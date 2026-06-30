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

API selalu mengembalikan **full URL tanpa prefix**:

```json
{
  "image_url": "https://cdn.example.com/ads/images/uuid.jpg"
}
```

Untuk URL eksternal:
```json
{
  "avatar": "https://lh3.googleusercontent.com/photo.jpg"
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

### GET Response (tidak berubah)
```json
{ "image_url": "https://cdn.example.com/ads/images/uuid.jpg" }
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

## 11. Catatan untuk Google Login

Setelah Google login, `avatar` di response session/me adalah URL Google langsung (`ext:`) — tidak di-upload ke S3 kita lagi. Response tetap mengembalikan full URL tanpa prefix:

```json
{
  "avatar": "https://lh3.googleusercontent.com/photo.jpg"
}
```
