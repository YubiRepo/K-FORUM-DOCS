# Directory Module — Database Schema

Database schema untuk Directory module KAI App dengan Product & Service merged dan Region ID optional.

---

## Overview Relasi

```
users (dari modul auth)
  └── companies
       ├── merchants
       │    ├── merchant_categories → category_master
       │    ├── merchant_items (product + service merged)
       │    ├── reviews
       │    └── inquiries
       └── favorites
```

---

## 1. `companies`

Master data perusahaan yang dimiliki Member Pro.

```sql
CREATE TABLE companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    logo_url TEXT,
    phone VARCHAR(20),
    email VARCHAR(100),
    website VARCHAR(200),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_companies_owner_id ON companies (owner_id);
CREATE INDEX idx_companies_status ON companies (status);
```

---

## 2. `merchants`

Toko/outlet/cabang dalam company. Bisa online atau punya lokasi fisik.

```sql
CREATE TABLE merchants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES users(id),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL,
    
    -- Location (OPTIONAL untuk online)
    address VARCHAR(300),
    city VARCHAR(100),
    province VARCHAR(100),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    region_id UUID REFERENCES regions(id),
    
    -- Contact
    phone VARCHAR(20),
    email VARCHAR(100),
    whatsapp VARCHAR(20),
    instagram VARCHAR(100),
    
    -- Hours & Images
    hours JSONB,
    images JSONB,
    
    -- Status & Approval
    status VARCHAR(20) NOT NULL,
    approval_status VARCHAR(20) NOT NULL,
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    rejection_reason VARCHAR(255),
    rejection_details TEXT,
    rejected_at TIMESTAMPTZ,
    
    -- Stats (cached)
    item_count INT NOT NULL DEFAULT 0,
    review_count INT NOT NULL DEFAULT 0,
    average_rating DECIMAL(3,2) NOT NULL DEFAULT 0,
    favorite_count INT NOT NULL DEFAULT 0,
    inquiry_count INT NOT NULL DEFAULT 0,
    total_views INT NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchants_company_id ON merchants (company_id);
CREATE INDEX idx_merchants_owner_id ON merchants (owner_id);
CREATE INDEX idx_merchants_region_id ON merchants (region_id);
CREATE INDEX idx_merchants_status ON merchants (status);
CREATE INDEX idx_merchants_approval_status ON merchants (approval_status);
CREATE INDEX idx_merchants_type ON merchants (type);
CREATE INDEX idx_merchants_published ON merchants (published_at DESC) WHERE status = 'published';
```

---

## 3. `merchant_items`

Produk atau Service yang dijual merchant. **MERGED** dalam satu table dengan type field.

```sql
CREATE TABLE merchant_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    
    -- Type & Identity
    type VARCHAR(20) NOT NULL, -- 'product' atau 'service'
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    
    -- Price & Unit (For Product)
    price BIGINT,
    currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
    unit VARCHAR(50),
    stock INT, -- null = unlimited
    
    -- Price Range (For Service)
    price_range_min BIGINT,
    price_range_max BIGINT,
    
    -- Duration (For Service)
    duration_minutes INT,
    
    -- Images
    images JSONB,
    
    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'available',
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchant_items_merchant_id ON merchant_items (merchant_id);
CREATE INDEX idx_merchant_items_type ON merchant_items (type);
CREATE INDEX idx_merchant_items_status ON merchant_items (status);
```

---

## 4. `merchant_categories`

Relasi many-to-many antara merchant dan category master.

```sql
CREATE TABLE merchant_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    category_id UUID NOT NULL REFERENCES category_master(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_merchant_categories_unique 
    ON merchant_categories (merchant_id, category_id);

CREATE INDEX idx_merchant_categories_category_id 
    ON merchant_categories (category_id);
```

---

## 5. `category_master`

Master data kategori bisnis yang dikelola superadmin.

```sql
CREATE TABLE category_master (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(10),
    is_active BOOLEAN NOT NULL DEFAULT true,
    order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_category_master_active ON category_master (is_active);
CREATE INDEX idx_category_master_order ON category_master (order);
```

---

## 6. `reviews`

Review rating dan text untuk merchant.

```sql
CREATE TABLE reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    rating INT NOT NULL,
    title VARCHAR(200),
    review_text TEXT,
    aspects JSONB,
    helpful_count INT NOT NULL DEFAULT 0,
    unhelpful_count INT NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'published',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_reviews_one_per_user 
    ON reviews (merchant_id, user_id);

CREATE INDEX idx_reviews_merchant_id ON reviews (merchant_id);
CREATE INDEX idx_reviews_user_id ON reviews (user_id);
CREATE INDEX idx_reviews_rating ON reviews (rating);
```

---

## 7. `inquiries`

Pertanyaan/kontak dari user ke merchant.

```sql
CREATE TABLE inquiries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_name VARCHAR(100) NOT NULL,
    user_email VARCHAR(100) NOT NULL,
    user_phone VARCHAR(20),
    subject VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    merchant_reply TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    replied_at TIMESTAMPTZ,
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inquiries_merchant_id ON inquiries (merchant_id);
CREATE INDEX idx_inquiries_user_id ON inquiries (user_id);
CREATE INDEX idx_inquiries_status ON inquiries (status);
```

---

## 8. `favorites`

Simpan/bookmark merchant oleh user.

```sql
CREATE TABLE favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_id UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    note TEXT,
    saved_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_favorites_unique 
    ON favorites (user_id, merchant_id);

CREATE INDEX idx_favorites_user_id ON favorites (user_id);
```

---

## 9. `directory_settings`

Konfigurasi global directory module.

```sql
CREATE TABLE directory_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Merchant Settings
    merchant_require_approval BOOLEAN NOT NULL DEFAULT true,
    merchant_max_per_member INT,
    merchant_min_items_before_publish INT NOT NULL DEFAULT 0,
    merchant_min_images_before_publish INT NOT NULL DEFAULT 1,
    
    -- Item Settings
    item_require_approval BOOLEAN NOT NULL DEFAULT false,
    item_max_per_merchant INT,
    item_allow_free_price BOOLEAN NOT NULL DEFAULT false,
    
    -- Review Settings
    review_allow_reviews BOOLEAN NOT NULL DEFAULT true,
    review_require_moderation BOOLEAN NOT NULL DEFAULT true,
    review_waiting_period_days INT NOT NULL DEFAULT 0,
    review_require_purchase_verification BOOLEAN NOT NULL DEFAULT false,
    
    -- Inquiry Settings
    inquiry_allow_inquiries BOOLEAN NOT NULL DEFAULT true,
    inquiry_auto_close_days INT NOT NULL DEFAULT 30,
    
    -- Verification Settings
    verification_require_phone BOOLEAN NOT NULL DEFAULT true,
    verification_require_identity BOOLEAN NOT NULL DEFAULT false,
    
    -- Feature Settings
    features_allow_online_only BOOLEAN NOT NULL DEFAULT true,
    features_allow_multiple_categories BOOLEAN NOT NULL DEFAULT true,
    
    -- Location Settings
    location_require_location_for_physical BOOLEAN NOT NULL DEFAULT true,
    location_auto_derive_region BOOLEAN NOT NULL DEFAULT true,
    
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## Golang Structs

### Company

```go
package directory

import "time"

type Company struct {
    ID          string    `db:"id"          json:"id"`
    OwnerID     string    `db:"owner_id"    json:"owner_id"`
    Name        string    `db:"name"        json:"name"`
    Description *string   `db:"description" json:"description,omitempty"`
    LogoURL     *string   `db:"logo_url"    json:"logo_url,omitempty"`
    Phone       *string   `db:"phone"       json:"phone,omitempty"`
    Email       *string   `db:"email"       json:"email,omitempty"`
    Website     *string   `db:"website"     json:"website,omitempty"`
    Status      string    `db:"status"      json:"status"`
    CreatedAt   time.Time `db:"created_at"  json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

const (
    CompanyStatusActive   = "active"
    CompanyStatusInactive = "inactive"
    CompanyStatusBanned   = "banned"
)
```

### Merchant

```go
package directory

import (
    "encoding/json"
    "time"
)

type Merchant struct {
    ID               string           `db:"id"               json:"id"`
    CompanyID        string           `db:"company_id"       json:"company_id"`
    OwnerID          string           `db:"owner_id"         json:"owner_id"`
    Name             string           `db:"name"             json:"name"`
    Description      *string          `db:"description"      json:"description,omitempty"`
    Type             string           `db:"type"             json:"type"`
    
    // Location (optional for online)
    Address          *string          `db:"address"          json:"address,omitempty"`
    City             *string          `db:"city"             json:"city,omitempty"`
    Province         *string          `db:"province"         json:"province,omitempty"`
    Latitude         *float64         `db:"latitude"         json:"latitude,omitempty"`
    Longitude        *float64         `db:"longitude"        json:"longitude,omitempty"`
    RegionID         *string          `db:"region_id"        json:"region_id,omitempty"`
    
    // Contact
    Phone            *string          `db:"phone"            json:"phone,omitempty"`
    Email            *string          `db:"email"            json:"email,omitempty"`
    WhatsApp         *string          `db:"whatsapp"         json:"whatsapp,omitempty"`
    Instagram        *string          `db:"instagram"        json:"instagram,omitempty"`
    
    // Hours & Images
    Hours            *json.RawMessage `db:"hours"            json:"hours,omitempty"`
    Images           *json.RawMessage `db:"images"           json:"images,omitempty"`
    
    // Status & Approval
    Status           string           `db:"status"           json:"status"`
    ApprovalStatus   string           `db:"approval_status"  json:"approval_status"`
    ApprovedBy       *string          `db:"approved_by"      json:"approved_by,omitempty"`
    ApprovedAt       *time.Time       `db:"approved_at"      json:"approved_at,omitempty"`
    RejectionReason  *string          `db:"rejection_reason" json:"rejection_reason,omitempty"`
    RejectionDetails *string          `db:"rejection_details" json:"rejection_details,omitempty"`
    RejectedAt       *time.Time       `db:"rejected_at"      json:"rejected_at,omitempty"`
    
    // Stats
    ItemCount        int              `db:"item_count"       json:"item_count"`
    ReviewCount      int              `db:"review_count"     json:"review_count"`
    AverageRating    float32          `db:"average_rating"   json:"average_rating"`
    FavoriteCount    int              `db:"favorite_count"   json:"favorite_count"`
    InquiryCount     int              `db:"inquiry_count"    json:"inquiry_count"`
    TotalViews       int              `db:"total_views"      json:"total_views"`
    
    // Timestamps
    CreatedAt        time.Time        `db:"created_at"       json:"created_at"`
    PublishedAt      *time.Time       `db:"published_at"     json:"published_at,omitempty"`
    UpdatedAt        time.Time        `db:"updated_at"       json:"updated_at"`
}

const (
    MerchantTypeRetail       = "retail"
    MerchantTypeOnline       = "online"
    MerchantTypeService      = "service"
    MerchantTypeFoodBeverage = "food_beverage"
    MerchantTypeBeauty       = "beauty"
    MerchantTypeOther        = "other"
    
    MerchantStatusDraft           = "draft"
    MerchantStatusPendingApproval = "pending_approval"
    MerchantStatusPublished       = "published"
    MerchantStatusRejected        = "rejected"
    MerchantStatusArchived        = "archived"
    MerchantStatusBanned          = "banned"
)
```

### MerchantItem (Product + Service Merged)

```go
package directory

import (
    "encoding/json"
    "time"
)

type MerchantItem struct {
    ID              string           `db:"id"              json:"id"`
    MerchantID      string           `db:"merchant_id"     json:"merchant_id"`
    
    // Type & Identity
    Type            string           `db:"type"            json:"type"` // "product" atau "service"
    Name            string           `db:"name"            json:"name"`
    Description     *string          `db:"description"     json:"description,omitempty"`
    Category        string           `db:"category"        json:"category"`
    
    // Price & Unit (For Product)
    Price           *int64           `db:"price"           json:"price,omitempty"`
    Currency        string           `db:"currency"        json:"currency"`
    Unit            *string          `db:"unit"            json:"unit,omitempty"`
    Stock           *int             `db:"stock"           json:"stock,omitempty"`
    
    // Price Range (For Service)
    PriceRangeMin   *int64           `db:"price_range_min" json:"price_range_min,omitempty"`
    PriceRangeMax   *int64           `db:"price_range_max" json:"price_range_max,omitempty"`
    
    // Duration (For Service)
    DurationMinutes *int             `db:"duration_minutes" json:"duration_minutes,omitempty"`
    
    // Images
    Images          *json.RawMessage `db:"images"          json:"images,omitempty"`
    
    // Status
    Status          string           `db:"status"          json:"status"`
    
    // Timestamps
    CreatedAt       time.Time        `db:"created_at"      json:"created_at"`
    UpdatedAt       time.Time        `db:"updated_at"      json:"updated_at"`
}

const (
    ItemTypeProduct  = "product"
    ItemTypeService  = "service"
    
    ItemStatusAvailable    = "available"
    ItemStatusUnavailable  = "unavailable"
    ItemStatusDiscontinued = "discontinued"
)
```

### Review

```go
package directory

import (
    "encoding/json"
    "time"
)

type Review struct {
    ID             string           `db:"id"              json:"id"`
    MerchantID     string           `db:"merchant_id"     json:"merchant_id"`
    UserID         string           `db:"user_id"         json:"user_id"`
    Rating         int              `db:"rating"          json:"rating"`
    Title          *string          `db:"title"           json:"title,omitempty"`
    ReviewText     *string          `db:"review_text"     json:"review_text,omitempty"`
    Aspects        *json.RawMessage `db:"aspects"         json:"aspects,omitempty"`
    HelpfulCount   int              `db:"helpful_count"   json:"helpful_count"`
    UnhelpfulCount int              `db:"unhelpful_count" json:"unhelpful_count"`
    Status         string           `db:"status"          json:"status"`
    CreatedAt      time.Time        `db:"created_at"      json:"created_at"`
    UpdatedAt      time.Time        `db:"updated_at"      json:"updated_at"`
}

const (
    ReviewStatusPublished        = "published"
    ReviewStatusPendingModeration = "pending_moderation"
    ReviewStatusRejected         = "rejected"
    ReviewStatusHidden           = "hidden"
)
```

### Inquiry

```go
package directory

import "time"

type Inquiry struct {
    ID            string    `db:"id"             json:"id"`
    MerchantID    string    `db:"merchant_id"    json:"merchant_id"`
    UserID        string    `db:"user_id"        json:"user_id"`
    UserName      string    `db:"user_name"      json:"user_name"`
    UserEmail     string    `db:"user_email"     json:"user_email"`
    UserPhone     *string   `db:"user_phone"     json:"user_phone,omitempty"`
    Subject       string    `db:"subject"        json:"subject"`
    Message       string    `db:"message"        json:"message"`
    MerchantReply *string   `db:"merchant_reply" json:"merchant_reply,omitempty"`
    Status        string    `db:"status"         json:"status"`
    RepliedAt     *time.Time `db:"replied_at"    json:"replied_at,omitempty"`
    ClosedAt      *time.Time `db:"closed_at"     json:"closed_at,omitempty"`
    CreatedAt     time.Time  `db:"created_at"    json:"created_at"`
    UpdatedAt     time.Time  `db:"updated_at"    json:"updated_at"`
}

const (
    InquiryStatusPending = "pending"
    InquiryStatusReplied = "replied"
    InquiryStatusClosed  = "closed"
    InquiryStatusSpam    = "spam"
)
```

---

*Database Schema untuk Directory Module. Product & Service merged. Region ID optional.*
