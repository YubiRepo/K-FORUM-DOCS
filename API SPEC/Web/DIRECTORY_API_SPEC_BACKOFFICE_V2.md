# API Spec — Directory Module (Web Backoffice Admin) v2.0

Dokumentasi lengkap API endpoint Directory module untuk backoffice admin dashboard (Superadmin & Admin Regional).

---

## Informasi Umum

- **Base URL**: `/api/v1/web/directory`
- **Auth Header**: `Authorization: Bearer <access_token>`
- **Content-Type**: `application/json`

### Response Envelope

```json
// Success
{ "data": { ... }, "message": "..." }

// List
{ "data": [...], "pagination": { "limit": 20, "offset": 0, "total": 145 } }

// Error
{ "message": "...", "errors": { ... } }
```

---

## ROLE-BASED ACCESS MATRIX

| Endpoint | Superadmin | Admin Regional |
|----------|-----------|----------------|
| **COMPANIES** | | |
| GET /companies | ✅ All | ✅ Own region only |
| GET /companies/:id | ✅ | ✅ Own region |
| POST /companies/:id/ban | ✅ | ❌ |
| POST /companies/:id/unban | ✅ | ❌ |
| DELETE /companies/:id | ✅ (force) | ❌ |
| **MERCHANTS** | | |
| GET /merchants | ✅ All | ✅ Own region only |
| GET /merchants/:id | ✅ | ✅ Own region |
| POST /merchants/:id/approve | ✅ | ✅ Own region |
| POST /merchants/:id/reject | ✅ | ✅ Own region |
| POST /merchants/:id/ban | ✅ | ✅ Own region |
| POST /merchants/:id/unban | ✅ | ❌ |
| POST /merchants/:id/feature | ✅ | ❌ |
| DELETE /merchants/:id | ✅ (force) | ❌ |
| **ITEMS** | | |
| GET /merchants/:id/items | ✅ | ✅ Own region |
| DELETE /merchants/:id/items/:item_id | ✅ | ✅ Own region |
| **REVIEWS** | | |
| GET /reviews | ✅ All | ✅ Own region |
| GET /reviews/:id | ✅ | ✅ Own region |
| POST /reviews/:id/approve | ✅ | ✅ Own region |
| POST /reviews/:id/reject | ✅ | ✅ Own region |
| POST /reviews/:id/hide | ✅ | ✅ Own region |
| DELETE /reviews/:id | ✅ | ❌ |
| **CATEGORIES** | | |
| GET /categories | ✅ | ✅ (read only) |
| POST /categories | ✅ | ❌ |
| PUT /categories/:id | ✅ | ❌ |
| DELETE /categories/:id | ✅ | ❌ |
| PUT /categories/reorder | ✅ | ❌ |
| **SETTINGS** | | |
| GET /settings | ✅ | ❌ |
| PUT /settings | ✅ | ❌ |
| **ANALYTICS** | | |
| GET /analytics/overview | ✅ All | ✅ Own region |
| GET /analytics/merchants | ✅ All | ✅ Own region |
| GET /analytics/reviews | ✅ All | ✅ Own region |

---

## ENDPOINT INDEX

| # | Method | Path | Akses |
|---|--------|------|-------|
| **COMPANIES** | | | |
| 1 | GET | `/companies` | SA / AR |
| 2 | GET | `/companies/{id}` | SA / AR |
| 3 | POST | `/companies/{id}/ban` | SA |
| 4 | POST | `/companies/{id}/unban` | SA |
| 5 | DELETE | `/companies/{id}` | SA |
| **MERCHANTS** | | | |
| 6 | GET | `/merchants` | SA / AR |
| 7 | GET | `/merchants/{id}` | SA / AR |
| 8 | POST | `/merchants/{id}/approve` | SA / AR |
| 9 | POST | `/merchants/{id}/reject` | SA / AR |
| 10 | POST | `/merchants/{id}/ban` | SA / AR |
| 11 | POST | `/merchants/{id}/unban` | SA |
| 12 | POST | `/merchants/{id}/feature` | SA |
| 13 | DELETE | `/merchants/{id}` | SA |
| **ITEMS** | | | |
| 14 | GET | `/merchants/{id}/items` | SA / AR |
| 15 | DELETE | `/merchants/{id}/items/{item_id}` | SA / AR |
| **REVIEWS** | | | |
| 16 | GET | `/reviews` | SA / AR |
| 17 | GET | `/reviews/{id}` | SA / AR |
| 18 | POST | `/reviews/{id}/approve` | SA / AR |
| 19 | POST | `/reviews/{id}/reject` | SA / AR |
| 20 | POST | `/reviews/{id}/hide` | SA / AR |
| 21 | DELETE | `/reviews/{id}` | SA |
| **INQUIRIES** | | | |
| 22 | GET | `/inquiries` | SA / AR |
| 23 | GET | `/inquiries/{id}` | SA / AR |
| 24 | POST | `/inquiries/{id}/close` | SA / AR |
| **CATEGORIES** | | | |
| 25 | GET | `/categories` | SA / AR |
| 26 | POST | `/categories` | SA |
| 27 | PUT | `/categories/{id}` | SA |
| 28 | DELETE | `/categories/{id}` | SA |
| 29 | PUT | `/categories/reorder` | SA |
| **SETTINGS** | | | |
| 30 | GET | `/settings` | SA |
| 31 | PUT | `/settings` | SA |
| **ANALYTICS** | | | |
| 32 | GET | `/analytics/overview` | SA / AR |
| 33 | GET | `/analytics/merchants` | SA / AR |
| 34 | GET | `/analytics/reviews` | SA / AR |

---

## COMPANIES

### 1. List All Companies

- **GET** `/api/v1/web/directory/companies`
- **Auth**: Superadmin / Admin Regional

**Query Params:**
| Param | Type | Deskripsi |
|-------|------|-----------|
| search | string | Search nama company / owner name |
| region_id | UUID | Filter by region (Superadmin only) |
| status | string | active/inactive/banned |
| has_banned_merchant | boolean | Filter yang punya merchant banned |
| sort | string | newest/name_asc/merchant_count_desc |
| limit | int | Default 20 |
| offset | int | Default 0 |

*Admin Regional: `region_id` filter otomatis sesuai region sendiri*

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-cmp-001",
      "owner_id": "uuid-user-001",
      "owner_name": "Andi Santoso",
      "owner_email": "andi@example.com",
      "owner_phone": "0812345678",
      "name": "PT Impor Korea",
      "logo_url": "https://cdn.../logo.jpg",
      "phone": "0212345678",
      "email": "contact@ptkore.com",
      "website": "https://ptkore.com",
      "merchant_count": 3,
      "active_merchant_count": 2,
      "status": "active",
      "created_at": "2026-01-15T00:00:00Z",
      "updated_at": "2026-05-20T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 87 }
}
```

---

### 2. Get Company Detail

- **GET** `/api/v1/web/directory/companies/{company_id}`
- **Auth**: Superadmin / Admin Regional

**Response 200:**
```json
{
  "data": {
    "id": "uuid-cmp-001",
    "owner": {
      "id": "uuid-user-001",
      "name": "Andi Santoso",
      "email": "andi@example.com",
      "phone": "0812345678",
      "member_since": "2026-01-01T00:00:00Z"
    },
    "name": "PT Impor Korea",
    "description": "Distributor produk Korea...",
    "logo_url": "https://cdn.../logo.jpg",
    "phone": "0212345678",
    "email": "contact@ptkore.com",
    "website": "https://ptkore.com",
    "status": "active",
    "ban_reason": null,
    "banned_at": null,
    "merchants": [
      {
        "id": "uuid-mch-001",
        "name": "Toko Jakarta Pusat",
        "type": "retail",
        "status": "published",
        "city": "Jakarta Pusat",
        "rating": 4.5
      }
    ],
    "created_at": "2026-01-15T00:00:00Z"
  }
}
```

---

### 3. Ban Company

- **POST** `/api/v1/web/directory/companies/{company_id}/ban`
- **Auth**: Superadmin only

**Request Body:**
```json
{ "reason": "Melanggar ketentuan: menjual produk palsu" }
```

**Response 200:**
```json
{ "message": "Company banned. All merchants hidden from directory." }
```

---

### 4. Unban Company

- **POST** `/api/v1/web/directory/companies/{company_id}/unban`
- **Auth**: Superadmin only

**Response 200:**
```json
{ "message": "Company unbanned. Review merchant statuses individually." }
```

---

### 5. Force Delete Company

- **DELETE** `/api/v1/web/directory/companies/{company_id}`
- **Auth**: Superadmin only
- **Warning**: Cascade delete semua merchant, items, reviews, inquiries, favorites

**Request Body:**
```json
{ "confirm": true, "reason": "Permintaan penghapusan dari owner" }
```

**Response 200:**
```json
{ "message": "Company and all related data permanently deleted" }
```

---

## MERCHANTS

### 6. List All Merchants

- **GET** `/api/v1/web/directory/merchants`
- **Auth**: Superadmin / Admin Regional

**Query Params:**
| Param | Type | Deskripsi |
|-------|------|-----------|
| search | string | Search nama merchant / owner |
| company_id | UUID | Filter by company |
| region_id | UUID | Filter by region (SA only) |
| category_id | UUID | Filter by kategori |
| type | string | retail/online/service/food_beverage/beauty/other |
| status | string | draft/pending_approval/published/rejected/archived/banned |
| approval_status | string | pending/approved/rejected |
| city | string | Filter by kota |
| is_featured | boolean | |
| has_pending | boolean | Tampilkan yang pending_approval saja (untuk quick filter) |
| sort | string | newest/rating_desc/name_asc/pending_first |
| limit | int | Default 20 |
| offset | int | Default 0 |

*Admin Regional: Otomatis filter hanya region sendiri*

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-mch-001",
      "company_name": "PT Impor Korea",
      "owner_name": "Andi Santoso",
      "owner_email": "andi@example.com",
      "name": "Toko Jakarta Pusat",
      "type": "retail",
      "categories": [{ "name": "Retail", "icon": "🏬" }],
      "city": "Jakarta Pusat",
      "province": "DKI Jakarta",
      "region_id": "uuid-region-jakarta",
      "region_name": "DKI Jakarta",
      "status": "pending_approval",
      "approval_status": "pending",
      "is_featured": false,
      "rating": 0,
      "review_count": 0,
      "item_count": 3,
      "primary_image": "https://cdn.../m1.jpg",
      "created_at": "2026-05-30T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 312 },
  "meta": {
    "pending_count": 15,
    "published_count": 287,
    "banned_count": 10
  }
}
```

---

### 7. Get Merchant Detail (Admin View)

- **GET** `/api/v1/web/directory/merchants/{merchant_id}`
- **Auth**: Superadmin / Admin Regional

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "company": {
      "id": "uuid-cmp-001",
      "name": "PT Impor Korea",
      "status": "active"
    },
    "owner": {
      "id": "uuid-user-001",
      "name": "Andi Santoso",
      "email": "andi@example.com",
      "phone": "0812345678"
    },
    "name": "Toko Jakarta Pusat",
    "description": "Outlet utama...",
    "type": "retail",
    "categories": [
      { "id": "uuid-cat-retail", "name": "Retail", "icon": "🏬", "is_primary": true }
    ],
    "images": [
      { "url": "https://cdn.../m1.jpg", "is_primary": true }
    ],
    "location": {
      "address": "Jl. Ahmad Yani No. 123",
      "city": "Jakarta Pusat",
      "province": "DKI Jakarta",
      "latitude": -6.2088,
      "longitude": 106.8905,
      "region_id": "uuid-region-jakarta",
      "region_name": "DKI Jakarta"
    },
    "contact": {
      "phone": "0212345678",
      "email": "jakarta@ptkore.com",
      "whatsapp": "0812345678",
      "instagram": "@ptkore_jakarta"
    },
    "hours": { ... },
    "status": "pending_approval",
    "approval_status": "pending",
    "rejection_reason": null,
    "ban_reason": null,
    "banned_by_name": null,
    "approved_by_name": null,
    "stats": {
      "item_count": 3,
      "review_count": 0,
      "rating": 0,
      "favorite_count": 0,
      "inquiry_count": 0,
      "view_count": 12
    },
    "is_featured": false,
    "settings": {
      "allow_reviews": true,
      "allow_inquiries": true
    },
    "created_at": "2026-05-30T10:00:00Z",
    "updated_at": "2026-05-30T10:00:00Z"
  }
}
```

---

### 8. Approve Merchant

- **POST** `/api/v1/web/directory/merchants/{merchant_id}/approve`
- **Auth**: Superadmin / Admin Regional (own region)

**Request Body:** *(optional)*
```json
{ "note": "Semua dokumen sudah lengkap" }
```

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "status": "published",
    "approval_status": "approved",
    "approved_at": "2026-05-30T11:00:00Z",
    "approved_by": "uuid-admin-001"
  },
  "message": "Merchant approved and published successfully"
}
```

---

### 9. Reject Merchant

- **POST** `/api/v1/web/directory/merchants/{merchant_id}/reject`
- **Auth**: Superadmin / Admin Regional (own region)

**Request Body:**
```json
{
  "reason": "Foto merchant tidak jelas. Mohon upload foto yang lebih terang dan menampilkan nama toko.",
  "note": "Owner sudah dihubungi via WA"
}
```

**Validasi:** `reason` required, min 10 char.

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "status": "rejected",
    "approval_status": "rejected",
    "rejection_reason": "Foto merchant tidak jelas..."
  },
  "message": "Merchant rejected. Owner will be notified."
}
```

---

### 10. Ban Merchant

- **POST** `/api/v1/web/directory/merchants/{merchant_id}/ban`
- **Auth**: Superadmin / Admin Regional (own region)

**Request Body:**
```json
{
  "reason": "Merchant terbukti menjual produk palsu dan menyesatkan konsumen",
  "notify_owner": true
}
```

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "status": "banned",
    "banned_at": "2026-05-30T11:00:00Z"
  },
  "message": "Merchant banned successfully"
}
```

---

### 11. Unban Merchant

- **POST** `/api/v1/web/directory/merchants/{merchant_id}/unban`
- **Auth**: Superadmin only

**Request Body:**
```json
{ "note": "Owner sudah memberikan klarifikasi dan bukti produk asli" }
```

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "status": "draft"
  },
  "message": "Merchant unbanned. Status set to draft, owner can republish."
}
```

---

### 12. Toggle Featured Merchant

- **POST** `/api/v1/web/directory/merchants/{merchant_id}/feature`
- **Auth**: Superadmin only

**Request Body:**
```json
{
  "is_featured": true,
  "featured_until": "2026-06-30T23:59:59Z"
}
```

**Response 200:**
```json
{
  "data": {
    "id": "uuid-mch-001",
    "is_featured": true,
    "featured_until": "2026-06-30T23:59:59Z"
  },
  "message": "Merchant featured status updated"
}
```

---

### 13. Force Delete Merchant

- **DELETE** `/api/v1/web/directory/merchants/{merchant_id}`
- **Auth**: Superadmin only
- **Warning**: Hard delete semua data terkait (items, reviews, inquiries, favorites)

**Request Body:**
```json
{ "confirm": true, "reason": "Duplikasi merchant dari owner yang sama" }
```

**Response 200:**
```json
{ "message": "Merchant and all related data permanently deleted" }
```

---

## ITEMS

### 14. List Merchant Items (Admin View)

- **GET** `/api/v1/web/directory/merchants/{merchant_id}/items`
- **Auth**: Superadmin / Admin Regional

**Query Params:** `type` (product/service), `status`, `limit`, `offset`

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-item-001",
      "type": "product",
      "name": "Korean Red Ginseng",
      "category": "Health & Wellness",
      "price": 150000,
      "currency": "IDR",
      "unit": "per pack",
      "stock": null,
      "status": "available",
      "image_count": 2,
      "primary_image": "https://cdn.../p1.jpg",
      "sort_order": 1,
      "created_at": "2026-03-10T00:00:00Z"
    },
    {
      "id": "uuid-item-002",
      "type": "service",
      "name": "Beauty Treatment",
      "category": "Beauty",
      "price_min": 200000,
      "price_max": 500000,
      "currency": "IDR",
      "duration_minutes": 60,
      "status": "available",
      "image_count": 1,
      "primary_image": "https://cdn.../s1.jpg",
      "sort_order": 2,
      "created_at": "2026-03-15T00:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 53 }
}
```

---

### 15. Force Delete Item

- **DELETE** `/api/v1/web/directory/merchants/{merchant_id}/items/{item_id}`
- **Auth**: Superadmin / Admin Regional

**Response 204:** (No Content)

---

## REVIEWS

### 16. List All Reviews

- **GET** `/api/v1/web/directory/reviews`
- **Auth**: Superadmin / Admin Regional

**Query Params:**
| Param | Type | Deskripsi |
|-------|------|-----------|
| merchant_id | UUID | Filter by merchant |
| region_id | UUID | Filter by region (SA only) |
| status | string | pending/published/rejected/hidden |
| rating | int | Filter 1-5 |
| has_text | boolean | Hanya yang ada review text |
| sort | string | newest/oldest/rating_asc/rating_desc |
| limit | int | Default 20 |
| offset | int | Default 0 |

*Admin Regional: Otomatis filter region sendiri*

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-rev-001",
      "merchant_id": "uuid-mch-001",
      "merchant_name": "Toko Jakarta Pusat",
      "user_id": "uuid-user-123",
      "user_name": "Budi Santoso",
      "user_email": "budi@example.com",
      "rating": 4,
      "title": "Produk bagus",
      "review_text": "...",
      "aspects": { "product_quality": 5, "service": 4, "price": 4 },
      "helpful_count": 23,
      "status": "pending",
      "created_at": "2026-05-28T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 45 },
  "meta": { "pending_count": 12 }
}
```

---

### 17. Get Review Detail

- **GET** `/api/v1/web/directory/reviews/{review_id}`
- **Auth**: Superadmin / Admin Regional

**Response 200:**
```json
{
  "data": {
    "id": "uuid-rev-001",
    "merchant": {
      "id": "uuid-mch-001",
      "name": "Toko Jakarta Pusat",
      "region_name": "DKI Jakarta"
    },
    "user": {
      "id": "uuid-user-123",
      "name": "Budi Santoso",
      "email": "budi@example.com"
    },
    "rating": 4,
    "title": "Produk bagus, tapi lama delivery",
    "review_text": "Produk original...",
    "aspects": { "product_quality": 5, "service": 4, "price": 4 },
    "helpful_count": 23,
    "unhelpful_count": 2,
    "status": "pending",
    "moderated_by": null,
    "moderation_note": null,
    "created_at": "2026-05-28T10:00:00Z"
  }
}
```

---

### 18. Approve Review

- **POST** `/api/v1/web/directory/reviews/{review_id}/approve`
- **Auth**: Superadmin / Admin Regional

**Request Body:** *(optional)*
```json
{ "note": "Review valid dan sesuai guidelines" }
```

**Response 200:**
```json
{
  "data": { "id": "uuid-rev-001", "status": "published" },
  "message": "Review approved and published"
}
```

---

### 19. Reject Review

- **POST** `/api/v1/web/directory/reviews/{review_id}/reject`
- **Auth**: Superadmin / Admin Regional

**Request Body:**
```json
{ "note": "Review mengandung kata-kata kasar dan tidak konstruktif" }
```

**Response 200:**
```json
{
  "data": { "id": "uuid-rev-001", "status": "rejected" },
  "message": "Review rejected"
}
```

---

### 20. Hide Review

- **POST** `/api/v1/web/directory/reviews/{review_id}/hide`
- **Auth**: Superadmin / Admin Regional
- **Deskripsi**: Sembunyikan dari publik tanpa delete (bisa di-approve kembali)

**Request Body:**
```json
{ "note": "Dikeluhkan merchant sebagai review tidak valid" }
```

**Response 200:**
```json
{
  "data": { "id": "uuid-rev-001", "status": "hidden" },
  "message": "Review hidden from public listing"
}
```

---

### 21. Delete Review (Force)

- **DELETE** `/api/v1/web/directory/reviews/{review_id}`
- **Auth**: Superadmin only

**Response 204:** (No Content)

---

## INQUIRIES

### 22. List All Inquiries

- **GET** `/api/v1/web/directory/inquiries`
- **Auth**: Superadmin / Admin Regional

**Query Params:** `merchant_id`, `status` (pending/replied/closed), `region_id`, `limit`, `offset`

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-inq-001",
      "merchant_name": "Toko Jakarta Pusat",
      "user_name": "Citra Dewi",
      "subject": "Apakah stok masih ada?",
      "status": "pending",
      "created_at": "2026-05-30T10:00:00Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 34 }
}
```

---

### 23. Get Inquiry Detail

- **GET** `/api/v1/web/directory/inquiries/{inquiry_id}`
- **Auth**: Superadmin / Admin Regional

**Response 200:**
```json
{
  "data": {
    "id": "uuid-inq-001",
    "merchant": { "id": "uuid-mch-001", "name": "Toko Jakarta Pusat" },
    "user": { "id": "uuid-user-123", "name": "Citra Dewi", "email": "citra@example.com" },
    "subject": "Apakah stok masih ada?",
    "message": "Halo, saya ingin menanyakan ketersediaan...",
    "merchant_reply": null,
    "replied_at": null,
    "status": "pending",
    "created_at": "2026-05-30T10:00:00Z"
  }
}
```

---

### 24. Force Close Inquiry

- **POST** `/api/v1/web/directory/inquiries/{inquiry_id}/close`
- **Auth**: Superadmin / Admin Regional

**Request Body:**
```json
{ "reason": "Ditutup admin: merchant tidak aktif" }
```

**Response 200:**
```json
{ "message": "Inquiry closed by admin" }
```

---

## CATEGORY MASTER

### 25. List Categories

- **GET** `/api/v1/web/directory/categories`
- **Auth**: Superadmin / Admin Regional (read only)

**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid-cat-retail",
      "name": "Retail",
      "description": "Toko penjualan produk retail",
      "icon": "🏬",
      "is_active": true,
      "sort_order": 1,
      "merchant_count": 87
    },
    {
      "id": "uuid-cat-food",
      "name": "Food & Beverage",
      "description": "Restaurant, cafe, food delivery",
      "icon": "🍽️",
      "is_active": true,
      "sort_order": 2,
      "merchant_count": 45
    }
  ]
}
```

---

### 26. Create Category

- **POST** `/api/v1/web/directory/categories`
- **Auth**: Superadmin only

**Request Body:**
```json
{
  "name": "Electronics",
  "description": "Electronic products and gadgets",
  "icon": "🔌",
  "sort_order": 12,
  "is_active": true
}
```

**Validasi:**
| Field | Rule |
|-------|------|
| name | Required, unique, max 100 |
| description | Optional, max 500 |
| icon | Optional, emoji, max 10 char |
| sort_order | Optional, int |

**Response 201:**
```json
{
  "data": {
    "id": "uuid-cat-new",
    "name": "Electronics",
    "icon": "🔌",
    "is_active": true,
    "sort_order": 12,
    "merchant_count": 0
  },
  "message": "Category created successfully"
}
```

---

### 27. Update Category

- **PUT** `/api/v1/web/directory/categories/{category_id}`
- **Auth**: Superadmin only

**Request Body:** *(partial update)*
```json
{
  "name": "Electronics & Gadgets",
  "icon": "📱",
  "is_active": true
}
```

**Response 200:**
```json
{
  "data": { ...updated category... },
  "message": "Category updated"
}
```

---

### 28. Delete Category

- **DELETE** `/api/v1/web/directory/categories/{category_id}`
- **Auth**: Superadmin only
- **Rule**: Tidak bisa delete jika masih ada merchant menggunakan kategori ini

**Response 204:** (No Content)

**Errors:**
- `409` — Kategori masih digunakan oleh {N} merchant. Deactivate saja.

---

### 29. Reorder Categories

- **PUT** `/api/v1/web/directory/categories/reorder`
- **Auth**: Superadmin only

**Request Body:**
```json
{
  "order": [
    { "category_id": "uuid-cat-retail", "sort_order": 1 },
    { "category_id": "uuid-cat-food",   "sort_order": 2 },
    { "category_id": "uuid-cat-beauty", "sort_order": 3 }
  ]
}
```

**Response 200:**
```json
{ "message": "Categories reordered successfully" }
```

---

## SETTINGS

### 30. Get Directory Settings

- **GET** `/api/v1/web/directory/settings`
- **Auth**: Superadmin only

**Response 200:**
```json
{
  "data": {
    "merchant": {
      "require_approval": true,
      "max_per_member": null,
      "min_items_before_publish": 0,
      "min_images_before_publish": 1
    },
    "item": {
      "require_approval": false,
      "max_per_merchant": null,
      "allow_free_price": false
    },
    "review": {
      "allow_reviews": true,
      "require_moderation": true,
      "waiting_period_days": 0,
      "require_purchase_verification": false
    },
    "inquiry": {
      "allow_inquiries": true,
      "auto_close_days": 30,
      "allow_auto_reply": true
    },
    "verification": {
      "require_phone": true,
      "require_identity": false
    },
    "features": {
      "allow_online_only": true,
      "allow_multiple_categories": true,
      "featured_merchant_enabled": false
    },
    "location": {
      "require_location_for_physical": true,
      "auto_derive_region": true,
      "enable_map_view": true
    },
    "updated_at": "2026-05-01T00:00:00Z",
    "updated_by_name": "Super Admin"
  }
}
```

---

### 31. Update Directory Settings

- **PUT** `/api/v1/web/directory/settings`
- **Auth**: Superadmin only
- **Note**: Partial update, hanya field yang dikirim yang diubah

**Request Body:**
```json
{
  "merchant": {
    "require_approval": false,
    "min_images_before_publish": 2
  },
  "review": {
    "require_moderation": false,
    "waiting_period_days": 7
  },
  "inquiry": {
    "auto_close_days": 14
  }
}
```

**Response 200:**
```json
{
  "data": { ...full updated settings... },
  "message": "Settings updated successfully"
}
```

---

## ANALYTICS

### 32. Overview Analytics

- **GET** `/api/v1/web/directory/analytics/overview`
- **Auth**: Superadmin / Admin Regional

**Query Params:** `region_id` (SA only), `period` (7d/30d/90d/all, default: 30d)

**Response 200:**
```json
{
  "data": {
    "period": "30d",
    "merchants": {
      "total": 312,
      "published": 287,
      "pending": 15,
      "rejected": 8,
      "banned": 2,
      "new_this_period": 45
    },
    "companies": {
      "total": 120,
      "active": 115,
      "banned": 5,
      "new_this_period": 18
    },
    "reviews": {
      "total": 1450,
      "published": 1380,
      "pending": 50,
      "rejected": 20,
      "avg_rating": 4.3,
      "new_this_period": 210
    },
    "inquiries": {
      "total": 890,
      "pending": 45,
      "replied": 720,
      "closed": 125,
      "new_this_period": 120
    },
    "top_categories": [
      { "name": "Retail", "merchant_count": 87, "percentage": 27.9 },
      { "name": "Food & Beverage", "merchant_count": 65, "percentage": 20.8 },
      { "name": "Beauty & Salon", "merchant_count": 52, "percentage": 16.7 }
    ],
    "merchant_by_type": {
      "retail": 87,
      "online": 72,
      "service": 55,
      "food_beverage": 65,
      "beauty": 52,
      "other": 31
    }
  }
}
```

---

### 33. Merchant Analytics

- **GET** `/api/v1/web/directory/analytics/merchants`
- **Auth**: Superadmin / Admin Regional

**Query Params:** `region_id`, `period`, `sort` (most_viewed/most_favorited/highest_rated/most_reviewed), `limit`, `offset`

**Response 200:**
```json
{
  "data": {
    "top_merchants": [
      {
        "id": "uuid-mch-001",
        "name": "Toko Jakarta Pusat",
        "city": "Jakarta Pusat",
        "type": "retail",
        "rating": 4.8,
        "review_count": 245,
        "favorite_count": 567,
        "view_count": 12430,
        "inquiry_count": 89,
        "published_at": "2026-01-15T00:00:00Z"
      }
    ],
    "growth": [
      { "date": "2026-05-01", "new_merchants": 3 },
      { "date": "2026-05-02", "new_merchants": 5 }
    ]
  }
}
```

---

### 34. Review Analytics

- **GET** `/api/v1/web/directory/analytics/reviews`
- **Auth**: Superadmin / Admin Regional

**Query Params:** `region_id`, `period`

**Response 200:**
```json
{
  "data": {
    "total_reviews": 1450,
    "avg_rating": 4.3,
    "rating_distribution": {
      "5": 620,
      "4": 450,
      "3": 230,
      "2": 100,
      "1": 50
    },
    "moderation_queue": 50,
    "trend": [
      { "date": "2026-05-01", "count": 12, "avg_rating": 4.2 },
      { "date": "2026-05-02", "count": 18, "avg_rating": 4.5 }
    ],
    "top_reviewed_merchants": [
      {
        "merchant_id": "uuid-mch-001",
        "merchant_name": "Toko Jakarta Pusat",
        "review_count": 245,
        "avg_rating": 4.8
      }
    ]
  }
}
```

---

## Error Responses

```json
// 401
{ "message": "Authentication required" }

// 403
{ "message": "Insufficient permissions" }
{ "message": "Admin Regional cannot access other regions" }
{ "message": "Only Superadmin can perform this action" }

// 404
{ "message": "Merchant not found" }
{ "message": "Company not found" }
{ "message": "Review not found" }

// 409
{ "message": "Category is still used by 87 merchants. Deactivate instead of deleting." }
{ "message": "Company still has active merchants. Remove merchants first or use force delete." }

// 422
{
  "message": "Validation failed",
  "errors": {
    "reason": ["Rejection reason is required"],
    "confirm": ["Must confirm force delete with confirm: true"]
  }
}
```

---

*API Spec Directory Module — Web Backoffice v2.0. KAI App. Last updated: 2026-05-30*
