# KAI Subscription API Spec — Mobile Client

Dokumentasi API endpoint untuk subscription management di mobile client (Flutter).

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/subscription`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Accept-Language: <lang_code>` (contoh: `ko`, `id`, `en`. Default: `ko`)
  - `X-Locale: <lang_code>` (contoh: `ko`, `id`, `en`. Default: `ko`)
  - `Authorization: Bearer <access_token>` (untuk endpoint yang membutuhkan autentikasi)

---

## 1. GET /subscription/me

**Description**: Ambil informasi subscription current user

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/mobile/subscription/me`

**Request**: None

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
      "status": "pending",
      "submitted_at": "2026-05-24T10:00:00.000Z",
      "amount": 80000
    }
  }
}
```

**Response (404 Not Found)**:
```json
{
  "message": "Subscription not found"
}
```

**Response (401 Unauthorized)**:
```json
{
  "message": "Authentication required"
}
```

---

## 2. GET /subscription/plans

**Description**: Ambil daftar semua plan yang tersedia

**Authentication**: Not required

**Method**: GET

**URL**: `/api/v1/mobile/subscription/plans`

**Query Parameters**: None

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "plan_standard",
      "name": "Standard",
      "price": 49000,
      "currency": "IDR",
      "duration_days": 30,
      "description": "Akses fitur utama: baca, join komunitas, tanya QnA",
      "is_default": true,
      "status": "active",
      "benefits": [
        {
          "key": "read_content",
          "description": "Baca news, event, direktori",
          "enabled": true
        },
        {
          "key": "join_community",
          "description": "Join komunitas",
          "enabled": true
        },
        {
          "key": "post_community",
          "description": "Post di komunitas (unlimited)",
          "enabled": true
        },
        {
          "key": "ask_qna",
          "description": "Tanya di QnA",
          "enabled": true
        }
      ]
    },
    {
      "id": "plan_pro",
      "name": "Pro",
      "price": 129000,
      "currency": "IDR",
      "duration_days": 30,
      "description": "Standard + posting news (approval), buat community, buat company/merchants, buat event",
      "is_default": false,
      "status": "active",
      "benefits": [
        {
          "key": "post_news",
          "description": "Post news (dengan approval)",
          "enabled": true
        },
        {
          "key": "create_community",
          "description": "Buat dan manage community",
          "enabled": true
        },
        {
          "key": "create_store",
          "description": "Buat store/merchant listing",
          "enabled": true
        },
        {
          "key": "create_event",
          "description": "Buat event",
          "enabled": true
        },
        {
          "key": "view_analytics",
          "description": "Lihat analytics",
          "enabled": true
        }
      ]
    }
  ]
}
```

---

## 3. POST /subscription/request

**Description**: Create upgrade request ke plan lain (manual verification atau payment gateway)

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/mobile/subscription/request`

> **Manual proof upload**: Gunakan presigned media endpoint (`POST /api/v1/mobile/media/presign` → upload → `POST /api/v1/mobile/media/confirm`) terlebih dahulu untuk upload bukti transfer. Hasilnya berupa `s3:` key yang dikirim sebagai `manual_proof_key`.

**Request Body**:
```json
{
  "requested_plan": "pro",
  "payment_method": "manual" | "stripe" | "midtrans",
  "manual_proof_key": "s3:/proofs/manual_upload_abc.jpg" | null
}
```

**Validation**:
- `requested_plan`: required, harus ada di plans
- `payment_method`: required, harus valid
- Jika `payment_method` === 'manual': `manual_proof_key` boleh null tapi recommended (format `s3:`)
- Jika `manual_proof_key` diisi, harus dimulai dengan `s3:`
- User tidak boleh punya pending request

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
    "payment_url": null,
    "submitted_at": "2026-05-24T10:00:00.000Z"
  },
  "message": "Upgrade request submitted. Admin akan verifikasi dalam 1x24 jam."
}
```

**Response (Payment Gateway - Stripe)**:
```json
{
  "data": {
    "request_id": "req_yyy",
    "current_plan": "standard",
    "requested_plan": "pro",
    "amount": 80000,
    "status": "processing",
    "payment_method": "stripe",
    "payment_url": "https://checkout.stripe.com/pay/...",
    "submitted_at": "2026-05-24T10:00:00.000Z"
  },
  "message": "Redirect ke Stripe checkout"
}
```

**Response (409 Conflict - Already have pending request)**:
```json
{
  "message": "You already have a pending upgrade request",
  "data": {
    "existing_request_id": "req_xxx",
    "status": "pending"
  }
}
```

**Response (400 Bad Request - Invalid plan)**:
```json
{
  "message": "Invalid plan requested",
  "errors": {
    "requested_plan": ["Plan not found or not available"]
  }
}
```

**Response (422 Unprocessable Entity)**:
```json
{
  "message": "Request validation failed",
  "errors": {
    "payment_method": ["Payment method tidak valid"]
  }
}
```

---

## 4. GET /subscription/request/:request_id

**Description**: Ambil detail subscription request

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/mobile/subscription/request/:request_id`

**Request**: None

**Response (200 OK)**:
```json
{
  "data": {
    "id": "req_xxx",
    "user_id": "user_123",
    "current_plan": "standard",
    "requested_plan": "pro",
    "amount": 80000,
    "status": "pending" | "completed" | "rejected" | "failed",
    "payment_method": "manual" | "stripe" | "midtrans",
    "payment_provider_transaction_id": "txn_xxx" | null,
    "manual_proof_url": "https://..." | null,
    "verified_by_admin": false | true,
    "rejection_reason": "Amount doesn't match" | null,
    "created_at": "2026-05-24T10:00:00.000Z",
    "approved_at": "2026-05-24T10:05:00.000Z" | null
  }
}
```

**Response (404 Not Found)**:
```json
{
  "message": "Request not found"
}
```

---

## 5. POST /subscription/request/:request_id/cancel

**Description**: Cancel pending upgrade request

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/mobile/subscription/request/:request_id/cancel`

**Request Body**: None

**Response (200 OK)**:
```json
{
  "message": "Upgrade request cancelled",
  "data": {
    "request_id": "req_xxx",
    "status": "cancelled"
  }
}
```

**Response (409 Conflict - Can't cancel non-pending request)**:
```json
{
  "message": "Cannot cancel non-pending request",
  "data": {
    "current_status": "completed"
  }
}
```

---

## 6. POST /subscription/downgrade

**Description**: Downgrade plan (e.g., Pro → Standard). Efektif di expiry date.

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/mobile/subscription/downgrade`

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
{
  "message": "Cannot downgrade from Standard plan"
}
```

---

## 7. POST /subscription/cancel-renewal

**Description**: Cancel renewal otomatis. Subscription akan expire di period_end.

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/mobile/subscription/cancel-renewal`

**Request Body**: None

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

**Description**: Ambil subscription history user saat ini

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/mobile/subscription/history`

**Query Parameters**:
- `limit`: 10 (default), max 50
- `offset`: 0 (default)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "hist_xxx",
      "old_plan": null,
      "new_plan": "standard",
      "action": "signup",
      "initiated_by": "system",
      "created_at": "2026-05-01T00:00:00.000Z"
    },
    {
      "id": "hist_yyy",
      "old_plan": "standard",
      "new_plan": "pro",
      "action": "upgrade",
      "initiated_by": "user",
      "created_at": "2026-05-10T10:00:00.000Z"
    },
    {
      "id": "hist_zzz",
      "old_plan": "pro",
      "new_plan": "pro",
      "action": "renewal",
      "initiated_by": "system",
      "created_at": "2026-06-10T00:00:00.000Z"
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0,
    "total": 3
  }
}
```

---

## 9. POST /subscription/verify-benefit

**Description**: Verify apakah user punya akses ke benefit tertentu. Untuk early check di frontend sebelum navigate.

**Authentication**: Required (Bearer token)

**Method**: POST

**URL**: `/api/v1/mobile/subscription/verify-benefit`

**Request Body**:
```json
{
  "benefit_key": "post_news" | "create_community" | "create_store" | "create_event" | "view_analytics"
}
```

**Response (200 OK - Has benefit)**:
```json
{
  "data": {
    "benefit_key": "post_news",
    "has_access": true,
    "plan": "pro"
  }
}
```

**Response (403 Forbidden - No access)**:
```json
{
  "data": {
    "benefit_key": "post_news",
    "has_access": false,
    "plan": "standard",
    "required_plan": "pro",
    "message": "Upgrade to Pro to post news"
  }
}
```

---

## 10. GET /subscription/benefits

**Description**: Ambil daftar benefit yang dimiliki user saat ini

**Authentication**: Required (Bearer token)

**Method**: GET

**URL**: `/api/v1/mobile/subscription/benefits`

**Response (200 OK)**:
```json
{
  "data": {
    "plan": "pro",
    "benefits": [
      {
        "key": "read_content",
        "description": "Baca news, event, direktori",
        "enabled": true
      },
      {
        "key": "join_community",
        "description": "Join komunitas",
        "enabled": true
      },
      {
        "key": "post_community",
        "description": "Post di komunitas (unlimited)",
        "enabled": true
      },
      {
        "key": "ask_qna",
        "description": "Tanya di QnA",
        "enabled": true
      },
      {
        "key": "post_news",
        "description": "Post news (dengan approval)",
        "enabled": true
      },
      {
        "key": "create_community",
        "description": "Buat dan manage community",
        "enabled": true
      },
      {
        "key": "create_store",
        "description": "Buat store/merchant listing",
        "enabled": true
      },
      {
        "key": "create_event",
        "description": "Buat event",
        "enabled": true
      },
      {
        "key": "view_analytics",
        "description": "Lihat analytics",
        "enabled": true
      }
    ]
  }
}
```

---

## 11. POST /subscription/webhook/stripe — ⚠️ DEPRECATED

> **DEPRECATED:** Stripe tidak lagi menjadi gateway aktif untuk KAI. Payment gateway resmi adalah **Midtrans** (lihat section 12). Endpoint ini dipertahankan hanya untuk referensi histori dan tidak diimplementasikan. Jangan dipakai untuk integrasi baru.

**Description**: Stripe webhook callback (automatic, called by Stripe)

**Authentication**: Required (Stripe signature verification via `X-Stripe-Signature` header)

**Method**: POST

**URL**: `/api/v1/mobile/subscription/webhook/stripe`

**Headers**:
- `X-Stripe-Signature: <signature>`

**Request Body** (example):
```json
{
  "id": "evt_xxx",
  "type": "charge.succeeded" | "charge.failed",
  "data": {
    "object": {
      "id": "ch_xxx",
      "amount": 8000000,
      "currency": "idr",
      "metadata": {
        "request_id": "req_xxx",
        "user_id": "user_123"
      }
    }
  }
}
```

**Response (200 OK)**:
```json
{
  "message": "Webhook processed successfully"
}
```

**Response (400 Bad Request - Invalid signature)**:
```json
{
  "message": "Invalid signature"
}
```

---

## 12. POST /subscription/webhook/midtrans

**Description**: Midtrans HTTP notification (payment callback). Dipanggil otomatis oleh server Midtrans setiap kali status transaksi berubah. Berlaku untuk mode **Snap maupun Core API** — payload notification-nya sama.

**Authentication**: Verifikasi `signature_key` (SHA-512 dari `order_id + status_code + gross_amount + ServerKey`). Request yang signature-nya tidak cocok harus ditolak `403`.

**Method**: POST

**URL**: `/api/v1/mobile/subscription/webhook/midtrans`

**Request Body** (contoh — settlement sukses):
```json
{
  "transaction_time": "2026-05-24 17:00:00",
  "transaction_status": "settlement",
  "transaction_id": "9aed5972-...",
  "status_message": "midtrans payment notification",
  "status_code": "200",
  "signature_key": "abc123...",
  "payment_type": "gopay",
  "order_id": "req_xxx",
  "merchant_id": "G12345",
  "gross_amount": "80000.00",
  "fraud_status": "accept",
  "currency": "IDR"
}
```

> **`order_id` = `subscription_requests.id`** (atau referensi unik yang di-generate saat create request). Backend mencari request berdasarkan field ini, bukan `transaction_id`.

**Mapping `transaction_status` → status internal:**

| Midtrans `transaction_status` | `fraud_status` | Aksi internal |
|---|---|---|
| `capture` | `accept` | request → `completed`, aktifkan subscription |
| `settlement` | — | request → `completed`, aktifkan subscription |
| `pending` | — | request tetap `pending` (belum bayar) |
| `deny` | — | request → `failed` |
| `cancel` / `expire` | — | request → `failed` |
| `capture` | `challenge` | request tetap `pending`, tandai untuk review manual |
| `refund` / `partial_refund` | — | trigger flow refund (lihat backoffice spec) |

**Processing (dalam satu DB transaction):**
```
1. Verifikasi signature_key (SHA-512). Mismatch → 403, jangan proses.
2. Cari subscription_request by order_id. Tidak ada → 404 (Midtrans akan retry).
3. Idempotency: jika request sudah completed/failed, balikin 200 OK tanpa proses ulang.
4. Map transaction_status → status internal (tabel di atas).
5. Jika completed:
   - UPDATE subscription_requests: status='completed',
     payment_provider_transaction_id=transaction_id, approved_at=NOW()
   - UPSERT subscriptions: plan_id baru, period start/end, status='active'
   - Tandai subscription lama 'expired'/'superseded'
   - INSERT subscription_history: action='upgrade', initiated_by='system'
   - (future hook) push accounting entry: source_ref=order_id, payment_provider='midtrans'
   - Kirim email + in-app notif: "Pembayaran berhasil, Pro aktif"
6. Jika failed: UPDATE status='failed'. Kirim notif gagal.
```

**Response (200 OK)** — selalu balikin 200 jika sudah diproses/diabaikan dengan benar (Midtrans retry jika non-200):
```json
{
  "message": "Notification processed"
}
```

**Response (403 Forbidden — invalid signature)**:
```json
{
  "message": "Invalid signature"
}
```

> **Catatan:** Webhook ini shared antara mobile dan web — keduanya membuat `subscription_requests` yang sama dan Midtrans memanggil URL notification yang dikonfigurasi di dashboard Midtrans (satu URL). Detail pembuatan transaksi (Snap token vs Core API) ada di `API_SPEC_SUBSCRIPTION_WEB.md`.

---

## Error Response Standard

### Format A: Standard Message Error (4xx / 5xx)
```json
{
  "message": "Pesan error deskriptif di sini"
}
```

### Format B: Validation Error (422)
```json
{
  "message": "Data input tidak valid",
  "errors": {
    "field_name": ["Error message 1", "Error message 2"],
    "another_field": ["Error message"]
  }
}
```

Frontend akan mengambil error pertama dari array untuk ditampilkan.

---

## Status Codes

- `200 OK` - Sukses
- `201 Created` - Resource created
- `400 Bad Request` - Validation error atau bad request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Access denied
- `404 Not Found` - Resource not found
- `409 Conflict` - Conflict (e.g., already have pending request)
- `422 Unprocessable Entity` - Validation error dengan detail per field
- `500 Internal Server Error` - Server error

---

## Rate Limiting

- Limit: 100 requests per minute per user
- Headers:
  - `X-RateLimit-Limit: 100`
  - `X-RateLimit-Remaining: 95`
  - `X-RateLimit-Reset: 1629856800`

---

## Caching Strategy

- `GET /subscription/me` - Cache 5 minutes (user dapat melihat perubahan dalam 5 menit)
- `GET /subscription/plans` - Cache 1 hour (plans jarang berubah)
- `GET /subscription/benefits` - Cache 5 minutes
- Lainnya: No cache

---

END OF DOCUMENT
