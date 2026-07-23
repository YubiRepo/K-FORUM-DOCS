# Community — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md).

## Ringkasan Domain

Community adalah grup yang dibuat oleh Member Pro (gated benefit `create_community`); pembuatnya otomatis jadi **leader**. Bergabung dan posting di dalam komunitas **terbuka untuk semua member** (Standard & Pro) — tier global cuma menggerbang **siapa yang boleh membuat** komunitas, bukan siapa yang boleh berpartisipasi.

Yang membuat domain ini berbeda dari domain lain: di atas tier global Standard/Pro, Community punya **sistem role lokalnya sendiri** — `leader`, `moderator`, `member` — yang scope-nya per-komunitas. Role ini **global & fixed** (tidak bisa ada role baru per komunitas), tapi **permission per role bisa berbeda antar komunitas** dan **diberikan oleh leader komunitas** (bukan oleh admin platform). Seorang member Standard sekalipun bisa dipromosikan jadi moderator di komunitas milik orang lain — status moderator tidak butuh tier Pro, yang butuh Pro hanya **mendirikan** komunitas (jadi leader dari komunitas baru).

Domain ini juga membawahi dua sub-fitur yang scoped per komunitas: **Papan Pengumuman** (info resmi read-only dari pengelola) dan **Schedule Komunitas** (agenda/kalender internal + RSVP) — keduanya beda dari modul Announcement platform dan modul Event.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Browse & cari komunitas (public & private) | ✅ | ✅ | Private terlihat tapi terkunci |
| Join komunitas public | ✅ | ✅ | Auto-join |
| Request join komunitas private | ✅ | ✅ | Menunggu approval leader/moderator |
| Terima undangan / redeem share link | ✅ | ✅ | Sama untuk kedua tier |
| Posting, komentar, like, save, share di komunitas | ✅ | ✅ | Selama punya permission `post_content` (default semua member) |
| Lihat pengumuman komunitas | ✅ | ✅ | Read-only |
| Lihat kalender & RSVP schedule komunitas | ✅ | ✅ | Occurrence aktif & belum lewat saja |
| Leave komunitas | ✅ | ✅ | Leader (selalu Pro) wajib transfer ownership dulu |
| **Membuat komunitas (jadi leader)** | ❌ | ✅ | Butuh benefit `create_community` |
| Jadi **moderator** di komunitas orang lain | ✅ | ✅ | Tidak digerbang tier — leader bisa promosikan Standard maupun Pro jadi moderator |
| Mengundang user / buat share link, approve join-request, kick/ban, publish pengumuman, buat schedule | ❌* | ✅ (default sbg leader) | *Standard/Pro bisa saja punya akses ini kalau di-promosikan moderator dan diberi permission terkait oleh leader |
| Customize permission per role di komunitas miliknya | ❌ | ✅ | Hanya berlaku untuk komunitas yang dia pimpin |
| Lihat analytics komunitas | ❌ | ✅ | Benefit `view_analytics` |

> Catatan: kolom "Pro" di atas mengasumsikan Pro yang bersangkutan **adalah leader** komunitas itu. Pro yang cuma jadi anggota biasa di komunitas lain punya hak persis sama seperti Standard (baris atas) — perbedaan Standard vs Pro murni soal **bisa/tidaknya mendirikan komunitas sendiri**.

## Journey 1: Member Join & Aktif di Komunitas — 🅢🅟 — 📱 Mobile

1. Member membuka tab **Community**, melihat daftar/discovery komunitas — bisa filter by kategori dan region (region cuma label/filter, bukan pembatas akses).
2. Cari komunitas via search. Komunitas **public** dan **private** sama-sama muncul di hasil; private ditandai terkunci (gembok) — kontennya tidak bisa dilihat sebelum jadi anggota.
3. **Join komunitas public** → tekan "Join" → langsung auto-join, langsung jadi anggota aktif, dapat role `member`, notifikasi ke pengelola (`member_joined`).
4. **Join komunitas private** → tekan "Request to Join", isi pesan opsional → status jadi "⏳ Menunggu persetujuan". Leader/moderator komunitas menerima notifikasi dan approve/reject. Kalau approve → jadi anggota aktif; kalau reject → request ditutup, member bisa request lagi nanti.
5. Setelah jadi anggota: bisa **posting** (teks plain, maks 10 gambar), **komentar** (plain text, maks 1 level reply — reply ke reply ditolak), **like/unlike**, **save/bookmark** post, dan **share** keluar via deep link (mis. WhatsApp — repost internal belum ada, Fase 2).
6. **Lihat Papan Pengumuman**: hanya pengumuman berstatus `published` dan belum `expires_at` yang tampil, urut pinned → priority (`important` di atas `normal`) → terbaru. Pengumuman **read-only** — tidak ada like/comment. Pengumuman `important` memicu push notification saat pertama publish; `normal` cuma in-app/badge.
7. **Lihat Schedule Komunitas**: buka kalender komunitas (window per bulan/tanggal), lihat semua occurrence termasuk hasil expand dari agenda berulang (mis. "tiap Sabtu"). Occurrence yang sudah dibatalkan tampil dengan tanda batal.
8. **RSVP** satu occurrence: pilih `going` / `maybe` / `not_going`. Hanya bisa untuk occurrence yang **aktif dan belum lewat** — occurrence batal atau sudah lewat menolak RSVP baru (tombol disabled). RSVP bisa diubah kapan saja selama occurrence masih berlaku.
9. **Leave komunitas**: tekan "Leave" di halaman komunitas → keanggotaan & role dihapus, `member_count` berkurang. **Kecuali** member tersebut adalah leader — sistem menolak dan mengarahkan ke alur transfer ownership dulu (lihat Journey 3).

**Empty/error state yang relevan:**
- Komunitas private tanpa jadi anggota → layar terkunci, hanya tombol "Request to Join", tidak ada preview konten.
- Member berstatus `banned` di suatu komunitas → tidak bisa join ulang sampai di-unban oleh leader/moderator.
- Pengumuman lewat `expires_at` → hilang dari daftar anggota (tetap tersimpan di arsip pengelola).
- RSVP untuk occurrence yang batal/lewat → tombol RSVP nonaktif, ditampilkan status "sudah lewat"/"dibatalkan".

## Journey 2: Member Menerima Invite ke Komunitas — 🅢🅟 — 📱 Mobile

Selain browse→join dan join-request di atas, member bisa masuk komunitas lewat dua jalur tambahan: **undangan personal** dan **share link/join code**.

1. **Undangan personal**: leader/moderator komunitas mengundang member by username/nama. Member menerima notifikasi `community_invitation_received` dan melihat daftar undangan masuk di **"Me → Community Invitations"**.
2. Member membuka detail undangan (nama komunitas, pesan opsional dari pengundang) → pilih **Accept** atau **Reject**.
   - **Accept** → langsung jadi anggota aktif, **bypass approval** sekalipun komunitas itu private (tidak membuat join-request). Role `member` langsung ter-assign.
   - **Reject** → undangan ditutup, tidak ada efek lain.
3. Undangan punya masa berlaku **default 7 hari**; lewat itu berstatus `expired` dan tidak bisa di-accept lagi (leader perlu mengundang ulang).
4. **Share link/join code**: member membuka link (mis. `kai.app/c/join/{code}`) yang dibagikan lewat WhatsApp/grup lain. App me-resolve link dan mencoba redeem.
   - Jika link `requires_approval=false` (default) atau komunitasnya public → **langsung join**.
   - Jika komunitas private dan link `requires_approval=true` → **membuat join-request** (sama seperti alur request manual, menunggu approval leader/moderator).
   - Jika member sudah jadi anggota → no-op, langsung diarahkan ke komunitas (bukan error).
   - Jika link sudah mati (dicabut/kedaluwarsa/kuota habis) → pesan jelas "Link sudah tidak berlaku" (410).
   - Member berstatus `banned` di komunitas tersebut ditolak di semua jalur ini (undangan lama, redeem link) sampai di-unban.

## Journey 3: Pro Member Membuat & Mengelola Komunitas — 🅟 — 📱 Mobile

1. Pro member membuka **"Create Community"** → isi nama, deskripsi, avatar, **kategori** (wajib, dipilih dari daftar yang dikelola superadmin), **visibility** (public/private), region (opsional).
2. Sistem cek benefit `create_community` ✅ → komunitas dibuat berstatus `active` → Pro member otomatis jadi **leader**, permission leader/moderator/member di-copy dari template default (`community_role_permissions_template`, dikelola superadmin).
3. **Mengelola anggota:**
   - **Undang user spesifik**: cari user terdaftar by username/nama → kirim undangan (bisa dibatalkan selama masih pending; tidak bisa mengundang user yang sudah jadi anggota atau berstatus banned). (not yet implemented fully)
   - **Buat share link**: atur `requires_approval`, `expires_at`, `max_uses` opsional → dapat kode/URL untuk disebar; bisa dicabut kapan saja (`is_active=false`), bisa punya beberapa link aktif sekaligus. (implemented, but button join masih belum berubah ketika user sudah join)
   - **Approve/reject join-request** komunitas private.
   - **Promote member → moderator**: pilih anggota, assign role `moderator`, atur permission spesifik untuk komunitas ini (mis. `manage_members`, `moderate_posts`, `manage_community_announcement`, `manage_community_schedule`) via bulk-assign — permission ini **tidak otomatis berlaku di komunitas lain** milik moderator yang sama. ()
   - **Kick/ban** anggota yang melanggar aturan.
4. **Customize permission per role**: leader bisa membuka Roles & Permissions komunitasnya, mencabut/menambah permission untuk role `moderator` (mis. cabut `delete_content` dari moderator) — perubahan berlaku langsung.
5. **Moderasi konten**: hapus post/komentar yang melanggar (butuh `moderate_posts`/`delete_content`, default dimiliki leader).
6. **Publish Papan Pengumuman**: isi judul, body (plain text), media opsional (maks 5 gambar), `priority` (`normal`/`important`), `is_pinned`, `expires_at` opsional. Bisa disimpan sebagai `draft` dulu (hanya terlihat pembuat & pengelola) sebelum `published`. Publish dengan priority `important` memicu push notification ke semua anggota.
7. **Buat Schedule Komunitas**: isi judul, waktu mulai/selesai, `all_day`, lokasi bebas teks, **timezone (wajib diisi eksplisit saat create)**, dan `recurrence` opsional (RRULE-style, mis. `FREQ=WEEKLY;BYDAY=SA` untuk agenda tiap Sabtu — satu entry mewakili semua occurrence-nya, di-expand on-the-fly per window kalender saat dilihat anggota).
8. **Kelola occurrence**: batalkan **satu tanggal** tertentu (exception, occurrence lain di seri yang sama tidak terpengaruh) atau batalkan **seluruh agenda** (`status=cancelled`, semua occurrence tampil batal). Lihat ringkasan RSVP per occurrence (`going`/`maybe`/`not_going` + daftar nama).
9. **Edit profil komunitas**: nama, deskripsi, avatar, visibility — kapan saja.
10. **Transfer ownership**: sebelum resign/leave, leader wajib transfer ke anggota aktif lain — leader lama didemote jadi member (atau keluar), anggota baru jadi leader.
11. **Hapus komunitas**: leader (atau superadmin) bisa hapus komunitas — cleanup penuh dalam satu transaksi: post, member, role, permission komunitas, pengumuman, schedule, RSVP, undangan, dan share link semuanya dibersihkan.

## Keterlibatan Admin — 💻 Web/Backoffice

Perlu ditegaskan: **leader dan moderator adalah role lokal komunitas**, diberikan oleh Pro member yang jadi leader — **bukan** role admin platform. Leader tidak bisa menyentuh komunitas lain, tidak bisa mengubah daftar kategori master, dan tidak bisa override role-permission di luar komunitasnya sendiri.

**Superadmin (KAI Pusat):**
- Melihat semua komunitas di seluruh platform, suspend/archive/hapus komunitas yang melanggar kebijakan.
- Menangani komunitas berstatus `orphaned` (terjadi kalau leader menghapus akun tanpa transfer ownership dulu) — assign leader baru atau archive komunitas.
- Moderasi konten global (post, pengumuman, schedule, undangan, share link) lintas semua komunitas sebagai override — sejalan dengan wewenang moderasi konten komunitas yang sudah ada.
- Mengelola daftar **kategori komunitas** (CRUD) yang dipilih leader saat create/edit komunitas.
- Mengelola **template default permission** (`community_role_permissions_template`) yang di-copy ke tiap komunitas baru saat dibuat.

**Admin Regional:** dokumen sumber (`COMMUNITY_RULES.md`, `COMMUNITY_ANNOUNCEMENT_SCHEDULE_RULES.md`, `COMMUNITY_INVITE_RULES.md`) **tidak menyebutkan** wewenang khusus Admin Regional atas domain Community — tabel "Siapa Bisa Apa" di ketiga dokumen hanya mencantumkan Superadmin secara eksplisit. Ini kemungkinan konsisten dengan prinsip "region cuma label & filter, bukan pembatas akses" pada modul ini, tapi tetap perlu diklarifikasi ke tim produk apakah Admin Regional dimaksudkan punya wewenang moderasi/monitoring atas komunitas di wilayahnya (mengikuti pola domain lain) atau memang murni terpusat di Superadmin untuk domain ini.

## Di Luar Cakupan Standard & Pro

- **Standard tidak bisa membuat komunitas** — mendirikan komunitas (jadi leader) mutlak butuh benefit `create_community` (Pro).
- **Tidak bisa membuat role baru** selain `leader`/`moderator`/`member` — role bersifat global & fixed di seluruh platform, tidak bisa ditambah per komunitas (mis. tidak ada "senior moderator").
- **Override permission per-user** (di luar per-role-per-komunitas) — belum didukung sistem (masih asumsi terbuka di `COMMUNITY_RULES.md`).
- **Reaction selain like, nested reply >1 level, repost internal** — ditunda ke Fase 2.
- **Mengubah detail satu occurrence** (mis. ganti jam khusus untuk satu tanggal tanpa membatalkannya) — ditunda Fase 2 (`exception.type=modified` sudah disiapkan sebagai hook tapi belum aktif).
- **Attendance/check-in resmi** — RSVP di Schedule Komunitas murni indikator minat, bukan catatan kehadiran otoritatif; fitur absensi beneran ditunda ke fase berikutnya kalau dibutuhkan.
- **Invite lewat email** ke non-user — ditunda Fase 2 (saat ini undangan cuma bisa ke user yang sudah terdaftar in-app).
- **Verifikasi badge resmi komunitas** — di luar cakupan dokumen Community; kalau komunitas punya badge terverifikasi, pemberiannya murni wewenang Superadmin lewat modul terpisah (lihat [09_VERIFICATION_BADGE_JOURNEY.md](09_VERIFICATION_BADGE_JOURNEY.md)), bukan hak leader/moderator komunitas.
- **Force-remove leader aktif oleh Superadmin** (di luar alur `orphaned`) — dokumen sumber hanya menjelaskan dua jalur pergantian leader: leader transfer ownership secara sukarela, atau leader menghapus akun sehingga komunitas jadi `orphaned` dan ditangani superadmin. Tidak ada alur superadmin mencopot leader aktif secara paksa tanpa lewat kondisi itu — kalau ini dibutuhkan, perlu klarifikasi tambahan ke tim produk.
- **Batas jumlah komunitas per Pro member** — belum dibatasi (asumsi terbuka, bisa jadi benefit terukur nanti).

## Edge Case & Catatan Tambahan

- **Leader tidak bisa leave tanpa transfer ownership dulu.** Kalau leader menghapus akunnya sendiri (bukan leave biasa), komunitas masuk status `orphaned` sampai superadmin assign owner baru atau archive.
- **Member `banned` menutup semua jalur masuk**: tidak bisa join biasa, tidak bisa diundang, tidak bisa accept undangan lama, tidak bisa redeem share link — sampai di-unban oleh leader/moderator.
- **Pengumuman yang expired tidak dihapus** — hanya tersaring dari tampilan anggota, tetap ada di arsip pengelola/backoffice.
- **Rekonstruksi occurrence schedule wajib pakai `timezone` milik entry**, bukan zona waktu server atau device — bug ini pernah ditemukan & diperbaiki (2026-07-17): membaca jam/tanggal langsung dari kolom `start_at` tanpa mengonversi ke `entry.timezone` bisa salah untuk agenda yang jam lokalnya menyebrang tengah malam UTC.
- **Share link dengan `requires_approval=true`**: `use_count` naik saat **join-request dibuat**, bukan saat di-approve — jadi kuota link membatasi jumlah pendaftar, bukan jumlah yang akhirnya diterima.
- **Satu undangan pending per pasangan (komunitas, invitee)** — mengundang ulang saat masih ada undangan pending akan menggantikan/menyegarkan undangan lama, bukan menumpuk duplikat.
- **Komunitas berbayar/berbadge/berregion** tetap mengikuti aturan visibility yang sama — region cuma label & filter, tidak membatasi siapa yang bisa menemukan atau join komunitas tersebut.

## Note
- ~~upload avatar komunitas beluma da presign url / atau mungkin mobile salah endpoint, sehingga gagal upload avatar komunitas~~ **Sebagian diperbaiki** (22 Jul 2026, k-forum-api): endpoint presign avatar generic (`POST /mobile/communities/media/avatar/presign`, tanpa `community_id`) ditambahkan khusus untuk mode create — sebelumnya memang belum ada sama sekali. Kebetulan path ini persis sama dengan yang sudah di-hardcode mobile, jadi mode **create** kemungkinan langsung jalan tanpa ubah kode mobile. Mode **edit** (ganti avatar komunitas existing) masih perlu fix di mobile (endpoint scoped `:community_id/media/avatar/presign` sudah benar di backend, tinggal dipakai mobile dengan benar). Detail: `COMMUNITY_MODULE_ISSUES.md` Issue 1.
- invite via account name masih belum sempurna, hasil invite tidak muncul di list invitation, padahal sudah dikirimkan. bahkan dimana list invitatiionnya, *(belum dikerjakan — root cause 100% di `k_forum` mobile, backend sudah benar & lengkap. Detail: `COMMUNITY_MODULE_ISSUES.md` Issue 2)*
- invite via link, ketika sudah join, tombol join masih muncul, seharusnya hilang, atau mungkin langsng diarahkan ke halaman komunitas *(belum dikerjakan — 100% di `k_forum` mobile. Detail: `COMMUNITY_MODULE_ISSUES.md` Issue 3)*
- ~~link share post masih belum sesuai domain url yg di pakai.~~ **Sudah diimplementasikan** (22 Jul 2026, k-forum-api): domain diganti ke `k-forum-app.yubicom.co.id` (config baru `AppConfig.AppBaseURL`, env `APP_BASE_URL`) dan path share-post dibetulkan jadi `/posts/{id}` (sebelumnya `/c/{communityId}/p/{postId}`, salah bentuk). Invite link juga ikut dibetulkan domainnya (path-nya sebelumnya sudah benar). Sudah diaudit ke modul lain (News/Event/Directory) — tidak ada pola hardcode domain salah yang sama. Detail: `COMMUNITY_MODULE_ISSUES.md` Issue 4.
- ~~moderator tidak bisa menghapus post, padahal sudah ada permission manage_content.~~ **Sudah diimplementasikan** (22 Jul 2026, k-forum-api): endpoint mobile baru `POST /mobile/communities/posts/:post_id/remove` ditambahkan (sebelumnya tidak ada jalur manapun, mobile maupun web, yang mengizinkan moderator biasa menghapus post orang lain), plus pengecekan permission per-komunitas (`CanModerateInCommunity`) ditambahkan langsung di `RemovePostUseCase` (sebelumnya tidak ada validasi ini sama sekali di usecase). Detail: `COMMUNITY_MODULE_ISSUES.md` Issue 5.
- ~~leader comunity dapet expose semua comunity di backoffice, seharusnya hanya comunity yang dia pimpin saja. ( di sini juga belum lebih spesifik sebagai halaman manage comunity untuk si leader, dan moderator,)~~ **Sudah diimplementasikan** (22 Jul 2026, k-forum-api + k-forum-backoffice): endpoint baru `GET /web/communities/mine` + otorisasi `CanModerateInCommunity` ditambahkan ke detail/member/moderate-member usecase (sebelumnya blanket `RequireAdmin()`, tidak ada jalur leader/moderator sama sekali). Halaman baru `/community/mine` di backoffice untuk self-service leader/moderator (list komunitas dikelola → drill-down kelola member). Sekalian ditemukan & diperbaiki lubang keamanan: halaman admin `/community` sebelumnya bisa ditembus member Pro biasa (benefit `create_community` di-OR dengan role check), sekarang murni admin-only. Detail: `COMMUNITY_MODULE_ISSUES.md` Issue 6.
- sudah tidak jadi leader, tapi masih punya session sebagai leader di eks comunity yang dia pimpin, seharusnya di logout otomatis. memang setelah sekian lama dia hilang, krn redis kmngkinan ( gapapa dulu). *(sengaja ditunda sesuai permintaan — root cause sudah ditemukan (cache principal tidak di-invalidate saat transfer/hapus komunitas), tapi belum dikerjakan. Detail: `COMMUNITY_MODULE_ISSUES.md` Issue 7)*
