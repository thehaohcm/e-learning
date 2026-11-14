-- 0088_daily_sets_subject_and_multi: generalize the "Daily 20" sets so (a) more
-- than one set can exist per day (the learner can generate another batch to
-- take, and can re-take a saved set in the future), and (b) sets belong to a
-- subject (PMP and now English each have their own Daily 20).
--
-- Adds subject_id (back-filled to PMP for existing rows) and seq (the batch
-- number within a day), and drops the old one-set-per-day UNIQUE(user_id, day)
-- so multiple sets per day are allowed. The AI review text is stored inside the
-- existing result JSONB (key "ai_review"), so no column is needed for it.
-- Idempotent and safe to re-run.

ALTER TABLE pmp_daily_sets ADD COLUMN IF NOT EXISTS subject_id BIGINT;
ALTER TABLE pmp_daily_sets ADD COLUMN IF NOT EXISTS seq INT NOT NULL DEFAULT 1;

-- existing rows were all PMP
UPDATE pmp_daily_sets
   SET subject_id = (SELECT id FROM subjects WHERE slug = 'pmp' LIMIT 1)
 WHERE subject_id IS NULL;

-- drop whatever UNIQUE constraint the table has (the original UNIQUE(user_id,day)),
-- found by name so this works regardless of the auto-generated constraint name
DO $$
DECLARE c text;
BEGIN
  SELECT conname INTO c
    FROM pg_constraint
   WHERE conrelid = 'pmp_daily_sets'::regclass AND contype = 'u'
   LIMIT 1;
  IF c IS NOT NULL THEN
    EXECUTE 'ALTER TABLE pmp_daily_sets DROP CONSTRAINT ' || quote_ident(c);
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS ix_daily_user_subject_day
    ON pmp_daily_sets (user_id, subject_id, day DESC, seq DESC);
