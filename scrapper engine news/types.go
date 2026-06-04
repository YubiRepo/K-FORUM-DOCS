package scraper

import "time"

// Source merepresentasikan satu news_source beserta konfigurasinya
// (selectors + categories), dimuat dari database.
type Source struct {
	ID               string
	Key              string
	Name             string
	BaseURL          string
	OriginalLanguage string
	Schedule         string // cron expression, mis. "0 */6 * * *"
	LastScrapedAt    *time.Time
	AutoPublish      bool
	AICleanup        bool
	AutoTranslate    bool
	IsActive         bool

	Selectors  Selectors
	Categories []Category
}

// Selectors menyimpan CSS selector fallback + author/tags + config tambahan.
// Konten utama diekstrak via Readability; Content hanya fallback.
type Selectors struct {
	Content     string
	Author      string
	Tags        string
	ExtraFields map[string]interface{}
}

// Category mendefinisikan satu kategori yang di-scrape dari sebuah source.
type Category struct {
	ID           string
	CategoryKey  string
	URLSuffix    string
	URLOverride  string
	ArticleLimit int
	IsActive     bool
}

// FeedItem adalah satu item hasil parsing RSS feed (sebelum fetch detail).
type FeedItem struct {
	Link        string
	Title       string
	PubDate     *time.Time
	Description string // excerpt dari RSS
	Thumbnail   string
}

// ArticleDetail adalah hasil fetch + ekstraksi halaman detail artikel.
type ArticleDetail struct {
	Content     string   // plain text — untuk translation & full-text search
	ContentHTML string   // HTML — untuk render di client
	Excerpt     string   // ringkasan
	Author      string   // nama penulis
	Tags        []string // tag artikel
	Err         error    // diisi jika ekstraksi gagal
}

// SaveScrapedInput adalah payload untuk menyimpan artikel hasil scraping.
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
	Thumbnail        string
	Tags             []string
	PubDate          *time.Time
	AutoPublish      bool
	AuthorLabel      string
}
