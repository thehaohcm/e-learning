-- 0056_remove_untagged_cheatsheet_labs: delete the AI-generated typed-command
-- (lab) drills in the Pentest "cheatsheet" deck that carry no section tag
-- (payload.topic empty/null). They were generated before cards were tagged with
-- their cheatsheet section, so they all fall under "(không mục)" in the Study
-- practice picker. Removing them lets the user regenerate per-section with
-- Call AI. Tagged drills and hand-made cards (source <> 'ai') are kept.
-- Idempotent: re-running deletes nothing once they are gone.

-- user_cards first (lab cards are normally never enrolled, but keep it safe
-- against any stray enrolment before the FK blocks the delete).
DELETE FROM user_cards uc
USING cards c, decks d, subjects s
WHERE uc.card_id = c.id
  AND c.deck_id = d.id AND d.subject_id = s.id
  AND s.slug = 'pentest' AND d.slug = 'cheatsheet'
  AND c.kind = 'lab' AND c.source = 'ai'
  AND COALESCE(NULLIF(c.payload->>'topic', ''), '') = '';

DELETE FROM cards c
USING decks d, subjects s
WHERE c.deck_id = d.id AND d.subject_id = s.id
  AND s.slug = 'pentest' AND d.slug = 'cheatsheet'
  AND c.kind = 'lab' AND c.source = 'ai'
  AND COALESCE(NULLIF(c.payload->>'topic', ''), '') = '';
