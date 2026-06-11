# Ads Module — Rules & Use Cases

Dokumen ini menjelaskan aturan bisnis modul Ads di aplikasi KAI. Ads adalah sistem iklan berbayar yang memungkinkan Superadmin (KAI Pusat) dan member dengan benefit `post_ads` untuk memasang iklan yang tampil di seluruh pengguna aplikasi KAI.

---

## 1. APA ITU MODUL ADS?

**Ads** adalah sistem advertising internal platform KAI. Iklan ditampilkan kepada semua pengguna aplikasi tanpa targeting khusus — siapapun yang membuka app berpotensi melihat iklan.

### Karakteristik:
- **Creator:** Superadmin (KAI Pusat) atau Member dengan benefit `post_ads`
- **Target:** Semua user yang membuka aplikasi KAI (tanpa targeting)
- **Purpose:** Monetisasi platform KAI — member Pro membayar untuk exposure iklan mereka
- **Approval:** Dikontrol via Ad Setting di backoffice (bisa auto-publish atau require review)
- **Tampil di:** Home screen (slider banner + native feed) dan halaman khusus "Iklan & Promo"

### Berbeda dari:
- **Announcement:** Komunikasi satu arah dari admin untuk broadcast informasi penting/darurat, bukan iklan komersial
- **News:** Konten editorial, bukan sponsored content
- **Event:** Event dengan registrasi, bukan iklan
- **Directory:** Listing bisnis permanen, bukan slot iklan berbatas waktu

---

## 2. SIAPA YANG BISA MEMBUAT ADS?

### Superadmin / KAI Pusat

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Membuat ads semua tipe | ✅ Ya | Tanpa batas jumlah, tanpa approval |
| Langsung publish | ✅ Ya | Tidak perlu review — langsung aktif |
| Moderasi ads dari member | ✅ Ya | Approve, reject, pause, atau hapus ads milik member |
| Konfigurasi Ad Setting | ✅ Ya | Atur approval mode, max durasi, max ads per member |
| Melihat analytics semua ads | ✅ Ya | Impressi, klik, CTR seluruh ads di platform |
| Edit / hapus ads milik sendiri | ✅ Ya | Kapan saja |
| Edit / hapus ads milik member | ✅ Ya | Sebagai moderator |

> Ads yang dibuat Superadmin mewakili KAI Pusat — untuk promosi event, promo tiket, atau kampanye resmi KAI. Tidak perlu approval karena Superadmin adalah pemegang otoritas tertinggi di platform.

### Member dengan Benefit `post_ads`

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Membuat ads | ✅ Ya* | *Hanya jika memiliki benefit `post_ads` (Pro plan) |
| Langsung publish | ✅ / ❌ | Tergantung Ad Setting — bisa auto atau require review |
| Moderasi ads orang lain | ❌ Tidak | Hanya bisa kelola ads milik sendiri |
| Melihat analytics ads sendiri | ✅ Ya | Hanya data ads milik sendiri |
| Edit ads sendiri | ✅ Ya* | *Hanya jika status masih `draft` atau `pending`. Jika sudah `active` tidak bisa edit — harus pause dulu |
| Hapus ads sendiri | ✅ Ya | Kapan saja, untuk semua status |

> **Benefit `post_ads`** harus di-assign ke plan oleh Superadmin di backoffice subscription. Default: hanya aktif di Pro plan. Superadmin bisa mengaktifkan atau menonaktifkan benefit ini kapan saja tanpa perlu deploy ulang.

### Member tanpa Benefit `post_ads`

Tidak bisa membuat ads. Di UI mobile, tombol "Pasang Iklan" akan menampilkan pesan ajakan upgrade ke Pro.

### Guest (belum login)

Tidak bisa membuat ads. Bisa melihat ads yang tampil di home dan halaman "Iklan & Promo".

---

## 3. TIPE ADS

Ada 4 tipe ads yang didukung. Creator memilih tipe saat membuat ads — setiap tipe punya field konten yang berbeda.

### Tipe 1: Image Banner

Iklan berupa gambar statis atau GIF animasi.

**Field konten:**
- `title` — judul internal untuk identifikasi (tidak tampil di mobile, max 100 char)
- `image_url` — URL gambar (wajib)
- `click_url` — URL tujuan saat user tap iklan (wajib)
- `description` — deskripsi opsional untuk backoffice reference

**Ketentuan media:**
- Format: JPG, PNG, GIF
- Ukuran maksimum: 2MB
- Rasio: 16:9 (slider banner) atau 4:1 (banner tipis)
- Resolusi minimum: 1200×675px untuk 16:9

**Cocok untuk placement:** Home slider banner, halaman Iklan & Promo (slider + list)

---

### Tipe 2: Video Banner

Iklan berupa video pendek yang autoplay.

**Field konten:**
- `title` — judul internal (max 100 char)
- `video_url` — URL file video (wajib)
- `thumbnail_url` — gambar preview sebelum video play (wajib)
- `click_url` — URL tujuan saat user tap (wajib)

**Ketentuan media:**
- Format: MP4
- Durasi maksimum: 30 detik
- Ukuran maksimum: 20MB
- Autoplay dengan muted — user tap untuk unmute

**Cocok untuk placement:** Home slider banner (selip di antara image banner), halaman Iklan & Promo

---

### Tipe 3: Text Ad

Iklan berbasis teks dengan CTA button — ringan di bandwidth, cocok untuk promo singkat.

**Field konten:**
- `headline` — teks utama iklan (wajib, max 60 char)
- `body_text` — teks pendukung (wajib, max 150 char)
- `cta_label` — label tombol CTA (wajib, max 20 char, contoh: "Beli Sekarang", "Daftar Gratis")
- `click_url` — URL tujuan saat user tap CTA (wajib)
- `icon_url` — ikon kecil opsional (maks 200×200px, PNG)

**Cocok untuk placement:** Native selip di home feed, halaman Iklan & Promo (list)

---

### Tipe 4: Native Ad

Iklan yang tampilannya menyerupai postingan konten biasa di feed — label "Iklan" kecil tetap muncul agar transparan kepada user.

**Field konten:**
- `title` — judul seperti judul artikel/postingan (wajib, max 100 char)
- `body_text` — isi singkat (wajib, max 200 char)
- `image_url` — gambar thumbnail opsional
- `click_url` — URL tujuan saat user tap (wajib)
- `sponsor_name` — nama pengiklan yang tampil (wajib, max 50 char)
- `sponsor_logo_url` — logo pengiklan kecil (opsional)

**Cocok untuk placement:** Selip di home feed (paling natural), halaman Iklan & Promo (list)

---

## 4. PLACEMENT — DI MANA ADS TAMPIL

### Placement 1: Home Slider Banner

Carousel swipe di bagian atas home screen, tepat di bawah top navigation bar.

**Karakteristik:**
- Menampilkan 1–5 ads aktif secara bergantian (bisa swipe manual atau auto-scroll)
- Setiap slide menampilkan label kecil "Iklan" di sudut atas
- Tipe yang didukung: Image Banner, Video Banner
- Urutan tampil: Ads Superadmin/KAI Pusat diprioritaskan lebih dulu, lalu ads member diurutkan berdasarkan `start_date` terbaru

**Auto-scroll:**
- Interval: setiap 5 detik (tidak auto-scroll saat user sedang swipe atau video sedang play)

---

### Placement 2: Native Selip di Home Feed

Ads muncul menyisip di antara konten feed (news, event, komunitas) di home screen.

**Karakteristik:**
- Muncul setiap N item konten (N dikonfigurasi di Ad Setting, default: setiap 5 item)
- Diberi highlight subtle (border kiri tipis) dan label "Iklan" kecil
- Tipe yang didukung: Native Ad, Text Ad
- Jika tidak ada ads aktif yang cukup, slot tetap kosong (tidak dipaksa isi)

---

### Placement 3: Halaman "Iklan & Promo"

Halaman dedicated yang bisa diakses dari bottom navigation bar (tab "Ads" / "Promo").

**Struktur halaman:**
1. **Slider banner** di bagian atas — sama seperti home, menampilkan Image Banner dan Video Banner
2. **List semua ads aktif** di bawah slider — menampilkan semua tipe (image, video, text, native) dalam format card list

**Karakteristik list:**
- Setiap item menampilkan: thumbnail/icon, badge tipe ads, judul, nama pengiklan, tanggal aktif, dan CTA link
- Diurutkan: Ads KAI Pusat paling atas, sisanya berdasarkan `start_date` terbaru
- User bisa scroll untuk lihat semua ads yang sedang aktif

---

### Matriks Tipe × Placement

| Tipe | Home slider | Home feed (native) | Halaman Iklan & Promo |
|------|------------|-------------------|----------------------|
| Image banner | ✅ | — | ✅ (slider + list) |
| Video banner | ✅ | — | ✅ (slider + list) |
| Text ad | — | ✅ | ✅ (list) |
| Native ad | — | ✅ | ✅ (list) |

---

## 5. AD SETTING (KONFIGURASI BACKOFFICE)

Superadmin bisa mengatur global setting untuk ads melalui backoffice. Setting ini berlaku untuk semua ads dari member (tidak berlaku untuk ads Superadmin sendiri).

### Setting yang tersedia:

| Setting | Tipe | Default | Keterangan |
|---------|------|---------|------------|
| `approval_mode` | enum | `require_review` | `auto_publish` = langsung aktif, `require_review` = masuk queue dulu |
| `max_active_ads_per_member` | integer | `3` | Batas jumlah ads berstatus `active` yang bisa dimiliki satu member sekaligus |
| `max_duration_days` | integer | `30` | Batas maksimum hari ads aktif — setelah lewat, otomatis expired |
| `feed_ads_interval` | integer | `5` | Setiap berapa item konten feed, ada 1 native/text ad yang muncul |
| `slider_max_items` | integer | `5` | Jumlah maksimum ads yang tampil di slider banner |

> Perubahan setting berlaku untuk ads **baru** yang dibuat setelah perubahan. Ads yang sudah aktif tidak terpengaruh perubahan setting (kecuali superadmin pause/hapus manual).

---

## 6. APPROVAL FLOW

### Jika `approval_mode = auto_publish`

```
Member buat ads → submit → status: active → langsung tampil di app
```

Superadmin tetap bisa moderasi (pause / reject / hapus) kapan saja setelah ads aktif.

---

### Jika `approval_mode = require_review`

```
Member buat ads → submit → status: pending
     ↓
Superadmin menerima notifikasi "Ada ads baru menunggu review"
     ↓
Superadmin review konten ads
     ↓
  [Approve] → status: active → tampil di app
  [Reject]  → status: rejected → member dapat notifikasi + alasan penolakan
```

**Catatan review:**
- Superadmin wajib mengisi alasan jika menolak (reject reason), dikirim ke member via notifikasi in-app
- Tidak ada batas waktu review — ads tetap di status `pending` sampai ditindak

---

## 7. STATUS LIFECYCLE ADS

```
draft → pending → active → expired
                ↘ rejected
         active → paused → active (jika di-resume)
         active → paused → expired (jika habis durasi saat paused)
```

| Status | Deskripsi | Siapa yang bisa set |
|--------|-----------|---------------------|
| `draft` | Dibuat tapi belum disubmit | Creator (belum submit) |
| `pending` | Menunggu review superadmin | Otomatis saat submit (jika require_review) |
| `active` | Sedang tayang di app | Otomatis (auto_publish / setelah approve) |
| `rejected` | Ditolak superadmin | Superadmin |
| `paused` | Dihentikan sementara | Superadmin atau creator sendiri |
| `expired` | Habis masa tayang (end_date tercapai) | Otomatis oleh sistem |

---

## 8. FIELD UMUM (SEMUA TIPE)

Selain field konten yang spesifik per tipe, setiap ads wajib punya field berikut:

| Field | Tipe | Wajib | Keterangan |
|-------|------|-------|------------|
| `title` | string | ✅ | Judul internal untuk identifikasi di backoffice (max 100 char) |
| `ad_type` | enum | ✅ | `image_banner`, `video_banner`, `text_ad`, `native_ad` |
| `click_url` | string | ✅ | URL tujuan saat user tap iklan (bisa eksternal atau deep link in-app) |
| `start_date` | date | ✅ | Tanggal mulai tayang |
| `end_date` | date | ✅ | Tanggal selesai tayang (max dari start_date + `max_duration_days`) |
| `status` | enum | ✅ | `draft`, `pending`, `active`, `rejected`, `paused`, `expired` |
| `created_by` | uuid | ✅ | User ID creator (superadmin atau member) |
| `reject_reason` | string | ❌ | Wajib diisi superadmin saat status → `rejected` |
| `notes` | string | ❌ | Catatan internal creator (tidak tampil di mobile) |

---

## 9. ANALYTICS & TRACKING

Setiap ads otomatis dicatat metrik berikut:

| Metrik | Deskripsi |
|--------|-----------|
| `impressions` | Jumlah kali ads ditampilkan ke user (per unique session) |
| `clicks` | Jumlah kali user tap/klik iklan |
| `ctr` | Click-through rate = clicks / impressions × 100% |

**Siapa yang bisa lihat:**
- Superadmin → analytics semua ads di platform (aggregate + per ads)
- Member creator → analytics ads milik sendiri saja

**Catatan:**
- Satu user yang melihat ads yang sama dalam satu session dihitung 1 impression (tidak double count dalam session yang sama)
- Analytics di-reset ke 0 jika ads di-pause lalu di-resume (fresh counting)

---

## 10. ATURAN BISNIS & EDGE CASES

### Member mencapai batas `max_active_ads_per_member`
Jika member sudah punya jumlah ads aktif sesuai batas setting, tombol "Buat Ads Baru" di-disable. Muncul pesan: "Kamu sudah mencapai batas maksimum X iklan aktif. Selesaikan atau hapus iklan lama untuk membuat yang baru."

### Member downgrade dari Pro ke Standard
Jika member kehilangan benefit `post_ads` karena downgrade atau subscription expired:
- Ads yang sedang `active` tetap tayang sampai `end_date` — tidak langsung dihapus
- Member tidak bisa membuat ads baru
- Member tidak bisa mengedit atau resume ads yang di-pause

### Ads expired otomatis
Sistem cron job cek setiap hari — ads dengan `end_date` < hari ini dan status `active` atau `paused` diubah otomatis ke `expired`. Member dan Superadmin tidak perlu aksi manual.

### Ads Superadmin tidak kena batasan setting
`max_active_ads_per_member` dan `max_duration_days` hanya berlaku untuk ads dari member. Superadmin bisa buat ads tanpa batas jumlah dan durasi.

### Tidak ada ads aktif
Jika tidak ada ads aktif sama sekali:
- Slider banner di home tidak muncul (section disembunyikan)
- Slot native feed tidak diisi (konten tetap normal tanpa gap)
- Halaman "Iklan & Promo" tetap bisa dibuka tapi tampilkan state kosong: "Belum ada iklan aktif saat ini"

### Click URL validasi
- Harus berformat URL valid (http/https) atau deep link in-app (format: `kai://...`)
- Superadmin bisa akses URL apapun
- Member hanya bisa input URL eksternal — deep link in-app hanya bisa diset oleh Superadmin

---

## 11. USER EXPERIENCE MOBILE

### Home Screen
1. User buka app → home load
2. Slider banner muncul di bagian atas (jika ada ads aktif tipe image/video)
3. User scroll feed — setiap N item konten, muncul 1 native/text ad
4. Semua ads memiliki label "Iklan" yang jelas
5. User tap ads → buka `click_url` di in-app browser atau deep link

### Halaman "Iklan & Promo"
1. User tap tab "Ads" / "Promo" di bottom navigation
2. Slider banner tampil di atas (sama seperti home)
3. Di bawah slider: list semua ads aktif, diurutkan KAI Pusat duluan
4. Setiap item: thumbnail, badge tipe, judul, sponsor, tanggal aktif, tombol CTA
5. User tap item → buka `click_url`

### Member yang ingin pasang iklan (mobile)
1. User buka halaman "Iklan & Promo"
2. Tap tombol "Pasang Iklanmu" (muncul jika punya benefit `post_ads`)
3. Pilih tipe ads → isi form konten → set jadwal tayang → submit
4. Jika `require_review`: muncul info "Iklanmu sedang direview oleh admin"
5. Jika `auto_publish`: muncul konfirmasi "Iklanmu sudah aktif!"
6. Member bisa pantau status & analytics ads dari halaman "Iklan Saya"

---

*Dokumen ini menjelaskan business rules dan use cases Ads Module KAI App.*
