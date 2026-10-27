-- 0020_course_notes: per-user free-form notes / cheat sheets for each subject
-- (course). Separate from word_notes (per-word English flashcards). Each item
-- is Markdown. course_note_summaries caches one AI-generated cheat sheet per
-- (user, subject), regenerated on demand.

CREATE TABLE IF NOT EXISTS course_notes (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id BIGINT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    title      TEXT NOT NULL DEFAULT '',
    body       TEXT NOT NULL DEFAULT '',
    sort_order INT  NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_course_notes_user_subject ON course_notes (user_id, subject_id, sort_order, id);

CREATE TABLE IF NOT EXISTS course_note_summaries (
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subject_id   BIGINT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    summary_md   TEXT NOT NULL DEFAULT '',
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, subject_id)
);
