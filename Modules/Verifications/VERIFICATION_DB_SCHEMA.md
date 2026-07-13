# Verification Badge Module — Database Schema (v1.0)

Database schema untuk **Verification Badge** module KAI App. Badge keaslian (centang) buat **User** & **Merchant**, diajukan manual, di-review Superadmin. Konsisten dengan konvensi project: `UUID` PK (`gen_random_uuid()`), `TIMESTAMPTZ`, append-only audit, rules-as-data config, resolve-live + cache.

---

## Daftar Isi

1. Overview Relasi
2. `verifications` — request + current state (polymorphic)
3. `verification_events` — audit log (append-only)
4. `verification_requirements` — config syarat dokumen (rules-as-data)
5. ALTER: cache `is_verified` di `users` & `merchants`
6. Go Models & Constants
7. Seed Data

---

## 1. Overview Relasi

```
users (modul auth)
  ├── verifications (polymorphic: entity_type='user',  entity_id=users.id)
  └── is_verified (cache column, ALTER)

merchants (modul directory)
  ├── verifications (polymorphic: entity_type='merchant', entity_id=merchants.id)
  └── is_verified (cache column, ALTER)

verifications (1) ──< verification_events (N)   ← audit append-only

verification_requirements (config, per entity_type)
  └── dibaca saat validasi dokumen pengajuan
```

> **Polymorphic:** `verifications` ga pakai FK ke `users`/`merchants` (beda tabel target). Integritas dijaga di app-layer + `entity_type` check. Badge type = `entity_type` (satu field, dua produk).
>
> **Resolve live:** sumber kebenaran badge = ada tidaknya row `verifications` `status='approved'`. Kolom `is_verified` di `users`/`merchants` cuma **cache** buat query listing cepat, di-maintain app-layer saat approve/revoke.

---

## PostgreSQL DDL Schema

```sql
-- ============================================================================
-- VERIFICATION BADGE MODULE DATABASE SCHEMA
-- Stack: PostgreSQL 13+
-- Created: 2026-07-13
-- ============================================================================

-- ============================================================================
-- 2. VERIFICATIONS
-- Satu row per PENGAJUAN. Lifecycle: pending -> approved | rejected | revoked.
-- Resubmit setelah rejected/revoked = row BARU (bukan update row lama).
-- ============================================================================
CREATE TABLE verifications (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type       VARCHAR(20) NOT NULL,          -- 'user' | 'merchant' (= badge type)
    entity_id         UUID NOT NULL,                 -- users.id | merchants.id (polymorphic, no FK)
    status            VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending'|'approved'|'rejected'|'revoked'

    -- Pengajuan
    submitted_by      UUID NOT NULL REFERENCES users(id), -- pemohon (owner user / merchant owner)
    documents         JSONB NOT NULL DEFAULT '[]',   -- [{doc_type, url, uploaded_at}]
    note              TEXT,                           -- catatan opsional dari pemohon

    -- Review (Superadmin)
    reviewed_by       UUID REFERENCES users(id),      -- superadmin yg approve/reject
    reviewed_at       TIMESTAMPTZ,
    rejection_reason  TEXT,                           -- WAJIB saat status='rejected'

    -- Revoke (Superadmin)
    revoked_by        UUID REFERENCES users(id),
    revoked_at        TIMESTAMPTZ,
    revoke_reason     TEXT,                           -- WAJIB saat status='revoked'

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_verif_entity_type CHECK (entity_type IN ('user','merchant')),
    CONSTRAINT chk_verif_status      CHECK (status IN ('pending','approved','rejected','revoked'))
);

-- Indexes
CREATE INDEX idx_verifications_entity      ON verifications (entity_type, entity_id);
CREATE INDEX idx_verifications_status      ON verifications (status);
CREATE INDEX idx_verifications_submitted   ON verifications (submitted_by);
CREATE INDEX idx_verifications_queue       ON verifications (created_at DESC) WHERE status = 'pending';

-- Max 1 pengajuan pending per entitas (anti-spam)
CREATE UNIQUE INDEX uq_verif_one_pending
    ON verifications (entity_type, entity_id) WHERE status = 'pending';
-- Max 1 badge approved aktif per entitas
CREATE UNIQUE INDEX uq_verif_one_approved
    ON verifications (entity_type, entity_id) WHERE status = 'approved';


-- ============================================================================
-- 3. VERIFICATION_EVENTS
-- Audit log APPEND-ONLY. Tiap aksi (submit/approve/reject/revoke/resubmit)
-- tercatat permanen. TIDAK PERNAH di-update / di-delete.
-- ============================================================================
CREATE TABLE verification_events (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    verification_id   UUID NOT NULL REFERENCES verifications(id) ON DELETE CASCADE,
    action            VARCHAR(20) NOT NULL,           -- 'submitted'|'approved'|'rejected'|'revoked'
    actor_id          UUID NOT NULL REFERENCES users(id),
    reason            TEXT,                           -- alasan reject/revoke bila ada
    metadata          JSONB DEFAULT '{}',             -- snapshot ringkas (dokumen count, dll)
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_verif_event_action
        CHECK (action IN ('submitted','approved','rejected','revoked'))
);

CREATE INDEX idx_verif_events_verification ON verification_events (verification_id, created_at);
CREATE INDEX idx_verif_events_actor        ON verification_events (actor_id);


-- ============================================================================
-- 4. VERIFICATION_REQUIREMENTS
-- Rules-as-data. Syarat dokumen per entity_type, bisa diubah Superadmin
-- tanpa deploy. match_mode: 'any_of' (min 1) | 'all_of' (wajib semua).
-- ============================================================================
CREATE TABLE verification_requirements (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type       VARCHAR(20) NOT NULL UNIQUE,    -- 'user' | 'merchant'
    match_mode        VARCHAR(10) NOT NULL,           -- 'any_of' | 'all_of'
    accepted_docs     JSONB NOT NULL,                 -- [{key,label,required,sensitive}]
    min_documents     INT NOT NULL DEFAULT 1,         -- utk any_of: minimal berapa dokumen
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    updated_by        UUID REFERENCES users(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_verif_req_entity CHECK (entity_type IN ('user','merchant')),
    CONSTRAINT chk_verif_req_mode   CHECK (match_mode IN ('any_of','all_of'))
);


-- ============================================================================
-- 5. ALTER: cache is_verified (additive, non-breaking)
-- ============================================================================
ALTER TABLE users     ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE merchants ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_users_is_verified     ON users (is_verified)     WHERE is_verified = TRUE;
CREATE INDEX idx_merchants_is_verified ON merchants (is_verified) WHERE is_verified = TRUE;
```

---

## 6. Go Models & Constants

```go
// Verification — satu pengajuan + current state
type Verification struct {
    ID              string     `db:"id"               json:"id"`
    EntityType      string     `db:"entity_type"      json:"entity_type"` // user|merchant
    EntityID        string     `db:"entity_id"        json:"entity_id"`
    Status          string     `db:"status"           json:"status"`
    SubmittedBy     string     `db:"submitted_by"     json:"submitted_by"`
    Documents       JSONB      `db:"documents"        json:"documents"`
    Note            *string    `db:"note"             json:"note,omitempty"`
    ReviewedBy      *string    `db:"reviewed_by"      json:"reviewed_by,omitempty"`
    ReviewedAt      *time.Time `db:"reviewed_at"      json:"reviewed_at,omitempty"`
    RejectionReason *string    `db:"rejection_reason" json:"rejection_reason,omitempty"`
    RevokedBy       *string    `db:"revoked_by"       json:"revoked_by,omitempty"`
    RevokedAt       *time.Time `db:"revoked_at"       json:"revoked_at,omitempty"`
    RevokeReason    *string    `db:"revoke_reason"    json:"revoke_reason,omitempty"`
    CreatedAt       time.Time  `db:"created_at"       json:"created_at"`
    UpdatedAt       time.Time  `db:"updated_at"       json:"updated_at"`
}

type VerificationEvent struct {
    ID             string    `db:"id"              json:"id"`
    VerificationID string    `db:"verification_id" json:"verification_id"`
    Action         string    `db:"action"          json:"action"`
    ActorID        string    `db:"actor_id"        json:"actor_id"`
    Reason         *string   `db:"reason"          json:"reason,omitempty"`
    Metadata       JSONB     `db:"metadata"        json:"metadata"`
    CreatedAt      time.Time `db:"created_at"      json:"created_at"`
}

type VerificationRequirement struct {
    ID           string    `db:"id"            json:"id"`
    EntityType   string    `db:"entity_type"   json:"entity_type"`
    MatchMode    string    `db:"match_mode"    json:"match_mode"`
    AcceptedDocs JSONB     `db:"accepted_docs" json:"accepted_docs"`
    MinDocuments int       `db:"min_documents" json:"min_documents"`
    IsActive     bool      `db:"is_active"     json:"is_active"`
    UpdatedBy    *string   `db:"updated_by"    json:"updated_by,omitempty"`
    CreatedAt    time.Time `db:"created_at"    json:"created_at"`
    UpdatedAt    time.Time `db:"updated_at"    json:"updated_at"`
}

const (
    VerifEntityUser     = "user"
    VerifEntityMerchant = "merchant"

    VerifStatusPending  = "pending"
    VerifStatusApproved = "approved"
    VerifStatusRejected = "rejected"
    VerifStatusRevoked  = "revoked"

    VerifActionSubmitted = "submitted"
    VerifActionApproved  = "approved"
    VerifActionRejected  = "rejected"
    VerifActionRevoked   = "revoked"

    VerifMatchAnyOf = "any_of"
    VerifMatchAllOf = "all_of"
)
```

---

## 7. Seed Data

```sql
-- USER: any_of, minimal 1 dokumen. KTP OPSIONAL (bukan wajib).
INSERT INTO verification_requirements (entity_type, match_mode, min_documents, accepted_docs) VALUES
('user', 'any_of', 1, '[
  {"key":"kta_kai",        "label":"KTA KAI (kartu anggota)",         "required":false, "sensitive":true},
  {"key":"kta_other",      "label":"KTA organisasi lain",             "required":false, "sensitive":true},
  {"key":"gov_id",         "label":"KTP / Paspor / SIM (opsional)",   "required":false, "sensitive":true},
  {"key":"membership_proof","label":"Bukti jabatan / keanggotaan",    "required":false, "sensitive":false},
  {"key":"social_proof",   "label":"Akun sosial media terverifikasi", "required":false, "sensitive":false}
]');

-- MERCHANT: all_of, legalitas usaha wajib.
INSERT INTO verification_requirements (entity_type, match_mode, min_documents, accepted_docs) VALUES
('merchant', 'all_of', 2, '[
  {"key":"business_license","label":"NIB / akta / izin usaha",  "required":true,  "sensitive":true},
  {"key":"owner_id",        "label":"Identitas pemilik",         "required":true,  "sensitive":true},
  {"key":"address_proof",   "label":"Bukti alamat usaha",        "required":false, "sensitive":false},
  {"key":"business_social", "label":"Akun sosmed bisnis",        "required":false, "sensitive":false}
]');
```

---

## Catatan Implementasi

- **Maintain cache `is_verified`:** di app-layer, dalam transaksi yang sama saat set `status='approved'` (→ TRUE) atau `status='revoked'` (→ FALSE). Bisa juga trigger, tapi app-layer lebih eksplisit & konsisten dengan modul lain.
- **Validasi dokumen:** saat submit, baca `verification_requirements` by `entity_type`. `any_of` → cek jumlah dokumen ≥ `min_documents`. `all_of` → cek semua yang `required=true` ada.
- **Dokumen sensitif** (`sensitive=true`): simpan di storage private, URL signed & short-lived, akses cuma Superadmin. Jangan pernah expose ke response member-facing.
- **Guard prasyarat:** sebelum insert, cek `users.status='active'` (user) atau `merchants.status='published'` (merchant).

---

*Verification Badge Module DB Schema v1.0 — KAI App. Step 2 dari pipeline. Last updated: 2026-07-13*
