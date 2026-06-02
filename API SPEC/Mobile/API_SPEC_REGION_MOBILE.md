# API Spec — Region Module (Mobile Client)

**Status:** Draft v1
**Last Updated:** 2026-06-02
**Base URL Prefix:** `/api/v1/mobile/regions`

Endpoint mobile modul Region: browse region, lihat region sendiri, request join, kelola undangan, dan keluar region. Endpoint backoffice (superadmin/admin region) ada di `API_SPEC_REGION_BACKOFFICE.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Model Data](#model-data)
3. [Endpoints](#endpoints)
   - [1. Get All Regions](#1-get-all-regions)
   - [2. Get My Region](#2-get-my-region)
   - [3. Get Region Detail](#3-get-region-detail)
   - [4. Get Region Members (Public List)](#4-get-region-members-public-list)
   - [5. Request Join Region](#5-request-join-region)
   - [6. Cancel Join Request](#6-cancel-join-request)
   - [7. Get Pending Invitations](#7-get-pending-invitations)
   - [8. Accept Invitation](#8-accept-invitation)
   - [9. Reject Invitation](#9-reject-invitation)
   - [10. Leave Region](#10-leave-region)
4. [Error Handling](#error-handling)

---

## Informasi Umum

### Headers Global
```
Content-Type: application/json
Accept: application/json
Authorization: Bearer <access_token> (Required, kecuali untuk public endpoints)
Accept-Language: <lang_code> (e.g., ko, id, en. Default: ko)
```

### Authentication & Authorization

- Sebagian besar endpoint require authenticated user (Bearer token).
- Beberapa endpoint bersifat **Optional auth** (browse) — hasil lebih lengkap (`your_status`/`your_role`) jika authenticated.

| Role | Akses Mobile |
|------|--------------|
| Superadmin | Semua endpoint mobile |
| Admin Region | Semua endpoint mobile |
| Member | Browse, request, accept/reject invite, leave |
| Guest | Browse region (endpoint dengan optional auth) |

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

## Endpoints

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
    "pagination": { "limit": 50, "offset": 0, "total": 8 }
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
  { "data": null, "message": "You are not a member of any region yet" }
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
      { "id": "uuid", "name": "Andi Pratama", "avatar": "https://...", "role": "admin" },
      { "id": "uuid", "name": "Citra Dewi", "avatar": "https://...", "role": "member" }
    ],
    "pagination": { "limit": 30, "offset": 0, "total": 1245 }
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
  { "message": "You are already a member of this region" }
  ```

- **Response (409 Conflict — Already pending):**
  ```json
  { "message": "You already have a pending request for this region" }
  ```

---

### 6. Cancel Join Request

User cancel pending request mereka sendiri.

- **URL:** `POST /api/v1/mobile/regions/{region_id}/request/cancel`
- **Autentikasi:** Required
- **Authorization:** User hanya bisa cancel own request

- **Response (200 OK):**
  ```json
  { "data": { "region_id": "region_jakarta", "status": "cancelled" }, "message": "Join request cancelled" }
  ```

- **Response (404 Not Found):**
  ```json
  { "message": "No pending request found for this region" }
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
    "pagination": { "limit": 20, "offset": 0, "total": 2 }
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
  { "message": "Invitation has expired. Admin can send a new one." }
  ```

- **Response (409 Conflict — Already accepted):**
  ```json
  { "message": "Invitation already accepted" }
  ```

---

### 9. Reject Invitation

User reject invitation.

- **URL:** `POST /api/v1/mobile/regions/invitations/{invitation_id}/reject`
- **Autentikasi:** Required
- **Authorization:** Invitation harus status=pending

- **Response (200 OK):**
  ```json
  { "data": { "invitation_id": "invite_123", "status": "rejected" }, "message": "Invitation rejected" }
  ```

---

### 10. Leave Region

User keluar dari region (meninggalkan membership aktif).

- **URL:** `POST /api/v1/mobile/regions/{region_id}/leave`
- **Autentikasi:** Required
- **Authorization:** User harus active member di region ini

- **Response (200 OK):**
  ```json
  { "data": { "region_id": "region_jakarta", "region_name": "KAI Jakarta" }, "message": "You have left KAI Jakarta" }
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
- `403 Forbidden` — Permission denied
- `404 Not Found`
- `409 Conflict` — e.g., already member
- `410 Gone` — Invitation expired
- `422 Unprocessable Entity` — Validation failed
- `500 Internal Server Error`

---

*Endpoint backoffice ada di `API_SPEC_REGION_BACKOFFICE.md`. Business logic: `REGION_SYSTEM_RULES.md`; schema: `REGION_DB_SCHEMA.md`.*
