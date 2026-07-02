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

Kurs mata uang untuk widget Home (default: **1 KRW → IDR**).

**Authentication**: Optional

**Method**: GET

**URL**: `/api/v1/mobile/home/exchange-rates`

**Query Parameters**:

| Parameter | Type   | Required | Default | Description |
|-----------|--------|----------|---------|-------------|
| `base`    | string | No       | `KRW`   | Mata uang sumber (ISO 4217) |
| `quote`   | string | No       | `IDR`   | Mata uang tujuan (ISO 4217) |

**Response (200 OK)**:

```json
{
  "status": "success",
  "message": "OK",
  "data": {
    "base_currency": "KRW",
    "quote_currency": "IDR",
    "rate": 11801.90,
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
| `change_percent`  | number | Yes      | Perubahan persentase vs periode referensi (biasanya 24 jam) |
| `updated_at`      | string | No       | ISO-8601 UTC — waktu rate di-fetch/cache |

**Contoh interpretasi**:

- `rate: 11801.90` dengan `base: KRW`, `quote: IDR` → **1 KRW = 11.801,90 IDR**
- `change_percent: -0.71` → turun 0,71% dibanding referensi

**Error Responses**:

| HTTP | Kondisi |
|------|---------|
| 404  | Endpoint belum diimplementasi — mobile pakai fallback |
| 400  | Pair mata uang tidak didukung |
| 502  | Upstream exchange provider gagal |

---

## Mobile Fallback Behaviour

| Skenario | Perilaku app |
|----------|--------------|
| HTTP 404 | Tampilkan data fallback statis |
| Network error / timeout | Tampilkan data fallback statis |
| HTTP 5xx | Tampilkan data fallback statis |
| Response `data` invalid | Tampilkan data fallback statis |
| HTTP 200 + payload valid | Tampilkan data live dari API |

**Fallback weather (saat ini)**:

```json
{
  "temperature": 26,
  "min_temperature": 22,
  "max_temperature": 30,
  "condition": "sunny"
}
```

**Fallback exchange rate (saat ini)**:

```json
{
  "base_currency": "KRW",
  "quote_currency": "IDR",
  "rate": 11801.90,
  "change_percent": -0.71
}
```

---

## Rekomendasi Implementasi Backend

1. **Cache** hasil upstream (weather + FX) minimal 15–30 menit untuk hemat quota API pihak ketiga.
2. **Jangan expose API key** third-party ke mobile — semua proxy lewat backend.
3. **Normalisasi `condition`** ke enum yang disepakati di tabel di atas.
4. **`updated_at`** wajib diisi agar mobile bisa menampilkan indikator freshness di masa depan.
5. Endpoint boleh **public** (tanpa auth) karena data non-personal; opsional auth untuk personalisasi region.

### Contoh Upstream (referensi internal)

- Weather: OpenWeatherMap, WeatherAPI, atau BMKG (jika tersedia)
- Exchange: Bank of Korea, exchangerate.host, atau sumber resmi KAI

---

## Changelog

| Version | Date       | Notes |
|---------|------------|-------|
| 1.0.0   | 2026-07-01 | Initial spec — weather + exchange rate widgets |