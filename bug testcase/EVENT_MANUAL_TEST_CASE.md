# Manual Test Case — Event (End-to-End: Web, API, App)

Dokumen test case manual untuk domain **Event** (kegiatan/forum event) — dipakai untuk persiapan presentasi/demo.
Meng-cover seluruh use case dari **buat event** sampai **notifikasi & reminder** diterima user,
lintas layer **Web (Backoffice)**, **API**, dan **App (Mobile)**.

- **Base URL API:** `https://k-forum-api.yubicom.co.id/api/v1`
- **Web:** `/api/v1/web/events/...` (Backoffice, butuh login Superadmin)
- **Mobile/App:** `/api/v1/mobile/events/...` (dipakai App — App belum ada source code di repo ini, jadi acuan perilaku App = kontrak API mobile ini + spec di `API SPEC/Mobile/API_SPEC_EVENT_MOBILE.md`)
- **Referensi:** `Modules/Event/EVENT_RULES.md`, `API SPEC/Web/API_SPEC_EVENT_BACKOFFICE.md`, `API SPEC/Mobile/API_SPEC_EVENT_MOBILE.md`, `bug testcase/EVENT_ERROR.md`

> ℹ️ **Update 2026-07-06**: dokumen versi awal ditulis setelah membaca kode backend dan menemukan sejumlah gap antara `EVENT_RULES.md`/API spec dan implementasi (settings module tidak ada, reject/cancel tidak mengirim notifikasi, reminder rusak total, gap otorisasi backoffice). Semua temuan itu **sudah diperbaiki** di backend (`internal/domain/event/`, `internal/app/usecase/event/`, `internal/interfaces/http/router/router.go`, `internal/interfaces/mq/`) dan diverifikasi lewat automated handler test (`go test ./internal/interfaces/http/handler/...`) — bukan cuma pembacaan kode. Dokumen ini sudah diupdate supaya expected result mencerminkan perilaku **setelah fix**. Sisa gap yang **belum** diperbaiki (keputusan sadar, bukan terlewat) tetap ditandai 🔴 dan dijelaskan di **Catatan Risiko**.
>
> **Update kedua (hari yang sama)**: 3 bug tambahan ditemukan dari manual test dan sudah diperbaiki: (1) `description` kategori event tidak pernah tersimpan (kolom tidak ada di tabel — lihat CAT-01/CAT-02), (2) banyak error code Event module (termasuk validasi umum & semua error settings baru) belum diterjemahkan, muncul sebagai kode mentah di response (lihat I-09), (3) tidak ada cara "Simpan Draft" sungguhan dari App — create selalu langsung submit/publish; sekarang ada `save_as_draft` di create dan endpoint baru `POST /events/{id}/submit` (lihat EVT-02c/02d/02e).

---

## Daftar Isi

1. [Role & Akun yang Dibutuhkan](#1-role--akun-yang-dibutuhkan)
2. [Peta Use Case](#2-peta-use-case)
3. [A — Event Module Settings (Konfigurasi Global)](#a--event-module-settings-konfigurasi-global)
4. [B — Category Management](#b--category-management)
5. [C — Event Lifecycle: Member Pro (App, butuh approval)](#c--event-lifecycle-member-pro-app-butuh-approval)
6. [D — Event Lifecycle: Superadmin (Backoffice)](#d--event-lifecycle-superadmin-backoffice)
7. [E — Konsumsi Konten: App/Mobile](#e--konsumsi-konten-appmobile)
8. [F — Konsumsi Konten: Web/Backoffice](#f--konsumsi-konten-webbackoffice)
9. [G — Interaksi: Save, Schedule + Kalender, Share (App)](#g--interaksi-save-schedule--kalender-share-app)
10. [H — Notifikasi & Reminder](#h--notifikasi--reminder)
11. [I — Permission & Negative Case](#i--permission--negative-case)
12. [Skenario End-to-End (urutan untuk demo)](#skenario-end-to-end-urutan-untuk-demo)
13. [Catatan Risiko yang Perlu Diwaspadai Saat Demo](#catatan-risiko-yang-perlu-diwaspadai-saat-demo)

---

## 1. Role & Akun yang Dibutuhkan

| Role | Kebutuhan | Bisa apa (ringkas) |
|---|---|---|
| **Superadmin/Usergod** | 1 akun | Satu-satunya role yang bisa akses `/api/v1/web/events/*` (approve/reject/delete/archive/settings/categories) — endpoint ini sekarang digerbang `RequireSuperAdmin()`, admin scope lain (region/community) **tidak** lolos |
| **Member Pro** | 1 akun, punya benefit `create_event` | Submit event (masuk `pending_approval` atau langsung `published` tergantung setting `auto_publish` — lihat bagian A), edit draft/rejected milik sendiri, cancel, **delete draft/rejected milik sendiri** (endpoint baru, lihat EVT-14) |
| **Member Standard** | 1 akun, tanpa benefit `create_event` | Baca, save/bookmark, schedule, share — tidak bisa create event |
| **Guest** | tanpa login | Baca event `published` saja (list/detail/featured/upcoming/categories) |

> Siapkan minimal 2 akun Member untuk test notifikasi broadcast publish (1 dengan preferensi modul "event" ON, 1 OFF) — lihat bagian H.
> Siapkan 1 akun Member Pro tambahan yang men-**schedule** event dengan `reminder_enabled=true` untuk test reminder — sekarang reminder benar-benar terjadwal per-user (lihat H).

---

## 2. Peta Use Case

| Use Case | Web | API (web) | API (mobile) | App |
|---|---|---|---|---|
| Kelola Event Module Settings (auto_publish, dll) | ✅ | ✅ **(baru)** | — | — |
| Kelola Category Event (dengan description) | ✅ | ✅ **(description fixed)** | (read-only) | (read-only, dropdown filter) |
| Simpan sebagai draft (Member Pro) | — | — | ✅ **(baru)** `save_as_draft: true` | ✅ |
| Submit draft ke review/publish | — | — | ✅ **(baru)** `POST /events/{id}/submit` | ✅ |
| Create/Submit event langsung (Member Pro) | — | — | ✅ (auto-publish atau pending_approval sesuai setting) | ✅ |
| Create event langsung published (Superadmin) | ✅ | ✅ | — | — |
| Approve/Reject event submission | ✅ | ✅ | — | — |
| Edit event (draft/rejected only, dengan edit-lock H-N hari) | — | — | ✅ | ✅ |
| Edit event apapun statusnya (Superadmin) | ✅ | ✅ | — | — |
| Cancel event (organizer/admin) | ✅ (force) | ✅ | ✅ | ✅ |
| **Delete event milik sendiri** (Member Pro, draft/rejected) | — | — | ✅ **(baru)** | ✅ |
| Archive event (Superadmin) | ✅ **(baru)** | ✅ **(baru)** | — | — |
| Delete event permanen (Superadmin, wajib archive dulu kalau published) | ✅ | ✅ | — | — |
| Toggle Featured | ✅ | ✅ | (read-only) | (read-only, carousel) |
| Feed & filter event (list/featured/upcoming/search) | (list admin) | ✅ | ✅ (3 endpoint 🔴 pernah 500 — lihat EVENT_ERROR.md, belum re-verifikasi di sesi ini) | ✅ |
| Save/Bookmark event | — | — | ✅ | ✅ |
| Tambah ke Jadwal + reminder + export kalender | — | — | ✅ | ✅ |
| Share event | — | — | ✅ | ✅ |
| Notifikasi broadcast (event baru published) | — | trigger otomatis | — | ✅ terima notif |
| Notifikasi approve ke organizer (unicast/bypass) | — | 🔴 masih cuma broadcast, lihat Catatan Risiko | — | — |
| Notifikasi reject ke organizer | — | ✅ **(fixed)** unicast + reason | — | — |
| Notifikasi cancel ke user yang save/schedule | — | ✅ **(fixed)** multicast | — | — |
| Reminder H-1jam/3jam/1hari/3hari/1minggu sebelum event | — | ✅ **(fixed)** job per-user, dibuat saat user schedule | — | ✅ |

---

## A — Event Module Settings (Konfigurasi Global)

> **Status:** endpoint `GET/PUT /api/v1/web/events/settings` sekarang **terimplementasi penuh** (tabel `event_settings` singleton, domain entity + repository, usecase, handler, route — superadmin only). Sudah dicek juga di sisi Backoffice: halaman `pages/events/index.vue` (tab Settings) sebelumnya sudah dibuat oleh tim frontend tapi nonaktif karena endpoint 500/422 mis-routed (lihat `k-forum-backoffice/docs/BACKEND_ISSUES.md`) — sekarang harusnya langsung berfungsi tanpa perubahan kode frontend, karena field response (`auto_publish`, `require_approval`, dst) sudah didesain match dengan tipe `EventSettings` yang sudah ada di frontend.

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| SET-01 | Login Superadmin | `GET /api/v1/web/events/settings` | `200` — data settings lengkap (`auto_publish`, `require_approval` [=`!auto_publish`, read-only computed], `max_events_per_member_per_month`, `allow_cancel_after_published`, `lock_editing_days_before_event`, `send_reminders`, `reminder_times`, `allow_comments`, `max_images_per_event`, `max_image_size_mb`, `updated_at`) |
| SET-02 | Login non-superadmin (admin region/community, atau member biasa) | `GET/PUT /api/v1/web/events/settings` | `403` |
| SET-03 | Superadmin set `auto_publish=true` via `PUT /settings` | Member Pro submit event baru via `POST /api/v1/mobile/events` | `201` — status **langsung `published`** (bukan `pending_approval`), broadcast notifikasi "Event Baru" langsung terpicu sama seperti alur approve |
| SET-04 | Superadmin set `auto_publish=false` (default) | Member Pro submit event baru | `201` — status `pending_approval`, masuk antrian review Superadmin |
| SET-05 | Superadmin set `max_events_per_member_per_month=2` | Member Pro submit event ke-3 dalam bulan kalender yang sama | `422` `EVENT_MAX_PER_MONTH_EXCEEDED` — dihitung dari `created_at` event pada bulan berjalan, per organizer |
| SET-06 | Setting `max_events_per_member_per_month=0` | Submit event ke berapa pun dalam sebulan | Selalu `201` — `0` berarti unlimited |
| SET-07 | Superadmin set `lock_editing_days_before_event=3`, event `draft` milik Member Pro dengan `event_date` 2 hari lagi | Member Pro `PUT` edit event tsb | `403` `EVENT_EDIT_LOCKED` — draft/rejected pun terkunci kalau tinggal < N hari menuju `event_date` |
| SET-08 | Event `draft` dengan `event_date` masih 10 hari lagi (di luar lock window) | Member Pro edit | `200` — masih boleh diedit |
| SET-09 | Superadmin set `max_images_per_event=3` | Member Pro create/update event dengan 5 images | `422` `DOMAIN_EVENT_IMAGES_EXCEEDED` — validasi sekarang ikut setting (bukan cuma hardcoded 10 di domain) |
| SET-10 | Superadmin set `allow_cancel_after_published=false` | Organizer (bukan admin) coba cancel event `published` miliknya | `403` `EVENT_CANCEL_AFTER_PUBLISHED_DISABLED` — organizer hanya bisa cancel saat masih `draft`. **Tidak berlaku untuk Superadmin force-cancel** (lihat ADM-06, admin selalu bisa cancel apapun) |
| SET-11 | Superadmin set `send_reminders=false` | User schedule event dengan `reminder_enabled=true` | `201` tetap sukses (preferensi tersimpan), tapi **tidak ada job reminder yang dibuat** — kill-switch global |
| SET-12 | — | Cek field `allow_comments` di response | Tersimpan & bisa diubah, tapi **belum ada efek apapun** — module Event memang belum punya fitur comment sama sekali di backend (beda dari News/Community). Field ini murni placeholder untuk kebutuhan masa depan |
| SET-13 | — | Cek field `max_image_size_mb` | Tersimpan & bisa diubah, tapi **belum divalidasi** saat presign upload gambar — lihat Catatan Risiko |

---

## B — Category Management

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| CAT-01 | Login Superadmin | `POST /api/v1/web/events/categories` `{ "name": "Sports", "description": "..." }` | `201`. ✅ **(fixed)** `description` sekarang benar-benar tersimpan dan dikembalikan di response — sebelumnya kolom ini tidak ada sama sekali di tabel/entity/DTO, jadi selalu hilang meski dikirim di request |
| CAT-02 | Kategori ada | `PUT /api/v1/web/events/categories/{id}` ubah nama & `description` | `200` — `description` ikut ter-update |
| CAT-03 | Kategori dipakai event aktif | `DELETE /api/v1/web/events/categories/{id}` | **Cek**: apakah ditolak, atau event terkait jadi tanpa kategori tanpa peringatan (pola serupa CAT-04 di News — lihat Catatan Risiko) |
| CAT-04 | Login non-superadmin | `POST/PUT/DELETE /events/categories` | `403` — endpoint ini ikut ke-gate `RequireSuperAdmin()` bersama seluruh grup `/web/events` |
| CAT-05 | — | `GET /api/v1/mobile/events/categories` (guest, tanpa token) | `200` — daftar kategori untuk dropdown, tanpa perlu login |

---

## C — Event Lifecycle: Member Pro (App, butuh approval)

> Endpoint create di App: `POST /api/v1/mobile/events` — gated `middleware.RequireBenefit("create_event")`. Status hasil create sekarang **tergantung setting `auto_publish`** (lihat bagian A) — default `pending_approval`. ✅ **(fixed, 2 lapis)** Backend sekarang mendukung `save_as_draft: true` untuk minta status draft. **Root cause asli**: app `k_forum` (Flutter) sudah punya UI "Simpan Draft"/"Submit" yang jalan, tapi mengirim field `status: "draft"` / `status: "pending_approval"` (bukan `save_as_draft`) — field yang sebelumnya tidak dibaca sama sekali oleh backend, jadi draft selalu ke-skip. Backend sekarang menerima **keduanya**: `save_as_draft: true` maupun `status: "draft"` dianggap sama (lihat `dto.EventInput.Status` di `event_dto.go`) — dipilih pendekatan ini (bikin backend kompatibel dengan app yang sudah ada) daripada nunggu rilis app baru.

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| EVT-01 | Login Member Standard (tanpa benefit `create_event`) | `POST /api/v1/mobile/events` | `403` — tombol "Buat Event" harusnya disable/tidak muncul di App, muncul ajakan upgrade ke Pro |
| EVT-02 | Login Member Pro, isi form event **offline** lengkap (title, description, category_id, venue_name, venue_address, event_date masa depan, event_time), setting `auto_publish=false` (default) | `POST /api/v1/mobile/events` | `201` — status `pending_approval`, `submitted_at` terisi |
| EVT-02b | Sama seperti EVT-02, tapi setting `auto_publish=true` (lihat SET-03) | `POST /api/v1/mobile/events` | `201` — status **`published`** langsung, `published_at` terisi, broadcast notifikasi "Event Baru" terpicu |
| EVT-02c | Login Member Pro, isi form event lengkap, tambahkan `save_as_draft: true` **atau** `status: "draft"` (payload asli App) | `POST /api/v1/mobile/events` | ✅ **(fixed)** `201` — status **`draft`** (bukan `pending_approval`/`published`), tidak ada notifikasi broadcast, event hanya terlihat organizer sendiri di "Event Saya" |
| EVT-02d | Event `draft` hasil EVT-02c, milik sendiri | `POST /api/v1/mobile/events/{id}/submit` (endpoint baru, tanpa body) | ✅ **(baru)** `200` — event pindah ke `pending_approval` (atau langsung `published` kalau `auto_publish=true`), broadcast notifikasi terpicu kalau auto-published |
| EVT-02e | Event `draft` milik Member Pro lain | `POST /api/v1/mobile/events/{id}/submit` | `403` — ownership check |
| EVT-02f | Event `draft` milik sendiri | App tap "Submit" di layar **edit** (bukan lewat EVT-02d) → `PUT /api/v1/mobile/events/{id}` dengan field sama + `status: "pending_approval"` | ✅ **(fixed)** `200` — event ikut pindah ke `pending_approval` (atau `published` kalau auto-publish). Ini jalur yang **benar-benar dipakai App saat ini** (belum ada tombol yang manggil endpoint `/submit` di EVT-02d — lihat Catatan Risiko) |
| EVT-03 | Sama seperti EVT-02 tapi `event_type: online` tanpa `online_platform`/`online_url` | `POST /api/v1/mobile/events` | `422` — validasi wajib untuk online |
| EVT-04 | `event_type: offline` tanpa `venue_name`/`venue_address` | `POST /api/v1/mobile/events` | `422` — validasi wajib untuk offline |
| EVT-05 | `event_date` di masa lalu | `POST /api/v1/mobile/events` | `422` — event date must be future |
| EVT-06 | `images` array melebihi `max_images_per_event` setting (default 10) | `POST /api/v1/mobile/events` | `422` `CodeEventImagesExceeded` |
| EVT-07 | Event `pending_approval` milik Member Pro | Member Pro coba `PUT /api/v1/mobile/events/{id}` | `403 CodeEventCannotEdit` — terkunci selama review (`CanEdit()` hanya izinkan draft/rejected) |
| EVT-08 | Event `pending_approval` | Member Pro lain (bukan pemilik) coba edit/cancel event tsb | `403 Forbidden` — ownership check (`OrganizerID != organizerID`) |
| EVT-09 | Login Superadmin, event `pending_approval` ada | `POST /api/v1/web/events/{id}/approve` `{ "notes": "..." }` | `200` — status jadi `published`, `published_at` terisi; broadcast notifikasi terpicu (lihat H-01) |
| EVT-10 | Event `published` | Cek ulang `POST /approve` pada event yang sama | `400` "Only pending_approval events can be approved" |
| EVT-11 | Event `pending_approval` | `POST /api/v1/web/events/{id}/reject` `{ "reason": "kurang dari 10 char" }` | `422` — reason min 10 karakter |
| EVT-12 | Event `pending_approval` | `POST /api/v1/web/events/{id}/reject` `{ "reason": "Alamat venue belum lengkap, mohon detail." }` | `200` — status jadi `rejected`, `rejection_reason` tersimpan. ✅ **Organizer sekarang menerima notifikasi** "Event Ditolak" berisi alasan (unicast, `RoutingForumEventRejected`) |
| EVT-13 | Event `rejected` milik Member Pro | Member Pro edit lalu simpan (`PUT`) | `200` — konten berubah, dan **otomatis re-submit** ke `pending_approval` lagi (`UpdateEventUseCase` panggil `Submit()` lagi jika `wasRejected`) |
| EVT-14 | Event `draft`/`rejected` milik sendiri | `DELETE /api/v1/mobile/events/{id}` | ✅ **(endpoint baru)** `200` — event terhapus permanen, gambar ikut ditandai untuk dibersihkan |
| EVT-14b | Event `pending_approval`/`published`/`cancelled` milik sendiri | `DELETE /api/v1/mobile/events/{id}` | `403` — hanya draft/rejected yang boleh dihapus organizer |
| EVT-14c | Event milik Member Pro lain | `DELETE /api/v1/mobile/events/{id}` | `403` — ownership check |
| EVT-15 | Event `published` milik Member Pro, `allow_cancel_after_published=true` (default) | `POST /api/v1/mobile/events/{id}/cancel` `{ "reason": "Venue tidak tersedia" }` | `200` — status `cancelled`. ✅ **Semua user yang sudah save/schedule event ini menerima notifikasi pembatalan** (multicast, `RoutingForumEventCancelled`) |
| EVT-16 | Event `cancelled` | Organizer coba edit event tsb | `403 CodeEventCannotEdit` — cancelled tidak bisa diedit |

---

## D — Event Lifecycle: Superadmin (Backoffice)

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| ADM-01 | Login Superadmin, kategori & images sudah diupload | `POST /api/v1/web/events` status lengkap | `201` — langsung `published` tanpa approval (`create_admin_event.go`) |
| ADM-02 | Event apapun statusnya | `PUT /api/v1/web/events/{id}` | `200` — Superadmin bisa edit event status apapun (`AdminUpdate()` skip `CanEdit()` check, dan tidak terkena `lock_editing_days_before_event` — lock itu hanya berlaku untuk organizer) |
| ADM-03 | Event `published` | `PATCH /api/v1/web/events/{id}/featured` `{ "is_featured": true }` | `200` — event masuk carousel featured mobile |
| ADM-04 | Event `draft`/`pending_approval`/`rejected`/`cancelled` | `PATCH .../featured` `{ "is_featured": true }` | `422` — hanya event `published` yang boleh di-feature |
| ADM-05 | Sudah ada 10 event featured aktif | `PATCH .../featured` `{ "is_featured": true }` pada event ke-11 | `200` tetap sukses + field `warning` di response (soft limit, tidak diblok) |
| ADM-06 | Event apapun (termasuk `published`) | `POST /api/v1/web/events/{id}/cancel` `{ "reason": "..." }` | `200` — status `cancelled`. Admin force-cancel **tidak** terikat `allow_cancel_after_published`. ✅ User yang sudah save/schedule event ini menerima notifikasi pembatalan (multicast) — tapi **organizer sendiri tidak dapat notifikasi khusus** dari force-cancel ini (lihat Catatan Risiko) |
| ADM-07 | Event `published` | `DELETE /api/v1/web/events/{id}` | ✅ **(fixed)** `409 EVENT_CANNOT_DELETE_PUBLISHED` — published harus di-archive dulu sebelum bisa di-hard-delete |
| ADM-07b | Event `draft`/`rejected`/`cancelled`/`archived` | `DELETE /api/v1/web/events/{id}` | `200` — hard delete permanen, gambar ikut ditandai untuk dibersihkan |
| ADM-08 | Event `published` | `POST /api/v1/web/events/{id}/archive` | ✅ **(endpoint baru)** `200` — status jadi `archived`, hilang dari listing publik & search index. Setelah ini, `DELETE` pada event yang sama akan sukses `200` (lihat ADM-07b) |

---

## E — Konsumsi Konten: App/Mobile

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| MOB-01 | Ada event published berbagai tipe | `GET /api/v1/mobile/events` (guest, tanpa token) | `200` — hanya event `published`, `is_saved`/`is_scheduled` tidak muncul/`false` karena guest |
| MOB-02 | — | `GET .../events?event_type=online` | Terfilter sesuai tipe |
| MOB-03 | — | `GET .../events?location=Surabaya` | Terfilter text-match pada `venue_address` |
| MOB-04 | — | `GET .../events?category_id=...&date_from=...&date_to=...` | Filter kombinasi jalan |
| MOB-05 | — | `GET /api/v1/mobile/events/featured?limit=5` | 🔴 Per `EVENT_ERROR.md` (23 Jun 2026) pernah **500 ERR_INTERNAL**. **Belum di-touch di sesi fix ini** (di luar scope temuan notifikasi/settings/otorisasi) — cek ulang status sebelum demo |
| MOB-06 | — | `GET /api/v1/mobile/events/upcoming?limit=10` | 🔴 Sama seperti MOB-05, belum diverifikasi ulang |
| MOB-07 | — | `GET /api/v1/mobile/events/search?q=korea` | 🔴 Sama seperti MOB-05, belum diverifikasi ulang. Catatan lama: `GET /events?search=korea` (list+filter) jalan normal — bisa jadi acuan perbaikan kalau masih error |
| MOB-08 | Login user, sudah save & schedule event X | `GET .../events/{X}` | `is_saved=true`, `is_scheduled=true` |
| MOB-09 | Guest | `GET .../events/{X}` | Tetap bisa baca detail, `is_saved`/`is_scheduled` default `false` |
| MOB-10 | — | `GET /api/v1/mobile/events/categories` | `200` — daftar kategori, auth optional |
| MOB-11 | App: buka tab Events | Scroll feed, ganti filter tipe/lokasi, buka detail, kembali ke list | Tidak ada crash; state filter & pagination konsisten |
| MOB-12 | — | Upload gambar sesuai flow presign: `POST /mobile/events/media/image/presign` → upload S3 → `POST /mobile/events/media/image/confirm` | Endpoint ini scoped ke `/events/media/image/*` (bukan `/mobile/media/presign` generic seperti disebut di draft awal `API_SPEC_EVENT_MOBILE.md`) — perbedaan ini murni dokumentasi spec vs implementasi, sudah dikonfirmasi rute aktual, tim mobile harus ikut rute ini |

---

## F — Konsumsi Konten: Web/Backoffice

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| WEB-01 | Campuran event semua status | `GET /api/v1/web/events` tanpa filter | Semua status muncul (admin lihat semua), plus `summary{total_pending,total_published,total_rejected,total_cancelled,total_featured}` |
| WEB-02 | — | `GET /api/v1/web/events/pending?sort=submitted_at` | Antrian review, oldest first — ini yang jadi "antrian approval" utama Superadmin |
| WEB-03 | — | `GET /api/v1/web/events?organizer_id=...` | Filter by organizer tertentu |
| WEB-04 | — | `GET /api/v1/web/events?is_featured=true` | Hanya event yang sedang di-feature |
| WEB-05 | — | Buka form create/edit event di Backoffice | Validasi field wajib (`category_id`, `event_type`, venue/online sesuai tipe, `event_date`, `event_time`) jalan di UI sebelum submit |
| WEB-06 | Login Superadmin | Buka tab **Settings** di halaman Events Backoffice | Form settings sekarang termuat dari `GET /events/settings` (bukan fallback default lagi) dan bisa disimpan via `PUT` — cek toggle `auto_publish`/`require_approval` saling terbalik otomatis |

---

## G — Interaksi: Save, Schedule + Kalender, Share (App)

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| INT-01 | Login user, event belum di-save | `POST .../events/{id}/save` `{ "note": "..." }` | `201` — masuk daftar bookmark |
| INT-02 | Sudah save | `POST` save lagi | Cek idempoten atau `409` duplicate — pastikan tidak nyimpen entry dobel |
| INT-03 | Sudah save | `DELETE .../events/{id}/save` | `200` — hilang dari `GET /events/me/saved` |
| INT-04 | Login user, belum schedule event X (published, `event_date` cukup jauh agar reminder punya waktu terjadwal) | `POST .../events/{id}/schedule` `{ "reminder_enabled": true, "reminder_time": "1 day before", "personal_notes": "...", "export_calendar": true }` | `201` — masuk "Jadwalku", response menyertakan `calendar_export{ics_url, google_calendar_url, apple_calendar_url}`. ✅ **Job reminder nyata dibuat** di belakang layar, dijadwalkan tepat 1 hari sebelum `event_date`+`event_time`, ditujukan ke user ini (lihat bagian H) |
| INT-04b | Event `published` dengan `event_date` besok jam 08:00, sekarang sudah jam 09:00 hari ini (offset "1 day before" jatuh di masa lalu) | `POST .../events/{id}/schedule` dengan `reminder_time: "1 day before"` | `201` tetap sukses (schedule tersimpan), tapi **reminder job tidak dibuat** — sistem skip kalau target waktu reminder sudah lewat |
| INT-05 | Sudah schedule event yang sama | `POST .../events/{id}/schedule` lagi | `409` Conflict (`ScheduleEventUseCase` cek `FindByEventAndUser` dulu) |
| INT-06 | `export_calendar` tidak dikirim/`false` | `POST .../events/{id}/schedule` | `201` — field `calendar_export` **tidak muncul** di response |
| INT-07 | Sudah schedule dengan `export_calendar` | `GET .../events/{id}/schedule/calendar` | `200` — bisa ambil ulang link kalender kapan saja |
| INT-08 | Event tidak ada di jadwal user | `GET .../events/{id}/schedule/calendar` | `404` "Event not in your schedule" |
| INT-09 | Sudah schedule | `DELETE .../events/{id}/schedule` | `200` — hilang dari `GET /events/me/schedule`. **Catatan**: reminder job yang sudah terlanjur terjadwal (INT-04) **tidak ikut dibatalkan** — lihat Catatan Risiko |
| INT-10 | Login user | `POST .../events/{id}/share` `{ "share_method": "whatsapp", "message": "..." }` | `201` — `share_count` di detail event naik, `share_link` dengan ref user dikembalikan |
| INT-11 | Guest (tanpa login) | Coba save/schedule/share | `401` — guest hanya boleh baca |

---

## H — Notifikasi & Reminder

> ✅ Bagian ini adalah area yang paling banyak diperbaiki. Reject, cancel, dan reminder sekarang benar-benar mengirim notifikasi (diverifikasi lewat automated test di `internal/interfaces/http/handler/{mobile,web}/event_handler_test.go` dan build/vet backend — bukan cuma baca kode). Approve tetap hanya trigger broadcast (bukan unicast/bypass khusus ke organizer) — ini keputusan sadar untuk membatasi scope perbaikan pada bug yang jelas (reject/cancel/reminder), bukan fitur baru (bypass notification key). Tetap disarankan verifikasi ulang secara live sebelum demo besar.

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| NOTIF-01 | 2 akun member: User A (preferensi modul "event" **ON**), User B (**OFF**) | Superadmin approve event pending (EVT-09) atau create langsung published (ADM-01/EVT-02b) | User A menerima notifikasi in-app + push "Event Baru" (`RoutingForumEventPublished`, broadcast). User B **tidak** menerima (opt-out bekerja) |
| NOTIF-02 | Event `pending_approval` milik Member Pro, preferensi modul "event" organizer **dimatikan** | Superadmin approve | 🔴 Organizer **hanya** dapat notifikasi kalau preferensi modul "event"-nya **ON** — belum ada bypass toggle khusus approve untuk Event (beda dari News yang punya `BypassEventKeys` untuk `news.article_approved`). Kalau organizer OFF, organizer tidak tahu event-nya sudah tayang kecuali cek manual ke "Event Saya". **Ini bukan bug baru, tapi gap yang sengaja belum ditutup** — lihat Catatan Risiko |
| NOTIF-03 | Event `pending_approval` | Superadmin **reject** dengan reason (EVT-12) | ✅ **(fixed)** Organizer menerima notifikasi in-app + push "Event Ditolak" berisi alasan reject (unicast, `RoutingForumEventRejected` → `HandleForumEventRejected`) |
| NOTIF-04 | Event `published` sudah di-save/schedule oleh beberapa user | Organizer atau Superadmin **cancel** event tsb (EVT-15/ADM-06) | ✅ **(fixed)** Semua user yang sudah save dan/atau schedule event tsb menerima notifikasi "Event Dibatalkan" (multicast ke union user_id dari `event_saves` + `event_schedules`, `RoutingForumEventCancelled` → `HandleForumEventCancelled`). Kalau tidak ada satupun yang save/schedule, tidak ada notifikasi terkirim (memang tidak ada penerima) |
| NOTIF-05 | User schedule event dengan `reminder_enabled=true`, `reminder_time="1 day before"` (INT-04), event masih jauh di masa depan | Tunggu sampai waktu reminder yang dipilih user (1 hari sebelum `event_date`+`event_time`) | ✅ **(fixed)** User menerima notifikasi push "Pengingat Event" — job dibuat per-user tepat saat `POST /schedule` dipanggil (bukan lagi satu job broken dengan `UserID` kosong di waktu approve) |
| NOTIF-06 | Setting `send_reminders=false` (SET-11) | User schedule event dengan reminder aktif | Tidak ada job reminder dibuat — kill-switch global bekerja, tapi preferensi user tetap tersimpan di DB |
| NOTIF-07 | User unschedule event yang reminder job-nya sudah terlanjur dibuat (INT-09) | Tunggu waktu reminder yang seharusnya | 🔴 Reminder **tetap terkirim** walau user sudah unschedule — job yang sudah dibuat di `scheduled_jobs` tidak otomatis dibatalkan saat unschedule/cancel event. Edge case minor, lihat Catatan Risiko |

---

## I — Permission & Negative Case

| ID | Steps | Expected Result |
|---|---|---|
| I-01 | Login user dengan `AdminScopes` non-kosong tapi bukan role `superadmin`/`usergod` (misal admin region/community) | Akses endpoint `/api/v1/web/events/*` (approve/reject/delete/settings/categories) | ✅ **(fixed)** `403` — grup route `/web/events` sekarang digerbang `middleware.RequireSuperAdmin()` (cek role `superadmin`/`usergod` secara eksplisit), bukan `RequireAdmin()` lagi. Admin non-superadmin **tidak lagi lolos** |
| I-02 | Login Member Standard | `POST /api/v1/mobile/events` | `403` — benefit `create_event` tidak ada |
| I-03 | Login Member Pro, coba edit event **milik Member Pro lain** | `PUT /api/v1/mobile/events/{id_bukan_miliknya}` | `403` — ownership check |
| I-04 | Login Member Pro, coba cancel event **milik orang lain** | `POST .../events/{id}/cancel` | `403` — ownership check |
| I-05 | Guest (tanpa login) | `POST/PUT/DELETE` apapun ke `/mobile/events/*` yang butuh auth | `401` |
| I-06 | Kirim payload dengan field asing `organizer_id`/`status: published`/`is_featured: true` saat create event dari mobile | `POST /api/v1/mobile/events` | Field tersebut harus diabaikan backend — status ditentukan setting `auto_publish` (bukan dari body), `organizer_id` diambil dari token |
| I-07 | Kirim XSS payload (`<script>alert(1)</script>`) di `title`/`description`/`personal_notes` | — | Pastikan tampil ter-escape di Web & App |
| I-08 | Superadmin `DELETE` event `published` yang masih disave banyak user | — | ✅ **(fixed)** `409 EVENT_CANNOT_DELETE_PUBLISHED` — wajib archive dulu (lihat ADM-07/ADM-08), sekarang konsisten dengan pola News |
| I-09 | Ganti header `Accept-Language`/locale user ke `id` (default) | Picu error validasi/domain apapun di module Event (mis. EVT-03, EVT-06, SET-05) | ✅ **(fixed)** Response `errors` sekarang berisi kalimat Bahasa Indonesia yang jelas (mis. "Platform online wajib diisi."), **bukan lagi kode mentah** seperti `DOMAIN_EVENT_ONLINE_PLATFORM_REQUIRED` — sebelumnya banyak error code Event module (termasuk semua yang baru ditambahkan untuk settings) belum masuk ke file terjemahan `locales/id.json`, jadi fallback ke kode itu sendiri. Sudah dilengkapi untuk `id`, `en`, dan `ko` |

---

## Skenario End-to-End (urutan untuk demo)

### Flow 1 — Member Pro Submit → Approve → Tampil di App → Notifikasi Broadcast
1. **[Web]** Superadmin pastikan setting `auto_publish=false` (default) di tab Settings.
2. **[App]** Login Member Pro → buka "Buat Event" → isi form offline lengkap → submit.
3. **[API]** Verifikasi `status=pending_approval`.
4. **[Web]** Login Superadmin → buka `Events — Pending Approval` → event Member Pro muncul di antrian.
5. **[Web]** Klik **Approve**.
6. **[App]** Feed publik → event sekarang muncul.
7. **[App]** Device User A (preferensi "event" ON) → cek notifikasi "Event Baru" masuk.
8. **[App]** Device User B (preferensi OFF) → tidak ada notifikasi.

### Flow 1b — Auto-Publish Aktif
1. **[Web]** Superadmin set `auto_publish=true` di tab Settings.
2. **[App]** Member Pro submit event baru → verifikasi `status=published` langsung (tanpa lewat antrian approval).
3. **[App]** Cek notifikasi broadcast "Event Baru" langsung masuk ke user lain.
4. **[Web]** Kembalikan `auto_publish=false` setelah demo selesai (supaya Flow 1 tetap bisa didemokan ulang).

### Flow 2 — Submission Ditolak
1. **[App]** Member Pro submit event lain.
2. **[Web]** Superadmin **Reject** dengan alasan jelas.
3. **[App]** Cek notifikasi masuk ke organizer — sekarang **notifikasi "Event Ditolak" berisi alasan seharusnya masuk**.
4. **[App]** Organizer buka event `rejected`, edit, simpan → otomatis balik ke `pending_approval` (auto-resubmit, EVT-13).

### Flow 3 — Cancel Event yang Sudah Disave/Schedule Banyak User
1. **[App]** User A save event X, User B schedule event X dengan reminder aktif.
2. **[App]** Organizer buka event X (published) → **Batalkan Event** dengan alasan.
3. **[App]** Cek device User A & User B — **notifikasi pembatalan seharusnya masuk ke keduanya**.
4. **[Web]** Cek `GET /web/events/{id}` — status `cancelled`.

### Flow 4 — Reminder Event
1. **[App]** User schedule event dengan `reminder_time: "1 day before"`, event dijadwalkan cukup jauh di masa depan.
2. Tunggu/simulasikan waktu H-1 hari sebelum event (atau cek tabel `scheduled_jobs` untuk verifikasi job dengan `reference_id = "<event_id>:<user_id>"` sudah tercatat dengan `next_run_at` yang benar, tanpa perlu menunggu real-time penuh).
3. Cek notifikasi push masuk pada waktunya.

### Flow 5 — Superadmin Create Langsung Published + Featured + Archive
1. **[Web]** Superadmin buat event baru lengkap → langsung `published` tanpa approval.
2. **[Web]** Toggle **Featured** pada event tsb.
3. **[App]** Cek carousel/banner featured di home → event muncul.
4. **[Web]** Buat 10 event lain jadi featured, lalu feature-kan 1 lagi (ke-11) → pastikan tetap sukses dengan `warning` di response (ADM-05), bukan ditolak.
5. **[Web]** Coba `DELETE` event published tsb → `409`, harus archive dulu.
6. **[Web]** `POST .../archive` pada event tsb → `200`, status jadi `archived`.
7. **[Web]** `DELETE` lagi → `200`, berhasil terhapus permanen.

### Flow 6 — Member Pro Hapus Event Draft Sendiri
1. **[App]** Member Pro buat event, jangan submit dulu (atau submit lalu ditolak Superadmin agar balik ke status yang bisa dihapus — cek apakah App punya jalur "simpan sebagai draft" tanpa submit).
2. **[App]** Buka event tsb (status draft/rejected) → **Hapus**.
3. **[API]** Verifikasi `DELETE /api/v1/mobile/events/{id}` → `200`, event hilang dari "Event Saya".

---

## Catatan Risiko yang Perlu Diwaspadai Saat Demo

- **[CRITICAL, fixed 2026-07-07] Reminder job & calendar export event ditambah 7 jam dari yang diinput (bug timezone WIB vs UTC)**: dilaporkan user langsung ("bug save schedule event, ditambah 7 jam dari yang diinputkan"). Root cause: `event_date` kolom Postgres bertipe `DATE` murni (timezone-naive — pgx selalu scan balik dengan `Location() == UTC` apapun isinya), dan `event_time` cuma string `"HH:mm"` tanpa info zona — padahal keduanya **wall-clock WIB** (organizer input jam lokal Indonesia, bukan UTC — konvensi yang sama seperti field DND di `module_preference_gate.go`). `combineDateAndTime()` (`helpers.go`, dipakai `ScheduleEventUseCase.scheduleReminderJob`) dan `GetCalendarExportUseCase` (`get_calendar_export.go`) menggabungkan date+time pakai `eventDate.Location()` / `time.Parse` biasa — keduanya jatuh ke UTC — jadi jam WIB yang diinput organizer (misal "19:00") disalahlabel sebagai "19:00 UTC" alih-alih dikonversi jadi "12:00 UTC". Dampak nyata:
  1. **Reminder job selalu meleset 7 jam LEBIH LAMBAT** dari waktu yang seharusnya (`next_run_at` dihitung dari instant yang salah, sehingga notifikasi "Pengingat Event" terkirim 7 jam lebih telat dari jadwal offset yang dipilih user — kadang bahkan reminder terlewat/tidak terkirim sama sekali kalau offset-nya kurang dari 7 jam, karena target waktu jadi di masa lalu).
  2. **Export kalender Outlook (`startdt`/`enddt`) menunjukkan jam yang salah** — RFC3339 dengan label "Z" (UTC) dari jam yang sebenarnya WIB, jadi kalau dibuka user di WIB, Outlook konversi balik jadi +7 jam dari jam asli event.
  - Fixed: tambah `jakartaLocation` (pola sama seperti `module_preference_gate.go`) di `internal/app/usecase/event/helpers.go`, dipakai `combineDateAndTime()` (bukan `eventDate.Location()`) dan `GetCalendarExportUseCase` (`time.ParseInLocation` bukan `time.Parse` biasa).
  - **Bonus bug ke-2 yang ketemu & ikut fixed sekaligus**: setelah field waktu Outlook jadi punya offset asli (`+07:00`, bukan `Z`), ketahuan `outlookURL` di `get_calendar_export.go` **tidak pernah** `url.QueryEscape()` field `startdt`/`enddt` — karakter `+` mentah di query string di-decode balik jadi spasi oleh parser standar (`application/x-www-form-urlencoded`), merusak offset zona jadi `"...10:00:00 07:00"` (spasi, bukan `+`). Bug ini sudah ada dari awal tapi **tidak pernah kelihatan** selama field-nya masih format UTC (`Z`, tanpa karakter `+`) — baru muncul setelah fix timezone di atas. Fixed: bungkus `startTime.Format(time.RFC3339)`/`endTime.Format(...)` dengan `url.QueryEscape()`, konsisten dengan field lain (title/description/location) di `fmt.Sprintf` yang sama.
  - **Bonus bug ke-3 (route mismatch, pola berulang ke-3x di sesi ini)**: sekalian audit endpoint ini, ketemu lagi `testhelper/testserver.go` daftar `GetCalendarExport` di path **`/calendar-export`** (dan grup **tanpa-auth**!) — production (`router.go`) pakai path **`/schedule/calendar`** dan **wajib login**. App (`k_forum`) sudah benar manggil `/schedule/calendar`. Test yang jalan sebelumnya (kalau ada) diam-diam test path+auth yang salah, tidak pernah nyentuh behavior production sesungguhnya (termasuk tidak pernah menguji bahwa endpoint ini benar-benar butuh login). Fixed: `testserver.go` dipindah ke grup authenticated dengan path yang benar, swagger comment ikut disamakan. **Pola ini sudah 3x ketemu di sesi ini** (lihat juga temuan `/me/schedule` vs `/my/schedule`) — kesimpulan: `testserver.go` TIDAK reuse `router.go`, jadi rawan menyimpang diam-diam kalau route production diubah tanpa disadari; audit endpoint apapun ke depan WAJIB diff kedua file.
  - **Test baru**: `TestMobileEvent_ScheduleEvent_ReminderJobUsesCorrectTimezone` (assert `next_run_at` di tabel `scheduled_jobs` = instant UTC yang benar, bukan +7 jam), `TestMobileEvent_GetCalendarExport_CorrectTimezone` (assert `startdt` Outlook URL punya offset `+07:00` yang benar & ter-escape dengan benar), `TestMobileEvent_GetCalendarExport_Unauthenticated`. Semua pass, `go build`/`go vet` bersih.
  - **Catatan**: fix ini HANYA untuk reminder job & calendar export. `lock_editing_days_before_event` (SET) di `update_event.go` juga pakai `ev.EventDate` mentah untuk hitung window hari — **tidak disentuh** karena granularitasnya per-hari (selisih 7 jam nyaris tidak berdampak di perhitungan level hari), dan tidak match keluhan spesifik user (bukan bagian dari laporan ini).
- **[fixed 2026-07-07] `GET /mobile/events/me/schedule` response tidak seragam dengan list event mobile lain**: `ListMyScheduleUseCase` cuma balikin record schedule mentah (`{id, event_id, reminder_enabled, reminder_time, personal_notes, status, created_at}`) — tidak ada satupun field kartu event (title, cover_image, venue, category, organizer) padahal read model (`MyScheduleReadItem`) sebenarnya sudah fetch semua itu (embed `EventListReadItem`), cuma dibuang saat mapping ke DTO. Ini beda pola dari `ListMySavedEventsUseCase` (endpoint kembar untuk "saved events") yang sudah benar pakai helper `toEventListItem()`. Dampak nyata di App: `EventScheduleItemModel.fromJson` (`k_forum/lib/features/event/data/models/event_schedule_model.dart`) baca `json['event']` (nested) untuk parse kartu event — karena backend tidak pernah kirim key `event`, App fallback parse objek schedule itu sendiri sebagai event (`eventJson = json`), jadi field event (title, dst) semua kosong, dan `event.id` App malah kebaca dari `id` schedule (bukan `event_id` yang benar) — record scheduled event tampil rusak total di layar "Jadwalku". Fixed: `dto.EventScheduleResponse` sekarang nest `Event EventListItem` (json key `event`), field `status` di top-level tetap punya arti **status schedule** (`scheduled`/`done`/`skipped`) — bukan status event, sengaja tidak di-flatten biar tidak collision. `ListMyScheduleUseCase` sekarang pakai `toEventListItem()` yang sama seperti endpoint list lain. **Response POST `/events/{id}/schedule` (bukan GET list)** sengaja dibiarkan `event: {id}` saja (title dkk kosong) — dicek ke source App, endpoint itu cuma dipakai App untuk baca field `message`, tidak parse body sebagai `EventScheduleItemModel`, jadi tidak ada gunanya query tambahan buat isi field yang App sendiri tidak baca.
  - **Bonus temuan (fixed juga)**: route production `GET /api/v1/mobile/events/me/schedule` (`router.go`) ternyata **beda path** dari yang dipakai testhelper (`/my/schedule`) dan swagger annotation (`/my/schedule`) — App sendiri sudah benar manggil `/me/schedule` (cocok dengan production), tapi test harness diam-diam test path yang salah/tidak pernah nyentuh route production yang sesungguhnya. Diperbaiki: `testhelper/testserver.go` dan swagger comment disamakan ke `/me/schedule`. **Pelajaran**: ini kejadian ke-2 di sesi ini `testserver.go` route table nyimpang dari `router.go` — kalau audit endpoint manapun di masa depan, selalu diff kedua file itu, jangan percaya salah satu saja.
  - **Belum diperbaiki (di App, bukan API, di luar scope task ini)**: `event_repository_impl.dart`'s `getMySchedule()` baca `raw['pagination']` untuk info paging — field yang **tidak pernah dikirim backend** (backend selalu balas `{data, meta}` standar, pola bug yang sama seperti sudah ditemukan berulang di News/Event Backoffice, lihat [[project_backoffice_event_pagination_image_bug]]). Kalau ada laporan "load more/pagination Jadwalku gak jalan di App", ini kemungkinan besar penyebabnya — perlu App-side fix (baca `meta`, bukan `pagination`), tidak disentuh di sesi ini karena fokusnya response shape, dan App-side edit sebelumnya di sesi ini pernah di-revert user.
- **[CRITICAL, fixed 2026-07-07] MQ worker: registry overwrite + queue binding hilang — NOTIF-01/03/04 sempat tidak jalan end-to-end di production**: ditemukan saat verifikasi ulang tabel NOTIF-01..04 di atas, bukan dari automated test (test harness `testhelper.TestServer` tidak wire RabbitMQ sama sekali — `find internal/interfaces/mq -iname "*_test.go"` kosong, jadi klaim "sudah fixed" sebelumnya cuma benar sampai level usecase/handler, belum diverifikasi sampai worker beneran consume dari queue). 3 gap di `cmd/worker/main.go`:
  1. **Registry di-share antar role, saling timpa**: `Registry.Register` cuma `map[routing]handler` biasa. `RegisterAll()` (notifikasi) dan `RegisterSearchSync()` (search indexing) dipanggil ke **registry yang sama**, dan beberapa routing key dipakai kedua role (`forumevent.new.published`, `forumevent.cancelled`, `communityevent.RoutingCommunityPostCreated`) — pemanggilan kedua diam-diam menimpa handler pertama, jadi role yang kalah race **tidak pernah jalan sama sekali**, walau queue-nya sendiri tetap menerima pesan (AMQP topic exchange fan-out ke semua queue yang bind, `msg.Type` di-set sama persis untuk semua — lihat `outbox_relay.go`). Ini artinya notifikasi "Event Baru" (NOTIF-01) berpotensi kalah race melawan search-reindex untuk routing key yang sama tergantung urutan register. Fixed: split jadi `notifReg`/`notifMiddleware` (khusus `RegisterAll`) dan `searchReg`/`searchMiddleware` (khusus `RegisterSearchSync`), masing-masing registry independen.
  2. **Binding queue untuk `forumevent.rejected` tidak ada sama sekali**: `registerConsumers()` punya list `{queue, routing}` yang benar-benar di-`consumer.Consume(...)` — dan routing `RoutingForumEventRejected` (dipakai fix NOTIF-03 sesi sebelumnya) **tidak terdaftar di list ini**. Artinya walau handler `HandleForumEventRejected` sudah benar dan outbox entry-nya sudah dipublish, tidak ada queue yang consume pesan itu sama sekali — event reject organizer notification **tidak pernah terkirim** di worker sungguhan. Fixed: tambah `{forumevent.QueueForumEventRejected, forumevent.RoutingForumEventRejected}` ke binding list.
  3. **Binding queue notifikasi untuk `forumevent.cancelled` juga tidak ada**: hanya ada binding search-role (`QueueSearchForumEventCancelled`) untuk routing ini; binding notifikasi (yang dipakai `HandleForumEventCancelled`, fix NOTIF-04) tidak pernah didaftarkan. Sama nasibnya dengan poin 2 — cancel notification tidak pernah sampai ke worker manapun. Fixed: tambah binding notifikasi terpisah.
  - Plus, terkait: `event.rejected`/`event.cancelled`/`event.reminder` belum masuk `BypassEventKeys` (`internal/domain/notification/service/module_preference_gate.go`) — kalau user/organizer mematikan preferensi modul "event", notifikasi wajib ini (hasil review + pembatalan event yang sudah di-save) ikut ke-block padahal seharusnya selalu tembus (pola sama seperti `news.article_rejected`, `community.member_kicked/banned`). Sudah ditambahkan.
  - **Verifikasi**: `go build ./...` dan `go vet ./...` bersih setelah fix (dicek di sesi ini). **Belum** diverifikasi live end-to-end dengan RabbitMQ+worker+App sungguhan (di luar kemampuan sandbox ini) — rekomendasi: sebelum demo, jalankan worker asli lalu trigger reject & cancel event sungguhan, konfirmasi notifikasi masuk ke device organizer/user.
- **Backoffice bug (fixed) — pagination "All Events"/"Review Queue" & gambar tidak muncul saat edit**: dua bug di `k-forum-backoffice` (`app/pages/events/index.vue`, `app/stores/eventStore.ts`), bukan di API.
  1. **Pagination cuma dummy di client**: kedua tab tidak pernah mengirim `page`/`limit` ke API — "All Events" mem-paginate ulang 10 baris pertama yang sudah kepalang termuat (pakai TanStack row model lokal), "Review Queue" malah tidak ada UI pagination sama sekali. Ditambah, `eventStore.ts` membaca `res.pagination`/`res.summary` yang **tidak pernah dikirim backend** (backend selalu balas `{data, meta}` sesuai standar `k-forum-api`, bukan `{data, pagination, summary}` seperti didokumentasikan di API spec yang sudah basi) — jadi walau pagination-nya dibetulkan, total count tetap 0 sebelum fix ini. Fixed: `eventStore.ts` sekarang punya `toMeta()` (pola sama seperti `newsStore.ts`) yang baca `res.meta`, kedua tab sekarang kirim `page`/`limit` sungguhan dan pagination-nya server-driven (pola sama seperti `news/articles/index.vue`). Bonus fix: parameter sort juga salah kirim (`sort=-created_at`) padahal backend expect `sort_by`+`sort_type` terpisah — sudah diperbaiki juga.
  2. **Gambar existing tidak muncul saat edit event**: form edit isi `eventForm.images` dari row `EventBackofficeListItem` (item tabel list) yang memang **tidak punya field images sama sekali** di backend (cuma `cover_image`) — jadi galeri gambar selalu kosong pas dibuka utk diedit, entah gambarnya S3 key atau URL eksternal. Fixed: `openEditSlideover` sekarang fetch detail penuh (`eventStore.fetchEventDetail`, yang sudah mengembalikan `images`/`images_raw`) sebelum isi form, plus tambah `imagesPreview[]` sebagai fallback URL supaya `getPreviewUrl()` bisa preview S3 key lama (yang tidak ada di cache upload sesi ini) — pola persis seperti `ArticleForm.vue` (News) yang sudah lebih dulu benar.
  - **Catatan sisa (belum diperbaiki, di luar scope 2 bug di atas)**: KPI card "Pending/Published/Rejected/Cancelled" di tab "All Events" akan tetap menampilkan 0 — field `summary{}` yang jadi sumbernya memang belum pernah diimplementasikan di backend sama sekali (bukan cuma salah baca field), butuh kerjaan backend terpisah kalau mau dibetulkan.
- **App bug (fixed) — `description`/`venue_name`/`venue_address` tampak kosong di App**: ini bug di App (`/home/yubi/Data/projects/k_forum`), **bukan** di API — sudah dikonfirmasi lewat test langsung ke backend bahwa `GET /mobile/events/{id}` selalu mengembalikan ketiga field itu dengan benar. Dua root cause terpisah, keduanya sudah diperbaiki:
  1. `EventModel._location` (`lib/features/event/data/models/event_model.dart`) cuma parse `json['location']` (objek nested) yang **tidak pernah** dikirim k-forum-api — venue selalu dikirim flat sebagai `venue_name`/`venue_address`. Akibatnya `event.location` selalu `null` di App manapun (detail, edit form, bahkan card di list). Fixed: `_location()` sekarang fallback baca `venue_name`/`venue_address` flat kalau tidak ada objek `location` nested.
  2. Layar edit event (`event_form_screen.dart`) prefill dari objek `Event` yang dilempar via router `extra` dari layar **list** ("Event Saya") — response list API memang sengaja tidak menyertakan `description` (beda dari detail). Fixed: `EventFormScreen` sekarang selalu re-fetch `GetEventDetailUseCase` saat masuk mode edit, lalu prefill ulang dari data lengkap.
- **EVT-02d vs EVT-02f — Endpoint `/submit` belum dipakai App**: dicek langsung ke source `k_forum` (Flutter, `lib/features/event/`) — App **tidak** punya tombol/kode yang manggil `POST /mobile/events/{id}/submit`. Alur "Submit" App yang sebenarnya jalan adalah: buka draft di layar **edit** → tombol Submit → `PUT /mobile/events/{id}` dengan `status: "pending_approval"` (jalur EVT-02f). Endpoint `/submit` (EVT-02d) tetap dipertahankan sebagai API yang lebih eksplisit/benar untuk klien masa depan atau App versi berikutnya, tapi App versi saat ini tidak memanggilnya. Root cause & fix detail: `CreateEventData.toJson()` di app mengirim `status`, bukan `save_as_draft` — backend sekarang menerima keduanya (lihat catatan di bagian C).
- **NOTIF-02 — Approve masih belum ada bypass notification khusus ke organizer**: berbeda dari News (`news.article_approved` masuk `BypassEventKeys`, selalu terkirim walau preferensi dimatikan), Event approve cuma memicu broadcast `event.new_published` yang tunduk pada preferensi modul "event" organizer sendiri. Kalau organizer mematikan notifikasi modul "event", dia tidak akan tahu event-nya sudah tayang kecuali cek manual. **Ini keputusan scope, bukan bug yang terlewat** — perbaikan penuh (bypass key khusus) butuh desain terpisah kalau memang diperlukan.
- **NOTIF-07 — Reminder tidak ikut dibatalkan saat unschedule/cancel**: job reminder yang sudah terlanjur dibuat saat `POST /schedule` tidak dihapus/dibatalkan kalau user kemudian `DELETE /schedule` (unschedule) atau event-nya dibatalkan. User bisa saja masih menerima "Pengingat Event" untuk event yang sudah tidak relevan lagi buat dia. Dampak minor (satu notifikasi nyasar), tidak fatal.
- **SET-13 — `max_image_size_mb` belum divalidasi saat upload**: field tersimpan di settings dan bisa diubah lewat UI, tapi endpoint presign gambar event (`GetEventImagePresignURLUseCase`) belum membaca setting ini untuk membatasi ukuran file. Validasi ukuran (kalau ada) masih bergantung pada implementasi storage/S3 di luar kontrol setting ini.
- **CAT-03 — Delete category tanpa cek referensi**: belum diverifikasi apakah hapus kategori yang masih dipakai event aktif ditolak atau membuat event jadi tanpa kategori diam-diam (pola serupa News CAT-04).
- **MOB-05/06/07 — 3 endpoint mobile yang pernah 500 ERR_INTERNAL** per `EVENT_ERROR.md` (23 Juni 2026): `/mobile/events/featured`, `/mobile/events/upcoming`, `/mobile/events/search`. **Di luar scope perbaikan sesi ini** (fokusnya settings/notifikasi/reminder/otorisasi) — cek ulang status sebelum demo, kalau masih error jangan andalkan carousel featured/upcoming di App untuk showcase.
- **MOB-12 — Endpoint presign media mobile beda dari draft awal spec**: rute aktual `/mobile/events/media/image/presign` (scoped ke event image), bukan `/mobile/media/presign` generic seperti sempat ditulis di draft `API_SPEC_EVENT_MOBILE.md`. Perbedaan ini murni dokumentasi, tim mobile harus pakai rute yang benar-benar terdaftar.
- **Konfirmasi ke tim frontend Backoffice**: tab Settings di `pages/events/index.vue` sebelumnya nonaktif karena endpoint backend belum ada (dicatat di `k-forum-backoffice/docs/BACKEND_ISSUES.md` baris 29-30). Field response backend baru ini didesain match 1:1 dengan tipe `EventSettings` yang sudah ada di frontend (`shared/types/Event.ts`), jadi seharusnya langsung jalan tanpa perubahan kode frontend — **tapi tetap perlu di-smoke-test live** di Backoffice sebelum demo, belum dicoba end-to-end lewat browser sungguhan di sesi ini (baru diverifikasi via automated handler test di backend).
