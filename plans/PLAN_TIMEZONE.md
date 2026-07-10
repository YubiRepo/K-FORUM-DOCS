# Implementasi User Settings — Bertahap per Modul

## Context

Modul **User Settings** (spec di `K-FORUM-DOCS/Modules/Users Settings/`) adalah modul self-scope kecil (1 tabel: `user_settings` — language + theme) yang sebagian besar isinya **agregasi** dari modul lain: device list dari FCM (`device_registrations`), linked account dari Auth (`credentials`/`user_identities`), dan bahasa aktif dari System Settings (`system_languages`). Karena scope-nya menyentuh banyak bounded context, implementasi dipecah per modul yang terdampak, bukan satu PR besar.

Riset ke codebase menemukan **3 gap nyata antara asumsi spec dan kondisi aktual**, yang mengubah urutan & isi fase:

1. **Tabel FCM sebenarnya bernama `device_registrations`**, bukan `fcm_tokens` — endpoint `/fcm/*` sudah ada tapi kolom `timezone` belum ada.
2. **Auth 100% stateless JWT** — tidak ada endpoint `/auth/logout`, tidak ada tabel session/refresh. TAPI: JWT access token sudah membawa `sid` (session_id) + `iat` (issued_at) di claims, dan Redis sudah dipakai untuk `principal:<userID>` cache (`internal/infrastructure/cache/redis_principal_cache.go`). Infra ini bisa **diperluas** (bukan dibangun dari nol) jadi revocation store beneran — ini titik krusial yang mengubah rencana dari "defer ke masa depan" jadi "bangun sekarang, scope kecil".
3. **DND (Notification) & Event reminder SUDAH SENGAJA di-hardcode ke `Asia/Jakarta`**, dengan komentar kode yang secara eksplisit mendokumentasikan insiden masa lalu: mengaitkan jam DND/reminder ke timezone yang bisa berubah (`default_timezone` system setting, sebuah nilai GLOBAL yang berubah untuk semua orang sekaligus) pernah dicoba dan menyebabkan reminder bergeser ~7 jam — lihat `internal/domain/notification/service/module_preference_gate.go:12-21` dan `internal/app/usecase/event/helpers.go:239-256`. **Tujuan refactor ini justru untuk mengganti hardcode itu dengan sumber kebenaran yang benar** (`device_registrations.timezone`, per-user, bukan setting global) — tapi Event-reminder (instant absolut, aman langsung dikonversi ke UTC) dan DND (rule wall-clock berulang, harus dievaluasi dinamis per-user, tidak boleh "dibekukan" ke UTC) butuh strategi berbeda dan level risiko berbeda; lihat Fase 7a/7b/7c untuk pembagiannya dan alasan urutannya.

Outcome yang diinginkan: 8 endpoint dari `API_SPEC_USER_SETTINGS.md` berjalan sesuai kontrak dengan enforcement logout yang **benar-benar real**, ditambah Event/DND yang akhirnya bersumber dari timezone device aktual user (bukan hardcode WIB) tanpa mereproduksi insiden lama.

---

## Fase 1 — User Settings Core (CRUD): App Preferences ✅ SELESAI (2026-07-10)

**Endpoint:** `GET /user-settings/preferences`, `PUT /user-settings/preferences` (spec #2, #3)

**Status:** Implementasi selesai, `go build ./...` + `go vet ./...` bersih, semua test baru pass (6/6) + full suite `mobile`/`web` handler tetap hijau. File yang dibuat/diubah:

- Migration: `internal/migrations/20260710014048_create_user_settings.up.sql` / `.down.sql`
- Domain: `internal/domain/usersettings/{entity,repository,constant}/*.go`
- Infra: `internal/infrastructure/persistence/postgres_user_settings_repository.go`
- Usecase: `internal/app/usecase/usersettings/{dependencies,get_preferences,update_preferences,helpers}.go`
- DTO: `internal/app/dto/user_settings_dto.go`
- Handler: `internal/interfaces/http/handler/mobile/user_settings_handler.go` + test-nya
- Router: `internal/interfaces/http/router/router.go` (registrasi di grup universal `protected`, path `/api/v1/user-settings/preferences`)
- Wiring: `cmd/app/main.go` (repo + usecases + handler, setelah `sysSettingsProvider` tersedia)
- Test harness: `internal/testhelper/testserver.go` — **catatan penting**: harness test punya `buildRouter()` sendiri yang independen dari `router.go` produksi (tidak memanggil `router.Setup()` sama sekali, dan sudah lebih dulu ketinggalan beberapa route produksi seperti `/auth/session`). Ditambahkan grup universal baru `universalProtected := v1.Group("", authMW...)` di `buildRouter()` khusus untuk endpoint ini, supaya path test (`/api/v1/user-settings/preferences`) sama dengan production, bukan ikut pola `/api/v1/mobile/fcm/...` yang dipakai FCM di test harness (yang sendiri berbeda dari path FCM produksi `/api/v1/fcm/...` — inkonsistensi pre-existing, di luar scope, tidak disentuh).

**Deviasi dari asumsi spec awal (ditemukan saat implementasi):**
- `users.id` di DB ternyata tipe **UUID** dan timestamp **TIMESTAMPTZ** (bukan `VARCHAR(36)`/`TIMESTAMP` seperti asumsi `USER_SETTINGS_DB_SCHEMA.md`) — migration `user_settings` disesuaikan ikut konvensi asli (`user_id UUID PRIMARY KEY REFERENCES users(id)`).
- Validasi `language` terhadap `system_languages`: dicek `IsActive` saja (match wording spec "harus ada & aktif"). Daftar `available_languages` di response GET difilter lebih ketat (`IsActive && IsUILanguage`) supaya konsisten dengan makna kolom `is_ui_language`.
- Partial update null-vs-absent (`language: null` = reset, field absen = jangan diubah) diimplementasikan dengan parsing body sebagai `map[string]json.RawMessage` di handler (bukan `ShouldBindJSON` langsung ke struct), karena `*string` polos tidak bisa membedakan dua kasus itu.

---

**Modul baru**, bounded context sendiri, pola auto-create meniru `notificationpreference` (`internal/app/usecase/notificationpreference/get_preference.go`, `internal/infrastructure/persistence/postgres_notification_module_preference_repository.go:35-49` — `FindOrCreateByUserID`).

- Migration via `make migrate-create NAME=create_user_settings`: tabel `user_settings (user_id PK FK, language NULL, theme NOT NULL DEFAULT 'system', created_at, updated_at)` sesuai `USER_SETTINGS_DB_SCHEMA.md`.
- `internal/domain/usersettings/entity/user_settings.go` — entity + factory default.
- `internal/domain/usersettings/repository/interfaces.go` — `UserSettingsRepository{FindOrCreateByUserID, Update}`.
- `internal/domain/usersettings/constant/user_settings_constant.go` — `CodeUserSettingsLanguageInvalid`, `CodeUserSettingsPersistenceFailed`, `CodeUserSettingsQueryFailed`.
- `internal/infrastructure/persistence/postgres_user_settings_repository.go` — implementasi + error mapping (§7 CLAUDE.md).
- `internal/app/usecase/usersettings/` — package baru: `dependencies.go`, `get_preferences.go`, `update_preferences.go`, `helpers.go` (`mapUserSettingsDomainError`, resolve `language_effective`).
  - Resolve chain (`user_settings.language` → header `X-Locale`/`Accept-Language` → `default_language`) inject dua dependency lintas-context yang **sudah ada**, bukan bikin baru:
    - `systemlanguage/repository.SystemLanguageRepository` (`FindByCode`, `FindAll`) — validasi & daftar `available_languages`.
    - `port.SystemSettingsProvider` (`GetPublic`) — fallback `default_language` (sudah di-cache Redis).
- `internal/app/dto/user_settings_dto.go` — `PreferencesResponse`, `UpdatePreferencesRequest`.
- Handler: **package `mobile`** (`internal/interfaces/http/handler/mobile/user_settings_handler.go`) — mengikuti presedan `device_handler.go` (endpoint universal tapi secara fisik hidup di package `mobile`, dipasang di grup router universal, bukan `protectedMobile`/`protectedWeb`). Test wajib di file yang sama (§11 CLAUDE.md).
- Router: daftar di grup universal `protected := v1.Group("")` (`internal/interfaces/http/router/router.go:152`, tempat yang sama dengan `/fcm`) → `protected.Group("/user-settings")`.
- Wiring: tambah 3 baris di `cmd/app/main.go` (repo, usecases, handler) — pola identik notification preferences (baris ~434, ~470-474).

---

## Fase 2 — FCM (`device_registrations`): kolom `timezone` ✅ SELESAI (2026-07-10)

**Tidak ada endpoint User Settings baru di fase ini** — ini prasyarat data untuk Fase 4.

**Status:** Implementasi selesai, `go build ./...` + `go vet ./...` bersih, test baru pass (termasuk validasi timezone invalid → 422) + full suite `mobile`/`web` handler tetap hijau. File yang dibuat/diubah:

- Migration: `internal/migrations/20260710015923_add_timezone_to_device_registrations.up.sql` / `.down.sql`
- Domain: `internal/domain/notification/entity/device_registration.go` (field `Timezone *string` + method `SetTimezone` + `validateTimezone` via `time.LoadLocation`, dipanggil dari `NewDeviceRegistration` juga), `internal/domain/notification/constant/error_code.go` (`CodeDeviceInvalidTimezone`)
- Infra: `internal/infrastructure/persistence/postgres_device_repository.go` (kolom `timezone` di `Save`/`Update` + kedua scan helper)
- DTO: `internal/app/dto/device_dto.go` (`RegisterDeviceInput`, `RefreshDeviceTokenInput`, `DeviceItem`, `RegisterDeviceResponse`) + `internal/app/dto/notification_tester_dto.go` (`AdminDeviceItem`, konsistensi admin view)
- Usecase: `internal/app/usecase/device/{register_device,refresh_device_token,helpers}.go` (thread `Timezone` di kedua alur re-registrasi & registrasi baru, termasuk `registerAsNew` di refresh), `internal/app/usecase/notificationtester/get_user_devices.go` (`toAdminDeviceItem`)
- Test: `internal/interfaces/http/handler/mobile/device_handler_test.go` — tambah assert timezone ikut kebawa di response register, + skenario baru `TestMobileDevice_RegisterDevice_InvalidTimezone` (422)

**Catatan implementasi:**
- `RefreshDeviceTokenUseCase.registerAsNew`'s error path untuk `NewDeviceRegistration` diubah dari `apperror.Internal(...)` generik jadi `mapDeviceDomainError(...)` — perbaikan kecil yang jadi perlu karena sekarang path itu juga bisa gagal dengan `CodeDeviceInvalidTimezone` (harus 422, bukan 500).
- Validasi timezone sengaja **tidak** reuse `valueobject.PlatformTimezone` (domain `system`) untuk menghindari import antar-domain di layer entity — cukup duplikasi `time.LoadLocation` check yang sama persis, domain `notification` tetap self-contained.
- Tidak ada fallback untuk kasus tzdata tidak tersedia di environment (beda dari `DefaultPlatformTimezone()` yang punya `FixedZone` fallback) — karena ini menerima value APAPUN dari client, tidak ada satu zona spesifik untuk di-hardcode sebagai fallback. Ini konsisten dengan keterbatasan yang sudah diterima di `NewPlatformTimezone` untuk kasus yang sama.

- Migration `make migrate-create NAME=add_timezone_to_device_registrations`: `ALTER TABLE device_registrations ADD COLUMN timezone VARCHAR(64) NULL`.
- `internal/domain/notification/entity/device_registration.go`: tambah field `Timezone *string` di struct, `DeviceRegistrationParams`, dan branch update di `RegisterDeviceUseCase`/`RefreshDeviceTokenUseCase`.
- `internal/infrastructure/persistence/postgres_device_repository.go`: tambah `timezone` ke kolom `Save`/`Update` (positional!) dan ke **kedua** scan helper (`scanDeviceRegistration`, `scanDeviceRegistrations`).
- `internal/app/dto/device_dto.go`: tambah `Timezone *string` ke `RegisterDeviceInput`, `RefreshDeviceTokenInput`, `DeviceItem`, `RegisterDeviceResponse` (opsional juga ke `AdminDeviceItem` di `notification_tester_dto.go` biar konsisten).
- `internal/app/usecase/device/helpers.go`: update `toDeviceItem`/`toRegisterDeviceResponse` untuk memetakan `Timezone`.
- Validasi ringan: kalau diisi, harus IANA valid (`time.LoadLocation`) → domain error baru `CodeDeviceInvalidTimezone` di `internal/domain/notification/constant/error_code.go`.
- Update `internal/interfaces/http/handler/mobile/device_handler_test.go` (tambah field di payload test, tidak wajib bikin skenario baru kecuali mau test validasi timezone invalid).

---

## Fase 3 — Auth: Session Revocation Nyata (perluasan infra Redis yang sudah ada) ✅ SELESAI (2026-07-10)

**Tidak ada endpoint User Settings di fase ini**, tapi menghasilkan `POST /auth/logout` (yang sebenarnya belum ada padahal diasumsikan spec sudah ada) + port revocation yang dipakai Fase 4.

**Status:** Implementasi selesai, `go build ./...` + `go vet ./...` bersih, test baru pass (401 unauth + 200 logout dengan assert token lama benar-benar ditolak di request berikutnya) + full suite `mobile`/`web` handler tetap hijau. File yang dibuat/diubah:

- Port: `internal/app/port/token_generator.go` (`TokenClaims.IssuedAt`, `RefreshTokenClaims{UserID, IssuedAt}` — dipakai sebagai return type baru `ParseRefreshToken`, bukan tuple `(string, time.Time, error)` seperti draft awal, lebih idiomatis Go dan konsisten dgn pola `TokenClaims`), `internal/app/port/session_revocation.go` (`SessionRevocationStore` — sesuai draft)
- Infra: `internal/infrastructure/external/jwt/token_generator.go` (isi `IssuedAt` di kedua Parse*), `internal/infrastructure/cache/redis_session_revocation.go` (persis pola draft: `session:revoked:<sessionID>`, `user:revoked_before:<userID>`)
- Middleware: `internal/interfaces/http/middleware/auth.go` (`JWTAuth` terima `port.SessionRevocationStore`, cek `IsSessionRevoked` + `RevokedBefore` vs `IssuedAt`, fail-open kalau store error)
- Usecase: `internal/app/usecase/auth/refresh_token.go` (cek `RevokedBefore` sebelum mint token baru — fail-open juga), `internal/app/usecase/auth/logout.go` (baru, `LogoutUseCase`)
- Handler: `internal/interfaces/http/handler/web/auth_handler.go` (`Logout` method baru, mengikuti presedan `GetSession` — universal endpoint yang secara fisik hidup di `web.AuthHandler` meski dipasang di grup universal)
- Router: `internal/interfaces/http/router/router.go` (`POST /auth/logout` di grup `sessionProtected`, param `sessionRevocationStore` baru di `Setup()` + semua 4 titik `middleware.JWTAuth(...)`)
- Wiring: `cmd/app/main.go` (`sessionRevocationStore` dari `cache.NewRedisSessionRevocationStore(redisClient)`, `logoutUC`, threading ke `refreshTokenUC`/`webAuthHandler`/`router.Setup`). **`cmd/worker/main.go` TIDAK disentuh** — dicek dulu, worker tidak pernah memanggil `middleware.JWTAuth` sama sekali (bukan HTTP server dengan auth), jadi tidak ada yang perlu diwiring di sana; draft plan menyebutnya sebagai kemungkinan titik wiring tapi ternyata tidak relevan.
- Test harness: `internal/testhelper/testserver.go` (param baru di `buildRouter`, route `/auth/logout` di `universalProtected` yang sudah ada dari Fase 1) + `internal/testhelper/noop.go` (`fakeSessionRevocationStore` — **bukan** no-op sungguhan seperti pola lain di file itu, melainkan implementasi in-memory asli dengan mutex+map, karena test butuh assert token lama benar-benar ditolak setelah logout, bukan cuma "tidak error")
- Test: `internal/interfaces/http/handler/web/auth_logout_handler_test.go` (baru)

**Catatan implementasi & deviasi dari draft:**
- `GET /api/v1/auth/session` (dipakai draft sebagai contoh endpoint "sebelum/sesudah logout" di test) **ternyata tidak terdaftar sama sekali** di `buildRouter()` test harness (gap pre-existing yang sama seperti dicatat di Fase 1) — test logout dialihkan pakai `GET /api/v1/user-settings/preferences` (sudah ada dari Fase 1, di grup universal yang sama) sebagai endpoint pembuktian revocation.
- Sesuai desain awal: **logout biasa cuma merevoke session/device sekarang** (via `RevokeSession`), belum ada endpoint publik untuk `RevokeAllBefore` (logout-all) — itu baru dipakai lewat port langsung oleh usecase Fase 4, tidak diexpose sebagai endpoint Auth generik.
- Enforcement bersifat **fail-open**: kalau Redis gagal dibaca saat cek revocation, request tetap lanjut (tidak diblokir). Ini best-effort by design (konsisten dengan pola cache lain di codebase ini), bukan bug — trade-off yang diterima supaya gangguan Redis tidak mengunci semua user keluar sekaligus.

Kunci: `accessClaims` sudah punya `SessionID` + `IssuedAt` (`internal/infrastructure/external/jwt/token_generator.go:28-31`), tapi `port.TokenClaims` (`internal/app/port/token_generator.go:3-6`) belum mengekspos `IssuedAt`, dan `ParseRefreshToken` cuma balikin `userID` polos. `middleware.JWTAuth` sudah set `session_id` di context tapi tidak pernah dipakai untuk apa pun. Redis sudah wired (`redis_principal_cache.go` jadi cetakan pola).

- `internal/app/port/token_generator.go`: tambah `IssuedAt time.Time` ke `TokenClaims`; ubah signature `ParseRefreshToken(token string) (string, error)` → `(userID string, issuedAt time.Time, err error)`.
- `internal/infrastructure/external/jwt/token_generator.go`: isi `IssuedAt` di `ParseAccessToken`/`ParseRefreshToken` dari claims yang sudah ada (tidak perlu ubah struct claims, cuma expose field yang sudah di-generate).
- Update semua caller `ParseRefreshToken` (`internal/app/usecase/auth/refresh_token.go`) + test double di `internal/testhelper/noop.go`.
- **Port baru** `internal/app/port/session_revocation.go`:
  ```go
  type SessionRevocationStore interface {
      RevokeSession(ctx context.Context, sessionID string, ttl time.Duration) error
      IsSessionRevoked(ctx context.Context, sessionID string) (bool, error)
      RevokeAllBefore(ctx context.Context, userID string, before time.Time, ttl time.Duration) error
      RevokedBefore(ctx context.Context, userID string) (*time.Time, error)
  }
  ```
- **Impl Redis** `internal/infrastructure/cache/redis_session_revocation.go` (pola sama seperti `RedisPrincipalCache`): key `session:revoked:<sessionID>` (TTL = access token TTL) untuk logout per-session; key `user:revoked_before:<userID>` (value = unix ts, TTL = refresh token TTL) untuk logout-all.
- `middleware.JWTAuth`: tambah parameter `store port.SessionRevocationStore`; setelah parse claims sukses, cek `IsSessionRevoked(sessionID)` dan `RevokedBefore(userID)` vs `claims.IssuedAt` → 401 kalau kena. Update semua call site (`router.go`, `cmd/app/main.go`, `cmd/worker/main.go`, `internal/testhelper/testserver.go`).
- `RefreshTokenUseCase`: inject store juga, tolak refresh kalau `issuedAt` refresh token < `RevokedBefore(userID)`.
- Usecase baru `internal/app/usecase/auth/logout.go` (`LogoutUseCase.Execute(ctx, userID, sessionID)` → `store.RevokeSession`).
- Endpoint baru **beneran**: `POST /api/v1/auth/logout`, didaftar di grup `sessionProtected` (`router.go:143-150`, sudah punya JWTAuth+PrincipalLoader, universal — sama tempat `/auth/session`).
- Tidak perlu endpoint publik untuk "revoke all" — itu dipakai lewat port langsung oleh Fase 4 (lihat di bawah, bukan lewat panggilan usecase-ke-usecase lintas modul, biar konsisten dengan pola cross-context yang sudah ada: modul lain saling berbagi **port**, bukan usecase satu sama lain).
- Test: handler test baru untuk `/auth/logout` (401 unauth, 200 success — assert token lama ditolak di request berikutnya).

---

## Fase 4 — User Settings: Session / Device Management ✅ SELESAI (2026-07-10)

**Endpoint:** `GET /user-settings/devices`, `DELETE /user-settings/devices/{id}`, `POST /user-settings/devices/logout-all` (spec #4, #5, #6)

Bergantung pada Fase 2 (kolom timezone) & Fase 3 (`port.SessionRevocationStore`).

**Status:** Implementasi selesai persis sesuai draft (tidak ada gap/deviasi baru yang ditemukan — semua asumsi Fase 2 & 3 terbukti cukup). `go build ./...` + `go vet ./...` bersih, 8 test baru pass (termasuk skenario penting: logout device LAIN tidak ikut merevoke session device sekarang) + full suite `mobile`/`web` handler tetap hijau. File yang dibuat/diubah:

- Domain: `internal/domain/notification/repository/interfaces.go` (`DeactivateAllByUserID` baru di `DeviceRegistrationRepository`)
- Infra: `internal/infrastructure/persistence/postgres_device_repository.go` (impl `DeactivateAllByUserID`, single `UPDATE ... WHERE user_id=$1`)
- DTO: `internal/app/dto/user_settings_dto.go` (`DeviceViewResponse`)
- Usecase: `internal/app/usecase/usersettings/{list_devices,logout_device,logout_all_devices}.go` (baru) + `dependencies.go` (tambah `DeviceRepo`, `RevocationStore`, `AccessTokenTTL`, `RefreshTokenTTL`) + `helpers.go` (`mapUserSettingsDeviceError` — mapper terpisah dari `mapUserSettingsDomainError` karena taksonomi error device milik domain `notification`, bukan `usersettings`)
- Handler: `internal/interfaces/http/handler/mobile/user_settings_handler.go` (`ListDevices`, `LogoutDevice`, `LogoutAllDevices`)
- Router: `internal/interfaces/http/router/router.go` (`GET/POST/DELETE` di bawah `protected.Group("/user-settings/devices")`)
- Wiring: `cmd/app/main.go`, `internal/testhelper/testserver.go` (thread `deviceRepo` + `sessionRevocationStore` + TTL ke `usersettingsUsecase.Dependencies`, route baru di `buildRouter`)
- Test: `internal/interfaces/http/handler/mobile/user_settings_devices_handler_test.go` (baru — 8 skenario: list 401/200, logout-device 401/404/200-current-device-revoke/200-other-device-no-revoke, logout-all 401/200-revoke)

**Catatan implementasi:**
- Endpoint `POST /user-settings/devices/logout-all` sekarang benar-benar **full enforcement instan** (bukan cuma kontrak API yang dipenuhi tapi efeknya parsial seperti yang ditoleransi spec) — berkat `SessionRevocationStore` dari Fase 3. Ini konkret menutup gap "dependency Auth belum siap" yang disebut spec asli.
- ~~Logout per-device untuk device selain device sekarang sengaja tetap partial~~ — **superseded, lihat Fase 4b di bawah.** Awalnya didesain begitu karena `sessionID` tidak stabil per-device (di-mint ulang tiap refresh token). Ternyata bisa diperbaiki tanpa perubahan arsitektur besar — lihat Fase 4b.

---

## Fase 4b — Fix: Logout device lain sekarang benar2 revoke (bukan cuma stop push) ✅ SELESAI (2026-07-10)

**Bug report dari user:** "logout satu device dari backoffice, tapi di device itu masih tetep bisa hit endpoint, gak unauthorized." Investigasi ulang menemukan batasan Fase 4 lebih longgar dari yang didokumentasikan — dokumen awal bilang "akses akhirnya putus saat refresh ditolak", tapi kenyataannya `RevokeAllBefore` (yang bikin refresh ditolak) **hanya dipanggil oleh Logout All**, bukan logout single-device. Jadi device lain yang di-logout satuan: push notif berhenti, TAPI access token tetap valid sampai expired alami, DAN refresh token-nya tidak pernah diblokir — bisa refresh terus tanpa batas.

**Root cause:** `sessionID` di-mint ulang acak setiap login/refresh dan tidak pernah disimpan terkait ke device manapun, jadi backend tidak punya cara tahu sessionID mana yang harus direvoke untuk device selain device yang sedang request.

**Fix:** tanam `device_id` (device_id yang sama dipakai saat register FCM) ke JWT access & refresh token sejak **login**, lalu pakai itu sebagai kunci revocation baru yang device-scoped — mirip pola `RevokeAllBefore`/`RevokedBefore` dari Fase 3, tapi di-scope ke `userID+deviceID` bukan cuma `userID`.

- Port: `internal/app/port/token_generator.go` (`TokenClaims.DeviceID`, `RefreshTokenClaims.DeviceID`; `GenerateAccessToken`/`GenerateRefreshToken` terima parameter `deviceID` baru), `internal/app/port/session_revocation.go` (`RevokeDeviceBefore`, `DeviceRevokedBefore` — key Redis `device:revoked_before:<userID>:<deviceID>`)
- Infra: `internal/infrastructure/external/jwt/token_generator.go` (claim `did` baru di access & refresh token), `internal/infrastructure/cache/redis_session_revocation.go` (impl key device-scoped)
- Middleware: `internal/interfaces/http/middleware/auth.go` (`JWTAuth` cek `DeviceRevokedBefore` kalau `claims.DeviceID != ""`; juga expose `c.Set("device_id", ...)`)
- Usecase: `internal/app/usecase/auth/refresh_token.go` (cek device-scoped revocation juga, dan **carry-forward** `claims.DeviceID` ke token baru — client tidak perlu resend device_id tiap refresh, cukup sekali saat login), `login.go`/`google_login.go` (kirim `req.DeviceID` ke `GenerateAccessToken`/`GenerateRefreshToken`), `verify_seamless_token.go` (deviceID="" — seamless handoff tidak punya device context, fallback aman ke perilaku lama)
- DTO: `internal/app/dto/auth_dto.go` (`LoginRequest.DeviceID`, `GoogleLoginRequest.DeviceID` — keduanya opsional, `omitempty`; tidak breaking untuk client lama yang belum kirim)
- Usecase: `internal/app/usecase/usersettings/logout_device.go` (cabang "device lain" sekarang panggil `RevokeDeviceBefore(userID, dr.DeviceID, now, refreshTokenTTL)` — merevoke access DAN refresh device itu sekaligus, bukan cuma nonaktifkan FCM)
- Test: `internal/interfaces/http/handler/mobile/user_settings_devices_handler_test.go` (`TestMobileUserSettingsDevices_LogoutDevice_OtherDevice_RevokesThatDevicesAccessAndRefresh` — 2 device login terpisah dengan `device_id` masing-masing via helper baru `mustLoginWithDevice`, buktikan device yang di-logout dari device lain langsung 401 di access DAN refresh, device yang melakukan logout tidak terpengaruh)
- Frontend (`k-forum-backoffice`): `app/stores/auth.ts` — `login()` & `loginWithGoogle()` sekarang kirim `device_id: useFcmStore().getDeviceId()` supaya sesi backoffice juga ikut terikat device_id (tanpa ini, fix di atas tidak berlaku untuk sesi yang login dari backoffice lama).

**Batasan yang MASIH ada (jujur, bukan disembunyikan):** fix ini hanya berlaku kalau device yang di-logout **login dengan mengirim `device_id`** — client lama (mobile app version lama, atau sesi yang sudah login sebelum fix ini deploy) tidak punya `did` di token-nya, sehingga device-scoped revocation tidak berlaku untuknya; efeknya balik ke perilaku lama (cuma stop push, akses jalan sampai token lama itu expired). **Mobile app Flutter (`k_forum`) belum disentuh di sesi ini** — perlu ditambahkan `device_id` di payload login/google-login-nya juga (pola sama seperti fix di `k-forum-backoffice`) supaya device mobile ikut terlindungi penuh, bukan cuma sesi backoffice.

**Diverifikasi end-to-end via Playwright** (2 browser context terisolasi = 2 "device" beda, device_id beda-beda dari localStorage masing-masing): login device A & B, device A register FCM, dari UI device B klik "Logout" di row device A → toast "Device logged out" → device A langsung 401 di endpoint manapun, device B tetap valid.

**Catatan operasional:** backend dev yang jalan di `localhost:8888` (proses `air`, root-owned) sempat ketahuan **stale** — binary belum rebuild sejak commit lama, tidak reflect perubahan Fase 4b (dan mungkin perubahan lain di sesi ini). Verifikasi E2E akhirnya dilakukan dengan menjalankan instance backend terpisah (`go run ./cmd/app` di port 8889, pakai Postgres/Redis/RabbitMQ dev yang sama via port yang di-expose). **Perlu restart proses `air` di 8888 supaya dev server utama benar-benar pakai kode terbaru.**

- **Identifikasi "current device"**: `GET /user-settings/devices?device_id=xxx` — query param opsional, client kirim `device_id` yang sama dipakai saat register FCM (keputusan yang sudah dikonfirmasi; tidak perlu header/middleware baru). Kalau tidak dikirim, semua `is_current=false`.
- Usecase baru di package `usersettings` (BUKAN reuse usecase `device` package — inject `notificationrepo.DeviceRegistrationRepository` langsung, pola cross-context yang sudah dipakai `news` untuk `SystemLanguageRepository`):
  - `list_devices.go` — `FindActiveByUserID`, map ke DTO, `is_current` = match `device_id` param.
  - `logout_device.go` — `FindByID` + cek ownership (`UserID`) → 404 kalau bukan milik user; `dr.Deactivate()` + `Update()`. **Kalau device yang di-logout == device sekarang** (device_id param match device row itu) → panggil `SessionRevocationStore.RevokeSession(ctx, sessionID, accessTokenTTL)` juga (sessionID diambil dari `c.GetString("session_id")` di handler, sudah di-set `JWTAuth`) → device sekarang benar-benar ter-invalidate seketika. Device lain: tetap partial (deactivate FCM saja) — ini **memang** batasan yang diterima spec sendiri karena sessionID tidak stabil per-device (di-mint ulang tiap refresh), jadi tidak perlu dipaksa full untuk kasus ini.
  - `logout_all_devices.go` — repo method baru `DeactivateAllByUserID(ctx, userID) error` di `DeviceRegistrationRepository` (+ impl Postgres, single `UPDATE ... WHERE user_id=$1`), **plus** `SessionRevocationStore.RevokeAllBefore(ctx, userID, now, refreshTokenTTL)` → ini **full enforcement instan**, termasuk device sekarang — menutup gap yang di spec ditandai sebagai "dependency Auth belum siap".
- DTO: `DeviceViewResponse` (id, platform, device_name, os_version, timezone, last_used_at, is_current).
- Handler + route di `protected.Group("/user-settings/devices")`.
- Test: 401, 404 (device bukan milik user / device_id salah), 200 list, 200 logout per-device (assert token lama current-device ditolak setelah logout), 200 logout-all (assert token lama ditolak setelah logout-all).

---

## Fase 5 — User Settings: Linked Accounts (Google) ✅ SELESAI (2026-07-10)

**Endpoint:** `GET /user-settings/linked-accounts`, `DELETE /user-settings/linked-accounts/google` (spec #7, #8)

Menyentuh domain Auth (`internal/domain/user`) — Google bukan field scalar, dimodelkan sebagai `Credential{Type: GOOGLE}` + `LoginIdentity{Kind: GOOGLE}` di aggregate `User`.

**Status:** Implementasi selesai persis sesuai draft (tidak ada gap baru), `go build ./...` + `go vet ./...` bersih, 6 test baru pass + full suite `mobile`/`web` handler tetap hijau. File yang dibuat/diubah:

- Domain: `internal/domain/user/constant/user_constant.go` (`CodeUserGoogleUnlinkRequiresPassword` **+ satu tambahan di luar draft**: `CodeUserGoogleNotLinked`, untuk kasus defensif unlink dipanggil saat Google memang belum tertaut — draft tidak menyebutkan case ini secara eksplisit), `internal/domain/user/entity/user.go` (`HasLocalPassword()`, `UnlinkGoogleCredential()`)
- Usecase: `internal/app/usecase/auth/change_password_local.go` (factor-out ke `HasLocalPassword()` sesuai draft), `internal/app/usecase/usersettings/{get_linked_accounts,unlink_google}.go` (baru) + `dependencies.go` (`UserRepo`) + `helpers.go` (`mapUserSettingsAuthError`)
- DTO: `internal/app/dto/user_settings_dto.go` (`GoogleLinkedAccount`, `LinkedAccountsResponse`)
- Handler: `internal/interfaces/http/handler/mobile/user_settings_handler.go` (`GetLinkedAccounts`, `UnlinkGoogle`)
- Router: `internal/interfaces/http/router/router.go` (`protected.Group("/user-settings/linked-accounts")`)
- Wiring: `cmd/app/main.go`, `internal/testhelper/testserver.go` (thread `userRepo` ke `usersettingsUsecase.Dependencies`, route baru di `buildRouter`)
- Test: `internal/interfaces/http/handler/mobile/user_settings_linked_accounts_handler_test.go` (baru — 6 skenario, termasuk helper `mustLinkGoogle`/`mustClearLocalPassword` yang insert langsung ke tabel `credentials`/`user_identities` via SQL karena `noopGoogleVerifier` di test environment selalu error, jadi flow Google login asli tidak bisa dipakai untuk setup test)

**Catatan implementasi penting (soft-delete in-place, bukan hapus dari slice):** `PostgresUserRepository.persistCredentials`/`persistLoginIdentitiesForCredential` hanya **UPSERT** setiap item yang masih ada di `user.Credentials`/`cred.LoginIdentities` slice — kalau item dihapus dari slice, repository tidak pernah tahu harus menghapus/menandai baris itu di DB (tidak ada logic "DELETE yang hilang dari slice"). Jadi `UnlinkGoogleCredential()` **tidak** menghapus credential/identity dari slice, melainkan set `DeletedAt` in-place sambil tetap membiarkannya di slice — supaya `Update()` benar-benar menulis `deleted_at` ke DB. Konsekuensinya: `FindCredential(GOOGLE)` (method yang sudah ada, dipakai di banyak tempat lain) **tidak** memfilter `DeletedAt` — jadi `GetLinkedAccountsUseCase` sengaja cek `google.DeletedAt == nil` secara eksplisit sendiri, bukan mengandalkan `FindCredential` untuk itu (tidak menyentuh/mengubah perilaku `FindCredential` yang sudah ada, biar tidak berisiko ke caller lain).

- `internal/domain/user/entity/user.go`: tambah method `HasLocalPassword() bool` (factor-out dari cek yang sudah ada di `change_password_local.go:67-70` — `FindCredential(LOCAL)` + `SecretHash != nil`) dan `UnlinkGoogleCredential() error` (hapus/soft-delete credential+identity GOOGLE dari aggregate; guard: return domain error kalau `!HasLocalPassword()`).
- `internal/domain/user/constant/user_constant.go`: tambah `CodeUserGoogleUnlinkRequiresPassword`.
- Usecase baru di package `usersettings` (inject `userrepo.UserRepository`, sudah ada):
  - `get_linked_accounts.go` — `FindByID`, derive `linked`/`email`/`can_unlink` dari `FindCredential(GOOGLE)` + `HasLocalPassword()`.
  - `unlink_google.go` — panggil `user.UnlinkGoogleCredential()`; map domain error → `apperror.Conflict` (409) via `helpers.go` (`mapUserSettingsAuthError`, pola CLAUDE.md); sukses → `userRepo.Update(user)`.
- DTO: `LinkedAccountsResponse`.
- Handler + route di `protected.Group("/user-settings/linked-accounts")`.
- Test: 401, 409 (belum punya password), 200 unlink sukses, 200 get status.

---

## Fase 6 — User Settings: Landing Aggregator ✅ SELESAI (2026-07-10)

**Endpoint:** `GET /user-settings` (spec #1) — **harus terakhir**, karena menggabungkan hasil Fase 1 + 4 + 5.

**Status:** Implementasi selesai persis sesuai draft, `go build ./...` + `go vet ./...` bersih, 2 test baru pass (assert ketiga section muncul + is_current ikut kebawa) + full suite `mobile`/`web` handler tetap hijau. **Ini menandai seluruh endpoint inti `API_SPEC_USER_SETTINGS.md` (#1–#8) sudah terimplementasi.** File yang dibuat/diubah:

- Usecase: `internal/app/usecase/usersettings/get_landing.go` (baru — compose `GetPreferencesUseCase`+`ListDevicesUseCase`+`GetLinkedAccountsUseCase` in-process) + `dependencies.go` (`GetLanding` field, dikonstruksi dari instance sibling yang sama, bukan instance baru — supaya tidak ada duplikasi logic)
- DTO: `internal/app/dto/user_settings_dto.go` (`LandingResponse`)
- Handler: `internal/interfaces/http/handler/mobile/user_settings_handler.go` (`GetLanding`)
- Router: `internal/interfaces/http/router/router.go`, `internal/testhelper/testserver.go` (`userSettings.GET("", ...)` — path kosong di dalam grup `/user-settings` yang sudah ada, jadi persis `/api/v1/user-settings`)
- Test: `internal/interfaces/http/handler/mobile/user_settings_landing_handler_test.go` (baru — reuse helper `mustLinkGoogle`/`mustRegisterDeviceWithDeviceID` dari Fase 4/5)

**Catatan implementasi:** `app_preferences` di response landing tetap menyertakan `available_languages` (field yang sama dipakai `GET /user-settings/preferences` biasa) karena `GetLandingUseCase` reuse `GetPreferencesUseCase` apa adanya — sedikit lebih dari contoh minimal di dokumen spec, tapi field tambahan yang harmless dan menghindari duplikasi logic mapping preferences.

- Usecase `get_landing.go` di package `usersettings` — compose langsung 3 usecase sibling (`GetPreferencesUseCase`, `ListDevicesUseCase`, `GetLinkedAccountsUseCase`) in-process, tanpa HTTP round-trip.
- Terima `device_id` query param opsional juga (untuk `is_current` di list device ringkas).
- Handler + route `protected.GET("/user-settings", ...)`.
- Test: 401, 200 (assert ketiga section muncul).

---

## Fase 7a — Event & Reminder: normalisasi ke UTC (absolute instant)

**Prinsip:** event punya satu momen nyata yang tetap, siapa pun yang lihat. Timezone device **organizer** (bukan device semua penonton) cuma dipakai **sekali**, saat create/update, untuk menerjemahkan input wall-clock mereka ke instant UTC yang benar. Setelah itu, evaluasi "apakah reminder harus fire sekarang" murni perbandingan UTC vs UTC — tidak butuh timezone siapa pun lagi saat runtime.

Kondisi sekarang: `events.event_date DATE` + `events.event_time VARCHAR(5)` (`internal/migrations/0005_create_event_tables.up.sql:30,32`), digabung lewat `combineDateAndTime()` (`internal/app/usecase/event/helpers.go:265`) yang selalu mengasumsikan `jakartaLocation` — dipakai di `schedule_event.go:104` (bikin `scheduled_job` reminder) dan `get_calendar_export.go:39,45` (ekspor ICS).

- **Migration data** (`make migrate-create NAME=add_event_starts_at_utc`): tambah kolom `events.starts_at_utc TIMESTAMPTZ`, `events.ends_at_utc TIMESTAMPTZ NULL`. Backfill one-time: `starts_at_utc = event_date + event_time interpreted sebagai Asia/Jakarta` (pola `combineDateAndTime` yang sudah ada, dijalankan sebagai migration data atau job satu-kali) — **additive**, tidak menghapus `event_date`/`event_time` lama dulu (biar rollback aman & UI lama yang masih baca field lama tidak patah selama masa transisi).
- **Input path**: saat create/update event, organizer's device timezone (dari `device_registrations` device organizer yang paling aktif, via `DeviceRegistrationRepository.FindActiveByUserID` — sama repo yang dipakai Fase 4) dipakai untuk interpretasi `event_date`+`event_time` → isi `starts_at_utc`. Fallback kalau organizer belum kirim timezone: tetap `Asia/Jakarta` (perilaku sekarang, bukan regresi).
- **Baca path**: `combineDateAndTime()` di `schedule_event.go` dan `get_calendar_export.go` diganti baca `starts_at_utc`/`ends_at_utc` langsung (tidak perlu kombinasi date+time+zone lagi) — reminder job & ICS export jadi murni UTC arithmetic, tidak ada zona yang perlu ditebak lagi saat runtime.
- **Bukan bagian fase ini**: menghapus kolom `event_date`/`event_time` lama, atau mengubah UI input organizer supaya explicit pilih timezone (masih asumsi device organizer = tempat event, cukup untuk kasus saat ini). Itu follow-up terpisah kalau nanti dibutuhkan organizer lintas-negara.
- **Test**: buat event dengan `event_time="10:00"` dari device dengan timezone `Asia/Seoul` → assert `starts_at_utc` = 01:00 UTC (bukan 03:00 UTC yang berarti masih baca sebagai WIB); assert reminder job & ICS export existing tetap benar untuk event lama yang di-backfill dari WIB.

## Fase 7b — DND: dari hardcoded WIB ke lookup dinamis per-user

**Prinsip:** window DND ("22:00–08:00") adalah **rule wall-clock berulang**, bukan instant — jadi **field-nya TETAP tersimpan sebagai string wall-clock, TIDAK dikonversi ke UTC**. Yang berubah cuma: zona yang dipakai untuk membaca "jam berapa sekarang" jadi dinamis per-user (device aktifnya), bukan hardcode `Asia/Jakarta`. Ini yang paling berisiko karena bentuknya paling mirip insiden lama — dieksekusi terakhir, setelah 7a stabil.

- `internal/domain/notification/service/module_preference_gate.go`: `NowInJakarta()` (baris 34) diganti/didampingi `NowInUserTimezone(ctx, userID string) time.Time` yang lookup `device_registrations` aktif terakhir user → fallback `default_timezone` system setting → fallback `Asia/Jakarta` (perilaku sekarang kalau semua lookup gagal — **tidak pernah lebih buruk dari kondisi sekarang**).
- `DefaultModulePreferenceGate` (baris 55-60) inject `notifRepo.DeviceRegistrationRepository` — legal secara layering (CLAUDE.md: domain service boleh depend ke repository interface domain).
- **Restrukturisasi `IsAllowedBulk`** (baris 74, `now := NowInJakarta()` di baris 82 dihitung SEKALI untuk seluruh batch): harus jadi per-user, pakai `DeviceRegistrationRepository.FindActiveByUserIDs` (bulk method yang **sudah ada**, `internal/infrastructure/persistence/postgres_device_repository.go:129-163`) supaya tetap 1 query batch, bukan N+1 — lalu evaluasi window per-user dengan zona masing-masing.
- `Dispatcher.SendNotification` (`internal/app/service/notification/dispatcher.go:373`) yang manggil `FilterChannels(ctx, userID, channels, notifservice.NowInJakarta())` untuk single user diganti jadi `NowInUserTimezone(ctx, userID)`.
- **Test krusial**: user set DND 22:00-08:00 saat device di Jakarta, lalu device aktif pindah ke Seoul (offset +2 jam dari WIB) → assert window ikut jam Seoul (sesuai Use Case 2 di `USER_SETTINGS_RULES.md`), BUKAN tetap freeze ke jam WIB yang sudah dikonversi. Test juga: user dengan `timezone NULL` (app version lama) tetap fallback ke `Asia/Jakarta`, tidak error/panic. Test window yang crossing midnight (22:00-08:00) tetap benar di zona baru manapun.
- **Rollout**: jalankan di belakang flag/percobaan bertahap kalau memungkinkan (mis. aktifkan dulu untuk notifikasi non-krusial, pantau sebelum full rollout) — mengingat riwayat insiden sebelumnya persis di area ini.

## Fase 7c (opsional, belum digarap) — Community Schedule (RRULE)

Modul `internal/domain/schedule` **belum punya konsep timezone sama sekali** (field `Location` yang ada cuma teks venue). Karena recurrence RRULE-nya sendiri kelihatannya belum diimplementasikan penuh, ini di level catatan saja: kalau nanti dibangun, anchor timezone-nya bisa langsung pakai pola yang sama dengan 7b (`device_registrations.timezone` aktif terakhir → `default_timezone` → fallback) sejak awal, tanpa utang seperti DND. Butuh eksplorasi & plan terpisah saat fitur recurrence-nya digarap.

---

## Verifikasi per Fase

- Setiap fase: `go build ./...` lalu `go test ./internal/... -count=1 -timeout 300s` (fokus ke package yang diubah dulu, lalu full suite sebelum fase dianggap selesai).
- Fase 1: `curl` manual — GET preferences (harus auto-create), PUT dengan `language` invalid → 422, PUT dengan `language=null` → balik ke resolve-live.
- Fase 2: register device dengan `timezone`, GET devices (lewat endpoint FCM lama) → pastikan timezone ikut kebawa.
- Fase 3: login → dapat access+refresh token → panggil `/auth/logout` → pastikan request berikutnya dengan access token lama 401; refresh token lama juga ditolak.
- Fase 4: register 2 device (device A jadi current, device B bukan) → `DELETE /user-settings/devices/{A}?device_id=A` → token device A langsung 401 di request berikutnya; device B row `is_active=false` tapi token device B (kalau ada) masih valid sampai expired (sesuai batasan yang diterima). `POST logout-all` → token device A **dan** B langsung 401.
- Fase 5: buat user Google-only (belum set password) → unlink → 409; set password → unlink → 200.
- Fase 6: pastikan response `GET /user-settings` sama datanya dengan gabungan 3 endpoint granular.
- Fase 7a: buat event dari device dengan timezone `Asia/Seoul`, `event_time="10:00"` → assert `starts_at_utc` = 01:00Z; jalankan backfill di data lama → assert reminder job & ICS export existing (device WIB) tidak berubah waktunya dibanding sebelum migrasi.
- Fase 7b: simulasikan device aktif user pindah dari `Asia/Jakarta` ke `Asia/Seoul` → assert evaluasi window DND ikut jam Seoul; assert user dengan `timezone NULL` tetap fallback ke WIB tanpa error; jalankan `IsAllowedBulk` dengan campuran user berbagai timezone → assert tetap 1 query batch (cek lewat log/EXPLAIN, bukan N+1).
- Ikuti §11 CLAUDE.md: tiap handler baru wajib ada test 401/422/404/409(bila ada)/200 di file `*_handler_test.go` sebelum fase dianggap selesai.
