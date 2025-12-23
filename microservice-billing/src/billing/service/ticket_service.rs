use bigdecimal::BigDecimal;
use std::fmt::Debug;
use uuid::Uuid;

use crate::billing::repository::CancelTicketError;
use crate::billing::{
    entity::{FolderEntity, PageRequest, PageResult, TicketEntity, TicketWithFolderEntity},
    repository::{FolderRepository, TicketRepository},
};

#[derive(Clone)]
pub struct TicketService {
    folder_repo: FolderRepository,
    ticket_repo: TicketRepository,
}

impl Debug for TicketService {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "TicketService")
    }
}

impl TicketService {
    pub fn new(folder_repo: FolderRepository, ticket_repo: TicketRepository) -> Self {
        Self {
            folder_repo,
            ticket_repo,
        }
    }
}

#[derive(Debug)]
pub enum CreateTicketError {
    FolderNotFound,
    FolderClosed,
    InternalError,
}

impl TicketService {
    #[tracing::instrument]
    pub async fn create_ticket(
        &self,
        folder_id: Uuid,
        total: BigDecimal,
        item_id: String,
        description: Option<String>,
    ) -> Result<(FolderEntity, TicketEntity), CreateTicketError> {
        let folder_entity = self
            .folder_repo
            .get_folder_by_id(&folder_id)
            .await
            .map_err(|_| CreateTicketError::InternalError)?
            .ok_or_else(|| {
                tracing::warn!("Attempted to create ticket (item_id={item_id}, total={total}) for nonexistent folder ID {folder_id}");
                CreateTicketError::FolderNotFound
            })?;

        if folder_entity.closed_at.is_some() {
            tracing::warn!("Attempted to create ticket for closed folder ID {folder_id}");
            return Err(CreateTicketError::FolderClosed);
        }

        let ticket_entity = self
            .ticket_repo
            .create_ticket(folder_id, total, item_id, description)
            .await
            .map_err(|_| CreateTicketError::InternalError)?;

        Ok((folder_entity, ticket_entity))
    }
}

#[derive(Debug)]
pub enum GetTicketError {
    NotFound,
    InternalError,
}

impl TicketService {
    #[tracing::instrument]
    pub async fn get_ticket_by_id(&self, id: &Uuid) -> Result<TicketEntity, GetTicketError> {
        self.ticket_repo
            .get_ticket_by_id(id)
            .await
            .map_err(|_| GetTicketError::InternalError)?
            .ok_or(GetTicketError::NotFound)
    }

    #[tracing::instrument]
    pub async fn get_ticket_with_folder_by_id(
        &self,
        id: &Uuid,
    ) -> Result<TicketWithFolderEntity, GetTicketError> {
        self.ticket_repo
            .get_ticket_with_folder_by_id(id)
            .await
            .map_err(|_| GetTicketError::InternalError)?
            .ok_or(GetTicketError::NotFound)
    }
}

#[derive(Debug)]
pub enum SearchTicketsError {
    InternalError,
}

impl TicketService {
    #[tracing::instrument]
    pub async fn search_tickets(
        &self,
        folder_id: Option<&Uuid>,
        paging: PageRequest,
    ) -> Result<PageResult<TicketWithFolderEntity>, SearchTicketsError> {
        self.ticket_repo
            .search_tickets(folder_id, paging)
            .await
            .map_err(|_| SearchTicketsError::InternalError)
    }

    #[tracing::instrument]
    pub async fn cancel_ticket(&self, id: &Uuid) -> Result<(), CancelTicketError> {
        return self.ticket_repo.cancel_ticket(id).await;
    }
}
