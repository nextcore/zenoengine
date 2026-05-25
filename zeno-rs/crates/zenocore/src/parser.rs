use crate::diagnostic::Diagnostic;
use crate::lexer::{Lexer, TokenType};
use std::fs;
use std::path::Path;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Node {
    pub name: String,
    pub value: Option<String>,
    pub children: Vec<Node>,
    pub line: usize,
    pub col: usize,
    pub filename: String,
}

struct ParserNode {
    name: String,
    value: Option<String>,
    children: Vec<usize>,
    #[allow(dead_code)]
    parent: Option<usize>,
    line: usize,
    col: usize,
}

pub fn parse_string(content: &str, filename: &str) -> Result<Node, Diagnostic> {
    let mut l = Lexer::new(content);

    let mut nodes = vec![ParserNode {
        name: "root".to_string(),
        value: None,
        children: Vec::new(),
        parent: None,
        line: 1,
        col: 1,
    }];

    let mut stack = vec![0];
    let mut last_node_idx: Option<usize> = None;

    loop {
        let tok = l.next_token();
        if tok.r#type == TokenType::EOF {
            break;
        }

        match tok.r#type {
            TokenType::Identifier => {
                let node_idx = nodes.len();
                let parent_idx = *stack.last().unwrap();

                nodes.push(ParserNode {
                    name: tok.literal,
                    value: None,
                    children: Vec::new(),
                    parent: Some(parent_idx),
                    line: tok.line,
                    col: tok.col,
                });

                nodes[parent_idx].children.push(node_idx);
                last_node_idx = Some(node_idx);
            }

            TokenType::Colon => {
                let current_line = tok.line;
                let mut value_parts = Vec::new();

                loop {
                    let peek = l.peek_token();
                    if peek.r#type == TokenType::EOF
                        || peek.line != current_line
                        || peek.r#type == TokenType::LBrace
                        || peek.r#type == TokenType::Colon
                    {
                        break;
                    }
                    if peek.r#type == TokenType::RBrace {
                        break;
                    }

                    let t = l.next_token();
                    value_parts.push(t.literal);
                }

                if !value_parts.is_empty() {
                    if let Some(idx) = last_node_idx {
                        nodes[idx].value = Some(value_parts.join(" "));
                    }
                }

                let peek = l.peek_token();
                if peek.r#type == TokenType::LBrace {
                    l.next_token(); // consume {
                    if let Some(idx) = last_node_idx {
                        stack.push(idx);
                    }
                } else if peek.r#type == TokenType::RBrace {
                    l.next_token(); // consume }
                    if stack.len() > 1 {
                        stack.pop();
                    }
                }
            }

            TokenType::LBrace => {
                if let Some(idx) = last_node_idx {
                    stack.push(idx);
                } else {
                    // Anonymous node
                    let node_idx = nodes.len();
                    let parent_idx = *stack.last().unwrap();

                    nodes.push(ParserNode {
                        name: String::new(),
                        value: None,
                        children: Vec::new(),
                        parent: Some(parent_idx),
                        line: tok.line,
                        col: tok.col,
                    });

                    nodes[parent_idx].children.push(node_idx);
                    stack.push(node_idx);
                }
            }

            TokenType::RBrace => {
                if stack.len() > 1 {
                    stack.pop();
                }
            }

            TokenType::Error => {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: format!("lexical error: unexpected character '{}'", tok.literal),
                    filename: filename.to_string(),
                    line: tok.line,
                    col: tok.col,
                    slot: None,
                });
            }
            _ => {}
        }
    }

    fn to_nested(nodes: &[ParserNode], idx: usize, filename: &str) -> Node {
        let pn = &nodes[idx];
        let children = pn
            .children
            .iter()
            .map(|&c_idx| to_nested(nodes, c_idx, filename))
            .collect();

        Node {
            name: pn.name.clone(),
            value: pn.value.clone(),
            children,
            line: pn.line,
            col: pn.col,
            filename: filename.to_string(),
        }
    }

    Ok(to_nested(&nodes, 0, filename))
}

pub fn parse_file<P: AsRef<Path>>(path: P) -> Result<Node, Diagnostic> {
    let path_ref = path.as_ref();
    let filename = path_ref.to_string_lossy().into_owned();
    let content = fs::read_to_string(path_ref).map_err(|e| Diagnostic {
        r#type: "error".to_string(),
        message: format!("failed to read file: {}", e),
        filename: filename.clone(),
        line: 1,
        col: 1,
        slot: None,
    })?;

    parse_string(&content, &filename)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_simple() {
        let code = r#"
            log: "🚀 Started"
            http.get: '/' {
                view: 'welcome'
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        assert_eq!(root.name, "root");
        assert_eq!(root.children.len(), 2);

        let log_node = &root.children[0];
        assert_eq!(log_node.name, "log");
        assert_eq!(log_node.value, Some("🚀 Started".to_string()));
        assert!(log_node.children.is_empty());

        let get_node = &root.children[1];
        assert_eq!(get_node.name, "http.get");
        assert_eq!(get_node.value, Some("/".to_string()));
        assert_eq!(get_node.children.len(), 1);

        let view_node = &get_node.children[0];
        assert_eq!(view_node.name, "view");
        assert_eq!(view_node.value, Some("welcome".to_string()));
    }
}
