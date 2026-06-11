# API Spec — User Management (Backoffice) v2

> **Update dari v1:**
> - Fix `User Object (List)` — `region` sekarang object `{id, name}`, bukan string flat
> - Fix `role` di list object — sekarang array, bukan single string
> - Tambah endpoint `POST /users/{id}/roles` — assign role single user
> - Tambah endpoint `DELETE /users/{id}/roles/{assignment_id}` — revoke role single user
> - Tambah endpoint `GET /users/{id}/roles/bulk` — assign multiple roles ke satu user
> - Tambah endpoint `GET /users/{id}/audit-log` — riwayat aksi admin pada user
> - Fix authorization matrix — suspend hanya superadmin, bukan admin region
> - Fix `GET /users` — admin region tidak bisa pakai region_id untuk akses region lain
> - Tambah endpoint `GET /users/stats` — ringkasan statistik platform

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/users`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin / Admin Region)
- **Authorization**:
  - Superadmin: akses semua user di semua region
  - Admin Region: hanya user di region mereka sendiri (auto-filtered dari JWT, tidak perlu kirim region_id)

---

## Authorization Matrix (Updated)

| Aksi | superadmin | admin region | Notes |
|---|:---:|:---:|---|
| List semua user platform | ✅ | ❌ | Admin region auto-filtered ke region sendiri |
| List user di region sendiri | ✅ | ✅ | |
| Lihat detail user | ✅ | ✅ (region sendiri) | |
| Edit profil user | ✅ | ✅ (region sendiri) | name, email, phone, avatar |
| Reset password | ✅ | ✅ (region sendiri) | |
| Activate / deactivate | ✅ | ✅ (region sendiri) | |
| **Suspend user** | ✅ | ❌ | Hanya superadmin |
| Soft delete akun | ✅ | ❌ | Hanya superadmin |
| Assign role member | ✅ | ✅ (region sendiri) | |
| Assign role admin region | ✅ | ✅ (region sendiri, + konfirmasi) | |
| **Assign role superadmin** | ✅ | ❌ | Hanya superadmin |
| Revoke role | ✅ | ✅ (yang dia assign saja) | |
| **Ubah subscription plan** | ✅ | ❌ | Hanya superadmin |
| Export CSV | ✅ | ✅ (region sendiri) | |
| Lihat audit log user | ✅ | ✅ (region sendiri) | |
| Kirim pesan ke user | ✅ | ✅ (region sendiri) | |

---

## Model Data

### 1. User Object (Full) — untuk GET /users/{id}

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
    "id": "region-jakarta-uuid",
    "name": "Jakarta"
  },
  "roles": [
    {
      "assignment_id": "ur_uuid",
      "role_id": "role-uuid",
      "role_name": "admin",
      "display_name": "Admin Region",
      "role_type": "system",
      "scope_type": "region",
      "scope_id": "region-jakarta-uuid",
      "scope_name": "Jakarta",
      "assigned_by": "superadmin_uuid",
      "assigned_at": "2026-05-20T00:00:00.000Z",
      "expired_at": null
    }
  ],
  "permissions": [
    { "key": "post_news",     "scope": "global" },
    { "key": "manage_region", "scope": "regional" }
  ],
  "stats": {
    "total_communities": 3,
    "posts_created": 45,
    "days_active": 130
  },
  "created_at": "2026-01-15T00:00:00.000Z",
  "updated_at": "2026-05-25T10:00:00.000Z",
  "last_login": "2026-05-25T09:30:00.000Z"
}
```

### 2. User Object (List) — untuk GET /users

```json
{
  "id": "uuid",
  "name": "Andi Pratama",
  "email": "andi@example.com",
  "username": "andipratama",
  "avatar": "https://example.com/avatar.jpg",
  "status": "active",
  "subscription_plan": "pro",
  "subscription_status": "active",
  "region": {
    "id": "region-jakarta-uuid",
    "name": "Jakarta"
  },
  "roles": [
    { "role_name": "admin", "scope_type": "region", "scope_name": "Jakarta" }
  ],
  "created_at": "2026-01-15T00:00:00.000Z",
  "last_login": "2026-05-25T09:30:00.000Z"
}
```

> **Catatan update:** `region` sekarang object `{id, name}` bukan string. `roles` sekarang array bukan single string — karena satu user bisa multi-role.

---

## Endpoints

### 1. List Users

- **URL**: `GET /api/v1/web/users`
- **Auth**: Required (Superadmin / Admin Region)

**Behavior per aktor:**
- Superadmin: return semua user, bisa filter by `region_id`
- Admin Region: auto-filtered ke region sendiri dari JWT, parameter `region_id` **di-ignore** (tidak bisa lihat region lain)

**Query Parameters:**

| Parameter | Type | Keterangan |
|---|---|---|
| `search` | string | Cari nama, email, atau username |
| `status` | string | `active`, `inactive`, `suspended` |
| `subscription_plan` | string | `standard`, `pro` |
| `subscription_status` | string | `active`, `expired`, `cancelled` |
| `role` | string | Filter by role name: `admin`, `member`, dll |
| `region_id` | string | **Superadmin only** — filter by region. Admin region: param ini di-ignore |
| `sort` | string | `created_at`, `last_login`, `name` (prefix `-` untuk desc) |
| `limit` | int | Default 20, max 100 |
| `offset` | int | Default 0 |
| `export` | string | `csv` atau `json` — return file download |

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "uuid1",
      "name": "Andi Pratama",
      "email": "andi@example.com",
      "username": "andipratama",
      "avatar": "https://example.com/avatar.jpg",
      "status": "active",
      "subscription_plan": "pro",
      "subscription_status": "active",
      "region": {
        "id": "region-jakarta-uuid",
        "name": "Jakarta"
      },
      "roles": [
        { "role_name": "admin", "scope_type": "region", "scope_name": "Jakarta" }
      ],
      "created_at": "2026-01-15T00:00:00.000Z",
      "last_login": "2026-05-25T09:30:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 245,
    "pages": 13
  },
  "meta": {
    "filtered_by_region": "Jakarta",
    "actor_role": "admin"
  }
}
```

> Field `meta.filtered_by_region` berguna untuk frontend — admin region bisa tahu filter aktif meski tidak kirim region_id manual.

---

### 2. Get User Detail

- **URL**: `GET /api/v1/web/users/{user_id}`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region sendiri saja

**Response (200 OK)**: Lihat User Object (Full) di atas.

**Response (403 Forbidden)**:
```json
{
  "message": "Admin can only view users in their own region"
}
```

---

### 3. Edit User Profile

- **URL**: `PATCH /api/v1/web/users/{user_id}`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region sendiri

**Request Body** (semua optional, kirim yang mau diubah saja):
```json
{
  "name": "Andi Pratama Updated",
  "email": "andi.new@example.com",
  "phone": "08123456790",
  "avatar": "https://example.com/new-avatar.jpg"
}
```

> username tidak bisa diubah oleh admin — hanya user sendiri.

**Response (200 OK)**:
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
  "message": "Profil user berhasil diperbarui"
}
```

---

### 4. Change User Status

- **URL**: `PATCH /api/v1/web/users/{user_id}/status`
- **Auth**: Required

**Authorization per status:**
- `active` ↔ `inactive`: superadmin dan admin region (region sendiri)
- `active` / `inactive` → `suspended`: **superadmin only**
- `suspended` → `active` / `inactive`: **superadmin only**

**Request Body**:
```json
{
  "status": "suspended",
  "reason": "Spam berulang di 3 komunitas dalam 24 jam"
}
```

> `reason` wajib untuk semua perubahan status — dicatat di audit log.

**Response (200 OK)**:
```json
{
  "data": {
    "user_id": "uuid",
    "old_status": "active",
    "new_status": "suspended",
    "reason": "Spam berulang di 3 komunitas dalam 24 jam",
    "changed_at": "2026-05-25T10:35:00.000Z",
    "changed_by": "superadmin_uuid"
  },
  "message": "Status user berubah dari active ke suspended"
}
```

**Response (403 — Admin region coba suspend)**:
```json
{
  "message": "Hanya superadmin yang bisa suspend atau unsuspend user"
}
```

---

### 5. Reset Password

- **URL**: `POST /api/v1/web/users/{user_id}/reset-password`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region sendiri

**Request Body**:
```json
{
  "send_email": true,
  "note": "User lupa password, diminta via support"
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "user_id": "uuid",
    "email": "andi@example.com",
    "reset_link_expires_in_hours": 24,
    "sent_at": "2026-05-25T10:40:00.000Z"
  },
  "message": "Link reset password dikirim ke email user"
}
```

---

### 6. Change Subscription Plan

- **URL**: `PATCH /api/v1/web/users/{user_id}/subscription`
- **Auth**: Required
- **Authorization**: **Superadmin only**

**Request Body**:
```json
{
  "plan": "pro",
  "effective_date": "2026-05-25",
  "reason": "Manual upgrade — transfer dikonfirmasi via WA",
  "skip_approval": true
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "user_id": "uuid",
    "old_plan": "standard",
    "new_plan": "pro",
    "effective_date": "2026-05-25",
    "changed_at": "2026-05-25T10:45:00.000Z",
    "changed_by": "superadmin_uuid"
  },
  "message": "Subscription berubah dari standard ke pro"
}
```

---

### 7. Assign Role ke User (Single)

**Endpoint baru** — assign satu role ke satu user dari halaman detail user.

- **URL**: `POST /api/v1/web/users/{user_id}/roles`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region (hanya ke region sendiri, tidak bisa assign superadmin)

**Request Body**:
```json
{
  "role_id": "role-uuid",
  "scope_type": "region",
  "scope_id": "region-jakarta-uuid",
  "expired_at": "2026-12-31T23:59:59.000Z",
  "notes": "Admin baru untuk region Jakarta"
}
```

**Validation:**
- `scope_id` wajib jika `scope_type` adalah `region` atau `community`
- `scope_id` null jika `scope_type` adalah `global`
- Admin region tidak bisa assign role `superadmin`
- Admin region hanya bisa set `scope_id` ke region mereka sendiri

**Response (201 Created)**:
```json
{
  "data": {
    "assignment_id": "ur_uuid",
    "user_id": "uuid",
    "role_id": "role-uuid",
    "role_name": "admin",
    "display_name": "Admin Region",
    "scope_type": "region",
    "scope_id": "region-jakarta-uuid",
    "scope_name": "Jakarta",
    "assigned_by": "superadmin_uuid",
    "assigned_at": "2026-05-25T10:50:00.000Z",
    "expired_at": "2026-12-31T23:59:59.000Z"
  },
  "message": "Role Admin Region berhasil di-assign ke Andi Pratama"
}
```

**Response (409 — Sudah punya role ini di scope ini)**:
```json
{
  "message": "User sudah memiliki role ini di scope yang sama",
  "data": {
    "existing_assignment_id": "ur_existing_uuid",
    "assigned_at": "2026-01-01T00:00:00.000Z"
  }
}
```

---

### 8. Assign Multiple Roles ke User (Multi-Role)

**Endpoint baru** — assign beberapa role sekaligus ke satu user dalam satu operasi.

- **URL**: `POST /api/v1/web/users/{user_id}/roles/bulk`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region (scope terbatas)

**Request Body**:
```json
{
  "assignments": [
    {
      "role_id": "role-admin-uuid",
      "scope_type": "region",
      "scope_id": "region-jakarta-uuid",
      "expired_at": "2026-12-31T23:59:59.000Z",
      "notes": "Admin Jakarta"
    },
    {
      "role_id": "role-moderator-uuid",
      "scope_type": "community",
      "scope_id": "community-futsal-uuid",
      "expired_at": null,
      "notes": "Moderator komunitas futsal"
    }
  ]
}
```

**Response (200 OK)** — partial success supported:
```json
{
  "data": {
    "user_id": "uuid",
    "total": 2,
    "successful": 2,
    "failed": 0,
    "skipped": 0,
    "results": [
      {
        "role_id": "role-admin-uuid",
        "role_name": "admin",
        "scope_name": "Jakarta",
        "status": "success",
        "assignment_id": "ur_new_001"
      },
      {
        "role_id": "role-moderator-uuid",
        "role_name": "moderator",
        "scope_name": "Futsal Jakarta",
        "status": "success",
        "assignment_id": "ur_new_002"
      }
    ]
  },
  "message": "2 role berhasil di-assign ke Andi Pratama"
}
```

---

### 9. Revoke Role dari User

- **URL**: `DELETE /api/v1/web/users/{user_id}/roles/{assignment_id}`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region (hanya role yang dia assign di region sendiri)

**Request Body**:
```json
{
  "reason": "Reorganisasi admin regional Q3 2026"
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "assignment_id": "ur_uuid",
    "user_id": "uuid",
    "role_name": "admin",
    "scope_name": "Jakarta",
    "revoked_at": "2026-05-25T11:00:00.000Z",
    "revoked_by": "superadmin_uuid"
  },
  "message": "Role Admin Region Jakarta berhasil di-revoke dari Andi Pratama"
}
```

**Response (403 — Admin coba revoke role yang bukan dia assign)**:
```json
{
  "message": "Admin hanya bisa revoke role yang dia sendiri assign"
}
```

**Response (422 — Coba revoke role member system)**:
```json
{
  "message": "Role member sistem tidak bisa di-revoke — ini adalah role default semua user"
}
```

---

### 10. Get User Audit Log

**Endpoint baru** — riwayat semua aksi admin yang dilakukan pada user ini.

- **URL**: `GET /api/v1/web/users/{user_id}/audit-log`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region sendiri

**Query Parameters:**
- `limit` (default 20, max 100)
- `offset` (default 0)
- `action` (optional) — filter by action: `role.assigned`, `user.status_changed`, dll

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "log_uuid",
      "actor_id": "superadmin_uuid",
      "actor_name": "Super Admin",
      "actor_role": "superadmin",
      "action": "role.assigned",
      "target_type": "role_assignment",
      "target_id": "ur_uuid",
      "old_value": null,
      "new_value": {
        "role_name": "admin",
        "scope_type": "region",
        "scope_name": "Jakarta"
      },
      "notes": "Admin baru untuk region Jakarta",
      "ip_address": "103.10.xx.xx",
      "created_at": "2026-05-25T10:50:00.000Z"
    },
    {
      "id": "log_uuid_2",
      "actor_id": "superadmin_uuid",
      "actor_name": "Super Admin",
      "actor_role": "superadmin",
      "action": "user.status_changed",
      "target_type": "user",
      "target_id": "user_uuid",
      "old_value": { "status": "active" },
      "new_value": { "status": "suspended" },
      "notes": "Spam berulang",
      "ip_address": "103.10.xx.xx",
      "created_at": "2026-05-20T08:00:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 12
  }
}
```

---

### 11. View User Subscription History

- **URL**: `GET /api/v1/web/users/{user_id}/subscription-history`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region sendiri

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "sub_hist_uuid",
      "action": "upgraded",
      "old_plan": "standard",
      "new_plan": "pro",
      "changed_by": "superadmin_uuid",
      "changed_by_name": "Super Admin",
      "reason": "Manual upgrade — transfer dikonfirmasi",
      "effective_date": "2026-05-25",
      "created_at": "2026-05-25T10:45:00.000Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 3 }
}
```

---

### 12. Send Message to User

- **URL**: `POST /api/v1/web/users/{user_id}/send-message`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region sendiri

**Request Body**:
```json
{
  "subject": "Pemberitahuan Penting",
  "message": "Akun kamu telah diperbarui.",
  "message_type": "email",
  "schedule_at": null
}
```

`message_type`: `email` | `in_app` | `both`

**Response (200 OK)**:
```json
{
  "data": {
    "message_id": "msg_001",
    "status": "sent",
    "sent_at": "2026-05-25T11:05:00.000Z"
  },
  "message": "Pesan berhasil dikirim"
}
```

---

### 13. Export Users

- **URL**: `GET /api/v1/web/users?export=csv`
- **Auth**: Required
- **Authorization**: Superadmin all, Admin region sendiri

Semua query filter berlaku. Admin region hanya bisa export user di region mereka.

**Response**: File download
```
Content-Type: text/csv
Content-Disposition: attachment; filename=users_20260525.csv

id,name,email,status,subscription_plan,region_name,created_at
uuid1,Andi Pratama,andi@example.com,active,pro,Jakarta,2026-01-15
```

---

### 14. Platform User Stats

**Endpoint baru** — ringkasan statistik user di platform (superadmin) atau di region (admin region).

- **URL**: `GET /api/v1/web/users/stats`
- **Auth**: Required

**Response (200 OK)**:
```json
{
  "data": {
    "scope": "region",
    "scope_name": "Jakarta",
    "total_users": 1240,
    "active": 1180,
    "inactive": 45,
    "suspended": 15,
    "standard_plan": 1050,
    "pro_plan": 190,
    "new_this_month": 48,
    "churned_this_month": 12
  }
}
```

---

## Error Responses

```json
// 403 — Admin region akses user di region lain
{ "message": "Admin hanya bisa mengelola user di region mereka sendiri" }

// 403 — Admin coba suspend
{ "message": "Hanya superadmin yang bisa suspend atau unsuspend user" }

// 403 — Admin coba assign superadmin role
{ "message": "Hanya superadmin yang bisa assign role superadmin" }

// 403 — Admin coba assign ke region lain
{ "message": "Admin hanya bisa assign role di region mereka sendiri" }

// 404
{ "message": "User tidak ditemukan" }

// 409 — Duplicate role assignment
{ "message": "User sudah memiliki role ini di scope yang sama" }

// 422 — Revoke role member default
{ "message": "Role member sistem tidak bisa di-revoke" }
```
