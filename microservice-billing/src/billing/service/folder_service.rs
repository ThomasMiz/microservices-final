use std::fmt::Debug;
use uuid::Uuid;

use crate::billing::repository::{CloseFolderError, DeleteFolderError};
use crate::billing::{entity::FolderEntity, repository::FolderRepository};

#[derive(Clone)]
pub struct FolderService {
    folder_repo: FolderRepository,
}

impl FolderService {
    pub fn new(folder_repo: FolderRepository) -> Self {
        Self { folder_repo }
    }
}

impl Debug for FolderService {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "FolderService")
    }
}

#[derive(Debug)]
pub enum CreateFolderError {
    InternalError,
}

impl FolderService {
    #[tracing::instrument]
    pub async fn create_folder(
        &self,
        name: Option<String>,
    ) -> Result<FolderEntity, CreateFolderError> {
        self.folder_repo
            .create_folder(name)
            .await
            .map_err(|_| CreateFolderError::InternalError)
    }
}

#[derive(Debug)]
pub enum GetFolderError {
    NotFound,
    InternalError,
}

impl FolderService {
    #[tracing::instrument]
    pub async fn get_folder_by_id(&self, id: &Uuid) -> Result<FolderEntity, GetFolderError> {
        match self.folder_repo.get_folder_by_id(id).await {
            Ok(Some(folder)) => Ok(folder),
            Ok(None) => Err(GetFolderError::NotFound),
            Err(_) => Err(GetFolderError::InternalError),
        }
    }

    #[tracing::instrument]
    pub async fn close_folder(&self, id: &Uuid) -> Result<FolderEntity, CloseFolderError> {
        self.folder_repo.close_folder(id).await
    }
}

impl FolderService {
    #[tracing::instrument]
    pub async fn delete_folder_by_id(&self, id: &Uuid) -> Result<(), DeleteFolderError> {
        self.folder_repo.delete_folder(id).await
    }
}
