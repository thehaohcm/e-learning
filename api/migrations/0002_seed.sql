-- 0002_seed: subjects, decks and a small PMP starter set.
-- PMP content here is written for this project (standard formulas and original
-- practice questions). Do not import real PMI exam questions: they are
-- confidential and copyrighted.

INSERT INTO subjects (slug, name, description, sort_order) VALUES
  ('english', 'English (TOEIC / IELTS)', 'Vocabulary, phrasal verbs, quizzes and writing practice', 1),
  ('pmp',     'PMI PMP',                 'Project management terms, formulas and practice questions', 2)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO decks (subject_id, slug, name, description, default_kind, sort_order)
SELECT s.id, d.slug, d.name, d.description, d.kind, d.sort_order
FROM (VALUES
  ('english', 'core-vocab',    'Core Vocabulary',        'TOEIC / IELTS words with everyday examples', 'vocab',   1),
  ('english', 'phrasal-verbs', 'Phrasal Verbs',          'Common phrasal verbs for work and daily life', 'phrasal', 2),
  ('pmp',     'terms',         'PMP Terms & Formulas',   'Key concepts and EVM formulas',               'concept', 1),
  ('pmp',     'practice',      'PMP Practice Questions', 'Original situational and calculation questions', 'mcq',  2)
) AS d(subject_slug, slug, name, description, kind, sort_order)
JOIN subjects s ON s.slug = d.subject_slug
ON CONFLICT (subject_id, slug) DO NOTHING;

-- ---------------------------------------------------------------- PMP concepts
INSERT INTO cards (deck_id, kind, front, payload, source)
SELECT d.id, 'concept', c.front, c.payload::jsonb, 'seed'
FROM (VALUES
  ('CPI (Cost Performance Index)',
   '{"back":"CPI = EV / AC","note_vi":"CPI < 1: vượt ngân sách; CPI > 1: dưới ngân sách.","example":"EV 40k, AC 50k → CPI 0.8"}'),
  ('SPI (Schedule Performance Index)',
   '{"back":"SPI = EV / PV","note_vi":"SPI < 1: chậm tiến độ; SPI > 1: nhanh hơn kế hoạch.","example":"EV 45k, PV 50k → SPI 0.9"}'),
  ('CV (Cost Variance)',
   '{"back":"CV = EV − AC","note_vi":"Âm: vượt ngân sách; dương: dưới ngân sách.","example":"EV 40k, AC 50k → CV −10k"}'),
  ('SV (Schedule Variance)',
   '{"back":"SV = EV − PV","note_vi":"Âm: chậm tiến độ; dương: vượt tiến độ.","example":"EV 45k, PV 50k → SV −5k"}'),
  ('EAC (typical variance)',
   '{"back":"EAC = BAC / CPI","note_vi":"Dự báo tổng chi phí khi hiệu suất chi phí hiện tại tiếp tục.","example":"BAC 100k, CPI 0.8 → EAC 125k"}'),
  ('ETC (Estimate to Complete)',
   '{"back":"ETC = EAC − AC","note_vi":"Chi phí còn cần để hoàn thành dự án."}'),
  ('VAC (Variance at Completion)',
   '{"back":"VAC = BAC − EAC","note_vi":"Âm: dự kiến vượt ngân sách khi kết thúc."}'),
  ('TCPI (to complete BAC)',
   '{"back":"TCPI = (BAC − EV) / (BAC − AC)","note_vi":"Hiệu suất chi phí cần đạt cho phần việc còn lại để không vượt BAC. > 1 nghĩa là khó hơn."}'),
  ('PERT (beta) estimate',
   '{"back":"(O + 4M + P) / 6","note_vi":"O: lạc quan, M: khả dĩ nhất, P: bi quan. Độ lệch chuẩn ≈ (P − O) / 6. Ước lượng tam giác: (O + M + P) / 3.","example":"O 4, M 6, P 14 → 7"}'),
  ('Communication channels',
   '{"back":"n(n − 1) / 2","note_vi":"n là số người liên quan (kể cả PM).","example":"8 người → 28 kênh"}'),
  ('Total float',
   '{"back":"Float = LS − ES = LF − EF","note_vi":"Thời gian một công việc có thể trễ mà không làm trễ ngày kết thúc dự án. Đường găng (critical path) thường có float = 0."}'),
  ('Project charter',
   '{"back":"Formally authorizes the project and gives the PM authority to use organizational resources.","note_vi":"Do sponsor (người tài trợ) ban hành, trước khi lập kế hoạch chi tiết."}'),
  ('Threat risk responses',
   '{"back":"Escalate, Avoid, Transfer, Mitigate, Accept","note_vi":"Với cơ hội (opportunity): Escalate, Exploit, Share, Enhance, Accept."}'),
  ('Servant leadership',
   '{"back":"Leading by serving the team: removing impediments, coaching and enabling self-organization.","note_vi":"Phong cách lãnh đạo hay gặp trong câu hỏi Agile."}')
) AS c(front, payload)
JOIN decks d ON d.slug = 'terms'
JOIN subjects s ON s.id = d.subject_id AND s.slug = 'pmp'
ON CONFLICT (deck_id, kind, lower(front)) DO NOTHING;

-- ---------------------------------------------------------------- PMP practice (MCQ)
-- payload: choices[], answer (0-based index), explanation
INSERT INTO cards (deck_id, kind, front, payload, source)
SELECT d.id, 'mcq', q.front, q.payload::jsonb, 'seed'
FROM (VALUES
  ('A project has an earned value of $40,000 and an actual cost of $50,000. What does this indicate?',
   '{"choices":["CPI is 0.8: the project is over budget","CPI is 1.25: the project is under budget","SPI is 0.8: the project is behind schedule","CV is +$10,000: the project is under budget"],"answer":0,"explanation":"CPI = EV / AC = 40,000 / 50,000 = 0.8. CPI below 1 means more is being spent than the value earned (over budget). Schedule cannot be judged without PV."}'),
  ('A team grows from 5 to 8 people. How many communication channels are added?',
   '{"choices":["3","13","18","28"],"answer":2,"explanation":"Channels = n(n − 1) / 2. 5 people: 10 channels; 8 people: 28 channels. Added: 28 − 10 = 18."}'),
  ('An activity has optimistic 4 days, most likely 6 days and pessimistic 14 days. What is the PERT (beta) estimate?',
   '{"choices":["6 days","7 days","8 days","8.5 days"],"answer":1,"explanation":"PERT = (O + 4M + P) / 6 = (4 + 24 + 14) / 6 = 7. 8 days is the triangular estimate (O + M + P) / 3."}'),
  ('On an agile team, who is primarily responsible for ordering the product backlog?',
   '{"choices":["Scrum master","Product owner","Project sponsor","The development team by vote"],"answer":1,"explanation":"The product owner owns and orders the backlog to maximize value. The scrum master facilitates and removes impediments."}'),
  ('During execution of a predictive project, a key stakeholder asks the team to add a new feature. What should the project manager do first?',
   '{"choices":["Add the feature if the team has spare capacity","Refuse because the scope is already baselined","Assess the impact and submit a change request through integrated change control","Escalate directly to the sponsor to decide"],"answer":2,"explanation":"In a predictive project, changes to the baseline go through integrated change control: analyse impact, then submit a change request for the CCB to decide. Adding it directly is gold plating / scope creep."}'),
  ('BAC is $100,000, EV is $60,000 and AC is $75,000. Assuming current cost performance continues, what is the EAC?',
   '{"choices":["$115,000","$125,000","$133,333","$140,000"],"answer":1,"explanation":"CPI = 60,000 / 75,000 = 0.8. EAC = BAC / CPI = 100,000 / 0.8 = 125,000."}')
) AS q(front, payload)
JOIN decks d ON d.slug = 'practice'
JOIN subjects s ON s.id = d.subject_id AND s.slug = 'pmp'
ON CONFLICT (deck_id, kind, lower(front)) DO NOTHING;
