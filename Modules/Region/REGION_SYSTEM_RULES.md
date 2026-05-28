# Region System — Rules & Use Cases

**Status:** Draft v1  
**Last Updated:** 2026-05-26  
**Module:** Region Management

---

## Daftar Isi

1. [Overview Konsep](#overview-konsep)
2. [Entitas Utama](#entitas-utama)
3. [Hierarki & Scope](#hierarki--scope)
4. [Member Default State](#member-default-state)
5. [Alur Membership](#alur-membership)
6. [Permission Matrix](#permission-matrix)
7. [Invitation Lifecycle](#invitation-lifecycle)
8. [Use Cases](#use-cases)
9. [Status & Transition](#status--transition)

---

## Overview Konsep

Region adalah pembagian geografis platform KAI yang memungkinkan admin lokal mengelola member di wilayah mereka. Sistem ini memiliki dua level:

1. **Platform Level (Superadmin)** — Kelola semua region, assign regional admins
2. **Region Level (Regional Admin)** — Kelola member di region mereka sendiri

Setiap member dapat bergabung **maksimal 1 region aktif** pada waktu bersamaan. Member dapat keluar dari region dan bergabung ke region lain kapan saja.

---

## Entitas Utama

### 1. Region

Mewakili wilayah geografis atau area administratif platform KAI.

```
id:          UUID (primary key)
name:        "KAI Jakarta", "KAI Surabaya", "KAI Bandung"
slug:        "jakarta", "surabaya", "bandung" (URL-friendly unique)
description: "Wilayah Jakarta dan sekitarnya" (max 500 chars)
image_url:   "https://cdn.example.com/regions/jakarta.jpg" (optional)
status:      "active" | "inactive" (superadmin bisa non-aktivkan)
created_by:  UUID (superadmin yang create)
created_at:  timestamp (UTC ISO 8601)
updated_at:  timestamp (UTC ISO 8601)
member_count: integer (denormalized, update setiap ada join/leave)
```

**Rules:**
- Region name harus unique
- Slug harus unique, derived dari name (lowercase, replace space dengan dash)
- Region tidak bisa di-delete, hanya di-deactivate
- Status "inactive" membuat region tidak tampil di listing, tapi existing members tetap tergabung

---

### 2. RegionMembership

Relasi antara user dan region. Merepresentasikan keanggotaan user di region tertentu.

```
id:               UUID (primary key)
user_id:          UUID (FK → users.id)
region_id:        UUID (FK → regions.id)
role:             "admin" | "member"
status:           "active" | "pending_approval" | "rejected"
joined_at:        timestamp (saat user jadi active member)
approval_notes:   string (catatan admin saat approve, optional)
rejection_reason: string (alasan ditolak, optional)
approved_by:      UUID (superadmin/admin yang approve/reject, optional)
created_at:       timestamp (saat record dibuat)
```

**Rules:**
- Satu user hanya boleh punya **satu active membership per region** (unique constraint: user_id + region_id where status='active')
- User bisa punya multiple memberships di region berbeda **asal hanya 1 yang active**
- Status `pending_approval` = member request join, belum diapprove
- Status `rejected` = admin tolak request, user bisa request lagi nanti
- Saat approve: `joined_at` diisi, `status` = "active"
- Saat reject: `rejection_reason` diisi, `status` = "rejected"

---

### 3. RegionInvitation

Undangan dari superadmin atau admin region mengajak user (via email) bergabung.

```
id:           UUID (primary key)
region_id:    UUID (FK → regions.id)
email:        "user@example.com" (email target)
invited_by:   UUID (FK → users.id, superadmin atau admin yang invite)
status:       "pending" | "accepted" | "rejected"
created_at:   timestamp (saat invite dibuat)
expires_at:   timestamp (created_at + 24 jam)
accepted_at:  timestamp (saat user accept invite)
rejected_at:  timestamp (saat user reject invite)
token:        string (unique, untuk email link verification)
```

**Rules:**
- Invitation berlaku 24 jam dari created_at
- Setelah 24 jam, user tidak bisa accept lagi → invitation expire
- Satu email boleh dapat invite ke region berbeda secara bersamaan
- Satu email tidak boleh dapat 2+ invites ke region yang **sama** dalam status pending
- Saat user accept: check apakah email sudah registered
  - Jika registered → auto create RegionMembership (active)
  - Jika belum registered → redirect ke signup, auto join region setelah signup selesai
- Saat user reject: status = "rejected", tidak bisa accept lagi tapi admin bisa invite ulang

---

## Hierarki & Scope

### Role Hierarchy

```
Superadmin (Platform Level)
    ├── Create regions
    ├── Assign regional admins
    ├── Monitor all regions
    └── Can do anything admin region can do

Admin Region (Region Level)
    ├── Invite members via email
    ├── Approve/reject member join requests
    ├── Remove members from region
    ├── View member list in their region
    └── Cannot create/delete/modify region itself

Member (Region Member)
    ├── View region info
    ├── Request join region
    ├── Accept invitations
    ├── Leave region
    └── Cannot manage anything
```

### Scope Definition

**Platform Scope:** Superadmin dan beberapa endpoint terbuka untuk all authenticated users.
- Create region → superadmin only
- List all regions → authenticated users
- Get region detail → authenticated users

**Region Scope:** Admin region bisa manage hanya region mereka sendiri.
- List members di my region → admin region
- Invite member → admin region
- Approve/reject join → admin region
- Remove member → admin region

**User Scope:** Member hanya bisa lihat region mereka sendiri dan region lain untuk browse.
- Get my region info → authenticated user
- Leave my region → authenticated user
- Request join another region → authenticated user
- Accept invitation → user dengan email yang match

---

## Member Default State

Setiap user saat pertama kali register:

```
region_id: null (tidak ada region)
RegionMembership: tidak ada record
Status: "no_region" atau tidak ada status

User dapat:
✅ Browse semua region (GET /regions)
✅ Request join region (POST /regions/{id}/request)
✅ Receive invitations via email
✅ Semua activity di platform tetap jalan tanpa region

User TIDAK harus punya region.
Region adalah opsional — user bisa tetap tidak ada region selamanya.
```

---

## Alur Membership

### Alur 1: Superadmin Invite User sebagai Admin Region

**Flow:**
```
Superadmin pilih region Jakarta
  ↓ [Invite as Admin]
Input email / select user
  ↓
RegionInvitation created
status: "pending"
invited_by: superadmin_id
role (implicit): "admin"
  ↓
User dapat email invitation: "Anda diundang jadi Admin KAI Jakarta"
  ↓
User accept:
  ├ RegionMembership created (user_id, region_jakarta, role=admin, status=active)
  ├ Notification: "Anda sekarang Admin KAI Jakarta"
  └ Admin dapat mulai manage region
  
User reject:
  ├ RegionInvitation.status = "rejected"
  └ Superadmin bisa invite lagi nanti
```

---

### Alur 2: Admin Region Invite Member via Email

**Flow:**
```
Admin Jakarta buka "Manage Members"
  ↓ [Invite New Member]
Input email: "budi@example.com"
  ↓
RegionInvitation created
status: "pending"
expires_at: 24 jam dari sekarang
  ↓
Email dikirim ke budi@example.com:
"Anda diundang join KAI Jakarta"
[Accept] [Reject] (links dengan token)
  ↓
User click [Accept]:
  ├ Cek: apakah budi@example.com sudah registered?
  │
  ├─ JIKA REGISTERED:
  │   ├ RegionMembership created
  │   │ (user_id, region_jakarta, role=member, status=active)
  │   ├ Notification: "Bergabung ke KAI Jakarta ✓"
  │   └ User langsung jadi member aktif
  │
  └─ JIKA BELUM REGISTERED:
      ├ Redirect ke signup form
      ├ Email pre-filled: budi@example.com
      ├ Saat signup selesai → auto create RegionMembership (active)
      └ First-time welcome message include region info
  
User click [Reject]:
  ├ RegionInvitation.status = "rejected"
  ├ RegionInvitation.rejected_at = now()
  └ User tidak join region, admin bisa invite lagi nanti
  
24 jam passed (no action):
  ├ RegionInvitation.status = "expired"
  └ User tidak bisa accept dari email link lagi
      (tapi admin bisa lihat status expired di backoffice)
```

---

### Alur 3: Member Request Join Region

**Flow:**
```
Member Andi (status: no_region)
  buka "Explore Regions"
  ↓ [Browse All Regions]
Lihat list: Jakarta (1.245 members), Surabaya (892 members), Bandung (456)
  ↓ [Join Jakarta]
  ↓
RegionMembership created:
  user_id: andi_id
  region_id: jakarta_id
  role: "member"
  status: "pending_approval"
  
Notification ke user:
  "Permintaan join KAI Jakarta dikirim. Admin akan review dalam 1-2 hari"
  ↓
Notification ke Admin Jakarta:
  "Andi meminta bergabung ke region ini"
  ↓
Admin Jakarta buka "Pending Requests"
  ├ Lihat: Andi [Approve] [Reject]
  │
  ├─ [Approve]:
  │   ├ RegionMembership.status = "active"
  │   ├ RegionMembership.joined_at = now()
  │   ├ RegionMembership.approved_by = admin_id
  │   ├ Notification ke Andi: "Bergabung ke KAI Jakarta ✓"
  │   └ member_count + 1
  │
  └─ [Reject]:
      ├ RegionMembership.status = "rejected"
      ├ RegionMembership.rejection_reason = "..." (optional)
      ├ RegionMembership.approved_by = admin_id
      ├ Notification ke Andi: "Permintaan ditolak. Alasan: ..."
      └ RegionMembership tidak dihapus, tetap ada di record (audit trail)
```

---

### Alur 4: Member Keluar dari Region

**Flow:**
```
Member Andi (active member KAI Jakarta)
  buka "My Region" → [Settings]
  ↓ [Leave This Region]
  ↓
Confirmation modal:
  "Yakin keluar dari KAI Jakarta?
   Anda bisa request join lagi nanti."
  
[Confirm Leave]:
  ├ RegionMembership status = "inactive" 
     atau soft-delete (tetap di record untuk audit trail)
  ├ member_count - 1
  ├ Notification: "Anda telah keluar dari KAI Jakarta"
  ├ Andi sekarang status "no_region"
  └ Bisa request join region lain atau terima undangan baru
```

---

## Permission Matrix

### Superadmin

| Action | Permission |
|--------|-----------|
| Create region | ✅ |
| Edit region info (name, slug, description, image) | ✅ |
| Deactivate/activate region | ✅ |
| Delete region | ❌ (hanya deactivate) |
| Assign admin to region | ✅ |
| Remove admin from region | ✅ |
| View all regions | ✅ |
| View members in any region | ✅ |
| Invite member to region (as admin or member) | ✅ |
| Approve member join request | ✅ |
| Reject member join request with reason | ✅ |
| Remove member from region | ✅ |
| Monitor region statistics | ✅ |

---

### Admin Region

| Action | Permission | Scope |
|--------|-----------|-------|
| View own region info | ✅ | My region only |
| View members in own region | ✅ | My region only |
| Invite member via email | ✅ | My region only |
| Approve member join request | ✅ | My region only |
| Reject member join request with reason | ✅ | My region only |
| Remove member from region | ✅ | My region only |
| Edit region info | ❌ | |
| Assign/remove admin | ❌ | |
| Create new region | ❌ | |
| Delete/deactivate region | ❌ | |
| View other regions | ✅ | Read-only (untuk info) |

---

### Member

| Action | Permission |
|--------|-----------|
| Browse all regions | ✅ |
| View region info | ✅ |
| Request join region | ✅ |
| Accept invitation | ✅ (if email match) |
| Reject invitation | ✅ |
| View my region info | ✅ (if active member) |
| Leave region | ✅ |
| View region members (list only) | ✅ (if active member, names without email) |
| Manage region | ❌ |

---

## Invitation Lifecycle

```
State Diagram:

Created (pending)
    │
    ├─→ [User Accept] ────────→ Accepted (24h window)
    │                              │
    │                              └─→ [Auto: 24h passed] → Expired
    │
    ├─→ [User Reject] ───────→ Rejected (permanent)
    │
    └─→ [24h passed] ────────→ Expired (automatic)

Duration: Created_at → Expires_at (24 jam)
Expiry: Jika user tidak accept dalam 24 jam, tidak bisa accept lagi dari email link

Admin actions:
- View pending invitations
- Resend invitation (create new RegionInvitation dengan token baru)
- Cancel invitation (set status to "cancelled" atau biarkan expire)
```

---

## Use Cases

### UC1: Superadmin Setup Region Baru dengan Admin

**Actor:** Superadmin (KAI Pusat)  
**Goal:** Buat region Jakarta dan assign Budi sebagai admin

**Flow:**
```
1. Superadmin buka "Manage Regions" di backoffice
2. [Create New Region]
3. Form: Name="KAI Jakarta", Slug="jakarta", Description="...", Image
4. [Create] → Region terbuat, status="active"
5. Buka region Jakarta → [Invite Admin]
6. Search user "Budi Santoso" atau input email "budi@example.com"
7. System check:
   ├ Jika sudah registered → show user card
   └ Jika belum → lanjut dengan email
8. [Invite as Admin]
9. RegionInvitation + email dikirim ke Budi
10. Budi accept dari email → RegionMembership created (role=admin, status=active)
11. Budi sekarang bisa manage KAI Jakarta
```

---

### UC2: Admin Invite Member via Email

**Actor:** Admin Region Jakarta  
**Goal:** Undang 5 orang anggota baru ke region

**Flow:**
```
1. Admin buka "Manage Members" → [Invite]
2. Bulk invite: input 5 email addresses
   - andi@example.com
   - citra@example.com
   - doni@example.com
   - eka@example.com
   - fauzi@example.com
3. [Send Invitations]
4. System create RegionInvitation untuk masing-masing email
5. Email dikirim ke semua 5 orang
6. Admin lihat status: "5 pending invitations"

Follow-up:
- Andi, Citra accept dalam 1 hari → Approved (auto-join)
- Doni, Eka reject → Rejected (stay in region? atau leave?)
- Fauzi tidak action, 24 jam expired → Auto-expire, Fauzi bisa invite ulang
```

---

### UC3: Member Browse & Request Join

**Actor:** Member (Andi, belum ada region)  
**Goal:** Lihat region apa saja dan request join Jakarta

**Flow:**
```
1. Andi buka home screen, lihat "You're not in any region"
2. [Browse Regions]
3. List region: Jakarta (1.245 members), Surabaya (892), Bandung (456)
4. Tap Jakarta → [Request to Join]
5. RegionMembership created: status="pending_approval"
6. Notification ke Andi: "Request sent"
7. Notification ke Admin Jakarta: "1 new join request"

Admin Jakarta review:
8. Buka "Pending Requests" → Lihat Andi
9. [Approve] → status="active", Andi jadi member
10. Notification ke Andi: "Bergabung ke KAI Jakarta ✓"
```

---

### UC4: Member Accept Invitation via Email

**Actor:** Member (Budi, dapat email invite)  
**Goal:** Accept invite ke KAI Jakarta

**Flow:**
```
1. Budi dapat email: "Anda diundang join KAI Jakarta"
2. [Accept] link di email
3. System verify token, check email match
4. Budi already registered? → Yes
5. Auto create RegionMembership (active)
6. Redirect ke Jakarta region page
7. Notification: "Bergabung ke KAI Jakarta ✓"
```

---

### UC5: Member Leave Region & Join Another

**Actor:** Member Andi (currently in Jakarta)  
**Goal:** Keluar Jakarta, join Surabaya

**Flow:**
```
1. Andi buka "My Region: KAI Jakarta"
2. [Leave Region]
3. Confirmation: "Yakin keluar?"
4. [Confirm] → RegionMembership.status = "inactive"
5. Notification: "Keluar dari KAI Jakarta"
6. Andi sekarang status "no_region"
7. [Browse Regions]
8. Tap Surabaya → [Request to Join]
9. New request submit
10. Admin Surabaya approve → Andi join Surabaya
```

---

## Status & Transition

### RegionMembership Status Flow

```
User (no region):
  ↓
[Admin invite] → pending_approval → [User reject] → rejected
                                  → [Admin approve] → active
                                  → [24h expire] → expired

                → [User request] → pending_approval → [Admin reject] → rejected
                                                     → [Admin approve] → active

Active member:
  ├→ [User leave] → inactive
  └→ [Admin remove] → inactive

Inactive:
  └→ User bisa request join lagi
```

---

### RegionInvitation Status Flow

```
pending ─→ [User accept] ──→ accepted (final, success)
        │
        ├─→ [User reject] ──→ rejected (final)
        │
        └─→ [24h expire] ───→ expired (final)
```

---

## Edge Cases & Rules

### Rule 1: One Active Membership Per User

Constraint: Unique index pada (user_id, region_id) where status='active'

```
User Andi:
  ✓ Active member KAI Jakarta
  ✓ Cannot have active membership di region lain secara bersamaan
  ✓ Harus leave Jakarta dulu sebelum join Surabaya
  ✓ Bisa punya "rejected" atau "pending_approval" di multiple regions
     (tapi tidak lebih dari 1 per region dalam status tersebut)
```

---

### Rule 2: Invitation 24-Hour Expiry

```
RegionInvitation.created_at: 2026-05-26 10:00:00
RegionInvitation.expires_at: 2026-05-27 10:00:00

Saat user click accept link setelah 24 jam:
  ├ System check: now() > expires_at?
  ├ Yes → Error: "Invitation expired"
  └ No → Proceed dengan accept

User tetap bisa request join region manual (tidak perlu invite)
```

---

### Rule 3: Admin vs Member Invitation Role

```
Saat superadmin invite via email:
  ├ [Invite as Admin] → role="admin"
  └ [Invite as Member] → role="member"

Saat admin region invite:
  └ [Invite] → role="member" (selalu member, admin cannot assign admin)

Superadmin bisa promote/demote member ↔ admin nanti (via backoffice endpoint)
```

---

### Rule 4: Removed Members Can Request Again

```
Scenario: Admin remove member dari region
  ├ RegionMembership.status = "inactive"
  ├ Member notification: "Anda dihapus dari KAI Jakarta by Admin"
  └ Member bisa:
     ├ Request join lagi (submit new request)
     └ Terima invite lagi (jika diundang)
```

---

*Dokumen ini menjelaskan bisnis logic region system. Untuk detail teknis lihat REGION_API_SPEC.md dan REGION_DB_SCHEMA.md*
