# Issue: Community — avatar upload, invite, share link, moderasi, backoffice leader

- **Modul**: Community — avatar, invite (account name & link), share post, post moderation, backoffice leader/moderator access
- **Severity**: 🔴 Tinggi untuk Issue 1, 4, 5 (fitur inti gagal total) — 🟠 Sedang untuk Issue 2, 3, 6 — 🟢 Rendah untuk Issue 7 (disepakati low priority)
- **Status**: 🔴 Open — 7 issue di dokumen ini, semua root cause sudah ditemukan by code review
- **Ditemukan**: 22 Jul 2026, saat review user journey `04_COMMUNITY_JOURNEY.md`
- **Pelapor**: review manual (dev), dikonfirmasi via code review langsung ke `k-forum-api`, `k_forum`, dan `k-forum-backoffice`

---

## Issue 1 — Upload avatar komunitas gagal total

- **Repo**: `k_forum` (root cause utama) + `k-forum-api` (gap desain sekunder)

### Ringkasan

Endpoint presign avatar komunitas **ada** di backend, tapi mobile app memanggil URL yang salah — jadi 404 setiap kali, di kedua mode (create maupun edit).

### Root cause (code review)

Backend route yang benar (`internal/interfaces/http/router/router.go:815`):
```go
mobileCommunities.POST("/:community_id/media/avatar/presign", mobileCommunityHandler.GetCommunityAvatarPresignURL)
```
— butuh `:community_id` di path.

Mobile (`lib/features/community/presentation/screens/community_form_screen.dart:84`) hardcode:
```dart
presignEndpoint: '/mobile/communities/media/avatar/presign',
```
— **tidak ada `community_id` sama sekali**, tidak di-interpolasi dari mana pun (dicek juga di edit mode yang jelas-jelas sudah punya `community.id` — tetap tidak dipakai). Path ini tidak match route manapun di backend → selalu gagal.

**Gap sekunder di backend**: screen `community_form_screen.dart` dipakai untuk CREATE dan EDIT sekaligus, dan avatar dipilih **sebelum** community dibuat (belum ada `community_id`). Untuk kasus create, endpoint avatar yang ada memang tidak bisa dipakai karena desainnya scoped-per-community. Bandingkan dengan pola post/announcement image yang endpoint presign-nya **generic, tidak butuh ID entity** (`router.go:563`: `communities.POST("/media/post/presign", ...)` — tanpa `:community_id`) — supaya bisa dipakai sebelum entity-nya ada. Avatar tidak punya varian generic serupa.

### Yang diminta

**k_forum (mobile)**:
1. Untuk mode **edit**: perbaiki `presignEndpoint` supaya include `community_id` yang benar, sesuai route backend: `/mobile/communities/${communityId}/media/avatar/presign`.
2. Untuk mode **create**: butuh keputusan desain (lihat poin backend di bawah) sebelum bisa diperbaiki — flow "pilih avatar dulu sebelum submit" tidak bisa jalan tanpa endpoint generic.

**k-forum-api**:
1. Tambah endpoint presign avatar **generic** (tanpa `:community_id`), mengikuti pola `media/post/presign`, khusus untuk dipakai saat create community (avatar dikonfirmasi setelah community berhasil dibuat, sama seperti pola presign/confirm di §9 `CLAUDE.md`).

### Kriteria selesai (acceptance)

- [ ] Edit komunitas existing → ganti avatar → berhasil upload & tersimpan.
- [ ] Create komunitas baru → pilih avatar sebelum submit → avatar ikut tersimpan setelah community berhasil dibuat.

---

## Issue 2 — Invite via account name: undangan tidak muncul di list mana pun (mobile)

- **Repo**: `k_forum` (100% mobile — backend sudah benar dan lengkap)

### Ringkasan

Saat leader/moderator mengundang user lain via pencarian nama akun, undangan **berhasil dibuat di server** (dikonfirmasi via code review `send_invitation.go` — response sukses berisi status `pending`). Tapi **user yang diundang tidak punya cara melihatnya di app** — tidak ada screen "My Invitations" sama sekali.

### Root cause (code review)

Backend sudah lengkap dan benar:
- `POST /mobile/communities/{community_id}/invitations` → buat undangan (leader/moderator).
- `GET /mobile/me/community-invitations` → list undangan milik user sendiri (recipient).

Di sisi mobile, endpoint kedua **sudah ada wiring-nya** sampai layer data source (`community_remote_data_source.dart:704`) dan usecase (`GetMyCommunityInvitationsUseCase` di `invitation_usecases.dart:77`) — tapi **tidak pernah dipanggil dari screen manapun**. Konfirmasi: `grep` untuk `GetMyCommunityInvitationsUseCase` di seluruh `lib/` hanya menemukan definisinya sendiri, nol pemanggilan.

Providers `AcceptInvitationUseCase`/`RejectInvitationUseCase` untuk Community juga sudah didaftarkan (`community_providers.dart:200-204`) tapi sama-sama tidak dipanggil dari screen manapun.

**Bandingkan dengan modul Region** yang punya pola identik tapi LENGKAP: `region_invitations_screen.dart` benar-benar memanggil `AcceptInvitationUseCase`/`RejectInvitationUseCase` (baris 102, 114). Community ketinggalan langkah terakhir ini — layer data/domain sudah siap, screen-nya belum pernah dibuat.

Satu-satunya screen terkait invite komunitas yang ada (`community_invite_join_screen.dart`) itu untuk flow **invite LINK** (`/c/join/{code}`), bukan untuk invite-by-account-name — beda usecase sama sekali (`PreviewInviteLinkUseCase`, bukan `GetMyCommunityInvitationsUseCase`).

### Yang diminta ke mobile (k_forum)

1. Buat screen "My Community Invitations" (bisa dicontoh langsung dari `region_invitations_screen.dart`), pakai `GetMyCommunityInvitationsUseCase` yang sudah ada.
2. Wire `AcceptInvitationUseCase`/`RejectInvitationUseCase` (Community) ke screen tersebut.
3. Tambahkan entry point yang jelas ke screen ini (badge notifikasi di tab Community atau di halaman profil), supaya user tahu di mana harus mencari undangan yang masuk.

### Kriteria selesai (acceptance)

- [ ] Leader undang user by account name → user yang diundang bisa lihat undangan tersebut di sebuah screen di app (bukan cuma notifikasi push).
- [ ] User bisa accept/reject dari screen tersebut, dan hasilnya konsisten dengan status member di komunitas terkait.

---

## Issue 3 — Invite via link: tombol "Join" masih tampil di list lain setelah berhasil join

- **Repo**: `k_forum`

### Ringkasan

Screen landing invite-link (`community_invite_join_screen.dart`) sendiri **sudah benar** — begitu redeem sukses (`redeem.joined == true`), langsung `context.go()` ke halaman detail komunitas, jadi tombol join di screen itu tidak "nyangkut". Masalahnya ada di list komunitas **lain** yang sudah ter-load sebelumnya di memori (tab Discover / My Communities) — list itu tidak ikut ter-refresh.

### Root cause (code review)

Di `communities_discover_view.dart:176-182`, ada mekanisme refresh-on-return, tapi **hanya untuk satu jalur navigasi**:
```dart
Future<void> _open(Community community) async {
  final changed = await context.push<bool>('/communities/${community.id}', extra: community);
  if (changed == true) _load();
}
```
Ini cuma jalan kalau user membuka detail komunitas **dari list Discover itu sendiri** (pakai `push<bool>`, nunggu hasil pop).

Tapi flow invite-link (`community_invite_join_screen.dart:88`) navigasi pakai:
```dart
context.go('/communities/${community.id}', extra: community);
```
`go()` **mengganti stack**, tidak ada mekanisme "return value" seperti `push<bool>` di atas. Jadi kalau list Discover/My Communities kebetulan sudah ter-load di memori sebelum user membuka invite link (umum terjadi karena `CommunityScreen` pakai `TabController` — state tab tidak di-dispose saat pindah tab), list itu **tidak pernah tahu** ada perubahan membership, dan tombol "Join" untuk komunitas yang baru saja di-join tetap tampil sampai user pull-to-refresh manual atau restart app.

### Yang diminta ke mobile (k_forum)

1. Setelah redeem invite link sukses, broadcast event "membership changed" (lewat provider/notifier yang sudah dipakai, atau `EventBus` kalau ada) yang didengarkan oleh `communities_discover_view.dart` dan `my_communities_view.dart` untuk trigger `_load()` ulang.
2. (Alternatif lebih sederhana) Sebelum navigasi ke detail komunitas dari invite-link screen, invalidate/refresh cache list komunitas kalau state management-nya sudah pakai shared store (bukan local State per screen).

### Kriteria selesai (acceptance)

- [ ] Join komunitas via invite link → kembali ke tab Discover tanpa pull-to-refresh manual → komunitas tersebut sudah tidak menampilkan tombol "Join" (atau berubah jadi "Joined"/hilang dari Discover kalau memang didesain begitu).

---

## Issue 4 — Link share post memakai domain & path yang salah (tidak bisa dibuka app)

- **Repo**: `k-forum-api`

### Ringkasan

Link yang dihasilkan tombol "share" post komunitas memakai domain yang **tidak terverifikasi sebagai universal link** oleh app — jadi kalau link itu dibuka orang lain, tidak akan membuka app K-Forum, hanya browser biasa (atau 404 kalau domainnya memang tidak live).

### Root cause (code review)

Domain universal link yang benar-benar dikonfigurasi & di-handle app (`lib/core/constants/deep_links.dart:16`):
```dart
static const String host = 'k-forum-app.yubicom.co.id';
```

Tapi backend hardcode domain **lain** di 2 tempat:
```go
// internal/app/usecase/community/share_post.go:32
shareURL := fmt.Sprintf("https://app.k-forum.id/c/%s/p/%s", post.CommunityID, postID)

// internal/app/usecase/community/create_invite_link.go:83
return "https://app.k-forum.id/c/join/" + code
```

`app.k-forum.id` ≠ `k-forum-app.yubicom.co.id` — dua domain yang sama sekali berbeda. Bahkan kalaupun domainnya benar, **path share-post juga salah bentuk**: backend generate `/c/{communityId}/p/{postId}`, sementara app sendiri expect `/posts/{postId}` (`deep_links.dart:29`: `static String post(String id) => '$baseUrl/posts/$id';`). Jadi share-post rusak di 2 lapis sekaligus (domain + path); invite-link cuma rusak di domain (path `/c/join/{code}` sudah cocok dengan `deep_links.dart:28`).

Tidak ada konstanta domain yang benar dikonfigurasi di manapun di `k-forum-api` — bukan salah value doang, memang belum pernah ada.

### Yang diminta ke backend (k-forum-api)

1. Tambah config/constant base URL universal link yang benar (`https://k-forum-app.yubicom.co.id`), idealnya via env var, dipakai konsisten di `share_post.go` dan `create_invite_link.go`.
2. Perbaiki path `share_post.go` jadi `/posts/{postId}` (samakan dengan `DeepLinks.post()` di app), bukan `/c/{communityId}/p/{postId}`.
3. Cek modul lain yang mungkin generate share link serupa (News, Event, Directory) — pola hardcode domain salah ini mungkin bukan cuma di Community.

### Kriteria selesai (acceptance)

- [ ] Share post dari app → link yang dibagikan pakai domain `k-forum-app.yubicom.co.id` dan path `/posts/{id}`.
- [ ] Link invite komunitas pakai domain yang sama.
- [ ] Link yang dibagikan, kalau dibuka di device yang sudah install app, membuka app langsung ke post/komunitas terkait (bukan browser).

---

## Issue 5 — Moderator tidak bisa menghapus post orang lain walau punya permission `manage_content`

- **Repo**: `k-forum-api`

### Ringkasan

Moderator komunitas (bukan platform admin) tidak punya cara menghapus post member lain dari mobile app, padahal role-permission system menetapkan moderator berhak melakukan moderasi konten (`manage_content`/`delete_content`).

### Root cause (code review)

Mobile hanya punya **satu** endpoint delete post: `DELETE /mobile/communities/posts/:post_id` → `DeletePostByAuthorUseCase.Execute`, yang cek:
```go
if !principal.CanActAsAuthor(p, post.AuthorID) {
    return apperror.Forbidden(string(apperror.CodeForbidden))
}
```
**Hanya mengizinkan penulis post itu sendiri.** Tidak ada cabang untuk moderator/leader sama sekali di jalur ini.

Usecase yang seharusnya menangani moderasi (`RemovePostUseCase`, komentar kode: `// Required — role: moderator, superadmin, admin | perm: delete_content`) memang ada dan diimplementasikan, tapi:
1. **Hanya di-wire ke web/backoffice**, tidak ada endpoint mobile untuk usecase ini sama sekali (`router.go:567`: `communities.POST("/posts/:post_id/remove", webCommunityHandler.RemovePost)` — grup `communities` di web).
2. Grup route web tersebut di-gate `middleware.RequireAdmin()` (`router.go:535`) — jadi walau ada endpoint web-nya, **community moderator biasa (bukan platform admin) tetap tidak bisa memanggilnya**, karena mereka tidak punya role admin platform.
3. `RemovePostUseCase.Execute` sendiri **tidak melakukan pengecekan permission `manage_content`/`delete_content` di level community sama sekali** (beda dengan `SendInvitationUseCase` yang explicit call `principal.CanModerateInCommunity(...)`) — jadi kalaupun endpoint-nya dibuka, permission moderator-nya belum tervalidasi di usecase.

Kesimpulan: **tidak ada jalur kode manapun** (mobile atau web) yang benar-benar mengizinkan community moderator (bukan platform admin) menghapus post orang lain saat ini.

### Yang diminta ke backend (k-forum-api)

1. Tambah endpoint mobile baru, misal `POST /mobile/communities/posts/:post_id/remove`, wired ke `RemovePostUseCase`.
2. Tambahkan pengecekan `principal.CanModerateInCommunity(p, post.CommunityID)` (atau permission check `manage_content`/`delete_content` yang setara) **di dalam** `RemovePostUseCase.Execute`, jangan cuma andalkan gate route.
3. Untuk endpoint web yang sudah ada: pertimbangkan apakah tetap perlu `RequireAdmin()`, atau diubah supaya moderator (bukan cuma platform admin) juga bisa lewat sana kalau backoffice memang dipakai moderator juga (lihat Issue 6 — ada indikasi backoffice memang dirancang bisa diakses leader/moderator).

### Kriteria selesai (acceptance)

- [ ] Moderator (bukan author, bukan platform admin) bisa hapus post member lain di komunitas yang dia moderasi, dari mobile app.
- [ ] Member biasa (bukan author, bukan moderator/leader) tetap ditolak (403) saat mencoba endpoint yang sama.
- [ ] Moderator komunitas LAIN (bukan komunitas tempat post itu berada) tetap ditolak — permission harus scoped per-community, bukan global.

---

## Issue 6 — Leader komunitas melihat SEMUA komunitas di backoffice, bukan cuma miliknya

- **Repo**: `k-forum-backoffice`

### Ringkasan

Halaman "Communities Management" di backoffice menampilkan seluruh komunitas di platform tanpa filter, walau yang login adalah leader (bukan platform admin) yang seharusnya cuma boleh urus komunitas yang dia pimpin. Belum ada halaman "manage community" khusus yang scoped untuk leader/moderator.

### Root cause (code review)

`app/pages/community/index.vue:13-17`:
```js
definePageMeta({
  title: 'Communities Management',
  middleware: 'permission',
  permission: { roles: ['usergod', 'super admin', 'superadmin', 'admin', 'admin region'], permissions: ['create_community', 'moderate_posts'] },
})
```
Halaman ini didesain untuk platform admin (daftar role yang diizinkan cuma role platform — `leader`/`moderator` community-level TIDAK ada di daftar `roles`). Tapi kondisi permission-nya **OR** dengan `permissions: ['create_community', ...]` — dan `create_community` itu sebenarnya **benefit subscription Pro**, bukan role-permission platform (lihat `PLAN_SUBSCRIPTION_SYSTEM.md`). Kalau middleware `permission` ini juga mengecek benefit selain role, ini yang membuka pintu bagi Pro member/leader untuk lolos ke halaman admin ini sama sekali.

Begitu masuk, `loadCommunities()` (baris ~72-83) memanggil `communityStore.fetchCommunities({ q, status, visibility, category_id, region_id, sort, limit, offset })` — **tidak ada filter `leader_id`/`owner_id`/"komunitas saya" dalam bentuk apapun**. Baik superadmin maupun leader yang somehow bisa masuk, memanggil query yang **identik**, mengembalikan seluruh komunitas di platform.

Ada middleware `app/middleware/community-role.ts` yang isinya justru dirancang untuk validasi role per-komunitas (`leader > moderator > member` hierarchy, baca `to.meta.requiredCommunityRole`) — tapi **middleware ini tidak pernah di-declare di halaman manapun** (`app/pages/community/[id].vue` tidak punya `definePageMeta` sama sekali, apalagi `middleware: 'community-role'`). Jadi middleware yang seharusnya jadi solusi untuk masalah ini sendiri adalah dead code, tidak pernah jalan.

### Yang diminta ke backoffice (k-forum-backoffice)

1. Buat halaman terpisah khusus untuk leader/moderator (misal `/community/mine` atau `/community/manage`), yang memanggil endpoint list yang di-scope ke komunitas yang dipimpin/dimoderasi user tersebut saja (perlu cek dulu apakah endpoint backend seperti ini sudah ada — kalau belum, ini juga PR ke `k-forum-api`).
2. Jangan andalkan benefit subscription (`create_community`) sebagai syarat masuk halaman **admin** platform — pisahkan benefit check (fitur mobile) dari role check (akses backoffice).
3. Wire ulang atau hapus `community-role.ts` — kalau memang masih relevan untuk halaman detail komunitas per-leader, pasang `definePageMeta({ middleware: 'community-role', requiredCommunityRole: 'moderator' })` di halaman yang sesuai. Kalau sudah tidak relevan (karena solusinya jadi halaman terpisah di poin 1), hapus supaya tidak membingungkan developer selanjutnya.

### Kriteria selesai (acceptance)

- [ ] Leader (bukan platform admin) yang login ke backoffice hanya bisa melihat & mengelola komunitas yang dia pimpin.
- [ ] Halaman "Communities Management" (list semua komunitas) tetap admin-only, tidak lolos oleh benefit Pro semata.
- [ ] Ada halaman terpisah yang jelas untuk leader/moderator kelola komunitas mereka sendiri.

---

## Issue 7 — Session leader masih aktif setelah tidak lagi jadi leader (cache Redis basi)

- **Repo**: `k-forum-api`
- **Severity**: 🟢 Rendah — disepakati bukan blocker, cache expire sendiri seiring waktu

### Ringkasan

User yang baru saja transfer/lepas kepemimpinan komunitas masih bisa melakukan aksi sebagai leader di komunitas eks-pimpinannya untuk sementara waktu, sampai cache-nya expire sendiri.

### Root cause (code review)

`Principal` (termasuk role/permission komunitas seperti leader) di-cache di Redis lewat `RedisPrincipalCache.Set(ctx, userID, p, ttl)` (`internal/infrastructure/cache/redis_principal_cache.go`).

Pola yang sudah dipakai di modul lain: setiap kali role/subscription seorang user berubah, cache principal-nya di-invalidate secara eksplisit — dikonfirmasi ada di `usermanagement/assign_user_role.go`, `usermanagement/revoke_user_role.go`, `usermanagement/assign_user_roles_bulk.go`, dan beberapa usecase subscription.

Tapi `internal/app/usecase/community/transfer_ownership.go` dan `internal/app/usecase/community/delete_community_by_leader.go` (dua usecase yang mengubah status leadership seseorang) **tidak memanggil invalidasi cache principal sama sekali** — jadi mengikuti pola yang sama seperti modul lain, ini murni gap konsistensi, bukan masalah desain fundamental Redis-nya.

### Yang diminta ke backend (k-forum-api) — non-blocking, boleh dikerjakan belakangan

1. Tambah pemanggilan invalidasi `RedisPrincipalCache` untuk user yang di-demote di `transfer_ownership.go` (leader lama) dan `delete_community_by_leader.go`, mengikuti pola yang sama seperti `usermanagement/revoke_user_role.go`.

### Kriteria selesai (acceptance)

- [ ] Setelah transfer ownership komunitas, leader lama langsung (tidak perlu menunggu TTL) kehilangan akses aksi leader di komunitas tersebut pada request berikutnya.

---

## Referensi

- Journey terkait: [`flows/user-journeys/04_COMMUNITY_JOURNEY.md`](../flows/user-journeys/04_COMMUNITY_JOURNEY.md).
- Spec: `Modules/Community/COMMUNITY_RULES.md`, `Modules/Community/COMMUNITY_INVITE_RULES.md`, `Modules/Role-Permission/COMMUNITY_ROLE_PERMISSION_CONFIG.md`.
- Kode kunci — `k-forum-api`: `internal/app/usecase/community/{send_invitation,remove_post,delete_post_by_author,share_post,create_invite_link,transfer_ownership,delete_community_by_leader}.go`, `internal/interfaces/http/router/router.go`, `internal/infrastructure/cache/redis_principal_cache.go`.
- Kode kunci — `k_forum`: `lib/features/community/presentation/screens/{community_form_screen,community_invite_join_screen}.dart`, `lib/features/community/presentation/widgets/communities_discover_view.dart`, `lib/features/community/domain/usecases/invitation_usecases.dart`, `lib/core/constants/deep_links.dart`.
- Kode kunci — `k-forum-backoffice`: `app/pages/community/index.vue`, `app/pages/community/[id].vue`, `app/middleware/community-role.ts`.
