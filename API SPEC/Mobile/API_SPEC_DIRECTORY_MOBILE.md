# API Spec — Directory Module (Mobile Client)

Dokumentasi API endpoint untuk Directory module di aplikasi mobile (Flutter).

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/directory`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (For authenticated endpoints)
- **Authentication**: Required untuk create/edit, optional untuk view
- **Error Format**: Standard message atau validation error

---

## 1. Upload Media

Sebelum create company/merchant/item, upload images terlebih dahulu.

- **URL**: `POST /api/v1/mobile/media/upload`
- **Autentikasi**: Yes (Member Pro)
- **Content-Type**: `multipart/form-data`

- **Request**:
  ```
  files[]: image1.jpg  (max 5 files, max 5MB per file)
  files[]: image2.jpg
  context: "directory"
  ```

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "urls": [
        "https://cdn.example.com/directory/img_uuid1.jpg",
        "https://cdn.example.com/directory/img_uuid2.jpg"
      ]
    },
    "message": "2 image(s) uploaded successfully"
  }
  ```

---

## 2. Company Endpoints

### 2.1 Create Company

- **URL**: `POST /api/v1/mobile/directory/company`
- **Autentikasi**: Yes (Member Pro)

- **Request Body**:
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

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "cmp_001",
      "owner_id": "user_pro_001",
      "name": "PT Impor Korea",
      "logo_url": "https://cdn.../logo.jpg",
      "merchant_count": 0,
      "status": "active",
      "created_at": "2026-05-28T10:00:00Z"
    },
    "message": "Company created successfully"
  }
  ```

### 2.2 Get My Companies

- **URL**: `GET /api/v1/mobile/directory/company`
- **Autentikasi**: Yes (Member Pro)
- **Query Parameters**: `limit`, `offset`

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "cmp_001",
        "name": "PT Impor Korea",
        "logo_url": "https://cdn.../logo.jpg",
        "merchant_count": 3,
        "status": "active",
        "created_at": "2026-05-28T10:00:00Z"
      }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 1 }
  }
  ```

### 2.3 Get Company Detail

- **URL**: `GET /api/v1/mobile/directory/company/{company_id}`
- **Autentikasi**: Yes (Member Pro)

### 2.4 Update Company

- **URL**: `PUT /api/v1/mobile/directory/company/{company_id}`
- **Autentikasi**: Yes (Member Pro, owner only)

---

## 3. Merchant Endpoints

### 3.1 Create Merchant

- **URL**: `POST /api/v1/mobile/directory/merchant`
- **Autentikasi**: Yes (Member Pro)

- **Request Body (Retail/Physical)**:
  ```json
  {
    "company_id": "cmp_001",
    "name": "Toko Jakarta Pusat",
    "description": "Outlet utama kami...",
    "type": "retail",
    "categories": ["cat_retail", "cat_food"],
    "location": {
      "address": "Jl. Ahmad Yani 123, Jakarta Pusat",
      "city": "Jakarta",
      "province": "DKI Jakarta",
      "latitude": -6.2088,
      "longitude": 106.8905
    },
    "contact": { "phone": "0212345678", "email": "jakarta@ptkore.com" },
    "hours": { "monday": { "open": "09:00", "close": "18:00" } },
    "images": ["https://cdn.../merchant_1.jpg"]
  }
  ```

- **Request Body (Online-only)**:
  ```json
  {
    "company_id": "cmp_001",
    "name": "Online Store",
    "type": "online",
    "categories": ["cat_retail"],
    "contact": { "email": "online@ptkore.com", "whatsapp": "0812345678" },
    "images": ["https://cdn.../store_logo.jpg"]
  }
  ```

### 3.2 Get My Merchants

- **URL**: `GET /api/v1/mobile/directory/merchant`
- **Autentikasi**: Yes (Member Pro)

### 3.3 Get Merchant Detail

- **URL**: `GET /api/v1/mobile/directory/merchant/{merchant_id}`
- **Autentikasi**: Optional

### 3.4 Update Merchant

- **URL**: `PUT /api/v1/mobile/directory/merchant/{merchant_id}`
- **Autentikasi**: Yes (Member Pro, owner only)

---

## 4. Item/Offering Endpoints (Product + Service Merged)

### 4.1 Create Item

- **URL**: `POST /api/v1/mobile/directory/merchant/{merchant_id}/item`
- **Autentikasi**: Yes (Member Pro, merchant owner)

- **Request Body (Product)**:
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
    "images": ["https://cdn.../product_1.jpg"]
  }
  ```

- **Request Body (Service)**:
  ```json
  {
    "type": "service",
    "name": "Korean Beauty Treatment",
    "description": "Authentic Korean skincare treatment...",
    "category": "Beauty Treatment",
    "price_range": { "min": 200000, "max": 500000, "currency": "IDR" },
    "duration_minutes": 60,
    "images": ["https://cdn.../service_1.jpg"]
  }
  ```

### 4.2 Get Merchant Items

- **URL**: `GET /api/v1/mobile/directory/merchant/{merchant_id}/item`
- **Autentikasi**: Optional
- **Query Parameters**: `type` (filter), `limit`, `offset`

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
        "currency": "IDR",
        "unit": "per pack",
        "stock": null,
        "status": "available"
      },
      {
        "id": "item_002",
        "type": "service",
        "name": "Korean Beauty Treatment",
        "category": "Beauty Treatment",
        "price_range": { "min": 200000, "max": 500000 },
        "duration_minutes": 60,
        "status": "available"
      }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 53 }
  }
  ```

### 4.3 Update Item

- **URL**: `PUT /api/v1/mobile/directory/merchant/{merchant_id}/item/{item_id}`
- **Autentikasi**: Yes (Member Pro, merchant owner)

### 4.4 Delete Item

- **URL**: `DELETE /api/v1/mobile/directory/merchant/{merchant_id}/item/{item_id}`
- **Autentikasi**: Yes (Member Pro, merchant owner)

---

## 5. Browse Directory

### 5.1 List All Merchants (Public)

- **URL**: `GET /api/v1/mobile/directory`
- **Autentikasi**: Optional

- **Query Parameters**:
  - `search`: Search by name
  - `category`: Filter by category
  - `city`: Filter by city
  - `rating`: Min rating (1-5)
  - `type`: Merchant type
  - `sort`: "newest", "rating", "popular"
  - `limit`, `offset`

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "mch_001",
        "name": "Toko Jakarta Pusat",
        "type": "retail",
        "image": "https://cdn.../merchant_1.jpg",
        "categories": [{ "id": "cat_retail", "name": "Retail" }],
        "city": "Jakarta",
        "rating": 4.5,
        "review_count": 123,
        "favorite_count": 234,
        "item_count": 53,
        "is_saved": false
      }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 456 }
  }
  ```

---

## 6. Reviews

### 6.1 Leave Review

- **URL**: `POST /api/v1/mobile/directory/merchant/{merchant_id}/review`
- **Autentikasi**: Yes (Any member)

- **Request Body**:
  ```json
  {
    "rating": 4,
    "title": "Produk bagus, tapi lama delivery",
    "review_text": "...",
    "aspects": {
      "product_quality": 5,
      "service": 4,
      "price": 4
    }
  }
  ```

### 6.2 Get Merchant Reviews

- **URL**: `GET /api/v1/mobile/directory/merchant/{merchant_id}/review`
- **Autentikasi**: Optional

---

## 7. Favorites/Wishlist

### 7.1 Save Merchant

- **URL**: `POST /api/v1/mobile/directory/merchant/{merchant_id}/save`
- **Autentikasi**: Yes (Any member)

### 7.2 Unsave Merchant

- **URL**: `DELETE /api/v1/mobile/directory/merchant/{merchant_id}/save`
- **Autentikasi**: Yes (Any member)

### 7.3 Get My Saved Merchants

- **URL**: `GET /api/v1/mobile/directory/me/saved`
- **Autentikasi**: Yes (Any member)

---

## 8. Inquiries

### 8.1 Send Inquiry (Customer)

- **URL**: `POST /api/v1/mobile/directory/merchant/{merchant_id}/inquiry`
- **Autentikasi**: Yes (Any member)

- **Request Body**:
  ```json
  {
    "subject": "Inquiry: Stok Korean Ginseng",
    "message": "Apakah masih ada stok?"
  }
  ```

### 8.2 Get My Inquiries (Customer)

- **URL**: `GET /api/v1/mobile/directory/me/inquiry`
- **Autentikasi**: Yes (Any member)

### 8.3 Get Merchant Inquiries (Merchant Owner)

- **URL**: `GET /api/v1/mobile/directory/merchant/{merchant_id}/inquiry`
- **Autentikasi**: Yes (Merchant owner)

### 8.4 Reply to Inquiry (Merchant Owner)

- **URL**: `POST /api/v1/mobile/directory/merchant/{merchant_id}/inquiry/{inquiry_id}/reply`
- **Autentikasi**: Yes (Merchant owner)

- **Request Body**:
  ```json
  {
    "reply": "Ya, masih ada stok..."
  }
  ```

---

## 9. Categories

### 9.1 Get Categories

- **URL**: `GET /api/v1/mobile/directory/categories`
- **Autentikasi**: Optional

- **Response (Success 200)**:
  ```json
  {
    "data": [
      { "id": "cat_retail", "name": "Retail", "icon": "🏬" },
      { "id": "cat_food", "name": "Food & Beverage", "icon": "🍽️" },
      { "id": "cat_service", "name": "Service", "icon": "🔧" }
    ]
  }
  ```

---

## Error Handling

Standard error responses:

```json
// 401 Unauthorized
{ "message": "Authentication required" }

// 403 Forbidden
{ "message": "Only Member Pro can create merchants" }

// 404 Not Found
{ "message": "Merchant not found" }

// 422 Unprocessable Entity
{
  "message": "Validation failed",
  "errors": {
    "name": ["Merchant name is required"],
    "categories": ["At least 1 category is required"]
  }
}
```

---

*API Spec Directory Module - Mobile Client. Product & Service merged. Region_id optional.*
