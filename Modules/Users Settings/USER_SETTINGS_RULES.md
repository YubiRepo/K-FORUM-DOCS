# User Settings — Rules & Use Cases

Dokumen ini menjelaskan aturan bisnis modul **User Settings**: pengaturan **per-user** yang dikelola sendiri oleh member lewat mobile app. Fokus pada **setting apa saja yang jadi milik modul ini**, **apa yang justru sudah punya rumah di modul lain**, dan **bagaimana setiap setting di-resolve** oleh backend. Untuk detail teknis lihat `USER_SETTINGS_DB_SCHEMA.md`.

**Status:** Draft v1
**Last Updated:** 2026-07-09
**Module:** User Settings

---

## Daftar Isi

1. [Overview Konsep](#overview-konsep)
2. [Prinsip Desain](#prinsip-desain)
3. [Apa yang BUKAN User Settings](#apa-yang-bukan-user-settings)
4. [Kategori Setting](#kategori-setting)
5. [App Preferences (Language & Theme)](#app-preferences-language--theme)
6. [Timezone (Device Metadata)](#timezone-device-metadata)
7. [Session / Device Management](#session--device-management)
8. [Linked Accounts (Google)](#linked-accounts-google)
9. [Data Export](#data-export)
10. [Storage Model](#storage-model)
11. [Endpoint Landing (GET /api/v1/user-settings)](#endpoint-landing)
12. [Siapa Bisa Apa](#siapa-bisa-apa)
13. [Use Cases](#use-cases)
14. [Ringkasan Aturan](#ringkasan-aturan)
15. [Cross-Module Follow-ups](#cross-module-follow-ups)
16. [Keputusan yang Sudah Dikunci](#keputusan-yang-sudah-dikunci)

---

## Overview Konsep

User Settings adalah **kumpulan preferensi & kontrol yang dimiliki dan diubah sendiri oleh satu user**, berbeda dari System Settings yang platform-wide dan dikelola admin. Semua setting di sini bersifat **self-scope**: user hanya bisa lihat & ubah settingnya sendiri.

Modul ini punya batas yang tricky karena banyak hal yang orang sebut "settings" sebenarnya **sudah tersebar di modul lain** (Profile, Auth, Notification Preferences, FCM). Supaya nggak duplikat, modul ini hanya memegang **yang belum punya rumah**, dan **mengagregasi** yang sudah ada.

Dua prinsip dasar:

- **Preference adalah pilihan user, bukan fakta device.** Kalau sebuah nilai punya satu jawaban benar yang device sudah tahu (mis. timezone), itu di-capture otomatis, bukan dijadikan picker.
- **Resolve live, jangan snapshot.** Nilai efektif (mis. bahasa yang dipakai) di-resolve dari chain fallback saat dibutuhkan, konsisten dengan pola ID Card & Member Point.

---

## Prinsip Desain

| Prinsip | Detail |
|---|---|
| **Self-scope only** | User hanya baca/ubah settingnya sendiri. Nggak ada admin yang ngedit User Settings orang lain (itu ranah User Management) |
| **Auto-create default** | Kalau user belum punya baris settings, server auto-create dengan default saat pertama diakses (sama seperti Notification Preferences) |
| **No duplication** | Setting yang sudah punya rumah di Profile/Auth/Notification/FCM TIDAK diulang di sini — dibaca/diagregasi, bukan disalin |
| **Device fact vs user choice** | Nilai deterministik dari OS (timezone) di-capture otomatis; nilai pilihan (language, theme) pakai picker |
| **Resolve live** | Bahasa & timezone efektif di-resolve dari chain, bukan di-cache mati |
| **Additive & hook-ready** | Field disiapkan buat Phase 2 (data export, 2FA) tanpa refactor schema |

---

## Apa yang BUKAN User Settings

Untuk mencegah duplikasi, hal-hal berikut **sudah punya rumah sendiri** dan tidak masuk modul ini:

| Yang sering dikira "settings" | Rumah aslinya |
|---|---|
| Edit nama, avatar, bio, alamat, dll | **Profile** (`API_SPEC_PROFILE_MEMBERSHIP`) |
| Ganti email (OTP) | **Profile** |
| Hapus akun / batal hapus | **Profile** |
| Ganti password (dari settings) | **Auth** (`POST /auth/password/change`) |
| Lupa password / reset | **Auth** |
| Logout & refresh token | **Auth** (`POST /auth/logout`, `POST /auth/refresh`) |
| Registry device + FCM token (list, revoke per device) | **FCM** (tabel `fcm_tokens`) |
| Toggle notifikasi, DND, per-modul, per-komunitas | **Notification Preferences** |
| Kelola/upgrade subscription | **Subscription** |
| Baca & setujui legal docs (T&C, Privacy) | **System Settings** (legal versioning) |
| Konfigurasi platform-wide | **System Settings** |

> **Catatan penting:** Registry device sudah ada di modul **FCM** (`fcm_tokens`: `device_id`, `platform`, `device_name`, `device_model`, `os_version`, `app_version`). Modul ini **TIDAK** bikin tabel session baru — Session Management di sini adalah **agregator/read-through** di atas `fcm_tokens` + endpoint Auth.

---

## Kategori Setting

Yang **beneran milik** modul ini vs yang **diagregasi**:

| Kategori | Isi | Milik modul ini? | User-facing (picker)? |
|---|---|---|---|
| **App Preferences** | Language, Theme | ✅ Ya (`user_settings`) | ✅ Ya |
| **Timezone** | IANA timezone device | Kolom di `fcm_tokens` (FCM) | ❌ Auto (metadata) |
| **Session / Device** | List device, logout | Agregasi FCM + Auth | ✅ Ya |
| **Linked Accounts** | Status Google, unlink | Agregasi Auth (`google_id`) | ✅ Ya |
| **Data Export** | Download data pribadi | ✅ Ya (Phase 2) | ✅ Ya (Phase 2) |

---

## App Preferences (Language & Theme)

### Language

Bahasa adalah **pilihan sadar**, bukan fakta device — krusial buat app diaspora (user Korea-Indonesia sering set HP ke `ko` tapi lebih nyaman baca app `id`, atau sebaliknya).

**Chain resolusi (dari prioritas tertinggi):**

1. **User preference** — kalau user sudah pernah pilih di settings
2. **Device language** — dari header `X-Locale`/`Accept-Language` yang dikirim client (default `ko`). Dipakai sebagai **default awal**, HANYA jika ada di `system_languages`
3. **`default_language`** dari System Settings — fallback terakhir

Aturan:

- Nilai valid dibatasi ke daftar aktif di `system_languages` (mis. `id`, `en`, `ko`). Kirim bahasa di luar daftar → ditolak.
- Begitu user override manual, device language **berhenti** jadi penentu — preference user menang, walau nanti dia ganti bahasa HP.
- Bahasa efektif dipakai **client-side** (UI copy) DAN **server-side** (template email, copy push notif, pemilihan bahasa legal docs kalau nanti multi-bahasa).

### Theme

| Value | Arti |
|---|---|
| `system` | Ikut setting OS (default) |
| `light` | Paksa terang |
| `dark` | Paksa gelap |

Theme murni client-side; backend menyimpan pilihannya saja supaya sinkron antar-device.

---

## Timezone (Device Metadata)

Timezone **bukan setting yang dipilih user** — device sudah tahu pasti dari OS, dan bikin picker manual malah rawan salah (user pindah negara, lupa update).

Aturan:

- App mengirim timezone device dalam **format IANA** (`Asia/Jakarta`, `Asia/Seoul`, `Asia/Makassar`, dll) secara otomatis saat **login**, **refresh token**, dan/atau **register/update FCM token**.
- **Disimpan sebagai kolom `timezone` di tabel `fcm_tokens`** (device row sudah ada di FCM) — **bukan** tabel baru. Ini penambahan kolom additive di modul FCM.
- **Tidak ada picker** di UI settings.
- Backend memakainya buat semua scheduling yang time-sensitive:
  - **DND** Notification (`22:00–08:00` itu jam lokal siapa?)
  - **Event reminder** (`reminder_hours_before`) kalau di-fire server-side
  - **Community Schedule** (RRULE recurrence butuh anchor timezone)
- Relevan banget buat diaspora: user bisa di WIB/WITA/WIT atau KST — nggak boleh di-hardcode ke satu zona.

> Kalau user aktif di banyak device dengan timezone beda, timezone efektif diambil dari **device yang paling terakhir aktif** (last-write-wins), atau device sumber request untuk aksi real-time.

---

## Session / Device Management

Menampilkan device mana saja yang aktif dan memungkinkan user memutus akses. **Modul ini tidak menyimpan session sendiri** — ini agregator di atas registry FCM + endpoint Auth.

**Sumber data device:** tabel `fcm_tokens` (FCM) — sudah berisi `device_id`, `platform`, `device_name`, `device_model`, `os_version`, `app_version`, plus `timezone` (kolom baru).

Dua aksi logout (final):

| Aksi | Mekanisme | Efek |
|---|---|---|
| **Logout per-device** | Revoke FCM token device tsb (`DELETE /fcm/revoke`) + invalidate session/refresh device itu di Auth | Device itu ke-logout & berhenti terima push. Device lain tetap aktif |
| **Logout all** | Auth revoke **semua** refresh token user + FCM revoke semua token user | **Semua** device ke-logout, termasuk yang sekarang. User harus login ulang |

Aturan:

- **List sessions** = query `fcm_tokens` milik user; tandai "device ini" berdasarkan `device_id` request.
- Enforcement logout bergantung ke **kapabilitas revoke di Auth**. Auth sudah punya `logout` & `refresh`; "logout all" memerlukan Auth bisa meng-invalidate seluruh refresh token user (revoke-all) — ini tanggung jawab modul Auth, User Settings hanya memanggilnya (lihat Cross-Module Follow-ups).
- Logout **tidak** menghapus data user, cuma memutus akses device.

---

## Linked Accounts (Google)

Menampilkan & mengelola akun sosial yang tertaut. Saat ini hanya **Google** (`google_id` sudah ada di Auth).

| Aksi | Aturan |
|---|---|
| **Lihat status** | Tampilkan apakah Google tertaut + email Google-nya |
| **Unlink** | **Hanya boleh** kalau user punya password aktif. Kalau akun cuma bisa login via Google (belum pernah set password), unlink ditolak biar user nggak ngunci diri sendiri |
| **Link** | **Ditunda** — nyusul kalau flow-nya sudah ada di Auth |

- Unlink yang ditolak → arahkan user set password dulu (via Auth) baru unlink.
- Phase 1 minimal: lihat status + unlink (dengan guard di atas).

---

## Data Export

**Phase 2.** User request download seluruh data pribadinya (profil, aktivitas, dll) — biasanya untuk kepatuhan privasi.

Hook yang disiapkan sekarang: endpoint & status tracking (`requested` → `processing` → `ready` → `expired`). Generasi file aktual + storage ditunda ke Phase 2. Nggak ada perubahan schema besar saat mengaktifkannya nanti.

---

## Storage Model

Modul ini **hanya memiliki satu tabel**; sisanya agregasi:

| Data | Model | Milik | Alasan |
|---|---|---|---|
| App Preferences (language, theme) | **Satu baris per user** (`user_settings`) | **User Settings** | Set-nya tetap & kecil; kolom eksplisit lebih simpel & type-safe |
| Device list + timezone | Kolom di `fcm_tokens` | **FCM** (nambah kolom `timezone`) | Registry device sudah ada; jangan duplikat |
| Session revoke | Endpoint | **Auth** | Sudah punya logout/refresh |
| Data Export job | Tabel job | User Settings (Phase 2) | Status tracking async |

Aturan:

- Baris `user_settings` **auto-created dengan default** saat pertama diakses; sistem tetap jalan walau user belum pernah buka settings.
- `user_settings` **hanya menyimpan** `language` + `theme` (+ hook Phase 2). Timezone, FCM token, device metadata **tidak** di sini — ada di `fcm_tokens`.
- Semua nilai punya default runtime-safe.

---

<a id="endpoint-landing"></a>
## Endpoint Landing (GET /api/v1/user-settings)

Selain endpoint granular per-kategori, ada **satu endpoint agregator** `GET /api/v1/user-settings` buat layar Settings utama — biar client nggak fetch 4-5 kali pas buka halaman. Endpoint **universal** (melayani Flutter mobile + Nuxt backoffice), mengikuti pola `API_SPEC_SESSION_BOOTSTRAP`.

Response menggabungkan (read-only aggregation):

- `app_preferences` (language, theme) — dari `user_settings`
- `linked_accounts` (status Google) — dari Auth
- `devices` (ringkas: nama, platform, last active) — dari `fcm_tokens`

Update tetap lewat endpoint granular masing-masing (PUT preference, DELETE device, dll). Landing ini murni buat baca cepat.

---

## Siapa Bisa Apa

| Aktor | Akses |
|---|---|
| **Member (Standard & Pro)** | Baca & ubah **settingnya sendiri**. Semua kategori |
| **Guest** | Tidak punya settings (belum ada akun) |
| **Admin / Superadmin** | **Tidak** mengubah User Settings member lewat modul ini. Aksi admin atas user (suspend, reset password, ubah plan) ada di **User Management**, terpisah |

Beda Standard vs Pro **tidak** memengaruhi struktur settings — semua member punya set setting yang sama. (Konsisten: perbedaan plan lewat benefit, bukan role/fitur setting.)

---

## Use Cases

### Use Case 1 — Ganti Bahasa
User baru install, HP-nya bahasa Korea → app default ke `ko` (karena `ko` ada di `system_languages`). User lebih nyaman `id`, buka Settings → pilih Indonesia. Mulai sekarang app & email dari server pakai `id`, walau HP tetap Korea.

### Use Case 2 — Timezone Otomatis
User pindah kerja Jakarta → Seoul. Buka app, login → device kirim `Asia/Seoul`, tersimpan di `fcm_tokens.timezone`. DND `22:00–08:00`-nya otomatis ngikut jam Korea tanpa user ngapa-ngapain. Nggak ada picker yang perlu disentuh.

### Use Case 3 — HP Hilang
User kehilangan HP lama. Dari HP baru → Settings → Devices → lihat device lama masih aktif → Logout per-device. Auth invalidate session device itu, FCM revoke token-nya, device lama ke-logout & berhenti terima push. Device sekarang tetap aktif.

### Use Case 4 — Logout Semua
User curiga akunnya kebobolan. Pilih "Logout all" → semua refresh token di-revoke Auth + semua FCM token di-revoke, termasuk device sekarang. User login ulang di device yang dia pegang.

### Use Case 5 — Lepas Google
User daftar via Google, sekarang mau lepas. Sistem cek: belum punya password → unlink ditolak, diarahkan set password dulu (via Auth). Setelah set password, unlink berhasil.

---

## Ringkasan Aturan

| Aturan | Detail |
|---|---|
| **Scope** | Self-only; tiap user cuma settingnya sendiri |
| **Auto-create** | Baris `user_settings` dibuat dengan default saat pertama diakses |
| **No duplication** | Profil, password, email, notif, subscription, legal, **device registry** TETAP di modulnya |
| **Owns 1 table** | Cuma `user_settings` (language + theme); device pakai `fcm_tokens`, session pakai Auth |
| **Language** | Pilihan user; default awal dari device kalau ada di `system_languages`; fallback ke `default_language` |
| **Theme** | `system`/`light`/`dark`, default `system` |
| **Timezone** | Auto dari device (IANA), kolom di `fcm_tokens`, dipakai untuk scheduling. Tanpa picker |
| **Sessions** | Dua opsi: logout per-device & logout all (all termasuk device sekarang); enforcement di Auth |
| **Linked accounts** | Unlink Google hanya jika punya password; link flow ditunda |
| **Landing** | `GET /api/v1/user-settings` agregator read-only (universal) buat layar Settings utama |
| **Data export** | Phase 2, hook status disiapkan |

---

## Cross-Module Follow-ups

Item yang perlu dikerjakan di modul lain supaya modul ini nyambung:

- **FCM** — tambah kolom `timezone` (IANA) di `fcm_tokens`; simpan saat register/update token. Pastikan ada endpoint "list devices" yang bisa dipakai layar Settings.
- **Auth** — sediakan kapabilitas **revoke-all** (invalidate semua refresh token user) untuk fitur "Logout all". Logout per-device & refresh sudah ada.
- **Notification** — konsumen `timezone` (DND & reminder) dari `fcm_tokens`.
- **System Settings** — `system_languages` & `default_language` jadi sumber validasi & fallback bahasa.
- **Role-Permission** — kemungkinan **tidak** perlu permission baru (semua aksi self-scope, cukup autentikasi). Konfirmasi saat DB schema.

---

## Keputusan yang Sudah Dikunci

| # | Topik | Keputusan |
|---|---|---|
| 1 | Bentuk modul | **Hybrid** — domain sendiri (`user_settings`) + agregator; plus endpoint universal `GET /api/v1/user-settings` sebagai landing |
| 2 | Session Management | **Phase 1**. Auth sudah ada; device list baca dari `fcm_tokens` (FCM), logout lewat Auth. Tidak ada tabel `user_sessions` baru |
| 3 | Opsi logout | **Dua opsi**: logout per-device & logout all (all = semua device termasuk yang sekarang) |
| 4 | Linked accounts | **Unlink Google** dengan guard "harus punya password"; **link flow ditunda** |

---

*Dokumen ini hasil brainstorming kebutuhan User Settings. Skema database menyusul di `USER_SETTINGS_DB_SCHEMA.md`.*
