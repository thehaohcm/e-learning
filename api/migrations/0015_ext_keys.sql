-- 0015_ext_keys: keys for the browser extension ("Add to E-Learn notes").
-- A key can only add/read the user's notes; only its SHA-256 hash is stored.
CREATE TABLE IF NOT EXISTS ext_keys (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name         TEXT NOT NULL DEFAULT '',
    token_hash   TEXT NOT NULL UNIQUE,
    prefix       TEXT NOT NULL DEFAULT '',      -- first characters, to recognise the key in Settings
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS ext_keys_user_idx ON ext_keys (user_id);
