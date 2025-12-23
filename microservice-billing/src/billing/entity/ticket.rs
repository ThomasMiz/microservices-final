use bigdecimal::BigDecimal;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::billing::entity::FolderEntity;

#[derive(sqlx::Type, Debug, Clone, Copy, PartialEq, Eq)]
#[sqlx(type_name = "TICKETSTATE", rename_all = "lowercase")]
#[derive(Serialize, Deserialize)]
pub enum TicketState {
    Pending,
    Paying,
    Paid,
    Canceled,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct TicketEntity {
    pub id: Uuid,
    pub folder_id: Uuid,
    pub total: BigDecimal,
    pub item_id: String,
    pub description: Option<String>,
    pub created_at: DateTime<Utc>,
    pub closed_at: Option<DateTime<Utc>>,
    pub state: TicketState,
}

impl TicketEntity {
    pub fn new(
        folder_id: Uuid,
        total: BigDecimal,
        item_id: String,
        description: Option<String>,
    ) -> Self {
        Self {
            id: Uuid::now_v7(),
            folder_id,
            total,
            item_id,
            description,
            created_at: Utc::now(),
            closed_at: None,
            state: TicketState::Pending,
        }
    }
}

#[derive(sqlx::FromRow, Debug, Clone, PartialEq, Eq)]
pub struct TicketWithFolderEntity {
    pub folder_id: Uuid,
    pub folder_created_at: DateTime<Utc>,
    pub folder_closed_at: Option<DateTime<Utc>>,
    pub folder_name: Option<String>,
    pub ticket_id: Uuid,
    pub ticket_folder_id: Uuid,
    pub ticket_total: BigDecimal,
    pub ticket_item_id: String,
    pub ticket_description: Option<String>,
    pub ticket_created_at: DateTime<Utc>,
    pub ticket_closed_at: Option<DateTime<Utc>>,
    pub ticket_state: TicketState,
}

impl TicketWithFolderEntity {
    pub fn separate(self) -> (FolderEntity, TicketEntity) {
        let folder = FolderEntity {
            id: self.folder_id,
            created_at: self.folder_created_at,
            closed_at: self.folder_closed_at,
            name: self.folder_name,
        };

        let ticket = TicketEntity {
            id: self.ticket_id,
            folder_id: self.ticket_folder_id,
            total: self.ticket_total,
            item_id: self.ticket_item_id,
            description: self.ticket_description,
            created_at: self.ticket_created_at,
            closed_at: self.ticket_closed_at,
            state: self.ticket_state,
        };

        return (folder, ticket);
    }
}
