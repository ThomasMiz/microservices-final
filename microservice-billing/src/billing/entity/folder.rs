use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct FolderEntity {
    pub id: Uuid,
    pub created_at: DateTime<Utc>,
    pub closed_at: Option<DateTime<Utc>>,
    pub name: Option<String>,
}

impl FolderEntity {
    pub fn new(name: Option<String>) -> Self {
        Self {
            id: Uuid::now_v7(),
            created_at: Utc::now(),
            closed_at: None,
            name,
        }
    }
}
