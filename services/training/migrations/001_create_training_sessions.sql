CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── Core session ───────────────────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS training_sessions (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID        NOT NULL,
    session_type     VARCHAR(50) NOT NULL
                         CHECK (session_type IN ('team_training','throwing','gym','conditioning','scrimmage','other')),
    duration_minutes INT         NOT NULL CHECK (duration_minutes > 0),
    intensity        VARCHAR(10) NOT NULL
                         CHECK (intensity IN ('low','medium','high')),
    location         VARCHAR(255),
    weather          VARCHAR(100),
    notes            TEXT,
    session_date     DATE        NOT NULL DEFAULT CURRENT_DATE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Throwing stats (V1 scope) ──────────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS throwing_stats (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id    UUID NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    backhand_reps INT  NOT NULL DEFAULT 0 CHECK (backhand_reps >= 0),
    forehand_reps INT  NOT NULL DEFAULT 0 CHECK (forehand_reps >= 0),
    hammer_reps   INT  NOT NULL DEFAULT 0 CHECK (hammer_reps   >= 0),
    scoober_reps  INT  NOT NULL DEFAULT 0 CHECK (scoober_reps  >= 0),
    break_throws  INT  NOT NULL DEFAULT 0 CHECK (break_throws  >= 0),
    hucks         INT  NOT NULL DEFAULT 0 CHECK (hucks         >= 0),
    turnovers     INT  NOT NULL DEFAULT 0 CHECK (turnovers     >= 0),
    UNIQUE (session_id)  -- one record per session
);

-- ── Conditioning stats (V1 scope) ─────────────────────────────────────────────

CREATE TABLE IF NOT EXISTS conditioning_stats (
    id              UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID           NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    sprints         INT            NOT NULL DEFAULT 0 CHECK (sprints      >= 0),
    distance_km     NUMERIC(6, 2)  CHECK (distance_km    >= 0),
    max_speed_kmh   NUMERIC(5, 2)  CHECK (max_speed_kmh  >= 0),
    heart_rate_avg  INT            CHECK (heart_rate_avg  > 0),
    heart_rate_max  INT            CHECK (heart_rate_max  > 0),
    UNIQUE (session_id)  -- one record per session
);

-- ── Indexes ────────────────────────────────────────────────────────────────────

CREATE INDEX IF NOT EXISTS idx_training_sessions_user_id
    ON training_sessions (user_id);

CREATE INDEX IF NOT EXISTS idx_training_sessions_session_date
    ON training_sessions (user_id, session_date DESC);

CREATE INDEX IF NOT EXISTS idx_throwing_stats_session_id
    ON throwing_stats (session_id);

CREATE INDEX IF NOT EXISTS idx_conditioning_stats_session_id
    ON conditioning_stats (session_id);

-- ── Future migrations (not yet) ────────────────────────────────────────────────
-- 002_add_offensive_stats.sql   (cuts, assists, drops)
-- 003_add_defensive_stats.sql   (blocks, layout_blocks, forced_turnovers)
-- 004_add_gym_exercises.sql     (exercise, sets, reps, weight)
-- 005_add_pull_stats.sql        (pulls, pull_distance_avg)
