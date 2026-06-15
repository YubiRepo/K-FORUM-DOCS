# API Spec — Schedule Backoffice (v1.1)

API specification untuk Schedule module KAI App — **backoffice surface**, diakses semua admin/staff backoffice (dikontrol via role/permission). Tidak ada surface mobile pada Phase 1.

Base URL: `/api/v1/admin/schedule`
Auth: Bearer token (admin/staff backoffice)

Sharing: tiga mode visibility — `private`, `all_admins`, `specific` (invite user tertentu, view-only).

---

## ENTRIES

### 1. POST /entries
Buat agenda baru.

**Auth:** Admin/staff

**Request Body:**
```json
{
  "title": "Review modul Schedule",
  "description": "Bahas progres",
  "entry_type_id": "type_agenda",
  "start_at": "2026-06-16T10:00:00.000Z",
  "end_at": "2026-06-16T11:00:00.000Z",
  "all_day": false,
  "location": "Ruang Rapat Lt.3",
  "visibility": "specific",
  "shared_with": ["usr_admin_b", "usr_admin_c"]
}
```

**Validation:**
- `title`: required, max 200
- `entry_type_id`: required, tipe harus aktif
- `start_at`: required
- `end_at`: opsional; jika ada harus ≥ `start_at`
- `all_day`: default `false`
- `visibility`: `private` (default) | `all_admins` | `specific`
- `shared_with`: array user_id. **Wajib & non-empty jika `visibility = specific`**; diabaikan untuk mode lain. User harus akun backoffice yang valid.

**Response (201 Created):**
```json
{
  "data": {
    "id": "sch_001",
    "title": "Review modul Schedule",
    "entry_type": { "id": "type_agenda", "code": "agenda", "name": "Agenda", "color": "#3B82F6" },
    "start_at": "2026-06-16T10:00:00.000Z",
    "end_at": "2026-06-16T11:00:00.000Z",
    "all_day": false,
    "location": "Ruang Rapat Lt.3",
    "visibility": "specific",
    "shared_with": [
      { "id": "usr_admin_b", "name": "Admin B" },
      { "id": "usr_admin_c", "name": "Admin C" }
    ],
    "status": "active",
    "source": "manual",
    "created_by": "usr_admin",
    "created_at": "2026-06-15T08:00:00.000Z"
  }
}
```

**Response (422 — specific tanpa shared_with):**
```json
{ "message": "Validation failed", "errors": { "shared_with": ["Required and must be non-empty when visibility is 'specific'"] } }
```

---

### 2. GET /entries
List agenda untuk tampilan kalender. Hasil otomatis ter-scope: agenda milik sendiri (semua mode) + `all_admins` + `specific` yang meng-invite user ini. Superadmin melihat semua.

**Auth:** Admin/staff

**Query Parameters:**
- `from`, `to` — rentang waktu (untuk render bulan/minggu/hari)
- `entry_type_id`
- `status` — `active` | `done` | `cancelled`
- `visibility` — `private` | `all_admins` | `specific`
- `mine` (bool) — hanya agenda milik sendiri
- `limit` (default 100, max 500), `offset`

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "sch_001",
      "title": "Review modul Schedule",
      "entry_type": { "code": "agenda", "color": "#3B82F6" },
      "start_at": "2026-06-16T10:00:00.000Z",
      "end_at": "2026-06-16T11:00:00.000Z",
      "all_day": false,
      "visibility": "specific",
      "status": "active",
      "source": "manual",
      "created_by": "usr_admin",
      "is_owner": true
    }
  ],
  "pagination": { "limit": 100, "offset": 0, "total": 1 }
}
```
> `is_owner` membantu UI menentukan apakah tombol edit/hapus ditampilkan.

---

### 3. GET /entries/:id
Detail satu agenda (termasuk `shared_with` jika mode `specific`).

**Auth:** Admin/staff (boleh lihat jika miliknya, `all_admins`, atau di-invite; Superadmin semua)

**Response (200 OK):** objek agenda lengkap + daftar `shared_with`.

**Response (403 — tidak berhak lihat):**
```json
{ "message": "You do not have access to this entry" }
```

---

### 4. PUT /entries/:id
Edit agenda (termasuk ganti `visibility` & `shared_with`). Hanya pembuat (atau Superadmin).

**Auth:** Pembuat / Superadmin

**Request Body:** field yang sama dengan POST (subset boleh). Mengirim `shared_with` akan menyinkronkan daftar invite (tambah/hapus selisih).

**Response (200 OK):** agenda ter-update.

**Response (403 — bukan pembuat):**
```json
{ "message": "Only the creator or a superadmin can edit this entry" }
```

---

### 5. POST /entries/:id/share
Kelola daftar invite (mode `specific`) tanpa harus PUT seluruh entri. Menambah/menghapus user.

**Auth:** Pembuat / Superadmin

**Request Body:**
```json
{ "add": ["usr_admin_d"], "remove": ["usr_admin_c"] }
```

**Response (200 OK):**
```json
{
  "data": {
    "id": "sch_001",
    "visibility": "specific",
    "shared_with": [
      { "id": "usr_admin_b", "name": "Admin B" },
      { "id": "usr_admin_d", "name": "Admin D" }
    ]
  }
}
```

**Response (409 — bukan mode specific):**
```json
{ "message": "Sharing list only applies when visibility is 'specific'" }
```

---

### 6. POST /entries/:id/status
Ubah status agenda (tandai selesai / batalkan).

**Auth:** Pembuat / Superadmin

**Request Body:**
```json
{ "status": "done" }
```
> `status`: `done` | `cancelled` | `active`

**Response (200 OK):**
```json
{ "data": { "id": "sch_001", "status": "done" } }
```

---

### 7. DELETE /entries/:id
Hapus agenda permanen. Hanya pembuat (atau Superadmin). Untuk agenda yang dibagikan, disarankan `cancelled` daripada delete.

**Auth:** Pembuat / Superadmin

**Response (200 OK):**
```json
{ "message": "Schedule entry deleted" }
```

---

## ENTRY TYPES

### 8. GET /entry-types
List tipe agenda (untuk dropdown & warna kalender).

**Auth:** Admin/staff

**Query Parameters:** `active` (bool)

**Response (200 OK):**
```json
{
  "data": [
    { "id": "type_agenda", "code": "agenda", "name": "Agenda", "color": "#3B82F6", "is_active": true },
    { "id": "type_reminder", "code": "reminder", "name": "Reminder", "color": "#F59E0B", "is_active": true },
    { "id": "type_milestone", "code": "milestone", "name": "Milestone", "color": "#10B981", "is_active": true }
  ]
}
```

### 9. POST /entry-types
Buat tipe baru.

**Auth:** Superadmin only

**Request Body:**
```json
{ "code": "deadline", "name": "Deadline", "color": "#EF4444" }
```

**Response (201 Created):** tipe baru.

### 10. PUT /entry-types/:id
Edit tipe (nama, warna, is_active). `code` tidak bisa diubah jika sudah dipakai.

**Auth:** Superadmin only

**Response (200 OK):** tipe ter-update.

> Tipe yang sudah dipakai agenda tidak bisa dihapus — hanya `is_active = false`.

---

## UTILITAS

### 11. GET /shareable-users
Daftar user backoffice yang bisa di-invite (untuk dropdown saat mode `specific`).

**Auth:** Admin/staff

**Query Parameters:** `q` (search nama), `limit`, `offset`

**Response (200 OK):**
```json
{
  "data": [
    { "id": "usr_admin_b", "name": "Admin B", "role": "admin" },
    { "id": "usr_admin_c", "name": "Admin C", "role": "staff" }
  ]
}
```

---

## CATATAN ARSITEKTUR
1. **Scope visibility wajib di-enforce di backend** — agenda yang tidak boleh dilihat (private/specific orang lain) tidak boleh bocor; Superadmin pengecualian.
2. **Sharing view-only** — user yang di-invite atau semua admin hanya bisa melihat; edit/hapus tetap pembuat (atau Superadmin).
3. **`shared_with` hanya berlaku saat `visibility = specific`** — diabaikan untuk mode lain (tetap tersimpan untuk dipakai bila kembali ke `specific`).
4. **Invite hanya user internal** — bukan email orang luar. Untuk berbagi keluar, gunakan export `.ics` (Phase 3).
5. **Field hook (`source`, `source_module`, `source_ref`, `recurrence`, `assigned_to`)** disiapkan tapi tidak diisi via endpoint manual Phase 1 — untuk linked entries & recurrence di Phase 2/3.
6. **Future region** — mode `region` cukup mengisi `schedule_entry_shares` otomatis; tidak ada perubahan endpoint inti.
