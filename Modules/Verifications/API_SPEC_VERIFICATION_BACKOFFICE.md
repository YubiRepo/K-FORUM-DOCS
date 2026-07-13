# API Specification — Verification Badge Module (Backoffice) v1.0

Endpoint Superadmin buat review antrian verifikasi, lihat dokumen, approve/reject/revoke, lihat audit trail, dan kelola syarat dokumen (rules-as-data).

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/web/verification`
- **Auth:** Wajib. `Authorization: Bearer <access_token>`
- **Role:** **Superadmin only** (semua endpoint). Admin Regional hanya boleh read status via modul lain, tidak punya akses ke sini.
- **Permission keys:** `verification.view_queue`, `verification.review`, `verification.revoke`.

### Error Responses (standar)

| Code | Arti |
|------|------|
| 400 | Validasi gagal (reason kosong saat reject/revoke) |
| 401 | Token invalid |
| 403 | Bukan Superadmin / permission kurang |
| 404 | Verification request tidak ditemukan |
| 409 | Transisi status invalid (mis. approve request yg bukan `pending`) |

---

## Daftar Isi

1. GET /requests — antrian & daftar pengajuan
2. GET /requests/:id — detail + dokumen
3. POST /requests/:id/approve
4. POST /requests/:id/reject
5. POST /requests/:id/revoke
6. GET /requests/:id/events — audit trail
7. GET /requirements — lihat config
8. PUT /requirements/:entity_type — ubah config

---

## Data Models

### VerificationAdminObject
```json
{
  "id": "verif_uuid",
  "entity_type": "user",
  "entity_id": "usr_uuid",
  "entity": { "id": "usr_uuid", "name": "Budi Santoso", "avatar_url": "..." },
  "status": "pending",
  "submitted_by": { "id": "usr_uuid", "name": "Budi Santoso" },
  "documents": [
    { "doc_type": "kta_kai", "url": "https://media.kai.app/signed/...", "uploaded_at": "2026-07-10T09:00:00.000Z" }
  ],
  "note": "Pengurus KAI wilayah Seoul",
  "reviewed_by": null,
  "reviewed_at": null,
  "rejection_reason": null,
  "revoke_reason": null,
  "created_at": "2026-07-10T09:00:00.000Z",
  "updated_at": "2026-07-10T09:00:00.000Z"
}
```
> `entity` di-resolve: `entity_type=user` → dari `users`; `merchant` → dari `merchants`; `community` → dari `communities`. `documents[].url` = signed URL short-lived (dokumen private).

---

## 1. List / Antrian Pengajuan

- **URL:** `GET /api/v1/web/verification/requests`
- **Auth:** Yes (Superadmin, `verification.view_queue`)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `status` | string | No | `pending` | `pending`\|`approved`\|`rejected`\|`revoked`\|`all` |
| `entity_type` | string | No | — | filter `user`\|`merchant`\|`community` |
| `q` | string | No | — | cari nama entitas / pemohon |
| `page` | int | No | 1 | |
| `per_page` | int | No | 20 | |

- **Response 200:**
```json
{
  "data": [ { "...VerificationAdminObject (ringkas, tanpa documents)": "..." } ],
  "pagination": { "page": 1, "per_page": 20, "total": 8, "total_pages": 1 }
}
```

---

## 2. Detail Pengajuan

- **URL:** `GET /api/v1/web/verification/requests/:id`
- **Auth:** Yes (Superadmin, `verification.view_queue`)
- **Response 200:** `{ "data": { ...VerificationAdminObject (lengkap + documents signed) } }`
- **Response 404:** `{ "error": { "code": "NOT_FOUND", "message": "Pengajuan tidak ditemukan." } }`

---

## 3. Approve

- **URL:** `POST /api/v1/web/verification/requests/:id/approve`
- **Auth:** Yes (Superadmin, `verification.review`)
- **Efek:** `status='approved'`, set `reviewed_by/at`, set cache `is_verified=true` di entitas (transaksi sama), tulis `verification_events` (`approved`), trigger notif `verification.approved`.
- **Request Body:** *(opsional)* `{ "internal_note": "Identitas & KTA valid" }`
- **Response 200:**
```json
{ "data": { "id": "verif_uuid", "status": "approved", "reviewed_at": "2026-07-12T14:00:00.000Z" } }
```
- **Response 409:** `{ "error": { "code": "INVALID_TRANSITION", "message": "Hanya pengajuan berstatus pending yang bisa di-approve." } }`

---

## 4. Reject

- **URL:** `POST /api/v1/web/verification/requests/:id/reject`
- **Auth:** Yes (Superadmin, `verification.review`)
- **Efek:** `status='rejected'`, simpan `rejection_reason`, tulis event (`rejected`), trigger notif `verification.rejected`. `is_verified` tetap false. Pemohon boleh resubmit.
- **Request Body:**
```json
{ "reason": "Foto KTA tidak terbaca, mohon upload ulang." }
```
- **Validasi:** `reason` **wajib**, non-empty.
- **Response 200:** `{ "data": { "id": "verif_uuid", "status": "rejected" } }`
- **Response 400:** `{ "error": { "code": "REASON_REQUIRED", "message": "Alasan penolakan wajib diisi." } }`

---

## 5. Revoke

Cabut badge yang sudah `approved`.

- **URL:** `POST /api/v1/web/verification/requests/:id/revoke`
- **Auth:** Yes (Superadmin, `verification.revoke`)
- **Efek:** `status='revoked'`, simpan `revoke_reason`, set cache `is_verified=false`, tulis event (`revoked`), trigger notif `verification.revoked`.
- **Request Body:**
```json
{ "reason": "Terindikasi impersonation / laporan pelanggaran." }
```
- **Validasi:** hanya dari status `approved`; `reason` **wajib**.
- **Response 200:** `{ "data": { "id": "verif_uuid", "status": "revoked" } }`
- **Response 409:** `{ "error": { "code": "INVALID_TRANSITION", "message": "Hanya badge approved yang bisa dicabut." } }`

---

## 6. Audit Trail

Riwayat append-only semua aksi pada satu pengajuan.

- **URL:** `GET /api/v1/web/verification/requests/:id/events`
- **Auth:** Yes (Superadmin, `verification.view_queue`)
- **Response 200:**
```json
{
  "data": [
    { "action": "submitted", "actor": { "id": "usr_uuid", "name": "Budi" },       "reason": null, "created_at": "2026-07-10T09:00:00.000Z" },
    { "action": "rejected",  "actor": { "id": "adm_uuid", "name": "Admin KAI" },   "reason": "Foto buram", "created_at": "2026-07-11T10:00:00.000Z" }
  ]
}
```

---

## 7. Lihat Config Syarat Dokumen

- **URL:** `GET /api/v1/web/verification/requirements`
- **Auth:** Yes (Superadmin)
- **Response 200:**
```json
{
  "data": [
    {
      "entity_type": "user",
      "match_mode": "any_of",
      "min_documents": 1,
      "accepted_docs": [
        { "key": "kta_kai", "label": "KTA KAI (kartu anggota)", "required": false, "sensitive": true }
      ],
      "is_active": true,
      "updated_at": "2026-07-13T00:00:00.000Z"
    }
  ]
}
```

---

## 8. Ubah Config Syarat Dokumen

Rules-as-data — ubah tanpa deploy.

- **URL:** `PUT /api/v1/web/verification/requirements/:entity_type`
- **Auth:** Yes (Superadmin)
- **Request Body:**
```json
{
  "match_mode": "any_of",
  "min_documents": 1,
  "accepted_docs": [
    { "key": "kta_kai", "label": "KTA KAI (kartu anggota)", "required": false, "sensitive": true },
    { "key": "gov_id",  "label": "KTP / Paspor / SIM (opsional)", "required": false, "sensitive": true }
  ]
}
```
- **Validasi:** `match_mode` ∈ {`any_of`,`all_of`}; kalau `all_of`, minimal 1 doc `required=true`.
- **Response 200:** `{ "data": { "entity_type": "user", "match_mode": "any_of", "updated_at": "..." } }`

---

## CATATAN

- **Transaksi atomik:** approve/revoke wajib update `verifications` + cache `is_verified` + insert `verification_events` dalam satu transaksi DB.
- **Signed URL:** dokumen sensitif dikasih signed URL short-lived; jangan simpan URL permanen di response yang bisa ke-cache.
- **Cross-module:** notif event (`verification.submitted/approved/rejected/revoked`) & permission keys didaftarin di modul Notification & Role-Permission (lihat RULES §7 & §5).

---

*Verification Badge Backoffice API v1.0 — KAI App. Step 3 dari pipeline. Last updated: 2026-07-13*
