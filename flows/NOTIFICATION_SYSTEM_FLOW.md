# Notification System Flow — K-Forum API

Dokumen ini menggambarkan arsitektur dan alur kerja notification system di K-Forum API, mencakup dua jalur pengiriman: **event-driven via MQ** (untuk notifikasi otomatis dari aksi user) dan **manual send via Backoffice** (untuk pengiriman oleh admin).

---

## Daftar Isi

- [Gambaran Umum](#gambaran-umum)
- [Komponen](#komponen)
- [Dua Jalur Pengiriman](#dua-jalur-pengiriman)
- [Jalur 1 — Event-Driven (via MQ)](#jalur-1--event-driven-via-mq)
- [Jalur 2 — Manual Send (Backoffice)](#jalur-2--manual-send-backoffice)
- [Studi Kasus — News Published → Notifikasi ke Semua User](#studi-kasus--news-published--notifikasi-ke-semua-user)
- [Preference Check & DND Window](#preference-check--dnd-window)
- [Device Resolution & FCM Delivery](#device-resolution--fcm-delivery)
- [Notification Status Lifecycle](#notification-status-lifecycle)
- [DeliveryAttempt & Retry](#deliveryattempt--retry)
- [Database Tables](#database-tables)
- [Folder Structure](#folder-structure)
- [Event Types yang Terdaftar](#event-types-yang-terdaftar)
- [Cara Menambah Notification Trigger Baru](#cara-menambah-notification-trigger-baru)
- [Status Implementasi](#status-implementasi)

---

## Gambaran Umum

```
┌─────────────────────────────────────────────────────────────────────────┐
│  JALUR 1 — EVENT-DRIVEN                                                 │
│                                                                         │
│  UseCase → event_outbox → OutboxRelay → RabbitMQ → MQHandler           │
│                                                    ↓                   │
│                                            Preference Check             │
│                                                    ↓                   │
│                                  Notification + DeliveryAttempt (QUEUED)│
│                                                    ↓                   │
│                                          Device Resolution              │
│                                                    ↓                   │
│                                            FCM Push Sender              │
│                                                    ↓                   │
│                               DeliveryAttempt (SUCCESS/RETRYING/FAILED) │
│                                                    ↓                   │
│                                  Notification (SENT / FAILED / READ)    │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│  JALUR 2 — MANUAL SEND (BACKOFFICE)                                     │
│                                                                         │
│  POST /send atau /broadcast (admin)                                     │
│                    ↓                                                    │
│          NotificationSendLog (audit)                                    │
│                    ↓                                                    │
│            Preference Check (per user)                                  │
│                    ↓                                                    │
│                Device Resolution                                        │
│                    ↓                                                    │
│              FCM Push Sender                                            │
│                    ↓                                                    │
│    [TODO] DeliveryAttempt (SUCCESS/RETRYING/FAILED) per channel         │
│                    ↓                                                    │
│       NotificationSendLogResult (per user)                              │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Komponen

| Komponen | Lokasi | Peran |
|---|---|---|
| `Notification` | `internal/domain/notification/entity/notification.go` | Aggregate root — record notifikasi per penerima |
| `DeviceRegistration` | `internal/domain/notification/entity/device_registration.go` | Token FCM/APNS per device user |
| `DeliveryAttempt` | `internal/domain/notification/entity/delivery_attempt.go` | Tracking pengiriman per channel dengan retry |
| `NotificationPreference` | `internal/domain/notification/entity/notification_preference.go` | Pengaturan notifikasi per user termasuk DND |
| `NotificationSendLog` | `internal/domain/notification/entity/notification_send_log.go` | Audit log untuk operasi send dari backoffice |
| `FCMPushSender` | `internal/infrastructure/external/fcm/fcm_push_sender.go` | Implementasi pengiriman via Firebase Cloud Messaging |
| `MQHandler` | `internal/interfaces/mq/handler/auth.go` | Consumer handler untuk event dari RabbitMQ |
| `NotificationHandler` | `internal/interfaces/http/handler/web/notification_handler.go` | HTTP handler untuk backoffice API |

---

## Dua Jalur Pengiriman

### Jalur 1 — Event-Driven (via MQ)

Notifikasi otomatis yang dipicu oleh aksi user. Event domain dipublish ke RabbitMQ via Outbox Pattern, kemudian dikonsumsi oleh MQ worker yang mengirim notifikasi.

**Contoh trigger:**
- User baru mendaftar → welcome notification
- User login dari device baru → security alert
- User diundang ke region → email undangan

### Jalur 2 — Manual Send (Backoffice)

Pengiriman langsung oleh admin melalui API backoffice. Digunakan untuk announcement, campaign marketing, atau notifikasi darurat.

**Dua mode:**
- `POST /send` — kirim ke user spesifik (by user_ids[])
- `POST /broadcast` — kirim ke semua user atau berdasarkan filter (platform, subscription_plan)

---

## Jalur 1 — Event-Driven (via MQ)

> Alur lengkap Outbox Pattern dan Job lifecycle ada di [JOB_SYSTEM_FLOW.md](./JOB_SYSTEM_FLOW.md). Bagian ini fokus pada notification-specific logic di sisi handler.

```
UseCase (API server)
  │
  ├── 1. Operasi DB utama (register user, login, dll.)
  │
  ├── 2. Serialize domain event → json.RawMessage
  │
  └── 3. outboxRepo.Save(entry)   ← satu transaksi DB
              │
              │  PENDING di event_outbox
              ▼
        OutboxRelay (polling 2 detik)
              │
              └── Publish ke RabbitMQ sebagai JobMessage
                          │
                          ▼
                   MQ Consumer (Worker)
                          │
                  JobConsumerMiddleware
                          │
                          ▼
                   MQHandler.Handle*()
                          │
                ┌─────────┴──────────┐
                │                    │
                ▼                    ▼
       Preference Check        (skip if disabled/DND)
                │
                ▼
        Device Resolution
         (active tokens)
                │
                ▼
        FCMPushSender.SendBatch()
                │
                ▼
          Firebase → User Device
```

### Handler yang terdaftar

```go
// internal/interfaces/mq/handler/auth.go
func (h *MQHandler) HandleUserRegistered(ctx context.Context, msg mq.JobMessage) error { ... }
func (h *MQHandler) HandleUserVerified(ctx context.Context, msg mq.JobMessage) error { ... }
func (h *MQHandler) HandleUserLoggedIn(ctx context.Context, msg mq.JobMessage) error { ... }
func (h *MQHandler) HandleRegionInvitationEmailRequested(ctx context.Context, msg mq.JobMessage) error { ... }
```

Handler wajib return `RetryableError` atau `FatalError` — tidak boleh return generic error. Middleware yang memutuskan apakah job diretry atau masuk DEAD.

---

## Jalur 2 — Manual Send (Backoffice)

### Send ke User Spesifik

```
POST /api/v1/web/notifications/send
  Body: { user_ids[], notification: {title, body, type, ...}, bypass_preferences }
  Auth: BearerAuth + Superadmin
  │
  ▼
SendToUsersUseCase.Execute()
  │
  ├── 1. Buat NotificationSendLog (audit, status: processing)
  │
  └── 2. Untuk setiap user_id:
          │
          ├── a. Verifikasi user exists
          │
          ├── b. Preference Check (skip jika bypass_preferences=true)
          │       └── jika semua notif disabled → AddResult(skipped)
          │
          ├── c. FindActiveDevices(user_id)
          │       └── jika tidak ada device aktif → AddResult(failed)
          │
          ├── d. Build PushMessage per device
          │
          ├── e. pushSender.SendBatch(messages)
          │       ├── OK  → AddResult(success, devices_sent)
          │       └── ERR → AddResult(failed, error)
          │
          │   [TODO] Setelah step e, seharusnya ada:
          │       ├── Buat/update DeliveryAttempt per channel
          │       │       success → attempt.MarkSuccess()
          │       │       fail    → attempt.ScheduleRetry() atau attempt.MarkFailed()
          │       └── Update Notification.Status berdasarkan hasil attempt
          │
          └── f. Update NotificationSendLog stats
                  (total_targets, total_sent, total_skipped, total_failed)

Response: { operation_id, stats, results[] }
```

> **Gap saat ini:** `processUserSend()` memanggil FCM langsung tanpa membuat record `Notification` maupun `DeliveryAttempt`. Artinya tidak ada tracking retry per channel, dan `Notification.Status` tidak pernah diupdate dari jalur ini. Lihat [DeliveryAttempt & Retry](#deliveryattempt--retry) untuk desain lengkapnya.

### Broadcast ke Semua User

```
POST /api/v1/web/notifications/broadcast
  Body: { notification, filters: {platform?, subscription_plan?}, bypass_preferences, confirm: true }
  Auth: BearerAuth + Superadmin
  │
  ▼
BroadcastUseCase.Execute()
  │
  ├── 1. Validasi confirm=true (mandatory — cegah broadcast tidak sengaja)
  │
  ├── 2. Resolve target user IDs:
  │       DeviceRepo.FindDistinctActiveUserIDs()
  │       Filter opsional:
  │         - platform: android | ios | web
  │         - subscription_plan: active (cek subscription aktif)
  │
  └── 3. Sama seperti SendToUsersUseCase untuk setiap user

Response: { operation_id, stats, filters_applied, sent_by_name, timestamps }
```

### Preview Broadcast (Dry Run)

```
POST /api/v1/web/notifications/broadcast/preview
  Body: { filters: {platform?, subscription_plan?} }
  │
  ▼
PreviewBroadcastUseCase.Execute()
  │
  └── Hitung jumlah user yang akan terdampak tanpa kirim notif

Response: { estimated_count }
```

---

## Studi Kasus — News Published → Notifikasi ke Semua User

Skenario: Admin mempublikasikan berita baru (announcement). Saat dipublish, sistem secara otomatis mengirim push notification ke seluruh user aktif — dengan tetap menghormati preferensi notifikasi masing-masing user.

Ini adalah jalur **event-driven**, bukan manual backoffice broadcast. Trigger-nya adalah aksi publish announcement, bukan admin yang menekan tombol kirim notifikasi secara terpisah.

---

### Konteks Domain Announcement

Di K-Forum API, "news" merupakan bagian dari domain **Announcement** dengan:

| Field | Nilai untuk News |
|---|---|
| `type` | `info` |
| `subtype` | `company_news` |
| `priority` | `low` atau `medium` (info type hanya boleh low/medium) |
| `scope` | `global` (semua user aktif) atau `regional` (member region tertentu) |

Priority menentukan perilaku notifikasi melalui tiga method di entity `Announcement`:

```go
// internal/domain/announcement/entity/announcement.go

func (a *Announcement) ShouldSendImmediatePush() bool {
    // hanya CRITICAL dan HIGH yang kirim push segera
    return a.Priority == critical || a.Priority == high
}

func (a *Announcement) ShouldBypassNotificationSettings() bool {
    // hanya CRITICAL yang abaikan preference user
    return a.Priority == critical
}

func (a *Announcement) ShouldSendEmail() bool {
    // INFO type tidak pernah kirim email
    switch a.Type {
    case info: return false
    ...
    }
}
```

Untuk `info` type (news) dengan priority `low`/`medium`:
- `ShouldSendImmediatePush()` = **false** → push hanya jika priority HIGH atau CRITICAL
- `ShouldBypassNotificationSettings()` = **false** → preferensi user **wajib dicek**
- `ShouldSendEmail()` = **false** → tidak ada email

---

### Alur dari Publish ke Notifikasi (Desain)

> **Status implementasi:** `PublishAnnouncementUseCase` saat ini sudah menghitung penerima dan menentukan apakah push/email harus dikirim, tapi **belum** membuat outbox entry. MQ handler untuk announcement event **belum** ada. Bagian di bawah mendokumentasikan desain lengkap yang perlu diimplementasikan.

```
Admin
  │
  └── POST /api/v1/web/announcements/{id}/publish
              │
              ▼
  PublishAnnouncementUseCase.Execute()
              │
              ├── [1] Fetch announcement dari DB
              │
              ├── [2] AnnouncementPublishProcessor.Process()
              │         │
              │         ├── countRecipients():
              │         │     scope=global  → COUNT users WHERE status='ACTIVE'
              │         │     scope=regional → COUNT region_memberships WHERE region_id=X
              │         │
              │         ├── ShouldSendImmediatePush()?
              │         │     HIGH/CRITICAL → totalSent = totalRecipients
              │         │     LOW/MEDIUM    → totalSent = 0  ← news tidak push langsung
              │         │
              │         └── Return AnnouncementPublishResult { totalRecipients, totalSent, events }
              │
              ├── [3] announcement.Publish(now, totalRecipients)
              │         └── Status: DRAFT → PUBLISHED
              │
              ├── [4] announcement.UpdateDeliveryStats(totalRecipients, totalSent)
              │
              ├── [5] announcementRepo.Update()  ← simpan ke DB
              │
              └── [6] TODO: outboxRepo.Save(OutboxEntry{
                            RoutingKey: "announcement.published",
                            Payload:    { announcement_id, type, subtype, priority, scope, region_id }
                          })
                          ← dalam satu transaksi yang sama dengan Update di atas
```

Setelah outbox entry tersimpan, flow berlanjut via OutboxRelay (sama seperti event lain):

```
OutboxRelay (polling 2 detik)
  │
  └── FindPending() → dapat entry "announcement.published"
            │
            └── Publish ke RabbitMQ sebagai JobMessage:
                  {
                    job_id:      "uuid",
                    type:        "announcement.published",
                    occurred_at: "...",
                    payload:     { announcement_id, type, priority, scope, ... }
                  }
                  │
                  ▼
            MarkProcessed() di event_outbox
```

---

### MQ Handler — `HandleAnnouncementPublished` (TODO)

```go
// internal/interfaces/mq/handler/announcement.go  ← perlu dibuat

func (h *MQHandler) HandleAnnouncementPublished(ctx context.Context, msg mq.JobMessage) error {
    var payload announcementevent.AnnouncementPublished
    if err := json.Unmarshal(msg.Payload, &payload); err != nil {
        return job.NewFatalError("gagal unmarshal AnnouncementPublished payload", err)
    }

    // Fetch announcement untuk baca priority, scope, dll.
    announcement, err := h.announcementRepo.FindByID(ctx, payload.AnnouncementID)
    if err != nil { ... }

    // Hanya kirim push jika priority HIGH atau CRITICAL
    if !announcement.ShouldSendImmediatePush() {
        return nil  // news (low/medium) tidak kirim push segera
    }

    // Resolve target users berdasarkan scope
    userIDs, err := h.resolveTargetUsers(ctx, announcement)
    if err != nil {
        return job.NewRetryableError("gagal resolve target users", err)
    }

    bypassPrefs := announcement.ShouldBypassNotificationSettings()

    // Kirim per user dengan preference check
    for _, userID := range userIDs {
        if err := h.sendToUser(ctx, userID, announcement, bypassPrefs); err != nil {
            // log partial failure, lanjut ke user berikutnya
        }
    }
    return nil
}
```

---

### Detail Per User — Preference Check & FCM Send

Untuk setiap user target, handler melakukan:

```
sendToUser(userID, announcement, bypassPrefs)
  │
  ├── [STEP 1] Preference Check
  │       │
  │       │   Dilewati jika bypassPrefs=true (hanya CRITICAL priority)
  │       │
  │       └── notifPrefRepo.FindDisabledUserIDs(ctx, [userID])
  │               │
  │               ├── User punya record dengan all_enabled = false?
  │               │       └── skip → tidak kirim, lanjut ke user berikutnya
  │               │
  │               └── User tidak punya record / all_enabled = true?
  │                       └── Dianggap enabled (default) → lanjut
  │
  ├── [STEP 2] Device Resolution
  │       │
  │       └── deviceRepo.FindActiveByUserID(userID)
  │               │
  │               ├── Tidak ada device aktif → skip (user belum login di device manapun)
  │               │
  │               └── Ada device aktif → lanjut
  │
  ├── [STEP 3] Build PushMessage per Device
  │       │
  │       └── Untuk setiap device aktif user:
  │               {
  │                 Token:    device.ProviderToken,
  │                 Title:    announcement.Title,
  │                 Body:     announcement.Body,
  │                 Priority: "high",
  │                 Payload: {
  │                   "type":         "announcement",
  │                   "entity_id":    announcement.ID,
  │                   "click_action": "OPEN_ANNOUNCEMENT"
  │                 }
  │               }
  │
  └── [STEP 4] FCM SendBatch
          │
          └── pushSender.SendBatch(ctx, []PushMessage)
                  ├── Firebase SendEach() per token
                  ├── UNREGISTERED token → skip, tidak error
                  └── Error FCM → log, catat sebagai partial failure
```

---

### Matriks Perilaku Berdasarkan Priority

| Priority | Type yang Boleh | `ShouldSendImmediatePush()` | `ShouldBypassNotificationSettings()` | `ShouldSendEmail()` |
|---|---|---|---|---|
| `critical` | disaster, urgent, system | ✓ Ya | ✓ Ya (abaikan preference) | ✓ Ya (semua type) |
| `high` | disaster, urgent, system | ✓ Ya | ✗ Tidak (cek preference) | ✓ Ya (disaster, system) |
| `medium` | system, info | ✗ Tidak | ✗ Tidak | ✗ Tidak |
| `low` | info | ✗ Tidak | ✗ Tidak | ✗ Tidak |

**Untuk news (`info` type, `low`/`medium` priority):**
- Push tidak dikirim segera (`ShouldSendImmediatePush()` = false)
- Preference tetap dicek jika suatu saat push ditambahkan
- Email tidak dikirim
- Notifikasi in-app tetap bisa dibuat (channel `in_app`, tersedia saat fitur in-app diimplementasikan)

---

### Kemungkinan Status per User

| Kondisi | Hasil |
|---|---|
| Priority `low`/`medium` (info/news) | Push tidak dikirim (return nil di handler) |
| Priority `critical` | Push dikirim, preference diabaikan |
| Priority `high`, user punya `all_enabled = false` | Skip (preference dihormati) |
| Priority `high`, user tidak punya device aktif | Skip (tidak ada token FCM) |
| Priority `high`, FCM berhasil | Push terkirim ke semua device aktif user |
| FCM token UNREGISTERED | Token di-skip, device lain tetap dicoba |

---

### Yang Perlu Diimplementasikan

Untuk flow ini berjalan penuh, beberapa hal perlu ditambahkan:

1. **Domain event** untuk announcement publish:
   ```go
   // internal/domain/announcement/event/routing.go
   const (
       RoutingAnnouncementPublished = "announcement.published"
       QueueAnnouncementPublished   = "announcement.published"
   )
   ```

2. **Outbox entry di `PublishAnnouncementUseCase`** — dalam satu transaksi dengan `announcementRepo.Update()`:
   ```go
   entry := outbox.NewOutboxEntry("announcement.published", payload)
   outboxRepo.Save(ctx, entry)
   ```

3. **MQ handler `HandleAnnouncementPublished`** — daftarkan di `RegisterAll()`:
   ```go
   r.Register(announcementevent.RoutingAnnouncementPublished, h.HandleAnnouncementPublished)
   ```

4. **Consumer di `cmd/worker/main.go`**:
   ```go
   consumer.Consume(
       announcementevent.QueueAnnouncementPublished,
       announcementevent.RoutingAnnouncementPublished,
       retryPolicy.MaxRetryFor(announcementevent.RoutingAnnouncementPublished),
       middleware.Handle,
   )
   ```

---

## Preference Check & DND Window

Sebelum mengirim, sistem mengecek pengaturan notifikasi user (kecuali `bypass_preferences=true`).

```go
// internal/domain/notification/entity/notification_preference.go

type NotificationPreference struct {
    AllEnabled     bool
    DoNotDisturb   bool
    DndWindow      DoNotDisturbWindow   // { StartTime, EndTime } format "HH:MM"
    Preferences    []PreferenceItem     // [{ EventType, Channel, Enabled }]
}
```

### Logika check

```
IsChannelEnabledFor(eventType, channel)
  │
  ├── AllEnabled == false → return false (semua channel mati)
  │
  ├── DoNotDisturb == true && DndWindow.IsActive(now) → return false untuk channel push
  │
  └── Cek per-item preferences:
          └── item.EventType == eventType && item.Channel == channel → return item.Enabled
```

### DND Window — midnight crossover

```go
// IsActive() mendukung DND window yang melewati tengah malam
// Contoh: 22:00 - 06:00

func (w DoNotDisturbWindow) IsActive(now time.Time) bool {
    if w.Start <= w.End {
        return now >= w.Start && now < w.End   // same day
    }
    return now >= w.Start || now < w.End       // crosses midnight
}
```

---

## Device Resolution & FCM Delivery

```
FindActiveByUserID(userID)
  │
  └── Returns []DeviceRegistration (platform, provider_token, push_provider)
              │
              ▼
        Build PushMessage per device:
          {
            Token:    device.ProviderToken,
            Title:    notification.Title,
            Body:     notification.Body,
            ImageURL: notification.Payload.ImageURL,
            Payload:  { type, entity_id, click_action, deep_link, ...extra }
          }
              │
              ▼
        FCMPushSender.SendBatch([]PushMessage)
              │
              ├── Android: SetAndroidConfig (priority: high)
              ├── iOS:     SetAPNSConfig (sound: default, content_available: true)
              └── Web:     WebPush config
              │
              ▼
        firebase.messaging.SendEach()
              │
              ├── Success → log success per token
              ├── UNREGISTERED token → skip (tidak error, token sudah invalid)
              └── Error lain → catat sebagai partial failure
              │
              ▼
        [DESAIN] Update DeliveryAttempt berdasarkan hasil:
              │
              ├── Semua token OK:
              │     attempt.MarkSuccess()
              │     deliveryAttemptRepo.Update(attempt)
              │     notification.MarkAsSent()
              │     notificationRepo.Update(notification)
              │
              ├── Gagal, masih bisa retry (RetryCount < 3):
              │     attempt.ScheduleRetry(nextRetryAt)
              │     deliveryAttemptRepo.Update(attempt)
              │     ← Notification tetap QUEUED, retry worker akan coba lagi
              │
              └── Gagal permanen (RetryCount >= 3):
                    attempt.MarkFailed(providerCode)
                    deliveryAttemptRepo.Update(attempt)
                    notification.MarkAsFailed()
                    notificationRepo.Update(notification)
```

> **Gap saat ini:** Langkah update `DeliveryAttempt` dan `Notification.Status` setelah FCM send **belum diimplementasikan** di kedua jalur. FCMPushSender hanya return error/nil — pemanggil tidak membuat record apapun dari hasil tersebut.

### NoopPushSender

Di environment development/test, jika credentials FCM tidak dikonfigurasi, sistem menggunakan `NoopPushSender` — menerima semua request tanpa benar-benar mengirim ke Firebase.

---

## Notification Status Lifecycle

`Notification` adalah aggregate root yang statusnya digerakkan oleh hasil `DeliveryAttempt`. Satu `Notification` dibuat per penerima, dan bisa memiliki banyak `DeliveryAttempt` (satu per channel, atau satu per retry attempt pada channel yang sama).

```
              ┌─────────┐
              │ PENDING │  ← Notification dibuat oleh handler/use case
              └────┬────┘
                   │  Enqueue()  ← setelah preference check lolos
                   ▼
             ┌──────────┐
             │  QUEUED  │  ← siap dikirim; DeliveryAttempt dibuat di sini
             └─────┬────┘        (status awal DeliveryAttempt: retrying)
                   │
                   │  FCM dipanggil → hasil dilaporkan ke DeliveryAttempt
                   │
          ┌────────┴─────────────────────┐
          │                              │
          │ Minimal 1 channel sukses     │ Semua channel gagal permanen
          ▼                              ▼
      ┌──────┐                      ┌────────┐
      │ SENT │                      │ FAILED │
      └──────┘                      └────────┘
       (MarkAsSent)                  (MarkAsFailed)
          │
          │  User membuka notif (MarkAsRead)
          ▼
       ┌──────┐
       │ READ │
       └──────┘
```

| Transisi | Method | Digerakkan oleh |
|---|---|---|
| PENDING → QUEUED | `Enqueue()` | Use case / MQ handler setelah preference check lolos |
| QUEUED → SENT | `MarkAsSent()` | Minimal 1 `DeliveryAttempt` di channel manapun berhasil |
| QUEUED → FAILED | `MarkAsFailed()` | Semua `DeliveryAttempt` sudah exhausted (RetryCount >= 3) |
| SENT → READ | `MarkAsRead()` | User membuka notifikasi (mobile client memanggil endpoint read) |

### Hubungan Notification ↔ DeliveryAttempt

```
Notification (1)
    │
    └── DeliveryAttempt (1..N) — satu per channel yang aktif
            │
            ├── channel: push    → FCMPushSender
            ├── channel: email   → EmailSender (TODO)
            ├── channel: in_app  → InAppStore  (TODO)
            └── channel: sms     → SMSSender   (TODO)
```

Setiap `DeliveryAttempt` berdiri sendiri per channel. Jika push gagal tapi email berhasil, `Notification` tetap SENT — karena minimal satu channel sukses. Status `Notification` bukan AND semua channel, melainkan OR (cukup satu yang berhasil).

---

## DeliveryAttempt & Retry

`DeliveryAttempt` adalah entitas yang mencatat **setiap upaya pengiriman per channel** dari sebuah notifikasi. Ia adalah lapisan di antara `Notification` (yang hanya tahu status akhir) dan provider pengiriman (FCM, email, dll.) yang bisa gagal dan perlu diretry.

### Struktur Entity

```go
// internal/domain/notification/entity/delivery_attempt.go

type DeliveryAttempt struct {
    ID             string
    NotificationID string
    Channel        NotificationChannel  // push | email | sms | in_app
    Status         DeliveryStatus       // retrying | success | failed
    ProviderCode   *string              // kode error dari provider (FCM error code, HTTP status, dll.)
    RetryCount     int                  // berapa kali sudah dicoba, max 3
    NextRetryAt    *time.Time           // kapan boleh dicoba lagi (nil jika belum ada jadwal)
    AttemptedAt    time.Time
}
```

**Penting:** Status awal `DeliveryAttempt` saat baru dibuat adalah **`retrying`** (bukan "pending"), karena ia dibuat tepat sebelum percobaan pertama — artinya "sedang dalam proses, siap dicoba".

### Lifecycle DeliveryAttempt

```
NewDeliveryAttempt(id, notificationID, channel)
  │
  └── Status: retrying, RetryCount: 0
              │
              ▼ percobaan kirim ke provider
              │
    ┌─────────┴────────────────┐
    │                          │
    ▼ Berhasil                 ▼ Gagal
    │                          │
attempt.MarkSuccess()    IsRetryable()?  (RetryCount < 3)
    │                     │         │
    │                     │ Ya      │ Tidak
    │                     ▼         ▼
    │          attempt.ScheduleRetry(at)   attempt.MarkFailed(providerCode)
    │            RetryCount++               Status: failed (permanen)
    │            NextRetryAt = at
    │            Status: tetap retrying
    │                     │
    ▼                      └──────────────┐
Status: success                           │
deliveryAttemptRepo.Update()              │
                              Background retry worker (TODO)
                              polling FindPendingRetries()
                              ketika NextRetryAt <= now
                              → ulangi percobaan kirim
```

### Method & Peran Masing-masing

| Method | Yang Dilakukan | Kapan Dipanggil |
|---|---|---|
| `NewDeliveryAttempt()` | Buat record baru, status: `retrying` | Sebelum percobaan pertama kirim ke provider |
| `MarkSuccess()` | Status → `success`, hapus ProviderCode | FCM/provider konfirmasi terkirim |
| `ScheduleRetry(at)` | RetryCount++, set NextRetryAt, status tetap `retrying` | Gagal tapi masih bisa diretry |
| `MarkFailed(code)` | Status → `failed`, simpan ProviderCode | RetryCount sudah mencapai 3 (permanent failure) |
| `IsRetryable()` | Return `RetryCount < 3` | Setelah gagal, untuk tentukan retry atau permanent fail |

### Integrasi dengan Notification.Status

`DeliveryAttempt` adalah yang menggerakkan transisi status `Notification`:

```
Setelah percobaan ke channel push selesai:
  │
  ├── attempt.Status == success
  │       └── notification.MarkAsSent()   → Notification: QUEUED → SENT
  │
  ├── attempt.Status == retrying (terjadwal retry)
  │       └── Notification tetap QUEUED  ← menunggu retry worker
  │
  └── attempt.Status == failed (semua retry exhausted)
          │
          └── Cek apakah SEMUA channel sudah failed?
                  Ya  → notification.MarkAsFailed()  → Notification: QUEUED → FAILED
                  Tidak → channel lain masih bisa sukses, Notification tetap QUEUED
```

### Retry Worker — Desain (TODO)

Background worker yang polling `delivery_attempts` untuk menemukan record yang perlu diretry:

```
RetryWorker (goroutine, interval N detik)  ← TODO: belum diimplementasi
  │
  └── deliveryAttemptRepo.FindPendingRetries()
        └── Query: status = 'retrying' AND next_retry_at <= NOW()
              │
              └── Untuk setiap attempt yang ditemukan:
                    │
                    ├── Ambil Notification terkait
                    │
                    ├── Lakukan send ulang ke provider (FCM/email/dll.)
                    │
                    ├── Berhasil:
                    │     attempt.MarkSuccess()
                    │     notification.MarkAsSent()
                    │
                    └── Gagal lagi:
                          attempt.IsRetryable()?
                            Ya  → attempt.ScheduleRetry(nextAt)  ← jadwalkan lagi
                            Tidak → attempt.MarkFailed(code)
                                    cek semua channel → notification.MarkAsFailed()
```

### Status Implementasi

| Komponen | Status |
|---|---|
| Entity `DeliveryAttempt` | ✓ Implemented (`delivery_attempt.go`) |
| Repository interface `DeliveryAttemptRepository` | ✓ Defined (`interfaces.go`) |
| Tabel `delivery_attempts` di DB | ✓ Migration ada |
| Pembuatan `DeliveryAttempt` saat send | ✗ TODO — `processUserSend()` langsung call FCM tanpa buat record |
| Update `DeliveryAttempt` setelah FCM | ✗ TODO — hasil FCM tidak direfleksikan ke entity |
| Update `Notification.Status` dari attempt | ✗ TODO — `Notification` record tidak pernah dibuat di current flow |
| Background retry worker | ✗ TODO — `FindPendingRetries()` tersedia, worker belum ada |

**Efek gap saat ini:** Tabel `delivery_attempts` dan `notifications` selalu kosong meskipun notifikasi berhasil terkirim via FCM. Observability pengiriman per user per channel saat ini hanya tersedia via `notification_send_log_results` (untuk jalur backoffice) atau log aplikasi (untuk jalur event-driven).

---

## Database Tables

### `notifications`

```sql
CREATE TABLE notifications (
    id             UUID PRIMARY KEY,
    recipient_id   UUID NOT NULL,
    type           TEXT NOT NULL,    -- system | social | content | reminder | marketing | security
    title          TEXT NOT NULL,
    body           TEXT NOT NULL,
    payload        JSONB,            -- { deep_link, image_url, extra: map[string]string }
    channels       TEXT[],           -- { in_app, push, email, sms }
    status         TEXT NOT NULL,    -- pending | queued | sent | failed | read
    reference_id   UUID,             -- entity yang memicu notif (post_id, comment_id, dll.)
    reference_type TEXT,             -- jenis entity (post, comment, event, dll.)
    sent_at        TIMESTAMP,
    read_at        TIMESTAMP,
    created_at     TIMESTAMP NOT NULL,
    updated_at     TIMESTAMP NOT NULL
);
```

### `device_registrations`

```sql
CREATE TABLE device_registrations (
    id             UUID PRIMARY KEY,
    user_id        UUID NOT NULL,
    platform       TEXT NOT NULL,        -- android | ios | web
    push_provider  TEXT NOT NULL,        -- fcm | apns | huawei | web_push
    provider_token TEXT NOT NULL UNIQUE, -- FCM registration token
    is_active      BOOLEAN DEFAULT true,
    last_seen_at   TIMESTAMP,
    device_meta    JSONB,                -- device_model, os_version, app_version, dll.
    created_at     TIMESTAMP NOT NULL,
    updated_at     TIMESTAMP NOT NULL
);
```

### `delivery_attempts`

```sql
CREATE TABLE delivery_attempts (
    id              UUID PRIMARY KEY,
    notification_id UUID NOT NULL REFERENCES notifications(id),
    channel         TEXT NOT NULL,    -- push | email | sms | in_app
    status          TEXT NOT NULL,    -- success | failed | retrying
    retry_count     INT DEFAULT 0,
    max_retry       INT DEFAULT 3,
    provider_code   TEXT,
    provider_msg    TEXT,
    next_retry_at   TIMESTAMP,
    attempted_at    TIMESTAMP NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL
);
```

### `notification_preferences`

```sql
CREATE TABLE notification_preferences (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL UNIQUE,
    all_enabled     BOOLEAN DEFAULT true,
    do_not_disturb  BOOLEAN DEFAULT false,
    dnd_start_time  TEXT,               -- "HH:MM" format, e.g. "22:00"
    dnd_end_time    TEXT,               -- "HH:MM" format, e.g. "06:00"
    preferences     JSONB,              -- [{ event_type, channel, enabled }]
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL
);
```

### `notification_send_logs` (Backoffice audit)

```sql
CREATE TABLE notification_send_logs (
    id                UUID PRIMARY KEY,
    title             TEXT NOT NULL,
    body              TEXT NOT NULL,
    notif_type        TEXT NOT NULL,    -- custom | announcement | event | news | community
    target_mode       TEXT NOT NULL,    -- user | broadcast
    bypass_prefs      BOOLEAN DEFAULT false,
    filters_applied   JSONB,            -- { platform, subscription_plan }
    total_targets     INT DEFAULT 0,
    total_sent        INT DEFAULT 0,
    total_skipped     INT DEFAULT 0,
    total_failed      INT DEFAULT 0,
    sent_by           UUID NOT NULL,    -- admin user_id
    sent_at           TIMESTAMP,
    created_at        TIMESTAMP NOT NULL,
    updated_at        TIMESTAMP NOT NULL
);
```

### `notification_send_log_results` (Per-user hasil backoffice send)

```sql
CREATE TABLE notification_send_log_results (
    id             UUID PRIMARY KEY,
    send_log_id    UUID NOT NULL REFERENCES notification_send_logs(id),
    user_id        UUID NOT NULL,
    status         TEXT NOT NULL,    -- success | skipped | failed
    devices_sent   INT DEFAULT 0,
    devices_total  INT DEFAULT 0,
    error_msg      TEXT,
    created_at     TIMESTAMP NOT NULL
);
```

---

## Folder Structure

```
internal/
├── domain/
│   └── notification/
│       ├── constant/
│       │   ├── notification_constant.go   # Status, Type, Channel enum
│       │   ├── device_constant.go         # Platform, PushProvider enum
│       │   └── error_code.go              # Domain error codes
│       └── entity/
│           ├── notification.go            # Aggregate root + lifecycle methods
│           ├── device_registration.go     # Device token management
│           ├── delivery_attempt.go        # Per-channel delivery tracking
│           ├── notification_preference.go # User preferences + DND
│           ├── notification_send_log.go   # Backoffice operation audit
│           └── notification_vo.go         # NotificationPayload, DoNotDisturbWindow
│
├── app/
│   ├── port/
│   │   └── push_sender.go                # PushSender interface + PushMessage struct
│   └── usecase/
│       └── notificationtester/
│           ├── send_to_users.go          # Kirim ke specific users
│           ├── broadcast.go              # Kirim ke semua/filtered users
│           ├── preview_broadcast.go      # Estimasi jumlah target
│           ├── get_history.go            # Query send logs
│           ├── get_send_detail.go        # Detail satu operasi + per-user results
│           └── get_user_devices.go       # Debug: lihat device tokens user
│
├── infrastructure/
│   └── external/
│       └── fcm/
│           └── fcm_push_sender.go        # FCM implementation + NoopPushSender
│
└── interfaces/
    ├── http/
    │   └── handler/web/
    │       └── notification_handler.go   # Backoffice HTTP endpoints
    └── mq/
        └── handler/
            └── auth.go                   # MQ handlers untuk user events

```

---

## Event Types yang Terdaftar

| Routing Key | Queue | Handler | Notification Action |
|---|---|---|---|
| `user.registered` | `auth.user.registered` | `HandleUserRegistered` | TODO: welcome notification + init preferences |
| `user.verified` | `auth.user.verified` | `HandleUserVerified` | TODO: notifikasi pasca verifikasi identity |
| `user.login` | `auth.user.login` | `HandleUserLoggedIn` | Email login notification (implemented) |
| `region.invitation.email.requested` | `region.invitation.email.requested` | `HandleRegionInvitationEmailRequested` | Email undangan region (implemented) |

---

## Cara Menambah Notification Trigger Baru

Contoh: menambah push notification ketika user mendapat komentar baru di postnya.

### Langkah 1 — Definisikan domain event

```go
// internal/domain/post/event/events.go
type PostCommented struct {
    PostID      string    `json:"post_id"`
    PostOwnerID string    `json:"post_owner_id"`
    CommenterID string    `json:"commenter_id"`
    CommentID   string    `json:"comment_id"`
    OccurredAt  time.Time `json:"occurred_at"`
}

// internal/domain/post/event/routing.go
const (
    RoutingPostCommented = "post.commented"
    QueuePostCommented   = "post.commented"
)
```

### Langkah 2 — Publish event dari UseCase (dalam satu transaksi)

```go
// internal/app/usecase/comment/create_comment.go

payload, _ := json.Marshal(postevent.PostCommented{
    PostID:      postID,
    PostOwnerID: post.OwnerID,
    CommenterID: commenterID,
    CommentID:   newComment.ID,
    OccurredAt:  time.Now(),
})
entry := outbox.NewOutboxEntry(postevent.RoutingPostCommented, payload)
outboxRepo.Save(ctx, entry)  // ← satu transaksi DB dengan operasi simpan comment
```

### Langkah 3 — Tulis MQ handler

```go
// internal/interfaces/mq/handler/post.go

func (h *MQHandler) HandlePostCommented(ctx context.Context, msg mq.JobMessage) error {
    var payload postevent.PostCommented
    if err := json.Unmarshal(msg.Payload, &payload); err != nil {
        return job.NewFatalError("gagal unmarshal PostCommented payload", err)
    }

    // Cek preference owner post
    pref, err := h.notifPrefRepo.FindByUserID(ctx, payload.PostOwnerID)
    if err == nil && !pref.IsChannelEnabledFor(notifconst.EventPostComment, notifconst.ChannelPush) {
        return nil // user opt-out, skip
    }

    // Ambil device aktif pemilik post
    devices, err := h.deviceRepo.FindActiveByUserID(ctx, payload.PostOwnerID)
    if err != nil || len(devices) == 0 {
        return nil // tidak ada device, skip
    }

    // Build dan kirim push message
    messages := buildPushMessages(devices, "Komentar Baru", "Ada yang mengomentari postinganmu", payload)
    if err := h.pushSender.SendBatch(ctx, messages); err != nil {
        return job.NewRetryableError("gagal kirim push notification PostCommented", err)
    }

    return nil
}
```

### Langkah 4 — Daftarkan ke registry

```go
// internal/interfaces/mq/handler/auth.go → RegisterAll()

func RegisterAll(r *registry.Registry, h *MQHandler) {
    // ... existing handlers ...
    r.Register(postevent.RoutingPostCommented, h.HandlePostCommented)
}
```

### Langkah 5 — Tambah consumer di worker/main.go

```go
// cmd/worker/main.go → registerConsumers()

if err := consumer.Consume(
    postevent.QueuePostCommented,
    postevent.RoutingPostCommented,
    retryPolicy.MaxRetryFor(postevent.RoutingPostCommented),
    middleware.Handle,
); err != nil {
    return fmt.Errorf("gagal consume %s: %w", postevent.RoutingPostCommented, err)
}
```

---

## Backoffice API Endpoints

Semua endpoint memerlukan `BearerAuth` + permission `Superadmin`.

| Method | Endpoint | Use Case | Deskripsi |
|---|---|---|---|
| `POST` | `/api/v1/web/notifications/send` | `SendToUsersUseCase` | Kirim ke user spesifik |
| `POST` | `/api/v1/web/notifications/broadcast` | `BroadcastUseCase` | Kirim ke semua/filtered users |
| `POST` | `/api/v1/web/notifications/broadcast/preview` | `PreviewBroadcastUseCase` | Estimasi jumlah target |
| `GET` | `/api/v1/web/notifications/history` | `GetHistoryUseCase` | List riwayat send operations |
| `GET` | `/api/v1/web/notifications/history/{operation_id}` | `GetSendDetailUseCase` | Detail satu operasi + per-user results |
| `GET` | `/api/v1/web/notifications/users/{user_id}/devices` | `GetUserDevicesUseCase` | Debug: list device tokens user |

---

## Status Implementasi

### Sudah Implemented

- Domain model lengkap: `Notification`, `DeviceRegistration`, `DeliveryAttempt`, `NotificationPreference`, `NotificationSendLog`
- Database schema: semua tabel notification (migration 0009 & 0010)
- Device registration & token management (register, refresh, revoke)
- FCM push sender integration (Android, iOS, Web)
- `NoopPushSender` untuk development tanpa credentials FCM
- Backoffice API: send specific users, broadcast, preview, history, detail
- Preference check dengan DND window (termasuk midnight crossover)
- Audit logging via `NotificationSendLog` + `NotificationSendLogResult`
- MQ handler untuk `region.invitation.email.requested` (kirim email undangan)

### TODO / In-Progress

- MQ handler `HandleUserRegistered` — welcome notification + init `NotificationPreference` default
- MQ handler `HandleUserVerified` — notifikasi pasca verifikasi identity
- **Event-driven announcement notification** (studi kasus news published):
  - Outbox entry di `PublishAnnouncementUseCase`
  - Domain event `announcement.published` + routing/queue constants
  - MQ handler `HandleAnnouncementPublished` — resolve users, cek preference, kirim FCM
  - Consumer di `cmd/worker/main.go`
- **DeliveryAttempt wiring** — entity dan repository sudah ada, belum dihubungkan ke alur send:
  - Buat `Notification` record sebelum kirim
  - Buat `DeliveryAttempt` per channel sebelum call FCM/provider
  - Update `DeliveryAttempt` (MarkSuccess/ScheduleRetry/MarkFailed) setelah hasil FCM
  - Update `Notification.Status` berdasarkan outcome semua channel
- **Background retry worker** — `FindPendingRetries()` tersedia di repository, worker belum ada
- Channel `in_app` — infrastructure tersedia, belum ada endpoint untuk mobile polling/SSE
- Channel `email` untuk notification (berbeda dengan email transaksional yang sudah ada)
- Channel `sms` — belum ada provider integration

---

## Ringkasan Pola Kunci

| Pola | Implementasi |
|---|---|
| **Reliable delivery** | Transactional Outbox — event tidak hilang meski RabbitMQ down saat event terjadi |
| **Preference-aware** | Cek `notification_preferences` sebelum kirim; DND window mendukung midnight crossover |
| **Bypass for urgent** | `bypass_preferences=true` di backoffice send untuk notifikasi darurat |
| **Audit trail** | `NotificationSendLog` + `NotificationSendLogResult` per operasi backoffice |
| **Partial failure** | `SendBatch` FCM melanjutkan meski sebagian device gagal; token UNREGISTERED di-skip |
| **Retry-safe** | `DeliveryAttempt.IsRetryable()` + max 3 retry dengan backoff |
| **Dev-friendly** | `NoopPushSender` aktif otomatis jika FCM tidak dikonfigurasi |
