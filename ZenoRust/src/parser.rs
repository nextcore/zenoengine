use crate::lexer::Token;
use logos::{Lexer, Logos};
use std::iter::Peekable;

#[derive(Debug, Clone, PartialEq)]
pub enum Expression {
    StringLiteral(String),
    Integer(i64),
    Boolean(bool),
    Identifier(String),
    BinaryOp(Box<Expression>, Op, Box<Expression>),
}

#[derive(Debug, Clone, PartialEq)]
pub enum Op {
    Add,
    Subtract,
    Multiply,
    Divide,
    Equal,
    NotEqual,
    LessThan,
    GreaterThan,
}

#[derive(Debug)]
pub enum Statement {
    Print(Expression),
    Let(String, Expression),
    Block(Vec<Statement>),
    If(Expression, Box<Statement>, Option<Box<Statement>>),
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
                        self.lexer.next();
                        if let Some(stmt) = self.parse_print() {
                            statements.push(stmt);
                        }
                    }
                    Token::Let => {
                        self.lexer.next();
                        if let Some(stmt) = self.parse_let() {
                            statements.push(stmt);
                        }
                    }
                    Token::If => {
                        self.lexer.next();
                        if let Some(stmt) = self.parse_if() {
                            statements.push(stmt);
                        }
                    }
                    Token::LBrace => {
                        if let Some(stmt) = self.parse_block() {
                            statements.push(stmt);
                        }
                    }
                    Token::RBrace => {
                        // Handled by parse_block, but if seen here it might be an extra brace
                        // For now, we stop parsing this block (or top level) if we see },
                        // but since parse() is top level, this usually means error or end of block logic
                        // We'll break loop if we see unexpected RBrace at top level so we don't infinite loop if caller didn't consume it
                        break;
                    }
                    _ => {
                        self.lexer.next();
                    }
                }
            } else {
                self.lexer.next();
            }
        }
        statements
    }

    fn parse_block(&mut self) -> Option<Statement> {
        // Expect '{'
        match self.lexer.next() {
            Some(Ok(Token::LBrace)) => {}
            _ => return None,
        }

        let mut statements = Vec::new();
        while let Some(token_result) = self.lexer.peek() {
            if let Ok(Token::RBrace) = token_result {
                self.lexer.next(); // consume '}'
                return Some(Statement::Block(statements));
            }

            // Parse statement
             if let Ok(token) = token_result {
                match token {
                    Token::Print => {
                         self.lexer.next();
                         if let Some(stmt) = self.parse_print() { statements.push(stmt); }
                    },
                    Token::Let => {
                         self.lexer.next();
                         if let Some(stmt) = self.parse_let() { statements.push(stmt); }
                    },
                     Token::If => {
                        self.lexer.next();
                        if let Some(stmt) = self.parse_if() { statements.push(stmt); }
                    }
                    Token::LBrace => {
                         if let Some(stmt) = self.parse_block() { statements.push(stmt); }
                    },
                    _ => { self.lexer.next(); } // skip unknown inside block
                }
             } else {
                 self.lexer.next();
             }
        }

        None // Missing closing brace
    }

    fn parse_if(&mut self) -> Option<Statement> {
        // Expect '('
        match self.lexer.next() {
            Some(Ok(Token::LParen)) => {}
            _ => return None,
        }

        let condition = self.parse_expression(0)?;

        // Expect ')'
        match self.lexer.next() {
            Some(Ok(Token::RParen)) => {}
            _ => return None,
        }

        // Parse Consequence (usually a Block)
        // Check if next is LBrace
        let consequence = if let Some(Ok(Token::LBrace)) = self.lexer.peek() {
             self.parse_block()?
        } else {
            // Single statement if support? Let's just enforce blocks for now or single statement
            // Ideally should support any statement.
             // For simplicity, let's just reuse top-level parse logic for one statement if we wanted,
             // but `parse_block` is standard for `if`. Let's assume block required for now.
             return None;
        };

        // Check for Else
        let mut alternative = None;
        if let Some(Ok(Token::Else)) = self.lexer.peek() {
            self.lexer.next(); // consume else

            // Check if next is 'if' (else if) or block
            if let Some(Ok(Token::If)) = self.lexer.peek() {
                 self.lexer.next();
                 alternative = Some(Box::new(self.parse_if()?));
            } else if let Some(Ok(Token::LBrace)) = self.lexer.peek() {
                 alternative = Some(Box::new(self.parse_block()?));
            } else {
                return None;
            }
        }

        Some(Statement::If(condition, Box::new(consequence), alternative))
    }

    fn parse_print(&mut self) -> Option<Statement> {
        match self.lexer.next() { Some(Ok(Token::LParen)) => {}, _ => return None }
        let expr = self.parse_expression(0)?;
        match self.lexer.next() { Some(Ok(Token::RParen)) => {}, _ => return None }
        if let Some(Ok(Token::Semicolon)) = self.lexer.peek() { self.lexer.next(); }
        Some(Statement::Print(expr))
    }

    fn parse_let(&mut self) -> Option<Statement> {
        let name = match self.lexer.next() { Some(Ok(Token::Identifier(s))) => s, _ => return None };
        match self.lexer.next() { Some(Ok(Token::Equals)) => {}, _ => return None }
        let expr = self.parse_expression(0)?;
        if let Some(Ok(Token::Semicolon)) = self.lexer.peek() { self.lexer.next(); }
        Some(Statement::Let(name, expr))
    }

    fn parse_expression(&mut self, min_bp: u8) -> Option<Expression> {
        let mut lhs = match self.lexer.next() {
            Some(Ok(Token::Integer(Some(i)))) => Expression::Integer(i),
            Some(Ok(Token::True)) => Expression::Boolean(true),
            Some(Ok(Token::False)) => Expression::Boolean(false),
            Some(Ok(Token::StringLiteral(s))) => Expression::StringLiteral(s),
            Some(Ok(Token::Identifier(s))) => Expression::Identifier(s),
            Some(Ok(Token::LParen)) => {
                let expr = self.parse_expression(0)?;
                match self.lexer.next() { Some(Ok(Token::RParen)) => expr, _ => return None }
            }
            _ => return None,
        };

        loop {
            let op = match self.lexer.peek() {
                Some(Ok(Token::Plus)) => Op::Add,
                Some(Ok(Token::Minus)) => Op::Subtract,
                Some(Ok(Token::Star)) => Op::Multiply,
                Some(Ok(Token::Slash)) => Op::Divide,
                Some(Ok(Token::DoubleEquals)) => Op::Equal,
                Some(Ok(Token::NotEquals)) => Op::NotEqual,
                Some(Ok(Token::LessThan)) => Op::LessThan,
                Some(Ok(Token::GreaterThan)) => Op::GreaterThan,
                _ => break,
            };

            let (l_bp, r_bp) = self.infix_binding_power(&op);
            if l_bp < min_bp { break; }

            self.lexer.next();
            let rhs = self.parse_expression(r_bp)?;
            lhs = Expression::BinaryOp(Box::new(lhs), op, Box::new(rhs));
        }
        Some(lhs)
    }

    fn infix_binding_power(&self, op: &Op) -> (u8, u8) {
        match op {
            Op::Equal | Op::NotEqual | Op::LessThan | Op::GreaterThan => (1, 2),
            Op::Add | Op::Subtract => (3, 4),
            Op::Multiply | Op::Divide => (5, 6),
        }
    }
}
