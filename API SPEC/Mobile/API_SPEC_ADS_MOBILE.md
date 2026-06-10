# API Spec — Ads Module (Mobile Client)

Dokumentasi API untuk mobile app menampilkan ads di home screen, halaman Iklan & Promo, dan member mengelola ads milik sendiri.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/ads`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (keterangan per endpoint)

---

## Daftar Endpoint

| Method | URL | Auth | Deskripsi |
|--------|-----|------|-----------|
| GET | `/home` | Optional | Ambil ads untuk home screen (slider + feed) |
| GET | `/` | Optional | List semua ads aktif untuk halaman Iklan & Promo |
| POST | `/impression` | Optional | Catat impressi ads |
| POST | `/click` | Optional | Catat klik ads |
| GET | `/my` | Required | List ads milik member sendiri |
| POST | `/` | Required | Buat ads baru (member) |
| GET | `/:id` | Required | Detail ads milik sendiri |
| PUT | `/:id` | Required | Update ads milik sendiri |
| DELETE | `/:id` | Required | Hapus ads milik sendiri |
| PATCH | `/:id/status` | Required | Pause / resume ads milik sendiri |
| GET | `/:id/analytics` | Required | Analytics ads milik sendiri |

---

## 1. GET /home

Ambil data ads khusus untuk home screen — slider banner dan slot native feed.  
Dipanggil saat home screen load atau refresh.

**Method**: GET  
**URL**: `/api/v1/mobile/ads/home`  
**Authentication**: Optional

**Response (200 OK)**:
```json
{
  "data": {
    "slider": [
      {
        "id": "ad_001",
        "ad_type": "image_banner",
        "image_url": "https://cdn.kai.id/ads/lebaran-2026.jpg",
        "click_url": "https://kai.id/promo/lebaran-2026",
        "is_kai_official": true
      },
      {
        "id": "ad_007",
        "ad_type": "video_banner",
        "video_url": "https://cdn.kai.id/ads/wisata-2026.mp4",
        "thumbnail_url": "https://cdn.kai.id/ads/wisata-2026-thumb.jpg",
        "click_url": "https://kai.id/wisata",
        "is_kai_official": true
      },
      {
        "id": "ad_003",
        "ad_type": "image_banner",
        "image_url": "https://cdn.hotelsentral.id/banner-kai.jpg",
        "click_url": "https://hotelsentral.id/promo-kai",
        "is_kai_official": false
      }
    ],
    "feed_ads": [
      {
        "id": "ad_005",
        "ad_type": "native_ad",
        "body_text": "Hotel terbaik dekat Stasiun Gambir. Tarif mulai Rp 350rb/malam.",
        "image_url": "https://cdn.hotelsentral.id/thumb.jpg",
        "sponsor_name": "Hotel Sentral Jakarta",
        "sponsor_logo_url": "https://cdn.hotelsentral.id/logo.png",
        "click_url": "https://hotelsentral.id/promo-kai",
        "is_kai_official": false
      },
      {
        "id": "ad_006",
        "ad_type": "text_ad",
        "headline": "Gratis ongkir semua destinasi!",
        "body_text": "Pesan tiket kereta sekarang dan nikmati promo gratis ongkir.",
        "cta_label": "Pesan Sekarang",
        "icon_url": "https://cdn.kai.id/ads/icon-train.png",
        "click_url": "https://kai.id/pesan-tiket",
        "is_kai_official": true
      }
    ],
    "feed_interval": 5
  }
}
```

> `is_kai_official: true` → ads dari superadmin/KAI Pusat. Mobile bisa tampilkan label berbeda jika diperlukan.  
> `feed_interval` → setiap berapa item feed, sisipkan 1 item dari `feed_ads`. Diambil dari ad settings.  
> Jika tidak ada slider ads aktif, `slider: []` — mobile hide section slider.  
> Jika tidak ada feed ads aktif, `feed_ads: []` — mobile skip slot native di feed.

---

## 2. GET /

List semua ads aktif untuk halaman "Iklan & Promo".  
Termasuk slider banner dan list lengkap semua tipe.

**Method**: GET  
**URL**: `/api/v1/mobile/ads`  
**Authentication**: Optional

**Query Parameters**:
- `ad_type` (optional): `image_banner` | `video_banner` | `text_ad` | `native_ad`
- `limit` (optional, default: 20, max: 50)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": {
    "slider": [
      {
        "id": "ad_001",
        "ad_type": "image_banner",
        "image_url": "https://cdn.kai.id/ads/lebaran-2026.jpg",
        "click_url": "https://kai.id/promo/lebaran-2026",
        "is_kai_official": true
      }
    ],
    "ads": [
      {
        "id": "ad_001",
        "ad_type": "image_banner",
        "title_display": "Promo Tiket Lebaran 2026",
        "image_url": "https://cdn.kai.id/ads/lebaran-2026.jpg",
        "click_url": "https://kai.id/promo/lebaran-2026",
        "sponsor_name": "KAI Official",
        "end_date": "2026-06-30",
        "cta_label": "Lihat Promo",
        "is_kai_official": true
      },
      {
        "id": "ad_005",
        "ad_type": "native_ad",
        "title_display": "Hotel dekat Stasiun Gambir",
        "image_url": "https://cdn.hotelsentral.id/thumb.jpg",
        "body_text": "Tarif mulai Rp 350rb/malam, free breakfast.",
        "click_url": "https://hotelsentral.id/promo-kai",
        "sponsor_name": "Hotel Sentral Jakarta",
        "sponsor_logo_url": "https://cdn.hotelsentral.id/logo.png",
        "end_date": "2026-06-30",
        "cta_label": "Pesan Sekarang",
        "is_kai_official": false
      },
      {
        "id": "ad_006",
        "ad_type": "text_ad",
        "title_display": "Gratis ongkir semua destinasi!",
        "headline": "Gratis ongkir semua destinasi!",
        "body_text": "Pesan tiket kereta sekarang dan nikmati promo gratis ongkir.",
        "cta_label": "Pesan Sekarang",
        "icon_url": "https://cdn.kai.id/ads/icon-train.png",
        "click_url": "https://kai.id/pesan-tiket",
        "sponsor_name": "KAI Official",
        "end_date": "2026-06-20",
        "is_kai_official": true
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 8
    }
  }
}
```

> `title_display` — field display-ready yang diisi server: untuk `image_banner` & `native_ad` pakai `title`, untuk `text_ad` pakai `headline`, untuk `video_banner` pakai `title`. Mobile tidak perlu logika ini.

---

## 3. POST /impression

Catat bahwa user melihat satu ads (impressi). Dipanggil mobile saat ads tampil di viewport.

**Method**: POST  
**URL**: `/api/v1/mobile/ads/impression`  
**Authentication**: Optional

**Request Body**:
```json
{
  "ad_id": "ad_001",
  "session_id": "sess_abc123xyz"
}
```

> `session_id` — string unik yang di-generate mobile saat app launch, dipakai untuk dedup impressi dalam satu sesi. Satu `session_id` + `ad_id` hanya dihitung 1 impressi.

**Response (200 OK)**:
```json
{
  "message": "ok"
}
```

> Response selalu 200 — mobile tidak perlu handle error analytics. Jika duplikat, server ignore silently.

---

## 4. POST /click

Catat bahwa user tap/klik satu ads. Dipanggil mobile sesaat sebelum membuka `click_url`.

**Method**: POST  
**URL**: `/api/v1/mobile/ads/click`  
**Authentication**: Optional

**Request Body**:
```json
{
  "ad_id": "ad_001",
  "session_id": "sess_abc123xyz"
}
```

**Response (200 OK)**:
```json
{
  "message": "ok"
}
```

---

## 5. GET /my

List semua ads milik member sendiri — semua status.

**Method**: GET  
**URL**: `/api/v1/mobile/ads/my`  
**Authentication**: Required

**Query Parameters**:
- `status` (optional): `draft` | `pending` | `active` | `rejected` | `paused` | `expired`
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "ad_010",
      "title": "Promo Toko Oleh-oleh Saya",
      "ad_type": "native_ad",
      "status": "active",
      "start_date": "2026-06-05",
      "end_date": "2026-06-30",
      "total_impressions": 840,
      "total_clicks": 22,
      "ctr": 2.62,
      "created_at": "2026-06-04T10:00:00.000Z"
    },
    {
      "id": "ad_011",
      "title": "Flash Sale Akhir Bulan",
      "ad_type": "text_ad",
      "status": "rejected",
      "reject_reason": "Klaim diskon tidak dapat diverifikasi. Mohon lampirkan bukti promo.",
      "start_date": "2026-06-15",
      "end_date": "2026-06-30",
      "total_impressions": 0,
      "total_clicks": 0,
      "ctr": 0,
      "created_at": "2026-06-08T14:00:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 2
  }
}
```

---

## 6. POST /

Buat ads baru dari member. Status awal tergantung `approval_mode` di ad settings:
- `require_review` → status: `pending`
- `auto_publish` → status: `active`

**Method**: POST  
**URL**: `/api/v1/mobile/ads`  
**Authentication**: Required (harus punya benefit `post_ads`)

**Request Body (native_ad)**:
```json
{
  "title": "Promo Toko Oleh-oleh Stasiun",
  "ad_type": "native_ad",
  "body_text": "Oleh-oleh khas Jawa lengkap, dekat pintu keluar Stasiun Tugu Yogyakarta.",
  "image_url": "https://cdn.tokooleholeh.id/banner.jpg",
  "sponsor_name": "Toko Oleh-oleh Bu Sari",
  "sponsor_logo_url": "https://cdn.tokooleholeh.id/logo.png",
  "click_url": "https://tokooleholeh.id",
  "start_date": "2026-06-10",
  "end_date": "2026-06-30",
  "notes": "Promo untuk pembeli dari KAI app"
}
```

**Request Body (text_ad)**:
```json
{
  "title": "Flash Sale Toko Saya",
  "ad_type": "text_ad",
  "headline": "Flash sale! Diskon 30% hari ini",
  "body_text": "Belanja oleh-oleh kereta api di toko kami dan hemat 30% untuk semua produk.",
  "cta_label": "Belanja Sekarang",
  "icon_url": null,
  "click_url": "https://tokosaya.id/flash-sale",
  "start_date": "2026-06-15",
  "end_date": "2026-06-30"
}
```

**Response (201 Created) — require_review**:
```json
{
  "data": {
    "id": "ad_012",
    "title": "Promo Toko Oleh-oleh Stasiun",
    "ad_type": "native_ad",
    "status": "pending",
    "start_date": "2026-06-10",
    "end_date": "2026-06-30",
    "created_at": "2026-06-09T08:00:00.000Z"
  },
  "message": "Iklanmu sedang direview oleh admin. Kamu akan mendapat notifikasi setelah diproses."
}
```

**Response (201 Created) — auto_publish**:
```json
{
  "data": {
    "id": "ad_012",
    "title": "Promo Toko Oleh-oleh Stasiun",
    "ad_type": "native_ad",
    "status": "active",
    "start_date": "2026-06-10",
    "end_date": "2026-06-30",
    "created_at": "2026-06-09T08:00:00.000Z"
  },
  "message": "Iklanmu sudah aktif dan mulai tayang!"
}
```

**Errors**:
```json
// 403 — tidak punya benefit post_ads
{
  "error": "FORBIDDEN",
  "message": "Fitur pasang iklan hanya tersedia untuk member Pro. Upgrade sekarang untuk mulai beriklan.",
  "upgrade_required": true
}

// 422 — sudah mencapai batas max_active_ads_per_member
{
  "error": "LIMIT_REACHED",
  "message": "Kamu sudah mencapai batas maksimum 3 iklan aktif. Selesaikan atau hapus iklan lama terlebih dahulu.",
  "current_active": 3,
  "max_allowed": 3
}

// 422 — end_date melebihi max_duration_days dari start_date
{
  "error": "VALIDATION_ERROR",
  "message": "Durasi iklan maksimum 30 hari",
  "field": "end_date"
}
```

---

## 7. GET /:id

Detail satu ads milik member sendiri.

**Method**: GET  
**URL**: `/api/v1/mobile/ads/:id`  
**Authentication**: Required

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ad_010",
    "title": "Promo Toko Oleh-oleh Stasiun",
    "ad_type": "native_ad",
    "body_text": "Oleh-oleh khas Jawa lengkap, dekat pintu keluar Stasiun Tugu Yogyakarta.",
    "image_url": "https://cdn.tokooleholeh.id/banner.jpg",
    "sponsor_name": "Toko Oleh-oleh Bu Sari",
    "sponsor_logo_url": "https://cdn.tokooleholeh.id/logo.png",
    "click_url": "https://tokooleholeh.id",
    "start_date": "2026-06-10",
    "end_date": "2026-06-30",
    "notes": "Promo untuk pembeli dari KAI app",
    "status": "active",
    "reject_reason": null,
    "total_impressions": 840,
    "total_clicks": 22,
    "ctr": 2.62,
    "created_at": "2026-06-09T08:00:00.000Z",
    "updated_at": "2026-06-10T00:00:00.000Z"
  }
}
```

**Errors**:
```json
// 403 — ads bukan milik member ini
{
  "error": "FORBIDDEN",
  "message": "Kamu tidak memiliki akses ke iklan ini"
}
```

---

## 8. PUT /:id

Update konten ads milik sendiri. Hanya bisa jika status `draft` atau `pending`.

**Method**: PUT  
**URL**: `/api/v1/mobile/ads/:id`  
**Authentication**: Required

**Request Body** (kirim hanya field yang ingin diubah):
```json
{
  "body_text": "Oleh-oleh khas Jawa lengkap dan terjangkau, dekat pintu keluar Stasiun Tugu.",
  "click_url": "https://tokooleholeh.id/promo-baru",
  "end_date": "2026-07-05"
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ad_010",
    "status": "pending",
    "updated_at": "2026-06-09T09:00:00.000Z"
  },
  "message": "Iklan berhasil diupdate"
}
```

**Errors**:
```json
// 422 — sudah active, tidak bisa edit
{
  "error": "INVALID_STATE",
  "message": "Iklan yang sedang aktif tidak bisa diedit. Pause terlebih dahulu.",
  "current_status": "active"
}
```

---

## 9. DELETE /:id

Hapus ads milik sendiri. Bisa di status apapun.

**Method**: DELETE  
**URL**: `/api/v1/mobile/ads/:id`  
**Authentication**: Required

**Response (200 OK)**:
```json
{
  "message": "Iklan berhasil dihapus"
}
```

---

## 10. PATCH /:id/status

Member bisa pause atau resume ads milik sendiri.

**Method**: PATCH  
**URL**: `/api/v1/mobile/ads/:id/status`  
**Authentication**: Required

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
    "id": "ad_010",
    "status": "paused",
    "updated_at": "2026-06-10T12:00:00.000Z"
  },
  "message": "Iklan berhasil di-pause"
}
```

**Transisi yang diizinkan untuk member**:
| Dari | Ke | Keterangan |
|------|----|------------|
| `active` | `paused` | Pause sementara |
| `paused` | `active` | Resume |

> Member tidak bisa approve/reject ads sendiri. Transisi lain (approve, reject, expire) hanya dilakukan superadmin atau sistem.

**Errors**:
```json
// 422 — transisi tidak valid
{
  "error": "INVALID_TRANSITION",
  "message": "Hanya bisa pause iklan yang sedang aktif",
  "current_status": "pending"
}
```

---

## 11. GET /:id/analytics

Analytics ads milik sendiri.

**Method**: GET  
**URL**: `/api/v1/mobile/ads/:id/analytics`  
**Authentication**: Required

**Query Parameters**:
- `date_from` (optional, default: `start_date` ads): `YYYY-MM-DD`
- `date_to` (optional, default: hari ini): `YYYY-MM-DD`

**Response (200 OK)**:
```json
{
  "data": {
    "ad_id": "ad_010",
    "ad_title": "Promo Toko Oleh-oleh Stasiun",
    "summary": {
      "total_impressions": 840,
      "total_clicks": 22,
      "ctr": 2.62
    },
    "chart": [
      {
        "date": "2026-06-10",
        "impressions": 240,
        "clicks": 6,
        "ctr": 2.5
      },
      {
        "date": "2026-06-11",
        "impressions": 310,
        "clicks": 9,
        "ctr": 2.9
      },
      {
        "date": "2026-06-12",
        "impressions": 290,
        "clicks": 7,
        "ctr": 2.41
      }
    ],
    "date_from": "2026-06-10",
    "date_to": "2026-06-12"
  }
}
```

---

## Error Codes Umum

| HTTP Status | Error Code | Keterangan |
|-------------|------------|------------|
| 401 | `UNAUTHORIZED` | Token tidak valid atau expired |
| 403 | `FORBIDDEN` | Bukan pemilik ads / tidak punya benefit `post_ads` |
| 404 | `NOT_FOUND` | Ads tidak ditemukan |
| 422 | `VALIDATION_ERROR` | Field tidak valid |
| 422 | `INVALID_STATE` | Operasi tidak valid di status saat ini |
| 422 | `INVALID_TRANSITION` | Transisi status tidak diizinkan |
| 422 | `LIMIT_REACHED` | Sudah mencapai batas max ads aktif |
| 500 | `INTERNAL_ERROR` | Server error |
