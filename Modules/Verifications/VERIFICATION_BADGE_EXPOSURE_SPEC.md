# Verification Badge — Exposure Spec (Public Read) v1.0

Spec pelengkap `API_SPEC_VERIFICATION_MOBILE.md`. Dokumen ini menjawab satu pertanyaan spesifik: **dari field mana user lain / publik tahu sebuah entitas (user / merchant / community) sudah terverifikasi?**

> `API_SPEC_VERIFICATION_MOBILE.md` §5 menyatakan badge "di-expose lewat object User/Merchant di modul masing-masing" — tapi per audit 2026-07-16 (cek `curl` langsung ke production) field itu **belum ada** di response endpoint manapun. Dokumen ini mengunci nama field, semantik, dan daftar endpoint yang wajib menambahkannya.

---

## 1. Field Kontrak

Setiap object entitas yang bisa dilihat publik menambahkan **satu field**:

| Field | Type | Wajib | Arti |
|-------|------|:---:|------|
| `is_verified` | boolean | ✅ | `true` jika entitas punya row `verifications` dengan `status='approved'` (badge centang aktif). Default `false`. |
| `verified_at` | string (ISO-8601) \| null | opsional | Kapan badge disetujui (`reviewed_at` dari approve). Untuk teks "Terverifikasi sejak …". Boleh ditunda; `is_verified` yang wajib. |

**Sumber kebenaran** = kolom cache `is_verified` di tabel `users` / `merchants` / `communities` (di-maintain app-layer saat approve/revoke — lihat `VERIFICATION_DB_SCHEMA.md` §Resolve-live). Endpoint publik cukup baca kolom cache ini, **tidak** perlu query tabel `verifications` per-row.

**Catatan penting:**
- **Tidak ada** field `verification_type` terpisah. "Tipe" badge = jenis entitasnya sendiri (badge di object merchant = merchant terverifikasi; di object user = user terverifikasi). Frontend menyimpulkan konteks dari object tempat field itu berada.
- Jangan bingung dengan field yang sudah ada tapi **beda makna**:
  - User `email_verified` / `phone_verified` → verifikasi kontak, **bukan** badge keaslian akun.
  - Merchant `approval_status` / `approved_at` → approval publikasi (draft → published), **bukan** badge verifikasi.
- `communities` juga punya kolom cache `is_verified` (DB Schema §tree entitas) — badge community wajib ikut di-expose, bukan cuma user & merchant.

---

## 2. Endpoint yang Wajib Di-update (Backend)

### 2.1 Merchant (modul Directory)

| Endpoint | Object | Aksi |
|----------|--------|------|
| `GET /api/v1/mobile/directory/merchants` | tiap item list | + `is_verified` |
| `GET /api/v1/mobile/directory/merchants/:id` | detail | + `is_verified` (+ `verified_at` opsional) |
| `GET /api/v1/mobile/directory/me/merchants` | tiap item | + `is_verified` (biar owner lihat status badge-nya di list) |
| `GET /api/v1/mobile/directory/merchants/:id/manage` | detail manage | + `is_verified` |

### 2.2 Community (modul Community)

| Endpoint | Object | Aksi |
|----------|--------|------|
| `GET /api/v1/mobile/communities` | tiap item list | + `is_verified` |
| `GET /api/v1/mobile/communities/:id` | detail | + `is_verified` (+ `verified_at` opsional) |

### 2.3 User (modul Profile + shared user object)

Badge user muncul di **banyak tempat** karena object user disematkan (embedded) di berbagai response. Cara paling bersih: tambahkan `is_verified` ke **serializer user publik yang dipakai bersama**, supaya otomatis ikut ke semua endpoint di bawah.

| Endpoint / lokasi object | Object | Aksi |
|--------------------------|--------|------|
| `GET /api/v1/mobile/profile/me` | data user (self) | + `is_verified` |
| `GET /api/v1/mobile/communities/:id/members` | `member.user` | + `is_verified` |
| `GET /api/v1/mobile/communities/:id` | `owner` | + `is_verified` |
| Post/komentar community | `author` | + `is_verified` |
| Event | `organizer` | + `is_verified` |
| Q&A, dan object user publik lain | user ref | + `is_verified` |

> Shared "user ref" object saat ini: `{ id, name, avatar, avatar_raw }`. Tambah jadi `{ id, name, avatar, avatar_raw, is_verified }`. Satu perubahan di serializer bersama = semua tempat kebadge.

---

## 3. Contoh Before/After

**Merchant list item:**
```jsonc
// SEBELUM
{ "id": "...", "name": "Seoul Galbi House", "type": "food_beverage", "rating": 0, "is_saved": false }
// SESUDAH
{ "id": "...", "name": "Seoul Galbi House", "type": "food_beverage", "rating": 0, "is_saved": false,
  "is_verified": true }
```

**Community list item:**
```jsonc
{ "id": "...", "name": "Bisnis Korea-Indonesia", "visibility": "public", "member_count": 6,
  "is_member": true, "is_verified": true }
```

**Shared user ref (author/owner/member.user/organizer):**
```jsonc
{ "id": "...", "name": "Test Member", "avatar": "https://…", "is_verified": true }
```

---

## 4. Follow-up Frontend (setelah field live)

Perubahan backend di atas adalah **prasyarat** — app tidak bisa menampilkan badge sebelum field ada. Setelah live, sisi mobile perlu:

1. **Model/entity:** tambah `bool isVerified` (default `false`) di `Merchant`, `Community`, dan user ref (`UserRef`), parse dari `is_verified` di masing-masing `fromJson`. Aman untuk di-deploy lebih dulu — default `false` selama field belum dikirim backend.
2. **UI badge:** tampilkan ikon centang di sebelah nama pada: kartu merchant (list + detail), kartu community (list + detail), nama user (profil, author post/komentar, member list, owner community). Sediakan satu widget `VerifiedBadge` reusable.
3. **Jangan** pakai `email_verified`/`phone_verified`/`approval_status` sebagai proxy badge — itu makna lain (§1).

---

## 5. Ringkasan

- **Field:** `is_verified` (boolean, wajib) + `verified_at` (opsional).
- **Belum ada** di response manapun per 2026-07-16 → backend wajib menambahkan di endpoint §2.
- **Sumber:** kolom cache `is_verified` (users/merchants/communities), sudah didefinisikan di DB Schema.
- **Frontend** menyusul setelah field live (§4).

---

*Verification Badge Exposure Spec v1.0 — KAI App. Pelengkap `API_SPEC_VERIFICATION_MOBILE.md` §5. Last updated: 2026-07-16.*
