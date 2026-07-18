# API Spec — Home Insights (Weather & Exchange Rate)

Dokumentasi API untuk widget **Cuaca** dan **Kurs** di halaman Home mobile.

Mobile client sudah mengonsumsi endpoint ini. Jika endpoint belum tersedia (404) atau error, app menampilkan **data fallback statis** tanpa mengganggu UX.

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/mobile/home`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`j
  - `Authorization: Bearer <access_token>` (optional — endpoint publik)
  - `Accept-Language: <lang_code>` (default: `id`)
  - `X-Locale: <lang_code>`

### Response Envelope

Semua response mengikuti envelope standar:

```json
{
  "status": "success",
  "message": "OK",
  "data": { }
}
```

Error:

```json
{
  "status": "error",
  "message": "Human readable message",
  "errors": {}
}
```

---

## 1. GET /weather

Cuaca saat ini untuk lokasi pengguna.

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/home/weather`

**Query Parameters**:

| Parameter | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `lat`     | number | No       | Latitude dari GPS device |
| `lng`     | number | No       | Longitude dari GPS device |

Jika `lat`/`lng` tidak dikirim, backend boleh fallback ke region default user (jika auth) atau lokasi default aplikasi.

**Response (200 OK)**:

```json
{
  "status": "success",
  "message": "OK",
  "data": {
    "temperature": 26,
    "min_temperature": 22,
    "max_temperature": 30,
    "condition": "sunny",
    "location_name": "Gangnam",
    "updated_at": "2026-07-01T08:00:00.000Z"
  }
}
```

**Field `data`**:

| Field             | Type   | Required | Description |
|-------------------|--------|----------|-------------|
| `temperature`     | number | Yes      | Suhu saat ini (°C) |
| `min_temperature` | number | No       | Suhu minimum hari ini (°C) |
| `max_temperature` | number | No       | Suhu maksimum hari ini (°C) |
| `condition`       | string | Yes      | Kondisi cuaca — lihat tabel di bawah |
| `location_name`   | string | No       | Nama lokasi tampilan (kota/kecamatan) |
| `updated_at`      | string | No       | ISO-8601 UTC — waktu data di-fetch/cache |

**Nilai `condition` yang didukung mobile**:

| Value            | Icon (mobile)   |
|------------------|-----------------|
| `sunny`          | Matahari        |
| `clear`          | Matahari        |
| `partly_cloudy`  | Awan + matahari |
| `cloudy`         | Awan            |
| `overcast`       | Awan            |
| `rain` / `rainy` | Hujan           |
| `drizzle`        | Hujan           |
| `thunderstorm`   | Petir           |
| `storm`          | Petir           |
| `snow` / `snowy` | Salju           |
| `fog` / `mist`   | Kabut           |

**Error Responses**:

| HTTP | Kondisi |
|------|---------|
| 404  | Endpoint belum diimplementasi — mobile pakai fallback |
| 502  | Upstream weather provider gagal |
| 503  | Layanan sementara tidak tersedia |

---

## 2. GET /exchange-rates

Kurs mata uang untuk widget Home. Mobile menampilkan **beberapa pair sekaligus** (per 2026-07-18): **KRW→IDR**, **USD→IDR**, **USD→KRW**. Endpoint ini melayani **satu pair per request**; mobile memanggilnya sekali per pair (paralel).

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/home/exchange-rates`

**Query Parameters**:

| Parameter | Type   | Required | Default | Description |
|-----------|--------|----------|---------|-------------|
| `base`    | string | No       | `KRW`   | Mata uang sumber (ISO 4217) |
| `quote`   | string | No       | `IDR`   | Mata uang tujuan (ISO 4217) |

Pair yang dipakai mobile saat ini: `KRW/IDR`, `USD/IDR`, `USD/KRW`.

**Response (200 OK)**:

```json
{
  "status": "success",
  "message": "OK",
  "data": {
    "base_currency": "KRW",
    "quote_currency": "IDR",
    "rate": 11.94,
    "change_percent": -0.71,
    "updated_at": "2026-07-01T08:00:00.000Z"
  }
}
```

**Field `data`**:

| Field             | Type   | Required | Description |
|-------------------|--------|----------|-------------|
| `base_currency`   | string | Yes      | ISO 4217 — mata uang sumber |
| `quote_currency`  | string | Yes      | ISO 4217 — mata uang tujuan |
| `rate`            | number | Yes      | Nilai tukar: **1 unit `base` = `rate` unit `quote`** |
| `change_percent`  | number | Yes      | Perubahan persentase vs periode referensi (biasanya 24 jam). Boleh `0` jika sumber tidak menyediakan tren |
| `updated_at`      | string | No       | ISO-8601 UTC — waktu rate di-fetch/cache |

**Contoh interpretasi** (nilai **per 1 unit base**, jangan di-scale ×1000):

- `KRW→IDR`, `rate: 11.94` → **1 KRW = 11,94 IDR**
- `USD→IDR`, `rate: 16300` → **1 USD = 16.300 IDR**
- `USD→KRW`, `rate: 1365` → **1 USD = 1.365 KRW**
- `change_percent: -0.71` → turun 0,71% dibanding referensi

> **Catatan (perbaikan 2026-07-18):** contoh lama `rate: 11801.90` untuk `KRW/IDR` salah skala ~1000× (nilai riil ±11–12 IDR per 1 KRW). Kirim rate **per 1 unit base**.

> **Rekomendasi (opsional, belum wajib):** sediakan juga endpoint batch
> `GET /exchange-rates/batch?pairs=KRW_IDR,USD_IDR,USD_KRW` yang mengembalikan
> `data: [ {…}, {…}, {…} ]` supaya mobile cukup 1 request. Sampai tersedia,
> mobile tetap memanggil per-pair.

**Error Responses**:

| HTTP | Kondisi |
|------|---------|
| 404  | Endpoint belum diimplementasi — mobile pakai fallback |
| 400  | Pair mata uang tidak didukung |
| 502  | Upstream exchange provider gagal |

---

## Mobile Fallback Behaviour

**Weather** — 1 tingkat fallback (statis) saat backend gagal.

**Exchange rate** — **3 tingkat**, per pair, independen:

1. **Backend** `/mobile/home/exchange-rates` (sumber utama).
2. **Public FX API** (client-side) jika backend gagal — ExchangeRate-API open
   endpoint `https://open.er-api.com/v6/latest/USD` (**key-less**, USD-based).
   Mobile menghitung cross-rate: `base→quote = usdRate[quote] / usdRate[base]`.
   `change_percent` = `0` (sumber ini tidak menyediakan tren). Dipanggil pakai
   Dio terpisah tanpa interceptor auth/locale aplikasi.
3. **Static** (terakhir) jika kedua sumber di atas gagal — tabel angka kasar
   agar widget tetap render offline.

| Skenario (per pair) | Perilaku app |
|---------------------|--------------|
| Backend 200 + valid | Live dari backend |
| Backend 404 / 5xx / timeout / invalid | Coba public FX API |
| Public FX API sukses | Live dari FX API (`change_percent: 0`, `is_fallback` internal) |
| Public FX API juga gagal | Static fallback |

**Fallback weather (statis)**:

```json
{ "temperature": 26, "min_temperature": 22, "max_temperature": 30, "condition": "sunny" }
```

**Static exchange-rate table (last resort, per 1 unit base)**:

```json
[
  { "base_currency": "KRW", "quote_currency": "IDR", "rate": 11.94,   "change_percent": 0 },
  { "base_currency": "USD", "quote_currency": "IDR", "rate": 16300.0, "change_percent": 0 },
  { "base_currency": "USD", "quote_currency": "KRW", "rate": 1365.0,  "change_percent": 0 }
]
```

> Catatan: begitu backend melayani semua pair dengan `change_percent` benar,
> tingkat 2 & 3 jarang terpakai. Public FX API hanya penyelamat sementara.

---

## Rekomendasi Implementasi Backend

1. **Cache** hasil upstream (weather + FX) minimal 15–30 menit untuk hemat quota API pihak ketiga.
2. **Jangan expose API key** third-party ke mobile — semua proxy lewat backend.
3. **Normalisasi `condition`** ke enum yang disepakati di tabel di atas.
4. **`updated_at`** wajib diisi agar mobile bisa menampilkan indikator freshness di masa depan.
5. Endpoint boleh **public** (tanpa auth) karena data non-personal; opsional auth untuk personalisasi region.

### Contoh Upstream (referensi internal)

- Weather: OpenWeatherMap, WeatherAPI, atau BMKG (jika tersedia)
- Exchange — sumber **wajib punya data historical/timeseries** agar bisa
  menghitung `change_percent` (tren ±24 jam). Sumber spot-only (harian, tanpa
  history) hanya cukup untuk `rate`, `change_percent` terpaksa `0`.

  | Sumber | API key | Historical (untuk `change_percent`) | KRW & IDR | Catatan |
  |--------|---------|-------------------------------------|-----------|---------|
  | **Frankfurter** (`api.frankfurter.dev`) | ❌ key-less | ✅ endpoint per-tanggal & range | ✅ (data ECB) | **Rekomendasi utama** — gratis, tanpa key, base bisa di-switch. `change_percent = (rate_today − rate_yesterday) / rate_yesterday × 100`. |
  | **exchangerate.host** | ✅ free key | ✅ `/timeseries` | ✅ luas | Alternatif jika butuh lebih banyak currency. |
  | **Bank of Korea ECOS** (한국은행) | ✅ free key | ✅ seri harian resmi | KRW otoritatif, IDR via cross | Paling akurat untuk KRW; cocok karena app fokus Korea. |
  | **Open Exchange Rates / CurrencyAPI / Fixer** | ✅ freemium | ✅ | ✅ | Pilih bila butuh SLA/volume tinggi. |
  | `open.er-api.com` | ❌ key-less | ❌ (spot harian, tanpa tren) | ✅ (USD-based) | Sama dengan fallback client-side mobile — **jangan sekadar di-proxy**, tidak menambah nilai (`change_percent` selalu `0`). |

---

## Changelog

| Version | Date       | Notes |
|---------|------------|-------|
| 1.1.1   | 2026-07-18 | Perluas rekomendasi upstream exchange: tabel sumber + syarat historical/timeseries untuk `change_percent`; tandai `open.er-api.com` (spot-only) jangan sekadar di-proxy; tambah Frankfurter sebagai rekomendasi utama key-less |
| 1.1.0   | 2026-07-18 | Multi-pair (KRW/IDR, USD/IDR, USD/KRW); fallback 3-tingkat exchange rate (backend → public FX API → static); perbaikan skala contoh `KRW/IDR` (per 1 unit base); rekomendasi endpoint batch |
| 1.0.0   | 2026-07-01 | Initial spec — weather + exchange rate widgets |