CREATE TABLE IF NOT EXISTS events
(
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    date        TIMESTAMPTZ NOT NULL,
    total_seats INT NOT NULL,
    booking_ttl INTERVAL NOT NULL DEFAULT '15 minutes'
);