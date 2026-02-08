use std::fs;
use std::env;
use std::net::SocketAddr;
use axum::{
    routing::any,
    Router,
    extract::{State, Request},
    body::Body,
    response::{Response, IntoResponse},
    http::{StatusCode, HeaderMap},
};
use serde::{Deserialize, Serialize};
use sqlx::{AnyPool, any::AnyPoolOptions};
use tower_http::trace::TraceLayer;
use tower_http::cors::CorsLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
use axum::body::Bytes;

// Use modules from the library crate
use ZenoRust::parser::Parser;
use ZenoRust::evaluator::{Evaluator, Value}; // Need Value for extracting response

#[derive(Clone)]
struct AppState {
    db_pool: Option<AnyPool>,
}

#[tokio::main]
async fn main() {
    // 1. Initialize Logging
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::new(
            std::env::var("RUST_LOG").unwrap_or_else(|_| "info".to_string()),
        ))
        .with(tracing_subscriber::fmt::layer())
        .init();

    // 2. Load .env
    dotenvy::dotenv().ok();

    let args: Vec<String> = env::args().collect();

    // 3. Initialize DB Pool
    sqlx::any::install_default_drivers();
    let db_url = env::var("DATABASE_URL").unwrap_or_else(|_| "sqlite://zeno.db?mode=rwc".to_string());

    let db_pool = AnyPoolOptions::new()
        .connect(&db_url)
        .await
        .ok();

    if db_pool.is_none() {
        tracing::warn!("Failed to connect to database at '{}'. Running without DB support.", db_url);
    } else {
        tracing::info!("Connected to database: {}", db_url);
    }

    if args.len() > 1 && args[1] == "server" {
        start_server(db_pool).await;
    } else {
        let file_path = if args.len() > 1 { &args[1] } else { "source/test.zl" };
        run_cli_mode(db_pool, file_path).await;
    }
}

async fn run_cli_mode(pool: Option<AnyPool>, file_path: &str) {
    tracing::info!("ZenoEngine Rust Edition (2024) - CLI Mode (Async)");
    tracing::info!("Executing ZenoLang script: {}", file_path);

    let contents = match fs::read_to_string(file_path) {
        Ok(c) => c,
        Err(e) => {
            tracing::error!("Could not read file {}: {}", file_path, e);
            return;
        }
    };

    let mut parser = Parser::new(&contents);
    let statements = parser.parse();

    let mut evaluator = Evaluator::new(pool);
    evaluator.eval(statements).await;
}

async fn start_server(pool: Option<AnyPool>) {
    let port = env::var("PORT").unwrap_or_else(|_| "3000".to_string());
    let addr_str = format!("0.0.0.0:{}", port);

    tracing::info!("ZenoEngine Rust Edition (2024) - Server Mode (Async)");
    tracing::info!("Listening on http://{}", addr_str);

    let state = AppState { db_pool: pool };

    // Wildcard route to handle everything
    let app = Router::new()
        .route("/{*path}", any(handle_request))
        .route("/", any(handle_request)) // Handle root explicitly if *path doesn't match empty
        .layer(TraceLayer::new_for_http())
        .layer(CorsLayer::permissive())
        .with_state(state);

    let addr: SocketAddr = addr_str.parse().expect("Invalid address");
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();

    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await
        .unwrap();
}

async fn shutdown_signal() {
    let ctrl_c = async {
        tokio::signal::ctrl_c()
            .await
            .expect("failed to install Ctrl+C handler");
    };

    #[cfg(unix)]
    let terminate = async {
        tokio::signal::unix::signal(tokio::signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    #[cfg(not(unix))]
    let terminate = std::future::pending::<()>();

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }

    tracing::info!("Signal received, starting graceful shutdown...");
}

// Universal Handler
async fn handle_request(
    State(state): State<AppState>,
    req: Request<Body>,
) -> Response {
    let (parts, body) = req.into_parts();
    let path = parts.uri.path().to_string();
    let method = parts.method.to_string();
    let query_str = parts.uri.query().unwrap_or("").to_string();

    // Read body as bytes -> string
    let body_bytes = axum::body::to_bytes(body, usize::MAX).await.unwrap_or(Bytes::new());
    let body_str = String::from_utf8_lossy(&body_bytes).to_string();

    // Entrypoint: source/index.zl
    let entrypoint = "source/index.zl";
    let contents = match fs::read_to_string(entrypoint) {
        Ok(c) => c,
        Err(e) => {
            return (
                StatusCode::INTERNAL_SERVER_ERROR,
                format!("Could not load entrypoint {}: {}", entrypoint, e)
            ).into_response();
        }
    };

    let mut parser = Parser::new(&contents);
    let statements = parser.parse();

    let mut evaluator = Evaluator::new(state.db_pool.clone());

    // Inject Request Context
    evaluator.set_request_context(&method, &path, &query_str, &body_str);

    evaluator.eval(statements).await;

    // Check if JSON response was set
    if let Some((code, json_body)) = evaluator.get_final_response() {
        // Return JSON
        let status = StatusCode::from_u16(code as u16).unwrap_or(StatusCode::OK);
        let json_str = serde_json::to_string(&json_body).unwrap_or_else(|_| "{}".to_string());

        return (
            status,
            [("Content-Type", "application/json")],
            json_str
        ).into_response();
    }

    // Default: Return text output (stdout)
    (StatusCode::OK, evaluator.get_output()).into_response()
}
