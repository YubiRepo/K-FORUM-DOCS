# ID Card Module — Database Schema (v2.0)

Database schema untuk ID Card module KAI App. Versi ini **decoupled dari Subscription**: kartu adalah identitas stabil, plan/benefit di-resolve live. Field `plan` dan `expiry_date` dihapus dari v1.0.

---

## Overview Relasi

```
users (dari modul auth)
  └── membership_id_cards (1:1)
       ├── orders (fulfillment, 1:0..1)
       └── card_scan_events (1:N)  ← event stream, fondasi future-proof

regions (dari modul region)
  └── membership_id_cards (1:N, via region_id)
```

> Catatan: TIDAK ada relasi langsung ke tabel `plans`/`subscriptions`. Plan di-resolve live saat verifikasi melalui modul Subscription.

---

## 1. `membership_id_cards`

Identitas keanggotaan. Satu user = satu kartu.

```sql
CREATE TABLE membership_id_cards (
    card_id          VARCHAR(50) PRIMARY KEY,            -- format: KAI-{YYYY}-{Random}
    user_id          VARCHAR(50) NOT NULL UNIQUE,
    region_id        VARCHAR(50) NOT NULL,
    issued_date      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status           VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active' | 'revoked'
    qr_version       VARCHAR(10) NOT NULL DEFAULT 'v1',
    qr_code_data     TEXT NOT NULL,                      -- {version}|{card_id}|{checksum}
    digital_format   JSONB NOT NULL,                     -- {design, user_name, avatar_url, qr_url}
    physical_ordered BOOLEAN NOT NULL DEFAULT FALSE,
    order_id         VARCHAR(50) NULL,
    revoked_reason   TEXT NULL,                          -- alasan revoke (identitas only)
    revoked_at       TIMESTAMPTZ NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_card_user   FOREIGN KEY (user_id)   REFERENCES users(id)    ON DELETE CASCADE,
    CONSTRAINT fk_card_region FOREIGN KEY (region_id) REFERENCES regions(id),
    CONSTRAINT fk_card_order  FOREIGN KEY (order_id)  REFERENCES orders(id),
    CONSTRAINT chk_card_status CHECK (status IN ('active', 'revoked'))
);

CREATE INDEX idx_card_user   ON membership_id_cards (user_id);
CREATE INDEX idx_card_status ON membership_id_cards (status);
CREATE INDEX idx_card_region ON membership_id_cards (region_id);
```

### Catatan field
- **TIDAK ADA `plan`** — di-resolve live dari Subscription saat verifikasi.
- **TIDAK ADA `expiry_date`** — identitas tidak kadaluarsa. Status hanya `active`/`revoked`.
- **`qr_version`** — memungkinkan format QR berkembang tanpa mematikan kartu lama.
- **`revoked_reason` / `revoked_at`** — audit revocation (alasan identitas only, bukan pembayaran).

---

## 2. `card_scan_events`

Event stream setiap scan kartu. Fondasi untuk attendance, redemption, loyalty, analytics, fraud detection. **Append-only.**

```sql
CREATE TABLE card_scan_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id      VARCHAR(50) NOT NULL,
    user_id      VARCHAR(50) NOT NULL,                   -- denormalized untuk query cepat
    scanned_by   VARCHAR(50) NULL,                       -- partner_id / event_gate_id / admin_id / null
    context      VARCHAR(20) NOT NULL,                   -- directory | event | admin | public_share | loyalty
    context_ref  VARCHAR(50) NULL,                       -- event_id / merchant_id / dll
    result       VARCHAR(20) NOT NULL,                   -- valid | revoked | invalid_checksum | not_found
    metadata     JSONB NULL,                             -- {benefit_applied, value, plan_at_scan, ...}
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_scan_card FOREIGN KEY (card_id) REFERENCES membership_id_cards(card_id),
    CONSTRAINT chk_scan_context CHECK (context IN ('directory','event','admin','public_share','loyalty')),
    CONSTRAINT chk_scan_result  CHECK (result IN ('valid','revoked','invalid_checksum','not_found'))
);

CREATE INDEX idx_scan_card       ON card_scan_events (card_id);
CREATE INDEX idx_scan_user       ON card_scan_events (user_id);
CREATE INDEX idx_scan_context    ON card_scan_events (context, context_ref);
CREATE INDEX idx_scan_created    ON card_scan_events (created_at);
CREATE INDEX idx_scan_scanned_by ON card_scan_events (scanned_by);
```

### Contoh records

```jsonc
// Scan di toko partner, member Pro dapat diskon
{
  "card_id": "KAI-2026-001",
  "user_id": "usr_123",
  "scanned_by": "merchant_045",
  "context": "directory",
  "context_ref": "mch_045",
  "result": "valid",
  "metadata": { "plan_at_scan": "pro", "benefit_applied": "directory_discount", "value": 15 }
}

// Scan di gate event
{
  "card_id": "KAI-2026-001",
  "user_id": "usr_123",
  "scanned_by": "gate_jkt_01",
  "context": "event",
  "context_ref": "evt_2026_summer",
  "result": "valid",
  "metadata": { "registered": true }
}

// Kartu revoked dicoba dipakai (fraud signal)
{
  "card_id": "KAI-2026-009",
  "user_id": "usr_999",
  "scanned_by": "merchant_012",
  "context": "directory",
  "result": "revoked",
  "metadata": { "reason": "user_banned" }
}
```

---

## 3. Query Patterns Umum

### Verifikasi kartu (resolve plan live)
```sql
-- 1. Ambil kartu
SELECT card_id, user_id, region_id, status
FROM membership_id_cards
WHERE card_id = $1 AND status = 'active';

-- 2. Resolve plan live (via modul Subscription)
SELECT plan FROM subscriptions WHERE user_id = $2 AND status = 'active';

-- 3. Catat scan (selalu, termasuk gagal)
INSERT INTO card_scan_events (card_id, user_id, scanned_by, context, context_ref, result, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7);
```

### Attendance event (future)
```sql
SELECT user_id, created_at
FROM card_scan_events
WHERE context = 'event' AND context_ref = $1 AND result = 'valid'
ORDER BY created_at;
```

### Laporan redemption partner (future)
```sql
SELECT context_ref AS merchant_id,
       COUNT(*) AS total_redemptions,
       SUM((metadata->>'value')::numeric) AS total_value
FROM card_scan_events
WHERE context = 'directory' AND result = 'valid'
  AND created_at BETWEEN $1 AND $2
GROUP BY context_ref;
```

### Fraud detection (kartu dipakai di 2 tempat berjauhan dalam waktu singkat)
```sql
SELECT card_id, COUNT(DISTINCT scanned_by) AS distinct_scanners
FROM card_scan_events
WHERE created_at > NOW() - INTERVAL '5 minutes' AND result = 'valid'
GROUP BY card_id
HAVING COUNT(DISTINCT scanned_by) > 1;
```

---

## 4. Migration Notes (dari v1.0)

Jika sudah ada tabel v1.0 dengan kolom `plan` dan `expiry_date`:

```sql
-- Hapus kolom yang sudah tidak dipakai
ALTER TABLE membership_id_cards DROP COLUMN IF EXISTS plan;
ALTER TABLE membership_id_cards DROP COLUMN IF EXISTS expiry_date;

-- Tambah kolom baru
ALTER TABLE membership_id_cards ADD COLUMN IF NOT EXISTS qr_version VARCHAR(10) NOT NULL DEFAULT 'v1';
ALTER TABLE membership_id_cards ADD COLUMN IF NOT EXISTS revoked_reason TEXT NULL;
ALTER TABLE membership_id_cards ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMPTZ NULL;

-- Normalisasi status 'expired' → 'active' (identitas tidak kadaluarsa)
UPDATE membership_id_cards SET status = 'active' WHERE status = 'expired';

-- Regenerate qr_code_data ke format versioned (jalankan via script aplikasi)
-- Format baru: v1|{card_id}|{SHA256(v1 + card_id + secret_salt)}
```

> **Penting:** regenerasi QR mengubah QR pada kartu **digital**. Kartu **fisik** yang sudah dicetak dengan QR v1 lama tetap valid jika checksum-nya cocok — inilah gunanya `qr_version` (backend handle per versi).
