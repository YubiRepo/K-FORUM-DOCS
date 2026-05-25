# Role & Permission System — KAI App

> **Dokumentasi Sistem**: Penjelasan menyeluruh tentang model role dan permission di KAI App sebelum implementasi API & Database Schema.

---

## Daftar Isi

1. [Overview Konsep](#overview-konsep)
2. [Entitas Utama](#entitas-utama)
3. [Hierarki & Scope](#hierarki--scope)
4. [Permission vs Benefit](#permission-vs-benefit)
5. [Community Role & Permission](#community-role--permission)
6. [Role Assignment Flow](#role-assignment-flow)
7. [Permission Check Logic](#permission-check-logic)
8. [Contoh Use Case](#contoh-use-case)

---

## Overview Konsep

KAI App memiliki **dual permission model**:

1. **System Roles & Permissions** — Global, managed oleh usergod & superadmin
   - Scope: Platform-wide
   - Contoh: usergod, superadmin, admin (region), member, guest

2. **Community Roles & Permissions** — Local per community, managed oleh community leader
   - Scope: Per komunitas
   - Contoh: leader, moderator, member (dalam komunitas X)

Keduanya **share permission master list** yang didefine oleh usergod, tapi diterapkan di konteks berbeda.

---

## Entitas Utama

### 1. Permission (Master)

Daftar **semua aksi/capability** yang ada di sistem. Didefine sekali oleh usergod, tidak berubah sering.

| Field | Deskripsi |
|-------|-----------|
| `id` | UUID, primary key |
| `key` | Unique identifier, e.g. `"post_news"`, `"moderate_posts"` |
| `display_name` | Human-readable, e.g. "Post News Article" |
| `description` | Penjelasan detail aksi ini |
| `scope` | Scope permission: `"global"` (system-wide), `"regional"` (per region), `"community"` (per community) |
| `category` | Grouping: `"content"`, `"moderation"`, `"member"`, `"admin"`, `"system"` |
| `risk_level` | `"low"`, `"medium"`, `"high"` — untuk audit/logging |
| `created_at` | Timestamp |

**Contoh permission:**

| Key | Display Name | Scope | Category | Risk Level |
|-----|--------------|-------|----------|------------|
| `post_content` | Post content | community | content | low |
| `moderate_posts` | Moderate posts | community | moderation | medium |
| `manage_members` | Manage community members | community | member | medium |
| `delete_content` | Delete content | community | moderation | high |
| `post_news` | Post news article | global | content | medium |
| `approve_news` | Approve news for publishing | global | admin | high |
| `assign_role` | Assign role to user | global | admin | high |
| `manage_region` | Manage regional settings | regional | admin | high |

---

### 2. Role (System & Community)

Kumpulan permission yang di-assign sebagai satu unit.

#### System Roles

Roles di level global/regional:

| Role | Scope | Deskripsi |
|------|-------|-----------|
| `usergod` | Platform | Developer/vendor, akses penuh semua sistem. **Tidak bisa di-assign, only untuk developer.** |
| `superadmin` | KAI Pusat | Admin pusat, kelola seluruh platform & konten KAI Pusat region. |
| `admin` | KAI Region | Admin per wilayah, kelola konten & member regional. |
| `member` | User | User terdaftar, punya subscription plan (standard/pro). |
| `guest` | Public | Belum login, read-only basic content. |

#### Community Roles

Roles dalam komunitas:

| Role | Scope | Deskripsi |
|------|-------|-----------|
| `leader` | Per Community | Founder/owner komunitas, manage anggota & settings. |
| `moderator` | Per Community | Bantu leader moderasi konten & anggota. (opsional) |
| `member` | Per Community | Anggota biasa komunitas tersebut. |

---

### 3. Role-Permission Relation

Tabel junction yang menghubungkan role ke permission.

#### System Role-Permission

```
system_role_permissions:
  - role_id → role (usergod, superadmin, admin, member, guest)
  - permission_id → permission
  - assigned_at
  - assigned_by (usergod ID)
```

Contoh assignment:
```
superadmin → [post_news, approve_news, manage_region, assign_role, ...]
admin → [post_news, moderate_posts, manage_members, ...]
member → [post_content, ask_qna, join_community, create_community, ...]
guest → [read_content]
```

#### Community Role-Permission

```
community_role_permissions:
  - community_id
  - role_id → role (leader, moderator, member)
  - permission_id → permission
  - assigned_at
  - assigned_by (community_leader_id)
```

Contoh assignment dalam Community "Futsal":
```
leader → [post_content, moderate_posts, manage_members, delete_content, ...]
moderator → [post_content, moderate_posts, delete_content]
member → [post_content, ask_qna]
```

---

### 4. User-Role Relation

Tabel yang assign role ke user dalam scope tertentu.

```
user_roles:
  - user_id
  - role_id → role
  - scope_type → "global", "region", "community"
  - scope_id → region_id atau community_id (null untuk global)
  - assigned_at
  - assigned_by (admin/superadmin ID)
  - expired_at (opsional, untuk temporary assignment)
```

Contoh user assignment:
```
user_123 → superadmin (scope: global)
user_456 → admin (scope: region_jakarta)
user_789 → leader (scope: community_futsal)
user_789 → moderator (scope: community_nature)
```

---

## Hierarki & Scope

### System Role Hierarchy

```
usergod (developer, akses semua)
  └── superadmin (KAI Pusat)
       └── admin (KAI Region)
            └── member (user terdaftar)
                 └── guest (public)
```

**Rule:**
- usergod dapat assign role ke siapa saja di level manapun
- superadmin dapat assign role ke admin & member (dalam scope KAI Pusat & region-nya)
- admin dapat assign role ke member (dalam scope region-nya saja)
- member & guest tidak bisa assign role

### Community Role Hierarchy

```
leader (community founder)
  └── moderator (optional)
       └── member
```

**Rule:**
- leader dapat assign/revoke role dalam komunitas mereka
- leader dapat appoint moderator
- moderator tidak bisa assign role, hanya moderate content
- member tidak bisa assign role

### Scope Independence

**System role & community role adalah independent**. Contoh:

```
user_456:
  - System role: member (global)
  - Community role: leader (dalam community_futsal)
  - Community role: member (dalam community_nature)
```

User ini:
- Sebagai `member` global → punya permission dari member role (post_content, create_community, dll)
- Sebagai `leader` di futsal → punya permission dari leader role (manage_members, moderate_posts, dll)
- Sebagai `member` di nature → punya permission dari member role (post_content, dll)

---

## Permission vs Benefit

### Permission (Role-based)

**Siapa yang bisa lakukan aksi ini?** → determined oleh role

```
Contoh:
- "approve_news" → hanya superadmin/admin bisa
- "manage_members" → hanya community leader bisa
- "assign_role" → hanya superadmin/usergod bisa
```

### Benefit (Subscription-based)

**Apa fitur yang bisa di-akses?** → determined oleh subscription plan

```
Contoh:
- standard plan → read_content, join_community, ask_qna
- pro plan → post_news, create_community, create_store, post_community
```

### Combined Logic

User harus memiliki **BOTH** untuk do aksi:

```
Contoh 1: User ingin post_news
├─ Benefit check → plan = pro? (subscription benefit)
├─ Permission check → role = superadmin/admin/member_pro? (role permission)
└─ Result: Allowed jika BOTH true

Contoh 2: User ingin approve_news
├─ Benefit check → plan = pro? (mungkin tidak applicable)
├─ Permission check → role = superadmin/admin? (role permission)
└─ Result: Allowed jika permission true (benefit mungkin diabaikan untuk admin actions)
```

---

## Community Role & Permission

### Permission Reuse di Community

Community roles **dapat reuse permission master list**, tapi hanya permission dengan `scope = "community"`.

```
Global permission list:
├─ post_content (scope: community)
├─ moderate_posts (scope: community)
├─ manage_members (scope: community)
├─ delete_content (scope: community)
├─ post_news (scope: global) ❌ tidak boleh di-assign ke community role
├─ approve_news (scope: global) ❌ tidak boleh di-assign ke community role
└─ assign_role (scope: global) ❌ tidak boleh di-assign ke community role
```

Community leader hanya bisa assign permission dengan `scope = "community"` ke role dalam komunitas mereka.

### Community Role Assignment

Alur:
1. **User membuat komunitas** → automatically jadi `leader`
2. **Leader invite user** → user jadi `member` (atau `moderator` jika leader pilih)
3. **Leader assign permission** → ke role (leader, moderator, member) dalam komunitas
4. **Member join** → automatically jadi `member` role (dengan permission default dari member role)

---

## Role Assignment Flow

### System Role Assignment

```
usergod → Define permission master list
    ↓
superadmin → Assign permission ke system roles (superadmin, admin, member, guest)
    ↓
admin (region) → Assign role ke user dalam region (member → admin, member → member, dll)
    ↓
User inherit permission dari role mereka
```

### Community Role Assignment

```
User creates community → automatically leader
    ↓
Leader assign permission ke community roles (leader, moderator, member)
    ↓
Leader invite/assign user ke role (moderator, member)
    ↓
User inherit permission dari community role mereka
```

---

## Permission Check Logic

### At Request Time

User request aksi (e.g., POST /news):

```
1. Get user → user_id
2. Get user roles → all roles (system + community in scope)
3. For each role → get permissions
4. Flatten permission list (union)
5. Check: aksi yang diinginkan ada dalam permission list?
6. Check subscription benefit (jika applicable)
7. Allow/Deny
```

### In Code (Pseudocode)

```go
func CanUserPostNews(userID, newsSourceType) (bool, error) {
    // 1. Get user's system role
    systemRole := getUserSystemRole(userID)
    
    // 2. Get subscription benefit
    plan := getUserSubscriptionPlan(userID)
    hasBenefit := plan.HasBenefit("post_news") // pro plan only
    
    // 3. Get permission from role
    permissions := getRolePermissions(systemRole)
    hasPermission := contains(permissions, "post_news")
    
    // 4. Combine check
    if newsSourceType == "kai_pusat" {
        // Only superadmin/admin can post news
        return hasPermission && (isRole(systemRole, "superadmin") || isRole(systemRole, "admin")), nil
    } else if newsSourceType == "regional" {
        // Admin region can post (with or without approval, depending on config)
        return hasPermission && isRole(systemRole, "admin"), nil
    } else if newsSourceType == "member" {
        // Member pro can post (needs benefit + permission)
        return hasBenefit && hasPermission, nil
    }
    
    return false, nil
}

func CanUserModerateInCommunity(userID, communityID) (bool, error) {
    // 1. Get user's community role in this community
    communityRole := getUserCommunityRole(userID, communityID)
    if communityRole == "" {
        return false, nil // not member of this community
    }
    
    // 2. Get permission from community role
    permissions := getCommunityRolePermissions(communityID, communityRole)
    hasPermission := contains(permissions, "moderate_posts")
    
    return hasPermission, nil
}
```

---

## Contoh Use Case

### Use Case 1: Member Pro ingin post news

**Flow:**
1. User (member pro) navigate ke News → click "Post News"
2. Frontend check: `GET /api/v1/backoffice/permissions/me`
   - Response: `{permissions: ["post_news", "post_content", ...], subscription_plan: "pro"}`
3. Frontend: User punya permission + plan → allow aksi
4. User submit form → `POST /api/v1/backoffice/news`
5. Backend check:
   - User's subscription plan = pro? ✅
   - User's permission includes "post_news"? ✅
   - Approval required (config)? → status = "draft" atau "published"
6. News created

### Use Case 2: Admin region ingin post news

**Flow:**
1. User (admin region_jakarta) navigate ke News → click "Post News"
2. Frontend check: `GET /api/v1/backoffice/permissions/me`
   - Response: `{permissions: ["post_news", "approve_news", "manage_region", ...], role: "admin", scope: "region_jakarta"}`
3. Frontend: User punya permission → allow aksi
4. User submit form → `POST /api/v1/backoffice/news`
5. Backend check:
   - User's role = admin? ✅
   - User's region = region_jakarta? ✅
   - User's permission includes "post_news"? ✅
   - Region admin tidak perlu approval → status = "published" langsung
6. News created & published

### Use Case 3: Community leader ingin assign moderator

**Flow:**
1. User (leader) navigate ke Community Settings → Members
2. Frontend check: `GET /api/v1/community/{communityID}/permissions/me`
   - Response: `{permissions: ["manage_members", "moderate_posts", ...], role: "leader", scope: "community_futsal"}`
3. Frontend: User punya permission manage_members → show assign option
4. Leader click "Make Moderator" on user → `POST /api/v1/community/{communityID}/members/{memberID}/role`
5. Backend check:
   - User's community role = leader? ✅
   - User's permission includes "manage_members"? ✅
   - Target permission (moderator) dalam community scope? ✅
6. User assigned as moderator → inherit moderator permissions

### Use Case 4: Community moderator moderate posts

**Flow:**
1. Moderator navigate ke Community → Posts
2. See delete button on post (because they're moderator, not just member)
3. Click delete → `DELETE /api/v1/community/{communityID}/posts/{postID}`
4. Backend check:
   - User's community role in this community = moderator? ✅
   - Moderator permission includes "delete_content"? ✅
5. Post deleted, log action

---

## Summary Tabel

| Aspek | System Role | Community Role |
|-------|-------------|----------------|
| **Scope** | Global / Regional | Per Community |
| **Managed by** | usergod, superadmin, admin | Community leader |
| **Permission source** | Master list (usergod define) | Master list (same, but filtered) |
| **Typical users** | All authenticated users | Members of specific community |
| **Hierarchy** | usergod > superadmin > admin > member > guest | leader > moderator > member |
| **Use case** | Control platform-wide access | Control community-level access |

---

## Next Step

1. **API Spec Backoffice** — CRUD permission, role-permission assignment (usergod/superadmin)
2. **API Spec Mobile** — Get user permissions (read-only)
3. **Database Schema** — permissions, user_roles, system_role_permissions, community_role_permissions

