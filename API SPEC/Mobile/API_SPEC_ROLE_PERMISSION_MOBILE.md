# API Spec — Role & Permission Module (Mobile Client)

Dokumentasi API untuk Mobile Client untuk membaca permission user sendiri. Endpoint ini **read-only** dan diakses oleh authenticated users untuk cek apa aksi yang bisa mereka lakukan.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/role-permission`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required untuk semua endpoint)
  - `Accept-Language: <lang_code>` (optional, e.g., `ko`, `id`, `en`)

---

## Model Data Utama

### 1. Permission Object (Mobile)

```json
{
  "id": "uuid",
  "key": "post_news",
  "display_name": "Post News Article",
  "scope": "global",
  "category": "content"
}
```

### 2. Role Object (Mobile)

```json
{
  "id": "uuid",
  "name": "admin",
  "display_name": "Admin Region",
  "role_type": "system"
}
```

### 3. User Role Assignment (Mobile)

```json
{
  "id": "uuid",
  "role_id": "uuid",
  "role": {
    "id": "uuid",
    "name": "admin",
    "display_name": "Admin Region",
    "role_type": "system"
  },
  "scope_type": "global",
  "scope_id": null,
  "scope_name": null
}
```

### 4. User Permissions Response (Mobile)

```json
{
  "user_id": "uuid",
  "subscription_plan": "pro",
  "roles": [
    {
      "id": "uuid",
      "role_id": "uuid",
      "role": {
        "name": "admin",
        "display_name": "Admin Region"
      },
      "scope_type": "region",
      "scope_id": "uuid",
      "scope_name": "Jakarta"
    }
  ],
  "permissions": [
    {
      "key": "post_news",
      "display_name": "Post News Article",
      "scope": "global",
      "category": "content"
    },
    {
      "key": "moderate_posts",
      "display_name": "Moderate Posts",
      "scope": "community",
      "category": "moderation"
    }
  ]
}
```

---

## Endpoints

### 1. Get My Permissions

Ambil semua permission yang dimiliki user saat ini (dari semua role mereka).

- **URL**: `GET /api/v1/mobile/role-permission/me`
- **Autentikasi**: Yes (Bearer token)
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "user_id": "usr_12345",
      "subscription_plan": "pro",
      "roles": [
        {
          "id": "user_role_uuid_1",
          "role_id": "role_admin_uuid",
          "role": {
            "id": "role_admin_uuid",
            "name": "admin",
            "display_name": "Admin Region",
            "role_type": "system"
          },
          "scope_type": "region",
          "scope_id": "region_jakarta_uuid",
          "scope_name": "Jakarta"
        },
        {
          "id": "user_role_uuid_2",
          "role_id": "role_leader_uuid",
          "role": {
            "id": "role_leader_uuid",
            "name": "leader",
            "display_name": "Community Leader",
            "role_type": "community"
          },
          "scope_type": "community",
          "scope_id": "community_futsal_uuid",
          "scope_name": "Futsal Players"
        }
      ],
      "permissions": [
        {
          "key": "post_news",
          "display_name": "Post News Article",
          "scope": "global",
          "category": "content"
        },
        {
          "key": "moderate_posts",
          "display_name": "Moderate Posts",
          "scope": "community",
          "category": "moderation"
        },
        {
          "key": "manage_members",
          "display_name": "Manage Community Members",
          "scope": "community",
          "category": "member"
        },
        {
          "key": "post_content",
          "display_name": "Post Content",
          "scope": "community",
          "category": "content"
        }
      ]
    }
  }
  ```

---

### 2. Check Single Permission

Quick endpoint untuk check apakah user punya permission tertentu.

- **URL**: `GET /api/v1/mobile/role-permission/check/{permission_key}`
- **Autentikasi**: Yes
- **URL Parameters**:
  - `permission_key`: Permission yang dicek (e.g., `post_news`, `moderate_posts`)

- **Response (Success 200 — Has Permission)**:
  ```json
  {
    "data": {
      "user_id": "usr_12345",
      "permission_key": "post_news",
      "has_permission": true,
      "permission": {
        "key": "post_news",
        "display_name": "Post News Article",
        "scope": "global",
        "category": "content"
      }
    }
  }
  ```

- **Response (Success 200 — No Permission)**:
  ```json
  {
    "data": {
      "user_id": "usr_12345",
      "permission_key": "post_news",
      "has_permission": false,
      "permission": {
        "key": "post_news",
        "display_name": "Post News Article",
        "scope": "global",
        "category": "content"
      }
    }
  }
  ```

---

### 3. Get My Roles

Ambil semua role yang di-assign ke user.

- **URL**: `GET /api/v1/mobile/role-permission/roles`
- **Autentikasi**: Yes
- **Query Parameters**:
  - `scope_type` (optional): Filter by scope type (`global`, `region`, `community`)
  - `scope_id` (optional): Filter by specific scope ID

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "user_role_uuid_1",
        "role_id": "role_admin_uuid",
        "role": {
          "id": "role_admin_uuid",
          "name": "admin",
          "display_name": "Admin Region",
          "role_type": "system"
        },
        "scope_type": "region",
        "scope_id": "region_jakarta_uuid",
        "scope_name": "Jakarta"
      },
      {
        "id": "user_role_uuid_2",
        "role_id": "role_leader_uuid",
        "role": {
          "id": "role_leader_uuid",
          "name": "leader",
          "display_name": "Community Leader",
          "role_type": "community"
        },
        "scope_type": "community",
        "scope_id": "community_futsal_uuid",
        "scope_name": "Futsal Players"
      }
    ]
  }
  ```

---

### 4. Get Permissions for Specific Role/Scope

Ambil semua permission untuk role tertentu dalam scope tertentu.

- **URL**: `GET /api/v1/mobile/role-permission/scope/{scope_type}/{scope_id}`
- **Autentikasi**: Yes
- **URL Parameters**:
  - `scope_type`: Type of scope (`global`, `region`, `community`)
  - `scope_id`: ID of the scope (null untuk global, region_id atau community_id untuk lainnya)

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "user_id": "usr_12345",
      "scope_type": "community",
      "scope_id": "community_futsal_uuid",
      "scope_name": "Futsal Players",
      "role": {
        "id": "role_leader_uuid",
        "name": "leader",
        "display_name": "Community Leader",
        "role_type": "community"
      },
      "permissions": [
        {
          "key": "post_content",
          "display_name": "Post Content",
          "scope": "community",
          "category": "content"
        },
        {
          "key": "moderate_posts",
          "display_name": "Moderate Posts",
          "scope": "community",
          "category": "moderation"
        },
        {
          "key": "manage_members",
          "display_name": "Manage Community Members",
          "scope": "community",
          "category": "member"
        },
        {
          "key": "delete_content",
          "display_name": "Delete Content",
          "scope": "community",
          "category": "moderation"
        }
      ]
    }
  }
  ```

- **Response (Error 404 — User not member of scope)**:
  ```json
  {
    "message": "You are not a member of this scope or scope not found"
  }
  ```

---

### 5. Get Role Details

Ambil detail satu role yang user miliki.

- **URL**: `GET /api/v1/mobile/role-permission/roles/{user_role_id}`
- **Autentikasi**: Yes
- **URL Parameters**:
  - `user_role_id`: ID of user's role assignment

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "user_role_uuid_1",
      "role_id": "role_admin_uuid",
      "role": {
        "id": "role_admin_uuid",
        "name": "admin",
        "display_name": "Admin Region",
        "role_type": "system"
      },
      "scope_type": "region",
      "scope_id": "region_jakarta_uuid",
      "scope_name": "Jakarta",
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "expired_at": null,
      "is_active": true,
      "permissions": [
        {
          "key": "post_news",
          "display_name": "Post News Article",
          "scope": "global",
          "category": "content"
        },
        {
          "key": "moderate_posts",
          "display_name": "Moderate Posts",
          "scope": "community",
          "category": "moderation"
        }
      ]
    }
  }
  ```

---

### 6. Check Bulk Permissions

Check multiple permissions at once (untuk optimize requests).

- **URL**: `POST /api/v1/mobile/role-permission/check-bulk`
- **Autentikasi**: Yes
- **Request Body**:
  ```json
  {
    "permission_keys": ["post_news", "moderate_posts", "manage_members", "assign_role"]
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "user_id": "usr_12345",
      "checks": [
        {
          "permission_key": "post_news",
          "has_permission": true
        },
        {
          "permission_key": "moderate_posts",
          "has_permission": true
        },
        {
          "permission_key": "manage_members",
          "has_permission": true
        },
        {
          "permission_key": "assign_role",
          "has_permission": false
        }
      ]
    }
  }
  ```

---

### 7. Get Communities Where User is Leader/Moderator

Ambil list komunitas dimana user adalah leader atau moderator.

- **URL**: `GET /api/v1/mobile/role-permission/communities`
- **Autentikasi**: Yes
- **Query Parameters**:
  - `role` (optional): Filter by role (`leader`, `moderator`, atau kosong untuk semua)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "community_id": "community_futsal_uuid",
        "community_name": "Futsal Players",
        "community_image": "https://example.com/futsal.jpg",
        "role": {
          "id": "role_leader_uuid",
          "name": "leader",
          "display_name": "Community Leader"
        },
        "user_role_id": "user_role_uuid_1",
        "assigned_at": "2026-01-15T00:00:00.000Z"
      },
      {
        "community_id": "community_nature_uuid",
        "community_name": "Nature Lovers",
        "community_image": "https://example.com/nature.jpg",
        "role": {
          "id": "role_moderator_uuid",
          "name": "moderator",
          "display_name": "Community Moderator"
        },
        "user_role_id": "user_role_uuid_3",
        "assigned_at": "2026-03-20T00:00:00.000Z"
      }
    ]
  }
  ```

---

### 8. Get Regions Where User is Admin

Ambil list region dimana user adalah admin.

- **URL**: `GET /api/v1/mobile/role-permission/regions`
- **Autentikasi**: Yes

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "region_id": "region_jakarta_uuid",
        "region_name": "Jakarta",
        "role": {
          "id": "role_admin_uuid",
          "name": "admin",
          "display_name": "Admin Region"
        },
        "user_role_id": "user_role_uuid_1",
        "assigned_at": "2026-05-20T00:00:00.000Z"
      },
      {
        "region_id": "region_surabaya_uuid",
        "region_name": "Surabaya",
        "role": {
          "id": "role_admin_uuid",
          "name": "admin",
          "display_name": "Admin Region"
        },
        "user_role_id": "user_role_uuid_2",
        "assigned_at": "2026-05-21T00:00:00.000Z"
      }
    ]
  }
  ```

---

## Feature Flag / UI Control Pattern

Mobile app dapat menggunakan permission check untuk show/hide UI:

```javascript
// Example: Check if user can post news
const canPostNews = await checkPermission('post_news');
if (canPostNews) {
  showPostNewsButton();
} else {
  hidePostNewsButton();
}

// Example: Get all permissions for conditional logic
const userPerms = await getMyPermissions();
const permissionMap = userPerms.permissions.reduce((acc, p) => {
  acc[p.key] = true;
  return acc;
}, {});

// In UI rendering
if (permissionMap['create_store']) {
  renderCreateStoreTab();
}
```

---

## Caching Strategy (Client-Side)

Karena permission changes tidak frequent, mobile dapat cache permission list:

1. **On first login** → fetch `/me` endpoint, cache hasil ke local storage
2. **On app resume** → fetch `/me` dengan short cache TTL (e.g., 5 minutes)
3. **On permission change notification** (via push/websocket) → invalidate cache dan refresh
4. **Fallback** → jika cache miss, fetch on-demand

---

## Error Handling

### 401 Unauthorized

```json
{
  "message": "Unauthorized. Token invalid or expired."
}
```

### 403 Forbidden

```json
{
  "message": "You don't have permission to access this resource"
}
```

### 404 Not Found

```json
{
  "message": "Role or permission not found"
}
```

### 500 Server Error

```json
{
  "message": "Internal server error. Please try again later."
}
```

---

## Common Use Cases

### Use Case 1: Show Post News Button

```
1. User navigate ke News page
2. Frontend GET /api/v1/mobile/role-permission/check/post_news
3. If has_permission == true AND subscription_plan == "pro"
4. Show "Post News" button
5. Else, show upgrade prompt or disabled button
```

### Use Case 2: Show Community Management Panel

```
1. User navigate ke Community
2. Frontend GET /api/v1/mobile/role-permission/communities?role=leader
3. If response.data.length > 0
4. Show "Manage Communities" tab
5. User click → GET /api/v1/mobile/role-permission/scope/community/{communityID}
6. Get list of permissions user can do in this community
7. Render management UI accordingly (manage members, moderate posts, etc)
```

### Use Case 3: Check Multiple Actions at Once

```
1. User open settings page
2. Frontend POST /api/v1/mobile/role-permission/check-bulk
   {
     "permission_keys": ["create_store", "create_community", "post_news"]
   }
3. Response shows which features are available
4. Render settings options based on availability
```

### Use Case 4: Show Admin Dashboard

```
1. User login
2. Frontend GET /api/v1/mobile/role-permission/me
3. If roles contains role_type="system" (superadmin, admin)
4. Show admin tab in navigation
5. User click → show regional or platform dashboard based on scope
```

---

## Integration with Subscription Plan

Mobile client harus combine **subscription plan** dengan **permissions** untuk determine akses user:

### Plan-Gated Permissions

Beberapa permission memerlukan subscription plan tertentu:

```javascript
// Example: Check if user can post news
const userPerms = await getMyPermissions(); // GET /api/v1/mobile/role-permission/me
const canPostNews = userPerms.permissions.some(p => p.key === 'post_news') 
                    && userPerms.subscription_plan === 'pro';

if (canPostNews) {
  showPostNewsButton();
} else if (userPerms.subscription_plan !== 'pro') {
  showUpgradePrompt('Upgrade to Pro to post news');
} else {
  hidePostNewsButton();
}
```

### Role-Only Permissions

Beberapa permission hanya gated oleh role (plan doesn't matter):

```javascript
const canApproveNews = userPerms.permissions.some(p => p.key === 'approve_news');
// no need to check plan
```

### Reference: Permission Categories

Lihat `ROLE_PERMISSION_SYSTEM.md` → "Permission Check Logic with Subscription Integration" untuk:
- **Role-Only Permissions** (approval, admin actions)
- **Plan-Gated Permissions** (post_news, create_store, create_community)
- **Combined Permissions** (post_content with quotas, view_analytics with depth)
- **Community-Scoped Permissions** (moderate_posts in community)

---

## Notes

1. **No Write Operations**: Mobile API adalah read-only. Semua write operations (create permission, assign role) hanya di backoffice API.
2. **Subscription Plan Included**: Response `/me` mencakup `subscription_plan` untuk client-side permission checks.
3. **Permission Caching**: Mobile client dapat cache permission list (including plan) untuk reduce server load.
4. **Real-time Updates**: Jika ada permission/plan change (via notification), client harus invalidate cache dan refetch.
5. **Scope Awareness**: Untuk permission yang scope-specific, client harus pass scope context.
6. **Plan-Permission Combo**: Client harus check BOTH subscription plan AND permission untuk plan-gated features.

