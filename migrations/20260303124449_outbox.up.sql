CREATE TABLE IF NOT EXISTS outbox
(
    id       BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL,
    email   TEXT NOT NULL
);