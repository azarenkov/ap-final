CREATE TABLE IF NOT EXISTS booking_audit (
    id          BIGSERIAL PRIMARY KEY,
    booking_id  UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    action      TEXT NOT NULL,
    old_status  TEXT NOT NULL,
    new_status  TEXT NOT NULL,
    actor       TEXT NOT NULL DEFAULT 'system',
    at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_booking_audit_booking ON booking_audit(booking_id, at DESC);
