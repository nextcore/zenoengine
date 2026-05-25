use crate::diagnostic::Diagnostic;
use crate::executor::{Engine, InputMeta, SlotMeta, Context};
use crate::parser::Node;
use crate::scope::{Scope, Value};
use std::collections::HashMap;
use std::sync::Arc;
use super::{eval_simple_condition, serialize_json};

pub fn register(engine: &mut Engine) {
    // ==========================================
    // SLOT: RETURN / STOP
    // ==========================================
    engine.register(
        "return",
        Arc::new(|_engine, _ctx, node, _scope| {
            Err(Diagnostic {
                r#type: "return".to_string(),
                message: "return signal".to_string(),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("return".to_string()),
            })
        }),
        SlotMeta {
            description: "Halt execution of the current block/handler.".to_string(),
            example: "return".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: SCOPE SET / VAR
    // ==========================================
    let var_handler = Arc::new(|engine: &Engine, _ctx: &mut Context, node: &Node, scope: &Arc<Scope>| {
        let mut key = String::new();
        let mut val = Value::Nil;

        // Support shorthand: scope.set: $nama_var
        if let Some(ref v) = node.value {
            if v.starts_with('$') {
                key = v[1..].to_string();
            } else {
                key = v.clone();
            }
        }

        for c in &node.children {
            if c.name == "key" || c.name == "name" {
                if let Some(ref cv) = c.value {
                    if cv.starts_with('$') {
                        key = cv[1..].to_string();
                    } else {
                        key = cv.clone();
                    }
                }
            }
            if c.name == "val" || c.name == "value" {
                val = engine.resolve_shorthand_value(c, scope);
            }
        }

        if !key.is_empty() {
            scope.set(&key, val);
        }
        Ok(())
    });

    let var_meta = SlotMeta {
        description: "Create a variable.".to_string(),
        example: "var: $my_var\n  val: 123".to_string(),
        inputs: {
            let mut m = HashMap::new();
            m.insert("key".to_string(), InputMeta { description: "Variable name".to_string(), required: false, r#type: "string".to_string() });
            m.insert("name".to_string(), InputMeta { description: "Variable name".to_string(), required: false, r#type: "string".to_string() });
            m.insert("val".to_string(), InputMeta { description: "Variable value".to_string(), required: false, r#type: "any".to_string() });
            m.insert("value".to_string(), InputMeta { description: "Variable value".to_string(), required: false, r#type: "any".to_string() });
            m
        },
        required_blocks: Vec::new(),
        value_type: String::new(),
    };

    engine.register("scope.set", var_handler.clone(), var_meta.clone());
    engine.register("var", var_handler, var_meta);

    // ==========================================
    // SLOT: LOGIC.COMPARE
    // ==========================================
    engine.register(
        "logic.compare",
        Arc::new(|engine, _ctx, node, scope| {
            let mut v1 = Value::Nil;
            let mut v2 = Value::Nil;
            let mut op = String::new();
            let mut target = "compare_result".to_string();

            for c in &node.children {
                if c.name == "v1" {
                    v1 = engine.resolve_shorthand_value(c, scope);
                }
                if c.name == "v2" {
                    v2 = engine.resolve_shorthand_value(c, scope);
                }
                if c.name == "op" {
                    op = c.value.clone().unwrap_or_default();
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            let res = match op.as_str() {
                "==" => v1.to_string_coerce() == v2.to_string_coerce(),
                "!=" => v1.to_string_coerce() != v2.to_string_coerce(),
                ">" => v1.to_float() > v2.to_float(),
                "<" => v1.to_float() < v2.to_float(),
                ">=" => v1.to_float() >= v2.to_float(),
                "<=" => v1.to_float() <= v2.to_float(),
                _ => false,
            };

            scope.set(&target, Value::Bool(res));
            Ok(())
        }),
        SlotMeta {
            description: "Compare two values.".to_string(),
            example: "logic.compare\n  v1: $age\n  op: '>'\n  v2: 17".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("v1".to_string(), InputMeta { description: "First value".to_string(), required: true, r#type: "any".to_string() });
                m.insert("v2".to_string(), InputMeta { description: "Second value".to_string(), required: true, r#type: "any".to_string() });
                m.insert("op".to_string(), InputMeta { description: "Operator".to_string(), required: true, r#type: "string".to_string() });
                m.insert("as".to_string(), InputMeta { description: "Result target".to_string(), required: false, r#type: "string".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: TRY / CATCH
    // ==========================================
    engine.register(
        "try",
        Arc::new(|engine, ctx, node, scope| {
            let mut do_node = None;
            let mut catch_node = None;
            let mut err_var = "error".to_string();

            for c in &node.children {
                if c.name == "do" {
                    do_node = Some(c);
                }
                if c.name == "catch" {
                    catch_node = Some(c);
                }
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        err_var = cv.trim_start_matches('$').to_string();
                    }
                }
            }

            if let Some(do_n) = do_node {
                for child in &do_n.children {
                    if let Err(err) = engine.execute(ctx, child, scope) {
                        // DON'T catch control flow errors
                        if err.r#type == "return" || err.r#type == "break" || err.r#type == "continue" {
                            return Err(err);
                        }
                        if let Some(catch_n) = catch_node {
                            scope.set(&err_var, Value::String(err.message.clone()));
                            for catch_child in &catch_n.children {
                                engine.execute(ctx, catch_child, scope)?;
                            }
                            return Ok(());
                        }
                        return Err(err);
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Handle errors using a try-catch block.".to_string(),
            example: "try {\n  do: { ... }\n  catch: { ... }\n}".to_string(),
            inputs: {
                let mut m = HashMap::new();
                m.insert("as".to_string(), InputMeta { description: "Error variable name".to_string(), required: false, r#type: "string".to_string() });
                m.insert("do".to_string(), InputMeta { description: "Main block".to_string(), required: false, r#type: "any".to_string() });
                m.insert("catch".to_string(), InputMeta { description: "Catch block".to_string(), required: false, r#type: "any".to_string() });
                m
            },
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: FOR LOOP / FOREACH
    // ==========================================
    let for_handler = Arc::new(|engine: &Engine, ctx: &mut Context, node: &Node, scope: &Arc<Scope>| {
        let mut raw_str = String::new();
        let mut loop_val = Value::Nil;

        if let Some(ref val) = node.value {
            if val.contains(';') {
                raw_str = val.clone();
            } else {
                loop_val = engine.resolve_shorthand_value(node, scope);
                raw_str = loop_val.to_string_coerce();
            }
        }

        let mut do_node = None;
        let mut item_name = "item".to_string();

        for c in &node.children {
            if c.name == "do" {
                do_node = Some(c);
            }
            if c.name == "as" {
                if let Some(ref cv) = c.value {
                    item_name = cv.trim_start_matches('$').to_string();
                }
            }
        }

        let do_n = match do_node {
            Some(n) => n,
            None => return Ok(()),
        };

        // A. C-Style For Loop (e.g. "$i = 0; $i < 10; $i++")
        if raw_str.chars().filter(|&c| c == ';').count() == 2 {
            let parts: Vec<&str> = raw_str.split(';').collect();
            let init_str = parts[0].trim();
            let cond_str = parts[1].trim();
            let upd_str = parts[2].trim();

            // 1. INIT
            let mut loop_var = String::new();
            if init_str.contains('=') {
                let init_parts: Vec<&str> = init_str.splitn(2, '=').collect();
                loop_var = init_parts[0].trim().trim_start_matches('$').to_string();
                let init_val = init_parts[1].trim().parse::<i64>().unwrap_or(0);
                scope.set(&loop_var, Value::Int(init_val));
            }

            // LOOP
            let max_loop = 10000;
            for _ in 0..max_loop {
                // 2. CONDITION
                if !eval_simple_condition(cond_str, scope) {
                    break;
                }

                // DO BODY
                for child in &do_n.children {
                    if let Err(err) = engine.execute(ctx, child, scope) {
                        if err.r#type == "break" {
                            return Ok(());
                        }
                        if err.r#type == "continue" {
                            break; // break children execution, continue to update
                        }
                        return Err(err);
                    }
                }

                // 3. UPDATE
                if !loop_var.is_empty() {
                    let v_raw = scope.get(&loop_var).unwrap_or(Value::Int(0));
                    let v_int = v_raw.to_int();
                    if upd_str.contains("++") {
                        scope.set(&loop_var, Value::Int(v_int + 1));
                    } else if upd_str.contains("--") {
                        scope.set(&loop_var, Value::Int(v_int - 1));
                    }
                }
            }
            return Ok(());
        }

        // B. Foreach Loop
        let source_list = loop_val.to_list();
        let count = source_list.len();
        let parent_loop = scope.get("loop");

        for (i, item) in source_list.into_iter().enumerate() {
            scope.set(&item_name, item);

            let mut loop_data = HashMap::new();
            loop_data.insert("index".to_string(), Value::Int(i as i64));
            loop_data.insert("iteration".to_string(), Value::Int((i + 1) as i64));
            loop_data.insert("remaining".to_string(), Value::Int((count - (i + 1)) as i64));
            loop_data.insert("count".to_string(), Value::Int(count as i64));
            loop_data.insert("first".to_string(), Value::Bool(i == 0));
            loop_data.insert("last".to_string(), Value::Bool(i == count - 1));
            loop_data.insert("even".to_string(), Value::Bool((i + 1) % 2 == 0));
            loop_data.insert("odd".to_string(), Value::Bool((i + 1) % 2 != 0));

            if let Some(ref p_loop) = parent_loop {
                loop_data.insert("parent".to_string(), p_loop.clone());
            }

            scope.set("loop", Value::Map(loop_data));

            // Execute children
            let mut skipped = false;
            for child in &do_n.children {
                if let Err(err) = engine.execute(ctx, child, scope) {
                    if err.r#type == "break" {
                        skipped = true;
                        break;
                    }
                    if err.r#type == "continue" {
                        break;
                    }
                    return Err(err);
                }
            }
            if skipped {
                break;
            }
        }

        if let Some(p_loop) = parent_loop {
            scope.set("loop", p_loop);
        } else {
            scope.delete("loop");
        }

        Ok(())
    });

    let for_meta = SlotMeta {
        description: "Iterate (loop) over a list or array.".to_string(),
        example: "for: $list\n  as: $item\n  do:\n    log: $item".to_string(),
        inputs: {
            let mut m = HashMap::new();
            m.insert("as".to_string(), InputMeta { description: "Alias variable".to_string(), required: false, r#type: "string".to_string() });
            m.insert("do".to_string(), InputMeta { description: "Do block".to_string(), required: true, r#type: "any".to_string() });
            m
        },
        required_blocks: vec!["do".to_string()],
        value_type: String::new(),
    };

    engine.register("for", for_handler.clone(), for_meta.clone());
    engine.register("foreach", for_handler, for_meta);

    // ==========================================
    // SLOT: WHILE LOOP / LOOP
    // ==========================================
    let while_handler = Arc::new(|engine: &Engine, ctx: &mut Context, node: &Node, scope: &Arc<Scope>| {
        let cond_raw = node.value.clone().unwrap_or_default();
        let mut do_c = None;
        for c in &node.children {
            if c.name == "do" {
                do_c = Some(c);
                break;
            }
        }

        let children_to_exec = match do_c {
            Some(do_node) => &do_node.children,
            None => &node.children,
        };

        let max_loop = 10000;
        for _ in 0..max_loop {
            if !eval_simple_condition(&cond_raw, scope) {
                break;
            }

            let mut skipped = false;
            for child in children_to_exec {
                if let Err(err) = engine.execute(ctx, child, scope) {
                    if err.r#type == "break" {
                        skipped = true;
                        break;
                    }
                    if err.r#type == "continue" {
                        break;
                    }
                    return Err(err);
                }
            }
            if skipped {
                break;
            }
        }
        Ok(())
    });

    let while_meta = SlotMeta {
        description: "While loop".to_string(),
        example: "while: $i < 10 {\n  do: {\n    log: $i\n  }\n}".to_string(),
        inputs: {
            let mut m = HashMap::new();
            m.insert("do".to_string(), InputMeta { description: "Do block".to_string(), required: false, r#type: "any".to_string() });
            m
        },
        required_blocks: Vec::new(),
        value_type: String::new(),
    };
    engine.register("while", while_handler.clone(), while_meta.clone());
    engine.register("loop", while_handler, while_meta);

    // ==========================================
    // SLOT: BREAK & CONTINUE
    // ==========================================
    engine.register(
        "break",
        Arc::new(|_engine, _ctx, node, scope| {
            if let Some(ref expr) = node.value {
                if !eval_simple_condition(expr, scope) {
                    return Ok(()); // Don't break if condition not met
                }
            }
            Err(Diagnostic {
                r#type: "break".to_string(),
                message: "break signal".to_string(),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("break".to_string()),
            })
        }),
        SlotMeta {
            description: "Force stop loop. Supports conditional.".to_string(),
            example: "break: $i == 5".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    engine.register(
        "continue",
        Arc::new(|_engine, _ctx, node, scope| {
            if let Some(ref expr) = node.value {
                if !eval_simple_condition(expr, scope) {
                    return Ok(()); // Don't continue if condition not met
                }
            }
            Err(Diagnostic {
                r#type: "continue".to_string(),
                message: "continue signal".to_string(),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("continue".to_string()),
            })
        }),
        SlotMeta {
            description: "Continue to next loop iteration. Supports conditional.".to_string(),
            example: "continue: $i % 2 == 0".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: DD & DUMP
    // ==========================================
    engine.register(
        "dump",
        Arc::new(|engine, _ctx, node, scope| {
            let val = engine.resolve_shorthand_value(node, scope);
            println!("[DUMP]: {:?} (Type: {:?})", val, val);
            Ok(())
        }),
        SlotMeta {
            description: "Dump variable content without stopping execution.".to_string(),
            example: "dump: $user".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    engine.register(
        "dd",
        Arc::new(|engine, _ctx, node, scope| {
            let val = engine.resolve_shorthand_value(node, scope);
            println!("[DD]: {:?} (Type: {:?})", val, val);
            Err(Diagnostic {
                r#type: "error".to_string(),
                message: "DD HALT: execution stopped by dd".to_string(),
                filename: node.filename.clone(),
                line: node.line,
                col: node.col,
                slot: Some("dd".to_string()),
            })
        }),
        SlotMeta {
            description: "Dump and Die. Display variable content and stop script immediately.".to_string(),
            example: "dd: $user".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: SWITCH
    // ==========================================
    engine.register(
        "switch",
        Arc::new(|engine, ctx, node, scope| {
            let val = engine.resolve_shorthand_value(node, scope);

            for child in &node.children {
                if child.name == "case" {
                    let case_val = engine.resolve_shorthand_value(child, scope);
                    if val == case_val {
                        if let Err(err) = engine.execute(ctx, child, scope) {
                            if err.r#type == "break" {
                                return Ok(());
                            }
                            return Err(err);
                        }
                        return Ok(());
                    }
                } else if child.name == "default" {
                    if let Err(err) = engine.execute(ctx, child, scope) {
                        if err.r#type == "break" {
                            return Ok(());
                        }
                        return Err(err);
                    }
                    return Ok(());
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Switch-case conditional branching.".to_string(),
            example: "switch: $role {\n  case 'admin': { ... }\n  default: { ... }\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: ISSET, EMPTY, UNLESS
    // ==========================================
    let is_empty_fn = |val: &Value| -> bool {
        match val {
            Value::Nil => true,
            Value::String(s) => s.is_empty(),
            Value::List(l) => l.is_empty(),
            Value::Map(m) => m.is_empty(),
            _ => false,
        }
    };

    engine.register(
        "isset",
        Arc::new(|engine, ctx, node, scope| {
            let val = engine.resolve_shorthand_value(node, scope);
            if val != Value::Nil {
                for child in &node.children {
                    if child.name == "do" {
                        return engine.execute(ctx, child, scope);
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Execute block if variable is set/defined.".to_string(),
            example: "isset: $user {\n  do: { ... }\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    engine.register(
        "empty",
        Arc::new(move |engine, ctx, node, scope| {
            let val = engine.resolve_shorthand_value(node, scope);
            if is_empty_fn(&val) {
                for child in &node.children {
                    if child.name == "do" {
                        return engine.execute(ctx, child, scope);
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Execute block if variable is empty.".to_string(),
            example: "empty: $list {\n  do: { ... }\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    engine.register(
        "unless",
        Arc::new(|engine, ctx, node, scope| {
            let val = engine.resolve_shorthand_value(node, scope);
            if !val.to_bool() {
                for child in &node.children {
                    if child.name == "do" {
                        return engine.execute(ctx, child, scope);
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Reverse of IF. Execute block if condition is FALSE.".to_string(),
            example: "unless: $is_admin {\n  do: { ... }\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: AUTH & GUEST (Check Login)
    // ==========================================
    let is_auth_fn = |scope: &Arc<Scope>| -> bool {
        let user = scope.get("user");
        let auth = scope.get("auth");
        (user.is_some() && user.unwrap() != Value::Nil) || (auth.is_some() && auth.unwrap() != Value::Nil)
    };

    engine.register(
        "auth",
        Arc::new(move |engine, ctx, node, scope| {
            if is_auth_fn(scope) {
                for child in &node.children {
                    if child.name == "do" {
                        return engine.execute(ctx, child, scope);
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Execute block if user is logged in.".to_string(),
            example: "auth {\n  do: { ... }\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    engine.register(
        "guest",
        Arc::new(move |engine, ctx, node, scope| {
            if !is_auth_fn(scope) {
                for child in &node.children {
                    if child.name == "do" {
                        return engine.execute(ctx, child, scope);
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Execute block if user is guest.".to_string(),
            example: "guest {\n  do: { ... }\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    engine.register(
        "auth.user",
        Arc::new(|_engine, _ctx, node, scope| {
            let mut target = "user".to_string();
            if let Some(ref v) = node.value {
                target = v.trim_start_matches('$').to_string();
            }
            for c in &node.children {
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }
            let user = scope.get("user").or_else(|| scope.get("auth")).unwrap_or(Value::Nil);
            scope.set(&target, user);
            Ok(())
        }),
        SlotMeta {
            description: "Retrieve current logged-in user data.".to_string(),
            example: "auth.user: $currentUser".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    engine.register(
        "auth.check",
        Arc::new(move |_engine, _ctx, node, scope| {
            let mut target = "is_auth".to_string();
            if let Some(ref v) = node.value {
                target = v.trim_start_matches('$').to_string();
            }
            for c in &node.children {
                if c.name == "as" {
                    if let Some(ref cv) = c.value {
                        target = cv.trim_start_matches('$').to_string();
                    }
                }
            }
            let is_auth = is_auth_fn(scope);
            scope.set(&target, Value::Bool(is_auth));
            Ok(())
        }),
        SlotMeta {
            description: "Check if user is logged in.".to_string(),
            example: "auth.check: $is_logged_in".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: CAN & CANNOT (RBAC Stubs)
    // ==========================================
    engine.register(
        "can",
        Arc::new(|engine, ctx, node, scope| {
            let user_auth = scope.get("user").or_else(|| scope.get("auth")).unwrap_or(Value::Nil);
            if user_auth != Value::Nil {
                for child in &node.children {
                    if child.name == "do" {
                        return engine.execute(ctx, child, scope);
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Execute block if user has specific permission.".to_string(),
            example: "can: 'edit'\n  do: { ... }".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    engine.register(
        "cannot",
        Arc::new(|engine, ctx, node, scope| {
            let user_auth = scope.get("user").or_else(|| scope.get("auth")).unwrap_or(Value::Nil);
            if user_auth == Value::Nil {
                for child in &node.children {
                    if child.name == "do" {
                        return engine.execute(ctx, child, scope);
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Execute block if user does NOT have specific permission.".to_string(),
            example: "cannot: 'edit'\n  do: { ... }".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: JSON OUTPUT
    // ==========================================
    engine.register(
        "json",
        Arc::new(|engine, _ctx, node, scope| {
            let val = engine.resolve_shorthand_value(node, scope);
            let bytes_str = serialize_json(&val);
            println!("{}", bytes_str);
            Ok(())
        }),
        SlotMeta {
            description: "Outputs value as JSON.".to_string(),
            example: "json: $data".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: FORELSE
    // ==========================================
    engine.register(
        "forelse",
        Arc::new(|engine, ctx, node, scope| {
            let val = engine.resolve_shorthand_value(node, scope);
            let list = val.to_list();

            if list.is_empty() {
                for child in &node.children {
                    if child.name == "forelse_empty" || child.name == "empty" {
                        for sub_child in &child.children {
                            engine.execute(ctx, sub_child, scope)?;
                        }
                        return Ok(());
                    }
                }
            } else {
                // Run the standard for loop behavior
                // Synthesize a "for" node
                let for_node = Node {
                    name: "for".to_string(),
                    value: node.value.clone(),
                    children: node.children.iter()
                        .filter(|c| c.name == "as" || c.name == "do")
                        .cloned()
                        .collect(),
                    line: node.line,
                    col: node.col,
                    filename: node.filename.clone(),
                };
                return engine.execute(ctx, &for_node, scope);
            }
            Ok(())
        }),
        SlotMeta {
            description: "Loop with empty fallback block.".to_string(),
            example: "forelse: $list {\n  as: $item\n  do: { ... }\n  forelse_empty: { ... }\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );

    // ==========================================
    // SLOT: VALIDATION ERROR
    // ==========================================
    engine.register(
        "error",
        Arc::new(|engine, ctx, node, scope| {
            let field_name = node.value.clone().unwrap_or_default();
            if let Some(Value::Map(errs)) = scope.get("errors") {
                if let Some(err_val) = errs.get(&field_name) {
                    let first_msg = match err_val {
                        Value::List(l) => l.first().map(|v| v.to_string_coerce()).unwrap_or_default(),
                        Value::String(s) => s.clone(),
                        _ => err_val.to_string_coerce(),
                    };
                    if !first_msg.is_empty() {
                        scope.set("message", Value::String(first_msg));
                        for child in &node.children {
                            if child.name == "do" {
                                return engine.execute(ctx, child, scope);
                            }
                        }
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Show validation error for field.".to_string(),
            example: "error: 'username' {\n  do: {\n    log: $message\n  }\n}".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: String::new(),
        },
    );
}
