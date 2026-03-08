# CLAUDE.md — Frisbee Training Platform

This file provides context for AI assistants (Claude and others) working in this codebase.
Read this before making changes.

---

## Project Overview

A microservices-based ultimate frisbee training platform where athletes log sessions,
track throwing volume, and view performance analytics.

**Stack:** Go · Gin · PostgreSQL · Redis · RabbitMQ · Next.js · Kubernetes

---

## Repository Structure

```
frisbee-platform/
├── services/
│   ├── auth-service/          # JWT auth (register, login)
│   ├── training-service/      # Core service: sessions, outbox, event publishing
│   ├── analytics-service/     # Consumes events, computes stats
│   └── notification-worker/   # Consumes events, sends reminders
├── gateway/                   # Nginx API gateway config
├── frontend/                  # Next.js app
├── infra/
│   ├── k8s/                   # Kubernetes manifests
│   ├── docker/                # Dockerfiles per service
│   └── observability/         # Prometheus, Grafana, Loki, Jaeger configs
├── proto/                     # Shared event schemas (if using protobuf)
└── CLAUDE.md
```

Each service is a **self-contained Go module** with its own `go.mod`.

---

## Services

### Auth Service (`/services/auth-service`)
- `POST /register` — create user account
- `POST /login` — returns signed JWT
- Database: `users` table
- No event publishing

### Training Service (`/services/training-service`)
- `POST /training` — log a session
- `GET /training` — list sessions for authenticated user
- `GET /training/:id` — get single session
- `PUT /training/:id` — update session
- `DELETE /training/:id` — soft delete (`deleted_at`)
- Implements the **outbox pattern** — never publish to RabbitMQ directly
- Background worker polls outbox and publishes events

### Analytics Service (`/services/analytics-service`)
- RabbitMQ consumer only (no direct HTTP writes from training service)
- `GET /analytics/weekly` — weekly throw volume
- `GET /analytics/stats` — totals, streaks, throw type breakdown
- Must be **idempotent** — duplicate events must not corrupt aggregates

### Notification Worker (`/services/notification-worker`)
- RabbitMQ consumer, no HTTP endpoints
- Sends reminders when user hasn't trained in 3+ days
- Sends weekly summaries

---

## Data Model

### training-service — Current Schema (`001_initial.sql`)

```sql
-- Core session record
training_sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL,
    session_type     VARCHAR(50) CHECK (IN 'team_training','throwing','gym','conditioning','scrimmage','other'),
    duration_minutes INT  CHECK (> 0),
    intensity        VARCHAR(10) CHECK (IN 'low','medium','high'),
    location         VARCHAR(255),
    weather          VARCHAR(100),
    notes            TEXT,
    session_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    updated_at       TIMESTAMPTZ DEFAULT NOW()
)

-- Throwing stats (one row per session, UNIQUE constraint enforced)
throwing_stats (
    id            UUID PRIMARY KEY,
    session_id    UUID REFERENCES training_sessions(id) ON DELETE CASCADE,
    backhand_reps INT DEFAULT 0,
    forehand_reps INT DEFAULT 0,
    hammer_reps   INT DEFAULT 0,
    scoober_reps  INT DEFAULT 0,
    break_throws  INT DEFAULT 0,   -- break-mark throws
    hucks         INT DEFAULT 0,   -- deep/long throws
    turnovers     INT DEFAULT 0,
    UNIQUE (session_id)
)

-- Conditioning stats (one row per session, UNIQUE constraint enforced)
conditioning_stats (
    id              UUID PRIMARY KEY,
    session_id      UUID REFERENCES training_sessions(id) ON DELETE CASCADE,
    sprints         INT           DEFAULT 0,
    distance_km     NUMERIC(6,2),
    max_speed_kmh   NUMERIC(5,2),
    heart_rate_avg  INT,
    heart_rate_max  INT,
    UNIQUE (session_id)
)

-- Transactional outbox for reliable event publishing
outbox_events (
    id           UUID PRIMARY KEY,
    event_type   VARCHAR(100) NOT NULL,
    payload      JSONB        NOT NULL,
    status       VARCHAR(20)  DEFAULT 'pending',  -- pending | published | failed
    created_at   TIMESTAMPTZ  DEFAULT NOW(),
    published_at TIMESTAMPTZ
)
```

### Indexes
```sql
idx_training_sessions_user_id       ON training_sessions (user_id)
idx_training_sessions_session_date  ON training_sessions (user_id, session_date DESC)
idx_throwing_stats_session_id       ON throwing_stats (session_id)
idx_conditioning_stats_session_id   ON conditioning_stats (session_id)
```

### Hard Deletes (no soft delete in V1)
The current schema uses `ON DELETE CASCADE` — deleting a `training_session` removes its
`throwing_stats` and `conditioning_stats` rows automatically. The training service publishes
a `training.session.deleted` event before deletion so analytics can reverse aggregates.

> **Note:** `deleted_at` soft-delete is planned but not yet in the schema. Do not add it
> without a migration file.

### Planned Migrations (not yet implemented)
```
002_add_offensive_stats.sql   -- cuts, assists, drops
003_add_defensive_stats.sql   -- blocks, layout_blocks, forced_turnovers
004_add_gym_exercises.sql     -- exercise, sets, reps, weight
005_add_pull_stats.sql        -- pulls, pull_distance_avg
```

---

## Event Architecture

### Outbox Pattern (training-service)
```
1. Begin transaction
2. INSERT INTO training_sessions
3. INSERT INTO outbox_events (status = 'pending')
4. Commit transaction
5. Background poller: SELECT unpublished → publish to RabbitMQ → mark published
```

**Never publish to RabbitMQ inside the HTTP request handler.**

### Event: `training.session.created`
```json
{
  "event_id": "uuid",
  "event_type": "training.session.created",
  "user_id": "uuid",
  "session_id": "uuid",
  "occurred_at": "RFC3339",
  "payload": {
    "date": "YYYY-MM-DD",
    "session_type": "drilling|scrimmage|conditioning|team_practice|solo",
    "duration_minutes": 90,
    "intensity": "low|medium|high",
    "throws": {
      "backhand": 80,
      "forehand": 60,
      "hammer": 20,
      "total": 160,
      "completion_rate": 0.94
    },
    "conditioning": {
      "rpe": 8,
      "distance_km": 4.2,
      "sprint_count": 30
    },
    "role": "handler|cutter|hybrid"
  }
}
```

### Event: `training.session.deleted`
Carries `session_id` and `user_id` only. Analytics service must reverse aggregates.

---

## API Gateway (Nginx)

Routes:
- `/api/auth/*` → auth-service:8081
- `/api/training/*` → training-service:8082
- `/api/analytics/*` → analytics-service:8083

Rate limits (Redis-backed):
- `POST /api/auth/login` → 5 req/min
- `POST /api/training` → 30 req/min

All routes except `/api/auth/register` and `/api/auth/login` require:
```
Authorization: Bearer <jwt>
```

---

## Frontend (`/frontend`)

**Framework:** Next.js with TanStack Query

Pages:
| Route | Description |
|---|---|
| `/login` | Auth form |
| `/dashboard` | Weekly volume, sessions this month, throw counts |
| `/training/new` | Log a new session form |
| `/training/history` | Paginated session list |
| `/analytics` | Charts: throw trends, conditioning load, streaks |

Data fetching uses TanStack Query — all server state goes through query hooks,
not component-level `useEffect` fetches.

---

## Observability

### Metrics (Prometheus)
Instrument every service with:
- `http_request_duration_seconds` (histogram, labeled by route + status)
- `http_requests_total` (counter)
- `outbox_events_published_total`
- `rabbitmq_messages_consumed_total`
- `redis_rate_limit_hits_total`

### Logging (zap + Loki)
Use structured logging everywhere. No `fmt.Println`.

```go
logger.Info("session created",
  zap.String("user_id", userID),
  zap.String("session_id", sessionID),
  zap.Int("throw_count", throws.Total),
)
```

Log levels: `DEBUG` locally, `INFO` in production.

### Tracing (OpenTelemetry + Jaeger)
Propagate trace context across service boundaries and RabbitMQ messages.
Every HTTP handler and RabbitMQ consumer should start/continue a span.

Trace the full flow: `Frontend → Training Service → RabbitMQ → Analytics Service`

---

## Kubernetes

Deployments: `auth-service`, `training-service`, `analytics-service`, `notification-worker`
StatefulSets: `postgres`, `redis`, `rabbitmq`

Each deployment requires:
- `readinessProbe` on `/health`
- `livenessProbe` on `/health`
- Resource requests + limits set
- Secrets via `Secret` objects (never hardcoded)
- Config via `ConfigMap`

---

## Local Development

```bash
# Start all infrastructure
docker compose up -d

# Run a service locally
cd services/training-service
go run ./cmd/server

# Run all services
docker compose --profile services up
```

Environment variables are loaded from `.env` files per service.
See `.env.example` in each service directory.

---

## Code Conventions

### Go
- One package per responsibility (`handler`, `service`, `repository`, `worker`)
- Errors wrapped with context: `fmt.Errorf("createSession: %w", err)`
- No global state — inject dependencies via constructor
- Repository interface defined in the service layer, not the repository layer
- Table-driven tests for handlers and service logic

### Database
- Migrations managed with `golang-migrate`
- Migration files in `services/<name>/migrations/`
- Never modify existing migration files — add new ones

### Git
- Branch naming: `feat/`, `fix/`, `infra/`, `docs/`
- Commit format: `feat(training): add outbox poller`
- PRs require passing tests and lint (`golangci-lint`)

---

## Common Pitfalls

- **Don't publish events inside a transaction** — use the outbox poller
- **Don't query analytics DB from training service** — services own their data
- **Don't use `time.Now()` directly in tests** — inject a clock interface
- **Don't forget idempotency keys** — analytics consumers must handle redelivery
- **Soft deletes only** — `deleted_at` not `DELETE FROM`
- **Fat events** — include enough payload so consumers don't need to call back