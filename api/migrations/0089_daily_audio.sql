-- 0089_daily_audio: learner-uploaded audio notes attached to a Daily set (an
-- "exam"). After taking a set, the learner can upload one or more audio files
-- (mp3 / m4a, e.g. a NotebookLM summary) and play or delete them on the web.
-- Stored as BYTEA so they survive deploys (the DB has its own volume) and need
-- no extra file storage. Deleted with the set or the user (ON DELETE CASCADE).

CREATE TABLE IF NOT EXISTS daily_audio (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    set_id     BIGINT NOT NULL REFERENCES pmp_daily_sets(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    mime       TEXT NOT NULL,
    bytes      INT  NOT NULL,
    data       BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_daily_audio_set ON daily_audio (set_id, id);
