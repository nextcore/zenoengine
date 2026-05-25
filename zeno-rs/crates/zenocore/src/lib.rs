pub mod diagnostic;
pub mod lexer;
pub mod parser;
pub mod scope;
pub mod executor;
pub mod slots;
pub mod analysis;

pub use diagnostic::Diagnostic;
pub use lexer::{Lexer, Token, TokenType};
pub use parser::{Node, parse_string, parse_file};
pub use scope::{Scope, Value};
pub use executor::{Context, Engine, InputMeta, SlotMeta, HandlerFn};
pub use slots::{register_logic_slots, register_collection_slots};
pub use analysis::{Analyzer, AnalysisResult};
