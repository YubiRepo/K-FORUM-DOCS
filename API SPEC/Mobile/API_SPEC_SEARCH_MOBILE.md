# API Spec — Global Search (Mobile)

Pencarian **lintas-modul** untuk search bar utama di home. Satu query mencari ke semua
konten yang relevan (announcement, news, community, post, event, directory/merchant,
Q&A) dan mengembalikan hasil ternormalisasi yang siap dirender + dibuka via deeplink.

- **Base URL Prefix**: `/api/v1/mobile`
- **Status implementasi**: 🔴 **belum ada di backend** — kontrak ini usulan mobile. Saat ini
  search bar home belum terhubung; tiap modul punya search sendiri (`/qna/search`,
  `/announcements/search`, `/events?search=`, dst.) tapi tidak ada endpoint terpadu.

---

## Headers

```
Accept: application/json
Accept-Language: <lang_code>     (ko | id | en — default: ko)
X-Locale: <lang_code>
Authorization: Bearer <access_token>
```

## Response Envelope

Mengikuti envelope standar app:

```json
{
  "status": "success | error",
  "status_code": 200,
  "message": "search results",
  "data": { ... },
  "error_code": null,
  "errors": null,
  "meta": null
}
```

## Tipe Konten (`type`)

| `type`         | Sumber                  | Deeplink dibuka ke           |
| -------------- | ----------------------- | ---------------------------- |
| `announcement` | Pengumuman              | `/announcements/{id}`        |
| `news`         | Artikel berita          | `/news/{id}`                 |
| `community`    | Komunitas               | `/communities/{id}`          |
| `post`         | Postingan komunitas     | `/posts/{id}`                |
| `event`        | Event                   | `/events/{id}`               |
| `merchant`     | Direktori (merchant)    | `/directory/merchant/{id}`   |
| `qna`          | Tanya Jawab / FAQ       | `/qna`                       |

> Set tipe bisa diperluas. `deeplink` cocok 1:1 dengan route go_router app (lihat
> `docs/modules/contracts/DEEPLINK_CONTRACT.md`) sehingga mobile cukup `push(deeplink)`.

## Bentuk Item Hasil (`SearchResultItem`)

Semua tipe dinormalisasi ke bentuk yang sama agar mobile bisa merender satu komponen
list generik:

```json
{
  "type": "news",
  "id": "763bb183-5f5b-4760-9e86-8da872e7a982",
  "title": "Hearts2Hearts kembali dengan single musim panas",
  "subtitle": "Korean Association Indonesia · 1j lalu",
  "thumbnail_url": "https://cdn.kai.id/news/abc.jpg",
  "badge": "NEWS",
  "deeplink": "/news/763bb183-5f5b-4760-9e86-8da872e7a982",
  "highlight": "…rilis single <em>musim panas</em> mereka…",
  "score": 0.92,
  "created_at": "2026-06-23T09:00:00.000Z"
}
```

| Field           | Type             | Keterangan                                                                 |
| --------------- | ---------------- | -------------------------------------------------------------------------- |
| `type`          | string (enum)    | Tipe konten — lihat tabel di atas.                                         |
| `id`            | string (uuid)    | ID entitas.                                                                |
| `title`         | string           | Judul utama (sudah terlokalisasi sesuai `Accept-Language`).                |
| `subtitle`      | string \| null   | Baris pendukung (kategori / author / tanggal / lokasi — bebas per tipe).   |
| `thumbnail_url` | string \| null   | Gambar/icon kecil. `null` → mobile pakai ikon default per tipe.            |
| `badge`         | string \| null   | Label kecil opsional (mis. `CRITICAL`, `Pro`, nama kategori).              |
| `deeplink`      | string           | Path in-app untuk dibuka saat di-tap (route go_router).                   |
| `highlight`     | string \| null   | Cuplikan teks yang match, `<em>` membungkus term. Boleh kosong.           |
| `score`         | number           | Skor relevansi `0..1` (untuk sorting/debug).                              |
| `created_at`    | string (ISO8601) | Untuk tiebreak by-recency & ditampilkan bila perlu.                       |

---

## 1. GET /search — Overview (grouped)

Mode **default**: mengembalikan hasil **dikelompokkan per tipe**, masing-masing berisi
beberapa item teratas + total per tipe. Dipakai untuk **layar hasil search utama**
("Top results") sebelum user memilih "Lihat semua di <tipe>".

**Method**: GET
**URL**: `/api/v1/mobile/search`
**Auth**: Required (member)

**Query Params**

| Param      | Type   | Required | Default | Keterangan                                                                 |
| ---------- | ------ | -------- | ------- | -------------------------------------------------------------------------- |
| `q`        | string | **Yes**  | —       | Keyword. **Min 2 karakter** (setelah trim). `<2` → `400`.                  |
| `types`    | string | No       | (semua) | Filter daftar tipe, dipisah koma. Mis. `news,event,community`.             |
| `per_type` | int    | No       | `5`     | Jumlah item per tipe di overview. Max: `10`.                               |
| `region_id`| string | No       | —       | Batasi hasil ber-region (announcement/community/event/merchant) ke region. |

**Response 200**

```json
{
  "data": {
    "query": "korea",
    "total": 27,
    "groups": [
      {
        "type": "news",
        "total": 12,
        "has_more": true,
        "items": [ /* up to per_type SearchResultItem */ ]
      },
      {
        "type": "event",
        "total": 3,
        "has_more": false,
        "items": [ /* ... */ ]
      },
      {
        "type": "community",
        "total": 8,
        "has_more": true,
        "items": [ /* ... */ ]
      },
      {
        "type": "merchant",
        "total": 4,
        "has_more": false,
        "items": [ /* ... */ ]
      }
    ]
  }
}
```

| Field            | Type   | Keterangan                                                              |
| ---------------- | ------ | ---------------------------------------------------------------------- |
| `query`          | string | Echo query (sudah ter-trim) — untuk verifikasi di client.              |
| `total`          | int    | Total hasil di semua tipe.                                             |
| `groups[]`       | array  | Hanya tipe yang **punya hasil** (tipe tanpa hasil tidak dikirim).      |
| `groups[].type`  | string | Tipe konten.                                                           |
| `groups[].total` | int    | Total hasil untuk tipe ini (bisa > jumlah `items`).                    |
| `groups[].has_more` | bool | `true` bila `total > per_type` → mobile tampilkan tombol "Lihat semua". |
| `groups[].items` | array  | `SearchResultItem[]` (maks `per_type`).                                |

> **Urutan group**: by `total` desc, atau urutan tetap (announcement → news → community
> → event → merchant → qna). Backend bebas pilih; mobile tidak mengasumsikan urutan.

---

## 2. GET /search — Per-type (flat, paginated)

Bila `type` (tunggal) diisi, endpoint mengembalikan **list datar paginated** untuk satu
tipe. Dipakai saat user tap **"Lihat semua di <tipe>"** dan untuk infinite scroll.

**Method**: GET
**URL**: `/api/v1/mobile/search?type=news`
**Auth**: Required (member)

**Query Params**

| Param       | Type   | Required | Default | Keterangan                                                  |
| ----------- | ------ | -------- | ------- | ----------------------------------------------------------- |
| `q`         | string | **Yes**  | —       | Keyword (min 2 karakter).                                   |
| `type`      | string | **Yes**  | —       | **Satu** tipe konten. Wajib di mode ini.                    |
| `limit`     | int    | No       | `20`    | Max: `50`.                                                  |
| `offset`    | int    | No       | `0`     | —                                                           |
| `region_id` | string | No       | —       | Filter region (bila relevan untuk tipe).                    |

**Response 200**

```json
{
  "data": [ /* SearchResultItem[] */ ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 12,
    "has_next": false,
    "has_prev": false
  }
}
```

> Bentuk `pagination` mengikuti list offset-based yang sudah dipakai modul lain
> (mis. `/communities`). Mobile sudah toleran ke `{ limit, offset, total, has_next, has_prev }`.

---

## 3. GET /search/suggestions — Autocomplete (opsional)

Saran cepat saat user mengetik (judul ringkas, tanpa body) untuk dropdown autocomplete.
Ringan & cepat; **opsional** — bila tidak diimplementasikan, mobile cukup pakai `/search`
overview dengan debounce.

**Method**: GET
**URL**: `/api/v1/mobile/search/suggestions`
**Auth**: Required (member)

**Query Params**

| Param   | Type   | Required | Default | Keterangan                          |
| ------- | ------ | -------- | ------- | ----------------------------------- |
| `q`     | string | **Yes**  | —       | Keyword (min 2 karakter).           |
| `limit` | int    | No       | `8`     | Max: `10`. Total saran lintas tipe. |

**Response 200**

```json
{
  "data": [
    { "type": "news", "id": "…", "title": "Korean Cultural Festival 2026", "deeplink": "/news/…" },
    { "type": "event", "id": "…", "title": "Korea Food Festival", "deeplink": "/events/…" },
    { "type": "community", "id": "…", "title": "Korea Lovers Jakarta", "deeplink": "/communities/…" }
  ]
}
```

---

## Aturan Visibilitas & Keamanan

Hasil **wajib** difilter di backend mengikuti hak akses & status konten:

| Aturan                                                                                       |
| -------------------------------------------------------------------------------------------- |
| Hanya konten **aktif/terbit**: announcement published, news published, event tidak `draft`, merchant `published`, ads tidak masuk search. |
| **Komunitas private**: metadata komunitas boleh muncul, tapi **post-nya tidak** boleh muncul untuk non-anggota. |
| **Post komunitas** hanya muncul bila user **anggota** komunitas tersebut.                    |
| Locale: `title`/`subtitle`/`highlight` mengikuti `Accept-Language`; fallback ke bahasa default konten bila terjemahan tidak ada. |
| `region_id`: bila diisi, batasi tipe ber-region; konten global tetap boleh muncul.          |

## Ranking

- Urutkan per `score` (relevansi) desc, lalu `created_at` desc sebagai tiebreak.
- Match di `title` diberi bobot lebih tinggi dari body/excerpt.
- Pengumuman `CRITICAL`/`urgent` boleh di-boost ke atas (opsional).

## Errors

```json
// 400 — query terlalu pendek / kosong
{ "status": "error", "status_code": 400, "error_code": "ERR_VALIDATION",
  "errors": "Query minimal 2 karakter." }

// 401 — belum login
{ "status": "error", "status_code": 401, "error_code": "ERR_UNAUTHORIZED",
  "errors": "Authentication required." }

// 422 — type tidak dikenal
{ "status": "error", "status_code": 422, "error_code": "ERR_UNPROCESSABLE_ENTITY",
  "errors": "Unknown search type: foo" }
```

> Bila `q` valid tapi tak ada hasil → **`200`** dengan `groups: []` (overview) atau
> `data: []` (per-type). Bukan error.

## Catatan untuk Mobile

1. **Debounce** input minimal **300ms** sebelum hit `/search`; jangan kirim saat `q < 2`.
2. **Batalkan** request sebelumnya saat user mengetik lagi (hindari hasil out-of-order).
3. Tap item → `context.push(item.deeplink)` (route go_router, tanpa parsing tambahan).
4. Render generik: thumbnail → `thumbnail_url` (fallback ikon per `type`), judul `title`,
   sub `subtitle`, badge `badge`. `highlight` boleh dirender sebagai teks rich (`<em>`).
5. Overview dulu (grouped) → tombol "Lihat semua di <tipe>" buka mode per-type (paginated).

## Catatan untuk Backend

- Endpoint terpadu bisa fan-out ke index tiap modul (atau satu index pencarian terpusat
  spt Postgres FTS / Meilisearch / Elasticsearch) lalu normalisasi ke `SearchResultItem`.
- Hindari N+1: ambil thumbnail/subtitle dalam query yang sama per tipe.
- Pertimbangkan **rate limit** ringan (mis. 10 req / 10 detik / user) karena dipanggil saat mengetik.
