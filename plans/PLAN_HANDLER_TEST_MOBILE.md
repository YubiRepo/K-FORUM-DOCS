# Rencana: Mobile API Integration Tests per Domain

## Context

Mobile handler di k-forum-api memiliki **18 domain** dengan **120+ endpoint** namun belum memiliki satu pun test di layer HTTP handler. Web handler sudah memiliki 20 test file lengkap (Phase 1–6 selesai semua). Rencana ini membangun mobile HTTP integration test dengan pola dan infrastruktur yang sudah ada — tidak perlu infrastruktur baru, hanya file test baru di `internal/interfaces/http/handler/mobile/`.

---

## Pendekatan Teknis

**Reuse penuh infrastruktur web:**
- `testhelper.MustStartTestServer()` — sudah serve mobile routes (buildRouter includes web + mobile handlers)
- `testhelper.BearerHeader()`, `seed.go`, `fixtures.go`, `noop.go` — sama persis
- Pola `TestMain` + `testutil_test.go` — mirror dari `web/main_test.go` dan `web/testutil_test.go`
- Package: `package mobile_test`
- URL prefix: `/api/v1/mobile`

**Mobile-specific nuances:**
- **OTP flow** — `noopSMSSender` tidak mengirim SMS nyata; test hanya verifikasi status 200 pada `request-otp`, lalu query OTP dari DB via `testSrv.DB` untuk `verify-otp`
- **Benefit-gated endpoints** — `CreateCommunity`, `CreatePost`, `CreateEvent`, `SubmitArticle`, `SubmitQuestion` butuh user dengan subscription aktif; tambah helper `mustCreateSubscribedUser(t)` yang seed plan + approve request langsung via DB
- **Presign → Confirm uploads** — test presign (200 + URL returned), skip actual S3 confirm karena `noopFileUploader` sudah return dummy URL
- **Google OAuth** — `noopGoogleVerifier` selalu error; test hanya verifikasi 401/502 response
- **Seamless token** — `GenerateSeamlessToken` test success path saja

---

## Infrastruktur yang Perlu Ditambah

### `mobile/main_test.go` (baru)
```go
var (
    testSrv        *testhelper.TestServer
    testCleanup    func()
)

func TestMain(m *testing.M) {
    testSrv, testCleanup = testhelper.MustStartTestServer()
    code := m.Run()
    testCleanup()
    os.Exit(code)
}
```

### `mobile/testutil_test.go` (baru)
```go
// uniqueEmail(prefix) string
// uniqueUsername(prefix) string
// mustRegisterUser(t, email, username, password) string  → userID
// mustLogin(t, identifier, password) (accessToken, refreshToken string)
// mustCreateSubscribedUser(t) (userID, token string)
//   → register + login + seed subscription plan + approve via DB
// mustGetOTPFromDB(t, userID, otpType) string
//   → query tabel OTP di testSrv.DB untuk verify-otp flow
```

Tidak perlu update `go.mod` — semua dependency sudah ada.

---

## Struktur File Test

```
k-forum-api/internal/interfaces/http/handler/mobile/
├── main_test.go                          # TestMain + shared vars
├── testutil_test.go                      # helper functions
├── auth_handler_test.go                  # Phase 1
├── profile_handler_test.go               # Phase 1
├── region_handler_test.go                # Phase 2
├── community_handler_test.go             # Phase 2
├── event_handler_test.go                 # Phase 3
├── news_handler_test.go                  # Phase 3
├── announcement_handler_test.go          # Phase 4
├── notification_handler_test.go          # Phase 4
├── notification_preference_handler_test.go # Phase 4
├── directory_handler_test.go             # Phase 5
├── ads_handler_test.go                   # Phase 5
├── subscription_handler_test.go          # Phase 6
├── qna_handler_test.go                   # Phase 6
├── device_handler_test.go                # Phase 6
├── bug_report_handler_test.go            # Phase 6
├── content_report_handler_test.go        # Phase 6
├── config_handler_test.go                # Phase 6
└── legal_handler_test.go                 # Phase 6
```

---

## Phase 1: auth, profile

### `mobile/auth_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Register sukses | `POST /api/v1/mobile/auth/register` | 201, user_id returned |
| Register email duplikat | `POST /api/v1/mobile/auth/register` | 409 |
| Register validasi gagal | `POST /api/v1/mobile/auth/register` | 400 |
| Login sukses | `POST /api/v1/mobile/auth/login` | 200, access+refresh token |
| Login password salah | `POST /api/v1/mobile/auth/login` | 401 |
| Google login (no-op) | `POST /api/v1/mobile/auth/login/google` | 401/502 |
| Refresh token | `POST /api/v1/mobile/auth/refresh-token` | 200, token baru |
| Request OTP | `POST /api/v1/mobile/auth/request-otp` | 200 |
| Verify OTP | `POST /api/v1/mobile/auth/verify-otp` | 200 (OTP dari DB) |
| Forgot password | `POST /api/v1/mobile/auth/password/forgot` | 200 |
| Reset password | `POST /api/v1/mobile/auth/password/reset` | 200 |
| Me (authenticated) | `GET /api/v1/mobile/auth/me` | 200, data user |
| Me (tanpa token) | `GET /api/v1/mobile/auth/me` | 401 |
| Change password | `POST /api/v1/mobile/auth/change-password` | 200 |
| Request change email | `POST /api/v1/mobile/auth/change-email/request` | 200 |
| Generate seamless token | `POST /api/v1/mobile/auth/seamless/generate-token` | 200, token |

### `mobile/profile_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Get profile | `GET /api/v1/mobile/profile/me` | 200, data profile |
| Update profile | `PATCH /api/v1/mobile/profile/me` | 200 |
| Change username | `PATCH /api/v1/mobile/profile/username` | 200 |
| Change username duplikat | `PATCH /api/v1/mobile/profile/username` | 409 |
| Get avatar presign URL | `POST /api/v1/mobile/profile/avatar` | 200, presign URL |
| Confirm avatar | `PUT /api/v1/mobile/profile/avatar` | 200 |
| Delete avatar | `DELETE /api/v1/mobile/profile/avatar` | 200 |
| Get subscription | `GET /api/v1/mobile/profile/subscription` | 200 |
| Get region memberships | `GET /api/v1/mobile/profile/memberships/regions` | 200 |
| Unauthorized | `GET /api/v1/mobile/profile/me` (no token) | 401 |

---

## Phase 2: region, community

### `mobile/region_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List regions | `GET /api/v1/mobile/regions` | 200, list |
| Get region detail | `GET /api/v1/mobile/regions/{id}` | 200, data region |
| Request join region | `POST /api/v1/mobile/regions/{id}/join` | 200/201 |
| Cancel join request | `DELETE /api/v1/mobile/regions/{id}/join` | 200 |
| Get my region | `GET /api/v1/mobile/region/me` | 200 |
| List region members | `GET /api/v1/mobile/regions/{id}/members` | 200 |
| List pending invitations | `GET /api/v1/mobile/region/invitations` | 200 |
| Accept invitation | `POST /api/v1/mobile/region/invitations/{id}/accept` | 200 |
| Reject invitation | `POST /api/v1/mobile/region/invitations/{id}/reject` | 200 |
| Leave region | `POST /api/v1/mobile/regions/{id}/leave` | 200 |

### `mobile/community_handler_test.go`

Community adalah domain terbesar (35+ endpoint). Fokus pada alur utama:

| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List categories | `GET /api/v1/mobile/communities/categories` | 200 |
| Browse communities | `GET /api/v1/mobile/communities` | 200, list |
| Create community (subscribed) | `POST /api/v1/mobile/communities` | 201 |
| Create community (no benefit) | `POST /api/v1/mobile/communities` | 403 |
| Get community | `GET /api/v1/mobile/communities/{id}` | 200 |
| Update community | `PATCH /api/v1/mobile/communities/{id}` | 200 |
| Delete community | `DELETE /api/v1/mobile/communities/{id}` | 200 |
| Join community | `POST /api/v1/mobile/communities/{id}/join` | 200/201 |
| Leave community | `DELETE /api/v1/mobile/communities/{id}/membership` | 200 |
| List members | `GET /api/v1/mobile/communities/{id}/members` | 200 |
| List join requests | `GET /api/v1/mobile/communities/{id}/join-requests` | 200 |
| Approve join request | `POST /api/v1/mobile/communities/{id}/join-requests/{req_id}/approve` | 200 |
| Get community feed | `GET /api/v1/mobile/communities/{id}/posts` | 200 |
| Create post (subscribed) | `POST /api/v1/mobile/communities/{id}/posts` | 201 |
| Create post (no benefit) | `POST /api/v1/mobile/communities/{id}/posts` | 403 |
| Get post | `GET /api/v1/mobile/communities/posts/{post_id}` | 200 |
| Delete post | `DELETE /api/v1/mobile/communities/posts/{post_id}` | 200 |
| Like post | `POST /api/v1/mobile/communities/posts/{post_id}/like` | 200 |
| Unlike post | `DELETE /api/v1/mobile/communities/posts/{post_id}/like` | 200 |
| List comments | `GET /api/v1/mobile/communities/posts/{post_id}/comments` | 200 |
| Create comment (subscribed) | `POST /api/v1/mobile/communities/posts/{post_id}/comments` | 201 |
| Save post | `POST /api/v1/mobile/communities/posts/{post_id}/save` | 200 |
| Unsave post | `DELETE /api/v1/mobile/communities/posts/{post_id}/save` | 200 |
| Get saved posts | `GET /api/v1/mobile/communities/posts/saved` | 200 |

---

## Phase 3: event, news

### `mobile/event_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List events | `GET /api/v1/mobile/events` | 200 |
| List categories | `GET /api/v1/mobile/events/categories` | 200 |
| Create event (subscribed) | `POST /api/v1/mobile/events` | 201 |
| Create event (no benefit) | `POST /api/v1/mobile/events` | 403 |
| Get event | `GET /api/v1/mobile/events/{id}` | 200 |
| Update event | `PUT /api/v1/mobile/events/{id}` | 200 |
| Cancel event | `POST /api/v1/mobile/events/{id}/cancel` | 200 |
| List my events | `GET /api/v1/mobile/events/my` | 200 |
| Save event | `POST /api/v1/mobile/events/{id}/save` | 200 |
| Unsave event | `DELETE /api/v1/mobile/events/{id}/unsave` | 200 |
| List saved events | `GET /api/v1/mobile/events/my/saved` | 200 |
| Schedule event | `POST /api/v1/mobile/events/{id}/schedule` | 200 |
| Unschedule event | `DELETE /api/v1/mobile/events/{id}/unschedule` | 200 |
| Get calendar export | `GET /api/v1/mobile/events/{id}/calendar-export` | 200 |
| Share event | `POST /api/v1/mobile/events/{id}/share` | 200 |
| Get image presign | `POST /api/v1/mobile/events/images/presign` | 200, URL |

### `mobile/news_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List articles (public) | `GET /api/v1/mobile/news/articles` | 200 |
| Get article (public) | `GET /api/v1/mobile/news/articles/{id}` | 200 |
| List categories (public) | `GET /api/v1/mobile/news/categories` | 200 |
| List comments (public) | `GET /api/v1/mobile/news/articles/{id}/comments` | 200 |
| Like article | `POST /api/v1/mobile/news/articles/{id}/like` | 200 |
| Unlike article | `DELETE /api/v1/mobile/news/articles/{id}/like` | 200 |
| Bookmark article | `POST /api/v1/mobile/news/articles/{id}/bookmark` | 200 |
| Unbookmark article | `DELETE /api/v1/mobile/news/articles/{id}/bookmark` | 200 |
| List bookmarks | `GET /api/v1/mobile/news/bookmarks` | 200 |
| Post comment | `POST /api/v1/mobile/news/articles/{id}/comments` | 201 |
| Delete comment | `DELETE /api/v1/mobile/news/comments/{id}` | 200 |
| Submit article (subscribed) | `POST /api/v1/mobile/news/articles` | 201 |
| Submit article (no benefit) | `POST /api/v1/mobile/news/articles` | 403 |
| Withdraw article | `DELETE /api/v1/mobile/news/articles/{id}/withdraw` | 200 |
| Unauthorized | `POST /api/v1/mobile/news/articles/{id}/like` (no token) | 401 |

---

## Phase 4: announcement, notification, notification_preference

### `mobile/announcement_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List announcements | `GET /api/v1/mobile/announcements` | 200 |
| Get announcement | `GET /api/v1/mobile/announcements/{id}` | 200 |
| Mark read | `POST /api/v1/mobile/announcements/{id}/mark-read` | 200 |
| Get unread count | `GET /api/v1/mobile/announcements/unread-count` | 200, count |
| List by type | `GET /api/v1/mobile/announcements/type/{type}` | 200 |
| List critical | `GET /api/v1/mobile/announcements/critical` | 200 |
| Search | `GET /api/v1/mobile/announcements/search?q=...` | 200 |
| Unauthorized | `GET /api/v1/mobile/announcements` (no token) | 401 |

### `mobile/notification_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List notifications | `GET /api/v1/mobile/notifications` | 200 |
| Get unread count | `GET /api/v1/mobile/notifications/unread-count` | 200, count |
| Mark read | `PATCH /api/v1/mobile/notifications/{id}/read` | 200 |
| Mark all read | `PATCH /api/v1/mobile/notifications/read-all` | 200 |
| Unauthorized | `GET /api/v1/mobile/notifications` (no token) | 401 |

### `mobile/notification_preference_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Get preferences | `GET /api/v1/mobile/notifications/preferences` | 200 |
| Update global | `PUT /api/v1/mobile/notifications/preferences/global` | 200 |
| Update module | `PUT /api/v1/mobile/notifications/preferences/modules/{module}` | 200 |
| Reset preferences | `POST /api/v1/mobile/notifications/preferences/reset` | 200 |

---

## Phase 5: directory, ads

### `mobile/directory_handler_test.go`

Domain besar; fokus pada alur merchant dan fitur sosial:

| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List categories | `GET /api/v1/mobile/directory/categories` | 200 |
| Browse merchants | `GET /api/v1/mobile/directory/merchants` | 200 |
| Create merchant (subscribed) | `POST /api/v1/mobile/directory/merchants` | 201 |
| Create merchant (no benefit) | `POST /api/v1/mobile/directory/merchants` | 403 |
| Get merchant detail | `GET /api/v1/mobile/directory/merchants/{id}` | 200 |
| Update merchant | `PUT /api/v1/mobile/directory/merchants/{id}` | 200 |
| Submit merchant | `POST /api/v1/mobile/directory/merchants/{id}/submit` | 200 |
| Archive merchant | `POST /api/v1/mobile/directory/merchants/{id}/archive` | 200 |
| Get my merchants | `GET /api/v1/mobile/directory/me/merchants` | 200 |
| Create item | `POST /api/v1/mobile/directory/merchants/{id}/items` | 201 |
| Get merchant items (public) | `GET /api/v1/mobile/directory/merchants/{id}/items` | 200 |
| Update item | `PUT /api/v1/mobile/directory/merchants/{id}/items/{item_id}` | 200 |
| Leave review | `POST /api/v1/mobile/directory/merchants/{id}/reviews` | 201 |
| Get merchant reviews (public) | `GET /api/v1/mobile/directory/merchants/{id}/reviews` | 200 |
| Vote review | `POST /api/v1/mobile/directory/reviews/{id}/vote` | 200 |
| Save merchant | `POST /api/v1/mobile/directory/merchants/{id}/save` | 200 |
| Unsave merchant | `DELETE /api/v1/mobile/directory/merchants/{id}/save` | 200 |
| Get saved merchants | `GET /api/v1/mobile/directory/me/saved` | 200 |
| Send inquiry | `POST /api/v1/mobile/directory/merchants/{id}/inquiries` | 201 |
| Get my inquiries | `GET /api/v1/mobile/directory/me/inquiries` | 200 |
| Get presign URL | `POST /api/v1/mobile/directory/merchants/images/presign` | 200 |

### `mobile/ads_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Get home ads (public) | `GET /api/v1/mobile/ads/home` | 200 |
| List active ads (public) | `GET /api/v1/mobile/ads` | 200 |
| Record impression (public) | `POST /api/v1/mobile/ads/impression` | 200 |
| Record click (public) | `POST /api/v1/mobile/ads/click` | 200 |
| Create ad | `POST /api/v1/mobile/ads` | 201 |
| List my ads | `GET /api/v1/mobile/ads/my` | 200 |
| Get my ad | `GET /api/v1/mobile/ads/{id}` | 200 |
| Update ad | `PUT /api/v1/mobile/ads/{id}` | 200 |
| Delete ad | `DELETE /api/v1/mobile/ads/{id}` | 200 |
| Update ad status | `PATCH /api/v1/mobile/ads/{id}/status` | 200 |
| Get ad analytics | `GET /api/v1/mobile/ads/{id}/analytics` | 200 |
| Get image presign | `POST /api/v1/mobile/ads/media/image/presign` | 200, URL |

---

## Phase 6: subscription, qna, device, bug_report, content_report, config, legal

### `mobile/subscription_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Get my subscription | `GET /api/v1/mobile/subscription/me` | 200 |
| List plans (public) | `GET /api/v1/mobile/subscription/plans` | 200 |
| Create subscription request | `POST /api/v1/mobile/subscription/request` | 201 |
| Get my request | `GET /api/v1/mobile/subscription/request/{id}` | 200 |
| Cancel request | `POST /api/v1/mobile/subscription/request/{id}/cancel` | 200 |
| Get my benefits | `GET /api/v1/mobile/subscription/benefits` | 200 |
| Verify benefit | `POST /api/v1/mobile/subscription/verify-benefit` | 200 |
| Get history | `GET /api/v1/mobile/subscription/history` | 200 |
| Get proof presign | `POST /api/v1/mobile/subscription/proof/presign` | 200 |

### `mobile/qna_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| List categories | `GET /api/v1/mobile/qna/categories` | 200 |
| List FAQ items | `GET /api/v1/mobile/qna/faq` | 200 |
| Get FAQ item | `GET /api/v1/mobile/qna/faq/{id}` | 200 |
| Vote FAQ helpful | `POST /api/v1/mobile/qna/faq/{id}/vote` | 200 |
| Submit question (subscribed) | `POST /api/v1/mobile/qna/questions` | 201 |
| Submit question (no benefit) | `POST /api/v1/mobile/qna/questions` | 403 |
| List my questions | `GET /api/v1/mobile/qna/questions/me` | 200 |
| Get question detail | `GET /api/v1/mobile/qna/questions/me/{id}` | 200 |

### `mobile/device_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Register device | `POST /api/v1/mobile/fcm/register` | 200/201 |
| Refresh device token | `PUT /api/v1/mobile/fcm/update` | 200 |
| Revoke device token | `DELETE /api/v1/mobile/fcm/revoke` | 200 |
| List devices | `GET /api/v1/mobile/fcm/devices` | 200 |
| Remove device | `DELETE /api/v1/mobile/fcm/devices/{id}` | 200 |

### `mobile/bug_report_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Submit bug report | `POST /api/v1/mobile/reports/bug` | 201 |
| Get my reports | `GET /api/v1/mobile/reports/bug/me` | 200 |
| Get report detail | `GET /api/v1/mobile/reports/bug/me/{id}` | 200 |
| Get attachment presign | `POST /api/v1/mobile/reports/bug/attachments/presign` | 200 |

### `mobile/content_report_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Get report reasons | `GET /api/v1/mobile/reports/content/reasons` | 200, list |
| Submit content report | `POST /api/v1/mobile/reports/content` | 201 |
| Get my reports | `GET /api/v1/mobile/reports/content/me` | 200 |

### `mobile/config_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Get public config (no auth needed) | `GET /api/v1/mobile/config` | 200 |

### `mobile/legal_handler_test.go`
| Test | Endpoint | Ekspektasi |
|------|----------|-----------|
| Get published legal (public) | `GET /api/v1/mobile/legal/{doc_type}` | 200 |
| Get pending acceptances | `GET /api/v1/mobile/legal/pending` | 200 |
| Accept legal doc | `POST /api/v1/mobile/legal/{doc_type}/accept` | 200 |

---

## Konvensi Skenario per Test

Sama dengan web tests:
```go
func TestMobileAuth_Login_Success(t *testing.T)
func TestMobileAuth_Login_WrongPassword(t *testing.T)
func TestMobileCommunity_CreatePost_NoSubscription(t *testing.T)
func TestMobileCommunity_CreatePost_Forbidden(t *testing.T)
```

---

## File Kritis yang Perlu Dimodifikasi/Dibuat

| File | Status | Keterangan |
|------|--------|------------|
| `internal/interfaces/http/handler/mobile/main_test.go` | Baru | TestMain + shared vars |
| `internal/interfaces/http/handler/mobile/testutil_test.go` | Baru | helpers: uniqueEmail, mustLogin, mustCreateSubscribedUser, mustGetOTPFromDB |
| `internal/interfaces/http/handler/mobile/auth_handler_test.go` | Baru | Phase 1 |
| `internal/interfaces/http/handler/mobile/profile_handler_test.go` | Baru | Phase 1 |
| ... (16 file test lainnya) | Baru | Phase 2–6 |

Referensi pattern yang wajib diikuti:
- `internal/interfaces/http/handler/web/main_test.go`
- `internal/interfaces/http/handler/web/testutil_test.go`
- `internal/interfaces/http/handler/web/auth_handler_test.go`

---

## Verifikasi

```bash
cd k-forum-api

# Jalankan semua mobile tests
go test ./internal/interfaces/http/handler/mobile/... -v -count=1 -timeout 300s

# Per phase (contoh phase 1)
go test ./internal/interfaces/http/handler/mobile/ -run "TestMobileAuth|TestMobileProfile" -v -timeout 180s

# Race check
go test ./internal/interfaces/http/handler/mobile/... -race -count=1 -timeout 300s
```

**Checklist per fase:**
- [ ] TestServer terinisialisasi tanpa error
- [ ] Semua test di file domain lulus
- [ ] Tidak ada race condition
- [ ] Benefit-gated endpoints diuji dengan dan tanpa subscription
