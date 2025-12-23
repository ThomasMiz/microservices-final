use std::env::VarError;
use axum::{
    Extension, Router,
    routing::{get, post},
};
use sqlx::{Pool, Postgres};

use crate::billing::handlers::folder_handlers::{
    close_folder_by_id_handler, delete_folder_by_id_handler,
};
use crate::billing::handlers::payment_handlers::make_payment_handler;
use crate::billing::handlers::ticket_handlers::delete_ticket_by_id_handler;
use crate::billing::repository::PaymentAttemptRepository;
use crate::billing::service::PaymentService;
use crate::billing::{
    handlers::{
        folder_handlers::{create_folder_handler, get_folder_by_id_handler},
        ticket_handlers::{
            create_ticket_handler, get_ticket_by_id_handler, search_tickets_handler,
        },
    },
    repository::{FolderRepository, TicketRepository},
    service::{FolderService, TicketService},
};

pub trait RegisterBillingEndpoints {
    fn register_billing_endpoints(self, db: Pool<Postgres>) -> Self;
}

fn get_payment_failure_chance() -> f64 {
    match std::env::var("PAYMENT_FAIL_CHANCE") {
        Ok(s) => {
            match s.trim().parse::<f64>() {
                Ok(v) => return v,
                Err(e) => {
                    tracing::error!("Could not parse PAYMENT_FAIL_CHANCE value as f64: {e}")
                }
            }
        }
        Err(VarError::NotPresent) => {},
        Err(e) => {
            tracing::error!("Failed to read PAYMENT_FAIL_CHANCE env var: {e:?}");
        }
    };

    // half of all payment attempts shall fail inshallah
    0.5
}

impl RegisterBillingEndpoints for Router {
    fn register_billing_endpoints(self, db: Pool<Postgres>) -> Self {
        let payment_failure_chance = get_payment_failure_chance();

        let folder_repo = FolderRepository::new(db.clone());
        let ticket_repo = TicketRepository::new(db.clone());
        let payment_repo = PaymentAttemptRepository::new(db.clone());

        let folder_service = FolderService::new(folder_repo.clone());
        let ticket_service = TicketService::new(folder_repo.clone(), ticket_repo.clone());
        let payment_service = PaymentService::new(ticket_repo, payment_repo, payment_failure_chance);

        self.route("/folders", post(create_folder_handler))
            .route(
                "/folders/{id}",
                get(get_folder_by_id_handler).delete(delete_folder_by_id_handler),
            )
            .route("/folders/{id}/close", post(close_folder_by_id_handler))
            .route(
                "/tickets",
                post(create_ticket_handler).get(search_tickets_handler),
            )
            .route(
                "/tickets/{id}",
                get(get_ticket_by_id_handler).delete(delete_ticket_by_id_handler),
            )
            .route("/payments", post(make_payment_handler))
            .layer(Extension(folder_service))
            .layer(Extension(ticket_service))
            .layer(Extension(payment_service))
    }
}
