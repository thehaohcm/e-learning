-- 0012_pron_track: pronunciation is reviewed on its own schedule.
-- A word whose meaning you know can still be said wrong (stress on the wrong
-- syllable). user_pron is a second FSRS state per (user, word), separate from
-- user_cards, so forgetting how to SAY a word never resets the meaning schedule.
CREATE TABLE IF NOT EXISTS user_pron (
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id        BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    state          SMALLINT NOT NULL DEFAULT 0 CHECK (state BETWEEN 0 AND 3),
    stability      DOUBLE PRECISION NOT NULL DEFAULT 0,
    difficulty     DOUBLE PRECISION NOT NULL DEFAULT 0,
    due            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_review    TIMESTAMPTZ,
    scheduled_days INT NOT NULL DEFAULT 0,
    reps           INT NOT NULL DEFAULT 0,
    lapses         INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, card_id)
);
CREATE INDEX IF NOT EXISTS ix_user_pron_due ON user_pron (user_id, due);

-- Every pronunciation answer: source = stress (tapped a syllable) | forgot
-- ("I forgot how to say it"). rating 1 = wrong/forgot, 3 = right.
CREATE TABLE IF NOT EXISTS pron_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id     BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    rating      SMALLINT NOT NULL,
    source      TEXT NOT NULL,
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_pron_logs_user ON pron_logs (user_id, reviewed_at);
