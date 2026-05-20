# Dokumentasi API Spec - Modul Profile & Membership (Mobile Client)

Dokumentasi ini dibuat untuk kebutuhan tim Backend agar skema request/response API sesuai dengan implementasi Clean Architecture pada Flutter Mobile Client.

## Informasi Umum

- **Base URL Prefix**: `/api/v1` (Seluruh endpoint di bawah ini menggunakan prefix `/api/v1/mobile/profile`)
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Accept-Language: <lang_code>` (Mengirimkan kode bahasa aktif, contoh: `ko`, `id`, `en`. Default: `ko`)
  - `X-Locale: <lang_code>` (Mengirimkan kode bahasa aktif, contoh: `ko`, `id`, `en`. Default: `ko`)
  - `Authorization: Bearer <access_token>` (Diperlukan untuk semua endpoint di modul ini)

---

## Model Data Utama

### 1. User Profile Object (Common Response Schema)
Setiap kali endpoint mengembalikan data Profile lengkap, strukturnya harus berupa format berikut di dalam field `data`:

```json
{
  "id": "string",
  "name": "string",
  "email": "string",
  "username": "string",
  "avatar": "string (nullable / URL)",
  "phone": "string (nullable)",
  "birth_date": "string (nullable, ISO 8601 format: YYYY-MM-DD)",
  "gender": "string (nullable, enum: 'male', 'female', 'other')",
  "bio": "string (nullable, max 500 char)",
  "marital_status": "string (nullable, enum: 'single', 'married', 'divorced', 'widowed')",
  "occupation": "string (nullable)",
  "interests": ["string"] (array of interest tags, nullable),
  "address": {
    "street": "string (nullable)",
    "city": "string (nullable)",
    "province": "string (nullable)",
    "postal_code": "string (nullable)",
    "country": "string (nullable)"
  },
  "plan": "string (e.g. 'standard', 'pro')",
  "roles": ["string"],
  "permissions": ["string"],
  "is_verified": "boolean",
  "created_at": "string (ISO 8601 UTC format)",
  "updated_at": "string (ISO 8601 UTC format)",
  "last_login": "string (nullable, ISO 8601 UTC format)"
}
```

### 2. Region Membership Object
Struktur untuk region yang diikuti user:

```json
{
  "id": "string",
  "name": "string (KAI Pusat / KAI Wilayah)",
  "code": "string (nullable, e.g. 'central', 'jakarta', 'bandung')",
  "joined_at": "string (ISO 8601 UTC format)"
}
```

### 3. Community Membership Object
Struktur untuk komunitas yang diikuti user:

```json
{
  "id": "string",
  "name": "string",
  "description": "string (nullable)",
  "avatar": "string (nullable / URL)",
  "community_role": "string (enum: 'member', 'moderator', 'leader')",
  "joined_at": "string (ISO 8601 UTC format)"
}
```

### 4. Subscription Object
Struktur detail subscription & plan user:

```json
{
  "plan": "string (enum: 'standard', 'pro')",
  "plan_id": "string",
  "status": "string (enum: 'active', 'expired', 'cancelled')",
  "start_date": "string (ISO 8601 UTC format)",
  "expiry_date": "string (nullable, ISO 8601 UTC format)",
  "auto_renew": "boolean",
  "price": "number (harga plan per bulan/tahun)",
  "currency": "string (e.g. 'IDR', 'USD')",
  "billing_cycle": "string (enum: 'monthly', 'yearly')",
  "next_billing_date": "string (nullable, ISO 8601 UTC format)",
  "remaining_days": "integer (sisa hari sebelum expired)"
}
```

### 5. Error Responses
Frontend menangani 2 skema error utama dari backend:

#### A. Standard Message Error (HTTP 4xx / 5xx)
```json
{
  "message": "Pesan error deskriptif di sini"
}
```

#### B. Validation Error (HTTP 422 Unprocessable Entity)
```json
{
  "message": "Data input tidak valid",
  "errors": {
    "email": ["Format email tidak valid."],
    "birth_date": ["Tanggal lahir tidak boleh di masa depan."],
    "phone": ["Format nomor telepon tidak valid."]
  }
}
```

---

## Daftar Endpoint

### 1. Get Profile (Fetch Full Profile Data)
Mengambil informasi profile lengkap pengguna yang sedang login beserta memberships dan subscription.

- **URL**: `GET /api/v1/mobile/profile/me`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**: None (Menggunakan Authorization Header)
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "usr_90210",
      "name": "Minji Park",
      "email": "user@example.com",
      "username": "minjipark",
      "avatar": "https://example.com/avatar.jpg",
      "phone": "08123456789",
      "birth_date": "1995-05-20",
      "gender": "female",
      "bio": "Seorang penggemar K-culture dan komunitas lokal.",
      "marital_status": "single",
      "occupation": "Marketing Manager",
      "interests": ["k-pop", "korean-food", "sports", "nature"],
      "address": {
        "street": "Jl. Sudirman No. 123",
        "city": "Jakarta",
        "province": "DKI Jakarta",
        "postal_code": "12190",
        "country": "Indonesia"
      },
      "plan": "pro",
      "roles": ["member"],
      "permissions": ["post_news", "create_community", "moderate_community"],
      "is_verified": true,
      "created_at": "2025-01-15T10:30:00.000Z",
      "updated_at": "2026-05-20T14:20:00.000Z",
      "last_login": "2026-05-20T08:45:00.000Z",
      "regions": [
        {
          "id": "reg_001",
          "name": "KAI Pusat",
          "code": "central",
          "joined_at": "2025-01-15T10:30:00.000Z"
        },
        {
          "id": "reg_002",
          "name": "KAI Jakarta",
          "code": "jakarta",
          "joined_at": "2025-06-10T14:20:00.000Z"
        }
      ],
      "communities": [
        {
          "id": "com_001",
          "name": "K-Pop Lovers",
          "description": "Komunitas penggemar musik Korea",
          "avatar": "https://example.com/kpop-avatar.jpg",
          "community_role": "member",
          "joined_at": "2025-02-01T09:00:00.000Z"
        },
        {
          "id": "com_002",
          "name": "Futsal Reguler",
          "description": "Komunitas olahraga futsal",
          "avatar": "https://example.com/futsal-avatar.jpg",
          "community_role": "leader",
          "joined_at": "2025-03-10T15:30:00.000Z"
        }
      ],
      "subscription": {
        "plan": "pro",
        "plan_id": "plan_pro_001",
        "status": "active",
        "start_date": "2026-04-20T00:00:00.000Z",
        "expiry_date": "2026-10-20T23:59:59.000Z",
        "auto_renew": true,
        "price": 99000,
        "currency": "IDR",
        "billing_cycle": "monthly",
        "next_billing_date": "2026-06-20T00:00:00.000Z",
        "remaining_days": 152
      }
    }
  }
  ```

---

### 2. Update Profile (Edit Personal Information)
Mengupdate informasi profile personal user (name, bio, address, personal details).

- **URL**: `PATCH /api/v1/mobile/profile/me`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "name": "Minji Park Updated",
    "phone": "08129876543",
    "birth_date": "1995-05-20",
    "gender": "female",
    "bio": "Seorang penggemar K-culture, community builder, dan tech enthusiast.",
    "marital_status": "married",
    "occupation": "Senior Marketing Manager",
    "interests": ["k-pop", "korean-food", "sports", "nature", "technology"],
    "address": {
      "street": "Jl. Sudirman No. 456",
      "city": "Jakarta",
      "province": "DKI Jakarta",
      "postal_code": "12190",
      "country": "Indonesia"
    }
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Profil berhasil diperbarui",
    "data": {
      "id": "usr_90210",
      "name": "Minji Park Updated",
      "email": "user@example.com",
      "username": "minjipark",
      "avatar": "https://example.com/avatar.jpg",
      "phone": "08129876543",
      "birth_date": "1995-05-20",
      "gender": "female",
      "bio": "Seorang penggemar K-culture, community builder, dan tech enthusiast.",
      "marital_status": "married",
      "occupation": "Senior Marketing Manager",
      "interests": ["k-pop", "korean-food", "sports", "nature", "technology"],
      "address": {
        "street": "Jl. Sudirman No. 456",
        "city": "Jakarta",
        "province": "DKI Jakarta",
        "postal_code": "12190",
        "country": "Indonesia"
      },
      "plan": "pro",
      "roles": ["member"],
      "permissions": ["post_news", "create_community", "moderate_community"],
      "is_verified": true,
      "created_at": "2025-01-15T10:30:00.000Z",
      "updated_at": "2026-05-20T16:45:00.000Z",
      "last_login": "2026-05-20T08:45:00.000Z"
    }
  }
  ```

---

### 3. Upload Avatar
Mengupload atau mengganti avatar profile user. File harus berupa image (jpg, png, webp).

- **URL**: `POST /api/v1/mobile/profile/avatar`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Content-Type**: `multipart/form-data`
- **Request Body**:
  ```
  file: <binary_image_file> (max 5MB, format: jpg, png, webp)
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Avatar berhasil diupload",
    "data": {
      "avatar": "https://example.com/avatars/usr_90210_1716277200.jpg",
      "avatar_thumbnail": "https://example.com/avatars/usr_90210_1716277200_thumb.jpg"
    }
  }
  ```
- **Response (Error 400 - Invalid File)**:
  ```json
  {
    "message": "File tidak valid",
    "errors": {
      "file": ["File harus berupa gambar (jpg, png, webp)", "Ukuran file maksimal 5MB"]
    }
  }
  ```

---

### 4. Delete Avatar
Menghapus avatar profile user (set ke default/null).

- **URL**: `DELETE /api/v1/mobile/profile/avatar`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**: None
- **Response (Success 200)**:
  ```json
  {
    "message": "Avatar berhasil dihapus",
    "data": {
      "avatar": null
    }
  }
  ```

---

### 5. Get Region Memberships
Mengambil daftar semua region yang diikuti user.

- **URL**: `GET /api/v1/mobile/profile/memberships/regions`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Query Parameters** (optional):
  - `page`: integer (default: 1)
  - `limit`: integer (default: 10, max: 50)
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "reg_001",
        "name": "KAI Pusat",
        "code": "central",
        "joined_at": "2025-01-15T10:30:00.000Z"
      },
      {
        "id": "reg_002",
        "name": "KAI Jakarta",
        "code": "jakarta",
        "joined_at": "2025-06-10T14:20:00.000Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 1,
      "total_items": 2,
      "items_per_page": 10
    }
  }
  ```

---

### 6. Get Community Memberships
Mengambil daftar semua komunitas yang diikuti user beserta community role-nya.

- **URL**: `GET /api/v1/mobile/profile/memberships/communities`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Query Parameters** (optional):
  - `page`: integer (default: 1)
  - `limit`: integer (default: 10, max: 50)
  - `role`: string (filter by role: 'member', 'moderator', 'leader')
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "com_001",
        "name": "K-Pop Lovers",
        "description": "Komunitas penggemar musik Korea",
        "avatar": "https://example.com/kpop-avatar.jpg",
        "community_role": "member",
        "joined_at": "2025-02-01T09:00:00.000Z"
      },
      {
        "id": "com_002",
        "name": "Futsal Reguler",
        "description": "Komunitas olahraga futsal",
        "avatar": "https://example.com/futsal-avatar.jpg",
        "community_role": "leader",
        "joined_at": "2025-03-10T15:30:00.000Z"
      }
    ],
    "pagination": {
      "current_page": 1,
      "total_pages": 1,
      "total_items": 2,
      "items_per_page": 10
    }
  }
  ```

---

### 7. Get Subscription Details
Mengambil informasi subscription & plan user secara detail.

- **URL**: `GET /api/v1/mobile/profile/subscription`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**: None
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "plan": "pro",
      "plan_id": "plan_pro_001",
      "status": "active",
      "start_date": "2026-04-20T00:00:00.000Z",
      "expiry_date": "2026-10-20T23:59:59.000Z",
      "auto_renew": true,
      "price": 99000,
      "currency": "IDR",
      "billing_cycle": "monthly",
      "next_billing_date": "2026-06-20T00:00:00.000Z",
      "remaining_days": 152,
      "benefits": [
        "Post unlimited news",
        "Create communities",
        "Moderate communities",
        "Access premium features",
        "Priority support"
      ]
    }
  }
  ```

---

### 8. Change Email (Request)
Melakukan request untuk mengubah email. System akan mengirimkan OTP verifikasi ke email lama dan email baru.

- **URL**: `POST /api/v1/mobile/profile/email/change-request`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "new_email": "newemail@example.com",
    "password": "currentpassword123"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Verifikasi OTP telah dikirim ke email lama dan email baru Anda",
    "data": {
      "temp_token": "temp_email_change_token_xyz",
      "expires_in": 3600
    }
  }
  ```
- **Response (Error 401 - Wrong Password)**:
  ```json
  {
    "message": "Password salah"
  }
  ```

---

### 9. Verify Email Change
Memverifikasi perubahan email dengan OTP yang diterima. User harus memberikan OTP dari email lama dan email baru.

- **URL**: `POST /api/v1/mobile/profile/email/verify-change`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "temp_token": "temp_email_change_token_xyz",
    "otp_old_email": "123456",
    "otp_new_email": "654321"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Email berhasil diperbarui",
    "data": {
      "email": "newemail@example.com"
    }
  }
  ```

---

### 10. Delete Account (Request)
Melakukan request untuk menghapus akun. System akan mengirimkan konfirmasi via email dan OTP.

> [!WARNING]
> **Catatan**: Penghapusan akun bersifat permanent. Semua data user akan dihapus dalam 30 hari. Selama periode 30 hari, user masih bisa cancel penghapusan dengan mengirimkan request cancel-delete dengan OTP yang sama.

- **URL**: `POST /api/v1/mobile/profile/account/delete-request`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "password": "currentpassword123",
    "reason": "string (nullable, alasan menghapus akun)"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Permintaan penghapusan akun telah dikirim. Cek email Anda untuk konfirmasi.",
    "data": {
      "deletion_token": "delete_token_xyz",
      "deletion_scheduled_date": "2026-06-20T00:00:00.000Z",
      "expires_in": 2592000
    }
  }
  ```

---

### 11. Confirm Account Deletion
Mengkonfirmasi penghapusan akun dengan OTP yang diterima via email.

- **URL**: `POST /api/v1/mobile/profile/account/confirm-delete`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "deletion_token": "delete_token_xyz",
    "otp": "123456"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Akun Anda telah dihapus dan akan dihapus sepenuhnya dalam 30 hari",
    "data": {
      "deletion_scheduled_date": "2026-06-20T00:00:00.000Z"
    }
  }
  ```

---

### 12. Cancel Account Deletion
Membatalkan permintaan penghapusan akun jika masih dalam periode 30 hari.

- **URL**: `POST /api/v1/mobile/profile/account/cancel-delete`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "deletion_token": "delete_token_xyz",
    "otp": "123456"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Penghapusan akun telah dibatalkan",
    "data": {
      "status": "active"
    }
  }
  ```

---

## Validasi Data & Constraints

### Personal Information Validation

| Field | Rules |
|-------|-------|
| `name` | Required, min 3 char, max 100 char |
| `email` | Required, valid email format, unique |
| `phone` | Optional, valid phone format |
| `birth_date` | Optional, must be valid date, not in future, min age 13 |
| `gender` | Optional, enum: 'male', 'female', 'other' |
| `bio` | Optional, max 500 char |
| `marital_status` | Optional, enum: 'single', 'married', 'divorced', 'widowed' |
| `occupation` | Optional, max 100 char |
| `interests` | Optional, array of strings, max 10 items, each max 50 char |
| `address.street` | Optional, max 200 char |
| `address.city` | Optional, max 100 char |
| `address.province` | Optional, max 100 char |
| `address.postal_code` | Optional, valid postal code format |
| `address.country` | Optional, valid country code or name |

---

## Status Code Reference

| Code | Meaning |
|------|---------|
| `200` | Success - Request berhasil |
| `201` | Created - Resource berhasil dibuat |
| `400` | Bad Request - Input tidak valid |
| `401` | Unauthorized - Token invalid/expired atau password salah |
| `403` | Forbidden - User tidak memiliki akses |
| `404` | Not Found - Resource tidak ditemukan |
| `409` | Conflict - Email sudah terdaftar/duplikasi |
| `422` | Unprocessable Entity - Validation error |
| `500` | Internal Server Error - Error di backend |

---

## Notes & Best Practices

1. **Token Refresh**: Jika user mendapat 401 pada endpoint apapun (kecuali login/register), mobile client harus refresh token menggunakan endpoint dari AUTH spec, lalu retry request.

2. **Rate Limiting**: Backend disarankan implementasi rate limiting untuk endpoint-endpoint sensitif seperti:
   - Email change request (max 3x per 24 jam)
   - Account deletion request (max 1x per 7 hari)

3. **Soft Delete**: Untuk account deletion, implementasi soft delete selama 30 hari grace period sebelum permanent delete.

4. **Avatar CDN**: Avatar harus di-serve dari CDN untuk performa optimal. Include thumbnail version untuk list views.

5. **Membership Pagination**: Untuk komunitas/region dengan banyak anggota, gunakan pagination untuk menghindari response yang terlalu besar.

6. **Timestamp Format**: Semua timestamp harus dalam format ISO 8601 UTC (contoh: `2026-05-20T14:20:00.000Z`).

7. **Language Support**: Gunakan `Accept-Language` header untuk mengirimkan error messages & benefits dalam bahasa user.
