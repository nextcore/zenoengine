use wasmtime::*;
use wasmtime_wasi::preview2::{self, WasiCtx, WasiCtxBuilder, WasiView, ResourceTable};
use wasmtime_wasi::preview2::preview1::{self, WasiPreview1View, WasiPreview1Adapter};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use crate::evaluator::Value;

struct PluginState {
    ctx: WasiCtx,
    table: ResourceTable,
    p1_ctx: WasiPreview1Adapter,
}

impl WasiView for PluginState {
    fn ctx(&mut self) -> &mut WasiCtx { &mut self.ctx }
    fn table(&mut self) -> &mut ResourceTable { &mut self.table }
}

impl WasiPreview1View for PluginState {
    fn adapter(&self) -> &WasiPreview1Adapter { &self.p1_ctx }
    fn adapter_mut(&mut self) -> &mut WasiPreview1Adapter { &mut self.p1_ctx }
}

pub struct WasmManager {
    engine: Engine,
    linker: Linker<PluginState>,
    modules: Arc<Mutex<HashMap<String, Module>>>,
}

impl WasmManager {
    pub fn new() -> Result<Self, anyhow::Error> {
        let mut config = Config::new();
        config.async_support(true);
        let engine = Engine::new(&config)?;

        let mut linker = Linker::new(&engine);
        preview2::preview1::add_to_linker_async(&mut linker)?;

        Ok(Self {
            engine,
            linker,
            modules: Arc::new(Mutex::new(HashMap::new())),
        })
    }

    pub async fn load_plugin(&self, name: String, path: String) -> Result<(), anyhow::Error> {
        let module = if path.ends_with(".wat") {
             // For testing
             let wat = tokio::fs::read_to_string(&path).await?;
             match Module::new(&self.engine, &wat) {
                 Ok(m) => m,
                 Err(e) => {
                     // Try to parse WAT manually or debug print error details
                     return Err(anyhow::anyhow!("Failed to compile WAT '{}': {}", path, e));
                 }
             }
        } else {
             Module::from_file(&self.engine, &path)?
        };

        if let Ok(mut modules) = self.modules.lock() {
            modules.insert(name, module);
            Ok(())
        } else {
            Err(anyhow::anyhow!("Failed to acquire module lock"))
        }
    }

    pub async fn call(&self, plugin_name: &str, func_name: &str, params: Value) -> Result<Value, String> {
        let module = {
            if let Ok(modules) = self.modules.lock() {
                modules.get(plugin_name).cloned().ok_or("Plugin not found")?
            } else {
                return Err("Failed to acquire module lock".to_string());
            }
        };

        let wasi = WasiCtxBuilder::new().inherit_stdio().build();
        let state = PluginState {
            ctx: wasi,
            table: ResourceTable::new(),
            p1_ctx: WasiPreview1Adapter::new(),
        };
        let mut store = Store::new(&self.engine, state);

        // Instantiate
        let instance = self.linker.instantiate_async(&mut store, &module).await
            .map_err(|e| format!("Failed to instantiate WASM: {}", e))?;

        // 1. Try String ABI (ptr, len) -> (ptr, len) or void
        // We look for `alloc` or `malloc` to allocate memory for params
        let alloc_func = instance.get_typed_func::<(i32), i32>(&mut store, "alloc")
            .or_else(|_| instance.get_typed_func::<(i32), i32>(&mut store, "malloc")).ok();

        let memory = instance.get_memory(&mut store, "memory").ok_or("Module exports no memory")?;

        // If params is not Null and we have alloc, try to pass data
        if params != Value::Null && alloc_func.is_some() {
            let json_str = params.to_string();
            let bytes = json_str.as_bytes();
            let len = bytes.len() as i32;

            // Force unwrap safe because we checked alloc_func.is_some()
            let alloc = alloc_func.unwrap();
            let ptr = alloc.call_async(&mut store, len).await
                .map_err(|e| format!("Alloc error: {}", e))?;

            memory.write(&mut store, ptr as usize, bytes)
                .map_err(|e| format!("Memory write error: {}", e))?;

            // Try calling function with (ptr, len) -> i64 (packed ptr/len result)
            // This is a common pattern: high 32 bits = len, low 32 bits = ptr
            if let Ok(func) = instance.get_typed_func::<(i32, i32), i64>(&mut store, func_name) {
                let packed_result = func.call_async(&mut store, (ptr, len)).await
                    .map_err(|e| format!("Exec error: {}", e))?;

                let res_ptr = (packed_result & 0xFFFFFFFF) as usize;
                let res_len = (packed_result >> 32) as usize;

                let mut buf = vec![0u8; res_len];
                memory.read(&mut store, res_ptr, &mut buf)
                    .map_err(|e| format!("Memory read error: {}", e))?;

                let res_str = String::from_utf8_lossy(&buf).to_string();
                // Try to parse as JSON, else return String
                if let Ok(v) = serde_json::from_str::<serde_json::Value>(&res_str) {
                    return Ok(crate::evaluator::json_to_value(v));
                } else {
                    return Ok(Value::String(res_str));
                }
            }

            // Try calling function with (ptr, len) -> void
            if let Ok(func) = instance.get_typed_func::<(i32, i32), ()>(&mut store, func_name) {
                func.call_async(&mut store, (ptr, len)).await
                    .map_err(|e| format!("Exec error: {}", e))?;
                return Ok(Value::Boolean(true));
            }
        }

        // 2. Fallback: Try generic function export (void -> void)
        if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, func_name) {
             func.call_async(&mut store, ()).await.map_err(|e| format!("Exec error: {}", e))?;
             return Ok(Value::Boolean(true));
        }

        // 3. Try _start if requested
        if func_name == "_start" || func_name == "main" {
             let func = instance.get_typed_func::<(), ()>(&mut store, "_start")
                .or_else(|_| instance.get_typed_func::<(), ()>(&mut store, "main"));

             if let Ok(f) = func {
                 f.call_async(&mut store, ()).await.map_err(|e| format!("Exec error: {}", e))?;
                 return Ok(Value::Boolean(true));
             }
        }

        Err(format!("Function '{}' not found or invalid signature (expected (ptr, len) or ())", func_name))
    }
}
