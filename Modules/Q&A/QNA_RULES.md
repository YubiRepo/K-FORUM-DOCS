# Q&A / FAQ Module — Rules & Use Cases

> **Revisi 2026-06-25:** Modul kini mendukung forum komunitas — member bisa menjawab pertanyaan publik member lain (gated benefit `answer_qna`), jawaban valid ditandai oleh expert/admin (boleh lebih dari satu), pertanyaan bisa di-assign ke expert, dan moderasi jawaban bisa diatur `auto`/`manual`. Word filter ditangani backend via modul System Settings.

Dokumen ini menjelaskan aturan bisnis modul Q&A / FAQ dan bagaimana user berinteraksi dengan fitur ini di aplikasi mobile. Bukan dokumen teknis — fokus pada **siapa bisa apa**, **kapan**, dan **kenapa**.

---

## Apa Itu Modul Q&A / FAQ?

Modul Q&A / FAQ adalah pusat informasi tanya jawab di aplikasi KAI. Berisi dua jenis konten:

**FAQ (Frequently Asked Questions)** — Kumpulan pertanyaan dan jawaban yang sudah dikurasi dan dipublish oleh superadmin. Member bisa membaca, mencari, dan memberi penilaian apakah konten ini membantu atau tidak.

**Pertanyaan Member (Q&A)** — Member bisa mengajukan pertanyaan sendiri jika tidak menemukan jawaban di FAQ. Saat mengajukan, member memilih visibilitas pertanyaan:

- **Privat** — pertanyaan hanya terlihat oleh penanya dan admin/expert yang ditugaskan. Dijawab secara 1-on-1, tidak masuk forum komunitas.
- **Publik** — setelah disetujui admin, pertanyaan masuk ke **forum komunitas**. Member lain (yang punya benefit menjawab) bisa ikut menjawab. Jawaban yang dinilai tepat bisa ditandai sebagai **rujukan valid** oleh expert/admin, dan member bisa memberi **upvote** pada jawaban yang membantu.

Pertanyaan publik maupun privat bisa **ditugaskan (assign)** oleh admin ke seorang expert atau staf sebagai penanggung jawab.

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
| Mengajukan pertanyaan (privat / publik) | ✅ Ya* | *Tergantung konfigurasi benefit `ask_qna` di plan |
| Melihat riwayat pertanyaan sendiri | ✅ Ya | Termasuk pertanyaan privat maupun publik milik sendiri |
| Melihat forum pertanyaan publik | ✅ Ya | Hanya pertanyaan publik berstatus `approved` ke atas |
| Menjawab pertanyaan publik orang lain | ✅ Ya** | **Tergantung konfigurasi benefit `answer_qna` di plan |
| Memberi upvote pada jawaban | ✅ Ya | Satu upvote per jawaban per member, bisa dibatalkan |
| Menandai jawaban sebagai rujukan valid | ❌ Tidak | Hanya assigned expert / admin / superadmin |
| Melihat pertanyaan **privat** orang lain | ❌ Tidak | Pertanyaan privat hanya antara penanya dan admin/expert |

> **Catatan benefit `ask_qna`:** Superadmin bisa mengatur apakah fitur mengajukan pertanyaan hanya tersedia untuk Pro, atau untuk semua member termasuk Standard. Jika tidak memiliki benefit ini, member melihat pesan ajakan upgrade.
>
> **Catatan benefit `answer_qna`:** Mekanisme sama — superadmin menentukan plan mana yang berhak menjawab pertanyaan publik member lain. Tanpa benefit ini, member tetap bisa membaca forum tapi tidak bisa menjawab.

### Guest (belum login)

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Membaca FAQ | ❌ Tidak | Harus login terlebih dahulu |
| Mengajukan pertanyaan | ❌ Tidak | Harus login |

### Expert / Penanggung Jawab (di-assign)

Expert adalah user (bisa member, staf, atau admin) yang **ditugaskan oleh admin** untuk menangani pertanyaan tertentu. Hak khusus expert berlaku **hanya untuk pertanyaan yang ditugaskan kepadanya**, dan ditentukan lewat permission `answer_qna` (menjawab) serta `validate_qna_answer` (menandai rujukan valid).

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Menjawab pertanyaan yang di-assign | ✅ Ya | Lewat mobile maupun backoffice |
| Menandai jawaban sebagai rujukan valid | ✅ Ya | Termasuk jawaban member lain di pertanyaan tsb |
| Approve / tolak jawaban member (manual mode) | ✅ Ya | Untuk pertanyaan yang ditugaskan kepadanya |
| Melihat isi pertanyaan privat yang di-assign | ✅ Ya | Hanya yang ditugaskan kepadanya |

### Superadmin

| Aksi | Bisa? | Catatan |
|------|-------|---------|
| Membuat & mengelola kategori FAQ | ✅ Ya | CRUD kategori, atur urutan, aktifkan/nonaktifkan |
| Membuat & mengelola FAQ | ✅ Ya | CRUD FAQ, pilih status draft/published/archived, pin item |
| Melihat semua pertanyaan masuk | ✅ Ya | Termasuk antrian pending dari semua member, public & private |
| Approve / tolak pertanyaan publik | ✅ Ya | Pertanyaan publik baru perlu disetujui agar tampil di forum |
| Menugaskan (assign) pertanyaan ke expert | ✅ Ya | Berlaku untuk pertanyaan publik maupun privat |
| Menjawab pertanyaan member | ✅ Ya | Jawaban bisa diset publik atau privat |
| Menandai jawaban sebagai rujukan valid | ✅ Ya | Bisa menandai jawaban siapa pun (member/expert/admin) |
| Approve / tolak / sembunyikan jawaban member | ✅ Ya | Sesuai mode moderasi jawaban |
| Menolak pertanyaan | ✅ Ya | Dengan menyertakan alasan yang dikirim ke member |
| Mengonversi pertanyaan menjadi FAQ | ✅ Ya | Jawaban yang valid bisa langsung dijadikan FAQ publik |
| Mengatur konfigurasi bot fallback & mode moderasi | ✅ Ya | Termasuk memilih mode `auto` / `manual` untuk jawaban |

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
  Isi form: kategori (opsional) + teks pertanyaan + pilih visibilitas (privat / publik)
          ↓
  Submit → validasi konten (word filter) dijalankan backend via System Settings
          ↓
  Bot fallback merespons otomatis (jika aktif)
          ↓
  Status: Menunggu (pending) — masuk antrian moderasi admin
          ↓
  Admin review antrian
       ↙                          ↘
  Setujui                       Tolak + alasan
     ↓                              ↓
  PRIVAT → status answered      Status: Ditolak
  saat dijawab admin/expert        ↓
  PUBLIK → status approved,     Member dapat notifikasi + alasan
  tampil di forum komunitas
     ↓
  Member lain (punya benefit answer_qna) bisa menjawab
     ↓
  Expert/admin menandai jawaban valid + member upvote
```

> **Catatan word filter:** Penyaringan kata terlarang **tidak ditangani di modul Q&A**. Validasi konten dijalankan di service layer backend memakai daftar kata yang dikelola di modul **System Settings**. Modul Q&A hanya memanggil validator tersebut sebelum menyimpan pertanyaan/jawaban; jika konten ditolak, request gagal dengan error validasi.

### Visibilitas Pertanyaan

Saat mengajukan, member memilih salah satu:

- **Privat** *(default)* — Pertanyaan hanya terlihat oleh penanya dan admin/expert yang ditugaskan. Cocok untuk hal personal/sensitif. Tidak masuk forum, tidak bisa dijawab member lain.
- **Publik** — Setelah disetujui admin, pertanyaan tampil di forum komunitas. Member lain bisa membaca dan (jika punya benefit `answer_qna`) menjawab.

### Bot Fallback

Saat pertanyaan baru masuk, sistem secara otomatis mengirim pesan konfirmasi kepada member bahwa pertanyaannya diterima dan sedang diproses. Pesan ini dikonfigurasi oleh superadmin — bukan jawaban sungguhan, melainkan acknowledgement otomatis. Bot juga bisa dikonfigurasi untuk memberikan saran kontak admin.

---

## Bagaimana Pertanyaan Publik Dijawab Komunitas?

Setelah pertanyaan publik disetujui dan tampil di forum:

1. **Member lain menjawab** — Member dengan benefit `answer_qna` bisa menulis jawaban. Satu member hanya boleh menulis satu jawaban per pertanyaan (bisa diedit). Expert yang ditugaskan dan admin juga bisa menjawab.

2. **Moderasi jawaban (konfigurable)** — Superadmin memilih mode lewat pengaturan:
   - **Auto** — Jawaban lolos word filter langsung tampil di forum.
   - **Manual** — Jawaban masuk antrian `pending` dulu; admin atau expert penanggung jawab harus menyetujui sebelum tampil.

3. **Upvote** — Member bisa memberi upvote pada jawaban yang membantu (satu upvote per jawaban, bisa dibatalkan). Jawaban diurutkan: rujukan valid di paling atas, lalu upvote terbanyak.

4. **Rujukan valid (accepted answer)** — Assigned expert, admin, atau superadmin bisa menandai satu atau **beberapa** jawaban sebagai rujukan valid. Jawaban yang ditandai akan diberi badge dan diposisikan di atas. Penanya **tidak** bisa menandai sendiri — penandaan ini wewenang authority agar kualitas terjaga.

5. **Konversi ke FAQ** — Admin bisa mengonversi jawaban valid mana pun menjadi FAQ publik baru.

---

## Penugasan Pertanyaan (Assignment)

Admin bisa menugaskan pertanyaan — publik maupun privat — kepada seorang expert atau staf sebagai penanggung jawab. Tujuannya merutekan pertanyaan ke orang yang paling kompeten (mis. pertanyaan pajak ke expert pajak).

- Expert yang ditugaskan mendapat notifikasi dan melihat pertanyaan tersebut di antrian "Ditugaskan ke Saya".
- Untuk pertanyaan **privat**, hanya expert yang ditugaskan (dan admin) yang bisa melihat isinya.
- Penugasan bisa diubah atau dicabut oleh admin kapan saja.
- Assignment bersifat penanggung jawab — tidak menghalangi member lain menjawab pertanyaan **publik**.

---

## Status Pertanyaan Member dan Artinya

### Menunggu *(pending)*
Pertanyaan sudah dikirim dan masuk antrian moderasi admin. Belum tampil di forum (untuk publik) dan belum dijawab (untuk privat).

### Disetujui & Tayang *(approved)*
Khusus pertanyaan **publik**. Admin sudah menyetujui dan pertanyaan kini tampil di forum komunitas, siap menerima jawaban dari member lain.

### Terjawab *(answered)*
Terutama dipakai untuk alur **privat** — admin/expert sudah memberikan jawaban. Member mendapat notifikasi push dan bisa membaca jawaban di halaman detail pertanyaannya.

### Ditolak *(rejected)*
Admin menolak pertanyaan — biasanya karena sudah ada di FAQ, duplikat, di luar scope, atau melanggar aturan. Member mendapat notifikasi beserta alasan penolakan.

### Dikonversi ke FAQ *(converted)*
Admin menilai pertanyaan ini cukup umum. Jawaban valid dijadikan FAQ publik yang bisa diakses semua orang. Penanya tetap mendapat notifikasi.

### Ditutup *(closed)*
Pertanyaan publik ditutup oleh admin — tidak menerima jawaban baru lagi, namun isi dan jawaban yang ada tetap bisa dibaca.

---

## Alur Lengkap Status Pertanyaan

```
Member submit pertanyaan (pilih privat / publik)
          ↓
       PENDING ───────── Tolak ──────→ REJECTED
          │
     Approve admin
       ↙        ↘
  PRIVAT       PUBLIK
     │            │
  dijawab     APPROVED (tampil di forum)
  admin/expert    │
     ↓         member lain menjawab + upvote
 ANSWERED      expert/admin tandai jawaban valid
     │            │
     │       (admin bisa CLOSED / CONVERTED)
     ↓            ↓
 CONVERTED ←──────┘
 (jika jawaban valid dijadikan FAQ)
```

---

## Aturan Privasi Pertanyaan

Visibilitas pertanyaan **dipilih oleh penanya** saat mengajukan:

- **Privat** *(default)* — Hanya penanya, admin, dan expert yang ditugaskan yang bisa melihat. Member lain tidak bisa melihat maupun menjawab. Cocok untuk pertanyaan personal atau sensitif.
- **Publik** — Setelah disetujui admin, pertanyaan dan jawaban-jawabannya terlihat oleh semua member di forum. Identitas penanya dan penjawab tampil (nama). Member lain bisa menjawab dan memberi upvote.

Untuk jawaban pada **alur privat**, admin tetap bisa memilih `is_public` saat menjawab — namun karena pertanyaannya privat, jawaban hanya terlihat oleh penanya. Flag `is_public` di sini relevan ketika sebuah jawaban privat hendak dikonversi menjadi FAQ.

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

## Sistem Upvote Jawaban (Forum Publik)

Berbeda dari vote helpful pada FAQ, jawaban di pertanyaan publik punya mekanisme **upvote** (hanya naik, tidak ada downvote):

- Satu member hanya bisa memberi satu upvote per jawaban, dan bisa membatalkannya.
- Member tidak bisa upvote jawabannya sendiri.
- Urutan jawaban di forum: **rujukan valid** (accepted) di paling atas, lalu **upvote terbanyak**, lalu yang terbaru.
- Upvote membantu jawaban berkualitas naik secara organik sebelum/selain ditandai valid oleh expert.

---

## Notifikasi Q&A

Member menerima notifikasi push dalam beberapa situasi:

1. **Pertanyaan disetujui (publik)** — Saat pertanyaan publik di-approve admin dan tayang di forum.
2. **Pertanyaan dijawab** — Saat admin/expert/member lain memberikan jawaban atas pertanyaan member.
3. **Pertanyaan ditolak** — Saat admin menolak pertanyaan, beserta alasan penolakan.
4. **Jawaban ditandai valid** — Saat jawaban member ditandai sebagai rujukan valid oleh expert/admin.
5. **Pertanyaan ditugaskan** — Expert mendapat notifikasi saat sebuah pertanyaan ditugaskan kepadanya.

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

### Use Case 9 — Member Bertanya Publik dan Dijawab Komunitas

**Siapa:** Choi Woo (penanya, Standard) & Kang Dae (penjawab, Pro dengan benefit `answer_qna`)

Choi Woo mengajukan pertanyaan "Rekomendasi co-working space murah di Jakarta Selatan?" dan memilih visibilitas **publik**. Admin menyetujui, pertanyaan tayang di forum.

Kang Dae yang sudah lama tinggal di Jakarta menulis jawaban lengkap berisi tiga rekomendasi. Karena mode moderasi diset `auto`, jawabannya langsung tampil. Beberapa member lain memberi upvote sehingga jawaban Kang Dae naik ke atas. Admin kemudian menandainya sebagai **rujukan valid**, dan jawaban itu mendapat badge.

---

### Use Case 10 — Admin Menugaskan Pertanyaan ke Expert Pajak

**Siapa:** Admin KAI & Expert Pajak (staf yang diberi permission `answer_qna` + `validate_qna_answer`)

Masuk pertanyaan teknis soal pelaporan SPT tahunan WNA. Admin merasa ini perlu jawaban akurat, jadi ia **menugaskan** pertanyaan tersebut ke Expert Pajak. Expert mendapat notifikasi, membuka antrian "Ditugaskan ke Saya" lewat mobile, lalu menulis jawaban resmi dan menandainya sebagai rujukan valid. Penanya mendapat notifikasi bahwa pertanyaannya dijawab.

---

### Use Case 11 — Mode Moderasi Manual untuk Jawaban Sensitif

**Siapa:** Admin KAI

Karena beberapa kategori rawan misinformasi, admin mengubah `answer_moderation_mode` menjadi **manual**. Kini setiap jawaban member di forum masuk status `pending` lebih dulu. Admin atau expert penanggung jawab meninjau, lalu approve jawaban yang akurat dan menolak yang menyesatkan — sebelum jawaban terlihat oleh member lain.

---

## Ringkasan Aturan

| Aturan | Detail |
|--------|--------|
| **Siapa yang bisa baca FAQ** | Semua member yang sudah login |
| **Siapa yang bisa ajukan pertanyaan** | Member dengan benefit `ask_qna` |
| **Siapa yang bisa jawab pertanyaan publik** | Member dengan benefit `answer_qna`, expert yang di-assign, admin |
| **Siapa yang bisa tandai jawaban valid** | Assigned expert / admin / superadmin (permission `validate_qna_answer`) — **bukan** penanya |
| **Siapa yang bisa buat/kelola FAQ** | Superadmin saja |
| **Visibilitas pertanyaan** | Dipilih penanya: `private` (default) atau `public` |
| **Privasi pertanyaan privat** | Hanya penanya + admin + expert yang ditugaskan |
| **Moderasi pertanyaan** | Selalu `pending` → di-approve admin (publik baru tayang setelah approve) |
| **Moderasi jawaban** | Konfigurable: `auto` (langsung tayang) atau `manual` (approve dulu) |
| **Word filter** | Ditangani backend via modul System Settings, bukan modul Q&A |
| **Assignment** | Admin assign pertanyaan (publik & privat) ke expert/staf |
| **Accepted answer** | Boleh **beberapa** per pertanyaan, ditandai authority |
| **Upvote jawaban** | Satu upvote per jawaban per member, bisa dibatalkan, tidak bisa upvote sendiri |
| **Status FAQ** | `draft` → `published` → `archived` |
| **Status pertanyaan** | `pending` → `approved` / `answered` / `rejected` / `converted` / `closed` |
| **Status jawaban** | `visible` / `pending` / `rejected` / `hidden` |
| **Bot fallback** | Pesan otomatis saat pertanyaan masuk, bisa diaktifkan/dinonaktifkan |
| **Vote helpful FAQ** | Satu vote per member per FAQ, bisa diubah kapan saja |
| **Konversi ke FAQ** | Admin bisa konversi jawaban valid menjadi FAQ publik |
| **Notifikasi** | Approve, dijawab, ditolak, ditandai valid, ditugaskan |
| **Hapus pertanyaan** | Member tidak bisa menghapus pertanyaan yang sudah diajukan |

---

*Dokumen rules Q&A / FAQ module — non-teknis. Untuk detail API lihat API_SPEC_QNA_MOBILE dan API_SPEC_QNA_BACKOFFICE. Untuk skema database lihat QNA_DB_SCHEMA.*
