use axum::{Extension, Json, http::StatusCode};
use serde::{Deserialize, Serialize};
use sqlx::{Pool, Postgres};

#[derive(Serialize, Deserialize)]
pub struct HealthCheckMessage {
    message: String,
}

struct SqlHealthCheck {
    _value: Option<i32>,
}

fn ok_response(message: String) -> (StatusCode, Json<HealthCheckMessage>) {
    (StatusCode::OK, Json(HealthCheckMessage { message }))
}

fn internal_error_response(message: String) -> (StatusCode, Json<HealthCheckMessage>) {
    (
        StatusCode::INTERNAL_SERVER_ERROR,
        Json(HealthCheckMessage { message }),
    )
}

pub async fn health_check_handler(
    Extension(pool): Extension<Pool<Postgres>>,
) -> (StatusCode, Json<HealthCheckMessage>) {
    let db_check = sqlx::query_as!(SqlHealthCheck, "SELECT 1 AS _value")
        .fetch_one(&pool)
        .await;

    if let Err(e) = db_check {
        let message = format!("Database error: {e}");
        tracing::error!("Database health check failed! {message}");
        return internal_error_response(message);
    }

    ok_response(String::from("All systems functional!"))
}
