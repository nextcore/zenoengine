mod php;

use std::io::{self, BufRead, Write};
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Serialize, Deserialize, Debug)]
struct RpcMessage {
    #[serde(default)]
    id: String,
    #[serde(default)]
    slot_name: String,
    #[serde(rename = "type", default = "default_type_fn")]
    msg_type: String,
    #[serde(flatten)]
    payload: Value,
}

fn default_type_fn() -> String {
    "legacy".to_string()
}

#[derive(Serialize)]
struct Response {
    #[serde(rename = "type", skip_serializing_if = "Option::is_none")]
    msg_type: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    id: Option<String>,
    success: bool,
    data: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
}

#[derive(Serialize)]
struct HostCall {
    #[serde(rename = "type")]
    msg_type: String,
    id: String,
    function: String,
    parameters: Value,
}

fn main() -> anyhow::Result<()> {
    // --- PHP Init ---
    if !php::init() {
        eprintln!("[Rust] Failed to initialize PHP Embed SAPI");
        std::process::exit(1);
    }

    // Ensure PHP shutdown on exit
    // Note: This simple defer might not run on process::exit, but works for normal loop exit.
    // Ideally use a wrapper struct with Drop impl or similar.
    defer_shutdown();

    let stdin = io::stdin();
    let mut stdout = io::stdout();
    let mut handle = stdin.lock();
    let mut buffer = String::new();

    while handle.read_line(&mut buffer)? > 0 {
        if buffer.trim().is_empty() {
            buffer.clear();
            continue;
        }

        // Catch parse errors to keep the loop alive
        let input: RpcMessage = match serde_json::from_str(&buffer) {
            Ok(msg) => msg,
            Err(e) => {
                eprintln!("[Rust] JSON Parse Error: {}", e);
                buffer.clear();
                continue;
            }
        };

        match input.slot_name.as_str() {
            "plugin_init" => {
                let resp = Response {
                    msg_type: None,
                    id: None,
                    success: true,
                    data: Some(serde_json::json!({
                        "name": "php-native",
                        "version": "1.3.0",
                        "description": "Rust-compiled PHP Bridge (Real Embed)"
                    })),
                    error: None,
                };
                writeln!(stdout, "{}", serde_json::to_string(&resp)?)?;
            },
            "plugin_register_slots" => {
                let resp = Response {
                    msg_type: None,
                    id: None,
                    success: true,
                    data: Some(serde_json::json!({
                        "slots": [
                            {"name": "php.run", "description": "Execute PHP code directly"},
                            {"name": "php.laravel", "description": "Invoke Laravel Artisan"},
                            {"name": "php.health", "description": "Check bridge health"},
                            {"name": "php.db_proxy", "description": "Execute DB query via Zeno pool"},
                            {"name": "php.crash", "description": "Simulate crash"}
                        ]
                    })),
                    error: None,
                };
                writeln!(stdout, "{}", serde_json::to_string(&resp)?)?;
            },
            "php.crash" => {
                eprintln!("[Rust] Simulating crash...");
                std::process::exit(1);
            },
            "php.health" => {
                let resp = Response {
                    msg_type: Some("guest_response".to_string()),
                    id: Some(input.id),
                    success: true,
                    data: Some(serde_json::json!({
                        "status": "healthy",
                        "uptime": "online",
                        "backend": "rust-embed"
                    })),
                    error: None,
                };
                writeln!(stdout, "{}", serde_json::to_string(&resp)?)?;
            },
            "php.db_proxy" => {
                // ... same mock implementation ...
                let host_call = HostCall {
                    msg_type: "host_call".to_string(),
                    id: "db1".to_string(),
                    function: "db_query".to_string(),
                    parameters: serde_json::json!({
                        "connection": "default",
                        "sql": "SELECT 1 as pool_check"
                    }),
                };
                writeln!(stdout, "{}", serde_json::to_string(&host_call)?)?;

                let resp = Response {
                    msg_type: Some("guest_response".to_string()),
                    id: Some(input.id),
                    success: true,
                    data: Some(serde_json::json!({
                        "message": "Query proxied to Zeno Pool"
                    })),
                    error: None,
                };
                writeln!(stdout, "{}", serde_json::to_string(&resp)?)?;
            },
            "php.run" | "php.laravel" => {
                // Determine entry point
                let entry_script = if input.slot_name == "php.laravel" {
                    // Laravel entry: artisan or public/index.php
                    // Assumes CWD is project root
                    "if (file_exists('artisan')) { require 'artisan'; } else { echo 'Laravel not found'; }"
                } else {
                    input.payload.get("code").and_then(|v| v.as_str()).unwrap_or("echo 'No code provided';")
                };

                // --- REQUEST LIFECYCLE START ---
                if !php::request_startup() {
                    eprintln!("[Rust] Request Startup Failed");
                    // Can continue but state might be dirty
                }

                // Inject Superglobals from payload
                let mut bootstrap = String::from("<?php ");

                // Extract Zeno Scope
                if let Some(scope) = input.payload.get("_zeno_scope") {
                    if let Ok(json) = serde_json::to_string(scope) {
                        bootstrap.push_str(&format!("$_SERVER['ZENO_SCOPE'] = json_decode('{}', true);", json.replace("'", "\\'")));
                    }
                }

                // Mock Request Data
                if let Some(req) = input.payload.get("request") {
                    if let Some(uri) = req.get("uri").and_then(|v| v.as_str()) {
                        bootstrap.push_str(&format!("$_SERVER['REQUEST_URI'] = '{}';", uri));
                    }
                    if let Some(method) = req.get("method").and_then(|v| v.as_str()) {
                        bootstrap.push_str(&format!("$_SERVER['REQUEST_METHOD'] = '{}';", method));
                    }
                }

                // Wrapper for capture
                let wrapped_code = format!(
                    "
                    ob_start();
                    try {{
                        // Bootstrap Logic
                        {}

                        // User Code
                        {}
                    }} catch (Throwable $e) {{
                        echo 'PHP Error: ' . $e->getMessage();
                    }}
                    $output = ob_get_clean();

                    // ALTERNATIVE: Write to a temporary file
                    $temp_file = sys_get_temp_dir() . '/zeno_php_' . getmypid() . '.out';
                    file_put_contents($temp_file, $output);
                    ",
                    bootstrap.trim_start_matches("<?php "),
                    entry_script
                );

                let success = php::eval(&wrapped_code);

                // --- REQUEST LIFECYCLE END ---
                php::request_shutdown();

                // Read back captured output
                let mut captured_output = String::new();
                let temp_path = std::env::temp_dir().join(format!("zeno_php_{}.out", std::process::id()));
                if temp_path.exists() {
                    if let Ok(content) = std::fs::read_to_string(&temp_path) {
                        captured_output = content;
                        let _ = std::fs::remove_file(temp_path);
                    }
                }

                let resp = Response {
                    msg_type: Some("guest_response".to_string()),
                    id: Some(input.id),
                    success: success,
                    data: Some(serde_json::json!({
                        "output": captured_output,
                        "status": if success { 200 } else { 500 }
                    })),
                    error: None,
                };
                writeln!(stdout, "{}", serde_json::to_string(&resp)?)?;
            },
            _ => {
                if input.msg_type != "host_response" {
                    let resp = Response {
                        msg_type: Some("guest_response".to_string()),
                        id: Some(input.id),
                        success: false,
                        data: None,
                        error: Some(format!("Unknown slot: {}", input.slot_name)),
                    };
                    writeln!(stdout, "{}", serde_json::to_string(&resp)?)?;
                }
            }
        }

        buffer.clear();
    }

    Ok(())
}

struct PhpGuard;
impl Drop for PhpGuard {
    fn drop(&mut self) {
        php::shutdown();
    }
}

fn defer_shutdown() -> PhpGuard {
    PhpGuard
}
