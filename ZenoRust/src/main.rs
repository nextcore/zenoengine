use std::fs;
use std::env;
use std::net::SocketAddr;
use axum::{
    routing::post,
    Router,
    Json,
    extract::State,
};
use serde::{Deserialize, Serialize};
use sqlx::{AnyPool, any::AnyPoolOptions};
use tower_http::trace::TraceLayer;
use tower_http::cors::CorsLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

// Use modules from the library crate
use ZenoRust::parser::Parser;
use ZenoRust::evaluator::Evaluator;

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
        run_cli_mode(db_pool).await;
    }
}

async fn run_cli_mode(pool: Option<AnyPool>) {
    tracing::info!("ZenoEngine Rust Edition (2024) - CLI Mode (Async)");

    let file_path = "source/test.zl";
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

    let app = Router::new()
        .route("/execute", post(execute_script))
        .layer(TraceLayer::new_for_http()) // Logging Middleware
        .layer(CorsLayer::permissive()) // CORS Middleware
        .with_state(state);

    let addr: SocketAddr = addr_str.parse().expect("Invalid address");
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();

    // Graceful Shutdown
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
