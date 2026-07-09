# User Settings — Database Schema (v1)

Database schema untuk modul **User Settings**. Prinsip inti: modul ini **hanya memiliki satu tabel** (`user_settings`) — device & session **tidak** disimpan di sini, melainkan diagregasi dari `fcm_tokens` (FCM) dan endpoint Auth. Untuk aturan bisnis lihat `USER_SETTINGS_RULES.md`.

**Status:** Draft v1
**Last Updated:** 2026-07-09
**Module:** User Settings

> **Konvensi tipe (disamakan dengan repo):** `users.id` & FK bertipe **`VARCHAR(36)`** (id string seperti `usr_90210`, bukan Postgres UUID), timestamp pakai **`TIMESTAMP`** — mengikuti tabel `fcm_tokens` yang jadi integrasi utama modul ini.

---

## Overview Relasi

```
users (dari modul Auth, id VARCHAR(36))
  ├── user_settings (1:1)              ← MILIK modul ini (language + theme)
  ├── fcm_tokens (1:N, dari FCM)       ← device list + timezone (kolom baru)
  └── refresh/session (Auth)           ← logout current (ada); logout-all (dependency)

system_languages (dari System Settings) ← sumber validasi & fallback bahasa
```

> Modul ini **tidak** membuat tabel `user_sessions` atau tabel device apa pun. Device registry sudah ada di `fcm_tokens`.

---

## 1. `user_settings`

Satu baris per user. Hanya menyimpan preferensi yang **beneran dipilih user** dan belum punya rumah di modul lain: `language` dan `theme`.

```sql
CREATE TABLE user_settings (
    user_id     VARCHAR(36) PRIMARY KEY,                 -- 1:1 dengan users.id
    language    VARCHAR(10) NULL,                        -- kode dari system_languages, mis. 'id','en','ko'. NULL = belum dipilih user (resolve live)
    theme       VARCHAR(10) NOT NULL DEFAULT 'system',   -- 'system' | 'light' | 'dark'
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_settings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_settings_theme CHECK (theme IN ('system', 'light', 'dark'))
);
```

### Catatan field

- **`language` NULLABLE** — ini kunci pola *resolve live*. `NULL` berarti user belum pernah override; backend me-resolve efektifnya dari chain **user → header `X-Locale`/`Accept-Language` → `default_language`** setiap kali dibutuhkan. Begitu user memilih eksplisit, nilainya diisi dan header tidak lagi jadi penentu.
- **`theme` NOT NULL default `system`** — theme selalu punya nilai konkret; `system` artinya ikut OS/browser.
- **Tidak ada `timezone` di sini** — timezone per-device, ada di `fcm_tokens` (lihat bagian 2).
- **Tidak ada kolom device/session** — diagregasi, bukan disimpan.

### Validasi di service layer

1. `language` (kalau diisi) harus **ada & aktif** di `system_languages`. Kalau tidak → tolak (422).
2. `theme` ∈ {`system`,`light`,`dark`}.
3. **Auto-create**: kalau baris belum ada saat diakses, insert default (`language = NULL`, `theme = 'system'`) lalu kembalikan.

---

## 2. Perubahan di `fcm_tokens` (modul FCM) — cross-module

Timezone device disimpan di registry device yang **sudah ada** (`/api/v1/mobile/fcm`), bukan tabel baru. Ini penambahan kolom **additive** yang dikerjakan di modul FCM:

```sql
-- Migration dijalankan di modul FCM
ALTER TABLE fcm_tokens
    ADD COLUMN IF NOT EXISTS timezone VARCHAR(64) NULL;   -- IANA tz, mis. 'Asia/Jakarta', 'Asia/Seoul'
```

Aturan:

- Diisi/diupdate saat **register** (`POST /api/v1/mobile/fcm/register`) & **update** (`PUT /api/v1/mobile/fcm/update`) — client menambahkan field `timezone` device pada body.
- Dibaca modul Notification untuk DND & reminder, dan modul Schedule untuk anchor RRULE.
- `last_used_at` / `updated_at` yang sudah ada dipakai untuk menentukan **device terakhir aktif** (sumber timezone efektif = last-write-wins).

> Tidak ada perubahan struktur lain di `fcm_tokens`. Kolom `id`, `device_name`, `platform`, `device_model`, `os_version`, `is_active`, `last_used_at` yang sudah ada langsung dipakai untuk layar Session/Device di Settings. **Device diidentifikasi lewat `fcm_tokens.id`** (bukan `fcm_token`, yang sensitif dan tidak di-expose).

---

## 3. `user_data_export_jobs` — Phase 2 (disiapkan, belum aktif)

Hook untuk fitur export data pribadi. **Belum dibangun di Phase 1**; dicantumkan agar aktivasi Phase 2 tidak butuh refactor.

```sql
-- PHASE 2 — jangan di-migrate dulu di Phase 1
CREATE TABLE user_data_export_jobs (
    id           VARCHAR(36) PRIMARY KEY,                   -- UUID digenerate app-side
    user_id      VARCHAR(36) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'requested',  -- requested|processing|ready|expired|failed
    file_url     TEXT NULL,                                 -- diisi saat ready
    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ready_at     TIMESTAMP NULL,
    expires_at   TIMESTAMP NULL,

    CONSTRAINT fk_export_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT chk_export_status CHECK (status IN ('requested','processing','ready','expired','failed'))
);
CREATE INDEX idx_export_user ON user_data_export_jobs (user_id);
```

---

## Go Structs

```go
// UserSettings — 1:1 dengan user. Language nullable (resolve live).
type UserSettings struct {
    UserID    string    `json:"user_id"    db:"user_id"`
    Language  *string   `json:"language"   db:"language"` // nil = belum dipilih user
    Theme     string    `json:"theme"      db:"theme"`    // system|light|dark
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DeviceView — bukan tabel; hasil query dari fcm_tokens untuk layar Settings.
type DeviceView struct {
    ID         string     `json:"id"           db:"id"`          // fcm_tokens.id — identifier device di Settings
    Platform   string     `json:"platform"     db:"platform"`    // android|ios|web
    DeviceName string     `json:"device_name"  db:"device_name"`
    OSVersion  string     `json:"os_version"   db:"os_version"`
    Timezone   *string    `json:"timezone"     db:"timezone"`
    LastUsedAt *time.Time `json:"last_used_at" db:"last_used_at"`
    IsCurrent  bool       `json:"is_current"`  // dihitung di service dari device request
}
```

---

## Query Patterns

### Ambil settings (auto-create kalau belum ada)

```sql
INSERT INTO user_settings (user_id) VALUES ($1)
ON CONFLICT (user_id) DO NOTHING;

SELECT user_id, language, theme, created_at, updated_at
FROM user_settings WHERE user_id = $1;
```

### Resolve bahasa efektif (chain: user → header → default)

```sql
-- 1. Preference user
SELECT language FROM user_settings WHERE user_id = $1;   -- bisa NULL

-- 2. Kalau NULL: pakai header X-Locale / Accept-Language, cek apakah aktif di system_languages
SELECT code FROM system_languages WHERE code = $header_locale AND is_active = true;

-- 3. Kalau tetap kosong: default_language dari system_settings (dibaca dari cache; default 'ko')
```

### Update preference

```sql
UPDATE user_settings
SET language = $2, theme = $3, updated_at = NOW()
WHERE user_id = $1;
-- Service validasi dulu: language ∈ system_languages (jika non-null), theme ∈ enum.
```

### List device untuk layar Settings (dari fcm_tokens)

```sql
SELECT id, platform, device_name, os_version, timezone, last_used_at
FROM fcm_tokens
WHERE user_id = $1 AND is_active = true
ORDER BY last_used_at DESC NULLS LAST;
```

### Logout per-device

```sql
-- FCM: nonaktifkan token device tsb (identifikasi via id)
UPDATE fcm_tokens SET is_active = false, updated_at = NOW()
WHERE user_id = $1 AND id = $2;
-- Auth: invalidate session/refresh device tsb — HANYA jika Auth sudah punya kemampuan
-- revoke per-device. Saat ini Auth baru punya logout (session sekarang). Lihat Cross-Module.
```

### Logout all

```sql
-- FCM: nonaktifkan semua token user
UPDATE fcm_tokens SET is_active = false, updated_at = NOW()
WHERE user_id = $1;
-- Auth: revoke SEMUA refresh token user (termasuk device sekarang) — DEPENDENCY:
-- butuh session/refresh store di Auth (belum ada). Lihat Cross-Module Follow-ups.
```

---

## Migration Order

1. `CREATE TABLE user_settings` (+ constraint + FK).
2. **Koordinasi modul FCM**: `ALTER TABLE fcm_tokens ADD COLUMN timezone` — kerjakan sebagai migration di modul FCM, bukan di sini.
3. `user_data_export_jobs` **ditunda** sampai Phase 2.

> Tidak ada seed data. `user_settings` diisi lazily per user (auto-create saat pertama diakses).

---

*Skema ini pasangan dari `USER_SETTINGS_RULES.md`. API di `API_SPEC_USER_SETTINGS.md` (universal, melayani Flutter + Nuxt).*
