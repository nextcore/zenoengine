use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_void};
use serde::{Deserialize, Serialize};
use serde_json::json;

// --- Host Function Imports ---

#[link(wasm_import_module = "env")]
extern "C" {
    fn host_log(level_ptr: *const c_char, level_len: u32, msg_ptr: *const c_char, msg_len: u32);
}

// --- Helper Functions ---

fn log(level: &str, message: &str) {
    unsafe {
        host_log(
            level.as_ptr() as *const c_char,
            level.len() as u32,
            message.as_ptr() as *const c_char,
            message.len() as u32,
        );
    }
}

// --- Plugin Exports ---

#[no_mangle]
pub extern "C" fn plugin_init() -> *mut c_char {
    let metadata = json!({
        "name": "robotics",
        "version": "1.0.0",
        "description": "High-performance robotics control"
    });
    CString::new(metadata.to_string()).unwrap().into_raw()
}

#[no_mangle]
pub extern "C" fn plugin_register_slots() -> *mut c_char {
    let slots = json!([
        {
            "name": "robot.sense",
            "description": "Read from robot sensors",
            "inputs": {
                "value": { "type": "string", "required": true, "description": "Sensor type (distance, battery)" },
                "as": { "type": "string", "required": false, "description": "Output variable" }
            }
        },
        {
            "name": "robot.act",
            "description": "Send command to actuators",
            "inputs": {
                "value": { "type": "string", "required": true, "description": "Action name" },
                "power": { "type": "int", "required": false, "description": "Actuator power level" }
            }
        }
    ]);
    CString::new(slots.to_string()).unwrap().into_raw()
}

#[derive(Deserialize)]
struct PluginRequest {
    slot_name: String,
    parameters: serde_json::Value,
}

#[no_mangle]
pub extern "C" fn plugin_execute(
    name_ptr: *const c_char, name_len: u32,
    params_ptr: *const c_char, params_len: u32
) -> *mut c_char {
    let slot_name = unsafe {
        let slice = std::slice::from_raw_parts(name_ptr as *const u8, name_len as usize);
        std::str::from_utf8(slice).unwrap()
    };

    let params_str = unsafe {
        let slice = std::slice::from_raw_parts(params_ptr as *const u8, params_len as usize);
        std::str::from_utf8(slice).unwrap()
    };

    let request: PluginRequest = serde_json::from_str(params_str).unwrap();

    let mut response = json!({ "success": true, "data": {} });

    match slot_name {
        "robot.sense" => {
            let sensor_type = request.parameters["value"].as_str().unwrap_or("unknown");
            let value = match sensor_type {
                "distance" => json!(120), // Simulated distance
                "battery" => json!(85),   // Simulated battery
                _ => json!("active")
            };
            log("info", &format!("🤖 [WASM Robot] Sensing {}: {}", sensor_type, value));
            response["data"]["sensor_data"] = value;
        },
        "robot.act" => {
            let action = request.parameters["value"].as_str().unwrap_or("none");
            let power = request.parameters["power"].as_i64().unwrap_or(100);
            log("info", &format!("⚙️ [WASM Robot] Actuating {}: {}%", action, power));
        },
        _ => {
            response = json!({ "success": false, "error": "Unknown slot" });
        }
    }

    CString::new(response.to_string()).unwrap().into_raw()
}

#[no_mangle]
pub extern "C" fn plugin_cleanup() {}

#[no_mangle]
pub extern "C" fn alloc(size: usize) -> *mut c_void {
    let mut buf = Vec::with_capacity(size);
    let ptr = buf.as_mut_ptr();
    std::mem::forget(buf);
    ptr as *mut c_void
}

fn main() {}
