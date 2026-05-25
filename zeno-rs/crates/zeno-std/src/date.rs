use zenocore::{Diagnostic, Engine, InputMeta, SlotMeta, Value};
use zenocore::slots::resolve_node_value;
use std::collections::HashMap;
use std::sync::Arc;
use super::{translate_layout_to_chrono, parse_flex_date, shift_datetime};

pub fn register(engine: &mut Engine) {
    // ==========================================
    // DATE.NOW
    // ==========================================
    engine.register(
        "date.now",
        Arc::new(|engine, _ctx, node, scope| {
            let mut target = "now".to_string();
            let mut layout = "%Y-%m-%dT%H:%M:%S%:z".to_string();

            for c in &node.children {
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
                if c.name == "layout" || c.name == "format" {
                    let layout_val = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                    layout = translate_layout_to_chrono(&layout_val);
                }
            }

            let now = chrono::Local::now();
            let formatted = now.format(&layout).to_string();

            scope.set(&target, Value::String(formatted));
            scope.set(&format!("{}_obj", target), Value::String(now.to_rfc3339()));
            Ok(())
        }),
        SlotMeta {
            description: "Get current date/time.".to_string(),
            example: "date.now: { as: $skarang }".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m.insert("layout".to_string(), InputMeta { description: "Layout format".to_string(), required: false, r#type: "string".to_string() });
                m.insert("format".to_string(), InputMeta { description: "Alias for layout".to_string(), required: false, r#type: "string".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // DATE.FORMAT
    // ==========================================
    engine.register(
        "date.format",
        Arc::new(|engine, _ctx, node, scope| {
            let mut input_str = String::new();
            let mut layout = "%Y-%m-%d %H:%M:%S".to_string();
            let mut target = "formatted_date".to_string();

            if node.value.is_some() {
                input_str = resolve_node_value(engine, node, scope).to_string_coerce();
            }

            for c in &node.children {
                if c.name == "val" || c.name == "date" {
                    input_str = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "layout" || c.name == "format" {
                    let layout_val = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                    layout = translate_layout_to_chrono(&layout_val);
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if input_str.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "date.format: input date is required".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("date.format".to_string()),
                });
            }

            let parsed_dt = parse_flex_date(&input_str).ok_or_else(|| Diagnostic {
                r#type: "error".to_string(),
                message: format!("date.format: failed to parse input date string '{}'", input_str),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("date.format".to_string()),
            })?;

            let formatted = parsed_dt.format(&layout).to_string();
            scope.set(&target, Value::String(formatted));
            Ok(())
        }),
        SlotMeta {
            description: "Format date/time string.".to_string(),
            example: "date.format: $created_at { layout: 'Human'; as: $tgl }".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("val".to_string(), InputMeta { description: "Source date".to_string(), required: false, r#type: "string".to_string() });
                m.insert("date".to_string(), InputMeta { description: "Alias for val".to_string(), required: false, r#type: "string".to_string() });
                m.insert("layout".to_string(), InputMeta { description: "Target format".to_string(), required: false, r#type: "string".to_string() });
                m.insert("format".to_string(), InputMeta { description: "Alias for layout".to_string(), required: false, r#type: "string".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // DATE.PARSE
    // ==========================================
    engine.register(
        "date.parse",
        Arc::new(|engine, _ctx, node, scope| {
            let mut input_str = String::new();
            let mut layout = String::new();
            let mut target = "parsed_date".to_string();

            if node.value.is_some() {
                input_str = resolve_node_value(engine, node, scope).to_string_coerce();
            }

            for c in &node.children {
                if c.name == "val" || c.name == "input" {
                    input_str = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "layout" || c.name == "format" {
                    layout = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if input_str.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "date.parse: input string is required".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("date.parse".to_string()),
                });
            }

            let parsed_dt = if !layout.is_empty() {
                use chrono::TimeZone;
                let chrono_layout = translate_layout_to_chrono(&layout);
                chrono::DateTime::parse_from_str(&input_str, &chrono_layout)
                    .map(|dt| dt.with_timezone(&chrono::Local))
                    .ok()
                    .or_else(|| {
                        chrono::NaiveDateTime::parse_from_str(&input_str, &chrono_layout)
                            .ok()
                            .and_then(|ndt| chrono::Local.from_local_datetime(&ndt).earliest())
                    })
            } else {
                parse_flex_date(&input_str)
            };

            let dt = parsed_dt.ok_or_else(|| Diagnostic {
                r#type: "error".to_string(),
                message: format!("date.parse: failed to parse date string '{}'", input_str),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("date.parse".to_string()),
            })?;

            scope.set(&target, Value::String(dt.to_rfc3339()));
            Ok(())
        }),
        SlotMeta {
            description: "Parse date string.".to_string(),
            example: "date.parse: '2023-12-25' { as: $tgl_obj }".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("input".to_string(), InputMeta { description: "Source string".to_string(), required: false, r#type: "string".to_string() });
                m.insert("val".to_string(), InputMeta { description: "Alias for input".to_string(), required: false, r#type: "string".to_string() });
                m.insert("layout".to_string(), InputMeta { description: "Format layout".to_string(), required: false, r#type: "string".to_string() });
                m.insert("format".to_string(), InputMeta { description: "Alias for layout".to_string(), required: false, r#type: "string".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // DATE.ADD
    // ==========================================
    engine.register(
        "date.add",
        Arc::new(|engine, _ctx, node, scope| {
            let mut input_str = String::new();
            let mut duration_str = String::new();
            let mut target = "shifted_date".to_string();

            if node.value.is_some() {
                input_str = resolve_node_value(engine, node, scope).to_string_coerce();
            }

            for c in &node.children {
                if c.name == "val" || c.name == "date" {
                    input_str = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "duration" || c.name == "add" {
                    duration_str = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            let base_dt = if input_str.is_empty() {
                chrono::Local::now()
            } else {
                parse_flex_date(&input_str).ok_or_else(|| Diagnostic {
                    r#type: "error".to_string(),
                    message: format!("date.add: failed to parse base date string '{}'", input_str),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("date.add".to_string()),
                })?
            };

            let shifted_dt = shift_datetime(&base_dt, &duration_str).ok_or_else(|| Diagnostic {
                r#type: "error".to_string(),
                message: format!("date.add: invalid duration format '{}' (e.g. 2h, -1d)", duration_str),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("date.add".to_string()),
            })?;

            scope.set(&target, Value::String(shifted_dt.to_rfc3339()));
            Ok(())
        }),
        SlotMeta {
            description: "Shift a date/time by a duration.".to_string(),
            example: "date.add: $now { duration: '2h'; as: $future }".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("date".to_string(), InputMeta { description: "Base date".to_string(), required: false, r#type: "string".to_string() });
                m.insert("val".to_string(), InputMeta { description: "Alias for date".to_string(), required: false, r#type: "string".to_string() });
                m.insert("duration".to_string(), InputMeta { description: "Duration string (e.g. 2h, -1d)".to_string(), required: false, r#type: "string".to_string() });
                m.insert("add".to_string(), InputMeta { description: "Alias for duration".to_string(), required: false, r#type: "string".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );
}
