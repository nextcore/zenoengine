use crate::parser::{Expression, Parser};
use crate::evaluator::{Evaluator, Value, Env};
use std::sync::{Arc, Mutex};
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub enum BladeNode {
    Text(String),
    Interpolation(Expression), // {{ expr }}
    If(Expression, Vec<BladeNode>, Option<Vec<BladeNode>>), // @if(expr) ... @else ... @endif
}

pub struct ZenoBladeParser<'a> {
    input: &'a str,
    pos: usize,
}

impl<'a> ZenoBladeParser<'a> {
    pub fn new(input: &'a str) -> Self {
        Self { input, pos: 0 }
    }

    pub fn parse(&mut self) -> Vec<BladeNode> {
        self.parse_nodes(false)
    }

    fn parse_nodes(&mut self, inside_block: bool) -> Vec<BladeNode> {
        let mut nodes = Vec::new();
        while self.pos < self.input.len() {
            if let Some(text) = self.parse_text() {
                nodes.push(BladeNode::Text(text));
            }

            if self.peek_str("{{") {
                if let Some(node) = self.parse_interpolation() {
                    nodes.push(node);
                }
            } else if self.peek_str("@if") {
                if let Some(node) = self.parse_if() {
                    nodes.push(node);
                }
            } else if self.peek_str("@else") || self.peek_str("@endif") {
                if inside_block {
                    break;
                } else {
                    self.pos += 1; // Stray tag
                }
            } else {
                if self.pos < self.input.len() {
                     // Should be handled by parse_text unless it's a specific char
                     // break loop to avoid infinite if parse_text returns None but we are not at end
                     // Actually parse_text consumes until tag. If we are at tag, we handle it.
                     // If we are at EOF, loop ends.
                     // If we are at unknown tag, parse_text should have consumed it? No, parse_text stops AT tag.
                     // So if we are here, we are at a tag start.
                     // If it's none of the above, skip 1 char.
                     if !self.peek_str("{{") && !self.peek_str("@if") && !self.peek_str("@else") && !self.peek_str("@endif") {
                         self.pos += 1;
                     }
                }
            }
        }
        nodes
    }

    fn parse_text(&mut self) -> Option<String> {
        let start = self.pos;
        while self.pos < self.input.len() {
            if self.peek_str("{{") || self.peek_str("@if") || self.peek_str("@else") || self.peek_str("@endif") {
                break;
            }
            self.pos += 1;
        }
        if self.pos > start {
            Some(self.input[start..self.pos].to_string())
        } else {
            None
        }
    }

    fn peek_str(&self, s: &str) -> bool {
        self.input[self.pos..].starts_with(s)
    }

    fn parse_interpolation(&mut self) -> Option<BladeNode> {
        self.pos += 2; // Skip {{
        // Find closing }}
        if let Some(end) = self.input[self.pos..].find("}}") {
            let expr_str = &self.input[self.pos..self.pos+end];
            let mut parser = Parser::new(expr_str);
            if let Some(expr) = parser.parse_expression_entry() {
                self.pos += end + 2;
                return Some(BladeNode::Interpolation(expr));
            }
        }
        None
    }

    fn parse_if(&mut self) -> Option<BladeNode> {
        self.pos += 3; // Skip @if
        // Expect (condition)
        if let Some(start) = self.input[self.pos..].find('(') {
             self.pos += start + 1;

             // Simplification: Find closing ')' by counting parens
             let mut depth = 1;
             let mut end = 0;
             for (i, c) in self.input[self.pos..].char_indices() {
                 if c == '(' { depth += 1; }
                 else if c == ')' { depth -= 1; }
                 if depth == 0 {
                     end = i;
                     break;
                 }
             }

             if end > 0 {
                 let expr_str = &self.input[self.pos..self.pos+end];
                 self.pos += end + 1;

                 let mut parser = Parser::new(expr_str);
                 let condition = parser.parse_expression_entry()?;

                 let true_block = self.parse_nodes(true);

                 let mut false_block = None;
                 if self.peek_str("@else") {
                     self.pos += 5;
                     false_block = Some(self.parse_nodes(true));
                 }

                 if self.peek_str("@endif") {
                     self.pos += 6;
                 }

                 return Some(BladeNode::If(condition, true_block, false_block));
             }
        }
        None
    }
}

pub struct ZenoBlade {
    evaluator: Evaluator,
}

impl ZenoBlade {
    pub fn render(template: &str, data: Value) -> String {
        // Create new evaluator for rendering scope
        let mut evaluator = Evaluator::new(None);

        // Inject data into env
        if let Value::Map(map) = data {
            let m = map.lock().unwrap();
            for (k, v) in m.iter() {
                evaluator.env.set(k.clone(), v.clone());
            }
        }

        let mut parser = ZenoBladeParser::new(template);
        let nodes = parser.parse();

        // We need an async render?
        // Our Evaluator is async.
        // But `render` built-in wrapper will be async.
        // We need internal `render_nodes`.

        // Since we are inside `evaluator.rs` usually, we can't easily spawn another evaluator if we want to share?
        // Actually, we should reuse the evaluator or create a new one.
        // But Evaluator::eval is async.
        // We can't block here easily inside synchronous function signature of `ZenoBlade::render`?
        // Wait, `ZenoBlade::render` should be called from `Builtin` which returns `BoxFuture`.
        // So `ZenoBlade::render` can be async!

        // However, I defined `render` above as sync. Let's fix.
        String::new() // Placeholder, logic moved to Evaluator builtin
    }
}
