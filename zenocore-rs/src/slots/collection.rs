use crate::diagnostic::Diagnostic;
use crate::executor::{Engine, InputMeta, SlotMeta};
use crate::scope::Value;
use std::collections::HashMap;
use std::sync::Arc;
use super::resolve_node_value;

pub fn register(engine: &mut Engine) {
    // ==========================================
    // ARRAY.PUSH
    // ==========================================
    engine.register(
        "array.push",
        Arc::new(|engine, _ctx, node, scope| {
            let mut target_name = String::new();
            let mut items = Vec::new();

            if let Some(ref val) = node.value {
                target_name = val.trim_start_matches('$').to_string();
            }

            for c in &node.children {
                if c.name == "in" || c.name == "list" {
                    if let Some(ref cv) = c.value {
                        target_name = cv.trim_start_matches('$').to_string();
                    }
                }
                if c.name == "val" || c.name == "value" || c.name == "item" {
                    items.push(engine.resolve_shorthand_value(c, scope));
                }
            }

            if target_name.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "array.push: target list not specified".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("array.push".to_string()),
                });
            }

            let mut list = match scope.get(&target_name) {
                Some(Value::List(l)) => l.clone(),
                Some(Value::Nil) | None => Vec::new(),
                Some(other) => vec![other],
            };

            list.extend(items);
            scope.set(&target_name, Value::List(list));
            Ok(())
        }),
        SlotMeta {
            description: "Add one or more items to the end of an array.".to_string(),
            example: "array.push: $my_list\n  val: 'New Item'".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("in".to_string(), InputMeta { description: "Target list".to_string(), required: false, r#type: "string".to_string() });
                m.insert("list".to_string(), InputMeta { description: "Target list".to_string(), required: false, r#type: "string".to_string() });
                m.insert("val".to_string(), InputMeta { description: "Value to push".to_string(), required: false, r#type: "any".to_string() });
                m.insert("value".to_string(), InputMeta { description: "Value to push".to_string(), required: false, r#type: "any".to_string() });
                m.insert("item".to_string(), InputMeta { description: "Value to push".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // COLLECTIONS.GET
    // ==========================================
    engine.register(
        "collections.get",
        Arc::new(|engine, _ctx, node, scope| {
            let mut list = Vec::new();
            let mut index = 0;
            let mut target = "item".to_string();

            if node.value.is_some() {
                list = resolve_node_value(engine, node, scope).to_list();
            }

            for c in &node.children {
                if c.name == "in" || c.name == "list" {
                    list = engine.resolve_shorthand_value(c, scope).to_list();
                }
                if c.name == "index" || c.name == "i" {
                    index = engine.resolve_shorthand_value(c, scope).to_int() as usize;
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if list.is_empty() {
                scope.set(&target, Value::Nil);
                return Ok(());
            }

            if index >= list.len() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "collections.get: index out of bounds".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("collections.get".to_string()),
                });
            }

            scope.set(&target, list[index].clone());
            Ok(())
        }),
        SlotMeta {
            description: "Get item from array at index.".to_string(),
            example: "collections.get: $list\n  index: 0\n  as: $item".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("in".to_string(), InputMeta { description: "Source list".to_string(), required: false, r#type: "list".to_string() });
                m.insert("list".to_string(), InputMeta { description: "Source list".to_string(), required: false, r#type: "list".to_string() });
                m.insert("index".to_string(), InputMeta { description: "Index".to_string(), required: true, r#type: "int".to_string() });
                m.insert("i".to_string(), InputMeta { description: "Index".to_string(), required: false, r#type: "int".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // ARRAY.POP
    // ==========================================
    engine.register(
        "array.pop",
        Arc::new(|_engine, _ctx, node, scope| {
            let mut target_name = String::new();
            let mut dst_name = "popped_item".to_string();

            if let Some(ref val) = node.value {
                target_name = val.trim_start_matches('$').to_string();
            }

            for c in &node.children {
                if c.name == "in" || c.name == "list" {
                    if let Some(ref cv) = c.value {
                        target_name = cv.trim_start_matches('$').to_string();
                    }
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        dst_name = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if target_name.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "array.pop: target list not specified".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("array.pop".to_string()),
                });
            }

            let current_val = scope.get(&target_name).unwrap_or(Value::Nil);
            let mut list = current_val.to_list();

            if list.is_empty() {
                scope.set(&dst_name, Value::Nil);
                return Ok(());
            }

            let popped = list.pop().unwrap();
            scope.set(&target_name, Value::List(list));
            scope.set(&dst_name, popped);
            Ok(())
        }),
        SlotMeta {
            description: "Remove and return the last item of an array.".to_string(),
            example: "array.pop: $stack\n  as: $item".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("in".to_string(), InputMeta { description: "Source list".to_string(), required: false, r#type: "string".to_string() });
                m.insert("list".to_string(), InputMeta { description: "Source list".to_string(), required: false, r#type: "string".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // ARRAY.JOIN
    // ==========================================
    engine.register(
        "array.join",
        Arc::new(|engine, _ctx, node, scope| {
            let mut list = Vec::new();
            let mut separator = ",".to_string();
            let mut target = "joined_string".to_string();

            if node.value.is_some() {
                list = resolve_node_value(engine, node, scope).to_list();
            }

            for c in &node.children {
                if c.name == "list" {
                    list = engine.resolve_shorthand_value(c, scope).to_list();
                }
                if c.name == "sep" || c.name == "separator" {
                    separator = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            let str_list: Vec<String> = list.iter().map(|item| item.to_string_coerce()).collect();
            scope.set(&target, Value::String(str_list.join(&separator)));
            Ok(())
        }),
        SlotMeta {
            description: "Join array elements into a string with a separator.".to_string(),
            example: "array.join: $tags\n  sep: ', '\n  as: $tag_str".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("list".to_string(), InputMeta { description: "Source list".to_string(), required: false, r#type: "list".to_string() });
                m.insert("sep".to_string(), InputMeta { description: "Separator".to_string(), required: false, r#type: "string".to_string() });
                m.insert("separator".to_string(), InputMeta { description: "Separator".to_string(), required: false, r#type: "string".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // MAP.SET
    // ==========================================
    engine.register(
        "map.set",
        Arc::new(|engine, _ctx, node, scope| {
            let mut target_name = String::new();

            if let Some(ref val) = node.value {
                target_name = val.trim_start_matches('$').to_string();
            }

            for c in &node.children {
                if c.name == "map" {
                    if let Some(ref cv) = c.value {
                        target_name = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if target_name.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "map.set: target map not specified".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("map.set".to_string()),
                });
            }

            let mut map_val = match scope.get(&target_name) {
                Some(Value::Map(m)) => m.clone(),
                _ => HashMap::new(),
            };

            let mut explicit_key = String::new();
            let mut explicit_val = Value::Nil;
            let mut has_explicit = false;

            for c in &node.children {
                if c.name == "key" {
                    explicit_key = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                    has_explicit = true;
                } else if c.name == "val" || c.name == "value" {
                    explicit_val = engine.resolve_shorthand_value(c, scope);
                } else if c.name != "map" {
                    map_val.insert(c.name.clone(), engine.resolve_shorthand_value(c, scope));
                }
            }

            if has_explicit && !explicit_key.is_empty() {
                map_val.insert(explicit_key, explicit_val);
            }

            scope.set(&target_name, Value::Map(map_val));
            Ok(())
        }),
        SlotMeta {
            description: "Set values in a map/object.".to_string(),
            example: "map.set: $user\n  age: 30".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("map".to_string(), InputMeta { description: "Target map".to_string(), required: false, r#type: "string".to_string() });
                m.insert("key".to_string(), InputMeta { description: "Key".to_string(), required: false, r#type: "string".to_string() });
                m.insert("val".to_string(), InputMeta { description: "Value".to_string(), required: false, r#type: "any".to_string() });
                m.insert("value".to_string(), InputMeta { description: "Value to push".to_string(), required: false, r#type: "any".to_string() });
                m.insert("*".to_string(), InputMeta { description: "Dynamic properties".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // MAP.KEYS
    // ==========================================
    engine.register(
        "map.keys",
        Arc::new(|engine, _ctx, node, scope| {
            let mut val = Value::Nil;
            let mut target = "keys".to_string();

            if node.value.is_some() {
                val = resolve_node_value(engine, node, scope);
            }

            for c in &node.children {
                if c.name == "map" {
                    val = engine.resolve_shorthand_value(c, scope);
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            let keys_list = match val {
                Value::Map(m) => m.keys().map(|k| Value::String(k.clone())).collect(),
                _ => Vec::new(),
            };

            scope.set(&target, Value::List(keys_list));
            Ok(())
        }),
        SlotMeta {
            description: "Get all keys of a map.".to_string(),
            example: "map.keys: $user\n  as: $fields".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("map".to_string(), InputMeta { description: "Source map".to_string(), required: false, r#type: "map".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // LEN
    // ==========================================
    engine.register(
        "len",
        Arc::new(|engine, _ctx, node, scope| {
            let mut val = Value::Nil;
            let mut target = "len".to_string();

            if node.value.is_some() {
                val = resolve_node_value(engine, node, scope);
            }

            for c in &node.children {
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
                if c.name == "in" || c.name == "list" || c.name == "val" || c.name == "value" {
                    val = engine.resolve_shorthand_value(c, scope);
                }
            }

            let length = match val {
                Value::Nil => 0,
                Value::String(s) => s.len() as i64,
                Value::List(l) => l.len() as i64,
                Value::Map(m) => m.len() as i64,
                _ => 0,
            };

            scope.set(&target, Value::Int(length));
            Ok(())
        }),
        SlotMeta {
            description: "Get the length of a string, array, or map.".to_string(),
            example: "len: $my_list\n  as: $count".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("in".to_string(), InputMeta { description: "Collection".to_string(), required: false, r#type: "any".to_string() });
                m.insert("list".to_string(), InputMeta { description: "Collection".to_string(), required: false, r#type: "any".to_string() });
                m.insert("val".to_string(), InputMeta { description: "Collection".to_string(), required: false, r#type: "any".to_string() });
                m.insert("value".to_string(), InputMeta { description: "Collection".to_string(), required: false, r#type: "any".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );
}
