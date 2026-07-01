# News Module — Backend API Issue Report

Dokumen ini merinci ketidaksesuaian (inconsistency) dan masalah (bugs) yang ditemukan pada endpoint API News Mobile (`/api/v1/mobile/news/*`), dites langsung ke backend dev (`http://192.168.1.29:8888`) menggunakan akun `test_member`. Masalah ini menyebabkan tombol **translate** tidak pernah muncul dan tidak ada cara menampilkan **link sumber asli** artikel hasil scraping.

Referensi: `docs/api_spec/API_SPEC_NEWS_MOBILE.md`, `docs/modules/News/NEWS_RULES.md`.

---

## 1. Masalah Utama: Tombol Translate Tidak Pernah Muncul

### Deskripsi

Tombol translate di detail artikel (mobile) hanya dirender jika `available_languages` punya lebih dari 1 entri:

```dart
// lib/features/news/presentation/screens/news_detail_screen.dart:289
if (a.availableLanguages.length > 1)
  PopupMenuButton<String>(...)
```

Namun field `available_languages` **tidak pernah dikirim** oleh backend — baik di list (`GET /articles`) maupun detail (`GET /articles/{id}`). Akibatnya `available_languages` selalu terparse jadi `[]` (list kosong) di client, jadi kondisi `> 1` selalu `false` dan tombol translate tidak pernah dirender, walaupun artikel yang sama sudah punya `is_translated: true` (artinya secara data translation-nya memang ada).

### Bukti Payload — `GET /mobile/news/articles/{id}`

```json
{
  "id": "7184b36e-3cb8-45fe-9591-ec5af47385eb",
  "language": "id",
  "title": "Samsung, Shinhan join Open USD stablecoin network",
  "author_label": "Korean Association Indonesia",
  "tags": [],
  "is_translated": true,
  "is_featured": false,
  "like_count": 1,
  "comment_count": 2,
  "view_count": 9,
  "is_liked": true,
  "is_bookmarked": false,
  "published_at": "2026-07-01T09:00:08.406863Z"
  // <-- tidak ada "available_languages" sama sekali (spec: array string, mis. ["id","en","ko"])
}
```

Field yang didokumentasikan di spec (`API_SPEC_NEWS_MOBILE.md` §Article Object Detail) tapi absen di response asli: `available_languages`.

### Dampak Pada Mobile

- Tombol translate (ikon 🌐 di app bar detail artikel) tidak pernah muncul untuk artikel manapun.
- User tidak bisa memilih bahasa baca lain walau backend sudah mendukungnya (`is_translated: true` membuktikan translation sudah pernah diproses untuk artikel ini).

---

## 2. Masalah: Response `POST/GET .../translate` Tidak Sesuai Spec — Berisiko Mengosongkan Artikel

### Deskripsi

Spec mendefinisikan dua bentuk response untuk `GET /articles/{id}/translate?language=...`:
- **200** → `data: { article_id, language, title, content, is_translated, translate_status: "done" }`
- **202** → `data: { ..., translate_status: "processing", fallback: { language, title, content } }`

Response asli dari backend dev untuk endpoint ini:

```json
{"status":"success","status_code":200,"message":"translation queued","data":null,"meta":null}
```

`data` bernilai `null` — bukan object seperti di spec. Ini konsisten untuk request bahasa `en` maupun `ko`.

### Dampak Pada Mobile

`TranslatedArticleMapper.fromJson` (`lib/features/news/data/models/news_models.dart`) memanggil `_objOf(raw)`, yang fallback ke `raw` itu sendiri kalau `data` bukan Map. Karena `data` adalah `null` (bukan Map), field yang dibaca (`article_id`, `title`, `content`, dst) semuanya tidak ketemu di envelope `{status, message, data, meta}` dan di-default jadi string kosong. Jika tombol translate suatu saat sudah muncul (setelah Masalah #1 diperbaiki) dan user memilih bahasa, artikel yang sedang dibaca akan **berubah jadi kosong** (`title=''`, `content=''`) karena hasil mapping ini langsung dipakai untuk update state artikel di `_translateTo()`.

Selain itu, semantik "queued" (antrian job) berbeda dari kontrak spec yang mengasumsikan hasil langsung tersedia atau setidaknya ada `fallback` berisi konten original untuk ditampilkan sementara. Dengan `data: null`, client tidak punya apa pun untuk ditampilkan selama proses translate berjalan.

---

## 3. Masalah: `original_url` Tidak Ada di Response — Tidak Bisa Menampilkan Link Sumber Asli

### Deskripsi

Untuk artikel hasil scraping, `NEWS_RULES.md` mendokumentasikan `original_url` sebagai atribut Article ("URL asli artikel di sumbernya... Contoh: `https://detik.com/sport/12345`"). Domain entity mobile (`NewsArticle.originalUrl`) dan mapper JSON (`_strOrNull(j['original_url'])`) sudah siap menerima field ini — tapi:

1. **Field ini bahkan tidak didokumentasikan** di skema response `Article Object (Detail)` pada `API_SPEC_NEWS_MOBILE.md` (baris 74–106) — spec-nya sendiri belum mencantumkan `original_url`.
2. **Response asli backend juga tidak mengirim field ini** (lihat payload lengkap di Masalah #1 — tidak ada `original_url`).

### Dampak Pada Mobile

Tidak ada cara untuk merender tombol/link "Baca sumber asli" di halaman detail artikel — field `a.originalUrl` di kode saat ini memang belum pernah dipakai di layer presentation, dan sekalipun ditambahkan, tidak ada data untuk ditampilkan.

### Catatan Terkait "Publisher Asli"

Per `NEWS_RULES.md` (bagian *Visibility: Tidak Scoped + Label Asal*), label yang **memang dimaksudkan** untuk tampil ke pembaca adalah `author_label` / `published_by_label` — yaitu label organisasi KAI (mis. "Korean Association Indonesia" atau "KAI Jakarta"), **bukan** nama portal berita asli (mis. "The Korea Herald"/`heraldcorp.com`). Ini terkonfirmasi juga secara eksplisit di rules: *"Artikel hasil scraping → label KAI Pusat"*.

Jadi ini **bukan bug** — secara desain, publisher yang ditampilkan adalah KAI, bukan portal sumber. Jika produk ingin pembaca tahu portal asli (mis. untuk kredibilitas/atribusi), maka yang perlu ditambahkan bukan field publisher baru, melainkan cukup `original_url` (Masalah #3 di atas) sehingga user bisa klik untuk melihat sumber aslinya langsung di website tersebut.

---

## Rekomendasi Perbaikan di Backend

1. **Isi `available_languages` di response `/articles` (list & detail).**
   Hitung dari baris `article_translations` yang `translate_status = 'done'` (atau minimal `[original_language]` jika belum ada translation lain). Tanpa ini, fitur translate di mobile tidak akan pernah terlihat oleh user.

2. **Perbaiki response `/articles/{id}/translate` agar `data` tidak `null`.**
   Minimal kembalikan bentuk sesuai spec: `{ article_id, language, translate_status: "processing", fallback: { language, title, content } }` saat job baru di-queue, supaya client punya konten fallback yang valid untuk ditampilkan sambil menunggu — dan `{ ..., title, content, translate_status: "done" }` saat translation sudah tersedia (idealnya endpoint ini dipanggil ulang / di-poll setelah job selesai).

3. **Tambahkan `original_url` ke response `/articles/{id}` (dan idealnya `/articles` list) untuk artikel hasil scraping.**
   Sekaligus update `API_SPEC_NEWS_MOBILE.md` untuk mencantumkan field ini di skema `Article Object (Detail)`, karena saat ini spec pun belum mendokumentasikannya.

4. **(Terkait, sudah dilaporkan terpisah ke tim)** Field `category` dan `scope` juga belum dikembalikan sama sekali oleh `/articles` (list maupun detail) walau sudah ada di spec — akan membuat badge kategori/scope di kartu & detail artikel tetap kosong sampai backend menambahkannya.
