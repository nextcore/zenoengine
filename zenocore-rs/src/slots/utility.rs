use crate::diagnostic::Diagnostic;
use crate::executor::{Engine, InputMeta, SlotMeta};
use crate::scope::Value;
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use super::{resolve_node_value, parse_json, serialize_json, parse_duration, FunctionRegistry};

pub fn register(engine: &mut Engine) {
    // ==========================================
    // FN (Define Function)
    // ==========================================
    engine.register(
        "fn",
        Arc::new(|_engine, ctx, node, _scope| {
            let func_name = node.value.clone().unwrap_or_default();
            let func_clean = func_name.trim_start_matches('$').to_string();
            if func_clean.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "fn: function name is required".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("fn".to_string()),
                });
            }

            let registry = if let Some(reg) = ctx.get::<FunctionRegistry>("functions") {
                reg
            } else {
                ctx.set("functions", FunctionRegistry { functions: Mutex::new(HashMap::new()) });
                ctx.get::<FunctionRegistry>("functions").ok_or_else(|| Diagnostic {
                    r#type: "error".to_string(),
                    message: "fn: failed to initialize function registry".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("fn".to_string()),
                })?
            };

            registry.functions.lock().unwrap().insert(func_clean, node.clone());
            Ok(())
        }),
        SlotMeta {
            description: "Define a reusable function code block.".to_string(),
            example: "fn: hitung_gaji {\n  log: $gaji\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // CALL (Invoke Function)
    // ==========================================
    engine.register(
        "call",
        Arc::new(|engine, ctx, node, scope| {
            let func_name = node.value.clone().unwrap_or_default();
            let func_clean = func_name.trim_start_matches('$').to_string();
            if func_clean.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "call: function name is required".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("call".to_string()),
                });
            }

            let registry = ctx.get::<FunctionRegistry>("functions").ok_or_else(|| Diagnostic {
                r#type: "error".to_string(),
                message: format!("call: function '{}' not found (no functions registered)", func_clean),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("call".to_string()),
            })?;

            let func_node = {
                let funcs = registry.functions.lock().unwrap();
                funcs.get(&func_clean).cloned().ok_or_else(|| Diagnostic {
                    r#type: "error".to_string(),
                    message: format!("call: function '{}' not found", func_clean),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("call".to_string()),
                })?
            };

            for child in &func_node.children {
                engine.execute(ctx, child, scope)?;
            }

            Ok(())
        }),
        SlotMeta {
            description: "Call a registered function code block.".to_string(),
            example: "call: hitung_gaji".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // JSON.PARSE
    // ==========================================
    engine.register(
        "json.parse",
        Arc::new(|engine, _ctx, node, scope| {
            let mut json_str = String::new();
            let mut target = "json_result".to_string();

            if node.value.is_some() {
                json_str = resolve_node_value(engine, node, scope).to_string_coerce();
            }

            for c in &node.children {
                if c.name == "val" || c.name == "value" {
                    json_str = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if json_str.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "json.parse: input value is empty".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("json.parse".to_string()),
                });
            }

            let val = parse_json(&json_str).map_err(|e| Diagnostic {
                r#type: "error".to_string(),
                message: format!("json.parse: invalid json format: {}", e),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("json.parse".to_string()),
            })?;

            scope.set(&target, val);
            Ok(())
        }),
        SlotMeta {
            description: "Parse a JSON string into a structured Value.".to_string(),
            example: "json.parse: $response_body\n  as: $data".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("val".to_string(), InputMeta { description: "JSON string".to_string(), required: false, r#type: "string".to_string() });
                m.insert("value".to_string(), InputMeta { description: "JSON string".to_string(), required: false, r#type: "string".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // JSON.STRINGIFY
    // ==========================================
    engine.register(
        "json.stringify",
        Arc::new(|engine, _ctx, node, scope| {
            let mut val = Value::Nil;
            let mut target = "json_string".to_string();

            if node.value.is_some() {
                val = resolve_node_value(engine, node, scope);
            }

            for c in &node.children {
                if c.name == "val" || c.name == "value" {
                    val = engine.resolve_shorthand_value(c, scope);
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            let serialized = serialize_json(&val);
            scope.set(&target, Value::String(serialized));
            Ok(())
        }),
        SlotMeta {
            description: "Serialize a structured Value into a JSON string.".to_string(),
            example: "json.stringify: $data\n  as: $json_str".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("val".to_string(), InputMeta { description: "Value to serialize".to_string(), required: false, r#type: "any".to_string() });
                m.insert("value".to_string(), InputMeta { description: "Value to serialize".to_string(), required: false, r#type: "any".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // TIME.SLEEP
    // ==========================================
    engine.register(
        "time.sleep",
        Arc::new(|engine, _ctx, node, scope| {
            let mut duration_str = String::new();

            if node.value.is_some() {
                duration_str = resolve_node_value(engine, node, scope).to_string_coerce();
            }

            for c in &node.children {
                if c.name == "duration" || c.name == "val" {
                    duration_str = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
            }

            if duration_str.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "time.sleep: duration is required".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("time.sleep".to_string()),
                });
            }

            let dur = parse_duration(&duration_str).ok_or_else(|| Diagnostic {
                r#type: "error".to_string(),
                message: format!("time.sleep: invalid duration format '{}'", duration_str),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("time.sleep".to_string()),
            })?;

            std::thread::sleep(dur);
            Ok(())
        }),
        SlotMeta {
            description: "Pause execution for a duration.".to_string(),
            example: "time.sleep: '1s'".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("duration".to_string(), InputMeta { description: "Duration string (e.g. 1s, 500ms)".to_string(), required: false, r#type: "string".to_string() });
                m.insert("val".to_string(), InputMeta { description: "Duration string (e.g. 1s, 500ms)".to_string(), required: false, r#type: "string".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );
}
