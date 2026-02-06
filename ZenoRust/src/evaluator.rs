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
}

// Manual PartialEq to ignore function pointer comparison
impl PartialEq for Value {
    fn eq(&self, other: &Self) -> bool {
        match (self, other) {
            (Value::Integer(a), Value::Integer(b)) => a == b,
            (Value::String(a), Value::String(b)) => a == b,
            (Value::Boolean(a), Value::Boolean(b)) => a == b,
            (Value::Null, Value::Null) => true,
            (Value::Function(params_a, body_a, _), Value::Function(params_b, body_b, _)) => {
                params_a == params_b && body_a == body_b
                // We ignore Env comparison for simplicity/cycles
            },
            (Value::Builtin(name_a, _), Value::Builtin(name_b, _)) => name_a == name_b,
            (Value::ReturnValue(a), Value::ReturnValue(b)) => a == b,
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
        }
    }
}

// Environment needs to be shared and mutable for closures/recursion
// We use Rc<RefCell<...>> for the chain
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
        self.store.borrow_mut().insert(name, val);
    }
}

pub struct Evaluator {
    env: Env,
}

impl Evaluator {
    pub fn new() -> Self {
        let mut evaluator = Self {
            env: Env::new(),
        };
        evaluator.register_builtins();
        evaluator
    }

    fn register_builtins(&mut self) {
        self.env.set("len".to_string(), Value::Builtin("len".to_string(), |args| {
            if args.len() != 1 {
                return Value::Null; // Should be error, but keeping simple
            }
            match &args[0] {
                Value::String(s) => Value::Integer(s.len() as i64),
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
    }

    pub fn eval(&mut self, statements: Vec<Statement>) {
        for stmt in statements {
            let result = self.eval_statement(stmt);
            if let Some(Value::ReturnValue(_)) = result {
                 // Top level return? Usually not allowed or just ignored/printed
                 // For now, we assume scripts don't return from top level
            }
        }
    }

    fn eval_statement(&mut self, stmt: Statement) -> Option<Value> {
        match stmt {
            Statement::Print(expr) => {
                let val = self.eval_expression(&expr)?;
                println!("{}", val);
                None
            }
            Statement::Let(name, expr) => {
                let val = self.eval_expression(&expr)?;
                self.env.set(name, val);
                None
            }
            Statement::Function(name, params, body) => {
                // Capture current env for closure
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
        // Blocks share the SAME env in our simplified previous implementation,
        // OR create a new one. To support `let` shadowing properly within blocks:
        // We should create a new inner env.

        let previous_env = self.env.clone();
        self.env = Env::new_with_outer(previous_env.clone());

        let mut result = None;
        for stmt in stmts {
            let val = self.eval_statement(stmt);
            if let Some(Value::ReturnValue(_)) = val {
                result = val;
                break;
            }
            result = val; // keep last expression result if we support that
        }

        self.env = previous_env; // restore
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

                // Create environment for execution: Closure Env + Args
                let mut extended_env = Env::new_with_outer(closure_env);
                for (param_name, arg_val) in params.iter().zip(args) {
                    extended_env.set(param_name.clone(), arg_val);
                }

                // Swap current env with extended env
                let previous_env = self.env.clone();
                self.env = extended_env;

                // Execute body (it's a Statement, usually Block, but could be others)
                // Note: body is stored as *Statement inside Value::Function?
                // We need to execute it.
                // Wait, Value::Function(..., Statement, ...) -> Statement is NOT Copy/Clone cheaply if complex.
                // We derived Clone on Statement, so we can clone it.
                let result = self.eval_statement(body);

                // Restore env
                self.env = previous_env;

                // Unwrap ReturnValue if present
                if let Some(Value::ReturnValue(val)) = result {
                    Some(*val)
                } else {
                    Some(Value::Null) // Function returns null if no return statement
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
}
