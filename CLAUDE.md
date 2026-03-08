# CLAUDE.md — Road to Niverse (Frisbee Training Platform)

This file is the source of truth for AI agents working in this codebase.
**Read this entire file before writing any code, creating files, or making changes.**

---

## What This Project Is

A microservices platform for ultimate frisbee athletes to log training sessions,
track throwing volume, and view performance analytics over time.

**Stack:** Go · Gin · PostgreSQL · Redis · RabbitMQ · Next.js · Auth0 · Kubernetes

---

## Repository Layout

```
road-to-niverse/
├── services/
│   ├── auth-service/          # User profile management (Auth0 is the identity provider)
│   ├── training-service/      # Core service: log sessions, publish events via outbox
│   ├── analytics-service/     # Consumes events, computes stats and trends
│   └── notification-worker/   # Consumes events, sends reminders and summaries
├── gateway/                   # Nginx config: JWT validation, routing, rate limiting
├── frontend/                  # Next.js app (TanStack Query for data fetching)
├── infra/
│   ├── k8s/                   # Kubernetes manifests
│   ├── docker/                # Dockerfiles per service
│   └── observability/         # Prometheus, Grafana, Loki, Jaeger configs
└── CLAUDE.md
```

Each service is a **self-contained Go module** with its own `go.mod`. Do not share
code between services via internal packages — duplicate small helpers if needed.

---

## Authentication Architecture

**Auth0 is the identity provider.** This system does not issue its own JWTs.

### Login flow
```
1. User logs in via Auth0 (browser redirect)
2. Auth0 redirects to POST /api/auth/callback with a short-lived code
3. Auth service exchanges code → receives access token from Auth0
4. Auth service sets an HttpOnly cookie on the response:
     Set-Cookie: access_token=<jwt>; HttpOnly; Secure; SameSite=Strict; Path=/
5. Browser stores cookie — frontend JS never reads the token directly
6. Every subsequent request: browser sends cookie automatically
```

### JWT validation flow
```
1. Request arrives at gateway with cookie
2. Gateway extracts JWT from the access_token cookie
3. Gateway validates JWT locally using Auth0's cached JWKS public key
     JWKS endpoint: https://<tenant>.auth0.com/.well-known/jwks.json
4. No network call to auth service — validation is a local crypto operation
5. Gateway extracts user identity and forwards as internal headers:
     X-User-ID: auth0|abc123
     X-User-Email: athlete@example.com
6. Gateway routes request to the target service
7. Target service also re-validates the JWT independently (defence in depth)
```

**The `sub` claim from the Auth0 JWT is the canonical `user_id` used across all
services and in all events.**

### Required env vars for JWT validation (every service)
```
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.road-to-niverse.com
```


## Services

### Auth Service — `/services/auth-service`

Responsibility: bridge between Auth0 and the platform's user profile data.
Auth0 owns credentials. This service owns profile fields.


**No event publishing from this service.**

### Training Service — `/services/training-service`

Responsibility: log and manage training sessions. The only service that writes
training data. Publishes events via the outbox pattern — never directly to RabbitMQ.

### Analytics Service — `/services/analytics-service`

Responsibility: consume training events and compute aggregate stats.
This service has its own database — never query it from other services.

### Notification Worker — `/services/notification-worker`

Responsibility: send timely reminders and summaries. No HTTP endpoints.

## Event Architecture

### Outbox Pattern (training-service only)

The outbox pattern guarantees events are never lost even if RabbitMQ is temporarily down.

```
1.  Begin database transaction
2.  INSERT INTO training_sessions (+ throwing_stats, conditioning_stats as needed)
3.  INSERT INTO outbox_events (status = 'pending', payload = fat event JSON)
4.  Commit transaction
    — both the data and the intent to publish are now durable —
5.  Background poller (runs every ~1s):
      SELECT * FROM outbox_events WHERE status = 'pending' ORDER BY created_at
6.  Publish each event to RabbitMQ
7.  UPDATE outbox_events SET status = 'published', published_at = NOW()
```

**Never publish to RabbitMQ inside an HTTP request handler or inside a transaction.**

### Event: `training.session.created`
```json
{
  "event_id":    "uuid",
  "event_type":  "training.session.created",
  "user_id":     "auth0|abc123",
  "session_id":  "uuid",
  "occurred_at": "2026-03-08T18:00:00Z",
  "payload": {
    "session_date":     "2026-03-08",
    "session_type":     "throwing",
    "duration_minutes": 90,
    "intensity":        "high",
    "throwing": {
      "backhand_reps": 80,
      "forehand_reps": 60,
      "hammer_reps":   20,
      "scoober_reps":  5,
      "break_throws":  15,
      "hucks":         10,
      "turnovers":     3
    },
    "conditioning": {
      "sprints":        20,
      "distance_km":    4.2,
      "max_speed_kmh":  24.5,
      "heart_rate_avg": 155,
      "heart_rate_max": 182
    }
  }
}
```

### Event: `training.session.deleted`
```json
{
  "event_id":    "uuid",
  "event_type":  "training.session.deleted",
  "user_id":     "auth0|abc123",
  "session_id":  "uuid",
  "occurred_at": "2026-03-08T18:05:00Z"
}
```

Analytics must reverse any aggregates built from the matching `training.session.created`
event. Events are fat — payloads must be self-contained so consumers never call back.

---

## API Gateway (Nginx)

The gateway is the **only external entry point**. Services refuse direct external traffic
via Kubernetes NetworkPolicy — only pods labelled `app: api-gateway` can reach them.

**Routing:**
```
/api/auth/*       → auth-service:8081
/api/training/*   → training-service:8082
/api/analytics/*  → analytics-service:8083
```

**What the gateway does on every request:**
```
1. Extract JWT from the access_token cookie
2. Validate JWT against cached Auth0 JWKS (local crypto — no network call)
3. Reject with 401 if token is missing or invalid
   (exceptions: /api/auth/callback and /api/auth/logout pass through unauthenticated)
4. Forward user identity to downstream service:
     X-User-ID: auth0|abc123
     X-User-Email: athlete@example.com
5. Forward internal secret header:
     X-Internal-Secret: <value from Kubernetes Secret>
6. Route to target service
```

**Rate limits (Redis-backed):**
```
POST /api/auth/callback  →  10 req/min
POST /api/training       →  30 req/min
```

**Internal secret:** every downstream service rejects requests missing `X-Internal-Secret`.
Value is injected via a Kubernetes Secret — never hardcoded.

---

## Frontend — `/frontend`

**Framework:** Next.js · TanStack Query

- All server state goes through TanStack Query hooks — never `useEffect` for fetching
- JWT lives in the HttpOnly cookie only — never read or store it in JavaScript
- Credentials are included on every fetch: `fetch(url, { credentials: 'include' })`

**Pages:**
| Route | Description |
|---|---|
| `/login` | Redirects to Auth0 Universal Login |
| `/dashboard` | Weekly volume, sessions this month, throw counts |
| `/training/new` | Form to log a new session |
| `/training/history` | Paginated session list |
| `/analytics` | Charts: throw trends, conditioning load, streaks |

---

## Observability

### Metrics — Prometheus
Every service exposes `/metrics`. Required instruments:
```
http_request_duration_seconds    histogram  labelled by route + status code
http_requests_total              counter
outbox_events_published_total    counter    training-service only
rabbitmq_messages_consumed_total counter    analytics + notification services
redis_rate_limit_hits_total      counter    gateway only
```

### Logging — zap → Loki
Structured logs everywhere. `fmt.Println` is not allowed.
```go
logger.Info("session created",
    zap.String("user_id", userID),
    zap.String("session_id", sessionID),
    zap.Int("duration_minutes", session.DurationMinutes),
)
```
Log level: `DEBUG` locally · `INFO` in production.
Never log the JWT value or cookie contents.

### Tracing — OpenTelemetry → Jaeger
Propagate trace context across all service boundaries including RabbitMQ message headers.
Every HTTP handler and every RabbitMQ consumer must start or continue a span.

Full trace path: `Frontend → Gateway → Training Service → RabbitMQ → Analytics Service`

---

## Kubernetes

| Resource | Kind |
|---|---|
| auth-service | Deployment |
| training-service | Deployment |
| analytics-service | Deployment |
| notification-worker | Deployment |
| postgres | StatefulSet |
| redis | StatefulSet |
| rabbitmq | StatefulSet |

Every Deployment requires:
- `readinessProbe` and `livenessProbe` hitting `GET /health`
- CPU and memory requests + limits defined
- Secrets from `Secret` objects — never literal values in manifests
- Config from `ConfigMap`
- `NetworkPolicy` allowing ingress only from `app: api-gateway`

---

## Local Development

```bash
# Start infrastructure (postgres, redis, rabbitmq)
docker compose up -d

# Run a single service
cd services/training-service && go run ./cmd/server

# Run all services
docker compose --profile services up
```

Each service reads config from a `.env` file.
Copy `.env.example` → `.env` and fill in values before running locally.

---

## Code Conventions

### Go
- Package layout per service: `cmd/` · `handler/` · `service/` · `repository/` · `worker/`
- Wrap errors with context: `fmt.Errorf("createSession: %w", err)`
- No global state — inject all dependencies via constructor functions
- Define repository interfaces in the `service` package, not `repository`
- Use table-driven tests for handlers and service logic
- Never call `time.Now()` directly — inject a `clock` interface for testability

### Database
- Migrations managed with `golang-migrate`
- Migration files live in `services/<name>/migrations/`
- **Never edit an existing migration file** — always add a new numbered file

### Git
- Branch prefixes: `feat/` · `fix/` · `infra/` · `docs/`
- Commit format: `feat(training): add outbox poller`
- PRs must pass tests and `golangci-lint` before merge

---

## Common Pitfalls — Read Before Coding

- **Never publish to RabbitMQ inside a request handler** — always use the outbox poller
- **Never publish inside a database transaction** — the outbox row is the durable record; the poller publishes after commit
- **Delete flow order matters** — write `training.session.deleted` to outbox first, then `DELETE FROM training_sessions`; cascade handles child rows
- **Analytics must be idempotent** — use `event_id` as a deduplication key; the same event arriving twice must not double-count
- **Services own their own data** — never query another service's database directly
- **No soft deletes in V1** — schema uses `ON DELETE CASCADE`; do not add `deleted_at` without a migration file
- **Fat events only** — payloads must be self-contained; consumers must never call back to fetch missing fields
- **Auth0 sub is the user_id everywhere** — in DB rows, event payloads, log fields, and internal headers
- **Never log the JWT or cookie value** — strip from access logs at the gateway level