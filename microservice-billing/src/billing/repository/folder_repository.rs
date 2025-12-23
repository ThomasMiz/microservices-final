use chrono::Utc;
use sqlx::{Pool, Postgres};
use std::ops::DerefMut;
use tracing::info;
use uuid::Uuid;

use crate::billing::entity::{FolderEntity, TicketEntity, TicketState};

#[derive(Clone)]
pub struct FolderRepository {
    executor: Pool<Postgres>,
}

impl FolderRepository {
    pub fn new(executor: Pool<Postgres>) -> Self {
        Self { executor }
    }

    pub async fn create_folder(&self, name: Option<String>) -> Result<FolderEntity, sqlx::Error> {
        let entity = FolderEntity::new(name);

        sqlx::query!(
            "INSERT INTO folder (id, created_at, closed_at, name) VALUES ($1, $2, NULL, $3)",
            entity.id,
            entity.created_at,
            entity.name,
        )
        .fetch_optional(&self.executor)
        .await
        .inspect_err(|e| tracing::error!("SQL error while creating folder: {e}"))?;

        info!("Created new folder with ID {}", entity.id);

        Ok(entity)
    }

    pub async fn get_folder_by_id(&self, id: &Uuid) -> Result<Option<FolderEntity>, sqlx::Error> {
        let maybe_entity = sqlx::query_as!(
            FolderEntity,
            "SELECT id, created_at, closed_at, name FROM folder WHERE id = $1",
            id
        )
        .fetch_optional(&self.executor)
        .await
        .inspect_err(|e| tracing::error!("SQL error while getting folder by ID: {e}"))?;

        Ok(maybe_entity)
    }
}

#[derive(Debug)]
pub enum CloseFolderError {
    NotFound,
    TicketsRemain(Vec<TicketEntity>),
    AlreadyClosed(FolderEntity),
    InternalError,
}

impl From<sqlx::Error> for CloseFolderError {
    fn from(_value: sqlx::Error) -> Self {
        Self::InternalError
    }
}

impl FolderRepository {
    pub async fn close_folder(&self, id: &Uuid) -> Result<FolderEntity, CloseFolderError> {
        tracing::debug!("Attempting to close folder with ID {id}");

        let mut folder = match self.get_folder_by_id(id).await? {
            Some(f) => f,
            None => return Err(CloseFolderError::NotFound),
        };

        if folder.closed_at.is_some() {
            tracing::error!("Cannot close folder ID {id}: already closed");
            return Err(CloseFolderError::AlreadyClosed(folder));
        }

        let remaining_tickets = sqlx::query_as!(
            TicketEntity,
            "
            SELECT
                id, folder_id, total, item_id, description, created_at, closed_at, state AS \"state: TicketState\"
            FROM ticket
            WHERE folder_id = $1 AND state IN ('pending', 'paying')
            ",
            id
        )
            .fetch_all(&self.executor)
            .await
            .inspect_err(|e| tracing::error!("SQL error while getting remaining tickets for folder: {e}"))?;

        if !remaining_tickets.is_empty() {
            let remaining_ticket_ids = remaining_tickets
                .iter()
                .map(|t| t.id.to_string())
                .collect::<Vec<_>>()
                .join(", ");

            tracing::warn!(
                "Could not close folder ID {id} because there are remaining tickets: {remaining_ticket_ids}"
            );
            return Err(CloseFolderError::TicketsRemain(remaining_tickets));
        }

        folder.closed_at = Some(Utc::now());

        sqlx::query!(
            "UPDATE folder SET closed_at = $2 WHERE id = $1",
            folder.id,
            folder.closed_at
        )
        .fetch_optional(&self.executor)
        .await
        .inspect_err(|e| tracing::error!("SQL error while closing folder: {e}"))?;

        tracing::info!(
            "Closed folder ID {} with date {}",
            folder.id,
            folder.closed_at.as_ref().unwrap()
        );

        Ok(folder)
    }
}

#[derive(Debug)]
pub enum DeleteFolderError {
    NotFound,
    TicketsRemain(Vec<TicketEntity>),
    AlreadyClosed(FolderEntity),
    InternalError,
}

impl From<sqlx::Error> for DeleteFolderError {
    fn from(_value: sqlx::Error) -> Self {
        Self::InternalError
    }
}

impl FolderRepository {
    pub async fn delete_folder(&self, id: &Uuid) -> Result<(), DeleteFolderError> {
        tracing::debug!("Attempting to delete folder with ID {id}");
        let mut tx = self.executor.begin().await?;

        let folder = match self.get_folder_by_id(id).await? {
            Some(f) => f,
            None => return Err(DeleteFolderError::NotFound),
        };

        if folder.closed_at.is_some() {
            tracing::error!("Cannot delete folder ID {id}: already closed");
            return Err(DeleteFolderError::AlreadyClosed(folder));
        }

        let remaining_tickets = sqlx::query_as!(
            TicketEntity,
            "
            SELECT
                id, folder_id, total, item_id, description, created_at, closed_at, state AS \"state: TicketState\"
            FROM ticket
            WHERE folder_id = $1 AND state != 'canceled'
            ",
            id
        )
            .fetch_all(tx.deref_mut())
            .await
            .inspect_err(|e| tracing::error!("SQL error while getting remaining tickets for folder: {e}"))?;

        if !remaining_tickets.is_empty() {
            let remaining_ticket_ids = remaining_tickets
                .iter()
                .map(|t| t.id.to_string())
                .collect::<Vec<_>>()
                .join(", ");

            tracing::warn!(
                "Could not delete folder ID {id} because there are remaining tickets: {remaining_ticket_ids}"
            );
            return Err(DeleteFolderError::TicketsRemain(remaining_tickets));
        }

        sqlx::query!(
            "DELETE FROM ticket WHERE folder_id = $1 AND state = 'canceled'",
            folder.id
        )
        .fetch_optional(tx.deref_mut())
        .await
        .inspect_err(|e| tracing::error!("SQL error while deleting tickets from folder: {e}"))?;

        sqlx::query!("DELETE FROM folder WHERE id = $1", folder.id)
            .fetch_optional(tx.deref_mut())
            .await
            .inspect_err(|e| tracing::error!("SQL error while deleting folder: {e}"))?;

        tx.commit()
            .await
            .inspect_err(|e| tracing::error!("SQL error while committing transaction: {e}"))?;

        tracing::info!("Deleted folder ID {}", folder.id);

        Ok(())
    }
}
