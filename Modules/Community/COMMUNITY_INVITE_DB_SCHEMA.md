# Database Schema — Community Invite & Share Link

> **Stack:** Golang + PostgreSQL
> **Berdasarkan:** `COMMUNITY_INVITE_RULES.md`
> **Module:** Community (sub-fitur: Undangan & Share Link)
> **Dibuat:** 2026-07-03

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [PostgreSQL DDL Schema](#postgresql-ddl-schema)
3. [Golang Structs](#golang-structs)
4. [Migrations](#migrations)
5. [Enum Values](#enum-values)
6. [Query Examples](#query-examples)

---

## Overview Relasi

```
communities (dari modul Community)
  ├── community_invitations              (undangan in-app targeted)
  ├── community_invite_links             (share link / join code)
  │     └── community_invite_link_redemptions   (audit siapa join via link)
  └── (permission check: manage_members via Role-Permission)

users (dari modul Auth)
  ├── community_invitations (invitee_id, invited_by)
  ├── community_invite_links (created_by)
  └── community_invite_link_redemptions (user_id)
```

---

## PostgreSQL DDL Schema

```sql
-- ============================================================================
-- COMMUNITY INVITE & SHARE LINK — DATABASE SCHEMA
-- Stack: PostgreSQL 13+
-- Created: 2026-07-03
-- Scope: sub-fitur modul Community (member-facing)
-- ============================================================================


-- ============================================================================
-- 1. COMMUNITY_INVITATIONS
-- Undangan in-app ke user terdaftar tertentu. Accept = bypass approval
-- (termasuk untuk komunitas private). Di-gate permission manage_members.
-- ============================================================================

CREATE TABLE community_invitations (
  id           UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  community_id UUID          NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
  invitee_id   UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  invited_by   UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  message      TEXT          NULL,
  status       VARCHAR(10)   NOT NULL DEFAULT 'pending',  -- pending|accepted|rejected|expired|cancelled
  expires_at   TIMESTAMPTZ   NULL,                        -- default diisi app now()+7d; NULL = tak kedaluwarsa
  responded_at TIMESTAMPTZ   NULL,

  created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_inv_status CHECK (status IN ('pending','accepted','rejected','expired','cancelled'))
);

-- Satu undangan PENDING per (komunitas, invitee) — partial unique index
CREATE UNIQUE INDEX uq_inv_pending
  ON community_invitations (community_id, invitee_id)
  WHERE status = 'pending';

CREATE INDEX idx_inv_invitee    ON community_invitations (invitee_id, status);
CREATE INDEX idx_inv_community  ON community_invitations (community_id, status);
CREATE INDEX idx_inv_expires    ON community_invitations (expires_at) WHERE status = 'pending';


-- ============================================================================
-- 2. COMMUNITY_INVITE_LINKS
-- Share link / join code. Bisa langsung-join atau lewat approval (flag).
-- Kontrol opsional: expiry, max_uses. Revocable via is_active.
-- ============================================================================

CREATE TABLE community_invite_links (
  id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
  community_id      UUID          NOT NULL REFERENCES communities(id) ON DELETE CASCADE,
  created_by        UUID          NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

  code              VARCHAR(32)   NOT NULL UNIQUE,           -- token pendek untuk URL
  requires_approval BOOLEAN       NOT NULL DEFAULT FALSE,    -- true = join-request; false = langsung join
  expires_at        TIMESTAMPTZ   NULL,                      -- NULL = tak kedaluwarsa
  max_uses          INTEGER       NULL,                      -- NULL = tak terbatas
  use_count         INTEGER       NOT NULL DEFAULT 0,
  is_active         BOOLEAN       NOT NULL DEFAULT TRUE,      -- revocable

  created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
  updated_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_link_maxuses CHECK (max_uses IS NULL OR max_uses > 0)
);

CREATE UNIQUE INDEX uq_link_code    ON community_invite_links (code);
CREATE INDEX idx_link_community     ON community_invite_links (community_id, is_active);


-- ============================================================================
-- 3. COMMUNITY_INVITE_LINK_REDEMPTIONS
-- Audit siapa yang gabung via link mana. Dedup + cegah double-count use_count.
-- ============================================================================

CREATE TABLE community_invite_link_redemptions (
  id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  link_id    UUID         NOT NULL REFERENCES community_invite_links(id) ON DELETE CASCADE,
  user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  result     VARCHAR(10)  NOT NULL,        -- 'joined' | 'requested'
  created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT chk_redeem_result CHECK (result IN ('joined','requested')),
  CONSTRAINT uq_redeem_once UNIQUE (link_id, user_id)   -- satu user redeem satu link sekali
);

CREATE INDEX idx_redeem_link ON community_invite_link_redemptions (link_id);
```

---

## Golang Structs

```go
package community

import "time"

// ---------- Enums ----------

type InvitationStatus string
type RedemptionResult string

const (
	InvPending   InvitationStatus = "pending"
	InvAccepted  InvitationStatus = "accepted"
	InvRejected  InvitationStatus = "rejected"
	InvExpired   InvitationStatus = "expired"
	InvCancelled InvitationStatus = "cancelled"

	RedeemJoined    RedemptionResult = "joined"
	RedeemRequested RedemptionResult = "requested"
)

// ---------- 1. CommunityInvitation ----------

type CommunityInvitation struct {
	ID          string           `json:"id" db:"id"`
	CommunityID string           `json:"community_id" db:"community_id"`
	InviteeID   string           `json:"invitee_id" db:"invitee_id"`
	InvitedBy   string           `json:"invited_by" db:"invited_by"`

	Message     *string          `json:"message,omitempty" db:"message"`
	Status      InvitationStatus `json:"status" db:"status"`
	ExpiresAt   *time.Time       `json:"expires_at,omitempty" db:"expires_at"`
	RespondedAt *time.Time       `json:"responded_at,omitempty" db:"responded_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (i *CommunityInvitation) IsAcceptable(now time.Time) bool {
	if i.Status != InvPending {
		return false
	}
	if i.ExpiresAt != nil && now.After(*i.ExpiresAt) {
		return false
	}
	return true
}

// ---------- 2. CommunityInviteLink ----------

type CommunityInviteLink struct {
	ID               string     `json:"id" db:"id"`
	CommunityID      string     `json:"community_id" db:"community_id"`
	CreatedBy        string     `json:"created_by" db:"created_by"`

	Code             string     `json:"code" db:"code"`
	RequiresApproval bool       `json:"requires_approval" db:"requires_approval"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	MaxUses          *int       `json:"max_uses,omitempty" db:"max_uses"`
	UseCount         int        `json:"use_count" db:"use_count"`
	IsActive         bool       `json:"is_active" db:"is_active"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// IsLive: link masih bisa dipakai?
func (l *CommunityInviteLink) IsLive(now time.Time) bool {
	if !l.IsActive {
		return false
	}
	if l.ExpiresAt != nil && now.After(*l.ExpiresAt) {
		return false
	}
	if l.MaxUses != nil && l.UseCount >= *l.MaxUses {
		return false
	}
	return true
}

// ---------- 3. CommunityInviteLinkRedemption ----------

type CommunityInviteLinkRedemption struct {
	ID        string           `json:"id" db:"id"`
	LinkID    string           `json:"link_id" db:"link_id"`
	UserID    string           `json:"user_id" db:"user_id"`
	Result    RedemptionResult `json:"result" db:"result"`
	CreatedAt time.Time        `json:"created_at" db:"created_at"`
}
```

---

## Migrations

Prasyarat: tabel `communities`, `community_members`, `community_join_requests`, `users` sudah ada.

### `migrations/2026070310_create_community_invitations.up.sql`
```sql
-- community_invitations (DDL §1) + partial unique index + indexes
```

### `migrations/2026070311_create_community_invite_links.up.sql`
```sql
-- community_invite_links        (DDL §2)
-- community_invite_link_redemptions (DDL §3)
-- + indexes, urutan sesuai FK
```

### `*.down.sql`
```sql
-- 2026070311 down:
DROP TABLE IF EXISTS community_invite_link_redemptions CASCADE;
DROP TABLE IF EXISTS community_invite_links CASCADE;
-- 2026070310 down:
DROP TABLE IF EXISTS community_invitations CASCADE;
```

> **Trigger `updated_at`:** pakai `update_timestamp()` yang sudah ada, pasang `BEFORE UPDATE` di `community_invitations` & `community_invite_links`.
>
> **Cleanup komunitas:** semua FK `ON DELETE CASCADE` ke `communities(id)`, jadi hapus komunitas otomatis bersih (Rule 7).
>
> **Redeem atomik:** operasi redeem (validasi link → insert member/request → insert redemption → `use_count++`) harus dalam **satu transaksi**, dengan `SELECT ... FOR UPDATE` pada baris link agar `max_uses` tidak race.

---

## Enum Values

| Tabel | Kolom | Nilai |
|---|---|---|
| `community_invitations` | `status` | `pending`, `accepted`, `rejected`, `expired`, `cancelled` |
| `community_invite_link_redemptions` | `result` | `joined`, `requested` |

---

## Query Examples

### Q1: Undangan masuk untuk seorang user (pending & belum expired)
```sql
SELECT ci.*, c.name AS community_name, c.avatar AS community_avatar
FROM community_invitations ci
JOIN communities c ON c.id = ci.community_id
WHERE ci.invitee_id = $1
  AND ci.status = 'pending'
  AND (ci.expires_at IS NULL OR ci.expires_at > NOW())
ORDER BY ci.created_at DESC;
```

### Q2: Cegah undangan duplikat (ditangani partial unique index)
```sql
INSERT INTO community_invitations (community_id, invitee_id, invited_by, message, expires_at)
VALUES ($1, $2, $3, $4, NOW() + INTERVAL '7 days')
ON CONFLICT (community_id, invitee_id) WHERE status = 'pending'
DO NOTHING
RETURNING *;
```

### Q3: Resolve & lock link saat redeem (anti-race)
```sql
SELECT *
FROM community_invite_links
WHERE code = $1
FOR UPDATE;   -- lalu cek is_active/expires_at/max_uses di app (IsLive)
```

### Q4: Naikkan use_count setelah redeem sukses
```sql
UPDATE community_invite_links
SET use_count = use_count + 1, updated_at = NOW()
WHERE id = $1;
```

### Q5: Statistik link untuk pengelola
```sql
SELECT l.id, l.code, l.requires_approval, l.expires_at, l.max_uses, l.use_count, l.is_active,
       COUNT(r.id) FILTER (WHERE r.result = 'joined')    AS joined_count,
       COUNT(r.id) FILTER (WHERE r.result = 'requested') AS requested_count
FROM community_invite_links l
LEFT JOIN community_invite_link_redemptions r ON r.link_id = l.id
WHERE l.community_id = $1
GROUP BY l.id
ORDER BY l.created_at DESC;
```

### Q6: Expire undangan lewat waktu (scheduler)
```sql
UPDATE community_invitations
SET status = 'expired', updated_at = NOW()
WHERE status = 'pending'
  AND expires_at IS NOT NULL
  AND expires_at <= NOW();
```

---

*Skema untuk sub-fitur Undangan & Share Link modul Community. Aturan di `COMMUNITY_INVITE_RULES.md`, endpoint di `API_SPEC_COMMUNITY_INVITE_MOBILE.md`.*
