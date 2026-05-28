# API Spec — Event Module (Web Backoffice)

Dokumentasi API event module untuk backoffice dashboard — manage, review, approve/reject events, dan configure event settings.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/events`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin only)
- **Authorization**: Semua endpoint backoffice event hanya untuk Superadmin
- **Error Format**: Same as other backoffice APIs (standard message or validation error)

---

## Image Upload Flow

Saat superadmin create atau edit event langsung dari backoffice, images tetap diupload via endpoint upload terlebih dahulu sebelum dikirim sebagai URL.

```
┌────────────────────────────────────────────────────────┐
│ IMAGE UPLOAD FLOW (Backoffice)                         │
├────────────────────────────────────────────────────────┤
│                                                        │
│  Step 1: Upload image(s)                               │
│  POST /api/v1/web/media/upload                         │
│  → Returns: { urls: ["https://cdn.../img1.jpg", ...] } │
│                                                        │
│  Step 2: Create/Edit event dengan image URLs           │
│  POST /api/v1/web/events                               │
│  { "images": ["https://cdn.../img1.jpg", ...] }        │
│                                                        │
└────────────────────────────────────────────────────────┘
```

---

## Model Data Utama

### 1. Event Object (Backoffice List)

```json
{
  "id": "uuid",
  "title": "Futsal Tournament 2026",
  "cover_image": "https://cdn.example.com/events/uploads/img_uuid1.jpg",
  "event_type": "offline",
  "venue_name": "GOR Senayan",
  "venue_address": "Jl. Pintu I Senayan, Jakarta Pusat",
  "online_platform": null,
  "online_url": null,
  "event_date": "2026-06-15",
  "event_time": "14:00",
  "category": {
    "id": "uuid",
    "name": "Sports"
  },
  "organizer": {
    "id": "uuid",
    "name": "Andi Pratama",
    "email": "andi@example.com"
  },
  "status": "pending_approval",
  "approval_status": "pending",
  "submitted_at": "2026-05-25T09:00:00.000Z",
  "created_at": "2026-05-25T08:00:00.000Z"
}
```

### 2. Event Object (Backoffice Detail)

```json
{
  "id": "uuid",
  "title": "Futsal Tournament 2026",
  "description": "Turnamen futsal tahunan terbesar di Jakarta...",
  "images": [
    "https://cdn.example.com/events/uploads/img_uuid1.jpg",
    "https://cdn.example.com/events/uploads/img_uuid2.jpg",
    "https://cdn.example.com/events/uploads/img_uuid3.jpg"
  ],
  "cover_image": "https://cdn.example.com/events/uploads/img_uuid1.jpg",
  "event_type": "offline",
  "venue_name": "GOR Senayan",
  "venue_address": "Jl. Pintu I Senayan, Jakarta Pusat",
  "online_platform": null,
  "online_url": null,
  "category": {
    "id": "uuid",
    "name": "Sports"
  },
  "event_date": "2026-06-15",
  "event_end_date": null,
  "event_time": "14:00",
  "registration_url": "https://eventbrite.com/e/futsal-2026",
  "organizer": {
    "id": "uuid",
    "name": "Andi Pratama",
    "email": "andi@example.com",
    "subscription_plan": "pro"
  },
  "status": "pending_approval",
  "approval_status": "pending",
  "approved_by": null,
  "approval_notes": null,
  "rejection_reason": null,
  "save_count": 0,
  "share_count": 0,
  "created_at": "2026-05-25T08:00:00.000Z",
  "submitted_at": "2026-05-25T09:00:00.000Z",
  "published_at": null,
  "updated_at": "2026-05-25T09:00:00.000Z"
}
```

> Untuk event **online**: `venue_name` dan `venue_address` bernilai `null`, `online_platform` dan `online_url` diisi.
> Untuk event **hybrid**: semua field venue dan online diisi.

---

## Endpoints

### 1. Upload Image(s) — Backoffice

Upload satu atau lebih gambar untuk event dari backoffice. Returns URL yang siap dipakai.

- **URL**: `POST /api/v1/web/media/upload`
- **Autentikasi**: Yes (Superadmin)
- **Content-Type**: `multipart/form-data`

- **Request**:
  ```
  files[]: image1.jpg  (max 5 files, max 5MB per file)
  files[]: image2.jpg
  context: "event"
  ```

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "urls": [
        "https://cdn.example.com/events/uploads/img_uuid1.jpg",
        "https://cdn.example.com/events/uploads/img_uuid2.jpg"
      ]
    },
    "message": "2 image(s) uploaded successfully"
  }
  ```

---

### 2. List All Events

Ambil semua events dengan filter lengkap — untuk management dan monitoring.

- **URL**: `GET /api/v1/web/events`
- **Autentikasi**: Yes (Superadmin)
- **Query Parameters**:
  - `search` (optional): Cari by title, venue_name, venue_address, atau organizer name/email
  - `status` (optional): `draft`, `pending_approval`, `published`, `rejected`, `cancelled`
  - `approval_status` (optional): `pending`, `approved`, `rejected`
  - `event_type` (optional): `offline`, `online`, `hybrid`
  - `category_id` (optional): Filter by kategori
  - `organizer_id` (optional): Filter by organizer user
  - `date_from` (optional): Filter event_date from
  - `date_to` (optional): Filter event_date to
  - `sort` (optional): `created_at`, `-created_at`, `event_date`, `-event_date`, `submitted_at` (default: `-created_at`)
  - `limit` (optional, default: 20, max: 100)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "event_123",
        "title": "Futsal Tournament 2026",
        "cover_image": "https://cdn.example.com/events/uploads/img_uuid1.jpg",
        "event_type": "offline",
        "venue_name": "GOR Senayan",
        "venue_address": "Jl. Pintu I Senayan, Jakarta Pusat",
        "event_date": "2026-06-15",
        "category": { "id": "uuid", "name": "Sports" },
        "organizer": {
          "id": "uuid",
          "name": "Andi Pratama",
          "email": "andi@example.com"
        },
        "status": "pending_approval",
        "approval_status": "pending",
        "submitted_at": "2026-05-25T09:00:00.000Z",
        "created_at": "2026-05-25T08:00:00.000Z"
      },
      {
        "id": "event_125",
        "title": "Webinar Product Management 2026",
        "cover_image": null,
        "event_type": "online",
        "venue_name": null,
        "venue_address": null,
        "event_date": "2026-06-20",
        "category": { "id": "uuid", "name": "Business" },
        "organizer": {
          "id": "uuid",
          "name": "Budi Santoso",
          "email": "budi@example.com"
        },
        "status": "published",
        "approval_status": "approved",
        "submitted_at": "2026-05-20T09:00:00.000Z",
        "created_at": "2026-05-20T08:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 134
    },
    "summary": {
      "total_pending": 5,
      "total_published": 89,
      "total_rejected": 12,
      "total_cancelled": 28
    }
  }
  ```

---

### 3. Get Pending Approvals

Ambil events yang menunggu approval — shortcut untuk review queue.

- **URL**: `GET /api/v1/web/events/pending`
- **Autentikasi**: Yes (Superadmin)
- **Query Parameters**:
  - `sort` (optional): `submitted_at`, `-submitted_at` (default: `submitted_at` — oldest first)
  - `limit` (optional, default: 20)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "event_124",
        "title": "Basketball Cup 2026",
        "cover_image": "https://cdn.example.com/events/uploads/img_uuid4.jpg",
        "event_type": "offline",
        "venue_name": "Hall A, ISTORA Senayan",
        "venue_address": "Jl. Pintu VI Senayan, Jakarta",
        "event_date": "2026-07-01",
        "category": { "id": "uuid", "name": "Sports" },
        "organizer": {
          "id": "uuid",
          "name": "Budi Santoso",
          "email": "budi@example.com"
        },
        "status": "pending_approval",
        "approval_status": "pending",
        "submitted_at": "2026-05-24T10:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 5
    }
  }
  ```

---

### 4. Get Event Detail

Ambil detail lengkap satu event untuk review.

- **URL**: `GET /api/v1/web/events/{event_id}`
- **Autentikasi**: Yes (Superadmin)

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "event_124",
      "title": "Basketball Cup 2026",
      "description": "Turnamen basket tingkat regional...",
      "images": [
        "https://cdn.example.com/events/uploads/img_uuid4.jpg",
        "https://cdn.example.com/events/uploads/img_uuid5.jpg"
      ],
      "cover_image": "https://cdn.example.com/events/uploads/img_uuid4.jpg",
      "event_type": "offline",
      "venue_name": "Hall A, ISTORA Senayan",
      "venue_address": "Jl. Pintu VI Senayan, Jakarta",
      "online_platform": null,
      "online_url": null,
      "category": { "id": "uuid", "name": "Sports" },
      "event_date": "2026-07-01",
      "event_end_date": "2026-07-03",
      "event_time": "09:00",
      "registration_url": "https://eventbrite.com/e/basketball-cup",
      "organizer": {
        "id": "uuid",
        "name": "Budi Santoso",
        "email": "budi@example.com",
        "subscription_plan": "pro"
      },
      "status": "pending_approval",
      "approval_status": "pending",
      "approved_by": null,
      "approval_notes": null,
      "rejection_reason": null,
      "save_count": 0,
      "share_count": 0,
      "created_at": "2026-05-24T09:00:00.000Z",
      "submitted_at": "2026-05-24T10:00:00.000Z",
      "published_at": null,
      "updated_at": "2026-05-24T10:00:00.000Z"
    }
  }
  ```


---

### 5. Approve Event

Approve event — status jadi `published` dan event langsung live.

- **URL**: `POST /api/v1/web/events/{event_id}/approve`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "notes": "Looks good, semua info lengkap. Approved!"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "event_124",
      "title": "Basketball Cup 2026",
      "status": "published",
      "approval_status": "approved",
      "approved_by": "superadmin_uuid",
      "approved_by_name": "Super Admin",
      "notes": "Looks good, semua info lengkap. Approved!",
      "approved_at": "2026-05-25T11:00:00.000Z",
      "published_at": "2026-05-25T11:00:00.000Z"
    },
    "message": "Event approved and published successfully"
  }
  ```

- **Response (Error — 400)**:
  ```json
  {
    "message": "Only pending_approval events can be approved"
  }
  ```

---

### 6. Reject Event

Reject event dengan alasan — organizer akan notified dan bisa edit & resubmit.

- **URL**: `POST /api/v1/web/events/{event_id}/reject`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "reason": "Informasi venue tidak lengkap. Mohon cantumkan alamat lengkap dan konfirmasi booking venue."
  }
  ```

- **Request Validation**:
  - `reason`: Required, string, min 10 chars, max 1000 chars

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "event_124",
      "title": "Basketball Cup 2026",
      "status": "rejected",
      "approval_status": "rejected",
      "rejected_by": "superadmin_uuid",
      "rejected_by_name": "Super Admin",
      "rejection_reason": "Informasi venue tidak lengkap. Mohon cantumkan alamat lengkap dan konfirmasi booking venue.",
      "rejected_at": "2026-05-25T11:05:00.000Z"
    },
    "message": "Event rejected. Organizer has been notified."
  }
  ```

- **Response (Error — 400)**:
  ```json
  {
    "message": "Only pending_approval events can be rejected"
  }
  ```

---

### 7. Force Cancel Event

Cancel event apapun statusnya — admin action.

- **URL**: `POST /api/v1/web/events/{event_id}/cancel`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "reason": "Pelanggaran community guidelines",
    "notify_organizer": true
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "event_123",
      "status": "cancelled",
      "reason": "Pelanggaran community guidelines",
      "cancelled_by": "superadmin_uuid",
      "cancelled_at": "2026-05-25T12:00:00.000Z"
    },
    "message": "Event cancelled by admin"
  }
  ```

---

### 8. Create Event (Backoffice)

Superadmin buat event langsung dari backoffice. Langsung published tanpa approval.

- **URL**: `POST /api/v1/web/events`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "title": "Official Platform Event 2026",
    "description": "Event resmi dari platform...",
    "images": [
      "https://cdn.example.com/events/uploads/img_uuid1.jpg",
      "https://cdn.example.com/events/uploads/img_uuid2.jpg"
    ],
    "category_id": "uuid_community",
    "event_type": "hybrid",
    "venue_name": "Jakarta Convention Center",
    "venue_address": "Jl. Gatot Subroto, Jakarta Selatan",
    "online_platform": "YouTube Live",
    "online_url": "https://youtube.com/live/abc123",
    "event_date": "2026-08-01",
    "event_end_date": "2026-08-03",
    "event_time": "09:00",
    "registration_url": "https://eventbrite.com/e/official-event-2026"
  }
  ```

- **Request Validation**:
  - `title`: Required, string, max 200 chars
  - `description`: Required, string
  - `images`: Optional, array of URLs, max 10 items. First item auto-set sebagai cover image
  - `category_id`: Required, must exist
  - `event_type`: Required, enum: `offline`, `online`, `hybrid`
  - `venue_name`: Required jika `event_type` adalah `offline` atau `hybrid`
  - `venue_address`: Required jika `event_type` adalah `offline` atau `hybrid`
  - `online_platform`: Required jika `event_type` adalah `online` atau `hybrid`
  - `online_url`: Required jika `event_type` adalah `online` atau `hybrid`, valid URL
  - `event_date`: Required, must be future date, format `YYYY-MM-DD`
  - `event_end_date`: Optional, must be >= `event_date`
  - `event_time`: Required, format `HH:mm`
  - `registration_url`: Optional, valid URL

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "event_200",
      "title": "Official Platform Event 2026",
      "status": "published",
      "images": [
        "https://cdn.example.com/events/uploads/img_uuid1.jpg",
        "https://cdn.example.com/events/uploads/img_uuid2.jpg"
      ],
      "cover_image": "https://cdn.example.com/events/uploads/img_uuid1.jpg",
      "published_at": "2026-05-25T12:00:00.000Z",
      "created_at": "2026-05-25T12:00:00.000Z"
    },
    "message": "Event created and published successfully"
  }
  ```

---

### 9. Edit Event (Backoffice)

Superadmin edit event apapun statusnya.

- **URL**: `PUT /api/v1/web/events/{event_id}`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "title": "Updated title",
    "images": [
      "https://cdn.example.com/events/uploads/img_new1.jpg"
    ],
    "event_type": "offline",
    "venue_name": "Gedung Serbaguna Senayan",
    "venue_address": "Jl. Asia Afrika, Jakarta Pusat",
    "event_date": "2026-08-05",
    "event_time": "10:00"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "event_200",
      "title": "Updated title",
      "cover_image": "https://cdn.example.com/events/uploads/img_new1.jpg",
      "images": [
        "https://cdn.example.com/events/uploads/img_new1.jpg"
      ],
      "updated_at": "2026-05-25T13:00:00.000Z"
    },
    "message": "Event updated successfully"
  }
  ```

---

### 10. Delete Event (Hard Delete)

Hapus event permanen dari database. Gunakan dengan hati-hati.

- **URL**: `DELETE /api/v1/web/events/{event_id}`
- **Autentikasi**: Yes (Superadmin)

- **Response (Success 200)**:
  ```json
  {
    "message": "Event permanently deleted",
    "data": {
      "deleted_id": "event_123",
      "deleted_at": "2026-05-25T13:05:00.000Z"
    }
  }
  ```

---

### 11. Manage Event Categories

CRUD untuk master data kategori event.

**Get All Categories:**
- **URL**: `GET /api/v1/web/events/categories`

- **Response (Success 200)**:
  ```json
  {
    "data": [
      { "id": "uuid", "name": "Sports", "description": "Sports events", "created_at": "2026-01-01T00:00:00.000Z" },
      { "id": "uuid", "name": "Business", "description": "Business & networking", "created_at": "2026-01-01T00:00:00.000Z" }
    ]
  }
  ```

**Create Category:**
- **URL**: `POST /api/v1/web/events/categories`

- **Request Body**:
  ```json
  {
    "name": "Health & Fitness",
    "description": "Health, fitness, and wellness events"
  }
  ```

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "name": "Health & Fitness",
      "description": "Health, fitness, and wellness events",
      "created_at": "2026-05-25T13:10:00.000Z"
    },
    "message": "Category created successfully"
  }
  ```

**Update Category:**
- **URL**: `PUT /api/v1/web/events/categories/{category_id}`

**Delete Category:**
- **URL**: `DELETE /api/v1/web/events/categories/{category_id}`

---

### 12. Event Module Settings

Superadmin configure pengaturan event module secara global.

- **URL GET**: `GET /api/v1/web/events/settings`
- **URL PUT**: `PUT /api/v1/web/events/settings`
- **Autentikasi**: Yes (Superadmin)

- **Response GET (Success 200)**:
  ```json
  {
    "data": {
      "auto_publish": false,
      "require_approval": true,
      "max_events_per_member_per_month": 5,
      "allow_cancel_after_published": true,
      "lock_editing_days_before_event": 3,
      "send_reminders": true,
      "reminder_times": ["1 week before", "3 days before", "1 day before"],
      "allow_comments": false,
      "max_images_per_event": 10,
      "max_image_size_mb": 5
    }
  }
  ```

- **Request Body PUT**:
  ```json
  {
    "auto_publish": true,
    "require_approval": false,
    "max_events_per_member_per_month": 10
  }
  ```

- **Response PUT (Success 200)**:
  ```json
  {
    "data": {
      "auto_publish": true,
      "require_approval": false,
      "max_events_per_member_per_month": 10,
      "updated_at": "2026-05-25T13:15:00.000Z"
    },
    "message": "Event module settings updated"
  }
  ```

---

## UI Flow Example

### Page: Pending Approvals Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│ Events — Pending Approval (5)                               │
├─────────────────────────────────────────────────────────────┤
│ Sort: [Oldest first ▼]                    [View All Events] │
│                                                             │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ [img] Basketball Cup 2026                           │    │
│ │       Budi Santoso · budi@example.com               │    │
│ │       🏟 Offline · Hall A, ISTORA  📅 2026-07-01    │    │
│ │       Submitted: 2026-05-24 10:00                   │    │
│ │                  [View Detail] [Approve] [Reject]   │    │
│ └─────────────────────────────────────────────────────┘    │
│                                                             │
│ ┌─────────────────────────────────────────────────────┐    │
│ │ [img] Badminton Open 2026                           │    │
│ │       Citra Dewi · citra@example.com                │    │
│ │       🏟 Offline · GOR Cipayung  📅 2026-07-10      │    │
│ │       Submitted: 2026-05-25 08:30                   │    │
│ │                  [View Detail] [Approve] [Reject]   │    │
│ └─────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Flow: Review & Approve Event

```
1. Superadmin click [View Detail]

2. Event Detail Modal:
   ┌─────────────────────────────────────────────────────┐
   │ Basketball Cup 2026                      [Close]    │
   ├─────────────────────────────────────────────────────┤
   │ Images: [img1] [img2]                               │
   │ Tipe: Offline                                       │
   │ Venue: Hall A, ISTORA Senayan                       │
   │ Alamat: Jl. Pintu VI Senayan, Jakarta               │
   │ Date: 2026-07-01 s/d 2026-07-03, 09:00             │
   │ Registration: https://eventbrite.com/...            │
   │                                                     │
   │ Organizer: Budi Santoso (Pro Member)                │
   │ Email: budi@example.com                             │
   │ Submitted: 2026-05-24 10:00                         │
   │                                                     │
   │ Description:                                        │
   │ Turnamen basket tingkat regional...                 │
   │                                                     │
   │ Approval Notes (optional):                          │
   │ [_________________________________]                 │
   │                                                     │
   │            [Reject with Reason] [✓ Approve]        │
   └─────────────────────────────────────────────────────┘

3. Click [✓ Approve]
   POST /api/v1/web/events/{event_id}/approve
   { "notes": "Looks good!" }

4. Response: Success
   → Event status = published
   → Organizer notified via email
   → Notification banner: "Basketball Cup 2026 approved and published"
```

### Flow: Reject Event

```
1. Click [Reject with Reason]

2. Modal: Reject Event
   ┌───────────────────────────────────────────┐
   │ Reject: Basketball Cup 2026               │
   ├───────────────────────────────────────────┤
   │ Rejection Reason (required):              │
   │ ┌────────────────────────────────────┐    │
   │ │ Informasi venue tidak lengkap.     │    │
   │ │ Mohon cantumkan alamat lengkap...  │    │
   │ └────────────────────────────────────┘    │
   │                                           │
   │              [Cancel] [Reject & Notify]  │
   └───────────────────────────────────────────┘

3. POST /api/v1/web/events/{event_id}/reject
   { "reason": "Informasi venue tidak lengkap..." }

4. Response: Success
   → Event status = rejected
   → Organizer notified dengan rejection reason
   → Organizer bisa edit & resubmit
```

---

## Important Notes

### ✅ DO:
- ✅ Upload images terpisah dulu, kirim sebagai URL saat create/edit
- ✅ Image pertama dalam array auto-set sebagai cover image
- ✅ Event scope global — tidak ada `region_id` di events
- ✅ Gunakan `event_type` untuk bedakan offline/online/hybrid
- ✅ `venue_name` + `venue_address` wajib diisi untuk offline/hybrid
- ✅ `online_platform` + `online_url` wajib diisi untuk online/hybrid
- ✅ Superadmin buat event langsung published tanpa approval
- ✅ Selalu sertakan rejection reason yang jelas dan actionable
- ✅ Log semua approval/rejection actions (audit trail)

### ❌ DON'T:
- ❌ Jangan include `region_id` dalam request/response event
- ❌ Tidak ada field `favorite_count` — hanya `save_count` dan `share_count`
- ❌ Jangan approve event tanpa review isi kontennya
- ❌ Jangan reject tanpa memberikan reason yang jelas

### Approval Authorization:

| Action | Superadmin | Admin | Notes |
|--------|-----------|-------|-------|
| View all events | ✅ | ❌ | Backoffice only for superadmin |
| View pending | ✅ | ❌ | |
| Approve event | ✅ | ❌ | Global approval |
| Reject event | ✅ | ❌ | Must provide reason |
| Force cancel | ✅ | ❌ | Admin action |
| Create event | ✅ | ❌ | Langsung published |
| Edit any event | ✅ | ❌ | |
| Delete event | ✅ | ❌ | Hard delete, irreversible |
| Manage categories | ✅ | ❌ | |
| Update settings | ✅ | ❌ | |

---

## Error Handling

Standard error responses:

```json
// 400 Bad Request
{
  "message": "Only pending_approval events can be approved"
}

// 401 Unauthorized
{
  "message": "Authentication required"
}

// 403 Forbidden
{
  "message": "Only superadmin can manage events in backoffice"
}

// 404 Not Found
{
  "message": "Event not found"
}

// 422 Unprocessable Entity
{
  "message": "Validation failed",
  "errors": {
    "reason": ["Rejection reason is required"],
    "event_date": ["Event date must be in the future"]
  }
}
```

| Scenario | HTTP | Reason |
|----------|------|--------|
| Approve non-pending event | 400 | Wrong status |
| Reject tanpa reason | 422 | Validation failed |
| venue_name kosong untuk offline event | 422 | Validation failed |
| online_url kosong untuk online event | 422 | Validation failed |
| Event not found | 404 | Invalid event_id |
| Non-superadmin access | 403 | Authorization failed |
| Image upload > 5 files | 422 | Upload limit per request |
| Delete published event | 200 | Hard delete, tapi warning dulu di UI |

---

*API spec event module untuk web backoffice. Superadmin only.*
