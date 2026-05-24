# Database Schema — KAI Subscription Module

> **Stack:** Golang + PostgreSQL  
> **Sumber:** `API_SPEC_Subscription_Mobile.md` & `API_SPEC_Subscription_Backoffice.md`  
> **Dibuat:** 2026-05-24

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [Tabel: plans](#1-plans)
3. [Tabel: benefit_master](#2-benefit_master)
4. [Tabel: plan_benefits](#3-plan_benefits)
5. [Tabel: subscriptions](#4-subscriptions)
6. [Tabel: subscription_requests](#5-subscription_requests)
7. [Tabel: subscription_history](#6-subscription_history)
8. [Tabel: refunds](#7-refunds)
9. [Ringkasan Enum Values](#ringkasan-enum-values)
10. [Catatan Implementasi Golang](#catatan-implementasi-golang)

---

## Overview Relasi

```
users
  └── subscriptions          (1 user → 1 active subscription)
  └── subscription_requests  (1 user → banyak requests)
  └── subscription_history   (1 user → banyak history entries)
  └── refunds                (1 user → banyak refunds)

plans
  └── plan_benefits          (1 plan → banyak assigned benefits)
        └── benefit_master   (referensi ke master benefit)

subscription_requests
  └── approved/rejected by → users (admin/superadmin)
  └── linked to → subscription_history
  └── linked to → refunds
```

### Entity Relationship Diagram (Teks)

```
plans ──────────────────── plan_benefits ─────────── benefit_master
  │                              │
  │                              │
  ├── subscriptions ─────── users ──── subscription_requests ── refunds
  │         │                              │
  └─────────┴── subscription_history ──────┘
```

---

## 1. `plans`

Master data plan yang tersedia. Di-manage hanya oleh superadmin. Jarang berubah — cocok di-cache di application layer.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `name` | `VARCHAR(100)` | NOT NULL | Contoh: `"Standard"`, `"Pro"` |
| `price` | `BIGINT` | NOT NULL | Harga dalam satuan terkecil mata uang (IDR = rupiah) |
| `currency` | `VARCHAR(10)` | NOT NULL, DEFAULT `'IDR'` | Kode mata uang ISO 4217 |
| `duration_days` | `INT` | NOT NULL, DEFAULT `30` | Durasi berlaku dalam hari |
| `description` | `TEXT` | NULLABLE | Deskripsi singkat plan |
| `is_default` | `BOOLEAN` | NOT NULL, DEFAULT `false` | Hanya boleh 1 row yang `true` |
| `status` | `VARCHAR(20)` | NOT NULL, DEFAULT `'active'` | Lihat enum: `active`, `disabled` |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | Auto-update via trigger |

**DDL Indexes & Constraints:**

```sql
-- Hanya 1 plan yang boleh jadi default
CREATE UNIQUE INDEX idx_plans_only_one_default
  ON plans (is_default)
  WHERE is_default = true;

-- Nama plan harus unik
CREATE UNIQUE INDEX idx_plans_name_unique
  ON plans (LOWER(name));
```

---

## 2. `benefit_master`

Master list semua benefit yang bisa di-assign ke plan. Hanya bisa dibuat/diubah oleh `usergod`. Tabel ini berperan sebagai "kamus" benefit yang valid di sistem.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `key` | `VARCHAR(100)` | NOT NULL, UNIQUE | Identifier unik benefit, e.g. `post_news` |
| `description` | `TEXT` | NULLABLE | Penjelasan benefit untuk ditampilkan di UI |
| `approval_required` | `BOOLEAN` | NOT NULL, DEFAULT `false` | Apakah aksi ini perlu persetujuan admin |
| `approval_role` | `VARCHAR(50)` | NULLABLE | Role yang berwenang approve, e.g. `superadmin` |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |

**Predefined benefit keys** (dari API spec):

| Key | Deskripsi | Approval Required | Approval Role |
|---|---|---|---|
| `read_content` | Baca news, event, direktori | `false` | — |
| `join_community` | Join komunitas | `false` | — |
| `post_community` | Post di komunitas (unlimited) | `false` | — |
| `ask_qna` | Tanya di Q&A | `false` | — |
| `post_news` | Post news | `true` | `superadmin` |
| `create_community` | Buat dan manage community | `false` | — |
| `create_store` | Buat store/merchant listing | `false` | — |
| `create_event` | Buat event | `false` | — |
| `view_analytics` | Lihat analytics | `false` | — |

---

## 3. `plan_benefits`

Join table antara `plans` dan `benefit_master`. Menentukan benefit mana yang aktif di plan tertentu. Satu plan bisa punya banyak benefit, dan konfigurasi bisa diubah oleh superadmin kapan saja.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `plan_id` | `UUID` | FK → `plans.id`, ON DELETE CASCADE | |
| `benefit_key` | `VARCHAR(100)` | FK → `benefit_master.key` | |
| `enabled` | `BOOLEAN` | NOT NULL, DEFAULT `true` | Toggle benefit on/off per plan |
| `metadata` | `JSONB` | NULLABLE | Konfigurasi tambahan, e.g. `{"approval_required": true}` |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |

**DDL Indexes & Constraints:**

```sql
-- Satu plan tidak boleh punya benefit key yang sama dua kali
CREATE UNIQUE INDEX idx_plan_benefits_unique
  ON plan_benefits (plan_id, benefit_key);

CREATE INDEX idx_plan_benefits_plan_id
  ON plan_benefits (plan_id);
```

---

## 4. `subscriptions`

State subscription **aktif saat ini** per user. Satu user hanya boleh punya satu row. Diupdate (bukan insert baru) setiap kali terjadi upgrade, renewal, atau downgrade.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `user_id` | `UUID` | NOT NULL, FK → `users.id`, ON DELETE CASCADE, UNIQUE | Enforce 1 subscription per user |
| `plan_id` | `UUID` | NOT NULL, FK → `plans.id` | Plan yang sedang aktif |
| `status` | `VARCHAR(20)` | NOT NULL | `active`, `expired`, `cancelled` |
| `current_period_start` | `DATE` | NOT NULL | Tanggal mulai periode aktif |
| `current_period_end` | `DATE` | NOT NULL | Tanggal berakhir periode aktif |
| `auto_renewal` | `BOOLEAN` | NOT NULL, DEFAULT `true` | Flag apakah akan auto-renew |
| `renewal_reminder_sent` | `BOOLEAN` | NOT NULL, DEFAULT `false` | Flag sudah kirim reminder expiry |
| `downgrade_to_plan_id` | `UUID` | NULLABLE, FK → `plans.id` | Diisi jika user request downgrade; efektif saat expiry |
| `downgrade_reason` | `TEXT` | NULLABLE | Alasan downgrade dari user |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |

**DDL Indexes:**

```sql
-- Enforce 1 subscription per user
CREATE UNIQUE INDEX idx_subscriptions_user_id
  ON subscriptions (user_id);

-- Query by status (e.g. cron job cek expiry)
CREATE INDEX idx_subscriptions_status
  ON subscriptions (status);

-- Cron job renewal reminder: cari yang period_end mendekati hari ini
CREATE INDEX idx_subscriptions_period_end
  ON subscriptions (current_period_end);

-- Compound index untuk query umum
CREATE INDEX idx_subscriptions_user_status
  ON subscriptions (user_id, status);
```

---

## 5. `subscription_requests`

Setiap request upgrade dari user — baik via transfer manual maupun payment gateway. Status berubah seiring approval flow. Satu user hanya boleh punya **satu pending request** dalam satu waktu.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `user_id` | `UUID` | NOT NULL, FK → `users.id`, ON DELETE CASCADE | |
| `current_plan_id` | `UUID` | NOT NULL, FK → `plans.id` | Plan user saat request dibuat |
| `requested_plan_id` | `UUID` | NOT NULL, FK → `plans.id` | Plan yang diminta |
| `amount` | `BIGINT` | NOT NULL | Nominal yang harus dibayar (dalam IDR) |
| `status` | `VARCHAR(20)` | NOT NULL, DEFAULT `'pending'` | `pending`, `completed`, `rejected`, `failed`, `cancelled` |
| `payment_method` | `VARCHAR(20)` | NOT NULL | `manual`, `stripe`, `midtrans` |
| `payment_provider_transaction_id` | `VARCHAR(255)` | NULLABLE | Transaction ID dari payment gateway |
| `manual_proof_url` | `TEXT` | NULLABLE | URL bukti transfer (manual payment) |
| `manual_proof_type` | `VARCHAR(20)` | NULLABLE | `image`, `video` |
| `verified_by_admin` | `BOOLEAN` | NOT NULL, DEFAULT `false` | Flag sudah dicek admin |
| `admin_notes` | `TEXT` | NULLABLE | Catatan admin saat approve |
| `rejection_reason` | `VARCHAR(100)` | NULLABLE | Reason singkat: `"Amount doesn't match"`, `"Invalid proof"`, dll |
| `rejection_details` | `TEXT` | NULLABLE | Detail panjang alasan reject |
| `approved_by` | `UUID` | NULLABLE, FK → `users.id` | Admin/superadmin yang approve atau reject |
| `approved_at` | `TIMESTAMPTZ` | NULLABLE | Timestamp saat di-approve atau di-reject |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |

**DDL Indexes & Constraints:**

```sql
CREATE INDEX idx_sub_requests_user_id
  ON subscription_requests (user_id);

CREATE INDEX idx_sub_requests_status
  ON subscription_requests (status);

CREATE INDEX idx_sub_requests_created_at
  ON subscription_requests (created_at DESC);

-- Enforce hanya 1 pending request per user dalam satu waktu
CREATE UNIQUE INDEX idx_sub_requests_one_pending
  ON subscription_requests (user_id)
  WHERE status = 'pending';

-- Query backoffice: filter by status + tanggal
CREATE INDEX idx_sub_requests_status_created
  ON subscription_requests (status, created_at DESC);
```

---

## 6. `subscription_history`

Audit trail lengkap setiap perubahan subscription. Tabel ini **append-only** — tidak ada UPDATE atau DELETE. Setiap aksi yang mengubah state subscription harus meninggalkan satu row di sini.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `user_id` | `UUID` | NOT NULL, FK → `users.id`, ON DELETE CASCADE | |
| `old_plan_id` | `UUID` | NULLABLE, FK → `plans.id` | `NULL` saat action = `signup` (belum punya plan sebelumnya) |
| `new_plan_id` | `UUID` | NOT NULL, FK → `plans.id` | Plan baru setelah aksi |
| `action` | `VARCHAR(30)` | NOT NULL | `signup`, `upgrade`, `downgrade`, `renewal`, `expired`, `cancelled` |
| `initiated_by` | `VARCHAR(20)` | NOT NULL | `user`, `admin`, `system` |
| `initiated_by_user_id` | `UUID` | NULLABLE, FK → `users.id` | Diisi hanya jika `initiated_by = 'admin'` |
| `request_id` | `UUID` | NULLABLE, FK → `subscription_requests.id` | Link ke request yang memicu history ini |
| `notes` | `TEXT` | NULLABLE | Catatan tambahan, e.g. alasan downgrade dari user |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |

**DDL Indexes:**

```sql
CREATE INDEX idx_sub_history_user_id
  ON subscription_history (user_id);

-- Query history terbaru per user
CREATE INDEX idx_sub_history_user_created
  ON subscription_history (user_id, created_at DESC);

-- Query backoffice: filter by action + tanggal
CREATE INDEX idx_sub_history_action_created
  ON subscription_history (action, created_at DESC);
```

> **Catatan:** Karena append-only, tabel ini tidak perlu `updated_at`. Jangan pernah `UPDATE` atau `DELETE` row di tabel ini — tambah row baru jika ada koreksi.

---

## 7. `refunds`

Record manual refund yang diproses oleh superadmin. Setiap refund mengakibatkan subscription di-cancel dan user di-downgrade ke plan default (Standard).

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `user_id` | `UUID` | NOT NULL, FK → `users.id` | |
| `request_id` | `UUID` | NULLABLE, FK → `subscription_requests.id` | Request yang menjadi dasar refund |
| `subscription_id` | `UUID` | NULLABLE, FK → `subscriptions.id` | Subscription yang dibatalkan |
| `refund_amount` | `BIGINT` | NOT NULL | Jumlah yang di-refund (dalam IDR) |
| `currency` | `VARCHAR(10)` | NOT NULL, DEFAULT `'IDR'` | |
| `reason` | `VARCHAR(100)` | NOT NULL | `User request`, `Payment error`, `Service issue`, `Other` |
| `notes` | `TEXT` | NULLABLE | Detail tambahan, e.g. nomor rekening tujuan refund |
| `old_plan_id` | `UUID` | NOT NULL, FK → `plans.id` | Plan sebelum refund |
| `new_plan_id` | `UUID` | NOT NULL, FK → `plans.id` | Plan setelah refund (biasanya Standard) |
| `refund_status` | `VARCHAR(20)` | NOT NULL, DEFAULT `'processed'` | `processed`, `pending`, `failed` |
| `processed_by` | `UUID` | NOT NULL, FK → `users.id` | Superadmin yang memproses refund |
| `refund_date` | `DATE` | NOT NULL | Tanggal refund diproses |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |

**DDL Indexes:**

```sql
CREATE INDEX idx_refunds_user_id
  ON refunds (user_id);

CREATE INDEX idx_refunds_created_at
  ON refunds (created_at DESC);
```

---

## Ringkasan Enum Values

### `plans.status`

| Value | Keterangan |
|---|---|
| `active` | Plan tersedia untuk dibeli/diassign |
| `disabled` | Plan disembunyikan, tidak bisa dibeli baru (existing subscriber tidak terpengaruh) |

---

### `subscriptions.status`

| Value | Keterangan |
|---|---|
| `active` | Subscription berlaku, dalam periode aktif |
| `expired` | Melewati `current_period_end`, belum diperbarui |
| `cancelled` | Dibatalkan paksa (refund / manual oleh admin) |

---

### `subscription_requests.status`

| Value | Trigger | Keterangan |
|---|---|---|
| `pending` | User submit request | Menunggu verifikasi admin (manual) atau callback gateway |
| `completed` | Admin approve / webhook sukses | Subscription sudah diaktifkan |
| `rejected` | Admin reject | Ditolak manual oleh admin |
| `failed` | Webhook gagal dari gateway | Pembayaran tidak berhasil di sisi gateway |
| `cancelled` | User cancel sendiri | User membatalkan request sebelum diproses |

---

### `subscription_requests.payment_method`

| Value | Keterangan | Flow |
|---|---|---|
| `manual` | Transfer bank manual | User upload bukti → admin verifikasi → approve/reject |
| `stripe` | Stripe payment gateway | Otomatis via Stripe webhook, tidak perlu approval admin |
| `midtrans` | Midtrans payment gateway | Otomatis via Midtrans webhook, tidak perlu approval admin |

---

### `subscription_history.action`

| Value | `initiated_by` | Keterangan |
|---|---|---|
| `signup` | `system` | Saat user pertama kali registrasi, auto-assign plan default |
| `upgrade` | `user` / `admin` | Naik ke plan yang lebih tinggi |
| `downgrade` | `user` / `admin` | Turun ke plan yang lebih rendah, efektif saat expiry |
| `renewal` | `system` | Auto-renewal di akhir periode |
| `expired` | `system` | Plan expired, tidak ada renewal |
| `cancelled` | `admin` | Dibatalkan paksa oleh admin (biasanya bersamaan dengan refund) |

---

### `subscription_history.initiated_by`

| Value | Keterangan |
|---|---|
| `user` | Aksi dipicu langsung oleh user (request upgrade, cancel renewal, dll) |
| `admin` | Aksi dipicu oleh admin/superadmin dari backoffice |
| `system` | Aksi otomatis oleh sistem (renewal, expiry check dari cron job) |

---

## Catatan Implementasi Golang

### Tipe Data Mapping

| PostgreSQL | Golang (nullable) | Golang (not null) |
|---|---|---|
| `UUID` | `*string` | `string` |
| `TIMESTAMPTZ` | `*time.Time` | `time.Time` |
| `DATE` | `*time.Time` | `time.Time` |
| `JSONB` | `json.RawMessage` | `json.RawMessage` |
| `BOOLEAN` | `*bool` | `bool` |
| `BIGINT` | `*int64` | `int64` |
| `INT` | `*int` | `int` |
| `VARCHAR` / `TEXT` | `*string` | `string` |

> Konvensi: gunakan **pointer** (`*T`) untuk semua kolom NULLABLE. Lebih idiomatik di Golang dan mudah dicek dengan `if field != nil`.

---

### Struct: `Plan`

Memetakan tabel `plans`.

```go
// internal/domain/subscription/plan.go

package subscription

import "time"

// Plan merepresentasikan satu paket langganan yang tersedia.
type Plan struct {
    ID           string    `db:"id"           json:"id"`
    Name         string    `db:"name"         json:"name"`
    Price        int64     `db:"price"        json:"price"`
    Currency     string    `db:"currency"     json:"currency"`
    DurationDays int       `db:"duration_days" json:"duration_days"`
    Description  *string   `db:"description"  json:"description,omitempty"`
    IsDefault    bool      `db:"is_default"   json:"is_default"`
    Status       string    `db:"status"       json:"status"`
    CreatedAt    time.Time `db:"created_at"   json:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"   json:"updated_at"`
}
```

---

### Struct: `BenefitMaster`

Memetakan tabel `benefit_master`.

```go
// internal/domain/subscription/benefit_master.go

package subscription

import "time"

// BenefitMaster adalah master data semua benefit yang dikenal sistem.
// Hanya bisa dibuat/diubah oleh usergod.
type BenefitMaster struct {
    ID               string    `db:"id"                json:"id"`
    Key              string    `db:"key"               json:"key"`
    Description      *string   `db:"description"       json:"description,omitempty"`
    ApprovalRequired bool      `db:"approval_required" json:"approval_required"`
    ApprovalRole     *string   `db:"approval_role"     json:"approval_role,omitempty"`
    CreatedAt        time.Time `db:"created_at"        json:"created_at"`
}

// Konstanta untuk benefit key yang sudah diketahui.
const (
    BenefitKeyReadContent      = "read_content"
    BenefitKeyJoinCommunity    = "join_community"
    BenefitKeyPostCommunity    = "post_community"
    BenefitKeyAskQnA           = "ask_qna"
    BenefitKeyPostNews         = "post_news"
    BenefitKeyCreateCommunity  = "create_community"
    BenefitKeyCreateStore      = "create_store"
    BenefitKeyCreateEvent      = "create_event"
    BenefitKeyViewAnalytics    = "view_analytics"
)
```

---

### Struct: `PlanBenefit`

Memetakan tabel `plan_benefits`.

```go
// internal/domain/subscription/plan_benefit.go

package subscription

import (
    "encoding/json"
    "time"
)

// PlanBenefit merepresentasikan assignment benefit ke satu plan tertentu.
type PlanBenefit struct {
    ID         string          `db:"id"          json:"id"`
    PlanID     string          `db:"plan_id"     json:"plan_id"`
    BenefitKey string          `db:"benefit_key" json:"benefit_key"`
    Enabled    bool            `db:"enabled"     json:"enabled"`
    Metadata   json.RawMessage `db:"metadata"    json:"metadata,omitempty"`
    CreatedAt  time.Time       `db:"created_at"  json:"created_at"`
    UpdatedAt  time.Time       `db:"updated_at"  json:"updated_at"`
}

// PlanBenefitMetadata adalah struktur opsional untuk field Metadata.
// Tidak semua benefit punya metadata.
type PlanBenefitMetadata struct {
    ApprovalRequired *bool `json:"approval_required,omitempty"`
}

// PlanWithBenefits digunakan untuk response API yang menyertakan
// detail benefit di dalam satu plan.
type PlanWithBenefits struct {
    Plan
    Benefits []PlanBenefit `db:"-" json:"benefits"`
}
```

---

### Struct: `Subscription`

Memetakan tabel `subscriptions`.

```go
// internal/domain/subscription/subscription.go

package subscription

import "time"

// Subscription merepresentasikan state langganan aktif seorang user.
// Satu user hanya boleh punya satu row (UNIQUE pada user_id).
type Subscription struct {
    ID                  string    `db:"id"                   json:"id"`
    UserID              string    `db:"user_id"              json:"user_id"`
    PlanID              string    `db:"plan_id"              json:"plan_id"`
    Status              string    `db:"status"               json:"status"`
    CurrentPeriodStart  time.Time `db:"current_period_start" json:"current_period_start"`
    CurrentPeriodEnd    time.Time `db:"current_period_end"   json:"current_period_end"`
    AutoRenewal         bool      `db:"auto_renewal"         json:"auto_renewal"`
    RenewalReminderSent bool      `db:"renewal_reminder_sent" json:"renewal_reminder_sent"`
    DowngradeToPlanID   *string   `db:"downgrade_to_plan_id" json:"downgrade_to_plan_id,omitempty"`
    DowngradeReason     *string   `db:"downgrade_reason"     json:"downgrade_reason,omitempty"`
    CreatedAt           time.Time `db:"created_at"           json:"created_at"`
    UpdatedAt           time.Time `db:"updated_at"           json:"updated_at"`
}

// Enum nilai yang valid untuk field Status.
const (
    SubscriptionStatusActive    = "active"
    SubscriptionStatusExpired   = "expired"
    SubscriptionStatusCancelled = "cancelled"
)

// SubscriptionWithPlan digunakan saat response API perlu menyertakan
// detail plan di dalam subscription (e.g. GET /subscription/me).
type SubscriptionWithPlan struct {
    Subscription
    Plan *Plan `db:"-" json:"plan,omitempty"`
}

// PendingRequestSummary adalah ringkasan pending request
// yang disertakan dalam response GET /subscription/me.
type PendingRequestSummary struct {
    ID            string    `json:"id"`
    RequestedPlan string    `json:"requested_plan"`
    Status        string    `json:"status"`
    SubmittedAt   time.Time `json:"submitted_at"`
    Amount        int64     `json:"amount"`
}
```

---

### Struct: `SubscriptionRequest`

Memetakan tabel `subscription_requests`.

```go
// internal/domain/subscription/subscription_request.go

package subscription

import "time"

// SubscriptionRequest merepresentasikan satu permintaan upgrade plan dari user.
type SubscriptionRequest struct {
    ID                           string     `db:"id"                             json:"id"`
    UserID                       string     `db:"user_id"                        json:"user_id"`
    CurrentPlanID                string     `db:"current_plan_id"                json:"current_plan_id"`
    RequestedPlanID              string     `db:"requested_plan_id"              json:"requested_plan_id"`
    Amount                       int64      `db:"amount"                         json:"amount"`
    Status                       string     `db:"status"                         json:"status"`
    PaymentMethod                string     `db:"payment_method"                 json:"payment_method"`
    PaymentProviderTransactionID *string    `db:"payment_provider_transaction_id" json:"payment_provider_transaction_id,omitempty"`
    ManualProofURL               *string    `db:"manual_proof_url"               json:"manual_proof_url,omitempty"`
    ManualProofType              *string    `db:"manual_proof_type"              json:"manual_proof_type,omitempty"`
    VerifiedByAdmin              bool       `db:"verified_by_admin"              json:"verified_by_admin"`
    AdminNotes                   *string    `db:"admin_notes"                    json:"admin_notes,omitempty"`
    RejectionReason              *string    `db:"rejection_reason"               json:"rejection_reason,omitempty"`
    RejectionDetails             *string    `db:"rejection_details"              json:"rejection_details,omitempty"`
    ApprovedBy                   *string    `db:"approved_by"                    json:"approved_by,omitempty"`
    ApprovedAt                   *time.Time `db:"approved_at"                    json:"approved_at,omitempty"`
    CreatedAt                    time.Time  `db:"created_at"                     json:"created_at"`
    UpdatedAt                    time.Time  `db:"updated_at"                     json:"updated_at"`
}

// Enum nilai yang valid untuk field Status.
const (
    RequestStatusPending   = "pending"
    RequestStatusCompleted = "completed"
    RequestStatusRejected  = "rejected"
    RequestStatusFailed    = "failed"
    RequestStatusCancelled = "cancelled"
)

// Enum nilai yang valid untuk field PaymentMethod.
const (
    PaymentMethodManual   = "manual"
    PaymentMethodStripe   = "stripe"
    PaymentMethodMidtrans = "midtrans"
)

// Enum nilai yang valid untuk field RejectionReason.
const (
    RejectionReasonAmountMismatch = "Amount doesn't match"
    RejectionReasonInvalidProof   = "Invalid proof"
    RejectionReasonDuplicate      = "Duplicate request"
    RejectionReasonOther          = "Other"
)

// SubscriptionRequestWithUser digunakan di backoffice saat response
// perlu menyertakan detail user di dalam request.
type SubscriptionRequestWithUser struct {
    SubscriptionRequest
    User *RequestUserInfo `db:"-" json:"user,omitempty"`
}

// RequestUserInfo adalah subset info user untuk ditampilkan
// di dalam subscription request (backoffice).
type RequestUserInfo struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Email     string  `json:"email"`
    Phone     *string `json:"phone,omitempty"`
    Region    string  `json:"region"`
    CreatedAt string  `json:"created_at"`
}
```

---

### Struct: `SubscriptionHistory`

Memetakan tabel `subscription_history`.

```go
// internal/domain/subscription/subscription_history.go

package subscription

import "time"

// SubscriptionHistory adalah audit trail perubahan subscription.
// Tabel ini append-only — tidak ada UPDATE atau DELETE.
type SubscriptionHistory struct {
    ID                  string    `db:"id"                    json:"id"`
    UserID              string    `db:"user_id"               json:"user_id"`
    OldPlanID           *string   `db:"old_plan_id"           json:"old_plan_id,omitempty"`
    NewPlanID           string    `db:"new_plan_id"           json:"new_plan_id"`
    Action              string    `db:"action"                json:"action"`
    InitiatedBy         string    `db:"initiated_by"          json:"initiated_by"`
    InitiatedByUserID   *string   `db:"initiated_by_user_id"  json:"initiated_by_user_id,omitempty"`
    RequestID           *string   `db:"request_id"            json:"request_id,omitempty"`
    Notes               *string   `db:"notes"                 json:"notes,omitempty"`
    CreatedAt           time.Time `db:"created_at"            json:"created_at"`
}

// Enum nilai yang valid untuk field Action.
const (
    HistoryActionSignup    = "signup"
    HistoryActionUpgrade   = "upgrade"
    HistoryActionDowngrade = "downgrade"
    HistoryActionRenewal   = "renewal"
    HistoryActionExpired   = "expired"
    HistoryActionCancelled = "cancelled"
)

// Enum nilai yang valid untuk field InitiatedBy.
const (
    InitiatedByUser   = "user"
    InitiatedByAdmin  = "admin"
    InitiatedBySystem = "system"
)

// SubscriptionHistoryWithPlans digunakan saat response API perlu
// menyertakan nama plan lama dan baru (bukan hanya ID-nya).
type SubscriptionHistoryWithPlans struct {
    SubscriptionHistory
    OldPlan *Plan `db:"-" json:"old_plan,omitempty"`
    NewPlan *Plan `db:"-" json:"new_plan,omitempty"`
}
```

---

### Struct: `Refund`

Memetakan tabel `refunds`.

```go
// internal/domain/subscription/refund.go

package subscription

import "time"

// Refund merepresentasikan satu proses refund manual oleh superadmin.
type Refund struct {
    ID             string    `db:"id"             json:"id"`
    UserID         string    `db:"user_id"        json:"user_id"`
    RequestID      *string   `db:"request_id"     json:"request_id,omitempty"`
    SubscriptionID *string   `db:"subscription_id" json:"subscription_id,omitempty"`
    RefundAmount   int64     `db:"refund_amount"  json:"refund_amount"`
    Currency       string    `db:"currency"       json:"currency"`
    Reason         string    `db:"reason"         json:"reason"`
    Notes          *string   `db:"notes"          json:"notes,omitempty"`
    OldPlanID      string    `db:"old_plan_id"    json:"old_plan_id"`
    NewPlanID      string    `db:"new_plan_id"    json:"new_plan_id"`
    RefundStatus   string    `db:"refund_status"  json:"refund_status"`
    ProcessedBy    string    `db:"processed_by"   json:"processed_by"`
    RefundDate     time.Time `db:"refund_date"    json:"refund_date"`
    CreatedAt      time.Time `db:"created_at"     json:"created_at"`
}

// Enum nilai yang valid untuk field Reason.
const (
    RefundReasonUserRequest  = "User request"
    RefundReasonPaymentError = "Payment error"
    RefundReasonServiceIssue = "Service issue"
    RefundReasonOther        = "Other"
)

// Enum nilai yang valid untuk field RefundStatus.
const (
    RefundStatusProcessed = "processed"
    RefundStatusPending   = "pending"
    RefundStatusFailed    = "failed"
)
```

---

### Webhook Flow (Stripe / Midtrans)

Saat webhook masuk dan pembayaran sukses, jalankan dalam satu **database transaction**:

```
1. Validasi signature webhook (X-Stripe-Signature / Midtrans signature)
2. Cari subscription_request berdasarkan payment_provider_transaction_id
3. UPDATE subscription_requests SET status = 'completed', approved_at = NOW()
4. UPSERT subscriptions    → update plan_id, period, status = 'active'
5. INSERT subscription_history → action = 'upgrade', initiated_by = 'system'
6. Kirim notifikasi ke user (push + email)
```

### Partial Unique Index — Satu Pending Request per User

Index berikut memastikan constraint langsung di level database, tanpa cek manual di application layer:

```sql
CREATE UNIQUE INDEX idx_sub_requests_one_pending
  ON subscription_requests (user_id)
  WHERE status = 'pending';
```

Jika user submit request kedua saat masih ada yang pending, PostgreSQL langsung throw `unique_violation` — aplikasi tinggal tangkap error ini dan return `409 Conflict`.

---

*Dokumen ini adalah schema & domain struct reference. Service layer, repository interface, dan migration files dibuat terpisah.*
