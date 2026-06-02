# API Specification — Content Report (Mobile Client)

Endpoint mobile untuk **pelaporan konten/user**. Member melaporkan konten yang melanggar; laporan masuk antrian moderasi (ditangani moderator komunitas / superadmin di backoffice).

Berdasarkan `REPORTING_RULES.md` (Bagian A) dan `REPORTING_DB_SCHEMA.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Data Models](#data-models)
3. [Endpoints](#endpoints)
   - [1. Get Report Reasons](#1-get-report-reasons)
   - [2. Submit Content Report](#2-submit-content-report)
   - [3. Get My Reports](#3-get-my-reports)
4. [Status Code Reference](#status-code-reference)
5. [Notes & Best Practices](#notes--best-practices)
6. [User Flow](#user-flow)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/mobile/reports/content`
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
| GET reasons | ✅ Required | member (Standard / Pro) |
| POST (submit) | ✅ Required | member; rate-limited |
| GET mine | ✅ Required | member (hanya laporan sendiri) |

> Member **tidak** bisa melihat laporan orang lain maupun memprosesnya. Pemrosesan dilakukan di backoffice.

---

## Data Models

### ReasonObject
```json
{
  "value": "harassment",
  "label": "Pelecehan / perundungan",
  "requires_detail": false
}
```
> `requires_detail = true` untuk `other` (wajib isi `detail`).

### ReportableType (enum, dikirim client)
`community_post`, `community_comment`, `community`, `qna_question`, `qna_answer`, `directory_listing`, `event`, `announcement`, `news_post`, `news_comment`, `user`

### MyReportObject
```json
{
  "id": "uuid",
  "reportable_type": "community_post",
  "reportable_id": "uuid",
  "reason": "spam",
  "detail": null,
  "status": "resolved",
  "resolution": "action_taken",
  "created_at": "2026-06-01T10:00:00.000Z",
  "reviewed_at": "2026-06-01T15:00:00.000Z"
}
```
> `resolution` & `reviewed_at` = `null` selama belum `resolved`.

### PaginationObject
```json
{ "limit": 20, "offset": 0, "total": 12, "has_next": false, "has_prev": false }
```

### Error Responses
```json
// 400 / 403 / 404 / 409 / 429
{ "message": "Deskripsi error" }

// 422 Validation Error
{
  "message": "Data input tidak valid",
  "errors": { "reason": ["Alasan wajib dipilih"], "detail": ["Wajib diisi untuk alasan lainnya"] }
}
```

---

## Endpoints

### 1. Get Report Reasons

Daftar kategori alasan untuk ditampilkan di sheet "Laporkan". Label mengikuti `Accept-Language`.

- **URL:** `GET /api/v1/mobile/reports/content/reasons`
- **Auth:** Required (member)
- **Response 200:**
```json
{
  "data": [
    { "value": "spam", "label": "Spam / promosi berlebihan", "requires_detail": false },
    { "value": "harassment", "label": "Pelecehan / perundungan", "requires_detail": false },
    { "value": "hate_speech", "label": "Ujaran kebencian / SARA", "requires_detail": false },
    { "value": "sexual_content", "label": "Konten seksual", "requires_detail": false },
    { "value": "violence", "label": "Kekerasan / ancaman", "requires_detail": false },
    { "value": "misinformation", "label": "Informasi menyesatkan", "requires_detail": false },
    { "value": "scam", "label": "Penipuan", "requires_detail": false },
    { "value": "impersonation", "label": "Pemalsuan identitas", "requires_detail": false },
    { "value": "other", "label": "Lainnya", "requires_detail": true }
  ]
}
```

---

### 2. Submit Content Report

Laporkan satu target. Satu user maksimal satu laporan per target.

- **URL:** `POST /api/v1/mobile/reports/content`
- **Auth:** Required (member) — rate-limited (default maks 20/hari)
- **Request Body:**
```json
{
  "reportable_type": "community_post",
  "reportable_id": "uuid",
  "reason": "harassment",
  "detail": "Berisi penghinaan ke anggota lain."
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `reportable_type` | string (enum) | **Yes** | Jenis target |
| `reportable_id` | string (UUID) | **Yes** | ID target |
| `reason` | string (enum) | **Yes** | Salah satu kategori |
| `detail` | string | Conditional | Wajib bila `reason = other` |

- **Side effects:** insert report (`status=pending`), set `community_id` otomatis bila target berada dalam komunitas, increment `report_count` target, cek ambang auto-flag.
- **Response 201:**
```json
{ "data": { "id": "uuid", "status": "pending" }, "message": "Laporan terkirim, terima kasih" }
```
- **Response 404:** `{ "message": "Konten yang dilaporkan tidak ditemukan" }`
- **Response 409:** `{ "message": "Anda sudah melaporkan konten ini" }`
- **Response 422:** validation error (mis. `detail` kosong untuk `other`).
- **Response 429:** `{ "message": "Terlalu banyak laporan hari ini, coba lagi nanti" }`

---

### 3. Get My Reports

Riwayat laporan milik user sendiri (transparansi status).

- **URL:** `GET /api/v1/mobile/reports/content/mine`
- **Auth:** Required (member)
- **Query Params:** `status` (opsional: `pending`/`reviewing`/`resolved`), `limit` (default 20, max 50), `offset`
- **Response 200:** array `MyReportObject` + `pagination`

---

## Status Code Reference

| Code | Makna |
|------|-------|
| 200 | OK |
| 201 | Laporan dibuat |
| 400 | Bad request |
| 401 | Token tidak valid / kedaluwarsa |
| 403 | Tidak diizinkan |
| 404 | Target tidak ditemukan |
| 409 | Sudah pernah melaporkan target ini |
| 422 | Validation error |
| 429 | Rate limit terlampaui |

---

## Notes & Best Practices

1. **Satu laporan per target.** UI sebaiknya menonaktifkan tombol "Laporkan" jika user sudah melaporkan target tersebut (server tetap menolak dengan `409`).
2. **`community_id` ditentukan server.** Client tidak mengirimnya; backend menurunkannya dari target agar routing scope akurat.
3. **Identitas pelapor dirahasiakan.** Tidak ada endpoint yang membuka siapa yang melaporkan suatu konten ke publik/pemilik konten.
4. **Konteks `other`.** Jika `reason = other`, paksa pengisian `detail` di UI sebelum submit.
5. **Tanpa janji aksi.** Setelah submit, jangan menjanjikan tindakan spesifik; status laporan bisa dipantau via endpoint #3.

---

## User Flow

```
Lihat konten → tap "Laporkan"
  → Get reasons (1) → pilih reason (+ detail jika 'other')
  → Submit (2) → 201 / 409 (sudah pernah)
Pantau status: My Reports (3) → pending → reviewing → resolved
```

---

*Pemrosesan laporan ada di `API_SPEC_CONTENT_REPORT_BACKOFFICE.md`. Selaras dengan `REPORTING_RULES.md` & `REPORTING_DB_SCHEMA.md`.*
