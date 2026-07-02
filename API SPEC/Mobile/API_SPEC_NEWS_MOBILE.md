# API Spec — News Module (Mobile Client)

Dokumentasi API News untuk aplikasi mobile — membaca artikel, memilih bahasa, like, comment, bookmark, dan share.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/news`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Optional untuk baca, Required untuk like/comment/bookmark)
  - `Accept-Language: <lang_code>` (default: `id`) — menentukan bahasa konten yang dikembalikan
- **Authentication**:
  - Baca artikel & comment: Optional (guest boleh)
  - Like / comment / bookmark: Required (member login)
- **Error Format**: Same as other mobile APIs (standard message or validation error)

---

## Konsep Bahasa (Accept-Language)

Konten artikel dikembalikan sesuai header `Accept-Language`:

```
Request dengan Accept-Language: en
        ↓
Backend cek translation bahasa 'en'
   ├── Ada (translate_status=done) → return versi EN, is_translated=true
   └── Tidak ada → return versi original artikel, is_translated=false
                   + available_languages = bahasa yang tersedia
```

Field `is_translated` dan `available_languages` membantu frontend menampilkan indikator bahasa atau opsi pilih bahasa lain.

---

## Model Data Utama

### 1. Article Object (List)

```json
{
  "id": "art_001",
  "title": "Timnas Indonesia Lolos ke Final",
  "summary": "Tim nasional sepakbola Indonesia berhasil...",
  "thumbnail_url": "https://cdn.kai.app/news/thumb_001.jpg",
  "category": {
    "id": "cat_sport",
    "name": "Olahraga",
    "slug": "olahraga"
  },
  "scope": {
    "id": "scp_id",
    "name": "Berita Indonesia",
    "slug": "indonesia"
  },
  "author_label": "Korean Association Indonesia",
  "language": "en",
  "is_translated": true,
  "available_languages": ["id", "en", "ko"],
  "like_count": 152,
  "comment_count": 23,
  "view_count": 4820,
  "is_liked": false,
  "is_bookmarked": false,
  "is_featured": false,
  "published_at": "2026-06-03T10:00:00.000Z"
}
```

### 2. Article Object (Detail)

```json
{
  "id": "art_001",
  "title": "Timnas Indonesia Lolos ke Final",
  "content": "<p>Tim nasional sepakbola Indonesia...</p>",
  "summary": "Tim nasional sepakbola Indonesia berhasil...",
  "thumbnail_url": "https://cdn.kai.app/news/thumb_001.jpg",
  "author": "Andi Pratama",
  "tags": ["sepakbola", "timnas", "piala asia"],
  "category": {
    "id": "cat_sport",
    "name": "Olahraga",
    "slug": "olahraga"
  },
  "scope": {
    "id": "scp_id",
    "name": "Berita Indonesia",
    "slug": "indonesia"
  },
  "author_label": "Korean Association Indonesia",
  "published_by_label": "Korean Association Indonesia",
  "original_url": "https://detik.com/sport/12345",
  "language": "en",
  "is_translated": true,
  "available_languages": ["id", "en", "ko"],
  "like_count": 152,
  "comment_count": 23,
  "view_count": 4820,
  "is_liked": false,
  "is_bookmarked": false,
  "is_featured": false,
  "published_at": "2026-06-03T10:00:00.000Z"
}
```

> `original_url` hanya muncul untuk artikel hasil scraping (field bernilai `null`/absen untuk artikel manual). Dipakai untuk menampilkan link "Baca sumber asli" — publisher yang ditampilkan ke pembaca tetap `author_label`/`published_by_label` (label KAI), bukan nama portal sumber.

### 3. Comment Object

```json
{
  "id": "cmt_001",
  "content": "Berita bagus, terima kasih!",
  "user": {
    "id": "usr_123",
    "name": "Budi Santoso",
    "avatar": "https://cdn.kai.app/avatars/budi.jpg"
  },
  "parent_id": null,
  "is_deleted": false,
  "reply_count": 2,
  "created_at": "2026-06-03T11:00:00.000Z"
}
```

---

## Endpoints

### 1. GET /articles

Ambil daftar artikel published. Konten mengikuti `Accept-Language`.

**Authentication**: Optional (jika login → `is_liked`/`is_bookmarked` akurat)

**Method**: GET

**URL**: `/api/v1/mobile/news/articles`

**Query Parameters**:
- `category` (optional): Filter by category slug (mis. `olahraga`)
- `scope` (optional): Filter by scope slug (`indonesia`, `korea`, `korea_indonesia`)
- `author_label` (optional): Filter by label asal (mis. `KAI Jakarta`)
- `featured` (optional): `true` untuk hanya artikel featured
- `q` (optional): Search keyword (judul + konten)
- `limit` (optional, default: 20, max: 100)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [
    { "...Article Object (List)" }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 156
  }
}
```

---

### 2. GET /articles/{article_id}

Ambil detail lengkap satu artikel. Otomatis increment view count.

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/news/articles/{article_id}`

**Behavior**:
- Increment `view_count` (total hit) setiap request
- Jika user login & belum pernah lihat → increment `unique_view_count`
- Konten mengikuti `Accept-Language`; jika translation tidak ada → fallback ke original + `is_translated: false`

**Response (200 OK)**: Article Object (Detail)

**Response (404 Not Found)**:
```json
{ "message": "Article not found" }
```

---

### 3. GET /articles/{article_id}/translate

Minta versi artikel dalam bahasa tertentu. Jika belum ada, sistem akan men-generate (sesuai `on_demand` setting). Mengembalikan versi original jika translation masih diproses.

**Authentication**: Required

**Method**: GET

**URL**: `/api/v1/mobile/news/articles/{article_id}/translate?language=ko`

**Query Parameters**:
- `language` (required): Kode bahasa target (mis. `ko`, `en`)

**Response (200 OK — sudah tersedia)**:
```json
{
  "data": {
    "article_id": "art_001",
    "language": "ko",
    "title": "...",
    "content": "...",
    "is_translated": true,
    "translate_status": "done"
  }
}
```

**Response (202 Accepted — sedang diproses)**:
```json
{
  "data": {
    "article_id": "art_001",
    "language": "ko",
    "translate_status": "processing",
    "fallback": {
      "language": "id",
      "title": "...",
      "content": "..."
    }
  },
  "message": "Translation in progress. Showing original for now."
}
```

> Jika `translation_enabled = false` di system settings → endpoint mengembalikan `403` dengan pesan fitur translation dinonaktifkan.

---

### 4. GET /categories

Ambil daftar kategori news aktif (untuk filter/tab).

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/news/categories`

**Response (200 OK)**:
```json
{
  "data": [
    { "id": "cat_sport", "name": "Olahraga", "slug": "olahraga", "sort_order": 1 },
    { "id": "cat_eco",   "name": "Ekonomi",  "slug": "ekonomi",  "sort_order": 2 }
  ]
}
```

---

### 5b. GET /scopes

Ambil daftar scope news aktif (untuk filter/tab asal berita: Indonesia / Korea / Korea di Indonesia).

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/news/scopes`

**Response (200 OK)**:
```json
{
  "data": [
    { "id": "scp_id", "name": "Berita Indonesia", "slug": "indonesia", "sort_order": 1 },
    { "id": "scp_kr", "name": "Berita Korea", "slug": "korea", "sort_order": 2 },
    { "id": "scp_kri", "name": "Berita Korea di Indonesia", "slug": "korea_indonesia", "sort_order": 3 }
  ]
}
```

---

### 5. POST /articles/{article_id}/like

Like artikel. Idempotent — like ganda diabaikan.

**Authentication**: Required

**Method**: POST

**URL**: `/api/v1/mobile/news/articles/{article_id}/like`

**Response (200 OK)**:
```json
{
  "data": { "article_id": "art_001", "is_liked": true, "like_count": 153 }
}
```

---

### 6. DELETE /articles/{article_id}/like

Unlike artikel.

**Authentication**: Required

**Method**: DELETE

**URL**: `/api/v1/mobile/news/articles/{article_id}/like`

**Response (200 OK)**:
```json
{
  "data": { "article_id": "art_001", "is_liked": false, "like_count": 152 }
}
```

---

### 7. POST /articles/{article_id}/bookmark

Simpan/bookmark artikel.

**Authentication**: Required

**Method**: POST

**URL**: `/api/v1/mobile/news/articles/{article_id}/bookmark`

**Response (200 OK)**:
```json
{
  "data": { "article_id": "art_001", "is_bookmarked": true }
}
```

---

### 8. DELETE /articles/{article_id}/bookmark

Hapus bookmark.

**Authentication**: Required

**Method**: DELETE

**URL**: `/api/v1/mobile/news/articles/{article_id}/bookmark`

**Response (200 OK)**:
```json
{
  "data": { "article_id": "art_001", "is_bookmarked": false }
}
```

---

### 9. GET /bookmarks

Ambil daftar artikel yang di-bookmark user.

**Authentication**: Required

**Method**: GET

**URL**: `/api/v1/mobile/news/bookmarks`

**Query Parameters**:
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [ { "...Article Object (List)" } ],
  "pagination": { "limit": 20, "offset": 0, "total": 12 }
}
```

---

### 10. GET /articles/{article_id}/comments

Ambil komentar level 1 (parent) untuk satu artikel. Reply diambil terpisah.

**Authentication**: Optional (guest boleh baca)

**Method**: GET

**URL**: `/api/v1/mobile/news/articles/{article_id}/comments`

**Query Parameters**:
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "cmt_001",
      "content": "Berita bagus!",
      "user": { "id": "usr_123", "name": "Budi", "avatar": "..." },
      "parent_id": null,
      "is_deleted": false,
      "reply_count": 2,
      "created_at": "2026-06-03T11:00:00.000Z"
    },
    {
      "id": "cmt_002",
      "content": null,
      "user": { "id": "usr_456", "name": "Sari", "avatar": "..." },
      "parent_id": null,
      "is_deleted": true,
      "reply_count": 0,
      "created_at": "2026-06-03T11:30:00.000Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 18 }
}
```

> Comment yang sudah dihapus: `is_deleted=true`, `content=null`. Frontend tampilkan "Komentar ini telah dihapus".

---

### 11. GET /comments/{comment_id}/replies

Ambil reply (level 2) untuk satu comment.

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/news/comments/{comment_id}/replies`

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "cmt_003",
      "content": "Setuju!",
      "user": { "id": "usr_789", "name": "Andi", "avatar": "..." },
      "parent_id": "cmt_001",
      "is_deleted": false,
      "created_at": "2026-06-03T12:00:00.000Z"
    }
  ]
}
```

---

### 12. POST /articles/{article_id}/comments

Tulis komentar baru (level 1) atau reply (level 2).

**Authentication**: Required

**Method**: POST

**URL**: `/api/v1/mobile/news/articles/{article_id}/comments`

**Request Body**:
```json
{
  "content": "Komentar saya...",
  "parent_id": null
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `content` | `string` | Yes | Isi komentar (max 1000 char) |
| `parent_id` | `string` | No | ID comment level 1 jika ini reply. Jika reply ke reply, backend tetap set ke parent level 1 |

**Response (201 Created)**:
```json
{
  "data": {
    "id": "cmt_004",
    "content": "Komentar saya...",
    "parent_id": null,
    "user": { "id": "usr_123", "name": "Budi", "avatar": "..." },
    "created_at": "2026-06-04T09:00:00.000Z"
  },
  "message": "Comment posted"
}
```

> Tidak ada edit comment. Untuk koreksi, user hapus lalu tulis ulang.

---

### 13. DELETE /comments/{comment_id}

Hapus komentar (soft delete). Hanya pemilik comment.

**Authentication**: Required

**Method**: DELETE

**URL**: `/api/v1/mobile/news/comments/{comment_id}`

**Response (200 OK)**:
```json
{
  "data": { "comment_id": "cmt_004", "is_deleted": true },
  "message": "Comment deleted"
}
```

**Response (403 Forbidden)**:
```json
{ "message": "You can only delete your own comments" }
```

---

### 14. POST /comments/{comment_id}/report

Laporkan komentar. Masuk ke sistem reporting (`reportable_type = news_comment`).

**Authentication**: Required

**Method**: POST

**URL**: `/api/v1/mobile/news/comments/{comment_id}/report`

**Request Body**:
```json
{
  "reason": "harassment",
  "detail": "Komentar mengandung pelecehan"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `reason` | `string` | Yes | Enum: `spam`, `harassment`, `hate_speech`, `sexual_content`, `violence`, `misinformation`, `scam`, `impersonation`, `other` |
| `detail` | `string` | No | Wajib jika `reason = other` |

**Response (201 Created)**:
```json
{ "message": "Report submitted" }
```

---

## Error Responses

### 400 Bad Request
```json
{ "message": "Invalid query parameter" }
```

### 401 Unauthorized
```json
{ "message": "Authentication required" }
```

### 403 Forbidden
```json
{ "message": "You can only delete your own comments" }
```

### 404 Not Found
```json
{ "message": "Article not found" }
```

### 422 Unprocessable Entity
```json
{
  "message": "Validation failed",
  "errors": { "content": ["Content exceeds 1000 character limit"] }
}
```

---

## Status Codes

- `200 OK` — Success
- `201 Created` — Resource created (comment, report)
- `202 Accepted` — Translation in progress
- `400 Bad Request` — Bad input
- `401 Unauthorized` — Auth required
- `403 Forbidden` — Permission denied
- `404 Not Found` — Resource not found
- `422 Unprocessable Entity` — Validation error
- `500 Internal Server Error` — Server error

---

## Caching Strategy

- `GET /articles` — Cache 2 menit (per bahasa)
- `GET /articles/{id}` — Cache 5 menit (per bahasa), invalidate saat edit
- `GET /categories` — Cache 1 jam
- `GET /scopes` — Cache 1 jam
- `GET /comments` — No cache (real-time)

---

*API spec News untuk mobile client.*
