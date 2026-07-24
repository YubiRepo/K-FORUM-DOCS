# Plan + Report: Translasi Master Data News (Kategori & Scope)

> **Dokumen hidup** — berdiri sendiri, terpisah dari
> [`PLAN_I18N_L10N_PLATFORM.md`](./PLAN_I18N_L10N_PLATFORM.md) yang
> platform-wide (sengaja ditangguhkan). Ini scope kecil, konkret: i18n/l10n
> untuk 2 tabel master data News saja. Latar belakang opsi desain
> translasi ada di [`BRAINSTORM_I18N_L10N_PLATFORM.md`](./BRAINSTORM_I18N_L10N_PLATFORM.md)
> (opsi (a)/(b)/(c)/(d)) — dokumen ini memakai pendekatan berbeda dari
> rekomendasi brainstorm (JSONB array, bukan tabel relasional) karena
> konteksnya beda: master data kecil dikelola admin, bukan konten dinamis
> volume tinggi seperti artikel/event.

## Context

`news_categories` & `news_scopes` (tabel master data yang jadi acuan
kategorisasi artikel News) hari ini cuma punya kolom `name` single-locale.
`article_translations` (translasi konten artikel) sudah matang dan
**tidak disentuh** — scope ini murni kategori & scope.

Keputusan desain:
- Translasi disimpan sebagai **array struct, di-persist sebagai JSONB**
  langsung di tabel `news_categories`/`news_scopes` — bukan tabel
  translasi terpisah seperti `article_translations`, karena volume kecil
  dan dikelola admin (bukan user-authored content).
- **Tidak ada auto-translate/LLM** untuk ini (beda dari artikel) — murni
  entry manual oleh admin via backoffice.
- GET/list endpoint publik (mobile) disesuaikan supaya nama yang
  ditampilkan sesuai locale request, dengan fallback ke nama default kalau
  translasi belum ada.

## Status

**DONE** — 2026-07-24 (termasuk follow-up: `category_name`/`scope_name` yang
ter-embed di response artikel kini juga locale-aware). 1 item non-blocking
sengaja ditunda, lihat *Catatan & Keputusan yang Diambil* di bawah.

## Checklist Implementasi

### Database
- [x] Migration `20260724071914_add_news_category_scope_translations`:
      `ALTER TABLE news_categories/news_scopes ADD COLUMN translations JSONB NOT NULL DEFAULT '[]'::jsonb`

### Domain (`internal/domain/news/`)
- [x] Value object `valueobject/name_translation.go`: `NameTranslation{Language, Name string}` + `ValidateNameTranslations`/`ResolveName` helper
- [x] `entity/news_category.go` & `entity/news_scope.go`: field `Translations []valueobject.NameTranslation`
- [x] Domain method `ResolveName(locale string) string` (fallback ke `Name` default)
- [x] Domain method `SetTranslations(...)` + tolak duplicate `Language`
- [x] Domain error code baru (`CodeNewsMasterDataTranslation{LanguageRequired,NameRequired,DuplicateLanguage}`) di `constant/news_constant.go`

### Persistence
- [x] `postgres_news_repository.go`: marshal/unmarshal `Translations` ↔ JSONB di Save/Update/FindAll/FindByID/FindBySlug (Category & Scope)

### Usecase (`internal/app/usecase/news/`)
- [x] `create_category.go`/`update_category.go`, `create_scope.go`/`update_scope.go`: terima & validasi `translations`
- [x] `list_scopes_mobile.go` + `list_categories_mobile.go` (baru, dipecah dari `list_categories.go` yang tadinya dipakai bersama web&mobile — lihat catatan di bawah): locale-aware via `ResolveName`, shape DTO tidak berubah
- [x] `list_categories.go`, `list_scopes_web.go` (admin): kembalikan `translations` lengkap utk edit

### Handler & DTO
- [x] `internal/app/dto/news_dto.go`: `NameTranslationItem` (dipakai dobel sebagai input & output), `translations` field di `NewsCategoryItem`/`NewsScopeItem` (`omitempty`)
- [x] Mobile handler: baca locale via `?lang=` fallback `c.GetString("locale")` (categories & scopes)
- [x] **Follow-up 2026-07-24 (selesai)**: response artikel yang meng-embed `category_name`/`scope_name` denormalized via JOIN kini ikut `ResolveName` — `list_articles_web.go` (+`?lang=`, baru), `list_articles_mobile.go`, `get_article_web.go` (+`?lang=` param baru di route/usecase/interface), `get_article_mobile.go`, `list_bookmarks.go` (cuma category, scope tak pernah ada di sana). Helper bersama `resolveNewsMasterDataName()` di `postgres_news_query.go` (reuse `valueobject.ResolveName`, infra→domain import diizinkan). `port.NewsArticleQueryReadModel.GetDetailForWeb` dapat parameter `lang` baru (breaking-internal, semua call site sudah diupdate). `PostgresNewsScopeQuery.List`/`port.NewsScopeQueryReadModel` dicek — ternyata dead code (dideklarasikan di `Dependencies` tapi tak pernah dipakai `NewUseCases`), sengaja tidak disentuh.

### Testing
- [x] Handler test: validation error (duplicate language) utk Category & Scope, locale-aware GET happy path (mobile), admin GET returns full translations (web) — lihat `TestWebNews_Create{Category,Scope}_{WithTranslations_Success,DuplicateTranslationLanguage}`, `TestWebNews_UpdateCategory_Translations_Success`, `TestMobileNews_List{Categories,Scopes}_LocaleAware`
- [x] Test tambahan utk `category_name` ter-embed di artikel: `TestWebNews_{ListArticles,GetArticle}_CategoryNameLocaleAware`, `TestMobileNews_{ListArticles,GetArticle,ListBookmarks}_CategoryNameLocaleAware` (termasuk kasus fallback locale tak didukung)
- [x] `go build ./...` lulus; full suite `go test ./internal/interfaces/http/handler/{mobile,web}/...` lulus tanpa regresi
- [x] **Bonus fix**: 2 gap test-infra pre-existing ditemukan & diperbaiki saat menulis test di atas (`internal/testhelper/testserver.go`) — `ScopeRepo` tidak pernah di-construct/wired ke `news.Dependencies` (bikin SEMUA usecase scope panic nil-pointer kalau benar-benar dipanggil), dan route web `/api/v1/web/news/scopes` (GET/POST/PUT/DELETE) tidak pernah didaftarkan di `buildRouter` test harness (404) — keduanya laten sejak fitur Scope pertama dibuat, baru ketahuan sekarang karena sebelumnya tidak ada test yang benar-benar memanggil usecase Scope lewat HTTP.

### Backoffice
- [x] `app/types/news.ts`: `NameTranslation` interface + `translations?: NameTranslation[]` di `NewsCategory`/`NewsScope`
- [x] `app/stores/newsStore.ts`: passthrough `translations` di `createCategory`/`createScope` (tipe payload diperluas); `updateCategory`/`updateScope` otomatis ikut karena pakai `Partial<NewsCategory/NewsScope>`
- [x] Komponen baru `app/components/news/TranslationsFieldset.vue` (`<NewsTranslationsFieldset>`, dipakai kedua halaman) — list input per bahasa aktif, tanpa status badge/tombol Auto-Translate, sesuai "murni manual"
- [x] Wire ke `app/pages/news/categories.vue` & `app/pages/news/scopes.vue` (dalam `UAccordion` collapsible di modal create/edit, supaya modal tidak terasa penuh utk kasus tanpa translasi)
- [x] Filter bahasa: `is_active && is_translate_target` dari `useLanguageStore` — lihat catatan soal locale default di bawah

## Catatan & Keputusan yang Diambil

1. **`ListCategoriesUseCase` ternyata dipakai bersama web (admin) & mobile
   (public)** — beda dari Scope yang sudah punya `ListNewsScopesWebUseCase`/
   `ListNewsScopesMobileUseCase` terpisah. Untuk konsistensi & supaya mobile
   tidak ikut membocorkan `translations` mentah, dibuat
   `ListCategoriesMobileUseCase` baru (mirror pola Scope), `ListCategoriesUseCase`
   lama tetap jadi versi admin/web. Route/path tidak berubah, cuma usecase
   internal yang dipanggil handler mobile.
2. **UI tidak mengecualikan locale default platform dari daftar input
   translasi** — tidak ada sumber "default_language" yang gampang diakses di
   backoffice tanpa nambah plumbing baru (`SystemLanguage` tidak punya flag
   `is_default`, general settings pakai form generik per-group). Semua bahasa
   `is_active && is_translate_target` ditampilkan; kalau salah satunya
   kebetulan sama dengan bahasa default, admin cukup biarkan kosong (input
   kosong tidak ikut terkirim). Redundansi UI minor, bukan bug — bisa
   disempurnakan nanti kalau benar-benar mengganggu.
3. **`article_translations` & pipeline LLM News sama sekali tidak disentuh**
   — sesuai scope yang diminta.

## Verifikasi

1. Migration jalan bersih; data existing tetap punya `name`, `translations` default `[]`.
2. Create/update kategori dengan translations via backoffice → tersimpan benar; duplicate `language` → ditolak.
3. Mobile `GET /categories?lang=en` → nama resolve ke translasi `en` kalau ada, fallback ke `name` default kalau tidak.
4. Web admin `GET /categories` → response berisi `translations` array lengkap.
5. Backoffice: modal edit kategori yang sudah punya translasi → input per bahasa terisi otomatis; simpan → ter-update benar.
6. `go build ./...` lulus; handler test kategori/scope existing tidak regresi; test baru sesuai skenario CLAUDE.md §11.

## Critical Files

**API:**
`internal/migrations/{0012_create_news_tables,0024_add_news_scopes}.up.sql` (referensi),
`internal/domain/news/entity/{news_category,news_scope}.go`,
`internal/domain/news/valueobject/name_translation.go` (baru),
`internal/domain/news/repository/interfaces.go`,
`internal/infrastructure/persistence/postgres_news_repository.go`,
`internal/app/usecase/news/{create_category,update_category,create_scope,update_scope,list_categories,list_scopes_web,list_scopes_mobile}.go`,
`internal/app/dto/news_dto.go`,
`internal/interfaces/http/handler/{web,mobile}/news_handler.go`

**Backoffice:**
`app/pages/news/categories.vue`, `app/pages/news/scopes.vue`,
`app/components/news/TranslationsFieldset.vue` (baru),
`app/types/news.ts`, `app/stores/newsStore.ts`
