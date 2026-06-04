package scraper

import (
	"context"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Pipeline mengorkestrasi proses scraping satu source dari awal sampai simpan DB.
type Pipeline struct {
	articleRepo ArticleRepo
	sourceRepo  SourceRepo
	cleaner     Cleaner
	translator  Translator
	maxParallel int64
}

// NewPipeline membuat pipeline baru. maxParallel mengatur jumlah fetch detail
// yang berjalan bersamaan per kategori (default disarankan 5).
func NewPipeline(ar ArticleRepo, sr SourceRepo, c Cleaner, t Translator, maxParallel int64) *Pipeline {
	if maxParallel <= 0 {
		maxParallel = 5
	}
	return &Pipeline{
		articleRepo: ar,
		sourceRepo:  sr,
		cleaner:     c,
		translator:  t,
		maxParallel: maxParallel,
	}
}

// ScrapeSource memproses semua kategori aktif dari satu source,
// lalu memperbarui last_scraped_at.
func (p *Pipeline) ScrapeSource(ctx context.Context, src Source) error {
	for _, cat := range src.Categories {
		if !cat.IsActive {
			continue
		}
		if err := p.scrapeCategory(ctx, src, cat); err != nil {
			slog.Error("scrape category failed",
				"source", src.Key, "category", cat.CategoryKey, "err", err)
			// lanjut ke kategori berikutnya — jangan hentikan seluruh source
		}
	}
	return p.sourceRepo.UpdateLastScraped(ctx, src.ID, time.Now())
}

func (p *Pipeline) scrapeCategory(ctx context.Context, src Source, cat Category) error {
	feedURL := BuildFeedURL(src.BaseURL, cat.URLSuffix, cat.URLOverride)

	// 1. Ambil daftar artikel dari RSS
	items, err := FetchFeed(ctx, feedURL)
	if err != nil {
		return err
	}

	// 2. Dedup SEBELUM fetch detail (hemat HTTP request)
	items = p.filterNew(ctx, items)

	// 3. Batasi sesuai article_limit
	if cat.ArticleLimit > 0 && len(items) > cat.ArticleLimit {
		items = items[:cat.ArticleLimit]
	}
	if len(items) == 0 {
		return nil
	}

	// 4. Fetch detail concurrent dengan worker pool terbatas
	sem := semaphore.NewWeighted(p.maxParallel)
	g, gctx := errgroup.WithContext(ctx)

	for _, it := range items {
		it := it
		if err := sem.Acquire(gctx, 1); err != nil {
			break
		}
		g.Go(func() error {
			defer sem.Release(1)
			return p.processItem(gctx, src, cat, it)
		})
	}
	return g.Wait()
}

func (p *Pipeline) processItem(ctx context.Context, src Source, cat Category, it FeedItem) error {
	detail := FetchDetail(ctx, it.Link, src.Selectors)
	if detail.Err != nil {
		slog.Warn("fetch detail failed", "url", it.Link, "err", detail.Err)
		return nil // skip artikel ini, jangan gagalkan batch
	}

	// 5. AI cleanup opsional
	contentHTML := detail.ContentHTML
	if src.AICleanup && p.cleaner != nil {
		if cleaned, cerr := p.cleaner.Clean(ctx, contentHTML); cerr == nil {
			contentHTML = cleaned
		} else {
			slog.Warn("ai cleanup failed, using raw", "url", it.Link, "err", cerr)
		}
	}

	// 6. Simpan ke DB
	articleID, err := p.articleRepo.SaveScraped(ctx, SaveScrapedInput{
		SourceID:         src.ID,
		CategoryKey:      cat.CategoryKey,
		OriginalLanguage: src.OriginalLanguage,
		OriginalURL:      it.Link,
		Title:            it.Title,
		Content:          detail.Content,
		ContentHTML:      contentHTML,
		Summary:          firstNonEmpty(detail.Excerpt, it.Description),
		Author:           detail.Author,
		Thumbnail:        it.Thumbnail,
		Tags:             detail.Tags,
		PubDate:          it.PubDate,
		AutoPublish:      src.AutoPublish,
		AuthorLabel:      "Korean Association Indonesia", // scraping selalu KAI Pusat
	})
	if err != nil {
		return err
	}
	if articleID == "" {
		// Sudah ada (ON CONFLICT DO NOTHING) — skip translation
		return nil
	}

	// 7. Enqueue translation jika auto_translate aktif
	if src.AutoTranslate && p.translator != nil {
		if terr := p.translator.EnqueueAll(ctx, articleID); terr != nil {
			slog.Warn("enqueue translation failed", "article", articleID, "err", terr)
		}
	}
	return nil
}

// filterNew membuang item yang original_url-nya sudah ada di DB.
func (p *Pipeline) filterNew(ctx context.Context, items []FeedItem) []FeedItem {
	if len(items) == 0 {
		return items
	}
	urls := make([]string, len(items))
	for i, it := range items {
		urls[i] = it.Link
	}

	existing, err := p.articleRepo.ExistingURLs(ctx, urls)
	if err != nil {
		slog.Warn("dedup query failed, processing all", "err", err)
		return items // fail-open
	}

	out := make([]FeedItem, 0, len(items))
	for _, it := range items {
		if !existing[it.Link] {
			out = append(out, it)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
