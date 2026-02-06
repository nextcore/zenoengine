pub mod lexer;
pub mod parser;
pub mod evaluator;

use std::fs;
use std::env;
use std::net::SocketAddr;
use axum::{
    routing::post,
    Router,
    Json,
};
use serde::{Deserialize, Serialize};
use parser::Parser;
use evaluator::Evaluator;

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() > 1 && args[1] == "server" {
        tokio::runtime::Builder::new_multi_thread()
            .enable_all()
            .build()
            .unwrap()
            .block_on(start_server());
    } else {
        run_cli_mode();
    }
}

fn run_cli_mode() {
    println!("ZenoEngine Rust Edition (2024) - CLI Mode");

    let file_path = "source/test.zl";
    println!("Executing ZenoLang script: {}", file_path);

    let contents = fs::read_to_string(file_path)
        .expect("Should have been able to read the file");

    let mut parser = Parser::new(&contents);
    let statements = parser.parse();

    let mut evaluator = Evaluator::new();
    evaluator.eval(statements);
}

async fn start_server() {
    println!("ZenoEngine Rust Edition (2024) - Server Mode");
    println!("Listening on http://localhost:3000");

    let app = Router::new()
        .route("/execute", post(execute_script));

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

async fn execute_script(Json(payload): Json<ScriptRequest>) -> Json<ScriptResponse> {
    let output = tokio::task::spawn_blocking(move || {
        let mut parser = Parser::new(&payload.script);
        let statements = parser.parse();

        let mut evaluator = Evaluator::new();
        evaluator.eval(statements);
        evaluator.get_output()
    })
    .await
    .unwrap();

    Json(ScriptResponse {
        output,
    })
}
