# Database Schema — News Module

> **Stack:** Golang + PostgreSQL  
> **Dibuat:** 2026-06-04  
> **Module:** News

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [PostgreSQL DDL Schema](#postgresql-ddl-schema)
3. [Enum Values](#enum-values)
4. [Query Examples](#query-examples)

---

## Overview Relasi

```
users
  └── articles              (created_by, published_by_user_id)
  └── article_comments      (user_id)
  └── article_likes         (user_id)
  └── article_bookmarks     (user_id)
  └── article_views         (user_id)
  └── news_sources          (created_by)
  └── source_categories     (updated_by)

news_categories
  └── articles              (category_id)
  └── source_categories     (category_id — mapping hasil scrape → kategori KAI)

news_scopes
  └── articles              (news_scope_id)
  └── news_sources          (default_scope_id)

news_sources
  └── source_selectors      (source_id — 1:1)
  └── source_categories     (source_id — 1:N)
  └── articles              (source_id — null jika manual)

articles
  └── article_translations  (article_id — 1:N, satu per bahasa)
  └── article_comments      (article_id)
  └── article_likes         (article_id)
  └── article_bookmarks     (article_id)
  └── article_views         (article_id)

article_comments
  └── article_comments      (parent_id — self-referential, max 2 level)

system_languages            (master bahasa aktif di platform)
news_system_settings        (singleton — config global translation)
```

---

## PostgreSQL DDL Schema

```sql
-- ============================================================================
-- NEWS MODULE DATABASE SCHEMA
-- Stack: PostgreSQL 13+
-- Created: 2026-06-04
-- ============================================================================


-- ============================================================================
-- 1. SYSTEM_LANGUAGES
-- Master bahasa yang aktif di platform. Dikelola oleh Usergod.
-- ============================================================================

CREATE TABLE system_languages (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  code                 VARCHAR(10)  NOT NULL UNIQUE,  -- 'id', 'en', 'ko', 'jp'
  name                 VARCHAR(100) NOT NULL,          -- 'Indonesian', 'English', 'Korean'
  is_ui_language       BOOLEAN      NOT NULL DEFAULT false,
  is_translate_target  BOOLEAN      NOT NULL DEFAULT false,
  is_active            BOOLEAN      NOT NULL DEFAULT true,
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_system_languages_code
  ON system_languages (code);

CREATE INDEX idx_system_languages_active
  ON system_languages (is_active, is_translate_target)
  WHERE is_active = true;

-- Seed default languages
INSERT INTO system_languages (code, name, is_ui_language, is_translate_target, is_active) VALUES
  ('id', 'Indonesian', true,  true,  true),
  ('en', 'English',    true,  true,  true),
  ('ko', 'Korean',     true,  true,  true)
ON CONFLICT (code) DO NOTHING;


-- ============================================================================
-- 2. NEWS_SYSTEM_SETTINGS
-- Konfigurasi global sistem News. Singleton (hanya 1 baris).
-- Dikelola oleh Usergod dan Superadmin sesuai permission.
-- ============================================================================

CREATE TABLE news_system_settings (
  id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  translation_enabled  BOOLEAN     NOT NULL DEFAULT true,
  -- Master toggle fitur translation.
  -- false = semua user hanya dapat bahasa original artikel.

  on_demand_enabled    BOOLEAN     NOT NULL DEFAULT true,
  -- Aktif jika translation_enabled = true.
  -- true  = auto-trigger translate saat user hit artikel dalam bahasa yang belum ada.
  -- false = user harus eksplisit pilih bahasa, baru translate di-generate.

  updated_by           UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_news_system_settings_singleton
  ON news_system_settings ((true));

-- Seed default settings
INSERT INTO news_system_settings (translation_enabled, on_demand_enabled)
VALUES (true, true)
ON CONFLICT DO NOTHING;


-- ============================================================================
-- 2b. NEWS_SCOPES
-- Klasifikasi asal/fokus geografis berita (dimensi terpisah dari kategori topikal).
-- Contoh: Berita Indonesia, Berita Korea, Berita Korea di Indonesia.
-- Dikelola oleh Superadmin. Phase 1 = master kecil & manual; FK agar mudah
-- menambah scope baru tanpa refactor (mis. 'bilateral', 'asean').
-- ============================================================================

CREATE TABLE news_scopes (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100) NOT NULL,        -- 'Berita Indonesia', 'Berita Korea'
  slug        VARCHAR(120) NOT NULL UNIQUE, -- 'indonesia', 'korea', 'korea_indonesia'
  is_active   BOOLEAN      NOT NULL DEFAULT true,
  sort_order  INTEGER      NOT NULL DEFAULT 0,
  created_by  UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by  UUID         NULL     REFERENCES users(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_news_scopes_slug
  ON news_scopes (slug);

CREATE INDEX idx_news_scopes_active_order
  ON news_scopes (is_active, sort_order)
  WHERE is_active = true;

-- Seed default scopes (created_by harus di-set saat insert)
-- INSERT INTO news_scopes (name, slug, sort_order, created_by) VALUES
--   ('Berita Indonesia',           'indonesia',       1, :usergod_id),
--   ('Berita Korea',               'korea',           2, :usergod_id),
--   ('Berita Korea di Indonesia',  'korea_indonesia', 3, :usergod_id);


-- ============================================================================
-- 3. NEWS_CATEGORIES
-- Kategori berita (Politik, Ekonomi, Olahraga, dll).
-- Dikelola oleh Superadmin.
-- ============================================================================

CREATE TABLE news_categories (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100) NOT NULL,
  slug        VARCHAR(120) NOT NULL UNIQUE,
  is_active   BOOLEAN      NOT NULL DEFAULT true,
  sort_order  INTEGER      NOT NULL DEFAULT 0,
  created_by  UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by  UUID         NULL     REFERENCES users(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_news_categories_slug
  ON news_categories (slug);

CREATE INDEX idx_news_categories_active_order
  ON news_categories (is_active, sort_order)
  WHERE is_active = true;


-- ============================================================================
-- 4. NEWS_SOURCES
-- Sumber berita eksternal untuk scraping.
-- Didaftarkan oleh Usergod. Config operasional oleh Usergod & Superadmin.
-- ============================================================================

CREATE TABLE news_sources (
  id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  key               VARCHAR(50)  NOT NULL UNIQUE,   -- 'detik', 'kompas', 'kai_official'
  name              VARCHAR(100) NOT NULL,
  base_url          VARCHAR(500) NOT NULL,
  original_language VARCHAR(10)  NOT NULL REFERENCES system_languages(code) ON DELETE RESTRICT,

  -- Default scope: diwariskan ke artikel hasil scraping dari source ini.
  -- Editor masih bisa override per-artikel via articles.news_scope_id.
  default_scope_id  UUID         NULL REFERENCES news_scopes(id) ON DELETE SET NULL,

  -- Scheduling (cron expression)
  schedule          VARCHAR(100) NOT NULL DEFAULT '0 */6 * * *',
  -- Contoh: '*/30 * * * *' = tiap 30 menit
  --         '0 * * * *'    = tiap jam
  --         '0 8,12,17 * * *' = jam 8, 12, 17 setiap hari
  last_scraped_at   TIMESTAMPTZ  NULL,

  -- Publish & content config
  auto_publish      BOOLEAN      NOT NULL DEFAULT false,
  -- true = hasil scraping langsung published, false = masuk draft

  ai_cleanup        BOOLEAN      NOT NULL DEFAULT false,
  -- true = jalankan AI cleanup setelah scraping untuk rapikan konten HTML

  auto_translate    BOOLEAN      NOT NULL DEFAULT false,
  -- true = generate translation ke semua bahasa aktif (is_translate_target=true)
  --        setelah scraping selesai

  is_active         BOOLEAN      NOT NULL DEFAULT true,
  created_by        UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by        UUID         NULL     REFERENCES users(id) ON DELETE SET NULL,
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_news_sources_key
  ON news_sources (key);

CREATE INDEX idx_news_sources_active
  ON news_sources (is_active, last_scraped_at)
  WHERE is_active = true;


-- ============================================================================
-- 5. SOURCE_SELECTORS
-- Konfigurasi CSS selector untuk parsing HTML per source.
-- Didaftarkan oleh Usergod. Relasi 1:1 dengan news_sources.
-- ============================================================================

CREATE TABLE source_selectors (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id         UUID        NOT NULL UNIQUE REFERENCES news_sources(id) ON DELETE CASCADE,
  content_selector  TEXT        NULL,       -- CSS selector fallback. Konten utama diekstrak via Readability; ini hanya dipakai jika Readability gagal
  author_selector   TEXT        NULL,       -- CSS selector untuk nama penulis
  tags_selector     TEXT        NULL,       -- CSS selector untuk tag artikel
  extra_fields      JSONB       NULL,       -- Edge case per source: aturan multi-page, override khusus (mis. Okezone, Tribun ?page=all)
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_source_selectors_source
  ON source_selectors (source_id);


-- ============================================================================
-- 6. SOURCE_CATEGORIES
-- Definisi kategori yang di-scrape per source + batas artikel per fetch.
-- Diatur oleh Superadmin.
-- ============================================================================

CREATE TABLE source_categories (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id     UUID         NOT NULL REFERENCES news_sources(id) ON DELETE CASCADE,
  category_key  VARCHAR(100) NOT NULL,     -- input manual: slug/path feed dari source ('sepakbola.xml', 'tekno')
  category_id   UUID         NULL REFERENCES news_categories(id) ON DELETE SET NULL,
  -- Mapping: hasil scrape dari category_key ini di-assign ke kategori KAI mana.
  -- Diisi articles.category_id saat scraping. NULL = belum dipetakan (fallback 'umum').
  url_suffix    VARCHAR(500) NULL,         -- mis. '/sport' — append ke base_url
  url_override  VARCHAR(500) NULL,         -- URL penuh override jika beda dari pola base_url
  article_limit INTEGER      NOT NULL DEFAULT 10,
  is_active     BOOLEAN      NOT NULL DEFAULT true,
  updated_by    UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT uq_source_category UNIQUE (source_id, category_key)
);

CREATE INDEX idx_source_categories_source_active
  ON source_categories (source_id, is_active)
  WHERE is_active = true;


-- ============================================================================
-- 7. ARTICLES
-- Entitas utama artikel berita.
-- ============================================================================

CREATE TABLE articles (
  id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Asal artikel
  source_id             UUID         NULL REFERENCES news_sources(id) ON DELETE SET NULL,
  -- null = artikel manual
  category_id           UUID         NULL REFERENCES news_categories(id) ON DELETE SET NULL,
  news_scope_id         UUID         NULL REFERENCES news_scopes(id) ON DELETE SET NULL,
  -- Asal/fokus geografis (indonesia / korea / korea_indonesia).
  -- Default diwariskan dari news_sources.default_scope_id saat scraping; editor bisa override.
  original_language     VARCHAR(10)  NOT NULL REFERENCES system_languages(code) ON DELETE RESTRICT,
  is_manual             BOOLEAN      NOT NULL DEFAULT false,
  original_url          VARCHAR(1000) NULL UNIQUE,
  -- URL asli dari scraping — UNIQUE untuk de-duplikasi. null jika manual.

  -- Status
  -- draft            : belum tayang
  -- pending_approval : menunggu review (khusus artikel Member Pro)
  -- published        : tayang ke semua user
  -- archived         : tidak muncul di listing, tetap tersimpan
  -- rejected         : ditolak (khusus artikel Member Pro)
  status                VARCHAR(20)  NOT NULL DEFAULT 'draft',

  -- Label & publisher
  author_label          VARCHAR(100) NOT NULL,
  -- 'Korean Association Indonesia' untuk KAI Pusat/scraping
  -- 'KAI Jakarta', 'KAI Bandung', dll untuk region
  author_region_id      UUID         NULL,
  -- FK ke regions (null jika dari KAI Pusat)

  -- Stats (denormalized untuk performa)
  view_count            INTEGER      NOT NULL DEFAULT 0,
  unique_view_count     INTEGER      NOT NULL DEFAULT 0,
  like_count            INTEGER      NOT NULL DEFAULT 0,
  comment_count         INTEGER      NOT NULL DEFAULT 0,

  -- Featured
  is_featured           BOOLEAN      NOT NULL DEFAULT false,

  -- Audit
  created_by            UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  -- null jika murni hasil scraping otomatis (tanpa intervensi editor)

  published_by_user_id  UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  -- Siapa yang men-publish. null jika auto_publish oleh sistem. Internal/audit only.

  published_by_label    VARCHAR(200) NULL,
  -- Label yang tampil ke publik. Sama dengan author_label pada umumnya.
  -- Diisi saat artikel published.

  -- Timestamps
  published_at          TIMESTAMPTZ  NULL,
  archived_at           TIMESTAMPTZ  NULL,
  created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE UNIQUE INDEX idx_articles_original_url
  ON articles (original_url)
  WHERE original_url IS NOT NULL;

CREATE INDEX idx_articles_published
  ON articles (status, published_at DESC)
  WHERE status = 'published';

CREATE INDEX idx_articles_source_status
  ON articles (source_id, status, created_at DESC);

CREATE INDEX idx_articles_category_published
  ON articles (category_id, status, published_at DESC)
  WHERE status = 'published';

CREATE INDEX idx_articles_scope_published
  ON articles (news_scope_id, status, published_at DESC)
  WHERE status = 'published';

CREATE INDEX idx_articles_featured
  ON articles (is_featured, published_at DESC)
  WHERE is_featured = true AND status = 'published';

CREATE INDEX idx_articles_pending_approval
  ON articles (created_at DESC)
  WHERE status = 'pending_approval';

CREATE INDEX idx_articles_author_region
  ON articles (author_region_id, status, published_at DESC)
  WHERE status = 'published';


-- ============================================================================
-- 8. ARTICLE_TRANSLATIONS
-- Konten artikel per bahasa. Satu artikel = N translations (satu per bahasa).
-- ============================================================================

CREATE TABLE article_translations (
  id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id       UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  language         VARCHAR(10)  NOT NULL REFERENCES system_languages(code) ON DELETE RESTRICT,

  -- Content
  title            TEXT         NOT NULL,
  content          TEXT         NOT NULL,
  summary          TEXT         NULL,
  author           VARCHAR(200) NULL,
  thumbnail_url    TEXT         NULL,
  tags             TEXT[]       NULL,

  -- Translation metadata
  is_original      BOOLEAN      NOT NULL DEFAULT false,
  -- true  = konten asli (scraping hasil parse / editor tulis manual)
  -- false = hasil translate

  translate_status VARCHAR(20)  NULL,
  -- pending / processing / done / failed
  -- hanya relevan jika is_original = false

  translated_by    VARCHAR(50)  NULL,
  -- 'google', 'openai', atau null jika diisi manual

  created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT uq_article_language UNIQUE (article_id, language)
);

CREATE UNIQUE INDEX idx_article_translations_unique
  ON article_translations (article_id, language);

CREATE INDEX idx_article_translations_article
  ON article_translations (article_id, is_original);

CREATE INDEX idx_article_translations_pending
  ON article_translations (translate_status, created_at ASC)
  WHERE translate_status = 'pending';

-- GIN index untuk full-text search
CREATE INDEX idx_article_translations_fts
  ON article_translations
  USING GIN (to_tsvector('simple', title || ' ' || content))
  WHERE translate_status = 'done' OR is_original = true;


-- ============================================================================
-- 9. ARTICLE_VIEWS
-- Tracking unique view per user. Total hit disimpan di articles.view_count.
-- ============================================================================

CREATE TABLE article_views (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id  UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  viewed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT uq_article_view UNIQUE (article_id, user_id)
);

CREATE UNIQUE INDEX idx_article_views_unique
  ON article_views (article_id, user_id);

CREATE INDEX idx_article_views_article
  ON article_views (article_id);


-- ============================================================================
-- 10. ARTICLE_LIKES
-- Like per artikel per user. Satu user satu like per artikel.
-- ============================================================================

CREATE TABLE article_likes (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id  UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT uq_article_like UNIQUE (article_id, user_id)
);

CREATE UNIQUE INDEX idx_article_likes_unique
  ON article_likes (article_id, user_id);

CREATE INDEX idx_article_likes_article
  ON article_likes (article_id);

CREATE INDEX idx_article_likes_user
  ON article_likes (user_id, created_at DESC);


-- ============================================================================
-- 11. ARTICLE_BOOKMARKS
-- Bookmark/save artikel per user.
-- ============================================================================

CREATE TABLE article_bookmarks (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id  UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT uq_article_bookmark UNIQUE (article_id, user_id)
);

CREATE UNIQUE INDEX idx_article_bookmarks_unique
  ON article_bookmarks (article_id, user_id);

CREATE INDEX idx_article_bookmarks_user
  ON article_bookmarks (user_id, created_at DESC);


-- ============================================================================
-- 12. ARTICLE_COMMENTS
-- Komentar artikel. Threaded 2 level (comment + reply).
-- Soft delete: is_deleted = true, content = null.
-- ============================================================================

CREATE TABLE article_comments (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id  UUID         NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  parent_id   UUID         NULL REFERENCES article_comments(id) ON DELETE SET NULL,
  -- null  = comment level 1
  -- terisi = reply level 2 (selalu FK ke comment level 1, tidak bisa nested lebih dalam)

  content     TEXT         NULL,
  -- null jika is_deleted = true

  is_deleted  BOOLEAN      NOT NULL DEFAULT false,
  deleted_by  UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  -- siapa yang delete: pemilik comment, Editor, atau Superadmin

  deleted_at  TIMESTAMPTZ  NULL,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_article_comments_article
  ON article_comments (article_id, parent_id, created_at ASC)
  WHERE is_deleted = false;

CREATE INDEX idx_article_comments_parent
  ON article_comments (parent_id, created_at ASC)
  WHERE parent_id IS NOT NULL;

CREATE INDEX idx_article_comments_user
  ON article_comments (user_id, created_at DESC);
```

---

## Enum Values

### `articles.status`

| Value | Keterangan |
|-------|-----------|
| `draft` | Belum tayang. Dari scraping (auto_publish=false) atau manual yang belum dipublish |
| `pending_approval` | Menunggu review — khusus artikel dari Member Pro |
| `published` | Tayang ke semua user (global) |
| `archived` | Tidak muncul di listing utama, tetap tersimpan |
| `rejected` | Ditolak — khusus artikel dari Member Pro |

### `article_translations.translate_status`

| Value | Keterangan |
|-------|-----------|
| `pending` | Job belum diproses |
| `processing` | Job sedang berjalan |
| `done` | Translation selesai dan tersedia |
| `failed` | Translation gagal — bisa di-retry |

### `article_translations.translated_by`

| Value | Keterangan |
|-------|-----------|
| `google` | Ditranslate via Google Translate API |
| `openai` | Ditranslate via OpenAI |
| `null` | Diisi manual oleh editor |

---

## Query Examples

```sql
-- 1. List artikel published (dengan filter bahasa)
-- Backend cek article_translations untuk bahasa yang diminta
SELECT
  a.id,
  a.author_label,
  a.published_by_label,
  a.like_count,
  a.comment_count,
  a.view_count,
  a.published_at,
  a.is_featured,
  t.title,
  t.summary,
  t.thumbnail_url,
  t.language,
  t.is_original
FROM articles a
JOIN article_translations t ON t.article_id = a.id
  AND t.language = $1                        -- bahasa yang diminta (Accept-Language)
  AND (t.translate_status = 'done' OR t.is_original = true)
WHERE a.status = 'published'
ORDER BY a.is_featured DESC, a.published_at DESC
LIMIT $2 OFFSET $3;


-- 2. Fallback ke original jika translation tidak ada
SELECT
  a.id,
  t.title,
  t.language,
  t.is_original
FROM articles a
JOIN article_translations t ON t.article_id = a.id
WHERE a.id = $1
  AND (
    t.language = $2                          -- coba bahasa yang diminta dulu
    OR (t.is_original = true AND NOT EXISTS (
      SELECT 1 FROM article_translations
      WHERE article_id = $1
        AND language = $2
        AND translate_status = 'done'
    ))
  )
ORDER BY (t.language = $2) DESC
LIMIT 1;


-- 3. De-duplikasi scraping — cek original_url sebelum insert
SELECT id FROM articles
WHERE original_url = $1
LIMIT 1;
-- Jika ada hasil → skip, tidak insert


-- 4. Ambil pending translation jobs (untuk background worker)
SELECT
  at.id,
  at.article_id,
  at.language,
  src.title    AS source_title,
  src.content  AS source_content
FROM article_translations at
JOIN article_translations src ON src.article_id = at.article_id AND src.is_original = true
WHERE at.translate_status = 'pending'
ORDER BY at.created_at ASC
LIMIT 10;


-- 5. Toggle like artikel (insert or delete)
-- Insert like:
INSERT INTO article_likes (article_id, user_id)
VALUES ($1, $2)
ON CONFLICT (article_id, user_id) DO NOTHING;

-- Update counter:
UPDATE articles
SET like_count = (SELECT COUNT(*) FROM article_likes WHERE article_id = $1),
    updated_at = NOW()
WHERE id = $1;

-- Delete like (unlike):
DELETE FROM article_likes WHERE article_id = $1 AND user_id = $2;


-- 6. Increment view_count (total hit) — selalu
UPDATE articles
SET view_count = view_count + 1, updated_at = NOW()
WHERE id = $1;

-- Insert unique view (per user) — jika belum pernah
INSERT INTO article_views (article_id, user_id)
VALUES ($1, $2)
ON CONFLICT (article_id, user_id) DO NOTHING;

-- Jika insert berhasil (bukan conflict) → increment unique_view_count
UPDATE articles
SET unique_view_count = unique_view_count + 1, updated_at = NOW()
WHERE id = $1;


-- 7. Ambil comment artikel (threaded, level 1 + replies)
-- Level 1: parent_id IS NULL
SELECT
  c.id,
  c.content,
  c.is_deleted,
  c.created_at,
  u.id   AS user_id,
  u.name AS user_name
FROM article_comments c
JOIN users u ON u.id = c.user_id
WHERE c.article_id = $1
  AND c.parent_id IS NULL
ORDER BY c.created_at ASC
LIMIT $2 OFFSET $3;

-- Level 2: replies untuk comment tertentu
SELECT
  c.id,
  c.content,
  c.is_deleted,
  c.created_at,
  u.id   AS user_id,
  u.name AS user_name
FROM article_comments c
JOIN users u ON u.id = c.user_id
WHERE c.parent_id = $1
ORDER BY c.created_at ASC;


-- 8. Soft delete comment
UPDATE article_comments
SET
  is_deleted = true,
  content    = null,
  deleted_by = $2,
  deleted_at = NOW(),
  updated_at = NOW()
WHERE id = $1
  AND (user_id = $2 OR $3 = true);  -- $3 = is_moderator (Editor/Superadmin)

-- Decrement comment_count di artikel
UPDATE articles
SET comment_count = GREATEST(0, comment_count - 1), updated_at = NOW()
WHERE id = (SELECT article_id FROM article_comments WHERE id = $1);


-- 9. Bookmark toggle
INSERT INTO article_bookmarks (article_id, user_id)
VALUES ($1, $2)
ON CONFLICT (article_id, user_id) DO NOTHING;

-- Remove bookmark:
DELETE FROM article_bookmarks WHERE article_id = $1 AND user_id = $2;


-- 10. Backoffice: list artikel pending approval (Member Pro)
SELECT
  a.id,
  a.author_label,
  a.created_at,
  u.name  AS submitted_by_name,
  u.email AS submitted_by_email,
  t.title,
  t.language
FROM articles a
JOIN users u ON u.id = a.created_by
JOIN article_translations t ON t.article_id = a.id AND t.is_original = true
WHERE a.status = 'pending_approval'
ORDER BY a.created_at ASC;


-- 11. Cek source yang sudah waktunya di-scrape (scheduler tick)
-- Evaluasi cron expression dilakukan di aplikasi (Go), bukan di SQL.
-- Query ini ambil semua source aktif beserta last_scraped_at-nya.
SELECT id, key, schedule, last_scraped_at, auto_publish, ai_cleanup, auto_translate
FROM news_sources
WHERE is_active = true
ORDER BY last_scraped_at ASC NULLS FIRST;
```

---

## Migrations

```sql
-- ============================================================================
-- Migration 1: system_languages
-- ============================================================================
CREATE TABLE IF NOT EXISTS system_languages (
  id                   UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  code                 VARCHAR(10)  NOT NULL UNIQUE,
  name                 VARCHAR(100) NOT NULL,
  is_ui_language       BOOLEAN      NOT NULL DEFAULT false,
  is_translate_target  BOOLEAN      NOT NULL DEFAULT false,
  is_active            BOOLEAN      NOT NULL DEFAULT true,
  created_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at           TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO system_languages (code, name, is_ui_language, is_translate_target, is_active) VALUES
  ('id', 'Indonesian', true, true, true),
  ('en', 'English',    true, true, true),
  ('ko', 'Korean',     true, true, true)
ON CONFLICT (code) DO NOTHING;


-- ============================================================================
-- Migration 2: news_system_settings
-- ============================================================================
CREATE TABLE IF NOT EXISTS news_system_settings (
  id                   UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  translation_enabled  BOOLEAN     NOT NULL DEFAULT true,
  on_demand_enabled    BOOLEAN     NOT NULL DEFAULT true,
  updated_by           UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_news_system_settings_singleton
  ON news_system_settings ((true));

INSERT INTO news_system_settings (translation_enabled, on_demand_enabled)
VALUES (true, true)
ON CONFLICT DO NOTHING;


-- ============================================================================
-- Migration 2b: news_scopes
-- ============================================================================
CREATE TABLE IF NOT EXISTS news_scopes (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100) NOT NULL,
  slug        VARCHAR(120) NOT NULL UNIQUE,
  is_active   BOOLEAN      NOT NULL DEFAULT true,
  sort_order  INTEGER      NOT NULL DEFAULT 0,
  created_by  UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by  UUID         NULL     REFERENCES users(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_news_scopes_active_order
  ON news_scopes (is_active, sort_order)
  WHERE is_active = true;


-- ============================================================================
-- Migration 3: news_categories
-- ============================================================================
CREATE TABLE IF NOT EXISTS news_categories (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  name        VARCHAR(100) NOT NULL,
  slug        VARCHAR(120) NOT NULL UNIQUE,
  is_active   BOOLEAN      NOT NULL DEFAULT true,
  sort_order  INTEGER      NOT NULL DEFAULT 0,
  created_by  UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by  UUID         NULL     REFERENCES users(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);


-- ============================================================================
-- Migration 4: news_sources
-- ============================================================================
CREATE TABLE IF NOT EXISTS news_sources (
  id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  key               VARCHAR(50)  NOT NULL UNIQUE,
  name              VARCHAR(100) NOT NULL,
  base_url          VARCHAR(500) NOT NULL,
  original_language VARCHAR(10)  NOT NULL REFERENCES system_languages(code) ON DELETE RESTRICT,
  default_scope_id  UUID         NULL REFERENCES news_scopes(id) ON DELETE SET NULL,
  schedule          VARCHAR(100) NOT NULL DEFAULT '0 */6 * * *',
  last_scraped_at   TIMESTAMPTZ  NULL,
  auto_publish      BOOLEAN      NOT NULL DEFAULT false,
  ai_cleanup        BOOLEAN      NOT NULL DEFAULT false,
  auto_translate    BOOLEAN      NOT NULL DEFAULT false,
  is_active         BOOLEAN      NOT NULL DEFAULT true,
  created_by        UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by        UUID         NULL     REFERENCES users(id) ON DELETE SET NULL,
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);


-- ============================================================================
-- Migration 5: source_selectors
-- ============================================================================
CREATE TABLE IF NOT EXISTS source_selectors (
  id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id         UUID        NOT NULL UNIQUE REFERENCES news_sources(id) ON DELETE CASCADE,
  content_selector  TEXT        NULL,
  author_selector   TEXT        NULL,
  tags_selector     TEXT        NULL,
  extra_fields      JSONB       NULL,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- ============================================================================
-- Migration 6: source_categories
-- ============================================================================
CREATE TABLE IF NOT EXISTS source_categories (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id     UUID         NOT NULL REFERENCES news_sources(id) ON DELETE CASCADE,
  category_key  VARCHAR(100) NOT NULL,
  category_id   UUID         NULL REFERENCES news_categories(id) ON DELETE SET NULL,
  url_suffix    VARCHAR(500) NULL,
  url_override  VARCHAR(500) NULL,
  article_limit INTEGER      NOT NULL DEFAULT 10,
  is_active     BOOLEAN      NOT NULL DEFAULT true,
  updated_by    UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_source_category UNIQUE (source_id, category_key)
);


-- ============================================================================
-- Migration 7: articles
-- ============================================================================
CREATE TABLE IF NOT EXISTS articles (
  id                    UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  source_id             UUID          NULL REFERENCES news_sources(id) ON DELETE SET NULL,
  category_id           UUID          NULL REFERENCES news_categories(id) ON DELETE SET NULL,
  news_scope_id         UUID          NULL REFERENCES news_scopes(id) ON DELETE SET NULL,
  original_language     VARCHAR(10)   NOT NULL REFERENCES system_languages(code) ON DELETE RESTRICT,
  is_manual             BOOLEAN       NOT NULL DEFAULT false,
  original_url          VARCHAR(1000) NULL UNIQUE,
  status                VARCHAR(20)   NOT NULL DEFAULT 'draft',
  author_label          VARCHAR(100)  NOT NULL,
  author_region_id      UUID          NULL,
  view_count            INTEGER       NOT NULL DEFAULT 0,
  unique_view_count     INTEGER       NOT NULL DEFAULT 0,
  like_count            INTEGER       NOT NULL DEFAULT 0,
  comment_count         INTEGER       NOT NULL DEFAULT 0,
  is_featured           BOOLEAN       NOT NULL DEFAULT false,
  created_by            UUID          NULL REFERENCES users(id) ON DELETE SET NULL,
  published_by_user_id  UUID          NULL REFERENCES users(id) ON DELETE SET NULL,
  published_by_label    VARCHAR(200)  NULL,
  published_at          TIMESTAMPTZ   NULL,
  archived_at           TIMESTAMPTZ   NULL,
  created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_articles_original_url
  ON articles (original_url) WHERE original_url IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_articles_published
  ON articles (status, published_at DESC) WHERE status = 'published';
CREATE INDEX IF NOT EXISTS idx_articles_pending_approval
  ON articles (created_at DESC) WHERE status = 'pending_approval';


-- ============================================================================
-- Migration 8: article_translations
-- ============================================================================
CREATE TABLE IF NOT EXISTS article_translations (
  id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id       UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  language         VARCHAR(10) NOT NULL REFERENCES system_languages(code) ON DELETE RESTRICT,
  title            TEXT        NOT NULL,
  content          TEXT        NOT NULL,
  summary          TEXT        NULL,
  author           VARCHAR(200) NULL,
  thumbnail_url    TEXT        NULL,
  tags             TEXT[]      NULL,
  is_original      BOOLEAN     NOT NULL DEFAULT false,
  translate_status VARCHAR(20) NULL,
  translated_by    VARCHAR(50) NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_article_language UNIQUE (article_id, language)
);

CREATE INDEX IF NOT EXISTS idx_article_translations_pending
  ON article_translations (translate_status, created_at ASC)
  WHERE translate_status = 'pending';


-- ============================================================================
-- Migration 9: article_views, article_likes, article_bookmarks
-- ============================================================================
CREATE TABLE IF NOT EXISTS article_views (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  viewed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_article_view UNIQUE (article_id, user_id)
);

CREATE TABLE IF NOT EXISTS article_likes (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_article_like UNIQUE (article_id, user_id)
);

CREATE TABLE IF NOT EXISTS article_bookmarks (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CONSTRAINT uq_article_bookmark UNIQUE (article_id, user_id)
);


-- ============================================================================
-- Migration 10: article_comments
-- ============================================================================
CREATE TABLE IF NOT EXISTS article_comments (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  article_id UUID        NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  parent_id  UUID        NULL REFERENCES article_comments(id) ON DELETE SET NULL,
  content    TEXT        NULL,
  is_deleted BOOLEAN     NOT NULL DEFAULT false,
  deleted_by UUID        NULL REFERENCES users(id) ON DELETE SET NULL,
  deleted_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_article_comments_article
  ON article_comments (article_id, parent_id, created_at ASC)
  WHERE is_deleted = false;


-- ============================================================================
-- Migration 11: updated_at trigger
-- ============================================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_system_languages_updated_at
  BEFORE UPDATE ON system_languages
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_news_categories_updated_at
  BEFORE UPDATE ON news_categories
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_news_sources_updated_at
  BEFORE UPDATE ON news_sources
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_source_selectors_updated_at
  BEFORE UPDATE ON source_selectors
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_source_categories_updated_at
  BEFORE UPDATE ON source_categories
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_articles_updated_at
  BEFORE UPDATE ON articles
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_article_translations_updated_at
  BEFORE UPDATE ON article_translations
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_article_comments_updated_at
  BEFORE UPDATE ON article_comments
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

---

*Database Schema untuk News Module. Stack: Golang + PostgreSQL 13+.*
