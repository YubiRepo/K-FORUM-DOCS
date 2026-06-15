# API Spec — ID Card Mobile (v2.0)

API specification untuk ID Card module KAI App — **mobile surface** (member-facing).

Versi decoupled dari Subscription: kartu = identitas, plan/benefit di-resolve live saat verifikasi. Kartu tidak menyimpan `plan` maupun `expiry_date`.

Base URL: `/api/v1/mobile`
Auth: Bearer token (hanya pemilik kartu)

---

## 1. GET /members/:user_id/id-card
Ambil ID card user.

**Auth:** Required (hanya pemilik)

**Response (200 OK):**
```json
{
  "data": {
    "card_id": "KAI-2026-001",
    "user_id": "usr_123",
    "region_id": "reg_jakarta",
    "issued_date": "2026-05-20T00:00:00.000Z",
    "status": "active",
    "qr_version": "v1",
    "qr_code_data": "v1|KAI-2026-001|a1b2c3d4e5f6g7h8",
    "digital_format": {
      "design": "standard-2026",
      "user_name": "John Doe",
      "avatar_url": "https://cdn.kai.app/avatars/usr_123.jpg",
      "qr_url": "https://kai.app/verify/KAI-2026-001"
    },
    "physical_ordered": false
  }
}
```
> Catatan: tidak ada `plan` di response. Plan adalah konsep Subscription — ambil dari `GET /mobile/subscription` bila perlu ditampilkan di UI kartu.

**Response (403 — bukan pemilik):**
```json
{ "message": "You can only access your own ID card" }
```

---

## 2. GET /members/:user_id/id-card/download-pdf
Download kartu digital sebagai PDF (self-contained, bisa dipakai offline).

**Auth:** Required (pemilik)

**Response:** Binary PDF (`Content-Type: application/pdf`)

---

## 3. POST /members/:user_id/id-card/request-physical
Request kartu fisik dikirim ke alamat.

**Auth:** Required (pemilik)

**Request Body:**
```json
{
  "shipping_address": "Jl. Sudirman No. 1, Jakarta",
  "phone": "+628123456789"
}
```

**Response (200 OK):**
```json
{
  "data": {
    "order_id": "ORD-001",
    "estimated_delivery": "2026-05-30",
    "status": "processing"
  }
}
```

**Response (409 — sudah pernah order):**
```json
{ "message": "Physical card already ordered", "data": { "order_id": "ORD-001" } }
```

---

## 4. POST /members/:user_id/id-card/request-replacement
Request kartu pengganti (hilang/rusak). Generate `card_id` baru dengan `user_id` sama. Maks 2x/tahun.

**Auth:** Required (pemilik)

**Request Body:**
```json
{
  "reason": "lost",
  "shipping_address": "Jl. Sudirman No. 1, Jakarta"
}
```
> `reason`: `lost` | `damaged` | `other`

**Response (200 OK):**
```json
{
  "data": {
    "new_card_id": "KAI-2026-001R1",
    "old_card_id": "KAI-2026-001",
    "replacements_used": 1,
    "replacements_remaining": 1
  }
}
```

**Response (429 — limit tercapai):**
```json
{ "message": "Replacement limit reached (max 2 per year)" }
```

---

## CATATAN
- Kartu **tidak menyimpan plan** — untuk menampilkan badge plan di UI, query modul Subscription terpisah.
- QR pada kartu digital ber-versi (`v1|...`).
- Endpoint verifikasi (scan) ada di dokumen **Public/Partner** terpisah.
