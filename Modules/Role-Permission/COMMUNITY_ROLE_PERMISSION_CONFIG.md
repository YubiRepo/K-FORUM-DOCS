# Community Role & Permission Configuration — Clarification

> **Questions from Backend Team:**
> 1. Tiap komunitas punya role & permission yang berbeda, atau cuma permission yang berbeda?
> 2. Saat member pro create community, dia auto-jadi leader dengan default permission apa?
> 3. Siapa yang set default permission untuk leader role?
> 4. Bisa di-custom setelahnya?

---

## Question 1: Role vs Permission — Apa Yang Berbeda Per Komunitas?

### ❌ WRONG Understanding
```
Komunitas A: 
├─ Role: leader, moderator, member
Komunitas B:
├─ Role: leader, moderator, member, senior_moderator
```
❌ Salah! Roles adalah **GLOBAL dan FIXED**, tidak bisa berbeda per komunitas.

### ✅ RIGHT Understanding
```
ROLES (Global, Fixed):
├─ leader
├─ moderator
├─ member

Komunitas A:
├─ leader role → permissions: [post_content, moderate_posts, manage_members, delete_content]

Komunitas B:
├─ leader role → permissions: [post_content, moderate_posts, manage_members]
└─ (no delete_content — leader di B tidak bisa delete)
```

### ✅ ANSWER

**Tiap komunitas punya PERMISSION YANG BERBEDA untuk ROLE YANG SAMA.**

- **Roles:** Global, fixed, tidak berubah (leader, moderator, member)
- **Permissions:** Bisa berbeda per komunitas untuk role yang sama

**Analogi:**
```
Roles = Posisi jabatan (president, vice-president, member)
Permissions = Kewenangan di perusahaan tersebut

Perusahaan A:
├─ President: [approve_budget, manage_hr, fire_employee]

Perusahaan B:
├─ President: [approve_budget, manage_hr]
└─ (tidak bisa fire_employee)

Sama role (president), tapi kewenangan berbeda.
```

---

## Question 2: Default Permission Saat Create Community

### Scenario: Member Pro Andi Bikin Community "Futsal Jakarta"

**Flow:**

```
1. Andi (member pro) → POST /community/create
   {
     "name": "Futsal Jakarta",
     "description": "..."
   }

2. Backend:
   - Check: Does Andi have "create_community" benefit? (subscription check) ✅
   - Create community entry
   - AUTO-ASSIGN USER ROLE:
     INSERT INTO user_roles:
     {
       "user_id": "andi",
       "role_id": "role_leader",
       "scope_type": "community",
       "scope_id": "comm_futsal",
       "assigned_by": "system"
     }

3. DEFAULT PERMISSION ASSIGNMENT:
   - Backend query: "Apa default permissions untuk leader role di komunitas baru?"
   - Answer depends: Siapa yang define default?
```

---

## Question 3: Siapa Set Default Permission?

### Option A: ❌ Hardcode di Backend (NOT RECOMMENDED)

```go
// backend/community/create.go

func (s *CommunityService) CreateCommunity(userID, name string) {
  // ... create community ...
  
  // Hardcode: leader always get these permissions
  defaultPermissions := []string{
    "post_content",
    "moderate_posts",
    "manage_members",
    "delete_content",
  }
  
  // Assign ke leader role di komunitas baru
  for _, permKey := range defaultPermissions {
    assignPermissionToRole(communityID, "leader", permKey)
  }
}
```

**Problem:**
- Kalau mau ubah default, harus re-deploy backend
- Tidak fleksibel
- Tidak bisa override per-role dari UI

---

### Option B: ✅ Seed Data / Migration (RECOMMENDED APPROACH)

**Step 1: Define default permissions di migration**

```sql
-- Migration: 202605_create_default_community_permissions.sql

-- Saat komunitas baru dibuat, copy default permissions dari template

-- Template permissions untuk leader role di komunitas baru:
CREATE TABLE community_role_permissions_template (
  role_name VARCHAR(50),
  permission_key VARCHAR(100),
  PRIMARY KEY (role_name, permission_key)
);

INSERT INTO community_role_permissions_template VALUES
  ('leader', 'post_content'),
  ('leader', 'moderate_posts'),
  ('leader', 'manage_members'),
  ('leader', 'delete_content'),
  ('moderator', 'post_content'),
  ('moderator', 'moderate_posts'),
  ('moderator', 'delete_content'),
  ('member', 'post_content');
```

**Step 2: Backend query template saat create community**

```go
func (s *CommunityService) CreateCommunity(userID, name string) {
  // ... create community ...
  
  // Query: Get default permissions dari template
  template := s.repo.GetCommunityRolePermissionsTemplate()
  
  // For each role in template:
  for roleName, permissions := range template {
    roleID := s.repo.GetRoleIDByName(roleName)
    
    // Assign permissions dari template ke komunitas baru
    for _, permKey := range permissions {
      permID := s.repo.GetPermissionIDByKey(permKey)
      s.repo.AssignPermissionToCommunityRole(
        communityID, 
        roleID, 
        permID,
      )
    }
  }
}
```

**Benefit:**
- Default ada di database (mudah lihat)
- Bisa ubah default tanpa re-deploy
- Superadmin bisa manage template di UI
- Setiap komunitas baru follow template yang sama

---

### Option C: ✅ Superadmin Configure (MOST FLEXIBLE)

**Flow:**

```
1. Usergod/Superadmin access backoffice
2. Go to: Settings → Community Role Permissions Template
3. Setup default permissions untuk komunitas baru:
   
   Leader role:
   ├─ ☑ post_content
   ├─ ☑ moderate_posts
   ├─ ☑ manage_members
   ├─ ☑ delete_content
   
   Moderator role:
   ├─ ☑ post_content
   ├─ ☑ moderate_posts
   ├─ ☑ delete_content
   
   Member role:
   ├─ ☑ post_content

4. Save template

5. Saat user buat komunitas baru:
   - Backend query: SELECT default permissions dari template
   - Auto-assign ke role sesuai template
```

**Benefit:**
- Paling fleksibel
- Superadmin punya full control
- Bisa change anytime tanpa backend deploy
- Template bisa di-version (v1, v2, dll)

---

## Question 4: Bisa Di-Custom Setelahnya?

### ✅ YES! Flow:

**Default (Saat Create Community):**
```
Community "Futsal":
├─ leader: [post_content, moderate_posts, manage_members, delete_content]
├─ moderator: [post_content, moderate_posts, delete_content]
├─ member: [post_content]
```

**Leader Customize (Anytime After):**
```
1. Leader access community settings
2. Go to: Roles & Permissions

3. Click "Moderator role" → edit permissions
4. Remove "delete_content"
5. Save

Result:
Community "Futsal" (AFTER CUSTOMIZATION):
├─ leader: [post_content, moderate_posts, manage_members, delete_content]
├─ moderator: [post_content, moderate_posts]  ← delete_content removed
├─ member: [post_content]
```

**API Flow (Leader customize):**

```
PUT /api/v1/mobile/role-permission/community-role-permissions/bulk-assign
{
  "community_id": "comm_futsal",
  "role_id": "role_moderator",
  "permission_ids": ["perm_post", "perm_moderate"],  // no delete
}

Backend:
1. Check: Is requester (leader) owner of comm_futsal? ✅
2. Delete old: DELETE FROM community_role_permissions 
   WHERE community_id = 'comm_futsal' AND role_id = 'role_moderator'
3. Insert new: INSERT INTO community_role_permissions 
   VALUES (comm_futsal, role_moderator, perm_post), 
          (comm_futsal, role_moderator, perm_moderate)
4. Response: Success
```

---

## Summary: The Right Approach

### 📋 Architecture Decision

```
┌─────────────────────────────────────────────────────────────┐
│ USERGOD/SUPERADMIN (Define Template)                        │
│ ├─ Set default permissions untuk komunitas baru             │
│ └─ Manage permission templates di backoffice                │
└────────────┬────────────────────────────────────────────────┘
             │
             ↓
┌─────────────────────────────────────────────────────────────┐
│ BACKEND (Auto-Assign on Create Community)                   │
│ ├─ Query template dari superadmin config                    │
│ ├─ Create community entry                                   │
│ ├─ Auto-assign leader role to creator                       │
│ └─ Copy default permissions dari template                   │
└────────────┬────────────────────────────────────────────────┘
             │
             ↓
┌─────────────────────────────────────────────────────────────┐
│ LEADER (Customize After Creation)                           │
│ ├─ Access community settings                                │
│ ├─ View current permissions untuk setiap role               │
│ ├─ Add/remove permissions as needed                         │
│ └─ Change takes effect immediately                          │
└─────────────────────────────────────────────────────────────┘
```

### 🔄 Complete Flow

#### Step 1: Superadmin Define Template (One-time)

```
Superadmin di Backoffice:
POST /api/v1/backoffice/community-templates/permissions
{
  "template_name": "default_v1",
  "roles": {
    "leader": ["post_content", "moderate_posts", "manage_members", "delete_content"],
    "moderator": ["post_content", "moderate_posts", "delete_content"],
    "member": ["post_content"]
  }
}

Result: Template saved di database
```

#### Step 2: Member Pro Create Community (Automatic)

```
Member Andi:
POST /api/v1/mobile/community/create
{
  "name": "Futsal Jakarta",
  "description": "..."
}

Backend:
1. Verify: Andi has "create_community" benefit ✅
2. Create community entry
3. Auto-assign: Andi → leader role
4. Query: SELECT default permissions template
5. Copy: leader permissions [post, moderate, manage, delete] 
         → comm_futsal, role_leader
6. Copy: moderator permissions [post, moderate, delete]
         → comm_futsal, role_moderator
7. Copy: member permissions [post]
         → comm_futsal, role_member

Result:
- Community created
- Permissions auto-assigned per template
- Andi (leader) can customize anytime
```

#### Step 3: Leader Customize (On-Demand)

```
Leader Andi (in Futsal settings):
PUT /api/v1/mobile/community-role-permissions/bulk-assign
{
  "community_id": "comm_futsal",
  "role_id": "role_moderator",
  "permission_ids": ["post_content", "moderate_posts"]
  // remove delete_content
}

Backend:
1. Check: Is Andi leader of comm_futsal? ✅
2. Has permission "manage_members"? ✅
3. Delete old permissions
4. Insert new permissions
5. Immediate effect

Result: Moderators di Futsal dapat't delete posts lagi
```

---

## Implementation Checklist

- [ ] **Create table: `community_role_permissions_template`**
  ```sql
  CREATE TABLE community_role_permissions_template (
    id UUID PRIMARY KEY,
    role_name VARCHAR(50),
    permission_key VARCHAR(100),
    UNIQUE(role_name, permission_key)
  );
  ```

- [ ] **Seed template data**
  ```sql
  INSERT INTO community_role_permissions_template VALUES
    (uuid(), 'leader', 'post_content'),
    (uuid(), 'leader', 'moderate_posts'),
    (uuid(), 'leader', 'manage_members'),
    (uuid(), 'leader', 'delete_content'),
    ...
  ```

- [ ] **Backoffice API: GET template**
  ```
  GET /api/v1/backoffice/community-templates/permissions
  → List current template
  ```

- [ ] **Backoffice API: UPDATE template**
  ```
  PUT /api/v1/backoffice/community-templates/permissions
  → Update default template for new communities
  ```

- [ ] **Backend: Query template saat create community**
  ```go
  template := s.repo.GetCommunityRolePermissionsTemplate()
  // Copy permissions ke komunitas baru
  ```

- [ ] **Mobile API: Leader customize permissions**
  ```
  PUT /api/v1/mobile/role-permission/community-role-permissions/bulk-assign
  → Leader customize role permissions
  ```

- [ ] **Update existing documentation**
  - [ ] `ROLE_PERMISSION_SYSTEM.md` - add template section
  - [ ] `API_SPEC_ROLE_PERMISSION_BACKOFFICE.md` - add template endpoints
  - [ ] Database schema comment

---

## Key Points untuk Backend Team

### ✅ DO:
- **Roles are global** — leader, moderator, member (tidak berubah per komunitas)
- **Permissions per community** — same role bisa punya permission berbeda
- **Template approach** — superadmin set default, backend copy, leader customize
- **Immediate effect** — permission change langsung berlaku

### ❌ DON'T:
- ❌ Hardcode default permissions di backend
- ❌ Create different roles per community
- ❌ Make roles template-driven (roles are fixed)
- ❌ Delay permission changes (async process)

---

*Dokumen ini menjelaskan complete flow untuk community role permission configuration. Gunakan untuk implementation guidance.*
