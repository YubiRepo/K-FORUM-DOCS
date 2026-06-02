# API Specification — Bug Report (Mobile Client)

Endpoint mobile untuk **pelaporan bug/masalah teknis**. Member mengirim laporan; ditangani tim support/dev di backoffice. Berdasarkan `REPORTING_RULES.md` (Bagian B) & `REPORTING_DB_SCHEMA.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Data Models](#data-models)
3. [Endpoints](#endpoints)
   - [1. Submit Bug Report](#1-submit-bug-report)
   - [2. Upload Attachment](#2-upload-attachment)
   - [3. Get My Bug Reports](#3-get-my-bug-reports)
   - [4. Get My Bug Report Detail](#4-get-my-bug-report-detail)
4. [Status Code Reference](#status-code-reference)
5. [Notes & Best Practices](#notes--best-practices)
6. [User Flow](#user-flow)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/mobile/reports/bug`
- **Headers Global:**
  ```
  Content-Type: application/json
  Accept: application/json
  Accept-Language: <lang_code>    (ko | id | en — default: ko)
  X-Locale: <lang_code>
  Authorization: Bearer <access_token>
  ```

- **Permission / Benefit:**

| Endpoint | Auth | Keterangan |
|----------|------|-----------|
| POST (submit) | ✅ Required | member (Standard / Pro) |
| POST attachment | ✅ Required | member |
| GET mine | ✅ Required | member (hanya laporan sendiri) |
| GET mine/:id | ✅ Required | member (hanya milik sendiri) |

---

## Data Models

### MyBugReportListObject
```json
{
  "id": "uuid",
  "title": "Aplikasi crash saat buka feed komunitas",
  "category": "crash",
  "severity": "high",
  "status": "in_progress",
  "created_at": "2026-06-01T10:00:00.000Z"
}
```

### MyBugReportDetailObject
```json
{
  "id": "uuid",
  "title": "Aplikasi crash saat buka feed komunitas",
  "description": "Setiap buka tab komunitas, aplikasi langsung tertutup.",
  "steps_to_reproduce": "1. Login\n2. Buka tab Komunitas\n3. Crash",
  "category": "crash",
  "severity": "high",
  "attachments": ["https://cdn/.../shot1.jpg"],
  "context": {
    "app_version": "2.3.1",
    "platform": "android",
    "os_version": "14",
    "device_model": "Samsung S23",
    "screen": "/communities/feed"
  },
  "status": "in_progress",
  "admin_note": "Sedang diinvestigasi tim.",
  "created_at": "2026-06-01T10:00:00.000Z",
  "resolved_at": null
}
```
> `admin_note` ditampilkan ke pelapor bila diisi (transparansi). Field internal (assigned_to, priority, external_issue_*) **tidak** dikembalikan ke mobile.

### AttachmentUploadObject
```json
{ "url": "https://cdn/.../shot1.jpg" }
```

### PaginationObject
```json
{ "limit": 20, "offset": 0, "total": 5, "has_next": false, "has_prev": false }
```

### Error Responses
```json
// 400 / 403 / 404 / 429
{ "message": "Deskripsi error" }

// 422 Validation Error
{
  "message": "Data input tidak valid",
  "errors": { "title": ["Judul wajib diisi"], "category": ["Kategori tidak valid"] }
}
```

---

## Endpoints

### 1. Submit Bug Report

Kirim laporan bug. Field `context` di-capture otomatis oleh client.

- **URL:** `POST /api/v1/mobile/reports/bug`
- **Auth:** Required (member) — rate-limited (mis. maks 10/hari)
- **Request Body:**
```json
{
  "title": "Aplikasi crash saat buka feed komunitas",
  "description": "Setiap buka tab komunitas, aplikasi langsung tertutup.",
  "steps_to_reproduce": "1. Login\n2. Buka tab Komunitas\n3. Crash",
  "category": "crash",
  "severity": "high",
  "attachments": ["https://cdn/.../shot1.jpg"],
  "context": {
    "app_version": "2.3.1",
    "platform": "android",
    "os_version": "14",
    "device_model": "Samsung S23",
    "screen": "/communities/feed"
  }
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `title` | string | **Yes** | Maks 200 karakter |
| `description` | string | **Yes** | Penjelasan masalah |
| `steps_to_reproduce` | string | No | Langkah memunculkan bug |
| `category` | string (enum) | **Yes** | `crash`/`ui`/`performance`/`data`/`auth`/`other` |
| `severity` | string (enum) | **Yes** | `low`/`medium`/`high`/`critical` |
| `attachments` | array(string) | No | URL hasil upload (maks 5) |
| `context` | object | No (disarankan) | Auto-captured: app_version, platform, os_version, device_model, screen |

- **Side effects:** insert bug report (`status=new`).
- **Response 201:**
```json
{ "data": { "id": "uuid", "status": "new" }, "message": "Terima kasih, laporan Anda kami terima" }
```
- **Response 422:** validation error.
- **Response 429:** `{ "message": "Terlalu banyak laporan, coba lagi nanti" }`

---

### 2. Upload Attachment

Upload screenshot sebelum submit. Kumpulkan URL lalu kirim di field `attachments`.

- **URL:** `POST /api/v1/mobile/reports/bug/attachments`
- **Auth:** Required (member)
- **Content-Type:** `multipart/form-data`
- **Form Field:** `file` (image; jpg/png/webp, maks 5MB)
- **Response 201:** `{ "data": { "url": "https://cdn/.../shot1.jpg" } }`
- **Response 422:** `{ "message": "Format / ukuran file tidak valid" }`

---

### 3. Get My Bug Reports

- **URL:** `GET /api/v1/mobile/reports/bug/mine`
- **Auth:** Required (member)
- **Query Params:** `status` (opsional), `limit` (default 20, max 50), `offset`
- **Response 200:** array `MyBugReportListObject` + `pagination`

---

### 4. Get My Bug Report Detail

- **URL:** `GET /api/v1/mobile/reports/bug/mine/{report_id}`
- **Auth:** Required (member) — hanya milik sendiri
- **Response 200:** `{ "data": { "...": "MyBugReportDetailObject" } }`
- **Response 403:** `{ "message": "Bukan laporan Anda" }`
- **Response 404:** `{ "message": "Laporan tidak ditemukan" }`

---

## Status Code Reference

| Code | Makna |
|------|-------|
| 200 | OK |
| 201 | Laporan / attachment dibuat |
| 400 | Bad request |
| 401 | Token tidak valid / kedaluwarsa |
| 403 | Bukan laporan milik user |
| 404 | Laporan tidak ditemukan |
| 422 | Validation error |
| 429 | Rate limit terlampaui |

---

## Notes & Best Practices

1. **Capture konteks otomatis.** Client sebaiknya mengisi `context` (versi app, OS, device, screen) tanpa membebani user mengetik — ini sangat membantu triage.
2. **Upload attachment dulu.** Sama seperti media di modul lain: upload tiap gambar via #2, baru sertakan URL di submit. Jangan kirim base64.
3. **Field internal disembunyikan.** `priority`, `assigned_to`, `external_issue_id/url` tidak pernah dikembalikan ke mobile.
4. **Transparansi terbatas.** Pelapor melihat `status` & `admin_note`, cukup untuk tahu progres tanpa membuka detail internal tim.

---

## User Flow

```
Temui bug → menu "Laporkan Masalah"
  → (opsional) Upload screenshot (2)
  → Isi judul/deskripsi/kategori/severity, context auto-terisi
  → Submit (1) → 201
Pantau: My Bug Reports (3) → Detail (4) → status new → triaged → in_progress → resolved
```

---

*Penanganan di `API_SPEC_BUG_REPORT_BACKOFFICE.md`. Selaras dengan `REPORTING_RULES.md` & `REPORTING_DB_SCHEMA.md`.*
