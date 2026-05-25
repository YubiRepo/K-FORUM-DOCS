# Community Role & Permission Scope — Deep Dive Analysis

> **Purpose**: Clarify community role scope, permission hierarchy, dan implementation flows untuk backend team

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Community Role Scope Definition](#community-role-scope-definition)
3. [Database Structure](#database-structure)
4. [Permission Hierarchy in Community](#permission-hierarchy-in-community)
5. [Real-World Case Examples](#real-world-case-examples)
6. [Flows (Tercover & Belum)](#flows-tercover--belum)
7. [Implementation Checklist](#implementation-checklist)

---

## Problem Statement

**Confusion Point:** "Apa itu community role scope? Gimana cara ngatur permission di setiap komunitas?"

**Root Issue:**
- Community role itu LOCAL per komunitas (bukan global)
- Satu user bisa punya role BERBEDA di komunitas berbeda
- Setiap komunitas bisa punya permission setup sendiri (leader-driven)
- Permission reuse dari master list, tapi scope terbatas ke community

---

## Community Role Scope Definition

### Level 1: Platform Level
```
User: Budi
├─ System Role: member (global)
└─ Subscription Plan: pro
```

Budi punya **system role = member** dan **plan = pro** di level global.
Ini bukan community role.

### Level 2: Community Level (Multiple Communities)
```
User: Budi
├─ Community: Futsal Jakarta
│   ├─ Role: leader
│   ├─ Scope: community_futsal_uuid
│   └─ Permissions: [post_content, moderate_posts, manage_members, delete_content]
│
└─ Community: Nature Lovers
    ├─ Role: moderator
    ├─ Scope: community_nature_uuid
    └─ Permissions: [post_content, moderate_posts, delete_content]
```

Budi punya **role yang berbeda** di 2 komunitas:
- Di Futsal: **leader** (full control)
- Di Nature: **moderator** (moderation only)

---

## Database Structure

### Table: `user_roles`

Menyimpan assignment user → role dalam scope tertentu

```
┌────────┬────────┬────────────┬──────────────────┬──────────────┐
│ user_id│ role_id│ scope_type │ scope_id         │ is_active    │
├────────┼────────┼────────────┼──────────────────┼──────────────┤
│ budi   │ leader │ community  │ community_futsal │ true         │
│ budi   │ modera │ community  │ community_nature │ true         │
│ andi   │ leader │ community  │ community_futsal │ true         │
│ andi   │ member │ community  │ community_nature │ true         │
└────────┴────────┴────────────┴──────────────────┴──────────────┘
```

**Key points:**
- `scope_type = 'community'` untuk community role
- `scope_id = community_uuid` untuk specify komunitas mana
- `role_id` reference ke `roles` table (leader, moderator, member, dll)
- Satu user bisa punya multiple rows untuk komunitas berbeda

### Table: `community_role_permissions`

Menyimpan permission assignment role → permission dalam komunitas

```
┌──────────────┬─────────┬──────────────┬─────────────────┐
│ community_id │ role_id │ permission_id│ permission_key  │
├──────────────┼─────────┼──────────────┼─────────────────┤
│ comm_futsal  │ leader  │ perm_post    │ post_content    │
│ comm_futsal  │ leader  │ perm_mod     │ moderate_posts  │
│ comm_futsal  │ leader  │ perm_manage  │ manage_members  │
│ comm_futsal  │ leader  │ perm_delete  │ delete_content  │
│              │         │              │                 │
│ comm_futsal  │ moderator│ perm_post   │ post_content    │
│ comm_futsal  │ moderator│ perm_mod    │ moderate_posts  │
│ comm_futsal  │ moderator│ perm_delete │ delete_content  │
│              │         │              │                 │
│ comm_nature  │ leader  │ perm_post    │ post_content    │
│ comm_nature  │ leader  │ perm_mod     │ moderate_posts  │
│ comm_nature  │ leader  │ perm_manage  │ manage_members  │
└──────────────┴─────────┴──────────────┴─────────────────┘
```

**Key points:**
- Setiap komunitas bisa punya permission setup berbeda
- `community_futsal` → leader punya 4 permissions
- `community_nature` → leader punya 3 permissions (tidak ada delete_content)
- Leader dari `comm_futsal` tidak otomatis punya permission di `comm_nature`

---

## Permission Hierarchy in Community

### Role Hierarchy (Within a Community)

```
INSIDE Community Futsal:

leader (founder/owner)
├─ can manage members (invite, remove, assign role)
├─ can post content (unlimited)
├─ can moderate posts (delete, approve)
├─ can delete content (dari siapa saja)
└─ can manage community settings

moderator (appointed by leader)
├─ can post content (unlimited)
├─ can moderate posts (delete, approve)
├─ can delete content
└─ cannot manage members
└─ cannot manage community settings

member (join request accepted)
├─ can post content (limited quota? atau unlimited?)
├─ can comment on posts
└─ cannot moderate atau delete
```

**PENTING:** Hierarchy ini HANYA berlaku INSIDE satu komunitas.

### Cross-Community Independence

```
User: Budi

Community Futsal:
├─ Role: leader
├─ Permissions: [post, moderate, manage_members, delete]
└─ Status: Owner/full control

Community Nature:
├─ Role: moderator
├─ Permissions: [post, moderate, delete]
└─ Status: Moderator only (NOT owner)
```

Budi di Nature **TIDAK** inherit permissions dari Futsal leader role.
Setiap komunitas adalah **isolated scope**.

---

## Real-World Case Examples

### Case 1: Create Community (User Becomes Leader)

**Scenario:** User Andi (plan: pro) ingin bikin komunitas "Badminton Jakarta"

**Flow:**

1. **Pre-check (permission check):**
   ```
   - Has permission "create_community"? (from subscription plan = pro) ✅
   - Allowed to proceed
   ```

2. **Create community:**
   ```
   POST /api/v1/mobile/community/create
   {
     "name": "Badminton Jakarta",
     "description": "...",
     "category": "sports"
   }
   
   Response:
   {
     "community_id": "comm_badminton",
     "owner_id": "andi",
     "created_at": "2026-05-25T10:00:00.000Z"
   }
   ```

3. **Auto-assign leader role:**
   ```
   INSERT INTO user_roles:
   {
     "user_id": "andi",
     "role_id": "role_leader",
     "scope_type": "community",
     "scope_id": "comm_badminton",
     "is_active": true
   }
   ```

4. **Auto-assign default permissions (leader role in this community):**
   ```
   INSERT INTO community_role_permissions (bulk):
   {
     "community_id": "comm_badminton",
     "role_id": "role_leader",
     "permissions": [
       "post_content",
       "moderate_posts",
       "manage_members",
       "delete_content"
     ]
   }
   ```

5. **Result:**
   - Andi adalah leader di "Badminton Jakarta"
   - Andi punya full permissions di komunitas itu
   - Andi tetap member biasa di komunitas lain (scope-independent)

---

### Case 2: Leader Invite User & Assign Role

**Scenario:** Andi (leader) invite Budi (member) ke "Badminton Jakarta", assign sebagai moderator

**Flow:**

1. **Leader invite (from community settings):**
   ```
   POST /api/v1/mobile/community/invite
   {
     "community_id": "comm_badminton",
     "user_id": "budi",
     "role": "moderator"
   }
   ```

2. **Check authorization:**
   ```
   - Is Andi (requester) leader in comm_badminton? ✅
   - Does Andi have "manage_members" permission? ✅
   - Is role "moderator" valid in this community? ✅
   - Proceed
   ```

3. **Create community membership + assign role:**
   ```
   INSERT INTO user_roles:
   {
     "user_id": "budi",
     "role_id": "role_moderator",
     "scope_type": "community",
     "scope_id": "comm_badminton",
     "is_active": true,
     "assigned_by": "andi"
   }
   ```

4. **Query: What are Budi's permissions in Badminton Jakarta?**
   ```sql
   SELECT p.* FROM permissions p
   INNER JOIN community_role_permissions crp 
     ON p.id = crp.permission_id
   WHERE crp.community_id = 'comm_badminton'
     AND crp.role_id = (
       SELECT role_id FROM user_roles 
       WHERE user_id = 'budi' 
         AND scope_id = 'comm_badminton'
     )
   
   Result: [post_content, moderate_posts, delete_content]
   (no manage_members karena moderator, bukan leader)
   ```

5. **Result:**
   - Budi sekarang member di Badminton Jakarta
   - Role: moderator
   - Permissions: [post, moderate, delete]
   - Budi tetap bisa punya role berbeda di komunitas lain

---

### Case 3: User Post Content (Permission Check)

**Scenario:** Budi (moderator di Badminton) ingin post di komunitas

**Flow:**

1. **User click "Post":**
   ```
   User: Budi
   Community: Badminton Jakarta
   Action: Create post
   ```

2. **Backend permission check:**
   ```
   QUERY: Does Budi have "post_content" permission in Badminton?
   
   SELECT crp.* FROM community_role_permissions crp
   WHERE crp.community_id = 'comm_badminton'
     AND crp.permission_id = (SELECT id FROM permissions WHERE key = 'post_content')
     AND crp.role_id = (
       SELECT role_id FROM user_roles
       WHERE user_id = 'budi'
         AND scope_id = 'comm_badminton'
     )
   
   Result: Found ✅
   ```

3. **Allow post:**
   ```
   INSERT INTO posts:
   {
     "community_id": "comm_badminton",
     "user_id": "budi",
     "content": "...",
     "created_at": "2026-05-25T10:30:00.000Z"
   }
   ```

4. **Result:**
   - Post created
   - Budi dapat post karena moderator punya "post_content" permission

---

### Case 4: Moderator Delete Post (Permission Check)

**Scenario:** Ada post di Badminton yang violate rules, Budi (moderator) ingin delete

**Flow:**

1. **Budi click "Delete" on post:**
   ```
   Moderator: Budi
   Post: post_id_123
   Action: Delete
   ```

2. **Backend permission check:**
   ```
   QUERY: Does Budi have "delete_content" permission in Badminton?
   
   SELECT crp.* FROM community_role_permissions crp
   WHERE crp.community_id = 'comm_badminton'
     AND crp.permission_id = (SELECT id FROM permissions WHERE key = 'delete_content')
     AND crp.role_id = (
       SELECT role_id FROM user_roles
       WHERE user_id = 'budi'
         AND scope_id = 'comm_badminton'
     )
   
   Result: Found ✅
   ```

3. **Allow delete + log action:**
   ```
   UPDATE posts SET deleted_at = NOW() WHERE id = 'post_id_123'
   
   INSERT INTO moderation_log:
   {
     "community_id": "comm_badminton",
     "moderator_id": "budi",
     "action": "delete_post",
     "post_id": "post_id_123",
     "reason": "...",
     "timestamp": "2026-05-25T10:45:00.000Z"
   }
   ```

4. **Result:**
   - Post deleted
   - Log created untuk audit

---

### Case 5: Regular Member Try Delete Post (Permission Denied)

**Scenario:** Citra (regular member di Badminton) coba delete post

**Flow:**

1. **Citra click "Delete" button:**
   ```
   Member: Citra
   Post: post_id_456
   Action: Delete
   ```

2. **Backend permission check:**
   ```
   QUERY: Does Citra have "delete_content" permission in Badminton?
   
   SELECT role_id FROM user_roles
   WHERE user_id = 'citra'
     AND scope_id = 'comm_badminton'
   
   Result: role_id = 'role_member'
   
   QUERY: Does member role have "delete_content" permission?
   
   SELECT crp.* FROM community_role_permissions crp
   WHERE crp.community_id = 'comm_badminton'
     AND crp.role_id = 'role_member'
     AND crp.permission_id = (SELECT id FROM permissions WHERE key = 'delete_content')
   
   Result: NOT FOUND ❌
   ```

3. **Deny action:**
   ```
   Response (403):
   {
     "message": "You don't have permission to delete posts in this community",
     "required_role": "moderator or higher"
   }
   ```

4. **Result:**
   - Action denied
   - Delete button probably tidak visible di UI (atau disabled)

---

### Case 6: User di Multiple Communities (Different Roles)

**Scenario:** Eka adalah leader di "Cooking Club" tapi member biasa di "Book Club"

**State:**

```
user_roles table:

┌──────────┬─────────────┬─────────────────┬─────────────────────┐
│ user_id  │ role_id     │ scope_type      │ scope_id            │
├──────────┼─────────────┼─────────────────┼─────────────────────┤
│ eka      │ role_leader │ community       │ comm_cooking        │
│ eka      │ role_member │ community       │ comm_book           │
└──────────┴─────────────┴─────────────────┴─────────────────────┘
```

**Permission checks:**

```
ACTION 1: Eka post di Cooking Club
- Query role in comm_cooking: leader ✅
- Query permissions: [post, moderate, manage, delete] ✅
- Result: ALLOW, post as leader

ACTION 2: Eka delete post di Cooking Club
- Query role in comm_cooking: leader ✅
- Query permissions: [post, moderate, manage, delete] ✅
- Result: ALLOW, delete

ACTION 3: Eka delete post di Book Club
- Query role in comm_book: member ✅
- Query permissions: [post] ✅ (only post_content)
- Result: DENY, no delete_content permission
```

**Key:** Eka's roles dan permissions **completely different** di dua komunitas.
Tidak ada inheritance atau cross-community effect.

---

### Case 7: Leader Customize Permissions (Advanced)

**Scenario:** Leader di "Coding Community" ingin moderators TIDAK bisa delete posts (hanya bisa moderate)

**Current state (default):**
```
Coding Community:
├─ moderator role permissions: [post_content, moderate_posts, delete_content]
```

**Leader want:**
```
Coding Community:
├─ moderator role permissions: [post_content, moderate_posts]
└─ (remove delete_content dari moderator)
```

**Flow:**

1. **Leader access community settings:**
   ```
   Leader: Eka
   Community: Coding
   ```

2. **Leader revoke permission from role:**
   ```
   DELETE FROM community_role_permissions
   WHERE community_id = 'comm_coding'
     AND role_id = 'role_moderator'
     AND permission_id = (SELECT id FROM permissions WHERE key = 'delete_content')
   ```

3. **Immediate effect:**
   ```
   Moderators di Coding Community sekarang tidak bisa delete posts.
   
   Permission check akan return DENY untuk delete action dari moderator.
   ```

4. **No affect on other communities:**
   ```
   Moderators di komunitas lain tetap bisa delete posts
   (jika permission masih assigned di komunitas mereka)
   ```

---

## Flows (Tercover & Belum)

### ✅ TERCOVER (dari dokumentasi)

| # | Flow | Tercover Di | Status |
|---|------|-------------|--------|
| 1 | Create community → auto-assign leader role | `ROLE_PERMISSION_SYSTEM.md` | ✅ Documented |
| 2 | Leader invite user → assign community role | `API_SPEC_ROLE_PERMISSION_BACKOFFICE.md` (endpoint) | ✅ Partially |
| 3 | User do action (post, moderate, delete) → permission check | `ROLE_PERMISSION_SYSTEM.md` (pseudocode) | ✅ Documented |
| 4 | Get user's community roles | `API_SPEC_ROLE_PERMISSION_MOBILE.md` (#7) | ✅ Documented |
| 5 | Get permissions in specific community | `API_SPEC_ROLE_PERMISSION_MOBILE.md` (#4) | ✅ Documented |
| 6 | Bulk check permissions | `API_SPEC_ROLE_PERMISSION_MOBILE.md` (#6) | ✅ Documented |
| 7 | Leader customize permissions (remove permission from role) | `API_SPEC_ROLE_PERMISSION_BACKOFFICE.md` (#3) | ✅ Documented |

### ⚠️ PERLU CLARIFICATION/UPDATE

| # | Flow | Issue | Action |
|---|------|-------|--------|
| 1 | Leader invite user flow (exact endpoint behavior) | Apakah via `/user-roles` endpoint atau ada community-specific endpoint? | Need API spec clarity |
| 2 | Default permission assignment saat create community | Siapa set default permissions? Leader otomatis dapat semua? | Need explicit flow |
| 3 | Permission check saat user action di community | Query structure belum di-code (just pseudocode) | Need implementation guide |
| 4 | Audit log untuk permission changes | Sudah ada di backoffice API, tapi perlu audit untuk community role changes | Might need endpoint |
| 5 | Community role removal/deactivation | Bagaimana kalau leader remove moderator? | Need flow doc |

### ❌ BELUM TERCOVER

| # | Flow | Need To Add |
|---|------|------------|
| 1 | Community leader leave/resign | Apa happen? Auto-assign ke user lain? |
| 2 | Community deletion → role cleanup | Gimana cleanup user_roles? |
| 3 | Permission change notification | Notify moderator kalau permission berubah? |
| 4 | Override permission per-user | Bisa override permission untuk specific user? (advanced) |

---

## Implementation Checklist

### Backend Implementation Steps

- [ ] **Table creation**
  - [ ] `permissions` table ✅ (sudah di schema)
  - [ ] `roles` table ✅ (sudah di schema)
  - [ ] `user_roles` table ✅ (sudah di schema)
  - [ ] `community_role_permissions` table ✅ (sudah di schema)

- [ ] **Seed data**
  - [ ] Insert predefined permissions (post_content, moderate_posts, etc)
  - [ ] Insert roles (leader, moderator, member, guest)
  - [ ] Insert default community role permissions mapping

- [ ] **API endpoints (Backoffice)**
  - [ ] `GET /role-permission/community-role-permissions` ✅ (documented)
  - [ ] `POST /role-permission/community-role-permissions` ✅ (documented)
  - [ ] `DELETE /role-permission/community-role-permissions/:id` ✅ (documented)
  - [ ] `POST /role-permission/community-role-permissions/bulk-assign` ✅ (documented)

- [ ] **API endpoints (Mobile)**
  - [ ] `GET /role-permission/scope/community/{community_id}` ✅ (documented)
  - [ ] `GET /role-permission/communities` ✅ (documented)
  - [ ] `POST /role-permission/check-bulk` ✅ (documented)

- [ ] **Permission check logic**
  - [ ] Query user's role in community
  - [ ] Query permissions for that role in that community
  - [ ] Return boolean (has permission or not)

- [ ] **Edge cases**
  - [ ] User with no role in community (not member) → all actions denied
  - [ ] User with multiple roles in same community (edge case) → combine permissions
  - [ ] Community role vs system role → keep separate

---

## Key Takeaways untuk Backend Team

### 1. **Community Role adalah LOCAL, Bukan Global**
```
❌ WRONG: Assign leader role globally to user
✅ RIGHT: Assign leader role to user IN SPECIFIC COMMUNITY
```

### 2. **Permission Scope:**
```
✅ Leader permission di community A: [post, moderate, manage, delete]
✅ Moderator permission di community B: [post, moderate]
❌ Leader permission di community B: NOT automatically [post, moderate, manage, delete]
   (harus explicitly assign/inherit di community B)
```

### 3. **Query Pattern untuk permission check:**
```
1. Get user_id + community_id
2. Query user_roles WHERE user_id AND scope_id = community_id
3. Get role_id dari result
4. Query community_role_permissions WHERE community_id AND role_id
5. Check apakah permission_id ada dalam result
6. Return boolean
```

### 4. **ACID: Scope isolation**
```
User X di Community A: leader
User X di Community B: member
User X di Community C: moderator

Setiap community: ISOLATED. Tidak ada inheritance.
Permission check selalu scoped ke community tertentu.
```

### 5. **Flexibility:**
```
Leader bisa customize permission per community.
- Community A: moderator = [post, moderate, delete]
- Community B: moderator = [post, moderate] (no delete)
Sama role name (moderator), tapi permission beda di tiap community.
```

---

## SQL Queries Reference

### Get user's role in community
```sql
SELECT * FROM user_roles
WHERE user_id = $1 
  AND scope_type = 'community'
  AND scope_id = $2
  AND is_active = true;
```

### Get permissions for role in community
```sql
SELECT p.* FROM permissions p
INNER JOIN community_role_permissions crp ON p.id = crp.permission_id
WHERE crp.community_id = $1 
  AND crp.role_id = $2;
```

### Check specific permission
```sql
SELECT COUNT(*) > 0 FROM community_role_permissions
WHERE community_id = $1
  AND role_id = $2
  AND permission_id = (SELECT id FROM permissions WHERE key = $3);
```

### Get all communities where user is leader/moderator
```sql
SELECT DISTINCT c.*, ur.role_id, r.name
FROM communities c
INNER JOIN user_roles ur ON c.id = ur.scope_id
INNER JOIN roles r ON ur.role_id = r.id
WHERE ur.user_id = $1
  AND ur.scope_type = 'community'
  AND r.name IN ('leader', 'moderator')
  AND ur.is_active = true;
```

---

*Document ini mencakup semua contoh case dan flow untuk community role scope. Gunakan sebagai reference untuk implementation di backend.*
