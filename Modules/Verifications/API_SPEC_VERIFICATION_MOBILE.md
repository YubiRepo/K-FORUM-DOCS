# API Specification — Verification Badge Module (Mobile) v1.0

Endpoint member-facing buat ngajuin verifikasi (User / Merchant), lihat syarat dokumen, dan cek status pengajuan sendiri. Review/approve ada di sisi Backoffice.

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/mobile/verification`
- **Auth:** Wajib. `Authorization: Bearer <access_token>`
- **Badge type = `entity_type`:** `user` | `merchant` | `community`.
- Member cuma bisa ngajuin/lihat verifikasi **entitas miliknya sendiri** (akun sendiri, merchant yang dia owner, atau komunitas di mana dia **Leader/owner**).
- Khusus `community`: requester **wajib** Leader/owner komunitas itu (permission `verification.request_community`). Member biasa → 403.

### Error Responses (standar)

| Code | Arti |
|------|------|
| 400 | Validasi gagal (dokumen kurang dari syarat, entity_type invalid) |
| 401 | Token invalid |
| 403 | Bukan pemilik entitas / bukan Leader komunitas / prasyarat belum terpenuhi (user non-active, merchant belum published, community non-active) |
| 409 | Sudah ada pengajuan `pending`, atau entitas sudah `approved` |
| 422 | Dokumen tidak memenuhi `match_mode` |

---

## Daftar Isi

1. GET /requirements — ambil syarat dokumen
2. POST /requests — ajukan verifikasi
3. GET /requests/mine — status verifikasi entitas sendiri

---

## 1. Ambil Syarat Dokumen

Dipakai UI buat nampilin dokumen apa aja yang diterima sebelum submit.

- **URL:** `GET /api/v1/mobile/verification/requirements`
- **Auth:** Yes (Member)
- **Query Params:**

| Param | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `entity_type` | string | Yes | `user` \| `merchant` \| `community` |

- **Response 200:**
```json
{
  "data": {
    "entity_type": "user",
    "match_mode": "any_of",
    "min_documents": 1,
    "accepted_docs": [
      { "key": "kta_kai",         "label": "KTA KAI (kartu anggota)",         "required": false },
      { "key": "kta_other",       "label": "KTA organisasi lain",             "required": false },
      { "key": "gov_id",          "label": "KTP / Paspor / SIM (opsional)",   "required": false },
      { "key": "membership_proof","label": "Bukti jabatan / keanggotaan",     "required": false },
      { "key": "social_proof",    "label": "Akun sosial media terverifikasi", "required": false }
    ],
    "hint": "Cukup lampirkan minimal 1 dokumen pendukung."
  }
}
```
> Field `sensitive` sengaja **tidak** di-expose ke mobile (internal only).

---

## 2. Ajukan Verifikasi

- **URL:** `POST /api/v1/mobile/verification/requests`
- **Auth:** Yes (Member — pemilik entitas)
- **Catatan:** dokumen di-upload dulu via media service (`context: "verification"`), lalu URL-nya dikirim di sini.

- **Request Body:**
```json
{
  "entity_type": "user",
  "entity_id": "usr_uuid_self",
  "documents": [
    { "doc_type": "kta_kai", "url": "https://media.kai.app/private/verif/abc.jpg" }
  ],
  "note": "Pengurus KAI wilayah Seoul"
}
```

- **Validasi:**
  - `entity_id` harus milik requester (user = dirinya; merchant = merchant yg dia owner; community = komunitas di mana dia **Leader/owner**).
  - Prasyarat: user `status='active'` / merchant `status='published'` / community `status='active'`.
  - Dokumen harus memenuhi `match_mode` dari requirements.
  - Tidak boleh ada pengajuan `pending` / sudah `approved`.

- **Response 201:**
```json
{
  "data": {
    "id": "verif_uuid",
    "entity_type": "user",
    "entity_id": "usr_uuid_self",
    "status": "pending",
    "submitted_at": "2026-07-13T09:00:00.000Z"
  }
}
```

- **Response 409:**
```json
{ "error": { "code": "ALREADY_PENDING", "message": "Sudah ada pengajuan yang sedang diproses." } }
```

---

## 3. Status Verifikasi Entitas Sendiri

Ambil pengajuan terbaru + status badge. Kalau `rejected`/`revoked`, `reason` ikut dikirim biar pemohon tau kenapa & bisa resubmit.

- **URL:** `GET /api/v1/mobile/verification/requests/mine`
- **Auth:** Yes (Member)
- **Query Params:**

| Param | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `entity_type` | string | Yes | `user` \| `merchant` \| `community` |
| `entity_id` | string | Yes | id entitas milik requester |

- **Response 200 (approved):**
```json
{
  "data": {
    "id": "verif_uuid",
    "entity_type": "user",
    "entity_id": "usr_uuid_self",
    "status": "approved",
    "is_verified": true,
    "reason": null,
    "submitted_at": "2026-07-10T09:00:00.000Z",
    "reviewed_at": "2026-07-12T14:00:00.000Z",
    "can_resubmit": false
  }
}
```

- **Response 200 (rejected):**
```json
{
  "data": {
    "status": "rejected",
    "is_verified": false,
    "reason": "Foto KTA tidak terbaca, mohon upload ulang.",
    "can_resubmit": true
  }
}
```

- **Response 200 (belum pernah ngajuin):**
```json
{ "data": { "status": null, "is_verified": false, "can_resubmit": true } }
```

> **Jangan** expose dokumen, `reviewed_by`, atau detail internal di response mobile. Cuma status + reason (buat reject/revoke).

---

## CATATAN

- Badge (`is_verified` + `verification_type`) di listing/profil di-expose lewat object User/Merchant di modul masing-masing — **bukan** endpoint ini. Endpoint ini khusus flow pengajuan.
- Resubmit = panggil `POST /requests` lagi (setelah `rejected`/`revoked`). Record lama tetap tersimpan (append-only).
- Notifikasi hasil (`approved`/`rejected`/`revoked`) dikirim via FCM oleh modul Notification.
- **Detail mekanisme upload dokumen** (presign → PUT storage → submit, handling presign expired) ada di `VERIFICATION_DOCUMENT_ACCESS_SPEC.md` §3 & §5.1 — tidak diulang di sini.

---

*Verification Badge Mobile API v1.0 — KAI App. Step 3 dari pipeline. Last updated: 2026-07-13*
