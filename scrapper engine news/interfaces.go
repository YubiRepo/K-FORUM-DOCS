package scraper

import (
	"context"
	"time"
)

// SourceRepo menyediakan akses ke konfigurasi source di database.
type SourceRepo interface {
	// ListActive mengembalikan semua source aktif beserta selectors & categories.
	ListActive(ctx context.Context) ([]Source, error)
	// UpdateLastScraped mencatat waktu scraping terakhir untuk satu source.
	UpdateLastScraped(ctx context.Context, sourceID string, t time.Time) error
}

// ArticleRepo menyediakan operasi penyimpanan artikel.
type ArticleRepo interface {
	// ExistingURLs mengembalikan set URL yang sudah ada di DB (untuk dedup).
	ExistingURLs(ctx context.Context, urls []string) (map[string]bool, error)
	// SaveScraped menyimpan satu artikel hasil scraping + translation original.
	// Mengembalikan article ID. Idempotent terhadap original_url (ON CONFLICT DO NOTHING).
	SaveScraped(ctx context.Context, in SaveScrapedInput) (string, error)
}

// Cleaner merapikan konten artikel via AI (opsional, per source).
type Cleaner interface {
	Clean(ctx context.Context, contentHTML string) (string, error)
}

// Translator meng-enqueue job translation untuk satu artikel.
type Translator interface {
	// EnqueueAll membuat job translation ke semua bahasa target aktif.
	EnqueueAll(ctx context.Context, articleID string) error
}
