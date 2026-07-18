# Merchant Like / Unlike — Spec (Mobile & Backoffice) v1.0

Spec fitur **Like/Unlike merchant** (❤️), terpisah dari fitur **Save/Bookmark** yang sudah ada.

> **Like ≠ Save.** *Save* (`is_saved`, `POST/DELETE /merchants/:id/save`, list `me/saved`) adalah bookmark pribadi untuk dibuka lagi nanti. *Like* adalah apresiasi publik yang **menaikkan `favorite_count`** merchant. Keduanya independen — satu merchant bisa di-save tanpa di-like, dan sebaliknya.

---

## 1. Model Data

Setiap object merchant (list & detail) menambahkan/menyertakan:

| Field | Type | Wajib | Arti |
|-------|------|:---:|------|
| `is_liked` | boolean | ✅ | `true` jika **user saat ini** sudah nge-like merchant ini. Default `false`. |
| `favorite_count` | integer | ✅ | Total like merchant (sudah ada di response saat ini; kini juga dipakai sebagai jumlah like). |

**Sumber kebenaran:** baris di tabel pivot `merchant_likes (user_id, merchant_id, created_at)`. `is_liked` = ada baris untuk (user, merchant). `favorite_count` = `COUNT(*)` per merchant (boleh di-cache kolom `favorite_count` di `merchants`, di-maintain saat like/unlike).

---

## 2. Endpoint (Mobile)

### 2.1 Like
- `POST /api/v1/mobile/directory/merchants/:id/like`
- **Auth:** Wajib (member)
- **Body:** kosong
- **Efek:** buat baris like kalau belum ada (idempotent — like dua kali tidak menggandakan), `favorite_count += 1`.
- **Response 200:**
  ```json
  { "data": { "is_liked": true, "favorite_count": 24 } }
  ```

### 2.2 Unlike
- `DELETE /api/v1/mobile/directory/merchants/:id/like`
- **Auth:** Wajib (member)
- **Efek:** hapus baris like kalau ada (idempotent), `favorite_count -= 1` (tidak pernah < 0).
- **Response 200:**
  ```json
  { "data": { "is_liked": false, "favorite_count": 23 } }
  ```

### 2.3 Error
| Status | Kode | Arti |
|--------|------|------|
| 401 | `ERR_UNAUTHORIZED` | Belum login |
| 404 | `NOT_FOUND` | Merchant tidak ada |
| 409 | (opsional) | Boleh diabaikan — like/unlike dibuat **idempotent**, jadi ulang aksi yang sama balas 200 tanpa error |

> Idempotent by design: FE pakai **optimistic toggle**, jadi endpoint tidak boleh error kalau state sudah sesuai (like saat sudah like → tetap 200).

---

## 3. Endpoint yang Wajib Meng-expose `is_liked`

Backend menambahkan `is_liked` (dan memastikan `favorite_count` terisi) di object merchant pada:

| Endpoint | Object |
|----------|--------|
| `GET /api/v1/mobile/directory/merchants` | tiap item list |
| `GET /api/v1/mobile/directory/merchants/:id` | detail |
| `GET /api/v1/mobile/directory/me/saved` | `merchant` di tiap entry (biar tombol like konsisten di list saved) |

> Owner-side (`me/merchants`, `manage`) tidak perlu — like adalah interaksi publik, bukan data kelola.

---

## 4. Perilaku Frontend (Mobile)

1. **Optimistic update:** saat tombol like ditekan, langsung balik state `is_liked` + `favorite_count` (±1) di UI, baru panggil endpoint. Kalau gagal → rollback ke nilai sebelumnya.
2. **Guard double-tap:** kunci tombol selama request in-flight (pola sama seperti `_savingBusy` di save).
3. **Tampilan:**
   - **Kartu merchant (list):** ikon hati (`favorite`/`favorite_border`) + `favorite_count` (kalau > 0).
   - **Detail merchant:** ikon hati + jumlah, terpisah dari tombol Save (bookmark).
4. **Guest:** kalau belum login, tekan like → arahkan ke login (atau sembunyikan) sesuai pola interaksi lain.

---

## 5. Backoffice
Tidak ada aksi khusus. `favorite_count` boleh ditampilkan sebagai metrik read-only di detail merchant (opsional). Superadmin tidak like/unlike atas nama user.

---

## 6. Checklist Implementasi

### Backend
- [ ] Tabel `merchant_likes` (atau reuse pivot favorite yang ada) + kolom cache `favorite_count`.
- [ ] `POST` & `DELETE /merchants/:id/like` (idempotent, transaksi update count).
- [ ] Tambah `is_liked` di serializer merchant (list, detail, saved).

### Mobile (Flutter)
- [ ] Entity/model: `bool isLiked` (parse `is_liked`, default false).
- [ ] Datasource: `like(id)` (POST), `unlike(id)` (DELETE).
- [ ] Repo: `toggleLike(id, like)`.
- [ ] UseCase `ToggleLikeMerchantUseCase` + registrasi DI.
- [ ] UI: tombol like optimistic di kartu & detail.

---

*Merchant Like/Unlike Spec v1.0 — KAI App. Pelengkap `DIRECTORY_API_SPEC_MOBILE_V2.md`. Last updated: 2026-07-17.*
