# AI-Powered E-Learning Platform – Architecture & Design Document

> **Role:** Solution Architect & Tech Lead  
> **Date:** 2026-06-10  
> **Stack:** Golang (Gin Framework) · PostgreSQL 16 · Redis · Vue 3 (Vite) · DeepSeek V4 Pro · Google Cloud Speech-to-Text · Browser Web Speech API (TTS)

---

## Table of Contents

1. [System Architecture Design](#1-system-architecture-design)
2. [Database Schema Design](#2-database-schema-design)
3. [Spaced Repetition Worker Logic](#3-spaced-repetition-worker-logic)
4. [Prompt Engineering Strategy](#4-prompt-engineering-strategy)
5. [Development Roadmap](#5-development-roadmap)

---

## 1. System Architecture Design

### 1.1 High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                                      CLIENT LAYER                                         │
│                                                                                          │
│   ┌──────────────────────────┐   ┌──────────────────────────┐   ┌──────────────────┐    │
│   │    Web App (Vue 3)       │   │    Mobile PWA             │   │  Admin Dashboard  │    │
│   │    Vite + TypeScript     │   │    (Responsive)           │   │  (Future)         │    │
│   │    • Flashcard UI        │   │    • Audio Recording      │   │  • Analytics      │    │
│   │    • Audio Capture       │   │    • Push Notifications   │   │  • Content Mgmt   │    │
│   └───────────┬──────────────┘   └───────────┬──────────────┘   └────────┬─────────┘    │
│               │                              │                            │              │
└───────────────┼──────────────────────────────┼────────────────────────────┼──────────────┘
                │                              │                            │
                └──────────────────────────────┼────────────────────────────┘
                                               │ HTTPS (TLS 1.3)
                                               ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                                    GATEWAY LAYER (Nginx)                                   │
│   • SSL Termination  • Rate Limiting (100 req/s per IP)  • Static Asset Cache (1y)        │
│   • Gzip/Brotli      • WebSocket Upgrade (for real-time alerts)                           │
└─────────────────────────────────────┬────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                                 GOLANG CORE SERVICES (:5173)                               │
│                                                                                           │
│  ┌─────────────────────────┐  ┌──────────────────────────┐  ┌─────────────────────────┐  │
│  │  API Server (Gin)       │  │  SRS Alert Worker Pool   │  │  AI Orchestrator         │  │
│  │                         │  │  (Background Goroutines) │  │  (Worker Pool Pattern)   │  │
│  │  • REST Endpoints       │  │                          │  │                          │  │
│  │  • JWT Auth Middleware  │  │  • Cron every 5 min      │  │  • Request Queue         │  │
│  │  • File Upload (Audio)  │  │  • Scan due_reviews      │  │  • Retry + Backoff       │  │
│  │  • Input Validation     │  │  • Push Notification     │  │  • Circuit Breaker       │  │
│  │  • Response Handler     │  │  • Reschedule NextReview │  │  • Response Cache (Redis) │  │
│  │  • CORS/Helmet          │  │  • APNs/FCM Integration  │  │  • Rate Limit Management │  │
│  └───────────┬─────────────┘  └───────────┬──────────────┘  └───────────┬─────────────┘  │
│              │                            │                             │                 │
│              └────────────────────────────┼─────────────────────────────┘                 │
│                                           │                                               │
│                    ┌──────────────────────┼──────────────────────┐                        │
│                    │                      ▼                      │                        │
│                    │          ┌─────────────────────┐           │                        │
│                    │          │   Redis (Cache+Queue)│           │                        │
│                    │          │                      │           │                        │
│                    │          │ • Session Store       │           │                        │
│                    │          │ • Vocab Cache          │           │                        │
│                    │          │ • AI Response Cache    │           │                        │
│                    │          │ • SRS Job Queue (List) │           │                        │
│                    │          │ • Pub/Sub: Alerts      │           │                        │
│                    │          │ • Rate Limit Counter   │           │                        │
│                    │          │ • User Presence        │           │                        │
│                    │          └─────────────────────┘           │                        │
│                    │                    │                        │                        │
│                    ▼                    ▼                        ▼                        │
│  ┌─────────────────────────────────────────────────────────────────────┐                 │
│  │                       PostgreSQL 16 (Primary)                        │                 │
│  │                                                                      │                 │
│  │  users │ vocabularies │ sentences │ user_vocab_progress               │                 │
│  │  writing_submissions │ quizzes │ quiz_attempts │ notification_log     │                 │
│  │  study_sessions │ daily_streaks │ topic_categories │ prompt_templates │                 │
│  └─────────────────────────────────────────────────────────────────────┘                 │
│                                                                                           │
│  ┌──────────────────────────────────────────────────────────────────────┐                │
│  │                    Object Storage (S3/MinIO - Local Dev)               │                │
│  │  • User audio recordings (.webm/.wav) • TTS audio caches (.mp3)       │                │
│  └──────────────────────────────────────────────────────────────────────┘                │
└──────────────────────────────────────────────────────────────────────────────────────────┘
                                               │
                                               │ HTTPS (TLS 1.3)
                                               ▼
┌──────────────────────────────────────────────────────────────────────────────────────────┐
│                                      AI SERVICES LAYER                                     │
│                                                                                           │
│  ┌─────────────────────────────────────┐  ┌────────────────────────────────────────────┐  │
│  │       DeepSeek V4 Pro API           │  │       Google Cloud Speech-to-Text          │  │
│  │                                     │  │                                            │  │
│  │  • DeepSeek-Chat (Fast Inference):  │  │  • Cloud Speech-to-Text (STT):             │  │
│  │    - 3-level vocabulary generation  │  │    - Streaming recognition                │  │
│  │    - 3-level example sentences      │  │    - Phoneme-level pronunciation scoring   │  │
│  │    - Search/semantic reference      │  │    - Confidence score per word             │  │
│  │    - Quiz generation                │  │    - Word-level alignment with reference   │  │
│  │                                     │  │                                            │  │
│  │  • DeepSeek-Reasoning (Deep Think): │  │                                            │  │
│  │    - Writing essay evaluation       │  │                                            │  │
│  │    - Grammar error analysis         │  │                                            │  │
│  │    - 3-level rewriting suggestions  │  │                                            │  │
│  │    - Personalized learning analysis │  │                                            │  │
│  │    - Complex exam generation        │  │                                            │  │
│  └─────────────────────────────────────┘  └────────────────────────────────────────────┘  │
│                                                                                           │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐  │
│  │                          BROWSER-BASED Text-to-Speech (FREE)                          │  │
│  │  • Web Speech API (window.speechSynthesis) - built into all modern browsers          │  │
│  │  • en-US: 5+ voices (Google US English, Microsoft David/Zira, etc.)                  │  │
│  │  • vi-VN: 2+ voices (Google tiếng Việt, Microsoft An)                                │  │
│  │  • No API cost, no latency, works offline with cached voices                         │  │
│  └─────────────────────────────────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Core Request Flows

#### Flow A: Generate Vocabulary with 3 Levels

```
  User           API Server              AI Orchestrator           Redis         DeepSeek-Chat
  ────           ──────────              ───────────────           ─────         ─────────────
   │                  │                        │                      │               │
   │ POST /api/v1/    │                        │                      │               │
   │ vocab/generate   │                        │                      │               │
   │ {word:"agile"}   │                        │                      │               │
   │─────────────────>│                        │                      │               │
   │                  │                        │                      │               │
   │                  │ 1. Validate JWT        │                      │               │
   │                  │ 2. Check Redis cache   │                      │               │
   │                  │─────────────────────────────────────────────>│               │
   │                  │<─────────────────────────────────────────────│               │
   │                  │   MISS                 │                      │               │
   │                  │                        │                      │               │
   │                  │ 3. Async call to       │                      │               │
   │                  │    AI Orchestrator     │                      │               │
   │                  │───────────────────────>│                      │               │
   │                  │                        │                      │               │
   │                  │                        │ 4. Build Prompt      │               │
   │                  │                        │    (System + User)   │               │
   │                  │                        │                      │               │
   │                  │                        │ 5. POST /chat        │               │
   │                  │                        │──────────────────────────────────────────>│
   │                  │                        │                      │               │
   │                  │                        │ 6. JSON Response     │               │
   │                  │                        │    {word, phonetics, │               │
   │                  │                        │     levels: [        │               │
   │                  │                        │      {level:1,       │               │
   │                  │                        │       meaning_vi,    │               │
   │                  │                        │       meaning_en,    │               │
   │                  │                        │       example_en,    │               │
   │                  │                        │       example_vi,    │               │
   │                  │                        │       notes},        │               │
   │                  │                        │      {level:2,...},  │               │
   │                  │                        │      {level:3,...}   │               │
   │                  │                        │    ]}                │               │
   │                  │                        │<──────────────────────────────────────────│
   │                  │                        │                      │               │
   │                  │                        │ 7. Parse + Validate  │               │
   │                  │                        │    Response          │               │
   │                  │                        │                      │               │
   │                  │ 8. DB Insert +         │                      │               │
   │                  │    Cache Set           │                      │               │
   │                  │─────────────────────────────────────────────>│               │
   │                  │   SETEX vocab:agile 1h  │                      │               │
   │                  │<─────────────────────────────────────────────│               │
   │                  │                        │                      │               │
   │  200 OK          │                        │                      │               │
   │<─────────────────│                        │                      │               │
```

#### Flow B: Spaced Repetition Alert Cycle

```
  Cron (every 5 min)      SRS Worker           PostgreSQL           Push/FCM/Email
  ─────────────────       ──────────           ──────────           ──────────────
         │                     │                    │                      │
         │ Trigger             │                    │                      │
         │────────────────────>│                    │                      │
         │                     │                    │                      │
         │                     │ 1. Acquire Redis   │                      │
         │                     │    distributed lock│                      │
         │                     │    SETNX srs:lock  │                      │
         │                     │                    │                      │
         │                     │ 2. SELECT uvp.*, v.*, u.*            │           │
         │                     │    FROM user_vocab_progress uvp    │           │
         │                     │    JOIN vocabularies v ON ...     │           │
         │                     │    JOIN users u ON ...            │           │
         │                     │    WHERE uvp.next_review_at <= NOW()         │
         │                     │      AND uvp.is_active = true     │           │
         │                     │      AND NOT EXISTS (             │           │
         │                     │        SELECT 1 FROM notification_log nl     │
         │                     │        WHERE nl.user_id = uvp.user_id        │
         │                     │          AND nl.vocab_id = uvp.vocab_id      │
         │                     │          AND nl.created_at > NOW()-INTERVAL '4 hours'
         │                     │      )                             │           │
         │                     │    LIMIT 1000                     │           │
         │                     │─────────────────────────────────>│           │
         │                     │<─── Result Set                   │           │
         │                     │                    │                      │
         │                     │ 3. For each row, dispatch         │           │
         │                     │    notification (goroutine pool)  │           │
         │                     │──────────────────────────────────────────────────>│
         │                     │                    │                      │
         │                     │ 4. Log notification │                      │
         │                     │─────────────────────────────────>│           │
         │                     │                    │                      │
         │                     │ 5. Unlock srs:lock  │                      │
         │                     │                    │                      │
```

#### Flow C: Pronunciation Assessment

```
  User           API Server           AI Orchestrator       S3/MinIO       Google STT
  ────           ──────────           ───────────────       ────────       ──────────
   │                  │                      │                  │               │
   │ POST /api/v1/    │                      │                  │               │
   │ speech/assess    │                      │                  │               │
   │ (multipart:      │                      │                  │               │
   │  file=audio.webm │                      │                  │               │
   │  vocab_id=42     │                      │                  │               │
   │  level=1)        │                      │                  │               │
   │─────────────────>│                      │                  │               │
   │                  │                      │                  │               │
   │                  │ 1. Validate file     │                  │               │
   │                  │    (size < 5MB,      │                  │               │
   │                  │     format: webm/wav)│                  │               │
   │                  │                      │                  │               │
   │                  │ 2. Upload to S3      │                  │               │
   │                  │─────────────────────────────────────>│               │
   │                  │<─── object_key       │                  │               │
   │                  │                      │                  │               │
   │                  │ 3. Load reference    │                  │               │
   │                  │    sentence from DB  │                  │               │
   │                  │                      │                  │               │
   │                  │ 4. Call AI Orchestrator                 │               │
   │                  │─────────────────────>│                  │               │
   │                  │                      │                  │               │
   │                  │                      │ 5. STT Request   │               │
   │                  │                      │──────────────────────────────────>│
   │                  │                      │                  │               │
   │                  │                      │ 6. STT Response  │               │
   │                  │                      │    {transcript:  │               │
   │                  │                      │     "I ...",     │               │
   │                  │                      │     words: [     │               │
   │                  │                      │      {word: "I", │               │
   │                  │                      │       confidence:│               │
   │                  │                      │       0.98,      │               │
   │                  │                      │       start_time,│               │
   │                  │                      │       end_time}  │               │
   │                  │                      │    ]}            │               │
   │                  │                      │<──────────────────────────────────│
   │                  │                      │                  │               │
   │                  │                      │ 7. Compute score │               │
   │                  │                      │    (Needleman-   │               │
   │                  │                      │     Wunsch align │               │
   │                  │                      │     reference vs │               │
   │                  │                      │     transcript)   │               │
   │                  │                      │                  │               │
   │                  │ 8. Return assessment │                  │               │
   │                  │<─────────────────────│                  │               │
   │                  │                      │                  │               │
   │  200 OK          │                      │                  │               │
   │  {overall_score: │                      │                  │               │
   │   85,            │                      │                  │               │
   │   word_scores:   │                      │                  │               │
   │   [{word: "I",   │                      │                  │               │
   │     score: 98},  │                      │                  │               │
   │    {word: "am",  │                      │                  │               │
   │     score: 92}], │                      │                  │               │
   │   feedback: "..."}│                     │                  │               │
   │<─────────────────│                      │                  │               │
```

### 1.3 Frontend Architecture

```
┌────────────────────────────────────────────────────────────┐
│                     Vue 3 + Vite + TypeScript               │
│                                                             │
│  src/                                                       │
│  ├── components/                                            │
│  │   ├── flashcard/                                         │
│  │   │   ├── FlashCard.vue          (Animated flip card)    │
│  │   │   └── LevelSlider.vue        (Level 1/2/3 toggle)    │
│  │   ├── audio/                                              │
│  │   │   ├── AudioRecorder.vue      (MediaRecorder API)     │
│  │   │   ├── AudioPlayer.vue        (TTS Playback)          │
│  │   │   └── WaveformVisualizer.vue                         │
│  │   ├── quiz/                                               │
│  │   │   ├── QuizCard.vue                                   │
│  │   │   └── QuizResults.vue                                │
│  │   ├── writing/                                            │
│  │   │   ├── WritingEditor.vue      (Rich text editor)      │
│  │   │   └── WritingFeedback.vue    (3-level rewrite view)  │
│  │   └── common/                                             │
│  │       ├── SpacedRepAlert.vue     (Alert modal/toast)     │
│  │       └── PronunciationChart.vue                         │
│  ├── composables/                                            │
│  │   ├── useAudioRecorder.ts        (MediaRecorder wrapper) │
│  │   ├── useSRS.ts                  (SRS state management)  │
│  │   └── useAISearch.ts             (Debounced search)      │
│  ├── stores/ (Pinia)                                         │
│  │   ├── vocabStore.ts                                       │
│  │   ├── userStore.ts                                        │
│  │   └── quizStore.ts                                        │
│  └── pages/                                                 │
│      ├── Dashboard.vue                                       │
│      ├── StudySession.vue                                    │
│      ├── WritingLab.vue                                      │
│      └── QuizRoom.vue                                        │
└────────────────────────────────────────────────────────────┘
```

---

## 2. Database Schema Design

### 2.1 Entity Relationship Diagram (Core)

```
users ──< user_vocab_progress >── vocabularies ──< sentences
  │                                      │
  │                                      ├──< quiz_questions
  │                                      │
  │                                      └──< vocab_tag_mappings >── vocab_tags
  │
  ├──< writing_submissions
  │        └──< writing_sentences
  │
  ├──< user_streaks
  │
  ├──< quiz_attempts >── quiz_questions
  │
  ├──< notification_log
  │
  └──< user_preferences

topic_categories ──< vocabularies
```

### 2.2 Complete DDL SQL

```sql
-- ============================================================================
-- MIGRATION 0001: CORE SCHEMA
-- ============================================================================

-- 1. USERS TABLE
CREATE TABLE IF NOT EXISTS users (
    id              BIGSERIAL PRIMARY KEY,
    email           TEXT UNIQUE NOT NULL,
    display_name    TEXT,
    firebase_uid    TEXT UNIQUE,           -- Firebase Auth UID (for mobile/web auth)
    avatar_url      TEXT,
    timezone        TEXT DEFAULT 'UTC',    -- For scheduling alerts in user's timezone
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);

-- 2. TOPIC CATEGORIES (e.g., "TOEIC", "IELTS", "Office English", "Movies")
CREATE TABLE IF NOT EXISTS topic_categories (
    id              SMALLSERIAL PRIMARY KEY,
    slug            TEXT UNIQUE NOT NULL,  -- e.g., "toeic-business"
    name_vi         TEXT NOT NULL,         -- "TOEIC - Tiếng Anh Thương Mại"
    name_en         TEXT NOT NULL,         -- "TOEIC - Business English"
    description     TEXT,
    parent_id       INT REFERENCES topic_categories(id),
    sort_order      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 3. VOCABULARIES (Core word/phrase table)
--    Enable pg_trgm extension for fuzzy text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS vocabularies (
    id              BIGSERIAL PRIMARY KEY,
    word            TEXT NOT NULL,          -- The English word/phrase
    phonetics       TEXT,                   -- /ˈædʒ.aɪl/
    pos             TEXT,                   -- Part of speech: noun, verb, adj, etc.
    topic_id        INT REFERENCES topic_categories(id),
    audio_url       TEXT,                   -- S3 URL for TTS pronunciation
    image_url       TEXT,                   -- Optional visual aid
    source          TEXT,                   -- Source: "deepseek-chat", "manual", "import"
    meta_json       JSONB DEFAULT '{}',     -- Flexible metadata
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_vocab_word ON vocabularies USING gin (word gin_trgm_ops);
CREATE INDEX idx_vocab_topic ON vocabularies(topic_id);
CREATE INDEX idx_vocab_meta ON vocabularies USING gin (meta_json);

-- 4. SENTENCES (3 Levels per vocabulary entry)
--    KEY DESIGN: One vocab maps to 3 sentences via 'level' column (1=Simple, 2=Normal, 3=Complex)
CREATE TABLE IF NOT EXISTS sentences (
    id              BIGSERIAL PRIMARY KEY,
    vocab_id        BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
    level           SMALLINT NOT NULL CHECK (level IN (1, 2, 3)),
    meaning_en      TEXT NOT NULL,          -- English meaning/definition at this level
    meaning_vi      TEXT NOT NULL,          -- Vietnamese translation
    example_en      TEXT NOT NULL,          -- Example sentence in English
    example_vi      TEXT NOT NULL,          -- Example sentence translated to Vietnamese
    usage_note      TEXT,                   -- Usage note (phrasal verbs, idioms, register)
    grammar_note    TEXT,                   -- Grammar explanation (e.g., subjunctive, passive)
    audio_url       TEXT,                   -- S3 URL for TTS of example_en
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(vocab_id, level)                -- One sentence per level per vocab
);
CREATE INDEX idx_sentences_vocab ON sentences(vocab_id);

-- 5. VOCABULARY TAGS (for flexible categorization)
CREATE TABLE IF NOT EXISTS vocab_tags (
    id              SERIAL PRIMARY KEY,
    name            TEXT UNIQUE NOT NULL,   -- e.g., "phrasal-verb", "business", "academic"
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS vocab_tag_mappings (
    vocab_id        BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,
    tag_id          INT NOT NULL REFERENCES vocab_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (vocab_id, tag_id)
);

-- ============================================================================
-- MIGRATION 0002: USER PROGRESS & SPACED REPETITION
-- ============================================================================

-- 6. USER_VOCAB_PROGRESS (SRS State per user per vocab)
--    This is the core table for the Spaced Repetition engine.
CREATE TABLE IF NOT EXISTS user_vocab_progress (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vocab_id        BIGINT NOT NULL REFERENCES vocabularies(id) ON DELETE CASCADE,

    -- Spaced Repetition State (FSRS-inspired algorithm)
    state           SMALLINT DEFAULT 0 NOT NULL,  -- 0=New, 1=Learning, 2=Review, 3=Relearning
    stability       DOUBLE PRECISION DEFAULT 2.5,  -- Memory stability (days)
    difficulty      DOUBLE PRECISION DEFAULT 0.3,  -- Item difficulty (0.0 to 1.0)
    elapsed_days    INT DEFAULT 0,                 -- Days since last review
    scheduled_days  INT DEFAULT 0,                 -- Next interval in days
    reps            INT DEFAULT 0,                 -- Total number of reviews
    lapses          INT DEFAULT 0,                 -- Times forgotten (rating < 3)
    last_review_at  TIMESTAMPTZ,
    next_review_at  TIMESTAMPTZ DEFAULT NOW(),

    -- Learning Metadata
    level_focus     SMALLINT DEFAULT 1 CHECK (level_focus IN (1, 2, 3)),
    is_active       BOOLEAN DEFAULT true,         -- User can pause learning a word
    mastery_score   DOUBLE PRECISION DEFAULT 0.0,  -- 0-100 computed mastery
    history_json    JSONB DEFAULT '[]',            -- Array of review events: [{date, rating, level, duration_ms}]

    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, vocab_id)
);
CREATE INDEX idx_uvp_next_review ON user_vocab_progress(user_id, next_review_at) WHERE is_active = true;
CREATE INDEX idx_uvp_user_state ON user_vocab_progress(user_id, state);
CREATE INDEX idx_uvp_due_count ON user_vocab_progress(next_review_at) WHERE next_review_at <= NOW() AND is_active = true;

-- 7. REVIEW_SESSIONS (Each time user reviews a batch of vocabulary)
CREATE TABLE IF NOT EXISTS review_sessions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vocab_count     INT NOT NULL,           -- Number of words reviewed
    duration_ms     BIGINT,                 -- Session duration in milliseconds
    avg_rating      DOUBLE PRECISION,       -- Average self-rating across the session
    session_type    TEXT DEFAULT 'full',     -- 'full', 'quick', 'quiz'
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_rs_user ON review_sessions(user_id, created_at DESC);

-- 8. USER_STREAKS (Daily learning streaks)
CREATE TABLE IF NOT EXISTS user_streaks (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    streak_date     DATE NOT NULL,
    vocab_reviewed  INT DEFAULT 0,
    writing_done    BOOLEAN DEFAULT false,
    quiz_done       BOOLEAN DEFAULT false,
    created_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(user_id, streak_date)
);
CREATE INDEX idx_streak_user_date ON user_streaks(user_id, streak_date DESC);

-- ============================================================================
-- MIGRATION 0003: WRITING SUBMISSIONS
-- ============================================================================

-- 9. WRITING_SUBMISSIONS
CREATE TABLE IF NOT EXISTS writing_submissions (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    topic_id        INT REFERENCES topic_categories(id),
    prompt          TEXT NOT NULL,          -- The writing prompt/topic
    original_text   TEXT NOT NULL,          -- User's original writing
    word_count      INT,
    required_vocabs JSONB,                  -- [{vocab_id, word}] required words user must use
    used_vocabs     JSONB,                  -- [{vocab_id, word}] words user actually used

    -- AI Evaluation Results
    overall_score   DOUBLE PRECISION,       -- 0-100 composite score
    grammar_score   DOUBLE PRECISION,       -- Grammar accuracy score
    vocab_score     DOUBLE PRECISION,       -- Vocabulary usage score
    coherence_score DOUBLE PRECISION,       -- Coherence & cohesion score
    ai_feedback_json JSONB DEFAULT '{}',    -- Full AI response: errors, suggestions, rewrites
    ai_model_used   TEXT,                   -- "deepseek-reasoning" or "gemini-2.5-pro"
    ai_tokens_used  INT,

    status          TEXT DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed'
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);
CREATE INDEX idx_ws_user ON writing_submissions(user_id, created_at DESC);
CREATE INDEX idx_ws_status ON writing_submissions(status) WHERE status IN ('pending', 'processing');

-- 10. WRITING_ERRORS (Normalized grammar/vocab errors detected by AI)
CREATE TABLE IF NOT EXISTS writing_errors (
    id              BIGSERIAL PRIMARY KEY,
    submission_id   BIGINT NOT NULL REFERENCES writing_submissions(id) ON DELETE CASCADE,
    error_type      TEXT NOT NULL,          -- 'grammar', 'spelling', 'word_choice', 'collocation', 'punctuation'
    error_span      TEXT,                   -- Original erroneous text
    correction      TEXT,                   -- Corrected text
    explanation     TEXT,                   -- AI's explanation
    severity        SMALLINT DEFAULT 1,     -- 1=Minor, 2=Moderate, 3=Critical
    position_start  INT,                    -- Character offset in original_text
    position_end    INT,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 11. WRITING_REWRITES (3-Level rewrite suggestions for a submission)
CREATE TABLE IF NOT EXISTS writing_rewrites (
    id              BIGSERIAL PRIMARY KEY,
    submission_id   BIGINT NOT NULL REFERENCES writing_submissions(id) ON DELETE CASCADE,
    level           SMALLINT NOT NULL CHECK (level IN (1, 2, 3)),
    rewritten_text  TEXT NOT NULL,
    explanation     TEXT,                   -- Why this version is level-appropriate
    created_at      TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(submission_id, level)
);

-- ============================================================================
-- MIGRATION 0004: QUIZZES
-- ============================================================================

-- 12. QUIZZES
CREATE TABLE IF NOT EXISTS quizzes (
    id              BIGSERIAL PRIMARY KEY,
    title           TEXT,
    description     TEXT,
    quiz_type       TEXT NOT NULL DEFAULT 'vocab', -- 'vocab', 'toeic', 'ielts', 'toefl', 'custom'
    topic_id        INT REFERENCES topic_categories(id),
    difficulty      TEXT DEFAULT 'medium',   -- 'easy', 'medium', 'hard'
    question_count  INT DEFAULT 10,
    time_limit_min  INT,                    -- Time limit in minutes (NULL = no limit)
    is_ai_generated BOOLEAN DEFAULT true,
    ai_prompt_used  TEXT,                   -- The prompt sent to AI to generate this quiz
    created_by      BIGINT REFERENCES users(id), -- NULL if system-generated
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

-- 13. QUIZ_QUESTIONS
CREATE TABLE IF NOT EXISTS quiz_questions (
    id              BIGSERIAL PRIMARY KEY,
    quiz_id         BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    vocab_id        BIGINT REFERENCES vocabularies(id),
    question_type   TEXT NOT NULL DEFAULT 'multiple_choice',
    prompt          TEXT NOT NULL,
    choices         JSONB NOT NULL,         -- [{text: "...", is_correct: true/false}]
    explanation     TEXT,                   -- Explanation shown after answering
    sort_order      INT DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_qq_quiz ON quiz_questions(quiz_id, sort_order);

-- 14. QUIZ_ATTEMPTS (User takes a quiz)
CREATE TABLE IF NOT EXISTS quiz_attempts (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    quiz_id         BIGINT NOT NULL REFERENCES quizzes(id),
    score           INT,                    -- Raw score (e.g., 8/10)
    total_questions INT,
    duration_ms     BIGINT,
    status          TEXT DEFAULT 'started',  -- 'started', 'completed', 'abandoned'
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);
CREATE INDEX idx_qa_user_quiz ON quiz_attempts(user_id, quiz_id);

-- 15. QUIZ_ANSWERS (Individual answers within an attempt)
CREATE TABLE IF NOT EXISTS quiz_answers (
    id              BIGSERIAL PRIMARY KEY,
    attempt_id      BIGINT NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id     BIGINT NOT NULL REFERENCES quiz_questions(id),
    selected_idx    INT,                    -- Index of user's selected choice (NULL if skipped)
    is_correct      BOOLEAN,
    time_spent_ms   INT,                    -- Time spent on this question
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_qanswer_attempt ON quiz_answers(attempt_id);

-- ============================================================================
-- MIGRATION 0005: NOTIFICATIONS & ALERTS
-- ============================================================================

-- 16. NOTIFICATION_LOG (Audit trail for all SRS alerts sent)
CREATE TABLE IF NOT EXISTS notification_log (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vocab_id        BIGINT NOT NULL REFERENCES vocabularies(id),
    level           SMALLINT,              -- Which level was sent
    channel         TEXT DEFAULT 'push',    -- 'push', 'email', 'sms', 'app'
    status          TEXT DEFAULT 'sent',    -- 'queued', 'sent', 'failed', 'delivered'
    error_message   TEXT,                  -- If failed
    created_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_nl_user_vocab_time ON notification_log(user_id, vocab_id, created_at DESC);

-- 17. USER_PREFERENCES (Notification & learning preferences)
CREATE TABLE IF NOT EXISTS user_preferences (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    daily_vocab_goal INT DEFAULT 10,       -- How many new words per day
    reminder_enabled BOOLEAN DEFAULT true,
    reminder_times  JSONB DEFAULT '["09:00", "13:00", "19:00"]',  -- Preferred alert times
    quiz_difficulty TEXT DEFAULT 'medium',
    tts_speed       DOUBLE PRECISION DEFAULT 1.0,  -- TTS playback speed
    ui_language     TEXT DEFAULT 'vi',     -- 'vi', 'en'
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- MIGRATION 0006: AI PROMPT MANAGEMENT
-- ============================================================================

-- 18. PROMPT_TEMPLATES (Store and version-control AI prompts)
CREATE TABLE IF NOT EXISTS prompt_templates (
    id              SERIAL PRIMARY KEY,
    name            TEXT UNIQUE NOT NULL,   -- 'vocab-generation-v1'
    description     TEXT,
    system_prompt   TEXT NOT NULL,
    user_prompt_template TEXT,             -- Template with {{placeholders}}
    ai_model        TEXT NOT NULL,          -- 'deepseek-chat', 'deepseek-reasoning', 'gemini'
    max_tokens      INT DEFAULT 2048,
    temperature     DOUBLE PRECISION DEFAULT 0.7,
    version         INT DEFAULT 1,
    is_active       BOOLEAN DEFAULT true,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- HELPFUL VIEWS
-- ============================================================================

-- View: Dashboard summary per user
CREATE OR REPLACE VIEW user_dashboard AS
SELECT
    u.id AS user_id,
    u.email,
    COUNT(uvp.id) FILTER (WHERE uvp.is_active = true) AS active_vocabs,
    COUNT(uvp.id) FILTER (WHERE uvp.mastery_score >= 80) AS mastered_vocabs,
    COUNT(uvp.id) FILTER (WHERE uvp.next_review_at <= NOW()) AS due_reviews,
    (SELECT COUNT(*) FROM user_streaks us
     WHERE us.user_id = u.id
       AND us.streak_date >= CURRENT_DATE - INTERVAL '7 days') AS streak_7d,
    (SELECT MAX(ws.overall_score) FROM writing_submissions ws
     WHERE ws.user_id = u.id
       AND ws.created_at >= NOW() - INTERVAL '7 days') AS best_writing_score_7d,
    (SELECT MAX(qa.score) FROM quiz_attempts qa
     WHERE qa.user_id = u.id
       AND qa.created_at >= NOW() - INTERVAL '7 days') AS best_quiz_score_7d
FROM users u
LEFT JOIN user_vocab_progress uvp ON uvp.user_id = u.id;
```

### 2.3 Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **3-level sentences in separate `sentences` table with `level` column** | Allows 1 vocab to have exactly 3 rows (one per level). Queries can filter by `level`. Extensible - if we add Level 4 later, just add a constraint. |
| **`user_vocab_progress` stores full SRS state** | Uses FSRS-like parameters (stability, difficulty, elapsed_days) for more accurate scheduling than SM-2. `history_json` keeps full review log for algorithm tuning. |
| **`writing_submissions` uses JSONB for AI feedback** | AI responses vary in structure. JSONB allows flexible storage while still being queryable (e.g., `ai_feedback_json->>'overall_score'`). |
| **Normalized `writing_errors` and `writing_rewrites`** | Enables analytics: "what errors do users make most?" and quick lookup of rewrites without parsing JSON. |
| **`prompt_templates` table** | Production-ready approach to prompt management - A/B test prompts, rollback, version control. |
| **GIN trigram index on `vocabularies.word`** | Enables fast fuzzy search (`ILIKE '%agil%'`) for the AI Search feature. |
| **Partial index on `user_vocab_progress.next_review_at`** | The SRS worker query only scans due items - this index makes it instant. |
| **`notification_log` deduplication** | Prevents sending the same alert twice within 4 hours via `NOT EXISTS` subquery. |

---

## 3. Spaced Repetition Worker Logic

### 3.1 Algorithm: FSRS (Free Spaced Repetition Scheduler)

The system uses **FSRS** (a modern, open-source algorithm that outperforms SM-2 and Anki's default algorithm) ported to Golang.

### 3.2 FSRS Core Formulas

```
FSRS computes the probability of recall (R) for a given item:

  R(t, S) = (1 + (t / (9 * S)) ^ -1) ^ -1

Where:
  t = elapsed days since last review
  S = stability (memory strength in days)

After a review with rating G (1=forgot, 2=hard, 3=good, 4=easy):

  New Difficulty:
    D' = D + w4 * (G - 3) * (1 / 17)    # D clamped to [0, 1]

  New Stability (after success, G >= 3):
    S' = S * (1 + exp(w8) * (11 - D) * S^(-w9) * (exp((1 - R) * w10) - 1))

  New Stability (after failure, G < 3):
    S' = S * w11 * D^(-w12) * ((1 - R)^w13 - 1)

  Next interval:
    I' = S' * ln(0.9) / ln(0.5)   # Interval to reach 90% retention
```

*Note: w4..w13 are trainable parameters. For MVP, we use the default FSRS-5 weights.*

### 3.3 Golang SRS Worker Architecture

```go
// File: api/internal/srs/fsrs.go
package srs

import (
    "math"
    "time"
)

// FSRS-5 default weights (pre-trained on Anki dataset)
var DefaultWeights = []float64{
    0.4872, 1.4003, 3.7145, 13.8206, 5.1618, 1.2298,
    0.8975, 0.031, 1.6474, 0.1367, 1.0461, 0.6461,
    0.0411, 1.0295, 0.1, 0.9906, 0.9714, 0.9561, 1.1687,
}

type FSRSParams struct {
    Stability   float64
    Difficulty  float64
    ElapsedDays int
    Reps        int
    Lapses      int
    State       int // 0=New, 1=Learning, 2=Review, 3=Relearning
}

type FSRSScheduleResult struct {
    NextStability   float64
    NextDifficulty  float64
    NextInterval    int
    NextReviewAt    time.Time
    NextState       int
    Retrievability  float64
}

// retrievability computes R(t, S)
func retrievability(elapsedDays int, stability float64) float64 {
    if elapsedDays <= 0 {
        return 1.0
    }
    t := float64(elapsedDays)
    return math.Pow(1+t/(9*stability), -1)
}

// NextInterval calculates optimal interval for given stability
func NextInterval(stability float64, retentionTarget float64) int {
    if retentionTarget <= 0 || retentionTarget >= 1 {
        retentionTarget = 0.9
    }
    return int(math.Ceil(stability * math.Log(retentionTarget) / math.Log(0.5)))
}

// Schedule computes the next SRS schedule given a rating (1=forgot, 2=hard, 3=good, 4=easy)
func Schedule(p FSRSParams, rating int) FSRSScheduleResult {
    w := DefaultWeights
    result := FSRSScheduleResult{}

    R := retrievability(p.ElapsedDays, p.Stability)

    // Update difficulty
    newD := p.Difficulty + w[4]*(float64(rating-3))*(1.0/17.0)
    newD = math.Max(0.0, math.Min(1.0, newD))
    newD = w[5]*p.Difficulty + (1-w[5])*newD

    var newS float64
    if rating >= 3 {
        factor := 1 + math.Exp(w[8])*(11-newD)*math.Pow(p.Stability, -w[9])*
            (math.Exp((1-R)*w[10])-1)
        newS = p.Stability * factor
        newS = w[6]*p.Stability + (1-w[6])*newS
    } else {
        factor := w[11]*math.Pow(newD, -w[12])*(math.Pow(1-R, w[13])-1)
        newS = math.Min(p.Stability, p.Stability*factor)
        newS = w[7]*p.Stability + (1-w[7])*newS
    }
    newS = math.Max(0.01, newS)

    result.NextStability = newS
    result.NextDifficulty = newD
    result.NextInterval = NextInterval(newS, 0.9)
    result.NextReviewAt = time.Now().Add(time.Duration(result.NextInterval)*24*time.Hour)
    result.Retrievability = retrievability(result.NextInterval, newS)

    if rating < 3 {
        result.NextState = 3
    } else if p.State == 0 || p.State == 1 {
        if result.NextInterval >= 7 {
            result.NextState = 2
        } else {
            result.NextState = 1
        }
    } else {
        result.NextState = 2
    }

    return result
}
```

### 3.4 SRS Alert Worker (Goroutine Pool + Redis Lock)

```go
// File: api/internal/worker/srs_alert_worker.go
package worker

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/go-redis/redis/v8"
    "gorm.io/gorm"
)

type DueReview struct {
    UserID     int64
    VocabID    int64
    Word       string
    Level      int
    ExampleEn  string
    ExampleVi  string
    AudioURL   string
}

type SRSAlertWorker struct {
    db       *gorm.DB
    redis    *redis.Client
    notifier NotificationService
    workers  int
    batchSize int
    interval  time.Duration
}

func NewSRSAlertWorker(db *gorm.DB, rds *redis.Client, n NotificationService) *SRSAlertWorker {
    return &SRSAlertWorker{
        db:        db,
        redis:     rds,
        notifier:  n,
        workers:   4,
        batchSize: 500,
        interval:  5 * time.Minute,
    }
}

func (w *SRSAlertWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()
    log.Println("[SRS Worker] Started")

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            w.processDueReviews(ctx)
        }
    }
}

func (w *SRSAlertWorker) processDueReviews(ctx context.Context) {
    // 1. Redis distributed lock prevents duplicate processing across instances
    lockKey := "srs:alert:lock"
    locked, _ := w.redis.SetNX(ctx, lockKey, time.Now().UnixNano(), 4*time.Minute).Result()
    if !locked {
        return
    }
    defer w.redis.Del(ctx, lockKey)

    // 2. Query due reviews with deduplication (no re-notify within 4 hours)
    var dueReviews []DueReview
    err := w.db.Raw(`
        SELECT uvp.user_id, uvp.vocab_id, v.word, uvp.level_focus AS level,
               s.example_en, s.example_vi, s.audio_url
        FROM user_vocab_progress uvp
        JOIN vocabularies v ON v.id = uvp.vocab_id
        JOIN sentences s ON s.vocab_id = v.id AND s.level = uvp.level_focus
        JOIN users u ON u.id = uvp.user_id
        JOIN user_preferences up ON up.user_id = u.id
        WHERE uvp.next_review_at <= NOW()
          AND uvp.is_active = true
          AND up.reminder_enabled = true
          AND NOT EXISTS (
              SELECT 1 FROM notification_log nl
              WHERE nl.user_id = uvp.user_id
                AND nl.vocab_id = uvp.vocab_id
                AND nl.created_at > NOW() - INTERVAL '4 hours'
          )
        ORDER BY uvp.next_review_at ASC
        LIMIT ?
    `, w.batchSize).Scan(&dueReviews).Error

    if err != nil || len(dueReviews) == 0 {
        return
    }

    log.Printf("[SRS Worker] Dispatching %d alerts", len(dueReviews))

    // 3. Goroutine pool for parallel notification dispatch
    jobs := make(chan DueReview, len(dueReviews))
    var wg sync.WaitGroup

    for i := 0; i < w.workers; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for review := range jobs {
                w.sendAlert(ctx, review)
            }
        }(i)
    }

    for _, review := range dueReviews {
        jobs <- review
    }
    close(jobs)
    wg.Wait()
}

func (w *SRSAlertWorker) sendAlert(ctx context.Context, r DueReview) {
    title := fmt.Sprintf("📚 Review: \"%s\"", r.Word)
    body := fmt.Sprintf("%s\n%s", r.ExampleEn, r.ExampleVi)

    err := w.notifier.SendPush(ctx, r.UserID, title, body, map[string]string{
        "vocab_id": fmt.Sprintf("%d", r.VocabID),
        "level":    fmt.Sprintf("%d", r.Level),
    })

    status := "sent"
    errMsg := ""
    if err != nil {
        status = "failed"
        errMsg = err.Error()
    }

    w.db.Exec(`
        INSERT INTO notification_log (user_id, vocab_id, level, channel, status, error_message)
        VALUES (?, ?, ?, 'push', ?, ?)
    `, r.UserID, r.VocabID, r.Level, status, errMsg)
}
```

### 3.5 Bottleneck Prevention Summary

| Problem | Solution |
|---------|----------|
| Duplicate alerts | `notification_log` with 4-hour window check |
| Multiple worker instances | Redis `SETNX` distributed lock |
| Large user base (100K+) | Batch `LIMIT`, goroutine pool, partial index |
| Slow push sending | Concurrent goroutine pool per batch |
| DB connection exhaustion | Single query + in-memory distribution |
| Missed alerts due to crash | Idempotent: re-picked on next scan if log missing |

---

## 4. Prompt Engineering Strategy

### 4.1 Prompt 1: Vocabulary Generation with 3 Levels

```
SYSTEM PROMPT (DeepSeek-Chat - Fast Inference)

You are VocabForge, an expert English linguistics engine specialized in creating
structured, multi-level vocabulary learning content for Vietnamese ESL learners.

Generate comprehensive vocabulary entries with precisely 3 proficiency levels
for each word/phrase provided.

=== OUTPUT FORMAT (STRICT JSON - NO MARKDOWN WRAPPING) ===
{
  "word": "<input word>",
  "phonetics": "<IPA pronunciation>",
  "pos": "<noun|verb|adj|adv|phrasal_verb|idiom>",
  "topic_tags": ["<tag1>", "<tag2>"],
  "levels": [
    {
      "level": 1,
      "label": "Simple (Functional)",
      "meaning_en": "<simple definition, max 10 words>",
      "meaning_vi": "<simple Vietnamese translation>",
      "example_en": "<SHORT S+V+O sentence, no idioms, no phrasal verbs, Filipino business English style>",
      "example_vi": "<Vietnamese translation>",
      "usage_note": "<why level 1: direct, functional, minimal complexity>"
    },
    {
      "level": 2,
      "label": "Normal (Native Casual)",
      "meaning_en": "<natural definition with context>",
      "meaning_vi": "<natural Vietnamese>",
      "example_en": "<Natural native-speaker sentence, may use phrasal verbs, light idioms, contractions>",
      "example_vi": "<Vietnamese, preserving casual tone>",
      "usage_note": "<why level 2: natural speech, conversational>"
    },
    {
      "level": 3,
      "label": "Complex (Academic/Professional)",
      "meaning_en": "<precise academic definition>",
      "meaning_vi": "<formal Vietnamese>",
      "example_en": "<Complex sentence: passive voice, subjunctive, advanced connectors, IELTS/TOEFL standard>",
      "example_vi": "<formal Vietnamese>",
      "usage_note": "<why level 3: academic, formal, exam-appropriate>"
    }
  ]
}

=== CRITICAL RULES ===
1. Level 1: Max 12 words, S+V+O only, NO phrasal verbs, NO idioms, NO contractions.
2. Level 2: Natural native speech, phrasal verbs allowed, light idioms, contractions OK.
3. Level 3: Academic/formal register, complex grammar, IELTS Writing Task 2 standard.
4. All example sentences must be original and use the target word naturally.
5. Vietnamese translations must be natural, fluent, preserve register.
6. Output ONLY valid JSON. No markdown, no extra text.
```

### 4.2 Prompt 2: Writing Evaluation with 3-Level Rewrite

```
SYSTEM PROMPT (DeepSeek-Reasoning - Deep Think Mode)

You are WriteSense, an expert English writing evaluator for Vietnamese ESL learners.
Analyze student writing surgically, provide corrective feedback, and 3-tier rewriting.

Be: constructive, specific, actionable, culturally aware of Vietnamese-English interference.

=== INPUT ===
1. TOPIC: The writing task description
2. STUDENT TEXT: The learner's original writing
3. REQUIRED VOCABULARY: Words they were required to use
4. LEARNER LEVEL: A2 / B1 / B2 / C1

=== OUTPUT FORMAT (STRICT JSON - NO MARKDOWN) ===
{
  "overall_score": <0-100>,
  "scores_breakdown": {
    "grammar": <0-100>,
    "vocabulary": <0-100>,
    "coherence_cohesion": <0-100>,
    "task_achievement": <0-100>
  },
  "required_vocab_usage": [
    {"word": "<word>", "used": <bool>, "correctly_used": <bool|null>, "comment": "<feedback>"}
  ],
  "errors": [
    {
      "type": "<grammar|spelling|word_choice|collocation|punctuation|article|preposition|tense>",
      "severity": <1=minor, 2=moderate, 3=critical>,
      "original": "<incorrect text span>",
      "correction": "<corrected version>",
      "explanation": "<Explanation in VIETNAMESE>",
      "rule": "<grammar rule reference>"
    }
  ],
  "strengths": ["<at least 2 items>"],
  "areas_for_improvement": ["<at least 2 items>"],
  "rewrites": [
    {
      "level": 1,
      "label": "Simple (Functional)",
      "rewritten_text": "<short sentences, S+V+O, clear, no complexity>",
      "changes_summary": "<what changed, why level 1>"
    },
    {
      "level": 2,
      "label": "Normal (Native Casual)",
      "rewritten_text": "<natural native style, phrasal verbs, flowing>",
      "changes_summary": "<what changed, why level 2>"
    },
    {
      "level": 3,
      "label": "Complex (Academic)",
      "rewritten_text": "<academic register, complex structures, IELTS 7.5+>",
      "changes_summary": "<what changed, why level 3>"
    }
  ],
  "vietnamese_explanation": "<3-4 sentence summary in Vietnamese>"
}

=== KEY VIETNAMESE LEARNER ERRORS TO WATCH ===
- Missing articles: "I went to store" -> "I went to the store"
- Wrong prepositions: "married with" -> "married to", "depend of" -> "depend on"
- Countable/uncountable: "many information" -> "much information"
- Tense mixing in narratives
- Word-for-word translation from Vietnamese
- Missing plural -s: "three apple" -> "three apples"

=== REWRITE LEVEL GUIDELINES ===
Level 1: Short (10-15 words), S+V+O, active voice, A2-B1 vocab, Filipino business style
Level 2: Mixed lengths, phrasal verbs OK, light idioms, contractions, conversational professional
Level 3: Complex structures, C1-C2 vocab, passive voice, advanced connectors, formal register

=== RULES ===
1. Output ONLY valid JSON. No markdown, no preamble.
2. Be honest but encouraging - Vietnamese learners need both.
3. Every error must have Vietnamese explanation.
4. Rewrites preserve original IDEAS while fixing EXPRESSION.
5. Required vocab feedback must be precise.
```

---

## 5. Development Roadmap

### Overview: 6 Phases (24 Weeks to Full MVP)

```
WEEK:  1  2  3  4  5  6  7  8  9  10 11 12 13 14 15 16 17 18 19 20 21 22 23 24
       ├─────────────┤├─────────────┤├─────────────┤├─────────────┤├─────────────┤
       │   Phase 1    │   Phase 2    │   Phase 3    │   Phase 4    │   Phase 5    │
       │ Foundation    │ AI + Content │ Spaced Rep   │ Writing+Quiz │ Polish+Launch│
```

---

### Phase 1: Foundation (Weeks 1-4)
**Theme: "Walking Skeleton" - deployable, minimal features**

#### Week 1-2: Infrastructure & Core Backend
- [x] Docker Compose setup (PostgreSQL 16 + Go API + Vue 3 Web) - *already done*
- [ ] Add Redis 7 Alpine to `docker-compose.yaml`
- [ ] Set up database migration framework (`golang-migrate` or GORM AutoMigrate)
- [ ] Run migration `0001_init.sql` (updated schema from Section 2)
- [ ] Implement JWT auth: `POST /api/v1/auth/login`, `POST /api/v1/auth/register`
- [ ] Basic CRUD: `POST /api/v1/vocabs`, `GET /api/v1/vocabs/:id`
- [ ] Set up MinIO container for local S3-compatible storage
- [ ] File upload: `POST /api/v1/audio/upload`

#### Week 3-4: Core Frontend
- [ ] Vue 3 structure: Pinia, Vue Router, TypeScript
- [ ] `FlashCard` component with flip animation + Level 1/2/3 toggle
- [ ] `AudioRecorder` component (MediaRecorder API)
- [ ] `AudioPlayer` component for TTS
- [ ] Dashboard page: user stats, daily streak, due reviews count
- [ ] Vocab Search page
- [ ] **Demo:** Login -> Search word -> See 3 levels on flashcard

---

### Phase 2: AI Integration & Content Generation (Weeks 5-8)
**Theme: "AI-Powered Content"**

#### Week 5-6: AI Orchestrator
- [ ] `AIOrchestrator` in Go with worker pool pattern
- [ ] DeepSeek-Chat API client (HTTP + retry + circuit breaker)
- [ ] Implement Prompt Template 1 (Vocab Generation) + parse JSON response
- [ ] `POST /api/v1/vocabs/generate` endpoint
- [ ] Redis caching: `vocab:{word}` key, TTL 24h
- [ ] Build `AudioPlayer` component with browser Web Speech API (TTS free, song ngữ Anh-Việt)
- [ ] Implement `useSpeechSynthesis` composable for voice selection (en-US + vi-VN)

#### Week 7-8: AI Search & Content
- [ ] `GET /api/v1/search?q=...` (full-text + semantic fallback via DeepSeek)
- [ ] Batch vocabulary import (CSV/JSON + AI generates 3 levels)
- [ ] Vocabulary detail page (3 levels side-by-side)
- [ ] AI "Explore" page: word of the day, topic browse
- [ ] **Demo:** Type word -> AI generates 3 levels -> TTS playback

---

### Phase 3: Spaced Repetition System (Weeks 9-12)
**Theme: "Never Forget" - the memory engine**

#### Week 9-10: SRS Engine
- [ ] Implement FSRS algorithm in Go (as in Section 3)
- [ ] `POST /api/v1/srs/review` endpoint (rate word -> FSRS recalculates)
- [ ] `GET /api/v1/srs/due-reviews?limit=20` endpoint
- [ ] Review Session page: flashcard carousel with Forgot/Hard/Good/Easy buttons
- [ ] Study session tracking (`review_sessions` table)

#### Week 11-12: SRS Alert Worker
- [ ] `SRSAlertWorker` goroutine (Section 3.4)
- [ ] Redis distributed lock (`SETNX`)
- [ ] Notification deduplication via `notification_log`
- [ ] Firebase Cloud Messaging integration
- [ ] `SpacedRepAlert` frontend component (toast + browser notification)
- [ ] Notification preferences page
- [ ] `user_streaks` tracking
- [ ] **Demo:** Review 20 flashcards -> SRS schedules -> Push notification fires on time

---

### Phase 4: Writing Lab & Quiz System (Weeks 13-18)
**Theme: "Produce, Don't Just Consume"**

#### Week 13-15: Writing Lab
- [ ] Writing Submission page (editor, prompt, required vocab list)
- [ ] `POST /api/v1/writing/submit` endpoint
- [ ] DeepSeek-Reasoning API client (longer timeout, higher tokens)
- [ ] Implement Prompt Template 2 (Writing Evaluation)
- [ ] Store: `writing_submissions` + `writing_errors` + `writing_rewrites`
- [ ] Writing Feedback page (scores, errors, 3-level rewrite tabs)
- [ ] Writing history page
- [ ] Enforce: user MUST use >=3 words from active vocab

#### Week 16-18: Quiz System
- [ ] `POST /api/v1/quizzes/generate` (AI generates 10-question quiz)
- [ ] `POST /api/v1/quizzes/:id/attempt` (submit answers)
- [ ] Quiz results page (scores, feedback, time stats)
- [ ] `POST /api/v1/quizzes/generate-exam?type=toeic` (full exam)
- [ ] Quiz history page
- [ ] **Demo:** Study 30 words -> AI quiz -> AI evaluates writing -> 3-level rewrite

---

### Phase 5: Polish & Launch (Weeks 19-24)
**Theme: "Production-Ready"**

#### Week 19-20: Performance & Reliability
- [ ] Query optimization (`EXPLAIN ANALYZE` on hot paths)
- [ ] Index audit (GIN, partial indexes)
- [ ] Redis cache hit rate monitoring
- [ ] API rate limiting (token bucket per user/IP)
- [ ] AI Orchestrator timeout & retry tuning
- [ ] Load testing (`k6` or `vegeta`): 100 concurrent users
- [ ] P99 targets: <500ms cached, <5s AI-generated

#### Week 21-22: Mobile PWA
- [ ] PWA manifest + service worker (offline flashcard)
- [ ] Responsive optimization (375px - 428px)
- [ ] Mobile audio recording testing (Safari, Chrome Android)
- [ ] "Add to Home Screen" prompt
- [ ] Mobile push notification testing

#### Week 23-24: Launch Preparation
- [ ] Security audit (JWT, input sanitization, SQL injection, CORS)
- [ ] Seed 1000 core TOEIC/IELTS vocabulary
- [ ] User onboarding flow (tutorial + level assessment)
- [ ] Admin dashboard (user stats, content stats, AI cost tracking)
- [ ] Monitoring: health checks, Prometheus metrics, error logging
- [ ] CI/CD: GitHub Actions -> Docker -> Deploy
- [ ] **LAUNCH MVP with all 6 core features**

---

### Phase 6: Post-Launch (Week 25+)
- [ ] A/B test prompt templates
- [ ] Train custom FSRS weights on user data
- [ ] Expand to Golang/Java/Linux/DevOps content
- [ ] Social features (leaderboards, study groups)
- [ ] AI personalized learning paths
- [ ] Offline mode (IndexedDB + sync)
- [ ] Multi-language (Korean, Japanese learners)

---

### Complete API Endpoint Map

| Method | Endpoint | Phase | Description |
|--------|----------|-------|-------------|
| POST | `/api/v1/auth/register` | 1 | User registration |
| POST | `/api/v1/auth/login` | 1 | JWT login |
| GET | `/api/v1/vocabs` | 1 | List/search vocabularies |
| GET | `/api/v1/vocabs/:id` | 1 | Get vocab with 3-level sentences |
| POST | `/api/v1/vocabs/generate` | 2 | AI generate vocab with 3 levels |
| POST | `/api/v1/audio/upload` | 1 | Upload user recording (for STT assessment) |
| POST | `/api/v1/speech/assess` | 3 | Pronunciation assessment (STT) |
| GET | `/api/v1/srs/due-reviews` | 3 | Get words due for SRS review |
| POST | `/api/v1/srs/review` | 3 | Submit review rating (1-4) |
| GET | `/api/v1/search` | 2 | AI-powered semantic search |
| POST | `/api/v1/writing/submit` | 4 | Submit writing for AI evaluation |
| GET | `/api/v1/writing/:id` | 4 | Get writing feedback + rewrites |
| GET | `/api/v1/writing/history` | 4 | User's writing history |
| POST | `/api/v1/quizzes/generate` | 4 | Generate quiz (AI or template) |
| GET | `/api/v1/quizzes/:id` | 4 | Get quiz questions |
| POST | `/api/v1/quizzes/:id/attempt` | 4 | Submit quiz answers |
| GET | `/api/v1/dashboard` | 1 | User dashboard stats |
| GET | `/api/v1/notifications/preferences` | 3 | Get alert preferences |
| PUT | `/api/v1/notifications/preferences` | 3 | Update alert preferences |

---

*End of Architecture & Design Document*