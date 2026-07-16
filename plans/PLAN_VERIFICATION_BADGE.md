# Plan — Verification Badge Module (KAI App)

Sumber: `K-FORUM-DOCS/Modules/Verifications/` (`VERIFICATION_RULES.md`, `VERIFICATION_DB_SCHEMA.md`, `API_SPEC_VERIFICATION_BACKOFFICE.md`, `API_SPEC_VERIFICATION_MOBILE.md`).

Modul baru: centang keaslian ("Verified") untuk **User** & **Merchant**, diajukan manual, di-review manual oleh Superadmin. Satu mesin (`verifications`), dua produk (`entity_type = user | merchant`). Layer trust terpisah dari `email_verified`/`phone_verified` (OTP) dan merchant `approval_status`.

Plan ini dipecah per layer arsitektur sesuai `CLAUDE.md` k-forum-api: **Domain → Usecase → Interfaces → Infrastructure**.

---

## ⚠️ Keputusan Penamaan (WAJIB dibaca sebelum mulai coding)

`internal/domain/verification/` **sudah dipakai** untuk sistem OTP (`VerificationCode` — email/phone verification, reset password, change identity). Itu konsep yang berbeda total dari modul ini (RULES §1 eksplisit bilang jangan ketuker keduanya).

Untuk menghindari collision paket Go & kebingungan penamaan, modul baru ini pakai nama **`verificationbadge`** di semua layer:

| Layer | Path |
|-------|------|
| Domain | `internal/domain/verificationbadge/` |
| Usecase | `internal/app/usecase/verificationbadge/` |
| DTO | `internal/app/dto/verificationbadge_dto.go` |
| Port (query model) | `internal/app/port/verificationbadge_query_model.go` |
| Persistence | `internal/infrastructure/persistence/postgres_verificationbadge_repository.go` (+ `_query.go`, `_event_repository.go`) |
| Handler | `.../handler/mobile/verificationbadge_handler.go`, `.../handler/web/verificationbadge_handler.go` |
| MQ handler | `internal/interfaces/mq/handler/verificationbadge_handler.go` |

Nama tabel DB **tetap** `verifications`, `verification_events`, `verification_requirements` (sesuai DB Schema doc) — collision hanya terjadi di level nama paket Go, bukan di level SQL.

---

## Rekap Model Data (dari DB Schema doc)

```
verifications              — 1 row per pengajuan (pending → approved|rejected|revoked)
  entity_type, entity_id    — polymorphic, no FK (user|merchant)
  submitted_by, documents (jsonb), note
  reviewed_by/at, rejection_reason
  revoked_by/at, revoke_reason
  UNIQUE(entity_type, entity_id) WHERE status='pending'   ← max 1 pending
  UNIQUE(entity_type, entity_id) WHERE status='approved'  ← max 1 approved aktif

verification_events        — audit append-only (submitted|approved|rejected|revoked)
verification_requirements  — rules-as-data config per entity_type (any_of|all_of)

users.is_verified           BOOLEAN DEFAULT FALSE   -- ALTER, cache
merchants.is_verified        BOOLEAN DEFAULT FALSE   -- ALTER, cache
```

Resolve-live: sumber kebenaran = row `verifications.status='approved'` aktif. `is_verified` cuma cache, di-maintain app-layer dalam transaksi yang sama saat approve/revoke.

---

## Phase 1 — Domain (`internal/domain/verificationbadge/`)

### 1.1 `constant/verificationbadge_constant.go`
- Entity type: `EntityTypeUser = "user"`, `EntityTypeMerchant = "merchant"`
- Status: `StatusPending`, `StatusApproved`, `StatusRejected`, `StatusRevoked`
- Action (buat event log): `ActionSubmitted`, `ActionApproved`, `ActionRejected`, `ActionRevoked`
- Match mode: `MatchAnyOf = "any_of"`, `MatchAllOf = "all_of"`
- Error codes (`domainerr.Code`): `CodeVerificationNotFound`, `CodeVerificationInvalidTransition`, `CodeVerificationAlreadyPending`, `CodeVerificationAlreadyApproved`, `CodeVerificationReasonRequired`, `CodeVerificationEntityTypeInvalid`, `CodeVerificationDocumentsInsufficient`, `CodeVerificationRequirementNotFound`, `CodeVerificationPersistenceFailed`, `CodeVerificationQueryFailed`

### 1.2 `entity/verification.go` — aggregate root
```go
type Verification struct {
    ID, EntityType, EntityID, Status string
    SubmittedBy string
    Documents   []VerificationDocument // {DocType, URL, UploadedAt}
    Note        *string
    ReviewedBy, RejectionReason *string
    ReviewedAt  *time.Time
    RevokedBy, RevokeReason *string
    RevokedAt   *time.Time
    CreatedAt, UpdatedAt time.Time
}
```
Domain methods (transisi status = satu-satunya jalur mutasi, bukan usecase yang set field langsung):
- `NewVerification(id, entityType, entityID, submittedBy string, docs []VerificationDocument, note *string) (*Verification, error)` — validasi `entity_type` valid, `documents` tidak kosong.
- `(v *Verification) Approve(reviewerID string) error` — guard: `Status == StatusPending` (else `CodeVerificationInvalidTransition`), set `Status=approved`, `ReviewedBy`, `ReviewedAt`.
- `(v *Verification) Reject(reviewerID, reason string) error` — guard pending + `reason` non-empty (`CodeVerificationReasonRequired`), set `Status=rejected`, `RejectionReason`.
- `(v *Verification) Revoke(revokerID, reason string) error` — guard: `Status == StatusApproved` (else `CodeVerificationInvalidTransition`) + `reason` non-empty, set `Status=revoked`, `RevokedBy/At`, `RevokeReason`.

### 1.3 `entity/verification_requirement.go`
```go
type VerificationRequirement struct {
    ID, EntityType, MatchMode string
    AcceptedDocs []AcceptedDoc // {Key, Label, Required, Sensitive}
    MinDocuments int
    IsActive bool
    UpdatedBy *string
}
```
- `(r *VerificationRequirement) Validate(docs []VerificationDocument) error` — domain method inti rules-as-data:
  - `any_of` → `len(docs) >= r.MinDocuments` → else `CodeVerificationDocumentsInsufficient`.
  - `all_of` → semua `AcceptedDocs` dengan `Required=true` harus ada key-nya di `docs` → else error yang sama.
- `Update(matchMode string, minDocs int, docs []AcceptedDoc, updatedBy string) error` — validasi: `match_mode` ∈ {any_of, all_of}; kalau `all_of`, minimal 1 doc `Required=true` (`CodeVerificationRequirementNotFound`/invalid — tambah code khusus kalau perlu `CodeVerificationRequirementInvalid`).

> **Kenapa validasi dokumen jadi domain method, bukan usecase if/else**: rule "any_of vs all_of" adalah core business rule rules-as-data yang disebut eksplisit di RULES §3 & §9 Catatan Implementasi DB doc — cocok di entity, bukan diulang di tiap usecase yang butuh validasi submit.

### 1.4 `repository/interfaces.go`
```go
type VerificationRepository interface {
    Save(ctx, *Verification) error
    FindByID(ctx, id string) (*Verification, error)
    FindActiveApprovedByEntity(ctx, entityType, entityID string) (*Verification, error) // utk cek 409 already-approved
    FindPendingByEntity(ctx, entityType, entityID string) (*Verification, error)          // utk cek 409 already-pending
    FindLatestByEntity(ctx, entityType, entityID string) (*Verification, error)           // utk GET /requests/mine
    Update(ctx, *Verification) error // approve/reject/revoke
}

type VerificationEventRepository interface {
    Append(ctx, *VerificationEvent) error
    ListByVerificationID(ctx, verificationID string) ([]*VerificationEvent, error)
}

type VerificationRequirementRepository interface {
    FindByEntityType(ctx, entityType string) (*VerificationRequirement, error)
    ListAll(ctx) ([]*VerificationRequirement, error)
    Update(ctx, *VerificationRequirement) error
}
```
`VerificationEvent` entity sederhana (bukan aggregate root, append-only log): `{ID, VerificationID, Action, ActorID, Reason *string, Metadata JSONB, CreatedAt}`.

### 1.5 `event/events.go` + `event/routing.go` (domain event → outbox, pola sama seperti `domain/news/event` & `domain/directory/event`)
```go
// events.go
type VerificationSubmitted struct { VerificationID, EntityType, EntityID string; OccurredAt time.Time }
type VerificationApproved  struct { VerificationID, EntityType, EntityID, OwnerUserID string; OccurredAt time.Time }
type VerificationRejected  struct { VerificationID, EntityType, EntityID, OwnerUserID, Reason string; OccurredAt time.Time }
type VerificationRevoked   struct { VerificationID, EntityType, EntityID, OwnerUserID, Reason string; OccurredAt time.Time }

// routing.go
const (
    RoutingVerificationSubmitted = "verificationbadge.submitted"
    RoutingVerificationApproved  = "verificationbadge.approved"
    RoutingVerificationRejected  = "verificationbadge.rejected"
    RoutingVerificationRevoked   = "verificationbadge.revoked"

    QueueNotifVerificationSubmitted = "notif.verificationbadge.submitted"
    QueueNotifVerificationApproved  = "notif.verificationbadge.approved"
    QueueNotifVerificationRejected  = "notif.verificationbadge.rejected"
    QueueNotifVerificationRevoked   = "notif.verificationbadge.revoked"
)
```
> `OwnerUserID` perlu di-resolve usecase saat approve/reject/revoke: kalau `entity_type=user` → `entity_id` itu sendiri; kalau `merchant` → owner user id dari `merchants.owner_id` (butuh baca `directory` repo dari usecase, **bukan** dari domain package lain — jaga dependency rule).

### Tidak perlu domain service terpisah
Guard prasyarat ("user harus `active`", "merchant harus `published`") **bukan** domain service `verificationbadge` — itu butuh baca aggregate dari bounded context lain (`user`, `directory`). Sesuai pola existing (mis. `directory.ApproveMerchantUseCase` baca `merchantRepo` langsung), taruh guard ini di **usecase** `submit_verification.go` yang inject `userRepo` & `merchantRepo` dari domain masing-masing.

---

## Phase 2 — Usecase (`internal/app/usecase/verificationbadge/`)

### Mobile
| File | Endpoint | Ringkasan |
|------|----------|-----------|
| `get_requirements.go` | `GET /mobile/verification/requirements` | Baca `VerificationRequirementRepository.FindByEntityType`, map ke DTO **tanpa** field `sensitive` (internal only, per Mobile API doc §1). |
| `submit_verification.go` | `POST /mobile/verification/requests` | Guard kepemilikan entitas (user=diri sendiri; merchant=owner) → guard prasyarat status (`userRepo.FindByID().Status==active` / `merchantRepo.FindByID().Status==published`, else 403) → guard no-pending & no-approved (`FindPendingByEntity`/`FindActiveApprovedByEntity`, else 409) → `requirementRepo.FindByEntityType` → `requirement.Validate(docs)` (422 kalau gagal) → `entity.NewVerification(...)` → simpan + `VerificationEvent{Action: submitted}` + outbox `VerificationSubmitted` dalam satu transaksi. |
| `get_my_verification.go` | `GET /mobile/verification/requests/mine` | `FindLatestByEntity` → map ke DTO ringkas (`status`, `is_verified`, `reason` kalau rejected/revoked, `can_resubmit`). Kalau belum pernah ngajuin → `status:null` (bukan 404). |

### Web / Backoffice
| File | Endpoint | Ringkasan |
|------|----------|-----------|
| `list_verification_requests.go` | `GET /web/verification/requests` | Pakai **port query model** (lihat Phase 2.1) — filter `status` (default `pending`), `entity_type`, `q`, pagination. |
| `get_verification_detail.go` | `GET /web/verification/requests/:id` | `FindByID` + resolve `entity` (join user/merchant) + signed URL dokumen. |
| `approve_verification.go` | `POST .../approve` | `FindByID` → `v.Approve(reviewerID)` (409 kalau bukan pending) → resolve `OwnerUserID` → transaksi: `Update` verifikasi + set cache `is_verified=true` di entitas (via `userRepo`/`merchantRepo`) + `Append` event `approved` + outbox `VerificationApproved`. |
| `reject_verification.go` | `POST .../reject` | Validasi `reason` non-empty (400 `REASON_REQUIRED`) → `v.Reject(reviewerID, reason)` → transaksi: `Update` + event + outbox `VerificationRejected`. |
| `revoke_verification.go` | `POST .../revoke` | Validasi `reason` → `v.Revoke(revokerID, reason)` (409 kalau bukan approved) → transaksi: `Update` + set cache `is_verified=false` + event + outbox `VerificationRevoked`. |
| `get_verification_events.go` | `GET .../events` | `ListByVerificationID` + resolve nama actor. |
| `get_verification_requirements_config.go` | `GET /web/verification/requirements` | `ListAll()`. |
| `update_verification_requirements_config.go` | `PUT /web/verification/requirements/:entity_type` | `FindByEntityType` → `requirement.Update(...)` → `Update` repo. |

### Shared
- `dependencies.go` — `Dependencies{VerificationRepo, EventRepo, RequirementRepo, UserRepo, MerchantRepo, OutboxRepo, Transactor, FileUploader/MediaUploadSvc}` + `NewUseCases()` merangkai semua di atas (pola identik `content_report/dependencies.go`).
- `helpers.go` — `mapVerificationDomainError()` (domain error → `apperror`, ikuti CLAUDE.md §7), `resolveDocumentURLs()` (signed/public URL per dokumen via `mediaUploadSvc.PublicURL` atau signed-URL helper — lihat catatan Phase 4 soal private bucket), `resolveOwnerUserID(entityType, entityID)`.

### 2.1 Port Query Model — `internal/app/port/verificationbadge_query_model.go`
Listing backoffice butuh join ke `users`/`merchants` (nama, avatar) + filter/pagination kompleks → **wajib** query model per CLAUDE.md §6 (bukan repository sederhana):
```go
type VerificationListReadQuery struct {
    Status, EntityType, Q *string
    PaginationParams
}
type VerificationListReadItem struct {
    ID, EntityType, EntityID, Status string
    EntityName, EntityAvatarURL      *string
    SubmittedByID, SubmittedByName   string
    CreatedAt, UpdatedAt             time.Time
}
type VerificationListReadResult struct {
    Items []VerificationListReadItem
    Meta  Meta
}
type VerificationQueryReadModel interface {
    List(ctx, VerificationListReadQuery) (*VerificationListReadResult, error)
}
```
`Execute()` usecase `list_verification_requests.go` wajib return `([]dto.Item, port.Meta, error)` tiga nilai terpisah (CLAUDE.md §8), handler pakai `respond.SuccessWithMeta`.

### 2.2 DTO — `internal/app/dto/verificationbadge_dto.go`
Semua request/response struct kedua sisi (mobile + web): `SubmitVerificationInput`, `VerificationMineResponse`, `VerificationRequirementResponse`, `ListVerificationRequestsQuery`, `VerificationAdminItem`, `VerificationAdminDetail`, `RejectVerificationInput{Reason}`, `RevokeVerificationInput{Reason}`, `UpdateRequirementInput`, `VerificationEventItem`.

---

## Phase 3 — Interfaces

### 3.1 HTTP Handler
- `internal/interfaces/http/handler/mobile/verificationbadge_handler.go` (+ `_test.go`): `GetRequirements`, `Submit`, `GetMine`.
- `internal/interfaces/http/handler/web/verificationbadge_handler.go` (+ `_test.go`): `List`, `GetDetail`, `Approve`, `Reject`, `Revoke`, `GetEvents`, `GetRequirementsConfig`, `UpdateRequirementsConfig`.
- Skenario test wajib per CLAUDE.md §11: unauthenticated 401, validation 400/422 (reason kosong, dokumen kurang), not found 404, forbidden 403 (bukan Superadmin di web; bukan owner entitas di mobile), success 2xx. Tambah kasus khusus: 409 already-pending, 409 invalid-transition.

### 3.2 Router
`internal/interfaces/http/router/router.go`:
```
mobile group /api/v1/mobile/verification  (auth required)
  GET  /requirements
  POST /requests
  GET  /requests/mine

web group /api/v1/web/verification  (auth + permission middleware)
  GET  /requests                              → perm: verification.view_queue
  GET  /requests/:id                           → perm: verification.view_queue
  POST /requests/:id/approve                   → perm: verification.review
  POST /requests/:id/reject                     → perm: verification.review
  POST /requests/:id/revoke                     → perm: verification.revoke
  GET  /requests/:id/events                    → perm: verification.view_queue
  GET  /requirements                            → perm: verification.view_queue (atau superadmin-only umum)
  PUT  /requirements/:entity_type              → perm: verification.review (config change tetap kurasi Superadmin)
```

### 3.3 Role-Permission (cross-module follow-up RULES §5)
`internal/seeders/permission_seeder.go` (+ `role_permission_seeder.go`): daftarkan 4 key baru — `verification.request` (assign ke role member), `verification.view_queue`, `verification.review`, `verification.revoke` (assign **hanya** ke role Superadmin — **jangan** assign ke Admin Regional, sesuai matrix RULES §5).

### 3.4 MQ — Notification Dispatch
- `internal/interfaces/mq/handler/verificationbadge_handler.go` — 4 handler func (pola identik `news_handler.go` / `subscription_handler.go`):
  - `HandleVerificationSubmitted` → target **Superadmin** (in-app, bypass preference). ⚠️ **Gap baru**: dispatcher saat ini cuma punya `TargetModeUnicast/Multicast/Broadcast` dengan ID eksplisit — **belum ada** cara resolve "semua user dengan role=Superadmin". Perlu tambahan kecil: method `userRepo.FindIDsByRole(ctx, roleName string) ([]string, error)` (atau reuse role-permission repo), lalu dispatch `TargetModeMulticast` dengan `RecipientIDs` hasil resolve itu. Ini satu-satunya potongan infra notifikasi yang genuinely baru — modul lain (news, subscription, content_report) semuanya unicast ke 1 user spesifik.
  - `HandleVerificationApproved/Rejected/Revoked` → `TargetModeUnicast` ke `OwnerUserID`, channel Push+InApp, bypass preference (transaksional, sama pola `SubscriptionUpgradeApproved`).
- `internal/interfaces/mq/router/router.go` — bind 4 routing key → queue → handler func.
- `cmd/worker/main.go` — register binding (ikuti pola `RoutingNewsArticleApproved` dkk).

---

## Phase 4 — Infrastructure

### 4.1 Migration
`make migrate-create NAME=create_verification_badge_tables` → hasil `internal/migrations/<timestamp>_create_verification_badge_tables.up.sql` / `.down.sql`.

**up.sql** — copy persis DDL dari `VERIFICATION_DB_SCHEMA.md` §2–5 (tabel `verifications`, `verification_events`, `verification_requirements`, semua index termasuk 2 unique partial index, ALTER `users`/`merchants` + index `is_verified`) **plus** seed default `verification_requirements` (2 INSERT dari DB doc §7) — pola sama seperti `event_settings` migration yang menyisipkan default row langsung di file up.sql (bukan seeder terpisah).

**down.sql** — drop tabel (cascade event), drop kolom `is_verified` dari `users`/`merchants`.

### 4.2 Persistence
- `postgres_verificationbadge_repository.go` — `Save/FindByID/FindActiveApprovedByEntity/FindPendingByEntity/FindLatestByEntity/Update` untuk `verifications`, plus repo terpisah untuk `verification_events` (`Append`/`ListByVerificationID`) dan `verification_requirements` (`FindByEntityType`/`ListAll`/`Update`). Ikuti CLAUDE.md §7: map `sql.ErrNoRows`→`CodeVerificationNotFound`, insert/update gagal→`CodeVerificationPersistenceFailed`, unique violation (`23505`) pada partial index pending/approved → **race-condition fallback** untuk error `CodeVerificationAlreadyPending`/`CodeVerificationAlreadyApproved` (usecase sudah cek duluan, tapi index ini jaring pengaman concurrent submit).
- `postgres_verificationbadge_query.go` — implementasi `VerificationQueryReadModel.List`: `LEFT JOIN` ke `users` atau `merchants` berdasarkan `entity_type` (butuh `UNION` atau `CASE`-based join / dua query digabung — desain query eksplisit karena polymorphic tanpa FK), `LEFT JOIN users submitted_by`, filter `q` ke nama entitas & pemohon (`ILIKE`), `resolvePaginationParams()` helper standar.

### 4.3 Media
- Tambah `ObjectPathVerificationDocument ObjectPath = "/verification/documents/"` di `internal/app/service/media/object_path.go`.
- Tambah context `"verification"` di validasi context media service (RULES §3, cross-module follow-up "Media service: tambah `context: verification`").
- **Private storage**: dokumen verifikasi (`sensitive=true` per config) **wajib** disimpan di bucket/prefix private, bukan CDN publik (RULES §3 & DO/DON'T §9). Ini butuh konfirmasi kebijakan bucket ke tim storage — RULES doc sendiri menandai ini sebagai follow-up terbuka ("konfirmasi bucket/policy sama tim storage"). Kalau `objectstorage` service saat ini belum punya konsep private-bucket + signed URL short-lived, itu prasyarat infra terpisah sebelum endpoint detail/list backoffice bisa expose `documents[].url` dengan aman.

### 4.4 Wiring
- `cmd/app/main.go` — instantiate 3 repo (`postgres_verificationbadge_*`), `verificationbadge.NewUseCases(...)`, handler mobile & web, daftarkan di router group (pola identik baris `contentReportRepo`/`webContentReportHandler` di sekitar cmd/app/main.go:616-794).
- `cmd/worker/main.go` — register 4 MQ handler binding baru.

---

## Urutan Kerja yang Disarankan (breakdown PR)

1. **PR 1 — Domain + Infra dasar**: migration, entity, constant, repository interfaces, persistence (repo + query), wiring kosong (belum ada route). Bisa di-test lewat unit test domain + integration test persistence langsung.
2. **PR 2 — Mobile flow**: usecase mobile (get_requirements, submit, get_mine) + handler mobile + route + handler test. Butuh media context "verification" & ObjectPath sudah ada.
3. **PR 3 — Backoffice flow + MQ/Notifikasi**: usecase web (list/detail/approve/reject/revoke/events/requirements config) + handler web + route + permission seeding + MQ handler & router binding + resolve target Superadmin (termasuk `FindIDsByRole` baru kalau belum ada).

Alasan split: mobile flow (submit) tidak butuh cache `is_verified` maupun notifikasi — bisa selesai & di-test independen dulu sebelum backoffice review flow yang lebih kompleks (transaksi cache + event + outbox + MQ).

---

## Verification (manual test plan setelah implementasi)

1. User submit verifikasi tanpa dokumen → 422 `documents tidak memenuhi match_mode`.
2. User submit lagi saat masih `pending` → 409 `ALREADY_PENDING`.
3. Superadmin approve → `GET /mobile/verification/requests/mine` langsung `is_verified:true`; cek kolom `users.is_verified=true`; cek badge muncul di profil (modul User/News byline lain, bukan endpoint ini).
4. Superadmin reject dengan reason kosong → 400 `REASON_REQUIRED`.
5. Superadmin reject dengan reason → user dapat notif push+in-app berisi reason; user resubmit → record baru (`verifications` count bertambah, bukan update row lama).
6. Superadmin revoke dari status approved → `is_verified` balik `false`, badge hilang, event `revoked` tercatat, row lama (approved) tidak berubah statusnya sendiri (row baru status revoked — **cek ulang di implementasi**: apakah revoke update row `approved` yang sama jadi `revoked`, atau insert row baru? DB Schema §4 "Aturan state" cuma bahas resubmit; revoke context RULES §4 bilang "dari status approved" — asumsikan **update row yang sama** jadi `revoked`, bukan insert baru, karena unique index `uq_verif_one_approved` akan bentrok kalau insert row baru berstatus approved dan approved lama belum diubah. **Konfirmasi ini di code review PR 3**.)
7. Merchant belum `published` coba ajukan verifikasi → 403.
8. Admin Regional coba akses endpoint web verification → 403 (permission `verification.*` tidak didaftarkan ke role-nya).
9. Response mobile `GET /requests/mine` **tidak** boleh mengandung `documents`, `reviewed_by`, atau detail internal lain.

---

## Catatan Terbuka (perlu keputusan/konfirmasi tambahan sebelum atau selama implementasi)

1. **Signed URL dokumen sensitif** — infra private bucket + signed-URL short-lived belum tentu tersedia di `objectstorage` service saat ini; cek dulu sebelum PR 1 kalau mau expose `documents[].url` di detail backoffice.
2. **Target notifikasi "Superadmin" (multicast by role)** — belum ada pola existing di codebase (semua notifikasi sejenis saat ini unicast ke 1 user). Perlu `userRepo.FindIDsByRole()` atau reuse role-permission query — masuk scope PR 3.
3. **Revoke: update row yang sama vs insert row baru** — perlu dikonfirmasi ulang terhadap unique index `uq_verif_one_approved` sebelum implementasi domain method `Revoke()` (lihat poin 6 di Verification plan di atas).
4. **KTA (ID Card) auto-attach** — disebut RULES §3 sebagai follow-up Phase 2 (out of scope sekarang), tidak perlu dikerjakan di modul ini.

---

*Plan v1.0 — Verification Badge Module. Dibuat 2026-07-13, mengikuti pipeline RULES → DB Schema → API Spec → Plan.*
