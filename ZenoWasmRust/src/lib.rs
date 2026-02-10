use wasm_bindgen::prelude::*;

#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = console)]
    fn log(s: &str);
}

#[wasm_bindgen]
pub fn greet() {
    log("Hello from ZenoWasmRust!");
}

#[wasm_bindgen]
pub fn zeno_version() -> String {
    // Ideally we'd get this from ZenoRust constant, but currently just return string
    "ZenoRust WASM v0.1.0".to_string()
}
