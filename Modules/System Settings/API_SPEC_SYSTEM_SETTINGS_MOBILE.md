# API Spec — System Settings (Mobile)

**Status:** Draft v1
**Last Updated:** 2026-06-12
**Base URL Prefix:** `/api/v1/mobile`

Sisi mobile hanya **mengonsumsi** settings — tidak ada endpoint write kecuali pencatatan persetujuan legal. Referensi: `SYSTEM_SETTINGS_RULES.md`.

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Public Config](#1-get-public-config)
3. [Legal Documents](#legal-documents-endpoints)
4. [Penambahan Field di Endpoint Auth](#penambahan-field-di-endpoint-auth)
5. [Maintenance Mode Contract](#maintenance-mode-contract)
6. [App Version Check Flow (Client)](#app-version-check-flow-client)
7. [Important Notes](#important-notes)
8. [Error Handling](#error-handling)

---

## Informasi Umum

### Headers Global

```
Content-Type: application/json
Accept: application/json
Accept-Language: <lang_code> (e.g., id, en, ko)
Authorization: Bearer <access_token> (hanya endpoint yang ditandai auth)
```

### Authentication Matrix

| Endpoint | Auth |
|---|---|
| `GET /config` | ❌ Public |
| `GET /legal/{doc_type}` | ❌ Public |
| `GET /legal/pending` | ✅ Required |
| `POST /legal/{doc_type}/accept` | ✅ Required |

---

## 1. Get Public Config

Dipanggil setiap **app cold-start** (sebelum login). Hanya berisi setting `is_public = true`. Response di-cache server-side (TTL ±60 detik) — aman dipanggil sering.

- **URL**: `GET /api/v1/mobile/config`
- **Autentikasi**: No
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "general": {
        "app_name": "KAI App",
        "tagline": "Korea Asosiasi Indonesia",
        "support_email": "support@kai.or.id",
        "platform_url": "https://kai.or.id",
        "default_language": "id"
      },
      "mobile_app": {
        "min_version_android": "1.2.0",
        "min_version_ios": "1.2.0",
        "latest_version_android": "1.4.0",
        "latest_version_ios": "1.4.0",
        "force_update_enabled": true,
        "playstore_url": "https://play.google.com/store/apps/details?id=org.kai.app",
        "appstore_url": "https://apps.apple.com/app/id0000000000",
        "update_message": "Versi baru tersedia. Silakan perbarui aplikasi Anda."
      },
      "maintenance": {
        "maintenance_mode_enabled": false,
        "maintenance_message": ""
      },
      "payment": {
        "payment_provider": "manual",
        "bank_name": "BCA",
        "bank_account_number": "1234567890",
        "bank_account_holder": "Korea Asosiasi Indonesia",
        "payment_instructions": "1. Transfer ke rekening di atas...\n2. Upload bukti transfer...",
        "payment_confirmation_deadline_hours": 24
      },
      "contact": {
        "whatsapp_number": "+6281234567890",
        "instagram_url": "https://instagram.com/kai.official",
        "website_url": "https://kai.or.id"
      }
    }
  }
  ```
- **Headers Response**:
  ```
  Cache-Control: public, max-age=60
  ```

> Endpoint ini **selalu hidup** — termasuk saat maintenance mode aktif (app butuh membaca `maintenance_message`).

---

## Legal Documents Endpoints

### 2. Get Legal Document (Published)

Konten dokumen legal versi aktif — dipakai halaman registrasi (S&K + Privacy), halaman Settings → About, dan dialog re-acceptance.

- **URL**: `GET /api/v1/mobile/legal/{doc_type}`
- **Autentikasi**: No (guest perlu baca S&K sebelum daftar)
- **Path Parameter**: `doc_type` = `terms` | `privacy` | `community_guidelines`
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "doc_type": "terms",
      "title": "Syarat & Ketentuan",
      "version": "2.1.0",
      "version_id": "uuid",
      "effective_date": "2026-05-15",
      "content": "# Syarat dan Ketentuan Penggunaan KAI App\n\n**Terakhir diperbarui: ...**\n\n(markdown)",
      "published_at": "2026-05-01T10:00:00.000Z"
    }
  }
  ```
- **Response (Error 404)**: `doc_type` tidak dikenal, atau belum ada versi published.

> `content` dalam Markdown — render di Flutter dengan markdown renderer. `version_id` dipakai client saat memanggil endpoint accept.

---

### 3. Get Pending Acceptances

Daftar dokumen yang **wajib disetujui ulang** oleh user saat ini. Dipanggil saat **app open** (cold start dengan session aktif). Data yang sama juga disertakan di response login/refresh/me sebagai field `pending_acceptances` — lihat [Penambahan Field di Endpoint Auth](#penambahan-field-di-endpoint-auth).

- **URL**: `GET /api/v1/mobile/legal/pending`
- **Autentikasi**: Yes
- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "doc_type": "terms",
        "title": "Syarat & Ketentuan",
        "version": "2.2.0",
        "version_id": "uuid",
        "effective_date": "2026-07-01"
      }
    ]
  }
  ```
  > Array kosong `[]` = tidak ada yang perlu disetujui, lanjut normal.
- **Perilaku client**: jika array tidak kosong → tampilkan **blocking dialog** berisi dokumen (fetch content via endpoint #2), user harus tap "Setuju" sebelum melanjutkan ke aplikasi.

---

### 4. Accept Legal Document

Mencatat persetujuan user terhadap satu versi dokumen. Idempotent — accept versi yang sama dua kali tetap 200.

- **URL**: `POST /api/v1/mobile/legal/{doc_type}/accept`
- **Autentikasi**: Yes
- **Request Body**:
  ```json
  { "version_id": "uuid" }
  ```
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "doc_type": "terms",
      "version": "2.2.0",
      "accepted_at": "2026-06-12T09:30:00.000Z"
    },
    "message": "Terms accepted"
  }
  ```
- **Rules**:
  - `version_id` harus versi **published saat ini** untuk `doc_type` tersebut. Jika versi sudah ter-archive (admin publish versi lebih baru di antara fetch & accept) → `409` dengan versi terbaru di response, client refresh dialog:
    ```json
    {
      "message": "A newer version has been published",
      "data": { "current_version_id": "uuid", "current_version": "2.3.0" }
    }
    ```
  - Backend mencatat `source = "app"` dan IP request.
  - Saat **registrasi**, backend otomatis mencatat acceptance `terms` + `privacy` versi published saat itu (`source = "registration"`) — client tidak perlu memanggil endpoint ini setelah daftar, cukup menampilkan checkbox persetujuan di form register.

---

## Penambahan Field di Endpoint Auth

Modul ini **menambahkan satu field** pada response beberapa endpoint existing di `API_SPEC_AUTH.md` — agar app tahu ada dokumen yang wajib disetujui ulang **tanpa extra round-trip** ke `GET /legal/pending` saat momen login.

> Endpoint `GET /legal/pending` tetap dibutuhkan — dipanggil saat **app open** (user mobile jarang login ulang, token awet). Field di bawah hanya optimasi untuk momen login/refresh.

### Field Baru: `pending_acceptances`

Tipe: array. Kosong `[]` = tidak ada yang perlu disetujui. Struktur item sama dengan response `GET /legal/pending`:

```json
"pending_acceptances": [
  {
    "doc_type": "terms",
    "title": "Syarat & Ketentuan",
    "version": "2.2.0",
    "version_id": "uuid",
    "effective_date": "2026-07-01"
  }
]
```

### Endpoint yang Ketambahan Field

| Endpoint (`API_SPEC_AUTH.md`) | Posisi field | Keterangan |
|---|---|---|
| `POST /api/v1/mobile/auth/login` | Root response (sejajar `token`) | Login email/password |
| `POST /api/v1/mobile/auth/login/google` | Root response | Google OAuth — termasuk auto-register |
| `POST /api/v1/mobile/auth/refresh` | Root response | Menangkap user dengan session lama yang aktif saat versi baru di-publish |
| `GET /api/v1/mobile/auth/me` | Root response | Dipanggil saat app open — alternatif `GET /legal/pending` |

**Contoh response login setelah penambahan:**

```json
{
  "token": "access_token_jwt_string",
  "refresh_token": "refresh_token_jwt_string",
  "pending_acceptances": [
    { "doc_type": "terms", "title": "Syarat & Ketentuan", "version": "2.2.0", "version_id": "uuid", "effective_date": "2026-07-01" }
  ],
  "data": {
    "id": "usr_90211",
    "name": "Minji Park",
    "email": "minji@example.com",
    "plan": "standard",
    "roles": ["member"]
  }
}
```

### Perilaku Client

```
Response login / refresh / me diterima
        ↓
pending_acceptances kosong?
  ├── Ya  → lanjut normal
  └── Tidak → fetch content via GET /legal/{doc_type}
              → tampilkan blocking dialog
              → user setuju → POST /legal/{doc_type}/accept
              → semua selesai → lanjut ke app
```

> **Catatan backward-compatible:** field ini *additive* — client versi lama yang belum membaca field ini tidak rusak; mereka tetap ter-cover oleh pengecekan `GET /legal/pending` saat app open.

---

## Maintenance Mode Contract

Saat `maintenance_mode_enabled = true`, semua endpoint mobile (kecuali `GET /config` dan health check) merespons:

- **HTTP Status**: `503 Service Unavailable`
- **Body**:
  ```json
  {
    "maintenance": true,
    "message": "Kami sedang melakukan pemeliharaan sistem. Silakan coba beberapa saat lagi."
  }
  ```

**Perilaku client (Flutter):**

```
HTTP interceptor menangkap status 503 + body.maintenance == true
        ↓
Navigasi ke halaman Maintenance (full screen, pesan dari body.message)
        ↓
Tombol "Coba Lagi" → re-hit GET /config
        ├── maintenance_mode_enabled masih true → tetap di halaman
        └── false → restart flow normal
```

> User dengan role usergod/superadmin **bypass** di sisi backend (cek role dari token) — mereka tetap dapat response normal.

---

## App Version Check Flow (Client)

Dijalankan di splash screen setiap cold start, berdasarkan response `GET /config`:

```
current = versi app terpasang (semver)
min     = mobile_app.min_version_<platform>
latest  = mobile_app.latest_version_<platform>

compare(current, min) < 0 && force_update_enabled == true
        → Dialog BLOCKING "Wajib Update"
          (tidak bisa di-dismiss; tombol → playstore_url / appstore_url)

compare(current, min) >= 0 && compare(current, latest) < 0
        → Dialog DISMISSIBLE "Update Tersedia"
          (pesan dari update_message; tampilkan maks 1x per hari, simpan flag lokal)

compare(current, latest) >= 0
        → Lanjut normal
```

Aturan perbandingan: **semantic versioning** `major.minor.patch`, numerik per segmen (`1.10.0` > `1.9.0`).

---

## Important Notes

### ✅ DO:
- ✅ Panggil `GET /config` di setiap cold start — sebelum auth, sebelum routing
- ✅ Cache response config secara lokal sebagai fallback offline (tapi jangan pakai cache untuk keputusan force update jika fetch berhasil)
- ✅ Render konten legal sebagai Markdown
- ✅ Kirim `version_id` (bukan string versi) saat accept — presisi & race-safe
- ✅ Cek `pending_acceptances` dari response login/refresh/me dulu — fallback ke `GET /legal/pending` hanya saat app open tanpa request auth
- ✅ Tangani `409` pada accept dengan refresh dokumen lalu minta persetujuan ulang
- ✅ Tampilkan group `payment` dari config di halaman upgrade Pro (jangan hardcode rekening)
- ✅ Branch UI pembayaran dari `payment_provider`: `manual` → rekening + upload bukti; `midtrans` → Snap flow; `both` → user pilih. Saat ini selalu `manual` — siapkan branching-nya sejak awal agar aktivasi Midtrans nanti tanpa rilis app

### ❌ DON'T:
- ❌ Jangan hardcode rekening bank, nomor WA, atau store URL di app — semua dari `/config`
- ❌ Jangan tampilkan setting di luar response `/config` — yang tidak dikirim memang privat
- ❌ Jangan biarkan user menutup dialog re-acceptance tanpa menyetujui (blocking)
- ❌ Jangan panggil endpoint legal accept saat registrasi — backend sudah mencatat otomatis
- ❌ Jangan bandingkan versi sebagai string biasa (`"1.10.0" < "1.9.0"` salah) — pakai semver compare

---

## Error Handling

### Format Standard
```json
{ "message": "Pesan error deskriptif" }
```

### Skenario Umum

| Scenario | HTTP | Reason |
|----------|------|--------|
| `doc_type` tidak dikenal | 404 | Hanya terms/privacy/community_guidelines |
| Belum ada versi published | 404 | Dokumen belum pernah di-publish |
| Accept tanpa auth | 401 | Token required |
| Accept `version_id` bukan versi aktif | 409 | Versi lebih baru sudah published — refresh |
| `version_id` tidak ditemukan | 404 | Invalid version_id |
| Maintenance mode aktif | 503 | Body `{ maintenance: true, message }` |
| Accept duplikat (sudah pernah) | 200 | Idempotent — bukan error |

### Status Codes

`200` OK · `401` Unauthorized · `404` Not Found · `409` Conflict · `422` Unprocessable Entity · `503` Service Unavailable (maintenance) · `500` Internal Server Error

---

*API spec System Settings untuk mobile client (Flutter). Backoffice spec di `API_SPEC_SYSTEM_SETTINGS_BACKOFFICE.md`.*
