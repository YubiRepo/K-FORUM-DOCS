# Database Schema — Community Module

> **Stack:** Golang + PostgreSQL
> **Berdasarkan:** `COMMUNITY_RULES.md`
> **Dibuat:** 2026-06-01

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [PostgreSQL DDL Schema](#postgresql-ddl-schema)
   - [1. communities](#1-communities)
   - [2. community_members](#2-community_members)
   - [3. community_join_requests](#3-community_join_requests)
   - [4. community_posts](#4-community_posts)
   - [5. community_post_likes](#5-community_post_likes)
   - [6. community_post_comments](#6-community_post_comments)
   - [7. community_post_saves](#7-community_post_saves)
3. [Enum Values](#enum-values)
4. [Golang Structs](#golang-structs)
5. [Sample Queries](#sample-queries)
6. [Catatan Integrasi](#catatan-integrasi)

---

## Overview Relasi

```
users (auth module)
  ├── communities            (owner_id → user pembuat/leader)
  ├── community_members       (keanggotaan: user ←→ community)
  ├── community_join_requests (request join untuk private)
  ├── community_posts         (konten/feed)
  ├── community_post_likes    (like, toggle)
  ├── community_post_comments (komentar, 1 level reply)
  └── community_post_saves    (bookmark)

communities
  ├── community_members
  ├── community_join_requests
  └── community_posts
        ├── community_post_likes
        ├── community_post_comments
        └── community_post_saves

regions (region module, OPTIONAL)
  └── communities.region_id  (nullable)

-- Role & permission TIDAK di modul ini:
user_roles                    (scope_type='community', scope_id=community_id)
community_role_permissions    (permission per role per community)
community_role_permissions_template (default saat create)
```

> **Catatan:** `user_roles` dan `community_role_permissions` dimiliki oleh **Role-Permission module**. Modul Community hanya mereferensikan `community_id` ke dalamnya.

---

## PostgreSQL DDL Schema

### 1. `communities`

Entitas inti komunitas.

```sql
CREATE TABLE communities (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  name          VARCHAR(150) NOT NULL,
  slug          VARCHAR(180) NOT NULL UNIQUE,
  description   TEXT         NULL,
  avatar_url    VARCHAR(500) NULL,
  owner_id      UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  region_id     UUID         NULL REFERENCES regions(id) ON DELETE SET NULL,
  visibility    VARCHAR(20)  NOT NULL DEFAULT 'public',  -- public | private
  status        VARCHAR(20)  NOT NULL DEFAULT 'active',   -- active | suspended | archived | orphaned
  member_count  INTEGER      NOT NULL DEFAULT 0,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_communities_slug ON communities (slug);
CREATE INDEX idx_communities_owner       ON communities (owner_id);
CREATE INDEX idx_communities_region      ON communities (region_id) WHERE region_id IS NOT NULL;
CREATE INDEX idx_communities_status_vis  ON communities (status, visibility);
```

### 2. `community_members`

Status keanggotaan. **Bukan** tempat menyimpan role.

```sql
CREATE TABLE community_members (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  community_id  UUID         NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
  user_id       UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status        VARCHAR(20)  NOT NULL DEFAULT 'active',  -- active | pending | banned
  joined_at     TIMESTAMPTZ  NULL,
  banned_at     TIMESTAMPTZ  NULL,
  banned_by     UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (community_id, user_id)
);

CREATE INDEX idx_comm_members_user      ON community_members (user_id);
CREATE INDEX idx_comm_members_community ON community_members (community_id, status);
```

### 3. `community_join_requests`

Hanya untuk komunitas private.

```sql
CREATE TABLE community_join_requests (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  community_id  UUID         NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
  user_id       UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status        VARCHAR(20)  NOT NULL DEFAULT 'pending', -- pending | approved | rejected
  message       TEXT         NULL,
  reviewed_by   UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  reviewed_at   TIMESTAMPTZ  NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Hanya boleh 1 request pending aktif per user per komunitas
CREATE UNIQUE INDEX idx_comm_join_unique_pending
  ON community_join_requests (community_id, user_id)
  WHERE status = 'pending';

CREATE INDEX idx_comm_join_community ON community_join_requests (community_id, status);
```

### 4. `community_posts`

Konten/feed komunitas. Interaksi (like, comment, save) ada di tabel #5–#7.

**Format konten: plain text.** `content` disimpan apa adanya sebagai teks biasa (tanpa markdown/HTML). Client boleh meng-auto-render entity ringan saat menampilkan (URL jadi tappable, `@mention`, `#hashtag`), tetapi yang tersimpan tetap plain text. Keputusan ini menjaga konten aman, mudah dimoderasi, dan konsisten antar-client. Jika kelak butuh formatting, migrasi ke markdown bisa dilakukan dengan menambah kolom format — tidak perlu sekarang.

```sql
CREATE TABLE community_posts (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  community_id  UUID         NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
  author_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  content       TEXT         NOT NULL,                    -- PLAIN TEXT (no markdown/HTML)
  media         JSONB        NULL,                        -- array of image objects, max 10 (lihat di bawah)
  like_count    INTEGER      NOT NULL DEFAULT 0,          -- denormalized
  comment_count INTEGER      NOT NULL DEFAULT 0,          -- denormalized (top-level + replies)
  save_count    INTEGER      NOT NULL DEFAULT 0,          -- denormalized
  share_count   INTEGER      NOT NULL DEFAULT 0,          -- denormalized (share link di-generate)
  status        VARCHAR(20)  NOT NULL DEFAULT 'published', -- published | removed
  removed_by    UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  removed_at    TIMESTAMPTZ  NULL,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comm_posts_feed
  ON community_posts (community_id, created_at DESC)
  WHERE status = 'published';
CREATE INDEX idx_comm_posts_author ON community_posts (author_id);
```

**Struktur `media` (JSONB array, maksimal 10 item):**

```jsonc
[
  { "url": "https://cdn/.../1.jpg", "thumb_url": ".../1_thumb.jpg", "width": 1080, "height": 1350, "order": 0 },
  { "url": "https://cdn/.../2.jpg", "thumb_url": ".../2_thumb.jpg", "width": 1080, "height": 720,  "order": 1 }
]
```

- Maksimal **10 gambar** per post (validasi di service layer).
- Upload gambar dilakukan **terpisah** sebelum create post: client upload ke endpoint upload/CDN → dapat URL → submit post membawa array `media`. Jangan kirim base64 dalam body post.
- `thumb_url` di-generate server saat upload; feed memakai thumbnail, full image saat di-tap.
- `order` menentukan urutan tampil di grid.

> **Counter denormalized:** `like_count`, `comment_count`, `save_count`, `share_count` di-increment/decrement saat aksi terjadi, supaya feed tidak perlu `COUNT(*)` tiap load. Update dilakukan dalam transaksi yang sama dengan insert/delete aksi terkait.

### 5. `community_post_likes`

Satu user maksimal satu like per post (toggle).

```sql
CREATE TABLE community_post_likes (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id     UUID         NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
  user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (post_id, user_id)
);

CREATE INDEX idx_comm_likes_post ON community_post_likes (post_id);
CREATE INDEX idx_comm_likes_user ON community_post_likes (user_id);
```

### 6. `community_post_comments`

Komentar dengan **maksimal 1 level reply**. Komentar top-level punya `parent_comment_id = NULL`; reply mengisi `parent_comment_id`. Reply terhadap reply tidak diizinkan (divalidasi di service layer).

```sql
CREATE TABLE community_post_comments (
  id                UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id           UUID         NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
  author_id         UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  parent_comment_id UUID         NULL REFERENCES community_post_comments(id) ON DELETE CASCADE,
  content           TEXT         NOT NULL,                     -- plain text
  reply_count       INTEGER      NOT NULL DEFAULT 0,           -- denormalized (untuk top-level)
  status            VARCHAR(20)  NOT NULL DEFAULT 'published', -- published | removed
  removed_by        UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  removed_at        TIMESTAMPTZ  NULL,
  created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Feed komentar top-level per post
CREATE INDEX idx_comm_comments_post
  ON community_post_comments (post_id, created_at ASC)
  WHERE parent_comment_id IS NULL AND status = 'published';

-- Reply per komentar induk
CREATE INDEX idx_comm_comments_parent
  ON community_post_comments (parent_comment_id, created_at ASC)
  WHERE parent_comment_id IS NOT NULL AND status = 'published';
```

> **Enforce 1 level:** sebelum insert reply, service cek bahwa `parent_comment_id` menunjuk komentar yang `parent_comment_id`-nya NULL. Jika tidak, tolak (`400`).

### 7. `community_post_saves`

Bookmark/simpan post per user.

```sql
CREATE TABLE community_post_saves (
  id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  post_id     UUID         NOT NULL REFERENCES community_posts(id) ON DELETE CASCADE,
  user_id     UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  UNIQUE (post_id, user_id)
);

CREATE INDEX idx_comm_saves_user ON community_post_saves (user_id, created_at DESC);
```

> **Share:** share di Fase 1 = **share eksternal via deep link** (mis. ke WhatsApp). Tidak butuh tabel — backend cukup generate URL dan increment `community_posts.share_count`. **Repost internal** (share post ke komunitas lain) ditunda ke Fase 2.

---

## Enum Values

### `communities.visibility`
| Value | Keterangan |
|---|---|
| `public` | Siapa saja bisa auto-join |
| `private` | Join lewat approval |

### `communities.status`
| Value | Keterangan |
|---|---|
| `active` | Normal beroperasi |
| `suspended` | Dibekukan superadmin (sementara) |
| `archived` | Diarsipkan / dihapus lunak |
| `orphaned` | Tanpa leader (owner hapus akun) — perlu intervensi superadmin |

### `community_members.status`
| Value | Keterangan |
|---|---|
| `active` | Anggota aktif |
| `pending` | Menunggu approval (private) |
| `banned` | Diblokir, tidak bisa re-join |

### `community_join_requests.status`
| Value | Keterangan |
|---|---|
| `pending` | Menunggu review |
| `approved` | Disetujui → jadi member |
| `rejected` | Ditolak |

### `community_posts.status`
| Value | Keterangan |
|---|---|
| `published` | Tampil di feed |
| `removed` | Dihapus moderasi |

### `community_post_comments.status`
| Value | Keterangan |
|---|---|
| `published` | Tampil |
| `removed` | Dihapus moderasi / penulis |

---

## Golang Structs

```go
type Community struct {
    ID          string     `db:"id"          json:"id"`
    Name        string     `db:"name"        json:"name"`
    Slug        string     `db:"slug"        json:"slug"`
    Description *string    `db:"description" json:"description"`
    AvatarURL   *string    `db:"avatar_url"  json:"avatar_url"`
    OwnerID     string     `db:"owner_id"    json:"owner_id"`
    RegionID    *string    `db:"region_id"   json:"region_id"`
    Visibility  string     `db:"visibility"  json:"visibility"`
    Status      string     `db:"status"      json:"status"`
    MemberCount int        `db:"member_count" json:"member_count"`
    CreatedAt   time.Time  `db:"created_at"  json:"created_at"`
    UpdatedAt   time.Time  `db:"updated_at"  json:"updated_at"`
}

type CommunityMember struct {
    ID          string     `db:"id"           json:"id"`
    CommunityID string     `db:"community_id" json:"community_id"`
    UserID      string     `db:"user_id"      json:"user_id"`
    Status      string     `db:"status"       json:"status"`
    JoinedAt    *time.Time `db:"joined_at"    json:"joined_at"`
    BannedAt    *time.Time `db:"banned_at"    json:"banned_at"`
    BannedBy    *string    `db:"banned_by"    json:"banned_by"`
    CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
    UpdatedAt   time.Time  `db:"updated_at"   json:"updated_at"`
}

type CommunityJoinRequest struct {
    ID          string     `db:"id"           json:"id"`
    CommunityID string     `db:"community_id" json:"community_id"`
    UserID      string     `db:"user_id"      json:"user_id"`
    Status      string     `db:"status"       json:"status"`
    Message     *string    `db:"message"      json:"message"`
    ReviewedBy  *string    `db:"reviewed_by"  json:"reviewed_by"`
    ReviewedAt  *time.Time `db:"reviewed_at"  json:"reviewed_at"`
    CreatedAt   time.Time  `db:"created_at"   json:"created_at"`
}

type CommunityMedia struct {
    URL      string `json:"url"`
    ThumbURL string `json:"thumb_url"`
    Width    int    `json:"width"`
    Height   int    `json:"height"`
    Order    int    `json:"order"`
}

type CommunityPost struct {
    ID           string           `db:"id"            json:"id"`
    CommunityID  string           `db:"community_id"  json:"community_id"`
    AuthorID     string           `db:"author_id"     json:"author_id"`
    Content      string           `db:"content"       json:"content"`       // plain text
    Media        []CommunityMedia `db:"media"         json:"media"`         // JSONB, max 10
    LikeCount    int              `db:"like_count"    json:"like_count"`
    CommentCount int              `db:"comment_count" json:"comment_count"`
    SaveCount    int              `db:"save_count"    json:"save_count"`
    ShareCount   int              `db:"share_count"   json:"share_count"`
    Status       string           `db:"status"        json:"status"`
    RemovedBy    *string          `db:"removed_by"    json:"removed_by"`
    RemovedAt    *time.Time       `db:"removed_at"    json:"removed_at"`
    CreatedAt    time.Time        `db:"created_at"    json:"created_at"`
    UpdatedAt    time.Time        `db:"updated_at"    json:"updated_at"`
}

type CommunityPostLike struct {
    ID        string    `db:"id"         json:"id"`
    PostID    string    `db:"post_id"    json:"post_id"`
    UserID    string    `db:"user_id"    json:"user_id"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type CommunityPostComment struct {
    ID              string     `db:"id"                json:"id"`
    PostID          string     `db:"post_id"           json:"post_id"`
    AuthorID        string     `db:"author_id"         json:"author_id"`
    ParentCommentID *string    `db:"parent_comment_id" json:"parent_comment_id"`
    Content         string     `db:"content"           json:"content"` // plain text
    ReplyCount      int        `db:"reply_count"       json:"reply_count"`
    Status          string     `db:"status"            json:"status"`
    RemovedBy       *string    `db:"removed_by"        json:"removed_by"`
    RemovedAt       *time.Time `db:"removed_at"        json:"removed_at"`
    CreatedAt       time.Time  `db:"created_at"        json:"created_at"`
    UpdatedAt       time.Time  `db:"updated_at"        json:"updated_at"`
}

type CommunityPostSave struct {
    ID        string    `db:"id"         json:"id"`
    PostID    string    `db:"post_id"    json:"post_id"`
    UserID    string    `db:"user_id"    json:"user_id"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
}
```

---

## Sample Queries

```sql
-- Q1: Browse komunitas public + private (active), dengan filter region opsional
SELECT * FROM communities
WHERE status = 'active'
  AND ($1::uuid IS NULL OR region_id = $1)
ORDER BY member_count DESC
LIMIT $2 OFFSET $3;

-- Q2: Cek apakah user sudah anggota (dan statusnya)
SELECT status FROM community_members
WHERE community_id = $1 AND user_id = $2;

-- Q3: Daftar member aktif sebuah komunitas
SELECT cm.*, u.fullname, u.avatar
FROM community_members cm
JOIN users u ON u.id = cm.user_id
WHERE cm.community_id = $1 AND cm.status = 'active';

-- Q4: Pending join requests untuk komunitas (private)
SELECT jr.*, u.fullname
FROM community_join_requests jr
JOIN users u ON u.id = jr.user_id
WHERE jr.community_id = $1 AND jr.status = 'pending'
ORDER BY jr.created_at ASC;

-- Q5: Feed komunitas (post published, terbaru dulu)
SELECT p.*, u.fullname AS author_name, u.avatar AS author_avatar
FROM community_posts p
JOIN users u ON u.id = p.author_id
WHERE p.community_id = $1 AND p.status = 'published'
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- Q6: Komunitas yang diikuti seorang user
SELECT c.*
FROM communities c
JOIN community_members cm ON cm.community_id = c.id
WHERE cm.user_id = $1 AND cm.status = 'active' AND c.status = 'active';

-- Q7: Transfer ownership (update owner) — jalankan dalam transaksi bersama update user_roles
UPDATE communities SET owner_id = $2, updated_at = NOW() WHERE id = $1;

-- Q8: Komunitas orphaned yang perlu ditangani superadmin
SELECT * FROM communities WHERE status = 'orphaned';

-- Q9: Toggle like (insert; jika sudah ada → caller hapus). Increment counter di transaksi sama.
INSERT INTO community_post_likes (post_id, user_id) VALUES ($1, $2)
ON CONFLICT (post_id, user_id) DO NOTHING;
-- UPDATE community_posts SET like_count = like_count + 1 WHERE id = $1;

-- Q10: Unlike
DELETE FROM community_post_likes WHERE post_id = $1 AND user_id = $2;
-- UPDATE community_posts SET like_count = like_count - 1 WHERE id = $1;

-- Q11: Komentar top-level sebuah post (+ flag apakah current user sudah like post)
SELECT c.*, u.fullname AS author_name, u.avatar AS author_avatar
FROM community_post_comments c
JOIN users u ON u.id = c.author_id
WHERE c.post_id = $1 AND c.parent_comment_id IS NULL AND c.status = 'published'
ORDER BY c.created_at ASC
LIMIT $2 OFFSET $3;

-- Q12: Reply dari satu komentar induk (1 level)
SELECT c.*, u.fullname AS author_name
FROM community_post_comments c
JOIN users u ON u.id = c.author_id
WHERE c.parent_comment_id = $1 AND c.status = 'published'
ORDER BY c.created_at ASC;

-- Q13: Post yang disimpan (saved) seorang user
SELECT p.*
FROM community_post_saves s
JOIN community_posts p ON p.id = s.post_id
WHERE s.user_id = $1 AND p.status = 'published'
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- Q14: Cek interaksi current user terhadap satu post (liked? saved?)
SELECT
  EXISTS(SELECT 1 FROM community_post_likes WHERE post_id = $1 AND user_id = $2) AS is_liked,
  EXISTS(SELECT 1 FROM community_post_saves WHERE post_id = $1 AND user_id = $2) AS is_saved;
```

---

## Catatan Integrasi

1. **Create community** = transaksi gabungan: insert `communities`, insert `community_members` (owner, active), lalu Role-Permission module assign `user_roles` (leader) + copy `community_role_permissions_template`.
2. **Permission check** untuk post/moderate/delete dilakukan via Role-Permission module — modul Community hanya memanggilnya, tidak menyimpan logika permission.
3. **Notification**: emit event `member_joined`, `member_left`, dan `new_posts` ke modul Notification (sesuai `notification-preferences-technical.md`).
4. **Region**: `region_id` nullable; jika diisi, hanya sebagai label & filter (pola modul Event), bukan pembatas akses.
5. **Cascade delete**: menghapus `communities` akan cascade ke members, join_requests, posts. Tetapi `user_roles` & `community_role_permissions` ada di modul lain — pastikan cleanup dipanggil eksplisit dalam transaksi.

---

*Dokumen ini adalah skema awal modul Community hasil breakdown. Sinkron dengan `COMMUNITY_RULES.md`.*
