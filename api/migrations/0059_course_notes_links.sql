-- 0059_course_notes_links: add a "links" jsonb column to course_notes so each
-- cheat-sheet item can carry resource links [{url,label}] to track later.
-- Existing rows get an empty array via the default. Idempotent.
ALTER TABLE course_notes ADD COLUMN IF NOT EXISTS links jsonb NOT NULL DEFAULT '[]'::jsonb;
