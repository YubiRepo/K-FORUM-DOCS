# Event Module — Rules & Use Cases

Dokumen ini menjelaskan aturan bisnis event module dan bagaimana user berinteraksi dengan fitur ini di aplikasi mobile. Bukan dokumen teknis — fokus pada **siapa bisa apa**, **kapan**, dan **kenapa**.

---

## Siapa yang Bisa Membuat Event?

Hanya **Member Pro** yang bisa membuat event. Member dengan paket Standard tidak bisa, dan akan melihat pesan ajakan upgrade.

| Siapa | Bisa Buat Event? | Catatan |
|-------|-----------------|---------|
| Member Pro | ✅ Ya | Bisa buat, edit, dan cancel event milik sendiri |
| Member Standard | ❌ Tidak | Lihat pesan "Upgrade ke Pro untuk membuat event" |
| Guest (belum login) | ❌ Tidak | Hanya bisa lihat event yang sudah published |
| Superadmin | ✅ Ya | Bisa buat event langsung dari backoffice, langsung published |

---

## Apa yang Bisa Dilakukan User di Event?

Semua user yang sudah login — apapun paketnya — bisa melakukan hal berikut terhadap event yang sudah published:

- **Melihat** daftar dan detail event
- **Menyimpan** event ke daftar bookmark (untuk dibuka lagi nanti)
- **Menambahkan** event ke jadwal pribadi, lengkap dengan pengingat dan opsi export ke kalender
- **Membagikan** event ke WhatsApp, media sosial, atau copy link

Yang **hanya bisa dilakukan Member Pro**:

- Membuat event baru
- Mengedit event milik sendiri (selama masih draft atau rejected)
- Membatalkan event milik sendiri

---

## Jangkauan Event

Event di platform ini **bersifat global** — artinya semua event terlihat oleh semua user di seluruh platform, tidak terbatas oleh region atau wilayah tertentu.

User bisa memfilter event berdasarkan tipe (offline/online/hybrid) atau mencari berdasarkan nama kota/area di field venue — tapi itu hanya untuk memudahkan pencarian, bukan pembatas akses.

---

## Tipe Event & Venue

Event bisa berlangsung secara offline, online, atau hybrid. Ini memengaruhi informasi venue yang perlu diisi.

| Tipe | Venue yang Diisi | Catatan |
|------|-----------------|---------|
| Offline | Nama venue + alamat (string) | Contoh: "GOR Senayan, Jl. Pintu I Senayan, Jakarta" |
| Online | Nama platform + link masuk | Contoh: "Zoom — https://zoom.us/j/..." |
| Hybrid | Keduanya | Ada lokasi fisik sekaligus link online |

Field venue adalah string bebas untuk saat ini — tidak terikat ke data koordinat atau maps. Cukup informatif agar user tahu harus ke mana atau masuk lewat mana.

---

## Bagaimana Event Menjadi Tayang?

Ada dua kemungkinan alur, tergantung konfigurasi yang diatur superadmin:

### Alur 1 — Auto-Publish Aktif

Member Pro membuat event → event **langsung tayang** tanpa perlu review. Cocok untuk platform yang ingin memberikan kepercayaan penuh kepada member pro-nya.

```
Member Pro buat event
        ↓
  Event langsung Published ✅
        ↓
  Semua user bisa lihat & akses
```

### Alur 2 — Perlu Persetujuan Superadmin

Member Pro membuat event → masuk antrian review → Superadmin approve atau reject.

```
Member Pro submit event
        ↓
  Status: Menunggu Persetujuan
        ↓
  Superadmin review
     ↙         ↘
Approve       Reject + alasan
   ↓               ↓
Published      Dikembalikan ke organizer
               (bisa diedit & disubmit ulang)
```

Mana yang dipakai ditentukan oleh superadmin lewat pengaturan platform.

---

## Status Event dan Artinya

Setiap event punya status yang menggambarkan posisinya saat ini.

### Draft
Event sudah dibuat tapi belum dikirim untuk ditayangkan. Hanya organizer yang bisa melihat. Bisa diedit atau dihapus kapan saja.

### Menunggu Persetujuan *(pending_approval)*
Event sudah disubmit oleh organizer dan sedang menunggu review superadmin. Selama menunggu, event tidak bisa diedit (terkunci). Organizer bisa menarik kembali submisi jika ingin merevisi.

### Tayang *(published)*
Event sudah disetujui dan live di platform. Semua user bisa melihat dan mengaksesnya. Tidak bisa dihapus — hanya bisa dibatalkan.

### Ditolak *(rejected)*
Superadmin menolak event dan memberikan alasan. Organizer mendapat notifikasi beserta penjelasan apa yang perlu diperbaiki. Event bisa diedit dan disubmit ulang.

### Dibatalkan *(cancelled)*
Event dibatalkan oleh organizer atau superadmin. Event tetap tersimpan di database tapi tidak muncul di listing publik.

---

## Alur Lengkap Status Event

```
                ┌─────────────────────────────────┐
                │                                 │
              DRAFT ──────────────────────────→ DELETE
                │
                │ (submit)
                ↓
        Menunggu Persetujuan
           ↙          ↘
       Approve        Reject
          ↓               ↓
       TAYANG          DRAFT (bisa edit & resubmit)
          │               │
          │ (cancel)       └──→ DELETE
          ↓
       DIBATALKAN
```

---

## Aturan Edit Event

Organizer hanya bisa mengedit event dalam kondisi tertentu:

| Status Event | Bisa Diedit? | Catatan |
|-------------|-------------|---------|
| Draft | ✅ Ya | Bebas edit semua field |
| Menunggu Persetujuan | ❌ Tidak | Terkunci selama review, tapi bisa ditarik kembali |
| Tayang | ❌ Tidak | Sudah live, tidak bisa diubah dari mobile |
| Ditolak | ✅ Ya | Bisa diedit lalu disubmit ulang |
| Dibatalkan | ❌ Tidak | Event sudah selesai |

Superadmin bisa mengedit event apapun dari backoffice, kapan saja, tanpa batasan status.

---

## Aturan Hapus Event

- Organizer hanya bisa menghapus event yang masih **draft** atau **rejected**
- Event yang sudah **tayang** tidak bisa dihapus — hanya bisa dibatalkan
- Superadmin bisa menghapus event apapun dari backoffice (permanent)

---

## Aturan Pembatalan Event

Organizer bisa membatalkan event milik sendiri, baik yang masih draft maupun yang sudah tayang. Superadmin bisa membatalkan event siapapun.

Saat event yang sudah tayang dibatalkan, semua user yang sudah menyimpan atau menjadwalkan event tersebut akan mendapat notifikasi.

---

## Fitur Jadwal & Kalender

Saat user menambahkan event ke jadwal pribadi, ada dua hal yang bisa dilakukan:

**1. Simpan ke jadwal in-app**
Event muncul di halaman "Jadwalku" di dalam aplikasi. User bisa mengaktifkan pengingat (reminder) yang akan mengirim notifikasi push dari aplikasi sebelum event berlangsung.

**2. Export ke kalender eksternal**
User bisa menambahkan event langsung ke kalender di perangkat mereka — Google Calendar, Apple Calendar, atau kalender lainnya yang mendukung format standar. Ini dilakukan dengan mengunduh file `.ics` atau melalui integrasi deep link ke aplikasi kalender.

Kedua opsi ini bisa dilakukan bersamaan — user bisa menyimpan ke jadwal in-app sekaligus export ke Google Calendar.

---

## Use Cases Mobile

Berikut skenario nyata bagaimana user berinteraksi dengan fitur event di aplikasi.

---

### Use Case 1 — Member Pro Buat Event Offline, Langsung Tayang

**Siapa:** Andi, Member Pro
**Kondisi:** Platform menggunakan auto-publish

Andi buka aplikasi, pergi ke menu Events, lalu tap tombol "Buat Event". Ia memilih tipe **Offline**, mengisi judul, deskripsi, mengunggah 3 foto venue, memilih kategori Sports, mengisi nama venue "GOR Senayan" beserta alamatnya, mengatur tanggal dan jam, lalu menambahkan link pendaftaran dari Eventbrite.

Setelah tap "Publish", event Andi langsung tayang. Ia bisa langsung lihat eventnya muncul di listing dan membagikannya ke teman-teman.

---

### Use Case 2 — Member Pro Buat Event Online, Perlu Persetujuan

**Siapa:** Budi, Member Pro
**Kondisi:** Platform menggunakan manual approval

Budi membuat event webinar dengan tipe **Online**. Ia mengisi nama platform "Zoom" dan link meeting. Setelah tap "Submit", eventnya masuk ke antrian review dengan status "Menunggu Persetujuan". Budi melihat banner: *"Event sedang direview oleh tim kami."*

Superadmin menerima notifikasi, membuka detail event Budi, memeriksanya, lalu menekan Approve. Budi langsung mendapat notifikasi bahwa eventnya sudah tayang.

---

### Use Case 3 — Event Ditolak, Organizer Revisi & Submit Ulang

**Siapa:** Budi, Member Pro
**Kondisi:** Superadmin menolak event karena info venue kurang lengkap

Budi mendapat notifikasi: *"Event kamu ditolak. Alasan: Alamat venue belum lengkap, mohon cantumkan alamat lengkap termasuk nama gedung dan lantai."*

Budi membuka eventnya, melihat status Ditolak beserta alasan dari superadmin. Ia melengkapi info venue, lalu menekan "Submit Ulang". Event kembali masuk antrian review.

---

### Use Case 4 — Member Standard Coba Buat Event

**Siapa:** Citra, Member Standard

Citra membuka halaman Events dan mencari tombol "Buat Event". Saat ia mengetuknya, muncul modal yang menjelaskan bahwa fitur ini hanya untuk Member Pro, beserta tombol "Upgrade ke Pro". Citra tidak bisa melanjutkan tanpa upgrade.

---

### Use Case 5 — User Bookmark & Tambahkan Event ke Jadwal + Kalender

**Siapa:** Doni, Member Standard

Doni melihat event "Futsal Tournament 2026" di feed. Ia mengetuk ikon bookmark untuk menyimpannya — event masuk ke halaman "Tersimpan" miliknya.

Lalu ia membuka detail event, mengetuk "Tambah ke Jadwal", mengaktifkan pengingat 1 hari sebelum event, dan menambahkan catatan pribadi. Di langkah yang sama, ia juga mengetuk "Tambah ke Google Calendar" — aplikasi Google Calendar terbuka dan event langsung masuk ke tanggal yang sesuai.

Sehari sebelum event, Doni mendapat notifikasi push dari aplikasi sebagai pengingat.

---

### Use Case 6 — User Bagikan Event

**Siapa:** Doni, Member Standard

Dari halaman detail event, Doni mengetuk tombol Share. Ia memilih WhatsApp dan mengetikkan pesan singkat sebelum mengirimkan link ke grup teman-temannya. Platform mencatat aksi share ini.

---

### Use Case 7 — Organizer Batalkan Event

**Siapa:** Andi, Member Pro (organizer)
**Kondisi:** Event sudah tayang, 2 hari sebelum pelaksanaan

Andi mendapat kabar bahwa venue tidak tersedia. Ia membuka halaman event miliknya, mengetuk "Batalkan Event", mengisi alasan pembatalan, lalu mengkonfirmasi. Event berubah status menjadi Dibatalkan, dan semua user yang sudah menjadwalkan event ini mendapat notifikasi pembatalan.

---

### Use Case 8 — User Browse Event by Tipe & Lokasi

**Siapa:** Eka, Member Standard

Eka membuka halaman Events. Ia memfilter event dengan tipe "Online" untuk mencari webinar yang bisa diikuti dari rumah. Lain waktu ia memfilter tipe "Offline" dan mencari "Surabaya" untuk menemukan event yang bisa didatangi langsung. Semua event tetap bisa ditemukan karena bersifat global — filter hanya membantu mempersempit pilihan.

---

## Ringkasan Aturan

| Aturan | Detail |
|--------|--------|
| **Siapa yang bisa buat** | Member Pro saja |
| **Persetujuan** | Superadmin (bisa diaktifkan/dinonaktifkan) |
| **Jangkauan event** | Global — semua user bisa lihat |
| **Tipe event** | Offline, Online, atau Hybrid |
| **Venue** | String bebas — nama venue/platform + alamat/link |
| **Gambar** | Bisa lebih dari satu, diunggah terpisah sebelum submit |
| **Edit** | Hanya saat draft atau rejected |
| **Hapus** | Hanya saat draft atau rejected |
| **Batalkan** | Bisa kapan saja selama event belum lewat |
| **Bookmark** | Semua user yang sudah login |
| **Jadwal & Kalender** | Semua user yang sudah login, bisa export ke kalender eksternal |

---

*Dokumen rules event module — non-teknis. Untuk detail API lihat API_SPEC_EVENT_MOBILE dan API_SPEC_EVENT_BACKOFFICE.*
