-- 0066_pmp_daily_sets: the PMP "Daily 20" practice set. One row per user per
-- day: 20 AI-written PMP-style questions (questions JSONB) and, once the
-- learner submits, their answers and score (result JSONB). Generated from the
-- learner's own PMP cheatsheet notes; no real exam questions are stored.

CREATE TABLE IF NOT EXISTS pmp_daily_sets (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    day          DATE NOT NULL,
    questions    JSONB NOT NULL,
    result       JSONB NOT NULL DEFAULT '{}',
    score        INT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    submitted_at TIMESTAMPTZ,
    UNIQUE (user_id, day)
);
