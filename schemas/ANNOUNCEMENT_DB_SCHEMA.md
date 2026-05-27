# Database Schema — Announcement Module

> **Stack:** Golang + PostgreSQL  
> **Dibuat:** 2026-05-25  

---

## Daftar Isi

1. [Overview Relasi](#overview-relasi)
2. [PostgreSQL DDL Schema](#postgresql-ddl-schema)
3. [Golang Structs (ORM Mapping)](#golang-structs-orm-mapping)
4. [Migrations (Using migrate/v4)](#migrations-using-migratev4)

---

## Overview Relasi

```
users
  └── announcement_reads           (audit trail read tracking)
  └── announcement_delivery_logs   (delivery status logging)
  └── announcements (created_by)

regions
  └── announcements (region_id)    (regional targeting)
```

---

## PostgreSQL DDL Schema

```sql
-- ============================================================================
-- ANNOUNCEMENT MODULE DATABASE SCHEMA
-- Stack: PostgreSQL 13+
-- Created: 2026-05-25
-- ============================================================================

-- ============================================================================
-- 1. MAIN ANNOUNCEMENTS TABLE
-- ============================================================================

CREATE TABLE announcements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  -- Content
  title VARCHAR(150) NOT NULL,
  body TEXT NOT NULL,
  image_url TEXT NULL,
  
  -- Classification
  type VARCHAR(20) NOT NULL,  -- disaster, system, urgent, info
  priority VARCHAR(20) NOT NULL,  -- critical, high, medium, low
  
  -- Scope
  scope VARCHAR(20) NOT NULL,  -- global, regional
  region_id UUID NULL REFERENCES regions(id),
  
  -- Key columns untuk efficient query (indexed)
  disaster_epicenter_lat NUMERIC NULL,
  disaster_epicenter_lng NUMERIC NULL,
  disaster_subtype VARCHAR(30) NULL,  -- earthquake, flood, landslide, etc
  
  system_subtype VARCHAR(30) NULL,  -- maintenance, outage, degraded, update, incident
  system_maintenance_start TIMESTAMPTZ NULL,
  system_maintenance_end TIMESTAMPTZ NULL,
  
  urgent_subtype VARCHAR(30) NULL,  -- security_alert, account_action, policy_change, etc
  urgent_deadline TIMESTAMPTZ NULL,
  
  info_subtype VARCHAR(30) NULL,  -- feature_launch, event, promo, policy_update, company_news
  
  -- Flexible metadata untuk type-specific data
  metadata JSONB NULL,
  
  -- Status & Lifecycle
  status VARCHAR(20) NOT NULL DEFAULT 'draft',  -- draft, published, archived
  published_at TIMESTAMPTZ NULL,
  expires_at TIMESTAMPTZ NULL,
  archived_at TIMESTAMPTZ NULL,
  
  -- Creator & Timestamps
  created_by UUID NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_by UUID NULL REFERENCES users(id),
  
  -- Analytics (cached)
  total_recipients INTEGER DEFAULT 0,
  total_sent INTEGER DEFAULT 0,
  total_read INTEGER DEFAULT 0,
  
  -- Constraints
  CONSTRAINT announcement_scope_check CHECK (
    (scope = 'global' AND region_id IS NULL) OR
    (scope = 'regional' AND region_id IS NOT NULL)
  ),
  
  CONSTRAINT announcement_valid_type CHECK (
    type IN ('disaster', 'system', 'urgent', 'info')
  ),
  
  CONSTRAINT announcement_valid_priority CHECK (
    priority IN ('critical', 'high', 'medium', 'low')
  ),
  
  CONSTRAINT announcement_valid_status CHECK (
    status IN ('draft', 'published', 'archived')
  ),
  
  CONSTRAINT announcement_expires_future CHECK (
    expires_at IS NULL OR expires_at > created_at
  )
);

-- Indexes untuk efficient queries
CREATE INDEX idx_announcements_type 
  ON announcements(type);

CREATE INDEX idx_announcements_scope 
  ON announcements(scope);

CREATE INDEX idx_announcements_status 
  ON announcements(status);

CREATE INDEX idx_announcements_priority 
  ON announcements(priority);

CREATE INDEX idx_announcements_published_at 
  ON announcements(published_at DESC) 
  WHERE status = 'published';

CREATE INDEX idx_announcements_created_at 
  ON announcements(created_at DESC);

CREATE INDEX idx_announcements_region_id 
  ON announcements(region_id) 
  WHERE scope = 'regional';

CREATE INDEX idx_announcements_created_by 
  ON announcements(created_by);

-- Disaster-specific indexes (for location queries)
CREATE INDEX idx_announcements_disaster_location 
  ON announcements(disaster_epicenter_lat, disaster_epicenter_lng) 
  WHERE type = 'disaster' AND status = 'published';

CREATE INDEX idx_announcements_disaster_subtype 
  ON announcements(disaster_subtype) 
  WHERE type = 'disaster';

-- System-specific indexes
CREATE INDEX idx_announcements_system_maintenance 
  ON announcements(system_maintenance_start, system_maintenance_end) 
  WHERE type = 'system' AND status = 'published';

-- Urgent-specific indexes
CREATE INDEX idx_announcements_urgent_deadline 
  ON announcements(urgent_deadline) 
  WHERE type = 'urgent' AND status = 'published';

-- JSONB index untuk metadata queries
CREATE INDEX idx_announcements_metadata 
  ON announcements USING GIN(metadata);

-- Composite indexes untuk common queries
CREATE INDEX idx_announcements_scope_status 
  ON announcements(scope, status);

CREATE INDEX idx_announcements_type_priority_published 
  ON announcements(type, priority, published_at DESC) 
  WHERE status = 'published';

CREATE INDEX idx_announcements_region_published 
  ON announcements(region_id, published_at DESC) 
  WHERE status = 'published' AND scope = 'regional';

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION update_announcements_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER announcements_updated_at_trigger
BEFORE UPDATE ON announcements
FOR EACH ROW
EXECUTE FUNCTION update_announcements_updated_at();

-- ============================================================================
-- 2. ANNOUNCEMENT READS TABLE (Audit trail)
-- ============================================================================

CREATE TABLE announcement_reads (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  read_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  platform VARCHAR(20) NULL,  -- android, ios, web
  
  -- Prevent duplicate reads dari same user
  CONSTRAINT announcement_reads_unique UNIQUE(announcement_id, user_id)
);

-- Indexes untuk efficient read tracking
CREATE INDEX idx_announcement_reads_announcement_id 
  ON announcement_reads(announcement_id);

CREATE INDEX idx_announcement_reads_user_id 
  ON announcement_reads(user_id);

CREATE INDEX idx_announcement_reads_read_at 
  ON announcement_reads(read_at DESC);

-- Composite index untuk analytics (read rate per announcement)
CREATE INDEX idx_announcement_reads_announcement_user 
  ON announcement_reads(announcement_id, user_id);

-- ============================================================================
-- 3. ANNOUNCEMENT DELIVERY LOG TABLE (Optional, untuk debugging)
-- ============================================================================

CREATE TABLE announcement_delivery_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  
  announcement_id UUID NOT NULL REFERENCES announcements(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  
  -- Delivery status
  status VARCHAR(20) NOT NULL,  -- sent, failed, skipped
  reason VARCHAR(255) NULL,  -- jika failed/skipped, apa alasannya
  
  -- Notification channels
  push_notification_sent BOOLEAN DEFAULT FALSE,
  email_sent BOOLEAN DEFAULT FALSE,
  in_app_notification BOOLEAN DEFAULT FALSE,
  
  sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  
  CONSTRAINT delivery_status_check CHECK (
    status IN ('sent', 'failed', 'skipped')
  )
);

-- Indexes untuk troubleshooting
CREATE INDEX idx_announcement_delivery_logs_announcement_id 
  ON announcement_delivery_logs(announcement_id);

CREATE INDEX idx_announcement_delivery_logs_user_id 
  ON announcement_delivery_logs(user_id);

CREATE INDEX idx_announcement_delivery_logs_status 
  ON announcement_delivery_logs(status);

CREATE INDEX idx_announcement_delivery_logs_sent_at 
  ON announcement_delivery_logs(sent_at DESC);

-- ============================================================================
-- 4. ANNOUNCEMENT VIEWS (Materialized Views untuk analytics)
-- ============================================================================

-- View: Announcement Statistics
CREATE MATERIALIZED VIEW announcement_stats AS
SELECT 
  a.id,
  a.title,
  a.type,
  a.priority,
  a.scope,
  a.region_id,
  a.status,
  a.published_at,
  a.total_recipients,
  a.total_sent,
  COUNT(DISTINCT ar.user_id) as total_read,
  CASE 
    WHEN a.total_sent > 0 
    THEN ROUND((COUNT(DISTINCT ar.user_id)::NUMERIC / a.total_sent) * 100, 2)
    ELSE 0 
  END as read_rate,
  COUNT(DISTINCT CASE WHEN ar.platform = 'android' THEN ar.user_id END) as android_reads,
  COUNT(DISTINCT CASE WHEN ar.platform = 'ios' THEN ar.user_id END) as ios_reads,
  COUNT(DISTINCT CASE WHEN ar.platform = 'web' THEN ar.user_id END) as web_reads
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id
WHERE a.status = 'published'
GROUP BY a.id, a.title, a.type, a.priority, a.scope, a.region_id, 
         a.status, a.published_at, a.total_recipients, a.total_sent;

CREATE INDEX idx_announcement_stats_type 
  ON announcement_stats(type);

CREATE INDEX idx_announcement_stats_published_at 
  ON announcement_stats(published_at DESC);

-- View: Disaster Announcements dengan Location
CREATE MATERIALIZED VIEW disaster_announcements_with_location AS
SELECT 
  a.id,
  a.title,
  a.body,
  a.disaster_subtype,
  a.disaster_epicenter_lat,
  a.disaster_epicenter_lng,
  a.priority,
  a.scope,
  a.region_id,
  a.status,
  a.published_at,
  a.expires_at,
  a.metadata->'magnitude' as magnitude,
  a.metadata->'depth_km' as depth_km,
  a.metadata->'affected_areas' as affected_areas,
  a.metadata->'evacuation_centers' as evacuation_centers,
  a.total_sent
FROM announcements a
WHERE a.type = 'disaster' 
  AND a.status = 'published'
  AND a.disaster_subtype IS NOT NULL;

-- View: Active Announcements untuk User (based on region & scope)
CREATE OR REPLACE VIEW user_active_announcements AS
SELECT 
  a.id,
  a.title,
  a.body,
  a.type,
  a.priority,
  a.scope,
  a.published_at,
  COALESCE(ar.read_at IS NOT NULL, FALSE) as is_read
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id
WHERE a.status = 'published'
  AND (a.expires_at IS NULL OR a.expires_at > NOW());

-- Refresh materialized views (perlu dijalankan periodically)
-- REFRESH MATERIALIZED VIEW announcement_stats;
-- REFRESH MATERIALIZED VIEW disaster_announcements_with_location;

-- ============================================================================
-- 5. HELPER FUNCTIONS
-- ============================================================================

-- Function: Calculate total recipients untuk announcement
CREATE OR REPLACE FUNCTION get_announcement_recipients(
  p_announcement_id UUID
) RETURNS INTEGER AS $$
DECLARE
  v_scope VARCHAR;
  v_region_id UUID;
  v_count INTEGER;
BEGIN
  -- Get scope dan region_id
  SELECT scope, region_id INTO v_scope, v_region_id
  FROM announcements
  WHERE id = p_announcement_id;
  
  -- Calculate recipients berdasarkan scope
  IF v_scope = 'global' THEN
    SELECT COUNT(*) INTO v_count
    FROM users
    WHERE deleted_at IS NULL;
  ELSIF v_scope = 'regional' THEN
    SELECT COUNT(*) INTO v_count
    FROM users
    WHERE deleted_at IS NULL AND region_id = v_region_id;
  END IF;
  
  RETURN COALESCE(v_count, 0);
END;
$$ LANGUAGE plpgsql;

-- Function: Mark announcement as published (update total_recipients & total_sent)
CREATE OR REPLACE FUNCTION publish_announcement(
  p_announcement_id UUID
) RETURNS TABLE (
  announcement_id UUID,
  status VARCHAR,
  total_recipients INTEGER,
  total_sent INTEGER
) AS $$
DECLARE
  v_recipients INTEGER;
  v_sent INTEGER;
BEGIN
  -- Calculate recipients
  v_recipients := get_announcement_recipients(p_announcement_id);
  
  -- Simulate: 98% delivery rate (real akan dari FCM queue result)
  v_sent := CEIL(v_recipients * 0.98);
  
  -- Update announcement
  UPDATE announcements
  SET 
    status = 'published',
    published_at = NOW(),
    total_recipients = v_recipients,
    total_sent = v_sent,
    updated_at = NOW()
  WHERE id = p_announcement_id
  RETURNING 
    p_announcement_id,
    status,
    total_recipients,
    total_sent
  INTO announcement_id, status, total_recipients, total_sent;
  
  RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- Function: Update read count (untuk trigger)
CREATE OR REPLACE FUNCTION update_announcement_read_count()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE announcements
  SET total_read = (
    SELECT COUNT(DISTINCT user_id)
    FROM announcement_reads
    WHERE announcement_id = NEW.announcement_id
  )
  WHERE id = NEW.announcement_id;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER announcement_reads_count_trigger
AFTER INSERT ON announcement_reads
FOR EACH ROW
EXECUTE FUNCTION update_announcement_read_count();

-- Function: Auto-archive expired announcements (run via cron)
CREATE OR REPLACE FUNCTION archive_expired_announcements()
RETURNS TABLE (archived_count INTEGER) AS $$
BEGIN
  UPDATE announcements
  SET 
    status = 'archived',
    archived_at = NOW(),
    updated_at = NOW()
  WHERE status = 'published'
    AND expires_at IS NOT NULL
    AND expires_at <= NOW();
  
  GET DIAGNOSTICS archived_count = ROW_COUNT;
  RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 6. SAMPLE QUERIES
-- ============================================================================

-- Query 1: Get all published announcements untuk user di Jakarta
/*
SELECT 
  a.id,
  a.title,
  a.type,
  a.priority,
  COALESCE(ar.read_at IS NOT NULL, FALSE) as is_read
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id 
  AND ar.user_id = 'user_123'
WHERE a.status = 'published'
  AND (a.expires_at IS NULL OR a.expires_at > NOW())
  AND (
    a.scope = 'global'
    OR (a.scope = 'regional' AND a.region_id = 'region_jakarta')
  )
ORDER BY a.published_at DESC;
*/

-- Query 2: Get disaster announcements with location dalam radius
/*
SELECT 
  a.id,
  a.title,
  a.disaster_epicenter_lat,
  a.disaster_epicenter_lng,
  a.metadata->'magnitude' as magnitude,
  a.metadata->'affected_areas' as affected_areas
FROM announcements a
WHERE a.type = 'disaster'
  AND a.status = 'published'
  AND ST_DWithin(
    ST_Point(a.disaster_epicenter_lng, a.disaster_epicenter_lat),
    ST_Point(-6.2088, 106.8905),  -- Jakarta
    100000  -- 100km radius
  )
ORDER BY a.published_at DESC;
*/

-- Query 3: Get announcement stats
/*
SELECT 
  a.id,
  a.title,
  a.total_recipients,
  a.total_sent,
  COUNT(DISTINCT ar.user_id) as total_read,
  ROUND((COUNT(DISTINCT ar.user_id)::NUMERIC / a.total_sent) * 100, 2) as read_rate
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id
WHERE a.status = 'published'
GROUP BY a.id, a.title, a.total_recipients, a.total_sent
ORDER BY a.published_at DESC;
*/

-- Query 4: Get unread announcements untuk user
/*
SELECT 
  a.id,
  a.title,
  a.type,
  a.priority,
  a.published_at
FROM announcements a
LEFT JOIN announcement_reads ar ON a.id = ar.announcement_id AND ar.user_id = 'user_123'
WHERE a.status = 'published'
  AND (a.expires_at IS NULL OR a.expires_at > NOW())
  AND ar.id IS NULL  -- not read
  AND (a.scope = 'global' OR (a.scope = 'regional' AND a.region_id = 'region_jakarta'))
ORDER BY a.priority DESC, a.published_at DESC;
*/

-- Query 5: Get announcements created by specific admin
/*
SELECT 
  a.id,
  a.title,
  a.type,
  a.status,
  a.scope,
  r.name as region_name,
  a.total_recipients,
  a.total_sent,
  a.created_at
FROM announcements a
LEFT JOIN regions r ON a.region_id = r.id
WHERE a.created_by = 'admin_jakarta_001'
ORDER BY a.created_at DESC;
*/

-- ============================================================================
-- 7. MAINTENANCE TASKS (Run via cron jobs)
-- ============================================================================

-- Daily: Archive expired announcements
-- SELECT archive_expired_announcements();

-- Weekly: Refresh materialized views
-- REFRESH MATERIALIZED VIEW announcement_stats;
-- REFRESH MATERIALIZED VIEW disaster_announcements_with_location;

-- Monthly: Cleanup old delivery logs (keep 90 days)
-- DELETE FROM announcement_delivery_logs WHERE sent_at < NOW() - INTERVAL '90 days';

-- ============================================================================
-- 8. GRANTS & SECURITY
-- ============================================================================

-- Grant appropriate permissions ke application role
-- (adjust sesuai setup security Anda)

-- GRANT SELECT ON announcements TO app_read_role;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON announcements TO app_write_role;
-- GRANT SELECT ON announcement_reads TO app_read_role;
-- GRANT SELECT, INSERT ON announcement_reads TO app_write_role;

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
```

---

## Golang Structs (ORM Mapping)

```go
// file: internal/domain/announcement/announcement.go

package announcement

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
)

// Announcement represents a broadcast announcement
type Announcement struct {
	ID                      string          `db:"id" json:"id"`
	Title                   string          `db:"title" json:"title"`
	Body                    string          `db:"body" json:"body"`
	ImageURL                *string         `db:"image_url" json:"image_url,omitempty"`
	
	Type                    string          `db:"type" json:"type"`                        // disaster, system, urgent, info
	Priority                string          `db:"priority" json:"priority"`                // critical, high, medium, low
	
	Scope                   string          `db:"scope" json:"scope"`                      // global, regional
	RegionID                *string         `db:"region_id" json:"region_id,omitempty"`
	
	// Key columns for disaster
	DisasterEpicenterLat    *float64        `db:"disaster_epicenter_lat" json:"epicenter_lat,omitempty"`
	DisasterEpicenterLng    *float64        `db:"disaster_epicenter_lng" json:"epicenter_lng,omitempty"`
	DisasterSubtype         *string         `db:"disaster_subtype" json:"disaster_subtype,omitempty"`
	
	// Key columns for system
	SystemSubtype           *string         `db:"system_subtype" json:"system_subtype,omitempty"`
	SystemMaintenanceStart  *time.Time      `db:"system_maintenance_start" json:"maintenance_start,omitempty"`
	SystemMaintenanceEnd    *time.Time      `db:"system_maintenance_end" json:"maintenance_end,omitempty"`
	
	// Key columns for urgent
	UrgentSubtype           *string         `db:"urgent_subtype" json:"urgent_subtype,omitempty"`
	UrgentDeadline          *time.Time      `db:"urgent_deadline" json:"urgent_deadline,omitempty"`
	
	// Key columns for info
	InfoSubtype             *string         `db:"info_subtype" json:"info_subtype,omitempty"`
	
	// Flexible metadata
	Metadata                json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	
	Status                  string          `db:"status" json:"status"`                    // draft, published, archived
	PublishedAt             *time.Time      `db:"published_at" json:"published_at,omitempty"`
	ExpiresAt               *time.Time      `db:"expires_at" json:"expires_at,omitempty"`
	ArchivedAt              *time.Time      `db:"archived_at" json:"archived_at,omitempty"`
	
	CreatedBy               string          `db:"created_by" json:"created_by"`
	CreatedAt               time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time       `db:"updated_at" json:"updated_at"`
	UpdatedBy               *string         `db:"updated_by" json:"updated_by,omitempty"`
	
	TotalRecipients         int             `db:"total_recipients" json:"total_recipients"`
	TotalSent               int             `db:"total_sent" json:"total_sent"`
	TotalRead               int             `db:"total_read" json:"total_read,omitempty"`
	
	// Runtime fields (not in DB)
	ReadRate                *float64        `db:"-" json:"read_rate,omitempty"`
	IsRead                  bool            `db:"-" json:"is_read,omitempty"`
}

// AnnouncementRead represents a read record
type AnnouncementRead struct {
	ID              string    `db:"id" json:"id"`
	AnnouncementID  string    `db:"announcement_id" json:"announcement_id"`
	UserID          string    `db:"user_id" json:"user_id"`
	ReadAt          time.Time `db:"read_at" json:"read_at"`
	Platform        *string   `db:"platform" json:"platform,omitempty"`
}

// AnnouncementDeliveryLog represents a delivery log
type AnnouncementDeliveryLog struct {
	ID                      string     `db:"id" json:"id"`
	AnnouncementID          string     `db:"announcement_id" json:"announcement_id"`
	UserID                  string     `db:"user_id" json:"user_id"`
	Status                  string     `db:"status" json:"status"`                      // sent, failed, skipped
	Reason                  *string    `db:"reason" json:"reason,omitempty"`
	PushNotificationSent    bool       `db:"push_notification_sent" json:"push_notification_sent"`
	EmailSent               bool       `db:"email_sent" json:"email_sent"`
	InAppNotification       bool       `db:"in_app_notification" json:"in_app_notification"`
	SentAt                  time.Time  `db:"sent_at" json:"sent_at"`
}

// DisasterMetadata represents disaster-specific metadata
type DisasterMetadata struct {
	Subtype            string              `json:"disaster_subtype"`
	Magnitude          *float64            `json:"magnitude,omitempty"`
	DepthKm            *int                `json:"depth_km,omitempty"`
	Location           Location            `json:"location"`
	AffectedAreas      []Location          `json:"affected_areas,omitempty"`
	EvacuationCenters  []EvacuationCenter  `json:"evacuation_centers,omitempty"`
	Helpline           *string             `json:"helpline,omitempty"`
}

// SystemMetadata represents system-specific metadata
type SystemMetadata struct {
	Subtype                string            `json:"system_subtype"`
	AffectedServices       []string          `json:"affected_services,omitempty"`
	MaintenanceWindow      MaintenanceWindow `json:"maintenance_window,omitempty"`
	EstimatedDurationHours *int              `json:"estimated_duration_hours,omitempty"`
	Impact                 *string           `json:"impact,omitempty"`
	StatusPageURL          *string           `json:"status_page_url,omitempty"`
}

// UrgentMetadata represents urgent-specific metadata
type UrgentMetadata struct {
	Subtype         string      `json:"urgent_subtype"`
	ActionRequired  bool        `json:"action_required"`
	ActionType      *string     `json:"action_type,omitempty"`
	Details         interface{} `json:"details,omitempty"`
	ActionButton    ActionButton `json:"action_button,omitempty"`
	Escalation      Escalation  `json:"escalation,omitempty"`
}

// InfoMetadata represents info-specific metadata
type InfoMetadata struct {
	Subtype             string      `json:"info_subtype"`
	ActionEncouraged    bool        `json:"action_encouraged"`
	ActionType          *string     `json:"action_type,omitempty"`
	Details             interface{} `json:"details,omitempty"`
	ActionButton        ActionButton `json:"action_button,omitempty"`
	EngagementSettings  Engagement  `json:"engagement,omitempty"`
}

// Supporting types
type Location struct {
	Name      string   `json:"name"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Address   *string  `json:"address,omitempty"`
	ImpactLevel *string `json:"impact_level,omitempty"`
}

type EvacuationCenter struct {
	Name     string   `json:"name"`
	Latitude float64  `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Capacity int      `json:"capacity"`
	Contact  *string  `json:"contact,omitempty"`
	Available *int    `json:"available,omitempty"`
}

type MaintenanceWindow struct {
	StartTime              time.Time `json:"start_time"`
	EndTime                time.Time `json:"end_time"`
	Timezone               *string   `json:"timezone,omitempty"`
}

type ActionButton struct {
	Label      string `json:"label"`
	URL        string `json:"url"`
	ActionType string `json:"action_type"`  // deeplink, external_url, phone_call
}

type Escalation struct {
	AutoBlockAfter string `json:"auto_block_after,omitempty"`
	SupportLink    string `json:"support_link,omitempty"`
}

type Engagement struct {
	ShareEnabled     bool `json:"share_enabled"`
	SaveEnabled      bool `json:"save_enabled"`
	AddToCalendar    bool `json:"add_to_calendar,omitempty"`
	BookmarkEnabled  bool `json:"bookmark_enabled,omitempty"`
}

// Constants
const (
	TypeDisaster = "disaster"
	TypeSystem   = "system"
	TypeUrgent   = "urgent"
	TypeInfo     = "info"

	PriorityCritical = "critical"
	PriorityHigh     = "high"
	PriorityMedium   = "medium"
	PriorityLow      = "low"

	ScopeGlobal    = "global"
	ScopeRegional  = "regional"

	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"

	DisasterEarthquake = "earthquake"
	DisasterFlood      = "flood"
	DisasterLandslide  = "landslide"

	SystemMaintenance = "maintenance"
	SystemOutage      = "outage"
	SystemDegraded    = "degraded_service"

	InfoFeature = "feature_launch"
	InfoEvent   = "event"
	InfoPromo   = "promo"
)

// Helper methods
func (a *Announcement) GetDisasterMetadata() (*DisasterMetadata, error) {
	if a.Type != TypeDisaster {
		return nil, ErrNotDisaster
	}
	var meta DisasterMetadata
	if err := json.Unmarshal(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (a *Announcement) GetSystemMetadata() (*SystemMetadata, error) {
	if a.Type != TypeSystem {
		return nil, ErrNotSystem
	}
	var meta SystemMetadata
	if err := json.Unmarshal(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (a *Announcement) GetUrgentMetadata() (*UrgentMetadata, error) {
	if a.Type != TypeUrgent {
		return nil, ErrNotUrgent
	}
	var meta UrgentMetadata
	if err := json.Unmarshal(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (a *Announcement) GetInfoMetadata() (*InfoMetadata, error) {
	if a.Type != TypeInfo {
		return nil, ErrNotInfo
	}
	var meta InfoMetadata
	if err := json.Unmarshal(a.Metadata, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

// IsPublished checks if announcement is published
func (a *Announcement) IsPublished() bool {
	return a.Status == StatusPublished && a.PublishedAt != nil
}

// IsExpired checks if announcement has expired
func (a *Announcement) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

// ByPlaintext returns plain text metadata (untuk logging)
func (a *Announcement) MetadataPlaintext() string {
	if a.Metadata == nil {
		return ""
	}
	return string(a.Metadata)
}
```

---

## Migrations (Using migrate/v4)

### migrations/000001_create_announcements_table.up.sql

```sql
-- Create announcements table with all fields and indexes
-- (copy the schema dari atas)
```

### migrations/000001_create_announcements_table.down.sql

```sql
DROP MATERIALIZED VIEW IF EXISTS disaster_announcements_with_location CASCADE;
DROP MATERIALIZED VIEW IF EXISTS announcement_stats CASCADE;
DROP VIEW IF EXISTS user_active_announcements CASCADE;

DROP TRIGGER IF EXISTS announcements_updated_at_trigger ON announcements;
DROP FUNCTION IF EXISTS update_announcements_updated_at();

DROP TRIGGER IF EXISTS announcement_reads_count_trigger ON announcement_reads;
DROP FUNCTION IF EXISTS update_announcement_read_count();

DROP TABLE IF EXISTS announcement_delivery_logs;
DROP TABLE IF EXISTS announcement_reads;
DROP TABLE IF EXISTS announcements;

DROP FUNCTION IF EXISTS archive_expired_announcements();
DROP FUNCTION IF EXISTS publish_announcement(UUID);
DROP FUNCTION IF EXISTS get_announcement_recipients(UUID);
```

---

*Dokumen ini adalah schema, domain struct, dan migration script reference untuk modul Announcement.*
