--liquibase formatted sql

--changeset ThomasMiz:001
CREATE TABLE reservation (
    id                         BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    room_id                    BIGINT NOT NULL,
    guest_id                   TEXT NOT NULL,
    start_date                 TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date                   TIMESTAMP WITH TIME ZONE NOT NULL,
    billing_folder_id          TEXT,
    reservation_billing_ticket TEXT,
    rented_hourly_price        DECIMAL(20, 2) NOT NULL,
    total_price                DECIMAL(20, 2) NOT NULL,
    CONSTRAINT reservation_fkey_room FOREIGN KEY (room_id) REFERENCES room (id),
    CONSTRAINT reservation_check_end_after_start CHECK (end_date > start_date)
);

CREATE INDEX reservation_btree_room_startdate_enddate ON reservation USING BTREE (room_id,  start_date, end_date);
