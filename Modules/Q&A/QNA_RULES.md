# Q&A / FAQ Module — Rules & Use Cases

Dokumen ini menjelaskan aturan bisnis modul Q&A / FAQ dan bagaimana user berinteraksi dengan fitur ini di aplikasi mobile. Bukan dokumen teknis — fokus pada **siapa bisa apa**, **kapan**, dan **kenapa**.

---

## Apa Itu Modul Q&A / FAQ?

Modul Q&A / FAQ adalah pusat informasi tanya jawab di aplikasi KAI. Berisi dua jenis konten:

**FAQ (Frequently Asked Questions)** — Kumpulan pertanyaan dan jawaban yang sudah dikurasi dan dipublish oleh superadmin. Member bisa membaca, mencari, dan memberi penilaian apakah konten ini membantu atau tidak.

**Pertanyaan Member (Q&A)** — Member bisa mengajukan pertanyaan sendiri jika tidak menemukan jawaban di FAQ. Pertanyaan masuk ke antrian moderasi, lalu dijawab oleh superadmin.

Topik yang dibahas mencakup dua area besar:
- **Administrasi di Indonesia** — imigrasi & visa, pajak, ketenagakerjaan, pendirian usaha, properti, dan sejenisnya
- **Sistem Aplikasi KAI** — panduan penggunaan fitur, subscription, komunitas, dan lain-lain

---

## Siapa yang Bisa Melakukan Apa?

### Member (Standard & Pro)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Membaca FAQ | ✅ Ya | Semua member bisa membaca semua FAQ yang sudah published |
| Mencari FAQ | ✅ Ya | Full-text search di seluruh FAQ |
| Memberi vote helpful / not helpful | ✅ Ya | Satu vote per FAQ per member, bisa diubah |
| Mengajukan pertanyaan | ✅ Ya* | *Tergantung konfigurasi benefit `ask_qna` di plan |
| Melihat riwayat pertanyaan sendiri | ✅ Ya | Hanya bisa lihat pertanyaan milik sendiri |
| Melihat pertanyaan orang lain | ❌ Tidak | Pertanyaan bersifat privat antara member dan admin |

> **Catatan benefit `ask_qna`:** Superadmin bisa mengatur apakah fitur mengajukan pertanyaan hanya tersedia untuk Pro, atau untuk semua member termasuk Standard. Jika Standard tidak memiliki benefit ini, mereka akan melihat pesan ajakan upgrade.

### Guest (belum login)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Membaca FAQ | ❌ Tidak | Harus login terlebih dahulu |
| Mengajukan pertanyaan | ❌ Tidak | Harus login |

### Superadmin

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Membuat & mengelola kategori FAQ | ✅ Ya | CRUD kategori, atur urutan, aktifkan/nonaktifkan |
| Membuat & mengelola FAQ | ✅ Ya | CRUD FAQ, pilih status draft/published/archived, pin item |
| Melihat semua pertanyaan masuk | ✅ Ya | Termasuk antrian pending dari semua member |
| Menjawab pertanyaan member | ✅ Ya | Jawaban bisa diset publik atau privat |
| Menolak pertanyaan | ✅ Ya | Dengan menyertakan alasan yang dikirim ke member |
| Mengonversi pertanyaan menjadi FAQ | ✅ Ya | Saat menjawab, jawaban bisa langsung dijadikan FAQ publik |
| Mengatur konfigurasi bot fallback | ✅ Ya | Aktifkan/nonaktifkan, edit pesan otomatis |

---

## Bagaimana FAQ Dikelola?

FAQ adalah konten resmi yang dikurasi oleh superadmin. Tidak ada member yang bisa membuat FAQ secara langsung — namun pertanyaan member yang dijawab bisa **dikonversi** menjadi FAQ publik oleh superadmin.

### Status FAQ

**Draft** — FAQ sudah dibuat tapi belum dipublish. Hanya terlihat oleh superadmin di backoffice. Bisa diedit kapan saja.

**Published** — FAQ aktif dan terlihat oleh semua member di aplikasi. Ini adalah kondisi normal FAQ yang berjalan.

**Archived** — FAQ dinonaktifkan dan tidak muncul di aplikasi, tapi data tetap tersimpan untuk referensi. Tidak dihapus permanen.

### Alur Status FAQ

```
Draft ──→ Published ──→ Archived
  ↑             │
  └─────────────┘
   (bisa kembali ke draft)
```

### Kategori FAQ

Setiap FAQ harus masuk ke dalam satu kategori. Kategori dikelola oleh superadmin — tidak bisa dibuat oleh member. Contoh kategori default:

- Imigrasi & Visa
- Pajak
- Ketenagakerjaan
- Pendirian Usaha
- Properti & Kepemilikan
- Sistem Aplikasi KAI
- Umum

Superadmin bisa menambah, mengubah, menonaktifkan, atau menghapus kategori. Kategori yang masih memiliki FAQ aktif tidak bisa dihapus langsung — FAQ harus diarsipkan atau dipindahkan terlebih dahulu.

### Pinned FAQ

Superadmin bisa menandai FAQ tertentu sebagai *pinned* — FAQ tersebut akan selalu muncul di posisi paling atas dalam kategorinya, terlepas dari urutan normal. Berguna untuk FAQ yang paling sering dicari atau paling penting.

---

## Bagaimana Member Mengajukan Pertanyaan?

Jika member tidak menemukan jawaban di FAQ setelah membaca dan mencari, ia bisa mengajukan pertanyaan sendiri.

### Alur Pengajuan Pertanyaan

```
Member tidak menemukan jawaban di FAQ
          ↓
  Tap "Ajukan Pertanyaan"
          ↓
  Isi form: pilih kategori (opsional) + tulis pertanyaan
          ↓
  Submit → Bot fallback merespons otomatis (jika aktif)
          ↓
  Status: Menunggu Jawaban (pending)
          ↓
  Superadmin review antrian
       ↙             ↘
  Jawab              Tolak + alasan
     ↓                    ↓
  Status: Terjawab    Status: Ditolak
     ↓                    ↓
  Member dapat         Member dapat
  notifikasi           notifikasi + alasan
```

### Bot Fallback

Saat pertanyaan baru masuk, sistem secara otomatis mengirim pesan konfirmasi kepada member bahwa pertanyaannya diterima dan sedang diproses. Pesan ini dikonfigurasi oleh superadmin — bukan jawaban sungguhan, melainkan acknowledgement otomatis.

Bot juga bisa dikonfigurasi untuk memberikan saran kontak admin jika member membutuhkan respons lebih cepat.

---

## Status Pertanyaan Member dan Artinya

### Menunggu Jawaban *(pending)*
Pertanyaan sudah dikirim dan masuk ke antrian moderasi superadmin. Member melihat status ini dan bisa menunggu — tidak ada tindakan lain yang perlu dilakukan.

### Terjawab *(answered)*
Superadmin sudah memberikan jawaban. Member mendapat notifikasi push dan bisa membaca jawaban lengkap di halaman detail pertanyaannya.

### Ditolak *(rejected)*
Superadmin menolak pertanyaan — biasanya karena sudah ada di FAQ, duplikat, atau di luar scope platform. Member mendapat notifikasi beserta alasan penolakan, dan bisa diarahkan ke FAQ yang relevan.

### Dikonversi ke FAQ *(converted)*
Superadmin menilai pertanyaan ini cukup umum dan penting untuk semua member. Jawaban yang diberikan sekaligus dijadikan FAQ publik yang bisa diakses semua orang. Member yang bertanya tetap mendapat notifikasi bahwa pertanyaannya sudah dijawab.

---

## Alur Lengkap Status Pertanyaan

```
Member submit pertanyaan
          ↓
       PENDING
       ↙     ↘
  Jawab      Tolak
     ↓          ↓
 ANSWERED    REJECTED
     │
     │ (jika jawaban dijadikan FAQ publik)
     ↓
 CONVERTED
```

---

## Aturan Privasi Pertanyaan

Pertanyaan yang diajukan member bersifat **privat secara default** — hanya member yang bersangkutan dan superadmin yang bisa melihatnya. Member lain tidak bisa melihat pertanyaan orang lain.

Namun jawaban bisa bersifat publik atau privat, tergantung pilihan superadmin saat menjawab:

- **Publik** — Jawaban terlihat oleh member yang bertanya (default). Jika dikonversi ke FAQ, jawaban menjadi konten publik yang bisa dibaca semua member.
- **Privat** — Jawaban hanya terlihat oleh member yang bertanya. Cocok untuk pertanyaan yang sifatnya personal atau sensitif.

---

## Sistem Vote Helpful

Setiap FAQ yang sudah published bisa diberi penilaian oleh member — apakah konten ini membantu atau tidak.

**Aturan vote:**
- Satu member hanya bisa memberi satu vote per FAQ
- Vote bisa diubah kapan saja (dari helpful ke not helpful, atau sebaliknya)
- Hasil vote ditampilkan sebagai angka di halaman detail FAQ
- Superadmin bisa melihat statistik vote di backoffice sebagai bahan evaluasi kualitas konten

Vote tidak mempengaruhi urutan tampil FAQ — itu tetap ditentukan oleh superadmin melalui `sort_order` dan `is_pinned`.

---

## Notifikasi Q&A

Member menerima notifikasi push dalam dua situasi:

1. **Pertanyaan dijawab** — Saat superadmin memberikan jawaban atas pertanyaan yang diajukan member
2. **Pertanyaan ditolak** — Saat superadmin menolak pertanyaan, beserta alasan penolakan

Member bisa mengatur preferensi notifikasi Q&A di pengaturan aplikasi — bisa dinonaktifkan secara keseluruhan atau per jenis notifikasi.

---

## Use Cases

Berikut skenario nyata bagaimana user berinteraksi dengan modul Q&A di aplikasi.

---

### Use Case 1 — Member Cari Informasi Pajak dan Menemukan Jawabannya di FAQ

**Siapa:** Minji Park, Member Standard

Minji baru tiba di Indonesia dan perlu tahu cara mengurus NPWP. Ia membuka aplikasi KAI, pergi ke modul Q&A, dan memilih kategori **Pajak**. Di sana ia menemukan FAQ yang berjudul "Bagaimana cara mendaftar NPWP sebagai WNA?" — lengkap dengan langkah-langkahnya.

Setelah membaca, Minji mengetuk tombol "Helpful" karena informasi ini menjawab kebutuhannya. Ia tidak perlu mengajukan pertanyaan baru.

---

### Use Case 2 — Member Cari Lewat Search dan Tidak Menemukan, Lalu Ajukan Pertanyaan

**Siapa:** Park Joon, Member Standard

Park Joon mencari informasi soal BPJS Ketenagakerjaan untuk WNA. Ia mengetik "BPJS WNA" di kolom pencarian — hasilnya ada beberapa FAQ tentang BPJS Kesehatan, tapi tidak ada yang menjawab pertanyaannya soal BPJS Ketenagakerjaan secara spesifik.

Ia mengetuk "Ajukan Pertanyaan", memilih kategori **Ketenagakerjaan**, lalu menuliskan pertanyaannya. Setelah submit, ia melihat pesan otomatis: *"Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja."*

Dua hari kemudian, Park Joon mendapat notifikasi bahwa pertanyaannya sudah dijawab. Ia membuka aplikasi dan membaca jawaban lengkap dari Admin KAI.

---

### Use Case 3 — Pertanyaan Member Dijadikan FAQ Publik

**Siapa:** Kim Soo, Member Pro

Kim Soo mengajukan pertanyaan tentang perbedaan KITAS dan KITAP yang cukup detail. Superadmin menilai pertanyaan ini sangat umum dan pasti akan ditanyakan oleh banyak member lain.

Saat menjawab, superadmin mengaktifkan opsi "Konversi ke FAQ" dan memilih kategori **Imigrasi & Visa**. Jawaban otomatis dipublish sebagai FAQ publik baru.

Kim Soo menerima notifikasi bahwa pertanyaannya sudah dijawab. Kini FAQ tentang KITAS vs KITAP bisa dinikmati semua member tanpa perlu mengajukan pertanyaan yang sama.

---

### Use Case 4 — Pertanyaan Ditolak karena Duplikat

**Siapa:** Lee Jae, Member Standard

Lee Jae mengajukan pertanyaan "Cara daftar NPWP WNA gimana ya?" — tanpa memeriksa FAQ lebih dulu. Superadmin melihat bahwa pertanyaan ini sudah terjawab di FAQ yang ada.

Superadmin menolak pertanyaan dan menyertakan alasan: *"Pertanyaan ini sudah tersedia di FAQ kami: 'Bagaimana cara mendaftar NPWP sebagai WNA?' — silakan cek di kategori Pajak."*

Lee Jae mendapat notifikasi beserta alasan penolakan, dan bisa langsung diarahkan ke FAQ yang relevan.

---

### Use Case 5 — Member Standard Coba Ajukan Pertanyaan tapi Tidak Punya Benefit

**Siapa:** Han Na, Member Standard
**Kondisi:** Superadmin mengatur benefit `ask_qna` hanya untuk Member Pro

Han Na ingin mengajukan pertanyaan tapi saat mengetuk "Ajukan Pertanyaan", muncul modal yang menjelaskan bahwa fitur ini memerlukan upgrade ke Pro. Ia bisa memilih untuk upgrade atau menutup modal dan hanya membaca FAQ yang tersedia.

---

### Use Case 6 — Superadmin Kelola Kategori dan FAQ Baru

**Siapa:** Admin KAI Pusat

Admin KAI mendapat masukan bahwa banyak member menanyakan soal investasi properti. Ia membuka backoffice, membuat kategori baru **Properti & Kepemilikan**, lalu membuat beberapa FAQ berisi pertanyaan-pertanyaan umum yang paling sering ditanyakan.

FAQ dibuat dengan status **Draft** terlebih dahulu — admin bisa memeriksa ulang konten, meminta review dari rekan, baru kemudian mengubah status ke **Published** agar langsung muncul di aplikasi member.

---

### Use Case 7 — Superadmin Atur Ulang Urutan FAQ

**Siapa:** Admin KAI Pusat

Admin ingin memastikan FAQ yang paling penting mudah ditemukan. Ia membuka backoffice, pergi ke kategori Pajak, lalu mengatur ulang urutan FAQ dengan drag-and-drop. FAQ soal NPWP ia pindahkan ke posisi pertama dan menandainya sebagai **pinned** — sehingga selalu muncul di bagian teratas meski urutan lain berubah.

---

### Use Case 8 — Member Lihat Riwayat Pertanyaan Sendiri

**Siapa:** Park Joon, Member Standard

Park Joon ingin mengecek apakah pertanyaannya dari minggu lalu sudah dijawab. Ia membuka tab **Riwayat Pertanyaan**, dan melihat daftar semua pertanyaan yang pernah ia ajukan beserta statusnya: satu *terjawab*, satu masih *menunggu jawaban*.

Ia mengetuk pertanyaan yang sudah terjawab untuk membaca jawaban lengkapnya.

---

## Ringkasan Aturan

| Aturan | Detail |
|--------|--------|
| **Siapa yang bisa baca FAQ** | Semua member yang sudah login |
| **Siapa yang bisa ajukan pertanyaan** | Member dengan benefit `ask_qna` (Standard atau Pro, tergantung konfigurasi) |
| **Siapa yang bisa buat/kelola FAQ** | Superadmin saja |
| **Privasi pertanyaan** | Privat — hanya bisa dilihat member yang mengajukan dan superadmin |
| **Status FAQ** | `draft` → `published` → `archived` |
| **Status pertanyaan** | `pending` → `answered` / `rejected` / `converted` |
| **Bot fallback** | Pesan otomatis saat pertanyaan masuk, bisa diaktifkan/dinonaktifkan |
| **Vote helpful** | Satu vote per member per FAQ, bisa diubah kapan saja |
| **Konversi ke FAQ** | Superadmin bisa konversi jawaban pertanyaan menjadi FAQ publik langsung saat menjawab |
| **Notifikasi** | Push notification saat pertanyaan dijawab atau ditolak |
| **Preferensi notifikasi** | Member bisa atur di pengaturan notifikasi per jenis |
| **Hapus pertanyaan** | Member tidak bisa menghapus pertanyaan yang sudah diajukan |

---

*Dokumen rules Q&A / FAQ module — non-teknis. Untuk detail API lihat API_SPEC_QNA_MOBILE dan API_SPEC_QNA_BACKOFFICE. Untuk skema database lihat QNA_DB_SCHEMA.*
