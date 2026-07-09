# API Spec — User Settings (Universal)

Spesifikasi API modul **User Settings**. **Satu endpoint set universal** yang melayani **Flutter mobile DAN Nuxt backoffice** — mengikuti pola `API_SPEC_SESSION_BOOTSTRAP` (`GET /api/v1/auth/session`). Datanya identik untuk kedua client (language, theme, devices, linked accounts), jadi tidak dipisah `/mobile/` vs `/web/`.

Semua setting bersifat **self-scope**: user hanya mengelola settingnya sendiri, apa pun role-nya. Aturan bisnis di `USER_SETTINGS_RULES.md`, skema di `USER_SETTINGS_DB_SCHEMA.md`.

- **Base URL**: `/api/v1/user-settings`
- **Dipakai oleh**: Flutter mobile + Nuxt backoffice
- **Auth**: semua endpoint butuh `Authorization: Bearer <access_token>`
- **Headers**:
  - `Content-Type: application/json`, `Accept: application/json`
  - `Accept-Language: <lang>` / `X-Locale: <lang>` (default `ko`) — dipakai untuk resolve bahasa efektif saat user belum override

> **Yang BUKAN di sini:** ganti password/email/hapus akun → **Profile/Auth**; toggle notifikasi & DND → **Notification Preferences**; register/refresh FCM token → **FCM** (`/api/v1/mobile/fcm`). Timezone **tidak** punya endpoint — dikirim otomatis saat register/update FCM token.
>
> **Hubungan dengan Session Bootstrap:** `GET /api/v1/auth/session` mengembalikan identitas/plan/permission setelah login. Modul ini melengkapi dengan **preferensi & kontrol akun** (bahasa, tema, device, linked account). Tidak tumpang tindih.

---

## Daftar Endpoint

| # | Method | Path | Fungsi |
|---|---|---|---|
| 1 | GET | `/user-settings` | Landing agregator (baca cepat) |
| 2 | GET | `/user-settings/preferences` | Ambil app preferences |
| 3 | PUT | `/user-settings/preferences` | Update language / theme |
| 4 | GET | `/user-settings/devices` | List device/session aktif |
| 5 | DELETE | `/user-settings/devices/{id}` | Logout per-device |
| 6 | POST | `/user-settings/devices/logout-all` | Logout semua device |
| 7 | GET | `/user-settings/linked-accounts` | Status akun tertaut (Google) |
| 8 | DELETE | `/user-settings/linked-accounts/google` | Unlink Google |

> `{id}` = `fcm_tokens.id` (identifier device di layar Settings). **Bukan** `fcm_token` — token tidak pernah di-expose.

---

## 1. GET /user-settings — Landing

Agregator read-only buat layar Settings utama, biar client nggak fetch 4-5 kali.

**Response `200`**
```json
{
  "data": {
    "app_preferences": {
      "language": "id",
      "language_effective": "id",
      "theme": "system"
    },
    "linked_accounts": {
      "google": { "linked": true, "email": "user.google@example.com", "can_unlink": true }
    },
    "devices": [
      {
        "id": "fcmrow_abc123",
        "platform": "android",
        "device_name": "Samsung Galaxy S21",
        "last_used_at": "2026-07-09T08:30:00.000Z",
        "is_current": true
      }
    ]
  }
}
```

- `language` = nilai tersimpan (`null` kalau belum dipilih user).
- `language_effective` = hasil resolve chain (user → header `X-Locale`/`Accept-Language` → `default_language`). Selalu terisi.
- `devices` bisa berisi campuran platform (`android`/`ios`/`web`) karena satu user bisa login di banyak client.

---

## 2. GET /user-settings/preferences

**Response `200`**
```json
{
  "data": {
    "language": null,
    "language_effective": "ko",
    "theme": "system",
    "available_languages": [
      { "code": "id", "name": "Bahasa Indonesia" },
      { "code": "en", "name": "English" },
      { "code": "ko", "name": "한국어" }
    ]
  }
}
```

> `available_languages` dari `system_languages` aktif (System Settings). Contoh `language_effective` = `ko` karena user belum override & header `X-Locale: ko`.

---

## 3. PUT /user-settings/preferences

Update language dan/atau theme. Kedua field opsional (partial update).

**Request**
```json
{ "language": "id", "theme": "dark" }
```

| Field | Type | Required | Aturan |
|---|---|---|---|
| `language` | `string`\|`null` | No | Harus ada di `system_languages` aktif. Kirim `null` untuk reset ke resolve-live (ikut header/default) |
| `theme` | `string` | No | `system` \| `light` \| `dark` (mengatur UI app maupun portal) |

**Response `200`**
```json
{ "data": { "language": "id", "language_effective": "id", "theme": "dark" } }
```

**Error `422`**
```json
{ "message": "Data input tidak valid", "errors": { "language": ["Bahasa 'jp' tidak tersedia."] } }
```

---

## 4. GET /user-settings/devices

List device/session aktif (sumber: `fcm_tokens`, `is_active = true`).

**Response `200`**
```json
{
  "data": [
    {
      "id": "fcmrow_abc123",
      "platform": "android",
      "device_name": "Samsung Galaxy S21",
      "os_version": "Android 13",
      "timezone": "Asia/Jakarta",
      "last_used_at": "2026-07-09T08:30:00.000Z",
      "is_current": true
    },
    {
      "id": "fcmrow_web01",
      "platform": "web",
      "device_name": "Chrome on macOS",
      "os_version": "macOS 14.5",
      "timezone": "Asia/Jakarta",
      "last_used_at": "2026-07-08T14:00:00.000Z",
      "is_current": false
    }
  ]
}
```

- `is_current` dihitung backend dari device request sekarang.

---

## 5. DELETE /user-settings/devices/{id} — Logout per-device

Cabut akses satu device. Device lain tetap aktif.

**Response `200`**
```json
{ "message": "Device berhasil di-logout" }
```

Efek Phase 1:
- **FCM**: token device di-nonaktifkan (`is_active = false`) → berhenti terima push.
- **Kalau device = device sekarang**: session di-akhiri via Auth `POST /auth/logout`.
- **Kalau device lain**: mematikan **auth session** device lain butuh kemampuan revoke per-device di Auth (belum ada). Sampai Auth siap, efek untuk device lain terbatas pada penghentian push + penonaktifan token; akses akhirnya putus saat access token device itu expired & refresh ditolak. Lihat catatan dependency di bawah.

**Error `404`** — device bukan milik user / tidak ditemukan.

---

## 6. POST /user-settings/devices/logout-all — Logout semua

Cabut akses **semua** device, **termasuk device sekarang**.

**Request**: none (Authorization header saja)

**Response `200`**
```json
{ "message": "Semua device telah di-logout. Silakan login kembali." }
```

Efek Phase 1:
- **FCM**: semua token user di-nonaktifkan.
- **Auth**: enforcement penuh butuh **revoke-all refresh token** di Auth — ini **dependency** (belum ada session/refresh store). Sampai tersedia, "logout all" memutus push untuk semua device & memicu logout device sekarang, tapi device lain baru benar-benar terputus saat refresh token berikutnya ditolak.

> ⚠️ **Dependency Auth:** fitur ini baru sepenuhnya enforce setelah modul Auth menambah session/refresh store + endpoint revoke-all. Sampai itu, kontrak endpoint tetap sama; backend mengembalikan `200` dan menjalankan bagian yang sudah bisa (FCM + logout current).

---

## 7. GET /user-settings/linked-accounts

**Response `200`**
```json
{
  "data": {
    "google": { "linked": true, "email": "user.google@example.com", "can_unlink": true }
  }
}
```

- `can_unlink = false` kalau user belum punya password (akun Google-only) — unlink ditolak sampai user set password.

---

## 8. DELETE /user-settings/linked-accounts/google — Unlink Google

**Response `200`**
```json
{ "message": "Akun Google berhasil dilepas" }
```

**Error `409`** — belum punya password
```json
{ "message": "Tidak bisa melepas Google karena akun belum punya password. Set password dulu lewat menu keamanan." }
```

> Link flow (menautkan Google baru) **belum tersedia di Phase 1**.

---

## Timezone (catatan, bukan endpoint)

Timezone **tidak** diatur lewat modul ini. Client mengirim IANA timezone (`Asia/Jakarta`, `Asia/Seoul`, dll) otomatis lewat body saat **register/update FCM token** (`/api/v1/mobile/fcm/register` & `/update`). Web (Nuxt) mengambilnya dari `Intl.DateTimeFormat().resolvedOptions().timeZone`. Disimpan di `fcm_tokens.timezone`, dipakai server-side untuk DND, reminder event, dan schedule.

---

## Authorization Notes

- Semua endpoint **self-scope** — user mengelola settingnya sendiri, apa pun role-nya. Tidak ada akses ke setting user lain (itu ranah **User Management**).
- Karena self-scope, **tidak perlu permission key khusus** — cukup terautentikasi.
- Data export (member-facing) ditunda ke **Phase 2**; endpoint belum diekspos.

---

## Error Responses (umum)

| Code | Arti |
|---|---|
| `401` | Token invalid/expired → refresh via Auth lalu retry |
| `404` | Resource tidak ditemukan (mis. device) |
| `409` | Konflik state (mis. unlink tanpa password) |
| `422` | Validation error (mis. bahasa tidak didukung) |
| `500` | Error backend |

---

## Cross-Module Dependencies

- **FCM** — kolom `timezone` di `fcm_tokens`; endpoint list device (identifikasi via `id`).
- **Auth** — **revoke-all** & **revoke per-device** untuk enforce "logout all" / logout device lain secara instan (butuh session/refresh store). Logout current & refresh sudah ada.
- **System Settings** — `system_languages` & `default_language` (default `ko`) untuk validasi & fallback bahasa.
- **Notification / Schedule** — konsumen `fcm_tokens.timezone`.

---

*Pasangan: `USER_SETTINGS_DB_SCHEMA.md`, `USER_SETTINGS_RULES.md`. Menggantikan draft terpisah Mobile/Backoffice.*
