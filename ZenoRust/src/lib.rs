pub mod lexer;
pub mod parser;
pub mod evaluator;
pub mod template;
pub mod middleware;
pub mod builtins;
pub mod db_builder;
pub mod plugins {
    pub mod sidecar;
    pub mod wasm;
}
