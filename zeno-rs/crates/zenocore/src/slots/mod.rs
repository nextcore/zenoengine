pub mod logic;
pub mod collection;
pub mod utility;

use crate::executor::Engine;
use crate::parser::Node;
use crate::scope::{Scope, Value};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

pub fn register_logic_slots(engine: &mut Engine) {
    logic::register(engine);
    register_collection_slots(engine);
}

pub fn register_collection_slots(engine: &mut Engine) {
    collection::register(engine);
    utility::register(engine);
}

// ==========================================
// SHARED HELPER FUNCTIONS
// ==========================================

pub fn resolve_node_value(engine: &Engine, node: &Node, scope: &Arc<Scope>) -> Value {
    let temp_node = Node {
        name: node.name.clone(),
        value: node.value.clone(),
        children: Vec::new(),
        line: node.line,
        col: node.col,
        filename: node.filename.clone(),
    };
    engine.resolve_shorthand_value(&temp_node, scope)
}

pub(crate) fn eval_simple_condition(expr: &str, scope: &Arc<Scope>) -> bool {
    let ops = ["<=", ">=", "==", "!=", "<", ">"];
    let mut found_op = None;
    for op in &ops {
        if expr.contains(op) {
            found_op = Some(*op);
            break;
        }
    }

    let op = match found_op {
        Some(o) => o,
        None => return false,
    };

    let parts: Vec<&str> = expr.splitn(2, op).collect();
    if parts.len() != 2 {
        return false;
    }

    let left = parts[0].trim();
    let right = parts[1].trim();

    let left_val = resolve_expression_value(left, scope);
    let right_val = resolve_expression_value(right, scope);

    match op {
        "<" => left_val.to_float() < right_val.to_float(),
        ">" => left_val.to_float() > right_val.to_float(),
        "<=" => left_val.to_float() <= right_val.to_float(),
        ">=" => left_val.to_float() >= right_val.to_float(),
        "==" => left_val.to_string_coerce() == right_val.to_string_coerce(),
        "!=" => left_val.to_string_coerce() != right_val.to_string_coerce(),
        _ => false,
    }
}

pub(crate) fn resolve_expression_value(s: &str, scope: &Arc<Scope>) -> Value {
    if s.starts_with('$') {
        let key = &s[1..];
        return scope.get(key).unwrap_or(Value::Nil);
    }
    if s == "true" {
        return Value::Bool(true);
    }
    if s == "false" {
        return Value::Bool(false);
    }
    if s == "null" || s == "nil" {
        return Value::Nil;
    }
    if let Ok(i) = s.parse::<i64>() {
        return Value::Int(i);
    }
    if let Ok(f) = s.parse::<f64>() {
        return Value::Float(f);
    }
    Value::String(s.to_string())
}

pub(crate) fn serialize_json(val: &Value) -> String {
    match val {
        Value::Nil => "null".to_string(),
        Value::Bool(b) => b.to_string(),
        Value::Int(i) => i.to_string(),
        Value::Float(f) => f.to_string(),
        Value::String(s) => format!("\"{}\"", s.replace('"', "\\\"")),
        Value::List(l) => {
            let parts: Vec<String> = l.iter().map(serialize_json).collect();
            format!("[{}]", parts.join(","))
        }
        Value::Map(m) => {
            let mut items: Vec<(String, String)> = m.iter()
                .map(|(k, v)| (k.clone(), serialize_json(v)))
                .collect();
            items.sort_by(|a, b| a.0.cmp(&b.0)); // Stable ordering for testing
            let parts: Vec<String> = items.into_iter()
                .map(|(k, v)| format!("\"{}\":{}", k.replace('"', "\\\""), v))
                .collect();
            format!("{{{}}}", parts.join(","))
        }
    }
}

pub struct FunctionRegistry {
    pub functions: Mutex<HashMap<String, Node>>,
}

pub(crate) fn parse_duration(s: &str) -> Option<std::time::Duration> {
    let s = s.trim();
    if s.ends_with("ms") {
        let val = s[..s.len()-2].parse::<u64>().ok()?;
        Some(std::time::Duration::from_millis(val))
    } else if s.ends_with('s') {
        let val = s[..s.len()-1].parse::<u64>().ok()?;
        Some(std::time::Duration::from_secs(val))
    } else if s.ends_with('m') {
        let val = s[..s.len()-1].parse::<u64>().ok()?;
        Some(std::time::Duration::from_secs(val * 60))
    } else {
        let val = s.parse::<u64>().ok()?;
        Some(std::time::Duration::from_secs(val))
    }
}


pub(crate) fn parse_json(s: &str) -> Result<Value, String> {
    let chars: Vec<char> = s.chars().collect();
    let mut index = 0;
    skip_whitespace(&chars, &mut index);
    let val = parse_json_value(&chars, &mut index)?;
    skip_whitespace(&chars, &mut index);
    if index < chars.len() {
        return Err("Trailing characters in JSON".to_string());
    }
    Ok(val)
}

fn skip_whitespace(chars: &[char], index: &mut usize) {
    while *index < chars.len() && chars[*index].is_whitespace() {
        *index += 1;
    }
}

fn parse_json_value(chars: &[char], index: &mut usize) -> Result<Value, String> {
    skip_whitespace(chars, index);
    if *index >= chars.len() {
        return Err("Unexpected EOF".to_string());
    }
    let c = chars[*index];
    match c {
        '{' => parse_json_object(chars, index),
        '[' => parse_json_array(chars, index),
        '"' => parse_json_string(chars, index),
        't' | 'f' => parse_json_bool(chars, index),
        'n' => parse_json_null(chars, index),
        '-' | '0'..='9' => parse_json_number(chars, index),
        _ => Err(format!("Unexpected character '{}' at {}", c, *index)),
    }
}

fn parse_json_object(chars: &[char], index: &mut usize) -> Result<Value, String> {
    *index += 1; // skip '{'
    let mut map = HashMap::new();
    loop {
        skip_whitespace(chars, index);
        if *index >= chars.len() {
            return Err("Unterminated object".to_string());
        }
        if chars[*index] == '}' {
            *index += 1;
            break;
        }
        if !map.is_empty() {
            if chars[*index] != ',' {
                return Err(format!("Expected ',' in object at {}", *index));
            }
            *index += 1; // skip ','
            skip_whitespace(chars, index);
        }
        if chars[*index] != '"' {
            return Err(format!("Expected key string at {}", *index));
        }
        let key = match parse_json_string(chars, index)? {
            Value::String(s) => s,
            _ => return Err("Expected key string".to_string()),
        };
        skip_whitespace(chars, index);
        if *index >= chars.len() || chars[*index] != ':' {
            return Err(format!("Expected ':' at {}", *index));
        }
        *index += 1; // skip ':'
        let val = parse_json_value(chars, index)?;
        map.insert(key, val);
    }
    Ok(Value::Map(map))
}

fn parse_json_array(chars: &[char], index: &mut usize) -> Result<Value, String> {
    *index += 1; // skip '['
    let mut list = Vec::new();
    loop {
        skip_whitespace(chars, index);
        if *index >= chars.len() {
            return Err("Unterminated array".to_string());
        }
        if chars[*index] == ']' {
            *index += 1;
            break;
        }
        if !list.is_empty() {
            if chars[*index] != ',' {
                return Err(format!("Expected ',' in array at {}", *index));
            }
            *index += 1; // skip ','
            skip_whitespace(chars, index);
        }
        let val = parse_json_value(chars, index)?;
        list.push(val);
    }
    Ok(Value::List(list))
}

fn parse_json_string(chars: &[char], index: &mut usize) -> Result<Value, String> {
    *index += 1; // skip start '"'
    let mut s = String::new();
    let mut escaped = false;
    while *index < chars.len() {
        let c = chars[*index];
        *index += 1;
        if escaped {
            match c {
                '"' => s.push('"'),
                '\\' => s.push('\\'),
                '/' => s.push('/'),
                'b' => s.push('\x08'),
                'f' => s.push('\x0c'),
                'n' => s.push('\n'),
                'r' => s.push('\r'),
                't' => s.push('\t'),
                _ => s.push(c),
            }
            escaped = false;
        } else if c == '\\' {
            escaped = true;
        } else if c == '"' {
            return Ok(Value::String(s));
        } else {
            s.push(c);
        }
    }
    Err("Unterminated string".to_string())
}

fn parse_json_bool(chars: &[char], index: &mut usize) -> Result<Value, String> {
    let mut s = String::new();
    while *index < chars.len() && chars[*index].is_alphabetic() {
        s.push(chars[*index]);
        *index += 1;
    }
    if s == "true" {
        Ok(Value::Bool(true))
    } else if s == "false" {
        Ok(Value::Bool(false))
    } else {
        Err(format!("Invalid boolean value: {}", s))
    }
}

fn parse_json_null(chars: &[char], index: &mut usize) -> Result<Value, String> {
    let mut s = String::new();
    while *index < chars.len() && chars[*index].is_alphabetic() {
        s.push(chars[*index]);
        *index += 1;
    }
    if s == "null" {
        Ok(Value::Nil)
    } else {
        Err(format!("Invalid null value: {}", s))
    }
}

fn parse_json_number(chars: &[char], index: &mut usize) -> Result<Value, String> {
    let mut s = String::new();
    while *index < chars.len() {
        let c = chars[*index];
        if c == '-' || c == '+' || c == '.' || c == 'e' || c == 'E' || c.is_ascii_digit() {
            s.push(c);
            *index += 1;
        } else {
            break;
        }
    }
    if s.contains('.') || s.contains('e') || s.contains('E') {
        if let Ok(f) = s.parse::<f64>() {
            Ok(Value::Float(f))
        } else {
            Err(format!("Invalid float number: {}", s))
        }
    } else {
        if let Ok(i) = s.parse::<i64>() {
            Ok(Value::Int(i))
        } else {
            Err(format!("Invalid integer number: {}", s))
        }
    }
}


#[cfg(test)]
mod tests {
    use super::*;
    use crate::parser::parse_string;
    use crate::executor::Context;

    #[test]
    fn test_logic_slots_loop_and_variables() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);

        let code = r#"
            var: $counter {
              val: 0
            }
            for: "$counter = 0; $counter < 3; $counter++" {
              do: {
                var: $nested {
                  val: $counter
                }
              }
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);

        engine.execute(&mut ctx, &root, &scope).unwrap();
        assert_eq!(scope.get("counter").unwrap(), Value::Int(3));
    }

    #[test]
    fn test_forelse_empty() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);

        let code = r#"
            var: $called {
              val: "no"
            }
            forelse: $list {
              as: $item
              do: {
                var: $called {
                  val: "yes"
                }
              }
              forelse_empty: {
                var: $called {
                  val: "empty_triggered"
                }
              }
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        println!("AST: {:#?}", root);
        let mut ctx = Context::new();
        let scope = Scope::new(None);
        let res = engine.execute(&mut ctx, &root, &scope);
        println!("EXECUTE RESULT: {:?}", res);
        res.unwrap();
        assert_eq!(scope.get("called").unwrap(), Value::String("empty_triggered".to_string()));
    }

    #[test]
    fn test_collection_slots() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);

        let code = r#"
            array.push: $my_list {
              val: "item1"
            }
            array.push: $my_list {
              val: "item2"
            }
            len: $my_list {
              as: $list_len
            }
            collections.get: $my_list {
              index: 1
              as: $second_item
            }
            array.pop: $my_list {
              as: $popped
            }
            array.join: $my_list {
              sep: "-"
              as: $joined
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);
        engine.execute(&mut ctx, &root, &scope).unwrap();

        assert_eq!(scope.get("list_len").unwrap(), Value::Int(2));
        assert_eq!(scope.get("second_item").unwrap(), Value::String("item2".to_string()));
        assert_eq!(scope.get("popped").unwrap(), Value::String("item2".to_string()));
        assert_eq!(scope.get("joined").unwrap(), Value::String("item1".to_string()));
    }

    #[test]
    fn test_map_slots() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);

        let code = r#"
            map.set: $my_map {
              name: "Alice"
              age: 25
            }
            map.keys: $my_map {
              as: $keys
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);
        engine.execute(&mut ctx, &root, &scope).unwrap();

        let keys_list = scope.get("keys").unwrap().to_list();
        assert_eq!(keys_list.len(), 2);
        let mut key_strs: Vec<String> = keys_list.iter().map(|k| k.to_string_coerce()).collect();
        key_strs.sort();
        assert_eq!(key_strs, vec!["age".to_string(), "name".to_string()]);
    }

    #[test]
    fn test_function_slots() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);

        let code = r#"
            fn: greet {
              var: $message {
                val: $name
              }
            }

            var: $name {
              val: "Budi"
            }
            call: greet
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);
        engine.execute(&mut ctx, &root, &scope).unwrap();

        assert_eq!(scope.get("message").unwrap(), Value::String("Budi".to_string()));
    }

    #[test]
    fn test_json_slots() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);

        let code = r#"
            var: $data {
              val: {
                name: "Alice"
                tags: {
                  0: "rust"
                  1: "zeno"
                }
                active: true
                score: 95
              }
            }
            json.stringify: $data {
              as: $json_str
            }
            json.parse: $json_str {
              as: $parsed_data
            }
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);
        engine.execute(&mut ctx, &root, &scope).unwrap();

        let json_str = scope.get("json_str").unwrap().to_string_coerce();
        assert!(json_str.contains("\"name\":\"Alice\""));
        assert!(json_str.contains("\"active\":true"));
        assert!(json_str.contains("\"score\":95"));
        assert!(json_str.contains("\"tags\":[\"rust\",\"zeno\"]"));

        let parsed = scope.get("parsed_data").unwrap().to_map();
        assert_eq!(parsed.get("name").unwrap(), &Value::String("Alice".to_string()));
        assert_eq!(parsed.get("active").unwrap(), &Value::Bool(true));
        assert_eq!(parsed.get("score").unwrap(), &Value::Int(95));
        let tags = parsed.get("tags").unwrap().to_list();
        assert_eq!(tags, vec![Value::String("rust".to_string()), Value::String("zeno".to_string())]);
    }

    #[test]
    fn test_sleep_slot() {
        let mut engine = Engine::new();
        register_logic_slots(&mut engine);

        let code = r#"
            time.sleep: "10ms"
        "#;
        let root = parse_string(code, "test.zl").unwrap();
        let mut ctx = Context::new();
        let scope = Scope::new(None);

        let start = std::time::Instant::now();
        engine.execute(&mut ctx, &root, &scope).unwrap();
        let duration = start.elapsed();

        assert!(duration >= std::time::Duration::from_millis(10));
    }


}
