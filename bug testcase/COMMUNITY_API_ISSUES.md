# Issue: Community API — 403 permission, PascalCase response, route hilang

- **Modul**: Community (`/api/v1/mobile/communities/*`) + sub-modul Announcement/Schedule/Invite
- **Severity**: High — fitur kelola komunitas (join request, announcement, schedule, invite) tidak bisa dipakai sama sekali dari mobile
- **Status**: 🔴 Open — menunggu fix backend
- **Ditemukan**: 13 Jul 2026, sweep penuh via curl di production (`k-forum-api.yubicom.co.id`), akun QA `test_member`
- **Konteks**: permission sudah ditambahkan di backoffice, tapi beberapa endpoint tetap `403`

---

## Metode uji

Login sebagai `test_member` → sweep 40+ endpoint mengikuti spec:
`API_SPEC_COMMUNITY_MOBILE.md`, `API_SPEC_COMMUNITY_ANNOUNCEMENT_SCHEDULE_MOBILE.md`, `API_SPEC_COMMUNITY_INVITE_MOBILE.md`.

Dua konteks komunitas:
1. **Komunitas existing** di mana test_member = **leader** (`K-Drama & Film Review Indonesia`, `1b4d19cb-…`).
2. **Komunitas baru** yang dibuat saat sweep (creator otomatis leader), dihapus lagi setelah selesai.

Pada keduanya, `GET /communities/{id}` melaporkan:

```
my_role       = leader
my_permissions = [manage_community, manage_members, post_content, delete_content, moderate_posts]
```

Spec menyatakan `my_permissions` adalah *source of truth* untuk aksi user (Notes #7).

---

## Issue 1 — 403 padahal permission ada (BUG UTAMA)

Semua endpoint berikut menolak **leader** yang `my_permissions`-nya memuat `manage_members`:

| Endpoint | Permission per spec | Hasil |
|---|---|---|
| `GET /{cid}/join-requests` | `manage_members` | **403** ❌ (di kedua komunitas) |
| `GET /{cid}/invitable-users` | `manage_members` | **403** ❌ |
| `GET /{cid}/invitations` | `manage_members` | **403** ❌ |
| `POST /{cid}/invite-links` | `manage_members` | **403** ❌ |
| `GET /{cid}/invite-links` | `manage_members` | **403** ❌ |
| `POST /{cid}/announcements` | `manage_community_announcement` | **403** ❌ |
| `POST /{cid}/schedule` | `manage_community_schedule` | **403** ❌ |

Body error seragam: `{"error_code":"ERR_FORBIDDEN","errors":"Anda tidak memiliki izin untuk melakukan tindakan ini."}`

### Akar masalah yang teridentifikasi

1. **Template role leader tidak memuat permission baru.** `my_permissions` leader (termasuk pada komunitas yang BARU dibuat) tidak berisi `manage_community_announcement` dan `manage_community_schedule` — dua permission yang didefinisikan spec announcement/schedule. Artinya saat create community, permission yang disalin dari template masih daftar lama. → announcement/schedule 403 itu "konsisten" dengan datanya, tapi datanya yang salah: **tambahkan kedua permission itu ke template role leader (dan moderator sesuai kebijakan), plus backfill komunitas yang sudah ada**.
2. **Join-requests & seluruh invite menolak walau `manage_members` ADA** di `my_permissions`. Ini murni bug pengecekan: middleware endpoint tsb kemungkinan memeriksa nama permission yang berbeda (mis. `manage_join_requests` / `manage_invites`) atau scope yang salah. → samakan dengan spec: `manage_members`.

### Dampak

Fitur kelola komunitas dari mobile mati total: leader tidak bisa melihat/approve join request komunitas private, tidak bisa mengundang anggota, tidak bisa membuat share link, tidak bisa membuat pengumuman/agenda. UI mobile-nya sudah jadi dan mengikuti `my_permissions` — begitu backend benar, fitur langsung hidup tanpa perubahan app.

### Acceptance

- [ ] Leader komunitas baru punya `my_permissions` berisi `manage_community_announcement` & `manage_community_schedule`; komunitas lama di-backfill.
- [ ] `GET /{cid}/join-requests` = 200 untuk pemilik `manage_members`.
- [ ] 5 endpoint invite/link = 200/201 untuk pemilik `manage_members`.
- [ ] `POST /announcements` & `POST /schedule` = 201 untuk pemilik permission masing-masing.

---

## Issue 2 — Response CREATE memakai PascalCase (inkonsisten dgn spec & GET)

`POST /{cid}/posts` dan `POST /posts/{id}/comments` membalas key **PascalCase** ala Go struct, sedangkan semua GET (feed, detail) membalas snake_case sesuai spec:

```json
// POST /posts → 201 (aktual)
{ "data": { "ID": "…", "CommunityID": "…", "Author": { "ID": "…", "Name": "" },
            "Content": "…", "LikeCount": 0, "IsLiked": false, "CreatedAt": "…" } }
```

- Kemungkinan penyebab: struct response create belum diberi json tag.
- Catatan: `Author.Name` juga selalu `""` pada response create (bonus bug kecil).
- **Mobile sudah workaround** (13 Jul 2026): normalizer `snakeCaseKeys` di `community_model.dart`, diterapkan di `PostModel`/`CommentModel` + unit test payload asli. Tetap mohon dirapikan di server supaya konsisten dengan spec.

## Issue 3 — Route `POST /communities/media/presign` tidak ada (404)

Spec #19 mendaftarkan dua route presign. Yang live hanya `POST /communities/media/post/presign` (201 ✅); route umum `POST /communities/media/presign` → **404**. Mobile memakai yang live, jadi tidak blocking — tapi spec & implementasi harus disamakan (implement atau hapus dari spec). Endpoint confirm (#20) & delete media (#21) belum diverifikasi.

### 3b — Upload media untuk Announcement juga tidak punya route (13 Jul 2026)

Spec announcement (Note #4) menyuruh upload via `POST /api/v1/mobile/media/upload` (context `community_announcement`) sebelum kirim `media[]` di `POST /announcements`. Hasil probe (semua dengan Bearer test_member, body `{"content_type":"image/jpeg"}`):

| Route | Hasil |
|---|---|
| `POST /mobile/media/upload` | **404** ❌ |
| `POST /communities/media/announcement/presign` | **404** ❌ |
| `POST /communities/media/post/presign` | 201 ✅ (satu-satunya yang live) |

**Workaround mobile**: UI announcement memakai presign **post** untuk lampiran pengumuman (key `s3:` dari presign dikirim di `media[].url` sesuai `MEDIA_UPLOAD_SPEC.md`). Konsekuensi: objek tersimpan di folder `community/posts/` dan lifecycle tracking-nya tercatat sebagai media post. Mohon backend: sediakan route upload sesuai spec announcement (atau update spec agar resmi memakai presign post), sekaligus konfirmasi `POST /announcements` menerima `media[]`. Verifikasi end-to-end create-with-media belum bisa dilakukan dari mobile karena `POST /announcements` masih 403 (Issue 1).

## Issue 4 — Bentuk pagination tidak sesuai spec (minor)

Spec: `pagination { limit, offset, total, has_next, has_prev }`. Aktual: `meta { current_page, per_page, total, total_pages }`. Mobile sudah membaca bentuk `meta`, jadi tidak blocking — cukup samakan spec-nya dengan kenyataan (atau sebaliknya, sekali saja, lalu konsisten).

---

## Yang sudah lolos (verified 200/201 ✅)

categories, browse (+filter), mine, detail (my_role/my_permissions terisi), members, join (409 saat sudah member — benar), create/update/delete community (create TIDAK dibatasi Pro saat ini — konfirmasi apakah memang begitu?), feed, post detail, create/delete post, comments + replies + delete comment, like/unlike, save/unsave, share (share_url OK), saved posts, presign post media, `GET /announcements`, `GET /schedule?from&to`, `GET /me/community-invitations`.

> ⚠️ Catatan kecil dari daftar di atas: spec #4 bilang create community butuh benefit Pro (`create_community`), tapi `test_member` (bukan Pro) berhasil membuat komunitas (201). Kalau gating Pro memang ditunda, update spec; kalau tidak, ini bug juga.

## Referensi

- Spec: `docs/api_spec/API_SPEC_COMMUNITY_MOBILE.md`, `docs/modules/Community/API_SPEC_COMMUNITY_ANNOUNCEMENT_SCHEDULE_MOBILE.md`, `docs/modules/Community/API_SPEC_COMMUNITY_INVITE_MOBILE.md`
- Workaround mobile (PascalCase): `lib/features/community/data/models/community_model.dart` → `snakeCaseKeys`
- Data uji dibersihkan: komunitas QA & post/komentar uji sudah dihapus.
