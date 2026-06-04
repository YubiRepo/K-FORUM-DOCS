# News Scraper Engine (Golang)

Implementasi scraper engine untuk News Module KAI App. Port dari prototipe TypeScript ke Golang production-ready.

## Struktur

```
internal/news/scraper/
├── types.go        # Struct: Source, Category, FeedItem, ArticleDetail, SaveScrapedInput
├── interfaces.go   # Interface: SourceRepo, ArticleRepo, Cleaner, Translator
├── fetcher.go      # FetchFeed — RSS/Atom → []FeedItem (pakai gofeed)
├── detail.go       # FetchDetail — Readability + selector + multipage handling
├── pipeline.go     # Pipeline — orkestrasi: fetch → dedup → detail → cleanup → save
└── scheduler.go    # Scheduler — cron-based, tick tiap menit
```

## Dependency

```bash
go mod tidy
```

Akan resolve:
- `github.com/mmcdole/gofeed` — RSS/Atom parsing
- `github.com/PuerkitoBio/goquery` — HTML selector
- `github.com/go-shiori/go-readability` — ekstraksi konten utama
- `github.com/robfig/cron/v3` — cron expression
- `golang.org/x/sync` — worker pool (semaphore + errgroup)

## Cara Pakai

Implementasikan 4 interface (`SourceRepo`, `ArticleRepo`, `Cleaner`, `Translator`) dengan backend DB / AI kamu, lalu:

```go
package main

import (
	"context"

	"github.com/kai/news-scraper/internal/news/scraper"
)

func main() {
	// Inject implementasi konkret (pgx repo, OpenAI cleaner, dll)
	var sourceRepo scraper.SourceRepo  = NewPgSourceRepo(db)
	var articleRepo scraper.ArticleRepo = NewPgArticleRepo(db)
	var cleaner scraper.Cleaner         = NewAICleaner(openaiClient) // boleh nil jika tidak dipakai
	var translator scraper.Translator   = NewTranslateQueue(db)      // boleh nil

	pipeline := scraper.NewPipeline(articleRepo, sourceRepo, cleaner, translator, 5)
	sched := scraper.NewScheduler(sourceRepo, pipeline)

	ctx := context.Background()
	sched.Run(ctx) // blocking; jalankan di goroutine jika perlu
}
```

## Catatan Desain

- **Eager fetch**: konten lengkap diambil saat scraping (translation butuh full content).
- **Dedup-first**: filter `original_url` sebelum fetch detail untuk hemat HTTP request.
- **Readability primary**: konten utama via Readability, CSS selector hanya fallback + author/tags.
- **Multipage**: artikel ter-paginate digabung otomatis (termasuk special case Tribun `?page=all`).
- **Fail-soft**: artikel gagal di-skip tanpa menggagalkan batch; dedup fail-open.
- **Idempotent**: `SaveScraped` harus pakai `ON CONFLICT (original_url) DO NOTHING`.

Lihat `NEWS_SCRAPER_TECHNICAL.md` untuk dokumen desain lengkap.
