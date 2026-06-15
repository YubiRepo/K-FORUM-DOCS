# API Spec — Accounting Backoffice (v1.0)

API specification untuk Accounting module KAI App — **backoffice surface** (Superadmin & Admin Region). Tidak ada surface mobile.

Base URL: `/api/v1/admin/accounting`
Auth: Bearer token
- **Admin Region** → scope otomatis ke region sendiri
- **Superadmin** → semua region + Pusat

---

## ENTRIES

### 1. POST /entries
Catat transaksi baru (IN/OUT).

**Auth:** Admin Region (region sendiri) / Superadmin

**Request Body:**
```json
{
  "direction": "OUT",
  "category_id": "cat_exp_event",
  "region_id": "reg_jakarta",
  "amount": 5000000,
  "currency": "IDR",
  "exchange_rate": 1,
  "description": "Sewa venue event Juni",
  "transaction_date": "2026-06-10",
  "attachment_url": "https://cdn.kai.app/receipts/abc.jpg",
  "source_ref": null
}
```

**Validation:**
- `direction`: required, `IN` | `OUT` (harus cocok dengan `direction` kategori)
- `category_id`: required, kategori harus aktif
- `amount`: required, > 0
- `currency`: default `IDR`; jika ≠ IDR maka `exchange_rate` wajib
- `transaction_date`: required
- Admin Region: `region_id` dipaksa ke region sendiri (abaikan input lain)
- Jika setting `require_attachment_for_out = true` dan `direction = OUT`: `attachment_url` wajib

**Response (201 Created):**
```json
{
  "data": {
    "id": "entry_001",
    "direction": "OUT",
    "category": { "id": "cat_exp_event", "code": "EXP_EVENT", "name": "Biaya Event" },
    "region_id": "reg_jakarta",
    "amount": 5000000,
    "currency": "IDR",
    "exchange_rate": 1,
    "amount_base": 5000000,
    "description": "Sewa venue event Juni",
    "transaction_date": "2026-06-10",
    "attachment_url": "https://cdn.kai.app/receipts/abc.jpg",
    "source": "manual",
    "status": "recorded",
    "created_by": "usr_admin_jkt",
    "created_at": "2026-06-11T08:00:00.000Z"
  }
}
```

**Response (422 — direction mismatch):**
```json
{ "message": "Direction does not match category", "errors": { "direction": ["Category EXP_EVENT is OUT"] } }
```

---

### 2. GET /entries
List & filter transaksi.

**Auth:** Admin Region (region sendiri) / Superadmin

**Query Parameters:**
- `region` — filter region (Admin Region ter-scope otomatis)
- `direction` — `IN` | `OUT`
- `category_id`
- `status` — `recorded` | `verified` | `void`
- `source` — `manual` | `system`
- `from`, `to` — rentang `transaction_date`
- `limit` (default 20, max 100), `offset`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "entry_001",
      "direction": "OUT",
      "category": { "code": "EXP_EVENT", "name": "Biaya Event" },
      "region_id": "reg_jakarta",
      "amount": 5000000,
      "currency": "IDR",
      "amount_base": 5000000,
      "description": "Sewa venue event Juni",
      "transaction_date": "2026-06-10",
      "status": "recorded",
      "source": "manual"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 1 }
}
```

---

### 3. GET /entries/:id
Detail satu transaksi.

**Auth:** Admin Region (region sendiri) / Superadmin

**Response (200 OK):** objek entry lengkap (semua field, termasuk audit & reconciliation).

**Response (403):**
```json
{ "message": "You can only access entries in your own region" }
```

---

### 4. PUT /entries/:id
Edit transaksi. Hanya bila status `recorded` (belum `verified`).

**Auth:** Pembuat entri (region sendiri) / Superadmin

**Request Body:** field yang sama dengan POST (subset boleh).

**Response (200 OK):** entry ter-update (`amount_base` dihitung ulang jika amount/rate berubah).

**Response (409 — sudah verified):**
```json
{ "message": "Cannot edit a verified entry. Void and create a new one instead." }
```

---

### 5. POST /entries/:id/void
Batalkan transaksi (tidak menghapus, tetap tersimpan untuk audit).

**Auth:** Pembuat (sebelum verified) / Superadmin (kapan saja)

**Request Body:**
```json
{ "reason": "Salah input nominal" }
```

**Response (200 OK):**
```json
{ "data": { "id": "entry_001", "status": "void", "void_reason": "Salah input nominal" } }
```

---

### 6. POST /entries/:id/verify
Verifikasi transaksi. Hanya aktif jika setting `verification_required = true`.

**Auth:** Superadmin only

**Response (200 OK):**
```json
{
  "data": {
    "id": "entry_001",
    "status": "verified",
    "verified_by": "usr_superadmin",
    "verified_at": "2026-06-11T10:00:00.000Z"
  }
}
```

**Response (409 — verifikasi nonaktif):**
```json
{ "message": "Verification is disabled in accounting settings" }
```

---

### 7. POST /entries/internal (INTERNAL — future auto-record)
Endpoint internal untuk modul lain (Subscription, Ads, payment webhook) push entri otomatis. **Bukan untuk UI.** Phase 2.

**Auth:** Service-to-service token (internal)

**Request Body:**
```json
{
  "direction": "IN",
  "category_code": "REV_SUBSCRIPTION",
  "region_id": null,
  "amount": 80000,
  "currency": "IDR",
  "transaction_date": "2026-06-11",
  "source": "system",
  "source_ref": "req_123",
  "external_txn_id": null,
  "payment_provider": "manual"
}
```
> Entri dibuat dengan `source = system`. Idempoten via `source_ref` (mencegah double-record bila modul retry).

**Response (201 Created):** entry yang dibuat.

---

## CATEGORIES

### 8. GET /categories
List kategori (hierarkis).

**Auth:** Admin Region / Superadmin

**Query Parameters:**
- `direction` (`IN`/`OUT`)
- `active` (bool)
- `tree` (bool, default `false`) — jika `true`, return bertingkat (parent berisi `children`); jika `false`, flat list

**Response (200 OK — tree=true):**
```json
{
  "data": [
    {
      "id": "cat_exp_op", "code": "EXP_OPERATIONAL", "name": "Operasional",
      "direction": "OUT", "parent_id": null, "is_active": true,
      "children": [
        { "id": "cat_op_listrik", "code": "EXP_OP_ELECTRICITY", "name": "Listrik", "direction": "OUT", "parent_id": "cat_exp_op", "is_active": true },
        { "id": "cat_op_internet", "code": "EXP_OP_INTERNET", "name": "Internet", "direction": "OUT", "parent_id": "cat_exp_op", "is_active": true }
      ]
    },
    {
      "id": "cat_rev_sub", "code": "REV_SUBSCRIPTION", "name": "Subscription Revenue",
      "direction": "IN", "parent_id": null, "is_active": true, "children": []
    }
  ]
}
```

**Response (200 OK — tree=false / flat):**
```json
{
  "data": [
    { "id": "cat_exp_op", "code": "EXP_OPERATIONAL", "name": "Operasional", "direction": "OUT", "parent_id": null, "is_active": true },
    { "id": "cat_op_listrik", "code": "EXP_OP_ELECTRICITY", "name": "Listrik", "direction": "OUT", "parent_id": "cat_exp_op", "is_active": true }
  ]
}
```

### 9. POST /categories
Buat kategori baru (parent atau child).

**Auth:** Superadmin only

**Request Body (parent / top-level):**
```json
{ "code": "REV_MERCHANDISE", "name": "Merchandise Sales", "direction": "IN", "description": "Penjualan merchandise KAI" }
```

**Request Body (child — sertakan `parent_id`):**
```json
{ "code": "REV_MERCH_APPAREL", "name": "Apparel", "direction": "IN", "parent_id": "cat_rev_merch" }
```

**Validation:**
- `code`: required, unik
- `direction`: required, `IN`/`OUT`. Jika child, harus sama dengan `direction` parent.
- `parent_id`: opsional. Jika diisi, parent harus top-level (tidak boleh child dari child → maks 2 level).

**Response (201 Created):** kategori baru.

**Response (422 — pelanggaran hierarki):**
```json
{ "message": "Invalid category hierarchy", "errors": { "parent_id": ["Parent must be a top-level category (max 2 levels)"], "direction": ["Child direction must match parent (OUT)"] } }
```

### 10. PUT /categories/:id
Edit kategori (nama, deskripsi, is_active). `code`, `direction`, dan `parent_id` tidak bisa diubah jika sudah dipakai transaksi.

**Auth:** Superadmin only

**Response (200 OK):** kategori ter-update.

**Response (409 — nonaktifkan parent yang punya child aktif):**
```json
{ "message": "Cannot deactivate parent with active children. Deactivate children first." }
```

> Kategori yang sudah dipakai entri tidak bisa dihapus — hanya `is_active = false`.

---

## REPORTS

### 11. GET /reports/summary
Ringkasan saldo & cashflow untuk periode.

**Auth:** Admin Region (region sendiri) / Superadmin (global atau filter region)

**Query Parameters:** `region`, `from`, `to`

**Response (200 OK):**
```json
{
  "data": {
    "period": { "from": "2026-06-01", "to": "2026-06-30" },
    "region_id": "reg_jakarta",
    "currency": "IDR",
    "total_in": 12000000,
    "total_out": 7500000,
    "balance": 4500000,
    "entry_count": 24
  }
}
```

### 12. GET /reports/by-category
Breakdown per kategori, mendukung roll-up ke parent atau rincian per child.

**Auth:** Admin Region (region sendiri) / Superadmin

**Query Parameters:** `region`, `from`, `to`, `direction`, `group_by` (`parent` | `child`, default `child`)

**Response (200 OK — group_by=parent, roll-up):**
```json
{
  "data": [
    { "code": "REV_SUBSCRIPTION", "name": "Subscription Revenue", "direction": "IN", "total": 8000000 },
    { "code": "REV_EVENT", "name": "Event Revenue", "direction": "IN", "total": 6000000 },
    { "code": "EXP_OPERATIONAL", "name": "Operasional", "direction": "OUT", "total": 2500000 }
  ]
}
```

**Response (200 OK — group_by=child, drill-down):**
```json
{
  "data": [
    { "code": "REV_EVENT_TICKET", "name": "Tiket", "parent_code": "REV_EVENT", "direction": "IN", "total": 4000000 },
    { "code": "REV_EVENT_SPONSOR", "name": "Sponsor", "parent_code": "REV_EVENT", "direction": "IN", "total": 2000000 },
    { "code": "EXP_OP_INTERNET", "name": "Internet", "parent_code": "EXP_OPERATIONAL", "direction": "OUT", "total": 1500000 },
    { "code": "EXP_OP_ELECTRICITY", "name": "Listrik", "parent_code": "EXP_OPERATIONAL", "direction": "OUT", "total": 1000000 }
  ]
}
```
> `group_by=parent` menjumlahkan semua child ke induknya (entri yang di-assign langsung ke parent juga ikut). `group_by=child` merinci per kategori asli.

### 13. GET /reports/by-region (Superadmin)
Cashflow per region untuk laporan global.

**Auth:** Superadmin only

**Query Parameters:** `from`, `to`

**Response (200 OK):**
```json
{
  "data": [
    { "region_id": "reg_jakarta", "total_in": 12000000, "total_out": 7500000, "balance": 4500000 },
    { "region_id": "reg_surabaya", "total_in": 6000000, "total_out": 4000000, "balance": 2000000 },
    { "region_id": null, "label": "Pusat", "total_in": 20000000, "total_out": 15000000, "balance": 5000000 }
  ]
}
```

### 14. GET /reports/export
Export transaksi ke CSV/JSON (untuk software akuntansi eksternal).

**Auth:** Admin Region (region sendiri) / Superadmin

**Query Parameters:** `region`, `from`, `to`, `format` (`csv` | `json`)

**Response:** File (`text/csv` atau `application/json`). Kolom termasuk `code` kategori standar untuk mapping ke software akuntansi.

---

## SETTINGS

### 15. GET /settings
Ambil setting accounting.

**Auth:** Superadmin only

**Response (200 OK):**
```json
{
  "data": {
    "verification_required": false,
    "default_currency": "IDR",
    "allow_region_admin_input": true,
    "require_attachment_for_out": false,
    "fiscal_year_start_month": 1
  }
}
```

### 16. PUT /settings
Update setting.

**Auth:** Superadmin only

**Request Body:** subset field setting.

**Response (200 OK):** setting ter-update.

---

## CATATAN ARSITEKTUR
1. **Scope region wajib di-enforce di backend** — Admin Region selalu di-filter ke region sendiri, jangan andalkan frontend.
2. **Verifikasi opsional** — endpoint `/verify` hanya berfungsi jika `verification_required = true`.
3. **Entri void tidak dihapus** — tetap tersimpan, dikeluarkan dari laporan.
4. **`amount_base` (IDR)** dipakai di semua laporan agar konsisten lintas currency.
5. **Endpoint `/entries/internal`** disiapkan untuk auto-record Phase 2; idempoten via `source_ref`.
