# User Settings (Profile & Keamanan Akun) — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Users Settings/USER_SETTINGS_RULES.md`, `Modules/Users Settings/API_SPEC_USER_SETTINGS.md`. Karena modul User Settings sengaja **tidak menduplikasi** pengaturan yang sudah punya rumah di modul lain, journey ini juga merujuk `API SPEC/Mobile/API_SPEC_PROFILE_MEMBERSHIP.md` (edit profil, avatar, ganti email, hapus akun) dan `API SPEC/Mobile/API_SPEC_AUTH.md` (ganti password) untuk melengkapi pengalaman "Pengaturan Akun" yang dilihat member sebagai satu layar utuh.

## Ringkasan Domain

User Settings adalah kumpulan preferensi dan kontrol yang **dimiliki dan diubah sendiri oleh satu user** (self-scope), berbeda dari System Settings yang platform-wide dan dikelola admin. Modul ini secara desain **hanya memiliki satu tabel sungguhan** (`user_settings`: language + theme) — sisanya adalah **agregasi read-through** di atas modul lain: device/timezone dari FCM (`fcm_tokens`/`device_registrations`), logout dari Auth, dan (di luar modul ini tapi ditampilkan di layar yang sama secara UX) profil personal dari modul Profile serta password/email dari Auth & Profile.

Ada satu endpoint agregator, `GET /api/v1/user-settings`, yang meringkas app preferences + linked accounts + device list dalam satu panggilan supaya layar "Pengaturan" utama tidak perlu fetch berkali-kali — tapi update tetap lewat endpoint granular masing-masing. Prinsip inti modul ini: **preference adalah pilihan user** (language, theme — pakai picker), sedangkan **fakta device di-capture otomatis** (timezone — tanpa picker, tidak bisa dipilih manual).

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Edit profil (nama, bio, alamat, data personal lain) | ✅ | ✅ | Sama persis, self-scope |
| Ganti/hapus foto profil (avatar) | ✅ | ✅ | Sama persis |
| Ganti password | ✅ | ✅ | Sama persis |
| Ganti email (dengan verifikasi OTP) | ✅ | ✅ | Sama persis |
| Atur bahasa & tema aplikasi | ✅ | ✅ | Sama persis |
| Lihat & kelola device/session aktif | ✅ | ✅ | Sama persis |
| Lepas (unlink) akun Google tertaut | ✅ | ✅ | Sama persis, dengan guard "harus punya password" |
| Request penghapusan akun | ✅ | ✅ | Sama persis; efek ke data domain lain (mis. konten Pro yang belum tayang) tidak dibahas eksplisit di source doc |
| Request data export | ✅ (Phase 2) | ✅ (Phase 2) | Belum aktif untuk siapa pun, bukan pembatasan tier |

Sesuai `USER_SETTINGS_RULES.md`: "Beda Standard vs Pro **tidak** memengaruhi struktur settings — semua member punya set setting yang sama." Perbedaan plan di platform KAI selalu lewat mekanisme **benefit** (mis. `post_news`, `create_community`), bukan lewat modul User Settings.

## Journey 1: Member Mengelola Profil — 🅢🅟 — 📱 Mobile

1. **Entry point**: member membuka menu Profil/Akun → `GET /api/v1/mobile/profile/me` menampilkan data personal: nama, email, username, avatar, phone, birth_date, gender, bio, marital_status, occupation, interests, address, serta field read-only (`plan`, `roles`, `permissions`, `is_verified`, `created_at`, `last_login`).
2. **Edit data personal**: member ubah field seperti nama, nomor telepon, tanggal lahir, gender, bio, status pernikahan, pekerjaan, minat (interests, maks 10 item), dan alamat lengkap → `PATCH /api/v1/mobile/profile/me` (partial update, kirim field yang berubah saja). Validasi utama: nama min 3/maks 100 karakter, bio maks 500 karakter, interests maks 10 item @ 50 karakter, dan field lain punya validasi format masing-masing (tanggal lahir tidak boleh di masa depan & minimal usia 13 tahun, dll).
3. **Field yang TIDAK bisa diubah lewat endpoint ini**: `email` dan `username` tidak termasuk payload update profil — email punya alur tersendiri dengan verifikasi OTP (lihat [Journey 2](#journey-2-member-mengelola-keamanan-akun--)), dan `permissions`/`roles`/`plan`/`is_verified` murni read-only (dikelola sistem/admin di domain lain: Role-Permission, Subscription, Verification Badge).
4. **Ganti foto profil (avatar)**, alur presigned upload (upload langsung ke S3, tidak lewat backend):
   - `POST /profile/avatar/presign` → dapat `presigned_url` + `s3_path` (prefix `s3:`), berlaku 1 jam.
   - Client upload file langsung ke S3 pakai `presigned_url`.
   - `POST /profile/avatar/confirm` dengan `s3_path` → backend konfirmasi & memperbarui avatar (mengembalikan versi full + thumbnail, di-serve dari CDN).
5. **Hapus foto profil**: `DELETE /profile/avatar` → avatar kembali ke default/null.
6. **Melihat profil sendiri seperti dilihat orang lain**: source doc (`API_SPEC_PROFILE_MEMBERSHIP.md`) tidak mendokumentasikan mode "preview publik" atau endpoint profil publik terpisah — hanya ada `GET /profile/me` (data lengkap milik sendiri). Tidak jelas apakah ada tampilan profil versi terbatas yang dilihat member lain (mis. di Community atau Directory); ini dicatat sebagai gap dokumentasi, bukan diasumsikan ada atau tidak ada.
7. **Lihat keanggotaan**: member juga bisa melihat (read-only dari modul ini) daftar region yang diikuti (`GET /profile/memberships/regions`) dan komunitas yang diikuti berikut role-nya (`GET /profile/memberships/communities`) — keduanya paginated, murni informasional di layar profil.

**Selesai:** profil member ter-update dan langsung terlihat di seluruh platform (mis. nama/avatar tampil di News, Community, Q&A) sesuai data terbaru.

## Journey 2: Member Mengelola Keamanan Akun — 🅢🅟 — 📱 Mobile

1. **Ganti password**: member masuk menu keamanan → isi password lama & password baru → `POST /api/v1/mobile/auth/password/change` (`old_password` + `password`). Backend memverifikasi password lama sebelum mengganti. Tidak ada langkah OTP tambahan untuk aksi ini — cukup mengetahui password lama.
2. **Ganti email** (butuh verifikasi OTP ganda, beda dari ganti password):
   - Member isi email baru + password saat ini → `POST /profile/email/change-request`. Sistem kirim OTP **ke email lama dan email baru sekaligus**, mengembalikan `temp_token` (berlaku 1 jam).
   - Member masukkan kedua OTP (dari email lama & email baru) → `POST /profile/email/verify-change` dengan `temp_token` + `otp_old_email` + `otp_new_email` → email resmi berganti.
   - Rate limit: maksimal 3x request ganti email per 24 jam.
   - Salah password saat request → ditolak 401.
3. **Ganti nomor telepon**: berbeda dari email, source doc **tidak** mendokumentasikan alur verifikasi OTP khusus untuk nomor telepon — `phone` adalah salah satu field biasa di `PATCH /profile/me` (Journey 1), diubah langsung tanpa re-verifikasi. Ini dicatat sebagai asimetri yang eksplisit ada di dokumen sumber (email selalu perlu OTP ganda, telepon tidak), bukan sesuatu yang diasumsikan salah ketik.
4. **Atur bahasa aplikasi**: member pilih bahasa di menu Pengaturan → `PUT /api/v1/user-settings/preferences` dengan `language` (harus salah satu dari `system_languages` aktif; kirim `null` untuk reset ke resolve otomatis). Begitu user override manual, bahasa device (`Accept-Language`/`X-Locale` header) **berhenti** jadi penentu — preference user selalu menang sampai diubah lagi. Bahasa efektif dipakai baik untuk UI (client-side) maupun konten dari server (template email, copy push notifikasi).
5. **Atur tema aplikasi**: `system` (ikut OS, default) / `light` / `dark`, di endpoint yang sama (`PUT /preferences`). Murni preferensi visual, disimpan agar sinkron antar-device.
6. **Timezone**: **tidak ada picker** — device mengirim timezone (format IANA, mis. `Asia/Jakarta`) otomatis saat login/refresh token/register FCM token, tersimpan di kolom device (`fcm_tokens`/`device_registrations`). Dipakai backend untuk semua penjadwalan sensitif-waktu (Do Not Disturb notifikasi, event reminder, community schedule).
7. **Kelola device/session aktif**:
   - `GET /user-settings/devices` menampilkan daftar device yang pernah login (platform, nama device, OS version, timezone, waktu terakhir aktif, serta flag `is_current` untuk device yang sedang dipakai).
   - **Logout satu device**: `DELETE /user-settings/devices/{id}` — device itu berhenti menerima push (token dinonaktifkan). Jika device yang di-logout adalah device sekarang, sesi langsung diakhiri lewat Auth. Untuk device **lain**, pemutusan sesi Auth secara instan **belum sepenuhnya didukung** (kapabilitas revoke per-device di Auth belum ada) — akses device itu baru benar-benar terputus saat access token-nya expired dan refresh ditolak. Ini dicatat sebagai dependency yang belum selesai, bukan bug tersembunyi.
   - **Logout semua device**: `POST /user-settings/devices/logout-all` — mematikan push semua device & mengakhiri sesi device sekarang; efek instan ke device lain juga bergantung pada kapabilitas "revoke-all refresh token" di Auth yang dicatat sebagai dependency belum siap. Kontrak endpoint tetap 200 dan menjalankan bagian yang sudah bisa.
   - Logout **tidak** menghapus data user, hanya memutus akses device.
8. **Kelola akun tertaut (Google)**: `GET /user-settings/linked-accounts` menampilkan status tertaut + email Google. Unlink (`DELETE /linked-accounts/google`) **hanya berhasil jika user sudah punya password aktif** — kalau akun murni login via Google (belum pernah set password), unlink ditolak (409) dan member diarahkan set password dulu lewat menu keamanan. Menautkan akun Google baru (**link**) belum tersedia di Phase 1.
9. **Hapus akun**:
   - Member request hapus akun (isi password + alasan opsional) → `POST /profile/account/delete-request`. Sistem kirim email konfirmasi + OTP, mengembalikan `deletion_token` dan `deletion_scheduled_date` (30 hari dari sekarang).
   - Member konfirmasi dengan OTP dari email → `POST /profile/account/confirm-delete` → akun masuk status "akan dihapus", grace period 30 hari berjalan.
   - Selama 30 hari itu, member masih bisa **batalkan** lewat `POST /profile/account/cancel-delete` (pakai `deletion_token` + OTP yang sama) → akun kembali `active`.
   - Setelah 30 hari tanpa pembatalan, data dihapus permanen (soft delete → hard delete). Rate limit: maksimal 1x request hapus akun per 7 hari.
10. **Data export**: belum tersedia untuk siapa pun — masih Phase 2. Hook status (`requested → processing → ready → expired`) sudah disiapkan di skema, tapi endpoint belum diekspos ke member.

**Selesai:** akun member tetap aman terkontrol olehnya sendiri — baik untuk ganti kredensial, mengatur preferensi tampilan, mengelola device aktif, maupun (kalau diperlukan) mengakhiri keanggotaannya sendiri.

## Keterlibatan Admin — 💻 Web/Backoffice

`USER_SETTINGS_RULES.md` menyatakan eksplisit: **admin/superadmin tidak mengubah User Settings member lewat modul ini** — modul ini murni self-scope, tidak ada permission admin untuk mengedit setting user lain. Aksi administratif atas akun member (suspend, reset password paksa, ubah plan, lihat detail akun) adalah kewenangan domain **User Management** yang terpisah — lihat [15_ADMIN_ONLY_DOMAINS_NOTE.md](15_ADMIN_ONLY_DOMAINS_NOTE.md).

Dengan kata lain: keterlibatan admin di domain User Settings/Profile-Keamanan-Akun untuk member adalah **tidak ada** — ini konsisten dengan tabel ringkasan di [00_OVERVIEW.md](00_OVERVIEW.md) yang menandai domain 14 tanpa keterlibatan admin.

## Di Luar Cakupan Standard & Pro

- **Melihat/mengubah profil atau settings member lain** — semua endpoint bersifat self-scope; tidak ada cara member mengakses data akun user lain lewat domain ini, itu murni ranah admin di User Management.
- **Memulihkan akun setelah 30 hari grace period lewat** — `USER_SETTINGS_RULES.md`/`API_SPEC_PROFILE_MEMBERSHIP.md` tidak menyebutkan mekanisme restore setelah data dihapus permanen; begitu grace period lewat, penghapusan bersifat final.
- **Menautkan (link) akun Google baru** — fitur ini eksplisit "ditunda", belum tersedia di Phase 1 untuk member mana pun.
- **Melakukan data export** — Phase 2, hook disiapkan tapi belum aktif; tidak bisa diakses member saat ini.
- **Menjamin logout instan untuk device lain** — sampai Auth menambah kapabilitas revoke per-device & revoke-all, member bisa memicu logout tapi efek instan ke device lain terbatas pada penghentian push saja; ini keterbatasan platform, bukan sesuatu yang bisa member percepat sendiri.
- **Mengatur permission/role sendiri** — field `roles`/`permissions`/`plan` di profil bersifat read-only; perubahan role/permission adalah ranah Role-Permission system, bukan hak edit member.

## Edge Case & Catatan Tambahan

- **Auto-create default**: baris `user_settings` (language + theme) otomatis dibuat dengan default saat pertama kali diakses — member tidak perlu setup manual apa pun sebelum bisa memakai app, sistem tetap jalan dengan nilai default runtime-safe.
- **Asimetri verifikasi email vs telepon**: ganti email selalu butuh OTP ganda (email lama + email baru) karena dianggap identitas login yang lebih sensitif, sedangkan ganti telepon cukup lewat update profil biasa tanpa OTP — ini pola yang eksplisit ada di dokumen sumber, dicatat di sini supaya tidak dianggap gap dokumentasi.
- **Timezone multi-device**: kalau member aktif di banyak device dengan timezone berbeda (mis. traveling), timezone efektif yang dipakai backend untuk DND/reminder mengikuti device yang **paling terakhir aktif** (last-write-wins) — bukan device spesifik yang memicu suatu aksi real-time, kecuali dicatat lain di domain terkait.
- **Rate limiting endpoint sensitif**: ganti email dibatasi 3x/24 jam, request hapus akun dibatasi 1x/7 hari — mencegah penyalahgunaan/spam pada aksi yang melibatkan pengiriman OTP berulang.
- **Ketergantungan pada modul Auth yang belum lengkap**: fitur "logout all" dan "logout device lain secara instan" baru akan sepenuhnya enforce setelah Auth menambahkan session/refresh store + endpoint revoke-all & revoke-per-device — dicatat eksplisit sebagai dependency terbuka di `USER_SETTINGS_RULES.md` dan `API_SPEC_USER_SETTINGS.md`, bukan diasumsikan sudah selesai.
