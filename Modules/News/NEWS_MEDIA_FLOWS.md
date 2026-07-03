# News Article — Media Lifecycle Flows

Dokumen ini menjelaskan mekanisme upload, konfirmasi, update, dan delete media (thumbnail) untuk artikel News.

---

## 1. Konsep Prefix Scheme

Semua media (thumbnail) disimpan di DB dengan **prefix**:

| Prefix | Arti | Contoh |
|--------|------|--------|
| `s3:` | Objek di S3/MinIO internal | `s3:/news/thumbnails/uuid.jpg` |
| `ext:` | URL eksternal | `ext:https://partner.com/logo.png` |
| *(tanpa prefix)* | Data lama (backward compat) | `/news/thumbnails/old.jpg` |

Tabel `media_uploads` melacak lifecycle file S3 (status: PENDING → CONFIRMED → DELETED).
File `ext:` tidak pernah masuk ke `media_uploads`.

---

## 2. Upload Flow (Presign → Confirm → Create)

```
UI                         API                          MinIO / media_uploads
──                         ───                          ─────────────────────
  │                          │                                │
  │── POST /presign ───────> │                                │
  │    { content_type }      │                                │
  │                          │── RequestUpload ────────────> │
  │                          │   Generate presigned PUT URL  │
  │                          │   Save record → PENDING       │
  │<── Response ──────────── │                                │
  │    { key: "s3:/news/     │                                │
  │       thumbnails/uuid    │                                │
  │       .jpg",             │                                │
  │      presign_url,        │                                │
  │      public_url }        │                                │
  │                          │                                │
  │── PUT <presign_url> ──── │ ──────────────────────────>   │
  │    (file langsung ke     │          MinIO                 │
  │     MinIO, tanpa API)    │                                │
  │                          │                                │
  │── POST /articles ──────> │                                │
  │    { thumbnail_url:      │                                │
  │      "s3:/news/          │                                │
  │       thumbnails/         │                                │
  │       uuid.jpg" }        │                                │
  │                          │── NormalizeValue ────────────> │
  │                          │   ("s3:..." → return as-is)    │
  │                          │                                │
  │                          │── Save article + translation ─>│
  │                          │                                │
  │                          │── ConfirmUpload ────────────> │
  │                          │   PENDING → CONFIRMED         │
  │<── Response ──────────── │                                │
```

### Detail Langkah

**Step 1 — Presign** (`POST /web/news/media/thumbnail/presign`)
- Validasi `content_type` (jpeg/png/webp/gif)
- Generate object path: `/news/thumbnails/{uuid}.jpg`
- Panggil `mediaUploadSvc.RequestUpload()`:
  - Generate presigned PUT URL via FileUploader
  - Simpan record `media_uploads` dengan status **PENDING**
- Return `key: "s3:/news/thumbnails/uuid.jpg"` (prefixed)

**Step 2 — Upload langsung ke MinIO**
- Client PUT ke `presign_url` dengan file binary
- Tidak melalui API backend

**Step 3a — Confirm eksplisit** (`POST /web/news/media/thumbnail/confirm?key=...`)
- `NormalizeValue(key)` → handle `s3:` / `ext:` / backward compat
- `ConfirmUpload` → PENDING → **CONFIRMED**
- Return `public_url`

**Step 3b — Confirm implisit via payload** (`POST /web/news/articles`)
- Dalam `CreateArticleUseCase.Execute()`:
  1. `normalizeNewsThumbnailKeyPtr(svc, req.Translation.ThumbnailURL)` → NormalizeValue
  2. Save translation (ThumbnailURL tersimpan dengan prefix `s3:`)
  3. `confirmNewsThumbnailKey(ctx, svc, translation.ThumbnailURL)` → ConfirmUpload

---

## 3. Update Flow

### 3a. Update — Thumbnail diganti

```
UI                          API
──                          ───
  │── GET /articles/{id} ──> │
  │<── { translations: [{    │
  │       thumbnail_url:     │
  │        "https://cdn.../  │
  │         uuid.jpg",       │
  │       thumbnail_raw:     │
  │        "s3:/news/        │
  │         thumbnails/      │
  │         uuid.jpg" }] }   │
  │                          │
  │── PUT /articles/{id} ──> │
  │    { translation: {      │
  │      thumbnail_url:      │
  │       "s3:/news/         │
  │        thumbnails/       │
  │        NEW.jpg" } }      │
  │                          │
  │── Fetch old translation ─>│
  │   old = "s3:/news/       │
  │          thumbnails/     │
  │          OLD.jpg"        │
  │                          │
  │── NormalizeValue(old) ──>│
  │── NormalizeValue(new) ──>│
  │                          │
  │── old != new?            │
  │   ├── Ya → MarkDeleted(  │
  │   │        old)          │
  │   │       CONFIRMED →    │
  │   │       DELETED        │
  │   │   → ConfirmUpload(   │
  │   │        new)          │
  │   │       PENDING →      │
  │   │       CONFIRMED      │
  │   └── Tidak → skip       │
  │                          │
  │── SaveOrUpdate ─────────>│
  │   translation            │
```

### 3b. Update — Thumbnail dihapus (nil / "")

```
UI                          API
──                          ───
  │── PUT /articles/{id} ──> │
  │    { translation: {      │
  │      thumbnail_url: null │
  │    } }                   │
  │                          │
  │── normalizeNewsThumbnail │
  │    KeyPtr(svc, nil)      │
  │    → return nil          │
  │                          │
  │── old != new?            │
  │   oldNorm != ""          │
  │   newNorm == ""          │
  │   → MarkDeleted(oldNorm) │
  │     CONFIRMED → DELETED  │
  │                          │
  │── Save translation       │
  │    ThumbnailURL = nil    │
```

### 3c. Update — Thumbnail tidak berubah

```
UI                          API
──                          ───
  │── PUT /articles/{id} ──> │
  │    { translation: {      │
  │      thumbnail_url:      │
  │       "https://cdn.../   │
  │        uuid.jpg" } }     │
  │    (CDN URL dari GET     │
  │     response)            │
  │                          │
  │── NormalizeValue(cdnURL) │
  │   KeyFromURL → extract   │
  │   → "s3:/news/thumbnails │
  │       /uuid.jpg"         │
  │                          │
  │── oldNorm == newNorm?    │
  │   → Skip (no-op)         │
  │                          │
  │── SaveOrUpdate ─────────>│
  │   translation            │
```

### Kode yang menangani (update_article.go)

```go
// 1. Ambil old translation
oldTranslation, _ := uc.translationRepo.FindByArticleAndLanguage(ctx, articleID, article.OriginalLanguage)

// 2. Normalize old vs new
oldNorm := uc.mediaUploadSvc.NormalizeValue(*oldThumbnail)
newNorm := uc.mediaUploadSvc.NormalizeValue(*newThumbnail)

// 3. Mark deleted jika berbeda
if oldNorm != "" && oldNorm != newNorm {
    _ = uc.mediaUploadSvc.MarkDeleted(ctx, oldNorm)
}

// 4. Simpan translation baru
uc.translationRepo.SaveOrUpdate(ctx, translation)

// 5. Confirm thumbnail baru
confirmNewsThumbnailKey(ctx, uc.mediaUploadSvc, translation.ThumbnailURL)
```

---

## 4. Delete Article Flow

```
UI                          API
──                          ───
  │── DELETE /articles/{id}  │
  │       ─────────────────> │
  │                          │
  │── EnsureDeletable() ────>│
  │   (hanya draft/rejected) │
  │                          │
  │── FindAllByArticle ─────>│
  │   Ambil semua            │
  │   translations           │
  │                          │
  │── Untuk setiap           │
  │   translation:           │
  │   MarkDeleted(           │
  │    translation.          │
  │    ThumbnailURL)         │
  │   CONFIRMED → DELETED    │
  │                          │
  │── articleRepo.Delete ───>│
  │   (CASCADE ke            │
  │    translations, likes,  │
  │    comments, dll)        │
```

---

## 5. Add Translation Flow

```
UI                          API
──                          ───
  │── PUT /articles/{id}/   │
  │    translations/{lang}  │
  │    { thumbnail_url:     │
  │      "s3:/..." }        │
  │       ─────────────────> │
  │                          │
  │── NormalizeValue ───────>│
  │── SaveOrUpdate ─────────>│
  │── ConfirmUpload ────────>│
```

Catatan: Add translation **tidak** melakukan mark deleted untuk thumbnail lama — karena setiap translation per bahasa punya thumbnail masing-masing yang independen.

---

## 6. Ringkasan Usecase & File

| Flow | Usecase File | Handler Endpoint | Media Action |
|------|-------------|-----------------|-------------|
| Presign | `get_thumbnail_presign_url.go` | `POST /web/news/media/thumbnail/presign` | `RequestUpload` → PENDING |
| Confirm eksplisit | `confirm_thumbnail.go` | `POST /web/news/media/thumbnail/confirm` | `ConfirmUpload` → CONFIRMED |
| Delete eksplisit | `delete_thumbnail.go` | `DELETE /web/news/media/thumbnail` | `MarkDeleted` → DELETED |
| Create article | `create_article.go` | `POST /web/news/articles` | `NormalizeValue` + `ConfirmUpload` |
| Submit article | `submit_article.go` | `POST /mobile/news/articles` | `NormalizeValue` + `ConfirmUpload` |
| Update article | `update_article.go` | `PUT /web/news/articles/{id}` | Compare old/new → `MarkDeleted` + `ConfirmUpload` |
| Add translation | `add_translation.go` | `PUT /web/news/articles/{id}/translations/{lang}` | `NormalizeValue` + `ConfirmUpload` |
| Delete article | `delete_article.go` | `DELETE /web/news/articles/{id}` | `MarkDeleted` semua thumbnail |

---

## 7. DTO Response Fields

| Field | Contoh | Kegunaan |
|-------|--------|----------|
| `thumbnail_url` | `https://cdn.../uuid.jpg` | URL akses untuk preview gambar |
| `thumbnail_raw` | `s3:/news/thumbnails/uuid.jpg` | Nilai asli dari DB (dengan prefix). UI kirim balik field ini saat PUT jika tidak ada perubahan |

Untuk `ext:` type, `thumbnail_raw` tetap muncul:
```json
{
  "thumbnail_url": "https://lh3.google.com/photo.jpg",
  "thumbnail_raw": "ext:https://lh3.google.com/photo.jpg"
}
```
