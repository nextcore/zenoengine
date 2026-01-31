use serde::{Deserialize, Serialize};
use serde_json::{json, Value};
use std::ffi::CString;
use std::os::raw::c_char;
use std::slice;

// ============================================================================
// PLUGIN METADATA
// ============================================================================

#[derive(Serialize)]
struct PluginMetadata {
    name: String,
    version: String,
    author: String,
    description: String,
    license: String,
}

#[derive(Serialize)]
struct SlotDefinition {
    name: String,
    description: String,
    example: String,
    inputs: Value,
}

#[derive(Deserialize)]
struct PluginRequest {
    slot_name: String,
    parameters: Value,
}

#[derive(Serialize)]
struct PluginResponse {
    success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
}

// ============================================================================
// WASM EXPORTS (Required by ZenoEngine)
// ============================================================================

#[no_mangle]
pub extern "C" fn plugin_init() -> *mut c_char {
    let metadata = PluginMetadata {
        name: "hello-rust".to_string(),
        version: "1.0.0".to_string(),
        author: "ZenoEngine Team".to_string(),
        description: "Hello World plugin written in Rust".to_string(),
        license: "MIT".to_string(),
    };

    alloc_json(&metadata)
}

#[no_mangle]
pub extern "C" fn plugin_register_slots() -> *mut c_char {
    let slots = vec![
        SlotDefinition {
            name: "rust.greet".to_string(),
            description: "Greet someone with a message (Rust edition)".to_string(),
            example: "rust.greet { name: 'World' }".to_string(),
            inputs: json!({
                "name": {
                    "type": "string",
                    "required": true,
                    "description": "Name to greet"
                }
            }),
        },
        SlotDefinition {
            name: "rust.calculate".to_string(),
            description: "Perform a simple calculation".to_string(),
            example: "rust.calculate { a: 10, b: 5, op: 'add' }".to_string(),
            inputs: json!({
                "a": {
                    "type": "number",
                    "required": true,
                    "description": "First number"
                },
                "b": {
                    "type": "number",
                    "required": true,
                    "description": "Second number"
                },
                "op": {
                    "type": "string",
                    "required": true,
                    "description": "Operation: add, sub, mul, div"
                }
            }),
        },
    ];

    alloc_json(&slots)
}

#[no_mangle]
pub extern "C" fn plugin_execute(
    slot_name_ptr: *const c_char,
    slot_name_len: usize,
    params_ptr: *const c_char,
    params_len: usize,
) -> *mut c_char {
    // Parse slot name
    let slot_name = unsafe {
        let bytes = slice::from_raw_parts(slot_name_ptr as *const u8, slot_name_len);
        String::from_utf8_lossy(bytes).to_string()
    };

    // Parse parameters JSON
    let params_json = unsafe {
        let bytes = slice::from_raw_parts(params_ptr as *const u8, params_len);
        String::from_utf8_lossy(bytes).to_string()
    };

    let request: PluginRequest = match serde_json::from_str(&params_json) {
        Ok(req) => req,
        Err(e) => {
            return alloc_json(&PluginResponse {
                success: false,
                data: None,
                error: Some(format!("Invalid request JSON: {}", e)),
            });
        }
    };

    // Route to appropriate handler
    let response = match slot_name.as_str() {
        "rust.greet" => execute_greet(&request.parameters),
        "rust.calculate" => execute_calculate(&request.parameters),
        _ => PluginResponse {
            success: false,
            data: None,
            error: Some(format!("Unknown slot: {}", slot_name)),
        },
    };

    alloc_json(&response)
}

#[no_mangle]
pub extern "C" fn plugin_cleanup() {
    // Nothing to cleanup for this simple plugin
}

#[no_mangle]
pub extern "C" fn alloc(size: usize) -> *mut u8 {
    let mut buf = Vec::with_capacity(size);
    let ptr = buf.as_mut_ptr();
    std::mem::forget(buf);
    ptr
}

// ============================================================================
// SLOT IMPLEMENTATIONS
// ============================================================================

fn execute_greet(params: &Value) -> PluginResponse {
    let name = params
        .get("name")
        .and_then(|v| v.as_str())
        .unwrap_or("World");

    let message = format!("Hello from Rust, {}! 🦀", name);

    // Log using host function
    call_host_log("info", &format!("Greeting: {}", message));

    PluginResponse {
        success: true,
        data: Some(json!({
            "message": message
        })),
        error: None,
    }
}

fn execute_calculate(params: &Value) -> PluginResponse {
    let a = params.get("a").and_then(|v| v.as_f64());
    let b = params.get("b").and_then(|v| v.as_f64());
    let op = params.get("op").and_then(|v| v.as_str());

    if a.is_none() || b.is_none() || op.is_none() {
        return PluginResponse {
            success: false,
            data: None,
            error: Some("Missing required parameters: a, b, op".to_string()),
        };
    }

    let a = a.unwrap();
    let b = b.unwrap();
    let op = op.unwrap();

    let result = match op {
        "add" => a + b,
        "sub" => a - b,
        "mul" => a * b,
        "div" => {
            if b == 0.0 {
                return PluginResponse {
                    success: false,
                    data: None,
                    error: Some("Division by zero".to_string()),
                };
            }
            a / b
        }
        _ => {
            return PluginResponse {
                success: false,
                data: None,
                error: Some(format!("Unknown operation: {}", op)),
            };
        }
    };

    call_host_log("info", &format!("Calculation: {} {} {} = {}", a, op, b, result));

    PluginResponse {
        success: true,
        data: Some(json!({
            "result": result
        })),
        error: None,
    }
}

// ============================================================================
// HOST FUNCTIONS (Imported from ZenoEngine)
// ============================================================================

#[link(wasm_import_module = "env")]
extern "C" {
    fn host_log(level_ptr: *const c_char, level_len: usize, msg_ptr: *const c_char, msg_len: usize);
}

fn call_host_log(level: &str, message: &str) {
    let level_bytes = level.as_bytes();
    let msg_bytes = message.as_bytes();

    unsafe {
        host_log(
            level_bytes.as_ptr() as *const c_char,
            level_bytes.len(),
            msg_bytes.as_ptr() as *const c_char,
            msg_bytes.len(),
        );
    }
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

fn alloc_json<T: Serialize>(value: &T) -> *mut c_char {
    let json_string = serde_json::to_string(value).unwrap_or_else(|_| "{}".to_string());
    let c_string = CString::new(json_string).unwrap();
    c_string.into_raw()
}
