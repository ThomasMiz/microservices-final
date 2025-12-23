--liquibase formatted sql

--changeset ThomasMiz:001
CREATE TABLE damage_report (
    id             BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    room_number    TEXT NOT NULL,
    reported_by_id BIGINT NOT NULL,
    item           TEXT,
    description    TEXT,
    guest_id       TEXT NOT NULL,
    fine_amount    DECIMAL(20, 2),
    billing_folder TEXT NOT NULL,
    billing_ticket TEXT NOT NULL,
    created_at     TIMESTAMP WITH TIME ZONE,
    CONSTRAINT damagereport_fkey_reportedby FOREIGN KEY (reported_by_id) REFERENCES cleaning_staff (id)
);
