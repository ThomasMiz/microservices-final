--liquibase formatted sql

--changeset ThomasMiz:001
CREATE TABLE room_occupancy (
    room_number      TEXT PRIMARY KEY,
    state            TEXT,
    current_guest_id TEXT,
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL
);
