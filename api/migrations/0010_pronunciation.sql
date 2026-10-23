-- 0010_pronunciation: how to say each English word, for Vietnamese learners:
-- syllables with the stressed one marked (hy·POTH·e·sis) and short tips in
-- Vietnamese about the hard sounds. Written once by AI, shared by all users.
CREATE TABLE IF NOT EXISTS card_pronunciations (
    card_id    BIGINT PRIMARY KEY REFERENCES cards(id) ON DELETE CASCADE,
    syllables  JSONB NOT NULL DEFAULT '[]'::jsonb, -- ["hy","poth","e","sis"]
    stress     INT   NOT NULL DEFAULT 0,           -- index of the stressed syllable
    tips_vi    JSONB NOT NULL DEFAULT '[]'::jsonb, -- ["th /θ/: đặt đầu lưỡi giữa hai răng..."]
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
