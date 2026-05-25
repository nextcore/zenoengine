use std::fmt;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Diagnostic {
    pub r#type: String, // "error", "warning", "panic"
    pub message: String,
    pub filename: String,
    pub line: usize,
    pub col: usize,
    pub slot: Option<String>,
}

impl std::error::Error for Diagnostic {}

impl fmt::Display for Diagnostic {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if !self.filename.is_empty() {
            if let Some(ref slot) = self.slot {
                write!(
                    f,
                    "[{}:{}:{}] {} (slot: {})",
                    self.filename, self.line, self.col, self.message, slot
                )
            } else {
                write!(
                    f,
                    "[{}:{}:{}] {}",
                    self.filename, self.line, self.col, self.message
                )
            }
        } else {
            write!(f, "{}", self.message)
        }
    }
}
