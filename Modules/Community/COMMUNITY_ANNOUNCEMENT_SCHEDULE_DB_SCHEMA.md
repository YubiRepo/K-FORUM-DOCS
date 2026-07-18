# Database Schema — Community Announcement & Schedule

> **Stack:** Golang + PostgreSQL
> **Berdasarkan:** `COMMUNITY_ANNOUNCEMENT_SCHEDULE_RULES.md`
> **Module:** Community (sub-fitur: Papan Pengumuman & Schedule)
> **Dibuat:** 2026-07-03

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [PostgreSQL DDL Schema](#postgresql-ddl-schema)
3. [Golang Structs (ORM Mapping)](#golang-structs-orm-mapping)
4. [Migrations](#migrations)
5. [Enum Values](#enum-values)
6. [Query Examples](#query-examples)
7. [Catatan Occurrence & Recurrence](#catatan-occurrence--recurrence)

---

## Overview Relasi

```
communities (dari modul Community)
  ├── community_announcements          (papan pengumuman per komunitas)
  ├── community_schedule_entries       (agenda per komunitas)
  │     ├── community_schedule_rsvps        (RSVP per occurrence)
  │     └── community_schedule_exceptions   (cancel/modify per occurrence)
  └── (permission check via Role-Permission: manage_community_announcement / manage_community_schedule)

users (dari modul Auth)
  ├── community_announcements (author_id)
  ├── community_schedule_entries (created_by)
  └── community_schedule_rsvps (user_id)
```

> **Occurrence tidak disimpan sebagai baris.** `community_schedule_rsvps` dan `community_schedule_exceptions` menunjuk ke satu occurrence lewat pasangan `(entry_id, occurrence_date)`. Daftar occurrence dihitung on-the-fly dari `entries.recurrence` (lihat §7).

---

## PostgreSQL DDL Schema

```sql
-- ============================================================================
-- COMMUNITY ANNOUNCEMENT & SCHEDULE — DATABASE SCHEMA
-- Stack: PostgreSQL 13+
-- Created: 2026-07-03
-- Scope: sub-fitur modul Community (per-community, member-facing)
-- ============================================================================


-- ============================================================================
-- 1. COMMUNITY_ANNOUNCEMENTS
-- Papan pengumuman read-only per komunitas. Dibuat oleh pemegang permission
-- manage_community_announcement (leader default; moderator bila di-grant).
-- ============================================================================

CREATE TABLE community_announcements (
  id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  community_id  UUID          NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
  author_id     UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  -- Konten
  title         VARCHAR(150)  NOT NULL,
  body          TEXT          NOT NULL,               -- plain text
  media         JSONB         NOT NULL DEFAULT '[]',  -- array of {url, thumb_url, width, height, order}, maks 5

  -- Klasifikasi & tampilan
  priority      VARCHAR(10)   NOT NULL DEFAULT 'normal',  -- 'normal' | 'important'
  is_pinned     BOOLEAN       NOT NULL DEFAULT FALSE,

  -- Lifecycle
  status        VARCHAR(10)   NOT NULL DEFAULT 'draft',   -- 'draft' | 'published' | 'archived'
  published_at  TIMESTAMPTZ   NULL,
  expires_at    TIMESTAMPTZ   NULL,                       -- setelah lewat: auto-hide dari member
  archived_at   TIMESTAMPTZ   NULL,

  created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_ann_priority CHECK (priority IN ('normal', 'important')),
  CONSTRAINT chk_ann_status   CHECK (status IN ('draft', 'published', 'archived'))
);

-- Feed member: pengumuman published & belum expired, urut pinned → priority → terbaru
CREATE INDEX idx_ann_community_feed
  ON community_announcements (community_id, is_pinned DESC, priority, published_at DESC)
  WHERE status = 'published';

CREATE INDEX idx_ann_community        ON community_announcements (community_id);
CREATE INDEX idx_ann_author           ON community_announcements (author_id);
CREATE INDEX idx_ann_expires          ON community_announcements (expires_at) WHERE expires_at IS NOT NULL;


-- ============================================================================
-- 2. COMMUNITY_SCHEDULE_ENTRIES
-- Agenda komunitas. Satu row bisa mewakili banyak tanggal bila recurrence diisi.
-- Dibuat oleh pemegang permission manage_community_schedule.
-- ============================================================================

CREATE TABLE community_schedule_entries (
  id            UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  community_id  UUID          NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
  created_by    UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  -- Konten
  title         VARCHAR(200)  NOT NULL,
  description   TEXT          NULL,
  location      TEXT          NULL,

  -- Waktu
  start_at      TIMESTAMPTZ   NOT NULL,               -- recurring: waktu occurrence pertama + time-of-day acuan
  end_at        TIMESTAMPTZ   NULL,                   -- opsional; jika ada harus >= start_at
  all_day       BOOLEAN       NOT NULL DEFAULT FALSE,

  -- Recurrence (AKTIF Phase 1)
  recurrence    TEXT          NULL,                   -- RRULE-style, mis. 'FREQ=WEEKLY;BYDAY=SA'. NULL = one-off

  -- Timezone (ditambahkan migrasi 20260717080434, 2026-07-17)
  timezone      VARCHAR(64)   NOT NULL DEFAULT 'Asia/Jakarta',  -- IANA identifier; wajib diisi eksplisit oleh creator saat create (lihat API spec mobile B4). Kolom ditambah via ALTER TABLE ADD COLUMN dengan DEFAULT ini, sehingga row lama otomatis ter-backfill ke 'Asia/Jakarta' (asumsi yang sudah berlaku sebelum migrasi, bukan regresi) — lihat COMMUNITY_ANNOUNCEMENT_SCHEDULE_RULES.md §Recurrence & Occurrence untuk kenapa kolom ini perlu.

  -- Lifecycle
  status        VARCHAR(10)   NOT NULL DEFAULT 'active',  -- 'active' | 'cancelled' (batal seluruh agenda)

  created_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_sch_status CHECK (status IN ('active', 'cancelled')),
  CONSTRAINT chk_sch_time   CHECK (end_at IS NULL OR end_at >= start_at)
);

CREATE INDEX idx_sch_community     ON community_schedule_entries (community_id);
CREATE INDEX idx_sch_start         ON community_schedule_entries (start_at);
CREATE INDEX idx_sch_created_by    ON community_schedule_entries (created_by);
CREATE INDEX idx_sch_active_range  ON community_schedule_entries (community_id, start_at) WHERE status = 'active';


-- ============================================================================
-- 3. COMMUNITY_SCHEDULE_RSVPS
-- RSVP indikator minat per occurrence (BUKAN absensi otoritatif).
-- Satu RSVP per (entry, occurrence_date, user).
-- ============================================================================

CREATE TABLE community_schedule_rsvps (
  id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  entry_id         UUID         NOT NULL REFERENCES community_schedule_entries(id) ON DELETE CASCADE,
  occurrence_date  DATE         NOT NULL,             -- tanggal occurrence yang di-RSVP
  user_id          UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,

  response         VARCHAR(10)  NOT NULL,             -- 'going' | 'maybe' | 'not_going'

  created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_rsvp_response CHECK (response IN ('going', 'maybe', 'not_going')),
  CONSTRAINT uq_rsvp_occurrence UNIQUE (entry_id, occurrence_date, user_id)
);

CREATE INDEX idx_rsvp_occurrence ON community_schedule_rsvps (entry_id, occurrence_date);
CREATE INDEX idx_rsvp_user       ON community_schedule_rsvps (user_id);


-- ============================================================================
-- 4. COMMUNITY_SCHEDULE_EXCEPTIONS
-- Override satu occurrence dari agenda berulang.
-- Phase 1: hanya type='cancelled' (skip satu tanggal).
-- Phase 2: type='modified' (ubah detail per occurrence) — hook disiapkan.
-- ============================================================================

CREATE TABLE community_schedule_exceptions (
  id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  entry_id         UUID         NOT NULL REFERENCES community_schedule_entries(id) ON DELETE CASCADE,
  occurrence_date  DATE         NOT NULL,             -- occurrence yang di-override

  type             VARCHAR(10)  NOT NULL DEFAULT 'cancelled',  -- 'cancelled' (P1) | 'modified' (P2)

  -- Hook Phase 2 (ubah detail per occurrence) — NULL di Phase 1
  override_start_at TIMESTAMPTZ NULL,
  override_end_at   TIMESTAMPTZ NULL,
  override_title    VARCHAR(200) NULL,
  override_location TEXT         NULL,

  created_by       UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_exc_type CHECK (type IN ('cancelled', 'modified')),
  CONSTRAINT uq_exc_occurrence UNIQUE (entry_id, occurrence_date)
);

CREATE INDEX idx_exc_entry ON community_schedule_exceptions (entry_id, occurrence_date);
```

---

## Golang Structs (ORM Mapping)

```go
package community

import (
	"encoding/json"
	"time"
)

// ---------- Enums ----------

type AnnouncementPriority string
type AnnouncementStatus   string

const (
	PriorityNormal    AnnouncementPriority = "normal"
	PriorityImportant AnnouncementPriority = "important"

	AnnStatusDraft     AnnouncementStatus = "draft"
	AnnStatusPublished AnnouncementStatus = "published"
	AnnStatusArchived  AnnouncementStatus = "archived"
)

type ScheduleStatus string
type RsvpResponse   string
type ExceptionType  string

const (
	SchStatusActive    ScheduleStatus = "active"
	SchStatusCancelled ScheduleStatus = "cancelled"

	RsvpGoing    RsvpResponse = "going"
	RsvpMaybe    RsvpResponse = "maybe"
	RsvpNotGoing RsvpResponse = "not_going"

	ExcCancelled ExceptionType = "cancelled"
	ExcModified  ExceptionType = "modified"
)

// ---------- Media ----------

type MediaItem struct {
	URL      string `json:"url"`
	ThumbURL string `json:"thumb_url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Order    int    `json:"order"`
}

// ---------- 1. CommunityAnnouncement ----------

type CommunityAnnouncement struct {
	ID          string               `json:"id" db:"id"`
	CommunityID string               `json:"community_id" db:"community_id"`
	AuthorID    string               `json:"author_id" db:"author_id"`

	Title string          `json:"title" db:"title"`
	Body  string          `json:"body" db:"body"`
	Media json.RawMessage `json:"media" db:"media"` // []MediaItem, maks 5

	Priority AnnouncementPriority `json:"priority" db:"priority"`
	IsPinned bool                 `json:"is_pinned" db:"is_pinned"`

	Status      AnnouncementStatus `json:"status" db:"status"`
	PublishedAt *time.Time         `json:"published_at,omitempty" db:"published_at"`
	ExpiresAt   *time.Time         `json:"expires_at,omitempty" db:"expires_at"`
	ArchivedAt  *time.Time         `json:"archived_at,omitempty" db:"archived_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (a *CommunityAnnouncement) IsVisibleToMember(now time.Time) bool {
	if a.Status != AnnStatusPublished {
		return false
	}
	if a.ExpiresAt != nil && now.After(*a.ExpiresAt) {
		return false
	}
	return true
}

// ---------- 2. CommunityScheduleEntry ----------

type CommunityScheduleEntry struct {
	ID          string `json:"id" db:"id"`
	CommunityID string `json:"community_id" db:"community_id"`
	CreatedBy   string `json:"created_by" db:"created_by"`

	Title       string  `json:"title" db:"title"`
	Description *string `json:"description,omitempty" db:"description"`
	Location    *string `json:"location,omitempty" db:"location"`

	StartAt time.Time  `json:"start_at" db:"start_at"`
	EndAt   *time.Time `json:"end_at,omitempty" db:"end_at"`
	AllDay  bool       `json:"all_day" db:"all_day"`

	Recurrence *string `json:"recurrence,omitempty" db:"recurrence"` // RRULE, NULL = one-off

	Timezone string `json:"timezone" db:"timezone"` // IANA identifier, wajib eksplisit (ditambahkan 2026-07-17)

	Status ScheduleStatus `json:"status" db:"status"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (e *CommunityScheduleEntry) IsRecurring() bool {
	return e.Recurrence != nil && *e.Recurrence != ""
}

// ---------- 3. CommunityScheduleRsvp ----------

type CommunityScheduleRsvp struct {
	ID             string       `json:"id" db:"id"`
	EntryID        string       `json:"entry_id" db:"entry_id"`
	OccurrenceDate time.Time    `json:"occurrence_date" db:"occurrence_date"` // DATE
	UserID         string       `json:"user_id" db:"user_id"`
	Response       RsvpResponse `json:"response" db:"response"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ---------- 4. CommunityScheduleException ----------

type CommunityScheduleException struct {
	ID             string        `json:"id" db:"id"`
	EntryID        string        `json:"entry_id" db:"entry_id"`
	OccurrenceDate time.Time     `json:"occurrence_date" db:"occurrence_date"` // DATE
	Type           ExceptionType `json:"type" db:"type"`

	// Phase 2 hook (NULL di Phase 1)
	OverrideStartAt  *time.Time `json:"override_start_at,omitempty" db:"override_start_at"`
	OverrideEndAt    *time.Time `json:"override_end_at,omitempty" db:"override_end_at"`
	OverrideTitle    *string    `json:"override_title,omitempty" db:"override_title"`
	OverrideLocation *string    `json:"override_location,omitempty" db:"override_location"`

	CreatedBy string    `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ---------- Occurrence (computed, bukan tabel) ----------

// ScheduleOccurrence adalah satu instance agenda pada tanggal tertentu,
// hasil expand recurrence. Dirakit di service layer, tidak disimpan di DB.
type ScheduleOccurrence struct {
	Entry          *CommunityScheduleEntry `json:"entry"`
	OccurrenceDate time.Time               `json:"occurrence_date"`
	StartAt        time.Time               `json:"start_at"`
	EndAt          *time.Time              `json:"end_at,omitempty"`
	IsCancelled    bool                    `json:"is_cancelled"`   // dari exception type=cancelled atau entry.status=cancelled
	MyResponse     *RsvpResponse           `json:"my_response,omitempty"`
	RsvpSummary    *RsvpSummary            `json:"rsvp_summary,omitempty"` // diisi untuk pengelola
}

type RsvpSummary struct {
	Going    int `json:"going"`
	Maybe    int `json:"maybe"`
	NotGoing int `json:"not_going"`
}
```

---

## Migrations

Urutan: tabel `communities` dan `users` harus sudah ada (modul Community & Auth). Migration ini menyusul setelahnya.

### `migrations/2026070300_create_community_announcements.up.sql`
```sql
-- community_announcements (lihat DDL §1)
-- + indexes
```

### `migrations/2026070301_create_community_schedule.up.sql`
```sql
-- community_schedule_entries       (DDL §2)
-- community_schedule_rsvps         (DDL §3)
-- community_schedule_exceptions    (DDL §4)
-- + indexes, urutan sesuai FK
```

### `migrations/2026070300_create_community_announcements.down.sql`
```sql
DROP TABLE IF EXISTS community_announcements CASCADE;
```

### `migrations/2026070301_create_community_schedule.down.sql`
```sql
DROP TABLE IF EXISTS community_schedule_exceptions CASCADE;
DROP TABLE IF EXISTS community_schedule_rsvps CASCADE;
DROP TABLE IF EXISTS community_schedule_entries CASCADE;
```

> **Trigger `updated_at`:** pakai fungsi `update_timestamp()` yang sudah ada (pola sama dengan modul lain). Pasang `BEFORE UPDATE` di keempat tabel.
>
> **Permission seed (di modul Role-Permission, bukan di sini):** insert `manage_community_announcement` & `manage_community_schedule` ke `permissions`, lalu tambahkan ke `community_role_permissions_template` untuk role leader.
>
> **Cleanup komunitas:** karena semua FK `ON DELETE CASCADE` ke `communities(id)`, penghapusan komunitas otomatis membersihkan pengumuman, agenda, RSVP, dan exception. Konsisten dengan Rule 7 di RULES.

### Migrasi susulan: `timezone` (2026-07-17)

Migrasi nyata di repo: `internal/migrations/20260717080434_add_community_schedule_entry_timezone.up.sql`:
```sql
ALTER TABLE community_schedule_entries ADD COLUMN timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Jakarta';
```
Additive, backfill otomatis via `DEFAULT` — row lama tidak perlu migrasi data terpisah. Lihat §Catatan Occurrence & Recurrence untuk alasan kolom ini dibutuhkan.

---

## Enum Values

| Tabel | Kolom | Nilai |
|---|---|---|
| `community_announcements` | `priority` | `normal`, `important` |
| `community_announcements` | `status` | `draft`, `published`, `archived` |
| `community_schedule_entries` | `status` | `active`, `cancelled` |
| `community_schedule_rsvps` | `response` | `going`, `maybe`, `not_going` |
| `community_schedule_exceptions` | `type` | `cancelled` (P1), `modified` (P2) |

---

## Query Examples

### Q1: Feed pengumuman untuk member
```sql
SELECT *
FROM community_announcements
WHERE community_id = $1
  AND status = 'published'
  AND (expires_at IS NULL OR expires_at > NOW())
ORDER BY is_pinned DESC,
         (priority = 'important') DESC,
         published_at DESC
LIMIT $2 OFFSET $3;
```

### Q2: Ambil agenda aktif dalam window (sebelum expand recurrence)
```sql
-- Ambil semua entry yang MUNGKIN punya occurrence dalam [from, to].
-- One-off: start_at dalam window. Recurring: start_at <= to (occurrence di-expand di app).
SELECT *
FROM community_schedule_entries
WHERE community_id = $1
  AND status = 'active'
  AND start_at <= $3                                  -- to
  AND (recurrence IS NOT NULL OR start_at >= $2);     -- one-off dibatasi from
```

### Q3: Exception (tanggal batal) untuk sekumpulan entry dalam window
```sql
SELECT entry_id, occurrence_date, type
FROM community_schedule_exceptions
WHERE entry_id = ANY($1)
  AND occurrence_date BETWEEN $2 AND $3;
```

### Q4: RSVP milik user untuk occurrence dalam window
```sql
SELECT entry_id, occurrence_date, response
FROM community_schedule_rsvps
WHERE user_id = $1
  AND entry_id = ANY($2)
  AND occurrence_date BETWEEN $3 AND $4;
```

### Q5: Ringkasan RSVP satu occurrence (untuk pengelola)
```sql
SELECT response, COUNT(*) AS total
FROM community_schedule_rsvps
WHERE entry_id = $1 AND occurrence_date = $2
GROUP BY response;
```

### Q6: Upsert RSVP (ubah/tambah)
```sql
INSERT INTO community_schedule_rsvps (entry_id, occurrence_date, user_id, response)
VALUES ($1, $2, $3, $4)
ON CONFLICT (entry_id, occurrence_date, user_id)
DO UPDATE SET response = EXCLUDED.response, updated_at = NOW()
RETURNING *;
```

---

## Catatan Occurrence & Recurrence

Bagian ini menegaskan cara occurrence dirakit (logika di service layer, bukan DB).

1. **Expand dibatasi window.** Endpoint kalender wajib menerima `from`/`to`. Untuk tiap entry recurring, expand `recurrence` (pakai library RRULE, mis. `github.com/teambition/rrule-go`) hanya dalam rentang itu. Recurring tanpa `UNTIL`/`COUNT` tetap aman karena dipotong window.

2. **`occurrence_date` = DATE.** Waktu jam diambil dari `entries.start_at` (dan `all_day` bila true). Satu tanggal cukup untuk mengidentifikasi occurrence secara unik dalam satu entry.
   - **Ditambahkan 2026-07-17**: kombinasi ulang `occurrence_date` + jam dari `start_at` WAJIB dilakukan di zona `entries.timezone` — `start_at` (TIMESTAMPTZ) selalu scan balik dari Postgres dengan Location UTC, jadi membaca komponen jamnya langsung (tanpa konversi eksplisit ke `entries.timezone` dulu) menghasilkan jam yang salah untuk entry dengan jam lokal yang menyebrang tengah malam UTC. Implementasi: `entry.StartAt.In(loc)` dulu baru ambil `.Hour()/.Minute()/.Second()`, bukan baca langsung dari `entry.StartAt` dalam Location aslinya.

3. **Merge exception & RSVP.** Setelah dapat daftar occurrence:
   - Tandai `is_cancelled` bila ada exception `type='cancelled'` pada `(entry_id, occurrence_date)`, atau bila `entry.status='cancelled'`.
   - Sisipkan `MyResponse` dari `community_schedule_rsvps` untuk user berjalan.
   - Untuk pengelola, isi `RsvpSummary` via Q5 (batch).

4. **Aturan RSVP.** Tolak RSVP bila occurrence: (a) sudah lewat (`occurrence_date` + time-of-day < now), atau (b) `is_cancelled`. Enforce di service, bukan hanya di client.

5. **Phase 2 — `type='modified'`.** Kolom `override_*` sudah ada tapi tidak diisi di Phase 1. Saat aktif nanti, occurrence yang punya exception `modified` memakai nilai override menggantikan nilai entry — tanpa perlu ubah schema.

---

*Dokumen ini adalah schema, domain struct, dan migration reference untuk sub-fitur Papan Pengumuman & Schedule modul Community. Untuk aturan bisnis lihat `COMMUNITY_ANNOUNCEMENT_SCHEDULE_RULES.md`. Untuk endpoint lihat API spec mobile & backoffice.*
