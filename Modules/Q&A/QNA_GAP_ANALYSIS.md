# Gap Analysis — QnA Module: Spec vs k-forum-api

> **Tanggal:** 2026-06-25  
> **Dibandingkan:** Spec docs (revisi 2026-06-25 — forum komunitas) vs kode aktual di `k-forum-api`  
> **Kesimpulan umum:** Implementasi sesuai spec lama (private Q&A + admin answer only). Semua fitur **forum komunitas** yang ditambahkan di revisi 2026-06-25 belum diimplementasikan — meliputi: public question feed, member answering, upvote, accepted answer, assignment, moderation jawaban.

---

## 1. Database Migration

File aktual: `internal/migrations/0009_create_qna_tables.up.sql` — ini versi lama. Perlu migration baru.

### 1.1 Tabel `qna_questions` — Kolom Hilang

| Kolom yang Hilang | Tipe | Default | Kebutuhan |
|---|---|---|---|
| `visibility` | VARCHAR(10) | `'private'` | Pilihan penanya: private/public |
| `assigned_to` | UUID NULL FK users | — | Assignment ke expert |
| `assigned_by` | UUID NULL FK users | — | Siapa yang assign |
| `assigned_at` | TIMESTAMPTZ NULL | — | Waktu assignment |
| `answer_count` | INTEGER | `0` | Counter denormalized listing |
| `accepted_count` | INTEGER | `0` | Jumlah jawaban accepted |
| `view_count` | INTEGER | `0` | Counter view |
| `approved_by` | UUID NULL FK users | — | Admin yang approve pertanyaan publik |
| `approved_at` | TIMESTAMPTZ NULL | — | Waktu approve |

**Index hilang:**
```sql
CREATE INDEX idx_qna_questions_public_feed ON qna_questions (category_id, created_at DESC)
  WHERE visibility = 'public' AND status IN ('approved', 'converted', 'closed');

CREATE INDEX idx_qna_questions_assigned ON qna_questions (assigned_to, status, created_at DESC)
  WHERE assigned_to IS NOT NULL;
```

### 1.2 Tabel `qna_answers` — Desain Lama, Perlu Major Rework

Kolom yang ada di DB sekarang vs yang seharusnya:

| Field Lama (harus dihapus) | Alasan |
|---|---|
| `is_primary BOOLEAN` | Diganti dengan `is_accepted` — semantik berbeda total |
| `idx_qna_answers_primary_unique` (UNIQUE on question_id) | Harus dihapus — sekarang boleh banyak jawaban per pertanyaan |

| Kolom yang Hilang | Tipe | Default | Kebutuhan |
|---|---|---|---|
| `answerer_type` | VARCHAR(10) | `'member'` | member/expert/admin |
| `status` | VARCHAR(10) | `'visible'` | visible/pending/rejected/hidden |
| `is_accepted` | BOOLEAN | `false` | Tandai jawaban valid |
| `accepted_by` | UUID NULL FK users | — | Siapa yang accept |
| `accepted_at` | TIMESTAMPTZ NULL | — | Waktu accept |
| `upvote_count` | INTEGER | `0` | Counter upvote denormalized |
| `moderated_by` | UUID NULL FK users | — | Moderator |
| `moderated_at` | TIMESTAMPTZ NULL | — | Waktu moderasi |
| `reject_reason` | TEXT NULL | — | Alasan tolak/hide |

**Index yang perlu ditambah:**
```sql
-- Hapus index lama yang salah:
DROP INDEX idx_qna_answers_primary_unique;

-- Tambah:
CREATE INDEX idx_qna_answers_question ON qna_answers
  (question_id, is_accepted DESC, upvote_count DESC, created_at ASC);  -- ganti yang lama

CREATE INDEX idx_qna_answers_pending ON qna_answers (question_id, created_at ASC)
  WHERE status = 'pending';

CREATE UNIQUE INDEX idx_qna_answers_user_question_unique ON qna_answers (question_id, answered_by);
```

### 1.3 Tabel `qna_answer_votes` — BELUM ADA

Seluruh tabel ini belum dibuat:
```sql
CREATE TABLE qna_answer_votes (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    answer_id  UUID        NOT NULL REFERENCES qna_answers(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id)        ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_qna_answer_votes_unique ON qna_answer_votes (answer_id, user_id);
CREATE INDEX        idx_qna_answer_votes_answer ON qna_answer_votes (answer_id);
```

### 1.4 Tabel `qna_bot_config` — Kolom Hilang

| Kolom yang Hilang | Tipe | Default |
|---|---|---|
| `answer_moderation_mode` | VARCHAR(10) | `'manual'` |
| `question_moderation_mode` | VARCHAR(10) | `'manual'` |

---

## 2. Domain Layer

### 2.1 Constants (`qna_constant.go`) — Tipe dan Value Hilang

**`QnaQuestionStatus` — 2 value hilang:**
```go
// Belum ada:
QnaQuestionStatusApproved QnaQuestionStatus = "approved"
QnaQuestionStatusClosed   QnaQuestionStatus = "closed"
```

**Tipe baru yang belum ada sama sekali:**
```go
// Visibility
type QnaQuestionVisibility string
const (
    QnaQuestionVisibilityPrivate QnaQuestionVisibility = "private"
    QnaQuestionVisibilityPublic  QnaQuestionVisibility = "public"
)

// Answerer type
type QnaAnswererType string
const (
    QnaAnswererTypeMember QnaAnswererType = "member"
    QnaAnswererTypeExpert QnaAnswererType = "expert"
    QnaAnswererTypeAdmin  QnaAnswererType = "admin"
)

// Answer status
type QnaAnswerStatus string
const (
    QnaAnswerStatusVisible  QnaAnswerStatus = "visible"
    QnaAnswerStatusPending  QnaAnswerStatus = "pending"
    QnaAnswerStatusRejected QnaAnswerStatus = "rejected"
    QnaAnswerStatusHidden   QnaAnswerStatus = "hidden"
)
```

**Domain error codes yang hilang:**
```go
// QnaQuestion — terkait fitur baru
CodeQnaQuestionNotPending         = "DOMAIN_QNA_QUESTION_NOT_PENDING"
CodeQnaQuestionNotPublic          = "DOMAIN_QNA_QUESTION_NOT_PUBLIC"
CodeQnaQuestionAlreadyApproved    = "DOMAIN_QNA_QUESTION_ALREADY_APPROVED"
CodeQnaQuestionClosed             = "DOMAIN_QNA_QUESTION_CLOSED"
CodeQnaQuestionAssignedToRequired = "DOMAIN_QNA_QUESTION_ASSIGNED_TO_REQUIRED"

// QnaAnswer — moderation & accept
CodeQnaAnswerAlreadyAnswered      = "DOMAIN_QNA_ANSWER_ALREADY_ANSWERED"
CodeQnaAnswerCannotVoteOwn        = "DOMAIN_QNA_ANSWER_CANNOT_VOTE_OWN"
CodeQnaAnswerInvalidModAction     = "DOMAIN_QNA_ANSWER_INVALID_MOD_ACTION"
CodeQnaAnswerReasonRequired       = "DOMAIN_QNA_ANSWER_REASON_REQUIRED"

// QnaAnswerVote — belum ada sama sekali
CodeQnaAnswerVotePersistenceFailed = "DOMAIN_QNA_ANSWER_VOTE_PERSISTENCE_FAILED"
CodeQnaAnswerVoteQueryFailed       = "DOMAIN_QNA_ANSWER_VOTE_QUERY_FAILED"
```

### 2.2 Entity `QnaQuestion` — Field & Method Hilang

Field yang hilang di struct:
```go
Visibility    string     // "private" | "public"
AssignedTo    *string
AssignedBy    *string
AssignedAt    *time.Time
AnswerCount   int
AcceptedCount int
ViewCount     int
ApprovedBy    *string
ApprovedAt    *time.Time
```

Method domain yang hilang:
```go
func (q *QnaQuestion) Approve(approvedBy string) error     // pending → approved (hanya public)
func (q *QnaQuestion) Assign(to, by string) error          // set assigned_to
func (q *QnaQuestion) Unassign() error                     // clear assignment
func (q *QnaQuestion) Close() error                        // approved → closed
func (q *QnaQuestion) IncrementAnswerCount()               // counter sync
func (q *QnaQuestion) IncrementAcceptedCount()
func (q *QnaQuestion) DecrementAcceptedCount()
```

### 2.3 Entity `QnaAnswer` — Desain Lama

Field yang salah/hilang:
```go
// HAPUS:
IsPrimary bool  // semantik lama, bukan forum

// TAMBAH:
AnswererType  string     // member/expert/admin
Status        string     // visible/pending/rejected/hidden
IsAccepted    bool
AcceptedBy    *string
AcceptedAt    *time.Time
UpvoteCount   int
ModeratedBy   *string
ModeratedAt   *time.Time
RejectReason  *string
```

Method domain yang hilang:
```go
func (a *QnaAnswer) Accept(acceptedBy string) error
func (a *QnaAnswer) Unaccept() error
func (a *QnaAnswer) Approve(moderatedBy string) error  // pending → visible
func (a *QnaAnswer) Reject(reason, moderatedBy string) error
func (a *QnaAnswer) Hide(reason, moderatedBy string) error
```

### 2.4 Entity `QnaBotConfig` — Field Hilang

```go
// Tambah di struct dan Update() method:
AnswerModerationMode   string  // "auto" | "manual"
QuestionModerationMode string  // "auto" | "manual"
```

### 2.5 Entity `QnaAnswerVote` — BELUM ADA

File `entity/qna_answer_vote.go` belum dibuat sama sekali.

---

## 3. Repository Interfaces

### 3.1 `QnaAnswerRepository` — Method Hilang

```go
// Perlu ditambah:
FindByID(ctx context.Context, id string) (*entity.QnaAnswer, error)
FindByUserAndQuestion(ctx context.Context, userID, questionID string) (*entity.QnaAnswer, error)
```

### 3.2 `QnaAnswerVoteRepository` — BELUM ADA

Interface baru yang dibutuhkan:
```go
type QnaAnswerVoteRepository interface {
    Save(ctx context.Context, vote *entity.QnaAnswerVote) error
    Delete(ctx context.Context, answerID, userID string) error
    FindByAnswerAndUser(ctx context.Context, answerID, userID string) (*entity.QnaAnswerVote, error)
    RecalcCount(ctx context.Context, answerID string) (int, error)
}
```

---

## 4. Application Layer — Usecase

### 4.1 Web/Backoffice — Usecase Hilang

| Spec | Usecase File | Status |
|---|---|---|
| B10 Reorder FAQ | `reorder_faq_items.go` | ❌ Belum ada |
| B17 Approve Question | `approve_question.go` | ❌ Belum ada |
| B18 Assign Question | `assign_question.go` | ❌ Belum ada |
| B19 Close Question | `close_question.go` | ❌ Belum ada |
| B20 List Answers (admin) | `list_answers_admin.go` | ❌ Belum ada |
| B21 Moderate Answer | `moderate_answer.go` | ❌ Belum ada |
| B22 Mark Answer Accepted | `mark_answer_accepted.go` | ❌ Belum ada (shared dengan mobile) |

### 4.2 Mobile — Usecase Hilang

| Spec | Usecase File | Status |
|---|---|---|
| #4 Search FAQ | `search_faq.go` | ❌ Belum ada |
| #9 Public Question Feed | `list_public_questions.go` | ❌ Belum ada |
| #10 Public Question Detail | `get_public_question_detail.go` | ❌ Belum ada |
| #11 Post Answer | `post_answer.go` | ❌ Belum ada |
| #12 Edit My Answer | `edit_answer.go` | ❌ Belum ada |
| #13 Upvote Answer | `upvote_answer.go` | ❌ Belum ada |
| #14 Accept Answer (mobile) | `mark_answer_accepted.go` | ❌ Belum ada (shared) |
| #15 Assigned Questions | `list_assigned_questions.go` | ❌ Belum ada |

---

## 5. Port Query Model (`qna_query_model.go`) — Method Hilang

```go
// Perlu ditambah ke interface QnaQueryReadModel:

SearchFaq(ctx context.Context, query SearchFaqQuery) ([]SearchFaqItem, port.Meta, error)
ListPublicQuestions(ctx context.Context, query ListPublicQuestionsQuery) ([]PublicQuestionListItem, port.Meta, error)
GetPublicQuestionDetail(ctx context.Context, questionID, userID string) (*PublicQuestionDetail, error)
ListAnswersForAdmin(ctx context.Context, questionID string, statusFilter *string) ([]AdminAnswerItem, error)
ListQuestionsAssignedToUser(ctx context.Context, userID string, query AssignedQuestionsQuery) ([]AssignedQuestionItem, port.Meta, error)
```

---

## 6. DTO (`qna_dto.go`) — Struct Hilang

**Response DTOs:**
- `SearchFaqItem` (id, question, answer_excerpt, category, helpful_count, relevance_score)
- `PublicQuestionListItem` (id, category, question_text, asker, answer_count, accepted_count, view_count, status, created_at)
- `PublicQuestionDetail` (lengkap dengan `answers`, `can_answer`, `can_accept`, `my_answer_id`)
- `PublicAnswerItem` (dengan user_upvoted, answerer, is_accepted, upvote_count)
- `AdminAnswerItem` (dengan status, moderation info)
- `AssignedQuestionListItem`

**Input DTOs:**
- `SearchFaqQuery` (q, category_id, limit, offset)
- `ListPublicQuestionsQuery` (category_id, filter, limit, offset)
- `PostAnswerInput` (answer_text, attachment_urls)
- `EditAnswerInput` (answer_text, attachment_urls)
- `AcceptAnswerInput` (is_accepted bool)
- `ApproveQuestionInput` (no body)
- `AssignQuestionInput` (assigned_to uuid|null)
- `CloseQuestionInput` (no body)
- `ModerateAnswerInput` (action string, reason *string)
- `ReorderFaqInput` (category_id, ordered_ids []string)
- `ListAnswersAdminQuery` (status filter)

---

## 7. HTTP Handlers

### 7.1 Web Handler — Handler Hilang

| Spec | Handler Method | Route | Status |
|---|---|---|---|
| B10 | `ReorderFaqItems` | `PUT /faqs/reorder` | ❌ |
| B17 | `ApproveQuestion` | `POST /questions/:id/approve` | ❌ |
| B18 | `AssignQuestion` | `PUT /questions/:id/assign` | ❌ |
| B19 | `CloseQuestion` | `POST /questions/:id/close` | ❌ |
| B20 | `ListAnswers` | `GET /questions/:id/answers` | ❌ |
| B21 | `ModerateAnswer` | `PATCH /answers/:id/moderate` | ❌ |
| B22 | `MarkAnswerAccepted` | `POST /answers/:id/accept` | ❌ |

### 7.2 Mobile Handler — Handler Hilang

| Spec | Handler Method | Route | Status |
|---|---|---|---|
| #4 | `SearchFaq` | `GET /search` | ❌ |
| #9 | `ListPublicQuestions` | `GET /questions/public` | ❌ |
| #10 | `GetPublicQuestionDetail` | `GET /questions/public/:id` | ❌ |
| #11 | `PostAnswer` | `POST /questions/:id/answers` | ❌ |
| #12 | `EditAnswer` | `PUT /answers/:id` | ❌ |
| #13 | `UpvoteAnswer` | `POST /answers/:id/upvote` | ❌ |
| #14 | `AcceptAnswer` | `POST /answers/:id/accept` | ❌ |
| #15 | `ListAssignedQuestions` | `GET /questions/assigned` | ❌ |

---

## 8. HTTP Method & Path Mismatch (Spec vs Router)

| Spec | Router Aktual | Catatan |
|---|---|---|
| `PUT /categories/{id}` (B3) | `PATCH /categories/:category_id` | PATCH lebih tepat untuk partial update; bisa jadi intentional |
| `PUT /faqs/{id}` (B8) | `PATCH /faqs/:faq_id` | Sama |
| `PUT /questions/{id}/reject` (B14) | `POST /questions/:id/reject` | Perlu disamakan |
| `GET /questions/my` (Mobile #7) | `GET /questions/me` | Path berbeda dari spec |
| `GET /questions/my/:id` (Mobile #8) | `GET /questions/me/:question_id` | Path berbeda dari spec |

> Catatan: PATCH vs PUT untuk update — PATCH secara semantik lebih benar untuk partial update (yang dipakai di spec). Ini mungkin intentional. Path `/me` vs `/my` perlu disesuaikan dengan kesepakatan tim.

---

## 9. Events & MQ Handler

### 9.1 Events Hilang (`event/events.go`)

| Event | Routing Key | Trigger |
|---|---|---|
| `QnAQuestionApproved` | `qna.question.approved` | Saat pertanyaan publik di-approve → notif ke penanya |
| `QnAQuestionAssigned` | `qna.question.assigned` | Saat pertanyaan di-assign → notif ke expert |
| `QnAAnswerAccepted` | `qna.answer.accepted` | Saat jawaban ditandai valid → notif ke pemilik jawaban |
| `QnAAnswerRejected` | `qna.answer.rejected` | Saat jawaban ditolak moderator → notif ke pemilik jawaban |

### 9.2 MQ Handlers Hilang (`interfaces/mq/handler/qna_handler.go`)

- `HandleQnAQuestionApproved()`
- `HandleQnAQuestionAssigned()`
- `HandleQnAAnswerAccepted()`
- `HandleQnAAnswerRejected()`

---

## 10. Ringkasan Priority

### 🔴 Blocker — DB Migration dulu

Tanpa migration baru, semua fitur forum komunitas tidak bisa diimplementasikan. Urutan:
1. Buat migration `0010_update_qna_for_community.up.sql` — tambah kolom ke `qna_questions`, rework `qna_answers`, buat `qna_answer_votes`, tambah kolom ke `qna_bot_config`
2. Update domain constants, entities, repository interfaces

### 🟡 Core Forum Features — Implementasikan berurutan

3. Public question approval (B17) + visibility support
4. Assignment (B18)
5. Public question feed (Mobile #9, #10)
6. Member posting answer (Mobile #11) + moderasi jawaban (B21)
7. Upvote answer (Mobile #13)
8. Accepted answer (B22, Mobile #14)
9. Close question (B19)

### 🟢 Secondary Features

10. Search FAQ (Mobile #4)
11. Reorder FAQ (B10)
12. Edit answer (Mobile #12)
13. Assigned questions queue (Mobile #15)
14. Events & notifikasi baru

---

## 11. Yang Sudah Benar ✅

Implementasi berikut sudah sesuai spec dan tidak perlu diubah:

- Seluruh kategori CRUD (B1–B4) ✅
- Seluruh FAQ CRUD (B5–B9, kecuali B10 reorder) ✅
- FAQ status workflow (draft → published → archived) ✅
- Vote FAQ helpful (B5 vote, Mobile #5) ✅
- Submit pertanyaan member (Mobile #6) ✅
- My Questions list & detail (Mobile #7, #8) ✅
- Admin answer question + convert to FAQ (B13) ✅
- Admin reject question (B14) ✅
- Bot config CRUD (B15, B16) ✅
- Media attachment flow (presign/confirm/delete) ✅
- Event + notifikasi untuk answered & rejected ✅
- Permission gating (`manage_qna`, `ask_qna`) ✅
