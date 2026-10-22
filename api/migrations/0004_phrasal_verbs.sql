-- 0004_phrasal_verbs: starter phrasal verbs common in TOEIC and daily work English.
-- payload: verb, particle, separable, senses[]

INSERT INTO cards (deck_id, kind, front, payload, source)
SELECT d.id, 'phrasal', p.verb || ' ' || p.particle,
       jsonb_build_object(
         'verb', p.verb, 'particle', p.particle, 'separable', p.separable,
         'senses', jsonb_build_array(jsonb_build_object(
            'level', 1, 'meaning_en', p.meaning_en, 'meaning_vi', p.meaning_vi,
            'example_en', p.example_en, 'example_vi', p.example_vi))),
       'seed'
FROM (VALUES
  ('give',  'up',          true,  'stop doing or trying something',     'từ bỏ',               'She gave up smoking last year.',                  'Cô ấy đã bỏ thuốc lá vào năm ngoái.'),
  ('look',  'after',       false, 'take care of',                       'chăm sóc',            'Can you look after my dog this weekend?',         'Bạn trông con chó giúp mình cuối tuần này được không?'),
  ('put',   'off',         true,  'postpone',                           'trì hoãn, dời lại',   'They put off the meeting until Friday.',          'Họ dời cuộc họp sang thứ Sáu.'),
  ('carry', 'out',         true,  'perform or complete a task',         'thực hiện, tiến hành','The team carried out a customer survey.',         'Nhóm đã tiến hành một cuộc khảo sát khách hàng.'),
  ('come',  'up with',     false, 'think of an idea or plan',           'nghĩ ra',             'We need to come up with a better plan.',          'Chúng ta cần nghĩ ra một kế hoạch tốt hơn.'),
  ('figure','out',         true,  'understand or solve',                'hiểu ra, tìm ra',     'I can''t figure out how this printer works.',     'Tôi không hiểu cái máy in này hoạt động thế nào.'),
  ('run',   'out of',      false, 'have none left',                     'hết, cạn',            'We have run out of printer paper.',               'Chúng ta hết giấy in rồi.'),
  ('bring', 'up',          true,  'mention a topic',                    'đề cập, nêu ra',      'She brought up the budget issue in the meeting.', 'Cô ấy nêu vấn đề ngân sách trong cuộc họp.'),
  ('turn',  'down',        true,  'refuse an offer or request',         'từ chối',             'He turned down the job offer.',                   'Anh ấy đã từ chối lời mời làm việc.'),
  ('set',   'up',          true,  'establish or arrange',               'thiết lập, thành lập','They set up a new office in Da Nang.',            'Họ mở một văn phòng mới ở Đà Nẵng.'),
  ('take',  'over',        true,  'take control of something',          'tiếp quản',           'A larger company took over the startup.',         'Một công ty lớn hơn đã tiếp quản công ty khởi nghiệp đó.'),
  ('look',  'forward to',  false, 'feel happy about something coming',  'mong đợi',            'I look forward to hearing from you.',             'Tôi mong sớm nhận được phản hồi từ bạn.'),
  ('follow','up',          false, 'check on or continue something',     'theo dõi, liên hệ lại','I will follow up with the client tomorrow.',     'Ngày mai tôi sẽ liên hệ lại với khách hàng.'),
  ('fill',  'out',         true,  'complete a form',                    'điền (đơn, biểu mẫu)','Please fill out this application form.',          'Vui lòng điền vào mẫu đơn này.')
) AS p(verb, particle, separable, meaning_en, meaning_vi, example_en, example_vi)
JOIN decks d ON d.slug = 'phrasal-verbs'
JOIN subjects s ON s.id = d.subject_id AND s.slug = 'english'
ON CONFLICT (deck_id, kind, lower(front)) DO NOTHING;
