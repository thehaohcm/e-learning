-- 0017_word_questions: AI multiple-choice questions about a word (usage in a
-- sentence, closest meaning, collocation), made when the learner practises the
-- words they missed on the study device, then reused.
CREATE TABLE IF NOT EXISTS word_questions (
    id          BIGSERIAL PRIMARY KEY,
    card_id     BIGINT NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    kind        TEXT NOT NULL,              -- usage | synonym | collocation
    prompt      TEXT NOT NULL,
    choices     JSONB NOT NULL,             -- 4 strings
    answer      INT NOT NULL,               -- index of the right choice
    explanation TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS word_questions_card_idx ON word_questions (card_id);
