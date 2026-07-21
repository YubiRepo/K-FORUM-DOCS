# Notifications (Inbox & Preferences) — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Notifications/NOTIFICATION_RULES_ENGINE.md`, `Modules/Notifications/notification-preferences-technical.md`, `API SPEC/Mobile/API_SPEC_NOTIFICATION_INBOX_MOBILE.md`, `API SPEC/Mobile/API_SPEC_NOTIFICATION_PREFERENCES.md`.

## Ringkasan Domain

Notifications adalah domain **cross-cutting** — ia tidak menghasilkan konten sendiri, tapi menerima event dari semua domain lain (Announcement, News, Community, Event, Q&A, Subscription, Reporting, Region) lalu menyalurkannya ke member lewat push (FCM) dan/atau inbox in-app. Semua logic penentuan "kirim ke siapa, lewat channel apa, bisa dimatikan atau tidak" berjalan di backend (Rules Engine + async queue asynq); Flutter murni **konsumen**: register FCM token, kirim/baca preferences, terima payload, render notifikasi, dan handle deep link dari `click_action`.

Bagi member, domain ini punya dua permukaan: **Inbox** (lihat & kelola notifikasi yang sudah diterima) dan **Preferences** (atur mana yang boleh masuk, lewat channel apa, kapan mode senyap aktif). Konten notifikasi yang benar-benar diterima seorang member sifatnya **turunan dari domain lain** — mis. member yang aktif post di Community akan menerima notifikasi post baru, sedangkan Pro Member yang submit artikel News akan menerima notifikasi approve/reject artikelnya (Standard tidak akan pernah dapat notifikasi ini karena tidak bisa submit artikel sama sekali).

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Terima & baca notifikasi inbox (in-app) | ✅ | ✅ | Sama persis — bukan fitur berbayar |
| Terima push notification (FCM) | ✅ | ✅ | Sama persis |
| Atur preferences global (`all_notifications_enabled`, Do Not Disturb) | ✅ | ✅ | Sama persis |
| Atur preferences per-modul (news, community, event, qna, announcement, subscription) | ✅ | ✅ | Sama persis |
| Atur preferences per-komunitas | ✅ | ✅ | Auto-created saat join komunitas, berlaku siapa saja |
| Reset preferences ke default | ✅ | ✅ | Sama persis |
| Menerima notifikasi transaksional milik domain yang hanya bisa diakses Pro (mis. approve/reject artikel News, karena hanya Pro yang bisa submit artikel) | ❌ (tidak relevan) | ✅ | Bukan pembatasan tier di domain Notifications sendiri — murni konsekuensi Pro punya lebih banyak aksi yang memicu notifikasi transaksional di domain lain |
| Bypass rules (notifikasi yang tidak bisa dimatikan) berlaku sama | ✅ | ✅ | Lihat [Journey 2](#journey-2-member-mengatur-preferensi-notifikasi--) |

Kesimpulan: domain ini **tidak plan-gated sama sekali**. Baris terakhir sengaja ditulis untuk menegaskan bahwa perbedaan "isi" notifikasi yang diterima Standard vs Pro murni cerminan dari fitur apa yang mereka pakai di domain lain — bukan karena akses ke domain Notifications itu sendiri dibedakan.

## Journey 1: Member Membuka Inbox Notifikasi — 🅢🅟 — 📱 Mobile

1. **Entry point**: member membuka app / masuk halaman inbox notifikasi. App menjalankan pola standar saat launch/login:
   - `GET /notifications` — memuat 20 notifikasi terbaru milik user (campuran sudah & belum dibaca, urut terbaru).
   - `GET /notifications/unread-count` — mengisi badge counter.
   - Connect ke `WS /ws/notifications` — mendengarkan notifikasi baru secara real-time tanpa polling (keepalive ping/pong tiap 50 detik).
2. **List notifikasi**: setiap item punya `type` (`security`, `social`, `content`, `system`, `reminder`, `marketing`), `title`, `body`, opsional `image_url`, serta metadata sumber (`module`, `event_type`, `entity_id`) dan status `is_read`/`read_at`. Tidak ada dokumentasi soal pengelompokan visual per kategori di UI — pengelompokan/filter tampilan (kalau ada) tidak dijelaskan di source doc.
3. **Real-time update**: saat WS mengirim pesan baru, item di-append ke list lokal dan badge count bertambah tanpa perlu re-fetch. Jika koneksi WS terputus, notifikasi tetap tersimpan di DB — client cukup `GET /notifications` lagi saat reconnect untuk catch-up (tidak hilang).
4. **Tap notifikasi → deep link**: member tap satu item, app membaca `click_action` (format `kai://.../...`, konsisten antara FCM push, WS, dan REST inbox) untuk navigasi langsung ke entity terkait (artikel, post komunitas, event, dll), lalu memanggil `PATCH /notifications/:id/read` untuk menandai dibaca.
5. **Mark all read**: member bisa tap "tandai semua dibaca" → `PATCH /notifications/read-all`, best-effort (kalau satu gagal ditandai, proses tetap lanjut ke sisanya).
6. **Notifikasi bypass**: item dengan `bypass: true` (mis. announcement darurat) tetap muncul di inbox seperti biasa — flag ini hanya relevan untuk logic pengiriman/preferences di backend, bukan tampilan berbeda di UI (tidak ada styling khusus yang didokumentasikan untuk item bypass).
7. **Empty state**: source doc tidak mendefinisikan copy/desain untuk kondisi inbox kosong (belum ada notifikasi sama sekali) — dicatat sebagai gap dokumentasi, bukan diasumsikan.

**Selesai:** member sudah membaca notifikasi terbaru dan/atau berpindah ke layar terkait via deep link; ini alur berulang tanpa "tahap akhir" formal, sama seperti konsumsi konten pada umumnya.

## Journey 2: Member Mengatur Preferensi Notifikasi — 🅢🅟 — 📱 Mobile

1. **Entry point**: member membuka menu "Pengaturan Notifikasi" → `GET /preferences` mengambil seluruh preferences miliknya. Jika user belum pernah punya baris preferences (user baru), server **auto-create dengan default** dan langsung mengembalikannya (semua modul `enabled: true` secara default, kecuali per-komunitas `member_joined`/`member_left` yang default `false`).
2. **Toggle global**:
   - `all_notifications_enabled` — master switch mematikan semua notifikasi (kecuali yang bypass, lihat poin 5).
   - `do_not_disturb_enabled` + jam mulai/selesai (`do_not_disturb_start_time`/`end_time`, format `HH:mm`) — mode senyap terjadwal. Jam ini adalah **jam lokal wall-clock**, bukan UTC — backend menentukan "sekarang jam berapa untuk user ini" dari timezone device user yang paling terakhir aktif (dikirim otomatis saat login/refresh/register FCM token), fallback ke `default_timezone` System Settings lalu `Asia/Jakarta` kalau itu pun tidak ada. Member tidak perlu (dan tidak bisa) memilih timezone eksplisit untuk field ini.
   - Update lewat `PUT /preferences/global`.
3. **Toggle per-modul** (`PUT /preferences/modules/{module}`), modul yang tersedia:
   - `announcement` → hanya sub-field `info_enabled` yang bisa diatur (untuk tipe `info` priority MEDIUM/LOW). Tipe darurat tidak muncul di sini karena memang tidak bisa dimatikan (lihat poin 5).
   - `news` → satu toggle `enabled` untuk semua notifikasi artikel baru (dari KAI Pusat, KAI Regional, maupun Pro Members).
   - `community` → toggle `enabled` untuk keseluruhan; pengaturan lebih granular per-komunitas ada di endpoint terpisah (poin 4).
   - `event` → `enabled`, `reminders_enabled`, `reminder_hours_before` (default 24 jam sebelum event), dan `interested_categories` (kosong = terima semua kategori event, terisi = hanya kategori yang match).
   - `qna` → `enabled` + `question_answered` (satu toggle untuk notifikasi "pertanyaan dijawab" maupun "pertanyaan ditolak", karena keduanya sama-sama update status pertanyaan dari sudut pandang member).
   - `subscription` → hanya `expiry_reminder_enabled` yang bisa diatur (reminder 7 & 3 hari sebelum plan Pro expired). Notifikasi status change (approved/rejected/expired/downgraded) tidak muncul di sini karena tidak bisa dimatikan (lihat poin 5).
4. **Toggle per-komunitas** (`PUT /preferences/communities/{community_id}`): member atur per komunitas yang diikutinya — `enabled`, `new_posts`, `member_joined`, `member_left`, `join_request_approved` (relevan khusus komunitas private; untuk komunitas public field ini bisa diisi tapi tidak akan pernah terpicu karena public pakai auto-join tanpa approval). Baris preference komunitas ini **auto-dibuat dengan default** saat member join komunitas — tidak perlu setup manual.
5. **Batas opt-out (tidak bisa dimatikan)** — beberapa notifikasi selalu terkirim meski `all_notifications_enabled = false` dan DND aktif:
   - Announcement tipe `disaster`/`system`/`urgent` dengan priority `CRITICAL`/`HIGH`.
   - Semua perubahan status subscription (upgrade approved/rejected, plan expired/downgraded).
   - Beberapa notifikasi transaksional lain di domain masing-masing yang ditandai bypass (mis. approve/reject artikel News Pro Member, kick/ban dari komunitas, member disetujui masuk region) — ini bukan sesuatu yang diatur lewat toggle preferences sama sekali karena sifatnya transaksional terhadap status akun/konten member sendiri.
6. **Reset ke default**: `POST /preferences/reset` mengembalikan seluruh preferences (global, per-modul, per-komunitas) ke nilai default pabrik.
7. **Fail-safe**: jika preferences tidak bisa diakses backend (mis. DB down) saat proses pengiriman, sistem **default ke tidak mengirim** (fail-closed) kecuali notifikasi itu bypass — jadi kegagalan infrastruktur tidak pernah membuat member kebanjiran notifikasi yang seharusnya sudah dia matikan, tapi juga tidak menjamin notifikasi non-bypass terkirim saat preferences tidak terbaca.

**Selesai:** preferences tersimpan dan langsung berlaku untuk pengiriman notifikasi berikutnya (dicek ulang setiap kali worker akan mengirim, bukan snapshot).

## Keterlibatan Admin — 💻 Web/Backoffice

Tidak ada admin tool yang mengelola preferensi atau inbox notifikasi member — domain ini murni self-scope, sama seperti prinsip di [14_USER_SETTINGS_JOURNEY.md](14_USER_SETTINGS_JOURNEY.md). Source doc tidak menyebutkan adanya notification tester/broadcast tool khusus di backoffice untuk mengirim notifikasi uji coba ke member.

Keterlibatan admin di domain ini murni **sebagai efek samping** dari aksi mereka di domain lain — mis. Superadmin approve/reject artikel Pro Member (News), publish announcement, atau approve join request (Community) — masing-masing memicu job notifikasi lewat Rules Engine yang sama, tapi aksi & UI-nya ada di journey domain masing-masing, bukan di domain Notifications ini.

Perlu dicatat juga: Superadmin, Admin Regional, dan Usergod sendiri **menerima** notifikasi operasional (mis. antrian approval baru, report masuk) — tapi selalu lewat **in-app backoffice queue**, tidak pernah lewat FCM push, dan ini adalah kotak masuk terpisah untuk persona admin, bukan sesuatu yang mereka kelola untuk member.

## Di Luar Cakupan Standard & Pro

- **Mengirim notifikasi ke user lain** — member tidak punya mekanisme broadcast/kirim notifikasi manual; semua notifikasi dipicu otomatis oleh event domain (publish, approve, join, dll), bukan aksi langsung member ke member lain.
- **Melihat delivery/engagement analytics** — data seperti `notification_delivery_logs`, status "partial delivery", atau statistik open-rate adalah tooling internal backend (asynqmon, DLQ inspection), tidak pernah diekspos ke member.
- **Mengubah/mematikan bypass rules** — notifikasi yang ditandai bypass (announcement darurat, perubahan status subscription, dll) tidak bisa dimatikan lewat preferences apa pun; ini keputusan desain keamanan/kepatuhan, bukan pembatasan tier.
- **Melihat atau mengubah preferences member lain** — self-scope sepenuhnya; tidak ada peran admin yang mengelola preferences notifikasi milik user lain lewat domain ini.
- **Mengatur channel Email** — dari Channel Matrix, Email dipakai untuk sebagian notifikasi transaksional penting (subscription, announcement CRITICAL), tapi tidak ada toggle Email terpisah untuk member di API preferences yang didokumentasikan — semua kontrol yang ada bersifat per-modul, bukan per-channel.

## Edge Case & Catatan Tambahan

- **Idempotency**: setiap notifikasi job punya `event_id` unik; worker cek apakah event sudah pernah diproses sebelum kirim ulang, jadi member tidak akan menerima notifikasi duplikat meski job di-retry akibat error sementara.
- **Retry & kegagalan pengiriman**: kegagalan FCM sementara (5xx/timeout) di-retry otomatis (maksimal 3x, backoff 1s/4s/16s) sebelum masuk Dead Letter Queue; token device yang sudah invalid langsung dinonaktifkan tanpa retry. Ini murni operasional backend — tidak ada yang perlu dilakukan member.
- **Konsistensi payload**: field `module`, `event_type`, `entity_id`, `click_action`, `bypass` identik di tiga channel (FCM push, WebSocket, REST inbox) — Flutter bisa pakai satu handler navigasi untuk ketiganya. Bedanya hanya bentuk field `extra`/`payload` (JSON object di WS/REST, JSON-encoded string di FCM data payload — perlu `jsonDecode` khusus di FCM).
- **Preferensi komunitas untuk komunitas baru**: begitu member join komunitas baru, baris preference untuk komunitas itu otomatis dibuat dengan default (`new_posts: true`, `member_joined`/`member_left: false`) — member tidak perlu setup manual sebelum bisa menerima notifikasi dari komunitas tersebut.
