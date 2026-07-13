# Issue: News API — filter `scope` & status `is_bookmarked` (backend)

- **Modul**: News — `GET /api/v1/mobile/news/articles` (+ `POST|DELETE .../bookmark`)
- **Severity**: High — fitur utama layar News (portal section per scope) bergantung pada filter ini
- **Status**: 🔴 Open — menunggu fix backend (2 issue di dokumen ini)
- **Ditemukan**: 13 Jul 2026, diverifikasi live di production (`k-forum-api.yubicom.co.id`)
- **Pelapor**: Mobile team (via QA visual — artikel yang sama muncul di section scope berbeda; ikon bookmark tidak menyala padahal sudah di-bookmark)

---

## Ringkasan

Layar News versi baru menampilkan berita **per scope** (Indonesia / Korea / Korea Indonesia Merdeka) sebagai section terpisah. Saat implementasi, filter scope di endpoint list artikel ternyata **tidak berfungsi sama sekali** — dua-duanya:

| Param | Perilaku seharusnya | Perilaku aktual |
|---|---|---|
| `scope` (slug, sesuai spec) | Filter by slug | **HTTP 500** setiap kali param dikirim |
| `scope_id` (uuid) | Filter by id | **Diabaikan diam-diam** — response identik dengan tanpa filter |

## Bukti reproduksi (curl, prod, 13 Jul 2026)

```bash
# Login → ambil token (akun QA test_member)
# Scope id dari GET /mobile/news/scopes:
#   Indonesia : 3375602e-fa3e-4f04-a80d-67a10601f5bd
#   Korea     : dd091caf-2c27-4191-af4d-ac8b814fce52
#   KIM       : 46368fa5-6e96-4edf-a611-7770c558f288

GET /mobile/news/articles?limit=20&scope_id=<Indonesia>   # → 20 artikel
GET /mobile/news/articles?limit=20&scope_id=<Korea>       # → 20 artikel YANG SAMA
GET /mobile/news/articles?limit=20                        # → 20 artikel YANG SAMA
```

Hasil: **ID set ketiga response identik**. Termasuk artikel dengan `scope_name: "Indonesia"` (mis. artikel IHSG/OJK) yang tetap muncul saat difilter `scope_id=Korea`.

Temuan tambahan:
- Dari 20 artikel teratas, **10 artikel `scope_name: null`** — belum ditandai scope sama sekali.
- Response list memakai field flat `scope_id`/`scope_name` (bukan objek `scope` nested seperti detail) — mobile sudah menyesuaikan parsing-nya.

## Kendala yang ditimbulkan

1. **Fungsional**: tanpa filter server, semua section scope menampilkan artikel yang sama → layar portal tidak ada artinya. Ini yang terlihat di QA ("kadang muncul sama").
2. **Workaround mobile boros & tidak akurat** (sudah dipasang sementara di `news_screen.dart`, helper `_onlyScope`):
   - App terpaksa **over-fetch** (ambil 20 untuk menampilkan 4 per section) lalu menyaring sendiri by `article.scope.id` → bandwidth & latency terbuang.
   - **Pagination "Lihat Semua" tidak akurat**: server memberi halaman berisi campuran scope; setelah disaring client, satu halaman bisa berisi sedikit (bahkan 0) artikel meski `has_more` masih true.
   - **Artikel ber-scope null tidak tampil di section mana pun** (hanya via search/featured) — keputusan produk sementara dari mobile, perlu dikonfirmasi.
3. **Tidak bisa dituntaskan dari sisi mobile**: penyaringan sejati harus di query DB; client tidak mungkin menyaring seluruh korpus.

## Yang diminta ke backend

1. **Implement filter `scope_id`** (uuid) di `GET /mobile/news/articles` — WHERE clause di query, bukan post-filter. Pagination (`total_pages`/`has_more`) harus dihitung SETELAH filter.
2. Perbaiki `scope` (slug) yang menyebabkan 500, atau hapus dari spec kalau memang tidak didukung.
3. **Keputusan artikel `scope_name: null`**: (a) wajibkan scope saat publish + backfill data lama, ATAU (b) definisikan perilakunya (mis. hanya muncul saat tanpa filter). Mobile mengikuti.

## Kriteria selesai (acceptance)

- [ ] `?scope_id=<Indonesia>` hanya mengembalikan artikel scope Indonesia; `meta` pagination konsisten.
- [ ] `?scope_id=<uuid-ngawur>` → response kosong / 400, bukan seluruh artikel.
- [ ] `?scope=<slug>` tidak lagi 500.
- [ ] Keputusan artikel null-scope terdokumentasi di `API_SPEC_NEWS_MOBILE.md`.
- [ ] Mobile menghapus workaround `_onlyScope` + over-fetch di `lib/features/news/presentation/screens/news_screen.dart` (ditandai komentar `WORKAROUND (13 Jul 2026)`).

---

# Issue 2: `is_bookmarked` hilang dari response LIST artikel

## Ringkasan

Ikon bookmark di daftar berita selalu tampil "belum di-bookmark" walaupun artikelnya sudah di-bookmark. Penyebab: **endpoint list tidak pernah mengirim `is_bookmarked`**, padahal `is_liked` dikirim.

## Bukti reproduksi (curl, prod, 13 Jul 2026)

```bash
POST /mobile/news/articles/{id}/bookmark      # → 200, tapi body: { "data": null }
GET  /mobile/news/articles?limit=5            # → item TIDAK punya field is_bookmarked
                                              #   (is_liked ADA — inkonsisten)
GET  /mobile/news/articles/{id}               # → detail: is_bookmarked: true ✅
GET  /mobile/news/bookmarks                   # → artikel tsb ada di daftar ✅
```

Jadi bookmark **tersimpan benar** di server; hanya response list yang tidak melaporkannya.

## Kendala yang ditimbulkan

1. Setiap kali list dimuat/di-refresh, semua artikel tampil "belum di-bookmark" — user mengira bookmark-nya hilang, atau malah men-tap lagi dan **tanpa sadar meng-unbookmark**.
2. `POST|DELETE /bookmark` membalas `data: null` (tanpa `is_bookmarked` baru). Mobile kini fallback ke state yang diminta, tapi server-reported state selalu lebih aman (mis. kalau request idempotent/konflik).

## Yang diminta ke backend

1. **Sertakan `is_bookmarked` di setiap item response LIST** `GET /mobile/news/articles` (dan `GET /mobile/news/bookmarks`), konsisten dengan `is_liked`.
2. (Nice to have) balas `POST|DELETE /bookmark` dengan `data: { "article_id": "...", "is_bookmarked": true|false }` seperti kontrak di spec — mobile sudah siap membacanya.

## Kriteria selesai (acceptance)

- [ ] Item list membawa `is_bookmarked` yang akurat per user.
- [ ] Ikon bookmark di list app menyala untuk artikel yang sudah di-bookmark, dan tetap menyala setelah pull-to-refresh.

## Catatan mobile (sudah dikerjakan, 13 Jul 2026)

- Bug ikon "balik sendiri sedetik setelah tap" sudah diperbaiki di app: mapper kini mentolerir `data: null` dan screen mempertahankan state yang diminta (`news_results.dart`, `news_models.dart`, `news_screen.dart`, `news_detail_screen.dart`).
- Yang TIDAK bisa diperbaiki dari app: status bookmark setelah reload list — butuh field dari server (issue di atas).

---

## Referensi

- Spec: `docs/api_spec/API_SPEC_NEWS_MOBILE.md` (bagian query parameters — ada callout warning yang menunjuk ke issue ini)
- Workaround mobile (scope): `lib/features/news/presentation/screens/news_screen.dart` → `_onlyScope`, `_loadSections`, `_load`, `_loadMore`
