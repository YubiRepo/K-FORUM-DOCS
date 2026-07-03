# Media Upload Rules — Backend Reference

Dokumen ini adalah acuan untuk implementasi upload/update/delete media di semua domain backend (BE). Pola yang didokumentasikan di sini adalah apa yang sudah diimplementasikan di modul **News Article** dan harus diikuti oleh semua domain lain.

---

## 1. Skema Prefix

Semua nilai media di DB disimpan dengan **prefix** untuk membedakan sumber:

| Prefix | Arti | Contoh |
|--------|------|--------|
| `s3:` | Objek di S3/MinIO internal | `s3:/ads/images/uuid.jpg` |
| `ext:` | URL eksternal (bukan S3 kita) | `ext:https://partner.com/logo.png` |
| *(tanpa prefix)* | Data lama (backward compat) | `/ads/images/old.jpg` |

Tabel `media_uploads` hanya melacak file `s3:` (lifecycle: PENDING → CONFIRMED → DELETED). File `ext:` tidak pernah masuk `media_uploads`.

---

## 2. Upload Flow

### 2.1. Presign Endpoint

Backend menyediakan endpoint `POST /[domain]/media/[type]/presign`:

```
Request:  { "content_type": "image/jpeg" }
Response: {
  "presign_url": "https://minio.../signed-PUT-url",
  "key": "s3:/path/uuid.jpg",
  "public_url": "https://cdn.../path/uuid.jpg",
  "expires_in": 900
}
```

Aturan implementasi:
- Validasi `content_type` sesuai tipe media (gambar: jpeg/png/webp/gif, video: mp4/webm/quicktime)
- Generate object path menggunakan constant `ObjectPath` yang sudah ditentukan
- Panggil `mediaUploadSvc.RequestUpload()`:
  - Generate presigned PUT URL
  - Simpan record di `media_uploads` dengan status **PENDING**
- Return key dengan prefix `s3:`

### 2.2. Client Upload (tidak melalui backend)

Client melakukan PUT langsung ke `presign_url` dengan file binary dan `Content-Type` yang sesuai. Backend tidak terlibat.

### 2.3. Confirm via Payload (WAJIB — bukan via endpoint)

Ketika entitas dibuat/diupdate, client mengirim `s3:` key langsung di payload bisnis:

```json
{
  "thumbnail_url": "s3:/news/thumbnails/uuid.jpg"
}
```

Backend akan auto-confirm di usecase setelah entity tersimpan:

```go
// 1. Normalize (handle s3:/ext:/backward compat)
thumbnailURL := normalizeKeyPtr(svc, req.ThumbnailURL)

// 2. Simpan ke DB
translation.ThumbnailURL = thumbnailURL
repo.Save(ctx, translation)

// 3. Confirm ke media_uploads (PENDING → CONFIRMED)
confirmMediaKey(ctx, svc, translation.ThumbnailURL)
```

> **WAJIB:** Confirm **hanya** dilakukan via payload bisnis. UI **tidak boleh** memanggil endpoint confirm terpisah. Endpoint `/confirm` tetap ada untuk backward compat dan debugging, tapi bukan flow utama.

### 2.4. Confirm Eksplisit (fallback — tidak untuk flow utama)

Endpoint `POST /[domain]/media/[type]/confirm?key=` tetap ada untuk:
- Debugging / testing
- Backward compat dengan client lama
- Flow manual oleh admin

Tapi **jangan** digunakan di flow normal create/update.

---

## 3. Update Flow

### 3.1. Thumbnail/Media Diganti

```
1. Fetch old entity / old translation dari DB
2. Normalize old value → oldNorm
3. Normalize new value dari request → newNorm
4. Bandingkan oldNorm vs newNorm
5. Jika berbeda:
   a. MarkDeleted(oldNorm)  → CONFIRMED → DELETED (cleanup S3)
   b. Simpan entity baru
   c. ConfirmUpload(newNorm) → PENDING → CONFIRMED
6. Jika sama:
   a. Skip (no-op), tetap simpan entity
```

```go
oldNorm := svc.NormalizeValue(*oldValue)
newNorm := svc.NormalizeValue(*newValue)

if oldNorm != "" && oldNorm != newNorm {
    _ = svc.MarkDeleted(ctx, oldNorm)
}

repo.Save(ctx, entity)
confirmMediaKey(ctx, svc, newThumbnail)
```

### 3.2. Media Dihapus (nil / "")

```
1. oldNorm != "" && newNorm == ""
2. MarkDeleted(oldNorm)
3. Simpan entity dengan field media = nil
```

```go
if oldNorm != "" && newNorm == "" {
    _ = svc.MarkDeleted(ctx, oldNorm)
}
```

### 3.3. Media Tidak Berubah

```
1. oldNorm == newNorm
2. Skip MarkDeleted dan ConfirmUpload (no-op)
3. Tetap simpan entity
```

---

## 4. Delete Entity Flow

Ketika entity dihapus (hard delete), semua media terkait harus di-mark deleted:

```go
func (uc *DeleteEntityUseCase) Execute(ctx context.Context, id string) error {
    // 1. Validasi bisa dihapus
    entity, err := uc.repo.FindByID(ctx, id)
    
    // 2. Mark deleted semua media sebelum hapus
    for _, media := range entity.Media {
        markMediaDeleted(ctx, uc.mediaUploadSvc, media.URL)
    }
    
    // 3. Hapus entity (CASCADE ke child tables)
    uc.repo.Delete(ctx, id)
}
```

---

## 5. DTO Response Fields

Setiap response DTO yang mengandung field media WAJIB mengembalikan dua field:

```json
{
  "thumbnail_url": "https://cdn.../uuid.jpg",
  "thumbnail_raw": "s3:/news/thumbnails/uuid.jpg"
}
```

| Field | Format | Kegunaan |
|-------|--------|----------|
| `thumbnail_url` | Full CDN URL | Preview gambar di UI |
| `thumbnail_raw` | Nilai dari DB (`s3:` / `ext:` / plain key) | UI kirim balik field ini saat PUT jika tidak ada perubahan |

Aturan:
- Untuk `s3:` type, `thumbnail_raw` berisi `s3:/path/file.jpg`
- Untuk `ext:` type, `thumbnail_raw` berisi `ext:https://external.com/img.jpg`
- Kedua field selalu diisi (tidak conditional)

---

## 6. Aturan Wajib Implementasi

| # | Aturan | Konsekuensi Jika Dilanggar |
|---|--------|---------------------------|
| 1 | Jangan panggil `/confirm` dari UI. Confirm hanya via payload bisnis | File PENDING jadi orphan jika create batal |
| 2 | Setiap create/update WAJIB bandingkan old vs new media key | File S3 jadi orphan (tidak pernah di-mark deleted) |
| 3 | Setiap delete entity WAJIB mark deleted semua media | File S3 jadi orphan selamanya |
| 4 | Setiap response DTO WAJIB return `thumbnail_raw` + `thumbnail_url` | UI tidak bisa deteksi s3: vs ext: untuk render mode upload/url |
| 5 | Gunakan `svc.NormalizeValue()` untuk semua input media (handle `s3:` / `ext:` / backward compat) | Input dari UI bisa corrupt |
| 6 | Gunakan `svc.PublicURL()` untuk semua output media | Client dapat raw key bukan CDN URL |
| 7 | Untuk `ext:` type, pastikan tidak masuk `media_uploads` | Udah di-handle oleh service (ConfirmUpload/MarkDeleted = no-op untuk ext:) |

---

## 7. Pola Kode yang Harus Diikuti

### Helper Functions (taruh di `helpers.go` masing-masing usecase)

```go
func normalizeKeyPtr(svc port.MediaUploadService, raw *string) *string {
    if svc == nil || raw == nil {
        return raw
    }
    key := svc.NormalizeValue(*raw)
    if key == "" {
        return nil
    }
    return &key
}

func confirmMediaKey(ctx context.Context, svc port.MediaUploadService, key *string) {
    if svc == nil || key == nil {
        return
    }
    normalized := svc.NormalizeValue(*key)
    if normalized != "" {
        _ = svc.ConfirmUpload(ctx, normalized)
    }
}

func markMediaDeleted(ctx context.Context, svc port.MediaUploadService, key *string) {
    if svc == nil || key == nil {
        return
    }
    normalized := svc.NormalizeValue(*key)
    if normalized != "" {
        _ = svc.MarkDeleted(ctx, normalized)
    }
}

func resolveMediaURL(svc port.MediaUploadService, key *string) *string {
    if svc == nil || key == nil {
        return key
    }
    v := svc.PublicURL(*key)
    return &v
}
```

### Usecase Pattern — Create

```go
func (uc *CreateUseCase) Execute(ctx context.Context, req dto.CreateRequest) (*dto.Response, error) {
    // 1. Normalize input media
    thumbnail := normalizeKeyPtr(uc.mediaUploadSvc, req.ThumbnailURL)

    // 2. Simpan entity
    entity := NewEntity(thumbnail)
    uc.repo.Save(ctx, entity)

    // 3. Confirm media
    confirmMediaKey(ctx, uc.mediaUploadSvc, thumbnail)

    return &dto.Response{ID: entity.ID}, nil
}
```

### Usecase Pattern — Update

```go
func (uc *UpdateUseCase) Execute(ctx context.Context, id string, req dto.UpdateRequest) (*dto.Response, error) {
    // 1. Ambil old entity
    old, _ := uc.repo.FindByID(ctx, id)
    var oldThumbnail string
    if old.ThumbnailURL != nil {
        oldThumbnail = *old.ThumbnailURL
    }

    // 2. Normalize new input
    newThumbnail := normalizeKeyPtr(uc.mediaUploadSvc, req.ThumbnailURL)

    // 3. Bandingkan dan cleanup jika berubah
    oldNorm := uc.mediaUploadSvc.NormalizeValue(oldThumbnail)
    newNorm := ""
    if newThumbnail != nil {
        newNorm = uc.mediaUploadSvc.NormalizeValue(*newThumbnail)
    }

    if oldNorm != "" && oldNorm != newNorm {
        _ = uc.mediaUploadSvc.MarkDeleted(ctx, oldNorm)
    }

    // 4. Update entity
    entity := old
    entity.ThumbnailURL = newThumbnail
    uc.repo.Update(ctx, entity)

    // 5. Confirm media baru (jika ada)
    confirmMediaKey(ctx, uc.mediaUploadSvc, newThumbnail)
}
```

### Usecase Pattern — Delete

```go
func (uc *DeleteUseCase) Execute(ctx context.Context, id string) error {
    old, _ := uc.repo.FindByID(ctx, id)
    
    // Mark deleted semua media
    markMediaDeleted(ctx, uc.mediaUploadSvc, old.ThumbnailURL)
    
    // Hapus entity
    uc.repo.Delete(ctx, id)
}
```

---

## 8. Ringkasan

| Flow | Action | MediaUploadSvc Method | media_uploads Status |
|------|--------|----------------------|---------------------|
| Presign | Minta upload URL | `RequestUpload` | PENDING |
| Upload ke S3 | Client langsung PUT | — | PENDING |
| Create entity | Simpan + confirm | `ConfirmUpload` | PENDING → **CONFIRMED** |
| Update entity (ganti media) | MarkDeleted old + Confirm new | `MarkDeleted` + `ConfirmUpload` | CONFIRMED → **DELETED**, new → **CONFIRMED** |
| Update entity (hapus media) | MarkDeleted old | `MarkDeleted` | CONFIRMED → **DELETED** |
| Delete entity | MarkDeleted semua | `MarkDeleted` | CONFIRMED → **DELETED** |
| Confirm endpoint (fallback) | Confirm manual | `ConfirmUpload` | PENDING → **CONFIRMED** |
| Delete endpoint (fallback) | Delete manual | `MarkDeleted` | CONFIRMED → **DELETED** |
