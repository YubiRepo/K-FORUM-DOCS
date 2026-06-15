# Schedule Module — Database Schema (v1.1)

Database schema untuk Schedule module KAI App. Kalender agenda backoffice dengan sharing tiga mode (private/all_admins/specific) dan field hook untuk future integrasi (Event, Announcement, recurrence, region).

---

## Overview Relasi

```
users (dari modul auth)
  ├── schedule_entries (created_by, assigned_to)
  ├── schedule_entry_shares (user_id) ← siapa yang di-invite
  └── schedule_entry_types (created_by)

schedule_entry_types
  └── schedule_entries (1:N, via entry_type_id)

schedule_entries
  └── schedule_entry_shares (1:N, dipakai saat visibility = 'specific')

-- Future link (tidak ada FK keras, menunjuk by id):
schedule_entries.source_ref → events.id / announcements.id (saat source=linked)
```

---

## 1. `schedule_entry_types`

Master tipe agenda. Dikelola Superadmin, extensible.

```sql
CREATE TABLE schedule_entry_types (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(30) NOT NULL UNIQUE,   -- 'agenda' | 'reminder' | 'milestone' | ...
    name        VARCHAR(50) NOT NULL,
    color       VARCHAR(7) NULL,               -- hex untuk render kalender, mis. '#3B82F6'
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_by  UUID NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO schedule_entry_types (code, name, color) VALUES
  ('agenda',    'Agenda',    '#3B82F6'),
  ('reminder',  'Reminder',  '#F59E0B'),
  ('milestone', 'Milestone', '#10B981');
```

---

## 2. `schedule_entries`

Tabel utama — satu row per agenda.

```sql
CREATE TABLE schedule_entries (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title         VARCHAR(200) NOT NULL,
    description   TEXT NULL,
    entry_type_id UUID NOT NULL REFERENCES schedule_entry_types(id),

    -- Waktu
    start_at      TIMESTAMPTZ NOT NULL,
    end_at        TIMESTAMPTZ NULL,            -- opsional; jika ada harus >= start_at
    all_day       BOOLEAN NOT NULL DEFAULT FALSE,

    location      TEXT NULL,

    -- Visibility & status
    visibility    VARCHAR(12) NOT NULL DEFAULT 'private', -- 'private' | 'all_admins' | 'specific'
    status        VARCHAR(10) NOT NULL DEFAULT 'active',  -- 'active' | 'done' | 'cancelled'

    -- Hook integrasi (future-proof)
    source        VARCHAR(10) NOT NULL DEFAULT 'manual',  -- 'manual' | 'linked'
    source_module VARCHAR(30) NULL,            -- 'event' | 'announcement' | ... (saat linked)
    source_ref    VARCHAR(100) NULL,           -- id entitas asal (event_id, dll)
    recurrence    TEXT NULL,                   -- RRULE-style string (future)
    assigned_to   UUID NULL REFERENCES users(id), -- future

    created_by    UUID NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_schedule_visibility CHECK (visibility IN ('private', 'all_admins', 'specific')),
    CONSTRAINT chk_schedule_status     CHECK (status IN ('active', 'done', 'cancelled')),
    CONSTRAINT chk_schedule_source     CHECK (source IN ('manual', 'linked')),
    CONSTRAINT chk_schedule_time       CHECK (end_at IS NULL OR end_at >= start_at)
);

CREATE INDEX idx_schedule_created_by ON schedule_entries (created_by);
CREATE INDEX idx_schedule_start      ON schedule_entries (start_at);
CREATE INDEX idx_schedule_visibility ON schedule_entries (visibility);
CREATE INDEX idx_schedule_status     ON schedule_entries (status);
CREATE INDEX idx_schedule_type       ON schedule_entries (entry_type_id);
CREATE INDEX idx_schedule_source     ON schedule_entries (source, source_module, source_ref);
```

---

## 3. `schedule_entry_shares`

Daftar user yang di-invite untuk melihat agenda. **Hanya relevan saat `visibility = 'specific'`.** View-only — tidak memberi hak edit.

```sql
CREATE TABLE schedule_entry_shares (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_entry_id UUID NOT NULL REFERENCES schedule_entries(id) ON DELETE CASCADE,
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_schedule_share UNIQUE (schedule_entry_id, user_id)
);

CREATE INDEX idx_share_entry ON schedule_entry_shares (schedule_entry_id);
CREATE INDEX idx_share_user  ON schedule_entry_shares (user_id);
```

### Catatan
- `ON DELETE CASCADE` → bila agenda dihapus atau user dihapus, baris share ikut bersih.
- Daftar share **tetap tersimpan** walau `visibility` diganti ke selain `specific` (tidak dipakai sampai mode kembali `specific`).
- **Future region:** mode `region` cukup mengisi tabel ini otomatis dari anggota region — struktur tak berubah.

---

## 4. Query Patterns Umum

### Kalender untuk satu admin (semua agenda yang boleh ia lihat) dalam rentang
```sql
SELECT e.*, t.code AS type_code, t.name AS type_name, t.color
FROM schedule_entries e
JOIN schedule_entry_types t ON t.id = e.entry_type_id
WHERE e.status != 'cancelled'
  AND e.start_at < $range_end
  AND COALESCE(e.end_at, e.start_at) >= $range_start
  AND (
        e.created_by = $current_user                       -- miliknya (mode apa pun)
        OR e.visibility = 'all_admins'                     -- agenda tim
        OR (e.visibility = 'specific' AND EXISTS (          -- di-invite
              SELECT 1 FROM schedule_entry_shares s
              WHERE s.schedule_entry_id = e.id AND s.user_id = $current_user
            ))
      )
ORDER BY e.start_at;
```

### Superadmin: lihat semua (override, termasuk private/specific orang lain)
```sql
SELECT e.*, t.code AS type_code, t.color
FROM schedule_entries e
JOIN schedule_entry_types t ON t.id = e.entry_type_id
WHERE e.start_at < $range_end
  AND COALESCE(e.end_at, e.start_at) >= $range_start
ORDER BY e.start_at;
```

### Ambil daftar user yang di-invite pada satu agenda
```sql
SELECT u.id, u.name
FROM schedule_entry_shares s
JOIN users u ON u.id = s.user_id
WHERE s.schedule_entry_id = $1;
```

### Filter by type & status
```sql
SELECT * FROM schedule_entries
WHERE entry_type_id = $1 AND status = $2
ORDER BY start_at;
```

### Entri linked dari modul tertentu (future)
```sql
SELECT * FROM schedule_entries
WHERE source = 'linked' AND source_module = $1 AND source_ref = $2;
```

---

## 5. Catatan Implementasi

- **Scope visibility di-enforce di backend** — jangan andalkan frontend untuk menyaring agenda yang tidak boleh dilihat.
- **Sinkron share saat update** — saat pembuat mengubah daftar invite, hitung selisih (tambah/hapus baris) di `schedule_entry_shares`.
- **Cancelled, bukan delete** untuk agenda yang dibagikan; delete permanen hanya oleh pembuat/Superadmin.
- **Trigger `updated_at`** disarankan pada `schedule_entries` dan `schedule_entry_types`.
- **Recurrence** di-expand di layer aplikasi saat render, bukan disimpan sebagai banyak row.
- **Linked entries (Phase 2)** dibuat via job sinkronisasi dari modul sumber, bukan input manual.
