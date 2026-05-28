# Announcement Module — Rules & Use Cases

Dokumentasi sistem announcement KAI App. Announcement adalah notifikasi broadcast dari admin untuk semua user di platform atau region tertentu.

---

## 1. WHAT IS ANNOUNCEMENT?

**Announcement** adalah sistem komunikasi one-way dari KAI Pusat atau Admin Regional kepada semua user di scope mereka. **Bukan untuk individual user**, tapi untuk broadcast informasi penting ke banyak orang sekaligus.

### Karakteristik:
- **Creator:** Superadmin (global) atau Admin Regional (regional)
- **Target:** All users dalam scope (global atau regional)
- **Purpose:** Broadcast informasi urgent atau penting
- **Approval:** Tidak perlu — langsung publish
- **Notification:** Push + In-app banner + Email

### Berbeda dari:
- **News:** Editorial content, perlu approval
- **Event:** Event registration, user-initiated
- **Community:** User-created groups
- **Direct Message:** 1-to-1 communication

---

## 2. ANNOUNCEMENT TYPES

### Type: DISASTER

**Untuk situasi darurat alam atau keadaan krisis.**

#### Subtypes:
- **earthquake** — Gempa bumi
- **flood** — Banjir
- **landslide** — Longsor
- **tsunami** — Tsunami
- **volcanic** — Erupsi vulkanik
- **storm** — Badai/puyuh
- **fire** — Kebakaran

#### Use Cases:
```
"Gempa 7.2 SR terjadi di Jawa Barat. 
 Epicenter: Bandung. Kedalaman: 12km.
 Pusat evakuasi: GOR Bandung, Capacity 500.
 Helpline: 119"

"Banjir besar di Jakarta Timur mencapai 150cm.
 Rute evakuasi tersedia: Jl. Ahmad Yani → Higher Ground.
 Shelter: Jakarta Timur Convention Hall, 1000 tempat."

"Kebakaran hutan Bogor berkembang pesat.
 Evakuasi mandatory untuk area radius 10km.
 Hotline: 112"
```

#### Who can create:
- Superadmin → Global disaster announcement (affects all regions)
- Admin Regional → Regional disaster announcement (own region only)

#### Priority:
- **CRITICAL** — User dapat push bahkan jika notification off
- **HIGH** — Push immediately

#### Action Required:
- Biasanya informasional, tidak ada action khusus
- Tapi bisa kasih evacuation center locations, helplines, safe routes

---

### Type: SYSTEM

**Untuk system maintenance, outage, atau technical updates yang affect service availability.**

#### Subtypes:
- **maintenance** — Scheduled maintenance
- **outage** — Service down (unplanned)
- **degraded_service** — Service degraded (slower)
- **update** — System update atau patch
- **incident** — Security incident atau critical issue

#### Use Cases:
```
"Scheduled Maintenance: Database upgrade.
 Affected: All regions.
 Duration: Tue 10pm-midnight (4 hours).
 Impact: All services may be unavailable.
 Status page: https://status.kai.app"

"Service Degradation Alert:
 Affected regions: Jakarta, Surabaya.
 Services: Community posting (slow).
 Expected resolution: 2 hours.
 We're working on it!"

"Security Update Required:
 Update your KAI app to latest version.
 Reason: Security vulnerability patched.
 All users should update immediately."
```

#### Who can create:
- Superadmin → Global system announcement
- Admin Regional → Regional-only maintenance (e.g., Jakarta server maintenance)

#### Priority:
- **CRITICAL** → If outage affects payment or security
- **HIGH** → If affects core features
- **MEDIUM** → If affects minor features

---

### Type: URGENT

**Untuk action-required notifications yang affects all users atau scope.**

⚠️ **IMPORTANT:** Bukan untuk individual user yang specific (e.g., "your payment failed").
Ini untuk **all users** yang perlu action. Contoh: "Everyone update password due to security breach".

#### Subtypes:
- **security_alert** — Security incident affecting all users
- **account_action** — Action required dari semua user atau kategori
- **policy_change** — Policy change immediate effect
- **urgent_update** — Update urgent yang harus dilakukan semua user

#### Use Cases:
```
"SECURITY ALERT: Password Reset Required.
 All users must reset password due to security breach.
 Process: Settings → Security → Change Password.
 Deadline: Tomorrow 11:59 PM.
 Support: support@kai.app"

"URGENT: Email Verification Required.
 Email verification mandatory for account security.
 Verify within 24 hours.
 Link: [verify-email]"

"URGENT: New Community Guidelines.
 Effective immediately.
 Violation consequences updated.
 Review: [guidelines-link]"
```

#### Who can create:
- Superadmin only → Global urgent announcements

#### Priority:
- **CRITICAL** → Security breach, immediate action needed
- **HIGH** → Policy enforcement, deadline-driven

---

### Type: INFO

**Untuk informasi umum, feature announcements, updates yang tidak urgent.**

#### Subtypes:
- **feature_launch** — Feature baru release
- **event** — Event announcement (e.g., KAI Summit)
- **promo** — Promo atau special offer
- **policy_update** — Policy update (for info only, tidak immediate)
- **company_news** — KAI company news atau milestone
- **general_info** — General informasi ke semua user

#### Use Cases:
```
"🎉 New Feature: Advanced Analytics!
 Pro members can now view detailed community metrics.
 Available now in Community Settings.
 Learn more: [feature-link]"

"📢 KAI Summit 2026 Registration Open!
 Date: June 15, 2026.
 Venue: Jakarta Convention Center (Hybrid).
 Register: [event-link]"

"🎁 Summer Sale - 50% Off Pro!
 Limited time: 5 days only.
 Upgrade now: [upgrade-link]"

"📋 Privacy Policy Updated.
 Effective June 1, 2026.
 Key changes: [policy-link]"
```

#### Who can create:
- Superadmin → Global info announcements
- Admin Regional → Regional info (e.g., "Jakarta meetup this weekend")

#### Priority:
- **LOW/MEDIUM** → Usually LOW karena tidak urgent
- No push notification (in-app only) atau delayed push

---

## 3. SCOPE: GLOBAL vs REGIONAL

### GLOBAL Scope

```
Announcement:
  scope: 'global'
  region_id: null

Recipients:
  - All users di platform
  - Semua region
  - Tidak peduli plan (standard/pro)

Contoh:
  - KAI Summit 2026 announcement
  - Platform-wide security alert
  - System maintenance (all servers)
  - New feature launch
  - Major policy change
```

### REGIONAL Scope

```
Announcement:
  scope: 'regional'
  region_id: 'region_jakarta'

Recipients:
  - Users dengan region = Jakarta saja
  - User di region lain tidak lihat
  
Contoh:
  - Banjir besar Jakarta Timur
  - Maintenance server Jakarta only
  - Jakarta regional event
  - Local policy atau rules Jakarta
```

---

## 4. WHO CAN CREATE & VIEW

### Superadmin (KAI Pusat)

| Action | Scope | Permission |
|--------|-------|------------|
| Create announcement | Global | ✅ Yes |
| Create announcement | Regional (any region) | ✅ Yes |
| Edit own announcement | Any | ✅ Yes |
| Delete own announcement | Any | ✅ Yes |
| View all announcements | Any | ✅ Yes |
| Publish announcement | Any | ✅ Yes (no approval needed) |

### Admin Regional (e.g., Admin Jakarta)

| Action | Scope | Permission |
|--------|-------|------------|
| Create announcement | Global | ❌ NO |
| Create announcement | Own region (Jakarta) | ✅ Yes |
| Create announcement | Other region (Surabaya) | ❌ NO |
| Edit own announcement | Own region | ✅ Yes |
| Delete own announcement | Own region | ✅ Yes |
| View announcements | Own region | ✅ Yes |
| View announcements | Other region | ❌ NO |
| View announcements | Global | ✅ Yes (but can't edit/delete) |
| Publish announcement | Own region | ✅ Yes |

### Member (All users)

| Action | Permission |
|--------|------------|
| View announcements | ✅ Yes (global + own region) |
| Create announcement | ❌ NO |
| Edit announcement | ❌ NO |
| Delete announcement | ❌ NO |

---

## 5. ANNOUNCEMENT LIFECYCLE

```
┌──────────┐
│  DRAFT   │  Admin create, belum tayang
└────┬─────┘
     │ publish
     ↓
┌──────────────┐
│ PUBLISHED    │  Live, user bisa lihat & dapat notif
└────┬─────────┘
     │
     │ (manual archive OR auto-expire)
     ↓
┌──────────────┐
│  ARCHIVED    │  Tidak muncul di listing utama
│  (EXPIRED)   │  Masih bisa dilihat di history
└──────────────┘
```

### Status Detail:

**DRAFT**
- Admin create tapi belum publish
- Hanya creator yang bisa lihat
- Bisa diedit atau dihapus kapan saja
- Tidak ada notif terkirim

**PUBLISHED**
- Langsung live ke semua user di scope
- Push notification dikirim immediately
- Tidak bisa dihapus (hanya archive)
- Bisa diedit (e.g., extend deadline jika urgent)

**ARCHIVED**
- Dihapus dari listing utama
- Bisa dari manual archive atau auto-expire via `expires_at`
- Masih bisa dilihat di halaman "Announcement History"
- User bisa filter by type/date

---

## 6. NOTIFICATION BEHAVIOR

### Push Notification

Sent immediately when announcement published.

| Priority | Behavior |
|----------|----------|
| CRITICAL | Bypass notification settings, force push to all devices |
| HIGH | Push immediately, respect user notification settings |
| MEDIUM | Push dalam batch (bisa delayed beberapa menit) |
| LOW | No push notification, in-app only |

### In-App Banner

Muncul di home screen mobile app.

| Priority | Display |
|----------|---------|
| CRITICAL | Sticky banner (always on top), highlight warna merah |
| HIGH | Pin di atas listing, highlight warna orange |
| MEDIUM | Normal position di listing |
| LOW | Bawah listing, tidak highlight |

### Email Notification

| Type | Email Sent? |
|------|------------|
| DISASTER | ✅ Yes (to all recipients) |
| SYSTEM (HIGH) | ✅ Yes (if outage/critical) |
| SYSTEM (MEDIUM) | ❌ No |
| URGENT (CRITICAL) | ✅ Yes |
| INFO | ❌ No (unless user enable preference) |

---

## 7. ANNOUNCEMENT TARGETING

### Calculation:

Saat publish, system hitung total recipients berdasarkan:

**Global announcement:**
```
Recipients = All users di platform
           = Total standard members + total pro members
           = Biasanya ~5.000-10.000 users
```

**Regional announcement:**
```
Recipients = Users dengan region = [region_id]
           = Biasanya ~500-2.000 users per region
```

### Sending:

```
1. Query recipients dari database
2. Get FCM tokens untuk setiap user
3. Send push notification (batch)
4. Log semua yang terkirim/failed
5. Update announcement.total_sent

Proses:
- Langsung (1-5 menit setelah publish)
- Via FCM queue (Google Firebase Cloud Messaging)
- Max rate: ~1000 notifications per second
```

---

## 8. ANNOUNCEMENT EXPIRY & ARCHIVE

### Auto-Expire

```
announcement:
  published_at: '2026-05-25T10:00:00Z'
  expires_at: '2026-06-01T00:00:00Z'

Behavior:
- Pada 2026-06-01 00:00:00, status otomatis → 'expired'
- Announcement hilang dari listing
- Masih di history dengan filter
- User yg belum baca tetap bisa baca via notification/history
```

### Manual Archive

```
Admin bisa klik "Archive" untuk announcements
Contoh: Disaster gempa sudah berakhir, archive dari listing
        Maintenance sudah selesai, archive dari listing
```

### Difference:

| Expired | Archived |
|---------|----------|
| Auto, based on date | Manual, by admin |
| `expires_at` field | No date needed |
| Informational | Actively managed |

---

## 9. USE CASE FLOWCHART

### Use Case 1: Disaster Gempa (Global)

```
Gempa terjadi di Jawa Barat

1. Superadmin buka backoffice
2. Click "Create Announcement"
3. Fill:
   - Type: DISASTER
   - Subtype: earthquake
   - Priority: CRITICAL
   - Scope: GLOBAL
   - Location: Map pick epicenter + map pick affected areas
   - Content: Title, body, evacuation info

4. Set metadata:
   - magnitude: 7.2
   - depth_km: 12
   - affected_areas: [Bandung, Cirebon]
   - evacuation_centers: [GOR Bandung]

5. Click "Publish Now"

6. System:
   - status → published
   - Query all 5.234 users
   - Send push notification (critical, bypass settings)
   - Send email to all
   - In-app banner appears (sticky, red)

7. Users receive:
   - Push notification immediately
   - In-app banner (tap to see detail)
   - Email alert
   - Can view map with epicenter + affected areas + evacuation centers
```

### Use Case 2: System Maintenance (Regional)

```
Server Jakarta perlu maintenance

1. Admin Jakarta buka backoffice
2. Click "Create Announcement"
3. Fill:
   - Type: SYSTEM
   - Subtype: maintenance
   - Priority: HIGH
   - Scope: REGIONAL
   - Region: Jakarta

4. Set metadata:
   - maintenance_start: 2026-05-26T22:00:00
   - maintenance_end: 2026-05-27T02:00:00
   - affected_services: [Database, API]
   - status_page_url: https://status.kai.app

5. Click "Publish Now"

6. System:
   - Only 856 users di Jakarta notified
   - Users di Surabaya tidak terima
   - Push notification (high priority)
   - In-app banner

7. Users di Jakarta see:
   - Maintenance window details
   - Affected services
   - Expected duration
   - Status page link
```

### Use Case 3: Feature Launch (Global)

```
Advanced Analytics feature released

1. Superadmin create announcement
2. Fill:
   - Type: INFO
   - Subtype: feature_launch
   - Priority: MEDIUM
   - Scope: GLOBAL

3. Set metadata:
   - feature_name: "Advanced Analytics"
   - availability: "pro_only"
   - features_list: [real-time metrics, custom reports, export]
   - learn_more_link: https://kai.app/features/analytics

4. Publish

5. System:
   - No push (priority MEDIUM)
   - In-app banner (normal position)
   - Email only if user enable feature announcements

6. Users see:
   - Feature description with images
   - Only available for Pro members
   - Button: "Learn More" atau "Upgrade"
```

---

## 10. RULES & CONSTRAINTS

### Announcement Content:

| Field | Rules |
|-------|-------|
| title | Max 150 chars, required |
| body | Max 2000 chars, required |
| image_url | Optional, max 1 image |
| type | Required, must be valid type |
| priority | Required, CRITICAL overrides user settings |
| scope | Required, global or regional |
| region_id | Required if scope=regional, null if global |
| expires_at | Optional, must be future date |

### Timing:

| Constraint | Rule |
|-----------|------|
| Publish immediately | ✅ Supported |
| Schedule for later | ❌ Not supported (publish now or draft) |
| Auto-renew | ❌ Not supported |
| Multi-region announce | ❌ Not supported (create separate per region) |

### Frequency:

| Rule | Detail |
|------|--------|
| How often can admin create | No limit |
| How many can be active | No limit |
| User can opt-out | ✅ Only non-critical |
| User can disable notifications | ✅ Except CRITICAL |

---

## 11. ADMIN RESPONSIBILITY

**Superadmin:**
- Create global announcements only when truly platform-wide
- Review metadata sebelum publish (especially disaster info)
- Set appropriate priority dan deadline
- Monitor delivery rate (total_sent vs total_recipients)
- Archive when announcement no longer relevant

**Admin Regional:**
- Create regional announcements only for their region
- Validate disaster information sebelum publish
- Keep emergency contact info updated
- Coordinate with superadmin jika disaster affects multiple regions
- Set appropriate expiry date

---

## 12. USER EXPERIENCE

### Mobile App:

1. **Notification receive**
   - Push notification appears
   - Badge counter on bell icon
   - Sound/vibration based on priority

2. **In-app banner**
   - Home screen shows sticky banner (if critical/high)
   - User tap to view detail
   - Swipe to dismiss

3. **Announcement list**
   - Dedicated "Announcements" tab
   - Filter by type (disaster, system, urgent, info)
   - Show unread count
   - Oldest first (for disaster/urgent)

4. **Read tracking**
   - Auto-marked as read when user open detail
   - Count of readers shown in stats

5. **History**
   - Can view past announcements
   - Searchable by keyword

### Web Backoffice:

1. **Create form**
   - Dynamic fields based on type
   - Map picker for location
   - Preview before publish

2. **Management**
   - List all announcements with filters
   - Publish, archive, delete buttons
   - View delivery stats (sent/read rate)

3. **Analytics**
   - Total recipients vs sent
   - Read rate per announcement
   - Breakdown by region (if applicable)

---

## 13. EDGE CASES

### Multiple Disasters in Different Regions?

```
Gempa di Jawa Barat + Banjir di Jakarta

Solution: Create 2 announcements
- Announcement 1: Gempa (scope=global, or regional Jawa Barat)
- Announcement 2: Banjir (scope=regional Jakarta)

Each reaches appropriate audience
```

### Announcement during System Outage?

```
Database down, cannot send announcement!

Solution:
- Pre-populate emergency message cache
- SMS fallback untuk CRITICAL announcements
- Or: Manual comms via social media
```

### User Location Changed?

```
User di Jakarta, pindah ke Surabaya

Solution:
- Announcement based on current region_id
- Jika switch region di app, should update region
- Previous regional announcements still visible in history
```

---

*Dokumen ini menjelaskan business rules dan use cases announcement module KAI App.*
