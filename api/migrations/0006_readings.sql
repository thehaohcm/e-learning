-- 0006_readings: AI-generated reading passages and daily conversations built
-- from the user's own vocabulary (learned words + a few upcoming ones),
-- with a TOEIC/IELTS-style comprehension quiz.

CREATE TABLE IF NOT EXISTS readings (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind          TEXT NOT NULL CHECK (kind IN ('reading', 'conversation')),
    level         TEXT NOT NULL CHECK (level IN ('easy', 'medium', 'ielts')),
    title         TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'processing',  -- processing | completed | failed
    error_message TEXT,
    -- the generated passage/dialogue, questions and the vocabulary used
    content       JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- the user's answers and score, once submitted
    answers       JSONB,
    score         INT,
    total         INT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at  TIMESTAMPTZ,
    answered_at   TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS ix_readings_user_time ON readings (user_id, created_at DESC);
