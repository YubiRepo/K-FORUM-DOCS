# Database Schema — Q&A / FAQ Module

> **Stack:** Golang + PostgreSQL  
> **Dibuat:** 2026-05-30  
> **Module:** QnA / FAQ

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [PostgreSQL DDL Schema](#postgresql-ddl-schema)
3. [Golang Structs (ORM Mapping)](#golang-structs-orm-mapping)
4. [Migrations](#migrations)
5. [Enum Values](#enum-values)
6. [Query Examples](#query-examples)

---

## Overview Relasi

```
users
  └── qna_questions          (pertanyaan yang diajukan member)
  └── qna_answers            (jawaban dari superadmin)
  └── qna_faq_helpful_votes  (vote helpful/not-helpful per FAQ item)

qna_categories
  └── qna_faq_items          (FAQ statis yang dikurasi admin)
  └── qna_questions          (kategori pertanyaan member)

qna_faq_items
  └── qna_faq_helpful_votes  (vote tracking per item)

qna_questions
  └── qna_answers            (jawaban admin per pertanyaan)

qna_bot_config               (konfigurasi bot fallback, singleton)
```

---

## PostgreSQL DDL Schema

```sql
-- ============================================================================
-- QnA / FAQ MODULE DATABASE SCHEMA
-- Stack: PostgreSQL 13+
-- Created: 2026-05-30
-- ============================================================================


-- ============================================================================
-- 1. QNA_CATEGORIES
-- Master data kategori FAQ (Imigrasi, Pajak, Sistem Aplikasi, dll.)
-- ============================================================================

CREATE TABLE qna_categories (
  id          UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100)  NOT NULL,
  slug        VARCHAR(120)  NOT NULL UNIQUE,
  description TEXT          NULL,
  icon        VARCHAR(100)  NULL,        -- icon name (e.g. tabler icon key)
  color       VARCHAR(20)   NULL,        -- hex color for UI (e.g. "#534AB7")
  sort_order  INTEGER       NOT NULL DEFAULT 0,
  is_active   BOOLEAN       NOT NULL DEFAULT true,
  created_by  UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE UNIQUE INDEX idx_qna_categories_slug
  ON qna_categories (slug);

CREATE INDEX idx_qna_categories_active_order
  ON qna_categories (is_active, sort_order)
  WHERE is_active = true;


-- ============================================================================
-- 2. QNA_FAQ_ITEMS
-- FAQ statis yang dikurasi dan dipublish oleh superadmin.
-- ============================================================================

CREATE TABLE qna_faq_items (
  id           UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id  UUID         NOT NULL REFERENCES qna_categories(id) ON DELETE RESTRICT,

  -- Content
  question     TEXT         NOT NULL,
  answer       TEXT         NOT NULL,
  tags         TEXT[]       NULL,         -- array tag untuk full-text search

  -- Status
  status       VARCHAR(20)  NOT NULL DEFAULT 'draft',   -- draft, published, archived

  -- Stats
  view_count   INTEGER      NOT NULL DEFAULT 0,
  helpful_count     INTEGER NOT NULL DEFAULT 0,
  not_helpful_count INTEGER NOT NULL DEFAULT 0,

  -- Sort
  sort_order   INTEGER      NOT NULL DEFAULT 0,
  is_pinned    BOOLEAN      NOT NULL DEFAULT false,

  -- Audit
  created_by   UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by   UUID         NULL     REFERENCES users(id) ON DELETE SET NULL,
  published_at TIMESTAMPTZ  NULL,
  archived_at  TIMESTAMPTZ  NULL,
  created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_qna_faq_items_category
  ON qna_faq_items (category_id, status, sort_order);

CREATE INDEX idx_qna_faq_items_published
  ON qna_faq_items (status, is_pinned, sort_order)
  WHERE status = 'published';

CREATE INDEX idx_qna_faq_items_tags
  ON qna_faq_items USING GIN (tags);

-- Full-text search index (Indonesian + English)
CREATE INDEX idx_qna_faq_items_fts
  ON qna_faq_items
  USING GIN (
    to_tsvector('simple', question || ' ' || answer)
  );


-- ============================================================================
-- 3. QNA_FAQ_HELPFUL_VOTES
-- Tracking vote helpful / not helpful per FAQ item per user.
-- Satu user hanya bisa vote sekali per item.
-- ============================================================================

CREATE TABLE qna_faq_helpful_votes (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  faq_id     UUID        NOT NULL REFERENCES qna_faq_items(id) ON DELETE CASCADE,
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  is_helpful BOOLEAN     NOT NULL,   -- true = helpful, false = not helpful
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Satu user hanya bisa vote sekali per FAQ item
CREATE UNIQUE INDEX idx_qna_helpful_votes_unique
  ON qna_faq_helpful_votes (faq_id, user_id);

CREATE INDEX idx_qna_helpful_votes_faq
  ON qna_faq_helpful_votes (faq_id);


-- ============================================================================
-- 4. QNA_QUESTIONS
-- Pertanyaan yang diajukan member jika tidak menemukan jawaban di FAQ.
-- ============================================================================

CREATE TABLE qna_questions (
  id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  category_id      UUID         NULL REFERENCES qna_categories(id) ON DELETE SET NULL,

  -- Content
  question_text    TEXT         NOT NULL,
  attachment_urls  TEXT[]       NULL,   -- optional attachments (gambar, dokumen)

  -- Status
  -- pending: belum dijawab
  -- answered: sudah dijawab admin
  -- rejected: ditolak (duplikat / tidak relevan)
  -- converted: dijadikan FAQ publik
  status           VARCHAR(20)  NOT NULL DEFAULT 'pending',

  -- Bot fallback
  bot_responded    BOOLEAN      NOT NULL DEFAULT false,
  bot_response     TEXT         NULL,

  -- Rejection
  rejection_reason TEXT         NULL,
  rejected_by      UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  rejected_at      TIMESTAMPTZ  NULL,

  -- Timestamps
  answered_at      TIMESTAMPTZ  NULL,
  created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_qna_questions_user
  ON qna_questions (user_id, created_at DESC);

CREATE INDEX idx_qna_questions_status
  ON qna_questions (status, created_at DESC);

CREATE INDEX idx_qna_questions_category_status
  ON qna_questions (category_id, status);

-- Pending questions (common backoffice query)
CREATE INDEX idx_qna_questions_pending
  ON qna_questions (created_at DESC)
  WHERE status = 'pending';


-- ============================================================================
-- 5. QNA_ANSWERS
-- Jawaban dari superadmin untuk pertanyaan member.
-- Satu pertanyaan bisa punya satu jawaban utama (is_primary=true).
-- ============================================================================

CREATE TABLE qna_answers (
  id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id     UUID         NOT NULL REFERENCES qna_questions(id) ON DELETE CASCADE,
  answered_by     UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  -- Content
  answer_text     TEXT         NOT NULL,
  attachment_urls TEXT[]       NULL,

  -- Visibility
  is_public       BOOLEAN      NOT NULL DEFAULT true,   -- false = hanya terlihat oleh member yang bertanya
  is_primary      BOOLEAN      NOT NULL DEFAULT true,   -- true = jawaban utama (trigger notif)

  -- Convert to FAQ
  converted_to_faq_id UUID     NULL REFERENCES qna_faq_items(id) ON DELETE SET NULL,

  -- Timestamps
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Satu pertanyaan hanya boleh punya satu jawaban utama
CREATE UNIQUE INDEX idx_qna_answers_primary_unique
  ON qna_answers (question_id)
  WHERE is_primary = true;

CREATE INDEX idx_qna_answers_question
  ON qna_answers (question_id, created_at DESC);

CREATE INDEX idx_qna_answers_answered_by
  ON qna_answers (answered_by);


-- ============================================================================
-- 6. QNA_BOT_CONFIG
-- Konfigurasi bot fallback (singleton — hanya satu row).
-- ============================================================================

CREATE TABLE qna_bot_config (
  id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  enabled             BOOLEAN     NOT NULL DEFAULT true,
  fallback_message    TEXT        NOT NULL DEFAULT
    'Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja.',
  suggest_contact     BOOLEAN     NOT NULL DEFAULT true,
  contact_info        TEXT        NULL,   -- nomor / email kontak yang disarankan

  -- Keyword matching threshold (0.0–1.0)
  -- Jika similarity score FAQ < threshold → bot menjawab
  similarity_threshold NUMERIC(3,2) NOT NULL DEFAULT 0.60,

  updated_by          UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Pastikan hanya ada satu row config
CREATE UNIQUE INDEX idx_qna_bot_config_singleton
  ON qna_bot_config ((true));
```

---

## Golang Structs (ORM Mapping)

```go
package qna

import (
    "time"
    "github.com/lib/pq"
)

// QnaCategory — Master kategori FAQ
type QnaCategory struct {
    ID          string     `db:"id"          json:"id"`
    Name        string     `db:"name"        json:"name"`
    Slug        string     `db:"slug"        json:"slug"`
    Description *string    `db:"description" json:"description,omitempty"`
    Icon        *string    `db:"icon"        json:"icon,omitempty"`
    Color       *string    `db:"color"       json:"color,omitempty"`
    SortOrder   int        `db:"sort_order"  json:"sort_order"`
    IsActive    bool       `db:"is_active"   json:"is_active"`
    CreatedBy   string     `db:"created_by"  json:"created_by"`
    CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
    UpdatedAt   time.Time  `db:"updated_at"  json:"updated_at"`
}

// QnaFaqItem — FAQ statis dikurasi superadmin
type QnaFaqItem struct {
    ID              string         `db:"id"               json:"id"`
    CategoryID      string         `db:"category_id"      json:"category_id"`
    Question        string         `db:"question"         json:"question"`
    Answer          string         `db:"answer"           json:"answer"`
    Tags            pq.StringArray `db:"tags"             json:"tags,omitempty"`
    Status          string         `db:"status"           json:"status"`
    ViewCount       int            `db:"view_count"       json:"view_count"`
    HelpfulCount    int            `db:"helpful_count"    json:"helpful_count"`
    NotHelpfulCount int            `db:"not_helpful_count" json:"not_helpful_count"`
    SortOrder       int            `db:"sort_order"       json:"sort_order"`
    IsPinned        bool           `db:"is_pinned"        json:"is_pinned"`
    CreatedBy       string         `db:"created_by"       json:"created_by"`
    UpdatedBy       *string        `db:"updated_by"       json:"updated_by,omitempty"`
    PublishedAt     *time.Time     `db:"published_at"     json:"published_at,omitempty"`
    ArchivedAt      *time.Time     `db:"archived_at"      json:"archived_at,omitempty"`
    CreatedAt       time.Time      `db:"created_at"       json:"created_at"`
    UpdatedAt       time.Time      `db:"updated_at"       json:"updated_at"`
}

const (
    FaqStatusDraft     = "draft"
    FaqStatusPublished = "published"
    FaqStatusArchived  = "archived"
)

// QnaFaqHelpfulVote — Tracking vote per user per FAQ
type QnaFaqHelpfulVote struct {
    ID        string    `db:"id"         json:"id"`
    FaqID     string    `db:"faq_id"     json:"faq_id"`
    UserID    string    `db:"user_id"    json:"user_id"`
    IsHelpful bool      `db:"is_helpful" json:"is_helpful"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// QnaQuestion — Pertanyaan yang diajukan member
type QnaQuestion struct {
    ID             string         `db:"id"              json:"id"`
    UserID         string         `db:"user_id"         json:"user_id"`
    CategoryID     *string        `db:"category_id"     json:"category_id,omitempty"`
    QuestionText   string         `db:"question_text"   json:"question_text"`
    AttachmentURLs pq.StringArray `db:"attachment_urls" json:"attachment_urls,omitempty"`
    Status         string         `db:"status"          json:"status"`
    BotResponded   bool           `db:"bot_responded"   json:"bot_responded"`
    BotResponse    *string        `db:"bot_response"    json:"bot_response,omitempty"`
    RejectionReason *string       `db:"rejection_reason" json:"rejection_reason,omitempty"`
    RejectedBy     *string        `db:"rejected_by"     json:"rejected_by,omitempty"`
    RejectedAt     *time.Time     `db:"rejected_at"     json:"rejected_at,omitempty"`
    AnsweredAt     *time.Time     `db:"answered_at"     json:"answered_at,omitempty"`
    CreatedAt      time.Time      `db:"created_at"      json:"created_at"`
    UpdatedAt      time.Time      `db:"updated_at"      json:"updated_at"`
}

const (
    QuestionStatusPending   = "pending"
    QuestionStatusAnswered  = "answered"
    QuestionStatusRejected  = "rejected"
    QuestionStatusConverted = "converted"
)

// QnaAnswer — Jawaban admin untuk pertanyaan member
type QnaAnswer struct {
    ID               string         `db:"id"                  json:"id"`
    QuestionID       string         `db:"question_id"         json:"question_id"`
    AnsweredBy       string         `db:"answered_by"         json:"answered_by"`
    AnswerText       string         `db:"answer_text"         json:"answer_text"`
    AttachmentURLs   pq.StringArray `db:"attachment_urls"     json:"attachment_urls,omitempty"`
    IsPublic         bool           `db:"is_public"           json:"is_public"`
    IsPrimary        bool           `db:"is_primary"          json:"is_primary"`
    ConvertedToFaqID *string        `db:"converted_to_faq_id" json:"converted_to_faq_id,omitempty"`
    CreatedAt        time.Time      `db:"created_at"          json:"created_at"`
    UpdatedAt        time.Time      `db:"updated_at"          json:"updated_at"`
}

// QnaBotConfig — Konfigurasi bot fallback (singleton)
type QnaBotConfig struct {
    ID                  string    `db:"id"                   json:"id"`
    Enabled             bool      `db:"enabled"              json:"enabled"`
    FallbackMessage     string    `db:"fallback_message"     json:"fallback_message"`
    SuggestContact      bool      `db:"suggest_contact"      json:"suggest_contact"`
    ContactInfo         *string   `db:"contact_info"         json:"contact_info,omitempty"`
    SimilarityThreshold float64   `db:"similarity_threshold" json:"similarity_threshold"`
    UpdatedBy           *string   `db:"updated_by"           json:"updated_by,omitempty"`
    UpdatedAt           time.Time `db:"updated_at"           json:"updated_at"`
}
```

---

## Migrations

```sql
-- ============================================================================
-- Migration 1: Create qna_categories
-- ============================================================================
CREATE TABLE IF NOT EXISTS qna_categories (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100) NOT NULL,
  slug        VARCHAR(120) NOT NULL UNIQUE,
  description TEXT         NULL,
  icon        VARCHAR(100) NULL,
  color       VARCHAR(20)  NULL,
  sort_order  INTEGER      NOT NULL DEFAULT 0,
  is_active   BOOLEAN      NOT NULL DEFAULT true,
  created_by  UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_qna_categories_slug
  ON qna_categories (slug);
CREATE INDEX IF NOT EXISTS idx_qna_categories_active_order
  ON qna_categories (is_active, sort_order) WHERE is_active = true;


-- ============================================================================
-- Migration 2: Create qna_faq_items
-- ============================================================================
CREATE TABLE IF NOT EXISTS qna_faq_items (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  category_id       UUID        NOT NULL REFERENCES qna_categories(id) ON DELETE RESTRICT,
  question          TEXT        NOT NULL,
  answer            TEXT        NOT NULL,
  tags              TEXT[]      NULL,
  status            VARCHAR(20) NOT NULL DEFAULT 'draft',
  view_count        INTEGER     NOT NULL DEFAULT 0,
  helpful_count     INTEGER     NOT NULL DEFAULT 0,
  not_helpful_count INTEGER     NOT NULL DEFAULT 0,
  sort_order        INTEGER     NOT NULL DEFAULT 0,
  is_pinned         BOOLEAN     NOT NULL DEFAULT false,
  created_by        UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by        UUID        NULL     REFERENCES users(id) ON DELETE SET NULL,
  published_at      TIMESTAMPTZ NULL,
  archived_at       TIMESTAMPTZ NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_qna_faq_items_category
  ON qna_faq_items (category_id, status, sort_order);
CREATE INDEX IF NOT EXISTS idx_qna_faq_items_published
  ON qna_faq_items (status, is_pinned, sort_order) WHERE status = 'published';
CREATE INDEX IF NOT EXISTS idx_qna_faq_items_tags
  ON qna_faq_items USING GIN (tags);
CREATE INDEX IF NOT EXISTS idx_qna_faq_items_fts
  ON qna_faq_items USING GIN (to_tsvector('simple', question || ' ' || answer));


-- ============================================================================
-- Migration 3: Create qna_faq_helpful_votes
-- ============================================================================
CREATE TABLE IF NOT EXISTS qna_faq_helpful_votes (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  faq_id     UUID        NOT NULL REFERENCES qna_faq_items(id) ON DELETE CASCADE,
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  is_helpful BOOLEAN     NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_qna_helpful_votes_unique
  ON qna_faq_helpful_votes (faq_id, user_id);
CREATE INDEX IF NOT EXISTS idx_qna_helpful_votes_faq
  ON qna_faq_helpful_votes (faq_id);


-- ============================================================================
-- Migration 4: Create qna_questions
-- ============================================================================
CREATE TABLE IF NOT EXISTS qna_questions (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id          UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  category_id      UUID        NULL REFERENCES qna_categories(id) ON DELETE SET NULL,
  question_text    TEXT        NOT NULL,
  attachment_urls  TEXT[]      NULL,
  status           VARCHAR(20) NOT NULL DEFAULT 'pending',
  bot_responded    BOOLEAN     NOT NULL DEFAULT false,
  bot_response     TEXT        NULL,
  rejection_reason TEXT        NULL,
  rejected_by      UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  rejected_at      TIMESTAMPTZ NULL,
  answered_at      TIMESTAMPTZ NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_qna_questions_user
  ON qna_questions (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_qna_questions_status
  ON qna_questions (status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_qna_questions_category_status
  ON qna_questions (category_id, status);
CREATE INDEX IF NOT EXISTS idx_qna_questions_pending
  ON qna_questions (created_at DESC) WHERE status = 'pending';


-- ============================================================================
-- Migration 5: Create qna_answers
-- ============================================================================
CREATE TABLE IF NOT EXISTS qna_answers (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id       UUID        NOT NULL REFERENCES qna_questions(id) ON DELETE CASCADE,
  answered_by       UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  answer_text       TEXT        NOT NULL,
  attachment_urls   TEXT[]      NULL,
  is_public         BOOLEAN     NOT NULL DEFAULT true,
  is_primary        BOOLEAN     NOT NULL DEFAULT true,
  converted_to_faq_id UUID      NULL REFERENCES qna_faq_items(id) ON DELETE SET NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_qna_answers_primary_unique
  ON qna_answers (question_id) WHERE is_primary = true;
CREATE INDEX IF NOT EXISTS idx_qna_answers_question
  ON qna_answers (question_id, created_at DESC);


-- ============================================================================
-- Migration 6: Create qna_bot_config + seed default
-- ============================================================================
CREATE TABLE IF NOT EXISTS qna_bot_config (
  id                   UUID           PRIMARY KEY DEFAULT gen_random_uuid(),
  enabled              BOOLEAN        NOT NULL DEFAULT true,
  fallback_message     TEXT           NOT NULL DEFAULT 'Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja.',
  suggest_contact      BOOLEAN        NOT NULL DEFAULT true,
  contact_info         TEXT           NULL,
  similarity_threshold NUMERIC(3,2)   NOT NULL DEFAULT 0.60,
  updated_by           UUID           NULL REFERENCES users(id) ON DELETE SET NULL,
  updated_at           TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_qna_bot_config_singleton
  ON qna_bot_config ((true));

-- Seed default config
INSERT INTO qna_bot_config (enabled, fallback_message, suggest_contact, similarity_threshold)
VALUES (
  true,
  'Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja.',
  true,
  0.60
)
ON CONFLICT DO NOTHING;


-- ============================================================================
-- Migration 7: Auto-update trigger for updated_at
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_qna_categories_updated_at
  BEFORE UPDATE ON qna_categories
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_qna_faq_items_updated_at
  BEFORE UPDATE ON qna_faq_items
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_qna_faq_helpful_votes_updated_at
  BEFORE UPDATE ON qna_faq_helpful_votes
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_qna_questions_updated_at
  BEFORE UPDATE ON qna_questions
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_qna_answers_updated_at
  BEFORE UPDATE ON qna_answers
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- Migration 8: Seed default categories
-- ============================================================================
-- (Jalankan setelah superadmin user tersedia, isi created_by dengan UUID superadmin)
-- INSERT INTO qna_categories (name, slug, icon, color, sort_order, created_by) VALUES
--   ('Imigrasi & Visa',          'imigrasi-visa',         'ti-passport',   '#534AB7', 1, '<superadmin_uuid>'),
--   ('Pajak',                    'pajak',                 'ti-receipt',    '#0F6E56', 2, '<superadmin_uuid>'),
--   ('Ketenagakerjaan',          'ketenagakerjaan',       'ti-briefcase',  '#993C1D', 3, '<superadmin_uuid>'),
--   ('Pendirian Usaha',          'pendirian-usaha',       'ti-building',   '#854F0B', 4, '<superadmin_uuid>'),
--   ('Properti & Kepemilikan',   'properti',              'ti-home',       '#185FA5', 5, '<superadmin_uuid>'),
--   ('Sistem Aplikasi KAI',      'sistem-aplikasi',       'ti-device-mobile', '#639922', 6, '<superadmin_uuid>'),
--   ('Umum',                     'umum',                  'ti-info-circle','#888780', 7, '<superadmin_uuid>');
```

---

## Enum Values

### `qna_faq_items.status`

| Value | Keterangan |
|-------|-----------|
| `draft` | Belum dipublish, hanya terlihat oleh admin |
| `published` | Aktif dan terlihat oleh member |
| `archived` | Nonaktif, tersimpan untuk referensi |

### `qna_questions.status`

| Value | Keterangan |
|-------|-----------|
| `pending` | Belum dijawab admin |
| `answered` | Sudah dijawab, notifikasi terkirim |
| `rejected` | Ditolak (duplikat / tidak relevan) |
| `converted` | Jawaban dijadikan FAQ publik |

---

## Query Examples

```sql
-- 1. Ambil semua kategori aktif terurut
SELECT id, name, slug, icon, color, sort_order
FROM qna_categories
WHERE is_active = true
ORDER BY sort_order ASC;


-- 2. Ambil FAQ published per kategori (dengan sort pinned dulu)
SELECT f.id, f.question, f.answer, f.helpful_count, f.not_helpful_count
FROM qna_faq_items f
WHERE f.category_id = $1
  AND f.status = 'published'
ORDER BY f.is_pinned DESC, f.sort_order ASC;


-- 3. Full-text search FAQ
SELECT id, question, answer, category_id,
       ts_rank(to_tsvector('simple', question || ' ' || answer),
               plainto_tsquery('simple', $1)) AS rank
FROM qna_faq_items
WHERE status = 'published'
  AND to_tsvector('simple', question || ' ' || answer) @@ plainto_tsquery('simple', $1)
ORDER BY rank DESC
LIMIT 10;


-- 4. Cek apakah user sudah vote FAQ ini
SELECT is_helpful
FROM qna_faq_helpful_votes
WHERE faq_id = $1 AND user_id = $2;


-- 5. Upsert vote helpful (toggle)
INSERT INTO qna_faq_helpful_votes (faq_id, user_id, is_helpful)
VALUES ($1, $2, $3)
ON CONFLICT (faq_id, user_id)
DO UPDATE SET is_helpful = EXCLUDED.is_helpful, updated_at = NOW();

-- Update counter di faq_items setelah vote
UPDATE qna_faq_items
SET
  helpful_count     = (SELECT COUNT(*) FROM qna_faq_helpful_votes WHERE faq_id = $1 AND is_helpful = true),
  not_helpful_count = (SELECT COUNT(*) FROM qna_faq_helpful_votes WHERE faq_id = $1 AND is_helpful = false),
  updated_at = NOW()
WHERE id = $1;


-- 6. Riwayat pertanyaan member (My Questions)
SELECT q.id, q.question_text, q.status, q.created_at,
       c.name AS category_name,
       a.answer_text, a.created_at AS answered_at
FROM qna_questions q
LEFT JOIN qna_categories c ON c.id = q.category_id
LEFT JOIN qna_answers a ON a.question_id = q.id AND a.is_primary = true
WHERE q.user_id = $1
ORDER BY q.created_at DESC
LIMIT $2 OFFSET $3;


-- 7. Backoffice: pertanyaan pending (antrian admin)
SELECT q.id, q.question_text, q.created_at,
       u.name AS user_name, u.email AS user_email,
       c.name AS category_name
FROM qna_questions q
JOIN users u ON u.id = q.user_id
LEFT JOIN qna_categories c ON c.id = q.category_id
WHERE q.status = 'pending'
ORDER BY q.created_at ASC;


-- 8. Convert jawaban ke FAQ baru (atomic)
BEGIN;
  -- Insert FAQ baru
  INSERT INTO qna_faq_items (category_id, question, answer, status, created_by, published_at)
  SELECT q.category_id, q.question_text, a.answer_text, 'published', a.answered_by, NOW()
  FROM qna_answers a
  JOIN qna_questions q ON q.id = a.question_id
  WHERE a.id = $1
  RETURNING id INTO v_faq_id;

  -- Link answer ke FAQ baru
  UPDATE qna_answers SET converted_to_faq_id = v_faq_id WHERE id = $1;

  -- Update status question jadi converted
  UPDATE qna_questions SET status = 'converted' WHERE id = (
    SELECT question_id FROM qna_answers WHERE id = $1
  );
COMMIT;
```

---

*Database Schema untuk QnA / FAQ Module. Stack: Golang + PostgreSQL 13+.*
