# Database Schema — Role & Permission Module

> **Stack:** Golang + PostgreSQL  
> **Berdasarkan:** `ROLE_PERMISSION_SYSTEM.md`  
> **Dibuat:** 2026-05-25

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [Tabel: permissions](#1-permissions)
3. [Tabel: roles](#2-roles)
4. [Tabel: system_role_permissions](#3-system_role_permissions)
5. [Tabel: user_roles](#4-user_roles)
6. [Tabel: community_role_permissions](#5-community_role_permissions)
7. [Ringkasan Enum Values](#ringkasan-enum-values)
8. [Catatan Implementasi Golang](#catatan-implementasi-golang)
9. [Contoh Query](#contoh-query)

---

## Overview Relasi

```
permissions
  ├── system_role_permissions ──────── roles (system)
  └── community_role_permissions ───── community (dari modul lain)

user_roles ─────────── users (dari modul auth)
               ├────── roles
               └────── regions / communities (polymorphic scope)

roles:
  ├── system roles (usergod, superadmin, admin, member, guest)
  └── community roles (leader, moderator, member)
```

### Entity Relationship Diagram (Teks)

```
┌─────────────────────────────────────────────────────────────┐
│                      permissions                            │
│  (master list: post_news, moderate_posts, manage_members...) │
└──────────┬──────────────────────────────────────────────────┘
           │
           ├─────────────────┬──────────────────────┐
           │                 │                      │
    ┌──────▼──────┐  ┌──────▼──────┐      ┌───────▼────────┐
    │   roles     │  │ system_role │      │ community_role │
    │ (system &   │  │ permissions │      │ permissions    │
    │ community)  │  └─────────────┘      └────────────────┘
    └──────┬──────┘
           │
      ┌────▼────────────────┐
      │   user_roles        │
      │ (assign role to user│
      │  in scope context)  │
      └─────────────────────┘
           │
        users (dari modul auth)
```

---

## 1. `permissions`

Master data semua permission yang tersedia di sistem. Didefine oleh usergod, tidak sering berubah. Cocok di-cache.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `key` | `VARCHAR(100)` | NOT NULL, UNIQUE | Identifier unik, e.g. `post_news`, `moderate_posts` |
| `display_name` | `VARCHAR(200)` | NOT NULL | Human-readable, e.g. "Post News Article" |
| `description` | `TEXT` | NULLABLE | Penjelasan detail |
| `scope` | `VARCHAR(20)` | NOT NULL | `global`, `regional`, `community` |
| `category` | `VARCHAR(50)` | NOT NULL | Grouping: `content`, `moderation`, `member`, `admin`, `system` |
| `risk_level` | `VARCHAR(20)` | NOT NULL, DEFAULT `'low'` | `low`, `medium`, `high` (untuk audit) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | Auto-update via trigger |

**DDL Indexes:**

```sql
-- Unique key constraint
CREATE UNIQUE INDEX idx_permissions_key
  ON permissions (LOWER(key));

-- Query by scope (filter permission untuk community vs global)
CREATE INDEX idx_permissions_scope
  ON permissions (scope);

-- Query by category (group di UI)
CREATE INDEX idx_permissions_category
  ON permissions (category);

-- Risk level (untuk audit logging)
CREATE INDEX idx_permissions_risk_level
  ON permissions (risk_level);
```

**Predefined Permission Keys:**

| Key | Display Name | Scope | Category | Risk Level |
|---|---|---|---|---|
| `post_content` | Post content | community | content | low |
| `moderate_posts` | Moderate posts | community | moderation | medium |
| `manage_members` | Manage community members | community | member | medium |
| `delete_content` | Delete content | community | moderation | high |
| `create_community` | Create community | global | content | low |
| `post_news` | Post news article | global | content | medium |
| `approve_news` | Approve news for publishing | global | admin | high |
| `create_store` | Create merchant store | global | content | low |
| `assign_role` | Assign role to user | global | admin | high |
| `manage_region` | Manage regional settings | regional | admin | high |
| `view_analytics` | View analytics | global | admin | low |
| `manage_news_sources` | Manage RSS news sources | global | system | high |
| `configure_auto_schedule` | Configure auto-schedule news | global | system | high |
| `read_content` | Read basic content | global | content | low |
| `join_community` | Join community | global | content | low |
| `ask_qna` | Ask question in Q&A | global | content | low |

---

## 2. `roles`

Daftar semua role yang ada di sistem. Mencakup system roles dan community roles.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `name` | `VARCHAR(100)` | NOT NULL, UNIQUE | Nama role, e.g. `usergod`, `superadmin`, `admin`, `leader` |
| `display_name` | `VARCHAR(200)` | NOT NULL | Human-readable, e.g. "Admin KAI Pusat" |
| `description` | `TEXT` | NULLABLE | Penjelasan role |
| `role_type` | `VARCHAR(20)` | NOT NULL | `system` atau `community` |
| `assignable` | `BOOLEAN` | NOT NULL, DEFAULT `true` | Apakah role ini bisa di-assign ke user? (usergod tidak assignable) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `updated_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |

**DDL Indexes:**

```sql
-- Unique name constraint
CREATE UNIQUE INDEX idx_roles_name
  ON roles (LOWER(name));

-- Query by type (filter system vs community roles)
CREATE INDEX idx_roles_type
  ON roles (role_type);

-- Query assignable roles only
CREATE INDEX idx_roles_assignable
  ON roles (assignable)
  WHERE assignable = true;
```

**Predefined Roles:**

| Name | Role Type | Display Name | Assignable | Description |
|---|---|---|---|---|
| `usergod` | system | Developer/Vendor | false | Akses penuh semua sistem. Hanya untuk developer. |
| `superadmin` | system | Admin KAI Pusat | true | Admin KAI Pusat, kelola seluruh platform. |
| `admin` | system | Admin Region | true | Admin per wilayah, kelola konten regional. |
| `member` | system | Member | true | User terdaftar, punya subscription plan. |
| `guest` | system | Guest | true | Belum login, read-only basic content. |
| `leader` | community | Community Leader | true | Founder/owner komunitas, manage anggota. |
| `moderator` | community | Community Moderator | true | Bantu leader moderasi konten. |
| `community_member` | community | Community Member | true | Anggota biasa komunitas. |

---

## 3. `system_role_permissions`

Join table antara system roles dan permissions. Menentukan permission mana yang dimiliki setiap system role.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `role_id` | `UUID` | NOT NULL, FK → `roles.id` (WHERE role_type='system'), ON DELETE CASCADE | |
| `permission_id` | `UUID` | NOT NULL, FK → `permissions.id`, ON DELETE CASCADE | |
| `assigned_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `assigned_by` | `UUID` | NOT NULL, FK → `users.id` | Admin/usergod yang assign |
| `notes` | `TEXT` | NULLABLE | Catatan kenapa permission ini di-assign |

**DDL Indexes & Constraints:**

```sql
-- Satu role tidak boleh punya permission yang sama dua kali
CREATE UNIQUE INDEX idx_system_role_perm_unique
  ON system_role_permissions (role_id, permission_id);

-- Query permission untuk role tertentu
CREATE INDEX idx_system_role_perm_role_id
  ON system_role_permissions (role_id);

-- Query role yang punya permission tertentu
CREATE INDEX idx_system_role_perm_permission_id
  ON system_role_permissions (permission_id);
```

---

## 4. `user_roles`

Assign system role atau community role ke user dalam scope tertentu. Satu user bisa punya multiple roles di scope berbeda.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `user_id` | `UUID` | NOT NULL, FK → `users.id`, ON DELETE CASCADE | |
| `role_id` | `UUID` | NOT NULL, FK → `roles.id`, ON DELETE CASCADE | |
| `scope_type` | `VARCHAR(20)` | NOT NULL | `global`, `region`, `community` |
| `scope_id` | `UUID` | NULLABLE | region_id atau community_id (NULL untuk global) |
| `assigned_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `assigned_by` | `UUID` | NOT NULL, FK → `users.id` | Admin/superadmin/leader yang assign |
| `expired_at` | `TIMESTAMPTZ` | NULLABLE | Temporary assignment expiry (opsional) |
| `is_active` | `BOOLEAN` | NOT NULL, DEFAULT `true` | Soft deactivate role |
| `deactivated_at` | `TIMESTAMPTZ` | NULLABLE | Kapan di-deactivate |

**DDL Indexes & Constraints:**

```sql
-- Query user's roles
CREATE INDEX idx_user_roles_user_id
  ON user_roles (user_id);

-- Query roles dalam scope tertentu
CREATE INDEX idx_user_roles_scope
  ON user_roles (scope_type, scope_id);

-- Query active roles only (common query)
CREATE INDEX idx_user_roles_active
  ON user_roles (user_id, is_active)
  WHERE is_active = true;

-- Compound: user + scope (get user's role in specific scope)
CREATE INDEX idx_user_roles_user_scope
  ON user_roles (user_id, scope_type, scope_id);

-- Ensure expired role is not used
CREATE INDEX idx_user_roles_expiry
  ON user_roles (user_id, expired_at)
  WHERE is_active = true AND (expired_at IS NULL OR expired_at > NOW());

-- Prevent user punya dua role yang sama dalam scope yang sama
CREATE UNIQUE INDEX idx_user_roles_no_duplicate
  ON user_roles (user_id, role_id, scope_type, scope_id)
  WHERE is_active = true;
```

---

## 5. `community_role_permissions`

Join table antara community roles dan permissions di level community. Memungkinkan leader customize permission per komunitas.

| Kolom | Tipe PostgreSQL | Constraint | Keterangan |
|---|---|---|---|
| `id` | `UUID` | PK, DEFAULT `gen_random_uuid()` | |
| `community_id` | `UUID` | NOT NULL, FK → `communities.id`, ON DELETE CASCADE | |
| `role_id` | `UUID` | NOT NULL, FK → `roles.id` (WHERE role_type='community'), ON DELETE CASCADE | |
| `permission_id` | `UUID` | NOT NULL, FK → `permissions.id`, ON DELETE CASCADE | |
| `assigned_at` | `TIMESTAMPTZ` | NOT NULL, DEFAULT `NOW()` | |
| `assigned_by` | `UUID` | NOT NULL, FK → `users.id` | Community leader yang assign |
| `override_permission` | `BOOLEAN` | NOT NULL, DEFAULT `false` | Apakah ini override default permission? |
| `notes` | `TEXT` | NULLABLE | Catatan permission override |

**DDL Indexes & Constraints:**

```sql
-- Satu role tidak boleh punya permission yang sama dua kali dalam community yang sama
CREATE UNIQUE INDEX idx_community_role_perm_unique
  ON community_role_permissions (community_id, role_id, permission_id);

-- Query permission untuk role dalam community tertentu
CREATE INDEX idx_community_role_perm_community_role
  ON community_role_permissions (community_id, role_id);

-- Query permission untuk role dalam semua community (untuk pre-load)
CREATE INDEX idx_community_role_perm_role_id
  ON community_role_permissions (role_id);

-- Query permission untuk community tertentu
CREATE INDEX idx_community_role_perm_community_id
  ON community_role_permissions (community_id);
```

---

## Ringkasan Enum Values

### `permissions.scope`

| Value | Keterangan |
|---|---|
| `global` | Permission berlaku di level platform (usergod, superadmin, member) |
| `regional` | Permission berlaku di level region (admin regional, regional operations) |
| `community` | Permission berlaku dalam komunitas (leader, moderator, community member) |

### `permissions.category`

| Value | Keterangan |
|---|---|
| `content` | Posting, membuat content, berbagi |
| `moderation` | Menghapus, approve, mengelola konten |
| `member` | Manage member, invite, role assignment |
| `admin` | Administrative actions (assign role, manage system) |
| `system` | System-level (manage sources, configure schedule) |

### `permissions.risk_level`

| Value | Keterangan |
|---|---|
| `low` | Aksi benign, rarely abused |
| `medium` | Aksi yang perlu perhatian, monitor |
| `high` | Aksi sensitive, log semua, need approval |

### `roles.role_type`

| Value | Keterangan |
|---|---|
| `system` | Role global (usergod, superadmin, admin, member, guest) |
| `community` | Role lokal per komunitas (leader, moderator, member) |

### `user_roles.scope_type`

| Value | Keterangan |
|---|---|
| `global` | Role berlaku di seluruh platform |
| `region` | Role berlaku di region tertentu (scope_id = region_id) |
| `community` | Role berlaku dalam komunitas tertentu (scope_id = community_id) |

---

## Catatan Implementasi Golang

### Tipe Data Mapping

| PostgreSQL | Golang (nullable) | Golang (not null) |
|---|---|---|
| `UUID` | `*string` | `string` |
| `TIMESTAMPTZ` | `*time.Time` | `time.Time` |
| `VARCHAR` / `TEXT` | `*string` | `string` |
| `BOOLEAN` | `*bool` | `bool` |

### Struct: `Permission`

```go
package permission

import "time"

// Permission merepresentasikan satu aksi/capability di sistem.
type Permission struct {
    ID        string    `db:"id"         json:"id"`
    Key       string    `db:"key"        json:"key"`
    DisplayName string  `db:"display_name" json:"display_name"`
    Description *string `db:"description"  json:"description,omitempty"`
    Scope     string    `db:"scope"      json:"scope"`
    Category  string    `db:"category"   json:"category"`
    RiskLevel string    `db:"risk_level" json:"risk_level"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Enum untuk scope
const (
    ScopeGlobal    = "global"
    ScopeRegional  = "regional"
    ScopeCommunity = "community"
)

// Enum untuk category
const (
    CategoryContent    = "content"
    CategoryModeration = "moderation"
    CategoryMember     = "member"
    CategoryAdmin      = "admin"
    CategorySystem     = "system"
)

// Enum untuk risk_level
const (
    RiskLevelLow    = "low"
    RiskLevelMedium = "medium"
    RiskLevelHigh   = "high"
)
```

### Struct: `Role`

```go
package role

import "time"

// Role merepresentasikan satu role yang bisa di-assign ke user.
type Role struct {
    ID          string    `db:"id"            json:"id"`
    Name        string    `db:"name"          json:"name"`
    DisplayName string    `db:"display_name"  json:"display_name"`
    Description *string   `db:"description"   json:"description,omitempty"`
    RoleType    string    `db:"role_type"     json:"role_type"`
    Assignable  bool      `db:"assignable"    json:"assignable"`
    CreatedAt   time.Time `db:"created_at"    json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"    json:"updated_at"`
}

// Enum untuk role_type
const (
    RoleTypeSystem    = "system"
    RoleTypeCommunity = "community"
)

// Predefined role names
const (
    RoleUsergod   = "usergod"
    RoleSuperadmin = "superadmin"
    RoleAdmin     = "admin"
    RoleMember    = "member"
    RoleGuest     = "guest"

    RoleLeader   = "leader"
    RoleModerator = "moderator"
    RoleCommunityMember = "community_member"
)
```

### Struct: `SystemRolePermission`

```go
package role

import "time"

// SystemRolePermission merepresentasikan permission yang di-assign ke system role.
type SystemRolePermission struct {
    ID           string    `db:"id"             json:"id"`
    RoleID       string    `db:"role_id"        json:"role_id"`
    PermissionID string    `db:"permission_id"  json:"permission_id"`
    AssignedAt   time.Time `db:"assigned_at"    json:"assigned_at"`
    AssignedBy   string    `db:"assigned_by"    json:"assigned_by"`
    Notes        *string   `db:"notes"          json:"notes,omitempty"`
}

// SystemRolePermissionWithDetails digunakan untuk response API
// yang menyertakan detail permission & role.
type SystemRolePermissionWithDetails struct {
    SystemRolePermission
    Role       *Role       `db:"-" json:"role,omitempty"`
    Permission *Permission `db:"-" json:"permission,omitempty"`
}
```

### Struct: `UserRole`

```go
package role

import "time"

// UserRole merepresentasikan assignment role ke user dalam scope tertentu.
type UserRole struct {
    ID        string     `db:"id"           json:"id"`
    UserID    string     `db:"user_id"      json:"user_id"`
    RoleID    string     `db:"role_id"      json:"role_id"`
    ScopeType string     `db:"scope_type"   json:"scope_type"`
    ScopeID   *string    `db:"scope_id"     json:"scope_id,omitempty"`
    AssignedAt time.Time `db:"assigned_at"  json:"assigned_at"`
    AssignedBy string    `db:"assigned_by"  json:"assigned_by"`
    ExpiredAt *time.Time `db:"expired_at"   json:"expired_at,omitempty"`
    IsActive  bool       `db:"is_active"    json:"is_active"`
    DeactivatedAt *time.Time `db:"deactivated_at" json:"deactivated_at,omitempty"`
}

// Enum untuk scope_type
const (
    ScopeTypeGlobal    = "global"
    ScopeTypeRegion    = "region"
    ScopeTypeCommunity = "community"
)

// UserRoleWithDetails digunakan untuk response API yang menyertakan
// detail role, scope name, dan permission list.
type UserRoleWithDetails struct {
    UserRole
    Role          *Role             `db:"-" json:"role,omitempty"`
    ScopeName     *string           `db:"-" json:"scope_name,omitempty"` // region name atau community name
    Permissions   []Permission      `db:"-" json:"permissions,omitempty"` // flattened permission list
}

// UserPermissionContext adalah struktur untuk caching user's permission
// di dalam request context (untuk permission check logic).
type UserPermissionContext struct {
    UserID       string
    Roles        []UserRoleWithDetails
    Permissions  map[string]bool // permission_key → true/false
    IsExpired    bool
}
```

### Struct: `CommunityRolePermission`

```go
package role

import "time"

// CommunityRolePermission merepresentasikan permission yang di-assign ke community role.
type CommunityRolePermission struct {
    ID                string    `db:"id"                  json:"id"`
    CommunityID       string    `db:"community_id"        json:"community_id"`
    RoleID            string    `db:"role_id"             json:"role_id"`
    PermissionID      string    `db:"permission_id"       json:"permission_id"`
    AssignedAt        time.Time `db:"assigned_at"         json:"assigned_at"`
    AssignedBy        string    `db:"assigned_by"         json:"assigned_by"`
    OverridePermission bool     `db:"override_permission" json:"override_permission"`
    Notes             *string   `db:"notes"               json:"notes,omitempty"`
}

// CommunityRolePermissionWithDetails digunakan untuk response API.
type CommunityRolePermissionWithDetails struct {
    CommunityRolePermission
    Role       *Role       `db:"-" json:"role,omitempty"`
    Permission *Permission `db:"-" json:"permission,omitempty"`
}

// CommunityRolePermissionSet adalah collection permission untuk satu role dalam community.
type CommunityRolePermissionSet struct {
    CommunityID string
    RoleID      string
    Permissions []Permission
}
```

---

## Contoh Query

### Query 1: Get User's All Permissions (Global + Community Scoped)

```sql
-- Ambil semua permission user (dari semua role mereka)
WITH user_all_roles AS (
    SELECT DISTINCT ur.role_id, ur.scope_type, ur.scope_id
    FROM user_roles ur
    WHERE ur.user_id = $1 AND ur.is_active = true
      AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
),
global_perms AS (
    SELECT DISTINCT p.* FROM permissions p
    INNER JOIN system_role_permissions srp ON p.id = srp.permission_id
    INNER JOIN user_all_roles uar ON srp.role_id = uar.role_id
    WHERE uar.scope_type = 'global'
),
regional_perms AS (
    SELECT DISTINCT p.* FROM permissions p
    INNER JOIN system_role_permissions srp ON p.id = srp.permission_id
    INNER JOIN user_all_roles uar ON srp.role_id = uar.role_id
    WHERE uar.scope_type = 'region'
),
community_perms AS (
    SELECT DISTINCT p.* FROM permissions p
    INNER JOIN community_role_permissions crp ON p.id = crp.permission_id
    INNER JOIN user_all_roles uar ON crp.role_id = uar.role_id AND crp.community_id = uar.scope_id
    WHERE uar.scope_type = 'community'
)
SELECT * FROM (
    SELECT * FROM global_perms
    UNION ALL
    SELECT * FROM regional_perms
    UNION ALL
    SELECT * FROM community_perms
) all_perms
ORDER BY key;
```

### Query 2: Get User's Role in Specific Community

```sql
SELECT ur.* FROM user_roles ur
WHERE ur.user_id = $1
  AND ur.scope_type = 'community'
  AND ur.scope_id = $2
  AND ur.is_active = true
  AND (ur.expired_at IS NULL OR ur.expired_at > NOW())
LIMIT 1;
```

### Query 3: Get Community Role Permissions

```sql
SELECT p.* FROM permissions p
INNER JOIN community_role_permissions crp ON p.id = crp.permission_id
WHERE crp.community_id = $1 AND crp.role_id = $2
  AND p.scope = 'community';
```

### Query 4: Check if User Can Do Action

```go
// Pseudocode untuk permission check
func (repo *PermissionRepo) CanUserDoAction(ctx context.Context, userID, action string) (bool, error) {
    // 1. Get user's all active roles
    userRoles, err := repo.GetUserRoles(ctx, userID)
    if err != nil {
        return false, err
    }
    
    // 2. Flatten permission from all roles
    permMap := make(map[string]bool)
    
    // From system roles
    for _, ur := range userRoles {
        if ur.ScopeType == "global" || ur.ScopeType == "region" {
            perms, err := repo.GetSystemRolePermissions(ctx, ur.RoleID)
            if err != nil {
                return false, err
            }
            for _, p := range perms {
                permMap[p.Key] = true
            }
        }
    }
    
    // 3. Check if action (permission key) exists in map
    return permMap[action], nil
}
```

### Query 5: Assign Permission to System Role

```sql
INSERT INTO system_role_permissions (role_id, permission_id, assigned_by)
VALUES ($1, $2, $3)
ON CONFLICT (role_id, permission_id) DO NOTHING
RETURNING *;
```

---

## Migration Notes

Saat deploy, jalankan migration dalam urutan:

1. **Create permissions table** — master list
2. **Create roles table** — define all roles
3. **Create system_role_permissions** — assign permission to system roles
4. **Create user_roles** — user role assignment
5. **Create community_role_permissions** — community-level permission (requires communities table exists)

Jalankan seed script untuk insert predefined permissions & roles.

---

*Dokumen ini adalah schema & domain struct reference. Service layer, repository interface, dan migration files dibuat terpisah.*
