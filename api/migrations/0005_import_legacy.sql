-- 0005_import_legacy: copy data from the old per-type tables (created by the
-- old migrate_db.py) into the new generic tables. Runs once. On a fresh
-- database the old tables do not exist and this migration does nothing.
--
-- The old tables are NOT dropped here, so nothing is lost. Once you have
-- checked the new data, you can drop them with docs/drop_legacy_tables.sql.
--
-- About the imported FSRS state: the old scheduler stored difficulty on a 0-1
-- scale and computed intervals ~6x too short, so those numbers are not valid
-- FSRS values. Cards that had been reviewed become "review" cards with a
-- neutral difficulty (5) and stability equal to their last interval; their due
-- date is kept. FSRS corrects the values after the next couple of reviews.

DO $$
DECLARE
  vocab_deck   BIGINT;
  phrasal_deck BIGINT;
BEGIN
  SELECT d.id INTO vocab_deck FROM decks d JOIN subjects s ON s.id = d.subject_id
   WHERE s.slug = 'english' AND d.slug = 'core-vocab';
  SELECT d.id INTO phrasal_deck FROM decks d JOIN subjects s ON s.id = d.subject_id
   WHERE s.slug = 'english' AND d.slug = 'phrasal-verbs';

  -- ---------------------------------------------------------- vocabularies
  IF to_regclass('public.vocabularies') IS NOT NULL AND to_regclass('public.sentences') IS NOT NULL THEN
    INSERT INTO cards (deck_id, kind, front, payload, source, created_at)
    SELECT DISTINCT ON (lower(trim(v.word)))
           vocab_deck, 'vocab', trim(v.word),
           jsonb_strip_nulls(jsonb_build_object(
             'ipa', NULLIF(trim(v.phonetics), ''),
             'pos', NULLIF(trim(v.pos), ''),
             'senses', COALESCE((
                SELECT jsonb_agg(jsonb_build_object(
                         'level', s.level, 'meaning_en', s.meaning_en, 'meaning_vi', s.meaning_vi,
                         'example_en', s.example_en, 'example_vi', s.example_vi,
                         'usage_note', NULLIF(s.usage_note, ''),
                         'grammar_note', NULLIF(s.grammar_note, '')) ORDER BY s.level)
                FROM sentences s WHERE s.vocab_id = v.id), '[]'::jsonb))),
           CASE WHEN v.source = 'ai-generated' THEN 'ai' ELSE 'legacy' END,
           COALESCE(v.created_at, NOW())
    FROM vocabularies v
    WHERE trim(COALESCE(v.word, '')) <> ''
    ORDER BY lower(trim(v.word)),
             (SELECT COUNT(*) FROM sentences s WHERE s.vocab_id = v.id) DESC, v.id
    ON CONFLICT (deck_id, kind, lower(front)) DO UPDATE
       SET payload = EXCLUDED.payload, source = EXCLUDED.source, updated_at = NOW()
     WHERE cards.source = 'seed'
       AND jsonb_array_length(EXCLUDED.payload->'senses') > jsonb_array_length(cards.payload->'senses');

    IF to_regclass('public.user_vocab_progress') IS NOT NULL THEN
      INSERT INTO user_cards (user_id, card_id, state, stability, difficulty, due, last_review,
                              scheduled_days, reps, lapses, suspended, created_at, updated_at)
      SELECT DISTINCT ON (p.user_id, c.id)
             p.user_id, c.id,
             CASE WHEN COALESCE(p.reps, 0) > 0 THEN 2 ELSE 0 END,
             CASE WHEN COALESCE(p.reps, 0) > 0 THEN GREATEST(1, COALESCE(p.scheduled_days, 1)) ELSE 0 END,
             CASE WHEN COALESCE(p.reps, 0) > 0 THEN 5 ELSE 0 END,
             COALESCE(p.next_review_at, NOW()),
             CASE WHEN COALESCE(p.reps, 0) > 0 THEN COALESCE(p.last_review_at, NOW()) END,
             COALESCE(p.scheduled_days, 0), COALESCE(p.reps, 0), COALESCE(p.lapses, 0),
             NOT COALESCE(p.is_active, TRUE),
             COALESCE(p.created_at, NOW()), NOW()
      FROM user_vocab_progress p
      JOIN vocabularies v ON v.id = p.vocab_id
      JOIN cards c ON c.deck_id = vocab_deck AND c.kind = 'vocab' AND lower(c.front) = lower(trim(v.word))
      JOIN users u ON u.id = p.user_id
      ORDER BY p.user_id, c.id, p.reps DESC NULLS LAST
      ON CONFLICT (user_id, card_id) DO NOTHING;
    END IF;
  END IF;

  -- ---------------------------------------------------------- phrasal verbs
  IF to_regclass('public.phrasal_verbs') IS NOT NULL AND to_regclass('public.phrasal_verb_examples') IS NOT NULL THEN
    INSERT INTO cards (deck_id, kind, front, payload, source, created_at)
    SELECT DISTINCT ON (lower(trim(pv.verb) || ' ' || trim(pv.particle)))
           phrasal_deck, 'phrasal', trim(pv.verb) || ' ' || trim(pv.particle),
           jsonb_build_object(
             'verb', trim(pv.verb), 'particle', trim(pv.particle),
             'separable', COALESCE(pv.separable, FALSE),
             'senses', COALESCE((
                SELECT jsonb_agg(jsonb_strip_nulls(jsonb_build_object(
                         'level', e.level, 'meaning_en', e.meaning_en, 'meaning_vi', e.meaning_vi,
                         'example_en', e.example_en, 'example_vi', e.example_vi,
                         'usage_note', NULLIF(e.usage_note, ''))) ORDER BY e.level)
                FROM phrasal_verb_examples e WHERE e.phrasal_verb_id = pv.id), '[]'::jsonb)),
           CASE WHEN pv.source = 'ai-generated' THEN 'ai' ELSE 'legacy' END,
           COALESCE(pv.created_at, NOW())
    FROM phrasal_verbs pv
    WHERE trim(COALESCE(pv.verb, '')) <> '' AND trim(COALESCE(pv.particle, '')) <> ''
    ORDER BY lower(trim(pv.verb) || ' ' || trim(pv.particle)),
             (SELECT COUNT(*) FROM phrasal_verb_examples e WHERE e.phrasal_verb_id = pv.id) DESC, pv.id
    ON CONFLICT (deck_id, kind, lower(front)) DO UPDATE
       SET payload = EXCLUDED.payload, source = EXCLUDED.source, updated_at = NOW()
     WHERE cards.source = 'seed'
       AND jsonb_array_length(EXCLUDED.payload->'senses') > jsonb_array_length(cards.payload->'senses');

    IF to_regclass('public.user_phrasal_verb_progress') IS NOT NULL THEN
      INSERT INTO user_cards (user_id, card_id, state, stability, difficulty, due, last_review,
                              scheduled_days, reps, lapses, suspended, created_at, updated_at)
      SELECT DISTINCT ON (p.user_id, c.id)
             p.user_id, c.id,
             CASE WHEN COALESCE(p.reps, 0) > 0 THEN 2 ELSE 0 END,
             CASE WHEN COALESCE(p.reps, 0) > 0 THEN GREATEST(1, COALESCE(p.scheduled_days, 1)) ELSE 0 END,
             CASE WHEN COALESCE(p.reps, 0) > 0 THEN 5 ELSE 0 END,
             COALESCE(p.next_review_at, NOW()),
             CASE WHEN COALESCE(p.reps, 0) > 0 THEN COALESCE(p.last_review_at, NOW()) END,
             COALESCE(p.scheduled_days, 0), COALESCE(p.reps, 0), COALESCE(p.lapses, 0),
             NOT COALESCE(p.is_active, TRUE),
             COALESCE(p.created_at, NOW()), NOW()
      FROM user_phrasal_verb_progress p
      JOIN phrasal_verbs pv ON pv.id = p.phrasal_verb_id
      JOIN cards c ON c.deck_id = phrasal_deck AND c.kind = 'phrasal'
                  AND lower(c.front) = lower(trim(pv.verb) || ' ' || trim(pv.particle))
      JOIN users u ON u.id = p.user_id
      ORDER BY p.user_id, c.id, p.reps DESC NULLS LAST
      ON CONFLICT (user_id, card_id) DO NOTHING;
    END IF;
  END IF;
END $$;
