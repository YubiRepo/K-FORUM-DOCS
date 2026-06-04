# API Spec — Notification Inbox (Web / Backoffice)

Dokumentasi API untuk web client mengambil, membaca, dan menerima notifikasi in-app secara real-time.

> **Catatan**: Endpoint ini untuk **inbox notifikasi user yang sedang login** (bukan admin tester).
> Endpoint admin untuk broadcast/send/history ada di file terpisah: `API_SPEC_NOTIFICATION_TESTER_BACKOFFICE.md`.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <access_token>` (Required semua endpoint)

---

## Daftar Endpoint

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/notifications/inbox` | List notifikasi belum dibaca |
| GET | `/notifications/inbox/unread-count` | Jumlah notifikasi belum dibaca |
| PATCH | `/notifications/inbox/:id/read` | Tandai satu notifikasi sebagai dibaca |
| PATCH | `/notifications/inbox/read-all` | Tandai semua notifikasi sebagai dibaca |
| WS | `/ws/notifications` | WebSocket untuk notifikasi real-time |

---

## 1. GET /notifications/inbox

Ambil daftar notifikasi yang belum dibaca milik user yang sedang login.

**Method**: GET

**URL**: `/api/v1/web/notifications/inbox`

**Authentication**: Required

**Response (200 OK)**:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "notifikasi berhasil diambil",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "type": "security",
      "title": "Login Berhasil",
      "body": "Anda berhasil masuk ke akun K-Forum. Jika bukan Anda, segera amankan akun Anda.",
      "is_read": false,
      "reference_id": null,
      "reference_type": null,
      "created_at": "2026-06-04T10:30:00+07:00",
      "read_at": null
    },
    {
      "id": "770e8400-e29b-41d4-a716-446655440002",
      "type": "system",
      "title": "Pengumuman Baru",
      "body": "Maintenance server dijadwalkan Selasa 10 malam",
      "is_read": false,
      "reference_id": "announcement-uuid-789",
      "reference_type": "announcement",
      "created_at": "2026-06-04T08:00:00+07:00",
      "read_at": null
    }
  ],
  "meta": null
}
```

---

## 2. GET /notifications/inbox/unread-count

Ambil jumlah notifikasi yang belum dibaca.

**Method**: GET

**URL**: `/api/v1/web/notifications/inbox/unread-count`

**Authentication**: Required

**Response (200 OK)**:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "unread count berhasil diambil",
  "data": {
    "count": 3
  },
  "meta": null
}
```

---

## 3. PATCH /notifications/inbox/:id/read

Tandai satu notifikasi sebagai sudah dibaca.

**Method**: PATCH

**URL**: `/api/v1/web/notifications/inbox/:id/read`

**Authentication**: Required

**Path Parameters**:
| Parameter | Type | Deskripsi |
|-----------|------|-----------|
| `id` | string (UUID) | ID notifikasi |

**Response (200 OK)**:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "notifikasi ditandai sudah dibaca",
  "data": null,
  "meta": null
}
```

**Response (403 Forbidden)**:
```json
{
  "status": "error",
  "status_code": 403,
  "errors": "notifikasi bukan milik user ini"
}
```

---

## 4. PATCH /notifications/inbox/read-all

Tandai semua notifikasi belum dibaca menjadi sudah dibaca.

**Method**: PATCH

**URL**: `/api/v1/web/notifications/inbox/read-all`

**Authentication**: Required

**Response (200 OK)**:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "semua notifikasi ditandai sudah dibaca",
  "data": {
    "marked": 3
  },
  "meta": null
}
```

---

## 5. WebSocket /ws/notifications

Koneksi real-time untuk menerima notifikasi in-app.

**URL**: `/api/v1/web/ws/notifications`

**Authentication**: `Authorization: Bearer <access_token>` header

**Format Pesan (diterima client)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "security",
  "title": "Login Berhasil",
  "body": "Anda berhasil masuk ke akun K-Forum.",
  "reference_id": null,
  "reference_type": null,
  "payload": {
    "ip_address": "192.168.1.1",
    "event": "login"
  },
  "created_at": "2026-06-04T10:30:00Z"
}
```

**Contoh koneksi (JavaScript — browser)**:
```javascript
const token = localStorage.getItem('access_token')
const ws = new WebSocket(`wss://api.k-forum.com/api/v1/web/ws/notifications`)

// WebSocket browser tidak support custom header.
// Gunakan query param untuk token jika diperlukan, atau gunakan endpoint berbeda.
// Disarankan: kirim token via first message setelah connect.

ws.onopen = () => {
  console.log('WS connected')
}

ws.onmessage = (event) => {
  const notif = JSON.parse(event.data)
  // Tampilkan toast atau update badge
  showNotificationToast(notif)
  updateBadgeCount(prev => prev + 1)
}

ws.onclose = () => {
  console.log('WS disconnected, fallback to polling')
  startPolling() // poll /unread-count setiap 30 detik
}
```

> **Catatan Browser**: Browser tidak dapat mengirim custom HTTP header saat upgrade WebSocket. Untuk autentikasi dari browser, disarankan menggunakan cookie `HttpOnly` atau mengatur backend agar mendukung session-based auth untuk endpoint WS. Diskusikan dengan backend team untuk solusi yang sesuai.

---

## Perbedaan Web vs Mobile

| Aspek | Mobile (`/mobile/notifications`) | Web (`/web/notifications/inbox`) |
|-------|----------------------------------|----------------------------------|
| Prefix path | `/notifications` | `/notifications/inbox` |
| WS path | `/mobile/ws/notifications` | `/web/ws/notifications` |
| Auth WS | Header (wscat/native app support) | Perlu pertimbangan (browser WS header terbatas) |
| Format response | ✓ Identik | ✓ Identik |
| Source of truth | DB `notifications` table | DB `notifications` table |

---

## Catatan Perubahan Schema

### Field `is_read` (boolean) menggantikan `status` (string)

**Sebelum**:
```json
{ "status": "sent" }   // atau "pending", "queued", "failed"
```

**Sesudah**:
```json
{ "is_read": false }   // selalu false di endpoint list (sudah difilter unread)
```

**Alasan**: Status delivery per channel (berhasil/gagal/retry) ada di `DeliveryAttempt`, bukan di `Notification`. Notification hanya perlu tahu apakah user sudah membaca atau belum. `is_read: bool` lebih semantik dan tidak membingungkan client dengan state delivery internal.
