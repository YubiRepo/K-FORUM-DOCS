# API Spec — Community Announcement & Schedule (Mobile Client)

Dokumentasi API endpoint untuk sub-fitur **Papan Pengumuman** & **Schedule Komunitas** di aplikasi mobile (Flutter). Semua endpoint terikat sebuah komunitas (`community_id`) dan mengikuti aturan keanggotaan/visibility modul Community.

Untuk aturan bisnis lihat `COMMUNITY_ANNOUNCEMENT_SCHEDULE_RULES.md`. Untuk schema lihat `COMMUNITY_ANNOUNCEMENT_SCHEDULE_DB_SCHEMA.md`.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/communities/{community_id}`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Accept-Language: <lang_code>` (`ko` | `id` | `en`, default `ko`)
  - `X-Locale: <lang_code>`
  - `Authorization: Bearer <access_token>`
- **Authentication**: Wajib untuk semua endpoint di modul ini (komunitas private hanya untuk anggota; public boleh baca sesuai visibility Community, tapi RSVP tetap butuh keanggotaan).
- **Permission**: Aksi kelola dicek per komunitas via Role-Permission:
  - `manage_community_announcement` → buat/edit/hapus/pin pengumuman
  - `manage_community_schedule` → buat/edit/hapus agenda, batalin occurrence, lihat daftar RSVP

---

## Model Data Utama

### 1. Announcement Object
```json
{
  "id": "ann_001",
  "community_id": "comm_futsal",
  "author": { "id": "usr_01", "name": "Budi", "avatar": "https://cdn/.../a.jpg" },
  "title": "Latihan pindah lapangan",
  "body": "Mulai minggu depan latihan pindah ke GOR Senayan.",
  "media": [
    { "url": "https://cdn/.../1.jpg", "thumb_url": "https://cdn/.../1_t.jpg", "width": 1080, "height": 720, "order": 0 }
  ],
  "priority": "important",
  "is_pinned": true,
  "status": "published",
  "published_at": "2026-07-03T09:00:00.000Z",
  "expires_at": null,
  "created_at": "2026-07-03T08:55:00.000Z"
}
```
> `author` di-embed ringkas. `status`, `expires_at` hanya relevan buat pengelola; member selalu menerima yang `published` & belum expired.

### 2. Schedule Entry Object
```json
{
  "id": "sch_001",
  "community_id": "comm_futsal",
  "created_by": { "id": "usr_01", "name": "Budi" },
  "title": "Latihan rutin",
  "description": "Bawa jersey terang & gelap.",
  "location": "GOR Senayan Lapangan 3",
  "start_at": "2026-07-05T01:00:00.000Z",
  "end_at": "2026-07-05T03:00:00.000Z",
  "all_day": false,
  "recurrence": "FREQ=WEEKLY;BYDAY=SA",
  "timezone": "Asia/Jakarta",
  "status": "active",
  "created_at": "2026-07-03T08:00:00.000Z"
}
```
> `timezone` (ditambahkan 2026-07-17): IANA identifier (mis. `Asia/Jakarta`,
> `Asia/Makassar`, `Asia/Jayapura`) yang menentukan zona waktu agenda ini —
> **wajib** diisi creator secara eksplisit saat create (lihat B4), bukan
> diinfer diam-diam dari device. Dipakai backend untuk merekonstruksi
> occurrence (jam & tanggal) recurring dengan benar; tanpa ini, agenda
> dengan jam lokal yang menyebrang tengah malam UTC (mis. 03:00 WIB) bisa
> salah hari/jam saat di-generate ulang (bug yang sudah diperbaiki 2026-07-17).

### 3. Occurrence Object
Satu instance agenda pada satu tanggal. Hasil expand recurrence (dirakit backend). Ini yang dipakai buat render kalender.
```json
{
  "entry_id": "sch_001",
  "occurrence_date": "2026-07-12",
  "title": "Latihan rutin",
  "location": "GOR Senayan Lapangan 3",
  "start_at": "2026-07-12T01:00:00.000Z",
  "end_at": "2026-07-12T03:00:00.000Z",
  "all_day": false,
  "is_cancelled": false,
  "is_past": false,
  "my_response": "going",
  "rsvp_summary": { "going": 12, "maybe": 4, "not_going": 2 }
}
```
> `my_response`: `null` bila belum RSVP. `rsvp_summary` hanya diisi bila requester punya `manage_community_schedule` (member biasa dapat `going` count saja atau null — lihat catatan).

### 4. Error Responses
Standar sama seperti modul lain:
```json
{ "message": "Pesan error deskriptif" }
```
Validation (422):
```json
{ "message": "Data input tidak valid", "errors": { "title": ["Judul wajib diisi."] } }
```

---

# BAGIAN A — PAPAN PENGUMUMAN

## A1. List Pengumuman
Daftar pengumuman komunitas. Member menerima yang `published` & belum expired; pengelola bisa minta termasuk draft/archived.

- **URL**: `GET /announcements`
- **Autentikasi**: Yes
- **Query Parameters**:
  - `limit` (default 20, max 50), `offset`
  - `include` (opsional, khusus pengelola): `draft`, `archived` — diabaikan bila tak punya permission
- **Response (200)**:
  ```json
  {
    "data": [
      { "id": "ann_001", "title": "Latihan pindah lapangan", "priority": "important", "is_pinned": true, "published_at": "2026-07-03T09:00:00.000Z", "...": "..." }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 5 }
  }
  ```
> Urutan: pinned → priority `important` → `published_at` terbaru.

## A2. Detail Pengumuman
- **URL**: `GET /announcements/{announcement_id}`
- **Autentikasi**: Yes
- **Response (200)**: `{ "data": { Announcement Object } }`
- **Rules**: Member tidak bisa buka pengumuman `draft`/`archived`/expired → `404`.

## A3. Buat Pengumuman
- **URL**: `POST /announcements`
- **Autentikasi**: Yes + permission `manage_community_announcement`
- **Request Body**:
  ```json
  {
    "title": "Latihan pindah lapangan",
    "body": "Mulai minggu depan latihan pindah ke GOR Senayan.",
    "media": [ { "url": "https://cdn/.../1.jpg", "thumb_url": "https://cdn/.../1_t.jpg", "width": 1080, "height": 720, "order": 0 } ],
    "priority": "important",
    "is_pinned": true,
    "expires_at": null,
    "status": "published"
  }
  ```
  > `status` boleh `draft` atau `published`. `media` maks 5. Upload file dulu via `POST /api/v1/mobile/media/upload` (context `community_announcement`).
- **Response (201)**: `{ "data": { Announcement Object }, "message": "Announcement created" }`
- **Efek**: `status=published` + `priority=important` → push ke anggota (event `community_announcement_published`). `normal` → in-app/badge saja.

## A4. Edit Pengumuman
- **URL**: `PUT /announcements/{announcement_id}`
- **Autentikasi**: Yes + `manage_community_announcement`
- **Request Body**: field yang sama (partial update diperbolehkan).
- **Response (200)**: `{ "data": { Announcement Object }, "message": "Announcement updated" }`
- **Rules**: Mengubah `draft` → `published` memicu notif seperti A3. Publish→publish tidak mengirim ulang push.

## A5. Pin / Unpin
- **URL**: `POST /announcements/{announcement_id}/pin`
- **Autentikasi**: Yes + `manage_community_announcement`
- **Request Body**: `{ "is_pinned": true }`
- **Response (200)**: `{ "data": { "id": "ann_001", "is_pinned": true } }`

## A6. Arsip / Hapus Pengumuman
- **URL**: `DELETE /announcements/{announcement_id}`
- **Autentikasi**: Yes + `manage_community_announcement`
- **Query**: `mode` = `archive` (default) | `delete`
- **Response (200)**: `{ "message": "Announcement archived" }`
> Disarankan `archive` (set `status=archived`) daripada hard delete untuk jejak.

---

# BAGIAN B — SCHEDULE KOMUNITAS

## B1. List Occurrence (Kalender)
Endpoint utama kalender. Mengembalikan **occurrence** (bukan entry mentah) dalam rentang tanggal, sudah di-expand dari recurrence, dikurangi occurrence yang dibatalkan, dan disisipi RSVP requester.

- **URL**: `GET /schedule`
- **Autentikasi**: Yes
- **Query Parameters**:
  - `from` (wajib, `YYYY-MM-DD`) — awal window
  - `to` (wajib, `YYYY-MM-DD`) — akhir window
  - `include_cancelled` (opsional, default `true`) — sertakan occurrence batal (ditandai)
- **Response (200)**:
  ```json
  {
    "data": [
      { "entry_id": "sch_001", "occurrence_date": "2026-07-05", "title": "Latihan rutin", "start_at": "2026-07-05T01:00:00.000Z", "is_cancelled": false, "is_past": true, "my_response": "going" },
      { "entry_id": "sch_001", "occurrence_date": "2026-07-12", "title": "Latihan rutin", "start_at": "2026-07-12T01:00:00.000Z", "is_cancelled": false, "is_past": false, "my_response": null }
    ],
    "window": { "from": "2026-07-01", "to": "2026-07-31" }
  }
  ```
- **Rules**:
  - `from`/`to` wajib; rentang maks disarankan **≤ 92 hari** (cegah expand berlebihan) → `422` bila melebihi.
  - `rsvp_summary` hanya di-embed bila requester punya `manage_community_schedule`.

## B2. Detail Entry
Metadata agenda + aturan recurrence (buat halaman edit / lihat seri).

- **URL**: `GET /schedule/{entry_id}`
- **Autentikasi**: Yes
- **Response (200)**: `{ "data": { Schedule Entry Object } }`

## B3. Detail Occurrence
Detail satu tanggal termasuk ringkasan & (untuk pengelola) daftar peserta.

- **URL**: `GET /schedule/{entry_id}/occurrences/{occurrence_date}`
- **Autentikasi**: Yes
- **Response (200)**:
  ```json
  {
    "data": {
      "entry_id": "sch_001",
      "occurrence_date": "2026-07-12",
      "title": "Latihan rutin",
      "start_at": "2026-07-12T01:00:00.000Z",
      "is_cancelled": false,
      "is_past": false,
      "my_response": "going",
      "rsvp_summary": { "going": 12, "maybe": 4, "not_going": 2 }
    }
  }
  ```

## B4. Buat Agenda
- **URL**: `POST /schedule`
- **Autentikasi**: Yes + permission `manage_community_schedule`
- **Request Body**:
  ```json
  {
    "title": "Latihan rutin",
    "description": "Bawa jersey terang & gelap.",
    "location": "GOR Senayan Lapangan 3",
    "start_at": "2026-07-05T01:00:00.000Z",
    "end_at": "2026-07-05T03:00:00.000Z",
    "all_day": false,
    "recurrence": "FREQ=WEEKLY;BYDAY=SA",
    "timezone": "Asia/Jakarta"
  }
  ```
  > `recurrence` opsional; NULL = agenda one-off. `end_at` opsional (harus ≥ `start_at`).
  > `timezone`: **Wajib** (`binding:"required"` — 422 kalau kosong atau bukan IANA identifier valid). Client boleh prefill dari timezone device creator, tapi field ini harus selalu dikirim eksplisit, tidak pernah diinfer diam-diam di backend.
- **Response (201)**: `{ "data": { Schedule Entry Object }, "message": "Schedule created" }`

## B5. Edit Agenda
- **URL**: `PUT /schedule/{entry_id}`
- **Autentikasi**: Yes + `manage_community_schedule`
- **Request Body**: field sama dengan B4 (partial).
- **Response (200)**: `{ "data": { Schedule Entry Object }, "message": "Schedule updated" }`
> **Catatan:** edit agenda mengubah **seluruh series**. Ubah detail **satu occurrence** saja belum didukung Phase 1 (Phase 2).
> **Catatan `timezone` saat edit (beda dengan B4):** **opsional** di sini — kalau di-omit dari body, entry tetap pakai `timezone` yang sudah tersimpan (partial update, bukan reset ke default). Kirim field ini hanya kalau memang ingin mengubah zona agenda.

## B6. Batalin Satu Occurrence
- **URL**: `POST /schedule/{entry_id}/cancel-occurrence`
- **Autentikasi**: Yes + `manage_community_schedule`
- **Request Body**: `{ "occurrence_date": "2026-07-19" }`
- **Response (200)**: `{ "message": "Occurrence cancelled", "data": { "entry_id": "sch_001", "occurrence_date": "2026-07-19", "is_cancelled": true } }`
> Membuat `community_schedule_exceptions` type=cancelled. Reversible via B7.

## B7. Batalkan Pembatalan Occurrence
- **URL**: `DELETE /schedule/{entry_id}/cancel-occurrence`
- **Autentikasi**: Yes + `manage_community_schedule`
- **Request Body**: `{ "occurrence_date": "2026-07-19" }`
- **Response (200)**: `{ "message": "Occurrence restored" }`

## B8. Batalin / Hapus Agenda (Seluruh Series)
- **URL**: `DELETE /schedule/{entry_id}`
- **Autentikasi**: Yes + `manage_community_schedule`
- **Query**: `mode` = `cancel` (default, set status=cancelled) | `delete`
- **Response (200)**: `{ "message": "Schedule cancelled" }`

## B9. Set / Ubah RSVP
- **URL**: `PUT /schedule/{entry_id}/rsvp`
- **Autentikasi**: Yes (anggota aktif)
- **Request Body**:
  ```json
  { "occurrence_date": "2026-07-12", "response": "going" }
  ```
- **Response (200)**:
  ```json
  { "data": { "entry_id": "sch_001", "occurrence_date": "2026-07-12", "response": "going" }, "message": "RSVP saved" }
  ```
- **Rules**:
  - Occurrence harus **aktif & belum lewat** → `409` bila batal/lewat.
  - Upsert: RSVP kedua kali menimpa yang lama.
  - `response` ∈ `going` | `maybe` | `not_going`.

## B10. Hapus RSVP
- **URL**: `DELETE /schedule/{entry_id}/rsvp`
- **Autentikasi**: Yes
- **Request Body**: `{ "occurrence_date": "2026-07-12" }`
- **Response (200)**: `{ "message": "RSVP removed" }`

## B11. Daftar RSVP per Occurrence (Pengelola)
Daftar nama peserta per respons — untuk pengelola melihat siapa yang ikut.

- **URL**: `GET /schedule/{entry_id}/occurrences/{occurrence_date}/rsvps`
- **Autentikasi**: Yes + `manage_community_schedule`
- **Query**: `response` (opsional filter), `limit`, `offset`
- **Response (200)**:
  ```json
  {
    "data": {
      "summary": { "going": 12, "maybe": 4, "not_going": 2 },
      "respondents": [
        { "user": { "id": "usr_05", "name": "Andi", "avatar": null }, "response": "going", "responded_at": "2026-07-06T10:00:00.000Z" }
      ]
    },
    "pagination": { "limit": 20, "offset": 0, "total": 18 }
  }
  ```
> Ingat: ini **indikator minat**, bukan absensi resmi (lihat Rule 6 di RULES).

---

## Status Code Reference

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request — input tidak valid |
| `401` | Unauthorized — token invalid/expired |
| `403` | Forbidden — bukan anggota / tak punya permission kelola |
| `404` | Not Found — resource/komunitas tidak ada, atau pengumuman tak terlihat member |
| `409` | Conflict — RSVP pada occurrence batal/lewat, dsb |
| `422` | Validation error / window terlalu lebar |
| `500` | Internal Server Error |

**Validation errors baru (2026-07-17, B4/B5):**

| Kode Domain | HTTP | Kondisi |
|---|---|---|
| `DOMAIN_COMMUNITY_SCHEDULE_ENTRY_TIMEZONE_REQUIRED` | 422 | `timezone` tidak dikirim saat create (B4) |
| `DOMAIN_COMMUNITY_SCHEDULE_ENTRY_TIMEZONE_INVALID` | 422 | `timezone` bukan IANA identifier valid (gagal validasi `time.LoadLocation`), create maupun edit |

---

## Notes & Best Practices

1. **Permission check per komunitas.** `403` bila requester bukan anggota (untuk baca private) atau tak punya permission kelola pada `community_id` tersebut. Jangan andalkan client-side gating.
2. **Kalender selalu ber-window.** `GET /schedule` wajib `from`/`to`. Client memuat per bulan/minggu, backend expand recurrence hanya dalam window (maks ~92 hari).
3. **Occurrence key = `occurrence_date` (YYYY-MM-DD).** Semua aksi RSVP/cancel menargetkan `(entry_id, occurrence_date)`. Time-of-day diambil dari `entry.start_at`, direkonstruksi ulang di zona `entry.timezone` (lihat §Model Data Utama — bukan zona server, bukan zona device requester).
4. **Upload media dulu.** Untuk pengumuman bermedia, `POST /api/v1/mobile/media/upload` (context `community_announcement`) lalu kirim URL di body. Maks 5 gambar.
5. **Timestamp ISO 8601 UTC**; `occurrence_date` memakai format tanggal `YYYY-MM-DD` (tanpa waktu).
6. **Caching:** feed pengumuman & kalender aman di-cache pendek (1–5 menit); invalidasi lokal setelah aksi RSVP/create.
7. **Notif reminder** dikirim otomatis sebelum `start_at` occurrence ke anggota yang RSVP `going`/`maybe` (Phase 1). Client tidak perlu menjadwalkan sendiri.
