INSERT INTO users (email) VALUES ('demo@user.local') ON CONFLICT DO NOTHING;

INSERT INTO vocabs (word, meaning_en, meaning_vi, example_en, example_vi, topic)
VALUES
('sustainable', 'able to be maintained', 'bền vững', 'We need sustainable solutions.', 'Chúng ta cần các giải pháp bền vững.', 'Environment'),
('negotiate', 'to discuss to reach agreement', 'đàm phán', 'They negotiated a better deal.', 'Họ đã đàm phán một thỏa thuận tốt hơn.', 'Business')
ON CONFLICT DO NOTHING;

-- Insert quiz questions only if they don't exist
INSERT INTO quiz_questions (vocab_id, prompt, choices, answer_idx)
SELECT 
    v.id,
    'What is the closest meaning of "sustainable"?',
    '["maintainable","temporary","fragile","random"]',
    0
FROM vocabs v
WHERE v.word = 'sustainable'
AND NOT EXISTS (
    SELECT 1 FROM quiz_questions q 
    WHERE q.prompt = 'What is the closest meaning of "sustainable"?'
);

INSERT INTO quiz_questions (vocab_id, prompt, choices, answer_idx)
SELECT 
    v.id,
    'Choose the best meaning of "negotiate".',
    '["argue","discuss to reach agreement","ignore","celebrate"]',
    1
FROM vocabs v
WHERE v.word = 'negotiate'
AND NOT EXISTS (
    SELECT 1 FROM quiz_questions q 
    WHERE q.prompt = 'Choose the best meaning of "negotiate".'
);
