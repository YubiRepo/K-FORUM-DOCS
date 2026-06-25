# Plan: Global Search System dengan Meilisearch

## Context

Saat ini setiap modul punya search sendiri menggunakan PostgreSQL `ILIKE` — tidak ada endpoint terpadu. Mobile butuh satu search bar yang mencari semua konten sekaligus. Tujuannya: endpoint `/api/v1/mobile/search` yang fan-out ke Meilisearch, mengembalikan hasil ternormalisasi grouped-by-type atau paginated per-type, dengan aturan visibilitas yang eksplisit di domain layer.

**Spec rujukan**: `K-FORUM-DOCS/API SPEC/Mobile/API_SPEC_SEARCH_MOBILE.md`

---

## Mengapa Ada Domain Layer untuk Search?

Aturan "konten apa yang boleh masuk index global search" adalah **bisnis rule** — bukan detail teknis. Contoh:
- Post hanya di-index kalau komunitas parent-nya `public` dan `active`
- Announcement hanya di-index kalau `published` dan belum expired

Rule ini harus **eksplisit dan testable**, bukan terpendam di SQL query di infrastruktur. Oleh karena itu:
- `internal/domain/search/` mendefinisikan kontrak dan value object
- Indexers tinggal di **app layer** (bukan infrastruktur) karena mereka mengandung orkestrasi bisnis rule
- Infrastruktur hanya implementasi storage Meilisearch

---

## Arsitektur Layer

```
interfaces/http → usecase/search → domain/search (SearchIndexRepository interface)
                                 → app/port (*SearchReadModel per domain)
                                 → app/port (SearchQueryPort)
                                       ↓
                           infrastructure/search (MeilisearchPort implements domain + port interfaces)
                           infrastructure/persistence (PostgresSearchQuery implements read models)
```

---

## Privacy Rules (eksplisit di Indexer, app layer)

| Type | Aturan |
|------|--------|
| `announcement` | `status == "published"` && belum expired |
| `news` | `status == "published"` |
| `community` | `status == "active"` (private boleh — hanya metadata) |
| `post` | `status == "published"` && community `status == "active"` && community `visibility == "public"` |
| `event` | `status == "published"` && `approval_status == "approved"` |
| `merchant` | `status == "published"` && `approval_status == "approved"` && tidak dibanned |
| `qna` | `status in ["published", "pinned"]` |

---

## File yang Dibuat / Dimodifikasi

### Domain layer (NEW)

```
internal/domain/search/
├── document.go
│   └── SearchDocument struct (value object — apa yang disimpan di index)
│       Fields: ID, Type, Title, TitleID, TitleEN, Body, BodyID, BodyEN,
│               Subtitle, ThumbnailURL, Badge, Deeplink, CreatedAt int64,
│               RegionID string, IsCritical bool, IsFeatured bool
└── repository.go
    └── SearchIndexRepository interface
        - Index(ctx, docType string, doc SearchDocument) error
        - BulkIndex(ctx, docType string, docs []SearchDocument) error
        - Delete(ctx, docType, id string) error
        - SetupIndices(ctx) error
```

### App layer — Port (NEW)

**`internal/app/port/search_read_model.go`** — interface per tipe, berisi field minimal untuk indexing:

```go
type AnnouncementSearchReadModel interface {
    FindSearchable(ctx context.Context) ([]AnnouncementSearchItem, error)
    FindSearchableByID(ctx context.Context, id string) (*AnnouncementSearchItem, error)
}
// ... NewsSearchReadModel, CommunitySearchReadModel, PostSearchReadModel,
//     EventSearchReadModel, MerchantSearchReadModel, QnASearchReadModel

type AnnouncementSearchItem struct {
    ID, Title, Body, AnnouncementType, Scope string
    RegionID   *string
    ImageURL   *string
    ExpiresAt  *time.Time
    CreatedAt  time.Time
    IsCritical bool
}
// ... item struct per tipe (minimal fields, bukan full DTO)
```

**`internal/app/port/search_query_port.go`** — query ke search engine:

```go
type SearchQueryPort interface {
    SearchType(ctx context.Context, query SearchQuery) (*SearchTypeResult, error)
    Suggest(ctx context.Context, q string, limit int, locale string) ([]SearchHit, error)
}

type SearchQuery struct {
    Type     string // "announcement" | "news" | ...
    Q        string
    Limit    int
    Offset   int
    RegionID string
    Locale   string // "ko" | "id" | "en"
}

type SearchHit struct {
    ID, Type, Title string
    Subtitle, ThumbnailURL, Badge, Highlight *string
    Deeplink  string
    Score     float64
    CreatedAt int64
}

type SearchTypeResult struct {
    Type  string
    Total int64
    Hits  []SearchHit
}
```

### App layer — Indexers (usecase/search/indexer, BUKAN infrastruktur)

```
internal/app/usecase/search/indexer/
├── announcement_indexer.go
├── news_indexer.go
├── community_indexer.go
├── post_indexer.go
├── event_indexer.go
├── merchant_indexer.go
└── qna_indexer.go
```

Setiap indexer: `struct { readModel port.XxxSearchReadModel; indexRepo domain.SearchIndexRepository }`

Methods:
- `BulkIndex(ctx)` — fetch semua, apply domain rule, map ke SearchDocument, push ke repo
- `IndexOne(ctx, id)` — untuk event-driven: fetch by ID, validasi rule, index atau skip (jika tidak lagi eligible)
- `DeleteOne(ctx, id)` — hapus dari index

**Contoh pola di `post_indexer.go`:**
```go
func (idx *PostIndexer) BulkIndex(ctx context.Context) error {
    items, _ := idx.readModel.FindSearchable(ctx)
    var docs []search.SearchDocument
    for _, item := range items {
        // Rule eksplisit dan testable — bukan SQL tersembunyi:
        if item.CommunityVisibility != "public" || item.CommunityStatus != "active" {
            continue
        }
        docs = append(docs, mapPostToDocument(item))
    }
    return idx.indexRepo.BulkIndex(ctx, "post", docs)
}
```

### App layer — DTOs

**`internal/app/dto/search_dto.go`:**
```go
type ListSearchQuery struct {
    Q        string  `form:"q" binding:"required,min=2"`
    Type     string  `form:"type"`         // kosong = overview, isi = per-type
    Types    string  `form:"types"`        // filter tipe di overview, comma-separated
    PerType  *int    `form:"per_type"`     // default 5, max 10
    Limit    *int    `form:"limit"`        // per-type mode, default 20, max 50
    Offset   *int    `form:"offset"`
    RegionID *string `form:"region_id"`
}

type SearchResultItem struct {
    Type         string   `json:"type"`
    ID           string   `json:"id"`
    Title        string   `json:"title"`
    Subtitle     *string  `json:"subtitle"`
    ThumbnailURL *string  `json:"thumbnail_url"`
    Badge        *string  `json:"badge"`
    Deeplink     string   `json:"deeplink"`
    Highlight    *string  `json:"highlight"`
    Score        float64  `json:"score"`
    CreatedAt    string   `json:"created_at"` // ISO8601
}

type SearchGroup struct {
    Type    string             `json:"type"`
    Total   int64              `json:"total"`
    HasMore bool               `json:"has_more"`
    Items   []SearchResultItem `json:"items"`
}

type SuggestionItem struct {
    Type     string `json:"type"`
    ID       string `json:"id"`
    Title    string `json:"title"`
    Deeplink string `json:"deeplink"`
}
```

### App layer — Usecases

```
internal/app/usecase/search/
├── dependencies.go
│   - Dependencies struct berisi: SearchQueryPort, 7 Indexer, *sql.DB (untuk reindex)
│   - UseCases struct: Search, Suggest, ReindexAll, ReindexType
├── search_usecase.go
│   - Execute(ctx, userID string, query dto.ListSearchQuery) → overview atau per-type
│   - Overview: goroutine fan-out ke semua tipe, collect & group results
│   - Per-type: single index query + pagination
├── suggest_usecase.go
│   - Execute(ctx, q string, limit int) → []dto.SuggestionItem
└── reindex_usecase.go
    - ExecuteAll(ctx) → BulkIndex() semua 7 indexer
    - ExecuteType(ctx, typeName string) → BulkIndex() indexer tertentu
```

### Infrastructure — Meilisearch implementation (NEW)

```
internal/infrastructure/search/
├── meilisearch_client.go
│   - NewMeilisearchClient(host, apiKey string) *meilisearch.Client
│   - Konstan nama index: IndexAnnouncements = "k_forum_announcements", dst.
└── meilisearch_search_port.go
    - MeilisearchPort struct
    - Implements domain.SearchIndexRepository (Index, BulkIndex, Delete, SetupIndices)
    - Implements port.SearchQueryPort (SearchType, Suggest)
    - SetupIndices() configures per index:
        searchableAttributes: ["title","title_id","title_en","body","body_id","body_en","subtitle"]
        filterableAttributes: ["region_id","is_critical","is_featured"]
        sortableAttributes:   ["created_at"]
        rankingRules:         ["words","typo","proximity","attribute","sort","exactness"]
```

### Infrastructure — DB Read Models (implementasi port)

**`internal/infrastructure/persistence/postgres_search_query.go`:**
- `PostgresSearchQuery` struct implements semua 7 `*SearchReadModel` interfaces
- Query per tipe dioptimalkan: JOIN untuk ambil relasi sekaligus (tidak N+1)
- SQL WHERE **mencerminkan** rule domain (konsisten, tapi rule aslinya di indexer):

```sql
-- PostSearchReadModel.FindSearchable — ambil dengan konteks komunitas
SELECT p.id, p.content, p.status, p.created_at,
       c.id as community_id, c.name, c.visibility, c.status as community_status
FROM community_posts p
JOIN communities c ON c.id = p.community_id
WHERE p.status = 'published'
  AND c.status = 'active'
  AND c.visibility = 'public'  -- pre-filter di SQL untuk efisiensi, rule tetap di indexer
```

### Interfaces layer

**`internal/interfaces/http/handler/mobile/search_handler.go`:**
```go
type SearchHandler struct{ useCases *searchusecase.UseCases }

func NewSearchHandler(useCases *searchusecase.UseCases) *SearchHandler

func (h *SearchHandler) Search(c *gin.Context)
    // Bind ListSearchQuery → dispatch: ada "type" → per-type, tidak ada → overview
    // Ambil locale dari c.GetString("locale")

func (h *SearchHandler) Suggestions(c *gin.Context)
```

**`internal/interfaces/http/handler/mobile/search_handler_test.go`** (wajib sebelum selesai):
- 401 tanpa token
- 400 q < 2 char
- 422 type tidak dikenal
- 200 overview kosong (q valid tapi tidak ada hasil)
- 200 overview dengan hasil (mock Meilisearch atau seed + reindex)
- 200 per-type paginated
- 200 suggestions

**`internal/interfaces/mq/handler/search_sync_handler.go`:**

| Routing key | Aksi |
|-------------|------|
| `announcement.published` | `announcementIndexer.IndexOne(id)` |
| `announcement.archived` | `announcementIndexer.DeleteOne(id)` |
| `news.article.published` | `newsIndexer.IndexOne(id)` |
| `news.article.archived` / `.rejected` | `newsIndexer.DeleteOne(id)` |
| `community.post.created` | `postIndexer.IndexOne(id)` |
| `community.post.removed` | `postIndexer.DeleteOne(id)` |
| `forum.event.published` | `eventIndexer.IndexOne(id)` |
| `forum.event.cancelled` / `.archived` | `eventIndexer.DeleteOne(id)` |

Merchant & QnA: belum ada domain events → di-handle via scheduled full re-index (gunakan `robfig/cron` yang sudah ada di codebase, trigger harian).

Error indexing di-log, tidak gagalkan message processing (best-effort).

### Modifikasi File Existing

**`internal/config/config.go`:**
```go
type MeilisearchConfig struct {
    Host   string
    APIKey string
}
// Tambah ke Config struct + Load():
Meilisearch: MeilisearchConfig{
    Host:   getEnv("MEILISEARCH_HOST", "http://localhost:7700"),
    APIKey: getEnv("MEILISEARCH_API_KEY", ""),
},
```

**`internal/interfaces/http/router/router.go`:**
- Tambah `*mobilehandler.SearchHandler` ke parameter `Setup()`
- Di `protectedMobile` group:
  ```go
  mobileSearch := protectedMobile.Group("/search")
  {
      mobileSearch.GET("", mobileSearchHandler.Search)
      mobileSearch.GET("/suggestions", mobileSearchHandler.Suggestions)
  }
  ```
- Di `protectedWeb` group (admin):
  ```go
  webSearch := protectedWeb.Group("/search", middleware.RequireAdmin())
  {
      webSearch.POST("/reindex", webSearchHandler.ReindexAll)
      webSearch.POST("/reindex/:type", webSearchHandler.ReindexType)
  }
  ```

**`cmd/app/main.go`:**
```go
// Inisialisasi Meilisearch
meiliClient := searchinfra.NewMeilisearchClient(cfg.Meilisearch.Host, cfg.Meilisearch.APIKey)
meiliPort := searchinfra.NewMeilisearchPort(meiliClient)
if err := meiliPort.SetupIndices(ctx); err != nil {
    logger.Warn("meilisearch setup failed", slog.Any("error", err))
}

// Search read models (di postgres_search_query.go)
searchQuery := persistence.NewPostgresSearchQuery(db)

// Indexers
annIndexer := searchindexer.NewAnnouncementIndexer(searchQuery, meiliPort)
newsIndexer := searchindexer.NewNewsIndexer(searchQuery, meiliPort)
// ... 5 lagi

// Usecases + handler
searchUseCases := searchusecase.NewUseCases(searchusecase.Dependencies{ ... })
mobileSearchHandler := mobilehandler.NewSearchHandler(searchUseCases)
```

---

## Dependency yang Sudah Ada (tidak perlu install baru)

- `github.com/meilisearch/meilisearch-go v0.36.3` — **sudah ditambahkan ke go.mod**
- `robfig/cron/v3` — sudah ada, untuk scheduled re-index
- `rabbitmq/amqp091-go` — sudah ada, untuk MQ sync handler

---

## Urutan Implementasi

1. `internal/config/config.go` — tambah MeilisearchConfig
2. `internal/domain/search/document.go` + `repository.go`
3. `internal/app/port/search_read_model.go` + `search_query_port.go`
4. `internal/app/dto/search_dto.go`
5. `internal/infrastructure/search/meilisearch_client.go` + `meilisearch_search_port.go`
6. `internal/infrastructure/persistence/postgres_search_query.go` (7 read models)
7. 7 indexers di `internal/app/usecase/search/indexer/`
8. `internal/app/usecase/search/` (search, suggest, reindex usecases)
9. `internal/interfaces/http/handler/mobile/search_handler.go` + test
10. Wire di `router.go` + `cmd/app/main.go`
11. `internal/interfaces/mq/handler/search_sync_handler.go`

---

## Verifikasi End-to-End

```bash
# 1. Build
go build ./...

# 2. Jalankan Meilisearch
docker run -p 7700:7700 getmeili/meilisearch

# 3. Env
export MEILISEARCH_HOST=http://localhost:7700
export MEILISEARCH_API_KEY=

# 4. Run server → cek log "meilisearch indices ready"

# 5. Trigger re-index
curl -X POST /api/v1/web/search/reindex -H "Authorization: Bearer <admin_token>"

# 6. Test endpoints
GET /api/v1/mobile/search?q=korea                          → grouped response
GET /api/v1/mobile/search?q=korea&type=news&limit=10       → paginated
GET /api/v1/mobile/search/suggestions?q=kor&limit=8        → suggestions

# 7. Test validasi
GET /api/v1/mobile/search?q=k    → 400 ERR_VALIDATION
GET /api/v1/mobile/search?q=test&type=unknown → 422 ERR_UNPROCESSABLE_ENTITY
GET /api/v1/mobile/search?q=test (tanpa token) → 401

# 8. Handler tests
go test ./internal/interfaces/http/handler/mobile/ -run TestSearch -v -timeout 120s
```
