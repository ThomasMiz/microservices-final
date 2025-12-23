use crate::billing::entity::PaymentMethod;
use crate::billing::handlers::model_json::MessageJson;
use crate::billing::service::{MakePaymentError, PaymentService};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::{Extension, Json};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Serialize, Deserialize)]
pub struct MakePaymentJson {
    payment_method: PaymentMethod,
    description: Option<String>,
    tickets: Vec<Uuid>,
}

#[tracing::instrument]
pub async fn make_payment_handler(
    Extension(payment_service): Extension<PaymentService>,
    Json(body): Json<MakePaymentJson>,
) -> impl IntoResponse {
    let result = payment_service
        .make_payment(body.payment_method, body.description, body.tickets)
        .await;

    let (message, status) = match result {
        Ok(payment) => (
            MessageJson::new(format!(
                "Your deeds have been paid: payment ID {}",
                payment.id
            )),
            StatusCode::OK,
        ),
        Err(MakePaymentError::PaymentProcessingError(payment)) => (
            MessageJson::new(format!(
                "Error processing payment: payment ID {}",
                payment.id
            )),
            StatusCode::SERVICE_UNAVAILABLE,
        ),
        Err(MakePaymentError::InternalError) => (
            MessageJson::from("I fucked up bro"),
            StatusCode::INTERNAL_SERVER_ERROR,
        ),
        Err(MakePaymentError::TicketAlreadyPaid(paid_tickets)) => (
            MessageJson::new(format!(
                "These tickets have already been paid: {}",
                paid_tickets
                    .iter()
                    .map(|t| t.to_string())
                    .collect::<Vec<_>>()
                    .join(", ")
            )),
            StatusCode::BAD_REQUEST,
        ),
        Err(MakePaymentError::TicketNotFound(not_found)) => (
            MessageJson::new(format!(
                "Tickets not found: {}",
                not_found
                    .iter()
                    .map(|t| t.to_string())
                    .collect::<Vec<_>>()
                    .join(", ")
            )),
            StatusCode::BAD_REQUEST,
        ),
    };

    (status, Json(message))
}
