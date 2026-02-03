use std::ffi::{CString};
use std::os::raw::{c_char, c_void};
use serde::{Deserialize};
use serde_json::json;

// --- Plugin Exports ---

#[no_mangle]
pub extern "C" fn plugin_init() -> *mut c_char {
    let metadata = json!({
        "name": "vision-ocr",
        "version": "1.0.0",
        "description": "License Plate Recognition System"
    });
    CString::new(metadata.to_string()).unwrap().into_raw()
}

#[no_mangle]
pub extern "C" fn plugin_register_slots() -> *mut c_char {
    let slots = json!([
        {
            "name": "vision.ocr",
            "description": "Detect and read license plate from image",
            "inputs": {
                "source": { "type": "string", "required": true, "description": "Path to image or camera stream" },
                "as": { "type": "string", "required": false, "description": "Output variable" }
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
        "vision.ocr" => {
            let _source = request.parameters["source"].as_str().unwrap_or("");
            // In real world, call OpenCV/Tesseract here
            let plate = "B 1234 ABC";
            let confidence = 0.98;

            response["data"]["plate_number"] = json!(plate);
            response["data"]["confidence"] = json!(confidence);
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
