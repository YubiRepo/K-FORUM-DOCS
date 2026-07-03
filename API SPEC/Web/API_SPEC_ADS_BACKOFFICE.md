# API Spec — Ads Module (Web Backoffice)

Dokumentasi API untuk superadmin mengelola ads, moderasi, ad settings, dan melihat analytics di dashboard backoffice.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/ads`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin only)

---

## Daftar Endpoint

| Method | URL | Deskripsi |
|--------|-----|-----------|
| GET | `/ads` | List semua ads |
| POST | `/ads` | Buat ads baru |
| GET | `/ads/:id` | Detail satu ads |
| PUT | `/ads/:id` | Update ads |
| DELETE | `/ads/:id` | Hapus ads |
| PATCH | `/ads/:id/status` | Update status ads (approve/reject/pause/resume) |
| GET | `/ads/:id/analytics` | Analytics detail satu ads |
| GET | `/analytics` | Analytics aggregate semua ads |
| GET | `/settings` | Ambil ad settings |
| PUT | `/settings` | Update ad settings |
| POST | `/media/image/presign` | Minta presigned URL upload gambar |
| POST | `/media/image/confirm` | Konfirmasi upload gambar |
| DELETE | `/media/image` | Tandai gambar sebagai deleted |
| POST | `/media/video/presign` | Minta presigned URL upload video |
| POST | `/media/video/confirm` | Konfirmasi upload video |
| DELETE | `/media/video` | Tandai video sebagai deleted |

---

## 1. GET /ads

List semua ads dengan filter dan pagination.

**Method**: GET  
**URL**: `/api/v1/web/ads`

**Query Parameters**:
- `status` (optional): `draft` | `pending` | `active` | `rejected` | `paused` | `expired`
- `ad_type` (optional): `image_banner` | `video_banner` | `text_ad` | `native_ad`
- `created_by` (optional): filter by user UUID (untuk lihat ads dari member tertentu)
- `search` (optional): search by title
- `date_from` (optional): filter `start_date` >= date (format: `YYYY-MM-DD`)
- `date_to` (optional): filter `end_date` <= date (format: `YYYY-MM-DD`)
- `limit` (optional, default: 20, max: 100)
- `offset` (optional, default: 0)
- `sort` (optional, default: `created_at_desc`): `created_at_desc` | `created_at_asc` | `start_date_asc` | `end_date_asc`

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "ad_001",
      "title": "Promo Tiket Lebaran 2026",
      "ad_type": "image_banner",
      "status": "active",
      "click_url": "https://kai.id/promo/lebaran-2026",
      "start_date": "2026-06-01",
      "end_date": "2026-06-30",
      "total_impressions": 12400,
      "total_clicks": 310,
      "ctr": 2.5,
      "created_by": {
        "id": "user_superadmin_001",
        "name": "KAI Pusat",
        "role": "superadmin"
      },
      "created_at": "2026-05-28T09:00:00.000Z",
      "updated_at": "2026-06-01T00:00:00.000Z"
    },
    {
      "id": "ad_002",
      "title": "Diskon Hotel Dekat Stasiun",
      "ad_type": "native_ad",
      "status": "pending",
      "click_url": "https://hotelsentral.id/promo",
      "start_date": "2026-06-10",
      "end_date": "2026-06-30",
      "total_impressions": 0,
      "total_clicks": 0,
      "ctr": 0,
      "created_by": {
        "id": "user_member_pro_001",
        "name": "Budi Santoso",
        "role": "member"
      },
      "created_at": "2026-06-05T14:00:00.000Z",
      "updated_at": "2026-06-05T14:00:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 34
  }
}
```

---

## 2. POST /ads

Buat ads baru. Superadmin langsung `active` tanpa perlu approval.

**Method**: POST  
**URL**: `/api/v1/web/ads`

**Request Body (image_banner)**:
```json
{
  "title": "Promo Tiket Lebaran 2026",
  "ad_type": "image_banner",
  "image_url": "s3:/ads/images/lebaran-2026.jpg",
  "click_url": "https://kai.id/promo/lebaran-2026",
  "start_date": "2026-06-01",
  "end_date": "2026-06-30",
  "notes": "Campaign lebaran Q2 2026"
}
```

**Request Body (video_banner)**:
```json
{
  "title": "Video Promo KAI Wisata",
  "ad_type": "video_banner",
  "video_url": "s3:/ads/videos/wisata-2026.mp4",
  "thumbnail_url": "s3:/ads/images/wisata-2026-thumb.jpg",
  "click_url": "https://kai.id/wisata",
  "start_date": "2026-06-01",
  "end_date": "2026-06-30",
  "notes": null
}
```

**Request Body (text_ad)**:
```json
{
  "title": "Text Ad Promo Member",
  "ad_type": "text_ad",
  "headline": "Gratis ongkir semua destinasi!",
  "body_text": "Pesan tiket kereta sekarang dan nikmati promo gratis ongkir untuk semua rute.",
  "cta_label": "Pesan Sekarang",
  "icon_url": "s3:/ads/images/icon-train.png",
  "click_url": "https://kai.id/pesan-tiket",
  "start_date": "2026-06-01",
  "end_date": "2026-06-20",
  "notes": null
}
```

**Request Body (native_ad)**:
```json
{
  "title": "Native Ad Hotel Sentral",
  "ad_type": "native_ad",
  "body_text": "Hotel terbaik dekat Stasiun Gambir. Tarif mulai Rp 350rb/malam, free breakfast.",
  "image_url": "ext:https://cdn.hotelsentral.id/banner.jpg",
  "sponsor_name": "Hotel Sentral Jakarta",
  "sponsor_logo_url": "ext:https://cdn.hotelsentral.id/logo.png",
  "click_url": "https://hotelsentral.id/promo-kai",
  "start_date": "2026-06-10",
  "end_date": "2026-06-30",
  "notes": "Partnership Q2"
}
```

**Response (201 Created)**:
```json
{
  "data": {
    "id": "ad_003",
    "title": "Promo Tiket Lebaran 2026",
    "ad_type": "image_banner",
    "status": "active",
    "click_url": "https://kai.id/promo/lebaran-2026",
    "start_date": "2026-06-01",
    "end_date": "2026-06-30",
    "created_by": "user_superadmin_001",
    "created_at": "2026-06-01T08:00:00.000Z"
  },
  "message": "Ads berhasil dibuat dan langsung aktif"
}
```

**Errors**:
```json
// 422 — field wajib tidak lengkap sesuai ad_type
{
  "error": "VALIDATION_ERROR",
  "message": "Field image_url wajib diisi untuk tipe image_banner",
  "field": "image_url"
}

// 422 — end_date terlalu jauh dari start_date
{
  "error": "VALIDATION_ERROR",
  "message": "Durasi iklan maksimum 30 hari",
  "field": "end_date"
}
```

---

## 3. GET /ads/:id

Detail satu ads beserta semua field konten.

**Method**: GET  
**URL**: `/api/v1/web/ads/:id`

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ad_002",
    "title": "Diskon Hotel Dekat Stasiun",
    "ad_type": "native_ad",
    "body_text": "Hotel terbaik dekat Stasiun Gambir. Tarif mulai Rp 350rb/malam, free breakfast.",
    "image_url": "https://cdn.hotelsentral.id/banner.jpg",
    "sponsor_name": "Hotel Sentral Jakarta",
    "sponsor_logo_url": "https://cdn.hotelsentral.id/logo.png",
    "click_url": "https://hotelsentral.id/promo-kai",
    "start_date": "2026-06-10",
    "end_date": "2026-06-30",
    "notes": null,
    "status": "pending",
    "reject_reason": null,
    "reviewed_by": null,
    "reviewed_at": null,
    "total_impressions": 0,
    "total_clicks": 0,
    "ctr": 0,
    "created_by": {
      "id": "user_member_pro_001",
      "name": "Budi Santoso",
      "role": "member",
      "avatar_url": "https://cdn.kai.id/avatars/budi.jpg"
    },
    "created_at": "2026-06-05T14:00:00.000Z",
    "updated_at": "2026-06-05T14:00:00.000Z"
  }
}
```

**Errors**:
```json
// 404
{
  "error": "NOT_FOUND",
  "message": "Ads tidak ditemukan"
}
```

---

## 4. PUT /ads/:id

Update konten ads. Hanya bisa update jika status `draft` atau `pending`.  
Jika `active` atau `paused`, superadmin harus pause dulu via PATCH `/status`.

**Method**: PUT  
**URL**: `/api/v1/web/ads/:id`

**Request Body** (kirim hanya field yang ingin diubah):
```json
{
  "title": "Promo Tiket Lebaran 2026 — Update",
  "image_url": "s3:/ads/images/lebaran-2026-v2.jpg",
  "click_url": "https://kai.id/promo/lebaran-2026-v2",
  "end_date": "2026-07-05",
  "notes": "Diperpanjang 5 hari"
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ad_001",
    "title": "Promo Tiket Lebaran 2026 — Update",
    "status": "active",
    "updated_at": "2026-06-10T11:00:00.000Z"
  },
  "message": "Ads berhasil diupdate"
}
```

**Errors**:
```json
// 422 — ads sedang active, tidak bisa edit langsung
{
  "error": "INVALID_STATE",
  "message": "Ads yang sedang active tidak bisa diedit. Pause terlebih dahulu.",
  "current_status": "active"
}
```

---

## 5. DELETE /ads/:id

Hapus ads permanen. Bisa dilakukan di status apapun.

**Method**: DELETE  
**URL**: `/api/v1/web/ads/:id`

**Response (200 OK)**:
```json
{
  "message": "Ads berhasil dihapus"
}
```

---

## 6. PATCH /ads/:id/status

Update status ads — untuk approve, reject, pause, atau resume.

**Method**: PATCH  
**URL**: `/api/v1/web/ads/:id/status`

**Request Body (approve)**:
```json
{
  "status": "active"
}
```

**Request Body (reject — reject_reason wajib)**:
```json
{
  "status": "rejected",
  "reject_reason": "Konten iklan mengandung klaim yang tidak dapat diverifikasi. Mohon revisi dan ajukan ulang."
}
```

**Request Body (pause)**:
```json
{
  "status": "paused"
}
```

**Request Body (resume)**:
```json
{
  "status": "active"
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ad_002",
    "status": "active",
    "reviewed_by": "user_superadmin_001",
    "reviewed_at": "2026-06-06T09:00:00.000Z"
  },
  "message": "Status ads berhasil diupdate"
}
```

**Transisi status yang valid**:
| Dari | Ke | Keterangan |
|------|----|------------|
| `pending` | `active` | Approve |
| `pending` | `rejected` | Reject — wajib isi `reject_reason` |
| `active` | `paused` | Pause sementara |
| `paused` | `active` | Resume |
| `draft` | `pending` | Submit untuk review (biasanya dari member, tapi superadmin bisa juga) |

**Errors**:
```json
// 422 — transisi status tidak valid
{
  "error": "INVALID_TRANSITION",
  "message": "Tidak bisa mengubah status dari 'expired' ke 'active'",
  "current_status": "expired"
}

// 422 — reject tanpa reject_reason
{
  "error": "VALIDATION_ERROR",
  "message": "reject_reason wajib diisi saat menolak ads",
  "field": "reject_reason"
}
```

---

## 7. GET /ads/:id/analytics

Analytics detail satu ads — impressi, klik, CTR.

**Method**: GET  
**URL**: `/api/v1/web/ads/:id/analytics`

**Query Parameters**:
- `date_from` (optional, default: `start_date` ads): `YYYY-MM-DD`
- `date_to` (optional, default: hari ini): `YYYY-MM-DD`
- `group_by` (optional, default: `day`): `day` | `week`

**Response (200 OK)**:
```json
{
  "data": {
    "ad_id": "ad_001",
    "ad_title": "Promo Tiket Lebaran 2026",
    "ad_type": "image_banner",
    "summary": {
      "total_impressions": 12400,
      "total_clicks": 310,
      "ctr": 2.5
    },
    "chart": [
      {
        "date": "2026-06-01",
        "impressions": 1800,
        "clicks": 45,
        "ctr": 2.5
      },
      {
        "date": "2026-06-02",
        "impressions": 2100,
        "clicks": 58,
        "ctr": 2.76
      },
      {
        "date": "2026-06-03",
        "impressions": 1950,
        "clicks": 47,
        "ctr": 2.41
      }
    ],
    "date_from": "2026-06-01",
    "date_to": "2026-06-03"
  }
}
```

---

## 8. GET /analytics

Analytics aggregate semua ads di platform — untuk dashboard overview superadmin.

**Method**: GET  
**URL**: `/api/v1/web/ads/analytics`

**Query Parameters**:
- `date_from` (optional, default: 30 hari lalu): `YYYY-MM-DD`
- `date_to` (optional, default: hari ini): `YYYY-MM-DD`

**Response (200 OK)**:
```json
{
  "data": {
    "summary": {
      "total_ads_active": 12,
      "total_ads_pending": 3,
      "total_impressions": 98400,
      "total_clicks": 2460,
      "ctr": 2.5
    },
    "by_type": [
      { "ad_type": "image_banner", "count": 5, "impressions": 52000, "clicks": 1400, "ctr": 2.69 },
      { "ad_type": "video_banner", "count": 2, "impressions": 18000, "clicks": 540, "ctr": 3.0 },
      { "ad_type": "text_ad",      "count": 3, "impressions": 14400, "clicks": 288, "ctr": 2.0 },
      { "ad_type": "native_ad",    "count": 2, "impressions": 14000, "clicks": 232, "ctr": 1.66 }
    ],
    "top_ads": [
      {
        "id": "ad_001",
        "title": "Promo Tiket Lebaran 2026",
        "ad_type": "image_banner",
        "impressions": 12400,
        "clicks": 310,
        "ctr": 2.5
      }
    ],
    "date_from": "2026-05-11",
    "date_to": "2026-06-10"
  }
}
```

---

## 9. GET /settings

Ambil konfigurasi ad settings saat ini.

**Method**: GET  
**URL**: `/api/v1/web/ads/settings`

**Response (200 OK)**:
```json
{
  "data": {
    "id": "adsetting_001",
    "approval_mode": "require_review",
    "max_active_ads_per_member": 3,
    "max_duration_days": 30,
    "feed_ads_interval": 5,
    "slider_max_items": 5,
    "updated_at": "2026-06-01T08:00:00.000Z",
    "updated_by": {
      "id": "user_superadmin_001",
      "name": "KAI Pusat"
    }
  }
}
```

---

## 10. PUT /settings

Update konfigurasi ad settings. Kirim hanya field yang ingin diubah.

**Method**: PUT  
**URL**: `/api/v1/web/ads/settings`

**Request Body**:
```json
{
  "approval_mode": "auto_publish",
  "max_active_ads_per_member": 5,
  "max_duration_days": 30,
  "feed_ads_interval": 7,
  "slider_max_items": 5
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "approval_mode": "auto_publish",
    "max_active_ads_per_member": 5,
    "max_duration_days": 30,
    "feed_ads_interval": 7,
    "slider_max_items": 5,
    "updated_at": "2026-06-10T10:00:00.000Z",
    "updated_by": "user_superadmin_001"
  },
  "message": "Ad settings berhasil diupdate. Berlaku untuk ads baru."
}
```

**Errors**:
```json
// 422
{
  "error": "VALIDATION_ERROR",
  "message": "max_active_ads_per_member minimal 1",
  "field": "max_active_ads_per_member"
}
```

---

## Error Codes Umum

| HTTP Status | Error Code | Keterangan |
|-------------|------------|------------|
| 400 | `BAD_REQUEST` | Request malformed |
| 401 | `UNAUTHORIZED` | Token tidak valid atau expired |
| 403 | `FORBIDDEN` | Bukan superadmin |
| 404 | `NOT_FOUND` | Ads tidak ditemukan |
| 422 | `VALIDATION_ERROR` | Field tidak valid |
| 422 | `INVALID_STATE` | Operasi tidak valid di status saat ini |
| 422 | `INVALID_TRANSITION` | Transisi status tidak diizinkan |
| 500 | `INTERNAL_ERROR` | Server error |
