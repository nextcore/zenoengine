use crate::parser::{Statement, Expression, Op};
use crate::template::{ZenoBladeParser, BladeNode};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use async_recursion::async_recursion;
use std::future::Future;
use std::pin::Pin;
use sqlx::{AnyPool, Row, Column, TypeInfo};
use reqwest::{Method, header::{HeaderMap, HeaderName, HeaderValue}};
use std::str::FromStr;
use serde::Serialize;
use tokio::fs;
use std::path::Path;
use chrono::prelude::*;

#[derive(Debug, Clone)]
pub enum Value {
    Integer(i64),
    String(String),
    Boolean(bool),
    Null,
    Function(Vec<String>, Statement, Env),
    Builtin(String, fn(Vec<Value>, Env, Option<AnyPool>) -> Pin<Box<dyn Future<Output = Value> + Send>>),
    ReturnValue(Box<Value>),
    Array(Arc<Mutex<Vec<Value>>>),
    Map(Arc<Mutex<HashMap<String, Value>>>),
}

impl Serialize for Value {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: serde::Serializer,
    {
        match self {
            Value::Integer(i) => serializer.serialize_i64(*i),
            Value::String(s) => serializer.serialize_str(s),
            Value::Boolean(b) => serializer.serialize_bool(*b),
            Value::Null => serializer.serialize_none(),
            Value::Array(arr) => {
                let vec = arr.lock().unwrap();
                use serde::ser::SerializeSeq;
                let mut seq = serializer.serialize_seq(Some(vec.len()))?;
                for element in vec.iter() {
                    seq.serialize_element(element)?;
                }
                seq.end()
            },
            Value::Map(map) => {
                let m = map.lock().unwrap();
                use serde::ser::SerializeMap;
                let mut map_ser = serializer.serialize_map(Some(m.len()))?;
                for (k, v) in m.iter() {
                    map_ser.serialize_entry(k, v)?;
                }
                map_ser.end()
            },
            _ => serializer.serialize_none(),
        }
    }
}

// Manual PartialEq
impl PartialEq for Value {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (Value::Integer(a), Value::Integer(b)) => a == b,
            (Value::String(a), Value::String(b)) => a == b,
            (Value::Boolean(a), Value::Boolean(b)) => a == b,
            (Value::Null, Value::Null) => true,
            (Value::Function(params_a, body_a, _), Value::Function(params_b, body_b, _)) => {
                params_a == params_b && body_a == body_b
            },
            (Value::Builtin(name_a, _), Value::Builtin(name_b, _)) => name_a == name_b,
            (Value::ReturnValue(a), Value::ReturnValue(b)) => a == b,
            (Value::Array(a), Value::Array(b)) => {
                let vec_a = a.lock().unwrap();
                let vec_b = b.lock().unwrap();
                vec_a.iter().zip(vec_b.iter()).all(|(x, y)| x == y) && vec_a.len() == vec_b.len()
            },
            (Value::Map(a), Value::Map(b)) => {
                let map_a = a.lock().unwrap();
                let map_b = b.lock().unwrap();
                map_a.len() == map_b.len() && map_a.iter().all(|(k, v)| map_b.get(k) == Some(v))
            }
            _ => false,
        }
    }
}

impl std::fmt::Display for Value {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Value::Integer(i) => write!(f, "{}", i),
            Value::String(s) => write!(f, "{}", s),
            Value::Boolean(b) => write!(f, "{}", b),
            Value::Null => write!(f, "null"),
            Value::Function(params, _, _) => write!(f, "fn({})", params.join(", ")),
            Value::Builtin(name, _) => write!(f, "builtin({})", name),
            Value::ReturnValue(val) => write!(f, "return {}", val),
            Value::Array(arr) => {
                let vec = arr.lock().unwrap();
                let elements: Vec<String> = vec.iter().map(|v| format!("{}", v)).collect();
                write!(f, "[{}]", elements.join(", "))
            },
            Value::Map(map) => {
                let m = map.lock().unwrap();
                let mut entries: Vec<String> = m.iter().map(|(k, v)| format!("\"{}\": {}", k, v)).collect();
                entries.sort();
                write!(f, "{{{}}}", entries.join(", "))
            }
        }
    }
}

#[derive(Debug, Clone)]
pub struct Env {
    store: Arc<Mutex<HashMap<String, Value>>>,
    outer: Option<Box<Env>>,
}

impl Env {
    pub fn new() -> Self {
        Self {
            store: Arc::new(Mutex::new(HashMap::new())),
            outer: None,
        }
    }

    pub fn new_with_outer(outer: Env) -> Self {
        Self {
            store: Arc::new(Mutex::new(HashMap::new())),
            outer: Some(Box::new(outer)),
        }
    }

    pub fn get(&self, name: &str) -> Option<Value> {
        if let Some(val) = self.store.lock().unwrap().get(name) {
            return Some(val.clone());
        }
        if let Some(ref outer) = self.outer {
            return outer.get(name);
        }
        None
    }

    pub fn set(&mut self, name: String, val: Value) {
        self.store.lock().unwrap().insert(name, val);
    }

    pub fn update(&mut self, name: &str, val: Value) -> bool {
         {
             let mut store = self.store.lock().unwrap();
             if store.contains_key(name) {
                 store.insert(name.to_string(), val);
                 return true;
             }
         }
         if let Some(ref mut outer) = self.outer {
             return outer.update(name, val);
         }
         false
    }
}

pub struct Evaluator {
    pub env: Env,
    output: String,
    db_pool: Option<AnyPool>,
}

impl Evaluator {
    pub fn new(db_pool: Option<AnyPool>) -> Self {
        let mut evaluator = Self {
            env: Env::new(),
            output: String::new(),
            db_pool,
        };
        evaluator.register_builtins();
        evaluator
    }

    pub fn get_output(&self) -> String {
        self.output.clone()
    }

    pub fn get_final_response(&self) -> Option<(i32, Value)> {
        let status = self.env.get("_response_status");
        let body = self.env.get("_response_body");

        if let (Some(Value::Integer(s)), Some(b)) = (status, body) {
            Some((s as i32, b))
        } else {
            None
        }
    }

    pub fn set_request_context(&mut self, method: &str, path: &str, query: &str, body: &str) {
        let mut req_map = HashMap::new();
        req_map.insert("method".to_string(), Value::String(method.to_string()));
        req_map.insert("path".to_string(), Value::String(path.to_string()));
        req_map.insert("query".to_string(), Value::String(query.to_string()));
        req_map.insert("body".to_string(), Value::String(body.to_string()));

        self.env.set("request".to_string(), Value::Map(Arc::new(Mutex::new(req_map))));
        self.env.set("_response_body".to_string(), Value::Null);
        self.env.set("_response_status".to_string(), Value::Null);
        // Track if a route has been handled
        self.env.set("_route_handled".to_string(), Value::Boolean(false));
    }

    fn register_builtins(&mut self) {
        // ... (Standard Built-ins)
        // I will preserve existing builtins and add router_get/post.

        self.env.set("len".to_string(), Value::Builtin("len".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            match &args[0] {
                Value::String(s) => Value::Integer(s.len() as i64),
                Value::Array(arr) => Value::Integer(arr.lock().unwrap().len() as i64),
                Value::Map(map) => Value::Integer(map.lock().unwrap().len() as i64),
                _ => Value::Integer(0),
            }
        })));

        self.env.set("upper".to_string(), Value::Builtin("upper".to_string(), |args, _, _| Box::pin(async move {
             if args.len() != 1 { return Value::Null; }
             match &args[0] { Value::String(s) => Value::String(s.to_uppercase()), _ => Value::Null }
        })));

        self.env.set("str".to_string(), Value::Builtin("str".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            Value::String(format!("{}", args[0]))
        })));

        self.env.set("str_concat".to_string(), Value::Builtin("str_concat".to_string(), |args, _, _| Box::pin(async move {
            let mut result = String::new();
            for arg in args {
                result.push_str(&format!("{}", arg));
            }
            Value::String(result)
        })));

        self.env.set("str_replace".to_string(), Value::Builtin("str_replace".to_string(), |args, _, _| Box::pin(async move {
            if args.len() < 3 { return Value::Null; }
            let s = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let from = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };
            let to = match &args[2] { Value::String(s) => s.clone(), _ => return Value::Null };
            Value::String(s.replace(&from, &to))
        })));

        self.env.set("coalesce".to_string(), Value::Builtin("coalesce".to_string(), |args, _, _| Box::pin(async move {
            for arg in args {
                if let Value::Null = arg { continue; }
                return arg;
            }
            Value::Null
        })));

        self.env.set("sleep".to_string(), Value::Builtin("sleep".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            if let Value::Integer(ms) = args[0] {
                tokio::time::sleep(std::time::Duration::from_millis(ms as u64)).await;
            }
            Value::Null
        })));

        self.env.set("time_now".to_string(), Value::Builtin("time_now".to_string(), |_, _, _| Box::pin(async move {
            let now = Local::now();
            Value::String(now.to_rfc3339())
        })));

        self.env.set("time_format".to_string(), Value::Builtin("time_format".to_string(), |args, _, _| Box::pin(async move {
            if args.len() < 2 { return Value::Null; }
            let time_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let format_str = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };

            if let Ok(dt) = DateTime::parse_from_rfc3339(&time_str) {
                return Value::String(dt.format(&format_str).to_string());
            }
            Value::Null
        })));

        self.env.set("push".to_string(), Value::Builtin("push".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 2 { return Value::Null; }
            if let Value::Array(arr) = &args[0] {
                arr.lock().unwrap().push(args[1].clone());
                return Value::Array(arr.clone());
            }
            Value::Null
        })));

        self.env.set("arr_join".to_string(), Value::Builtin("arr_join".to_string(), |args, _, _| Box::pin(async move {
            if args.len() < 2 { return Value::Null; }
            if let Value::Array(arr) = &args[0] {
                let sep = match &args[1] { Value::String(s) => s.clone(), _ => ",".to_string() };
                let vec = arr.lock().unwrap();
                let strings: Vec<String> = vec.iter().map(|v| format!("{}", v)).collect();
                return Value::String(strings.join(&sep));
            }
            Value::Null
        })));

        self.env.set("arr_pop".to_string(), Value::Builtin("arr_pop".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            if let Value::Array(arr) = &args[0] {
                return arr.lock().unwrap().pop().unwrap_or(Value::Null);
            }
            Value::Null
        })));

        self.env.set("file_read".to_string(), Value::Builtin("file_read".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            match fs::read_to_string(&path_str).await {
                Ok(content) => Value::String(content),
                Err(_) => Value::Null
            }
        })));

        self.env.set("file_write".to_string(), Value::Builtin("file_write".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 2 { return Value::Null; }
            let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let content = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };

            // Security check
            let path = Path::new(&path_str);
            if let Some(ext) = path.extension() {
                let ext_str = ext.to_string_lossy().to_lowercase();
                if (ext_str == "zl" || ext_str == "env") && std::env::var("APP_ENV").unwrap_or_default() != "development" {
                    eprintln!("Security Error: Cannot write to sensitive files in production");
                    return Value::Null;
                }
            }

            if let Some(parent) = path.parent() {
                let _ = fs::create_dir_all(parent).await;
            }

            match fs::write(&path_str, content).await {
                Ok(_) => Value::Boolean(true),
                Err(_) => Value::Boolean(false)
            }
        })));

        self.env.set("file_delete".to_string(), Value::Builtin("file_delete".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            match fs::remove_file(&path_str).await {
                Ok(_) => Value::Boolean(true),
                Err(_) => Value::Boolean(false)
            }
        })));

        self.env.set("dir_create".to_string(), Value::Builtin("dir_create".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            match fs::create_dir_all(&path_str).await {
                Ok(_) => Value::Boolean(true),
                Err(_) => Value::Boolean(false)
            }
        })));

        self.env.set("map_keys".to_string(), Value::Builtin("map_keys".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            if let Value::Map(map) = &args[0] {
                let m = map.lock().unwrap();
                let mut keys = Vec::new();
                for k in m.keys() {
                    keys.push(Value::String(k.clone()));
                }
                return Value::Array(Arc::new(Mutex::new(keys)));
            }
            Value::Null
        })));

        self.env.set("http_get".to_string(), Value::Builtin("http_get".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            let url = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            match reqwest::get(&url).await {
                Ok(resp) => match resp.text().await { Ok(t) => Value::String(t), Err(_) => Value::Null },
                Err(_) => Value::Null
            }
        })));

        self.env.set("http_post".to_string(), Value::Builtin("http_post".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 2 { return Value::Null; }
            let url = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let body = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };
            let client = reqwest::Client::new();
            match client.post(&url).body(body).send().await {
                Ok(resp) => match resp.text().await { Ok(t) => Value::String(t), Err(_) => Value::Null },
                Err(_) => Value::Null
            }
        })));

        self.env.set("db_execute".to_string(), Value::Builtin("db_execute".to_string(), |args, _, pool| Box::pin(async move {
            if pool.is_none() {
                eprintln!("DB Error: No database pool configured");
                return Value::Null;
            }
            let pool = pool.unwrap();
            if args.len() < 1 { return Value::Null; }
            let sql = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let mut query = sqlx::query(&sql);
            if args.len() > 1 {
                if let Value::Array(params) = &args[1] {
                    for param in params.lock().unwrap().iter() {
                        match param {
                            Value::Integer(i) => query = query.bind(*i),
                            Value::String(s) => query = query.bind(s.clone()),
                            Value::Boolean(b) => query = query.bind(*b),
                            Value::Null => query = query.bind(None::<String>),
                            _ => query = query.bind(format!("{}", param)),
                        }
                    }
                }
            }
            match query.execute(&pool).await {
                Ok(result) => Value::Integer(result.rows_affected() as i64),
                Err(e) => {
                    eprintln!("DB Execute Error: {}", e);
                    Value::Null
                }
            }
        })));

        self.env.set("db_query".to_string(), Value::Builtin("db_query".to_string(), |args, _, pool| Box::pin(async move {
            if pool.is_none() {
                eprintln!("DB Error: No database pool configured");
                return Value::Null;
            }
            let pool = pool.unwrap();
            if args.len() < 1 { return Value::Null; }
            let sql = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let mut query = sqlx::query(&sql);
            if args.len() > 1 {
                if let Value::Array(params) = &args[1] {
                    for param in params.lock().unwrap().iter() {
                        match param {
                            Value::Integer(i) => query = query.bind(*i),
                            Value::String(s) => query = query.bind(s.clone()),
                            Value::Boolean(b) => query = query.bind(*b),
                            Value::Null => query = query.bind(None::<String>),
                            _ => query = query.bind(format!("{}", param)),
                        }
                    }
                }
            }
            match query.fetch_all(&pool).await {
                Ok(rows) => {
                    let mut result_rows = Vec::new();
                    for row in rows {
                        let mut map = HashMap::new();
                        for col in row.columns() {
                            let name = col.name();
                            let val = if let Ok(v) = row.try_get::<i64, _>(name) { Value::Integer(v) }
                            else if let Ok(v) = row.try_get::<String, _>(name) { Value::String(v) }
                            else if let Ok(v) = row.try_get::<bool, _>(name) { Value::Boolean(v) }
                            else { Value::Null };
                            map.insert(name.to_string(), val);
                        }
                        result_rows.push(Value::Map(Arc::new(Mutex::new(map))));
                    }
                    Value::Array(Arc::new(Mutex::new(result_rows)))
                },
                Err(e) => {
                    eprintln!("DB Query Error: {}", e);
                    Value::Null
                }
            }
        })));

        self.env.set("view".to_string(), Value::Builtin("view".to_string(), |args, _, pool| Box::pin(async move {
            if args.len() < 1 { return Value::Null; }
            let tpl_input = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };

            // Check if it's a file path (simple check: usually views are files)
            // But here the input might be raw string (as per previous tests) or a file path.
            // ZenoGo usually takes a file path or name.
            // If it ends with .zl, load file.

            let tpl = if tpl_input.ends_with(".zl") || tpl_input.contains("/") {
                 match tokio::fs::read_to_string(&tpl_input).await {
                     Ok(c) => c,
                     Err(_) => tpl_input.clone() // Fallback to raw string
                 }
            } else {
                tpl_input.clone()
            };

            let mut renderer = Evaluator::new(pool);
            if args.len() > 1 {
                let data = args[1].clone();
                if let Value::Map(map) = data {
                    let m = map.lock().unwrap();
                    for (k, v) in m.iter() {
                        renderer.env.set(k.clone(), v.clone());
                    }
                }
            }

            let mut parser = ZenoBladeParser::new(&tpl);
            let nodes = parser.parse();
            let mut output = renderer.render_nodes(nodes).await;

            // Handle @extends logic
            // If __layout was set during render, we need to render the layout now.
            if let Some(Value::String(layout_path)) = renderer.env.get("__layout") {
                 if let Ok(layout_content) = tokio::fs::read_to_string(&layout_path).await {
                     let mut layout_parser = ZenoBladeParser::new(&layout_content);
                     let layout_nodes = layout_parser.parse();
                     // Render layout with the same env (which now contains sections)
                     output = renderer.render_nodes(layout_nodes).await;
                 }
            }

            Value::String(output)
        })));

        self.env.set("response_json".to_string(), Value::Builtin("response_json".to_string(), |args, mut env, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            env.update("_response_body", args[0].clone());
            if let Some(Value::Null) = env.get("_response_status") {
                env.update("_response_status", Value::Integer(200));
            }
            Value::Null
        })));

        self.env.set("response_status".to_string(), Value::Builtin("response_status".to_string(), |args, mut env, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            if let Value::Integer(_) = args[0] {
                env.update("_response_status", args[0].clone());
            }
            Value::Null
        })));

        self.env.set("fetch".to_string(), Value::Builtin("fetch".to_string(), |args, _, _| Box::pin(async move {
            if args.len() < 1 { return Value::Null; }
            let url = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let mut method = Method::GET;
            let mut headers = HeaderMap::new();
            let mut body = String::new();
            if args.len() > 1 {
                if let Value::Map(opts_arc) = &args[1] {
                    let opts = opts_arc.lock().unwrap();
                    if let Some(Value::String(m)) = opts.get("method") {
                        method = Method::from_str(&m.to_uppercase()).unwrap_or(Method::GET);
                    }
                    if let Some(Value::String(b)) = opts.get("body") { body = b.clone(); }
                    if let Some(Value::Map(h_arc)) = opts.get("headers") {
                        let h_map = h_arc.lock().unwrap();
                        for (k, v) in h_map.iter() {
                            if let Value::String(v_str) = v {
                                if let (Ok(hn), Ok(hv)) = (HeaderName::from_str(k), HeaderValue::from_str(v_str)) {
                                    headers.insert(hn, hv);
                                }
                            }
                        }
                    }
                }
            }
            let client = reqwest::Client::new();
            let resp = client.request(method, &url).headers(headers).body(body).send().await;
            match resp {
                Ok(r) => {
                    let status = r.status().as_u16() as i64;
                    let mut resp_headers = HashMap::new();
                    for (k, v) in r.headers() {
                        if let Ok(v_str) = v.to_str() {
                            resp_headers.insert(k.to_string(), Value::String(v_str.to_string()));
                        }
                    }
                    let body_text = r.text().await.unwrap_or_default();
                    let mut result_map = HashMap::new();
                    result_map.insert("status".to_string(), Value::Integer(status));
                    result_map.insert("headers".to_string(), Value::Map(Arc::new(Mutex::new(resp_headers))));
                    result_map.insert("body".to_string(), Value::String(body_text));
                    Value::Map(Arc::new(Mutex::new(result_map)))
                },
                Err(_) => Value::Null
            }
        })));

        self.env.set("json_parse".to_string(), Value::Builtin("json_parse".to_string(), |args, _, _| Box::pin(async move {
            if args.len() != 1 { return Value::Null; }
            let json_str = match &args[0] { Value::String(s) => s, _ => return Value::Null };
            match serde_json::from_str::<serde_json::Value>(json_str) {
                Ok(v) => json_to_value(v),
                Err(_) => Value::Null
            }
        })));

        // NEW: Router Helpers
        self.env.set("router_get".to_string(), Value::Builtin("router_get".to_string(), |args, mut env, pool| Box::pin(async move {
            if args.len() != 2 { return Value::Null; }
            let path = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let handler = args[1].clone();

            let request_val = env.get("request");
            let route_handled = env.get("_route_handled");

            // Check if already handled
            if let Some(Value::Boolean(true)) = route_handled {
                return Value::Null;
            }

            let mut match_found = false;
            if let Some(Value::Map(req_map)) = request_val {
                let (method, current_path) = {
                    let req = req_map.lock().unwrap();
                    (
                        req.get("method").cloned().unwrap_or(Value::Null),
                        req.get("path").cloned().unwrap_or(Value::Null)
                    )
                };

                if method == Value::String("GET".to_string()) && current_path == Value::String(path) {
                    match_found = true;
                }
            }

            if match_found {
                env.update("_route_handled", Value::Boolean(true));
                if let Value::Function(_, body, closure_env) = handler {
                    let mut sub_evaluator = Evaluator::new(pool);
                    sub_evaluator.env = Env::new_with_outer(closure_env);
                    sub_evaluator.eval_statement(body).await;
                }
            }
            Value::Null
        })));

        self.env.set("router_post".to_string(), Value::Builtin("router_post".to_string(), |args, mut env, pool| Box::pin(async move {
            if args.len() != 2 { return Value::Null; }
            let path = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
            let handler = args[1].clone();

            let request_val = env.get("request");
            let route_handled = env.get("_route_handled");

            if let Some(Value::Boolean(true)) = route_handled { return Value::Null; }

            let mut match_found = false;
            if let Some(Value::Map(req_map)) = request_val {
                let (method, current_path) = {
                    let req = req_map.lock().unwrap();
                    (
                        req.get("method").cloned().unwrap_or(Value::Null),
                        req.get("path").cloned().unwrap_or(Value::Null)
                    )
                };

                if method == Value::String("POST".to_string()) && current_path == Value::String(path) {
                    match_found = true;
                }
            }

            if match_found {
                env.update("_route_handled", Value::Boolean(true));
                if let Value::Function(_, body, closure_env) = handler {
                    let mut sub_evaluator = Evaluator::new(pool);
                    sub_evaluator.env = Env::new_with_outer(closure_env);
                    sub_evaluator.eval_statement(body).await;
                }
            }
            Value::Null
        })));
    }

    #[async_recursion]
    async fn render_nodes(&mut self, nodes: Vec<BladeNode>) -> String {
        let mut output = String::new();
        for node in nodes {
            match node {
                BladeNode::Text(t) => output.push_str(&t),
                BladeNode::Interpolation(expr) => {
                    if let Some(val) = self.eval_expression(&expr).await {
                        output.push_str(&format!("{}", val));
                    }
                },
                BladeNode::If(cond, true_block, false_block) => {
                    if let Some(val) = self.eval_expression(&cond).await {
                        if self.is_truthy(val) {
                            output.push_str(&self.render_nodes(true_block).await);
                        } else if let Some(fb) = false_block {
                            output.push_str(&self.render_nodes(fb).await);
                        }
                    }
                }
                BladeNode::ForEach(collection_expr, item_name, block) => {
                    if let Some(collection_val) = self.eval_expression(&collection_expr).await {
                        if let Value::Array(arr) = collection_val {
                             let elements = {
                                 let vec = arr.lock().unwrap();
                                 vec.clone()
                             };

                             for element in elements {
                                 let previous_env = self.env.clone();
                                 self.env = Env::new_with_outer(previous_env.clone());
                                 self.env.set(item_name.clone(), element);

                                 output.push_str(&self.render_nodes(block.clone()).await);

                                 self.env = previous_env;
                             }
                        }
                    }
                }
                BladeNode::Include(path) => {
                    // TODO: Ideally resolve path relative to views, but for now assuming direct path or handled by caller
                    if let Ok(content) = tokio::fs::read_to_string(&path).await {
                        // Render included content with current env
                        // We need to parse it first
                        let mut parser = ZenoBladeParser::new(&content);
                        let included_nodes = parser.parse();
                        output.push_str(&self.render_nodes(included_nodes).await);
                    }
                }
                BladeNode::Section(name, block) => {
                    // Sections are captured, not rendered immediately if we are extending
                    // But if we are just defining a section in a normal view (not extending), it might be ignored or rendered?
                    // Usually @section defines content for a parent.
                    // We need a way to store sections.
                    let content = self.render_nodes(block).await;
                    let section_key = format!("__section_{}", name);
                    self.env.set(section_key, Value::String(content));
                }
                BladeNode::Yield(name) => {
                    let section_key = format!("__section_{}", name);
                    if let Some(Value::String(content)) = self.env.get(&section_key) {
                        output.push_str(&content);
                    }
                }
                BladeNode::Extends(path) => {
                    // This is the tricky part.
                    // 1. We must assume we have already parsed and executed the rest of the file (which contains @sections).
                    // 2. Actually, when we encounter @extends, it usually means THIS file is a child.
                    //    The parser returns a list of nodes.
                    //    If we are processing nodes and hit Extends, we should probably stop rendering current output
                    //    and instead switch to rendering the parent, but carrying over the sections we defined.

                    // However, `render_nodes` processes sequentially.
                    // Sections might be defined BEFORE or AFTER @extends.
                    // Standard Blade often puts @extends at top.
                    // If @extends is present, the "main" output of this file (outside sections) is usually discarded?

                    // For simplicity: We store the layout path in env, and after rendering the current view,
                    // if layout is set, we render the layout.

                    self.env.set("__layout".to_string(), Value::String(path));
                }
            }
        }
        output
    }

    pub async fn eval(&mut self, statements: Vec<Statement>) {
        for stmt in statements {
            let result = self.eval_statement(stmt).await;
            if let Some(Value::ReturnValue(_)) = result {
            }
        }
    }

    #[async_recursion]
    async fn eval_statement(&mut self, stmt: Statement) -> Option<Value> {
        match stmt {
            Statement::Print(expr) => {
                let val = self.eval_expression(&expr).await?;
                let out = format!("{}\n", val);
                self.output.push_str(&out);
                print!("{}", out);
                None
            }
            Statement::Let(name, expr) => {
                let val = self.eval_expression(&expr).await?;
                self.env.set(name, val);
                None
            }
             Statement::Assign(lhs, rhs) => {
                let val = self.eval_expression(&rhs).await?;
                match lhs {
                    Expression::Identifier(name) => {
                        if !self.env.update(&name, val.clone()) {
                             eprintln!("Runtime Error: Variable '{}' not declared before assignment", name);
                             return None;
                        }
                    },
                    Expression::Index(target_expr, index_expr) => {
                         let target = self.eval_expression(&target_expr).await?;
                         let index = self.eval_expression(&index_expr).await?;

                         match (target, index) {
                             (Value::Array(arr), Value::Integer(i)) => {
                                 let mut vec = arr.lock().unwrap();
                                 if i >= 0 && (i as usize) < vec.len() {
                                     vec[i as usize] = val;
                                 } else {
                                     eprintln!("Runtime Error: Index out of bounds");
                                 }
                             },
                             (Value::Map(map), Value::String(key)) => {
                                 let mut m = map.lock().unwrap();
                                 m.insert(key, val);
                             },
                             _ => {
                                 eprintln!("Runtime Error: Invalid assignment target");
                             }
                         }
                    },
                    _ => {
                        eprintln!("Runtime Error: Invalid assignment target");
                    }
                }
                None
            }
            Statement::Function(name, params, body) => {
                let func = Value::Function(params, *body, self.env.clone());
                self.env.set(name, func);
                None
            }
            Statement::Return(expr) => {
                let val = self.eval_expression(&expr).await?;
                Some(Value::ReturnValue(Box::new(val)))
            }
            Statement::Block(stmts) => {
                self.eval_block(stmts).await
            }
            Statement::If(condition, consequence, alternative) => {
                let cond_val = self.eval_expression(&condition).await?;
                if self.is_truthy(cond_val) {
                    self.eval_statement(*consequence).await
                } else if let Some(alt) = alternative {
                    self.eval_statement(*alt).await
                } else {
                    None
                }
            }
            Statement::Expression(expr) => {
                self.eval_expression(&expr).await
            }
        }
    }

    async fn eval_block(&mut self, stmts: Vec<Statement>) -> Option<Value> {
        let previous_env = self.env.clone();
        self.env = Env::new_with_outer(previous_env.clone());

        let mut result = None;
        for stmt in stmts {
            let val = self.eval_statement(stmt).await;
            if let Some(Value::ReturnValue(_)) = val {
                result = val;
                break;
            }
            result = val;
        }

        self.env = previous_env;
        result
    }

    #[async_recursion]
    async fn eval_expression(&mut self, expr: &Expression) -> Option<Value> {
        match expr {
            Expression::Integer(i) => Some(Value::Integer(*i)),
            Expression::StringLiteral(s) => Some(Value::String(s.clone())),
            Expression::Boolean(b) => Some(Value::Boolean(*b)),
            Expression::Null => Some(Value::Null),
            Expression::Identifier(name) => {
                if let Some(val) = self.env.get(name) {
                    Some(val)
                } else {
                    eprintln!("Runtime Error: Variable '{}' not found", name);
                    None
                }
            },
            Expression::Array(elements) => {
                let mut vals = Vec::new();
                for e in elements {
                    vals.push(self.eval_expression(e).await?);
                }
                Some(Value::Array(Arc::new(Mutex::new(vals))))
            },
             Expression::Map(pairs) => {
                let mut map = HashMap::new();
                for (k, v_expr) in pairs {
                    map.insert(k.clone(), self.eval_expression(v_expr).await?);
                }
                Some(Value::Map(Arc::new(Mutex::new(map))))
            },
            Expression::Index(left, index) => {
                 let left_val = self.eval_expression(left).await?;
                 let index_val = self.eval_expression(index).await?;

                 match (left_val, index_val) {
                     (Value::Array(arr), Value::Integer(i)) => {
                         let vec = arr.lock().unwrap();
                         if i >= 0 && (i as usize) < vec.len() {
                             Some(vec[i as usize].clone())
                         } else {
                             eprintln!("Runtime Error: Index out of bounds");
                             Some(Value::Null)
                         }
                     },
                     (Value::Map(map), Value::String(key)) => {
                         let m = map.lock().unwrap();
                         m.get(&key).cloned().or(Some(Value::Null))
                     },
                     _ => {
                         eprintln!("Runtime Error: Index operation not supported on this type");
                         None
                     }
                 }
            },
            Expression::BinaryOp(lhs, op, rhs) => {
                let l_val = self.eval_expression(lhs).await?;
                let r_val = self.eval_expression(rhs).await?;
                self.eval_infix_expression(op, l_val, r_val)
            },
            Expression::Call(func_expr, args) => {
                let func = self.eval_expression(func_expr).await?;
                let mut evaluated_args = Vec::new();
                for arg in args {
                    evaluated_args.push(self.eval_expression(arg).await?);
                }
                self.apply_function(func, evaluated_args).await
            }
            Expression::Function(params, body) => {
                Some(Value::Function(params.clone(), *body.clone(), self.env.clone()))
            }
        }
    }

    async fn apply_function(&mut self, func: Value, args: Vec<Value>) -> Option<Value> {
        match func {
            Value::Function(params, body, closure_env) => {
                if args.len() != params.len() {
                    eprintln!("Runtime Error: Function expected {} args, got {}", params.len(), args.len());
                    return None;
                }
                let mut extended_env = Env::new_with_outer(closure_env);
                for (param_name, arg_val) in params.iter().zip(args) {
                    extended_env.set(param_name.clone(), arg_val);
                }
                let previous_env = self.env.clone();
                self.env = extended_env;
                let result = self.eval_statement(body).await;
                self.env = previous_env;

                if let Some(Value::ReturnValue(val)) = result {
                    Some(*val)
                } else {
                    Some(Value::Null)
                }
            },
            Value::Builtin(_, func_ptr) => {
                Some(func_ptr(args, self.env.clone(), self.db_pool.clone()).await)
            },
            _ => {
                eprintln!("Runtime Error: Not a function");
                None
            }
        }
    }

    fn eval_infix_expression(&self, op: &Op, left: Value, right: Value) -> Option<Value> {
        match (left, op, right) {
            (Value::Integer(l), Op::Add, Value::Integer(r)) => Some(Value::Integer(l + r)),
            (Value::Integer(l), Op::Subtract, Value::Integer(r)) => Some(Value::Integer(l - r)),
            (Value::Integer(l), Op::Multiply, Value::Integer(r)) => Some(Value::Integer(l * r)),
            (Value::Integer(l), Op::Divide, Value::Integer(r)) => {
                if r == 0 { eprintln!("Runtime Error: Division by zero"); None } else { Some(Value::Integer(l / r)) }
            },

            (Value::String(l), Op::Add, Value::String(r)) => Some(Value::String(l + &r)),
            (Value::String(l), Op::Add, Value::Integer(r)) => Some(Value::String(format!("{}{}", l, r))),
            (Value::Integer(l), Op::Add, Value::String(r)) => Some(Value::String(format!("{}{}", l, r))),

            (Value::Integer(l), Op::Equal, Value::Integer(r)) => Some(Value::Boolean(l == r)),
            (Value::Integer(l), Op::NotEqual, Value::Integer(r)) => Some(Value::Boolean(l != r)),
            (Value::Integer(l), Op::LessThan, Value::Integer(r)) => Some(Value::Boolean(l < r)),
            (Value::Integer(l), Op::GreaterThan, Value::Integer(r)) => Some(Value::Boolean(l > r)),

            (Value::Boolean(l), Op::Equal, Value::Boolean(r)) => Some(Value::Boolean(l == r)),
            (Value::Boolean(l), Op::NotEqual, Value::Boolean(r)) => Some(Value::Boolean(l != r)),

            (Value::String(l), Op::Equal, Value::String(r)) => Some(Value::Boolean(l == r)),
            (Value::String(l), Op::NotEqual, Value::String(r)) => Some(Value::Boolean(l != r)),

            // Null Handling
            (Value::Null, Op::Equal, Value::Null) => Some(Value::Boolean(true)),
            (Value::Null, Op::NotEqual, Value::Null) => Some(Value::Boolean(false)),
            (_, Op::Equal, Value::Null) => Some(Value::Boolean(false)),
            (Value::Null, Op::Equal, _) => Some(Value::Boolean(false)),
            (_, Op::NotEqual, Value::Null) => Some(Value::Boolean(true)),
            (Value::Null, Op::NotEqual, _) => Some(Value::Boolean(true)),

            _ => {
                eprintln!("Runtime Error: Mismatched types in binary operation");
                None
            }
        }
    }

    fn is_truthy(&self, val: Value) -> bool {
        match val {
            Value::Boolean(b) => b,
            Value::Integer(i) => i != 0,
            Value::String(s) => !s.is_empty(),
            Value::Null => false,
            _ => false,
        }
    }
}

// Helper to convert Serde JSON to Zeno Value
fn json_to_value(v: serde_json::Value) -> Value {
    match v {
        serde_json::Value::Null => Value::Null,
        serde_json::Value::Bool(b) => Value::Boolean(b),
        serde_json::Value::Number(n) => {
            if let Some(i) = n.as_i64() {
                Value::Integer(i)
            } else {
                Value::Integer(n.as_f64().unwrap_or(0.0) as i64)
            }
        },
        serde_json::Value::String(s) => Value::String(s),
        serde_json::Value::Array(arr) => {
            let zeno_arr: Vec<Value> = arr.into_iter().map(json_to_value).collect();
            Value::Array(Arc::new(Mutex::new(zeno_arr)))
        },
        serde_json::Value::Object(obj) => {
            let mut map = HashMap::new();
            for (k, v) in obj {
                map.insert(k, json_to_value(v));
            }
            Value::Map(Arc::new(Mutex::new(map)))
        }
    }
}

#[cfg(test)]
mod tests {
    use crate::parser::Parser;
    use crate::evaluator::{Evaluator, Value};

    #[tokio::test]
    async fn test_functions() {
        let input = r#"
            fn add(a, b) {
                return a + b;
            }
            let res = add(5, 5);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new(None);
        evaluator.eval(statements).await;

        let val = evaluator.env.get("res");
        if let Some(Value::Integer(v)) = val {
            assert_eq!(v, 10);
        } else {
            panic!("Function return value not correct, got {:?}", val);
        }
    }

     #[tokio::test]
    async fn test_closure() {
        let input = r#"
            let factor = 2;
            fn multiply(a) {
                return a * factor;
            }
            let res = multiply(10);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new(None);
        evaluator.eval(statements).await;

        let val = evaluator.env.get("res");
        if let Some(Value::Integer(v)) = val {
            assert_eq!(v, 20);
        } else {
            panic!("Closure capture failed");
        }
    }

    #[tokio::test]
    async fn test_builtin() {
        let input = r#"
            let l = len("hello");
            let u = upper("hello");
            let s = str(123);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new(None);
        evaluator.eval(statements).await;

        assert_eq!(evaluator.env.get("l"), Some(Value::Integer(5)));
        assert_eq!(evaluator.env.get("u"), Some(Value::String("HELLO".to_string())));
        assert_eq!(evaluator.env.get("s"), Some(Value::String("123".to_string())));
    }

    #[tokio::test]
    async fn test_arrays() {
        let input = r#"
            let arr = [1, 2, 3];
            let first = arr[0];
            push(arr, 4);
            let length = len(arr);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new(None);
        evaluator.eval(statements).await;

        assert_eq!(evaluator.env.get("first"), Some(Value::Integer(1)));
        assert_eq!(evaluator.env.get("length"), Some(Value::Integer(4)));

        // Check array content manually
        if let Some(Value::Array(arr)) = evaluator.env.get("arr") {
             let vec = arr.lock().unwrap();
             assert_eq!(vec.len(), 4);
             assert_eq!(vec[3], Value::Integer(4));
        } else {
            panic!("arr not found or not array");
        }
    }

    #[tokio::test]
    async fn test_maps() {
        let input = r#"
            let user = {"name": "Zeno", "id": 1};
            let n = user["name"];
            user["id"] = 99;
            let i = user["id"];
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new(None);
        evaluator.eval(statements).await;

        assert_eq!(evaluator.env.get("n"), Some(Value::String("Zeno".to_string())));
        assert_eq!(evaluator.env.get("i"), Some(Value::Integer(99)));
    }
}
