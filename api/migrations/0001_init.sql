-- 0001_init: core schema shared by every subject (English, PMP, ...).
--
-- Idea: every piece of learnable content is a CARD that lives in a DECK, and a
-- deck belongs to a SUBJECT. A card has a `kind` that decides how it is shown
-- (vocab, phrasal, concept, mcq) and a JSONB `payload` with kind-specific data.
-- Spaced-repetition state is stored per (user, card) in user_cards, so the same
-- FSRS scheduler works for every subject.
--
-- Written to be safe on a fresh database AND on the legacy database created by
-- the old migrate_db.py (all CREATE ... IF NOT EXISTS / ADD COLUMN IF NOT EXISTS).

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ---------------------------------------------------------------- users
CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL PRIMARY KEY,
    email      TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE users ADD COLUMN IF NOT EXISTS display_name  TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS role          TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone      TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS settings      JSONB;
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at    TIMESTAMPTZ;

UPDATE users SET display_name  = COALESCE(display_name, ''),
                 password_hash = COALESCE(password_hash, ''),
                 role          = COALESCE(role, 'user'),
                 timezone      = CASE WHEN timezone IS NULL OR timezone = '' OR timezone = 'UTC'
                                      THEN 'Asia/Ho_Chi_Minh' ELSE timezone END,
                 settings      = COALESCE(settings, '{}'::jsonb),
                 updated_at    = COALESCE(updated_at, created_at, NOW());

ALTER TABLE users ALTER COLUMN display_name  SET DEFAULT '', ALTER COLUMN display_name  SET NOT NULL;
-- '' means "no password yet" (legacy accounts can claim it once via /auth/register).
ALTER TABLE users ALTER COLUMN password_hash SET DEFAULT '', ALTER COLUMN password_hash SET NOT NULL;
ALTER TABLE users ALTER COLUMN role          SET DEFAULT 'user', ALTER COLUMN role SET NOT NULL;
ALTER TABLE users ALTER COLUMN timezone      SET DEFAULT 'Asia/Ho_Chi_Minh', ALTER COLUMN timezone SET NOT NULL;
ALTER TABLE users ALTER COLUMN settings      SET DEFAULT '{}'::jsonb, ALTER COLUMN settings SET NOT NULL;
ALTER TABLE users ALTER COLUMN updated_at    SET DEFAULT NOW(), ALTER COLUMN updated_at SET NOT NULL;
ALTER TABLE users ALTER COLUMN created_at    SET DEFAULT NOW();

-- ---------------------------------------------------------------- content
CREATE TABLE IF NOT EXISTS subjects (
    id          BIGSERIAL PRIMARY KEY,
    slug        TEXT UNIQUE NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order  INT  NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS decks (
    id           BIGSERIAL PRIMARY KEY,
    subject_id   BIGINT NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    slug         TEXT NOT NULL,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    -- the card kind new content in this deck should have (used by AI generation)
    default_kind TEXT NOT NULL DEFAULT 'vocab'
                 CHECK (default_kind IN ('vocab', 'phrasal', 'concept', 'mcq')),
    sort_order   INT  NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (subject_id, slug)
);

CREATE TABLE IF NOT EXISTS cards (
    id         BIGSERIAL PRIMARY KEY,
    deck_id    BIGINT NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
    kind       TEXT   NOT NULL CHECK (kind IN ('vocab', 'phrasal', 'concept', 'mcq')),
    -- vocab: the word | phrasal: "give up" | concept: the term | mcq: the question stem
    front      TEXT   NOT NULL CHECK (length(trim(front)) > 0),
    payload    JSONB  NOT NULL DEFAULT '{}'::jsonb,
    source     TEXT   NOT NULL DEFAULT 'manual',   -- manual | seed | ai | legacy
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_cards_deck_kind_front ON cards (deck_id, kind, lower(front));
CREATE INDEX IF NOT EXISTS ix_cards_front_trgm ON cards USING gin (front gin_trgm_ops);

-- ---------------------------------------------------------------- devices (ESP32)
CREATE TABLE IF NOT EXISTS devices (
    id                   BIGSERIAL PRIMARY KEY,
    user_id              BIGINT REFERENCES users(id) ON DELETE CASCADE,  -- NULL until paired
    name                 TEXT NOT NULL DEFAULT 'Study device',
    token_hash           TEXT UNIQUE NOT NULL,      -- sha256 of the device's secret token
    pair_code            TEXT,
    pair_code_expires_at TIMESTAMPTZ,
    firmware             TEXT NOT NULL DEFAULT '',
    settings             JSONB NOT NULL DEFAULT '{}'::jsonb,
    last_seen_at         TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_devices_pair_code ON devices (pair_code) WHERE pair_code IS NOT NULL;
CREATE INDEX IF NOT EXISTS ix_devices_user ON devices (user_id);

-- ---------------------------------------------------------------- learning state
CREATE TABLE IF NOT EXISTS user_cards (
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id        BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    -- FSRS state: 0 new, 1 learning, 2 review, 3 relearning
    state          SMALLINT NOT NULL DEFAULT 0 CHECK (state BETWEEN 0 AND 3),
    stability      DOUBLE PRECISION NOT NULL DEFAULT 0,
    difficulty     DOUBLE PRECISION NOT NULL DEFAULT 0,
    due            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_review    TIMESTAMPTZ,
    scheduled_days INT NOT NULL DEFAULT 0,
    reps           INT NOT NULL DEFAULT 0,
    lapses         INT NOT NULL DEFAULT 0,
    suspended      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, card_id)
);
CREATE INDEX IF NOT EXISTS ix_user_cards_due ON user_cards (user_id, due) WHERE NOT suspended;

-- Append-only history of every answer. Used for stats/streaks today and for
-- optimising FSRS weights per user later.
CREATE TABLE IF NOT EXISTS review_logs (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id          BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    rating           SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 4),
    state_before     SMALLINT NOT NULL,
    elapsed_days     INT NOT NULL DEFAULT 0,
    scheduled_days   INT NOT NULL DEFAULT 0,
    duration_ms      INT,
    source           TEXT NOT NULL DEFAULT 'web',   -- web | device | quiz
    device_id        BIGINT REFERENCES devices(id) ON DELETE SET NULL,
    client_review_id TEXT,                          -- idempotency key sent by the device
    reviewed_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS ix_review_logs_user_time ON review_logs (user_id, reviewed_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS ux_review_logs_client
    ON review_logs (user_id, client_review_id) WHERE client_review_id IS NOT NULL;

-- ---------------------------------------------------------------- writing lab (English)
CREATE TABLE IF NOT EXISTS writing_submissions (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    prompt           TEXT NOT NULL,
    original_text    TEXT NOT NULL,
    word_count       INT,
    overall_score    DOUBLE PRECISION,
    grammar_score    DOUBLE PRECISION,
    vocab_score      DOUBLE PRECISION,
    coherence_score  DOUBLE PRECISION,
    ai_feedback_json JSONB DEFAULT '{}'::jsonb,
    ai_model_used    TEXT,
    status           TEXT DEFAULT 'pending',
    created_at       TIMESTAMPTZ DEFAULT NOW(),
    completed_at     TIMESTAMPTZ
);
ALTER TABLE writing_submissions ADD COLUMN IF NOT EXISTS error_message TEXT;
CREATE INDEX IF NOT EXISTS ix_writing_user_time ON writing_submissions (user_id, created_at DESC);
UPDATE writing_submissions SET word_count = COALESCE(word_count, 0), status = COALESCE(status, 'pending')
 WHERE word_count IS NULL OR status IS NULL;
