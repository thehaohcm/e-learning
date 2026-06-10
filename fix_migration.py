#!/usr/bin/env python3
"""Fix migration issues - reads credentials from .env"""

import os
import sys
import psycopg2


def load_env():
    env_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), ".env")
    if not os.path.exists(env_path):
        print("ERROR: .env file not found!")
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

FIXES = [
    """
    DROP INDEX IF EXISTS idx_uvp_due_count;
    CREATE INDEX IF NOT EXISTS idx_uvp_due_count ON user_vocab_progress(next_review_at, is_active);
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

print("Applying fixes...")
conn = psycopg2.connect(**DB_CONFIG)
conn.autocommit = True
cur = conn.cursor()

for i, sql in enumerate(FIXES, 1):
    try:
        cur.execute(sql)
        first_line = sql.strip().split('\n')[0].strip()
        print(f"[FIX {i}] OK  {first_line[:80]}")
    except Exception as e:
        print(f"[FIX {i}] FAIL  Error: {e}")

cur.close()

print("\nVerifying view...")
cur2 = conn.cursor()
cur2.execute("SELECT * FROM user_dashboard LIMIT 0;")
cols = [desc[0] for desc in cur2.description]
print(f"user_dashboard columns: {cols}")

cur2.execute("""
    SELECT table_name FROM information_schema.tables
    WHERE table_schema = 'public' ORDER BY table_name;
""")
tables = [row[0] for row in cur2.fetchall()]
print(f"\nTotal tables: {len(tables)}")
for t in tables:
    print(f"  - {t}")

cur2.close()
conn.close()
print("\nAll fixes applied successfully!")