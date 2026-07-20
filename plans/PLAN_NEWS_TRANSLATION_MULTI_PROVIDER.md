# Plan — News Translation: Multi-Provider (Google Cloud Translation, AWS Translate, LLM) + Pilihan Provider di Backoffice

> **Status (2026-07-18): Provider `llm` (Claude) SUDAH DIIMPLEMENTASIKAN**, memakai **Anthropic Message Batches API** (bukan panggilan sinkron per-item) — lihat bagian [§3 LLM (Claude via Anthropic Go SDK)](#3-llm-claude-via-anthropic-go-sdk) untuk desain final yang benar-benar dikerjakan, sedikit berbeda dari draft awal di bawah (menambah dukungan async batch submit+poll yang tidak ada di draft pertama). Google Cloud Translation dan AWS Translate **belum** dikerjakan.
>
> **Update (2026-07-18): §A/§B/§C (backoffice provider selection + registry + request-level override) SUDAH DIIMPLEMENTASIKAN JUGA** — desain final sedikit disederhanakan dari draft (`translation_llm_model` di §A **tidak** dibuat, lihat alasan di §A yang sudah direvisi). Lihat bagian masing-masing yang ditandai ✅ untuk detail final vs draft awal.

## Context

Fitur AI translate untuk News saat ini **belum benar-benar terhubung ke provider apa pun** — hanya `NoopTranslationProvider` (mengembalikan konten mentah tanpa terjemahan), disengaja sejak desain awal supaya build/test tidak butuh API key (`PLAN_NEWS_SCRAPER_ENGINE.md`). Provider ini di-wire satu kali secara global saat worker start:

```go
// cmd/worker/main.go:204
mqrelay.NewTranslationRelay(translationBatchRepo, &port.NoopTranslationProvider{}, 30*time.Second, 10, logger)
```

Ada **field yang sudah ada di API tapi tidak pernah dipakai**: `TriggerNewsTranslationRequest.Provider *string` (`internal/app/dto/news_dto.go:42-45`) — backoffice sudah mengirim field ini (`newsStore.ts` → `translateArticle(id, languages, provider)`), tapi `TriggerTranslationUseCase.Execute` mengabaikannya sepenuhnya, dengan komentar eksplisit di kode (`trigger_translation.go:36-40`) yang menjelaskan kenapa: relay hanya punya **satu provider global**, jadi pemilihan provider per-request tidak mungkin berfungsi tanpa dispatch multi-provider di level relay. Dokumen ini adalah plan untuk **mengisi gap itu**.

`NEWS_RULES.md:490-497` sudah menyebut niat "Google Translate sebagai default, OpenAI sebagai alternatif" — dokumen ini menggantikan asumsi tersebut dengan implementasi konkret untuk **3 provider**: Google Cloud Translation, AWS Translate, dan LLM (Claude), plus **pilihan provider dari backoffice** sesuai permintaan user.

**Pelajaran yang wajib dihindari** (dari `PLAN_SYSTEM_SETTINGS_IMPLEMENTATION.md`): audit system settings menemukan banyak key yang *"bisa diubah admin tanpa efek apa pun"* karena backend tidak pernah membaca nilainya — hardcode tetap dipakai. Setting provider translation di plan ini **wajib** benar-benar dikonsumsi end-to-end sebelum dianggap selesai; tidak boleh berakhir sebagai dropdown kosmetik.

---

## Tujuan

1. Implementasi provider translation nyata: **Google Cloud Translation (Advanced v3)**, **AWS Translate**, **LLM (Claude)** — masing-masing mengimplementasikan `port.NewsTranslationProvider` yang sudah ada.
2. **Registry/dispatch multi-provider** di worker — bukan lagi satu instance global tunggal — supaya lebih dari satu provider bisa hidup berdampingan dan dipilih per-job.
3. **Backoffice bisa memilih provider aktif** (default + fallback) lewat halaman News Settings yang sudah ada.
4. **Routing yang benar di sistem**: field `Provider` yang sudah ada di request trigger — sekarang benar-benar dihormati, disimpan di job row, dan dibaca ulang oleh relay saat eksekusi async (bukan lagi diabaikan).
5. Provider yang belum dikonfigurasi (API key kosong) **tidak boleh diam-diam sukses sebagai Noop** — job harus gagal dengan alasan jelas, bisa dilihat admin.
6. Integritas HTML terjaga di title/content/summary untuk ketiga provider baru (konten News berupa HTML, lihat Non-Goals di `PLAN_NEWS_SCRAPER_ENGINE.md`).
7. Observability: provider mana yang menangani tiap job (`translated_by` sudah ada), status gagal per provider, dan tombol "Test Connection" per provider di backoffice.

### Non-Goals (v1)

- Pilihan provider **per-bahasa** (mis. Google utk id/en, Claude utk ko) — v1 hanya global default + fallback. Ditandai sebagai kandidat v2 di Open Questions.
- API key yang bisa diubah runtime dari backoffice (terenkripsi) — v1 pakai env var, ops-managed.
- Dashboard biaya/rate-limit per provider — flagged sebagai follow-up, bukan blocker rilis.
- DeepL — sempat direkomendasikan di diskusi sebelumnya tapi tidak diminta di sini; pola registry di plan ini membuatnya trivial ditambah nanti sebagai provider ke-4.

---

## Keputusan Arsitektur

### A. Di mana "provider mana yang aktif" disimpan? — ✅ DIIMPLEMENTASIKAN 2026-07-18

Ada dua mekanisme settings yang hidup berdampingan di codebase (lihat riset arsitektur):

| Mekanisme | Lokasi | Karakteristik |
|---|---|---|
| `system_settings` (generic key-value) | `internal/migrations/0017_create_system_settings_tables.up.sql` | Row-per-key, punya `is_sensitive` (encrypted+masked) — cocok untuk secret, tapi belum ada `group_key='news'` |
| `news_system_settings` (singleton per-modul) | `internal/migrations/0012_create_news_tables.up.sql:32-45` | Sudah menyimpan `translation_enabled`/`on_demand_enabled`, sudah dibaca langsung oleh `TriggerTranslationUseCase` |

**Keputusan: extend `news_system_settings`** — provider translation secara konsep sama persis dengan `translation_enabled` (knob runtime domain News), dan menghindari config News terpecah di dua sistem berbeda. Kolom baru:

```sql
translation_provider                  VARCHAR(20) NOT NULL DEFAULT 'noop'   -- 'noop' | 'google' | 'aws' | 'llm'
translation_fallback_provider         VARCHAR(20) NULL                      -- opsional, salah satu enum di atas
translation_allow_request_override    BOOLEAN NOT NULL DEFAULT true         -- izinkan field `Provider` di request trigger override default
translation_llm_model                 VARCHAR(50) NULL                      -- mis. 'claude-haiku-4-5', hanya relevan jika provider=llm
```

API key **tetap di env var** (ops-managed), bukan di tabel ini — menghindari perlu duplikasi plumbing enkripsi `system_settings.is_sensitive` untuk v1. Ditandai eksplisit sebagai open question kalau nanti mau key yang bisa diputar dari backoffice (baru pindah ke `system_settings` group `news`).

**Realisasi final (beda dari draft di atas):** hanya 3 kolom yang dibuat, BUKAN 4 — `translation_llm_model` **sengaja tidak dibuat** untuk hindari "setting kosmetik yang tidak berpengaruh" (lihan lesson-learned di Context di atas): model LLM tetap murni env var (`ANTHROPIC_TRANSLATION_MODEL`), karena `LLMTranslationProvider` dibuat sekali saat startup worker dengan model tetap — expose field ini di backoffice tanpa benar-benar mengubah model yang dipakai berjalan akan persis jadi anti-pattern yang mau dihindari. Migration asli: `20260718081110_add_news_translation_provider_settings` — menambah `translation_provider VARCHAR(20) DEFAULT 'llm'`, `translation_fallback_provider VARCHAR(20) DEFAULT 'noop'`, `translation_allow_request_override BOOLEAN DEFAULT true` ke `news_system_settings`. **Default `'llm'`/`'noop'` (bukan `'noop'`/NULL seperti draft)** — dipilih supaya migration TIDAK mengubah behavior yang sudah jalan sejak sesi sebelumnya (worker yang sudah pakai `llm` via env var lanjut jalan sama persis; environment tanpa API key otomatis fallback ke `noop` — identik dengan behavior lama).

Domain: `constant.TranslationProvider` (`noop`/`google`/`aws`/`llm`) + `IsValid()` di `internal/domain/news/constant/news_constant.go`. Entity `NewsSystemSettings` (`internal/domain/news/entity/news_system_settings.go`) dan repo Get/Update (`postgres_news_repository.go`) sudah baca-tulis 3 kolom ini. API: `GetSettingsUseCase`/`UpdateSettingsUseCase` extend + validasi enum (422 `DOMAIN_NEWS_INVALID_TRANSLATION_PROVIDER` kalau bukan salah satu dari 4 nilai). Backoffice: section baru "Translation Provider" di `k-forum-backoffice/app/pages/news/settings.vue` (select default+fallback provider, toggle override, badge availability).

### B. Registry/dispatch provider (menghilangkan batasan single-global-instance) — ✅ DIIMPLEMENTASIKAN 2026-07-18

Package baru `internal/infrastructure/translation/`:

```
internal/infrastructure/translation/
├── google_translate_provider.go   // GoogleTranslateProvider, Name() = "google"
├── aws_translate_provider.go      // AWSTranslateProvider,   Name() = "aws"
├── llm_translation_provider.go    // LLMTranslationProvider, Name() = "llm" (Claude)
└── registry.go                    // ProviderRegistry
```

```go
// registry.go
type ProviderRegistry struct {
    providers map[string]port.NewsTranslationProvider
}

// NewProviderRegistry membangun registry sekali saat worker start.
// Provider HANYA didaftarkan kalau env var wajibnya lengkap — provider
// yang belum dikonfigurasi tidak bikin worker crash, cukup absen dari map.
func NewProviderRegistry(cfg Config) *ProviderRegistry

func (r *ProviderRegistry) Get(name string) (port.NewsTranslationProvider, bool)
func (r *ProviderRegistry) Available() []string // untuk endpoint status
```

`TranslationRelay` (`internal/interfaces/mq/relay/translation_relay.go`) berubah dari menyimpan satu `provider port.NewsTranslationProvider` menjadi menyimpan `registry *ProviderRegistry` + `settingsRepo` (untuk baca `translation_provider`/`translation_fallback_provider` terkini — bukan snapshot beku saat startup, supaya perubahan setting di backoffice langsung berlaku tanpa restart worker).

**Aturan resolusi provider per job** (dipanggil sebelum `provider.Translate(...)`):

1. `job.ProviderRequested` (kolom baru, lihat bagian C) — kalau ada **dan** tersedia di registry → pakai ini.
2. Kalau tidak → `news_system_settings.translation_provider` — kalau tersedia di registry → pakai ini.
3. Kalau tidak tersedia → `news_system_settings.translation_fallback_provider` (kalau di-set) — kalau tersedia di registry → pakai ini.
4. Kalau masih tidak ada yang tersedia → **job di-mark `failed`** dengan `provider_error = "no_provider_configured"`. **Tidak pernah** diam-diam jatuh ke Noop sebagai fallback tersembunyi — Noop hanya dipakai kalau memang itu yang secara eksplisit dipilih di langkah 1-3 (mis. admin sengaja set default=`noop` di environment dev/staging).

Aturan #4 ini secara langsung menghindari kelas bug yang sudah pernah terjadi (`project_news_translation_lifecycle_bugs.md`: `translated_by` kosong, dispatch gagal total) — status job harus selalu jujur mencerminkan apakah terjemahan sungguhan terjadi.

**Realisasi final:** persis seperti draft, dengan 2 detail tambahan yang draft belum antisipasi:
- Interface bernama `port.ProviderRegistry` (bukan `translation.ProviderRegistry` struct langsung) — implementasinya `translation.Registry` (`registry.go`), constructor `NewRegistry(providers ...port.NewsTranslationProvider)`. `Get`/`Available()` sama seperti draft.
- **Bucketing per-provider dalam satu tick** — hal yang draft belum bahas: satu batch `ClaimPending` bisa berisi campuran job dengan `ProviderRequested` berbeda-beda (mis. satu job override ke `llm`, sisanya pakai default). `TranslationRelay.processBatch` sekarang resolve provider PER JOB dulu, kelompokkan (`bucket`) per nama provider, baru dispatch tiap bucket (submit-batch kalau providernya implement `NewsBatchTranslationProvider`, sekuensial kalau tidak) — supaya per-request override benar-benar bisa jalan bersamaan dengan default provider di tick yang sama tanpa saling mengganggu.
- Karena `PollBatch` untuk batch Anthropic butuh tahu provider MANA yang submit (batch ID itu opaque per-provider), ditambah kolom `article_translations.batch_provider` (bukan bagian draft awal) yang direkam saat `MarkBatchSubmitted`.

### C. Menyimpan provider yang diminta saat trigger — ✅ DIIMPLEMENTASIKAN 2026-07-18

`article_translations` dapat kolom baru:

```sql
provider_requested  VARCHAR(20)   NULL   -- provider yg diminta caller (dari request trigger), null = pakai default settings
```

**Realisasi final:** `provider_error` **tidak dibuat sebagai kolom baru** — kolom `translate_error` yang sudah ada dari migration batch-tracking sesi sebelumnya (`20260718033455_add_news_translation_batch_tracking`) dipakai ulang untuk alasan gagal (termasuk `"no_provider_configured"`), jadi cukup satu kolom saja, bukan dua terpisah seperti draft. `ProviderRequested` juga ditambahkan ke domain entity `ArticleTranslation` (+ method `SetProviderRequested`), bukan cuma di level SQL — supaya usecase tetap ikut pola domain-first yang dipakai di seluruh codebase (lihat CLAUDE.md §"Pola Kerja Usecase").

`TriggerTranslationUseCase.Execute` (`internal/app/usecase/news/trigger_translation.go`) sekarang:

1. Validasi `req.Provider` (kalau diisi) adalah salah satu enum yang dikenal (`google`/`aws`/`llm`/`noop`) → 422 (`DOMAIN_NEWS_INVALID_TRANSLATION_PROVIDER`) kalau bukan.
2. **Open Question #1 SUDAH DIPUTUSKAN**: kalau `translation_allow_request_override = false`, field `Provider` **diabaikan secara diam-diam** (bukan 400/422) — request tetap sukses 202, job tetap ter-enqueue, hanya providernya nanti ikut default settings bukan permintaan caller. Alasan: field ini cuma metadata routing, bukan hal yang layak menggagalkan seluruh operasi trigger-translate kalau admin sengaja mengunci override. Validasi enum tetap jalan duluan (jadi provider yang memang bukan salah satu dari 4 nilai yang dikenal tetap 422, terlepas dari status override).
3. Simpan hasil validasi ke kolom `provider_requested` pada row `article_translations` yang dibuat/di-mark pending — inilah yang nanti dibaca relay.

Test coverage: `TestWebNews_TriggerTranslation_InvalidProvider_Unprocessable`, `TestWebNews_TriggerTranslation_ProviderOverrideDisabled_StillEnqueues`, `TestWebNews_UpdateSettings_TranslationProvider_Success`, `TestWebNews_UpdateSettings_InvalidTranslationProvider_Unprocessable`, `TestWebNews_GetSettings_IncludesAvailableProviders` di `news_handler_test.go` — semua lolos bersama seluruh test News existing (tidak ada regresi).

---

## Rencana Implementasi Provider

### 1. Google Cloud Translation (Advanced v3) — risiko implementasi paling rendah

- **SDK**: `cloud.google.com/go/translate/apiv3` (dependency baru — `google.golang.org/api` yang sudah ada di `go.mod` dipakai untuk FCM, bukan Translation, jadi tetap perlu tambah modul baru)
- **Auth**: service-account JSON via `GOOGLE_APPLICATION_CREDENTIALS` (path file) atau `GOOGLE_TRANSLATE_CREDENTIALS_JSON` (inline base64) — pilih satu pola, didokumentasikan di `.env.example`
- **HTML**: request `TranslateText` dengan `MimeType: "text/html"` → tag HTML otomatis dipertahankan, tidak perlu kerja tambahan (dikonfirmasi dari dokumentasi resmi Google Cloud)
- **Pricing** (referensi keputusan sebelumnya): $20/juta karakter, free tier 500K/bulan

### 2. AWS Translate — butuh spike sebelum implementasi penuh

- **SDK**: AWS SDK for Go v2 — `github.com/aws/aws-sdk-go-v2/service/translate` (dependency baru; project saat ini pakai **MinIO** untuk object storage, bukan AWS SDK, jadi tidak ada precedent AWS credential chain yang bisa dipakai ulang)
- **Auth**: `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_REGION` (atau IAM role kalau nanti deploy di infra AWS)
- **⚠️ Open engineering question — perlu spike Phase 0**: `TranslateText` dasar AWS **tidak** punya mode HTML-aware setara `mimeType` Google. Preservasi HTML di AWS dilakukan lewat `TranslateDocument` (menerima raw document bytes + `ContentType: text/html`, didesain untuk dokumen utuh, bukan snippet per-field) atau `StartTextTranslationJob` (batch async, tidak cocok untuk pola `Translate()` sinkron per-field yang ada sekarang). Spike Phase 0 harus memastikan `TranslateDocument` bisa dipanggil sinkron per potongan HTML kecil (title/content/summary) dengan hasil yang benar. Kalau tidak praktis, fallback: strip tag HTML jadi placeholder sebelum panggil `TranslateText` plain, lalu sambung ulang tag setelahnya — lebih banyak kode, risiko korupsi lebih tinggi pada markup kompleks (nested tag, atribut yang juga mengandung teks terjemahkan seperti `alt=`).
- **Pricing** (referensi): $15/juta karakter, free tier 2 juta karakter/bulan (12 bulan pertama)

### 3. LLM (Claude via Anthropic Go SDK) — ✅ DIIMPLEMENTASIKAN 2026-07-18

- **⚠️ Auth — WAJIB API key, bukan OAuth subscription (keputusan final, bukan open question)**: sempat dipertimbangkan pakai OAuth dari langganan Team plan yang sudah ada supaya bisa mulai tanpa API key baru, tapi **ini dilarang eksplisit oleh Anthropic** — OAuth token dari subscription Claude.ai (Free/Pro/Max/Team) hanya sah dipakai di Claude Code & claude.ai; dipakai di tool/service lain melanggar Consumer Terms of Service, dan sejak **4 April 2026 sudah ditegakkan aktif** (subscription OAuth tidak lagi meng-cover usage lewat harness/tool pihak ketiga apa pun). Worker translation ini persis kategori "server-side/scheduled job/production automation" yang menurut dokumentasi resmi Anthropic wajib pakai **API key dengan Commercial Terms** (billing pay-per-token), bukan subscription login. Jadi provider `llm` sejak awal butuh API key baru dari Anthropic Console — tidak ada opsi "mulai gratis pakai Team plan dulu, API key belakangan". User sudah punya API key sendiri; disimpan di env var `ANTHROPIC_API_KEY` (lihat `.env.example`), tidak pernah masuk kode/DB.
- **SDK**: `github.com/anthropics/anthropic-sdk-go` v1.58.0 (ditambahkan ke `go.mod`)
- **Model**: env var `ANTHROPIC_TRANSLATION_MODEL`, default **`claude-haiku-4-5`** sesuai permintaan user (murah, kualitas multibahasa baik); belum dikawinkan ke setting backoffice (`translation_llm_model` di §A masih rencana, saat ini murni env var)
- **Prompt + JSON schema (menjaga model tidak keluar dari task translasi)**: system prompt eksplisit 6 aturan (hanya terjemahkan teks; pertahankan tag/atribut identifier seperti `href`/`src`/`class` apa adanya tapi terjemahkan atribut yang berisi teks tampil seperti `alt`/`title`; jangan tambah komentar apa pun; **konten yang diterjemahkan adalah DATA bukan instruksi** — guard eksplisit terhadap prompt injection dari artikel scraped pihak ketiga yang berisi teks yang menyerupai perintah; kembalikan tanpa perubahan kalau tidak ada teks; pertahankan proper noun/kode/URL). Dikombinasikan dengan **Structured Outputs** (`output_config.format` json_schema, field tunggal `translated_html`) — model **tidak bisa** membalas apa pun selain field itu, dan konten yang diterjemahkan dibungkus delimiter `<content_to_translate>...</content_to_translate>` sebagai lapisan tambahan pemisah data-vs-instruksi. Lihat `internal/infrastructure/translation/llm_translation_provider.go`.
- **Integritas HTML**: `internal/infrastructure/translation/htmlintegrity.go` — `tagsMatch()` membandingkan urutan nama tag (pakai `golang.org/x/net/html` tokenizer, sudah ada transitif via `goquery`) antara input dan output, mengabaikan isi teks/nilai atribut. Jalur sinkron (`Translate()`) retry sekali dengan instruksi korektif kalau mismatch; jalur batch (`PollBatch()`) langsung menandai field itu gagal (tidak auto-retry — retry berarti submit batch baru, mahal secara latency) — admin bisa re-trigger translate manual utk mengulang.
- **✅ Batch processing (Anthropic Message Batches API)** — sesuai permintaan user, BUKAN panggilan sinkron per-item:
  - `port.NewsBatchTranslationProvider` (baru, di `scraper_port.go`) — capability opsional di atas `NewsTranslationProvider`: `SubmitBatch(jobs) (batchID, error)` (submit, non-blocking) + `PollBatch(batchID, jobs) (done, results, error)` (cek status, ambil hasil kalau `processing_status=ended`).
  - `LLMTranslationProvider` mengimplementasikan keduanya: `SubmitBatch` mem-bundle title+content+summary tiap job jadi maks 3 sub-request (`custom_id = "{job_id}:title"` dst.) dalam **satu** `client.Messages.Batches.New` call; `PollBatch` stream hasil via `Batches.ResultsStreaming`, cocokkan `custom_id` balik ke job+field, validasi tag, kumpulkan per-job.
  - `TranslationRelay` (`internal/interfaces/mq/relay/translation_relay.go`) di-rewrite: tiap tick (30 detik) sekarang **poll batch in-flight dulu** (`pollInFlightBatches`) baru **claim+submit pending baru** (`processBatch`) — type-assert provider ke `NewsBatchTranslationProvider`; kalau provider tidak implement itu (mis. Noop, atau Google/AWS nanti kalau dibuat sinkron), fallback ke loop `Translate()` sekuensial yang lama, tidak ada perubahan perilaku untuk provider lain.
  - Migration `20260718033455_add_news_translation_batch_tracking` — kolom baru `article_translations.llm_batch_id` (batch mana job ini disubmit, dipakai poll ulang bahkan lintas restart worker) dan `translate_error` (alasan gagal, utk observability — mengisi gap yang sebelumnya cuma set status `failed` tanpa alasan).
  - **⚠️ Trade-off latency yang perlu diketahui**: batch Anthropic biasanya selesai **dalam 1 jam, maksimum 24 jam** (bukan detik seperti panggilan sinkron) — trigger translate dari backoffice sekarang **tidak** langsung terlihat hasilnya, harus menunggu sampai tick relay berikutnya berhasil poll batch yang sudah `ended`. Trade-off ini disengaja sesuai permintaan eksplisit user ("pake batch processing") untuk mengejar diskon ~50% biaya dibanding panggilan sinkron per-item — cocok karena translasi News memang bukan pekerjaan yang butuh hasil real-time (relay sendiri sudah async/polling sejak awal).
  - **✅ Toggle env `ANTHROPIC_TRANSLATION_USE_BATCH`** (default `true`) — supaya latensi ~1 jam di atas tidak mengganggu saat testing lokal, `false` memaksa jalur sinkron per-field (near-instant, harga penuh). Mekanisme: `translation.DisableBatch(provider)` (`internal/infrastructure/translation/batch_toggle.go`) membungkus provider dengan struct yang meng-embed `port.NewsTranslationProvider` sebagai *interface* (bukan concrete type) — trik Go ini membuat method `SubmitBatch`/`PollBatch` milik `LLMTranslationProvider` tidak ikut ter-promote, sehingga type-assertion `TranslationRelay` ke `NewsBatchTranslationProvider` gagal dan relay otomatis jatuh ke loop `Translate()` sekuensial yang lama. Diuji di `batch_toggle_test.go`.
  - `MarkFailed` di `TranslationBatchRepo` berubah signature — sekarang wajib `(ctx, id, reason string)`, dipakai baik jalur sekuensial lama maupun jalur batch baru.
- **Testing**: `internal/infrastructure/translation/htmlintegrity_test.go` — unit test `tagsMatch` (8 skenario: identik, atribut beda bebas, multipage divider, tag hilang/reorder/nambah terdeteksi, teks polos) + round-trip `customID`/`splitCustomID`. Belum ada test dengan HTTP client di-mock untuk `SubmitBatch`/`PollBatch` (butuh key asli atau `option.WithBaseURL` ke mock server) — ditandai sebagai follow-up, bukan blocker karena logic inti (parsing, validasi tag, resolusi job) sudah diuji lewat helper murni di atas.

---

## Perubahan Backoffice

Extend **`app/pages/news/settings.vue`** (halaman bespoke Vue yang sudah ada, backed by `newsStore.ts`) — **bukan** `GroupForm.vue` generic, karena keputusan Arsitektur A menaruh config ini di `news_system_settings`, bukan `system_settings`.

Section baru "Translation Provider":

- **Select** — Default Provider (`noop`/`google`/`aws`/`llm`), tiap opsi diberi badge status (configured/not configured) dari endpoint status baru — supaya admin tidak memilih provider yang pasti gagal karena belum ada API key
- **Select** — Fallback Provider (opsional, enum sama)
- **Toggle** — Allow request-level override
- **Select** — LLM Model (hanya aktif kalau Default atau Fallback = `llm`)
- **Tombol "Test Connection"** per provider — memicu 1 kali terjemahan trivial ("Hello" id→en) secara sinkron dan melaporkan sukses/gagal + latency, mengikuti pola yang sudah ada di System Settings (`POST /email/test`)

Update `newsStore.ts` + `app/types/news.ts` untuk field-field baru ini.

---

## Perubahan API

| Endpoint | Perubahan |
|---|---|
| `GET/PUT /api/v1/web/news/settings` | Extend DTO dengan 4 field baru (`translation_provider`, `translation_fallback_provider`, `translation_allow_request_override`, `translation_llm_model`) |
| `GET /api/v1/web/news/settings/translation-providers/status` (baru) | `[{provider, configured: bool, reason?}]` — dibaca dari `ProviderRegistry.Available()` |
| `POST /api/v1/web/news/settings/translation-providers/{provider}/test` (baru) | Test-translate sinkron, `{success, latency_ms, error?}` |
| `POST /api/v1/web/news/articles/{id}/translate` (existing) | `TriggerNewsTranslationRequest.Provider` sekarang benar-benar dipakai (lihat Keputusan Arsitektur C) |

---

## Migration

```sql
-- 00XX_add_news_translation_provider_config.up.sql

ALTER TABLE news_system_settings
    ADD COLUMN translation_provider VARCHAR(20) NOT NULL DEFAULT 'noop'
        CHECK (translation_provider IN ('noop', 'google', 'aws', 'llm')),
    ADD COLUMN translation_fallback_provider VARCHAR(20)
        CHECK (translation_fallback_provider IS NULL OR translation_fallback_provider IN ('noop', 'google', 'aws', 'llm')),
    ADD COLUMN translation_allow_request_override BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN translation_llm_model VARCHAR(50);

ALTER TABLE article_translations
    ADD COLUMN provider_requested VARCHAR(20)
        CHECK (provider_requested IS NULL OR provider_requested IN ('noop', 'google', 'aws', 'llm')),
    ADD COLUMN provider_error VARCHAR(255);
```

CHECK constraint dipakai sebagai defense-in-depth (bukan cuma validasi di application layer) — mengingat riwayat bug lifecycle translation sebelumnya, status/data yang tidak konsisten di level DB sudah pernah jadi sumber masalah nyata.

---

## Fase Rollout

| Fase | Isi | Perilaku prod selama fase ini |
|---|---|---|
| **0 — Spike** | Konfirmasi pendekatan HTML AWS (`TranslateDocument` vs strip/splice) | Tidak ada perubahan |
| **1** | Implementasi 3 provider + registry, unit test. Belum di-wire ke relay/settings | Tidak ada perubahan (dead code di belakang registry) |
| **2** | Migration + rewrite dispatch relay (aturan resolusi provider) | Default tetap `noop` di DB → tidak ada perubahan perilaku |
| **3** | UI backoffice (select provider, status, test connection) | Admin bisa lihat & test provider, tapi belum switch default di prod |
| **4** | Aktifkan `google` sebagai default di staging, monitor `translated_by`/`provider_error` pada sample artikel, verifikasi rendering HTML tidak rusak di web/app/backoffice preview | Staging only |
| **5** | Broaden ke prod, update `NEWS_RULES.md` + `.env.example` | Rollout penuh |

---

## Rencana Testing

- Unit test per provider dengan HTTP client yang di-mock (Google & AWS SDK mendukung custom HTTP client/endpoint; Anthropic SDK mendukung `base_url` override ke mock server lokal)
- **Golden-file HTML integrity test**: jalankan set sample `ContentHTML` hasil scraping asli (nested tag, `<a href>`, `<img>`, divider multipage `<hr class="page-divider">`) lewat tiap provider, assert struktur tag tetap utuh
- Unit test aturan resolusi provider di relay (job punya `provider_requested` / tidak / fallback dipakai / tidak ada provider tersedia → `failed`)
- Verifikasi manual E2E (perlu dilihat langsung di UI) didokumentasikan sebagai checklist untuk siapa pun yang menjalankan Phase 4

---

## Open Questions

1. ✅ **RESOLVED 2026-07-18**: kalau `translation_allow_request_override = false` dan caller tetap kirim `Provider` di request → diam-diam pakai default settings (bukan reject) — lihat §C.
2. Apakah API key provider perlu bisa diedit dari backoffice (terenkripsi) di masa depan, atau selamanya env-managed? Masih env-managed sejauh ini (`ANTHROPIC_API_KEY`).
3. ✅ **RESOLVED 2026-07-18**: `translation_llm_model` **tidak dibuat** — model LLM tetap murni `ANTHROPIC_TRANSLATION_MODEL` env var (Haiku dipakai user), untuk hindari setting kosmetik yang tidak benar-benar mengubah behavior. Lihat §A.
4. Pilihan provider per-bahasa (bukan cuma global) — worth v2, atau memang tidak pernah dibutuhkan? Masih belum, tapi arsitektur bucket-per-provider di §B (TranslationRelay) sudah siap kalau nanti mau ditambah karena resolusi sudah per-job, bukan per-tick.
5. Rate-limiting/cost cap per provider — perlu sebelum broad rollout ke banyak artikel sekaligus, atau cukup dipantau manual dulu? Masih belum digarap.

---

## Referensi

- `k-forum-api/internal/app/port/scraper_port.go:15-45` — interface `NewsTranslationProvider` + `NoopTranslationProvider`
- `k-forum-api/cmd/worker/main.go:204` — wiring global saat ini
- `k-forum-api/internal/interfaces/mq/relay/translation_relay.go` — relay polling
- `k-forum-api/internal/app/usecase/news/trigger_translation.go:36-40` — komentar yang menjelaskan kenapa `Provider` diabaikan
- `k-forum-api/internal/app/dto/news_dto.go:42-45` — `TriggerNewsTranslationRequest`
- `k-forum-api/internal/migrations/0012_create_news_tables.up.sql` — `news_system_settings`, `article_translations`, `system_languages`
- `k-forum-api/internal/migrations/0017_create_system_settings_tables.up.sql` — pola `system_settings` generic (tidak dipakai di plan ini, tapi jadi pembanding keputusan A)
- `K-FORUM-DOCS/Modules/News/NEWS_RULES.md:490-497,658`
- `K-FORUM-DOCS/plans/PLAN_NEWS_SCRAPER_ENGINE.md` — desain awal Noop & alasan disengaja
- `K-FORUM-DOCS/plans/PLAN_SYSTEM_SETTINGS_IMPLEMENTATION.md` — pelajaran "setting tidak terhubung ke backend"
- `k-forum-backoffice/app/pages/news/settings.vue`, `app/stores/newsStore.ts`
- `go.mod` — dependency baru yang perlu ditambah: `cloud.google.com/go/translate`, `github.com/aws/aws-sdk-go-v2/service/translate` (+ config/credentials), `github.com/anthropics/anthropic-sdk-go` (semuanya belum ada saat ini — storage pakai MinIO bukan AWS SDK, `google.golang.org/api` yang ada dipakai untuk FCM bukan Translation)
