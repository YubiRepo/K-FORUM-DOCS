# Plan — Guest Actor Support

## Context

Saat ini semua field actor (`created_by`, `author_id`, `user_id`, dll.) di seluruh domain mengacu ke `users(id)` dengan constraint `NOT NULL REFERENCES users(id)`. Sistem mengharuskan user login (JWT valid) sebelum bisa melakukan interaksi apapun yang melibatkan actor. Endpoint interaktif seluruhnya berada di bawah middleware `JWTAuth`.

Ke depan diperlukan **guest mode**: user yang belum login harus tetap bisa menjadi actor untuk operasi tertentu — misalnya save event, register device FCM, submit QnA, mark read announcement, like/save post komunitas. Dua skenario yang harus di-handle:

1. **Guest tidak pernah login** → data dihapus otomatis setelah TTL berakhir (cascade)
2. **Guest login** → semua data actor di-*merge* (transfer) ke real user ID, kemudian guest record dihapus

---

## Strategi Inti

**Guest user = record nyata di tabel `users`** dengan tambahan kolom `user_type = 'guest'` dan `expires_at` sebagai TTL. Client (mobile) menyimpan opaque token `guest_session_token` yang dikirim via header `X-Guest-Token`. Semua FK yang sudah ada (`REFERENCES users(id)`) tidak perlu diubah — guest ID tetap UUID valid di tabel `users`. Saat login, jalankan UPDATE bulk untuk transfer semua actor references ke real user ID, lalu DELETE guest record (CASCADE menangani tabel anak secara otomatis).

---

## Perubahan Per Layer

### 1. Migration: `0017_add_guest_user_support.up.sql`

```sql
ALTER TABLE users
    ADD COLUMN user_type           VARCHAR(20) NOT NULL DEFAULT 'registered'
        CHECK (user_type IN ('registered', 'guest')),
    ADD COLUMN guest_session_token VARCHAR(255) UNIQUE,
    ADD COLUMN expires_at          TIMESTAMPTZ;

CREATE INDEX idx_users_user_type   ON users (user_type);
CREATE INDEX idx_users_expires_at  ON users (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_users_guest_token ON users (guest_session_token) WHERE guest_session_token IS NOT NULL;
```

Down migration: DROP INDEX ketiga, DROP ketiga kolom.

Guest rows memakai synthetic email `guest_<uuid>@guest.local` dan username `guest_<uuid[:8]>` agar constraint NOT NULL dan UNIQUE pada `users.email` / `users.username` tidak perlu diubah.

---

### 2. Domain Layer

**`internal/domain/user/constant/`** — tambahkan:
```go
type UserType string
const (
    UserTypeRegistered UserType = "registered"
    UserTypeGuest      UserType = "guest"
)
// Error codes baru:
// CodeGuestUserExpired, CodeGuestSessionTokenRequired
```

**`internal/domain/user/entity/user.go`** — tambahkan fields ke struct `User`:
```go
UserType           constant.UserType
GuestSessionToken  *string
ExpiresAt          *time.Time
```

Tambahkan constructor dan methods:
```go
func NewGuestUser(id, sessionToken string, ttl time.Duration) (*User, error)
func (u *User) IsGuest() bool
func (u *User) IsExpired() bool
func (u *User) EnsureNotExpired() error
```

`NewUser` yang ada tidak berubah — default `UserType = UserTypeRegistered`.

---

### 3. Repository Layer

**`internal/domain/user/repository/interfaces.go`** — tambahkan 4 method baru:
```go
SaveGuest(ctx context.Context, user *entity.User) error
FindByGuestSessionToken(ctx context.Context, token string) (*entity.User, error)
MergeGuestToUser(ctx context.Context, guestID, realUserID string) error
DeleteExpiredGuests(ctx context.Context, before time.Time) (int64, error)
```

**`internal/infrastructure/persistence/postgres_user_repository.go`**

- Update semua SELECT / scan function untuk include kolom `user_type, guest_session_token, expires_at`
- `SaveGuest`: INSERT dengan synthetic email/username + guest fields
- `FindByGuestSessionToken`: `SELECT ... WHERE guest_session_token = $1 AND deleted_at IS NULL`
- `MergeGuestToUser` — jalankan dalam transaction caller menggunakan `execFromContext`:
  - Tabel dengan **simple UPDATE** (tidak ada unique constraint per-user):
    - `device_registrations.user_id`
    - `notification_preferences.user_id`
    - `community_posts.author_id`
    - `community_post_comments.author_id`
    - `qna_questions.user_id`
  - Tabel dengan **INSERT + DELETE** (ada `UNIQUE (entity_id, user_id)` — hindari conflict):
    - `community_post_likes`, `community_post_saves`, `event_saves`, `announcement_reads`
    ```sql
    -- Pattern INSERT+DELETE untuk junction tables:
    INSERT INTO community_post_likes (id, post_id, user_id, created_at)
    SELECT id, post_id, $realUserID, created_at
    FROM community_post_likes WHERE user_id = $guestID
    ON CONFLICT (post_id, user_id) DO NOTHING;
    DELETE FROM community_post_likes WHERE user_id = $guestID;
    ```
  - Setelah semua tabel: `DELETE FROM users WHERE id = $guestID` (CASCADE bersihkan credentials/identities/dll)
- `DeleteExpiredGuests`: `DELETE FROM users WHERE user_type = 'guest' AND expires_at < $before`

---

### 4. Port

Tambahkan **`internal/app/port/guest_session_resolver.go`**:
```go
type GuestSessionResolver interface {
    FindByGuestSessionToken(ctx context.Context, token string) (*userEntity.User, error)
}
```
`PostgresUserRepository` otomatis satisfy interface ini setelah implementasi di atas.

---

### 5. Application Layer — Usecases

**Baru: `internal/app/usecase/auth/create_guest_session.go`**
- Input: tidak ada
- Proses: generate `sessionToken` via `crypto/rand` (32 bytes, hex), buat `NewGuestUser(uuid, token, 30*24*time.Hour)`, simpan via `userRepo.SaveGuest`
- Output DTO baru `GuestSessionResponse { GuestID, GuestSessionToken, ExpiresAt }`

**Modifikasi: `internal/app/usecase/auth/login.go`**
- Tambahkan field `transactor port.Transactor` ke struct `LoginUseCase` dan inject di constructor
- Tambahkan parameter `guestToken string` ke `Execute` (string kosong = tidak ada guest session)
- Setelah login berhasil dan JWT dibuat, jika `guestToken != ""`:
  ```go
  func (uc *LoginUseCase) mergeGuestIfPresent(ctx, guestToken, realUserID string) error {
      guest, _ := uc.userRepo.FindByGuestSessionToken(ctx, guestToken)
      if guest == nil || !guest.IsGuest() || guest.IsExpired() { return nil }
      return uc.transactor.WithTx(ctx, func(txCtx context.Context) error {
          return uc.userRepo.MergeGuestToUser(txCtx, guest.ID, realUserID)
      })
  }
  ```
  Merge bersifat **best-effort** — jika gagal log error, tapi login tetap berhasil.

**Modifikasi: `internal/app/usecase/auth/google_login.go`** — pola identik dengan login.go

**DTO: `internal/app/dto/auth_dto.go`**
- Tambah struct `GuestSessionResponse`
- Tambah field opsional ke `LoginRequest`: `GuestToken *string \`json:"guest_token,omitempty"\``

---

### 6. Middleware

**`internal/interfaces/http/middleware/auth.go`** — tambahkan `GuestOrJWTAuth`:
```go
func GuestOrJWTAuth(tokenGen port.TokenGenerator, guestResolver port.GuestSessionResolver) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Coba JWT dulu (path tidak berubah dari JWTAuth)
        if authHeader := c.GetHeader("Authorization"); authHeader != "" {
            // parse + set context seperti JWTAuth, lalu return
        }
        // 2. Coba guest token
        if token := c.GetHeader("X-Guest-Token"); token != "" {
            user, err := guestResolver.FindByGuestSessionToken(c.Request.Context(), token)
            if err == nil && user != nil && !user.IsExpired() {
                c.Set("user_id", user.ID)
                c.Set("is_guest", true)
                c.Set("roles", []string{})
                c.Set("permissions", []string{})
                c.Set("plans", []string{})
                c.Set("benefits", []string{})
            }
        }
        c.Next()
    }
}

func IsGuestFromContext(c *gin.Context) bool { ... }
```

---

### 7. Router

**`internal/interfaces/http/router/router.go`**

- Tambah parameter `guestResolver port.GuestSessionResolver` ke fungsi `Setup`
- Tambah route publik di grup `mobileAuth`:
  ```go
  mobileAuth.POST("/guest-session", mobileAuthHandler.CreateGuestSession)
  ```
- Buat grup baru untuk endpoint yang support guest:
  ```go
  optionalMobile := mobile.Group("")
  optionalMobile.Use(middleware.GuestOrJWTAuth(tokenGen, guestResolver))
  ```
- **Pindahkan** routes berikut dari `protectedMobile` ke `optionalMobile`:

  | Route | Alasan |
  |---|---|
  | `POST /mobile/fcm/register` | Guest perlu daftar device untuk notifikasi push |
  | `PUT /mobile/fcm/update` | Sama |
  | `DELETE /mobile/fcm/revoke` | Sama |
  | `POST /mobile/events/:event_id/save` | Guest bisa simpan event |
  | `DELETE /mobile/events/:event_id/save` | Sama |
  | `POST /mobile/communities/posts/:post_id/like` | Guest bisa like post |
  | `DELETE /mobile/communities/posts/:post_id/like` | Sama |
  | `POST /mobile/communities/posts/:post_id/save` | Guest bisa save post |
  | `DELETE /mobile/communities/posts/:post_id/save` | Sama |
  | `POST /mobile/qna/questions` | Guest bisa submit pertanyaan |
  | `POST /mobile/announcements/:announcement_id/read` | Guest bisa mark read |

  Semua route yang butuh real user (profile, subscription, moderation, dll.) **tetap di `protectedMobile`**.

---

### 8. HTTP Handler

**`internal/interfaces/http/handler/mobile/auth_handler.go`**
- Tambahkan method `CreateGuestSession` yang memanggil `CreateGuestSessionUseCase`
- Update `Login` dan `GoogleLogin` handler untuk ekstrak `X-Guest-Token` dari header dan teruskan ke usecase

---

### 9. Cleanup Worker

**Baru: `internal/interfaces/mq/relay/guest_cleanup_relay.go`**
- Ikuti pola `delivery_retry_relay.go`
- Setiap interval (default 1 jam): panggil `userRepo.DeleteExpiredGuests(ctx, time.Now())`

**`cmd/worker/main.go`** — tambahkan:
```go
go relay.NewGuestCleanupRelay(userRepo, 1*time.Hour, logger).Start(ctx)
```

---

### 10. Wiring (`cmd/app/main.go`)

1. Instantiate `createGuestSessionUC` dengan TTL 30 hari dari config
2. Inject `transactor` ke `NewLoginUseCase` (param baru)
3. Teruskan `userRepo` sebagai `guestResolver` ke `router.Setup`
4. Teruskan `createGuestSessionUC` ke `mobileAuthHandler` constructor

---

## Urutan Implementasi

| # | Layer | File |
|---|---|---|
| 1 | Migration | `0017_add_guest_user_support.up/down.sql` |
| 2 | Domain constant | `domain/user/constant/` |
| 3 | Domain entity | `domain/user/entity/user.go` |
| 4 | Repository interface | `domain/user/repository/interfaces.go` |
| 5 | Repository impl | `infrastructure/persistence/postgres_user_repository.go` |
| 6 | Port | `app/port/guest_session_resolver.go` |
| 7 | Usecase baru | `app/usecase/auth/create_guest_session.go` |
| 8 | DTO | `app/dto/auth_dto.go` |
| 9 | Usecase modifikasi | `app/usecase/auth/login.go` + `google_login.go` |
| 10 | Middleware | `interfaces/http/middleware/auth.go` |
| 11 | Handler | `interfaces/http/handler/mobile/auth_handler.go` |
| 12 | Router | `interfaces/http/router/router.go` |
| 13 | Cleanup relay | `interfaces/mq/relay/guest_cleanup_relay.go` |
| 14 | Wiring | `cmd/app/main.go` + `cmd/worker/main.go` |

---

## Verifikasi

1. **Migration**: `go run cmd/migrate/main.go up` → kolom dan index terbentuk di tabel `users`
2. **Create guest session**: `POST /api/v1/mobile/auth/guest-session` → dapat `guest_session_token` + `expires_at`
3. **Guest interaction**: `POST /mobile/events/:id/save` dengan header `X-Guest-Token: <token>` → sukses tanpa JWT
4. **Guest merge saat login**: Login dengan body `{ ..., "guest_token": "<token>" }` → cek di DB bahwa data guest (event_saves, dll.) berpindah ke real user ID dan guest record terhapus
5. **TTL cleanup**: Set `expires_at` ke masa lalu manual, jalankan relay → data terhapus cascade
6. **JWT-only endpoint tidak terpengaruh**: `GET /mobile/auth/me` tanpa JWT tetap 401
7. **Duplicate handling saat merge**: Buat data identik di guest dan real user (misal like post yang sama) → merge tidak menghasilkan unique constraint violation
