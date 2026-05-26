# API Spec — Seamless Login (Mobile to Web)

Dokumentasi ini menjelaskan *flow* dan *endpoint* yang dibutuhkan untuk melakukan **Seamless Login** (One-Time Token / Magic Link) dari aplikasi Mobile (Flutter) ke Web Backoffice (Nuxt.js). 

Tujuannya agar user yang sudah *login* di Mobile App dapat membuka Backoffice tanpa harus melakukan *login* ulang.

---

## Architecture Flow

1. **Flutter** memanggil API `Generate Web Token` menggunakan `Authorization: Bearer <Mobile_JWT>`.
2. **Backend** memvalidasi user, men-*generate* sebuah token unik yang bersifat *one-time use* (sekali pakai) dengan *expired time* yang sangat singkat (misal: 1 menit), dan mereturn token tersebut ke Flutter.
3. **Flutter** membuka browser/WebView dengan URL: `https://[DOMAIN_BACKOFFICE]/auth/seamless?token=<One_Time_Token>`.
4. **Nuxt.js** membaca parameter `token` di URL, lalu menembakkan API `Verify Web Token` ke Backend.
5. **Backend** memvalidasi *One-Time Token*. Jika valid:
   - Hapus/hanguskan token tersebut agar tidak bisa dipakai lagi (One-Time Use).
   - Generate *Web Access Token* (JWT standar untuk Web) dan mereturn-nya.
6. **Nuxt.js** menyimpan *Web Access Token* ke *cookies*, meng-*update state auth*, lalu me-redirect user ke halaman *Dashboard*.

---

## 1. Endpoint: Generate Web Token (Dipanggil oleh Flutter)

Endpoint ini diakses dari *Mobile App* (Flutter) ketika user menekan tombol "Buka Backoffice".

- **URL**: `POST /api/v1/auth/seamless/generate-token`
- **Headers**:
  - `Content-Type: application/json`
  - `Authorization: Bearer <Mobile_Access_Token>` (Wajib)
- **Request Body**: (Kosong atau opsional untuk *metadata* tracking)
- **Response (Success 201)**:
  ```json
  {
    "status": "success",
    "message": "One-time token generated successfully",
    "data": {
      "seamless_token": "a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6",
      "expires_in": 60, // Token expired dalam 60 detik
      "redirect_url": "https://backoffice.domain.com/auth/seamless?token=a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"
    }
  }
  ```

> **Security Note untuk Backend:**
> - Token yang di-*generate* (`seamless_token`) wajib disimpan di *Database* atau *Redis*.
> - Beri batasan kedaluwarsa maksimal **1 - 2 menit** saja.
> - Relasikan token ini dengan ID User yang sedang login di Mobile.

---

## 2. Endpoint: Verify Web Token (Dipanggil oleh Nuxt Web)

Endpoint ini diakses oleh Web Backoffice (Nuxt.js) untuk menukarkan `seamless_token` menjadi `Web Access Token` yang sah.

- **URL**: `POST /api/v1/auth/seamless/verify-token`
- **Headers**:
  - `Content-Type: application/json`
- **Request Body**:
  ```json
  {
    "seamless_token": "a1b2c3d4-e5f6-7g8h-9i0j-k1l2m3n4o5p6"
  }
  ```
- **Response (Success 200)**:
  ```json
  {
    "status": "success",
    "message": "Token exchanged successfully",
    "data": {
      "user": {
        "id": "uuid-user-123",
        "fullname": "Andi Darmawan",
        "email": "andi@example.com",
        "role": "superadmin"
      },
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.web_token_asli...",
      "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.refresh..."
    }
  }
  ```
- **Response (Error 401 - Expired/Invalid)**:
  ```json
  {
    "status": "error",
    "status_code": 401,
    "message": "Seamless token is invalid or has expired."
  }
  ```

> **Security Note untuk Backend:**
> - Saat token berhasil divalidasi, backend **WAJIB LANGSUNG MENGHAPUS** (atau menandai `is_used = true`) token tersebut dari Database/Redis. Hal ini krusial untuk mencegah serangan celah keamanan jika ada pihak lain yang mencoba memanggil URL tersebut lagi.

---

## 3. Implementasi Frontend (Nuxt.js)

Sebagai gambaran untuk *Web Developer*, di Nuxt.js kita hanya perlu membuat satu halaman `pages/auth/seamless.vue` yang bertugas sebagai penengah (*middleware*).

```vue
<!-- file: app/pages/auth/seamless.vue -->
<script setup lang="ts">
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore() // Asumsi menggunakan Pinia
const toast = useToast()

const token = route.query.token as string

onMounted(async () => {
  if (!token) {
    toast.add({ title: 'Invalid Link', description: 'No token provided.', color: 'error' })
    return router.push('/auth/login')
  }

  try {
    // 1. Tembak API Verify
    const response = await $fetch('/api/v1/auth/seamless/verify-token', {
      method: 'POST',
      body: { seamless_token: token }
    })

    // 2. Simpan token asli dan session user
    authStore.setToken(response.data.token)
    authStore.setUser(response.data.user)

    // 3. Redirect ke Dashboard
    toast.add({ title: 'Welcome Back', description: 'Seamless login successful!', color: 'success' })
    router.push('/')

  } catch (error) {
    toast.add({ title: 'Login Failed', description: 'Link expired or invalid.', color: 'error' })
    router.push('/auth/login')
  }
})
</script>

<template>
  <div class="h-screen flex flex-col items-center justify-center">
    <!-- UI Loading Indicator -->
    <UIcon name="i-lucide-loader-2" class="w-12 h-12 animate-spin text-primary-500 mb-4" />
    <h2 class="text-xl font-medium">Authenticating securely...</h2>
    <p class="text-sm text-gray-500">Please wait while we log you in.</p>
  </div>
</template>
```

---

## 4. Implementasi Mobile (Flutter)

Sebagai gambaran untuk *Mobile Developer*, pemanggilan dapat dilakukan seperti ini menggunakan paket `url_launcher` dan `dio`.

```dart
Future<void> openBackoffice() async {
  try {
    // 1. Request One-Time Token dari backend
    final response = await dio.post('/api/v1/auth/seamless/generate-token');
    
    // 2. Ambil URL Redirect yang sudah berisi token
    final redirectUrl = response.data['data']['redirect_url'];
    
    // 3. Buka URL menggunakan In-App Browser atau Eksternal Browser
    final Uri url = Uri.parse(redirectUrl);
    if (await canLaunchUrl(url)) {
      await launchUrl(
        url, 
        // Disarankan menggunakan inAppWebView agar user tidak pindah ke Chrome
        mode: LaunchMode.inAppWebView 
      );
    }
  } catch (e) {
    print('Gagal membuka backoffice: $e');
  }
}
```
