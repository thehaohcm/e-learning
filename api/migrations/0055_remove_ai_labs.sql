-- 0055_remove_ai_labs: delete the AI-generated command drills that the earlier
-- "Call AI (typed-command)" flow created inside the Pentest "Labs" deck. Those
-- drills now go to the Cheatsheet deck (studied via SRS) instead, so the ones
-- left in the Labs deck are clutter. Hand-authored labs/assignments and the
-- challenge self-quizzes use source 'manual' and are kept. Idempotent.

DELETE FROM cards
WHERE kind = 'lab'
  AND source = 'ai'
  AND deck_id IN (
    SELECT d.id FROM decks d
    JOIN subjects s ON s.id = d.subject_id
    WHERE s.slug = 'pentest' AND d.slug = 'labs'
  );
