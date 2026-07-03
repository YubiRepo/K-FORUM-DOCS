# API Spec — Community Invite & Share Link (Mobile Client)

Dokumentasi API endpoint untuk **Undangan (invite in-app)** & **Share Link / Join Code** modul Community (Flutter). Untuk aturan lihat `COMMUNITY_INVITE_RULES.md`, skema `COMMUNITY_INVITE_DB_SCHEMA.md`.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Accept-Language: <lang_code>` (`ko` | `id` | `en`, default `ko`)
  - `Authorization: Bearer <access_token>`
- **Authentication**: Wajib untuk semua endpoint.
- **Permission**: Aksi kelola (undang, buat/cabut link) butuh `manage_members` pada komunitas terkait. Accept/reject undangan & redeem link tidak butuh permission (tunduk aturan bisnis).

Endpoint dibagi dua kelompok path:
- Kelola per komunitas → `/communities/{community_id}/...`
- Aksi milik user (undangan masuk, redeem) → `/community-invitations/...`, `/community-invite-links/...`

---

## Model Data Utama

### 1. Invitation Object
```json
{
  "id": "inv_001",
  "community": { "id": "comm_futsal", "name": "Futsal Jakarta", "avatar": "https://cdn/.../c.jpg", "visibility": "private" },
  "invitee": { "id": "usr_05", "name": "Andi", "username": "andi", "avatar": null },
  "invited_by": { "id": "usr_01", "name": "Budi" },
  "message": "Gabung yuk latihan tiap Sabtu!",
  "status": "pending",
  "expires_at": "2026-07-10T08:00:00.000Z",
  "created_at": "2026-07-03T08:00:00.000Z"
}
```

### 2. Invite Link Object
```json
{
  "id": "lnk_001",
  "community_id": "comm_futsal",
  "code": "f4Kx9aQ2",
  "url": "https://kai.app/c/join/f4Kx9aQ2",
  "requires_approval": false,
  "expires_at": null,
  "max_uses": null,
  "use_count": 12,
  "is_active": true,
  "created_at": "2026-07-03T08:00:00.000Z"
}
```

### 3. Error Responses
```json
{ "message": "Pesan error deskriptif" }
```
Validation (422): `{ "message": "Data input tidak valid", "errors": { "invitee_id": ["User tidak ditemukan."] } }`

---

# BAGIAN A — UNDANGAN (INVITE IN-APP)

## A1. Cari User untuk Diundang
Cari user terdaftar yang belum jadi anggota (untuk autocomplete saat undang).

- **URL**: `GET /communities/{community_id}/invitable-users`
- **Autentikasi**: Yes + `manage_members`
- **Query**: `q` (cari username/nama, min 2 char), `limit` (default 10, max 20)
- **Response (200)**:
  ```json
  {
    "data": [
      { "id": "usr_05", "name": "Andi", "username": "andi", "avatar": null, "already_member": false, "pending_invite": false }
    ]
  }
  ```
> Backend menandai `already_member` / `pending_invite` supaya client bisa disable pilihan yang tak relevan.

## A2. Kirim Undangan
- **URL**: `POST /communities/{community_id}/invitations`
- **Autentikasi**: Yes + `manage_members`
- **Request Body**:
  ```json
  { "invitee_id": "usr_05", "message": "Gabung yuk latihan tiap Sabtu!" }
  ```
- **Response (201)**: `{ "data": { Invitation Object }, "message": "Invitation sent" }`
- **Rules / Error**:
  - `409` bila invitee sudah anggota aktif, atau ada undangan `pending` duplikat.
  - `403` bila invitee `banned` di komunitas ini.
  - Sukses → emit `community_invitation_received` (push ke invitee).

## A3. Daftar Undangan Komunitas (Pengelola)
- **URL**: `GET /communities/{community_id}/invitations`
- **Autentikasi**: Yes + `manage_members`
- **Query**: `status` (`pending`|`accepted`|`rejected`|`expired`|`cancelled`), `limit`, `offset`
- **Response (200)**:
  ```json
  {
    "data": [ { "id": "inv_001", "invitee": { "id": "usr_05", "name": "Andi" }, "status": "pending", "created_at": "2026-07-03T08:00:00.000Z" } ],
    "pagination": { "limit": 20, "offset": 0, "total": 3 }
  }
  ```

## A4. Cancel Undangan
- **URL**: `DELETE /communities/{community_id}/invitations/{invitation_id}`
- **Autentikasi**: Yes + `manage_members`
- **Response (200)**: `{ "message": "Invitation cancelled" }`
- **Rules**: Hanya undangan `pending` yang bisa di-cancel (`409` bila sudah accepted/rejected).

## A5. Undangan Masuk (Milik User)
Daftar undangan yang ditujukan ke user yang sedang login.

- **URL**: `GET /me/community-invitations`
- **Autentikasi**: Yes
- **Query**: `status` (default `pending`), `limit`, `offset`
- **Response (200)**:
  ```json
  {
    "data": [ { Invitation Object } ],
    "pagination": { "limit": 20, "offset": 0, "total": 1 }
  }
  ```

## A6. Accept Undangan
- **URL**: `POST /community-invitations/{invitation_id}/accept`
- **Autentikasi**: Yes (harus invitee-nya)
- **Response (200)**:
  ```json
  { "message": "Joined community", "data": { "community": { "id": "comm_futsal", "name": "Futsal Jakarta" } } }
  ```
- **Rules**:
  - Undangan harus `pending` & belum expired → `410` bila expired, `409` bila sudah direspons.
  - **Bypass approval** walau komunitas private: langsung `community_members(active)` + role member.
  - Idempoten bila sudah anggota (Rule 3) → tetap `200`, tidak duplikat.
  - `403` bila user `banned`.

## A7. Reject Undangan
- **URL**: `POST /community-invitations/{invitation_id}/reject`
- **Autentikasi**: Yes (harus invitee-nya)
- **Response (200)**: `{ "message": "Invitation rejected" }`

---

# BAGIAN B — SHARE LINK / JOIN CODE

## B1. Buat Share Link
- **URL**: `POST /communities/{community_id}/invite-links`
- **Autentikasi**: Yes + `manage_members`
- **Request Body** (semua opsional):
  ```json
  { "requires_approval": false, "expires_at": null, "max_uses": null }
  ```
  > Default: langsung-join, tanpa expiry, tanpa kuota. Untuk komunitas public, `requires_approval` diabaikan.
- **Response (201)**: `{ "data": { Invite Link Object }, "message": "Invite link created" }`

## B2. Daftar Share Link (Pengelola)
- **URL**: `GET /communities/{community_id}/invite-links`
- **Autentikasi**: Yes + `manage_members`
- **Query**: `active` (bool filter)
- **Response (200)**:
  ```json
  {
    "data": [
      { "id": "lnk_001", "code": "f4Kx9aQ2", "url": "https://kai.app/c/join/f4Kx9aQ2", "requires_approval": false, "expires_at": null, "max_uses": null, "use_count": 12, "joined_count": 12, "requested_count": 0, "is_active": true }
    ]
  }
  ```

## B3. Cabut Share Link
- **URL**: `DELETE /communities/{community_id}/invite-links/{link_id}`
- **Autentikasi**: Yes + `manage_members`
- **Response (200)**: `{ "message": "Invite link revoked" }`
> Set `is_active=false`. Redemption yang sudah terjadi tidak terpengaruh.

## B4. Preview Link (Sebelum Join)
Dipanggil app saat user membuka deep link, untuk menampilkan info komunitas sebelum konfirmasi join.

- **URL**: `GET /community-invite-links/{code}`
- **Autentikasi**: Yes
- **Response (200)**:
  ```json
  {
    "data": {
      "community": { "id": "comm_futsal", "name": "Futsal Jakarta", "avatar": "https://cdn/.../c.jpg", "visibility": "private", "member_count": 42 },
      "requires_approval": false,
      "link_status": "active",
      "already_member": false
    }
  }
  ```
- **Rules**: `410` bila link mati (dicabut/expired/kuota habis).

## B5. Redeem Link (Join / Request)
- **URL**: `POST /community-invite-links/{code}/redeem`
- **Autentikasi**: Yes
- **Response (200) — langsung join**:
  ```json
  { "joined": true, "requested": false, "data": { "community": { "id": "comm_futsal", "name": "Futsal Jakarta" } }, "message": "Joined community" }
  ```
- **Response (200) — perlu approval**:
  ```json
  { "joined": false, "requested": true, "message": "Join request submitted" }
  ```
- **Rules / Error**:
  - `410` bila link mati.
  - `403` bila user `banned` di komunitas.
  - Sudah anggota → `200` idempoten (`joined:true`, tidak menambah `use_count`), arahkan ke komunitas.
  - Public OR `requires_approval=false` → langsung join. Private + `requires_approval=true` → buat `community_join_requests(pending)`.
  - Operasi atomik (lock baris link) supaya `max_uses` tidak kelewat.

---

## Status Code Reference

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request |
| `401` | Unauthorized |
| `403` | Forbidden — tak punya `manage_members`, atau user banned |
| `404` | Not Found — komunitas/undangan/link tidak ada |
| `409` | Conflict — sudah anggota, undangan duplikat, atau sudah direspons |
| `410` | Gone — undangan/link kedaluwarsa, dicabut, atau kuota habis |
| `422` | Validation error |
| `500` | Internal Server Error |

---

## Notes & Best Practices

1. **Permission per komunitas.** Endpoint kelola (`/communities/{id}/...`) cek `manage_members` pada komunitas itu; jangan andalkan gating client.
2. **Deep link.** `url` berformat `https://kai.app/c/join/{code}`. App meng-handle → panggil B4 (preview) → user konfirmasi → B5 (redeem). Universal/App Links disarankan.
3. **Bypass approval.** Accept undangan (A6) & redeem link `requires_approval=false` (B5) melewati alur approval private. Join-request manual biasa tetap butuh approval leader (di luar spec ini).
4. **Idempotensi.** Accept/redeem oleh anggota existing = sukses no-op, bukan error.
5. **410 vs 404.** Link/undangan yang pernah ada tapi mati → `410 Gone` (beda dari `404` untuk yang memang tidak pernah ada) supaya client bisa kasih pesan tepat.
6. **Timestamp ISO 8601 UTC.**
7. **Rate limiting** disarankan pada A2 (kirim undangan) & B5 (redeem) untuk cegah spam.
8. **Email invite** belum ada di Phase 1 — semua undangan menargetkan user terdaftar (`invitee_id`). Invite non-user via email = Phase 2.
