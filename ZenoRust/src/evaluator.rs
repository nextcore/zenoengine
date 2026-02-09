use crate::parser::{Statement, Expression, Op, Parser};
use crate::template::{ZenoBladeParser, BladeNode};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use async_recursion::async_recursion;
use std::future::Future;
use std::pin::Pin;
use sqlx::AnyPool;
use serde::Serialize;

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
    // Wrapper for Rust Objects (like QueryBuilder)
    // We use Arc<Mutex<Any>> but need Any + Send + Sync.
    // Since we know the types we want, let's make a specific variant for now or use a trait object.
    // For simplicity, let's add specific variant for QueryBuilder.
    QueryBuilder(Arc<Mutex<crate::db_builder::ZenoQueryBuilder>>),
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
            },
            Value::QueryBuilder(_) => write!(f, "<QueryBuilder>"),
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
        if let Ok(store) = self.store.lock() {
            if let Some(val) = store.get(name) {
                return Some(val.clone());
            }
        }
        if let Some(ref outer) = self.outer {
            return outer.get(name);
        }
        None
    }

    pub fn set(&mut self, name: String, val: Value) {
        if let Ok(mut store) = self.store.lock() {
            store.insert(name, val);
        }
    }

    pub fn update(&mut self, name: &str, val: Value) -> bool {
         {
             if let Ok(mut store) = self.store.lock() {
                 if store.contains_key(name) {
                     store.insert(name.to_string(), val);
                     return true;
                 }
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
        crate::builtins::register(&mut self.env);
    }

    #[async_recursion]
    pub async fn render_nodes(&mut self, nodes: Vec<BladeNode>) -> String {
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
    pub async fn eval_statement(&mut self, stmt: Statement) -> Option<Value> {
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
pub fn json_to_value(v: serde_json::Value) -> Value {
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
