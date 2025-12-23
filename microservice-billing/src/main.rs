use axum::{Extension, Router, routing::get};
use axum_tracing_opentelemetry::middleware::{OtelAxumLayer, OtelInResponseLayer};
use tracing::{error, info};

use crate::billing::router::RegisterBillingEndpoints;

#[cfg(test)]
mod test_base_path;
#[cfg(test)]
mod tests;

mod billing;
mod database;
mod healthcheck;
mod tracing_config;
mod migrations;

#[tokio::main]
async fn main() -> Result<(), ()> {
    let _guard = tracing_config::init_logging();

    let listen_at = std::env::var("LISTEN_AT").unwrap_or_else(|_| String::from("0.0.0.0:5000"));
    let base_path = std::env::var("BASE_PATH").unwrap_or_else(|_| "/api/billing".to_string());

    let db_pool = match database::connect_to_db().await {
        Ok(pool) => pool,
        Err(db_error) => {
            error!("{db_error}");
            panic!("{db_error}");
        }
    };

    migrations::run_migrations(&db_pool).await;

    let app = Router::new()
        .nest(
            &base_path,
            Router::new().register_billing_endpoints(db_pool.clone()),
        )
        .layer(OtelInResponseLayer::default())
        .layer(OtelAxumLayer::default())
        .route("/health", get(healthcheck::health_check_handler))
        .layer(Extension(db_pool));

    let listener = match tokio::net::TcpListener::bind(&listen_at).await {
        Ok(l) => l,
        Err(e) => {
            error!("Could not bind TCP socket \"{listen_at}\": {e}");
            panic!("Could not bind TCP socket \"{listen_at}\": {e}");
        }
    };

    info!("Server listening on {0}", listen_at);
    axum::serve(listener, app).await.unwrap();

    Ok(())
}
