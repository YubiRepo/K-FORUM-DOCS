# Job System Flow — K-Forum API

Dokumen ini menggambarkan arsitektur dan alur kerja job system yang digunakan K-Forum API untuk mengirim domain event secara reliable via RabbitMQ, menggunakan pola **Transactional Outbox + Worker Consumer**.

---

## Daftar Isi

- [Gambaran Umum](#gambaran-umum)
- [Komponen](#komponen)
- [Tiga Konsep Terpisah](#tiga-konsep-terpisah)
- [Producer Flow — Outbox Pattern](#producer-flow--outbox-pattern)
- [Consumer Flow — Worker](#consumer-flow--worker)
- [Job Lifecycle](#job-lifecycle)
- [Error Classification](#error-classification)
- [Retry Policy](#retry-policy)
- [Job Types yang Terdaftar](#job-types-yang-terdaftar)
- [Database Tables](#database-tables)
- [Folder Structure](#folder-structure)
- [Cara Menambah Job Type Baru](#cara-menambah-job-type-baru)

---

## Gambaran Umum

```
┌─────────────────────────────────────────────────────────────────────────┐
│  API Server (cmd/app)                                                   │
│                                                                         │
│  UseCase → save domain event ke event_outbox (satu transaksi DB)        │
└──────────────────────┬──────────────────────────────────────────────────┘
                       │ event_outbox row (PENDING)
                       ▼
┌──────────────────────────────────────────────────────────────────────────┐
│  OutboxRelay (goroutine, polling tiap 2 detik)                           │
│                                                                          │
│  FindPending → Publish ke RabbitMQ → MarkProcessed / MarkFailed          │
└──────────────────────┬───────────────────────────────────────────────────┘
                       │ JobMessage envelope (JSON)
                       ▼
               ┌──────────────┐
               │  RabbitMQ    │
               │  Exchange    │
               └──────┬───────┘
                      │ routing_key → queue
                      ▼
┌─────────────────────────────────────────────────────────────────────────┐
│  Worker (cmd/worker)                                                    │
│                                                                         │
│  RabbitMQConsumer                                                       │
│      ↓                                                                  │
│  JobConsumerMiddleware  ← parse envelope, job lifecycle (DB)            │
│      ↓                                                                  │
│  Registry.Dispatch()   ← route by msg.Type                             │
│      ↓                                                                  │
│  MQHandler.Handle*()   ← business logic (email, notif, dll.)           │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Komponen

| Komponen | Lokasi | Peran |
|---|---|---|
| `OutboxEntry` | `internal/app/service/outbox/` | Representasi event yang perlu dikirim ke queue |
| `OutboxRelay` | `internal/interfaces/mq/relay/` | Goroutine polling outbox → publish ke RabbitMQ |
| `JobMessage` | `internal/interfaces/mq/message.go` | Transport envelope yang dikirim via RabbitMQ |
| `Job` | `internal/app/service/job/job.go` | Tracking eksekusi job di sisi consumer (status, retry) |
| `JobRepository` | `internal/app/service/job/repository.go` | Interface persistence untuk Job |
| `RetryPolicy` | `internal/app/service/job/retry_policy.go` | Konfigurasi max retry per job type |
| `JobConsumerMiddleware` | `internal/interfaces/mq/middleware/` | Intersep message → manage job lifecycle |
| `Registry` | `internal/interfaces/mq/registry/` | Memetakan job type ke handler function |
| `MQHandler` | `internal/interfaces/mq/handler/` | Business logic handler per event type |

---

## Tiga Konsep Terpisah

Penting: ketiga konsep ini **tidak boleh dicampur**.

### 1. Domain Event
Representasi bisnis "sesuatu terjadi".

```go
// internal/domain/user/event/events.go
type UserLoggedIn struct {
    UserID     string    `json:"user_id"`
    IPAddress  string    `json:"ip_address"`
    OccurredAt time.Time `json:"occurred_at"`
}
```

### 2. OutboxEntry
Representasi "event ini perlu dikirim ke queue". Dibuat satu transaksi DB dengan perubahan data.

```go
// internal/app/service/outbox/outbox.go
type OutboxEntry struct {
    ID         uuid.UUID
    RoutingKey string          // e.g. "user.login"
    Payload    json.RawMessage // serialized domain event
    Status     OutboxStatus    // PENDING → PROCESSED / FAILED
    CreatedAt  time.Time
}
```

### 3. JobMessage
Transport envelope yang dikirim ke/dari RabbitMQ.

```go
// internal/interfaces/mq/message.go
type JobMessage struct {
    JobID      string          `json:"job_id"`
    Type       string          `json:"type"`       // sama dengan routing_key
    Version    int             `json:"version"`
    OccurredAt time.Time       `json:"occurred_at"`
    Payload    json.RawMessage `json:"payload"`    // domain event payload
}
```

---

## Producer Flow — Outbox Pattern

Tujuan: **guarantee delivery** — event tidak akan hilang meski RabbitMQ down saat event terjadi.

```
UseCase (API server)
  │
  ├── 1. Lakukan operasi DB utama (simpan user, update status, dll.)
  │
  ├── 2. Serialize domain event → json.RawMessage
  │
  ├── 3. outbox.NewOutboxEntry(routingKey, payload)
  │
  └── 4. outboxRepo.Save(ctx, entry)   ← satu transaksi DB
              │
              │   event_outbox row: PENDING
              ▼
        OutboxRelay (goroutine, interval 2 detik)
              │
              ├── FindPending(limit=50) → ambil semua PENDING
              │
              └── per entry:
                    ├── jobRepo.Save()           ← idempotent (ON CONFLICT DO NOTHING)
                    ├── publisher.Publish()      ← kirim ke RabbitMQ
                    ├── OK  → outboxRepo.MarkProcessed()
                    └── ERR → outboxRepo.MarkFailed(errMsg)
```

**Kenapa pakai Outbox?**

Tanpa outbox, jika API server crash setelah tulis DB tapi sebelum publish ke Rabbit → event hilang.
Dengan outbox, event tersimpan di DB dulu. Relay akan retry publish saat server kembali up.

---

## Consumer Flow — Worker

```
RabbitMQ delivery (raw bytes)
  │
  ▼
JobConsumerMiddleware.Handle(body []byte)
  │
  ├── 1. json.Unmarshal → JobMessage{JobID, Type, Payload, ...}
  │
  ├── 2. jobRepo.FindByID(jobID)
  │       ├── found → pakai job yang ada
  │       └── not found → NewJob() + jobRepo.Save()   ← idempotent
  │
  ├── 3. job.MarkProcessing() + jobRepo.Update()
  │
  ├── 4. Registry.Dispatch(ctx, msg)
  │       └── handler := handlers[msg.Type]
  │               └── handler(ctx, msg)  ← business logic
  │
  └── 5. berdasarkan hasil dispatch:
          ├── nil         → job.MarkCompleted()
          ├── RetryableError → job.MarkRetrying(err) → return err (RabbitMQ akan re-queue)
          └── FatalError     → job.MarkDead(err)     → return err (masuk DLQ)
```

**Desain middleware ini analog dengan HTTP middleware** — handler tidak perlu tahu soal job lifecycle, hanya perlu return error yang benar (`RetryableError` vs `FatalError`).

---

## Job Lifecycle

```
                 ┌─────────┐
                 │ PENDING │  ← dibuat oleh OutboxRelay atau saat consumer terima message
                 └────┬────┘
                      │  MarkProcessing()
                      ▼
               ┌────────────┐
               │ PROCESSING │  ← middleware set sebelum dispatch ke handler
               └─────┬──────┘
                     │
          ┌──────────┼─────────────┐
          │          │             │
          ▼          ▼             ▼
    ┌──────────┐ ┌──────────┐ ┌──────┐
    │COMPLETED │ │ RETRYING │ │ DEAD │
    └──────────┘ └──────────┘ └──────┘
    (sukses)    (error sementara) (error fatal /
                (retry count++)   max retry habis)
```

**Field yang berubah per transisi:**

| Transisi | Status | Field yang diupdate |
|---|---|---|
| `MarkProcessing()` | PROCESSING | `processing_at`, `updated_at` |
| `MarkCompleted()` | COMPLETED | `completed_at`, `updated_at` |
| `MarkRetrying(err)` | RETRYING | `retry_count++`, `last_error`, `updated_at` |
| `MarkDead(err)` | DEAD | `last_error`, `updated_at` |

**`CanProcess()`** — cek apakah job boleh diproses:
```go
func (j *Job) CanProcess() bool {
    return j.Status == PENDING || j.Status == RETRYING
}
```

---

## Error Classification

Handler wajib membungkus error dengan tipe yang benar agar middleware tahu apa yang harus dilakukan.

### RetryableError — error sementara

Digunakan untuk kegagalan yang bisa hilang sendiri (network timeout, service down, DB busy).

```go
// RabbitMQ akan re-queue message → job masuk RETRYING
return job.NewRetryableError("gagal kirim login notification email", err)
```

Contoh kasus:
- SMTP timeout
- HTTP 5xx dari external service
- DB connection refused

### FatalError — error permanen

Digunakan untuk kegagalan yang tidak akan berubah meski diulang.

```go
// Job masuk DEAD langsung → tidak ada retry
return job.NewFatalError("gagal unmarshal UserLoggedIn payload", err)
```

Contoh kasus:
- Payload JSON corrupt / schema tidak cocok
- Job type tidak ada di registry
- Data bisnis invalid yang tidak bisa diperbaiki

### Kombinasi keduanya (contoh nyata)

```go
func (h *MQHandler) HandleUserLoggedIn(ctx context.Context, msg mq.JobMessage) error {
    var payload event.UserLoggedIn
    if err := json.Unmarshal(msg.Payload, &payload); err != nil {
        // Payload rusak → tidak ada gunanya retry
        return job.NewFatalError("gagal unmarshal UserLoggedIn payload", err)
    }

    if err := h.emailSender.SendLoginNotification(...); err != nil {
        // Email server bisa down sementara → retry masuk akal
        return job.NewRetryableError("gagal kirim login notification email", err)
    }

    return nil
}
```

---

## Retry Policy

Dikonfigurasi di `cmd/worker/main.go`:

```go
retryPolicy := job.NewRetryPolicy(3)  // default max 3 retry untuk semua job type
```

Untuk override per job type:
```go
retryPolicy.Register("user.login", 5)           // login notif: 5x retry
retryPolicy.Register("region.invitation.*", 2)  // undangan: 2x retry
```

`MaxRetryFor(jobType)` → fallback ke `defaultMax` jika tidak ada override spesifik.

> **Catatan**: Saat ini semua job type menggunakan default max retry = 3. Job yang gagal setelah 3x retry masuk status DEAD dan tidak diproses ulang secara otomatis — perlu manual review atau DLQ monitoring.

---

## Job Types yang Terdaftar

Semua handler didaftarkan di `internal/interfaces/mq/handler/auth.go` via `RegisterAll()`.

| Routing Key | Queue | Handler | Aksi |
|---|---|---|---|
| `user.registered` | `auth.user.registered` | `HandleUserRegistered` | TODO: welcome email, init preferences |
| `user.verified` | `auth.user.verified` | `HandleUserVerified` | TODO: post-verification notif |
| `user.login` | `auth.user.login` | `HandleUserLoggedIn` | Kirim email + push notif login |
| `region.invitation.email.requested` | `region.invitation.email.requested` | `HandleRegionInvitationEmailRequested` | Kirim email undangan region |

Routing key = nama event bisnis. Queue = consumer group yang memproses event tersebut.

---

## Database Tables

### `event_outbox`

Menyimpan event yang antri untuk dikirim ke RabbitMQ (producer side).

```sql
CREATE TABLE event_outbox (
    id           UUID PRIMARY KEY,
    routing_key  TEXT NOT NULL,       -- e.g. "user.login"
    payload      JSONB NOT NULL,      -- serialized domain event
    status       TEXT NOT NULL,       -- PENDING | PROCESSED | FAILED
    created_at   TIMESTAMP NOT NULL,
    processed_at TIMESTAMP,           -- diisi saat berhasil publish
    error_msg    TEXT                 -- diisi saat gagal publish
);
```

**Status lifecycle:**
```
PENDING → PROCESSED  (berhasil publish ke RabbitMQ)
PENDING → FAILED     (gagal publish, perlu investigasi)
```

### `jobs`

Menyimpan tracking eksekusi setiap job di sisi consumer.

```sql
CREATE TABLE jobs (
    id            UUID PRIMARY KEY,
    type          TEXT NOT NULL,       -- routing_key / job type
    payload       JSONB NOT NULL,
    status        TEXT NOT NULL,       -- PENDING | PROCESSING | COMPLETED | RETRYING | DEAD
    retry_count   INT DEFAULT 0,
    max_retry     INT DEFAULT 3,
    scheduled_at  TIMESTAMP NOT NULL,
    processing_at TIMESTAMP,
    completed_at  TIMESTAMP,
    last_error    TEXT,
    created_at    TIMESTAMP NOT NULL,
    updated_at    TIMESTAMP NOT NULL
);
```

**Idempotency**: `Save()` menggunakan `ON CONFLICT (id) DO NOTHING` — aman jika Relay mengirim duplikat.

---

## Folder Structure

```
internal/
├── app/
│   └── service/
│       ├── job/
│       │   ├── job.go          ← Job struct + MarkProcessing/Completed/Retrying/Dead
│       │   ├── error.go        ← RetryableError, FatalError, IsRetryable(), IsFatal()
│       │   ├── retry_policy.go ← RetryPolicy: per-type max retry config
│       │   └── repository.go   ← JobRepository interface
│       └── outbox/
│           ├── outbox.go       ← OutboxEntry struct + NewOutboxEntry()
│           └── repository.go   ← OutboxRepository interface
│
├── interfaces/
│   └── mq/
│       ├── message.go          ← JobMessage (transport envelope)
│       ├── handler/
│       │   ├── auth.go         ← MQHandler + handlers untuk user events
│       │   └── region_invitation.go ← handler untuk region invitation event
│       ├── middleware/
│       │   └── job_lifecycle.go ← JobConsumerMiddleware: parse → lifecycle → dispatch
│       ├── registry/
│       │   └── registry.go     ← Registry: Register() + Dispatch()
│       └── relay/
│           └── outbox_relay.go ← OutboxRelay: polling outbox → publish RabbitMQ
│
└── infrastructure/
    └── persistence/
        ├── postgres_job_repository.go    ← implementasi JobRepository
        └── postgres_outbox_repository.go ← implementasi OutboxRepository

cmd/
└── worker/
    └── main.go    ← entry point: wiring semua komponen + registerConsumers()
```

---

## Cara Menambah Job Type Baru

Contoh: menambah handler untuk event `community.member_joined`.

### Langkah 1 — Definisikan routing & queue constants

```go
// internal/domain/community/event/routing.go
const (
    RoutingCommunityMemberJoined = "community.member_joined"
    QueueCommunityMemberJoined   = "community.member_joined"
)
```

### Langkah 2 — Definisikan event struct

```go
// internal/domain/community/event/events.go
type CommunityMemberJoined struct {
    CommunityID string    `json:"community_id"`
    UserID      string    `json:"user_id"`
    JoinedAt    time.Time `json:"joined_at"`
}
```

### Langkah 3 — Tulis handler

```go
// internal/interfaces/mq/handler/community.go

func (h *MQHandler) HandleCommunityMemberJoined(ctx context.Context, msg mq.JobMessage) error {
    var payload communityevent.CommunityMemberJoined
    if err := json.Unmarshal(msg.Payload, &payload); err != nil {
        return job.NewFatalError("gagal unmarshal CommunityMemberJoined payload", err)
    }

    // ... business logic
    if err := h.notifUseCases.NotifyCommunityMemberJoined.Execute(ctx, payload); err != nil {
        return job.NewRetryableError("gagal kirim notifikasi member joined", err)
    }

    return nil
}
```

### Langkah 4 — Daftarkan ke registry

```go
// internal/interfaces/mq/handler/auth.go → RegisterAll()

func RegisterAll(r *registry.Registry, h *MQHandler) {
    // ... existing ...
    r.Register(communityevent.RoutingCommunityMemberJoined, h.HandleCommunityMemberJoined)
}
```

### Langkah 5 — Tambah consumer di worker/main.go

```go
// cmd/worker/main.go → registerConsumers()

if err := consumer.Consume(
    communityevent.QueueCommunityMemberJoined,
    communityevent.RoutingCommunityMemberJoined,
    retryPolicy.MaxRetryFor(communityevent.RoutingCommunityMemberJoined),
    middleware.Handle,
); err != nil {
    return fmt.Errorf("gagal consume %s: %w", communityevent.RoutingCommunityMemberJoined, err)
}
```

### Langkah 6 — Publish event dari UseCase (producer side)

```go
// Di usecase yang memicu event

payload, _ := json.Marshal(communityevent.CommunityMemberJoined{
    CommunityID: communityID,
    UserID:      userID,
    JoinedAt:    time.Now(),
})
entry := outbox.NewOutboxEntry(communityevent.RoutingCommunityMemberJoined, payload)
outboxRepo.Save(ctx, entry)  // dalam satu transaksi DB dengan operasi utama
```

---

## Ringkasan Pola Kunci

| Pola | Implementasi |
|---|---|
| **Reliable delivery** | Transactional Outbox — event disimpan DB sebelum dikirim ke Rabbit |
| **Idempotency** | `ON CONFLICT DO NOTHING` di job save; consumer harus handle duplikat |
| **At-least-once** | RabbitMQ default; handler harus idempotent |
| **Error classification** | `RetryableError` vs `FatalError` — handler yang memutuskan |
| **Observability** | Status job di tabel `jobs` + `last_error` untuk debugging |
| **Extensibility** | Registry pattern — tambah job type tanpa ubah infrastruktur |
