# API Spec — Community Role Permissions Template Management (Backoffice)

Dokumentasi API untuk superadmin manage default permissions template komunitas baru di backoffice dashboard.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/role-permission/templates`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin only)
- **Error Format**: Same as other backoffice APIs (standard message or validation error)

---

## Model Data Utama

### 1. Template Permission Object

```json
{
  "id": "uuid",
  "role_id": "uuid",
  "role": {
    "id": "uuid",
    "name": "leader",
    "display_name": "Community Leader",
    "role_type": "community"
  },
  "permission_id": "uuid",
  "permission": {
    "id": "uuid",
    "key": "post_content",
    "display_name": "Post Content",
    "scope": "community",
    "category": "content"
  },
  "created_at": "2026-05-25T00:00:00.000Z",
  "updated_at": "2026-05-25T00:00:00.000Z"
}
```

### 2. Template By Role Object

```json
{
  "role_id": "uuid",
  "role": {
    "id": "uuid",
    "name": "leader",
    "display_name": "Community Leader",
    "role_type": "community"
  },
  "permissions": [
    {
      "id": "uuid",
      "key": "post_content",
      "display_name": "Post Content",
      "scope": "community"
    },
    {
      "id": "uuid",
      "key": "moderate_posts",
      "display_name": "Moderate Posts",
      "scope": "community"
    }
  ]
}
```

---

## Endpoints

### 1. Get All Template Permissions

Ambil semua template permissions (seluruh role dan permission mapping).

- **URL**: `GET /api/v1/web/role-permission/templates`
- **Autentikasi**: Yes (Superadmin only)
- **Query Parameters**:
  - `role_id` (optional): Filter by specific role
  - `limit` (optional, default: 100)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "role_id": "uuid",
        "role": {
          "id": "uuid",
          "name": "leader",
          "display_name": "Community Leader",
          "role_type": "community"
        },
        "permission_id": "uuid",
        "permission": {
          "id": "uuid",
          "key": "post_content",
          "display_name": "Post Content",
          "scope": "community",
          "category": "content"
        },
        "created_at": "2026-05-25T00:00:00.000Z",
        "updated_at": "2026-05-25T00:00:00.000Z"
      },
      {
        "id": "uuid",
        "role_id": "uuid",
        "role": {
          "id": "uuid",
          "name": "leader",
          "display_name": "Community Leader",
          "role_type": "community"
        },
        "permission_id": "uuid",
        "permission": {
          "id": "uuid",
          "key": "moderate_posts",
          "display_name": "Moderate Posts",
          "scope": "community",
          "category": "moderation"
        },
        "created_at": "2026-05-25T00:00:00.000Z",
        "updated_at": "2026-05-25T00:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 100,
      "offset": 0,
      "total": 8
    }
  }
  ```

---

### 2. Get Template By Role

Ambil template permissions untuk satu role (untuk UI form).

- **URL**: `GET /api/v1/web/role-permission/templates/role/{role_id}`
- **Autentikasi**: Yes (Superadmin only)
- **URL Parameters**:
  - `role_id`: ID dari community role (leader, moderator, member)

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "role_id": "uuid",
      "role": {
        "id": "uuid",
        "name": "leader",
        "display_name": "Community Leader",
        "role_type": "community"
      },
      "permissions": [
        {
          "id": "uuid",
          "key": "post_content",
          "display_name": "Post Content",
          "scope": "community",
          "category": "content"
        },
        {
          "id": "uuid",
          "key": "moderate_posts",
          "display_name": "Moderate Posts",
          "scope": "community",
          "category": "moderation"
        },
        {
          "id": "uuid",
          "key": "manage_members",
          "display_name": "Manage Members",
          "scope": "community",
          "category": "member"
        },
        {
          "id": "uuid",
          "key": "delete_content",
          "display_name": "Delete Content",
          "scope": "community",
          "category": "moderation"
        }
      ]
    }
  }
  ```

---

### 3. Assign Permission to Template Role

Add satu permission ke template role.

- **URL**: `POST /api/v1/web/role-permission/templates`
- **Autentikasi**: Yes (Superadmin only)
- **Request Body**:
  ```json
  {
    "role_id": "uuid",
    "permission_id": "uuid"
  }
  ```

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "role_id": "uuid",
      "role": {
        "id": "uuid",
        "name": "moderator",
        "display_name": "Community Moderator"
      },
      "permission_id": "uuid",
      "permission": {
        "id": "uuid",
        "key": "create_event",
        "display_name": "Create Event",
        "scope": "community"
      },
      "created_at": "2026-05-25T10:00:00.000Z",
      "updated_at": "2026-05-25T10:00:00.000Z"
    },
    "message": "Permission added to template role successfully"
  }
  ```

- **Response (Error 409 — Already exists)**:
  ```json
  {
    "message": "Permission already assigned to this role in template"
  }
  ```

---

### 4. Revoke Permission from Template Role

Remove satu permission dari template role.

- **URL**: `DELETE /api/v1/web/role-permission/templates/{template_id}`
- **Autentikasi**: Yes (Superadmin only)
- **URL Parameters**:
  - `template_id`: ID dari template_permission assignment

- **Response (Success 200)**:
  ```json
  {
    "message": "Permission removed from template role successfully",
    "data": {
      "role_id": "uuid",
      "permission_id": "uuid"
    }
  }
  ```

---

### 5. Bulk Update Template Role Permissions

Ganti semua permissions untuk satu role di template (replace all).

- **URL**: `PUT /api/v1/web/role-permission/templates/role/{role_id}`
- **Autentikasi**: Yes (Superadmin only)
- **URL Parameters**:
  - `role_id`: ID dari community role

- **Request Body**:
  ```json
  {
    "permission_ids": ["uuid1", "uuid2", "uuid3"]
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "role_id": "uuid",
      "role": {
        "id": "uuid",
        "name": "moderator",
        "display_name": "Community Moderator"
      },
      "old_permissions": ["uuid_a", "uuid_b", "uuid_c"],
      "new_permissions": ["uuid1", "uuid2", "uuid3"],
      "added": ["uuid3"],
      "removed": ["uuid_c"],
      "updated_at": "2026-05-25T10:05:00.000Z"
    },
    "message": "Template updated successfully. New communities will use updated permissions."
  }
  ```

---

### 6. Get Available Permissions for Template

Ambil daftar permissions yang BISA di-assign ke template (filter by scope).

- **URL**: `GET /api/v1/web/role-permission/templates/available-permissions`
- **Autentikasi**: Yes (Superadmin only)
- **Query Parameters**:
  - `scope` (optional): Filter by scope (default: `community`)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "key": "post_content",
        "display_name": "Post Content",
        "description": "Allow posting content in community",
        "scope": "community",
        "category": "content"
      },
      {
        "id": "uuid",
        "key": "moderate_posts",
        "display_name": "Moderate Posts",
        "description": "Allow moderating posts",
        "scope": "community",
        "category": "moderation"
      },
      {
        "id": "uuid",
        "key": "manage_members",
        "display_name": "Manage Members",
        "description": "Allow managing community members",
        "scope": "community",
        "category": "member"
      },
      {
        "id": "uuid",
        "key": "delete_content",
        "display_name": "Delete Content",
        "description": "Allow deleting content",
        "scope": "community",
        "category": "moderation"
      }
    ]
  }
  ```

---

### 7. Get Community Roles (For UI Dropdown)

Ambil daftar community roles yang bisa di-template.

- **URL**: `GET /api/v1/web/role-permission/templates/community-roles`
- **Autentikasi**: Yes (Superadmin only)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "name": "leader",
        "display_name": "Community Leader",
        "description": "Founder/owner komunitas",
        "role_type": "community",
        "assignable": true
      },
      {
        "id": "uuid",
        "name": "moderator",
        "display_name": "Community Moderator",
        "description": "Bantu leader moderasi",
        "role_type": "community",
        "assignable": true
      },
      {
        "id": "uuid",
        "name": "community_member",
        "display_name": "Community Member",
        "description": "Anggota biasa komunitas",
        "role_type": "community",
        "assignable": true
      }
    ]
  }
  ```

---

### 8. Reset Template to Default

Reset template ke default (factory reset). Useful untuk undo changes.

- **URL**: `POST /api/v1/web/role-permission/templates/reset`
- **Autentikasi**: Yes (Superadmin only)

- **Request Body**:
  ```json
  {
    "confirm": true
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "message": "Template reset to default successfully",
    "data": {
      "reset_at": "2026-05-25T10:10:00.000Z",
      "affected_roles": 3,
      "total_permissions": 8
    }
  }
  ```

- **Response (Error 400 — Missing confirmation)**:
  ```json
  {
    "message": "Please confirm reset by setting confirm=true"
  }
  ```

---

## UI Flow Example

### Page: Community Role Permissions Template

```
┌─────────────────────────────────────────────────────────┐
│ Community Role Permissions Template                     │
│ (Default permissions untuk komunitas baru)              │
├─────────────────────────────────────────────────────────┤
│                                                         │
│ LEADER role                                   [Edit]    │
│ ├─ ☑ Post content                                      │
│ ├─ ☑ Moderate posts                                    │
│ ├─ ☑ Manage members                                    │
│ ├─ ☑ Delete content                                    │
│                                                         │
│ MODERATOR role                                [Edit]    │
│ ├─ ☑ Post content                                      │
│ ├─ ☑ Moderate posts                                    │
│ ├─ ☑ Delete content                                    │
│                                                         │
│ MEMBER role                                   [Edit]    │
│ ├─ ☑ Post content                                      │
│                                                         │
│                          [Reset to Default] [Save All]  │
└─────────────────────────────────────────────────────────┘
```

### Flow: Edit Moderator Permissions

```
1. Click [Edit] next to MODERATOR role

2. Modal: Edit Moderator Template
   ┌─────────────────────────────────────────┐
   │ Community Role: Moderator               │
   ├─────────────────────────────────────────┤
   │ Available permissions:                  │
   │ ☑ Post content                          │
   │ ☑ Moderate posts                        │
   │ ☑ Delete content                        │
   │ ☐ Create event                          │
   │ ☐ Manage members                        │
   │                                         │
   │              [Cancel] [Save]            │
   └─────────────────────────────────────────┘

3. User uncheck "Delete content"

4. Click [Save]
   → PUT /api/v1/web/role-permission/templates/role/{role_id}
      {
        "permission_ids": ["post_content", "moderate_posts"]
      }
   
5. Response: Success
   → Modal close
   → Template updated
   → UI show updated state
   → Message: "Template updated. New communities will use updated permissions."
```

---

## Important Notes

### ✅ DO:
- Template hanya affect komunitas BARU
- Perubahan immediate, no approval needed
- Superadmin bisa reset anytime
- Clear UI untuk show permission impact

### ❌ DON'T:
- ❌ Jangan override existing communities
- ❌ Jangan hardcode default di backend
- ❌ Jangan require migration untuk update template
- ❌ Jangan auto-sync ke existing communities

### When Template Changes Take Effect:

```
Timeline:
┌─────────────────────────────────────────────────────────┐
│ T0: Superadmin update template                          │
│     Leader role: remove "delete_content"                │
│                                                          │
│ T1: Member create community "A" (new)                   │
│     → use UPDATED template                              │
│     → moderator NOT have delete_content                 │
│                                                          │
│ T2: Member create community "B" (new)                   │
│     → use UPDATED template                              │
│     → moderator NOT have delete_content                 │
│                                                          │
│ Community "C" (created BEFORE update):                  │
│     → moderator STILL have delete_content               │
│     → template change NOT affect                        │
└─────────────────────────────────────────────────────────┘
```

---

## Error Handling

Standard error responses:

```json
// 400 Bad Request
{
  "message": "Invalid role_id or permission_id"
}

// 401 Unauthorized
{
  "message": "Authentication required"
}

// 403 Forbidden
{
  "message": "Only superadmin can manage templates"
}

// 404 Not Found
{
  "message": "Role or permission not found"
}

// 409 Conflict
{
  "message": "Permission already assigned to this role in template"
}
```

---

*API spec untuk community role permissions template management. Gunakan untuk backoffice implementation.*
