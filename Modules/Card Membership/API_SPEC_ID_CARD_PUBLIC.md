# API Spec — ID Card Public / Partner (v2.0)

API specification untuk ID Card module KAI App — **public & partner surface** (verifikasi via scan QR).

Inti versi v2.0: kartu = identitas; **plan & benefit di-resolve LIVE** dari modul Subscription saat scan, bukan dibaca dari kartu. Setiap scan dicatat ke `card_scan_events`.

Base URL: `/api/v1/public`

---

## 1. POST /verify-card
Verifikasi kartu via scan QR. Return konteks identitas; plan/benefit di-resolve live (hanya untuk partner ter-autentikasi).

**Auth:** Optional
- **Dengan partner token** → dapat plan + active_benefits (resolve live)
- **Tanpa token** → hanya info publik (nama, region, status)

**Request Body:**
```json
{
  "qr_data": "v1|KAI-2026-001|a1b2c3d4e5f6g7h8",
  "context": "directory",
  "context_ref": "mch_045"
}
```
> `context`: `directory` | `event` | `public_share` | `loyalty`
> `context_ref`: opsional, mis. `merchant_id` atau `event_id`

**Response (200 OK — partner authenticated):**
```json
{
  "data": {
    "valid": true,
    "card_id": "KAI-2026-001",
    "user_id": "usr_123",
    "user_name": "John Doe",
    "region_id": "reg_jakarta",
    "status": "active",
    "plan": "pro",
    "active_benefits": ["directory_discount", "event_early_booking"]
  }
}
```
> `plan` & `active_benefits` di-resolve **live** dari Subscription saat scan. Kalau benefit Pro diubah superadmin, response ini otomatis ikut berubah tanpa cetak ulang kartu.

**Response (200 OK — public, tanpa token):**
```json
{
  "data": {
    "valid": true,
    "user_name": "John Doe",
    "region_id": "reg_jakarta",
    "status": "active"
  }
}
```
> Tanpa token: TIDAK ada plan/benefit/user_id (cegah enumerasi & kebocoran data).

**Response (200 OK — kartu di-revoke):**
```json
{ "data": { "valid": false, "status": "revoked" } }
```

**Response (400 — checksum tidak valid):**
```json
{ "data": { "valid": false, "reason": "invalid_checksum" } }
```

**Response (404 — kartu tidak ditemukan):**
```json
{ "data": { "valid": false, "reason": "not_found" } }
```

---

## 2. GET /verify/:card_id
Halaman verifikasi publik (untuk shared link, mis. `kai.app/verify/KAI-2026-001`).

**Auth:** Not required

**Response (200 OK):**
```json
{
  "data": {
    "valid": true,
    "user_name": "John Doe",
    "region_id": "reg_jakarta",
    "status": "active"
  }
}
```
> Hanya info publik. Rate-limited untuk cegah enumerasi.

---

## VERIFICATION RULES (urutan wajib di backend)
1. Parse versi QR → handle sesuai `qr_version`
2. Validasi checksum: `SHA256(version + card_id + secret_salt)`
3. Cek `status = active`
4. (Jika partner authenticated) resolve user → plan & benefit live dari Subscription
5. Catat ke `card_scan_events` — **selalu**, termasuk hasil gagal (`invalid_checksum`, `not_found`, `revoked`)

## SECURITY
- Rate-limit semua endpoint verifikasi (cegah enumerasi card_id).
- QR tidak memuat `plan` maupun `user_id` mentah — hanya `{version}|{card_id}|{checksum}`.
- Benefit bernilai tinggi (diskon, entry event) **wajib** backend call; jangan andalkan offline-only karena status bisa berubah (revoked) sejak kartu dicetak.
