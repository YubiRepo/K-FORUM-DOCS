# Database Schema — Q&A / FAQ Module

> **Stack:** Golang + PostgreSQL  
> **Dibuat:** 2026-05-30 · **Revisi:** 2026-06-25 (forum komunitas: jawab antar-member, accepted answer, assignment, moderasi jawaban konfigurable)
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
  └── qna_questions          (pertanyaan yang diajukan member; public/private)
  └── qna_answers            (jawaban dari member / expert / admin)
  └── qna_answer_votes       (upvote jawaban di forum publik)
  └── qna_faq_helpful_votes  (vote helpful/not-helpful per FAQ item)

qna_categories
  └── qna_faq_items          (FAQ statis yang dikurasi admin)
  └── qna_questions          (kategori pertanyaan member)

qna_faq_items
  └── qna_faq_helpful_votes  (vote tracking per item)

qna_questions
  ├── qna_answers            (banyak jawaban per pertanyaan; >=0 ditandai accepted)
  ├── assigned_to → users    (expert/staff penanggung jawab — public & private)
  └── approved_by → users    (admin yang approve pertanyaan publik)

qna_answers
  ├── qna_answer_votes       (upvote per jawaban)
  ├── accepted_by → users    (expert/admin yang menandai jawaban valid/rujukan)
  └── converted_to_faq_id    (link bila jawaban dijadikan FAQ)

qna_bot_config               (konfigurasi bot fallback + mode moderasi jawaban, singleton)
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

  -- Visibility (dipilih penanya saat submit)
  -- private: hanya penanya + admin/expert yang ditugaskan (model 1-on-1, tidak masuk forum)
  -- public:  masuk forum komunitas, member lain bisa lihat & jawab (setelah di-approve)
  visibility       VARCHAR(10)  NOT NULL DEFAULT 'private',

  -- Status
  -- pending:   menunggu moderasi admin (belum tampil di forum / belum dijawab)
  -- approved:  disetujui & tampil publik di forum (hanya untuk visibility=public)
  -- answered:  sudah dijawab (private: oleh admin/expert; tetap dipakai utk alur private)
  -- rejected:  ditolak (duplikat / tidak relevan / melanggar)
  -- converted: dijadikan FAQ publik
  -- closed:    ditutup, tidak menerima jawaban baru lagi (opsional, oleh admin)
  status           VARCHAR(20)  NOT NULL DEFAULT 'pending',

  -- Assignment (penanggung jawab — berlaku utk public & private)
  assigned_to      UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  assigned_by      UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  assigned_at      TIMESTAMPTZ  NULL,

  -- Counters (denormalized untuk listing forum)
  answer_count     INTEGER      NOT NULL DEFAULT 0,   -- jumlah jawaban yang sudah tampil (approved/auto)
  accepted_count   INTEGER      NOT NULL DEFAULT 0,   -- jumlah jawaban yang ditandai valid/rujukan
  view_count       INTEGER      NOT NULL DEFAULT 0,

  -- Bot fallback
  bot_responded    BOOLEAN      NOT NULL DEFAULT false,
  bot_response     TEXT         NULL,

  -- Rejection
  rejection_reason TEXT         NULL,
  rejected_by      UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  rejected_at      TIMESTAMPTZ  NULL,

  -- Approval (untuk public yang lolos moderasi)
  approved_by      UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  approved_at      TIMESTAMPTZ  NULL,

  -- Timestamps
  answered_at      TIMESTAMPTZ  NULL,   -- diisi saat jawaban pertama tampil / accepted pertama
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

-- Forum publik: pertanyaan approved, terbaru dulu
CREATE INDEX idx_qna_questions_public_feed
  ON qna_questions (category_id, created_at DESC)
  WHERE visibility = 'public' AND status IN ('approved', 'converted', 'closed');

-- Antrian per penanggung jawab (assigned)
CREATE INDEX idx_qna_questions_assigned
  ON qna_questions (assigned_to, status, created_at DESC)
  WHERE assigned_to IS NOT NULL;


-- ============================================================================
-- 5. QNA_ANSWERS
-- Jawaban dari superadmin untuk pertanyaan member.
-- Banyak jawaban per pertanyaan (member/expert/admin); >=0 ditandai accepted.
-- ============================================================================

CREATE TABLE qna_answers (
  id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id     UUID         NOT NULL REFERENCES qna_questions(id) ON DELETE CASCADE,
  answered_by     UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  -- Tipe penjawab (resolved dari role/permission saat menjawab, disimpan utk display)
  -- member: jawaban dari sesama member (butuh benefit answer_qna)
  -- expert: jawaban dari user yang di-assign sbg penanggung jawab pertanyaan ini
  -- admin:  jawaban dari admin/superadmin
  answerer_type   VARCHAR(10)  NOT NULL DEFAULT 'member',

  -- Content
  answer_text     TEXT         NOT NULL,
  attachment_urls TEXT[]       NULL,

  -- Moderasi jawaban (mode diatur via qna_config.answer_moderation_mode)
  -- visible:  tampil ke publik (auto-mode lolos word filter, atau sudah di-approve)
  -- pending:  menunggu approve admin/expert (manual-mode)
  -- rejected: ditolak moderator
  -- hidden:   disembunyikan setelah tampil (mis. hasil report)
  status          VARCHAR(10)  NOT NULL DEFAULT 'visible',

  -- Accepted / rujukan valid (boleh lebih dari satu per pertanyaan)
  -- Ditandai oleh assigned expert / admin / superadmin — BUKAN penanya.
  is_accepted     BOOLEAN      NOT NULL DEFAULT false,
  accepted_by     UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  accepted_at     TIMESTAMPTZ  NULL,

  -- Upvote (denormalized counter; detail di qna_answer_votes)
  upvote_count    INTEGER      NOT NULL DEFAULT 0,

  -- Visibility privat (dipertahankan utk alur private 1-on-1)
  is_public       BOOLEAN      NOT NULL DEFAULT true,

  -- Convert to FAQ
  converted_to_faq_id UUID     NULL REFERENCES qna_faq_items(id) ON DELETE SET NULL,

  -- Moderation audit
  moderated_by    UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  moderated_at    TIMESTAMPTZ  NULL,
  reject_reason   TEXT         NULL,

  -- Timestamps
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Jawaban per pertanyaan, urut accepted dulu lalu upvote terbanyak
CREATE INDEX idx_qna_answers_question
  ON qna_answers (question_id, is_accepted DESC, upvote_count DESC, created_at ASC);

-- Antrian moderasi jawaban (manual mode)
CREATE INDEX idx_qna_answers_pending
  ON qna_answers (question_id, created_at ASC)
  WHERE status = 'pending';

CREATE INDEX idx_qna_answers_answered_by
  ON qna_answers (answered_by);

-- Satu user hanya boleh punya satu jawaban per pertanyaan (cegah spam)
CREATE UNIQUE INDEX idx_qna_answers_user_question_unique
  ON qna_answers (question_id, answered_by);


-- ============================================================================
-- 6. QNA_ANSWER_VOTES
-- Upvote jawaban di pertanyaan publik. Upvote-only, satu per user per jawaban.
-- ============================================================================

CREATE TABLE qna_answer_votes (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  answer_id  UUID        NOT NULL REFERENCES qna_answers(id) ON DELETE CASCADE,
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Satu user hanya bisa upvote sekali per jawaban (toggle via delete)
CREATE UNIQUE INDEX idx_qna_answer_votes_unique
  ON qna_answer_votes (answer_id, user_id);

CREATE INDEX idx_qna_answer_votes_answer
  ON qna_answer_votes (answer_id);


-- ============================================================================
-- 7. QNA_BOT_CONFIG
-- Konfigurasi modul QnA (singleton — hanya satu row).
-- Mencakup bot fallback + mode moderasi jawaban forum.
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

  -- Mode moderasi jawaban member di pertanyaan publik (fleksibel, diatur superadmin)
  -- auto:   jawaban lolos word filter (System Settings) langsung tampil (status=visible)
  -- manual: jawaban masuk status=pending, perlu approve admin/expert baru tampil
  answer_moderation_mode VARCHAR(10) NOT NULL DEFAULT 'manual',

  -- Mode moderasi pertanyaan publik baru
  -- Selalu pending → approve admin (tidak dikonfigurasi, hardcoded di rules).
  -- Disediakan di sini bila kelak ingin dibuat fleksibel.
  question_moderation_mode VARCHAR(10) NOT NULL DEFAULT 'manual',

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
    Visibility     string         `db:"visibility"      json:"visibility"`
    Status         string         `db:"status"          json:"status"`
    AssignedTo     *string        `db:"assigned_to"     json:"assigned_to,omitempty"`
    AssignedBy     *string        `db:"assigned_by"     json:"assigned_by,omitempty"`
    AssignedAt     *time.Time     `db:"assigned_at"     json:"assigned_at,omitempty"`
    AnswerCount    int            `db:"answer_count"    json:"answer_count"`
    AcceptedCount  int            `db:"accepted_count"  json:"accepted_count"`
    ViewCount      int            `db:"view_count"      json:"view_count"`
    BotResponded   bool           `db:"bot_responded"   json:"bot_responded"`
    BotResponse    *string        `db:"bot_response"    json:"bot_response,omitempty"`
    RejectionReason *string       `db:"rejection_reason" json:"rejection_reason,omitempty"`
    RejectedBy     *string        `db:"rejected_by"     json:"rejected_by,omitempty"`
    RejectedAt     *time.Time     `db:"rejected_at"     json:"rejected_at,omitempty"`
    ApprovedBy     *string        `db:"approved_by"     json:"approved_by,omitempty"`
    ApprovedAt     *time.Time     `db:"approved_at"     json:"approved_at,omitempty"`
    AnsweredAt     *time.Time     `db:"answered_at"     json:"answered_at,omitempty"`
    CreatedAt      time.Time      `db:"created_at"      json:"created_at"`
    UpdatedAt      time.Time      `db:"updated_at"      json:"updated_at"`
}

const (
    QuestionVisibilityPrivate = "private"
    QuestionVisibilityPublic  = "public"

    QuestionStatusPending   = "pending"
    QuestionStatusApproved  = "approved"
    QuestionStatusAnswered  = "answered"
    QuestionStatusRejected  = "rejected"
    QuestionStatusConverted = "converted"
    QuestionStatusClosed    = "closed"
)

// QnaAnswer — Jawaban untuk pertanyaan member (dari member / expert / admin)
type QnaAnswer struct {
    ID               string         `db:"id"                  json:"id"`
    QuestionID       string         `db:"question_id"         json:"question_id"`
    AnsweredBy       string         `db:"answered_by"         json:"answered_by"`
    AnswererType     string         `db:"answerer_type"       json:"answerer_type"`
    AnswerText       string         `db:"answer_text"         json:"answer_text"`
    AttachmentURLs   pq.StringArray `db:"attachment_urls"     json:"attachment_urls,omitempty"`
    Status           string         `db:"status"              json:"status"`
    IsAccepted       bool           `db:"is_accepted"         json:"is_accepted"`
    AcceptedBy       *string        `db:"accepted_by"         json:"accepted_by,omitempty"`
    AcceptedAt       *time.Time     `db:"accepted_at"         json:"accepted_at,omitempty"`
    UpvoteCount      int            `db:"upvote_count"        json:"upvote_count"`
    IsPublic         bool           `db:"is_public"           json:"is_public"`
    ConvertedToFaqID *string        `db:"converted_to_faq_id" json:"converted_to_faq_id,omitempty"`
    ModeratedBy      *string        `db:"moderated_by"        json:"moderated_by,omitempty"`
    ModeratedAt      *time.Time     `db:"moderated_at"        json:"moderated_at,omitempty"`
    RejectReason     *string        `db:"reject_reason"       json:"reject_reason,omitempty"`
    CreatedAt        time.Time      `db:"created_at"          json:"created_at"`
    UpdatedAt        time.Time      `db:"updated_at"          json:"updated_at"`
}

const (
    AnswererTypeMember = "member"
    AnswererTypeExpert = "expert"
    AnswererTypeAdmin  = "admin"

    AnswerStatusVisible  = "visible"
    AnswerStatusPending  = "pending"
    AnswerStatusRejected = "rejected"
    AnswerStatusHidden   = "hidden"
)

// QnaAnswerVote — Upvote jawaban (upvote-only, satu per user per jawaban)
type QnaAnswerVote struct {
    ID        string    `db:"id"         json:"id"`
    AnswerID  string    `db:"answer_id"  json:"answer_id"`
    UserID    string    `db:"user_id"    json:"user_id"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// QnaBotConfig — Konfigurasi bot fallback + moderasi (singleton)
type QnaBotConfig struct {
    ID                     string    `db:"id"                       json:"id"`
    Enabled                bool      `db:"enabled"                  json:"enabled"`
    FallbackMessage        string    `db:"fallback_message"         json:"fallback_message"`
    SuggestContact         bool      `db:"suggest_contact"          json:"suggest_contact"`
    ContactInfo            *string   `db:"contact_info"             json:"contact_info,omitempty"`
    SimilarityThreshold    float64   `db:"similarity_threshold"     json:"similarity_threshold"`
    AnswerModerationMode   string    `db:"answer_moderation_mode"   json:"answer_moderation_mode"`
    QuestionModerationMode string    `db:"question_moderation_mode" json:"question_moderation_mode"`
    UpdatedBy              *string   `db:"updated_by"               json:"updated_by,omitempty"`
    UpdatedAt              time.Time `db:"updated_at"               json:"updated_at"`
}

const (
    ModerationModeAuto   = "auto"
    ModerationModeManual = "manual"
)
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
  visibility       VARCHAR(10) NOT NULL DEFAULT 'private',
  status           VARCHAR(20) NOT NULL DEFAULT 'pending',
  assigned_to      UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  assigned_by      UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  assigned_at      TIMESTAMPTZ NULL,
  answer_count     INTEGER     NOT NULL DEFAULT 0,
  accepted_count   INTEGER     NOT NULL DEFAULT 0,
  view_count       INTEGER     NOT NULL DEFAULT 0,
  bot_responded    BOOLEAN     NOT NULL DEFAULT false,
  bot_response     TEXT        NULL,
  rejection_reason TEXT        NULL,
  rejected_by      UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  rejected_at      TIMESTAMPTZ NULL,
  approved_by      UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  approved_at      TIMESTAMPTZ NULL,
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
CREATE INDEX IF NOT EXISTS idx_qna_questions_public_feed
  ON qna_questions (category_id, created_at DESC)
  WHERE visibility = 'public' AND status IN ('approved', 'converted', 'closed');
CREATE INDEX IF NOT EXISTS idx_qna_questions_assigned
  ON qna_questions (assigned_to, status, created_at DESC)
  WHERE assigned_to IS NOT NULL;


-- ============================================================================
-- Migration 5: Create qna_answers
-- ============================================================================
CREATE TABLE IF NOT EXISTS qna_answers (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  question_id       UUID        NOT NULL REFERENCES qna_questions(id) ON DELETE CASCADE,
  answered_by       UUID        NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  answerer_type     VARCHAR(10) NOT NULL DEFAULT 'member',
  answer_text       TEXT        NOT NULL,
  attachment_urls   TEXT[]      NULL,
  status            VARCHAR(10) NOT NULL DEFAULT 'visible',
  is_accepted       BOOLEAN     NOT NULL DEFAULT false,
  accepted_by       UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  accepted_at       TIMESTAMPTZ NULL,
  upvote_count      INTEGER     NOT NULL DEFAULT 0,
  is_public         BOOLEAN     NOT NULL DEFAULT true,
  converted_to_faq_id UUID      NULL REFERENCES qna_faq_items(id) ON DELETE SET NULL,
  moderated_by      UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  moderated_at      TIMESTAMPTZ NULL,
  reject_reason     TEXT        NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_qna_answers_question
  ON qna_answers (question_id, is_accepted DESC, upvote_count DESC, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_qna_answers_pending
  ON qna_answers (question_id, created_at ASC) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_qna_answers_answered_by
  ON qna_answers (answered_by);
CREATE UNIQUE INDEX IF NOT EXISTS idx_qna_answers_user_question_unique
  ON qna_answers (question_id, answered_by);


-- ============================================================================
-- Migration 5b: Create qna_answer_votes
-- ============================================================================
CREATE TABLE IF NOT EXISTS qna_answer_votes (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  answer_id  UUID        NOT NULL REFERENCES qna_answers(id) ON DELETE CASCADE,
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_qna_answer_votes_unique
  ON qna_answer_votes (answer_id, user_id);
CREATE INDEX IF NOT EXISTS idx_qna_answer_votes_answer
  ON qna_answer_votes (answer_id);


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
  answer_moderation_mode   VARCHAR(10) NOT NULL DEFAULT 'manual',
  question_moderation_mode VARCHAR(10) NOT NULL DEFAULT 'manual',
  updated_by           UUID           NULL REFERENCES users(id) ON DELETE SET NULL,
  updated_at           TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_qna_bot_config_singleton
  ON qna_bot_config ((true));

-- Seed default config
INSERT INTO qna_bot_config (enabled, fallback_message, suggest_contact, similarity_threshold, answer_moderation_mode, question_moderation_mode)
VALUES (
  true,
  'Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja.',
  true,
  0.60,
  'manual',
  'manual'
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

-- Note: qna_answer_votes tidak punya updated_at (upvote bersifat insert/delete saja).


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

### `qna_questions.visibility`

| Value | Keterangan |
|-------|-----------|
| `private` | Hanya penanya + admin/expert yang ditugaskan. Tidak masuk forum. |
| `public` | Masuk forum komunitas, member lain bisa lihat & jawab (setelah approved) |

### `qna_questions.status`

| Value | Keterangan |
|-------|-----------|
| `pending` | Menunggu moderasi admin (belum tampil / belum dijawab) |
| `approved` | Disetujui & tampil publik di forum (khusus `visibility=public`) |
| `answered` | Sudah dijawab (terutama dipakai utk alur private) |
| `rejected` | Ditolak (duplikat / tidak relevan / melanggar) |
| `converted` | Jawaban dijadikan FAQ publik |
| `closed` | Ditutup, tidak menerima jawaban baru |

### `qna_answers.status`

| Value | Keterangan |
|-------|-----------|
| `visible` | Tampil ke publik (auto-mode lolos filter, atau sudah di-approve) |
| `pending` | Menunggu approve admin/expert (manual-mode) |
| `rejected` | Ditolak moderator |
| `hidden` | Disembunyikan setelah sempat tampil (mis. hasil report) |

### `qna_answers.answerer_type`

| Value | Keterangan |
|-------|-----------|
| `member` | Jawaban dari sesama member (butuh benefit `answer_qna`) |
| `expert` | Jawaban dari user yang di-assign sbg penanggung jawab |
| `admin` | Jawaban dari admin / superadmin |

### `qna_bot_config.answer_moderation_mode` / `question_moderation_mode`

| Value | Keterangan |
|-------|-----------|
| `auto` | Jawaban lolos word filter (System Settings) langsung `visible` |
| `manual` | Jawaban masuk `pending`, perlu approve admin/expert |

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
SELECT q.id, q.question_text, q.status, q.visibility, q.created_at,
       c.name AS category_name,
       q.answer_count, q.accepted_count
FROM qna_questions q
LEFT JOIN qna_categories c ON c.id = q.category_id
WHERE q.user_id = $1
ORDER BY q.created_at DESC
LIMIT $2 OFFSET $3;


-- 6b. Forum publik: daftar pertanyaan approved per kategori
SELECT q.id, q.question_text, q.answer_count, q.accepted_count, q.view_count,
       q.created_at, c.name AS category_name,
       u.name AS asker_name
FROM qna_questions q
JOIN users u ON u.id = q.user_id
LEFT JOIN qna_categories c ON c.id = q.category_id
WHERE q.visibility = 'public'
  AND q.status IN ('approved', 'converted', 'closed')
  AND ($1::uuid IS NULL OR q.category_id = $1)
ORDER BY q.created_at DESC
LIMIT $2 OFFSET $3;


-- 6c. Jawaban yang tampil untuk satu pertanyaan (accepted dulu, lalu upvote)
SELECT a.id, a.answer_text, a.answerer_type, a.is_accepted, a.upvote_count,
       a.created_at, u.name AS answerer_name,
       EXISTS (
         SELECT 1 FROM qna_answer_votes v
         WHERE v.answer_id = a.id AND v.user_id = $2
       ) AS user_upvoted
FROM qna_answers a
JOIN users u ON u.id = a.answered_by
WHERE a.question_id = $1
  AND a.status = 'visible'
ORDER BY a.is_accepted DESC, a.upvote_count DESC, a.created_at ASC;


-- 6d. Upvote jawaban (toggle): insert jika belum, hapus jika sudah
-- Insert:
INSERT INTO qna_answer_votes (answer_id, user_id)
VALUES ($1, $2)
ON CONFLICT (answer_id, user_id) DO NOTHING;
-- Sinkron counter:
UPDATE qna_answers
SET upvote_count = (SELECT COUNT(*) FROM qna_answer_votes WHERE answer_id = $1),
    updated_at = NOW()
WHERE id = $1;


-- 6e. Tandai jawaban sebagai accepted/rujukan (oleh expert/admin)
UPDATE qna_answers
SET is_accepted = true, accepted_by = $2, accepted_at = NOW(), updated_at = NOW()
WHERE id = $1;
-- Sinkron counter accepted di question:
UPDATE qna_questions
SET accepted_count = (SELECT COUNT(*) FROM qna_answers WHERE question_id = $3 AND is_accepted = true),
    updated_at = NOW()
WHERE id = $3;


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
