# Event Module — Error Report

Endpoint Event yang **error** hasil test mobile.

- **Base URL:** `https://k-forum-api.yubicom.co.id/api/v1`
- **Tanggal:** 23 Juni 2026
- **Auth:** login member biasa (token valid)
- **Sifat:** semua error = **`500 ERR_INTERNAL`** (unhandled exception backend), bukan validasi/permission.

> Pola: endpoint **list & detail jalan normal**, tapi handler "khusus" (featured/upcoming/search) melempar 500. Kemungkinan bug query/join/agregasi. Mohon cek **log server** tiap handler.

---

## 🔴 EVENT · Featured
- **Endpoint:** `GET /mobile/events/featured?limit=5`
- **Harusnya:** `200` + daftar event unggulan
- **Aktual:** `500` `ERR_INTERNAL`
- **Repro:**
  ```bash
  curl -H "Authorization: Bearer <token>" \
    "https://k-forum-api.yubicom.co.id/api/v1/mobile/events/featured?limit=5"
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
- **Dampak app:** layar Search event gagal.
- **Catatan:** `GET /mobile/events?search=korea` (list + filter) **JALAN** → bisa jadi acuan untuk benerin handler `/search`.

---

## Ringkasan

| Error | Method | Endpoint | Status |
| --- | --- | --- | --- |
| Featured | GET | `/mobile/events/featured` | 🔴 500 |
| Upcoming | GET | `/mobile/events/upcoming` | 🔴 500 |
| Search | GET | `/mobile/events/search` | 🔴 500 |

**Total: 3 endpoint** error `500` di module Event. Endpoint Event lain (list, detail, save, schedule, share) **sehat**.

> Test lengkap semua module: `API_ENDPOINT_TEST.md`. Rekap error semua module: `API_ERROR_REPORT.md`.
