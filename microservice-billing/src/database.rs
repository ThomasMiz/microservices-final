use std::{env, fmt};

use sqlx::{Pool, Postgres, postgres::PgPoolOptions};
use tracing::info;

#[derive(Debug)]
pub enum ConnectDatabaseError {
    EnvUrlNotSet,
    EnvUrlNotUnicode,
    ConnectError(String, sqlx::Error),
}

impl fmt::Display for ConnectDatabaseError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::EnvUrlNotSet => write!(
                f,
                "The DATABASE_URL environment variable must be set with the database's IP:PORT"
            ),
            Self::EnvUrlNotUnicode => write!(
                f,
                "Could not read DATABASE_URL environment variable because it is not unicode. How about you fix your shit before starting this service?"
            ),
            Self::ConnectError(url, error) => {
                write!(f, "Could not connect to the database at \"{url}\": {error}")
            }
        }
    }
}

impl From<env::VarError> for ConnectDatabaseError {
    fn from(value: env::VarError) -> Self {
        match value {
            env::VarError::NotPresent => Self::EnvUrlNotSet,
            env::VarError::NotUnicode(_) => Self::EnvUrlNotUnicode,
        }
    }
}

pub async fn connect_to_db() -> Result<Pool<Postgres>, ConnectDatabaseError> {
    let database_url = std::env::var("DATABASE_URL")?;

    let maybe_pool = PgPoolOptions::new().connect(&database_url).await;

    match maybe_pool {
        Ok(pool) => {
            info!("Connected to the database");
            Ok(pool)
        }
        Err(e) => Err(ConnectDatabaseError::ConnectError(database_url, e)),
    }
}
