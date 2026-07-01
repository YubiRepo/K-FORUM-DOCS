# Media Upload Spec — Central Reference

Dokumen ini menjadi referensi utama untuk frontend (web & mobile) dan backend tentang sistem media upload menggunakan prefix scheme `s3:` dan `ext:`.

---

## 1. Konsep Prefix Scheme

Setiap nilai media yang disimpan di database kini memiliki **prefix** untuk membedakan sumber:

| Prefix | Arti | Contoh |
|--------|------|--------|
| `s3:` | Objek di S3/MinIO internal | `s3:/ads/images/uuid.jpg` |
| `ext:` | URL eksternal (bukan S3 kita) | `ext:https://lh3.googleusercontent.com/photo.jpg` |
| *(tanpa prefix)* | Data lama (backward compat) | `/ads/images/uuid.jpg` |

Tabel `media_uploads` (lifecycle tracking) tetap menyimpan key **tanpa prefix** — internal service concern.

---

## 2. Generic Presign Flow

Flow yang berlaku untuk semua domain:

```
1. UI → POST /[domain]/media/[type]/presign
         Body: { "content_type": "image/jpeg" }
         Response: { "key": "s3:/path/uuid.jpg", "presign_url": "https://...", "public_url": "https://...", "expires_in": 3600 }

2. UI → PUT <presign_url>    (langsung ke MinIO, tanpa melalui API)

3. UI → POST /[domain]/create  atau  PUT /[domain]/:id
         Payload: { "image_url": "s3:/path/uuid.jpg" }
         (key dari step 1 dipakai langsung tanpa transformasi)
```

---

## 3. Format Key di Request vs Response

### Request (Write) — Create / Update

UI mengirim nilai **dengan prefix**:

| Skenario | Yang UI kirim | Contoh |
|----------|--------------|--------|
| Upload baru via presign | `key` dari presign response (sudah `s3:`) | `s3:/ads/images/abc.jpg` |
| URL eksternal | Manual prefix `ext:` + URL | `ext:https://partner.com/logo.png` |
| Tidak berubah (PATCH) | Kirim ulang nilai dari GET response (full URL) | `https://cdn.../abc.jpg` → API detect `s3:` |
| Hapus media | `""` atau `null` | |

### Response (Read) — GET / Detail

API mengembalikan **dua field** untuk setiap media:

| Field | Format | Kegunaan |
|-------|--------|----------|
| `xxx_url` | Full CDN URL | Preview gambar di UI (`<img src="...">`) |
| `xxx_raw` | Nilai asli DB (`s3:` / `ext:` / plain key) | Deteksi tipe media & kirim balik saat PUT |

Contoh:
```json
{
  "thumbnail_url": "https://cdn.example.com/news/thumbnails/uuid.jpg",
  "thumbnail_raw": "s3:/news/thumbnails/uuid.jpg"
}
```
Untuk URL eksternal:
```json
{
  "avatar_url": "https://lh3.googleusercontent.com/photo.jpg",
  "avatar_raw": "ext:https://lh3.googleusercontent.com/photo.jpg"
}
```

---

## 4. Contoh JSON Lengkap

### Presign Response
```json
// SEBELUM (OLD)
{ "key": "/ads/images/uuid.png", "presign_url": "...", "public_url": "..." }

// SESUDAH (NEW)
{ "key": "s3:/ads/images/uuid.png", "presign_url": "...", "public_url": "..." }
```

### Create/Update Request
```json
// SEBELUM (OLD) — full URL
{ "image_url": "https://cdn.example.com/ads/images/uuid.jpg" }

// SESUDAH (NEW) — upload via presign
{ "image_url": "s3:/ads/images/uuid.jpg" }

// SESUDAH (NEW) — URL eksternal
{ "image_url": "ext:https://partner.com/logo.png" }
```

### GET Response
```json
{ "image_url": "https://cdn.example.com/ads/images/uuid.jpg", "image_raw": "s3:/ads/images/uuid.jpg" }
```

---

## 5. Daftar Endpoint Presign / Confirm / Delete per Domain

### Web Backoffice

| Domain | Presign | Confirm | Delete | Tipe Media |
|--------|---------|---------|--------|------------|
| Ads | `POST /web/ads/media/image/presign` | `POST /web/ads/media/image/confirm` | `DELETE /web/ads/media/image` | Image |
| Ads | `POST /web/ads/media/video/presign` | `POST /web/ads/media/video/confirm` | `DELETE /web/ads/media/video` | Video |
| Region | `POST /web/regions/media/image/presign` | `POST /web/regions/media/image/confirm` | `DELETE /web/regions/media/image` | Image |
| Announcement | `POST /web/announcements/media/image/presign` | `POST /web/announcements/media/image/confirm` | `DELETE /web/announcements/media/image` | Image |
| Community | `POST /web/communities/media/avatar/presign` | `POST /web/communities/media/avatar/confirm` | `DELETE /web/communities/media/avatar` | Avatar |
| Community | `POST /web/communities/media/post/presign` | `POST /web/communities/media/post/confirm` | `DELETE /web/communities/media/post` | Post Media |
| Event | `POST /web/events/media/image/presign` | `POST /web/events/media/image/confirm` | `DELETE /web/events/media/image` | Image |
| News | `POST /web/news/media/thumbnail/presign` | `POST /web/news/media/thumbnail/confirm` | `DELETE /web/news/media/thumbnail` | Thumbnail |
| User Management | `POST /web/users/media/avatar/presign` | `POST /web/users/media/avatar/confirm` | `DELETE /web/users/media/avatar` | Avatar |
| QnA | `POST /web/qna/media/attachment/presign` | `POST /web/qna/media/attachment/confirm` | `DELETE /web/qna/media/attachment` | Attachment |
| Directory | `POST /web/directory/media/image/presign` | `POST /web/directory/media/image/confirm` | `DELETE /web/directory/media/image` | Image |
| Subscription | `POST /web/subscriptions/media/proof/presign` | `POST /web/subscriptions/media/proof/confirm` | `DELETE /web/subscriptions/media/proof` | Proof |
| Bug Report | `POST /web/bug-reports/media/attachment/presign` | `POST /web/bug-reports/media/attachment/confirm` | `DELETE /web/bug-reports/media/attachment` | Attachment |

### Mobile

| Domain | Presign | Confirm | Delete | Tipe Media |
|--------|---------|---------|--------|------------|
| Profile | `POST /mobile/profile/avatar/presign` | `POST /mobile/profile/avatar/confirm` | `DELETE /mobile/profile/avatar` | Avatar |
| Community | `POST /mobile/communities/media/avatar/presign` | `POST /mobile/communities/media/avatar/confirm` | `DELETE /mobile/communities/media/avatar` | Avatar |
| QnA | `POST /mobile/qna/media/attachment/presign` | `POST /mobile/qna/media/attachment/confirm` | `DELETE /mobile/qna/media/attachment` | Attachment |
| Bug Report | `POST /mobile/bug-reports/media/attachment/presign` | `POST /mobile/bug-reports/media/attachment/confirm` | `DELETE /mobile/bug-reports/media/attachment` | Attachment |
| Subscription | `POST /mobile/subscriptions/media/proof/presign` | `POST /mobile/subscriptions/media/proof/confirm` | `DELETE /mobile/subscriptions/media/proof` | Proof |
| Event | `POST /mobile/events/media/image/presign` | `POST /mobile/events/media/image/confirm` | `DELETE /mobile/events/media/image` | Image |

---

## 6. Aturan Confirm / Delete

- **`s3:` keys** → tracked di tabel `media_uploads`. Confirm menandai record sebagai CONFIRMED, Delete menandai sebagai DELETED untuk cleanup async.
- **`ext:` keys** → **no-op**. Tidak ada entry di `media_uploads`. Confirm/Delete langsung return sukses tanpa efek samping.
- **Nilai kosong** → **no-op**.
- **Backward compat** → key lama tanpa prefix (dimulai `/`) diterima dan otomatis dikenali sebagai `s3:`.

---

## 7. Backward Compatibility

Data lama di DB yang belum memiliki prefix tetap bisa dibaca dan ditulis:

| Kondisi | Input | Perilaku |
|---------|-------|----------|
| Tanpa prefix, dimulai `/` | `/ads/images/old.jpg` | Dianggap `s3:` → simpan & baca seperti biasa |
| Full CDN URL | `https://cdn.example.com/ads/images/old.jpg` | API strip ke key → simpan sebagai `s3:` |
| Full URL eksternal | `https://partner.com/logo.png` | API simpan sebagai `ext:` (perilaku baru) |
| Sudah prefix | `s3:/ads/images/new.jpg` | Langsung pakai apa adanya |

Tidak ada migrasi DB yang dibutuhkan.

---

## 8. Content Type yang Diizinkan

| Tipe | MIME Types |
|------|------------|
| Gambar | `image/jpeg`, `image/png`, `image/webp`, `image/gif` |
| Video | `video/mp4`, `video/webm`, `video/quicktime` |

Validasi dilakukan di usecase presign, bukan di handler.

---

## 9. Alur Upload S3 (Lengkap)

```
1. UI → POST /[domain]/media/[type]/presign
         Body: { "content_type": "image/png" }
         Response: {
           "key": "s3:/ads/images/uuid.png",
           "presign_url": "https://minio.../signed-PUT-url",
           "public_url": "https://cdn.../ads/images/uuid.png",
           "expires_in": 3600
         }

2. UI → PUT <presign_url>
         Header: Content-Type: image/png
         Body: <binary file>
         (langsung ke MinIO, tanpa melalui API)

3. UI → POST /[domain]/create
         Body: { "image_url": "s3:/ads/images/uuid.png" }
         (key dari step 1 dipakai langsung — tanpa transformasi)
```

## 10. Alur URL Eksternal

```
1. UI punya URL eksternal: "https://partner.com/logo.png"
2. UI prefix manual: "ext:https://partner.com/logo.png"
3. UI → POST /[domain]/create
         Body: { "icon_url": "ext:https://partner.com/logo.png" }
```

---

## 11. Refactoring Frontend — Deteksi Tipe Media via `_raw` Field

### 11.1. Latar Belakang

Response DTO sekarang mengembalikan dua field per media: `xxx_url` (CDN untuk preview) dan `xxx_raw` (nilai asli dari DB dengan prefix `s3:` / `ext:`). FE harus menggunakan `xxx_raw` untuk menentukan mode komponen upload (upload mode vs paste URL mode).

### 11.2. Aturan Deteksi

| Nilai `xxx_raw` | Arti | Mode Komponen | Kirim ke PUT/UPDATE |
|-----------------|------|---------------|---------------------|
| Dimulai `s3:` | File S3 internal | **Upload mode** (drag & drop / file picker) | Kirim `xxx_raw` apa adanya |
| Dimulai `ext:` | URL eksternal | **URL mode** (text input) | Kirim `xxx_raw` apa adanya |
| Plain key (`/path`) | Data lama (backward compat) | **Upload mode** atau biarkan URL | Kirim `xxx_raw` apa adanya |
| `null` / `""` | Tidak ada media | — | Kirim `null` |

### 11.3. Yang Harus Diubah di FE

Untuk setiap komponen form yang memiliki field media dengan dua mekanisme (paste URL + upload via presign):

| # | Perubahan | Penjelasan |
|---|-----------|------------|
| 1 | **Gunakan `xxx_raw` sebagai `modelValue`** | Ganti binding dari `xxx_url` ke `xxx_raw` agar komponen upload bisa deteksi `s3:` vs `ext:` |
| 2 | **Pass `xxx_url` sebagai `preview-url` prop** | Untuk preview gambar saat `modelValue` adalah `s3:` key (tidak bisa langsung di-render sebagai `<img src>`) |
| 3 | **Kirim `xxx_raw` di payload PUT** | Saat media tidak berubah, kirim nilai `xxx_raw` (bukan `xxx_url`). Backend `NormalizeValue` tetap handle CDN URL, tapi `xxx_raw` lebih eksplisit |
| 4 | **Hapus panggilan `/confirm` setelah upload** | Confirm dilakukan otomatis oleh backend saat entity disimpan. FE cukup kirim `s3:` key di payload bisnis |
| 5 | **Hapus mode toggle manual** | Mode upload/URL ditentukan otomatis dari prefix `xxx_raw`, bukan dari toggle user |

### 11.4. Contoh Implementasi (Vue)

```vue
<MediaUploadField
  v-model="form.thumbnail_raw"           // ← pakai _raw, bukan _url
  :preview-url="form.thumbnail_url"       // ← CDN URL untuk preview
  domain="news"
  type="thumbnail"
  kind="image"
/>
```

Di dalam `MediaUploadField`:

```typescript
// Deteksi mode dari modelValue
onMounted(() => {
  if (props.modelValue?.startsWith('s3:')) {
    mode.value = 'upload'
  } else if (props.modelValue?.startsWith('ext:')) {
    mode.value = 'url'
  } else if (props.modelValue) {
    mode.value = 'url' // backward compat
  }
})

// Preview: fallback dari cache lokal, lalu dari props.previewUrl
const previewSrc = computed(() => {
  if (!value.value) return ''
  if (value.value.startsWith('s3:')) {
    return localCache[value.value] || props.previewUrl || ''
  }
  if (value.value.startsWith('ext:')) {
    return value.value.slice(4)
  }
  return value.value
})
```

### 11.5. Domain yang Perlu Refactoring

Berlaku untuk semua domain yang memiliki dua mekanisme (paste URL + upload presign):

| Domain | Field `_url` | Field `_raw` |
|--------|-------------|--------------|
| Ads | `image_url`, `video_url`, `thumbnail_url`, `icon_url`, `sponsor_logo_url` | `image_raw`, `video_raw`, `thumbnail_raw`, `icon_raw`, `sponsor_logo_raw` |
| Region | `image_url` | `image_raw` |
| Announcement | `image_url` | `image_raw` |
| Community | `avatar_url`, `url`, `thumb_url` | `avatar_raw`, `url_raw`, `thumb_url_raw` |
| Event | `cover_image`, `images[]`, `avatar` (organizer) | `cover_image_raw`, `images_raw`, `avatar_raw` |
| News | `thumbnail_url` | `thumbnail_raw` |
| User Management | `avatar_url` | `avatar_raw` |
| QnA | `attachment_urls[]`, `user_avatar`, `answered_by_avatar`, `asker_avatar` | `attachment_urls_raw`, `user_avatar_raw`, `answered_by_avatar_raw`, `asker_avatar_raw` |
| Directory | `logo_url`, `images[]`, `primary_image`, `image` | `logo_raw`, `images_raw`, `primary_image_raw`, `image_raw` |
| Subscription | `manual_proof_url` | `manual_proof_raw` |
| Bug Report | `attachments[]` | `attachments_raw` |
| Accounting | `attachment_url` | `attachment_raw` |
| Profile | `avatar`, `image_url` | `avatar_raw`, `image_raw` |

---

## 12. Catatan untuk Google Login

Setelah Google login, `avatar` di response session/me adalah URL Google langsung (`ext:`) — tidak di-upload ke S3 kita lagi. Response tetap mengembalikan full URL tanpa prefix:

```json
{
  "avatar": "https://lh3.googleusercontent.com/photo.jpg"
}
```
