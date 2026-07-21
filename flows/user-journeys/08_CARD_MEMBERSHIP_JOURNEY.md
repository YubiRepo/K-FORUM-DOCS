# Card Membership (ID Card Digital) — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md).

## Ringkasan Domain

ID Card adalah bukti identitas keanggotaan KAI yang **stabil** — bukan representasi status subscription. Kartu sengaja dipisah dari plan: tidak menyimpan `plan`, benefit, atau `expiry_date`, dan tampilannya sama sekali tidak berubah antara Standard dan Pro. Setiap member — Standard maupun Pro — otomatis mendapat satu kartu saat registrasi selesai; tidak ada aksi "buat kartu" terpisah yang perlu dilakukan member. Fungsinya tiga lapis: identitas ("orang ini member KAI"), verification token (QR yang di-scan partner/event/admin), dan gateway ke benefit yang di-resolve **live** oleh modul lain (Subscription) saat kartu di-scan — bukan disimpan di kartu itu sendiri. Karena itu, upgrade/downgrade Pro↔Standard tidak menyentuh kartu sama sekali.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Punya ID Card (auto-provisioned saat registrasi) | ✅ | ✅ | Identik untuk kedua tier |
| Lihat kartu digital di Profile | ✅ | ✅ | |
| Download kartu sebagai PDF | ✅ | ✅ | |
| Share QR link untuk verifikasi online | ✅ | ✅ | |
| Request kartu fisik (print + kirim) | ✅ | ✅ | Biaya cetak fisik mungkin jadi benefit Pro (`physical_id_card`) atau shipping-only — keputusan bisnis terpisah, **tidak mempengaruhi validitas kartu** |
| Request replacement (hilang/rusak) | ✅ | ✅ | Maks 2×/tahun untuk kedua tier |
| Tampilan/desain kartu berubah sesuai plan | ❌ | ❌ | Desain kartu sama untuk semua tier — kartu tidak punya "plan badge" |
| Kartu ikut nonaktif saat downgrade/telat bayar | ❌ | ❌ | Kartu hanya `revoked` untuk alasan identitas (banned/akun dihapus/keluar), bukan pembayaran |

## Journey 1: Member Melihat & Menggunakan ID Card — 🅢🅟 — 📱 Mobile

1. **Provisioning otomatis** — begitu member menyelesaikan registrasi, sistem langsung membuat kartu di background: generate `card_id` unik (format `KAI-{YYYY}-{Random}`), set `region_id` dari region user saat itu, `status = active`, dan build data QR (`qr_version`, `qr_code_data`) serta tampilan (`digital_format`: nama, avatar, qr_url). Member tidak melakukan apa pun untuk memicu ini — kartu sudah "ada" begitu akun aktif.
2. **Lihat kartu** — Member membuka Profile > Membership ID Card, melihat kartu digital berisi nama, avatar, dan QR code. Tidak ada indikator plan/benefit di tampilan ini.
3. **Download PDF** — Member menekan [Download], mendapat kartu dalam bentuk PDF self-contained (bisa dipakai offline, tanpa perlu buka app).
4. **Share QR link** — Member menekan [Share], mendapat link publik (`kai.app/verify/{card_id}`) yang bisa dikirim ke siapa saja. Siapa pun yang membuka link hanya melihat info publik: nama, region, status valid — **tanpa** plan/benefit/data sensitif lain.
5. **Tunjukkan QR untuk di-scan** — di lokasi partner (toko direktori) atau gate event, member menunjukkan QR dari app untuk di-scan staff/gate.
   - Backend memvalidasi checksum QR dan `status = active`.
   - Jika staff/partner ter-autentikasi, sistem me-resolve plan & benefit member **secara live** dari modul Subscription saat itu juga (misal diskon Pro) — bukan dari data yang tersimpan di kartu.
   - Setiap scan (berhasil maupun gagal) tercatat sebagai `card_scan_events`, jadi dari sisi member setiap penggunaan kartu punya jejak yang bisa diaudit.
6. **Request kartu fisik** — Member mengisi alamat pengiriman & nomor telepon, submit → sistem membuat fulfillment order (`physical_ordered = true`), status "processing" dengan estimasi tanggal kirim.
7. **Request replacement** (jika hilang/rusak) — Member submit alasan (`lost`/`damaged`/`other`) + alamat kirim → sistem menerbitkan `card_id` baru (kartu lama otomatis tidak berlaku), dibatasi maksimal 2× per tahun untuk cegah abuse fulfillment.

**Selesai:** Member memiliki kartu identitas yang bisa dilihat, diunduh, dibagikan, dan di-scan kapan saja — tanpa perlu memperbarui apa pun saat plan mereka berubah.

## Keterlibatan Admin — 💻 Web/Backoffice

Keterlibatan admin di domain ini **minim** dan sebagian besar bersifat read/monitoring, bukan approval alur member:

- **Superadmin & Admin Regional** — bisa scan & verifikasi kartu (peran yang sama seperti partner/staff saat memvalidasi identitas member secara langsung/tatap muka), dan bisa melihat/filter daftar kartu (Admin Regional terbatas pada region miliknya, Superadmin semua region).
- **Revoke kartu** — Admin Regional (region sendiri) atau Superadmin bisa me-revoke kartu, tapi **hanya** untuk alasan keanggotaan: member di-ban, akun dihapus, atau keluar dari KAI. Ini **bukan** aksi rutin dalam journey member normal — tidak ada proses "approve kartu" seperti pada domain lain (news, upgrade, dsb) karena kartu memang auto-provisioned tanpa gate admin.
- **Lihat scan analytics** — Admin Regional (scope region), Superadmin (scope global) bisa melihat riwayat scan untuk analitik/fraud detection.
- Tidak ada langkah admin yang menghambat member menerima, melihat, atau menggunakan kartunya — beda dari domain seperti Subscription Upgrade atau News yang punya antrian approval eksplisit.

## Di Luar Cakupan Standard & Pro

- **Member tidak bisa self-revoke atau reissue kartu sendiri di luar flow replacement resmi** — revoke murni wewenang admin (alasan identitas/keanggotaan), dan reissue hanya lewat flow "request replacement" resmi (dibatasi 2×/tahun).
- **Member tidak bisa mengubah desain/tampilan kartu** — `digital_format` (design, layout) ditentukan sistem, bukan dikustomisasi member; kustomisasi desain baru direncanakan sebagai fitur masa depan (Phase 3 roadmap), bukan hak member saat ini.
- **Member tidak bisa scan kartu member lain** — aksi scan/verifikasi adalah hak Partner/Staff dan Admin, bukan sesama member.
- **Member tidak bisa melihat riwayat scan kartunya sendiri secara mandiri** — `card_scan_events` dipakai untuk analitik sisi partner/admin; permission matrix di rules tidak memberi member akses langsung ke log scan mereka sendiri.

## Edge Case & Catatan Tambahan

- **Downgrade/upgrade plan tidak memicu perubahan kartu sama sekali** — next scan otomatis me-resolve plan/benefit terbaru dari Subscription; tidak perlu cetak ulang atau update kartu fisik.
- **Tidak ada status "expired"** pada kartu — identitas tidak kedaluwarsa; hanya `active` atau `revoked` secara permanen.
- **QR bersifat versioned** (`v1|card_id|checksum`) sehingga format bisa berkembang ke depan tanpa mematikan kartu fisik lama yang sudah dicetak.
- **Verifikasi offline vs online**: quick check dasar (checksum) bisa dilakukan offline, tapi benefit bernilai tinggi (diskon, entry event) wajib panggilan backend karena status kartu bisa berubah (mis. di-revoke) sejak dicetak.
