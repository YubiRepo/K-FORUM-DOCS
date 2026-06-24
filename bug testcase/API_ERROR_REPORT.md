# Backend Issues — Mobile API

Daftar **endpoint yang error** hasil test mobile. Semua di environment **produksi/publik**.

- **Base URL:** `https://k-forum-api.yubicom.co.id/api/v1`
- **Tanggal test:** 23 Juni 2026
- **Auth:** login sebagai member biasa (token valid)
- **Sifat:** semua error = **`500 ERR_INTERNAL`** (unhandled exception di server) — **bukan** error validasi/permission.

> Pola umum: endpoint **list & detail jalan normal**, tapi handler "khusus" (featured/upcoming/search/like/comments) melempar 500. Kemungkinan bug di query/join/aggregasi handler tersebut. Mohon cek **log server** di tiap handler.

---

## 🔴 EVENT · Featured
- **Endpoint:** `GET /mobile/events/featured?limit=5`
- **Harusnya:** `200` + daftar event unggulan
- **Aktual:** `500` `ERR_INTERNAL`
- **Repro:**
  ```bash
  curl -H "Authorization: Bearer <token>" \
    "https://k-forum-api.yubicom.co.id/api/v1/mobile/events/featured?limit=5"
  # → {"status":"error","status_code":500,"error_code":"ERR_INTERNAL",...}
  ```
- **Dampak app:** strip "Featured" di layar Event blank/gagal.

## 🔴 EVENT · Upcoming
- **Endpoint:** `GET /mobile/events/upcoming?limit=10`
- **Harusnya:** `200` + daftar event mendatang
- **Aktual:** `500` `ERR_INTERNAL`
- **Repro:**
  ```bash
  curl -H "Authorization: Bearer <token>" \
    "https://k-forum-api.yubicom.co.id/api/v1/mobile/events/upcoming?limit=10"
  ```
- **Dampak app:** strip "Upcoming" blank/gagal.

## 🔴 EVENT · Search
- **Endpoint:** `GET /mobile/events/search?q=korea&page=1&limit=5`
- **Harusnya:** `200` + hasil pencarian
- **Aktual:** `500` `ERR_INTERNAL`
- **Repro:**
  ```bash
  curl -H "Authorization: Bearer <token>" \
    "https://k-forum-api.yubicom.co.id/api/v1/mobile/events/search?q=korea&page=1&limit=5"
  ```
- **Dampak app:** layar Search event gagal. (Catatan: `GET /mobile/events?search=korea` **jalan** — bisa jadi acuan untuk benerin handler `/search`.)

## ✅ NEWS · Like / Unlike — 500 SUDAH FIX (ada bug sisa)
- **Endpoint:** `POST` / `DELETE /mobile/news/articles/{id}/like`
- **Status:** sekarang `200` ✅ (`like_count` ikut naik/turun dengan benar).
- 🟠 **Bug sisa — `is_liked` selalu `false`:** setelah `POST like`, `GET /mobile/news/articles/{id}` tetap mengembalikan `is_liked: false` (padahal `like_count` sudah jadi 1). Akibatnya app nggak tahu user sudah like → ikon like balik ke status belum-disuka tiap kali detail di-reload.
  ```bash
  # before: is_liked=false like_count=0
  curl -H "Authorization: Bearer <token>" -X POST ".../mobile/news/articles/<id>/like"
  curl -H "Authorization: Bearer <token>" ".../mobile/news/articles/<id>"
  # after:  is_liked=false  like_count=1   ← is_liked harusnya true
  ```
- 🟠 **Bug sisa — response body kosong:** `POST`/`DELETE like` balikin `data:null` (tanpa `is_liked`/`like_count`). Idealnya balikin state terbaru `{ is_liked, like_count }` biar mobile bisa rekonsiliasi tanpa nebak.
- 🟠 **Bug sisa — list tanpa `is_liked`:** `GET /mobile/news/articles` (list) **nggak punya field `is_liked`** sama sekali (cuma `like_count`). Akibatnya kartu di list nggak tahu artikel sudah disuka → ikon love selalu kosong saat load. Mohon tambahkan `is_liked` per artikel di list.

> **Status app:** pakai **optimistic + rekonsiliasi** — ikon langsung berubah saat di-tap (UX), dan akan otomatis ikut nilai server begitu ke-3 perbaikan di atas (`is_liked` di like-response, detail, & list) tersedia. Tanpa perubahan kode lagi.

## ✅ NEWS · Comments — 500 SUDAH FIX (ada bug sisa)
- **Endpoint:** `GET /mobile/news/articles/{id}/comments`
- **Status:** sekarang `200` ✅ dan komentar non-deleted mengembalikan `content`.
- 🟠 **Bug sisa — `created_at` epoch-zero:** semua komentar balikin `"created_at":"0001-01-01T00:00:00Z"` → timestamp "x menit lalu" jadi salah. (Workaround app: tanggal epoch dianggap kosong.)
- 🟡 **Catatan bentuk:** author dikirim sebagai field flat `user_id`/`user_name`/`user_avatar` (bukan objek `user{}`). Mobile sudah disesuaikan, tapi mohon dikonsistenkan dengan modul lain bila perlu.
- 🟡 **`POST .../comments`** balikin cuma `{id}` (tanpa `content`/author) → app harus susun komentar baru dari input lokal. Idealnya balikin objek komentar lengkap.

---

## 🟠 ADS · `/my` — envelope tidak konsisten (nested `data.data`)

Bukan `500`, tapi **bentuk response beda sendiri** dari endpoint ads lain → bikin parser mobile crash
(`type '_Map<String, dynamic>' is not a subtype of type 'List<dynamic>?'`).

- **Endpoint:** `GET /mobile/ads/my`
- **Aktual** — list dibungkus dobel di `data.data`:
  ```json
  { "status":"success", "data": { "data": [ … ], "pagination": { "limit":20,"offset":0,"total":0 } } }
  ```
- **Endpoint ads lain konsisten pakai named key** (tidak nested ganda):
  | Endpoint | `data` berisi |
  | --- | --- |
  | `GET /mobile/ads/home` | `{ slider[], feed_ads[], feed_interval }` |
  | `GET /mobile/ads` (promo) | `{ slider[], ads[], pagination }` |
  | `GET /mobile/ads/my` | `{ data[], pagination }` ← **beda sendiri** |
- **Saran fix backend (pilih salah satu, biar seragam):**
  - **Opsi A (disarankan):** ganti key jadi `ads` → `{ "data": { "ads": [...], "pagination": {...} } }` (samain dgn promo list).
  - **Opsi B:** flat → `{ "data": [...], "pagination": {...} }` (sesuai `API_SPEC_ADS_MOBILE.md` yang sekarang).
- **Status app:** sudah di-workaround di mapper (`AdPageMapper` toleran ke `data[]`/`items`/`ads`/flat), jadi app **nggak nunggu** fix ini — tapi bagusnya tetap diseragamkan.

## 🟡 ADS · pesan error pakai kode domain, bukan kalimat

- **Endpoint:** `POST /mobile/ads` (create) — contoh saat body kurang lengkap.
- **Aktual:** `422` dengan `errors: "DOMAIN_AD_BODY_TEXT_REQUIRED"` (kode, bukan kalimat) dan `message: null`.
  ```json
  { "status":"error","status_code":422,"error_code":"ERR_UNPROCESSABLE_ENTITY","errors":"DOMAIN_AD_BODY_TEXT_REQUIRED" }
  ```
- **Dampak app:** user lihat string mentah `DOMAIN_AD_BODY_TEXT_REQUIRED`. Mohon isi `message` dengan kalimat human-readable (idealnya ter-lokalisasi via `Accept-Language`).
- **Catatan:** `text_ad` ternyata **wajib** `body_text` (di spec opsional) — mohon disinkronkan dengan `API_SPEC_ADS_MOBILE.md`.

> Endpoint ads yang sehat saat dites: `GET /home` ✅, `GET /` (promo) ✅, `GET /my` ✅ (200, walau envelope beda), `POST /impression` ✅ (200 `ok`), `POST /click` ✅ (200 `ok`). `GET /:id` → `404 DOMAIN_AD_NOT_FOUND` saat id bukan milik member (kemungkinan detail owner-scoped — belum bisa dites karena member belum punya ads).

---

## Ringkasan

| Module | Error | Method | Endpoint | Status |
| --- | --- | --- | --- | --- |
| EVENT | Featured | GET | `/mobile/events/featured` | 🔴 500 |
| EVENT | Upcoming | GET | `/mobile/events/upcoming` | 🔴 500 |
| EVENT | Search | GET | `/mobile/events/search` | 🔴 500 |
| NEWS | Like/Unlike (500) | POST/DELETE | `/mobile/news/articles/{id}/like` | ✅ fixed (🟠 `is_liked` selalu false) |
| NEWS | Get Comments (500) | GET | `/mobile/news/articles/{id}/comments` | ✅ fixed (🟠 `created_at` epoch-zero) |
| ADS | Envelope `/my` nested `data.data` | GET | `/mobile/ads/my` | 🟠 inkonsisten |
| ADS | Pesan error pakai kode domain | POST | `/mobile/ads` | 🟡 UX |

**Total: 5 endpoint `500`** + **2 catatan ADS** (1 inkonsistensi envelope, 1 pesan error). Yang `500` wajib diperbaiki backend; yang ADS sebaiknya diseragamkan walau app sudah workaround.

> Detail test lengkap per-module ada di `API_ENDPOINT_TEST.md`. Khusus News dibahas lebih dalam di `NEWS_LIKE_COMMENT_BUG.md`.
