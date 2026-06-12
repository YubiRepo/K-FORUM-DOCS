# API Spec — System Settings (Backoffice)

**Status:** Draft v1
**Last Updated:** 2026-06-12
**Base URL Prefix:** `/api/v1/web/system-settings` & `/api/v1/web/legal-documents`

Referensi bisnis: `SYSTEM_SETTINGS_RULES.md` · Schema: `SYSTEM_SETTINGS_DB_SCHEMA.md`

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Model Data](#model-data)
3. [Settings Endpoints](#settings-endpoints)
4. [Legal Documents Endpoints](#legal-documents-endpoints)
5. [Admin Flow](#admin-flow)
6. [Important Notes](#important-notes)
7. [Error Handling](#error-handling)

---

## Informasi Umum

### Headers Global

```
Content-Type: application/json
Accept: application/json
Authorization: Bearer <access_token> (Required)
```

### Authorization

| Endpoint group | Usergod | Superadmin | Admin Region |
|---|---|---|---|
| Read settings | ✅ | ✅ | ❌ |
| Update settings `editable_by = usergod` | ✅ | ❌ (403) | ❌ |
| Update settings `editable_by = superadmin` | ✅ | ✅ | ❌ |
| Legal documents (full) | ✅ | ✅ | ❌ |

Permission yang dicek: `manage_system_settings` dan `manage_legal_documents`. Di atas permission, middleware juga mengecek kolom `editable_by` per setting key.

---

## Model Data

### Setting Object

```json
{
  "group_key": "mobile_app",
  "setting_key": "min_version_android",
  "value": "1.2.0",
  "value_type": "string",
  "is_public": true,
  "is_sensitive": false,
  "editable_by": "superadmin",
  "description": "Versi minimum Android (semver)",
  "updated_by_name": "Super Admin",
  "updated_at": "2026-06-10T08:00:00.000Z"
}
```

> Untuk setting `is_sensitive = true`, field `value` dikembalikan **masked**: `"ap***ey"`. Nilai asli tidak pernah keluar lewat GET.

### Legal Document Object

```json
{
  "id": "uuid",
  "doc_type": "terms",
  "title": "Syarat & Ketentuan",
  "published_version": {
    "id": "uuid",
    "version": "2.1.0",
    "effective_date": "2026-05-15",
    "require_reacceptance": true,
    "published_at": "2026-05-01T10:00:00.000Z",
    "acceptance_count": 12450
  },
  "draft_version": {
    "id": "uuid",
    "version": "2.2.0",
    "updated_at": "2026-06-10T08:00:00.000Z"
  },
  "total_versions": 5
}
```

### Legal Document Version Object

```json
{
  "id": "uuid",
  "document_id": "uuid",
  "doc_type": "terms",
  "version": "2.1.0",
  "content": "# Syarat dan Ketentuan...\n\n(markdown)",
  "status": "published",
  "effective_date": "2026-05-15",
  "require_reacceptance": true,
  "acceptance_count": 12450,
  "published_at": "2026-05-01T10:00:00.000Z",
  "published_by_name": "Super Admin",
  "created_by_name": "Super Admin",
  "created_at": "2026-04-20T10:00:00.000Z",
  "updated_at": "2026-05-01T10:00:00.000Z"
}
```

---

## Settings Endpoints

### 1. Get All Settings (Grouped)

Ambil semua settings, dikelompokkan per group — untuk render seluruh halaman System Settings sekaligus.

- **URL**: `GET /api/v1/web/system-settings`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Query Parameters**:
  - `group` (optional): Filter satu group saja, e.g. `?group=security`
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "general": [
        { "setting_key": "app_name", "value": "KAI App", "value_type": "string", "editable_by": "superadmin", "is_sensitive": false, "description": "Nama platform", "updated_at": "2026-06-01T00:00:00.000Z" }
      ],
      "mobile_app": [ "..." ],
      "security": [ "..." ],
      "email": [
        { "setting_key": "smtp_username", "value": "ap***ey", "value_type": "string", "editable_by": "usergod", "is_sensitive": true, "description": "Username SMTP — password di env SMTP_PASSWORD", "updated_at": "2026-06-01T00:00:00.000Z" }
      ],
      "storage": [ "..." ],
      "payment": [ "..." ],
      "moderation": [ "..." ],
      "maintenance": [ "..." ],
      "contact": [ "..." ]
    }
  }
  ```

> UI backoffice render form per group berdasarkan `value_type` (string → text input, boolean → toggle, number → number input, array → tags/textarea). Field dengan `editable_by = usergod` ditampilkan **disabled/read-only** untuk superadmin.

---

### 2. Update Settings (Bulk per Group)

Update satu atau beberapa setting sekaligus — dipakai tombol "Save Changes" per tab.

- **URL**: `PATCH /api/v1/web/system-settings`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin (+ check `editable_by` per key)
- **Request Body**:
  ```json
  {
    "settings": [
      { "setting_key": "min_version_android", "value": "1.5.0" },
      { "setting_key": "force_update_enabled", "value": true },
      { "setting_key": "update_message", "value": "Update wajib untuk keamanan akun Anda." }
    ]
  }
  ```
- **Validasi** (semua dijalankan sebelum write — all-or-nothing dalam 1 transaksi):
  1. Semua `setting_key` harus exist (tidak ada create key baru via API).
  2. Tipe `value` harus cocok dengan `value_type`.
  3. Aktor berhak per `editable_by`.
  4. Validasi khusus per key: semver valid, `min_version <= latest_version`, `smtp_port` 1–65535, `default_language` ada di `system_languages`, `report_auto_flag_threshold >= 1`, dst.
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "updated_keys": ["min_version_android", "force_update_enabled", "update_message"],
      "cache_invalidated": true
    },
    "message": "3 settings updated successfully"
  }
  ```
- **Side effects**: invalidate settings cache; tulis Audit Log per key (old → new, masked jika sensitive).

> Berbeda dengan bulk role assignment, update settings bersifat **all-or-nothing**: satu key gagal validasi → seluruh request ditolak 422. Setting saling terkait (mis. min vs latest version), partial write berbahaya.

---

### 3. Send Test Email

Verifikasi konfigurasi SMTP dengan mengirim email percobaan. Memakai konfigurasi `email.*` dari DB + password dari env.

- **URL**: `POST /api/v1/web/system-settings/email/test`
- **Autentikasi**: Yes
- **Authorization**: usergod
- **Request Body**:
  ```json
  { "to": "test@example.com" }
  ```
- **Response (Success 200)**:
  ```json
  { "message": "Test email sent to test@example.com" }
  ```
- **Response (Error 400)**:
  ```json
  { "message": "SMTP connection failed: dial tcp: connection refused" }
  ```

---

### 4. Toggle Maintenance Mode (Shortcut)

Sama dengan PATCH `maintenance_mode_enabled`, tapi sebagai endpoint khusus agar mudah di-guard, di-log prioritas tinggi, dan dipakai tombol toggle besar di UI.

- **URL**: `POST /api/v1/web/system-settings/maintenance/toggle`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Request Body**:
  ```json
  {
    "enabled": true,
    "message": "Pemeliharaan terjadwal. Estimasi selesai pukul 03.00 WIB." 
  }
  ```
  > `message` optional — jika kosong, `maintenance_message` existing tetap dipakai.
- **Response (Success 200)**:
  ```json
  {
    "data": { "maintenance_mode_enabled": true },
    "message": "Maintenance mode activated"
  }
  ```

---

## Legal Documents Endpoints

### 1. Get All Legal Documents

Ringkasan 3 dokumen + versi published & draft aktifnya. Untuk render card row di tab Legal & Policies.

- **URL**: `GET /api/v1/web/legal-documents`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Response (Success 200)**:
  ```json
  {
    "data": [
      { "...LegalDocumentObject (terms)" },
      { "...LegalDocumentObject (privacy)" },
      { "...LegalDocumentObject (community_guidelines)" }
    ]
  }
  ```

---

### 2. Get Version List

- **URL**: `GET /api/v1/web/legal-documents/{doc_type}/versions`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Query Parameters**:
  - `status` (optional): `draft` | `published` | `archived`
  - `limit` (optional, default: 20), `offset` (optional, default: 0)
- **Response (Success 200)**:
  ```json
  {
    "data": [
      { "id": "uuid", "version": "2.1.0", "status": "published", "effective_date": "2026-05-15", "require_reacceptance": true, "acceptance_count": 12450, "created_at": "2026-04-20T10:00:00.000Z" },
      { "id": "uuid", "version": "2.0.0", "status": "archived", "effective_date": "2026-01-01", "require_reacceptance": false, "acceptance_count": 11800, "created_at": "2025-12-10T10:00:00.000Z" }
    ],
    "pagination": { "limit": 20, "offset": 0, "total": 5 }
  }
  ```
  > List tidak menyertakan `content` (bisa besar) — ambil via detail.

---

### 3. Get Version Detail

- **URL**: `GET /api/v1/web/legal-documents/{doc_type}/versions/{version_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Response (Success 200)**:
  ```json
  { "data": { "...LegalDocumentVersionObject (termasuk content)" } }
  ```

---

### 4. Create Draft Version

- **URL**: `POST /api/v1/web/legal-documents/{doc_type}/versions`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin (`manage_legal_documents`)
- **Request Body**:
  ```json
  {
    "version": "2.2.0",
    "content": "# Syarat dan Ketentuan...\n\n(markdown, max 100.000 chars)",
    "copy_from_version_id": "uuid (optional — prefill content dari versi lama)"
  }
  ```
  > Jika `copy_from_version_id` diisi dan `content` kosong → content disalin dari versi tersebut (use case "New Version" di UI).
- **Response (Success 201)**:
  ```json
  {
    "data": { "...LegalDocumentVersionObject (status: draft)" },
    "message": "Draft version 2.2.0 created"
  }
  ```
- **Rules**: `version` unique per dokumen (409 jika duplikat). Boleh ada lebih dari satu draft sekaligus.

---

### 5. Update Draft Version

- **URL**: `PUT /api/v1/web/legal-documents/{doc_type}/versions/{version_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Request Body**:
  ```json
  {
    "version": "2.2.0",
    "content": "# Syarat dan Ketentuan (revisi)..."
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": { "...LegalDocumentVersionObject" },
    "message": "Draft updated"
  }
  ```
- **Rules**: hanya status `draft` yang bisa diedit. Versi `published`/`archived` → 400.

---

### 6. Publish Version

Mengaktifkan satu versi. Versi published sebelumnya otomatis `archived` (1 transaksi).

- **URL**: `POST /api/v1/web/legal-documents/{doc_type}/versions/{version_id}/publish`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Request Body**:
  ```json
  {
    "effective_date": "2026-07-01",
    "require_reacceptance": true
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "published_version": { "id": "uuid", "version": "2.2.0", "status": "published", "effective_date": "2026-07-01", "require_reacceptance": true },
      "archived_version": { "id": "uuid", "version": "2.1.0", "status": "archived" }
    },
    "message": "Version 2.2.0 published. Previous version archived."
  }
  ```
- **Rules**:
  - Hanya status `draft` yang bisa di-publish (400 jika bukan).
  - `effective_date` wajib, tidak boleh di masa lalu (boleh hari ini).
  - Jika `require_reacceptance = true` → invalidate cache `pending_acceptances`; user akan diminta setuju ulang saat next login/app open.

---

### 7. Delete Draft Version

- **URL**: `DELETE /api/v1/web/legal-documents/{doc_type}/versions/{version_id}`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Response (Success 200)**:
  ```json
  { "message": "Draft version deleted" }
  ```
- **Rules**: hanya `draft` (400 jika published/archived — versi yang pernah tayang tidak boleh dihapus, ada jejak acceptance).

---

### 8. Get Acceptance Stats

Statistik persetujuan untuk satu versi — untuk monitoring rollout re-acceptance.

- **URL**: `GET /api/v1/web/legal-documents/{doc_type}/versions/{version_id}/acceptance-stats`
- **Autentikasi**: Yes
- **Authorization**: usergod or superadmin
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "version": "2.2.0",
      "acceptance_count": 8210,
      "total_active_users": 12450,
      "acceptance_rate": 0.659,
      "by_source": { "registration": 320, "app": 7800, "web": 90 }
    }
  }
  ```

---

## Admin Flow

```
SYSTEM SETTINGS
  GET   /system-settings                      → render semua tab
  PATCH /system-settings                      → save changes per tab (bulk)
  POST  /system-settings/email/test           → tombol "Send Test"
  POST  /system-settings/maintenance/toggle   → toggle maintenance

LEGAL & POLICIES
  GET    /legal-documents                                     → card 3 dokumen
  GET    /legal-documents/{type}/versions                     → riwayat versi
  GET    /legal-documents/{type}/versions/{id}                → detail + content
  POST   /legal-documents/{type}/versions                     → "New Version" (draft)
  PUT    /legal-documents/{type}/versions/{id}                → "Save Draft"
  POST   /legal-documents/{type}/versions/{id}/publish        → "Publish"
  DELETE /legal-documents/{type}/versions/{id}                → hapus draft
  GET    /legal-documents/{type}/versions/{id}/acceptance-stats → monitoring
```

---

## Important Notes

### ✅ DO:
- ✅ Update settings all-or-nothing dalam 1 transaksi — invalidate cache setelah commit
- ✅ Log setiap perubahan setting & publish legal version ke Audit Log (old → new)
- ✅ Mask nilai `is_sensitive = true` di semua response & audit log
- ✅ Render field `editable_by = usergod` sebagai read-only untuk superadmin
- ✅ Validasi `min_version <= latest_version` saat PATCH group `mobile_app`
- ✅ Konfirmasi dialog di UI sebelum publish dengan `require_reacceptance = true` (berdampak ke semua user)

### ❌ DON'T:
- ❌ Jangan buat endpoint create/delete setting key — daftar key dikunci migration
- ❌ Jangan kembalikan nilai sensitive utuh lewat GET, bahkan untuk usergod
- ❌ Jangan simpan SMTP password / kredensial apa pun lewat API ini (env only)
- ❌ Jangan izinkan edit content versi yang sudah published — buat versi baru
- ❌ Jangan hard-delete versi published/archived (jejak legal acceptance)
- ❌ Jangan terapkan partial success di PATCH settings

---

## Error Handling

### Format A: Standard Message Error
```json
{ "message": "Pesan error deskriptif" }
```

### Format B: Validation Error (422)
```json
{
  "message": "Validation failed",
  "errors": {
    "min_version_android": ["Must be valid semver"],
    "min_version_android ": ["Cannot be greater than latest_version_android (1.4.0)"]
  }
}
```

### Format C: Permission Error (403)
```json
{
  "message": "Insufficient permissions for this action",
  "required_role": "usergod"
}
```

### Skenario Umum

| Scenario | HTTP | Reason |
|----------|------|--------|
| Setting key tidak dikenal | 422 | Key tidak ada di seed — tidak boleh create via API |
| Superadmin update setting `editable_by = usergod` | 403 | Cek `editable_by` |
| Tipe value tidak cocok `value_type` | 422 | Validation |
| `min_version > latest_version` | 422 | Validation khusus mobile_app |
| Publish versi non-draft | 400 | Wrong status |
| Edit/delete versi published | 400 | Immutable setelah publish |
| Versi duplikat per dokumen | 409 | Conflict |
| `effective_date` di masa lalu | 422 | Validation |
| `doc_type` tidak dikenal | 404 | Hanya terms/privacy/community_guidelines |
| SMTP test gagal koneksi | 400 | Pesan error koneksi diteruskan |
| Non-admin akses | 403 | Authorization failed |

### Status Codes

`200` OK · `201` Created · `400` Bad Request · `401` Unauthorized · `403` Forbidden · `404` Not Found · `409` Conflict · `422` Unprocessable Entity · `500` Internal Server Error

---

*API spec System Settings untuk web backoffice. Mobile spec di `API_SPEC_SYSTEM_SETTINGS_MOBILE.md`.*
