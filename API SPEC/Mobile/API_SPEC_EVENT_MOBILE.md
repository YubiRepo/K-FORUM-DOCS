# API Spec — Event Module (Mobile)

Dokumentasi API event module untuk aplikasi mobile — create, browse, save, schedule, dan share events.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/events`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (kecuali List & Detail Events)
- **Authorization**:
  - Create/Edit/Cancel Event: Member Pro only
  - View/Save/Schedule/Share: All authenticated users
- **Error Format**: Same as other mobile APIs (standard message or validation error)

---

## Image Upload Flow

Sebelum create/edit event, images harus diupload terlebih dahulu via endpoint upload. Backend returns URL yang kemudian dikirim saat create/edit event.

```
┌────────────────────────────────────────────────────────┐
│ IMAGE UPLOAD FLOW                                      │
├────────────────────────────────────────────────────────┤
│                                                        │
│  Step 1: Upload image(s)                               │
│  POST /api/v1/mobile/media/upload                      │
│  → Returns: { urls: ["https://cdn.../img1.jpg", ...] } │
│                                                        │
│  Step 2: Create/Edit event dengan image URLs           │
│  POST /api/v1/mobile/events                            │
│  { "images": ["https://cdn.../img1.jpg", ...] }        │
│                                                        │
└────────────────────────────────────────────────────────┘
```

---

## Model Data Utama

### 1. Event Object (List)

```json
{
  "id": "uuid",
  "title": "Futsal Tournament 2026",
  "cover_image": "https://cdn.example.com/events/img1.jpg",
  "event_type": "offline",
  "venue_name": "GOR Senayan",
  "venue_address": "Jl. Pintu I Senayan, Jakarta",
  "event_date": "2026-06-15",
  "event_time": "14:00",
  "category": {
    "id": "uuid",
    "name": "Sports"
  },
  "organizer": {
    "id": "uuid",
    "name": "Andi Pratama",
    "avatar": "https://cdn.example.com/avatars/andi.jpg"
  },
  "is_saved": false,
  "is_scheduled": false,
  "status": "published"
}
```

### 2. Event Object (Detail)

```json
{
  "id": "uuid",
  "title": "Futsal Tournament 2026",
  "description": "Turnamen futsal tahunan terbesar di Jakarta...",
  "images": [
    "https://cdn.example.com/events/img1.jpg",
    "https://cdn.example.com/events/img2.jpg",
    "https://cdn.example.com/events/img3.jpg"
  ],
  "cover_image": "https://cdn.example.com/events/img1.jpg",
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
    "avatar": "https://cdn.example.com/avatars/andi.jpg"
  },
  "status": "published",
  "is_saved": true,
  "is_scheduled": true,
  "save_count": 45,
  "share_count": 8,
  "published_at": "2026-05-20T10:00:00.000Z",
  "created_at": "2026-05-19T09:00:00.000Z"
}
```

> Untuk event **online**: `venue_name` dan `venue_address` bernilai `null`, `online_platform` dan `online_url` diisi.
> Untuk event **hybrid**: semua field venue dan online diisi.

---

## Endpoints

### 1. Upload Image(s)

Upload satu atau lebih gambar untuk event. Harus dilakukan **sebelum** create/edit event. Returns URL yang siap dipakai.

- **URL**: `POST /api/v1/mobile/media/upload`
- **Autentikasi**: Yes (Member Pro)
- **Content-Type**: `multipart/form-data`

- **Request**:
  ```
  files[]: image1.jpg  (max 5 files, max 5MB per file)
  files[]: image2.jpg
  context: "event"     (form field, untuk organizing di storage)
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

- **Response (Error — 422)**:
  ```json
  {
    "message": "Validation failed",
    "errors": {
      "files": ["Maximum 5 files per upload"],
      "files.0": ["File size exceeds 5MB limit"],
      "files.1": ["Only image files are allowed (jpg, png, webp)"]
    }
  }
  ```

---

### 2. Create Event

Buat event baru. Images dikirim sebagai array URL hasil dari endpoint upload.

- **URL**: `POST /api/v1/mobile/events`
- **Autentikasi**: Yes (Member Pro only)
- **Authorization**: Requires `subscription_plan = pro` dan permission `create_event`

- **Request Body (Offline)**:
  ```json
  {
    "title": "Futsal Tournament 2026",
    "description": "Turnamen futsal tahunan terbesar di Jakarta...",
    "images": [
      "https://cdn.example.com/events/uploads/img_uuid1.jpg",
      "https://cdn.example.com/events/uploads/img_uuid2.jpg"
    ],
    "category_id": "uuid_sports",
    "event_type": "offline",
    "venue_name": "GOR Senayan",
    "venue_address": "Jl. Pintu I Senayan, Jakarta Pusat",
    "event_date": "2026-06-15",
    "event_end_date": null,
    "event_time": "14:00",
    "registration_url": "https://eventbrite.com/e/futsal-2026"
  }
  ```

- **Request Body (Online)**:
  ```json
  {
    "title": "Webinar Product Management 2026",
    "description": "...",
    "images": [],
    "category_id": "uuid_business",
    "event_type": "online",
    "online_platform": "Zoom",
    "online_url": "https://zoom.us/j/123456789",
    "event_date": "2026-06-20",
    "event_end_date": null,
    "event_time": "10:00",
    "registration_url": "https://eventbrite.com/e/webinar-pm"
  }
  ```

- **Request Body (Hybrid)**:
  ```json
  {
    "title": "KAI Annual Summit 2026",
    "description": "...",
    "images": ["https://cdn.example.com/events/uploads/img_uuid1.jpg"],
    "category_id": "uuid_community",
    "event_type": "hybrid",
    "venue_name": "Jakarta Convention Center",
    "venue_address": "Jl. Gatot Subroto, Jakarta",
    "online_platform": "YouTube Live",
    "online_url": "https://youtube.com/live/abc123",
    "event_date": "2026-07-10",
    "event_end_date": "2026-07-11",
    "event_time": "09:00",
    "registration_url": null
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

- **Response (Success 201 — Auto Published)**:
  ```json
  {
    "data": {
      "id": "event_123",
      "title": "Futsal Tournament 2026",
      "status": "published",
      "cover_image": "https://cdn.example.com/events/uploads/img_uuid1.jpg",
      "published_at": "2026-05-25T10:00:00.000Z",
      "created_at": "2026-05-25T10:00:00.000Z"
    },
    "message": "Event created and published successfully"
  }
  ```

- **Response (Success 201 — Pending Approval)**:
  ```json
  {
    "data": {
      "id": "event_124",
      "title": "Futsal Tournament 2026",
      "status": "pending_approval",
      "cover_image": "https://cdn.example.com/events/uploads/img_uuid1.jpg",
      "created_at": "2026-05-25T10:00:00.000Z"
    },
    "message": "Event submitted for approval"
  }
  ```

- **Response (Error — 403 Forbidden)**:
  ```json
  {
    "message": "Upgrade to Pro to create events"
  }
  ```

---

### 3. List Events (Published)

Ambil daftar events yang sudah published. Dapat difilter by tipe, venue/kota, kategori, tanggal.

- **URL**: `GET /api/v1/mobile/events`
- **Autentikasi**: Optional (is_saved/is_scheduled hanya muncul kalau authenticated)
- **Query Parameters**:
  - `search` (optional): Cari by title, venue_name, atau venue_address
  - `category_id` (optional): Filter by kategori
  - `event_type` (optional): `offline`, `online`, `hybrid`
  - `location` (optional): Filter by kota/area — text match pada `venue_address`
  - `date_from` (optional): Filter from date `YYYY-MM-DD`
  - `date_to` (optional): Filter to date `YYYY-MM-DD`
  - `sort` (optional): `event_date`, `created_at`, `-event_date`, `-created_at` (default: `event_date`)
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
        "event_time": "14:00",
        "category": { "id": "uuid", "name": "Sports" },
        "organizer": {
          "id": "uuid",
          "name": "Andi Pratama",
          "avatar": "https://cdn.example.com/avatars/andi.jpg"
        },
        "registration_url": "https://eventbrite.com/e/futsal-2026",
        "is_saved": false,
        "is_scheduled": false
      },
      {
        "id": "event_125",
        "title": "Webinar Product Management 2026",
        "cover_image": null,
        "event_type": "online",
        "venue_name": null,
        "venue_address": null,
        "event_date": "2026-06-20",
        "event_time": "10:00",
        "category": { "id": "uuid", "name": "Business" },
        "organizer": {
          "id": "uuid",
          "name": "Budi Santoso",
          "avatar": "https://cdn.example.com/avatars/budi.jpg"
        },
        "registration_url": "https://eventbrite.com/e/webinar-pm",
        "is_saved": true,
        "is_scheduled": false
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 87
    }
  }
  ```

---

### 4. View Event Detail

Ambil detail lengkap satu event.

- **URL**: `GET /api/v1/mobile/events/{event_id}`
- **Autentikasi**: Optional

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "event_123",
      "title": "Futsal Tournament 2026",
      "description": "Turnamen futsal tahunan terbesar di Jakarta...",
      "images": [
        "https://cdn.example.com/events/uploads/img_uuid1.jpg",
        "https://cdn.example.com/events/uploads/img_uuid2.jpg"
      ],
      "cover_image": "https://cdn.example.com/events/uploads/img_uuid1.jpg",
      "event_type": "offline",
      "venue_name": "GOR Senayan",
      "venue_address": "Jl. Pintu I Senayan, Jakarta Pusat",
      "online_platform": null,
      "online_url": null,
      "category": { "id": "uuid", "name": "Sports" },
      "event_date": "2026-06-15",
      "event_end_date": null,
      "event_time": "14:00",
      "registration_url": "https://eventbrite.com/e/futsal-2026",
      "organizer": {
        "id": "uuid",
        "name": "Andi Pratama",
        "avatar": "https://cdn.example.com/avatars/andi.jpg"
      },
      "status": "published",
      "is_saved": true,
      "is_scheduled": true,
      "save_count": 45,
      "share_count": 8,
      "published_at": "2026-05-20T10:00:00.000Z",
      "created_at": "2026-05-19T09:00:00.000Z"
    }
  }
  ```

---

### 5. Edit Own Event

Edit event milik sendiri. Hanya bisa edit event berstatus `draft` atau `rejected`.

- **URL**: `PUT /api/v1/mobile/events/{event_id}`
- **Autentikasi**: Yes (Member Pro, owner only)

- **Request Body**:
  ```json
  {
    "title": "Futsal Tournament 2026 — Updated",
    "description": "...",
    "images": [
      "https://cdn.example.com/events/uploads/img_new1.jpg"
    ],
    "event_type": "offline",
    "venue_name": "GOR Sumantri Brojonegoro",
    "venue_address": "Jl. HR Rasuna Said, Kuningan, Jakarta",
    "event_date": "2026-06-20",
    "event_time": "15:00"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "event_123",
      "status": "draft",
      "cover_image": "https://cdn.example.com/events/uploads/img_new1.jpg",
      "updated_at": "2026-05-25T11:00:00.000Z"
    },
    "message": "Event updated successfully"
  }
  ```

- **Response (Error — 403)**:
  ```json
  {
    "message": "Only draft or rejected events can be edited"
  }
  ```

---

### 6. Cancel Event

Cancel event yang sudah dibuat (bisa dari draft atau published).

- **URL**: `POST /api/v1/mobile/events/{event_id}/cancel`
- **Autentikasi**: Yes (Member Pro, owner only)

- **Request Body**:
  ```json
  {
    "reason": "Venue tidak tersedia"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "event_123",
      "status": "cancelled",
      "reason": "Venue tidak tersedia",
      "cancelled_at": "2026-05-25T11:00:00.000Z"
    },
    "message": "Event cancelled"
  }
  ```

---

### 7. Get My Events

Ambil daftar events milik sendiri (semua status).

- **URL**: `GET /api/v1/mobile/events/me`
- **Autentikasi**: Yes (Member Pro)
- **Query Parameters**:
  - `status` (optional): `draft`, `pending_approval`, `published`, `rejected`, `cancelled`
  - `limit` (optional, default: 20)
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
        "status": "published",
        "event_date": "2026-06-15",
        "created_at": "2026-05-19T09:00:00.000Z"
      },
      {
        "id": "event_124",
        "title": "Webinar PM",
        "cover_image": null,
        "event_type": "online",
        "venue_name": null,
        "status": "pending_approval",
        "event_date": "2026-07-01",
        "created_at": "2026-05-24T08:00:00.000Z"
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

### 8. Save / Bookmark Event

Simpan event ke daftar bookmark.

- **URL**: `POST /api/v1/mobile/events/{event_id}/save`
- **Autentikasi**: Yes

- **Request Body**:
  ```json
  {
    "note": "Menarik, coba daftar nanti"
  }
  ```

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "event_save_123",
      "event_id": "event_123",
      "saved_at": "2026-05-25T10:00:00.000Z",
      "note": "Menarik, coba daftar nanti"
    },
    "message": "Event saved"
  }
  ```

---

### 9. Unsave Event

Hapus event dari daftar bookmark.

- **URL**: `DELETE /api/v1/mobile/events/{event_id}/save`
- **Autentikasi**: Yes

- **Response (Success 200)**:
  ```json
  {
    "message": "Event unsaved"
  }
  ```

---

### 10. Add to Schedule

Tambahkan event ke jadwal pribadi. User bisa sekaligus meminta data untuk export ke kalender eksternal.

- **URL**: `POST /api/v1/mobile/events/{event_id}/schedule`
- **Autentikasi**: Yes

- **Request Body**:
  ```json
  {
    "reminder_enabled": true,
    "reminder_time": "1 day before",
    "personal_notes": "Invite Budi & Citra",
    "export_calendar": true
  }
  ```

- **Request Validation**:
  - `reminder_enabled`: Optional, boolean, default `false`
  - `reminder_time`: Optional, enum: `1 hour before`, `3 hours before`, `1 day before`, `3 days before`, `1 week before`
  - `personal_notes`: Optional, string max 500 chars
  - `export_calendar`: Optional, boolean. Jika `true`, response menyertakan `calendar_export`

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "event_schedule_123",
      "event_id": "event_123",
      "scheduled_at": "2026-05-25T10:00:00.000Z",
      "status": "scheduled",
      "reminder_enabled": true,
      "reminder_time": "1 day before",
      "calendar_export": {
        "ics_url": "https://api.example.com/events/event_123/calendar.ics",
        "google_calendar_url": "https://calendar.google.com/calendar/r/eventedit?text=Futsal+Tournament+2026&dates=20260615T140000/20260615T160000&location=GOR+Senayan&details=...",
        "apple_calendar_url": "webcal://api.example.com/events/event_123/calendar.ics"
      }
    },
    "message": "Event added to schedule"
  }
  ```

  > Jika `export_calendar: false` atau tidak dikirim, field `calendar_export` tidak muncul di response.

---

### 11. Get Calendar Export Links

Ambil ulang link export kalender untuk event yang sudah dijadwalkan — berguna jika user ingin export ke kalender lain di lain waktu.

- **URL**: `GET /api/v1/mobile/events/{event_id}/schedule/calendar`
- **Autentikasi**: Yes

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "event_id": "event_123",
      "title": "Futsal Tournament 2026",
      "event_date": "2026-06-15",
      "event_time": "14:00",
      "calendar_export": {
        "ics_url": "https://api.example.com/events/event_123/calendar.ics",
        "google_calendar_url": "https://calendar.google.com/calendar/r/eventedit?...",
        "apple_calendar_url": "webcal://api.example.com/events/event_123/calendar.ics"
      }
    }
  }
  ```

- **Response (Error — 404)**:
  ```json
  {
    "message": "Event not in your schedule"
  }
  ```

---

### 12. Remove from Schedule

Hapus event dari jadwal pribadi.

- **URL**: `DELETE /api/v1/mobile/events/{event_id}/schedule`
- **Autentikasi**: Yes

- **Response (Success 200)**:
  ```json
  {
    "message": "Event removed from schedule"
  }
  ```

---

### 13. Share Event

Catat event yang di-share (untuk analytics & share link generation).

- **URL**: `POST /api/v1/mobile/events/{event_id}/share`
- **Autentikasi**: Yes

- **Request Body**:
  ```json
  {
    "share_method": "whatsapp",
    "message": "Ayok ikutan futsal tournament!"
  }
  ```

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "event_share_123",
      "share_method": "whatsapp",
      "shared_at": "2026-05-25T10:00:00.000Z",
      "share_link": "https://kai.app/events/event_123?ref=user_abc"
    },
    "message": "Event shared"
  }
  ```

---

### 14. Get My Saved Events

Ambil daftar event yang sudah di-bookmark.

- **URL**: `GET /api/v1/mobile/events/me/saved`
- **Autentikasi**: Yes
- **Query Parameters**:
  - `limit` (optional, default: 20)
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
        "event_date": "2026-06-15",
        "saved_at": "2026-05-25T10:00:00.000Z",
        "note": "Menarik, coba daftar nanti"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 3
    }
  }
  ```

---

### 15. Get My Schedule

Ambil daftar event yang ada di jadwal pribadi.

- **URL**: `GET /api/v1/mobile/events/me/schedule`
- **Autentikasi**: Yes
- **Query Parameters**:
  - `status` (optional): `scheduled`, `done`, `skipped`
  - `limit` (optional, default: 20)
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
        "event_date": "2026-06-15",
        "event_time": "14:00",
        "status": "scheduled",
        "reminder_enabled": true,
        "reminder_time": "1 day before",
        "personal_notes": "Invite Budi & Citra"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 1
    }
  }
  ```

---

### 16. Get Event Categories

Ambil daftar kategori event untuk dropdown saat create event.

- **URL**: `GET /api/v1/mobile/events/categories`
- **Autentikasi**: Optional

- **Response (Success 200)**:
  ```json
  {
    "data": [
      { "id": "uuid", "name": "Sports" },
      { "id": "uuid", "name": "Business" },
      { "id": "uuid", "name": "Cultural" },
      { "id": "uuid", "name": "Educational" },
      { "id": "uuid", "name": "Community" },
      { "id": "uuid", "name": "Other" }
    ]
  }
  ```

---

## UI Flow Example

### Flow: Create Event Offline dengan Multiple Images

```
1. User buka "Create Event"

2. User upload 3 gambar:
   POST /api/v1/mobile/media/upload
   → Response: {
       "urls": ["https://cdn.../img1.jpg", "https://cdn.../img2.jpg", "https://cdn.../img3.jpg"]
     }

3. User isi form:
   ┌─────────────────────────────────────────────────────┐
   │ Create Event                                        │
   ├─────────────────────────────────────────────────────┤
   │ Title:        [Futsal Tournament 2026]              │
   │ Description:  [...]                                 │
   │ Images:       [img1 ✓] [img2 ✓] [img3 ✓] [+ Add]  │
   │               (cover)                               │
   │ Category:     [Sports ▼]                            │
   │                                                     │
   │ Tipe Event:   ( ) Offline  ( ) Online  ( ) Hybrid  │
   │               [● Offline]                           │
   │                                                     │
   │ Nama Venue:   [GOR Senayan]                         │
   │ Alamat:       [Jl. Pintu I Senayan, Jakarta Pusat]  │
   │                                                     │
   │ Tanggal:      [2026-06-15]                          │
   │ Waktu:        [14:00]                               │
   │ Reg. Link:    [https://eventbrite.com/...]          │
   │                                                     │
   │               [Simpan Draft] [Publish / Submit]     │
   └─────────────────────────────────────────────────────┘

4. POST /api/v1/mobile/events → success
```

### Flow: Add to Schedule + Export ke Google Calendar

```
1. User buka detail event "Futsal Tournament 2026"
2. Tap "Tambah ke Jadwal"

3. Modal:
   ┌─────────────────────────────────────────────────────┐
   │ Tambah ke Jadwal                                    │
   ├─────────────────────────────────────────────────────┤
   │ Pengingat:   [✓] Aktifkan                           │
   │ Waktu:       [1 hari sebelum ▼]                     │
   │ Catatan:     [Invite Budi & Citra]                  │
   │                                                     │
   │ Export ke Kalender (opsional):                      │
   │ [✓] Google Calendar                                 │
   │ [ ] Apple Calendar                                  │
   │ [ ] Download .ics                                   │
   │                                                     │
   │                            [Simpan]                 │
   └─────────────────────────────────────────────────────┘

4. Tap [Simpan]
   POST /api/v1/mobile/events/{event_id}/schedule
   {
     "reminder_enabled": true,
     "reminder_time": "1 day before",
     "personal_notes": "Invite Budi & Citra",
     "export_calendar": true
   }

5. Response menyertakan google_calendar_url
   → App buka Google Calendar URL di browser/deep link
   → Event langsung masuk ke Google Calendar user
```

---

## Important Notes

### ✅ DO:
- ✅ Upload images terpisah dulu sebelum create/edit event
- ✅ Image pertama dalam array auto-set sebagai cover image
- ✅ Event scope global — tidak ada `region_id`
- ✅ Gunakan `event_type` untuk bedakan offline/online/hybrid
- ✅ `venue_name` + `venue_address` untuk offline/hybrid (string bebas)
- ✅ `online_platform` + `online_url` untuk online/hybrid
- ✅ Sertakan `export_calendar: true` untuk dapatkan calendar export links

### ❌ DON'T:
- ❌ Jangan kirim base64 atau file langsung saat create/edit event
- ❌ Jangan include `region_id` dalam request body event
- ❌ Member standard tidak bisa create event
- ❌ Jangan edit event yang sedang `pending_approval` atau sudah `published`
- ❌ Tidak ada fitur favorite — hanya save/bookmark

### Status Flow:

```
draft ──→ pending_approval ──→ published
  │                                │
  └──→ DELETE                   CANCEL
                ↑
          rejected ←────────────────┘
              │
         (edit & resubmit)
```

---

## Error Handling

Standard error responses:

```json
// 401 Unauthorized
{
  "message": "Authentication required"
}

// 403 Forbidden
{
  "message": "Upgrade to Pro to create events"
}

// 404 Not Found
{
  "message": "Event not found"
}

// 422 Unprocessable Entity
{
  "message": "Validation failed",
  "errors": {
    "venue_name": ["Venue name is required for offline events"],
    "online_url": ["Online URL is required for online events"],
    "event_date": ["Event date must be in the future"],
    "images": ["Maximum 10 images allowed"]
  }
}
```

| Scenario | HTTP | Reason |
|----------|------|--------|
| Member standard create event | 403 | Pro plan required |
| Upload > 5 files sekaligus | 422 | Upload limit per request |
| Image > 5MB | 422 | File size exceeded |
| Event date di masa lalu | 422 | Validation failed |
| venue_name kosong untuk offline event | 422 | Validation failed |
| online_url kosong untuk online event | 422 | Validation failed |
| Edit published event | 403 | Status not editable |
| Event not found | 404 | Invalid event_id |
| Unauthenticated save/schedule | 401 | Auth required |

---

*API spec event module untuk mobile. Upload images terpisah sebelum create/edit event.*
