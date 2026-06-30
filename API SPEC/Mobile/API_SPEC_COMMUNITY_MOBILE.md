# API Specification — Community Module (Mobile Client)

Spesifikasi endpoint mobile untuk modul Community: pengelolaan komunitas, keanggotaan, join request (private), feed/postingan, serta interaksi (like, comment 1-level, save, share).

Berdasarkan `COMMUNITY_RULES.md` dan `COMMUNITY_DB_SCHEMA.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Data Models](#data-models)
3. [Endpoints — Community](#endpoints--community)
4. [Endpoints — Membership](#endpoints--membership)
5. [Endpoints — Join Request (Private)](#endpoints--join-request-private)
6. [Endpoints — Posts](#endpoints--posts)
7. [Endpoints — Interaksi](#endpoints--interaksi)
8. [Status Code Reference](#status-code-reference)
9. [Notes & Best Practices](#notes--best-practices)
10. [User Flow](#user-flow)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/mobile/communities`
- **Headers Global:**
  ```
  Content-Type: application/json
  Accept: application/json
  Accept-Language: <lang_code>    (ko | id | en — default: ko)
  X-Locale: <lang_code>
  Authorization: Bearer <access_token>
  ```

- **Permission / Benefit yang dibutuhkan:**

| Endpoint                                   | Auth       | Benefit / Keterangan                         |
| ------------------------------------------ | ---------- | -------------------------------------------- |
| GET communities (browse)                   | ✅ Required | member (Standard / Pro)                      |
| GET communities/:id                        | ✅ Required | member                                       |
| GET communities/mine                       | ✅ Required | member                                       |
| POST communities                           | ✅ Required | `create_community` (benefit Pro)             |
| PATCH communities/:id                      | ✅ Required | leader, permission `manage_community`        |
| DELETE communities/:id                     | ✅ Required | leader                                       |
| POST communities/:id/join                  | ✅ Required | member                                       |
| DELETE communities/:id/membership          | ✅ Required | member (bukan leader)                        |
| GET communities/:id/members                | ✅ Required | member komunitas                             |
| POST communities/:id/transfer-ownership    | ✅ Required | leader                                       |
| PATCH communities/:id/members/:uid         | ✅ Required | `manage_members`                             |
| GET / POST / PATCH join-requests           | ✅ Required | `manage_members` (kecuali POST oleh pemohon) |
| GET communities/:id/posts                  | ✅ Required | member komunitas                             |
| POST communities/:id/posts                 | ✅ Required | anggota aktif, permission `post_content`     |
| DELETE posts/:id                           | ✅ Required | author ATAU permission `delete_content`      |
| POST media/upload                          | ✅ Required | anggota aktif                                |
| POST posts/:id/like, comments, save, share | ✅ Required | anggota aktif komunitas                      |

> Komunitas **private** hanya bisa dilihat metadata-nya (nama, deskripsi, jumlah anggota) oleh non-anggota; konten/feed terkunci sampai menjadi anggota.

---

## Data Models

### CommunityListObject (List — simplified)
```json
{
  "id": "uuid",
  "name": "Komunitas WNI Seoul",
  "slug": "komunitas-wni-seoul",
  "avatar_url": "https://cdn/.../avatar.jpg",
  "category": { "id": "uuid", "name": "Sosial & Budaya" },
  "visibility": "public",
  "region": { "id": "uuid", "name": "Seoul" },
  "member_count": 1240,
  "is_member": true,
  "membership_status": "active"
}
```
> `region` bernilai `null` jika komunitas global. `membership_status`: `active` | `pending` | `banned` | `null` (belum gabung).

### CommunityDetailObject
```json
{
  "id": "uuid",
  "name": "Komunitas WNI Seoul",
  "slug": "komunitas-wni-seoul",
  "description": "Wadah berbagi info WNI di Seoul dan sekitarnya.",
  "avatar_url": "https://cdn/.../avatar.jpg",
  "category": { "id": "uuid", "name": "Sosial & Budaya" },
  "visibility": "public",
  "region": { "id": "uuid", "name": "Seoul" },
  "owner": { "id": "uuid", "name": "Budi Santoso", "avatar": "https://cdn/.../u.jpg" },
  "member_count": 1240,
  "status": "active",
  "is_member": true,
  "membership_status": "active",
  "my_role": "member",
  "my_permissions": ["post_content"],
  "created_at": "2026-02-01T08:00:00.000Z"
}
```
> `my_role`: `leader` | `moderator` | `member` | `null`. `my_permissions`: daftar permission user di komunitas ini (dari Role-Permission module).

### MemberObject
```json
{
  "user": { "id": "uuid", "name": "Siti Aminah", "avatar": "https://cdn/.../s.jpg" },
  "role": "moderator",
  "status": "active",
  "joined_at": "2026-02-10T09:00:00.000Z"
}
```

### JoinRequestObject
```json
{
  "id": "uuid",
  "user": { "id": "uuid", "name": "Andi", "avatar": "https://cdn/.../a.jpg" },
  "message": "Saya WNI yang baru pindah ke Seoul.",
  "status": "pending",
  "created_at": "2026-05-20T10:00:00.000Z"
}
```

### PostObject
```json
{
  "id": "uuid",
  "community_id": "uuid",
  "author": { "id": "uuid", "name": "Budi Santoso", "avatar": "https://cdn/.../u.jpg" },
  "content": "Ada yang tahu tempat makan halal di Gangnam?",
  "media": [
    { "url": "https://cdn/.../1.jpg", "thumb_url": "https://cdn/.../1_t.jpg", "width": 1080, "height": 1350, "order": 0 }
  ],
  "like_count": 24,
  "comment_count": 8,
  "save_count": 3,
  "share_count": 1,
  "is_liked": true,
  "is_saved": false,
  "status": "published",
  "created_at": "2026-05-21T12:00:00.000Z"
}
```

### CommentObject
```json
{
  "id": "uuid",
  "post_id": "uuid",
  "parent_comment_id": null,
  "author": { "id": "uuid", "name": "Siti", "avatar": "https://cdn/.../s.jpg" },
  "content": "Coba ke restoran X di Gangnam-gu, enak.",
  "reply_count": 2,
  "status": "published",
  "created_at": "2026-05-21T12:30:00.000Z"
}
```
> `parent_comment_id` = `null` untuk komentar top-level; berisi UUID induk untuk reply (maks 1 level).

### MediaUploadObject
```json
{
  "url": "https://cdn/.../1.jpg",
  "thumb_url": "https://cdn/.../1_t.jpg",
  "width": 1080,
  "height": 1350
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
  "errors": {
    "name": ["Nama komunitas wajib diisi"],
    "media": ["Maksimal 10 gambar per post"]
  }
}
```

---

## Endpoints — Community

### 0. List Categories

Daftar kategori komunitas yang aktif — dipakai untuk dropdown saat create/edit komunitas dan filter di browse.

- **URL:** `GET /api/v1/mobile/communities/categories`
- **Auth:** Required (member)
- **Response 200:**
```json
{
  "data": [
    { "id": "uuid", "name": "Sosial & Budaya", "slug": "sosial-budaya" },
    { "id": "uuid", "name": "Pendidikan", "slug": "pendidikan" }
  ]
}
```

---

### 1. Browse Communities

Daftar komunitas `active`. Filter region & visibility, plus pencarian.

- **URL:** `GET /api/v1/mobile/communities`
- **Auth:** Required (member)
- **Query Params:**

| Param         | Type          | Required | Default   | Keterangan                           |
| ------------- | ------------- | -------- | --------- | ------------------------------------ |
| `q`           | string        | No       | —         | Cari nama komunitas (min 2 karakter) |
| `category_id` | string (UUID) | No       | —         | Filter kategori                      |
| `region_id`   | string (UUID) | No       | —         | Filter region; kosong = semua        |
| `visibility`  | string        | No       | —         | `public` \| `private`                |
| `sort`        | string        | No       | `popular` | `popular` (member_count) \| `newest` |
| `limit`       | int           | No       | 20        | Max: 50                              |
| `offset`      | int           | No       | 0         | —                                    |

- **Response 200:**
```json
{
  "data": [ { "...": "CommunityListObject" } ],
  "pagination": { "limit": 20, "offset": 0, "total": 42, "has_next": true, "has_prev": false }
}
```

---

### 2. Get Community Detail

- **URL:** `GET /api/v1/mobile/communities/{community_id}`
- **Auth:** Required (member)
- **Response 200:** `{ "data": { "...": "CommunityDetailObject" } }`
- **Response 404:** `{ "message": "Komunitas tidak ditemukan" }`

> Jika komunitas `private` dan user bukan anggota: field metadata tetap dikembalikan, tetapi `my_role` = `null` dan akses feed ditolak di endpoint posts.

---

### 3. Get My Communities

Komunitas yang diikuti user (status `active`).

- **URL:** `GET /api/v1/mobile/communities/mine`
- **Auth:** Required (member)
- **Query Params:** `limit` (default 20, max 50), `offset` (default 0)
- **Response 200:** array `CommunityListObject` + `pagination`

---

### 4. Create Community

- **URL:** `POST /api/v1/mobile/communities`
- **Auth:** Required — benefit `create_community` (Pro)
- **Request Body:**
```json
{
  "name": "Komunitas WNI Seoul",
  "description": "Wadah berbagi info WNI di Seoul.",
  "avatar_url": "s3:/uploads/communities/avatar.jpg",
  "category_id": "uuid",
  "visibility": "public",
  "region_id": "uuid"
}
```

| Field         | Type          | Required              | Keterangan                                     |
| ------------- | ------------- | --------------------- | ---------------------------------------------- |
| `name`        | string        | **Yes**               | 3–150 karakter                                 |
| `description` | string        | No                    | —                                              |
| `avatar_url`  | string        | No                    | File key `s3:` hasil upload media              |
| `category_id` | string (UUID) | **Yes**               | Pilih dari daftar kategori aktif (endpoint #0) |
| `visibility`  | string        | No (default `public`) | `public` \| `private`                          |
| `region_id`   | string (UUID) | No                    | Kosong = komunitas global                      |

- **Side effects:** creator auto jadi `leader`, permission disalin dari template, `community_members` (status active) dibuat.
- **Response 201:** `{ "data": { "...": "CommunityDetailObject" } }`
- **Response 403:** `{ "message": "Fitur membuat komunitas hanya untuk member Pro" }`
- **Response 422:** validation error.

---

### 5. Update Community

- **URL:** `PATCH /api/v1/mobile/communities/{community_id}`
- **Auth:** Required — leader (permission `manage_community`)
- **Request Body (partial):** `name`, `description`, `avatar_url`, `category_id`, `visibility`, `region_id`
- **Response 200:** `{ "data": { "...": "CommunityDetailObject" } }`
- **Response 403:** `{ "message": "Anda tidak punya izin mengubah komunitas ini" }`

---

### 6. Delete Community

- **URL:** `DELETE /api/v1/mobile/communities/{community_id}`
- **Auth:** Required — leader
- **Side effects:** cascade hapus members, join_requests, posts, dan cleanup `user_roles` + `community_role_permissions` (transaksional).
- **Response 200:** `{ "message": "Komunitas berhasil dihapus" }`
- **Response 403:** `{ "message": "Hanya leader yang dapat menghapus komunitas" }`

---

## Endpoints — Membership

### 7. Join Community

- **URL:** `POST /api/v1/mobile/communities/{community_id}/join`
- **Auth:** Required (member)
- **Request Body (opsional, untuk private):**
```json
{ "message": "Saya WNI yang baru pindah ke Seoul." }
```
- **Behaviour:**
  - `public` → langsung jadi anggota (`status=active`), assign role `member`, emit `member_joined`.
  - `private` → buat `community_join_requests` (`status=pending`), notify leader/moderator.
- **Response 200 (public):** `{ "data": { "membership_status": "active" } }`
- **Response 202 (private):** `{ "data": { "membership_status": "pending" } }`
- **Response 409:** `{ "message": "Anda sudah menjadi anggota / sudah mengajukan permintaan" }`
- **Response 403:** `{ "message": "Anda diblokir dari komunitas ini" }` (status banned)

---

### 8. Leave Community

- **URL:** `DELETE /api/v1/mobile/communities/{community_id}/membership`
- **Auth:** Required (member, **bukan** leader)
- **Side effects:** hapus `community_members`, hapus `user_roles` scope ini, `member_count -= 1`, emit `member_left`.
- **Response 200:** `{ "message": "Berhasil keluar dari komunitas" }`
- **Response 403:** `{ "message": "Leader harus transfer ownership sebelum keluar" }`

---

### 9. Get Community Members

- **URL:** `GET /api/v1/mobile/communities/{community_id}/members`
- **Auth:** Required — anggota komunitas
- **Query Params:**

| Param    | Type   | Default | Keterangan                                 |
| -------- | ------ | ------- | ------------------------------------------ |
| `role`   | string | —       | Filter `leader` \| `moderator` \| `member` |
| `q`      | string | —       | Cari nama anggota                          |
| `limit`  | int    | 20      | Max 50                                     |
| `offset` | int    | 0       | —                                          |

- **Response 200:** array `MemberObject` + `pagination`

---

### 10. Transfer Ownership

- **URL:** `POST /api/v1/mobile/communities/{community_id}/transfer-ownership`
- **Auth:** Required — leader
- **Request Body:** `{ "new_owner_id": "uuid" }`
- **Validasi:** `new_owner` harus anggota aktif komunitas.
- **Side effects:** leader lama → `member`; new_owner → `leader`; update `communities.owner_id` (transaksional).
- **Response 200:** `{ "message": "Kepemilikan berhasil dialihkan" }`
- **Response 422:** `{ "message": "Calon owner bukan anggota aktif komunitas" }`

---

### 11. Manage Member (Promote / Kick / Ban / Unban)

- **URL:** `PATCH /api/v1/mobile/communities/{community_id}/members/{user_id}`
- **Auth:** Required — permission `manage_members`
- **Request Body:**
```json
{ "action": "promote" }
```

| `action`  | Efek                                                |
| --------- | --------------------------------------------------- |
| `promote` | Member → moderator (assign role)                    |
| `demote`  | Moderator → member                                  |
| `kick`    | Keluarkan dari komunitas (hapus membership + role)  |
| `ban`     | Set `status=banned`, hapus role; tidak bisa re-join |
| `unban`   | Set `status=active` kembali                         |

- **Aturan:** tidak bisa mengubah role leader lewat endpoint ini (pakai transfer-ownership). Moderator tidak bisa mem-promote/ban moderator lain kecuali permission mengizinkan.
- **Response 200:** `{ "message": "Aksi berhasil diterapkan" }`
- **Response 403:** `{ "message": "Anda tidak punya izin mengelola anggota" }`

---

## Endpoints — Join Request (Private)

### 12. List Join Requests

- **URL:** `GET /api/v1/mobile/communities/{community_id}/join-requests`
- **Auth:** Required — permission `manage_members`
- **Query Params:** `status` (default `pending`), `limit`, `offset`
- **Response 200:** array `JoinRequestObject` + `pagination`

---

### 13. Approve Join Request

- **URL:** `POST /api/v1/mobile/communities/{community_id}/join-requests/{request_id}/approve`
- **Auth:** Required — permission `manage_members`
- **Side effects:** request `status=approved`, buat `community_members` (active) + role `member`, emit `member_joined`.
- **Response 200:** `{ "message": "Permintaan disetujui" }`

---

### 14. Reject Join Request

- **URL:** `POST /api/v1/mobile/communities/{community_id}/join-requests/{request_id}/reject`
- **Auth:** Required — permission `manage_members`
- **Response 200:** `{ "message": "Permintaan ditolak" }`

---

## Endpoints — Posts

### 15. Get Community Feed

- **URL:** `GET /api/v1/mobile/communities/{community_id}/posts`
- **Auth:** Required — anggota komunitas (private terkunci untuk non-anggota)
- **Query Params:** `limit` (default 20, max 50), `offset`
- **Response 200:** array `PostObject` (status `published`, terbaru dulu) + `pagination`
- **Response 403:** `{ "message": "Anda harus menjadi anggota untuk melihat konten" }`

---

### 16. Get Post Detail

- **URL:** `GET /api/v1/mobile/communities/posts/{post_id}`
- **Auth:** Required — anggota komunitas
- **Response 200:** `{ "data": { "...": "PostObject" } }`
- **Response 404:** `{ "message": "Postingan tidak ditemukan" }`

---

### 17. Create Post

- **URL:** `POST /api/v1/mobile/communities/{community_id}/posts`
- **Auth:** Required — anggota aktif, permission `post_content`
- **Request Body:**
```json
{
  "content": "Ada yang tahu tempat makan halal di Gangnam?",
  "media": [
    { "url": "s3:/uploads/community/posts/1.jpg", "thumb_url": "s3:/uploads/community/posts/1_t.jpg", "width": 1080, "height": 1350, "order": 0 }
  ]
}
```

| Field     | Type   | Required | Keterangan                       |
| --------- | ------ | -------- | -------------------------------- |
| `content` | string | **Yes**  | Plain text (tanpa markdown/HTML) |
| `media`   | array  | No       | Maks **10** objek (hasil upload) |

- **Side effects:** emit event `new_posts` ke Notification.
- **Response 201:** `{ "data": { "...": "PostObject" } }`
- **Response 422:** `{ "message": "Data input tidak valid", "errors": { "media": ["Maksimal 10 gambar per post"] } }`

---

### 18. Delete Post

- **URL:** `DELETE /api/v1/mobile/communities/posts/{post_id}`
- **Auth:** Required — author postingan **atau** permission `delete_content` (moderasi)
- **Behaviour:** soft delete (`status=removed`, isi `removed_by`/`removed_at`).
- **Response 200:** `{ "message": "Postingan dihapus" }`
- **Response 403:** `{ "message": "Anda tidak punya izin menghapus postingan ini" }`

---

### 19. Presign Media Upload

Minta presigned URL untuk upload file ke S3. Client upload langsung ke `upload_url` yang dikembalikan, lalu panggil confirm.

- **URL:** `POST /api/v1/mobile/communities/media/presign`
- **Auth:** Required — anggota aktif
- **Request Body:**
```json
{
  "filename": "photo.jpg",
  "type": "post"
}
```

| Field      | Type   | Required | Keterangan                    |
| ---------- | ------ | -------- | ----------------------------- |
| `filename` | string | **Yes**  | Nama file dengan ekstensi     |
| `type`     | string | **Yes**  | `avatar` / `post`             |

- **Response 200:**
```json
{
  "data": {
    "upload_url": "https://s3.ap-northeast-2.amazonaws.com/kforum-uploads/...",
    "file_key": "s3:/uploads/communities/abc123.jpg",
    "expires_in": 900
  }
}
```

### 20. Confirm Media Upload

Konfirmasi upload selesai. Server memproses file (generate thumbnail, dsb).

- **URL:** `POST /api/v1/mobile/communities/media/confirm`
- **Auth:** Required — anggota aktif
- **Request Body:**
```json
{
  "file_key": "s3:/uploads/communities/abc123.jpg"
}
```

- **Response 200:**
```json
{
  "data": {
    "url": "https://cdn.k-forum.id/uploads/communities/abc123.jpg",
    "thumb_url": "https://cdn.k-forum.id/uploads/communities/abc123_t.jpg",
    "width": 1080,
    "height": 1350
  }
}
```

### 21. Delete Media

Hapus media yang sudah di-upload (jika batal dipakai).

- **URL:** `DELETE /api/v1/mobile/communities/media`
- **Auth:** Required — anggota aktif
- **Request Body:**
```json
{
  "file_key": "s3:/uploads/communities/abc123.jpg"
}
```
- **Response 200:** `{ "message": "Media berhasil dihapus" }`
- **Response 404:** `{ "message": "Media tidak ditemukan" }`

> Client memanggil presign, upload langsung ke S3, lalu confirm. File key `s3:` dipakai di request create/update. Jangan kirim base64.

---

## Endpoints — Interaksi

### 22. Like / Unlike Post (Toggle)

- **Like:** `POST /api/v1/mobile/communities/posts/{post_id}/like`
- **Unlike:** `DELETE /api/v1/mobile/communities/posts/{post_id}/like`
- **Auth:** Required — anggota aktif komunitas
- **Side effects:** update `like_count`; like baru emit notifikasi ke author (sesuai preferensi).
- **Response 200:** `{ "data": { "is_liked": true, "like_count": 25 } }`

---

### 23. Get Comments (Top-Level)

- **URL:** `GET /api/v1/mobile/communities/posts/{post_id}/comments`
- **Auth:** Required — anggota komunitas
- **Query Params:** `limit` (default 20, max 50), `offset`
- **Response 200:** array `CommentObject` (top-level, `parent_comment_id=null`) + `pagination`

---

### 24. Get Replies of a Comment

- **URL:** `GET /api/v1/mobile/communities/comments/{comment_id}/replies`
- **Auth:** Required — anggota komunitas
- **Query Params:** `limit` (default 20, max 50), `offset`
- **Response 200:** array `CommentObject` (reply) + `pagination`

---

### 25. Create Comment / Reply

- **URL:** `POST /api/v1/mobile/communities/posts/{post_id}/comments`
- **Auth:** Required — anggota aktif komunitas
- **Request Body:**
```json
{
  "content": "Coba ke restoran X di Gangnam-gu.",
  "parent_comment_id": null
}
```

| Field               | Type          | Required | Keterangan                                               |
| ------------------- | ------------- | -------- | -------------------------------------------------------- |
| `content`           | string        | **Yes**  | Plain text                                               |
| `parent_comment_id` | string (UUID) | No       | Diisi untuk reply; **harus** menunjuk komentar top-level |

- **Validasi 1 level:** jika `parent_comment_id` menunjuk komentar yang sendirinya sudah reply → `400`.
- **Side effects:** update `comment_count` post & `reply_count` induk; emit notifikasi `someone_replied`.
- **Response 201:** `{ "data": { "...": "CommentObject" } }`
- **Response 400:** `{ "message": "Tidak bisa membalas sebuah balasan (maks 1 level)" }`

---

### 26. Delete Comment

- **URL:** `DELETE /api/v1/mobile/communities/comments/{comment_id}`
- **Auth:** Required — author komentar **atau** permission `moderate_posts`
- **Behaviour:** soft delete (`status=removed`); decrement counter terkait.
- **Response 200:** `{ "message": "Komentar dihapus" }`

---

### 27. Save / Unsave Post (Toggle)

- **Save:** `POST /api/v1/mobile/communities/posts/{post_id}/save`
- **Unsave:** `DELETE /api/v1/mobile/communities/posts/{post_id}/save`
- **Auth:** Required — anggota aktif komunitas
- **Response 200:** `{ "data": { "is_saved": true, "save_count": 4 } }`

---

### 28. Get My Saved Posts

- **URL:** `GET /api/v1/mobile/communities/posts/saved`
- **Auth:** Required (member)
- **Query Params:** `limit` (default 20, max 50), `offset`
- **Response 200:** array `PostObject` (yang disimpan, terbaru disimpan dulu) + `pagination`

---

### 29. Share Post

Generate deep link untuk dibagikan keluar (mis. WhatsApp). Menaikkan `share_count`.

- **URL:** `POST /api/v1/mobile/communities/posts/{post_id}/share`
- **Auth:** Required — anggota komunitas
- **Response 200:**
```json
{ "data": { "share_url": "https://app.k-forum.id/c/komunitas-wni-seoul/p/uuid", "share_count": 2 } }
```

> Repost internal (membagikan post ke komunitas lain di dalam app) **belum** didukung — direncanakan Fase 2.

---

## Status Code Reference

| Code | Makna                                               |
| ---- | --------------------------------------------------- |
| 200  | OK                                                  |
| 201  | Created (community, post, comment, media)           |
| 202  | Accepted (join request private — menunggu approval) |
| 400  | Bad request (mis. reply ke reply)                   |
| 401  | Token tidak valid / kedaluwarsa                     |
| 403  | Tidak punya izin / benefit / diblokir               |
| 404  | Resource tidak ditemukan                            |
| 409  | Konflik (sudah anggota / sudah ada request)         |
| 422  | Validation error                                    |

---

## Notes & Best Practices

1. **Plain text only.** Field `content` (post & comment) disimpan apa adanya. Client boleh auto-render URL/`@mention`/`#hashtag` saat tampil, tetapi tidak mengirim markup.
2. **Upload media via presign.** Panggil presign (#19), upload langsung ke S3, confirm (#20), lalu pakai `s3:` key di request create post. Jangan kirim base64.
3. **Toggle, bukan duplikat.** Like & save bersifat idempoten — `POST` saat sudah ada cukup mengembalikan state terkini, tidak menambah ganda.
4. **Counter dari server.** Selalu pakai `like_count`/`comment_count`/`save_count`/`share_count` dari response, jangan hitung di client.
5. **Private community.** Non-anggota hanya melihat metadata; semua endpoint feed/post/komentar menolak dengan `403` sampai menjadi anggota.
6. **Leader tidak bisa leave.** Harus transfer ownership dulu (endpoint #10).
7. **Permission dinamis.** `my_permissions` pada detail komunitas adalah sumber kebenaran untuk menampilkan/menyembunyikan tombol aksi (post, moderate, manage member).

---

## User Flow

### Flow A — Menemukan & Bergabung
```
Browse (1) → Detail (2)
  ├─ public  → Join (7) → langsung anggota → lihat Feed (15)
  └─ private → Join (7, kirim pesan) → status pending
              → leader Approve (13) → anggota → Feed (15)
```

### Flow B — Posting & Interaksi
```
Presign media (19) → Upload ke S3 → Confirm (20) → Create post (17)
Lihat feed (15) → Like (20) / Save (25) / Share (27)
              → Comments (21) → Reply (23, parent_comment_id)
              → Replies (22)
```

### Flow C — Mengelola Komunitas (Leader/Moderator)
```
Join requests (12) → Approve (13) / Reject (14)
Members (9) → Manage member (11: promote/kick/ban)
Transfer ownership (10) → lalu Leave (8)  [jika ingin keluar]
Moderasi: Delete post (18) / Delete comment (24)
```

---

*Spesifikasi ini selaras dengan `COMMUNITY_RULES.md` dan `COMMUNITY_DB_SCHEMA.md`. Endpoint backoffice (superadmin) dijabarkan terpisah di `API_SPEC_COMMUNITY_BACKOFFICE.md`.*