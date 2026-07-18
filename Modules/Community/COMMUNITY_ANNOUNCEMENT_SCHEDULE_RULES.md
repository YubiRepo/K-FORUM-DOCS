# Community Announcement & Schedule — Rules & Use Cases (v1.0)

Dokumen ini menjelaskan aturan bisnis dua fitur tambahan di modul **Community**: **Papan Pengumuman** dan **Schedule Komunitas**. Fokus pada **siapa bisa apa**, **kapan**, dan **kenapa** — bukan detail teknis. Untuk detail teknis lihat `COMMUNITY_ANNOUNCEMENT_SCHEDULE_DB_SCHEMA.md` serta API spec mobile & backoffice.

> **Posisi fitur:** Dua fitur ini berada **di dalam** modul Community (scoped per komunitas), dibuat oleh leader/moderator, dilihat oleh anggota komunitas.
>
> **Bukan pengganti modul lain:**
> - **Papan Pengumuman ≠ modul Announcement platform.** Announcement platform = broadcast superadmin/admin regional ke semua user (global/regional). Papan Pengumuman = info resmi leader untuk anggota satu komunitas.
> - **Schedule Komunitas ≠ modul Schedule backoffice.** Schedule backoffice = kalender internal admin. Schedule Komunitas = agenda komunitas yang dilihat anggota di mobile.
> - **Schedule Komunitas ≠ modul Event.** Event = acara publik dengan registrasi, dilihat semua member. Schedule Komunitas = agenda internal komunitas (latihan, rapat, kegiatan), hanya untuk anggota.

---

## Daftar Isi

1. [Overview Konsep](#overview-konsep)
2. [Permission Baru](#permission-baru)
3. [Fitur A — Papan Pengumuman](#fitur-a--papan-pengumuman)
4. [Fitur B — Schedule Komunitas](#fitur-b--schedule-komunitas)
5. [Recurrence & Occurrence](#recurrence--occurrence)
6. [RSVP (Indikator Minat)](#rsvp-indikator-minat)
7. [Siapa Bisa Apa](#siapa-bisa-apa)
8. [Alur Utama](#alur-utama)
9. [Status & Transition](#status--transition)
10. [Hubungan dengan Modul Lain](#hubungan-dengan-modul-lain)
11. [Edge Cases & Rules](#edge-cases--rules)
12. [Phasing](#phasing)
13. [Keputusan yang Masih Terbuka](#keputusan-yang-masih-terbuka)

---

## Overview Konsep

Dua fitur, satu prinsip yang sama: **konten dibuat oleh pengelola komunitas, dikonsumsi oleh anggota.**

| Fitur | Apa | Dibuat oleh | Dilihat oleh |
|---|---|---|---|
| **Papan Pengumuman** | Info resmi/penting dari komunitas (read-only) | Leader/moderator (permission) | Anggota komunitas |
| **Schedule Komunitas** | Agenda/kalender komunitas (latihan, rapat, kegiatan) | Leader/moderator (permission) | Anggota komunitas |

Prinsip desain yang dipegang (konsisten dengan modul lain):

- **Permission-driven, bukan role-hardcoded** — hak bikin pengumuman/agenda diputuskan lewat permission key, bukan dicek "apakah dia leader". Leader bisa mendelegasikan ke moderator per komunitas.
- **Scope = komunitas** — semua data terikat `community_id`. Komunitas private → hanya anggota yang lihat.
- **Phase 1 solid, hook Phase 2 disiapkan** — pola sama dengan modul lain (`source`, `recurrence`, dll disiapkan walau belum semua dipakai).

---

## Permission Baru

Dua permission key baru didaftarkan di modul **Role-Permission** (master `permissions`) dan masuk ke **template community role**:

| Key | Display | Default di template |
|---|---|---|
| `manage_community_announcement` | Kelola Pengumuman Komunitas | Leader ✅ · Moderator ❌ (bisa di-grant) · Member ❌ |
| `manage_community_schedule` | Kelola Agenda Komunitas | Leader ✅ · Moderator ❌ (bisa di-grant) · Member ❌ |

> **Pola identik dengan `moderate_posts`:** leader otomatis dapat, moderator ditentukan leader per komunitas via Role-Permission bulk-assign. Modul Community **tidak** menyimpan logika permission sendiri — cuma memanggil permission check.
>
> **Follow-up wajib:** registrasi kedua key ini di `ROLE_PERMISSION_*` (master + template) dan pertimbangkan gating di `PLAN_SUBSCRIPTION` bila mau dijadikan benefit (mengikuti pola key Q&A: `answer_qna`, `validate_qna_answer`, `manage_qna`).

---

## Fitur A — Papan Pengumuman

### Konsep

Pengumuman adalah **info resmi read-only** dari pengelola komunitas untuk anggotanya. Berbeda dari `community_posts` (feed sosial): pengumuman tidak untuk di-like/comment (Phase 1), punya **prioritas** dan **kedaluwarsa**, serta bisa **di-pin** di atas.

### Entitas: CommunityAnnouncement

| Atribut | Keterangan |
|---|---|
| `community_id`, `author_id` | Konteks komunitas & penulis |
| `title` | Judul pengumuman (wajib) |
| `body` | Isi — **plain text** (konsisten dengan post; client boleh auto-render URL/mention) |
| `media` | Lampiran gambar opsional, maks 5 (lebih kecil dari post karena bukan konten sosial) |
| `priority` | `normal` \| `important` — `important` memicu push notification |
| `is_pinned` | Pin di atas daftar pengumuman |
| `expires_at` | Nullable — setelah lewat, pengumuman auto-hide dari anggota (tetap tersimpan) |
| `status` | `draft` \| `published` \| `archived` |

### Aturan

- **Read-only untuk anggota (Phase 1).** Tidak ada like/comment/save. Interaksi ditunda Phase 2 bila diperlukan.
- **Priority menentukan notifikasi:** `important` → push + in-app; `normal` → in-app/badge saja (tanpa push).
- **`expires_at` opsional:** kalau diisi, pengumuman hilang dari view anggota setelah lewat, tapi tetap ada di arsip backoffice/leader.
- **Pin:** boleh lebih dari satu pengumuman ter-pin; urutan tampil = pinned dulu (priority `important` di atas `normal`), lalu terbaru.
- **Visibility ikut komunitas:** komunitas private → hanya anggota aktif yang bisa baca pengumuman.
- **Draft:** hanya terlihat oleh pembuat & pengelola; tidak dikirim notif sampai `published`.

---

## Fitur B — Schedule Komunitas

### Konsep

Kalender agenda milik komunitas: latihan rutin, rapat, kegiatan, turnamen. Anggota bisa **melihat** agenda dan **menandai minat** (RSVP). Agenda bisa **berulang** (recurrence, aktif Phase 1).

### Entitas: CommunityScheduleEntry

| Atribut | Keterangan |
|---|---|
| `community_id`, `created_by` | Konteks & pembuat |
| `title`, `description` | Judul & detail (description opsional) |
| `start_at` | Waktu mulai (untuk recurring = waktu occurrence pertama) |
| `end_at` | Waktu selesai (opsional; jika ada harus ≥ `start_at`) |
| `all_day` | `true` = sepanjang hari (abaikan komponen jam) |
| `location` | Teks bebas (opsional) — mis. "GOR Senayan Lapangan 3" |
| `recurrence` | Nullable — aturan berulang RRULE-style (mis. `FREQ=WEEKLY;BYDAY=SA`). **Aktif Phase 1** |
| `timezone` | **Ditambahkan 2026-07-17.** IANA identifier (mis. `Asia/Jakarta`) — **wajib** diisi eksplisit oleh creator saat create (client boleh prefill dari device, tidak pernah diinfer diam-diam di backend), **opsional** saat edit (partial update, fallback ke nilai lama kalau di-omit). Menentukan zona waktu yang dipakai untuk merekonstruksi `start_at`/occurrence dengan benar — lihat §Recurrence & Occurrence. |
| `status` | `active` \| `cancelled` (batal seluruh series/agenda) |

### Entitas: CommunityScheduleRsvp

Menandai minat anggota terhadap **satu occurrence tertentu** (bukan seluruh series).

| Atribut | Keterangan |
|---|---|
| `entry_id`, `occurrence_date`, `user_id` | Kunci unik gabungan |
| `response` | `going` \| `maybe` \| `not_going` |

### Entitas: CommunityScheduleException

Menandai satu occurrence dari agenda berulang yang **dibatalkan** (mis. "Sabtu depan libur").

| Atribut | Keterangan |
|---|---|
| `entry_id`, `occurrence_date` | Occurrence yang di-override |
| `type` | `cancelled` (Phase 1). `modified` disiapkan untuk Phase 2 (ubah detail per-occurrence) |

---

## Recurrence & Occurrence

Ini bagian paling rawan, jadi aturannya dibikin eksplisit.

**Konsep:**
- Satu `community_schedule_entries` bisa mewakili **banyak tanggal** kalau `recurrence` diisi (mis. tiap Sabtu).
- Sistem **tidak** menyimpan satu baris per tanggal. Occurrence dihitung **on-the-fly** dari `start_at` + `recurrence`, dibatasi oleh **rentang query** (mis. kalender bulan berjalan).
- Agenda one-off = `recurrence` NULL → cuma punya satu occurrence, yaitu tanggal `start_at`.

**Aturan:**
- **Occurrence selalu dibatasi window.** Backend tidak pernah generate occurrence tak-terbatas; selalu dalam rentang tanggal yang diminta (mis. `?from=2026-07-01&to=2026-07-31`).
- **RSVP & exception nempel ke `occurrence_date`**, bukan ke entry secara keseluruhan. Jadi "gw ikut Sabtu ini doang" tidak menyeret RSVP ke semua Sabtu.
- **Batalin satu tanggal** = insert `community_schedule_exceptions (entry_id, occurrence_date, type='cancelled')`. Occurrence itu tampil dengan tanda batal; RSVP dinonaktifkan.
- **Batalin seluruh agenda** = set `entries.status='cancelled'`. Semua occurrence tampil batal.
- **Ubah detail satu occurrence** (mis. ganti jam Sabtu ini doang) → **ditunda Phase 2** (`exception.type='modified'` sudah disiapkan sebagai hook).
- **Rekonstruksi occurrence memakai `entry.timezone` (ditambahkan 2026-07-17), bukan zona server atau zona apa pun yang menempel di kolom `start_at` (TIMESTAMPTZ) pasca-baca dari database.** `start_at` disimpan sebagai instant absolut (Postgres menormalisasi ke UTC internal), sehingga membaca komponen jam/tanggalnya langsung tanpa mengonversi dulu ke `entry.timezone` bisa salah untuk agenda yang jam lokalnya menyebrang tengah malam UTC (mis. jam 03:00 WIB = 20:00 UTC hari sebelumnya) — baik jam occurrence-nya maupun *hari apa* occurrence itu jatuh (relevan untuk `recurrence` mingguan/bulanan tanpa `BYDAY` eksplisit). Bug ini ditemukan & diperbaiki 2026-07-17 bersamaan dengan penambahan kolom `timezone`.

---

## RSVP (Indikator Minat)

Poin penting: **RSVP di sini indikator minat, bukan absensi otoritatif.**

- Anggota menandai `going` / `maybe` / `not_going` untuk membantu pengelola memperkirakan keramaian.
- **Ini bukan catatan kehadiran (attendance) yang bisa dipercaya sebagai kebenaran.** Check-in/absensi beneran ditunda ke fase berikut kalau memang dibutuhkan.
- **Satu RSVP per (entry, occurrence_date, user)** — bisa diubah kapan saja selama occurrence belum lewat.
- **RSVP hanya untuk occurrence aktif & belum lewat.** Occurrence yang batal atau sudah lewat tidak menerima RSVP baru.
- Pengelola bisa melihat **ringkasan hitungan** (`going: 12, maybe: 4, not_going: 2`) dan daftar nama per occurrence.

---

## Siapa Bisa Apa

### Anggota Komunitas (member aktif)

| Aksi | Bisa? | Catatan |
|---|---|---|
| Lihat pengumuman komunitas | ✅ Ya | Yang `published` & belum expired |
| Like/comment pengumuman | ❌ Tidak | Read-only Phase 1 |
| Lihat agenda/kalender komunitas | ✅ Ya | Termasuk occurrence dari recurring |
| RSVP occurrence (going/maybe/not_going) | ✅ Ya | Occurrence aktif & belum lewat |
| Bikin pengumuman / agenda | ❌ Tidak | Perlu permission |

### Leader / Moderator (sesuai permission per komunitas)

| Aksi | Perlu permission |
|---|---|
| Buat/edit/hapus pengumuman | `manage_community_announcement` |
| Pin/unpin, set priority & expiry | `manage_community_announcement` |
| Buat/edit/hapus agenda | `manage_community_schedule` |
| Set recurrence, batalin occurrence, batalin agenda | `manage_community_schedule` |
| Lihat ringkasan & daftar RSVP | `manage_community_schedule` |

> Leader mendapat kedua permission secara default dari template. Moderator hanya jika di-grant leader (via Role-Permission bulk-assign). Member tidak pernah punya.

### Non-anggota / Guest

| Konteks | Akses |
|---|---|
| Komunitas public | Bisa lihat pengumuman & agenda **hanya jika** boleh lihat konten komunitas (ikut aturan visibility Community); RSVP tetap butuh keanggotaan |
| Komunitas private | Tidak ada akses sampai jadi anggota |

### Superadmin (Backoffice)

Bisa melihat & memoderasi pengumuman/agenda semua komunitas (override), sejalan dengan wewenang moderasi konten komunitas yang sudah ada.

---

## Alur Utama

### Alur 1: Leader Publish Pengumuman

```
1. Leader/mod → POST /communities/{id}/announcements
   { title, body, priority, is_pinned?, expires_at?, media?[], status: 'published' }
2. Backend:
   - Cek permission manage_community_announcement pada scope community ini ✅
   - Insert community_announcements (status=published)
   - Jika priority=important → emit event community_announcement_published → Notification (push)
   - Jika priority=normal → in-app/badge saja
3. Response: announcement object
```

### Alur 2: Member Baca Pengumuman

```
1. Member → GET /communities/{id}/announcements
2. Backend:
   - Cek keanggotaan / visibility ✅
   - Return published & (expires_at IS NULL OR expires_at > now), urut pinned→priority→terbaru
3. Response: list
```

### Alur 3: Leader Buat Agenda Berulang

```
1. Leader/mod → POST /communities/{id}/schedule
   { title, start_at, end_at?, all_day?, location?, recurrence: 'FREQ=WEEKLY;BYDAY=SA' }
2. Backend:
   - Cek permission manage_community_schedule ✅
   - Insert community_schedule_entries (status=active, recurrence set)
3. Response: entry object (satu baris mewakili semua Sabtu)
```

### Alur 4: Member Lihat Kalender & RSVP

```
1. Member → GET /communities/{id}/schedule?from=2026-07-01&to=2026-07-31
2. Backend:
   - Ambil entries komunitas
   - Expand occurrence dari recurrence dalam window
   - Kurangi occurrence yang punya exception type=cancelled
   - Sisipkan status RSVP milik user per occurrence
3. Member → PUT /communities/{id}/schedule/{entry_id}/rsvp
   { occurrence_date: '2026-07-12', response: 'going' }
4. Backend:
   - Cek occurrence valid, aktif, belum lewat ✅
   - Upsert community_schedule_rsvps (entry_id, occurrence_date, user_id)
```

### Alur 5: Leader Batalin Satu Occurrence

```
1. Leader/mod → POST /communities/{id}/schedule/{entry_id}/cancel-occurrence
   { occurrence_date: '2026-07-19' }
2. Backend:
   - Insert community_schedule_exceptions (type=cancelled)
   - (opsional) notify member yang sudah RSVP going/maybe
3. Occurrence 19 Juli tampil dengan tanda batal, RSVP dinonaktifkan
```

### Alur 6: Reminder Agenda

```
- Scheduler cek occurrence yang akan datang dalam X waktu sebelum start_at
- Untuk tiap occurrence aktif → emit event community_schedule_reminder → Notification
- Target: anggota yang RSVP going/maybe (Phase 1); target lebih luas bisa dikonfigurasi nanti
```

---

## Status & Transition

### Announcement

```
draft ──publish──> published ──archive──> archived
published ──(expires_at lewat)──> auto-hidden dari member (status tetap published, tersaring by query)
published ──archive──> archived
```

### Schedule Entry

```
active ──cancel──> cancelled        (batalin seluruh agenda/series)
```

### Occurrence (turunan, bukan baris tabel)

```
(default) active
active ──exception cancelled──> cancelled (satu tanggal saja)
active ──(start_at lewat)──> past (read-only, RSVP terkunci)
```

### RSVP

```
(belum ada) ──set──> going | maybe | not_going ──ubah──> going | maybe | not_going
(occurrence lewat/batal) → RSVP terkunci
```

---

## Hubungan dengan Modul Lain

| Kebutuhan | Ditangani oleh | Bagaimana fitur ini memakainya |
|---|---|---|
| Hak bikin pengumuman/agenda | Role-Permission | Permission check `manage_community_announcement` / `manage_community_schedule`, scope=community |
| Keanggotaan & visibility | Community (core) | Reuse `community_members` + `visibility` untuk gating baca |
| Push & reminder | Notification | Event `community_announcement_published`, `community_schedule_reminder` |
| Region (opsional) | Region | Tidak dipakai langsung — ikut region komunitas induk kalau perlu label |

> Fitur ini **tidak** menyentuh modul Announcement platform maupun Schedule backoffice. Nama sengaja diprefiks `community_*` supaya jelas terpisah.

---

## Edge Cases & Rules

### Rule 1: Permission dicek per komunitas
Moderator yang punya `manage_community_schedule` di komunitas A **tidak otomatis** punya di komunitas B. Scope isolation, konsisten dengan `COMMUNITY_ROLE_SCOPE_ANALYSIS.md`.

### Rule 2: Pengumuman expired tidak dihapus
`expires_at` lewat → tersaring dari view anggota, tetap tersimpan & terlihat di arsip pengelola.

### Rule 3: RSVP nempel ke occurrence, bukan series
Batalin/ubah RSVP satu tanggal tidak memengaruhi tanggal lain dalam series yang sama.

### Rule 4: Cancel occurrence vs cancel agenda
`exception.type=cancelled` membatalkan **satu tanggal**; `entries.status=cancelled` membatalkan **seluruh agenda**. Keduanya reversible (hapus exception / set active lagi) selama belum lewat.

### Rule 5: Occurrence dibatasi window
Query kalender **wajib** menyertakan rentang tanggal; recurring tanpa `UNTIL`/`COUNT` tetap aman karena expansion dibatasi window backend.

### Rule 6: RSVP = indikator, bukan kebenaran
Angka RSVP tidak boleh dipakai sebagai catatan kehadiran resmi. Attendance/check-in adalah fitur terpisah di fase berikut bila dibutuhkan.

### Rule 7: Hapus komunitas = cleanup penuh
Penghapusan komunitas harus ikut membersihkan `community_announcements`, `community_schedule_entries`, `community_schedule_rsvps`, dan `community_schedule_exceptions` dalam transaksi cleanup Community yang sudah ada.

---

## Phasing

| Fitur | Phase 1 | Phase 2+ |
|---|---|---|
| Pengumuman | CRUD, priority, pin, expiry, push (important), read-only | Like/comment, penjadwalan publish, kategori |
| Schedule | CRUD, recurrence (RRULE), cancel occurrence, cancel agenda | Ubah detail per-occurrence (`exception.modified`), assign PIC |
| RSVP | going/maybe/not_going per occurrence, ringkasan hitungan | Check-in/attendance otoritatif, kuota/cap |
| Reminder | Notif sebelum start_at ke RSVP going/maybe | Reminder configurable (lead time, target), reminder pengumuman |
| Integrasi | Event `published` & `reminder` ke Notification | Tarik agenda ke Event bila relevan (hook) |

---

## Keputusan yang Masih Terbuka

Asumsi default di dokumen ini, **mudah diubah**:

| # | Topik | Asumsi Sementara |
|---|---|---|
| 1 | Lead time reminder | Default satu nilai platform (mis. 1 jam sebelum). Configurable = Phase 2 |
| 2 | Target reminder | Anggota yang RSVP going/maybe. Alternatif: semua anggota |
| 3 | Batas jumlah pinned announcement | Tidak dibatasi (bisa di-cap nanti) |
| 4 | Media pengumuman | Maks 5 gambar (post = 10). Bisa disamakan bila perlu |
| 5 | Pengumuman ke non-anggota (public community) | Ikut aturan visibility Community (boleh baca, tak boleh RSVP) |
| 6 | Gating sebagai benefit plan | Belum — kedua permission murni di Role-Permission. Bisa dijadikan benefit kalau diputuskan |

---

*Dokumen ini adalah hasil breakdown fitur Papan Pengumuman & Schedule di modul Community. Untuk skema database lihat `COMMUNITY_ANNOUNCEMENT_SCHEDULE_DB_SCHEMA.md`. Untuk endpoint lihat API spec mobile & backoffice.*
