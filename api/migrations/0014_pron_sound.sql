-- 0014_pron_sound: "Which sound?" questions. For each word, the vowel (letter
-- group) Vietnamese learners most often misread from the spelling, e.g. the
-- "u" of FLUC in fluctuate is /ʌ/ ("â"), not /u/.
-- NULL = not prepared yet, {} = nothing tricky in this word.
ALTER TABLE card_pronunciations ADD COLUMN IF NOT EXISTS sound JSONB;
