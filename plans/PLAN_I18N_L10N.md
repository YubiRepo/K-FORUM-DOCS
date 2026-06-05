# Plan: Localization & Internationalization (i18n/l10n) — k-forum-api

## Context

Semua pesan error dan response di-hardcode dalam Bahasa Indonesia di layer use case.
Tujuan: support multi-bahasa (Indonesia default + English) tanpa merusak arsitektur Clean Architecture / DDD yang sudah ada.

**Insight kritis:**
- Domain errors (556 kode) **tidak dikirim langsung ke user** — use case menangkapnya dan membuat `AppError` baru
- Yang sampai ke user adalah `AppError.Message` — itulah yang perlu di-translate
- Jumlah pesan unik yang perlu di-translate jauh lebih kecil (~50-100 message key) dari 556 domain codes
- Edge case: beberapa use case pakai `apperror.Forbidden(de.Message)` — domain message langsung jadi user message, ini harus dibenahi

**Error flow saat ini:**
```
DomainError{Code, Message} → UseCase (catch + map) → AppError{Code, Message (hardcode ID)} → httperror.Handle() → JSON response
```

**Error flow setelah i18n:**
```
DomainError{Code} → UseCase (catch + map) → AppError{Code, MsgKey} → httperror.Handle() → translate(locale, msgKey) → JSON response
```

---

## Pendekatan: Message Key di AppError, Translate di HTTP Boundary

Domain & App layer **tidak tahu bahasa**. Use case menggunakan **message key** (string konstan) sebagai `Message` di AppError. HTTP layer mentranslate key tersebut menggunakan locale dari request context.

**Konsekuensi penting:** Semua string Bahasa Indonesia yang hardcode di use case (`"identifier atau password salah"`, `"anda tidak memiliki akses"`, dll.) **dihapus dan diganti konstanta key**. Pesan manusianya hidup sepenuhnya di `locales/*.json`. `AppError.Message` bukan lagi pesan yang dikirim ke user, melainkan kunci lookup.

### Mengapa approach ini?
- Terjemahan hanya di satu tempat: `httperror/http_error.go`
- Tidak perlu inject i18n service ke semua use case
- Backward compatible jika key tidak ditemukan di translation file (fallback ke key itu sendiri)
- Sesuai Clean Architecture: translation adalah concern HTTP layer, bukan domain/app
- Tambah bahasa baru cukup tambah file JSON, zero change di use case

---

## Phase 0 — Simplifikasi DomainError (Hapus Message Field)

Field `Message` di `DomainError` redundant karena `Code` sudah deskriptif sendiri (e.g. `CodeUserBanned`). Hapus field ini untuk menyederhanakan domain layer.

### 0.1 Ubah Struct DomainError

File: `internal/domain/errors/error.go`

```go
// SEBELUM
type DomainError struct {
    Code    Code
    Message string   // ← dihapus
    Err     error
}
func New(code Code, message string) *DomainError  // ← signature berubah
func Wrap(code Code, message string, err error) *DomainError  // ← signature berubah

// SESUDAH
type DomainError struct {
    Code Code
    Err  error
}
func New(code Code) *DomainError
func Wrap(code Code, err error) *DomainError
```

`DomainError.Error()` cukup tampilkan `code` (dan wrapped error jika ada) untuk keperluan log.

### 0.2 Update Semua Pemanggil

Cari semua tempat yang memanggil `domainerr.New(code, "message")` dan `domainerr.Wrap(code, "message", err)` di seluruh domain layer dan hapus parameter message:

```go
// SEBELUM
return domainerr.New(CodeUserBanned, "user diblokir")
return domainerr.Wrap(CodeUserQueryFailed, "gagal query user", err)

// SESUDAH
return domainerr.New(CodeUserBanned)
return domainerr.Wrap(CodeUserQueryFailed, err)
```

File yang terdampak: semua entity, value object, dan repository di `internal/domain/**/`.

### 0.3 Hapus Penggunaan `de.Message` di Use Cases

Cari semua `apperror.Xxx(de.Message)` di use cases dan ganti ke key eksplisit:

```go
// SEBELUM
return nil, apperror.Forbidden(de.Message)

// SESUDAH
return nil, apperror.Forbidden(apperror.MsgUserBanned)
```

---

## Phase 1 — Foundation

### 1.1 Dependency

```
go get github.com/nicksnyder/go-i18n/v2
go get golang.org/x/text
```

### 1.2 Message Key Constants

File baru: `internal/app/apperror/msg_key.go`

```go
package apperror

// Message keys untuk translation — dipakai sebagai AppError.Message
const (
    MsgInvalidCredentials    = "err.auth.invalid_credentials"
    MsgUserDeleted           = "err.auth.user_deleted"
    MsgUserBanned            = "err.auth.user_banned"
    MsgUserNoCredential      = "err.auth.no_credential"
    MsgIdentityUnverified    = "err.auth.identity_unverified"
    MsgUnauthorized          = "err.common.unauthorized"
    MsgForbidden             = "err.common.forbidden"
    MsgNotFound              = "err.common.not_found"
    MsgConflict              = "err.common.conflict"
    MsgInternalError         = "err.common.internal"
    MsgInvalidInput          = "err.common.invalid_input"
    // ... tambah sesuai kebutuhan per domain
)
```

### 1.3 Translation Files

```
k-forum-api/
└── locales/
    ├── id.json    # Bahasa Indonesia (default)
    └── en.json    # English
```

Format (go-i18n JSON):
```json
// locales/id.json
{
  "err.auth.invalid_credentials": "Identifier atau password salah.",
  "err.auth.user_deleted": "Akun ini telah dihapus.",
  "err.auth.user_banned": "Akun ini telah diblokir.",
  "err.common.unauthorized": "Anda tidak memiliki akses.",
  "err.common.not_found": "Data tidak ditemukan.",
  "success.auth.login": "Berhasil masuk.",
  "success.auth.register": "Registrasi berhasil."
}
```

### 1.4 I18n Loader (Infrastructure)

File baru: `internal/infrastructure/i18n/bundle.go`

- Load translation files dari `locales/*.json` saat startup
- Expose fungsi `Translate(lang, key string) string` — thread-safe, di-cache
- Fallback: jika key tidak ditemukan, return key itu sendiri (graceful degradation)
- Supported: `id` (default), `en`

### 1.5 Locale Detection Middleware

File baru: `internal/interfaces/http/middleware/locale.go`

Urutan deteksi:
1. Query param `?lang=id`
2. `Accept-Language` header (parse RFC 5646, ambil tag pertama)
3. Default: `id`

```go
func Locale() gin.HandlerFunc {
    return func(c *gin.Context) {
        lang := c.Query("lang")
        if lang == "" {
            lang = parseAcceptLanguage(c.GetHeader("Accept-Language"))
        }
        if !isSupportedLocale(lang) {
            lang = "id"
        }
        c.Set("locale", lang)
        c.Next()
    }
}
```

Daftarkan di router (sebelum auth middleware):
File: `internal/interfaces/http/router/router.go`

---

## Phase 2 — Refactor AppError Messages ke Message Keys

### 2.1 Refactor Use Cases

**Pattern yang perlu diubah** di semua use case (`internal/app/usecase/**/*.go`):

```go
// SEBELUM
return nil, apperror.Unauthorized("identifier atau password salah")
return nil, apperror.Forbidden(de.Message)    // ⚠️ domain message langsung ke user

// SESUDAH
return nil, apperror.Unauthorized(apperror.MsgInvalidCredentials)
return nil, apperror.Forbidden(apperror.MsgUserBanned)    // gunakan key, bukan de.Message
```

**Strategi untuk `apperror.Xxx(de.Message)`:**
- Untuk setiap domain error code yang dipakai langsung, buat mapping eksplisit ke message key
- Contoh di login usecase:
  ```go
  case usererr.CodeUserDeleted:
      return nil, apperror.Forbidden(apperror.MsgUserDeleted)
  case usererr.CodeUserBanned:
      return nil, apperror.Forbidden(apperror.MsgUserBanned)
  ```

### 2.2 Translate di HTTP Error Handler

File yang dimodifikasi: `internal/interfaces/http/httperror/http_error.go`

Di fungsi `Handle()`, setelah mendapatkan `appErr.message`, translate menggunakan locale dari Gin context:

```go
func Handle(c *gin.Context, err error, bundle *i18n.Bundle) {
    appErr := mapError(err)
    locale := c.GetString("locale")   // set oleh locale middleware
    translatedMsg := bundle.Translate(locale, appErr.message)

    payload := any(translatedMsg)
    if appErr.details != nil {
        payload = appErr.details  // validation errors punya format sendiri
    }
    respond.Error(c, appErr.statusCode, payload)
}
```

Inject `*i18n.Bundle` ke `Handle()` via closure atau struct — bundle dibuat sekali saat startup di `cmd/app/main.go`.

### 2.3 Translate Success Messages di Handlers

```go
// SEBELUM
respond.OK(c, "login success", resp)

// SESUDAH
msg := bundle.Translate(c.GetString("locale"), "success.auth.login")
respond.OK(c, msg, resp)
```

---

## Phase 3 — Validation Error Messages

File yang dimodifikasi: `internal/interfaces/http/httperror/validation.go`

Buat validator message translator yang memetakan tag → message key → translated string.

```go
func ParseValidationError(locale string, bundle *i18n.Bundle, errs validator.ValidationErrors) map[string]string {
    result := make(map[string]string)
    for _, e := range errs {
        key := "validation." + e.Tag()   // e.g. "validation.required"
        result[e.Field()] = bundle.Translate(locale, key)
    }
    return result
}
```

Translation entries tambahan:
```json
{
  "validation.required": "Wajib diisi.",
  "validation.email": "Format email tidak valid.",
  "validation.min": "Minimal {{ .Param }} karakter.",
  "validation.max": "Maksimal {{ .Param }} karakter."
}
```

---

## Phase 4 — Email & Notification (Opsional)

### 4.1 Email Templates

```
internal/infrastructure/external/smtp/templates/
├── id/
│   ├── otp.html
│   └── welcome.html
└── en/
    ├── otp.html
    └── welcome.html
```

SMTP service menerima parameter `lang string` untuk memilih template folder.

### 4.2 Push Notification

FCM service menerima `lang string`, ambil teks dari translation bundle untuk judul/body notifikasi.

---

## Phase 5 — User Language Preference (Future)

- Tambah kolom `preferred_language VARCHAR(10)` di tabel `user_profiles`
- Auth middleware load preferensi user dari DB/cache setelah verify JWT
- Inject ke Gin context sebagai override locale
- Locale middleware baca ini sebagai prioritas tertinggi

---

## Urutan Implementasi

| Step | Task | File |
|------|------|------|
| 1 | Hapus `Message` field dari `DomainError`, update signature `New()` & `Wrap()` | `internal/domain/errors/error.go` |
| 2 | Update semua pemanggil `domainerr.New/Wrap` di domain layer | `internal/domain/**/` |
| 3 | Add go-i18n dependency | `go.mod` |
| 4 | Buat message key constants | `internal/app/apperror/msg_key.go` |
| 5 | Buat translation files (id + en) | `locales/*.json` |
| 6 | Buat i18n bundle loader | `internal/infrastructure/i18n/bundle.go` |
| 7 | Buat locale middleware | `internal/interfaces/http/middleware/locale.go` |
| 8 | Daftarkan locale middleware di router | `internal/interfaces/http/router/router.go` |
| 9 | Refactor `httperror.Handle()` untuk translate | `internal/interfaces/http/httperror/http_error.go` |
| 10 | Refactor semua use case: hardcode string → message key, ganti `de.Message` → key eksplisit | `internal/app/usecase/**/*.go` |
| 11 | Translate validation errors | `internal/interfaces/http/httperror/validation.go` |
| 12 | Translate success messages di handlers | `internal/interfaces/http/handler/**/*.go` |
| 13 | Localize email templates | `internal/infrastructure/external/smtp/` |
| 14 | Localize push notifications | `internal/infrastructure/external/fcm/` |

---

## Verifikasi

1. **Unit test i18n bundle**: test `Translate()` untuk berbagai locale dan fallback ke key jika tidak ditemukan
2. **Integration test**: kirim request dengan `Accept-Language: en`, pastikan response dalam English; default tanpa header → Indonesia
3. **Manual test via curl/Postman**:
   - Login salah tanpa header → error dalam Bahasa Indonesia
   - Login salah dengan `Accept-Language: en` → error dalam English
   - Login salah dengan `?lang=en` → error dalam English
4. **Edge case**: key tidak ada di translation file → response menampilkan key itu sendiri (bukan crash)
5. **Lint check**: pastikan tidak ada string hardcode Bahasa Indonesia tersisa di use cases
