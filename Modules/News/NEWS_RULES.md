# News Module — Rules & Use Cases

Dokumen ini menjelaskan aturan bisnis modul News dan bagaimana user berinteraksi dengan fitur ini. Fokus pada **siapa bisa apa**, **kapan**, dan **kenapa** — bukan dokumen teknis. Untuk detail teknis lihat `NEWS_DB_SCHEMA.md` serta API spec mobile & backoffice (menyusul).

---

## Daftar Isi

1. [Overview Konsep](#overview-konsep)
2. [Entitas Utama](#entitas-utama)
3. [Cara News Masuk ke Sistem](#cara-news-masuk-ke-sistem)
4. [Siapa Bisa Apa](#siapa-bisa-apa)
5. [Visibility: Tidak Scoped + Label Asal](#visibility-tidak-scoped--label-asal)
6. [Kategori News](#kategori-news)
7. [News Scope](#news-scope)
8. [Fitur Scraping & Scheduling](#fitur-scraping--scheduling)
9. [Status & Lifecycle](#status--lifecycle)
10. [Multi-Language & Translation](#multi-language--translation)
11. [AI Content Cleanup](#ai-content-cleanup)
12. [Interaksi Member](#interaksi-member)
13. [Article Views](#article-views)
14. [Comment](#comment)
15. [Use Cases](#use-cases)
16. [Ringkasan Aturan](#ringkasan-aturan)
17. [Keputusan yang Masih Terbuka](#keputusan-yang-masih-terbuka)

---

## Overview Konsep

News adalah modul konten editorial di platform KAI. Berbeda dari Announcement (broadcast satu arah) dan Community (konten user-generated), News adalah **konten berita yang dikurasi oleh Editor/admin** atau **diambil otomatis dari sumber eksternal** melalui scraping.

Prinsip dasar:

- **News dibuat oleh Editor/admin**, bukan member. Yang bisa membuat news adalah Usergod, Superadmin, dan Editor. Member adalah konsumen — hanya melihat dan berinteraksi (like, comment, bookmark, share).
- **News tidak scoped.** Semua news bisa dilihat secara global oleh semua user di platform. Tidak ada pembatasan "hanya region tertentu yang bisa lihat".
- **Yang membedakan news hanyalah label asal** — siapa yang posting (KAI Pusat atau region tertentu). Label ini informatif, bukan pembatas akses.
- **Scraping hanya milik KAI Pusat.** Hanya Superadmin yang bisa mengatur scheduling/scraping per source. Usergod yang mendaftarkan sumber dan selector-nya.
- **Pemisahan tanggung jawab scraping**: Usergod = mendaftarkan & mengonfigurasi *sumber* (URL, selector, cara baca HTML). Superadmin = mengatur *operasional* per source category (jadwal, mode publish, limit artikel).
- **Tidak ada approval flow untuk Editor/admin.** Status artikel Editor/admin hanya `draft → published → archived`.
- **Member Pro bisa post news** (jika benefit dikonfigurasi) dengan flow approval: `draft → pending_approval → published / rejected`.
- **Multi-language via translation.** Artikel ditulis dalam satu bahasa utama. Translation dikontrol lewat 3 level config: global system, per source, dan on-demand.

---

## Entitas Utama

### 1. Article (News)

Entitas inti yang merepresentasikan satu artikel berita.

| Atribut | Keterangan & Contoh |
|---|---|
| `source_id` | FK ke `news_sources`. Kosong (null) jika artikel ditulis manual, terisi jika hasil scraping |
| `news_scope_id` | FK ke `news_scopes`. Klasifikasi **asal/fokus geografis** berita: `indonesia`, `korea`, atau `korea_indonesia`. Saat scraping diwariskan dari `news_sources.default_scope_id`; Editor bisa override. Boleh kosong (belum dipetakan) |
| `original_language` | Bahasa asli artikel. Contoh: `id`. Mengikuti source jika scraping, dipilih Editor jika manual |
| `is_manual` | `true` jika ditulis manual oleh Editor/member, `false` jika hasil scraping |
| `status` | Status artikel: `draft`, `pending_approval`, `published`, `archived`, `rejected` |
| `author_label` | Label asal yang tampil ke pembaca. Contoh: `Korean Association Indonesia`, `KAI Jakarta` |
| `author_region_id` | FK region asal — kosong (null) jika dari KAI Pusat, terisi jika dari region tertentu |
| `original_url` | URL asli artikel di sumbernya (hasil scraping). Dipakai untuk cek duplikat. Contoh: `https://detik.com/sport/12345`. Kosong jika manual |
| `is_featured` | `true` = artikel di-pin/disorot di halaman utama |
| `view_count` | Total berapa kali artikel dibuka (termasuk guest & buka berulang). Contoh: `1523` |
| `unique_view_count` | Berapa user login berbeda yang membuka. Contoh: `840` |
| `like_count` | Total like (disimpan langsung untuk performa, tidak hitung ulang tiap query) |
| `comment_count` | Total komentar (disimpan langsung untuk performa) |
| `created_by` | FK ke users — Editor/admin/member yang membuat. Kosong (null) jika murni scraping otomatis tanpa campur tangan editor |
| `published_by_user_id` | FK ke users — siapa yang menekan tombol publish. Kosong jika auto-publish oleh sistem. **Hanya untuk audit internal, tidak tampil ke publik** |
| `published_by_label` | Label penerbit yang **tampil ke pembaca**. Contoh: `Korean Association Indonesia` atau `KAI Jakarta` |
| `published_at` | Waktu artikel tayang |

> **Penting — bedakan dua makna "scope":**
> - **Visibility scope** (global/regional) → Article **tidak** punya ini. Semua article visible global ke semua user. `author_region_id` di sini hanya **label asal**, bukan pembatas siapa yang boleh lihat.
> - **News scope** (`news_scope_id`) → ini **klasifikasi konten** (asal/fokus geografis: Indonesia / Korea / Korea di Indonesia), dipakai sebagai **filter** di UI. Tidak membatasi visibility — artikel scope `korea` tetap bisa dilihat semua orang, cuma bisa difilter terpisah. Lihat bagian [News Scope](#news-scope).

### 2. Article Translation

Menyimpan konten artikel per bahasa. Satu artikel bisa punya banyak baris translation (satu per bahasa).

| Atribut | Keterangan & Contoh |
|---|---|
| `article_id` | FK ke `articles` |
| `language` | Kode bahasa baris ini. Contoh: `id`, `en`, `ko` |
| `title` | Judul dalam bahasa ini |
| `content` | Isi lengkap artikel dalam bahasa ini |
| `summary` | Ringkasan singkat (excerpt) dalam bahasa ini |
| `author` | Nama penulis (hasil scraping atau diisi manual) |
| `thumbnail_url` | URL gambar thumbnail artikel |
| `tags` | Daftar tag artikel. Contoh: `["sepakbola", "timnas", "piala dunia"]` |
| `is_original` | `true` = konten asli (hasil scraping/ditulis editor). `false` = hasil terjemahan otomatis |
| `translate_status` | Status proses terjemahan: `pending`, `processing`, `done`, `failed`. Hanya berlaku jika `is_original=false` |
| `translated_by` | Mesin penerjemah yang dipakai: `google`, `openai`, atau kosong (null) jika diisi manual |

### 3. News Source

Sumber berita eksternal untuk scraping. Didaftarkan oleh **Usergod**, dikonfigurasi operasionalnya oleh **Usergod dan Superadmin**.

| Atribut | Keterangan & Contoh |
|---|---|
| `key` | Identifier unik sumber, huruf kecil tanpa spasi. Contoh: `detik`, `kompas`, `kai_official` |
| `name` | Nama tampil sumber. Contoh: `Detik.com`, `Kompas`, `KAI Official` |
| `base_url` | URL dasar feed sumber. Contoh: `https://www.cnnindonesia.com/` |
| `original_language` | Kode bahasa asli konten sumber ini. Contoh: `id` (untuk portal Indonesia), `ko` (portal Korea) |
| `default_scope_id` | FK ke `news_scopes`. Scope default yang diwariskan ke setiap artikel hasil scraping dari source ini. Contoh: Antara/Detik → `indonesia`, Yonhap/Korea Times → `korea`. Boleh kosong; Editor tetap bisa set scope per artikel |
| `schedule` | Jadwal scraping dalam format cron expression. Contoh: `0 */6 * * *` = tiap 6 jam, `0 8,12,17 * * *` = jam 8/12/17 setiap hari, `0 8 * * 1-5` = jam 8 tiap hari kerja |
| `last_scraped_at` | Timestamp terakhir kali scraping berhasil. Diisi sistem otomatis, dipakai scheduler untuk hitung jadwal berikutnya |
| `auto_publish` | `true` = hasil scraping langsung tayang ke user. `false` = masuk antrian draft untuk direview Editor dulu. Default: `false` |
| `ai_cleanup` | `true` = setelah scraping, konten dirapikan AI (buang sisa noise, format ulang). Aktifkan untuk source yang HTML-nya berantakan. Default: `false` |
| `auto_translate` | `true` = setelah scraping, langsung generate terjemahan ke semua bahasa aktif. `false` = terjemahan menyusul (on-demand). Default: `false` |
| `is_active` | `true` = source ikut dijadwalkan scraping. `false` = di-pause tanpa hapus konfigurasi |
| `created_by` | Usergod yang mendaftarkan source |

> `schedule`, `auto_publish`, `ai_cleanup`, dan `auto_translate` bisa di-toggle oleh **Usergod dan Superadmin** sesuai permission. Scheduler sistem berjalan tiap menit, mengecek source mana yang sudah waktunya berdasarkan cron expression + `last_scraped_at`.

### 4. Source Selectors

Konfigurasi teknis cara parsing HTML per source. Didaftarkan oleh **Usergod**.

> **Penting — metode ekstraksi konten:** Konten utama artikel **tidak** diekstrak murni dari CSS selector, melainkan menggunakan library **Readability** (algoritma yang sama dipakai mode "Reader" di browser — otomatis mendeteksi badan artikel dan membuang noise seperti menu, iklan, sidebar). CSS selector di sini berperan sebagai **fallback** (jika Readability gagal) dan untuk mengambil **author + tags** yang sering tidak terdeteksi Readability.

| Atribut | Keterangan & Contoh |
|---|---|
| `source_id` | FK ke `news_sources` |
| `content_selector` | CSS selector fallback untuk badan artikel, jika Readability gagal. **Boleh kosong** — Readability jadi andalan utama. Contoh: `.detail-text, #detikdetailtext` |
| `author_selector` | CSS selector untuk nama penulis. Contoh: `.author, .writer, .byline, .reporter` |
| `tags_selector` | CSS selector untuk tag artikel. Contoh: `.detail-tag a, .tag-links a` |
| `extra_fields` | JSONB — konfigurasi tambahan untuk edge case per source. Contoh: aturan multi-page (`{"multipage": true, "next_text": ["selanjutnya", "next"]}`) atau override khusus seperti Okezone |

**Tentang multi-page:** Sebagian portal memecah satu artikel jadi beberapa halaman ("halaman 1, 2, 3" atau tombol "Selanjutnya"). Scraper otomatis mendeteksi dan menggabungkan semua halaman jadi satu konten utuh. Aturan deteksi per source bisa diatur di `extra_fields` jika polanya tidak standar (mis. Tribun yang pakai `?page=all`).

### 5. Source Categories

Definisi kategori yang di-scrape dari suatu source beserta batas artikel per fetch. Diatur oleh **Superadmin**. Scheduling dan mode publish dikontrol di level `news_sources`, bukan di sini.

| Atribut | Keterangan & Contoh |
|---|---|
| `source_id` | FK ke `news_sources` |
| `category_key` | **Input manual** — nama/slug/path kategori sesuai feed source. String bebas. Contoh: `sport`, `ekonomi`, `lifestyle`, `sepakbola.xml`. Dipakai untuk membangun URL feed |
| `category_id` | FK ke `news_categories` — **mapping** hasil scrape dari `category_key` ini ke kategori KAI mana. Saat scraping, `articles.category_id` diisi dari sini. Kosong (null) = belum dipetakan (artikel fallback ke kategori `umum`) |
| `url_suffix` | Bagian URL yang ditambahkan ke `base_url` untuk kategori ini. Contoh: jika base_url = `https://rss.tempo.co/` dan suffix = `nasional`, maka feed = `https://rss.tempo.co/nasional` |
| `url_override` | URL feed penuh, dipakai jika pola tidak cocok dengan base_url + suffix. Contoh: `https://www.antaranews.com/rss/terkini.xml`. Kosongkan jika cukup pakai `url_suffix` |
| `article_limit` | Maksimal artikel yang diambil per sekali fetch dari kategori ini. Contoh: `10` (ambil 10 artikel terbaru tiap scraping) |
| `is_active` | `true` = kategori ikut di-scrape. `false` = di-skip tanpa hapus konfigurasi |
| `updated_by` | Superadmin yang terakhir mengubah konfigurasi ini |

### 6. System Languages

Bahasa yang aktif di platform. Dikelola oleh **Usergod**.

| Atribut | Keterangan & Contoh |
|---|---|
| `code` | Kode bahasa standar (ISO). Contoh: `id`, `en`, `ko`, `jp` |
| `name` | Nama bahasa untuk ditampilkan. Contoh: `Indonesian`, `English`, `Korean` |
| `is_ui_language` | `true` = bahasa ini bisa dipakai untuk antarmuka aplikasi (menu, tombol, dll) |
| `is_translate_target` | `true` = artikel boleh diterjemahkan ke bahasa ini. Saat `auto_translate` jalan, hanya bahasa dengan flag ini yang di-generate |
| `is_active` | `true` = bahasa aktif dan tersedia di sistem |

### 7. News System Settings

Konfigurasi global sistem News. Dikelola oleh Usergod dan Superadmin (sesuai permission).

| Atribut | Keterangan & Contoh |
|---|---|
| `translation_enabled` | Saklar utama fitur terjemahan. `false` = fitur mati, semua user hanya dapat bahasa asli artikel (paling hemat biaya) |
| `on_demand_enabled` | Berlaku jika `translation_enabled=true`. `true` = sistem otomatis menerjemahkan saat ada user membuka artikel dalam bahasa yang belum tersedia. `false` = user harus memilih bahasa secara eksplisit dulu baru diterjemahkan |

---

## Cara News Masuk ke Sistem

Ada dua jalur utama.

### Jalur 1 — Manual

Editor/admin membuat artikel langsung di backoffice. Editor memilih bahasa utama, menulis konten, lalu publish atau simpan sebagai draft. Article bertanda `is_manual=true`, `source_id=null`.

Jika Member Pro punya benefit `post_news` → bisa submit artikel dari mobile, masuk `pending_approval` sebelum tayang.

### Jalur 2 — Scraping Otomatis (KAI Pusat Saja)

Scheduler membaca source aktif dari DB sesuai interval yang dikonfigurasi Superadmin. Sistem:
1. Fetch RSS
2. Cek duplikat via `original_url` — jika sudah ada di DB, skip
3. Scrape detail HTML menggunakan selector dari DB
4. Opsional AI cleanup (jika `ai_cleanup=true`)
5. Simpan ke `articles` + `article_translations` (`is_original=true`)
6. Opsional auto-translate semua bahasa aktif (jika `auto_translate=true`)
7. Masuk `draft` atau langsung `published` tergantung `auto_publish` per source category

**Scraping eksklusif KAI Pusat** — Admin Region tidak punya jalur ini.

---

## Siapa Bisa Apa

### Usergod

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Daftarkan & konfigurasi news source (URL, selector) | ✅ Ya | Setup teknis sumber |
| Edit/hapus source | ✅ Ya | |
| Toggle `ai_cleanup` & `auto_translate` per source | ✅ Ya | Bersama Superadmin |
| Kelola system languages | ✅ Ya | Tambah/nonaktifkan bahasa |
| Kelola news system settings (translation_enabled, on_demand_enabled) | ✅ Ya | Sesuai permission |
| Semua aksi Superadmin | ✅ Ya | Akses penuh |

### Superadmin (KAI Pusat)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Konfigurasi source categories (jadwal, auto_publish, limit) | ✅ Ya | Eksklusif KAI Pusat |
| Toggle `ai_cleanup` & `auto_translate` per source | ✅ Ya | Bersama Usergod |
| Activate/deactivate source | ✅ Ya | |
| Kelola news system settings | ✅ Ya | Sesuai permission |
| Buat artikel manual | ✅ Ya | Label asal = KAI Pusat |
| Publish / archive artikel | ✅ Ya | Semua artikel |
| Edit artikel apapun | ✅ Ya | |
| Approve / reject artikel dari Member Pro | ✅ Ya | |
| Kelola kategori news | ✅ Ya | CRUD, urutan, aktif/nonaktif |
| Daftarkan source baru | ❌ Tidak | Hanya Usergod |

### Editor

Role sistem. Di-assign oleh Usergod atau Superadmin (sesuai permission). Fokus pada operasional konten sehari-hari.

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Buat artikel manual | ✅ Ya | Label asal sesuai region Editor |
| Review & publish artikel dari draft | ✅ Ya | Termasuk hasil scraping yang masuk draft |
| Edit artikel | ✅ Ya | Termasuk artikel yang sudah published — tanpa re-approval |
| Archive artikel | ✅ Ya | |
| Approve / reject artikel dari Member Pro | ✅ Ya | Jika punya permission `approve_news` |
| Konfigurasi source / scheduling | ❌ Tidak | Eksklusif Usergod/Superadmin |
| Kelola kategori | ❌ Tidak | Hanya Superadmin |
| Kelola system languages / settings | ❌ Tidak | Hanya Usergod/Superadmin |

### Member Pro (jika benefit `post_news` aktif)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Submit artikel | ✅ Ya | Masuk `pending_approval` |
| Edit artikel milik sendiri (status draft/rejected) | ✅ Ya | |
| Withdraw artikel (batalkan submission) | ✅ Ya | Sebelum di-approve |
| Publish langsung | ❌ Tidak | Selalu lewat approval |

### Member (Standard & Pro — sebagai konsumen)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Baca semua artikel published | ✅ Ya | Global — semua artikel, apapun asalnya |
| Pilih bahasa baca | ✅ Ya | Sesuai bahasa yang tersedia |
| Like artikel | ✅ Ya | |
| Comment artikel | ✅ Ya | Detail flow comment dibahas terpisah |
| Save/bookmark artikel | ✅ Ya | Untuk dibaca lagi nanti |
| Share artikel | ✅ Ya | Ke WhatsApp, sosmed, copy link |
| Filter berdasarkan kategori / label asal | ✅ Ya | Opsional di UI |

### Guest (belum login)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Baca artikel published | ✅ Ya | Bisa baca tanpa login |
| Like / comment / bookmark / share | ❌ Tidak | Harus login |

---

## Visibility: Tidak Scoped + Label Asal

Perbedaan kunci News dengan modul Announcement:

- **Announcement** punya scope global/regional → menentukan **siapa yang bisa lihat**.
- **News** tidak punya scope → **semua artikel bisa dilihat semua orang**, global.

Yang membedakan artikel hanyalah **label asal** (`author_label`):

- Artikel dari Superadmin/Editor KAI Pusat → label **KAI Pusat**
- Artikel dari Editor Jakarta → label **KAI Jakarta**
- Artikel hasil scraping → label **KAI Pusat**
- Artikel dari Member Pro → label nama member (atau KAI Pusat jika diapprove pusat)

Fungsi label asal bersifat informatif dan filter opsional UI — bukan pembatas akses.

---

## Kategori News

Sistem kategori seperti portal berita umum (mis. Politik, Ekonomi, Olahraga, Teknologi, Hiburan, dll).

- Dikelola oleh **Superadmin** — CRUD, atur urutan tampil, aktif/nonaktif.
- Setiap artikel wajib punya satu kategori.
- Hasil scraping diberi kategori default per source, bisa dikoreksi Editor saat review.
- Member bisa filter dan browse berdasarkan kategori.

---

## News Scope

Dimensi klasifikasi **terpisah dan orthogonal** dari kategori. Kalau kategori menjawab *"berita tentang apa"* (Olahraga, Ekonomi), scope menjawab *"berita dari/tentang mana"* secara geografis.

### Nilai Scope (Phase 1)

| Slug | Nama | Maksud |
|---|---|---|
| `indonesia` | Berita Indonesia | Berita umum dari/tentang Indonesia |
| `korea` | Berita Korea | Berita dari/tentang Korea |
| `korea_indonesia` | Berita Korea di Indonesia | Komunitas Korea di Indonesia, hubungan bilateral, topik lintas keduanya |

> Scope disimpan sebagai master table (`news_scopes`) dengan FK, bukan enum hardcoded — supaya scope baru (mis. `bilateral`, `asean`) bisa ditambah tanpa ubah skema. Dikelola oleh **Superadmin** (CRUD, urutan, aktif/nonaktif).

### Aturan

- **Orthogonal dengan kategori.** Satu artikel punya satu kategori *dan* satu scope. Contoh: artikel Olahraga + scope Korea, atau Ekonomi + scope Korea di Indonesia.
- **Satu artikel = satu scope** (bukan multi). Boleh kosong jika belum dipetakan.
- **Bukan visibility.** Scope hanya untuk filtering/browsing di UI. Semua artikel tetap visible global ke semua user, apa pun scope-nya. (Beda dengan Announcement yang scope-nya membatasi siapa yang lihat.)
- **Default dari source.** Source punya `default_scope_id`. Saat scraping, artikel mewarisi scope dari source-nya. Editor bisa override per artikel (mis. artikel Antara yang membahas komunitas Korea di Jakarta → ubah ke `korea_indonesia`).
- **Artikel manual.** Editor pilih scope sendiri saat tulis artikel.

---

## Fitur Scraping & Scheduling

Eksklusif KAI Pusat. Dipisah menjadi dua tingkat tanggung jawab:

### Usergod — Definisi Sumber (Setup Teknis)

1. Tambah source (key, name, base_url, bahasa asli)
2. Input CSS selector (content fallback, author, tags, extra fields) — catatan: konten utama diekstrak via Readability, selector hanya fallback + author/tags
3. Set cron schedule (mis. `"0 */6 * * *"` = tiap 6 jam, `"0 8 * * 1-5"` = jam 8 tiap hari kerja)
4. Toggle `auto_publish` — langsung tayang atau masuk draft
5. Toggle `ai_cleanup` — aktifkan jika konten source biasanya berantakan
6. Toggle `auto_translate` — aktifkan jika artikel source ini harus langsung tersedia semua bahasa

### Superadmin — Operasional Source & Categories

Superadmin bisa update semua config di `news_sources` (schedule, auto_publish, ai_cleanup, auto_translate, is_active) dan mengatur `source_categories` per source:

1. Tambah/edit kategori yang di-scrape (`category_key`, url_suffix/override)
2. Mapping `category_key` → `category_id` (kategori KAI) agar hasil scrape masuk kategori yang benar
3. Set `article_limit` — maksimal artikel per fetch per kategori
4. Set `default_scope_id` di level source (Indonesia/Korea/Korea di Indonesia)
5. Activate/deactivate kategori tanpa hapus konfigurasi

### Cara Scheduler Bekerja

Scheduler sistem berjalan **tiap menit**. Setiap tick, scheduler cek semua source aktif:

```
Untuk tiap source aktif:
  Evaluasi cron expression + last_scraped_at
  Sudah waktunya? → jalankan scraping job
  Belum waktunya? → skip
```

Ini memungkinkan jadwal yang sangat fleksibel:
```
"*/30 * * * *"      → tiap 30 menit
"0 * * * *"         → tiap jam tepat
"0 8,12,17 * * *"   → jam 8, 12, 17 setiap hari
"0 8 * * 1,3,5"     → jam 8, Senin/Rabu/Jumat saja
"0 0 * * *"         → sekali sehari tengah malam
```

### Metode Ekstraksi Konten

Saat scraping halaman detail artikel, konten utama diekstrak dengan kombinasi:

1. **Readability (utama)** — algoritma yang otomatis mengenali badan artikel dan membuang elemen non-konten (menu, iklan, sidebar, footer). Tidak perlu selector spesifik per source untuk body artikel.
2. **CSS Selector (fallback & pelengkap)** — dipakai jika Readability gagal, dan untuk mengambil **author** + **tags** yang sering tidak terdeteksi Readability. Diatur di `source_selectors`.
3. **Multi-page handling** — jika artikel terpecah ke beberapa halaman ("Selanjutnya", "halaman 2", atau `?page=all`), scraper otomatis mengikuti dan menggabungkan jadi satu konten utuh.

### Alur Scraping

```
Scheduler tick tiap menit → cek source yang sudah waktunya
        ↓
Fetch RSS feed → dapat list artikel (title, url, pubDate, excerpt, thumbnail)
        ↓
Dedup: filter via original_url → buang artikel yang sudah ada di DB
        ↓
Ambil maksimal article_limit dari artikel BARU saja
        ↓
Fetch detail tiap artikel baru (concurrent, dengan rate limit):
   → Readability ekstrak konten utama
   → Selector ambil author + tags
   → Gabungkan multi-page jika ada
        ↓
[ai_cleanup = true?] → kirim ke AI → konten bersih
        ↓
Set category_id  ← dari source_categories.category_id (mapping), fallback 'umum' jika null
Set news_scope_id ← dari news_sources.default_scope_id
        ↓
Simpan articles + article_translations (is_original=true, language=original_language)
Update last_scraped_at
        ↓
[auto_translate = true?] → enqueue translation job ke semua bahasa aktif (is_translate_target=true)
        ↓
auto_publish = true?
├── Ya  → status = published
└── Tidak → status = draft → Editor review → published / archived
```

> **Optimasi penting:** Dedup dilakukan **sebelum** fetch detail. RSS bisa mengembalikan 20 artikel tapi mungkin hanya 3 yang benar-benar baru — jadi hanya 3 itu yang di-fetch detail-nya. Hemat banyak HTTP request.

---

## Status & Lifecycle

### Artikel dari Editor / Admin (Tanpa Approval)

```
draft ──publish──> published ──archive──> archived
                      │
                   (bisa diedit langsung, tanpa re-approval)
```

### Artikel dari Member Pro (Dengan Approval)

```
draft ──submit──> pending_approval ──approve──> published ──archive──> archived
                        │
                     reject
                        ↓
                    rejected (bisa diedit & resubmit)
```

**Status detail:**

- **DRAFT** — belum tayang. Dari scraping (auto_publish=false), manual Editor yang disimpan dulu, atau artikel Member Pro yang belum disubmit.
- **PENDING_APPROVAL** — khusus artikel Member Pro. Menunggu review dari Editor/Superadmin yang punya permission `approve_news`.
- **PUBLISHED** — tayang ke semua user. Editor/admin bisa edit langsung tanpa re-approval.
- **REJECTED** — artikel Member Pro ditolak. Member bisa edit dan resubmit.
- **ARCHIVED** — tidak muncul di listing utama, tetap tersimpan untuk arsip/history.

---

## Multi-Language & Translation

### 3 Level Config Translation

```
Level 1 — Global System Setting (News System Settings)
├── translation_enabled = false → fitur translate mati total, semua dapat original
└── translation_enabled = true → fitur translate aktif, lanjut ke level berikutnya
        │
        ├── on_demand_enabled = true
        │   → auto-trigger translate saat user hit artikel dalam bahasa yang belum ada
        │   → return original dulu, translation selesai → cached, next request dapat hasil
        │
        └── on_demand_enabled = false
            → tidak ada auto-trigger
            → user eksplisit pilih bahasa sendiri, baru translate di-generate

Level 2 — Per Source (news_sources)
└── auto_translate = true
    → setelah scraping selesai, langsung generate translation ke semua bahasa aktif
    → override on_demand: user request bahasa apapun langsung dapat hasil

Level 3 — Artikel Manual
└── ikut global setting (source_id = null, tidak ada per-source config)
```

### Flow Tampil Artikel ke User

```
Request artikel + Accept-Language: en
        ↓
Cek article_translations WHERE language='en' AND translate_status='done'
        ├── Ada → return versi EN ✅
        └── Tidak ada
                ↓
            [on_demand_enabled = true?]
            ├── Ya → trigger background job translate
            │        return versi original + flag: { is_translated: false }
            │        job selesai → cached → next request dapat EN
            └── Tidak → return versi original + flag: { is_translated: false }
                         user bisa eksplisit pilih bahasa lain dari daftar available_languages
```

### Kombinasi Mode

| `translation_enabled` | `on_demand_enabled` | `auto_translate` (source) | Hasil |
|---|---|---|---|
| false | - | - | Semua dapat bahasa original saja |
| true | false | false | User pilih bahasa eksplisit → baru translate |
| true | true | false | Auto-trigger saat user hit bahasa yang belum ada |
| true | - | true | Semua bahasa siap sejak artikel scraping selesai |

### Provider Translate

- **Google Translate API** — default (murah, cepat, support banyak bahasa)
- **OpenAI** — alternatif untuk kualitas lebih natural (lebih mahal per token)
- Field `translated_by` di `article_translations` menyimpan provider yang dipakai — traceable

### Artikel Manual — Translation

Editor hanya wajib mengisi **satu bahasa utama** saat membuat artikel. Bahasa lain di-generate mengikuti global setting (on_demand atau eksplisit oleh user).

---

## AI Content Cleanup

Scraping HTML bisa menghasilkan konten berantakan (noise dari navigasi, iklan, whitespace berlebih). Ada dua layer:

### Layer 1 — Technical Cleanup (Selalu Berjalan)

Ditangani goquery selector: strip HTML tags sisa, buang elemen non-konten via selector, normalize whitespace.

### Layer 2 — AI Cleanup (Opsional, Per Source)

Jika `news_sources.ai_cleanup = true`:

```
Raw content hasil scraping → kirim ke AI dengan prompt cleanup → konten bersih tersimpan
```

Diaktifkan per source sesuai kebutuhan — source berantakan aktifkan, source rapi tidak perlu. Toggle oleh **Usergod dan Superadmin** sesuai permission.

---

## Interaksi Member

Member adalah konsumen artikel. Yang bisa dilakukan terhadap artikel published:

- **Membaca** daftar dan detail artikel (semua artikel, global)
- **Memilih bahasa** baca sesuai preferensi — artikel tampil dalam bahasa setting user (`Accept-Language`), fallback ke original jika translation belum ada
- **Like** artikel
- **Comment** artikel — threaded 2 level (comment + reply)
- **Menyimpan / bookmark** artikel untuk dibaca nanti
- **Membagikan** artikel ke WhatsApp, media sosial, atau copy link
- **Memfilter** berdasarkan kategori atau label asal (opsional di UI)
- **Menerima notifikasi** artikel baru (sesuai preferensi notification)

---

## Article Views

Dua jenis view count disimpan terpisah di tabel `articles`:

| Field | Keterangan |
|---|---|
| `view_count` | Total hit — setiap kali artikel dibuka, termasuk guest dan user yang buka berkali-kali |
| `unique_view_count` | Unique per user login — setiap user dihitung satu kali saja |

Guest hanya dihitung di `view_count`, tidak di `unique_view_count` karena tidak ada user ID.

---

## Comment

### Struktur

Threaded 2 level — comment dan reply. Tidak ada nesting lebih dalam.

```
Comment A (level 1)
  └── Reply A1 (level 2)
  └── Reply A2 (level 2)
Comment B (level 1)
  └── Reply B1 (level 2)
```

Jika user tap "Reply" pada reply (level 2) → tetap masuk sebagai reply ke parent comment (level 1), bukan nested lebih dalam. Behavior umum seperti Instagram/Twitter.

### Aturan Comment

| Aturan | Detail |
|---|---|
| **Siapa yang bisa comment** | Member login (Standard & Pro) |
| **Guest** | ✅ Bisa baca comment, ❌ tidak bisa comment |
| **Edit comment** | ❌ Tidak bisa |
| **Delete comment** | ✅ Soft delete — oleh pemilik comment sendiri atau Editor/Superadmin |
| **Tampilan setelah delete** | "Komentar ini telah dihapus" — reply di bawahnya tetap tampil |
| **Like comment** | ❌ Tidak ada — like hanya di level artikel |
| **Moderasi** | Via reporting system (`reportable_type = news_comment`) |
| **Max level** | 2 (comment → reply, tidak bisa reply ke reply) |

### Flow Comment

```
Member tulis comment → submit
        ↓
Simpan dengan parent_id = null (level 1)
comment_count artikel += 1

Member tap Reply pada comment/reply → submit
        ↓
Simpan dengan parent_id = comment level 1
(jika reply ke reply → parent_id tetap ke level 1)
comment_count artikel += 1

Member/Editor/Superadmin delete comment
        ↓
Soft delete: is_deleted = true, content = null
Tampil: "Komentar ini telah dihapus"
Reply di bawahnya tetap tampil

Member report comment
        ↓
Masuk reporting system (reportable_type = news_comment, reportable_id = comment_id)
```

---

## Use Cases

### Use Case 1 — Editor Buat Artikel Manual

Editor membuka backoffice, klik "Buat Artikel", memilih bahasa utama "ID", mengisi konten, memilih kategori, lalu publish. Artikel langsung tayang. User dengan `Accept-Language: en` yang membuka artikel — jika `on_demand_enabled=true` → sistem auto-trigger translate di background, user dapat versi ID dulu, lalu EN tersedia untuk request berikutnya.

### Use Case 2 — Scheduling Fleksibel Per Source

Superadmin mengkonfigurasi source "Detik" dengan schedule `"0 */2 * * *"` (tiap 2 jam) dan `auto_publish=false` — hasil masuk draft dulu. Source "KAI Official" dikonfigurasi `"0 8 * * *"` (sekali sehari jam 8 pagi) dengan `auto_publish=true` dan `auto_translate=true` — langsung tayang dalam semua bahasa aktif. Tiap source berjalan sesuai jadwal masing-masing secara independen.

### Use Case 3 — Scraping dengan Draft Mode & AI Cleanup

Source "Okezone" diset `ai_cleanup=true` karena HTML berantakan, `auto_publish=false`. Artikel masuk draft setelah dirapikan AI. Editor review antrian draft, edit seperlunya, publish yang layak.

### Use Case 4 — Translation Off, User Pilih Bahasa Sendiri

Global setting: `translation_enabled=true`, `on_demand_enabled=false`. User membuka artikel — dapat versi original. Di UI tersedia dropdown pilihan bahasa (`available_languages`). User pilih "EN" → sistem trigger translate → cached → user dapat versi EN.

### Use Case 5 — Member Pro Submit Artikel

Member Pro menulis artikel di mobile, submit. Status berubah ke `pending_approval`. Editor/Superadmin yang punya permission `approve_news` menerima notifikasi, review, lalu approve. Artikel tayang. Jika ditolak, member bisa edit dan resubmit.

### Use Case 6 — De-duplikasi Scraping

Scheduler jalan, ambil feed RSS Detik. Artikel "Pertandingan Sepakbola Indonesia" sudah pernah di-scrape 3 jam lalu (`original_url` sudah ada di DB). Sistem skip artikel itu, hanya insert artikel yang benar-benar baru.

### Use Case 7 — Editor Edit Artikel Published

Ditemukan typo di artikel yang sudah tayang. Editor buka artikel, edit, simpan — langsung terupdate tanpa approval ulang.

---

## Ringkasan Aturan

| Aturan | Detail |
|--------|--------|
| **Siapa yang bisa buat artikel** | Usergod, Superadmin, Editor — dan Member Pro jika benefit aktif |
| **Approval flow** | Hanya untuk artikel Member Pro (`pending_approval`) — Editor/admin tidak perlu approval |
| **Status artikel Editor/admin** | `draft → published → archived` |
| **Status artikel Member Pro** | `draft → pending_approval → published / rejected` |
| **Edit artikel published** | Editor/admin langsung, tanpa re-approval |
| **Scheduling scraping** | Cron expression per source — fleksibel (per menit, jam, hari tertentu, dll) |
| **Scraping** | Eksklusif KAI Pusat — schedule & auto_publish di level source, article_limit di level category |
| **Daftarkan & config source** | Usergod saja |
| **Metode ekstraksi konten** | Readability (utama) + CSS selector (fallback/author/tags) + multi-page merge |
| **De-duplikasi scraping** | Cek `original_url` sebelum fetch detail — skip jika sudah ada (MVP) |
| **AI cleanup** | Per source (`ai_cleanup`), toggle by Usergod & Superadmin |
| **Auto translate** | Per source (`auto_translate`), toggle by Usergod & Superadmin |
| **Translation global** | Dikontrol `translation_enabled` + `on_demand_enabled` di News System Settings |
| **Bahasa artikel ke user** | Berdasarkan `Accept-Language` header — fallback ke original jika tidak ada |
| **Translation provider** | Google Translate (default) atau OpenAI — traceable via `translated_by` |
| **Reaksi member** | Like saja di level artikel (MVP) — bisa extend ke emoji reaction nanti |
| **Comment** | Threaded 2 level, soft delete, no edit, moderasi via reporting system |
| **Like comment** | ❌ Tidak ada — like hanya di artikel |
| **Published by (publik)** | Label organisasi — `"Korean Association Indonesia"` (KAI Pusat/scraping) atau `"KAI [Region]"` (admin region) |
| **Published by (internal)** | `published_by_user_id` — FK ke users, untuk audit. Null jika auto_publish oleh sistem |
| **Guest access** | Bisa baca artikel & comment tanpa login — tidak bisa like, comment, bookmark, share |
| **Kelola kategori** | Superadmin saja |
| **System languages** | Dikelola Usergod |
| **News system settings** | Dikelola Usergod & Superadmin sesuai permission |
| **Article views** | `view_count` (total hit) + `unique_view_count` (unique per user login) |

---

## Keputusan yang Masih Terbuka

1. **Rate-limit translate on-demand** — berapa concurrent job yang diizinkan? Dikontrol di level queue.
2. **Reaksi** — hanya like untuk MVP, extend ke emoji reaction di iterasi berikutnya.
3. **De-duplikasi lanjutan** — hash konten sebagai pelengkap cek URL, dipertimbangkan jika ada kasus duplikat lolos.
4. **Notifikasi** — push notif artikel baru & notifikasi approval artikel Member Pro dibahas di modul Notifikasi.
5. **Label asal artikel Member Pro** — nama member, atau mengikuti region member?

---

*Dokumen rules News module — non-teknis. Untuk detail API lihat `API_SPEC_NEWS_MOBILE` & `API_SPEC_NEWS_BACKOFFICE` (menyusul). Untuk skema database lihat `NEWS_DB_SCHEMA` (menyusul).*
