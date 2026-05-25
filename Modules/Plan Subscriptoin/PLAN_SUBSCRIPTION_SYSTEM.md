# KAI Subscription System — Functional Requirements Breakdown (UPDATED)

## Ringkasan

KAI punya 2 plan: **Standard** (Rp 49rb/bulan, default untuk semua member baru) dan **Pro** (Rp 129rb/bulan, optional upgrade). Sementara ini pakai manual verification (transfer + admin approve), siap integrasi payment gateway nanti. Sistem harus track history, manage renewal, dan gate fitur berdasarkan plan.

**Key difference:**
- **Standard:** Akses semua fitur utama, KECUALI posting news, membuat community, dan membuat company/merchants di directory
- **Pro:** Standard + posting news (dengan approval) + buat community + buat company/merchants + buat event

---

## 1. ENTITIES & MASTER DATA

### 1.1 Plan Master

Entitas: `plans`

```
Plan: Standard (DEFAULT)
- Price: Rp 49,000/month
- Duration: 30 days (auto-renew)
- Status: Active
- Default: YES (all new members start here automatically)
- Description: "Akses fitur utama: baca, join komunitas, tanya QnA, lihat direktori"

Plan: Pro
- Price: Rp 129,000/month
- Duration: 30 days (auto-renew)
- Status: Active
- Default: NO (upgrade optional)
- Description: "Standard + posting news (approval), buat community, buat company/merchants, buat event"
```

**Admin dapat:**
- View all plans
- Edit price, description
- Enable/disable plan (pause new registration)
- Add new plan tier (future)
- Configure benefits per plan (fleksibel, tim bisnis yang atur)

---

### 1.2 Plan Benefits Architecture

Benefits punya 3-tier hierarchy:

#### **Tier 1: Benefits Master** (managed by Usergod)
Entitas: `benefits_master`

Benefit keys yang tersedia di sistem (immutable, defined once):

```
Contoh benefits_master (created by usergod):

post_news
- description: "Member can post news"
(Approval flow adalah ROLE-BASED, bukan benefit-based)

create_community
- description: "Member can create and manage community"

create_store
- description: "Member can create store/merchant listing"

create_event
- description: "Member can host event"

view_analytics
- description: "Member can view analytics dashboard"

priority_support
- description: "Priority email/chat support"

[Future benefits ditambah di sini oleh usergod]
```

**Usergod responsibility:**
- Create benefit master (new feature unlock → usergod define benefit key)
- Coordinate dengan developer (key harus ada di code)
- **NOTE**: Approval flow (siapa yang approve) dihandle oleh ROLE/PERMISSION system, bukan benefit system

**IMPORTANT - Benefit vs Approval:**
- **Benefit** = "Apa fitur yang bisa diakses?" (gated by subscription plan)
- **Approval** = "Siapa yang bisa approve aksi?" (gated by role/permission system)

Contoh `post_news`:
- Benefit `post_news` unlock fitur posting untuk pro members
- Permission `post_news` allow member pro to post
- Permission `approve_news` allow admin/superadmin to approve posts
- Member pro posts → saved as draft
- Admin sees draft → approves/rejects via approve_news permission

---

#### **Tier 2: Plan Benefits Assignment** (managed by Superadmin)
Entitas: `plan_benefits`

Superadmin choose dari benefits_master, assign ke plans (enable/disable anytime):

```
Standard plan (default):
- post_news: disabled ❌
- create_community: disabled ❌
- create_store: disabled ❌
- create_event: disabled ❌
- view_analytics: disabled ❌
- priority_support: disabled ❌

Pro plan (optional upgrade):
- post_news: enabled ✅ (dengan approval flow)
- create_community: enabled ✅ (owner/moderator, can manage members)
- create_store: enabled ✅ (merchant profile, upload produk/jasa)
- create_event: enabled ✅ (host event, manage attendees)
- view_analytics: enabled ✅ (community stats, post views, engagement)
- priority_support: enabled ✅
```

**Superadmin responsibility:**
- Choose benefits dari benefits_master (tidak bisa create benefit baru)
- Enable/disable benefits per plan
- Can toggle anytime via admin panel
- No code change needed

**Developer responsibility (Tier 3 - Code Implementation):**
- Implement permission checks: `if (hasUserBenefit(userId, 'post_news'))`
- Show lock message di UI jika benefit disabled
- Reference hardcoded benefit keys (match dengan benefits_master)

---

### 1.3 Why Three-Tier?

✅ **Controlled & Safe:** Usergod define system, no random benefits
✅ **Flexible:** Superadmin can enable/disable without code change
✅ **Easy A/B test:** Different Pro v1 vs Pro v2 benefit combinations
✅ **Future-proof:** New benefits easily added via benefits_master
✅ **Tim bisnis empowered:** Manage benefits without developer dependency

---

### 1.3 Subscription History Tracking

Entitas: `subscription_history`

Setiap perubahan subscription tercatat:

```
{
  id: uuid
  user_id: uuid
  old_plan: 'standard' | 'pro'
  new_plan: 'standard' | 'pro'
  action: 'signup' | 'upgrade' | 'downgrade' | 'renewal' | 'cancel' | 'expired'
  reason: string (optional)
  initiated_by: 'user' | 'admin' | 'system'
  created_at: timestamp
}
```

**Contoh entries:**
- User A: signup → standard (automatic saat register)
- User B: standard → pro (upgrade manual verification)
- User A: standard → standard (renewal auto)
- User C: pro → standard (user downgrade request)
- User D: pro → standard (expired, no renewal)

---

## 2. FLOW PER ROLE

### ROLE 1: MEMBER (User terdaftar dengan Standard plan default)

**Subscription status:** `plan: 'standard'` (automatic saat signup, no payment required for default)

**Apa saja fitur yang bisa diakses?**
- Baca news, event, direktori (read-only)
- Join komunitas (unlimited)
- Post di komunitas (unlimited)
- Tanya di QnA
- Lihat/cari direktori
- Edit profile pribadi
- Lihat profile orang lain

**Fitur yang TIDAK bisa (locked behind Pro paywall):**
- Post news → error "Upgrade to Pro to post news"
- Buat community → error "Upgrade to Pro to create community"
- Buat company/merchants di directory → error "Upgrade to Pro to create store listing"
- Buat event → error "Upgrade to Pro to create event"

---

#### FLOW 1A: Member Request Upgrade to Pro (Manual Verification)

**Trigger:** User click "Upgrade to Pro" di profile/pricing page

**Step 1: Konfirmasi upgrade**
```
Screen: Upgrade to Pro

Current plan: Standard (Rp 49,000/month)
Upgrade to: Pro (Rp 129,000/month)
Additional cost: Rp 80,000

Benefits unlock:
✓ Post news (with approval)
✓ Create community
✓ Create store listing (company/merchants)
✓ Create events
✓ View analytics
✓ Priority support

[Continue] [Cancel]
```

**Step 2: Payment method selection**
```
Screen: Choose Payment Method

[ ] Manual transfer (saat ini)
[ ] Stripe (coming soon)
[ ] Midtrans (coming soon)

[Continue]
```

**Step 3: Manual transfer pathway**
```
Screen: Transfer Details

Bank account target:
- BCA: 1234567890 (PT KAI)
- Mandiri: 0987654321 (PT KAI)
- GCash: GCash ID (KAI Central)

Amount: Rp 80,000
Reference: UPGRADE_USER_[USER_ID]_[TIMESTAMP]

[ ] I've transferred the amount
[Upload proof] → image/video dari transfer (optional tapi recommended)

[Submit Request]
```

**Step 4: Request created, status = PENDING**
```
Entitas: subscription_requests
{
  id: req_xxx
  user_id: user_123
  current_plan: 'standard'
  requested_plan: 'pro'
  status: 'pending'
  payment_provider: 'manual'
  payment_provider_transaction_id: null
  manual_note: 'user uploaded proof image url'
  verified_by_admin: false
  created_at: 2026-05-24T10:00:00Z
  approved_at: null
}
```

**Step 5: Member sees pending state**
```
Screen: Profile → Subscription

Current plan: Standard
Pending upgrade request:
- Plan: Pro
- Amount: Rp 80,000
- Status: ⏳ Awaiting admin verification
- Submitted: 24 May 2026, 10:00

[View proof] [Cancel request]

Note: "Biasanya diverifikasi dalam 1x24 jam"
```

Member TIDAK bisa:
- Request upgrade lagi (while pending)
- Downgrade (hanya cancel request)

---

#### FLOW 1B: Member Upgrade via Payment Gateway (Future)

**Same as 1A Step 1-2, tapi Step 3:**

```
Screen: Stripe Checkout

Amount: Rp 80,000
Plan: Pro
Currency: IDR

[Pay with Stripe]
↓
User complete payment di Stripe
↓
Webhook call: /payment/webhook (success)
↓
subscription_requests status: completed
↓
Automatic activate subscription
```

Member sees: "Payment successful! Your Pro plan is now active."

---

#### FLOW 1C: Member Downgrade / Cancel

**Scenario 1: Downgrade Pro → Standard**
```
Current plan: Pro (active)
Period end: 25 June 2026

User click "Downgrade to Standard" di subscription settings

Confirmation modal:
"Downgrading means you'll lose Pro benefits:
- Can't post news
- Can't create community
- Can't create store listings
- Can't create events

Your downgrade takes effect on 25 June 2026 (expiry date).
Current Pro features will still work until then.

[Confirm downgrade] [Cancel]"

↓ [Confirm downgrade]

subscription.status = 'cancelled'
subscription_history: Pro → Standard, action = 'downgrade'

At 25 June:
- Auto-downgrade to Standard
- User notified: "Your Pro plan expired. You're now on Standard."
```

**Scenario 2: Renewal reminder & cancellation**
```
Current plan: Pro (active)
Period end: 25 June 2026

7 days before expiry (18 June):
Email: "Your Pro plan expires in 7 days"
- Amount to renew: Rp 129,000
- Next period: 25 June - 25 July 2026
- Auto-renewal enabled

User can:
[Continue] → renewal process
[Cancel renewal] → Pro ends on 25 June, downgrade to Standard

3 days before (22 June):
In-app notification: "Your Pro plan expires in 3 days!"

If no action → auto-renews on 25 June (expect admin verification for manual)
If [Cancel renewal] → Pro ends on 25 June
```

---

### ROLE 2: PRO MEMBER (User dengan Pro plan aktif)

**Subscription status:** `plan: 'pro'`, `status: 'active'`

#### FLOW 2A: Pro member post news (dengan approval)

**New capability:** Can post news

```
Screen: Create News

Title: "Event besar KAI tahun ini"
Content: "..."
Category: "Business"
Region: "KAI Pusat"

[Save as draft] [Submit for approval]

↓ [Submit for approval]

Entitas created:
{
  news_id: news_xxx
  user_id: user_pro
  title: "Event besar..."
  status: 'pending_approval'
  posted_by_plan: 'pro'
  created_at: 2026-05-24T10:00:00Z
}

Member sees:
Status: ⏳ Pending approval
"Your news is being reviewed by admin. Usually 24-48 hours."
[View] [Edit draft] [Withdraw]
```

**Admin approves/rejects:**
```
Admin dashboard: Pending approvals
- News from John (Pro) - "Event besar..."
  [Approve] [Reject with reason]

If approved → news.status = 'published'
If rejected → news.status = 'rejected', reason sent to user
```

---

#### FLOW 2B: Pro member create community

**New capability:** Can create & own community

```
Screen: Create Community

Name: "Futsal Jakarta"
Description: "Komunitas futsal di Jakarta"
Category: [dropdown: sports, nature, business, etc.]
Rules: "..."
Image: [upload]

[Create community]

↓

Entitas created:
{
  community_id: comm_xxx
  owner_user_id: user_pro
  name: "Futsal Jakarta"
  status: 'active'
  created_by_plan: 'pro'
  created_at: 2026-05-24T10:00:00Z
}

Member becomes owner automatically.
Can add moderators, manage members, set rules.
```

---

#### FLOW 2C: Pro member create store listing (company/merchants)

**New capability:** Can create company & list products/services

```
Screen: Create Store Listing

Company Name: "PT KAI Electronics"
Description: "Distributor elektronik Korea"
Category: [dropdown]
Address: "Jakarta Pusat"
Phone: "+62812345678"
Email: "info@kai-electronics.com"
Image: [upload]

[Create store]

↓

Entitas created:
{
  company_id: comp_xxx
  owner_user_id: user_pro
  name: "PT KAI Electronics"
  status: 'active'
  created_by_plan: 'pro'
  created_at: 2026-05-24T10:00:00Z
}

After store created, Pro member can:
- Add products/services (name, price, description, image)
- Edit store info
- View store stats
- Manage orders/inquiries
```

---

#### FLOW 2D: Pro member create event

**New capability:** Can host event

```
Screen: Create Event

Event name: "KAI Family Gathering 2026"
Date: "25 June 2026"
Time: "09:00 - 17:00"
Location: "Senayan Jakarta"
Description: "..."
Max attendees: 200
Image: [upload]

[Create event]

↓

Entitas created:
{
  event_id: event_xxx
  user_id: user_pro
  name: "KAI Family Gathering 2026"
  status: 'published'
  created_by_plan: 'pro'
  created_at: 2026-05-24T10:00:00Z
}

Pro member can:
- Manage attendees (approve, reject, see RSVP list)
- Update event details
- Send announcement to attendees
- View event stats
```

---

#### FLOW 2E: Pro member view analytics

**New capability:** View basic analytics

```
Screen: My Analytics

Community: "Futsal Jakarta"
Period: This month

Stats:
- Total members: 450
- New members: 23
- Total posts: 156
- Total comments: 892
- Engagement rate: 12.5%

Top posts:
[Post 1] - 145 likes, 28 comments
[Post 2] - 98 likes, 15 comments
```

---

### ROLE 3: SUPERADMIN (KAI Pusat)

**Subscription management visibility:** All users, all regions

**Fitur:**
- View all subscription requests (manual & gateway)
- View subscription history (all users)
- Verify/reject upgrade requests
- Create refund
- View analytics: conversion, MRR, churn rate
- Manage plan master (add new plan, edit price, enable/disable)
- Manage plan benefits (add/remove features per plan)
- Send bulk renewal reminders
- Configure payment gateway settings

---

#### FLOW 3A: Superadmin verify upgrade request

```
Screen: Subscription Verification Dashboard

Filter: [All] [Pending] [Completed] [Rejected]
Sort: [Newest] [Oldest]

Pending requests (5):
┌──────────────────────────────────────────────┐
│ ID: req_123                                   │
│ User: John Doe (john@example.com)            │
│ Current plan: Standard                        │
│ Requested plan: Pro                           │
│ Amount: Rp 80,000                            │
│ Submitted: 24 May 2026, 10:00                │
│ Proof: [View image/video]                    │
│                                               │
│ [Approve]  [Reject]  [Request more info]    │
└──────────────────────────────────────────────┘
```

**Action: Approve**
```
Modal: Confirm Approval

[Approve] [Cancel]

↓ [Approve]

Backend:
1. Update subscription_requests: status = 'completed', approved_at = now()
2. Find current subscription (user_standard)
3. Create new subscription entry:
   - user_id: john
   - plan: pro
   - status: active
   - current_period_start: 24 May 2026
   - current_period_end: 24 June 2026
4. Mark old subscription as expired (soft delete or status = 'superseded')
5. Create subscription_history: john, standard → pro, action = 'upgrade'
6. Send email to user: "Your Pro plan is now active!"
7. Send in-app notification: "Upgrade successful!"

Admin sees: ✅ Approved

Member sees in app:
- Current plan: Pro (active)
- Benefits unlock immediately
```

**Action: Reject**
```
Modal: Reject Request

Reason: [dropdown]
- Duplicate request
- Invalid proof
- Amount doesn't match
- Other: [text field]

[Reject] [Cancel]

↓ [Reject]

Backend:
1. Update subscription_requests: status = 'rejected', rejection_reason = "..."
2. Send email to user: "Your upgrade request was rejected. Reason: ..."

Member sees:
- Request status: ❌ Rejected
- [Retry upgrade]
```

---

#### FLOW 3B: View subscription history

```
Screen: Subscription History

Filter: 
- User: [search by name/email]
- Action: [All] [Signup] [Upgrade] [Downgrade] [Renewal] [Expired] [Cancelled]
- Date range: [From] [To]
- Plan: [All] [Standard] [Pro]

Results:
────────────────────────────────────────────────────────
Date          User          Old Plan  New Plan  Action
────────────────────────────────────────────────────────
24 May 2026   John Doe      Standard  Pro       Upgrade
23 May 2026   Jane Smith    -         Standard  Signup
22 May 2026   Mike Park     Standard  Standard  Renewal
────────────────────────────────────────────────────────

[Click row for details]
```

---

#### FLOW 3C: Manage plan master & benefits

```
Screen: Plan Management

┌──────────────────────┬──────────┬────────────────────────┐
│ Plan Name            │ Price    │ Status                 │
├──────────────────────┼──────────┼────────────────────────┤
│ Standard (DEFAULT)   │ Rp 49k   │ Active                 │
│ Pro                  │ Rp 129k  │ Active                 │
│ Premium (draft)      │ Rp 299k  │ Disabled               │
└──────────────────────┴──────────┴────────────────────────┘

[Edit Standard] [Edit Pro] [Add New Plan]

────────────────

Click [Edit Pro]:

Form:
- Name: "Pro" [read-only]
- Price: "129,000" → can change
- Duration: "30 days" [dropdown: 7, 14, 30, 90]
- Status: [Active / Disabled]
- Description: "Standard + posting news (approval), buat community..." [editable]

─ BENEFITS (tim bisnis atur ini) ─

✅ Post news (dengan approval)
✅ Buat community
✅ Buat company/merchants
✅ Buat event
✅ View analytics
✅ Priority support

[Add benefit] [Remove] [Save changes] [Cancel]
```

---

#### FLOW 3D: View subscription analytics

```
Screen: Subscription Analytics

Date range: [This month] [Last month] [Custom]

────────────────────────────────────

KPI Cards:
┌─────────────────┐ ┌──────────────┐ ┌─────────────┐
│ Total Users     │ │ Pro Users    │ │ MRR         │
│ 5,234           │ │ 380          │ │ Rp 49.02M   │
└─────────────────┘ └──────────────┘ └─────────────┘

┌─────────────────┐ ┌──────────────┐ ┌─────────────┐
│ Conversion      │ │ Churn Rate   │ │ Avg Revenue │
│ Pro/Total: 7.3% │ │ 8.2%         │ │ Rp 94k      │
└─────────────────┘ └──────────────┘ └─────────────┘

────────────────────────────────────

Breakdown by plan:

Standard (DEFAULT):
- Users: 4,854 (default saat signup)
- Revenue: N/A (default, no payment)
- Note: Semua member baru start di sini

Pro:
- Users: 380
- Revenue: Rp 49.02M (380 × 129k)
- Growth: +5% from last month

────────────────────────────────────

Churn analysis:
- 31 downgrade Pro→Standard this month
- Top reason: "Features not needed" (45%)
- Top reason: "Too expensive" (35%)
```

---

#### FLOW 3E: Configure payment gateway (future)

```
Screen: Payment Gateway Settings

Current payment method: Manual verification

Add payment gateway:
[ ] Stripe
    API Key: [hidden]
    Webhook URL: https://api.kai.app/payment/webhook
    Status: [Configure] [Test] [Enable] [Disable]

[ ] Midtrans
    Merchant ID: [hidden]
    Webhook URL: https://api.kai.app/payment/webhook
    Status: [Configure] [Test] [Enable] [Disable]

────────────────────────────────

Manual verification settings:
- Status: [Enabled] [Disabled]
- Grace period (days): 7 (↑ ↓)
- Auto-cancel pending after X days: 14 (↑ ↓)

[Save settings]
```

---

### ROLE 4: ADMIN REGION (KAI Wilayah)

**Subscription management visibility:** Users dalam region mereka only

**Fitur:**
- View subscription requests dari region (monitoring only)
- View subscription history dari region
- Forward request ke superadmin untuk approval (optional)

---

#### FLOW 4A: Regional admin monitoring

```
Screen: Subscription Requests - My Region (KAI Jakarta)

Filter: [All] [Pending] [Completed] [Rejected]

Pending requests (3):
┌──────────────────────────────────────────┐
│ User: John Doe (Jakarta Pusat)           │
│ Current plan: Standard                   │
│ Requested plan: Pro                      │
│ Amount: Rp 80,000                        │
│ Proof: [View]                            │
│                                          │
│ Status: Awaiting KAI Pusat approval      │
│ [Forward to superadmin] [View details]   │
└──────────────────────────────────────────┘
```

Regional admin cannot approve directly. Can monitor & forward ke pusat.

---

## 3. FITUR UMUM (Cross-role)

### 3.1 Profile & Subscription Widget

Semua member bisa lihat di profile section:

```
┌─────────────────────────────────┐
│ SUBSCRIPTION STATUS             │
├─────────────────────────────────┤
│ Current plan: Standard          │
│ Status: Active ✓                │
│                                 │
│ Period: 24 May - 24 June 2026   │
│ Renews in: 30 days              │
│                                 │
│ [View benefits] [Manage]        │
│ [Upgrade to Pro] [Settings]     │
└─────────────────────────────────┘
```

---

### 3.2 Benefits Comparison Table

Di pricing page:

```
               Standard        Pro
─────────────────────────────────────
Baca konten      ✓              ✓
Join komunitas   ✓              ✓
Post komunitas   ✓              ✓
─────────────────────────────────────
Post news        ❌             ✓ (approval)
Buat community   ❌             ✓
Store listing    ❌             ✓
Buat event       ❌             ✓
Analytics        ❌             ✓
─────────────────────────────────────
Price            Rp 49k/mo      Rp 129k/mo
                 (DEFAULT)      [Upgrade]
─────────────────────────────────────
```

---

### 3.3 Notification System

**Email notifications:**
1. Signup (automatically Standard) → welcome email
2. Upgrade request submitted → confirmation
3. Upgrade approved → account upgraded, benefits unlock
4. Upgrade rejected → reason provided, [retry]
5. 7 days before renewal → reminder
6. 3 days before renewal → final reminder
7. Renewal successful → confirmation
8. Subscription expired/cancelled → notice

**In-app notifications:**
1. Subscription status changed (upgrade/downgrade/renewal)
2. Renewal reminder
3. Subscription expired
4. Feature locked → suggest upgrade (e.g., "Upgrade to Pro to post news")
5. Request status changed (approved/rejected)

---

### 3.4 Feature Lock Messages

When Standard member try access Pro feature:

```
Error modal:

"This feature is only for Pro members"

Description: "Post news, create community, create store, and host events 
are Pro-exclusive features. Upgrade to unlock."

Benefits unlock:
✓ Post news
✓ Create community  
✓ Create store listing
✓ Create events
✓ View analytics

[Upgrade to Pro] [Learn more] [Cancel]
```

---

## 4. SPECIAL CASES & EDGE CASES

### Case 1: Signup flow (automatic Standard)

**Scenario:** User baru register

```
Step 1: Registration form (email, password, name, etc.)
Step 2: Email verification
Step 3: Complete profile (optional)
Step 4: Redirect to home

Automatic:
- subscription entry created: plan = 'standard', status = 'active'
- subscription_history: action = 'signup'
- User can use Standard features immediately (no payment needed)
- Email welcome: "Welcome to KAI! You're on Standard plan."
```

---

### Case 2: Multiple pending requests

**Scenario:** User submit 2x upgrade request (mistake)

```
First request: req_001 (Standard → Pro) - pending
Second request: req_002 (Standard → Pro) - submit attempt

System should:
- Block second request
- Show error: "You already have a pending upgrade request"
- [View existing request]
```

---

### Case 3: Upgrade while renewal pending

**Scenario:** User Standard dengan renewal pending, request upgrade to Pro

```
Current subscription: Standard
Renewal period end: 24 June 2026
Renewal request pending: req_std_renewal

User submit upgrade to Pro: req_pro_upgrade

System should:
- Create new upgrade request (req_pro_upgrade)
- If approved: new Pro subscription start 24 June 2026
- Old Standard renewal cancelled automatically
- Clean transition
```

---

### Case 4: Downgrade during active period

**Scenario:** User Pro ingin downgrade ke Standard

```
Current plan: Pro (active)
Period end: 24 June 2026

User request downgrade.

System:
- Confirm downgrade
- Mark current Pro subscription for cancellation
- Downgrade effective: 24 June 2026 (expiry date)
- During grace period: Pro features still work
- After: Auto-downgrade to Standard
```

---

### Case 5: Admin refund/manual downgrade

**Scenario:** Admin need downgrade user (refund atau error)

```
Admin action: 
- Find user Pro subscription
- [Downgrade to Standard] button
- Reason: [dropdown] "Refund requested", "Duplicate transaction", etc.

System:
- Mark Pro subscription: cancelled, reason = "refund"
- Create Standard subscription: start immediately
- Create subscription_history: Pro → Standard, action = "downgrade", initiated_by = "admin"
- Email user: "Your Pro plan has been downgraded to Standard. Refund: Rp 129,000"
```

---

### Case 6: Grace period & auto-downgrade

**Scenario:** User Pro expire, admin want 7-day grace period

```
Subscription: Pro
Expire date: 24 June 2026
Grace period: +7 days (until 1 July 2026)

During grace period (24 June - 1 July):
- User still have Pro access
- In-app warning: "Your Pro plan expires in X days"

After grace period (1 July):
- Auto-downgrade to Standard
- Lose Pro features
- Email: "Your grace period ended. Downgraded to Standard."
```

---

## 5. ADMIN DASHBOARD SUMMARY

**Superadmin dashboard:**

```
Dashboard Overview:
─────────────────────────────────────
Total Users:        5,234
Standard (default): 4,854 (92.7%)
Pro (paid):         380 (7.3%)
Active Requests:    12 (pending verification)

Revenue (this month): 
- Pro MRR: Rp 49.02M

Conversion:
- Standard → Pro: 7.3%

Quick actions:
[ ] Verify requests (12)
[ ] Manage plans & benefits
[ ] View analytics
[ ] Configure payment gateway

Latest requests:
[ ] John Doe - Standard → Pro - 2h ago
[ ] Jane Smith - Standard → Pro - 5h ago
```

---

## 6. BENEFITS FLEXIBILITY & ARCHITECTURE

### Benefits Three-Tier Model

**Tier 1: Usergod (System Master)**
- Define benefit keys (e.g., 'post_news', 'create_community')
- No UI for this, typically via database or developer console
- Add new benefit when new feature released
- Immutable once created

**Tier 2: Superadmin (Plan Configurator)**
- Choose from available benefits (benefits_master)
- Assign/enable/disable benefits per plan
- Flexible: can toggle anytime via admin panel
- No code change needed

**Tier 3: Developer (Code Implementation)**
- Implement permission checks at code level
- Use hardcoded benefit keys
- Show lock message if benefit disabled
- Feature gates in UI (show/hide based on benefit)

---

### Developer Implementation Pattern

```typescript
// Permission check function
async function hasUserBenefit(userId: string, benefitKey: string) {
  const user = await getUser(userId)
  const plan = user.plan  // 'standard' | 'pro'
  const benefit = await getPlanBenefit(plan, benefitKey)
  return benefit?.enabled === true
}

// Controller example: POST /news/create
router.post('/news/create', async (req, res) => {
  if (!await hasUserBenefit(req.user.id, 'post_news')) {
    return res.status(403).json({
      message: 'Upgrade to Pro to post news'
    })
  }
  // Create news logic...
})

// Frontend example: React component
<button onClick={() => {
  if (!hasUserBenefit(userId, 'create_community')) {
    showUpgradeModal('Upgrade to Pro to create community')
  } else {
    navigateTo('/create-community')
  }
}}>
  Create Community
</button>
```

---

### Workflow: Adding New Feature (Future)

**Example: Tim bisnis ingin "advanced_analytics" untuk Pro**

1. **Usergod:** Create benefit key
   ```
   INSERT benefits_master:
   - key: 'advanced_analytics'
   - description: "Advanced charts, export reports, custom filters"
   - approval_required: false
   ```

2. **Developer:** Implement code
   ```typescript
   if (await hasUserBenefit(userId, 'advanced_analytics')) {
     showAdvancedAnalyticsUI()
   } else {
     showBasicAnalyticsUI()
   }
   ```

3. **Superadmin:** Assign to Pro plan
   ```
   Admin panel → Plans → Pro → Add Benefit
   Select: "advanced_analytics" ✓ Enable
   ```

4. **Done!** Pro members instantly get feature. No redeploy.

---

### Permission Matrix

```
Who can do what?

                        Create Benefit  Assign to Plan  Toggle Enable/Disable
Usergod                 ✓              ✗              ✗
Superadmin              ✗              ✓              ✓
Admin Region            ✗              ✗              ✗
Member                  ✗              ✗              ✗
Developer (code)        -              -              Reference hardcoded keys
```

---

---

## 7. DATA MODEL REFERENCE

### benefits_master (Managed by Usergod)
```
{
  id: uuid
  key: string (unique, immutable)  // 'post_news', 'create_community', etc
  description: string
  approval_required: boolean
  approval_role: string | null     // 'superadmin', 'admin', or null
  metadata: {
    internal_notes?: string
  }
  created_at: timestamp
  created_by: 'usergod'
  updated_at: timestamp
}
```

### plan_benefits (Managed by Superadmin)
```
{
  id: uuid
  plan_id: uuid (FK to plans)
  benefit_key: string (FK to benefits_master.key)
  enabled: boolean
  metadata: {
    description_override?: string  // Optional, if different from master
  }
  created_at: timestamp
  updated_at: timestamp
  updated_by: superadmin_id
}
```

### Permission Check (at Runtime)
```
Query: SELECT * FROM plan_benefits 
       WHERE plan_id = (SELECT plan FROM users WHERE id = ?) 
       AND benefit_key = ?
       AND enabled = true

Result: boolean (has benefit or not)
```

---

## 8. FUTURE ENHANCEMENTS

1. **Tiered pricing:** Add more plan tiers (Premium, Enterprise)
2. **Annual billing:** 20% discount untuk yearly subscription
3. **Free trial:** 14 days Pro free before charging
4. **Usage-based pricing:** Per-event hosting fee, per-store listing
5. **Loyalty rewards:** Referral bonus, long-term discount
6. **Family/group plan:** Share 1 subscription for multiple users
7. **Pause subscription:** Pause instead of cancel (up to 3 months)
8. **Dunning management:** Auto-retry failed payments
9. **Region-specific pricing:** Adjust price per region/country
10. **Affiliate program:** Partner merchants get revenue share

---

## 9. SUCCESS METRICS

**What we're tracking:**

1. **Conversion rate:** Standard → Pro (target: 7-10%)
2. **MRR:** Monthly Recurring Revenue Pro tier (target: Rp 50M+)
3. **Churn rate:** Cancelled Pro subscriptions (target: <10% monthly)
4. **ARPU:** Average Revenue Per User (target: Rp 10k+ monthly)
5. **Time to upgrade:** Days from signup to first Pro upgrade (target: <60 days)
6. **Payment success rate:** Successful vs failed transactions (target: >95%)
7. **Renewal rate:** Users who renew after first month (target: >85%)
8. **Feature adoption:** % Pro users utilizing each feature (e.g., 60% create news, 45% create community)

---

## INTEGRATION WITH ROLE & PERMISSION SYSTEM

**Subscription Benefits** vs **Role Permissions** adalah TWO SEPARATE systems yang bekerja bersama:

### Benefit (Subscription-driven)
- **Defines**: Apa fitur yang bisa diakses berdasarkan subscription plan
- **Examples**: `post_news`, `create_store`, `create_community`, `view_analytics`
- **Managed by**: superadmin (assign benefits to plans)
- **Effect**: User dengan plan pro unlock fitur tertentu

### Permission (Role-driven) 
- **Defines**: Siapa yang bisa lakukan aksi berdasarkan role
- **Examples**: `approve_news` (admin/superadmin only), `manage_members` (leader only)
- **Managed by**: usergod/superadmin (assign permissions to roles)
- **Effect**: User dengan role admin/superadmin dapat approve aksi

### Combined Check Example

**User wants to post news:**
1. Check benefit: Does user's plan include `post_news` benefit? (subscription check)
2. If yes → allow to post
   - If admin role → auto-publish
   - If member role → save as draft (awaiting approval via approve_news permission)
3. If no → deny with "upgrade required" message

**Admin approves news:**
1. Check permission: Does user have `approve_news` permission? (role check)
2. If yes → allow to approve draft posts
3. If no → deny

**Reference**: See `ROLE_PERMISSION_SYSTEM.md` → "Permission Check Logic with Subscription Integration" for detailed logic.

---

END OF DOCUMENT
