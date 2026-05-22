# KAI Membership ID Card Specification

## Overview
ID Card Keanggotaan adalah **bukti formal** yang diberikan kepada setiap member KAI. Card ini tersedia dalam dua format:
- **Digital** (in-app, shareable, downloadable as PDF)
- **Fisik** (printed card, QR-enabled, shipped to user)

Setiap member (baik standard maupun pro) mendapat ID card saat registration.

---

## Data Model: Membership ID Card

### Tabel: `membership_id_cards`

```
+──────────────────┬──────────────────┬───────────────────────────────────────+
│ Field            │ Type             │ Keterangan                            │
+──────────────────┬──────────────────┬───────────────────────────────────────+
│ card_id          │ String (PK)      │ Unique ID, format: KAI-{YYYY}-{Random}
│                  │                  │ Contoh: KAI-2026-USR001              │
│ user_id          │ String (FK)      │ Reference ke User.id                  │
│ plan             │ Enum             │ Snapshot dari user.plan:              │
│                  │                  │ - standard                            │
│                  │                  │ - pro                                 │
│ region_id        │ String (FK)      │ Reference ke Region saat card issued  │
│ issued_date      │ DateTime (UTC)   │ Waktu card dibuat (auto-set)         │
│ expiry_date      │ DateTime (nullable) │ Null = permanent                    │
│                  │                  │ Set = auto-expire setelah periode     │
│ qr_code_data     │ String           │ Encoded QR: {card_id}|{user_id}|{plan}│
│                  │                  │ Untuk offline verification            │
│ status           │ Enum             │ - active (valid)                      │
│                  │                  │ - revoked (dibatalkan)                │
│                  │                  │ - expired (kadaluarsa)                │
│ digital_format   │ JSON             │ Design & data untuk digital variant   │
│                  │                  │ {design, user_name, avatar_url,      │
│                  │                  │  plan_badge, qr_url}                 │
│ physical_ordered │ Boolean          │ Apakah sudah order fisik?             │
│ order_id         │ String (FK, null)│ Reference ke fulfillment order        │
│ created_at       │ DateTime         │ Saat record dibuat                    │
│ updated_at       │ DateTime         │ Saat terakhir di-update              │
+──────────────────┬──────────────────┬───────────────────────────────────────+
```

---

## Lifecycle

### 1. Card Creation
- **Trigger**: User menyelesaikan registration
- **Action**: 
  - System auto-generate unique `card_id` 
  - Snapshot user.plan ke dalam card.plan
  - Set `issued_date = now()`
  - Generate QR code data
  - Create digital_format JSON
  - Set `status = active`

### 2. Digital Card Access
User dapat:
- **View** di Profile > Membership ID Card
- **Download** sebagai PDF
- **Share** QR code link ke teman (untuk verifikasi online)

### 3. Physical Card Request
- User bisa request kartu fisik dari dalam app (atau auto-shipped)
- Trigger fulfillment order (print + ship)
- Card dikirim ke alamat user
- Set `physical_ordered = true`

### 4. Verification (Digital & Physical)
- **Toko partner** scan QR → backend verify `status` & `plan`
- **Entry gate** (event) scan → confirm registered member
- **Admin** scan → view member details di backend

### 5. Plan Upgrade/Downgrade
- **Decision**: Satu card selamanya, atau new card?
  - **Option A**: Update existing card (simple, satu card per user)
  - **Option B**: Create new card (track history, cleaner audit)
  - Rekomendasi: **Option A** untuk simplicity

### 6. Card Expiry / Revocation
- **Auto-expire**: Jika `expiry_date` set dan sudah lewat → status = expired
- **Manual revoke**: Admin dapat revoke card (status = revoked)
  - Trigger: user banned, membership cancelled, etc.

---

## Digital Card Format

### UI Display (Mobile)
```
┌─────────────────────────────┐
│  KAI MEMBER ID CARD         │
├─────────────────────────────┤
│                             │
│   [User Avatar / Photo]     │
│                             │
│   John Doe                  │
│   Member ID: KAI-2026-001   │
│   Plan: PRO                 │
│   Valid Until: Dec 31, 2027 │
│                             │
│   [QR Code]                 │
│                             │
│   Status: Active            │
├─────────────────────────────┤
│ [ Download PDF ] [ Share ]  │
└─────────────────────────────┘
```

### Digital Attributes
- **Responsive**: Optimized untuk mobile (800x600px)
- **QR Code**: Scannable (contains: card_id, user_id, plan, checksum)
- **PDF Export**: Self-contained, no internet needed
- **Share Link**: Generates shareable URL (e.g., kai.app/member/KAI-2026-001)

---

## Physical Card Format

### Design Specifications
```
FRONT SIDE:
┌─────────────────────────────────┐
│         KAI KOREA               │ (top)
│   Asosiasi Indonesia            │
│                                 │
│   [User Photo - 30x40mm]        │
│   John Doe                      │
│   Member since: 2026-05-20      │
│                                 │
│   Card ID: KAI-2026-001         │
│   Plan: PRO                     │
│   Valid: 2026-05-20 to 2027-... │
│                                 │
│   [QR Code - 20x20mm]           │
└─────────────────────────────────┘

BACK SIDE:
┌─────────────────────────────────┐
│   KAI Benefits (Pro Member):    │
│   ✓ Exclusive content access    │
│   ✓ Directory discounts         │
│   ✓ Event early booking         │
│   ✓ Priority support            │
│                                 │
│   Validity: Scan QR to verify   │
│   Contact: support@kai.app      │
│   www.kai.app                   │
└─────────────────────────────────┘
```

### Physical Card Specs
- **Size**: ISO/IEC 7810 ID-1 (85.6 × 53.98 mm) — standar
- **Material**: PVC or Polycarbonate (durability + QR embedding)
- **Finish**: Matte or glossy (user choice?)
- **QR Code**: Embedded/printed, contains verification data
- **Photo**: High-res user photo (minimum 200dpi)

---

## Use Cases & Verification Flow

### Use Case 1: Directory / Toko Partner
```
Member → Show physical card to shop staff
  ↓
Staff → Scan QR code with app
  ↓
Backend → Verify: card_id valid? status=active? plan=?
  ↓
App → Display: "Member John Doe - PRO - Valid"
  ↓
Staff → Apply discount/benefit based on plan
```

### Use Case 2: Event Entry
```
Member → Arrive at event gate
  ↓
Gate staff → Scan physical card QR
  ↓
Backend → Check: member registered for this event? status=active?
  ↓
System → Grant/deny entry
```

### Use Case 3: Digital Share (Online Verification)
```
Member → Share digital QR link: kai.app/member/KAI-2026-001
  ↓
Recipient → Open link → See member's public profile
  ↓
Can verify: "This person is a pro member of KAI"
```

### Use Case 4: Admin Panel
```
Admin → Member Management → Search "John Doe"
  ↓
View → Card status, issued_date, plan, physical_ordered status
  ↓
Actions → Revoke card (if banned), extend expiry, view history
```

---

## API Endpoints (Proposed)

### 1. Get User's ID Card
```
GET /api/v1/mobile/members/{user_id}/id-card
Response: {
  "card_id": "KAI-2026-001",
  "plan": "pro",
  "issued_date": "2026-05-20T00:00:00Z",
  "expiry_date": "2027-05-20T00:00:00Z",
  "status": "active",
  "digital_format": {
    "design": "standard-2026",
    "user_name": "John Doe",
    "avatar_url": "https://...",
    "plan_badge": "PRO",
    "qr_url": "https://kai.app/verify/KAI-2026-001"
  },
  "physical_ordered": false,
  "qr_code_data": "KAI-2026-001|usr_123|pro|abc123"
}
```

### 2. Download Digital Card as PDF
```
GET /api/v1/mobile/members/{user_id}/id-card/download-pdf
Response: Binary PDF file
```

### 3. Request Physical Card
```
POST /api/v1/mobile/members/{user_id}/id-card/request-physical
Body: {
  "shipping_address": "...",
  "phone": "..."
}
Response: {
  "order_id": "ORD-001",
  "estimated_delivery": "2026-05-30",
  "status": "processing"
}
```

### 4. Verify Card (for QR scan)
```
POST /api/v1/public/verify-card
Body: {
  "card_id": "KAI-2026-001",
  "qr_data": "KAI-2026-001|usr_123|pro|abc123"
}
Response: {
  "valid": true,
  "user_name": "John Doe",
  "plan": "pro",
  "status": "active",
  "issued_date": "2026-05-20"
}
```

### 5. Admin: List Cards
```
GET /api/v1/admin/id-cards?region=jakarta&status=active&plan=pro
Response: [
  {
    "card_id": "KAI-2026-001",
    "user_name": "John Doe",
    "plan": "pro",
    "status": "active",
    "issued_date": "2026-05-20",
    "physical_ordered": true
  },
  ...
]
```

### 6. Admin: Revoke Card
```
POST /api/v1/admin/id-cards/{card_id}/revoke
Body: {
  "reason": "User banned"
}
Response: {
  "card_id": "KAI-2026-001",
  "status": "revoked",
  "revoked_at": "2026-05-22"
}
```

---

## Open Decisions

1. **Card Lifetime Strategy**
   - One card per user forever (update on plan change)?
   - New card on every plan upgrade?
   - **Recommendation**: One card, update fields on change.

2. **Expiry Policy**
   - Permanent (never expire)?
   - Auto-expire yearly?
   - Auto-expire on plan downgrade?

3. **Physical Card Cost**
   - Free to all members?
   - Charged for shipping only?
   - Pro-only benefit?

4. **QR Verification Offline**
   - QR encodes enough data to verify offline?
   - Or always requires backend call?

5. **Card Replacement**
   - User can request replacement (lost/damaged)?
   - How many replacements per year?

---

## Integration Points

### With Other Modules

**Profile Module**
- Store & display ID card in user profile
- Link to order/fulfillment system for physical card

**Directory Module**
- Toko staff scan card → verify plan → apply benefits

**Event Module**
- Event gate scan card → check registration + membership

**Notification Module**
- Alert user when card expires (30 days before)
- Notify when physical card ships

**Admin Panel**
- Bulk view/manage cards
- Revoke suspicious cards
- Export card usage reports

---

## Security & Validation

### QR Code Content
```
Format: {card_id}|{user_id}|{plan}|{checksum}
Checksum: SHA256(card_id + user_id + plan + secret_salt)

Example:
KAI-2026-001|usr_123|pro|a1b2c3d4e5f6g7h8
```

### Backend Verification
- Always verify QR checksum on scan
- Check `status = active`
- Check `expiry_date >= today` (if set)
- Log all verification attempts (audit trail)

### Digital Share Security
- Shareable link shows public info only (name, plan, valid date)
- No sensitive data in shareable link
- Rate-limit verification requests (prevent enumeration)

---

## Implementation Roadmap

### Phase 1: Digital Card Only
- User registration → auto-create card
- View card in app
- Download PDF
- Share QR link
- Basic verification endpoint

### Phase 2: Physical Card
- Print-on-demand integration
- Shipping integration
- Track fulfillment status
- User request flow

### Phase 3: Enhanced Features
- Card customization (design themes)
- Bulk card management (admin)
- Card analytics (usage, verification frequency)
- Expiry auto-renewal workflows

---

## Database Schema (Example SQL)

```sql
CREATE TABLE membership_id_cards (
  card_id VARCHAR(50) PRIMARY KEY,
  user_id VARCHAR(50) NOT NULL UNIQUE,
  plan VARCHAR(20) NOT NULL, -- 'standard', 'pro'
  region_id VARCHAR(50) NOT NULL,
  issued_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  expiry_date TIMESTAMP NULL,
  qr_code_data TEXT NOT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active', -- 'active', 'revoked', 'expired'
  digital_format JSON NOT NULL,
  physical_ordered BOOLEAN DEFAULT FALSE,
  order_id VARCHAR(50) NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (region_id) REFERENCES regions(id),
  FOREIGN KEY (order_id) REFERENCES orders(id) -- fulfillment
);

CREATE INDEX idx_card_user ON membership_id_cards(user_id);
CREATE INDEX idx_card_status ON membership_id_cards(status);
CREATE INDEX idx_card_expiry ON membership_id_cards(expiry_date);
```

---

## Summary

**ID Card Keanggotaan** adalah aset kunci untuk verifikasi membership & building brand presence. Dual-format (digital + physical) memastikan fleksibilitas akses dan keamanan. QR-based verification memungkinkan integrasi mudah dengan partner ecosystem (toko, event, etc.).

Next: Finalisasi skema data, setup fulfillment integration, dan design digital card UI.
