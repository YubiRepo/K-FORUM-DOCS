# Notification Breakdown Per Modul

Dokumen ini adalah referensi ringkas semua event notifikasi di platform KAI, dikelompokkan per modul. Mencakup siapa penerimanya, channel pengiriman, dan apakah bisa dikontrol user via preferences.

---

## Legend

| Role | Keterangan |
|---|---|
| `member` | User biasa (Standard / Pro) |
| `superadmin` | Admin KAI Pusat |
| `admin` | Admin KAI Regional |
| `leader` | Community Leader |
| `moderator` | Community Moderator |
| `usergod` | Developer / Vendor |

| Kolom | Keterangan |
|---|---|
| **Bypass** | ✅ = selalu dikirim, tidak bisa diblokir user |
| **User Kontrol** | Preference key yang mengontrol notif ini. `—` = tidak ada (notif role internal) |

---

## 1. Announcement

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Announcement `disaster`/`system`/`urgent` dipublish (priority CRITICAL/HIGH) | `member` (global atau per region sesuai scope) | Push + Email + In-app | ✅ | — |
| Announcement `info` dipublish (priority MEDIUM/LOW) | `member` (global atau per region sesuai scope) | Push + In-app | ❌ | `announcement.info_enabled` |

---

## 2. News

### Untuk Member

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Artikel baru dipublish | `member` | Push + In-app | ❌ | `news.enabled` |
| Artikel Pro Member di-approve → published | `member` (penulis) | Push + In-app | ✅ | — |
| Artikel Pro Member di-reject | `member` (penulis) | Push + In-app | ✅ | — |

### Untuk Superadmin (Operasional)

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Artikel Pro Member masuk `pending_approval` | `superadmin` | In-app backoffice | ✅ | — |

---

## 3. Community

### Untuk Member

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Post baru dipublish di komunitas | `member` (anggota komunitas) | Push + In-app | ❌ | `community.communities[id].new_posts` |
| Member lain bergabung | `member` (anggota komunitas) | In-app | ❌ | `community.communities[id].member_joined` |
| Member lain keluar | `member` (anggota komunitas) | In-app | ❌ | `community.communities[id].member_left` |
| Join request di-approve (komunitas private) | `member` (pemohon) | Push + In-app | ❌ | `community.communities[id].join_request_approved` |
| Join request di-reject (komunitas private) | `member` (pemohon) | Push + In-app | ✅ | — |
| Kena kick dari komunitas | `member` | Push + In-app | ✅ | — |
| Kena ban dari komunitas | `member` | Push + In-app | ✅ | — |
| Komunitas di-suspend oleh superadmin | `leader` | Push + In-app | ✅ | — |

### Untuk Leader / Moderator (Operasional)

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Join request masuk baru (komunitas private) | `leader` + `moderator` komunitas tersebut | Push + In-app | ✅ | — |
| Laporan konten masuk (scope komunitas) | `leader` + `moderator` (dengan `manage_reports`) | In-app | ✅ | — |
| Auto-flag: report_count ≥ threshold (scope komunitas) | `leader` + `moderator` + `superadmin` | Push + In-app | ✅ | — |

### Untuk Superadmin (Operasional)

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Komunitas masuk status `orphaned` (leader hapus akun) | `superadmin` | In-app backoffice | ✅ | — |

---

## 4. Event

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Event baru dipublish | `member` (sesuai `interested_categories`, atau semua jika kosong) | Push + In-app | ❌ | `event.enabled` + `event.interested_categories` |
| Reminder sebelum event berlangsung | `member` | Push + In-app | ❌ | `event.reminders_enabled` + `event.reminder_hours_before` |

---

## 5. Q&A

### Untuk Member

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Pertanyaan dijawab oleh superadmin | `member` (penanya) | Push + In-app | ❌ | `qna.question_answered` |
| Pertanyaan ditolak (beserta alasan) | `member` (penanya) | Push + In-app | ❌ | `qna.question_answered` |
| Pertanyaan dikonversi ke FAQ | `member` (penanya) | Push + In-app | ❌ | `qna.question_answered` |

### Untuk Superadmin (Operasional)

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Pertanyaan baru masuk dari member | `superadmin` | In-app backoffice | ✅ | — |

---

## 6. Subscription

### Untuk Member

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Request upgrade disubmit (konfirmasi) | `member` | Push + Email + In-app | ✅ | — |
| Upgrade ke Pro di-approve | `member` | Push + Email + In-app | ✅ | — |
| Upgrade ke Pro di-reject (+ alasan) | `member` | Push + Email + In-app | ✅ | — |
| Plan expired / downgraded ke Standard | `member` | Push + Email + In-app | ✅ | — |
| Expiry reminder 7 hari sebelum expired | `member` (Pro aktif) | Push + Email + In-app | ❌ | `subscription.expiry_reminder_enabled` |
| Expiry reminder 3 hari sebelum expired | `member` (Pro aktif) | Push + In-app | ❌ | `subscription.expiry_reminder_enabled` |

### Untuk Superadmin (Operasional)

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Request upgrade baru masuk dari member | `superadmin` | In-app backoffice | ✅ | — |

---

## 7. Reporting

### Content Report

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Report baru masuk — scope komunitas | `leader` + `moderator` komunitas (dengan `manage_reports`) | In-app | ✅ | — |
| Auto-flag — scope komunitas (report_count ≥ threshold) | `leader` + `moderator` + `superadmin` | Push + In-app | ✅ | — |
| Report baru masuk — scope global (QnA, event, user, dll) | `superadmin` | In-app backoffice | ✅ | — |
| Auto-flag — scope global (report_count ≥ threshold) | `superadmin` | Push + In-app backoffice | ✅ | — |
| Laporan selesai diproses (opsional) | `member` (pelapor) | In-app | ❌ | — (opsional per implementasi) |

### Bug Report

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Bug report diterima (konfirmasi ke pelapor) | `member` (pelapor) | In-app | ✅ | — |
| Bug report baru masuk | `usergod` + `superadmin` | In-app backoffice | ✅ | — |

---

## 8. Region

| Event / Trigger | Penerima | Channel | Bypass | User Kontrol |
|---|---|---|---|---|
| Member disetujui masuk region | `member` | Push + In-app | ✅ | — |
| Member baru bergabung ke region | `admin` (region terkait) | In-app backoffice | ✅ | — |

---

## Ringkasan Bypass vs User-Controlled

### Notifikasi yang Selalu Bypass (tidak bisa dimatikan user)

| Kategori | Event |
|---|---|
| Announcement darurat | Tipe disaster / system / urgent (priority CRITICAL/HIGH) |
| News transaksional | Artikel Pro Member approved / rejected |
| Community status akun | Join rejected, kick, ban, komunitas suspended |
| Community operasional | Join request masuk ke leader/mod, auto-flag report |
| Subscription status | Submit, approved, rejected, expired, downgraded |
| Region | Member disetujui masuk region |
| Reporting | Auto-flag semua scope, bug report konfirmasi & masuk |

### Notifikasi yang Bisa Dikontrol User (via Preferences)

| Preference Key | Mengontrol |
|---|---|
| `announcement.info_enabled` | Announcement tipe info (priority MEDIUM/LOW) |
| `news.enabled` | Semua notif artikel baru published |
| `community.enabled` | Master toggle semua notif komunitas |
| `community.communities[id].new_posts` | Post baru di komunitas tertentu |
| `community.communities[id].member_joined` | Member bergabung di komunitas tertentu |
| `community.communities[id].member_left` | Member keluar di komunitas tertentu |
| `community.communities[id].join_request_approved` | Join request di-approve di komunitas tertentu |
| `event.enabled` | Master toggle semua notif event |
| `event.reminders_enabled` | Reminder event |
| `event.reminder_hours_before` | Berapa jam sebelum event untuk reminder |
| `event.interested_categories` | Filter kategori event |
| `qna.enabled` | Master toggle semua notif Q&A |
| `qna.question_answered` | Notif pertanyaan dijawab / ditolak / jadi FAQ |
| `subscription.expiry_reminder_enabled` | Expiry reminder 7 & 3 hari |

---

*Referensi terkait: `NOTIFICATION_RULES_ENGINE.md` (backend logic), `API_SPEC_NOTIFICATION_PREFERENCES.md` (API mobile), `API_SPEC_FCM.md` (token management)*
