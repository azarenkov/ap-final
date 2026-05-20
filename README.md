# Train Tickets System — Final Project

**Course:** Advanced Programming 2
**Author:** Alexey, Rauan, Diana
**Scope:** Train Service + API Gateway (Alexey's responsibility in the team microservices project).

---

## Overview

Microservices-based train ticket platform. The system as a whole consists of four
services and one API Gateway per team member; this repository implements
**Alexey's two pieces**:

| Component | Tech | Port |
|---|---|---|
| Train Service | Go · gRPC · Postgres · Redis · NATS | `50051` |
| API Gateway (Alexey) | Go · Gin · gRPC clients · JWT | `8080` |

The other services owned by teammates (User — Rauan, Booking — Diana, Notification —
common) live in their own repositories. The proto contracts for those services are
checked in here under `proto/{user,booking,notification}/v1/` so the gateway can
talk to them via gRPC.

## Architecture

```
                       ┌────────────────────────┐
   HTTP/JSON           │      API Gateway       │
   client  ───────────▶│  Gin · JWT · CORS · RL │
                       └──────────┬─────────────┘
                                  │ gRPC
        ┌─────────────────────────┼─────────────────────────┐
        ▼                         ▼                         ▼
  ┌───────────┐            ┌─────────────┐           ┌───────────┐
  │ User Svc  │            │ Train Svc   │           │ Booking   │
  │ (Rauan)   │            │   (this)    │           │ (Diana)   │
  └───────────┘            └──────┬──────┘           └───────────┘
                                  │
                  ┌───────────────┼────────────────┐
                  ▼               ▼                ▼
              Postgres          Redis            NATS  ──▶ Notification Svc
            (train_db)        (cache)         (events)
```

Clean Architecture is enforced inside the Train Service:

```
cmd/train-service/main.go       — entrypoint, calls app.MustRun()
internal/
  app/                          — wires config + db + redis + nats + gRPC server
  config/                       — env-driven config
  domain/                       — Train, Route, SearchFilter, errors, invariants
  repository/postgres/          — TrainRepo, RouteRepo (database/sql + pq)
  cache/                        — Redis search-result cache (30s TTL, invalidated on writes)
  usecase/                      — business logic — depends only on interfaces above
  events/                       — NATS publisher for train.* events
  transport/grpc/               — TrainServer implementing trainv1.TrainServiceServer
migrations/001_init.sql         — schema: routes, trains
```

## Train Service — gRPC endpoints (12)

| # | RPC | Description |
|---|---|---|
| 1 | `CreateTrain` | Create a train; requires existing route. |
| 2 | `GetTrainById` | Single train by id. |
| 3 | `UpdateTrain` | Mutate name / times / price / status; emits `train.delayed` / `train.cancelled` on status change. |
| 4 | `DeleteTrain` | Delete train; publishes cancellation event. |
| 5 | `SearchTrains` | Filter by origin / destination / departure window; cached for 30 s. |
| 6 | `GetTrainSchedule` | Synthesised 3-stop schedule from route + times. |
| 7 | `CreateRoute` | Create a route (origin / destination / distance / minutes). |
| 8 | `GetRouteById` | Single route. |
| 9 | `UpdateRoute` | Update mutable route fields. |
| 10 | `DeleteRoute` | Delete (referential-integrity check on trains). |
| 11 | `GetAvailableSeats` | Returns total + available seats. |
| 12 | `UpdateSeatAvailability` | Atomic delta (`-N` reserve, `+N` release); used by Booking Service. |

### NATS events published

- `train.created` — payload `{type, train_id, status, at}`
- `train.updated` — on any update without status change
- `train.delayed` — when status flips to `DELAYED`
- `train.cancelled` — on delete OR status flips to `CANCELLED`

Subjects are dot-separated so the Notification Service can wildcard-subscribe with
`train.>`.

## API Gateway — HTTP routes

Public:

```
POST   /v1/users/register
POST   /v1/users/login
POST   /v1/users/reset-password

GET    /v1/trains
GET    /v1/trains/:id
GET    /v1/trains/:id/schedule
GET    /v1/trains/:id/seats
GET    /v1/routes/:id
```

Authenticated (Bearer JWT, HS256, secret from `JWT_SECRET`):

```
GET    /v1/users/me
PATCH  /v1/users/me
POST   /v1/users/me/change-password

POST   /v1/trains
PATCH  /v1/trains/:id
DELETE /v1/trains/:id
POST   /v1/trains/:id/seats        (body: {"delta": -N|+N})
POST   /v1/routes
PATCH  /v1/routes/:id
DELETE /v1/routes/:id

POST   /v1/bookings
GET    /v1/bookings/:id
GET    /v1/bookings/me
POST   /v1/bookings/:id/cancel
POST   /v1/bookings/:id/confirm
GET    /v1/bookings/:id/status
```

`/healthz` is always public.

Middleware applied to every request:
- **CORS** — allow-all for dev, configurable in `app.go`.
- **Rate limit** — fixed-window 60 req/min per IP via Redis (fails open if Redis is down).
- **JWT** on the `authedAPI` group.

## Quickstart

### Local (Docker Compose)

```bash
cp .env.example .env
make up                       # builds + starts train-db, redis, nats, train-service, api-gateway
curl -s http://localhost:8080/healthz
```

### Local (go run, no Docker)

```bash
# 1. Start infra
docker compose up -d train-db redis nats

# 2. Train service
cd train-service
TRAIN_DATABASE_URL='postgres://postgres:postgres@localhost:5432/train_db?sslmode=disable' \
REDIS_ADDR=localhost:6379 \
NATS_URL=nats://localhost:4222 \
go run ./cmd/train-service

# 3. Gateway (in a second terminal)
cd api-gateway
TRAIN_SERVICE_ADDR=localhost:50051 \
JWT_SECRET=dev-secret \
go run ./cmd/api-gateway
```

### Example HTTP flow

```bash
# Create a route + train (auth — use a JWT signed by your team's user service).
TOKEN=eyJ...your-jwt-here

curl -X POST http://localhost:8080/v1/routes \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"origin":"Astana","destination":"Almaty","distance_km":1200,"estimated_minutes":960}'

curl -X POST http://localhost:8080/v1/trains \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"code\":\"IC-101\",\"name\":\"Talgo\",\"route_id\":\"$ROUTE_ID\",\
       \"departure_time\":\"2026-06-01T08:00:00Z\",\
       \"arrival_time\":\"2026-06-01T20:00:00Z\",\
       \"total_seats\":300,\"price_cents\":1500000}"

# Search trains
curl "http://localhost:8080/v1/trains?origin=Astana&destination=Almaty"

# Reserve 2 seats (booking service would call this)
curl -X POST http://localhost:8080/v1/trains/$TRAIN_ID/seats \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"delta":-2}'
```

## Proto code-generation

The generated Go code lives at `gen/` as its own module
(`github.com/azarenkov/ap2-final-gen`), referenced from `train-service` and
`api-gateway` via a `replace` directive.

```bash
make proto    # regenerates {train,user,booking,notification}/v1/*.pb.go + *_grpc.pb.go
```

## Tests

```bash
make test
```

- **Unit (domain)** — `internal/domain/train_test.go` — invariants for `NewTrain`,
  `ApplySeatDelta`, `IsValidStatus`, `NewRoute`.
- **Unit (usecase)** — `internal/usecase/usecase_test.go` — happy paths, status
  transitions emit the correct NATS event, over-reservation rejection, etc. Uses
  in-memory fake repositories + a spy publisher.
- **Integration (manual)** — start `docker compose up -d train-db redis nats`,
  then `make test` for the full suite. The repository layer is tested implicitly
  by the gRPC integration path via the gateway.

Coverage on `internal/domain` and `internal/usecase` is the main quality bar
(`go test -cover ./...`).

## Repository layout

```
final/
├── proto/                      .proto contracts for all 4 services
├── gen/                        generated Go code (own module)
├── train-service/              clean-architecture service
├── api-gateway/                Alexey's HTTP-to-gRPC gateway
├── docker-compose.yml          local infra + services
├── Makefile                    proto / build / test / up / down / lint
└── README.md
```
