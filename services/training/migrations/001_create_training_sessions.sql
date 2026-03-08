CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS training_sessions (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          UUID NOT NULL,
    title            VARCHAR(255) NOT NULL,
    description      TEXT,
    duration_minutes INT NOT NULL CHECK (duration_minutes > 0),
    throw_count      INT NOT NULL DEFAULT 0 CHECK (throw_count >= 0),
    session_date     TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_training_sessions_user_id
    ON training_sessions (user_id);

CREATE INDEX IF NOT EXISTS idx_training_sessions_session_date
    ON training_sessions (session_date DESC);
