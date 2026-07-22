# Event — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Event/EVENT_RULES.md`.

## Ringkasan Domain

Event di KAI App bersifat **global** — semua member bisa melihat dan berinteraksi dengan semua event di platform, tidak dibatasi region. Semua member yang sudah login (Standard maupun Pro) bisa browse, bookmark, menjadwalkan (dengan reminder & export kalender), membagikan, dan memberi feedback setelah event berlangsung. Membuat, mengedit, dan mengelola event adalah eksklusif untuk **Member Pro** — Standard yang mencoba akan diarahkan ke ajakan upgrade.

Event bisa berupa Offline, Online, atau Hybrid, dan melalui alur publikasi yang bisa auto-publish atau butuh persetujuan Superadmin, tergantung konfigurasi platform. Fitur pendaftaran/kehadiran murni via **link eksternal** yang diisi organizer (misal Eventbrite, Zoom) — platform ini sendiri **tidak punya sistem RSVP/attendee tracking internal** (lihat catatan di "Di Luar Cakupan" dan Journey 2).

## Batasan Standard vs Pro di Domain Ini

| Aksi                                                                | Standard | Pro | Catatan                                                                        |
| ------------------------------------------------------------------- | :------: | :-: | ------------------------------------------------------------------------------ |
| Lihat daftar & detail event                                         |    ✅    | ✅  | Semua event published, global                                                  |
| Filter berdasarkan tipe (offline/online/hybrid) & cari lokasi venue |    ✅    | ✅  | Hanya mempersempit pencarian, bukan pembatas akses                             |
| Bookmark event                                                      |    ✅    | ✅  |                                                                                |
| Tambah ke jadwal pribadi (in-app) + reminder                        |    ✅    | ✅  |                                                                                |
| Export ke kalender eksternal (.ics / deep link)                     |    ✅    | ✅  |                                                                                |
| Bagikan event (share)                                               |    ✅    | ✅  |                                                                                |
| Ikuti/daftar event via link eksternal organizer                     |    ✅    | ✅  | Bukan fitur RSVP internal — sekadar link registrasi yang dicantumkan organizer |
| Isi feedback setelah event berlangsung                              |    ✅    | ✅  | Kecuali untuk event yang ia organisir sendiri                                  |
| Buat event baru                                                     |    ❌    | ✅  | Standard melihat modal "Upgrade ke Pro untuk membuat event"                    |
| Edit event milik sendiri                                            |    ❌    | ✅  | Hanya saat status draft atau rejected                                          |
| Hapus event milik sendiri                                           |    ❌    | ✅  | Hanya saat status draft atau rejected                                          |
| Batalkan event milik sendiri                                        |    ❌    | ✅  | Bisa kapan saja sebelum/sesudah tayang                                         |
| Lihat feedback & ringkasan statistik event miliknya                 |    ❌    | ✅  | Hanya untuk event yang ia organisir                                            |

## Journey 1: Member Menghadiri Event — 🅢🅟 — 📱 Mobile (✅ Done tanggal : 22 Juli 2026)

1. **Browse & filter** — Member membuka menu Events, melihat listing event global. Bisa memfilter berdasarkan tipe (Offline/Online/Hybrid) atau mencari nama kota/area di field venue untuk mempersempit pilihan.
2. **Lihat detail event** — Tap salah satu event untuk melihat judul, deskripsi, foto venue, kategori, tanggal & jam, info venue (nama+alamat untuk offline, nama platform+link untuk online, keduanya untuk hybrid), dan link pendaftaran eksternal jika organizer mencantumkannya.
3. **Bookmark (opsional)** — Tap ikon bookmark, event masuk ke halaman "Tersimpan" untuk dibuka lagi nanti.
4. **Tambah ke jadwal (opsional)** — Tap "Tambah ke Jadwal": event masuk ke halaman "Jadwalku" in-app, member bisa mengaktifkan reminder (push notification sebelum event) dan menambahkan catatan pribadi. Di langkah yang sama bisa sekaligus "Tambah ke Google/Apple Calendar" (unduh file `.ics` atau deep link) — kedua opsi bisa dipakai bersamaan.
5. **Daftar/hadir** — Jika event menyediakan link pendaftaran eksternal (Eventbrite, Zoom, dsb), member mengetuknya untuk mendaftar/bergabung di luar aplikasi. Untuk event offline tanpa link pendaftaran, member cukup datang ke venue yang tercantum.
6. **Bagikan (opsional)** — Tap tombol Share, pilih WhatsApp/media sosial/copy link. Platform mencatat aksi share ini.
7. **Event berlangsung** — Member menghadiri/mengikuti event sesuai tipe & venue-nya.
8. **Isi feedback (setelah event selesai)** — Begitu waktu mulai event (`event_date` + `event_time`, sesuai `timezone` event) sudah lewat, banner "Bagaimana pengalamanmu di event ini?" muncul di halaman event. Member mengisi:
   - Rating keseluruhan (1–5, wajib)
   - Rating venue/lokasi (1–5, opsional)
   - Rating penyelenggaraan (1–5, opsional)
   - Rekomendasi: Ya/Tidak — "Apakah kamu akan datang lagi ke event dari organizer ini?" (opsional)
   - Komentar bebas (opsional, tunduk filter kata terlarang platform)
   - Opsi "Kirim sebagai anonim" — organizer tetap bisa membaca isi feedback tapi tidak melihat identitas pengisi
9. **Edit/hapus feedback (opsional)** — Selama masih dalam window waktu feedback (default 30 hari sejak event selesai, bisa dikonfigurasi Superadmin), member bisa mengedit atau menghapus feedback miliknya sendiri. Tombol berubah jadi "Edit Feedback" setelah submit pertama.

**Selesai:** Member sudah menghadiri/mengikuti event dan (opsional) memberi feedback yang tersimpan untuk dilihat organizer.

### Empty/Error State di Journey Ini

- **Belum saatnya feedback** — Sebelum waktu mulai event lewat, tombol feedback belum muncul sama sekali.
- **Window feedback sudah lewat** — Setelah 30 hari (atau durasi yang dikonfigurasi Superadmin) berlalu, feedback baru tidak bisa disubmit lagi; feedback yang sudah ada tetap tersimpan dan bisa dilihat member yang bersangkutan.
- **Event dibatalkan** — Feedback tidak bisa diisi untuk event berstatus Dibatalkan. Jika event yang sudah pernah dibookmark/dijadwalkan tiba-tiba dibatalkan organizer/superadmin, member mendapat notifikasi pembatalan.
- **Organizer tidak bisa feedback event sendiri** — Jika member adalah organizer dari event tersebut, tombol/opsi feedback tidak tersedia untuknya (berlaku per-event, bukan per-role — organizer event lain tetap boleh feedback ke event tersebut).
- **Guest (belum login)** — Hanya bisa melihat event published; tidak bisa bookmark, menjadwalkan, atau mengisi feedback.
- **Duplikasi submit** — Satu member hanya bisa mengirim satu feedback per event; percobaan submit kedua kali akan ditolak (harus edit feedback yang sudah ada).
- **Feedback dihapus admin** — Kalau organizer/Superadmin menghapus feedback member (moderasi), feedback tersebut tidak bisa dikembalikan oleh member.

## Journey 2: Pro Member Membuat & Mengelola Event — 🅟 — 📱 Mobile (✅ Done tanggal : 22 Juli 2026)

1. **Mulai buat event** — Pro member membuka menu Events, tap "Buat Event".
2. **Isi detail** — Pilih tipe (Offline/Online/Hybrid), isi judul, deskripsi, kategori, upload foto venue (bisa lebih dari satu, diunggah terpisah sebelum submit), isi info venue sesuai tipe (nama venue+alamat untuk offline; nama platform+link untuk online; keduanya untuk hybrid — field venue adalah string bebas, tidak terikat data koordinat/maps), atur tanggal & jam, dan opsional menambahkan link pendaftaran eksternal (mis. Eventbrite/Zoom).
3. **Publish/Submit** — Tergantung konfigurasi platform:
   - **Auto-publish aktif:** tap "Publish" → event langsung tayang, langsung terlihat semua user dan bisa dibagikan.
   - **Perlu persetujuan:** tap "Submit" → status berubah "Menunggu Persetujuan", banner "Event sedang direview oleh tim kami" muncul. Event terkunci (tidak bisa diedit) selama menunggu, tapi organizer bisa menarik kembali submisi untuk revisi. Superadmin mereview lalu approve (event jadi Tayang, organizer dapat notifikasi) atau reject dengan alasan (event kembali ke status bisa-diedit, organizer dapat notifikasi + alasan, lalu bisa disubmit ulang).
4. **Edit event** — Hanya bisa dilakukan saat status **draft** (bebas edit semua field) atau **rejected** (edit lalu submit ulang). Event yang sudah **Menunggu Persetujuan** atau sudah **Tayang** tidak bisa diedit dari mobile.
5. **Hapus event** — Hanya bisa untuk event berstatus **draft** atau **rejected**. Event yang sudah tayang tidak bisa dihapus, hanya bisa dibatalkan.
6. **Batalkan event** — Organizer bisa membatalkan event miliknya sendiri kapan saja (baik masih draft maupun sudah tayang), dengan mengisi alasan pembatalan. Jika event yang sudah tayang dibatalkan, semua user yang sudah bookmark/menjadwalkan event tersebut mendapat notifikasi pembatalan.
7. **Lihat feedback & ringkasan** — Di halaman event miliknya, organizer membuka tab "Feedback" untuk melihat:
   - Ringkasan: rata-rata rating, persentase yang menjawab "akan datang lagi", distribusi rating (jumlah bintang 5, 4, 3, dst)
   - Daftar feedback satu per satu (nama pengisi tersembunyi untuk yang kirim anonim, tapi isi komentar tetap terlihat)
   - Notifikasi setiap kali ada feedback baru masuk (bisa dimatikan lewat preferensi notifikasi)

**Selesai:** Event Pro member berhasil dipublikasikan (langsung atau via approval), dikelola sesuai status, dan organizer bisa memantau feedback dari peserta setelah event selesai.

### Catatan Penting — Tidak Ada Manajemen Attendee/RSVP Internal

`EVENT_RULES.md` **tidak mendokumentasikan** fitur manajemen attendee di dalam aplikasi — tidak ada sistem RSVP internal, tidak ada daftar peserta yang bisa di-approve/reject organizer, dan tidak ada fitur "kirim pengumuman ke attendee". Pendaftaran/kehadiran sepenuhnya bergantung pada link eksternal (Eventbrite, Zoom, dll) yang dicantumkan organizer di field event — pengelolaannya terjadi di luar platform KAI App. Yang tersedia sebagai bentuk terdekat dari "melihat siapa yang tertarik/terlibat" hanyalah **daftar & ringkasan feedback** pasca-event (poin 7 di atas), bukan daftar attendee pra-event. Bagian ini sengaja tidak diisi dengan flow rekaan — kalau fitur ini memang ada di modul lain atau versi mendatang, perlu dicek ulang ke dokumen sumber yang lebih baru.

## Keterlibatan Admin — 💻 Web/Backoffice

**Superadmin — Approval event (jika mode manual approval aktif):**

1. Menerima notifikasi saat Pro member submit event baru.
2. Membuka detail event di antrian review.
3. **Approve** → event berubah jadi Tayang, organizer dapat notifikasi.
4. **Reject** → mengisi alasan penolakan, event dikembalikan ke organizer (bisa diedit & disubmit ulang), organizer dapat notifikasi beserta alasan.

**Superadmin — Kelola event tanpa batasan status:**

- Bisa membuat event langsung dari backoffice — langsung berstatus Tayang, tanpa approval.
- Bisa mengedit event **apapun**, kapan saja, tanpa terikat batasan status yang berlaku untuk organizer di mobile.
- Bisa menghapus event apapun secara permanen dari backoffice.
- Bisa membatalkan event siapapun.

**Superadmin — Moderasi feedback:**

- Bisa melihat semua feedback dari event manapun (termasuk identitas pengisi anonim — meski poin ini tidak dijabarkan eksplisit di `EVENT_RULES.md`, hanya disebutkan Superadmin bisa lihat "semua feedback, event manapun" untuk monitoring & moderasi; perlu konfirmasi eksplisit ke sumber lain jika detail identitas-anonim-ke-superadmin perlu dipastikan).
- Bisa menghapus feedback apapun (hard delete, permanen) — misalnya karena pelanggaran atau spam.

**Superadmin — Pengaturan platform terkait event:**

- Menentukan mode publikasi: auto-publish atau perlu persetujuan.
- Mengatur window waktu submit feedback (default 30 hari).

**Admin Regional:** `EVENT_RULES.md` tidak menyebutkan peran Admin Regional sama sekali dalam modul event — event bersifat global dan seluruh kewenangan approval/moderasi yang terdokumentasi ada di tangan Superadmin. Tidak ada pemisahan cakupan wilayah untuk event pada dokumen ini, jadi bagian ini tidak diisi dengan asumsi scoping regional.

## Di Luar Cakupan Standard & Pro

- **Standard tidak bisa membuat, mengedit, menghapus, atau membatalkan event** — murni benefit Pro; Standard yang mencoba diarahkan ke ajakan upgrade.
- **Member (Standard maupun Pro) tidak bisa menyetujui/menolak event miliknya sendiri atau event orang lain** — approval mutlak di tangan Superadmin (kalau mode manual approval aktif).
- **Member tidak bisa mengedit event yang sudah Tayang dari mobile** — perubahan pada event live hanya bisa dilakukan Superadmin dari backoffice.
- **Member tidak bisa melihat feedback event orang lain** — feedback bukan konten publik; organizer hanya melihat feedback event miliknya sendiri, user biasa tidak bisa melihat feedback siapapun kecuali miliknya sendiri.
- **Organizer tidak bisa memberi feedback untuk event miliknya sendiri** — batasan berlaku per-event.
- **Tidak ada manajemen attendee/RSVP internal (approve/reject peserta) maupun fitur kirim pengumuman ke attendee** — tidak terdokumentasi di `EVENT_RULES.md` sama sekali; pendaftaran/kehadiran sepenuhnya via link eksternal yang dikelola organizer di luar platform.
- **Tidak ada konsep kapasitas/kuota event ("event full") atau status "pendaftaran ditutup"** — tidak disebutkan di sumber; kapasitas kemungkinan diatur di sisi platform eksternal (mis. Eventbrite), bukan oleh KAI App.
- **Guest (belum login) tidak bisa bookmark, menjadwalkan, atau mengisi feedback** — hanya bisa melihat event yang sudah published.

## Edge Case & Catatan Tambahan

- **Alur status event lengkap:** Draft → (submit) → Menunggu Persetujuan → Approve → Tayang → (cancel) → Dibatalkan; atau Menunggu Persetujuan → Reject → Draft (bisa edit & submit ulang) → (hapus) → terhapus permanen. Draft yang belum pernah disubmit juga bisa langsung dihapus kapan saja oleh organizer.
- **Venue adalah string bebas** — tidak terikat ke data koordinat/maps; cukup informatif (nama venue+alamat untuk offline, nama platform+link untuk online).
- **Anonymous feedback:** organizer tetap bisa membaca isi komentar/rating feedback anonim, hanya identitas pengisinya yang disembunyikan dari organizer.
- **Window feedback dikonfigurasi Superadmin** (default 30 hari sejak event selesai) — berlaku sebagai pengaturan platform, bukan per-event.
- **Feedback hanya berlaku untuk event berstatus published** (termasuk yang tanggalnya sudah lewat); event yang dibatalkan tidak bisa menerima feedback baru.
- **Satu feedback per user per event** — tidak bisa submit dua kali; harus edit feedback yang sudah ada jika ingin mengubah.
