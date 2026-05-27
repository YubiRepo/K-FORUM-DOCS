# API Spec — Announcement Module (Web Backoffice)

Dokumentasi API untuk admin manage announcements di dashboard backoffice.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/announcements`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin or Admin Regional)
- **Authorization**:
  - Superadmin: all operations, global + regional
  - Admin Regional: only own region operations

---

## 1. POST /announcements

Create announcement baru.

**Authentication**: Required (Superadmin or Admin Regional)

**Method**: POST

**URL**: `/api/v1/web/announcements`

**Request Body (Disaster Example)**:
```json
{
  "title": "Gempa Bumi 7.2 SR",
  "body": "Gempa bumi magnitude 7.2 terjadi di Jawa Barat. Epicenter dekat Bandung...",
  "image_url": "https://cdn.../disaster_banner.jpg",
  
  "type": "disaster",
  "priority": "critical",
  "scope": "global",
  "region_id": null,
  
  "status": "draft",
  "expires_at": "2026-06-01T00:00:00Z",
  
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
```

**Response (201 Created)**:
```json
{
  "data": {
    "id": "ann_disaster_001",
    "title": "Gempa Bumi 7.2 SR",
    "type": "disaster",
    "priority": "critical",
    "scope": "global",
    "status": "draft",
    "created_by": "superadmin_001",
    "created_at": "2026-05-25T10:00:00.000Z",
    "total_recipients": 0  // belum publish, jadi 0
  },
  "message": "Announcement created successfully"
}
```

**Request Body (System Example)**:
```json
{
  "title": "Maintenance Server Jakarta",
  "body": "Database upgrade dijadwalkan Selasa 10 malam - tengah malam (4 jam).",
  
  "type": "system",
  "priority": "high",
  "scope": "regional",
  "region_id": "region_jakarta",
  
  "status": "draft",
  
  "metadata": {
    "system_subtype": "maintenance",
    "affected_services": ["Database", "API Gateway"],
    "maintenance_window": {
      "start_time": "2026-05-26T22:00:00Z",
      "end_time": "2026-05-27T02:00:00Z"
    },
    "estimated_duration_hours": 4,
    "status_page_url": "https://status.kai.app"
  }
}
```

---

## 2. GET /announcements

List semua announcements dengan filter.

**Authentication**: Required

**Method**: GET

**URL**: `/api/v1/web/announcements`

**Query Parameters**:
- `type` (optional): `disaster`, `system`, `urgent`, `info`
- `priority` (optional): `critical`, `high`, `medium`, `low`
- `status` (optional): `draft`, `published`, `archived`
- `scope` (optional): `global`, `regional`
- `region_id` (optional): Filter by region (superadmin: any, admin: own region only)
- `search` (optional): Search title/body
- `created_from` (optional): Filter from date
- `created_to` (optional): Filter to date
- `sort` (optional): `published_at`, `-published_at`, `created_at` (default: `-created_at`)
- `limit` (optional, default: 20, max: 100)
- `offset` (optional, default: 0)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "ann_disaster_001",
      "title": "Gempa Bumi 7.2 SR",
      "type": "disaster",
      "priority": "critical",
      "scope": "global",
      "status": "published",
      "created_by": "superadmin_001",
      "published_at": "2026-05-25T10:05:00.000Z",
      "total_recipients": 5234,
      "total_sent": 5200,
      "read_count": 3421,
      "created_at": "2026-05-25T10:00:00.000Z"
    },
    {
      "id": "ann_system_001",
      "title": "Maintenance Server Jakarta",
      "type": "system",
      "priority": "high",
      "scope": "regional",
      "region": {
        "id": "region_jakarta",
        "name": "Jakarta"
      },
      "status": "draft",
      "created_by": "admin_jakarta_001",
      "total_recipients": 0,
      "created_at": "2026-05-24T08:00:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 45
  },
  "summary": {
    "total_draft": 3,
    "total_published": 35,
    "total_archived": 7
  }
}
```

---

## 3. GET /announcements/{announcement_id}

Get detail satu announcement.

**Authentication**: Required

**Method**: GET

**URL**: `/api/v1/web/announcements/{announcement_id}`

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ann_disaster_001",
    "title": "Gempa Bumi 7.2 SR",
    "body": "Gempa bumi magnitude 7.2...",
    "image_url": "https://cdn.../disaster_banner.jpg",
    "type": "disaster",
    "priority": "critical",
    "scope": "global",
    "region_id": null,
    
    "status": "published",
    "published_at": "2026-05-25T10:05:00.000Z",
    "expires_at": "2026-06-01T00:00:00Z",
    
    "created_by": "superadmin_001",
    "created_by_name": "Admin KAI",
    "created_at": "2026-05-25T10:00:00.000Z",
    "updated_at": "2026-05-25T10:05:00.000Z",
    
    "metadata": {
      "disaster_subtype": "earthquake",
      "magnitude": 7.2,
      "depth_km": 12,
      "location": {...},
      "affected_areas": [...],
      "evacuation_centers": [...]
    },
    
    "total_recipients": 5234,
    "total_sent": 5200,
    "read_count": 3421,
    "read_rate": 65.7
  }
}
```

---

## 4. PUT /announcements/{announcement_id}

Edit announcement (hanya saat draft atau published).

**Authentication**: Required

**Method**: PUT

**URL**: `/api/v1/web/announcements/{announcement_id}`

**Request Body**:
```json
{
  "title": "Updated Title",
  "body": "Updated body...",
  "priority": "high",
  "expires_at": "2026-06-05T00:00:00Z",
  "metadata": {
    "disaster_subtype": "earthquake",
    "magnitude": 7.2,
    "evacuation_centers": [...]
  }
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ann_disaster_001",
    "title": "Updated Title",
    "updated_at": "2026-05-25T10:30:00.000Z"
  },
  "message": "Announcement updated successfully"
}
```

---

## 5. POST /announcements/{announcement_id}/publish

Publish announcement (dari draft → published).

**Authentication**: Required

**Method**: POST

**URL**: `/api/v1/web/announcements/{announcement_id}/publish`

**Request Body**: Empty atau bisa ada:
```json
{
  "schedule_for": null  // jika mau schedule di masa depan (not yet supported)
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ann_disaster_001",
    "status": "published",
    "published_at": "2026-05-25T10:05:00.000Z",
    "total_recipients": 5234,
    "total_sent": 5200,
    "notification_sent": true
  },
  "message": "Announcement published to 5.234 users",
  "events": [
    {
      "type": "push_notification_sent",
      "count": 5200
    },
    {
      "type": "email_sent",
      "count": 5234,
      "reason": "CRITICAL priority"
    }
  ]
}
```

---

## 6. POST /announcements/{announcement_id}/archive

Archive announcement (dari published/draft → archived).

**Authentication**: Required

**Method**: POST

**URL**: `/api/v1/web/announcements/{announcement_id}/archive`

**Request Body**: Empty

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ann_disaster_001",
    "status": "archived",
    "archived_at": "2026-05-25T11:00:00.000Z"
  },
  "message": "Announcement archived"
}
```

---

## 7. DELETE /announcements/{announcement_id}

Hapus announcement (hanya untuk draft).

**Authentication**: Required

**Method**: DELETE

**URL**: `/api/v1/web/announcements/{announcement_id}`

**Response (200 OK)**:
```json
{
  "message": "Announcement deleted",
  "data": {
    "deleted_id": "ann_disaster_001"
  }
}
```

**Response (409 Conflict - Cannot delete published)**:
```json
{
  "message": "Cannot delete published announcement. Please archive instead."
}
```

---

## 8. GET /announcements/{announcement_id}/analytics

Get analytics untuk satu announcement.

**Authentication**: Required

**Method**: GET

**URL**: `/api/v1/web/announcements/{announcement_id}/analytics`

**Response (200 OK)**:
```json
{
  "data": {
    "id": "ann_disaster_001",
    "title": "Gempa Bumi 7.2 SR",
    
    "delivery_stats": {
      "total_recipients": 5234,
      "total_sent": 5200,
      "total_failed": 34,
      "delivery_rate": 99.4
    },
    
    "read_stats": {
      "total_read": 3421,
      "read_rate": 65.7
    },
    
    "breakdown_by_platform": {
      "android": {
        "sent": 2600,
        "read": 1820,
        "read_rate": 70.0
      },
      "ios": {
        "sent": 1800,
        "read": 1050,
        "read_rate": 58.3
      },
      "web": {
        "sent": 800,
        "read": 551,
        "read_rate": 68.9
      }
    },
    
    "timeline": {
      "created_at": "2026-05-25T10:00:00.000Z",
      "published_at": "2026-05-25T10:05:00.000Z",
      "first_read_at": "2026-05-25T10:07:00.000Z",
      "last_read_at": "2026-05-25T14:35:00.000Z"
    }
  }
}
```

---

## 9. GET /announcements/region/{region_id}

Get announcements untuk specific region (admin regional only lihat own region).

**Authentication**: Required

**Method**: GET

**URL**: `/api/v1/web/announcements/region/{region_id}`

**Query Parameters**:
- `type` (optional)
- `status` (optional)
- `limit` (optional)
- `offset` (optional)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "ann_system_001",
      "title": "Maintenance Server Jakarta",
      "type": "system",
      "scope": "regional",
      "region": {
        "id": "region_jakarta",
        "name": "Jakarta"
      },
      "status": "published"
    }
  ]
}
```

---

## 10. POST /announcements/batch-create

Create multiple announcements sekaligus (for bulk disaster alerts).

**Authentication**: Required (Superadmin only)

**Method**: POST

**URL**: `/api/v1/web/announcements/batch-create`

**Request Body**:
```json
{
  "announcements": [
    {
      "title": "Gempa Bandung",
      "body": "...",
      "type": "disaster",
      "priority": "critical",
      "scope": "regional",
      "region_id": "region_jawa_barat",
      "metadata": {...}
    },
    {
      "title": "Gempa Jakarta",
      "body": "...",
      "type": "disaster",
      "priority": "critical",
      "scope": "regional",
      "region_id": "region_jakarta",
      "metadata": {...}
    }
  ]
}
```

**Response (201 Created)**:
```json
{
  "data": {
    "created": 2,
    "failed": 0,
    "announcements": [
      {
        "id": "ann_disaster_001",
        "title": "Gempa Bandung",
        "status": "draft"
      },
      {
        "id": "ann_disaster_002",
        "title": "Gempa Jakarta",
        "status": "draft"
      }
    ]
  },
  "message": "2 announcements created successfully"
}
```

---

## Error Responses

### 400 Bad Request
```json
{
  "message": "Invalid request",
  "errors": {
    "title": ["Title is required"],
    "type": ["Invalid type"]
  }
}
```

### 401 Unauthorized
```json
{
  "message": "Authentication required"
}
```

### 403 Forbidden
```json
{
  "message": "You don't have permission to manage announcements in other regions"
}
```

### 404 Not Found
```json
{
  "message": "Announcement not found"
}
```

### 409 Conflict
```json
{
  "message": "Cannot delete published announcement"
}
```

---

## Status Codes

- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Validation error
- `401 Unauthorized` - Auth required
- `403 Forbidden` - Permission denied
- `404 Not Found` - Resource not found
- `409 Conflict` - State conflict
- `500 Internal Server Error` - Server error

---

## Rate Limiting

- Limit: 200 requests per minute per admin
- Headers:
  - `X-RateLimit-Limit: 200`
  - `X-RateLimit-Remaining: 195`
  - `X-RateLimit-Reset: 1629856800`

---

*API spec announcement untuk web backoffice.*
