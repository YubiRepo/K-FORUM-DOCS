# Rencana: API Integration Tests per Domain

## Context

Project k-forum-api memiliki 20+ domain dengan 100+ endpoint (web + mobile) namun belum memiliki test di layer HTTP handler. Test yang sudah ada hanya di persistence layer (`postgres_*_test.go`). Rencana ini membangun HTTP integration test menggunakan Go `httptest` + testcontainers yang sudah dipakai di project, dengan pola full-stack: dari request HTTP masuk hingga response keluar, melibatkan middleware JWT, handler, usecase, dan DB nyata.

---

## Pendekatan Teknis

**Strategi utama:** Satu `TestServer` shared per domain test file, diinisialisasi di `TestMain`. Setiap test case menyiapkan data sendiri dan membersihkannya via `t.Cleanup()`.

**Yang tidak di-mock:**
- PostgreSQL (testcontainers)
- Semua repository & usecase

**Yang di-mock/skip:**
- Redis → gunakan **miniredis** (`github.com/alicebob/miniredis/v2`, perlu ditambah ke `go.mod`)
- RabbitMQ → no-op publisher (event publish gagal tidak menghentikan test)
- MinIO, Firebase, SMTP, Google OAuth → no-op atau stub

---

## Infrastruktur Test (File Baru)

### 1. `internal/testhelper/testserver.go`

```go
type TestServer struct {
    DB     *sql.DB
    Engine *gin.Engine
    Redis  *miniredis.Miniredis  // akses langsung untuk inspect/flush
}

// MustStartTestServer membangun server lengkap dengan testcontainer DB.
// Panggil sekali dari TestMain, cleanup otomatis via t.Cleanup.
func MustStartTestServer(t *testing.T) *TestServer

// Helper HTTP request (mengembalikan *httptest.ResponseRecorder)
func (s *TestServer) POST(path string, body any, headers ...map[string]string) *httptest.ResponseRecorder
func (s *TestServer) GET(path string, headers ...map[string]string) *httptest.ResponseRecorder
func (s *TestServer) PUT(path string, body any, headers ...map[string]string) *httptest.ResponseRecorder
func (s *TestServer) DELETE(path string, headers ...map[string]string) *httptest.ResponseRecorder

// BearerHeader membangun header Authorization dari token
func BearerHeader(token string) map[string]string
```

**Cara kerja `MustStartTestServer`:**
1. `testhelper.MustStartTestDB()` → dapat `*sql.DB` + cleanup
2. `miniredis.Run()` → dapat mock Redis
3. Inisialisasi semua repository (`persistence.NewPostgres*Repository(db)`)
4. Inisialisasi semua usecase dengan repo + deps (no-op publisher, stub external)
5. Inisialisasi semua handler (web + mobile)
6. `router.Setup(allHandlers...)` → dapat `*gin.Engine`
7. Return `TestServer{DB, Engine, Redis}`

### 2. `internal/testhelper/fixtures.go` (extended)

Tambah fixture helper yang dibutuhkan test:
- `MustInsertCredential(ctx, t, db, userID, email, hashedPw)` — seed credential user
- `MustInsertSubscriptionPlan(ctx, t, db)` — seed plan untuk test subscription
- `MustInsertRegion(ctx, t, db)` — seed region untuk test komunitas/event
- `MustCreateTestUser(t, srv)` → register + login, kembalikan `(userID, token string)`
- `MustCreateAdminUser(t, srv)` → buat user dengan role admin, kembalikan token

### 3. Update `go.mod`

Tambah dependency:
- `github.com/alicebob/miniredis/v2` — mock Redis

---

## Struktur File Test

```
k-forum-api/internal/interfaces/http/handler/
├── web/
│   ├── auth_handler_test.go          # Phase 1
│   ├── profile_handler_test.go       # Phase 1
│   ├── role_permission_handler_test.go # Phase 1
│   ├── region_handler_test.go        # Phase 2
│   ├── community_handler_test.go     # Phase 2
│   ├── event_handler_test.go         # Phase 2
│   ├── announcement_handler_test.go  # Phase 3
│   ├── subscription_handler_test.go  # Phase 3
│   ├── qna_handler_test.go           # Phase 3
│   ├── directory_handler_test.go     # Phase 4
│   ├── news_handler_test.go          # Phase 4
│   ├── notification_handler_test.go  # Phase 4
│   ├── user_management_handler_test.go # Phase 5
│   ├── ads_handler_test.go           # Phase 5
│   ├── schedule_handler_test.go      # Phase 5
│   ├── reporting_handler_test.go     # Phase 6
│   ├── accounting_handler_test.go    # Phase 6
│   └── system_settings_handler_test.go # Phase 6
└── mobile/
    ├── auth_handler_test.go          # Phase 1
    ├── profile_handler_test.go       # Phase 1
    └── ... (mengikuti fase yang sama)
```

Setiap file menggunakan satu `TestServer` bersama via package-level var, diinisialisasi di `TestMain`:

```go
var testSrv *testhelper.TestServer

func TestMain(m *testing.M) {
    testSrv = testhelper.MustStartTestServer(nil)  // nil → gunakan testing.M
    os.Exit(m.Run())
}
```

---

## Phase 1: auth, profile, role-permission

### `web/auth_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Register sukses | `POST /api/v1/web/auth/register` | 200, token kembali |
| Register email duplikat | `POST /api/v1/web/auth/register` | 409 |
| Register validasi gagal | `POST /api/v1/web/auth/register` | 400 |
| Login sukses | `POST /api/v1/web/auth/login` | 200, access+refresh token |
| Login password salah | `POST /api/v1/web/auth/login` | 401 |
| Refresh token | `POST /api/v1/web/auth/token/refresh` | 200, token baru |
| Me (authenticated) | `GET /api/v1/web/auth/me` | 200, data user |
| Me (tanpa token) | `GET /api/v1/web/auth/me` | 401 |
| Forgot password | `POST /api/v1/web/auth/password/forgot` | 200 |

### `web/profile_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Get profile | `GET /api/v1/web/profile` | 200, data profile |
| Update profile | `PUT /api/v1/web/profile` | 200 |
| Update profile tanpa token | `PUT /api/v1/web/profile` | 401 |
| Change username | `PUT /api/v1/web/profile/username` | 200 |

### `web/role_permission_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List permissions (admin) | `GET /api/v1/web/role-permission/permissions` | 200, list |
| List permissions (non-admin) | `GET /api/v1/web/role-permission/permissions` | 403 |
| Create permission | `POST /api/v1/web/role-permission/permissions` | 201 |
| List roles | `GET /api/v1/web/role-permission/roles` | 200 |
| Create role | `POST /api/v1/web/role-permission/roles` | 201 |
| Assign role ke user | `POST /api/v1/web/role-permission/user-roles` | 201 |

---

## Phase 2: regions, communities, events

### `web/region_handler_test.go`
- CRUD region (superadmin token)
- List regions (semua user)
- Approve/reject member
- Send invitation + accept

### `web/community_handler_test.go`
- Create community (admin)
- List communities
- List + moderate posts
- Suspend community

### `web/event_handler_test.go`
- CRUD event
- Approval workflow (pending → approved)
- List events (dengan filter)

---

## Phase 3: announcements, subscription, qna

### `web/announcement_handler_test.go`
- CRUD announcement
- Publish + archive lifecycle
- Image upload presign

### `web/subscription_handler_test.go`
- CRUD plans + benefits
- Approve/reject subscription request
- History + analytics

### `web/qna_handler_test.go`
- CRUD FAQ categories + items
- Submit question → answer workflow
- Vote helpful

---

## Phase 4: directory, news, notifications

### `web/directory_handler_test.go`
- CRUD merchant
- Approve/reject/ban
- Item + review management

### `web/news_handler_test.go`
- CRUD article + source
- Approve/publish workflow
- Category management

### `web/notification_handler_test.go`
- Send notification ke user
- Broadcast
- History + inbox

---

## Phase 5: user-management, ads, schedule

### `web/user_management_handler_test.go`
- List users
- CRUD user (admin)

### `web/ads_handler_test.go`
- Create + list ads
- Click/impression tracking (mobile)

### `web/schedule_handler_test.go`
- CRUD schedule entries + types
- Status updates

---

## Phase 6: reporting, accounting, system-settings

### `web/reporting_handler_test.go`
- Submit content report
- Submit bug report
- List + moderate reports (admin)

### `web/accounting_handler_test.go`
- CRUD entries + categories
- Analytics overview

### `web/system_settings_handler_test.go`
- Get + update system settings
- Language management
- Legal documents CRUD

---

## Konvensi Skenario per Test

Setiap test file mengikuti struktur:
```go
func TestXxx_Success(t *testing.T)        // happy path
func TestXxx_Unauthorized(t *testing.T)   // tanpa/invalid token → 401
func TestXxx_Forbidden(t *testing.T)      // token valid tapi role kurang → 403
func TestXxx_ValidationError(t *testing.T) // payload tidak valid → 400
func TestXxx_NotFound(t *testing.T)       // resource tidak ada → 404
```

---

## Verifikasi & Cara Menjalankan

```bash
# Semua test (butuh Docker untuk testcontainers)
cd k-forum-api
go test ./internal/interfaces/http/handler/... -v -count=1 -timeout 300s

# Per phase (contoh phase 1)
go test ./internal/interfaces/http/handler/web/ -run "TestAuth|TestProfile|TestRolePermission" -v -timeout 180s

# Tambah target di Makefile:
# test-api: go test ./internal/interfaces/http/handler/... -v -count=1 -timeout 300s
```

**Checklist per fase selesai:**
- [ ] `TestServer` terinisialisasi tanpa error
- [ ] Semua test di file domain lulus
- [ ] Tidak ada race condition (`go test -race`)
- [ ] `t.Cleanup()` membersihkan data test
