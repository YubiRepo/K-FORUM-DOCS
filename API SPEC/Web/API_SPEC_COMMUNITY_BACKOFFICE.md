# API Specification — Community Module (Backoffice)

> **Dibuat:** 2026-06-01
> **Stack:** Web Backoffice (Superadmin)
> **Base URL Prefix:** `/api/v1/web/communities`

Spesifikasi endpoint backoffice untuk modul Community: pengawasan & moderasi global oleh Superadmin — melihat seluruh komunitas, suspend/archive, moderasi konten, mengelola anggota lintas komunitas, dan menangani komunitas `orphaned`.

Berdasarkan `COMMUNITY_RULES.md` dan `COMMUNITY_DB_SCHEMA.md`. Endpoint mobile (member-facing) ada di `API_SPEC_COMMUNITY_MOBILE.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Data Models](#data-models)
3. [Community Endpoints](#community-endpoints)
   - [B1. List Communities](#b1-list-communities)
   - [B2. Get Community Detail](#b2-get-community-detail)
   - [B3. Suspend Community](#b3-suspend-community)
   - [B4. Unsuspend Community](#b4-unsuspend-community)
   - [B5. Archive Community](#b5-archive-community)
   - [B6. Delete Community](#b6-delete-community)
4. [Orphaned Handling](#orphaned-handling)
   - [B7. List Orphaned Communities](#b7-list-orphaned-communities)
   - [B8. Assign New Owner](#b8-assign-new-owner)
5. [Member Endpoints](#member-endpoints)
   - [B9. List Members](#b9-list-members)
   - [B10. Force Remove / Ban Member](#b10-force-remove--ban-member)
6. [Content Moderation Endpoints](#content-moderation-endpoints)
   - [B11. List Posts](#b11-list-posts)
   - [B12. Get Post Detail](#b12-get-post-detail)
   - [B13. Remove Post](#b13-remove-post)
   - [B14. Restore Post](#b14-restore-post)
   - [B15. List Comments of Post](#b15-list-comments-of-post)
   - [B16. Remove Comment](#b16-remove-comment)
7. [Statistics](#statistics)
   - [B17. Community Stats](#b17-community-stats)
8. [Status Code Reference](#status-code-reference)
9. [Notification Triggers](#notification-triggers)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/web/communities`
- **Headers Global:**
  ```
  Content-Type: application/json
  Accept: application/json
  Authorization: Bearer <access_token>
  ```

- **Autentikasi:** Semua endpoint di bawah ini **hanya bisa diakses oleh Superadmin**.
  - Request tanpa token atau role bukan superadmin → `401` / `403`.
  - Superadmin punya akses lintas komunitas (termasuk komunitas `private`), tanpa perlu menjadi anggota.

> **Catatan:** Pengelolaan rutin sebuah komunitas (approve join, promote, moderasi harian) dilakukan oleh **leader/moderator** lewat endpoint mobile. Backoffice difokuskan pada **intervensi platform-level**: pelanggaran kebijakan, sengketa, dan komunitas tanpa pemilik.

---

## Data Models

### CommunityAdminListObject
```json
{
  "id": "uuid",
  "name": "Komunitas WNI Seoul",
  "slug": "komunitas-wni-seoul",
  "owner": { "id": "uuid", "name": "Budi Santoso" },
  "region": { "id": "uuid", "name": "Seoul" },
  "visibility": "public",
  "status": "active",
  "member_count": 1240,
  "post_count": 312,
  "created_at": "2026-02-01T08:00:00.000Z"
}
```
> `region` = `null` jika komunitas global.

### CommunityAdminDetailObject
```json
{
  "id": "uuid",
  "name": "Komunitas WNI Seoul",
  "slug": "komunitas-wni-seoul",
  "description": "Wadah berbagi info WNI di Seoul.",
  "avatar_url": "https://cdn/.../avatar.jpg",
  "owner": { "id": "uuid", "name": "Budi Santoso", "email": "budi@example.com" },
  "region": { "id": "uuid", "name": "Seoul" },
  "visibility": "public",
  "status": "active",
  "member_count": 1240,
  "post_count": 312,
  "moderation": {
    "suspended_by": null,
    "suspended_at": null,
    "reason": null
  },
  "created_at": "2026-02-01T08:00:00.000Z",
  "updated_at": "2026-05-30T12:00:00.000Z"
}
```

### AdminMemberObject
```json
{
  "user": { "id": "uuid", "name": "Siti Aminah", "email": "siti@example.com" },
  "role": "moderator",
  "status": "active",
  "joined_at": "2026-02-10T09:00:00.000Z"
}
```

### AdminPostObject
```json
{
  "id": "uuid",
  "community": { "id": "uuid", "name": "Komunitas WNI Seoul" },
  "author": { "id": "uuid", "name": "Budi Santoso" },
  "content": "Ada yang tahu tempat makan halal di Gangnam?",
  "media_count": 1,
  "like_count": 24,
  "comment_count": 8,
  "status": "published",
  "removed_by": null,
  "removed_at": null,
  "created_at": "2026-05-21T12:00:00.000Z"
}
```

### AdminCommentObject
```json
{
  "id": "uuid",
  "post_id": "uuid",
  "parent_comment_id": null,
  "author": { "id": "uuid", "name": "Siti" },
  "content": "Coba ke restoran X di Gangnam-gu.",
  "status": "published",
  "created_at": "2026-05-21T12:30:00.000Z"
}
```

### PaginationObject
```json
{ "limit": 20, "offset": 0, "total": 85, "has_next": true, "has_prev": false }
```

### Error Responses
```json
// 400 / 403 / 404
{ "message": "Deskripsi error" }

// 422 Validation Error
{
  "message": "Data input tidak valid",
  "errors": { "reason": ["Alasan wajib diisi"] }
}
```

---

## Community Endpoints

### B1. List Communities

Daftar seluruh komunitas (semua status & visibility) dengan filter.

- **URL:** `GET /api/v1/web/communities`
- **Auth:** Superadmin
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `q` | string | No | — | Cari nama / slug |
| `status` | string | No | — | `active` \| `suspended` \| `archived` \| `orphaned` |
| `visibility` | string | No | — | `public` \| `private` |
| `region_id` | string (UUID) | No | — | Filter region |
| `owner_id` | string (UUID) | No | — | Filter pemilik |
| `sort` | string | No | `newest` | `newest` \| `popular` \| `most_reported` |
| `limit` | int | No | 20 | Max: 100 |
| `offset` | int | No | 0 | — |

- **Response 200:** array `CommunityAdminListObject` + `pagination`

---

### B2. Get Community Detail

- **URL:** `GET /api/v1/web/communities/{community_id}`
- **Auth:** Superadmin
- **Response 200:** `{ "data": { "...": "CommunityAdminDetailObject" } }`
- **Response 404:** `{ "message": "Komunitas tidak ditemukan" }`

---

### B3. Suspend Community

Bekukan komunitas sementara (pelanggaran kebijakan). Komunitas tidak hilang, tetapi tidak bisa diakses member.

- **URL:** `POST /api/v1/web/communities/{community_id}/suspend`
- **Auth:** Superadmin
- **Request Body:**
```json
{ "reason": "Pelanggaran pedoman komunitas — konten SARA." }
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `reason` | string | **Yes** | Alasan suspend (audit & notifikasi ke leader) |

- **Side effects:** `status = suspended`, simpan `suspended_by`/`suspended_at`/`reason`; notify leader.
- **Response 200:** `{ "message": "Komunitas dibekukan" }`
- **Response 422:** `{ "message": "Data input tidak valid", "errors": { "reason": ["Alasan wajib diisi"] } }`

---

### B4. Unsuspend Community

- **URL:** `POST /api/v1/web/communities/{community_id}/unsuspend`
- **Auth:** Superadmin
- **Side effects:** `status = active`, kosongkan field moderasi; notify leader.
- **Response 200:** `{ "message": "Komunitas diaktifkan kembali" }`
- **Response 409:** `{ "message": "Komunitas tidak dalam status suspended" }`

---

### B5. Archive Community

Arsipkan komunitas (read-only permanen, tidak dihapus). Berbeda dari suspend yang sifatnya sementara.

- **URL:** `POST /api/v1/web/communities/{community_id}/archive`
- **Auth:** Superadmin
- **Request Body:** `{ "reason": "Tidak aktif > 12 bulan." }`
- **Side effects:** `status = archived`.
- **Response 200:** `{ "message": "Komunitas diarsipkan" }`

---

### B6. Delete Community

Hapus permanen beserta seluruh data terkait. **Irreversible.**

- **URL:** `DELETE /api/v1/web/communities/{community_id}`
- **Auth:** Superadmin
- **Side effects (transaksional):** cascade hapus members, join_requests, posts, likes, comments, saves; cleanup `user_roles` + `community_role_permissions` scope ini.
- **Response 200:** `{ "message": "Komunitas dihapus permanen" }`

---

## Orphaned Handling

Komunitas masuk status `orphaned` ketika owner/leader menghapus akun (lihat `COMMUNITY_RULES.md` Rule 1 & 2). Superadmin wajib menugaskan owner baru atau mengarsipkannya.

### B7. List Orphaned Communities

- **URL:** `GET /api/v1/web/communities/orphaned`
- **Auth:** Superadmin
- **Query Params:** `limit` (default 20, max 100), `offset`
- **Response 200:** array `CommunityAdminListObject` (status `orphaned`) + `pagination`

---

### B8. Assign New Owner

Tugaskan member yang ada sebagai leader baru, kembalikan komunitas ke `active`.

- **URL:** `POST /api/v1/web/communities/{community_id}/assign-owner`
- **Auth:** Superadmin
- **Request Body:** `{ "new_owner_id": "uuid" }`
- **Validasi:** `new_owner` harus anggota aktif komunitas tersebut.
- **Side effects:** assign role `leader`, update `communities.owner_id`, `status = active`; notify owner baru.
- **Response 200:** `{ "message": "Owner baru ditetapkan, komunitas aktif kembali" }`
- **Response 422:** `{ "message": "Calon owner bukan anggota aktif komunitas" }`

> Jika tidak ada kandidat anggota yang layak, gunakan **B5 (Archive)** sebagai gantinya.

---

## Member Endpoints

### B9. List Members

- **URL:** `GET /api/v1/web/communities/{community_id}/members`
- **Auth:** Superadmin
- **Query Params:**

| Param | Type | Default | Keterangan |
|-------|------|---------|-----------|
| `role` | string | — | `leader` \| `moderator` \| `member` |
| `status` | string | — | `active` \| `pending` \| `banned` |
| `q` | string | — | Cari nama / email |
| `limit` | int | 20 | Max 100 |
| `offset` | int | 0 | — |

- **Response 200:** array `AdminMemberObject` + `pagination`

---

### B10. Force Remove / Ban Member

Intervensi platform-level terhadap anggota (mis. user bermasalah lintas komunitas).

- **URL:** `PATCH /api/v1/web/communities/{community_id}/members/{user_id}`
- **Auth:** Superadmin
- **Request Body:**
```json
{ "action": "ban", "reason": "Spam berulang." }
```

| `action` | Efek |
|----------|------|
| `kick` | Keluarkan dari komunitas (hapus membership + role) |
| `ban` | `status=banned`, hapus role; tidak bisa re-join |
| `unban` | Kembalikan ke `active` |

- **Catatan:** berbeda dari endpoint mobile (#11), superadmin **boleh** menindak leader/moderator. Jika yang di-ban/kick adalah leader → komunitas menjadi `orphaned` (perlu B8).
- **Response 200:** `{ "message": "Aksi berhasil diterapkan" }`

---

## Content Moderation Endpoints

### B11. List Posts

Lihat post lintas/dalam komunitas untuk moderasi.

- **URL:** `GET /api/v1/web/communities/posts`
- **Auth:** Superadmin
- **Query Params:**

| Param | Type | Default | Keterangan |
|-------|------|---------|-----------|
| `community_id` | string (UUID) | — | Filter per komunitas |
| `author_id` | string (UUID) | — | Filter per penulis |
| `status` | string | — | `published` \| `removed` |
| `q` | string | — | Cari isi konten |
| `limit` | int | 20 | Max 100 |
| `offset` | int | 0 | — |

- **Response 200:** array `AdminPostObject` + `pagination`

---

### B12. Get Post Detail

- **URL:** `GET /api/v1/web/communities/posts/{post_id}`
- **Auth:** Superadmin
- **Response 200:** `{ "data": { "...": "AdminPostObject (dengan media lengkap)" } }`

---

### B13. Remove Post

- **URL:** `POST /api/v1/web/communities/posts/{post_id}/remove`
- **Auth:** Superadmin
- **Request Body:** `{ "reason": "Konten melanggar pedoman." }`
- **Side effects:** `status=removed`, isi `removed_by`/`removed_at`; notify author (opsional).
- **Response 200:** `{ "message": "Postingan dihapus" }`

---

### B14. Restore Post

Kembalikan post yang sebelumnya di-remove.

- **URL:** `POST /api/v1/web/communities/posts/{post_id}/restore`
- **Auth:** Superadmin
- **Side effects:** `status=published`, kosongkan `removed_by`/`removed_at`.
- **Response 200:** `{ "message": "Postingan dipulihkan" }`
- **Response 409:** `{ "message": "Postingan tidak dalam status removed" }`

---

### B15. List Comments of Post

- **URL:** `GET /api/v1/web/communities/posts/{post_id}/comments`
- **Auth:** Superadmin
- **Query Params:** `status` (`published` \| `removed`), `limit` (default 20, max 100), `offset`
- **Response 200:** array `AdminCommentObject` (top-level + reply, ditandai via `parent_comment_id`) + `pagination`

---

### B16. Remove Comment

- **URL:** `POST /api/v1/web/communities/comments/{comment_id}/remove`
- **Auth:** Superadmin
- **Request Body:** `{ "reason": "Ujaran kebencian." }`
- **Side effects:** `status=removed`; decrement counter terkait.
- **Response 200:** `{ "message": "Komentar dihapus" }`

---

## Statistics

### B17. Community Stats

Ringkasan metrik sebuah komunitas (untuk panel admin).

- **URL:** `GET /api/v1/web/communities/{community_id}/stats`
- **Auth:** Superadmin
- **Query Params:** `period` (`7d` \| `30d` \| `90d`, default `30d`)
- **Response 200:**
```json
{
  "data": {
    "member_count": 1240,
    "new_members": 64,
    "active_members": 380,
    "post_count": 312,
    "new_posts": 47,
    "total_likes": 1820,
    "total_comments": 540,
    "pending_join_requests": 5,
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
| 403 | Bukan superadmin |
| 404 | Resource tidak ditemukan |
| 409 | Konflik status (mis. unsuspend komunitas yang tidak suspended) |
| 422 | Validation error (mis. `reason` kosong) |

---

## Notification Triggers

| Aksi Backoffice | Notifikasi |
|---|---|
| B3 Suspend | Ke leader: komunitas dibekukan + alasan |
| B4 Unsuspend | Ke leader: komunitas aktif kembali |
| B5 Archive | Ke leader: komunitas diarsipkan |
| B8 Assign Owner | Ke owner baru: ditetapkan sebagai leader |
| B10 Ban/Kick | Ke user terdampak (opsional, sesuai kebijakan) |
| B13 Remove Post | Ke author (opsional) |
| B16 Remove Comment | Ke author (opsional) |

> Semua event dikirim ke modul Notification sesuai `notification-preferences-technical.md`.

---

*Spesifikasi ini selaras dengan `COMMUNITY_RULES.md`, `COMMUNITY_DB_SCHEMA.md`, dan `API_SPEC_COMMUNITY_MOBILE.md`.*
