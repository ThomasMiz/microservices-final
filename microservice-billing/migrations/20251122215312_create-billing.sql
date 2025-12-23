-- Table for billind folders. If `closed_at` is not null, then the folder is
-- closed and can no longer be used. The IDs should be UUIDv7.
CREATE TABLE folder (
    id         UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at  TIMESTAMP WITH TIME ZONE,
    name       VARCHAR(1000)
);

CREATE TYPE TICKETSTATE AS ENUM ('pending', 'paying', 'paid', 'canceled');

-- Table for tickets. Tickets are also identified by an UUIDv7, reference the
-- folder they belong to, an opaque "item ID" sent by the microservice that
-- requested creating the ticket, and a state as per the enum above.
CREATE TABLE ticket (
    id          UUID PRIMARY KEY,
    folder_id   UUID NOT NULL REFERENCES folder (id),
    total       DECIMAL(24, 2) NOT NULL,
    item_id     VARCHAR(1000) NOT NULL,
    description VARCHAR(1000), -- Opaque value provided by the consumer microservice
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at   TIMESTAMP WITH TIME ZONE,
    state       TICKETSTATE NOT NULL
);

CREATE UNIQUE INDEX ticket_btree_folder_and_id ON ticket USING BTREE (folder_id, id);

CREATE TYPE PAYMENTATTEMPTSTATE AS ENUM ('ongoing', 'successful', 'failed');
CREATE TYPE PAYMENTMETHOD AS ENUM ('cash', 'credit', 'debit', 'dishwashing', 'other');

-- Represents an attempt to pay one or multiple tickets.
CREATE TABLE payment_attempt (
    id          UUID PRIMARY KEY,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL,
    closed_at   TIMESTAMP WITH TIME ZONE,
    method      PAYMENTMETHOD,
    description VARCHAR(1000), -- Opaque value provided by the consumer microservice
    state       PAYMENTATTEMPTSTATE NOT NULL,
    total       DECIMAL(24, 2) NOT NULL
);

-- Relates tickets to a payment attempt. There must be at least one per payment_attempt.
CREATE TABLE payment_attempt_ticket (
    payment_attempt_id UUID NOT NULL,
    ticket_id          UUID NOT NULL,
    PRIMARY KEY(payment_attempt_id, ticket_id)
);
