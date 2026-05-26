use std::collections::HashMap;
use std::sync::{Arc, Mutex, OnceLock};
use std::path::Path;
use zenocore::{Engine, Node, Scope, SlotMeta, Diagnostic, Value};

use crate::transpiler::transpile_blade_native;

pub struct HtmlBuffer(pub Mutex<String>);

#[derive(Clone)]
pub struct SectionMap(pub Arc<Mutex<HashMap<String, Node>>>);

#[derive(Clone)]
pub struct StackMap(pub Arc<Mutex<HashMap<String, Vec<Node>>>>);

static BLADE_CACHE: OnceLock<Mutex<HashMap<String, Node>>> = OnceLock::new();

fn get_blade_cache() -> &'static Mutex<HashMap<String, Node>> {
    BLADE_CACHE.get_or_init(|| Mutex::new(HashMap::new()))
}

pub fn clear_blade_cache() {
    if let Some(cache) = BLADE_CACHE.get() {
        let mut guard = cache.lock().unwrap();
        guard.clear();
    }
}

fn get_cached_or_parse(full_path: &str) -> Result<Node, String> {
    let cache = get_blade_cache();
    {
        let guard = cache.lock().unwrap();
        if let Some(node) = guard.get(full_path) {
            return Ok(node.clone());
        }
    }

    let content = std::fs::read_to_string(full_path)
        .map_err(|e| format!("view not found: {}. Error: {}", full_path, e))?;
    let node = transpile_blade_native(&content, full_path)?;

    let mut guard = cache.lock().unwrap();
    guard.insert(full_path.to_string(), node.clone());
    Ok(node)
}

fn make_error(node: &Node, msg: String, slot: Option<String>) -> Diagnostic {
    Diagnostic {
        r#type: "error".to_string(),
        message: msg,
        filename: node.filename.clone(),
        line: node.line,
        col: node.col,
        slot,
    }
}

fn escape_html(s: &str) -> String {
    let mut escaped = String::with_capacity(s.len());
    for c in s.chars() {
        match c {
            '&' => escaped.push_str("&amp;"),
            '\'' => escaped.push_str("&#39;"),
            '"' => escaped.push_str("&quot;"),
            '<' => escaped.push_str("&lt;"),
            '>' => escaped.push_str("&gt;"),
            _ => escaped.push(c),
        }
    }
    escaped
}

fn get_view_root(scope: &Arc<Scope>) -> String {
    if let Some(val) = scope.get("_view_root") {
        let s = val.to_string_coerce();
        if !s.is_empty() {
            return s;
        }
    }
    "views".to_string()
}

fn ensure_blade_ext(path: &str) -> String {
    if !path.ends_with(".blade.zl") {
        format!("{}.blade.zl", path)
    } else {
        path.to_string()
    }
}

fn resolve_node_value(engine: &Engine, node: &Node, scope: &Arc<Scope>) -> Value {
    if let Some(ref val_str) = node.value {
        let dummy = Node {
            name: String::new(),
            value: Some(val_str.clone()),
            children: Vec::new(),
            line: node.line,
            col: node.col,
            filename: node.filename.clone(),
        };
        engine.resolve_shorthand_value(&dummy, scope)
    } else {
        Value::Nil
    }
}

pub fn register_blade_slots(eng: &mut Engine) {
    // 1. __native_write
    eng.register(
        "__native_write",
        Arc::new(|_engine, ctx, node, _scope| {
            if let Some(ref val) = node.value {
                if let Some(buf) = ctx.get::<HtmlBuffer>("httpWriter") {
                    let mut guard = buf.0.lock().unwrap();
                    guard.push_str(val);
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Writes raw string value to output buffer".to_string(),
            example: "__native_write: 'hello'".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "any".to_string(),
        },
    );

    // 2. __native_write_safe
    eng.register(
        "__native_write_safe",
        Arc::new(|engine, ctx, node, scope| {
            let val = resolve_node_value(engine, node, scope);
            if let Some(buf) = ctx.get::<HtmlBuffer>("httpWriter") {
                let mut guard = buf.0.lock().unwrap();
                guard.push_str(&escape_html(&val.to_string_coerce()));
            }
            Ok(())
        }),
        SlotMeta {
            description: "Writes HTML-escaped string value to output buffer".to_string(),
            example: "__native_write_safe: $name".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "any".to_string(),
        },
    );

    // 3. view.root
    eng.register(
        "view.root",
        Arc::new(|engine, _ctx, node, scope| {
            let val = resolve_node_value(engine, node, scope);
            let root_path = val.to_string_coerce();
            if root_path.is_empty() {
                return Err(make_error(node, "view.root path is required".to_string(), Some("view.root".to_string())));
            }
            scope.set("_view_root", Value::String(root_path));
            Ok(())
        }),
        SlotMeta {
            description: "Sets base directory for Blade views".to_string(),
            example: "view.root: 'apps/blog/resources/views'".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        },
    );

    // 4. view.blade
    eng.register(
        "view.blade",
        Arc::new(|engine, ctx, node, scope| {
            // Injects CSRF
            let token = ctx.get::<String>("csrf_token").map(|t| t.to_string()).unwrap_or_else(|| "".to_string());
            scope.set("csrf_token", Value::String(token.clone()));
            scope.set("csrf_field", Value::String(format!(r#"<input type="hidden" name="gorilla.csrf.Token" value="{}">"#, token)));

            let val = resolve_node_value(engine, node, scope);
            let view_file = val.to_string_coerce();
            if view_file.is_empty() {
                return Err(make_error(node, "view.blade view file is required".to_string(), Some("view.blade".to_string())));
            }

            let view_root = get_view_root(scope);
            let filename = ensure_blade_ext(&view_file);
            let full_path = Path::new(&view_root).join(filename);
            let full_path_str = full_path.to_string_lossy();

            let program_node = get_cached_or_parse(&full_path_str)
                .map_err(|e| make_error(node, e, Some("view.blade".to_string())))?;

            // Bind other children attributes to scope
            for child in &node.children {
                if child.name == "file" {
                    continue;
                }
                let val = engine.resolve_shorthand_value(child, scope);
                scope.set(&child.name, val);
            }

            // Ensure sections/stacks maps are initialized in context
            if ctx.get::<SectionMap>("__sections").is_none() {
                ctx.set("__sections", SectionMap(Arc::new(Mutex::new(HashMap::new()))));
            }
            if ctx.get::<StackMap>("__stacks").is_none() {
                ctx.set("__stacks", StackMap(Arc::new(Mutex::new(HashMap::new()))));
            }

            engine.execute(ctx, &program_node, scope)?;
            Ok(())
        }),
        SlotMeta {
            description: "Renders Blade view template".to_string(),
            example: "view.blade: 'welcome' { $title: 'Home' }".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        },
    );

    // 5. view.extends
    eng.register(
        "view.extends",
        Arc::new(|engine, ctx, node, scope| {
            let val = resolve_node_value(engine, node, scope);
            let layout_file = val.to_string_coerce();
            if layout_file.is_empty() {
                return Err(make_error(node, "view.extends layout file is required".to_string(), Some("view.extends".to_string())));
            }

            let view_root = get_view_root(scope);
            let filename = ensure_blade_ext(&layout_file);
            let full_path = Path::new(&view_root).join(filename);
            let full_path_str = full_path.to_string_lossy();

            let layout_node = get_cached_or_parse(&full_path_str)
                .map_err(|e| make_error(node, e, Some("view.extends".to_string())))?;

            engine.execute(ctx, &layout_node, scope)?;
            Ok(())
        }),
        SlotMeta {
            description: "Extends layout template".to_string(),
            example: "view.extends: 'layouts.app'".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        },
    );

    // 6. section.define
    eng.register(
        "section.define",
        Arc::new(|engine, ctx, node, scope| {
            let val = resolve_node_value(engine, node, scope);
            let name = val.to_string_coerce();
            let body = node.children.iter().find(|c| c.name == "do");
            if let Some(b) = body {
                if let Some(sections) = ctx.get::<SectionMap>("__sections") {
                    let mut guard = sections.0.lock().unwrap();
                    guard.insert(name, b.clone());
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Defines layout section content".to_string(),
            example: "section.define: 'content' { ... }".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        },
    );

    // 7. section.yield
    eng.register(
        "section.yield",
        Arc::new(|engine, ctx, node, scope| {
            let val = resolve_node_value(engine, node, scope);
            let name = val.to_string_coerce();
            if let Some(sections) = ctx.get::<SectionMap>("__sections") {
                let guard = sections.0.lock().unwrap();
                if let Some(body) = guard.get(&name) {
                    for child in &body.children {
                        engine.execute(ctx, child, scope)?;
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Yields defined layout section content".to_string(),
            example: "section.yield: 'content'".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        },
    );

    // 8. view.include
    eng.register(
        "view.include",
        Arc::new(|engine, ctx, node, scope| {
            let val = resolve_node_value(engine, node, scope);
            let view_file = val.to_string_coerce();
            if view_file.is_empty() {
                return Ok(());
            }

            let mut include_data = HashMap::new();
            if !node.children.is_empty() {
                let data_node = &node.children[0];
                if data_node.name == "data_map" {
                    for child in &data_node.children {
                        let val = engine.resolve_shorthand_value(child, scope);
                        include_data.insert(child.name.clone(), val);
                    }
                } else if data_node.name == "data_var" {
                    let val = engine.resolve_shorthand_value(data_node, scope);
                    if let Value::Map(m) = val {
                        include_data = m;
                    }
                }
            }

            let inner_scope = Scope::new(Some(scope.clone()));
            for (k, v) in include_data {
                inner_scope.set(&k, v);
            }

            let view_root = get_view_root(scope);
            let filename = ensure_blade_ext(&view_file);
            let full_path = Path::new(&view_root).join(filename);
            let full_path_str = full_path.to_string_lossy();

            let include_node = get_cached_or_parse(&full_path_str)
                .map_err(|e| make_error(node, e, Some("view.include".to_string())))?;

            engine.execute(ctx, &include_node, &inner_scope)?;
            Ok(())
        }),
        SlotMeta {
            description: "Includes partial template".to_string(),
            example: "view.include: 'partials.header' { $user: $user }".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        },
    );

    // 9. view.push
    eng.register(
        "view.push",
        Arc::new(|engine, ctx, node, scope| {
            let val = resolve_node_value(engine, node, scope);
            let name = val.to_string_coerce();
            let body = node.children.iter().find(|c| c.name == "do");
            if let Some(b) = body {
                if let Some(stacks) = ctx.get::<StackMap>("__stacks") {
                    let mut guard = stacks.0.lock().unwrap();
                    guard.entry(name).or_insert_with(Vec::new).push(b.clone());
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Pushes content block to a stack".to_string(),
            example: "view.push: 'scripts' { ... }".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        },
    );

    // 10. view.stack
    eng.register(
        "view.stack",
        Arc::new(|engine, ctx, node, scope| {
            let val = resolve_node_value(engine, node, scope);
            let name = val.to_string_coerce();
            if let Some(stacks) = ctx.get::<StackMap>("__stacks") {
                let guard = stacks.0.lock().unwrap();
                if let Some(nodes) = guard.get(&name) {
                    for n in nodes {
                        for child in &n.children {
                            engine.execute(ctx, child, scope)?;
                        }
                    }
                }
            }
            Ok(())
        }),
        SlotMeta {
            description: "Renders pushed stack content".to_string(),
            example: "view.stack: 'scripts'".to_string(),
            inputs: HashMap::new(),
            required_blocks: Vec::new(),
            value_type: "string".to_string(),
        },
    );
}

#[cfg(test)]
mod tests {
    use super::*;
    use zenocore::Context;
    use zenocore::parser::parse_string;

    #[test]
    fn test_slots_blade_flow() {
        let _ = std::fs::create_dir_all("test_views/layouts");
        std::fs::write("test_views/layouts/app.blade.zl", "Header\n@yield('content')\nFooter").unwrap();
        std::fs::write("test_views/home.blade.zl", "Welcome {{ $name }}!").unwrap();
        std::fs::write("test_views/post.blade.zl", "@extends('layouts/app')\n@section('content')\nPost content: {{ $title }}\n@endsection").unwrap();

        let mut engine = Engine::new();
        zenocore::slots::register_logic_slots(&mut engine);
        register_blade_slots(&mut engine);

        let mut ctx = Context::new();
        let buf = HtmlBuffer(Mutex::new(String::new()));
        ctx.set("httpWriter", buf);

        let scope = Scope::new(None);
        scope.set("_view_root", Value::String("test_views".to_string()));
        scope.set("name", Value::String("Budi".to_string()));
        scope.set("title", Value::String("My Post".to_string()));

        let node = parse_string("view.blade: 'post'", "test.zl").unwrap();
        engine.execute(&mut ctx, &node, &scope).unwrap();

        let writer = ctx.get::<HtmlBuffer>("httpWriter").unwrap();
        let output = writer.0.lock().unwrap().clone();
        println!("=== OUTPUT: {:?} ===", output);
        if let Some(sections) = ctx.get::<SectionMap>("__sections") {
            let guard = sections.0.lock().unwrap();
            println!("=== SECTIONS keys: {:?} ===", guard.keys().collect::<Vec<_>>());
        }

        assert!(output.contains("Header"));
        assert!(output.contains("Post content: My Post"));
        assert!(output.contains("Footer"));

        // Clean up
        let _ = std::fs::remove_dir_all("test_views");
    }
}
