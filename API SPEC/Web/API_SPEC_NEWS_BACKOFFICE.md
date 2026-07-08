# API Spec — News Module (Web Backoffice)

Dokumentasi API News untuk dashboard backoffice — manajemen artikel, sumber scraping, kategori, bahasa, dan setting sistem.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/news`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required
- **Authorization** (berbasis permission, fleksibel via role):
  - `manage_news_source` — daftar/edit source & selector (Usergod)
  - `manage_news_source_config` — schedule, auto_publish, ai_cleanup, auto_translate (Usergod & Superadmin)
  - `manage_news_category` — CRUD kategori (Superadmin)
  - `manage_news_settings` — system settings & languages (Usergod & Superadmin)
  - `create_news` / `edit_news` / `publish_news` — kelola artikel (Editor, Superadmin)
  - `approve_news` — approve/reject artikel Member Pro (Editor, Superadmin)
- **Error Format**: Same as other backoffice APIs

---

## Daftar Endpoint — Media Thumbnail Upload

| Method | Endpoint | Deskripsi |
|--------|----------|-----------|
| POST | `/api/v1/web/media/news/presign` | Generate presigned URL for thumbnail upload |
| POST | `/api/v1/web/media/news/confirm` | Confirm thumbnail upload |
| DELETE | `/api/v1/web/media/news/{file_id}` | Delete uploaded thumbnail |

---

## Bagian A — Article Management

### 1. GET /articles

List artikel dengan filter (semua status, untuk backoffice).

**Permission**: `create_news` atau `edit_news`

**Method**: GET — `/api/v1/web/news/articles`

**Query Parameters**:
- `status` (optional): `draft`, `pending_approval`, `published`, `archived`, `rejected`
- `source_id` (optional): Filter by source
- `category_id` (optional): Filter by kategori
- `news_scope_id` (optional): Filter by scope (asal/fokus geografis)
- `is_manual` (optional): `true`/`false`
- `q` (optional): Search keyword
- `limit` (optional, default: 20), `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "art_001",
      "title": "Timnas Indonesia Lolos ke Final",
      "status": "published",
      "is_manual": false,
      "original_language": "id",
      "author_label": "Korean Association Indonesia",
      "author": "Andi Pratama",
      "thumbnail_url": "https://cdn.kai.app/news/thumb_001.jpg",
      "source": { "id": "src_detik", "name": "Detik.com" },
      "category": { "id": "cat_sport", "name": "Olahraga" },
      "scope": { "id": "scp_id", "name": "Berita Indonesia", "slug": "indonesia" },
      "view_count": 4820,
      "like_count": 152,
      "comment_count": 23,
      "is_featured": false,
      "created_by": { "id": "usr_ed1", "name": "Editor Satu" },
      "published_by_label": "Korean Association Indonesia",
      "published_at": "2026-06-03T10:00:00.000Z",
      "created_at": "2026-06-03T09:30:00.000Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 312 }
}
```

---

### 2. GET /articles/{article_id}

Detail artikel + semua translations.

**Permission**: `create_news` atau `edit_news`

**Method**: GET — `/api/v1/web/news/articles/{article_id}`

**Response (200 OK)**:
```json
{
  "data": {
    "id": "art_001",
    "source_id": "src_detik",
    "category_id": "cat_sport",
    "news_scope_id": "scp_id",
    "original_language": "id",
    "is_manual": false,
    "original_url": "https://detik.com/sport/12345",
    "status": "published",
    "author_label": "Korean Association Indonesia",
    "author_region_id": null,
    "author": "Andi Pratama",
    "thumbnail_url": "https://cdn.kai.app/news/thumb_001.jpg",
    "thumbnail_raw": "s3:/news/thumbnails/thumb_001.jpg",
    "is_featured": false,
    "view_count": 4820,
    "unique_view_count": 3110,
    "like_count": 152,
    "comment_count": 23,
    "created_by": "usr_ed1",
    "published_by_user_id": "usr_ed1",
    "published_by_label": "Korean Association Indonesia",
    "published_at": "2026-06-03T10:00:00.000Z",
    "translations": [
      {
        "language": "id",
        "title": "Timnas Indonesia Lolos ke Final",
        "content": "<p>...</p>",
        "summary": "...",
        "tags": ["sepakbola", "timnas"],
        "is_original": true,
        "translate_status": null,
        "translated_by": null
      },
      {
        "language": "en",
        "title": "Indonesia National Team Reaches Final",
        "content": "<p>...</p>",
        "is_original": false,
        "translate_status": "done",
        "translated_by": "google"
      }
    ]
  }
}
```

---

### 3. POST /articles

Buat artikel manual. Editor pilih bahasa utama + tulis konten.

**Permission**: `create_news`

**Method**: POST — `/api/v1/web/news/articles`

**Request Body**:
```json
{
  "category_id": "cat_sport",
  "news_scope_id": "scp_id",
  "original_language": "id",
  "author_label": "KAI Jakarta",
  "author_region_id": "region_jakarta",
  "author": "Nama Penulis",
  "thumbnail_url": "s3:/news/thumbnails/thumb.jpg",
  "is_featured": false,
  "status": "draft",
  "translation": {
    "title": "Judul Berita",
    "content": "<p>Isi berita...</p>",
    "summary": "Ringkasan singkat",
    "tags": ["tag1", "tag2"]
  }
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `category_id` | `string` | Yes | FK kategori |
| `news_scope_id` | `string` | No | FK scope (asal/fokus geografis). Editor pilih saat tulis artikel manual |
| `original_language` | `string` | Yes | Kode bahasa utama (mis. `id`) |
| `author_label` | `string` | Yes | Label asal (mis. `KAI Jakarta`) |
| `author_region_id` | `string` | No | null jika KAI Pusat |
| `author` | `string` | No | Nama penulis asli artikel. Satu nilai per artikel, sama untuk semua bahasa hasil translate |
| `thumbnail_url` | `string` | No | Thumbnail artikel. Format `s3:` (upload) atau `ext:` (URL eksternal). Satu nilai per artikel, sama untuk semua bahasa hasil translate |
| `is_featured` | `bool` | No | Default false |
| `status` | `string` | No | `draft` (default) atau `published` |
| `translation` | `object` | Yes | Konten dalam bahasa utama (title, content, summary, tags — tidak termasuk author/thumbnail_url) |

**Response (201 Created)**:
```json
{
  "data": { "id": "art_new", "status": "draft" },
  "message": "Article created"
}
```

> Jika `status=published`, sistem set `published_at`, `published_by_user_id`, `published_by_label`, lalu enqueue translation ke bahasa lain (mengikuti system settings).

---

### 4. PUT /articles/{article_id}

Edit artikel. Bisa edit artikel published langsung tanpa re-approval.

**Permission**: `edit_news`

**Method**: PUT — `/api/v1/web/news/articles/{article_id}`

**Request Body**: Sama seperti POST (field yang diubah saja). Bisa juga update translation per bahasa.

**Response (200 OK)**:
```json
{ "data": { "id": "art_001" }, "message": "Article updated" }
```

---

### 5. PUT /articles/{article_id}/translations/{language}

Tambah/edit translation manual untuk bahasa tertentu (override hasil AI).

**Permission**: `edit_news`

**Method**: PUT — `/api/v1/web/news/articles/{article_id}/translations/{language}`

**Request Body**:
```json
{
  "title": "...",
  "content": "<p>...</p>",
  "summary": "...",
  "tags": ["..."]
}
```

**Response (200 OK)**:
```json
{ "message": "Translation saved", "data": { "language": "en", "translated_by": null } }
```

> Saat diisi manual, `translated_by=null`, `translate_status=done`, `is_original=false`.
> `author` dan `thumbnail_url` tidak ada di endpoint ini — keduanya article-level, diubah lewat `PUT /articles/{article_id}` (endpoint 4).

---

### 6. POST /articles/{article_id}/publish

Publish artikel dari draft.

**Permission**: `publish_news`

**Method**: POST — `/api/v1/web/news/articles/{article_id}/publish`

**Response (200 OK)**:
```json
{
  "data": { "id": "art_001", "status": "published", "published_at": "2026-06-04T10:00:00.000Z" },
  "message": "Article published"
}
```

> Set `published_by_user_id` = user yang request, `published_by_label` = `author_label`.

---

### 7. POST /articles/{article_id}/archive

Arsipkan artikel.

**Permission**: `edit_news`

**Method**: POST — `/api/v1/web/news/articles/{article_id}/archive`

**Response (200 OK)**:
```json
{ "data": { "id": "art_001", "status": "archived" }, "message": "Article archived" }
```

---

### 8. POST /articles/{article_id}/approve

Approve artikel dari Member Pro (status `pending_approval` → `published`).

**Permission**: `approve_news`

**Method**: POST — `/api/v1/web/news/articles/{article_id}/approve`

**Response (200 OK)**:
```json
{ "data": { "id": "art_pro_1", "status": "published" }, "message": "Article approved" }
```

---

### 9. POST /articles/{article_id}/reject

Reject artikel Member Pro (`pending_approval` → `rejected`).

**Permission**: `approve_news`

**Method**: POST — `/api/v1/web/news/articles/{article_id}/reject`

**Request Body**:
```json
{ "reason": "Konten tidak relevan dengan platform" }
```

**Response (200 OK)**:
```json
{ "data": { "id": "art_pro_1", "status": "rejected" }, "message": "Article rejected" }
```

---

### 10. POST /articles/{article_id}/translate

Trigger generate translation manual ke bahasa tertentu (atau semua).

**Permission**: `edit_news`

**Method**: POST — `/api/v1/web/news/articles/{article_id}/translate`

**Request Body**:
```json
{ "languages": ["en", "ko"], "provider": "openai" }
```

| Field | Type | Required | Description |
|---|---|---|---|
| `languages` | `string[]` | No | Bahasa target. Kosong = semua bahasa aktif |
| `provider` | `string` | No | `google` (default) atau `openai`. **Belum berfungsi** — diterima untuk kompatibilitas kontrak, tapi provider yang benar-benar dipakai worker masih fixed satu instance global (lihat `NEWS_RULES.md` §Provider Translate) |

**Response (202 Accepted)**:
```json
{ "message": "Translation jobs enqueued", "data": { "enqueued": ["en", "ko"] } }
```

> Bahasa yang belum pernah punya baris translation akan dibuat baru (status `pending`); bahasa yang sudah ada di-reset ke `pending` untuk diproses ulang. `translated_by` hasil akhirnya bisa `google`/`openai`/`noop` tergantung provider yang aktif di worker saat job diproses — lihat `NEWS_DB_SCHEMA.md` §Enum Values.

---

### 11. DELETE /articles/{article_id}

Hapus artikel (hard delete). Hanya untuk draft/rejected.

**Permission**: `edit_news`

**Method**: DELETE — `/api/v1/web/news/articles/{article_id}`

**Response (200 OK)**:
```json
{ "message": "Article deleted" }
```

**Response (409 Conflict)**:
```json
{ "message": "Cannot delete published article. Archive it instead." }
```

---

## Bagian B — Source Management

### 12. GET /sources

List semua news source + config.

**Permission**: `manage_news_source` atau `manage_news_source_config`

**Method**: GET — `/api/v1/web/news/sources`

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "src_detik",
      "key": "detik",
      "name": "Detik.com",
      "base_url": "https://rss.detik.com/",
      "original_language": "id",
      "default_scope": { "id": "scp_id", "name": "Berita Indonesia", "slug": "indonesia" },
      "schedule": "0 */2 * * *",
      "last_scraped_at": "2026-06-04T08:00:00.000Z",
      "auto_publish": false,
      "ai_cleanup": true,
      "auto_translate": false,
      "is_active": true,
      "selectors": {
        "content_selector": ".detail-text",
        "author_selector": ".author",
        "tags_selector": ".detail-tag a",
        "extra_fields": null
      },
      "categories": [
        { "id": "sc_1", "category_key": "sport", "category_id": "cat_sport", "url_suffix": "sepakbola", "article_limit": 10, "is_active": true }
      ]
    }
  ]
}
```

---

### 13. POST /sources

Daftarkan source baru + selector.

**Permission**: `manage_news_source` (Usergod)

**Method**: POST — `/api/v1/web/news/sources`

**Request Body**:
```json
{
  "key": "kompas",
  "name": "Kompas",
  "base_url": "https://www.kompas.com/",
  "original_language": "id",
  "default_scope_id": "scp_id",
  "schedule": "0 */6 * * *",
  "auto_publish": false,
  "ai_cleanup": false,
  "auto_translate": false,
  "is_active": true,
  "selectors": {
    "content_selector": ".read__content",
    "author_selector": ".credit-title-name",
    "tags_selector": ".tag__article__item a",
    "extra_fields": { "multipage": true }
  }
}
```

**Response (201 Created)**:
```json
{ "data": { "id": "src_kompas", "key": "kompas" }, "message": "Source created" }
```

---

### 14. PUT /sources/{source_id}

Edit source. Usergod bisa ubah semua; Superadmin bisa ubah config operasional (schedule, auto_publish, ai_cleanup, auto_translate, default_scope_id, is_active).

**Permission**: `manage_news_source` (full) atau `manage_news_source_config` (config saja)

**Method**: PUT — `/api/v1/web/news/sources/{source_id}`

**Response (200 OK)**:
```json
{ "message": "Source updated" }
```

---

### 15. DELETE /sources/{source_id}

Hapus source (cascade ke selectors & categories).

**Permission**: `manage_news_source` (Usergod)

**Method**: DELETE — `/api/v1/web/news/sources/{source_id}`

**Response (200 OK)**:
```json
{ "message": "Source deleted" }
```

---

### 16. POST /sources/{source_id}/scrape-now

Trigger scraping manual untuk source ini (di luar jadwal).

**Permission**: `manage_news_source_config`

**Method**: POST — `/api/v1/web/news/sources/{source_id}/scrape-now`

**Response (202 Accepted)**:
```json
{ "message": "Scraping triggered" }
```

---

## Bagian C — Source Categories

### 17. POST /sources/{source_id}/categories

Tambah kategori scraping untuk source.

**Permission**: `manage_news_source_config` (Superadmin)

**Method**: POST — `/api/v1/web/news/sources/{source_id}/categories`

**Request Body**:
```json
{
  "category_key": "ekonomi",
  "category_id": "cat_eco",
  "url_suffix": "ekonomi/rss",
  "url_override": null,
  "article_limit": 15,
  "is_active": true
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `category_key` | `string` | Yes | Input manual — slug/path feed sesuai source |
| `category_id` | `string` | No | Mapping ke kategori KAI. Hasil scrape dari `category_key` ini di-assign ke kategori ini. Null = fallback `umum` |
| `url_suffix` | `string` | No | Append ke base_url |
| `url_override` | `string` | No | URL feed penuh jika pola beda |
| `article_limit` | `int` | No | Default 10 |
| `is_active` | `bool` | No | Default true |

**Response (201 Created)**:
```json
{ "data": { "id": "sc_new" }, "message": "Category added" }
```

---

### 18. PUT /sources/categories/{category_id}

Edit kategori scraping.

**Permission**: `manage_news_source_config`

**Method**: PUT — `/api/v1/web/news/sources/categories/{category_id}`

**Response (200 OK)**:
```json
{ "message": "Category updated" }
```

---

### 19. DELETE /sources/categories/{category_id}

Hapus kategori scraping.

**Permission**: `manage_news_source_config`

**Method**: DELETE — `/api/v1/web/news/sources/categories/{category_id}`

**Response (200 OK)**:
```json
{ "message": "Category deleted" }
```

---

## Bagian D — News Categories (Master)

### 20. GET /categories

List semua kategori news.

**Permission**: `manage_news_category`

**Method**: GET — `/api/v1/web/news/categories`

**Response (200 OK)**:
```json
{
  "data": [
    { "id": "cat_sport", "name": "Olahraga", "slug": "olahraga", "is_active": true, "sort_order": 1 }
  ]
}
```

---

### 21. POST /categories

Buat kategori news.

**Permission**: `manage_news_category` (Superadmin)

**Method**: POST — `/api/v1/web/news/categories`

**Request Body**:
```json
{ "name": "Teknologi", "slug": "teknologi", "is_active": true, "sort_order": 5 }
```

**Response (201 Created)**:
```json
{ "data": { "id": "cat_tech" }, "message": "Category created" }
```

---

### 22. PUT /categories/{category_id} & DELETE /categories/{category_id}

Edit / hapus kategori news.

**Permission**: `manage_news_category`

**Response (200 OK)**:
```json
{ "message": "Category updated" }
```

---

## Bagian D2 — News Scopes (Master)

Master scope asal/fokus geografis berita. Dikelola Superadmin via permission `manage_news_category` (sama dengan kategori — keduanya master klasifikasi konten).

### 22b. GET /scopes

List semua scope news.

**Permission**: `manage_news_category`

**Method**: GET — `/api/v1/web/news/scopes`

**Response (200 OK)**:
```json
{
  "data": [
    { "id": "scp_id", "name": "Berita Indonesia", "slug": "indonesia", "is_active": true, "sort_order": 1 },
    { "id": "scp_kr", "name": "Berita Korea", "slug": "korea", "is_active": true, "sort_order": 2 },
    { "id": "scp_kri", "name": "Berita Korea di Indonesia", "slug": "korea_indonesia", "is_active": true, "sort_order": 3 }
  ]
}
```

---

### 22c. POST /scopes

Buat scope baru.

**Permission**: `manage_news_category` (Superadmin)

**Method**: POST — `/api/v1/web/news/scopes`

**Request Body**:
```json
{ "name": "Berita Korea di Indonesia", "slug": "korea_indonesia", "is_active": true, "sort_order": 3 }
```

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | `string` | Yes | Nama tampil |
| `slug` | `string` | Yes | Unik, huruf kecil (mis. `korea_indonesia`) |
| `is_active` | `bool` | No | Default true |
| `sort_order` | `int` | No | Default 0 |

**Response (201 Created)**:
```json
{ "data": { "id": "scp_kri" }, "message": "Scope created" }
```

---

### 22d. PUT /scopes/{scope_id} & DELETE /scopes/{scope_id}

Edit / hapus scope news. Saat scope dihapus, `articles.news_scope_id` & `news_sources.default_scope_id` yang merujuknya di-set null (ON DELETE SET NULL).

**Permission**: `manage_news_category`

**Method**: PUT / DELETE — `/api/v1/web/news/scopes/{scope_id}`

**Response (200 OK)**:
```json
{ "message": "Scope updated" }
```

---

## Bagian E — System Languages

### 23. GET /languages

List semua bahasa sistem.

**Permission**: `manage_news_settings`

**Method**: GET — `/api/v1/web/news/languages`

**Response (200 OK)**:
```json
{
  "data": [
    { "code": "id", "name": "Indonesian", "is_ui_language": true, "is_translate_target": true, "is_active": true },
    { "code": "en", "name": "English", "is_ui_language": true, "is_translate_target": true, "is_active": true }
  ]
}
```

---

### 24. POST /languages & PUT /languages/{code}

Tambah / edit bahasa sistem.

**Permission**: `manage_news_settings` (Usergod)

**Method**: POST — `/api/v1/web/news/languages`

**Request Body**:
```json
{ "code": "jp", "name": "Japanese", "is_ui_language": false, "is_translate_target": true, "is_active": true }
```

**Response (201 Created)**:
```json
{ "data": { "code": "jp" }, "message": "Language added" }
```

---

## Bagian F — System Settings

### 25. GET /settings

Ambil konfigurasi global News.

**Permission**: `manage_news_settings`

**Method**: GET — `/api/v1/web/news/settings`

**Response (200 OK)**:
```json
{
  "data": {
    "translation_enabled": true,
    "on_demand_enabled": true,
    "updated_by": "usr_god",
    "updated_at": "2026-06-01T00:00:00.000Z"
  }
}
```

---

### 26. PUT /settings

Update konfigurasi global.

**Permission**: `manage_news_settings`

**Method**: PUT — `/api/v1/web/news/settings`

**Request Body**:
```json
{ "translation_enabled": true, "on_demand_enabled": false }
```

**Response (200 OK)**:
```json
{ "message": "Settings updated" }
```

---

## Error Responses

### 400 / 401 / 403 / 404 / 409 / 422
```json
{ "message": "..." }
```

```json
// 403
{ "message": "You don't have permission to manage news sources" }

// 409
{ "message": "Cannot delete published article. Archive it instead." }

// 422
{
  "message": "Validation failed",
  "errors": { "schedule": ["Invalid cron expression"] }
}
```

---

## Status Codes

- `200 OK` — Success
- `201 Created` — Resource created
- `202 Accepted` — Async job enqueued (scrape, translate)
- `400 Bad Request` — Validation error
- `401 Unauthorized` — Auth required
- `403 Forbidden` — Permission denied
- `404 Not Found` — Resource not found
- `409 Conflict` — State conflict
- `422 Unprocessable Entity` — Validation error
- `500 Internal Server Error` — Server error

---

## Rate Limiting

- Limit: 200 requests per minute per admin
- Headers: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset`

---

*API spec News untuk web backoffice.*
