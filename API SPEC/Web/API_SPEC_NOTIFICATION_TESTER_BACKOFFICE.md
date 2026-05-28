# API Spec — Notification Tester (Web Backoffice)

Dokumentasi API untuk superadmin mengirim push notification langsung dari backoffice — untuk keperluan testing, broadcast pengumuman, atau kirim pesan ke user tertentu tanpa harus lewat kode.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/notifications`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin only)
- **Authorization**: Semua endpoint hanya untuk Superadmin
- **Error Format**: Same as other backoffice APIs (standard message or validation error)

---

## Konsep

```
┌────────────────────────────────────────────────────────────┐
│ NOTIFICATION TESTER — DUA MODE                             │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Mode 1: Per User                                          │
│  Kirim ke satu atau beberapa user spesifik                 │
│  POST /api/v1/web/notifications/send                       │
│  { "target": "user", "user_ids": ["uuid1", "uuid2"] }      │
│                                                            │
│  Mode 2: Broadcast                                         │
│  Kirim ke semua user aktif di platform                     │
│  POST /api/v1/web/notifications/broadcast                  │
│  { "target": "all" }                                       │
│                                                            │
│  Keduanya respect notification preferences user            │
│  kecuali jika di-override dengan bypass_preferences: true  │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

> **Catatan penting:** Notifikasi yang dikirim lewat tester ini tetap melewati FCM token management yang sudah ada. Backend akan query active FCM tokens milik target user sebelum mengirim.

---

## Model Data Utama

### 1. Notification Payload

```json
{
  "title": "string",
  "body": "string",
  "image_url": "string (nullable, URL gambar thumbnail)",
  "type": "string (enum: custom, event, news, community, announcement)",
  "entity_id": "string (nullable, ID entitas terkait)",
  "click_action": "string (nullable, deep link: kai://... )",
  "data": {
    "key": "value"
  }
}
```

### 2. Send Result Per User

```json
{
  "user_id": "uuid",
  "user_name": "Andi Pratama",
  "email": "andi@example.com",
  "status": "success" | "skipped" | "failed",
  "reason": "string (nullable, diisi jika skipped atau failed)",
  "devices_sent": 2,
  "devices_total": 3
}
```

### 3. Notification Log Object

```json
{
  "id": "uuid",
  "title": "string",
  "body": "string",
  "type": "string",
  "target_mode": "user" | "broadcast",
  "total_targets": 100,
  "total_sent": 97,
  "total_skipped": 2,
  "total_failed": 1,
  "bypass_preferences": false,
  "sent_by": "uuid",
  "sent_by_name": "Super Admin",
  "sent_at": "2026-05-26T10:00:00.000Z"
}
```

---

## Endpoints

### 1. Send Notification ke User Tertentu

Kirim push notification ke satu atau beberapa user spesifik berdasarkan `user_id`.

- **URL**: `POST /api/v1/web/notifications/send`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "user_ids": ["uuid1", "uuid2", "uuid3"],
    "notification": {
      "title": "Halo dari Admin!",
      "body": "Ini adalah test notifikasi dari backoffice.",
      "image_url": null,
      "type": "custom",
      "entity_id": null,
      "click_action": null
    },
    "bypass_preferences": false
  }
  ```

- **Request Validation**:
  - `user_ids`: Required, array, min 1 item, max 500 items
  - `notification.title`: Required, string, max 100 chars
  - `notification.body`: Required, string, max 300 chars
  - `notification.image_url`: Optional, valid URL
  - `notification.type`: Required, enum: `custom`, `event`, `news`, `community`, `announcement`
  - `notification.entity_id`: Optional, string
  - `notification.click_action`: Optional, string (deep link)
  - `bypass_preferences`: Optional, boolean, default `false`. Jika `true`, notifikasi dikirim meski user matikan notifikasi

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "operation_id": "notif_op_20260526_001",
      "total_targets": 3,
      "total_sent": 3,
      "total_skipped": 0,
      "total_failed": 0,
      "results": [
        {
          "user_id": "uuid1",
          "user_name": "Andi Pratama",
          "email": "andi@example.com",
          "status": "success",
          "devices_sent": 2,
          "devices_total": 2
        },
        {
          "user_id": "uuid2",
          "user_name": "Budi Santoso",
          "email": "budi@example.com",
          "status": "success",
          "devices_sent": 1,
          "devices_total": 1
        },
        {
          "user_id": "uuid3",
          "user_name": "Citra Dewi",
          "email": "citra@example.com",
          "status": "success",
          "devices_sent": 1,
          "devices_total": 1
        }
      ]
    },
    "message": "Notification sent to 3 users successfully"
  }
  ```

- **Response (Partial — Some Skipped or Failed)**:
  ```json
  {
    "data": {
      "operation_id": "notif_op_20260526_002",
      "total_targets": 3,
      "total_sent": 1,
      "total_skipped": 1,
      "total_failed": 1,
      "results": [
        {
          "user_id": "uuid1",
          "user_name": "Andi Pratama",
          "email": "andi@example.com",
          "status": "success",
          "devices_sent": 2,
          "devices_total": 2
        },
        {
          "user_id": "uuid2",
          "user_name": "Budi Santoso",
          "email": "budi@example.com",
          "status": "skipped",
          "reason": "User has disabled all notifications",
          "devices_sent": 0,
          "devices_total": 1
        },
        {
          "user_id": "uuid3",
          "user_name": "Citra Dewi",
          "email": "citra@example.com",
          "status": "failed",
          "reason": "No active FCM tokens found",
          "devices_sent": 0,
          "devices_total": 0
        }
      ]
    },
    "message": "1 of 3 users received notification (1 skipped, 1 failed)"
  }
  ```

- **Response (Error — 422)**:
  ```json
  {
    "message": "Validation failed",
    "errors": {
      "user_ids": ["Must provide at least 1 user"],
      "notification.title": ["Title is required"],
      "notification.type": ["Invalid notification type"]
    }
  }
  ```

---

### 2. Broadcast Notification ke Semua User

Kirim push notification ke semua user aktif di platform. Cocok untuk pengumuman platform-wide.

- **URL**: `POST /api/v1/web/notifications/broadcast`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "notification": {
      "title": "Pembaruan Platform KAI",
      "body": "Kami baru saja merilis fitur baru! Update aplikasi kamu sekarang.",
      "image_url": "https://cdn.example.com/announcements/update_banner.jpg",
      "type": "announcement",
      "click_action": "kai://home"
    },
    "filters": {
      "subscription_plan": "pro",
      "platform": "android"
    },
    "bypass_preferences": false,
    "confirm": true
  }
  ```

- **Request Validation**:
  - `notification.title`: Required, string, max 100 chars
  - `notification.body`: Required, string, max 300 chars
  - `notification.type`: Required, enum: `custom`, `event`, `news`, `community`, `announcement`
  - `filters`: Optional object untuk mempersempit target
    - `filters.subscription_plan`: Optional, enum: `standard`, `pro`
    - `filters.platform`: Optional, enum: `android`, `ios`, `web`
  - `bypass_preferences`: Optional, boolean, default `false`
  - `confirm`: **Required**, harus bernilai `true` sebagai konfirmasi broadcast

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "operation_id": "broadcast_op_20260526_001",
      "total_targets": 1250,
      "total_sent": 1198,
      "total_skipped": 42,
      "total_failed": 10,
      "filters_applied": {
        "subscription_plan": "pro",
        "platform": "android"
      },
      "bypass_preferences": false,
      "sent_by": "superadmin_uuid",
      "sent_by_name": "Super Admin",
      "started_at": "2026-05-26T10:00:00.000Z",
      "completed_at": "2026-05-26T10:00:08.000Z"
    },
    "message": "Broadcast sent to 1198 of 1250 users (42 skipped, 10 failed)"
  }
  ```

- **Response (Error — 400, Missing Confirm)**:
  ```json
  {
    "message": "Broadcast requires confirm: true to prevent accidental sends"
  }
  ```

---

### 3. Preview Broadcast (Dry Run)

Simulasikan broadcast tanpa benar-benar mengirim — hanya hitung berapa user yang akan terkena dampak berdasarkan filter yang dipilih.

- **URL**: `POST /api/v1/web/notifications/broadcast/preview`
- **Autentikasi**: Yes (Superadmin)

- **Request Body**:
  ```json
  {
    "filters": {
      "subscription_plan": "pro",
      "platform": "android"
    },
    "bypass_preferences": false
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "estimated_targets": 1250,
      "estimated_sent": 1198,
      "estimated_skipped": 42,
      "filters_applied": {
        "subscription_plan": "pro",
        "platform": "android"
      },
      "breakdown": {
        "with_active_tokens": 1230,
        "without_active_tokens": 20,
        "notifications_disabled": 42
      }
    },
    "message": "Preview only — no notifications were sent"
  }
  ```

---

### 4. Get Notification History

Ambil riwayat semua notifikasi yang pernah dikirim dari backoffice.

- **URL**: `GET /api/v1/web/notifications/history`
- **Autentikasi**: Yes (Superadmin)
- **Query Parameters**:
  - `type` (optional): Filter by type (`custom`, `announcement`, dll)
  - `target_mode` (optional): `user` atau `broadcast`
  - `date_from` (optional): Filter from date
  - `date_to` (optional): Filter to date
  - `sent_by` (optional): Filter by superadmin user_id
  - `limit` (optional, default: 20)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "notif_log_001",
        "title": "Pembaruan Platform KAI",
        "body": "Kami baru saja merilis fitur baru!...",
        "type": "announcement",
        "target_mode": "broadcast",
        "total_targets": 1250,
        "total_sent": 1198,
        "total_skipped": 42,
        "total_failed": 10,
        "bypass_preferences": false,
        "sent_by": "superadmin_uuid",
        "sent_by_name": "Super Admin",
        "sent_at": "2026-05-26T10:00:00.000Z"
      },
      {
        "id": "notif_log_002",
        "title": "Halo dari Admin!",
        "body": "Ini adalah test notifikasi dari backoffice.",
        "type": "custom",
        "target_mode": "user",
        "total_targets": 3,
        "total_sent": 3,
        "total_skipped": 0,
        "total_failed": 0,
        "bypass_preferences": false,
        "sent_by": "superadmin_uuid",
        "sent_by_name": "Super Admin",
        "sent_at": "2026-05-26T09:30:00.000Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 47
    }
  }
  ```

---

### 5. Get Notification Detail

Ambil detail lengkap satu pengiriman notifikasi, termasuk hasil per user.

- **URL**: `GET /api/v1/web/notifications/history/{operation_id}`
- **Autentikasi**: Yes (Superadmin)

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "notif_log_002",
      "title": "Halo dari Admin!",
      "body": "Ini adalah test notifikasi dari backoffice.",
      "image_url": null,
      "type": "custom",
      "entity_id": null,
      "click_action": null,
      "target_mode": "user",
      "bypass_preferences": false,
      "total_targets": 3,
      "total_sent": 3,
      "total_skipped": 0,
      "total_failed": 0,
      "sent_by": "superadmin_uuid",
      "sent_by_name": "Super Admin",
      "sent_at": "2026-05-26T09:30:00.000Z",
      "results": [
        {
          "user_id": "uuid1",
          "user_name": "Andi Pratama",
          "email": "andi@example.com",
          "status": "success",
          "devices_sent": 2,
          "devices_total": 2
        },
        {
          "user_id": "uuid2",
          "user_name": "Budi Santoso",
          "email": "budi@example.com",
          "status": "success",
          "devices_sent": 1,
          "devices_total": 1
        },
        {
          "user_id": "uuid3",
          "user_name": "Citra Dewi",
          "email": "citra@example.com",
          "status": "success",
          "devices_sent": 1,
          "devices_total": 1
        }
      ]
    }
  }
  ```

---

### 6. Get User Devices & Token Status

Sebelum kirim notifikasi ke user tertentu, superadmin bisa cek dulu device apa saja yang dimiliki user dan apakah token-nya aktif. Berguna untuk debugging kenapa notifikasi tidak sampai.

- **URL**: `GET /api/v1/web/notifications/users/{user_id}/devices`
- **Autentikasi**: Yes (Superadmin)

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "user_id": "uuid1",
      "user_name": "Andi Pratama",
      "email": "andi@example.com",
      "notifications_enabled": true,
      "devices": [
        {
          "id": "fcm_tok_001",
          "device_name": "Samsung Galaxy S21",
          "platform": "android",
          "device_model": "SM-G991B",
          "os_version": "Android 13",
          "app_version": "1.2.3",
          "is_active": true,
          "last_used_at": "2026-05-26T08:45:00.000Z",
          "created_at": "2026-04-01T10:30:00.000Z"
        },
        {
          "id": "fcm_tok_002",
          "device_name": "iPad Air",
          "platform": "ios",
          "device_model": "iPad13,1",
          "os_version": "iOS 17.2",
          "app_version": "1.2.3",
          "is_active": false,
          "last_used_at": "2026-04-10T16:30:00.000Z",
          "created_at": "2026-02-20T09:00:00.000Z"
        }
      ],
      "active_devices_count": 1,
      "total_devices_count": 2
    }
  }
  ```

---

## UI Flow Example

### Page: Notification Tester

```
┌─────────────────────────────────────────────────────────────┐
│ Notification Tester                                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ TARGET:                                                     │
│ ( ) Kirim ke User Tertentu   (●) Broadcast ke Semua User   │
│                                                             │
│ ── jika Broadcast ──────────────────────────────────────── │
│ Filter (opsional):                                          │
│ Subscription: [Semua ▼]    Platform: [Semua ▼]             │
│                                                             │
│ [Preview — estimasi 1.250 user akan menerima]               │
│                                                             │
│ MESSAGE:                                                    │
│ Title:   [Pembaruan Platform KAI           ] (max 100)     │
│ Body:    [Kami baru saja merilis fitur baru] (max 300)     │
│ Image:   [https://cdn.../banner.jpg        ] (opsional)    │
│                                                             │
│ Tipe:    [Announcement ▼]                                  │
│ Deep Link:[kai://home                      ] (opsional)    │
│                                                             │
│ [✓] Bypass notification preferences user                   │
│                                                             │
│          [Preview] [Kirim ke 1.250 User →]                 │
└─────────────────────────────────────────────────────────────┘
```

### Flow: Broadcast dengan Preview dulu

```
1. Superadmin pilih mode "Broadcast"
2. Set filter: Subscription = Pro, Platform = Semua

3. Klik [Preview]
   POST /api/v1/web/notifications/broadcast/preview
   { "filters": { "subscription_plan": "pro" } }

4. Response preview:
   ┌───────────────────────────────────────────┐
   │ Preview Broadcast                         │
   ├───────────────────────────────────────────┤
   │ Estimasi target:   1.250 user             │
   │ Estimasi terkirim: 1.198 user             │
   │ Estimasi dilewati: 42 user                │
   │ (user matikan notifikasi)                 │
   │                                           │
   │ Token aktif:  1.230 device                │
   │ Token tidak aktif: 20 device              │
   │                          [Lanjut Kirim]   │
   └───────────────────────────────────────────┘

5. Klik [Lanjut Kirim]
   POST /api/v1/web/notifications/broadcast
   { "confirm": true, "filters": {...}, "notification": {...} }

6. Response: Success
   → Banner: "Broadcast dikirim ke 1.198 dari 1.250 user"
   → Log masuk ke Notification History
```

### Flow: Kirim ke User Tertentu (Testing)

```
1. Superadmin pilih mode "Kirim ke User Tertentu"

2. Cari dan pilih user:
   ┌──────────────────────────────────────────────────────┐
   │ Cari user: [andi@...       ]                         │
   │ ☑ Andi Pratama (andi@example.com) — 2 device aktif  │
   │ ☑ Budi Santoso (budi@example.com) — 1 device aktif  │
   │ ☐ Citra Dewi   (citra@example.com) — 0 device aktif │
   └──────────────────────────────────────────────────────┘

3. Isi pesan, klik [Kirim ke 2 User]
   POST /api/v1/web/notifications/send
   {
     "user_ids": ["uuid_andi", "uuid_budi"],
     "notification": { "title": "Test", "body": "...", "type": "custom" },
     "bypass_preferences": false
   }

4. Response:
   ┌────────────────────────────────────────────┐
   │ Notifikasi Terkirim                        │
   ├────────────────────────────────────────────┤
   │ ✓ Andi Pratama  — 2 device  (Success)      │
   │ ✓ Budi Santoso  — 1 device  (Success)      │
   └────────────────────────────────────────────┘
```

---

## Important Notes

### ✅ DO:
- ✅ Gunakan Preview sebelum broadcast untuk estimasi dampak
- ✅ Broadcast selalu butuh `confirm: true` untuk cegah kirim tidak sengaja
- ✅ Log semua pengiriman — bisa dilihat di Notification History
- ✅ Cek device user dulu jika notifikasi tidak sampai (endpoint devices)
- ✅ Gunakan `bypass_preferences: true` hanya untuk notifikasi kritis / urgent

### ❌ DON'T:
- ❌ Jangan broadcast tanpa Preview dulu
- ❌ Jangan set `bypass_preferences: true` untuk notifikasi rutin — hormati preferensi user
- ❌ Broadcast bukan untuk notifikasi personal — gunakan Send ke user tertentu

### Status Pengiriman Per User:

| Status | Artinya |
|--------|---------|
| `success` | Notifikasi berhasil dikirim ke minimal 1 device |
| `skipped` | User menonaktifkan notifikasi (dan bypass_preferences = false) |
| `failed` | Tidak ada active FCM token, atau semua token invalid |

### Notification Types:

| Type | Kapan Dipakai |
|------|--------------|
| `custom` | Testing bebas, tidak terkait entitas apapun |
| `announcement` | Pengumuman resmi platform |
| `event` | Notifikasi terkait event (isi `entity_id` dengan event_id) |
| `news` | Notifikasi terkait berita (isi `entity_id` dengan news_id) |
| `community` | Notifikasi terkait komunitas (isi `entity_id` dengan community_id) |

---

## Error Handling

Standard error responses:

```json
// 400 Bad Request
{
  "message": "Broadcast requires confirm: true to prevent accidental sends"
}

// 401 Unauthorized
{
  "message": "Authentication required"
}

// 403 Forbidden
{
  "message": "Only superadmin can send notifications from backoffice"
}

// 404 Not Found
{
  "message": "User not found"
}

// 422 Unprocessable Entity
{
  "message": "Validation failed",
  "errors": {
    "notification.title": ["Title is required"],
    "notification.body": ["Body exceeds 300 character limit"],
    "user_ids": ["Must provide at least 1 user"]
  }
}
```

| Scenario | HTTP | Reason |
|----------|------|--------|
| Broadcast tanpa `confirm: true` | 400 | Safety check |
| Title/body kosong | 422 | Validation failed |
| user_ids kosong | 422 | Validation failed |
| user_ids > 500 | 422 | Batch limit |
| User tidak ditemukan | 404 | Invalid user_id |
| Non-superadmin access | 403 | Authorization failed |
| Semua token invalid (FCM) | 200 | Partial — status `failed` per user |

---

*API spec notification tester untuk web backoffice. Superadmin only. Semua pengiriman tercatat di history.*
