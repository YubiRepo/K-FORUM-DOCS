# API Specification — Bug Report (Backoffice)

> **Dibuat:** 2026-06-02
> **Stack:** Web Backoffice (Support / Dev / Superadmin)
> **Base URL Prefix:** `/api/v1/web/reports/bug`

Endpoint backoffice untuk **triage & penanganan bug report**: antrian, klasifikasi, assign, update status, dan slot integrasi issue tracker (Linear/Jira). Berdasarkan `REPORTING_RULES.md` (Bagian B) & `REPORTING_DB_SCHEMA.md`. Pengiriman oleh member ada di `API_SPEC_BUG_REPORT_MOBILE.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Data Models](#data-models)
3. [Endpoints](#endpoints)
   - [B1. List Bug Reports (Queue)](#b1-list-bug-reports-queue)
   - [B2. Get Bug Report Detail](#b2-get-bug-report-detail)
   - [B3. Triage (Set Priority)](#b3-triage-set-priority)
   - [B4. Assign](#b4-assign)
   - [B5. Update Status](#b5-update-status)
   - [B6. Link External Issue](#b6-link-external-issue)
   - [B7. Bug Stats](#b7-bug-stats)
4. [Status Code Reference](#status-code-reference)
5. [Notification Triggers](#notification-triggers)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/web/reports/bug`
- **Headers Global:**
  ```
  Content-Type: application/json
  Accept: application/json
  Authorization: Bearer <access_token>
  ```

- **Autentikasi:** Superadmin atau role dengan permission `manage_bug_reports` (support/dev).
  - Request tanpa token / tanpa hak → `401` / `403`.

---

## Data Models

### BugReportListObject
```json
{
  "id": "uuid",
  "title": "Aplikasi crash saat buka feed komunitas",
  "category": "crash",
  "severity": "high",
  "priority": "urgent",
  "status": "in_progress",
  "reporter": { "id": "uuid", "name": "Andi" },
  "assignee": { "id": "uuid", "name": "Dev Rina" },
  "platform": "android",
  "external_issue_id": "ENG-1421",
  "created_at": "2026-06-01T10:00:00.000Z"
}
```
> `priority`, `assignee`, `external_issue_id` = `null` bila belum diisi.

### BugReportDetailObject
```json
{
  "id": "uuid",
  "title": "Aplikasi crash saat buka feed komunitas",
  "description": "Setiap buka tab komunitas, aplikasi langsung tertutup.",
  "steps_to_reproduce": "1. Login\n2. Buka tab Komunitas\n3. Crash",
  "category": "crash",
  "severity": "high",
  "priority": "urgent",
  "attachments": ["https://cdn/.../shot1.jpg"],
  "context": {
    "app_version": "2.3.1",
    "platform": "android",
    "os_version": "14",
    "device_model": "Samsung S23",
    "screen": "/communities/feed"
  },
  "status": "in_progress",
  "reporter": { "id": "uuid", "name": "Andi", "email": "andi@example.com" },
  "assignee": { "id": "uuid", "name": "Dev Rina" },
  "admin_note": "Sedang diinvestigasi tim.",
  "external_issue_id": "ENG-1421",
  "external_issue_url": "https://linear.app/.../ENG-1421",
  "created_at": "2026-06-01T10:00:00.000Z",
  "updated_at": "2026-06-01T12:00:00.000Z",
  "resolved_at": null
}
```

### PaginationObject
```json
{ "limit": 20, "offset": 0, "total": 85, "has_next": true, "has_prev": false }
```

### Error Responses
```json
// 400 / 403 / 404 / 409
{ "message": "Deskripsi error" }

// 422 Validation Error
{ "message": "Data input tidak valid", "errors": { "status": ["Status tidak valid"] } }
```

---

## Endpoints

### B1. List Bug Reports (Queue)

- **URL:** `GET /api/v1/web/reports/bug`
- **Auth:** Superadmin atau `manage_bug_reports`
- **Query Params:**

| Param | Type | Default | Keterangan |
|-------|------|---------|-----------|
| `status` | string | — | `new`/`triaged`/`in_progress`/`resolved`/`wont_fix`/`closed` |
| `category` | string | — | Filter kategori |
| `severity` | string | — | Filter severity |
| `priority` | string | — | Filter priority |
| `platform` | string | — | `ios`/`android`/`web` |
| `assigned_to` | string (UUID) | — | Filter assignee (`unassigned` untuk yang kosong) |
| `q` | string | — | Cari judul/deskripsi |
| `sort` | string | `severity` | `severity` (kritis dulu) \| `newest` |
| `limit` | int | 20 | Max 100 |
| `offset` | int | 0 | — |

- **Response 200:** array `BugReportListObject` + `pagination`

---

### B2. Get Bug Report Detail

- **URL:** `GET /api/v1/web/reports/bug/{report_id}`
- **Auth:** Superadmin atau `manage_bug_reports`
- **Response 200:** `{ "data": { "...": "BugReportDetailObject" } }`
- **Response 404:** `{ "message": "Laporan tidak ditemukan" }`

---

### B3. Triage (Set Priority)

Klasifikasi awal: tetapkan `priority` dan pindahkan ke `triaged`.

- **URL:** `POST /api/v1/web/reports/bug/{report_id}/triage`
- **Auth:** Superadmin atau `manage_bug_reports`
- **Request Body:**
```json
{ "priority": "urgent", "note": "Crash reproducible, prioritas tinggi." }
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `priority` | string (enum) | **Yes** | `low`/`medium`/`high`/`urgent` |
| `note` | string | No | Catatan triage (masuk `admin_note`) |

- **Side effects:** `status = triaged` (jika sebelumnya `new`), isi `priority`.
- **Response 200:** `{ "message": "Laporan ditriage" }`

---

### B4. Assign

- **URL:** `POST /api/v1/web/reports/bug/{report_id}/assign`
- **Auth:** Superadmin atau `manage_bug_reports`
- **Request Body:** `{ "assignee_id": "uuid" }` (kirim `null` untuk unassign)
- **Side effects:** isi `assigned_to`; status boleh otomatis ke `in_progress` bila masih `triaged` (opsional).
- **Response 200:** `{ "message": "Laporan ditugaskan" }`
- **Response 422:** `{ "message": "Assignee bukan staff yang valid" }`

---

### B5. Update Status

- **URL:** `PATCH /api/v1/web/reports/bug/{report_id}/status`
- **Auth:** Superadmin atau `manage_bug_reports`
- **Request Body:**
```json
{ "status": "resolved", "note": "Fix dirilis di v2.3.2." }
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `status` | string (enum) | **Yes** | `triaged`/`in_progress`/`resolved`/`wont_fix`/`closed` |
| `note` | string | No | Masuk `admin_note` (terlihat pelapor) |

- **Side effects:** update `status`; isi `resolved_at` saat masuk `resolved`/`wont_fix`/`closed`; emit notifikasi ke pelapor (opsional).
- **Transisi tidak valid** (mis. dari `closed` balik ke `new`) → `409`.
- **Response 200:** `{ "message": "Status diperbarui" }`

---

### B6. Link External Issue

Hubungkan ke issue Linear/Jira (Fase 2 — slot sudah tersedia).

- **URL:** `POST /api/v1/web/reports/bug/{report_id}/external-issue`
- **Auth:** Superadmin atau `manage_bug_reports`
- **Request Body:**
```json
{ "external_issue_id": "ENG-1421", "external_issue_url": "https://linear.app/.../ENG-1421" }
```
- **Side effects:** simpan `external_issue_id`/`external_issue_url`.
- **Response 200:** `{ "message": "Issue eksternal tertaut" }`

> Saat integrasi penuh aktif, endpoint ini dapat diperluas untuk **membuat** issue otomatis (bukan sekadar menautkan ID yang sudah ada) dan menyinkronkan status balik via webhook.

---

### B7. Bug Stats

- **URL:** `GET /api/v1/web/reports/bug/stats`
- **Auth:** Superadmin atau `manage_bug_reports`
- **Query Params:** `period` (`7d`/`30d`/`90d`, default `30d`)
- **Response 200:**
```json
{
  "data": {
    "new": 12,
    "in_progress": 8,
    "resolved": 140,
    "wont_fix": 20,
    "by_category": { "crash": 40, "ui": 30, "performance": 25, "data": 15, "auth": 10, "other": 20 },
    "by_severity": { "critical": 10, "high": 35, "medium": 60, "low": 35 },
    "by_platform": { "android": 90, "ios": 60, "web": 30 },
    "avg_resolution_hours": 48.2,
    "period": "30d"
  }
}
```

---

## Status Code Reference

| Code | Makna |
|------|-------|
| 200 | OK |
| 400 | Bad request |
| 401 | Token tidak valid / kedaluwarsa |
| 403 | Tidak punya izin `manage_bug_reports` |
| 404 | Laporan tidak ditemukan |
| 409 | Transisi status tidak valid |
| 422 | Validation error |

---

## Notification Triggers

| Aksi | Notifikasi |
|---|---|
| B4 Assign | Ke assignee: bug ditugaskan |
| B5 Update Status → `resolved`/`wont_fix` | Ke pelapor (opsional): hasil penanganan |
| B6 Link External Issue | — (internal) |

> Semua event dikirim ke modul Notification sesuai `notification-preferences-technical.md`.

---

*Selaras dengan `REPORTING_RULES.md`, `REPORTING_DB_SCHEMA.md`, dan `API_SPEC_BUG_REPORT_MOBILE.md`.*
