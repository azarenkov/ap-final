CREATE TABLE IF NOT EXISTS routes (
    id                 UUID PRIMARY KEY,
    origin             TEXT NOT NULL,
    destination        TEXT NOT NULL,
    distance_km        INT NOT NULL CHECK (distance_km > 0),
    estimated_minutes  INT NOT NULL CHECK (estimated_minutes > 0)
);

CREATE INDEX IF NOT EXISTS idx_routes_origin_destination
    ON routes (origin, destination);

CREATE TABLE IF NOT EXISTS trains (
    id                 UUID PRIMARY KEY,
    code               TEXT NOT NULL UNIQUE,
    name               TEXT NOT NULL,
    route_id           UUID NOT NULL REFERENCES routes(id) ON DELETE RESTRICT,
    departure_time     TIMESTAMPTZ NOT NULL,
    arrival_time       TIMESTAMPTZ NOT NULL,
    total_seats        INT NOT NULL CHECK (total_seats > 0),
    available_seats    INT NOT NULL CHECK (available_seats >= 0),
    price_cents        BIGINT NOT NULL CHECK (price_cents > 0),
    status             TEXT NOT NULL DEFAULT 'SCHEDULED',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (arrival_time > departure_time),
    CHECK (available_seats <= total_seats)
);

CREATE INDEX IF NOT EXISTS idx_trains_route ON trains (route_id);
CREATE INDEX IF NOT EXISTS idx_trains_departure ON trains (departure_time);
