-- 0009_notes: personal decks ("My words") and the floating word notes.

-- A deck with an owner is private to that user (NULL = shared deck).
ALTER TABLE decks ADD COLUMN IF NOT EXISTS owner_id BIGINT REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS ix_decks_owner ON decks (owner_id);

-- Words the user noted while studying, to learn later.
CREATE TABLE IF NOT EXISTS word_notes (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text       TEXT NOT NULL,                  -- the word/phrase as noted
    note       TEXT NOT NULL DEFAULT '',       -- the user's own remark
    context    TEXT NOT NULL DEFAULT '',       -- sentence where it was met
    source     TEXT NOT NULL DEFAULT '',       -- where: "Reading: Office notice", "Quiz"...
    card_id    BIGINT REFERENCES cards(id) ON DELETE SET NULL, -- existing card, or the one created from it
    status     TEXT NOT NULL DEFAULT 'new',    -- new | ready (meaning known) | added (in Study)
    info       JSONB NOT NULL DEFAULT '{}'::jsonb, -- meaning, IPA, example... (from the card or AI)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_word_notes_user_text ON word_notes (user_id, lower(text));
CREATE INDEX IF NOT EXISTS ix_word_notes_user_time ON word_notes (user_id, created_at DESC);
