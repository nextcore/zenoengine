use crate::parser::{Statement, Expression, Op};
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub enum Value {
    Integer(i64),
    String(String),
}

impl std::fmt::Display for Value {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Value::Integer(i) => write!(f, "{}", i),
            Value::String(s) => write!(f, "{}", s),
        }
    }
}

pub struct Evaluator {
    env: HashMap<String, Value>,
}

impl Evaluator {
    pub fn new() -> Self {
        Self {
            env: HashMap::new(),
        }
    }

    pub fn eval(&mut self, statements: Vec<Statement>) {
        for stmt in statements {
            match stmt {
                Statement::Print(expr) => {
                    if let Some(val) = self.eval_expression(&expr) {
                        println!("{}", val);
                    } else {
                         eprintln!("Runtime Error: Could not evaluate expression in print statement");
                    }
                }
                Statement::Let(name, expr) => {
                    if let Some(val) = self.eval_expression(&expr) {
                        self.env.insert(name, val);
                    } else {
                        eprintln!("Runtime Error: Could not evaluate expression in let statement");
                    }
                }
            }
        }
    }

    fn eval_expression(&self, expr: &Expression) -> Option<Value> {
        match expr {
            Expression::Integer(i) => Some(Value::Integer(*i)),
            Expression::StringLiteral(s) => Some(Value::String(s.clone())),
            Expression::Identifier(name) => self.env.get(name).cloned(),
            Expression::BinaryOp(lhs, op, rhs) => {
                let l_val = self.eval_expression(lhs)?;
                let r_val = self.eval_expression(rhs)?;

                match (l_val, op, r_val) {
                    (Value::Integer(l), Op::Add, Value::Integer(r)) => Some(Value::Integer(l + r)),
                    (Value::Integer(l), Op::Subtract, Value::Integer(r)) => Some(Value::Integer(l - r)),
                    (Value::Integer(l), Op::Multiply, Value::Integer(r)) => Some(Value::Integer(l * r)),
                    (Value::Integer(l), Op::Divide, Value::Integer(r)) => {
                        if r == 0 {
                            eprintln!("Runtime Error: Division by zero");
                            None
                        } else {
                            Some(Value::Integer(l / r))
                        }
                    },
                    (Value::String(l), Op::Add, Value::String(r)) => Some(Value::String(l + &r)),
                    (Value::String(l), Op::Add, Value::Integer(r)) => Some(Value::String(format!("{}{}", l, r))),
                    (Value::Integer(l), Op::Add, Value::String(r)) => Some(Value::String(format!("{}{}", l, r))),
                    _ => {
                        eprintln!("Runtime Error: Mismatched types in binary operation");
                        None
                    }
                }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use crate::parser::Parser;
    use crate::evaluator::{Evaluator, Value};

    #[test]
    fn test_print_math() {
        let input = r#"print(10 + 20 * 2);"#; // 10 + 40 = 50
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        assert_eq!(statements.len(), 1);

        // We can't easily capture stdout in this simple test harness without redirection,
        // but we can verify execution doesn't panic.
        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);
    }

    #[test]
    fn test_variable_assignment() {
        let input = r#"
            let x = 10;
            let y = 5;
            print(x + y);
        "#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
        assert_eq!(statements.len(), 3);

        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);

        // Verify internal state
        if let Some(Value::Integer(v)) = evaluator.env.get("x") {
            assert_eq!(*v, 10);
        } else {
            panic!("Variable x not found or not integer");
        }
    }
}
