# Database Schema — Reporting System

> **Stack:** Golang + PostgreSQL
> **Berdasarkan:** `REPORTING_RULES.md`
> **Dibuat:** 2026-06-02

Dua subsistem terpisah: `content_reports` (pelaporan konten/user) dan `bug_reports` (pelaporan teknis).

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [PostgreSQL DDL Schema](#postgresql-ddl-schema)
   - [1. content_reports](#1-content_reports)
   - [2. bug_reports](#2-bug_reports)
   - [3. Denormalized report_count](#3-denormalized-report_count)
3. [Enum Values](#enum-values)
4. [Golang Structs](#golang-structs)
5. [Sample Queries](#sample-queries)
6. [Catatan Integrasi](#catatan-integrasi)

---

## Overview Relasi

```
users (auth module)
  ├── content_reports.reporter_id
  ├── content_reports.reviewed_by
  ├── bug_reports.reporter_id
  └── bug_reports.assigned_to

content_reports (polymorphic)
  ├── reportable_type + reportable_id  → target konten/user (tanpa FK keras)
  └── community_id (nullable)          → routing scope komunitas (FK communities)

bug_reports
  └── external_issue_id / external_issue_url → Linear/Jira (Fase 2, tanpa FK)

-- Counter denormalized hidup di tabel target masing-masing (lihat bagian 3):
community_posts.report_count, community_post_comments.report_count, dst.
```

> **Polymorphic tanpa FK keras:** `reportable_id` tidak memakai foreign key karena bisa menunjuk ke banyak tabel. Validasi keberadaan target dilakukan di service layer berdasarkan `reportable_type`.

---

## PostgreSQL DDL Schema

### 1. `content_reports`

```sql
CREATE TABLE content_reports (
  id              UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  reporter_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reportable_type VARCHAR(40)  NOT NULL,   -- community_post | community_comment | community | qna_question | ...
  reportable_id   UUID         NOT NULL,
  community_id    UUID         NULL REFERENCES communities(id) ON DELETE CASCADE, -- routing scope komunitas
  reason          VARCHAR(30)  NOT NULL,   -- spam | harassment | hate_speech | ...
  detail          TEXT         NULL,
  status          VARCHAR(20)  NOT NULL DEFAULT 'pending', -- pending | reviewing | resolved
  resolution      VARCHAR(20)  NULL,       -- action_taken | dismissed (saat resolved)
  resolution_note TEXT         NULL,
  reviewed_by     UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at     TIMESTAMPTZ  NULL,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  -- Satu user maks 1 report per target
  UNIQUE (reporter_id, reportable_type, reportable_id)
);

-- Antrian per target (hitung agregat, lihat report untuk satu konten)
CREATE INDEX idx_content_reports_target
  ON content_reports (reportable_type, reportable_id);

-- Antrian penanganan (filter status), terlama dulu
CREATE INDEX idx_content_reports_queue
  ON content_reports (status, created_at ASC);

-- Routing scope komunitas
CREATE INDEX idx_content_reports_community
  ON content_reports (community_id, status)
  WHERE community_id IS NOT NULL;

-- Laporan oleh user (false-report tracking)
CREATE INDEX idx_content_reports_reporter
  ON content_reports (reporter_id, resolution);
```

### 2. `bug_reports`

```sql
CREATE TABLE bug_reports (
  id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  reporter_id        UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title              VARCHAR(200) NOT NULL,
  description        TEXT         NOT NULL,
  steps_to_reproduce TEXT         NULL,
  category           VARCHAR(20)  NOT NULL,  -- crash | ui | performance | data | auth | other
  severity           VARCHAR(20)  NOT NULL,  -- low | medium | high | critical (persepsi user)
  priority           VARCHAR(20)  NULL,      -- low | medium | high | urgent (diisi saat triage)
  attachments        JSONB        NULL,      -- array URL screenshot

  -- Konteks auto-captured oleh client
  app_version        VARCHAR(30)  NULL,
  platform           VARCHAR(15)  NULL,      -- ios | android | web
  os_version         VARCHAR(40)  NULL,
  device_model       VARCHAR(80)  NULL,
  screen             VARCHAR(120) NULL,      -- route/layar saat report

  status             VARCHAR(20)  NOT NULL DEFAULT 'new', -- new | triaged | in_progress | resolved | wont_fix | closed
  assigned_to        UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  admin_note         TEXT         NULL,

  -- Integrasi issue tracker (Fase 2)
  external_issue_id  VARCHAR(80)  NULL,
  external_issue_url VARCHAR(300) NULL,

  created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  resolved_at        TIMESTAMPTZ  NULL
);

-- Antrian triage (filter status & priority)
CREATE INDEX idx_bug_reports_queue    ON bug_reports (status, priority, created_at DESC);
CREATE INDEX idx_bug_reports_reporter ON bug_reports (reporter_id, created_at DESC);
CREATE INDEX idx_bug_reports_assignee ON bug_reports (assigned_to) WHERE assigned_to IS NOT NULL;
```

### 3. Denormalized `report_count`

Agar sort `most_reported` & auto-flag cepat, tabel target yang sering dilaporkan menyimpan counter. Tambahkan kolom berikut pada tabel terkait (contoh untuk Community):

```sql
ALTER TABLE community_posts          ADD COLUMN report_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE community_post_comments  ADD COLUMN report_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE communities              ADD COLUMN report_count INTEGER NOT NULL DEFAULT 0;
-- (modul lain — qna, directory, event, user — mengikuti pola sama bila perlu sort)
```

> Counter di-increment dalam transaksi yang sama dengan insert `content_reports`, dan menjadi sumber sort `most_reported` (mis. Community backoffice B1). Untuk target tanpa kebutuhan sort, jumlah cukup dihitung via query agregat dari `content_reports`.

---

## Enum Values

### `content_reports.reportable_type`
`community_post`, `community_comment`, `community`, `qna_question`, `qna_answer`, `directory_listing`, `event`, `announcement`, `news_post`, `news_comment`, `user`

### `content_reports.reason`
`spam`, `harassment`, `hate_speech`, `sexual_content`, `violence`, `misinformation`, `scam`, `impersonation`, `other`

### `content_reports.status`
| Value | Keterangan |
|---|---|
| `pending` | Baru masuk |
| `reviewing` | Sedang ditinjau |
| `resolved` | Selesai |

### `content_reports.resolution`
| Value | Keterangan |
|---|---|
| `action_taken` | Tindakan moderasi diambil |
| `dismissed` | Laporan ditolak (tidak melanggar) |

### `bug_reports.category`
`crash`, `ui`, `performance`, `data`, `auth`, `other`

### `bug_reports.severity` / `priority`
severity: `low`, `medium`, `high`, `critical` — priority: `low`, `medium`, `high`, `urgent`

### `bug_reports.status`
| Value | Keterangan |
|---|---|
| `new` | Baru masuk |
| `triaged` | Sudah diklasifikasi & diberi priority |
| `in_progress` | Sedang dikerjakan |
| `resolved` | Selesai diperbaiki |
| `wont_fix` | Diputuskan tidak diperbaiki |
| `closed` | Ditutup (final) |

### `bug_reports.platform`
`ios`, `android`, `web`

---

## Golang Structs

```go
type ContentReport struct {
    ID             string     `db:"id"              json:"id"`
    ReporterID     string     `db:"reporter_id"     json:"reporter_id"`
    ReportableType string     `db:"reportable_type" json:"reportable_type"`
    ReportableID   string     `db:"reportable_id"   json:"reportable_id"`
    CommunityID    *string    `db:"community_id"    json:"community_id"`
    Reason         string     `db:"reason"          json:"reason"`
    Detail         *string    `db:"detail"          json:"detail"`
    Status         string     `db:"status"          json:"status"`
    Resolution     *string    `db:"resolution"      json:"resolution"`
    ResolutionNote *string    `db:"resolution_note" json:"resolution_note"`
    ReviewedBy     *string    `db:"reviewed_by"     json:"reviewed_by"`
    ReviewedAt     *time.Time `db:"reviewed_at"     json:"reviewed_at"`
    CreatedAt      time.Time  `db:"created_at"      json:"created_at"`
    UpdatedAt      time.Time  `db:"updated_at"      json:"updated_at"`
}

type BugReport struct {
    ID               string     `db:"id"                 json:"id"`
    ReporterID       string     `db:"reporter_id"        json:"reporter_id"`
    Title            string     `db:"title"              json:"title"`
    Description      string     `db:"description"        json:"description"`
    StepsToReproduce *string    `db:"steps_to_reproduce" json:"steps_to_reproduce"`
    Category         string     `db:"category"           json:"category"`
    Severity         string     `db:"severity"           json:"severity"`
    Priority         *string    `db:"priority"           json:"priority"`
    Attachments      []string   `db:"attachments"        json:"attachments"`
    AppVersion       *string    `db:"app_version"        json:"app_version"`
    Platform         *string    `db:"platform"           json:"platform"`
    OSVersion        *string    `db:"os_version"         json:"os_version"`
    DeviceModel      *string    `db:"device_model"       json:"device_model"`
    Screen           *string    `db:"screen"             json:"screen"`
    Status           string     `db:"status"             json:"status"`
    AssignedTo       *string    `db:"assigned_to"        json:"assigned_to"`
    AdminNote        *string    `db:"admin_note"         json:"admin_note"`
    ExternalIssueID  *string    `db:"external_issue_id"  json:"external_issue_id"`
    ExternalIssueURL *string    `db:"external_issue_url" json:"external_issue_url"`
    CreatedAt        time.Time  `db:"created_at"         json:"created_at"`
    UpdatedAt        time.Time  `db:"updated_at"         json:"updated_at"`
    ResolvedAt       *time.Time `db:"resolved_at"        json:"resolved_at"`
}
```

---

## Sample Queries

```sql
-- Q1: Submit content report (cek dedup via UNIQUE; increment counter di transaksi sama)
INSERT INTO content_reports (reporter_id, reportable_type, reportable_id, community_id, reason, detail)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (reporter_id, reportable_type, reportable_id) DO NOTHING;
-- UPDATE community_posts SET report_count = report_count + 1 WHERE id = $3; (sesuai type)

-- Q2: Antrian report global (superadmin) — pending terlama dulu
SELECT * FROM content_reports
WHERE status = 'pending' AND community_id IS NULL
ORDER BY created_at ASC
LIMIT $1 OFFSET $2;

-- Q3: Antrian report scope satu komunitas (moderator)
SELECT * FROM content_reports
WHERE community_id = $1 AND status IN ('pending','reviewing')
ORDER BY created_at ASC;

-- Q4: Semua report untuk satu target (lihat konteks sebelum memutuskan)
SELECT cr.*, u.fullname AS reporter_name
FROM content_reports cr
JOIN users u ON u.id = cr.reporter_id
WHERE cr.reportable_type = $1 AND cr.reportable_id = $2
ORDER BY cr.created_at DESC;

-- Q5: Resolve report
UPDATE content_reports
SET status = 'resolved', resolution = $2, resolution_note = $3,
    reviewed_by = $4, reviewed_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- Q6: False-report ratio seorang user
SELECT
  COUNT(*) FILTER (WHERE resolution = 'dismissed')::float
    / NULLIF(COUNT(*) FILTER (WHERE status = 'resolved'), 0) AS dismiss_ratio
FROM content_reports
WHERE reporter_id = $1;

-- Q7: Target yang melewati ambang auto-flag (mis. komentar)
SELECT reportable_id, COUNT(*) AS total
FROM content_reports
WHERE reportable_type = $1 AND status <> 'resolved'
GROUP BY reportable_id
HAVING COUNT(*) >= $2;

-- Q8: Antrian bug report (triage) — kritis & terbaru dulu
SELECT * FROM bug_reports
WHERE status IN ('new','triaged','in_progress')
ORDER BY
  CASE severity WHEN 'critical' THEN 0 WHEN 'high' THEN 1 WHEN 'medium' THEN 2 ELSE 3 END,
  created_at DESC
LIMIT $1 OFFSET $2;

-- Q9: Bug report milik seorang user
SELECT * FROM bug_reports WHERE reporter_id = $1 ORDER BY created_at DESC;

-- Q10: Update status bug + isi resolved_at saat selesai
UPDATE bug_reports
SET status = $2, priority = COALESCE($3, priority), assigned_to = COALESCE($4, assigned_to),
    admin_note = $5,
    resolved_at = CASE WHEN $2 IN ('resolved','wont_fix','closed') THEN NOW() ELSE resolved_at END,
    updated_at = NOW()
WHERE id = $1;
```

---

## Catatan Integrasi

1. **Submit content report** = transaksi gabungan: insert `content_reports` + increment `report_count` di tabel target + cek ambang auto-flag.
2. **Resolusi `action_taken`** memanggil endpoint moderasi yang sudah ada (mis. remove post / ban user di Community backoffice). Sistem report tidak menduplikasi logika moderasi.
3. **Notification**: emit `report_resolved` (ke pelapor, opsional) & `report_threshold_reached` (ke penangan) sesuai `notification-preferences-technical.md`.
4. **Permission baru**: `manage_reports` (content, scope komunitas) & `manage_bug_reports` (bug) ditambahkan di Role-Permission module.
5. **Polymorphic validation**: service layer memvalidasi `reportable_id` benar-benar ada sesuai `reportable_type` sebelum insert.
6. **Issue tracker (Fase 2)**: `external_issue_id`/`external_issue_url` diisi saat bug dibuatkan issue di Linear/Jira; status disinkronkan via webhook.

---

*Dokumen ini selaras dengan `REPORTING_RULES.md`.*
