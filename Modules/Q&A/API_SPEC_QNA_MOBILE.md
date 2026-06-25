# API Specification — Q&A / FAQ Module (Mobile Client)

> **Dibuat:** 2026-05-30 · **Revisi:** 2026-06-25 (forum komunitas: jawab antar-member, accepted answer, assignment)
> **Stack:** Flutter Mobile Client
> **Base URL Prefix:** `/api/v1/mobile/qna`

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Data Models](#data-models)
3. [Endpoints](#endpoints)
   - [1. Get Categories](#1-get-categories)
   - [2. Get FAQ List](#2-get-faq-list)
   - [3. Get FAQ Detail](#3-get-faq-detail)
   - [4. Search FAQ](#4-search-faq)
   - [5. Vote FAQ Helpful](#5-vote-faq-helpful)
   - [6. Submit Question](#6-submit-question)
   - [7. Get My Questions](#7-get-my-questions)
   - [8. Get My Question Detail](#8-get-my-question-detail)
   - [9. Get Public Question Feed](#9-get-public-question-feed)
   - [10. Get Public Question Detail](#10-get-public-question-detail)
   - [11. Post Answer](#11-post-answer)
   - [12. Edit My Answer](#12-edit-my-answer)
   - [13. Upvote Answer](#13-upvote-answer)
   - [14. Mark Answer Accepted (Expert/Admin)](#14-mark-answer-accepted-expertadmin)
   - [15. Get Questions Assigned To Me](#15-get-questions-assigned-to-me)
4. [Status Code Reference](#status-code-reference)
5. [Notes & Best Practices](#notes--best-practices)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/mobile/qna`
- **Headers Global:**
  ```
  Content-Type: application/json
  Accept: application/json
  Accept-Language: <lang_code>    (ko | id | en — default: ko)
  X-Locale: <lang_code>
  Authorization: Bearer <access_token>
  ```

- **Permission / Benefit yang dibutuhkan:**

| Endpoint | Auth | Benefit / Keterangan |
|----------|------|---------------------|
| GET categories | ✅ Required | `read_content` |
| GET faqs | ✅ Required | `read_content` |
| GET faqs/:id | ✅ Required | `read_content` |
| GET search | ✅ Required | `read_content` |
| POST faqs/:id/vote | ✅ Required | member (Standard / Pro) |
| POST questions | ✅ Required | `ask_qna` (benefit subscription) |
| GET questions/my | ✅ Required | member (Standard / Pro) |
| GET questions/my/:id | ✅ Required | member, hanya milik sendiri |
| GET questions/public | ✅ Required | member (Standard / Pro) |
| GET questions/public/:id | ✅ Required | member (Standard / Pro) |
| POST questions/:id/answers | ✅ Required | `answer_qna`, atau expert yang di-assign, atau admin |
| PUT answers/:id | ✅ Required | pemilik jawaban |
| POST answers/:id/upvote | ✅ Required | member (Standard / Pro), bukan pemilik jawaban |
| POST answers/:id/accept | ✅ Required | permission `validate_qna_answer` (expert assigned / admin) |
| GET questions/assigned | ✅ Required | permission `answer_qna` + ada assignment |

> **Catatan permission expert:** Endpoint `accept` dan akses pertanyaan privat yang di-assign dikontrol lewat permission, bukan label plan. Backend mem-resolve hak ini saat request (lihat `GET /api/v1/auth/session`).

---

## Data Models

### CategoryObject
```json
{
  "id": "uuid",
  "name": "Imigrasi & Visa",
  "slug": "imigrasi-visa",
  "description": "Informasi seputar visa, izin tinggal, dan imigrasi di Indonesia",
  "icon": "ti-passport",
  "color": "#534AB7",
  "sort_order": 1,
  "faq_count": 12
}
```

### FaqItemListObject (List — simplified)
```json
{
  "id": "uuid",
  "question": "Bagaimana cara mendaftar NPWP sebagai WNA?",
  "category_id": "uuid",
  "is_pinned": false,
  "helpful_count": 87,
  "view_count": 342
}
```

### FaqItemDetailObject
```json
{
  "id": "uuid",
  "category": {
    "id": "uuid",
    "name": "Pajak",
    "slug": "pajak"
  },
  "question": "Bagaimana cara mendaftar NPWP sebagai WNA?",
  "answer": "WNA yang bekerja di Indonesia wajib memiliki NPWP...",
  "tags": ["npwp", "wna", "pajak"],
  "is_pinned": false,
  "view_count": 342,
  "helpful_count": 87,
  "not_helpful_count": 3,
  "user_vote": null,
  "status": "published",
  "published_at": "2026-01-10T08:00:00.000Z",
  "updated_at": "2026-04-15T12:30:00.000Z"
}
```

> `user_vote`: `true` = helpful, `false` = not helpful, `null` = belum vote.

### UserQuestionListObject
```json
{
  "id": "uuid",
  "category": {
    "id": "uuid",
    "name": "Imigrasi & Visa"
  },
  "question_text": "Apakah WNA bisa memiliki properti di Indonesia?",
  "status": "answered",
  "has_answer": true,
  "created_at": "2026-05-19T14:00:00.000Z",
  "answered_at": "2026-05-20T10:00:00.000Z"
}
```

### UserQuestionDetailObject
```json
{
  "id": "uuid",
  "category": {
    "id": "uuid",
    "name": "Imigrasi & Visa"
  },
  "question_text": "Apakah WNA bisa memiliki properti di Indonesia?",
  "attachment_urls": [],
  "status": "answered",
  "bot_responded": true,
  "bot_response": "Terima kasih, pertanyaan Anda sedang diproses oleh tim kami.",
  "answer": {
    "id": "uuid",
    "answer_text": "WNA dapat memiliki properti melalui skema Hak Pakai...",
    "attachment_urls": [],
    "is_public": true,
    "converted_to_faq_id": null,
    "answered_by": {
      "id": "uuid",
      "name": "Admin KAI Pusat"
    },
    "created_at": "2026-05-20T10:00:00.000Z"
  },
  "created_at": "2026-05-19T14:00:00.000Z",
  "answered_at": "2026-05-20T10:00:00.000Z"
}
```

### PaginationObject
```json
{
  "limit": 20,
  "offset": 0,
  "total": 85,
  "has_next": true,
  "has_prev": false
}
```

### Error Responses
```json
// 400 / 403 / 404
{
  "message": "Deskripsi error"
}

// 422 Validation Error
{
  "message": "Data input tidak valid",
  "errors": {
    "question_text": ["Pertanyaan tidak boleh kosong"],
    "category_id": ["Kategori tidak valid"]
  }
}
```

---

## Endpoints

### 1. Get Categories

Ambil semua kategori FAQ yang aktif, terurut berdasarkan `sort_order`.

- **URL:** `GET /api/v1/mobile/qna/categories`
- **Auth:** Required (member)
- **Query Params:** —

- **Response 200:**
```json
{
  "data": [
    {
      "id": "cat_uuid_1",
      "name": "Imigrasi & Visa",
      "slug": "imigrasi-visa",
      "description": "Informasi seputar visa dan imigrasi",
      "icon": "ti-passport",
      "color": "#534AB7",
      "sort_order": 1,
      "faq_count": 12
    },
    {
      "id": "cat_uuid_2",
      "name": "Pajak",
      "slug": "pajak",
      "description": "NPWP, SPT, dan kewajiban pajak di Indonesia",
      "icon": "ti-receipt",
      "color": "#0F6E56",
      "sort_order": 2,
      "faq_count": 8
    },
    {
      "id": "cat_uuid_3",
      "name": "Sistem Aplikasi KAI",
      "slug": "sistem-aplikasi",
      "description": "Panduan penggunaan aplikasi KAI",
      "icon": "ti-device-mobile",
      "color": "#639922",
      "sort_order": 6,
      "faq_count": 15
    }
  ]
}
```

---

### 2. Get FAQ List

Ambil daftar FAQ published. Pinned item selalu muncul di atas.

- **URL:** `GET /api/v1/mobile/qna/faqs`
- **Auth:** Required (member)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `category_id` | string (UUID) | No | — | Filter per kategori |
| `limit` | int | No | 20 | Max: 50 |
| `offset` | int | No | 0 | — |

- **Response 200:**
```json
{
  "data": [
    {
      "id": "faq_uuid_1",
      "question": "Bagaimana cara mendaftar NPWP sebagai WNA?",
      "category_id": "cat_uuid_2",
      "is_pinned": true,
      "helpful_count": 87,
      "view_count": 342
    },
    {
      "id": "faq_uuid_2",
      "question": "Berapa lama proses pengurusan NPWP?",
      "category_id": "cat_uuid_2",
      "is_pinned": false,
      "helpful_count": 45,
      "view_count": 210
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 8,
    "has_next": false,
    "has_prev": false
  }
}
```

---

### 3. Get FAQ Detail

Ambil detail satu FAQ item beserta jawaban lengkap. Auto-increment `view_count` setiap kali diakses.

- **URL:** `GET /api/v1/mobile/qna/faqs/{faq_id}`
- **Auth:** Required (member)
- **Path Params:** `faq_id` — UUID FAQ item

- **Response 200:**
```json
{
  "data": {
    "id": "faq_uuid_1",
    "category": {
      "id": "cat_uuid_2",
      "name": "Pajak",
      "slug": "pajak"
    },
    "question": "Bagaimana cara mendaftar NPWP sebagai WNA?",
    "answer": "WNA yang bekerja atau memiliki penghasilan di Indonesia wajib memiliki NPWP. Caranya:\n\n1. Kunjungi Kantor Pelayanan Pajak (KPP) terdekat\n2. Bawa paspor dan dokumen izin tinggal (KITAS/KITAP)\n3. Isi formulir pendaftaran\n4. NPWP diterbitkan dalam 1 hari kerja",
    "tags": ["npwp", "wna", "pajak", "kitas"],
    "is_pinned": true,
    "view_count": 343,
    "helpful_count": 87,
    "not_helpful_count": 3,
    "user_vote": null,
    "status": "published",
    "published_at": "2026-01-10T08:00:00.000Z",
    "updated_at": "2026-04-15T12:30:00.000Z"
  }
}
```

- **Response 404:**
```json
{ "message": "FAQ tidak ditemukan" }
```

---

### 4. Search FAQ

Full-text search di seluruh FAQ published. Mengembalikan hasil terurut berdasarkan relevansi.

- **URL:** `GET /api/v1/mobile/qna/search`
- **Auth:** Required (member)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `q` | string | **Yes** | — | Keyword pencarian (min 2 karakter) |
| `category_id` | string (UUID) | No | — | Filter hasil per kategori |
| `limit` | int | No | 10 | Max: 30 |
| `offset` | int | No | 0 | — |

- **Response 200:**
```json
{
  "data": [
    {
      "id": "faq_uuid_1",
      "question": "Bagaimana cara mendaftar NPWP sebagai WNA?",
      "answer_excerpt": "WNA yang bekerja di Indonesia wajib memiliki NPWP. Caranya: kunjungi KPP terdekat...",
      "category": {
        "id": "cat_uuid_2",
        "name": "Pajak"
      },
      "helpful_count": 87,
      "relevance_score": 0.92
    },
    {
      "id": "faq_uuid_5",
      "question": "Berapa tarif PPh untuk WNA yang bekerja di Indonesia?",
      "answer_excerpt": "WNA yang berstatus sebagai subjek pajak dalam negeri dikenakan tarif PPh...",
      "category": {
        "id": "cat_uuid_2",
        "name": "Pajak"
      },
      "helpful_count": 32,
      "relevance_score": 0.74
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0,
    "total": 2,
    "has_next": false,
    "has_prev": false
  },
  "query": "npwp wna"
}
```

- **Response 400** (query < 2 karakter):
```json
{ "message": "Kata kunci pencarian minimal 2 karakter" }
```

---

### 5. Vote FAQ Helpful

Member memberi vote apakah FAQ bermanfaat atau tidak. Jika sudah pernah vote, vote akan diperbarui (toggle/update).

- **URL:** `POST /api/v1/mobile/qna/faqs/{faq_id}/vote`
- **Auth:** Required (member)
- **Path Params:** `faq_id` — UUID FAQ item

- **Request Body:**
```json
{
  "is_helpful": true
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `is_helpful` | boolean | **Yes** | `true` = helpful, `false` = not helpful |

- **Response 200:**
```json
{
  "data": {
    "faq_id": "faq_uuid_1",
    "user_vote": true,
    "helpful_count": 88,
    "not_helpful_count": 3
  },
  "message": "Vote berhasil disimpan"
}
```

- **Response 404:**
```json
{ "message": "FAQ tidak ditemukan" }
```

---

### 6. Submit Question

Member mengajukan pertanyaan baru. Pertanyaan masuk ke antrian moderasi superadmin. Bot fallback otomatis merespons jika diaktifkan.

- **URL:** `POST /api/v1/mobile/qna/questions`
- **Auth:** Required (member, butuh benefit `ask_qna`)

- **Request Body:**
```json
{
  "category_id": "cat_uuid_1",
  "question_text": "Apakah WNA bisa memiliki properti di Indonesia dengan status KITAS?",
  "visibility": "public",
  "attachment_urls": [
    "https://cdn.example.com/uploads/doc_uuid.jpg"
  ]
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `category_id` | string (UUID) | No | Pilihan kategori (opsional, membantu routing) |
| `question_text` | string | **Yes** | Teks pertanyaan. Min: 10 char, max: 1000 char |
| `visibility` | string | No | `private` (default) atau `public`. Public tampil di forum setelah di-approve admin |
| `attachment_urls` | string[] | No | URL lampiran yang sudah di-upload. Max: 3 item |

> Validasi kata terlarang dijalankan backend (modul System Settings) sebelum disimpan. Jika konten ditolak → `422`.

- **Response 201:**
```json
{
  "data": {
    "id": "q_uuid_1",
    "question_text": "Apakah WNA bisa memiliki properti di Indonesia dengan status KITAS?",
    "category": {
      "id": "cat_uuid_1",
      "name": "Properti & Kepemilikan"
    },
    "visibility": "public",
    "status": "pending",
    "bot_response": "Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja.",
    "created_at": "2026-05-30T10:00:00.000Z"
  },
  "message": "Pertanyaan berhasil dikirim. Menunggu persetujuan moderator."
}
```

- **Response 403** (benefit `ask_qna` tidak tersedia di plan user):
```json
{
  "message": "Fitur mengajukan pertanyaan memerlukan upgrade plan. Silakan upgrade ke Pro untuk menggunakan fitur ini."
}
```

- **Response 422:**
```json
{
  "message": "Data input tidak valid",
  "errors": {
    "question_text": ["Pertanyaan minimal 10 karakter"],
    "attachment_urls": ["Maksimal 3 lampiran yang diizinkan"]
  }
}
```

---

### 7. Get My Questions

Ambil daftar semua pertanyaan yang pernah diajukan oleh member yang sedang login, diurutkan dari yang terbaru.

- **URL:** `GET /api/v1/mobile/qna/questions/my`
- **Auth:** Required (member)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `status` | string | No | — | Filter: `pending`, `answered`, `rejected`, `converted` |
| `limit` | int | No | 20 | — |
| `offset` | int | No | 0 | — |

- **Response 200:**
```json
{
  "data": [
    {
      "id": "q_uuid_1",
      "category": {
        "id": "cat_uuid_1",
        "name": "Properti & Kepemilikan"
      },
      "question_text": "Apakah WNA bisa memiliki properti di Indonesia?",
      "status": "answered",
      "has_answer": true,
      "created_at": "2026-05-19T14:00:00.000Z",
      "answered_at": "2026-05-20T10:00:00.000Z"
    },
    {
      "id": "q_uuid_2",
      "category": null,
      "question_text": "Bagaimana proses perpanjangan KITAS?",
      "status": "pending",
      "has_answer": false,
      "created_at": "2026-05-28T09:00:00.000Z",
      "answered_at": null
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 2,
    "has_next": false,
    "has_prev": false
  }
}
```

---

### 8. Get My Question Detail

Ambil detail satu pertanyaan beserta jawaban admin (jika sudah dijawab). Hanya bisa diakses oleh member yang mengajukan pertanyaan tersebut.

- **URL:** `GET /api/v1/mobile/qna/questions/my/{question_id}`
- **Auth:** Required (member, hanya milik sendiri)
- **Path Params:** `question_id` — UUID pertanyaan

- **Response 200:**
```json
{
  "data": {
    "id": "q_uuid_1",
    "category": {
      "id": "cat_uuid_1",
      "name": "Properti & Kepemilikan"
    },
    "question_text": "Apakah WNA bisa memiliki properti di Indonesia dengan status KITAS?",
    "attachment_urls": [],
    "status": "answered",
    "bot_responded": true,
    "bot_response": "Terima kasih, pertanyaan Anda sedang diproses oleh tim kami.",
    "answer": {
      "id": "ans_uuid_1",
      "answer_text": "WNA dengan status KITAS dapat memiliki properti melalui skema Hak Pakai (HP). Berikut ketentuan lengkapnya:\n\n1. Minimal masa tinggal 1 tahun\n2. Nilai properti sesuai batas yang ditetapkan pemerintah daerah\n3. Tidak bisa dijual kepada WNA lain tanpa izin khusus",
      "attachment_urls": [],
      "is_public": true,
      "converted_to_faq_id": null,
      "answered_by": {
        "id": "admin_uuid",
        "name": "Admin KAI Pusat"
      },
      "created_at": "2026-05-20T10:00:00.000Z"
    },
    "created_at": "2026-05-19T14:00:00.000Z",
    "answered_at": "2026-05-20T10:00:00.000Z"
  }
}
```

- **Response 403** (bukan milik user):
```json
{ "message": "Anda tidak memiliki akses ke pertanyaan ini" }
```

- **Response 404:**
```json
{ "message": "Pertanyaan tidak ditemukan" }
```

---

### 9. Get Public Question Feed

Daftar pertanyaan **publik** yang sudah di-approve (forum komunitas). Tidak menampilkan pertanyaan privat milik siapa pun.

- **URL:** `GET /api/v1/mobile/qna/questions/public`
- **Auth:** Required (member)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `category_id` | string (UUID) | No | — | Filter per kategori |
| `filter` | string | No | `latest` | `latest`, `unanswered`, `most_upvoted` |
| `limit` | int | No | 20 | Max: 50 |
| `offset` | int | No | 0 | — |

- **Response 200:**
```json
{
  "data": [
    {
      "id": "q_uuid_10",
      "category": { "id": "cat_uuid_1", "name": "Properti & Kepemilikan" },
      "question_text": "Rekomendasi co-working space murah di Jakarta Selatan?",
      "asker": { "id": "user_uuid_9", "name": "Choi Woo" },
      "answer_count": 3,
      "accepted_count": 1,
      "view_count": 58,
      "status": "approved",
      "created_at": "2026-06-01T09:00:00.000Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 1, "has_next": false, "has_prev": false }
}
```

---

### 10. Get Public Question Detail

Detail pertanyaan publik beserta semua jawaban yang tampil. Auto-increment `view_count`. Jawaban diurut: accepted dulu, lalu upvote terbanyak.

- **URL:** `GET /api/v1/mobile/qna/questions/public/{question_id}`
- **Auth:** Required (member)

- **Response 200:**
```json
{
  "data": {
    "id": "q_uuid_10",
    "category": { "id": "cat_uuid_1", "name": "Properti & Kepemilikan" },
    "question_text": "Rekomendasi co-working space murah di Jakarta Selatan?",
    "attachment_urls": [],
    "asker": { "id": "user_uuid_9", "name": "Choi Woo" },
    "status": "approved",
    "answer_count": 3,
    "accepted_count": 1,
    "view_count": 59,
    "can_answer": true,
    "can_accept": false,
    "my_answer_id": null,
    "answers": [
      {
        "id": "ans_uuid_1",
        "answer_text": "Tiga rekomendasi: ...",
        "answerer": { "id": "user_uuid_5", "name": "Kang Dae" },
        "answerer_type": "member",
        "is_accepted": true,
        "accepted_by": { "id": "admin_uuid", "name": "Admin KAI" },
        "upvote_count": 12,
        "user_upvoted": false,
        "created_at": "2026-06-01T11:00:00.000Z"
      },
      {
        "id": "ans_uuid_2",
        "answer_text": "Bisa coba juga ...",
        "answerer": { "id": "user_uuid_7", "name": "Lim Ho" },
        "answerer_type": "member",
        "is_accepted": false,
        "accepted_by": null,
        "upvote_count": 4,
        "user_upvoted": true,
        "created_at": "2026-06-01T12:30:00.000Z"
      }
    ],
    "created_at": "2026-06-01T09:00:00.000Z"
  }
}
```

> - `can_answer`: apakah user yang login boleh menjawab (punya benefit `answer_qna` / expert / admin, dan belum punya jawaban di sini, dan status belum `closed`).
> - `can_accept`: apakah user boleh menandai jawaban valid (expert assigned / admin).
> - `my_answer_id`: ID jawaban milik user di pertanyaan ini, bila ada.

---

### 11. Post Answer

Menulis jawaban pada pertanyaan publik. Satu user maksimal satu jawaban per pertanyaan. Tergantung `answer_moderation_mode`, jawaban langsung `visible` (auto) atau `pending` (manual).

- **URL:** `POST /api/v1/mobile/qna/questions/{question_id}/answers`
- **Auth:** Required (benefit `answer_qna`, atau expert yang di-assign, atau admin)

- **Request Body:**
```json
{
  "answer_text": "Tiga rekomendasi co-working space: ...",
  "attachment_urls": []
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `answer_text` | string | **Yes** | Min: 10 char, max: 5000 char. Support Markdown |
| `attachment_urls` | string[] | No | Max: 3 item |

> Word filter dijalankan backend (System Settings) sebelum simpan. Konten ditolak → `422`.

- **Response 201 (auto mode):**
```json
{
  "data": {
    "id": "ans_uuid_9",
    "question_id": "q_uuid_10",
    "status": "visible",
    "answerer_type": "member",
    "created_at": "2026-06-02T08:00:00.000Z"
  },
  "message": "Jawaban berhasil dikirim"
}
```

- **Response 201 (manual mode):**
```json
{
  "data": {
    "id": "ans_uuid_9",
    "question_id": "q_uuid_10",
    "status": "pending",
    "answerer_type": "member",
    "created_at": "2026-06-02T08:00:00.000Z"
  },
  "message": "Jawaban dikirim dan sedang menunggu peninjauan moderator"
}
```

- **Response 403** (tidak punya benefit `answer_qna`):
```json
{ "message": "Fitur menjawab pertanyaan memerlukan upgrade plan." }
```

- **Response 409** (sudah pernah menjawab):
```json
{ "message": "Anda sudah menjawab pertanyaan ini. Silakan edit jawaban Anda." }
```

- **Response 400** (pertanyaan ditutup / bukan publik):
```json
{ "message": "Pertanyaan ini tidak menerima jawaban baru" }
```

---

### 12. Edit My Answer

Mengubah jawaban sendiri. Jika mode `manual`, jawaban yang sudah `visible` bisa kembali `pending` setelah diedit (tergantung kebijakan backend).

- **URL:** `PUT /api/v1/mobile/qna/answers/{answer_id}`
- **Auth:** Required (pemilik jawaban)

- **Request Body:**
```json
{
  "answer_text": "Update: tambahan rekomendasi ...",
  "attachment_urls": []
}
```

- **Response 200:**
```json
{
  "data": { "id": "ans_uuid_9", "status": "visible" },
  "message": "Jawaban berhasil diperbarui"
}
```

- **Response 403** (bukan pemilik):
```json
{ "message": "Anda tidak memiliki akses ke jawaban ini" }
```

---

### 13. Upvote Answer

Toggle upvote pada sebuah jawaban. Memanggil endpoint saat sudah upvote akan **membatalkan** upvote.

- **URL:** `POST /api/v1/mobile/qna/answers/{answer_id}/upvote`
- **Auth:** Required (member, bukan pemilik jawaban)

- **Response 200:**
```json
{
  "data": {
    "answer_id": "ans_uuid_1",
    "user_upvoted": true,
    "upvote_count": 13
  },
  "message": "Upvote disimpan"
}
```

- **Response 400** (upvote jawaban sendiri):
```json
{ "message": "Anda tidak bisa memberi upvote pada jawaban sendiri" }
```

---

### 14. Mark Answer Accepted (Expert/Admin)

Menandai (atau membatalkan tanda) sebuah jawaban sebagai rujukan valid. Boleh ada lebih dari satu jawaban valid per pertanyaan. Hanya untuk user dengan permission `validate_qna_answer` (expert yang di-assign pada pertanyaan tsb, atau admin/superadmin).

- **URL:** `POST /api/v1/mobile/qna/answers/{answer_id}/accept`
- **Auth:** Required (permission `validate_qna_answer`)

- **Request Body:**
```json
{
  "is_accepted": true
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `is_accepted` | boolean | **Yes** | `true` = tandai valid, `false` = batalkan |

- **Response 200:**
```json
{
  "data": {
    "answer_id": "ans_uuid_1",
    "is_accepted": true,
    "accepted_by": { "id": "expert_uuid", "name": "Expert Pajak" },
    "question_accepted_count": 1
  },
  "message": "Jawaban ditandai sebagai rujukan valid"
}
```

- **Response 403** (tidak berwenang):
```json
{ "message": "Anda tidak berwenang menandai jawaban pada pertanyaan ini" }
```

---

### 15. Get Questions Assigned To Me

Daftar pertanyaan (publik & privat) yang ditugaskan ke user yang sedang login sebagai penanggung jawab.

- **URL:** `GET /api/v1/mobile/qna/questions/assigned`
- **Auth:** Required (permission `answer_qna`, hanya menampilkan yang di-assign ke user)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `status` | string | No | — | Filter status pertanyaan |
| `limit` | int | No | 20 | — |
| `offset` | int | No | 0 | — |

- **Response 200:**
```json
{
  "data": [
    {
      "id": "q_uuid_22",
      "category": { "id": "cat_uuid_2", "name": "Pajak" },
      "question_text": "Bagaimana pelaporan SPT tahunan untuk WNA?",
      "visibility": "private",
      "status": "answered",
      "answer_count": 1,
      "assigned_at": "2026-06-03T07:00:00.000Z",
      "created_at": "2026-06-02T15:00:00.000Z"
    }
  ],
  "pagination": { "limit": 20, "offset": 0, "total": 1, "has_next": false, "has_prev": false }
}
```

---


| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request — input tidak valid |
| `401` | Unauthorized — token expired / invalid |
| `403` | Forbidden — tidak punya akses atau benefit |
| `404` | Not Found — resource tidak ditemukan |
| `409` | Conflict — sudah pernah menjawab pertanyaan ini |
| `422` | Unprocessable — validation error (termasuk konten ditolak word filter) |
| `500` | Internal Server Error |

---

## Notes & Best Practices

1. **Token Refresh:** Jika mendapat 401 pada endpoint manapun, mobile client harus refresh token via Auth spec, kemudian retry request.

2. **Benefit `ask_qna` / `answer_qna`:** Sebelum menampilkan form submit / tombol jawab, client cek `permissions`/`benefits` dari user object (via `GET /api/v1/auth/session`) untuk menghindari 403.

9. **Forum Publik:** Gunakan `can_answer`, `can_accept`, dan `my_answer_id` dari detail pertanyaan untuk menentukan tombol mana yang ditampilkan, tanpa perlu request tambahan.

10. **Word Filter:** Validasi kata terlarang dijalankan backend (modul System Settings). Client tidak perlu memvalidasi sendiri; cukup tangani respons `422` dan tampilkan pesan error ke user.

11. **Moderasi Jawaban:** Respons `POST answers` mengembalikan `status` (`visible`/`pending`). Jika `pending`, beritahu user bahwa jawaban menunggu peninjauan.

3. **View Count:** `GET /faqs/{faq_id}` selalu increment `view_count`. Jangan panggil endpoint ini saat prefetch / background sync — hanya panggil saat user benar-benar membuka halaman detail.

4. **Search Debounce:** Implementasikan debounce minimal 300ms pada input search sebelum memanggil `GET /search` untuk menghindari request berlebihan.

5. **Vote Toggle:** Endpoint vote bersifat upsert. Jika user vote `true` lalu vote `false`, nilai diperbarui. Client cukup kirim request baru tanpa perlu cek state sebelumnya.

6. **Notifikasi Q&A:** Member menerima push notification saat pertanyaan dijawab (`question_answered`) atau ditolak (`question_rejected`). Deep link notifikasi ke `GET /questions/my/{question_id}`.

7. **Attachment Upload:** Upload lampiran terlebih dahulu ke media upload endpoint sebelum submit pertanyaan. Kirim URL hasil upload ke `attachment_urls`.

8. **Polling Riwayat:** Untuk status `pending`, client bisa refresh `GET /questions/my` setelah menerima push notification, atau tampilkan badge unread pada tab riwayat.

---

## User Flow

```
Buka Modul Q&A
  ↓
GET /categories
  ↓
Pilih kategori → GET /faqs?category_id=
  ↓
Buka FAQ → GET /faqs/{faq_id}   (view_count++)
  ↓
Vote → POST /faqs/{faq_id}/vote
  ↓
Tidak menemukan jawaban?
  ↓
GET /search?q=   (cari dulu)
  ↓
Masih tidak ada? → POST /questions
  ↓
Lihat riwayat → GET /questions/my
  ↓
Notifikasi masuk → GET /questions/my/{question_id}
```

---

*API Specification — QnA Mobile Client. Stack: Flutter + Go Backend.*
