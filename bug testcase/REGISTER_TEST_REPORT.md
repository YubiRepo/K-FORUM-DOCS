# Laporan Test — Register (`POST /mobile/auth/register`)

- **Tanggal:** 18 Juni 2026 (re-test)
- **Endpoint:** `POST {baseUrl}/mobile/auth/register`
- **Environment:** dev — `http://192.168.1.29:8888/api/v1`
- **Data dikirim:** `{ username, fullname?, email, phone?, password }`
- **Cara test:** manual via curl (40+ skenario, locale en/id/ko)

---

## Kesimpulan singkat

Endpoint **berfungsi**. Validasi input jalan, cek data ganda (unik) jalan, dan
**tidak ada celah naik-hak-akses** (aman). **Kabar baik vs test sebelumnya:** pesan
error sekarang **terlokalisasi penuh** (en/id/ko) dan tiap response punya `error_code`.
Masih ada beberapa **ketidakkonsistenan** yang sebaiknya dirapikan.

> Catatan: bisa login walau email belum diverifikasi itu **memang disengaja**
> (lihat bagian "Catatan"), jadi bukan bug.

---

## Bentuk response (envelope)

Semua response pakai bentuk seragam:

```json
// sukses
{ "status": "success", "status_code": 201, "message": "register success",
  "data": { "user_id": "..." }, "meta": null }

// error
{ "status": "error", "status_code": 422, "error_code": "ERR_UNPROCESSABLE_ENTITY",
  "errors": { "email": "This field is required." } }
```

`error_code` yang ditemui: `ERR_UNPROCESSABLE_ENTITY` (422), `ERR_CONFLICT` (409),
`ERR_UNAUTHORIZED` (401).

> **Bentuk `errors` belum konsisten** (lihat B-6): kadang **objek per-field**
> `{"<field>": "<pesan>"}`, kadang **string polos** `"<pesan>"`.

---

## Aturan yang sudah dipastikan

| Field | Aturan |
| --- | --- |
| email | wajib, format harus valid + ada domain (mis. `.com`), **unik** (huruf besar/kecil dianggap sama), tidak di-trim |
| username | wajib, **min 3 karakter**, hanya huruf/angka/underscore (tidak boleh spasi, titik, `@`), **unik tapi huruf besar/kecil dianggap beda** |
| phone | ada batas panjang min & max, boleh diawali `+`, **unik**; wajib kalau dikirim — **tapi kalau field-nya tidak dikirim sama sekali malah lolos** (bug B-2) |
| password | wajib, **min 8 karakter**, **wajib ada huruf kapital**, **wajib ada angka** — tapi **tidak wajib huruf kecil** (lihat B-8) |
| fullname | opsional; **disimpan apa adanya (tidak dibersihkan)** |
| Sukses | `201` + `{ "data": { "user_id": "..." } }` |

---

## Hasil pengujian

> **Catatan kolom *Pesan*:** ditulis sesuai response server **tanpa diterjemahkan**;
> yang ditampilkan adalah versi **locale `en`**. Pesan ini **terlokalisasi** — versi
> `id`/`ko` ada (lihat bagian "Catatan — pesan error tergantung locale"). Acuan utama
> app tetap **status code + field key**, bukan teks pesan.

### 1. Cek data ganda (unik)
| Skenario | Status | Pesan (`en`) |
| --- | --- | --- |
| Semua data baru | **201** ✅ | `register success` |
| Email sama (username & phone baru) | **409** ✅ | `"A data conflict occurred."` |
| Username sama (email & phone baru) | **409** ✅ | `"A data conflict occurred."` |
| Phone sama (email & username baru) | **409** ✅ | `"A data conflict occurred."` |
| Email sama tapi HURUF BESAR (`QA.X@` vs `qa.x@`) | **409** ✅ | email tidak bedakan besar/kecil |
| Username beda besar/kecil (`QaCase` vs `qacase`) | **201 + 201** ⚠️ | keduanya lolos (lihat B-3) |

Isi `409`: string polos `"A data conflict occurred."` — **tidak menyebut field mana** yang bentrok.

### 2. Field wajib / kosong
| Skenario | Status | `errors` (`en`) |
| --- | --- | --- |
| Email tidak dikirim | 422 | `{"email": "This field is required."}` |
| Username tidak dikirim | 422 | `{"username": "This field is required."}` |
| Password tidak dikirim | 422 | `{"password": "This field is required."}` |
| **Phone tidak dikirim** | **201** ⚠️ | akun dibuat tanpa phone (lihat B-2) |
| Email diisi `""` (kosong) | 422 | `{"email": "This field is required."}` |
| Phone diisi `""` (kosong) | 422 | `"Invalid phone number format."` |

### 3. Batas nilai & password policy
| Skenario | Status | `errors` (`en`) |
| --- | --- | --- |
| Username 3 karakter | **201** ✅ | — |
| Username 2 karakter | 422 | `{"username": "Minimum length is 3 characters."}` |
| Password valid (`Password123`) | **201** ✅ | — |
| Password 7 karakter (`Pass12`) | 422 | `{"password": "Minimum length is 8 characters."}` |
| Password tanpa huruf kapital (`password1`) | 422 | `"Password must contain an uppercase letter."` |
| Password tanpa angka (`Passwordxx`) | 422 | `"Password must contain a digit."` |
| **Password tanpa huruf kecil (`PASSWORD123`)** | **201** ⚠️ | lolos — huruf kecil tidak diwajibkan (B-8) |
| Phone terlalu pendek (`0812`) | 422 | `"Invalid phone number format."` |
| Phone terlalu panjang (21 digit) | 422 | `"Invalid phone number format."` |

### 4. Format / karakter
| Skenario | Status | `errors` (`en`) |
| --- | --- | --- |
| Email salah format (`not-an-email`) | 422 | `{"email": "Invalid email format."}` |
| Email tanpa domain (`a@b`) | 422 | `{"email": "Invalid email format."}` |
| Email ada spasi di depan/belakang | 422 | `{"email": "Invalid email format."}` (tidak di-trim dulu) |
| Phone berisi huruf (`abcde`) | 422 | `"Invalid phone number format."` |
| Phone diawali `+62…` | **201** ✅ | boleh pakai `+` |
| Username pakai spasi / `@` | 422 | `"Invalid username format."` |
| Fullname tidak dikirim | **201** ✅ | memang opsional |

### 5. Uji keamanan
| Skenario | Status | Catatan |
| --- | --- | --- |
| Sisipkan `role`/`roles`/`is_admin`/`plan` di body | **201** | ✅ **Diabaikan.** Login → `/auth/me` = user biasa (`status: ACTIVE`, **tanpa** field role/plan). Tidak bisa naik jadi admin |
| Sisipkan XSS di fullname `<script>alert(1)</script>` | **201** | ⚠️ tersimpan apa adanya, tidak dibersihkan (lihat B-5) |

---

## Daftar masalah (untuk tim backend)

| Kode | Tingkat | Masalah | Saran perbaikan |
| --- | --- | --- | --- |
| B-1 | 🟢 Ringan | **Data ganda — tidak reproduce di re-test ini.** Dup email/username/phone konsisten ditolak `409`. (Di test awal sempat lolos 1×; kemungkinan race condition / daftar ulang akun belum verifikasi.) | Tetap pasang unik di level DB sebagai jaring pengaman. Coba reproduksi: daftar → biarkan belum verifikasi → daftar lagi dgn email sama. |
| B-2 | 🔴 Penting | **Phone tidak konsisten.** Phone diisi `""` → 422, tapi **field phone tidak dikirim sama sekali → lolos 201** (akun tanpa phone). | Anggap "tidak dikirim" = "kosong": kalau wajib → 422; atau jadikan phone benar-benar opsional secara konsisten. |
| B-3 | 🟡 Sedang | **Username besar/kecil dianggap beda**, padahal email tidak. `QaCase` & `qacase` dua-duanya bisa ada → rawan penyamaran. | Cek username unik tanpa bedakan huruf besar/kecil (simpan/bandingkan versi lowercase). |
| B-4 | 🟡 Sedang | **Pesan `409` terlalu umum** (`"A data conflict occurred."`) — tidak menyebut field mana yang bentrok, app tidak bisa kasih pesan spesifik ("email sudah dipakai" vs "username sudah dipakai"). | Sertakan field, mis. `{"errors":{"email":"..."}}` dengan `error_code` spesifik. |
| B-5 | 🟡 Sedang | **Fullname tidak dibersihkan** — `<script>` tersimpan mentah. Aman hanya kalau semua tampilan meng-escape teks. | Bersihkan input, atau pastikan semua tampilan (terutama backoffice) tidak pakai `v-html`. |
| B-6 | 🟡 Sedang | **Bentuk `errors` belum seragam** — kadang objek per-field (`{"email":"..."}` untuk required / min-length / email-format), kadang string polos (`"Password must contain..."`, `"Invalid phone number format."`, `"Invalid username format."`, konflik 409). | Selalu pakai format objek per-field biar app gampang map ke input yang salah. |
| B-7 | 🟢 Ringan | **Email tidak dirapikan dulu** — spasi di depan/belakang langsung dianggap salah format. | Trim email (dan username) sebelum divalidasi. |
| B-8 | 🟡 Sedang | **Password policy tidak konsisten** — wajib huruf kapital & angka, tapi **huruf kecil tidak diwajibkan** (`PASSWORD123` lolos). | Samakan aturan: kalau memang butuh kompleksitas, wajibkan juga huruf kecil; atau dokumentasikan aturan resminya supaya app bisa validasi di sisi klien. |

---

## Bukti ringkas (locale `en`)

```
semua data baru        → 201  {"data":{"user_id":"..."}}  "register success"
email/user/phone sama  → 409  ERR_CONFLICT  "A data conflict occurred."
email sama (HURUF BESAR)→ 409  (email tidak bedakan besar/kecil)
QaCase + qacase        → 201 + 201  (username bedakan besar/kecil)        ← B-3
phone tidak dikirim    → 201  (akun tanpa phone)                          ← B-2
phone "" (kosong)      → 422  "Invalid phone number format."
no email/user/pass     → 422  {"<field>":"This field is required."}
username 3 / 2 char    → 201 / 422 {"username":"Minimum length is 3 characters."}
password Password123   → 201
password Pass12 (7)    → 422  {"password":"Minimum length is 8 characters."}
password password1     → 422  "Password must contain an uppercase letter."
password Passwordxx    → 422  "Password must contain a digit."
password PASSWORD123   → 201  (huruf kecil tidak wajib)                   ← B-8
phone 0812 / 21 digit  → 422  "Invalid phone number format."
phone +62…             → 201  (boleh pakai +)
username spasi / @     → 422  "Invalid username format."
email a@b / "  x  "    → 422  {"email":"Invalid email format."}
sisip role=superadmin  → 201, /auth/me user biasa tanpa role (diabaikan, aman)
fullname <script>      → 201, tersimpan mentah                            ← B-5
```

---

## Catatan — pesan error tergantung locale

API memakai header `Accept-Language` / `X-Locale` (lihat `LocaleInterceptor`).
Di re-test ini dikonfirmasi **pesan error sudah terlokalisasi penuh** untuk `en` / `id` / `ko`
— berlaku untuk error per-field **maupun** string polos, termasuk konflik 409.

Contoh body invalid yang sama, beda locale:

```
[en] {"email":"This field is required.","username":"Minimum length is 3 characters."}
[id] {"email":"Wajib diisi.","username":"Minimal 3 karakter."}
[ko] {"email":"필수 입력 항목입니다.","username":"최소 3자 이상이어야 합니다."}

[en] "Password must contain an uppercase letter."
[id] "Password harus mengandung huruf kapital."
[ko] "비밀번호에는 대문자가 포함되어야 합니다."

[en] "A data conflict occurred."
[id] "Terjadi konflik data."
[ko] "데이터 충돌이 발생했습니다."
```

Implikasi untuk app:

- Teks pesan **sudah** dilokalisasi server sesuai `X-Locale`, jadi bisa langsung
  ditampilkan ke user **selama** locale request diset benar.
- Tetap **jangan** parsing/mencocokkan teks pesan untuk menentukan jenis error —
  pakai **status code** (`409`/`422`/`401`) + **`error_code`** + **field key** dari objek `errors`.
- Untuk error berbentuk **string polos** (B-6), app tidak punya field key → susah
  memetakan ke input tertentu. Minta backend seragamkan ke format objek per-field.

---

## Catatan — memang disengaja (bukan bug)

- **Bisa login walau belum verifikasi.** Akun baru langsung berstatus `ACTIVE`
  dengan `is_email_verified: false` dan bisa langsung login. Sudah dikonfirmasi
  memang disengaja oleh tim (18 Juni 2026) — verifikasi hanya informasi, tidak
  menghalangi login.

---

## Akun test yang perlu dihapus (DB dev)

Semua akun dengan prefik berikut dibuat selama re-test ini (18 Juni 2026), aman dihapus
— semuanya berdomain `@kftest.com`:

- `qa.*`, `qacase*`, `QaCase*` (uji unik & case username)
- `r3.*` … `r14.*`, dan `r4_*` (akun tanpa phone) (uji field wajib & batas nilai)
- `r11.*` / `PASSWORD123` (uji password tanpa huruf kecil)
- `f5.*`, `f8.*` (phone `+62` & tanpa fullname yang lolos)
- `s1.*`, `s2.*` (uji role-injection & XSS fullname)
- `esc3.*`, `esc4.*` (verifikasi escalation + login)
- `g1*`, `g2*` (uji lokalisasi pesan)

> Tip cepat: hapus semua user dev dengan email `LIKE '%@kftest.com'`.
