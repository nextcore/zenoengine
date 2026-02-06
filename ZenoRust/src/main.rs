pub mod lexer;
pub mod parser;
pub mod evaluator;

use std::fs;
use std::env;
use std::net::SocketAddr;
use std::sync::Arc;
use axum::{
    routing::post,
    Router,
    Json,
    extract::State,
};
use serde::{Deserialize, Serialize};
use parser::Parser;
use evaluator::Evaluator;
use sqlx::{SqlitePool, sqlite::SqlitePoolOptions};

#[derive(Clone)]
struct AppState {
    db_pool: Option<SqlitePool>,
}

#[tokio::main]
async fn main() {
    let args: Vec<String> = env::args().collect();

    // Initialize DB Pool (In-memory for demo/testing)
    // In production, read from DATABASE_URL env var
    let db_pool = SqlitePoolOptions::new()
        .connect("sqlite::memory:")
        .await
        .ok(); // Ignore error for now, or panic if strict

    if args.len() > 1 && args[1] == "server" {
        start_server(db_pool).await;
    } else {
        run_cli_mode(db_pool).await;
    }
}

async fn run_cli_mode(pool: Option<SqlitePool>) {
    println!("ZenoEngine Rust Edition (2024) - CLI Mode (Async)");

    let file_path = "source/test.zl";
    println!("Executing ZenoLang script: {}", file_path);

    let contents = fs::read_to_string(file_path)
        .expect("Should have been able to read the file");

    let mut parser = Parser::new(&contents);
    let statements = parser.parse();

    let mut evaluator = Evaluator::new(pool);
    evaluator.eval(statements).await;
}

async fn start_server(pool: Option<SqlitePool>) {
    println!("ZenoEngine Rust Edition (2024) - Server Mode (Async)");
    println!("Listening on http://localhost:3000");

    let state = AppState { db_pool: pool };

    let app = Router::new()
        .route("/execute", post(execute_script))
        .with_state(state);

    let addr = SocketAddr::from(([127, 0, 0, 1], 3000));
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

#[derive(Deserialize)]
struct ScriptRequest {
    script: String,
}

#[derive(Serialize)]
struct ScriptResponse {
    output: String,
}

async fn execute_script(
    State(state): State<AppState>,
    Json(payload): Json<ScriptRequest>
) -> Json<ScriptResponse> {
    let mut parser = Parser::new(&payload.script);
    let statements = parser.parse();

    let mut evaluator = Evaluator::new(state.db_pool.clone());
    evaluator.eval(statements).await;

    Json(ScriptResponse {
        output: evaluator.get_output(),
    })
}
