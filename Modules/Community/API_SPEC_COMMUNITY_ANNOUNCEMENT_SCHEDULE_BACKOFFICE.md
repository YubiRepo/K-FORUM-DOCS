# API Spec — Community Announcement & Schedule (Backoffice)

Endpoint **oversight & moderasi** untuk superadmin/staff terhadap pengumuman & agenda di **semua komunitas**. Pengelolaan sehari-hari (buat/edit) dilakukan leader/moderator lewat mobile (lihat `API_SPEC_COMMUNITY_ANNOUNCEMENT_SCHEDULE_MOBILE.md`). Backoffice fokus pada **memantau lintas komunitas** dan **moderasi** (arsip/hapus pengumuman, batalin agenda yang melanggar).

Untuk aturan bisnis lihat `COMMUNITY_ANNOUNCEMENT_SCHEDULE_RULES.md`.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/backoffice/community`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>`
- **Authentication**: Wajib. Akses & aksi dikontrol via Role-Permission (superadmin, atau staff dengan permission moderasi konten komunitas global).
- **Scope**: Lintas komunitas (tidak dibatasi `community_id`). Superadmin adalah override — bisa lihat draft/expired/private.

---

## Model Data
Reuse Announcement Object, Schedule Entry Object, dan Occurrence Object dari spec Mobile, dengan tambahan konteks komunitas:
```json
{
  "community": { "id": "comm_futsal", "name": "Futsal Jakarta", "visibility": "public" },
  "...": "field lain sama seperti Mobile"
}
```

---

# BAGIAN A — PENGUMUMAN (Oversight)

## A1. List Pengumuman Lintas Komunitas
- **URL**: `GET /announcements`
- **Autentikasi**: Yes (moderasi)
- **Query Parameters**:
  - `community_id` (opsional filter)
  - `status` (`draft` | `published` | `archived`)
  - `priority` (`normal` | `important`)
  - `q` (cari di title/body), `from`, `to` (rentang `created_at`)
  - `limit` (default 20, max 100), `offset`
- **Response (200)**:
  ```json
  {
    "data": [
      { "id": "ann_001", "community": { "id": "comm_futsal", "name": "Futsal Jakarta" }, "title": "Latihan pindah lapangan", "priority": "important", "status": "published", "author": { "id": "usr_01", "name": "Budi" }, "published_at": "2026-07-03T09:00:00.000Z" }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 134 }
  }
  ```

## A2. Detail Pengumuman
- **URL**: `GET /announcements/{announcement_id}`
- **Autentikasi**: Yes
- **Response (200)**: `{ "data": { Announcement Object + community } }`
> Superadmin bisa lihat semua status (termasuk draft/expired).

## A3. Moderasi Pengumuman (Arsip / Hapus)
- **URL**: `POST /announcements/{announcement_id}/moderate`
- **Autentikasi**: Yes (moderasi)
- **Request Body**:
  ```json
  { "action": "archive", "reason": "Melanggar pedoman komunitas" }
  ```
  > `action`: `archive` | `delete` | `restore`. `reason` opsional (tercatat di audit).
- **Response (200)**: `{ "message": "Announcement archived", "data": { "id": "ann_001", "status": "archived" } }`
> Aksi superadmin menimpa milik leader — leader diberi tahu via Notification (opsional Phase 2).

---

# BAGIAN B — SCHEDULE (Oversight)

## B1. List Agenda Lintas Komunitas
- **URL**: `GET /schedule`
- **Autentikasi**: Yes (moderasi)
- **Query Parameters**:
  - `community_id` (opsional), `status` (`active` | `cancelled`)
  - `has_recurrence` (bool), `q`, `from`, `to` (rentang `start_at`)
  - `limit`, `offset`
- **Response (200)**:
  ```json
  {
    "data": [
      { "id": "sch_001", "community": { "id": "comm_futsal", "name": "Futsal Jakarta" }, "title": "Latihan rutin", "recurrence": "FREQ=WEEKLY;BYDAY=SA", "status": "active", "created_by": { "id": "usr_01", "name": "Budi" }, "start_at": "2026-07-05T01:00:00.000Z", "timezone": "Asia/Jakarta" }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 58 }
  }
  ```
> List backoffice menampilkan **entry** (bukan expand occurrence) — oversight ga butuh kalender per tanggal. Untuk detail RSVP satu occurrence, pakai endpoint mobile B11 (superadmin override).
> `timezone` (ditambahkan 2026-07-17): IANA identifier agenda ini — ditambahkan supaya moderator bisa melihat konteks zona waktu saat me-review `start_at`, konsisten dengan yang sudah ada di response mobile.

## B2. Detail Agenda
- **URL**: `GET /schedule/{entry_id}`
- **Autentikasi**: Yes
- **Response (200)**: `{ "data": { Schedule Entry Object + community + rsvp_total + timezone } }`

## B3. Moderasi Agenda (Batalin / Hapus)
- **URL**: `POST /schedule/{entry_id}/moderate`
- **Autentikasi**: Yes (moderasi)
- **Request Body**:
  ```json
  { "action": "cancel", "reason": "Konten tidak sesuai" }
  ```
  > `action`: `cancel` (status=cancelled) | `delete` | `restore`.
- **Response (200)**: `{ "message": "Schedule cancelled", "data": { "id": "sch_001", "status": "cancelled" } }`

---

## Status Code Reference

| Code | Meaning |
|------|---------|
| `200` | Success |
| `401` | Unauthorized |
| `403` | Forbidden — tak punya permission moderasi global |
| `404` | Not Found |
| `422` | Validation error |
| `500` | Internal Server Error |

---

## Notes

1. **Audit trail.** Setiap aksi moderasi (`archive`/`delete`/`cancel`/`restore`) sebaiknya tercatat dengan `actor_id`, `reason`, timestamp — mengikuti pola audit modul lain.
2. **Backoffice tidak membuat konten.** Superadmin tidak membuat pengumuman/agenda atas nama komunitas; itu ranah leader di mobile. Backoffice hanya memantau & memoderasi.
3. **Override visibility.** Superadmin bisa melihat komunitas private, draft, dan pengumuman expired untuk keperluan moderasi.
4. **Notifikasi ke pengelola** saat konten dimoderasi = opsional Phase 2 (event ke leader komunitas terkait).
