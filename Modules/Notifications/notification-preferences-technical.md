# Notification Preferences — Technical Implementation

Dokumentasi teknis untuk implementasi sistem notification preferences. Mencakup database schema, backend Golang, dan scalability considerations.

---

## Daftar Isi

- [Database Schema](#database-schema)
  - [Approach 1: Normalized (Recommended)](#approach-1-normalized-recommended)
  - [Approach 2: Single JSON](#approach-2-single-json-simpler)
- [Backend Implementation (Golang)](#backend-implementation-golang)
  - [Domain & Repository Layer](#domain--repository-layer)
  - [Service Layer](#service-layer)
  - [Notification Engine Integration](#notification-engine-integration)
- [Scalability Considerations](#scalability-considerations)
- [Best Practices](#best-practices)

---

## Database Schema

### Approach 1: Normalized (Recommended)

Lebih baik untuk long-term scalability. Bisa menambah preference type baru tanpa breaking schema.

#### Table: `notification_preferences`

Primary table menyimpan global settings dan module-level toggles per user.

```sql
CREATE TABLE notification_preferences (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE,

  -- Global toggles (level 1)
  all_notifications_enabled BOOLEAN DEFAULT true,
  do_not_disturb_enabled BOOLEAN DEFAULT false,
  do_not_disturb_start_time TIME,
  do_not_disturb_end_time TIME,

  -- Module-level toggles (level 2)
  news_enabled BOOLEAN DEFAULT true,
  community_enabled BOOLEAN DEFAULT true,
  event_enabled BOOLEAN DEFAULT true,
  qna_enabled BOOLEAN DEFAULT true,
  directory_enabled BOOLEAN DEFAULT false,

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

#### Table: `notification_preferences_news`

Granular settings untuk modul News.

```sql
CREATE TABLE notification_preferences_news (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  preference_id UUID NOT NULL,

  from_kai_pusat BOOLEAN DEFAULT true,
  from_kai_region BOOLEAN DEFAULT true,
  from_pro_members BOOLEAN DEFAULT false,

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  FOREIGN KEY (preference_id) REFERENCES notification_preferences(id) ON DELETE CASCADE,
  UNIQUE(preference_id)
);
```

#### Table: `notification_preferences_community`

Settings per komunitas (dynamic, bisa banyak row per user).

```sql
CREATE TABLE notification_preferences_community (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  preference_id UUID NOT NULL,
  community_id UUID NOT NULL,

  enabled BOOLEAN DEFAULT true,
  new_posts BOOLEAN DEFAULT true,
  member_joined BOOLEAN DEFAULT false,
  member_left BOOLEAN DEFAULT false,

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  FOREIGN KEY (preference_id) REFERENCES notification_preferences(id) ON DELETE CASCADE,
  FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
  UNIQUE(preference_id, community_id)
);
```

#### Table: `notification_preferences_event`

Settings untuk modul Event.

```sql
CREATE TABLE notification_preferences_event (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  preference_id UUID NOT NULL,

  reminders_enabled BOOLEAN DEFAULT true,
  reminder_hours_before INT DEFAULT 24,
  interested_categories TEXT[],

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  FOREIGN KEY (preference_id) REFERENCES notification_preferences(id) ON DELETE CASCADE,
  UNIQUE(preference_id)
);
```

#### Table: `notification_preferences_qna`

Settings untuk modul Q&A.

```sql
CREATE TABLE notification_preferences_qna (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  preference_id UUID NOT NULL,

  someone_replied BOOLEAN DEFAULT true,
  reply_from_moderator BOOLEAN DEFAULT true,

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  FOREIGN KEY (preference_id) REFERENCES notification_preferences(id) ON DELETE CASCADE,
  UNIQUE(preference_id)
);
```

> **Index yang direkomendasikan:**
> ```sql
> CREATE INDEX idx_notif_pref_user_id ON notification_preferences(user_id);
> CREATE INDEX idx_notif_pref_community ON notification_preferences_community(preference_id, community_id);
> ```

---

### Approach 2: Single JSON (Simpler)

Cocok jika app masih baru dan belum butuh scalability tinggi. Read lebih cepat karena single query.

```sql
CREATE TABLE notification_preferences (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE,

  settings JSONB NOT NULL DEFAULT '{
    "global": {
      "all_enabled": true,
      "dnd_enabled": false,
      "dnd_start": null,
      "dnd_end": null
    },
    "modules": {
      "news": {
        "enabled": true,
        "from_kai_pusat": true,
        "from_kai_region": true,
        "from_pro_members": false
      },
      "community": { "enabled": true },
      "event": {
        "enabled": true,
        "reminders_enabled": true,
        "reminder_hours_before": 24
      },
      "qna": { "enabled": true },
      "communities": {
        "comm_123": { "enabled": true, "new_posts": true }
      }
    }
  }',

  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),

  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**Kapan pakai mana:**

| Kriteria | Normalized | Single JSON |
|---|---|---|
| Long-term scalability | ✅ Lebih baik | ⚠️ Sulit extend |
| Query complexity | ⚠️ JOIN banyak | ✅ Single query |
| Add preference baru | ✅ Tambah kolom/tabel | ⚠️ Ubah default JSON |
| Reporting/analytics | ✅ Mudah di-query | ⚠️ Butuh JSON extraction |
| Early-stage app | ⚠️ Over-engineered | ✅ Cukup |

---

## Backend Implementation (Golang)

### Domain & Repository Layer

**Lokasi:** `internal/domain/notification.go`, `internal/repository/postgres/notification_repo.go`

```go
// internal/domain/notification.go

type NotificationPreference struct {
  ID                      string         `json:"id"`
  UserID                  string         `json:"user_id"`
  AllNotificationsEnabled bool           `json:"all_notifications_enabled"`
  DoNotDisturbEnabled     bool           `json:"do_not_disturb_enabled"`
  DoNotDisturbStartTime   sql.NullString `json:"do_not_disturb_start_time"`
  DoNotDisturbEndTime     sql.NullString `json:"do_not_disturb_end_time"`

  News      NewsPreference      `json:"news"`
  Community CommunityPreference `json:"community"`
  Event     EventPreference     `json:"event"`
  QnA       QnAPreference       `json:"qna"`

  CreatedAt string `json:"created_at"`
  UpdatedAt string `json:"updated_at"`
}

type NewsPreference struct {
  Enabled        bool `json:"enabled"`
  FromKaiPusat   bool `json:"from_kai_pusat"`
  FromKaiRegion  bool `json:"from_kai_region"`
  FromProMembers bool `json:"from_pro_members"`
}

type CommunityPreference struct {
  Enabled     bool                        `json:"enabled"`
  Communities map[string]CommunitySpecific `json:"communities"`
}

type CommunitySpecific struct {
  Enabled      bool `json:"enabled"`
  NewPosts     bool `json:"new_posts"`
  MemberJoined bool `json:"member_joined"`
  MemberLeft   bool `json:"member_left"`
}

type EventPreference struct {
  Enabled             bool     `json:"enabled"`
  RemindersEnabled    bool     `json:"reminders_enabled"`
  ReminderHoursBefore int      `json:"reminder_hours_before"`
  InterestedCategories []string `json:"interested_categories"`
}

type QnAPreference struct {
  Enabled            bool `json:"enabled"`
  SomeoneReplied     bool `json:"someone_replied"`
  ReplyFromModerator bool `json:"reply_from_moderator"`
}

// Repository interface (contract)
type NotificationPreferenceRepository interface {
  GetByUserID(ctx context.Context, userID string) (*NotificationPreference, error)
  Create(ctx context.Context, preference *NotificationPreference) error
  Update(ctx context.Context, preference *NotificationPreference) error
  DeleteByUserID(ctx context.Context, userID string) error
}
```

```go
// internal/repository/postgres/notification_repo.go

type notificationPostgresRepo struct {
  db *sqlx.DB
}

func (r *notificationPostgresRepo) GetByUserID(ctx context.Context, userID string) (*domain.NotificationPreference, error) {
  query := `
    SELECT
      np.id, np.user_id,
      np.all_notifications_enabled,
      np.do_not_disturb_enabled,
      np.do_not_disturb_start_time,
      np.do_not_disturb_end_time,
      np.news_enabled, np.community_enabled, np.event_enabled, np.qna_enabled,
      npn.from_kai_pusat, npn.from_kai_region, npn.from_pro_members,
      npe.reminders_enabled, npe.reminder_hours_before, npe.interested_categories,
      npq.someone_replied, npq.reply_from_moderator,
      np.created_at, np.updated_at
    FROM notification_preferences np
    LEFT JOIN notification_preferences_news npn ON np.id = npn.preference_id
    LEFT JOIN notification_preferences_event npe ON np.id = npe.preference_id
    LEFT JOIN notification_preferences_qna npq ON np.id = npq.preference_id
    WHERE np.user_id = $1
  `

  var pref domain.NotificationPreference
  err := r.db.GetContext(ctx, &pref, query, userID)
  if errors.Is(err, sql.ErrNoRows) {
    return nil, domain.ErrNotFound
  }
  if err != nil {
    return nil, err
  }

  // Load community-specific preferences separately
  pref.Community.Communities, _ = r.getCommunityPreferences(ctx, pref.ID)

  return &pref, nil
}

func (r *notificationPostgresRepo) Update(ctx context.Context, pref *domain.NotificationPreference) error {
  tx, err := r.db.BeginTxx(ctx, nil)
  if err != nil {
    return err
  }
  defer tx.Rollback()

  // Update main table
  _, err = tx.ExecContext(ctx, `
    UPDATE notification_preferences
    SET all_notifications_enabled=$1, do_not_disturb_enabled=$2,
        do_not_disturb_start_time=$3, do_not_disturb_end_time=$4,
        news_enabled=$5, community_enabled=$6, event_enabled=$7, qna_enabled=$8,
        updated_at=NOW()
    WHERE id=$9`,
    pref.AllNotificationsEnabled, pref.DoNotDisturbEnabled,
    pref.DoNotDisturbStartTime, pref.DoNotDisturbEndTime,
    pref.News.Enabled, pref.Community.Enabled, pref.Event.Enabled, pref.QnA.Enabled,
    pref.ID,
  )
  if err != nil {
    return err
  }

  // Upsert module-specific tables
  tx.ExecContext(ctx, `
    UPDATE notification_preferences_news
    SET from_kai_pusat=$1, from_kai_region=$2, from_pro_members=$3
    WHERE preference_id=$4`,
    pref.News.FromKaiPusat, pref.News.FromKaiRegion, pref.News.FromProMembers, pref.ID,
  )

  tx.ExecContext(ctx, `
    UPDATE notification_preferences_event
    SET reminders_enabled=$1, reminder_hours_before=$2, interested_categories=$3
    WHERE preference_id=$4`,
    pref.Event.RemindersEnabled, pref.Event.ReminderHoursBefore,
    pq.Array(pref.Event.InterestedCategories), pref.ID,
  )

  return tx.Commit()
}
```

---

### Service Layer

**Lokasi:** `internal/service/notification_service.go`

```go
type notificationService struct {
  repo  domain.NotificationPreferenceRepository
  cache CacheService
}

// Core method: dipanggil sebelum setiap notifikasi dikirim
func (s *notificationService) ShouldSendNotification(ctx context.Context, userID string, notif *NotificationEvent) (bool, error) {
  pref, err := s.repo.GetByUserID(ctx, userID)
  if err != nil {
    if errors.Is(err, domain.ErrNotFound) {
      // Auto-create with defaults for new users
      pref = s.createDefaultPreferences(userID)
      s.repo.Create(ctx, pref)
    } else {
      return false, err
    }
  }

  // Level 1: Global toggle
  if !pref.AllNotificationsEnabled {
    return false, nil
  }

  // Level 2: Do Not Disturb
  if pref.DoNotDisturbEnabled && s.isInDoNotDisturbWindow(pref) {
    return false, nil
  }

  // Level 3: Module-specific
  switch notif.Type {
  case "news_posted":
    if !pref.News.Enabled {
      return false, nil
    }
    switch notif.Source {
    case "kai_pusat":
      return pref.News.FromKaiPusat, nil
    case "kai_region":
      return pref.News.FromKaiRegion, nil
    case "pro_member":
      return pref.News.FromProMembers, nil
    }

  case "community_post":
    if !pref.Community.Enabled {
      return false, nil
    }
    if spec, ok := pref.Community.Communities[notif.CommunityID]; ok {
      return spec.Enabled && spec.NewPosts, nil
    }
    return true, nil // Default: enabled if not yet configured

  case "event_reminder":
    if !pref.Event.Enabled || !pref.Event.RemindersEnabled {
      return false, nil
    }
    if len(pref.Event.InterestedCategories) > 0 {
      return contains(pref.Event.InterestedCategories, notif.EventCategory), nil
    }
    return true, nil

  case "qna_reply":
    if !pref.QnA.Enabled {
      return false, nil
    }
    if notif.ReplyFromModerator {
      return pref.QnA.ReplyFromModerator, nil
    }
    return pref.QnA.SomeoneReplied, nil
  }

  return true, nil
}

func (s *notificationService) createDefaultPreferences(userID string) *domain.NotificationPreference {
  return &domain.NotificationPreference{
    UserID:                  userID,
    AllNotificationsEnabled: true,
    DoNotDisturbEnabled:     false,
    News: domain.NewsPreference{
      Enabled: true, FromKaiPusat: true, FromKaiRegion: true, FromProMembers: false,
    },
    Community: domain.CommunityPreference{
      Enabled:     true,
      Communities: make(map[string]domain.CommunitySpecific),
    },
    Event: domain.EventPreference{
      Enabled: true, RemindersEnabled: true, ReminderHoursBefore: 24,
    },
    QnA: domain.QnAPreference{
      Enabled: true, SomeoneReplied: true, ReplyFromModerator: true,
    },
  }
}

// Auto-create community preference when user joins
func (s *notificationService) OnUserJoinedCommunity(ctx context.Context, userID, communityID string) error {
  pref, err := s.repo.GetByUserID(ctx, userID)
  if err != nil && !errors.Is(err, domain.ErrNotFound) {
    return err
  }
  if pref == nil {
    pref = s.createDefaultPreferences(userID)
    s.repo.Create(ctx, pref)
  }
  if pref.Community.Communities == nil {
    pref.Community.Communities = make(map[string]domain.CommunitySpecific)
  }
  pref.Community.Communities[communityID] = domain.CommunitySpecific{
    Enabled: true, NewPosts: true, MemberJoined: false, MemberLeft: false,
  }
  return s.repo.Update(ctx, pref)
}
```

---

### Notification Engine Integration

```go
type NotificationEngine struct {
  fcm     *FirebaseMessaging
  prefSvc domain.NotificationService
}

// Single notification
func (e *NotificationEngine) SendNotification(ctx context.Context, userID string, notif *NotificationEvent) error {
  should, err := e.prefSvc.ShouldSendNotification(ctx, userID, notif)
  if err != nil {
    return err
  }
  if !should {
    log.Debugf("notification suppressed for user %s: %s", userID, notif.Type)
    return nil
  }
  return e.fcm.Send(ctx, userID, notif.ToFCMPayload())
}

// Broadcast ke banyak user — gunakan batch query, bukan loop
func (e *NotificationEngine) Broadcast(ctx context.Context, event interface{}) error {
  rule := e.getRuleForEvent(event)
  if rule == nil {
    return fmt.Errorf("no rule for event type")
  }

  recipients := rule.Recipients(ctx, event)

  // ✅ Batch check — satu query untuk semua user
  prefs, err := e.prefSvc.GetByUserIDBatch(ctx, recipients)
  if err != nil {
    return err
  }

  filtered := []string{}
  for _, userID := range recipients {
    pref := prefs[userID]
    should, _ := e.prefSvc.ShouldSendWithPref(pref, rule.ToNotifEvent(event))
    if should {
      filtered = append(filtered, userID)
    }
  }

  if len(filtered) == 0 {
    return nil
  }

  return e.fcm.SendMulticast(ctx, filtered, rule.ToFCMPayload(ctx, event))
}
```

---

## Scalability Considerations

### Database Level

- Gunakan **normalized approach** untuk long-term scalability
- Tambah index pada kolom yang sering di-query:
  ```sql
  CREATE INDEX idx_notif_pref_user_id ON notification_preferences(user_id);
  CREATE INDEX idx_notif_community_pref ON notification_preferences_community(preference_id, community_id);
  ```
- Community preferences bersifat dynamic — auto-created saat user join komunitas
- Untuk user millions+, pertimbangkan partitioning table `notification_preferences` by `user_id`

### Application Level

**Cache preferences di Redis** — preferences bersifat read-heavy, write-rarely. Jangan query DB setiap kali cek notifikasi.

```go
const prefCacheKey = "notif_pref:%s" // %s = userID

func (s *notificationService) getPreference(ctx context.Context, userID string) (*domain.NotificationPreference, error) {
  key := fmt.Sprintf(prefCacheKey, userID)

  // Check cache first
  if cached, err := s.cache.Get(ctx, key); err == nil {
    var pref domain.NotificationPreference
    json.Unmarshal([]byte(cached), &pref)
    return &pref, nil
  }

  // Miss: query DB and cache
  pref, err := s.repo.GetByUserID(ctx, userID)
  if err != nil {
    return nil, err
  }

  data, _ := json.Marshal(pref)
  s.cache.Set(ctx, key, string(data), 1*time.Hour)

  return pref, nil
}

// Invalidate cache setiap kali preferences diupdate
func (s *notificationService) invalidateCache(ctx context.Context, userID string) {
  s.cache.Delete(ctx, fmt.Sprintf(prefCacheKey, userID))
}
```

**Batch preference checking** — untuk broadcast, ambil semua preferences dalam satu query:

```go
// ❌ Jangan — N+1 query
for _, userID := range recipients {
  pref, _ := repo.GetByUserID(ctx, userID)
}

// ✅ Batch — single query
prefs, _ := repo.GetByUserIDBatch(ctx, recipients)
```

### Notification Engine Level

- **Async preference checking** — jangan block notification send jika preference check fail, gunakan default allow
- **Don't let preferences block critical notifications** — misal security alerts, OTP, harus selalu dikirim regardless of preferences

---

## Best Practices

### Do's ✅

1. **Always have defaults** — user baru auto-create dengan sensible defaults saat pertama kali login
2. **Cache aggressively** — preferences are read-heavy, write-rarely; cache di Redis dengan TTL 1 jam
3. **Validate on backend** — jangan trust client; selalu validasi body di handler layer
4. **Deprecate features gracefully** — jika module dihapus, tetap handle old preferences tanpa error
5. **Log preference changes** — untuk audit trail & debugging production issues
6. **Auto-create on community join** — jangan require user setup preferences manual saat join komunitas

### Don'ts ❌

1. **Don't send notification tanpa cek preference** — selalu lewat `ShouldSendNotification()`
2. **Don't expose all settings to user** — hanya show setting yang relevan dengan fitur yang aktif
3. **Don't let preferences block critical notifications** — security alerts, OTP harus bypass preference check
4. **Don't make community preferences required** — auto-create dengan default saat user join
5. **Don't loop N+1 queries for broadcast** — selalu batch query preferences untuk broadcast
