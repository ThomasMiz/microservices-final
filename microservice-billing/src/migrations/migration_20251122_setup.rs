use sqlx_migrator::vec_box;

// Folder table operations
const CREATE_FOLDER_TABLE_SQL: &'static str = r#"
-- Table for billind folders. If `closed_at` is not null, then the folder is
-- closed and can no longer be used. The IDs should be UUIDv7.
CREATE TABLE IF NOT EXISTS folder (
    id         UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at  TIMESTAMP WITH TIME ZONE,
    name       VARCHAR(1000)
);
"#;

const DROP_FOLDER_TABLE_SQL: &'static str = r#"
DROP TABLE IF EXISTS folder;
"#;

// Ticket state enum operations
const CREATE_TICKETSTATE_TYPE_SQL: &'static str = r#"
DO $$ BEGIN
    CREATE TYPE TICKETSTATE AS ENUM ('pending', 'paying', 'paid', 'canceled');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
"#;

const DROP_TICKETSTATE_TYPE_SQL: &'static str = r#"
DROP TYPE TICKETSTATE;
"#;

// Ticket table operations
const CREATE_TICKET_TABLE_SQL: &'static str = r#"
-- Table for tickets. Tickets are also identified by an UUIDv7, reference the
-- folder they belong to, an opaque "item ID" sent by the microservice that
-- requested creating the ticket, and a state as per the enum above.
CREATE TABLE IF NOT EXISTS ticket (
    id          UUID PRIMARY KEY,
    folder_id   UUID NOT NULL REFERENCES folder (id),
    total       DECIMAL(24, 2) NOT NULL,
    item_id     VARCHAR(1000) NOT NULL,
    description VARCHAR(1000), -- Opaque value provided by the consumer microservice
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at   TIMESTAMP WITH TIME ZONE,
    state       TICKETSTATE NOT NULL
);
"#;

const DROP_TICKET_TABLE_SQL: &'static str = r#"
DROP TABLE IF EXISTS ticket;
"#;

// Ticket index operations
const CREATE_TICKET_INDEX_SQL: &'static str = r#"
CREATE UNIQUE INDEX IF NOT EXISTS ticket_btree_folder_and_id ON ticket USING BTREE (folder_id, id);
"#;

const DROP_TICKET_INDEX_SQL: &'static str = r#"
DROP INDEX IF EXISTS ticket_btree_folder_and_id;
"#;

// Payment attempt state enum operations
const CREATE_PAYMENTATTEMPTSTATE_TYPE_SQL: &'static str = r#"
DO $$ BEGIN
    CREATE TYPE PAYMENTATTEMPTSTATE AS ENUM ('ongoing', 'successful', 'failed');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
"#;

const DROP_PAYMENTATTEMPTSTATE_TYPE_SQL: &'static str = r#"
DROP TYPE PAYMENTATTEMPTSTATE;
"#;

// Payment method enum operations
const CREATE_PAYMENTMETHOD_TYPE_SQL: &'static str = r#"
DO $$ BEGIN
    CREATE TYPE PAYMENTMETHOD AS ENUM ('cash', 'credit', 'debit', 'dishwashing', 'other');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
"#;

const DROP_PAYMENTMETHOD_TYPE_SQL: &'static str = r#"
DROP TYPE IF EXISTS PAYMENTMETHOD;
"#;

// Payment attempt table operations
const CREATE_PAYMENT_ATTEMPT_TABLE_SQL: &'static str = r#"
-- Represents an attempt to pay one or multiple tickets.
CREATE TABLE IF NOT EXISTS payment_attempt (
    id          UUID PRIMARY KEY,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at   TIMESTAMP WITH TIME ZONE,
    method      PAYMENTMETHOD,
    description VARCHAR(1000), -- Opaque value provided by the consumer microservice
    state       PAYMENTATTEMPTSTATE NOT NULL,
    total       DECIMAL(24, 2) NOT NULL
);
"#;

const DROP_PAYMENT_ATTEMPT_TABLE_SQL: &'static str = r#"
DROP TABLE IF EXISTS payment_attempt;
"#;

// Payment attempt ticket table operations
const CREATE_PAYMENT_ATTEMPT_TICKET_TABLE_SQL: &'static str = r#"
-- Relates tickets to a payment attempt. There must be at least one per payment_attempt.
CREATE TABLE IF NOT EXISTS payment_attempt_ticket (
    payment_attempt_id UUID NOT NULL,
    ticket_id          UUID NOT NULL,
    PRIMARY KEY(payment_attempt_id, ticket_id)
);
"#;

const DROP_PAYMENT_ATTEMPT_TICKET_TABLE_SQL: &'static str = r#"
DROP TABLE IF EXISTS payment_attempt_ticket;
"#;

pub(crate) struct Migration20251122Setup;

sqlx_migrator::postgres_migration!(
    Migration20251122Setup,
    "billing",
    "Migration20251122Setup",
    vec_box![],
    vec_box![
        (CREATE_FOLDER_TABLE_SQL, DROP_FOLDER_TABLE_SQL),
        (CREATE_TICKETSTATE_TYPE_SQL, DROP_TICKETSTATE_TYPE_SQL),
        (CREATE_TICKET_TABLE_SQL, DROP_TICKET_TABLE_SQL),
        (CREATE_TICKET_INDEX_SQL, DROP_TICKET_INDEX_SQL),
        (
            CREATE_PAYMENTATTEMPTSTATE_TYPE_SQL,
            DROP_PAYMENTATTEMPTSTATE_TYPE_SQL
        ),
        (CREATE_PAYMENTMETHOD_TYPE_SQL, DROP_PAYMENTMETHOD_TYPE_SQL),
        (
            CREATE_PAYMENT_ATTEMPT_TABLE_SQL,
            DROP_PAYMENT_ATTEMPT_TABLE_SQL
        ),
        (
            CREATE_PAYMENT_ATTEMPT_TICKET_TABLE_SQL,
            DROP_PAYMENT_ATTEMPT_TICKET_TABLE_SQL
        )
    ]
);
