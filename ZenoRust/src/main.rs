pub mod lexer;
pub mod parser;
pub mod evaluator;

use std::fs;
use parser::Parser;
use evaluator::Evaluator;

fn main() {
    println!("ZenoEngine Rust Edition (2024) Initialized!");

    let file_path = "source/test.zl";
    println!("Executing ZenoLang script: {}", file_path);

    let contents = fs::read_to_string(file_path)
        .expect("Should have been able to read the file");

    let mut parser = Parser::new(&contents);
    let statements = parser.parse();

    let mut evaluator = Evaluator::new();
    evaluator.eval(statements);
}
