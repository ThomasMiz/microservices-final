--liquibase formatted sql

--changeset ThomasMiz:001
CREATE TABLE cleaning_job (
    id              BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    room_number     TEXT,
    assigned_to_id  BIGINT,
    created_at      TIMESTAMP WITH TIME ZONE,
    started_at      TIMESTAMP WITH TIME ZONE,
    finished_at     TIMESTAMP WITH TIME ZONE,
    CONSTRAINT cleaningjob_fkey_room FOREIGN KEY (room_number) REFERENCES room_occupancy (room_number),
    CONSTRAINT cleaningjob_fkey_assignedto FOREIGN KEY (assigned_to_id) REFERENCES cleaning_staff (id)
);

CREATE INDEX cleaningjob_btree_roomnumber_and_id ON cleaning_job USING BTREE(room_number, id);
CREATE INDEX cleaningjob_btree_assignedto_and_id ON cleaning_job USING BTREE(assigned_to_id, id);
