# Issue: Q&A — attachment tak terlihat, assign expert 404, error jawab pertanyaan publik

- **Modul**: Q&A — pertanyaan privat (attachment), assignment ke expert, forum pertanyaan publik
- **Severity**: 🔴 Tinggi untuk Issue 2 & 3 (fitur inti gagal total, bukan cuma UX) — 🟠 Sedang untuk Issue 1
- **Status**: 🔴 Open — 3 issue di dokumen ini, semua root cause sudah ditemukan by code review
- **Ditemukan**: 22 Jul 2026, saat review user journey `05_QNA_JOURNEY.md`
- **Pelapor**: review manual (dev), dikonfirmasi via code review langsung ke `k-forum-api` dan `k-forum-backoffice`

---

## Issue 1 — Attachment pertanyaan privat tidak tampil di backoffice admin

- **Repo**: `k-forum-backoffice` (root cause utama) + `k-forum-api` (temuan tambahan: attachment sebenarnya tidak private secara teknis)

### Ringkasan

Saat admin/expert membuka detail pertanyaan privat yang ada lampirannya, modal detail hanya menampilkan teks pertanyaan dan teks jawaban — lampiran (gambar/dokumen) tidak ditampilkan sama sekali, walau datanya sudah dikirim backend dengan benar.

### Root cause (code review)

**Backend sudah benar** — `getQuestion()` (`internal/infrastructure/persistence/postgres_qna_query.go:864,875,887`) meng-query `qu.attachment_urls` dan mengisi `item.AttachmentURLs`. Mapper `mapPortQnaQuestionToDTO` (`internal/app/usecase/qna/qna_mapper.go:66-67`) juga benar memetakan `AttachmentURLs`/`AttachmentURLsRaw` untuk pertanyaan **dan** jawaban ke response JSON (`item.Answer.AttachmentURLs` di baris 88-89). Jadi field `attachment_urls` **ada** di response API `GET /api/v1/web/qna/questions/{question_id}`.

**Frontend yang tidak render** — di `app/pages/qa/index.vue`, modal detail pertanyaan (mulai baris ~1828) menampilkan `question_text` (baris 1870), `bot_response` (1875-1882), `answer` teks (1887-1896), `rejection_reason` (1902-1909) — **tidak ada satupun elemen yang membaca `attachment_urls`**. Field ini cuma dideklarasikan di type TypeScript store (`app/stores/qnaStore.ts:215`, khusus untuk payload jawab pertanyaan, bukan untuk display), tidak pernah dipakai di template manapun.

**Temuan tambahan (terkait tapi beda kategori)**: attachment pertanyaan/jawaban QnA di-upload dengan `port.PrivacyPublic` (`internal/app/usecase/qna/get_qna_attachment_presign_url.go:77`), **bukan** `PrivacyPrivate`. Artinya secara teknis, klaim "attachment hanya bisa diakses penanya dan expert/admin yang ditugaskan" **tidak benar-benar ditegakkan** di level storage — siapapun yang punya link langsung (bocor lewat screenshot, share, dsb.) bisa akses file-nya, karena disimpan di bucket public, bukan private dengan signed-URL time-limited seperti pola yang sudah dipakai modul lain (lihat `CLAUDE.md` §13 — verification badge document sudah pakai `PrivacyPrivate`). Ini bukan penyebab langsung dari "admin tidak bisa lihat attachment", tapi ditemukan saat menyusuri kode yang sama dan berkaitan dengan pernyataan akses-restriction yang sama.

### Yang diminta

**k-forum-backoffice**:
1. Tambah UI di modal detail pertanyaan (`app/pages/qa/index.vue`) untuk menampilkan `attachment_urls` milik pertanyaan, dan `answer.attachment_urls` milik jawaban — cukup list thumbnail/link yang bisa diklik.

**k-forum-api** (non-blocking untuk Issue 1, tapi disarankan dikerjakan bersamaan karena topiknya sama):
1. Pertimbangkan pindah attachment QnA (minimal untuk pertanyaan **privat**) dari `PrivacyPublic` ke `PrivacyPrivate`, mengikuti checklist migrasi di `CLAUDE.md` §13, supaya klaim "hanya penanya dan expert/admin yang ditugaskan" benar-benar ditegakkan, bukan cuma UI yang menyembunyikan.

### Kriteria selesai (acceptance)

- [ ] Admin buka pertanyaan privat berlampiran → lampiran tampil dan bisa dibuka di modal detail backoffice.
- [ ] (Kalau poin privacy dikerjakan) Attachment pertanyaan privat tidak bisa diakses lewat URL langsung oleh user yang bukan penanya/expert-ditugaskan/admin.

---

## Issue 2 — Admin gagal assign pertanyaan privat ke expert (404 route not found)

- **Repo**: `k-forum-api`

### Ringkasan

Tombol assign expert di backoffice selalu gagal dengan error seolah endpoint tidak ada.

### Root cause (code review)

Spec resmi (`K-FORUM-DOCS/Modules/Q&A/API_SPEC_QNA_BACKOFFICE.md`, B18):
```
PUT /api/v1/web/qna/questions/{question_id}/assign
```

Backoffice **sudah benar** mengikuti spec — `app/stores/qnaStore.ts:263`:
```ts
const res = await api.put(`${BASE}/questions/${id}/assign`, { assigned_to: assignedTo })
```

Tapi backend meregistrasi route ini dengan verb yang **berbeda** dari spec-nya sendiri — `internal/interfaces/http/router/router.go:461`:
```go
qna.POST("/questions/:question_id/assign", webQnaHandler.AssignQuestion)
```

Gin membedakan route berdasarkan method+path — `PUT` ke path yang cuma register `POST` akan dianggap "no matching route" → 404, persis seperti yang dilaporkan ("route not found"). Backend yang salah, bukan backoffice.

### Yang diminta ke backend (k-forum-api)

1. Ubah `router.go:461` dari `qna.POST(...)` jadi `qna.PUT(...)`, supaya sesuai `API_SPEC_QNA_BACKOFFICE.md` B18 dan sesuai yang sudah dipanggil backoffice.
2. Cek endpoint lain di QnA (dan modul lain) untuk verb yang tidak sesuai spec — pola ini (spec bilang satu verb, implementasi pakai verb lain) berpotensi terulang di endpoint lain yang belum ketauan karena belum dicoba.

### Kriteria selesai (acceptance)

- [ ] Admin assign pertanyaan (publik maupun privat) ke expert → berhasil, `assigned_to` tersimpan dan tampil di detail pertanyaan.
- [ ] Kirim `assigned_to: null` → penugasan tercabut (sesuai spec B18).

---

## Issue 3 — Lihat pertanyaan publik di mobile → error "answer query failed"

- **Repo**: `k-forum-api`

### Ringkasan

Endpoint detail pertanyaan publik di mobile selalu gagal saat memuat daftar jawabannya (pertanyaan sendiri berhasil termuat, tapi bagian jawaban error).

### Root cause (code review)

`GET /api/v1/mobile/qna/questions/public/{question_id}` → `GetPublicQuestionDetail` (`internal/infrastructure/persistence/postgres_qna_query.go:501-618`). Query pertama (ambil data pertanyaan) berhasil — errornya spesifik muncul di query KEDUA (ambil daftar jawaban, baris 543-555):

```sql
SELECT a.id, a.answered_by, u.fullname, ...
FROM qna_answers a
JOIN users u ON u.id = a.answered_by
WHERE a.question_id=$1 AND a.status='visible'
  AND ($2 = '' OR a.answered_by NOT IN (SELECT blocked_id FROM blocked_users WHERE blocker_id=$2))
ORDER BY a.is_accepted DESC, a.upvote_count DESC, a.created_at ASC
```

Parameter `$2` (userID) dipakai dalam **dua konteks tipe yang berbeda dalam satu query**:
- `$2 = ''` → PostgreSQL infer sebagai perbandingan teks.
- `blocker_id=$2` → `blocker_id` bertipe `UUID`, jadi PostgreSQL infer `$2` sebagai `uuid`.

PostgreSQL (lewat extended query protocol yang dipakai `lib/pq`) butuh **satu tipe konsisten** untuk tiap placeholder di seluruh query — kombinasi dua konteks ini rawan menghasilkan error tipe saat prepare statement (mis. `operator does not exist: uuid = unknown` atau `could not determine data type of parameter $2`), yang ter-wrap jadi `CodeQnaAnswerQueryFailed` (baris 557/575/614).

**Bandingkan dengan pola yang benar** di fungsi tetangga `ListPublicQuestions` (baris 440-444, ditambahkan di commit yang sama) — di sana klausa blocked-user **ditambahkan secara kondisional di kode Go**, bukan di-embed statis di SQL:
```go
if query.UserID != "" {
    whereClauses = append(whereClauses, fmt.Sprintf("qu.user_id NOT IN (SELECT blocked_id FROM blocked_users WHERE blocker_id = $%d)", argIdx))
    args = append(args, query.UserID)
    argIdx++
}
```
Di sini parameter hanya pernah dipakai dalam SATU konteks tipe (selalu `uuid`, klausanya cuma ada kalau userID memang terisi) — jadi tidak kena masalah ini. `GetPublicQuestionDetail` memakai pola yang berbeda (guard `$2 = ''` di dalam SQL) untuk kasus yang sama, dan pola itulah yang bermasalah.

### Yang diminta ke backend (k-forum-api)

1. Ubah query jawaban di `GetPublicQuestionDetail` supaya mengikuti pola yang sama seperti `ListPublicQuestions` — bangun klausa `WHERE` secara kondisional di Go (tambahkan filter blocked-user hanya kalau `userID != ""`), bukan pakai guard `$2 = ''` di dalam SQL statis.
2. Setelah fix, test dengan user yang benar-benar login (userID terisi) — ini kasus yang selalu kena karena mobile mengharuskan login untuk sebagian besar aksi.

### Kriteria selesai (acceptance)

- [ ] `GET /mobile/qna/questions/public/{question_id}` oleh user yang login → berhasil 200, daftar jawaban termuat lengkap.
- [ ] Jawaban dari user yang di-block oleh requester tetap tersembunyi dari daftar (fungsi block-nya jangan sampai regresi saat diperbaiki).
- [ ] Guest (belum login, kalau endpoint ini memang dibuka untuk guest) tetap bisa lihat jawaban tanpa error.

---

## Referensi

- Journey terkait: [`flows/user-journeys/05_QNA_JOURNEY.md`](../flows/user-journeys/05_QNA_JOURNEY.md).
- Spec: `Modules/Q&A/QNA_RULES.md`, `Modules/Q&A/API_SPEC_QNA_BACKOFFICE.md` (B18), `API SPEC/Mobile/API_SPEC_QNA_MOBILE.md`.
- Kode kunci — `k-forum-api`: `internal/infrastructure/persistence/postgres_qna_query.go` (`getQuestion`, `GetPublicQuestionDetail`, `ListPublicQuestions`), `internal/app/usecase/qna/qna_mapper.go`, `internal/app/usecase/qna/get_qna_attachment_presign_url.go`, `internal/interfaces/http/router/router.go`.
- Kode kunci — `k-forum-backoffice`: `app/pages/qa/index.vue`, `app/stores/qnaStore.ts`.
