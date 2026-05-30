# API Specification — Q&A / FAQ Module (Backoffice)

> **Dibuat:** 2026-05-30
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
6. [Settings Endpoints](#settings-endpoints)
   - [B15. Get Bot Config](#b15-get-bot-config)
   - [B16. Update Bot Config](#b16-update-bot-config)
7. [Status Code Reference](#status-code-reference)
8. [Notification Triggers](#notification-triggers)

---

## Informasi Umum

- **Base URL Prefix:** `/api/v1/web/qna`
- **Headers Global:**
  ```
  Content-Type: application/json
  Accept: application/json
  Authorization: Bearer <access_token>
  ```

- **Autentikasi:** Semua endpoint di bawah ini **hanya bisa diakses oleh Superadmin**.
  - Request tanpa token atau role bukan superadmin → `401` / `403`

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
  "status": "pending",
  "bot_responded": true,
  "created_at": "2026-05-28T09:00:00.000Z"
}
```

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
  "status": "pending",
  "bot_responded": true,
  "bot_response": "Terima kasih, pertanyaan Anda sedang diproses oleh tim kami.",
  "answer": null,
  "rejection_reason": null,
  "created_at": "2026-05-28T09:00:00.000Z",
  "answered_at": null
}
```

### BotConfigObject
```json
{
  "enabled": true,
  "fallback_message": "Terima kasih atas pertanyaan Anda. Tim kami akan segera merespons dalam 1–2 hari kerja.",
  "suggest_contact": true,
  "contact_info": "admin@kai-indonesia.org",
  "similarity_threshold": 0.60,
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
| `status` | string | No | `pending` | `pending`, `answered`, `rejected`, `converted`, atau `all` |
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

- **URL:** `POST /api/v1/web/qna/questions/{question_id}/answer`
- **Auth:** Yes (Superadmin)

- **Request Body:**
```json
{
  "answer_text": "WNA dengan KITAS dapat memiliki properti melalui skema Hak Pakai (HP). Berikut ketentuan lengkapnya:\n\n1. Minimal masa tinggal 1 tahun\n2. Nilai properti sesuai batas yang ditetapkan pemerintah daerah\n3. Tidak bisa dijual kepada WNA lain tanpa izin khusus\n\nReferensi: PP No. 18 Tahun 2021",
  "attachment_urls": [],
  "is_public": true,
  "convert_to_faq": true,
  "faq_category_id": "cat_uuid_1"
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `answer_text` | string | **Yes** | Teks jawaban. Max: 10.000 char. Support Markdown |
| `attachment_urls` | string[] | No | URL lampiran. Max: 3 item |
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
  "similarity_threshold": 0.70
}
```

| Field | Type | Required | Keterangan |
|-------|------|----------|-----------|
| `enabled` | boolean | No | Toggle bot aktif/nonaktif |
| `fallback_message` | string | No | Pesan otomatis ke member saat pertanyaan masuk. Max: 500 char |
| `suggest_contact` | boolean | No | Tampilkan info kontak di dalam bot response |
| `contact_info` | string | No | Info kontak yang ditampilkan (email, nomor WA, dll.) Max: 300 char |
| `similarity_threshold` | float (0.0–1.0) | No | Ambang batas kemiripan dengan FAQ. Lebih tinggi = lebih ketat. Default: 0.60 |

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
| Admin menjawab pertanyaan | `POST /questions/:id/answer` | Member yang bertanya | `question_answered` |
| Admin menolak pertanyaan | `PUT /questions/:id/reject` | Member yang bertanya | `question_rejected` |

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
  GET  /questions                     → antrian (default: pending, oldest first)
  GET  /questions/{id}                → detail pertanyaan
  POST /questions/{id}/answer         → jawab → (notif member)
  PUT  /questions/{id}/reject         → tolak → (notif member)

PENGATURAN
  GET  /settings/bot                  → lihat konfigurasi bot
  PUT  /settings/bot                  → update konfigurasi bot
```

---

*API Specification — QnA Backoffice. Stack: Web Admin Panel + Go Backend.*
