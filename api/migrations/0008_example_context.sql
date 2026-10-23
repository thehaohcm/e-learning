-- 0008_example_context: situation label of each example sentence (work, daily
-- life, news...), shown next to the sentence in "More examples".
ALTER TABLE card_examples ADD COLUMN IF NOT EXISTS context TEXT NOT NULL DEFAULT '';
