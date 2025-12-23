--liquibase formatted sql

--changeset ThomasMiz:001
CREATE TABLE cleaning_staff (
    id         BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name       TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT cleaningstaff_unique_name UNIQUE (name)
);

--add initial cleaning staff
INSERT INTO cleaning_staff (name, created_at) VALUES ('Marcelo', now()), ('Mariana', now());
