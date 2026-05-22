# Notification Preferences — API Specification

Base URL: `/api/v1/mobile/notifications/preferences`
Auth: semua endpoint memerlukan `Authorization: Bearer <token>`

---

## Daftar Isi

- [Data Models](#data-models)
- [Endpoints](#endpoints)
  - [GET /preferences](#1-get-preferences)
  - [PUT /preferences/global](#2-put-preferencesglobal)
  - [PUT /preferences/modules/{module}](#3-put-preferencesmodulesmodule)
  - [PUT /preferences/communities/{community_id}](#4-put-preferencescommunitiescommunity_id)
  - [POST /preferences/reset](#5-post-preferencesreset)
- [Error Responses](#error-responses)
- [Sample Preference Configs](#sample-preference-configs)

---

## Data Models

### `NotificationPreference` (Full Response Object)

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "all_notifications_enabled": true,
  "do_not_disturb_enabled": false,
  "do_not_disturb_start_time": "22:00",
  "do_not_disturb_end_time": "08:00",
  "news": {
    "enabled": true,
    "from_kai_pusat": true,
    "from_kai_region": true,
    "from_pro_members": false
  },
  "community": {
    "enabled": true,
    "communities": {
      "comm_123": {
        "enabled": true,
        "new_posts": true,
        "member_joined": false,
        "member_left": false
      }
    }
  },
  "event": {
    "enabled": true,
    "reminders_enabled": true,
    "reminder_hours_before": 24,
    "interested_categories": ["sports", "culture"]
  },
  "qna": {
    "enabled": true,
    "someone_replied": true,
    "reply_from_moderator": true
  },
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

---

## Endpoints

### 1. GET /preferences

Mengambil seluruh notification preferences milik user yang sedang login.

**Request**
```
GET /api/v1/mobile/notifications/preferences
Authorization: Bearer <token>
```

**Response `200 OK`**
```json
{
  "data": { ...NotificationPreference }
}
```

> Jika user belum punya preferences, server akan auto-create dengan default values dan mengembalikan hasilnya.

---

### 2. PUT /preferences/global

Update pengaturan global: master toggle dan Do Not Disturb.

**Request**
```
PUT /api/v1/mobile/notifications/preferences/global
Authorization: Bearer <token>
Content-Type: application/json
```

**Body**
```json
{
  "all_notifications_enabled": true,
  "do_not_disturb_enabled": true,
  "do_not_disturb_start_time": "22:00",
  "do_not_disturb_end_time": "08:00"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `all_notifications_enabled` | `bool` | No | Master toggle semua notifikasi |
| `do_not_disturb_enabled` | `bool` | No | Enable/disable mode DND |
| `do_not_disturb_start_time` | `string` (HH:mm) | No | Waktu mulai DND. Wajib jika DND enabled |
| `do_not_disturb_end_time` | `string` (HH:mm) | No | Waktu selesai DND. Wajib jika DND enabled |

**Response `200 OK`**
```json
{
  "data": { ...NotificationPreference }
}
```

---

### 3. PUT /preferences/modules/{module}

Update pengaturan per-modul. Module yang tersedia: `news`, `community`, `event`, `qna`.

**Request**
```
PUT /api/v1/mobile/notifications/preferences/modules/{module}
Authorization: Bearer <token>
Content-Type: application/json
```

#### Module: `news`

```json
{
  "enabled": true,
  "from_kai_pusat": true,
  "from_kai_region": true,
  "from_pro_members": false
}
```

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Toggle seluruh notifikasi berita |
| `from_kai_pusat` | `bool` | Notifikasi dari KAI Pusat |
| `from_kai_region` | `bool` | Notifikasi dari KAI Regional |
| `from_pro_members` | `bool` | Notifikasi dari Pro Members |

---

#### Module: `community`

```json
{
  "enabled": true
}
```

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Toggle seluruh notifikasi komunitas |

> Pengaturan per-komunitas diatur lewat endpoint `/communities/{community_id}`.

---

#### Module: `event`

```json
{
  "enabled": true,
  "reminders_enabled": true,
  "reminder_hours_before": 24,
  "interested_categories": ["sports", "culture", "business"]
}
```

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Toggle seluruh notifikasi event |
| `reminders_enabled` | `bool` | Aktifkan pengingat event |
| `reminder_hours_before` | `int` | Jam sebelum event untuk mengirim pengingat. Default: `24` |
| `interested_categories` | `string[]` | Filter kategori event. Kosong = semua kategori |

---

#### Module: `qna`

```json
{
  "enabled": true,
  "someone_replied": true,
  "reply_from_moderator": true
}
```

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Toggle seluruh notifikasi Q&A |
| `someone_replied` | `bool` | Notifikasi ketika ada yang membalas |
| `reply_from_moderator` | `bool` | Notifikasi ketika moderator membalas |

**Response `200 OK`**
```json
{
  "data": { ...NotificationPreference }
}
```

---

### 4. PUT /preferences/communities/{community_id}

Update pengaturan notifikasi untuk satu komunitas spesifik.

**Request**
```
PUT /api/v1/mobile/notifications/preferences/communities/{community_id}
Authorization: Bearer <token>
Content-Type: application/json
```

**Path Params**

| Param | Type | Description |
|---|---|---|
| `community_id` | `uuid` | ID komunitas yang ingin diatur |

**Body**
```json
{
  "enabled": true,
  "new_posts": true,
  "member_joined": false,
  "member_left": false
}
```

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Toggle notifikasi untuk komunitas ini |
| `new_posts` | `bool` | Notifikasi ketika ada post baru |
| `member_joined` | `bool` | Notifikasi ketika ada member baru bergabung |
| `member_left` | `bool` | Notifikasi ketika ada member keluar |

**Response `200 OK`**
```json
{
  "data": { ...NotificationPreference }
}
```

---

### 5. POST /preferences/reset

Reset semua notification preferences ke default values.

**Request**
```
POST /api/v1/mobile/notifications/preferences/reset
Authorization: Bearer <token>
```

**Response `200 OK`**
```json
{
  "data": { ...NotificationPreference }
}
```

> Setelah reset, response mengembalikan preferences dengan default values yang baru.

---

## Error Responses

Semua error menggunakan format standar:

```json
{
  "error": "ERROR_CODE",
  "message": "human-readable error message"
}
```

| HTTP Status | Error Code | Keterangan |
|---|---|---|
| `400` | `BAD_REQUEST` | Request body tidak valid atau field tidak sesuai |
| `401` | `UNAUTHORIZED` | Token tidak ada atau expired |
| `404` | `NOT_FOUND` | Community ID tidak ditemukan |
| `422` | `VALIDATION_ERROR` | Field gagal validasi (format waktu salah, dll) |
| `500` | `INTERNAL_ERROR` | Server error |

---

## Sample Preference Configs

### Conservative User (Minimal Notifications)

```json
{
  "all_notifications_enabled": true,
  "news": {
    "enabled": true,
    "from_kai_pusat": true,
    "from_kai_region": false,
    "from_pro_members": false
  },
  "community": { "enabled": false },
  "event": { "enabled": false },
  "qna": { "enabled": false }
}
```

### Active User (All Notifications)

```json
{
  "all_notifications_enabled": true,
  "news": {
    "enabled": true,
    "from_kai_pusat": true,
    "from_kai_region": true,
    "from_pro_members": true
  },
  "community": {
    "enabled": true,
    "communities": {
      "comm_sports": { "enabled": true, "new_posts": true, "member_joined": true },
      "comm_nature": { "enabled": true, "new_posts": true, "member_joined": false }
    }
  },
  "event": {
    "enabled": true,
    "reminders_enabled": true,
    "reminder_hours_before": 24,
    "interested_categories": ["sports", "culture"]
  },
  "qna": {
    "enabled": true,
    "someone_replied": true,
    "reply_from_moderator": true
  }
}
```

### Quiet Hours User

```json
{
  "all_notifications_enabled": true,
  "do_not_disturb_enabled": true,
  "do_not_disturb_start_time": "22:00",
  "do_not_disturb_end_time": "08:00"
}
```
