# Reporting (Content Report & Bug Report) — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Reporting/REPORTING_RULES.md`.

## Ringkasan Domain

Reporting terdiri dari **dua subsistem terpisah** yang sengaja disimpan sebagai tabel berbeda (`content_reports` dan `bug_reports`) agar field tetap bersih dan routing penanganan tidak campur, meski di sisi mobile keduanya memakai **pola UI "Laporkan" yang konsisten**:

1. **Content Reporting** — melaporkan konten atau user yang melanggar kebijakan (post/comment komunitas, komunitas itu sendiri, pertanyaan/jawaban Q&A, listing direktori, event, announcement, konten/comment News, atau profil user) lewat konsep **polymorphic target** (`reportable_type` + `reportable_id`). Ditangani moderator komunitas (untuk target scope komunitas) atau Superadmin (untuk semua, termasuk scope global) → berujung tindakan moderasi (hapus konten/ban user) atau ditolak (`dismissed`).
2. **Bug Reporting** — melaporkan masalah teknis aplikasi, disertai konteks otomatis (versi app, device, OS, layar saat ini) plus deskripsi/langkah reproduksi dari user. Ditangani tim dev/support → berujung perbaikan teknis (fix/deploy) atau `wont_fix`.

Kedua subsistem ini **tidak dibatasi tier sama sekali** — Member Standard dan Pro punya hak yang identik untuk submit laporan di kedua jalur.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Submit content report (konten/user apa pun) | ✅ | ✅ | Sama untuk semua `reportable_type` |
| Lihat status content report miliknya sendiri | ❌ | ❌ | `REPORTING_RULES.md` eksplisit: Member "Lihat report ❌ Tidak" — member tidak diberi visibilitas atas laporan yang sudah disubmit |
| Proses/resolve content report | ❌ | ❌ | Eksklusif moderator komunitas (`manage_reports`, hanya scope komunitasnya) atau Superadmin (semua scope) |
| Submit bug report | ✅ | ✅ | |
| Lihat status bug report miliknya sendiri | ✅ | ✅ | Berbeda dari content report — di bug report, member eksplisit **bisa** lihat laporannya sendiri |
| Triage/resolve bug report | ❌ | ❌ | Eksklusif role support/dev (`manage_bug_reports`) atau Superadmin |

## Journey 1: Member Melaporkan Konten/User — 🅢🅟 — 📱 Mobile

1. Member menemukan konten atau user yang dianggap melanggar — bisa berupa post/comment komunitas, komunitas itu sendiri, pertanyaan/jawaban Q&A, listing direktori, event, announcement, konten/comment News, atau profil user lain — tap "Laporkan".
2. Pilih kategori `reason`: `spam`, `harassment`, `hate_speech`, `sexual_content`, `violence`, `misinformation`, `scam`, `impersonation`, atau `other` (wajib isi `detail` teks bebas jika memilih `other`).
3. Submit. Backend memvalidasi target benar-benar ada dan **belum pernah dilaporkan user yang sama** — satu user maksimal 1 report aktif per target (unique constraint `reporter_id + reportable_type + reportable_id`).
4. Report tersimpan dengan status `pending`; jika target berscope komunitas (post/comment komunitas), `community_id` diisi untuk routing ke moderator komunitas terkait. Untuk target selain itu (komunitas itu sendiri, Q&A, direktori, event, announcement, News, user), scope selalu global — hanya bisa ditangani Superadmin.
5. `report_count` pada target bertambah 1. Jika melewati ambang (default 5, configurable) → sistem **auto-flag** target untuk prioritas perhatian dan menotifikasi penangan — auto-flag **tidak** otomatis menghapus konten, keputusan tetap di tangan moderator/Superadmin.
6. Member **tidak melihat status lanjutan** dari laporannya sendiri di UI (tidak ada halaman "laporan saya" untuk content report) — dokumen sumber eksplisit menyebut member tidak diberi akses "Lihat report". Notifikasi ke pelapor saat laporan `resolved` bersifat opsional (default aktif, per catatan "Keputusan yang Masih Terbuka" di dokumen sumber).
7. **Anti-abuse**: member dibatasi maksimal N report per hari (default 20). Sistem juga melacak rasio laporan yang berakhir `dismissed` milik seorang user — rasio tinggi bisa menurunkan bobot laporan berikutnya atau memicu peringatan.
8. Identitas pelapor **tidak pernah dibuka** ke pemilik konten yang dilaporkan.

**Selesai:** report tersimpan dan masuk antrian penanganan; dari sisi member, alur berakhir di titik submit (tanpa visibilitas lanjutan atas keputusan, kecuali notifikasi opsional saat resolved).

## Journey 2: Member Melaporkan Bug Teknis — 🅢🅟 — 📱 Mobile

1. Member membuka form bug report (dokumen sumber tidak mendetailkan entry point spesifik di navigasi mobile — dicatat sebagai gap, bukan diasumsikan lokasinya di menu tertentu).
2. Sistem otomatis menangkap konteks tanpa input user: `app_version`, `platform` (ios/android/web), `os_version`, `device_model`, dan `screen` (route/layar saat laporan dibuat).
3. Member mengisi field wajib: `title` (ringkasan singkat), `description` (penjelasan masalah), `category` (`crash`/`ui`/`performance`/`data`/`auth`/`other`), `severity` (`low`/`medium`/`high`/`critical` — persepsi member sendiri, bukan keputusan tim). Opsional: `steps_to_reproduce`, `attachments` (screenshot, maksimum 5).
4. Submit → status awal `new`.
5. Member bisa memantau status laporannya sendiri (hanya miliknya) mengikuti lifecycle: `new` → `triaged` (sudah diklasifikasi & diberi `priority` oleh tim) → `in_progress` (`assigned_to` terisi) → `resolved` → `closed`; atau bisa juga bercabang dari `triaged` langsung ke `wont_fix` → `closed`.
6. Jika bug sudah terhubung ke issue tracker eksternal (`external_issue_id`/`external_issue_url` — bagian dari Fase 2 yang disiapkan strukturnya tapi **belum diimplementasikan**, saat ini masih Fase 1 standalone), member berpotensi melihat referensi link issue tersebut — ditandai sebagai kemampuan masa depan, bukan yang sudah aktif.

**Selesai:** member submit bug report dan bisa memantau progresnya sendiri sampai `resolved`/`closed`/`wont_fix` — beda dengan content report, di sini member punya visibilitas penuh atas laporannya.

## Keterlibatan Admin — 💻 Web/Backoffice

1. **Content report** — moderator/leader komunitas dengan permission `manage_reports` menangani report berscope komunitasnya sendiri; Superadmin menangani semua (komunitas + global). Alur: lihat antrian (filter scope/status) → ambil (`reviewing`) → putuskan `action_taken` (memanggil endpoint moderasi yang sudah ada, mis. hapus post/ban user di Community backoffice — sistem report tidak menduplikasi logika moderasi) **atau** `dismissed` → isi `resolution_note` → status `resolved`. Opsional: notifikasi ke pelapor saat selesai.
2. **Bug report** — role khusus support/dev (permission `manage_bug_reports`) atau Superadmin bisa lihat semua laporan, melakukan triage (menentukan `priority` & `assigned_to`), memproses `in_progress`, lalu menutup sebagai `resolved`/`wont_fix`/`closed`.
3. **Integrasi issue tracker** (Fase 2, belum berjalan): saat bug ditriage, sistem berpotensi membuat issue di Linear/Jira via API dan menyimpan `external_issue_id`/`external_issue_url`, lalu sinkronisasi status balik — struktur sudah disiapkan tanpa perlu migrasi besar nantinya.
4. **Integrasi dengan modul lain**: aksi `action_taken` memanggil endpoint moderasi modul terkait (Community, dll); notifikasi (report resolved, auto-flag) dipancarkan lewat modul Notification; permission `manage_reports`/`manage_bug_reports` dikelola lewat modul Role-Permission; `report_count` dikonsumsi modul konten lain untuk sort `most_reported`.

## Di Luar Cakupan Standard & Pro

- **Melihat status content report miliknya sendiri** — dokumen sumber eksplisit tidak memberi member akses ini; hanya moderator/Superadmin yang bisa melihat antrian report.
- **Memproses/resolve report apa pun** — baik content report maupun bug report, keputusan tindak lanjut selalu di tangan moderator komunitas/Superadmin (content) atau support-dev/Superadmin (bug), bukan hak member.
- **Mengakses issue tracker eksternal (Linear/Jira)** — itu tools internal tim dev/support, tidak diekspos ke member bahkan setelah Fase 2 (integrasi tracker) berjalan.
- **Melihat identitas pelapor lain atau detail report_count granular** — member paling banyak melihat efek tidak langsung seperti sort "most_reported" di listing komunitas, itu pun angka agregat, bukan rincian siapa yang melapor.
- **Melihat/mengubah keputusan `resolution` (action_taken vs dismissed) secara langsung** — keputusan ini murni domain moderator/Superadmin, member paling banter menerima notifikasi opsional saat laporannya selesai diproses.

## Edge Case & Catatan Tambahan

- **Ambang auto-flag** default 5 report (configurable) — tidak memicu penghapusan otomatis, tetap perlu keputusan manusia dari moderator/Superadmin.
- **Rate-limit** default 20 report/hari per user, plus pelacakan rasio laporan `dismissed` untuk mendeteksi pola pelaporan yang tidak akurat/abusive.
- **Beberapa keputusan masih terbuka** di `REPORTING_RULES.md` sendiri (bukan diasumsikan, memang ditandai open di dokumen sumber): apakah permission moderator komunitas memakai `manage_reports` baru atau reuse `moderate_posts`; definisi role `support`/`dev` khusus bug report di backoffice belum final; auto-create issue tracker (Fase 2) masih ditunda.
- **Notifikasi ke pelapor saat report resolved** bersifat opsional dengan asumsi sementara default aktif — belum menjadi keputusan final di dokumen sumber.
