# System Settings — Rules & Use Cases

Dokumen ini menjelaskan aturan bisnis modul **System Settings**: konfigurasi global platform KAI yang dikelola lewat backoffice. Fokus pada **setting apa saja yang ada**, **siapa bisa mengubah apa**, dan **bagaimana setting dikonsumsi** oleh backend & mobile app. Untuk detail teknis lihat `SYSTEM_SETTINGS_DB_SCHEMA.md`.

**Status:** Draft v1
**Last Updated:** 2026-06-12
**Module:** System Settings

---

## Daftar Isi

1. [Overview Konsep](#overview-konsep)
2. [Prinsip Desain](#prinsip-desain)
3. [Apa yang BUKAN System Settings](#apa-yang-bukan-system-settings)
4. [Setting Groups](#setting-groups)
5. [Legal & Policies (Versioned Documents)](#legal--policies-versioned-documents)
6. [Siapa Bisa Apa](#siapa-bisa-apa)
7. [Public Config untuk Mobile](#public-config-untuk-mobile)
8. [Maintenance Mode](#maintenance-mode)
9. [Mobile App Version Control](#mobile-app-version-control)
10. [Caching & Invalidasi](#caching--invalidasi)
11. [Audit & Keamanan](#audit--keamanan)
12. [Use Cases](#use-cases)
13. [Ringkasan Aturan](#ringkasan-aturan)
14. [Keputusan yang Masih Terbuka](#keputusan-yang-masih-terbuka)

---

## Overview Konsep

System Settings adalah **satu tempat untuk konfigurasi platform-wide yang boleh berubah saat runtime tanpa deploy**. Terdiri dari dua bagian:

1. **Key-Value Settings** — nilai konfigurasi sederhana (toggle, angka, teks, list) yang dikelompokkan per *group*. Disimpan di tabel `system_settings`.
2. **Legal Documents** — dokumen legal (Syarat & Ketentuan, Kebijakan Privasi, Pedoman Komunitas) yang **versioned** dan butuh tracking persetujuan user. Disimpan di tabel terpisah karena strukturnya beda.

Dua prinsip dasar:

- **Setting adalah nilai, bukan logika.** Modul lain (Auth, Reporting, Subscription, dll) yang punya logikanya — System Settings hanya menyimpan angka/toggle yang dibaca modul tersebut.
- **Sesuai kebutuhan nyata.** Tidak ada setting untuk fitur yang belum ada. Tidak ada Public API key generator, webhook, atau theming — KAI adalah platform internal mobile-first.

---

## Prinsip Desain

| Prinsip | Detail |
|---|---|
| **Hybrid storage** | Key-value (`system_settings`) untuk nilai scalar; tabel relasional khusus untuk Legal Documents yang versioned |
| **Single source of truth** | Setting yang sudah punya rumah di modul lain TIDAK diduplikat di sini (lihat bagian berikut) |
| **Secrets bukan setting** | Kredensial (SMTP password, S3 secret key, Firebase key, Google OAuth secret) disimpan di environment variables / secret manager — BUKAN di database |
| **Typed value** | Setiap setting punya `value_type` (`string`/`number`/`boolean`/`array`/`json`) untuk validasi |
| **Runtime-safe default** | Setiap key punya default yang di-seed via migration; sistem tetap jalan walau admin belum menyentuh settings sama sekali |
| **Audited** | Setiap perubahan tercatat di Audit Log (old value → new value, siapa, kapan) |

---

## Apa yang BUKAN System Settings

Untuk mencegah duplikasi, hal-hal berikut **sudah punya rumah sendiri** dan tidak masuk modul ini:

| Konfigurasi | Rumahnya | Alasan |
|---|---|---|
| Translation toggle news (`translation_enabled`, `on_demand_enabled`) | `news_system_settings` (modul News) | Sudah exist, singleton sendiri |
| Bahasa platform (`system_languages`) | Modul News / Languages | Sudah exist, dikelola Usergod |
| Konfigurasi scraping per source | `news_sources` (modul News) | Per-source, bukan global |
| Ads setting (`approval_mode`, `max_active_ads_per_member`, dll) | `ad_settings` (modul Ads) | Sudah didefinisikan di `ADS_RULES.md` |
| Harga plan Standard/Pro | Tabel plans (modul Subscription) | Data master, bukan setting |
| Permission & role | Modul Role-Permission | Sistem sendiri |
| Notification preference | Modul Notification Preferences | Per-user, bukan global |
| Kredensial FCM / Google OAuth / S3 | Environment / secret manager | Keamanan |
| Theme/appearance backoffice | Frontend (localStorage per admin) | Preferensi personal, bukan platform |
| Public API + rate limit + webhook | ❌ Tidak dibuat | Tidak ada third-party consumer |
| Debug mode toggle dari UI | ❌ Tidak dibuat | Berbahaya di production; pakai env variable |

---

## Setting Groups

### Group: `general` — Platform Identity & Registrasi

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `app_name` | string | `KAI App` | Nama platform — dipakai di email, notifikasi, About page |
| `tagline` | string | `Korea Asosiasi Indonesia` | Subtitle/tagline |
| `support_email` | string | `support@kai.or.id` | Email support yang tampil ke user |
| `platform_url` | string | `https://kai.or.id` | URL utama platform |
| `default_timezone` | string | `Asia/Jakarta` | Timezone default untuk tampilan tanggal |
| `default_language` | string | `id` | Bahasa default UI — harus valid di `system_languages` |
| `public_registration_enabled` | boolean | `true` | `false` = endpoint register ditutup (hanya invite/admin-created) |
| `email_verification_required` | boolean | `true` | `true` = user wajib verifikasi OTP email sebelum bisa akses penuh (sesuai flow `API_SPEC_AUTH.md`) |

### Group: `mobile_app` — App Version Control

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `min_version_android` | string | `1.0.0` | Versi minimum yang masih didukung (semver) |
| `min_version_ios` | string | `1.0.0` | Versi minimum iOS |
| `latest_version_android` | string | `1.0.0` | Versi terbaru di Play Store — untuk soft-update prompt |
| `latest_version_ios` | string | `1.0.0` | Versi terbaru di App Store |
| `force_update_enabled` | boolean | `true` | `true` = app di bawah min version diblokir sampai update |
| `playstore_url` | string | `""` | Link Play Store |
| `appstore_url` | string | `""` | Link App Store |
| `update_message` | string | `Versi baru tersedia...` | Pesan yang tampil di dialog update |

### Group: `security` — Auth & Password Policy

Nilai-nilai ini **dibaca oleh modul Auth** — logika lockout/OTP/JWT tetap di Auth.

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `max_login_attempts` | number | `5` | Gagal login berturut-turut sebelum akun di-lock sementara |
| `lockout_duration_minutes` | number | `30` | Lama lockout setelah mencapai max attempts |
| `otp_expiry_minutes` | number | `5` | Masa berlaku kode OTP email |
| `otp_max_attempts` | number | `5` | Maksimal salah input OTP sebelum kode hangus |
| `otp_resend_cooldown_seconds` | number | `60` | Jeda minimal antar kirim ulang OTP |
| `access_token_expiry_minutes` | number | `60` | Masa berlaku JWT access token |
| `refresh_token_expiry_days` | number | `30` | Masa berlaku refresh token |
| `seamless_token_expiry_seconds` | number | `60` | TTL one-time token seamless login (spec: maks 1–2 menit) |
| `password_min_length` | number | `8` | Panjang minimum password |
| `password_require_number` | boolean | `true` | Wajib mengandung angka |
| `password_require_uppercase` | boolean | `false` | Wajib huruf kapital |
| `password_require_symbol` | boolean | `false` | Wajib simbol |

> **Catatan:** 2FA tidak masuk — belum ada di auth flow. Tambahkan setting-nya nanti bersamaan fiturnya.

### Group: `email` — SMTP & Email Behavior

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `smtp_host` | string | `""` | Host SMTP |
| `smtp_port` | number | `587` | Port SMTP |
| `smtp_username` | string | `""` | Username SMTP |
| `smtp_encryption` | string | `tls` | `tls` / `ssl` / `none` |
| `smtp_from_name` | string | `KAI App` | Nama pengirim |
| `smtp_from_email` | string | `noreply@kai.or.id` | Alamat pengirim |
| `welcome_email_enabled` | boolean | `true` | Kirim welcome email setelah registrasi terverifikasi |
| `email_footer_text` | string | `© 2026 KAI...` | Footer semua email keluar |

> **SMTP password TIDAK disimpan di DB** — di env (`SMTP_PASSWORD`). UI backoffice hanya menampilkan placeholder masked. Tombol "Send Test Email" memakai konfigurasi DB + password dari env.

### Group: `storage` — Upload & Media

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `storage_provider` | string | `local` | `local` / `s3` / `r2` — kredensial provider di env |
| `max_upload_size_mb` | number | `10` | Batas umum ukuran upload |
| `max_avatar_size_mb` | number | `2` | Batas khusus avatar |
| `max_payment_proof_size_mb` | number | `5` | Batas bukti transfer subscription |
| `allowed_mime_types` | array | `["image/jpeg","image/png","image/webp","application/pdf"]` | Whitelist MIME upload |
| `cdn_base_url` | string | `""` | Opsional — prefix CDN untuk serving file |

### Group: `payment` — Pembayaran Manual Subscription

Dibaca oleh modul Subscription dan ditampilkan di mobile saat user upgrade ke Pro.

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `bank_name` | string | `""` | Nama bank tujuan transfer, mis. `BCA` |
| `bank_account_number` | string | `""` | Nomor rekening |
| `bank_account_holder` | string | `""` | Nama pemilik rekening |
| `payment_instructions` | string | `""` | Instruksi pembayaran (markdown) — tampil di halaman upgrade |
| `payment_confirmation_deadline_hours` | number | `24` | Batas waktu upload bukti transfer setelah memulai upgrade |
| `payment_provider` | string | `manual` | `manual` / `midtrans` / `both` — menentukan flow pembayaran yang aktif. App branch UI dari nilai ini |
| `midtrans_environment` | string | `sandbox` | `sandbox` / `production`. Hanya dibaca backend saat `payment_provider` melibatkan midtrans |
| `midtrans_enabled_channels` | array | `[]` | Channel yang diaktifkan, mis. `["gopay","qris","bank_transfer"]`. Di-pass backend ke Snap |

> **Midtrans — disiapkan sejak awal:** tiga key terakhir sudah di-seed dengan default inert (`manual`, `sandbox`, `[]`) — tidak ada perilaku berubah sampai integrasi dibangun. Credentials (`MIDTRANS_SERVER_KEY`, `MIDTRANS_CLIENT_KEY`) tetap di env — pola sama dengan `storage_provider`. Logika transaksi (Snap token, webhook handler, status mapping) milik modul Subscription, bukan modul ini. UI backoffice menyembunyikan field `midtrans_*` saat `payment_provider = manual`.

### Group: `moderation` — Knob Global Reporting

`REPORTING_RULES.md` menandai dua nilai ini sebagai *"configurable"* tanpa menentukan rumahnya — rumahnya di sini. **Logika tetap di modul Reporting.**

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `banned_keywords` | array | `[]` | Kata terlarang — konten yang mengandung kata ini ditolak/di-flag saat submit (community post, comment, QnA) |
| `report_auto_flag_threshold` | number | `5` | Jumlah report sebelum target di-auto-flag (referensi: REPORTING_RULES) |
| `report_rate_limit_per_day` | number | `20` | Maksimal report per user per hari (referensi: REPORTING_RULES) |

### Group: `maintenance` — Maintenance Mode

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `maintenance_mode_enabled` | boolean | `false` | Lihat bagian [Maintenance Mode](#maintenance-mode) |
| `maintenance_message` | string | `Kami sedang melakukan pemeliharaan...` | Pesan yang tampil di mobile & backoffice |

### Group: `contact` — Kontak & Sosial Media

Untuk halaman About/Help di mobile app.

| Key | Type | Default | Keterangan |
|---|---|---|---|
| `whatsapp_number` | string | `""` | Nomor WhatsApp admin/CS |
| `instagram_url` | string | `""` | Link Instagram resmi |
| `website_url` | string | `""` | Website resmi |

---

## Legal & Policies (Versioned Documents)

### Konsep

Tiga jenis dokumen legal, masing-masing **versioned**:

| `doc_type` | Keterangan |
|---|---|
| `terms` | Syarat & Ketentuan |
| `privacy` | Kebijakan Privasi |
| `community_guidelines` | Pedoman Komunitas |

Struktur tiga tabel:

1. `legal_documents` — satu baris per jenis dokumen (master).
2. `legal_document_versions` — setiap versi konten (markdown). Status: `draft` → `published` → `archived`.
3. `user_legal_acceptances` — jejak persetujuan: user X menyetujui versi Y pada waktu Z.

### Aturan Versi

- Hanya **satu versi `published` per dokumen** pada satu waktu. Publish versi baru → versi lama otomatis `archived`.
- Versi memakai format semver-ish bebas (mis. `2.1.0`) + `effective_date`.
- Versi `published` **tidak bisa diedit** — perubahan = buat versi baru. (Typo kecil: buat versi patch.)
- Versi punya flag `require_reacceptance`:
  - `true` → setelah publish, semua user diminta menyetujui ulang saat login/buka app berikutnya (blocking dialog di mobile).
  - `false` → versi baru berlaku diam-diam, acceptance lama tetap dianggap sah.
- Acceptance dicatat sekali per user per versi (idempotent).
- Saat registrasi, user otomatis dianggap menyetujui versi `published` saat itu untuk `terms` & `privacy` (dicatat di `user_legal_acceptances`).

### Konsumsi di Mobile

- `GET /api/v1/mobile/legal/:doc_type` → versi published (markdown) — bisa diakses guest (untuk halaman registrasi).
- Saat login, response menyertakan `pending_acceptances: []` — daftar dokumen yang butuh persetujuan ulang.
- `POST /api/v1/mobile/legal/:doc_type/accept` → catat persetujuan versi aktif.

---

## Siapa Bisa Apa

Dua permission baru di master permission list:

| Key | Scope | Risk |
|---|---|---|
| `manage_system_settings` | global | high |
| `manage_legal_documents` | global | medium |

Matrix per group:

| Group | Usergod | Superadmin | Admin Region | Member |
|---|---|---|---|---|
| `general` | ✅ edit | ✅ edit | ❌ | ❌ |
| `mobile_app` | ✅ edit | ✅ edit | ❌ | ❌ |
| `security` | ✅ edit | 👁 read-only | ❌ | ❌ |
| `email` | ✅ edit | 👁 read-only | ❌ | ❌ |
| `storage` | ✅ edit | 👁 read-only | ❌ | ❌ |
| `payment` | ✅ edit | ✅ edit | ❌ | ❌ |
| `moderation` | ✅ edit | ✅ edit | ❌ | ❌ |
| `maintenance` | ✅ edit | ✅ edit | ❌ | ❌ |
| `contact` | ✅ edit | ✅ edit | ❌ | ❌ |
| Legal documents | ✅ full | ✅ full | ❌ | 👁 baca versi published |

Logikanya: **Usergod = setting teknis/infrastruktur** (security, email, storage); **Superadmin = setting operasional/konten** (identity, payment, legal, moderation, maintenance). Implementasi: kolom `editable_by` di tiap setting (`usergod` | `superadmin`), dicek di middleware selain permission check.

---

## Public Config untuk Mobile

Mobile app butuh sebagian setting **tanpa login** (cek versi, maintenance, kontak). Endpoint publik:

```
GET /api/v1/mobile/config
```

Hanya setting dengan flag `is_public = true` yang diekspos. Response dikelompokkan:

```json
{
  "general":    { "app_name": "KAI App", "default_language": "id" },
  "mobile_app": { "min_version_android": "1.2.0", "force_update_enabled": true, "playstore_url": "..." },
  "maintenance":{ "maintenance_mode_enabled": false, "maintenance_message": "" },
  "contact":    { "whatsapp_number": "+62...", "instagram_url": "..." },
  "payment":    { "bank_name": "BCA", "bank_account_number": "...", "bank_account_holder": "...", "payment_instructions": "..." }
}
```

> Setting sensitif (`smtp_*`, `security.*`, `storage_provider`, `banned_keywords`) **tidak pernah** masuk endpoint publik (`is_public = false`). `banned_keywords` sengaja privat agar tidak bisa di-bypass.

Endpoint ini di-cache agresif (CDN/Redis) karena dipanggil setiap app cold-start.

---

## Maintenance Mode

Perilaku saat `maintenance_mode_enabled = true`:

| Aktor | Perilaku |
|---|---|
| Guest & Member (mobile/API) | Semua endpoint return `503` + body `{ "maintenance": true, "message": "<maintenance_message>" }` — mobile tampilkan halaman maintenance |
| Usergod & Superadmin | **Bypass** — tetap bisa akses semua endpoint (agar bisa mematikan maintenance mode dari backoffice) |
| Endpoint yang tetap hidup | `GET /api/v1/mobile/config`, `GET /health`, endpoint auth backoffice |

Aktivasi/deaktivasi tercatat di Audit Log dengan prioritas tinggi.

---

## Mobile App Version Control

Flow saat app cold-start:

```
App start → GET /api/v1/mobile/config
        ↓
Bandingkan versi app (semver) dengan min_version_<platform>
        ├── app < min_version & force_update_enabled = true
        │       → dialog blocking "Wajib Update" → tombol ke store URL
        ├── app < latest_version (tapi ≥ min_version)
        │       → dialog dismissible "Update Tersedia"
        └── app ≥ latest_version → lanjut normal
```

Aturan:
- Perbandingan versi memakai **semantic versioning** (`major.minor.patch`).
- Validasi backoffice: `min_version` tidak boleh lebih besar dari `latest_version`.
- Force update adalah jalan keluar saat ada breaking change API atau celah keamanan di versi lama.

---

## Caching & Invalidasi

- Semua settings di-load ke **cache (Redis / in-memory)** saat boot — modul lain membaca dari cache, bukan query DB per request.
- Update setting → tulis DB → **invalidate cache** (publish event / hapus key Redis) → reload.
- TTL cache fallback: 5 menit (jaga-jaga jika invalidation gagal di multi-instance).
- `GET /api/v1/mobile/config` punya layer cache sendiri (response-level, TTL pendek mis. 60 detik).

---

## Audit & Keamanan

- **Setiap perubahan** setting & legal document masuk Audit Log: `actor_id`, `setting_key` / `doc_type+version`, `old_value`, `new_value`, `timestamp`, `ip`.
- Setting bertanda `is_sensitive = true` (mis. `smtp_username`): nilai di Audit Log dan response list **dimask** (`ap***ey`).
- Validasi server-side per `value_type` + validasi khusus per key (mis. `smtp_port` 1–65535, `default_language` harus ada di `system_languages`, semver valid untuk versi app).
- Endpoint settings backoffice di-rate-limit normal; tidak ada endpoint settings di mobile selain public config & legal.

---

## Use Cases

### Use Case 1 — Force Update Setelah Breaking Change

Backend rilis API v2 yang tidak kompatibel dengan app < 1.5.0. Superadmin set `min_version_android = 1.5.0`, `min_version_ios = 1.5.0`, `force_update_enabled = true`. User dengan app 1.4.x membuka app → dialog blocking → diarahkan ke store.

### Use Case 2 — Ganti Rekening Pembayaran

KAI ganti rekening bank. Superadmin update group `payment` di backoffice. Cache di-invalidate. User yang membuka halaman upgrade Pro detik berikutnya melihat rekening baru — tanpa deploy.

### Use Case 3 — Publish Terms Baru dengan Wajib Setuju Ulang

Tim legal merevisi S&K (perubahan signifikan soal data pribadi). Superadmin buat versi `2.0.0` (draft), preview, lalu publish dengan `require_reacceptance = true`. Versi `1.x` otomatis archived. Semua user saat next login mendapat dialog persetujuan; sebelum setuju, tidak bisa lanjut. Acceptance tercatat per user.

### Use Case 4 — Maintenance Migrasi DB

Usergod akan migrasi database besar. Aktifkan `maintenance_mode_enabled` + pesan estimasi selesai. Mobile user melihat halaman maintenance; Usergod tetap bisa akses backoffice. Selesai migrasi → matikan toggle.

### Use Case 5 — Ketatkan Anti-Spam

Gelombang spam report palsu terdeteksi. Superadmin turunkan `report_rate_limit_per_day` dari 20 → 10 dan tambah kata di `banned_keywords`. Modul Reporting & Community membaca nilai baru dari cache pada request berikutnya.

---

## Ringkasan Aturan

| Aturan | Detail |
|---|---|
| **Storage model** | Key-value (`system_settings`) + tabel khusus legal (3 tabel) |
| **Groups** | `general`, `mobile_app`, `security`, `email`, `storage`, `payment`, `moderation`, `maintenance`, `contact` |
| **Secrets** | Tidak pernah di DB — env/secret manager |
| **Editor** | Usergod (semua), Superadmin (operasional: general, mobile_app, payment, moderation, maintenance, contact, legal) |
| **Permission baru** | `manage_system_settings`, `manage_legal_documents` |
| **Public config** | `GET /api/v1/mobile/config` — hanya `is_public = true`, tanpa auth |
| **Legal docs** | Versioned, 1 published per type, `require_reacceptance` opsional, acceptance tercatat per user per versi |
| **Maintenance** | 503 untuk member/guest, bypass untuk usergod/superadmin |
| **Version check** | Semver compare di app start via public config |
| **Caching** | Redis/in-memory + invalidate on update, fallback TTL 5 menit |
| **Audit** | Semua perubahan tercatat, nilai sensitif dimask |
| **Duplikasi** | DILARANG — news settings, ads settings, languages, plans tetap di modulnya |

---

## Keputusan yang Masih Terbuka

| # | Topik | Asumsi Sementara |
|---|---|---|
| 1 | Banned keywords: reject saat submit atau auto-flag saja? | Reject saat submit dengan pesan generik (lebih simple); bisa diubah ke flag-only nanti |
| 2 | Re-acceptance blocking total atau grace period? | Blocking saat login berikutnya, tanpa grace period |
| 3 | `payment_confirmation_deadline_hours` dipakai untuk auto-cancel pending upgrade? | Ya — cron modul Subscription membaca nilai ini |
| 4 | Multi-bahasa untuk legal documents? | Fase 1 satu bahasa (id). Struktur versi siap ditambah kolom `language` nanti |
| 5 | Setting per-region (mis. kontak per region)? | Tidak — semua global. Kebutuhan regional ditangani modul Region |
| 6 | Integrasi Midtrans | Key config sudah disiapkan di group `payment` (default inert: `manual`/`sandbox`/`[]`). Saat integrasi dibangun, sisanya masuk modul Subscription: endpoint Snap token, webhook notification handler, mapping status, dan update `PLAN_SUBSCRIPTION_SYSTEM.md` |

---

*Dokumen ini hasil brainstorming kebutuhan System Settings. Skema database di `SYSTEM_SETTINGS_DB_SCHEMA.md`.*
