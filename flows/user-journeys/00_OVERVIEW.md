# User Journey — Member Standard & Pro (per Domain)

Kumpulan dokumen user journey untuk member **Standard** dan **Pro** di setiap domain KAI App, dari awal masuk aplikasi sampai selesai menyelesaikan satu fitur tertentu. Fokus dokumen ini adalah **pengalaman pengguna** (langkah, layar, notifikasi, error state) — bukan spesifikasi API atau skema database. Untuk itu lihat `Modules/` dan `API SPEC/`.

---

## Legend

### Platform

| Simbol | Platform | Keterangan |
|--------|----------|------------|
| 📱 | **Mobile** | Aplikasi member (Flutter app — `k_forum`). Semua journey member terjadi di sini. |
| 💻 | **Web/Backoffice** | Panel admin (Nuxt — `k-forum-backoffice`). **Tidak ada** aplikasi web untuk member — Web selalu berarti sisi admin/superadmin/admin regional. |

> Catatan penting: di KAI App, member (Standard maupun Pro) **hanya punya akses via Mobile**. Kalau sebuah dokumen journey menyebut "Web", itu selalu bagian **keterlibatan admin**, bukan alternatif akses member.

### Tier Member

| Simbol | Tier | Keterangan |
|--------|------|------------|
| 🅢 | **Standard** | Plan default semua member baru (Rp 49rb/bulan, otomatis saat signup, tanpa perlu bayar di awal). |
| 🅟 | **Pro** | Upgrade opsional (Rp 129rb/bulan, verifikasi manual transfer saat ini). |

### Role Admin

| Role | Scope | Keterangan |
|------|-------|------------|
| **Usergod** | Platform (developer/vendor) | Akses penuh, konfigurasi sistem inti. Tidak pernah muncul di journey member. |
| **Superadmin** | KAI Pusat, global | Approve upgrade, approve konten Pro (news, verification badge, dll), kelola plan & benefit. |
| **Admin Regional** | Per wilayah (KAI Wilayah) | Kelola & moderasi member/konten di region miliknya saja. |

---

## Ringkasan Tier: Apa Bedanya Standard vs Pro?

Sumber: [`Modules/Plan Subscriptoin/PLAN_SUBSCRIPTION_SYSTEM.md`](../../Modules/Plan%20Subscriptoin/PLAN_SUBSCRIPTION_SYSTEM.md)

| Fitur | Standard | Pro |
|---|:---:|:---:|
| Baca news, event, direktori | ✅ | ✅ |
| Join komunitas | ✅ | ✅ |
| Post & komentar di komunitas | ✅ | ✅ |
| Tanya/jawab di Q&A | ✅* | ✅* |
| Lihat & cari direktori | ✅ | ✅ |
| Edit profile, ID Card, notification preferences | ✅ | ✅ |
| **Post news** (dengan approval admin) | ❌ | ✅ |
| **Buat & kelola community** (jadi leader) | ❌ | ✅ |
| **Buat company/merchant** di direktori | ❌ | ✅ |
| **Buat & host event** | ❌ | ✅ |
| **Lihat analytics** (community/store/event) | ❌ | ✅ |
| **Pasang Ads** (`post_ads` benefit) | ❌ | ✅ |
| Priority support | ❌ | ✅ |

`*` Ask/answer Q&A digerbang oleh benefit `ask_qna` / `answer_qna` yang **bisa dikonfigurasi ulang oleh Superadmin** per plan — secara default kedua benefit ini bisa saja diberikan ke Standard juga. Lihat [05_QNA_JOURNEY.md](05_QNA_JOURNEY.md) untuk detail.

**Prinsip penting (jangan ketuker):**
- **Benefit** (subscription-driven) = fitur apa yang bisa diakses. Dikelola superadmin lewat Plan Benefits, bisa toggle kapan saja tanpa deploy ulang.
- **Permission** (role-driven) = siapa yang boleh approve/moderate aksi tersebut. Dikelola lewat Role & Permission system.
- Kedua sistem ini **independen** — plan menentukan "boleh coba posting", role menentukan "siapa yang approve postingan itu".

Detail lengkap upgrade/downgrade/renewal ada di [02_SUBSCRIPTION_UPGRADE_JOURNEY.md](02_SUBSCRIPTION_UPGRADE_JOURNEY.md).

---

## Daftar Journey per Domain

| # | Domain | File | Ada keterlibatan admin? |
|---|--------|------|:---:|
| 01 | Onboarding & Account (register, verifikasi, region, subscription default) | [01_ONBOARDING_ACCOUNT_JOURNEY.md](01_ONBOARDING_ACCOUNT_JOURNEY.md) | Minim |
| 02 | Subscription & Upgrade ke Pro | [02_SUBSCRIPTION_UPGRADE_JOURNEY.md](02_SUBSCRIPTION_UPGRADE_JOURNEY.md) | Ya |
| 03 | News | [03_NEWS_JOURNEY.md](03_NEWS_JOURNEY.md) | Ya |
| 04 | Community (join, post, create, announcement, schedule, invite) | [04_COMMUNITY_JOURNEY.md](04_COMMUNITY_JOURNEY.md) | Ya |
| 05 | Q&A / FAQ | [05_QNA_JOURNEY.md](05_QNA_JOURNEY.md) | Ya |
| 06 | Directory (merchant/company) | [06_DIRECTORY_JOURNEY.md](06_DIRECTORY_JOURNEY.md) | Ya |
| 07 | Event | [07_EVENT_JOURNEY.md](07_EVENT_JOURNEY.md) | Ya |
| 08 | Card Membership (ID Card digital) | [08_CARD_MEMBERSHIP_JOURNEY.md](08_CARD_MEMBERSHIP_JOURNEY.md) | Minim |
| 09 | Verification Badge | [09_VERIFICATION_BADGE_JOURNEY.md](09_VERIFICATION_BADGE_JOURNEY.md) | Ya (Superadmin-only) |
| 10 | Ads | [10_ADS_JOURNEY.md](10_ADS_JOURNEY.md) | Ya |
| 11 | Announcement (broadcast admin) | [11_ANNOUNCEMENT_JOURNEY.md](11_ANNOUNCEMENT_JOURNEY.md) | Ya (admin sebagai creator) |
| 12 | Reporting (content & bug report) | [12_REPORTING_JOURNEY.md](12_REPORTING_JOURNEY.md) | Ya |
| 13 | Notifications (inbox & preferences) | [13_NOTIFICATIONS_JOURNEY.md](13_NOTIFICATIONS_JOURNEY.md) | Tidak |
| 14 | User Settings (profile & keamanan akun) | [14_USER_SETTINGS_JOURNEY.md](14_USER_SETTINGS_JOURNEY.md) | Tidak |
| 15 | Domain admin-only (tidak ada journey member) | [15_ADMIN_ONLY_DOMAINS_NOTE.md](15_ADMIN_ONLY_DOMAINS_NOTE.md) | — |

---

## Cara Membaca Tiap Dokumen

Setiap file journey mengikuti struktur yang sama:

1. **Ringkasan domain** — apa domain ini & siapa aktornya.
2. **Batasan Standard vs Pro** — tabel aksi yang boleh/tidak per tier di domain ini.
3. **Journey Standard** — langkah demi langkah 📱 Mobile dari entry point sampai fitur selesai dipakai.
4. **Journey Pro** — tambahan langkah yang hanya berlaku untuk Pro (kalau ada perbedaan).
5. **Keterlibatan Admin** — langkah 💻 Web/Backoffice kalau ada approval/moderasi.
6. **Di luar cakupan Standard & Pro** — fitur yang memang bukan hak member sama sekali (baik Standard maupun Pro), ditulis eksplisit beserta alasannya.

Tujuannya: memastikan tidak ada gap pengalaman — dari member baru buka aplikasi pertama kali, sampai berhasil menuntaskan aksi spesifik di tiap domain.

---

## Lihat Juga: Journey Superadmin (Backoffice)

Dokumen di atas fokus ke pengalaman **member**, dengan bagian admin hanya ditulis sebagai "Keterlibatan Admin" secukupnya. Untuk journey lengkap dari sisi **Superadmin** mengelola satu domain di backoffice (menu apa saja, tombol apa saja, alur kerja sehari-hari), lihat folder terpisah [`flows/superadmin/`](../superadmin/):

| Domain | File |
|---|---|
| News | [NEWS_JOURNEY.md](../superadmin/NEWS_JOURNEY.md) |

Folder ini akan bertambah seiring domain lain didokumentasikan dari sisi Superadmin.
