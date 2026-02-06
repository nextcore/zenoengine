use crate::parser::{Statement, Expression, Op};
use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq)]
pub enum Value {
    Integer(i64),
    String(String),
    Boolean(bool),
    Null,
}

impl std::fmt::Display for Value {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Value::Integer(i) => write!(f, "{}", i),
            Value::String(s) => write!(f, "{}", s),
            Value::Boolean(b) => write!(f, "{}", b),
            Value::Null => write!(f, "null"),
        }
    }
}

pub struct Evaluator {
    // Vector of environments to handle scoping (stack)
    // envs[0] is global, envs.last() is current local scope
    envs: Vec<HashMap<String, Value>>,
}

impl Evaluator {
    pub fn new() -> Self {
        Self {
            envs: vec![HashMap::new()],
        }
    }

    pub fn eval(&mut self, statements: Vec<Statement>) {
        for stmt in statements {
            self.eval_statement(stmt);
        }
    }

    fn eval_statement(&mut self, stmt: Statement) {
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
                    // Insert into current (top) scope
                    if let Some(current_scope) = self.envs.last_mut() {
                        current_scope.insert(name, val);
                    }
                } else {
                    eprintln!("Runtime Error: Could not evaluate expression in let statement");
                }
            }
            Statement::Block(stmts) => {
                self.enter_scope();
                for s in stmts {
                    self.eval_statement(s);
                }
                self.exit_scope();
            }
            Statement::If(condition, consequence, alternative) => {
                let cond_val = self.eval_expression(&condition);
                if self.is_truthy(cond_val) {
                    self.eval_statement(*consequence);
                } else if let Some(alt) = alternative {
                    self.eval_statement(*alt);
                }
            }
        }
    }

    fn eval_expression(&self, expr: &Expression) -> Option<Value> {
        match expr {
            Expression::Integer(i) => Some(Value::Integer(*i)),
            Expression::StringLiteral(s) => Some(Value::String(s.clone())),
            Expression::Boolean(b) => Some(Value::Boolean(*b)),
            Expression::Identifier(name) => self.lookup_variable(name),
            Expression::BinaryOp(lhs, op, rhs) => {
                let l_val = self.eval_expression(lhs)?;
                let r_val = self.eval_expression(rhs)?;
                self.eval_infix_expression(op, l_val, r_val)
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

    fn enter_scope(&mut self) {
        self.envs.push(HashMap::new());
    }

    fn exit_scope(&mut self) {
        self.envs.pop();
    }

    fn lookup_variable(&self, name: &str) -> Option<Value> {
        // Look from inner-most scope to outer-most
        for env in self.envs.iter().rev() {
            if let Some(val) = env.get(name) {
                return Some(val.clone());
            }
        }
        eprintln!("Runtime Error: Variable '{}' not found", name);
        None
    }

    fn is_truthy(&self, val: Option<Value>) -> bool {
        match val {
            Some(Value::Boolean(b)) => b,
            Some(Value::Integer(i)) => i != 0,
            Some(Value::String(s)) => !s.is_empty(),
            Some(Value::Null) => false,
            None => false,
        }
    }
}

#[cfg(test)]
mod tests {
    use crate::parser::Parser;
    use crate::evaluator::{Evaluator, Value};

    #[test]
    fn test_print_math() {
        let input = r#"print(10 + 20 * 2);"#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();
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
        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);

        let val = evaluator.lookup_variable("x");
        if let Some(Value::Integer(v)) = val {
            assert_eq!(v, 10);
        } else {
            panic!("Variable x not found");
        }
    }

    #[test]
    fn test_if_else() {
        let input = r#"
            let result = 0;
            if (10 > 5) {
                let result = 100;
            } else {
                let result = 0;
            }
        "#;
        // Note: 'let result' inside block creates a NEW variable in that scope (shadowing),
        // it does NOT update the outer 'result'.
        // Our current implementation is "Let" always declares. We don't have Re-assignment (=) yet.
        // So this test is tricky. We can check if code runs.

        let mut parser = Parser::new(input);
        let statements = parser.parse();
        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);
    }
}
