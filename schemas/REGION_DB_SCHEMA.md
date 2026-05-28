# Database Schema — Region Module

**Status:** Draft v1  
**Stack:** Golang + PostgreSQL  
**Last Updated:** 2026-05-26

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [Tabel: regions](#1-regions)
3. [Tabel: region_memberships](#2-region_memberships)
4. [Tabel: region_invitations](#3-region_invitations)
5. [Enum Values](#enum-values)
6. [Golang Structs](#golang-structs)
7. [Sample Queries](#sample-queries)

---

## Overview Relasi

```
regions
  ├── region_memberships ────── users (one user can be member/admin)
  └── region_invitations ────── users (inviter)
                           └── users (target, via email)
```

### Entity Relationship (Text Diagram)

```
┌─────────────┐
│   users     │
├─────────────┤
│ id (PK)     │
│ email       │
│ name        │
└──────┬──────┘
       │
       ├─── region_memberships (user_id FK) ──┐
       │                                       │
       └─── region_invitations (invited_by) ──┤
                                              │
                                    ┌─────────▼────────┐
                                    │   regions        │
                                    ├──────────────────┤
                                    │ id (PK)          │
                                    │ name (UNIQUE)    │
                                    │ slug (UNIQUE)    │
                                    │ status           │
                                    │ created_by (FK)  │
                                    └──────────────────┘
                                              │
                                    region_invitations (region_id FK)
```

---

## 1. `regions`

Master data untuk semua region di platform. Dikelola hanya oleh superadmin.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|-------|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `name` | `VARCHAR(100)` | NOT NULL, UNIQUE | Contoh: "KAI Jakarta", "KAI Surabaya" |
| `slug` | `VARCHAR(50)` | NOT NULL, UNIQUE | URL-friendly, lowercase, no spaces |
| `description` | `TEXT` | NULLABLE | Deskripsi wilayah, max 500 chars |
| `image_url` | `TEXT` | NULLABLE | URL logo/cover region |
| `status` | `VARCHAR(20)` | NOT NULL, DEFAULT `'active'` | `active` atau `inactive` |
| `created_by` | `UUID` | NOT NULL, FK → `users.id` | Superadmin yang create |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | Auto-update via trigger |

**DDL Indexes & Constraints:**

```sql
-- Unique constraints
CREATE UNIQUE INDEX idx_regions_name_unique ON regions(LOWER(name));
CREATE UNIQUE INDEX idx_regions_slug_unique ON regions(LOWER(slug));

-- Query by status (for listing active regions only)
CREATE INDEX idx_regions_status ON regions(status);

-- Query by created_by (untuk tracking superadmin actions)
CREATE INDEX idx_regions_created_by ON regions(created_by);

-- Full-text search or simple text search
CREATE INDEX idx_regions_name_trgm ON regions USING GIN(name gin_trgm_ops);
```

---

## 2. `region_memberships`

Relasi user dengan region. Menyimpan membership status, role, dan approval info.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|-------|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `user_id` | `UUID` | NOT NULL, FK → `users.id`, ON DELETE CASCADE | |
| `region_id` | `UUID` | NOT NULL, FK → `regions.id`, ON DELETE CASCADE | |
| `role` | `VARCHAR(20)` | NOT NULL | `admin` atau `member` |
| `status` | `VARCHAR(20)` | NOT NULL, DEFAULT `'active'` | `active`, `pending_approval`, `rejected` |
| `joined_at` | `TIMESTAMPTZ` | NULLABLE | Saat user jadi active member |
| `approval_notes` | `TEXT` | NULLABLE | Catatan admin saat approve |
| `rejection_reason` | `TEXT` | NULLABLE | Alasan ditolak |
| `approved_by` | `UUID` | NULLABLE, FK → `users.id` | Admin/superadmin yang approve/reject |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | Saat record dibuat (request atau invite accepted) |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |

**DDL Indexes & Constraints:**

```sql
-- Enforce: 1 active membership per user per region
CREATE UNIQUE INDEX idx_region_memberships_active_unique
  ON region_memberships(user_id, region_id)
  WHERE status = 'active';

-- Query: find all members in a region
CREATE INDEX idx_region_memberships_region_id
  ON region_memberships(region_id, status);

-- Query: find all regions user is member of
CREATE INDEX idx_region_memberships_user_id
  ON region_memberships(user_id, status);

-- Query: find pending approvals by region
CREATE INDEX idx_region_memberships_pending
  ON region_memberships(region_id, status)
  WHERE status = 'pending_approval';

-- Query: track rejections
CREATE INDEX idx_region_memberships_rejected
  ON region_memberships(region_id, status)
  WHERE status = 'rejected';

-- Foreign key for approved_by
CREATE INDEX idx_region_memberships_approved_by
  ON region_memberships(approved_by)
  WHERE approved_by IS NOT NULL;
```

**Foreign Key Constraints:**

```sql
ALTER TABLE region_memberships
ADD CONSTRAINT fk_region_memberships_user_id
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE region_memberships
ADD CONSTRAINT fk_region_memberships_region_id
  FOREIGN KEY (region_id) REFERENCES regions(id) ON DELETE CASCADE;

ALTER TABLE region_memberships
ADD CONSTRAINT fk_region_memberships_approved_by
  FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL;
```

---

## 3. `region_invitations`

Undangan untuk bergabung region via email. Invitation dapat accept dalam 24 jam.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|-------|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `region_id` | `UUID` | NOT NULL, FK → `regions.id`, ON DELETE CASCADE | Region yang mengundang |
| `email` | `VARCHAR(255)` | NOT NULL | Email target undangan |
| `invited_by` | `UUID` | NOT NULL, FK → `users.id` | Superadmin atau admin yang invite |
| `status` | `VARCHAR(20)` | NOT NULL, DEFAULT `'pending'` | `pending`, `accepted`, `rejected`, `expired` |
| `token` | `VARCHAR(255)` | NULLABLE, UNIQUE | JWT atau hash untuk email link verification |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | Saat invitation dibuat |
| `expires_at` | `TIMESTAMPTZ` | NOT NULL | created_at + 24 jam |
| `accepted_at` | `TIMESTAMPTZ` | NULLABLE | Saat user accept |
| `rejected_at` | `TIMESTAMPTZ` | NULLABLE | Saat user reject |

**DDL Indexes & Constraints:**

```sql
-- Prevent duplicate pending invitations to same email in same region
CREATE UNIQUE INDEX idx_region_invitations_pending_unique
  ON region_invitations(region_id, email)
  WHERE status = 'pending';

-- Token untuk link verification
CREATE UNIQUE INDEX idx_region_invitations_token
  ON region_invitations(token)
  WHERE token IS NOT NULL;

-- Query: find pending invitations for email
CREATE INDEX idx_region_invitations_email
  ON region_invitations(email, status);

-- Query: find expiring invitations (untuk cleanup atau reminder)
CREATE INDEX idx_region_invitations_expires_at
  ON region_invitations(expires_at);

-- Query: find invitations sent by user (untuk tracking)
CREATE INDEX idx_region_invitations_invited_by
  ON region_invitations(invited_by);

-- Query: find pending invitations in region
CREATE INDEX idx_region_invitations_region_pending
  ON region_invitations(region_id, status)
  WHERE status = 'pending';
```

**Foreign Key Constraints:**

```sql
ALTER TABLE region_invitations
ADD CONSTRAINT fk_region_invitations_region_id
  FOREIGN KEY (region_id) REFERENCES regions(id) ON DELETE CASCADE;

ALTER TABLE region_invitations
ADD CONSTRAINT fk_region_invitations_invited_by
  FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE SET NULL;
```

---

## Enum Values

### `regions.status`

| Value | Keterangan |
|-------|-----------|
| `active` | Region tersedia, tampil di listing, members bisa join |
| `inactive` | Region disembunyikan, tidak tampil di listing, existing members tetap tergabung |

---

### `region_memberships.status`

| Value | Keterangan |
|-------|-----------|
| `active` | Member aktif di region, punya akses penuh |
| `pending_approval` | Member request join, menunggu approval dari admin |
| `rejected` | Admin tolak request join, member bisa request lagi nanti |

---

### `region_memberships.role`

| Value | Keterangan |
|-------|-----------|
| `admin` | Admin region, bisa manage members, invite, approve/reject |
| `member` | Regular member, dapat akses region content |

---

### `region_invitations.status`

| Value | Keterangan |
|-------|-----------|
| `pending` | Invitation baru, menunggu user accept/reject |
| `accepted` | User accept invitation, RegionMembership created |
| `rejected` | User reject invitation, tidak bisa accept lagi |
| `expired` | 24 jam passed, expired otomatis, tidak bisa accept dari email link |

---

## Golang Structs

### `Region`

```go
package region

import "time"

// Region merepresentasikan satu region di platform.
type Region struct {
    ID        string    `db:"id"         json:"id"`
    Name      string    `db:"name"       json:"name"`
    Slug      string    `db:"slug"       json:"slug"`
    Description *string `db:"description" json:"description,omitempty"`
    ImageURL  *string   `db:"image_url"  json:"image_url,omitempty"`
    Status    string    `db:"status"     json:"status"`  // active, inactive
    CreatedBy string    `db:"created_by" json:"created_by"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// RegionWithMemberCount digunakan untuk response dengan member count.
type RegionWithMemberCount struct {
    Region
    MemberCount int `db:"member_count" json:"member_count"`
}

// RegionStatus enum
const (
    RegionStatusActive   = "active"
    RegionStatusInactive = "inactive"
)
```

---

### `RegionMembership`

```go
package region

import "time"

// RegionMembership merepresentasikan membership user di region.
type RegionMembership struct {
    ID               string     `db:"id"                 json:"id"`
    UserID           string     `db:"user_id"            json:"user_id"`
    RegionID         string     `db:"region_id"          json:"region_id"`
    Role             string     `db:"role"               json:"role"`
    Status           string     `db:"status"             json:"status"`
    JoinedAt         *time.Time `db:"joined_at"          json:"joined_at,omitempty"`
    ApprovalNotes    *string    `db:"approval_notes"     json:"approval_notes,omitempty"`
    RejectionReason  *string    `db:"rejection_reason"   json:"rejection_reason,omitempty"`
    ApprovedBy       *string    `db:"approved_by"        json:"approved_by,omitempty"`
    CreatedAt        time.Time  `db:"created_at"         json:"created_at"`
    UpdatedAt        time.Time  `db:"updated_at"         json:"updated_at"`
}

// RegionMembershipWithUser digunakan untuk response dengan user details.
type RegionMembershipWithUser struct {
    RegionMembership
    User *UserInfo `db:"-" json:"user,omitempty"`
}

// UserInfo adalah subset dari user untuk ditampilkan dalam membership.
type UserInfo struct {
    ID     string  `json:"id"`
    Name   string  `json:"name"`
    Email  string  `json:"email"`
    Avatar *string `json:"avatar,omitempty"`
}

// RegionMembershipRole enum
const (
    RoleMember = "member"
    RoleAdmin  = "admin"
)

// RegionMembershipStatus enum
const (
    StatusActive            = "active"
    StatusPendingApproval   = "pending_approval"
    StatusRejected          = "rejected"
)
```

---

### `RegionInvitation`

```go
package region

import "time"

// RegionInvitation merepresentasikan undangan bergabung region via email.
type RegionInvitation struct {
    ID        string     `db:"id"          json:"id"`
    RegionID  string     `db:"region_id"   json:"region_id"`
    Email     string     `db:"email"       json:"email"`
    InvitedBy string     `db:"invited_by"  json:"invited_by"`
    Status    string     `db:"status"      json:"status"`
    Token     *string    `db:"token"       json:"token,omitempty"`
    CreatedAt time.Time  `db:"created_at"  json:"created_at"`
    ExpiresAt time.Time  `db:"expires_at"  json:"expires_at"`
    AcceptedAt *time.Time `db:"accepted_at" json:"accepted_at,omitempty"`
    RejectedAt *time.Time `db:"rejected_at" json:"rejected_at,omitempty"`
}

// RegionInvitationWithDetails digunakan untuk response dengan region/inviter details.
type RegionInvitationWithDetails struct {
    RegionInvitation
    RegionName      string  `db:"region_name"        json:"region_name"`
    InvitedByName   string  `db:"invited_by_name"    json:"invited_by_name"`
    InvitedByAvatar *string `db:"invited_by_avatar"  json:"invited_by_avatar,omitempty"`
}

// RegionInvitationStatus enum
const (
    InvitationStatusPending  = "pending"
    InvitationStatusAccepted = "accepted"
    InvitationStatusRejected = "rejected"
    InvitationStatusExpired  = "expired"
)
```

---

## Sample Queries

### Query 1: Get All Active Regions

```sql
SELECT id, name, slug, description, image_url, status, created_at
FROM regions
WHERE status = 'active'
ORDER BY created_at DESC;
```

### Query 2: Get Region with Member Count

```sql
SELECT
    r.id, r.name, r.slug, r.description, r.image_url, r.status,
    COUNT(CASE WHEN rm.status = 'active' THEN 1 END) as member_count
FROM regions r
LEFT JOIN region_memberships rm ON r.id = rm.region_id
WHERE r.status = 'active'
GROUP BY r.id
ORDER BY r.created_at DESC;
```

### Query 3: Get User's Region (if any)

```sql
SELECT r.*, rm.role, rm.status, rm.joined_at
FROM region_memberships rm
INNER JOIN regions r ON rm.region_id = r.id
WHERE rm.user_id = $1 AND rm.status = 'active'
LIMIT 1;
```

### Query 4: Get All Members in Region with Status

```sql
SELECT
    rm.id, rm.user_id, rm.role, rm.status, rm.joined_at,
    u.name, u.email, u.avatar
FROM region_memberships rm
INNER JOIN users u ON rm.user_id = u.id
WHERE rm.region_id = $1
ORDER BY rm.status DESC, rm.joined_at DESC;
```

### Query 5: Get Pending Approvals in Region

```sql
SELECT
    rm.id, rm.user_id, u.name, u.email,
    rm.created_at as requested_at
FROM region_memberships rm
INNER JOIN users u ON rm.user_id = u.id
WHERE rm.region_id = $1 AND rm.status = 'pending_approval'
ORDER BY rm.created_at ASC;
```

### Query 6: Get User's All Memberships (not just active)

```sql
SELECT
    rm.*, r.name as region_name, r.slug
FROM region_memberships rm
INNER JOIN regions r ON rm.region_id = r.id
WHERE rm.user_id = $1
ORDER BY rm.created_at DESC;
```

### Query 7: Get Pending Invitations for Email

```sql
SELECT
    ri.*, r.name as region_name, r.image_url,
    u.name as invited_by_name, u.avatar as invited_by_avatar
FROM region_invitations ri
INNER JOIN regions r ON ri.region_id = r.id
INNER JOIN users u ON ri.invited_by = u.id
WHERE ri.email = $1 AND ri.status = 'pending'
ORDER BY ri.created_at DESC;
```

### Query 8: Check if User Already Member of Region

```sql
SELECT 1
FROM region_memberships
WHERE user_id = $1 AND region_id = $2 AND status = 'active'
LIMIT 1;
```

### Query 9: Get Expired Invitations (for cleanup)

```sql
SELECT id, email, region_id
FROM region_invitations
WHERE status = 'pending' AND expires_at < NOW();
```

### Query 10: Approve Member Request

```sql
UPDATE region_memberships
SET
    status = 'active',
    joined_at = NOW(),
    approved_by = $1,
    updated_at = NOW()
WHERE id = $2
RETURNING *;
```

### Query 11: Remove Member from Region

```sql
UPDATE region_memberships
SET status = 'inactive', updated_at = NOW()
WHERE user_id = $1 AND region_id = $2
RETURNING *;
```

### Query 12: Get Region Admin(s)

```sql
SELECT u.*
FROM region_memberships rm
INNER JOIN users u ON rm.user_id = u.id
WHERE rm.region_id = $1 AND rm.role = 'admin' AND rm.status = 'active';
```

---

## Migrations (Pseudocode)

```sql
-- Migration 1: Create regions table
CREATE TABLE regions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL UNIQUE,
  slug VARCHAR(50) NOT NULL UNIQUE,
  description TEXT,
  image_url TEXT,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Migration 2: Create region_memberships table
CREATE TABLE region_memberships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  region_id UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
  role VARCHAR(20) NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  joined_at TIMESTAMPTZ,
  approval_notes TEXT,
  rejection_reason TEXT,
  approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Migration 3: Create region_invitations table
CREATE TABLE region_invitations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  region_id UUID NOT NULL REFERENCES regions(id) ON DELETE CASCADE,
  email VARCHAR(255) NOT NULL,
  invited_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'pending',
  token VARCHAR(255) UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL,
  accepted_at TIMESTAMPTZ,
  rejected_at TIMESTAMPTZ
);

-- Migration 4: Add indexes (as documented above)
-- CREATE INDEX statements...
```

---

*Dokumen ini menjelaskan schema database region system. Untuk API lihat REGION_API_SPEC.md, untuk business logic lihat REGION_SYSTEM.md*
