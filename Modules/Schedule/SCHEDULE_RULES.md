# Schedule Module — Rules & Use Cases (v1.1)

Dokumentasi sistem rules modul Schedule KAI App. Schedule adalah **kalender agenda KAI yang dikelola di backoffice oleh admin/staff** — tempat mencatat agenda apa pun (rapat, deadline, kegiatan internal, milestone, dll), diinput manual dulu, dengan hook untuk menarik jadwal dari modul lain (Event, Announcement) di masa depan.

> **Posisi modul:** Schedule BUKAN pengganti modul Event. Event = konten acara publik untuk dilihat member di mobile. Schedule = agenda internal/operasional yang dikelola admin di backoffice. Schedule juga berbeda dari fitur "Add to Schedule" milik member di mobile (itu jadwal pribadi user terhadap event publik).
>
> **Catatan v1.1:** Model sharing diperjelas menjadi tiga mode (`private` / `all_admins` / `specific`-invite). Schedule **tidak region-aware** pada versi ini, tapi desain sharing dibuat agar region bisa ditambahkan nanti sebagai mode baru tanpa perombakan.

---

## 1. APA ITU MODUL SCHEDULE?

**Schedule** adalah kalender agenda backoffice. Admin/staff mencatat agenda dan melihatnya dalam tampilan kalender (bulan/minggu/hari).

### Karakteristik:
- **Surface:** Backoffice only (tidak ada di mobile member pada Phase 1)
- **Pengguna:** Semua admin/staff backoffice (dikontrol via role/permission)
- **Fungsi inti:** CRUD agenda, tampilan kalender, filter, tandai selesai, sharing
- **Sumber data:** Manual (diketik admin) + future linked dari modul lain (Event, Announcement)
- **Sharing:** Tiga mode — `private`, `all_admins`, `specific` (invite user tertentu)

### Berbeda dari:
- **Event:** acara publik untuk member. Schedule = agenda internal admin.
- **Event "Add to Schedule" (mobile):** jadwal pribadi member terhadap event. Schedule = agenda yang dikelola admin di backoffice.
- **Announcement:** broadcast ke member. Schedule = catatan agenda, bukan komunikasi keluar.

---

## 2. KONSEP INTI — SCHEDULE ENTRY

Setiap agenda = satu **entry**:

```
schedule_entry
  ├── title          : judul agenda
  ├── description    : detail (opsional)
  ├── entry_type     : agenda | reminder | milestone (extensible)
  ├── start_at       : waktu mulai
  ├── end_at         : waktu selesai (opsional)
  ├── all_day        : true = sepanjang hari (abaikan jam)
  ├── location       : teks bebas (opsional)
  ├── visibility     : private | all_admins | specific   (lihat §4)
  ├── status         : active | done | cancelled
  ├── source         : manual | linked         (Phase 1 selalu 'manual')
  ├── source_module  : null | event | announcement | ...   (future)
  ├── source_ref     : id entitas asal, mis. event_id      (future)
  ├── recurrence     : null | rule (RRULE-style)           (future)
  ├── assigned_to    : user_id admin             (future, opsional)
  ├── created_by
  ├── created_at
  └── updated_at

schedule_entry_share   (dipakai hanya saat visibility = specific)
  ├── schedule_entry_id
  └── user_id          : admin/staff yang diundang melihat
```

### Kenapa desain ini?
- **`visibility` + tabel share terpisah** → fleksibel: pribadi, ke semua tim, atau ke orang tertentu. Tabel `schedule_entry_share` jadi fondasi yang bisa diperluas (mis. region nanti tinggal mengisi daftar share otomatis).
- **`source` + `source_module` + `source_ref`** → hook future-proof: nanti Event/Announcement bisa muncul di kalender tanpa duplikasi data (entri `linked` menunjuk balik ke sumbernya). Phase 1 selalu `manual`.
- **`recurrence`** → disiapkan untuk agenda berulang (rapat mingguan), belum aktif di Phase 1.
- **`all_day`** → membedakan agenda berjam vs sepanjang hari (deadline, hari libur).

---

## 3. ENTRY TYPE

| Type | Keterangan |
|------|-----------|
| `agenda` | Agenda umum (rapat, kegiatan, acara internal) |
| `reminder` | Pengingat singkat (deadline, follow-up) |
| `milestone` | Penanda penting (target, tenggat proyek) |

> Daftar ini extensible — bisa ditambah Superadmin via master data tanpa deploy ulang (pola sama dengan kategori di modul lain).

---

## 4. SHARING / VISIBILITY (inti v1.1)

Setiap agenda punya satu **mode visibility**:

| Mode | Siapa yang bisa LIHAT | Kegunaan |
|------|----------------------|----------|
| `private` | Hanya pembuat | Agenda pribadi (catatan personal, draft) |
| `all_admins` | Semua admin/staff backoffice | Agenda yang seluruh tim perlu tahu |
| `specific` | Pembuat + daftar user yang di-invite | Rapat/agenda untuk beberapa orang tertentu |

### Aturan sharing
- **Default `private`** saat buat (aman; admin pilih mode lain bila ingin dibagikan).
- **Sharing = view-only.** User yang di-invite (mode `specific`) atau semua admin (mode `all_admins`) hanya bisa **melihat** agenda di kalendernya — **tidak bisa mengedit/menghapus**.
- **Edit/hapus tetap milik pembuat** (atau Superadmin), terlepas dari mode.
- **Mode `specific`** memakai tabel `schedule_entry_share`: pembuat memilih satu/lebih user backoffice yang sudah ada (bukan email orang luar).
- **Mengubah mode** boleh kapan saja oleh pembuat. Pindah ke selain `specific` akan mengabaikan daftar share (tetap tersimpan, tidak dipakai).
- **Superadmin** bisa melihat semua agenda apa pun mode-nya (override).

### Catatan region (future)
Schedule **tidak region-aware** di v1.1. Bila nanti perlu "share ke semua admin region X", itu ditambahkan sebagai **mode baru** (`region`) di atas struktur yang sama — region hanya menjadi cara cepat mengisi daftar `schedule_entry_share`. Struktur inti tidak berubah. Karena itu keputusan region bisa ditunda tanpa risiko refactor.

---

## 5. ACTOR & PERMISSION MATRIX

Diakses semua admin/staff backoffice, dikontrol via role/permission.

| Aksi | Admin/Staff | Superadmin |
|------|:---:|:---:|
| Buat agenda | ✅ | ✅ |
| Lihat agenda sendiri (semua mode milik sendiri) | ✅ | ✅ |
| Lihat agenda `all_admins` | ✅ | ✅ |
| Lihat agenda `specific` yang meng-invite dirinya | ✅ | ✅ |
| Lihat agenda `private`/`specific` milik orang lain (tak di-invite) | ❌ | ✅ |
| Invite/keluarkan user pada agenda sendiri (`specific`) | ✅ | ✅ |
| Edit/hapus agenda sendiri | ✅ | ✅ |
| Edit/hapus agenda orang lain | ❌ | ✅ |
| Tandai selesai (done) agenda sendiri | ✅ | ✅ |
| Kelola entry_type master | ❌ | ✅ |

---

## 6. STATUS LIFECYCLE

```
active → done        (agenda selesai dijalankan)
active → cancelled   (agenda dibatalkan)
```

| Status | Deskripsi | Masuk kalender aktif? |
|--------|-----------|:---:|
| `active` | Agenda terjadwal, belum lewat/selesai | ✅ |
| `done` | Sudah dilaksanakan/diselesaikan | ✅ (ditandai selesai) |
| `cancelled` | Dibatalkan — tetap tersimpan, ditandai batal | ditampilkan dengan tanda batal |

> Agenda tidak dihapus permanen kecuali oleh pembuat/Superadmin. `cancelled` lebih disarankan daripada delete untuk jejak.

---

## 7. USE CASES

### Use Case 1: Rapat tim (all_admins)
```
Admin → Schedule → "Buat Agenda"
  ↓
title="Rapat Koordinasi Mingguan", entry_type=agenda
start_at=2026-06-16 10:00, end_at=11:00, visibility=all_admins
  ↓
Submit → muncul di kalender semua admin/staff
```

### Use Case 2: Reminder pribadi (private)
```
Admin → Schedule → "Buat Agenda"
  ↓
title="Follow up vendor X", entry_type=reminder
start_at=2026-06-18, all_day=true, visibility=private
  ↓
Submit → hanya muncul di kalender admin tsb
```

### Use Case 3: Rapat dengan orang tertentu (specific + invite)
```
Admin → Schedule → "Buat Agenda"
  ↓
title="Review modul Schedule", visibility=specific
invite: [admin_B, admin_C]
  ↓
Submit → muncul di kalender pembuat, admin_B, admin_C (view-only)
```

### Use Case 4: Lihat kalender bulan ini
```
Admin → Schedule → tampilan kalender
  ↓
Filter: rentang 2026-06-01 s/d 2026-06-30
  ↓
Tampil: agenda milik sendiri (semua mode) + all_admins
        + specific yang meng-invite dirinya
```

### Use Case 5: Tandai agenda selesai
```
Admin → buka agenda → "Tandai Selesai" → status=done
```

---

## 8. ATURAN BISNIS & EDGE CASES

- **Default private** — agenda baru `private` kecuali admin memilih mode lain.
- **Sharing view-only** — yang di-invite/semua admin hanya bisa lihat; edit/hapus tetap pembuat (atau Superadmin).
- **Invite hanya user internal** — mode `specific` memilih user backoffice yang sudah ada, bukan email orang luar. (Untuk berbagi keluar, pakai export `.ics` di Phase 3.)
- **end_at opsional** — jika kosong, agenda dianggap titik waktu tunggal (atau sepanjang hari jika `all_day`).
- **all_day** — saat true, komponen jam diabaikan; agenda menempati seluruh tanggal.
- **Validasi waktu** — `end_at` (jika ada) harus ≥ `start_at`.
- **Cancelled tidak hilang** — tetap tampil dengan penanda batal; bisa dihapus permanen oleh pembuat.
- **Ganti mode** — boleh kapan saja; daftar share (`specific`) tetap tersimpan walau mode dipindah, dipakai lagi bila kembali ke `specific`.
- **Phase 1 manual** — `source` selalu `manual`. Entri `linked` baru muncul di Phase 2.

---

## 9. FUTURE-PROOF — INTEGRASI YANG DISIAPKAN

Pola: **simpan hook sekarang, fungsi menyusul tanpa refactor.**

| Integrasi masa depan | Yang sudah disiapkan sekarang |
|----------------------|-------------------------------|
| **Share ke region** (semua admin region X) | Mode `region` baru di atas tabel `schedule_entry_share` (region mengisi daftar share otomatis) — struktur inti tak berubah |
| **Tarik Event ke kalender** | `source=linked`, `source_module=event`, `source_ref=event_id` |
| **Tarik Announcement terjadwal** | `source_module=announcement` |
| **Agenda berulang** (rapat mingguan) | Field `recurrence` (RRULE-style) |
| **Reminder/notifikasi** ke admin/diundang sebelum agenda | Hook ke modul Notification (FCM) yang sudah ada |
| **Export kalender** (.ics / Google Calendar) | Pola sama dengan Event (`.ics` + deep link) |
| **Jadwal publik** (agenda KAI tampil ke member) | Mode visibility diperluas + endpoint mobile read-only nanti |
| **Assignment** (agenda di-assign ke admin tertentu) | Field `assigned_to` |

### Roadmap
- **Phase 1:** CRUD agenda manual, kalender, filter, sharing (private/all_admins/specific+invite), tandai selesai. Field hook disiapkan tapi belum dipakai.
- **Phase 2:** Linked entries dari Event & Announcement (read-only). Reminder via Notification. (Opsional) mode `region` bila diputuskan perlu.
- **Phase 3:** Recurrence, export .ics, assignment, jadwal publik ke mobile.

---

## 10. INTEGRATION POINTS

| Modul | Hubungan |
|-------|----------|
| **Event** | Future: event tayang muncul sebagai entri `linked` (`source_module=event`). |
| **Announcement** | Future: announcement terjadwal muncul sebagai entri `linked`. |
| **Notification** | Future: reminder agenda dikirim via FCM ke pembuat & user yang di-invite. |
| **Role/Permission** | Menentukan siapa boleh akses & kelola agenda. |
| **Region (future, opsional)** | Bila diaktifkan: mode `region` mengisi daftar share dari anggota region. |

---

## 11. SUMMARY

Schedule = **kalender agenda backoffice** dengan sharing tiga mode: `private`, `all_admins`, dan `specific` (invite user tertentu, view-only). Tabel `schedule_entry_share` menjadi fondasi yang fleksibel — region bisa ditambahkan nanti sebagai mode baru tanpa perombakan. Phase 1 murni manual, tapi setiap entri dirancang bisa di-link dari modul lain (Event, Announcement) lewat `source`/`source_ref`, menjadikan Schedule siap berkembang jadi pusat agregasi agenda KAI tanpa duplikasi data.
