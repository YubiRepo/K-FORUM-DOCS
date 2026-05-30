# Directory Module — Business Rules & Structure (REVISED)

Dokumentasi struktur dan aturan Directory module KAI App - dengan Product & Service di-merge dan Region ID optional.

---

## 1. HIERARCHY STRUCTURE

```
Member Pro (User)
  └── Company (Bisnis utama - owner)
       ├── Company Settings (name, logo, description)
       ├── Merchant 1 (Toko/outlet)
       │    ├── Categories (multiple)
       │    ├── Items/Offerings (Product + Service merged)
       │    └── Reviews
       │
       ├── Merchant 2 (Outlet lain)
       │    ├── Categories (multiple)
       │    ├── Items/Offerings
       │    └── Reviews
       │
       └── Merchant N
            ├── Categories (multiple)
            ├── Items/Offerings
            └── Reviews
```

### Key Structure:

**Company:**
- Bisnis utama yang di-own oleh Member Pro
- Satu Member Pro bisa punya multiple companies
- Contoh: "PT Impor Korea" adalah company

**Merchant:**
- Toko/outlet/cabang/sales channel dalam company
- Satu company bisa punya multiple merchants
- Contoh: "Toko Jakarta Pusat", "Online Store", "Service Center"
- Merchant bisa punya lokasi fisik (optional)

**Category:**
- Kategori bisnis tempat merchant beroperasi
- Satu merchant bisa punya multiple categories
- Contoh: "Retail", "Food & Beverage", "Service", "Beauty"

**Item/Offering:**
- Product ATAU Service yang dijual merchant (di-merge dalam satu entity)
- Satu merchant bisa punya mix of products & services
- Product: punya harga tunggal, stock (optional)
- Service: punya price range, duration

---

## 2. WHO CAN CREATE WHAT

### Member Pro Permissions:

| Action | Permission | Notes |
|--------|-----------|-------|
| Create Company | ✅ Yes | Unlimited per Member Pro |
| Create Merchant | ✅ Yes | Unlimited per company (subject to limits) |
| Add Categories to Merchant | ✅ Yes | Multiple categories per merchant |
| Add Items (Product + Service) | ✅ Yes | Unlimited per merchant (subject to limits) |
| Edit own Company/Merchant | ✅ Yes | Only owner atau authorized |
| Delete Company/Merchant | ✅ Yes | If no items |
| Archive Merchant | ✅ Yes | Keep data, hide dari directory |
| View own Directory Dashboard | ✅ Yes | See own companies & merchants |
| View all Directory | ✅ Yes | Public listing |
| Save/Favorite Merchant | ✅ Yes | Bookmark |
| Review Merchant | ✅ Yes | After conditions (if enabled) |
| Contact Merchant | ✅ Yes | Inquiry form |

### Admin Permissions:

| Action | Superadmin | Admin Regional |
|--------|-----------|-----------------|
| View all companies | ✅ Yes | ✅ Yes (filter by region) |
| View all merchants | ✅ Yes | ✅ Yes (own region) |
| Approve/Reject merchant | ✅ Yes | ✅ Yes (own region) |
| Manage categories master | ✅ Yes | ❌ No |
| Ban merchant | ✅ Yes | ✅ Yes (own region) |
| Delete merchant (force) | ✅ Yes | ❌ No |
| Manage Directory Settings | ✅ Yes | ❌ No |
| View analytics & stats | ✅ Yes | ✅ Yes (own region) |

### Regular Member Permissions:

| Action | Permission |
|--------|-----------|
| View directory | ✅ Yes |
| Search merchants | ✅ Yes |
| Filter by category/location/rating | ✅ Yes |
| View merchant detail | ✅ Yes |
| Save/Favorite merchant | ✅ Yes |
| Review merchant | ✅ Yes |
| Share merchant | ✅ Yes |
| Contact merchant (inquiry) | ✅ Yes |
| Create company/merchant | ❌ No (requires Member Pro) |

---

## 3. COMPANY STRUCTURE

### Company Object:

```json
{
  "id": "cmp_001",
  "owner_id": "user_pro_001",
  "name": "PT Impor Korea",
  "description": "Distributor produk Korea authentic dengan berbagai kategori...",
  "logo_url": "https://cdn.../logo.jpg",
  
  "contact": {
    "phone": "0212345678",
    "email": "contact@ptkore.com",
    "website": "https://ptkore.com"
  },
  
  "merchant_count": 3,
  "status": "active",
  
  "created_at": "2026-01-15T00:00:00Z",
  "updated_at": "2026-05-20T10:00:00Z"
}
```

### Company Lifecycle:

```
Created → Active → Inactive (manual) / Banned (by admin)
```

---

## 4. MERCHANT STRUCTURE

### Merchant Object:

```json
{
  "id": "mch_001",
  "company_id": "cmp_001",
  "owner_id": "user_pro_001",
  
  "name": "Toko Jakarta Pusat",
  "description": "Outlet utama kami di Jakarta dengan koleksi lengkap...",
  "type": "retail",
  
  "categories": [
    { "id": "cat_retail", "name": "Retail" },
    { "id": "cat_food", "name": "Food & Beverage" }
  ],
  
  "location": {
    "address": "Jl. Ahmad Yani No. 123, Jakarta Pusat",
    "city": "Jakarta",
    "province": "DKI Jakarta",
    "latitude": -6.2088,
    "longitude": 106.8905,
    "region_id": "region_jakarta"  // AUTO-DERIVED dari lat/lng, OPTIONAL
  },
  
  "contact": {
    "phone": "0212345678",
    "email": "jakarta@ptkore.com",
    "whatsapp": "0812345678",
    "instagram": "@ptkore_jakarta"
  },
  
  "hours": {
    "monday": { "open": "09:00", "close": "18:00", "closed": false },
    "tuesday": { "open": "09:00", "close": "18:00", "closed": false },
    "wednesday": { "open": "09:00", "close": "18:00", "closed": false },
    "thursday": { "open": "09:00", "close": "18:00", "closed": false },
    "friday": { "open": "09:00", "close": "20:00", "closed": false },
    "saturday": { "open": "10:00", "close": "20:00", "closed": false },
    "sunday": { "closed": true }
  },
  
  "stats": {
    "item_count": 53,  // product + service combined
    "review_count": 123,
    "rating": 4.5,
    "favorite_count": 234,
    "inquiry_count": 45,
    "total_views": 5420
  },
  
  "images": [
    { "url": "https://cdn.../merchant_1.jpg", "order": 1, "is_primary": true },
    { "url": "https://cdn.../merchant_2.jpg", "order": 2, "is_primary": false }
  ],
  
  "status": "published",
  "approval_status": "approved",
  
  "settings": {
    "allow_reviews": true,
    "allow_inquiries": true,
    "auto_reply_inquiries": false,
    "auto_reply_message": null,
    "featured": false
  },
  
  "created_at": "2026-02-10T00:00:00Z",
  "published_at": "2026-02-12T10:00:00Z",
  "approved_at": "2026-02-12T10:00:00Z",
  "approved_by": "admin_001"
}
```

### Merchant Types:

| Type | Description | Requires Location? |
|------|-------------|-------------------|
| retail | Toko penjualan retail | If `allow_online_only = false` |
| online | Online shop / e-commerce | No |
| service | Penyedia jasa | If `allow_online_only = false` |
| food_beverage | Restaurant / cafe | If `allow_online_only = false` |
| beauty | Salon / spa | If `allow_online_only = false` |
| other | Lainnya | Depends on setting |

**Location Rules:**
- **Online-only merchant** (type: online) → address & region_id OPTIONAL
- **Physical location merchant** (retail/service/food/beauty) → address wajib, lat/lng wajib, region_id auto-derived
- Region ID tidak perlu di-input user, di-derive otomatis dari latitude/longitude

---

## 5. MERCHANT APPROVAL FLOW

### State Diagram:

```
Member Pro create merchant → validation check
   ├─ Name required ✓
   ├─ Description required ✓
   ├─ At least 1 category ✓
   ├─ Location (depends on type & settings) ✓
   └─ At least 1 image (recommended) ✓
   
   ↓
   
require_merchant_approval setting?
   ├─ YES → status: pending_approval
   │        (Admin review)
   │        ├─ Approve → status: published
   │        └─ Reject → status: rejected (can resubmit)
   │
   └─ NO → status: auto_published (immediate)
```

---

## 6. ITEMS/OFFERINGS (Product + Service Merged)

### Single Item Object (Product):

```json
{
  "id": "item_001",
  "merchant_id": "mch_001",
  
  "type": "product",
  
  "name": "Korean Red Ginseng",
  "description": "Authentic Korean red ginseng from Geumsan...",
  "category": "Health & Wellness",
  
  "price": 150000,
  "currency": "IDR",
  "unit": "per pack",
  
  "images": [
    { "url": "https://cdn.../product_1.jpg", "order": 1, "is_primary": true },
    { "url": "https://cdn.../product_2.jpg", "order": 2, "is_primary": false }
  ],
  
  "stock": null,  // null = unlimited, atau number untuk limited stock
  "status": "available",
  
  "created_at": "2026-03-10T00:00:00Z",
  "updated_at": "2026-03-10T00:00:00Z"
}
```

### Single Item Object (Service):

```json
{
  "id": "item_002",
  "merchant_id": "mch_001",
  
  "type": "service",
  
  "name": "Korean Beauty Treatment",
  "description": "Authentic Korean skincare treatment dengan produk Korea...",
  "category": "Beauty Treatment",
  
  "price_range": {
    "min": 200000,
    "max": 500000,
    "currency": "IDR"
  },
  
  "duration_minutes": 60,
  
  "images": [
    { "url": "https://cdn.../service_1.jpg", "order": 1, "is_primary": true }
  ],
  
  "status": "available",
  
  "created_at": "2026-03-15T00:00:00Z",
  "updated_at": "2026-03-15T00:00:00Z"
}
```

### Item Rules:

```
✅ One merchant = Multiple items (product + service mix)
✅ Unlimited items per merchant (unless settings limit)
✅ Product: harga tunggal, stock optional
✅ Service: price range, duration
✅ Category: free text (bukan master list)
✅ Image minimum 1, recommended 3+
✅ Stock bisa unlimited (null) atau limited number
✅ Cannot set price = 0 (free tidak diizinkan MVP)
✅ Type: "product" atau "service"
```

---

## 7. MERCHANT CATEGORIES

### Category Master (Admin Maintained):

```json
[
  {
    "id": "cat_retail",
    "name": "Retail",
    "description": "Toko penjualan produk retail",
    "icon": "🏬",
    "is_active": true,
    "order": 1
  },
  {
    "id": "cat_food",
    "name": "Food & Beverage",
    "description": "Restaurant, cafe, food delivery",
    "icon": "🍽️",
    "is_active": true,
    "order": 2
  },
  {
    "id": "cat_service",
    "name": "Service",
    "description": "Consulting, repair, rental, dll",
    "icon": "🔧",
    "is_active": true,
    "order": 3
  },
  {
    "id": "cat_beauty",
    "name": "Beauty & Salon",
    "description": "Salon, spa, beauty treatment",
    "icon": "💄",
    "is_active": true,
    "order": 4
  }
]
```

### Merchant Category Association:

```json
{
  "merchant_id": "mch_001",
  "categories": [
    { "id": "cat_retail", "name": "Retail", "primary": true },
    { "id": "cat_food", "name": "Food & Beverage", "primary": false }
  ]
}
```

---

## 8. REVIEWS & RATINGS

### Review Object:

```json
{
  "id": "rev_001",
  "merchant_id": "mch_001",
  "user_id": "user_123",
  "user_name": "Budi Santoso",
  "user_avatar": "https://cdn.../avatar.jpg",
  
  "rating": 4,
  "title": "Produk bagus, tapi lama delivery",
  "review_text": "Produk original dan kualitas bagus. Hanya saja pengiriman agak lama...",
  
  "aspects": {
    "product_quality": 5,
    "service": 4,
    "price": 4
  },
  
  "helpful_count": 23,
  "unhelpful_count": 2,
  
  "status": "published",
  "created_at": "2026-04-20T10:00:00Z",
  "updated_at": "2026-04-20T10:00:00Z"
}
```

### Review Rules:

```
✅ One user = One review per merchant
✅ User dapat edit review
✅ Cannot delete review (dapat hide)
✅ Rating 1-5 stars (wajib)
✅ Review text optional
✅ Moderation dapat diaktifkan via settings
✅ User dapat mark helpful/unhelpful
```

---

## 9. INQUIRIES & CONTACT

### Inquiry Object:

```json
{
  "id": "inq_001",
  "merchant_id": "mch_001",
  "user_id": "user_123",
  "user_name": "Citra Dewi",
  "user_email": "citra@example.com",
  "user_phone": "0812345678",
  
  "subject": "Inquiry: Stok Korean Ginseng",
  "message": "Apakah masih ada stok untuk produk Korean Red Ginseng?",
  
  "status": "pending",
  "replied_at": null,
  "merchant_reply": null,
  
  "created_at": "2026-05-20T15:00:00Z",
  "closed_at": null
}
```

### Inquiry States:

```
pending → replied → closed
```

---

## 10. FAVORITE/SAVE MERCHANT

### Favorite Object:

```json
{
  "id": "fav_001",
  "user_id": "user_123",
  "merchant_id": "mch_001",
  
  "note": "Toko favorit untuk beli ginseng",
  "saved_at": "2026-05-20T10:00:00Z"
}
```

---

## 11. DIRECTORY SETTINGS & CONFIGURATION

### Directory Settings:

```json
{
  "merchant": {
    "require_approval": true,
    "max_per_member": null,
    "min_items_before_publish": 0,
    "min_images_before_publish": 1
  },
  
  "item": {
    "require_approval": false,
    "max_per_merchant": null,
    "allow_free_price": false
  },
  
  "review": {
    "allow_reviews": true,
    "require_moderation": true,
    "waiting_period_days": 0,
    "require_purchase_verification": false
  },
  
  "inquiry": {
    "allow_inquiries": true,
    "auto_close_days": 30,
    "allow_auto_reply": true
  },
  
  "verification": {
    "require_phone": true,
    "require_identity": false
  },
  
  "features": {
    "allow_online_only": true,
    "allow_multiple_categories": true,
    "featured_merchant_enabled": false
  },
  
  "location": {
    "require_location_for_physical": true,
    "auto_derive_region": true,
    "enable_map_view": true
  }
}
```

---

## 12. SCOPES & VISIBILITY

```
Directory Scope: GLOBAL
  - All users dapat lihat merchant
  - Merchant dari any region visible ke all users
  - Admin dapat filter by region untuk manage

User Filtering:
  - Category
  - Location/City (bukan region_id)
  - Rating
  - Merchant Name (search)
  - Favorites
```

### Merchant Visibility Status:

| Status | Visible | Notes |
|--------|---------|-------|
| draft | ❌ No | Only owner |
| pending_approval | ❌ No | Owner & admin |
| published | ✅ Yes | All users |
| rejected | ❌ No | Owner see reason |
| archived | ❌ No | Can unarchive |
| banned | ❌ No | Permanently |

---

## 13. USE CASES

### Use Case 1: Member Pro Create Company & Merchant (Online-only)

```
Scenario: Ani ingin create toko online Korea

Step 1: Create Company "Toko Online Korea"
Step 2: Create Merchant "Online Store"
        - Type: online
        - NO address required
        - Only contact: email, whatsapp
        
Step 3: Add items (mix of product & service):
        - Product: Korean Ginseng, Rp 150k
        - Product: Korean Cosmetics, Rp 75k
        - Service: Online consultation, Rp 200-500k
        
Step 4: Submit → Published (or pending approval)
        - Visible to all users
        - No location filter (karena online)
```

### Use Case 2: Member Pro Create Merchant (Physical Location)

```
Scenario: Budi create salon Korea dengan lokasi fisik

Step 1: Create Company "PT Salon Korea"
Step 2: Create Merchant "Salon Jakarta Pusat"
        - Type: beauty
        - Address: Jl. Ahmad Yani 123, Jakarta
        - Lat/Lng: auto-derive region_id = "region_jakarta"
        - Contact: phone, instagram
        - Hours: Mon-Sat 10:00-19:00, Sun closed
        
Step 3: Add items:
        - Service: Basic facial, Rp 300-500k, 60 min
        - Service: Hair treatment, Rp 200-400k, 45 min
        - Product: Korean skincare, Rp 150k (retail penjualan)
        
Step 4: Submit → Published or pending approval
```

### Use Case 3: User Browse & Review

```
Scenario: User browse directory

Step 1: Search "salon" atau filter "Beauty"
Step 2: See merchants list dengan:
        - Photo
        - Name & rating
        - Location (city)
        - Opening hours (if physical)
        
Step 3: Click merchant → detail:
        - Semua items (product + service mix)
        - Reviews & rating
        - Location on map (if physical)
        - Contact info
        
Step 4: Actions:
        - Save merchant
        - Leave review
        - Contact merchant
        - Share ke social media
```

---

## 14. RULES & CONSTRAINTS

### Location Rules:

```
Online-only merchant (type: online):
  ✅ Address OPTIONAL
  ✅ Latitude/Longitude OPTIONAL
  ✅ region_id OPTIONAL
  ✅ Only need: email, phone, instagram, website

Physical merchant (retail/service/food/beauty):
  ✅ Address REQUIRED
  ✅ Latitude/Longitude REQUIRED
  ✅ region_id AUTO-DERIVED dari lat/lng
  ✅ Hours RECOMMENDED
```

### Item Rules:

```
✅ Product + Service dalam satu table
✅ Type: "product" atau "service" (wajib)
✅ Product: price (single), unit, stock (optional)
✅ Service: price_range, duration_minutes
✅ Unlimited per merchant (unless limits)
✅ Image minimum 1, recommended 3+
✅ Cannot price = 0
```

### Admin Regional Scope:

```
Admin region can:
  ✅ View merchants dalam own region
  ✅ Approve merchants dalam own region
  ✅ Ban merchants dalam own region
  ✅ View analytics per region
  
Admin regional cannot:
  ❌ View merchants dari region lain
  ❌ Manage global settings
  ❌ Manage category master
```

---

## 15. FUTURE FEATURES

- ❌ Paid featured merchant
- ❌ Promote item/product
- ❌ Transaction system
- ❌ Merchant dashboard & analytics
- ❌ Team member management
- ❌ Inventory tracking
- ❌ Merchant verification badge
- ❌ Booking system
- ❌ Wishlist for items
- ❌ Live chat
- ❌ Video items

---

*Dokumen ini menjelaskan business rules Directory module KAI App. Product & Service merged. Region_id optional untuk online merchants. Last updated: 2026-05-28*
