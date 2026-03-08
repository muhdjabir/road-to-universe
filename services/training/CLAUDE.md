### Training Service — `/services/training-service`

Responsibility: log and manage training sessions. The only service that writes
training data. Publishes events via the outbox pattern — never directly to RabbitMQ.

**Endpoints:**
```
POST   /api/training       — log a new session (creates session + optional stats rows)
GET    /api/training       — list sessions for the authenticated user (paginated)
GET    /api/training/:id   — get a single session with throwing + conditioning stats
PUT    /api/training/:id   — update session metadata or stats
DELETE /api/training/:id   — hard delete (cascades to stats rows, publishes deleted event)
```

**Database: `training_sessions`, `throwing_stats`, `conditioning_stats`, `outbox_events`**
```sql
-- Core session record
training_sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL,        -- Auth0 sub value
    session_type     VARCHAR(50) NOT NULL
                         CHECK (session_type IN
                           ('team_training','throwing','gym','conditioning','scrimmage','other')),
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    intensity        VARCHAR(10) NOT NULL CHECK (intensity IN ('low','medium','high')),
    location         VARCHAR(255),
    weather          VARCHAR(100),
    notes            TEXT,
    session_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
)

-- Throwing stats — one row per session (UNIQUE enforced)
throwing_stats (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id    UUID NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    backhand_reps INT NOT NULL DEFAULT 0,
    forehand_reps INT NOT NULL DEFAULT 0,  -- flick
    hammer_reps   INT NOT NULL DEFAULT 0,
    scoober_reps  INT NOT NULL DEFAULT 0,
    break_throws  INT NOT NULL DEFAULT 0,  -- break-mark throws
    hucks         INT NOT NULL DEFAULT 0,  -- deep throws 25+ yards
    turnovers     INT NOT NULL DEFAULT 0,
    UNIQUE (session_id)
)

-- Conditioning stats — one row per session (UNIQUE enforced)
conditioning_stats (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    sprints         INT NOT NULL DEFAULT 0,
    distance_km     NUMERIC(6,2) CHECK (distance_km >= 0),
    max_speed_kmh   NUMERIC(5,2) CHECK (max_speed_kmh >= 0),
    heart_rate_avg  INT CHECK (heart_rate_avg > 0),
    heart_rate_max  INT CHECK (heart_rate_max > 0),
    UNIQUE (session_id)
)

-- Transactional outbox for reliable event publishing
outbox_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type   VARCHAR(100) NOT NULL,
    payload      JSONB NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending | published | failed
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
)
```

**Indexes:**
```sql
idx_training_sessions_user_id      ON training_sessions (user_id)
idx_training_sessions_session_date ON training_sessions (user_id, session_date DESC)
idx_throwing_stats_session_id      ON throwing_stats (session_id)
idx_conditioning_stats_session_id  ON conditioning_stats (session_id)
```

**Planned migrations (do not implement until the migration file exists):**
```
002_add_offensive_stats.sql   — cuts, assists, drops
003_add_defensive_stats.sql   — blocks, layout_blocks, forced_turnovers
004_add_gym_exercises.sql     — exercise, sets, reps, weight
005_add_pull_stats.sql        — pulls, pull_distance_avg
```