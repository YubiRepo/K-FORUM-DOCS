# News Scraper Engine — Technical Design (Golang)

> **Stack:** Golang + PostgreSQL  
> **Dibuat:** 2026-06-04  
> **Referensi:** Port dari prototipe TypeScript (crawler-agents) ke Golang production

---

## Daftar Isi

1. [Overview](#overview)
2. [Library Golang](#library-golang)
3. [Struktur Package](#struktur-package)
4. [Komponen Inti](#komponen-inti)
5. [Flow End-to-End](#flow-end-to-end)
6. [Scheduler & Cron](#scheduler--cron)
7. [Concurrency & Rate Limit](#concurrency--rate-limit)
8. [Error Handling & Retry](#error-handling--retry)
9. [Mapping dari Prototipe TS](#mapping-dari-prototipe-ts)

---

## Overview

Scraper engine bertugas mengambil berita dari sumber RSS eksternal secara terjadwal, mengekstrak konten lengkap, lalu menyimpannya ke database. Prosesnya:

```
RSS Feed → List artikel → Dedup → Fetch detail → Extract konten
        → (AI cleanup) → Simpan DB → (enqueue translation) → publish/draft
```

Prinsip desain:
- **Eager fetch** — konten lengkap diambil saat scraping, bukan saat user baca (translation butuh full content)
- **Dedup-first** — filter `original_url` sebelum fetch detail untuk hemat HTTP request
- **Readability-based extraction** — ekstraksi konten utama via algoritma Readability, selector hanya fallback + author/tags
- **Per-source config** — jadwal, auto_publish, ai_cleanup, auto_translate diatur per source di DB

---

## Library Golang

| Kebutuhan | Library | Alasan |
|---|---|---|
| RSS/Atom parsing | `github.com/mmcdole/gofeed` | Handle RSS + Atom + JSON feed otomatis, parse pubDate & enclosure |
| HTML query/selector | `github.com/PuerkitoBio/goquery` | Equivalent cheerio — jQuery-style selector |
| Readability extraction | `github.com/go-shiori/go-readability` | Port Mozilla Readability ke Go |
| Cron scheduling | `github.com/robfig/cron/v3` | Parse & eval cron expression |
| HTTP client | `net/http` (stdlib) | Cukup, dengan custom transport untuk timeout & header |
| PostgreSQL | `github.com/jackc/pgx/v5` | Driver Postgres performa tinggi |
| Worker pool | `golang.org/x/sync/errgroup` + semaphore | Concurrency terkontrol |

Instalasi:
```bash
go get github.com/mmcdole/gofeed
go get github.com/PuerkitoBio/goquery
go get github.com/go-shiori/go-readability
go get github.com/robfig/cron/v3
go get github.com/jackc/pgx/v5
go get golang.org/x/sync/errgroup
```

---

## Struktur Package

```
internal/news/
├── scraper/
│   ├── scheduler.go        # Cron scheduler — trigger scraping per source
│   ├── fetcher.go          # Fetch RSS feed → list artikel
│   ├── detail.go           # Fetch & extract detail artikel (Readability + selector + multipage)
│   ├── dedup.go            # Cek original_url di DB
│   ├── cleanup.go          # AI content cleanup (opsional)
│   ├── pipeline.go         # Orkestrasi: fetch → dedup → detail → save
│   └── types.go            # Struct: Source, FeedItem, ArticleDetail
├── translate/
│   ├── translator.go       # Interface + impl Google/OpenAI
│   ├── worker.go           # Background worker proses translation queue
│   └── queue.go            # Enqueue/dequeue translation jobs
├── repository/
│   ├── source_repo.go      # Query news_sources, source_selectors, source_categories
│   ├── article_repo.go     # Insert/update articles + article_translations
│   └── settings_repo.go    # news_system_settings
└── service/
    └── news_service.go     # Business logic, dipanggil dari HTTP handler
```

---

## Komponen Inti

### 1. Types

```go
package scraper

import "time"

// Source merepresentasikan satu news_source + config-nya.
type Source struct {
	ID               string
	Key              string
	Name             string
	BaseURL          string
	OriginalLanguage string
	Schedule         string // cron expression
	LastScrapedAt    *time.Time
	AutoPublish      bool
	AICleanup        bool
	AutoTranslate    bool
	IsActive         bool

	Selectors  Selectors
	Categories []Category
}

type Selectors struct {
	Content     string                 // fallback selector (boleh kosong)
	Author      string
	Tags        string
	ExtraFields map[string]interface{} // multipage rules, dll
}

type Category struct {
	ID           string
	CategoryKey  string
	URLSuffix    string
	URLOverride  string
	ArticleLimit int
	IsActive     bool
}

// FeedItem = satu item dari RSS feed (sebelum fetch detail).
type FeedItem struct {
	Link        string
	Title       string
	PubDate     *time.Time
	Description string // excerpt dari RSS
	Thumbnail   string
}

// ArticleDetail = hasil fetch + extract halaman detail.
type ArticleDetail struct {
	Content     string   // plain text (untuk translation & search)
	ContentHTML string   // HTML (untuk render)
	Excerpt     string
	Author      string
	Tags        []string
	Err         error
}
```

### 2. Fetcher (RSS → List)

```go
package scraper

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// FetchFeed mengambil RSS feed untuk satu category URL dan mengembalikan list item.
func FetchFeed(ctx context.Context, feedURL string) ([]FeedItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned status %d", resp.StatusCode)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse feed: %w", err)
	}

	items := make([]FeedItem, 0, len(feed.Items))
	for _, it := range feed.Items {
		fi := FeedItem{
			Link:        it.Link,
			Title:       it.Title,
			Description: cleanHTML(it.Description),
		}
		if it.PublishedParsed != nil {
			fi.PubDate = it.PublishedParsed
		}
		// thumbnail dari enclosure atau image
		if len(it.Enclosures) > 0 {
			fi.Thumbnail = it.Enclosures[0].URL
		} else if it.Image != nil {
			fi.Thumbnail = it.Image.URL
		}
		items = append(items, fi)
	}
	return items, nil
}

// BuildFeedURL menggabungkan base_url + suffix, atau pakai override.
func BuildFeedURL(base, suffix, override string) string {
	if override != "" {
		return override
	}
	return base + suffix
}
```

### 3. Detail Crawler (Readability + Selector + Multipage)

```go
package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	readability "github.com/go-shiori/go-readability"
)

const maxPages = 10

// FetchDetail mengambil & mengekstrak konten artikel dari URL detail.
// Mengikuti multipage jika ada.
func FetchDetail(ctx context.Context, articleURL string, sel Selectors) ArticleDetail {
	var (
		allContent  strings.Builder
		allHTML     strings.Builder
		firstAuthor string
		firstTags   []string
		visited     = map[string]bool{}
		current     = articleURL
		pageCount   = 0
	)

	for current != "" && !visited[current] && pageCount < maxPages {
		visited[current] = true
		pageCount++

		body, rawHTML, err := httpGet(ctx, current)
		if err != nil {
			if pageCount == 1 {
				return ArticleDetail{Err: err}
			}
			break // halaman lanjutan gagal — pakai yang sudah terkumpul
		}

		// Readability ekstrak konten utama
		parsedURL, _ := url.Parse(current)
		article, rerr := readability.FromReader(strings.NewReader(rawHTML), parsedURL)
		if rerr == nil {
			if allContent.Len() > 0 {
				allContent.WriteString("\n\n")
			}
			allContent.WriteString(strings.TrimSpace(article.TextContent))

			if allHTML.Len() > 0 {
				allHTML.WriteString(`<hr class="page-divider">`)
			}
			allHTML.WriteString(article.Content)
		}

		// Author & tags hanya dari halaman pertama
		if pageCount == 1 {
			doc, derr := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
			if derr == nil {
				firstAuthor = extractAuthor(doc, sel.Author, article)
				firstTags = extractTags(doc, sel.Tags)
			}
		}

		// Deteksi halaman berikutnya
		current = detectNextPage(body, current, pageCount, sel.ExtraFields)
	}

	if allContent.Len() == 0 {
		return ArticleDetail{Err: fmt.Errorf("no content extracted")}
	}

	return ArticleDetail{
		Content:     strings.TrimSpace(allContent.String()),
		ContentHTML: strings.TrimSpace(allHTML.String()),
		Author:      firstAuthor,
		Tags:        firstTags,
	}
}

func httpGet(ctx context.Context, target string) (*goquery.Document, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("Referer", "https://www.google.com/")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", err
	}
	html, _ := doc.Html()
	return doc, html, nil
}

func extractAuthor(doc *goquery.Document, selector string, article *readability.Article) string {
	if article != nil && article.Byline != "" {
		return normalizeSpace(article.Byline)
	}
	if selector == "" {
		selector = ".author, .writer, .byline, .reporter, .penulis"
	}
	return normalizeSpace(doc.Find(selector).First().Text())
}

func extractTags(doc *goquery.Document, selector string) []string {
	if selector == "" {
		selector = ".tags a, .tag-links a, .detail-tag a, [href*='/tag/']"
	}
	var tags []string
	seen := map[string]bool{}
	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		t := strings.TrimSpace(s.Text())
		if t != "" && !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	})
	return tags
}

// detectNextPage mengembalikan URL halaman berikutnya, atau "" jika tidak ada.
func detectNextPage(doc *goquery.Document, current string, pageCount int, extra map[string]interface{}) string {
	// Special case: Tribun pakai ?page=all
	if strings.Contains(current, "tribunnews.com") && !strings.Contains(current, "page=all") {
		if strings.Contains(current, "?") {
			return current + "&page=all"
		}
		return current + "?page=all"
	}

	var next string
	doc.Find("a").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := strings.ToLower(strings.TrimSpace(s.Text()))
		href, exists := s.Attr("href")
		if !exists {
			return true
		}
		if text == "selanjutnya" || text == "next" || text == fmt.Sprint(pageCount+1) {
			if u, err := resolveURL(current, href); err == nil {
				next = u
				return false
			}
		}
		return true
	})
	return next
}

func resolveURL(base, href string) (string, error) {
	b, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	r, err := url.Parse(href)
	if err != nil {
		return "", err
	}
	return b.ResolveReference(r).String(), nil
}

func normalizeSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func cleanHTML(s string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
	if err != nil {
		return s
	}
	return strings.TrimSpace(doc.Text())
}
```

### 4. Dedup

```go
package scraper

import "context"

// FilterNew membuang item yang original_url-nya sudah ada di DB.
func (p *Pipeline) FilterNew(ctx context.Context, items []FeedItem) []FeedItem {
	urls := make([]string, len(items))
	for i, it := range items {
		urls[i] = it.Link
	}

	existing, err := p.articleRepo.ExistingURLs(ctx, urls) // SELECT original_url WHERE original_url = ANY($1)
	if err != nil {
		return items // fail-open: lebih baik proses ulang daripada kehilangan berita
	}

	out := items[:0]
	for _, it := range items {
		if !existing[it.Link] {
			out = append(out, it)
		}
	}
	return out
}
```

### 5. Pipeline (Orkestrasi)

```go
package scraper

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Pipeline struct {
	articleRepo ArticleRepo
	sourceRepo  SourceRepo
	cleaner     Cleaner    // AI cleanup
	translator  Translator // enqueue translation
	maxParallel int64      // mis. 5
}

// ScrapeSource memproses satu source: semua category aktif-nya.
func (p *Pipeline) ScrapeSource(ctx context.Context, src Source) error {
	for _, cat := range src.Categories {
		if !cat.IsActive {
			continue
		}
		if err := p.scrapeCategory(ctx, src, cat); err != nil {
			slog.Error("scrape category failed",
				"source", src.Key, "category", cat.CategoryKey, "err", err)
			// lanjut ke category lain — jangan stop semua
		}
	}
	// update last_scraped_at sekali per source
	return p.sourceRepo.UpdateLastScraped(ctx, src.ID, time.Now())
}

func (p *Pipeline) scrapeCategory(ctx context.Context, src Source, cat Category) error {
	feedURL := BuildFeedURL(src.BaseURL, cat.URLSuffix, cat.URLOverride)

	// 1. Fetch RSS list
	items, err := FetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}

	// 2. Dedup SEBELUM fetch detail
	items = p.FilterNew(ctx, items)

	// 3. Batasi sesuai article_limit
	if len(items) > cat.ArticleLimit {
		items = items[:cat.ArticleLimit]
	}
	if len(items) == 0 {
		return nil // tidak ada yang baru
	}

	// 4. Fetch detail concurrent dengan worker pool
	sem := semaphore.NewWeighted(p.maxParallel)
	g, gctx := errgroup.WithContext(ctx)

	for _, it := range items {
		it := it
		if err := sem.Acquire(gctx, 1); err != nil {
			break
		}
		g.Go(func() error {
			defer sem.Release(1)

			detail := FetchDetail(gctx, it.Link, src.Selectors)
			if detail.Err != nil {
				slog.Warn("fetch detail failed", "url", it.Link, "err", detail.Err)
				return nil // skip artikel ini, jangan gagalkan batch
			}

			// 5. AI cleanup opsional
			content := detail.Content
			contentHTML := detail.ContentHTML
			if src.AICleanup {
				if cleaned, err := p.cleaner.Clean(gctx, contentHTML); err == nil {
					contentHTML = cleaned
				}
			}

			// 6. Simpan ke DB
			articleID, err := p.articleRepo.SaveScraped(gctx, SaveScrapedInput{
				SourceID:         src.ID,
				CategoryKey:      cat.CategoryKey,
				OriginalLanguage: src.OriginalLanguage,
				OriginalURL:      it.Link,
				Title:            it.Title,
				Content:          content,
				ContentHTML:      contentHTML,
				Summary:          firstNonEmpty(detail.Excerpt, it.Description),
				Author:           detail.Author,
				Thumbnail:        it.Thumbnail,
				Tags:             detail.Tags,
				PubDate:          it.PubDate,
				AutoPublish:      src.AutoPublish,
				AuthorLabel:      "Korean Association Indonesia", // scraping = KAI Pusat
			})
			if err != nil {
				return err
			}

			// 7. Enqueue translation jika auto_translate
			if src.AutoTranslate {
				p.translator.EnqueueAll(gctx, articleID)
			}
			return nil
		})
	}
	return g.Wait()
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
```

---

## Flow End-to-End

```
┌─────────────────────────────────────────────────────────────┐
│ SCHEDULER (tick tiap menit)                                  │
│  └── Untuk tiap source aktif:                                │
│        evaluasi cron + last_scraped_at → due? → trigger      │
└────────────────────────┬────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ PIPELINE.ScrapeSource(source)                                │
│  └── Untuk tiap category aktif:                              │
│        1. BuildFeedURL(base, suffix/override)                │
│        2. FetchFeed → []FeedItem                             │
│        3. FilterNew (dedup via original_url)                 │
│        4. Limit ke article_limit                            │
│        5. Worker pool (max 5 paralel):                       │
│             FetchDetail (Readability + selector + multipage) │
│             → AICleanup? → SaveScraped → EnqueueTranslation? │
│  └── UpdateLastScraped(now)                                 │
└────────────────────────┬────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────────┐
│ TRANSLATION WORKER (background, terpisah)                    │
│  └── Poll article_translations WHERE status='pending'        │
│        → call provider (Google/OpenAI)                       │
│        → simpan hasil, status='done'                         │
└─────────────────────────────────────────────────────────────┘
```

---

## Scheduler & Cron

```go
package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	sourceRepo SourceRepo
	pipeline   *Pipeline
	parser     cron.Parser
}

func NewScheduler(repo SourceRepo, pipe *Pipeline) *Scheduler {
	return &Scheduler{
		sourceRepo: repo,
		pipeline:   pipe,
		parser:     cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	}
}

// Run mulai loop scheduler. Tick tiap menit.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			s.tick(ctx, now)
		}
	}
}

func (s *Scheduler) tick(ctx context.Context, now time.Time) {
	sources, err := s.sourceRepo.ListActive(ctx)
	if err != nil {
		slog.Error("list active sources", "err", err)
		return
	}

	for _, src := range sources {
		if !s.isDue(src, now) {
			continue
		}
		src := src
		go func() {
			if err := s.pipeline.ScrapeSource(ctx, src); err != nil {
				slog.Error("scrape source", "source", src.Key, "err", err)
			}
		}()
	}
}

// isDue mengevaluasi apakah source sudah waktunya di-scrape.
func (s *Scheduler) isDue(src Source, now time.Time) bool {
	sched, err := s.parser.Parse(src.Schedule)
	if err != nil {
		slog.Warn("invalid cron", "source", src.Key, "schedule", src.Schedule)
		return false
	}
	if src.LastScrapedAt == nil {
		return true // belum pernah → jalankan
	}
	// next run setelah last_scraped — apakah sudah lewat?
	next := sched.Next(*src.LastScrapedAt)
	return !next.After(now)
}
```

---

## Concurrency & Rate Limit

- **Per-category worker pool**: maksimal `maxParallel` (default 5) fetch detail berjalan bersamaan via `semaphore.Weighted`. Mencegah membanjiri satu source dengan ratusan request sekaligus.
- **Per-source goroutine**: tiap source di-scrape di goroutine terpisah, jadi source lambat tidak memblokir yang lain.
- **HTTP timeout**: 15 detik untuk RSS, 30 detik untuk detail (halaman artikel lebih berat).
- **Rate limit antar request** (opsional): tambah `time.Sleep` kecil atau token bucket per source jika source sensitif terhadap rapid requests.

---

## Error Handling & Retry

| Skenario | Penanganan |
|---|---|
| RSS feed gagal di-fetch | Log error, skip category ini, lanjut category lain |
| Detail artikel gagal | Log warning, skip artikel ini, lanjut artikel lain (tidak gagalkan batch) |
| Halaman multipage ke-N gagal | Pakai konten yang sudah terkumpul dari halaman sebelumnya |
| Dedup query gagal | Fail-open — proses semua item (lebih baik duplikat-cek di insert daripada kehilangan berita) |
| Insert DB gagal | Return error, batch goroutine berhenti untuk artikel itu |
| AI cleanup gagal | Fallback ke konten asli (sebelum cleanup) |
| Translation gagal | `translate_status = failed`, bisa di-retry worker |
| Invalid cron expression | Log warning, skip source (jangan crash scheduler) |

**Idempotency:** `original_url` UNIQUE di DB. Jika dedup lolos tapi ada race, insert akan kena unique constraint — tangani dengan `ON CONFLICT DO NOTHING`.

---

## Mapping dari Prototipe TS

| Prototipe TS | Golang | Catatan |
|---|---|---|
| `axios.get` | `net/http` | Header User-Agent sama |
| `xml2js` + `xmlParser.ts` | `gofeed` | gofeed handle parsing & date otomatis, tidak perlu manual |
| `@mozilla/readability` + `jsdom` | `go-shiori/go-readability` | Port langsung Readability |
| `cheerio` | `goquery` | Selector API mirip |
| `parseRssItems` | `FetchFeed` | List parsing |
| `fetchArticleDetail` | `FetchDetail` | Termasuk multipage logic |
| `sources.ts` (config statis) | `news_sources` + `source_selectors` + `source_categories` (DB) | Config pindah ke DB, bisa diatur lewat backoffice |
| Multi-page loop (while) | `for` loop dengan `visited` map | Logika identik |
| `cleanHtml` | `cleanHTML` (goquery) | |
| Tribun `?page=all` special case | `detectNextPage` | Dipertahankan |

**Perubahan penting:** Config source yang tadinya hardcoded di `sources.ts` sekarang pindah ke DB (`news_sources`, `source_selectors`, `source_categories`), sehingga bisa ditambah/diubah lewat backoffice tanpa deploy ulang.

---

*Dokumen teknis scraper engine News Module. Stack: Golang + PostgreSQL.*
