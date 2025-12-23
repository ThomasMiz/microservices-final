use axum::{Extension, Json, extract::Path, http::StatusCode};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::billing::handlers::model_json::MessageJson;
use crate::billing::repository::{CloseFolderError, DeleteFolderError};
use crate::billing::{
    handlers::model_json::FolderJson,
    service::{FolderService, GetFolderError},
};
use std::fmt::Write;

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateFolderJson {
    name: Option<String>,
}

#[tracing::instrument]
pub async fn create_folder_handler(
    Extension(folder_service): Extension<FolderService>,
    Json(body): Json<CreateFolderJson>,
) -> Result<Json<FolderJson>, StatusCode> {
    let entity = folder_service
        .create_folder(body.name)
        .await
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)?;

    Ok(Json(FolderJson::from(entity)))
}

impl From<GetFolderError> for StatusCode {
    fn from(value: GetFolderError) -> Self {
        match value {
            GetFolderError::NotFound => Self::NOT_FOUND,
            GetFolderError::InternalError => Self::INTERNAL_SERVER_ERROR,
        }
    }
}

#[tracing::instrument]
pub async fn get_folder_by_id_handler(
    Extension(folder_service): Extension<FolderService>,
    Path(id): Path<Uuid>,
) -> Result<Json<FolderJson>, StatusCode> {
    let folder_entity = folder_service
        .get_folder_by_id(&id)
        .await
        .map_err(|err| StatusCode::from(err))?;

    Ok(Json(FolderJson::from(folder_entity)))
}

#[tracing::instrument]
pub async fn delete_folder_by_id_handler(
    Extension(folder_service): Extension<FolderService>,
    Path(id): Path<Uuid>,
) -> (StatusCode, Json<MessageJson>) {
    let result = folder_service.delete_folder_by_id(&id).await;

    let (message, status) = match result {
        Ok(()) => (String::from("Folder deleted"), StatusCode::OK),
        Err(DeleteFolderError::NotFound) => (
            String::from("Bro what ((not found))"),
            StatusCode::NOT_FOUND,
        ),
        Err(DeleteFolderError::TicketsRemain(tickets)) => {
            let mut msg = String::from("The folder still has the following tickets:\n");
            for t in tickets {
                let _ = write!(msg, "    - {}\n", t.id);
            }
            (msg, StatusCode::BAD_REQUEST)
        }
        Err(DeleteFolderError::AlreadyClosed(_)) => (
            String::from("Folder already closed"),
            StatusCode::BAD_REQUEST,
        ),
        Err(DeleteFolderError::InternalError) => (
            String::from("I fucked up m8"),
            StatusCode::INTERNAL_SERVER_ERROR,
        ),
    };

    (status, Json(MessageJson::new(message)))
}

#[tracing::instrument]
pub async fn close_folder_by_id_handler(
    Extension(folder_service): Extension<FolderService>,
    Path(id): Path<Uuid>,
) -> (StatusCode, Json<MessageJson>) {
    let result = folder_service.close_folder(&id).await;

    let (message, status) = match result {
        Ok(folder) => (format!("Folder {} closed", folder.id), StatusCode::OK),
        Err(CloseFolderError::NotFound) => (
            String::from("Bro what ((not found))"),
            StatusCode::NOT_FOUND,
        ),
        Err(CloseFolderError::TicketsRemain(tickets)) => {
            let mut msg = String::from("The folder has the following remaining tickets:\n");
            for t in tickets {
                let _ = write!(msg, "    - {}\n", t.id);
            }
            (msg, StatusCode::BAD_REQUEST)
        }
        Err(CloseFolderError::AlreadyClosed(_)) => (
            String::from("Folder already closed"),
            StatusCode::BAD_REQUEST,
        ),
        Err(CloseFolderError::InternalError) => (
            String::from("I fucked up m8"),
            StatusCode::INTERNAL_SERVER_ERROR,
        ),
    };

    (status, Json(MessageJson::new(message)))
}
