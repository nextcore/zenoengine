use wasmtime::*;
use wasmtime_wasi::{WasiCtx, WasiCtxBuilder, WasiView};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use serde_json::Value;

struct PluginState {
    ctx: WasiCtx,
    table: ResourceTable,
}

impl WasiView for PluginState {
    fn ctx(&mut self) -> &mut WasiCtx { &mut self.ctx }
    fn table(&mut self) -> &mut ResourceTable { &mut self.table }
}

pub struct WasmManager {
    engine: Engine,
    linker: Linker<PluginState>,
    // For MVP, we store instantiation logic or instantiated stores?
    // WASM execution usually requires a Store per instance or execution.
    // We will load modules on demand or cache them.
    modules: Arc<Mutex<HashMap<String, Module>>>,
}

impl WasmManager {
    pub fn new() -> Result<Self, anyhow::Error> {
        let mut config = Config::new();
        config.async_support(true);
        let engine = Engine::new(&config)?;

        let mut linker = Linker::new(&engine);
        wasmtime_wasi::add_to_linker_async(&mut linker)?;

        Ok(Self {
            engine,
            linker,
            modules: Arc::new(Mutex::new(HashMap::new())),
        })
    }

    pub async fn load_plugin(&self, name: String, path: String) -> Result<(), anyhow::Error> {
        let module = Module::from_file(&self.engine, path)?;
        self.modules.lock().unwrap().insert(name, module);
        Ok(())
    }

    pub async fn call(&self, plugin_name: &str, func_name: &str, params: Value) -> Result<Value, String> {
        let module = {
            let modules = self.modules.lock().unwrap();
            modules.get(plugin_name).cloned().ok_or("Plugin not found")?
        };

        let wasi = WasiCtxBuilder::new().inherit_stdio().build();
        let state = PluginState { ctx: wasi, table: ResourceTable::new() };
        let mut store = Store::new(&self.engine, state);

        let instance = self.linker.instantiate_async(&mut store, &module).await
            .map_err(|e| format!("Failed to instantiate WASM: {}", e))?;

        let func = instance.get_typed_func::<(), ()>(&mut store, func_name)
            .map_err(|e| format!("Function '{}' not found: {}", func_name, e))?;

        // For MVP, we only support void -> void or simple number types to prove execution.
        // Complex JSON passing requires memory bindings (alloc/dealloc) which is complex for a quick port.
        // Alternatively, we use `wasi` to read inputs from env or stdin?
        // Let's assume the user just wants to run a function.

        func.call_async(&mut store, ()).await.map_err(|e| format!("Exec error: {}", e))?;

        Ok(Value::Bool(true))
    }
}
