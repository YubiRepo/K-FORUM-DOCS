# Notification Preferences — API Specification

Base URL: `/api/v1/mobile/notifications/preferences`
Auth: semua endpoint memerlukan `Authorization: Bearer <token>`

---

## Daftar Isi

- [Data Models](#data-models)
- [Bypass Rules](#bypass-rules)
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
  "announcement": {
    "info_enabled": true
  },
  "news": {
    "enabled": true
  },
  "community": {
    "enabled": true,
    "communities": {
      "comm_123": {
        "enabled": true,
        "new_posts": true,
        "member_joined": false,
        "member_left": false,
        "join_request_approved": true
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
    "question_answered": true
  },
  "subscription": {
    "expiry_reminder_enabled": true
  },
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z"
}
```

---

## Bypass Rules

Beberapa jenis notifikasi **selalu dikirim** dan tidak bisa dimatikan oleh user — terlepas dari pengaturan preferences maupun mode Do Not Disturb.

| Modul | Kondisi Bypass | Keterangan |
|---|---|---|
| `announcement` | Tipe `disaster`, `system`, `urgent` dengan priority `CRITICAL` atau `HIGH` | Pengumuman darurat dan sistem kritikal tidak bisa diblokir user |
| `announcement` | Tipe `info` dengan priority `MEDIUM` atau `LOW` | Bisa dikontrol user via `announcement.info_enabled` |
| `subscription` | Status changes: upgrade approved, upgrade rejected, plan expired, plan downgraded | Perubahan status akun tidak bisa diblokir user |
| `subscription` | Expiry reminder (7 hari & 3 hari sebelum expired) | Bisa dikontrol user via `subscription.expiry_reminder_enabled` |

> Aturan ini align dengan behavior di modul Announcement: `priority: CRITICAL` selalu force-push ke semua device, sementara `priority: LOW` hanya muncul in-app tanpa push notification.

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

> **Catatan:** Master toggle (`all_notifications_enabled = false`) dan DND **tidak mempengaruhi** notifikasi announcement yang bersifat bypass (disaster, system, urgent CRITICAL/HIGH).

> **Catatan timezone (2026-07-17):** `do_not_disturb_start_time`/`do_not_disturb_end_time` adalah jam lokal wall-clock — **tidak ada field timezone terpisah untuk DND**, backend menentukan sendiri "jam berapa sekarang" versi user berdasarkan timezone device user yang paling terakhir aktif (field `timezone` yang sudah dikirim device saat register FCM/refresh token — lihat kolom `device_registrations.timezone`, ditambahkan 2026-07-10). Kalau device belum pernah kirim timezone (app lama), fallback ke `default_timezone` System Settings, lalu ke `Asia/Jakarta` kalau itu juga tidak ada. Jadi user cukup isi jam DND sesuai jam lokal mereka saat ini — tidak perlu (dan tidak bisa) memilih timezone eksplisit untuk field ini, beda dengan Event yang wajib pilih timezone eksplisit karena Event terikat lokasi, bukan device.
>
> **Drift dokumentasi ditemukan (bukan disebabkan perubahan ini, tidak diperbaiki di sini):** `API_SPEC_FCM.md` di direktori yang sama mendokumentasikan tabel `fcm_tokens` dengan field `fcm_token`/`device_id`/`platform` saja — tabel sebenarnya bernama `device_registrations` dan sudah punya kolom `timezone` (serta field lain) sejak 2026-07-10, tidak pernah didokumentasikan ulang. Spec itu perlu direvisi terpisah, di luar scope perubahan DND ini.

**Response `200 OK`**
```json
{
  "data": { ...NotificationPreference }
}
```

---

### 3. PUT /preferences/modules/{module}

Update pengaturan per-modul. Module yang tersedia: `announcement`, `news`, `community`, `event`, `qna`, `subscription`.

**Request**
```
PUT /api/v1/mobile/notifications/preferences/modules/{module}
Authorization: Bearer <token>
Content-Type: application/json
```

---

#### Module: `announcement`

Hanya notifikasi tipe `info` yang bisa dikontrol user. Tipe `disaster`, `system`, dan `urgent` dengan priority CRITICAL/HIGH selalu dikirim — lihat [Bypass Rules](#bypass-rules).

```json
{
  "info_enabled": true
}
```

| Field | Type | Description |
|---|---|---|
| `info_enabled` | `bool` | Toggle notifikasi announcement tipe `info` (priority MEDIUM/LOW) |

---

#### Module: `news`

```json
{
  "enabled": true
}
```

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Toggle seluruh notifikasi berita (dari KAI Pusat, KAI Regional, maupun Pro Members) |

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

> Pengaturan per-komunitas (termasuk `join_request_approved`) diatur lewat endpoint `/communities/{community_id}`.

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
  "question_answered": true
}
```

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Toggle seluruh notifikasi Q&A |
| `question_answered` | `bool` | Notifikasi ketika pertanyaan dijawab atau ditolak oleh superadmin |

> Modul Q&A hanya menghasilkan dua jenis notifikasi untuk member: pertanyaan dijawab dan pertanyaan ditolak. Keduanya dikontrol oleh satu toggle `question_answered` karena sama-sama merupakan update status dari pertanyaan yang diajukan.

---

#### Module: `subscription`

Hanya expiry reminder yang bisa dikontrol user. Notifikasi status changes (approved, rejected, expired, downgraded) selalu dikirim — lihat [Bypass Rules](#bypass-rules).

```json
{
  "expiry_reminder_enabled": true
}
```

| Field | Type | Description |
|---|---|---|
| `expiry_reminder_enabled` | `bool` | Toggle pengingat expiry plan (7 hari & 3 hari sebelum expired) |

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
  "member_left": false,
  "join_request_approved": true
}
```

| Field | Type | Description |
|---|---|---|
| `enabled` | `bool` | Toggle notifikasi untuk komunitas ini |
| `new_posts` | `bool` | Notifikasi ketika ada post baru |
| `member_joined` | `bool` | Notifikasi ketika ada member baru bergabung |
| `member_left` | `bool` | Notifikasi ketika ada member keluar |
| `join_request_approved` | `bool` | Notifikasi ketika join request user di-approve (hanya relevan untuk komunitas private) |

> `join_request_approved` tetap bisa di-set untuk komunitas public, namun tidak akan pernah terpicu karena komunitas public menggunakan auto-join tanpa approval.

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

## Default Values

Nilai default saat preferences pertama kali di-create untuk user baru:

| Field | Default |
|---|---|
| `all_notifications_enabled` | `true` |
| `do_not_disturb_enabled` | `false` |
| `announcement.info_enabled` | `true` |
| `news.enabled` | `true` |
| `community.enabled` | `true` |
| `event.enabled` | `true` |
| `event.reminders_enabled` | `true` |
| `event.reminder_hours_before` | `24` |
| `event.interested_categories` | `[]` (semua kategori) |
| `qna.enabled` | `true` |
| `qna.question_answered` | `true` |
| `subscription.expiry_reminder_enabled` | `true` |
| Per komunitas: `new_posts` | `true` |
| Per komunitas: `member_joined` | `false` |
| Per komunitas: `member_left` | `false` |
| Per komunitas: `join_request_approved` | `true` |

---

## Sample Preference Configs

### Conservative User (Minimal Notifications)

```json
{
  "all_notifications_enabled": true,
  "announcement": {
    "info_enabled": false
  },
  "news": {
    "enabled": true
  },
  "community": { "enabled": false },
  "event": { "enabled": false },
  "qna": { "enabled": false },
  "subscription": { "expiry_reminder_enabled": false }
}
```

### Active User (All Notifications)

```json
{
  "all_notifications_enabled": true,
  "announcement": {
    "info_enabled": true
  },
  "news": {
    "enabled": true
  },
  "community": {
    "enabled": true,
    "communities": {
      "comm_sports": {
        "enabled": true,
        "new_posts": true,
        "member_joined": true,
        "member_left": false,
        "join_request_approved": true
      },
      "comm_nature": {
        "enabled": true,
        "new_posts": true,
        "member_joined": false,
        "member_left": false,
        "join_request_approved": true
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
    "question_answered": true
  },
  "subscription": {
    "expiry_reminder_enabled": true
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

### Emergency Only (Mute Semua Kecuali Darurat)

```json
{
  "all_notifications_enabled": false,
  "announcement": {
    "info_enabled": false
  },
  "news": { "enabled": false },
  "community": { "enabled": false },
  "event": { "enabled": false },
  "qna": { "enabled": false },
  "subscription": { "expiry_reminder_enabled": false }
}
```

> Meskipun semua dimatikan, notifikasi announcement tipe `disaster`, `system`, dan `urgent` dengan priority CRITICAL/HIGH tetap akan diterima.
