# Domain Admin-Only — Tidak Ada Journey Member

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md).

## Kenapa Dokumen Ini Berbeda

Lima domain di bawah ini **tidak punya journey member sama sekali** — bukan cuma "minim keterlibatan admin" seperti domain lain di daftar, tapi murni 💻 Web/Backoffice. Tidak ada layar 📱 Mobile untuk domain-domain ini, bahkan tidak ada bentuk tidak-langsung seperti tombol terkunci atau prompt upgrade yang biasanya menandai fitur Pro-only. Member Standard maupun Pro tidak pernah membuka, melihat, atau memicu UI dari domain-domain ini — baik lewat aksi sengaja maupun sebagai efek samping aksi lain.

Karena itu domain-domain ini tidak layak dapat file journey penuh (yang formatnya "Journey Standard → Journey Pro → Keterlibatan Admin") — tidak ada langkah member untuk didokumentasikan. Sebagai gantinya, catatan ini menjelaskan singkat: domain itu untuk apa, siapa di internal KAI yang memakainya, kenapa member tidak pernah menyentuhnya langsung, dan — kalau relevan — bagaimana domain itu tetap memengaruhi pengalaman member secara tidak langsung meski UI-nya tidak pernah terlihat.

Lima domain: **Role & Permission System**, **System Settings**, **User Management (Backoffice)**, **Accounting**, dan **Schedule (kalender agenda internal)**.

## Role & Permission System

Role & Permission adalah mesin otorisasi platform: master daftar permission, definisi role (`usergod`, `superadmin`, `admin` region, `member`, `guest` di level sistem; `leader`, `moderator`, `member` di level komunitas), dan aturan siapa boleh assign apa ke siapa. Yang benar-benar mengelolanya adalah **Usergod** (define permission master list, satu-satunya yang bisa touch role `usergod`) dan **Superadmin** (assign permission ke system role). Di level komunitas, pengelolanya adalah **community leader** — tapi itu bagian dari fitur Community yang memang muncul di journey member (lihat [04_COMMUNITY_JOURNEY.md](04_COMMUNITY_JOURNEY.md)), bukan bagian dari modul Role & Permission itu sendiri.

Member tidak pernah membuka layar "kelola permission" — tidak ada UI mobile untuk itu. Yang member alami hanyalah **efeknya**: sistem mengecek permission (role-based) dan benefit (subscription-based) di balik layar setiap kali member mencoba aksi seperti post news atau approve konten, lalu diam-diam meng-allow/deny. Contoh dari dokumen sumber: `approve_news` hanya bisa dilakukan superadmin/admin — member pro tidak pernah melihat opsi approve itu sama sekali, bukan karena disembunyikan, tapi karena permission-nya memang tidak pernah diberikan ke role `member`.

## System Settings

System Settings adalah "satu tempat untuk konfigurasi platform-wide yang boleh berubah saat runtime tanpa deploy" — mencakup group seperti `general`, `mobile_app`, `security`, `email`, `storage`, `payment`, `moderation`, `maintenance`, dan `contact`, plus dokumen legal versioned (Terms, Privacy, Community Guidelines). Yang mengelola: **Usergod** untuk setting teknis/infrastruktur (security, email, storage), **Superadmin** untuk setting operasional/konten (general, payment, moderation, maintenance, contact, legal). Admin Regional dan Member sama sekali tidak punya akses edit di modul ini (matrix "Siapa Bisa Apa" di dokumen sumber menandai kolom Admin Region dan Member semua ❌, kecuali baca versi published dokumen legal).

Member tidak pernah membuka layar System Settings — tidak ada menu itu di app. Tapi dampaknya ke member sangat langsung meski tidak lewat UI: **Maintenance Mode** adalah contoh paling jelas — begitu Usergod/Superadmin menyalakan `maintenance_mode_enabled`, semua endpoint mobile untuk guest & member langsung balas `503` dan app menampilkan halaman maintenance, padahal member tidak pernah membuka atau tahu menu itu ada. Contoh lain: perubahan `min_version_android`/`force_update_enabled` memicu dialog "Wajib Update" di app member; perubahan rekening di group `payment` langsung muncul di halaman upgrade Pro member; publish versi Terms baru dengan `require_reacceptance=true` memaksa semua member menyetujui ulang saat login berikutnya.

## User Management (Backoffice)

User Management adalah modul tempat **Superadmin** dan **Admin Region** mengelola akun member: melihat/edit profil, mengubah status akun (active/inactive/suspended), assign & revoke role (single atau bulk, termasuk lewat CSV), assign role dengan scope region/community, reset password, dan mengubah subscription plan secara manual. Superadmin punya akses platform-wide (semua user, semua region); Admin Region hanya bisa mengelola user di region-nya sendiri dan tidak bisa suspend user, assign role superadmin, atau mengubah subscription. Member biasa dan Guest eksplisit disebutkan **tidak bisa** mengakses backoffice user management.

Member tidak pernah membuka dashboard ini — tapi ini justru domain yang efeknya paling terasa dari sisi member karena berulang kali "meng-atur ulang" akun mereka dari sisi admin. Saat Superadmin mengubah plan subscription member secara manual (`subscription.changed` di audit log, "langsung efektif tanpa menunggu konfirmasi user"), member hanya melihat efeknya: benefit Pro tiba-tiba aktif atau nonaktif di app mereka sendiri — lihat [02_SUBSCRIPTION_UPGRADE_JOURNEY.md](02_SUBSCRIPTION_UPGRADE_JOURNEY.md) untuk journey upgrade normal member sendiri (yang jalurnya beda — lewat request & approval, bukan manual admin change ini). Begitu juga saat status akun diubah jadi `suspended`: member langsung logout paksa (semua session di-invalidate) dan menerima email notifikasi berisi alasan, tanpa pernah membuka layar admin yang melakukan itu. Assign/revoke role region juga menentukan siapa yang jadi Admin Region di sisi member — tapi ini murni pekerjaan admin-ke-admin, bukan sesuatu yang member picu atau lihat.

## Accounting

Accounting adalah buku besar (ledger) pencatatan uang masuk (IN) dan keluar (OUT) untuk operasional KAI Pusat — mencatat revenue dari Subscription/Ads, plus pengeluaran manual (gaji, sewa, biaya event) yang tidak punya modul sendiri. Dokumen sumber menyebut eksplisit: **"Surface: Backoffice only (tidak ada di mobile member)"**, dan pengguna: **"Superadmin (KAI Pusat) saja"** — tidak ada Admin Region, karena modul ini "tidak menggunakan region" sama sekali (tidak ada scope/pemisahan per region).

Domain ini benar-benar tidak terhubung ke aksi member apa pun. Ini murni ledger internal — mencatat *hasil* dari transaksi yang sudah terjadi di modul lain (Subscription, Ads), bukan bagian dari alur transaksi member itu sendiri. Saat member upgrade ke Pro atau memasang iklan, member tidak berinteraksi dengan Accounting sedikit pun — modul ini hanya mencatat angka revenue tersebut di belakang layar untuk kepentingan pelaporan keuangan KAI Pusat. Tidak ada dampak langsung maupun tidak langsung yang bisa dirasakan member dari perubahan apa pun di modul ini.

## Schedule (Kalender Agenda Internal)

Schedule adalah kalender agenda backoffice tempat admin/staff mencatat agenda internal (rapat, deadline, milestone, kegiatan) dengan tiga mode sharing (`private`, `all_admins`, `specific`-invite). Dokumen sumber menandai eksplisit: **"Surface: Backoffice only (tidak ada di mobile member pada Phase 1)"**, dan penggunanya adalah "semua admin/staff backoffice" (dikontrol via role/permission), dengan Superadmin punya hak override melihat semua agenda apa pun mode-nya.

Penting untuk membedakan modul ini dari dua hal lain yang terdengar mirip tapi berbeda:
- **Event** — modul konten acara publik yang memang dilihat & diikuti member di mobile (lihat [07_EVENT_JOURNEY.md](07_EVENT_JOURNEY.md)). Schedule bukan pengganti Event; Schedule adalah agenda internal admin, bukan listing acara publik.
- **"Add to Schedule" milik member di mobile** — fitur berbeda di sisi member untuk menyimpan jadwal pribadi terhadap sebuah event publik. Ini adalah fitur mobile-side yang berdiri sendiri, bukan bagian dari modul Schedule backoffice ini. Dokumen sumber menyebut eksplisit keduanya "berbeda".

Karena Schedule murni untuk agenda internal admin (rapat tim, follow-up, milestone proyek) dan belum punya endpoint mobile read-only apa pun di Phase 1, tidak ada dampak — langsung maupun tidak langsung — ke pengalaman member sama sekali. Rencana masa depan (Phase 3) menyebut kemungkinan "jadwal publik ke mobile", tapi itu belum ada dan belum mengubah status "tidak ada journey member" saat ini.

## Ringkasan

| Domain | Siapa Pakai | Dampak ke Member (langsung/tidak langsung/tidak ada) |
|---|---|---|
| Role & Permission System | Usergod (define permission master), Superadmin (assign ke system role) | Tidak langsung — menentukan aksi apa yang di-allow/deny saat member memakai fitur, tanpa UI yang pernah member lihat |
| System Settings | Usergod (teknis: security/email/storage), Superadmin (operasional: general/payment/moderation/maintenance/contact/legal) | Tidak langsung tapi sangat terasa — mis. Maintenance Mode meng-gate seluruh app member (503) meski member tidak pernah buka menunya |
| User Management (Backoffice) | Superadmin (platform-wide), Admin Region (region sendiri) | Tidak langsung — perubahan role/status/subscription dari sisi admin langsung mengubah apa yang member alami di app-nya sendiri (suspend, ganti plan, dll) |
| Accounting | Superadmin (KAI Pusat) saja | Tidak ada — murni ledger internal, tidak terhubung ke aksi member mana pun |
| Schedule (kalender internal) | Semua admin/staff backoffice, Superadmin (override lihat semua) | Tidak ada (Phase 1) — agenda internal admin, terpisah dari Event publik maupun "Add to Schedule" member |
