use crate::parser::{Expression, Parser};
use crate::evaluator::{Evaluator, Value, Env};
use std::sync::{Arc, Mutex};
use std::collections::HashMap;

#[derive(Debug, Clone)]
pub enum BladeNode {
    Text(String),
    Interpolation(Expression), // {{ expr }}
    If(Expression, Vec<BladeNode>, Option<Vec<BladeNode>>), // @if(expr) ... @else ... @endif
    ForEach(Expression, String, Vec<BladeNode>), // @foreach(collection as item) ... @endforeach
    Include(String), // @include('path')
    Extends(String), // @extends('path')
    Section(String, Vec<BladeNode>), // @section('name') ... @endsection
    Yield(String), // @yield('name')
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
            } else if self.peek_str("@foreach") {
                if let Some(node) = self.parse_foreach() {
                    nodes.push(node);
                }
            } else if self.peek_str("@include") {
                if let Some(node) = self.parse_include() {
                    nodes.push(node);
                }
            } else if self.peek_str("@extends") {
                if let Some(node) = self.parse_extends() {
                    nodes.push(node);
                }
            } else if self.peek_str("@section") {
                if let Some(node) = self.parse_section() {
                    nodes.push(node);
                }
            } else if self.peek_str("@yield") {
                if let Some(node) = self.parse_yield() {
                    nodes.push(node);
                }
            } else if self.peek_str("@else") || self.peek_str("@endif") || self.peek_str("@endforeach") || self.peek_str("@endsection") {
                if inside_block {
                    break;
                } else {
                    self.pos += 1; // Stray tag
                }
            } else {
                if self.pos < self.input.len() {
                     if !self.peek_str("{{") && !self.peek_str("@") {
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
            if self.peek_str("{{") || self.peek_str("@") {
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

    fn parse_foreach(&mut self) -> Option<BladeNode> {
        self.pos += 8; // Skip @foreach
        // Expect (collection as item)
        if let Some(start) = self.input[self.pos..].find('(') {
             self.pos += start + 1;

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
                 let content_str = &self.input[self.pos..self.pos+end];
                 self.pos += end + 1;

                 // Split by " as "
                 if let Some(pos_as) = content_str.find(" as ") {
                     let collection_str = &content_str[..pos_as];
                     let item_str = &content_str[pos_as+4..].trim(); // Extract identifier

                     let mut parser = Parser::new(collection_str);
                     if let Some(collection_expr) = parser.parse_expression_entry() {
                         let loop_block = self.parse_nodes(true);
                         if self.peek_str("@endforeach") {
                             self.pos += 11;
                         }
                         return Some(BladeNode::ForEach(collection_expr, item_str.to_string(), loop_block));
                     }
                 }
             }
        }
        None
    }

    fn parse_include(&mut self) -> Option<BladeNode> {
        self.pos += 8; // @include
        let path = self.parse_string_arg()?;
        Some(BladeNode::Include(path))
    }

    fn parse_extends(&mut self) -> Option<BladeNode> {
        self.pos += 8; // @extends
        let path = self.parse_string_arg()?;
        Some(BladeNode::Extends(path))
    }

    fn parse_yield(&mut self) -> Option<BladeNode> {
        self.pos += 6; // @yield
        let name = self.parse_string_arg()?;
        Some(BladeNode::Yield(name))
    }

    fn parse_section(&mut self) -> Option<BladeNode> {
        self.pos += 8; // @section
        let name = self.parse_string_arg()?;

        let block = self.parse_nodes(true);

        if self.peek_str("@endsection") {
            self.pos += 11;
        }
        Some(BladeNode::Section(name, block))
    }

    fn parse_string_arg(&mut self) -> Option<String> {
        if let Some(start) = self.input[self.pos..].find('(') {
            self.pos += start + 1;
            // Handle quotes
            let remainder = &self.input[self.pos..];
            if remainder.starts_with('\'') {
                if let Some(end) = remainder[1..].find('\'') {
                    let s = remainder[1..1+end].to_string();
                    self.pos += end + 2; // skip ' and '
                    // skip closing paren if present
                    if self.pos < self.input.len() && self.input[self.pos..].starts_with(')') {
                        self.pos += 1;
                    }
                    return Some(s);
                }
            } else if remainder.starts_with('"') {
                if let Some(end) = remainder[1..].find('"') {
                    let s = remainder[1..1+end].to_string();
                    self.pos += end + 2;
                    if self.pos < self.input.len() && self.input[self.pos..].starts_with(')') {
                        self.pos += 1;
                    }
                    return Some(s);
                }
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
