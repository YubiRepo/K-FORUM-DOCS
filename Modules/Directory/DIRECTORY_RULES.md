# Directory Module — System Rules (v2.0)

Dokumentasi lengkap system rules Directory module KAI App. Mencakup semua aturan bisnis, validasi, lifecycle, dan constraint.

---

## 1. OVERVIEW & TUJUAN

Directory adalah fitur yang memungkinkan **Member Pro** mendaftarkan bisnis/toko mereka ke dalam katalog bisnis komunitas Korea-Indonesia. Member biasa dapat menelusuri, menyimpan, mereview, dan menghubungi merchant.

**Scope**: Global — semua user dapat melihat semua merchant dari seluruh region.

---

## 2. HIERARCHY STRUCTURE

```
Member Pro (User)
  └── Company (Bisnis utama)
       ├── Company Profile (name, logo, description, contact)
       │
       ├── Merchant 1 (Toko/outlet/cabang)
       │    ├── Categories (multiple, dari category_master)
       │    ├── Images (min 1)
       │    ├── Location (optional untuk online, wajib untuk fisik)
       │    ├── Operating Hours
       │    ├── Contact Info
       │    ├── Items (Product + Service merged, unlimited)
       │    │    ├── Item Type: "product" (harga tunggal, stock optional)
       │    │    └── Item Type: "service" (price_range, duration)
       │    ├── Reviews (user → merchant, 1 review per user)
       │    └── Inquiries (user → merchant, tanya jawab)
       │
       ├── Merchant 2
       └── Merchant N
```

---

## 3. ACTOR & PERMISSION MATRIX

### 3.1 Member Pro (Merchant Owner)

| Action | Rule |
|--------|------|
| Create Company | ✅ Unlimited |
| Edit/Delete Company | ✅ Only own company |
| Create Merchant | ✅ Unlimited per company (subject to settings) |
| Edit Merchant | ✅ Only own merchant |
| Archive Merchant | ✅ Hides dari listing, data tetap ada |
| Delete Merchant | ✅ Hanya jika belum punya reviews/inquiries, atau force by admin |
| Publish Merchant | ✅ Trigger approval flow (lihat §5) |
| Add/Edit/Delete Items | ✅ Only own merchant |
| Reply Inquiry | ✅ Only own merchant |
| View own stats/analytics | ✅ |
| Review own merchant | ❌ Tidak diperbolehkan |

### 3.2 Regular Member

| Action | Rule |
|--------|------|
| Browse directory | ✅ Lihat semua published merchant |
| Search & filter | ✅ Category, kota, rating, nama |
| View merchant detail | ✅ |
| Save/unsave merchant | ✅ |
| Leave review | ✅ 1 review per merchant per user |
| Edit own review | ✅ |
| Send inquiry | ✅ |
| Create company/merchant | ❌ Requires Member Pro plan |

### 3.3 Admin Regional

| Action | Rule |
|--------|------|
| View companies | ✅ Hanya region sendiri |
| View & filter merchants | ✅ Hanya region sendiri |
| Approve merchant | ✅ Hanya region sendiri |
| Reject merchant | ✅ Hanya region sendiri (wajib sertakan reason) |
| Ban merchant | ✅ Hanya region sendiri |
| Delete merchant (force) | ❌ Hanya Superadmin |
| Moderate reviews | ✅ Region sendiri |
| View analytics | ✅ Hanya region sendiri |
| Manage category master | ❌ |
| Manage directory settings | ❌ |

### 3.4 Superadmin

| Action | Rule |
|--------|------|
| Semua aksi Admin Regional | ✅ Semua region |
| Delete company/merchant (force) | ✅ |
| Manage category master | ✅ |
| Manage directory settings | ✅ |
| View global analytics | ✅ |
| Unban merchant | ✅ |

---

## 4. COMPANY RULES

### 4.1 Validasi Create/Update Company

| Field | Required | Constraint |
|-------|----------|-----------|
| name | ✅ | Min 3, max 200 char |
| description | ❌ | Max 2000 char |
| logo_url | ❌ | Valid URL, upload via media API terlebih dahulu |
| phone | ❌ | Format Indonesia (+62 atau 0xxx) |
| email | ❌ | Valid email format |
| website | ❌ | Valid URL |

### 4.2 Company Lifecycle

```
create → active
active → inactive  (owner set manual)
active → banned    (by admin)
inactive → active  (owner reaktivasi)
banned → active    (Superadmin only, unban)
```

### 4.3 Company Rules

- Satu Member Pro dapat punya **unlimited companies**
- Company yang `banned` → semua merchantnya otomatis tidak tampil di listing
- Company tidak bisa dihapus jika masih punya merchant aktif
- Company `inactive` → merchant tetap bisa publish (company tidak affect merchant visibility secara langsung, hanya status company sendiri)

---

## 5. MERCHANT RULES

### 5.1 Merchant Types

| Type | Deskripsi | Location Wajib? |
|------|-----------|-----------------|
| `retail` | Toko penjualan produk fisik | ✅ Ya |
| `food_beverage` | Restaurant, cafe, catering | ✅ Ya |
| `beauty` | Salon, spa, klinik kecantikan | ✅ Ya |
| `service` | Jasa konsultasi, reparasi, rental | ✅ Ya |
| `online` | Online shop, e-commerce murni | ❌ Opsional |
| `other` | Lainnya | Depends on `require_location_for_physical` setting |

### 5.2 Validasi Create/Update Merchant

| Field | Required | Constraint |
|-------|----------|-----------|
| company_id | ✅ | Must own company |
| name | ✅ | Min 3, max 200 char |
| description | ✅ | Min 20, max 3000 char |
| type | ✅ | Enum: retail, online, service, food_beverage, beauty, other |
| categories | ✅ | Min 1, max 5, dari category_master |
| images | ✅ (jika `min_images_before_publish` > 0) | Min 1, max 10, upload terlebih dahulu |
| address | ✅ (jika type ≠ online) | Max 500 char |
| latitude | ✅ (jika type ≠ online) | Valid lat: -90 to 90 |
| longitude | ✅ (jika type ≠ online) | Valid lng: -180 to 180 |
| region_id | ❌ | AUTO-DERIVED dari lat/lng via geo-lookup |
| city | ❌ | Auto-fill dari geocoding |
| province | ❌ | Auto-fill dari geocoding |
| phone | ❌ (setidaknya 1 contact) | Format valid |
| email | ❌ | Format valid |
| whatsapp | ❌ | Format valid |
| instagram | ❌ | Max 50 char |
| hours | ❌ | JSON per hari (open, close, closed) |

### 5.3 Merchant Approval Flow

```
[Member Pro] → Create/Submit Merchant
       │
       ▼
[System] Validasi field wajib
       │
       ├── FAIL → Return validation error (status: draft)
       │
       └── PASS → Cek setting require_merchant_approval
                     │
                     ├── TRUE → status: pending_approval
                     │          → Notifikasi ke Admin Regional (jika ada) / Superadmin
                     │          │
                     │          ├── [Admin] Approve → status: published, published_at = NOW()
                     │          │                  → Notifikasi ke owner: "Merchant disetujui"
                     │          │
                     │          └── [Admin] Reject  → status: rejected, rejection_reason wajib
                     │                             → Notifikasi ke owner: "Merchant ditolak: {reason}"
                     │                             → Owner dapat edit & resubmit
                     │
                     └── FALSE → status: published (auto-approve)
                                → published_at = NOW()
```

### 5.4 Merchant Status & Visibility

| Status | Visible di Listing | Owner Lihat | Admin Lihat |
|--------|--------------------|-------------|-------------|
| `draft` | ❌ | ✅ | ✅ |
| `pending_approval` | ❌ | ✅ (dengan status info) | ✅ |
| `published` | ✅ | ✅ | ✅ |
| `rejected` | ❌ | ✅ (dengan reason) | ✅ |
| `archived` | ❌ | ✅ | ✅ |
| `banned` | ❌ | ✅ (dengan reason) | ✅ |

### 5.5 Merchant Archive vs Delete vs Ban

```
Archive (by owner):
  - Merchant disembunyikan dari public listing
  - Data tetap lengkap (items, reviews, inquiries)
  - Owner masih bisa lihat & unarchive kapan saja
  - Bisa jadi: published kembali

Delete (by owner):
  - Hanya bisa jika: reviews_count = 0 AND inquiries_count = 0
  - Jika sudah ada data, gunakan Archive
  - Hard delete dari database

Force Delete (Superadmin only):
  - Bisa delete meski ada reviews/inquiries
  - Cascade delete semua related data

Ban (by Admin):
  - Merchant tidak tampil, tidak bisa diubah owner
  - Merchant bisa di-unban oleh Superadmin
  - Alasan ban wajib dicatat
```

### 5.6 Merchant Stats (Computed/Cached)

Stats berikut di-update secara realtime atau via background job:

| Field | Source | Update Trigger |
|-------|--------|----------------|
| `item_count` | COUNT items WHERE status != archived | Item add/delete/archive |
| `review_count` | COUNT reviews WHERE status = published | Review create/moderate |
| `rating` | AVG(rating) dari published reviews | Review create/update |
| `favorite_count` | COUNT favorites | Favorite add/remove |
| `inquiry_count` | COUNT inquiries | Inquiry create |
| `total_views` | Increment per unique page view | Merchant detail viewed |

---

## 6. ITEM RULES (Product + Service Merged)

### 6.1 Validasi Item

#### Untuk Product (`type: "product"`):

| Field | Required | Constraint |
|-------|----------|-----------|
| name | ✅ | Min 3, max 200 char |
| type | ✅ | `"product"` |
| price | ✅ | > 0, max 999.999.999 |
| currency | ✅ | Default: `IDR` |
| unit | ❌ | Max 50 char (contoh: "per pack", "per kg") |
| description | ❌ | Max 2000 char |
| category | ❌ | Free text, max 100 char |
| images | ✅ (min 1) | Max 10 images |
| stock | ❌ | null = unlimited; integer ≥ 0 = limited |
| status | ✅ | `available` atau `unavailable` |

#### Untuk Service (`type: "service"`):

| Field | Required | Constraint |
|-------|----------|-----------|
| name | ✅ | Min 3, max 200 char |
| type | ✅ | `"service"` |
| price_min | ✅ | > 0 |
| price_max | ✅ | ≥ price_min |
| currency | ✅ | Default: `IDR` |
| duration_minutes | ❌ | Positive integer |
| description | ❌ | Max 2000 char |
| category | ❌ | Free text, max 100 char |
| images | ✅ (min 1) | Max 10 images |
| status | ✅ | `available` atau `unavailable` |

### 6.2 Item Business Rules

```
✅ Item category adalah FREE TEXT (bukan dari master list)
✅ Satu merchant bisa mix product + service
✅ Tidak ada batasan jumlah item per merchant (kecuali `item_max_per_merchant` setting)
✅ Item tidak bisa harga Rp 0 (kecuali `allow_free_price` setting = true)
✅ Image item min 1, max 10
✅ Merchant yang `banned` atau `archived` → item tidak tampil
✅ Item status `unavailable` → tetap tampil tapi dengan label "Tidak Tersedia"
✅ Item tidak perlu approval terpisah (kecuali `item_require_approval` setting = true)
```

---

## 7. CATEGORY MASTER RULES

### 7.1 Struktur

Category master dikelola oleh Superadmin. Merchant dapat pilih 1-5 kategori dari daftar ini.

```
Kategori default:
  - Retail 🏬
  - Food & Beverage 🍽️
  - Service 🔧
  - Beauty & Salon 💄
  - Health & Wellness 🏥
  - Education 📚
  - Entertainment 🎭
  - Travel & Tourism ✈️
  - Finance & Insurance 💰
  - Technology 💻
  - Fashion 👗
  - Other 📦
```

### 7.2 Category Rules

```
✅ Superadmin dapat add/edit/deactivate kategori
✅ Kategori yang di-deactivate: tidak muncul di pilihan baru, existing merchant tidak terpengaruh
✅ Kategori tidak bisa dihapus jika masih ada merchant menggunakannya
✅ Merchant wajib pilih minimum 1 kategori
✅ Merchant bisa pilih maximum 5 kategori
✅ Urutan tampil di UI ditentukan oleh field `order`
```

---

## 8. REVIEW RULES

### 8.1 Validasi Review

| Field | Required | Constraint |
|-------|----------|-----------|
| rating | ✅ | Integer 1-5 |
| title | ❌ | Max 200 char |
| review_text | ❌ | Max 2000 char |
| aspects.product_quality | ❌ | Integer 1-5 |
| aspects.service | ❌ | Integer 1-5 |
| aspects.price | ❌ | Integer 1-5 |

### 8.2 Review Business Rules

```
✅ 1 user = 1 review per merchant (unique constraint)
✅ User dapat EDIT review yang sudah ada (update, bukan create baru)
✅ User TIDAK bisa delete review
✅ Merchant owner TIDAK bisa review merchant sendiri
✅ Rating 1-5 WAJIB; title & text OPSIONAL
✅ Merchant rating (avg) di-update setelah setiap review create/update
✅ Jika `require_moderation = true` → review status: `pending` dulu
   - Admin approve → status: `published`
   - Admin reject → status: `rejected`
   - Default (jika moderation OFF) → status: `published` langsung
✅ Helpful/unhelpful voting: 1 user 1 vote per review
✅ Review `pending` tidak tampil di public listing
```

### 8.3 Review Moderation States

```
pending → published  (admin approve)
pending → rejected   (admin reject)
published → hidden   (admin hide, tidak hapus)
rejected → pending   (admin re-queue)
```

---

## 9. INQUIRY RULES

### 9.1 Validasi Inquiry

| Field | Required | Constraint |
|-------|----------|-----------|
| subject | ✅ | Min 5, max 200 char |
| message | ✅ | Min 10, max 2000 char |

### 9.2 Inquiry Business Rules

```
✅ Inquiry dibuat oleh member (login wajib)
✅ Merchant owner mendapat notifikasi saat ada inquiry baru
✅ Merchant owner dapat reply 1x (setelah reply, status → replied)
✅ Inquiry dapat di-close oleh owner atau auto-close setelah `auto_close_days` hari
✅ User dapat lihat semua inquiry mereka (ke semua merchant)
✅ Merchant owner dapat lihat semua inquiry ke merchant mereka
✅ Jika `allow_inquiries = false` pada merchant settings → inquiry form tidak muncul
✅ Jika `allow_auto_reply = true` → sistem kirim auto_reply_message sebelum owner reply
```

### 9.3 Inquiry States

```
pending → replied  (merchant owner reply)
pending → closed   (auto-close atau manual)
replied → closed   (auto-close atau manual)
```

---

## 10. LOCATION & REGION RULES

### 10.1 Region Derivation

```
Merchant fisik dengan lat/lng → sistem call geo-lookup API
  └── Return: region_id, city, province
  └── Simpan ke merchant record
  └── User tidak perlu input region_id manual
```

### 10.2 Location Rules per Merchant Type

| Type | address | lat/lng | region_id |
|------|---------|---------|-----------|
| online | Optional | Optional | Optional |
| retail | **WAJIB** | **WAJIB** | Auto-derived |
| service | **WAJIB** | **WAJIB** | Auto-derived |
| food_beverage | **WAJIB** | **WAJIB** | Auto-derived |
| beauty | **WAJIB** | **WAJIB** | Auto-derived |
| other | Depends on setting | Depends | Auto-derived if lat/lng provided |

### 10.3 Admin Regional Scope Filter

Admin Regional hanya melihat merchant dari region mereka:
```sql
WHERE merchant.region_id = admin.region_id
-- OR
WHERE merchant.region_id IS NULL  -- online merchants tidak di-filter by region untuk admin
```

**Note:** Online merchants (no region) tetap visible ke semua admin untuk moderation purposes.

---

## 11. DIRECTORY SETTINGS RULES

Settings dikelola oleh Superadmin dan berlaku secara global.

| Setting Key | Default | Deskripsi |
|-------------|---------|-----------|
| `merchant.require_approval` | `true` | Merchant baru butuh approval admin |
| `merchant.max_per_member` | `null` | Max merchant per Member Pro (null = unlimited) |
| `merchant.min_items_before_publish` | `0` | Min item sebelum bisa publish |
| `merchant.min_images_before_publish` | `1` | Min foto merchant |
| `item.require_approval` | `false` | Item baru butuh approval |
| `item.max_per_merchant` | `null` | Max item per merchant |
| `item.allow_free_price` | `false` | Boleh input harga Rp 0 |
| `review.allow_reviews` | `true` | Fitur review aktif |
| `review.require_moderation` | `true` | Review perlu dimoderasi sebelum tampil |
| `review.waiting_period_days` | `0` | Hari tunggu sebelum boleh review |
| `inquiry.allow_inquiries` | `true` | Fitur inquiry aktif |
| `inquiry.auto_close_days` | `30` | Auto-close inquiry setelah N hari |
| `features.allow_online_only` | `true` | Boleh daftar merchant tanpa lokasi |
| `features.allow_multiple_categories` | `true` | Merchant bisa pilih > 1 kategori |
| `location.require_location_for_physical` | `true` | Merchant fisik wajib isi koordinat |
| `location.auto_derive_region` | `true` | Region di-derive dari lat/lng otomatis |
| `location.enable_map_view` | `true` | Tampilkan peta di detail merchant |

---

## 12. NOTIFICATION RULES

| Event | Notif To | Channel |
|-------|----------|---------|
| Merchant submitted (pending) | Admin (regional jika ada, else Superadmin) | In-app + FCM |
| Merchant approved | Merchant owner | In-app + FCM |
| Merchant rejected | Merchant owner | In-app + FCM |
| Merchant banned | Merchant owner | In-app + FCM |
| New inquiry received | Merchant owner | In-app + FCM |
| Inquiry replied | Inquiry sender | In-app + FCM |
| New review (if moderation ON) | Admin | In-app |
| Review approved/rejected | Reviewer | In-app |

---

## 13. MEDIA / IMAGE UPLOAD RULES

```
✅ Semua image di-upload terlebih dahulu via /api/v1/mobile/media/upload
✅ Context: "directory"
✅ Max file size: 5 MB per file
✅ Max 5 file per request upload
✅ Format accepted: JPG, JPEG, PNG, WEBP
✅ Return: array of CDN URLs
✅ URLs yang valid digunakan saat create/update merchant atau item
```

---

## 14. SEARCH & FILTER RULES

### Public Listing Filters:

| Filter | Type | Deskripsi |
|--------|------|-----------|
| `q` | Text search | Nama merchant, deskripsi |
| `category_id` | UUID | Filter by category_master |
| `type` | Enum | retail, online, service, food_beverage, beauty |
| `city` | Text | Filter by kota (case-insensitive) |
| `rating_min` | Float | Minimum rating (1.0–5.0) |
| `is_open_now` | Boolean | Filter merchant yang sedang buka |
| `has_product` | Boolean | Merchant yang punya item product |
| `has_service` | Boolean | Merchant yang punya item service |
| `sort` | Enum | `rating_desc`, `newest`, `name_asc`, `most_reviewed` |

### Default Sort: `rating_desc` (merchant rating tertinggi muncul pertama)

---

## 15. SEARCH RANKING LOGIC

Merchant di-sort berdasarkan:
1. Featured merchant (jika `featured_merchant_enabled = true`) → muncul pertama
2. Rating tertinggi (default sort)
3. Review count lebih banyak → tiebreaker
4. Newest created → tiebreaker kedua

---

## 16. DATA RETENTION

```
✅ Review: Permanent (tidak auto-delete)
✅ Inquiry: Auto-close setelah `auto_close_days`, data tetap
✅ Favorites: Delete saat user atau merchant dihapus
✅ Banned merchant: Data tetap di DB, hanya tidak tampil
✅ Deleted merchant: Cascade delete items, favorites; reviews dan inquiries dipertahankan
```

---

## 17. FUTURE FEATURES (OUT OF SCOPE MVP)

- Paid featured merchant / promoted listing
- Booking/appointment system
- Transaction & payment
- Team member management per merchant
- Inventory tracking & stock alert
- Merchant verification badge
- Live chat
- Wishlist per item (bukan merchant)
- Video items
- Analytics dashboard untuk merchant owner

---

*Directory System Rules v2.0 — KAI App. Last updated: 2026-05-30*
