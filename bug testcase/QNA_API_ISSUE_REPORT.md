# Q&A Module — Backend API Issue Report

Dokumen ini merinci ketidaksesuaian (inconsistency) dan masalah (bugs) yang ditemukan pada endpoint API Q&A Mobile. Masalah ini menyebabkan bug visual dan fungsional di aplikasi Mobile Client.

---

## 1. Masalah Utama: Relasi `answer` Bernilai `null` pada List Endpoint

### Deskripsi
Pada endpoint list pertanyaan saya (`GET /api/v1/mobile/qna/questions/me`), pertanyaan yang sudah berstatus `"answered"` mengembalikan field `"answer": null`. 
Namun, ketika memanggil endpoint detail (`GET /api/v1/mobile/qna/questions/me/{id}`), objek `"answer"` terisi secara lengkap.

### Dampak Pada Mobile
Aplikasi mobile menggunakan field `q.answer` untuk merender preview boks jawaban (`QnaAnswerBox`) langsung di halaman riwayat pertanyaan. Karena list API mengembalikan `null`, boks preview ini tidak muncul di list utama dan baru muncul ketika user mengetuk item tersebut untuk membuka lembar detail.

### Bukti Payload JSON

#### `GET /mobile/qna/questions/me` (List)
```json
{
  "id": "d10211d8-9d4f-43b0-94b2-bc494f73cc12",
  "status": "answered",
  "answer": null  // <--- BUG: Harusnya berisi preview objek jawaban
}
```

#### `GET /mobile/qna/questions/me/d10211d8-9d4f-43b0-94b2-bc494f73cc12` (Detail)
```json
{
  "id": "d10211d8-9d4f-43b0-94b2-bc494f73cc12",
  "status": "answered",
  "answer": {  // <--- BENAR: Berisi detail jawaban
    "id": "d7a687d9-c275-4ba7-ab9d-b291dd13df38",
    "answered_by": "0e4238fa-ef6d-45d2-be80-614931526212",
    "answered_by_name": "User God Seed",
    "answer_text": "jokoooooow wowowowowoowwowowoww",
    "created_at": "2026-06-26T04:15:41.483442Z"
  }
}
```

---

## 2. Masalah: Field `visibility` Kosong pada List Endpoint

### Deskripsi
* Pada endpoint list (`GET /mobile/qna/questions/me`), field `"visibility"` dikembalikan sebagai string kosong (`""`).
* Pada endpoint detail (`GET /mobile/qna/questions/me/{id}`), field `"visibility"` dikembalikan dengan benar (misal: `"public"` atau `"private"`).

### Dampak Pada Mobile
Aplikasi tidak dapat menampilkan status visibilitas pertanyaan (apakah ini pertanyaan forum publik atau privat 1-on-1) pada kartu di halaman utama.

---

## 3. Masalah: `answer_count` dan `accepted_count` Selalu `0`

### Deskripsi
Pada pertanyaan yang sudah berstatus `"answered"` dan memiliki jawaban valid, field `"answer_count"` dan `"accepted_count"` tetap mengembalikan nilai `0` baik di list maupun detail endpoint.

### Dampak Pada Mobile
Aplikasi tidak dapat menampilkan jumlah jawaban yang benar untuk forum komunitas di halaman daftar pertanyaan publik/forum.

---

## 4. Perbedaan Skema Serialisasi `category`

### Deskripsi
Terdapat perbedaan representasi objek kategori (category) di antara kedua endpoint:
* **List Endpoint:** Menggunakan flat keys (`category_id`, `category_name`).
* **Detail Endpoint:** Menggunakan nested object (`category: { id: "...", name: "..." }`).

---

## Rekomendasi Perbaikan di Backend

1. **Eager Loading Relasi `answer`:**
   Pastikan query untuk list endpoint (`GET /mobile/qna/questions/me`) melakukan eager load/join ke relasi `answer` sehingga data jawaban tetap dikirim di list view.
   
2. **Standardisasi Serializer:**
   Gunakan serializer (atau resource transformer) yang sama untuk model `Question` di list dan detail endpoint agar field seperti `visibility`, `category`, dan `answer_count` memiliki format serta nilai yang konsisten.
