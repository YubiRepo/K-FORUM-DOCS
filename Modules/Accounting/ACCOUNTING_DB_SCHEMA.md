# Accounting Module — Database Schema (v1.0)

Database schema untuk Accounting module KAI App. Ledger pencatatan in/out dengan scope region, multi-currency, dan field hook untuk future integrasi.

---

## Overview Relasi

```
regions (dari modul region)
  └── accounting_entries (1:N, via region_id; null = Pusat)

accounting_categories
  └── accounting_entries (1:N, via category_id)

users (dari modul auth)
  ├── accounting_entries (created_by, verified_by)
  └── accounting_categories (created_by)

-- Future (Phase 3):
accounting_budgets
accounting_recurring_templates
```

---

## 1. `accounting_categories`

Master kategori pemasukan/pengeluaran, **hierarkis (parent → child, maks 2 level)**. Dikelola Superadmin.

```sql
CREATE TABLE accounting_categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        VARCHAR(50) NOT NULL UNIQUE,        -- mis. 'REV_SUBSCRIPTION', 'EXP_OP_INTERNET'
    name        VARCHAR(100) NOT NULL,
    direction   VARCHAR(3) NOT NULL,                -- 'IN' | 'OUT'
    parent_id   UUID NULL REFERENCES accounting_categories(id), -- null = parent/top-level
    description TEXT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_by  UUID NULL REFERENCES users(id),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_category_direction CHECK (direction IN ('IN', 'OUT')),
    CONSTRAINT chk_category_not_self  CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE INDEX idx_category_direction ON accounting_categories (direction);
CREATE INDEX idx_category_active    ON accounting_categories (is_active);
CREATE INDEX idx_category_parent    ON accounting_categories (parent_id);
```

### Aturan hierarki (di-enforce di aplikasi)
- Maksimum 2 level: parent yang punya `parent_id` sendiri tidak boleh dijadikan parent lagi (no grandchild).
- `direction` child harus sama dengan parent-nya.
- Parent tidak bisa di-`is_active = false` jika masih punya child aktif.
- `parent_id` & `direction` tidak bisa diubah jika kategori sudah dipakai transaksi.

### Seed default (parent dulu, lalu child)
```sql
-- Parent / top-level
INSERT INTO accounting_categories (code, name, direction) VALUES
  ('REV_SUBSCRIPTION', 'Subscription Revenue', 'IN'),
  ('REV_ADS',          'Ads Revenue',          'IN'),
  ('REV_EVENT',        'Event Revenue',        'IN'),
  ('REV_DONATION',     'Donation',             'IN'),
  ('REV_OTHER',        'Other Income',         'IN'),
  ('EXP_SALARY',       'Gaji & Honor',         'OUT'),
  ('EXP_RENT',         'Sewa',                 'OUT'),
  ('EXP_EVENT',        'Biaya Event',          'OUT'),
  ('EXP_OPERATIONAL',  'Operasional',          'OUT'),
  ('EXP_MARKETING',    'Marketing',            'OUT'),
  ('EXP_OTHER',        'Other Expense',        'OUT');

-- Child (contoh) — parent_id diambil dari code parent
INSERT INTO accounting_categories (code, name, direction, parent_id)
SELECT 'REV_EVENT_TICKET',  'Tiket',  'IN',  id FROM accounting_categories WHERE code = 'REV_EVENT';
INSERT INTO accounting_categories (code, name, direction, parent_id)
SELECT 'REV_EVENT_SPONSOR', 'Sponsor','IN',  id FROM accounting_categories WHERE code = 'REV_EVENT';
INSERT INTO accounting_categories (code, name, direction, parent_id)
SELECT 'EXP_OP_ELECTRICITY','Listrik','OUT', id FROM accounting_categories WHERE code = 'EXP_OPERATIONAL';
INSERT INTO accounting_categories (code, name, direction, parent_id)
SELECT 'EXP_OP_INTERNET',   'Internet','OUT',id FROM accounting_categories WHERE code = 'EXP_OPERATIONAL';
INSERT INTO accounting_categories (code, name, direction, parent_id)
SELECT 'EXP_OP_SUPPLIES',   'ATK & Perlengkapan','OUT', id FROM accounting_categories WHERE code = 'EXP_OPERATIONAL';
```

---

## 2. `accounting_entries`

Tabel utama — satu row per transaksi (in/out).

```sql
CREATE TABLE accounting_entries (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    direction        VARCHAR(3) NOT NULL,            -- 'IN' | 'OUT'
    category_id      UUID NOT NULL REFERENCES accounting_categories(id),
    region_id        VARCHAR(50) NULL,               -- null = Pusat; atau region tertentu

    -- Nominal & currency
    amount           NUMERIC(18,2) NOT NULL,         -- nominal dalam currency asli
    currency         VARCHAR(3) NOT NULL DEFAULT 'IDR',
    exchange_rate    NUMERIC(18,6) NOT NULL DEFAULT 1, -- kurs ke IDR
    amount_base      NUMERIC(18,2) NOT NULL,         -- amount × exchange_rate (IDR), untuk agregasi

    -- Deskripsi
    description      TEXT NULL,
    transaction_date DATE NOT NULL,                  -- kapan transaksi terjadi
    attachment_url   TEXT NULL,                      -- bukti: nota/kwitansi/transfer

    -- Sumber & hook integrasi (future-proof)
    source           VARCHAR(10) NOT NULL DEFAULT 'manual', -- 'manual' | 'system'
    source_ref       VARCHAR(100) NULL,              -- link balik: subscription_request_id, ad_id, dll
    external_txn_id  VARCHAR(100) NULL,              -- id transaksi payment gateway (future)
    payment_provider VARCHAR(20) NULL,               -- 'manual' | 'stripe' | 'midtrans' (future)

    -- Status & audit
    status           VARCHAR(10) NOT NULL DEFAULT 'recorded', -- 'recorded' | 'verified' | 'void'
    void_reason      TEXT NULL,
    verified_by      UUID NULL REFERENCES users(id),
    verified_at      TIMESTAMPTZ NULL,

    -- Reconciliation (future)
    reconciled       BOOLEAN NOT NULL DEFAULT FALSE,
    reconciled_at    TIMESTAMPTZ NULL,

    created_by       UUID NOT NULL REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_entry_region FOREIGN KEY (region_id) REFERENCES regions(id),
    CONSTRAINT chk_entry_direction CHECK (direction IN ('IN', 'OUT')),
    CONSTRAINT chk_entry_source    CHECK (source IN ('manual', 'system')),
    CONSTRAINT chk_entry_status    CHECK (status IN ('recorded', 'verified', 'void')),
    CONSTRAINT chk_entry_amount    CHECK (amount > 0)
);

CREATE INDEX idx_entry_region     ON accounting_entries (region_id);
CREATE INDEX idx_entry_category   ON accounting_entries (category_id);
CREATE INDEX idx_entry_direction  ON accounting_entries (direction);
CREATE INDEX idx_entry_status     ON accounting_entries (status);
CREATE INDEX idx_entry_txn_date   ON accounting_entries (transaction_date);
CREATE INDEX idx_entry_source     ON accounting_entries (source, source_ref);
CREATE INDEX idx_entry_created_by ON accounting_entries (created_by);
```

### Catatan field
- **`amount_base`** dihitung saat input (`amount × exchange_rate`). Semua laporan agregat pakai ini agar konsisten lintas currency.
- **`source` / `source_ref`** memisahkan entri manual dari entri auto (future). Phase 1 selalu `manual`.
- **`external_txn_id` / `payment_provider`** disiapkan untuk integrasi gateway, kosong di Phase 1.
- **`transaction_date`** dipakai untuk laporan keuangan (bukan `created_at`).
- **`region_id` null** = transaksi tingkat Pusat (bukan region tertentu).

---

## 3. `accounting_settings`

Konfigurasi global accounting. Satu row, dikelola Superadmin.

```sql
CREATE TABLE accounting_settings (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    verification_required    BOOLEAN NOT NULL DEFAULT FALSE,
    default_currency         VARCHAR(3) NOT NULL DEFAULT 'IDR',
    allow_region_admin_input BOOLEAN NOT NULL DEFAULT TRUE,
    require_attachment_for_out BOOLEAN NOT NULL DEFAULT FALSE,
    fiscal_year_start_month  INTEGER NOT NULL DEFAULT 1,
    updated_by               UUID NULL REFERENCES users(id),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_fiscal_month CHECK (fiscal_year_start_month BETWEEN 1 AND 12)
);

INSERT INTO accounting_settings (verification_required, default_currency)
VALUES (FALSE, 'IDR');
```

---

## 4. Tabel Future (Phase 3) — disiapkan, belum dipakai

### `accounting_budgets`
```sql
CREATE TABLE accounting_budgets (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id     VARCHAR(50) NULL REFERENCES regions(id), -- null = Pusat
    category_id   UUID NULL REFERENCES accounting_categories(id),
    period_year   INTEGER NOT NULL,
    period_month  INTEGER NULL,                  -- null = budget tahunan
    planned_amount NUMERIC(18,2) NOT NULL,
    currency      VARCHAR(3) NOT NULL DEFAULT 'IDR',
    created_by    UUID NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### `accounting_recurring_templates`
```sql
CREATE TABLE accounting_recurring_templates (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    direction     VARCHAR(3) NOT NULL,
    category_id   UUID NOT NULL REFERENCES accounting_categories(id),
    region_id     VARCHAR(50) NULL REFERENCES regions(id),
    amount        NUMERIC(18,2) NOT NULL,
    currency      VARCHAR(3) NOT NULL DEFAULT 'IDR',
    description   TEXT NULL,
    frequency     VARCHAR(20) NOT NULL,          -- 'monthly' | 'weekly' | 'yearly'
    next_run_date DATE NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_by    UUID NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_recurring_direction CHECK (direction IN ('IN','OUT')),
    CONSTRAINT chk_recurring_frequency CHECK (frequency IN ('monthly','weekly','yearly'))
);
```

---

## 5. Query Patterns Umum

### Saldo region untuk periode tertentu
```sql
SELECT
  COALESCE(SUM(amount_base) FILTER (WHERE direction = 'IN'), 0)  AS total_in,
  COALESCE(SUM(amount_base) FILTER (WHERE direction = 'OUT'), 0) AS total_out,
  COALESCE(SUM(amount_base) FILTER (WHERE direction = 'IN'), 0)
    - COALESCE(SUM(amount_base) FILTER (WHERE direction = 'OUT'), 0) AS balance
FROM accounting_entries
WHERE status != 'void'
  AND region_id = $1
  AND transaction_date BETWEEN $2 AND $3;
```

### Breakdown per kategori
```sql
SELECT c.code, c.name, c.direction,
       SUM(e.amount_base) AS total
FROM accounting_entries e
JOIN accounting_categories c ON c.id = e.category_id
WHERE e.status != 'void'
  AND e.transaction_date BETWEEN $1 AND $2
GROUP BY c.code, c.name, c.direction
ORDER BY c.direction, total DESC;
```

### Breakdown roll-up ke parent (gabungkan child ke induknya)
```sql
-- COALESCE(parent_id, id): entri di child dijumlahkan ke parent;
-- entri yang di-assign langsung ke parent tetap masuk ke parent.
SELECT p.code AS parent_code, p.name AS parent_name, p.direction,
       SUM(e.amount_base) AS total
FROM accounting_entries e
JOIN accounting_categories c ON c.id = e.category_id
JOIN accounting_categories p ON p.id = COALESCE(c.parent_id, c.id)
WHERE e.status != 'void'
  AND e.transaction_date BETWEEN $1 AND $2
GROUP BY p.code, p.name, p.direction
ORDER BY p.direction, total DESC;
```

### Cashflow per region (laporan global Superadmin)
```sql
SELECT region_id,
       SUM(amount_base) FILTER (WHERE direction = 'IN')  AS total_in,
       SUM(amount_base) FILTER (WHERE direction = 'OUT') AS total_out
FROM accounting_entries
WHERE status != 'void'
  AND transaction_date BETWEEN $1 AND $2
GROUP BY region_id;
```

### Entri dari sumber sistem (future auto-record audit)
```sql
SELECT * FROM accounting_entries
WHERE source = 'system' AND source_ref = $1; -- mis. subscription_request_id
```

---

## 6. Catatan Implementasi

- **`amount_base` dihitung di aplikasi** saat create/update, bukan di DB, agar kurs yang dipakai tercatat eksplisit di `exchange_rate`.
- **Void, bukan delete** — entri tidak pernah dihapus fisik; set `status = 'void'` + `void_reason`.
- **Trigger `updated_at`** disarankan pada `accounting_entries` dan `accounting_categories` (pola sama dengan `ad_settings`).
- **Scope enforcement di query layer** — Admin Region selalu di-filter `region_id = <region-nya>` di backend, jangan andalkan frontend.
