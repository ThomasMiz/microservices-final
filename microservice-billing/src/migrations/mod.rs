use sqlx::Postgres;
use sqlx_migrator::{Info, Migrate, Migrator, Plan};
use std::ops::DerefMut;

mod migration_20251122_setup;

fn create_migrator() -> Migrator<Postgres> {
    let mut migrator = Migrator::default();
    migrator
        .add_migration(Box::new(migration_20251122_setup::Migration20251122Setup))
        .unwrap();
    migrator
}

pub async fn run_migrations(pool: &sqlx::Pool<Postgres>) {
    let migrator = create_migrator();
    let mut conn = pool.acquire().await.unwrap();
    migrator
        .run(conn.deref_mut(), &Plan::apply_all())
        .await
        .unwrap();
}
