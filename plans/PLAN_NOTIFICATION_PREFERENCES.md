# Plan — Notification Preferences: Gap Resolution

## Context

Spec lama (`API_SPEC_NOTIFICATION_PREFERENCES.md`) dibuat sebelum domain model final. Spec itu mendefinisikan struktur hierarkis per-modul (`news.from_kai_pusat`, `community.communities.{id}`) yang tidak pernah diimplementasikan di BE. Domain model aktual menggunakan struktur generik flat: array `EventChannelPreference` dengan `{event_type, channel, enabled}`.

Selain itu, seluruh HTTP layer (handler, use case, route) untuk preferences belum ada sama sekali. Dispatcher sudah menggunakan `preferenceRepo` tapi tidak ada endpoint untuk user mengaturnya.

Tujuan dokumen ini: memetakan gap, menentukan format preference yang benar, dan mendefinisikan endpoint baru yang sesuai domain model aktual.

---

## Bagaimana SelectChannels Bekerja (Current Implementation)

```
Dispatcher.dispatchToOne()
  └─ resolveChannels(ctx, userID, tmpl)
      ├─ prefRepo.FindByUserID(ctx, userID)
      ├─ eventType = string(tmpl.Type)   // "security", "social", "content", dll
      └─ selector.SelectChannels(pref, eventType, tmpl.Channels)
          ├─ if AllEnabled = false → return []
          ├─ if DND aktif + push → skip push
          └─ for each channel in requestedChannels:
               cek pref.IsChannelEnabledFor(eventType, channel)
               → iterate Preferences array cari {EventType, Channel}
               → jika tidak ada: default true (enabled)
```

**Kunci**: `eventType` yang digunakan adalah `string(NotificationType)` — bukan routing key event.

---

## Format Preference yang Tepat

Preferences disimpan **sparse** — hanya entry yang berbeda dari default (true = enabled):

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "all_enabled": true,
  "do_not_disturb": false,
  "dnd_window": {
    "start_time": "22:00",
    "end_time": "08:00"
  },
  "preferences": [
    {"event_type": "social",    "channel": "push",  "enabled": false},
    {"event_type": "marketing", "channel": "push",  "enabled": false},
    {"event_type": "marketing", "channel": "email", "enabled": false}
  ],
  "created_at": "2026-06-04T10:00:00Z",
  "updated_at": "2026-06-04T10:00:00Z"
}
```

**Granularitas**: per `NotificationType` × `NotificationChannel`

### Nilai `event_type` yang valid

| event_type | Artinya |
|------------|---------|
| `security` | Login, perubahan password |
| `social` | Like, komentar, mention, balasan |
| `content` | Post baru di komunitas |
| `system` | Broadcast admin, pengumuman sistem |
| `reminder` | Pengingat event, jadwal |
| `marketing` | Promosi, informasi produk |

### Nilai `channel` yang valid

| channel | Artinya |
|---------|---------|
| `in_app` | WebSocket / notification list |
| `push` | FCM push notification |
| `email` | SMTP email |
| `sms` | Fonnte WhatsApp |

### Sparse storage — default = enabled

Hanya simpan entry yang **dimatikan (enabled: false)**. Jika tidak ada entry untuk kombinasi tertentu, default-nya adalah `true` (enabled). Ini membuat storage efisien dan response bersih.

---

## Endpoint Baru (Sesuai Domain Model)

Base URL: `/api/v1/mobile/notifications/preferences`

### GET /preferences

Ambil preferences user. Auto-create dengan default jika user belum punya.

**Response 200**:
```json
{
  "status": "success",
  "status_code": 200,
  "message": "preferences berhasil diambil",
  "data": {
    "id": "uuid",
    "user_id": "uuid",
    "all_enabled": true,
    "do_not_disturb": false,
    "dnd_window": null,
    "preferences": [],
    "created_at": "...",
    "updated_at": "..."
  }
}
```

---

### PUT /preferences/global

Update `all_enabled` dan DND window.

**Request body**:
```json
{
  "all_enabled": true,
  "do_not_disturb": true,
  "dnd_start_time": "22:00",
  "dnd_end_time": "08:00"
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|------------|
| `all_enabled` | bool | No | Master toggle semua notifikasi |
| `do_not_disturb` | bool | No | Aktifkan DND |
| `dnd_start_time` | string (HH:MM) | Conditional | Wajib jika `do_not_disturb: true` |
| `dnd_end_time` | string (HH:MM) | Conditional | Wajib jika `do_not_disturb: true` |

**Response 200**: sama seperti GET /preferences

---

### PUT /preferences/channels

Batch upsert per-event_type + channel. Hanya kirim entry yang ingin diubah.

**Request body**:
```json
{
  "preferences": [
    {"event_type": "social",    "channel": "push",  "enabled": false},
    {"event_type": "marketing", "channel": "email", "enabled": false},
    {"event_type": "social",    "channel": "email", "enabled": true}
  ]
}
```

| Field | Type | Keterangan |
|-------|------|------------|
| `event_type` | string | Salah satu dari 6 NotificationType |
| `channel` | string | `in_app`, `push`, `email`, atau `sms` |
| `enabled` | bool | false = matikan, true = aktifkan |

**Behaviour**: upsert — jika entry sudah ada di-update, jika belum ada di-insert.

**Response 200**: sama seperti GET /preferences (state terbaru)

---

### POST /preferences/reset

Reset semua preference ke default: `all_enabled = true`, DND = false, `preferences = []`.

**Request body**: tidak diperlukan

**Response 200**: sama seperti GET /preferences (state setelah reset)

---

## Gap dari Spec Lama

| Spec Lama | Status | Resolusi |
|-----------|--------|----------|
| Struktur nested per-modul (`news`, `community`, `event`, `qna`) | ❌ Tidak ada di domain | **Hapus dari spec** — ganti dengan generic `event_type` |
| `all_notifications_enabled` | ✓ Ada sebagai `AllEnabled` | Rename ke `all_enabled` di response |
| `do_not_disturb_enabled` + `start_time` + `end_time` | ✓ Ada (`DoNotDisturb` + `DoNotDisturbWindow`) | Flatten jadi 3 field di request, nested `dnd_window` di response |
| `PUT /preferences/modules/{module}` | ❌ | **Ganti** dengan `PUT /preferences/channels` |
| `PUT /preferences/communities/{community_id}` | ❌ Tidak ada di domain | **Hapus** — domain tidak model per-community preference |
| `POST /preferences/reset` | ❌ Belum ada handler | **Tetap ada**, implementasi hapus `Preferences` array |
| HTTP layer keseluruhan | ❌ Tidak ada handler/usecase/route | **Implementasi baru** |

---

## Yang Perlu Diimplementasikan di BE

### DTO (baru)
```
internal/app/dto/notification_preference_dto.go
  - NotificationPreferenceResponse
  - UpdateGlobalPreferenceInput
  - UpdateChannelPreferencesInput
  - EventChannelPreferenceItem
```

### Use Cases (baru)
```
internal/app/usecase/notification/
  get_preferences.go             — FindByUserID, auto-create jika tidak ada
  update_global_preferences.go   — update AllEnabled + DND
  update_channel_preferences.go  — batch upsert EventChannelPreference
  reset_preferences.go           — reset ke default
```

Tambah ke `NotificationUseCases` struct di `notification_usecases.go`.

### Handler (baru)
```
internal/interfaces/http/handler/mobile/notification_preference_handler.go
  - GetPreferences()
  - UpdateGlobal()
  - UpdateChannels()
  - Reset()
```

### Routes (tambah ke router.go, di bawah `/notifications` group)
```
GET  /notifications/preferences
PUT  /notifications/preferences/global
PUT  /notifications/preferences/channels
POST /notifications/preferences/reset
```

### Spec (update)
```
K-FORUM-DOCS/API SPEC/Mobile/API_SPEC_NOTIFICATION_PREFERENCES.md
  — Ganti struktur hierarkis dengan generic format
  — Update semua endpoint dan request/response example
```

---

## File Referensi BE

| File | Status | Peran |
|------|--------|-------|
| `domain/notification/entity/notification_preference.go` | ✓ Ada | Domain entity, tidak perlu diubah |
| `domain/notification/service/channel_selector.go` | ✓ Ada | Selector logic, tidak perlu diubah |
| `domain/notification/repository/interfaces.go` | ✓ Ada | `NotificationPreferenceRepository` |
| `infrastructure/persistence/postgres_notification_preference_repository.go` | ✓ Ada | Implementasi persistence |
| `app/usecase/notification/notification_usecases.go` | ✓ Ada | Tambah use case preference di sini |

---

## Verification

1. `GET /preferences` saat user baru → auto-create dan return default (`all_enabled: true`, `preferences: []`)
2. `PUT /preferences/global` matikan `all_enabled` → login lagi → tidak ada notif tersimpan di DB, tidak ada push
3. `PUT /preferences/channels` matikan push untuk `social` → like post → push tidak dikirim, in_app tetap berjalan
4. Set DND `22:00-07:00`, test jam 23:00 → push tidak dikirim, in_app tetap
5. `POST /preferences/reset` → state kembali ke default, semua channel enabled
