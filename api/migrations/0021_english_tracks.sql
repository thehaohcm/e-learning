-- 0021_english_tracks: extra English decks for two parallel study tracks the
-- learner chose — TOEIC (workplace register, for the cert) and Conversational
-- (everyday / forum / comment register, for natural writing). Decks are created
-- empty; cards are generated with the in-app AI, which reads each deck's name
-- and description to target the right words and register. IELTS is intentionally
-- left out for now (added later when the learner moves on to it).

INSERT INTO decks (subject_id, slug, name, description, default_kind, sort_order)
SELECT s.id, d.slug, d.name, d.description, d.kind, d.sort_order
FROM (VALUES
  ('toeic-vocab', 'TOEIC Vocabulary',
   'Workplace and business English for the TOEIC Listening & Reading test: offices, meetings, email, commerce, contracts, finance, travel, logistics and customer service. Common, practical words at B1-C1, in a professional but everyday register. Examples should sound like real workplace situations (short emails, announcements, conversations).',
   'vocab', 4),
  ('conversational-vocab', 'Everyday & forum vocab',
   'Casual, conversational English as actually used in daily life and online — Reddit, forums, comment replies, chat and texting. High-frequency everyday words, informal register, including common slang and internet usage where natural. NOT academic or exam-formal. Examples should read like a natural comment or message, not an essay.',
   'vocab', 5),
  ('conversational-reactions', 'Informal connectors & reactions',
   'Short natural phrases for replying and writing casually online: reacting (agreeing, disagreeing, being unsure), softening opinions, and connecting ideas — e.g. informal discourse markers like "to be fair", "that said", "I guess", "tbh", "no worries". For each, give the phrase, what it means and when to use it. Keep it casual (forum/comment tone), not formal essay connectors.',
   'concept', 6)
) AS d(slug, name, description, kind, sort_order)
JOIN subjects s ON s.slug = 'english'
ON CONFLICT (subject_id, slug) DO NOTHING;
