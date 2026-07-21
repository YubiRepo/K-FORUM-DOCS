# Verification Badge — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md).

## Ringkasan Domain

Verification Badge adalah centang keaslian ("Verified") yang di-grant selektif oleh **Superadmin** kepada tiga jenis entitas: **User** (akun member), **Merchant** (listing bisnis di Directory), dan **Community** (komunitas). Ini adalah layer trust yang terpisah tegas dari hal lain yang kelihatannya mirip — bukan konfirmasi kontak (`email_verified`/`phone_verified`), bukan gate publish (merchant `approval_status`), dan bukan ID Card (identitas universal semua member). Prosesnya manual di kedua sisi: member/owner/leader mengajukan dengan dokumen pendukung, Superadmin me-review satu per satu, dan setiap pengajuan/approve/reject/revoke tercatat permanen (append-only) untuk audit. Ketiga tipe badge dibedakan hanya dari field `type` — satu mesin (tabel `verifications`), tiga "produk" dengan label, syarat dokumen, dan tempat tampil yang berbeda.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Ajukan verifikasi **User** (badge akun sendiri) | ✅ | ✅ | Syarat cuma akun `active`, tidak digerbang plan |
| Lihat status pengajuan sendiri & resubmit setelah reject/revoke | ✅ | ✅ | |
| Ajukan verifikasi **Merchant** | ❌ | ✅ | Tergerbang **tidak langsung** oleh Pro: mengajukan mensyaratkan jadi owner merchant yang `approved` & `published`, dan membuat merchant sendiri adalah benefit Pro-only |
| Ajukan verifikasi **Community** | ❌ | ✅ | Sama polanya: mengajukan mensyaratkan jadi Leader komunitas yang `active`, dan membuat/memimpin komunitas adalah benefit Pro-only |
| Approve/reject/revoke badge (tipe apa pun) | ❌ | ❌ | Mutlak wewenang Superadmin, tidak ada di sisi member manapun |

## Journey 1: Member Mengajukan User Verification Badge — 🅢🅟 — 📱 Mobile

1. Trigger: Member (Standard atau Pro, akun `status = active`) membuka Profile > Ajukan Verifikasi.
2. Member melampirkan **minimal satu** dokumen pendukung dari daftar yang diterima (model `any_of`) — misalnya KTA KAI, KTA organisasi lain, kartu identitas (KTP/Paspor/SIM — opsional, **tidak wajib**), bukti jabatan/keanggotaan, atau link sosial media terverifikasi. Selfie/foto pemegang bersifat opsional, hanya diminta Superadmin kalau perlu klarifikasi tambahan.
3. Member mengisi alasan pengajuan, lalu [Submit] → sistem membuat request berstatus `pending`, dan Superadmin menerima notifikasi in-app "Ada pengajuan verifikasi baru".
4. Member melihat status pengajuannya di layar yang sama: "⏳ Pending review". Selama pending, member **tidak bisa** membuat pengajuan baru untuk entitas yang sama.
5. **Jalur approve** — Superadmin menyetujui → status jadi `approved`, badge "Verified Member" langsung tampil di profil (dan di semua tempat nama member muncul: author byline, komentar/post, dsb), member menerima notifikasi "Selamat, akun lu udah terverifikasi".
6. **Jalur reject** — Superadmin menolak dengan alasan wajib → member menerima notifikasi berisi alasan penolakan, badge tidak tampil, dan member boleh mengajukan ulang (resubmit) kapan saja — pengajuan lama tetap tersimpan sebagai riwayat, tidak dihapus.
7. **Revoke di kemudian hari** — Superadmin bisa mencabut badge kapan saja setelah approved (misal ditemukan pelanggaran/impersonation), dengan alasan wajib. Member menerima notifikasi "Verifikasi dicabut: {alasan}", badge langsung hilang dari semua tempat tampil.

**Selesai:** Member Standard maupun Pro punya jalur yang identik untuk mendapatkan badge "Verified Member" — tidak digerbang plan.

## Journey 2: Pro Member (Merchant Owner) Mengajukan Merchant Badge — 🅟 — 📱 Mobile

1. Prasyarat: Member adalah owner merchant (aksi buat merchant sendiri hanya benefit Pro), dan merchant tersebut sudah `approval_status = approved` **dan** `status = published` — merchant yang masih draft/pending/rejected/banned tidak bisa diajukan.
2. Owner membuka Merchant > Ajukan Verifikasi, melampirkan dokumen **wajib semua** (model `all_of`, lebih ketat dari User) — dokumen legalitas usaha (NIB/akta perusahaan/izin usaha) dan foto identitas pemilik. Bukti alamat usaha & akun sosmed bisnis bersifat opsional.
3. Submit → status `pending`, Superadmin menerima notifikasi pengajuan baru.
4. Superadmin review → **Approve**: badge "Verified Merchant" tampil di listing (card) dan halaman detail merchant, owner menerima notifikasi. **Reject**: owner menerima alasan, boleh resubmit dengan dokumen yang diperbaiki.
5. Badge bisa dicabut kapan saja oleh Superadmin (misal bisnis tutup, dokumen ternyata palsu), dengan alasan wajib — badge hilang dari listing & detail, owner menerima notifikasi.

**Selesai:** Merchant terverifikasi tampil dengan centang bisnis di seluruh listing Directory, memberi trust signal ke member lain.

## Journey 3: Pro Member (Community Leader) Mengajukan Community Badge — 🅟 — 📱 Mobile

1. Prasyarat: Member adalah Leader/owner komunitas (membuat/memimpin komunitas adalah benefit Pro-only), komunitas berstatus `active` (bukan archived/suspended). Hanya Leader yang punya izin ini — member biasa di komunitas tersebut tidak bisa mengajukan.
2. Leader membuka Community Settings > Ajukan Verifikasi, melampirkan **minimal satu** bukti legitimasi organisasi (model `any_of`) — akta/SK organisasi, bukti afiliasi/partner KAI, akun sosmed resmi komunitas, atau surat keterangan pengurus.
3. Submit atas nama komunitas → status `pending`, Superadmin menerima notifikasi.
4. Superadmin review legitimasi komunitas → **Approve**: badge "Verified Community" tampil di header komunitas serta di list komunitas, Leader menerima notifikasi. **Reject**: Leader menerima alasan, boleh resubmit.
5. Superadmin bisa mencabut badge kapan saja (misal komunitas bubar atau melanggar), dengan alasan wajib — badge hilang, Leader menerima notifikasi pencabutan.

**Selesai:** Komunitas terverifikasi tampil dengan centang di header dan listing, menandakan komunitas resmi/terafiliasi KAI.

## Keterlibatan Admin — 💻 Web/Backoffice

Untuk **ketiga tipe badge** (User, Merchant, Community), seluruh kurasi ada **sepenuhnya** di tangan Superadmin:

1. Superadmin membuka Backoffice > Verification Queue, memfilter pengajuan `pending` (bisa filter per `type`: user/merchant/community).
2. Superadmin membuka detail pengajuan — melihat dokumen yang dilampirkan (akses dokumen sensitif ini **eksklusif** Superadmin; disimpan di storage private, tidak pernah lewat CDN publik/response member-facing).
3. **Approve** → status `approved`, kolom cache `is_verified` di-set `true` pada entitas terkait (users/merchants/communities), badge langsung tampil, notifikasi terkirim ke pemohon.
4. **Reject** → status `rejected`, alasan wajib diisi, notifikasi berisi alasan terkirim ke pemohon, pemohon boleh mengajukan ulang.
5. **Revoke** (dari status `approved` kapan saja) → status `revoked`, alasan wajib diisi, `is_verified` di-set `false`, badge hilang, notifikasi terkirim ke pemilik/Leader.
6. Semua tindakan (submit/approve/reject/revoke) tercatat append-only di tabel `verifications` — tidak ada hard-delete, jadi seluruh riwayat pengajuan (termasuk yang lama sudah reject/revoke) tetap bisa diaudit.

**Admin Regional secara eksplisit TIDAK punya peran approval/reject/revoke di domain ini** — satu-satunya kewenangan Admin Regional adalah **melihat (read-only)** status verified entitas di region miliknya untuk keperluan monitoring. Ini berbeda dari domain lain (misal Merchant approval biasa) di mana Admin Regional biasa terlibat — Verification Badge sengaja disentralisasi penuh ke Superadmin karena sifatnya global trust signal, bukan moderasi konten regional.

## Di Luar Cakupan Standard & Pro

- **Member tidak bisa self-approve/reject/revoke badge miliknya sendiri** — kurasi mutlak wewenang Superadmin, memastikan badge tetap bernilai sebagai sinyal trust yang selektif.
- **Admin Regional sama sekali tidak bisa memproses (approve/reject/revoke) pengajuan verifikasi apa pun** — beda dari domain lain, verifikasi tidak didelegasikan ke region, hanya read-only monitoring.
- **Tidak ada approval otomatis/instan** — Phase 1 murni manual-first; hook untuk auto-KYC/integrasi pihak ketiga sudah disiapkan tapi belum dibangun.
- **Member tidak bisa mengajukan lebih dari satu request aktif (`pending`) untuk entitas yang sama** — sistem menolak pengajuan kedua selama yang pertama masih menunggu review, untuk mencegah spam.
- **Member biasa (bukan Leader) tidak bisa mengajukan verifikasi komunitas** — hanya Leader/owner komunitas yang punya permission `verification.request_community`.
- **Tidak ada banding (appeal) terstruktur** atas penolakan — member hanya bisa resubmit dengan dokumen baru, bukan mengajukan keberatan formal atas keputusan Superadmin.
- **Tidak ada tier badge** (misal "Official Organization" vs "Verified Individual") — setiap tipe (user/merchant/community) hanya punya satu level badge: terverifikasi atau tidak.

## Edge Case & Catatan Tambahan

- **Downgrade Pro→Standard tidak otomatis mencabut badge Merchant/Community yang sudah dimiliki** — rules tidak menyebut revoke otomatis saat downgrade; revoke tetap harus aksi manual Superadmin dengan alasan (mis. kalau merchant/komunitas jadi tidak aktif akibat downgrade, itu dievaluasi kasus per kasus, bukan trigger otomatis dari perubahan plan). Ini konsisten dengan prinsip "resolve live dari record verifikasi terbaru", bukan snapshot yang terikat status plan.
- **Badge tidak digerbang oleh subscription secara langsung** — hook untuk menjadikan badge sebagai benefit berbayar/di-gate subscription sudah disiapkan tapi sengaja tidak dibangun di Phase 1; Pro hanya jadi prasyarat *tidak langsung* karena Pro dibutuhkan untuk *memiliki* merchant/komunitas yang bisa diajukan verifikasinya.
- **Rejected/revoked boleh resubmit sebagai record baru** — bukan update record lama; riwayat penolakan/pencabutan sebelumnya tetap tersimpan permanen untuk audit.
- Dokumen yang diklaim di rules soal "verifikasi member scoped per-komunitas" (Leader verifikasi member di dalam komunitasnya sendiri, badge cuma berlaku lokal) secara eksplisit **ditunda** dan bukan bagian dari Phase 1 — badge yang ada sekarang selalu global (dikelola Superadmin), bukan per-komunitas.
