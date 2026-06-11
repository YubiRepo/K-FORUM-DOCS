# API Spec — Session Bootstrap

> **Tujuan:** Single endpoint universal yang dipanggil sekali setelah login berhasil, baik dari mobile (Flutter) maupun web backoffice (Nuxt).
> Return semua data yang dibutuhkan client untuk render UI tanpa perlu hit endpoint lain.
> Hasil di-cache client-side (TTL 5 menit atau sampai invalidasi via push notif).

---

## Overview

### Kenapa satu endpoint untuk mobile dan web?

Data yang dibutuhkan Flutter dan Nuxt setelah login pada dasarnya **sama** — profil, plan, benefits, roles, permissions, admin_scopes. Tidak ada field yang exclusive untuk satu client saja. Memisahkan hanya karena konvensi prefix `/mobile/` vs `/web/` justru duplikasi logic di backend tanpa manfaat nyata.

Kondisi sebelumnya client harus hit beberapa endpoint terpisah:

| Endpoint lama | Data yang didapat |
|---|---|
| `POST /auth/login` | token + plan + roles + `permissions: []` (kosong) |
| `GET /role-permission/me` | roles + permissions + subscription_plan |
| `GET /subscription/benefits` | plan + benefits[] |

Dengan endpoint ini, semua cukup **satu request** setelah login — dari client manapun.

---

## Endpoint

### GET /api/v1/auth/session

- **URL:** `GET /api/v1/auth/session`
- **Auth:** Required (Bearer token)
- **Dipakai oleh:** Flutter mobile + Nuxt backoffice
- **Cache TTL (client-side):** 5 menit
- **Invalidasi cache:** Saat menerima push notif tipe `session_changed`

---

## Response Struktur

### Fields

| Field | Keterangan | Dipakai |
|---|---|---|
| `user` | Profil dasar | Mobile + Web |
| `subscription` | Plan aktif, status, expiry | Mobile + Web |
| `benefits[]` | Fitur yang bisa diakses berdasarkan plan | Mobile (primary), Web (show upgrade prompt) |
| `roles[]` | Semua role + scope | Mobile + Web |
| `permissions[]` | Permission efektif | Mobile + Web |
| `admin_scopes[]` | Derived view — region/platform yang dikelola | Web backoffice (primary), Mobile (show/hide admin tab) |

---

### Response 200 — Member biasa

```json
{
  "data": {
    "user": {
      "id": "usr_90210",
      "name": "Minji Park",
      "email": "user@example.com",
      "username": "minjipark",
      "avatar": "https://cdn.kai.app/avatars/minji.jpg"
    },

    "subscription": {
      "plan": "pro",
      "status": "active",
      "current_period_end": "2026-07-11",
      "auto_renewal": true
    },

    "benefits": [
      { "key": "read_content",     "enabled": true },
      { "key": "join_community",   "enabled": true },
      { "key": "post_community",   "enabled": true },
      { "key": "ask_qna",          "enabled": true },
      { "key": "post_news",        "enabled": true },
      { "key": "create_community", "enabled": true },
      { "key": "create_store",     "enabled": true },
      { "key": "create_event",     "enabled": true },
      { "key": "view_analytics",   "enabled": true },
      { "key": "post_ads",         "enabled": true }
    ],

    "roles": [
      {
        "name": "member",
        "role_type": "system",
        "scope_type": "global",
        "scope_id": null,
        "scope_name": null
      },
      {
        "name": "leader",
        "role_type": "community",
        "scope_type": "community",
        "scope_id": "community_futsal_uuid",
        "scope_name": "Futsal Players"
      }
    ],

    "permissions": [
      { "key": "read_content",   "scope": "global" },
      { "key": "join_community", "scope": "global" },
      { "key": "post_content",   "scope": "community" },
      { "key": "post_news",      "scope": "global" },
      { "key": "moderate_posts", "scope": "community" },
      { "key": "manage_members", "scope": "community" },
      { "key": "delete_content", "scope": "community" }
    ],

    "admin_scopes": []
  }
}
```

---

### Response 200 — Admin Region (multi-region)

```json
{
  "data": {
    "user": {
      "id": "usr_admin_001",
      "name": "Andi Pratama",
      "email": "andi@kai.app",
      "username": "andipratama",
      "avatar": "https://cdn.kai.app/avatars/andi.jpg"
    },

    "subscription": {
      "plan": "pro",
      "status": "active",
      "current_period_end": "2026-07-11",
      "auto_renewal": true
    },

    "benefits": [
      { "key": "read_content",     "enabled": true },
      { "key": "join_community",   "enabled": true },
      { "key": "post_community",   "enabled": true },
      { "key": "ask_qna",          "enabled": true },
      { "key": "post_news",        "enabled": true },
      { "key": "create_community", "enabled": true },
      { "key": "create_store",     "enabled": true },
      { "key": "create_event",     "enabled": true },
      { "key": "view_analytics",   "enabled": true },
      { "key": "post_ads",         "enabled": true }
    ],

    "roles": [
      {
        "name": "member",
        "role_type": "system",
        "scope_type": "global",
        "scope_id": null,
        "scope_name": null
      },
      {
        "name": "admin",
        "role_type": "system",
        "scope_type": "region",
        "scope_id": "region-jakarta-uuid",
        "scope_name": "Jakarta"
      },
      {
        "name": "admin",
        "role_type": "system",
        "scope_type": "region",
        "scope_id": "region-bandung-uuid",
        "scope_name": "Bandung"
      }
    ],

    "permissions": [
      { "key": "read_content",   "scope": "global" },
      { "key": "post_news",      "scope": "global" },
      { "key": "approve_news",   "scope": "global" },
      { "key": "manage_region",  "scope": "regional" },
      { "key": "manage_reports", "scope": "community" }
    ],

    "admin_scopes": [
      {
        "role": "admin",
        "scope_type": "region",
        "scope_id": "region-jakarta-uuid",
        "scope_name": "Jakarta"
      },
      {
        "role": "admin",
        "scope_type": "region",
        "scope_id": "region-bandung-uuid",
        "scope_name": "Bandung"
      }
    ]
  }
}
```

---

### Response 200 — Superadmin

```json
{
  "data": {
    "user": {
      "id": "usr_superadmin_001",
      "name": "KAI Pusat",
      "email": "superadmin@kai.app",
      "username": "kaipusat",
      "avatar": "https://cdn.kai.app/avatars/kai.jpg"
    },

    "subscription": {
      "plan": "pro",
      "status": "active",
      "current_period_end": null,
      "auto_renewal": false
    },

    "benefits": [
      { "key": "read_content",     "enabled": true },
      { "key": "join_community",   "enabled": true },
      { "key": "post_community",   "enabled": true },
      { "key": "ask_qna",          "enabled": true },
      { "key": "post_news",        "enabled": true },
      { "key": "create_community", "enabled": true },
      { "key": "create_store",     "enabled": true },
      { "key": "create_event",     "enabled": true },
      { "key": "view_analytics",   "enabled": true },
      { "key": "post_ads",         "enabled": true }
    ],

    "roles": [
      {
        "name": "superadmin",
        "role_type": "system",
        "scope_type": "global",
        "scope_id": null,
        "scope_name": null
      }
    ],

    "permissions": [
      { "key": "read_content",          "scope": "global" },
      { "key": "post_news",             "scope": "global" },
      { "key": "approve_news",          "scope": "global" },
      { "key": "assign_role",           "scope": "global" },
      { "key": "manage_region",         "scope": "regional" },
      { "key": "manage_news_sources",   "scope": "global" },
      { "key": "configure_auto_schedule","scope": "global" },
      { "key": "manage_reports",        "scope": "community" },
      { "key": "manage_bug_reports",    "scope": "global" }
    ],

    "admin_scopes": [
      {
        "role": "superadmin",
        "scope_type": "global",
        "scope_id": null,
        "scope_name": "KAI Pusat (All Regions)"
      }
    ]
  }
}
```

---

## Field `admin_scopes` — Detail

Field ini adalah **derived view** yang di-flatten dari `roles[]` — hanya system roles dengan scope `regional` atau `global`. Community roles (leader, moderator) tidak masuk ke sini.

| Kebutuhan client | Cara pakai `admin_scopes` |
|---|---|
| Cek apakah user adalah admin | `admin_scopes.length > 0` |
| Cek superadmin | `admin_scopes[0].scope_type === "global"` |
| Tampilkan region switcher di backoffice | Render dropdown dari `admin_scopes` |
| Pre-fill scope saat assign role | Pakai `scope_id` dari entry aktif |
| Breadcrumb "Admin Region Jakarta" | Pakai `scope_name` dari entry aktif |
| Show/hide backoffice tab di mobile | `admin_scopes.length > 0` |
| Guard route backoffice di Nuxt | `admin_scopes.length > 0`, redirect jika kosong |

> **Backend filtering tetap otomatis dari JWT** — admin region tidak perlu kirim `region_id` di setiap request. `admin_scopes` di session hanya untuk kebutuhan UI, bukan untuk menentukan akses di backend.

---

## Perubahan pada Login Response

Login (`POST /auth/login`) dan Google OAuth tetap return token + data minimal saja.
Client **wajib hit `/session`** setelah dapat token untuk data lengkap.

```json
{
  "token": "access_token_jwt",
  "refresh_token": "refresh_token_jwt",
  "data": {
    "id": "usr_90210",
    "name": "Minji Park",
    "email": "user@example.com",
    "username": "minjipark",
    "avatar": "https://cdn.kai.app/avatars/minji.jpg",
    "plan": "pro",
    "roles": ["member"]
  }
}
```

> `permissions: []` di login response **dihapus** — tidak berguna kalau kosong. Client langsung hit `/session` setelah dapat token.

---

## Seamless Login (Mobile → Backoffice)

Setelah Nuxt dapat Web Access Token dari flow seamless login, langkah berikutnya sama:

```
Flutter → generate seamless token → buka WebView
  └─ Nuxt verify token → dapat Web Access Token
       └─ Nuxt hit GET /api/v1/auth/session
            └─ Simpan ke state (Pinia/Vuex) → render dashboard
```

Tidak perlu endpoint session terpisah untuk web.

---

## Kapan Client Hit Endpoint Ini?

| Kondisi | Aksi |
|---|---|
| Setelah login berhasil (mobile & web) | Fetch `/session`, simpan ke cache |
| App resume dari background | Jika cache expired (>5 menit), refetch |
| Terima push notif `session_changed` | Invalidate cache, refetch |
| Dapat response `403` di endpoint mana pun | Invalidate cache, refetch, retry sekali |
| Nuxt route guard — user buka halaman backoffice | Cek cache, fetch jika belum ada |

---

## Endpoint Subscription yang Tetap Ada

| Endpoint | Kapan dipakai |
|---|---|
| `GET /api/v1/mobile/subscription/me` | Halaman "Kelola Langganan" — expiry, history, pending request |
| `GET /api/v1/mobile/subscription/plans` | Halaman upgrade/compare plan |
| `POST /api/v1/mobile/subscription/verify-benefit` | On-demand check sebelum aksi spesifik (fallback) |
| `GET /api/v1/mobile/subscription/history` | Halaman history pembayaran |

### Endpoint yang Dihapus / Digabung

| Endpoint lama | Alasan |
|---|---|
| `GET /api/v1/mobile/subscription/benefits` | Redundant — sudah masuk ke `/session` |
| `GET /api/v1/mobile/profile/subscription` | Redundant dengan `GET /subscription/me` |
| `GET /api/v1/mobile/role-permission/me` | Redundant — sudah masuk ke `/session` |
| `GET /api/v1/web/role-permission/me` | Redundant — sudah masuk ke `/session` |

---

## Backend Enforcement

> Client cache hanya untuk **UX** (show/hide button, guard route).
> Backend **selalu** re-check dari DB/Redis di setiap protected endpoint — tidak trust claim dari client.

```
Request masuk ke protected endpoint
  ├─ Middleware: verify JWT
  ├─ Middleware: check role permission (DB/Redis)
  ├─ Middleware: check benefit/plan (DB/Redis)
  ├─ Middleware: auto-inject region scope untuk admin region
  └─ Allow → handler / Deny → 403
```

### Redis Cache Backend

```
Key:   session:{user_id}
Value: { plan, benefits[], permissions[], roles[], admin_scopes[] }
TTL:   5 menit

Invalidasi saat:
- Plan user berubah (upgrade/downgrade/expired)
- Role user berubah (assign/revoke)
- Superadmin toggle benefit di plan
```

---

## Client-Side Implementation

### Flutter (Mobile)

```dart
Future<void> onLoginSuccess(String token) async {
  await storage.write('access_token', token);
  final session = await api.get('/api/v1/auth/session');
  await sessionCache.save(session.data);
  await sessionCache.setExpiry(Duration(minutes: 5));
}

bool isAdmin() => (sessionCache.get('admin_scopes') as List).isNotEmpty;
bool isSuperAdmin() => (sessionCache.get('admin_scopes') as List)
    .any((s) => s['role'] == 'superadmin');

List getManagedRegions() => (sessionCache.get('admin_scopes') as List)
    .where((s) => s['scope_type'] == 'region').toList();

bool canPostNews() => (sessionCache.get('benefits') as List)
    .any((b) => b['key'] == 'post_news' && b['enabled'] == true);
```

### Nuxt (Web Backoffice)

```typescript
// stores/session.ts (Pinia)
export const useSessionStore = defineStore('session', {
  state: () => ({
    user: null,
    subscription: null,
    benefits: [],
    roles: [],
    permissions: [],
    admin_scopes: [],
  }),

  actions: {
    async fetchSession() {
      const data = await $fetch('/api/v1/auth/session', {
        headers: { Authorization: `Bearer ${getToken()}` }
      });
      this.$patch(data.data);
    }
  },

  getters: {
    isAdmin: (state) => state.admin_scopes.length > 0,
    isSuperAdmin: (state) => state.admin_scopes.some(s => s.role === 'superadmin'),
    managedRegions: (state) => state.admin_scopes.filter(s => s.scope_type === 'region'),
  }
});

// middleware/auth.ts — route guard backoffice
export default defineNuxtRouteMiddleware(async () => {
  const session = useSessionStore();
  if (!session.isAdmin) {
    return navigateTo('/403');
  }
});
```

---

## Catatan Desain

**Kenapa satu endpoint universal, bukan `/mobile/` dan `/web/` terpisah?**
Data yang dibutuhkan kedua client identik. Memisahkan hanya karena konvensi prefix justru menduplikasi backend logic. Client yang tidak butuh field tertentu cukup ignore — `admin_scopes` kosong untuk member biasa tidak merugikan siapapun.

**Kenapa `admin_scopes` dipisah dari `roles[]`?**
`roles[]` berisi semua role termasuk community (leader, moderator). Backoffice hanya butuh system roles regional/global. `admin_scopes` adalah derived view yang sudah difilter sehingga client tidak perlu logic tambahan untuk mengekstraknya.

**Kenapa `permissions[]` dan `benefits[]` tetap dipisah?**
- `benefits` → show/hide fitur berdasarkan plan, trigger upgrade prompt jika disabled
- `permissions` → show/hide aksi berdasarkan role, tidak ada upgrade prompt

**Kenapa backend tetap enforce meski `admin_scopes` ada di session?**
Kalau hanya andalkan client-side, user bisa manipulasi cache. Backend selalu re-derive scope dari JWT + DB, tidak dari claim di request body atau header.
