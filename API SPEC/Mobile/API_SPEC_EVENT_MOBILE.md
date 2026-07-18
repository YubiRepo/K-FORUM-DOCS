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

## Image Upload Flow (Presigned URL)

Sebelum create/edit event, images harus diupload terlebih dahulu via **presigned URL flow**. Client mendapat presigned URL + `s3:` key, upload langsung ke S3, lalu konfirmasi. `s3:` key inilah yang dikirim saat create/edit event.

```
┌──────────────────────────────────────────────────────────────────┐
│ IMAGE UPLOAD FLOW (Presigned URL)                                │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Step 1: Request presigned URL + s3 key                          │
│  POST /api/v1/mobile/media/presign                               │
│  → Returns: { url: "https://s3....", key: "s3:/events/..." }    │
│                                                                  │
│  Step 2: Upload file langsung ke S3 via presigned URL            │
│  PUT <presigned_url>                                             │
│  (Body: file binary, Content-Type: image/jpeg)                   │
│                                                                  │
│  Step 3: Confirm upload selesai                                  │
│  POST /api/v1/mobile/media/confirm                               │
│  → Returns: { status: "confirmed", key: "s3:/events/..." }      │
│                                                                  │
│  Step 4: Create/Edit event dengan s3 key                         │
│  POST /api/v1/mobile/events                                      │
│  { "images": ["s3:/events/img1.jpg", ...] }                      │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

Untuk menghapus media yang sudah diupload (sebelum dipakai), gunakan endpoint **Delete Media**.

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
  "timezone": "Asia/Jakarta",
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
  "is_featured": false,
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
  "timezone": "Asia/Jakarta",
  "registration_url": "https://eventbrite.com/e/futsal-2026",
  "organizer": {
    "id": "uuid",
    "name": "Andi Pratama",
    "avatar": "https://cdn.example.com/avatars/andi.jpg"
  },
  "status": "published",
  "is_saved": true,
  "is_scheduled": true,
  "is_featured": false,
  "save_count": 45,
  "share_count": 8,
  "published_at": "2026-05-20T10:00:00.000Z",
  "created_at": "2026-05-19T09:00:00.000Z"
}
```

> `timezone`: IANA timezone identifier tempat event berlangsung (contoh: `Asia/Jakarta`, `Asia/Makassar`, `Asia/Jayapura`). Wajib diisi saat create/edit — dipakai untuk menghitung reminder dan export kalender dalam waktu absolut (UTC) yang benar.
> Untuk event **online**: `venue_name` dan `venue_address` bernilai `null`, `online_platform` dan `online_url` diisi.
> Untuk event **hybrid**: semua field venue dan online diisi.

### 3. Event Feedback Object

```json
{
  "id": "uuid",
  "event_id": "event_123",
  "user": {
    "id": "uuid",
    "name": "Doni Saputra",
    "avatar": "https://cdn.example.com/avatars/doni.jpg"
  },
  "rating": 5,
  "venue_rating": 4,
  "organization_rating": 5,
  "would_recommend": true,
  "comment": "Acaranya seru banget, venue nyaman",
  "is_anonymous": false,
  "created_at": "2026-06-16T09:00:00.000Z",
  "updated_at": "2026-06-16T09:00:00.000Z"
}
```

> Kalau `is_anonymous: true`, field `user` bernilai `null` di response yang dilihat organizer/superadmin — isi feedback tetap tampil apa adanya.

### 4. Event Feedback Summary Object

```json
{
  "event_id": "event_123",
  "total_feedback": 42,
  "average_rating": 4.6,
  "average_venue_rating": 4.3,
  "average_organization_rating": 4.7,
  "recommend_percentage": 92.9,
  "rating_distribution": {
    "5": 30,
    "4": 8,
    "3": 3,
    "2": 1,
    "1": 0
  }
}
```

---

## Endpoints

### 1a. Request Presigned URL

Mendapatkan presigned URL dan `s3:` key untuk upload file langsung ke S3. Satu request = satu file. Untuk multiple files, lakukan berulang.

- **URL**: `POST /api/v1/mobile/media/presign`
- **Autentikasi**: Yes (Member Pro)
- **Content-Type**: `application/json`

- **Request**:
  ```json
  {
    "filename": "img1.jpg",
    "mime_type": "image/jpeg",
    "context": "event"
  }
  ```

- **Validation**:
  - `filename`: Required, max 200 chars, must have image extension (jpg, png, webp)
  - `mime_type`: Required, must match extension
  - `context`: Required, enum: `event`, `subscription_proof`, `avatar`, `news`

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "url": "https://s3.ap-southeast-1.amazonaws.com/kai-uploads/events/abc123?X-Amz-Algorithm=...",
      "key": "s3:/events/uploads/img_uuid1.jpg",
      "expires_in": 3600
    },
    "message": "Presigned URL generated"
  }
  ```

- **Response (Error — 422)**:
  ```json
  {
    "message": "Validation failed",
    "errors": {
      "filename": ["Invalid file extension. Allowed: jpg, png, webp"]
    }
  }
  ```

---

### 1b. Confirm Upload

Konfirmasi bahwa file sudah berhasil diupload ke S3. Backend akan memverifikasi keberadaan file dan mencatatnya di database.

- **URL**: `POST /api/v1/mobile/media/confirm`
- **Autentikasi**: Yes (Member Pro)
- **Content-Type**: `application/json`

- **Request**:
  ```json
  {
    "key": "s3:/events/uploads/img_uuid1.jpg"
  }
  ```

- **Validation**:
  - `key`: Required, must start with `s3:`

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "key": "s3:/events/uploads/img_uuid1.jpg",
      "status": "confirmed"
    },
    "message": "Upload confirmed"
  }
  ```

- **Response (Error — 404)**:
  ```json
  {
    "message": "File not found on storage. Upload may have failed."
  }
  ```

---

### 1c. Delete Media

Menghapus media yang sudah diupload (sebelum digunakan di create/edit). Media yang sudah terpakai di event tidak bisa dihapus — hapus dari event dulu.

- **URL**: `DELETE /api/v1/mobile/media`
- **Autentikasi**: Yes (Member Pro)
- **Content-Type**: `application/json`

- **Request**:
  ```json
  {
    "key": "s3:/events/uploads/img_uuid1.jpg"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "message": "Media deleted successfully"
  }
  ```

- **Response (Error — 409)**:
  ```json
  {
    "message": "Media is in use by an existing event. Remove it from the event first."
  }
  ```

---

### 2. Create Event

Buat event baru. Images dikirim sebagai array `s3:` key hasil dari presigned URL flow.

- **URL**: `POST /api/v1/mobile/events`
- **Autentikasi**: Yes (Member Pro only)
- **Authorization**: Requires `subscription_plan = pro` dan permission `create_event`

- **Request Body (Offline)**:
  ```json
  {
    "title": "Futsal Tournament 2026",
    "description": "Turnamen futsal tahunan terbesar di Jakarta...",
    "images": [
      "s3:/events/uploads/img_uuid1.jpg",
      "s3:/events/uploads/img_uuid2.jpg"
    ],
    "category_id": "uuid_sports",
    "event_type": "offline",
    "venue_name": "GOR Senayan",
    "venue_address": "Jl. Pintu I Senayan, Jakarta Pusat",
    "event_date": "2026-06-15",
    "event_end_date": null,
    "event_time": "14:00",
    "timezone": "Asia/Jakarta",
    "registration_url": "ext:https://eventbrite.com/e/futsal-2026"
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
    "online_url": "ext:https://zoom.us/j/123456789",
    "event_date": "2026-06-20",
    "event_end_date": null,
    "event_time": "10:00",
    "timezone": "Asia/Jakarta",
    "registration_url": "ext:https://eventbrite.com/e/webinar-pm"
  }
  ```

- **Request Body (Hybrid)**:
  ```json
  {
    "title": "KAI Annual Summit 2026",
    "description": "...",
    "images": ["s3:/events/uploads/img_uuid1.jpg"],
    "category_id": "uuid_community",
    "event_type": "hybrid",
    "venue_name": "Jakarta Convention Center",
    "venue_address": "Jl. Gatot Subroto, Jakarta",
    "online_platform": "YouTube Live",
    "online_url": "ext:https://youtube.com/live/abc123",
    "event_date": "2026-07-10",
    "event_end_date": "2026-07-11",
    "event_time": "09:00",
    "timezone": "Asia/Jayapura",
    "registration_url": null
  }
  ```

- **Request Validation**:
  - `title`: Required, string, max 200 chars
  - `description`: Required, string
  - `images`: Optional, array of `s3:` keys, max 10 items. First item auto-set sebagai cover image
  - `category_id`: Required, must exist
  - `event_type`: Required, enum: `offline`, `online`, `hybrid`
  - `venue_name`: Required jika `event_type` adalah `offline` atau `hybrid`
  - `venue_address`: Required jika `event_type` adalah `offline` atau `hybrid`
  - `online_platform`: Required jika `event_type` adalah `online` atau `hybrid`
  - `online_url`: Required jika `event_type` adalah `online` atau `hybrid`, valid URL
  - `event_date`: Required, must be future date, format `YYYY-MM-DD`
  - `event_end_date`: Optional, must be >= `event_date`
  - `event_time`: Required, format `HH:mm`
  - `timezone`: **Required** (mandatory, breaking change per 2026-07-17), string, must be a valid IANA timezone identifier (validated server-side via Go `time.LoadLocation`). Contoh nilai valid: `Asia/Jakarta`, `Asia/Makassar`, `Asia/Jayapura`. Dipakai untuk mengubah `event_date`+`event_time` (wall-clock lokal event) menjadi instant UTC absolut untuk reminder scheduling & calendar export
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
        "timezone": "Asia/Jakarta",
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
        "timezone": "Asia/Jakarta",
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

### 4. Get Featured Events

Ambil daftar event yang di-highlight (featured) oleh Superadmin — buat carousel/banner di home. Status featured murni dikontrol dari backoffice; mobile cuma konsumsi.

- **URL**: `GET /api/v1/mobile/events/featured`
- **Autentikasi**: Optional (`is_saved`/`is_scheduled` hanya muncul kalau authenticated)
- **Query Parameters**:
  - `limit` (optional, default: 5, max: 20)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "event_125",
        "title": "Official Platform Event 2026",
        "cover_image": "https://cdn.example.com/events/uploads/img_uuid1.jpg",
        "event_type": "hybrid",
        "venue_name": "Jakarta Convention Center",
        "venue_address": "Jl. Gatot Subroto, Jakarta Selatan",
        "event_date": "2026-08-01",
        "event_time": "09:00",
        "timezone": "Asia/Jakarta",
        "category": { "id": "uuid", "name": "Community" },
        "organizer": {
          "id": "uuid",
          "name": "Admin KAI",
          "avatar": "https://cdn.example.com/avatars/admin.jpg"
        },
        "registration_url": "https://eventbrite.com/e/official-event-2026",
        "is_saved": false,
        "is_scheduled": false,
        "is_featured": true
      }
    ]
  }
  ```

> Hanya event berstatus `published` yang muncul. Default urut by `event_date` ascending (event terdekat duluan).

---

### 5. Get Upcoming Events

Ambil daftar event yang akan datang, diurutkan dari yang paling dekat (`event_date` ascending). Hanya event `published` yang `event_date`-nya >= hari ini.

- **URL**: `GET /api/v1/mobile/events/upcoming`
- **Autentikasi**: Optional (`is_saved`/`is_scheduled` hanya muncul kalau authenticated)
- **Query Parameters**:
  - `category_id` (optional): Filter by kategori
  - `event_type` (optional): `offline`, `online`, `hybrid`
  - `limit` (optional, default: 10, max: 50)
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
        "timezone": "Asia/Jakarta",
        "category": { "id": "uuid", "name": "Sports" },
        "organizer": {
          "id": "uuid",
          "name": "Andi Pratama",
          "avatar": "https://cdn.example.com/avatars/andi.jpg"
        },
        "registration_url": "https://eventbrite.com/e/futsal-2026",
        "is_saved": false,
        "is_scheduled": false,
        "is_featured": false
      }
    ],
    "pagination": {
      "limit": 10,
      "offset": 0,
      "total": 42
    }
  }
  ```

---

### 6. View Event Detail

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
      "timezone": "Asia/Jakarta",
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

### 7. Edit Own Event

Edit event milik sendiri. Images menggunakan `s3:` key hasil dari presigned URL flow. Hanya bisa edit event berstatus `draft` atau `rejected`.

- **URL**: `PUT /api/v1/mobile/events/{event_id}`
- **Autentikasi**: Yes (Member Pro, owner only)

- **Request Body**:
  ```json
  {
    "title": "Futsal Tournament 2026 — Updated",
    "description": "...",
    "images": [
      "s3:/events/uploads/img_new1.jpg"
    ],
    "event_type": "offline",
    "venue_name": "GOR Sumantri Brojonegoro",
    "venue_address": "Jl. HR Rasuna Said, Kuningan, Jakarta",
    "event_date": "2026-06-20",
    "event_time": "15:00",
    "timezone": "Asia/Jakarta"
  }
  ```

  > `timezone` **wajib** dikirim juga saat edit (bukan hanya create) — sama seperti `event_date`/`event_time`, memakai IANA identifier (contoh: `Asia/Jakarta`, `Asia/Makassar`, `Asia/Jayapura`).

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

### 8. Cancel Event

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

### 9. Get My Events

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

### 10. Save / Bookmark Event

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

### 11. Unsave Event

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

### 12. Add to Schedule

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

### 13. Get Calendar Export Links

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

### 14. Remove from Schedule

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

### 15. Share Event

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

### 16. Get My Saved Events

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

### 17. Get My Schedule

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
        "timezone": "Asia/Jakarta",
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

### 18. Get Event Categories

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

### 19. Submit Event Feedback

Isi feedback (angket) untuk event yang sudah berlangsung. Tidak bisa dipakai oleh organizer event itu sendiri.

- **URL**: `POST /api/v1/mobile/events/{event_id}/feedback`
- **Autentikasi**: Yes

- **Request Body**:
  ```json
  {
    "rating": 5,
    "venue_rating": 4,
    "organization_rating": 5,
    "would_recommend": true,
    "comment": "Acaranya seru banget, venue nyaman",
    "is_anonymous": false
  }
  ```

- **Request Validation**:
  - `rating`: Required, integer, 1–5
  - `venue_rating`: Optional, integer, 1–5
  - `organization_rating`: Optional, integer, 1–5
  - `would_recommend`: Optional, boolean
  - `comment`: Optional, string, max 1000 chars. Tunduk pada filter kata terlarang platform (banned keywords) — comment yang mengandung kata terlarang ditolak `422`
  - `is_anonymous`: Optional, boolean, default `false`
  - Event harus berstatus `published` dan waktu mulainya (`event_date`+`event_time` sesuai `timezone`) sudah lewat
  - Event masih dalam window feedback (default 30 hari setelah event berlangsung, diatur superadmin)
  - User bukan organizer event tsb
  - User belum pernah submit feedback untuk event ini (satu feedback per user per event — kalau sudah ada, pakai `PUT` untuk edit)

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "feedback_123",
      "event_id": "event_123",
      "rating": 5,
      "venue_rating": 4,
      "organization_rating": 5,
      "would_recommend": true,
      "comment": "Acaranya seru banget, venue nyaman",
      "is_anonymous": false,
      "created_at": "2026-06-16T09:00:00.000Z"
    },
    "message": "Feedback submitted successfully"
  }
  ```

- **Response (Error — 403)**:
  ```json
  {
    "message": "Organizer cannot submit feedback for their own event"
  }
  ```

- **Response (Error — 409)**:
  ```json
  {
    "message": "You have already submitted feedback for this event. Use edit instead."
  }
  ```

- **Response (Error — 422)**:
  ```json
  {
    "message": "Feedback is not open yet — event has not started"
  }
  ```

---

### 20. Get My Feedback for Event

Ambil feedback milik sendiri untuk satu event — dipakai untuk cek apakah sudah submit dan untuk prefill form edit.

- **URL**: `GET /api/v1/mobile/events/{event_id}/feedback/me`
- **Autentikasi**: Yes

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "feedback_123",
      "event_id": "event_123",
      "rating": 5,
      "venue_rating": 4,
      "organization_rating": 5,
      "would_recommend": true,
      "comment": "Acaranya seru banget, venue nyaman",
      "is_anonymous": false,
      "created_at": "2026-06-16T09:00:00.000Z",
      "updated_at": "2026-06-16T09:00:00.000Z"
    }
  }
  ```

- **Response (Error — 404)**:
  ```json
  {
    "message": "You have not submitted feedback for this event"
  }
  ```

---

### 21. Edit My Feedback

Edit feedback milik sendiri. Hanya bisa selama masih dalam window feedback.

- **URL**: `PUT /api/v1/mobile/events/{event_id}/feedback`
- **Autentikasi**: Yes (owner only)

- **Request Body**: Sama seperti Submit Feedback — field yang tidak dikirim tetap memakai nilai sebelumnya.
  ```json
  {
    "rating": 4,
    "comment": "Direvisi: venue agak sempit tapi tetap seru"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "feedback_123",
      "event_id": "event_123",
      "rating": 4,
      "venue_rating": 4,
      "organization_rating": 5,
      "would_recommend": true,
      "comment": "Direvisi: venue agak sempit tapi tetap seru",
      "is_anonymous": false,
      "updated_at": "2026-06-18T10:00:00.000Z"
    },
    "message": "Feedback updated successfully"
  }
  ```

- **Response (Error — 403)**:
  ```json
  {
    "message": "Feedback window has closed for this event"
  }
  ```

---

### 22. Delete My Feedback

Hapus feedback milik sendiri.

- **URL**: `DELETE /api/v1/mobile/events/{event_id}/feedback`
- **Autentikasi**: Yes (owner only)

- **Response (Success 200)**:
  ```json
  {
    "message": "Feedback deleted"
  }
  ```

---

### 23. Get Event Feedback List (Organizer)

Ambil semua feedback untuk event milik sendiri. Hanya organizer event tsb yang boleh akses — user lain (bukan organizer) mendapat `403`.

- **URL**: `GET /api/v1/mobile/events/{event_id}/feedback`
- **Autentikasi**: Yes (organizer/owner only)
- **Query Parameters**:
  - `sort` (optional): `created_at`, `-created_at`, `rating`, `-rating` (default: `-created_at`)
  - `limit` (optional, default: 20, max: 100)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "feedback_123",
        "event_id": "event_123",
        "user": {
          "id": "uuid",
          "name": "Doni Saputra",
          "avatar": "https://cdn.example.com/avatars/doni.jpg"
        },
        "rating": 5,
        "venue_rating": 4,
        "organization_rating": 5,
        "would_recommend": true,
        "comment": "Acaranya seru banget, venue nyaman",
        "is_anonymous": false,
        "created_at": "2026-06-16T09:00:00.000Z"
      },
      {
        "id": "feedback_124",
        "event_id": "event_123",
        "user": null,
        "rating": 4,
        "venue_rating": 3,
        "organization_rating": 4,
        "would_recommend": true,
        "comment": "Cukup baik, parkiran kurang luas",
        "is_anonymous": true,
        "created_at": "2026-06-16T11:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 42
    }
  }
  ```

- **Response (Error — 403)**:
  ```json
  {
    "message": "Only the event organizer can view its feedback"
  }
  ```

---

### 24. Get Event Feedback Summary (Organizer)

Ambil ringkasan statistik feedback untuk event milik sendiri.

- **URL**: `GET /api/v1/mobile/events/{event_id}/feedback/summary`
- **Autentikasi**: Yes (organizer/owner only)

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "event_id": "event_123",
      "total_feedback": 42,
      "average_rating": 4.6,
      "average_venue_rating": 4.3,
      "average_organization_rating": 4.7,
      "recommend_percentage": 92.9,
      "rating_distribution": {
        "5": 30,
        "4": 8,
        "3": 3,
        "2": 1,
        "1": 0
      }
    }
  }
  ```

- **Response (Error — 403)**:
  ```json
  {
    "message": "Only the event organizer can view its feedback"
  }
  ```

---

## UI Flow Example

### Flow: Create Event Offline dengan Multiple Images

```
1. User buka "Create Event"

2. User upload 3 gambar via presigned URL flow (ulangi per file):
   POST /api/v1/mobile/media/presign  →  dapat presigned URL + s3:key
   PUT <presigned_url>                →  upload langsung ke S3
   POST /api/v1/mobile/media/confirm  →  konfirmasi
   → Hasil: 3 s3 key siap pakai

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
   │ Zona Waktu:   [Asia/Jakarta ▼]  (wajib diisi)       │
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

### Flow: Isi Feedback Setelah Event Selesai

```
1. User buka detail event yang sudah lewat tanggalnya
   → Banner muncul: "Bagaimana pengalamanmu di event ini?" [Isi Feedback]

2. Modal Feedback:
   ┌─────────────────────────────────────────────────────┐
   │ Feedback: Futsal Tournament 2026                    │
   ├─────────────────────────────────────────────────────┤
   │ Rating Keseluruhan:     ★★★★★                       │
   │ Rating Venue:           ★★★★☆                       │
   │ Rating Penyelenggaraan: ★★★★★                       │
   │ Rekomendasi:            [✓] Ya  [ ] Tidak            │
   │ Komentar:               [Acaranya seru banget...]   │
   │ [ ] Kirim sebagai anonim                             │
   │                                                     │
   │                                    [Kirim Feedback] │
   └─────────────────────────────────────────────────────┘

3. POST /api/v1/mobile/events/{event_id}/feedback
   { "rating": 5, "venue_rating": 4, "organization_rating": 5,
     "would_recommend": true, "comment": "...", "is_anonymous": false }

4. Response: Success → tombol berubah jadi "Edit Feedback"
   → Organizer mendapat notifikasi feedback baru
```

---

## Important Notes

### ✅ DO:
- ✅ Gunakan presigned URL flow sebelum create/edit event (presign → upload → confirm)
- ✅ Kirim `s3:` key pada field `images` di create/edit request
- ✅ Gunakan `ext:https://...` untuk URL eksternal (registration_url, online_url)
- ✅ Image pertama dalam array auto-set sebagai cover image
- ✅ Event scope global — tidak ada `region_id`
- ✅ Gunakan `event_type` untuk bedakan offline/online/hybrid
- ✅ `venue_name` + `venue_address` untuk offline/hybrid (string bebas)
- ✅ `online_platform` + `online_url` untuk online/hybrid
- ✅ Sertakan `export_calendar: true` untuk dapatkan calendar export links
- ✅ **`timezone` wajib diisi** (mandatory) saat create maupun edit event — IANA identifier seperti `Asia/Jakarta`, `Asia/Makassar`, atau `Asia/Jayapura`. Field ini menganchor `event_date`+`event_time` (wall-clock lokal) ke instant UTC absolut, dipakai untuk reminder scheduling & calendar export
- ✅ Feedback hanya bisa diisi setelah event berlangsung, dan hanya oleh user yang bukan organizer event tsb
- ✅ Satu feedback per user per event — pakai `PUT` untuk revisi, bukan `POST` berulang
- ✅ Hanya organizer event yang bisa lihat daftar & ringkasan feedback event miliknya

### ❌ DON'T:
- ❌ Jangan kirim base64 atau file langsung saat create/edit event
- ❌ Jangan kirim full CDN URL saat create/edit — gunakan `s3:` key
- ❌ Jangan include `region_id` dalam request body event
- ❌ Member standard tidak bisa create event
- ❌ Jangan edit event yang sedang `pending_approval` atau sudah `published`
- ❌ Tidak ada fitur favorite — hanya save/bookmark
- ❌ `is_featured` read-only di mobile — tidak ada cara set/toggle dari client (editorial, dikontrol Superadmin via backoffice)
- ❌ **Breaking change (2026-07-17)**: Jangan omit `timezone` saat create/edit — request tanpa `timezone` akan ditolak `422` (`DOMAIN_EVENT_TIMEZONE_REQUIRED`). Client lama yang belum kirim field ini wajib update dulu sebelum create/edit event akan berhasil lagi
- ❌ Organizer tidak bisa submit feedback untuk eventnya sendiri
- ❌ User selain organizer tidak bisa lihat daftar/ringkasan feedback event (bukan konten publik)
- ❌ Feedback tidak bisa disubmit sebelum event dimulai, atau setelah window feedback tertutup

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
    "images": ["Maximum 10 images allowed"],
    "timezone": ["Timezone is required and must be a valid IANA identifier (e.g. Asia/Jakarta)"]
  }
}
```

| Scenario | HTTP | Reason |
|----------|------|--------|
| Member standard create event | 403 | Pro plan required |
| Presign dengan file extension tidak valid | 422 | Validation failed |
| Image > 5MB di S3 (presign content-length) | 422 | File size exceeded |
| Confirm dengan key tidak ditemukan di S3 | 404 | Upload may have failed |
| Hapus media yang sedang dipakai event | 409 | Media in use |
| Event date di masa lalu | 422 | Validation failed |
| venue_name kosong untuk offline event | 422 | Validation failed |
| online_url kosong untuk online event | 422 | Validation failed |
| `timezone` tidak dikirim saat create/edit | 422 | Validation failed (`DOMAIN_EVENT_TIMEZONE_REQUIRED`) |
| `timezone` bukan IANA identifier valid (gagal `time.LoadLocation`) | 422 | Validation failed (`DOMAIN_EVENT_TIMEZONE_INVALID`) |
| Edit published event | 403 | Status not editable |
| Event not found | 404 | Invalid event_id |
| Unauthenticated save/schedule | 401 | Auth required |
| Organizer submit feedback event sendiri | 403 | Organizer cannot review own event |
| Submit feedback sebelum event mulai | 422 | Feedback not open yet |
| Submit feedback dua kali untuk event yang sama | 409 | Already submitted, use PUT |
| Edit/delete feedback setelah window tertutup | 403 | Feedback window closed |
| Comment feedback mengandung kata terlarang | 422 | Validation failed (banned keyword) |
| Non-organizer akses list/summary feedback | 403 | Only organizer can view |
| Get feedback/me tanpa pernah submit | 404 | No feedback found |

---

*API spec event module untuk mobile. Gunakan presigned URL flow (presign → upload → confirm) sebelum create/edit event.*
