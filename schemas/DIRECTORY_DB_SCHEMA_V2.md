# Directory Module — Database Schema (v2.0)

Schema lengkap untuk Directory module KAI App. PostgreSQL dialect. Product & Service merged dalam satu tabel `merchant_items`. Region ID optional (auto-derived dari koordinat).

---

## Overview Relasi

```
users (auth module)
  └── companies
       └── merchants
            ├── merchant_categories  ──→  category_master
            ├── merchant_images
            ├── merchant_items
            │    └── merchant_item_images
            ├── reviews
            │    └── review_votes
            ├── inquiries
            └── merchant_views (log)

favorites  (user ←→ merchant)
directory_settings (global config, singleton)
```

---

## 1. `companies`

Master data perusahaan milik Member Pro.

```sql
CREATE TABLE companies (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name        VARCHAR(200) NOT NULL,
    description TEXT,
    logo_url    TEXT,

    -- Contact
    phone       VARCHAR(20),
    email       VARCHAR(100),
    website     VARCHAR(200),

    -- Lifecycle
    status      VARCHAR(20)  NOT NULL DEFAULT 'active',
    -- status: active | inactive | banned

    banned_at       TIMESTAMPTZ,
    banned_by       UUID REFERENCES users(id),
    ban_reason      TEXT,

    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_companies_owner_id ON companies(owner_id);
CREATE INDEX idx_companies_status   ON companies(status);

COMMENT ON TABLE  companies             IS 'Bisnis utama / parent dari merchant';
COMMENT ON COLUMN companies.status      IS 'active | inactive | banned';
```

---

## 2. `merchants`

Toko/outlet/cabang dalam company. Bisa fisik atau online.

```sql
CREATE TABLE merchants (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID        NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    owner_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Identity
    name        VARCHAR(200) NOT NULL,
    description TEXT         NOT NULL,
    type        VARCHAR(30)  NOT NULL,
    -- type: retail | online | service | food_beverage | beauty | other

    -- Location (OPTIONAL untuk online, REQUIRED untuk fisik)
    address     TEXT,
    city        VARCHAR(100),
    province    VARCHAR(100),
    latitude    NUMERIC(10, 7),
    longitude   NUMERIC(10, 7),
    region_id   UUID REFERENCES regions(id) ON DELETE SET NULL,
    -- region_id auto-derived dari lat/lng via geo-lookup

    -- Contact
    phone       VARCHAR(20),
    email       VARCHAR(100),
    whatsapp    VARCHAR(20),
    instagram   VARCHAR(50),

    -- Operating Hours (JSONB per hari)
    -- Format: { "monday": { "open": "09:00", "close": "18:00", "closed": false }, ... }
    hours       JSONB,

    -- Merchant-level settings
    allow_reviews           BOOLEAN NOT NULL DEFAULT true,
    allow_inquiries         BOOLEAN NOT NULL DEFAULT true,
    auto_reply_enabled      BOOLEAN NOT NULL DEFAULT false,
    auto_reply_message      TEXT,

    -- Status
    status          VARCHAR(30) NOT NULL DEFAULT 'draft',
    -- status: draft | pending_approval | published | rejected | archived | banned
    approval_status VARCHAR(30) NOT NULL DEFAULT 'pending',
    -- approval_status: pending | approved | rejected

    rejection_reason    TEXT,
    ban_reason          TEXT,
    banned_at           TIMESTAMPTZ,
    banned_by           UUID REFERENCES users(id),
    approved_at         TIMESTAMPTZ,
    approved_by         UUID REFERENCES users(id),
    published_at        TIMESTAMPTZ,
    archived_at         TIMESTAMPTZ,

    -- Cached Stats (update via trigger/job)
    stat_item_count     INT NOT NULL DEFAULT 0,
    stat_review_count   INT NOT NULL DEFAULT 0,
    stat_rating         NUMERIC(3, 2) NOT NULL DEFAULT 0.00,
    stat_favorite_count INT NOT NULL DEFAULT 0,
    stat_inquiry_count  INT NOT NULL DEFAULT 0,
    stat_view_count     INT NOT NULL DEFAULT 0,

    -- Feature flags
    is_featured         BOOLEAN NOT NULL DEFAULT false,
    featured_until      TIMESTAMPTZ,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchants_company_id  ON merchants(company_id);
CREATE INDEX idx_merchants_owner_id    ON merchants(owner_id);
CREATE INDEX idx_merchants_region_id   ON merchants(region_id);
CREATE INDEX idx_merchants_status      ON merchants(status);
CREATE INDEX idx_merchants_type        ON merchants(type);
CREATE INDEX idx_merchants_city        ON merchants(city);
CREATE INDEX idx_merchants_rating      ON merchants(stat_rating DESC);
CREATE INDEX idx_merchants_published   ON merchants(published_at DESC) WHERE status = 'published';
CREATE INDEX idx_merchants_location    ON merchants USING GIST (
    ll_to_earth(latitude::float, longitude::float)
) WHERE latitude IS NOT NULL AND longitude IS NOT NULL;

-- Full text search index
CREATE INDEX idx_merchants_fts ON merchants USING GIN(
    to_tsvector('indonesian', coalesce(name,'') || ' ' || coalesce(description,''))
);

COMMENT ON TABLE  merchants             IS 'Toko/outlet/cabang dalam company';
COMMENT ON COLUMN merchants.type        IS 'retail | online | service | food_beverage | beauty | other';
COMMENT ON COLUMN merchants.status      IS 'draft | pending_approval | published | rejected | archived | banned';
COMMENT ON COLUMN merchants.region_id   IS 'Auto-derived dari lat/lng, tidak diinput manual';
COMMENT ON COLUMN merchants.hours       IS 'JSON: {monday:{open,close,closed}, tuesday:..., ...}';
```

---

## 3. `merchant_images`

Foto-foto merchant. Relasi terpisah untuk mendukung reorder.

```sql
CREATE TABLE merchant_images (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID    NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

    url         TEXT    NOT NULL,
    is_primary  BOOLEAN NOT NULL DEFAULT false,
    sort_order  INT     NOT NULL DEFAULT 0,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merchant_images_merchant_id ON merchant_images(merchant_id);
CREATE UNIQUE INDEX idx_merchant_images_primary
    ON merchant_images(merchant_id) WHERE is_primary = true;

COMMENT ON TABLE merchant_images IS 'Foto merchant, max 10 per merchant';
```

---

## 4. `category_master`

Daftar kategori yang dikelola Superadmin.

```sql
CREATE TABLE category_master (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon        VARCHAR(10),  -- emoji
    is_active   BOOLEAN      NOT NULL DEFAULT true,
    sort_order  INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_category_master_active ON category_master(is_active);
CREATE INDEX idx_category_master_order  ON category_master(sort_order);

-- Seed data default
INSERT INTO category_master (name, icon, sort_order) VALUES
    ('Retail',            '🏬', 1),
    ('Food & Beverage',   '🍽️', 2),
    ('Service',           '🔧', 3),
    ('Beauty & Salon',    '💄', 4),
    ('Health & Wellness', '🏥', 5),
    ('Education',         '📚', 6),
    ('Entertainment',     '🎭', 7),
    ('Travel & Tourism',  '✈️', 8),
    ('Finance',           '💰', 9),
    ('Technology',        '💻', 10),
    ('Fashion',           '👗', 11),
    ('Other',             '📦', 99);
```

---

## 5. `merchant_categories`

Many-to-many antara merchant dan category_master.

```sql
CREATE TABLE merchant_categories (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID    NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    category_id UUID    NOT NULL REFERENCES category_master(id) ON DELETE CASCADE,
    is_primary  BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_merchant_categories_unique
    ON merchant_categories(merchant_id, category_id);
CREATE UNIQUE INDEX idx_merchant_categories_primary
    ON merchant_categories(merchant_id) WHERE is_primary = true;
CREATE INDEX idx_merchant_categories_category_id
    ON merchant_categories(category_id);

COMMENT ON TABLE merchant_categories IS 'Many-to-many: merchant punya 1-5 kategori';
```

---

## 6. `merchant_items`

Product DAN Service dalam satu tabel (merged).

```sql
CREATE TABLE merchant_items (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID        NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,

    -- Type
    item_type   VARCHAR(10) NOT NULL,
    -- item_type: product | service

    -- Identity
    name        VARCHAR(200) NOT NULL,
    description TEXT,
    category    VARCHAR(100), -- free text, bukan dari master

    -- Pricing — Product pakai price, Service pakai price_min/price_max
    price       NUMERIC(15, 2),           -- untuk product
    price_min   NUMERIC(15, 2),           -- untuk service
    price_max   NUMERIC(15, 2),           -- untuk service
    currency    VARCHAR(5) NOT NULL DEFAULT 'IDR',
    unit        VARCHAR(50),              -- untuk product: "per pack", "per kg"

    -- Product specific
    stock       INT,                      -- null = unlimited, 0 = habis, N = ada stok

    -- Service specific
    duration_minutes INT,                 -- estimasi durasi layanan

    -- Status
    status      VARCHAR(20) NOT NULL DEFAULT 'available',
    -- status: available | unavailable | archived

    sort_order  INT NOT NULL DEFAULT 0,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT chk_item_type       CHECK (item_type IN ('product', 'service')),
    CONSTRAINT chk_product_price   CHECK (item_type != 'product' OR price IS NOT NULL),
    CONSTRAINT chk_service_price   CHECK (item_type != 'service' OR (price_min IS NOT NULL AND price_max IS NOT NULL)),
    CONSTRAINT chk_price_range     CHECK (price_min IS NULL OR price_max IS NULL OR price_max >= price_min),
    CONSTRAINT chk_price_positive  CHECK (price IS NULL OR price > 0),
    CONSTRAINT chk_price_min_pos   CHECK (price_min IS NULL OR price_min > 0)
);

CREATE INDEX idx_merchant_items_merchant_id ON merchant_items(merchant_id);
CREATE INDEX idx_merchant_items_type        ON merchant_items(item_type);
CREATE INDEX idx_merchant_items_status      ON merchant_items(status);
CREATE INDEX idx_merchant_items_order       ON merchant_items(merchant_id, sort_order);

COMMENT ON TABLE  merchant_items             IS 'Product dan Service dalam satu tabel';
COMMENT ON COLUMN merchant_items.item_type   IS 'product | service';
COMMENT ON COLUMN merchant_items.price       IS 'Untuk product: harga tunggal';
COMMENT ON COLUMN merchant_items.price_min   IS 'Untuk service: harga minimum';
COMMENT ON COLUMN merchant_items.price_max   IS 'Untuk service: harga maksimum';
COMMENT ON COLUMN merchant_items.stock       IS 'null=unlimited, 0=habis, N=stok ada (product only)';
```

---

## 7. `merchant_item_images`

Foto per item. Relasi terpisah untuk mendukung reorder.

```sql
CREATE TABLE merchant_item_images (
    id      UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    item_id UUID    NOT NULL REFERENCES merchant_items(id) ON DELETE CASCADE,

    url         TEXT    NOT NULL,
    is_primary  BOOLEAN NOT NULL DEFAULT false,
    sort_order  INT     NOT NULL DEFAULT 0,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_item_images_item_id ON merchant_item_images(item_id);
CREATE UNIQUE INDEX idx_item_images_primary
    ON merchant_item_images(item_id) WHERE is_primary = true;
```

---

## 8. `reviews`

Review dari member ke merchant. 1 user = 1 review per merchant.

```sql
CREATE TABLE reviews (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID    NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    user_id     UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Content
    rating      INT         NOT NULL,
    title       VARCHAR(200),
    review_text TEXT,

    -- Aspect ratings (JSONB)
    -- Format: { "product_quality": 4, "service": 5, "price": 3 }
    aspects     JSONB,

    -- Engagement
    helpful_count   INT NOT NULL DEFAULT 0,
    unhelpful_count INT NOT NULL DEFAULT 0,

    -- Moderation
    status          VARCHAR(20) NOT NULL DEFAULT 'published',
    -- status: pending | published | rejected | hidden
    moderated_by    UUID REFERENCES users(id),
    moderated_at    TIMESTAMPTZ,
    moderation_note TEXT,

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_rating_range CHECK (rating BETWEEN 1 AND 5)
);

-- 1 user hanya bisa 1 review per merchant
CREATE UNIQUE INDEX idx_reviews_one_per_user
    ON reviews(merchant_id, user_id);

CREATE INDEX idx_reviews_merchant_id ON reviews(merchant_id);
CREATE INDEX idx_reviews_user_id     ON reviews(user_id);
CREATE INDEX idx_reviews_status      ON reviews(status);
CREATE INDEX idx_reviews_rating      ON reviews(rating);

COMMENT ON TABLE  reviews        IS '1 user = 1 review per merchant';
COMMENT ON COLUMN reviews.status IS 'pending | published | rejected | hidden';
```

---

## 9. `review_votes`

Helpful/unhelpful vote pada review.

```sql
CREATE TABLE review_votes (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id   UUID    NOT NULL REFERENCES reviews(id) ON DELETE CASCADE,
    user_id     UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote        VARCHAR(10) NOT NULL,
    -- vote: helpful | unhelpful
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_vote CHECK (vote IN ('helpful', 'unhelpful'))
);

-- 1 user 1 vote per review
CREATE UNIQUE INDEX idx_review_votes_unique
    ON review_votes(review_id, user_id);

CREATE INDEX idx_review_votes_review_id ON review_votes(review_id);
```

---

## 10. `inquiries`

Pertanyaan/pesan dari user ke merchant.

```sql
CREATE TABLE inquiries (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID    NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    user_id     UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Inquiry content
    subject     VARCHAR(200) NOT NULL,
    message     TEXT         NOT NULL,

    -- Reply
    merchant_reply  TEXT,
    replied_at      TIMESTAMPTZ,

    -- Lifecycle
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    -- status: pending | replied | closed
    closed_at   TIMESTAMPTZ,
    closed_by   VARCHAR(20), -- 'owner' | 'system' | 'admin'

    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_inquiry_status CHECK (status IN ('pending', 'replied', 'closed'))
);

CREATE INDEX idx_inquiries_merchant_id ON inquiries(merchant_id);
CREATE INDEX idx_inquiries_user_id     ON inquiries(user_id);
CREATE INDEX idx_inquiries_status      ON inquiries(status);
CREATE INDEX idx_inquiries_created_at  ON inquiries(created_at DESC);
```

---

## 11. `favorites`

Merchant yang di-save/bookmark oleh user.

```sql
CREATE TABLE favorites (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    merchant_id UUID    NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    note        TEXT,
    saved_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_favorites_unique
    ON favorites(user_id, merchant_id);
CREATE INDEX idx_favorites_user_id     ON favorites(user_id);
CREATE INDEX idx_favorites_merchant_id ON favorites(merchant_id);
```

---

## 12. `merchant_views`

Log view untuk stat `total_views`. Deduplicate per user per hari.

```sql
CREATE TABLE merchant_views (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id UUID    NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    user_id     UUID    REFERENCES users(id) ON DELETE SET NULL, -- null = guest
    ip_address  INET,
    viewed_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    view_date   DATE    NOT NULL DEFAULT CURRENT_DATE -- untuk dedup per hari
);

-- Deduplicate: 1 user = 1 view per merchant per hari
CREATE UNIQUE INDEX idx_merchant_views_daily_user
    ON merchant_views(merchant_id, user_id, view_date)
    WHERE user_id IS NOT NULL;

CREATE INDEX idx_merchant_views_merchant_id ON merchant_views(merchant_id);
CREATE INDEX idx_merchant_views_viewed_at   ON merchant_views(viewed_at DESC);
```

---

## 13. `directory_settings`

Konfigurasi global directory. Singleton (hanya 1 row).

```sql
CREATE TABLE directory_settings (
    id  INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),  -- singleton

    -- Merchant
    merchant_require_approval           BOOLEAN     NOT NULL DEFAULT true,
    merchant_max_per_member             INT,                          -- null = unlimited
    merchant_min_items_before_publish   INT         NOT NULL DEFAULT 0,
    merchant_min_images_before_publish  INT         NOT NULL DEFAULT 1,

    -- Item
    item_require_approval   BOOLEAN NOT NULL DEFAULT false,
    item_max_per_merchant   INT,                                      -- null = unlimited
    item_allow_free_price   BOOLEAN NOT NULL DEFAULT false,

    -- Review
    review_allow_reviews                BOOLEAN NOT NULL DEFAULT true,
    review_require_moderation           BOOLEAN NOT NULL DEFAULT true,
    review_waiting_period_days          INT     NOT NULL DEFAULT 0,
    review_require_purchase_verification BOOLEAN NOT NULL DEFAULT false,

    -- Inquiry
    inquiry_allow_inquiries     BOOLEAN NOT NULL DEFAULT true,
    inquiry_auto_close_days     INT     NOT NULL DEFAULT 30,
    inquiry_allow_auto_reply    BOOLEAN NOT NULL DEFAULT true,

    -- Verification
    verification_require_phone      BOOLEAN NOT NULL DEFAULT true,
    verification_require_identity   BOOLEAN NOT NULL DEFAULT false,

    -- Features
    features_allow_online_only          BOOLEAN NOT NULL DEFAULT true,
    features_allow_multiple_categories  BOOLEAN NOT NULL DEFAULT true,
    features_featured_merchant_enabled  BOOLEAN NOT NULL DEFAULT false,

    -- Location
    location_require_location_for_physical  BOOLEAN NOT NULL DEFAULT true,
    location_auto_derive_region             BOOLEAN NOT NULL DEFAULT true,
    location_enable_map_view                BOOLEAN NOT NULL DEFAULT true,

    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_by  UUID REFERENCES users(id)
);

-- Insert singleton row
INSERT INTO directory_settings (id) VALUES (1) ON CONFLICT DO NOTHING;
```

---

## 14. Triggers & Functions

### 14.1 Update `updated_at` Otomatis

```sql
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply ke semua tabel relevan
CREATE TRIGGER trg_companies_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_merchants_updated_at
    BEFORE UPDATE ON merchants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_merchant_items_updated_at
    BEFORE UPDATE ON merchant_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_reviews_updated_at
    BEFORE UPDATE ON reviews
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_inquiries_updated_at
    BEFORE UPDATE ON inquiries
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
```

### 14.2 Update Merchant Stats Setelah Review

```sql
CREATE OR REPLACE FUNCTION refresh_merchant_review_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE merchants
    SET
        stat_review_count = (
            SELECT COUNT(*) FROM reviews
            WHERE merchant_id = COALESCE(NEW.merchant_id, OLD.merchant_id)
              AND status = 'published'
        ),
        stat_rating = COALESCE((
            SELECT ROUND(AVG(rating)::NUMERIC, 2) FROM reviews
            WHERE merchant_id = COALESCE(NEW.merchant_id, OLD.merchant_id)
              AND status = 'published'
        ), 0.00),
        updated_at = NOW()
    WHERE id = COALESCE(NEW.merchant_id, OLD.merchant_id);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_review_stats
    AFTER INSERT OR UPDATE OR DELETE ON reviews
    FOR EACH ROW EXECUTE FUNCTION refresh_merchant_review_stats();
```

### 14.3 Update Merchant Favorite Count

```sql
CREATE OR REPLACE FUNCTION refresh_merchant_favorite_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE merchants
    SET
        stat_favorite_count = (
            SELECT COUNT(*) FROM favorites
            WHERE merchant_id = COALESCE(NEW.merchant_id, OLD.merchant_id)
        ),
        updated_at = NOW()
    WHERE id = COALESCE(NEW.merchant_id, OLD.merchant_id);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_favorite_count
    AFTER INSERT OR DELETE ON favorites
    FOR EACH ROW EXECUTE FUNCTION refresh_merchant_favorite_count();
```

### 14.4 Update Merchant Item Count

```sql
CREATE OR REPLACE FUNCTION refresh_merchant_item_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE merchants
    SET
        stat_item_count = (
            SELECT COUNT(*) FROM merchant_items
            WHERE merchant_id = COALESCE(NEW.merchant_id, OLD.merchant_id)
              AND status != 'archived'
        ),
        updated_at = NOW()
    WHERE id = COALESCE(NEW.merchant_id, OLD.merchant_id);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_item_count
    AFTER INSERT OR UPDATE OR DELETE ON merchant_items
    FOR EACH ROW EXECUTE FUNCTION refresh_merchant_item_count();
```

---

## 15. Golang Structs

```go
package directory

import (
    "time"
    "github.com/google/uuid"
)

// --- Company ---

type Company struct {
    ID          uuid.UUID  `db:"id"          json:"id"`
    OwnerID     uuid.UUID  `db:"owner_id"    json:"owner_id"`
    Name        string     `db:"name"        json:"name"`
    Description *string    `db:"description" json:"description,omitempty"`
    LogoURL     *string    `db:"logo_url"    json:"logo_url,omitempty"`
    Phone       *string    `db:"phone"       json:"phone,omitempty"`
    Email       *string    `db:"email"       json:"email,omitempty"`
    Website     *string    `db:"website"     json:"website,omitempty"`
    Status      string     `db:"status"      json:"status"`
    BannedAt    *time.Time `db:"banned_at"   json:"banned_at,omitempty"`
    BanReason   *string    `db:"ban_reason"  json:"ban_reason,omitempty"`
    CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
    UpdatedAt   time.Time  `db:"updated_at"  json:"updated_at"`
}

// --- Merchant ---

type Merchant struct {
    ID          uuid.UUID  `db:"id"           json:"id"`
    CompanyID   uuid.UUID  `db:"company_id"   json:"company_id"`
    OwnerID     uuid.UUID  `db:"owner_id"     json:"owner_id"`
    Name        string     `db:"name"         json:"name"`
    Description string     `db:"description"  json:"description"`
    Type        string     `db:"type"         json:"type"`

    // Location
    Address   *string    `db:"address"   json:"address,omitempty"`
    City      *string    `db:"city"      json:"city,omitempty"`
    Province  *string    `db:"province"  json:"province,omitempty"`
    Latitude  *float64   `db:"latitude"  json:"latitude,omitempty"`
    Longitude *float64   `db:"longitude" json:"longitude,omitempty"`
    RegionID  *uuid.UUID `db:"region_id" json:"region_id,omitempty"`

    // Contact
    Phone     *string `db:"phone"     json:"phone,omitempty"`
    Email     *string `db:"email"     json:"email,omitempty"`
    Whatsapp  *string `db:"whatsapp"  json:"whatsapp,omitempty"`
    Instagram *string `db:"instagram" json:"instagram,omitempty"`

    Hours *map[string]DayHours `db:"hours" json:"hours,omitempty"`

    // Settings
    AllowReviews       bool    `db:"allow_reviews"        json:"allow_reviews"`
    AllowInquiries     bool    `db:"allow_inquiries"      json:"allow_inquiries"`
    AutoReplyEnabled   bool    `db:"auto_reply_enabled"   json:"auto_reply_enabled"`
    AutoReplyMessage   *string `db:"auto_reply_message"   json:"auto_reply_message,omitempty"`

    // Status
    Status         string  `db:"status"          json:"status"`
    ApprovalStatus string  `db:"approval_status" json:"approval_status"`
    RejectionReason *string `db:"rejection_reason" json:"rejection_reason,omitempty"`
    BanReason       *string `db:"ban_reason"       json:"ban_reason,omitempty"`

    // Timestamps
    ApprovedAt  *time.Time `db:"approved_at"  json:"approved_at,omitempty"`
    PublishedAt *time.Time `db:"published_at" json:"published_at,omitempty"`
    ArchivedAt  *time.Time `db:"archived_at"  json:"archived_at,omitempty"`
    BannedAt    *time.Time `db:"banned_at"    json:"banned_at,omitempty"`

    // Stats
    StatItemCount     int     `db:"stat_item_count"     json:"item_count"`
    StatReviewCount   int     `db:"stat_review_count"   json:"review_count"`
    StatRating        float64 `db:"stat_rating"         json:"rating"`
    StatFavoriteCount int     `db:"stat_favorite_count" json:"favorite_count"`
    StatInquiryCount  int     `db:"stat_inquiry_count"  json:"inquiry_count"`
    StatViewCount     int     `db:"stat_view_count"     json:"view_count"`

    IsFeatured bool `db:"is_featured" json:"is_featured"`

    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type DayHours struct {
    Open   string `json:"open"`
    Close  string `json:"close"`
    Closed bool   `json:"closed"`
}

// --- MerchantItem ---

type MerchantItem struct {
    ID          uuid.UUID `db:"id"          json:"id"`
    MerchantID  uuid.UUID `db:"merchant_id" json:"merchant_id"`
    ItemType    string    `db:"item_type"   json:"type"`
    Name        string    `db:"name"        json:"name"`
    Description *string   `db:"description" json:"description,omitempty"`
    Category    *string   `db:"category"    json:"category,omitempty"`

    // Product fields
    Price    *float64 `db:"price"    json:"price,omitempty"`
    Unit     *string  `db:"unit"     json:"unit,omitempty"`
    Stock    *int     `db:"stock"    json:"stock,omitempty"`

    // Service fields
    PriceMin        *float64 `db:"price_min"        json:"price_min,omitempty"`
    PriceMax        *float64 `db:"price_max"        json:"price_max,omitempty"`
    DurationMinutes *int     `db:"duration_minutes" json:"duration_minutes,omitempty"`

    Currency  string `db:"currency" json:"currency"`
    Status    string `db:"status"   json:"status"`
    SortOrder int    `db:"sort_order" json:"sort_order"`

    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// --- Review ---

type Review struct {
    ID          uuid.UUID `db:"id"          json:"id"`
    MerchantID  uuid.UUID `db:"merchant_id" json:"merchant_id"`
    UserID      uuid.UUID `db:"user_id"     json:"user_id"`
    Rating      int       `db:"rating"      json:"rating"`
    Title       *string   `db:"title"       json:"title,omitempty"`
    ReviewText  *string   `db:"review_text" json:"review_text,omitempty"`
    Aspects     *map[string]int `db:"aspects" json:"aspects,omitempty"`
    HelpfulCount   int    `db:"helpful_count"   json:"helpful_count"`
    UnhelpfulCount int    `db:"unhelpful_count" json:"unhelpful_count"`
    Status      string    `db:"status"      json:"status"`
    CreatedAt   time.Time `db:"created_at"  json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"  json:"updated_at"`
}

// --- Inquiry ---

type Inquiry struct {
    ID            uuid.UUID  `db:"id"             json:"id"`
    MerchantID    uuid.UUID  `db:"merchant_id"    json:"merchant_id"`
    UserID        uuid.UUID  `db:"user_id"        json:"user_id"`
    Subject       string     `db:"subject"        json:"subject"`
    Message       string     `db:"message"        json:"message"`
    MerchantReply *string    `db:"merchant_reply" json:"merchant_reply,omitempty"`
    RepliedAt     *time.Time `db:"replied_at"     json:"replied_at,omitempty"`
    Status        string     `db:"status"         json:"status"`
    ClosedAt      *time.Time `db:"closed_at"      json:"closed_at,omitempty"`
    CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
    UpdatedAt     time.Time  `db:"updated_at"     json:"updated_at"`
}

// --- Favorite ---

type Favorite struct {
    ID         uuid.UUID `db:"id"          json:"id"`
    UserID     uuid.UUID `db:"user_id"     json:"user_id"`
    MerchantID uuid.UUID `db:"merchant_id" json:"merchant_id"`
    Note       *string   `db:"note"        json:"note,omitempty"`
    SavedAt    time.Time `db:"saved_at"    json:"saved_at"`
}
```

---

*Directory DB Schema v2.0 — KAI App. PostgreSQL. Last updated: 2026-05-30*
