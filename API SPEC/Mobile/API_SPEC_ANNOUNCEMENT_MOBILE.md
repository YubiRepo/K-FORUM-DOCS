# API Spec — Announcement Module (Mobile Client)

Dokumentasi API untuk mobile aplikasi mengambil dan mengelola announcement.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/announcements`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Optional untuk list, required untuk read tracking)
  - `Accept-Language: <lang_code>` (default: `id`)

---

## 1. GET /announcements

Ambil daftar announcement yang published. User hanya lihat:
- Global announcements
- Regional announcements untuk region mereka

**Authentication**: Optional (akan lebih accurate jika auth karena get user region)

**Method**: GET

**URL**: `/api/v1/mobile/announcements`

**Query Parameters**:
- `type` (optional): Filter by type (`disaster`, `system`, `urgent`, `info`)
- `priority` (optional): Filter by priority (`critical`, `high`, `medium`, `low`)
- `limit` (optional, default: 20, max: 100)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "ann_disaster_001",
      "title": "Gempa Bumi 7.2 SR",
      "body": "Gempa bumi magnitude 7.2 terjadi di Jawa Barat...",
      "image_url": "https://cdn.../disaster_banner.jpg",
      "type": "disaster",
      "priority": "critical",
      "scope": "global",
      "is_read": false,
      "published_at": "2026-05-25T10:05:00.000Z",
      "expires_at": "2026-06-01T00:00:00.000Z"
    },
    {
      "id": "ann_system_001",
      "title": "Maintenance Server Jakarta",
      "body": "Database upgrade dijadwalkan Selasa 10 malam...",
      "type": "system",
      "priority": "high",
      "scope": "regional",
      "region": {
        "id": "region_jakarta",
        "name": "Jakarta"
      },
      "is_read": false,
      "published_at": "2026-05-24T08:00:00.000Z"
    },
    {
      "id": "ann_info_001",
      "title": "Advanced Analytics Available",
      "body": "Fitur baru: Advanced Analytics untuk Pro members...",
      "type": "info",
      "priority": "medium",
      "scope": "global",
      "is_read": true,
      "published_at": "2026-05-20T09:00:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 45
  }
}
```

---

## 2. GET /announcements/{announcement_id}

Ambil detail lengkap satu announcement + metadata.

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/announcements/{announcement_id}`

**Response (200 OK - Disaster)**:
```json
{
  "data": {
    "id": "ann_disaster_001",
    "title": "Gempa Bumi 7.2 SR",
    "body": "Gempa bumi magnitude 7.2 terjadi di Jawa Barat...",
    "image_url": "https://cdn.../disaster_banner.jpg",
    "type": "disaster",
    "priority": "critical",
    "scope": "global",
    "is_read": false,
    "published_at": "2026-05-25T10:05:00.000Z",
    "expires_at": "2026-06-01T00:00:00.000Z",
    
    "metadata": {
      "disaster_subtype": "earthquake",
      "magnitude": 7.2,
      "depth_km": 12,
      "location": {
        "name": "Jawa Barat",
        "latitude": -6.9271,
        "longitude": 107.6693,
        "address": "Near Bandung"
      },
      "affected_areas": [
        {
          "area_name": "Bandung",
          "latitude": -6.9175,
          "longitude": 107.6095,
          "impact_level": "severe"
        },
        {
          "area_name": "Cirebon",
          "latitude": -6.7034,
          "longitude": 108.3469,
          "impact_level": "moderate"
        }
      ],
      "evacuation_centers": [
        {
          "name": "GOR Bandung",
          "latitude": -6.8944,
          "longitude": 107.6047,
          "capacity": 500,
          "contact": "081234567890"
        }
      ],
      "helpline": "119"
    }
  }
}
```

**Response (200 OK - System)**:
```json
{
  "data": {
    "id": "ann_system_001",
    "title": "Maintenance Server Jakarta",
    "body": "Database upgrade dijadwalkan...",
    "type": "system",
    "priority": "high",
    "scope": "regional",
    "region": {
      "id": "region_jakarta",
      "name": "Jakarta"
    },
    "is_read": false,
    "published_at": "2026-05-24T08:00:00.000Z",
    
    "metadata": {
      "system_subtype": "maintenance",
      "affected_services": ["Database", "API Gateway"],
      "maintenance_window": {
        "start_time": "2026-05-25T22:00:00Z",
        "end_time": "2026-05-26T02:00:00Z",
        "timezone": "UTC+7"
      },
      "estimated_duration_hours": 4,
      "impact": "All services may be unavailable",
      "status_page_url": "https://status.kai.app"
    }
  }
}
```

**Response (200 OK - Info)**:
```json
{
  "data": {
    "id": "ann_info_001",
    "title": "Advanced Analytics Available",
    "body": "Fitur baru untuk Pro members...",
    "type": "info",
    "priority": "medium",
    "scope": "global",
    "is_read": true,
    "published_at": "2026-05-20T09:00:00.000Z",
    
    "metadata": {
      "info_subtype": "feature_launch",
      "feature_name": "Advanced Analytics",
      "availability": "pro_only",
      "features": [
        {
          "name": "Real-time Engagement Metrics",
          "description": "Track posts and comments in real-time"
        },
        {
          "name": "Custom Reports",
          "description": "Generate and export PDF reports"
        }
      ],
      "learn_more_link": "https://kai.app/features/analytics"
    }
  }
}
```

---

## 3. POST /announcements/{announcement_id}/read

Mark announcement sebagai sudah dibaca. Auto-triggered saat user open detail, tapi bisa juga explicit call.

**Authentication**: Required

**Method**: POST

**URL**: `/api/v1/mobile/announcements/{announcement_id}/read`

**Request Body**:
```json
{
  "platform": "android"  // atau "ios", "web"
}
```

**Response (200 OK)**:
```json
{
  "message": "Announcement marked as read",
  "data": {
    "announcement_id": "ann_disaster_001",
    "read_at": "2026-05-25T10:10:00.000Z"
  }
}
```

---

## 4. GET /announcements/unread/count

Ambil jumlah announcement unread.

**Authentication**: Required

**Method**: GET

**URL**: `/api/v1/mobile/announcements/unread/count`

**Response (200 OK)**:
```json
{
  "data": {
    "unread_count": 3,
    "critical_count": 1
  }
}
```

---

## 5. GET /announcements/by-type/{type}

Ambil announcement spesifik type (untuk filtering).

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/announcements/by-type/{type}`

**URL Parameters**:
- `type`: `disaster`, `system`, `urgent`, `info`

**Query Parameters**:
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "ann_disaster_001",
      "title": "Gempa Bumi 7.2 SR",
      "type": "disaster",
      "published_at": "2026-05-25T10:05:00.000Z"
    },
    {
      "id": "ann_disaster_002",
      "title": "Banjir Bandung",
      "type": "disaster",
      "published_at": "2026-05-24T15:30:00.000Z"
    }
  ]
}
```

---

## 6. GET /announcements/critical

Ambil hanya announcement dengan priority CRITICAL (urgent access).

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/announcements/critical`

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "ann_disaster_001",
      "title": "Gempa Bumi 7.2 SR",
      "type": "disaster",
      "priority": "critical",
      "published_at": "2026-05-25T10:05:00.000Z"
    }
  ]
}
```

---

## 7. GET /announcements/search

Search announcement by keyword.

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/announcements/search`

**Query Parameters**:
- `q` (required): Search keyword
- `limit` (optional, default: 20)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "ann_disaster_001",
      "title": "Gempa Bumi 7.2 SR",
      "body": "Gempa bumi magnitude 7.2 terjadi di Jawa Barat...",
      "type": "disaster",
      "published_at": "2026-05-25T10:05:00.000Z"
    }
  ]
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "message": "Invalid type parameter"
}
```

### 401 Unauthorized
```json
{
  "message": "Authentication required"
}
```

### 404 Not Found
```json
{
  "message": "Announcement not found"
}
```

---

## Status Codes

- `200 OK` - Success
- `400 Bad Request` - Bad input
- `401 Unauthorized` - Auth required
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

---

## Caching Strategy

- `GET /announcements` - Cache 5 minutes
- `GET /announcements/{id}` - Cache 1 hour
- `GET /announcements/critical` - Cache 2 minutes (more urgent)
- Other queries - No cache

---

*API spec announcement untuk mobile client.*
