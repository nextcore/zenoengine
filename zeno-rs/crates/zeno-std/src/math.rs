use zenocore::{Diagnostic, Engine, InputMeta, SlotMeta, Value};
use zenocore::slots::resolve_node_value;
use evalexpr::{ContextWithMutableFunctions, ContextWithMutableVariables};
use std::collections::HashMap;
use std::sync::Arc;

pub fn register(engine: &mut Engine) {
    // ==========================================
    // MATH.CALC
    // ==========================================
    engine.register(
        "math.calc",
        Arc::new(|engine, _ctx, node, scope| {
            let mut expression_str = String::new();
            let mut target = "calc_result".to_string();

            if node.value.is_some() {
                expression_str = resolve_node_value(engine, node, scope).to_string_coerce();
            }

            for c in &node.children {
                if c.name == "val" || c.name == "expr" {
                    expression_str = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if expression_str.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "math.calc: expression is required".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("math.calc".to_string()),
                });
            }

            let clean_expr = expression_str.replace('$', "");
            let mut eval_context = evalexpr::HashMapContext::new();

            for (k, v) in scope.to_map() {
                let eval_val = match v {
                    Value::Nil => evalexpr::Value::Empty,
                    Value::Bool(b) => evalexpr::Value::Boolean(b),
                    Value::Int(i) => evalexpr::Value::Float(i as f64),
                    Value::Float(f) => evalexpr::Value::Float(f),
                    Value::String(s) => {
                        if let Ok(i) = s.parse::<i64>() {
                            evalexpr::Value::Float(i as f64)
                        } else if let Ok(f) = s.parse::<f64>() {
                            evalexpr::Value::Float(f)
                        } else {
                            evalexpr::Value::String(s)
                        }
                    }
                    _ => evalexpr::Value::Empty,
                };
                let _ = eval_context.set_value(k.into(), eval_val);
            }

            // Inject standard math functions
            let _ = eval_context.set_function("ceil".into(), evalexpr::Function::new(|v| {
                let f = match v {
                    evalexpr::Value::Float(f) => *f,
                    evalexpr::Value::Int(i) => *i as f64,
                    _ => return Err(evalexpr::EvalexprError::expected_number(v.clone())),
                };
                Ok(evalexpr::Value::Float(f.ceil()))
            }));
            let _ = eval_context.set_function("floor".into(), evalexpr::Function::new(|v| {
                let f = match v {
                    evalexpr::Value::Float(f) => *f,
                    evalexpr::Value::Int(i) => *i as f64,
                    _ => return Err(evalexpr::EvalexprError::expected_number(v.clone())),
                };
                Ok(evalexpr::Value::Float(f.floor()))
            }));
            let _ = eval_context.set_function("round".into(), evalexpr::Function::new(|v| {
                let f = match v {
                    evalexpr::Value::Float(f) => *f,
                    evalexpr::Value::Int(i) => *i as f64,
                    _ => return Err(evalexpr::EvalexprError::expected_number(v.clone())),
                };
                Ok(evalexpr::Value::Float(f.round()))
            }));
            let _ = eval_context.set_function("abs".into(), evalexpr::Function::new(|v| {
                match v {
                    evalexpr::Value::Float(f) => Ok(evalexpr::Value::Float((*f).abs())),
                    evalexpr::Value::Int(i) => Ok(evalexpr::Value::Int((*i).abs())),
                    _ => Err(evalexpr::EvalexprError::expected_number(v.clone())),
                }
            }));
            let _ = eval_context.set_function("sqrt".into(), evalexpr::Function::new(|v| {
                let f = match v {
                    evalexpr::Value::Float(f) => *f,
                    evalexpr::Value::Int(i) => *i as f64,
                    _ => return Err(evalexpr::EvalexprError::expected_number(v.clone())),
                };
                Ok(evalexpr::Value::Float(f.sqrt()))
            }));
            let _ = eval_context.set_function("pow".into(), evalexpr::Function::new(|v| {
                let args = match v {
                    evalexpr::Value::Tuple(vec) => vec,
                    _ => return Err(evalexpr::EvalexprError::expected_tuple(v.clone())),
                };
                if args.len() != 2 {
                    return Err(evalexpr::EvalexprError::expected_tuple(v.clone()));
                }
                let base = match &args[0] {
                    evalexpr::Value::Float(f) => *f,
                    evalexpr::Value::Int(i) => *i as f64,
                    _ => return Err(evalexpr::EvalexprError::expected_number(args[0].clone())),
                };
                let exp = match &args[1] {
                    evalexpr::Value::Float(f) => *f,
                    evalexpr::Value::Int(i) => *i as f64,
                    _ => return Err(evalexpr::EvalexprError::expected_number(args[1].clone())),
                };
                Ok(evalexpr::Value::Float(base.powf(exp)))
            }));
            let _ = eval_context.set_function("max".into(), evalexpr::Function::new(|v| {
                let args = match v {
                    evalexpr::Value::Tuple(vec) => vec,
                    _ => return Err(evalexpr::EvalexprError::expected_tuple(v.clone())),
                };
                if args.is_empty() {
                    return Err(evalexpr::EvalexprError::expected_tuple(v.clone()));
                }
                let mut max_val = f64::MIN;
                for arg in args {
                    let f = match arg {
                        evalexpr::Value::Float(f) => *f,
                        evalexpr::Value::Int(i) => *i as f64,
                        _ => return Err(evalexpr::EvalexprError::expected_number(arg.clone())),
                    };
                    if f > max_val {
                        max_val = f;
                    }
                }
                Ok(evalexpr::Value::Float(max_val))
            }));
            let _ = eval_context.set_function("min".into(), evalexpr::Function::new(|v| {
                let args = match v {
                    evalexpr::Value::Tuple(vec) => vec,
                    _ => return Err(evalexpr::EvalexprError::expected_tuple(v.clone())),
                };
                if args.is_empty() {
                    return Err(evalexpr::EvalexprError::expected_tuple(v.clone()));
                }
                let mut min_val = f64::MAX;
                for arg in args {
                    let f = match arg {
                        evalexpr::Value::Float(f) => *f,
                        evalexpr::Value::Int(i) => *i as f64,
                        _ => return Err(evalexpr::EvalexprError::expected_number(arg.clone())),
                    };
                    if f < min_val {
                        min_val = f;
                    }
                }
                Ok(evalexpr::Value::Float(min_val))
            }));

            let eval_result = evalexpr::eval_with_context(&clean_expr, &eval_context).map_err(|e| Diagnostic {
                r#type: "error".to_string(),
                message: format!("math.calc: error evaluating '{}': {}", expression_str, e),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("math.calc".to_string()),
            })?;

            let final_val = match eval_result {
                evalexpr::Value::Int(i) => Value::Int(i),
                evalexpr::Value::Float(f) => Value::Float(f),
                evalexpr::Value::Boolean(b) => Value::Bool(b),
                evalexpr::Value::String(s) => Value::String(s),
                _ => Value::Nil,
            };

            scope.set(&target, final_val);
            Ok(())
        }),
        SlotMeta {
            description: "Evaluate a mathematical expression string.".to_string(),
            example: "math.calc: ceil($total / 10)\n  as: $pages".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m.insert("val".to_string(), InputMeta { description: "Math expression".to_string(), required: false, r#type: "string".to_string() });
                m.insert("expr".to_string(), InputMeta { description: "Math expression".to_string(), required: false, r#type: "string".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // MONEY.CALC
    // ==========================================
    engine.register(
        "money.calc",
        Arc::new(|engine, _ctx, node, scope| {
            let mut expression_str = String::new();
            let mut target = "money_result".to_string();

            if node.value.is_some() {
                expression_str = resolve_node_value(engine, node, scope).to_string_coerce();
            }

            for c in &node.children {
                if c.name == "val" {
                    expression_str = engine.resolve_shorthand_value(c, scope).to_string_coerce();
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if expression_str.is_empty() {
                return Err(Diagnostic {
                    r#type: "error".to_string(),
                    message: "money.calc: expression is required".to_string(),
                    filename: node.filename.clone(),
                    line: node.line,
                    col: node.col,
                    slot: Some("money.calc".to_string()),
                });
            }

            let clean_expr = expression_str.replace('$', "");
            let mut eval_context = evalexpr::HashMapContext::new();

            for (k, v) in scope.to_map() {
                let eval_val = match v {
                    Value::Nil => evalexpr::Value::Empty,
                    Value::Bool(b) => evalexpr::Value::Boolean(b),
                    Value::Int(i) => evalexpr::Value::Int(i),
                    Value::Float(f) => evalexpr::Value::Float(f),
                    Value::String(s) => {
                        if let Ok(i) = s.parse::<i64>() {
                            evalexpr::Value::Int(i)
                        } else if let Ok(f) = s.parse::<f64>() {
                            evalexpr::Value::Float(f)
                        } else {
                            evalexpr::Value::String(s)
                        }
                    }
                    _ => evalexpr::Value::Empty,
                };
                let _ = eval_context.set_value(k.into(), eval_val);
            }

            let eval_result = evalexpr::eval_with_context(&clean_expr, &eval_context).map_err(|e| Diagnostic {
                r#type: "error".to_string(),
                message: format!("money.calc: error evaluating '{}': {}", expression_str, e),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("money.calc".to_string()),
            })?;

            let decimal_str = match eval_result {
                evalexpr::Value::Int(i) => {
                    use rust_decimal::prelude::FromPrimitive;
                    if let Some(d) = rust_decimal::Decimal::from_i64(i) {
                        d.to_string()
                    } else {
                        i.to_string()
                    }
                }
                evalexpr::Value::Float(f) => {
                    use rust_decimal::prelude::FromPrimitive;
                    if let Some(d) = rust_decimal::Decimal::from_f64(f) {
                        d.normalize().to_string()
                    } else {
                        f.to_string()
                    }
                }
                evalexpr::Value::String(s) => s,
                _ => "0".to_string(),
            };

            scope.set(&target, Value::String(decimal_str));
            Ok(())
        }),
        SlotMeta {
            description: "Financial math calculations using high precision decimals.".to_string(),
            example: "money.calc: ($harga * $qty) - $diskon\n  as: $total".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("as".to_string(), InputMeta { description: "Target variable".to_string(), required: false, r#type: "any".to_string() });
                m.insert("val".to_string(), InputMeta { description: "Financial expression".to_string(), required: false, r#type: "string".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );
}
