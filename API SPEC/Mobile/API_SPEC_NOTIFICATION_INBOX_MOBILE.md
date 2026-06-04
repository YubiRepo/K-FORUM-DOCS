# API Spec — Notification Inbox (Mobile Client)

Dokumentasi API untuk mobile mengambil, membaca, dan menerima notifikasi in-app secara real-time.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <access_token>` (Required semua endpoint)

---

## Daftar Endpoint

| Method | Path | Deskripsi |
|--------|------|-----------|
| GET | `/notifications` | List notifikasi belum dibaca |
| GET | `/notifications/unread-count` | Jumlah notifikasi belum dibaca |
| PATCH | `/notifications/:id/read` | Tandai satu notifikasi sebagai dibaca |
| PATCH | `/notifications/read-all` | Tandai semua notifikasi sebagai dibaca |
| WS | `/ws/notifications` | WebSocket untuk notifikasi real-time |

---

## 1. GET /notifications

Ambil daftar notifikasi yang belum dibaca milik user yang sedang login.

**Method**: GET

**URL**: `/api/v1/mobile/notifications`

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
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "type": "social",
      "title": "Post kamu disukai",
      "body": "budi_santoso menyukai post kamu",
      "is_read": false,
      "reference_id": "post-uuid-123",
      "reference_type": "post",
      "created_at": "2026-06-04T09:15:00+07:00",
      "read_at": null
    }
  ],
  "meta": null
}
```

**Catatan**:
- Hanya mengembalikan notifikasi dengan `status = 'unread'`
- Diurutkan dari yang terbaru (`created_at DESC`)
- Tidak ada pagination — ambil semua yang belum dibaca
- `is_read` selalu `false` di endpoint ini (sudah difilter unread); field tersedia untuk konsistensi saat WS push juga mengandung data yang sama

**Tipe notifikasi (`type`)**:
| Value | Deskripsi |
|-------|-----------|
| `security` | Aktivitas keamanan (login, perubahan password) |
| `social` | Interaksi sosial (like, komentar, mention) |
| `content` | Konten baru (post komunitas, pengumuman) |
| `system` | Pesan sistem (broadcast admin) |
| `reminder` | Pengingat (event, jadwal) |
| `marketing` | Promosi dan informasi produk |

**Response (401 Unauthorized)**:
```json
{
  "status": "error",
  "status_code": 401,
  "errors": "user tidak terautentikasi"
}
```

---

## 2. GET /notifications/unread-count

Ambil jumlah notifikasi yang belum dibaca.

**Method**: GET

**URL**: `/api/v1/mobile/notifications/unread-count`

**Authentication**: Required

**Response (200 OK)**:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "unread count berhasil diambil",
  "data": {
    "count": 5
  },
  "meta": null
}
```

**Catatan**:
- Endpoint ringan, cocok untuk di-poll secara berkala (misal setiap 30 detik) sebagai fallback jika WS terputus
- Return `0` jika tidak ada notifikasi belum dibaca

---

## 3. PATCH /notifications/:id/read

Tandai satu notifikasi sebagai sudah dibaca.

**Method**: PATCH

**URL**: `/api/v1/mobile/notifications/:id/read`

**Authentication**: Required

**Path Parameters**:
| Parameter | Type | Deskripsi |
|-----------|------|-----------|
| `id` | string (UUID) | ID notifikasi |

**Request Body**: Tidak diperlukan

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

**Response (403 Forbidden)** — notifikasi bukan milik user:
```json
{
  "status": "error",
  "status_code": 403,
  "errors": "notifikasi bukan milik user ini"
}
```

**Response (404 Not Found)**:
```json
{
  "status": "error",
  "status_code": 404,
  "errors": "notifikasi tidak ditemukan"
}
```

---

## 4. PATCH /notifications/read-all

Tandai semua notifikasi belum dibaca menjadi sudah dibaca.

**Method**: PATCH

**URL**: `/api/v1/mobile/notifications/read-all`

**Authentication**: Required

**Request Body**: Tidak diperlukan

**Response (200 OK)**:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "semua notifikasi ditandai sudah dibaca",
  "data": {
    "marked": 5
  },
  "meta": null
}
```

**Catatan**:
- `marked` = jumlah notifikasi yang berhasil ditandai dibaca
- Best-effort: jika satu gagal, proses tetap lanjut ke notifikasi berikutnya

---

## 5. WebSocket /ws/notifications

Koneksi real-time untuk menerima notifikasi in-app tanpa polling.

**Protocol**: WebSocket (`ws://` atau `wss://`)

**URL**: `/api/v1/mobile/ws/notifications`

**Authentication**: `Authorization: Bearer <access_token>` header

**Lifecycle**:
```
Client                          Server
  │                               │
  ├── Upgrade: websocket ────────►│ (JWTAuth middleware validate token)
  │◄── 101 Switching Protocols ───│
  │                               │ (Register client ke hub)
  │                               │
  │◄── {"id":"...","type":"..."} ─│ (Push notifikasi real-time)
  │◄── {"id":"...","type":"..."} ─│
  │                               │
  │── ping ──────────────────────►│ (keepalive setiap 50 detik)
  │◄── pong ──────────────────────│
  │                               │
  │── close ─────────────────────►│ (disconnect)
```

**Format Pesan (diterima client)**:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "social",
  "title": "Post kamu disukai",
  "body": "budi_santoso menyukai post kamu",
  "reference_id": "post-uuid-123",
  "reference_type": "post",
  "payload": {
    "community_id": "comm-uuid-456",
    "post_id": "post-uuid-123"
  },
  "created_at": "2026-06-04T10:30:00Z"
}
```

**Field payload per event type**:
| type | payload keys |
|------|-------------|
| `security` | `ip_address`, `event` |
| `social` (post liked) | `community_id`, `post_id` |
| `social` (commented) | `community_id`, `post_id`, `comment_id` |
| `social` (replied) | `community_id`, `post_id`, `comment_id` |
| `social` (mentioned) | `community_id`, `post_id`, `context` |
| `system` (announcement) | `announcement_id` |

**Catatan Penting**:
- Server **hanya mengirim**, client tidak perlu kirim data (kecuali ping keepalive)
- Jika koneksi WS terputus, notifikasi tetap tersimpan di DB — client cukup hit `GET /notifications` saat reconnect untuk catch-up
- Pesan broadcast (announcement) dikirim ke semua user yang terkoneksi, tapi **tidak** tersimpan di tabel `notifications` — user lihat via `/announcements`

**Contoh koneksi (wscat)**:
```bash
wscat -c "ws://localhost:8888/api/v1/mobile/ws/notifications" \
      -H "Authorization: Bearer <access_token>"
```

---

## Pola Penggunaan yang Direkomendasikan (Client)

```
App Launch / User Login
  │
  ├─ 1. Hit GET /notifications           ← Load semua unread dari DB
  ├─ 2. Hit GET /notifications/unread-count ← Update badge count
  ├─ 3. Connect WS /ws/notifications     ← Listen real-time
  │
  │  [Saat WS menerima pesan]
  ├─ 4. Append ke list lokal             ← Tidak perlu re-fetch
  ├─ 5. Update badge count (count + 1)
  │
  │  [User tap notifikasi]
  ├─ 6. PATCH /notifications/:id/read    ← Mark dibaca
  ├─ 7. Remove dari list lokal / set is_read: true
  │
  │  [User tap "baca semua"]
  └─ 8. PATCH /notifications/read-all   ← Mark semua dibaca
```
