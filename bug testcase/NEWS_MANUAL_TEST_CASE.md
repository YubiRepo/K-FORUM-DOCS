# Manual Test Case — News (End-to-End: Web, API, App)

Dokumen test case manual untuk domain **News** — dipakai untuk persiapan presentasi/demo.
Meng-cover seluruh use case dari **buat sumber/artikel** sampai **notifikasi spesifik**
diterima user, lintas layer **Web (Backoffice)**, **API**, dan **App (Mobile)**.

- **Base URL API:** `https://k-forum-api.yubicom.co.id/api/v1`
- **Web:** `/api/v1/web/news/...` (Backoffice, butuh login admin)
- **Mobile/App:** `/api/v1/mobile/news/...` (dipakai App — App belum ada source code di repo ini, jadi acuan perilaku App = kontrak API mobile ini + spec di `API SPEC/Mobile/API_SPEC_NEWS_MOBILE.md`)
- **Referensi:** `Modules/News/NEWS_RULES.md`, `API SPEC/Web/API_SPEC_NEWS_BACKOFFICE.md`, `API SPEC/Mobile/API_SPEC_NEWS_MOBILE.md`, `flows/NOTIFICATION_DELIVERY_FLOW.md`

---

## Daftar Isi

1. [Role & Akun yang Dibutuhkan](#1-role--akun-yang-dibutuhkan)
2. [Peta Use Case](#2-peta-use-case)
3. [A — Source Management (Scraper)](#a--source-management-scraper)
4. [B — Category & Scope Management](#b--category--scope-management)
5. [C — Article Lifecycle: Jalur Manual (Editor/Admin)](#c--article-lifecycle-jalur-manual-editoradmin)
6. [D — Article Lifecycle: Jalur Member Pro (App, butuh approval)](#d--article-lifecycle-jalur-member-pro-app-butuh-approval)
7. [E — Article dari Scraper (Auto)](#e--article-dari-scraper-auto)
8. [F — Konsumsi Konten: Web/Backoffice](#f--konsumsi-konten-webbackoffice)
9. [G — Konsumsi Konten: App/Mobile](#g--konsumsi-konten-appmobile)
10. [H — Interaksi: Like, Bookmark, Comment (App)](#h--interaksi-like-bookmark-comment-app)
11. [I — Notifikasi: Broadcast saat Publish](#i--notifikasi-broadcast-saat-publish)
12. [J — Notifikasi: Approve/Reject ke Author](#j--notifikasi-approvereject-ke-author)
13. [K — Permission & Negative Case](#k--permission--negative-case)
14. [Skenario End-to-End (urutan untuk demo)](#skenario-end-to-end-urutan-untuk-demo)
15. [Catatan Risiko yang Perlu Diwaspadai Saat Demo](#catatan-risiko-yang-perlu-diwaspadai-saat-demo)

---

## 1. Role & Akun yang Dibutuhkan

| Role | Kebutuhan | Bisa apa (ringkas) |
|---|---|---|
| **Usergod** | 1 akun | Semua akses + register/edit/delete News Source |
| **Superadmin** | 1 akun | Semua operasi artikel, category, scope, source-category config; **tidak bisa** register source baru |
| **Editor** | 1 akun | Create/edit/publish/archive artikel manual, approve/reject (jika permission diberi); **tidak bisa** akses source/category/settings |
| **Member Pro** | 1 akun, punya benefit `post_news` | Submit artikel (butuh approval), edit draft sendiri, withdraw |
| **Member biasa** | 1 akun, tanpa benefit `post_news` | Baca, like, comment, bookmark saja — tidak bisa submit artikel |
| **Guest** | tanpa login | Baca artikel published + comment saja, tidak bisa interaksi |

> Siapkan minimal 2 akun Member biasa untuk test notifikasi broadcast (1 dengan preference "news" ON, 1 dengan OFF) — lihat bagian I.

---

## 2. Peta Use Case

| Use Case | Web | API (web) | API (mobile) | App |
|---|---|---|---|---|
| Kelola News Source (scraper) | ✅ | ✅ | — | — |
| Trigger scrape manual | ✅ | ✅ | — | — |
| Kelola Category & Scope | ✅ | ✅ | (read-only) | (read-only, dropdown filter) |
| Create/Edit/Publish/Archive artikel (admin) | ✅ | ✅ | — | — |
| Approve/Reject artikel (Member Pro submission) | ✅ | ✅ | — | — |
| Submit/Withdraw artikel (Member Pro) | — | — | ✅ (belum ada di spec mobile — cek Catatan Risiko) | ✅ |
| Feed & filter artikel | (list admin) | ✅ | ✅ | ✅ |
| Detail + translate on-demand | ✅ | ✅ | ✅ | ✅ |
| Like / Bookmark / Comment / Report comment | — | — | ✅ | ✅ |
| Notifikasi broadcast (artikel baru publish) | — | trigger otomatis | — | ✅ terima notif |
| Notifikasi approve/reject ke author | — | trigger otomatis | — | ✅ terima notif |
| Preferensi notifikasi modul "news" | — | — | ✅ | ✅ toggle di setting |

---

## A — Source Management (Scraper)

> Hanya **Usergod** yang boleh register/edit/delete source. **Superadmin** hanya boleh atur source-category & config operasional per source, **tidak** register source baru.

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| SRC-01 | Login sebagai Usergod | `POST /api/v1/web/news/sources` dengan `base_url`, `schedule` cron, `auto_publish=false`, `is_active=true` | `201` — source tersimpan; cron job scraping terdaftar (auto-run sesuai schedule) |
| SRC-02 | Login sebagai Superadmin, source sudah ada | Coba `POST /api/v1/web/news/sources` | `403` — Superadmin tidak boleh register source baru |
| SRC-03 | Source aktif ada | `PUT /api/v1/web/news/sources/{id}` ubah `schedule` cron | `200` — job lama di-reschedule, bukan dobel job |
| SRC-04 | Source aktif ada | `PUT .../sources/{id}` set `is_active=false` | `200` — cron job source **berhenti** jalan (verifikasi tidak ada artikel baru dari source ini setelah nonaktif) |
| SRC-05 | Source ada | `POST /api/v1/web/news/sources/{id}/scrape-now` | `200`/`202` — job scrape dienqueue segera (tidak nunggu jadwal cron) |
| SRC-06 | Source dengan `auto_publish=true` | Trigger scrape-now, tunggu proses selesai | Artikel baru hasil scrape berstatus **`published`** langsung |
| SRC-07 | Source dengan `auto_publish=false` | Trigger scrape-now | Artikel baru hasil scrape berstatus **`draft`**, menunggu review admin |
| SRC-08 | Source & item sudah pernah di-scrape (`original_url` sama) | Trigger scrape-now lagi | Item **tidak** dobel — dedup by `original_url` bekerja sebelum fetch detail |
| SRC-09 | Source-category mapping | `POST/PUT/DELETE .../sources/{source_id}/categories` | Mapping feed → kategori internal berhasil dibuat/diubah/dihapus |
| SRC-10 | Source ada, punya cron job aktif | `DELETE /api/v1/web/news/sources/{id}` | `200` — source terhapus. **Cek manual**: pastikan cron job tidak orphan (lihat Catatan Risiko) |

---

## B — Category & Scope Management

> Kategori & Scope: **flat list**, hanya Superadmin ke atas. Editor **tidak** bisa akses.

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| CAT-01 | Login Superadmin | `POST /api/v1/web/news/categories` dengan slug valid (`olahraga`) | `201` |
| CAT-02 | Kategori dengan slug `olahraga` sudah ada | `POST` kategori baru dengan slug sama | `409` `CodeNewsCategoryDuplicateSlug` |
| CAT-03 | — | `POST` kategori dengan slug mengandung kapital/spasi (`Olah Raga`) | `422` — slug harus match `^[a-z0-9]+(?:-[a-z0-9]+)*$` |
| CAT-04 | Kategori dipakai oleh artikel aktif | `DELETE /api/v1/web/news/categories/{id}` | **Cek**: apakah ditolak atau tetap terhapus (artikel jadi tanpa kategori)? Lihat Catatan Risiko |
| CAT-05 | Login Editor | `POST/PUT/DELETE` category atau scope | `403` — Editor tidak punya `manage_news_category` |
| SCP-01 | Login Superadmin | `POST /api/v1/web/news/scopes` (mis. `indonesia`, `korea`, `korea_indonesia`) | `201` |
| SCP-02 | Scope dipakai artikel & source | `DELETE /api/v1/web/news/scopes/{id}` | `200` — `articles.news_scope_id` & `news_sources.default_scope_id` yang refer jadi `NULL` (bukan artikel/source ikut terhapus) |
| SCP-03 | — | `GET /api/v1/mobile/news/scopes` | Hanya scope `is_active=true` yang muncul (mobile selalu `activeOnly`) |

---

## C — Article Lifecycle: Jalur Manual (Editor/Admin)

> `news_scope_id` dan `author_region_id` di sini **hanya filter**, bukan access-control — artikel tetap tampil global ke semua region meski scope-nya "korea" misalnya.

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| ART-01 | Login Editor, kategori tersedia | `POST /api/v1/web/news/articles` status `draft`, isi `translation` (title/content/summary) | `201` — status `draft`, belum tampil di mobile |
| ART-02 | Artikel draft ART-01 | `PUT /api/v1/web/news/articles/{id}` ubah judul/konten | `200` — draft masih bisa diedit bebas |
| ART-03 | Artikel draft | `POST /api/v1/web/news/articles/{id}/publish` | `200` — status `published`, `published_at`/`published_by_label` terisi; translation ke bahasa lain otomatis di-enqueue (sesuai system settings) |
| ART-04 | Artikel `published` | `PUT /api/v1/web/news/articles/{id}` ubah konten | `200` — **published boleh diedit langsung tanpa re-approval** (beda dengan jalur Member Pro) |
| ART-05 | Artikel `published` | `POST /api/v1/web/news/articles/{id}/archive` | `200` — status `archived`, hilang dari feed publik mobile, hilang dari search index |
| ART-06 | Artikel `published` atau `archived` | `DELETE /api/v1/web/news/articles/{id}` | `409` — "Cannot delete published article. Archive it instead." (hanya `draft`/`rejected` yang boleh hard-delete) |
| ART-07 | Artikel `draft` | `DELETE /api/v1/web/news/articles/{id}` | `200` — hard delete + thumbnail ikut dibersihkan |
| ART-08 | Artikel dibuat langsung dengan `status: published` di body `POST /articles` | — | `201` langsung published (tanpa lewat draft dulu) — cek `published_at` & notifikasi broadcast tetap terpicu |
| ART-09 | `POST /articles/{id}/translate` pada artikel published | — | `200`/`202` — trigger translate ke bahasa target sesuai system settings |
| ART-10 | `PUT /articles/{id}/translations/{lang}` isi manual | — | `200` — override hasil AI, `translated_by=null`, `is_original=false` |

---

## D — Article Lifecycle: Jalur Member Pro (App, butuh approval)

> Endpoint submit di App: `POST /api/v1/mobile/news/articles` — **gated `RequireBenefit("post_news")`**.

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| MPR-01 | Login Member biasa (tanpa benefit `post_news`) di App | Coba submit artikel baru | `403` — tidak ada benefit, tombol "Tulis Berita" seharusnya tidak muncul/di-disable di App |
| MPR-02 | Login Member Pro (benefit `post_news` aktif) | Submit artikel baru dari App | Artikel dibuat status `pending_approval` |
| MPR-03 | Artikel `pending_approval` milik Member Pro | Member Pro edit draft-nya sebelum di-approve | Berhasil (masih miliknya, belum publish) |
| MPR-04 | Artikel `pending_approval` | Member Pro lain (bukan pemilik) coba **withdraw** artikel | `403` — withdraw hanya untuk pemilik (ownership-checked) |
| MPR-05 | Artikel `pending_approval` milik sendiri | `DELETE /api/v1/mobile/news/articles/{id}/withdraw` | `200` — status balik ke `draft` |
| MPR-06 | Artikel `pending_approval` | Login Editor/Admin di Web, `POST /articles/{id}/approve` | `200` — status `published`; **author menerima notifikasi** (lihat J-01) |
| MPR-07 | Artikel `pending_approval` | Login Editor/Admin, `POST /articles/{id}/reject` dengan body `{ "reason": "..." }` | `200` — status `rejected`; **author menerima notifikasi berisi alasan** (lihat J-02) |
| MPR-08 | Artikel `rejected` milik Member Pro | Member Pro edit lalu submit ulang | Berhasil kembali jadi `pending_approval` (`CanEdit()` mengizinkan status `rejected`) |
| MPR-09 | Login akun tanpa permission `approve_news`/`approve_article` | Coba approve/reject | `403` — cek juga inkonsistensi permission (lihat Catatan Risiko K-04) |

---

## E — Article dari Scraper (Auto)

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| SCR-01 | Source `auto_publish=true`, AI Cleanup ON | Scrape-now sukses | Artikel `published` otomatis, konten sudah melalui goquery cleanup (layer 1) + AI cleanup (layer 2) |
| SCR-02 | Source `auto_translate=true` | Artikel baru dari scrape | Translation ke bahasa lain otomatis tergenerate (lihat matrix translation di `NEWS_RULES.md`) |
| SCR-03 | Source `auto_publish=false` | Scrape-now sukses | Artikel masuk sebagai `draft`, muncul di list admin untuk direview manual sebelum publish |
| SCR-04 | Artikel hasil scrape | Cek field `is_manual` & `original_url` | `is_manual=false`, `original_url` terisi URL sumber asli |

---

## F — Konsumsi Konten: Web/Backoffice

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| WEB-01 | Ada campuran artikel draft/published/pending/archived | `GET /api/v1/web/news/articles` tanpa filter | Semua status muncul (admin lihat semua, bukan cuma published) |
| WEB-02 | — | `GET /api/v1/web/news/articles?status=pending_approval` | Hanya artikel menunggu approval — ini yang jadi "antrian review" Editor |
| WEB-03 | — | Buka halaman `news.vue` list, filter by kategori/scope/status di UI | Data konsisten dengan hasil API (cek kontrak `pagination{limit,offset,total}` vs `meta{...}` — lihat Catatan Risiko K-05) |
| WEB-04 | — | Buka form create/edit artikel di Backoffice | Validasi field wajib (`category_id`, `original_language`, `author_label`, `translation.title/content`) jalan di UI sebelum submit |

---

## G — Konsumsi Konten: App/Mobile

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| MOB-01 | Ada artikel published berbagai kategori/scope | `GET /api/v1/mobile/news/articles` (guest, tanpa token) | `200` — hanya artikel `published` yang muncul |
| MOB-02 | — | `GET .../articles?category=olahraga` | Terfilter sesuai slug kategori |
| MOB-03 | — | `GET .../articles?scope=korea` | Terfilter sesuai slug scope |
| MOB-04 | — | `GET .../articles?featured=true` | Hanya artikel `is_featured=true` |
| MOB-05 | — | `GET .../articles?q=keyword` | Search di judul + konten |
| MOB-06 | Login user, sudah like/bookmark artikel X | `GET .../articles/{X}` | `is_liked`/`is_bookmarked` = `true` (akurat sesuai user login) |
| MOB-07 | Sebagai guest (tanpa login) | `GET .../articles/{X}` | Tetap bisa baca detail, `is_liked`/`is_bookmarked` default `false`, view counter tetap naik (`view_count` total) |
| MOB-08 | User login, buka artikel yang sama 2x | Cek `unique_view_count` | Naik hanya **1x** per user (unique), sedangkan `view_count` naik tiap kali buka |
| MOB-09 | Artikel translation bahasa target belum ada, `on_demand_enabled=true` | `GET .../articles/{id}/translate?language=en` | `202` + fallback konten original, translate diproses async |
| MOB-10 | Ulangi request setelah translate selesai | `GET .../articles/{id}/translate?language=en` | `200` — konten sudah dalam bahasa target |
| MOB-11 | — | `GET /api/v1/mobile/news/categories`, `GET /scopes` | Hanya item `is_active=true` |
| MOB-12 | App: buka tab/menu News | Scroll feed, ganti filter kategori/scope, buka detail, kembali ke list | Tidak ada crash; state filter & pagination tetap konsisten |

---

## H — Interaksi: Like, Bookmark, Comment (App)

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| INT-01 | Login user, artikel belum di-like | `POST .../articles/{id}/like` | `200` — `is_liked=true`, `like_count` naik 1 |
| INT-02 | Sudah like (INT-01) | `POST` like lagi (double-tap) | Idempotent — `like_count` **tidak** naik dobel |
| INT-03 | Sudah like | `DELETE .../articles/{id}/like` | `200` — `is_liked=false`, `like_count` turun 1 |
| INT-04 | Login user | `POST/DELETE .../articles/{id}/bookmark`, lalu `GET .../bookmarks` | Bookmark list menampilkan artikel yang di-bookmark, hilang setelah unbookmark |
| INT-05 | Login user | `POST .../articles/{id}/comments` body `{content, parent_id: null}` | `201` — comment level 1 tersimpan |
| INT-06 | Comment level 1 ada | Reply ke comment tsb dengan `parent_id` = id comment level 1 | `201` — jadi level 2 |
| INT-07 | Comment level 2 (reply) ada | Coba reply lagi ke comment level 2 itu | Backend **flatten** — tetap tersimpan sebagai level 2 dengan `parent_id` = level 1 asal (tidak ada level 3) |
| INT-08 | Comment milik user lain | `DELETE /mobile/news/comments/{id}` | `403` — hanya pemilik comment yang boleh hapus (soft delete); **tidak ada** jalur moderator-delete di App meski usecase back-end mendukungnya |
| INT-09 | Comment sendiri ada | `DELETE /mobile/news/comments/{id}` | `200` — soft delete, comment hilang dari list tapi tidak bisa post ulang dengan isi sama sebagai "edit" (tidak ada fitur edit comment) |
| INT-10 | Login user | `POST /mobile/news/comments/{id}/report` | `200`/`201` — laporan tersimpan untuk moderasi |
| INT-11 | Guest (tanpa login) | Coba like/bookmark/comment | `401` — guest hanya boleh baca artikel & comment |

---

## I — Notifikasi: Broadcast saat Publish

> Flow: `PublishArticleUseCase` → outbox `news.article.published` → fanout ke **semua user** (in_app + push), difilter oleh preferensi modul `news` per user. DND hanya menekan channel **push**, in-app tetap masuk.

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| NOTIF-01 | 2 akun member: User A (preferensi "news" **ON**, default), User B | Admin publish artikel baru (draft→publish) | User A menerima notifikasi in-app **dan** push berisi info artikel baru |
| NOTIF-02 | User B set `PUT /mobile/notifications/preferences/modules/news` → `enabled=false` | Admin publish artikel baru lagi | User B **tidak** menerima notifikasi sama sekali (in-app maupun push) — tervalidasi opt-out bekerja |
| NOTIF-03 | User A aktifkan Do Not Disturb (DND) di jam saat ini | Admin publish artikel baru | User A tetap terima notifikasi **in-app**, tapi **push tidak terkirim** selama jam DND |
| NOTIF-04 | Artikel dibuat langsung `status: published` via `POST /articles` (ART-08) | — | Notifikasi broadcast tetap terpicu sama seperti lewat tombol publish |
| NOTIF-05 | User A tap notifikasi di App | — | Deep link membawa user langsung ke detail artikel yang dipublish (cek `Domain`/deeplink docs jika ada) |
| NOTIF-06 | Artikel di-archive (bukan publish) | — | **Tidak ada** notifikasi terkirim untuk archive/reject-dari-search (hanya publish/approve/reject yang notify) |

---

## J — Notifikasi: Approve/Reject ke Author

> Event `news.article_approved` / `news.article_rejected` masuk **BypassEventKeys** — selalu terkirim ke author **meski preferensi modul "news" dimatikan** (beda dengan broadcast publish di atas).

| ID | Precondition | Steps | Expected Result |
|---|---|---|---|
| J-01 | Member Pro submit artikel (MPR-02), preferensi "news" **dimatikan** oleh author | Admin approve artikel tsb | Author **tetap** menerima notifikasi approved (unicast) — bypass toggle |
| J-02 | Member Pro submit artikel | Admin reject dengan `reason` | Author menerima notifikasi berisi alasan reject |
| J-03 | — | Bandingkan notifikasi approve/reject: apakah **unicast** hanya ke author, tidak broadcast ke user lain | User lain (bukan author) **tidak** menerima notifikasi approve/reject ini |

---

## K — Permission & Negative Case

| ID | Steps | Expected Result |
|---|---|---|
| K-01 | Login Editor, coba akses endpoint source (`/sources`, `/sources/*/categories`) | `403` — Editor tanpa akses source sama sekali |
| K-02 | Login Editor, coba `POST /categories` atau `/scopes` | `403` — hanya Superadmin ke atas |
| K-03 | Login user dengan permission `create_news` tapi tanpa `publish_news` | Bisa create draft, tapi `POST /articles/{id}/publish` → `403` |
| K-04 | Login user dengan permission `approve_news` **tapi tidak** `approve_article` (atau sebaliknya) | Cek konsisten: route-level cek `approve_news`, tapi usecase re-check `approve_article` via `CanApproveContent()` — pastikan tidak ada kombinasi permission yang bikin user "punya akses di router tapi ditolak di usecase" atau sebaliknya |
| K-05 | Coba `DELETE` kategori yang masih dipakai artikel aktif | Pastikan tidak menyebabkan artikel jadi tanpa kategori diam-diam tanpa peringatan |
| K-06 | Coba inject field `role`/`is_admin`/`published_by_user_id` di body `POST /articles` sebagai Member Pro | Field tidak berpengaruh — tidak bisa bikin artikel langsung `published` atau naikkan privilege |
| K-07 | Kirim XSS payload (`<script>alert(1)</script>`) di `translation.content`/comment | Pastikan tampil ter-escape di Web & App, tidak tereksekusi sebagai script |
| K-08 | `DELETE` source yang punya cron job aktif (SRC-10) | Cek log/scheduler — pastikan job tidak tetap jalan mencoba scrape source yang sudah dihapus (orphan job) |

---

## Skenario End-to-End (urutan untuk demo)

Urutan ini didesain untuk demo presentasi — tunjukkan flow lengkap **create → publish → muncul di app → notifikasi**, lalu **submission-approval**, lalu **scraper**.

### Flow 1 — Artikel Manual: Admin Create → Publish → Tampil di App → Notifikasi
1. **[Web]** Login Editor → buat artikel baru status `draft` (kategori + scope + konten lengkap).
2. **[Web]** Buka list artikel → pastikan artikel muncul dengan status `draft`.
3. **[App]** Cek feed News (`GET /mobile/news/articles`) → artikel **belum** muncul (masih draft).
4. **[Web]** Klik **Publish** pada artikel tsb.
5. **[API]** Verifikasi `status=published`, `published_at` terisi.
6. **[App]** Refresh feed News → artikel **sekarang muncul** di paling atas/sesuai sort.
7. **[App]** Cek notifikasi masuk (in-app + push) di device User A (preferensi ON).
8. **[App]** Tap notifikasi → langsung ke detail artikel.
9. **[App]** Cek device User B (preferensi OFF) → **tidak** ada notifikasi masuk.

### Flow 2 — Submission Member Pro: Submit → Approve → Publish → Notifikasi Author
1. **[App]** Login Member Pro → buka "Tulis Berita" → submit artikel baru.
2. **[Web]** Login Editor → buka antrian `pending_approval` → artikel Member Pro muncul di sana.
3. **[Web]** Klik **Approve**.
4. **[App]** Login sebagai Member Pro tsb → cek notifikasi "artikel disetujui" masuk.
5. **[App]** Cek feed News publik → artikel sudah tampil untuk semua user.

### Flow 3 — Submission Ditolak
1. **[App]** Member Pro submit artikel lain.
2. **[Web]** Editor **Reject** dengan alasan tertentu.
3. **[App]** Author menerima notifikasi reject berisi alasan.
4. **[App]** Author edit artikel yang rejected lalu submit ulang → balik ke `pending_approval`.

### Flow 4 — Withdraw
1. **[App]** Member Pro submit artikel.
2. **[App]** Sebelum di-approve, Member Pro tarik lagi via **Withdraw**.
3. **[Web]** Cek artikel balik ke status `draft`, hilang dari antrian approval Editor.

### Flow 5 — Scraper Otomatis
1. **[Web]** Usergod pastikan ada News Source aktif dengan `auto_publish=true`.
2. **[Web]** Klik **Scrape Now** pada source tsb.
3. **[Web]** Tunggu beberapa saat → cek list artikel, muncul artikel baru berstatus `published` dengan `is_manual=false`.
4. **[App]** Cek artikel hasil scrape muncul di feed + trigger notifikasi broadcast sama seperti Flow 1.
5. Ulangi **Scrape Now** → pastikan artikel yang sama **tidak dobel** (dedup by `original_url`).

### Flow 6 — Delete Rules
1. **[Web]** Coba `DELETE` artikel yang sudah `published` → tampilkan pesan error `409` (harus di-archive dulu).
2. **[Web]** Archive artikel tsb → lalu coba `DELETE` lagi → masih `409` (archived pun tidak bisa langsung delete, hanya draft/rejected).
3. **[Web]** Buat artikel draft baru → langsung `DELETE` → berhasil `200`.

---

## Catatan Risiko yang Perlu Diwaspadai Saat Demo

- **K-04 — Inkonsistensi permission approve/reject**: route pakai `approve_news`, tapi usecase cek `approve_article`. Kalau demo approve gagal padahal permission "kelihatan" benar di UI, ini kemungkinan penyebabnya.
- **SRC-10 — Orphan cron job**: `DeleteSourceUseCase` tidak membersihkan job terjadwal milik source yang dihapus. Kalau demo delete source lalu source itu "masih jalan scraping", ini penyebabnya — bukan bug baru.
- **CAT-04 — Delete category tanpa cek referensi**: hapus kategori yang masih dipakai artikel tidak divalidasi; artikel terkait bisa jadi tanpa kategori tanpa peringatan.
- **WEB-03 — Kontrak pagination**: spec API pakai `pagination{limit,offset,total}`, tapi store frontend (`newsStore.ts`) kadang ekspektasi `meta{current_page,per_page,total,total_pages}` — kalau list di Backoffice terlihat kosong padahal API 200 dengan data, cek shape ini dulu.
- **MPR-09 — Endpoint submit App**: `POST /mobile/news/articles` (submit Member Pro) ada di usecase back-end tapi belum eksplisit terdokumentasi di `API_SPEC_NEWS_MOBILE.md` — konfirmasi dulu ke tim mobile bahwa endpoint ini sudah live sebelum dipakai di demo Flow 2.
- Untuk histori bug yang **sudah** pernah ditemukan & sebagian sudah fix (like/comment 500, `is_liked` tidak sinkron, dll), lihat `bug testcase/API_ERROR_REPORT.md` dan `bug testcase/NEWS_LIKE_COMMENT_BUG.md` — cek ulang status-nya sebelum demo supaya tidak kaget kalau muncul lagi.
