# API Spec — User Management (Backoffice)

Dokumentasi API untuk superadmin/admin manage users di backoffice dashboard — view, search, edit profile, manage subscription, view roles & permissions.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/users`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin / Admin Region)
- **Authorization**:
  - Superadmin dapat view/edit all users
  - Admin Region hanya dapat view/edit users dalam region mereka
- **Error Format**: Same as other backoffice APIs (standard message or validation error)

---

## Model Data Utama

### 1. User Object (Full)

```json
{
  "id": "uuid",
  "name": "Andi Pratama",
  "email": "andi@example.com",
  "username": "andipratama",
  "phone": "08123456789",
  "avatar": "https://example.com/avatar.jpg",
  "status": "active",
  "email_verified": true,
  "phone_verified": false,
  "subscription_plan": "pro",
  "subscription_status": "active",
  "subscription_expired_at": "2026-06-24T23:59:59.000Z",
  "region": {
    "id": "uuid",
    "name": "Jakarta"
  },
  "roles": [
    {
      "id": "uuid",
      "role_id": "uuid",
      "role": {
        "id": "uuid",
        "name": "admin",
        "display_name": "Admin Region",
        "role_type": "system"
      },
      "scope_type": "region",
      "scope_id": "region_jakarta",
      "scope_name": "Jakarta",
      "assigned_at": "2026-05-20T00:00:00.000Z"
    }
  ],
  "created_at": "2026-01-15T00:00:00.000Z",
  "updated_at": "2026-05-25T10:00:00.000Z",
  "last_login": "2026-05-25T09:30:00.000Z"
}
```

### 2. User Object (List)

```json
{
  "id": "uuid",
  "name": "Andi Pratama",
  "email": "andi@example.com",
  "status": "active",
  "subscription_plan": "pro",
  "subscription_status": "active",
  "region": "Jakarta",
  "role": "admin",
  "created_at": "2026-01-15T00:00:00.000Z",
  "last_login": "2026-05-25T09:30:00.000Z"
}
```

---

## Endpoints

### 1. List Users

Ambil daftar users dengan filtering, search, pagination.

- **URL**: `GET /api/v1/web/users`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin see all, Admin see own region only
- **Query Parameters**:
  - `search` (optional): Search by name/email/phone
  - `status` (optional): Filter by status (`active`, `inactive`, `suspended`)
  - `subscription_plan` (optional): Filter by plan (`standard`, `pro`)
  - `subscription_status` (optional): Filter by subscription (`active`, `expired`, `cancelled`)
  - `role` (optional): Filter by role (`admin`, `member`, `guest`, etc)
  - `region_id` (optional): Filter by region
  - `sort` (optional): Sort field (`created_at`, `last_login`, `name`), prefix `-` for desc (e.g., `-created_at`)
  - `limit` (optional, default: 20, max: 100): Pagination limit
  - `offset` (optional, default: 0): Pagination offset
  - `export` (optional): Export format (`csv`, `json`) — returns file instead of JSON

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid1",
        "name": "Andi Pratama",
        "email": "andi@example.com",
        "status": "active",
        "subscription_plan": "pro",
        "subscription_status": "active",
        "region": "Jakarta",
        "role": "admin",
        "created_at": "2026-01-15T00:00:00.000Z",
        "last_login": "2026-05-25T09:30:00.000Z"
      },
      {
        "id": "uuid2",
        "name": "Budi Santoso",
        "email": "budi@example.com",
        "status": "active",
        "subscription_plan": "standard",
        "subscription_status": "active",
        "region": "Jakarta",
        "role": "member",
        "created_at": "2026-02-20T00:00:00.000Z",
        "last_login": "2026-05-24T15:45:00.000Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 245,
      "pages": 13
    },
    "filters": {
      "status": "active",
      "region_id": "region_jakarta"
    }
  }
  ```

---

### 2. Get User Detail

Ambil detail lengkap satu user.

- **URL**: `GET /api/v1/web/users/{user_id}`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin see all, Admin see own region only
- **URL Parameters**:
  - `user_id`: ID dari user

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "name": "Andi Pratama",
      "email": "andi@example.com",
      "username": "andipratama",
      "phone": "08123456789",
      "avatar": "https://example.com/avatar.jpg",
      "status": "active",
      "email_verified": true,
      "phone_verified": false,
      "subscription_plan": "pro",
      "subscription_status": "active",
      "subscription_expired_at": "2026-06-24T23:59:59.000Z",
      "region": {
        "id": "uuid",
        "name": "Jakarta"
      },
      "roles": [
        {
          "id": "uuid",
          "role_id": "uuid",
          "role": {
            "id": "uuid",
            "name": "admin",
            "display_name": "Admin Region",
            "role_type": "system"
          },
          "scope_type": "region",
          "scope_id": "region_jakarta",
          "scope_name": "Jakarta",
          "assigned_at": "2026-05-20T00:00:00.000Z"
        }
      ],
      "permissions": [
        {
          "key": "post_news",
          "display_name": "Post News",
          "scope": "global"
        },
        {
          "key": "manage_region",
          "display_name": "Manage Regional Settings",
          "scope": "regional"
        }
      ],
      "created_at": "2026-01-15T00:00:00.000Z",
      "updated_at": "2026-05-25T10:00:00.000Z",
      "last_login": "2026-05-25T09:30:00.000Z"
    },
    "stats": {
      "total_communities": 3,
      "posts_created": 45,
      "days_active": 130
    }
  }
  ```

---

### 3. Edit User Profile

Edit profil user (name, email, phone, avatar).

- **URL**: `PUT /api/v1/web/users/{user_id}`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin edit all, Admin only own region

- **Request Body**:
  ```json
  {
    "name": "Andi Pratama Updated",
    "email": "andi.new@example.com",
    "phone": "08123456790",
    "avatar": "https://example.com/new-avatar.jpg"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "uuid",
      "name": "Andi Pratama Updated",
      "email": "andi.new@example.com",
      "phone": "08123456790",
      "avatar": "https://example.com/new-avatar.jpg",
      "updated_at": "2026-05-25T10:30:00.000Z"
    },
    "message": "User profile updated successfully"
  }
  ```

---

### 4. Change User Status

Suspend, activate, atau deactivate user account.

- **URL**: `PATCH /api/v1/web/users/{user_id}/status`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin only untuk suspend, Admin bisa activate/deactivate own region

- **Request Body**:
  ```json
  {
    "status": "active" | "inactive" | "suspended",
    "reason": "User requested account deactivation"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "user_id": "uuid",
      "old_status": "active",
      "new_status": "suspended",
      "reason": "User requested account deactivation",
      "changed_at": "2026-05-25T10:35:00.000Z",
      "changed_by": "superadmin_uuid"
    },
    "message": "User status changed from active to suspended"
  }
  ```

---

### 5. Reset User Password

Admin reset password untuk user (send reset link via email).

- **URL**: `POST /api/v1/web/users/{user_id}/reset-password`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin all, Admin own region

- **Request Body**:
  ```json
  {
    "send_email": true,
    "note": "User forgot password"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "user_id": "uuid",
      "reset_token_sent": true,
      "email": "andi@example.com",
      "reset_link_expires_in_hours": 24
    },
    "message": "Password reset link sent to user's email"
  }
  ```

---

### 6. Change User Subscription

Admin upgrade/downgrade user subscription plan.

- **URL**: `PATCH /api/v1/web/users/{user_id}/subscription`
- **Autentikasi**: Yes (Superadmin only)
- **Authorization**: Superadmin only

- **Request Body**:
  ```json
  {
    "plan": "pro",
    "effective_date": "2026-05-25",
    "reason": "Manual admin upgrade",
    "skip_approval": true
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "user_id": "uuid",
      "old_plan": "standard",
      "new_plan": "pro",
      "old_status": "active",
      "new_status": "active",
      "effective_date": "2026-05-25",
      "reason": "Manual admin upgrade",
      "changed_at": "2026-05-25T10:40:00.000Z"
    },
    "message": "Subscription changed from standard to pro"
  }
  ```

---

### 7. View User Subscription History

Ambil riwayat subscription changes untuk user.

- **URL**: `GET /api/v1/web/users/{user_id}/subscription-history`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin all, Admin own region
- **Query Parameters**:
  - `limit` (optional, default: 20)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "action": "upgrade",
        "old_plan": "standard",
        "new_plan": "pro",
        "initiated_by": "system",
        "reason": "User requested upgrade",
        "created_at": "2026-05-10T10:00:00.000Z"
      },
      {
        "id": "uuid",
        "action": "renewal",
        "plan": "pro",
        "initiated_by": "system",
        "created_at": "2026-04-10T00:00:00.000Z"
      },
      {
        "id": "uuid",
        "action": "signup",
        "plan": "standard",
        "initiated_by": "system",
        "created_at": "2026-01-15T00:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 3
    }
  }
  ```

---

### 8. View User Activity Log

Ambil activity log user (login, actions, etc).

- **URL**: `GET /api/v1/web/users/{user_id}/activity-log`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin all, Admin own region
- **Query Parameters**:
  - `action_type` (optional): Filter by type (`login`, `post`, `comment`, etc)
  - `date_from` (optional): Filter from date
  - `date_to` (optional): Filter to date
  - `limit` (optional, default: 50, max: 200)
  - `offset` (optional, default: 0)

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "action_type": "login",
        "details": "Logged in from 192.168.1.1",
        "created_at": "2026-05-25T09:30:00.000Z"
      },
      {
        "id": "uuid",
        "action_type": "post_created",
        "details": "Created post in community Futsal",
        "resource_id": "post_123",
        "created_at": "2026-05-25T08:15:00.000Z"
      },
      {
        "id": "uuid",
        "action_type": "profile_updated",
        "details": "Updated profile information",
        "created_at": "2026-05-24T14:20:00.000Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "total": 156
    }
  }
  ```

---

### 9. Send Message to User

Send announcement atau notification ke user.

- **URL**: `POST /api/v1/web/users/{user_id}/send-message`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin all, Admin own region

- **Request Body**:
  ```json
  {
    "subject": "Important Update",
    "message": "We have updated our community guidelines. Please review them.",
    "message_type": "email" | "in_app" | "both",
    "schedule_at": "2026-05-26T10:00:00.000Z"
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": {
      "user_id": "uuid",
      "message_id": "msg_001",
      "subject": "Important Update",
      "type": "email",
      "status": "scheduled",
      "scheduled_at": "2026-05-26T10:00:00.000Z",
      "sent_at": null
    },
    "message": "Message scheduled to send on 2026-05-26 at 10:00"
  }
  ```

---

### 10. Export Users

Export user list ke CSV atau JSON.

- **URL**: `GET /api/v1/web/users?export=csv` atau `?export=json`
- **Autentikasi**: Yes (Superadmin / Admin)
- **Authorization**: Superadmin all, Admin own region
- **Query Parameters**:
  - `export`: Required, `csv` atau `json`
  - Semua filter parameters dari endpoint List Users juga berlaku

- **Response (Success 200)**: File download
  ```
  Content-Type: text/csv
  Content-Disposition: attachment; filename=users_20260525.csv

  id,name,email,status,subscription_plan,region,created_at
  uuid1,Andi Pratama,andi@example.com,active,pro,Jakarta,2026-01-15
  uuid2,Budi Santoso,budi@example.com,active,standard,Jakarta,2026-02-20
  ...
  ```

---

## UI Flow Example

### Page: User Management

```
┌─────────────────────────────────────────────────────────────┐
│ User Management Dashboard                                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ SEARCH & FILTER:                                            │
│ Search: [____________] [Search]                            │
│                                                             │
│ Filters:                                                    │
│ Status: [All ▼]  Plan: [All ▼]  Region: [All ▼]           │
│ Sort: [Created Date ▼] [Export as CSV] [Export as JSON]   │
│                                                             │
│ USERS LIST:                                                 │
│ ┌───┬──────────┬──────────────┬────────┬──────┬───────┐    │
│ │ # │ Name     │ Email        │ Plan   │ Role │ Action│    │
│ ├───┼──────────┼──────────────┼────────┼──────┼───────┤    │
│ │1  │Andi P.   │andi@ex.com   │Pro     │Admin │[View] │    │
│ │2  │Budi S.   │budi@ex.com   │Std     │Memb  │[View] │    │
│ │3  │Citra D.  │citra@ex.com  │Pro     │Memb  │[View] │    │
│ └───┴──────────┴──────────────┴────────┴──────┴───────┘    │
│                                                             │
│ Page 1 of 13 | Showing 1-20 of 245 users                   │
│                                                             │
│ [Previous] [1] [2] [3] ... [13] [Next]                     │
└─────────────────────────────────────────────────────────────┘
```

### Modal: User Detail & Actions

```
┌─────────────────────────────────────────────────────────────┐
│ User Detail: Andi Pratama                          [Close]  │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ PROFILE:                                                    │
│ Name: Andi Pratama        | Edit                           │
│ Email: andi@example.com   | Verified ✓                     │
│ Phone: 08123456789        | Not verified                   │
│ Status: Active            | [Change Status ▼]              │
│                                                             │
│ SUBSCRIPTION:                                               │
│ Plan: Pro                 | [Change Plan]                  │
│ Status: Active            |                                │
│ Expires: 2026-06-24       |                                │
│                                                             │
│ ROLES & PERMISSIONS:                                        │
│ Roles:                                                      │
│ ├─ Admin (Region: Jakarta) - Assigned 2026-05-20          │
│                                                             │
│ Permissions:                                                │
│ ├─ post_news                                               │
│ ├─ manage_region                                           │
│                                                             │
│ ACCOUNT ACTIONS:                                            │
│ [Reset Password] [Send Message] [View Activity]            │
│ [View Subscription History] [Export User Data]             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Important Notes

### ✅ DO:
- ✅ Check authorization per user/region
- ✅ Log all admin actions (audit trail)
- ✅ Support search dan filtering
- ✅ Show user stats (communities, posts, etc)
- ✅ Allow export untuk reporting

### ❌ DON'T:
- ❌ Admin region bisa lihat users dari region lain
- ❌ Allow change user role tanpa explicit bulk-assign endpoint
- ❌ Silent fail (always return meaningful errors)
- ❌ Expose password hashes atau sensitive data

### Authorization Matrix:

| Action | Superadmin | Admin Region | Notes |
|--------|-----------|--------------|-------|
| List users | All | Own region only | Filter by region |
| View detail | All | Own region only | Check region before return |
| Edit profile | All | Own region only | Name, email, phone, avatar |
| Change status | All | Own region only | Active, inactive, suspended |
| Reset password | All | Own region only | Send via email |
| Change subscription | All | No | Superadmin only |
| View activity | All | Own region only | Login, posts, etc |
| Send message | All | Own region only | Email or in-app |

---

## Error Handling

Standard error responses:

```json
// 400 Bad Request
{
  "message": "Invalid request body or parameters"
}

// 401 Unauthorized
{
  "message": "Authentication required"
}

// 403 Forbidden
{
  "message": "Admin can only manage users in their own region"
}

// 404 Not Found
{
  "message": "User not found"
}

// 422 Unprocessable Entity
{
  "message": "Validation failed",
  "errors": {
    "email": ["Email already in use"],
    "status": ["Invalid status value"]
  }
}
```

---

*API spec untuk user management backoffice. Gunakan untuk admin dashboard implementation.*
