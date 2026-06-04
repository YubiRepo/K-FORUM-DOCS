# Notification Delivery Flow — K-Forum API (Current Implementation)

Dokumen ini menggambarkan arsitektur pengiriman notifikasi sistem terbaru, mencakup semua channel (in_app WebSocket, push FCM, email SMTP, SMS Fonnte) dan ketiga mode pengiriman (unicast, multicast, broadcast).

> Dokumen ini adalah **versi terbaru** yang menggantikan/melengkapi `NOTIFICATION_SYSTEM_FLOW.md`.
> Fokus pada implementasi aktual dengan RabbitMQ pub/sub untuk in_app cross-process delivery.

---

## Daftar Isi

1. [Arsitektur Komponen](#1-arsitektur-komponen)
2. [Channel Delivery](#2-channel-delivery)
3. [Target Mode: Unicast / Multicast / Broadcast](#3-target-mode)
4. [Flow Lengkap: Event → Notifikasi](#4-flow-lengkap-event--notifikasi)
5. [In-App Cross-Process via RabbitMQ](#5-in-app-cross-process-via-rabbitmq)
6. [Retry & Fallback](#6-retry--fallback)
7. [User Preference & DND](#7-user-preference--dnd)
8. [Event Types yang Terdaftar](#8-event-types-yang-terdaftar)
9. [Source of Truth untuk Client](#9-source-of-truth-untuk-client)
10. [Cara Menambah Event Trigger Baru](#10-cara-menambah-event-trigger-baru)

---

## 1. Arsitektur Komponen

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         K-FORUM API PROCESSES                            │
│                                                                          │
│  ┌─────────────────────────┐      ┌──────────────────────────────────┐  │
│  │   cmd/app (HTTP Server) │      │      cmd/worker (Worker)         │  │
│  │                         │      │                                  │  │
│  │  ┌──────────────┐       │      │  ┌────────────────────────────┐  │  │
│  │  │  WS Hub      │◄──────┼──────┼──│  NotificationDispatcher    │  │  │
│  │  │  (connected  │  RMQ  │      │  │  ┌──────┬──────┬─────────┐ │  │  │
│  │  │   clients)   │  pub/ │      │  │  │Push  │Email │ SMS     │ │  │  │
│  │  └──────┬───────┘  sub  │      │  │  │(FCM) │(SMTP)│(Fonnte) │ │  │  │
│  │         │               │      │  │  └──────┴──────┴─────────┘ │  │  │
│  │  ┌──────▼───────┐       │      │  └──────────────┬─────────────┘  │  │
│  │  │  WS Handler  │       │      │                 │                 │  │
│  │  │  /ws/notifs  │       │      │  ┌──────────────▼─────────────┐  │  │
│  │  └──────────────┘       │      │  │  MQ Handler                │  │  │
│  │                         │      │  │  (HandleUserLoggedIn, etc.) │  │  │
│  │  ┌──────────────┐       │      │  └──────────────┬─────────────┘  │  │
│  │  │InApp Consumer│◄──────┼──────┼──────────────── │                │  │
│  │  │(RMQ Consumer)│  RMQ  │      │  ┌──────────────▼─────────────┐  │  │
│  │  └──────────────┘  queue│      │  │  MQ Middleware + Registry  │  │  │
│  └─────────────────────────┘      │  └──────────────┬─────────────┘  │  │
│                                   │                 │                  │  │
└───────────────────────────────────┼─────────────────┼──────────────────┘  │
                                    │                 │
           ┌────────────────────────┘                 │
           │                                          │
           ▼                                          ▼
     RabbitMQ Topic Exchange                   RabbitMQ Queues
     (k-forum.events)                          (auth.user.login,
                                                notif.community.post.liked,
                                                ws.notification.inapp, ...)
```

---

## 2. Channel Delivery

| Channel | Infra | Persistent? | Real-time? | Retry? |
|---------|-------|-------------|------------|--------|
| `in_app` | WebSocket Hub + DB | ✓ DB (saat offline) | ✓ jika online | ✗ (DB sudah jadi fallback) |
| `push` | FCM (Firebase) | ✗ (FCM best-effort) | ✓ | ✓ DeliveryRetryRelay |
| `email` | SMTP | ✗ | ✗ (async) | ✗ |
| `sms` | Fonnte (WhatsApp) | ✗ | ✗ (async) | ✗ |

### In-App: Online vs Offline

```
User ONLINE (WS connected):
  Dispatcher → RabbitMQ (ws.notification.inapp)
    → InApp Consumer (HTTP server)
      → Hub.SendToUser() → WS client ← notif muncul real-time

User OFFLINE (WS disconnected):
  Dispatcher → simpan ke DB (notifications table)
    → [tidak ada WS delivery]
      → User reconnect → GET /notifications → load dari DB
```

---

## 3. Target Mode

### Unicast (1 user)

```
Dispatcher.Dispatch(unicast, recipientID)
  │
  ├─ prefRepo.FindByUserID()        ← ambil preference
  ├─ channelSelector.SelectChannels() ← filter DND, per-event toggle
  │
  ├─ (dalam 1 transaksi DB)
  │   ├─ notifRepo.Save(notification)
  │   └─ deliveryRepo.Save(inapp_attempt) → status: SUCCESS
  │
  ├─ sendInApp   → RabbitMQ → WS hub
  ├─ sendPush    → FCM batch → DeliveryAttempt
  ├─ sendEmail   → userRepo.FindByID → email → SMTP
  └─ sendSMS     → userRepo.FindByID → phone → Fonnte
```

### Multicast (list of users)

```
Dispatcher.Dispatch(multicast, recipientIDs)
  │
  ├─ Proses per batch (50 user)
  │   │
  │   ├─ Per user: resolve channels + persist notification + sendInApp
  │   │
  │   ├─ Kumpulkan push messages dari semua user dalam batch
  │   ├─ FCM.SendBatch(allPushMsgs)   ← 1 call untuk semua
  │   └─ Update DeliveryAttempt semua user berdasarkan hasil batch
```

**Perbedaan kunci vs unicast**: push di-batch per 50 user untuk efisiensi FCM call.

### Broadcast (semua user)

```
Dispatcher.Dispatch(broadcast)
  │
  ├─ in_app: Hub.Broadcast() via RabbitMQ → semua WS client online
  │
  └─ push:
      ├─ deviceRepo.FindDistinctActiveUserIDs()   ← semua user punya device
      ├─ Per batch 50 user: FCM.SendBatch()
      └─ Loop sampai semua user diproses

TIDAK membuat individual Notification record di DB.
User lihat broadcast via /announcements, bukan /notifications.
```

---

## 4. Flow Lengkap: Event → Notifikasi

### Contoh: User Login

```
1. User login via POST /mobile/auth/login
   │
   ▼
2. LoginUseCase berhasil
   └─ outboxRepo.Save({routing: "user.login", payload: {user_id, ip_address}})
      └─ dalam 1 transaksi bersama auth logic

   ▼ [ASYNC — OutboxRelay polling setiap 2 detik]

3. OutboxRelay
   ├─ FindPendingOutbox()
   ├─ jobRepo.Save(job)
   ├─ publisher.Publish("user.login", JobMessage)
   └─ outbox.MarkProcessed()

   ▼

4. RabbitMQ routes ke queue "auth.user.login"

   ▼

5. Worker MQHandler.HandleUserLoggedIn()
   └─ dispatcher.Dispatch(
        template: {type:security, title:"Login Berhasil", channels:[in_app,push,email]},
        target: {mode:unicast, recipientID: userID}
      )

   ▼

6. Dispatcher.dispatchToOne()
   ├─ [DB TX] Save notification + inapp DeliveryAttempt (SUCCESS)
   │
   ├─ sendInApp:
   │   └─ RabbitMQNotifHub.SendToUser()
   │       └─ publish ke "ws.notification.inapp"
   │
   ├─ sendPush:
   │   ├─ deviceRepo.FindActiveByUserID()
   │   ├─ FCM.SendBatch()
   │   └─ Save DeliveryAttempt (SUCCESS/RETRYING)
   │
   └─ sendEmail:
       ├─ userRepo.FindByID()
       └─ smtp.SendNotification(email, title, body)

   ▼ [parallel — HTTP Server]

7. InApp Consumer (HTTP server)
   └─ consume "ws.notification.inapp"
       └─ Hub.SendToUser(userID, msg)
           ├─ user ONLINE  → WS push → client menerima notif
           └─ user OFFLINE → message dropped (sudah ada di DB dari step 6)
```

### Contoh: Post Disukai

```
1. User like post → LikePostUseCase
   └─ outboxRepo.Save({routing: "community.post.liked", payload: {...}})

   ▼ [ASYNC]

2. Worker MQHandler.HandlePostLiked()
   ├─ Cek: likedByID != postAuthorID (tidak notif ke diri sendiri)
   └─ dispatcher.Dispatch(
        template: {type:social, title:"Post kamu disukai",
                   channels:[in_app,push]},
        target: {mode:unicast, recipientID: postAuthorID}
      )
   └─ [flow sama seperti login di atas]

```

### Contoh: Announcement Published (Broadcast)

```
1. Admin publish announcement → POST /web/announcements/:id/publish
   └─ outboxRepo.Save({routing: "announcement.published", payload: {...}})

   ▼ [ASYNC]

2. Worker MQHandler.HandleAnnouncementPublished()
   └─ dispatcher.Dispatch(
        template: {type:system, title: ann.Title, channels:[in_app,push]},
        target: {mode:broadcast}
      )

   ▼

3. dispatchBroadcast()
   ├─ sendInApp: RabbitMQNotifHub.Broadcast()
   │   └─ publish ke "ws.notification.inapp" (UserID kosong = broadcast)
   │       → InApp Consumer → Hub.Broadcast() → semua WS client online
   │
   └─ sendPush:
       ├─ deviceRepo.FindDistinctActiveUserIDs()
       └─ per batch 50: FCM.SendBatch()

[TIDAK ada row di tabel notifications — user lihat via /announcements]
```

---

## 5. In-App Cross-Process via RabbitMQ

Solusi untuk masalah worker dan HTTP server adalah proses terpisah dengan hub berbeda.

```
Worker Process                         HTTP Server Process
┌─────────────────────┐                ┌──────────────────────────┐
│ RabbitMQNotifHub    │                │ InApp Consumer           │
│ (implements         │   RabbitMQ     │ (ws.StartInAppConsumer)  │
│  port.NotifHub)     │──────────────► │                          │
│                     │ queue:         │  unmarshal               │
│ SendToUser(uid,msg) │ ws.notif.inapp │  → Hub.SendToUser()      │
│ Broadcast(msg)      │                │  → Hub.Broadcast()       │
└─────────────────────┘                └──────────┬───────────────┘
                                                  │
                                       ┌──────────▼───────────┐
                                       │  ws.Hub              │
                                       │  (connected clients) │
                                       └──────────┬───────────┘
                                                  │ WebSocket
                                                  ▼
                                            Mobile/Web clients
```

**Routing key**: `notification.inapp`
**Queue**: `ws.notification.inapp`
**Format message**:
```json
{
  "user_id": "uuid",   // kosong = broadcast ke semua
  "message": {
    "id": "uuid",
    "type": "security",
    "title": "...",
    "body": "...",
    "payload": {...},
    "created_at": "2026-06-04T10:30:00Z"
  }
}
```

**Karakteristik queue ini**:
- Best-effort: maxRetry = 0 (tidak perlu retry — data sudah di DB)
- Handler selalu return nil (ACK) — pesan rusak di-drop, tidak di-retry
- Tidak ada DLQ yang perlu dimonitor

---

## 6. Retry & Fallback

### Push Channel Retry (via DeliveryRetryRelay)

```
DeliveryAttempt status = RETRYING
  │
  ▼ [DeliveryRetryRelay polling setiap 5 menit]

RetryPushDeliveryUseCase
  ├─ deviceRepo.FindActiveByUserID()
  ├─ FCM.Send(token, msg)
  └─ Update DeliveryAttempt:
      ├─ SUCCESS → status = success
      ├─ Masih gagal & retryable → ScheduleRetry (5m → 10m → 20m exponential)
      └─ Maks retry (3x) → status = failed
```

### In-App Fallback (jika user offline saat push WS)

```
User offline → WS message dropped dari Hub
→ Notification sudah tersimpan di DB (step 6 flow di atas)
→ User online lagi → GET /notifications → load dari DB
→ Tidak ada data yang hilang
```

---

## 7. User Preference & DND

`NotificationChannelSelector.SelectChannels()` dijalankan per user sebelum dispatch.

```
AllEnabled = false → kembalikan [] (tidak ada channel)
  │
  ▼
DND aktif + dalam window → keluarkan PUSH dari channels
(in_app tetap jalan selama DND)
  │
  ▼
EventChannelPreference per event type → filter per channel toggle
(misal: user matikan push untuk event "social")
  │
  ▼
Interseksi dengan requestedChannels dari template
→ channels final yang akan digunakan
```

**Contoh**: user punya DND 22:00–07:00, login jam 23:00:
- Requested channels: [in_app, push, email]
- Setelah selector: [in_app, email] (push diblok DND)

---

## 8. Event Types yang Terdaftar

| Routing Key | Queue | Target Mode | Channels Default |
|-------------|-------|-------------|-----------------|
| `user.login` | `auth.user.login` | unicast | in_app, push, email |
| `user.registered` | `auth.user.registered` | unicast | — (TODO) |
| `user.verified` | `auth.user.verified` | unicast | — (TODO) |
| `community.post.liked` | `notif.community.post.liked` | unicast | in_app, push |
| `community.post.commented` | `notif.community.post.commented` | unicast | in_app, push |
| `community.comment.replied` | `notif.community.comment.replied` | unicast | in_app, push |
| `community.user.mentioned` | `notif.community.user.mentioned` | unicast | in_app, push |
| `community.post.created` | `notif.community.post.created` | multicast | in_app, push (TODO) |
| `announcement.published` | `notif.announcement.published` | broadcast | in_app, push |
| `region.invitation.email.requested` | `region.invitation.email.requested` | — | email only (direct SMTP) |

---

## 9. Source of Truth untuk Client

| Data | Source | Endpoint |
|------|--------|----------|
| Notifikasi user (unicast/multicast) | `notifications` table | `GET /notifications` |
| Unread count | `notifications` table | `GET /notifications/unread-count` |
| Notifikasi broadcast (announcement) | `announcements` table | `GET /announcements` |
| Real-time delivery | WebSocket Hub (in-memory) | `WS /ws/notifications` |

**Prinsip**: DB adalah sumber kebenaran. WebSocket hanya interupsi real-time — client selalu bisa recover dari DB jika WS terputus.

---

## 10. Cara Menambah Event Trigger Baru

Misal: tambah notifikasi saat user di-follow.

### Langkah 1 — Buat domain event

```go
// internal/domain/user/event/events.go
type UserFollowed struct {
    FollowedUserID string    `json:"followed_user_id"`
    FollowerID     string    `json:"follower_id"`
    FollowerName   string    `json:"follower_name"`
    OccurredAt     time.Time `json:"occurred_at"`
}
```

```go
// internal/domain/user/event/routing.go
const (
    RoutingUserFollowed = "user.followed"
    QueueUserFollowed   = "notif.user.followed"
)
```

### Langkah 2 — Publish event dari use case

```go
// Dalam FollowUserUseCase.Execute()
outboxRepo.Save(ctx, &outbox.OutboxEntry{
    RoutingKey: event.RoutingUserFollowed,
    Payload:    json.Marshal(event.UserFollowed{...}),
})
// Simpan dalam 1 transaksi bersama business logic
```

### Langkah 3 — Tambah MQ handler

```go
// internal/interfaces/mq/handler/handler.go
func (h *MQHandler) HandleUserFollowed(ctx context.Context, msg mq.JobMessage) error {
    var payload event.UserFollowed
    json.Unmarshal(msg.Payload, &payload)

    refType := "user"
    h.dispatcher.Dispatch(ctx, notifservice.NotificationTemplate{
        Type:  notifconst.NotificationTypeSocial,
        Title: "Follower baru",
        Body:  payload.FollowerName + " mengikuti kamu",
        Channels: []notifconst.NotificationChannel{
            notifconst.NotificationChannelInApp,
            notifconst.NotificationChannelPush,
        },
        ReferenceID:   &payload.FollowerID,
        ReferenceType: &refType,
    }, notifservice.Target{
        Mode:        notifservice.TargetModeUnicast,
        RecipientID: payload.FollowedUserID,
    })
    return nil
}
```

### Langkah 4 — Daftarkan handler dan consumer

```go
// RegisterAll()
r.Register(event.RoutingUserFollowed, h.HandleUserFollowed)

// registerConsumers() di worker/main.go
{event.QueueUserFollowed, event.RoutingUserFollowed},
```

Selesai — notifikasi akan berjalan otomatis melewati semua layer (preference check, DB persist, WS, push, email).
