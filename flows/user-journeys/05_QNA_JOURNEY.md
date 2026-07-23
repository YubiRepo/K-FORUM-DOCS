# Q&A / FAQ — User Journey (Member Standard & Pro)

> Legend platform & tier: lihat [00_OVERVIEW.md](00_OVERVIEW.md). Sumber: `Modules/Q&A/QNA_RULES.md`.

## Ringkasan Domain

Modul Q&A / FAQ punya dua jenis konten. **FAQ** adalah konten resmi yang dikurasi dan dipublish superadmin — semua member cuma bisa membaca, mencari, dan memberi vote helpful/not helpful, tidak bisa membuat atau mengedit. **Pertanyaan Member** adalah pertanyaan yang diajukan member sendiri saat tidak menemukan jawaban di FAQ, dengan dua pilihan visibilitas: **privat** (1-on-1 dengan admin/expert yang ditugaskan, tidak masuk forum) atau **publik** (setelah disetujui admin, tayang di forum komunitas dan bisa dijawab member lain, di-upvote, dan ditandai sebagai rujukan valid oleh expert/admin).

**Nuansa penting di domain ini** (beda dari domain lain yang biasanya Pro-only): mengajukan pertanyaan (`ask_qna`) dan menjawab pertanyaan publik (`answer_qna`) **bukan hardcode Pro-only**. Kedua benefit ini **dikonfigurasi Superadmin per plan** — Superadmin bisa memilih apakah Standard juga mendapat benefit ini atau tidak. Jangan asumsikan Standard selalu terkunci dari fitur ini; dokumen ini menandai baris terkait sebagai "tergantung konfigurasi", bukan "Pro only" secara mutlak. Referensi tabel perbandingan tier lengkap ada di [00_OVERVIEW.md](00_OVERVIEW.md).

## Batasan Standard vs Pro di Domain Ini

| Aksi | Standard | Pro | Catatan |
|---|:---:|:---:|---|
| Membaca & mencari FAQ | ✅ | ✅ | Semua member login, tanpa gate benefit |
| Vote helpful / not helpful pada FAQ | ✅ | ✅ | Satu vote per FAQ per member, bisa diubah kapan saja |
| Melihat forum pertanyaan publik | ✅ | ✅ | Hanya pertanyaan publik berstatus `approved` ke atas |
| Mengajukan pertanyaan (privat/publik) | ✅* | ✅* | *Tergantung konfigurasi benefit `ask_qna` oleh Superadmin — default bisa berbeda per plan, tidak otomatis Pro-only |
| Melihat riwayat pertanyaan sendiri | ✅ | ✅ | Termasuk privat maupun publik milik sendiri |
| Menjawab pertanyaan publik member lain | ✅** | ✅** | **Tergantung konfigurasi benefit `answer_qna` oleh Superadmin — default bisa berbeda per plan, tidak otomatis Pro-only |
| Upvote jawaban di forum publik | ✅ | ✅ | Satu upvote per jawaban per member, tidak bisa upvote jawaban sendiri |
| Menandai jawaban sebagai rujukan valid | ❌ | ❌ | Hanya assigned expert / admin / superadmin — penanya pun tidak bisa menandai jawabannya sendiri |
| Melihat pertanyaan privat member lain | ❌ | ❌ | Privat hanya terlihat penanya + admin/expert yang ditugaskan |
| Membuat/mengelola FAQ | ❌ | ❌ | Hanya Superadmin |

## Journey 1: Membaca & Mencari FAQ — 🅢🅟 — 📱 Mobile

1. Member membuka modul Q&A dari menu utama aplikasi.
2. Melihat daftar kategori FAQ (mis. Imigrasi & Visa, Pajak, Ketenagakerjaan, Pendirian Usaha, Properti & Kepemilikan, Sistem Aplikasi KAI, Umum) — kategori dan urutannya dikelola Superadmin.
3. Memilih kategori, atau langsung mengetik kata kunci di kolom pencarian (full-text search di seluruh FAQ published).
4. FAQ yang di-pin oleh Superadmin selalu tampil di posisi teratas kategorinya, terlepas dari urutan normal.
5. Membuka detail FAQ, membaca isi jawaban.
6. Mengetuk "Helpful" atau "Not Helpful" — satu vote per FAQ per member, bisa diubah kapan saja (dari helpful ke not helpful atau sebaliknya). Vote **tidak** memengaruhi urutan tampil FAQ (itu murni `sort_order` + `is_pinned` yang diatur Superadmin).
7. Jika jawaban yang dicari ditemukan, selesai — member tidak perlu mengajukan pertanyaan baru.

**Selesai:** Member menemukan informasi yang dibutuhkan tanpa perlu berinteraksi dengan admin/expert.

## Journey 2: Mengajukan Pertanyaan Privat — 🅢🅟 — 📱 Mobile

1. Member sudah membaca & mencari FAQ tapi tidak menemukan jawaban yang sesuai.
2. Mengetuk "Ajukan Pertanyaan". Jika member **tidak** punya benefit `ask_qna` (sesuai konfigurasi Superadmin untuk plan-nya), muncul modal ajakan upgrade — member bisa pilih upgrade atau menutup modal dan kembali membaca FAQ saja.
3. Mengisi form: kategori (opsional), teks pertanyaan, pilih visibilitas **Privat** (ini pilihan default).
4. Submit → konten divalidasi word filter di backend (dikelola modul System Settings, bukan Q&A).
5. Sistem otomatis mengirim pesan bot fallback ("Terima kasih atas pertanyaan Anda, tim kami akan segera merespons...") — ini acknowledgement otomatis, bukan jawaban sungguhan.
6. Status pertanyaan: **Menunggu (pending)** — masuk antrian moderasi admin. Karena privat, tidak ada proses approve/tolak untuk "tayang" (privat memang tidak pernah tayang ke forum) — admin/expert langsung bisa melihat dan menjawabnya di antrian mereka.
7. Admin bisa menugaskan (assign) pertanyaan ini ke expert tertentu sebagai penanggung jawab; hanya expert yang ditugaskan (dan admin) yang bisa melihat isi pertanyaan privat ini.
8. Admin/expert menjawab → status berubah jadi **Terjawab (answered)**. Member menerima notifikasi push dan membaca jawaban lengkap di halaman detail pertanyaannya.
9. Member bisa melihat riwayat semua pertanyaan yang pernah diajukan (privat maupun publik) lewat tab "Riwayat Pertanyaan", lengkap dengan status masing-masing (menunggu / terjawab / ditolak / dst).

**Selesai:** Member mendapat jawaban personal atas pertanyaannya tanpa pertanyaan tersebut terekspos ke member lain.

**Catatan:** Admin juga bisa menolak pertanyaan (privat maupun publik) dengan menyertakan alasan — status jadi **Ditolak (rejected)**, member mendapat notifikasi beserta alasannya (mis. karena sudah ada di FAQ, duplikat, atau di luar scope).

## Journey 3: Mengajukan & Berinteraksi di Pertanyaan Publik — 🅢🅟 — 📱 Mobile

1. Member mengajukan pertanyaan (langkah sama seperti Journey 2), tapi memilih visibilitas **Publik** di form. Gate benefit `ask_qna` berlaku sama seperti pertanyaan privat.
2. Status **Menunggu (pending)** — pertanyaan publik **wajib** disetujui admin dulu sebelum tampil di forum (berbeda dari privat yang langsung bisa dilihat admin/expert untuk dijawab).
3. Admin approve → status **Disetujui & Tayang (approved)**, pertanyaan tampil di forum komunitas, member penanya dapat notifikasi.
4. Member lain yang punya benefit `answer_qna` (sesuai konfigurasi Superadmin) bisa membaca dan menulis jawaban. Satu member hanya boleh menulis satu jawaban per pertanyaan (bisa diedit). Member yang **tidak** punya benefit ini tetap bisa membaca forum, hanya tidak bisa menjawab.
5. Jawaban tayang tergantung mode moderasi yang diatur Superadmin:
   - **Auto** — jawaban yang lolos word filter langsung tampil di forum.
   - **Manual** — jawaban masuk antrian `pending`, harus disetujui admin/expert penanggung jawab dulu sebelum terlihat member lain.
6. Member (termasuk penanya) bisa memberi **upvote** pada jawaban yang membantu — satu upvote per jawaban per member, bisa dibatalkan, dan **tidak bisa upvote jawaban sendiri**. Jawaban di forum diurutkan: rujukan valid dulu, lalu upvote terbanyak, lalu yang terbaru.
7. Expert yang ditugaskan atau admin/superadmin menandai satu atau beberapa jawaban sebagai **rujukan valid** (accepted answer) — jawaban itu dapat badge dan naik ke posisi atas. **Penanya sendiri tidak bisa menandai jawabannya sebagai valid** — ini murni wewenang authority agar kualitas terjaga.
8. Member penanya mendapat notifikasi setiap kali ada perkembangan: disetujui, dijawab, jawaban ditandai valid.
9. Pertanyaan publik bisa ditutup (**closed**) oleh admin — tidak menerima jawaban baru, tapi isi & jawaban lama tetap bisa dibaca. Bisa juga dikonversi jadi FAQ publik (**converted**) jika admin menilai jawabannya cukup umum untuk semua member.

**Selesai:** Pertanyaan publik terjawab lewat kontribusi komunitas, jawaban terbaik naik secara organik lewat upvote dan ditandai resmi oleh authority.

## Keterlibatan Admin — 💻 Web/Backoffice

**Superadmin — Moderasi pertanyaan & jawaban:**
- Melihat semua pertanyaan masuk (pending queue), publik maupun privat, dari semua member.
- Approve/tolak pertanyaan publik (wajib sebelum tayang di forum); tolak disertai alasan yang dikirim ke member.
- Approve/tolak/sembunyikan jawaban member sesuai mode moderasi (`auto`/`manual`).
- Mengatur mode moderasi jawaban (`answer_moderation_mode`) dan konfigurasi bot fallback.
- Menandai jawaban siapa pun (member/expert/admin) sebagai rujukan valid.
- Mengonversi jawaban valid menjadi FAQ publik baru.

**Superadmin — Assignment:**
- Menugaskan (assign) pertanyaan — publik maupun privat — ke expert atau staf sebagai penanggung jawab, untuk merutekan ke orang paling kompeten (mis. pertanyaan pajak ke expert pajak).
- Mengubah atau mencabut penugasan kapan saja. Assignment bersifat penanggung jawab, tidak menghalangi member lain menjawab pertanyaan publik.

**Superadmin — Kelola FAQ:**
- CRUD kategori FAQ (buat, ubah, aktifkan/nonaktifkan, hapus — kategori dengan FAQ aktif tidak bisa dihapus langsung, harus diarsipkan/dipindah dulu).
- CRUD FAQ, atur status `draft` → `published` → `archived` (bisa kembali ke draft), atur urutan (`sort_order`), tandai `pinned`.
- Melihat statistik vote helpful sebagai bahan evaluasi kualitas konten.

**Expert / Penanggung Jawab (di-assign, permission `answer_qna` + `validate_qna_answer`):**
- Menjawab pertanyaan yang ditugaskan kepadanya (lewat mobile maupun backoffice).
- Menandai jawaban (termasuk jawaban member lain) sebagai rujukan valid, khusus untuk pertanyaan yang di-assign kepadanya.
- Approve/tolak jawaban member dalam mode moderasi manual, untuk pertanyaan yang ditugaskan kepadanya.
- Melihat isi pertanyaan privat yang di-assign kepadanya.

## Di Luar Cakupan Standard & Pro

- **Membuat, mengedit, atau menghapus FAQ** — FAQ murni konten kurasi Superadmin; pertanyaan member hanya bisa "dikonversi" jadi FAQ oleh Superadmin, tidak bisa diajukan langsung sebagai FAQ.
- **Membuat/mengelola kategori FAQ** — kewenangan Superadmin saja.
- **Menandai jawaban sebagai rujukan valid** (termasuk jawaban sendiri) — hanya assigned expert/admin/superadmin, agar penandaan kualitas tetap objektif.
- **Melihat pertanyaan privat member lain** — hanya penanya dan admin/expert yang ditugaskan yang bisa melihat isinya.
- **Menugaskan (assign) pertanyaan ke expert** — murni keputusan admin.
- **Menghapus pertanyaan yang sudah diajukan** — member tidak punya kemampuan ini sama sekali.
- **Mengatur mode moderasi jawaban (auto/manual) atau bot fallback** — konfigurasi sistem oleh Superadmin.

## Edge Case & Catatan Tambahan

- **Word filter bukan bagian modul Q&A** — validasi konten terlarang (untuk pertanyaan maupun jawaban) dijalankan di service layer backend memakai daftar kata dari modul System Settings. Modul Q&A hanya memanggil validator ini sebelum menyimpan; jika ditolak, request gagal dengan error validasi.
- **`is_public` pada jawaban di alur privat** — admin tetap bisa menandai jawaban sebagai `is_public` saat menjawab pertanyaan privat, tapi karena pertanyaannya privat, jawaban tetap hanya terlihat penanya. Flag ini baru relevan kalau jawaban privat tersebut nantinya dikonversi jadi FAQ.
- **Accepted answer bisa lebih dari satu** per pertanyaan publik — bukan single "best answer" seperti forum Q&A pada umumnya.
- **Assignment tidak eksklusif untuk pertanyaan publik** — expert yang ditugaskan tetap harus bersaing/berdampingan dengan jawaban member lain yang punya benefit `answer_qna`; assignment hanya menandai siapa yang "bertanggung jawab", bukan menutup pertanyaan dari kontribusi member lain.
- **Guest (belum login) sama sekali tidak bisa mengakses modul ini** — baca FAQ maupun mengajukan pertanyaan sama-sama mensyaratkan login terlebih dahulu.


## Note
- ~~saat pertanyaan private mengandung attachment, dari backoffice admin tidak bisa melihat attachment tersebut, hanya bisa melihat pertanyaan dan jawaban saja. Attachment hanya bisa diakses oleh penanya dan expert/admin yang ditugaskan.~~ **Sudah diimplementasikan** (22 Jul 2026): UI attachment ditambahkan di modal detail pertanyaan backoffice (`app/pages/qa/index.vue`) — pertanyaan, jawaban privat legacy, dan tiap jawaban forum publik. Sekalian ditemukan & diperbaiki 2 bug bonus di baris yang sama: jawaban privat sebelumnya tampil sebagai raw JSON blob (tipe TS salah, seharusnya object bukan string), dan nama penjawab di forum selalu kosong (field flat API dikira nested object). Belum dicek visual langsung di browser (baru sampai `npm run build` sukses). Migrasi attachment dari `PrivacyPublic` ke `PrivacyPrivate` di k-forum-api **sengaja ditunda** (blast radius lebar, nilainya baru maksimal setelah UI ini jadi — sudah tercapai, jadi migrasi ini jadi kandidat kuat buat sesi berikutnya). Detail: `QNA_MODULE_ISSUES.md` Issue 1.
- ~~admin failed to assign pertanyaan private ke expert, route not found.~~ **Sudah diimplementasikan** (22 Jul 2026, k-forum-api): route salah verb (`POST` seharusnya `PUT` sesuai spec B18). Audit menyeluruh endpoint QnA menemukan 3 mismatch verb tambahan yang sama-sama 404 diam-diam (Update Category, Reject Question, Moderate Answer) — semua sudah diperbaiki sekalian. Detail: `QNA_MODULE_ISSUES.md` Issue 2.
- ~~saat liat pertanyaan public di mobile , endpoint error answer query failed.~~ **Sudah diimplementasikan** (22 Jul 2026, k-forum-api): query jawaban di `GetPublicQuestionDetail` memakai parameter yang sama dalam 2 konteks tipe berbeda (text vs uuid) dalam satu SQL, bikin Postgres gagal resolve tipe — diperbaiki jadi WHERE clause kondisional di Go, mengikuti pola yang sudah benar di `ListPublicQuestions`. Detail: `QNA_MODULE_ISSUES.md` Issue 3.
- lifecycle pertanyaan public/private masih belum jelas. *(belum diinvestigasi — di luar cakupan `QNA_MODULE_ISSUES.md`, belum ada analisis kode untuk poin ini)*
