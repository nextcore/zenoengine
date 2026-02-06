use crate::parser::{Statement, Expression, Op};
use std::collections::HashMap;
use std::rc::Rc;
use std::cell::RefCell;

#[derive(Debug, Clone)]
pub enum Value {
    Integer(i64),
    String(String),
    Boolean(bool),
    Null,
    // Parameters, Body, Closure Environment (captured)
    Function(Vec<String>, Statement, Env),
    // Builtin Function: Name, Function Pointer
    Builtin(String, fn(Vec<Value>) -> Value),
    ReturnValue(Box<Value>), // Wrapper to signal return
    // Array: Mutable shared list
    Array(Rc<RefCell<Vec<Value>>>),
    // Map: Mutable shared key-value store
    Map(Rc<RefCell<HashMap<String, Value>>>),
}

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
                let vec_a = a.borrow();
                let vec_b = b.borrow();
                vec_a.iter().zip(vec_b.iter()).all(|(x, y)| x == y) && vec_a.len() == vec_b.len()
            },
            (Value::Map(a), Value::Map(b)) => {
                let map_a = a.borrow();
                let map_b = b.borrow();
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
                let vec = arr.borrow();
                let elements: Vec<String> = vec.iter().map(|v| format!("{}", v)).collect();
                write!(f, "[{}]", elements.join(", "))
            },
            Value::Map(map) => {
                let m = map.borrow();
                let mut entries: Vec<String> = m.iter().map(|(k, v)| format!("\"{}\": {}", k, v)).collect();
                entries.sort(); // Sort for deterministic output
                write!(f, "{{{}}}", entries.join(", "))
            }
        }
    }
}

#[derive(Debug, Clone, PartialEq)]
pub struct Env {
    store: Rc<RefCell<HashMap<String, Value>>>,
    outer: Option<Box<Env>>,
}

impl Env {
    pub fn new() -> Self {
        Self {
            store: Rc::new(RefCell::new(HashMap::new())),
            outer: None,
        }
    }

    pub fn new_with_outer(outer: Env) -> Self {
        Self {
            store: Rc::new(RefCell::new(HashMap::new())),
            outer: Some(Box::new(outer)),
        }
    }

    pub fn get(&self, name: &str) -> Option<Value> {
        if let Some(val) = self.store.borrow().get(name) {
            return Some(val.clone());
        }
        if let Some(ref outer) = self.outer {
            return outer.get(name);
        }
        None
    }

    pub fn set(&mut self, name: String, val: Value) {
        // Simple set logic: always set in current scope (like `let`)?
        // Or update if exists?
        // Our `Let` statement uses `set` on current scope.
        // Assignment logic needs to check existence.
        self.store.borrow_mut().insert(name, val);
    }

    pub fn update(&mut self, name: &str, val: Value) -> bool {
         if self.store.borrow().contains_key(name) {
             self.store.borrow_mut().insert(name.to_string(), val);
             return true;
         }
         if let Some(ref mut outer) = self.outer {
             return outer.update(name, val);
         }
         false
    }
}

pub struct Evaluator {
    env: Env,
    output: String,
}

impl Evaluator {
    pub fn new() -> Self {
        let mut evaluator = Self {
            env: Env::new(),
            output: String::new(),
        };
        evaluator.register_builtins();
        evaluator
    }

    pub fn get_output(&self) -> String {
        self.output.clone()
    }

    fn register_builtins(&mut self) {
        self.env.set("len".to_string(), Value::Builtin("len".to_string(), |args| {
            if args.len() != 1 { return Value::Null; }
            match &args[0] {
                Value::String(s) => Value::Integer(s.len() as i64),
                Value::Array(arr) => Value::Integer(arr.borrow().len() as i64),
                Value::Map(map) => Value::Integer(map.borrow().len() as i64),
                _ => Value::Integer(0),
            }
        }));

        self.env.set("upper".to_string(), Value::Builtin("upper".to_string(), |args| {
             if args.len() != 1 { return Value::Null; }
             match &args[0] {
                 Value::String(s) => Value::String(s.to_uppercase()),
                 _ => Value::Null,
             }
        }));

        self.env.set("str".to_string(), Value::Builtin("str".to_string(), |args| {
            if args.len() != 1 { return Value::Null; }
            Value::String(format!("{}", args[0]))
        }));

        self.env.set("push".to_string(), Value::Builtin("push".to_string(), |args| {
            if args.len() != 2 { return Value::Null; }
            if let Value::Array(arr) = &args[0] {
                arr.borrow_mut().push(args[1].clone());
                return Value::Array(arr.clone());
            }
            Value::Null
        }));

        // HTTP Client Built-ins
        self.env.set("http_get".to_string(), Value::Builtin("http_get".to_string(), |args| {
            if args.len() != 1 { return Value::Null; }
            let url = match &args[0] {
                Value::String(s) => s,
                _ => return Value::Null,
            };

            match reqwest::blocking::get(url) {
                Ok(resp) => {
                     match resp.text() {
                         Ok(text) => Value::String(text),
                         Err(_) => Value::Null
                     }
                },
                Err(e) => {
                    eprintln!("HTTP Error: {}", e);
                    Value::Null
                }
            }
        }));

        self.env.set("http_post".to_string(), Value::Builtin("http_post".to_string(), |args| {
            if args.len() != 2 { return Value::Null; }
            let url = match &args[0] {
                Value::String(s) => s,
                _ => return Value::Null,
            };
            let body = match &args[1] {
                Value::String(s) => s.clone(),
                _ => return Value::Null,
            };

            let client = reqwest::blocking::Client::new();
            match client.post(url).body(body).send() {
                Ok(resp) => {
                     match resp.text() {
                         Ok(text) => Value::String(text),
                         Err(_) => Value::Null
                     }
                },
                Err(e) => {
                    eprintln!("HTTP Error: {}", e);
                    Value::Null
                }
            }
        }));
    }

    pub fn eval(&mut self, statements: Vec<Statement>) {
        for stmt in statements {
            let result = self.eval_statement(stmt);
            if let Some(Value::ReturnValue(_)) = result {
                 // Top level return
            }
        }
    }

    fn eval_statement(&mut self, stmt: Statement) -> Option<Value> {
        match stmt {
            Statement::Print(expr) => {
                let val = self.eval_expression(&expr)?;
                let out = format!("{}\n", val);
                self.output.push_str(&out);
                print!("{}", out); // Still print to stdout for CLI feedback
                None
            }
            Statement::Let(name, expr) => {
                let val = self.eval_expression(&expr)?;
                self.env.set(name, val);
                None
            }
             Statement::Assign(lhs, rhs) => {
                let val = self.eval_expression(&rhs)?;
                match lhs {
                    Expression::Identifier(name) => {
                        if !self.env.update(&name, val.clone()) {
                             eprintln!("Runtime Error: Variable '{}' not declared before assignment", name);
                             return None;
                        }
                    },
                    Expression::Index(target_expr, index_expr) => {
                         let target = self.eval_expression(&target_expr)?;
                         let index = self.eval_expression(&index_expr)?;

                         match (target, index) {
                             (Value::Array(arr), Value::Integer(i)) => {
                                 let mut vec = arr.borrow_mut();
                                 if i >= 0 && (i as usize) < vec.len() {
                                     vec[i as usize] = val;
                                 } else {
                                     eprintln!("Runtime Error: Index out of bounds");
                                 }
                             },
                             (Value::Map(map), Value::String(key)) => {
                                 let mut m = map.borrow_mut();
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
                let val = self.eval_expression(&expr)?;
                Some(Value::ReturnValue(Box::new(val)))
            }
            Statement::Block(stmts) => {
                self.eval_block(stmts)
            }
            Statement::If(condition, consequence, alternative) => {
                let cond_val = self.eval_expression(&condition)?;
                if self.is_truthy(cond_val) {
                    self.eval_statement(*consequence)
                } else if let Some(alt) = alternative {
                    self.eval_statement(*alt)
                } else {
                    None
                }
            }
            Statement::Expression(expr) => {
                self.eval_expression(&expr)
            }
        }
    }

    fn eval_block(&mut self, stmts: Vec<Statement>) -> Option<Value> {
        let previous_env = self.env.clone();
        self.env = Env::new_with_outer(previous_env.clone());

        let mut result = None;
        for stmt in stmts {
            let val = self.eval_statement(stmt);
            if let Some(Value::ReturnValue(_)) = val {
                result = val;
                break;
            }
            result = val;
        }

        self.env = previous_env;
        result
    }

    fn eval_expression(&mut self, expr: &Expression) -> Option<Value> {
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
                    vals.push(self.eval_expression(e)?);
                }
                Some(Value::Array(Rc::new(RefCell::new(vals))))
            },
             Expression::Map(pairs) => {
                let mut map = HashMap::new();
                for (k, v_expr) in pairs {
                    map.insert(k.clone(), self.eval_expression(v_expr)?);
                }
                Some(Value::Map(Rc::new(RefCell::new(map))))
            },
            Expression::Index(left, index) => {
                 let left_val = self.eval_expression(left)?;
                 let index_val = self.eval_expression(index)?;

                 match (left_val, index_val) {
                     (Value::Array(arr), Value::Integer(i)) => {
                         let vec = arr.borrow();
                         if i >= 0 && (i as usize) < vec.len() {
                             Some(vec[i as usize].clone())
                         } else {
                             eprintln!("Runtime Error: Index out of bounds");
                             Some(Value::Null)
                         }
                     },
                     (Value::Map(map), Value::String(key)) => {
                         let m = map.borrow();
                         m.get(&key).cloned().or(Some(Value::Null))
                     },
                     _ => {
                         eprintln!("Runtime Error: Index operation not supported on this type");
                         None
                     }
                 }
            },
            Expression::BinaryOp(lhs, op, rhs) => {
                let l_val = self.eval_expression(lhs)?;
                let r_val = self.eval_expression(rhs)?;
                self.eval_infix_expression(op, l_val, r_val)
            },
            Expression::Call(func_expr, args) => {
                let func = self.eval_expression(func_expr)?;
                let mut evaluated_args = Vec::new();
                for arg in args {
                    evaluated_args.push(self.eval_expression(arg)?);
                }
                self.apply_function(func, evaluated_args)
            }
        }
    }

    fn apply_function(&mut self, func: Value, args: Vec<Value>) -> Option<Value> {
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
                let result = self.eval_statement(body);
                self.env = previous_env;

                if let Some(Value::ReturnValue(val)) = result {
                    Some(*val)
                } else {
                    Some(Value::Null)
                }
            },
            Value::Builtin(_, func_ptr) => {
                Some(func_ptr(args))
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
    use std::rc::Rc;
    use std::cell::RefCell;

    #[test]
    fn test_functions() {
        let input = r#"
            fn add(a, b) {
                return a + b;
            }
            let res = add(5, 5);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);

        let val = evaluator.env.get("res");
        if let Some(Value::Integer(v)) = val {
            assert_eq!(v, 10);
        } else {
            panic!("Function return value not correct, got {:?}", val);
        }
    }

     #[test]
    fn test_closure() {
        let input = r#"
            let factor = 2;
            fn multiply(a) {
                return a * factor;
            }
            let res = multiply(10);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);

        let val = evaluator.env.get("res");
        if let Some(Value::Integer(v)) = val {
            assert_eq!(v, 20);
        } else {
            panic!("Closure capture failed");
        }
    }

    #[test]
    fn test_builtin() {
        let input = r#"
            let l = len("hello");
            let u = upper("hello");
            let s = str(123);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);

        assert_eq!(evaluator.env.get("l"), Some(Value::Integer(5)));
        assert_eq!(evaluator.env.get("u"), Some(Value::String("HELLO".to_string())));
        assert_eq!(evaluator.env.get("s"), Some(Value::String("123".to_string())));
    }

    #[test]
    fn test_arrays() {
        let input = r#"
            let arr = [1, 2, 3];
            let first = arr[0];
            push(arr, 4);
            let length = len(arr);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);

        assert_eq!(evaluator.env.get("first"), Some(Value::Integer(1)));
        assert_eq!(evaluator.env.get("length"), Some(Value::Integer(4)));

        // Check array content manually
        if let Some(Value::Array(arr)) = evaluator.env.get("arr") {
             let vec = arr.borrow();
             assert_eq!(vec.len(), 4);
             assert_eq!(vec[3], Value::Integer(4));
        } else {
            panic!("arr not found or not array");
        }
    }

    #[test]
    fn test_maps() {
        let input = r#"
            let user = {"name": "Zeno", "id": 1};
            let n = user["name"];
            user["id"] = 99;
            let i = user["id"];
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);

        assert_eq!(evaluator.env.get("n"), Some(Value::String("Zeno".to_string())));
        assert_eq!(evaluator.env.get("i"), Some(Value::Integer(99)));
    }
}
