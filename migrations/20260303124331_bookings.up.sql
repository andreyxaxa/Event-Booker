CREATE TABLE IF NOT EXISTS bookings
(
    id         BIGSERIAL PRIMARY KEY,
    event_id   BIGINT REFERENCES events(id),
    email      TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx
    ON bookings(expires_at) WHERE status = 'pending';