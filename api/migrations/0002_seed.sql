INSERT INTO users (email) VALUES ('demo@user.local') ON CONFLICT DO NOTHING;

INSERT INTO vocabs (word, meaning_en, meaning_vi, example_en, example_vi, topic)
VALUES
('sustainable', 'able to be maintained', 'bền vững', 'We need sustainable solutions.', 'Chúng ta cần các giải pháp bền vững.', 'Environment'),
('negotiate', 'to discuss to reach agreement', 'đàm phán', 'They negotiated a better deal.', 'Họ đã đàm phán một thỏa thuận tốt hơn.', 'Business');

INSERT INTO quiz_questions (vocab_id, prompt, choices, answer_idx)
VALUES
((SELECT id FROM vocabs WHERE word='sustainable'),
 'What is the closest meaning of "sustainable"?',
 '["maintainable","temporary","fragile","random"]',
 0),
((SELECT id FROM vocabs WHERE word='negotiate'),
 'Choose the best meaning of "negotiate".',
 '["argue","discuss to reach agreement","ignore","celebrate"]',
 1);

