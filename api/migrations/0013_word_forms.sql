-- 0013_word_forms: TOEIC Part 5 style options for fill-in-the-blank sentences.
-- form_options = 3 wrong forms of the same word family for this sentence
-- ("fluctuation", "fluctuating", "fluctuant" when the answer is "fluctuated").
-- NULL = not prepared yet, [] = the word has no usable family (use other words).
ALTER TABLE card_examples ADD COLUMN IF NOT EXISTS form_options JSONB;
