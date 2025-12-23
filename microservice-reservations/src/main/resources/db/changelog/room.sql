--liquibase formatted sql

--changeset ThomasMiz:001
CREATE TABLE room (
    id           BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    number       TEXT NOT NULL,
    active       BOOLEAN NOT NULL,
    name         TEXT,
    description  TEXT,
    max_capacity INT NOT NULL,
    category     TEXT,
    hourly_price DECIMAL(20, 2)
);

CREATE UNIQUE INDEX room_unique_number_where_active ON room (number) WHERE active = true;
