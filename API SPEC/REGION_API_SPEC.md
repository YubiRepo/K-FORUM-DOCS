# API Spec — Region Module

**Status:** Draft v1  
**Last Updated:** 2026-05-26  
**Base URL Prefix Mobile:** `/api/v1/mobile/regions`  
**Base URL Prefix Backoffice:** `/api/v1/web/regions`

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Model Data](#model-data)
3. [Mobile Client Endpoints](#mobile-client-endpoints)
4. [Backoffice Endpoints](#backoffice-endpoints)
5. [Error Handling](#error-handling)

---

## Informasi Umum

### Headers Global

**Mobile & Backoffice:**
```
Content-Type: application/json
Accept: application/json
Authorization: Bearer <access_token> (Required, kecuali untuk public endpoints)
Accept-Language: <lang_code> (e.g., ko, id, en. Default: ko)
```

### Authentication

- Mobile endpoints: Semua require authenticated user (Bearer token)
- Backoffice endpoints: Semua require superadmin atau admin region (role check)

### Authorization

| Role | Mobile | Backoffice |
|------|--------|-----------|
| Superadmin | All mobile endpoints | All backoffice endpoints |
| Admin Region | All mobile endpoints | Own region endpoints only |
| Member | Browse, request, accept invite | ❌ No access |
| Guest | Browse regions (some) | ❌ No access |

---

## Model Data

### Region Object

```json
{
  "id": "uuid",
  "name": "KAI Jakarta",
  "slug": "jakarta",
  "description": "Wilayah Jakarta dan sekitarnya",
  "image_url": "https://cdn.example.com/regions/jakarta.jpg",
  "status": "active",
  "member_count": 1245,
  "your_status": "active" | "pending_approval" | "rejected" | null,
  "your_role": "admin" | "member" | null,
  "created_at": "2026-05-20T10:00:00.000Z",
  "updated_at": "2026-05-26T09:30:00.000Z"
}
```

### RegionMembership Object

```json
{
  "id": "uuid",
  "user_id": "uuid",
  "region_id": "uuid",
  "region_name": "KAI Jakarta",
  "role": "admin" | "member",
  "status": "active" | "pending_approval" | "rejected",
  "joined_at": "2026-05-20T10:00:00.000Z",
  "approval_notes": "Welcome to region",
  "rejection_reason": null,
  "approved_by": "uuid",
  "created_at": "2026-05-20T09:00:00.000Z"
}
```

### RegionInvitation Object

```json
{
  "id": "uuid",
  "region_id": "uuid",
  "region_name": "KAI Jakarta",
  "email": "user@example.com",
  "status": "pending" | "accepted" | "rejected" | "expired",
  "invited_by_name": "Admin Name",
  "invited_by_avatar": "https://...",
  "created_at": "2026-05-26T10:00:00.000Z",
  "expires_at": "2026-05-27T10:00:00.000Z",
  "accepted_at": null,
  "message": "Join region untuk akses content lokal"
}
```

---

## Mobile Client Endpoints

### 1. Get All Regions

Ambil daftar semua region yang tersedia (active only).

- **URL:** `GET /api/v1/mobile/regions`
- **Autentikasi:** Optional (better results jika authenticated)
- **Query Parameters:**
  - `search` (optional): Search by name atau slug
  - `limit` (optional, default: 50, max: 100)
  - `offset` (optional, default: 0)

- **Response (200 OK):**
  ```json
  {
    "data": [
      {
        "id": "region_jakarta",
        "name": "KAI Jakarta",
        "slug": "jakarta",
        "description": "Wilayah Jakarta dan sekitarnya",
        "image_url": "https://cdn.example.com/regions/jakarta.jpg",
        "status": "active",
        "member_count": 1245,
        "your_status": null,
        "your_role": null,
        "created_at": "2026-05-20T10:00:00.000Z"
      },
      {
        "id": "region_surabaya",
        "name": "KAI Surabaya",
        "slug": "surabaya",
        "description": "Wilayah Surabaya dan sekitarnya",
        "image_url": "https://cdn.example.com/regions/surabaya.jpg",
        "status": "active",
        "member_count": 892,
        "your_status": "pending_approval",
        "your_role": "member",
        "created_at": "2026-05-20T10:00:00.000Z"
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

### 2. Get My Region

Ambil info region yang user sekarang aktif, atau null jika belum bergabung.

- **URL:** `GET /api/v1/mobile/regions/me`
- **Autentikasi:** Required

- **Response (200 OK — User is member):**
  ```json
  {
    "data": {
      "id": "region_jakarta",
      "name": "KAI Jakarta",
      "slug": "jakarta",
      "description": "Wilayah Jakarta dan sekitarnya",
      "image_url": "https://cdn.example.com/regions/jakarta.jpg",
      "status": "active",
      "member_count": 1245,
      "your_status": "active",
      "your_role": "member",
      "joined_at": "2026-05-20T10:00:00.000Z",
      "created_at": "2026-05-20T10:00:00.000Z"
    }
  }
  ```

- **Response (200 OK — User no region):**
  ```json
  {
    "data": null,
    "message": "You are not a member of any region yet"
  }
  ```

---

### 3. Get Region Detail

Ambil detail lengkap satu region.

- **URL:** `GET /api/v1/mobile/regions/{region_id}`
- **Autentikasi:** Optional

- **Response (200 OK):**
  ```json
  {
    "data": {
      "id": "region_jakarta",
      "name": "KAI Jakarta",
      "slug": "jakarta",
      "description": "Wilayah Jakarta dan sekitarnya",
      "image_url": "https://cdn.example.com/regions/jakarta.jpg",
      "status": "active",
      "member_count": 1245,
      "your_status": "active",
      "your_role": "member",
      "created_at": "2026-05-20T10:00:00.000Z"
    }
  }
  ```

---

### 4. Get Region Members (Public List)

Ambil daftar member region — names only, no sensitive data.

- **URL:** `GET /api/v1/mobile/regions/{region_id}/members`
- **Autentikasi:** Optional
- **Query Parameters:**
  - `limit` (optional, default: 30, max: 50)
  - `offset` (optional, default: 0)

- **Response (200 OK):**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "name": "Andi Pratama",
        "avatar": "https://...",
        "role": "admin" | "member"
      },
      {
        "id": "uuid",
        "name": "Citra Dewi",
        "avatar": "https://...",
        "role": "member"
      }
    ],
    "pagination": {
      "limit": 30,
      "offset": 0,
      "total": 1245
    }
  }
  ```

---

### 5. Request Join Region

Member request bergabung ke region tertentu.

- **URL:** `POST /api/v1/mobile/regions/{region_id}/request`
- **Autentikasi:** Required
- **Authorization:** User tidak boleh sudah punya active membership di region ini

- **Request Body:** (Empty)

- **Response (201 Created — Success):**
  ```json
  {
    "data": {
      "membership_id": "uuid",
      "region_id": "region_jakarta",
      "region_name": "KAI Jakarta",
      "status": "pending_approval",
      "created_at": "2026-05-26T10:00:00.000Z"
    },
    "message": "Join request submitted. Admin akan review dalam 1-2 hari."
  }
  ```

- **Response (409 Conflict — Already member):**
  ```json
  {
    "message": "You are already a member of this region"
  }
  ```

- **Response (409 Conflict — Already pending):**
  ```json
  {
    "message": "You already have a pending request for this region"
  }
  ```

---

### 6. Cancel Join Request

User cancel pending request mereka sendiri.

- **URL:** `POST /api/v1/mobile/regions/{region_id}/request/cancel`
- **Autentikasi:** Required
- **Authorization:** User hanya bisa cancel own request

- **Response (200 OK):**
  ```json
  {
    "data": {
      "region_id": "region_jakarta",
      "status": "cancelled"
    },
    "message": "Join request cancelled"
  }
  ```

- **Response (404 Not Found):**
  ```json
  {
    "message": "No pending request found for this region"
  }
  ```

---

### 7. Get Pending Invitations

Ambil daftar invitations yang menunggu approval user.

- **URL:** `GET /api/v1/mobile/regions/invitations/pending`
- **Autentikasi:** Required
- **Query Parameters:**
  - `limit` (optional, default: 20, max: 50)
  - `offset` (optional, default: 0)

- **Response (200 OK):**
  ```json
  {
    "data": [
      {
        "id": "invite_123",
        "region_id": "region_jakarta",
        "region_name": "KAI Jakarta",
        "region_image": "https://...",
        "invited_by_name": "Admin Jakarta",
        "invited_by_avatar": "https://...",
        "status": "pending",
        "created_at": "2026-05-26T10:00:00.000Z",
        "expires_at": "2026-05-27T10:00:00.000Z",
        "time_left_hours": 23
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 2
    }
  }
  ```

---

### 8. Accept Invitation

User accept invitation bergabung ke region.

- **URL:** `POST /api/v1/mobile/regions/invitations/{invitation_id}/accept`
- **Autentikasi:** Required (email di token harus match invitation.email)
- **Authorization:** Invitation harus belum expired dan status=pending

- **Response (200 OK):**
  ```json
  {
    "data": {
      "membership_id": "uuid",
      "region_id": "region_jakarta",
      "region_name": "KAI Jakarta",
      "status": "active",
      "role": "member",
      "joined_at": "2026-05-26T10:00:00.000Z"
    },
    "message": "Bergabung ke KAI Jakarta ✓"
  }
  ```

- **Response (410 Gone — Invitation expired):**
  ```json
  {
    "message": "Invitation has expired. Admin can send a new one."
  }
  ```

- **Response (409 Conflict — Already accepted):**
  ```json
  {
    "message": "Invitation already accepted"
  }
  ```

---

### 9. Reject Invitation

User reject invitation.

- **URL:** `POST /api/v1/mobile/regions/invitations/{invitation_id}/reject`
- **Autentikasi:** Required
- **Authorization:** Invitation harus status=pending

- **Response (200 OK):**
  ```json
  {
    "data": {
      "invitation_id": "invite_123",
      "status": "rejected"
    },
    "message": "Invitation rejected"
  }
  ```

---

### 10. Leave Region

User keluar dari region (meninggalkan membership aktif).

- **URL:** `POST /api/v1/mobile/regions/{region_id}/leave`
- **Autentikasi:** Required
- **Authorization:** User harus active member di region ini

- **Response (200 OK):**
  ```json
  {
    "data": {
      "region_id": "region_jakarta",
      "region_name": "KAI Jakarta"
    },
    "message": "You have left KAI Jakarta"
  }
  ```

---

## Backoffice Endpoints

### 1. Create Region (Superadmin Only)

- **URL:** `POST /api/v1/web/regions`
- **Autentikasi:** Required
- **Authorization:** Superadmin only

- **Request Body:**
  ```json
  {
    "name": "KAI Bandung",
    "slug": "bandung",
    "description": "Wilayah Bandung dan sekitarnya",
    "image_url": "https://cdn.example.com/regions/bandung.jpg"
  }
  ```

- **Validation:**
  - `name`: Required, string, 3-100 chars, unique
  - `slug`: Required, string, 3-50 chars, unique, lowercase, no spaces
  - `description`: Optional, string, max 500 chars
  - `image_url`: Optional, valid URL

- **Response (201 Created):**
  ```json
  {
    "data": {
      "id": "region_bandung",
      "name": "KAI Bandung",
      "slug": "bandung",
      "description": "Wilayah Bandung dan sekitarnya",
      "image_url": "https://cdn.example.com/regions/bandung.jpg",
      "status": "active",
      "member_count": 0,
      "created_at": "2026-05-26T10:00:00.000Z"
    }
  }
  ```

---

### 2. Update Region (Superadmin Only)

- **URL:** `PUT /api/v1/web/regions/{region_id}`
- **Autentikasi:** Required
- **Authorization:** Superadmin only

- **Request Body:**
  ```json
  {
    "name": "KAI Bandung Raya",
    "description": "...",
    "image_url": "https://..."
  }
  ```

- **Response (200 OK):**
  ```json
  {
    "data": {
      "id": "region_bandung",
      "name": "KAI Bandung Raya",
      "slug": "bandung",
      "updated_at": "2026-05-26T10:05:00.000Z"
    }
  }
  ```

---

### 3. Deactivate/Activate Region (Superadmin Only)

- **URL:** `PATCH /api/v1/web/regions/{region_id}/status`
- **Autentikasi:** Required
- **Authorization:** Superadmin only

- **Request Body:**
  ```json
  {
    "status": "inactive"
  }
  ```

- **Response (200 OK):**
  ```json
  {
    "data": {
      "region_id": "region_bandung",
      "status": "inactive"
    }
  }
  ```

---

### 4. Get All Regions (Backoffice)

- **URL:** `GET /api/v1/web/regions`
- **Autentikasi:** Required
- **Authorization:** Superadmin only (untuk admin region, gunakan endpoint khusus)

- **Query Parameters:**
  - `search` (optional)
  - `status` (optional): `active` | `inactive`
  - `limit`, `offset`

- **Response (200 OK):**
  ```json
  {
    "data": [
      {
        "id": "region_jakarta",
        "name": "KAI Jakarta",
        "slug": "jakarta",
        "status": "active",
        "member_count": 1245,
        "admin_name": "Budi Santoso",
        "admin_email": "budi@example.com",
        "created_at": "2026-05-20T10:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 8
    }
  }
  ```

---

### 5. Get Region Detail (Admin Can View Own Region)

- **URL:** `GET /api/v1/web/regions/{region_id}`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR (Admin Region yang manage region ini)

- **Response (200 OK):**
  ```json
  {
    "data": {
      "id": "region_jakarta",
      "name": "KAI Jakarta",
      "slug": "jakarta",
      "description": "Wilayah Jakarta...",
      "image_url": "https://...",
      "status": "active",
      "member_count": 1245,
      "admin_id": "uuid",
      "admin_name": "Budi Santoso",
      "created_at": "2026-05-20T10:00:00.000Z"
    }
  }
  ```

---

### 6. Assign Admin to Region (Superadmin Only)

- **URL:** `POST /api/v1/web/regions/{region_id}/assign-admin`
- **Autentikasi:** Required
- **Authorization:** Superadmin only

- **Request Body:**
  ```json
  {
    "user_id": "uuid_user" | null,
    "email": "budi@example.com" | null
  }
  ```

- **Response (201 Created — User already registered):**
  ```json
  {
    "data": {
      "region_id": "region_jakarta",
      "user_id": "uuid",
      "user_name": "Budi Santoso",
      "user_email": "budi@example.com",
      "role": "admin",
      "status": "active",
      "assigned_at": "2026-05-26T10:00:00.000Z"
    },
    "message": "Budi assigned as admin for KAI Jakarta"
  }
  ```

- **Response (201 Created — User not registered yet, send invite):**
  ```json
  {
    "data": {
      "region_id": "region_jakarta",
      "email": "newadmin@example.com",
      "status": "invitation_pending",
      "invitation_expires_at": "2026-05-27T10:00:00.000Z"
    },
    "message": "Invitation sent to newadmin@example.com"
  }
  ```

---

### 7. Get Region Members (Admin Can View Own Region Members)

- **URL:** `GET /api/v1/web/regions/{region_id}/members`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR (Admin Region yang manage region ini)

- **Query Parameters:**
  - `status` (optional): `active` | `pending_approval` | `rejected`
  - `role` (optional): `admin` | `member`
  - `search` (optional): Search by name atau email
  - `limit`, `offset`

- **Response (200 OK):**
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "name": "Andi Pratama",
        "email": "andi@example.com",
        "avatar": "https://...",
        "role": "member",
        "status": "active",
        "joined_at": "2026-05-20T10:00:00.000Z"
      },
      {
        "id": "uuid",
        "name": "Citra Dewi",
        "email": "citra@example.com",
        "avatar": "https://...",
        "role": "member",
        "status": "pending_approval",
        "requested_at": "2026-05-25T10:00:00.000Z"
      }
    ],
    "pagination": {
      "limit": 20,
      "offset": 0,
      "total": 1245
    }
  }
  ```

---

### 8. Approve Member Join Request

- **URL:** `POST /api/v1/web/regions/{region_id}/members/{user_id}/approve`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region yang manage region ini

- **Request Body:**
  ```json
  {
    "notes": "Welcome to KAI Jakarta"
  }
  ```

- **Response (200 OK):**
  ```json
  {
    "data": {
      "user_id": "uuid",
      "user_name": "Citra Dewi",
      "user_email": "citra@example.com",
      "region_id": "region_jakarta",
      "status": "active",
      "joined_at": "2026-05-26T10:00:00.000Z"
    },
    "message": "Citra approved and joined KAI Jakarta"
  }
  ```

---

### 9. Reject Member Join Request

- **URL:** `POST /api/v1/web/regions/{region_id}/members/{user_id}/reject`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region

- **Request Body:**
  ```json
  {
    "reason": "Belum memenuhi syarat"
  }
  ```

- **Response (200 OK):**
  ```json
  {
    "data": {
      "user_id": "uuid",
      "user_name": "Citra Dewi",
      "region_id": "region_jakarta",
      "status": "rejected",
      "reason": "Belum memenuhi syarat",
      "rejected_at": "2026-05-26T10:00:00.000Z"
    }
  }
  ```

---

### 10. Invite Member via Email (Admin Can Invite for Own Region)

- **URL:** `POST /api/v1/web/regions/{region_id}/invite`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region yang manage region ini

- **Request Body:**
  ```json
  {
    "emails": ["user1@example.com", "user2@example.com"],
    "role": "member" | "admin"
  }
  ```

- **Response (201 Created):**
  ```json
  {
    "data": {
      "invitations": [
        {
          "id": "invite_123",
          "email": "user1@example.com",
          "status": "pending",
          "expires_at": "2026-05-27T10:00:00.000Z"
        },
        {
          "id": "invite_124",
          "email": "user2@example.com",
          "status": "pending",
          "expires_at": "2026-05-27T10:00:00.000Z"
        }
      ]
    },
    "message": "2 invitations sent"
  }
  ```

---

### 11. Resend Invitation

- **URL:** `POST /api/v1/web/regions/{region_id}/invitations/{invitation_id}/resend`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region

- **Response (200 OK):**
  ```json
  {
    "data": {
      "invitation_id": "invite_123",
      "status": "pending",
      "expires_at": "2026-05-27T12:00:00.000Z"
    },
    "message": "Invitation resent"
  }
  ```

---

### 12. Remove Member from Region

- **URL:** `DELETE /api/v1/web/regions/{region_id}/members/{user_id}`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region

- **Response (200 OK):**
  ```json
  {
    "data": {
      "user_id": "uuid",
      "user_name": "Andi Pratama",
      "region_id": "region_jakarta"
    },
    "message": "Andi removed from KAI Jakarta"
  }
  ```

---

## Error Handling

### Standard Error Response

```json
{
  "message": "Error description here"
}
```

### Validation Error (422)

```json
{
  "message": "Validation failed",
  "errors": {
    "field_name": ["Error message 1"],
    "another_field": ["Error message 2"]
  }
}
```

### Common HTTP Status Codes

- `200 OK` — Success
- `201 Created` — Resource created
- `400 Bad Request` — Bad request
- `401 Unauthorized` — Auth required
- `403 Forbidden` — Permission denied
- `404 Not Found` — Resource not found
- `409 Conflict` — Conflict (e.g., already member)
- `410 Gone` — Invitation expired
- `422 Unprocessable Entity` — Validation failed
- `500 Internal Server Error` — Server error

---

*Dokumen ini menjelaskan API untuk region system. Untuk business logic lihat REGION_SYSTEM.md, untuk database schema lihat REGION_DB_SCHEMA.md*
