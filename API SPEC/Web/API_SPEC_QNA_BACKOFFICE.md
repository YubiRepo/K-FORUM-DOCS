# API Specification — Q&A / FAQ Module (Backoffice)

> **Dibuat:** 2026-05-30 · **Revisi:** 2026-06-25 (forum komunitas: jawab antar-member, accepted answer, assignment)
> **Stack:** Web Backoffice (Superadmin)
> **Base URL Prefix:** `/api/v1/web/qna`

---

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Data Models](#data-models)
3. [Category Endpoints](#category-endpoints)
   - [B1. List Categories](#b1-list-categories)
   - [B2. Create Category](#b2-create-category)
   - [B3. Update Category](#b3-update-category)
   - [B4. Delete Category](#b4-delete-category)
4. [FAQ Endpoints](#faq-endpoints)
   - [B5. List FAQ Items](#b5-list-faq-items)
   - [B6. Get FAQ Detail](#b6-get-faq-detail)
   - [B7. Create FAQ Item](#b7-create-faq-item)
   - [B8. Update FAQ Item](#b8-update-faq-item)
   - [B9. Delete FAQ Item](#b9-delete-faq-item)
   - [B10. Reorder FAQ Items](#b10-reorder-faq-items)
5. [Question (Incoming) Endpoints](#question-incoming-endpoints)
   - [B11. List Questions](#b11-list-questions)
   - [B12. Get Question Detail](#b12-get-question-detail)
   - [B13. Answer Question](#b13-answer-question)
   - [B14. Reject Question](#b14-reject-question)
   - [B17. Approve Public Question](#b17-approve-public-question)
   - [B18. Assign Question](#b18-assign-question)
   - [B19. Close Question](#b19-close-question)
6. [Answer Moderation Endpoints](#answer-moderation-endpoints)
   - [B20. List Answers (per question)](#b20-list-answers-per-question)
   - [B21. Moderate Answer (approve/reject/hide)](#b21-moderate-answer)
   - [B22. Mark Answer Accepted](#b22-mark-answer-accepted)
7. [Settings Endpoints](#settings-endpoints)
   - [B15. Get Bot Config](#b15-get-bot-config)
   - [B16. Update Bot Config](#b16-update-bot-config)
8. [Status Code Reference](#status-code-reference)
9. [Notification Triggers](#notification-triggers)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/web/qna`
- **Headers Global:**
  ```
  Content-Type: application/json
  Accept: application/json
  Authorization: Bearer <access_token>
  ```

- **Autentikasi & Otorisasi:** Endpoint dikontrol lewat **permission**, bukan sekadar role label:
  - Kategori & FAQ, konfigurasi bot: superadmin (`manage_qna`).
  - Menjawab / approve / reject / assign pertanyaan: `answer_qna` (expert yang di-assign) atau admin/superadmin (`manage_qna`).
  - Menandai jawaban valid: `validate_qna_answer`.
  - Request tanpa permission yang sesuai → `403`.

> Word filter untuk konten pertanyaan & jawaban dijalankan di service layer backend memakai daftar kata dari modul **System Settings** — bukan tanggung jawab modul Q&A.

---

## Data Models

### CategoryAdminObject
```json
{
  "id": "uuid",
  "name": "Imigrasi & Visa",
  "slug": "imigrasi-visa",
  "description": "Informasi seputar visa, izin tinggal, dan imigrasi di Indonesia",
  "icon": "ti-passport",
  "color": "#534AB7",
  "sort_order": 1,
  "is_active": true,
  "faq_count": 12,
  "created_by": {
    "id": "uuid",
    "name": "Admin KAI"
  },
  "created_at": "2026-01-01T00:00:00.000Z",
  "updated_at": "2026-01-01T00:00:00.000Z"
}
```

### FaqAdminListObject
```json
{
  "id": "uuid",
  "category": {
    "id": "uuid",
    "name": "Pajak"
  },
  "question": "Bagaimana cara mendaftar NPWP sebagai WNA?",
  "status": "published",
  "is_pinned": true,
  "sort_order": 1,
  "view_count": 342,
  "helpful_count": 87,
  "not_helpful_count": 3,
  "created_by": {
    "id": "uuid",
    "name": "Admin KAI"
  },
  "published_at": "2026-01-10T08:00:00.000Z",
  "updated_at": "2026-04-15T12:30:00.000Z"
}
```

### FaqAdminDetailObject
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
  "status": "published",
  "is_pinned": true,
  "sort_order": 1,
  "view_count": 342,
  "helpful_count": 87,
  "not_helpful_count": 3,
  "created_by": {
    "id": "uuid",
    "name": "Admin KAI"
  },
  "updated_by": {
    "id": "uuid",
    "name": "Admin KAI"
  },
  "published_at": "2026-01-10T08:00:00.000Z",
  "archived_at": null,
  "created_at": "2026-01-05T10:00:00.000Z",
  "updated_at": "2026-04-15T12:30:00.000Z"
}
```

### QuestionAdminListObject
```json
{
  "id": "uuid",
  "user": {
    "id": "uuid",
    "name": "Kim Minji",
    "email": "minji@example.com",
    "avatar": "https://cdn.example.com/avatar.jpg"
  },
  "category": {
    "id": "uuid",
    "name": "Properti & Kepemilikan"
  },
  "question_text": "Apakah WNA bisa memiliki properti di Indonesia?",
  "visibility": "public",
  "status": "pending",
  "assigned_to": {
    "id": "uuid",
    "name": "Expert Pajak"
  },
  "answer_count": 0,
  "accepted_count": 0,
  "bot_responded": true,
  "created_at": "2026-05-28T09:00:00.000Z"
}
```

> `assigned_to` bernilai `null` jika belum ditugaskan.

### QuestionAdminDetailObject
```json
{
  "id": "uuid",
  "user": {
    "id": "uuid",
    "name": "Kim Minji",
    "email": "minji@example.com",
    "avatar": "https://cdn.example.com/avatar.jpg",
    "plan": "standard"
  },
  "category": {
    "id": "uuid",
    "name": "Properti & Kepemilikan"
  },
  "question_text": "Apakah WNA bisa memiliki properti di Indonesia?",
  "attachment_urls": [],
  "visibility": "public",
  "status": "approved",
  "assigned_to": { "id": "uuid", "name": "Expert Pajak" },
  "assigned_by": { "id": "uuid", "name": "Admin KAI" },
  "assigned_at": "2026-05-28T10:00:00.000Z",
  "approved_by": { "id": "uuid", "name": "Admin KAI" },
  "approved_at": "2026-05-28T09:30:00.000Z",
  "answer_count": 2,
  "accepted_count": 1,
  "bot_responded": true,
  "bot_response": "Terima kasih, pertanyaan Anda sedang diproses oleh tim kami.",
  "answers": [
    {
      "id": "ans_uuid_1",
      "answer_text": "WNA dapat memiliki properti melalui skema Hak Pakai ...",
      "answered_by": { "id": "uuid", "name": "Kang Dae" },
      "answerer_type": "member",
      "status": "visible",
      "is_accepted": true,
      "accepted_by": { "id": "uuid", "name": "Admin KAI" },
      "upvote_count": 12,
      "converted_to_faq_id": null,
      "created_at": "2026-05-28T11:00:00.000Z"
    }
  ],
  "rejection_reason": null,
  "created_at": "2026-05-28T09:00:00.000Z",
  "answered_at": "2026-05-28T11:00:00.000Z"
}
```

> Untuk pertanyaan `private`, `answers` umumnya berisi jawaban dari admin/expert saja. Untuk `public`, berisi semua jawaban termasuk yang `pending` (agar moderator bisa meninjau).

### BotConfigObject
```json
{
  "enabled": true,
  "fallback_message": "Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja.",
  "suggest_contact": true,
  "contact_info": "admin@kai-indonesia.org",
  "similarity_threshold": 0.60,
  "answer_moderation_mode": "manual",
  "question_moderation_mode": "manual",
  "updated_by": {
    "id": "uuid",
    "name": "Admin KAI"
  },
  "updated_at": "2026-05-01T08:00:00.000Z"
}
```

### PaginationObject
```json
{
  "limit": 20,
  "offset": 0,
  "total": 45,
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

// 409 Conflict
{
  "message": "Slug sudah digunakan oleh kategori lain"
}

// 422 Validation Error
{
  "message": "Data input tidak valid",
  "errors": {
    "name": ["Nama kategori tidak boleh kosong"],
    "slug": ["Slug hanya boleh berisi huruf kecil, angka, dan tanda hubung"]
  }
}
```

---

## Category Endpoints

### B1. List Categories

Ambil semua kategori. Superadmin bisa melihat kategori nonaktif juga via query param.

- **URL:** `GET /api/v1/web/qna/categories`
- **Auth:** Yes (Superadmin)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `include_inactive` | boolean | No | false | Tampilkan kategori nonaktif juga |

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
      "is_active": true,
      "faq_count": 12,
      "created_by": {
        "id": "uuid",
        "name": "Admin KAI"
      },
      "created_at": "2026-01-01T00:00:00.000Z",
      "updated_at": "2026-01-01T00:00:00.000Z"
    },
    {
      "id": "cat_uuid_7",
      "name": "Investasi",
      "slug": "investasi",
      "description": null,
      "icon": "ti-chart-bar",
      "color": "#185FA5",
      "sort_order": 8,
      "is_active": false,
      "faq_count": 0,
      "created_by": {
        "id": "uuid",
        "name": "Admin KAI"
      },
      "created_at": "2026-03-10T00:00:00.000Z",
      "updated_at": "2026-04-01T00:00:00.000Z"
    }
  ]
}
```

---

### B2. Create Category

- **URL:** `POST /api/v1/web/qna/categories`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "name": "Ketenagakerjaan",
  "slug": "ketenagakerjaan",
  "description": "BPJS, izin kerja, dan peraturan ketenagakerjaan di Indonesia",
  "icon": "ti-briefcase",
  "color": "#993C1D",
  "sort_order": 3
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `name` | string | **Yes** | Nama kategori. Max: 100 char |
| `slug` | string | **Yes** | URL-friendly, lowercase, unique. Contoh: `ketenagakerjaan` |
| `description` | string | No | Deskripsi singkat. Max: 500 char |
| `icon` | string | No | Nama icon Tabler (contoh: `ti-briefcase`) |
| `color` | string | No | Hex color (contoh: `#993C1D`) |
| `sort_order` | int | No | Urutan tampil. Default: 0 |

- **Response 201:**
```json
{
  "data": {
    "id": "cat_uuid_3",
    "name": "Ketenagakerjaan",
    "slug": "ketenagakerjaan",
    "description": "BPJS, izin kerja, dan peraturan ketenagakerjaan di Indonesia",
    "icon": "ti-briefcase",
    "color": "#993C1D",
    "sort_order": 3,
    "is_active": true,
    "faq_count": 0,
    "created_by": { "id": "admin_uuid", "name": "Admin KAI" },
    "created_at": "2026-05-30T10:00:00.000Z",
    "updated_at": "2026-05-30T10:00:00.000Z"
  },
  "message": "Kategori berhasil dibuat"
}
```

- **Response 409** (slug sudah dipakai):
```json
{ "message": "Slug sudah digunakan oleh kategori lain" }
```

---

### B3. Update Category

Semua field opsional — hanya field yang dikirim yang akan diperbarui.

- **URL:** `PUT /api/v1/web/qna/categories/{category_id}`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "name": "Ketenagakerjaan & BPJS",
  "description": "Panduan BPJS Kesehatan, BPJS Ketenagakerjaan, dan izin kerja WNA",
  "is_active": true,
  "sort_order": 3
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `name` | string | No | Max: 100 char |
| `slug` | string | No | Harus tetap unique |
| `description` | string | No | Max: 500 char |
| `icon` | string | No | Nama icon Tabler |
| `color` | string | No | Hex color |
| `sort_order` | int | No | — |
| `is_active` | boolean | No | Toggle aktif/nonaktif |

- **Response 200:**
```json
{
  "data": { ...CategoryAdminObject },
  "message": "Kategori berhasil diupdate"
}
```

- **Response 404:**
```json
{ "message": "Kategori tidak ditemukan" }
```

---

### B4. Delete Category

Hanya bisa dihapus jika tidak ada FAQ published atau pertanyaan pending di dalamnya.

- **URL:** `DELETE /api/v1/web/qna/categories/{category_id}`
- **Auth:** Yes (Superadmin)

- **Response 200:**
```json
{ "message": "Kategori berhasil dihapus" }
```

- **Response 409** (masih ada FAQ aktif):
```json
{ "message": "Kategori tidak bisa dihapus karena masih memiliki 12 FAQ aktif. Arsipkan atau pindahkan FAQ terlebih dahulu." }
```

---

## FAQ Endpoints

### B5. List FAQ Items

Daftar semua FAQ item dengan filtering dan pagination.

- **URL:** `GET /api/v1/web/qna/faqs`
- **Auth:** Yes (Superadmin)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `category_id` | UUID | No | — | Filter per kategori |
| `status` | string | No | — | `draft`, `published`, `archived` |
| `search` | string | No | — | Cari di teks question / answer |
| `sort` | string | No | `-created_at` | `created_at`, `-created_at`, `view_count`, `-view_count`, `helpful_count`, `-helpful_count` |
| `limit` | int | No | 20 | Max: 100 |
| `offset` | int | No | 0 | — |

- **Response 200:**
```json
{
  "data": [
    {
      "id": "faq_uuid_1",
      "category": { "id": "uuid", "name": "Pajak" },
      "question": "Bagaimana cara mendaftar NPWP sebagai WNA?",
      "status": "published",
      "is_pinned": true,
      "sort_order": 1,
      "view_count": 342,
      "helpful_count": 87,
      "not_helpful_count": 3,
      "created_by": { "id": "uuid", "name": "Admin KAI" },
      "published_at": "2026-01-10T08:00:00.000Z",
      "updated_at": "2026-04-15T12:30:00.000Z"
    },
    {
      "id": "faq_uuid_6",
      "category": { "id": "uuid", "name": "Imigrasi & Visa" },
      "question": "Apa syarat perpanjangan KITAS?",
      "status": "draft",
      "is_pinned": false,
      "sort_order": 0,
      "view_count": 0,
      "helpful_count": 0,
      "not_helpful_count": 0,
      "created_by": { "id": "uuid", "name": "Admin KAI" },
      "published_at": null,
      "updated_at": "2026-05-29T08:00:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 45,
    "has_next": true,
    "has_prev": false
  }
}
```

---

### B6. Get FAQ Detail

- **URL:** `GET /api/v1/web/qna/faqs/{faq_id}`
- **Auth:** Yes (Superadmin)

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
    "answer": "WNA yang bekerja atau memiliki penghasilan di Indonesia wajib memiliki NPWP...",
    "tags": ["npwp", "wna", "pajak", "kitas"],
    "status": "published",
    "is_pinned": true,
    "sort_order": 1,
    "view_count": 342,
    "helpful_count": 87,
    "not_helpful_count": 3,
    "created_by": { "id": "uuid", "name": "Admin KAI" },
    "updated_by": { "id": "uuid", "name": "Admin KAI" },
    "published_at": "2026-01-10T08:00:00.000Z",
    "archived_at": null,
    "created_at": "2026-01-05T10:00:00.000Z",
    "updated_at": "2026-04-15T12:30:00.000Z"
  }
}
```

---

### B7. Create FAQ Item

- **URL:** `POST /api/v1/web/qna/faqs`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "category_id": "cat_uuid_2",
  "question": "Apa perbedaan KITAS dan KITAP?",
  "answer": "KITAS (Kartu Izin Tinggal Terbatas) adalah izin tinggal sementara bagi WNA...\n\nKITAP (Kartu Izin Tinggal Tetap) adalah izin tinggal permanen bagi WNA yang sudah...",
  "tags": ["kitas", "kitap", "imigrasi", "izin tinggal"],
  "status": "published",
  "is_pinned": false,
  "sort_order": 5
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `category_id` | UUID | **Yes** | ID kategori yang valid dan aktif |
| `question` | string | **Yes** | Teks pertanyaan. Max: 500 char |
| `answer` | string | **Yes** | Teks jawaban. Max: 10.000 char. Support Markdown |
| `tags` | string[] | No | Tag pencarian. Max: 10 tag, masing-masing max 50 char |
| `status` | string | No | `draft` (default) atau `published` |
| `is_pinned` | boolean | No | Default: false |
| `sort_order` | int | No | Default: 0 |

- **Response 201:**
```json
{
  "data": { ...FaqAdminDetailObject },
  "message": "FAQ berhasil dibuat"
}
```

- **Response 422:**
```json
{
  "message": "Data input tidak valid",
  "errors": {
    "question": ["Pertanyaan tidak boleh kosong"],
    "category_id": ["Kategori tidak ditemukan atau tidak aktif"]
  }
}
```

---

### B8. Update FAQ Item

Semua field opsional — hanya field yang dikirim yang akan diperbarui.

- **URL:** `PUT /api/v1/web/qna/faqs/{faq_id}`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "question": "Apa perbedaan KITAS dan KITAP? (Diperbarui 2026)",
  "answer": "...",
  "tags": ["kitas", "kitap", "imigrasi"],
  "status": "published",
  "is_pinned": true,
  "sort_order": 2
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `category_id` | UUID | No | Pindah kategori |
| `question` | string | No | Max: 500 char |
| `answer` | string | No | Max: 10.000 char |
| `tags` | string[] | No | Menggantikan seluruh tags lama |
| `status` | string | No | `draft`, `published`, atau `archived` |
| `is_pinned` | boolean | No | — |
| `sort_order` | int | No | — |

- **Response 200:**
```json
{
  "data": { ...FaqAdminDetailObject },
  "message": "FAQ berhasil diupdate"
}
```

> **Note:** Mengubah `status` ke `published` akan otomatis mengisi `published_at` jika sebelumnya null. Mengubah ke `archived` akan mengisi `archived_at`.

---

### B9. Delete FAQ Item

Hard delete. FAQ yang sudah dikonversi dari pertanyaan member tidak bisa dihapus langsung — arsipkan dulu.

- **URL:** `DELETE /api/v1/web/qna/faqs/{faq_id}`
- **Auth:** Yes (Superadmin)

- **Response 200:**
```json
{ "message": "FAQ berhasil dihapus" }
```

- **Response 404:**
```json
{ "message": "FAQ tidak ditemukan" }
```

---

### B10. Reorder FAQ Items

Atur ulang urutan tampil FAQ di dalam satu kategori. Kirim array `ordered_ids` sesuai urutan yang diinginkan.

- **URL:** `PUT /api/v1/web/qna/faqs/reorder`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "category_id": "cat_uuid_2",
  "ordered_ids": [
    "faq_uuid_3",
    "faq_uuid_1",
    "faq_uuid_5",
    "faq_uuid_2"
  ]
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `category_id` | UUID | **Yes** | Kategori yang akan direorder |
| `ordered_ids` | string[] | **Yes** | Array UUID FAQ sesuai urutan baru. Semua FAQ dalam kategori harus disertakan |

- **Response 200:**
```json
{ "message": "Urutan FAQ berhasil diperbarui" }
```

- **Response 422** (ID tidak lengkap):
```json
{ "message": "ordered_ids harus berisi semua FAQ dalam kategori ini (4 item diharapkan, 3 diterima)" }
```

---

## Question (Incoming) Endpoints

### B11. List Questions

Daftar pertanyaan masuk dari member. Default menampilkan antrian `pending` dari yang paling lama (oldest first).

- **URL:** `GET /api/v1/web/qna/questions`
- **Auth:** Yes (Superadmin)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `status` | string | No | `pending` | `pending`, `approved`, `answered`, `rejected`, `converted`, `closed`, atau `all` |
| `visibility` | string | No | — | `private`, `public` |
| `assigned_to` | UUID | No | — | Filter pertanyaan yang ditugaskan ke user tertentu. `me` = ditugaskan ke saya |
| `category_id` | UUID | No | — | Filter per kategori |
| `search` | string | No | — | Cari di teks pertanyaan atau nama/email user |
| `sort` | string | No | `created_at` | `created_at` (asc), `-created_at` (desc) |
| `limit` | int | No | 20 | Max: 100 |
| `offset` | int | No | 0 | — |

- **Response 200:**
```json
{
  "data": [
    {
      "id": "q_uuid_1",
      "user": {
        "id": "user_uuid_1",
        "name": "Kim Minji",
        "email": "minji@example.com",
        "avatar": "https://cdn.example.com/avatar.jpg"
      },
      "category": {
        "id": "cat_uuid_1",
        "name": "Properti & Kepemilikan"
      },
      "question_text": "Apakah WNA bisa memiliki properti di Indonesia?",
      "status": "pending",
      "bot_responded": true,
      "created_at": "2026-05-26T09:00:00.000Z"
    },
    {
      "id": "q_uuid_2",
      "user": {
        "id": "user_uuid_2",
        "name": "Park Joon",
        "email": "joon@example.com",
        "avatar": null
      },
      "category": null,
      "question_text": "Bagaimana cara mendaftar BPJS Kesehatan untuk WNA?",
      "status": "pending",
      "bot_responded": true,
      "created_at": "2026-05-27T14:30:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 7,
    "has_next": false,
    "has_prev": false
  },
  "summary": {
    "total_pending": 7,
    "total_answered_today": 3,
    "total_rejected_today": 1
  }
}
```

---

### B12. Get Question Detail

- **URL:** `GET /api/v1/web/qna/questions/{question_id}`
- **Auth:** Yes (Superadmin)

- **Response 200:**
```json
{
  "data": {
    "id": "q_uuid_1",
    "user": {
      "id": "user_uuid_1",
      "name": "Kim Minji",
      "email": "minji@example.com",
      "avatar": "https://cdn.example.com/avatar.jpg",
      "plan": "standard"
    },
    "category": {
      "id": "cat_uuid_1",
      "name": "Properti & Kepemilikan"
    },
    "question_text": "Apakah WNA bisa memiliki properti di Indonesia?",
    "attachment_urls": [],
    "status": "pending",
    "bot_responded": true,
    "bot_response": "Terima kasih, pertanyaan Anda sedang diproses oleh tim kami.",
    "answer": null,
    "rejection_reason": null,
    "created_at": "2026-05-26T09:00:00.000Z",
    "answered_at": null
  }
}
```

---

### B13. Answer Question

Admin menjawab pertanyaan member. Otomatis trigger notifikasi push ke member. Opsional: konversi jawaban menjadi FAQ publik baru dalam satu request.

> Jawaban yang dibuat lewat endpoint ini otomatis `answerer_type` = `admin` (atau `expert` bila pemanggil adalah expert yang di-assign) dan langsung berstatus `visible`. Untuk **menandai** jawaban (milik siapa pun) sebagai rujukan valid, gunakan B22. Endpoint ini dipakai terutama untuk alur **privat** dan untuk admin yang menjawab langsung; pada forum publik, member menjawab lewat endpoint mobile.

- **URL:** `POST /api/v1/web/qna/questions/{question_id}/answer`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "answer_text": "WNA dengan KITAS dapat memiliki properti melalui skema Hak Pakai (HP). Berikut ketentuan lengkapnya:\n\n1. Minimal masa tinggal 1 tahun\n2. Nilai properti sesuai batas yang ditetapkan pemerintah daerah\n3. Tidak bisa dijual kepada WNA lain tanpa izin khusus\n\nReferensi: PP No. 18 Tahun 2021",
  "attachment_urls": ["s3:/qna/attachments/uuid.pdf"],
  "is_public": true,
  "convert_to_faq": true,
  "faq_category_id": "cat_uuid_1"
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `answer_text` | string | **Yes** | Teks jawaban. Max: 10.000 char. Support Markdown |
| `attachment_urls` | string[] | No | Path lampiran (format `s3:/path` untuk internal, `ext:https://...` untuk eksternal). Max: 3 item |
| `is_public` | boolean | No | Default: `true`. `false` = jawaban hanya terlihat oleh member yang bertanya |
| `convert_to_faq` | boolean | No | Default: `false`. `true` = jawaban otomatis dibuat sebagai FAQ publik baru |
| `faq_category_id` | UUID | Conditional | **Wajib** jika `convert_to_faq: true`. Kategori FAQ yang akan dibuat |

- **Response 201** (tanpa convert):
```json
{
  "data": {
    "answer_id": "ans_uuid_1",
    "question_id": "q_uuid_1",
    "question_status": "answered",
    "converted_to_faq_id": null,
    "notification_sent": true
  },
  "message": "Pertanyaan berhasil dijawab dan notifikasi telah dikirim"
}
```

- **Response 201** (dengan `convert_to_faq: true`):
```json
{
  "data": {
    "answer_id": "ans_uuid_1",
    "question_id": "q_uuid_1",
    "question_status": "converted",
    "converted_to_faq_id": "faq_uuid_new",
    "notification_sent": true
  },
  "message": "Pertanyaan dijawab dan telah dikonversi menjadi FAQ publik"
}
```

- **Response 400** (pertanyaan sudah dijawab):
```json
{ "message": "Pertanyaan ini sudah dijawab sebelumnya" }
```

- **Response 422** (`convert_to_faq: true` tapi `faq_category_id` kosong):
```json
{
  "message": "Data input tidak valid",
  "errors": {
    "faq_category_id": ["Kategori FAQ wajib diisi jika convert_to_faq diaktifkan"]
  }
}
```

---

### B14. Reject Question

Tolak pertanyaan yang tidak relevan, duplikat, atau tidak sesuai dengan scope modul. Member menerima notifikasi beserta alasan penolakan.

- **URL:** `PUT /api/v1/web/qna/questions/{question_id}/reject`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "rejection_reason": "Pertanyaan ini sudah terjawab di FAQ: 'Bagaimana cara mendaftar NPWP sebagai WNA?'. Silakan cek di kategori Pajak."
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `rejection_reason` | string | **Yes** | Alasan penolakan (akan dikirim ke member). Max: 500 char |

- **Response 200:**
```json
{
  "data": {
    "question_id": "q_uuid_1",
    "status": "rejected",
    "notification_sent": true
  },
  "message": "Pertanyaan berhasil ditolak"
}
```

- **Response 400** (pertanyaan bukan dalam status pending):
```json
{ "message": "Hanya pertanyaan berstatus 'pending' yang bisa ditolak" }
```

---

### B17. Approve Public Question

Menyetujui pertanyaan **publik** berstatus `pending` agar tayang di forum komunitas. Hanya berlaku untuk `visibility=public`. Pertanyaan privat tidak perlu di-approve — admin langsung menjawab via B13.

- **URL:** `POST /api/v1/web/qna/questions/{question_id}/approve`
- **Auth:** `manage_qna` (admin/superadmin)

- **Response 200:**
```json
{
  "data": {
    "question_id": "q_uuid_10",
    "status": "approved",
    "notification_sent": true
  },
  "message": "Pertanyaan disetujui dan kini tayang di forum"
}
```

- **Response 400** (bukan publik / bukan pending):
```json
{ "message": "Hanya pertanyaan publik berstatus 'pending' yang bisa disetujui" }
```

---

### B18. Assign Question

Menugaskan pertanyaan (publik maupun privat) ke seorang expert/staf sebagai penanggung jawab. Kirim `assigned_to: null` untuk mencabut penugasan.

- **URL:** `PUT /api/v1/web/qna/questions/{question_id}/assign`
- **Auth:** `manage_qna` (admin/superadmin)

- **Request Body:**
```json
{
  "assigned_to": "expert_uuid"
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `assigned_to` | UUID \| null | **Yes** | User yang ditugaskan. `null` = cabut penugasan. User harus punya permission `answer_qna` |

- **Response 200:**
```json
{
  "data": {
    "question_id": "q_uuid_22",
    "assigned_to": { "id": "expert_uuid", "name": "Expert Pajak" },
    "assigned_at": "2026-06-03T07:00:00.000Z",
    "notification_sent": true
  },
  "message": "Pertanyaan berhasil ditugaskan"
}
```

- **Response 422** (user tidak punya permission `answer_qna`):
```json
{
  "message": "Data input tidak valid",
  "errors": { "assigned_to": ["User tidak memiliki hak untuk menjawab pertanyaan"] }
}
```

---

### B19. Close Question

Menutup pertanyaan publik agar tidak menerima jawaban baru. Isi & jawaban tetap bisa dibaca.

- **URL:** `POST /api/v1/web/qna/questions/{question_id}/close`
- **Auth:** `manage_qna` (admin/superadmin)

- **Response 200:**
```json
{
  "data": { "question_id": "q_uuid_10", "status": "closed" },
  "message": "Pertanyaan ditutup"
}
```

---

## Answer Moderation Endpoints

### B20. List Answers (per question)

Daftar semua jawaban untuk satu pertanyaan, termasuk yang `pending` dan `hidden` (untuk keperluan moderasi).

- **URL:** `GET /api/v1/web/qna/questions/{question_id}/answers`
- **Auth:** `answer_qna` (expert assigned) atau `manage_qna` (admin)
- **Query Params:**

| Param | Type | Required | Default | Keterangan |
|-------|------|----------|---------|-----------|
| `status` | string | No | `all` | `visible`, `pending`, `rejected`, `hidden`, `all` |

- **Response 200:**
```json
{
  "data": [
    {
      "id": "ans_uuid_1",
      "answer_text": "WNA dapat memiliki properti melalui skema Hak Pakai ...",
      "answered_by": { "id": "uuid", "name": "Kang Dae", "plan": "pro" },
      "answerer_type": "member",
      "status": "visible",
      "is_accepted": true,
      "accepted_by": { "id": "uuid", "name": "Admin KAI" },
      "upvote_count": 12,
      "converted_to_faq_id": null,
      "created_at": "2026-06-01T11:00:00.000Z"
    },
    {
      "id": "ans_uuid_3",
      "answer_text": "Setahu saya gak boleh sama sekali ...",
      "answered_by": { "id": "uuid", "name": "Anon Member", "plan": "standard" },
      "answerer_type": "member",
      "status": "pending",
      "is_accepted": false,
      "accepted_by": null,
      "upvote_count": 0,
      "converted_to_faq_id": null,
      "created_at": "2026-06-02T08:00:00.000Z"
    }
  ]
}
```

---

### B21. Moderate Answer

Approve, tolak, atau sembunyikan jawaban member. Dipakai pada mode `manual` (approve jawaban `pending`) atau untuk menyembunyikan jawaban yang sudah tayang (mis. hasil report).

- **URL:** `PATCH /api/v1/web/qna/answers/{answer_id}/moderate`
- **Auth:** `answer_qna` (expert assigned pada pertanyaan tsb) atau `manage_qna` (admin)

- **Request Body:**
```json
{
  "action": "approve",
  "reason": null
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `action` | string | **Yes** | `approve` (→ visible), `reject` (→ rejected), `hide` (→ hidden) |
| `reason` | string | Conditional | Wajib untuk `reject` / `hide`. Max: 500 char |

- **Response 200:**
```json
{
  "data": { "answer_id": "ans_uuid_3", "status": "visible" },
  "message": "Jawaban disetujui dan kini tampil di forum"
}
```

- **Response 422** (`reject`/`hide` tanpa reason):
```json
{
  "message": "Data input tidak valid",
  "errors": { "reason": ["Alasan wajib diisi untuk aksi reject/hide"] }
}
```

---

### B22. Mark Answer Accepted

Menandai (atau membatalkan) jawaban sebagai rujukan valid. Boleh lebih dari satu per pertanyaan. Identik dengan endpoint mobile `accept`, disediakan juga di backoffice.

- **URL:** `POST /api/v1/web/qna/answers/{answer_id}/accept`
- **Auth:** `validate_qna_answer` (expert assigned atau admin/superadmin)

- **Request Body:**
```json
{
  "is_accepted": true
}
```

- **Response 200:**
```json
{
  "data": {
    "answer_id": "ans_uuid_1",
    "is_accepted": true,
    "accepted_by": { "id": "admin_uuid", "name": "Admin KAI" },
    "question_accepted_count": 1
  },
  "message": "Jawaban ditandai sebagai rujukan valid"
}
```

---

## Settings Endpoints

### B15. Get Bot Config

Ambil konfigurasi bot fallback yang sedang aktif.

- **URL:** `GET /api/v1/web/qna/settings/bot`
- **Auth:** Yes (Superadmin)

- **Response 200:**
```json
{
  "data": {
    "enabled": true,
    "fallback_message": "Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja.",
    "suggest_contact": true,
    "contact_info": "admin@kai-indonesia.org",
    "similarity_threshold": 0.60,
    "updated_by": {
      "id": "admin_uuid",
      "name": "Admin KAI"
    },
    "updated_at": "2026-05-01T08:00:00.000Z"
  }
}
```

---

### B16. Update Bot Config

Perbarui konfigurasi bot fallback. Semua field opsional — hanya yang dikirim yang akan diperbarui.

- **URL:** `PUT /api/v1/web/qna/settings/bot`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "enabled": true,
  "fallback_message": "Terima kasih sudah bertanya! Admin kami akan merespons dalam 1 hari kerja.",
  "suggest_contact": true,
  "contact_info": "admin@kai-indonesia.org | WA: 0812xxxxxxxx",
  "similarity_threshold": 0.70,
  "answer_moderation_mode": "manual",
  "question_moderation_mode": "manual"
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `enabled` | boolean | No | Toggle bot aktif/nonaktif |
| `fallback_message` | string | No | Pesan otomatis ke member saat pertanyaan masuk. Max: 500 char |
| `suggest_contact` | boolean | No | Tampilkan info kontak di dalam bot response |
| `contact_info` | string | No | Info kontak yang ditampilkan (email, nomor WA, dll.) Max: 300 char |
| `similarity_threshold` | float (0.0–1.0) | No | Ambang batas kemiripan dengan FAQ. Lebih tinggi = lebih ketat. Default: 0.60 |
| `answer_moderation_mode` | string | No | `auto` (jawaban langsung tayang) atau `manual` (approve dulu). Default: `manual` |
| `question_moderation_mode` | string | No | Mode moderasi pertanyaan publik. Default: `manual` |

- **Response 200:**
```json
{
  "data": { ...BotConfigObject },
  "message": "Konfigurasi bot berhasil diperbarui"
}
```

---

## Status Code Reference

| Code | Meaning |
|------|---------|
| `200` | Success |
| `201` | Created |
| `400` | Bad Request — kondisi bisnis tidak terpenuhi |
| `401` | Unauthorized — token expired / invalid |
| `403` | Forbidden — bukan superadmin |
| `404` | Not Found — resource tidak ditemukan |
| `409` | Conflict — duplikasi data (slug, dll.) |
| `422` | Unprocessable — validation error |
| `500` | Internal Server Error |

---

## Notification Triggers

Semua notifikasi push dikirim via FCM ke member yang bersangkutan.

| Event | Endpoint | Notifikasi ke | Sub-type |
|-------|----------|---------------|----------|
| Pertanyaan publik disetujui | `POST /questions/:id/approve` | Penanya | `question_approved` |
| Admin/expert/member menjawab | `POST /questions/:id/answer` & mobile `POST /questions/:id/answers` | Penanya | `question_answered` |
| Admin menolak pertanyaan | `PUT /questions/:id/reject` | Penanya | `question_rejected` |
| Jawaban ditandai valid | `POST /answers/:id/accept` | Pemilik jawaban | `answer_accepted` |
| Pertanyaan ditugaskan | `PUT /questions/:id/assign` | Expert yang ditugaskan | `question_assigned` |
| Jawaban ditolak moderator | `PATCH /answers/:id/moderate` (reject) | Pemilik jawaban | `answer_rejected` |

**Notification Payload:**
```json
// question_answered
{
  "type": "qna",
  "sub_type": "question_answered",
  "title": "Pertanyaan Anda telah dijawab",
  "body": "Admin KAI telah menjawab pertanyaan Anda.",
  "data": {
    "question_id": "q_uuid_1",
    "answer_id": "ans_uuid_1"
  }
}

// question_rejected
{
  "type": "qna",
  "sub_type": "question_rejected",
  "title": "Pertanyaan Anda tidak dapat diproses",
  "body": "Pertanyaan ini sudah terjawab di FAQ. Silakan cek di kategori Pajak.",
  "data": {
    "question_id": "q_uuid_1"
  }
}
```

---

## Admin Flow

```
KELOLA KATEGORI
  GET  /categories                    → lihat semua kategori (+ inactive)
  POST /categories                    → buat kategori baru
  PUT  /categories/{id}               → edit kategori
  DELETE /categories/{id}             → hapus (jika tidak ada FAQ aktif)

KELOLA FAQ
  GET  /faqs                          → daftar FAQ dengan filter & search
  GET  /faqs/{id}                     → detail FAQ
  POST /faqs                          → buat FAQ baru
  PUT  /faqs/{id}                     → edit FAQ
  DELETE /faqs/{id}                   → hapus FAQ
  PUT  /faqs/reorder                  → atur urutan

ANTRIAN PERTANYAAN
  GET  /questions                     → antrian (filter status/visibility/assigned_to)
  GET  /questions/{id}                → detail pertanyaan + daftar jawaban
  POST /questions/{id}/approve        → setujui pertanyaan publik → tayang di forum (notif)
  PUT  /questions/{id}/assign         → tugaskan ke expert (notif expert)
  POST /questions/{id}/answer         → admin/expert jawab langsung (notif)
  PUT  /questions/{id}/reject         → tolak (notif)
  POST /questions/{id}/close          → tutup forum pertanyaan

MODERASI JAWABAN
  GET   /questions/{id}/answers       → daftar jawaban (termasuk pending/hidden)
  PATCH /answers/{id}/moderate        → approve / reject / hide jawaban
  POST  /answers/{id}/accept          → tandai jawaban sbg rujukan valid (notif)

MEDIA / LAMPIRAN
  POST   /attachments/presign             → dapatkan presigned URL untuk upload lampiran
  POST   /attachments/confirm             → konfirmasi upload selesai
  DELETE /attachments/{attachment_id}     → hapus lampiran

PENGATURAN
  GET  /settings/bot                  → lihat konfigurasi bot + mode moderasi
  PUT  /settings/bot                  → update konfigurasi (termasuk answer_moderation_mode)
```

---

*API Specification — QnA Backoffice. Stack: Web Admin Panel + Go Backend.*
