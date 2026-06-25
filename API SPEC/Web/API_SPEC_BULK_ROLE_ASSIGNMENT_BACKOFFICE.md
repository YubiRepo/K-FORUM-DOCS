# API Spec — Bulk Role Assignment to Users (Backoffice)

Dokumentasi API untuk superadmin/admin assign multiple roles ke multiple users dalam satu operasi di backoffice dashboard — termasuk assign multiple roles ke satu user sekaligus.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/role-permission/bulk-assign`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin / Admin Region)
- **Authorization**:
  - Superadmin dapat assign role ke siapa saja, any scope
  - Admin Region hanya dapat assign ke users dalam region mereka
- **Error Format**: Same as other backoffice APIs (standard message or validation error)

---

## Model Data Utama

### 1. Bulk Assignment Request

```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3"],
  "role_id": "uuid",
  "scope_type": "region",
  "scope_id": "uuid",
  "expired_at": "2026-12-31T23:59:59.000Z",
  "notes": "Bulk assign admin to new region"
}
```

### 2. Bulk Assignment Result Item

```json
{
  "user_id": "uuid",
  "user_name": "Andi",
  "email": "andi@example.com",
  "status": "success" | "error",
  "role_assignment_id": "uuid",
  "assigned_at": "2026-05-25T10:00:00.000Z"
}
```

### 3. Bulk Operation Result

```json
{
  "operation_id": "uuid",
  "total_users": 3,
  "successful": 2,
  "failed": 1,
  "results": [
    {
      "user_id": "uuid1",
      "user_name": "Andi",
      "email": "andi@example.com",
      "status": "success",
      "role_assignment_id": "uuid"
    },
    {
      "user_id": "uuid2",
      "user_name": "Budi",
      "email": "budi@example.com",
      "status": "error",
      "reason": "User already has this role in this scope",
      "error_code": "DUPLICATE_ROLE"
    }
  ]
}
```

---

## Endpoints

### 1. Bulk Assign Role to Users

Assign satu role ke multiple users sekaligus (dalam scope yang sama).

- **URL**: `POST /api/v1/web/role-permission/bulk-assign`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**:
  - Superadmin: dapat assign ke user apapun, scope apapun
  - Admin: hanya dapat assign ke users dalam region mereka

- **Request Body**:
  ```json
  {
    "user_ids": ["uuid1", "uuid2", "uuid3", "uuid4"],
    "role_id": "uuid",
    "scope_type": "region",
    "scope_id": "uuid",
    "expired_at": "2026-12-31T23:59:59.000Z",
    "notes": "Assign admin role untuk region Jakarta expansion"
  }
  ```

- **Request Validation**:
  - `user_ids`: Required, array, min 1 item, max 100 items
  - `role_id`: Required, must exist
  - `scope_type`: Required, enum: `global`, `region`, `community`
  - `scope_id`: Required for region/community scope, null for global
  - `expired_at`: Optional, must be future date
  - `notes`: Optional, string max 500 chars

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "operation_id": "bulk_op_20260525_001",
      "total_users": 4,
      "successful": 4,
      "failed": 0,
      "results": [
        {
          "user_id": "uuid1",
          "user_name": "Andi Admin",
          "email": "andi@example.com",
          "status": "success",
          "role_assignment_id": "ur_001",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        },
        {
          "user_id": "uuid2",
          "user_name": "Budi Admin",
          "email": "budi@example.com",
          "status": "success",
          "role_assignment_id": "ur_002",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        },
        {
          "user_id": "uuid3",
          "user_name": "Citra Admin",
          "email": "citra@example.com",
          "status": "success",
          "role_assignment_id": "ur_003",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        },
        {
          "user_id": "uuid4",
          "user_name": "Doni Admin",
          "email": "doni@example.com",
          "status": "success",
          "role_assignment_id": "ur_004",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        }
      ]
    },
    "message": "4 users assigned role successfully"
  }
  ```

- **Response (Partial Success — Some Failed)**:
  ```json
  {
    "data": {
      "operation_id": "bulk_op_20260525_002",
      "total_users": 3,
      "successful": 2,
      "failed": 1,
      "results": [
        {
          "user_id": "uuid1",
          "user_name": "Andi",
          "email": "andi@example.com",
          "status": "success",
          "role_assignment_id": "ur_005"
        },
        {
          "user_id": "uuid2",
          "user_name": "Budi",
          "email": "budi@example.com",
          "status": "error",
          "reason": "User already has this role in this scope",
          "error_code": "DUPLICATE_ROLE"
        },
        {
          "user_id": "uuid3",
          "user_name": "Citra",
          "email": "citra@example.com",
          "status": "success",
          "role_assignment_id": "ur_006"
        }
      ]
    },
    "message": "2 of 3 users assigned successfully (1 failed)"
  }
  ```

- **Response (Validation Error — 422)**:
  ```json
  {
    "message": "Validation failed",
    "errors": {
      "user_ids": ["Must provide at least 1 user"],
      "role_id": ["Role ID is required"],
      "scope_id": ["Scope ID required for region scope"]
    }
  }
  ```

- **Response (Authorization Error — 403)**:
  ```json
  {
    "message": "Admin can only assign roles to users in their own region",
    "data": {
      "user_region": "Surabaya",
      "admin_region": "Jakarta"
    }
  }
  ```

---

### 2. Get Bulk Assignment Status

Check status dari bulk assignment operation yang sedang/sudah selesai.

- **URL**: `GET /api/v1/web/role-permission/bulk-assign/{operation_id}`
- **Autentikasi**: Yes (Superadmin / Admin)
- **URL Parameters**:
  - `operation_id`: ID dari bulk assignment operation

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "operation_id": "bulk_op_20260525_001",
      "status": "completed",
      "total_users": 4,
      "successful": 4,
      "failed": 0,
      "started_at": "2026-05-25T10:00:00.000Z",
      "completed_at": "2026-05-25T10:00:05.000Z",
      "initiated_by": "uuid",
      "initiated_by_name": "Super Admin",
      "results": [...]
    }
  }
  ```

---

### 3. Bulk Assign from CSV/File

Upload CSV file dengan user list, assign role ke semua.

- **URL**: `POST /api/v1/web/role-permission/bulk-assign/from-file`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Content-Type**: `multipart/form-data`

- **Request**:
  ```
  file: users.csv (CSV file)
  role_id: uuid (form field)
  scope_type: region (form field)
  scope_id: uuid (form field)
  notes: optional (form field)
  ```

- **CSV Format**:
  ```
  user_id,email,name
  uuid1,andi@example.com,Andi Admin
  uuid2,budi@example.com,Budi Admin
  uuid3,citra@example.com,Citra Admin
  ```

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "operation_id": "bulk_op_csv_001",
      "total_users": 3,
      "successful": 3,
      "failed": 0,
      "file_name": "users.csv",
      "processed_at": "2026-05-25T10:05:00.000Z",
      "results": [...]
    },
    "message": "CSV processed successfully, 3 users assigned"
  }
  ```

- **Response (CSV Parse Error — 400)**:
  ```json
  {
    "message": "CSV file format invalid",
    "errors": {
      "line_2": "Missing required column: user_id",
      "line_5": "Invalid user_id format"
    }
  }
  ```

---

### 4. Bulk Revoke Role from Users

Remove role dari multiple users sekaligus.

- **URL**: `DELETE /api/v1/web/role-permission/bulk-assign`
- **Autentikasi**: Yes (Superadmin / Admin)

- **Request Body**:
  ```json
  {
    "user_ids": ["uuid1", "uuid2", "uuid3"],
    "role_id": "uuid",
    "scope_type": "region",
    "scope_id": "uuid",
    "reason": "Role no longer needed"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "operation_id": "bulk_revoke_001",
      "total_users": 3,
      "successful": 3,
      "failed": 0,
      "results": [
        {
          "user_id": "uuid1",
          "user_name": "Andi",
          "status": "success",
          "revoked_at": "2026-05-25T10:10:00.000Z"
        },
        {
          "user_id": "uuid2",
          "user_name": "Budi",
          "status": "success",
          "revoked_at": "2026-05-25T10:10:00.000Z"
        },
        {
          "user_id": "uuid3",
          "user_name": "Citra",
          "status": "success",
          "revoked_at": "2026-05-25T10:10:00.000Z"
        }
      ]
    },
    "message": "Role revoked from 3 users"
  }
  ```

---

### 5. Bulk Update Expiry

Update expiry date untuk multiple role assignments sekaligus.

- **URL**: `PATCH /api/v1/web/role-permission/bulk-assign/expiry`
- **Autentikasi**: Yes (Superadmin / Admin)

- **Request Body**:
  ```json
  {
    "user_ids": ["uuid1", "uuid2", "uuid3"],
    "role_id": "uuid",
    "scope_type": "region",
    "scope_id": "uuid",
    "new_expiry": "2026-12-31T23:59:59.000Z"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "operation_id": "bulk_expiry_update_001",
      "total_users": 3,
      "successful": 3,
      "failed": 0,
      "results": [
        {
          "user_id": "uuid1",
          "user_name": "Andi",
          "old_expiry": "2026-06-25T23:59:59.000Z",
          "new_expiry": "2026-12-31T23:59:59.000Z",
          "status": "success"
        },
        {
          "user_id": "uuid2",
          "user_name": "Budi",
          "old_expiry": "2026-06-25T23:59:59.000Z",
          "new_expiry": "2026-12-31T23:59:59.000Z",
          "status": "success"
        },
        {
          "user_id": "uuid3",
          "user_name": "Citra",
          "old_expiry": "2026-06-25T23:59:59.000Z",
          "new_expiry": "2026-12-31T23:59:59.000Z",
          "status": "success"
        }
      ]
    },
    "message": "Expiry updated for 3 role assignments"
  }
  ```

---

### 6. Assign Multiple Roles to Single User

Assign beberapa roles sekaligus ke satu user (bisa beda scope per role).

- **URL**: `POST /api/v1/web/role-permission/bulk-assign/user/{user_id}`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**:
  - Superadmin: dapat assign role apapun, scope apapun
  - Admin: hanya dapat assign ke user dalam region mereka

- **URL Parameters**:
  - `user_id`: ID dari target user

- **Request Body**:
  ```json
  {
    "roles": [
      {
        "role_id": "uuid_admin",
        "scope_type": "region",
        "scope_id": "uuid_jakarta",
        "expired_at": "2026-12-31T23:59:59.000Z"
      },
      {
        "role_id": "uuid_moderator",
        "scope_type": "community",
        "scope_id": "uuid_community_futsal",
        "expired_at": null
      },
      {
        "role_id": "uuid_leader",
        "scope_type": "community",
        "scope_id": "uuid_community_basket",
        "expired_at": null
      }
    ],
    "notes": "Assign multiple roles untuk Andi sebagai admin region sekaligus leader komunitas"
  }
  ```

- **Request Validation**:
  - `roles`: Required, array, min 1 item, max 20 items
  - `roles[].role_id`: Required, must exist
  - `roles[].scope_type`: Required, enum: `global`, `region`, `community`
  - `roles[].scope_id`: Required for region/community scope, null for global
  - `roles[].expired_at`: Optional, must be future date if provided
  - `notes`: Optional, string max 500 chars

- **Response (Success 201)**:
  ```json
  {
    "data": {
      "operation_id": "multi_role_op_20260525_001",
      "user_id": "uuid",
      "user_name": "Andi Pratama",
      "email": "andi@example.com",
      "total_roles": 3,
      "successful": 3,
      "failed": 0,
      "results": [
        {
          "role_id": "uuid_admin",
          "role_name": "Admin Region",
          "scope_type": "region",
          "scope_id": "uuid_jakarta",
          "scope_name": "Jakarta",
          "status": "success",
          "role_assignment_id": "ur_001",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        },
        {
          "role_id": "uuid_moderator",
          "role_name": "Community Moderator",
          "scope_type": "community",
          "scope_id": "uuid_community_futsal",
          "scope_name": "Futsal Jakarta",
          "status": "success",
          "role_assignment_id": "ur_002",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        },
        {
          "role_id": "uuid_leader",
          "role_name": "Community Leader",
          "scope_type": "community",
          "scope_id": "uuid_community_basket",
          "scope_name": "Basket Selatan",
          "status": "success",
          "role_assignment_id": "ur_003",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        }
      ]
    },
    "message": "3 roles assigned to Andi Pratama successfully"
  }
  ```

- **Response (Partial Success — Some Failed)**:
  ```json
  {
    "data": {
      "operation_id": "multi_role_op_20260525_002",
      "user_id": "uuid",
      "user_name": "Andi Pratama",
      "email": "andi@example.com",
      "total_roles": 3,
      "successful": 2,
      "failed": 1,
      "results": [
        {
          "role_id": "uuid_admin",
          "role_name": "Admin Region",
          "scope_type": "region",
          "scope_id": "uuid_jakarta",
          "scope_name": "Jakarta",
          "status": "success",
          "role_assignment_id": "ur_004",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        },
        {
          "role_id": "uuid_moderator",
          "role_name": "Community Moderator",
          "scope_type": "community",
          "scope_id": "uuid_community_futsal",
          "scope_name": "Futsal Jakarta",
          "status": "error",
          "reason": "User already has this role in this scope",
          "error_code": "DUPLICATE_ROLE"
        },
        {
          "role_id": "uuid_leader",
          "role_name": "Community Leader",
          "scope_type": "community",
          "scope_id": "uuid_community_basket",
          "scope_name": "Basket Selatan",
          "status": "success",
          "role_assignment_id": "ur_005",
          "assigned_at": "2026-05-25T10:00:00.000Z"
        }
      ]
    },
    "message": "2 of 3 roles assigned to Andi Pratama successfully (1 failed)"
  }
  ```

- **Response (Validation Error — 422)**:
  ```json
  {
    "message": "Validation failed",
    "errors": {
      "roles": ["Must provide at least 1 role"],
      "roles.1.scope_id": ["Scope ID required for community scope"]
    }
  }
  ```

- **Response (Authorization Error — 403)**:
  ```json
  {
    "message": "Admin can only assign roles to users in their own region",
    "data": {
      "user_region": "Surabaya",
      "admin_region": "Jakarta"
    }
  }
  ```

---

### 7. Revoke Multiple Roles from Single User

Remove beberapa roles sekaligus dari satu user.

- **URL**: `DELETE /api/v1/web/role-permission/bulk-assign/user/{user_id}`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**:
  - Superadmin: dapat revoke role apapun
  - Admin: hanya dapat revoke dari user dalam region mereka

- **URL Parameters**:
  - `user_id`: ID dari target user

- **Request Body**:
  ```json
  {
    "role_assignment_ids": ["ur_001", "ur_002", "ur_003"],
    "reason": "User pindah region, semua role lama dilepas"
  }
  ```

- **Request Validation**:
  - `role_assignment_ids`: Required, array, min 1 item
  - `reason`: Optional, string max 500 chars

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "operation_id": "multi_revoke_op_20260525_001",
      "user_id": "uuid",
      "user_name": "Andi Pratama",
      "total_roles": 3,
      "successful": 3,
      "failed": 0,
      "results": [
        {
          "role_assignment_id": "ur_001",
          "role_name": "Admin Region",
          "scope_name": "Jakarta",
          "status": "success",
          "revoked_at": "2026-05-25T11:00:00.000Z"
        },
        {
          "role_assignment_id": "ur_002",
          "role_name": "Community Moderator",
          "scope_name": "Futsal Jakarta",
          "status": "success",
          "revoked_at": "2026-05-25T11:00:00.000Z"
        },
        {
          "role_assignment_id": "ur_003",
          "role_name": "Community Leader",
          "scope_name": "Basket Selatan",
          "status": "success",
          "revoked_at": "2026-05-25T11:00:00.000Z"
        }
      ]
    },
    "message": "3 roles revoked from Andi Pratama successfully"
  }
  ```

---

## UI Flow Example

### Page: Bulk Role Assignment

```
┌─────────────────────────────────────────────────────────┐
│ Bulk Role Assignment                                    │
├─────────────────────────────────────────────────────────┤
│                                                         │
│ SELECT USERS:                                           │
│ ┌──────────────────────────────────────┐               │
│ │ [Search or select users...]          │               │
│ │ ☑ Andi (andi@example.com)            │               │
│ │ ☑ Budi (budi@example.com)            │               │
│ │ ☑ Citra (citra@example.com)          │               │
│ │ ☑ Doni (doni@example.com)            │               │
│ │ Selected: 4 users                    │               │
│ └──────────────────────────────────────┘               │
│                                                         │
│ ASSIGN ROLE:                                            │
│ Role: [Dropdown: Admin Region ▼]                      │
│                                                         │
│ SCOPE:                                                  │
│ Scope Type: [Dropdown: Region ▼]                      │
│ Region: [Dropdown: Jakarta ▼]                         │
│                                                         │
│ EXPIRY (Optional):                                      │
│ [ ] Set expiry date                                    │
│ Expired at: [2026-12-31] ───────────                  │
│                                                         │
│ NOTES:                                                  │
│ [Text area: Assign admin to Jakarta region...]         │
│                                                         │
│                [Cancel] [Assign to 4 Users]            │
└─────────────────────────────────────────────────────────┘
```

### Flow: Assign Role

```
1. Select users: Andi, Budi, Citra, Doni (4 users)
2. Select role: Admin Region
3. Select scope: Region → Jakarta
4. Click [Assign to 4 Users]

5. Backend:
   POST /api/v1/web/role-permission/bulk-assign
   {
     "user_ids": ["uuid1", "uuid2", "uuid3", "uuid4"],
     "role_id": "role_admin",
     "scope_type": "region",
     "scope_id": "region_jakarta"
   }

6. Response: Success (4/4 assigned)
   ┌─────────────────────────────────────┐
   │ Operation Complete                  │
   │ ✓ 4 users assigned successfully     │
   ├─────────────────────────────────────┤
   │ • Andi - Success                    │
   │ • Budi - Success                    │
   │ • Citra - Success                   │
   │ • Doni - Success                    │
   │                                     │
   │ Operation ID: bulk_op_001           │
   │ Completed at: 10:05:30              │
   │              [Close] [View Details] │
   └─────────────────────────────────────┘
```

### Flow: Assign Multiple Roles to Single User

```
1. Admin buka User Detail: Andi Pratama
2. Click [Assign Roles]

3. Modal: Assign Multiple Roles
   ┌─────────────────────────────────────────────────────┐
   │ Assign Roles to: Andi Pratama                       │
   ├─────────────────────────────────────────────────────┤
   │ Role #1                                   [Remove]  │
   │ Role:       [Admin Region ▼]                        │
   │ Scope Type: [Region ▼]                              │
   │ Region:     [Jakarta ▼]                             │
   │ Expiry:     [2026-12-31]                            │
   │                                                     │
   │ Role #2                                   [Remove]  │
   │ Role:       [Community Moderator ▼]                 │
   │ Scope Type: [Community ▼]                           │
   │ Community:  [Futsal Jakarta ▼]                      │
   │ Expiry:     [ No expiry ]                           │
   │                                                     │
   │ Role #3                                   [Remove]  │
   │ Role:       [Community Leader ▼]                    │
   │ Scope Type: [Community ▼]                           │
   │ Community:  [Basket Selatan ▼]                      │
   │ Expiry:     [ No expiry ]                           │
   │                                                     │
   │ [+ Add Another Role]                                │
   │                                                     │
   │ Notes: [Assign role untuk Andi...]                  │
   │                                                     │
   │              [Cancel] [Assign 3 Roles]              │
   └─────────────────────────────────────────────────────┘

4. Click [Assign 3 Roles]
   → POST /api/v1/web/role-permission/bulk-assign/user/{user_id}
      {
        "roles": [
          { "role_id": "uuid_admin", "scope_type": "region", "scope_id": "uuid_jakarta", "expired_at": "2026-12-31T23:59:59.000Z" },
          { "role_id": "uuid_moderator", "scope_type": "community", "scope_id": "uuid_futsal", "expired_at": null },
          { "role_id": "uuid_leader", "scope_type": "community", "scope_id": "uuid_basket", "expired_at": null }
        ],
        "notes": "Assign role untuk Andi..."
      }

5. Response: Success (3/3 assigned)
   ┌─────────────────────────────────────┐
   │ Roles Assigned                      │
   │ ✓ 3 roles assigned to Andi Pratama  │
   ├─────────────────────────────────────┤
   │ • Admin Region (Jakarta) - Success  │
   │ • Moderator (Futsal Jakarta) - ✓    │
   │ • Leader (Basket Selatan) - ✓       │
   │                                     │
   │              [Close] [View Detail]  │
   └─────────────────────────────────────┘
```

---

## Important Notes

### ✅ DO:
- ✅ Limit bulk operations ke max 100 users per request
- ✅ Return partial success (some succeed, some fail)
- ✅ Log semua bulk operations (audit trail)
- ✅ Show detailed results (per-user status)
- ✅ Support CSV import untuk large operations

### ❌ DON'T:
- ❌ Jangan silent fail (return success kalau sebenarnya failed)
- ❌ Jangan allow assign ke users di scope lain (authorization check)
- ❌ Jangan override existing roles tanpa warning
- ❌ Jangan delete old assignments without tracking

### Transaction Strategy:

```
For each user:
1. Check authorization
2. Check if role already assigned
3. If not, insert new assignment

If any user fails:
→ Return partial success (not rollback all)
→ Allow client to retry failed users

This is more forgiving than all-or-nothing approach
```

---

## Error Handling

Standard error responses:

```json
// 400 Bad Request
{
  "message": "CSV file format invalid"
}

// 401 Unauthorized
{
  "message": "Authentication required"
}

// 403 Forbidden
{
  "message": "Admin can only assign roles to users in their own region"
}

// 404 Not Found
{
  "message": "Role not found"
}

// 422 Unprocessable Entity
{
  "message": "Validation failed",
  "errors": {
    "user_ids": ["Must provide at least 1 user"],
    "scope_id": ["Scope ID required for region scope"]
  }
}
```

| Scenario | HTTP | Reason |
|----------|------|--------|
| No users selected | 422 | Empty user_ids array |
| No roles provided | 422 | Empty roles array |
| Role not found | 404 | Invalid role_id |
| Admin unauthorized | 403 | Trying to assign to different region |
| User already has role | 201 (partial) | Skip, continue with others |
| Scope invalid | 422 | Missing scope_id for region/community |
| CSV parse error | 400 | Invalid CSV format |
| Role assignment not found | 404 | Invalid role_assignment_id saat revoke |

---

*API spec untuk bulk role assignment. Gunakan untuk backoffice implementation.*
