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
| GET | `/notifications` | List 20 notifikasi terbaru |
| GET | `/notifications/unread-count` | Jumlah notifikasi belum dibaca |
| PATCH | `/notifications/:id/read` | Tandai satu notifikasi sebagai dibaca |
| PATCH | `/notifications/read-all` | Tandai semua notifikasi sebagai dibaca |
| WS | `/ws/notifications` | WebSocket untuk notifikasi real-time |

---

## 1. GET /notifications

Ambil 20 notifikasi terbaru milik user yang sedang login, diurutkan dari yang terbaru.

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
      "type": "content",
      "title": "Post Baru di Komunitas Pecinta Kereta",
      "body": "Budi Santoso: \"Foto perjalanan Argo Bromo kemarin keren banget!\"",
      "image_url": "https://cdn.kai.id/news/thumb_uuid.jpg",
      "module": "community",
      "event_type": "post_new",
      "entity_id": "post-uuid-123",
      "click_action": "kai://communities/comm-uuid/posts/post-uuid-123",
      "bypass": false,
      "extra": {
        "community_id": "comm-uuid",
        "community_name": "Pecinta Kereta"
      },
      "is_read": false,
      "created_at": "2026-06-04T10:30:00+07:00",
      "read_at": null
    },
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "type": "security",
      "title": "⚠️ Peringatan Darurat",
      "body": "Terdapat gangguan layanan di area Jakarta Pusat",
      "module": "announcement",
      "event_type": "critical_published",
      "entity_id": "ann-uuid-456",
      "click_action": "kai://announcements/ann-uuid-456",
      "bypass": true,
      "extra": {
        "priority": "CRITICAL",
        "type": "disaster"
      },
      "is_read": true,
      "created_at": "2026-06-04T09:15:00+07:00",
      "read_at": "2026-06-04T09:20:00+07:00"
    }
  ],
  "meta": null
}
```

### Field Reference

| Field | Type | Keterangan |
|-------|------|------------|
| `id` | string (UUID) | ID delivery attempt in-app (gunakan ini untuk mark read) |
| `type` | string | Tipe notifikasi — lihat tabel di bawah |
| `title` | string | Judul notifikasi |
| `body` | string | Isi notifikasi |
| `image_url` | string \| absent | URL gambar/thumbnail. Tidak ada jika notif tidak punya visual |
| `module` | string | Modul sumber: `announcement`, `news`, `community`, `event`, `qna`, `subscription`, `region` |
| `event_type` | string | Event spesifik dalam modul — sama persis dengan `event_type` di FCM data payload |
| `entity_id` | string | UUID entitas terkait (article, community, event, dll). `""` jika tidak ada |
| `click_action` | string | Deep link navigasi Flutter — sama persis dengan `click_action` di FCM data payload |
| `bypass` | bool | `true` jika notif ini bypass preferences user (notif kritis) |
| `extra` | object \| absent | Data tambahan per event type. Map key-value string. Absent jika tidak ada |
| `is_read` | bool | `true` jika sudah dibaca |
| `created_at` | string (RFC3339) | Waktu notifikasi diterima |
| `read_at` | string (RFC3339) \| null | Waktu dibaca. `null` jika belum dibaca |

**Tipe notifikasi (`type`)**:
| Value | Deskripsi |
|-------|-----------|
| `security` | Aktivitas keamanan (login, perubahan password) |
| `social` | Interaksi sosial |
| `content` | Konten baru (post komunitas, berita, pengumuman) |
| `system` | Pesan sistem |
| `reminder` | Pengingat (event, jadwal) |
| `marketing` | Promosi dan informasi produk |

**Catatan**:
- Mengembalikan **semua** notifikasi (read & unread), maksimal 20 terbaru
- Field `module`, `event_type`, `entity_id`, `click_action`, `bypass` konsisten dengan structure FCM `data` payload — Flutter bisa pakai handler yang sama untuk FCM push dan inbox
- Field `extra` di inbox adalah JSON object (bukan JSON-encoded string seperti di FCM) — tidak perlu `jsonDecode`

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
| `id` | string (UUID) | ID notifikasi (field `id` dari response GET /notifications) |

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
  "type": "content",
  "title": "Post Baru di Komunitas Pecinta Kereta",
  "body": "Budi Santoso: \"Foto perjalanan Argo Bromo kemarin keren banget!\"",
  "image_url": "https://cdn.kai.id/news/thumb_uuid.jpg",
  "module": "community",
  "event_type": "post_new",
  "entity_id": "post-uuid-123",
  "click_action": "kai://communities/comm-uuid/posts/post-uuid-123",
  "bypass": false,
  "reference_id": "comm-uuid",
  "reference_type": "community",
  "payload": {
    "community_id": "comm-uuid",
    "community_name": "Pecinta Kereta"
  },
  "created_at": "2026-06-04T10:30:00Z"
}
```

| Field | Type | Keterangan |
|-------|------|------------|
| `id` | string (UUID) | ID delivery attempt — gunakan untuk PATCH /:id/read |
| `type` | string | Tipe notifikasi |
| `title` | string | Judul |
| `body` | string | Isi |
| `image_url` | string \| absent | URL gambar/thumbnail |
| `module` | string | Modul sumber notif — sama dengan FCM `data.module` |
| `event_type` | string | Event spesifik — sama dengan FCM `data.event_type` |
| `entity_id` | string | UUID entitas terkait — sama dengan FCM `data.entity_id` |
| `click_action` | string | Deep link navigasi — sama dengan FCM `data.click_action` |
| `bypass` | bool | `true` jika notif bypass preferences |
| `reference_id` | string \| absent | ID referensi entitas (field lama, tetap ada) |
| `reference_type` | string \| absent | Tipe referensi (field lama, tetap ada) |
| `payload` | object \| absent | Map extra data dari notifikasi (field lama, tetap ada) |
| `created_at` | string (RFC3339) | Waktu notifikasi dibuat |

**Catatan Penting**:
- Server **hanya mengirim**, client tidak perlu kirim data (kecuali ping keepalive)
- Jika koneksi WS terputus, notifikasi tetap tersimpan di DB — client cukup hit `GET /notifications` saat reconnect untuk catch-up
- Field `module`, `event_type`, `entity_id`, `click_action`, `bypass` konsisten dengan FCM push dan REST inbox — Flutter pakai handler navigasi yang sama untuk ketiga channel

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
  ├─ 1. Hit GET /notifications              ← Load 20 notif terbaru dari DB
  ├─ 2. Hit GET /notifications/unread-count ← Update badge count
  ├─ 3. Connect WS /ws/notifications        ← Listen real-time
  │
  │  [Saat WS menerima pesan]
  ├─ 4. Append ke list lokal                ← Tidak perlu re-fetch
  ├─ 5. Update badge count (count + 1)
  │
  │  [User tap notifikasi]
  ├─ 6. Baca click_action dari item         ← Navigasi ke deep link
  ├─ 7. PATCH /notifications/:id/read       ← Mark dibaca
  ├─ 8. Set is_read: true di list lokal
  │
  │  [User tap "baca semua"]
  └─ 9. PATCH /notifications/read-all       ← Mark semua dibaca
```

## Konsistensi dengan FCM Push

Field `module`, `event_type`, `entity_id`, `click_action`, `bypass` di REST response **identik** dengan field `data` di FCM push message. Flutter dapat menggunakan handler navigasi yang sama:

```dart
// Berlaku untuk FCM push AND inbox REST item
void handleNotificationTap(String clickAction) {
  NavigationService.navigateTo(clickAction);
}

// FCM push: extra adalah JSON-encoded string → perlu jsonDecode
// Inbox REST: extra adalah JSON object → langsung pakai
```
