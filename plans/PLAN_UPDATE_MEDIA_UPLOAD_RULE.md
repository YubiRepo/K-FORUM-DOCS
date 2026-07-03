# Plan: Media Storage Prefix Scheme (`s3:` dan `ext:`)

## Context

Saat ini semua field media di domain entities menyimpan **raw key** saja (contoh: `/ads/images/uuid.jpg`). Tidak ada cara membedakan apakah nilai tersebut adalah objek S3 internal atau URL eksternal.

Kebutuhan baru:
- Beberapa field media menerima **URL eksternal** (tidak di-upload ke S3 kita), misalnya Google avatar.
- MediaService harus bisa membedakan keduanya untuk: lifecycle tracking (`media_uploads`), cleanup, dan resolusi public URL.
- UI perlu skema yang jelas untuk memberi tahu API: "ini S3" vs "ini URL eksternal".

Solusi: tambahkan **prefix** pada nilai yang disimpan di DB dan dikirim UI:
- `s3:<key>` → objek di S3/MinIO kita (contoh: `s3:/ads/images/uuid.jpg`)
- `ext:<url>` → URL eksternal langsung (contoh: `ext:https://lh3.googleusercontent.com/photo.jpg`)

---

## 1. Prefix Scheme

### Penyimpanan di DB (domain entity tables)

| Sebelum | Sesudah |
|---------|---------|
| `/ads/images/uuid.jpg` | `s3:/ads/images/uuid.jpg` |
| `https://lh3.googleusercontent.com/photo.jpg` | `ext:https://lh3.googleusercontent.com/photo.jpg` |
| *(kosong)* | *(kosong / null)* |

Tabel `media_uploads` (lifecycle tracking) tetap menyimpan key **tanpa prefix** — ini internal concern service.

### Backward Compatibility

Untuk data lama di DB (tanpa prefix): `NormalizeValue` dan `PublicURL` akan auto-detect:
- Dimulai dengan `/` → anggap `s3:` key (backward compat)
- URL CDN kita → strip ke key, anggap `s3:`
- Tidak ada perubahan migrasi DB yang dibutuhkan

---

## 2. Perubahan MediaUploadService

### Port Interface — `internal/app/port/media_upload_service.go`

Rename `NormalizeKey` → `NormalizeValue`, update signature dan doc:

```go
type MediaUploadService interface {
    RequestUpload(ctx, objectName, contentType, expires, privacy) (presignURL, prefixedKey, publicURL string, err error)
    ConfirmUpload(ctx context.Context, prefixedValue string) error  // no-op untuk ext:
    MarkDeleted(ctx context.Context, prefixedValue string) error    // no-op untuk ext:
    PublicURL(prefixedValue string) string                          // handles s3: dan ext:
    NormalizeValue(rawInput string) string                          // returns s3:<key> or ext:<url>
}
```

### Service Implementation — `internal/app/service/media/media_upload_service.go`

Tambahkan helper internal:

```go
func parsePrefix(v string) (prefix, raw string) {
    if after, ok := strings.CutPrefix(v, "s3:"); ok {
        return "s3", after
    }
    if after, ok := strings.CutPrefix(v, "ext:"); ok {
        return "ext", after
    }
    return "", v
}
```

Implementasi method-method baru:

```go
// NormalizeValue: input dari UI → nilai dengan prefix untuk disimpan di DB
func (s *mediaUploadService) NormalizeValue(rawInput string) string {
    if rawInput == "" {
        return ""
    }
    prefix, raw := parsePrefix(rawInput)
    if prefix == "s3" || prefix == "ext" {
        return rawInput // sudah punya prefix, langsung return
    }
    // Backward compat: cek apakah URL CDN kita
    key := s.fileUploader.KeyFromURL(raw)
    if key != "" {
        return "s3:" + key // URL CDN kita → s3: key
    }
    if strings.HasPrefix(raw, "/") {
        return "s3:" + raw // plain key → s3:
    }
    return "ext:" + raw // URL eksternal lainnya
}

// PublicURL: nilai dari DB → URL publik untuk response DTO
func (s *mediaUploadService) PublicURL(prefixedValue string) string {
    if prefixedValue == "" {
        return ""
    }
    prefix, raw := parsePrefix(prefixedValue)
    switch prefix {
    case "s3":
        return s.fileUploader.PublicURL(raw)
    case "ext":
        return raw // URL eksternal dikembalikan langsung
    default:
        // backward compat: plain key atau CDN URL lama
        if strings.HasPrefix(raw, "/") {
            return s.fileUploader.PublicURL(raw)
        }
        return raw
    }
}

// ConfirmUpload: hanya act untuk s3:, no-op untuk ext:
func (s *mediaUploadService) ConfirmUpload(ctx context.Context, prefixedValue string) error {
    prefix, raw := parsePrefix(prefixedValue)
    if prefix == "ext" || prefixedValue == "" {
        return nil
    }
    return s.doConfirmUpload(ctx, raw) // raw = key tanpa prefix
}

// MarkDeleted: hanya act untuk s3:, no-op untuk ext:
func (s *mediaUploadService) MarkDeleted(ctx context.Context, prefixedValue string) error {
    prefix, raw := parsePrefix(prefixedValue)
    if prefix == "ext" || prefixedValue == "" {
        return nil
    }
    return s.doMarkDeleted(ctx, raw)
}

// RequestUpload: tambahkan s3: prefix pada key yang dikembalikan
func (s *mediaUploadService) RequestUpload(...) (presignURL, prefixedKey, publicURL string, err error) {
    // ... existing logic ...
    prefixedKey = "s3:" + objectKey // objectKey = "/ads/images/uuid.jpg"
    return presignURL, prefixedKey, publicURL, nil
}
```

---

## 3. Perubahan Presign Response (semua domain)

Semua presign response DTO mengembalikan field `key`. Setelah perubahan, `key` akan berisi prefix `s3:`.

**Tidak perlu ubah DTO struct** — nilainya otomatis berubah karena `RequestUpload()` sekarang return prefixed key.

Contoh response baru:
```json
{
  "presign_url": "https://minio.../signed-PUT-url",
  "key": "s3:/ads/images/uuid.png",
  "public_url": "https://cdn.../ads/images/uuid.png",
  "expires_in": 3600
}
```

UI tinggal gunakan `key` langsung di payload create/update tanpa transformasi tambahan.

---

## 4. Update Semua Usecase: `NormalizeKey` → `NormalizeValue`

**13 domain helpers** perlu di-update. Pattern perubahan identik di semua domain:

```go
// SEBELUM (di create/update usecase)
key := u.mediaUploadSvc.NormalizeKey(input.ImageURL)
entity.ImageURL = key
// ...save...
u.confirmMediaKey(ctx, key)

// SESUDAH
value := u.mediaUploadSvc.NormalizeValue(input.ImageURL)
entity.ImageURL = value
// ...save...
u.confirmMediaKey(ctx, value) // ConfirmUpload handles ext: as no-op
```

**File helpers yang perlu diubah** (panggilan `markMediaDeleted` juga sudah handle prefix via `MarkDeleted`):
- `internal/app/usecase/ads/helpers.go`
- `internal/app/usecase/announcement/helpers.go`
- `internal/app/usecase/community/helpers.go`
- `internal/app/usecase/event/helpers.go`
- `internal/app/usecase/news/helpers.go`
- `internal/app/usecase/region/helpers.go`
- `internal/app/usecase/directory/helpers.go`
- `internal/app/usecase/profile/helpers.go`
- `internal/app/usecase/usermanagement/helpers.go`
- `internal/app/usecase/qna/helpers.go`
- `internal/app/usecase/subscription/helpers.go`
- `internal/app/usecase/accounting/helpers.go`
- `internal/app/usecase/bug_report/helpers.go`

**File create/update usecase** yang langsung call `NormalizeKey` (di luar helpers): perlu grep dan update juga.

**DTO building** — semua tempat yang call `mediaUploadSvc.PublicURL(entity.XxxURL)` atau `fileUploader.PublicURL(entity.XxxURL)` akan tetap bekerja karena `PublicURL` diupdate untuk handle prefix. Pastikan semua call melalui `mediaUploadSvc.PublicURL()`, bukan langsung ke `fileUploader.PublicURL()`.

---

## 5. Perubahan Google Login — `internal/app/usecase/auth/google_login.go`

**Hapus** fungsi `fetchAndUploadGoogleAvatar()` (download + re-upload ke S3).

**Ganti** dengan simpan URL Google langsung dengan prefix `ext:`:

```go
// SEBELUM (lines ~310-346)
if pictureURL != "" {
    if key, err := u.fetchAndUploadGoogleAvatar(ctx, userID, pictureURL); err == nil {
        profile.SetAvatar(key)
    }
}

// SESUDAH
if pictureURL != "" {
    profile.SetAvatar("ext:" + pictureURL)
}
```

Hapus seluruh fungsi `fetchAndUploadGoogleAvatar()` dan dependensi terkait (`fileUploader` injection jika hanya dipakai untuk ini).

---

## 6. UI Schema — Cara Membedakan s3: vs ext:

### Kontrak untuk semua field media di Create/Update Request

| Skenario | Yang UI kirim | Contoh |
|----------|--------------|--------|
| Upload baru via presign | `key` dari presign response (sudah `s3:`) | `s3:/ads/images/abc.jpg` |
| URL eksternal | Manual prefix `ext:` + URL | `ext:https://example.com/img.jpg` |
| Tidak berubah (PATCH) | Kirim ulang nilai yang diterima dari GET response (API akan re-normalize) | `https://cdn.../ads/images/abc.jpg` → API detect ke `s3:` |
| Hapus media | `""` atau `null` | |

### Alur Upload S3 (Presign Flow)

```
1. UI → POST /[domain]/media/[type]/presign
         Response: { key: "s3:/ads/images/uuid.jpg", presign_url: "...", ... }

2. UI → PUT presign_url  (langsung ke MinIO, tanpa API)

3. UI → POST /[domain]/create  dengan payload: { image_url: "s3:/ads/images/uuid.jpg" }
         (key dari step 1 dipakai langsung — tidak perlu transformasi)
```

### Alur URL Eksternal

```
1. UI punya URL eksternal: "https://partner.com/logo.png"
2. UI prefix manual: "ext:https://partner.com/logo.png"
3. UI → POST /[domain]/create  dengan payload: { icon_url: "ext:https://partner.com/logo.png" }
```

### Confirm/Delete Endpoints

- Menerima format baru: `?key=s3:/ads/images/uuid.jpg`
- Menerima format lama (backward compat): `?key=/ads/images/uuid.jpg`
- `ext:` keys: no-op (tidak ada entry di `media_uploads`)

### GET Response (tidak berubah)

API selalu mengembalikan **full URL tanpa prefix** di response:
```json
{
  "image_url": "https://cdn.example.com/ads/images/uuid.jpg"
}
```
atau untuk eksternal:
```json
{
  "avatar": "https://lh3.googleusercontent.com/photo.jpg"  
}
```

UI tidak perlu tahu tentang prefix saat membaca data.

---

## 7. Domain & Use Case Checklist

| Domain | Fields | Action |
|--------|--------|--------|
| **Profile** | `AvatarURL` | Update NormalizeKey → NormalizeValue |
| **User Management** | `AvatarURL` | Update NormalizeKey → NormalizeValue |
| **Community** | `AvatarURL`, `Media[].URL`, `Media[].ThumbURL` | Update NormalizeKey → NormalizeValue |
| **Event** | `Images[]`, `CoverImage` | Update NormalizeKey → NormalizeValue |
| **Announcement** | `ImageURL` | Update NormalizeKey → NormalizeValue |
| **Region** | `ImageURL` | Update NormalizeKey → NormalizeValue |
| **Ads** | `ImageURL`, `VideoURL`, `ThumbnailURL`, `IconURL`, `SponsorLogoURL` | Update NormalizeKey → NormalizeValue |
| **News** | `ThumbnailURL` | Update NormalizeKey → NormalizeValue |
| **Directory** | Merchant/Item image fields | Update NormalizeKey → NormalizeValue |
| **QnA** | `AttachmentURLs[]` | Update NormalizeKey → NormalizeValue |
| **Subscription** | `ManualProofURL` | Update NormalizeKey → NormalizeValue |
| **Accounting** | `AttachmentURL` | Update NormalizeKey → NormalizeValue |
| **Bug Report** | Attachment fields | Update NormalizeKey → NormalizeValue |
| **Auth/Google Login** | `AvatarURL` | Hapus re-upload logic → simpan `ext:` |

---

## 8. File Kritis yang Dimodifikasi

| File | Perubahan |
|------|-----------|
| `internal/app/port/media_upload_service.go` | Rename `NormalizeKey` → `NormalizeValue`, update docs |
| `internal/app/service/media/media_upload_service.go` | Tambah `parsePrefix()`, update semua method |
| `internal/app/usecase/auth/google_login.go` | Ganti `fetchAndUploadGoogleAvatar` dengan `ext:` prefix |
| `internal/app/usecase/*/helpers.go` (13 files) | `NormalizeKey` → `NormalizeValue` |
| Semua create/update usecases yang call NormalizeKey langsung | Rename panggilan |

---

---

## 10. Pembuatan & Update API Spec

### Buat File Spec Baru (Central Reference)

**File**: `K-FORUM-DOCS/API SPEC/MEDIA_UPLOAD_SPEC.md`

Dokumen ini menjadi referensi utama untuk seluruh tim (frontend web, mobile, backend) tentang sistem media upload. Isinya:

1. **Konsep prefix scheme** — apa itu `s3:` vs `ext:`, kapan dipakai
2. **Generic presign flow** — step-by-step yang berlaku untuk semua domain
3. **Format key di request vs response**:
   - Request (write): `s3:<key>` atau `ext:<url>`
   - Response (read): full CDN URL tanpa prefix
4. **Contoh JSON** untuk setiap skenario
5. **Daftar presign/confirm/delete endpoints** per domain (tabel lengkap)
6. **Aturan confirm/delete**: `s3:` → tracked di `media_uploads`; `ext:` → no-op
7. **Backward compat**: nilai lama tanpa prefix diterima API, tetap bisa dibaca

---

### Update Spec Backoffice Web (K-FORUM-DOCS/API SPEC/Web/)

Semua spec ini perlu update karena saat ini hanya menampilkan full URL di request — padahal setelah perubahan ini, request harus pakai format `s3:` atau `ext:`.

| File | Yang Perlu Diubah |
|------|------------------|
| `API_SPEC_ADS_BACKOFFICE.md` | (1) Tambah endpoint presign/confirm/delete untuk image, video, icon, sponsor_logo; (2) Update contoh request `image_url`, `video_url`, `thumbnail_url`, `icon_url`, `sponsor_logo_url` → format `s3:` atau `ext:` |
| `API_SPEC_REGION_BACKOFFICE.md` | Update contoh request `image_url` → format `s3:`; tambahkan presign endpoint jika belum ada |
| `API_SPEC_ANNOUNCEMENT_BACKOFFICE.md` | Tambah presign/confirm/delete endpoints; update `image_url` di request → `s3:` |
| `API_SPEC_COMMUNITY_BACKOFFICE.md` | Tambah presign/confirm/delete untuk avatar dan post media; update `avatar_url` di request |
| `API_SPEC_EVENT_BACKOFFICE.md` | Tambah presign/confirm/delete; update `images[]` di request → format `s3:` |
| `API_SPEC_NEWS_BACKOFFICE.md` | Tambah presign/confirm/delete; update `thumbnail_url` → format `s3:` |
| `API_SPEC_USER_MANAGEMENT_BACKOFFICE.md` | Update `avatar_url` di request → format `s3:`; tambah presign endpoint |
| `API_SPEC_QNA_BACKOFFICE.md` | Tambah presign/confirm/delete untuk attachments; update `attachment_urls[]` |
| `DIRECTORY_API_SPEC_BACKOFFICE_V2.md` | Tambah presign endpoints; update field image di request → `s3:` |
| `API_SPEC_SUBSCRIPTION_WEB.md` | Update `manual_proof_url` di request → format `s3:` |

**Pola perubahan contoh JSON yang dipakai di semua file ini:**

```json
// Request (CREATE/UPDATE) — SEBELUM
{ "image_url": "https://cdn.example.com/ads/images/uuid.jpg" }

// Request (CREATE/UPDATE) — SESUDAH  
{ "image_url": "s3:/ads/images/uuid.jpg" }      // untuk upload via presign
{ "image_url": "ext:https://partner.com/logo.png" }  // untuk URL eksternal

// Response (GET) — TIDAK BERUBAH
{ "image_url": "https://cdn.example.com/ads/images/uuid.jpg" }
```

**Pola contoh presign response yang diupdate:**

```json
// SEBELUM
{ "key": "/ads/images/uuid.png", "presign_url": "...", "public_url": "..." }

// SESUDAH
{ "key": "s3:/ads/images/uuid.png", "presign_url": "...", "public_url": "..." }
```

---

### Update Spec Mobile (K-FORUM-DOCS/API SPEC/Mobile/)

Mobile **beralih ke presign + s3:/ext: scheme** (sama dengan backoffice). Spec lama yang menggunakan `multipart/form-data` perlu diganti total.

| File | Yang Perlu Diubah |
|------|------------------|
| `API_SPEC_PROFILE_MEMBERSHIP.md` | **Major update**: Hapus endpoint `POST /mobile/profile/avatar` dengan multipart. Ganti dengan 3 endpoint presign/confirm/delete. Update contoh request avatar upload ke format `s3:`. |
| `API_SPEC_AUTH.md` | Tambah catatan: setelah Google login, `avatar` di response adalah URL Google langsung (ext:) — tidak di-upload ke S3 kita lagi |
| `API_SPEC_COMMUNITY_MOBILE.md` | Tambah/update presign endpoints untuk community avatar dan post media; update contoh request media → `s3:` format |
| `API_SPEC_QNA_MOBILE.md` | Tambah presign/confirm/delete untuk attachment; update `attachment_urls[]` → `s3:` format |
| `API_SPEC_BUG_REPORT_MOBILE.md` | Tambah presign endpoints untuk attachment |
| `API_SPEC_EVENT_MOBILE.md` | Update media field format jika ada upload dari mobile |
| `API_SPEC_Subscription_Mobile.md` | Update `manual_proof_url` → presign flow + `s3:` format |

**Pola presign endpoints baru untuk mobile** (contoh avatar):
```
POST /api/v1/mobile/profile/avatar/presign   → { key: "s3:/user/uuid.jpg", presign_url, public_url }
PUT  <presign_url>                           → (direct to MinIO)
POST /api/v1/mobile/profile/avatar/confirm   → ?key=s3:/user/uuid.jpg
DELETE /api/v1/mobile/profile/avatar         → ?key=s3:/user/uuid.jpg
```

---

### Update Contract File (k-forum-backoffice/docs/contracts/)

**File**: `ADS_MEDIA_UPLOAD_CONTRACT.md`
- Update status (dari "NOT YET IMPLEMENTED" → implemented)
- Update contoh `key` di response presign → `s3:/ads/...`
- Tambah dokumentasi `ext:` option untuk `icon_url` dan `sponsor_logo_url`

---

## 11. Verifikasi

1. **Unit test MediaUploadService**: Test `NormalizeValue`, `PublicURL`, `ConfirmUpload`, `MarkDeleted` dengan input `s3:`, `ext:`, dan unprefixed (backward compat).

2. **Integration test Google login**: Login dengan Google → cek `user_profiles.avatar_url` di DB berisi `ext:https://...` (bukan re-upload key).

3. **Presign flow test**: 
   - Panggil `/ads/media/image/presign`
   - Cek response `key` field berisi `s3:/ads/images/...`
   - Buat ad dengan `image_url: "s3:/ads/images/..."` 
   - Cek DB menyimpan `s3:/ads/images/...`
   - GET ad → response `image_url` adalah full CDN URL (tanpa prefix)

4. **Ext flow test**: Buat entity dengan `image_url: "ext:https://example.com/img.jpg"` → DB simpan `ext:...` → GET response kembalikan `https://example.com/img.jpg`.

5. **Backward compat test**: Data lama di DB (plain key `/ads/images/old.jpg`) → GET response tetap kembalikan full CDN URL dengan benar.
