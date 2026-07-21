# Ads — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Ads/ADS_RULES.md`.

## Ringkasan Domain

Ads adalah sistem iklan berbayar internal platform KAI — iklan tampil ke **semua pengguna** (member maupun guest) tanpa targeting khusus, di home screen (slider banner + native feed) dan halaman khusus "Iklan & Promo". Superadmin (KAI Pusat) bisa membuat ads apa pun tanpa batas jumlah/durasi dan **tanpa approval** — representasi promosi resmi KAI Pusat. Member baru bisa membuat ads kalau punya benefit `post_ads` (default hanya aktif di plan Pro, tapi Superadmin bisa toggle kapan saja tanpa deploy ulang). Bahkan dengan benefit aktif, publish bisa **langsung aktif** atau **masuk antrian review** tergantung `approval_mode` yang dikonfigurasi Superadmin di Ad Setting backoffice — bukan sesuatu yang otomatis pasti satu arah untuk semua member. Member tanpa benefit hanya melihat ajakan upgrade ke Pro di tempat tombol "Pasang Iklan" seharusnya muncul.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Melihat ads (home slider, feed, halaman Iklan & Promo) | ✅ | ✅ | Berlaku juga untuk guest — tidak ada gating tampilan sama sekali |
| Tap ads → buka `click_url` | ✅ | ✅ | Di in-app browser atau deep link |
| Membuat ads | ❌ | ✅* | *Butuh benefit `post_ads` aktif di plan — default hanya Pro, tapi bisa dikonfigurasi ulang Superadmin kapan saja |
| Publish langsung tanpa review | ❌ | ✅ / ❌ | Tergantung `approval_mode` di Ad Setting (`auto_publish` vs `require_review`) — bukan hak pasti Pro, murni konfigurasi global Superadmin |
| Edit ads sendiri | ❌ | ✅* | *Hanya saat status `draft`/`pending`; kalau sudah `active` harus di-pause dulu sebelum bisa edit |
| Pause/resume ads sendiri | ❌ | ✅ | |
| Hapus ads sendiri | ❌ | ✅ | Kapan saja, untuk semua status |
| Lihat analytics ads sendiri (impressions/clicks/CTR) | ❌ | ✅ | Hanya data ads miliknya sendiri |
| Moderasi ads member lain | ❌ | ❌ | Eksklusif Superadmin |
| Konfigurasi Ad Setting (approval mode, max durasi, dll) | ❌ | ❌ | Eksklusif Superadmin |

## Journey 1: Member Melihat Ads — 🅢🅟 — 📱 Mobile

1. **Entry point 1 — Home screen**: member (atau guest) membuka app → slider banner muncul di bagian atas home, tepat di bawah top navigation bar (hanya jika ada ads aktif tipe Image Banner/Video Banner).
2. Slider menampilkan 1–5 ads aktif secara bergantian, auto-scroll setiap 5 detik (berhenti otomatis saat user sedang swipe manual atau video sedang play). Ads Superadmin/KAI Pusat diprioritaskan tampil lebih dulu, baru ads member diurutkan berdasarkan `start_date` terbaru.
3. Saat scroll feed (news/event/komunitas), setiap N item konten muncul 1 native/text ad yang menyisip (N dikonfigurasi Superadmin lewat `feed_ads_interval`, default setiap 5 item) — diberi highlight border kiri tipis. Jika ads aktif tidak cukup, slot tetap kosong (tidak dipaksa isi).
4. Semua ads (di slider maupun feed) memiliki label "Iklan" yang jelas di sudut, agar transparan ke user.
5. Member/guest tap ads → `click_url` terbuka di in-app browser, atau deep link in-app kalau `click_url` diset Superadmin sebagai deep link (member creator hanya bisa set URL eksternal).
6. **Entry point 2 — Halaman "Iklan & Promo"**: dari tab khusus di bottom navigation. Struktur: slider banner (sama seperti home) di atas, lalu list semua ads aktif (semua tipe: image, video, text, native) dalam format card — thumbnail/icon, badge tipe, judul, nama pengiklan, tanggal aktif, tombol CTA. Diurutkan: ads KAI Pusat paling atas, sisanya berdasarkan `start_date` terbaru.
7. Jika tidak ada ads aktif sama sekali: section slider di home disembunyikan, slot native feed tidak diisi (feed tetap normal tanpa gap), dan halaman "Iklan & Promo" tetap bisa dibuka tapi menampilkan state kosong "Belum ada iklan aktif saat ini".

**Selesai:** member/guest melihat dan (opsional) berinteraksi dengan ads yang sedang tayang; tidak ada "tahap akhir" formal karena ini alur konsumsi pasif berulang setiap kali membuka app.

## Journey 2: Pro Member Memasang Ads — 🅟 — 📱 Mobile

**Prasyarat**: benefit `post_ads` harus aktif pada plan member (default Pro, tapi selalu dikonfigurasi terpisah oleh Superadmin di backoffice subscription — bukan otomatis menyala hanya karena upgrade ke Pro). Tanpa benefit ini, tombol "Pasang Iklanmu" di halaman "Iklan & Promo" menampilkan pesan ajakan upgrade ke Pro, bukan form pembuatan ads.

1. Member membuka halaman "Iklan & Promo" → tap tombol "Pasang Iklanmu" (hanya muncul jika benefit `post_ads` aktif).
2. Pilih salah satu dari 4 tipe ads — **Image Banner**, **Video Banner**, **Text Ad**, atau **Native Ad** — masing-masing punya field konten & ketentuan media berbeda (lihat `ADS_RULES.md` §3 untuk detail tiap tipe).
3. Isi form konten sesuai tipe, lalu set jadwal tayang (`start_date`, `end_date` — durasi maksimum dibatasi `max_duration_days`, default 30 hari) → submit.
4. **Jika `approval_mode = require_review`** (default): status jadi `pending`, muncul info "Iklanmu sedang direview oleh admin". Selama `pending`, member masih bisa edit konten dan menghapusnya kapan saja.
5. **Jika `approval_mode = auto_publish`**: status langsung `active`, muncul konfirmasi "Iklanmu sudah aktif!" — tayang langsung di app tanpa menunggu review (Superadmin tetap bisa pause/reject/hapus kapan saja setelahnya sebagai moderasi).
6. **Kalau ditolak (`rejected`)**: member menerima notifikasi berisi alasan penolakan (`reject_reason`, wajib diisi Superadmin) — member bisa edit lalu submit ulang, kembali masuk status `pending`.
7. Member memantau ads-nya dari halaman "Iklan Saya": edit (hanya saat `draft`/`pending` — kalau sudah `active` harus di-pause dulu), pause/resume (`active ⇄ paused`), atau hapus (kapan saja, semua status).
8. Member melihat analytics ads miliknya sendiri: `impressions`, `clicks`, `ctr` — hanya data ads sendiri, bukan platform-wide.
9. **Mencapai batas `max_active_ads_per_member`** (default 3): tombol "Buat Ads Baru" di-disable dengan pesan "Kamu sudah mencapai batas maksimum X iklan aktif. Selesaikan atau hapus iklan lama untuk membuat yang baru."
10. **Ads expired otomatis**: sistem cron harian mengubah status `active`/`paused` dengan `end_date` terlewati menjadi `expired` — tidak perlu aksi manual member.
11. **Downgrade dari Pro ke Standard** (kehilangan benefit `post_ads`): ads yang sedang `active` tetap tayang sampai `end_date` (tidak langsung dihapus), tapi member tidak bisa membuat ads baru, tidak bisa edit, dan tidak bisa resume ads yang sedang `paused`.

**Selesai (jalur sukses):** ads member tayang `active` di app dan bisa diklik user lain, dengan analytics terpantau dari halaman "Iklan Saya".

**Selesai (jalur ditolak):** member melihat status `rejected` beserta alasan, bisa edit & submit ulang kapan saja.

## Keterlibatan Admin — 💻 Web/Backoffice

1. **Buat ads langsung**: Superadmin bisa membuat ads tipe apa pun tanpa batas jumlah dan tanpa approval — langsung `active`. Batasan `max_active_ads_per_member` dan `max_duration_days` hanya berlaku untuk ads dari member, tidak untuk Superadmin.
2. **Moderasi ads member**: approve/reject (wajib isi `reject_reason` saat reject, dikirim ke member via notifikasi in-app), pause, atau hapus ads member kapan saja — juga bisa edit ads member sebagai moderator.
3. **Konfigurasi Ad Setting**: `approval_mode` (`auto_publish`/`require_review`), `max_active_ads_per_member`, `max_duration_days`, `feed_ads_interval`, `slider_max_items`. Perubahan berlaku untuk ads **baru** setelah perubahan — ads yang sudah aktif tidak terpengaruh kecuali di-pause/hapus manual.
4. **Notifikasi antrian**: saat `approval_mode = require_review`, Superadmin menerima notifikasi "Ada ads baru menunggu review" setiap ada submission member.
5. **Analytics platform-wide**: Superadmin melihat impressi, klik, CTR seluruh ads di platform (aggregate maupun per ads), sedangkan member hanya melihat data ads miliknya sendiri.

## Di Luar Cakupan Standard & Pro

- **Standard tidak bisa membuat ads sama sekali** — benefit `post_ads` secara default hanya diberikan ke plan Pro; tanpa benefit ini di plan-nya, tombol pembuatan ads tidak muncul di UI.
- **Moderasi ads member lain** — eksklusif Superadmin; member Pro hanya bisa kelola ads miliknya sendiri.
- **Konfigurasi Ad Setting** (`approval_mode`, `max_active_ads_per_member`, `max_duration_days`, `feed_ads_interval`, `slider_max_items`) — eksklusif Superadmin lewat backoffice.
- **Targeting user segment tertentu untuk ads** — sistem Ads memang tidak mendukung targeting sama sekali (semua ads tampil ke semua user tanpa kecuali); ini bukan pembatasan tier, tapi karakter desain sistem.
- **Set `click_url` sebagai deep link in-app (`kai://...`)** — hanya Superadmin yang bisa akses deep link; member hanya bisa input URL eksternal (http/https).
- **Bebas dari batas jumlah/durasi ads** — `max_active_ads_per_member` dan `max_duration_days` hanya berlaku untuk ads member; Superadmin dikecualikan.

## Edge Case & Catatan Tambahan

- **Analytics reset saat pause → resume**: counter impressions/clicks di-reset ke 0 (fresh counting), bukan melanjutkan angka sebelumnya.
- **Satu impression per session**: user yang melihat ads sama berkali-kali dalam satu session tetap dihitung 1 impression saja (tidak double count).
- **`ADS_RULES.md` tidak mendetailkan** apakah member menerima notifikasi khusus saat ads-nya mendekati `end_date` (mis. reminder "iklanmu akan expired 3 hari lagi") — tidak disebutkan di dokumen sumber, jadi tidak diasumsikan ada.
