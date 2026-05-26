# Migration Explanation: Community Role Permissions Template

> **File**: `migrations/202605251600_create_community_role_permissions_template.sql`
> **Purpose**: Menjelaskan migration untuk template default permissions komunitas baru
> **Audience**: Backend team, database engineers

---

## Table of Contents

1. [Apa Itu Migration Ini?](#apa-itu-migration-ini)
2. [Mengapa Diperlukan?](#mengapa-diperlukan)
3. [Table Structure](#table-structure)
4. [SQL Breakdown](#sql-breakdown)
5. [Seed Data Explanation](#seed-data-explanation)
6. [Rollback Process](#rollback-process)
7. [Usage Flow](#usage-flow)
8. [Important Notes](#important-notes)

---

## Apa Itu Migration Ini?

### Tujuan

Migration ini membuat **template database table** yang menyimpan default permissions untuk setiap community role. Template ini digunakan saat member pro membuat komunitas baru.

### Kapan Dijalankan

- Dijalankan **SEKALI** saat database initialization
- Atau saat deployment awal (sebelum komunitas pertama dibuat)
- **JANGAN** di-run ulang (migration idempotent — aman jika run 2x)

### Apa Hasilnya

Setelah migration:
- ✅ Table `community_role_permissions_template` dibuat
- ✅ Default permissions di-seed (8 rows: 4 leader, 3 moderator, 1 member)
- ✅ Siap digunakan oleh backend saat create community

---

## Mengapa Diperlukan?

### Problem

Sebelum migration ini, backend **hardcode** default permissions saat create community:

```go
// BAD: Hardcoded
defaultPermissions := []string{
  "post_content",
  "moderate_posts",
  "manage_members",
  "delete_content",
}
// Harus re-deploy backend untuk ubah default!
```

### Solution

Dengan migration ini, default permissions **disimpan di database**:

```sql
-- GOOD: Di-database, bisa diubah tanpa re-deploy
SELECT * FROM community_role_permissions_template
WHERE role_id = 'leader'
-- Result: [post_content, moderate_posts, manage_members, delete_content]
```

### Benefit

✅ **Flexible** — Superadmin bisa ubah default via backoffice
✅ **No redeploy** — Ubah template tanpa touch code
✅ **Versioning** — Bisa track changes di database
✅ **Different defaults** — Bisa punya multiple template versions di future

---

## Table Structure

### Tabel: `community_role_permissions_template`

```sql
CREATE TABLE community_role_permissions_template (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(role_id, permission_id)
);
```

### Kolom Breakdown

| Kolom | Type | Meaning | Contoh |
|-------|------|---------|--------|
| `id` | UUID | Unique identifier untuk setiap template entry | `550e8400-e29b-41d4-a716-446655440000` |
| `role_id` | UUID | Reference ke role (leader, moderator, member) | `550e8400-e29b-41d4-a716-446655440001` |
| `permission_id` | UUID | Reference ke permission (post_content, etc) | `550e8400-e29b-41d4-a716-446655440002` |
| `created_at` | TIMESTAMPTZ | Saat entry dibuat | `2026-05-25 10:00:00 UTC` |
| `updated_at` | TIMESTAMPTZ | Saat entry terakhir diupdate | `2026-05-25 10:00:00 UTC` |

### Constraints

**PRIMARY KEY: `id`**
- Setiap entry punya unique ID
- Auto-generated sebagai UUID

**FOREIGN KEY: `role_id`**
- Harus reference role yang exist di tabel `roles`
- `ON DELETE CASCADE` — jika role dihapus, template entry juga terhapus

**FOREIGN KEY: `permission_id`**
- Harus reference permission yang exist di tabel `permissions`
- `ON DELETE CASCADE` — jika permission dihapus, template entry juga terhapus

**UNIQUE: `(role_id, permission_id)`**
- Kombinasi role_id + permission_id harus unik
- Tidak boleh ada 2 entry dengan role_id=A dan permission_id=B yang sama
- **Tujuan**: Mencegah duplicate permissions di template

### Indexes

```sql
CREATE INDEX idx_template_role_id 
  ON community_role_permissions_template(role_id);

CREATE INDEX idx_template_permission_id 
  ON community_role_permissions_template(permission_id);
```

**Mengapa perlu index?**

Saat backend query template:
```sql
SELECT * FROM community_role_permissions_template
WHERE role_id = 'role_leader'
```

Index pada `role_id` membuat query lebih cepat (DB bisa jump langsung ke rows dengan role_id=leader, tidak perlu scan seluruh table).

**Impact**: Dari O(n) menjadi O(log n)

---

## SQL Breakdown

### Bagian 1: Create Table

```sql
CREATE TABLE IF NOT EXISTS community_role_permissions_template (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ...
);
```

**`IF NOT EXISTS`**: Jika table sudah ada, skip (idempotent — aman run 2x)

**`UUID PRIMARY KEY`**: Setiap row punya unique ID

**`DEFAULT gen_random_uuid()`**: Jika insert tanpa specify id, auto-generate

---

### Bagian 2: Indexes

```sql
CREATE INDEX idx_template_role_id 
  ON community_role_permissions_template(role_id);
```

**Untuk speed up queries**:
```sql
-- Query ini akan cepat karena ada index
SELECT * FROM community_role_permissions_template
WHERE role_id = ?
```

---

### Bagian 3: Auto-update Trigger

```sql
CREATE TRIGGER update_community_role_permissions_template_timestamp
  BEFORE UPDATE ON community_role_permissions_template
  FOR EACH ROW
  EXECUTE FUNCTION update_timestamp();
```

**Apa ini?**

Setiap kali ada UPDATE ke table, `updated_at` field otomatis update ke `NOW()`.

**Contoh**:
```sql
-- Admin update: remove 'delete_content' dari moderator
UPDATE community_role_permissions_template
SET deleted_at = '2026-05-25'
WHERE role_id = 'moderator' AND permission_id = 'delete_content'

-- RESULT: updated_at automatically set to current timestamp
```

**Tujuan**: Audit trail — bisa tau kapan terakhir template diubah

---

## Seed Data Explanation

### Apa Itu Seed?

**Seed** = Data awal yang di-insert saat migration jalan.

Setelah migration selesai, table sudah berisi default permissions tanpa perlu manual insert.

### Seed Script

```sql
DO $$ 
DECLARE
  v_role_leader UUID;
  v_role_moderator UUID;
  v_role_member UUID;
  v_perm_post UUID;
  v_perm_moderate UUID;
  v_perm_manage UUID;
  v_perm_delete UUID;
BEGIN
  -- Get role IDs dari tabel roles
  SELECT id INTO v_role_leader 
    FROM roles WHERE name = 'leader' AND role_type = 'community';
  
  SELECT id INTO v_role_moderator 
    FROM roles WHERE name = 'moderator' AND role_type = 'community';
  
  SELECT id INTO v_role_member 
    FROM roles WHERE name = 'community_member' AND role_type = 'community';
  
  -- Get permission IDs dari tabel permissions
  SELECT id INTO v_perm_post FROM permissions WHERE key = 'post_content';
  SELECT id INTO v_perm_moderate FROM permissions WHERE key = 'moderate_posts';
  SELECT id INTO v_perm_manage FROM permissions WHERE key = 'manage_members';
  SELECT id INTO v_perm_delete FROM permissions WHERE key = 'delete_content';
  
  -- Insert template permissions
  INSERT INTO community_role_permissions_template (role_id, permission_id)
  VALUES 
    (v_role_leader, v_perm_post),
    (v_role_leader, v_perm_moderate),
    (v_role_leader, v_perm_manage),
    (v_role_leader, v_perm_delete);
  
  -- ... insert untuk moderator dan member ...
END $$;
```

### Step by Step

**Step 1: Query Role IDs**

```sql
SELECT id INTO v_role_leader 
  FROM roles WHERE name = 'leader' AND role_type = 'community';
```

Query tabel `roles` untuk cari role dengan name='leader'. Simpan ID-nya ke variable `v_role_leader`.

**Contoh:**
```
Name: leader
Role Type: community
ID: 550e8400-e29b-41d4-a716-446655440001 ← diambil ini
```

**Step 2: Query Permission IDs**

```sql
SELECT id INTO v_perm_post FROM permissions WHERE key = 'post_content';
```

Query tabel `permissions` untuk cari permission dengan key='post_content'. Simpan ID-nya.

**Contoh:**
```
Key: post_content
Display Name: Post Content
ID: 550e8400-e29b-41d4-a716-446655440002 ← diambil ini
```

**Step 3: Insert Mappings**

```sql
INSERT INTO community_role_permissions_template (role_id, permission_id)
VALUES 
  (v_role_leader, v_perm_post),
  (v_role_leader, v_perm_moderate),
  (v_role_leader, v_perm_manage),
  (v_role_leader, v_perm_delete);
```

Insert 4 rows untuk leader role:
- leader → post_content
- leader → moderate_posts
- leader → manage_members
- leader → delete_content

**Tabel setelah insert:**

```
┌─────────────────────────────────────────────────────┐
│ community_role_permissions_template                 │
├──────┬─────────────────┬──────────────────────┐
│ id   │ role_id         │ permission_id        │
├──────┼─────────────────┼──────────────────────┤
│ uuid1│ role_leader     │ perm_post            │
│ uuid2│ role_leader     │ perm_moderate        │
│ uuid3│ role_leader     │ perm_manage          │
│ uuid4│ role_leader     │ perm_delete          │
│ uuid5│ role_moderator  │ perm_post            │
│ uuid6│ role_moderator  │ perm_moderate        │
│ uuid7│ role_moderator  │ perm_delete          │
│ uuid8│ role_member     │ perm_post            │
└──────┴─────────────────┴──────────────────────┘
```

### Kondisi Prerequisites

**Script ini HANYA jalan jika:**

1. ✅ Table `roles` sudah exist
2. ✅ Role 'leader', 'moderator', 'community_member' sudah di-seed di `roles` table
3. ✅ Table `permissions` sudah exist
4. ✅ Permissions 'post_content', 'moderate_posts', 'manage_members', 'delete_content' sudah di-seed

**Jika tidak, script error:**
```
ERROR: relation "roles" does not exist
```

**Solusi**: Jalankan migration untuk `roles` dan `permissions` terlebih dahulu!

---

## Rollback Process

### Migration Down

```sql
DROP TRIGGER IF EXISTS update_community_role_permissions_template_timestamp 
  ON community_role_permissions_template;

DROP INDEX IF EXISTS idx_template_permission_id;
DROP INDEX IF EXISTS idx_template_role_id;

DROP TABLE IF EXISTS community_role_permissions_template;
```

### Step by Step

**Step 1: Drop Trigger**

```sql
DROP TRIGGER IF EXISTS update_community_role_permissions_template_timestamp 
  ON community_role_permissions_template;
```

Hapus trigger yang auto-update `updated_at`.

**`IF EXISTS`**: Jika trigger tidak ada, skip (aman, tidak error)

**Step 2: Drop Indexes**

```sql
DROP INDEX IF EXISTS idx_template_role_id;
DROP INDEX IF EXISTS idx_template_permission_id;
```

Hapus indexes yang speedup queries.

**`IF EXISTS`**: Jika index tidak ada, skip

**Step 3: Drop Table**

```sql
DROP TABLE IF EXISTS community_role_permissions_template;
```

Hapus table dan semua data di dalamnya.

**`IF EXISTS`**: Jika table tidak ada, skip

### Konsekuensi Rollback

❌ **Semua template data hilang**
- Jika ada komunitas yang dibuat setelah migration, permission-nya tidak berubah
- Template yang disimpan hilang

✅ **Aman untuk development/testing**
- Bisa rollback tanpa khawatir corrupted data
- Bisa run migration lagi (idempotent)

---

## Usage Flow

### Flow 1: Migration Run

```
Developer:
$ npm run migrate:up

Database:
1. Create table community_role_permissions_template
2. Create indexes
3. Create trigger
4. Seed initial data (8 rows)

Result:
✅ Template ready to use
✅ Backend dapat query dari template
```

### Flow 2: Backend Query Template (Saat Create Community)

```go
// Backend create community

// Query: Get template permissions untuk semua role
template := db.Query(`
  SELECT role_id, ARRAY_AGG(permission_id) as permission_ids
  FROM community_role_permissions_template
  GROUP BY role_id
`)

// Result:
// {
//   role_id: 'role_leader',
//   permission_ids: [perm_post, perm_moderate, perm_manage, perm_delete]
// },
// {
//   role_id: 'role_moderator',
//   permission_ids: [perm_post, perm_moderate, perm_delete]
// },
// ...

// For each role in template:
for role in template {
  // Insert permissions ke community_role_permissions
  for perm in role.permission_ids {
    db.Exec(`
      INSERT INTO community_role_permissions (community_id, role_id, permission_id)
      VALUES (?, ?, ?)
    `, communityID, role.role_id, perm)
  }
}
```

### Flow 3: Superadmin Update Template (Via Backoffice)

```
Superadmin:
1. Access: Settings → Community Role Permissions Template
2. Edit: Moderator role
3. Remove: 'delete_content' permission
4. Save

Backoffice API:
DELETE FROM community_role_permissions_template
WHERE role_id = 'role_moderator'
  AND permission_id = 'perm_delete'

Result:
✅ Template updated
✅ New communities: moderator tidak bisa delete
✅ Existing communities: tidak affected
```

---

## Important Notes

### ✅ DO:

- ✅ Run migration sebelum backend run
- ✅ Seed hanya default permissions (yang essential)
- ✅ Gunakan indexes untuk optimize queries
- ✅ Track updated_at untuk audit

### ❌ DON'T:

- ❌ Jangan hardcode permissions di backend
- ❌ Jangan run migration multiple times (idempotent — safe, tapi unnecessary)
- ❌ Jangan hapus template row manual (gunakan API)
- ❌ Jangan change default template untuk existing communities (only affects new)

### Safety Considerations:

**Referential Integrity**:
```
Jika role_id di-delete dari tabel roles:
→ Template entry otomatis delete (CASCADE)
→ Table tetap consistent

Jika permission_id di-delete:
→ Template entry otomatis delete (CASCADE)
→ Table tetap consistent
```

**Uniqueness**:
```
Tidak boleh insert:
(role_leader, perm_post) 2x

Akan error:
UNIQUE constraint violation
```

---

## Summary

### Apa Migration Ini Lakukan?

| Aspek | Detail |
|-------|--------|
| **Membuat** | Table `community_role_permissions_template` |
| **Seed** | 8 default permissions (4 leader, 3 moderator, 1 member) |
| **Indexes** | Untuk optimize queries by role_id & permission_id |
| **Trigger** | Auto-update `updated_at` saat ada perubahan |
| **Purpose** | Store default permissions untuk komunitas baru |

### Kapan Digunakan?

| Waktu | Yang Terjadi |
|-------|-------------|
| **Migration Run** | Template di-seed ke database |
| **Create Community** | Backend query template, copy ke komunitas baru |
| **Edit Template** | Superadmin via backoffice, affect komunitas baru saja |

### Benefit?

| Benefit | Penjelasan |
|---------|-----------|
| **Flexible** | Bisa ubah default tanpa re-deploy |
| **Maintainable** | Template di-database, easy to track & audit |
| **Scalable** | Bisa support multiple template versions |
| **Safe** | Existing communities tidak affected |

---

*Penjelasan lengkap migration untuk community role permissions template. Gunakan sebagai reference untuk mengerti flow dan implementation.*
