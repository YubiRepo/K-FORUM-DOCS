# Issue: Region — undangan email, list dobel, template email seragam

- **Modul**: Region & Onboarding — undangan region, list region (mobile), email transaksional
- **Severity**: 🟠 Sedang-Tinggi — bukan blocker fungsional (in-app tetap jalan), tapi UX undangan & list region rusak, dan copy email reset password berpotensi bikin user bingung/curiga (dianggap phishing)
- **Status**: 🔴 Open — 3 issue di dokumen ini, root cause sudah ditemukan by code review, belum ada fix
- **Ditemukan**: 22 Jul 2026, saat review user journey `01_ONBOARDING_ACCOUNT_JOURNEY.md`
- **Pelapor**: review manual (dev), dikonfirmasi via code review langsung (bukan cuma reproduksi runtime)

---

## Issue 1 — Email undangan region tidak pernah terkirim

- **Repo**: `k-forum-api` (backend only, tidak ada yang perlu diubah di mobile/backoffice)
- **Endpoint terkait**: `POST /api/v1/web/regions/:region_id/invite` (dan `resend`)

### Ringkasan

Saat admin mengundang member ke region, undangan **muncul di akun member terkait** (in-app), tapi **email undangan tidak pernah terkirim** — bukan cuma gagal kirim (error di log), tapi memang **tidak pernah dicoba kirim**.

### Root cause (code review)

Ada 2 jalur yang independen:

1. **In-app (jalan)**: `InviteMembersUseCase.Execute` (`internal/app/usecase/region/invite_members.go`) langsung `invitationRepo.Save(...)` ke tabel `region_invitations` dalam transaksi yang sama → langsung terbaca oleh endpoint list-invitations member. Ini yang membuat undangan "muncul di akun".
2. **Email (rusak)**: di transaksi yang sama, usecase juga publish event ke outbox dengan routing key `region.invitation.email.requested` (`invite_members.go:104`, konstanta di `internal/domain/region/event/routing.go:4`). **Tidak ada consumer yang terdaftar untuk routing key ini** — cek `internal/interfaces/mq/router/router.go`, hanya ada:
   ```go
   r.Register(regionevent.RoutingRegionJoinApproved, h.HandleRegionJoinApproved)  // baris 67
   ```
   Tidak ada baris setara untuk `RoutingRegionInvitationEmailRequested`. Event ini publish ke queue yang **tidak pernah dikonsumsi** — bukan soal SMTP dikonfigurasi atau tidak (`cmd/app/main.go:141-152` fallback ke `NoopEmailSender` kalau `SMTP_HOST` kosong), tapi handler-nya memang belum pernah dibuat. `emailSender.SendRegionInvitation(...)` (yang sudah ada implementasinya di `internal/infrastructure/external/smtp/email_sender.go:48`) **tidak pernah terpanggil sama sekali** untuk flow ini.
3. Fungsi `resend_invitation.go` (tombol "resend" undangan) punya bug yang sama — publish ke routing key yang sama, sama-sama tidak ada konsumen.

### Kendala yang ditimbulkan

- Member yang diundang via email tidak akan pernah tahu ada undangan kecuali mereka kebetulan membuka app dan cek notifikasi/list undangan sendiri.
- Fitur "resend invitation" di backoffice terlihat berhasil (200 OK) tapi sama-sama tidak mengirim apa-apa — false positive bagi admin yang menekan resend.

### Yang diminta ke backend (k-forum-api)

1. Tambah handler `HandleRegionInvitationEmailRequested` di `internal/interfaces/mq/handler/region_handler.go`, isinya panggil `emailSender.SendRegionInvitation(...)` dengan data dari payload `regionevent.InvitationEmailRequested`.
2. Register di `internal/interfaces/mq/router/router.go`: `r.Register(regionevent.RoutingRegionInvitationEmailRequested, h.HandleRegionInvitationEmailRequested)`.
3. Daftarkan queue-nya juga di `cmd/worker/main.go` (pola persis seperti baris `{regionevent.QueueRegionJoinApproved, regionevent.RoutingRegionJoinApproved}` di baris 416).
4. Pastikan payload event `InvitationEmailRequested` (di `internal/domain/region/event/events.go`) sudah bawa semua data yang dibutuhkan template email (region name — lihat Issue 3, bukan cuma region ID).

### Kriteria selesai (acceptance)

- [ ] Undang member baru → email diterima di inbox (verifiable di environment dengan SMTP aktif, atau log noop sender menunjukkan `SendRegionInvitation` terpanggil).
- [ ] Tombol "resend invitation" di backoffice benar-benar mengirim ulang email, bukan cuma 200 OK kosong.
- [ ] Log worker menunjukkan queue `region.invitation.email.requested` terkonsumsi (tidak menumpuk di outbox).

---

## Issue 2 — List region di mobile tampil dobel untuk region milik sendiri

Ada **2 root cause independen** yang berkontribusi ke bug yang sama secara visual. Keduanya perlu diperbaiki.

### 2a. Mobile (`k_forum`) — penyebab utama, selalu terjadi

- **Repo**: `k_forum` (Flutter app)
- **File**: `lib/features/region/presentation/screens/regions_browse_screen.dart`

**Ringkasan**: Screen "Browse Region" memanggil 2 use case terpisah lalu render keduanya tanpa dedup:
- `GetMyRegionUseCase` → disimpan ke `_myRegion`, dirender sebagai card khusus "region saya" di bagian atas (baris ~274-328).
- `GetRegionsUseCase` → disimpan ke `_regions`, dirender sebagai grid/list umum di bawahnya (baris ~251, `itemCount: _regions.length`).

`_regions` (list umum dari API) **tidak difilter** untuk membuang region yang `id`-nya sama dengan `_myRegion?.id`. Karena API list umum secara desain memang mengembalikan semua region termasuk milik sendiri (dengan field `your_status`/`your_role` terisi — lihat `API_SPEC_REGION_MOBILE.md`), region yang sama muncul dua kali di layar: sekali di card atas, sekali lagi di grid bawah.

**Yang diminta ke mobile (k_forum)**:
1. Di `_load()` dan `_loadMore()`, exclude region dengan `id == _myRegion?.id` dari `_regions` sebelum `setState`. Cukup satu baris filter setelah assignment `_regions = page.items;` (baris ~97) dan di bagian merge `fresh` (baris ~118-121).

### 2b. Backend (`k-forum-api`) — edge case tambahan, muncul kalau user pernah `rejected` lalu request lagi

- **Repo**: `k-forum-api`
- **File**: `internal/infrastructure/persistence/postgres_region_query.go`, fungsi `ListRegionsForMobile` (baris ~141-161)

**Ringkasan**: Query list region mobile join ke `region_memberships` tanpa filter status:

```go
userJoin = fmt.Sprintf(`LEFT JOIN region_memberships my_m ON my_m.region_id = r.id AND my_m.user_id = $%d`, argIdx)
```

Unique index di tabel ini cuma scoped ke status aktif:

```sql
-- internal/migrations/0004_create_region_tables.up.sql:43
CREATE UNIQUE INDEX idx_region_memberships_active_unique ON region_memberships (user_id, region_id) WHERE status = 'active';
```

Artinya user **bisa punya lebih dari 1 row** untuk `(user_id, region_id)` yang sama selama cuma 1 yang `active` — misalnya 1 row `rejected` (dari request join yang ditolak) + 1 row `pending_approval`/`active` (dari request ulang). `RequestJoinRegionUseCase.Execute` (`internal/app/usecase/region/request_join_region.go:38-46`) hanya block kalau existing membership `IsActive()` atau `IsPending()` — kalau statusnya `rejected`, tidak di-block dan tidak dibersihkan, langsung bikin row baru:

```go
if existing != nil {
    if existing.IsActive() { /* conflict */ }
    if existing.IsPending() { /* conflict */ }
    // rejected: lolos ke bawah, row baru dibuat, row rejected lama TIDAK dihapus
}
```

Karena `LEFT JOIN my_m` tidak difilter status dan `GROUP BY` menyertakan `your_status`/`your_role`, satu region menghasilkan **2 row di response API itu sendiri** untuk user yang pernah ditolak lalu join lagi.

**Yang diminta ke backend (k-forum-api)**:
1. Ubah join di `ListRegionsForMobile` supaya hanya ambil membership row **terbaru** per `(user_id, region_id)`, misal via `LEFT JOIN LATERAL (... ORDER BY created_at DESC LIMIT 1) my_m ON true`, bukan join biasa ke semua row historis.
2. (Opsional, defense-in-depth) Di `RequestJoinRegionUseCase`, hapus/supersede row `rejected` lama sebelum membuat row baru saat user request join ulang — supaya tabel `region_memberships` tidak menumpuk row basi per user per region.

### Kriteria selesai (acceptance)

- [ ] User dengan 1 region aktif, tanpa histori rejected: region itu tampil **1x** di layar Browse Region (bukan di card atas + grid bawah).
- [ ] User yang pernah `rejected` di region X lalu join lagi (approved/pending): `GET /mobile/regions` hanya balikin **1 row** untuk region X.
- [ ] Test regresi: user tanpa region sama sekali → tidak ada card "region saya", grid tampil normal tanpa row dobel.

---

## Issue 3 — Template email OTP/undangan/reset password seragam & kurang informatif

- **Repo**: `k-forum-api` (tidak ada yang perlu diubah di mobile/backoffice)
- **File**: `internal/app/port/email_sender.go`, `internal/infrastructure/external/smtp/email_sender.go`, `internal/app/usecase/auth/forgot_password.go:89`

### Ringkasan

Ini **lebih dari sekadar "belum dipersonalisasi"** — salah satu kasusnya adalah copy email yang **salah konteks**:

1. **Reset password memakai template OTP registrasi, apa adanya.** `EmailSender` interface (`internal/app/port/email_sender.go`) cuma punya 3 method: `SendOTP`, `SendRegionInvitation`, `SendNotification` — **tidak ada method khusus untuk reset password**. `forgot_password.go:89` memanggil `emailSender.SendOTP(...)`, method yang sama dengan verifikasi email saat registrasi, termasuk body-nya:

   > "Kode OTP verifikasi email Anda adalah: ... **Jika Anda tidak melakukan registrasi, abaikan email ini.**"

   User yang lupa password dan minta reset akan terima email yang menyebut "registrasi" — membingungkan, dan berpotensi membuat user curiga (mengira ini bukan email yang mereka minta / phishing).

2. **Email undangan region print raw UUID, bukan nama region.** `SendRegionInvitation` (`email_sender.go:48-74`) menyusun body dengan `Region ID: %s` diisi UUID mentah — user terima email berisi ID region yang tidak bisa dibaca manusia, bukan nama region-nya (misal "KAI Jakarta").

3. Ketiga jenis email (OTP, undangan, notifikasi generik) memakai gaya/struktur plain-text yang sama, tanpa branding atau pembeda konteks per alasan pengiriman.

### Kendala yang ditimbulkan

- Email reset password yang menyebut "registrasi" berisiko bikin user tidak percaya/abaikan email tersebut, padahal itu OTP yang mereka minta sendiri.
- Email undangan region dengan raw UUID tidak actionable buat user awam (tidak tahu ini undangan ke region mana).

### Yang diminta ke backend (k-forum-api)

1. Tambah method baru di `EmailSender` interface khusus reset password (misal `SendPasswordResetOTP(toEmail, username, otp string) error`), dengan body yang menyebut "reset password", bukan "registrasi". Update `forgot_password.go` untuk memanggil method baru ini, bukan `SendOTP`.
2. Ubah signature `SendRegionInvitation` untuk menerima `regionName` (bukan cuma `regionID`), dan pastikan usecase pemanggil (setelah Issue 1 diperbaiki) mengambil nama region sebelum publish event, supaya field ini tersedia di payload.
3. (Nice to have, tidak blocking) Rapikan template supaya ada minimal subject/signature yang konsisten dengan branding KAI App, bukan "Tim k-forum-api".

### Kriteria selesai (acceptance)

- [ ] Email reset password tidak lagi menyebut kata "registrasi"/"registration".
- [ ] Email undangan region menampilkan nama region yang manusiawi, bukan UUID.
- [ ] `EmailSender` punya method terpisah untuk OTP-registrasi vs OTP-reset-password (boleh share helper internal, tapi body/subject harus beda).

---

## Referensi

- Journey terkait: [`flows/user-journeys/01_ONBOARDING_ACCOUNT_JOURNEY.md`](../flows/user-journeys/01_ONBOARDING_ACCOUNT_JOURNEY.md) — Journey 5 (Diundang ke Region) & Journey 6 (Pindah/Keluar Region).
- Spec: `Modules/Region/REGION_SYSTEM_RULES.md`, `API SPEC/Mobile/API_SPEC_REGION_MOBILE.md`.
- Kode kunci yang dirujuk: `k-forum-api/internal/app/usecase/region/{invite_members,resend_invitation,request_join_region}.go`, `k-forum-api/internal/interfaces/mq/router/router.go`, `k-forum-api/internal/infrastructure/persistence/postgres_region_query.go`, `k-forum-api/internal/infrastructure/external/smtp/email_sender.go`, `k-forum-api/internal/app/usecase/auth/forgot_password.go`, `k_forum/lib/features/region/presentation/screens/regions_browse_screen.dart`.
