# Announcement — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Announcement/ANNOUNCEMENT_RULES.md`.

## Ringkasan Domain

Announcement adalah komunikasi **satu arah** (one-way) dari Superadmin (KAI Pusat, scope global) atau Admin Regional (scope wilayahnya sendiri) ke semua user di scope tersebut — bukan komunikasi 1-ke-1, dan bukan konten yang perlu approval sebelum tayang (admin publish langsung). Di domain ini, **member (Standard maupun Pro) murni berperan sebagai penerima** — tidak ada satu pun aksi "membuat/mengedit/menghapus announcement" yang menjadi hak member, terlepas dari tier apa pun. Perbedaan yang ada di domain ini semata soal **scope penerima** (global vs regional) dan **prioritas/tipe announcement** (disaster, system, urgent, info) — sama sekali bukan soal Standard vs Pro.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Menerima push notification | ✅ | ✅ | Perilaku push mengikuti prioritas (CRITICAL/HIGH/MEDIUM/LOW), sama untuk semua tier |
| Menerima in-app banner di home | ✅ | ✅ | |
| Menerima email (untuk tipe/prioritas tertentu) | ✅ | ✅ | |
| Melihat daftar & baca detail announcement (global + region sendiri) | ✅ | ✅ | |
| Filter berdasarkan tipe/tanggal, cari riwayat | ✅ | ✅ | |
| Opt-out notifikasi non-critical | ✅ | ✅ | Semua kecuali CRITICAL bisa dimatikan lewat notification preferences |
| **Membuat announcement** | ❌ | ❌ | Bukan hak member sama sekali — eksklusif Superadmin/Admin Regional |
| **Edit/hapus/archive announcement** | ❌ | ❌ | Eksklusif admin, tidak ada pengecualian tier |

## Journey 1: Member Menerima & Membaca Announcement — 🅢🅟 — 📱 Mobile

1. **Push notification** terkirim segera setelah admin publish. Perilaku mengikuti prioritas:
   - `CRITICAL` — bypass notification settings, force push ke semua device meski notifikasi dimatikan.
   - `HIGH` — push segera, tetap menghormati setting notifikasi user.
   - `MEDIUM` — dikirim dalam batch, bisa delay beberapa menit.
   - `LOW` — tidak ada push, hanya muncul di in-app.
2. Badge counter muncul di ikon bell, disertai sound/vibration sesuai prioritas.
3. **In-app banner** muncul di home screen: sticky (selalu di atas) dan highlight merah untuk `CRITICAL`, pin di atas listing dengan highlight orange untuk `HIGH`, posisi normal di listing untuk `MEDIUM`, di bawah listing tanpa highlight untuk `LOW`. Member bisa tap untuk lihat detail, atau swipe untuk dismiss.
4. **Email** dikirim untuk kombinasi tipe/prioritas tertentu: semua `DISASTER`, `SYSTEM` dengan prioritas HIGH (outage/critical), `URGENT` dengan prioritas CRITICAL. Untuk `INFO`, tidak ada email kecuali member mengaktifkan preferensi khusus.
5. **Tab "Announcements"** dedicated: filter berdasarkan tipe (disaster/system/urgent/info), menampilkan unread count, dan untuk tipe disaster/urgent diurutkan **oldest first**.
6. **Read tracking**: begitu member membuka detail, otomatis ditandai sudah dibaca (hitungan pembaca ini yang ditampilkan sebagai statistik ke admin, bukan ke member).
7. **Khusus tipe DISASTER** (mis. gempa, banjir, kebakaran): detail bisa memuat info epicenter/lokasi kejadian, kedalaman, area terdampak, **lokasi evakuasi & kapasitasnya**, **rute evakuasi**, dan **helpline** darurat — ditampilkan lewat peta di halaman detail. Biasanya sifatnya informasional (tidak ada tombol aksi khusus), tapi memuat info krusial untuk keselamatan.
8. **History**: member bisa melihat announcement lama (termasuk yang sudah `archived`/`expired`) di halaman riwayat, bisa dicari berdasarkan keyword.
9. **Scope penerima**: member hanya melihat announcement `global` + announcement `regional` sesuai region member saat ini. Jika member pindah region, ia mulai menerima announcement region barunya; announcement region lama yang sudah diterima tetap tersimpan di riwayat.

**Selesai:** member sudah menerima notifikasi dan membaca detail announcement yang relevan dengan scope-nya; tidak ada "tahap akhir" formal karena member tidak melakukan aksi lanjutan apa pun terhadap announcement (murni penerima).

## Keterlibatan Admin — 💻 Web/Backoffice

1. **Superadmin** bisa membuat announcement scope `global` (semua region) maupun scope `regional` untuk region mana pun, langsung publish tanpa approval, dan mengedit/menghapus announcement miliknya sendiri kapan saja. Superadmin juga bisa melihat semua announcement (termasuk yang dibuat Admin Regional) tapi tidak bisa edit/hapus milik Admin Regional.
2. **Admin Regional** hanya bisa membuat announcement scope `regional` untuk **region miliknya sendiri** (tidak bisa untuk region lain, dan tidak bisa scope `global`), langsung publish tanpa approval. Bisa lihat announcement global (read-only, tidak bisa edit/hapus) plus kelola penuh announcement regionnya sendiri.
3. **4 tipe announcement** yang tersedia:
   - **DISASTER** — subtype: earthquake, flood, landslide, tsunami, volcanic, storm, fire. Bisa dibuat Superadmin (global) atau Admin Regional (regional).
   - **SYSTEM** — subtype: maintenance, outage, degraded_service, update, incident. Bisa dibuat Superadmin (global) atau Admin Regional (regional-only, mis. maintenance server Jakarta).
   - **URGENT** — subtype: security_alert, account_action, policy_change, urgent_update. **Eksklusif Superadmin** — Admin Regional tidak bisa membuat tipe ini.
   - **INFO** — subtype: feature_launch, event, promo, policy_update, company_news, general_info. Bisa dibuat Superadmin (global) atau Admin Regional (regional).
4. **Publish langsung** memicu: query recipients dari database, ambil FCM token, kirim push (batch via Firebase Cloud Messaging, ~1-5 menit), kirim email jika berlaku, log terkirim/gagal, update `total_sent`.
5. **Archive**: manual oleh admin (mis. bencana sudah berakhir) atau otomatis via `expires_at` yang terlewati.
6. **Monitoring**: admin memantau delivery rate (total_sent vs total_recipients) dan read rate per announcement, termasuk breakdown per region untuk announcement regional.

## Di Luar Cakupan Standard & Pro

- **Membuat/broadcast announcement** — sama sekali bukan hak member, baik Standard maupun Pro; eksklusif Superadmin (global) dan Admin Regional (region sendiri).
- **Edit/hapus/archive announcement** — eksklusif admin yang membuatnya (atau Superadmin untuk kasus tertentu); member tidak punya kontrol apa pun atas konten yang sudah tayang.
- **Menargetkan user individual tertentu** — sistem ini secara eksplisit **bukan** untuk komunikasi ke user spesifik (mis. "pembayaranmu gagal"); selalu broadcast ke seluruh scope (global/regional).
- **Membuat announcement tipe URGENT** — bahkan Admin Regional tidak bisa, eksklusif Superadmin; jelas jauh di luar cakupan member.
- **Menonaktifkan notifikasi prioritas CRITICAL** — dipaksa tampil (bypass notification settings) untuk semua user, tidak bisa di-opt-out oleh siapa pun.
- **Melihat announcement region lain** — member (dan Admin Regional) hanya melihat announcement global + region sendiri, bukan soal tier tapi soal scope wilayah.

## Edge Case & Catatan Tambahan

- **Multi-bencana di region berbeda dalam waktu bersamaan**: ditangani sebagai announcement terpisah per kejadian (bukan digabung), masing-masing menjangkau audiens yang sesuai.
- **Sistem down saat butuh kirim announcement darurat**: dokumen sumber menyebut mitigasi berupa cache pesan darurat pre-populated dan SMS fallback untuk `CRITICAL`, atau komunikasi manual lewat media sosial — bukan bagian dari journey member, tapi relevan sebagai konteks keandalan sistem.
- **Tidak ada fitur "schedule for later"** — announcement hanya bisa langsung publish atau disimpan sebagai draft; tidak ada auto-renew; dan tidak bisa satu announcement menjangkau multi-region sekaligus (harus dibuat terpisah per region).
- **`ANNOUNCEMENT_RULES.md` tidak mendetailkan** apakah member bisa memberi "acknowledge" eksplisit (selain read-tracking otomatis) untuk announcement bertipe URGENT yang mensyaratkan aksi (mis. reset password) — dicatat sebagai gap dokumentasi, bukan diasumsikan ada tombol konfirmasi terpisah.
