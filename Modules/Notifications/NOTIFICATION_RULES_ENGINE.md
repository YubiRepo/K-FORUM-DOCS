# Notification Rules Engine

Dokumen ini menjelaskan **logic backend** yang menentukan kapan, kepada siapa, dan melalui channel apa notifikasi dikirim di platform KAI. Ini adalah referensi utama untuk implementasi Golang — Flutter hanya menerima hasil akhir via FCM.

---

## Daftar Isi

- [Arsitektur Umum](#arsitektur-umum)
- [Async Queue Architecture](#async-queue-architecture)
- [Job Types & Priority](#job-types--priority)
- [Konsep Dasar](#konsep-dasar)
- [Pipeline Pengiriman](#pipeline-pengiriman)
- [Rules Per Modul](#rules-per-modul)
  - [Announcement](#1-announcement)
  - [News](#2-news)
  - [Community](#3-community)
  - [Event](#4-event)
  - [Q&A](#5-qa)
  - [Subscription](#6-subscription)
  - [Reporting](#7-reporting)
  - [Region](#8-region)
- [Bypass Rules](#bypass-rules)
- [Channel Matrix](#channel-matrix)
- [Recipient Resolution](#recipient-resolution)
- [FCM Payload Structure](#fcm-payload-structure)
- [Error Handling & Retry](#error-handling--retry)

---

## Arsitektur Umum

Notifikasi berjalan **sepenuhnya async** — request yang memicu event (publish announcement, approve subscription, dll) selesai segera setelah menyimpan ke DB dan melempar job ke queue. Pengiriman FCM terjadi di background oleh worker pool, sehingga tidak ada delay ke end user maupun ke admin.

```
[HTTP Request — admin publish / aksi user]
      │
      ▼
[Backend: simpan ke DB]  →  return 200 OK  ← selesai, cepat
      │
      ▼  async, non-blocking
[Enqueue NotificationJob ke Redis (asynq)]
      │
      ▼
[Worker Pool]
      ├── Resolve recipients (batch dari DB)
      ├── Check preferences per user
      ├── Check bypass rules
      ├── Build FCM payload
      └── FCM Batch Send (max 500 token/request)
            ├── Success  → log delivered
            ├── Invalid token → deactivate di DB
            └── Failure  → retry (max 3x) → DLQ
```

**Stack:** Redis + [asynq](https://github.com/hibiken/asynq) (Go-native task queue). Dipilih karena:
- Go-native, API bersih, zero boilerplate
- Redis kemungkinan sudah ada di stack (caching/session) — no infra tambahan
- Built-in: retry, DLQ, priority queue, scheduled tasks (untuk reminder cron)
- asynqmon: UI monitoring bawaan untuk inspect antrian dan DLQ

Flutter **tidak punya logic di sini**. Flutter hanya:
- Mendaftarkan FCM token ke backend
- Mengirim/update preferences via API
- Menerima FCM payload dan me-render notifikasi
- Handle deep link dari `click_action`


---

## Async Queue Architecture

### Mengapa Async?

Tanpa async, flow synchronous akan sangat lambat:

```
Publish announcement → query 10.000 users → query FCM tokens
→ 20.000+ FCM requests → return response ke admin
                         ↑ bisa 30-60 detik
```

Dengan async queue:

```
Publish announcement → simpan DB → enqueue job → return 200 OK
                                                  ↑ < 100ms
                          [background]
                          Worker process → FCM batch → selesai dalam detik
```

### Job Structure

Setiap job yang masuk ke queue berisi minimal:

```go
type NotificationJob struct {
    EventType   string          `json:"event_type"`
    EventID     string          `json:"event_id"`    // idempotency key
    Payload     json.RawMessage `json:"payload"`     // data spesifik per event
    IsBypass    bool            `json:"is_bypass"`
    Priority    string          `json:"priority"`    // CRITICAL, HIGH, MEDIUM, LOW
    EnqueuedAt  time.Time       `json:"enqueued_at"`
}
```

### Idempotency

Setiap job punya `event_id` unik. Worker cek apakah event ini sudah pernah diproses sebelum eksekusi — penting untuk mencegah double-send jika job di-retry:

```go
func (e *Engine) Process(ctx context.Context, task *asynq.Task) error {
    var job NotificationJob
    json.Unmarshal(task.Payload(), &job)

    // Cek apakah sudah diproses (idempotency)
    if e.isAlreadyProcessed(ctx, job.EventID) {
        return nil // skip, sudah pernah kirim
    }

    // Mark as processing (dengan TTL sesuai job retention)
    e.markProcessing(ctx, job.EventID)

    // ... proses normal
}
```

### Scheduled Jobs (Reminder)

Cron-based jobs untuk reminder dijadwalkan langsung via asynq tanpa perlu cron daemon terpisah:

```go
// Setup saat server start
scheduler := asynq.NewScheduler(redisOpt, nil)

// Event reminder — cek setiap jam
scheduler.Register("0 * * * *", asynq.NewTask("reminder:event", nil),
    asynq.Queue("low"))

// Expiry reminder — cek setiap hari jam 08:00
scheduler.Register("0 8 * * *", asynq.NewTask("reminder:subscription_expiry", nil),
    asynq.Queue("low"))

scheduler.Run()
```

---

## Job Types & Priority

Daftar semua job types yang terdaftar di worker:

| Job Type | Queue | Trigger |
|---|---|---|
| `notification:announcement` | critical / default | Announcement dipublish |
| `notification:news_published` | default | Artikel baru published |
| `notification:news_status` | critical | Approve/reject artikel Pro Member |
| `notification:community_post` | default | Post baru di komunitas |
| `notification:community_membership` | critical | Join approved/rejected, kick, ban |
| `notification:community_join_request` | critical | Join request masuk ke leader/mod |
| `notification:community_report` | default | Report masuk / auto-flag |
| `notification:event_published` | default | Event baru published |
| `reminder:event` | low | Reminder sebelum event (cron) |
| `notification:qna_answered` | default | Pertanyaan dijawab/ditolak |
| `notification:subscription_status` | critical | Approve/reject/expired/downgrade |
| `reminder:subscription_expiry` | low | Expiry reminder 7 & 3 hari (cron) |
| `notification:region_approved` | critical | Member disetujui masuk region |
| `notification:report_autoflag` | critical | Auto-flag threshold tercapai |


---

## Konsep Dasar

### Tipe Notifikasi

| Tipe | Keterangan |
|---|---|
| `transactional` | Dipicu oleh aksi spesifik user (approved, rejected, answered). Umumnya bypass preferences. |
| `activity` | Aktivitas sosial di platform (post baru, member join). Dikontrol preferences. |
| `system` | Dari sistem/admin (announcement, broadcast). Bisa bypass tergantung priority. |
| `reminder` | Pengingat terjadwal (event reminder, expiry reminder). Dikontrol preferences. |
| `operational` | Notif internal untuk role admin/leader/moderator. Tidak lewat preferences member. |

### Preference Check

Sebelum kirim push ke member, backend **wajib** cek:

```
1. all_notifications_enabled == false  → skip (kecuali bypass)
2. do_not_disturb aktif di jam ini     → skip (kecuali bypass)
3. module.enabled == false             → skip (kecuali bypass)
4. sub-field spesifik == false         → skip (kecuali bypass)
```

### Bypass

Notifikasi dengan flag bypass **mengabaikan semua preference check** dan langsung dikirim. Lihat [Bypass Rules](#bypass-rules) untuk daftar lengkap.

### Recipient Types

| Type | Siapa | Lewat Mana |
|---|---|---|
| `member` | User biasa (standard/pro) | FCM mobile |
| `superadmin` | Admin KAI Pusat | In-app backoffice (web) |
| `admin_region` | Admin KAI Region | In-app backoffice (web) |
| `community_leader` | Leader komunitas tertentu | FCM mobile |
| `community_moderator` | Moderator komunitas tertentu | FCM mobile |
| `usergod` | Developer/vendor | In-app backoffice (web) |

---

## Pipeline Pengiriman

Ada dua tahap pipeline yang berjalan terpisah:

**Tahap 1 — Enqueue (sync, di dalam HTTP request):**
```go
// Dipanggil oleh setiap modul setelah aksi berhasil disimpan ke DB
func (d *Dispatcher) Emit(event DomainEvent) error {
    job := asynq.NewTask(event.Type(), event.Payload())

    // Priority menentukan antrian mana yang dipakai
    queue := d.resolveQueue(event) // "critical", "default", "low"

    _, err := d.client.Enqueue(job,
        asynq.Queue(queue),
        asynq.MaxRetry(3),
    )
    return err // jika enqueue gagal, log tapi jangan gagalkan request utama
}
```

**Tahap 2 — Process (async, di worker pool):**
```go
// Worker mengambil job dari queue dan memprosesnya
func (e *Engine) Process(ctx context.Context, task *asynq.Task) error {
    var event DomainEvent
    json.Unmarshal(task.Payload(), &event)

    // 1. Resolve recipients — query DB dalam batch
    recipients, err := e.resolveRecipients(ctx, event)

    // 2. Split ke batch FCM (max 500 token per request)
    batches := chunkBy(recipients, 500)

    for _, batch := range batches {
        filtered := []Recipient{}

        for _, r := range batch {
            // 3. Cek preferences (kecuali bypass)
            if !event.IsBypass() {
                if !e.checkPreferences(r, event) {
                    continue
                }
            }
            filtered = append(filtered, r)
        }

        // 4. Kirim ke FCM secara batch
        e.sendFCMBatch(ctx, filtered, event)
    }
    return nil
}
```

---

## Rules Per Modul

---

### 1. Announcement

**Event source:** Saat superadmin/admin publish announcement.

| Event | Recipient | Bypass | Channel | Preference Key |
|---|---|---|---|---|
| Announcement publish — tipe `disaster`/`system`/`urgent`, priority `CRITICAL`/`HIGH` | Semua member (global atau per region sesuai scope) | ✅ Ya | Push + Email + In-app | — |
| Announcement publish — tipe `info`, priority `MEDIUM`/`LOW` | Semua member (global atau per region) | ❌ Tidak | Push + In-app | `announcement.info_enabled` |

**Recipient Resolution:**
```
if announcement.scope == "global"  → semua user aktif
if announcement.scope == "regional" → semua user dengan region_id == announcement.region_id
```

**Logic priority:**
```
CRITICAL → force push semua device, bypass DND, bypass all_notifications_enabled
HIGH     → push segera, bypass DND, cek preferences modul
MEDIUM   → push batch (delay beberapa menit), cek preferences
LOW      → in-app only, tidak ada push
```

---

### 2. News

**Event source:** Saat artikel dipublish atau status artikel Member Pro berubah.

| Event | Recipient | Bypass | Channel | Preference Key |
|---|---|---|---|---|
| Artikel baru dipublish | Semua member (subscriber) | ❌ Tidak | Push + In-app | `news.enabled` |
| Artikel Pro Member di-approve | Member penulis | ✅ Ya | Push + In-app | — |
| Artikel Pro Member di-reject | Member penulis | ✅ Ya | Push + In-app | — |
| Artikel Pro Member masuk `pending_approval` | Superadmin (semua) | ✅ Ya | In-app backoffice | — |

**Catatan:**
- Notif artikel baru **hanya dikirim setelah status `published`** — bukan saat `pending_approval` atau `draft`.
- Notif approval/rejection ke penulis adalah transaksional — bypass karena menyangkut status konten milik user.
- Notif ke superadmin masuk antrian backoffice, bukan FCM.

---

### 3. Community

**Event source:** Berbagai aksi di dalam komunitas.

#### Notif ke Member

| Event | Recipient | Bypass | Channel | Preference Key |
|---|---|---|---|---|
| Post baru dipublish di komunitas | Semua anggota aktif komunitas | ❌ Tidak | Push + In-app | `community.communities[id].new_posts` |
| Member lain bergabung | Semua anggota aktif komunitas | ❌ Tidak | In-app | `community.communities[id].member_joined` |
| Member lain keluar | Semua anggota aktif komunitas | ❌ Tidak | In-app | `community.communities[id].member_left` |
| Join request di-approve (private) | Member pemohon | ❌ Tidak | Push + In-app | `community.communities[id].join_request_approved` |
| Join request di-reject (private) | Member pemohon | ✅ Ya | Push + In-app | — |
| Kena kick dari komunitas | Member yang di-kick | ✅ Ya | Push + In-app | — |
| Kena ban dari komunitas | Member yang di-ban | ✅ Ya | Push + In-app | — |
| Komunitas di-suspend superadmin | Leader komunitas | ✅ Ya | Push + In-app | — |

#### Notif ke Leader / Moderator (Operational)

| Event | Recipient | Bypass | Channel |
|---|---|---|---|
| Join request masuk baru (private community) | Leader + Moderator komunitas tersebut | ✅ Ya | Push + In-app |
| Report masuk — scope komunitas | Leader + Moderator komunitas tersebut (yang punya `manage_reports`) | ✅ Ya | In-app |
| Auto-flag report (report_count ≥ threshold) | Leader + Moderator + Superadmin | ✅ Ya | Push + In-app |

#### Notif ke Superadmin (Operational)

| Event | Recipient | Channel |
|---|---|---|
| Komunitas masuk status `orphaned` | Superadmin | In-app backoffice |

**Recipient Resolution untuk community notifications:**
```
// Notif ke anggota komunitas
recipients = community_members
    WHERE community_id = event.community_id
    AND status = "active"
    AND user_id != event.actor_id  // jangan notif diri sendiri

// Notif ke leader/moderator
recipients = user_roles
    WHERE scope_type = "community"
    AND scope_id = event.community_id
    AND role IN ("leader", "moderator")
```

---

### 4. Event

**Event source:** Saat event dipublish atau scheduler reminder berjalan.

| Event | Recipient | Bypass | Channel | Preference Key |
|---|---|---|---|---|
| Event baru dipublish | Member yang `interested_categories` match (atau semua jika kosong) | ❌ Tidak | Push + In-app | `event.enabled` |
| Reminder sebelum event | Member | ❌ Tidak | Push + In-app | `event.reminders_enabled` |

**Reminder Logic:**
```
// Cron job berjalan setiap jam
events = SELECT * FROM events
    WHERE start_time BETWEEN NOW() + reminder_hours AND NOW() + reminder_hours + 1h
    AND status = "published"

for each event:
    recipients = members WHERE event.reminders_enabled == true
    // kirim dengan reminder_hours_before sesuai preference masing-masing user
```

**Category Filter:**
```
if member.interested_categories == [] → kirim semua event
if member.interested_categories != [] → kirim hanya jika event.category IN interested_categories
```

---

### 5. Q&A

**Event source:** Saat superadmin menjawab atau menolak pertanyaan member.

| Event | Recipient | Bypass | Channel | Preference Key |
|---|---|---|---|---|
| Pertanyaan dijawab | Member penanya | ❌ Tidak | Push + In-app | `qna.question_answered` |
| Pertanyaan ditolak (+ alasan) | Member penanya | ❌ Tidak | Push + In-app | `qna.question_answered` |
| Pertanyaan dikonversi ke FAQ | Member penanya | ❌ Tidak | Push + In-app | `qna.question_answered` |
| Pertanyaan baru masuk | Superadmin | ✅ Ya | In-app backoffice | — |

**Catatan:** Kedua kondisi (dijawab dan ditolak) menggunakan preference key yang sama `qna.question_answered` karena keduanya adalah update status pertanyaan dari sudut pandang member.

---

### 6. Subscription

**Event source:** Saat status subscription member berubah atau scheduler reminder berjalan.

#### Notif ke Member

| Event | Recipient | Bypass | Channel | Preference Key |
|---|---|---|---|---|
| Request upgrade disubmit (konfirmasi) | Member | ✅ Ya | Push + Email + In-app | — |
| Upgrade di-approve | Member | ✅ Ya | Push + Email + In-app | — |
| Upgrade di-reject (+ alasan) | Member | ✅ Ya | Push + Email + In-app | — |
| Plan expired / downgraded | Member | ✅ Ya | Push + Email + In-app | — |
| Expiry reminder 7 hari | Member (Pro aktif) | ❌ Tidak | Push + Email + In-app | `subscription.expiry_reminder_enabled` |
| Expiry reminder 3 hari | Member (Pro aktif) | ❌ Tidak | Push + In-app | `subscription.expiry_reminder_enabled` |

#### Notif ke Superadmin (Operational)

| Event | Recipient | Channel |
|---|---|---|
| Request upgrade baru masuk | Superadmin | In-app backoffice |

**Expiry Reminder Logic:**
```
// Cron job berjalan setiap hari jam 08:00
subscriptions = SELECT * FROM user_subscriptions
    WHERE plan = "pro"
    AND status = "active"
    AND expires_at BETWEEN NOW() + 7d AND NOW() + 8d  // window 7 hari
    → kirim reminder "7 hari lagi"

subscriptions = SELECT * FROM user_subscriptions
    WHERE plan = "pro"
    AND status = "active"
    AND expires_at BETWEEN NOW() + 3d AND NOW() + 4d  // window 3 hari
    → kirim reminder "3 hari lagi"

// Cek preferences sebelum kirim
if user.preferences.subscription.expiry_reminder_enabled == false → skip
```

---

### 7. Reporting

**Event source:** Saat report masuk atau mencapai threshold auto-flag.

#### Content Report

| Event | Recipient | Bypass | Channel |
|---|---|---|---|
| Report baru masuk — scope komunitas | Leader + Moderator komunitas (dengan `manage_reports`) | ✅ Ya | In-app |
| Auto-flag — scope komunitas (report_count ≥ threshold) | Leader + Moderator komunitas + Superadmin | ✅ Ya | Push + In-app |
| Report baru masuk — scope global | Superadmin | ✅ Ya | In-app backoffice |
| Auto-flag — scope global (report_count ≥ threshold) | Superadmin | ✅ Ya | Push + In-app backoffice |
| Laporan selesai diproses (opsional) | Member pelapor | ❌ Tidak | In-app |

**Auto-flag Threshold Logic:**
```
// Dipanggil setiap kali report baru masuk
if target.report_count >= config.auto_flag_threshold (default: 5):
    target.is_flagged = true
    emit AutoFlagEvent → notify penangan
```

#### Bug Report

| Event | Recipient | Channel |
|---|---|---|
| Bug report baru masuk | Usergod + Superadmin | In-app backoffice |
| Bug report diterima (konfirmasi ke pelapor) | Member pelapor | In-app |

---

### 8. Region

**Event source:** Saat member join atau disetujui masuk ke region.

| Event | Recipient | Bypass | Channel |
|---|---|---|---|
| Member disetujui masuk region | Member | ✅ Ya | Push + In-app |
| Member baru bergabung ke region | Admin region | ✅ Ya | In-app backoffice |

---

## Bypass Rules

Notifikasi berikut **mengabaikan semua preference check** dan selalu dikirim, termasuk saat `all_notifications_enabled = false` dan saat DND aktif.

```go
func (e *Engine) isBypass(event DomainEvent) bool {
    switch event.Type {
    // Announcement
    case AnnouncementPublished:
        return event.Priority == "CRITICAL" || event.Priority == "HIGH"

    // News — transaksional ke penulis
    case NewsApproved, NewsRejected:
        return true

    // Community — status akun/keanggotaan
    case JoinRequestRejected, MemberKicked, MemberBanned, CommunitySuspended:
        return true

    // Community — operasional leader/moderator
    case JoinRequestReceived, ReportAutoFlagged:
        return true

    // Subscription — semua perubahan status akun
    case SubscriptionSubmitted, SubscriptionApproved,
         SubscriptionRejected, SubscriptionExpired, SubscriptionDowngraded:
        return true

    // Region
    case RegionApproved:
        return true

    default:
        return false
    }
}
```

---

## Channel Matrix

Ringkasan channel per tipe notifikasi:

| Channel | Siapa | Kapan dipakai |
|---|---|---|
| **FCM Push (mobile)** | Member, Leader, Moderator | Notif real-time yang perlu perhatian segera |
| **Email** | Member | Transaksional penting (subscription, announcement CRITICAL) |
| **In-app mobile** | Member | Semua notif, tersimpan di notification center |
| **In-app backoffice (web)** | Superadmin, Admin region, Usergod | Antrian operasional — pending approval, report masuk, dll |

**Catatan penting:** Superadmin, Admin region, dan Usergod **tidak menerima FCM push** untuk notif operasional — semua masuk antrian in-app backoffice. FCM hanya untuk notif yang perlu segera dibaca di mobile.

---

## Recipient Resolution

Helper functions standar yang digunakan Rules Engine:

```go
// Semua member aktif di platform
func AllActiveMembers(db) []User

// Semua member aktif di region tertentu
func MembersByRegion(db, regionID) []User

// Semua anggota aktif komunitas (kecuali actor)
func CommunityMembers(db, communityID, excludeUserID) []User

// Leader + moderator komunitas dengan permission tertentu
func CommunityStaff(db, communityID, permission string) []User

// Semua superadmin
func AllSuperadmins(db) []User

// Admin region tertentu
func RegionAdmins(db, regionID) []User

// Satu user spesifik (untuk transaksional)
func SingleUser(db, userID) []User
```

---

## FCM Payload Structure

Standar payload yang dikirim ke FCM untuk semua notifikasi push:

```json
{
  "title": "string",
  "body": "string",
  "image_url": "string | null",
  "data": {
    "type": "announcement | news | community | event | qna | subscription | report",
    "entity_id": "uuid | null",
    "click_action": "kai://screen/entity_id",
    "bypass": "true | false"
  }
}
```

**Deep link convention (`click_action`):**

| Modul | Format |
|---|---|
| Announcement | `kai://announcements/{id}` |
| News | `kai://news/{id}` |
| Community post | `kai://communities/{community_id}/posts/{post_id}` |
| Community join request | `kai://communities/{community_id}/members` |
| Event | `kai://events/{id}` |
| Q&A | `kai://qna/questions/{id}` |
| Subscription | `kai://subscription` |

---

## Error Handling & Retry

### FCM-level errors

```
Token invalid / not registered
    → deactivate token di DB (is_active = false)
    → lanjut ke token lain milik user yang sama
    → tidak trigger retry job (bukan error yang bisa diperbaiki dengan retry)

FCM server error (5xx, timeout)
    → asynq otomatis retry job (max 3x)
    → backoff: 1s, 4s, 16s (exponential)
    → jika 3x masih gagal → job masuk Dead Letter Queue (DLQ)

Partial delivery (sebagian token sukses, sebagian gagal)
    → catat per-token result di notification_delivery_logs
    → status keseluruhan = "partial"
```

### Dead Letter Queue (DLQ)

Job yang gagal setelah max retry masuk ke DLQ. Bisa di-inspect dan di-retry manual via asynqmon UI.

```
DLQ retention: 7 hari (setelah itu auto-delete)
Monitoring: asynqmon dashboard (internal tool)
Action options:
    → Retry manual (jika root cause sudah diperbaiki)
    → Archive (jika sudah tidak relevan, misalnya event lama)
    → Delete
```

### Preferences DB tidak bisa diakses

```
→ fail-safe: JANGAN kirim (default ke opt-out)
→ kecuali event adalah bypass → tetap kirim
→ log error, job tetap selesai (tidak masuk retry/DLQ)
  karena ini infrastructure issue, bukan job failure
```

### Queue priorities

asynq mendukung multiple queues dengan priority berbeda. Notifikasi di-route sesuai urgency:

| Queue | Priority | Dipakai untuk |
|---|---|---|
| `critical` | Tertinggi | Announcement CRITICAL, subscription status change, kick/ban |
| `default` | Normal | Semua notifikasi bypass lainnya, activity notif |
| `low` | Rendah | Reminder (event, expiry), notif low-priority |

```go
func (d *Dispatcher) resolveQueue(event DomainEvent) string {
    if event.IsBypass() && event.Priority() == "CRITICAL" {
        return "critical"
    }
    if event.Type() == ReminderEvent {
        return "low"
    }
    return "default"
}
```

### Scaling worker pool

Worker count bisa di-tune sesuai load. Starting point yang aman:

```go
srv := asynq.NewServer(redisOpt, asynq.Config{
    Queues: map[string]int{
        "critical": 6,  // 6 goroutine untuk antrian critical
        "default":  3,
        "low":      1,
    },
    // Total max 10 concurrent workers
})
```

Untuk broadcast ke 10.000 user dengan 2 device rata-rata = 20.000 FCM requests.
Dengan batch 500 token = 40 FCM batch requests. Dengan 6 worker di queue critical, selesai dalam hitungan detik.

---

*Dokumen ini adalah referensi backend (Golang) untuk Notification Rules Engine. Untuk API preferences mobile lihat `API_SPEC_NOTIFICATION_PREFERENCES.md`. Untuk FCM token management lihat `API_SPEC_FCM.md`.*
