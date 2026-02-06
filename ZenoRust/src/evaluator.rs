use crate::parser::Statement;

pub struct Evaluator;

impl Evaluator {
    pub fn new() -> Self {
        Self
    }

    pub fn eval(&self, statements: Vec<Statement>) {
        for stmt in statements {
            match stmt {
                Statement::Print(content) => {
                    println!("{}", content);
                }
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use crate::parser::Parser;
    use crate::evaluator::Evaluator;

    #[test]
    fn test_print_statement() {
        let input = r#"print("Hello Test");"#;
        let mut parser = Parser::new(input);
        let statements = parser.parse();

        assert_eq!(statements.len(), 1);

        let evaluator = Evaluator::new();
        evaluator.eval(statements);
    }
}
