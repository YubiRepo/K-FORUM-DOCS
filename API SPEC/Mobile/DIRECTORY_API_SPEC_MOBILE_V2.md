# API Spec — Directory Module (Mobile Client) v2.1

Dokumentasi lengkap API endpoint Directory module untuk aplikasi mobile Flutter.

---

## Informasi Umum

- **Base URL**: `/api/v1/mobile/directory`
- **Auth Header**: `Authorization: Bearer <access_token>`
- **Content-Type**: `application/json`
- **Pagination**: `limit` (default 20, max 100) + `offset`

### Response Envelope

```json
// Success
{
  "data": { ... },
  "message": "Success message"
}

// Success (list)
{
  "data": [ ... ],
  "pagination": { "limit": 20, "offset": 0, "total": 145 }
}

// Error
{
  "message": "Error message",
  "errors": { "field": ["validation message"] }  // hanya untuk 422
}
```

### HTTP Status Codes

| Code | Meaning                       |
| ---- | ----------------------------- |
| 200  | OK                            |
| 201  | Created                       |
| 204  | No Content (delete)           |
| 400  | Bad Request                   |
| 401  | Unauthorized                  |
| 403  | Forbidden (kurang permission) |
| 404  | Not Found                     |
| 409  | Conflict (duplicate)          |
| 422  | Validation Error              |
| 500  | Internal Server Error         |

---

## ENDPOINT INDEX

| #                     | Method | Path                                    | Auth     | Akses              |
| --------------------- | ------ | --------------------------------------- | -------- | ------------------ |
| **MEDIA**             |        |                                         |          |                    |
| 1                     | POST   | `/media/upload`                         | ✅        | Member Pro         |
| **COMPANY**           |        |                                         |          |                    |
| 2                     | GET    | `/company`                              | ✅        | Member Pro         |
| 3                     | POST   | `/company`                              | ✅        | Member Pro         |
| 4                     | GET    | `/company/{id}`                         | ✅        | Member Pro (owner) |
| 5                     | PUT    | `/company/{id}`                         | ✅        | Member Pro (owner) |
| 6                     | DELETE | `/company/{id}`                         | ✅        | Member Pro (owner) |
| **MERCHANT — Owner**  |        |                                         |          |                    |
| 7                     | POST   | `/merchants`                             | ✅        | Member Pro         |
| 8                     | GET    | `/me/merchants`                          | ✅        | Member Pro         |
| 9                     | GET    | `/merchants/{id}/manage`                 | ✅        | Member Pro (owner) |
| 10                    | PUT    | `/merchants/{id}`                        | ✅        | Member Pro (owner) |
| 11                    | POST   | `/merchants/{id}/publish`                | ✅        | Member Pro (owner) |
| 12                    | POST   | `/merchants/{id}/archive`                | ✅        | Member Pro (owner) |
| 13                    | POST   | `/merchants/{id}/unarchive`              | ✅        | Member Pro (owner) |
| 14                    | DELETE | `/merchants/{id}`                        | ✅        | Member Pro (owner) |
| **MERCHANT — Public** |        |                                         |          |                    |
| 15                    | GET    | `/merchants`                             | Optional | All                |
| 16                    | GET    | `/merchants/{id}`                        | Optional | All                |
| **ITEMS**             |        |                                         |          |                    |
| 17                    | GET    | `/merchants/{id}/item`                   | Optional | All                |
| 18                    | GET    | `/merchants/{id}/item/{item_id}`         | Optional | All                |
| 19                    | POST   | `/merchants/{id}/item`                   | ✅        | Member Pro (owner) |
| 20                    | PUT    | `/merchants/{id}/item/{item_id}`         | ✅        | Member Pro (owner) |
| 21                    | DELETE | `/merchants/{id}/item/{item_id}`         | ✅        | Member Pro (owner) |
| 22                    | PUT    | `/merchants/{id}/item/reorder`           | ✅        | Member Pro (owner) |
| **REVIEWS**           |        |                                         |          |                    |
| 23                    | GET    | `/merchants/{id}/review`                 | Optional | All                |
| 24                    | POST   | `/merchants/{id}/review`                 | ✅        | Any member         |
| 25                    | PUT    | `/merchants/{id}/review`                 | ✅        | Reviewer (own)     |
| 26                    | POST   | `/review/{review_id}/vote`              | ✅        | Any member         |
| **FAVORITES (Save/Bookmark)** |  |                                     |          |                    |
| 27                    | GET    | `/me/saved`                             | ✅        | Any member         |
| 28                    | POST   | `/merchants/{id}/save`                   | ✅        | Any member         |
| 29                    | DELETE | `/merchants/{id}/save`                   | ✅        | Any member         |
| **LIKES (Reaction ❤️)** |      |                                         |          |                    |
| 27a                   | POST   | `/merchants/{id}/like`                   | ✅        | Any member         |
| 27b                   | DELETE | `/merchants/{id}/like`                   | ✅        | Any member         |
| **INQUIRIES**         |        |                                         |          |                    |
| 30                    | POST   | `/merchants/{id}/inquiry`                | ✅        | Any member         |
| 31                    | GET    | `/me/inquiry`                           | ✅        | Any member         |
| 32                    | GET    | `/merchants/{id}/inquiry`                | ✅        | Owner              |
| 33                    | POST   | `/merchants/{id}/inquiry/{inq_id}/reply` | ✅        | Owner              |
| 34                    | POST   | `/merchants/{id}/inquiry/{inq_id}/close` | ✅        | Owner              |
| **CATEGORIES**        |        |                                         |          |                    |
| 35                    | GET    | `/categories`                           | Optional | All                |

---

## MEDIA

### 1. Upload Media

Upload image sebelum create merchant/item. Return CDN URL.

- **POST** `/api/v1/mobile/media/upload`
- **Auth**: Member Pro
- **Content-Type**: `multipart/form-data`

**Request:**
```
files[]: image1.jpg  (max 5 files, max 5MB each, format: jpg/jpeg/png/webp)
context: "directory"
```

**Response 201:**
```json
{
  "data": {
    "urls": [
      "https://cdn.example.com/directory/img_abc123.jpg",
      "https://cdn.example.com/directory/img_def456.jpg"
    ]
  },
  "message": "2 image(s) uploaded successfully"
}
```

**Errors:**
- `400` — format file tidak didukung
- `422` — file terlalu besar (max 5MB), atau melebihi 5 file

---

## COMPANY

### 2. List My Companies

- **GET** `/api/v1/mobile/directory/company`
- **Auth**: Member Pro

**Query Params:**
| Param  | Type | Default | Deskripsi |
| ------ | ---- | ------- | --------- |
| limit  | int  | 20      |           |
| offset | int  | 0       |           |

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-cmp-001",
      "name": "PT Impor Korea",
      "logo_url": "https://cdn.../logo.jpg",
      "phone": "0212345678",
      "email": "contact@ptkore.com",
      "website": "https://ptkore.com",
      "merchant_count": 3,
      "status": "active",
      "created_at": "2026-01-15T00:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 2 }
}
```

---

### 3. Create Company

- **POST** `/api/v1/mobile/directory/company`
- **Auth**: Member Pro

**Request Body:**
```json
{
  "name": "PT Impor Korea",
  "description": "Distributor produk Korea authentic...",
  "logo_url": "https://cdn.../logo.jpg",
  "phone": "0212345678",
  "email": "contact@ptkore.com",
  "website": "https://ptkore.com"
}
```

**Validasi:**
| Field       | Rule                       |
| ----------- | -------------------------- |
| name        | Required, min 3, max 200   |
| description | Optional, max 2000         |
| logo_url    | Optional, valid URL        |
| phone       | Optional, format Indonesia |
| email       | Optional, valid email      |
| website     | Optional, valid URL        |

**Response 201:**
```json
{
  "data": {
    "id": "uuid-cmp-001",
    "owner_id": "uuid-user-001",
    "name": "PT Impor Korea",
    "logo_url": "https://cdn.../logo.jpg",
    "merchant_count": 0,
    "status": "active",
    "created_at": "2026-05-30T10:00:00Z"
  },
  "message": "Company created successfully"
}
```

---

### 4. Get Company Detail

- **GET** `/api/v1/mobile/directory/company/{company_id}`
- **Auth**: Member Pro (owner only)

**Response 200:**
```json
{
  "data": {
    "id": "uuid-cmp-001",
    "owner_id": "uuid-user-001",
    "name": "PT Impor Korea",
    "description": "Distributor produk Korea...",
    "logo_url": "https://cdn.../logo.jpg",
    "phone": "0212345678",
    "email": "contact@ptkore.com",
    "website": "https://ptkore.com",
    "merchant_count": 3,
    "status": "active",
    "created_at": "2026-01-15T00:00:00Z",
    "updated_at": "2026-05-20T10:00:00Z"
  }
}
```

---

### 5. Update Company

- **PUT** `/api/v1/mobile/directory/company/{company_id}`
- **Auth**: Member Pro (owner only)

**Request Body:** *(semua field optional, partial update)*
```json
{
  "name": "PT Impor Korea Baru",
  "logo_url": "https://cdn.../new_logo.jpg",
  "phone": "0812999999"
}
```

**Response 200:**
```json
{
  "data": { ...updated company object... },
  "message": "Company updated successfully"
}
```

---

### 6. Delete Company

- **DELETE** `/api/v1/mobile/directory/company/{company_id}`
- **Auth**: Member Pro (owner only)
- **Rule**: Hanya bisa delete jika `merchant_count = 0`

**Response 204:** (No Content)

**Errors:**
- `409` — Company masih punya merchant aktif

---

## MERCHANT — OWNER ENDPOINTS

### 7. Create Merchant

- **POST** `/api/v1/mobile/directory/merchants`
- **Auth**: Member Pro

**Request Body:**
```json
{
  "company_id": "uuid-cmp-001",
  "name": "Toko Jakarta Pusat",
  "description": "Outlet utama kami di Jakarta dengan koleksi lengkap produk Korea...",
  "type": "retail",
  "category_ids": ["uuid-cat-retail", "uuid-cat-food"],
  "images": [
    { "url": "https://cdn.../merchants_1.jpg", "is_primary": true, "sort_order": 1 },
    { "url": "https://cdn.../merchants_2.jpg", "is_primary": false, "sort_order": 2 }
  ],
  "address": "Jl. Ahmad Yani No. 123, Jakarta Pusat",
  "latitude": -6.2088,
  "longitude": 106.8905,
  "phone": "0212345678",
  "email": "jakarta@ptkore.com",
  "whatsapp": "0812345678",
  "instagram": "@ptkore_jakarta",
  "hours": {
    "monday":    { "open": "09:00", "close": "18:00", "closed": false },
    "tuesday":   { "open": "09:00", "close": "18:00", "closed": false },
    "wednesday": { "open": "09:00", "close": "18:00", "closed": false },
    "thursday":  { "open": "09:00", "close": "18:00", "closed": false },
    "friday":    { "open": "09:00", "close": "20:00", "closed": false },
    "saturday":  { "open": "10:00", "close": "20:00", "closed": false },
    "sunday":    { "closed": true }
  },
  "allow_reviews": true,
  "allow_inquiries": true
}
```

**Validasi:**
| Field        | Rule                                                       |
| ------------ | ---------------------------------------------------------- |
| company_id   | Required, must own company                                 |
| name         | Required, min 3, max 200                                   |
| description  | Required, min 20, max 3000                                 |
| type         | Required: retail/online/service/food_beverage/beauty/other |
| category_ids | Required, min 1, max 5, valid category_master IDs          |
| images       | Min 1 (jika setting `min_images` = 1)                      |
| address      | Required jika type != online                               |
| latitude     | Required jika type != online, valid range                  |
| longitude    | Required jika type != online, valid range                  |

**Response 201:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "company_id": "uuid-cmp-001",
    "name": "Toko Jakarta Pusat",
    "status": "pending_approval",
    "created_at": "2026-05-30T10:00:00Z"
  },
  "message": "Merchant submitted for approval"
}
```

*Note: `status` bisa `published` langsung jika setting `require_approval = false`*

---

### 8. List My Merchants

- **GET** `/api/v1/mobile/directory/me/merchants`
- **Auth**: Member Pro

**Query Params:** `company_id` (optional), `status`, `limit`, `offset`

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-mch-001",
      "company_id": "uuid-cmp-001",
      "company_name": "PT Impor Korea",
      "name": "Toko Jakarta Pusat",
      "type": "retail",
      "status": "published",
      "approval_status": "approved",
      "city": "Jakarta",
      "rating": 4.5,
      "review_count": 123,
      "item_count": 45,
      "primary_image": "https://cdn.../merchants_1.jpg",
      "created_at": "2026-05-30T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 3 }
}
```

---

### 9. Get My Merchant (Owner Detail)

- **GET** `/api/v1/mobile/directory/merchants/{merchant_id}/manage`
- **Auth**: Member Pro (owner only)
- **Deskripsi**: Versi lengkap dengan settings, stats, rejection_reason, dll.

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "company_id": "uuid-cmp-001",
    "name": "Toko Jakarta Pusat",
    "description": "...",
    "type": "retail",
    "categories": [
      { "id": "uuid-cat-retail", "name": "Retail", "icon": "🏬", "is_primary": true }
    ],
    "images": [
      { "id": "uuid-img-001", "url": "https://cdn.../m1.jpg", "is_primary": true, "sort_order": 1 }
    ],
    "address": "Jl. Ahmad Yani No. 123",
    "city": "Jakarta Pusat",
    "province": "DKI Jakarta",
    "latitude": -6.2088,
    "longitude": 106.8905,
    "phone": "0212345678",
    "email": "jakarta@ptkore.com",
    "whatsapp": "0812345678",
    "instagram": "@ptkore_jakarta",
    "hours": { ... },
    "status": "published",
    "approval_status": "approved",
    "rejection_reason": null,
    "settings": {
      "allow_reviews": true,
      "allow_inquiries": true,
      "auto_reply_enabled": false,
      "auto_reply_message": null
    },
    "stats": {
      "item_count": 45,
      "review_count": 123,
      "rating": 4.5,
      "favorite_count": 234,
      "inquiry_count": 45,
      "view_count": 5420
    },
    "published_at": "2026-02-12T10:00:00Z",
    "created_at": "2026-02-10T00:00:00Z",
    "updated_at": "2026-05-20T10:00:00Z"
  }
}
```

**Field `stats`** (dipakai untuk kartu Views / Rating / Items / Inquiries di layar kelola merchant):

| Field            | Type   | Arti                                                                 |
|------------------|--------|----------------------------------------------------------------------|
| `item_count`     | int    | Jumlah item/produk merchant                                          |
| `review_count`   | int    | Jumlah ulasan                                                        |
| `rating`         | number | Rata-rata rating (0–5); `0` jika belum ada ulasan                    |
| `favorite_count` | int    | Jumlah **like** ❤️ (digerakkan endpoint `/like` #27a/27b)            |
| `inquiry_count`  | int    | Jumlah inquiry masuk                                                 |
| `view_count`     | int    | Jumlah tampilan (increment 1×/user/hari saat buka detail publik #16) |

> ⚠️ **Nama field wajib `view_count`** (bukan `total_views`). Mobile sempat membaca `total_views` — sudah diselaraskan ke `view_count`, tapi juga menerima `total_views` sebagai fallback. Kirim `view_count`. Semua field `stats` **wajib ada** (kirim `0` bila belum ada data) agar kartu tidak menampilkan kosong.

---

### 10. Update Merchant

- **PUT** `/api/v1/mobile/directory/merchants/{merchant_id}`
- **Auth**: Member Pro (owner only)
- **Note**: Jika merchant sudah `published`, update akan set status ke `pending_approval` kembali (jika `require_approval = true`)

**Request Body:** *(partial update, semua field optional)*
```json
{
  "name": "Toko Jakarta Pusat Updated",
  "description": "Deskripsi baru...",
  "category_ids": ["uuid-cat-retail"],
  "images": [
    { "url": "https://cdn.../new1.jpg", "is_primary": true, "sort_order": 1 }
  ],
  "phone": "0219999999",
  "hours": { ... },
  "allow_inquiries": false
}
```

**Response 200:**
```json
{
  "data": { ...updated merchant object... },
  "message": "Merchant updated successfully"
}
```

---

### 11. Publish Merchant

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/publish`
- **Auth**: Member Pro (owner only)

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "status": "pending_approval",
    "message": "Merchant submitted for review"
  }
}
```

*Jika auto-approve:*
```json
{
  "data": {
    "id": "uuid-mch-001",
    "status": "published",
    "published_at": "2026-05-30T10:00:00Z"
  },
  "message": "Merchant published successfully"
}
```

---

### 12 & 13. Archive / Unarchive Merchant

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/archive`
- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/unarchive`
- **Auth**: Member Pro (owner only)

**Response 200:**
```json
{ "message": "Merchant archived successfully" }
{ "message": "Merchant unarchived. Status set to draft." }
```

---

### 14. Delete Merchant

- **DELETE** `/api/v1/mobile/directory/merchants/{merchant_id}`
- **Auth**: Member Pro (owner only)
- **Rule**: Hanya bisa jika belum ada reviews atau inquiries

**Response 204:** (No Content)

**Errors:**
- `409` — Merchant punya reviews/inquiries, gunakan archive

---

## MERCHANT — PUBLIC ENDPOINTS

### 15. List Merchants (Public Directory)

- **GET** `/api/v1/mobile/directory/merchants`
- **Auth**: Optional

**Query Params:**
| Param       | Type    | Deskripsi                                                  |
| ----------- | ------- | ---------------------------------------------------------- |
| q           | string  | Full-text search nama/deskripsi                            |
| category_id | UUID    | Filter by kategori                                         |
| type        | string  | retail/online/service/food_beverage/beauty/other           |
| city        | string  | Filter by kota (case-insensitive)                          |
| rating_min  | float   | Min rating (1.0–5.0)                                       |
| is_open_now | boolean | Filter yang sedang buka                                    |
| has_product | boolean | Ada item product                                           |
| has_service | boolean | Ada item service                                           |
| sort        | string  | `rating_desc`(default)/`newest`/`name_asc`/`most_reviewed` |
| limit       | int     | Default 20, max 100                                        |
| offset      | int     | Default 0                                                  |

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-mch-001",
      "name": "Toko Jakarta Pusat",
      "type": "retail",
      "categories": [
        { "id": "uuid-cat-retail", "name": "Retail", "icon": "🏬" }
      ],
      "primary_image": "https://cdn.../merchants_1.jpg",
      "city": "Jakarta Pusat",
      "province": "DKI Jakarta",
      "rating": 4.5,
      "review_count": 123,
      "item_count": 45,
      "is_open_now": true,
      "is_featured": false,
      "is_saved": false,  // null jika tidak login
      "is_liked": false,  // null jika tidak login
      "favorite_count": 234,
      "published_at": "2026-02-12T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 245 }
}
```

---

### 16. Get Merchant Detail (Public)

- **GET** `/api/v1/mobile/directory/merchants/{merchant_id}`
- **Auth**: Optional
- **Side Effect**: Increment view count (1x per user per hari)

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "name": "Toko Jakarta Pusat",
    "description": "Outlet utama kami...",
    "type": "retail",
    "categories": [
      { "id": "uuid-cat-retail", "name": "Retail", "icon": "🏬" }
    ],
    "images": [
      { "url": "https://cdn.../m1.jpg", "is_primary": true },
      { "url": "https://cdn.../m2.jpg", "is_primary": false }
    ],
    "location": {
      "address": "Jl. Ahmad Yani No. 123, Jakarta Pusat",
      "city": "Jakarta Pusat",
      "province": "DKI Jakarta",
      "latitude": -6.2088,
      "longitude": 106.8905
    },
    "contact": {
      "phone": "0212345678",
      "email": "jakarta@ptkore.com",
      "whatsapp": "0812345678",
      "instagram": "@ptkore_jakarta"
    },
    "hours": {
      "monday":    { "open": "09:00", "close": "18:00", "closed": false },
      "tuesday":   { "open": "09:00", "close": "18:00", "closed": false },
      "wednesday": { "open": "09:00", "close": "18:00", "closed": false },
      "thursday":  { "open": "09:00", "close": "18:00", "closed": false },
      "friday":    { "open": "09:00", "close": "20:00", "closed": false },
      "saturday":  { "open": "10:00", "close": "20:00", "closed": false },
      "sunday":    { "closed": true }
    },
    "is_open_now": true,
    "rating": 4.5,
    "review_count": 123,
    "item_count": 45,
    "favorite_count": 234,
    "view_count": 5420,
    "is_saved": false,
    "is_liked": false,
    "is_featured": false,
    "allow_reviews": true,
    "allow_inquiries": true,
    "company": {
      "id": "uuid-cmp-001",
      "name": "PT Impor Korea",
      "logo_url": "https://cdn.../logo.jpg"
    },
    "published_at": "2026-02-12T10:00:00Z"
  }
}
```

---

## ITEMS

### 17. List Merchant Items (Public)

- **GET** `/api/v1/mobile/directory/merchants/{merchant_id}/item`
- **Auth**: Optional

**Query Params:**
| Param  | Type   | Deskripsi             |
| ------ | ------ | --------------------- |
| type   | string | product/service       |
| status | string | available/unavailable |
| q      | string | Search nama item      |
| limit  | int    | Default 20            |
| offset | int    | Default 0             |

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-item-001",
      "type": "product",
      "name": "Korean Red Ginseng",
      "description": "Authentic Korean red ginseng...",
      "category": "Health & Wellness",
      "price": 150000,
      "currency": "IDR",
      "unit": "per pack",
      "stock": null,
      "status": "available",
      "primary_image": "https://cdn.../product_1.jpg",
      "sort_order": 1
    },
    {
      "id": "uuid-item-002",
      "type": "service",
      "name": "Korean Beauty Treatment",
      "description": "Authentic skincare treatment...",
      "category": "Beauty",
      "price_min": 200000,
      "price_max": 500000,
      "currency": "IDR",
      "duration_minutes": 60,
      "status": "available",
      "primary_image": "https://cdn.../service_1.jpg",
      "sort_order": 2
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 45 }
}
```

---

### 18. Get Item Detail (Public)

- **GET** `/api/v1/mobile/directory/merchants/{merchant_id}/item/{item_id}`
- **Auth**: Optional

**Response 200:**
```json
{
  "data": {
    "id": "uuid-item-001",
    "merchant_id": "uuid-mch-001",
    "type": "product",
    "name": "Korean Red Ginseng",
    "description": "Authentic Korean red ginseng from Geumsan...",
    "category": "Health & Wellness",
    "price": 150000,
    "currency": "IDR",
    "unit": "per pack",
    "stock": null,
    "status": "available",
    "images": [
      { "url": "https://cdn.../product_1.jpg", "is_primary": true, "sort_order": 1 },
      { "url": "https://cdn.../product_2.jpg", "is_primary": false, "sort_order": 2 }
    ],
    "sort_order": 1,
    "created_at": "2026-03-10T00:00:00Z"
  }
}
```

---

### 19. Create Item

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/item`
- **Auth**: Member Pro (owner only)

**Request Body (Product):**
```json
{
  "type": "product",
  "name": "Korean Red Ginseng",
  "description": "Authentic Korean red ginseng...",
  "category": "Health & Wellness",
  "price": 150000,
  "currency": "IDR",
  "unit": "per pack",
  "stock": null,
  "images": [
    { "url": "https://cdn.../p1.jpg", "is_primary": true, "sort_order": 1 }
  ],
  "status": "available"
}
```

**Request Body (Service):**
```json
{
  "type": "service",
  "name": "Korean Beauty Treatment",
  "description": "Full Korean skincare treatment...",
  "category": "Beauty",
  "price_min": 200000,
  "price_max": 500000,
  "currency": "IDR",
  "duration_minutes": 60,
  "images": [
    { "url": "https://cdn.../s1.jpg", "is_primary": true, "sort_order": 1 }
  ],
  "status": "available"
}
```

**Validasi:**
| Field     | Rule                                     |
| --------- | ---------------------------------------- |
| type      | Required: product/service                |
| name      | Required, min 3, max 200                 |
| price     | Required jika type=product, > 0          |
| price_min | Required jika type=service, > 0          |
| price_max | Required jika type=service, >= price_min |
| images    | Min 1, max 10                            |

**Response 201:**
```json
{
  "data": { ...item object... },
  "message": "Item created successfully"
}
```

---

### 20. Update Item

- **PUT** `/api/v1/mobile/directory/merchants/{merchant_id}/item/{item_id}`
- **Auth**: Member Pro (owner only)

**Request Body:** *(partial update)*
```json
{
  "price": 175000,
  "stock": 50,
  "status": "available"
}
```

**Response 200:**
```json
{
  "data": { ...updated item object... },
  "message": "Item updated successfully"
}
```

---

### 21. Delete Item

- **DELETE** `/api/v1/mobile/directory/merchants/{merchant_id}/item/{item_id}`
- **Auth**: Member Pro (owner only)

**Response 204:** (No Content)

---

### 22. Reorder Items

- **PUT** `/api/v1/mobile/directory/merchants/{merchant_id}/item/reorder`
- **Auth**: Member Pro (owner only)

**Request Body:**
```json
{
  "order": [
    { "item_id": "uuid-item-002", "sort_order": 1 },
    { "item_id": "uuid-item-001", "sort_order": 2 },
    { "item_id": "uuid-item-003", "sort_order": 3 }
  ]
}
```

**Response 200:**
```json
{ "message": "Items reordered successfully" }
```

---

## REVIEWS

### 23. List Merchant Reviews (Public)

- **GET** `/api/v1/mobile/directory/merchants/{merchant_id}/review`
- **Auth**: Optional

**Query Params:** `rating` (1-5 filter), `sort` (newest/most_helpful), `limit`, `offset`

**Response 200:**
```json
{
  "data": {
    "summary": {
      "average_rating": 4.5,
      "total_reviews": 123,
      "distribution": { "5": 65, "4": 35, "3": 15, "2": 5, "1": 3 }
    },
    "reviews": [
      {
        "id": "uuid-rev-001",
        "user_name": "Budi Santoso",
        "user_avatar": "https://cdn.../avatar.jpg",
        "rating": 4,
        "title": "Produk bagus, tapi lama delivery",
        "review_text": "Produk original dan kualitas bagus...",
        "aspects": {
          "product_quality": 5,
          "service": 4,
          "price": 4
        },
        "helpful_count": 23,
        "unhelpful_count": 2,
        "my_vote": null,
        "created_at": "2026-04-20T10:00:00Z"
      }
    ]
  },
  "pagination": { "limit": 20, "offset": 0, "total": 123 }
}
```

---

### 24. Create Review

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/review`
- **Auth**: Any member (tidak boleh owner sendiri)

**Request Body:**
```json
{
  "rating": 4,
  "title": "Produk bagus, tapi lama delivery",
  "review_text": "Produk original dan kualitas bagus...",
  "aspects": {
    "product_quality": 5,
    "service": 4,
    "price": 4
  }
}
```

**Response 201:**
```json
{
  "data": { ...review object... },
  "message": "Review submitted. Waiting for moderation."
}
```

*Jika moderation OFF: `"message": "Review published successfully"`*

**Errors:**
- `409` — User sudah pernah review merchant ini (gunakan PUT untuk edit)
- `403` — Owner tidak bisa review sendiri

---

### 25. Update My Review

- **PUT** `/api/v1/mobile/directory/merchants/{merchant_id}/review`
- **Auth**: Reviewer (own review only)

**Request Body:** *(partial update)*
```json
{
  "rating": 5,
  "review_text": "Update: delivery sudah lebih cepat sekarang!"
}
```

**Response 200:**
```json
{
  "data": { ...updated review object... },
  "message": "Review updated"
}
```

---

### 26. Vote Review (Helpful/Unhelpful)

- **POST** `/api/v1/mobile/directory/review/{review_id}/vote`
- **Auth**: Any member

**Request Body:**
```json
{ "vote": "helpful" }
```

*Vote: `helpful` atau `unhelpful`. Vote ulang dengan value yang sama → unvote.*

**Response 200:**
```json
{
  "data": {
    "helpful_count": 24,
    "unhelpful_count": 2,
    "my_vote": "helpful"
  }
}
```

---

## FAVORITES

### 27. Get My Saved Merchants

- **GET** `/api/v1/mobile/directory/me/saved`
- **Auth**: Any member

**Query Params:** `limit`, `offset`

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-fav-001",
      "merchant": {
        "id": "uuid-mch-001",
        "name": "Toko Jakarta Pusat",
        "type": "retail",
        "city": "Jakarta Pusat",
        "rating": 4.5,
        "primary_image": "https://cdn.../m1.jpg",
        "is_open_now": true
      },
      "note": "Toko favorit untuk beli ginseng",
      "saved_at": "2026-05-20T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 12 }
}
```

---

### 28. Save Merchant

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/save`
- **Auth**: Any member

**Request Body:** *(optional)*
```json
{ "note": "Toko favorit untuk beli ginseng" }
```

**Response 201:**
```json
{ "message": "Merchant saved to favorites" }
```

**Errors:**
- `409` — Sudah di-save sebelumnya

---

### 29. Unsave Merchant

- **DELETE** `/api/v1/mobile/directory/merchants/{merchant_id}/save`
- **Auth**: Any member

**Response 204:** (No Content)

---

## LIKES (Reaction ❤️)

> **Save vs Like — dua hal berbeda:**
> - **Save** (`/save`, #28–29) = *bookmark*. Menyimpan merchant ke daftar **"Saved"** milik user (muncul di `GET /me/saved`). Menggerakkan field `is_saved`.
> - **Like** (`/like`, di bawah) = *reaction* ❤️ (tombol hati di merchant detail). Menggerakkan hitungan `favorite_count` (dipakai di stats) dan field per-user `is_liked`. Tidak masuk daftar "Saved".
>
> Keduanya idempotent: like saat sudah like → tetap 200 (tanpa dobel hitung); unlike saat belum like → tetap 204/200.

### 27a. Like Merchant

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/like`
- **Auth**: Any member (login wajib)
- **Deskripsi**: Menandai merchant sebagai disukai oleh user saat ini. Menaikkan `favorite_count` sebanyak 1 (hanya jika sebelumnya belum like) dan set `is_liked = true`.

**Request Body**: — (kosong)

**Response 200:**
```json
{
  "data": {
    "merchant_id": "uuid-mch-001",
    "is_liked": true,
    "favorite_count": 235
  }
}
```

**Error Responses:**

| HTTP | error_code                | Kondisi                          |
|------|---------------------------|----------------------------------|
| 401  | `ERR_UNAUTHORIZED`        | Belum login                      |
| 404  | `ERR_NOT_FOUND`           | Merchant tidak ditemukan/arsip   |

> Catatan: jika user sudah like sebelumnya, tetap balas **200** dengan state terkini (idempotent, `favorite_count` tidak bertambah lagi).

### 27b. Unlike Merchant

- **DELETE** `/api/v1/mobile/directory/merchants/{merchant_id}/like`
- **Auth**: Any member (login wajib)
- **Deskripsi**: Membatalkan like. Menurunkan `favorite_count` sebanyak 1 (tidak pernah di bawah 0) dan set `is_liked = false`.

**Response 200:**
```json
{
  "data": {
    "merchant_id": "uuid-mch-001",
    "is_liked": false,
    "favorite_count": 234
  }
}
```

> Boleh juga balas **204 No Content**; mobile menangani keduanya (jika body ada, dipakai untuk sinkron `favorite_count`).

---

## INQUIRIES

### 30. Send Inquiry

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/inquiry`
- **Auth**: Any member

**Request Body:**
```json
{
  "subject": "Apakah stok Korean Ginseng masih ada?",
  "message": "Halo, saya ingin menanyakan ketersediaan stok Korean Red Ginseng ukuran 100g. Terima kasih."
}
```

**Response 201:**
```json
{
  "data": {
    "id": "uuid-inq-001",
    "merchant_name": "Toko Jakarta Pusat",
    "subject": "Apakah stok Korean Ginseng masih ada?",
    "status": "pending",
    "created_at": "2026-05-30T10:00:00Z"
  },
  "message": "Inquiry sent successfully"
}
```

**Errors:**
- `403` — `allow_inquiries = false` pada merchant

---

### 31. My Inquiries (Customer View)

- **GET** `/api/v1/mobile/directory/me/inquiry`
- **Auth**: Any member

**Query Params:** `status` (pending/replied/closed), `limit`, `offset`

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-inq-001",
      "merchant": {
        "id": "uuid-mch-001",
        "name": "Toko Jakarta Pusat",
        "primary_image": "https://cdn.../m1.jpg"
      },
      "subject": "Apakah stok Korean Ginseng masih ada?",
      "message": "Halo, saya ingin menanyakan...",
      "status": "replied",
      "merchant_reply": "Ya, masih ada stok. Silakan order langsung.",
      "replied_at": "2026-05-30T14:00:00Z",
      "created_at": "2026-05-30T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 5 }
}
```

---

### 32. Merchant Inquiries (Owner View)

- **GET** `/api/v1/mobile/directory/merchants/{merchant_id}/inquiry`
- **Auth**: Member Pro (owner only)

**Query Params:** `status`, `limit`, `offset`

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-inq-001",
      "user_name": "Citra Dewi",
      "user_avatar": "https://cdn.../avatar.jpg",
      "subject": "Apakah stok Korean Ginseng masih ada?",
      "message": "...",
      "status": "pending",
      "merchant_reply": null,
      "created_at": "2026-05-30T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 8 }
}
```

---

### 33. Reply to Inquiry

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/inquiry/{inquiry_id}/reply`
- **Auth**: Member Pro (owner only)

**Request Body:**
```json
{ "reply": "Ya, masih ada stok. Silakan order langsung ke WA kami." }
```

**Response 200:**
```json
{
  "data": {
    "id": "uuid-inq-001",
    "status": "replied",
    "merchant_reply": "Ya, masih ada stok...",
    "replied_at": "2026-05-30T14:00:00Z"
  },
  "message": "Reply sent successfully"
}
```

---

### 34. Close Inquiry

- **POST** `/api/v1/mobile/directory/merchants/{merchant_id}/inquiry/{inquiry_id}/close`
- **Auth**: Member Pro (owner only)

**Response 200:**
```json
{ "message": "Inquiry closed" }
```

---

## CATEGORIES

### 35. List Categories (Public)

- **GET** `/api/v1/mobile/directory/categories`
- **Auth**: Optional

**Response 200:**
```json
{
  "data": [
    { "id": "uuid-cat-retail", "name": "Retail", "icon": "🏬", "sort_order": 1 },
    { "id": "uuid-cat-food", "name": "Food & Beverage", "icon": "🍽️", "sort_order": 2 },
    { "id": "uuid-cat-service", "name": "Service", "icon": "🔧", "sort_order": 3 },
    { "id": "uuid-cat-beauty", "name": "Beauty & Salon", "icon": "💄", "sort_order": 4 },
    { "id": "uuid-cat-health", "name": "Health & Wellness", "icon": "🏥", "sort_order": 5 }
  ]
}
```

---

## Error Responses

```json
// 401 Unauthorized
{ "message": "Authentication required" }

// 403 Forbidden
{ "message": "Only Member Pro can create merchants" }
{ "message": "You are not the owner of this merchant" }
{ "message": "Cannot review your own merchant" }

// 404 Not Found
{ "message": "Merchant not found" }
{ "message": "Item not found" }

// 409 Conflict
{ "message": "You have already reviewed this merchant" }
{ "message": "Merchant already saved to favorites" }

// 422 Validation Error
{
  "message": "Validation failed",
  "errors": {
    "name": ["Merchant name is required", "Minimum 3 characters"],
    "latitude": ["Latitude is required for physical merchant types"],
    "category_ids": ["At least 1 category is required"]
  }
}
```

---

## Changelog

| Versi | Tanggal    | Perubahan |
|-------|------------|-----------|
| 2.1   | 2026-07-21 | Tambah endpoint **Like/Unlike** merchant (`POST`/`DELETE /merchants/{id}/like`, #27a/27b) + field `is_liked` di list #15 & detail publik #16; tegaskan beda **Save** (bookmark) vs **Like** (reaction ❤️ → `favorite_count`); dokumentasikan arti tiap field `stats` dan tegaskan nama field views = **`view_count`** (bukan `total_views`) |
| 2.0   | 2026-05-30 | Rilis awal v2 |

---

*API Spec Directory Module — Mobile Client v2.1. KAI App. Last updated: 2026-07-21*
