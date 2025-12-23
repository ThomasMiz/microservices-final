use bigdecimal::BigDecimal;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::billing::entity::TicketEntity;

#[derive(sqlx::Type, Debug, Clone, Copy, PartialEq, Eq)]
#[sqlx(type_name = "PAYMENTATTEMPTSTATE", rename_all = "lowercase")]
#[derive(Serialize, Deserialize)]
pub enum PaymentAttemptState {
    Ongoing,
    Successful,
    Failed,
}

#[derive(sqlx::Type, Debug, Clone, Copy, PartialEq, Eq)]
#[sqlx(type_name = "PAYMENTMETHOD", rename_all = "lowercase")]
#[derive(Serialize, Deserialize)]
pub enum PaymentMethod {
    Cash,
    Credit,
    Debit,
    Dishwashing,
    Other,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct PaymentAttemptEntity {
    pub id: Uuid,
    pub created_at: DateTime<Utc>,
    pub closed_at: Option<DateTime<Utc>>,
    pub method: PaymentMethod,
    pub description: Option<String>,
    pub state: PaymentAttemptState,
    pub total: BigDecimal,
    pub tickets: Vec<TicketEntity>,
}

impl PaymentAttemptEntity {
    pub fn new(
        method: PaymentMethod,
        description: Option<String>,
        state: PaymentAttemptState,
        tickets: Vec<TicketEntity>,
    ) -> Self {
        let total = tickets.iter().map(|t| &t.total).sum();

        Self {
            id: Uuid::now_v7(),
            created_at: Utc::now(),
            closed_at: None,
            method,
            description,
            state,
            total,
            tickets,
        }
    }
}
