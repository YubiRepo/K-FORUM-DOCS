# News — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/News/NEWS_RULES.md`, `Modules/News/NEWS_MEDIA_FLOWS.md`.

## Ringkasan Domain

News adalah modul konten editorial — bukan konten user-generated seperti Community. Secara default, **semua artikel dibuat oleh Editor/admin** (manual) atau **hasil scraping otomatis** (eksklusif KAI Pusat). Member (Standard & Pro) pada dasarnya adalah **konsumen**: baca, like, comment, bookmark, share. Satu-satunya hal yang membedakan Pro di domain ini adalah kemampuan opsional **submit artikel sendiri** (jika benefit `post_news` aktif di plan-nya) — tapi tetap tidak bisa publish langsung, selalu lewat antrian approval admin.

News **tidak scoped secara visibility** — semua artikel published bisa dilihat semua orang (member maupun guest), apa pun label asalnya (KAI Pusat atau region tertentu). Yang membedakan artikel hanyalah **label asal** (informatif) dan **News Scope** (`indonesia` / `korea` / `korea_indonesia`, dimensi filter geografis yang orthogonal terhadap kategori) — keduanya bukan pembatas akses.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Baca artikel published (list & detail) | ✅ | ✅ | Global, tidak scoped — termasuk artikel label region lain |
| Pilih/ganti bahasa baca | ✅ | ✅ | Mengikuti `Accept-Language`, fallback ke original |
| Like artikel | ✅ | ✅ | Reaksi hanya "like" (MVP) |
| Comment & reply artikel | ✅ | ✅ | Threaded 2 level, harus login |
| Bookmark/save artikel | ✅ | ✅ | |
| Share artikel (WA/sosmed/copy link) | ✅ | ✅ | |
| Filter berdasarkan kategori / News Scope / label asal | ✅ | ✅ | Opsional di UI |
| Terima notifikasi artikel baru | ✅ | ✅ | Sesuai preferensi notifikasi |
| **Submit artikel sendiri** | ❌ | ✅* | *Butuh benefit `post_news` aktif di plan Pro (dikonfigurasi Superadmin, bukan otomatis by-design) |
| Edit artikel sendiri (status `draft`/`rejected`) | ❌ | ✅ | Hanya sebelum/setelah approval, bukan saat `pending_approval` |
| Withdraw submission sebelum di-approve | ❌ | ✅ | |
| Publish langsung tanpa approval | ❌ | ❌ | Pro tetap selalu lewat `pending_approval` — tidak ada auto-publish untuk member |
| Edit artikel sendiri yang sudah `published` | ❌ | ❌ | Tidak tercantum sebagai hak Member Pro di `NEWS_RULES.md` — edit published hanya hak Editor/admin |

## Journey 1: Member Membaca & Berinteraksi dengan News — 🅢🅟 — 📱 Mobile

1. **Entry point**: member membuka tab News di aplikasi → melihat listing artikel `published`, urut terbaru, dengan artikel `is_featured` di-pin/disorot di bagian atas.
2. **Filter/browse** (opsional): filter berdasarkan kategori (Politik, Ekonomi, Olahraga, dll — dikelola Superadmin), News Scope (Berita Indonesia / Berita Korea / Berita Korea di Indonesia), atau label asal (KAI Pusat / KAI [Region]). Filter ini murni UI, tidak membatasi artikel mana yang boleh dilihat.
3. Member tap satu artikel → masuk halaman detail: judul, konten, thumbnail, author, label asal (`author_label`), kategori, dan counter like/comment.
4. **Bahasa tampil**: artikel dirender sesuai `Accept-Language` device member.
   - Jika translation untuk bahasa itu sudah `done` → langsung tampil versi tersebut.
   - Jika belum ada, dan `on_demand_enabled=true` → sistem trigger translate di background, member tetap dapat versi original dulu (dengan flag `is_translated: false`), request berikutnya baru dapat versi terjemahan.
   - Jika `on_demand_enabled=false` → member dapat versi original, dengan opsi dropdown pilih bahasa lain dari daftar `available_languages` — begitu dipilih, translate di-trigger dan di-cache untuk pembaca berikutnya.
   - Jika `translation_enabled=false` secara global → semua member hanya dapat bahasa original, tanpa opsi ganti bahasa.
   - Catatan implementasi saat ini: provider translate masih `NoopTranslationProvider` — job translate tetap berjalan (`pending → processing → done`) tapi isi hasil terjemahan **sama persis dengan original** (`translated_by=noop`), belum benar-benar diterjemahkan.
5. **View count** tercatat otomatis tiap artikel dibuka: `view_count` (semua hit, termasuk guest & buka berulang) dan `unique_view_count` (khusus user login, satu kali per user).
6. **Like** — member tap ikon like, `like_count` bertambah. Reaksi hanya satu jenis (like) di MVP.
7. **Comment**:
   - Member tulis comment baru → tersimpan sebagai level 1, `comment_count` artikel bertambah.
   - Member tap "Reply" pada comment atau reply mana pun → tersimpan sebagai reply ke comment level 1 (tap reply pada reply tidak membuat nesting lebih dalam — tetap masuk sebagai reply ke parent level 1, seperti Instagram/Twitter).
   - Comment **tidak bisa diedit**, hanya bisa di-soft-delete oleh pemilik comment sendiri (atau Editor/Superadmin). Setelah dihapus, tampil "Komentar ini telah dihapus"; reply di bawahnya tetap tampil.
   - Member bisa report comment yang tidak pantas → masuk reporting system (`reportable_type = news_comment`).
8. **Bookmark/save** — member tap simpan untuk dibaca lagi nanti dari daftar bookmark pribadi.
9. **Share** — member share artikel ke WhatsApp, media sosial, atau copy link.
10. **Notifikasi artikel baru** — member menerima push notification sesuai preferensi notifikasi masing-masing.
11. **Guest (belum login)**: bisa membuka & membaca artikel serta comment tanpa login, tapi tombol like/comment/bookmark/share tidak aktif — diarahkan ke login jika mencoba.

**Selesai:** member selesai membaca/berinteraksi dengan satu atau lebih artikel sesuai preferensinya; tidak ada "tahap akhir" formal karena ini alur konsumsi konten yang berulang.

> Catatan: `NEWS_RULES.md` & `NEWS_MEDIA_FLOWS.md` tidak mendetailkan copy/desain untuk empty state (mis. "belum ada artikel di kategori ini") atau error state jaringan spesifik pada listing/detail — bagian ini tidak ditulis lebih detail karena tidak ada dasar dari sumber dokumen.

## Journey 2: Pro Member Posting News — 🅟 — 📱 Mobile

**Prasyarat**: benefit `post_news` harus aktif pada plan Pro member (dikonfigurasi Superadmin lewat Plan Benefits — bukan otomatis menyala hanya karena upgrade ke Pro). Jika benefit tidak aktif, entry point "Tulis Artikel" tidak muncul di aplikasi meski member sudah Pro.

1. Member Pro membuka News → tap "Tulis Artikel".
2. **Compose draft**: pilih satu bahasa utama, isi judul/konten/summary/tag, pilih kategori (wajib — setiap artikel harus punya kategori), News Scope (opsional).
3. **Upload thumbnail** (opsional): app request presign URL → upload file langsung ke MinIO/S3 (tidak lewat API backend) → key tersimpan dengan prefix `s3:` (mis. `s3:/news/thumbnails/uuid.jpg`).
4. Member menyimpan sebagai **draft** dulu, atau langsung **submit**.
   - Saat submit (`POST /mobile/news/articles`), sistem confirm thumbnail (`NormalizeValue` + `ConfirmUpload`, status `PENDING → CONFIRMED`) dan menyimpan article + translation asli sekaligus.
5. Status berubah dari `draft` → `pending_approval`. Member melihat status "⏳ Menunggu review admin" di daftar artikel miliknya.
6. **Selama `pending_approval`**:
   - Member **tidak bisa mengedit** konten langsung (edit hanya diizinkan saat status `draft` atau `rejected`).
   - Member **bisa withdraw** (batalkan submission) sebelum di-approve — mengembalikan artikel ke status yang bisa diedit lagi.
7. **Jika disetujui (approve)** → status `published`, artikel tayang global ke semua user seperti artikel lainnya, dengan label asal (`author_label`) tampil ke pembaca.
8. **Jika ditolak (reject)** → status `rejected`. Member bisa membuka kembali artikelnya, mengedit, lalu resubmit — kembali masuk antrian `pending_approval`.
9. Member Pro **tidak pernah** bisa publish langsung — semua submission, disetujui atau tidak, selalu melalui approval Editor/Superadmin yang punya permission `approve_news`.
10. Setelah artikel `published`, Member Pro **tidak bisa mengedit lagi** — perubahan pada artikel published (mis. perbaiki typo) hanya kewenangan Editor/admin, bukan hak Member Pro.

**Selesai (jalur sukses):** artikel Member Pro tayang published dan bisa dibaca/like/comment seperti artikel editorial lainnya.

**Selesai (jalur ditolak):** member melihat status `rejected` di daftar artikelnya, bisa edit & resubmit kapan saja.

## Keterlibatan Admin — 💻 Web/Backoffice

> Catatan peran: `NEWS_RULES.md` mendefinisikan role operasional **Editor** (di-assign oleh Usergod/Superadmin, "label asal sesuai region Editor") sebagai analog terdekat dari **Admin Regional** di legend `00_OVERVIEW.md`. Dokumen sumber tidak memakai istilah "Admin Regional" secara eksplisit untuk modul News — istilah yang dipakai konsisten adalah "Editor".

1. **Antrian approval**: Editor atau Superadmin yang punya permission `approve_news` melihat daftar artikel Member Pro berstatus `pending_approval`.
2. **Approve** → status berubah jadi `published`, artikel tayang ke semua user secara global.
3. **Reject** → status berubah jadi `rejected`; member pemilik bisa edit & resubmit. `NEWS_RULES.md` tidak mendetailkan apakah reject wajib disertai kolom alasan terstruktur (berbeda dari flow reject upgrade subscription di [02_SUBSCRIPTION_UPGRADE_JOURNEY.md](02_SUBSCRIPTION_UPGRADE_JOURNEY.md) yang eksplisit punya daftar alasan) — ini ditandai sebagai gap dokumentasi, bukan diasumsikan.
4. **Moderasi konten published**: Editor/Superadmin bisa edit artikel apa pun (termasuk yang sudah `published`) tanpa perlu re-approval, dan bisa archive artikel kapan saja.
5. **Moderasi comment**: Editor/Superadmin bisa soft-delete comment siapa pun (bukan hanya milik sendiri), dan menangani laporan comment yang masuk lewat reporting system.
6. **Kelola kategori & News Scope**: eksklusif Superadmin (CRUD, urutan tampil, aktif/nonaktif) — di luar keterlibatan Editor.
7. **Operasional scraping** (jadwal, `auto_publish`, `ai_cleanup`, `auto_translate` per source, mapping kategori): eksklusif Usergod (setup sumber) & Superadmin (operasional source/category) — sama sekali tidak bersinggungan dengan journey member, disebut di sini hanya sebagai konteks kenapa sebagian artikel non-member-Pro juga melalui status `draft` sebelum tayang.

## Di Luar Cakupan Standard & Pro

- **Auto-publish tanpa approval** — Member Pro selalu lewat `pending_approval`; tidak ada mekanisme publish langsung meski sudah berlangganan Pro.
- **Edit artikel milik member/editor lain** — hanya Editor/Superadmin yang bisa edit artikel siapa pun; Member Pro hanya boleh edit artikelnya sendiri, dan hanya saat status `draft`/`rejected`.
- **Edit artikel sendiri yang sudah published** — begitu tayang, revisi jadi kewenangan Editor/admin.
- **Kelola kategori & News Scope** — CRUD kategori dan scope eksklusif Superadmin.
- **Mendaftarkan/konfigurasi news source untuk scraping** — eksklusif Usergod (pendaftaran) dan Superadmin (operasional jadwal/publish/limit); scraping sendiri eksklusif KAI Pusat, Admin Regional/Editor daerah tidak punya jalur ini.
- **Kelola system languages & News System Settings** (`translation_enabled`, `on_demand_enabled`) — eksklusif Usergod/Superadmin.
- **Approve/reject submission diri sendiri atau member lain** — mutlak kewenangan Editor/Superadmin dengan permission `approve_news`.
- **Reaksi selain like** (mis. emoji reaction) & **like pada comment** — belum ada di MVP untuk siapa pun, bukan pembatasan tier.
- **Guest berinteraksi** (like/comment/bookmark/share) — guest hanya bisa baca, harus login dulu untuk berinteraksi; ini bukan soal tier Standard/Pro.

## Edge Case & Catatan Tambahan

- **Downgrade Pro → Standard saat punya artikel `draft`/`rejected`/`pending_approval`**: `NEWS_RULES.md` tidak membahas dampaknya sama sekali (berbeda dengan domain Subscription yang eksplisit menjelaskan transisi benefit). Tidak jelas apakah artikel yang belum tayang tetap tersimpan, otomatis di-withdraw, atau tetap bisa diproses admin meski member sudah bukan Pro lagi — ini ditandai sebagai celah dokumentasi, bukan diasumsikan.
- **Label asal artikel Member Pro** — termasuk salah satu "Keputusan yang Masih Terbuka" di `NEWS_RULES.md`: apakah `author_label` memakai nama member sendiri atau mengikuti region member. Belum final.
- **Kepemilikan aksi "delete" artikel** — `NEWS_MEDIA_FLOWS.md` menyebut endpoint delete artikel (`EnsureDeletable`) hanya bisa dieksekusi pada status `draft`/`rejected`, tapi tidak menyebutkan secara eksplisit apakah Member Pro punya akses ke aksi delete permanen ini, atau hanya "withdraw" (yang sifatnya membatalkan submission, bukan menghapus). Kedua istilah ini kemungkinan merujuk ke aksi berbeda — dicatat sebagai ambiguitas, bukan disamakan.
- **Thumbnail saat edit draft/rejected**: mengikuti flow update standar — jika member ganti thumbnail, thumbnail lama ditandai `DELETED` dan yang baru `CONFIRMED`; jika thumbnail dihapus (dikosongkan), thumbnail lama langsung ditandai `DELETED`; jika tidak berubah (kirim balik `thumbnail_raw` yang sama), sistem skip tanpa aksi media apa pun.
- **Notifikasi approval Pro**: `NEWS_RULES.md` (Use Case 5) menyebutkan Editor/Superadmin menerima notifikasi saat ada submission baru untuk direview, tapi tidak mendetailkan bentuk notifikasi ke member saat artikelnya di-reject (dibanding saat di-approve) — dicatat sebagai gap, bukan diasumsikan bentuknya.
