use crate::lexer::Token;
use logos::{Lexer, Logos};

#[derive(Debug)]
pub enum Statement {
    Print(String),
}

pub struct Parser<'a> {
    lexer: Lexer<'a, Token>,
}

impl<'a> Parser<'a> {
    pub fn new(input: &'a str) -> Self {
        Self {
            lexer: Token::lexer(input),
        }
    }

    pub fn parse(&mut self) -> Vec<Statement> {
        let mut statements = Vec::new();
        while let Some(token_result) = self.lexer.next() {
            match token_result {
                Ok(Token::Print) => {
                    if let Some(stmt) = self.parse_print() {
                        statements.push(stmt);
                    }
                }
                _ => {} // Skip unexpected tokens for now
            }
        }
        statements
    }

    fn parse_print(&mut self) -> Option<Statement> {
        // Expect '('
        match self.lexer.next() {
            Some(Ok(Token::LParen)) => {}
            _ => return None,
        }

        // Expect String
        let content = match self.lexer.next() {
            Some(Ok(Token::StringLiteral(s))) => s,
            _ => return None,
        };

        // Expect ')'
        match self.lexer.next() {
            Some(Ok(Token::RParen)) => {}
            _ => return None,
        }

        // Optional ';'
        // We peek or just consume if it exists, for now let's leniently handle it
        // But Logos iterator consumes. We'll check next token.
        // If it's semicolon, consume it. If it's something else, we might need to put it back?
        // Recursive descent usually needs peekable. For now, simple assumption: script ends or next stmt.

        Some(Statement::Print(content))
    }
}
