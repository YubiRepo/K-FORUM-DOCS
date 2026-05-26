# Domain Breakdown — Notification

Dokumen ini menerjemahkan brainstorming notification system ke struktur domain konkret yang mengikuti arsitektur DDD proyek ini. Rujukan arsitektur: `CONTEXT.md`.

---

## Daftar Isi

- [Bounded Context](#bounded-context)
- [Folder Structure](#folder-structure)
- [Aggregates & Entities](#aggregates--entities)
- [Value Objects](#value-objects)
- [Constants & Error Codes](#constants--error-codes)
- [Repository Interfaces](#repository-interfaces)
- [Domain Service](#domain-service)
- [Application Port (Messaging)](#application-port-messaging)
- [Layer Boundary](#layer-boundary)
- [Event & Queue Topology](#event--queue-topology)
- [Catatan Penting](#catatan-penting)

---

## Bounded Context

Notification adalah bounded context tersendiri. Dia **consumer** dari domain lain (Post, Comment, Event, dll.), bukan bagian dari domain tersebut.

```
internal/domain/notification/
```

Aturan batas:

- Domain notification tidak boleh import domain lain.
- Domain lain tidak boleh import domain notification secara langsung.
- Komunikasi antar domain dilakukan via domain event melalui message queue (RabbitMQ).

---

## Folder Structure

Mengikuti pola bounded context yang sudah ada (contoh: `subscription`):

```
internal/domain/notification/
├── constant/
│   ├── notification_constant.go      # NotificationStatus, NotificationType, NotificationChannel
│   ├── device_constant.go            # Platform, PushProvider
│   └── error_code.go                 # Semua domain error code
├── entity/
│   ├── notification.go               # Aggregate root
│   ├── delivery_attempt.go           # Entity dalam aggregate Notification
│   ├── device_registration.go        # Aggregate root tersendiri
│   └── notification_preference.go    # Aggregate root tersendiri
├── valueobject/
│   └── notification_vo.go            # NotificationPayload, DoNotDisturbWindow
└── repository/
    └── interfaces.go                 # Semua interface repository notification
```

> **Catatan file `port`**: Query model (listing/pagination) untuk notification diletakkan di `internal/app/port/notification_query_model.go`, bukan di domain. Domain hanya berisi interface repository dan aggregate.

---

## Aggregates & Entities

### 1. `Notification` — Aggregate Root

Source of truth untuk setiap notifikasi yang dikirim ke seorang user.

**Fields:**

| Field           | Type                    | Keterangan                                                        |
| --------------- | ----------------------- | ----------------------------------------------------------------- |
| `ID`            | `string`                | UUID                                                              |
| `RecipientID`   | `string`                | User yang menerima notif                                          |
| `Type`          | `NotificationType`      | Jenis notif (SYSTEM, SOCIAL, CONTENT, dll.)                       |
| `Title`         | `string`                | Judul notif                                                       |
| `Body`          | `string`                | Isi notif                                                         |
| `Payload`       | `NotificationPayload`   | Value object: data kontekstual tambahan (deep link, reference id) |
| `Status`        | `NotificationStatus`    | PENDING / QUEUED / SENT / FAILED / READ                           |
| `Channels`      | `[]NotificationChannel` | IN_APP, PUSH, EMAIL, SMS                                          |
| `ReferenceID`   | `*string`               | ID entitas sumber (post_id, comment_id, dll.)                     |
| `ReferenceType` | `*string`               | Tipe entitas sumber ("post", "comment", dll.)                     |
| `CreatedAt`     | `time.Time`             |                                                                   |
| `ReadAt`        | `*time.Time`            | Nil jika belum dibaca                                             |
| `SentAt`        | `*time.Time`            | Nil jika belum terkirim                                           |

**Domain Methods:**

- `NewNotification(...)` — constructor dengan validasi.
- `MarkAsRead()` — set status READ, isi ReadAt.
- `MarkAsSent()` — set status SENT, isi SentAt.
- `MarkAsFailed()` — set status FAILED.
- `Enqueue()` — transisi dari PENDING ke QUEUED.

---

### 2. `DeliveryAttempt` — Entity (dalam aggregate Notification)

Mencatat setiap percobaan pengiriman notif per channel. Digunakan untuk retry logic dan audit.

**Fields:**

| Field            | Type                  | Keterangan                          |
| ---------------- | --------------------- | ----------------------------------- |
| `ID`             | `string`              | UUID                                |
| `NotificationID` | `string`              | FK ke Notification                  |
| `Channel`        | `NotificationChannel` | PUSH / EMAIL / SMS                  |
| `Status`         | `DeliveryStatus`      | SUCCESS / FAILED / RETRYING         |
| `ProviderCode`   | `*string`             | Kode error dari provider (opsional) |
| `RetryCount`     | `int`                 |                                     |
| `NextRetryAt`    | `*time.Time`          |                                     |
| `AttemptedAt`    | `time.Time`           |                                     |

**Domain Methods:**

- `NewDeliveryAttempt(...)` — constructor.
- `MarkSuccess()` — set status SUCCESS.
- `MarkFailed(providerCode string)` — set status FAILED, simpan kode error.
- `ScheduleRetry(at time.Time)` — set status RETRYING, isi NextRetryAt, increment RetryCount.
- `IsRetryable() bool` — cek apakah masih bisa diretry (contoh: RetryCount < maxRetry).

---

### 3. `DeviceRegistration` — Aggregate Root

Representasi device user yang terdaftar untuk menerima push notification.

> Nama field menggunakan `PushProvider` dan `ProviderToken`, **bukan** `FCMToken`. Ini agar bisa support FCM, APNS, Huawei, WebPush.

**Fields:**

| Field           | Type           | Keterangan                         |
| --------------- | -------------- | ---------------------------------- |
| `ID`            | `string`       | UUID                               |
| `UserID`        | `string`       | Pemilik device                     |
| `Platform`      | `Platform`     | ANDROID / IOS / WEB                |
| `PushProvider`  | `PushProvider` | FCM / APNS / HUAWEI / WEB_PUSH     |
| `ProviderToken` | `string`       | Token dari push provider           |
| `Active`        | `bool`         | false jika token sudah tidak valid |
| `LastSeenAt`    | `time.Time`    |                                    |
| `CreatedAt`     | `time.Time`    |                                    |
| `UpdatedAt`     | `time.Time`    |                                    |

**Domain Methods:**

- `NewDeviceRegistration(...)` — constructor dengan validasi.
- `UpdateToken(newToken string)` — ganti token, update LastSeenAt.
- `Deactivate()` — set Active = false (dipanggil saat provider return UNREGISTERED).
- `RecordSeen()` — update LastSeenAt.

---

### 4. `NotificationPreference` — Aggregate Root

Setting notifikasi milik satu user. Menentukan channel mana yang aktif untuk setiap jenis event.

**Fields:**

| Field                | Type                       | Keterangan                     |
| -------------------- | -------------------------- | ------------------------------ |
| `ID`                 | `string`                   | UUID                           |
| `UserID`             | `string`                   | Unique per user                |
| `AllEnabled`         | `bool`                     | Global toggle                  |
| `DoNotDisturb`       | `bool`                     | Mode senyap                    |
| `DoNotDisturbWindow` | `*DoNotDisturbWindow`      | Value object: start & end time |
| `Preferences`        | `[]EventChannelPreference` | Per event type per channel     |
| `CreatedAt`          | `time.Time`                |                                |
| `UpdatedAt`          | `time.Time`                |                                |

**Embedded type `EventChannelPreference`:**

| Field       | Type                  | Keterangan                       |
| ----------- | --------------------- | -------------------------------- |
| `EventType` | `string`              | e.g. "comment", "event_reminder" |
| `Channel`   | `NotificationChannel` | PUSH / EMAIL / SMS / IN_APP      |
| `Enabled`   | `bool`                |                                  |

**Domain Methods:**

- `NewNotificationPreference(userID string)` — constructor dengan default semua enabled.
- `IsChannelEnabledFor(eventType string, channel NotificationChannel) bool` — cek apakah channel aktif untuk event tertentu, dengan mempertimbangkan global toggle dan DND.
- `SetPreference(eventType string, channel NotificationChannel, enabled bool)` — update satu preference.
- `EnableAll()` / `DisableAll()` — toggle global.
- `SetDoNotDisturb(window DoNotDisturbWindow)` / `ClearDoNotDisturb()`.

---

## Value Objects

### `NotificationPayload`

Data kontekstual yang dibawa oleh notif. Immutable.

```go
// internal/domain/notification/valueobject/notification_vo.go

type NotificationPayload struct {
    DeepLink    string            // URL deep link ke konten
    ImageURL    string            // Opsional, untuk rich push
    Extra       map[string]string // Data tambahan (key-value)
}
```

### `DoNotDisturbWindow`

Rentang waktu DND. Immutable.

```go
type DoNotDisturbWindow struct {
    StartTime string // Format "HH:MM" (e.g. "22:00")
    EndTime   string // Format "HH:MM" (e.g. "07:00")
}

func NewDoNotDisturbWindow(start, end string) (DoNotDisturbWindow, error) // validasi format
func (w DoNotDisturbWindow) IsActive(now time.Time) bool
```

---

## Constants & Error Codes

### `constant/notification_constant.go`

```go
type NotificationStatus string

const (
    NotificationStatusPending NotificationStatus = "pending"
    NotificationStatusQueued  NotificationStatus = "queued"
    NotificationStatusSent    NotificationStatus = "sent"
    NotificationStatusFailed  NotificationStatus = "failed"
    NotificationStatusRead    NotificationStatus = "read"
)

type NotificationType string

const (
    NotificationTypeSystem    NotificationType = "system"
    NotificationTypeSocial    NotificationType = "social"
    NotificationTypeContent   NotificationType = "content"
    NotificationTypeReminder  NotificationType = "reminder"
    NotificationTypeMarketing NotificationType = "marketing"
    NotificationTypeSecurity  NotificationType = "security"
)

type NotificationChannel string

const (
    NotificationChannelInApp NotificationChannel = "in_app"
    NotificationChannelPush  NotificationChannel = "push"
    NotificationChannelEmail NotificationChannel = "email"
    NotificationChannelSMS   NotificationChannel = "sms"
)

type DeliveryStatus string

const (
    DeliveryStatusSuccess  DeliveryStatus = "success"
    DeliveryStatusFailed   DeliveryStatus = "failed"
    DeliveryStatusRetrying DeliveryStatus = "retrying"
)
```

### `constant/device_constant.go`

```go
type Platform string

const (
    PlatformAndroid Platform = "android"
    PlatformIOS     Platform = "ios"
    PlatformWeb     Platform = "web"
)

type PushProvider string

const (
    PushProviderFCM     PushProvider = "fcm"
    PushProviderAPNS    PushProvider = "apns"
    PushProviderHuawei  PushProvider = "huawei"
    PushProviderWebPush PushProvider = "web_push"
)
```

### `constant/error_code.go`

```go
const (
    // Notification
    CodeNotificationIDRequired        domainerr.Code = "NOTIF_ID_REQUIRED"
    CodeNotificationRecipientRequired domainerr.Code = "NOTIF_RECIPIENT_REQUIRED"
    CodeNotificationTitleRequired     domainerr.Code = "NOTIF_TITLE_REQUIRED"
    CodeNotificationBodyRequired      domainerr.Code = "NOTIF_BODY_REQUIRED"
    CodeNotificationInvalidStatus     domainerr.Code = "NOTIF_INVALID_STATUS"
    CodeNotificationNotFound          domainerr.Code = "NOTIF_NOT_FOUND"
    CodeNotificationPersistenceFailed domainerr.Code = "NOTIF_PERSISTENCE_FAILED"
    CodeNotificationQueryFailed       domainerr.Code = "NOTIF_QUERY_FAILED"

    // DeliveryAttempt
    CodeDeliveryAttemptIDRequired         domainerr.Code = "DELIVERY_ID_REQUIRED"
    CodeDeliveryAttemptNotFound           domainerr.Code = "DELIVERY_NOT_FOUND"
    CodeDeliveryAttemptPersistenceFailed  domainerr.Code = "DELIVERY_PERSISTENCE_FAILED"

    // DeviceRegistration
    CodeDeviceIDRequired             domainerr.Code = "DEVICE_ID_REQUIRED"
    CodeDeviceUserIDRequired         domainerr.Code = "DEVICE_USER_ID_REQUIRED"
    CodeDeviceTokenRequired          domainerr.Code = "DEVICE_TOKEN_REQUIRED"
    CodeDeviceInvalidPlatform        domainerr.Code = "DEVICE_INVALID_PLATFORM"
    CodeDeviceInvalidProvider        domainerr.Code = "DEVICE_INVALID_PROVIDER"
    CodeDeviceNotFound               domainerr.Code = "DEVICE_NOT_FOUND"
    CodeDeviceDuplicateToken         domainerr.Code = "DEVICE_DUPLICATE_TOKEN"
    CodeDevicePersistenceFailed      domainerr.Code = "DEVICE_PERSISTENCE_FAILED"
    CodeDeviceQueryFailed            domainerr.Code = "DEVICE_QUERY_FAILED"

    // NotificationPreference
    CodePreferenceUserIDRequired     domainerr.Code = "PREF_USER_ID_REQUIRED"
    CodePreferenceInvalidChannel     domainerr.Code = "PREF_INVALID_CHANNEL"
    CodePreferenceNotFound           domainerr.Code = "PREF_NOT_FOUND"
    CodePreferencePersistenceFailed  domainerr.Code = "PREF_PERSISTENCE_FAILED"

    // DoNotDisturbWindow
    CodeDNDInvalidTimeFormat         domainerr.Code = "DND_INVALID_TIME_FORMAT"
)
```

---

## Repository Interfaces

`internal/domain/notification/repository/interfaces.go`

```go
package repository

import (
    "context"
    "k-forum-api/internal/domain/notification/entity"
)

// NotificationRepository menyimpan dan mengambil aggregate Notification.
type NotificationRepository interface {
    Save(ctx context.Context, n *entity.Notification) error
    Update(ctx context.Context, n *entity.Notification) error
    FindByID(ctx context.Context, id string) (*entity.Notification, error)
    FindUnreadByRecipient(ctx context.Context, recipientID string) ([]*entity.Notification, error)
    CountUnreadByRecipient(ctx context.Context, recipientID string) (int64, error)
}

// DeliveryAttemptRepository menyimpan dan mengambil DeliveryAttempt.
type DeliveryAttemptRepository interface {
    Save(ctx context.Context, da *entity.DeliveryAttempt) error
    Update(ctx context.Context, da *entity.DeliveryAttempt) error
    FindByID(ctx context.Context, id string) (*entity.DeliveryAttempt, error)
    FindPendingRetries(ctx context.Context) ([]*entity.DeliveryAttempt, error)
}

// DeviceRegistrationRepository menyimpan dan mengambil aggregate DeviceRegistration.
type DeviceRegistrationRepository interface {
    Save(ctx context.Context, dr *entity.DeviceRegistration) error
    Update(ctx context.Context, dr *entity.DeviceRegistration) error
    FindByID(ctx context.Context, id string) (*entity.DeviceRegistration, error)
    FindByUserIDAndToken(ctx context.Context, userID, token string) (*entity.DeviceRegistration, error)
    FindActiveByUserID(ctx context.Context, userID string) ([]*entity.DeviceRegistration, error)
    DeactivateByToken(ctx context.Context, token string) error
}

// NotificationPreferenceRepository menyimpan dan mengambil aggregate NotificationPreference.
type NotificationPreferenceRepository interface {
    Save(ctx context.Context, pref *entity.NotificationPreference) error
    Update(ctx context.Context, pref *entity.NotificationPreference) error
    FindByUserID(ctx context.Context, userID string) (*entity.NotificationPreference, error)
}
```

> **Query listing** (pagination notif history, filter per type/channel) diletakkan di `internal/app/port/notification_query_model.go` sebagai `NotificationQueryModel` interface terpisah, bukan di sini.

---

## Domain Service

Digunakan jika aturan bisnis melibatkan lebih dari satu aggregate atau tidak cocok ditempatkan di salah satu entity.

### `NotificationChannelSelector`

Menentukan channel apa saja yang harus digunakan untuk mengirim notif berdasarkan preference user dan jenis event.

**Lokasi:** `internal/domain/notification/` (file: `channel_selector.go`)

```go
// Kontrak
type NotificationChannelSelector interface {
    SelectChannels(
        preference *entity.NotificationPreference,
        eventType string,
        requestedChannels []constant.NotificationChannel,
    ) []constant.NotificationChannel
}
```

**Rule bisnis yang dikandung:**

- Jika `AllEnabled = false`, kembalikan slice kosong.
- Jika DND aktif, keluarkan channel PUSH dari hasil (in_app tetap jalan).
- Filter berdasarkan `EventChannelPreference` per event type.
- Interseksi antara `requestedChannels` dan channel yang diizinkan user.

---

## Application Port (Messaging)

Ini bukan domain, tapi **kontrak** untuk infrastruktur messaging. Diletakkan di `internal/app/port/`.

### `PushSender` — port untuk push notification

```go
// internal/app/port/notification_port.go

type PushMessage struct {
    Token    string
    Title    string
    Body     string
    Payload  map[string]string
    Priority string // "high" | "normal"
}

// PushSender adalah kontrak untuk kirim push notification.
// Implementasi konkret (FCM, APNS) ada di infrastructure/external.
type PushSender interface {
    Send(ctx context.Context, msg PushMessage) error
    SendBatch(ctx context.Context, msgs []PushMessage) error
}
```

### `NotificationEventPublisher` — port untuk publish domain event ke queue

```go
// internal/app/port/notification_port.go (lanjutan)

type DeliveryJob struct {
    NotificationID string
    Channel        string // "push" | "email" | "sms"
}

// NotificationEventPublisher digunakan usecase untuk enqueue delivery job.
type NotificationEventPublisher interface {
    EnqueueDeliveryJob(ctx context.Context, job DeliveryJob) error
}
```

**Implementasi** ada di `internal/infrastructure/external/` (RabbitMQ publisher). Usecase hanya bergantung pada interface ini.

---

## Layer Boundary

### Yang BOLEH ada di domain notification:

```
entity.Notification
entity.DeliveryAttempt
entity.DeviceRegistration
entity.NotificationPreference
valueobject.NotificationPayload
valueobject.DoNotDisturbWindow
constant.NotificationStatus
constant.NotificationChannel
constant.NotificationType
constant.Platform
constant.PushProvider
constant.Error codes
repository.NotificationRepository (interface)
repository.DeviceRegistrationRepository (interface)
repository.NotificationPreferenceRepository (interface)
domain service: NotificationChannelSelector
```

### Yang TIDAK BOLEH masuk domain (pindahkan ke port/infrastructure):

```
FCM SDK
RabbitMQ client
HTTP client
Konfigurasi token / credential
Logika retry scheduling (ini usecase concern)
"SendFCMNotification" (terlalu infrastructure-centric)
```

### Usecase yang tepat (application layer):

```go
// BENAR — bicara bahasa domain/bisnis
NotifyPostPublishedFollowersUseCase
NotifyUserMentionedUseCase
RegisterDeviceTokenUseCase
UpdateNotificationPreferenceUseCase
MarkNotificationReadUseCase
ProcessPushDeliveryUseCase    // worker: ambil notif, cari token, call PushSender

// SALAH — terlalu infrastructure-centric
SendFCMNotificationUseCase
PushToFCMUseCase
```

---

## Event & Queue Topology

Sesuai brainstorming, queue yang dibutuhkan:

```
events.topic                        (domain events dari service lain)
 ├── post.published
 ├── comment.created
 ├── event.reminder_due
 └── user.followed

notification.push.high              (transactional: mention, direct interaction)
notification.push.normal            (fanout: post baru, event baru)
notification.email
notification.push.fanout_batch      (massive recipients, pakai FCM multicast)
notification.retry                  (retry queue dengan delay)
notification.dlq                    (dead letter: gagal semua retry)
```

**Worker / Consumer mapping:**

| Queue                            | Consumer                    | Keterangan                                     |
| -------------------------------- | --------------------------- | ---------------------------------------------- |
| `events.topic`                   | `NotificationEventConsumer` | Create notification records + enqueue delivery |
| `notification.push.high`         | `PushDeliveryWorker`        | Low latency, concurrency kecil                 |
| `notification.push.normal`       | `PushDeliveryWorker`        | Throughput lebih tinggi                        |
| `notification.push.fanout_batch` | `FanoutBatchWorker`         | FCM multicast, proses per batch                |
| `notification.retry`             | `RetryDeliveryWorker`       | Delay-based retry                              |
| `notification.dlq`               | Manual review / alerting    |                                                |

**Lokasi consumer/worker:** `internal/interfaces/mq/` (consumer handler) dan `cmd/worker/main.go` sebagai entry point.

---

## Catatan Penting

1. **Nama field `ProviderToken`, bukan `FCMToken`** — agar tidak coupling ke FCM. Domain tidak tahu siapa provider-nya.

2. **`NotificationEventConsumer` adalah consumer, bukan publisher** — dia terima event dari Post/Comment/Event service, lalu membuat `Notification` record dan enqueue delivery job.

3. **Dua langkah terpisah: create notif & deliver notif** — ini disengaja. Create notif adalah application concern, deliver notif adalah infrastructure concern yang dijembatani lewat port `PushSender`.

4. **Invalid token (UNREGISTERED dari FCM)** → panggil `DeviceRegistration.Deactivate()` via repository, bukan hapus record. Ini penting untuk audit.

5. **Error mapping di persistence** mengikuti aturan CONTEXT.md:
   - `sql.ErrNoRows` → `domainerr.New(constant.CodeXxxNotFound, "...")`
   - Unique constraint → `domainerr.New(constant.CodeDeviceDuplicateToken, "...")`
   - Gagal simpan → `domainerr.Wrap(constant.CodeXxxPersistenceFailed, "...", err)`

6. **Query listing notif** (history, pagination, filter by type) → `port.NotificationQueryModel`, bukan domain repository. Ikuti aturan CONTEXT.md section 6.
