use std::process::Stdio;
use tokio::process::Command;
use tokio::io::{BufReader, AsyncBufReadExt, AsyncWriteExt};
use tokio::sync::{Mutex, oneshot, mpsc};
use std::collections::HashMap;
use std::sync::Arc;
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Serialize, Deserialize)]
pub struct PluginRequest {
    #[serde(rename = "type")]
    pub msg_type: String,
    pub id: String,
    pub slot_name: String,
    pub parameters: Value,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PluginResponse {
    pub success: bool,
    pub data: Value,
    pub error: Option<String>,
}

#[derive(Clone)]
pub struct PluginHandle {
    tx: mpsc::Sender<(PluginRequest, oneshot::Sender<PluginResponse>)>,
}

pub struct SidecarManager {
    plugins: Arc<Mutex<HashMap<String, PluginHandle>>>,
}

impl SidecarManager {
    pub fn new() -> Self {
        Self {
            plugins: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub async fn load_plugin(&self, name: String, binary: String, work_dir: String) -> Result<(), String> {
        let mut cmd = Command::new(&binary);
        cmd.current_dir(&work_dir);
        cmd.env("ZENO_PLUGIN_NAME", &name);
        cmd.env("ZENO_SIDECAR", "true");
        cmd.stdin(Stdio::piped());
        cmd.stdout(Stdio::piped());
        cmd.stderr(Stdio::inherit());

        let mut child = cmd.spawn().map_err(|e| format!("Failed to spawn sidecar '{}': {}", binary, e))?;
        let mut stdin = child.stdin.take().ok_or("Failed to open stdin")?;
        let stdout = child.stdout.take().ok_or("Failed to open stdout")?;

        let (tx, mut rx) = mpsc::channel::<(PluginRequest, oneshot::Sender<PluginResponse>)>(32);
        let pending_calls: Arc<Mutex<HashMap<String, oneshot::Sender<PluginResponse>>>> = Arc::new(Mutex::new(HashMap::new()));

        // Writer Task
        let pending_calls_writer = pending_calls.clone();
        tokio::spawn(async move {
            while let Some((req, resp_tx)) = rx.recv().await {
                // Register pending call
                {
                    let mut calls = pending_calls_writer.lock().await;
                    calls.insert(req.id.clone(), resp_tx);
                }

                if let Ok(json) = serde_json::to_string(&req) {
                    if stdin.write_all(json.as_bytes()).await.is_err() { break; }
                    if stdin.write_all(b"\n").await.is_err() { break; }
                    if stdin.flush().await.is_err() { break; }
                }
            }
        });

        // Reader Task
        let pending_calls_reader = pending_calls.clone();
        tokio::spawn(async move {
            let reader = BufReader::new(stdout);
            let mut lines = reader.lines();

            while let Ok(Some(line)) = lines.next_line().await {
                if let Ok(msg) = serde_json::from_str::<serde_json::Value>(&line) {
                    if let Some(msg_type) = msg.get("type").and_then(|v| v.as_str()) {
                        if msg_type == "guest_response" {
                            if let Some(id) = msg.get("id").and_then(|v| v.as_str()) {
                                let mut calls = pending_calls_reader.lock().await;
                                if let Some(tx) = calls.remove(id) {
                                    let success = msg.get("success").and_then(|v| v.as_bool()).unwrap_or(false);
                                    let data = msg.get("data").cloned().unwrap_or(Value::Null);
                                    let error = msg.get("error").and_then(|v| v.as_str()).map(|s| s.to_string());

                                    let _ = tx.send(PluginResponse { success, data, error });
                                }
                            }
                        } else if msg_type == "host_call" {
                            eprintln!("DEBUG: Host Call received: {}", line);
                            // TODO: Implement Host Call Loopback
                        }
                    }
                }
            }
        });

        let mut plugins = self.plugins.lock().await;
        plugins.insert(name, PluginHandle { tx });

        Ok(())
    }

    pub async fn call(&self, plugin_name: &str, method: &str, params: Value) -> Result<Value, String> {
        let handle = {
            let plugins = self.plugins.lock().await;
            plugins.get(plugin_name).cloned().ok_or_else(|| format!("Plugin '{}' not found", plugin_name))?
        };

        let id = uuid::Uuid::new_v4().to_string();
        let request = PluginRequest {
            msg_type: "guest_call".to_string(),
            id,
            slot_name: method.to_string(),
            parameters: params,
        };

        let (resp_tx, resp_rx) = oneshot::channel();

        handle.tx.send((request, resp_tx)).await.map_err(|_| "Plugin process died".to_string())?;

        let response = resp_rx.await.map_err(|_| "Plugin response channel closed".to_string())?;

        if response.success {
            Ok(response.data)
        } else {
            Err(response.error.unwrap_or_else(|| "Unknown plugin error".to_string()))
        }
    }
}
