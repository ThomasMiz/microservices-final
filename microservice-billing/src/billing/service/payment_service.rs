use crate::billing::entity::{
    PaymentAttemptEntity, PaymentAttemptState, PaymentMethod, TicketState,
};
use crate::billing::repository::{PaymentAttemptRepository, TicketRepository};
use std::collections::HashSet;
use std::fmt::Debug;
use uuid::Uuid;

#[derive(Clone)]
pub struct PaymentService {
    ticket_repo: TicketRepository,
    payment_repo: PaymentAttemptRepository,
    failure_chance: f64,
}

impl Debug for PaymentService {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "PaymentService")
    }
}

impl PaymentService {
    pub fn new(
        ticket_repo: TicketRepository,
        payment_repo: PaymentAttemptRepository,
        failure_chance: f64,
    ) -> PaymentService {
        Self {
            ticket_repo,
            payment_repo,
            failure_chance
        }
    }
}

pub enum MakePaymentError {
    TicketNotFound(HashSet<Uuid>),
    TicketAlreadyPaid(HashSet<Uuid>),
    PaymentProcessingError(PaymentAttemptEntity),
    InternalError,
}

impl PaymentService {
    #[tracing::instrument]
    pub async fn make_payment(
        &self,
        method: PaymentMethod,
        description: Option<String>,
        tickets: Vec<Uuid>,
    ) -> Result<PaymentAttemptEntity, MakePaymentError> {
        tracing::info!(
            "Attempting to make payments for tickets: {tickets:?} with method {method:?}"
        );

        let ticket_entities = self
            .ticket_repo
            .get_all_tickets_by_ids(&tickets)
            .await
            .map_err(|_| MakePaymentError::InternalError)?;

        if ticket_entities.len() != tickets.len() {
            let mut missing = HashSet::new();
            for t in tickets.iter() {
                missing.insert(*t);
            }

            for t in ticket_entities {
                missing.remove(&t.id);
            }

            tracing::warn!("Attempted to make payment for non-existent tickets: {missing:?}");
            return Err(MakePaymentError::TicketNotFound(missing));
        }

        let mut paid_tickets = HashSet::new();
        for t in ticket_entities.iter() {
            if t.state != TicketState::Pending {
                paid_tickets.insert(t.id);
            }
        }

        if !paid_tickets.is_empty() {
            tracing::warn!("Attempted to make payment for already-paid tickets: {paid_tickets:?}");
            return Err(MakePaymentError::TicketAlreadyPaid(paid_tickets));
        }

        let mut payment_attempt = self
            .payment_repo
            .create_payment_attempt(
                method,
                description,
                PaymentAttemptState::Ongoing,
                ticket_entities,
            )
            .await
            .map_err(|_| MakePaymentError::InternalError)?;

        // Simulate bullshit service
        let sleep_time_millis = rand::random::<u64>() % 2500 + 500;
        tracing::info!("Simulating payment service delay of {sleep_time_millis}ms");
        tokio::time::sleep(std::time::Duration::from_millis(sleep_time_millis)).await;

        // Let's say 50% of payment attempts fail for no fucken reason
        let randy = rand::random_range(0.0..=1.0);
        let is_failure = randy < self.failure_chance;
        tracing::info!("Should the payment go through? randy={randy} is_failure={is_failure}");
        if is_failure {
            self.payment_repo
                .mark_as_failed(&mut payment_attempt)
                .await
                .map_err(|_| MakePaymentError::InternalError)?;

            return Err(MakePaymentError::PaymentProcessingError(payment_attempt));
        }

        self.payment_repo
            .mark_as_success(&mut payment_attempt)
            .await
            .map_err(|_| MakePaymentError::InternalError)?;

        tracing::info!(
            "Successfully processed payments for tickets: {tickets:?} with method {method:?}"
        );
        Ok(payment_attempt)
    }
}
