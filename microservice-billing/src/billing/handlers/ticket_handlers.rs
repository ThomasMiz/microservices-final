use std::num::NonZeroU32;

use axum::{
    Extension, Json,
    extract::{Path, Query},
    http::{HeaderMap, StatusCode},
};
use bigdecimal::BigDecimal;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use crate::billing::handlers::model_json::MessageJson;
use crate::billing::repository::CancelTicketError;
use crate::billing::{
    entity::PageRequest,
    handlers::model_json::TicketJson,
    service::{CreateTicketError, GetTicketError, SearchTicketsError, TicketService},
};

#[derive(Debug, Serialize, Deserialize)]
pub struct CreateTicketJson {
    folder_id: Uuid,
    total: BigDecimal,
    item_id: String,
    description: Option<String>,
}

impl From<GetTicketError> for StatusCode {
    fn from(value: GetTicketError) -> Self {
        match value {
            GetTicketError::NotFound => StatusCode::NOT_FOUND,
            GetTicketError::InternalError => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }
}

impl From<CreateTicketError> for StatusCode {
    fn from(value: CreateTicketError) -> Self {
        match value {
            CreateTicketError::FolderNotFound => StatusCode::BAD_REQUEST,
            CreateTicketError::FolderClosed => StatusCode::BAD_REQUEST,
            CreateTicketError::InternalError => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }
}

#[tracing::instrument]
pub async fn create_ticket_handler(
    Extension(ticket_service): Extension<TicketService>,
    Json(body): Json<CreateTicketJson>,
) -> Result<Json<TicketJson>, StatusCode> {
    let (folder_entity, ticket_entity) = ticket_service
        .create_ticket(body.folder_id, body.total, body.item_id, body.description)
        .await?;

    Ok(Json(TicketJson::from((folder_entity, ticket_entity))))
}

#[tracing::instrument]
pub async fn get_ticket_by_id_handler(
    Extension(ticket_service): Extension<TicketService>,
    Path(id): Path<Uuid>,
) -> Result<Json<TicketJson>, StatusCode> {
    let ticket_entity = ticket_service
        .get_ticket_with_folder_by_id(&id)
        .await
        .map_err(|err| StatusCode::from(err))?;

    Ok(Json(TicketJson::from(ticket_entity)))
}

#[tracing::instrument]
pub async fn delete_ticket_by_id_handler(
    Extension(ticket_service): Extension<TicketService>,
    Path(id): Path<Uuid>,
) -> (StatusCode, Json<MessageJson>) {
    let result = ticket_service.cancel_ticket(&id).await;

    let (message, status) = match result {
        Ok(()) => (String::from("Ticket deleted"), StatusCode::OK),
        Err(CancelTicketError::NotFound) => (
            String::from("Bro what ((not found))"),
            StatusCode::NOT_FOUND,
        ),
        Err(CancelTicketError::AlreadyCanceled) => (
            String::from("Ticket already canceled"),
            StatusCode::BAD_REQUEST,
        ),
        Err(CancelTicketError::AlreadyPaid) => (
            String::from("Cannot cancel a paid ticket"),
            StatusCode::BAD_REQUEST,
        ),
        Err(CancelTicketError::InternalError) => (
            String::from("I fucked up m8"),
            StatusCode::INTERNAL_SERVER_ERROR,
        ),
    };

    (status, Json(MessageJson::new(message)))
}

#[derive(Debug, Deserialize)]
pub struct SearchTicketsQueryParams {
    pub page: Option<u32>,

    #[serde(rename = "pageSize")]
    pub page_size: Option<u32>,

    #[serde(rename = "folderId")]
    pub folder_id: Option<Uuid>,
}

impl From<SearchTicketsError> for StatusCode {
    fn from(value: SearchTicketsError) -> Self {
        match value {
            SearchTicketsError::InternalError => StatusCode::INTERNAL_SERVER_ERROR,
        }
    }
}

#[tracing::instrument]
pub async fn search_tickets_handler(
    Extension(ticket_service): Extension<TicketService>,
    Query(query): Query<SearchTicketsQueryParams>,
) -> Result<(HeaderMap, Json<Vec<TicketJson>>), StatusCode> {
    let page_number = query.page.unwrap_or(0);
    let page_size_raw = query.page_size.unwrap_or(20);
    let page_size = NonZeroU32::new(page_size_raw).ok_or(StatusCode::BAD_REQUEST)?;

    let page_request = PageRequest::new(page_number, page_size);

    let page = ticket_service
        .search_tickets(query.folder_id.as_ref(), page_request)
        .await?;

    let total_elements = page.total_elements();
    let elems = page
        .into_elements()
        .into_iter()
        .map(|e| TicketJson::from(e))
        .collect();

    let mut headers = HeaderMap::new();
    headers.insert(
        "X-Total-Elements",
        format!("{total_elements}").parse().unwrap(),
    );

    Ok((headers, Json(elems)))
}
