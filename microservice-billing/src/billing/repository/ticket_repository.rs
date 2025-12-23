use bigdecimal::BigDecimal;
use sqlx::{FromRow, Pool, Postgres, QueryBuilder};
use uuid::Uuid;

use crate::billing::entity::{
    PageCountEntity, PageRequest, PageResult, TicketEntity, TicketState, TicketWithFolderEntity,
};

#[derive(Clone)]
pub struct TicketRepository {
    executor: Pool<Postgres>,
}

impl TicketRepository {
    pub fn new(executor: Pool<Postgres>) -> Self {
        Self { executor }
    }

    pub async fn create_ticket(
        &self,
        folder_id: Uuid,
        total: BigDecimal,
        item_id: String,
        description: Option<String>,
    ) -> Result<TicketEntity, sqlx::Error> {
        let entity = TicketEntity::new(folder_id, total, item_id, description);

        sqlx::query!(
            "
            INSERT INTO ticket (id, folder_id, total, item_id, description, created_at, closed_at, state)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            ",
            entity.id,
            entity.folder_id,
            entity.total,
            entity.item_id,
            entity.description,
            entity.created_at,
            entity.closed_at,
            entity.state as _,
        )
            .fetch_optional(&self.executor)
            .await
            .inspect_err(|e| tracing::error!("SQL error while creating ticket: {e}"))?;

        Ok(entity)
    }

    pub async fn get_ticket_by_id(&self, id: &Uuid) -> Result<Option<TicketEntity>, sqlx::Error> {
        let maybe_entity = sqlx::query_as!(
            TicketEntity,
            "
            SELECT
                id, folder_id, total, item_id, description, created_at, closed_at, state AS \"state: TicketState\"
            FROM ticket
            WHERE id = $1
            ",
            id
        )
            .fetch_optional(&self.executor)
            .await
            .inspect_err(|e| tracing::error!("SQL error while getting ticket by ID: {e}"))?;

        Ok(maybe_entity)
    }

    pub async fn get_ticket_with_folder_by_id(
        &self,
        id: &Uuid,
    ) -> Result<Option<TicketWithFolderEntity>, sqlx::Error> {
        let maybe_entity = sqlx::query_as!(
            TicketWithFolderEntity,
            "
            SELECT
                t.id AS ticket_id, t.folder_id AS ticket_folder_id, t.total AS ticket_total, t.item_id AS ticket_item_id, t.description AS ticket_description, t.created_at AS ticket_created_at, t.closed_at AS ticket_closed_at, t.state AS \"ticket_state: TicketState\",
                f.id AS folder_id, f.created_at AS folder_created_at, f.closed_at AS folder_closed_at, f.name AS folder_name
            FROM ticket t JOIN folder f ON t.folder_id = f.id
            WHERE t.id = $1
            ",
            id
        )
            .fetch_optional(&self.executor)
            .await
            .inspect_err(|e| tracing::error!("SQL error while getting ticket with folder by ID: {e}"))?;

        Ok(maybe_entity)
    }

    pub async fn get_all_tickets_by_ids(
        &self,
        ids: &Vec<Uuid>,
    ) -> Result<Vec<TicketEntity>, sqlx::Error> {
        let tickets_vec = sqlx::query_as!(
            TicketEntity,
            "
            SELECT
                id, folder_id, total, item_id, description, created_at, closed_at, state AS \"state: TicketState\"
            FROM ticket
            WHERE id =ANY($1)
            ",
            ids
        )
            .fetch_all(&self.executor)
            .await
            .inspect_err(|e| tracing::error!("SQL error while getting tickets by ID: {e}"))?;

        Ok(tickets_vec)
    }

    pub async fn get_all_tickets_with_folder_by_ids(
        &self,
        ids: &Vec<Uuid>,
    ) -> Result<Vec<TicketWithFolderEntity>, sqlx::Error> {
        let tickets_vec = sqlx::query_as!(
            TicketWithFolderEntity,
            "
            SELECT
                t.id AS ticket_id, t.folder_id AS ticket_folder_id, t.total AS ticket_total, t.item_id AS ticket_item_id, t.description AS ticket_description, t.created_at AS ticket_created_at, t.closed_at AS ticket_closed_at, t.state AS \"ticket_state: TicketState\",
                f.id AS folder_id, f.created_at AS folder_created_at, f.closed_at AS folder_closed_at, f.name AS folder_name
            FROM ticket t JOIN folder f ON t.folder_id = f.id
            WHERE t.id =ANY($1)
            ",
            ids
        )
            .fetch_all(&self.executor)
            .await
            .inspect_err(|e| tracing::error!("SQL error while getting ticket with folder by ID: {e}"))?;

        Ok(tickets_vec)
    }
}

fn add_search_tickets_sql_where<'a>(
    qb: &mut QueryBuilder<'a, Postgres>,
    folder_id: Option<&'a Uuid>,
) {
    let mut where_added = false;
    let mut prepare_where = || {
        if where_added {
            qb.push(" AND ");
        } else {
            where_added = true;
            qb.push(" WHERE ");
        }
    };

    if let Some(fid) = folder_id {
        prepare_where();
        qb.push("t.folder_id = ").push_bind(fid);
    }
}

impl TicketRepository {
    async fn search_tickets_elements(
        &self,
        folder_id: Option<&Uuid>,
        paging: PageRequest,
    ) -> Result<Vec<TicketWithFolderEntity>, sqlx::Error> {
        let mut qb = sqlx::QueryBuilder::new("
            SELECT
                t.id AS ticket_id, t.folder_id AS ticket_folder_id, t.total AS ticket_total, t.item_id AS ticket_item_id, t.description AS ticket_description, t.created_at AS ticket_created_at, t.closed_at AS ticket_closed_at, t.state AS ticket_state,
                f.id AS folder_id, f.created_at AS folder_created_at, f.closed_at AS folder_closed_at, f.name AS folder_name
            FROM ticket t JOIN folder f ON t.folder_id = f.id
            ");

        add_search_tickets_sql_where(&mut qb, folder_id);

        qb.push(" ORDER BY t.id DESC");
        qb.push(" OFFSET ").push_bind(paging.offset() as i64);
        qb.push(" LIMIT ").push_bind(paging.page_size_u32() as i32);

        let rows = qb
            .build()
            .fetch_all(&self.executor)
            .await
            .inspect_err(|err| {
                tracing::error!("Could not fetch elements on ticket search: {err}")
            })?;

        let mut result = Vec::new();
        for ele in rows {
            let entity = TicketWithFolderEntity::from_row(&ele).inspect_err(|err| {
                tracing::error!(
                    "Could not convert row to TicketWithFolderEntity on ticket search: {err}"
                )
            })?;
            result.push(entity);
        }

        Ok(result)
    }

    async fn search_tickets_count(&self, folder_id: Option<&Uuid>) -> Result<u64, sqlx::Error> {
        let mut qb = sqlx::QueryBuilder::new(
            "SELECT COUNT(*) FROM ticket t JOIN folder f ON t.folder_id = f.id",
        );

        add_search_tickets_sql_where(&mut qb, folder_id);

        let row = qb
            .build()
            .fetch_one(&self.executor)
            .await
            .inspect_err(|err| tracing::error!("Could not fetch count on ticket search: {err}"))?;

        let count = PageCountEntity::from_row(&row)
            .inspect_err(|err| {
                tracing::error!("Could not convert row to PageCountEntity on ticket search: {err}")
            })?
            .count as u64;

        Ok(count)
    }

    pub async fn search_tickets(
        &self,
        folder_id: Option<&Uuid>,
        paging: PageRequest,
    ) -> Result<PageResult<TicketWithFolderEntity>, sqlx::Error> {
        let (elements, total_elements) = tokio::try_join!(
            self.search_tickets_elements(folder_id, paging),
            self.search_tickets_count(folder_id)
        )?;

        Ok(PageResult::new(paging, elements, total_elements))
    }
}

#[derive(Debug)]
pub enum CancelTicketError {
    NotFound,
    AlreadyCanceled,
    AlreadyPaid,
    InternalError,
}

impl TicketRepository {
    pub async fn cancel_ticket(&self, id: &Uuid) -> Result<(), CancelTicketError> {
        tracing::info!("Attempting to cancel ticket ID {id}");

        let ticket_entity = self
            .get_ticket_by_id(id)
            .await
            .map_err(|e| {
                tracing::error!("SQL error while getting ticket ID: {e}");
                CancelTicketError::InternalError
            })?
            .ok_or_else(|| {
                tracing::warn!("Attempted to cancel nonexistent ticket ID {id}");
                CancelTicketError::NotFound
            })?;

        if ticket_entity.state == TicketState::Paid {
            tracing::warn!("Attempted to cancel already-paid ticket ID {id}");
            return Err(CancelTicketError::AlreadyPaid);
        }

        if ticket_entity.state == TicketState::Canceled {
            tracing::warn!("Attempted to cancel already-canceled ticket ID {id}");
            return Err(CancelTicketError::AlreadyCanceled);
        }

        sqlx::query!("UPDATE ticket SET state = 'canceled' WHERE id = $1", id)
            .fetch_optional(&self.executor)
            .await
            .map_err(|e| {
                tracing::error!("SQL error while canceling ticket: {e}");
                CancelTicketError::InternalError
            })?;

        Ok(())
    }
}
