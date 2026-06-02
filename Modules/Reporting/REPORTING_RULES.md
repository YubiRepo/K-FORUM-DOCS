# Reporting System — Rules & Use Cases

Dokumen ini menjelaskan aturan bisnis **sistem pelaporan** di K-Forum, yang terdiri dari dua subsistem terpisah namun berbagi pola UI:

1. **Content Reporting** — pelaporan konten/user yang melanggar (post, komentar, komunitas, Q&A, listing, event, user) untuk ditindak moderasi.
2. **Bug Reporting** — pelaporan masalah teknis aplikasi untuk ditindak tim dev/support.

Untuk detail teknis lihat `REPORTING_DB_SCHEMA.md` dan API spec mobile/backoffice.

---

## Daftar Isi

1. [Kenapa Dua Subsistem Terpisah](#kenapa-dua-subsistem-terpisah)
2. [Bagian A — Content Reporting](#bagian-a--content-reporting)
   - [Konsep Polymorphic Target](#konsep-polymorphic-target)
   - [Scope: Komunitas vs Global](#scope-komunitas-vs-global)
   - [Kategori Alasan](#kategori-alasan)
   - [Lifecycle Report](#lifecycle-report)
   - [Agregasi & Auto-Flag](#agregasi--auto-flag)
   - [Anti-Abuse](#anti-abuse)
   - [Siapa Bisa Apa (Content)](#siapa-bisa-apa-content)
   - [Alur Content Report](#alur-content-report)
3. [Bagian B — Bug Reporting](#bagian-b--bug-reporting)
   - [Field & Konteks Otomatis](#field--konteks-otomatis)
   - [Lifecycle Bug](#lifecycle-bug)
   - [Integrasi Issue Tracker](#integrasi-issue-tracker)
   - [Siapa Bisa Apa (Bug)](#siapa-bisa-apa-bug)
4. [Integrasi dengan Modul Lain](#integrasi-dengan-modul-lain)
5. [Keputusan yang Masih Terbuka](#keputusan-yang-masih-terbuka)

---

## Kenapa Dua Subsistem Terpisah

| Aspek | Content Report | Bug Report |
|---|---|---|
| Yang dilaporkan | Konten / user | Masalah teknis aplikasi |
| Penangan | Moderator / Superadmin | Tim dev / support |
| Tujuan | Penegakan kebijakan → moderasi | Perbaikan teknis → fix |
| Konteks utama | Alasan & kategori pelanggaran | Versi app, device, langkah reproduksi |
| Tindak lanjut | Hapus konten / ban user | Triage → fix → deploy |

Disimpan sebagai **dua tabel terpisah** (`content_reports`, `bug_reports`) agar field tetap bersih dan routing tidak campur. Di sisi mobile, keduanya boleh memakai **pola UI "Laporkan" yang konsisten**.

---

## Bagian A — Content Reporting

### Konsep Polymorphic Target

Satu sistem report menerima banyak jenis target tanpa tabel per jenis, lewat pasangan `reportable_type` + `reportable_id`:

| `reportable_type` | Scope default | Keterangan |
|---|---|---|
| `community_post` | Komunitas | Postingan di komunitas |
| `community_comment` | Komunitas | Komentar |
| `community` | Global | Komunitas itu sendiri |
| `qna_question` | Global | Pertanyaan Q&A |
| `qna_answer` | Global | Jawaban Q&A |
| `directory_listing` | Global | Listing direktori/merchant |
| `event` | Global | Event |
| `announcement` | Global | Pengumuman |
| `news_post` | Global | Konten News (saat modul News rilis) |
| `news_comment` | Global | Komentar News (menyusul) |
| `user` | Global | Profil/akun user |

### Scope: Komunitas vs Global

Setiap report punya **scope** yang menentukan siapa yang berhak menanganinya:

- **Komunitas** (`community_post`, `community_comment`): bisa ditangani **leader/moderator komunitas terkait** (permission `manage_reports`) ATAU superadmin. Report menyimpan `community_id` untuk routing.
- **Global** (sisanya): hanya **superadmin** lewat backoffice. `community_id` = `null`.

> Komunitas yang dilaporkan (`reportable_type = community`) bersifat **global** — karena menyangkut komunitas secara keseluruhan, bukan urusan internal yang dipegang leader-nya sendiri.

### Kategori Alasan

`reason` (enum) + `detail` (teks bebas opsional):

| Value | Keterangan |
|---|---|
| `spam` | Spam / promosi berlebihan |
| `harassment` | Pelecehan / perundungan |
| `hate_speech` | Ujaran kebencian / SARA |
| `sexual_content` | Konten seksual |
| `violence` | Kekerasan / ancaman |
| `misinformation` | Informasi menyesatkan |
| `scam` | Penipuan |
| `impersonation` | Pemalsuan identitas |
| `other` | Lainnya (wajib isi `detail`) |

### Lifecycle Report

```
pending ──ambil review──> reviewing ──selesai──> resolved
                                                    ├─ resolution = action_taken
                                                    └─ resolution = dismissed
```

- `pending` — baru masuk, belum disentuh.
- `reviewing` — sedang ditinjau moderator/superadmin.
- `resolved` — selesai, dengan `resolution`:
  - `action_taken` — tindakan diambil (hapus konten / ban user).
  - `dismissed` — laporan ditolak (tidak melanggar).

### Agregasi & Auto-Flag

- **Dedup:** satu user maksimal **1 report aktif per target** (UNIQUE `reporter_id + reportable_type + reportable_id`).
- **Agregasi:** target menyimpan `report_count` denormalized untuk prioritisasi & sort `most_reported` (dipakai mis. Community backoffice). Counter di-increment saat report baru masuk.
- **Auto-flag:** bila `report_count` melewati ambang (default **5**, configurable), sistem menandai target untuk perhatian prioritas dan menotifikasi penangan. Auto-flag **tidak** otomatis menghapus konten — keputusan tetap di tangan manusia.

### Anti-Abuse

- **Rate-limit:** maksimal **N report/hari per user** (default 20).
- **False-report tracking:** simpan rasio report user yang berakhir `dismissed`; rasio tinggi → turunkan bobot / beri peringatan.
- Report tidak membuka identitas pelapor ke pemilik konten.

### Siapa Bisa Apa (Content)

| Aktor | Submit report | Lihat report | Proses (resolve) |
|---|---|---|---|
| Member (Standard/Pro) | ✅ Ya | ❌ Tidak | ❌ Tidak |
| Moderator/Leader komunitas | ✅ Ya | ✅ Hanya scope komunitasnya | ✅ Scope komunitasnya (`manage_reports`) |
| Superadmin | ✅ Ya | ✅ Semua | ✅ Semua + global |

### Alur Content Report

```
1. Member menemukan konten melanggar → tap "Laporkan"
2. Pilih reason + detail (opsional) → submit
3. Backend:
   - Validasi target ada & belum dilaporkan user ini
   - Insert content_reports (status=pending), set community_id jika scope komunitas
   - report_count += 1 di target
   - Jika report_count >= threshold → auto-flag + notify penangan
4. Penangan (moderator/superadmin):
   - List antrian report (filter scope/status) → ambil (reviewing)
   - Putuskan: action_taken (panggil aksi moderasi yg sudah ada) ATAU dismissed
   - Tulis resolution_note → status=resolved
5. (Opsional) Notify pelapor bahwa laporannya selesai
```

> **Penting:** report adalah **antrian + audit**, bukan aksi. Saat `action_taken`, sistem memanggil endpoint moderasi yang sudah ada (mis. remove post / ban user di Community backoffice). Sistem report tidak menduplikasi logika moderasi.

---

## Bagian B — Bug Reporting

### Field & Konteks Otomatis

Field yang diisi **user**:

| Field | Wajib | Keterangan |
|---|---|---|
| `title` | ✅ | Ringkasan singkat |
| `description` | ✅ | Penjelasan masalah |
| `steps_to_reproduce` | ❌ | Langkah memunculkan bug |
| `category` | ✅ | `crash`/`ui`/`performance`/`data`/`auth`/`other` |
| `severity` | ✅ | `low`/`medium`/`high`/`critical` (persepsi user) |
| `attachments` | ❌ | Screenshot (maks mis. 5) |

Field **auto-captured** oleh client (tanpa input user):

| Field | Keterangan |
|---|---|
| `app_version` | Versi aplikasi |
| `platform` | `ios` / `android` / `web` |
| `os_version` | Versi OS |
| `device_model` | Model perangkat |
| `screen` | Route/layar saat report dibuat |

### Lifecycle Bug

```
new ──triage──> triaged ──> in_progress ──> resolved ──> closed
                   │                          │
                   └──> wont_fix ─────────────┘
```

- `new` — baru masuk.
- `triaged` — sudah diklasifikasi & diberi `priority`.
- `in_progress` — sedang dikerjakan (`assigned_to` diisi).
- `resolved` — perbaikan selesai.
- `wont_fix` — diputuskan tidak diperbaiki.
- `closed` — ditutup (final).

> `severity` = persepsi pelapor; `priority` = keputusan tim saat triage. Keduanya disimpan terpisah.

### Integrasi Issue Tracker

Fase 1 **standalone** (disimpan di DB, dikelola di backoffice). Disiapkan field:

- `external_issue_id` — ID issue di Linear/Jira (nullable).
- `external_issue_url` — link issue (nullable).

Fase 2: saat bug ditriage, sistem bisa membuat issue di Linear/Jira via API dan menyimpan ID/URL-nya, lalu sinkronkan status balik. Struktur sekarang sudah siap untuk itu tanpa migrasi besar.

### Siapa Bisa Apa (Bug)

| Aktor | Submit | Lihat | Triage/Resolve |
|---|---|---|---|
| Member (Standard/Pro) | ✅ Ya | ✅ Hanya miliknya | ❌ Tidak |
| Support/Dev (role khusus) | ✅ Ya | ✅ Semua | ✅ Ya (`manage_bug_reports`) |
| Superadmin | ✅ Ya | ✅ Semua | ✅ Ya |

---

## Integrasi dengan Modul Lain

| Kebutuhan | Modul | Cara pakai |
|---|---|---|
| Aksi saat `action_taken` (hapus/ban) | Community backoffice (& modul konten lain) | Report memanggil endpoint moderasi yang ada |
| Notifikasi (report resolved, auto-flag) | Notification | Report memancarkan event |
| Permission `manage_reports`, `manage_bug_reports` | Role-Permission | Tambah permission baru |
| Sort `most_reported` | Community (B1) & modul konten lain | Konsumsi `report_count` |
| Issue tracker (Fase 2) | Linear / Jira (Atlassian) | Via `external_issue_id`/`url` |

---

## Keputusan yang Masih Terbuka

| # | Topik | Asumsi Sementara |
|---|---|---|
| 1 | Ambang auto-flag | 5 report (configurable) |
| 2 | Rate-limit report/hari | 20 per user |
| 3 | Notifikasi ke pelapor saat resolved | Opsional, default aktif |
| 4 | Permission moderator komunitas memproses report | Pakai permission baru `manage_reports` (bukan reuse `moderate_posts`) |
| 5 | Role khusus bug report | Role `support` / `dev` di backoffice (perlu didefinisikan) |
| 6 | Auto-create issue tracker | Ditunda Fase 2 |

---

*Dokumen ini hasil breakdown awal sistem pelaporan. Skema database di `REPORTING_DB_SCHEMA.md`.*
