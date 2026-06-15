# API Spec — ID Card Backoffice (v2.0)

API specification untuk ID Card module KAI App — **backoffice surface** (admin & superadmin).

Versi decoupled dari Subscription. Revoke hanya untuk alasan identitas/keanggotaan, bukan pembayaran.

Base URL: `/api/v1/admin`
Auth: Bearer token (Admin Regional → region sendiri; Superadmin → semua region)

---

## 1. GET /id-cards
List & filter kartu.

**Auth:** Required (Admin region sendiri / Superadmin semua)

**Query Parameters:**
- `region` — filter by region (Admin Regional otomatis ter-scope ke region sendiri)
- `status` — `active` | `revoked`
- `q` — search nama user
- `limit` (default 20, max 50), `offset` (default 0)

**Response (200 OK):**
```json
{
  "data": [
    {
      "card_id": "KAI-2026-001",
      "user_name": "John Doe",
      "region_id": "reg_jakarta",
      "status": "active",
      "issued_date": "2026-05-20",
      "physical_ordered": true
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 1 }
}
```
> Filter `plan` TIDAK tersedia di sini — plan bukan atribut kartu. Untuk filter per plan, gunakan modul Subscription.

---

## 2. GET /id-cards/:card_id
Detail satu kartu.

**Auth:** Required (Admin region sendiri / Superadmin)

**Response (200 OK):**
```json
{
  "data": {
    "card_id": "KAI-2026-001",
    "user_id": "usr_123",
    "user_name": "John Doe",
    "region_id": "reg_jakarta",
    "status": "active",
    "qr_version": "v1",
    "issued_date": "2026-05-20T00:00:00.000Z",
    "physical_ordered": true,
    "order_id": "ORD-001",
    "revoked_reason": null,
    "revoked_at": null,
    "created_at": "2026-05-20T00:00:00.000Z",
    "updated_at": "2026-05-20T00:00:00.000Z"
  }
}
```

---

## 3. POST /id-cards/:card_id/revoke
Revoke kartu. **Hanya untuk alasan identitas/keanggotaan** (user banned, akun dihapus, keluar dari KAI). Bukan untuk subscription lapse/downgrade/telat bayar.

**Auth:** Required (Admin region sendiri / Superadmin)

**Request Body:**
```json
{ "reason": "User banned" }
```

**Response (200 OK):**
```json
{
  "data": {
    "card_id": "KAI-2026-001",
    "status": "revoked",
    "revoked_reason": "User banned",
    "revoked_at": "2026-05-22T10:00:00.000Z"
  }
}
```

**Response (400 — alasan tidak valid):**
```json
{ "message": "Revocation reason must be identity-related, not payment-related" }
```

> Revocation bersifat permanen. Tidak ada status `expired` — identitas tidak kadaluarsa.

---

## 4. GET /id-cards/:card_id/scan-history
Riwayat scan kartu (audit & fraud detection).

**Auth:** Required (Admin region sendiri / Superadmin)

**Query Parameters:** `context`, `result`, `limit`, `offset`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "scan_001",
      "scanned_by": "merchant_045",
      "context": "directory",
      "context_ref": "mch_045",
      "result": "valid",
      "metadata": { "plan_at_scan": "pro", "benefit_applied": "directory_discount" },
      "created_at": "2026-05-21T14:30:00.000Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 1 }
}
```

---

## 5. GET /id-cards/scan-analytics
Agregat scan untuk dashboard (usage, redemption, fraud signals).

**Auth:** Required (Admin region → region sendiri; Superadmin → global)

**Query Parameters:** `region`, `context`, `from`, `to`

**Response (200 OK):**
```json
{
  "data": {
    "total_scans": 1240,
    "valid_scans": 1198,
    "failed_scans": 42,
    "by_context": {
      "directory": 890,
      "event": 310,
      "public_share": 40
    },
    "top_merchants": [
      { "merchant_id": "mch_045", "redemptions": 230 }
    ],
    "fraud_signals": [
      { "card_id": "KAI-2026-009", "distinct_scanners_5min": 3 }
    ]
  }
}
```

---

## CATATAN ARSITEKTUR
1. Kartu tidak menyimpan plan — semua plan/benefit di-resolve live dari Subscription.
2. Revoke hanya alasan identitas. Subscription lapse/downgrade tidak mempengaruhi validitas kartu.
3. Setiap scan dicatat ke `card_scan_events` — fondasi untuk attendance, redemption, loyalty, fraud detection.
