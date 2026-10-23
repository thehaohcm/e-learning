-- 0011_push: Web Push study reminders.
-- app_secrets keeps server-generated keys (the VAPID key pair for Web Push is
-- created on first start, so nothing has to be configured in .env).
CREATE TABLE IF NOT EXISTS app_secrets (
    name       TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- One row per browser/phone that allowed notifications.
CREATE TABLE IF NOT EXISTS push_subscriptions (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    endpoint     TEXT NOT NULL UNIQUE,
    p256dh       TEXT NOT NULL,
    auth         TEXT NOT NULL,
    user_agent   TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_sent_at TIMESTAMPTZ,
    fail_count   INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS push_subscriptions_user_idx ON push_subscriptions (user_id);

-- Reminder slots already handled ("2026-10-23 20:00" in the user's time zone),
-- so a reminder is never sent twice (restarts, several API instances).
CREATE TABLE IF NOT EXISTS push_reminders_sent (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot    TEXT NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, slot)
);
