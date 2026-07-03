# Community Invite & Share Link — Rules & Use Cases (v1.0)

Dokumen ini menjelaskan aturan bisnis dua cara baru bergabung ke komunitas di modul **Community**: **Undangan (invite in-app)** dan **Share Link / Join Code**. Melengkapi cara gabung yang sudah ada (public auto-join, private join-request).

Untuk skema lihat `COMMUNITY_INVITE_DB_SCHEMA.md`. Untuk endpoint lihat `API_SPEC_COMMUNITY_INVITE_MOBILE.md`. Aturan komunitas inti tetap di `COMMUNITY_RULES.md`.

---

## Daftar Isi

1. [Overview](#overview)
2. [Cara Gabung Komunitas (Lengkap)](#cara-gabung-komunitas-lengkap)
3. [Fitur A — Undangan (Invite In-App)](#fitur-a--undangan-invite-in-app)
4. [Fitur B — Share Link / Join Code](#fitur-b--share-link--join-code)
5. [Permission](#permission)
6. [Siapa Bisa Apa](#siapa-bisa-apa)
7. [Alur Utama](#alur-utama)
8. [Status & Transition](#status--transition)
9. [Hubungan dengan Modul Lain](#hubungan-dengan-modul-lain)
10. [Edge Cases & Rules](#edge-cases--rules)
11. [Phasing](#phasing)
12. [Keputusan yang Masih Terbuka](#keputusan-yang-masih-terbuka)

---

## Overview

Dua mekanisme yang saling melengkapi:

| Fitur | Sifat | Target | Kapan dipakai |
|---|---|---|---|
| **Undangan (invite)** | Ditargetkan | Satu user terdaftar tertentu | Leader ngajak orang spesifik |
| **Share Link** | Terbuka | Siapa saja yang pegang link | Sebar di grup WA/IG, viral |

Prinsip:
- **Undangan bersifat personal & butuh persetujuan target** (target accept/reject). Undangan ke komunitas private **bypass approval** — leader sudah percaya, jadi tinggal target setuju.
- **Share link bisa langsung nge-join** (default) atau **tetap lewat approval** (opsional per-link, via flag `requires_approval`).
- **Keduanya di-gate `manage_members`** (permission yang sudah ada untuk kelola anggota). Tidak ada permission key baru.

---

## Cara Gabung Komunitas (Lengkap)

Setelah fitur ini, total ada 4 jalur masuk:

| Jalur | Komunitas Public | Komunitas Private |
|---|---|---|
| **Browse → Join** | Auto-join langsung | Kirim join-request → approval |
| **Undangan (invite)** | Accept → langsung join | Accept → **bypass approval**, langsung join |
| **Share link** (`requires_approval=false`) | Pegang link → langsung join | Pegang link → **bypass approval**, langsung join |
| **Share link** (`requires_approval=true`) | Link → langsung join¹ | Link → buat join-request → approval |

> ¹ Untuk komunitas **public**, `requires_approval` diabaikan — public tidak punya alur approval, jadi selalu langsung join.

---

## Fitur A — Undangan (Invite In-App)

### Konsep

Leader/moderator (pemegang `manage_members`) mengundang **user terdaftar** dengan mencari username/nama. Undangan muncul sebagai notifikasi + daftar undangan di app target. Target **accept** → jadi anggota; **reject** → undangan ditutup.

### Entitas: CommunityInvitation

| Atribut | Keterangan |
|---|---|
| `community_id`, `invitee_id`, `invited_by` | Komunitas, yang diundang, yang mengundang |
| `message` | Pesan opsional dari pengundang |
| `status` | `pending` \| `accepted` \| `rejected` \| `expired` \| `cancelled` |
| `expires_at` | Nullable — default **7 hari** dari dibuat; lewat itu → `expired` |
| `responded_at` | Waktu accept/reject |

### Aturan

- **Hanya user terdaftar** yang bisa diundang (in-app; invite via email ditunda — lihat Phasing).
- **Accept = bypass approval**, termasuk untuk komunitas private. Tidak membuat join-request.
- **Satu undangan pending per (komunitas, invitee).** Undangan ulang saat masih pending → tolak/ganti yang lama.
- **Tidak bisa mengundang** user yang sudah anggota aktif (`409`), atau yang berstatus `banned` di komunitas itu.
- **Undangan kedaluwarsa** otomatis setelah `expires_at`; target tidak bisa accept lagi (bisa diundang ulang).
- **Pengundang/leader bisa cancel** undangan yang masih pending.

---

## Fitur B — Share Link / Join Code

### Konsep

Link/kode yang bisa disebar. Siapa pun yang membukanya (dan login) bisa gabung sesuai aturan link. Ala "invite link" — cocok disebar di grup chat.

### Entitas: CommunityInviteLink

| Atribut | Keterangan |
|---|---|
| `community_id`, `created_by` | Komunitas & pembuat |
| `code` | Token unik pendek (dipakai di URL, mis. `kai.app/c/join/{code}`) |
| `requires_approval` | `false` (default) = pegang link langsung join; `true` = buat join-request dulu |
| `expires_at` | Nullable — kalau diisi, link mati setelah tanggal itu |
| `max_uses` | Nullable — kalau diisi, link mati setelah dipakai sekian kali |
| `use_count` | Berapa kali sudah berhasil dipakai (join) |
| `is_active` | Revocable — leader bisa matikan kapan saja |

### Aturan

- **Default paling simpel:** `requires_approval=false`, `expires_at=null`, `max_uses=null` → link abadi, langsung-join, sampai dicabut.
- **Link mati** bila salah satu: `is_active=false` (dicabut), lewat `expires_at`, atau `use_count >= max_uses`.
- **Public community:** `requires_approval` diabaikan (selalu langsung join).
- **Banned user** tidak bisa pakai link (ditolak walau link valid).
- **User yang sudah anggota** yang pakai link → no-op (tidak nambah `use_count`, tidak error keras — arahkan ke komunitas).
- **Beberapa link aktif sekaligus diperbolehkan** (mis. satu abadi + satu berkuota untuk event tertentu).
- **`use_count` naik hanya saat join berhasil** (bukan saat link dibuka). Redemption dicatat untuk audit & mencegah double-count.

---

## Permission

Tidak ada permission key baru. Kedua fitur di-gate oleh **`manage_members`** yang sudah ada:

| Aksi | Perlu |
|---|---|
| Kirim/cancel undangan | `manage_members` |
| Buat/cabut/lihat share link | `manage_members` |
| Accept/reject undangan yang ditujukan ke diri sendiri | — (siapa pun yang diundang) |
| Redeem share link | — (user login mana pun, tunduk aturan link) |

> Leader dapat `manage_members` default. Moderator bila di-grant leader (pola sama seperti approve/kick/ban).

---

## Siapa Bisa Apa

### Leader / Moderator (dengan `manage_members`)

| Aksi | Bisa? |
|---|---|
| Cari & undang user terdaftar | ✅ |
| Cancel undangan pending | ✅ |
| Lihat daftar undangan komunitas | ✅ |
| Buat share link (atur approval/expiry/kuota) | ✅ |
| Cabut/lihat share link & statistik pakai | ✅ |

### Member (yang diundang / penerima link)

| Aksi | Bisa? |
|---|---|
| Lihat undangan yang ditujukan ke dirinya | ✅ |
| Accept / reject undangan | ✅ |
| Redeem share link (join / request) | ✅ (kecuali banned) |
| Undang orang lain / buat link | ❌ (butuh `manage_members`) |

### Superadmin (Backoffice)

Bisa melihat & mencabut undangan/link komunitas mana pun sebagai bagian moderasi (override). Detail endpoint menyusul bila diperlukan.

---

## Alur Utama

### Alur 1: Leader Undang User (In-App)

```
1. Leader → GET /communities/{id}/invitable-users?q=andi   (cari user)
2. Leader → POST /communities/{id}/invitations { invitee_id, message? }
3. Backend:
   - Cek permission manage_members ✅
   - Cek invitee bukan anggota aktif & bukan banned ✅
   - Cek tak ada undangan pending duplikat ✅
   - Insert community_invitations (status=pending, expires_at=now+7d)
   - Emit event community_invitation_received → Notification (ke invitee)
4. Response: invitation object
```

### Alur 2: Target Accept Undangan

```
1. Invitee → GET /me/community-invitations         (lihat undangan masuk)
2. Invitee → POST /community-invitations/{id}/accept
3. Backend:
   - Cek status=pending & belum expired ✅
   - Insert community_members (status=active)   ← bypass approval walau private
   - Assign user_roles: member, scope=community
   - member_count += 1
   - Update invitation status=accepted, responded_at=now
   - Emit member_joined (+ opsional community_invitation_accepted ke pengundang)
4. Response: success + community object
```

### Alur 3: Leader Buat Share Link

```
1. Leader → POST /communities/{id}/invite-links
   { requires_approval?: false, expires_at?: null, max_uses?: null }
2. Backend:
   - Cek permission manage_members ✅
   - Generate code unik
   - Insert community_invite_links (is_active=true, use_count=0)
3. Response: { code, url: "https://kai.app/c/join/{code}", ... }
```

### Alur 4: User Redeem Share Link

```
1. User buka link → app resolve → POST /community-invite-links/{code}/redeem
2. Backend validasi link:
   - is_active=true, belum expired, use_count < max_uses (jika ada) ✅
   - user bukan banned di komunitas ✅
   - jika user sudah anggota → no-op, arahkan ke komunitas
3. Tentukan aksi:
   - Public OR requires_approval=false → join langsung:
       insert community_members(active) + user_roles member + member_count++
       + insert redemption, use_count++
   - Private AND requires_approval=true → buat community_join_requests(pending)
       (tidak menambah use_count sampai di-approve? → lihat Rule 6)
4. Response: { joined: true } atau { requested: true }
```

### Alur 5: Leader Cabut Link

```
1. Leader → DELETE /communities/{id}/invite-links/{link_id}   (set is_active=false)
2. Link langsung mati; redemption yang sudah terjadi tidak terpengaruh.
```

---

## Status & Transition

### Invitation

```
pending ──accept──> accepted        (jadi anggota)
pending ──reject──> rejected
pending ──cancel (leader)──> cancelled
pending ──(lewat expires_at)──> expired
```

### Invite Link (status turunan, bukan enum tunggal)

```
active  = is_active AND (expires_at IS NULL OR expires_at > now)
                    AND (max_uses IS NULL OR use_count < max_uses)
dead    = is_active=false  OR  expired  OR  kuota habis
```

---

## Hubungan dengan Modul Lain

| Kebutuhan | Ditangani oleh | Cara pakai |
|---|---|---|
| Hak mengundang / kelola link | Role-Permission | Permission check `manage_members`, scope=community |
| Keanggotaan & ban check | Community (core) | Reuse `community_members`, `community_join_requests` |
| Notifikasi undangan | Notification | Event `community_invitation_received`, `member_joined` |
| Cari user terdaftar | User Management | Endpoint search user (username/nama) |
| Deep link | App/Client | URL `kai.app/c/join/{code}` di-resolve app |

---

## Edge Cases & Rules

### Rule 1: Bypass approval hanya lewat undangan / link yang diizinkan
Join-request manual ke komunitas private **tetap** butuh approval. Bypass hanya terjadi via undangan (accept) atau share link `requires_approval=false`.

### Rule 2: Banned menutup semua jalur
User `banned` di sebuah komunitas tidak bisa diundang, tidak bisa accept undangan lama, dan tidak bisa redeem link — sampai di-unban.

### Rule 3: Idempoten untuk anggota existing
Accept undangan / redeem link oleh user yang sudah anggota = no-op sukses (arahkan ke komunitas), bukan error atau duplikat membership.

### Rule 4: Satu undangan pending per pasangan
`(community_id, invitee_id)` hanya boleh punya satu undangan `pending`. Undang ulang menggantikan/menyegarkan yang lama.

### Rule 5: Link mati = tegas
Redeem link mati → `410 Gone` (link kedaluwarsa/dicabut/kuota habis), pesan jelas ke user.

### Rule 6: `use_count` untuk link approval
Bila `requires_approval=true`, `use_count` naik saat **join-request dibuat** (bukan saat approve), supaya kuota membatasi jumlah pendaftar, bukan jumlah yang diterima. (Asumsi default — mudah diubah.)

### Rule 7: Cleanup komunitas
Hapus komunitas ikut membersihkan `community_invitations`, `community_invite_links`, dan `community_invite_link_redemptions` (FK cascade), konsisten dengan cleanup Community.

---

## Phasing

| Aspek | Phase 1 | Phase 2+ |
|---|---|---|
| Undangan | In-app targeted, accept/reject, expiry, cancel | Invite via **email** (narik non-user, pola Region), bulk invite |
| Share link | Buat/cabut, approval flag, expiry, max-uses, redemption audit | QR code, per-link label/analytics, member-generated link |
| Siapa boleh invite | `manage_members` (leader/mod) | Opsi: izinkan member biasa undang (viral) |
| Notifikasi | `invitation_received`, `member_joined` | `invitation_accepted` ke pengundang, reminder undangan hampir expired |

---

## Keputusan yang Masih Terbuka

Asumsi default, mudah diubah:

| # | Topik | Asumsi Sementara |
|---|---|---|
| 1 | Masa berlaku undangan | 7 hari (nullable) |
| 2 | Siapa boleh undang/buat link | `manage_members` saja (leader/mod). Member biasa = Phase 2 |
| 3 | `use_count` untuk link approval | Naik saat request dibuat (Rule 6) |
| 4 | Invite via email | Ditunda Phase 2 (butuh alur onboarding non-user, pola Region) |
| 5 | Batas jumlah link aktif per komunitas | Tidak dibatasi |
| 6 | Redemption tracking | Dicatat untuk audit & dedup (tabel terpisah) |

---

*Dokumen ini adalah breakdown fitur Undangan & Share Link modul Community. Skema di `COMMUNITY_INVITE_DB_SCHEMA.md`, endpoint di `API_SPEC_COMMUNITY_INVITE_MOBILE.md`.*
