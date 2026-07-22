# Issue: Directory — like, inquiry, upload gambar, notifikasi approval, jam operasional, harga service

- **Modul**: Directory — merchant (like, jam operasional, notifikasi approval), item (upload gambar, harga service, detail), inquiry (validasi, status, reply), backoffice self-service
- **Severity**: 🔴 Tinggi untuk Issue 3, 6, 7 (fitur inti gagal/salah data) — 🟠 Sedang untuk Issue 1, 4, 5, 8 — 🟢 Rendah untuk Issue 2 (UX, bukan blocker fungsional) dan Issue 9 (gap desain, bukan regresi)
- **Status**: 🔴 Open — 9 issue di dokumen ini, semua root cause sudah ditemukan by code review
- **Ditemukan**: 22 Jul 2026, saat review user journey `06_DIRECTORY_JOURNEY.md`
- **Pelapor**: review manual (dev), dikonfirmasi via code review langsung ke `k-forum-api`, `k_forum`, dan `k-forum-backoffice`

---

## Issue 1 — Like merchant belum diimplementasi sama sekali

- **Repo**: `k-forum-api`

### Ringkasan

`DIRECTORY_API_SPEC_MOBILE_V2.md` mendefinisikan endpoint like (baris 93-95: `POST/DELETE /merchants/{id}/like`, terpisah dari save/bookmark) dan field `is_liked`/`favorite_count` di response merchant — tapi tidak ada satupun bagian dari fitur ini yang benar-benar dibangun di backend.

### Root cause (code review)

- Tidak ada usecase `like_merchant.go`/`unlike_merchant.go` di `internal/app/usecase/directory/` — hanya ada `save_merchant.go`/`unsave_merchant.go` (fitur **save/bookmark**, secara spec memang terpisah dari like ❤️, lihat baris 88-95 spec).
- Tidak ada handler/route apapun untuk `like` di `internal/interfaces/http/handler/mobile/directory_handler.go` maupun `router.go`.
- Field `favorite_count` ADA di beberapa DTO (`directory_dto.go`) dan dipetakan dari kolom `FavoriteCount` — tapi karena tidak ada aksi like/unlike yang meng-increment/decrement-nya, nilainya tidak pernah berubah dari default (kemungkinan selalu 0 kecuali di-seed manual).
- Field `is_liked` (wajib ada di response list per spec baris 590) **tidak ada sama sekali** di DTO manapun — bukan cuma logic-nya kosong, field-nya memang tidak dideklarasikan.

### Yang diminta ke backend (k-forum-api)

1. Buat usecase `LikeMerchant`/`UnlikeMerchant`, tabel relasi like (mirip pola `dir_merchant_saves` untuk save/bookmark yang sudah ada), dan endpoint `POST|DELETE /mobile/directory/merchants/{id}/like`.
2. Tambah field `is_liked` (boolean, `null` jika belum login) ke semua DTO yang me-return merchant list/detail, dan pastikan `favorite_count` benar-benar dihitung dari tabel like yang baru.

### Kriteria selesai (acceptance)

- [ ] Member like merchant → `favorite_count` bertambah, `is_liked: true` di response berikutnya.
- [ ] Unlike → berkurang lagi, `is_liked: false`.
- [ ] Guest (belum login) → `is_liked: null` (bukan `false`, sesuai spec).

---

## Issue 2 — Error validasi inquiry (subject/message terlalu pendek) menampilkan kode mentah, bukan pesan manusiawi

- **Repo**: `k-forum-api`

### Ringkasan

Saat subject atau message inquiry terlalu pendek, pesan error yang diterima bukan kalimat yang bisa dipahami user, tapi kode internal mentah.

### Root cause (code review)

Validasi panjang minimum SUDAH benar diterapkan di `internal/domain/directory/entity/inquiry.go:107-127` (`minInquirySubjectLength = 5`, `minInquiryMessageLength = 10`, baris 12-15) — bukan ini yang jadi masalah utama, tapi cara errornya disampaikan ke user:

1. `send_inquiry.go` menangkap error validasi lewat `mapDomainValidationError()` (`internal/app/usecase/directory/helpers.go:164-170`), yang memanggil `apperror.Unprocessable(string(de.Code), nil)` — **parameter pertama `Unprocessable` itu memang bernama `message`**, jadi kode error mentah (`"DOMAIN_DIRECTORY_INQUIRY_SUBJECT_TOO_SHORT"`) dikirim langsung sebagai isi pesan, bukan sebagai key untuk translasi.
2. Di layer HTTP (`internal/interfaces/http/httperror/http_error.go:80,90-95`), pesan ini memang DICOBA di-translate lewat `translate(locale, httpErr.message)` — tapi fallback-nya, sesuai komentar kode sendiri ("Fallback ke key itu sendiri"), adalah **mengembalikan key itu sendiri kalau tidak ketemu di locale bundle**.
3. Dicek langsung: `locales/id.json`, `en.json`, `ko.json` **tidak punya entry** untuk `DOMAIN_DIRECTORY_INQUIRY_SUBJECT_TOO_SHORT` / `..._MESSAGE_TOO_SHORT` (atau varian TOO_LONG-nya).

Hasil akhir: user melihat literally `"DOMAIN_DIRECTORY_INQUIRY_SUBJECT_TOO_SHORT"` sebagai pesan error, bukan kalimat seperti "Subjek minimal 5 karakter".

**Temuan tambahan (skala lebih besar dari sekadar inquiry)**: dari 101 domain error code yang didefinisikan di `internal/domain/directory/constant/directory_constant.go`, cuma **16 yang punya entry locale**. 85 kode lainnya — kemungkinan besar termasuk validasi merchant, item, review lain di modul Directory — berpotensi punya masalah yang sama persis, cuma belum ketauan karena belum dicoba satu-satu.

### Yang diminta ke backend (k-forum-api)

1. Tambah entry locale (id/en/ko) untuk minimal 4 kode inquiry ini: `DOMAIN_DIRECTORY_INQUIRY_SUBJECT_TOO_SHORT`, `..._SUBJECT_TOO_LONG`, `..._MESSAGE_TOO_SHORT`, `..._MESSAGE_TOO_LONG`.
2. Audit menyeluruh: bandingkan semua 101 kode di `directory_constant.go` dengan isi `locales/*.json`, lengkapi yang masih kosong (bukan cuma inquiry).

### Kriteria selesai (acceptance)

- [ ] Kirim inquiry dengan subject < 5 karakter → pesan error berbahasa manusia sesuai locale request (id/en/ko), bukan kode mentah.
- [ ] Sama untuk message < 10 karakter, dan untuk batas atas (subject > 200, message > 2000).

---

## Issue 3 — Upload gambar create merchant & create item: preview tidak sesuai, yang terkirim ke server malah foto placeholder acak

- **Repo**: `k_forum`
- **Severity**: 🔴 Tinggi — user tidak sadar fotonya gagal terupload karena tidak ada pesan error apapun

### Ringkasan

Saat membuat merchant baru atau item baru dan memilih gambar, foto yang benar-benar diupload ke server BUKAN foto yang dipilih user — melainkan foto stok acak dari internet (`picsum.photos`) — dan ini terjadi tanpa notifikasi error apapun ke user.

### Root cause (code review)

Backend sudah benar — route presign yang terdaftar di `internal/interfaces/http/router/router.go:1020-1024`:
```
POST /mobile/directory/merchants/media/image/presign
POST /mobile/directory/items/media/image/presign
```

Tapi mobile di KEDUA form memanggil path yang **hilang satu segmen** (`merchants/`/`items/`):

`lib/features/directory/presentation/screens/merchant_form_screen.dart:267`:
```dart
presignEndpoint: '/mobile/directory/media/image/presign',
```

`lib/features/directory/presentation/screens/item_form_screen.dart:128`:
```dart
presignEndpoint: '/mobile/directory/media/image/presign',
```

Path ini tidak match route manapun → request presign selalu gagal (404). Masalah jadi lebih parah karena `_uploadOrFallback()` (merchant_form_screen.dart:258-278, pola identik di item_form_screen.dart) membungkus panggilan presign dengan `try { ... } catch (_) { // ignore — use placeholder }` — begitu presign gagal, **tanpa pesan error apapun ke user**, fungsi ini diam-diam mengembalikan:
```dart
final fallbackUrl = 'https://picsum.photos/seed/${file.path.hashCode.abs()}/600/400';
```
— foto stok acak dari layanan Picsum — dan URL inilah yang masuk ke `_images` list, ditampilkan di preview, dan akhirnya benar-benar dikirim ke `create merchant`/`create item` sebagai gambar produk. User mengira upload sukses karena TIDAK ADA sinyal kegagalan sama sekali.

### Yang diminta ke mobile (k_forum)

1. Perbaiki `presignEndpoint` di kedua file jadi `/mobile/directory/merchants/media/image/presign` dan `/mobile/directory/items/media/image/presign` sesuai route yang benar.
2. **Hapus fallback placeholder senyap.** Kalau presign gagal, tampilkan error yang jelas ke user (snackbar) dan JANGAN masukkan apapun ke `_images` — jangan biarkan flow lanjut seolah berhasil. Silent-fallback seperti ini berbahaya karena bisa menutupi kegagalan lain di masa depan (network flaky, dsb.) tanpa disadari siapapun.

### Kriteria selesai (acceptance)

- [ ] Create merchant/item, pilih gambar → foto yang benar-benar dipilih user tampil di preview DAN tersimpan di server/S3 (bukan foto Picsum).
- [ ] Simulasikan presign gagal (misal matikan endpoint sementara) → user melihat error yang jelas, tidak ada foto placeholder yang menyusup diam-diam.

---

## Issue 4 — Merchant di-approve admin tidak kirim notifikasi ke owner, meski sudah live di listing publik

- **Repo**: `k-forum-api`

### Ringkasan

Setelah admin approve pengajuan merchant, merchant tersebut langsung tampil di direktori publik (bagian ini benar) — tapi owner-nya tidak pernah diberi tahu bahwa pengajuannya disetujui.

### Root cause (code review)

`ApproveMerchantUseCase` (`internal/app/usecase/directory/approve_merchant.go`) publish event `direvent.MerchantPublished{ MerchantID, OccurredAt }` — **payload-nya bahkan tidak menyertakan owner user ID**. Satu-satunya consumer event ini adalah `SearchSyncHandler.HandleMerchantPublished` (`internal/interfaces/mq/handler/search_sync_handler.go:181-190`), yang isinya cuma `h.merchantIndexer.IndexOne(ctx, payload.MerchantID)` — meng-index merchant ke search engine, itu sebabnya merchant langsung muncul di listing. **Tidak ada consumer yang mengirim notifikasi apapun ke owner.**

Dicek juga `reject_merchant.go` — usecase ini **tidak publish event apapun sama sekali**, jadi owner juga tidak diberitahu saat pengajuannya ditolak.

### Yang diminta ke backend (k-forum-api)

1. Tambah dispatch notifikasi (in-app + push, ikuti pola modul lain seperti `HandleRegionJoinApproved`) ke owner saat merchant di-approve — bisa lewat consumer baru untuk event yang sama, atau event terpisah khusus notifikasi.
2. Tambah event + notifikasi serupa untuk `reject_merchant.go` (owner juga berhak tahu kalau ditolak, termasuk alasannya kalau ada).

### Kriteria selesai (acceptance)

- [ ] Admin approve merchant → owner menerima notifikasi in-app/push "Merchant Anda disetujui" dalam waktu wajar.
- [ ] Admin reject merchant → owner menerima notifikasi serupa, idealnya menyertakan alasan reject.

---

## Issue 5 — Tidak ada halaman detail item, cuma list

- **Repo**: `k_forum`

### Ringkasan

Item milik merchant hanya tampil sebagai kartu di list halaman merchant — tidak bisa dibuka untuk melihat detailnya (deskripsi lengkap, semua gambar, dst).

### Root cause (code review)

Backend sudah lengkap — route `GET /mobile/directory/merchants/:merchant_id/items/:item_id` terdaftar (`router.go:1032`) dan usecase `GetMerchantItemDetail` sudah diimplementasikan.

Mobile juga sudah punya usecase yang wired ke endpoint ini (`lib/features/directory/domain/usecases/get_item_detail_usecase.dart`) — tapi **tidak pernah dipanggil dari manapun** (nol pemanggilan di seluruh `lib/`, dan tidak ada file `item_detail_screen.dart`). `MerchantItemCard` (`lib/features/directory/presentation/widgets/merchant_item_card.dart`), widget yang dipakai untuk render tiap item di `merchant_detail_screen.dart:400`, juga tidak punya `onTap`/`GestureDetector`/`InkWell` apapun — jadi item benar-benar tidak bisa di-tap sama sekali.

### Yang diminta ke mobile (k_forum)

1. Buat screen `item_detail_screen.dart` yang memanggil `GetItemDetailUseCase` yang sudah ada.
2. Tambah `onTap` di `MerchantItemCard` (atau di pembungkusnya) untuk navigasi ke screen tersebut.

### Kriteria selesai (acceptance)

- [ ] Tap item di halaman merchant → terbuka halaman detail item (deskripsi, semua gambar, harga/range harga, dst).

---

## Issue 6 — Status buka/tutup merchant tidak sesuai jam operasional yang di-set

- **Repo**: `k-forum-api`

### Ringkasan

`is_open_now` sering tidak sesuai jam operasional yang sudah diatur owner — bukan karena data jam-nya salah, tapi karena logic pengecekannya memakai jam server, bukan jam yang relevan untuk platform.

### Root cause (code review)

`isOpenNow()` (`internal/app/usecase/directory/hours_helper.go:16-42`) memakai `now := time.Now()` langsung — waktu proses Go di server (biasanya UTC di container), **tanpa konversi ke timezone apapun**, lalu dibandingkan mentah-mentah dengan jam `HH:MM` yang di-input owner (yang mengasumsikan waktu lokal mereka, bukan UTC).

Ini persis kelas bug yang sama dengan yang sudah diperbaiki di modul Ads (`internal/app/usecase/ads/helpers.go:20-32`, komentar kode: *"active today", dianchor ke setting `default_timezone` platform — bukan jam proses Go (yang bisa UTC di server)*) — Directory business hours memang **belum ikut dapat fix ini**, sudah tercatat sebagai item belum selesai sejak fix Ads timezone sebelumnya.

### Yang diminta ke backend (k-forum-api)

1. Di `hours_helper.go`, ganti `time.Now()` mentah dengan resolve `default_timezone` dari System Settings, mengikuti pola yang sama seperti `internal/app/usecase/ads/helpers.go` (`sysvo.DefaultPlatformTimezone()` + override dari setting `default_timezone` kalau ada).

### Kriteria selesai (acceptance)

- [ ] Merchant dengan jam operasional 09:00–17:00 WIB menunjukkan `is_open_now: true` saat benar-benar jam 09:00-17:00 WIB, terlepas dari timezone server.
- [ ] Kasus overnight (mis. 22:00–02:00) tetap benar setelah fix (logic overnight yang sudah ada di baris 37-39 tidak boleh regresi).

---

## Issue 7 — Create item service: harga (price_min/price_max) tidak konsisten — shape response salah, nilai tidak tampil di list

- **Repo**: `k-forum-api`

### Ringkasan

Untuk item tipe "service" (pakai range harga, bukan harga tunggal), nilai harga yang di-set tidak muncul saat item ditampilkan di list.

### Root cause (code review)

Validasi di domain entity (`internal/domain/directory/entity/merchant_item.go:151-160`) sudah benar sesuai spec (`price_min` wajib > 0, `price_max` wajib >= `price_min`) — begini juga DTO request (`CreateItemInput` di `directory_dto.go:336-350`, field `price_min`/`price_max` sudah match nama JSON di spec.

Masalahnya ada di **response**: `buildItemResponse()` (`internal/app/usecase/directory/create_item.go:112-138`) membungkus harga service jadi objek nested:
```go
resp.PriceRange = &dto.ItemPriceRangeResponse{
    Min:      it.PriceMin,
    Max:      it.PriceMax,
    Currency: it.Currency,
}
```
Tapi spec (`DIRECTORY_API_SPEC_MOBILE_V2.md` baris 707-708, dan mobile: `lib/features/directory/data/models/directory_models.dart:229-230` — `priceMin: _intN(j, 'price_min')`) mengharapkan field **flat** `price_min`/`price_max` langsung di object item, bukan di-nest di bawah `price_range`. Karena backend kirim `price_range.min`/`price_range.max` sementara mobile parsing `j['price_min']`/`j['price_max']` langsung — hasilnya selalu `null`, sehingga harga service tidak pernah tampil di list mobile manapun.

Untuk gejala "belum bisa disimpan": dari code review, validasi backend (domain entity) dan validasi client (`item_form_screen.dart:150-160`) **keduanya sudah benar secara logic** — tidak ditemukan bug yang jelas-jelas menolak input valid saat create. Kemungkinan besar gejala ini adalah **efek lanjutan dari shape mismatch di atas**: item sebenarnya BERHASIL dibuat di server, tapi begitu kembali ke list (yang harganya tidak muncul karena bug di atas), terlihat seperti "gagal tersimpan" walau sebenarnya tersimpan tanpa harga yang terbaca dengan benar. Perlu diverifikasi ulang setelah Issue ini diperbaiki — kalau ternyata masih ada kasus create benar-benar gagal, itu perlu dilaporkan sebagai issue terpisah.

### Yang diminta ke backend (k-forum-api)

1. Ubah `buildItemResponse()` supaya `price_min`/`price_max` dikirim flat langsung di object item (sesuai spec), bukan di-nest di `price_range`. Cek juga endpoint list/detail item lain yang mungkin memakai pola nested serupa untuk konsistensi.

### Kriteria selesai (acceptance)

- [ ] Create item tipe service dengan `price_min`/`price_max` → response create dan response list/detail selanjutnya sama-sama punya field flat `price_min`/`price_max` sesuai spec.
- [ ] List item merchant menampilkan range harga service dengan benar (bukan kosong).
- [ ] Setelah fix, retest apakah "tidak bisa disimpan" masih terjadi — kalau masih, buka issue baru dengan detail reproduksi yang lebih spesifik.

---

## Issue 8 — Status inquiry tidak berubah jadi "replied", dan penanya tidak bisa lihat balasan merchant

- **Repo**: `k_forum` (100% mobile — backend write & read path sudah dikonfirmasi benar)

### Ringkasan

Setelah merchant membalas inquiry, dari sisi penanya (customer) statusnya tetap terlihat seperti belum dibalas, dan tidak ada tempat untuk melihat isi balasannya sama sekali.

### Root cause (code review)

**Backend sudah benar di semua layer** — dikonfirmasi dengan membaca kode dari ujung ke ujung:
- `Inquiry.Reply()` (`internal/domain/directory/entity/inquiry.go:77-91`) benar meng-set `i.Status = constant.InquiryStatusReplied` saat dibalas.
- `PostgresInquiryRepository.Update()` (`internal/infrastructure/persistence/postgres_directory_repository.go:795-808`) benar menyimpan `status=$3` ke kolom `dir_inquiries.status`.
- `ListMyInquiries` (query untuk `GET /me/inquiry`, `postgres_directory_query.go:706-743`) benar membaca `i.status` langsung dari tabel, tidak ada cache/denormalisasi yang bisa basi.

**Yang benar-benar hilang: tidak ada screen di mobile untuk customer melihat inquiry mereka sendiri sama sekali.** Usecase `get_my_inquiries_usecase.dart` sudah ada dan wired ke repository/data source, tapi **nol pemanggilan** di seluruh `lib/` (dicek dengan grep menyeluruh) — tidak ada file `presentation/screens/*inquir*` untuk sisi customer. Satu-satunya pemakaian `ReplyInquiryUseCase` yang ditemukan ada di `merchant_manage_screen.dart:315` — itu punya **owner** (untuk membalas), bukan punya **penanya** (untuk melihat balasan).

Karena tidak ada screen sama sekali, penanya tidak punya cara melihat status ATAU balasan yang update — persepsi "status masih tetap open" kemungkinan besar berasal dari state awal saat inquiry pertama dikirim (`status: "pending"`) yang tidak pernah di-refresh karena tidak ada screen yang men-fetch ulang dari server.

### Yang diminta ke mobile (k_forum)

1. Buat screen "My Inquiries" untuk sisi customer, memakai `GetMyInquiriesUseCase` yang sudah ada — tampilkan daftar inquiry beserta status (pending/replied/closed) dan, kalau sudah dibalas, tampilkan `merchant_reply` beserta `replied_at`.
2. Tambahkan entry point yang jelas (misal dari halaman profil, atau dari merchant detail setelah kirim inquiry).

### Kriteria selesai (acceptance)

- [ ] Customer kirim inquiry, merchant balas → customer bisa lihat status berubah jadi "Replied" dan bisa baca isi balasannya, dari sebuah screen di app (bukan cuma dari notifikasi push kalau ada).

---

## Issue 9 — Backoffice belum punya halaman self-service untuk owner (company/merchant/item/review/inquiry milik sendiri)

- **Repo**: `k-forum-backoffice`
- **Severity**: 🟢 Rendah — ini gap desain/fitur, bukan regresi; mobile app sudah punya self-service untuk sebagian besar hal ini (create/edit merchant, create/edit item, balas inquiry via `merchant_manage_screen.dart`)

### Ringkasan

Tidak ada halaman di backoffice untuk seorang Pro member (owner merchant) mengelola company/merchant/item/review/inquiry miliknya sendiri lewat web — satu-satunya halaman Directory di backoffice murni untuk admin platform.

### Root cause (code review)

`app/pages/directory/index.vue:2-5`:
```js
definePageMeta({
  middleware: 'permission',
  permission: { roles: ['usergod', 'super admin', 'superadmin', 'admin', 'admin region'], permissions: ['manage_directory'] },
})
```
Cuma satu halaman, khusus role admin platform + permission `manage_directory`. Tidak ada tab/halaman terpisah yang di-scope ke "merchant/company/item/review/inquiry milik user yang login", dan tidak ada mekanisme apapun (mirip masalah yang sama ditemukan di Community — lihat `COMMUNITY_MODULE_ISSUES.md` Issue 6) untuk owner mengelola resource miliknya sendiri dari sisi backoffice.

### Yang diminta ke backoffice (k-forum-backoffice)

1. Kalau memang diinginkan self-service via web (bukan cuma mobile) — buat halaman terpisah (misal `/directory/mine`) yang di-scope ke company/merchant/item/review/inquiry milik user yang login, mengikuti pola yang sama seperti rekomendasi di Community Issue 6.
2. Kalau ternyata self-service via mobile sudah cukup dan web tidak diperlukan — dokumentasikan keputusan ini di `DIRECTORY_RULES.md` supaya tidak dianggap "belum selesai" di review berikutnya.

### Kriteria selesai (acceptance)

- [ ] Keputusan produk jelas: self-service directory disediakan di backoffice atau tidak. Kalau ya, owner bisa login ke backoffice dan hanya melihat/mengelola company/merchant/item/review/inquiry miliknya sendiri.

---

## Referensi

- Journey terkait: [`flows/user-journeys/06_DIRECTORY_JOURNEY.md`](../flows/user-journeys/06_DIRECTORY_JOURNEY.md).
- Spec: `Modules/Directory/DIRECTORY_RULES.md`, `Modules/Directory/MERCHANT_LIKE_SPEC.md`, `API SPEC/Mobile/DIRECTORY_API_SPEC_MOBILE_V2.md`.
- Kode kunci — `k-forum-api`: `internal/app/usecase/directory/{send_inquiry,reply_inquiry,approve_merchant,reject_merchant,create_item,hours_helper}.go`, `internal/domain/directory/entity/{inquiry,merchant_item}.go`, `internal/infrastructure/persistence/{postgres_directory_query,postgres_directory_repository}.go`, `internal/interfaces/mq/handler/search_sync_handler.go`, `internal/interfaces/http/httperror/http_error.go`, `locales/*.json`.
- Kode kunci — `k_forum`: `lib/features/directory/presentation/screens/{merchant_form_screen,item_form_screen,merchant_detail_screen,merchant_manage_screen}.dart`, `lib/features/directory/presentation/widgets/merchant_item_card.dart`, `lib/features/directory/domain/usecases/{get_item_detail_usecase,get_my_inquiries_usecase}.dart`.
- Kode kunci — `k-forum-backoffice`: `app/pages/directory/index.vue`.
