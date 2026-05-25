# API Spec — Role & Permission Module (Backoffice)

Dokumentasi API untuk manajemen role dan permission di dashboard backoffice. Endpoint ini diakses oleh **usergod** dan **superadmin** untuk mendefinisikan, mengatur, dan mengaudit permission dan role assignment.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/role-permission`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required — hanya usergod/superadmin)
- **Error Format**: Same as `API_SPEC_AUTH` (standard message error atau validation error)

---

## Model Data Utama

### 1. Permission Object

```json
{
  "id": "uuid",
  "key": "post_news",
  "display_name": "Post News Article",
  "description": "Allow user to post news articles",
  "scope": "global",
  "category": "content",
  "risk_level": "medium",
  "created_at": "2026-05-20T00:00:00.000Z",
  "updated_at": "2026-05-20T00:00:00.000Z"
}
```

### 2. Role Object

```json
{
  "id": "uuid",
  "name": "admin",
  "display_name": "Admin Region",
  "description": "Admin untuk wilayah tertentu",
  "role_type": "system",
  "assignable": true,
  "created_at": "2026-05-20T00:00:00.000Z",
  "updated_at": "2026-05-20T00:00:00.000Z"
}
```

### 3. System Role Permission Object

```json
{
  "id": "uuid",
  "role_id": "uuid",
  "permission_id": "uuid",
  "assigned_at": "2026-05-20T00:00:00.000Z",
  "assigned_by": "uuid",
  "notes": "Optional notes about this assignment"
}
```

### 4. User Role Object

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "role_id": "uuid",
  "scope_type": "global",
  "scope_id": null,
  "assigned_at": "2026-05-20T00:00:00.000Z",
  "assigned_by": "uuid",
  "expired_at": null,
  "is_active": true,
  "deactivated_at": null
}
```

### 5. Detailed Responses

Beberapa endpoint return object dengan detail nested (role details, permission details, dll).

```json
{
  "id": "uuid",
  "role_id": "uuid",
  "role": {
    "id": "uuid",
    "name": "admin",
    "display_name": "Admin Region",
    "role_type": "system",
    "assignable": true
  },
  "permission_id": "uuid",
  "permission": {
    "id": "uuid",
    "key": "post_news",
    "display_name": "Post News Article",
    "scope": "global",
    "category": "content"
  },
  "assigned_at": "2026-05-20T00:00:00.000Z",
  "assigned_by": "uuid"
}
```

---

## Permission Management Endpoints

### 1. List All Permissions

Ambil semua permission yang ada di sistem. Cocok untuk pre-load/cache di frontend.

- **URL**: `GET /api/v1/web/role-permission/permissions`
- **Autentikasi**: Yes (Bearer token)
- **Authorization**: usergod only
- **Query Parameters**:
  - `scope` (optional): Filter by scope (`global`, `regional`, `community`)
  - `category` (optional): Filter by category (`content`, `moderation`, `member`, `admin`, `system`)
  - `risk_level` (optional): Filter by risk level (`low`, `medium`, `high`)
  - `limit` (optional, default: 100): Pagination limit
  - `offset` (optional, default: 0): Pagination offset

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "key": "post_news",
        "display_name": "Post News Article",
        "description": "Allow user to post news articles",
        "scope": "global",
        "category": "content",
        "risk_level": "medium",
        "created_at": "2026-05-20T00:00:00.000Z",
        "updated_at": "2026-05-20T00:00:00.000Z"
      },
      {
        "id": "uuid",
        "key": "moderate_posts",
        "display_name": "Moderate Posts",
        "description": "Moderate posts in community",
        "scope": "community",
        "category": "moderation",
        "risk_level": "medium",
        "created_at": "2026-05-20T00:00:00.000Z",
        "updated_at": "2026-05-20T00:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 100,
      "offset": 0,
      "total": 25
    }
  }
  ```

---

### 2. Get Permission Detail

Ambil detail satu permission.

- **URL**: `GET /api/v1/web/role-permission/permissions/{permission_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod only
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "key": "post_news",
      "display_name": "Post News Article",
      "description": "Allow user to post news articles",
      "scope": "global",
      "category": "content",
      "risk_level": "medium",
      "created_at": "2026-05-20T00:00:00.000Z",
      "updated_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```

---

### 3. Create Permission

Buat permission baru. **Only usergod can do this.**

- **URL**: `POST /api/v1/web/role-permission/permissions`
- **Autentikasi**: Yes
- **Authorization**: usergod only
- **Request Body**:
  ```json
  {
    "key": "custom_permission",
    "display_name": "Custom Permission Name",
    "description": "Description of what this permission allows",
    "scope": "global",
    "category": "content",
    "risk_level": "medium"
  }
  ```
- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "key": "custom_permission",
      "display_name": "Custom Permission Name",
      "description": "Description of what this permission allows",
      "scope": "global",
      "category": "content",
      "risk_level": "medium",
      "created_at": "2026-05-20T00:00:00.000Z",
      "updated_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```
- **Response (Error 409)**:
  ```json
  {
    "message": "Permission with key 'custom_permission' already exists"
  }
  ```

---

### 4. Update Permission

Update permission yang sudah ada. **Only usergod can do this.**

- **URL**: `PUT /api/v1/web/role-permission/permissions/{permission_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod only
- **Request Body**:
  ```json
  {
    "display_name": "Updated Display Name",
    "description": "Updated description",
    "risk_level": "high"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "key": "custom_permission",
      "display_name": "Updated Display Name",
      "description": "Updated description",
      "scope": "global",
      "category": "content",
      "risk_level": "high",
      "created_at": "2026-05-20T00:00:00.000Z",
      "updated_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```

---

### 5. Delete Permission

Hapus permission (soft delete atau restrict based on usage). **Only usergod.**

- **URL**: `DELETE /api/v1/web/role-permission/permissions/{permission_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod only
- **Response (Success 200)**:
  ```json
  {
    "message": "Permission deleted successfully"
  }
  ```
- **Response (Error 409)**:
  ```json
  {
    "message": "Cannot delete permission: still assigned to roles"
  }
  ```

---

## Role Management Endpoints

### 1. List All Roles

Ambil semua role di sistem.

- **URL**: `GET /api/v1/web/role-permission/roles`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Query Parameters**:
  - `role_type` (optional): Filter by type (`system`, `community`)
  - `assignable` (optional): Filter by assignable (`true`, `false`)
  - `limit` (optional, default: 50)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "name": "superadmin",
        "display_name": "Admin KAI Pusat",
        "description": "Admin untuk KAI Pusat",
        "role_type": "system",
        "assignable": true,
        "created_at": "2026-05-20T00:00:00.000Z",
        "updated_at": "2026-05-20T00:00:00.000Z"
      },
      {
        "id": "uuid",
        "name": "leader",
        "display_name": "Community Leader",
        "description": "Leader komunitas",
        "role_type": "community",
        "assignable": true,
        "created_at": "2026-05-20T00:00:00.000Z",
        "updated_at": "2026-05-20T00:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 8
    }
  }
  ```

---

### 2. Get Role Detail

Ambil detail satu role + permission yang di-assign.

- **URL**: `GET /api/v1/web/role-permission/roles/{role_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "name": "admin",
      "display_name": "Admin Region",
      "description": "Admin untuk wilayah tertentu",
      "role_type": "system",
      "assignable": true,
      "created_at": "2026-05-20T00:00:00.000Z",
      "updated_at": "2026-05-20T00:00:00.000Z",
      "permissions": [
        {
          "id": "uuid",
          "key": "post_news",
          "display_name": "Post News Article",
          "scope": "global",
          "category": "content",
          "risk_level": "medium"
        },
        {
          "id": "uuid",
          "key": "moderate_posts",
          "display_name": "Moderate Posts",
          "scope": "community",
          "category": "moderation",
          "risk_level": "medium"
        }
      ]
    }
  }
  ```

---

### 3. Create Role

Buat role baru (system atau community). **Only usergod.**

- **URL**: `POST /api/v1/web/role-permission/roles`
- **Autentikasi**: Yes
- **Authorization**: usergod only
- **Request Body**:
  ```json
  {
    "name": "custom_role",
    "display_name": "Custom Role Name",
    "description": "Description of this role",
    "role_type": "system",
    "assignable": true
  }
  ```
- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "name": "custom_role",
      "display_name": "Custom Role Name",
      "description": "Description of this role",
      "role_type": "system",
      "assignable": true,
      "created_at": "2026-05-20T00:00:00.000Z",
      "updated_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```

---

### 4. Update Role

Update role information. **Only usergod.**

- **URL**: `PUT /api/v1/web/role-permission/roles/{role_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod only
- **Request Body**:
  ```json
  {
    "display_name": "Updated Role Name",
    "description": "Updated description",
    "assignable": false
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "name": "custom_role",
      "display_name": "Updated Role Name",
      "description": "Updated description",
      "role_type": "system",
      "assignable": false,
      "created_at": "2026-05-20T00:00:00.000Z",
      "updated_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```

---

## System Role-Permission Assignment Endpoints

### 1. List System Role Permissions

Ambil semua permission yang di-assign ke satu system role.

- **URL**: `GET /api/v1/web/role-permission/system-role-permissions`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Query Parameters**:
  - `role_id` (optional): Filter by role
  - `permission_id` (optional): Filter by permission
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
          "name": "admin",
          "display_name": "Admin Region"
        },
        "permission_id": "uuid",
        "permission": {
          "id": "uuid",
          "key": "post_news",
          "display_name": "Post News Article",
          "scope": "global",
          "category": "content",
          "risk_level": "medium"
        },
        "assigned_at": "2026-05-20T00:00:00.000Z",
        "assigned_by": "uuid",
        "notes": null
      }
    ],
    "pagination": {
      "limit": 100,
      "offset": 0,
      "total": 45
    }
  }
  ```

---

### 2. Assign Permission to System Role

Assign satu permission ke satu system role.

- **URL**: `POST /api/v1/web/role-permission/system-role-permissions`
- **Autentikasi**: Yes
- **Authorization**: superadmin (untuk superadmin/admin role), usergod (untuk semua role)
- **Request Body**:
  ```json
  {
    "role_id": "uuid",
    "permission_id": "uuid",
    "notes": "Optional notes about why this permission was assigned"
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
        "name": "admin",
        "display_name": "Admin Region"
      },
      "permission_id": "uuid",
      "permission": {
        "id": "uuid",
        "key": "post_news",
        "display_name": "Post News Article",
        "scope": "global",
        "category": "content",
        "risk_level": "medium"
      },
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "assigned_by": "uuid",
      "notes": null
    }
  }
  ```
- **Response (Error 409)**:
  ```json
  {
    "message": "Permission already assigned to this role"
  }
  ```

---

### 3. Revoke Permission from System Role

Hapus assignment permission dari system role.

- **URL**: `DELETE /api/v1/web/role-permission/system-role-permissions/{assignment_id}`
- **Autentikasi**: Yes
- **Authorization**: superadmin or usergod
- **Response (Success 200)**:
  ```json
  {
    "message": "Permission revoked from role successfully"
  }
  ```

---

### 4. Bulk Assign Permissions to Role

Assign multiple permissions sekaligus ke satu role (convenience endpoint).

- **URL**: `POST /api/v1/web/role-permission/system-role-permissions/bulk-assign`
- **Autentikasi**: Yes
- **Authorization**: superadmin or usergod
- **Request Body**:
  ```json
  {
    "role_id": "uuid",
    "permission_ids": ["uuid1", "uuid2", "uuid3"],
    "notes": "Bulk assignment for new role setup"
  }
  ```
- **Response (Success 201)**:
  ```json
  {
    "data": {
      "role_id": "uuid",
      "assigned_count": 3,
      "skipped_count": 0,
      "assignments": [
        {
          "id": "uuid",
          "permission_id": "uuid1",
          "permission_key": "post_news",
          "status": "assigned"
        },
        {
          "id": "uuid",
          "permission_id": "uuid2",
          "permission_key": "moderate_posts",
          "status": "assigned"
        },
        {
          "id": "uuid",
          "permission_id": "uuid3",
          "permission_key": "manage_members",
          "status": "assigned"
        }
      ]
    }
  }
  ```

---

## User Role Assignment Endpoints

### 1. List User Roles

Ambil semua role assignment untuk satu user atau filter user tertentu.

- **URL**: `GET /api/v1/web/role-permission/user-roles`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or admin (filter own region only)
- **Query Parameters**:
  - `user_id` (optional): Filter by user
  - `role_id` (optional): Filter by role
  - `scope_type` (optional): Filter by scope (`global`, `region`, `community`)
  - `scope_id` (optional): Filter by scope ID
  - `is_active` (optional): Filter by active status (`true`, `false`)
  - `limit` (optional, default: 50)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "user": {
          "id": "uuid",
          "name": "Budi Admin",
          "email": "budi@example.com"
        },
        "role_id": "uuid",
        "role": {
          "id": "uuid",
          "name": "admin",
          "display_name": "Admin Region"
        },
        "scope_type": "region",
        "scope_id": "uuid",
        "scope_name": "Jakarta",
        "assigned_at": "2026-05-20T00:00:00.000Z",
        "assigned_by": "uuid",
        "expired_at": null,
        "is_active": true,
        "deactivated_at": null
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 1
    }
  }
  ```

---

### 2. Get User Role Detail

Ambil detail satu user role assignment.

- **URL**: `GET /api/v1/web/role-permission/user-roles/{user_role_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or admin (own region only)
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "user_id": "uuid",
      "user": {
        "id": "uuid",
        "name": "Budi Admin",
        "email": "budi@example.com",
        "phone": "08123456789"
      },
      "role_id": "uuid",
      "role": {
        "id": "uuid",
        "name": "admin",
        "display_name": "Admin Region"
      },
      "scope_type": "region",
      "scope_id": "uuid",
      "scope_name": "Jakarta",
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "assigned_by": "uuid",
      "expired_at": null,
      "is_active": true,
      "deactivated_at": null,
      "permissions": [
        {
          "id": "uuid",
          "key": "post_news",
          "display_name": "Post News Article",
          "scope": "global",
          "category": "content",
          "risk_level": "medium"
        },
        {
          "id": "uuid",
          "key": "moderate_posts",
          "display_name": "Moderate Posts",
          "scope": "community",
          "category": "moderation",
          "risk_level": "medium"
        }
      ]
    }
  }
  ```

---

### 3. Assign Role to User

Assign role ke user dalam scope tertentu.

- **URL**: `POST /api/v1/web/role-permission/user-roles`
- **Autentikasi**: Yes
- **Authorization**: usergod (global role), superadmin (region/community), admin (own region users)
- **Request Body**:
  ```json
  {
    "user_id": "uuid",
    "role_id": "uuid",
    "scope_type": "region",
    "scope_id": "uuid",
    "expired_at": null,
    "notes": "Assigned as admin for Jakarta region"
  }
  ```
- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "user_id": "uuid",
      "role_id": "uuid",
      "scope_type": "region",
      "scope_id": "uuid",
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "assigned_by": "uuid",
      "expired_at": null,
      "is_active": true,
      "deactivated_at": null
    }
  }
  ```
- **Response (Error 409)**:
  ```json
  {
    "message": "User already has this role in this scope"
  }
  ```

---

### 4. Update User Role

Update user role (e.g., extend expiry, change scope).

- **URL**: `PUT /api/v1/web/role-permission/user-roles/{user_role_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or admin (own region only)
- **Request Body**:
  ```json
  {
    "expired_at": "2026-12-31T23:59:59.000Z"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "user_id": "uuid",
      "role_id": "uuid",
      "scope_type": "region",
      "scope_id": "uuid",
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "assigned_by": "uuid",
      "expired_at": "2026-12-31T23:59:59.000Z",
      "is_active": true,
      "deactivated_at": null
    }
  }
  ```

---

### 5. Deactivate User Role

Deactivate role tanpa delete (soft delete dengan timestamp).

- **URL**: `POST /api/v1/web/role-permission/user-roles/{user_role_id}/deactivate`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or admin (own region only)
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "user_id": "uuid",
      "role_id": "uuid",
      "scope_type": "region",
      "scope_id": "uuid",
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "assigned_by": "uuid",
      "expired_at": null,
      "is_active": false,
      "deactivated_at": "2026-05-25T10:30:00.000Z"
    }
  }
  ```

---

### 6. Reactivate User Role

Reactivate deactivated role.

- **URL**: `POST /api/v1/web/role-permission/user-roles/{user_role_id}/reactivate`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or admin (own region only)
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "user_id": "uuid",
      "role_id": "uuid",
      "scope_type": "region",
      "scope_id": "uuid",
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "assigned_by": "uuid",
      "expired_at": null,
      "is_active": true,
      "deactivated_at": null
    }
  }
  ```

---

### 7. Delete User Role

Hapus role assignment (hard delete).

- **URL**: `DELETE /api/v1/web/role-permission/user-roles/{user_role_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Response (Success 200)**:
  ```json
  {
    "message": "User role deleted successfully"
  }
  ```

---

## Community Role-Permission Assignment Endpoints

### 1. List Community Role Permissions

Ambil semua permission yang di-assign ke community roles dalam satu komunitas.

- **URL**: `GET /api/v1/web/role-permission/community-role-permissions`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or community leader
- **Query Parameters**:
  - `community_id` (required): Filter by community
  - `role_id` (optional): Filter by role
  - `permission_id` (optional): Filter by permission
  - `limit` (optional, default: 100)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "community_id": "uuid",
        "role_id": "uuid",
        "role": {
          "id": "uuid",
          "name": "leader",
          "display_name": "Community Leader"
        },
        "permission_id": "uuid",
        "permission": {
          "id": "uuid",
          "key": "post_content",
          "display_name": "Post Content",
          "scope": "community",
          "category": "content",
          "risk_level": "low"
        },
        "assigned_at": "2026-05-20T00:00:00.000Z",
        "assigned_by": "uuid",
        "override_permission": false,
        "notes": null
      }
    ],
    "pagination": {
      "limit": 100,
      "offset": 0,
      "total": 20
    }
  }
  ```

---

### 2. Assign Permission to Community Role

Assign permission ke community role dalam komunitas tertentu.

- **URL**: `POST /api/v1/web/role-permission/community-role-permissions`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or community leader
- **Request Body**:
  ```json
  {
    "community_id": "uuid",
    "role_id": "uuid",
    "permission_id": "uuid",
    "override_permission": false,
    "notes": "Allow moderators to delete posts"
  }
  ```
- **Response (Success 201)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "community_id": "uuid",
      "role_id": "uuid",
      "role": {
        "id": "uuid",
        "name": "moderator",
        "display_name": "Community Moderator"
      },
      "permission_id": "uuid",
      "permission": {
        "id": "uuid",
        "key": "delete_content",
        "display_name": "Delete Content",
        "scope": "community",
        "category": "moderation",
        "risk_level": "high"
      },
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "assigned_by": "uuid",
      "override_permission": false,
      "notes": "Allow moderators to delete posts"
    }
  }
  ```

---

### 3. Revoke Permission from Community Role

Hapus permission dari community role.

- **URL**: `DELETE /api/v1/web/role-permission/community-role-permissions/{assignment_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or community leader (own community)
- **Response (Success 200)**:
  ```json
  {
    "message": "Permission revoked from community role successfully"
  }
  ```

---

### 4. Bulk Assign Permissions to Community Role

Assign multiple permissions sekaligus ke community role.

- **URL**: `POST /api/v1/web/role-permission/community-role-permissions/bulk-assign`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin or community leader
- **Request Body**:
  ```json
  {
    "community_id": "uuid",
    "role_id": "uuid",
    "permission_ids": ["uuid1", "uuid2", "uuid3"],
    "notes": "Setup default permissions for new moderator role"
  }
  ```
- **Response (Success 201)**:
  ```json
  {
    "data": {
      "community_id": "uuid",
      "role_id": "uuid",
      "assigned_count": 3,
      "skipped_count": 0,
      "assignments": [
        {
          "id": "uuid",
          "permission_id": "uuid1",
          "permission_key": "post_content",
          "status": "assigned"
        },
        {
          "id": "uuid",
          "permission_id": "uuid2",
          "permission_key": "moderate_posts",
          "status": "assigned"
        },
        {
          "id": "uuid",
          "permission_id": "uuid3",
          "permission_key": "delete_content",
          "status": "assigned"
        }
      ]
    }
  }
  ```

---

## Audit & Reporting Endpoints

### 1. Get Audit Log (Permission Changes)

Ambil log semua perubahan permission & role assignment.

- **URL**: `GET /api/v1/web/role-permission/audit-logs`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Query Parameters**:
  - `action` (optional): Filter by action (`create_permission`, `assign_permission`, `assign_role`, `deactivate_role`, dll)
  - `risk_level` (optional): Filter by affected permission risk level (`high`, `medium`, `low`)
  - `assigned_by` (optional): Filter by who made the change
  - `start_date` (optional): ISO date, filter changes after this date
  - `end_date` (optional): ISO date, filter changes before this date
  - `limit` (optional, default: 50)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "action": "assign_permission",
        "target_type": "system_role_permission",
        "target_id": "uuid",
        "target_description": "Assigned 'post_news' to admin role",
        "affected_permission": {
          "key": "post_news",
          "risk_level": "medium"
        },
        "changed_by": "uuid",
        "changed_by_name": "Super Admin",
        "timestamp": "2026-05-20T10:30:00.000Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 125
    }
  }
  ```

---

## Error Handling

Semua endpoint mengikuti format error dari `API_SPEC_AUTH`:

### Standard Message Error

```json
{
  "message": "Pesan error deskriptif"
}
```

### Validation Error (422)

```json
{
  "message": "Data input tidak valid",
  "errors": {
    "role_id": ["Role ID is required"],
    "permission_id": ["Permission ID is not valid"]
  }
}
```

### Common Error Codes

| HTTP Code | Message               | Keterangan                                           |
| --------- | --------------------- | ---------------------------------------------------- |
| 400       | Bad Request           | Invalid request format                               |
| 401       | Unauthorized          | Token invalid/expired                                |
| 403       | Forbidden             | User tidak punya permission untuk aksi ini           |
| 404       | Not Found             | Resource tidak ditemukan                             |
| 409       | Conflict              | Duplicate assignment / conflict dengan existing data |
| 422       | Unprocessable Entity  | Validation error                                     |
| 500       | Internal Server Error | Server error                                         |

---

## Notes

1. **Authorization Hierarchy**: usergod > superadmin > admin. Endpoint yang meminta superadmin authorization juga accept usergod.
2. **Audit Trail**: Semua aksi (create, update, delete, assign) harus di-log ke audit log dengan changed_by, timestamp, dan action detail.
3. **Cache Invalidation**: Setiap perubahan permission/role harus invalidate cache user permission (jika di-implement).
4. **Soft Deletes**: User role assignment menggunakan soft delete (is_active flag) bukan hard delete, kecuali ada explicit delete request dari usergod.
5. **Scope Validation**: Saat assign role ke user, validate bahwa scope_id benar-benar ada (region exist, community exist).
