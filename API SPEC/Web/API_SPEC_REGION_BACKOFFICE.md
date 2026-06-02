# API Spec — Region Module (Backoffice)

**Status:** Draft v1
**Last Updated:** 2026-06-02
**Base URL Prefix:** `/api/v1/web/regions`

Endpoint backoffice modul Region: pengelolaan region oleh Superadmin & Admin Region — CRUD region, assign admin, kelola member (approve/reject/remove), dan undangan via email. Endpoint mobile (member-facing) ada di `API_SPEC_REGION_MOBILE.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Model Data](#model-data)
3. [Endpoints](#endpoints)
   - [B1. Create Region (Superadmin Only)](#b1-create-region-superadmin-only)
   - [B2. Update Region (Superadmin Only)](#b2-update-region-superadmin-only)
   - [B3. Deactivate/Activate Region (Superadmin Only)](#b3-deactivateactivate-region-superadmin-only)
   - [B4. Get All Regions (Backoffice)](#b4-get-all-regions-backoffice)
   - [B5. Get Region Detail](#b5-get-region-detail)
   - [B6. Assign Admin to Region (Superadmin Only)](#b6-assign-admin-to-region-superadmin-only)
   - [B7. Get Region Members](#b7-get-region-members)
   - [B8. Approve Member Join Request](#b8-approve-member-join-request)
   - [B9. Reject Member Join Request](#b9-reject-member-join-request)
   - [B10. Invite Member via Email](#b10-invite-member-via-email)
   - [B11. Resend Invitation](#b11-resend-invitation)
   - [B12. Remove Member from Region](#b12-remove-member-from-region)
4. [Error Handling](#error-handling)

---

## Informasi Umum

### Headers Global
```
Content-Type: application/json
Accept: application/json
Authorization: Bearer <access_token> (Required)
Accept-Language: <lang_code> (e.g., ko, id, en. Default: ko)
```

### Authentication & Authorization

Semua endpoint require **superadmin** atau **admin region** (role check). Member & guest tidak punya akses.

| Role | Akses Backoffice |
|------|------------------|
| Superadmin | Semua endpoint |
| Admin Region | Hanya region yang dikelolanya (lihat anotasi per-endpoint) |
| Member / Guest | ❌ Tidak ada akses |

> Endpoint bertanda **(Superadmin Only)** tidak dapat diakses Admin Region. Endpoint lain dapat diakses Admin Region **terbatas pada region yang ia kelola**.

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
  "created_at": "2026-05-26T10:00:00.000Z",
  "expires_at": "2026-05-27T10:00:00.000Z",
  "accepted_at": null
}
```

---

## Endpoints

### B1. Create Region (Superadmin Only)

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

### B2. Update Region (Superadmin Only)

- **URL:** `PUT /api/v1/web/regions/{region_id}`
- **Autentikasi:** Required
- **Authorization:** Superadmin only

- **Request Body:**
  ```json
  { "name": "KAI Bandung Raya", "description": "...", "image_url": "https://..." }
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

### B3. Deactivate/Activate Region (Superadmin Only)

- **URL:** `PATCH /api/v1/web/regions/{region_id}/status`
- **Autentikasi:** Required
- **Authorization:** Superadmin only

- **Request Body:**
  ```json
  { "status": "inactive" }
  ```

- **Response (200 OK):**
  ```json
  { "data": { "region_id": "region_bandung", "status": "inactive" } }
  ```

---

### B4. Get All Regions (Backoffice)

- **URL:** `GET /api/v1/web/regions`
- **Autentikasi:** Required
- **Authorization:** Superadmin only (untuk admin region, gunakan B5)

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
    "pagination": { "limit": 20, "offset": 0, "total": 8 }
  }
  ```

---

### B5. Get Region Detail

- **URL:** `GET /api/v1/web/regions/{region_id}`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region yang manage region ini

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

### B6. Assign Admin to Region (Superadmin Only)

- **URL:** `POST /api/v1/web/regions/{region_id}/assign-admin`
- **Autentikasi:** Required
- **Authorization:** Superadmin only

- **Request Body:**
  ```json
  { "user_id": "uuid_user" | null, "email": "budi@example.com" | null }
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

### B7. Get Region Members

- **URL:** `GET /api/v1/web/regions/{region_id}/members`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region yang manage region ini

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
    "pagination": { "limit": 20, "offset": 0, "total": 1245 }
  }
  ```

---

### B8. Approve Member Join Request

- **URL:** `POST /api/v1/web/regions/{region_id}/members/{user_id}/approve`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region yang manage region ini

- **Request Body:**
  ```json
  { "notes": "Welcome to KAI Jakarta" }
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

### B9. Reject Member Join Request

- **URL:** `POST /api/v1/web/regions/{region_id}/members/{user_id}/reject`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region

- **Request Body:**
  ```json
  { "reason": "Belum memenuhi syarat" }
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

### B10. Invite Member via Email

- **URL:** `POST /api/v1/web/regions/{region_id}/invite`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region yang manage region ini

- **Request Body:**
  ```json
  { "emails": ["user1@example.com", "user2@example.com"], "role": "member" | "admin" }
  ```

- **Response (201 Created):**
  ```json
  {
    "data": {
      "invitations": [
        { "id": "invite_123", "email": "user1@example.com", "status": "pending", "expires_at": "2026-05-27T10:00:00.000Z" },
        { "id": "invite_124", "email": "user2@example.com", "status": "pending", "expires_at": "2026-05-27T10:00:00.000Z" }
      ]
    },
    "message": "2 invitations sent"
  }
  ```

---

### B11. Resend Invitation

- **URL:** `POST /api/v1/web/regions/{region_id}/invitations/{invitation_id}/resend`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region

- **Response (200 OK):**
  ```json
  {
    "data": { "invitation_id": "invite_123", "status": "pending", "expires_at": "2026-05-27T12:00:00.000Z" },
    "message": "Invitation resent"
  }
  ```

---

### B12. Remove Member from Region

- **URL:** `DELETE /api/v1/web/regions/{region_id}/members/{user_id}`
- **Autentikasi:** Required
- **Authorization:** Superadmin OR Admin Region

- **Response (200 OK):**
  ```json
  {
    "data": { "user_id": "uuid", "user_name": "Andi Pratama", "region_id": "region_jakarta" },
    "message": "Andi removed from KAI Jakarta"
  }
  ```

---

## Error Handling

### Standard Error Response
```json
{ "message": "Error description here" }
```

### Validation Error (422)
```json
{ "message": "Validation failed", "errors": { "field_name": ["Error message 1"] } }
```

### Common HTTP Status Codes
- `200 OK` — Success
- `201 Created` — Resource created
- `400 Bad Request`
- `401 Unauthorized` — Auth required
- `403 Forbidden` — Permission denied (mis. admin region akses region lain)
- `404 Not Found`
- `409 Conflict`
- `422 Unprocessable Entity` — Validation failed
- `500 Internal Server Error`

---

*Endpoint mobile ada di `API_SPEC_REGION_MOBILE.md`. Business logic: `REGION_SYSTEM_RULES.md`; schema: `REGION_DB_SCHEMA.md`.*
