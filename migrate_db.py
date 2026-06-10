#!/usr/bin/env python3
"""
Database Migration Script for AI-Powered E-Learning Platform
Reads credentials from .env file.
"""

import os
import sys


def load_env():
    """Load .env file into os.environ"""
    env_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), ".env")
    if not os.path.exists(env_path):
        print("ERROR: .env file not found!")
        print("  cp .env.example .env")
        print("  Then edit .env with your credentials")
        sys.exit(1)

    with open(env_path) as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if "=" in line:
                key, _, value = line.partition("=")
                key = key.strip()
                value = value.strip().strip('"').strip("'")
                if key and value and key not in os.environ:
                    os.environ[key] = value


load_env()

DB_CONFIG = {
    "host": os.environ.get("DB_HOST"),
    "port": int(os.environ.get("DB_PORT", "5432")),
    "user": os.environ.get("DB_USER"),
    "password": os.environ.get("DB_PASSWORD"),
    "database": os.environ.get("DB_NAME"),
}

# Verify all required values
missing = [k for k, v in DB_CONFIG.items() if not v]
if missing:
    print(f"ERROR: Missing required env vars: {missing}")
    print("Check your .env file")
    sys.exit(1)

# All CREATE TABLE statements in order (from ARCHITECTURE.md)
MIGRATIONS = [
    # =========================================================================
    # MIGRATION 0001: CORE SCHEMA
    # =========================================================================
    """
    CREATE TABLE IF NOT EXISTS users (
        id              BIGSERIAL PRIMARY KEY,
        email           TEXT UNIQUE NOT NULL,
        display_name    TEXT,
        firebase_uid    TEXT UNIQUE,
        avatar_url      TEXT,
        timezone        TEXT DEFAULT 'UTC',
        is_active       BOOLEAN DEFAULT true,
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        updated_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_users_firebase_uid ON users(firebase_uid);",

    """
    CREATE TABLE IF NOT EXISTS topic_categories (
        id              SMALLSERIAL PRIMARY KEY,
        slug            TEXT UNIQUE NOT NULL,
        name_vi         TEXT NOT NULL,
        name_en         TEXT NOT NULL,
        description     TEXT,
        parent_id       INT REFERENCES topic_categories(id),
        sort_order      INT DEFAULT 0,
        created_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,

    "CREATE EXTENSION IF NOT EXISTS pg_trgm;",

    """
    CREATE TABLE IF NOT EXISTS vocabularies (
        id              BIGSERIAL PRIMARY KEY,
        word            TEXT NOT NULL,
        phonetics       TEXT,
        pos             TEXT,
        topic_id        INT REFERENCES topic_categories(id),
        audio_url       TEXT,
        image_url       TEXT,
        source          TEXT,
        meta_json       JSONB DEFAULT '{}',
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        updated_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_vocab_word ON vocabularies USING gin (word gin_trgm_ops);",
    "CREATE INDEX IF NOT EXISTS idx_vocab_topic ON vocabularies(topic_id);",
    "CREATE INDEX IF NOT EXISTS idx_vocab_meta ON vocabularies USING gin (meta_json);",

    """
    CREATE TABLE IF NOT EXISTS sentences (
        id              BIGSERIAL PRIMARY KEY,
        vocab_id        BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
        level           SMALLINT NOT NULL CHECK (level IN (1, 2, 3)),
        meaning_en      TEXT NOT NULL,
        meaning_vi      TEXT NOT NULL,
        example_en      TEXT NOT NULL,
        example_vi      TEXT NOT NULL,
        usage_note      TEXT,
        grammar_note    TEXT,
        audio_url       TEXT,
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        updated_at      TIMESTAMPTZ DEFAULT NOW(),
        UNIQUE(vocab_id, level)
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_sentences_vocab ON sentences(vocab_id);",

    """
    CREATE TABLE IF NOT EXISTS vocab_tags (
        id              SERIAL PRIMARY KEY,
        name            TEXT UNIQUE NOT NULL,
        created_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,

    """
    CREATE TABLE IF NOT EXISTS vocab_tag_mappings (
        vocab_id        BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
        tag_id          INT NOT NULL REFERENCES vocab_tags(id) ON DELETE CASCADE,
        PRIMARY KEY (vocab_id, tag_id)
    );
    """,

    """
    CREATE TABLE IF NOT EXISTS user_vocab_progress (
        id              BIGSERIAL PRIMARY KEY,
        user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        vocab_id        BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
        state           SMALLINT DEFAULT 0 NOT NULL,
        stability       DOUBLE PRECISION DEFAULT 2.5,
        difficulty      DOUBLE PRECISION DEFAULT 0.3,
        elapsed_days    INT DEFAULT 0,
        scheduled_days  INT DEFAULT 0,
        reps            INT DEFAULT 0,
        lapses          INT DEFAULT 0,
        last_review_at  TIMESTAMPTZ,
        next_review_at  TIMESTAMPTZ DEFAULT NOW(),
        level_focus     SMALLINT DEFAULT 1 CHECK (level_focus IN (1, 2, 3)),
        is_active       BOOLEAN DEFAULT true,
        mastery_score   DOUBLE PRECISION DEFAULT 0.0,
        history_json    JSONB DEFAULT '[]',
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        updated_at      TIMESTAMPTZ DEFAULT NOW(),
        UNIQUE(user_id, vocab_id)
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_uvp_next_review ON user_vocab_progress(user_id, next_review_at) WHERE is_active = true;",
    "CREATE INDEX IF NOT EXISTS idx_uvp_user_state ON user_vocab_progress(user_id, state);",
    "CREATE INDEX IF NOT EXISTS idx_uvp_due_count ON user_vocab_progress(next_review_at, is_active);",

    """
    CREATE TABLE IF NOT EXISTS review_sessions (
        id              BIGSERIAL PRIMARY KEY,
        user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        vocab_count     INT NOT NULL,
        duration_ms     BIGINT,
        avg_rating      DOUBLE PRECISION,
        session_type    TEXT DEFAULT 'full',
        created_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_rs_user ON review_sessions(user_id, created_at DESC);",

    """
    CREATE TABLE IF NOT EXISTS user_streaks (
        id              BIGSERIAL PRIMARY KEY,
        user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        streak_date     DATE NOT NULL,
        vocab_reviewed  INT DEFAULT 0,
        writing_done    BOOLEAN DEFAULT false,
        quiz_done       BOOLEAN DEFAULT false,
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        UNIQUE(user_id, streak_date)
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_streak_user_date ON user_streaks(user_id, streak_date DESC);",

    """
    CREATE TABLE IF NOT EXISTS writing_submissions (
        id              BIGSERIAL PRIMARY KEY,
        user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        topic_id        INT REFERENCES topic_categories(id),
        prompt          TEXT NOT NULL,
        original_text   TEXT NOT NULL,
        word_count      INT,
        required_vocabs JSONB,
        used_vocabs     JSONB,
        overall_score   DOUBLE PRECISION,
        grammar_score   DOUBLE PRECISION,
        vocab_score     DOUBLE PRECISION,
        coherence_score DOUBLE PRECISION,
        ai_feedback_json JSONB DEFAULT '{}',
        ai_model_used   TEXT,
        ai_tokens_used  INT,
        status          TEXT DEFAULT 'pending',
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        completed_at    TIMESTAMPTZ
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_ws_user ON writing_submissions(user_id, created_at DESC);",
    "CREATE INDEX IF NOT EXISTS idx_ws_status ON writing_submissions(status) WHERE status IN ('pending', 'processing');",

    """
    CREATE TABLE IF NOT EXISTS writing_errors (
        id              BIGSERIAL PRIMARY KEY,
        submission_id   BIGINT NOT NULL REFERENCES writing_submissions(id) ON DELETE CASCADE,
        error_type      TEXT NOT NULL,
        error_span      TEXT,
        correction      TEXT,
        explanation     TEXT,
        severity        SMALLINT DEFAULT 1,
        position_start  INT,
        position_end    INT,
        created_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,

    """
    CREATE TABLE IF NOT EXISTS writing_rewrites (
        id              BIGSERIAL PRIMARY KEY,
        submission_id   BIGINT NOT NULL REFERENCES writing_submissions(id) ON DELETE CASCADE,
        level           SMALLINT NOT NULL CHECK (level IN (1, 2, 3)),
        rewritten_text  TEXT NOT NULL,
        explanation     TEXT,
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        UNIQUE(submission_id, level)
    );
    """,

    """
    CREATE TABLE IF NOT EXISTS quizzes (
        id              BIGSERIAL PRIMARY KEY,
        title           TEXT,
        description     TEXT,
        quiz_type       TEXT NOT NULL DEFAULT 'vocab',
        topic_id        INT REFERENCES topic_categories(id),
        difficulty      TEXT DEFAULT 'medium',
        question_count  INT DEFAULT 10,
        time_limit_min  INT,
        is_ai_generated BOOLEAN DEFAULT true,
        ai_prompt_used  TEXT,
        created_by      BIGINT REFERENCES users(id),
        created_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,

    """
    CREATE TABLE IF NOT EXISTS quiz_questions (
        id              BIGSERIAL PRIMARY KEY,
        quiz_id         BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
        vocab_id        BIGINT REFERENCES vocabularies(id),
        question_type   TEXT NOT NULL DEFAULT 'multiple_choice',
        prompt          TEXT NOT NULL,
        choices         JSONB NOT NULL,
        explanation     TEXT,
        sort_order      INT DEFAULT 0,
        created_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_qq_quiz ON quiz_questions(quiz_id, sort_order);",

    """
    CREATE TABLE IF NOT EXISTS quiz_attempts (
        id              BIGSERIAL PRIMARY KEY,
        user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        quiz_id         BIGINT NOT NULL REFERENCES quizzes(id),
        score           INT,
        total_questions INT,
        duration_ms     BIGINT,
        status          TEXT DEFAULT 'started',
        started_at      TIMESTAMPTZ DEFAULT NOW(),
        completed_at    TIMESTAMPTZ
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_qa_user_quiz ON quiz_attempts(user_id, quiz_id);",

    """
    CREATE TABLE IF NOT EXISTS quiz_answers (
        id              BIGSERIAL PRIMARY KEY,
        attempt_id      BIGINT NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
        question_id     BIGINT NOT NULL REFERENCES quiz_questions(id),
        selected_idx    INT,
        is_correct      BOOLEAN,
        time_spent_ms   INT,
        created_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_qanswer_attempt ON quiz_answers(attempt_id);",

    """
    CREATE TABLE IF NOT EXISTS notification_log (
        id              BIGSERIAL PRIMARY KEY,
        user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        vocab_id        BIGINT NOT NULL REFERENCES vocabularies(id),
        level           SMALLINT,
        channel         TEXT DEFAULT 'push',
        status          TEXT DEFAULT 'sent',
        error_message   TEXT,
        created_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,
    "CREATE INDEX IF NOT EXISTS idx_nl_user_vocab_time ON notification_log(user_id, vocab_id, created_at DESC);",

    """
    CREATE TABLE IF NOT EXISTS user_preferences (
        id              BIGSERIAL PRIMARY KEY,
        user_id         BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        daily_vocab_goal INT DEFAULT 10,
        reminder_enabled BOOLEAN DEFAULT true,
        reminder_times  JSONB DEFAULT '["09:00", "13:00", "19:00"]',
        quiz_difficulty TEXT DEFAULT 'medium',
        tts_speed       DOUBLE PRECISION DEFAULT 1.0,
        ui_language     TEXT DEFAULT 'vi',
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        updated_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,

    """
    CREATE TABLE IF NOT EXISTS prompt_templates (
        id              SERIAL PRIMARY KEY,
        name            TEXT UNIQUE NOT NULL,
        description     TEXT,
        system_prompt   TEXT NOT NULL,
        user_prompt_template TEXT,
        ai_model        TEXT NOT NULL,
        max_tokens      INT DEFAULT 2048,
        temperature     DOUBLE PRECISION DEFAULT 0.7,
        version         INT DEFAULT 1,
        is_active       BOOLEAN DEFAULT true,
        created_at      TIMESTAMPTZ DEFAULT NOW(),
        updated_at      TIMESTAMPTZ DEFAULT NOW()
    );
    """,

    """
    CREATE OR REPLACE VIEW user_dashboard AS
    SELECT
        u.id AS user_id,
        u.email,
        COUNT(uvp.id) FILTER (WHERE uvp.is_active = true) AS active_vocabs,
        COUNT(uvp.id) FILTER (WHERE uvp.mastery_score >= 80) AS mastered_vocabs,
        COUNT(uvp.id) FILTER (WHERE uvp.next_review_at <= NOW()) AS due_reviews,
        COALESCE(
            (SELECT COUNT(*) FROM user_streaks us
             WHERE us.user_id = u.id
               AND us.streak_date >= CURRENT_DATE - INTERVAL '7 days'), 0
        ) AS streak_7d,
        (SELECT MAX(ws.overall_score) FROM writing_submissions ws
         WHERE ws.user_id = u.id
           AND ws.created_at >= NOW() - INTERVAL '7 days') AS best_writing_score_7d,
        (SELECT MAX(qa.score) FROM quiz_attempts qa
         WHERE qa.user_id = u.id
           AND qa.started_at >= NOW() - INTERVAL '7 days') AS best_quiz_score_7d
    FROM users u
    LEFT JOIN user_vocab_progress uvp ON uvp.user_id = u.id
    GROUP BY u.id, u.email;
    """,
]


def run_migration():
    try:
        import psycopg2
    except ImportError:
        print("ERROR: psycopg2 not installed. Run: pip install psycopg2-binary")
        sys.exit(1)

    print(f"Connecting to PostgreSQL at {DB_CONFIG['host']}:{DB_CONFIG['port']}...")
    try:
        conn = psycopg2.connect(**DB_CONFIG)
        conn.autocommit = True
        cur = conn.cursor()
        print("Connected successfully!")
    except Exception as e:
        print(f"Connection failed: {e}")
        sys.exit(1)

    total = len(MIGRATIONS)
    success = 0
    failed = 0

    for i, sql in enumerate(MIGRATIONS, 1):
        first_line = sql.strip().split('\n')[0].strip()
        if first_line.startswith('--'):
            first_line = sql.strip().split('\n')[1].strip()

        try:
            cur.execute(sql)
            success += 1
            print(f"[{i:02d}/{total:02d}] OK  {first_line[:80]}")
        except Exception as e:
            failed += 1
            print(f"[{i:02d}/{total:02d}] FAIL {first_line[:80]}")
            print(f"          Error: {str(e)[:120]}")

    cur.close()
    conn.close()

    print(f"\n{'='*60}")
    print(f"Migration completed: {success} success, {failed} failed out of {total}")

    conn2 = psycopg2.connect(**DB_CONFIG)
    cur2 = conn2.cursor()
    cur2.execute("""
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = 'public'
        ORDER BY table_name;
    """)
    tables = [row[0] for row in cur2.fetchall()]
    cur2.close()
    conn2.close()

    print(f"\nTables created ({len(tables)}):")
    for t in tables:
        print(f"  - {t}")

    expected = [
        "users", "topic_categories", "vocabularies", "sentences",
        "vocab_tags", "vocab_tag_mappings", "user_vocab_progress",
        "review_sessions", "user_streaks", "writing_submissions",
        "writing_errors", "writing_rewrites", "quizzes", "quiz_questions",
        "quiz_attempts", "quiz_answers", "notification_log",
        "user_preferences", "prompt_templates",
    ]
    missing = [t for t in expected if t not in tables]
    if missing:
        print(f"\nWARNING: Missing tables: {missing}")
    else:
        print(f"\nAll {len(expected)} expected tables created successfully!")


if __name__ == "__main__":
    run_migration()