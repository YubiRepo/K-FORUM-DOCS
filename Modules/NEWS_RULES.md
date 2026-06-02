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
7. [Fitur Scraping & Scheduling](#fitur-scraping--scheduling)
8. [Setting Approval](#setting-approval)
9. [Status & Lifecycle](#status--lifecycle)
10. [Interaksi Member](#interaksi-member)
11. [Multi-Language / Translation (Open)](#multi-language--translation-open)
12. [Use Cases](#use-cases)
13. [Ringkasan Aturan](#ringkasan-aturan)
14. [Keputusan yang Masih Terbuka](#keputusan-yang-masih-terbuka)

---

## Overview Konsep

News adalah modul konten editorial di platform KAI. Berbeda dari Announcement (broadcast satu arah) dan Community (konten user-generated), News adalah **konten berita yang dikurasi oleh admin** atau **diambil otomatis dari sumber eksternal** melalui scraping.

Lima prinsip dasar:

- **News dibuat oleh admin**, bukan member. Yang bisa membuat news adalah Superadmin (KAI Pusat) dan Admin Region. Member murni konsumen — hanya melihat dan berinteraksi (save, like, share).
- **News tidak scoped.** Semua news bisa dilihat secara global oleh semua user di platform. Tidak ada pembatasan "hanya region tertentu yang bisa lihat".
- **Yang membedakan news hanyalah label asal** — siapa yang posting (KAI Pusat atau region tertentu). Label ini informatif, bukan pembatas akses.
- **Scraping hanya milik KAI Pusat.** Hanya Superadmin yang bisa mengatur scheduling/scraping. Admin Region hanya bisa posting manual dan approval. Usergod yang mendaftarkan sumber RSS-nya.
- **Pemisahan tanggung jawab scraping**: Usergod mendaftarkan & mengonfigurasi *sumber* (cara baca RSS), sedangkan Superadmin mengatur *operasional* tiap sumber (jadwal & mode publish).

---

## Entitas Utama

### 1. News

Entitas inti yang merepresentasikan satu artikel berita.

| Atribut | Keterangan |
|---|---|
| `title`, `content`, `excerpt` | Isi berita |
| `cover_image`, `images` | Gambar berita |
| `category_id` | Kategori berita (FK ke News Category) |
| `source_id` | Sumber berita — null jika dibuat manual, terisi jika hasil scraping |
| `author_label` | Label asal news: `KAI Pusat` atau nama region (mis. `KAI Jakarta`) |
| `author_region_id` | Region asal author — null jika dari KAI Pusat, terisi jika dari Admin Region |
| `origin` | `manual` atau `scraped` — menandakan asal news |
| `status` | `draft`, `pending_approval`, `published`, `archived`, `rejected` |
| `created_by` | Admin yang membuat (null jika murni hasil scraping otomatis) |
| `published_at` | Waktu tayang |

> **Penting:** News **tidak** punya field `scope` (global/regional) seperti Announcement. Semua news visible global. Field region di sini hanya **label asal** untuk menunjukkan siapa yang posting.

### 2. News Category

Kategori berita umum seperti pada portal berita pada umumnya.

| Atribut | Keterangan |
|---|---|
| `name`, `slug` | Identitas kategori (mis. "Politik", "Ekonomi", "Olahraga", "Teknologi") |
| `is_active` | Aktif/nonaktif |
| `sort_order` | Urutan tampil |

Kategori dikelola oleh Superadmin (CRUD, atur urutan, aktif/nonaktif).

### 3. News Source

Sumber berita eksternal untuk fitur scraping. Didaftarkan dan dikonfigurasi oleh **Usergod**.

| Atribut | Keterangan |
|---|---|
| `name` | Nama sumber (mis. "Detik", "Kompas") |
| `feed_url` | URL feed RSS sumber (mis. `https://rss.detik.com/...`) |
| `tag_mapping` | Mapping tag RSS yang dibutuhkan untuk parsing — tag mana untuk judul, konten, gambar, tanggal, dll |
| `default_category_id` | Kategori default untuk hasil scraping dari sumber ini (opsional) |
| `is_registered` | Status pendaftaran oleh Usergod |

> **Catatan:** Isi `tag_mapping` adalah inti dari konfigurasi RSS. Usergod menentukan field RSS mana yang dipetakan ke field news, karena tiap portal punya struktur feed yang berbeda.

### 4. Source Config (Pengaturan Operasional Sumber)

Pengaturan operasional per sumber. Diatur oleh **Superadmin saja** (KAI Pusat), terpisah dari pendaftaran sumber oleh Usergod.

| Atribut | Keterangan |
|---|---|
| `source_id` | Sumber yang dikonfigurasi (FK ke News Source) |
| `schedule` | Frekuensi scraping (mis. tiap 1 jam, tiap 6 jam, harian) |
| `publish_mode` | `auto_publish` (langsung tayang) atau `draft` (masuk antrian review) |
| `is_enabled` | Aktif/nonaktif scraping tanpa menghapus konfigurasi |

> **Pemisahan penting:** Usergod = "ini cara baca Detik.com (URL + tag)". Superadmin = "scrape Detik.com tiap 3 jam, hasilnya masuk draft". Admin Region **tidak** punya akses ke konfigurasi ini sama sekali.

---

## Cara News Masuk ke Sistem

Ada dua jalur, dan jalur ini menentukan banyak aturan turunannya.

### Jalur 1 — Manual

Superadmin atau Admin Region mengetik berita langsung di backoffice. News bertanda `origin = manual`, dengan `created_by` terisi. Label asal mengikuti pembuatnya (KAI Pusat atau region tersebut).

### Jalur 2 — Scraping Otomatis (KAI Pusat Saja)

Sistem mengambil berita dari sumber yang sudah dikonfigurasi, sesuai jadwal yang diatur Superadmin. News bertanda `origin = scraped`, dengan `source_id` terisi dan label asal `KAI Pusat`. Hasil scraping masuk ke `draft` atau langsung `published` tergantung `publish_mode` sumbernya. **Scraping eksklusif milik KAI Pusat** — region tidak punya jalur ini.

---

## Siapa Bisa Apa

### Usergod

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Daftarkan & konfigurasi news source (URL + tag RSS) | ✅ Ya | Termasuk mapping tag untuk parsing |
| Edit/hapus source | ✅ Ya | Definisi teknis sumber |
| Semua aksi Superadmin | ✅ Ya | Akses penuh |

### Superadmin (KAI Pusat)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Konfigurasi scheduling per source (jadwal, mode publish) | ✅ Ya | **Eksklusif** — hanya KAI Pusat |
| Buat news manual | ✅ Ya | Label asal = KAI Pusat |
| Review & publish hasil scraping (draft) | ✅ Ya | Termasuk edit sebelum tayang |
| Approve / reject news (jika approval aktif) | ✅ Ya | Termasuk news dari Admin Region |
| Edit/hapus news apapun | ✅ Ya | Semua news |
| Kelola kategori news | ✅ Ya | CRUD, urutan, aktif/nonaktif |
| Toggle setting approval global | ✅ Ya | On/off apakah news perlu approval |
| Daftarkan source baru | ❌ Tidak | Hanya Usergod |

### Admin Region (mis. Admin Jakarta)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Buat news manual | ✅ Ya | Label asal = region-nya (mis. KAI Jakarta) |
| Submit news untuk approval | ✅ Ya | Jika setting approval aktif |
| Edit/hapus news milik sendiri | ✅ Ya | Sebelum tayang / sesuai status |
| Approval news | ⚠️ Lihat catatan | Approval ditangani Superadmin (lihat bagian Setting Approval) |
| Konfigurasi source / scheduling | ❌ Tidak | **Eksklusif KAI Pusat** |
| Scraping | ❌ Tidak | Region hanya manual posting |
| Kelola kategori | ❌ Tidak | Hanya Superadmin |

### Member (Standard & Pro)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Lihat semua news published | ✅ Ya | Global — semua news, apapun asalnya |
| Save/bookmark news | ✅ Ya | Untuk dibaca lagi nanti |
| Reaksi (like, dll) | ✅ Ya | Sesuai fitur reaksi yang tersedia |
| Share news | ✅ Ya | Ke WhatsApp, sosmed, copy link |
| Buat / edit / hapus news | ❌ Tidak | Member adalah konsumen, bukan kontributor |

> **Catatan:** Membuat news **tidak** di-gate oleh subscription plan (Pro/Standard) — member sama sekali tidak bisa membuat news terlepas dari plan-nya. Pembuatan news murni dikontrol oleh role (Superadmin / Admin Region).

### Guest (belum login)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Lihat news published | ✅ Terbatas | Hanya konten publik (jika diaktifkan) |
| Save / reaksi / share | ❌ Tidak | Harus login |

---

## Visibility: Tidak Scoped + Label Asal

Ini perbedaan kunci News dengan modul Announcement.

- **Announcement** punya scope global/regional → menentukan **siapa yang bisa lihat**.
- **News** tidak punya scope → **semua news bisa dilihat semua orang**, global.

Yang membedakan news hanyalah **label asal** (`author_label`):

- News dari Superadmin → label **KAI Pusat**
- News dari Admin Jakarta → label **KAI Jakarta**
- News hasil scraping → label **KAI Pusat** (karena scraping milik pusat)

Fungsi label asal:

- **Informatif** — user tahu berita ini dari pusat atau region tertentu.
- **Filter opsional di UI** — user *boleh* memfilter "tampilkan news dari KAI Jakarta", tapi ini bukan pembatasan akses, hanya kenyamanan browsing.
- **Bukan pembatas** — user di region manapun tetap bisa melihat news dari region lain.

---

## Kategori News

Modul News memiliki sistem kategori seperti portal berita pada umumnya (mis. Politik, Ekonomi, Olahraga, Teknologi, Hiburan, dll).

- Kategori dikelola oleh **Superadmin** — bisa tambah, edit, nonaktifkan, dan atur urutan tampil.
- Setiap news wajib punya satu kategori.
- Hasil scraping bisa diberi kategori default per sumber, atau dikoreksi manual saat review.
- Member bisa memfilter dan browse news berdasarkan kategori.

---

## Fitur Scraping & Scheduling

Inti dari otomasi News. **Eksklusif milik KAI Pusat** — Admin Region tidak terlibat sama sekali. Dipisah menjadi dua tingkat tanggung jawab.

### Yang Diatur Usergod — Definisi Sumber (Level Teknis)

Usergod mendaftarkan sumber dan menentukan cara membacanya:

1. **Tambah source** — mis. `Detik`
2. **Input feed URL** — URL RSS sumber tersebut
3. **Input tag mapping** — petakan tag RSS ke field news (judul, konten, gambar, tanggal, dll)

Konfigurasi ini sifatnya sekali setup dan jarang berubah — ibarat "kamus" cara membaca tiap sumber. Karena tiap portal punya struktur feed berbeda, mapping ini harus didefinisikan per sumber.

### Yang Diatur Superadmin (KAI Pusat) — Operasional Sumber

Untuk tiap source yang sudah didaftarkan Usergod, **Superadmin saja** yang mengatur:

1. **Schedule** — seberapa sering sistem melakukan scraping (mis. tiap 1 jam, 6 jam, harian)
2. **Publish mode** — pilih salah satu:
   - `auto_publish` — hasil scraping langsung tayang tanpa review
   - `draft` — hasil scraping masuk antrian draft untuk direview manual dulu
3. **Enable/disable** — pause scraping suatu sumber tanpa menghapus konfigurasinya

> **Penting:** Admin Region **tidak** punya akses ke scheduling/scraping. Region hanya manual posting + approval.

### Alur Scraping

```
Source (didaftarkan Usergod)
        ↓
Source Config (jadwal & mode oleh Superadmin / KAI Pusat)
        ↓
[Scheduler jalan sesuai jadwal]
        ↓
Ambil & parsing feed RSS (pakai tag mapping)
        ↓
   ┌────────────────┬────────────────┐
   │ publish_mode = │ publish_mode = │
   │  auto_publish  │     draft      │
   ↓                ↓
PUBLISHED        DRAFT → review admin → PUBLISHED / REJECTED
```

> Mode publish hasil scraping diatur **per source** (di Source Config), terpisah dari setting approval global untuk news manual.

---

## Setting Approval

Ada konfigurasi global yang menentukan apakah news manual perlu approval sebelum tayang atau bisa langsung publish.

### Setting Approval Global (News Manual)

- **Level:** Global — satu setting berlaku untuk semua news manual di platform.
- **Diatur oleh:** Superadmin.
- **Dua mode:**
  - **Approval ON** — news manual (dari Admin Region maupun Superadmin) masuk status `pending_approval` dulu, harus di-approve sebelum `published`.
  - **Approval OFF** — news manual langsung `published` saat di-submit.

### Mode Publish Scraping (Per Source)

Terpisah dari setting approval global di atas. Hasil scraping punya konfigurasinya sendiri **per source** (`auto_publish` atau `draft`), diatur di Source Config oleh Superadmin. Jadi:

- News manual → dikontrol **setting approval global**
- News scraping → dikontrol **publish_mode per source**

### Catatan tentang Approval

Karena member tidak bisa post news, approval hanya relevan untuk news yang dibuat admin. Saat approval ON, alur approval ditangani Superadmin (KAI Pusat) sebagai otoritas approval terpusat.

---

## Status & Lifecycle

```
┌──────────┐
│  DRAFT   │  Belum tayang (hasil scraping draft mode, atau manual disimpan dulu)
└────┬─────┘
     │ submit
     ↓
┌────────────────────┐
│ PENDING_APPROVAL   │  Menunggu approval (jika setting approval ON)
└────┬───────────────┘
     │ approve            │ reject
     ↓                    ↓
┌──────────────┐    ┌──────────────┐
│ PUBLISHED    │    │  REJECTED    │
└────┬─────────┘    └──────────────┘
     │ archive (manual atau expire)
     ↓
┌──────────────┐
│  ARCHIVED    │  Tidak muncul di listing utama, tetap bisa diakses via history
└──────────────┘
```

### Status Detail

**DRAFT** — Belum tayang. Dari scraping (mode draft) atau manual yang belum di-submit. Bisa diedit & dihapus.

**PENDING_APPROVAL** — Menunggu approval Superadmin. Hanya muncul jika setting approval global ON. Jika approval OFF, status ini dilewati.

**PUBLISHED** — Tayang ke semua user (global). Bisa diedit (koreksi typo) atau diarsipkan.

**REJECTED** — Ditolak admin saat review/approval. Tidak tayang.

**ARCHIVED** — Tidak muncul di listing utama, tetap tersimpan untuk arsip/history.

---

## Interaksi Member

Member adalah konsumen news. Yang bisa dilakukan member terhadap news yang sudah published:

- **Membaca** daftar dan detail news (semua news, global)
- **Menyimpan / bookmark** news untuk dibaca lagi nanti
- **Memberi reaksi** (like, dll sesuai fitur yang tersedia)
- **Membagikan** news ke WhatsApp, media sosial, atau copy link
- **Memfilter** news berdasarkan kategori atau label asal (opsional)
- **Menerima notifikasi** news baru (sesuai preferensi: dari KAI Pusat, KAI Region — lihat modul Notification Preferences)

Member **tidak bisa** membuat, mengedit, atau menghapus news.

---

## Multi-Language / Translation (Open)

> **Status: Belum diputuskan.** Bagian ini didokumentasikan sebagai pertimbangan desain, belum jadi aturan final.

### Konteks Masalah

News ditargetkan multi-language, namun konten masuk dalam Bahasa Indonesia — baik dari scraping sumber lokal (mis. Detik) maupun input manual admin. Tantangannya: bagaimana menyediakan versi bahasa lain (mis. EN, KO) secara konsisten?

### Opsi Pendekatan

**Opsi A — Translate on-the-fly.** Simpan konten 1 bahasa (ID), terjemahkan saat dibaca. Hemat storage, tapi kualitas inkonsisten dan berat di performa.

**Opsi B — Simpan multi-language di DB.** Tiap news punya versi ID/EN/KO tersimpan. Kualitas terkontrol & bisa di-cache, tapi butuh proses pengisian versi non-ID.

**Opsi C — Hybrid (rekomendasi awal).** Simpan original (ID) sebagai sumber kebenaran. Translation digenerate sekali saat publish (bukan tiap baca), lalu disimpan dan bisa dikoreksi manual sebelum tayang. Paling cocok untuk integrasi AI ke depan.

### Rekomendasi Sementara

Untuk fase sekarang, siapkan struktur data yang **ready untuk multi-language** (mengikuti Opsi C) tetapi **isi Bahasa Indonesia dulu**, dengan field translation bersifat nullable. Dengan begitu, saat integrasi AI translation dilakukan nanti, tidak perlu migrasi schema besar — cukup isi field yang sudah disiapkan.

### Rencana Integrasi AI (Future)

Ke depan, translation kemungkinan ditangani AI: saat news akan dipublish → trigger AI translate → simpan hasil → admin bisa review/koreksi sebelum tayang. Belum masuk scope fase ini.

---

## Use Cases

### Use Case 1 — Superadmin Buat News Manual (Approval OFF)

Superadmin membuka backoffice News, klik "Buat News", mengisi judul, konten, memilih kategori "Ekonomi", mengunggah cover image, lalu klik Publish. Karena setting approval OFF, news langsung tayang dengan label asal "KAI Pusat" dan bisa dilihat semua user di platform.

### Use Case 2 — Admin Region Buat News (Approval ON)

Admin Jakarta membuat news tentang acara lokal Korea-Indonesia di Jakarta. Karena setting approval ON, news masuk status `pending_approval`. Superadmin menerima notifikasi, mereview, lalu approve. News tayang dengan label asal "KAI Jakarta" — dan tetap bisa dilihat oleh user di seluruh platform (tidak terbatas Jakarta).

### Use Case 3 — Usergod Daftarkan Sumber Baru

Usergod menambahkan sumber "Detik", memasukkan URL feed RSS-nya, lalu memetakan tag RSS: `<title>` → judul, `<description>` → excerpt, `<enclosure>` → gambar, `<pubDate>` → tanggal. Sumber tersimpan dan siap dikonfigurasi Superadmin.

### Use Case 4 — Superadmin Atur Scheduling Sumber

Superadmin membuka konfigurasi sumber "Detik", set schedule tiap 3 jam, dan memilih mode `draft` agar hasil scraping direview dulu. Sistem mulai mengambil berita tiap 3 jam ke antrian draft. Admin Region tidak terlibat dalam proses ini.

### Use Case 5 — Review Hasil Scraping

Tiap beberapa jam, berita baru dari Detik masuk ke draft. Superadmin membuka antrian, mengoreksi kategori bila perlu, lalu publish berita yang layak dan reject yang tidak relevan.

### Use Case 6 — Sumber dengan Auto-Publish

Untuk sumber tepercaya, Superadmin set mode `auto_publish`. Hasil scraping langsung tayang tanpa review manual, cocok untuk feed resmi yang sudah terkurasi.

### Use Case 7 — Member Baca & Simpan News

Member di Surabaya membuka tab News, memfilter kategori "Olahraga", membaca artikel dari KAI Pusat sekaligus artikel berlabel "KAI Jakarta" (karena news global), lalu mengetuk bookmark untuk menyimpan dan membagikan ke grup WhatsApp.

---

## Ringkasan Aturan

| Aturan | Detail |
|--------|--------|
| **Siapa yang bisa buat news** | Superadmin (KAI Pusat) dan Admin Region (manual) |
| **Member buat news** | ❌ Tidak bisa — member hanya konsumen |
| **Scraping / scheduling** | Superadmin (KAI Pusat) saja — eksklusif |
| **Daftarkan & config source RSS** | Usergod saja |
| **Admin Region** | Hanya manual posting + (di-)approval |
| **Visibility news** | Tidak scoped — semua news global, bisa dilihat semua user |
| **Field region di news** | Hanya label asal (siapa yang posting), bukan pembatas akses |
| **Kelola kategori** | Superadmin saja |
| **Setting approval (news manual)** | Global, diatur Superadmin — on/off |
| **Mode publish scraping** | Per source (auto_publish / draft), diatur Superadmin |
| **Status news** | draft → pending_approval → published → archived (+ rejected) |
| **Interaksi member** | Lihat, save/bookmark, reaksi, share, filter |
| **Multi-language** | Belum final — struktur disiapkan ready (Opsi C), isi ID dulu |
| **Integrasi AI translation** | Future, belum masuk scope |

---

## Keputusan yang Masih Terbuka

1. **Bahasa yang didukung** — ID wajib; perlu konfirmasi apakah EN dan/atau KO.
2. **Translation scraping** — otomatis (AI) atau tetap ID saja sampai ada yang translate manual?
3. **Translation input manual** — admin wajib isi semua bahasa, atau isi 1 bahasa + bantuan sistem?
4. **Translation blocking publish?** — apakah news boleh tayang ID-only dulu, versi bahasa lain menyusul?
5. **Reaksi member** — jenis reaksi apa saja yang tersedia? (like saja, atau lebih?)
6. **De-duplikasi scraping** — bagaimana mencegah berita yang sama ter-scrape dua kali dari sumber yang sama?
7. **Edit setelah approve** — apakah edit news yang sudah published perlu re-approval, atau langsung berlaku?

---

*Dokumen rules News module — non-teknis. Untuk detail API lihat API_SPEC_NEWS_MOBILE & API_SPEC_NEWS_BACKOFFICE (menyusul). Untuk skema database lihat NEWS_DB_SCHEMA (menyusul).*
