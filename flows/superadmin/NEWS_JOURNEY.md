# News — Journey Superadmin (Backoffice)

> Dokumen ini menggambarkan apa yang **benar-benar bisa dilakukan Superadmin hari ini** di panel admin (backoffice) untuk modul News, berdasarkan tampilan dan tombol yang sudah ada di aplikasi — bukan dokumen rencana/spec. Untuk journey member (yang baca/nulis berita di HP), lihat [03_NEWS_JOURNEY.md](../user-journeys/03_NEWS_JOURNEY.md).

## Ringkasan

News di backoffice adalah tempat mengelola semua hal soal berita: menerbitkan artikel, mengatur kategori & wilayah berita, mengatur sumber berita otomatis (scraping), dan mengatur terjemahan. Superadmin punya akses ke hampir semua bagian ini. Sebagian kecil pengaturan teknis (soal cara mesin scraping bekerja) dikunci khusus untuk **Usergod** (peran developer/vendor) — Superadmin tidak bisa membukanya.

Cara masuk: dari menu sebelah kiri, pilih **Content Management → News**. Halaman News punya 5 tab di bagian atas: **Articles, Sources, Categories, Scopes, Settings**.

## Journey 1: Mengelola Artikel (tab Articles)

Ini halaman utama, isinya daftar semua artikel — baik yang ditulis admin, hasil scraping otomatis, maupun kiriman Member Pro — semua tercampur dalam satu tabel.

1. Superadmin membuka tab **Articles**. Di atas tabel ada kartu ringkasan: Total, Published, Pending Approval, Draft.
2. **Menyaring daftar**: bisa pilih status (Semua/Published/Draft/Pending/Rejected/Archived), cari judul/penulis, filter kategori, filter wilayah berita, atau filter berdasarkan sumber (manual atau salah satu sumber otomatis).
3. **Menerbitkan draft**: artikel berstatus Draft (baik tulisan admin sendiri atau hasil scraping) bisa langsung di-**Publish** dari menu titik-tiga.
4. **Menandai unggulan**: artikel yang sudah Published bisa ditandai **Feature/Unfeature** (bintang) supaya tampil menonjol di aplikasi member.
5. **Mengarsipkan**: artikel Published bisa di-**Archive** kalau sudah tidak relevan lagi.
6. **Menghapus**: hanya artikel berstatus Draft atau Rejected yang bisa dihapus permanen (ada konfirmasi sebelum terhapus).
7. **Aksi massal**: centang beberapa artikel sekaligus untuk **Publish** banyak draft sekaligus atau **Delete** banyak sekaligus.
8. **Membuat artikel baru**: tombol **New Article** → isi judul, ringkasan, isi berita, pilih kategori (wajib), wilayah berita (opsional), bahasa utama, label asal (mis. "KAI Jakarta"), nama penulis, unggah gambar sampul, tag, lalu simpan sebagai **Draft** atau langsung **Publish**.
9. **Mengedit artikel**: buka artikel → **Edit**. Semua field bisa diubah kecuali bahasa utama (terkunci setelah artikel dibuat, supaya tidak membingungkan sistem terjemahan).
10. **Melihat detail artikel**: ada halaman detail berisi statistik (jumlah dilihat, disukai, dikomentari), info asal artikel, alasan penolakan (kalau pernah ditolak), dan tombol Publish/Archive/Edit.

**Yang belum ada**: tidak ada halaman khusus untuk moderasi komentar berita (hapus komentar tidak pantas dari pembaca) di dalam menu News — itu ditangani lewat menu **Reports** yang sifatnya umum untuk semua jenis konten, bukan khusus News.

## Journey 2: Menerjemahkan Artikel

Fitur ini ada di dalam halaman **Edit Artikel**, tab **Translations** — dipakai untuk menyiapkan versi bahasa lain dari satu artikel yang sama.

1. Superadmin membuka artikel yang mau diterjemahkan → **Edit** → pindah ke tab **Translations**.
2. Muncul daftar semua bahasa aktif di sistem (selain bahasa asli artikel), masing-masing dengan status saat ini:
   - **Original** — ini bahasa asli artikel, tidak perlu diterjemahkan.
   - **Not Translated** — belum ada versi bahasa ini sama sekali.
   - **Pending / Processing (AI)** — baru saja diminta diterjemahkan otomatis, sedang diproses AI di belakang layar.
   - **AI Failed** — proses terjemahan otomatis gagal (perlu dicoba ulang atau ditulis manual).
   - **AI (nama provider)** — sudah selesai diterjemahkan otomatis, menyebutkan AI apa yang dipakai (mis. Claude/Google).
   - **Manual** — versi bahasa ini ditulis/disunting langsung oleh admin, bukan hasil AI.
3. **Menerjemahkan otomatis**: klik tombol **Auto-Translate** pada baris bahasa yang diinginkan. Status baris itu akan berubah bertahap dari Not Translated → Pending → Processing → selesai jadi "AI (provider)" (atau "AI Failed" kalau gagal). Ini berjalan di belakang layar, jadi Superadmin bisa klik **Reload** beberapa saat kemudian untuk melihat status terbaru tanpa perlu keluar-masuk halaman.
4. **Menerjemahkan/menyunting manual**: klik **Edit Manual** untuk membuka form berisi Judul, Ringkasan, Isi, dan Tag dalam bahasa tersebut — dipakai kalau hasil AI kurang tepat, atau memang ingin ditulis sendiri tanpa AI sama sekali. Setelah disimpan, status baris berubah jadi **Manual**.
5. Kedua cara (AI maupun manual) bisa dipakai bergantian per bahasa — misalnya bahasa Inggris dipakai hasil AI, tapi bahasa Korea ditulis manual karena butuh istilah khusus.
6. **Keterbatasan saat ini**: tombol **Auto-Translate** di layar ini otomatis memakai satu penyedia AI tertentu secara tetap — belum ada pilihan "pakai provider yang mana" langsung di sini per artikel. Pengaturan provider (utama & cadangan) hanya bisa diatur secara global untuk seluruh modul News lewat tab **Settings** (lihat Journey 6), bukan per artikel/per klik.
7. Kalau terjemahan global dimatikan (lihat Journey 6), tab Translations ini tetap ada tapi hasilnya tidak akan pernah dibaca member — pembaca akan selalu melihat bahasa asli.

## Journey 3: Mengelola Kategori Berita (tab Categories)

1. Superadmin membuka tab **Categories** — daftar master kategori (Politik, Ekonomi, Olahraga, dll) yang dipakai untuk mengelompokkan semua artikel.
2. Bisa lihat dalam bentuk kartu atau tabel, cari berdasarkan nama.
3. **Menambah kategori**: tombol **New Category** → isi nama (kode/slug terisi otomatis mengikuti nama, bisa diubah manual), atur aktif/nonaktif.
4. **Mengubah urutan tampil**: aktifkan mode **Reorder**, geser-geser kategori ke urutan yang diinginkan, lalu **Save Order**.
5. **Mengaktif/nonaktifkan, mengedit, atau menghapus** kategori langsung dari kartu/baris masing-masing.

## Journey 4: Mengelola Wilayah Berita / News Scope (tab Scopes)

Alurnya sama persis seperti Categories di atas, hanya isinya wilayah cakupan berita (misalnya "Berita Indonesia", "Berita Korea", "Berita Korea di Indonesia") — dipakai sebagai filter tambahan yang terpisah dari kategori.

## Journey 5: Mengelola Sumber Berita Otomatis / Scraping (tab Sources)

Tab ini untuk berita yang ditarik otomatis dari situs berita lain, tanpa perlu ditulis manual satu-satu.

1. **Active** — saklar untuk menyalakan/mematikan satu sumber berita. Kalau dimatikan, sumber itu berhenti ditarik otomatis sesuai jadwalnya (tapi artikel yang sudah pernah masuk tidak ikut terhapus).
2. **Auto Publish** — saklar untuk menentukan apakah artikel hasil tarikan otomatis dari sumber ini langsung tayang (Published) begitu selesai ditarik, atau masuk dulu sebagai **Draft** supaya bisa diperiksa manual sebelum diterbitkan.
3. **Auto Translate** — saklar untuk menentukan apakah artikel dari sumber ini otomatis diterjemahkan ke semua bahasa aktif begitu selesai ditarik, atau dibiarkan dalam bahasa asli sampai ada yang menerjemahkan (manual atau on-demand, lihat Journey 2 & 6).
4. **Scrape Now** — tombol untuk langsung menarik berita terbaru dari sumber tersebut saat itu juga, tanpa menunggu jadwal otomatis berikutnya (berguna untuk uji coba atau kalau ada berita penting yang perlu segera masuk).

Selain 4 hal di atas, Superadmin juga bisa:

- Melihat berapa banyak kategori scraping yang sudah dipetakan per sumber.
- Membuka tombol **Categories** untuk mengatur pemetaan tiap kategori dari sumber itu ke kategori berita tujuan di sistem (mis. kategori "Sports" di situs sumber dipetakan ke kategori "Olahraga" di News), termasuk mengaktif/nonaktifkan kategori tertentu supaya tidak ikut ditarik.

**Yang tetap eksklusif Usergod** (developer/vendor), tidak bisa disentuh Superadmin:

- Menambah sumber berita baru atau menghapus sumber yang sudah terdaftar.
- Mengatur jadwal mentah (kapan persisnya scraping berjalan otomatis) dan selector HTML (petunjuk teknis cara sistem "membaca" bagian judul/isi/tag dari halaman situs sumber).

## Journey 6: Mengatur Terjemahan Global (tab Settings)

1. Superadmin membuka tab **Settings** — ini pengaturan terjemahan yang berlaku untuk **seluruh** modul News, bukan per artikel.
2. **Saklar Aktifkan Terjemahan**: kalau dimatikan, semua pembaca hanya melihat bahasa asli artikel (paling hemat, tidak ada proses terjemahan sama sekali).
3. **Saklar Terjemahan Otomatis (On-Demand)**: kalau menyala, begitu ada pembaca membuka artikel dalam bahasa yang belum pernah diterjemahkan, sistem otomatis menerjemahkan di belakang layar untuk pembaca berikutnya.
4. **Pilih Penyedia Terjemahan**: ada pilihan penyedia utama (mis. AI Claude, Google, AWS, atau tanpa penerjemah sama sekali) dan penyedia cadangan (dipakai kalau penyedia utama sedang tidak tersedia). Pilihan yang belum terpasang di server akan tampil tapi tidak bisa dipilih.
5. Semua perubahan di halaman ini disimpan sekaligus lewat satu tombol **Save Settings** di bagian bawah (bukan tersimpan otomatis per kolom).

## Batasan & Catatan Penting untuk Superadmin

- **Yang eksklusif Usergod, bukan Superadmin**: menambah/menghapus sumber berita otomatis, serta mengatur jadwal mentah & selector HTML-nya. Setelah perbaikan backoffice selesai (lihat Journey 5), Superadmin akan bisa mengatur Active, Auto Publish, Auto Translate, Scrape Now, dan pemetaan kategori langsung dari sumber yang sudah terdaftar — tanpa perlu masuk ke halaman teknis Usergod.
- **Moderasi komentar berita** tidak punya halaman khusus di menu News — dilakukan lewat menu Reports yang umum untuk semua modul.
- **Kata terlarang (banned keywords)** untuk judul/isi artikel Member Pro diatur secara global di menu Settings → Moderation, berlaku sama untuk News dan lima modul lain — tidak ada pengaturan kata terlarang khusus News.
- **Tidak ada tombol "reindex pencarian"** di menu News — kalau ada masalah pencarian artikel tidak muncul, itu perlu ditangani lewat cara lain di luar tampilan backoffice ini.
- **Memilih penyedia terjemahan per-artikel**: opsi untuk memilih penyedia terjemahan secara spesifik saat menerjemahkan satu artikel tertentu, sebenarnya sudah disiapkan sebagai pengaturan ("Allow Per-Request Override") di halaman Settings, tapi tombol "Auto-Translate" di halaman edit artikel belum menyediakan pilihan itu — jadi pengaturannya ada, tapi belum bisa dipakai dari layar edit artikel.
