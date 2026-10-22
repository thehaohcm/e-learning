-- 0007_practice: extra example sentences per card, written by AI, for the
-- "practise forgotten words" drill (translate to Vietnamese, then read aloud).
-- Sentences are shared by all users; each user rotates through them.

CREATE TABLE IF NOT EXISTS card_examples (
    id             BIGSERIAL PRIMARY KEY,
    card_id        BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    sentence_en    TEXT NOT NULL,
    translation_vi TEXT NOT NULL DEFAULT '',
    source         TEXT NOT NULL DEFAULT 'ai',   -- ai | card (the card's own example)
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_card_examples_sentence ON card_examples (card_id, lower(sentence_en));

-- How many times a user has practised a card: picks the next set of sentences.
CREATE TABLE IF NOT EXISTS practice_rounds (
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_id    BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    round      INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, card_id)
);
