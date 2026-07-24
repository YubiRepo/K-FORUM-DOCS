# Plan + Report: Multi-language & Localization — Platform-wide (i18n/l10n)

> **Dokumen hidup** — di-update seiring progres implementasi (mengikuti
> konvensi `PLAN_TIMEZONE_WALLCLOCK_INPUTS.md`: status per phase diubah
> jadi DONE dengan tanggal begitu selesai, bukan dibiarkan statis).
>
> Melengkapi `../PLAN_I18N_L10N.md` (translasi pesan error API — sudah
> diimplementasikan, scope sempit, tidak diubah oleh dokumen ini). Latar
> belakang analisis & trade-off ada di
> [`BRAINSTORM_I18N_L10N_PLATFORM.md`](./BRAINSTORM_I18N_L10N_PLATFORM.md).
>
> **Catatan silang penting**: Phase 5 di `../PLAN_I18N_L10N.md`
> ("User Language Preference — Future") ternyata sudah diimplementasikan
> sebagai alur terpisah (`usersettings/helpers.go: resolveLanguageEffective`)
> yang tidak dipakai ulang oleh `middleware/locale.go`. Ini akar penyebab
> masalah yang diperbaiki di Phase 0 di bawah.

## Log Perubahan

- **2026-07-24 (lanjutan)**: 3 file plan-detail (Phase 0/3/4) **dihapus**.
  Upaya platform-wide ini ditangguhkan sementara — fokus dialihkan ke
  fitur scope-kecil yang berdiri sendiri: translasi master data News
  (kategori & scope), lihat
  [`NEWS_MASTER_DATA_TRANSLATIONS.md`](./NEWS_MASTER_DATA_TRANSLATIONS.md).
  Dokumen ini (dan brainstorm) tetap valid sebagai referensi kalau/ketika
  upaya platform-wide dilanjutkan lagi.
- **2026-07-24**: Dokumen dibuat & folder direorganisasi ke `plans/i18n&l10n/`.
  Plan detail dibuat untuk Phase 0, 3, 4 (file terpisah, lihat tabel
  status). Phase 1, 2, 5, 6 masih level ringkasan — akan dipecah jadi file
  sendiri kalau/ketika mulai dikerjakan.

## Status per Phase

| Phase | Deskripsi | Status | Plan Detail |
|---|---|---|---|
| 0 | Fix locale-resolution inconsistency & timezone-display bug | NOT STARTED | (ringkasan di bawah, belum dipecah) |
| 1 | Single source of truth untuk locale master data | NOT STARTED | (ringkasan di bawah, belum dipecah) |
| 2 | Central source string statis lintas platform | NOT STARTED | (ringkasan di bawah, belum dipecah) |
| 3 | Generalisasi translasi konten ke modul lain (pilot: Event) | NOT STARTED | (ringkasan di bawah, belum dipecah) |
| 4 | Redesain search indexing multi-locale | NOT STARTED | (ringkasan di bawah, belum dipecah) |
| 5 | Generalisasi UI manajemen translasi di backoffice | NOT STARTED | (ringkasan di bawah, belum dipecah) |
| 6 | Audit testing & regresi | NOT STARTED | (ringkasan di bawah, belum dipecah) |

> Lihat juga upaya paralel scope-kecil yang sedang berjalan:
> [`NEWS_MASTER_DATA_TRANSLATIONS.md`](./NEWS_MASTER_DATA_TRANSLATIONS.md)
> (translasi kategori & scope News, independen dari tracker platform-wide ini).

**Cara update dokumen ini**: begitu sebuah phase mulai/selesai dikerjakan,
ubah kolom Status (`NOT STARTED` → `IN PROGRESS` → `DONE (tanggal)`), dan
tambahkan entri baru di *Log Perubahan* di atas. Kalau Phase 1/2/5/6 mulai
dikerjakan dan butuh detail teknis sedalam Phase 0/3/4, pecah jadi file
`PHASE_N_....md` sendiri mengikuti pola yang sama, lalu update kolom "Plan
Detail" di tabel ini.

---

## Phase 0 — Perbaikan Bug & Inkonsistensi (quick win, risiko rendah)

Locale resolution punya 2 alur yang dipakai untuk hal berbeda (bukan cuma
beda gaya kode — konten News mobile ternyata sama sekali tidak memakai
preferensi bahasa tersimpan user), plus bug timezone-display di mobile
(pakai `.toLocal()` bukan timezone entity) dan backoffice (tidak ada
konversi sama sekali). *(Plan detail file terpisah pernah dibuat untuk
phase ini lalu dihapus saat scope ditangguhkan — lihat Log Perubahan di
atas. Kalau dilanjutkan lagi, tulis ulang detailnya sebagai file baru.)*

---

## Phase 1 — Single Source of Truth Locale Master Data

*(Ringkasan — belum dipecah jadi file detail)*

### 1.1 Hilangkan hardcoded `supportedLocales` di Go bundle

`internal/infrastructure/i18n/bundle.go` baca daftar locale didukung dari
`system_languages` (dengan cache in-memory, invalidate saat admin
ubah data), bukan hardcode array.

### 1.2 Pastikan kedua sisi (setelah Phase 0.1 disatukan) validasi ke sumber yang sama

Tabel `system_languages.is_active` jadi validator tunggal untuk locale valid.

---

## Phase 2 — Central Source String Statis Lintas Platform

*(Ringkasan — belum dipecah jadi file detail)*

### 2.1 Definisikan skema key & sumber sentral

Buat file sumber (mis. `K-FORUM-DOCS/i18n/keys.json` atau lokasi yang
disepakati) dengan skema `{"key.path": {"id": "...", "en": "...", "ko": "..."}}`,
mengikuti konvensi key yang sudah ada di `locales/id.json` (mis.
`err.auth.invalid_credentials`) supaya tidak perlu rename besar-besaran.

### 2.2 Bangun generator per platform

Script yang membaca sumber sentral dan menghasilkan:
- `k-forum-api/locales/{id,en,ko}.json` (format sudah sesuai)
- `k-forum-backoffice/app/i18n/locales/{en,id,ko}.ts`
- `k_forum/lib/l10n/app_{en,id,ko}.arb`

### 2.3 Migrasi bertahap per modul

Pindahkan string hardcoded backoffice (News settings, Event, Community,
Directory screens) ke sumber sentral satu modul per satu, tanpa mengubah UX.

### 2.4 CI drift-check

Job yang gagal kalau ada key hilang/berbeda antara sumber sentral dan hasil
generate di salah satu platform.

---

## Phase 3 — Generalisasi Translasi Konten ke Modul Lain

Pilot di modul Event: tambah `events.original_language` + tabel
`event_translations` (kontrak kolom identik `article_translations`),
ekstrak kontrak `TranslatableContentRepository` generik supaya reusable,
pilot HANYA jalur sync/on-demand (batch ditunda — `BatchTranslationJob`
News-shaped, generalisasinya berisiko regresi News kalau dipaksakan
sekarang). *(Plan detail file terpisah pernah dibuat untuk phase ini lalu
dihapus saat scope ditangguhkan — lihat Log Perubahan di atas.)*

Rollout ke modul lain (Community/Directory/QnA/Ads) — **[BUTUH KONFIRMASI
USER]** urutan prioritas.

> Catatan: fitur translasi master data News (kategori & scope) yang
> sedang dikerjakan di [`NEWS_MASTER_DATA_TRANSLATIONS.md`](./NEWS_MASTER_DATA_TRANSLATIONS.md)
> **bukan** bagian dari Phase 3 ini — itu translasi master data (JSONB,
> manual), bukan translasi konten dinamis modul lain (tabel row-based,
> LLM) yang dibahas Phase 3.

---

## Phase 4 — Redesain Search Indexing Multi-locale

7 index (satu per tipe konten) berbagi 1 mapping dengan field paralel per
locale (`title_id/title_en/...`) — cuma News yang pernah mengisinya, 6
index lain membawa field kosong hari ini juga (quick-win independen bisa
dibersihkan duluan). Desain baru: pisahkan kebutuhan **search**
(field dikelompokkan per rumpun bahasa/skrip, mis. `title_latin`/`title_ko`
— bukan per locale individual) dari kebutuhan **display** (field
`flattened` `translations` yang tidak ikut menyumbang field-count growth
walau locale bertambah). *(Plan detail file terpisah pernah dibuat untuk
phase ini lalu dihapus saat scope ditangguhkan — lihat Log Perubahan di
atas.)*

---

## Phase 5 — Generalisasi UI Manajemen Translasi Backoffice

*(Ringkasan — belum dipecah jadi file detail)*

### 5.1 Ekstrak komponen dari `ArticleForm.vue` (Translations tab)

Status badge per bahasa, tombol Auto-Translate, modal edit manual — jadikan
komponen reusable, dipakai modul lain begitu Phase 3 selesai untuk modul
itu.

### 5.2 **[BUTUH KONFIRMASI USER]** Lokalisasi penuh admin chrome

Belum diputuskan apakah backoffice perlu dilokalkan penuh (saat ini mayoritas
hardcoded English) atau cukup tetap English selamanya untuk chrome admin.
Kalau ya, jadi bagian dari migrasi Phase 2.3.

---

## Phase 6 — Testing & Audit Regresi

*(Ringkasan — belum dipecah jadi file detail)*

### 6.1 Audit test harness untuk gap "i18n bundle tidak ter-load"

Sudah pernah ditemukan di test Directory/Community (translasi diam-diam
tidak pernah teruji). Audit ulang test suite modul yang kena Phase 2/3 untuk
memastikan bundle/fixture locale benar-benar ter-load.

### 6.2 Test matrix per modul yang dapat translasi konten baru

(a) create dengan locale asli, (b) request locale lain sebelum ada
translasi → fallback ke original, (c) setelah translasi ada → tampil
translasi, (d) locale tidak didukung → fallback default sistem, (e)
search/listing menampilkan title sesuai locale request.

### 6.3 Regression test locale-resolution setelah Phase 0.1

Pastikan penyatuan alur tidak mengubah behavior mayoritas kasus existing
(user dengan preferensi tersimpan tetap dapat locale yang sama seperti
sebelumnya).

---

## Urutan Eksekusi Disarankan

Phase 0 (quick win, bug fix) → Phase 1 (fondasi locale master) → Phase 2
(string statis, bisa paralel dengan Phase 3) → Phase 3 (translasi konten,
mulai dari pilot Event) → Phase 4 (search, setelah Phase 3 pilot selesai
supaya desain index sudah mempertimbangkan modul baru) → Phase 5 (UI
backoffice, ikut progress Phase 3) → Phase 6 (berjalan terus di tiap phase,
bukan di akhir saja).

## Verifikasi (ringkasan lintas phase — detail per phase ada di masing-masing dokumen)

1. Phase 0: user dengan preferensi tersimpan "id" tapi `Accept-Language:
   en-US` harus konsisten dapat "id" di semua endpoint (error message
   maupun konten News/Event); mobile & backoffice tampilkan waktu event
   sesuai timezone entity, bukan device/tanpa konversi.
2. Phase 3 pilot: buat Event, minta translasi ke locale lain, verifikasi
   fallback & status translasi bekerja sama seperti News.
3. Phase 4: bandingkan hasil search sebelum/sesudah reindex untuk query yang
   sama di 3 locale — pastikan tidak ada regresi relevansi.
4. Phase 6.1: jalankan test suite modul yang disentuh, pastikan gagal kalau
   translasi sengaja dirusak (bukan cuma "tidak error").

## Pertanyaan Terbuka (belum dikonfirmasi user)

1. Urutan prioritas modul rollout Phase 3 (Community/Directory/QnA/Ads).
2. Currency: multi-currency per locale, atau satu currency + format angka locale-aware?
3. Central-source (Phase 2) cukup custom, atau ada preferensi vendor TMS?
4. Backoffice perlu dilokalkan penuh (Phase 5.2), atau tetap English-only untuk chrome admin selamanya?
5. Opsi UX timezone-display backoffice (Phase 0.3): tampilkan waktu asli event apa adanya, atau konversi ke timezone browser admin?
