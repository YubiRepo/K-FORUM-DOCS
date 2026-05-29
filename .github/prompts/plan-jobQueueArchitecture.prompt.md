# Plan: Job Queue Architecture Foundation

**TL;DR** — Extend RabbitMQ setup yang ada menjadi production-grade job system dengan 3 pilar: (1) **Job entity + DB lifecycle**, (2) **Outbox Pattern** untuk atomic publish, (3) **DLQ + retry classification** di consumer. Arsitektur ini mengikuti DDD conventions dari CONTEXT.md dan staging approach dari JOB_DOMAIN.md.

---

## Phase 1 — Domain Layer

### 1.1 Job Domain — `internal/domain/job/`

| File                           | Isi                                                                                                                                                                                                        |
| ------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `entity/job.go`                | Job aggregate: ID, Type, Payload, Status, RetryCount, MaxRetry, ProcessingAt, CompletedAt, LastError. Methods: `MarkProcessing()`, `MarkCompleted()`, `MarkRetrying(err)`, `MarkDead(err)`, `CanProcess()` |
| `valueobject/job_status.go`    | const: PENDING, PROCESSING, COMPLETED, RETRYING, DEAD                                                                                                                                                      |
| `repository/job_repository.go` | Interface: `Save`, `FindByID`, `Update`                                                                                                                                                                    |
| `service/retry_policy.go`      | `MaxRetryFor(jobType) int` (default=3, per-type configurable), `IsRetryable(count, max) bool`                                                                                                              |
| `message.go`                   | **JobMessage envelope** — `JobID`, `Type`, `Version int`, `OccurredAt`, `Payload json.RawMessage`                                                                                                          |
| `error.go`                     | `RetryableError`, `FatalError` types + `IsRetryable(err) bool` helper                                                                                                                                      |

### 1.2 Outbox Domain — `internal/domain/outbox/`

| File                              | Isi                                                                                                       |
| --------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `entity/outbox_entry.go`          | OutboxEntry: ID, RoutingKey, Payload, Status (PENDING/PROCESSED/FAILED), CreatedAt, ProcessedAt, ErrorMsg |
| `repository/outbox_repository.go` | Interface: `Save`, `FindPending(ctx, limit)`, `MarkProcessed`, `MarkFailed`                               |

---

## Phase 2 — Migrations

**`internal/migrations/0004_create_jobs_table.up.sql`**

```sql
CREATE TABLE jobs (
  id UUID PRIMARY KEY,
  type VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
  retry_count INT NOT NULL DEFAULT 0,
  max_retry INT NOT NULL DEFAULT 3,
  scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  processing_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  last_error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_jobs_status_scheduled ON jobs(status, scheduled_at);
```

**`internal/migrations/0004_create_jobs_table.down.sql`**

```sql
DROP TABLE jobs;
```

**`internal/migrations/0005_create_event_outbox_table.up.sql`**

```sql
CREATE TABLE event_outbox (
  id UUID PRIMARY KEY,
  routing_key VARCHAR(100) NOT NULL,
  payload JSONB NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  processed_at TIMESTAMPTZ,
  error_msg TEXT
);
CREATE INDEX idx_outbox_status_created ON event_outbox(status, created_at);
```

**`internal/migrations/0005_create_event_outbox_table.down.sql`**

```sql
DROP TABLE event_outbox;
```

---

## Phase 3 — Persistence Layer

### 3.1 Transactor Pattern — `internal/infrastructure/persistence/transactor.go`

- `PostgresTransactor{db *sql.DB}` implements `port.Transactor`
- `WithTx(ctx, fn)`: begin tx → `context.WithValue(ctx, txKey{}, tx)` → fn(txCtx) → commit/rollback
- Export `TxFromContext(ctx) *sql.Tx` — dipakai oleh repo yang perlu ikut TX
- Semua repo yang perlu TX-aware: panggil `execerFromContext(ctx)` yang return `*sql.Tx` jika ada, else `*sql.DB`

### 3.2 Repository Implementations

**`internal/infrastructure/persistence/postgres_job_repository.go`**

- `PostgresJobRepository{db *sql.DB}`
- `Save`: INSERT INTO jobs
- `FindByID`: SELECT; `sql.ErrNoRows` → `domainerr.New(CodeJobNotFound, ...)`
- `Update`: UPDATE jobs SET status, retry_count, processing_at, completed_at, last_error, updated_at WHERE id
- Error mapping per CONTEXT.md §7

**`internal/infrastructure/persistence/postgres_outbox_repository.go`**

- `PostgresOutboxRepository{db *sql.DB}`
- `Save`: INSERT INTO event_outbox — TX-aware via `TxFromContext`
- `FindPending`: `SELECT ... WHERE status='PENDING' ORDER BY created_at LIMIT $1 FOR UPDATE SKIP LOCKED`
- `MarkProcessed`: UPDATE status='PROCESSED', processed_at=NOW() WHERE id=$1
- `MarkFailed`: UPDATE status='FAILED', error_msg=$2 WHERE id=$1

### 3.3 Domain Error Constants

Tambah di package constant job/outbox masing-masing:

- `CodeJobNotFound`, `CodeJobPersistenceFailed`, `CodeJobQueryFailed`
- `CodeOutboxPersistenceFailed`, `CodeOutboxQueryFailed`

---

## Phase 4 — Transport Layer

### Update `internal/infrastructure/external/rabbitmq/client.go`

**Consumer changes:**

- Set QoS di init: `ch.Qos(10, 0, false)` (prefetch=10)
- Declare DLQ exchange `k-forum.dlq` (type: `direct`) di `NewRabbitMQConsumer` dan `NewRabbitMQPublisher`
- Update `Consume()` signature: tambah `maxRetry int` param
- `QueueDeclare` tambah args:
  ```go
  amqp.Table{
    "x-dead-letter-exchange":    "k-forum.dlq",
    "x-dead-letter-routing-key": queueName + ".dlq",
  }
  ```
- Declare DLQ queue `{queueName}.dlq` — bound ke exchange `k-forum.dlq` dengan routing key `{queueName}.dlq`
- Consumer goroutine logic baru:
  ```
  retryCount := getXDeathCount(d.Headers)  // baca x-death header
  err := handler(d.Body)
  if err == nil            → d.Ack(false)
  if IsFatal(err) || retryCount >= maxRetry → d.Nack(false, false)  [→ DLQ]
  else                     → d.Nack(false, true)  [requeue]
  ```
- Helper baru: `getXDeathCount(headers amqp.Table) int` — baca `x-death` array, sum semua count

---

## Phase 5 — Application Layer

### 5.1 New Port — `internal/app/port/transactor.go`

```go
type Transactor interface {
    WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

### 5.2 Update Usecases

Langkah pertama: grep `publisher.Publish` di seluruh `internal/app/usecase/` untuk identifikasi semua usecase yang perlu dimigrasi.

Pola perubahan untuk setiap usecase yang terpengaruh (contoh `register.go`):

- **Hapus** `EventPublisher` dependency
- **Inject** `outboxRepo outbox_repository.OutboxRepository` dan `transactor port.Transactor`
- **Wrap** domain save + outbox save dalam `transactor.WithTx()`:
  ```go
  transactor.WithTx(ctx, func(txCtx context.Context) error {
      if err := uc.userRepo.Save(txCtx, user); err != nil { return err }
      if err := uc.profileRepo.Save(txCtx, profile); err != nil { return err }
      return uc.outboxRepo.Save(txCtx, &outbox.Entry{
          ID:         uuid.New(),
          RoutingKey: event.RoutingUserRegistered,
          Payload:    marshaledEventPayload,
          Status:     outbox.StatusPending,
          CreatedAt:  time.Now(),
      })
  })
  ```
- Update `dependencies.go` di paket usecase yang terpengaruh untuk wire deps baru

### 5.3 Update `cmd/app/main.go`

- Init `PostgresTransactor{db}`
- Init `PostgresOutboxRepository{db}`
- Wire ke usecase yang terpengaruh
- **Hapus** `eventPublisher` dari wire usecase (publisher tetap diinit untuk worker, bukan untuk usecase)

---

## Phase 6 — Worker Layer

### 6.1 Handler Registry — `internal/interfaces/mq/registry.go`

```go
type HandlerFunc func(ctx context.Context, msg job.JobMessage) error

type Registry struct {
    handlers map[string]HandlerFunc
}

func NewRegistry() *Registry
func (r *Registry) Register(jobType string, h HandlerFunc)
func (r *Registry) Dispatch(ctx context.Context, msg job.JobMessage) error  // error jika jobType tidak terdaftar
```

### 6.2 Consumer Middleware — `internal/interfaces/mq/consumer_middleware.go`

`JobConsumerMiddleware{registry *Registry, jobRepo job_repository.JobRepository}`

Method `Handle(body []byte) error`:

1. Unmarshal body → `job.JobMessage` (FatalError jika gagal unmarshal)
2. `jobRepo.FindByID(ctx, msg.JobID)` — jika not found: buat Job baru dan `Save`
3. `job.MarkProcessing()` → `jobRepo.Update()`
4. `registry.Dispatch(ctx, msg)`
5. Success → `job.MarkCompleted()` → `jobRepo.Update()`; return nil
6. `RetryableError` → `job.MarkRetrying(err)` → `jobRepo.Update()`; return error (consumer requeue)
7. `FatalError` → `job.MarkDead(err)` → `jobRepo.Update()`; return `job.FatalError` (consumer → DLQ)

### 6.3 Outbox Relay — `internal/interfaces/mq/outbox_relay.go`

```go
type OutboxRelay struct {
    outboxRepo outbox.Repository
    jobRepo    job.Repository
    publisher  port.EventPublisher
    interval   time.Duration
}

func (r *OutboxRelay) Start(ctx context.Context)
```

Setiap tick (`interval`, misal 2s):

1. `outboxRepo.FindPending(ctx, 50)`
2. Untuk setiap entry:
   a. Buat `job.Job{ID: entry.ID, Type: routingKeyToType(entry.RoutingKey), Payload: entry.Payload, Status: PENDING}`
   b. `jobRepo.Save(ctx, job)` — upsert/ignore duplicate (relay restart safe)
   c. Build `job.JobMessage{JobID: entry.ID.String(), Type: job.Type, Version: 1, OccurredAt: entry.CreatedAt, Payload: entry.Payload}`
   d. `publisher.Publish(entry.RoutingKey, jobMessage)`
   e. Success → `outboxRepo.MarkProcessed(ctx, entry.ID)`
   f. Error → `outboxRepo.MarkFailed(ctx, entry.ID, err.Error())` (retry next tick)

### 6.4 Update MQ Handlers — `internal/interfaces/mq/handlers.go`

- Update signature setiap handler: `func(ctx context.Context, msg job.JobMessage) error`
- Unmarshal `msg.Payload` → specific event struct (bukan `body []byte` langsung)
- Tambah fungsi `RegisterAll(r *Registry, handler *MQHandler)` — register semua handlers ke registry

### 6.5 Update `cmd/worker/main.go`

```go
// Init
db := initDB(cfg)
jobRepo := persistence.NewPostgresJobRepository(db)
outboxRepo := persistence.NewPostgresOutboxRepository(db)
publisher := initPublisher(cfg)  // RabbitMQPublisher

// Registry + middleware
registry := mq.NewRegistry()
handler := mq.NewMQHandler(logger, /* deps */)
mq.RegisterAll(registry, handler)
middleware := mq.NewJobConsumerMiddleware(registry, jobRepo)

// Consumer — setiap queue pakai middleware.Handle
consumer.Consume(event.QueueUserRegistered, event.RoutingUserRegistered, retryPolicy.MaxRetryFor("user.registered"), middleware.Handle)
consumer.Consume(event.QueueUserVerified,   event.RoutingUserVerified,   retryPolicy.MaxRetryFor("user.verified"),   middleware.Handle)
consumer.Consume(event.QueueUserLogin,      event.RoutingUserLogin,      retryPolicy.MaxRetryFor("user.login"),      middleware.Handle)

// Outbox relay goroutine
relay := mq.NewOutboxRelay(outboxRepo, jobRepo, publisher, 2*time.Second)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go relay.Start(ctx)
```

---

## File Summary

### File Baru

| Path                                                                | Keterangan                         |
| ------------------------------------------------------------------- | ---------------------------------- |
| `internal/domain/job/entity/job.go`                                 | Job aggregate                      |
| `internal/domain/job/valueobject/job_status.go`                     | JobStatus consts                   |
| `internal/domain/job/repository/job_repository.go`                  | Repository interface               |
| `internal/domain/job/service/retry_policy.go`                       | RetryPolicy domain service         |
| `internal/domain/job/message.go`                                    | JobMessage envelope                |
| `internal/domain/job/error.go`                                      | RetryableError, FatalError         |
| `internal/domain/outbox/entity/outbox_entry.go`                     | OutboxEntry aggregate              |
| `internal/domain/outbox/repository/outbox_repository.go`            | Repository interface               |
| `internal/infrastructure/persistence/transactor.go`                 | PostgresTransactor + TxFromContext |
| `internal/infrastructure/persistence/postgres_job_repository.go`    | Job repo impl                      |
| `internal/infrastructure/persistence/postgres_outbox_repository.go` | Outbox repo impl                   |
| `internal/app/port/transactor.go`                                   | Transactor port interface          |
| `internal/interfaces/mq/registry.go`                                | Handler registry                   |
| `internal/interfaces/mq/consumer_middleware.go`                     | Job lifecycle middleware           |
| `internal/interfaces/mq/outbox_relay.go`                            | Outbox relay worker                |
| `internal/migrations/0004_create_jobs_table.up.sql`                 | Migration                          |
| `internal/migrations/0004_create_jobs_table.down.sql`               | Migration                          |
| `internal/migrations/0005_create_event_outbox_table.up.sql`         | Migration                          |
| `internal/migrations/0005_create_event_outbox_table.down.sql`       | Migration                          |

### File Dimodifikasi

| Path                                                  | Perubahan                                                               |
| ----------------------------------------------------- | ----------------------------------------------------------------------- |
| `internal/infrastructure/external/rabbitmq/client.go` | DLQ topology, prefetch, x-death retry logic, signature `Consume`        |
| `internal/interfaces/mq/handlers.go`                  | Signature handler → `(ctx, job.JobMessage) error`, tambah `RegisterAll` |
| `internal/app/usecase/auth/register.go`               | Replace publisher → outbox + transactor                                 |
| _(usecase lain yang publish event)_                   | Sama seperti register.go                                                |
| `cmd/worker/main.go`                                  | Registry, middleware, outbox relay goroutine, DB wiring                 |
| `cmd/app/main.go`                                     | Transactor, outbox repo, hapus publisher dari usecase wire              |

---

## Verification Checklist

- [ ] `go build ./...` — tidak ada compile error
- [ ] `make migrate-up` — tabel `jobs` dan `event_outbox` terbuat
- [ ] Register user → `event_outbox` ada 1 row PENDING, `jobs` belum ada row
- [ ] Start worker → outbox row → PROCESSED; `jobs` row muncul dengan status COMPLETED
- [ ] Verif RabbitMQ Management UI: message terdeliver ke queue, consumer count = 1
- [ ] Paksa `FatalError` di handler → `jobs` status = DEAD; pesan masuk ke `{queue}.dlq`
- [ ] Paksa `RetryableError` → pesan requeue N kali (lihat `x-death` header count di RabbitMQ UI) → masuk DLQ setelah maxRetry; `jobs` status = DEAD

---

## Decisions & Constraints

- `EventPublisher` port tetap ada — hanya outbox relay yang memanggilnya, usecase tidak lagi bergantung langsung
- Job record **dibuat oleh relay** (bukan usecase) — usecase hanya menyimpan outbox entry
- Relay berjalan di `cmd/worker` process; `FOR UPDATE SKIP LOCKED` menjamin keamanan multi-instance
- Retry menggunakan `x-death` header counting — tidak butuh plugin RabbitMQ tambahan
- Transactor pattern via context key — non-invasif terhadap repo yang sudah ada; hanya repo yang opt-in yang memeriksa `TxFromContext`

## Out of Scope (Fase Berikutnya)

- Event types baru (notification, announcement, FCM)
- Implementasi business logic di handler (kirim welcome email, init notification preferences, dll)
- Delayed retry dengan exponential backoff (butuh TTL queue chain atau plugin)
- Outbox relay berbasis PostgreSQL `LISTEN/NOTIFY` (menggantikan polling)
