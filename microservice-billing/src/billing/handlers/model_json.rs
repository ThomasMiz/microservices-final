use bigdecimal::BigDecimal;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::fmt::Display;
use uuid::Uuid;

use crate::billing::entity::{FolderEntity, TicketEntity, TicketState, TicketWithFolderEntity};

#[derive(Serialize, Deserialize)]
pub struct MessageJson {
    message: String,
}

impl MessageJson {
    pub fn new(message: String) -> Self {
        Self { message }
    }
}

impl<T: Display> From<T> for MessageJson {
    fn from(value: T) -> Self {
        return Self::new(value.to_string());
    }
}

#[derive(Serialize, Deserialize)]
pub struct FolderJson {
    pub id: Uuid,
    pub created_at: DateTime<Utc>,
    pub closed_at: Option<DateTime<Utc>>,
    pub name: Option<String>,
}

impl From<FolderEntity> for FolderJson {
    fn from(value: FolderEntity) -> Self {
        Self {
            id: value.id,
            created_at: value.created_at,
            closed_at: value.closed_at,
            name: value.name,
        }
    }
}

#[derive(Serialize, Deserialize)]
pub struct TicketJson {
    pub id: Uuid,
    pub folder: FolderJson,
    pub total: BigDecimal,
    pub item_id: String,
    pub description: Option<String>,
    pub created_at: DateTime<Utc>,
    pub closed_at: Option<DateTime<Utc>>,
    pub state: TicketState,
}

impl From<TicketWithFolderEntity> for TicketJson {
    fn from(value: TicketWithFolderEntity) -> Self {
        return TicketJson::from(value.separate());
    }
}

impl From<(FolderEntity, TicketEntity)> for TicketJson {
    fn from(value: (FolderEntity, TicketEntity)) -> Self {
        Self {
            id: value.1.id,
            folder: FolderJson::from(value.0),
            total: value.1.total,
            item_id: value.1.item_id,
            description: value.1.description,
            created_at: value.1.created_at,
            closed_at: value.1.closed_at,
            state: value.1.state,
        }
    }
}
