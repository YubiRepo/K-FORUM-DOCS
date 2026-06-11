# User Management — Rules & Use Cases

> **Status:** Draft v1
> **Last Updated:** 2026-06-11
> **Module:** User Management (Backoffice)
> **Berkaitan dengan:** `ROLE_PERMISSION_SYSTEM.md`, `REGION_SYSTEM_RULES.md`, `PLAN_SUBSCRIPTION_SYSTEM.md`, `API_SPEC_BULK_ROLE_ASSIGNMENT_BACKOFFICE.md`

---

## Daftar Isi

1. [Overview](#1-overview)
2. [Aktor & Scope Akses](#2-aktor--scope-akses)
3. [Entitas Utama](#3-entitas-utama)
4. [Permission Matrix](#4-permission-matrix)
5. [Manajemen Data User](#5-manajemen-data-user)
6. [Assign Role — Single User](#6-assign-role--single-user)
7. [Assign Role — Bulk (Multiple Users)](#7-assign-role--bulk-multiple-users)
8. [Assign Role per Region](#8-assign-role-per-region)
9. [Status & Lifecycle User Account](#9-status--lifecycle-user-account)
10. [Subscription Management oleh Admin](#10-subscription-management-oleh-admin)
11. [Audit Log](#11-audit-log)
12. [Use Cases](#12-use-cases)
13. [Business Rules & Constraints](#13-business-rules--constraints)
14. [Edge Cases](#14-edge-cases)
15. [Integrasi dengan Modul Lain](#15-integrasi-dengan-modul-lain)

---

## 1. Overview

Modul **User Management** adalah modul backoffice yang memungkinkan Superadmin dan Admin Region untuk mengelola akun user di platform KAI. Lingkupnya mencakup:

- Melihat dan mengedit profil user
- Mengelola status akun (aktif, suspend, nonaktif)
- Assign dan revoke role — per user maupun bulk sekaligus
- Assign role dengan scope region atau scope community
- Mengelola subscription plan user secara manual
- Reset password
- Melihat riwayat aktivitas dan audit log

Modul ini **tidak mencakup** manajemen komunitas internal (itu di modul Community) dan tidak mencakup pembuatan user baru (itu di modul Auth/Register).

---

## 2. Aktor & Scope Akses

### Aktor yang Bisa Mengakses Modul Ini

| Aktor | Scope Data | Keterangan |
|---|---|---|
| `usergod` | Semua user, semua region | Akses penuh, hanya developer |
| `superadmin` | Semua user, semua region | Admin KAI Pusat — akses platform-wide |
| `admin` (region) | Hanya user di region-nya | Admin wilayah — akses terbatas pada region sendiri |

> Member biasa dan Guest **tidak bisa** mengakses backoffice user management.

### Perbedaan Scope Superadmin vs Admin Region

```
Superadmin:
  ✅ Lihat semua user di semua region
  ✅ Assign role apapun ke user manapun
  ✅ Suspend user (termasuk admin region)
  ✅ Upgrade/downgrade subscription user
  ✅ Assign role superadmin ke user lain
  ✅ Hapus (soft delete) akun user
  ✅ Export data semua user (CSV)

Admin Region:
  ✅ Lihat user di region mereka sendiri
  ✅ Assign role "member" di region mereka
  ✅ Assign role "admin" di region mereka (perlu konfirmasi)
  ✅ Reset password user di region mereka
  ✅ Activate/deactivate user di region mereka
  ❌ Tidak bisa suspend user
  ❌ Tidak bisa assign role superadmin
  ❌ Tidak bisa ubah subscription
  ❌ Tidak bisa lihat user di luar region mereka
  ❌ Tidak bisa export semua user
```

---

## 3. Entitas Utama

### User (dalam konteks User Management)

Data user yang dikelola di backoffice:

```
id:                   UUID (primary key)
name:                 string
email:                string (unique)
username:             string (unique)
phone:                string (nullable)
avatar:               string (URL, nullable)
status:               "active" | "inactive" | "suspended"
email_verified:       boolean
phone_verified:       boolean
subscription_plan:    "standard" | "pro"
subscription_status:  "active" | "expired" | "cancelled"
subscription_expired_at: timestamp (nullable)
region_id:            UUID (nullable, FK → regions.id)
created_at:           timestamp
updated_at:           timestamp
last_login:           timestamp (nullable)
deleted_at:           timestamp (nullable, soft delete)
```

### UserRole (assignment)

```
id:           UUID
user_id:      UUID (FK → users.id)
role_id:      UUID (FK → roles.id)
scope_type:   "global" | "region" | "community"
scope_id:     UUID (nullable — null untuk global scope)
assigned_by:  UUID (FK → users.id, admin yang assign)
assigned_at:  timestamp
expired_at:   timestamp (nullable, untuk temporary role)
is_active:    boolean (soft revoke)
notes:        text (nullable, alasan assignment)
```

---

## 4. Permission Matrix

### Aksi yang Bisa Dilakukan per Aktor

| Aksi | usergod | superadmin | admin region |
|---|:---:|:---:|:---:|
| List semua user platform | ✅ | ✅ | ❌ |
| List user di region sendiri | ✅ | ✅ | ✅ |
| Lihat detail user | ✅ | ✅ | ✅ (region sendiri) |
| Edit profil user | ✅ | ✅ | ✅ (region sendiri) |
| Reset password user | ✅ | ✅ | ✅ (region sendiri) |
| Activate / deactivate user | ✅ | ✅ | ✅ (region sendiri) |
| Suspend user | ✅ | ✅ | ❌ |
| Soft delete akun | ✅ | ✅ | ❌ |
| Assign role member | ✅ | ✅ | ✅ (di region sendiri) |
| Assign role admin region | ✅ | ✅ | ✅ (di region sendiri, perlu konfirmasi) |
| Assign role superadmin | ✅ | ✅ | ❌ |
| Revoke role | ✅ | ✅ | ✅ (role yang mereka assign saja) |
| Bulk assign role | ✅ | ✅ | ✅ (scope region sendiri) |
| Bulk assign via CSV | ✅ | ✅ | ✅ (scope region sendiri) |
| Ubah subscription plan | ✅ | ✅ | ❌ |
| Export data user (CSV) | ✅ | ✅ | ✅ (region sendiri saja) |
| Lihat audit log | ✅ | ✅ | ✅ (region sendiri) |

---

## 5. Manajemen Data User

### 5.1 List & Search User

Admin bisa filter user berdasarkan:

- `search` — nama, email, atau username
- `status` — `active`, `inactive`, `suspended`
- `subscription_plan` — `standard`, `pro`
- `subscription_status` — `active`, `expired`, `cancelled`
- `role` — filter by system role
- `region_id` — filter by region (superadmin only)
- `sort` — `created_at`, `last_login`, `name` (prefix `-` untuk desc)

Superadmin melihat semua user. Admin Region hanya melihat user di region mereka.

### 5.2 Detail User

Response detail user mencakup:
- Data profil lengkap
- Status akun dan verifikasi
- Subscription aktif + riwayat singkat
- Semua role yang aktif + scope masing-masing
- Daftar permission efektif (union dari semua role)
- Region membership
- Stats: total komunitas, posts, hari aktif

### 5.3 Edit Profil User

Admin dapat mengedit: `name`, `email`, `phone`, `avatar`.

**Rules:**
- Email baru tidak boleh konflik dengan user lain
- Perubahan email **tidak** membutuhkan re-verifikasi oleh admin (opsional kirim notif ke user)
- Admin tidak bisa edit `username` — hanya user sendiri yang bisa
- Admin tidak bisa edit `password` langsung — hanya bisa kirim reset link

### 5.4 Reset Password

Admin mengirim reset link ke email user. User yang klik link bisa set password baru.

**Rules:**
- Reset link expired dalam 24 jam
- Hanya satu reset link aktif per user (request baru invalidate link lama)
- Admin tidak bisa melihat atau set password secara langsung

---

## 6. Assign Role — Single User

### Assign Satu Role ke Satu User

Admin membuka detail user → klik "Assign Role" → pilih role, scope type, dan scope id.

**Fields saat assign:**

| Field | Required | Keterangan |
|---|---|---|
| `role_id` | ✅ | Role yang akan di-assign |
| `scope_type` | ✅ | `global`, `region`, atau `community` |
| `scope_id` | Conditional | Wajib jika scope_type = region atau community |
| `expired_at` | ❌ | Opsional — untuk temporary role |
| `notes` | ❌ | Alasan assignment (untuk audit) |

**Rules:**
- Satu user bisa punya **multiple roles di scope berbeda**
- Tidak bisa assign role yang sudah ada di scope yang sama (duplicate check)
- Admin region hanya bisa assign ke scope region mereka sendiri
- Admin region tidak bisa assign role `superadmin`
- `usergod` tidak bisa di-assign via UI — hanya lewat DB/developer
- Jika `expired_at` diisi, role otomatis revoked setelah tanggal tersebut (cron job)

### Assign Multiple Roles ke Satu User (Multi-Role Assignment)

Dari halaman detail user, admin bisa assign beberapa role sekaligus dalam satu operasi:

```
User: Andi Pratama

Role #1: Admin Region → Scope: Region → Jakarta
Role #2: Community Moderator → Scope: Community → Futsal Jakarta
Role #3: Community Leader → Scope: Community → Basket Selatan

[Assign 3 Roles]
```

Ini adalah satu atomic operation — kalau salah satu gagal, yang lain tetap dilanjutkan dan hasilnya dilaporkan per-role.

### Revoke Role dari User

Admin bisa revoke role aktif dari user. Revoke menggunakan soft delete (`is_active = false`) bukan hard delete, untuk keperluan audit trail.

**Rules:**
- Admin region hanya bisa revoke role yang **dia sendiri assign** di region mereka
- Superadmin bisa revoke role apapun
- Revoking role `member` (system role default) tidak diperbolehkan — user selalu punya setidaknya role member

---

## 7. Assign Role — Bulk (Multiple Users)

### Bulk Assign: Satu Role ke Banyak User

Admin memilih multiple user dari list → pilih role dan scope → assign sekaligus.

**Input:**
```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3", ...],
  "role_id": "uuid",
  "scope_type": "region",
  "scope_id": "region_uuid",
  "expired_at": null,
  "notes": "Batch assignment untuk regional admins baru"
}
```

**Rules:**
- Maksimal **100 user** per satu bulk operation
- User yang sudah punya role yang sama di scope yang sama → di-skip (tidak error, terhitung `skipped`)
- User yang tidak ditemukan → error per-item, tidak stop keseluruhan operasi
- Hasil operasi selalu return `operation_id` untuk tracking

### Bulk Assign via CSV Upload

Upload CSV dengan daftar user untuk di-assign role yang sama.

**Format CSV:**
```csv
user_id,email,name
uuid1,andi@example.com,Andi Pratama
uuid2,budi@example.com,Budi Santoso
```

**Rules:**
- Kolom `user_id` atau `email` wajib ada (boleh salah satu, kalau keduanya ada → `user_id` dipakai)
- Maksimal **500 baris** per upload
- CSV divalidasi dulu sebelum diproses — kalau ada error format, seluruh file ditolak
- File non-CSV (Excel, dll) tidak diterima — harus .csv murni
- Setelah upload sukses, hasilnya bisa di-download sebagai laporan

### Bulk Revoke

Remove satu role dari banyak user sekaligus.

**Input:**
```json
{
  "user_ids": ["uuid1", "uuid2", "uuid3"],
  "role_id": "uuid",
  "scope_type": "region",
  "scope_id": "region_uuid",
  "reason": "Reorganisasi admin regional"
}
```

---

## 8. Assign Role per Region

### Konsep Scope Region

Role yang di-assign dengan `scope_type: region` hanya berlaku di region tersebut. User yang punya role `admin` di Region Jakarta **tidak punya kekuatan admin** di Region Surabaya.

### Siapa yang Bisa Assign Admin Region

| Yang Assign | Target Role | Scope | Boleh? |
|---|---|---|---|
| Superadmin | admin | Region Jakarta | ✅ |
| Superadmin | admin | Region Surabaya | ✅ |
| Admin Jakarta | admin | Region Jakarta | ✅ (perlu konfirmasi popup) |
| Admin Jakarta | admin | Region Surabaya | ❌ |
| Admin Jakarta | superadmin | Region apapun | ❌ |

**Konfirmasi popup saat admin region assign admin baru di region mereka:**
> "Kamu akan assign role Admin Region Jakarta ke [Nama User]. User ini akan punya akses yang sama dengan kamu di region ini. Lanjutkan?"

### Transfer Admin Region

Jika admin region ingin "transfer" posisinya ke user lain:
1. Superadmin assign admin baru di region tersebut
2. Superadmin revoke role admin dari user lama
3. Tidak ada fitur "transfer" otomatis — selalu manual dua langkah

### Multi-Region Admin

Satu user bisa jadi admin di lebih dari satu region. Tidak ada batasan jumlah region, tapi setiap assignment harus eksplisit.

---

## 9. Status & Lifecycle User Account

### Status User

| Status | Keterangan | Bisa Login? | Konten Tampil? |
|---|---|---|---|
| `active` | Normal | ✅ Ya | ✅ Ya |
| `inactive` | Dinonaktifkan (by user atau admin) | ❌ Tidak | ✅ Ya (konten lama tetap) |
| `suspended` | Diblokir karena pelanggaran | ❌ Tidak | ❌ Konten disembunyikan |

### Siapa yang Bisa Ubah Status

| Dari → Ke | superadmin | admin region |
|---|:---:|:---:|
| active → inactive | ✅ | ✅ (region sendiri) |
| active → suspended | ✅ | ❌ |
| inactive → active | ✅ | ✅ (region sendiri) |
| suspended → active | ✅ | ❌ |
| suspended → inactive | ✅ | ❌ |

**Rules:**
- Semua perubahan status wajib menyertakan `reason` (dicatat di audit log)
- User yang di-suspend menerima notifikasi email dengan alasan singkat
- Superadmin tidak bisa suspend sesama superadmin — hanya usergod yang bisa
- User yang di-suspend bisa appeal via email ke KAI Pusat (di luar scope sistem ini)

### Soft Delete (Hapus Akun)

Hanya superadmin yang bisa soft delete akun user. Ini berbeda dari suspend:

- **Suspend:** User masih ada di DB, konten disembunyikan, bisa di-unsuspend
- **Soft Delete:** `deleted_at` diisi, user tidak tampil di listing, tidak bisa login, tapi data tetap di DB untuk keperluan audit

**Rules:**
- Soft delete hanya bisa dilakukan jika user bukan superadmin aktif lain
- Sebelum soft delete: system auto-revoke semua role aktif user
- Konten yang dibuat user (news, posts, events) tidak ikut dihapus — hanya akun yang ditandai deleted
- Tidak ada fitur restore dari soft delete via UI — harus manual oleh usergod jika diperlukan

---

## 10. Subscription Management oleh Admin

### Superadmin: Manual Plan Change

Superadmin bisa upgrade atau downgrade plan user secara manual — tanpa perlu approval flow biasa.

**Fields:**
```
plan:           "standard" | "pro"
effective_date: tanggal berlaku (default: hari ini)
reason:         wajib diisi (untuk audit)
skip_approval:  true (sudah implicit untuk superadmin)
```

**Rules:**
- Manual change oleh superadmin langsung efektif tanpa menunggu konfirmasi user
- History tetap tercatat di `subscription_history` dengan `changed_by: superadmin_id`
- Jika ada pending upgrade request dari user yang sama → auto-cancelled saat superadmin manual change
- Tidak ada refund logic di sini — refund diproses terpisah secara manual

### Admin Region: Tidak Bisa Ubah Subscription

Admin region tidak punya akses untuk mengubah plan subscription user. Jika ada permintaan upgrade di region mereka, admin region cukup forward ke superadmin atau arahkan user ke flow upgrade biasa.

---

## 11. Audit Log

Semua aksi di modul User Management dicatat di audit log dengan struktur:

```
id:          UUID
actor_id:    UUID (user yang melakukan aksi)
actor_role:  "superadmin" | "admin" | "usergod"
action:      string (lihat daftar di bawah)
target_type: "user" | "role_assignment" | "subscription"
target_id:   UUID
old_value:   JSON (state sebelum)
new_value:   JSON (state sesudah)
notes:       text (reason dari admin, nullable)
ip_address:  string
created_at:  timestamp
```

**Daftar action yang di-log:**

| Action | Trigger |
|---|---|
| `user.profile_updated` | Edit profil user |
| `user.status_changed` | Activate / deactivate / suspend |
| `user.deleted` | Soft delete akun |
| `user.password_reset_sent` | Reset password link dikirim |
| `role.assigned` | Assign role ke user |
| `role.revoked` | Revoke role dari user |
| `role.bulk_assigned` | Bulk assign role |
| `role.bulk_revoked` | Bulk revoke role |
| `role.expired` | Role expired otomatis (by cron) |
| `subscription.changed` | Manual plan change oleh admin |

Audit log **tidak bisa diedit atau dihapus** — immutable. Hanya bisa di-query (read-only).

---

## 12. Use Cases

### Use Case 1: Superadmin Assign Admin Baru untuk Region Jakarta

```
1. Superadmin buka User Management → search "Andi"
2. Buka detail Andi Pratama
3. Klik [Assign Role]
4. Pilih: Role = "Admin Region", Scope = "Region", Region = "Jakarta"
5. Isi expired_at = 2026-12-31 (kontrak 6 bulan)
6. Isi notes = "Admin baru untuk region Jakarta"
7. Klik [Assign]
8. System:
   - Cek: superadmin punya hak assign admin? ✅
   - Cek: Andi belum punya role admin di Jakarta? ✅
   - Insert user_roles (scope_type=region, scope_id=region_jakarta, expired_at=2026-12-31)
   - Log audit: role.assigned
   - Kirim notifikasi ke Andi: "Kamu telah di-assign sebagai Admin Region Jakarta"
9. Andi sekarang bisa akses backoffice region Jakarta
```

### Use Case 2: Bulk Assign Admin ke 5 User Sekaligus

```
1. Superadmin buka User Management → filter Region = "Surabaya"
2. Centang 5 user (checkbox)
3. Klik [Bulk Actions] → [Assign Role]
4. Pilih: Role = "Member", Scope = "Region", Region = "Surabaya"
5. Klik [Assign to 5 Users]
6. System proses:
   - User 1 (Budi): belum punya role ini → assigned ✅
   - User 2 (Citra): sudah punya role ini → skipped ℹ️
   - User 3 (Doni): belum punya → assigned ✅
   - User 4 (Eka): user tidak aktif → error ⚠️ (tapi lanjut)
   - User 5 (Fani): belum punya → assigned ✅
7. Hasil: 3 assigned, 1 skipped, 1 error
8. Admin bisa download laporan hasil operasi
```

### Use Case 3: Admin Region Surabaya Assign Admin Baru

```
1. Admin Surabaya login → buka User Management
2. Hanya melihat user di Region Surabaya
3. Cari "Gita" → buka detail Gita Wulandari
4. Klik [Assign Role]
5. Pilih: Role = "Admin Region"
   → Scope otomatis = "Region Surabaya" (tidak bisa pilih region lain)
6. Muncul konfirmasi: "Gita akan punya akses admin yang sama dengan kamu di Surabaya. Lanjutkan?"
7. Konfirmasi → assign
8. Audit log: actor = admin_surabaya, action = role.assigned, target = gita
```

### Use Case 4: Suspend User yang Melanggar

```
1. Superadmin terima laporan user "Hadi" spam di komunitas
2. Buka detail Hadi → klik [Change Status] → pilih "Suspended"
3. Wajib isi reason: "Spam berulang di 3 komunitas berbeda dalam 24 jam"
4. Klik [Confirm Suspend]
5. System:
   - Status Hadi = suspended
   - Semua session aktif Hadi di-invalidate (logout paksa)
   - Konten Hadi disembunyikan dari publik
   - Email notifikasi dikirim ke Hadi
   - Log audit: user.status_changed (active → suspended)
6. Hadi tidak bisa login sampai superadmin unsuspend
```

### Use Case 5: Bulk Import Admin Regional via CSV

```
1. Superadmin siapkan file admins_bandung.csv:
   user_id,email,name
   uuid-001,andi@email.com,Andi
   uuid-002,budi@email.com,Budi
   uuid-003,citra@email.com,Citra

2. Buka User Management → [Bulk Assign] → [Upload CSV]
3. Upload file, pilih Role = "Admin Region", Region = "Bandung"
4. System validasi CSV:
   - Format valid ✅
   - 3 user ditemukan ✅
5. Preview hasil: 3 user akan di-assign
6. Konfirmasi → proses
7. Hasil: 3/3 berhasil
8. Download laporan operasi (CSV)
```

### Use Case 6: Role Temporary — Expired Otomatis

```
1. Superadmin assign role Admin Region Bali ke Irfan
   expired_at = 2026-07-01 (ganti admin setelah event musim panas)

2. Tanggal 2026-07-01 02:00 WIB:
   - Cron job jalan: scan user_roles WHERE expired_at <= NOW() AND is_active = true
   - Temukan: Irfan, role Admin Region Bali, expired
   - Set is_active = false
   - Log audit: role.expired
   - Kirim notifikasi ke Irfan: "Role Admin Region Bali kamu telah berakhir"

3. Irfan kehilangan akses admin Bali secara otomatis
```

---

## 13. Business Rules & Constraints

### Role Assignment

1. **Tidak boleh assign role yang sama di scope yang sama dua kali** (unique constraint)
2. **Usergod tidak bisa di-assign via UI** — hanya lewat database/developer
3. **Role `member` (system) tidak bisa di-revoke** — ini adalah role default semua user terdaftar
4. **Admin region tidak bisa assign ke luar region sendiri**
5. **Admin region tidak bisa assign role superadmin**
6. **Kalau user di-suspend, semua session aktif harus di-invalidate** (JWT blacklist atau revoke refresh token)

### Data Integrity

7. **Soft delete akun tidak menghapus konten** — konten yang dibuat user tetap ada, hanya akun yang tidak bisa diakses
8. **Revoke role menggunakan soft delete** (`is_active = false`) bukan hard delete — untuk keperluan audit
9. **Audit log immutable** — tidak ada endpoint untuk edit atau hapus audit log

### Scope Validation

10. **Saat assign dengan scope region → validasi region_id benar-benar ada**
11. **Saat assign dengan scope community → validasi community_id benar-benar ada**
12. **Admin region yang di-revoke jabatannya → akses backoffice region langsung hilang** (tergantung implementasi session/token)

### Bulk Operation

13. **Maksimal 100 user per bulk assign (via JSON)**, 500 user via CSV
14. **Error pada sebagian user tidak menghentikan keseluruhan operasi** (partial success)
15. **Setiap bulk operation menghasilkan `operation_id`** yang bisa dipakai untuk tracking dan download laporan

---

## 14. Edge Cases

### User Berganti Region, Role Admin Tidak Otomatis Pindah

Jika seorang user yang punya role Admin Region Jakarta pindah ke Region Surabaya (update region membership), role admin Jakarta-nya **tidak otomatis hilang**. Superadmin harus revoke manual.

### User di-Suspend tapi Punya Role Admin

Jika admin region di-suspend oleh superadmin:
- Role admin tetap ada di DB (`is_active = true` di user_roles)
- Tapi user tidak bisa login → de facto tidak bisa gunakan akses admin
- Saat unsuspend, akses admin langsung pulih tanpa perlu re-assign

### Bulk Assign — User Sudah Punya Role yang Sama

User yang sudah punya role yang di-assign di scope yang sama → **di-skip** (bukan error). Total `skipped_count` dilaporkan di response.

### Expired Role — User Tidak Tahu

Jika role expired dan cron job berhasil revoke, notifikasi email dikirim ke user. Jika cron job gagal (downtime), role akan expired pada cron run berikutnya. Tidak ada SLA real-time untuk expired role.

### Admin Assign Admin Lain di Region Sendiri — Circular

Admin A assign Admin B di region yang sama. Admin B sekarang bisa assign admin lain juga. Tidak ada pembatasan "siapa yang assign pertama kali" — semua admin di region yang sama punya kekuatan yang setara untuk user management di region tersebut.

### Bulk CSV — Email Tidak Terdaftar

Jika CSV menggunakan email dan email tersebut belum terdaftar di platform:
- Row tersebut di-error (`user_not_found`)
- Proses dilanjutkan untuk baris lain
- Laporan mencatat email yang tidak ditemukan

---

## 15. Integrasi dengan Modul Lain

| Modul | Titik Integrasi |
|---|---|
| **Auth** | Invalidate session saat suspend; reset password link |
| **Role & Permission** | Sumber data role & permission master; `user_roles` table |
| **Region** | Validasi `scope_id` saat assign role regional; filter user per region |
| **Subscription** | Manual plan change oleh superadmin; lihat status subscription di detail user |
| **Notification (FCM)** | Kirim notif saat: role assigned/revoked/expired, status changed, password reset |
| **Community** | Informasi community role tampil di detail user (read-only di modul ini) |
| **Audit Log** | Semua write operation di-log untuk compliance dan traceability |

---

*Dokumen ini adalah rules & use cases untuk modul User Management. API spec untuk modul ini ada di `API_SPEC_USER_MANAGEMENT_BACKOFFICE.md` dan `API_SPEC_BULK_ROLE_ASSIGNMENT_BACKOFFICE.md`.*
