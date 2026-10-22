-- Run MANUALLY, only after you have checked that the new app shows your old
-- words, phrasal verbs and progress correctly (migration 0005 copied them).
-- Take a backup first:  pg_dump -Fc -f before_drop.dump <db>
--
--   psql "$DATABASE_URL" -f docs/drop_legacy_tables.sql

BEGIN;
DROP VIEW  IF EXISTS user_dashboard;
DROP TABLE IF EXISTS user_phrasal_verb_progress;
DROP TABLE IF EXISTS phrasal_verb_examples;
DROP TABLE IF EXISTS phrasal_verbs;
DROP TABLE IF EXISTS notification_log;
DROP TABLE IF EXISTS quiz_answers;
DROP TABLE IF EXISTS quiz_attempts;
DROP TABLE IF EXISTS quiz_questions;
DROP TABLE IF EXISTS quizzes;
DROP TABLE IF EXISTS writing_errors;
DROP TABLE IF EXISTS writing_rewrites;
DROP TABLE IF EXISTS user_streaks;
DROP TABLE IF EXISTS review_sessions;
DROP TABLE IF EXISTS user_vocab_progress;
DROP TABLE IF EXISTS vocab_tag_mappings;
DROP TABLE IF EXISTS vocab_tags;
DROP TABLE IF EXISTS sentences;
DROP TABLE IF EXISTS vocabularies;
DROP TABLE IF EXISTS user_preferences;
DROP TABLE IF EXISTS prompt_templates;
-- writing_submissions.topic_id still references topic_categories in old databases
ALTER TABLE writing_submissions DROP COLUMN IF EXISTS topic_id;
ALTER TABLE writing_submissions DROP COLUMN IF EXISTS required_vocabs;
ALTER TABLE writing_submissions DROP COLUMN IF EXISTS used_vocabs;
DROP TABLE IF EXISTS topic_categories;
-- also from the very first version of the schema
DROP TABLE IF EXISTS user_vocabs;
DROP TABLE IF EXISTS study_sessions;
DROP TABLE IF EXISTS vocabs;
COMMIT;
