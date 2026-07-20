# Plan — System Settings: Progress Implementasi per Key

## Context

`SYSTEM_SETTINGS_RULES.md` & `SYSTEM_SETTINGS_DB_SCHEMA.md` mendefinisikan modul System Settings sebagai key-value store yang **dibaca modul lain** saat runtime (Auth, Reporting, Subscription, Upload, dst). Audit kode (2026-07-07) menemukan gap besar antara spec dan implementasi: seluruh key sudah di-seed via migration `0017_create_system_settings_tables.up.sql` dan bisa diedit lewat backoffice (`app/components/settings/GroupForm.vue`), tapi **mayoritas key tidak punya konsumen di backend** — nilai bisa diubah admin tanpa efek apa pun, karena logika terkait masih pakai konstanta hardcoded / env var. Beberapa hardcode bahkan bernilai beda dari default yang di-seed (grup `security`).

Dokumen ini adalah **tracker hidup**: setiap kali satu key selesai benar-benar "dikawinkan" ke logic yang relevan (backend baca dari settings, bukan hardcode), checklist di bawah di-update dari `[ ]` ke `[x]` beserta catatan file yang diubah. Ini sumber kebenaran progress lintas sesi — bukan plan file sementara Claude Code.

Legend kolom **Status**: `[ ]` belum disentuh · `[~]` sebagian (ada bug/gap) · `[x]` selesai dikonsumsi backend sesuai spec.

---

## Group: `general`

| Key | Status | Catatan |
|---|---|---|
| `app_name` | [ ] | Tidak dibaca modul apa pun (email/notif masih hardcode "K-Forum"/"KAI App") |
| `tagline` | [ ] | Diparse di Flutter, tidak ditampilkan di mana pun |
| `support_email` | [ ] | Tidak dibaca modul apa pun |
| `platform_url` | [ ] | Tidak dibaca modul apa pun |
| `default_timezone` | **[~] validasi saja** (2026-07-07, direvisi) | **Percobaan pertama (mengaitkan ke notification DND gate & event reminder/calendar) DIBATALKAN — menyebabkan bug live**: reminder job gagal tersimpan diam-diam saat admin mengubah nilai selain `Asia/Jakarta`. Root cause & sisa implementasi yang aman: lihat section khusus di bawah |
| `default_language` | [ ] | Tidak drive locale backend maupun mobile (mobile pakai picker manual) |
| `public_registration_enabled` | [ ] | `RegisterUseCase` tidak pernah cek setting ini — registrasi selalu terbuka |
| `email_verification_required` | [ ] | Gate verifikasi di `login.go` ter-comment-out, tidak terhubung ke setting ini |

## Group: `mobile_app` — ✅ satu-satunya group end-to-end lengkap

| Key | Status | Catatan |
|---|---|---|
| `min_version_android` / `min_version_ios` | [x] | Passthrough via `/mobile/config`, Flutter compare semver sendiri |
| `latest_version_android` / `latest_version_ios` | [x] | Sama, drive dialog "update tersedia" |
| `force_update_enabled` | [x] | Flutter block navigasi di `splash_screen.dart` |
| `playstore_url` / `appstore_url` | [x] | Dibuka via `url_launcher` dari dialog update |
| `update_message` | [x] | Ditampilkan sebagai body dialog |

## Group: `security`

| Key | Status | Catatan |
|---|---|---|
| `max_login_attempts` | [ ] | Const `MaxLoginAttempts=5` di `user_constant.go:10` — kebetulan sama dengan seed, tapi tidak baca dari settings |
| `lockout_duration_minutes` | [~] BUG | Const `LockoutDuration=15*time.Minute` (`user_constant.go:11`) vs seed `30` menit — **mismatch nyata** |
| `otp_expiry_minutes` | [ ] | Literal `5*time.Minute` di `request_identity_otp.go:79` |
| `otp_max_attempts` | [ ] | Tidak ada implementasi attempt-counter sama sekali di `VerificationCode.Verify()` |
| `otp_resend_cooldown_seconds` | [ ] | Tidak ada rate-limit resend sama sekali |
| `access_token_expiry_minutes` | [~] BUG | Env `JWT_ACCESS_TTL` default 15 menit (`config.go:165`) vs seed `60` menit — **mismatch nyata** |
| `refresh_token_expiry_days` | [ ] | Env `JWT_REFRESH_TTL` default 30 hari — kebetulan cocok, tapi tidak baca settings |
| `seamless_token_expiry_seconds` | [ ] | Const `seamlessTokenTTL=60s` di `token_generator.go:102`, duplikat literal di `generate_seamless_token.go:34` |
| `password_min_length` | [ ] | Literal `len(raw) < 8` di `user_vo.go:93` |
| `password_require_number` | [ ] | Unconditional check, tidak togglable |
| `password_require_uppercase` | [~] BUG | Selalu enforced (`user_vo.go:105-106`) walau default seed `false` — **arah bug terbalik, lebih strict dari yang diminta** |
| `password_require_symbol` | [ ] | Tidak ada code path sama sekali |

## Group: `email`

| Key | Status | Catatan |
|---|---|---|
| `smtp_host` / `smtp_port` / `smtp_username` / `smtp_encryption` / `smtp_from_name` / `smtp_from_email` | [ ] | Semua dari env (`config.go:174-178`), di-set sekali saat boot; form backoffice ada tapi tidak berefek |
| `smtp_password` | N/A | Key ini **tidak ada di spec dokumen** — anomali seed, tapi karena password harus di env, sudah benar secara prinsip meski kontradiksi "kolom ini tidak boleh ada" |
| `welcome_email_enabled` | [ ] | Tidak ada branch registrasi yang cek ini |
| `email_footer_text` | [ ] | Tidak ada outgoing email yang append ini |
| *(bug tambahan)* | — | `SendTestEmailUseCase` tidak pakai nilai form yang baru diisi admin, hanya sender boot-time; payload FE `{to}` tidak cocok DTO BE (`to_email`,`subject`,`body`) → request gagal validasi |

## Group: `storage`

| Key | Status | Catatan |
|---|---|---|
| `storage_provider` / `max_upload_size_mb` / `max_avatar_size_mb` / `max_payment_proof_size_mb` / `allowed_mime_types` / `cdn_base_url` | [ ] | Semua presign usecase (ads/avatar/profile/subscription) hardcode allowlist masing-masing; tidak ada size-limit check di `media_upload_service.go` |

## Group: `payment`

| Key | Status | Catatan |
|---|---|---|
| `bank_name` / `bank_account_number` / `bank_account_holder` / `payment_instructions` | [~] BUG | Masuk public config dengan benar, tapi screen upgrade Flutter **hardcode 2 rekening bank palsu**, tidak baca `AppConfig.payment` — ganti rekening di backoffice tidak berefek ke app |
| `payment_confirmation_deadline_hours` | [ ] | Tidak ada cron auto-cancel (klaim spec dokumen soal ini salah) |
| `payment_provider` | [ ] | Hanya dicek sebagai enum literal di domain constant, bukan dari settings |
| `midtrans_environment` / `midtrans_enabled_channels` | [ ] | Tidak ada kode Midtrans sama sekali |
| `midtrans_integration_mode` | MISSING | **Hilang dari migration 0017** walau didokumentasikan di spec — perlu migration baru kalau mau dipakai |

## Group: `moderation`

| Key | Status | Catatan |
|---|---|---|
| `banned_keywords` | [x] (2026-07-16) | Diimplementasikan penuh — reject 422 `CONTENT_BANNED_KEYWORD` di 6 modul (Community: post/comment/community/announcement/schedule/invitation; QnA: question/answer/edit_answer; News: comment+submit/update article; Event: create/update/admin variant/share; Directory: company/merchant/item/review/inquiry; Ads: create/update member+web). Substring case-insensitive, fail-open kalau settings gagal dibaca, no bypass utk admin. Lihat detail arsitektur di bawah |
| `report_auto_flag_threshold` / `report_rate_limit_per_day` | [ ] | Di luar scope kerja banned_keywords — logic tetap milik modul Reporting sendiri, cuma numpang grup `moderation` di settings |

## Group: `maintenance`

| Key | Status | Catatan |
|---|---|---|
| `maintenance_mode_enabled` / `maintenance_message` | [x] (2026-07-19) | Semua endpoint mobile non-`/config` sekarang ter-gate (termasuk `mobileAuth`, legal, subscription plans, ads publik, home insights, news publik) — 503 untuk guest/member, bypass usergod/superadmin, `/config` & `/health` tetap hidup. Aktivasi/deaktivasi (dan perubahan setting lain) sekarang tercatat di `system_settings_audit_log`. Detail arsitektur & keputusan di section khusus di bawah |

## Group: `contact`

| Key | Status | Catatan |
|---|---|---|
| `whatsapp_number` | [~] | Dipakai di SOS emergency sheet Flutter — bukan halaman About/Help seperti niat spec (halaman itu tidak ada) |
| `instagram_url` / `website_url` | [ ] | Masuk public config, tidak dipakai di mana pun di Flutter |

## Legal Documents

| Item | Status | Catatan |
|---|---|---|
| Versioning/publish/single-published/immutability/`require_reacceptance` | [x] | Semua benar di backend |
| `pending_acceptances` di response login/refresh/me | [ ] | Field ada di DTO, tidak pernah diisi |
| Mobile blocking re-acceptance dialog | [ ] | Data layer ada, tidak pernah dipanggil dari flow app-open |
| Checkbox terms/privacy di registrasi | [ ] | Tidak ada di `register_screen.dart` |

## Cross-cutting

| Item | Status | Catatan |
|---|---|---|
| Public config response-level cache (60s, sesuai spec) | [~] | Redis cache utk `SystemSettingsProvider` (TTL 5 menit) diaktifkan 2026-07-16 (Option B di `main.go`, sebelumnya Option A/langsung DB) — bukan cache di level response endpoint 60s sesuai spec, tapi provider-level. Dipicu oleh `banned_keywords` yang dibaca di ~20 titik submit-konten per request |
| Permission `manage_system_settings`/`manage_legal_documents` | [ ] | Ter-seed tapi route pakai `RequireRole`, permission ini tidak pernah dicek |
| `editable_by` (usergod vs superadmin) | [x] backend / [ ] frontend | Backend benar (`update_settings.go`); frontend `GroupForm.vue` punya `ENFORCE_EDITABLE_BY=false` (TODO) |
| Validasi per-key khusus (semver, timezone, dll) | [~] | `default_timezone` sudah (IANA validation di `update_settings.go`) — key lain (semver versi app, `default_language` vs `system_languages`, dll) belum |
| Audit log perubahan setting | [~] (2026-07-19) | `system_settings_audit_log` tercatat untuk setiap `UpdateSettingsUseCase`/`ToggleMaintenanceUseCase` (semua key, bukan cuma maintenance) — lihat section `maintenance_mode_enabled` di bawah. Transisi status legal document version (publish/archive) BELUM disentuh — masih di luar scope sesi ini |

---

## `default_timezone` — percobaan pertama, ditemukan bug live, direvisi (2026-07-07)

### Percobaan pertama (DIBATALKAN)

Desain awal: value object `PlatformTimezone` + kontrak domain `PlatformTimezoneRepository` + adapter infra baca lewat `port.SystemSettingsProvider`, dikonsumsi oleh notification DND gate (`module_preference_gate.go`) dan event reminder/calendar (`event/helpers.go`, `schedule_event.go`, `get_calendar_export.go`) — mengganti hardcode `jakartaLocation`/`NowInJakarta()`/`combineDateAndTime()` supaya "settingnya benar-benar dipakai".

**Bug yang ditemukan** (dilaporkan user langsung setelah deploy ke dev env, dikonfirmasi lewat log container `k-forum-api_app`/`k-forum-api_worker` + query tabel `scheduled_jobs` + test reproduksi): begitu admin mengubah `default_timezone` ke selain `Asia/Jakarta`, `POST /mobile/events/:id/schedule` tetap sukses (201) TAPI reminder job (`scheduled_jobs` row) **gagal dibuat sama sekali, tanpa error yang terlihat** — best-effort swallow di `scheduleReminderJob` diam-diam menganggap `reminderAt` sudah lewat.

**Root cause:** `event_date`/`event_time` (dan field DND `HH:MM`) adalah wall-clock value yang **selalu** diinput sebagai WIB oleh mobile app — app-nya sendiri tidak punya timezone picker dan tidak tahu apa-apa soal setting `default_timezone`. Ini konvensi bisnis yang tetap (fixed), bukan sesuatu yang seharusnya ikut berubah kalau admin mengganti setting display-only tersebut. Menyambungkan `default_timezone` ke logic ini membuat backend **re-interpretasi** wall-clock value yang sama sebagai timezone lain — menggeser instant absolut yang dihitung (mis. WIB→Tokyo menggeser 2 jam lebih awal), yang untuk event/reminder dekat waktu sekarang bisa menggeser `reminderAt` jadi di masa lalu.

**Fix:** revert semua konsumsi tersebut balik ke hardcode `Asia/Jakarta` (persis seperti sebelumnya) — lihat "Yang dipertahankan" & "Yang di-revert" di bawah. Diverifikasi dengan test reproduksi: `default_timezone` diganti ke `Asia/Tokyo`, jadwalkan event dekat waktu sekarang, reminder job **tetap dibuat** (sebelumnya 0 row, sesudah fix 1 row).

### Yang dipertahankan (aman, tidak menyebabkan bug di atas)

- `internal/domain/system/valueobject/platform_timezone.go` (+ test) — value object `PlatformTimezone` tetap ada, TAPI sekarang **hanya dipakai untuk validasi**, bukan untuk menafsirkan wall-clock input di mana pun.
- `internal/domain/system/constant/system_constant.go` — `CodeTimezoneInvalid` tetap ada.
- `internal/app/usecase/system/update_settings.go` — `validateSettingValue()` tetap ada: `default_timezone` wajib lolos `sysvo.NewPlatformTimezone(value)` (IANA timezone valid) sebelum ditulis, reject 422 kalau tidak. Ini validasi per-key pertama yang ada di usecase ini — murni hygiene, tidak mengubah behavior apa pun di luar endpoint update settings itu sendiri.
- `internal/interfaces/http/handler/web/system_settings_handler_test.go` — `TestWebSystemSettings_UpdateSettings_InvalidTimezone` (422) & `_ValidTimezone` (200 + restore ke default) tetap ada, masih valid untuk validasi di atas.

### Yang di-revert (balik ke hardcode Asia/Jakarta, persis seperti sebelumnya)

- `internal/domain/notification/service/module_preference_gate.go` + `module_preference_gate_test.go` — `jakartaLocation`/`NowInJakarta()` kembali, `DefaultModulePreferenceGate` kembali single-arg constructor, `FilterChannels` kembali menerima param `now time.Time`. Komentar diperluas menjelaskan KENAPA ini sengaja tidak boleh dikaitkan ke `default_timezone`.
- `internal/app/service/notification/dispatcher.go` — call site `FilterChannels` kembali pass `notifservice.NowInJakarta()`.
- `internal/app/usecase/event/helpers.go` — `jakartaLocation` var + `combineDateAndTime()` free function kembali, dengan komentar diperluas menjelaskan root cause bug di atas sebagai regression note.
- `internal/app/usecase/event/schedule_event.go`, `get_calendar_export.go`, `dependencies.go` — `timezoneRepo`/`TimezoneRepo` dihapus, kembali panggil `combineDateAndTime()` langsung.
- `internal/domain/system/repository/interfaces.go` — interface `PlatformTimezoneRepository` dihapus (tidak ada konsumen lagi).
- `internal/infrastructure/cache/platform_timezone_provider.go` (+ test) — **dihapus**, tidak ada konsumen lagi.
- `cmd/app/main.go`, `cmd/worker/main.go`, `internal/testhelper/testserver.go` — semua wiring `platformTimezoneRepo` dihapus; `cmd/app/main.go` kembali ke ordering aslinya (System Settings block sesudah UserManagement, bukan sebelum Event).

### Kalau mau dicoba lagi nanti

`default_timezone` baru bisa aman dikaitkan ke interpretasi wall-clock event/DND KALAU mobile app juga diubah untuk (a) benar-benar timezone-aware saat organizer input jam, atau (b) selalu mengirim instant absolut (UTC) bukan wall-clock string, sehingga backend tidak perlu menerka timezone apa yang dimaksud. Sampai itu terjadi, key ini tetap [~] (tervalidasi tapi tidak dikonsumsi untuk logic apa pun) — status yang sama dengan sebelum sesi ini, ditambah validasi write-time yang aman.

**Verifikasi yang sudah dijalankan:** `go build ./...`, `go vet ./...` bersih. Test lulus: value object, infra adapter, `internal/domain/notification/service` (termasuk regresi WIB+7 offset), handler web system-settings (termasuk 2 test baru), handler mobile event — termasuk `TestMobileEvent_GetCalendarExport_CorrectTimezone` dan `TestMobileEvent_ScheduleEvent_ReminderJobUsesCorrectTimezone` yang sudah ada sebelumnya (regresi bug +7 jam) — semua tetap hijau setelah refactor dari hardcode ke settings-driven.

---

## `banned_keywords` — implementasi penuh (2026-07-16)

### Desain

- **Domain service** `internal/domain/moderation/service/keyword_matcher.go` — `FindBannedKeyword(keywords []string, texts ...string) (matched string, found bool)`, substring case-insensitive, pure (tanpa I/O).
- **App service** `internal/app/service/moderation/keyword_checker.go` — `KeywordChecker.Check(ctx, texts ...string) error`, baca `banned_keywords` via `SystemSettingsProvider.GetAll` (**bukan** `GetPublic` — key ini `is_public=false`). Fail-open kalau settings gagal dibaca (log WARN) — alasan: fail-closed akan menjadikan settings store sebagai titik gagal tunggal utk ~20 endpoint submit-konten sekaligus, blast radius jauh lebih besar daripada risiko konten tidak ter-filter sesaat. Match → `apperror.CodeContentBannedKeyword` ("CONTENT_BANNED_KEYWORD", 422), pesan generik — kata yang match tidak pernah dikembalikan ke client (hanya dicatat di log server untuk audit), supaya user tidak bisa membisect kata mana yang match.
- Dipanggil sebagai statement pertama tiap `Execute()`, sebelum repository read/write apa pun.
- Validasi sisi tulis: `case "banned_keywords"` baru di `validateSettingValue()` (`internal/app/usecase/system/update_settings.go`) — reject non-array, entri kosong, duplikat (case-insensitive), >100 karakter/kata, >500 entri.

### Cakupan (6 modul, sesuai UI backoffice `moderation.vue` yang sudah live — bukan cuma 3 modul minimal di RULES.md)

- **Community**: `create_post.go`, `create_comment.go`, `create_community.go`, `update_community.go`, `create_announcement.go`/`edit_announcement.go`, `create_schedule_entry.go`/`edit_schedule_entry.go`, `send_invitation.go`.
- **QnA**: `submit_question.go`, `post_answer.go`, `answer_question.go` (admin-authored, tetap dicek), `edit_answer.go`.
- **News**: `post_comment.go`, `submit_article.go`, `update_article.go` (admin-authored, tetap dicek).
- **Event**: `create_event.go`, `update_event.go`, `create_admin_event.go`, `admin_update_event.go`, `share_event.go`.
- **Directory/Merchant**: `create_company.go`/`update_company.go`, `create_merchant.go`/`update_merchant.go`, `create_item.go`/`update_item.go`, `leave_review.go`/`update_review.go`, `send_inquiry.go`, `reply_inquiry.go`.
- **Ads/Promotions**: `create_ad.go`, `create_ad_web.go`, `update_ad.go`, `update_ad_web.go`.

Tidak ada bypass role — konten admin/superadmin juga dicek (keputusan eksplisit, konsisten satu code path).

### Efek samping: Redis cache utk SystemSettingsProvider diaktifkan

Karena `banned_keywords` sekarang dibaca di ~20 titik submit-konten per request, `cmd/app/main.go` di-flip dari Option A (langsung Postgres) ke Option B (`CachedSystemSettingsProvider`, TTL 5 menit, sudah terhubung ke `Invalidate()` di `update_settings.go`). Ini mengubah behavior SEMUA konsumen settings (`maintenance_mode_enabled`, `default_language`, dll), bukan cuma moderation — risiko dianggap nihil karena Redis sudah jadi hard dependency (app fatal error saat startup kalau Redis unreachable).

### Verifikasi yang sudah dijalankan

`go build ./...`, `go vet ./...` bersih. Test baru: `internal/domain/moderation/service` (unit test matcher), `internal/app/service/moderation` (unit test checker — fail-open, no-match, match, regresi "harus panggil GetAll bukan GetPublic"), `internal/interfaces/http/handler/web/system_settings_handler_test.go` (validasi banned_keywords invalid/valid), 3 handler test representatif (`TestMobileCommunity_CreatePost_BannedKeyword`, `TestMobileQna_SubmitQuestion_BannedKeyword`, `TestMobileNews_PostComment_BannedKeyword`). Full suite (`mobile`, `web`, `domain/...`, `app/...`) tetap hijau setelah seluruh wiring 6 modul.

---

## `maintenance_mode_enabled` / `maintenance_message` — gating penuh + audit log (2026-07-19)

### Gap yang ditemukan (audit 2026-07-07, sebelum sesi ini)

`middleware.MaintenanceGate` (`internal/interfaces/http/middleware/maintenance.go`) sudah benar secara logic (baca `maintenance_mode_enabled`/`maintenance_message` via `SystemSettingsProvider.GetAll`, fail-open kalau settings gagal dibaca, bypass role `usergod`/`superadmin`, 503 + body `{"maintenance":true,"message":...}` untuk yang lain) — TAPI hanya **dipasang satu kali**, di grup `protectedMobile` (`internal/interfaces/http/router/router.go`). Semua grup mobile publik lain luput sama sekali: `mobileAuth` (register/login/OTP/refresh/password), `/legal/:doc_type`, `/subscription/plans`, `mobileAdsPublic` (`/ads`, `/ads/home`, impression/click), `mobileHome` (weather/exchange-rates), `mobileNewsPublic` (articles/categories/scopes/comments). Guest tetap dapat 200 dari semua endpoint itu walau admin sudah mengaktifkan maintenance mode. Tidak ada audit log sama sekali untuk perubahan setting apa pun (bukan cuma maintenance).

### Keputusan: `mobileAuth` (login/register/OTP) ikut di-gate

`SYSTEM_SETTINGS_RULES.md` §Maintenance Mode eksplisit: *"Guest & Member (mobile/API) | Semua endpoint return 503"* — hanya `GET /mobile/config`, `GET /health`, dan auth backoffice yang dikecualikan. Tidak ada pengecualian untuk mobile login/register. Diputuskan untuk ikut menggerbang `mobileAuth` (bukan cuma 3 kelompok yang disebut di audit awal: legal/ads/subscription-plans) supaya benar-benar sesuai spec "semua endpoint", bukan cuma menutup 3 gap yang kebetulan ditemukan duluan. Backoffice (`/web/auth/*`) TIDAK disentuh — itulah jalur usergod/superadmin tetap bisa masuk untuk mematikan maintenance mode, sesuai catatan spec "agar bisa mematikan maintenance mode dari backoffice".

### Implementasi gating

- `internal/interfaces/http/router/router.go` — `middleware.MaintenanceGate(sysSettingsProvider)` ditambahkan ke: `mobileAuth` group, grup baru `mobilePublicGated` (menaungi `/legal/:doc_type` dan `/subscription/plans`, sebelumnya route langsung di `mobile` group), `mobileAdsPublic`, `mobileHome`, `mobileNewsPublic`. `GET /mobile/config` sengaja TIDAK disentuh (harus tetap hidup — mobile app butuh endpoint ini untuk tahu status maintenance itu sendiri).
- Middleware `MaintenanceGate` sendiri TIDAK diubah — hanya titik pemasangannya yang diperluas.

### Bug test-infra yang ditemukan & diperbaiki: `noopSystemSettingsProvider`

Saat menulis test untuk verifikasi gating, ditemukan `internal/testhelper/testserver.go` menyuntik `noopSystemSettingsProvider{}` (selalu return empty map) ke `middleware.MaintenanceGate` di router test — **terpisah dari** provider Postgres asli (`sysSettingsRepo`) yang dipakai untuk usecase settings lain. Akibatnya `MaintenanceGate` **selalu fail-open di seluruh test suite**, termasuk untuk `protectedMobile` yang sudah lama ter-gate di production — gap ini sudah ada sejak sebelum sesi ini dan membuat gating maintenance tidak pernah benar-benar teruji. Diperbaiki: `sysSettings` di `MustStartTestServer`/`buildRouter` sekarang memakai `sysSettingsRepo` (Postgres langsung, sama seperti yang dipakai `systemUseCases.Provider`) bukan noop. Type `noopSystemSettingsProvider` (beserta importnya) dihapus dari `internal/testhelper/noop.go` karena sudah tidak ada pemakai. Fix ini murni test-infra — tidak mengubah behavior production sama sekali (default `maintenance_mode_enabled=false` di kedua kasus menghasilkan pass-through yang sama).

### Audit log — tabel baru `system_settings_audit_log`

Spec (`SYSTEM_SETTINGS_RULES.md` §Maintenance Mode: *"Aktivasi/deaktivasi tercatat di Audit Log dengan prioritas tinggi"*, §Audit & Keamanan: *"Setiap perubahan setting ... masuk Audit Log: actor_id, setting_key, old_value, new_value, timestamp, ip"*) mensyaratkan audit log untuk maintenance toggle secara eksplisit, dan untuk SEMUA perubahan setting secara umum. Karena `maintenance_mode_enabled`/`maintenance_message` bisa diubah lewat DUA jalur — `POST /web/system-settings/maintenance/toggle` (`ToggleMaintenanceUseCase`) DAN `PATCH /web/system-settings` generik (`UpdateSettingsUseCase`, yang tidak membatasi key apa pun) — audit ditulis di kedua usecase sekaligus, bukan cuma di endpoint toggle, supaya tidak ada jalur bypass yang luput dari audit.

- Migration baru `internal/migrations/20260719104143_create_system_settings_audit_log.up.sql` — tabel `system_settings_audit_log(id, actor_id, setting_key, old_value, new_value, ip_address, created_at)`, mengikuti pola persis `user_audit_logs` (migration `0015_usermanagement_additions.up.sql`) tapi tanpa kolom `actor_role`/`target_type`/`target_id`/`notes` (tidak relevan untuk setting key-value).
- Domain: `internal/domain/system/entity/settings_audit_log.go` (`SettingsAuditLog` + `NewSettingsAuditLog`, immutable), `internal/domain/system/repository/interfaces.go` (`SettingsAuditLogRepository.Save`), error code baru di `internal/domain/system/constant/system_constant.go`.
- Infra: `internal/infrastructure/persistence/postgres_system_settings_audit_log_repository.go` (`PostgresSettingsAuditLogRepository`).
- Usecase: `writeSettingsAuditLogs()` helper (di `internal/app/usecase/system/toggle_maintenance.go`, dipakai juga oleh `update_settings.go`) menulis satu row per key yang berubah, best-effort (gagal tidak menggagalkan operasi utama). Nilai untuk setting `is_sensitive=true` dimask jadi literal `"***"` di kolom `old_value`/`new_value` (bukan algoritma mask prefix/suffix yang dipakai `SystemSetting.MaskedValue()` — cukup untuk audit trail, tidak perlu presisi yang sama dengan response list). `clientIP` (`c.ClientIP()`) di-thread dari handler (`system_settings_handler.go`) ke kedua usecase sebagai parameter baru.
- Wiring: `Dependencies.AuditLogRepo` baru di `internal/app/usecase/system/use_cases.go`; `cmd/app/main.go` dan `internal/testhelper/testserver.go` mengkonstruksi `PostgresSettingsAuditLogRepository` dan mengoper ke `systemUsecase.Dependencies`.

### Yang BELUM disentuh (di luar scope sesi ini)

- Audit log untuk transisi status legal document version (publish/archive) — masih pakai jalur lama tanpa audit trail terpisah (row cross-cutting tetap `[~]`, bukan `[x]`).
- `report_auto_flag_threshold`/`report_rate_limit_per_day` (modul Reporting) dan seluruh key group lain yang masih `[ ]`/`[~]` di tabel utama — tidak disentuh.

### Verifikasi yang sudah dijalankan

`go build ./...`, `go vet ./...` bersih. Test baru: `internal/interfaces/http/handler/mobile/maintenance_gate_test.go` (`TestMobileMaintenanceGate_BlocksPublicRoutes` — 8 subtest per route yang baru di-gate termasuk `mobileAuth`, `TestMobileMaintenanceGate_ConfigStaysAlive`, `TestMobileMaintenanceGate_DisabledAllowsPublicRoutes`), 2 test baru di `internal/interfaces/http/handler/web/system_settings_handler_test.go` (`TestWebSystemSettings_ToggleMaintenance_WritesAuditLog`, `TestWebSystemSettings_UpdateSettings_WritesAuditLog` — assert row masuk ke `system_settings_audit_log`). Full suite `internal/interfaces/http/handler/mobile/...` dan `internal/interfaces/http/handler/web/...` (bukan cuma yang baru) tetap hijau setelah fix noop provider — mengonfirmasi tidak ada regresi dari mengaktifkan gate yang sebelumnya selalu no-op di test.
