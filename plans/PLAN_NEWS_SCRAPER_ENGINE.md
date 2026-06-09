# Plan: News Scraper Engine — k-forum-api

## Context

Domain `newssource` (news_sources, source_selectors, source_categories) + backoffice CRUD endpoints sudah selesai (migration 0025). Scraper engine bertanggung jawab membaca config sumber berita dari DB, melakukan crawl RSS feed secara terjadwal, mengekstrak konten artikel (Readability + CSS selector + multipage), dedup berdasarkan `original_url`, menyimpan ke tabel `articles` + `article_translations`, dan mengantri translation otomatis antar bahasa.

**Arsitektur saat ini:**
- `cmd/app` — HTTP API server
- `cmd/worker` — background worker: OutboxRelay, DeliveryRetryRelay, BroadcastFanoutRelay + consumer MQ
- `internal/app/service/` — app-layer services (job, outbox, media, notification)
- `internal/app/port/` — port interfaces untuk dependency injection
- `internal/infrastructure/persistence/` — implementasi persistence
- `internal/interfaces/mq/relay/` — relay polling (outbox, delivery, dll)

**Keputusan arsitektur:**
Scraper engine diletakkan di `internal/app/service/scraper/` mengikuti pola yang sama dengan service lain (job, outbox, media). Port interfaces (Cleaner, TranslationEnqueuer) di `internal/app/port/`. Scheduler berjalan sebagai goroutine di `cmd/worker`.

---

## Dependency Baru

```bash
go get github.com/mmcdole/gofeed            # RSS/Atom feed parser
go get github.com/PuerkitoBio/goquery       # CSS selector (author, tags fallback)
go get github.com/go-shiori/go-readability  # content extraction (Readability primary)
go get github.com/robfig/cron/v3            # cron expression evaluator
go get golang.org/x/sync                    # semaphore untuk worker pool
```

---

## Layer 1 — App Service: `internal/app/service/scraper/`

### `types.go`

```go
package scraper

import "time"

type Source struct {
    ID               string
    Key              string
    Name             string
    BaseURL          string
    OriginalLanguage string
    Schedule         string      // cron expression, e.g. "0 */6 * * *"
    LastScrapedAt    *time.Time
    AutoPublish      bool
    AICleanup        bool
    AutoTranslate    bool
    Selectors        Selectors
    Categories       []Category
}

type Selectors struct {
    ContentSelector *string
    AuthorSelector  *string
    TagsSelector    *string
    ExtraFields     map[string]any
}

type Category struct {
    ID          string
    CategoryKey string
    URLSuffix   *string
    URLOverride *string
    ArticleLimit int
}

type FeedItem struct {
    Title       string
    Link        string
    Description string
    PubDate     *time.Time
    Thumbnail   *string
}

type ArticleDetail struct {
    Content     string
    ContentHTML string
    Excerpt     string
    Author      string
    Tags        []string
}

type SaveScrapedInput struct {
    SourceID         string
    CategoryKey      string
    OriginalLanguage string
    OriginalURL      string
    Title            string
    Content          string
    ContentHTML      string
    Summary          string
    Author           string
    Thumbnail        *string
    Tags             []string
    PubDate          *time.Time
    AutoPublish      bool
    AuthorLabel      string
}
```

### `repository.go`

```go
package scraper

import (
    "context"
    "time"
)

type SourceRepo interface {
    ListActive(ctx context.Context) ([]Source, error)
    UpdateLastScraped(ctx context.Context, id string, t time.Time) error
}

type ArticleRepo interface {
    // Return "" if original_url already exists (ON CONFLICT DO NOTHING)
    SaveScraped(ctx context.Context, input SaveScrapedInput) (articleID string, err error)
    ExistingURLs(ctx context.Context, urls []string) (map[string]bool, error)
}
```

### `fetcher.go`

```go
// BuildFeedURL returns urlOverride if set, else base+suffix, else base alone.
func BuildFeedURL(base, suffix, override string) string

// FetchFeed fetches and parses an RSS/Atom feed. Timeout 15s, User-Agent browser.
func FetchFeed(ctx context.Context, feedURL string) ([]FeedItem, error)
```

### `detail.go`

Strategi:
1. `go-readability` untuk content extraction (primary)
2. `goquery` sebagai fallback untuk `author` dan `tags` via CSS selector
3. Multipage loop: jika ada link "next page", fetch + gabungkan content (max 10 halaman)
4. Special case: domain Tribun append `?page=all` ke URL sebelum fetch

```go
func FetchDetail(ctx context.Context, articleURL string, sel Selectors) ArticleDetail
```

### `dedup.go`

```go
// FilterNew filters out items whose links already exist in the DB.
// Fail-open: jika ExistingURLs error, return semua items (tidak block scraping).
func FilterNew(ctx context.Context, items []FeedItem, repo ArticleRepo) []FeedItem
```

### `pipeline.go`

```go
type Pipeline struct {
    articleRepo ArticleRepo
    sourceRepo  SourceRepo
    cleaner     port.NewsCleaner              // pluggable: NoopNewsCleaner atau AI cleaner
    translator  port.NewsTranslationEnqueuer  // pluggable: Postgres enqueuer atau Noop
    maxParallel int64                         // worker pool size, default 5
}

func NewPipeline(
    articleRepo ArticleRepo,
    sourceRepo  SourceRepo,
    cleaner     port.NewsCleaner,
    translator  port.NewsTranslationEnqueuer,
    maxParallel int64,
) *Pipeline

// ScrapeSource menjalankan full pipeline untuk satu source.
// Per-category:
//   BuildFeedURL → FetchFeed → FilterNew → limit(cat.ArticleLimit)
//   → worker pool (semaphore + errgroup):
//       FetchDetail → cleaner.Clean(ContentHTML)
//       → SaveScraped → jika AutoTranslate && articleID != "" → translator.EnqueueAll
// UpdateLastScraped sekali per source setelah semua category selesai.
// Artikel yang gagal di-skip (log warning), bukan error fatal.
func (p *Pipeline) ScrapeSource(ctx context.Context, src Source) error
```

### `scheduler.go`

```go
type Scheduler struct {
    sourceRepo SourceRepo
    pipeline   *Pipeline
}

func NewScheduler(sourceRepo SourceRepo, pipeline *Pipeline) *Scheduler

// Run adalah blocking goroutine. Tick setiap 1 menit.
// Untuk tiap source aktif: evaluasi isDue(src, now) → ScrapeSource di goroutine terpisah.
func (s *Scheduler) Run(ctx context.Context)

// isDue mengecek apakah source perlu di-scrape berdasarkan cron expression + last_scraped_at.
func isDue(src Source, now time.Time) bool
```

---

## Layer 2 — Port Interfaces: `internal/app/port/scraper_port.go`

```go
package port

import "context"

// NewsCleaner membersihkan HTML artikel (AI-based atau noop).
type NewsCleaner interface {
    Clean(ctx context.Context, html string) (string, error)
}

// NewsTranslationEnqueuer mengantri article_translation rows untuk setiap target language.
type NewsTranslationEnqueuer interface {
    EnqueueAll(ctx context.Context, articleID string) error
}

// NoopNewsCleaner — implementasi default, AI cleanup diisi nanti.
type NoopNewsCleaner struct{}

func (n *NoopNewsCleaner) Clean(_ context.Context, html string) (string, error) {
    return html, nil
}

// NoopNewsTranslationEnqueuer — implementasi default untuk build tanpa provider.
type NoopNewsTranslationEnqueuer struct{}

func (n *NoopNewsTranslationEnqueuer) EnqueueAll(_ context.Context, _ string) error {
    return nil
}
```

---

## Layer 3 — Persistence Adapters: `internal/infrastructure/persistence/`

### `postgres_scraper_source_repo.go`

Implements `scraper.SourceRepo`:

```go
// ListActive mengambil semua source aktif beserta selectors + categories.
// Query: SELECT news_sources + LEFT JOIN source_selectors + source_categories WHERE is_active=true
// Assemble per-source struct dari rows (multiple rows per source karena JOIN categories).
func (r *PostgresScraperSourceRepo) ListActive(ctx context.Context) ([]scraper.Source, error)

// UpdateLastScraped: UPDATE news_sources SET last_scraped_at=$2, updated_at=NOW() WHERE id=$1
func (r *PostgresScraperSourceRepo) UpdateLastScraped(ctx context.Context, id string, t time.Time) error
```

### `postgres_scraper_article_repo.go`

Implements `scraper.ArticleRepo`:

```go
// SaveScraped:
// 1. Lookup news_category_id FROM source_categories WHERE source_id=$1 AND category_key=$2
// 2. INSERT INTO articles (..., is_manual=false, status='draft'|'published' based on AutoPublish)
//    ON CONFLICT (original_url) DO NOTHING RETURNING id
// 3. rows_affected == 0 → return "", nil (artikel sudah ada)
// 4. INSERT INTO article_translations (is_original=true, translate_status=NULL)
// 5. return articleID
func (r *PostgresScraperArticleRepo) SaveScraped(ctx context.Context, input scraper.SaveScrapedInput) (string, error)

// ExistingURLs: SELECT original_url FROM articles WHERE original_url = ANY($1)
func (r *PostgresScraperArticleRepo) ExistingURLs(ctx context.Context, urls []string) (map[string]bool, error)
```

**Pre-condition**: `articles.original_url` harus punya UNIQUE constraint. Verifikasi di migration 0023. Jika belum ada, tambah migration 0026:
```sql
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS uq_articles_original_url
ON articles (original_url) WHERE original_url IS NOT NULL;
```

### `postgres_translation_enqueuer.go`

Implements `port.NewsTranslationEnqueuer`:

```go
// EnqueueAll:
// 1. SELECT original_language FROM articles WHERE id=$articleID
// 2. SELECT code FROM system_languages WHERE is_translate_target=true AND code != $originalLanguage
// 3. INSERT INTO article_translations (article_id, language, translate_status='pending', is_original=false)
//    ON CONFLICT (article_id, language) DO NOTHING
func (e *PostgresTranslationEnqueuer) EnqueueAll(ctx context.Context, articleID string) error
```

---

## Layer 4 — Translation Relay: `internal/interfaces/mq/relay/translation_relay.go`

Mengikuti pola `outbox_relay.go` (blocking `Run`, polling interval):

```go
type TranslationRelay struct {
    db       *sql.DB
    provider port.NewsTranslationProvider  // interface sudah ada di internal/app/port/news_port.go
    interval time.Duration                 // default 30 detik
    batch    int                           // default 10
}

func NewTranslationRelay(db *sql.DB, provider port.NewsTranslationProvider,
    interval time.Duration, batch int) *TranslationRelay

// Run loop setiap interval:
//   BEGIN TX
//   SELECT ... FROM article_translations
//   WHERE translate_status = 'pending'
//   LIMIT $batch FOR UPDATE SKIP LOCKED
//   Untuk tiap row:
//     UPDATE translate_status = 'processing'
//     fetch original translation (is_original=true) untuk source content
//     provider.Translate(ctx, content, srcLang, targetLang)
//     UPDATE title/content/summary/translate_status = 'done'
//   Jika error → translate_status = 'failed'
//   COMMIT
func (r *TranslationRelay) Run(ctx context.Context)
```

`FOR UPDATE SKIP LOCKED` penting agar multi-instance worker tidak memproses translation yang sama.

Default provider: `NoopTranslationProvider` (return original text) untuk build + test tanpa API key.

---

## Layer 5 — Scrape-Now Wiring (HTTP → Outbox → Worker)

### `internal/domain/news/event/routing.go` — tambahkan constant:

```go
const (
    RoutingNewsScrapeNow = "news.source.scrape_now"
    QueueNewsScrapeNow   = "scraper.source.scrape_now"
)
```

### `internal/app/usecase/newssource/trigger_scrape_now.go`

```go
type ScrapeNowPayload struct {
    SourceID string `json:"source_id"`
}

type TriggerScrapeNowUseCase struct {
    sourceRepo nsrepo.NewsSourceRepository
    outboxRepo outboxsvc.Repository
}

// Execute: verifikasi source ada → save outbox entry {RoutingKey: RoutingNewsScrapeNow, Payload: ScrapeNowPayload}
func (uc *TriggerScrapeNowUseCase) Execute(ctx context.Context, sourceID string) error
```

Tambah `TriggerScrapeNow *TriggerScrapeNowUseCase` ke `UseCases` dan `Dependencies` di `dependencies.go`.

### `internal/interfaces/http/handler/web/news_source_handler.go` — update dari placeholder:

```go
func (h *NewsSourceHandler) ScrapeNow(c *gin.Context) {
    sourceID := c.Param("source_id")
    if err := h.useCases.TriggerScrapeNow.Execute(c.Request.Context(), sourceID); err != nil {
        httperror.Handle(c, err)
        return
    }
    respond.OK(c, "scrape triggered", nil)
}
```

### `internal/interfaces/mq/handlers/scrape_now_handler.go`

```go
type ScrapeNowHandler struct {
    sourceRepo scraper.SourceRepo
    pipeline   *scraper.Pipeline
}

// Handle: unmarshal payload → cari source by ID dari ListActive → pipeline.ScrapeSource
func (h *ScrapeNowHandler) Handle(ctx context.Context, msg mq.JobMessage) error
```

---

## `cmd/worker/main.go` Changes

```go
// Scraper persistence adapters
scraperSourceRepo := persistence.NewPostgresScraperSourceRepo(db)
scraperArticleRepo := persistence.NewPostgresScraperArticleRepo(db)
translationEnqueuer := persistence.NewPostgresTranslationEnqueuer(db, langRepo)

// Scraper pipeline + scheduler
scraperPipeline := scraper.NewPipeline(
    scraperArticleRepo, scraperSourceRepo,
    &port.NoopNewsCleaner{}, translationEnqueuer, 5,
)
scraperScheduler := scraper.NewScheduler(scraperSourceRepo, scraperPipeline)
go scraperScheduler.Run(ctx)

// Translation relay
translationRelay := relay.NewTranslationRelay(db, &port.NoopTranslationProvider{}, 30*time.Second, 10)
go translationRelay.Run(ctx)

// Scrape-now consumer
scrapeNowHandler := mqhandlers.NewScrapeNowHandler(scraperSourceRepo, scraperPipeline)
registry.Register(newsevent.RoutingNewsScrapeNow, scrapeNowHandler.Handle)
// tambah ke bindings: Queue=QueueNewsScrapeNow, RoutingKey=RoutingNewsScrapeNow, MaxRetry=1
```

---

## Implementation Order

1. `go get` semua 5 dependency baru
2. `internal/app/port/scraper_port.go` — NewsCleaner, NewsTranslationEnqueuer + Noop impls
3. `internal/app/service/scraper/types.go` + `repository.go`
4. `internal/app/service/scraper/fetcher.go` + `detail.go` + `dedup.go`
5. `internal/app/service/scraper/pipeline.go` + `scheduler.go`
6. Verifikasi UNIQUE constraint `articles.original_url` (migration 0023) → tambah 0026 jika belum ada
7. `internal/infrastructure/persistence/postgres_scraper_source_repo.go`
8. `internal/infrastructure/persistence/postgres_scraper_article_repo.go`
9. `internal/infrastructure/persistence/postgres_translation_enqueuer.go`
10. `internal/domain/news/event/routing.go` — tambah scrape_now routing constants
11. `internal/app/usecase/newssource/trigger_scrape_now.go` + update `dependencies.go`
12. `internal/interfaces/http/handler/web/news_source_handler.go` — update ScrapeNow dari placeholder
13. `internal/interfaces/mq/relay/translation_relay.go`
14. `internal/interfaces/mq/handlers/scrape_now_handler.go`
15. `cmd/worker/main.go` — wire semua komponen
16. `go build ./...` — verifikasi compile bersih

---

## Catatan Penting

- **Idempotency**: `SaveScraped` menggunakan `ON CONFLICT (original_url) DO NOTHING` — scraping ulang tidak duplikat artikel.
- **Fail-open dedup**: Jika `ExistingURLs` gagal (DB timeout dll), pipeline tetap lanjut dengan semua items — lebih baik duplikat sementara daripada skip artikel baru.
- **NoopCleaner + NoopTranslationProvider**: Build bersih tanpa API key. Pluggable nanti ketika AI cleanup atau translation provider tersedia.
- **`articles.source_id`**: harus nullable karena artikel manual tidak punya source.
- **`FOR UPDATE SKIP LOCKED`** di TranslationRelay: aman untuk multi-instance deployment.
- **Scheduler vs cron library**: Scheduler menggunakan `robfig/cron/v3` hanya sebagai parser expression, bukan sebagai cron runner penuh — loop tick-per-menit lebih mudah dikontrol dengan `ctx.Done()`.

---

## Verifikasi End-to-End

1. `go build ./...` — clean compile
2. `cmd/worker` start → log "scraper scheduler started", tick log setiap menit
3. POST `/api/v1/web/news/sources/:id/scrape-now` → row baru di `event_outbox` dengan routing_key `news.source.scrape_now`
4. OutboxRelay publish ke RabbitMQ → worker consumer handle → `pipeline.ScrapeSource` dipanggil
5. Cek `articles` table: row baru dengan `is_manual=false` + `source_id` terisi
6. Cek `article_translations` table: row `is_original=true` + konten terisi
7. Jika `auto_translate=true`: row tambahan dengan `translate_status='pending'` per target language
8. TranslationRelay pick up pending rows → set status=`done` (NoopProvider: content = original)
