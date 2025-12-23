use crate::billing::entity::{
    PaymentAttemptEntity, PaymentAttemptState, PaymentMethod, TicketEntity, TicketState,
};
use chrono::Utc;
use sqlx::{Pool, Postgres};
use std::ops::DerefMut;
use uuid::Uuid;

#[derive(Clone)]
pub struct PaymentAttemptRepository {
    executor: Pool<Postgres>,
}

impl PaymentAttemptRepository {
    pub fn new(executor: Pool<Postgres>) -> Self {
        Self { executor }
    }

    pub async fn create_payment_attempt(
        &self,
        method: PaymentMethod,
        description: Option<String>,
        state: PaymentAttemptState,
        tickets: Vec<TicketEntity>,
    ) -> Result<PaymentAttemptEntity, sqlx::Error> {
        let mut tx = self.executor.begin().await?;
        let mut payment_attempt = PaymentAttemptEntity::new(method, description, state, tickets);

        sqlx::query!(
            "
            INSERT INTO payment_attempt (id, created_at, closed_at, method, description, state, total)
            VALUES ($1, $2, NULL, $3, $4, $5, $6)
            ",
            payment_attempt.id,
            payment_attempt.created_at,
            payment_attempt.method as _,
            payment_attempt.description,
            payment_attempt.state as _,
            payment_attempt.total
        )
            .fetch_optional(tx.deref_mut())
            .await
            .inspect_err(|e| tracing::error!("SQL error while inserting payment attempt: {e}"))?;

        tracing::info!("Inserted payment attempt with ID {}", payment_attempt.id);

        let ticket_ids: Vec<Uuid> = payment_attempt.tickets.iter().map(|t| t.id).collect();

        sqlx::query!(
            "UPDATE ticket SET state = $1 WHERE id =ANY($2::uuid[])",
            TicketState::Paying as _,
            &ticket_ids
        )
        .fetch_optional(tx.deref_mut())
        .await
        .inspect_err(|e| tracing::error!("SQL error while marking tickets as paying: {e}"))?;

        tracing::info!(
            "Updated tickets of payment attempt {} to state \"Paying\": {ticket_ids:?}",
            payment_attempt.id
        );

        sqlx::query!(
            "
            INSERT INTO payment_attempt_ticket (payment_attempt_id, ticket_id)
            SELECT $1, u.ticket_id FROM UNNEST($2::uuid[]) AS u(ticket_id)
            ",
            payment_attempt.id,
            &ticket_ids
        )
        .fetch_optional(tx.deref_mut())
        .await
        .inspect_err(|e| {
            tracing::error!("SQL error while inserting payment attempt tickets: {e}")
        })?;

        tracing::info!(
            "Inserted payment attempt ticket relations for payment attempt ID {}",
            payment_attempt.id
        );

        tx.commit()
            .await
            .inspect_err(|e| tracing::error!("SQL error while committing transaction: {e}"))?;

        tracing::info!(
            "Committed creation of payment attempt with ID {}",
            payment_attempt.id
        );

        payment_attempt
            .tickets
            .iter_mut()
            .for_each(|t| t.state = TicketState::Paying);

        Ok(payment_attempt)
    }

    pub async fn mark_as_success(
        &self,
        payment_attempt: &mut PaymentAttemptEntity,
    ) -> Result<(), sqlx::Error> {
        let mut tx = self.executor.begin().await?;
        let close_date = Utc::now();

        sqlx::query!(
            "UPDATE payment_attempt SET state = $1, closed_at = $2 WHERE id = $3",
            PaymentAttemptState::Successful as _,
            close_date,
            payment_attempt.id
        )
        .fetch_optional(tx.deref_mut())
        .await
        .inspect_err(|e| {
            tracing::error!("SQL error while marking payment attempt as successful: {e}")
        })?;

        tracing::info!(
            "Updated state of payment attempt with ID {} to \"Successful\"",
            payment_attempt.id
        );

        let ticket_ids: Vec<Uuid> = payment_attempt.tickets.iter().map(|t| t.id).collect();

        sqlx::query!(
            "UPDATE ticket SET state = $1, closed_at = $2 WHERE id =ANY($3::uuid[])",
            TicketState::Paid as _,
            close_date,
            &ticket_ids
        )
        .fetch_optional(tx.deref_mut())
        .await
        .inspect_err(|e| tracing::error!("SQL error while marking tickets as paid: {e}"))?;

        tracing::info!(
            "Updated tickets as paid for payment attempt ID {} ticket IDs {ticket_ids:?}",
            payment_attempt.id
        );

        tx.commit()
            .await
            .inspect_err(|e| tracing::error!("SQL error while committing transaction: {e}"))?;

        tracing::info!(
            "Committed marking successful of payment attempt with ID {}",
            payment_attempt.id
        );

        payment_attempt.state = PaymentAttemptState::Successful;
        payment_attempt.closed_at = Some(close_date);
        payment_attempt.tickets.iter_mut().for_each(|t| {
            t.state = TicketState::Paid;
            t.closed_at = Some(close_date);
        });

        Ok(())
    }

    pub async fn mark_as_failed(
        &self,
        payment_attempt: &mut PaymentAttemptEntity,
    ) -> Result<(), sqlx::Error> {
        let mut tx = self.executor.begin().await?;
        let close_date = Utc::now();

        sqlx::query!(
            "UPDATE payment_attempt SET state = $1, closed_at = $2 WHERE id = $3",
            PaymentAttemptState::Failed as _,
            close_date,
            payment_attempt.id
        )
        .fetch_optional(tx.deref_mut())
        .await
        .inspect_err(|e| {
            tracing::error!("SQL error while marking payment attempt as failed: {e}")
        })?;

        let ticket_ids: Vec<Uuid> = payment_attempt.tickets.iter().map(|t| t.id).collect();

        sqlx::query!(
            "UPDATE ticket SET state = $1 WHERE id =ANY($2::uuid[])",
            TicketState::Pending as _,
            &ticket_ids
        )
        .fetch_optional(tx.deref_mut())
        .await
        .inspect_err(|e| tracing::error!("SQL error while marking tickets as pending: {e}"))?;

        tx.commit()
            .await
            .inspect_err(|e| tracing::error!("SQL error while committing transaction: {e}"))?;

        tracing::info!(
            "Committed marking failure of payment attempt with ID {}",
            payment_attempt.id
        );

        payment_attempt.state = PaymentAttemptState::Failed;
        payment_attempt.closed_at = Some(close_date);
        payment_attempt.tickets.iter_mut().for_each(|t| {
            t.state = TicketState::Pending;
        });

        Ok(())
    }
}
