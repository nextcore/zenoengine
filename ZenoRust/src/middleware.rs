use axum::{
    body::Body,
    extract::Request,
    http::{StatusCode, Response},
    middleware::Next,
    response::IntoResponse,
};
use std::env;
use std::collections::HashSet;
use std::sync::OnceLock;

static BLOCKED_IPS: OnceLock<HashSet<String>> = OnceLock::new();

pub async fn ip_blocker(req: Request, next: Next) -> Result<impl IntoResponse, (StatusCode, String)> {
    let blocked = BLOCKED_IPS.get_or_init(|| {
        let raw = env::var("BLOCKED_IPS").unwrap_or_default();
        raw.split(',')
           .map(|s| s.trim().to_string())
           .filter(|s| !s.is_empty())
           .collect()
    });

    // Extract IP (Simplified: assumes ConnectInfo or X-Forwarded-For handled by upstream/Axum)
    // For now, we check X-Forwarded-For header manually if present, or just proceed.
    // Axum extract::ConnectInfo requires setup.
    // Let's check headers.

    if let Some(forwarded) = req.headers().get("x-forwarded-for") {
        if let Ok(ip_str) = forwarded.to_str() {
            // Take first IP in comma separated list
            let client_ip = ip_str.split(',').next().unwrap_or("").trim();
            if blocked.contains(client_ip) {
                tracing::warn!("Blocked request from IP: {}", client_ip);
                return Err((StatusCode::FORBIDDEN, "Access Denied".to_string()));
            }
        }
    }

    Ok(next.run(req).await)
}

pub async fn security_headers(req: Request, next: Next) -> impl IntoResponse {
    let mut response = next.run(req).await;
    let headers = response.headers_mut();

    headers.insert("X-Content-Type-Options", "nosniff".parse().unwrap());
    headers.insert("X-Frame-Options", "DENY".parse().unwrap());
    headers.insert("X-XSS-Protection", "1; mode=block".parse().unwrap());
    headers.insert("X-Powered-By", "ZenoRust".parse().unwrap());

    response
}
