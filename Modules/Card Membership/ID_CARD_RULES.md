# KAI ID Card Module — Rules & Use Cases (v2.0)

Dokumentasi sistem rules ID Card KAI App. Versi ini **memisahkan ID Card dari Subscription** — kartu adalah identitas keanggotaan yang stabil, bukan representasi status pembayaran. Semua logika benefit/plan di-resolve secara live saat verifikasi, bukan disimpan di kartu.

> **Perubahan utama dari v1.0:** field `plan` dihapus dari kartu & QR, `expiry_date` dihapus, status disederhanakan jadi `active`/`revoked`, verifikasi return konteks identitas (bukan benefit), dan ditambahkan fondasi future-proof (`card_scan_events`, QR versioning).

---

## 1. WHAT IS ID CARD?

**ID Card** adalah bukti formal keanggotaan setiap member KAI. Fungsinya tiga lapis:

1. **Identitas** — bukti "orang ini member KAI" (digital + fisik).
2. **Verification token** — QR yang di-scan untuk membuktikan keanggotaan secara real-time.
3. **Gateway** — pintu masuk ke ekosistem (benefit partner, event, loyalty) — semua di-resolve live oleh modul lain, bukan oleh kartu.

### Prinsip Desain (PALING PENTING)

| Prinsip | Penjelasan |
|---------|-----------|
| **Kartu = identitas stabil** | Tidak menyimpan plan, benefit, atau status pembayaran. Identitas tidak hilang saat plan berubah. |
| **Resolve live, jangan snapshot** | Plan & benefit di-resolve dari modul Subscription saat scan, bukan disalin ke kartu. |
| **Satu sumber kebenaran** | Plan tetap di Subscription. Benefit tetap di sistem benefit. Kartu tidak menduplikasi. |
| **Kartu "bodoh", consumer "pintar"** | Semua fungsi (benefit, attendance, loyalty) lahir di sisi consumer/backend saat scan — kartu tidak pernah perlu berubah. |

### Berbeda dari:
- **Subscription:** menjawab "apa yang boleh dilakukan" (entitlement, berubah-ubah). ID Card menjawab "siapa orang ini" (identitas, stabil).
- **Profile:** data lengkap user. ID Card adalah representasi ringkas + verifiable.
- **Role/Permission:** kontrol aksi di dalam app. ID Card untuk verifikasi eksternal (offline/partner).

---

## 2. KENAPA TERPISAH DARI SUBSCRIPTION?

Keputusan arsitektur inti. Alasan:

1. **Pertanyaan berbeda.** Identitas ("member KAI?") ≠ entitlement ("boleh apa?"). Downgrade Pro→Standard tidak menghilangkan keanggotaan — hanya benefit-nya.
2. **Identitas stabil, plan volatile.** Subscription penuh state (upgrade/downgrade/renewal/expired/cancel). Kalau kartu harus sinkron tiap perubahan = sumber bug & sinkronisasi tanpa manfaat.
3. **Salah konsep kalau dicampur.** Kalau kartu mati saat subscription lapse, member sah yang telat bayar kehilangan identitasnya. Itu keliru.

**Hubungan yang benar:** kartu menyentuh plan/benefit **hanya saat verifikasi (read-time)**, dengan resolve live — bukan dengan menyimpan plan di kartu.

```
SALAH (coupled):                      BENAR (decoupled):
card.plan = "pro" (snapshot)          card = identitas murni
verify → baca card.plan               verify → resolve user.plan + benefit LIVE
       → kasih benefit                       → kasih benefit sesuai kondisi saat itu
```

---

## 3. DATA MODEL

### Tabel: `membership_id_cards`

| Field | Type | Keterangan |
|-------|------|-----------|
| `card_id` | String (PK) | Unique ID, format: `KAI-{YYYY}-{Random}`. Contoh: `KAI-2026-USR001` |
| `user_id` | String (FK, UNIQUE) | Reference ke `users.id`. 1 user = 1 kartu |
| `region_id` | String (FK) | Region saat kartu issued (untuk filtering admin & akses region-based) |
| `issued_date` | DateTime (UTC) | Waktu kartu dibuat (auto-set) |
| `status` | Enum | `active` (valid) / `revoked` (dibatalkan permanen) |
| `qr_version` | String | Versi format QR, mis. `v1`. Untuk forward-compat |
| `qr_code_data` | String | `{version}\|{card_id}\|{checksum}` — **TANPA plan/user_id mentah** |
| `digital_format` | JSON | Data tampilan: `{design, user_name, avatar_url, qr_url}` — **TANPA plan_badge** |
| `physical_ordered` | Boolean | Apakah sudah order fisik? |
| `order_id` | String (FK, nullable) | Reference ke fulfillment order |
| `created_at` | DateTime | Saat record dibuat |
| `updated_at` | DateTime | Saat terakhir di-update |

> **Catatan perubahan:** `plan` dan `expiry_date` **dihapus** dari schema. Plan di-resolve live; identitas tidak punya konsep "kadaluarsa pembayaran".

### Tabel: `card_scan_events` (BARU — fondasi future-proof)

Event stream setiap kali kartu di-scan. Fondasi untuk attendance, redemption, loyalty, analytics, dan fraud detection.

| Field | Type | Keterangan |
|-------|------|-----------|
| `id` | UUID (PK) | |
| `card_id` | String (FK) | Kartu yang di-scan |
| `user_id` | String (FK) | User pemilik kartu (denormalized untuk query cepat) |
| `scanned_by` | String (nullable) | Siapa yang scan: `partner_id`, `event_gate_id`, `admin_id`, atau null (public) |
| `context` | Enum | `directory` / `event` / `admin` / `public_share` / `loyalty` |
| `context_ref` | String (nullable) | Reference kontekstual (mis. `event_id`, `merchant_id`) |
| `result` | Enum | `valid` / `revoked` / `invalid_checksum` / `not_found` |
| `metadata` | JSON (nullable) | Data tambahan per konteks (benefit_applied, value, dll) |
| `created_at` | DateTime | Timestamp scan |

---

## 4. LIFECYCLE

### 4.1 Card Creation
- **Trigger:** User menyelesaikan registration
- **Action:**
  - Generate unique `card_id`
  - Set `region_id` dari region user saat ini
  - Set `issued_date = now()`, `status = active`, `qr_version = v1`
  - Generate `qr_code_data` (lihat §7)
  - Build `digital_format` JSON
- **Catatan:** TIDAK menyalin plan. Kartu lahir sebagai identitas murni.

### 4.2 Digital Card Access
User dapat:
- **View** di Profile > Membership ID Card
- **Download** sebagai PDF (self-contained)
- **Share** QR link untuk verifikasi online

### 4.3 Physical Card Request
- User request kartu fisik dari app
- Trigger fulfillment order (print + ship)
- Set `physical_ordered = true`, simpan `order_id`

### 4.4 Verification
- Partner/event/admin scan QR → backend verifikasi (lihat §6 & §7)
- Setiap scan tercatat di `card_scan_events`

### 4.5 Plan Change → TIDAK MENYENTUH KARTU
- Upgrade/downgrade/renewal/expired → **kartu tidak berubah sama sekali**
- Next scan otomatis resolve plan/benefit terbaru
- History plan ditangani modul Subscription (`subscription_history`), bukan kartu

### 4.6 Revocation (HANYA alasan identitas)
Kartu jadi `revoked` **hanya** untuk alasan keanggotaan, **bukan pembayaran**:
- ✅ User di-ban
- ✅ Akun dihapus
- ✅ Keluar dari KAI
- ❌ BUKAN karena: subscription lapse, downgrade, telat bayar

> Revocation bersifat permanen. Tidak ada status `expired` — identitas tidak kadaluarsa.

---

## 5. ACTOR & PERMISSION MATRIX

| Action | Member (pemilik) | Partner/Staff | Admin Regional | Superadmin |
|--------|:---:|:---:|:---:|:---:|
| View own card | ✅ | ❌ | — | — |
| Download PDF | ✅ | ❌ | — | — |
| Share QR link | ✅ | ❌ | — | — |
| Request physical card | ✅ | ❌ | — | — |
| Scan & verify card | ❌ | ✅ | ✅ | ✅ |
| List/filter cards | ❌ | ❌ | ✅ (region sendiri) | ✅ (semua) |
| Revoke card | ❌ | ❌ | ✅ (region sendiri) | ✅ |
| View scan analytics | ❌ | ✅ (scan sendiri) | ✅ (region) | ✅ (global) |

---

## 6. USE CASES & VERIFICATION FLOW

### Use Case 1: Directory / Toko Partner
```
Member → Tunjukkan kartu ke staff toko
  ↓
Staff → Scan QR
  ↓
Backend → Validasi checksum + status=active
        → Resolve user.plan + benefit LIVE dari Subscription
        → Catat card_scan_events (context=directory)
  ↓
App → "Member John Doe — Valid — Benefit: directory_discount"
  ↓
Staff → Apply benefit sesuai data live
```

### Use Case 2: Event Entry
```
Member → Tiba di gate event
  ↓
Gate → Scan QR
  ↓
Backend → Validasi → cek registrasi event (jika RSVP aktif)
        → Catat card_scan_events (context=event, context_ref=event_id)
  ↓
System → Grant/deny entry + catat attendance
```

### Use Case 3: Digital Share (Online Verification)
```
Member → Share link: kai.app/verify/KAI-2026-001
  ↓
Recipient → Buka link → lihat info publik (nama, region, status valid)
  ↓
Verifikasi: "Orang ini member KAI yang valid"
```
> Link publik **hanya** tampilkan: nama, region, status. TANPA plan/benefit/data sensitif.

### Use Case 4: Admin Panel
```
Admin → Member Management → cari "John Doe"
  ↓
View → status kartu, issued_date, region, physical_ordered, riwayat scan
  ↓
Actions → Revoke (jika banned), lihat scan history
```

---

## 7. SECURITY & VALIDATION

### 7.1 QR Code Format (versioned, identity-only)
```
Format: {version}|{card_id}|{checksum}
Checksum: SHA256(version + card_id + secret_salt)

Contoh: v1|KAI-2026-001|a1b2c3d4e5f6g7h8
```

**Kenapa tidak ada plan & user_id mentah di QR?**
- **Plan** di-resolve live saat scan — supaya perubahan benefit langsung kepakai tanpa cetak ulang kartu.
- **user_id mentah** tidak diekspos — kurangi permukaan serangan enumerasi. Backend resolve user dari `card_id`.
- **Versioning** (`v1`) membuat format bisa berkembang tanpa mematikan kartu fisik lama.

### 7.2 Backend Verification (urutan wajib)
1. Parse versi → handle sesuai versi
2. Validasi checksum
3. Cek `status = active`
4. Resolve user → plan & benefit live (jika consumer butuh)
5. Catat ke `card_scan_events` (selalu, termasuk yang gagal)

### 7.3 Digital Share Security
- Link publik: nama, region, status saja
- Rate-limit verifikasi (cegah enumerasi)
- Checksum wajib valid untuk semua verifikasi

---

## 8. FUNGSI DI LUAR IDENTITAS (Future-Proof)

Semua fungsi ini lahir dari **QR + scan event stream**, bukan dari data di kartu. Artinya bisa ditambahkan tanpa mengubah kartu (terutama fisik yang sudah dicetak).

| Fungsi | Penjelasan | Yang perlu disiapkan sekarang |
|--------|-----------|-------------------------------|
| **Verification token** | Scan → resolve benefit live → apply diskon partner | Verify return konteks (§6) |
| **Event attendance** | Scan di gate → catat presensi, sertifikat, analytics | `context=event` di scan events |
| **Partner redemption & settlement** | Log tiap redeem → laporan partner, revenue-share, promo terbatas | `card_scan_events` + `metadata` |
| **Region-based access** | Scan → tahu region → event/harga/lounge khusus region | `region_id` sudah di schema |
| **Loyalty & points** | Scan = earning poin; tier loyalty terpisah dari plan | Scan events sebagai sumber poin |
| **Digital wallet pass** | Add to Apple/Google Wallet, push update otomatis | `digital_format` lengkap + QR stabil |
| **Cross-partner identity** | Satu kartu = identity lintas partner ekosistem | QR versioning |

### Tiga keputusan kecil yang membuka semua di atas
1. **`card_scan_events` sebagai event stream** sejak Phase 1.
2. **Verify return konteks kaya** (`{valid, user_id, region_id, status}`), bukan boolean. Benefit di-resolve live oleh consumer.
3. **QR versioning + identifier stabil** (`v1|card_id|checksum`), tanpa plan/benefit di dalamnya.

---

## 9. INTEGRATION POINTS

| Modul | Hubungan |
|-------|----------|
| **Subscription** | Sumber kebenaran plan/benefit. Di-query live saat verify. Kartu TIDAK menyimpan plan. |
| **Profile** | Host tampilan kartu. Sumber `user_name`, `avatar_url`. |
| **Region** | Sumber `region_id` saat issued. Dasar akses region-based. |
| **Directory** | Consumer verifikasi → apply benefit partner. |
| **Event** | Consumer verifikasi → entry + attendance. |
| **Notification** | Notify saat kartu fisik dikirim, saat kartu di-revoke. |
| **Admin Panel** | Kelola, revoke, lihat scan analytics. |

---

## 10. OPEN DECISIONS (resolved di v2.0)

| # | Open Decision (v1) | Resolusi v2.0 |
|---|-------------------|---------------|
| 1 | Plan change → update vs new card? | **Tidak relevan** — kartu tidak menyimpan plan. Tidak berubah saat plan berubah. |
| 2 | Expiry policy? | **Tidak ada expiry** — identitas tidak kadaluarsa. (Jika kartu fisik perlu re-print berkala, itu siklus terpisah.) |
| 3 | Physical card cost? | Bisa jadi benefit Pro (`physical_id_card`) atau shipping-only. Diputuskan tim bisnis — TIDAK mempengaruhi validitas kartu. |
| 4 | QR offline verification? | **Hybrid** — quick check offline (checksum), tapi benefit bernilai tinggi WAJIB backend call (karena status bisa berubah). |
| 5 | Card replacement? | Diizinkan, batasi 2x/tahun untuk cegah abuse fulfillment. Replacement = `card_id` baru, `user_id` sama. |

---

## 11. IMPLEMENTATION ROADMAP

### Phase 1: Digital Card + Fondasi Future-Proof
- Auto-create kartu saat registration (identitas murni)
- View, download PDF, share QR
- Endpoint verify (return konteks, bukan boolean)
- **`card_scan_events` table** (fondasi)
- **QR versioning** (`v1`)

### Phase 2: Physical Card
- Print-on-demand + shipping integration
- Track fulfillment
- Replacement flow (max 2x/tahun)

### Phase 3: Ecosystem Functions
- Event attendance via scan
- Partner redemption & settlement reports
- Loyalty/points dari scan events
- Digital wallet pass (Apple/Google)
- Region-based access
- Card customization (design themes)

---

## 12. SUMMARY

ID Card = **identitas stabil + QR sebagai gateway**. Kartu tidak menyimpan apa-apa selain identitas; semua fungsi (benefit, attendance, loyalty, settlement) lahir dari **scan event + resolve live**. Inilah yang membuatnya future-proof: fungsi baru ditambahkan di sisi consumer/backend, sedangkan kartu — terutama yang fisik dan sudah dicetak — tidak pernah perlu berubah.
