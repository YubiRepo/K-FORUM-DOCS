# Dokumentasi API Spec - Modul Authentication (Mobile Client)

Dokumentasi ini dibuat untuk kebutuhan tim Backend agar skema request/response API sesuai dengan implementasi Clean Architecture pada Flutter Mobile Client.

## Informasi Umum

- **Base URL Prefix**: `/api/v1` (Seluruh endpoint di bawah ini menggunakan prefix `/api/v1/mobile/auth`)
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Accept-Language: <lang_code>` (Mengirimkan kode bahasa aktif, contoh: `ko`, `id`, `en`. Default: `ko`)
  - `X-Locale: <lang_code>` (Mengirimkan kode bahasa aktif, contoh: `ko`, `id`, `en`. Default: `ko`)
  - `Authorization: Bearer <access_token>` (Hanya untuk endpoint yang membutuhkan autentikasi)

---

## Model Data Utama

### 1. User Object (Common Response Schema)
Setiap kali endpoint mengembalikan data User, strukturnya harus berupa format berikut di dalam field `data`:

```json
{
  "id": "string",
  "fullname": "string",
  "email": "string",
  "username": "string",
  "avatar": "string (nullable / URL)",
  "phone": "string (nullable)",
  "google_id": "string (nullable)",
  "plan": "string (e.g. 'standart', 'free')",
  "roles": ["string"],
  "permissions": ["string"],
  "created_at": "string (ISO 8601 UTC format, e.g. 2026-05-20T00:00:00.000Z)"
}
```

### 2. Error Responses
Frontend menangani 2 skema error utama dari backend:

#### A. Standard Message Error (HTTP 4xx / 5xx)
```json
{
  "message": "Pesan error deskriptif di sini"
}
```

#### B. Validation Error (HTTP 422 Unprocessable Entity)
Jika terdapat kesalahan input form, backend disarankan mengembalikan objek `errors` yang berisi key field dan array pesan kesalahan. Mobile client akan mengambil pesan error pertama untuk ditampilkan.
```json
{
  "message": "Data input tidak valid",
  "errors": {
    "email": ["Format email tidak valid."],
    "password": ["Password minimal harus terdiri dari 8 karakter."]
  }
}
```

---

## Daftar Endpoint

### 1. Login (Email & Password)
Melakukan autentikasi menggunakan email /username dan password.

- **URL**: `POST /api/v1/mobile/auth/login`
- **Autentikasi**: Tidak
- **Request Body**:
  ```json
  {
    "identifier": "user@example.com",
    "password": "userpassword"
  }
  ```
- **Response (Success 200/201)**:
  ```json
  {
    "token": "access_token_jwt_string",
    "refresh_token": "refresh_token_jwt_string",
    "data": {
      "id": "usr_90210",
      "fullname": "Minji Park",
      "email": "user@example.com",
      "username": "minjipark",
      "avatar": "https://example.com/avatar.jpg",
      "phone": "08123456789",
      "google_id": null,
      "plan": "standart",
      "roles": ["member"],
      "permissions": [],
      "created_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```

---

### 2. Google Authentication (Login & Sign-Up)
Unified Auth Flow untuk login atau registrasi secara otomatis menggunakan ID Token dari Google Sign-In.

- **URL**: `POST /api/v1/mobile/auth/login/google`
- **Autentikasi**: Tidak
- **Request Body**:
  ```json
  {
    "id_token": "google_id_token_jwt_string",
    "access_token": "optional_google_access_token_string"
  }
  ```
- **Response (Success 200/201)**:
  - Jika email Google belum terdaftar ➡️ Registrasi akun secara otomatis dan kembalikan response token + user.
  - Jika email Google sudah terdaftar ➡️ Login dan kembalikan response token + user.
  ```json
  {
    "token": "access_token_jwt_string",
    "refresh_token": "refresh_token_jwt_string",
    "data": {
      "id": "usr_90211",
      "name": "Minji Google",
      "email": "user.google@example.com",
      "username": "minjigoogle",
      "avatar": "https://lh3.googleusercontent.com/a/avatar_url",
      "phone": null,
      "google_id": "google_oauth_sub_id",
      "plan": "standart",
      "roles": ["member"],
      "permissions": [],
      "created_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```

---

### 3. Register (Pendaftaran Akun Baru)
Melakukan registrasi akun baru dengan kredensial biasa.

- **URL**: `POST /api/v1/mobile/auth/register`
- **Autentikasi**: Tidak
- **Request Body**:
  ```json
  {
    "username": "Minji Park",
    "fullname": "Minji Park", // optional
    "email": "minji-park@mail.com",
    "phone": "08123456789",
    "password": "userpassword"
  }
  ```
- **Response (Success 200/201)**:
  ```json
  {
    "message": "Registration successful. Please verify OTP.",
    "data": {
      "id": "usr_90212",
      "fullname": "Minji Park",
      "email": "minji-park@mail.com",
      "username": "minjipark123",
      "avatar": null,
      "phone": "08123456789",
      "google_id": null,
      "plan": "standart",
      "roles": ["member"],
      "permissions": [],
      "created_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```

---

### 4. Logout
Mengakhiri session token JWT aktif.

- **URL**: `POST /api/v1/mobile/auth/logout`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**: None (Menggunakan Authorization Header)
- **Response (Success 200)**:
  ```json
  {
    "message": "Logout successful"
  }
  ```

---

### 5. Refresh Token
Mendapatkan Access Token baru menggunakan Refresh Token.

- **URL**: `POST /api/v1/mobile/auth/refresh`
- **Autentikasi**: Tidak (Menggunakan payload refresh token)
- **Request Body**:
  ```json
  {
    "refresh_token": "refresh_token_jwt_string"
  }
  ```
- **Response (Success 200/201)**:
  ```json
  {
    "token": "new_access_token_jwt_string",
    "refresh_token": "optional_new_refresh_token_jwt_string"
  }
  ```

---

### 6. Send OTP (Kirim OTP ke Email)
Mengirimkan kode OTP verifikasi ke email pengguna.

- **URL**: `POST /api/v1/mobile/auth/otp/send`
- **Autentikasi**: Tidak
- **Request Body**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Kode OTP berhasil dikirim ke email Anda"
  }
  ```

---

### 7. Verify OTP (Verifikasi OTP)
Memverifikasi OTP yang dimasukkan oleh pengguna.

- **URL**: `POST /api/v1/mobile/auth/otp/verify`
- **Autentikasi**: Tidak
- **Request Body**:
  ```json
  {
    "identifier": "user@example.com",
    "otp": "1234"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "OTP verified successfully",
    "token": "temp_verification_token_for_reset_password_or_auth"
  }
  ```

---

### 8. Forgot Password (Lupa Password - Kirim OTP/Link)
Mengajukan permintaan reset password jika pengguna lupa kata sandi mereka. Backend akan mengirimkan kode OTP atau link ke email pengguna.

- **URL**: `POST /api/v1/mobile/auth/password/forgot`
- **Autentikasi**: Tidak
- **Request Body**:
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Link/Kode OTP reset password telah dikirim ke email Anda"
  }
  ```

---

### 9. Reset Password (Mengubah Password via OTP/Token)
Mengubah password lama ke password baru menggunakan kode verifikasi/token yang valid.

> [!NOTE]
> **Alur Forgot & Reset Password (User Belum Login)**:
> 1. User memanggil `POST /password/forgot` ➡️ Backend mengirim OTP ke email.
> 2. User memverifikasi OTP dengan `POST /otp/verify` ➡️ Backend memvalidasi dan mengembalikan `token` (token sementara JWT).
> 3. User memanggil `POST /password/reset` dengan mengirimkan `email`, `token` (dari langkah 2), dan `password` baru. 
> *Catatan*: Jika backend memilih untuk tidak mengembalikan token pada langkah 2, backend dapat mengizinkan `token` pada langkah 3 diisi langsung dengan kode OTP yang dimasukkan user.

- **URL**: `POST /api/v1/mobile/auth/password/reset`
- **Autentikasi**: Tidak
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "token": "temp_verification_token_or_otp",
    "password": "newpassword123"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Password berhasil diperbarui"
  }
  ```

---

### 10. Get Me (Profil Pengguna Aktif)
Mengambil informasi profil dari pengguna yang sedang login saat ini.

- **URL**: `GET /api/v1/mobile/auth/me`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**: None (Menggunakan Authorization Header)
- **Response (Success 200)**:
  ```json
  {
    "data": {
      "id": "usr_90210",
      "name": "Minji Park",
      "email": "user@example.com",
      "username": "minjipark",
      "avatar": "https://example.com/avatar.jpg",
      "phone": "08123456789",
      "google_id": null,
      "plan": "standart",
      "roles": ["member"],
      "permissions": [],
      "created_at": "2026-05-20T00:00:00.000Z"
    }
  }
  ```

---

### 11. Change Password (Ubah Password dari Pengaturan)
Mengubah password untuk pengguna yang sedang aktif/login. Pengguna harus memasukkan password lama untuk verifikasi sebelum menggantinya dengan password baru.

- **URL**: `POST /api/v1/mobile/auth/password/change`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "old_password": "currentpassword123",
    "password": "newpassword123"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "message": "Password changed successfully"
  }
  ```

