# API Spec — Directory Module (Web Backoffice Admin)

Dokumentasi API endpoint untuk Directory management di backoffice dashboard (Superadmin & Admin Regional).

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/directory`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin / Admin Regional)

---

## ROLE-BASED ACCESS

| Endpoint | Superadmin | Admin Regional |
|----------|-----------|-----------------|
| GET /companies | ✅ (all) | ✅ (own region) |
| GET /merchants | ✅ (all) | ✅ (own region) |
| POST /merchants/:id/approve | ✅ | ✅ (own region) |
| POST /merchants/:id/reject | ✅ | ✅ (own region) |
| POST /merchants/:id/ban | ✅ | ✅ (own region) |
| DELETE /merchants/:id | ✅ | ❌ |
| GET /settings | ✅ | ❌ |
| PUT /settings | ✅ | ❌ |
| GET /categories | ✅ | ❌ |
| POST /categories | ✅ | ❌ |
| GET /analytics | ✅ | ✅ (own region) |

---

## 1. Company Management

### 1.1 List All Companies

- **URL**: `GET /api/v1/web/directory/companies`
- **Autentikasi**: Yes (Superadmin / Admin Regional)
- **Query Parameters**: `search`, `region_id`, `status`, `limit`, `offset`

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "cmp_001",
        "owner_id": "user_pro_001",
        "owner_name": "Andi Santoso",
        "owner_email": "andi@example.com",
        "name": "PT Impor Korea",
        "logo_url": "https://cdn.../logo.jpg",
        "phone": "0212345678",
        "email": "contact@ptkore.com",
        "merchant_count": 3,
        "status": "active",
        "created_at": "2026-01-15T00:00:00Z"
      }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 45 }
  }
  ```

### 1.2 Get Company Detail

- **URL**: `GET /api/v1/web/directory/companies/{company_id}`
- **Autentikasi**: Yes

---

## 2. Merchant Management

### 2.1 List All Merchants

- **URL**: `GET /api/v1/web/directory/merchants`
- **Autentikasi**: Yes (Superadmin / Admin Regional)

- **Query Parameters**:
  - `search`, `company_id`, `region_id`, `status`, `approval_status`
  - `type`: retail, online, service, food_beverage, beauty
  - `category`, `sort`, `limit`, `offset`

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "mch_001",
        "company_id": "cmp_001",
        "owner_name": "Andi Santoso",
        "name": "Toko Jakarta Pusat",
        "type": "retail",
        "categories": ["Retail", "Food & Beverage"],
        "city": "Jakarta",
        "region_id": "region_jakarta",
        "status": "published",
        "approval_status": "approved",
        "item_count": 53,
        "rating": 4.5,
        "review_count": 123,
        "favorite_count": 234,
        "total_views": 5420,
        "created_at": "2026-02-10T00:00:00Z",
        "published_at": "2026-02-12T10:00:00Z"
      }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 234 },
    "summary": {
      "total_pending": 5,
      "total_published": 200,
      "total_rejected": 12,
      "total_banned": 8
    }
  }
  ```

### 2.2 Get Pending Merchants

- **URL**: `GET /api/v1/web/directory/merchants/pending`
- **Autentikasi**: Yes

### 2.3 Get Merchant Detail

- **URL**: `GET /api/v1/web/directory/merchants/{merchant_id}`
- **Autentikasi**: Yes

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "mch_001",
      "company_id": "cmp_001",
      "owner_id": "user_pro_001",
      "name": "Toko Jakarta Pusat",
      "description": "Outlet utama kami di Jakarta...",
      "type": "retail",
      "categories": ["Retail", "Food & Beverage"],
      "location": {
        "address": "Jl. Ahmad Yani 123, Jakarta Pusat",
        "city": "Jakarta",
        "province": "DKI Jakarta",
        "latitude": -6.2088,
        "longitude": 106.8905,
        "region_id": "region_jakarta"
      },
      "contact": {
        "phone": "0212345678",
        "email": "jakarta@ptkore.com",
        "whatsapp": "0812345678",
        "instagram": "@ptkore_jakarta"
      },
      "hours": {
        "monday": { "open": "09:00", "close": "18:00" },
        "tuesday": { "open": "09:00", "close": "18:00" }
      },
      "stats": {
        "item_count": 53,
        "review_count": 123,
        "rating": 4.5,
        "favorite_count": 234,
        "inquiry_count": 45,
        "total_views": 5420
      },
      "status": "published",
      "approval_status": "approved",
      "approved_by": "admin_001",
      "approved_at": "2026-02-12T10:00:00Z"
    }
  }
  ```

### 2.4 Approve Merchant

- **URL**: `POST /api/v1/web/directory/merchants/{merchant_id}/approve`
- **Autentikasi**: Yes (Superadmin / Admin Regional)

- **Request Body**:
  ```json
  {
    "notes": "All info complete and verified"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "mch_001",
      "status": "published",
      "approval_status": "approved",
      "approved_at": "2026-05-28T10:00:00Z"
    },
    "message": "Merchant approved and published successfully"
  }
  ```

### 2.5 Reject Merchant

- **URL**: `POST /api/v1/web/directory/merchants/{merchant_id}/reject`
- **Autentikasi**: Yes

- **Request Body**:
  ```json
  {
    "reason": "Phone number tidak valid",
    "rejection_details": "Phone number provided does not match with company registration"
  }
  ```

### 2.6 Ban Merchant

- **URL**: `POST /api/v1/web/directory/merchants/{merchant_id}/ban`
- **Autentikasi**: Yes

- **Request Body**:
  ```json
  {
    "reason": "Violation of community guidelines",
    "notes": "Multiple fake reviews and spam inquiries"
  }
  ```

### 2.7 Delete Merchant (Force)

- **URL**: `DELETE /api/v1/web/directory/merchants/{merchant_id}`
- **Autentikasi**: Yes (Superadmin only)

---

## 3. Items Management

### 3.1 List Merchant Items

- **URL**: `GET /api/v1/web/directory/merchants/{merchant_id}/items`
- **Autentikasi**: Yes

- **Query Parameters**: `type` (product/service), `status`, `limit`, `offset`

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "item_001",
        "type": "product",
        "name": "Korean Red Ginseng",
        "category": "Health & Wellness",
        "price": 150000,
        "unit": "per pack",
        "stock": null,
        "status": "available",
        "image_count": 2,
        "created_at": "2026-03-10T00:00:00Z"
      },
      {
        "id": "item_002",
        "type": "service",
        "name": "Beauty Treatment",
        "category": "Beauty",
        "price_range": { "min": 200000, "max": 500000 },
        "duration_minutes": 60,
        "status": "available",
        "image_count": 1,
        "created_at": "2026-03-15T00:00:00Z"
      }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 53 }
  }
  ```

### 3.2 Delete Item

- **URL**: `DELETE /api/v1/web/directory/merchants/{merchant_id}/items/{item_id}`
- **Autentikasi**: Yes (Superadmin / Admin Regional)

---

## 4. Category Master

### 4.1 List Categories

- **URL**: `GET /api/v1/web/directory/categories`
- **Autentikasi**: Yes (Superadmin)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "cat_retail",
        "name": "Retail",
        "description": "Toko penjualan produk retail",
        "icon": "🏬",
        "is_active": true,
        "order": 1
      },
      {
        "id": "cat_food",
        "name": "Food & Beverage",
        "description": "Restaurant, cafe, food delivery",
        "icon": "🍽️",
        "is_active": true,
        "order": 2
      }
    ]
  }
  ```

### 4.2 Create Category

- **URL**: `POST /api/v1/web/directory/categories`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "name": "Electronics",
    "description": "Electronic products and gadgets",
    "icon": "🔌",
    "order": 10
  }
  ```

### 4.3 Update Category

- **URL**: `PUT /api/v1/web/directory/categories/{category_id}`
- **Autentikasi**: Yes (Superadmin)

### 4.4 Delete Category

- **URL**: `DELETE /api/v1/web/directory/categories/{category_id}`
- **Autentikasi**: Yes (Superadmin)

---

## 5. Directory Settings

### 5.1 Get Settings

- **URL**: `GET /api/v1/web/directory/settings`
- **Autentikasi**: Yes (Superadmin)

- **Response (Success 200)**:
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
        "waiting_period_days": 0
      },
      "inquiry": {
        "allow_inquiries": true,
        "auto_close_days": 30
      },
      "verification": {
        "require_phone": true,
        "require_identity": false
      },
      "features": {
        "allow_online_only": true,
        "allow_multiple_categories": true
      },
      "location": {
        "require_location_for_physical": true,
        "auto_derive_region": true
      }
    }
  }
  ```

### 5.2 Update Settings

- **URL**: `PUT /api/v1/web/directory/settings`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "merchant": { "require_approval": false },
    "review": { "require_moderation": false }
  }
  ```

---

## 6. Analytics & Reports

### 6.1 Get Directory Analytics

- **URL**: `GET /api/v1/web/directory/analytics`
- **Autentikasi**: Yes (Superadmin / Admin Regional)

- **Query Parameters**: `date_from`, `date_to`, `period`

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "period": "2026-05",
      "kpis": {
        "total_companies": 45,
        "total_merchants": 234,
        "published_merchants": 200,
        "pending_merchants": 5,
        "banned_merchants": 8,
        "total_items": 12500,
        "product_count": 11000,
        "service_count": 1500,
        "total_reviews": 5234,
        "avg_rating": 4.3,
        "total_inquiries": 2341
      },
      "merchant_breakdown": {
        "by_type": {
          "retail": { "count": 100, "percentage": 42.7 },
          "online": { "count": 65, "percentage": 27.8 },
          "service": { "count": 40, "percentage": 17.1 },
          "food_beverage": { "count": 20, "percentage": 8.5 },
          "beauty": { "count": 9, "percentage": 3.9 }
        },
        "by_region": {
          "region_jakarta": { "count": 120, "percentage": 51.3 },
          "region_surabaya": { "count": 65, "percentage": 27.8 },
          "region_bandung": { "count": 49, "percentage": 20.9 }
        }
      },
      "item_breakdown": {
        "products": { "count": 11000, "percentage": 88 },
        "services": { "count": 1500, "percentage": 12 }
      },
      "activity": {
        "new_merchants_this_period": 23,
        "new_items_this_period": 456,
        "new_reviews_this_period": 234,
        "new_inquiries_this_period": 342
      },
      "approval_metrics": {
        "total_pending": 5,
        "avg_approval_time_hours": 8,
        "rejection_rate": 5.2
      }
    }
  }
  ```

### 6.2 Get Merchant Performance

- **URL**: `GET /api/v1/web/directory/merchants/{merchant_id}/performance`
- **Autentikasi**: Yes (Superadmin / Admin Regional)

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "merchant_id": "mch_001",
      "merchant_name": "Toko Jakarta Pusat",
      "period": "2026-05",
      "stats": {
        "views": 5420,
        "items": 53,
        "reviews": 123,
        "rating": 4.5,
        "favorites": 234,
        "inquiries": 45,
        "inquiry_response_rate": 95.6,
        "avg_response_time_hours": 2.4
      },
      "trends": {
        "views_trend": "+15%",
        "reviews_trend": "+12%",
        "inquiries_trend": "+8%"
      }
    }
  }
  ```

---

*API Spec Directory Module - Backoffice Admin Panel. Superadmin & Admin Regional only.*
