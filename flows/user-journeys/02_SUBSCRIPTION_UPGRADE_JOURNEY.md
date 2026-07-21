# Subscription & Upgrade ke Pro — User Journey

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Plan Subscriptoin/PLAN_SUBSCRIPTION_SYSTEM.md`.

## Ringkasan Domain

Setiap member baru otomatis berada di plan **Standard** (gratis diakses langsung, tanpa perlu bayar di awal — biaya Rp 49rb/bulan berjalan sebagai biaya keanggotaan berkelanjutan, bukan syarat aktivasi akun). Member bisa upgrade ke **Pro** kapan saja lewat verifikasi transfer manual (payment gateway masih rencana ke depan). Domain ini adalah "pintu gerbang" ke semua domain lain yang men-gate fitur berdasarkan plan.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Lihat status subscription & benefit sendiri | ✅ | ✅ | |
| Request upgrade ke Pro | ✅ | — | |
| Downgrade ke Standard | — | ✅ | Efektif di akhir periode aktif |
| Cancel/renewal reminder | ✅ | ✅ | |
| Approve/reject upgrade orang lain | ❌ | ❌ | Hanya Superadmin |
| Ubah harga plan / benefit master | ❌ | ❌ | Hanya Superadmin (benefit assignment) & Usergod (benefit key) |

## Journey 1: Signup — Otomatis Jadi Standard — 🅢 — 📱 Mobile

1. User mengisi form registrasi (email, password, nama, dll) → submit.
2. Verifikasi email (OTP/link).
3. Melengkapi profil (opsional, bisa dilewati).
4. Diarahkan ke home.
5. Di background, sistem otomatis:
   - Membuat entri `subscription`: `plan = standard`, `status = active`.
   - Mencatat `subscription_history`: action = `signup`.
6. Member menerima email selamat datang: "Welcome to KAI! You're on Standard plan."
7. Member langsung bisa memakai semua fitur Standard tanpa proses pembayaran tambahan.

**Selesai:** Member berstatus Standard aktif dan siap memakai fitur inti (baca konten, join komunitas, dll).

## Journey 2: Member Request Upgrade ke Pro — 🅢→🅟 — 📱 Mobile

1. Trigger: Member menekan "Upgrade to Pro" di halaman profil/pricing.
2. **Konfirmasi upgrade** — layar menampilkan perbandingan plan (Standard Rp49rb vs Pro Rp129rb, selisih Rp80rb) beserta daftar benefit yang akan terbuka (post news, create community, create store, create event, view analytics, priority support). [Continue] / [Cancel].
3. **Pilih metode pembayaran** — saat ini hanya "Manual transfer" yang aktif (Stripe/Midtrans masih "coming soon").
4. **Transfer manual** — layar menampilkan rekening tujuan (BCA/Mandiri/GCash), nominal, dan kode referensi unik. Member mencentang "saya sudah transfer" dan bisa upload bukti transfer (opsional tapi direkomendasikan). [Submit Request].
5. Sistem membuat `subscription_requests` berstatus `pending`.
6. Member melihat status di Profile → Subscription: "⏳ Awaiting admin verification — biasanya diverifikasi dalam 1×24 jam", dengan opsi [View proof] dan [Cancel request].
7. Selama status pending, member **tidak bisa** mengajukan request upgrade baru (sistem menolak dengan pesan "You already have a pending upgrade request").

**Selesai (jalur sukses):** lanjut ke keterlibatan admin di bawah → setelah disetujui, member menerima notifikasi email + in-app "Upgrade successful!" dan semua benefit Pro langsung aktif tanpa perlu logout/login ulang.

**Selesai (jalur ditolak):** member melihat status "❌ Rejected" beserta alasan, dan tombol [Retry upgrade].

## Journey 3: Pro Member Downgrade / Cancel — 🅟→🅢 — 📱 Mobile

1. Member membuka Subscription Settings, menekan "Downgrade to Standard".
2. Modal konfirmasi menjelaskan benefit yang akan hilang (post news, create community, create store, create event) dan bahwa downgrade efektif di tanggal expiry periode berjalan (fitur Pro tetap jalan sampai tanggal itu).
3. Member menekan [Confirm downgrade].
4. Sistem menandai subscription `cancelled` dan mencatat `subscription_history` (Pro → Standard, action = downgrade).
5. Di tanggal expiry: auto-downgrade ke Standard, member menerima notifikasi "Your Pro plan expired. You're now on Standard."

## Journey 4: Renewal Reminder — 🅟 — 📱 Mobile

1. H-7 sebelum expiry: email reminder ("Your Pro plan expires in 7 days", nominal renewal, opsi lanjut/cancel).
2. H-3 sebelum expiry: notifikasi in-app "Your Pro plan expires in 3 days!".
3. Jika tidak ada aksi → auto-renew di tanggal expiry (tetap menunggu verifikasi admin untuk metode manual).
4. Jika member menekan [Cancel renewal] → Pro berakhir tepat di tanggal expiry, otomatis turun ke Standard.

## Keterlibatan Admin — 💻 Web/Backoffice

**Superadmin — Verifikasi upgrade request:**
1. Membuka "Subscription Verification Dashboard", filter [Pending].
2. Melihat detail request: user, plan saat ini → yang diminta, nominal, bukti transfer.
3. [Approve] → sistem update request jadi `completed`, buat subscription Pro baru, arsipkan subscription Standard lama, catat history, kirim email + notifikasi ke member.
4. [Reject] → pilih alasan (duplicate/invalid proof/amount mismatch/lainnya), sistem update request jadi `rejected` + kirim alasan ke member.

**Superadmin — Kelola plan & benefit:**
- Edit harga/deskripsi plan, enable/disable plan, tambah plan baru (future).
- Assign/toggle benefit per plan (`plan_benefits`) — perubahan berlaku live tanpa deploy ulang.
- Lihat analytics subscription (total user, conversion rate, MRR, churn).
- Manual refund/downgrade member (misal karena kesalahan transaksi).

**Admin Regional — Monitoring only:**
- Melihat subscription request & history di region miliknya saja.
- **Tidak bisa approve langsung** — hanya bisa forward request ke Superadmin.

**Usergod (jarang tampil ke member):**
- Mendefinisikan benefit key baru di `benefits_master` (misal `advanced_analytics`) saat ada fitur baru — ini murni pekerjaan developer/konfigurasi sistem, bukan bagian journey member.

## Di Luar Cakupan Standard & Pro

- **Member tidak bisa approve upgrade sendiri atau member lain** — approval mutlak di tangan Superadmin.
- **Member tidak bisa mengatur harga plan atau benefit** yang tersedia di tiap plan — ini kewenangan Superadmin (assignment) dan Usergod (definisi benefit baru).
- **Member tidak bisa mengonfigurasi payment gateway** (Stripe/Midtrans) — pengaturan ini murni backoffice Superadmin.
- **Admin Regional tidak bisa menyetujui/menolak upgrade** — hanya monitoring & forward ke pusat.

## Edge Case & Catatan Tambahan

- Downgrade saat masih ada renewal pending: request upgrade baru tetap bisa dibuat; jika disetujui, renewal Standard lama otomatis dibatalkan demi transisi bersih.
- Grace period expiry (opsional, dikonfigurasi admin): Pro yang expired tetap dapat akses beberapa hari (default pola: 7 hari) sebelum benar-benar turun ke Standard, dengan warning in-app setiap hari selama masa itu.
