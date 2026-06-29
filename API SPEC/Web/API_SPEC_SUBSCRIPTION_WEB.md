# KAI Subscription API Spec — Web Client (Member)

Dokumentasi API endpoint untuk subscription management di **web client** (Nuxt member-facing web). Member bisa lihat plans, upgrade, downgrade, cancel renewal, lihat history, dan bayar via **Midtrans** langsung dari browser.

> **Hubungan dengan spec lain:**
> - `API_SPEC_Subscription_Mobile.md` — endpoint setara untuk Flutter (prefix `/mobile`). Logika bisnis & response shape identik; yang beda hanya prefix dan detail return URL pembayaran.
> - Webhook Midtrans **shared** dengan mobile (satu URL notification dikonfigurasi di dashboard Midtrans). Lihat section 12.
> - `API_SPEC_PLAN_SUBSCRIPTION.md` — endpoint backoffice (admin verify request, refund, plan management).
> - Mode integrasi (`snap`/`core`), environment (`sandbox`/`production`), dan channel aktif dikontrol dari **System Settings** group `payment` — bukan hardcode (lihat `SYSTEM_SETTINGS_RULES.md`).

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/subscription`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Accept-Language: <lang_code>` (contoh: `ko`, `id`, `en`. Default: `ko`)
  - `X-Locale: <lang_code>`
  - `Authorization: Bearer <access_token>` (untuk endpoint yang membutuhkan autentikasi)

### Catatan arsitektur pembayaran

Member memilih `payment_method` saat create request. Untuk `midtrans`, **bentuk response `POST /request` berbeda tergantung `midtrans_integration_mode` di System Settings**:

| Mode (setting) | Yang dibalikin backend | Yang dilakukan web frontend |
|---|---|---|
| `snap` | `snap_token` + `redirect_url` | Buka Snap popup via `snap.js` (pakai `snap_token`), atau redirect ke `redirect_url` (hosted page). Setelah selesai, redirect ke `finish_url` member. |
| `core` | `payment_instructions` per-channel (VA number / QRIS string / deeplink) | Render instruksi bayar di halaman sendiri (custom UI), polling status via `GET /request/:id` |

> **Sumber kebenaran tetap webhook.** Aktivasi subscription **hanya** terjadi dari notification Midtrans (section 12), bukan dari frontend. `finish_url`/`redirect_url` hanya untuk UX — frontend tetap konfirmasi status via `GET /request/:id`.

---

## 1. GET /subscription/me

**Description**: Ambil informasi subscription current user.

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/web/subscription/me`

**Response (200 OK)**:
```json
{
  "data": {
    "plan": "standard" | "pro",
    "status": "active" | "expired" | "cancelled",
    "current_period_start": "2026-05-24",
    "current_period_end": "2026-06-24",
    "auto_renewal": true,
    "renewal_reminder_sent": false,
    "pending_request": {
      "id": "req_xxx" | null,
      "requested_plan": "pro",
      "status": "pending" | "processing",
      "submitted_at": "2026-05-24T10:00:00.000Z",
      "amount": 80000
    }
  }
}
```

**Response (404 / 401)**: sama seperti mobile spec.

---

## 2. GET /subscription/plans

**Description**: Ambil daftar semua plan yang tersedia (untuk pricing page web).

**Authentication**: Not required

**Method**: GET

**URL**: `/api/v1/web/subscription/plans`

**Response (200 OK)**: identik dengan mobile spec — array plan dengan `id`, `name`, `price`, `currency`, `duration_days`, `description`, `is_default`, `status`, dan `benefits[]` (`key`, `description`, `enabled`).

---

## 3. POST /subscription/request

**Description**: Create upgrade request. Mendukung `manual` dan `midtrans`. Untuk midtrans, response menyesuaikan `midtrans_integration_mode`.

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/web/subscription/request`

**Request Body**:
```json
{
  "requested_plan": "pro",
  "payment_method": "manual" | "midtrans",
  "manual_proof_url": "https://..." | null,
  "manual_proof_type": "image" | "video" | null
}
```

**Validation**:
- `requested_plan`: required, harus ada di plans dan `status = active`
- `payment_method`: required; nilai valid mengikuti `payment_provider` di System Settings (`manual` saja, `midtrans` saja, atau keduanya bila `both`)
- Jika `payment_method = manual`: `manual_proof_url` boleh null tapi recommended
- User tidak boleh punya pending/processing request lain
- Channel yang dikirim ke Midtrans dibatasi oleh `midtrans_enabled_channels`

### Response 3A — `payment_method = manual`

**Response (201 Created)**:
```json
{
  "data": {
    "request_id": "req_xxx",
    "current_plan": "standard",
    "requested_plan": "pro",
    "amount": 80000,
    "status": "pending",
    "payment_method": "manual",
    "submitted_at": "2026-05-24T10:00:00.000Z"
  },
  "message": "Upgrade request submitted. Admin akan verifikasi dalam 1x24 jam."
}
```

### Response 3B — `payment_method = midtrans`, mode `snap`

**Response (201 Created)**:
```json
{
  "data": {
    "request_id": "req_yyy",
    "current_plan": "standard",
    "requested_plan": "pro",
    "amount": 80000,
    "status": "processing",
    "payment_method": "midtrans",
    "midtrans_mode": "snap",
    "order_id": "req_yyy",
    "snap_token": "66e4fa55-fdac-4ef9-91b5-733b97d1b862",
    "redirect_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/66e4fa55-...",
    "expires_at": "2026-05-24T11:00:00.000Z"
  },
  "message": "Lanjutkan pembayaran via Midtrans Snap"
}
```

> Web frontend: panggil `window.snap.pay(snap_token, { onSuccess, onPending, onError, onClose })`, atau redirect ke `redirect_url`. `order_id = request_id`.

### Response 3C — `payment_method = midtrans`, mode `core`

**Response (201 Created)** — contoh channel `bank_transfer` (VA):
```json
{
  "data": {
    "request_id": "req_zzz",
    "current_plan": "standard",
    "requested_plan": "pro",
    "amount": 80000,
    "status": "processing",
    "payment_method": "midtrans",
    "midtrans_mode": "core",
    "order_id": "req_zzz",
    "transaction_id": "9aed5972-3a3e-...",
    "payment_instructions": {
      "channel": "bank_transfer",
      "bank": "bca",
      "va_number": "12345678901",
      "qris_string": null,
      "deeplink_url": null,
      "expires_at": "2026-05-24T11:00:00.000Z"
    }
  },
  "message": "Selesaikan pembayaran sesuai instruksi"
}
```

Contoh channel lain pada mode `core`:
- `qris` → `payment_instructions.qris_string` berisi payload QR (frontend render jadi QR image), `va_number` null.
- `gopay` → `payment_instructions.deeplink_url` (buka di app GoPay / scan QR), `va_number` null.

**Response (409 Conflict — Already have pending request)**:
```json
{
  "message": "You already have a pending upgrade request",
  "data": { "existing_request_id": "req_xxx", "status": "pending" }
}
```

**Response (400 — gateway disabled)**:
```json
{
  "message": "Midtrans payment is not enabled",
  "data": { "payment_provider": "manual" }
}
```

**Response (422)**: validation error per-field (format standar).

---

## 4. GET /subscription/request/:request_id

**Description**: Ambil detail request. Web frontend memakai ini untuk **polling status** setelah user kembali dari Midtrans (mode snap) atau saat menunggu pembayaran (mode core).

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/web/subscription/request/:request_id`

**Response (200 OK)**:
```json
{
  "data": {
    "id": "req_xxx",
    "user_id": "user_123",
    "current_plan": "standard",
    "requested_plan": "pro",
    "amount": 80000,
    "status": "pending" | "processing" | "completed" | "rejected" | "failed",
    "payment_method": "manual" | "midtrans",
    "midtrans_mode": "snap" | "core" | null,
    "payment_provider_transaction_id": "txn_xxx" | null,
    "manual_proof_url": "https://..." | null,
    "verified_by_admin": false,
    "rejection_reason": null,
    "created_at": "2026-05-24T10:00:00.000Z",
    "approved_at": null
  }
}
```

> `status = completed` adalah sinyal final bahwa subscription sudah aktif (di-set oleh webhook). Frontend boleh berhenti polling saat `completed`/`failed`/`rejected`.

---

## 5. POST /subscription/request/:request_id/cancel

**Description**: Cancel pending request (sebelum bayar). Untuk request midtrans yang masih `processing`, ini juga membatalkan transaksi di sisi Midtrans (best-effort).

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/web/subscription/request/:request_id/cancel`

**Response (200 OK)**:
```json
{
  "message": "Upgrade request cancelled",
  "data": { "request_id": "req_xxx", "status": "cancelled" }
}
```

**Response (409 Conflict)**:
```json
{
  "message": "Cannot cancel non-pending request",
  "data": { "current_status": "completed" }
}
```

---

## 6. POST /subscription/downgrade

**Description**: Downgrade plan (Pro → Standard). Efektif di expiry date.

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/web/subscription/downgrade`

**Request Body**:
```json
{
  "requested_plan": "standard",
  "reason": "No longer need features" | "Too expensive" | "Other"
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "current_plan": "pro",
    "new_plan": "standard",
    "effective_date": "2026-06-24",
    "message": "Your plan will downgrade to Standard on 2026-06-24 (expiry date)"
  }
}
```

**Response (400 Bad Request)**:
```json
{ "message": "Cannot downgrade from Standard plan" }
```

---

## 7. POST /subscription/cancel-renewal

**Description**: Cancel renewal otomatis. Subscription expire di period_end.

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/web/subscription/cancel-renewal`

**Response (200 OK)**:
```json
{
  "data": {
    "plan": "pro",
    "period_end": "2026-06-24",
    "auto_renewal": false,
    "message": "Renewal cancelled. Your Pro plan will expire on 2026-06-24"
  }
}
```

---

## 8. GET /subscription/history

**Description**: Subscription history user saat ini.

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/web/subscription/history`

**Query Parameters**: `limit` (10 default, max 50), `offset` (0 default)

**Response (200 OK)**: array history (`id`, `old_plan`, `new_plan`, `action`, `initiated_by`, `created_at`) + `pagination`. Identik dengan mobile spec.

---

## 9. POST /subscription/verify-benefit

**Description**: Cek apakah user punya akses ke benefit tertentu (early check sebelum navigate).

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/web/subscription/verify-benefit`

**Request Body**:
```json
{ "benefit_key": "post_news" | "create_community" | "create_store" | "create_event" | "view_analytics" }
```

**Response (200 OK — has access)** / **(403 — no access)**: identik dengan mobile spec.

---

## 10. GET /subscription/benefits

**Description**: Daftar benefit yang dimiliki user saat ini.

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/web/subscription/benefits`

**Response (200 OK)**: `{ "data": { "plan": "...", "benefits": [ { "key", "description", "enabled" } ] } }` — identik dengan mobile spec.

---

## 11. GET /subscription/payment-config

**Description**: Ambil konfigurasi pembayaran aktif agar web frontend bisa branch UI **tanpa hardcode**. Membaca System Settings group `payment`. Tidak pernah mengembalikan secret (server key tetap di env).

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/web/subscription/payment-config`

**Response (200 OK)**:
```json
{
  "data": {
    "payment_provider": "manual" | "midtrans" | "both",
    "midtrans": {
      "enabled": true,
      "environment": "sandbox" | "production",
      "integration_mode": "snap" | "core",
      "client_key": "SB-Mid-client-xxxxxxxx",
      "enabled_channels": ["gopay", "qris", "bank_transfer"],
      "snap_js_url": "https://app.sandbox.midtrans.com/snap/snap.js"
    },
    "manual": {
      "enabled": true,
      "bank_name": "BCA",
      "bank_account_number": "1234567890",
      "bank_account_holder": "PT KAI",
      "payment_instructions": "Transfer ke rekening di atas...",
      "confirmation_deadline_hours": 24
    }
  }
}
```

> `client_key` aman dipublish ke frontend (memang client-side). `snap_js_url` menyesuaikan environment (`app.sandbox...` vs `app.midtrans.com`). Jika `payment_provider = manual`, blok `midtrans.enabled = false`.

---

## 12. POST /subscription/webhook/midtrans

**Description**: Midtrans HTTP notification (payment callback). **Shared dengan mobile** — Midtrans hanya memanggil satu URL notification yang dikonfigurasi di dashboard. Detail mapping status, signature, dan processing **sama persis** dengan `API_SPEC_Subscription_Mobile.md` section 12.

**Method**: POST

**URL** (notification URL terdaftar di dashboard Midtrans): `/api/v1/web/subscription/webhook/midtrans` *atau* `/api/v1/mobile/subscription/webhook/midtrans` — implementasi backend boleh mengarahkan keduanya ke handler yang sama. Disarankan satu canonical URL, mis. `/api/v1/payments/midtrans/notification`.

**Authentication**: Verifikasi `signature_key` = SHA-512(`order_id + status_code + gross_amount + ServerKey`). Mismatch → `403`, jangan proses.

**Mapping status & processing**: identik dengan mobile spec section 12 (`settlement`/`capture+accept` → `completed` + aktifkan subscription; `deny`/`cancel`/`expire` → `failed`; `pending` → tetap pending; idempotent terhadap notifikasi berulang).

**Response (200 OK)** / **(403 invalid signature)**: sama dengan mobile spec.

---

## Error Response Standard

**Format A (4xx/5xx)**: `{ "message": "..." }`
**Format B (422)**: `{ "message": "...", "errors": { "field": ["..."] } }`

## Status Codes

`200`, `201`, `400`, `401`, `403`, `404`, `409`, `422`, `500` — sama dengan mobile spec.

## Caching Strategy

- `GET /subscription/me` — cache 5 menit
- `GET /subscription/plans` — cache 1 jam
- `GET /subscription/benefits` — cache 5 menit
- `GET /subscription/payment-config` — cache 5 menit (invalidate saat setting payment berubah)
- Lainnya: no cache

---

## Ringkasan perbedaan dari Mobile Spec

| Aspek | Mobile (`/mobile`) | Web (`/web`) |
|---|---|---|
| Gateway aktif | Midtrans (Stripe deprecated) | Midtrans |
| Payment flow | Snap via webview / Core API in-app | Snap popup/redirect di browser / Core API custom page |
| Endpoint payment-config | (dibaca dari System Settings mobile) | `GET /subscription/payment-config` (eksplisit) |
| Webhook | shared, satu handler | shared, satu handler |
| Business logic | identik | identik |
