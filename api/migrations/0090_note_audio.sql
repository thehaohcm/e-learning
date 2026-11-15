-- 0090_note_audio: learner-uploaded audio attached to a cheatsheet section
-- (course_note). A section can have several audio files (e.g. a recorded
-- explanation or a NotebookLM summary of that topic). Stored as BYTEA so it
-- survives deploys. Works for every subject's cheatsheet (PMP, pentest,
-- English, AWS) since it is keyed by the note. Deleted with the note or user.

CREATE TABLE IF NOT EXISTS note_audio (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    note_id    BIGINT NOT NULL REFERENCES course_notes(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    mime       TEXT NOT NULL,
    bytes      INT  NOT NULL,
    data       BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS ix_note_audio_note ON note_audio (note_id, id);
