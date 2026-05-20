CREATE TABLE IF NOT EXISTS bookings (
    id            UUID PRIMARY KEY,
    user_id       UUID NOT NULL,
    train_id      UUID NOT NULL,
    seat_count    INT NOT NULL CHECK (seat_count > 0),
    amount_cents  BIGINT NOT NULL CHECK (amount_cents >= 0),
    status        TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bookings_user ON bookings(user_id);
CREATE INDEX IF NOT EXISTS idx_bookings_train ON bookings(train_id);

CREATE TABLE IF NOT EXISTS tickets (
    id         UUID PRIMARY KEY,
    booking_id UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    code       TEXT NOT NULL,
    issued_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tickets_booking ON tickets(booking_id);
