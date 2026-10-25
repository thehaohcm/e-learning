-- 0016_tts_audio: English words read aloud for the study device (WAV, made by
-- the "tts" container once and reused). Rows not played for TTS_KEEP_DAYS are
-- deleted automatically; they are made again when needed.
CREATE TABLE IF NOT EXISTS tts_audio (
    key          TEXT PRIMARY KEY,               -- sha256(voice | text), see internal/tts
    text         TEXT NOT NULL,
    voice        TEXT NOT NULL,
    wav          BYTEA NOT NULL,
    bytes        INT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS tts_audio_last_used_idx ON tts_audio (last_used_at);
