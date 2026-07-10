# Spec: Perbaikan App Store Rejection — iOS 0.1.1 (7)

- **Submission ID**: eb31eed7-8c1a-4862-96c5-079dd5a3e274
- **Review date**: 10 Juli 2026 (device review: iPad Air 11-inch M3)
- **Status dokumen**: living doc untuk resubmission berikutnya

Apple menolak dengan **4 guideline**. Ringkasan status:

| # | Guideline | Kebutuhan | Mobile (Flutter) | Backend | Ops/ASC |
|---|-----------|-----------|------------------|---------|---------|
| 1 | 4.8 Login Services | Sign in with Apple | ✅ Selesai | ❌ Endpoint baru | Enable capability di App ID |
| 2 | 1.2 UGC | Block user | ❌ Belum | ❌ Endpoint baru | Screen recording |
| 3 | 2.1 Information Needed | Akun demo subscription expired | — | ❌ Siapkan akun | Isi App Review Information |
| 4 | 5.1.1(v) Account Deletion | Hapus akun in-app | ✅ Selesai | ✅ Sudah ada | Screen recording |

---

## 1. Guideline 4.8 — Sign in with Apple

### Yang sudah dikerjakan (Flutter) ✅

- Package `sign_in_with_apple: ^8.1.0` + entitlement `com.apple.developer.applesignin` di `ios/Runner/Runner.entitlements`.
- Tombol **Sign in with Apple** di layar Login & Register (hanya tampil di iOS), di atas tombol Google.
- Alur lengkap: `AuthStateProvider.loginWithApple()` → `AppleLoginUseCase` → `AuthRepository.loginWithApple()` → `POST /mobile/auth/login/apple`.
- String lokal 3 bahasa (`loginWithApple`).

### Yang harus dikerjakan backend ❌

**Endpoint baru**: `POST /api/v1/mobile/auth/login/apple` — unified login/register (sama seperti `login/google`).

Request:
```json
{
  "identity_token": "<JWT dari Apple>",
  "authorization_code": "<authorization code dari Apple>",
  "fullname": "Nama User",     // HANYA terkirim di login pertama; opsional
  "email": "user@icloud.com"   // HANYA terkirim di login pertama; opsional
}
```

Langkah verifikasi server (wajib, jangan percaya payload mentah):
1. Ambil public keys Apple dari `https://appleid.apple.com/auth/keys` (JWKS, boleh di-cache).
2. Verifikasi signature `identity_token` (RS256) dengan key yang `kid`-nya cocok.
3. Validasi claim: `iss == "https://appleid.apple.com"`, `aud == "com.yubitech.kforum.ios"` (bundle ID), `exp` belum lewat.
4. Ekstrak `sub` → ini **Apple user ID permanen**. Simpan sebagai kolom `apple_id` di tabel user (sejajar dengan `google_id`).
5. Lookup user by `apple_id`; kalau tidak ada, coba link by `email` (kalau ada di token); kalau tetap tidak ada → buat akun baru memakai `fullname`/`email` dari request body.
6. Response: **format sama persis dengan `login/google`** (`token`, `refresh_token`, `data` user) supaya mobile tidak perlu parsing khusus.

Catatan penting:
- Email dari Apple bisa berupa **private relay** (`xxx@privaterelay.appleid.com`) — perlakukan sebagai email valid.
- `fullname`/`email` hanya dikirim Apple **sekali** (login pertama). Jika hilang (user pernah revoke lalu login lagi), fallback: nama = "Apple User", email dari claim token jika ada.
- User yang dibuat via Apple **tidak punya password** → relevan untuk flow hapus akun (lihat §4, catatan password).

### Konfigurasi Apple Developer (ops)

- Certificates, Identifiers & Profiles → **Identifiers** → App ID `com.yubitech.kforum.ios` → centang capability **Sign In with Apple** → Save → regenerate provisioning profile (automatic signing di Xcode cukup dibuka ulang). Tanpa ini sign-in sheet error (code 1000) di device asli.
- **Keys** → Create Key → centang "Sign in with Apple" → download **`.p8`** (⚠️ hanya bisa didownload SEKALI) → catat **Key ID** + **Team ID** → serahkan ke tim backend (dipakai untuk client secret & revoke, lihat bawah).
- Tidak perlu Services ID / redirect URL selama hanya native iOS (bukan web/Android).
- Update screenshot metadata di ASC agar menampilkan layar login baru (diminta eksplisit oleh reviewer).
- Testing: device/simulator harus login Apple ID di Settings; tes final di device asli/TestFlight (simulator sering flaky).

### Kredensial untuk backend (status: ✅ sudah dibuat, 10 Jul 2026)

| Item | Nilai | Dipakai untuk |
|---|---|---|
| Team ID (`iss` di client secret) | `6HF6B53G32` | Client secret JWT |
| Key ID (header `kid` di client secret) | `W6WT543ZNX` | Client secret JWT |
| Client ID / Bundle ID (`sub` di client secret; `aud` saat verifikasi identity token) | `com.yubitech.kforum.ios` | Verifikasi + client secret |
| Private key | `AuthKey_W6WT543ZNX.p8` | Tanda tangan client secret (ES256) |
| JWKS Apple (public keys, tanpa kredensial) | `https://appleid.apple.com/auth/keys` | Verifikasi identity token |

⚠️ File `.p8` adalah private key: **tidak di-commit ke repo dan tidak ditempel di doc/chat** — diserahkan terpisah via password manager/secret manager. Team ID, Key ID, dan bundle ID bukan rahasia (tercantum di binary app), aman didokumentasikan di sini.

Catatan: verifikasi `identity_token` di endpoint login **hanya butuh JWKS publik** — `.p8` baru diperlukan untuk exchange `authorization_code` dan revoke token (bagian berikut).

### Revoke token Apple saat hapus akun (backend, WAJIB)

Kebijakan Apple (sejak Juni 2022): app dengan Sign in with Apple **wajib me-revoke token Apple ketika user menghapus akun** — nyambung dengan guideline 5.1.1 di §4.

1. **Client secret**: JWT ES256 ditandatangani `.p8`; claims: `iss` = Team ID, `sub` = `com.yubitech.kforum.ios`, `aud` = `https://appleid.apple.com`, `exp` ≤ 6 bulan. Generate on-the-fly atau cache.
2. **Saat login Apple pertama**: tukar `authorization_code` (sudah dikirim mobile di payload) via `POST https://appleid.apple.com/auth/token` (`grant_type=authorization_code`) → simpan `refresh_token` Apple di record user. Catatan: authorization code hanya valid ±5 menit, tukar segera.
3. **Saat hapus akun** (confirm-delete) untuk user ber-`apple_id`: `POST https://appleid.apple.com/auth/revoke` dengan refresh token tsb (`token_type_hint=refresh_token`), best-effort (jangan gagalkan penghapusan bila revoke error, tapi log).

---

## 2. Guideline 1.2 — UGC: Block User

Reviewer minta spesifik: *"a mechanism for users to block abusive users (blocking should also notify the developer of the inappropriate content and should remove it from the user's feed instantly)"*.

Yang **sudah ada** (tinggal didemokan di screen recording):
- ✅ EULA/ToS sebelum login — flow `pending_acceptances` (terms, privacy, community guidelines) sudah jalan via `accept_legal_document_usecase`.
- ✅ Report/flag konten — `showReportContentSheet` dipakai di 9 titik (post, komentar, event, dsb).

Yang **belum ada** — Block User:

### Backend ❌ (endpoint baru) — kontrak detail

**Mobile sudah terintegrasi dengan kontrak ini** (lihat bagian Mobile di bawah) — backend tinggal implement. Semua response memakai envelope standar `{status, status_code, message, data, meta}`.

#### `POST /api/v1/mobile/users/{user_id}/block`
- Auth: Bearer. Body (opsional):
  ```json
  { "reason": "string | null" }
  ```
- Response `200`:
  ```json
  { "status": "success", "message": "user blocked", "data": { "blocked_user_id": "uuid", "blocked_at": "ISO8601" } }
  ```
- Error: `400` block diri sendiri; `404` user tidak ada; `409` sudah diblok (idempotent-friendly: boleh juga balas 200).
- Efek server-side (WAJIB, ini yang dinilai reviewer):
  1. Konten user yang diblok **disaring dari semua endpoint list** milik si pemblokir: community posts, comments, Q&A, replies. (Event/news/merchant tidak wajib — bukan feed UGC bebas.)
  2. **Notifikasi ke admin/developer**: entry di backoffice moderation queue ATAU email ke support@kai.or.id, memuat blocker, blocked, reason, timestamp — reviewer eksplisit minta "blocking should also notify the developer".

#### `DELETE /api/v1/mobile/users/{user_id}/block`
- Response `200`: `{ "data": { "unblocked_user_id": "uuid" } }`. Error `404` bila tidak sedang diblok (atau 200 idempotent).

#### `GET /api/v1/mobile/users/me/blocked?page=1&limit=20`
- Response `200`:
  ```json
  {
    "data": [
      { "id": "uuid", "username": "string", "fullname": "string | null", "avatar": "url | null", "blocked_at": "ISO8601" }
    ],
    "meta": { "current_page": 1, "total_pages": 1, "total_items": 3, "items_per_page": 20 }
  }
  ```

### Mobile ✅ (sudah dikerjakan — jalan penuh setelah backend live)

- Aksi **"Block user"** di menu yang sama dengan "Report" pada post card & komentar (fitur `lib/features/blocking/`).
- Dialog konfirmasi → panggil endpoint → konten user tsb **hilang seketika dari list yang sedang tampil** (filter client-side via `BlockedUsersProvider`, tidak menunggu refresh — reviewer menekankan "instantly"). Daftar ID user terblokir juga di-persist lokal (SharedPreferences) sehingga filter tetap aktif di feed berikutnya meski backend belum menyaring.
- Halaman **Profile → Blocked Users** (route `/profile/blocked-users`) untuk melihat daftar & unblock.
- Fallback: jika endpoint belum tersedia (404), mobile tetap memblokir secara lokal (optimistic) supaya UX bisa didemokan — tapi **jangan resubmit sebelum backend live**, karena filter lintas-device & notifikasi admin hanya bisa dari server.

---

## 3. Guideline 2.1 — Akun Demo Subscription Expired

**Bukan bug dan bukan perubahan kode.** Reviewer tidak bisa mengakses flow pembelian subscription karena tidak punya akun dengan subscription **expired**.

Tugas (backend/ops):
1. Buat akun demo di production, contoh: `appreview_expired@k-forum.dev` / password bebas yang kuat.
2. Set akun tsb punya subscription **PRO yang sudah expired** (end date di masa lalu) supaya reviewer melihat flow renew/upgrade/purchase dari kondisi expired.
3. Isi kredensial di **App Store Connect → App Review Information → Sign-In Information** (bersama akun demo member biasa yang sudah ada).

### Keputusan model subscription (final)

**Model "multiplatform services" (guideline 3.1.3(b))** — purchase/renewal/upgrade hanya di web; app mobile hanya menampilkan status:

- ✅ App menampilkan: plan aktif, benefit, periode, sisa hari (layar Subscription).
- ✅ Purchase flow sudah tergerbang `AppFlags.inAppUpgradeEnabled = false` (compile-time) — tombol upgrade/beli tidak pernah tampil, termasuk di build yang direview.
- ✅ Harga plan di layar "Compare Plans" kini juga disembunyikan saat flag off (menampilkan harga tier yang tak bisa dibeli in-app berisiko dianggap advertising external purchase / 3.1.1). Perbandingan benefit tetap tampil.
- ❌ **Jangan** menambahkan teks/link "beli di website" di app iOS — anti-steering (kecuali storefront AS, tidak relevan untuk kita).
- Web tetap jualan seperti biasa; app tidak menyebutnya.

**Balasan ke reviewer untuk 2.1**: jelaskan bahwa subscription tidak dapat dibeli di dalam app iOS (tidak ada purchase flow); app hanya menampilkan status plan yang dikelola di luar app. Tetap sediakan akun demo expired di atas agar reviewer bisa memverifikasi bahwa dari kondisi expired pun tidak ada jalur pembelian.

---

## 4. Guideline 5.1.1(v) — Hapus Akun In-App

### Yang sudah dikerjakan ✅

- Menu **Hapus Akun** (merah) di Profile → bagian Account.
- Screen `/profile/delete-account` (`delete_account_screen.dart`), 2 langkah:
  1. Warning permanen + input password + alasan (opsional) → `POST /mobile/profile/account/delete-request` → OTP terkirim ke email.
  2. Input OTP → `POST /mobile/profile/account/confirm-delete` → sukses → auto-logout → kembali ke login.
- Backend endpoint-nya **sudah ada** dari awal; yang hilang hanya UI.

### Kontrak: hapus akun untuk akun social-login (tanpa password) ⚠️

**Mobile sudah terintegrasi** dengan kontrak berikut — backend tinggal implement:

1. **Payload user** (`/mobile/auth/login*`, `/mobile/auth/me`, `/mobile/profile/me`) menambah dua field:
   ```json
   {
     "apple_id": "string | null",
     "has_password": true
   }
   ```
   - `has_password`: `false` untuk akun yang dibuat via Google/Apple dan belum pernah set password lokal; `true` selainnya.
   - Perilaku mobile: jika `has_password == false`, layar Hapus Akun **menyembunyikan field password** dan mengirim `delete-request` **tanpa** field `password`. Jika field `has_password` tidak dikirim backend (build lama), mobile fallback ke `true` (password tetap diminta — perilaku lama).

2. **`POST /mobile/profile/account/delete-request`**: field `password` menjadi **opsional**.
   - `password` terkirim → validasi seperti sekarang.
   - `password` absen → hanya boleh diterima jika akun memang `has_password == false` (verifikasi identitas cukup via OTP email yang sudah ada di step confirm). Jika akun punya password tapi request tanpa password → `400`.

3. Pastikan penghapusan adalah **deletion** (data dihapus/dianonimkan sesuai kebijakan), bukan sekadar deactivate — reviewer eksplisit menolak "temporarily deactivate or disable".

---

## 5. Checklist Resubmission

**Kode/infra:**
- [ ] Backend: `POST /mobile/auth/login/apple` live di production
- [ ] Backend: block-user endpoints + filter feed + notifikasi admin
- [ ] Backend: `has_password` + `apple_id` di payload user, dan `password` opsional di `delete-request` (kontrak §4 — mobile sudah terintegrasi)
- [ ] Mobile: UI block user + halaman blocked users
- [ ] Ops: enable Sign In with Apple di App ID + provisioning profile baru
- [ ] Ops: buat Key "Sign in with Apple" (.p8) → serahkan Key ID + Team ID + .p8 ke backend
- [ ] Backend: exchange `authorization_code` → simpan Apple refresh token; revoke saat hapus akun
- [ ] Ops: akun demo subscription expired dibuat & diisi di App Review Information
- [ ] ASC: update screenshot metadata (login screen baru dengan tombol Apple)

**Bukti untuk reviewer (direkam di device fisik, taruh link di Notes App Review Information):**
- [ ] Recording 1 (guideline 1.2): tampilkan EULA/ToS saat register/login → flag/report sebuah konten → block seorang user → tunjukkan kontennya hilang seketika dari feed
- [ ] Recording 2 (guideline 5.1.1): buat akun baru / login akun demo → navigasi ke Profile → Hapus Akun → flow lengkap sampai konfirmasi & logout

**Balasan pesan reviewer:** setelah semua siap, reply di ASC dengan ringkasan per guideline + di mana menemukan tiap fitur.
