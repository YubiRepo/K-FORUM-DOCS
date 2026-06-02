# API Specification — Content Report (Backoffice)

> **Dibuat:** 2026-06-02
> **Stack:** Web Backoffice (Superadmin & Moderator komunitas)
> **Base URL Prefix:** `/api/v1/web/reports/content`

Endpoint backoffice untuk **memproses laporan konten/user**: antrian moderasi, peninjauan, dan resolusi (action_taken / dismissed). Berdasarkan `REPORTING_RULES.md` (Bagian A) & `REPORTING_DB_SCHEMA.md`. Pengiriman laporan oleh member ada di `API_SPEC_CONTENT_REPORT_MOBILE.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Data Models](#data-models)
3. [Endpoints](#endpoints)
   - [B1. List Reports (Queue)](#b1-list-reports-queue)
   - [B2. Get Report Detail](#b2-get-report-detail)
   - [B3. Get Reports by Target](#b3-get-reports-by-target)
   - [B4. Claim / Start Review](#b4-claim--start-review)
   - [B5. Resolve Report](#b5-resolve-report)
   - [B6. Bulk Resolve](#b6-bulk-resolve)
   - [B7. Reporter Trust Info](#b7-reporter-trust-info)
   - [B8. Report Stats](#b8-report-stats)
4. [Status Code Reference](#status-code-reference)
5. [Notification Triggers](#notification-triggers)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/web/reports/content`
- **Headers Global:**
  ```
  Content-Type: application/json
  Accept: application/json
  Authorization: Bearer <access_token>
  ```

- **Autentikasi & Scope:**

| Aktor | Akses |
|---|---|
| Superadmin | Semua laporan (global & semua komunitas) |
| Moderator/Leader komunitas (`manage_reports`) | **Hanya** laporan scope komunitasnya (`community_id` cocok) |

  - Request tanpa token / tanpa hak → `401` / `403`.
  - Laporan global (`community_id = null`) **hanya** dapat diakses superadmin.

> **Prinsip:** resolusi `action_taken` **memanggil endpoint moderasi yang sudah ada** (mis. remove post / ban user di Community backoffice). Endpoint di sini mencatat keputusan & audit, lalu memicu aksi tersebut — bukan menduplikasi logikanya.

---

## Data Models

### ReportListObject
```json
{
  "id": "uuid",
  "reportable_type": "community_post",
  "reportable_id": "uuid",
  "community": { "id": "uuid", "name": "Komunitas WNI Seoul" },
  "reason": "harassment",
  "status": "pending",
  "target_report_count": 6,
  "reporter": { "id": "uuid", "name": "Andi" },
  "created_at": "2026-06-01T10:00:00.000Z"
}
```
> `community` = `null` untuk laporan global. `target_report_count` = total laporan pada target yang sama (prioritisasi).

### ReportDetailObject
```json
{
  "id": "uuid",
  "reportable_type": "community_post",
  "reportable_id": "uuid",
  "community": { "id": "uuid", "name": "Komunitas WNI Seoul" },
  "reason": "harassment",
  "detail": "Berisi penghinaan ke anggota lain.",
  "status": "reviewing",
  "resolution": null,
  "resolution_note": null,
  "reporter": { "id": "uuid", "name": "Andi", "email": "andi@example.com" },
  "reviewed_by": null,
  "reviewed_at": null,
  "target_snapshot": {
    "type": "community_post",
    "author": { "id": "uuid", "name": "Budi" },
    "content": "Isi konten yang dilaporkan...",
    "status": "published",
    "created_at": "2026-05-30T09:00:00.000Z"
  },
  "target_report_count": 6,
  "created_at": "2026-06-01T10:00:00.000Z"
}
```
> `target_snapshot` = ringkasan konten target (di-resolve server sesuai `reportable_type`) agar peninjau tidak perlu memanggil modul lain.

### ReporterTrustObject
```json
{
  "user": { "id": "uuid", "name": "Andi" },
  "total_reports": 40,
  "resolved": 30,
  "dismissed": 24,
  "dismiss_ratio": 0.8,
  "flag": "high_false_rate"
}
```
> `flag`: `none` | `high_false_rate` (rasio dismiss tinggi → pelapor patut diragukan).

### PaginationObject
```json
{ "limit": 20, "offset": 0, "total": 85, "has_next": true, "has_prev": false }
```

### Error Responses
```json
// 400 / 403 / 404 / 409
{ "message": "Deskripsi error" }

// 422 Validation Error
{ "message": "Data input tidak valid", "errors": { "resolution": ["Resolusi wajib dipilih"] } }
```

---

## Endpoints

### B1. List Reports (Queue)

Antrian laporan. Superadmin lihat semua; moderator otomatis difilter ke komunitasnya.

- **URL:** `GET /api/v1/web/reports/content`
- **Auth:** Superadmin atau `manage_reports`
- **Query Params:**

| Param | Type | Default | Keterangan |
|-------|------|---------|-----------|
| `status` | string | `pending` | `pending` \| `reviewing` \| `resolved` |
| `reportable_type` | string | — | Filter jenis target |
| `reason` | string | — | Filter kategori |
| `community_id` | string (UUID) | — | Superadmin: filter komunitas; moderator: diabaikan (dipaksa ke scope-nya) |
| `scope` | string | — | `community` \| `global` (superadmin) |
| `sort` | string | `oldest` | `oldest` \| `most_reported` (by target_report_count) |
| `limit` | int | 20 | Max 100 |
| `offset` | int | 0 | — |

- **Response 200:** array `ReportListObject` + `pagination`

---

### B2. Get Report Detail

- **URL:** `GET /api/v1/web/reports/content/{report_id}`
- **Auth:** Superadmin atau `manage_reports` (scope cocok)
- **Response 200:** `{ "data": { "...": "ReportDetailObject" } }`
- **Response 403:** `{ "message": "Laporan di luar scope Anda" }`
- **Response 404:** `{ "message": "Laporan tidak ditemukan" }`

---

### B3. Get Reports by Target

Semua laporan untuk satu target (lihat seluruh konteks sebelum memutuskan).

- **URL:** `GET /api/v1/web/reports/content/target`
- **Auth:** Superadmin atau `manage_reports`
- **Query Params:** `reportable_type` (**Yes**), `reportable_id` (**Yes**), `limit`, `offset`
- **Response 200:** array `ReportDetailObject` (tanpa target_snapshot berulang) + `pagination`

---

### B4. Claim / Start Review

Tandai laporan sedang ditinjau (`reviewing`) & catat peninjau.

- **URL:** `POST /api/v1/web/reports/content/{report_id}/review`
- **Auth:** Superadmin atau `manage_reports`
- **Side effects:** `status = reviewing`, `reviewed_by = current_user`.
- **Response 200:** `{ "message": "Laporan dalam peninjauan" }`
- **Response 409:** `{ "message": "Laporan sudah diselesaikan" }`

---

### B5. Resolve Report

Selesaikan laporan dengan keputusan. Bila `action_taken`, sertakan aksi moderasi yang dijalankan.

- **URL:** `POST /api/v1/web/reports/content/{report_id}/resolve`
- **Auth:** Superadmin atau `manage_reports` (scope cocok)
- **Request Body:**
```json
{
  "resolution": "action_taken",
  "action": "remove_content",
  "note": "Konten menghina, dihapus sesuai pedoman."
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `resolution` | string (enum) | **Yes** | `action_taken` \| `dismissed` |
| `action` | string (enum) | Conditional | Wajib bila `action_taken`: `remove_content` \| `ban_author` \| `suspend_community` \| `none` |
| `note` | string | No | Catatan resolusi (audit) |

- **Side effects:**
  - `status = resolved`, isi `resolution`/`resolution_note`/`reviewed_by`/`reviewed_at`.
  - Bila `action` ≠ `none`, **panggil endpoint moderasi terkait** (Community backoffice: remove post/comment, ban member, suspend community) sesuai `reportable_type`.
  - Laporan **lain** pada target yang sama dapat ikut otomatis di-resolve sebagai `action_taken` (opsional, lihat catatan).
  - Emit notifikasi `report_resolved` ke pelapor (opsional).
- **Response 200:** `{ "message": "Laporan diselesaikan" }`
- **Response 422:** `{ "message": "Data input tidak valid", "errors": { "action": ["Wajib diisi saat action_taken"] } }`

> **Catatan:** untuk target dengan banyak laporan, resolusi tunggal sebaiknya menutup seluruh laporan target itu agar antrian tidak menumpuk duplikat. Perilaku ini dapat dikontrol flag `cascade_target` (default `true`).

---

### B6. Bulk Resolve

Selesaikan banyak laporan sekaligus (mis. dismiss massal hasil spam-report).

- **URL:** `POST /api/v1/web/reports/content/bulk-resolve`
- **Auth:** Superadmin atau `manage_reports`
- **Request Body:**
```json
{ "report_ids": ["uuid1", "uuid2"], "resolution": "dismissed", "note": "Tidak melanggar." }
```
- **Response 200:** `{ "data": { "resolved": 2, "skipped": 0 }, "message": "Selesai" }`

---

### B7. Reporter Trust Info

Statistik kepercayaan seorang pelapor (deteksi penyalahgunaan).

- **URL:** `GET /api/v1/web/reports/content/reporters/{user_id}`
- **Auth:** Superadmin
- **Response 200:** `{ "data": { "...": "ReporterTrustObject" } }`

---

### B8. Report Stats

Ringkasan metrik antrian (panel admin).

- **URL:** `GET /api/v1/web/reports/content/stats`
- **Auth:** Superadmin atau `manage_reports`
- **Query Params:** `period` (`7d`/`30d`/`90d`, default `30d`), `community_id` (opsional)
- **Response 200:**
```json
{
  "data": {
    "pending": 18,
    "reviewing": 4,
    "resolved": 240,
    "action_taken": 150,
    "dismissed": 90,
    "by_reason": { "spam": 60, "harassment": 40, "hate_speech": 25, "other": 30 },
    "avg_resolution_hours": 6.4,
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
| 403 | Bukan superadmin / di luar scope moderator |
| 404 | Laporan tidak ditemukan |
| 409 | Konflik status (mis. resolve laporan yang sudah resolved) |
| 422 | Validation error |

---

## Notification Triggers

| Aksi | Notifikasi |
|---|---|
| Target tembus ambang auto-flag | Ke penangan (moderator/superadmin) |
| B5 Resolve (`report_resolved`) | Ke pelapor (opsional, sesuai preferensi) |
| B5 `action_taken` (remove/ban) | Mengikuti notifikasi modul moderasi terkait (mis. ke author konten) |

> Semua event dikirim ke modul Notification sesuai `notification-preferences-technical.md`.

---

*Selaras dengan `REPORTING_RULES.md`, `REPORTING_DB_SCHEMA.md`, dan `API_SPEC_CONTENT_REPORT_MOBILE.md`.*
