# API Spec — Bulk Role Assignment (Backoffice) v2

> **Update dari v1:**
> - Tambah field `skipped` dan `skipped_count` di semua bulk result — user yang sudah punya role → skipped, bukan error
> - Tambah `expired_at` support di CSV bulk assign (via extra form field)
> - Fix `error_codes` — lebih konsisten dan lengkap
> - Tambah endpoint `GET /bulk-assign/history` — list semua bulk operation yang pernah dilakukan
> - Tambah endpoint `POST /bulk-assign/preview` — dry-run sebelum eksekusi
> - Klarifikasi limit: 100 user via JSON, 500 via CSV

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/web/role-permission/bulk-assign`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <access_token>` (Required)
- **Authentication**: Required (Superadmin / Admin Region)
- **Authorization**:
  - Superadmin: assign ke siapa saja, any scope
  - Admin Region: hanya ke user di region mereka, scope terbatas ke region sendiri

---

## Model Data

### Bulk Assignment Request

```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3"],
  "role_id": "uuid",
  "scope_type": "region",
  "scope_id": "uuid",
  "expired_at": "2026-12-31T23:59:59.000Z",
  "notes": "Batch admin assignment Q3"
}
```

### Bulk Result Item (Updated)

```json
{
  "user_id": "uuid",
  "user_name": "Andi",
  "email": "andi@example.com",
  "status": "success | error | skipped",
  "role_assignment_id": "uuid",
  "assigned_at": "2026-05-25T10:00:00.000Z",
  "reason": "User already has this role in this scope",
  "error_code": "DUPLICATE_ROLE | USER_NOT_FOUND | UNAUTHORIZED_SCOPE | INVALID_ROLE | USER_SUSPENDED"
}
```

> **Status `skipped` vs `error`:**
> - `skipped`: user valid, tapi sudah punya role yang sama di scope yang sama → tidak ada yang diubah, bukan error
> - `error`: user tidak ditemukan, scope tidak valid, atau unauthorized

### Bulk Operation Result (Updated)

```json
{
  "operation_id": "bulk_op_uuid",
  "total_users": 5,
  "successful": 3,
  "skipped": 1,
  "failed": 1,
  "results": [ ... ]
}
```

---

## Endpoints

### 1. Bulk Assign Role ke Banyak User

Assign satu role ke multiple users sekaligus (scope yang sama untuk semua).

- **URL**: `POST /api/v1/web/role-permission/bulk-assign`
- **Auth**: Required
- **Limit**: Maksimal **100 user** per request

**Request Body**:
```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3", "uuid4"],
  "role_id": "role-admin-uuid",
  "scope_type": "region",
  "scope_id": "region-jakarta-uuid",
  "expired_at": null,
  "notes": "Admin baru region Jakarta batch Juli 2026"
}
```

**Response (201 Created)**:
```json
{
  "data": {
    "operation_id": "bulk_op_20260525_001",
    "total_users": 4,
    "successful": 2,
    "skipped": 1,
    "failed": 1,
    "processed_at": "2026-05-25T10:00:00.000Z",
    "initiated_by": "superadmin_uuid",
    "results": [
      {
        "user_id": "uuid1",
        "user_name": "Andi",
        "email": "andi@example.com",
        "status": "success",
        "role_assignment_id": "ur_new_001",
        "assigned_at": "2026-05-25T10:00:00.000Z"
      },
      {
        "user_id": "uuid2",
        "user_name": "Budi",
        "email": "budi@example.com",
        "status": "skipped",
        "reason": "User already has this role in this scope",
        "error_code": "DUPLICATE_ROLE"
      },
      {
        "user_id": "uuid3",
        "user_name": "Citra",
        "email": "citra@example.com",
        "status": "success",
        "role_assignment_id": "ur_new_002",
        "assigned_at": "2026-05-25T10:00:00.000Z"
      },
      {
        "user_id": "uuid4",
        "user_name": null,
        "email": null,
        "status": "error",
        "reason": "User not found",
        "error_code": "USER_NOT_FOUND"
      }
    ]
  },
  "message": "2 berhasil, 1 di-skip (duplikat), 1 error dari 4 user"
}
```

**Error Codes:**

| Code | Keterangan |
|---|---|
| `DUPLICATE_ROLE` | User sudah punya role ini di scope yang sama → skipped |
| `USER_NOT_FOUND` | user_id tidak ditemukan di DB |
| `USER_SUSPENDED` | User di-suspend, tidak bisa di-assign role |
| `UNAUTHORIZED_SCOPE` | Admin region coba assign ke region lain |
| `INVALID_ROLE` | role_id tidak valid atau admin coba assign superadmin |
| `INVALID_SCOPE` | scope_id tidak valid (region/community tidak ada) |

**Response (403 — Admin region scope violation)**:
```json
{
  "message": "Admin hanya bisa assign role di region mereka sendiri",
  "data": {
    "admin_region": "Jakarta",
    "requested_scope": "Surabaya"
  }
}
```

---

### 2. Preview Bulk Assign (Dry Run)

**Endpoint baru** — simulasikan operasi tanpa benar-benar mengeksekusi. Return prediksi hasilnya: berapa yang akan berhasil, skip, error.

- **URL**: `POST /api/v1/web/role-permission/bulk-assign/preview`
- **Auth**: Required

**Request Body**: Sama persis dengan bulk assign biasa.

**Response (200 OK)**:
```json
{
  "data": {
    "is_preview": true,
    "total_users": 4,
    "estimated_successful": 2,
    "estimated_skipped": 1,
    "estimated_failed": 1,
    "breakdown": [
      {
        "user_id": "uuid1",
        "user_name": "Andi",
        "email": "andi@example.com",
        "predicted_status": "success"
      },
      {
        "user_id": "uuid2",
        "user_name": "Budi",
        "email": "budi@example.com",
        "predicted_status": "skipped",
        "reason": "Already has role in this scope",
        "error_code": "DUPLICATE_ROLE"
      },
      {
        "user_id": "uuid3",
        "user_name": "Citra",
        "email": "citra@example.com",
        "predicted_status": "success"
      },
      {
        "user_id": "uuid4",
        "user_name": null,
        "email": null,
        "predicted_status": "error",
        "reason": "User not found",
        "error_code": "USER_NOT_FOUND"
      }
    ]
  },
  "message": "Preview only — tidak ada yang diubah"
}
```

---

### 3. Get Bulk Operation Status

Check status dari bulk operation yang sedang atau sudah selesai.

- **URL**: `GET /api/v1/web/role-permission/bulk-assign/{operation_id}`
- **Auth**: Required

**Response (200 OK)**:
```json
{
  "data": {
    "operation_id": "bulk_op_20260525_001",
    "type": "bulk_assign",
    "status": "completed",
    "total_users": 4,
    "successful": 2,
    "skipped": 1,
    "failed": 1,
    "role_name": "admin",
    "scope_type": "region",
    "scope_name": "Jakarta",
    "started_at": "2026-05-25T10:00:00.000Z",
    "completed_at": "2026-05-25T10:00:03.000Z",
    "initiated_by": "superadmin_uuid",
    "initiated_by_name": "Super Admin",
    "results": [ ... ]
  }
}
```

---

### 4. Bulk Assign dari CSV

Upload CSV dengan daftar user, assign role yang sama ke semua.

- **URL**: `POST /api/v1/web/role-permission/bulk-assign/from-file`
- **Auth**: Required
- **Content-Type**: `multipart/form-data`
- **Limit**: Maksimal **500 baris** per file

**Request form fields**:

| Field | Required | Keterangan |
|---|---|---|
| `file` | ✅ | File .csv |
| `role_id` | ✅ | UUID role yang akan di-assign |
| `scope_type` | ✅ | `global`, `region`, atau `community` |
| `scope_id` | Conditional | Wajib jika scope_type bukan global |
| `expired_at` | ❌ | ISO 8601 datetime, opsional |
| `notes` | ❌ | Catatan operasi |

**Format CSV** (kolom `user_id` atau `email` wajib ada salah satu):
```csv
user_id,email,name
uuid1,andi@example.com,Andi Admin
uuid2,budi@example.com,Budi Admin
,citra@example.com,Citra (pakai email karena tidak tahu UUID)
```

> Jika `user_id` dan `email` keduanya ada, `user_id` yang dipakai.

**Response (201 Created)**:
```json
{
  "data": {
    "operation_id": "bulk_op_csv_001",
    "total_users": 3,
    "successful": 3,
    "skipped": 0,
    "failed": 0,
    "file_name": "admins_jakarta.csv",
    "expired_at_applied": "2026-12-31T23:59:59.000Z",
    "processed_at": "2026-05-25T10:05:00.000Z",
    "results": [ ... ]
  },
  "message": "CSV berhasil diproses, 3 user di-assign"
}
```

**Response (400 — CSV parse error)**:
```json
{
  "message": "Format CSV tidak valid",
  "errors": {
    "header": "Kolom user_id atau email wajib ada",
    "line_3": "Format UUID tidak valid di kolom user_id",
    "line_7": "Email kosong"
  }
}
```

> Jika ada error di CSV format, **seluruh file ditolak** — tidak ada yang diproses. Admin harus fix CSV dulu.

---

### 5. Bulk Revoke Role dari Banyak User

Remove satu role dari multiple users sekaligus.

- **URL**: `DELETE /api/v1/web/role-permission/bulk-assign`
- **Auth**: Required

**Request Body**:
```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3"],
  "role_id": "role-admin-uuid",
  "scope_type": "region",
  "scope_id": "region-jakarta-uuid",
  "reason": "Reorganisasi admin regional Q3 2026"
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "operation_id": "bulk_revoke_001",
    "total_users": 3,
    "successful": 2,
    "skipped": 1,
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
        "status": "skipped",
        "reason": "User does not have this role assignment",
        "error_code": "ROLE_NOT_ASSIGNED"
      },
      {
        "user_id": "uuid3",
        "user_name": "Citra",
        "status": "success",
        "revoked_at": "2026-05-25T10:10:00.000Z"
      }
    ]
  },
  "message": "2 role berhasil di-revoke, 1 di-skip"
}
```

---

### 6. Bulk Update Expiry

Update tanggal expired untuk multiple role assignments sekaligus.

- **URL**: `PATCH /api/v1/web/role-permission/bulk-assign/expiry`
- **Auth**: Required

**Request Body**:
```json
{
  "assignment_ids": ["ur_uuid1", "ur_uuid2", "ur_uuid3"],
  "expired_at": "2026-12-31T23:59:59.000Z",
  "notes": "Perpanjang kontrak sampai akhir tahun"
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "total": 3,
    "successful": 3,
    "failed": 0,
    "new_expired_at": "2026-12-31T23:59:59.000Z",
    "updated_at": "2026-05-25T10:15:00.000Z"
  },
  "message": "Expiry 3 role assignment berhasil diperbarui"
}
```

---

### 7. Download Laporan Bulk Operation

Download hasil operasi bulk sebagai file CSV.

- **URL**: `GET /api/v1/web/role-permission/bulk-assign/{operation_id}/download`
- **Auth**: Required

**Response**: File download
```
Content-Type: text/csv
Content-Disposition: attachment; filename=bulk_op_20260525_001_result.csv

user_id,user_name,email,status,role_name,scope_name,assigned_at,error_code
uuid1,Andi,andi@email.com,success,admin,Jakarta,2026-05-25T10:00:00Z,
uuid2,Budi,budi@email.com,skipped,admin,Jakarta,,DUPLICATE_ROLE
```

---

### 8. List Bulk Operation History

**Endpoint baru** — riwayat semua bulk operation yang pernah dilakukan.

- **URL**: `GET /api/v1/web/role-permission/bulk-assign/history`
- **Auth**: Required
- **Query Parameters**: `limit` (default 20), `offset`, `type` (`bulk_assign` | `bulk_revoke` | `from_file`)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "operation_id": "bulk_op_20260525_001",
      "type": "bulk_assign",
      "role_name": "admin",
      "scope_type": "region",
      "scope_name": "Jakarta",
      "total_users": 4,
      "successful": 2,
      "skipped": 1,
      "failed": 1,
      "initiated_by_name": "Super Admin",
      "processed_at": "2026-05-25T10:00:00.000Z"
    },
    {
      "operation_id": "bulk_op_csv_001",
      "type": "from_file",
      "file_name": "admins_jakarta.csv",
      "role_name": "admin",
      "scope_type": "region",
      "scope_name": "Jakarta",
      "total_users": 3,
      "successful": 3,
      "skipped": 0,
      "failed": 0,
      "initiated_by_name": "Super Admin",
      "processed_at": "2026-05-25T10:05:00.000Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 15 }
}
```

---

## Error Codes Lengkap

| Code | HTTP | Keterangan |
|---|---|---|
| `DUPLICATE_ROLE` | - | User sudah punya role di scope yang sama (→ skipped) |
| `ROLE_NOT_ASSIGNED` | - | User tidak punya role ini (saat revoke → skipped) |
| `USER_NOT_FOUND` | - | user_id tidak ditemukan |
| `USER_SUSPENDED` | - | User sedang di-suspend |
| `UNAUTHORIZED_SCOPE` | 403 | Admin region coba assign ke scope lain |
| `INVALID_ROLE` | 403 | Role tidak valid atau tidak boleh di-assign |
| `INVALID_SCOPE` | 422 | scope_id tidak ditemukan |
| `LIMIT_EXCEEDED` | 422 | Melebihi batas 100 (JSON) atau 500 (CSV) |
| `CSV_PARSE_ERROR` | 400 | Format CSV tidak valid |
| `CSV_MISSING_COLUMN` | 400 | Kolom wajib tidak ada di CSV |

---

## Catatan Implementasi

**Partial success adalah default** — satu error tidak menghentikan seluruh operasi. Setiap item diproses independen dan hasilnya dilaporkan per-item.

**Urutan proses CSV**: validasi format → preview internal → eksekusi → return hasil. Jika validasi gagal, tidak ada yang dieksekusi.

**Idempotency**: Mengirim request yang sama dua kali → request kedua semua item akan berstatus `skipped` (bukan error). Aman untuk retry.

**Audit log**: Setiap bulk operation menghasilkan satu log entry di level operasi, plus log entry individual per successful assignment/revoke.
