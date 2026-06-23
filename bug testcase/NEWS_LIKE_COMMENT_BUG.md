# Laporan Bug — News Like & Comments (Mobile API)

- **Tanggal:** 19 Juni 2026
- **Environment:** dev — `http://192.168.1.29:8888/api/v1`
- **Cara test:** manual via curl, login sebagai member biasa (token valid, 248 char)
- **Sisi:** **Backend** — server mengembalikan `500 ERR_INTERNAL`. App mobile memanggil endpoint dengan benar.

---

## Ringkasan

| Aksi | Endpoint | Status | Verdict |
| --- | --- | --- | --- |
| Like artikel | `POST /mobile/news/articles/{id}/like` | **500** `ERR_INTERNAL` | 🔴 Backend error |
| Unlike artikel | `DELETE /mobile/news/articles/{id}/like` | **500** `ERR_INTERNAL` | 🔴 Backend error |
| Ambil daftar komentar | `GET /mobile/news/articles/{id}/comments` | **500** `ERR_INTERNAL` | 🔴 Backend error |
| Kirim komentar | `POST /mobile/news/articles/{id}/comments` | **201** `comment posted` | 🟢 OK |
| Detail artikel | `GET /mobile/news/articles/{id}` | **200** | 🟢 OK |

Dikonfirmasi **konsisten di 3 artikel berbeda** (bukan masalah data 1 artikel).

Isi response error: `{"status":"error","status_code":500,"error_code":"ERR_INTERNAL","errors":"An internal error occurred. Please try again."}`

---

## Dampak ke user (mobile)

1. **Tombol Like tidak berfungsi** — setiap tap → 500, app menampilkan SnackBar error.
2. **Komentar tampak "tidak bisa"** — POST komentar **sebenarnya berhasil (201)**, tapi
   `GET .../comments` selalu 500, jadi komentar yang baru dikirim **tidak pernah muncul**
   di daftar. Dari sisi user terlihat seperti gagal padahal datanya masuk ke DB.

---

## Bukti (curl)

```bash
API=http://192.168.1.29:8888/api/v1
TOKEN=<member access_token>   # dari POST /mobile/auth/login {identifier,password}
ART=8b9bad7f-7edd-4ff1-90f8-f6a6a292b6a5

# LIKE → 500
curl -H "Authorization: Bearer $TOKEN" -X POST   "$API/mobile/news/articles/$ART/like"
# → 500 {"error_code":"ERR_INTERNAL","errors":"An internal error occurred. Please try again."}

# UNLIKE → 500
curl -H "Authorization: Bearer $TOKEN" -X DELETE "$API/mobile/news/articles/$ART/like"
# → 500

# GET COMMENTS → 500
curl -H "Authorization: Bearer $TOKEN" "$API/mobile/news/articles/$ART/comments?limit=5&offset=0"
# → 500

# POST COMMENT → 201 (OK)
curl -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
     -X POST "$API/mobile/news/articles/$ART/comments" -d '{"content":"Mantap beritanya"}'
# → 201 {"data":{"id":"e064d13a-48e5-4aec-b3aa-81613a94e754"},"message":"comment posted"}
```

Article id lain yang juga 500: `851bdf7e-f4a0-4b6e-8ba9-8f45cba9005e`, `0b14473c-caf7-4103-8442-9a85d53ad4df`.

---

## Untuk tim backend

- Cek log server untuk handler **News Like** (`POST`/`DELETE /mobile/news/articles/{id}/like`)
  dan **List Comments** (`GET /mobile/news/articles/{id}/comments`). Keduanya melempar
  500 (unhandled exception), bukan error validasi.
- Endpoint POST comment & GET detail sehat → kemungkinan masalah spesifik di query/join
  bagian like-count & comment-list (mis. kolom/relasi yang belum ada, null pointer, atau
  agregasi `like_count`/`comment_count`).

---

## Mitigasi sementara di app (sudah dikerjakan)

Di [news_comments_bottom_sheet.dart](../../lib/features/news/presentation/widgets/news_comments_bottom_sheet.dart):

- **Error daftar komentar tidak lagi disembunyikan** — saat `GET comments` gagal & daftar
  kosong, ditampilkan state **error + tombol "Retry"** (sebelumnya hanya muncul list kosong
  "Be the first to comment." yang menyesatkan).
- **Komentar baru di-prepend optimistis** — setelah POST 201, komentar langsung tampil di
  daftar tanpa menunggu `GET comments` (yang masih 500), jadi tidak hilang.

Tombol **Like** sudah benar (menampilkan SnackBar error saat 500) — tidak ada perubahan
app yang diperlukan; tinggal menunggu fix backend.
