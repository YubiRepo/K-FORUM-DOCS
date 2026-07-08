# Plan: Pindahkan `author` & `thumbnail_url` News dari `article_translations` ke `articles`

> **Status: Sudah dieksekusi (2026-07-04).** Migration, k-forum-api, dan k-forum-backoffice sudah diimplementasikan sesuai plan ini dalam satu sesi (bukan bertahap seperti §8 — lihat "Catatan Eksekusi" di akhir dokumen untuk deviasi & temuan tambahan).

## Context

Saat ini `author` (nama penulis asli) dan `thumbnail_url` (gambar artikel) disimpan per-baris di `article_translations`. Ini salah secara model data: kedua field ini **tidak berubah** ketika artikel diterjemahkan — penulis asli dan gambar artikel tetap sama di semua bahasa. Menyimpannya per-translation menyebabkan:

- Duplikasi nilai yang identik di setiap baris translation (termasuk hasil auto-translate yang sebenarnya tidak butuh field ini sama sekali).
- Query mobile/web harus `COALESCE` lintas translation (original vs bahasa aktif) untuk resolve satu nilai — kompleksitas yang tidak perlu.
- Lifecycle thumbnail (`NormalizeValue` / `ConfirmUpload` / `MarkDeleted`) berjalan berulang per-translation padahal seharusnya sekali per artikel.
- `AddTranslationUseCase` menerima `author`/`thumbnail_url` di request padahal translation hasil AI tidak pernah mengisi keduanya (bug laten, lihat `request_translation.go`).

Keputusan: pindahkan `author` dan `thumbnail_url` menjadi kolom di `articles` (satu nilai per artikel). Spec sudah diupdate — lihat:
- `schemas/NEWS_DB_SCHEMA.md` §7 (DDL) & Migration 12 (ALTER + backfill)
- `Modules/News/NEWS_RULES.md` (deskripsi entitas)
- `Modules/News/NEWS_MEDIA_FLOWS.md` (lifecycle thumbnail)
- `API SPEC/Web/API_SPEC_NEWS_BACKOFFICE.md` & `API SPEC/Mobile/API_SPEC_NEWS_MOBILE.md` (request/response)

> **Catatan penamaan:** `articles.author_label` (label publisher, mis. "KAI Jakarta") sudah ada dan **beda makna** dari `articles.author` (nama penulis/byline) yang baru ditambahkan. Jangan disatukan.

---

## 1. Database Migration (k-forum-api)

Repo `k-forum-api` pakai `database/sql` + `lib/pq` dengan migration file bernomor (lihat `internal/migrations/0012_create_news_tables.up.sql`), bukan GORM. Tambahkan migration baru (nomor lanjutan, ikuti konvensi CLAUDE.md §12):

**`00XX_move_news_author_thumbnail_to_article.up.sql`**
```sql
ALTER TABLE articles
  ADD COLUMN IF NOT EXISTS author        VARCHAR(200) NULL,
  ADD COLUMN IF NOT EXISTS thumbnail_url TEXT         NULL;

UPDATE articles a
SET author        = t.author,
    thumbnail_url = t.thumbnail_url
FROM article_translations t
WHERE t.article_id = a.id
  AND t.is_original = true;
```

**`00XX_move_news_author_thumbnail_to_article.down.sql`**
```sql
ALTER TABLE articles
  DROP COLUMN IF EXISTS author,
  DROP COLUMN IF EXISTS thumbnail_url;
```

Sengaja **tidak** langsung `DROP COLUMN` di `article_translations` pada migration yang sama — lakukan di migration terpisah setelah kode yang membaca kolom lama benar-benar berhenti jalan (lihat urutan rilis di §5). Migration drop-nya:

**`00XY_drop_news_author_thumbnail_from_translation.up.sql`**
```sql
ALTER TABLE article_translations
  DROP COLUMN IF EXISTS author,
  DROP COLUMN IF EXISTS thumbnail_url;
```

---

## 2. k-forum-api — Domain Entity

**`internal/domain/news/entity/article.go`**
- Tambah `Author *string` dan `ThumbnailURL *string` ke `ArticleParams` (line ~11-22) dan struct `Article` (line ~24-47).
- `NewArticle()` (line ~49-77): terima & normalize kedua field baru (pakai pola `normalizeOptionalStr` yang sudah ada di `article_translation.go`).

**`internal/domain/news/entity/article_translation.go`**
- Hapus `Author *string` (line 18, 33) dan `ThumbnailURL *string` (line 19, 34) dari `ArticleTranslationParams` dan `ArticleTranslation`.
- `NewArticleTranslation()` (line ~43-69): hapus normalisasi `Author`/`ThumbnailURL` (line 59-60).

---

## 3. k-forum-api — DTO (`internal/app/dto/news_dto.go`)

| Struct | Perubahan |
|---|---|
| `CreateNewsArticleRequest` (line 7-18) | Tambah `Author *string`, `ThumbnailURL *string` di level top (sejajar `AuthorLabel`) |
| `UpdateNewsArticleRequest` (line 20-28) | Sama seperti create |
| `UpsertNewsTranslationRequest` (line 197-204) | Hapus `Author`, `ThumbnailURL` (line 201-202) — dipakai juga oleh `AddTranslationUseCase`, jadi ini otomatis membersihkan field yang selama ini tidak seharusnya ada di sana |
| `NewsArticleWebDetail` (line 102-126) | Tambah `Author *string`, `ThumbnailURL *string`, `ThumbnailRaw *string` di level artikel (sejajar `AuthorLabel`/`AuthorRegionID`) |
| `NewsTranslationItem` (line 206-219) | Hapus `Author`, `ThumbnailURL`, `ThumbnailRaw` (line 212-214) — translation response tidak lagi bawa field ini |
| `NewsArticleWebListItem` (line 87-100) | Opsional: tambah `ThumbnailURL` jika backoffice list ingin tampilkan thumbnail (saat ini web list belum menampilkan) |
| `NewsArticleMobileListItem` / `NewsArticleMobileDetail` | Tidak berubah struktur (sudah punya `ThumbnailURL`/`ThumbnailRaw` di level atas) — hanya sumber datanya yang pindah (lihat §4) |

Port layer (`internal/app/port/news_port.go`) ikut menyesuaikan:
- `NewsArticleWebDetailReadItem` (line 82-106): tambah `Author *string`, `ThumbnailURL *string` di level artikel; `NewsArticleTranslationReadItem` (line 68-80) hapus `Author`/`ThumbnailURL`.
- `NewsArticleMobileDetailReadItem` (line 108-134): pindahkan `Author`/`ThumbnailURL` (line 126-127, saat ini di bawah komentar "Resolved translation") ke section article-level bersama `AuthorLabel` (line 114).

---

## 4. k-forum-api — Repository (raw SQL)

**`internal/infrastructure/persistence/postgres_news_repository.go`**
- `PostgresArticleRepository.Save/Update/FindByID` (line 25-103): tambah kolom `author`, `thumbnail_url` ke INSERT/UPDATE/SELECT + scan.
- `PostgresArticleTranslationRepository` (line 127-252): hapus `author`/`thumbnail_url` dari `Save`, `SaveOrUpdate`, `FindByArticleAndLanguage`, `FindAllByArticle`, `scanTranslation` (line 133-244).

**`internal/infrastructure/persistence/postgres_news_query.go`** — ini yang paling banyak berubah karena pola `COALESCE(tl.x, to2.x)` lintas translation harus diganti jadi baca langsung dari `articles`:
- `ListForMobile` (line 145-278): ganti `COALESCE(tl.thumbnail_url, to2.thumbnail_url)` (line 203) → `a.thumbnail_url`; tambah `a.author` bila perlu.
- `GetDetailForWeb` (line 282-358): tambah `a.author`, `a.thumbnail_url` ke SELECT artikel (line 283-294); hapus `author, thumbnail_url` dari query translation (line 327-357).
- `GetDetailForMobile` (line 362-437): ganti `COALESCE(tl.author, to2.author)` dan `COALESCE(tl.thumbnail_url, to2.thumbnail_url)` (line 375-376) → `a.author`, `a.thumbnail_url` langsung, hapus join tambahan yang cuma dipakai untuk ini kalau tidak dipakai field lain.
- `ListByUser` (bookmarks, line 518-579): ganti `COALESCE(tl.thumbnail_url, to2.thumbnail_url)` (line 537) → `a.thumbnail_url`.

**`internal/infrastructure/persistence/postgres_scraper_article_repo.go`**
- `SaveScraped` (line 24-109): tambah `author`, `thumbnail_url` ke INSERT `articles` (line 48-64, dari `input.Author`/`input.Thumbnail`); hapus dari INSERT `article_translations` (line 93-103).

**`internal/infrastructure/persistence/postgres_translation_enqueuer.go`**
- `EnqueueAll` (line 20-69): hapus `author, thumbnail_url` dari `INSERT ... SELECT` (line 51-63) — translation baru tidak lagi copy field ini sama sekali.

**Search indexing — `internal/infrastructure/persistence/postgres_search_query.go`**
- `FindSearchableNews` (line 74-105) & `FindSearchableNewsByID` (line 107-133): ganti subquery `(SELECT t.thumbnail_url FROM article_translations t WHERE ... is_original) AS thumbnail_url` (line 82, 116) → `a.thumbnail_url` langsung.

---

## 5. k-forum-api — Usecase

| File | Perubahan |
|---|---|
| `create_article.go` (line 47-67) | `ArticleTranslation` tidak lagi diisi `Author`/`ThumbnailURL`; field ini sekarang jadi bagian `Article` yang dibangun dari `req.Author`/`req.ThumbnailURL`. `confirmNewsThumbnailKey` jalan atas `article.ThumbnailURL` |
| `update_article.go` (line 56-98) | **Perubahan paling besar.** Diff old/new thumbnail pindah dari level translation ke level artikel: fetch `oldArticle` (bukan `oldTranslation` by language), compare `oldArticle.ThumbnailURL` vs `req.ThumbnailURL`, `MarkDeleted`/`ConfirmUpload` jalan **sekali per update**, bukan tiap kali translation disave |
| `add_translation.go` (line 25-45) | Hapus seluruh blok `Author`/`ThumbnailURL`/`normalizeNewsThumbnailKeyPtr`/`confirmNewsThumbnailKey` (line 32-33, 45) — translation tambahan cuma title/content/summary/tags |
| `delete_article.go` (line 38-41) | Ganti loop `for _, t := range translations { markNewsThumbnailDeleted(t.ThumbnailURL) }` jadi satu panggilan `markNewsThumbnailDeleted(ctx, svc, article.ThumbnailURL)` |
| `submit_article.go` (line 46-69) | Sama seperti `create_article.go` — `Author`/`ThumbnailURL` dari request masuk ke `Article`, bukan `ArticleTranslation` |
| `get_article_web.go` (line 53-68) | `mapToNewsArticleWebDetail`: set `Author`/`ThumbnailURL`/`ThumbnailRaw` di `NewsArticleWebDetail` top-level (dari `article.Author`/`article.ThumbnailURL`), hapus dari `NewsTranslationItem` per bahasa |
| `get_article_mobile.go` (line 37-59) | `ThumbnailURL`/`ThumbnailRaw` diambil dari read-model artikel (sudah article-level di read model setelah §4), tambahkan `Author` juga bila field ini mau ditampilkan mobile (spec sudah expose `author` di Article Object List — lihat `API_SPEC_NEWS_MOBILE.md`) |
| `list_articles_mobile.go` (line 34-56) | Tambah `Author: item.Author` ke `NewsArticleMobileListItem` (field baru sesuai update spec mobile list) |
| `request_translation.go` (line 65-73) | Tidak perlu perubahan — sebelumnya memang tidak copy `Author`/`ThumbnailURL` (bug lama), sekarang jadi correct-by-construction karena field ini memang bukan bagian translation lagi |

**`helpers.go`** (`resolveNewsThumbnailURL`, `normalizeNewsThumbnailKeyPtr`, `confirmNewsThumbnailKey`, `markNewsThumbnailDeleted`, line 85-126) — tidak perlu diubah, generic terhadap `*string`. Tinggal ganti caller-nya dari `translation.ThumbnailURL` → `article.ThumbnailURL`.

Endpoint media standalone (`get_thumbnail_presign_url.go`, `confirm_thumbnail.go`, `delete_thumbnail.go`) — **tidak berubah**, sudah field-agnostic.

---

## 6. k-forum-api — Test & Docs

- Regenerate swagger (`swag init`) setelah DTO berubah — jangan edit `docs/docs.go` manual.
- Tambah/update handler test di `internal/interfaces/http/handler/web/news_handler_test.go` dan `.../mobile/news_handler_test.go` untuk cover `author`/`thumbnail_url` di level artikel (create, update, get detail, list mobile) — sesuai wajib handler test per CLAUDE.md §11.

---

## 7. k-forum-backoffice — Frontend

Semua pemakaian ada di 3 file inti (`app/components/news/ArticleForm.vue`, `app/components/news/ArticleDetail.vue`, `app/types/news.ts`) — **tidak ada** duplikasi field ini di per-language tab, jadi migrasinya murni "ubah sumber data", bukan restrukturisasi UI:

**`app/types/news.ts`**
- `ArticleTranslation` (line 12-23): hapus `author?: string` (line 17), `thumbnail_url?: string | null` (line 18).
- `Article` (line 26-49) / `ArticleDetail` (line 52-62): tambah `author?: string | null`, `thumbnail_url?: string | null`, `thumbnail_raw?: string | null`.

**`app/components/news/ArticleForm.vue`**
- `buildPayload()` (line 87-102): pindahkan `author`/`thumbnail_url` dari objek `translation` ke root payload (sejajar `author_label`):
  ```js
  const buildPayload = () => ({
    category_id: form.category_id,
    news_scope_id: ...,
    original_language: form.original_language,
    author_label: form.author_label,
    author: form.author,                        // pindah ke sini
    thumbnail_url: form.thumbnail_url || null,   // pindah ke sini
    is_featured: form.is_featured,
    status: form.status,
    translation: {
      title: form.title,
      content: form.content,
      summary: form.summary,
      tags: form.tagsText.split(',').map(s => s.trim()).filter(Boolean)
    }
  })
  ```
- Hydration on load (line 50-68): baca `form.author`/`form.thumbnail_url` dari `detail.author`/`detail.thumbnail_raw` (article root), bukan dari `orig` (translation asli). `originalThumbnailUrl` (line 18, 54) ikut baca dari `detail.thumbnail_url`.
- Field input `author`/thumbnail (line 293, 311-320) tetap di tab "General Content" — tidak perlu pindah UI, sudah tampil sekali (bukan per-language) secara visual, cuma binding datanya yang berubah.

**`app/components/news/ArticleDetail.vue`**
- `cover` computed (line 26): ganti `activeTr.value?.thumbnail_url` → `article.value?.thumbnail_url` (article-level, tidak berubah saat user switch bahasa via language switcher line 128-140).
- Tambahkan tampilan `author` bila diinginkan (saat ini fetched tapi tidak pernah dirender).

**`app/stores/newsStore.ts`** — tidak perlu perubahan struktural (`payload: any`, passthrough langsung), pastikan saja endpoint `saveTranslation` (line 96-105) tetap tidak mengirim `author`/`thumbnail_url` (sudah begitu).

---

## 8. Urutan Rilis (hindari breaking change di tengah jalan)

1. **DB migration additive**: jalankan migration §1 bagian "up" pertama (`ADD COLUMN` + backfill ke `articles`). Kolom lama di `article_translations` **belum** dihapus — aman untuk kode lama yang masih jalan.
2. **Deploy k-forum-api** dengan semua perubahan §2–§6 (baca/tulis `articles.author`/`articles.thumbnail_url`, request/response API sudah pakai bentuk baru).
3. **Deploy k-forum-backoffice** dengan perubahan §7 (payload & hydration article-level).
4. Setelah dipastikan stabil (tidak ada trafik yang masih kirim payload lama `translation.author`/`translation.thumbnail_url`), jalankan migration susulan untuk `DROP COLUMN` di `article_translations`.

> Karena `CreateNewsArticleRequest`/`UpdateNewsArticleRequest` berubah bentuk (field pindah keluar dari `translation`), ini **breaking change** di kontrak API backoffice — pastikan step 2 & 3 dirilis berdekatan (idealnya sekaligus) supaya tidak ada window backoffice lama memanggil API baru atau sebaliknya.

---

## 9. Checklist Ringkas

- [x] Migration: `ADD COLUMN articles.author/thumbnail_url` + backfill
- [x] Domain entity: `Article` dapat field baru, `ArticleTranslation` field dihapus
- [x] DTO + port: pindahkan field top-level, regenerate swagger
- [x] Repository: update semua SELECT/INSERT/UPDATE (news_repository, news_query, scraper_article_repo, translation_enqueuer, search_query)
- [x] Usecase: create/update/add_translation/delete/submit/get_article_web/get_article_mobile/list_articles_mobile
- [x] Handler test baru untuk field article-level
- [x] Backoffice: types, `ArticleForm.vue` (buildPayload + hydration), `ArticleDetail.vue` (cover computed)
- [x] Rilis API + backoffice berdekatan (breaking contract change) — dieksekusi sekaligus dalam satu sesi
- [x] Migration susulan: `DROP COLUMN article_translations.author/thumbnail_url` — digabung ke migration yang sama (lihat Catatan Eksekusi)

---

## Catatan Eksekusi (2026-07-04)

Deviasi dari plan asli, dan temuan tambahan selama implementasi:

1. **Migration digabung jadi satu file**, bukan dua (additive lalu drop terpisah). Karena eksekusi dilakukan sekaligus dalam satu sesi kerja (bukan rilis produksi bertahap), `ADD COLUMN` + backfill + `DROP COLUMN article_translations` digabung di migration `20260704043220_move_news_author_thumbnail_to_article`. **Jika ini dijalankan ke database produksi yang sudah punya trafik berjalan**, sebaiknya split kembali jadi dua migration terpisah sesuai urutan rilis §8 untuk menghindari window downtime.
2. **Migration dibuat manual** (bukan via `make migrate-create`) karena `make`/`docker` di lingkungan eksekusi minta password sudo interaktif yang tidak bisa dipenuhi non-interaktif. File tetap mengikuti konvensi penamaan timestamp (`YYYYMMDDHHMMSS_nama_migrasi.up/down.sql`) sesuai CLAUDE.md §12.
3. **Bug pre-existing ditemukan & diperbaiki**: `internal/testhelper/testserver.go` tidak meng-inject `MediaUploadSvc` ke `newsusecase.Dependencies` (beda dengan wiring produksi di `cmd/app/main.go` yang sudah benar). Akibatnya semua resolusi thumbnail URL di News lewat test harness selalu no-op/nil-panic — gap ini sudah ada sebelum migration ini dan baru ketahuan karena test baru untuk `thumbnail_url` menabrak nil pointer. Sudah diperbaiki dengan menambahkan `MediaUploadSvc: mediaUploadSvc` ke wiring test.
4. **`ListArticlesWebUseCase` bertambah dependency** `MediaUploadSvc` (sebelumnya tidak ada) karena field `thumbnail_url` baru ditambahkan ke `NewsArticleWebListItem` (opsional di plan asli, diputuskan untuk diimplementasikan agar backoffice list bisa tampilkan thumbnail tanpa request tambahan).
5. **`NewsArticleMobileDetail.Author`** ditambahkan sebagai field baru — sebelumnya field ini didokumentasikan di `API_SPEC_NEWS_MOBILE.md` tapi tidak pernah diimplementasikan di DTO (gap lama, bukan regresi).
6. Semua test News (web + mobile handler, existing + baru) hijau: `go test ./internal/interfaces/http/handler/web/...` dan `.../mobile/...` full pass, tidak ada regresi ke test lain.
7. Backoffice: `npx vue-tsc --noEmit` bersih (proyek tidak punya `vue-tsc` sebagai devDependency terpasang — dijalankan via `npx` sekali untuk verifikasi, tidak ditambahkan ke `package.json`).
