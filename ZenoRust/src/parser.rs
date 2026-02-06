use crate::lexer::Token;
use logos::{Lexer, Logos};
use std::iter::Peekable;

#[derive(Debug, Clone, PartialEq)]
pub enum Expression {
    StringLiteral(String),
    Integer(i64),
    Boolean(bool),
    Identifier(String),
    Array(Vec<Expression>),
    Map(Vec<(String, Expression)>), // Key (String), Value
    Index(Box<Expression>, Box<Expression>), // Target, Index
    BinaryOp(Box<Expression>, Op, Box<Expression>),
    Call(Box<Expression>, Vec<Expression>), // FuncName, Args
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

#[derive(Debug, Clone, PartialEq)]
pub enum Statement {
    Print(Expression),
    Let(String, Expression),
    Assign(Expression, Expression), // Target (LHS), Value (RHS)
    Block(Vec<Statement>),
    If(Expression, Box<Statement>, Option<Box<Statement>>),
    Function(String, Vec<String>, Box<Statement>), // Name, Params, Body
    Return(Expression),
    Expression(Expression),
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
                    Token::Fn => {
                        self.lexer.next();
                        if let Some(stmt) = self.parse_function() { statements.push(stmt); }
                    },
                    Token::Return => {
                        self.lexer.next();
                        if let Some(stmt) = self.parse_return() { statements.push(stmt); }
                    },
                    Token::Print => {
                        self.lexer.next();
                        if let Some(stmt) = self.parse_print() { statements.push(stmt); }
                    }
                    Token::Let => {
                        self.lexer.next();
                        if let Some(stmt) = self.parse_let() { statements.push(stmt); }
                    }
                    Token::If => {
                        self.lexer.next();
                        if let Some(stmt) = self.parse_if() { statements.push(stmt); }
                    }
                    Token::LBrace => {
                        // Ambiguity: Block or Map?
                        // At top level (Statement), `{` starts a Block.
                        // Maps are Expressions.
                        // So here it's definitely a Block.
                        if let Some(stmt) = self.parse_block() { statements.push(stmt); }
                    }
                    Token::RBrace => { break; }
                    _ => {
                        // Statement that starts with expression (Assignment or Call)
                        if let Some(expr) = self.parse_expression(0) {
                            // Check for Assignment (=)
                             if let Some(Ok(Token::Equals)) = self.lexer.peek() {
                                 self.lexer.next(); // consume '='
                                 let rhs = match self.parse_expression(0) { Some(e) => e, None => break }; // Error handling simplified
                                 if let Some(Ok(Token::Semicolon)) = self.lexer.peek() { self.lexer.next(); }
                                 statements.push(Statement::Assign(expr, rhs));
                             } else {
                                 // Expression statement
                                 if let Some(Ok(Token::Semicolon)) = self.lexer.peek() { self.lexer.next(); }
                                 statements.push(Statement::Expression(expr));
                             }
                        } else {
                            self.lexer.next(); // Skip unknown
                        }
                    }
                }
            } else {
                self.lexer.next();
            }
        }
        statements
    }

    fn parse_function(&mut self) -> Option<Statement> {
        let name = match self.lexer.next() { Some(Ok(Token::Identifier(s))) => s, _ => return None };
        match self.lexer.next() { Some(Ok(Token::LParen)) => {}, _ => return None }

        let mut params = Vec::new();
        if let Some(Ok(Token::RParen)) = self.lexer.peek() {
            self.lexer.next();
        } else {
            loop {
                match self.lexer.next() { Some(Ok(Token::Identifier(s))) => params.push(s), _ => return None };
                match self.lexer.peek() {
                    Some(Ok(Token::Comma)) => { self.lexer.next(); },
                    Some(Ok(Token::RParen)) => { self.lexer.next(); break; },
                    _ => return None
                }
            }
        }

        let body = self.parse_block()?;
        Some(Statement::Function(name, params, Box::new(body)))
    }

    fn parse_return(&mut self) -> Option<Statement> {
        let expr = self.parse_expression(0)?;
        if let Some(Ok(Token::Semicolon)) = self.lexer.peek() { self.lexer.next(); }
        Some(Statement::Return(expr))
    }

    fn parse_block(&mut self) -> Option<Statement> {
        match self.lexer.next() { Some(Ok(Token::LBrace)) => {}, _ => return None }
        let mut statements = Vec::new();
        while let Some(token_result) = self.lexer.peek() {
            if let Ok(Token::RBrace) = token_result {
                self.lexer.next(); return Some(Statement::Block(statements));
            }
            if let Ok(token) = token_result {
                 match token {
                    Token::Fn => { self.lexer.next(); if let Some(stmt) = self.parse_function() { statements.push(stmt); } },
                    Token::Return => { self.lexer.next(); if let Some(stmt) = self.parse_return() { statements.push(stmt); } },
                    Token::Print => { self.lexer.next(); if let Some(stmt) = self.parse_print() { statements.push(stmt); } },
                    Token::Let => { self.lexer.next(); if let Some(stmt) = self.parse_let() { statements.push(stmt); } },
                    Token::If => { self.lexer.next(); if let Some(stmt) = self.parse_if() { statements.push(stmt); } },
                    Token::LBrace => { if let Some(stmt) = self.parse_block() { statements.push(stmt); } },
                    _ => {
                         // Expression or Assignment
                         if let Some(expr) = self.parse_expression(0) {
                             if let Some(Ok(Token::Equals)) = self.lexer.peek() {
                                 self.lexer.next();
                                 let rhs = match self.parse_expression(0) { Some(e) => e, None => break };
                                 if let Some(Ok(Token::Semicolon)) = self.lexer.peek() { self.lexer.next(); }
                                 statements.push(Statement::Assign(expr, rhs));
                             } else {
                                if let Some(Ok(Token::Semicolon)) = self.lexer.peek() { self.lexer.next(); }
                                statements.push(Statement::Expression(expr));
                             }
                         } else { self.lexer.next(); }
                    }
                }
            } else { self.lexer.next(); }
        }
        None
    }

    fn parse_if(&mut self) -> Option<Statement> {
        match self.lexer.next() { Some(Ok(Token::LParen)) => {}, _ => return None }
        let condition = self.parse_expression(0)?;
        match self.lexer.next() { Some(Ok(Token::RParen)) => {}, _ => return None }
        let consequence = if let Some(Ok(Token::LBrace)) = self.lexer.peek() { self.parse_block()? } else { return None; };
        let mut alternative = None;
        if let Some(Ok(Token::Else)) = self.lexer.peek() {
            self.lexer.next();
            if let Some(Ok(Token::If)) = self.lexer.peek() {
                 self.lexer.next(); alternative = Some(Box::new(self.parse_if()?));
            } else if let Some(Ok(Token::LBrace)) = self.lexer.peek() {
                 alternative = Some(Box::new(self.parse_block()?));
            } else { return None; }
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

    pub fn parse_expression_entry(&mut self) -> Option<Expression> {
        self.parse_expression(0)
    }

    fn parse_expression(&mut self, min_bp: u8) -> Option<Expression> {
        let mut lhs = match self.lexer.next() {
            Some(Ok(Token::Integer(Some(i)))) => Expression::Integer(i),
            Some(Ok(Token::True)) => Expression::Boolean(true),
            Some(Ok(Token::False)) => Expression::Boolean(false),
            Some(Ok(Token::StringLiteral(s))) => Expression::StringLiteral(s),
            Some(Ok(Token::Identifier(s))) => Expression::Identifier(s),
            Some(Ok(Token::LBracket)) => {
                let mut elements = Vec::new();
                if let Some(Ok(Token::RBracket)) = self.lexer.peek() {
                    self.lexer.next();
                } else {
                    loop {
                        elements.push(self.parse_expression(0)?);
                        match self.lexer.peek() {
                            Some(Ok(Token::Comma)) => { self.lexer.next(); },
                            Some(Ok(Token::RBracket)) => { self.lexer.next(); break; },
                            _ => return None
                        }
                    }
                }
                Expression::Array(elements)
            },
            Some(Ok(Token::LBrace)) => { // Map Literal {"key": val}
                let mut pairs = Vec::new();
                if let Some(Ok(Token::RBrace)) = self.lexer.peek() {
                    self.lexer.next();
                } else {
                    loop {
                        // Key must be string literal (for now)
                        let key = match self.lexer.next() {
                            Some(Ok(Token::StringLiteral(s))) => s,
                            _ => return None
                        };
                        match self.lexer.next() { Some(Ok(Token::Colon)) => {}, _ => return None }
                        let value = self.parse_expression(0)?;
                        pairs.push((key, value));

                        match self.lexer.peek() {
                            Some(Ok(Token::Comma)) => { self.lexer.next(); },
                            Some(Ok(Token::RBrace)) => { self.lexer.next(); break; },
                            _ => return None
                        }
                    }
                }
                Expression::Map(pairs)
            },
            Some(Ok(Token::LParen)) => {
                let expr = self.parse_expression(0)?;
                match self.lexer.next() { Some(Ok(Token::RParen)) => expr, _ => return None }
            }
            _ => return None,
        };

        loop {
            // Handle Indexing: `arr[index]`
            if let Some(Ok(Token::LBracket)) = self.lexer.peek() {
                 let index_bp = 10;
                 if index_bp < min_bp { break; }

                 self.lexer.next();
                 let index_expr = self.parse_expression(0)?;
                 match self.lexer.next() { Some(Ok(Token::RBracket)) => {}, _ => return None }

                 lhs = Expression::Index(Box::new(lhs), Box::new(index_expr));
                 continue;
            }

            // Handle Call
            if let Some(Ok(Token::LParen)) = self.lexer.peek() {
                let call_bp = 10;
                if call_bp < min_bp { break; }

                self.lexer.next();
                let mut args = Vec::new();
                if let Some(Ok(Token::RParen)) = self.lexer.peek() {
                    self.lexer.next();
                } else {
                    loop {
                        args.push(self.parse_expression(0)?);
                        match self.lexer.peek() {
                            Some(Ok(Token::Comma)) => { self.lexer.next(); },
                            Some(Ok(Token::RParen)) => { self.lexer.next(); break; },
                            _ => return None
                        }
                    }
                }
                lhs = Expression::Call(Box::new(lhs), args);
                continue;
            }

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
