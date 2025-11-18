CREATE TABLE IF NOT EXISTS users (
  id BIGSERIAL PRIMARY KEY,
  email TEXT UNIQUE,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vocabs (
  id BIGSERIAL PRIMARY KEY,
  word TEXT,
  phonetics TEXT,
  meaning_en TEXT,
  meaning_vi TEXT,
  example_en TEXT,
  example_vi TEXT,
  audio_url TEXT,
  topic TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_vocabs (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id),
  vocab_id BIGINT REFERENCES vocabs(id),
  difficulty TEXT,
  last_result BOOLEAN,
  stability DOUBLE PRECISION DEFAULT 2.5,
  interval_days INT DEFAULT 1,
  next_review_at TIMESTAMP DEFAULT NOW(),
  history_json TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS quiz_questions (
  id BIGSERIAL PRIMARY KEY,
  vocab_id BIGINT REFERENCES vocabs(id),
  prompt TEXT,
  choices TEXT,
  answer_idx INT,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS quiz_attempts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id),
  question_id BIGINT REFERENCES quiz_questions(id),
  correct BOOLEAN,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS study_sessions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT REFERENCES users(id),
  topic TEXT,
  started_at TIMESTAMP DEFAULT NOW(),
  ended_at TIMESTAMP
);

