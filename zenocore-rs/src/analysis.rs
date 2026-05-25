use crate::diagnostic::Diagnostic;
use crate::executor::Engine;
use crate::parser::{Node, parse_file};
use std::collections::HashSet;
use std::path::Path;

#[derive(Debug, Clone, Default)]
pub struct AnalysisResult {
    pub errors: Vec<Diagnostic>,
    pub warnings: Vec<Diagnostic>,
}

pub struct Analyzer<'a> {
    engine: &'a Engine,
    visited: HashSet<String>,
}

impl<'a> Analyzer<'a> {
    pub fn new(engine: &'a Engine) -> Self {
        Self {
            engine,
            visited: HashSet::new(),
        }
    }

    pub fn analyze(&mut self, root: &Node) -> AnalysisResult {
        let mut res = AnalysisResult::default();
        self.visited.clear();
        self.walk(root, None, &mut res);
        res
    }

    fn walk(&mut self, node: &Node, parent_name: Option<&str>, res: &mut AnalysisResult) {
        // A. Include Analysis
        if node.name == "include" {
            let mut path = String::new();
            if let Some(ref val) = node.value {
                path = val.trim_matches(|c| c == '"' || c == '\'' || c == '`').to_string();
            }

            if !path.is_empty() {
                if self.visited.contains(&path) {
                    // Cyclic dependency detected
                    res.warnings.push(Diagnostic {
                        r#type: "warning".to_string(),
                        message: format!("static warning: cyclic inclusion of '{}' detected", path),
                        filename: node.filename.clone(),
                        line: node.line,
                        col: node.col,
                        slot: Some("include".to_string()),
                    });
                    return;
                }
                self.visited.insert(path.clone());

                let path_obj = Path::new(&path);
                match parse_file(path_obj) {
                    Ok(included_root) => {
                        self.walk(&included_root, None, res);
                    }
                    Err(diag) => {
                        res.errors.push(diag);
                    }
                }
            }
            return;
        }

        // Skip root node logic, just recurse
        if node.name == "root" {
            for child in &node.children {
                self.walk(child, Some("root"), res);
            }
            return;
        }

        // B. Slot Validation
        if !node.name.is_empty() && !node.name.starts_with('$') {
            if let Some(meta) = self.engine.docs.get(&node.name) {
                // 1. Check Main Value Type
                if !meta.value_type.is_empty() && meta.value_type != "any" {
                    if let Some(ref val_str) = node.value {
                        if !val_str.is_empty() && !val_str.starts_with('$') && !val_str.contains("??") {
                            let mut dummy = node.clone();
                            dummy.children = Vec::new();
                            let parsed_val = self.engine.resolve_shorthand_value(&dummy, &crate::scope::Scope::new(None));
                            if let Err(err) = self.engine.validate_value_type(&parsed_val, &meta.value_type, node, &node.name) {
                                res.errors.push(err);
                            }
                        }
                    }
                }

                // 2. Check Required Attributes
                let mut first_required = String::new();
                for (name, input) in &meta.inputs {
                    if input.required {
                        first_required = name.clone();
                        break;
                    }
                }

                for (name, input) in &meta.inputs {
                    if input.required {
                        let mut found = false;
                        for child in &node.children {
                            if child.name == *name {
                                found = true;
                                break;
                            }
                        }

                        // Positional Value Satisfaction
                        let is_common_positional = name == "id" || name == "spreadsheet_id" || name == "path" || name == "name" || name == "url" || name == "file";
                        if !found && (is_common_positional || *name == first_required) && node.value.is_some() {
                            if let Some(ref val_str) = node.value {
                                if !val_str.is_empty() {
                                    found = true;
                                }
                            }
                        }

                        if !found {
                            res.errors.push(Diagnostic {
                                r#type: "error".to_string(),
                                message: format!("static error: missing required attribute '{}' for slot '{}'", name, node.name),
                                filename: node.filename.clone(),
                                line: node.line,
                                col: node.col,
                                slot: Some(node.name.clone()),
                            });
                        }
                    }
                }

                // 3. Check Required Blocks
                for block_name in &meta.required_blocks {
                    let mut found = false;
                    for child in &node.children {
                        if child.name == *block_name {
                            found = true;
                            break;
                        }
                    }
                    if !found {
                        res.errors.push(Diagnostic {
                            r#type: "error".to_string(),
                            message: format!("static error: missing required block '{}:' for slot '{}'", block_name, node.name),
                            filename: node.filename.clone(),
                            line: node.line,
                            col: node.col,
                            slot: Some(node.name.clone()),
                        });
                    }
                }

                // 4. Check Type for Constant Values
                for child in &node.children {
                    if let Some(input) = meta.inputs.get(&child.name) {
                        if !input.r#type.is_empty() && input.r#type != "any" {
                            if let Some(ref val_str) = child.value {
                                if !val_str.is_empty() && !val_str.starts_with('$') && !val_str.contains("??") {
                                    let parsed_val = self.engine.resolve_shorthand_value(child, &crate::scope::Scope::new(None));
                                    if let Err(err) = self.engine.validate_value_type(&parsed_val, &input.r#type, child, &node.name) {
                                        res.errors.push(err);
                                    }
                                }
                            }
                        }
                    }
                }

                // 5. Recurse nested slots (Only children that are NOT attributes)
                for child in &node.children {
                    if !meta.inputs.contains_key(&child.name) {
                        self.walk(child, Some(&node.name), res);
                    }
                }
                return;
            } else {
                // If not registered, check if it's a known keyword in the executor/slots
                let keywords = [
                    "if", "for", "foreach", "while", "switch", "try", "do", "then", "else", "catch",
                    "as", "rules", "rules_map", "data", "break", "continue", "return", "scope.set", "var",
                    "logic.compare", "dump", "dd", "isset", "empty", "unless", "auth", "guest",
                    "auth.user", "auth.check", "can", "cannot", "json", "forelse", "error"
                ];

                if !keywords.contains(&node.name.as_str()) {
                    let mut is_call_arg = false;
                    if let Some(ref p_name) = parent_name {
                        if let Some(p_meta) = self.engine.docs.get(*p_name) {
                            if p_meta.inputs.is_empty() {
                                is_call_arg = true;
                            }
                        }
                        if *p_name == "call" || *p_name == "data" || *p_name == "array.pop" || *p_name == "date.now" || *p_name == "system.env" || *p_name == "coalesce" {
                            is_call_arg = true;
                        }
                    }

                    if !is_call_arg {
                        res.errors.push(Diagnostic {
                            r#type: "error".to_string(),
                            message: format!("static error: unknown slot '{}'", node.name),
                            filename: node.filename.clone(),
                            line: node.line,
                            col: node.col,
                            slot: Some(node.name.clone()),
                        });
                    }
                }
            }
        }

        // Default recursion
        for child in &node.children {
            self.walk(child, Some(&node.name), res);
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::parser::parse_string;
    use crate::slots::register_logic_slots;

    #[test]
    fn test_analyzer_unknown_slot() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);

        let root = parse_string("unknown_slot_name: 123", "test.zl").unwrap();
        let mut analyzer = Analyzer::new(&engine);
        let res = analyzer.analyze(&root);

        assert!(!res.errors.is_empty());
        assert!(res.errors[0].message.contains("unknown slot 'unknown_slot_name'"));
    }

    #[test]
    fn test_analyzer_cyclic_includes() {
        let engine = Engine::new();
        
        let path_a = std::env::current_dir().unwrap().join("test_a.zl");
        let path_b = std::env::current_dir().unwrap().join("test_b.zl");
        
        std::fs::write(&path_a, "include: 'test_b.zl'").unwrap();
        std::fs::write(&path_b, "include: 'test_a.zl'").unwrap();
        
        let root = parse_file(&path_a).unwrap();
        let mut analyzer = Analyzer::new(&engine);
        let res = analyzer.analyze(&root);
        
        // Clean up temp files
        let _ = std::fs::remove_file(&path_a);
        let _ = std::fs::remove_file(&path_b);
        
        assert!(!res.warnings.is_empty());
        assert!(res.warnings[0].message.contains("cyclic inclusion"));
    }
}
