use axum::{
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::net::SocketAddr;
use zenoengine::{
    apidoc::{self, APIRegistry, RouteDoc, ResponseDoc, RequestBodyDoc, MediaTypeDoc, SchemaDoc, Property},
    new_engine,
    parser::parse_string,
    executor::Context,
    scope::{Scope, Value},
};

#[derive(Deserialize)]
struct ExecuteRequest {
    script: String,
}

#[derive(Serialize)]
struct ExecuteResponse {
    success: bool,
    error: Option<String>,
    variables: serde_json::Value,
}

// Convert Zeno Engine Scope Values to standard JSON for web response
fn convert_value_to_json(val: &Value) -> serde_json::Value {
    match val {
        Value::Nil => serde_json::Value::Null,
        Value::String(s) => serde_json::Value::String(s.clone()),
        Value::Int(i) => serde_json::Value::Number(serde_json::Number::from(*i)),
        Value::Float(f) => {
            if let Some(n) = serde_json::Number::from_f64(*f) {
                serde_json::Value::Number(n)
            } else {
                serde_json::Value::Null
            }
        }
        Value::Bool(b) => serde_json::Value::Bool(*b),
        Value::List(l) => {
            let arr = l.iter().map(convert_value_to_json).collect();
            serde_json::Value::Array(arr)
        }
        Value::Map(m) => {
            let mut map = serde_json::Map::new();
            for (k, v) in m {
                map.insert(k.clone(), convert_value_to_json(v));
            }
            serde_json::Value::Object(map)
        }
    }
}

// GET /openapi.json
async fn handle_openapi_json() -> Json<serde_json::Value> {
    let spec = APIRegistry::global().generate_openapi();
    Json(spec)
}

// GET /docs (Swagger UI)
async fn handle_swagger_ui() -> axum::response::Html<String> {
    let html = apidoc::swagger_ui_html("/openapi.json");
    axum::response::Html(html)
}

// POST /execute (evaluates a ZenoLang script)
async fn handle_execute(Json(payload): Json<ExecuteRequest>) -> Json<ExecuteResponse> {
    let engine = new_engine();
    let mut ctx = Context::new();
    let scope = Scope::new(None);

    match parse_string(&payload.script, "request.zl") {
        Ok(root) => {
            match engine.execute(&mut ctx, &root, &scope) {
                Ok(_) => {
                    // Collect all variables from scope into a JSON map
                    let scope_map = scope.to_map();
                    let mut json_map = serde_json::Map::new();
                    for (k, v) in scope_map {
                        json_map.insert(k, convert_value_to_json(&v));
                    }
                    Json(ExecuteResponse {
                        success: true,
                        error: None,
                        variables: serde_json::Value::Object(json_map),
                    })
                }
                Err(err) => {
                    Json(ExecuteResponse {
                        success: false,
                        error: Some(format!("Execution error: {}", err.message)),
                        variables: serde_json::Value::Object(serde_json::Map::new()),
                    })
                }
            }
        }
        Err(err) => {
            Json(ExecuteResponse {
                success: false,
                error: Some(format!("Parsing error at {}:{}: {}", err.line, err.col, err.message)),
                variables: serde_json::Value::Object(serde_json::Map::new()),
            })
        }
    }
}

#[tokio::main]
async fn main() {
    // 1. Initialize API Documentation Registry
    let registry = APIRegistry::global();
    {
        let mut title = registry.title.write().unwrap();
        *title = "ZenoEngine Axum Web Server".to_string();
        let mut desc = registry.description.write().unwrap();
        *desc = "A real-world project example showing zenoengine-rs embedded inside an Axum API server.".to_string();
    }

    // 2. Register endpoint documentation
    let mut execute_responses = HashMap::new();
    execute_responses.insert("200".to_string(), ResponseDoc {
        description: "Evaluation results containing the scope variables".to_string(),
    });

    let mut execute_properties = HashMap::new();
    execute_properties.insert("script".to_string(), Property {
        r#type: "string".to_string(),
    });

    let mut content_map = HashMap::new();
    content_map.insert("application/json".to_string(), MediaTypeDoc {
        schema: SchemaDoc {
            r#type: "object".to_string(),
            properties: execute_properties,
        },
    });

    registry.register("POST", "/execute", RouteDoc {
        method: "POST".to_string(),
        path: "/execute".to_string(),
        summary: "Execute ZenoLang Script".to_string(),
        description: "Evaluates the provided script code and returns the variables populated in the script scope.".to_string(),
        tags: vec!["Execution".to_string()],
        params: Vec::new(),
        request_body: Some(RequestBodyDoc {
            content: content_map,
        }),
        responses: execute_responses,
    });

    // 3. Setup Axum Router
    let app = Router::new()
        .route("/openapi.json", get(handle_openapi_json))
        .route("/docs", get(handle_swagger_ui))
        .route("/execute", post(handle_execute));

    // 4. Start Server
    let addr = SocketAddr::from(([127, 0, 0, 1], 3000));
    println!("🚀 ZenoEngine Axum web server running at http://{}", addr);
    println!("📖 Swagger UI documentation available at http://{}/docs", addr);
    
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
