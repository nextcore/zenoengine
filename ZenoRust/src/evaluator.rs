use crate::parser::{Statement, Expression, Op};
use crate::template::{ZenoBladeParser, BladeNode};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use async_recursion::async_recursion;
use std::future::Future;
use std::pin::Pin;
use sqlx::{AnyPool, Row, Column, TypeInfo, ValueRef};

#[derive(Debug, Clone)]
pub enum Value {
    Integer(i64),
    String(String),
    Boolean(bool),
    Null,
    Function(Vec<String>, Statement, Env),
    // Builtin now holds a function pointer that returns a BoxFuture
    Builtin(String, fn(Vec<Value>, Option<AnyPool>) -> Pin<Box<dyn Future<Output = Value> + Send>>),
    ReturnValue(Box<Value>),
    Array(Arc<Mutex<Vec<Value>>>),
    Map(Arc<Mutex<HashMap<String, Value>>>),
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
         } // Drop lock before checking outer

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

    fn register_builtins(&mut self) {
        // ... (existing built-ins omitted for brevity in thought, but included in file write)
        self.env.set("len".to_string(), Value::Builtin("len".to_string(), |args, _| {
            Box::pin(async move {
                if args.len() != 1 { return Value::Null; }
                match &args[0] {
                    Value::String(s) => Value::Integer(s.len() as i64),
                    Value::Array(arr) => Value::Integer(arr.lock().unwrap().len() as i64),
                    Value::Map(map) => Value::Integer(map.lock().unwrap().len() as i64),
                    _ => Value::Integer(0),
                }
            })
        }));

        self.env.set("upper".to_string(), Value::Builtin("upper".to_string(), |args, _| {
             Box::pin(async move {
                 if args.len() != 1 { return Value::Null; }
                 match &args[0] {
                     Value::String(s) => Value::String(s.to_uppercase()),
                     _ => Value::Null,
                 }
             })
        }));

        self.env.set("str".to_string(), Value::Builtin("str".to_string(), |args, _| {
            Box::pin(async move {
                if args.len() != 1 { return Value::Null; }
                Value::String(format!("{}", args[0]))
            })
        }));

        self.env.set("push".to_string(), Value::Builtin("push".to_string(), |args, _| {
            Box::pin(async move {
                if args.len() != 2 { return Value::Null; }
                if let Value::Array(arr) = &args[0] {
                    arr.lock().unwrap().push(args[1].clone());
                    return Value::Array(arr.clone());
                }
                Value::Null
            })
        }));

        self.env.set("http_get".to_string(), Value::Builtin("http_get".to_string(), |args, _| {
            Box::pin(async move {
                if args.len() != 1 { return Value::Null; }
                let url = match &args[0] {
                    Value::String(s) => s.clone(),
                    _ => return Value::Null,
                };

                match reqwest::get(&url).await {
                    Ok(resp) => {
                         match resp.text().await {
                             Ok(text) => Value::String(text),
                             Err(_) => Value::Null
                         }
                    },
                    Err(e) => {
                        eprintln!("HTTP Error: {}", e);
                        Value::Null
                    }
                }
            })
        }));

        self.env.set("http_post".to_string(), Value::Builtin("http_post".to_string(), |args, _| {
            Box::pin(async move {
                if args.len() != 2 { return Value::Null; }
                let url = match &args[0] {
                    Value::String(s) => s.clone(),
                    _ => return Value::Null,
                };
                let body = match &args[1] {
                    Value::String(s) => s.clone(),
                    _ => return Value::Null,
                };

                let client = reqwest::Client::new();
                match client.post(&url).body(body).send().await {
                    Ok(resp) => {
                         match resp.text().await {
                             Ok(text) => Value::String(text),
                             Err(_) => Value::Null
                         }
                    },
                    Err(e) => {
                        eprintln!("HTTP Error: {}", e);
                        Value::Null
                    }
                }
            })
        }));

        self.env.set("db_execute".to_string(), Value::Builtin("db_execute".to_string(), |args, pool| {
            Box::pin(async move {
                if pool.is_none() {
                    eprintln!("DB Error: No database pool configured");
                    return Value::Null;
                }
                let pool = pool.unwrap();

                if args.len() < 1 { return Value::Null; }
                let sql = match &args[0] {
                    Value::String(s) => s.clone(),
                    _ => return Value::Null,
                };

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
            })
        }));

        self.env.set("db_query".to_string(), Value::Builtin("db_query".to_string(), |args, pool| {
            Box::pin(async move {
                if pool.is_none() {
                    eprintln!("DB Error: No database pool configured");
                    return Value::Null;
                }
                let pool = pool.unwrap();

                if args.len() < 1 { return Value::Null; }
                let sql = match &args[0] {
                    Value::String(s) => s.clone(),
                    _ => return Value::Null,
                };

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
                                let val = if let Ok(v) = row.try_get::<i64, _>(name) {
                                    Value::Integer(v)
                                } else if let Ok(v) = row.try_get::<String, _>(name) {
                                    Value::String(v)
                                } else if let Ok(v) = row.try_get::<bool, _>(name) {
                                    Value::Boolean(v)
                                } else {
                                    Value::Null
                                };
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
            })
        }));

        // VIEW / RENDER
        self.env.set("view".to_string(), Value::Builtin("view".to_string(), |args, pool| {
            Box::pin(async move {
                if args.len() != 2 { return Value::Null; }
                let tpl = match &args[0] {
                    Value::String(s) => s.clone(),
                    _ => return Value::Null,
                };
                let data = args[1].clone();

                // Create a temporary evaluator to render
                let mut renderer = Evaluator::new(pool);

                // Inject data
                if let Value::Map(map) = data {
                    let m = map.lock().unwrap();
                    for (k, v) in m.iter() {
                        renderer.env.set(k.clone(), v.clone());
                    }
                }

                let mut parser = ZenoBladeParser::new(&tpl);
                let nodes = parser.parse();

                let output = renderer.render_nodes(nodes).await;
                Value::String(output)
            })
        }));
    }

    // Helper for rendering blade nodes
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
            }
        }
        output
    }

    // ... (rest of Evaluator implementation: eval, eval_statement, etc.)
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
                Some(func_ptr(args, self.db_pool.clone()).await)
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
