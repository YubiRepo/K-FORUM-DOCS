# Verification Document — Upload & Access Spec (Mobile & Backoffice) v1.0

Spec pelengkap `API_SPEC_VERIFICATION_MOBILE.md` & `API_SPEC_VERIFICATION_BACKOFFICE.md`. Dokumen ini fokus khusus ke **mekanisme file** dokumen verifikasi (KTA, NIB, KTP, dll): siapa yang boleh akses, kenapa URL-nya terbatas, dan gimana FE (mobile & backoffice) harus nge-handle URL yang short-lived/kadaluarsa.

> Dokumen verifikasi **selalu privat** (RULES §3 & DO/DON'T). File-nya disimpan di bucket private, **beda dari** media lain (avatar, thumbnail news, gambar event, dll) yang disimpan di bucket public dan URL-nya statis/permanen. Aturan di doc ini **khusus** buat dokumen verifikasi — jangan disamain sama pola upload media public.

---

## 1. Model Akses — Siapa Boleh Apa

| Aktor | Upload (kirim file) | Lihat/download file setelah submit |
|-------|:---:|:---:|
| Pemohon (owner user / owner merchant / Leader komunitas) — via **Mobile** | ✅ (entitas miliknya sendiri saja) | ❌ — mobile **tidak pernah** dapat URL lihat-ulang dokumen. Setelah upload, file di sisi pemohon dianggap sudah "terkirim"; app tidak menyediakan halaman review dokumen yang dikirim. |
| Member lain / publik | ❌ | ❌ |
| Admin Regional | ❌ | ❌ |
| Superadmin — via **Backoffice** | ❌ (Superadmin ga pernah upload dokumen atas nama pemohon) | ✅ — satu-satunya pihak yang dapat URL lihat/download dokumen, khusus saat review pengajuan (`verification.view_queue`) |

**Kesimpulan:**
- **Write access** (kirim file) → dibatasi ke **pemilik entitas**, lewat **Mobile**, dienforce di usecase Submit (cek ownership `entity_id`) — bukan di endpoint presign (presign cuma butuh login, validasi ownership terjadi saat `POST /requests`).
- **Read access** (lihat/download file) → dibatasi ke **Superadmin**, lewat **Backoffice**, dienforce oleh permission `verification.view_queue` + endpoint `GET /requests/:id`.
- Tidak ada aktor lain yang pernah menerima URL dokumen dalam bentuk apapun. `GET /requests/mine` (mobile) sengaja **tidak** mengembalikan field `documents` sama sekali (lihat `API_SPEC_VERIFICATION_MOBILE.md` §3) — cuma `status` + `reason`.

---

## 2. Kenapa URL-nya Terbatas (Private Bucket + Signed URL)

File dokumen disimpan di **bucket private** (beda fisik dari bucket public yang dipakai avatar/thumbnail/dll). Bucket private ini:
- Tidak punya URL publik yang bisa diakses langsung — **tidak ada CDN URL statis** seperti media lain.
- Satu-satunya cara akses adalah lewat **presigned URL** (URL yang di-generate BE dengan signature + waktu kadaluarsa/TTL), untuk dua keperluan berbeda:
  1. **Presigned PUT URL** (upload) — dipakai Mobile buat kirim file dari device ke storage.
  2. **Presigned GET URL** (download/lihat) — dipakai Backoffice buat Superadmin lihat isi dokumen saat review.
- Kedua jenis URL ini **short-lived** — begitu TTL lewat, URL langsung tidak valid (storage balas `403 Forbidden`), walaupun object-nya sendiri masih ada di storage.

**Implikasi penting buat FE:** URL presign (baik PUT maupun GET) **tidak boleh disimpan/dipakai lama**. Jangan cache URL ini di local storage, jangan simpan permanen di state management buat dipakai lintas sesi — selalu treat sebagai nilai sekali-pakai yang bisa basi.

---

## 3. Upload Dokumen (Mobile)

Alurnya **sama seperti pola upload media lain** yang sudah dipakai di app (avatar, community post, bukti transfer subscription, dll — presign → PUT langsung ke storage → confirm implisit). Yang beda **cuma dua hal**:

1. Presign-nya minta privacy **private**, bukan public → response presign **tidak ada field `public_url`** (dokumen lain biasanya balikin `public_url` juga, verifikasi sengaja tidak, karena URL publik ke bucket private itu tidak valid/404).
2. Endpoint & object path-nya khusus modul verification.

### Flow

```
Mobile App                          API (k-forum-api)                 Storage (private bucket)
    |                                      |                                    |
    |-- POST /requests/documents/presign ->|                                    |
    |   { content_type: "image/jpeg" }     |                                    |
    |<-- { presign_url, key, expires_in } -|                                    |
    |                                      |                                    |
    |-- PUT presign_url (body: file) ----------------------------------------->|  (langsung ke storage,
    |<-- 200 OK -------------------------------------------------------------->|   bukan lewat API)
    |                                      |                                    |
    |-- POST /requests                     |                                    |
    |   { documents: [{doc_type, url: key}], ... } ->|                          |
    |<-- 201 { id, status: "pending" } ----|                                    |
```

### Endpoint

**1. Minta presigned upload URL**

- `POST /api/v1/mobile/verification/requests/documents/presign`
- **Auth:** Wajib (member manapun boleh minta — ownership dicek nanti pas submit, bukan di sini)
- **Request:**
  ```json
  { "content_type": "image/jpeg" }
  ```
- **Content-Type yang diterima:** `image/jpeg`, `image/png`, `image/webp`, `application/pdf` — selain itu, `422 CONTENT_TYPE_NOT_ALLOWED`.
- **Response 200:**
  ```json
  {
    "data": {
      "presign_url": "https://storage.internal/private/verification/documents/<uuid>.jpg?X-Amz-...",
      "key": "s3:/verification/documents/<uuid>.jpg",
      "expires_in": 900
    }
  }
  ```
  - `presign_url` — URL buat `PUT` file langsung ke storage. **Berlaku 900 detik (15 menit)** sejak diterbitkan.
  - `key` — identifier yang harus dikirim balik di field `documents[].url` saat submit (§bawah). **Bukan** URL yang bisa dibuka browser — cuma reference internal.
  - Tidak ada `public_url` di response ini (beda dari presign media lain) — jangan expect field itu ada.

**2. Upload file ke storage**

- `PUT <presign_url>` — body = raw file bytes, header `Content-Type` **harus sama** dengan yang dikirim saat minta presign.
- Request ini **langsung ke storage**, bukan ke `k-forum-api`. Tidak butuh `Authorization` header k-forum-api (signature presign yang jadi otorisasinya).
- Sukses = `200 OK` dari storage (bukan dari backend).

**3. Submit pengajuan (pakai `key` dari langkah 1)**

- `POST /api/v1/mobile/verification/requests`
- Field `documents[].url` diisi dengan **`key`** dari langkah 1, bukan `presign_url`.
  ```json
  {
    "entity_type": "user",
    "entity_id": "usr_uuid_self",
    "documents": [
      { "doc_type": "kta_kai", "url": "s3:/verification/documents/<uuid>.jpg" }
    ],
    "note": "Pengurus KAI wilayah Seoul"
  }
  ```
- Saat request ini diproses, backend otomatis "confirm" key yang dipakai (menandai file sebagai resmi terpakai, bukan orphan) — FE **tidak perlu** panggil endpoint confirm terpisah. Ini beda dari beberapa modul media lain yang punya endpoint `/confirm` eksplisit; untuk verification, konfirmasi implisit terjadi di dalam `POST /requests`.

> Detail lengkap request/response `POST /requests` (validasi, error 400/403/409/422) sudah ada di `API_SPEC_VERIFICATION_MOBILE.md` §2 — tidak diulang di sini.

---

## 4. Lihat Dokumen (Backoffice)

Superadmin **tidak** akses dokumen lewat URL statis. Setiap kali halaman detail pengajuan dibuka, backend generate **signed GET URL baru** untuk tiap dokumen — URL ini yang dipakai buat menampilkan/download file.

- `GET /api/v1/web/verification/requests/:id`
- **Auth:** Superadmin, permission `verification.view_queue`
- **Response 200** (potongan relevan):
  ```json
  {
    "data": {
      "documents": [
        {
          "doc_type": "kta_kai",
          "url": "https://storage.internal/private/verification/documents/<uuid>.jpg?X-Amz-...",
          "uploaded_at": "2026-07-10T09:00:00.000Z"
        }
      ]
    }
  }
  ```
- `documents[].url` di response ini adalah **signed GET URL, berlaku 15 menit** sejak endpoint dipanggil — **bukan** URL permanen. Dibuka di tab baru / `<img src>` / didownload langsung bisa, tapi kalau dibuka setelah 15 menit dari saat `GET /requests/:id` dipanggil, storage akan balas `403 Forbidden`.
- **Tidak ada endpoint terpisah** untuk "minta URL dokumen" — satu-satunya cara mendapatkan signed URL adalah lewat response detail ini. Kalau butuh URL baru (karena yang lama sudah/hampir kadaluarsa), **panggil ulang `GET /requests/:id`** — endpoint ini idempotent dan setiap panggilan menghasilkan signed URL baru untuk semua dokumen di pengajuan itu.

---

## 5. Mekanisme Ambil-Ulang URL yang Kadaluarsa

Karena kedua jenis URL (presign PUT di mobile, presign GET di backoffice) short-lived, FE **wajib** punya strategi kalau URL basi di tengah proses. Ini bukan skenario tepi yang boleh diabaikan — TTL 15 menit itu **realistis kena** kalau user lama milih file, koneksi lambat, atau admin buka halaman review lalu ditinggal lama sebelum benar-benar lihat dokumennya.

### 5.1 Mobile — Presign PUT (upload) expired

**Kapan bisa terjadi:** user minta presign, lalu berlama-lama (pilih ulang file, ganti pikiran, app di-background lalu balik lagi) sebelum benar-benar `PUT` filenya → lewat 15 menit.

**Cara deteksi:** `PUT presign_url` balas `403 Forbidden` dari storage (bukan error JSON k-forum-api, karena request ini langsung ke storage).

**Mekanisme wajib di FE:**
1. **Minta presign sedekat mungkin dengan aksi upload** — jangan minta presign begitu user masuk halaman "Ajukan Verifikasi", minta presign **tepat setelah** user pilih file dan siap upload (per dokumen, satu presign = satu file).
2. Kalau `PUT` gagal dengan `403`, **otomatis minta presign baru** (`POST .../documents/presign` ulang dengan `content_type` yang sama) lalu retry `PUT` sekali. Kalau gagal lagi, baru tampilkan error ke user ("Upload gagal, coba lagi").
3. **Jangan** reuse `key`/`presign_url` lama yang sudah gagal — presign baru menghasilkan `key` baru (object baru), jadi hasil retry harus pakai `key` yang baru diterima, bukan yang lama.
4. Kalau user submit form (`POST /requests`) sebelum semua dokumen selesai di-`PUT`, tahan submit sampai semua upload sukses — jangan kirim `key` yang belum ter-upload.

### 5.2 Backoffice — Signed GET (lihat dokumen) expired

**Kapan bisa terjadi:** Superadmin buka halaman detail pengajuan, lalu meninggalkannya terbuka (multitasking, meeting, dll) lebih dari 15 menit sebelum benar-benar klik lihat dokumen; atau membuka banyak tab dokumen berurutan sampai total waktu review lewat TTL.

**Cara deteksi:** gambar gagal load (`<img>` error event) atau, kalau dibuka di tab baru/download, browser/storage balas `403 Forbidden`.

**Mekanisme wajib di FE:**
1. **Jangan fetch detail sekali di awal lalu dipakai selamanya** selama halaman terbuka. Minimal: refetch `GET /requests/:id` setiap kali tab/section "Documents" dibuka atau di-fokus ulang (misal setiap kali user pindah dari tab lain balik ke tab Documents), supaya signed URL yang ditampilkan selalu segar.
2. Untuk viewer gambar (bukan cuma link download polos): pasang **error handler** pada elemen gambar — kalau gagal load, otomatis refetch `GET /requests/:id`, ambil `documents[].url` yang baru, lalu retry render sekali sebelum menampilkan pesan "Gagal memuat dokumen, coba refresh".
3. Untuk aksi "buka di tab baru"/"download": sediakan tombol refresh di dekat dokumen yang re-fetch detail dulu **sebelum** membuka URL — supaya URL yang dibuka selalu yang baru digenerate, bukan yang sudah nempel di state sejak halaman dibuka.
4. **Jangan** taruh `documents[].url` di URL address bar / bookmark / share link — karena akan basi dalam 15 menit dan tidak bisa "direfresh" tanpa akses balik ke halaman detail pengajuan.

> Referensi pola UI viewer gambar dengan lightbox/carousel yang sudah ada di codebase backoffice: modul Event (`UCarousel` + zoom modal). Modul Verification saat ini masih pakai link "View" polos tanpa viewer/error-handling — kalau viewer dokumen mau di-upgrade, pola Event bisa dijadikan referensi, ditambah poin 2 & 3 di atas yang belum ada di modul manapun.

---

## 6. Error Handling Reference

| Sumber | Kode/Status | Arti | Aksi FE |
|--------|-------------|------|---------|
| k-forum-api (presign mobile) | `422 CONTENT_TYPE_NOT_ALLOWED` | Content-type file tidak diterima | Validasi tipe file di FE sebelum minta presign (jpeg/png/webp/pdf) |
| Storage (PUT presign, mobile) | `403 Forbidden` (bukan JSON, response langsung dari storage) | Presign PUT sudah kadaluarsa (>15 menit) atau signature tidak cocok | Minta presign baru, retry sekali (§5.1) |
| k-forum-api (submit) | `400`/`403`/`409`/`422` | Lihat `API_SPEC_VERIFICATION_MOBILE.md` §2 | Tangani sesuai tabel di spec tersebut |
| Storage (GET signed, backoffice) | `403 Forbidden` saat load/`<img>` gagal | Signed GET URL sudah kadaluarsa (>15 menit sejak `GET /requests/:id`) | Refetch `GET /requests/:id`, retry sekali (§5.2) |
| k-forum-api (detail backoffice) | `404 NOT_FOUND` | Pengajuan tidak ditemukan | Redirect ke list / tampilkan not-found state |

---

## 7. Checklist Implementasi FE

### Mobile (belum ada implementasinya sama sekali — fitur baru)

- [ ] Screen "Ajukan Verifikasi" (per `entity_type`) — ambil `GET /requirements` dulu buat render daftar dokumen yang diterima + hint `any_of`/`all_of`.
- [ ] Per dokumen yang dipilih user: minta presign **saat user pilih file** (bukan di awal screen), simpan `key` di state form.
- [ ] Implementasi retry presign-expired sesuai §5.1 (request presign baru + retry PUT sekali).
- [ ] Submit `POST /requests` cuma setelah **semua** dokumen wajib ter-upload (semua `key` valid).
- [ ] **Tidak perlu** bikin halaman "lihat dokumen yang sudah disubmit" — backend sengaja tidak menyediakan endpoint itu untuk mobile (§1).
- [ ] Halaman status (`GET /requests/mine`) cukup tampilkan `status` + `reason` (kalau rejected/revoked) + tombol resubmit — **jangan** expect/render field dokumen di response ini.

### Backoffice (sudah ada, perlu upgrade)

- [ ] Tab "Documents" di halaman detail: refetch `GET /requests/:id` setiap kali tab dibuka/difokus ulang (§5.2 poin 1), bukan cuma sekali saat halaman mount.
- [ ] Ganti/lengkapi link "View" polos (`target="_blank"`, `:to="doc.url"` saat ini) dengan viewer yang: (a) render `<img>`/preview inline kalau content-type gambar, dengan `@error` handler yang trigger refetch+retry (§5.2 poin 2); (b) untuk PDF, tetap link download tapi refresh detail dulu sebelum membuka (§5.2 poin 3).
- [ ] Jangan taruh `doc.url` di computed/ref yang dipertahankan lintas navigasi — treat sebagai nilai sekali-render.

---

## 8. Referensi

- `API_SPEC_VERIFICATION_MOBILE.md` — kontrak lengkap endpoint mobile (requirements, submit, mine).
- `API_SPEC_VERIFICATION_BACKOFFICE.md` — kontrak lengkap endpoint backoffice (list, detail, approve/reject/revoke, events, requirements config).
- `VERIFICATION_RULES.md` §3 & §9 — kenapa dokumen harus privat, DO/DON'T.
- `VERIFICATION_DB_SCHEMA.md` — struktur `documents` JSONB di tabel `verifications`.

---

*Verification Document Upload & Access Spec v1.0 — KAI App. Pelengkap pipeline RULES → DB Schema → API Spec. Last updated: 2026-07-16.*
