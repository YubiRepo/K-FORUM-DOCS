# Onboarding & Account — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md).

## Ringkasan Domain

Domain ini adalah journey **pertama** yang dilalui setiap member — dari buka aplikasi KAI App pertama kali, daftar/login, verifikasi, melengkapi profil, sampai (opsional) bergabung ke sebuah region dan mendarat di home siap pakai sebagai member **Standard**. Semua akses ke KAI App hanya lewat 📱 Mobile — tidak ada website member. Domain ini juga mencakup siklus akun dasar yang relevan di sisi member: lupa/reset password, ganti/pindah/keluar region, dan request hapus akun (soft-delete 30 hari). Sumber: `API_SPEC_AUTH.md`, `API_SPEC_PROFILE_MEMBERSHIP.md`, `API_SPEC_REGION_MOBILE.md`, `REGION_SYSTEM_RULES.md`.

Assignment plan Standard otomatis saat signup dijelaskan detail di [02_SUBSCRIPTION_UPGRADE_JOURNEY.md](02_SUBSCRIPTION_UPGRADE_JOURNEY.md) (Journey 1) — di dokumen ini cukup disebutkan terjadi di background, tidak diulang detailnya.

## Batasan Standard vs Pro di Domain Ini

Domain ini terjadi **sebelum** ada perbedaan tier yang berarti — semua akun baru otomatis Standard (lihat 02), dan region bersifat opsional serta terbuka untuk kedua tier dengan aturan yang sama persis.

| Aksi                                                 | Standard | Pro  | Catatan                                                                                                                           |
| ---------------------------------------------------- | :------: | :--: | --------------------------------------------------------------------------------------------------------------------------------- |
| Registrasi & verifikasi OTP                          |    ✅    |  ✅  | Sama untuk semua akun baru — semua mulai sebagai Standard                                                                         |
| Login via Google (Sign-In/Sign-Up unified)           |    ✅    |  ✅  | Tidak ada perbedaan tier                                                                                                          |
| Melengkapi profil (bio, alamat, avatar, dll)         |    ✅    |  ✅  | Tidak digerbang oleh plan                                                                                                         |
| Request join region                                  |    ✅    |  ✅  | Tidak digerbang oleh plan                                                                                                         |
| Terima/tolak undangan region                         |    ✅    |  ✅  | Tidak digerbang oleh plan                                                                                                         |
| Pindah / keluar region                               |    ✅    |  ✅  | Tidak digerbang oleh plan                                                                                                         |
| Lupa/reset password, ganti email, request hapus akun |    ✅    |  ✅  | Sama untuk semua tier                                                                                                             |
| Approve/reject join request member lain              |    ❌    |  ❌  | Butuh role Admin Region/Superadmin, bukan soal tier                                                                               |
| Jadi Admin Region                                    |   ❌\*   | ❌\* | \*Bisa terjadi kalau di-invite Superadmin sebagai admin — tapi itu perubahan **role**, bukan hak yang didapat dari upgrade ke Pro |

## Journey 1: Registrasi Akun Baru & Verifikasi OTP — 🅢🅟 — 📱 Mobile ( ✅ DONE 21 Juli 20026 (soft test, no detail) )

1. **Buka aplikasi pertama kali** → layar welcome dengan pilihan [Login] / [Register] / [Sign in with Google].
2. Tap **Register** → isi form: `username`, `fullname` (opsional), `email`, `phone`, `password`.
3. Submit → `POST /api/v1/mobile/auth/register`.
   - **Sukses**: akun terbuat, response mengembalikan data user (`plan: "standart"`, `roles: ["member"]`) beserta pesan **"Registration successful. Please verify OTP."**
   - **Error 422**: validation error per field (email format salah, password < 8 karakter, dll) — pesan pertama ditampilkan ke user.
4. Sistem mengirim kode OTP ke email yang didaftarkan (berdasarkan pesan response di atas — spec tidak eksplisit menyatakan trigger otomatis, lihat catatan di bagian Edge Case). Jika belum menerima, user bisa tap "Resend OTP" → `POST /otp/send`.
5. User memasukkan kode OTP → `POST /api/v1/mobile/auth/otp/verify` dengan `identifier` (email) + `otp`.
   - **Sukses**: `"OTP verified successfully"` + token verifikasi sementara.
   - **Error**: kode salah/expired → tampilkan pesan error, izinkan retry/resend.
6. Setelah verifikasi, user diarahkan untuk **Login** (atau, jika desain app menyatukan alur, langsung dianggap login menggunakan token dari step sebelumnya).
7. Login sukses → app menyimpan `token` + `refresh_token`, lanjut ke **Journey 3 (Melengkapi Profil, opsional)** atau langsung ke home.

**Selesai:** Akun terverifikasi, berstatus Standard aktif (detail assignment plan lihat 02), siap dipakai atau melengkapi profil terlebih dulu.

> **Login biasa (returning user):** `POST /api/v1/mobile/auth/login` dengan `identifier` + `password` (+ `device_id` kalau app punya fitur kelola device). Tidak ada langkah OTP tambahan di sini — OTP hanya di alur registrasi & forgot password.

## Journey 2: Login/Registrasi via Google (Unified) — 🅢🅟 — 📱 Mobile (✅ DONE 21 Juli 20026 (soft test, no detail))

1. Tap **[Sign in with Google]** di layar welcome → proses Google Sign-In native, dapat `id_token`.
2. App memanggil `POST /api/v1/mobile/auth/login/google` dengan `id_token` (+ `device_id` opsional).
3. Backend cek email Google:
   - **Belum terdaftar** → akun baru otomatis dibuat (skip form registrasi & OTP sama sekali), langsung dapat `token` + `refresh_token`.
   - **Sudah terdaftar** → langsung login.
4. Avatar diambil langsung dari URL Google (`https://lh3.googleusercontent.com/...`) — tidak lewat proses upload S3 seperti avatar biasa.
5. User mendarat di home (akun baru) atau lanjut sesi sebelumnya (akun lama).

**Selesai:** Tidak ada friksi tambahan — jalur Google adalah shortcut penuh menuju home, tanpa OTP maupun form registrasi manual.

## Journey 3: Melengkapi Profil (Opsional, Kapan Saja) — 🅢🅟 — 📱 Mobile (✅ DONE 21 Juli 20026 (soft test, no detail))

1. Dari home atau menu Profile, user membuka "Complete your profile" (biasanya prompted setelah registrasi, tapi **bisa dilewati/skip**).
2. Isi data: `name`, `phone`, `birth_date`, `gender`, `bio`, `marital_status`, `occupation`, `interests`, `address` → submit → `PATCH /api/v1/mobile/profile/me`.
   - **Error 422**: per-field validation (mis. `birth_date` di masa depan, umur < 13 tahun, `bio` > 500 karakter).
3. **Upload avatar** (terpisah dari form di atas):
   - `POST /avatar/presign` (kirim `filename`, `mime_type`) → dapat `presigned_url` + `s3_path`.
   - Upload file langsung ke S3 pakai `presigned_url`.
   - `POST /avatar/confirm` (kirim `s3_path`) → avatar + thumbnail langsung ter-update di profile.
   - User juga bisa `DELETE /avatar` untuk kembali ke avatar default.
4. Profil bisa dilihat kembali kapan saja lewat `GET /profile/me` (data personal saja — memberships/subscription di-fetch terpisah, lihat catatan strategi endpoint di spec).

**Selesai:** Profil lengkap (atau tetap minimal — melengkapi profil tidak wajib untuk memakai fitur inti aplikasi).

> **Catatan penting**: field `is_verified` yang muncul di response `GET /profile/me` **kemungkinan besar bukan** status "email terverifikasi via OTP" — OTP verification tidak menghasilkan field semacam ini di response Auth manapun. Lebih mungkin field ini terkait fitur **Verification Badge** (lihat [09_VERIFICATION_BADGE_JOURNEY.md](09_VERIFICATION_BADGE_JOURNEY.md)). Ini dicatat sebagai **open question** — bukan diasumsikan, karena tidak ada definisi eksplisit di source docs domain ini.

## Journey 4: Join Region Pertama Kali (Browse & Request) — 🅢🅟 — 📱 Mobile (✅ DONE 21 Juli 20026 (soft test, no detail))

Region bersifat **opsional** — user baru default `region_id: null` ("no_region") dan tetap bisa memakai semua fitur platform tanpa region selamanya.

1. Dari home, user melihat indikator "You're not in any region yet" → tap **[Browse Regions]** → `GET /api/v1/mobile/regions` (list semua region status aktif, bisa `search` by nama/slug).
2. Tap salah satu region (mis. KAI Jakarta) → lihat detail (`GET /regions/{id}`) dan daftar member publik (nama + avatar saja, tanpa data sensitif).
3. Tap **[Request to Join]** → `POST /regions/{id}/request` (body kosong).
   - **Sukses (201)**: membership dibuat berstatus `pending_approval`. Pesan: _"Join request submitted. Admin akan review dalam 1-2 hari."_
   - **Error 409**: sudah jadi active member di region ini, atau sudah punya request pending ke region ini.
4. User bisa membatalkan request yang masih pending: `POST /regions/{id}/request/cancel`.
5. Menunggu review Admin Region (lihat bagian Keterlibatan Admin) → notifikasi masuk saat disetujui/ditolak:
   - **Approved** → status jadi `active`, `joined_at` terisi, notifikasi "Bergabung ke KAI Jakarta ✓".
   - **Rejected** → status `rejected` beserta alasan dari admin (kalau diisi); record tetap tersimpan (audit trail), user bisa request lagi kapan saja.

**Selesai:** User jadi active member sebuah region (atau tetap tanpa region kalau memilih tidak melanjutkan — tidak ada konsekuensi terhadap fitur lain).

## Journey 5: Diundang ke Region (Invitation, Sebelum atau Sesudah Registrasi) — 🅢🅟 — 📱 Mobile (✅ DONE 21 Juli 2026 (soft test, no detail))

Alur ini dipicu Admin Region/Superadmin mengundang lewat email (lihat Keterlibatan Admin), bukan inisiatif member — tapi penyelesaiannya ada di sisi member.

**Kasus A — email sudah terdaftar sebagai member:**

1. User menerima notifikasi/email undangan → buka "Pending Invitations" → `GET /regions/invitations/pending` (menampilkan `time_left_hours` tersisa dari window 24 jam).
2. Tap **[Accept]** → `POST /invitations/{id}/accept`. Syarat: email di access token harus match email undangan.
   - **Sukses**: membership langsung `active` (**tidak** melalui `pending_approval` — beda dari alur request join biasa), pesan _"Bergabung ke KAI Jakarta ✓"_.
   - **Error 410**: undangan sudah lewat 24 jam ("Admin can send a new one").
   - **Error 409**: undangan sudah pernah di-accept sebelumnya.
3. Atau tap **[Reject]** → `POST /invitations/{id}/reject` — status jadi rejected permanen, admin bisa invite ulang nanti.

**Kasus B — email belum pernah registrasi:**

1. User klik link undangan dari email → diarahkan ke form **Register** dengan email **sudah pre-filled**.
2. User menyelesaikan registrasi & verifikasi OTP normal (Journey 1).
3. Setelah registrasi selesai, sistem otomatis membuat `RegionMembership` berstatus `active` untuk region yang mengundang — **tanpa** perlu request/approval manual.
4. First-time welcome message menyertakan info region tersebut.

> **Catatan role**: kalau yang mengundang adalah **Superadmin** dan mengundang "sebagai Admin" (bukan member biasa), hasil accept adalah `role: admin` — user tersebut menjadi **Admin Region**, bukan lagi murni member Standard/Pro. Ini disebut untuk kelengkapan alur, tapi begitu jadi Admin Region, aktivitasnya sudah keluar dari cakupan "member journey" dokumen ini.

## Journey 6: Pindah / Keluar Region — 🅢🅟 — 📱 Mobile (✅ DONE 21 Juli 2026 (soft test, no detail))

1. Active member membuka "My Region" → `GET /regions/me` (menampilkan `your_status: "active"`, `your_role`, `joined_at`).
2. Tap **[Leave This Region]** → modal konfirmasi: _"Yakin keluar dari KAI Jakarta? Anda bisa request join lagi nanti."_
3. Confirm → `POST /regions/{id}/leave`. Syarat: harus active member di region tersebut.
   - Membership jadi `inactive`, notifikasi "You have left KAI Jakarta", user kembali ke status "no_region".
4. Untuk pindah ke region lain: user mengulang **Journey 4** (browse → request join region baru) — proses join region baru **sepenuhnya sama** untuk user yang baru pertama kali join maupun yang baru saja leave dari region lain.

**Selesai:** User berhasil keluar (dan opsional langsung lanjut request join region baru — kembali ke status `pending_approval` di region tujuan).

## Journey 7: Lupa Password (Sebelum Login) — 🅢🅟 — 📱 Mobile (✅ DONE 21 Juli 2026 (soft test, no detail))

1. Di layar login, tap **"Forgot Password?"** → masukkan email → `POST /password/forgot`. Response generik: _"Link/Kode OTP reset password telah dikirim ke email Anda"_ (tidak membocorkan apakah email terdaftar).
2. User memasukkan kode dari email → `POST /otp/verify` (`identifier` + `otp`) → dapat token verifikasi sementara.
3. User memasukkan password baru → `POST /password/reset` dengan `email`, `token` (dari step 2), `password` baru.
   - **Sukses**: _"Password berhasil diperbarui"_ → user diarahkan ke layar login untuk masuk dengan password baru.

## Journey 8: Request Hapus Akun (Soft-Delete 30 Hari) — 🅢🅟 — 📱 Mobile (✅ DONE 21 Juli 2026 (soft test, no detail))

1. Dari pengaturan akun, user tap **[Delete Account]** → masukkan password saat ini + alasan (opsional) → `POST /profile/account/delete-request`.
   - **Error 401**: password salah.
2. **Sukses**: sistem kirim OTP konfirmasi ke email, response menyertakan `deletion_scheduled_date` (30 hari dari sekarang) dan `deletion_token`.
3. User memasukkan OTP → `POST /profile/account/confirm-delete` → akun masuk masa **grace period 30 hari**; pesan: _"Akun Anda telah dihapus dan akan dihapus sepenuhnya dalam 30 hari."_
4. **Selama 30 hari**, user masih bisa membatalkan: `POST /profile/account/cancel-delete` (kirim `deletion_token` + OTP yang sama) → status kembali `active`.
5. Setelah 30 hari tanpa pembatalan → data dihapus permanen (proses backend, di luar interaksi member).

**Selesai (batal):** akun aktif normal kembali. **Selesai (tidak dibatalkan):** akun terhapus permanen setelah 30 hari.

> Rate limit disebutkan di spec: request hapus akun maksimal 1x per 7 hari, request ganti email maksimal 3x per 24 jam — relevan kalau member mencoba berulang kali dalam waktu singkat.

## Keterlibatan Admin — 💻 Web/Backoffice

**Registrasi & verifikasi akun (Journey 1, 2, 7, 8):** **Tidak ada** keterlibatan admin sama sekali — seluruhnya self-service otomatis lewat OTP/email, tidak ada langkah approval manual dari Superadmin/Admin Regional untuk pendaftaran, verifikasi, reset password, maupun penghapusan akun.

**Region — Superadmin:**

- Membuat region baru & mengisi info (name, slug, description, image) — region tidak bisa dihapus, hanya di-deactivate.
- Assign/remove Admin Region (invite user tertentu jadi admin sebuah region).
- Bisa melakukan semua yang Admin Region bisa lakukan, untuk region manapun.

**Region — Admin Region (scope: region miliknya saja):**

- Invite member baru via email (`POST` sejenis, hasilkan `RegionInvitation`, expire 24 jam).
- Meninjau **Pending Requests** dari Journey 4 → **[Approve]** (membership jadi active, `approved_by` tercatat) atau **[Reject]** (isi `rejection_reason` opsional).
- Remove member dari region (status jadi `inactive` — member yang di-remove bisa request join lagi atau menerima undangan baru kapan saja, lihat Rule 4 di REGION_SYSTEM_RULES).
- Tidak bisa mengedit info region, tidak bisa assign/remove admin, tidak bisa membuat/menghapus/deactivate region.

## Di Luar Cakupan Standard & Pro

- **Membuat region baru** — hanya Superadmin; region adalah struktur platform-level, bukan sesuatu yang dibuat member.
- **Assign/remove Admin Region** — hanya Superadmin; ini keputusan penempatan role, bukan hak subscription.
- **Approve/reject join request member lain, atau invite member ke suatu region** — butuh role Admin Region/Superadmin; member biasa (Standard maupun Pro) hanya bisa mengelola request/undangan miliknya sendiri.
- **Mengedit info region (nama, deskripsi, gambar) atau meng-nonaktifkan region** — hanya Superadmin.
- **Memverifikasi/menyetujui pendaftaran akun member lain secara manual** — tidak ada mekanisme semacam ini di source docs; verifikasi akun sepenuhnya via OTP otomatis, bukan review admin.

## Edge Case & Catatan Tambahan

- **Trigger OTP saat registrasi**: pesan response register ("Please verify OTP") mengindikasikan backend mengirim OTP otomatis begitu akun dibuat, tapi spec tidak menyatakan ini secara eksplisit sebagai side-effect wajib — dicatat sebagai asumsi, bukan fakta pasti dari dokumen.
- **Login sebelum OTP diverifikasi**: tidak ditemukan pernyataan eksplisit di `API_SPEC_AUTH.md` bahwa endpoint `login` menolak akun yang belum verifikasi OTP. Apakah backend benar-benar men-gate ini adalah **open question**.
- **Endpoint OTP verify bersifat generik**: `POST /otp/verify` dipakai baik untuk alur registrasi maupun forgot password (tokennya secara harfiah bernama `temp_verification_token_for_reset_password_or_auth`) — tidak ada parameter "purpose" untuk membedakan konteks pemakaian di spec.
- **Inkonsistensi penamaan plan**: response Auth (`register`, `login`, `me`) memakai string `"standart"`, sedangkan Profile & Membership memakai `"standard"`. Tidak jelas mana yang benar-benar dikirim backend — dicatat sebagai potensi bug/inkonsistensi dokumentasi, bukan diasumsikan salah satu benar.
- **Batas "1 region aktif" vs constraint yang didokumentasikan**: `REGION_SYSTEM_RULES.md` menyatakan di level konsep bahwa member maksimal 1 region aktif **secara keseluruhan** ("harus leave Jakarta dulu sebelum join Surabaya"), tapi constraint teknis yang disebutkan cuma unique index per `(user_id, region_id)` — yang secara harfiah hanya mencegah duplikat status aktif di region **yang sama**, bukan lintas region. `API_SPEC_REGION_MOBILE.md` untuk endpoint Request Join juga hanya mendokumentasikan 409 untuk kasus "sudah member"/"sudah pending" di region **itu juga** — tidak eksplisit menyebut penolakan request ke region lain saat user masih active di region berbeda. Perilaku pastinya (apakah request ke region kedua ditolak otomatis, atau harus leave manual dulu) adalah **open question**.
- **Status "suspended" akun secara global**: tidak ditemukan mekanisme atau field status semacam ini di keempat source docs (Auth, Profile & Membership, Region Mobile, Region System Rules). Yang ada hanya (a) status soft-delete dari alur hapus akun (Journey 8), dan (b) status `RegionMembership` (`active`/`pending_approval`/`rejected`/`inactive`) yang sifatnya scoped ke satu region, bukan status akun keseluruhan. Dokumen ini **tidak** mengasumsikan adanya fitur suspend akun member secara global karena tidak didukung source docs.
- **Login saat masa grace period penghapusan akun**: tidak dijelaskan apakah user masih bisa login normal selama 30 hari grace period (Journey 8) sebelum akun benar-benar terhapus — **open question**.
- **Undangan region duplikat**: satu email boleh menerima undangan ke beberapa region berbeda secara bersamaan, tapi tidak boleh menerima 2+ undangan pending ke region yang **sama**.

## Note on soft test 26 Juli 2026: semua alur di atas sudah dites secara soft (manual, tanpa automation) di versi mobile terbaru. Ada catatan bug :

- ~~ketika kirim undangn masuk region, email tidak terkirim. tapi masuk info undangan di akun terkait.~~ **Sudah diimplementasikan** (22 Jul 2026, k-forum-api): consumer MQ `HandleRegionInvitationEmailRequested` ditambahkan + didaftarkan ke router & queue worker — sebelumnya event ini dipublish tapi tak pernah dikonsumsi. Detail: `REGION_ONBOARDING_ISSUES.md` Issue 1. Belum diverifikasi end-to-end pakai SMTP asli (baru sampai unit/handler test).
- ~~mobile: list regions double untuk region dirinya sendiri~~ **Sebagian diperbaiki** (22 Jul 2026, k-forum-api): edge case duplikasi row akibat query `ListRegionsForMobile` (join tanpa filter row terbaru) sudah diperbaiki + `RequestJoinRegionUseCase` sekarang bersihkan row `rejected` lama sebelum request ulang. **Penyebab utama & paling umum masih belum diperbaiki** — ada di sisi `k_forum` mobile (`regions_browse_screen.dart` tidak exclude region sendiri dari list umum). Detail: `REGION_ONBOARDING_ISSUES.md` Issue 2a/2b.
- ~~semua template email (OTP, undangan, reset password) masih pakai template yang seragam, belum disesuaikan dengan reason/region masing-masing. ini masih wajar karena alur email belum diimplementasikan sepenuhnya.~~ **Sudah diimplementasikan** (22 Jul 2026, k-forum-api): reset password sekarang punya method & body sendiri (`SendPasswordResetOTP`, tidak lagi menyebut "registrasi"); email undangan region menampilkan nama region, bukan UUID mentah. Branding/signature konsisten ("Tim k-forum-api" di semua email) masih belum dirapikan — nice-to-have, tidak blocking. Detail: `REGION_ONBOARDING_ISSUES.md` Issue 3.
- account yang dihapus (soft-delete) belum ada implementasi cancel delete,
- ~~account yang login via google belum ada implementasi set password local.~~ **Sudah diimplementasikan** (2026-07-22): endpoint `POST /api/v1/{web,mobile}/auth/set-password` (lihat `API_SPEC_AUTH.md` §12) + UI backoffice di halaman Profile > Password. Endpoint mobile app-side sudah tersedia di backend; UI Flutter app belum (app repo di luar cakupan sesi ini).