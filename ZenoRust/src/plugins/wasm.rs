use wasmtime::*;
use wasmtime_wasi::preview2::{self, WasiCtx, WasiCtxBuilder, WasiView, ResourceTable};
use wasmtime_wasi::preview2::preview1::{self, WasiPreview1View, WasiPreview1Adapter};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use serde_json::Value;

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
             Module::new(&self.engine, &wat)?
        } else {
             Module::from_file(&self.engine, &path)?
        };
        self.modules.lock().unwrap().insert(name, module);
        Ok(())
    }

    pub async fn call(&self, plugin_name: &str, func_name: &str, _params: Value) -> Result<Value, String> {
        let module = {
            let modules = self.modules.lock().unwrap();
            modules.get(plugin_name).cloned().ok_or("Plugin not found")?
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

        // 1. Try generic function export (void -> void)
        if let Ok(func) = instance.get_typed_func::<(), ()>(&mut store, func_name) {
             func.call_async(&mut store, ()).await.map_err(|e| format!("Exec error: {}", e))?;
             return Ok(Value::Bool(true));
        }

        // 2. Try _start if requested
        if func_name == "_start" || func_name == "main" {
             let func = instance.get_typed_func::<(), ()>(&mut store, "_start")
                .or_else(|_| instance.get_typed_func::<(), ()>(&mut store, "main"));

             if let Ok(f) = func {
                 f.call_async(&mut store, ()).await.map_err(|e| format!("Exec error: {}", e))?;
                 return Ok(Value::Bool(true));
             }
        }

        Err(format!("Function '{}' not found or invalid signature (expected () -> ())", func_name))
    }
}
