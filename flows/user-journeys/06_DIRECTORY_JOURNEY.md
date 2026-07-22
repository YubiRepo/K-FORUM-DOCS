# Directory (Merchant/Company) — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Directory/DIRECTORY_RULES.md`, `Modules/Directory/MERCHANT_LIKE_SPEC.md`.

## Ringkasan Domain

Directory adalah katalog bisnis komunitas Korea-Indonesia. Semua member (Standard maupun Pro) bisa menelusuri, mencari, menyimpan (save), menyukai (like), mereview, dan mengirim inquiry ke merchant — scope-nya **global**, semua user bisa lihat merchant dari seluruh region. Yang jadi pembeda tier: hanya **Member Pro** yang bisa mendaftarkan bisnisnya sendiri. Struktur data mengikuti hierarki `Member Pro → Company → Merchant (toko/outlet) → Item (product/service)`, ditambah `Reviews` dan `Inquiries` yang menempel di tiap merchant.

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Browse & search direktori (kategori, kota, rating, dll) | ✅ | ✅ | |
| Lihat detail merchant | ✅ | ✅ | |
| Save/unsave (bookmark) merchant | ✅ | ✅ | Independen dari Like |
| Like/unlike merchant | ✅ | ✅ | Menaikkan `favorite_count`, independen dari Save |
| Tulis & edit review (1 per merchant) | ✅ | ✅ | Tidak bisa delete review sendiri |
| Kirim inquiry (tanya jawab) ke merchant | ✅ | ✅ | |
| Buat Company | ❌ | ✅ | Unlimited company per Member Pro |
| Buat Merchant (toko/outlet) di bawah company | ❌ | ✅ | Subject to setting `merchant.max_per_member` |
| Edit/archive/delete merchant sendiri | ❌ | ✅ | Hanya merchant milik sendiri |
| Tambah/edit/hapus Item (product/service) | ❌ | ✅ | Hanya di merchant sendiri |
| Balas inquiry masuk (sebagai owner) | ❌ | ✅ | Hanya 1x reply per inquiry |
| Lihat stats/analytics merchant sendiri | ❌ | ✅* | *Lihat catatan di bagian Edge Case — ada kontradiksi di dokumen sumber |
| Review merchant sendiri | ❌ | ❌ | Owner dilarang review merchant miliknya sendiri, meski dia Pro |

## Journey 1: Member Menjelajah Direktori — 🅢🅟 — 📱 Mobile

1. Member membuka tab **Directory** dari home/menu utama.
2. **Browse listing default** — merchant tampil dengan sort `rating_desc` (default): featured merchant (jika diaktifkan) tampil pertama, lalu diurutkan rating tertinggi, tiebreaker review count terbanyak, lalu terbaru.
3. **Search & filter** — member bisa:
   - Cari via teks bebas (`q`, mencocokkan nama & deskripsi merchant).
   - Filter: kategori (`category_id`), tipe merchant (retail/online/service/food_beverage/beauty), kota, rating minimum, "sedang buka" (`is_open_now`), punya produk (`has_product`), punya jasa (`has_service`).
   - Ubah sort: rating tertinggi, terbaru, nama A-Z, paling banyak direview.
   - **Empty state:** filter tidak menemukan hasil → tampilkan pesan "Tidak ada merchant ditemukan" + opsi reset filter.
4. **Buka detail merchant** — member menekan salah satu kartu merchant, melihat: profil (nama, deskripsi, kategori, tipe), foto, jam operasional, kontak (phone/email/whatsapp/instagram), lokasi (peta jika `location.enable_map_view` aktif dan merchant punya lat/lng), daftar item (product + service), serta reviews.
   - `total_views` merchant bertambah setiap kali detail dilihat (unique view).
5. **Save/unsave (bookmark)** — member menekan ikon bookmark untuk menyimpan merchant ke daftar "Tersimpan" (`me/saved`), bisa unsave kapan saja.
6. **Like/unlike merchant** — terpisah dari Save. Member menekan ikon hati; sistem toggle `is_liked` dan `favorite_count` secara **optimistic** (langsung berubah di UI, baru memanggil endpoint) — kalau gagal, rollback ke state sebelumnya. Aksi ini **idempotent**: like yang ditekan dua kali tidak menggandakan hitungan, tombol dikunci sementara selama request berjalan agar tidak double-tap.
   - **Guest/belum login:** menekan like diarahkan ke login (bukan error), mengikuti pola interaksi lain di app.
7. **Tulis review** — member mengisi rating (1-5, wajib), title & teks (opsional), serta rating per-aspek opsional (kualitas produk, layanan, harga). Ketentuan:
   - 1 user hanya bisa 1 review per merchant; kalau sudah pernah review, aksi berikutnya adalah **edit** review lama (bukan review baru).
   - Member **tidak bisa menghapus** review sendiri.
   - Jika setting `review.require_moderation` aktif → review masuk status `pending`, belum tampil publik sampai admin approve. Kalau moderasi mati → langsung `published`.
   - Rating merchant (avg) otomatis di-update setiap kali review dibuat/diubah.
8. **Kirim inquiry (tanya jawab)** — kalau merchant mengaktifkan `allow_inquiries`, member mengisi subject (min 5 char) & pesan (min 10 char) lalu kirim. Kalau merchant punya auto-reply, member langsung menerima pesan otomatis sebelum owner sempat membalas manual. Member bisa melihat semua inquiry yang pernah dia kirim ke berbagai merchant.
9. **Hubungi merchant langsung** — dari layar detail, member bisa langsung tap kontak yang tersedia (telepon/WhatsApp/Instagram) untuk keluar ke aplikasi terkait; ini terpisah dari alur Inquiry di dalam app.
10. **Error/edge state saat interaksi:**
    - Like/save/review/inquiry tanpa login → `401 ERR_UNAUTHORIZED`, diarahkan ke login.
    - Merchant sudah tidak ada/dihapus → `404 NOT_FOUND`, tampilkan halaman "Merchant tidak ditemukan".
    - Merchant berstatus `banned`/`archived`/`rejected` tidak akan muncul di listing publik sama sekali (hanya owner & admin yang bisa melihatnya).

**Selesai:** Member berhasil menemukan, mengeksplorasi, dan berinteraksi (save/like/review/inquiry) dengan merchant yang relevan — tanpa perlu jadi Pro.

## Journey 2: Pro Member Membuat & Mengelola Listing — 🅟 — 📱 Mobile

1. **Entry point** — dari tab Directory, Member Pro melihat CTA "Buat Bisnis Saya" / "My Business" (tidak tersedia untuk Standard).
2. **Buat Company** — isi nama (wajib, 3-200 char), deskripsi (opsional, max 2000 char), logo (upload dulu via media API, context `directory`), phone/email/website (opsional, format tervalidasi). Company langsung berstatus `active`. Satu Member Pro bisa punya **unlimited company**.
3. **Buat Merchant (toko/outlet) di bawah company** — pilih company pemilik, isi:
   - Nama (3-200 char), deskripsi (wajib, 20-3000 char), tipe (`retail`, `food_beverage`, `beauty`, `service`, `online`, `other`).
   - Kategori: pilih 1-5 dari category master (Superadmin yang kelola daftar kategorinya).
   - Foto: minimal 1 (kalau setting `min_images_before_publish` > 0), maksimal 10, wajib upload dulu ke CDN sebelum dipakai.
   - Lokasi: alamat + lat/lng **wajib** untuk tipe fisik (retail, food_beverage, beauty, service); **opsional** untuk `online`; untuk `other` tergantung setting `require_location_for_physical`. Kalau lat/lng diisi, `region_id`/kota/provinsi **otomatis di-derive** via geo-lookup — member tidak input manual.
   - Kontak: minimal salah satu dari phone/email/whatsapp/instagram.
   - Jam operasional per hari (opsional).
4. **Submit/Publish merchant** — sistem validasi field wajib dulu:
   - Gagal validasi → tetap status `draft`, tampilkan error per field.
   - Lolos validasi → cek setting `merchant.require_approval`:
     - **Aktif (default `true`)** → status jadi `pending_approval`, notifikasi terkirim ke Admin Regional (kalau ada di region tsb) atau Superadmin. Merchant belum tampil di listing publik, tapi owner tetap bisa melihatnya lengkap dengan info status "Menunggu approval".
     - **Nonaktif** → langsung `published` (auto-approve), `published_at` diisi saat itu juga, langsung tampil di listing.
5. **Cek status di "My Merchants"** — owner memantau status tiap merchant: `draft` / `pending_approval` / `published` / `rejected` (beserta alasan) / `archived` / `banned` (beserta alasan). Semua status ini tetap terlihat penuh oleh owner meski tidak tampil ke publik.
6. **Kalau disetujui** — owner menerima notifikasi in-app + FCM "Merchant disetujui", merchant langsung tampil di listing publik.
7. **Kalau ditolak** — owner menerima notifikasi berisi alasan penolakan, bisa mengedit merchant lalu resubmit untuk direview ulang.
8. **Tambah Item (product/service)** — di dalam merchant, owner menambah item tanpa batas (kecuali setting `item.max_per_merchant`), boleh campur product & service dalam 1 merchant:
   - **Product:** nama, harga (>0, kecuali `allow_free_price` aktif), currency, unit (opsional), stock (opsional — null berarti unlimited), foto min 1 max 10, status `available`/`unavailable`.
   - **Service:** nama, price_min & price_max (price_max ≥ price_min), currency, durasi menit (opsional), foto min 1 max 10, status `available`/`unavailable`.
   - Kategori item adalah **teks bebas**, bukan dari master list.
   - Item **tidak butuh approval terpisah**, kecuali setting `item.require_approval` diaktifkan Superadmin.
   - Item berstatus `unavailable` tetap tampil di listing tapi berlabel "Tidak Tersedia".
9. **Edit merchant/company** — owner hanya bisa edit company & merchant miliknya sendiri.
10. **Archive vs Delete:**
    - **Archive:** merchant hilang dari listing publik, tapi semua data (item, review, inquiry) tetap utuh; owner bisa unarchive/publish ulang kapan saja.
    - **Delete (oleh owner):** hanya diizinkan kalau `reviews_count = 0` dan `inquiries_count = 0` — kalau sudah ada data, owner wajib pakai Archive, bukan delete.
    - **Force delete:** hanya Superadmin, bisa hapus meski ada review/inquiry (cascade delete semua data terkait).
11. **Kelola inquiry masuk** — owner menerima notifikasi tiap ada inquiry baru, membalas (hanya 1x reply per inquiry, status berubah jadi `replied`), atau menutup manual. Inquiry auto-close setelah `auto_close_days` (default 30 hari) kalau tidak ada aksi.
12. **Lihat stats/analytics merchant** — owner melihat metrik yang di-update realtime/background job: `item_count`, `review_count`, `rating` (avg dari review published), `favorite_count`, `inquiry_count`, `total_views`.
13. **Company lifecycle** — owner bisa set company `active` ↔ `inactive` manual (status company tidak langsung memengaruhi visibilitas merchant di bawahnya); company tidak bisa dihapus selama masih punya merchant aktif.

**Selesai:** Member Pro berhasil membuat listing bisnis, melewati (atau menunggu) approval, mengisi katalog item, dan memantau performa toko lewat stats.

## Keterlibatan Admin — 💻 Web/Backoffice

**Admin Regional (scope: hanya region miliknya):**
1. Melihat & memfilter company/merchant di region sendiri (merchant tipe `online` tanpa region tetap terlihat oleh semua admin untuk keperluan moderasi).
2. **Approve merchant** yang `pending_approval` → status `published`, `published_at` diisi, notifikasi terkirim ke owner.
3. **Reject merchant** → wajib mengisi `rejection_reason`, notifikasi ke owner berisi alasan tsb, owner bisa edit & resubmit.
4. **Ban merchant** (region sendiri) → merchant langsung tidak tampil dan tidak bisa diubah owner; alasan ban wajib dicatat; notifikasi ke owner.
5. **Moderasi review** (region sendiri) — kalau `review.require_moderation` aktif, approve/reject review yang `pending`; bisa juga menyembunyikan (`hidden`) review yang sudah `published` tanpa menghapusnya.
6. Melihat analytics — hanya untuk region sendiri.
7. **Tidak bisa**: force delete merchant, kelola category master, ubah directory settings.

**Superadmin (scope: global):**
1. Semua aksi Admin Regional, tapi lintas semua region.
2. **Force delete** company/merchant (cascade ke semua data terkait) — kewenangan eksklusif Superadmin.
3. **Unban merchant** — hanya Superadmin yang bisa mencabut ban.
4. Kelola **category master** (tambah/edit/nonaktifkan kategori bisnis) — kategori yang dinonaktifkan hilang dari pilihan baru tapi tidak memengaruhi merchant existing, dan tidak bisa dihapus selama masih dipakai merchant.
5. Kelola **directory settings** global — mis. `merchant.require_approval`, `review.require_moderation`, `item.allow_free_price`, `location.require_location_for_physical`, dan seluruh setting lain yang disebut di section 11 `DIRECTORY_RULES.md`.
6. Melihat analytics global (semua region).

## Di Luar Cakupan Standard & Pro

- **Verifikasi badge resmi merchant** (verified/official badge) — keputusan akhir ada di tangan Superadmin, lihat [09_VERIFICATION_BADGE_JOURNEY.md](09_VERIFICATION_BADGE_JOURNEY.md). Catatan: `DIRECTORY_RULES.md` v2.0 (§17, per 2026-05-30) masih mencantumkan "Merchant verification badge" sebagai future feature/out-of-scope MVP — kalau fitur ini sudah aktif di modul lain, kemungkinan dokumen Directory ini belum di-update; perlu dikonfirmasi ke pemilik modul sebelum dijadikan acuan pasti.
- **Kelola category master** (tambah/nonaktifkan kategori bisnis) — hanya Superadmin, bukan hak member sama sekali.
- **Kelola directory settings global** (approval requirement, moderation toggle, dll) — hanya Superadmin.
- **Force delete merchant/company yang sudah ada review/inquiry** — hanya Superadmin; owner hanya bisa Archive.
- **Unban merchant** — hanya Superadmin, Admin Regional maupun owner tidak bisa mencabut ban sendiri.
- **Approve/reject merchant, item, atau review milik sendiri** — moderasi mutlak di tangan admin (Regional/Superadmin), owner tidak bisa self-approve.

## Edge Case & Catatan Tambahan

- **Kontradiksi soal analytics merchant owner:** §3.1 `DIRECTORY_RULES.md` secara eksplisit mendaftar "View own stats/analytics ✅" sebagai hak Member Pro, tapi §17 (Future Features/Out of Scope MVP) mencantumkan "Analytics dashboard untuk merchant owner" sebagai fitur yang **belum** dikerjakan. Dokumen ini mengikuti §3.1 (menyertakan Journey 2 langkah 12) tapi ambiguitas ini perlu dikonfirmasi — kemungkinan yang dimaksud §17 adalah dashboard analytics yang lebih kaya/visual, sedangkan metrik dasar (item_count, review_count, rating, favorite_count, inquiry_count, total_views) di §5.6 sudah tersedia.
- **Company `inactive` tidak memengaruhi visibilitas merchant** — hanya status company itu sendiri yang berubah; merchant di bawahnya tetap bisa `published` dan tampil normal.
- **Company/merchant `banned`** → seluruh merchant di bawah company yang banned otomatis ikut tidak tampil di listing.
- **Like bersifat idempotent by design** — like/unlike dua kali berturut-turut tidak dianggap error (409 boleh diabaikan FE), selaras dengan pola optimistic-update di mobile.
- **Reviewer tidak bisa review merchant sendiri** — berlaku juga untuk Member Pro yang jadi owner merchant tersebut.
- **Data retention:** review bersifat permanen (tidak auto-delete); merchant yang dihapus meng-cascade delete item & favorite, tapi review dan inquiry tetap dipertahankan di database untuk keperluan audit.

## Note
- belum implement like merchant
- inquiry masih ada bug subject min dan message min chars
- create merchant gambar yang di upload tidak muncul di preload confirmation, dan yng di kirim ke backend atau s3 malah placeholder image.
- create item gambar yang di upload tidak muncul di preload confirmation, dan yng di kirim ke backend atau s3 malah placeholder image.
- aproved merchant request belum mengirim notifikasi ke owner, tapi merchant sudah muncul di listing publik.
- tidak ada detail item di halaman merchant, hanya menampilkan list item saja.
- status closed and open merchant belum sesuai jam operasional yang di set.
- create item, min dan max price untuk service belum sesuai validasi, belum bisa di simpan juga, nilai yang di tampilkan di list masih belum muncul.
- jika inquiry sudah di reply, status inquiry masih tetap open, seharusnya berubah menjadi replied. dimana liat reply nya dari sisi penanya juga belum ada, seharusnya bisa liat reply dari merchant.
- di halaman backoffice, belum ada halaman untuk manage company, merchant, item, review, inquiry miliknya sendiri.
