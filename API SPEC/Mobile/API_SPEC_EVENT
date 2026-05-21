# Dokumentasi API Spec - Modul Event (Mobile Client)

Dokumentasi ini dibuat untuk kebutuhan tim Backend agar skema request/response API sesuai dengan implementasi Clean Architecture pada Flutter Mobile Client.

## Informasi Umum

- **Base URL Prefix**: `/api/v1` (Seluruh endpoint di bawah ini menggunakan prefix `/api/v1/mobile/events`)
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Accept-Language: <lang_code>` (Mengirimkan kode bahasa aktif, contoh: `ko`, `id`, `en`. Default: `ko`)
  - `X-Locale: <lang_code>` (Mengirimkan kode bahasa aktif, contoh: `ko`, `id`, `en`. Default: `ko`)
  - `Authorization: Bearer <access_token>` (Optional - untuk event yang require authentication atau future RSVP feature)

---

## Model Data Utama

### 1. Event Object (Common Response Schema)
Setiap kali endpoint mengembalikan data Event, strukturnya harus berupa format berikut di dalam field `data`:

```json
{
  "id": "string",
  "title": "string",
  "slug": "string (URL-friendly identifier)",
  "description": "string (full description, support markdown/html)",
  "short_description": "string (nullable, max 200 char for list view)",
  "banner": "string (nullable / URL)",
  "thumbnail": "string (nullable / URL, smaller version for list)",
  "category": {
    "id": "string",
    "name": "string (e.g. 'Cultural', 'Sports', 'Networking', 'Educational')",
    "slug": "string"
  },
  "tags": ["string"] (array of tags, e.g. ['k-pop', 'concert', 'jakarta']),
  "organizer": {
    "id": "string",
    "name": "string (e.g. 'KAI Pusat', 'KAI Jakarta')",
    "type": "string (enum: 'kai_pusat', 'kai_region', 'community', 'external')",
    "avatar": "string (nullable / URL)"
  },
  "region": {
    "id": "string",
    "name": "string (e.g. 'KAI Pusat', 'KAI Jakarta')",
    "code": "string (e.g. 'central', 'jakarta')"
  },
  "location": {
    "name": "string (venue name)",
    "address": "string",
    "city": "string",
    "province": "string (nullable)",
    "country": "string",
    "latitude": "number (nullable)",
    "longitude": "number (nullable)",
    "maps_url": "string (nullable, Google Maps link)"
  },
  "start_date": "string (ISO 8601 UTC format)",
  "end_date": "string (ISO 8601 UTC format)",
  "timezone": "string (e.g. 'Asia/Jakarta')",
  "status": "string (enum: 'draft', 'published', 'cancelled', 'completed')",
  "registration_status": "string (enum: 'open', 'closed', 'full', 'not_required')",
  "capacity": "integer (nullable, total capacity)",
  "registered_count": "integer (current number of registrations)",
  "is_free": "boolean",
  "price": "number (nullable, if paid event)",
  "currency": "string (e.g. 'IDR', 'USD')",
  "registration_url": "string (nullable, external registration link)",
  "contact_email": "string (nullable)",
  "contact_phone": "string (nullable)",
  "is_featured": "boolean (highlighted event)",
  "views_count": "integer",
  "created_by": {
    "id": "string",
    "name": "string",
    "role": "string"
  },
  "created_at": "string (ISO 8601 UTC format)",
  "updated_at": "string (ISO 8601 UTC format)",
  "published_at": "string (nullable, ISO 8601 UTC format)"
}
```

### 2. Event List Item (Simplified for List View)
Untuk endpoint list, response bisa menggunakan simplified version untuk performance:

```json
{
  "id": "string",
  "title": "string",
  "slug": "string",
  "short_description": "string (nullable)",
  "thumbnail": "string (nullable / URL)",
  "category": {
    "id": "string",
    "name": "string"
  },
  "region": {
    "id": "string",
    "name": "string"
  },
  "location": {
    "name": "string",
    "city": "string"
  },
  "start_date": "string (ISO 8601 UTC format)",
  "end_date": "string (ISO 8601 UTC format)",
  "registration_status": "string",
  "is_free": "boolean",
  "price": "number (nullable)",
  "currency": "string",
  "is_featured": "boolean"
}
```

### 3. Event Category Object
Struktur untuk kategori event:

```json
{
  "id": "string",
  "name": "string",
  "slug": "string",
  "description": "string (nullable)",
  "icon": "string (nullable, icon name or URL)",
  "color": "string (nullable, hex color code)",
  "event_count": "integer (total events in this category)"
}
```

### 4. Pagination Metadata
```json
{
  "current_page": "integer",
  "total_pages": "integer",
  "total_items": "integer",
  "items_per_page": "integer",
  "has_next": "boolean",
  "has_prev": "boolean"
}
```

### 5. Error Responses
Frontend menangani 2 skema error utama dari backend:

#### A. Standard Message Error (HTTP 4xx / 5xx)
```json
{
  "message": "Pesan error deskriptif di sini"
}
```

#### B. Validation Error (HTTP 422 Unprocessable Entity)
```json
{
  "message": "Data input tidak valid",
  "errors": {
    "field_name": ["Error message 1", "Error message 2"]
  }
}
```

---

## Daftar Endpoint

### 1. Get Events List
Mengambil daftar event dengan support filtering, searching, sorting, dan pagination.

- **URL**: `GET /api/v1/mobile/events`
- **Autentikasi**: Optional (`Bearer <access_token>` untuk personalized content)
- **Query Parameters**:
  ```
  page              : integer (default: 1)
  limit             : integer (default: 10, max: 50)
  search            : string (search in title, description, location)
  category_id       : string (filter by category)
  region_id         : string (filter by region)
  status            : string (filter by status: 'published', 'cancelled', 'completed')
  registration_status : string (filter by: 'open', 'closed', 'full', 'not_required')
  is_free           : boolean (filter free events: true/false)
  is_featured       : boolean (filter featured events: true/false)
  date_from         : string (ISO 8601 date, filter events starting from this date)
  date_to           : string (ISO 8601 date, filter events until this date)
  city              : string (filter by city)
  tags              : string (comma-separated tags, e.g. 'k-pop,concert')
  sort              : string (sort field: 'start_date', 'created_at', 'views_count', 'title')
  order             : string (sort order: 'asc', 'desc', default: 'asc' for dates, 'desc' for counts)
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "evt_001",
        "title": "Korean Cultural Festival 2026",
        "slug": "korean-cultural-festival-2026",
        "short_description": "Perayaan budaya Korea dengan berbagai kegiatan menarik untuk seluruh keluarga.",
        "thumbnail": "https://example.com/events/evt_001_thumb.jpg",
        "category": {
          "id": "cat_001",
          "name": "Cultural"
        },
        "region": {
          "id": "reg_001",
          "name": "KAI Pusat"
        },
        "location": {
          "name": "Gelora Bung Karno",
          "city": "Jakarta"
        },
        "start_date": "2026-06-15T10:00:00.000Z",
        "end_date": "2026-06-15T18:00:00.000Z",
        "registration_status": "open",
        "is_free": true,
        "price": null,
        "currency": "IDR",
        "is_featured": true
      },
      {
        "id": "evt_002",
        "title": "K-Pop Dance Workshop",
        "slug": "kpop-dance-workshop",
        "short_description": "Workshop dance K-Pop untuk pemula hingga advanced level.",
        "thumbnail": "https://example.com/events/evt_002_thumb.jpg",
        "category": {
          "id": "cat_002",
          "name": "Educational"
        },
        "region": {
          "id": "reg_002",
          "name": "KAI Jakarta"
        },
        "location": {
          "name": "Studio Dance ABC",
          "city": "Jakarta"
        },
        "start_date": "2026-06-20T14:00:00.000Z",
        "end_date": "2026-06-20T17:00:00.000Z",
        "registration_status": "open",
        "is_free": false,
        "price": 150000,
        "currency": "IDR",
        "is_featured": false
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 5,
      "total_items": 48,
      "items_per_page": 10,
      "has_next": true,
      "has_prev": false
    }
  }
  ```

---

### 2. Get Event Detail
Mengambil detail lengkap sebuah event berdasarkan ID atau slug.

- **URL**: `GET /api/v1/mobile/events/:id_or_slug`
- **Autentikasi**: Optional (`Bearer <access_token>` untuk tracking views & personalized content)
- **Path Parameters**:
  - `id_or_slug`: Event ID (e.g. `evt_001`) atau slug (e.g. `korean-cultural-festival-2026`)
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "evt_001",
      "title": "Korean Cultural Festival 2026",
      "slug": "korean-cultural-festival-2026",
      "description": "Perayaan budaya Korea terbesar tahun ini! Nikmati berbagai pertunjukan musik, tarian tradisional, pameran budaya, dan kuliner Korea. Event ini terbuka untuk umum dan gratis.\n\n**Agenda:**\n- 10:00 - Opening Ceremony\n- 11:00 - Traditional Dance Performance\n- 13:00 - K-Pop Cover Dance Competition\n- 15:00 - Korean Food Festival\n- 17:00 - Closing & Door Prize",
      "short_description": "Perayaan budaya Korea dengan berbagai kegiatan menarik untuk seluruh keluarga.",
      "banner": "https://example.com/events/evt_001_banner.jpg",
      "thumbnail": "https://example.com/events/evt_001_thumb.jpg",
      "category": {
        "id": "cat_001",
        "name": "Cultural",
        "slug": "cultural"
      },
      "tags": ["korean-culture", "festival", "family-friendly", "jakarta"],
      "organizer": {
        "id": "org_001",
        "name": "KAI Pusat",
        "type": "kai_pusat",
        "avatar": "https://example.com/avatars/kai_pusat.jpg"
      },
      "region": {
        "id": "reg_001",
        "name": "KAI Pusat",
        "code": "central"
      },
      "location": {
        "name": "Gelora Bung Karno",
        "address": "Jl. Pintu Satu Senayan, Jakarta Pusat",
        "city": "Jakarta",
        "province": "DKI Jakarta",
        "country": "Indonesia",
        "latitude": -6.218485,
        "longitude": 106.802316,
        "maps_url": "https://maps.google.com/?q=-6.218485,106.802316"
      },
      "start_date": "2026-06-15T10:00:00.000Z",
      "end_date": "2026-06-15T18:00:00.000Z",
      "timezone": "Asia/Jakarta",
      "status": "published",
      "registration_status": "open",
      "capacity": 5000,
      "registered_count": 1247,
      "is_free": true,
      "price": null,
      "currency": "IDR",
      "registration_url": null,
      "contact_email": "info@kai-pusat.org",
      "contact_phone": "+62215551234",
      "is_featured": true,
      "views_count": 8432,
      "created_by": {
        "id": "usr_admin_001",
        "name": "Admin KAI Pusat",
        "role": "superadmin"
      },
      "created_at": "2026-04-01T08:00:00.000Z",
      "updated_at": "2026-05-20T10:30:00.000Z",
      "published_at": "2026-04-15T09:00:00.000Z"
    }
  }
  ```
- **Response (Error 404 - Not Found)**:
  ```json
  {
    "message": "Event tidak ditemukan"
  }
  ```

---

### 3. Get Event Categories
Mengambil daftar semua kategori event yang tersedia.

- **URL**: `GET /api/v1/mobile/events/categories`
- **Autentikasi**: Tidak
- **Query Parameters**:
  ```
  include_count  : boolean (include event count per category, default: false)
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "cat_001",
        "name": "Cultural",
        "slug": "cultural",
        "description": "Acara yang berkaitan dengan budaya Korea",
        "icon": "culture_icon",
        "color": "#FF6B6B",
        "event_count": 15
      },
      {
        "id": "cat_002",
        "name": "Educational",
        "slug": "educational",
        "description": "Workshop, seminar, dan kelas pembelajaran",
        "icon": "education_icon",
        "color": "#4ECDC4",
        "event_count": 23
      },
      {
        "id": "cat_003",
        "name": "Sports",
        "slug": "sports",
        "description": "Kegiatan olahraga dan kompetisi",
        "icon": "sports_icon",
        "color": "#95E1D3",
        "event_count": 12
      },
      {
        "id": "cat_004",
        "name": "Networking",
        "slug": "networking",
        "description": "Pertemuan bisnis dan networking",
        "icon": "networking_icon",
        "color": "#F38181",
        "event_count": 8
      }
    ]
  }
  ```

---

### 4. Get Featured Events
Mengambil daftar event yang di-highlight (featured).

- **URL**: `GET /api/v1/mobile/events/featured`
- **Autentikasi**: Optional
- **Query Parameters**:
  ```
  limit  : integer (default: 5, max: 20)
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "evt_001",
        "title": "Korean Cultural Festival 2026",
        "slug": "korean-cultural-festival-2026",
        "short_description": "Perayaan budaya Korea dengan berbagai kegiatan menarik untuk seluruh keluarga.",
        "thumbnail": "https://example.com/events/evt_001_thumb.jpg",
        "category": {
          "id": "cat_001",
          "name": "Cultural"
        },
        "region": {
          "id": "reg_001",
          "name": "KAI Pusat"
        },
        "location": {
          "name": "Gelora Bung Karno",
          "city": "Jakarta"
        },
        "start_date": "2026-06-15T10:00:00.000Z",
        "end_date": "2026-06-15T18:00:00.000Z",
        "registration_status": "open",
        "is_free": true,
        "price": null,
        "currency": "IDR",
        "is_featured": true
      }
    ]
  }
  ```

---

### 5. Get Upcoming Events
Mengambil daftar event yang akan datang (upcoming), sorted by start_date ascending.

- **URL**: `GET /api/v1/mobile/events/upcoming`
- **Autentikasi**: Optional
- **Query Parameters**:
  ```
  limit       : integer (default: 10, max: 50)
  region_id   : string (filter by region)
  category_id : string (filter by category)
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "evt_002",
        "title": "K-Pop Dance Workshop",
        "slug": "kpop-dance-workshop",
        "short_description": "Workshop dance K-Pop untuk pemula hingga advanced level.",
        "thumbnail": "https://example.com/events/evt_002_thumb.jpg",
        "category": {
          "id": "cat_002",
          "name": "Educational"
        },
        "region": {
          "id": "reg_002",
          "name": "KAI Jakarta"
        },
        "location": {
          "name": "Studio Dance ABC",
          "city": "Jakarta"
        },
        "start_date": "2026-06-20T14:00:00.000Z",
        "end_date": "2026-06-20T17:00:00.000Z",
        "registration_status": "open",
        "is_free": false,
        "price": 150000,
        "currency": "IDR",
        "is_featured": false
      }
    ]
  }
  ```

---

### 6. Search Events
Endpoint khusus untuk search dengan support advanced query.

- **URL**: `GET /api/v1/mobile/events/search`
- **Autentikasi**: Optional
- **Query Parameters**:
  ```
  q        : string (required, search query)
  page     : integer (default: 1)
  limit    : integer (default: 10, max: 50)
  filters  : string (optional, JSON string of additional filters)
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "evt_001",
        "title": "Korean Cultural Festival 2026",
        "slug": "korean-cultural-festival-2026",
        "short_description": "Perayaan budaya Korea dengan berbagai kegiatan menarik untuk seluruh keluarga.",
        "thumbnail": "https://example.com/events/evt_001_thumb.jpg",
        "category": {
          "id": "cat_001",
          "name": "Cultural"
        },
        "region": {
          "id": "reg_001",
          "name": "KAI Pusat"
        },
        "location": {
          "name": "Gelora Bung Karno",
          "city": "Jakarta"
        },
        "start_date": "2026-06-15T10:00:00.000Z",
        "end_date": "2026-06-15T18:00:00.000Z",
        "registration_status": "open",
        "is_free": true,
        "price": null,
        "currency": "IDR",
        "is_featured": true
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 1,
      "total_items": 3,
      "items_per_page": 10,
      "has_next": false,
      "has_prev": false
    },
    "search_meta": {
      "query": "korean festival",
      "total_results": 3,
      "search_time_ms": 45
    }
  }
  ```

---

## Future Feature: RSVP / Registration

> [!NOTE]
> **Placeholder untuk future implementation**. Endpoint-endpoint berikut belum final dan masih dalam perencanaan.

### 7. Register to Event (Future)
Mendaftarkan user ke sebuah event.

- **URL**: `POST /api/v1/mobile/events/:id/register`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "attendee_name": "string (optional, default to user's name)",
    "attendee_email": "string (optional, default to user's email)",
    "attendee_phone": "string (optional)",
    "notes": "string (nullable, additional notes from attendee)"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Anda berhasil terdaftar untuk event ini",
    "data": {
      "registration_id": "reg_12345",
      "event_id": "evt_001",
      "status": "confirmed",
      "registered_at": "2026-05-21T10:30:00.000Z"
    }
  }
  ```

### 8. Cancel Event Registration (Future)
Membatalkan pendaftaran user dari sebuah event.

- **URL**: `DELETE /api/v1/mobile/events/:id/register`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Response (Success 200)**:
  ```json
  {
    "message": "Pendaftaran berhasil dibatalkan"
  }
  ```

### 9. Get My Registered Events (Future)
Mengambil daftar event yang sudah didaftarkan oleh user.

- **URL**: `GET /api/v1/mobile/events/my-registrations`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Query Parameters**:
  ```
  page   : integer (default: 1)
  limit  : integer (default: 10, max: 50)
  status : string (filter by: 'upcoming', 'past', 'cancelled')
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "registration_id": "reg_12345",
        "event": {
          "id": "evt_001",
          "title": "Korean Cultural Festival 2026",
          "slug": "korean-cultural-festival-2026",
          "thumbnail": "https://example.com/events/evt_001_thumb.jpg",
          "start_date": "2026-06-15T10:00:00.000Z",
          "end_date": "2026-06-15T18:00:00.000Z",
          "location": {
            "name": "Gelora Bung Karno",
            "city": "Jakarta"
          }
        },
        "status": "confirmed",
        "registered_at": "2026-05-21T10:30:00.000Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 1,
      "total_items": 5,
      "items_per_page": 10,
      "has_next": false,
      "has_prev": false
    }
  }
  ```

---

## Filter & Sort Guidelines

### Common Filters
Berikut adalah kombinasi filter yang umum digunakan:

1. **Upcoming Free Events in Jakarta**
   ```
   GET /events?is_free=true&city=Jakarta&date_from=2026-05-21&sort=start_date&order=asc
   ```

2. **Featured Cultural Events**
   ```
   GET /events?is_featured=true&category_id=cat_001&sort=start_date&order=asc
   ```

3. **Events by Specific Region**
   ```
   GET /events?region_id=reg_002&status=published&sort=start_date&order=asc
   ```

4. **Search K-Pop Related Events**
   ```
   GET /events/search?q=k-pop&sort=start_date&order=asc
   ```

### Sort Options
| Sort Field | Description | Default Order |
|------------|-------------|---------------|
| `start_date` | Event start date | asc |
| `created_at` | Creation date | desc |
| `views_count` | Popularity by views | desc |
| `title` | Alphabetical by title | asc |

---

## Status Code Reference

| Code | Meaning |
|------|---------|
| `200` | Success - Request berhasil |
| `400` | Bad Request - Invalid query parameters |
| `401` | Unauthorized - Token invalid/expired (untuk endpoint yang require auth) |
| `404` | Not Found - Event tidak ditemukan |
| `422` | Unprocessable Entity - Validation error |
| `500` | Internal Server Error - Error di backend |

---

## Notes & Best Practices

1. **Pagination**: Selalu gunakan pagination untuk list endpoint. Default 10 items per page, max 50.

2. **Caching Strategy**:
   - `/events` (list) → Cache 5-10 menit
   - `/events/:id` (detail) → Cache 15 menit atau sampai event updated
   - `/events/featured` → Cache 30 menit
   - `/events/categories` → Cache 1 jam (jarang berubah)

3. **Image Optimization**: Gunakan thumbnail untuk list view, banner untuk detail view. CDN highly recommended.

4. **Date Filtering**: 
   - Gunakan ISO 8601 format untuk date filters
   - `date_from` dan `date_to` adalah inclusive
   - Default timezone adalah Asia/Jakarta untuk Indonesia users

5. **Search Performance**: 
   - Implement full-text search di backend (ElasticSearch recommended)
   - Search minimum 3 characters
   - Debounce search input di client (300-500ms)

6. **Empty States**: 
   - Jika tidak ada event yang match filter, return empty array dengan pagination info
   - Berikan suggestion untuk broaden search criteria

7. **Featured Events**: Max 10 featured events active at any time untuk menjaga kualitas highlight.

8. **View Tracking**: Increment `views_count` setiap kali user membuka detail event (deduplicate by user + 24h window).

9. **Registration Status Logic**:
   - `open`: Available untuk registrasi
   - `closed`: Registrasi sudah ditutup
   - `full`: Capacity sudah penuh
   - `not_required`: Event tidak memerlukan registrasi

10. **Localization**: 
    - Gunakan `Accept-Language` header untuk response messages
    - Event description & title bisa multilingual (tergantung backend implementation)
