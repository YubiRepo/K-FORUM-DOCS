# Database Schema — Ads Module

> **Stack:** Golang + PostgreSQL  
> **Dibuat:** 2026-06-10  

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [PostgreSQL DDL Schema](#postgresql-ddl-schema)
3. [Golang Structs (ORM Mapping)](#golang-structs-orm-mapping)
4. [Migrations](#migrations)

---

## Overview Relasi

```
users
  └── ads (created_by)              -- siapa yang buat iklan
  └── ads (reviewed_by)             -- siapa yang approve/reject
  └── ad_analytics (per impression & klik — aggregated)

ads
  └── ad_analytics                  -- tracking impressi & klik per ads

ad_settings                         -- global config, single row, managed by superadmin
```

---

## PostgreSQL DDL Schema

```sql
-- ============================================================================
-- ADS MODULE DATABASE SCHEMA
-- Stack: PostgreSQL 13+
-- Created: 2026-06-10
-- ============================================================================


-- ============================================================================
-- 1. AD SETTINGS TABLE
-- Global konfigurasi ads, hanya 1 row. Dikelola oleh superadmin.
-- ============================================================================

CREATE TABLE ad_settings (
  id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Approval
  approval_mode               VARCHAR(20) NOT NULL DEFAULT 'require_review',
  -- 'auto_publish' | 'require_review'

  -- Batas per member
  max_active_ads_per_member   INTEGER NOT NULL DEFAULT 3,
  max_duration_days           INTEGER NOT NULL DEFAULT 30,

  -- Tampilan
  feed_ads_interval           INTEGER NOT NULL DEFAULT 5,
  -- Setiap N item feed, sisipkan 1 native/text ad
  slider_max_items            INTEGER NOT NULL DEFAULT 5,
  -- Max slide yang tampil di home slider banner

  -- Audit
  updated_at                  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_by                  UUID NULL REFERENCES users(id),

  -- Constraints
  CONSTRAINT ad_settings_approval_mode_check CHECK (
    approval_mode IN ('auto_publish', 'require_review')
  ),
  CONSTRAINT ad_settings_max_active_check CHECK (max_active_ads_per_member > 0),
  CONSTRAINT ad_settings_max_duration_check CHECK (max_duration_days > 0),
  CONSTRAINT ad_settings_feed_interval_check CHECK (feed_ads_interval > 0),
  CONSTRAINT ad_settings_slider_max_check CHECK (slider_max_items > 0)
);

-- Seed: pastikan selalu ada tepat 1 row
INSERT INTO ad_settings (approval_mode, max_active_ads_per_member, max_duration_days, feed_ads_interval, slider_max_items)
VALUES ('require_review', 3, 30, 5, 5);

-- Trigger: update updated_at otomatis
CREATE OR REPLACE FUNCTION update_ad_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ad_settings_updated_at_trigger
BEFORE UPDATE ON ad_settings
FOR EACH ROW
EXECUTE FUNCTION update_ad_settings_updated_at();


-- ============================================================================
-- 2. ADS TABLE
-- Tabel utama — satu row per iklan.
-- ============================================================================

CREATE TABLE ads (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  -- Identifikasi
  title           VARCHAR(100) NOT NULL,
  -- Judul internal untuk backoffice, tidak tampil di mobile

  ad_type         VARCHAR(20) NOT NULL,
  -- 'image_banner' | 'video_banner' | 'text_ad' | 'native_ad'

  -- -----------------------------------------------------------------------
  -- Field konten per tipe (nullable, diisi sesuai ad_type)
  -- -----------------------------------------------------------------------

  -- image_banner & native_ad
  image_url       TEXT NULL,

  -- video_banner
  video_url       TEXT NULL,
  thumbnail_url   TEXT NULL,

  -- text_ad & native_ad
  headline        VARCHAR(60) NULL,
  body_text       VARCHAR(200) NULL,
  cta_label       VARCHAR(20) NULL,
  icon_url        TEXT NULL,

  -- native_ad
  sponsor_name    VARCHAR(50) NULL,
  sponsor_logo_url TEXT NULL,

  -- -----------------------------------------------------------------------
  -- Field umum semua tipe
  -- -----------------------------------------------------------------------

  click_url       TEXT NOT NULL,
  -- URL tujuan saat user tap. Format: https:// atau kai:// (deep link)
  -- Deep link kai:// hanya boleh diset oleh superadmin (dicek di application layer)

  -- Jadwal tayang
  start_date      DATE NOT NULL,
  end_date        DATE NOT NULL,

  -- Catatan internal creator (tidak tampil di mobile)
  notes           TEXT NULL,

  -- -----------------------------------------------------------------------
  -- Status & Approval
  -- -----------------------------------------------------------------------

  status          VARCHAR(20) NOT NULL DEFAULT 'draft',
  -- 'draft' | 'pending' | 'active' | 'rejected' | 'paused' | 'expired'

  reject_reason   TEXT NULL,
  -- Wajib diisi superadmin saat reject

  reviewed_by     UUID NULL REFERENCES users(id),
  reviewed_at     TIMESTAMPTZ NULL,

  -- -----------------------------------------------------------------------
  -- Analytics (cached counter, di-update via background job)
  -- -----------------------------------------------------------------------

  total_impressions   BIGINT NOT NULL DEFAULT 0,
  total_clicks        BIGINT NOT NULL DEFAULT 0,
  -- CTR dihitung on-the-fly: total_clicks / total_impressions * 100

  -- -----------------------------------------------------------------------
  -- Creator & Timestamps
  -- -----------------------------------------------------------------------

  created_by      UUID NOT NULL REFERENCES users(id),
  created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_by      UUID NULL REFERENCES users(id),

  -- -----------------------------------------------------------------------
  -- Constraints
  -- -----------------------------------------------------------------------

  CONSTRAINT ads_valid_type CHECK (
    ad_type IN ('image_banner', 'video_banner', 'text_ad', 'native_ad')
  ),

  CONSTRAINT ads_valid_status CHECK (
    status IN ('draft', 'pending', 'active', 'rejected', 'paused', 'expired')
  ),

  CONSTRAINT ads_date_range_check CHECK (
    end_date > start_date
  ),

  -- image_banner wajib punya image_url
  CONSTRAINT ads_image_banner_check CHECK (
    ad_type != 'image_banner' OR image_url IS NOT NULL
  ),

  -- video_banner wajib punya video_url dan thumbnail_url
  CONSTRAINT ads_video_banner_check CHECK (
    ad_type != 'video_banner' OR (video_url IS NOT NULL AND thumbnail_url IS NOT NULL)
  ),

  -- text_ad wajib punya headline, body_text, cta_label
  CONSTRAINT ads_text_ad_check CHECK (
    ad_type != 'text_ad' OR (headline IS NOT NULL AND body_text IS NOT NULL AND cta_label IS NOT NULL)
  ),

  -- native_ad wajib punya title (reuse kolom title), body_text, sponsor_name
  CONSTRAINT ads_native_ad_check CHECK (
    ad_type != 'native_ad' OR (body_text IS NOT NULL AND sponsor_name IS NOT NULL)
  ),

  -- reject_reason wajib ada saat status = rejected
  CONSTRAINT ads_reject_reason_check CHECK (
    status != 'rejected' OR reject_reason IS NOT NULL
  ),

  -- reviewed_by & reviewed_at harus sama-sama null atau sama-sama isi
  CONSTRAINT ads_reviewed_consistency CHECK (
    (reviewed_by IS NULL AND reviewed_at IS NULL) OR
    (reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL)
  )
);

-- Indexes
CREATE INDEX idx_ads_status
  ON ads(status);

CREATE INDEX idx_ads_ad_type
  ON ads(ad_type);

CREATE INDEX idx_ads_created_by
  ON ads(created_by);

CREATE INDEX idx_ads_start_end_date
  ON ads(start_date, end_date);

-- Index untuk query slider & feed: ads yang sedang aktif
CREATE INDEX idx_ads_active_slider
  ON ads(start_date, created_by)
  WHERE status = 'active' AND ad_type IN ('image_banner', 'video_banner');

CREATE INDEX idx_ads_active_feed
  ON ads(start_date, created_by)
  WHERE status = 'active' AND ad_type IN ('text_ad', 'native_ad');

-- Index untuk moderasi queue superadmin
CREATE INDEX idx_ads_pending
  ON ads(created_at DESC)
  WHERE status = 'pending';

-- Index untuk cron job expired check harian
CREATE INDEX idx_ads_expiry_check
  ON ads(end_date)
  WHERE status IN ('active', 'paused');

-- Trigger: update updated_at otomatis
CREATE OR REPLACE FUNCTION update_ads_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ads_updated_at_trigger
BEFORE UPDATE ON ads
FOR EACH ROW
EXECUTE FUNCTION update_ads_updated_at();


-- ============================================================================
-- 3. AD ANALYTICS TABLE
-- Event log per impressi dan klik — untuk data granular.
-- Total aggregated disimpan di tabel ads (total_impressions, total_clicks).
-- ============================================================================

CREATE TABLE ad_analytics (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),

  ad_id       UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,

  event_type  VARCHAR(20) NOT NULL,
  -- 'impression' | 'click'

  user_id     UUID NULL REFERENCES users(id) ON DELETE SET NULL,
  -- NULL jika guest (belum login) — tetap catat

  session_id  VARCHAR(100) NULL,
  -- Untuk dedup impressi dalam satu session

  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT ad_analytics_event_type_check CHECK (
    event_type IN ('impression', 'click')
  )
);

-- Indexes
CREATE INDEX idx_ad_analytics_ad_id
  ON ad_analytics(ad_id);

CREATE INDEX idx_ad_analytics_ad_event
  ON ad_analytics(ad_id, event_type);

CREATE INDEX idx_ad_analytics_created_at
  ON ad_analytics(created_at DESC);

-- Composite index untuk dedup impressi: satu session per ads
CREATE UNIQUE INDEX idx_ad_analytics_impression_dedup
  ON ad_analytics(ad_id, session_id)
  WHERE event_type = 'impression' AND session_id IS NOT NULL;
```

---

## Golang Structs (ORM Mapping)

```go
// internal/domain/ads/ad_settings.go

package ads

import "time"

// AdSettings merepresentasikan konfigurasi global ads.
// Selalu hanya ada 1 row di tabel ad_settings.
type AdSettings struct {
  ID                       string    `db:"id"                          json:"id"`
  ApprovalMode             string    `db:"approval_mode"               json:"approval_mode"`
  // "auto_publish" | "require_review"
  MaxActiveAdsPerMember    int       `db:"max_active_ads_per_member"   json:"max_active_ads_per_member"`
  MaxDurationDays          int       `db:"max_duration_days"           json:"max_duration_days"`
  FeedAdsInterval          int       `db:"feed_ads_interval"           json:"feed_ads_interval"`
  SliderMaxItems           int       `db:"slider_max_items"            json:"slider_max_items"`
  UpdatedAt                time.Time `db:"updated_at"                  json:"updated_at"`
  UpdatedBy                *string   `db:"updated_by"                  json:"updated_by,omitempty"`
}

const (
  ApprovalModeAutoPublish   = "auto_publish"
  ApprovalModeRequireReview = "require_review"
)
```

```go
// internal/domain/ads/ad.go

package ads

import "time"

// Ad merepresentasikan satu iklan di platform KAI.
type Ad struct {
  ID    string `db:"id"      json:"id"`
  Title string `db:"title"   json:"title"`

  AdType string `db:"ad_type" json:"ad_type"`
  // "image_banner" | "video_banner" | "text_ad" | "native_ad"

  // Field konten — nullable, diisi sesuai AdType
  ImageUrl        *string `db:"image_url"         json:"image_url,omitempty"`
  VideoUrl        *string `db:"video_url"         json:"video_url,omitempty"`
  ThumbnailUrl    *string `db:"thumbnail_url"     json:"thumbnail_url,omitempty"`
  Headline        *string `db:"headline"          json:"headline,omitempty"`
  BodyText        *string `db:"body_text"         json:"body_text,omitempty"`
  CtaLabel        *string `db:"cta_label"         json:"cta_label,omitempty"`
  IconUrl         *string `db:"icon_url"          json:"icon_url,omitempty"`
  SponsorName     *string `db:"sponsor_name"      json:"sponsor_name,omitempty"`
  SponsorLogoUrl  *string `db:"sponsor_logo_url"  json:"sponsor_logo_url,omitempty"`

  // Field umum
  ClickUrl  string    `db:"click_url"  json:"click_url"`
  StartDate time.Time `db:"start_date" json:"start_date"`
  EndDate   time.Time `db:"end_date"   json:"end_date"`
  Notes     *string   `db:"notes"      json:"notes,omitempty"`

  // Status & Approval
  Status       string    `db:"status"        json:"status"`
  RejectReason *string   `db:"reject_reason" json:"reject_reason,omitempty"`
  ReviewedBy   *string   `db:"reviewed_by"   json:"reviewed_by,omitempty"`
  ReviewedAt   *time.Time `db:"reviewed_at"  json:"reviewed_at,omitempty"`

  // Analytics (cached)
  TotalImpressions int64 `db:"total_impressions" json:"total_impressions"`
  TotalClicks      int64 `db:"total_clicks"      json:"total_clicks"`

  // Timestamps
  CreatedBy string    `db:"created_by" json:"created_by"`
  CreatedAt time.Time `db:"created_at" json:"created_at"`
  UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
  UpdatedBy *string   `db:"updated_by" json:"updated_by,omitempty"`
}

// AdWithCreator dipakai untuk response yang include info user pembuat.
type AdWithCreator struct {
  Ad
  CreatorName   string  `db:"creator_name"   json:"creator_name"`
  CreatorAvatar *string `db:"creator_avatar" json:"creator_avatar,omitempty"`
}

// Konstanta AdType
const (
  AdTypeImageBanner = "image_banner"
  AdTypeVideoBanner = "video_banner"
  AdTypeTextAd      = "text_ad"
  AdTypeNativeAd    = "native_ad"
)

// Konstanta Status
const (
  AdStatusDraft    = "draft"
  AdStatusPending  = "pending"
  AdStatusActive   = "active"
  AdStatusRejected = "rejected"
  AdStatusPaused   = "paused"
  AdStatusExpired  = "expired"
)

// IsSliderType returns true jika tipe ini tampil di slider banner
func (a *Ad) IsSliderType() bool {
  return a.AdType == AdTypeImageBanner || a.AdType == AdTypeVideoBanner
}

// IsFeedType returns true jika tipe ini tampil selip di feed
func (a *Ad) IsFeedType() bool {
  return a.AdType == AdTypeTextAd || a.AdType == AdTypeNativeAd
}

// CTR menghitung click-through rate dalam persen
func (a *Ad) CTR() float64 {
  if a.TotalImpressions == 0 {
    return 0
  }
  return float64(a.TotalClicks) / float64(a.TotalImpressions) * 100
}
```

```go
// internal/domain/ads/ad_analytics.go

package ads

import "time"

// AdAnalytic merepresentasikan satu event impressi atau klik.
type AdAnalytic struct {
  ID        string    `db:"id"         json:"id"`
  AdID      string    `db:"ad_id"      json:"ad_id"`
  EventType string    `db:"event_type" json:"event_type"`
  // "impression" | "click"
  UserID    *string   `db:"user_id"    json:"user_id,omitempty"`
  SessionID *string   `db:"session_id" json:"session_id,omitempty"`
  CreatedAt time.Time `db:"created_at" json:"created_at"`
}

const (
  EventTypeImpression = "impression"
  EventTypeClick      = "click"
)

// AdAnalyticsSummary dipakai untuk response analytics per ads
type AdAnalyticsSummary struct {
  AdID             string  `json:"ad_id"`
  TotalImpressions int64   `json:"total_impressions"`
  TotalClicks      int64   `json:"total_clicks"`
  CTR              float64 `json:"ctr"`
}
```

---

## Migrations

```sql
-- migrations/000001_create_ad_settings.up.sql

CREATE TABLE ad_settings (
  id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  approval_mode             VARCHAR(20) NOT NULL DEFAULT 'require_review',
  max_active_ads_per_member INTEGER NOT NULL DEFAULT 3,
  max_duration_days         INTEGER NOT NULL DEFAULT 30,
  feed_ads_interval         INTEGER NOT NULL DEFAULT 5,
  slider_max_items          INTEGER NOT NULL DEFAULT 5,
  updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_by                UUID NULL REFERENCES users(id),

  CONSTRAINT ad_settings_approval_mode_check CHECK (
    approval_mode IN ('auto_publish', 'require_review')
  ),
  CONSTRAINT ad_settings_max_active_check CHECK (max_active_ads_per_member > 0),
  CONSTRAINT ad_settings_max_duration_check CHECK (max_duration_days > 0),
  CONSTRAINT ad_settings_feed_interval_check CHECK (feed_ads_interval > 0),
  CONSTRAINT ad_settings_slider_max_check CHECK (slider_max_items > 0)
);

INSERT INTO ad_settings (approval_mode, max_active_ads_per_member, max_duration_days, feed_ads_interval, slider_max_items)
VALUES ('require_review', 3, 30, 5, 5);
```

```sql
-- migrations/000002_create_ads.up.sql

CREATE TABLE ads (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title             VARCHAR(100) NOT NULL,
  ad_type           VARCHAR(20) NOT NULL,
  image_url         TEXT NULL,
  video_url         TEXT NULL,
  thumbnail_url     TEXT NULL,
  headline          VARCHAR(60) NULL,
  body_text         VARCHAR(200) NULL,
  cta_label         VARCHAR(20) NULL,
  icon_url          TEXT NULL,
  sponsor_name      VARCHAR(50) NULL,
  sponsor_logo_url  TEXT NULL,
  click_url         TEXT NOT NULL,
  start_date        DATE NOT NULL,
  end_date          DATE NOT NULL,
  notes             TEXT NULL,
  status            VARCHAR(20) NOT NULL DEFAULT 'draft',
  reject_reason     TEXT NULL,
  reviewed_by       UUID NULL REFERENCES users(id),
  reviewed_at       TIMESTAMPTZ NULL,
  total_impressions BIGINT NOT NULL DEFAULT 0,
  total_clicks      BIGINT NOT NULL DEFAULT 0,
  created_by        UUID NOT NULL REFERENCES users(id),
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_by        UUID NULL REFERENCES users(id),

  CONSTRAINT ads_valid_type CHECK (
    ad_type IN ('image_banner', 'video_banner', 'text_ad', 'native_ad')
  ),
  CONSTRAINT ads_valid_status CHECK (
    status IN ('draft', 'pending', 'active', 'rejected', 'paused', 'expired')
  ),
  CONSTRAINT ads_date_range_check CHECK (end_date > start_date),
  CONSTRAINT ads_image_banner_check CHECK (
    ad_type != 'image_banner' OR image_url IS NOT NULL
  ),
  CONSTRAINT ads_video_banner_check CHECK (
    ad_type != 'video_banner' OR (video_url IS NOT NULL AND thumbnail_url IS NOT NULL)
  ),
  CONSTRAINT ads_text_ad_check CHECK (
    ad_type != 'text_ad' OR (headline IS NOT NULL AND body_text IS NOT NULL AND cta_label IS NOT NULL)
  ),
  CONSTRAINT ads_native_ad_check CHECK (
    ad_type != 'native_ad' OR (body_text IS NOT NULL AND sponsor_name IS NOT NULL)
  ),
  CONSTRAINT ads_reject_reason_check CHECK (
    status != 'rejected' OR reject_reason IS NOT NULL
  ),
  CONSTRAINT ads_reviewed_consistency CHECK (
    (reviewed_by IS NULL AND reviewed_at IS NULL) OR
    (reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL)
  )
);

CREATE INDEX idx_ads_status ON ads(status);
CREATE INDEX idx_ads_ad_type ON ads(ad_type);
CREATE INDEX idx_ads_created_by ON ads(created_by);
CREATE INDEX idx_ads_start_end_date ON ads(start_date, end_date);
CREATE INDEX idx_ads_active_slider ON ads(start_date, created_by)
  WHERE status = 'active' AND ad_type IN ('image_banner', 'video_banner');
CREATE INDEX idx_ads_active_feed ON ads(start_date, created_by)
  WHERE status = 'active' AND ad_type IN ('text_ad', 'native_ad');
CREATE INDEX idx_ads_pending ON ads(created_at DESC) WHERE status = 'pending';
CREATE INDEX idx_ads_expiry_check ON ads(end_date) WHERE status IN ('active', 'paused');

CREATE OR REPLACE FUNCTION update_ads_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ads_updated_at_trigger
BEFORE UPDATE ON ads
FOR EACH ROW
EXECUTE FUNCTION update_ads_updated_at();
```

```sql
-- migrations/000003_create_ad_analytics.up.sql

CREATE TABLE ad_analytics (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ad_id       UUID NOT NULL REFERENCES ads(id) ON DELETE CASCADE,
  event_type  VARCHAR(20) NOT NULL,
  user_id     UUID NULL REFERENCES users(id) ON DELETE SET NULL,
  session_id  VARCHAR(100) NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT ad_analytics_event_type_check CHECK (
    event_type IN ('impression', 'click')
  )
);

CREATE INDEX idx_ad_analytics_ad_id ON ad_analytics(ad_id);
CREATE INDEX idx_ad_analytics_ad_event ON ad_analytics(ad_id, event_type);
CREATE INDEX idx_ad_analytics_created_at ON ad_analytics(created_at DESC);
CREATE UNIQUE INDEX idx_ad_analytics_impression_dedup
  ON ad_analytics(ad_id, session_id)
  WHERE event_type = 'impression' AND session_id IS NOT NULL;
```

```sql
-- Down migrations

-- migrations/000003_create_ad_analytics.down.sql
DROP TABLE IF EXISTS ad_analytics;

-- migrations/000002_create_ads.down.sql
DROP TRIGGER IF EXISTS ads_updated_at_trigger ON ads;
DROP FUNCTION IF EXISTS update_ads_updated_at;
DROP TABLE IF EXISTS ads;

-- migrations/000001_create_ad_settings.down.sql
DROP TABLE IF EXISTS ad_settings;
```

---

## Catatan Desain

**Kenapa `total_impressions` dan `total_clicks` disimpan di tabel `ads` (bukan hanya di `ad_analytics`)?**
Query COUNT dari `ad_analytics` untuk setiap tampilan mobile akan berat seiring data membesar. Kolom cached di `ads` di-update via background job setiap beberapa menit — cukup untuk display analytics, dan `ad_analytics` tetap disimpan untuk data granular jika sewaktu-waktu perlu breakdown per waktu atau per user.

**Kenapa dedup impressi pakai `session_id` bukan `user_id`?**
Guest (belum login) tidak punya `user_id`, tapi tetap bisa melihat ads. `session_id` dari client (generated saat app launch) memastikan satu sesi tidak menggandakan impressi — lebih akurat dari pure user-based dedup.

**Kenapa `ad_settings` pakai single-row pattern?**
Konfigurasi ads adalah setting global yang tunggal — tidak perlu multi-row. Lebih simpel dari key-value store karena schema-nya jelas, typed, dan mudah di-query tanpa parsing.
