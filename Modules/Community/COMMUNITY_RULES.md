# Community System — Rules & Use Cases

Dokumen ini menjelaskan aturan bisnis modul Community dan bagaimana user berinteraksi dengan fitur komunitas. Fokus pada **siapa bisa apa**, **kapan**, dan **kenapa** — bukan detail teknis. Untuk detail teknis lihat `COMMUNITY_DB_SCHEMA.md` serta API spec mobile & backoffice.

---

## Daftar Isi

1. [Overview Konsep](#overview-konsep)
2. [Entitas Utama](#entitas-utama)
3. [Hubungan dengan Modul Lain](#hubungan-dengan-modul-lain)
4. [Visibility: Public vs Private](#visibility-public-vs-private)
5. [Region: Opsional](#region-opsional)
6. [Siapa Bisa Apa](#siapa-bisa-apa)
7. [Alur Utama](#alur-utama)
8. [Status & Transition](#status--transition)
9. [Edge Cases & Rules](#edge-cases--rules)
10. [Keputusan yang Masih Terbuka](#keputusan-yang-masih-terbuka)

---

## Overview Konsep

Community adalah grup yang dibuat oleh **Member Pro**. Pembuat komunitas otomatis menjadi **leader**. Di dalam komunitas terdapat anggota, konten/postingan, dan moderasi.

Tiga prinsip dasar:

- **Membuat komunitas** hanya untuk Member Pro (gated oleh benefit `create_community`). **Bergabung dan posting** terbuka untuk semua member (Standard & Pro).
- **Role bersifat global & fixed** (`leader`, `moderator`, `member`), tetapi **permission per role bisa berbeda antar komunitas**. Aturan ini ditangani modul Role-Permission, bukan modul Community.
- Komunitas **boleh** terikat region, tapi tidak wajib (`region_id` opsional).

---

## Entitas Utama

### 1. Community

Entitas inti yang merepresentasikan satu komunitas.

| Atribut                         | Keterangan                                                                              |
| ------------------------------- | --------------------------------------------------------------------------------------- |
| `name`, `description`, `avatar` | Identitas komunitas                                                                     |
| `owner_id`                      | User yang saat ini menjadi leader/pemilik                                               |
| `category_id`                   | **Wajib** — satu kategori komunitas (FK ke `community_categories`); dikelola superadmin |
| `visibility`                    | `public` atau `private`                                                                 |
| `region_id`                     | Opsional — komunitas bisa lokal (terikat region) atau global                            |
| `status`                        | `active`, `suspended`, `archived`, `orphaned`                                           |
| `member_count`                  | Jumlah anggota aktif (denormalized untuk performa)                                      |

### 2. CommunityMember

Menyimpan **status keanggotaan** seorang user di sebuah komunitas. Berbeda dari `user_roles` (yang menyimpan role). Satu user bisa anggota banyak komunitas.

| Atribut                   | Keterangan                                        |
| ------------------------- | ------------------------------------------------- |
| `user_id`, `community_id` | Pasangan keanggotaan                              |
| `status`                  | `active`, `pending` (menunggu approval), `banned` |
| `joined_at`               | Waktu bergabung (saat status menjadi `active`)    |

> **Catatan penting:** Role (leader/moderator/member) TIDAK disimpan di sini. Role ada di `user_roles` dengan `scope_type='community'` dan `scope_id=community_id`. Tabel ini hanya status keanggotaan.

### 3. CommunityJoinRequest

Hanya relevan untuk komunitas **private**. Menyimpan permintaan bergabung yang menunggu persetujuan leader/moderator.

| Atribut                   | Keterangan                        |
| ------------------------- | --------------------------------- |
| `user_id`, `community_id` | Pemohon                           |
| `status`                  | `pending`, `approved`, `rejected` |
| `message`                 | Pesan opsional dari pemohon       |
| `reviewed_by`             | Leader/moderator yang memproses   |

### 4. CommunityPost

Konten yang dipublish dalam komunitas (feed).

| Atribut                     | Keterangan                                                                                                                                                 |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `community_id`, `author_id` | Konteks & penulis                                                                                                                                          |
| `content`                   | Teks postingan — **plain text** (tanpa markdown/HTML). Client boleh auto-render URL/`@mention`/`#hashtag` saat tampil, tapi yang disimpan tetap plain text |
| `media`                     | Lampiran gambar, maksimal **10** (array of object: url, thumb_url, width, height, order)                                                                   |
| `status`                    | `published`, `removed` (dihapus moderasi)                                                                                                                  |

### 5. Interaksi Post (Like, Comment, Save, Share)

Interaksi sosial atas sebuah post. Semua masuk **Fase 1** kecuali repost internal.

| Entitas                | Fungsi           | Aturan                                                                     |
| ---------------------- | ---------------- | -------------------------------------------------------------------------- |
| `CommunityPostLike`    | Like/unlike post | Toggle — satu user maks satu like per post                                 |
| `CommunityPostComment` | Komentar         | Plain text, **maks 1 level reply** (reply ke reply ditolak)                |
| `CommunityPostSave`    | Bookmark post    | Toggle — tersimpan privat di list user                                     |
| Share                  | Bagikan keluar   | Fase 1 = **deep link eksternal** (mis. WhatsApp); repost internal = Fase 2 |

Counter (`like_count`, `comment_count`, `save_count`, `share_count`) disimpan denormalized di post agar feed cepat.

> Hak berinteraksi mengikuti keanggotaan: hanya **anggota aktif** komunitas yang bisa like/comment/save. Komunitas private hanya bisa diinteraksi oleh anggotanya.

> Komentar (`community_post_comments`), like (`community_post_likes`), dan save (`community_post_saves`) termasuk **Fase 1** (lihat entitas #5). Yang ditunda ke **Fase 2**: reaction beragam (selain like), reply nested >1 level, dan **repost internal**.

---

## Hubungan dengan Modul Lain

Modul Community **mengkonsumsi** modul lain, bukan menduplikasi:

| Kebutuhan                                | Ditangani oleh                                 | Bagaimana Community memakainya             |
| ---------------------------------------- | ---------------------------------------------- | ------------------------------------------ |
| Role leader/moderator/member             | Role-Permission (global, fixed)                | Reference via `user_roles` scope=community |
| Permission per komunitas                 | `community_role_permissions` + template        | Saat create, copy dari template            |
| Izin membuat komunitas                   | Plan-Subscription (benefit `create_community`) | Cek benefit sebelum create                 |
| Statistik komunitas                      | Plan-Subscription (benefit `view_analytics`)   | Tampilkan analytics untuk Pro              |
| Notifikasi (post baru, member join/left) | Notification                                   | Community memancarkan event                |
| Region (opsional)                        | Region module                                  | `region_id` nullable, reference saja       |

---

## Visibility: Public vs Private

| Aspek                      | Public             | Private                                   |
| -------------------------- | ------------------ | ----------------------------------------- |
| Muncul di discovery/browse | ✅ Ya               | ✅ Ya (terlihat, tapi terkunci)            |
| Cara bergabung             | Auto-join langsung | Kirim join request → approval             |
| Lihat konten sebelum join  | ✅ Boleh            | ❌ Hanya anggota                           |
| Persetujuan                | Tidak perlu        | Leader/moderator (perlu `manage_members`) |

---

## Region: Opsional

`region_id` bersifat **nullable**:

- **Tanpa region** → komunitas global, semua member bisa menemukan & join.
- **Dengan region** → komunitas lokal; tetap bisa di-browse semua orang (region hanya label & filter, mengikuti pola modul Event), bukan pembatas akses, kecuali diputuskan lain.

> Region di sini berfungsi sebagai **label & filter**, sejalan dengan klarifikasi region pada modul Event. Tidak ada pembatasan akses berdasarkan region membership user.

---

## Siapa Bisa Apa

### Member Standard

| Aksi                              | Bisa?   | Catatan                                                      |
| --------------------------------- | ------- | ------------------------------------------------------------ |
| Browse & cari komunitas           | ✅ Ya    | Public & private (private terkunci); bisa filter by category |
| Join komunitas public             | ✅ Ya    | Auto-join                                                    |
| Request join komunitas private    | ✅ Ya    | Menunggu approval                                            |
| Posting di komunitas yang diikuti | ✅ Ya    | Selama punya permission `post_content`                       |
| Leave komunitas                   | ✅ Ya    |                                                              |
| **Membuat komunitas**             | ❌ Tidak | Perlu benefit `create_community` (Pro)                       |

### Member Pro

| Aksi                         | Bisa? | Catatan                  |
| ---------------------------- | ----- | ------------------------ |
| Semua aksi Member Standard   | ✅ Ya  |                          |
| **Membuat komunitas**        | ✅ Ya  | Auto jadi leader         |
| Mengelola komunitas miliknya | ✅ Ya  | Sesuai permission leader |
| Lihat analytics komunitas    | ✅ Ya  | Benefit `view_analytics` |

### Leader (dalam komunitasnya)

| Aksi                          | Bisa? | Catatan                                   |
| ----------------------------- | ----- | ----------------------------------------- |
| Edit profil komunitas         | ✅ Ya  | Nama, deskripsi, avatar, visibility       |
| Approve/reject join request   | ✅ Ya  | Komunitas private                         |
| Promote member → moderator    | ✅ Ya  | Perlu `manage_members`                    |
| Kick / ban member             | ✅ Ya  | Perlu `manage_members`                    |
| Customize permission per role | ✅ Ya  | Via Role-Permission bulk-assign           |
| Moderasi & hapus konten       | ✅ Ya  | Perlu `moderate_posts` / `delete_content` |
| Transfer ownership            | ✅ Ya  | Wajib sebelum resign                      |
| Hapus komunitas               | ✅ Ya  | Cascade cleanup                           |

### Moderator (dalam komunitasnya)

| Aksi                    | Bisa?        | Catatan                                |
| ----------------------- | ------------ | -------------------------------------- |
| Moderasi & hapus konten | ✅ Ya*        | *Sesuai permission yang di-set leader  |
| Approve join request    | ✅ Ya*        | *Jika `manage_members` diberikan       |
| Manage member           | ⚠️ Tergantung | Permission diatur leader per komunitas |

### Superadmin (Backoffice)

| Aksi                         | Bisa? | Catatan                     |
| ---------------------------- | ----- | --------------------------- |
| Lihat semua komunitas        | ✅ Ya  |                             |
| Suspend / archive komunitas  | ✅ Ya  | Pelanggaran kebijakan       |
| Hapus komunitas              | ✅ Ya  |                             |
| Tangani komunitas `orphaned` | ✅ Ya  | Assign owner baru / archive |
| Moderasi konten global       | ✅ Ya  |                             |

### Guest (belum login)

Tidak ada akses. Harus login terlebih dahulu.

---

## Alur Utama

### Alur 1: Member Pro Membuat Komunitas

```
1. Pro user → POST /communities { name, description, category_id, visibility, region_id? }
2. Backend:
   - Cek benefit create_community ✅
   - Validasi category_id ada dan aktif ✅
   - Buat entry communities (status=active, owner_id=user)
   - Auto-assign user_roles: leader, scope=community, scope_id=baru
   - Copy permission dari community_role_permissions_template
   - Insert community_members (user, status=active)
3. Response: community object
```

### Alur 2: Member Join Komunitas Public

```
1. Member → POST /communities/{id}/join
2. Backend:
   - Cek visibility = public ✅
   - Insert community_members (status=active)
   - Assign user_roles: member, scope=community
   - member_count += 1
   - Emit event member_joined → Notification
3. Response: success
```

### Alur 3: Member Request Join Komunitas Private

```
1. Member → POST /communities/{id}/join { message? }
2. Backend:
   - Cek visibility = private
   - Insert community_join_requests (status=pending)
   - Notify leader/moderator
3. Leader → POST /communities/{id}/requests/{rid}/approve
4. Backend:
   - Update request status=approved
   - Insert community_members (status=active) + user_roles member
   - Emit member_joined
```

### Alur 4: Member Leave Komunitas

```
1. Member → DELETE /communities/{id}/membership
2. Backend:
   - Hapus community_members row (atau set inactive)
   - Hapus user_roles untuk scope ini
   - member_count -= 1
   - Emit member_left
- Edge: jika requester adalah leader → tolak, harus transfer ownership dulu
```

### Alur 5: Leader Transfer Ownership

```
1. Leader → POST /communities/{id}/transfer-ownership { new_owner_id }
2. Backend:
   - Cek new_owner adalah member aktif komunitas ✅
   - Demote leader lama → member (atau keluar)
   - Promote new_owner → leader, update communities.owner_id
3. Response: success
```

### Alur 6: Hapus Komunitas

```
1. Leader/Superadmin → DELETE /communities/{id}
2. Backend (transactional):
   - Hapus/soft-delete community_posts
   - Hapus community_members
   - Hapus user_roles (scope=community ini)
   - Hapus community_role_permissions (scope ini)
   - Set communities.status = archived (atau hard delete)
   - Emit event ke anggota
```

---

## Status & Transition

### Community Status Flow

```
active ──suspend──> suspended ──unsuspend──> active
active ──archive──> archived
active ──owner hapus akun──> orphaned ──superadmin assign owner──> active
                                       └──superadmin archive──> archived
```

### CommunityMember Status Flow

```
(request private) pending ──approve──> active
                  pending ──reject──> (dihapus)
active ──ban──> banned ──unban──> active
active ──leave/kick──> (dihapus)
```

### JoinRequest Status Flow

```
pending ──approve──> approved
pending ──reject──> rejected
```

---

## Edge Cases & Rules

### Rule 1: Leader Tidak Bisa Leave Tanpa Transfer
Leader wajib transfer ownership ke member lain sebelum keluar. Jika leader menghapus akun, komunitas masuk status `orphaned` untuk ditangani superadmin.

### Rule 2: Satu Komunitas Minimal Punya Satu Leader
Sistem tidak boleh membiarkan komunitas `active` tanpa leader. `orphaned` adalah state sementara.

### Rule 3: Banned Member Tidak Bisa Re-join
Member dengan status `banned` tidak bisa join ulang sampai di-unban oleh leader/moderator.

### Rule 4: Hapus Komunitas = Cleanup Penuh
Penghapusan harus membersihkan `user_roles`, `community_members`, `community_role_permissions`, dan konten — dalam satu transaksi.

### Rule 5: Permission Mengikuti Role-Permission Module
Hak posting/moderasi/hapus diputuskan oleh `community_role_permissions`. Modul Community hanya memanggil permission check, tidak menyimpan logika permission sendiri.

---

## Keputusan yang Masih Terbuka

Asumsi default berikut dipakai dalam dokumen ini dan **mudah diubah**:

| #   | Topik                                 | Asumsi Sementara                                                                                                                                                 |
| --- | ------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Batas jumlah komunitas per Member Pro | Tidak dibatasi (bisa dijadikan benefit terukur nanti)                                                                                                            |
| 2   | Tipe konten post                      | **Plain text** + gambar maks 10 (diputuskan). Interaksi Fase 1: like, comment 1-level, save, share link. Fase 2: reaction beragam, nested reply, repost internal |
| 3   | Region sebagai pembatas akses         | Tidak — region hanya label & filter (ikut pola Event)                                                                                                            |
| 4   | Override permission per-user          | Belum didukung (hanya per-role per-komunitas)                                                                                                                    |
| 5   | Notifikasi perubahan permission       | Belum diputuskan                                                                                                                                                 |
| 6   | Pengelolaan daftar category           | Dikelola superadmin via backoffice (CRUD). Leader hanya memilih dari daftar yang tersedia saat create/edit komunitas                                             |

---

*Dokumen ini adalah hasil breakdown awal modul Community. Untuk skema database lihat `COMMUNITY_DB_SCHEMA.md`.*