use crate::lexer::Token;
use logos::{Lexer, Logos};
use std::iter::Peekable;

#[derive(Debug, Clone, PartialEq)]
pub enum Expression {
    StringLiteral(String),
    Integer(i64),
    Identifier(String),
    BinaryOp(Box<Expression>, Op, Box<Expression>),
}

#[derive(Debug, Clone, PartialEq)]
pub enum Op {
    Add,
    Subtract,
    Multiply,
    Divide,
}

#[derive(Debug)]
pub enum Statement {
    Print(Expression),
    Let(String, Expression),
}

pub struct Parser<'a> {
    lexer: Peekable<Lexer<'a, Token>>,
}

impl<'a> Parser<'a> {
    pub fn new(input: &'a str) -> Self {
        Self {
            lexer: Token::lexer(input).peekable(),
        }
    }

    pub fn parse(&mut self) -> Vec<Statement> {
        let mut statements = Vec::new();
        while let Some(token_result) = self.lexer.peek() {
            if let Ok(token) = token_result {
                match token {
                    Token::Print => {
                        self.lexer.next(); // consume 'print'
                        if let Some(stmt) = self.parse_print() {
                            statements.push(stmt);
                        }
                    }
                    Token::Let => {
                        self.lexer.next(); // consume 'let'
                        if let Some(stmt) = self.parse_let() {
                            statements.push(stmt);
                        }
                    }
                    _ => {
                        // Skip unknown or unexpected tokens at statement level
                        self.lexer.next();
                    }
                }
            } else {
                // Lexer error, skip
                self.lexer.next();
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

        // Expect Expression
        let expr = self.parse_expression(0)?;

        // Expect ')'
        match self.lexer.next() {
            Some(Ok(Token::RParen)) => {}
            _ => return None,
        }

        // Optional ';'
        if let Some(Ok(Token::Semicolon)) = self.lexer.peek() {
            self.lexer.next();
        }

        Some(Statement::Print(expr))
    }

    fn parse_let(&mut self) -> Option<Statement> {
        // Expect Identifier
        let name = match self.lexer.next() {
            Some(Ok(Token::Identifier(s))) => s,
            _ => return None,
        };

        // Expect '='
        match self.lexer.next() {
            Some(Ok(Token::Equals)) => {}
            _ => return None,
        }

        // Expect Expression
        let expr = self.parse_expression(0)?;

        // Optional ';'
        if let Some(Ok(Token::Semicolon)) = self.lexer.peek() {
            self.lexer.next();
        }

        Some(Statement::Let(name, expr))
    }

    // Pratt Parser for expressions
    fn parse_expression(&mut self, min_bp: u8) -> Option<Expression> {
        let mut lhs = match self.lexer.next() {
            Some(Ok(Token::Integer(Some(i)))) => Expression::Integer(i),
            Some(Ok(Token::StringLiteral(s))) => Expression::StringLiteral(s),
            Some(Ok(Token::Identifier(s))) => Expression::Identifier(s),
            Some(Ok(Token::LParen)) => {
                let expr = self.parse_expression(0)?;
                match self.lexer.next() {
                    Some(Ok(Token::RParen)) => expr,
                    _ => return None,
                }
            }
            _ => return None,
        };

        loop {
            let op = match self.lexer.peek() {
                Some(Ok(Token::Plus)) => Op::Add,
                Some(Ok(Token::Minus)) => Op::Subtract,
                Some(Ok(Token::Star)) => Op::Multiply,
                Some(Ok(Token::Slash)) => Op::Divide,
                _ => break,
            };

            let (l_bp, r_bp) = self.infix_binding_power(&op);
            if l_bp < min_bp {
                break;
            }

            self.lexer.next(); // consume op
            let rhs = self.parse_expression(r_bp)?;
            lhs = Expression::BinaryOp(Box::new(lhs), op, Box::new(rhs));
        }

        Some(lhs)
    }

    fn infix_binding_power(&self, op: &Op) -> (u8, u8) {
        match op {
            Op::Add | Op::Subtract => (1, 2),
            Op::Multiply | Op::Divide => (3, 4),
        }
    }
}
