use crate::evaluator::{Value, Env};
use crate::template::ZenoBladeParser;
use crate::evaluator::Evaluator;
use sqlx::{AnyPool, Row, Column, TypeInfo};
use std::sync::{Arc, Mutex};
use std::collections::HashMap;
use std::future::Future;
use std::pin::Pin;
use reqwest::{Method, header::{HeaderMap, HeaderName, HeaderValue}};
use std::str::FromStr;
use tokio::fs;
use std::path::Path;
use chrono::prelude::*;
use sha2::{Sha256, Digest};
use uuid::Uuid;
use rand::Rng; // For .random()
use regex::Regex;
use base64::prelude::*;
use bcrypt::{hash, verify, DEFAULT_COST};
use crate::plugins::sidecar::SidecarManager;
use crate::plugins::wasm::WasmManager;
use std::sync::OnceLock;

// Global Sidecar Manager (Lazy Init)
static SIDECAR_MANAGER: OnceLock<Arc<SidecarManager>> = OnceLock::new();
static WASM_MANAGER: OnceLock<Arc<WasmManager>> = OnceLock::new();

fn get_sidecar_manager() -> Arc<SidecarManager> {
    SIDECAR_MANAGER.get_or_init(|| Arc::new(SidecarManager::new())).clone()
}

fn get_wasm_manager() -> Arc<WasmManager> {
    WASM_MANAGER.get_or_init(|| Arc::new(WasmManager::new().expect("Failed to init WASM"))).clone()
}

type BuiltinFn = fn(Vec<Value>, Env, Option<AnyPool>) -> Pin<Box<dyn Future<Output = Value> + Send>>;

pub fn register(env: &mut Env) {
    // --- CORE ---
    env.set("len".to_string(), Value::Builtin("len".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        match &args[0] {
            Value::String(s) => Value::Integer(s.len() as i64),
            Value::Array(arr) => Value::Integer(arr.lock().unwrap().len() as i64),
            Value::Map(map) => Value::Integer(map.lock().unwrap().len() as i64),
            _ => Value::Integer(0),
        }
    })));

    env.set("sleep".to_string(), Value::Builtin("sleep".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        if let Value::Integer(ms) = args[0] {
            tokio::time::sleep(std::time::Duration::from_millis(ms as u64)).await;
        }
        Value::Null
    })));

    // --- STRING ---
    env.set("upper".to_string(), Value::Builtin("upper".to_string(), |args, _, _| Box::pin(async move {
         if args.len() != 1 { return Value::Null; }
         match &args[0] { Value::String(s) => Value::String(s.to_uppercase()), _ => Value::Null }
    })));

    env.set("str".to_string(), Value::Builtin("str".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        Value::String(format!("{}", args[0]))
    })));

    env.set("str_concat".to_string(), Value::Builtin("str_concat".to_string(), |args, _, _| Box::pin(async move {
        let mut result = String::new();
        for arg in args {
            result.push_str(&format!("{}", arg));
        }
        Value::String(result)
    })));

    env.set("str_replace".to_string(), Value::Builtin("str_replace".to_string(), |args, _, _| Box::pin(async move {
        if args.len() < 3 { return Value::Null; }
        let s = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let from = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };
        let to = match &args[2] { Value::String(s) => s.clone(), _ => return Value::Null };
        Value::String(s.replace(&from, &to))
    })));

    // --- ARRAY/MAP ---
    env.set("push".to_string(), Value::Builtin("push".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 2 { return Value::Null; }
        if let Value::Array(arr) = &args[0] {
            arr.lock().unwrap().push(args[1].clone());
            return Value::Array(arr.clone());
        }
        Value::Null
    })));

    env.set("arr_join".to_string(), Value::Builtin("arr_join".to_string(), |args, _, _| Box::pin(async move {
        if args.len() < 2 { return Value::Null; }
        if let Value::Array(arr) = &args[0] {
            let sep = match &args[1] { Value::String(s) => s.clone(), _ => ",".to_string() };
            let vec = arr.lock().unwrap();
            let strings: Vec<String> = vec.iter().map(|v| format!("{}", v)).collect();
            return Value::String(strings.join(&sep));
        }
        Value::Null
    })));

    env.set("arr_pop".to_string(), Value::Builtin("arr_pop".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        if let Value::Array(arr) = &args[0] {
            return arr.lock().unwrap().pop().unwrap_or(Value::Null);
        }
        Value::Null
    })));

    env.set("map_keys".to_string(), Value::Builtin("map_keys".to_string(), |args, _, _| Box::pin(async move {
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

    // --- HTTP ---
    env.set("http_get".to_string(), Value::Builtin("http_get".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let url = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        match reqwest::get(&url).await {
            Ok(resp) => match resp.text().await { Ok(t) => Value::String(t), Err(_) => Value::Null },
            Err(_) => Value::Null
        }
    })));

    env.set("http_post".to_string(), Value::Builtin("http_post".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 2 { return Value::Null; }
        let url = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let body = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };
        let client = reqwest::Client::new();
        match client.post(&url).body(body).send().await {
            Ok(resp) => match resp.text().await { Ok(t) => Value::String(t), Err(_) => Value::Null },
            Err(_) => Value::Null
        }
    })));

    env.set("fetch".to_string(), Value::Builtin("fetch".to_string(), |args, _, _| Box::pin(async move {
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

    env.set("response_json".to_string(), Value::Builtin("response_json".to_string(), |args, mut env, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        env.update("_response_body", args[0].clone());
        if let Some(Value::Null) = env.get("_response_status") {
            env.update("_response_status", Value::Integer(200));
        }
        Value::Null
    })));

    env.set("response_status".to_string(), Value::Builtin("response_status".to_string(), |args, mut env, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        if let Value::Integer(_) = args[0] {
            env.update("_response_status", args[0].clone());
        }
        Value::Null
    })));

    // --- DB ---
    env.set("db_execute".to_string(), Value::Builtin("db_execute".to_string(), |args, _, pool| Box::pin(async move {
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

    env.set("db_query".to_string(), Value::Builtin("db_query".to_string(), |args, _, pool| Box::pin(async move {
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
                        // try_get is strict with types in AnyRow if not using raw
                        // But for SQLite/Any, type mapping can be tricky.
                        // We try int first (covers boolean in sqlite), then string.
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

    // --- TEMPLATE ---
    env.set("view".to_string(), Value::Builtin("view".to_string(), |args, _, pool| Box::pin(async move {
        if args.len() < 1 { return Value::Null; }
        let tpl_input = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };

        let tpl = if tpl_input.ends_with(".zl") || tpl_input.contains("/") {
             match tokio::fs::read_to_string(&tpl_input).await {
                 Ok(c) => c,
                 Err(_) => tpl_input.clone()
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

        if let Some(Value::String(layout_path)) = renderer.env.get("__layout") {
             if let Ok(layout_content) = tokio::fs::read_to_string(&layout_path).await {
                 let mut layout_parser = ZenoBladeParser::new(&layout_content);
                 let layout_nodes = layout_parser.parse();
                 output = renderer.render_nodes(layout_nodes).await;
             }
        }

        Value::String(output)
    })));

    // --- FILESYSTEM ---
    env.set("file_read".to_string(), Value::Builtin("file_read".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        match fs::read_to_string(&path_str).await {
            Ok(content) => Value::String(content),
            Err(_) => Value::Null
        }
    })));

    env.set("file_write".to_string(), Value::Builtin("file_write".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 2 { return Value::Null; }
        let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let content = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };

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

    env.set("file_delete".to_string(), Value::Builtin("file_delete".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        match fs::remove_file(&path_str).await {
            Ok(_) => Value::Boolean(true),
            Err(_) => Value::Boolean(false)
        }
    })));

    env.set("dir_create".to_string(), Value::Builtin("dir_create".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        match fs::create_dir_all(&path_str).await {
            Ok(_) => Value::Boolean(true),
            Err(_) => Value::Boolean(false)
        }
    })));

    // --- TIME ---
    env.set("time_now".to_string(), Value::Builtin("time_now".to_string(), |_, _, _| Box::pin(async move {
        let now = Local::now();
        Value::String(now.to_rfc3339())
    })));

    env.set("time_format".to_string(), Value::Builtin("time_format".to_string(), |args, _, _| Box::pin(async move {
        if args.len() < 2 { return Value::Null; }
        let time_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let format_str = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };

        if let Ok(dt) = DateTime::parse_from_rfc3339(&time_str) {
            return Value::String(dt.format(&format_str).to_string());
        }
        Value::Null
    })));

    // --- JSON ---
    env.set("json_parse".to_string(), Value::Builtin("json_parse".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let json_str = match &args[0] { Value::String(s) => s, _ => return Value::Null };
        match serde_json::from_str::<serde_json::Value>(json_str) {
            Ok(v) => crate::evaluator::json_to_value(v), // Helper needs to be public or we move it
            Err(_) => Value::Null
        }
    })));

    env.set("json_stringify".to_string(), Value::Builtin("json_stringify".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        match serde_json::to_string(&args[0]) {
            Ok(s) => Value::String(s),
            Err(_) => Value::Null
        }
    })));

    // --- ENCODING ---
    env.set("base64_encode".to_string(), Value::Builtin("base64_encode".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let s = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        Value::String(BASE64_STANDARD.encode(s))
    })));

    env.set("base64_decode".to_string(), Value::Builtin("base64_decode".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let s = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        if let Ok(bytes) = BASE64_STANDARD.decode(s) {
            if let Ok(utf8) = String::from_utf8(bytes) {
                return Value::String(utf8);
            }
        }
        Value::Null
    })));

    env.set("hex_encode".to_string(), Value::Builtin("hex_encode".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let s = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        Value::String(hex::encode(s))
    })));

    env.set("hex_decode".to_string(), Value::Builtin("hex_decode".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let s = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        if let Ok(bytes) = hex::decode(s) {
            if let Ok(utf8) = String::from_utf8(bytes) {
                return Value::String(utf8);
            }
        }
        Value::Null
    })));

    // --- REGEX ---
    env.set("regex_match".to_string(), Value::Builtin("regex_match".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 2 { return Value::Null; }
        let pattern = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let text = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };

        if let Ok(re) = Regex::new(&pattern) {
            return Value::Boolean(re.is_match(&text));
        }
        Value::Boolean(false)
    })));

    env.set("regex_replace".to_string(), Value::Builtin("regex_replace".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 3 { return Value::Null; }
        let pattern = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let text = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };
        let replacement = match &args[2] { Value::String(s) => s.clone(), _ => return Value::Null };

        if let Ok(re) = Regex::new(&pattern) {
            let res = re.replace_all(&text, replacement.as_str());
            return Value::String(res.to_string());
        }
        Value::Null
    })));

    // --- VALIDATOR ---
    env.set("is_email".to_string(), Value::Builtin("is_email".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Boolean(false); }
        let s = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Boolean(false) };
        // Simple regex for email
        let re = Regex::new(r"^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$").unwrap();
        Value::Boolean(re.is_match(&s))
    })));

    env.set("is_numeric".to_string(), Value::Builtin("is_numeric".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Boolean(false); }
        match &args[0] {
            Value::Integer(_) => Value::Boolean(true),
            Value::String(s) => Value::Boolean(s.parse::<f64>().is_ok()),
            _ => Value::Boolean(false)
        }
    })));

    // --- UTILS ---
    env.set("env_get".to_string(), Value::Builtin("env_get".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let key = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        match std::env::var(&key) {
            Ok(val) => Value::String(val),
            Err(_) => Value::Null
        }
    })));

    env.set("hash_sha256".to_string(), Value::Builtin("hash_sha256".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let s = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let mut hasher = Sha256::new();
        hasher.update(s);
        let result = hasher.finalize();
        Value::String(hex::encode(result))
    })));

    env.set("password_hash".to_string(), Value::Builtin("password_hash".to_string(), |args, _, _| Box::pin(async move {
        if args.len() < 1 { return Value::Null; }
        let password = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let cost = if args.len() > 1 {
            match &args[1] { Value::Integer(i) => *i as u32, _ => DEFAULT_COST }
        } else {
            DEFAULT_COST
        };

        match hash(&password, cost) {
            Ok(h) => Value::String(h),
            Err(_) => Value::Null
        }
    })));

    env.set("password_verify".to_string(), Value::Builtin("password_verify".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 2 { return Value::Boolean(false); }
        let password = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Boolean(false) };
        let hash_str = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Boolean(false) };

        match verify(&password, &hash_str) {
            Ok(valid) => Value::Boolean(valid),
            Err(_) => Value::Boolean(false)
        }
    })));

    env.set("uuid_v4".to_string(), Value::Builtin("uuid_v4".to_string(), |_, _, _| Box::pin(async move {
        Value::String(Uuid::new_v4().to_string())
    })));

    env.set("random_int".to_string(), Value::Builtin("random_int".to_string(), |args, _, _| Box::pin(async move {
        let mut rng = rand::rng();
        if args.len() == 2 {
            if let (Value::Integer(min), Value::Integer(max)) = (&args[0], &args[1]) {
                return Value::Integer(rng.random_range(*min..*max));
            }
        }
        Value::Integer(rng.random())
    })));

    env.set("coalesce".to_string(), Value::Builtin("coalesce".to_string(), |args, _, _| Box::pin(async move {
        for arg in args {
            if let Value::Null = arg { continue; }
            return arg;
        }
        Value::Null
    })));

    // --- MODULARITY ---
    env.set("include".to_string(), Value::Builtin("include".to_string(), |args, mut env, pool| Box::pin(async move {
        if args.len() != 1 { return Value::Null; }
        let path_str = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };

        if let Ok(content) = tokio::fs::read_to_string(&path_str).await {
            use crate::parser::Parser;
            let mut parser = Parser::new(&content);
            let statements = parser.parse();

            let mut sub_evaluator = Evaluator::new(pool);
            sub_evaluator.env = env;
            sub_evaluator.eval(statements).await;
        }
        Value::Null
    })));

    // --- PLUGINS (Sidecar) ---
    env.set("plugin_load_sidecar".to_string(), Value::Builtin("plugin_load_sidecar".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 3 { return Value::Boolean(false); }
        let name = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Boolean(false) };
        let binary = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Boolean(false) };
        let cwd = match &args[2] { Value::String(s) => s.clone(), _ => return Value::Boolean(false) };

        let manager = get_sidecar_manager();
        match manager.load_plugin(name, binary, cwd).await {
            Ok(_) => Value::Boolean(true),
            Err(e) => {
                eprintln!("Plugin Load Error: {}", e);
                Value::Boolean(false)
            }
        }
    })));

    env.set("plugin_call".to_string(), Value::Builtin("plugin_call".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 3 { return Value::Null; }
        let name = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let method = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };
        // params handling: convert Value to serde_json::Value
        // For simple test, we pass args[2] assuming it is a Map or String
        let params_val = args[2].clone();

        // Helper to convert Value -> serde_json::Value
        // We rely on Serialize implementation of Value
        let params_json = serde_json::to_value(&params_val).unwrap_or(serde_json::Value::Null);

        let manager = get_sidecar_manager();
        match manager.call(&name, &method, params_json).await {
            Ok(v) => crate::evaluator::json_to_value(v),
            Err(e) => {
                eprintln!("Plugin Call Error: {}", e);
                Value::Null
            }
        }
    })));

    // --- WASM ---
    env.set("wasm_load".to_string(), Value::Builtin("wasm_load".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 2 { return Value::Boolean(false); }
        let name = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Boolean(false) };
        let path = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Boolean(false) };

        let manager = get_wasm_manager();
        match manager.load_plugin(name, path).await {
            Ok(_) => Value::Boolean(true),
            Err(e) => {
                eprintln!("WASM Load Error: {}", e);
                Value::Boolean(false)
            }
        }
    })));

    env.set("wasm_call".to_string(), Value::Builtin("wasm_call".to_string(), |args, _, _| Box::pin(async move {
        if args.len() != 2 { return Value::Null; }
        let name = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let func = match &args[1] { Value::String(s) => s.clone(), _ => return Value::Null };
        let params = serde_json::Value::Null; // Params not fully implemented for MVP WASM

        let manager = get_wasm_manager();
        match manager.call(&name, &func, params).await {
            Ok(_) => Value::Boolean(true), // Returns boolean success for void calls
            Err(e) => {
                eprintln!("WASM Call Error: {}", e);
                Value::Null
            }
        }
    })));

    // --- ROUTER ---
    env.set("router_get".to_string(), Value::Builtin("router_get".to_string(), |args, mut env, pool| Box::pin(async move {
        if args.len() != 2 { return Value::Null; }
        let path = match &args[0] { Value::String(s) => s.clone(), _ => return Value::Null };
        let handler = args[1].clone();

        let request_val = env.get("request");
        let route_handled = env.get("_route_handled");

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

    env.set("router_post".to_string(), Value::Builtin("router_post".to_string(), |args, mut env, pool| Box::pin(async move {
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
