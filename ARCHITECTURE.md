# Architecture

(The previous 73 KB design document described a 24-week plan that the code
never matched. This file describes what is actually implemented. See README.md
for setup and the device protocol.)

## Components

```
 Browser (Vue 3)          ESP32-S3 device
      │  JWT                    │  device token (dev_...)
      ▼                         ▼
   nginx (web container, :8085) ── /api ──► Go API (:8000) ──► PostgreSQL
                                               │
                                               └──► DeepSeek (optional: card generation, writing feedback)
```

- One Go binary. On start it connects to Postgres and applies `api/migrations/*.sql`
  in name order inside transactions, guarded by an advisory lock. There is no
  other place that changes the schema (no AutoMigrate, no Python scripts).
- No background job is needed to find due cards: `user_cards(user_id, due)` is
  indexed and the queue is computed per request. A small hourly worker only
  cleans up never-paired devices and stuck writing evaluations.
- Redis was removed (it was configured but unused).

## Study flow

1. `GET /api/study/queue`: due cards (learning/review, `due <= now`) first, then
   new cards up to `settings.daily_new` per day (counted from `review_logs` in the
   user's time zone). If the user has not added enough cards, new ones are taken
   automatically from `settings.active_deck_ids` (default: English core vocabulary).
2. `POST /api/study/review {card_id, rating}`: `internal/srs` runs FSRS
   (go-fsrs v3, 90% target retention, short-term learning steps 1m/5m/10m, fuzz on)
   inside a transaction with a row lock, then appends a `review_logs` row.
3. Quizzes build 4-choice questions from `mcq` cards or from other cards of the
   same deck (meaning ↔ term). A quiz answer on a card that is due counts as a
   review (correct = Good, wrong = Again); otherwise the schedule is untouched.

## Why one generic `cards` table

The old schema had a separate set of tables, progress table and handlers per
content type (vocabularies, phrasal verbs). Adding PMP would have meant copying
all of it again. Now:

- content differences live in `cards.kind` + `cards.payload` (validated in
  `internal/content`),
- scheduling, quiz, stats and the device API are written once,
- a new subject is data (subject + decks), not code.

## Security notes

- Passwords: bcrypt. Sessions: HS256 JWT (30 days), secret from `JWT_SECRET`
  (required in production).
- Devices: random 256-bit token, only its SHA-256 is stored. Pairing codes are
  6 digits, valid 15 minutes, and require a logged-in user to claim.
- Admin-only: creating, editing and deleting cards, AI generation.
- Login and device registration are rate-limited per IP.
- Errors returned to clients are generic; details go to the server log.
