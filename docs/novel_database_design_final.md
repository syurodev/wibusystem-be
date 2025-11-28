# Novel System - Database Design Document (Final Review)

**Version:** 3.1 (Reviewed & Updated)
**Date:** 2025-11-18
**Status:** Ready for Implementation

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Design Review & Findings](#design-review--findings)
3. [Audit & Change Tracking Strategy](#audit--change-tracking-strategy)
4. [Core Requirements](#core-requirements)
5. [Database Architecture](#database-architecture)
6. [Detailed Schema Design](#detailed-schema-design)
7. [Workflows](#workflows)
8. [Performance Considerations](#performance-considerations)
9. [Security & Permissions](#security--permissions)
10. [Implementation Roadmap](#implementation-roadmap)

---

## Overview

### System Purpose

Platform quản lý và dịch light novel với:

- **Dual ownership**: User hoặc Tenant sở hữu novel
- **Ownership transfer**: Chuyển quyền 2 chiều với approval workflow
- **Separate translations**: Synopsis và Chapter độc lập
- **Multi-team support**: Nhiều team dịch cùng ngôn ngữ (collaborative hoặc exclusive)
- **Community contributions**: User đóng góp bản dịch với review process
- **Audit tracking**: Track mọi thay đổi ở application layer

---

## Design Review & Findings

### ✅ Strengths

1. **Clear Separation of Concerns**

   - Synopsis vs Chapter translations tách biệt → Đúng với use case
   - Ownership system rõ ràng với polymorphic design
   - Team collaboration model flexible

2. **Scalability**

   - JSONB cho flexible content
   - Separate tables cho translation types
   - Support horizontal scaling

3. **Flexibility**
   - Multi-team per language
   - Report system cho conflicts
   - Contribution workflow có thể customize

### ⚠️ Issues Found & Fixed

#### Issue 1: Missing Audit Fields

**Problem:** Tables thiếu audit fields (created_by, updated_by, version)
**Fix:** Thêm audit fields vào tất cả core tables

#### Issue 2: Polymorphic Reference Không An Toàn

**Problem:** `novels.owner_id` reference đến 2 tables khác nhau → Không có FK constraint
**Fix:**

- Keep polymorphic design cho flexibility
- Add application-level validation
- Add check constraints để validate owner exists

#### Issue 3: Missing Index cho Ownership Queries

**Problem:** Query theo owner sẽ chậm
**Fix:** Add composite index `(owner_type, owner_id)`

#### Issue 4: Translation Team Assignment Thiếu Status

**Problem:** Không track được team có đang active hay blocked
**Fix:** Add `status` field: active, inactive, blocked

#### Issue 5: Missing Default Language Tracking

**Problem:** Novel phải có ngôn ngữ gốc nhưng không enforce
**Fix:**

- Make `original_language` NOT NULL
- Auto-create `novel_languages` record với `is_original=true`

#### Issue 6: Chapter Content History Có Thể Rất Lớn

**Problem:** Lưu full JSONB content mỗi version sẽ chiếm nhiều storage
**Fix:** Chỉ track metadata changes, không lưu full content snapshot

#### Issue 7: Missing Cascade Behaviors

**Problem:** Một số foreign keys thiếu ON DELETE behaviors
**Fix:** Review và set đúng cascade: CASCADE, SET NULL, RESTRICT

#### Issue 8: Translation Contributions Thiếu Contribution Quality Metrics

**Problem:** Không có cách đánh giá quality của contribution
**Fix:** Add quality_score, reviewer_rating fields

#### Issue 9: Missing Novel Publication Workflow

**Problem:** Không có workflow cho việc publish novel (draft → published)
**Fix:** Có status enum rồi, nhưng cần document workflow rõ ràng

#### Issue 10: Team Member Permissions Quá Đơn Giản

**Problem:** JSONB permissions không structured, khó validate
**Fix:** Define clear permission enum values

---

## Audit & Change Tracking Strategy

### Decision: Application Layer Logging

**Why Not Database Triggers:**

```
❌ Performance overhead on every write
❌ Synchronous blocking
❌ Hard to optimize/disable
❌ Cannot add rich context (IP, user agent, request ID)
❌ Lock contention at scale
❌ Complex debugging
```

**Why Application Layer:**

```
✅ Async/non-blocking
✅ Batch inserts for performance
✅ Rich context (IP, user agent, session, etc.)
✅ Selective logging (skip bulk operations)
✅ Horizontal scaling (multiple workers)
✅ Separate audit database possible
✅ Easy testing and debugging
```

### Architecture

```
┌─────────────────┐
│  API Service    │
│  (Go/Gin)       │
└────────┬────────┘
         │
         ├──────────────────┬────────────────────┐
         │                  │                    │
         v                  v                    v
┌─────────────────┐  ┌─────────────┐   ┌──────────────────┐
│  PostgreSQL     │  │ Redis Queue │   │  Context Info    │
│  (Main DB)      │  │             │   │  (IP, UA, etc.)  │
│  - novels       │  │             │   └──────────────────┘
│  - chapters     │  │             │
│  - translations │  │             │
└─────────────────┘  └──────┬──────┘
                            │
                            v
                     ┌──────────────┐
                     │ Audit Worker │
                     │ (Background) │
                     └──────┬───────┘
                            │
                            v
                     ┌──────────────┐
                     │ PostgreSQL   │
                     │ (catalog.    │
                     │  *_history)  │
                     └──────────────┘
```

### Implementation

#### 1. Minimal Database Support

```sql
-- Add audit fields to core tables
ALTER TABLE catalog.novels
ADD COLUMN created_by UUID NOT NULL REFERENCES identify.users(id),
ADD COLUMN updated_by UUID REFERENCES identify.users(id),
ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- Simple trigger to increment version
CREATE OR REPLACE FUNCTION catalog.increment_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version := OLD.version + 1;
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_novels_version
    BEFORE UPDATE ON catalog.novels
    FOR EACH ROW
    EXECUTE FUNCTION catalog.increment_version();
```

#### 2. History Tables (Detailed Audit)

```sql
CREATE TABLE catalog.novel_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Version & change metadata
    version INTEGER NOT NULL,
    changed_by UUID NOT NULL REFERENCES identify.users(id),
    change_type VARCHAR(20) NOT NULL, -- created, updated, deleted, ownership_changed

    -- What changed (JSONB)
    -- {"title": {"old": "...", "new": "..."}}
    changes JSONB NOT NULL,

    -- Context
    ip_address INET,
    user_agent TEXT,
    request_id UUID,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_novel_history_novel_id ON catalog.novel_history(novel_id);
CREATE INDEX idx_novel_history_changed_by ON catalog.novel_history(changed_by);
CREATE INDEX idx_novel_history_created_at ON catalog.novel_history(created_at DESC);
```

#### 3. Application Service

```go
// AuditService handles change tracking
type AuditService struct {
    queue    chan *AuditLog
    repo     *AuditRepository
    ctx      context.Context
    cancel   context.CancelFunc
}

type AuditLog struct {
    EntityType  string          // "novel", "chapter", "translation"
    EntityID    uuid.UUID
    Version     int
    ChangedBy   uuid.UUID
    ChangeType  string          // "created", "updated", "deleted"
    Changes     json.RawMessage // {"title": {"old": "...", "new": "..."}}
    IPAddress   string
    UserAgent   string
    RequestID   uuid.UUID
}

func (s *AuditService) Start(workers int) {
    s.ctx, s.cancel = context.WithCancel(context.Background())

    for i := 0; i < workers; i++ {
        go s.worker()
    }
}

func (s *AuditService) worker() {
    batch := make([]*AuditLog, 0, 100)
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case log := <-s.queue:
            batch = append(batch, log)

            // Flush when batch is full
            if len(batch) >= 100 {
                s.flushBatch(batch)
                batch = batch[:0]
            }

        case <-ticker.C:
            // Periodic flush
            if len(batch) > 0 {
                s.flushBatch(batch)
                batch = batch[:0]
            }

        case <-s.ctx.Done():
            // Shutdown: flush remaining
            if len(batch) > 0 {
                s.flushBatch(batch)
            }
            return
        }
    }
}

func (s *AuditService) flushBatch(logs []*AuditLog) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := s.repo.InsertBatch(ctx, logs); err != nil {
        log.Error().Err(err).Msg("Failed to insert audit logs")
    }
}

func (s *AuditService) LogChange(log *AuditLog) {
    select {
    case s.queue <- log:
        // Queued successfully
    default:
        // Queue full - log error but don't block
        log.Error().Msg("Audit queue full, dropping log")
    }
}
```

#### 4. Usage in Service Layer

```go
func (s *NovelService) UpdateNovel(
    ctx context.Context,
    id uuid.UUID,
    updates *NovelUpdateDTO,
    userID uuid.UUID,
) error {
    // Get current state
    oldNovel, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    // Apply updates
    newNovel := s.applyUpdates(oldNovel, updates)
    newNovel.UpdatedBy = &userID

    // Update in DB (trigger auto-increments version)
    if err := s.repo.Update(ctx, newNovel); err != nil {
        return err
    }

    // Async audit log (non-blocking)
    s.auditService.LogChange(&AuditLog{
        EntityType: "novel",
        EntityID:   id,
        Version:    newNovel.Version,
        ChangedBy:  userID,
        ChangeType: "updated",
        Changes:    s.detectChanges(oldNovel, newNovel),
        IPAddress:  getIPFromContext(ctx),
        UserAgent:  getUserAgentFromContext(ctx),
        RequestID:  getRequestIDFromContext(ctx),
    })

    return nil
}

func (s *NovelService) detectChanges(old, new *Novel) json.RawMessage {
    changes := make(map[string]any)

    if old.Title != new.Title {
        changes["title"] = map[string]any{
            "old": old.Title,
            "new": new.Title,
        }
    }

    if old.Status != new.Status {
        changes["status"] = map[string]any{
            "old": old.Status,
            "new": new.Status,
        }
    }

    // ... detect other changes

    data, _ := json.Marshal(changes)
    return data
}
```

---

## Core Requirements

### 1. Ownership Model

**Requirements:**

- Novel thuộc về User (cá nhân) hoặc Tenant (nhóm)
- Owner có thể transfer cho nhau
- Transfer cần approval từ receiver

**Design:**

```sql
novels.owner_type: 'user' | 'tenant'
novels.owner_id: UUID (polymorphic reference)
novels.created_by: UUID (actual creator, never changes)
```

### 2. Translation Model

**Requirements:**

- Synopsis và Chapter translations riêng biệt
- Mỗi ngôn ngữ có thể có nhiều contributors
- Official translations vs Community contributions

**Design:**

```sql
-- Synopsis (separate)
novel_synopsis_translations
synopsis_translation_contributions

-- Chapters (separate)
chapter_translations
translation_contributions
```

### 3. Team Collaboration

**Requirements:**

- Nhiều teams có thể dịch cùng ngôn ngữ
- Support exclusive rights (1 team only)
- Report system cho conflicts

**Design:**

```sql
translation_teams (linked to tenants)
novel_team_assignments (team → novel → language)
exclusive_translation_reports (conflict resolution)
```

---

## Database Architecture

### Entity Relationship Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    CATALOG SCHEMA                        │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────┐                                           │
│  │  novels  │ (owner_type, owner_id)                    │
│  └────┬─────┘                                           │
│       │                                                  │
│       ├─── volumes ──── chapters                        │
│       │                                                  │
│       ├─── novel_genres ──── genres                     │
│       ├─── novel_authors ──── authors                   │
│       ├─── novel_artists ──── artists                   │
│       ├─── novel_translators ──── translators           │
│       ├─── novel_team_assignments ──── translation_teams│
│       ├─── novel_synopsis_translations                  │
│       ├─── synopsis_translation_contributions           │
│       ├─── novel_languages                              │
│       ├─── ownership_transfers                          │
│       ├─── exclusive_translation_reports                │
│       └─── novel_history (audit)                        │
│                                                          │
│  chapters                                                │
│       ├─── chapter_translations                         │
│       ├─── translation_contributions                    │
│       └─── chapter_history (audit)                      │
│                                                          │
│  volumes                                                 │
│       └─── volume_history (audit)                       │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

---

## Detailed Schema Design

### 1. Core Tables

#### `catalog.novels`

```sql
CREATE TABLE catalog.novels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic info
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL UNIQUE,

    -- Ownership (Polymorphic - validated in application)
    owner_type VARCHAR(20) NOT NULL CHECK (owner_type IN ('user', 'tenant')),
    owner_id UUID NOT NULL,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Content
    synopsis JSONB, -- Synopsis in ORIGINAL language only
    cover_image_url VARCHAR(1000),
    thumbnail_url VARCHAR(1000),

    -- Original info
    original_language VARCHAR(10) NOT NULL, -- ISO 639-1: vi, en, zh, ja, ko
    original_title VARCHAR(500),

    -- Status
    status catalog.novel_status NOT NULL DEFAULT 'draft',

    -- Statistics (auto-updated by triggers or application)
    total_volumes INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,
    total_words BIGINT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00 CHECK (rating_average >= 0 AND rating_average <= 5),
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Additional metadata
    metadata JSONB DEFAULT '{}',

    -- Dates
    first_published_at TIMESTAMP WITH TIME ZONE,
    last_chapter_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Audit timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_novels_owner ON catalog.novels(owner_type, owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_created_by ON catalog.novels(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_status ON catalog.novels(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_original_language ON catalog.novels(original_language) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_slug ON catalog.novels(slug);
CREATE INDEX idx_novels_synopsis ON catalog.novels USING GIN(synopsis) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_metadata ON catalog.novels USING GIN(metadata) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_created_at ON catalog.novels(created_at DESC) WHERE deleted_at IS NULL;

-- Comments
COMMENT ON COLUMN catalog.novels.owner_type IS 'Polymorphic owner type: user or tenant';
COMMENT ON COLUMN catalog.novels.owner_id IS 'Reference to users.id OR tenants.id based on owner_type';
COMMENT ON COLUMN catalog.novels.created_by IS 'User who created the novel (never changes)';
COMMENT ON COLUMN catalog.novels.updated_by IS 'User who last updated the novel';
COMMENT ON COLUMN catalog.novels.version IS 'Version number, auto-incremented on each update';
COMMENT ON COLUMN catalog.novels.synopsis IS 'Synopsis in ORIGINAL language only. Translations go to novel_synopsis_translations';
```

#### `catalog.volumes`

```sql
CREATE TABLE catalog.volumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Basic info
    volume_number INTEGER NOT NULL CHECK (volume_number > 0),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,
    description TEXT,

    -- Images
    cover_image_url VARCHAR(1000),

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Statistics (auto-updated)
    chapter_count INTEGER NOT NULL DEFAULT 0 CHECK (chapter_count >= 0),
    word_count BIGINT NOT NULL DEFAULT 0 CHECK (word_count >= 0),

    -- Ordering
    display_order INTEGER NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    UNIQUE(novel_id, volume_number),
    UNIQUE(novel_id, slug)
);

CREATE INDEX idx_volumes_novel_id ON catalog.volumes(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_display_order ON catalog.volumes(novel_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_published ON catalog.volumes(novel_id, published_at DESC) WHERE is_published = TRUE AND deleted_at IS NULL;
```

#### `catalog.chapters`

```sql
CREATE TABLE catalog.chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    volume_id UUID REFERENCES catalog.volumes(id) ON DELETE SET NULL,

    -- Basic info
    chapter_number INTEGER NOT NULL CHECK (chapter_number > 0),
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,

    -- Content in ORIGINAL language only
    content JSONB NOT NULL,

    -- Audit fields
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Metrics
    word_count INTEGER NOT NULL DEFAULT 0 CHECK (word_count >= 0),
    character_count INTEGER NOT NULL DEFAULT 0 CHECK (character_count >= 0),

    -- Access control
    is_free BOOLEAN NOT NULL DEFAULT TRUE,
    price DECIMAL(10,2) DEFAULT 0.00 CHECK (price >= 0),
    currency VARCHAR(3) DEFAULT 'VND',

    -- Status
    status catalog.chapter_status NOT NULL DEFAULT 'draft',

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,

    -- Ordering
    display_order INTEGER NOT NULL,

    -- Author notes
    author_notes JSONB,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    deleted_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    UNIQUE(novel_id, chapter_number),
    UNIQUE(novel_id, slug)
);

CREATE INDEX idx_chapters_novel_id ON catalog.chapters(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_volume_id ON catalog.chapters(volume_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_status ON catalog.chapters(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_display_order ON catalog.chapters(novel_id, display_order) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_published ON catalog.chapters(published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
CREATE INDEX idx_chapters_content ON catalog.chapters USING GIN(content) WHERE deleted_at IS NULL;

COMMENT ON COLUMN catalog.chapters.content IS 'Chapter content in ORIGINAL language only. Translations go to chapter_translations';
```

---

### 2. Ownership System

#### `catalog.ownership_transfers`

```sql
CREATE TABLE catalog.ownership_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- From owner
    from_owner_type VARCHAR(20) NOT NULL CHECK (from_owner_type IN ('user', 'tenant')),
    from_owner_id UUID NOT NULL,

    -- To owner
    to_owner_type VARCHAR(20) NOT NULL CHECK (to_owner_type IN ('user', 'tenant')),
    to_owner_id UUID NOT NULL,

    -- Transfer metadata
    initiated_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    approved_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    rejected_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),

    reason TEXT,
    notes TEXT,
    admin_notes TEXT,

    -- Dates
    transferred_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ownership_transfers_novel_id ON catalog.ownership_transfers(novel_id);
CREATE INDEX idx_ownership_transfers_status ON catalog.ownership_transfers(status);
CREATE INDEX idx_ownership_transfers_to_owner ON catalog.ownership_transfers(to_owner_type, to_owner_id) WHERE status = 'pending';
CREATE INDEX idx_ownership_transfers_from_owner ON catalog.ownership_transfers(from_owner_type, from_owner_id);
```

---

### 3. Translation Teams

#### `catalog.translation_teams`

```sql
CREATE TABLE catalog.translation_teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES identify.tenants(id) ON DELETE CASCADE,

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,
    description TEXT,
    logo_url VARCHAR(1000),

    -- Languages
    primary_language VARCHAR(10) NOT NULL,
    supported_languages JSONB NOT NULL DEFAULT '[]',

    -- Contact
    website VARCHAR(500),
    discord_url VARCHAR(500),
    facebook_url VARCHAR(500),
    email VARCHAR(200),

    -- Statistics
    novel_count INTEGER NOT NULL DEFAULT 0,
    chapter_count INTEGER NOT NULL DEFAULT 0,
    member_count INTEGER NOT NULL DEFAULT 0,

    -- Status
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_translation_teams_tenant_id ON catalog.translation_teams(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_teams_slug ON catalog.translation_teams(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_teams_primary_language ON catalog.translation_teams(primary_language) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_teams_languages ON catalog.translation_teams USING GIN(supported_languages);
```

#### `catalog.team_members`

```sql
CREATE TABLE catalog.team_members (
    team_id UUID NOT NULL REFERENCES catalog.translation_teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    role VARCHAR(50) NOT NULL DEFAULT 'translator' CHECK (role IN ('leader', 'translator', 'editor', 'proofreader', 'qc')),

    -- Structured permissions
    can_translate BOOLEAN NOT NULL DEFAULT TRUE,
    can_edit BOOLEAN NOT NULL DEFAULT FALSE,
    can_publish BOOLEAN NOT NULL DEFAULT FALSE,
    can_manage_members BOOLEAN NOT NULL DEFAULT FALSE,

    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,

    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_team_members_user_id ON catalog.team_members(user_id) WHERE left_at IS NULL;
CREATE INDEX idx_team_members_role ON catalog.team_members(team_id, role) WHERE left_at IS NULL;
```

#### `catalog.novel_team_assignments`

```sql
CREATE TABLE catalog.novel_team_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES catalog.translation_teams(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    -- Status (UPDATED)
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'blocked')),

    -- Exclusive rights
    is_exclusive BOOLEAN NOT NULL DEFAULT FALSE,
    exclusive_approved_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    exclusive_approved_at TIMESTAMP WITH TIME ZONE,

    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, team_id, language)
);

CREATE INDEX idx_novel_team_assignments_novel_id ON catalog.novel_team_assignments(novel_id);
CREATE INDEX idx_novel_team_assignments_team_id ON catalog.novel_team_assignments(team_id);
CREATE INDEX idx_novel_team_assignments_language ON catalog.novel_team_assignments(language);
CREATE INDEX idx_novel_team_assignments_exclusive ON catalog.novel_team_assignments(novel_id, language) WHERE is_exclusive = TRUE;
CREATE INDEX idx_novel_team_assignments_status ON catalog.novel_team_assignments(status);
```

#### `catalog.exclusive_translation_reports`

```sql
CREATE TABLE catalog.exclusive_translation_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    -- Reporter (novel owner or authorized user)
    reported_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,

    -- Team being reported
    reported_team_id UUID NOT NULL REFERENCES catalog.translation_teams(id) ON DELETE CASCADE,

    -- Report details
    reason TEXT NOT NULL,
    evidence TEXT,

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'dismissed')),

    -- Action taken
    action_taken VARCHAR(50) CHECK (action_taken IN ('team_blocked', 'warning_issued', 'dismissed', 'under_review')),

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exclusive_reports_novel_id ON catalog.exclusive_translation_reports(novel_id);
CREATE INDEX idx_exclusive_reports_team_id ON catalog.exclusive_translation_reports(reported_team_id);
CREATE INDEX idx_exclusive_reports_status ON catalog.exclusive_translation_reports(status);
```

---

### 4. Synopsis Translations (SEPARATE)

#### `catalog.novel_synopsis_translations`

```sql
CREATE TABLE catalog.novel_synopsis_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    -- Content
    synopsis JSONB NOT NULL,

    -- Translator info
    translator_id UUID REFERENCES catalog.translators(id) ON DELETE SET NULL,
    translation_team_id UUID REFERENCES catalog.translation_teams(id) ON DELETE SET NULL,

    -- Audit
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Metrics
    word_count INTEGER NOT NULL DEFAULT 0,

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(novel_id, language)
);

CREATE INDEX idx_synopsis_translations_novel_id ON catalog.novel_synopsis_translations(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_language ON catalog.novel_synopsis_translations(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_translator_id ON catalog.novel_synopsis_translations(translator_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_team_id ON catalog.novel_synopsis_translations(translation_team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_translations_synopsis ON catalog.novel_synopsis_translations USING GIN(synopsis) WHERE deleted_at IS NULL;
```

#### `catalog.synopsis_translation_contributions`

```sql
CREATE TABLE catalog.synopsis_translation_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    contributor_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    language VARCHAR(10) NOT NULL,

    -- Content
    synopsis JSONB NOT NULL,
    contributor_notes TEXT,

    -- Review
    status catalog.translation_status NOT NULL DEFAULT 'pending_review',

    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,

    -- Quality (ADDED)
    quality_score INTEGER CHECK (quality_score >= 0 AND quality_score <= 100),
    reviewer_rating INTEGER CHECK (reviewer_rating >= 1 AND reviewer_rating <= 5),

    -- Link to official
    official_translation_id UUID REFERENCES catalog.novel_synopsis_translations(id) ON DELETE SET NULL,

    -- Credits
    credit_points INTEGER NOT NULL DEFAULT 0,
    is_credited BOOLEAN NOT NULL DEFAULT FALSE,

    -- Community feedback
    upvote_count INTEGER NOT NULL DEFAULT 0,
    downvote_count INTEGER NOT NULL DEFAULT 0,

    word_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_synopsis_contributions_novel_id ON catalog.synopsis_translation_contributions(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_contributor_id ON catalog.synopsis_translation_contributions(contributor_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_status ON catalog.synopsis_translation_contributions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_synopsis_contributions_language ON catalog.synopsis_translation_contributions(language) WHERE deleted_at IS NULL;
```

---

### 5. Chapter Translations (SEPARATE)

#### `catalog.chapter_translations`

```sql
CREATE TABLE catalog.chapter_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    -- Content
    title VARCHAR(500) NOT NULL,
    content JSONB NOT NULL,
    translator_notes JSONB,

    -- Translator info
    translator_id UUID REFERENCES catalog.translators(id) ON DELETE SET NULL,
    translation_team_id UUID REFERENCES catalog.translation_teams(id) ON DELETE SET NULL,

    -- Audit
    created_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE RESTRICT,
    updated_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    version INTEGER NOT NULL DEFAULT 1,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Metrics
    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Dates
    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(chapter_id, language)
);

CREATE INDEX idx_chapter_translations_chapter_id ON catalog.chapter_translations(chapter_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_language ON catalog.chapter_translations(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_translator_id ON catalog.chapter_translations(translator_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_team_id ON catalog.chapter_translations(translation_team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapter_translations_status ON catalog.chapter_translations(status) WHERE deleted_at IS NULL;
-- NOTE: Do NOT index full content (too large). Index specific fields if needed.
```

#### `catalog.translation_contributions`

```sql
CREATE TABLE catalog.translation_contributions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    contributor_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    language VARCHAR(10) NOT NULL,
    contribution_type catalog.contribution_type NOT NULL,

    -- Content
    title VARCHAR(500),
    content JSONB NOT NULL,
    contributor_notes TEXT,

    -- Review
    status catalog.translation_status NOT NULL DEFAULT 'pending_review',

    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,

    -- Quality (ADDED)
    quality_score INTEGER CHECK (quality_score >= 0 AND quality_score <= 100),
    reviewer_rating INTEGER CHECK (reviewer_rating >= 1 AND reviewer_rating <= 5),

    -- Link to official
    official_translation_id UUID REFERENCES catalog.chapter_translations(id) ON DELETE SET NULL,

    -- Credits
    credit_points INTEGER NOT NULL DEFAULT 0,
    is_credited BOOLEAN NOT NULL DEFAULT FALSE,

    -- Community feedback
    upvote_count INTEGER NOT NULL DEFAULT 0,
    downvote_count INTEGER NOT NULL DEFAULT 0,

    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_translation_contributions_chapter_id ON catalog.translation_contributions(chapter_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_contributor_id ON catalog.translation_contributions(contributor_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_status ON catalog.translation_contributions(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_language ON catalog.translation_contributions(language) WHERE deleted_at IS NULL;
CREATE INDEX idx_translation_contributions_quality ON catalog.translation_contributions(quality_score DESC) WHERE status = 'approved';
```

---

### 6. Audit/History Tables

#### `catalog.novel_history`

```sql
CREATE TABLE catalog.novel_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- Version & metadata
    version INTEGER NOT NULL,
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE SET NULL,
    change_type VARCHAR(20) NOT NULL CHECK (change_type IN ('created', 'updated', 'deleted', 'ownership_changed', 'published')),

    -- What changed
    changes JSONB NOT NULL,

    -- Context
    ip_address INET,
    user_agent TEXT,
    request_id UUID,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_novel_history_novel_id ON catalog.novel_history(novel_id);
CREATE INDEX idx_novel_history_changed_by ON catalog.novel_history(changed_by);
CREATE INDEX idx_novel_history_created_at ON catalog.novel_history(created_at DESC);
CREATE INDEX idx_novel_history_change_type ON catalog.novel_history(change_type);
```

#### `catalog.volume_history`

```sql
CREATE TABLE catalog.volume_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    volume_id UUID NOT NULL REFERENCES catalog.volumes(id) ON DELETE CASCADE,
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    version INTEGER NOT NULL,
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE SET NULL,
    change_type VARCHAR(20) NOT NULL,

    changes JSONB NOT NULL,

    ip_address INET,
    user_agent TEXT,
    request_id UUID,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_volume_history_volume_id ON catalog.volume_history(volume_id);
CREATE INDEX idx_volume_history_novel_id ON catalog.volume_history(novel_id);
CREATE INDEX idx_volume_history_created_at ON catalog.volume_history(created_at DESC);
```

#### `catalog.chapter_history`

```sql
CREATE TABLE catalog.chapter_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    version INTEGER NOT NULL,
    changed_by UUID NOT NULL REFERENCES identify.users(id) ON DELETE SET NULL,
    change_type VARCHAR(20) NOT NULL,

    -- Only metadata changes, NOT full content
    changes JSONB NOT NULL,

    ip_address INET,
    user_agent TEXT,
    request_id UUID,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chapter_history_chapter_id ON catalog.chapter_history(chapter_id);
CREATE INDEX idx_chapter_history_novel_id ON catalog.chapter_history(novel_id);
CREATE INDEX idx_chapter_history_created_at ON catalog.chapter_history(created_at DESC);

COMMENT ON COLUMN catalog.chapter_history.changes IS 'Only metadata changes logged here. Full content not stored due to size.';
```

---

### 7. Other Tables (Keep from previous)

```sql
-- Genres
catalog.genres
catalog.novel_genres

-- Contributors
catalog.authors
catalog.artists
catalog.translators
catalog.novel_authors
catalog.novel_artists
catalog.novel_translators

-- Language tracking
catalog.novel_languages
```

_(Same as v3.0, keep unchanged)_

---

### 8. Enum Types

```sql
-- Novel status
CREATE TYPE catalog.novel_status AS ENUM (
    'draft',
    'ongoing',
    'completed',
    'hiatus',
    'dropped'
);

-- Chapter status
CREATE TYPE catalog.chapter_status AS ENUM (
    'draft',
    'published',
    'scheduled'
);

-- Translation status
CREATE TYPE catalog.translation_status AS ENUM (
    'draft',
    'pending_review',
    'approved',
    'rejected',
    'published'
);

-- Contribution type
CREATE TYPE catalog.contribution_type AS ENUM (
    'new_translation',
    'improvement',
    'proofreading',
    'correction'
);
```

---

## Workflows

### 1. Novel Creation

```
1. User/Tenant creates novel
2. Set owner_type and owner_id
3. Set created_by = current user
4. Set original_language (required)
5. Enter synopsis in original language
6. System auto-creates novel_languages record (is_original=true)
7. Novel status = 'draft'
```

### 2. Ownership Transfer

```
User → Tenant:
1. User initiate transfer
2. Create ownership_transfer (status=pending, to_owner_type=tenant)
3. Tenant admin notified
4. Admin approve/reject
5. If approved:
   - Update novels.owner_type = 'tenant'
   - Update novels.owner_id = tenant_id
   - Update transfer.status = 'approved'
   - Log to novel_history

Tenant → User:
1. Tenant admin initiate
2. Create ownership_transfer (status=pending, to_owner_type=user)
3. User notified
4. User accept/reject
5. Similar update process
```

### 3. Team Translation Assignment

```
1. Team wants to translate novel to language X
2. Check existing assignments for (novel, language)
3. If exclusive rights exist → Cannot assign (blocked)
4. Else → Create novel_team_assignments (status=active)
5. Team can start translating
```

### 4. Exclusive Rights Request

```
1. Novel owner discovers unauthorized team
2. Owner submit exclusive_translation_reports
3. Provide evidence
4. Admin review
5. If approved:
   - Update authorized team: is_exclusive=true
   - Update unauthorized team: status=blocked
   - Optional: Remove unauthorized translations
```

### 5. Synopsis Translation

```
Official:
1. Translator/Team select novel + language
2. Translate synopsis
3. Create novel_synopsis_translations (status=draft)
4. Publish when ready
5. Update novel_languages.synopsis_translated=true

Community:
1. User contribute synopsis translation
2. Create synopsis_translation_contributions (status=pending_review)
3. Moderator review, rate quality
4. If approved → Award credit points
5. Optional: Merge to official translation
```

### 6. Chapter Translation

```
Official:
1. Team assigned to novel + language
2. For each chapter:
   - Translate content
   - Create chapter_translations (status=draft)
   - Team editor review
   - Publish
3. Update novel_languages.chapters_translated++

Community:
1. User contribute chapter translation
2. Choose contribution_type
3. Create translation_contributions
4. Review workflow
5. If approved → Merge to official
```

---

## Performance Considerations

### 1. Index Strategy

**High Priority Indexes:**

```sql
-- Most common queries
idx_novels_owner (owner_type, owner_id)
idx_novels_status
idx_chapters_novel_id
idx_chapter_translations_chapter_id
```

**Partial Indexes:**

```sql
-- Only index active records
WHERE deleted_at IS NULL
WHERE status = 'published'
WHERE is_active = TRUE
```

### 2. JSONB Optimization

**Use GIN indexes selectively:**

```sql
-- Index synopsis for search
CREATE INDEX idx_novels_synopsis ON catalog.novels USING GIN(synopsis);

-- DON'T index large chapter content
-- Index specific extracted fields instead
```

### 3. Pagination

```sql
-- Always use LIMIT and OFFSET
-- Use cursor-based for large datasets
SELECT * FROM catalog.novels
WHERE created_at < $cursor
ORDER BY created_at DESC
LIMIT 20;
```

### 4. Statistics Updates

**Use application triggers, not DB triggers:**

```go
// Update statistics asynchronously
go updateNovelStatistics(novelID)
```

### 5. Audit Log Performance

**Batch inserts:**

```go
// Buffer 100 logs before insert
auditService.FlushBatch(logs)
```

---

## Security & Permissions

### Application-Level Validation

```go
// Validate polymorphic owner exists
func ValidateOwner(ownerType string, ownerID uuid.UUID) error {
    switch ownerType {
    case "user":
        return userRepo.Exists(ownerID)
    case "tenant":
        return tenantRepo.Exists(ownerID)
    default:
        return errors.New("invalid owner_type")
    }
}
```

### Permission Checks

```go
// Check if user can modify novel
func CanModifyNovel(user *User, novel *Novel) bool {
    // Owner check
    if novel.OwnerType == "user" && novel.OwnerID == user.ID {
        return true
    }

    // Tenant member check
    if novel.OwnerType == "tenant" {
        return isTenantMember(user.ID, novel.OwnerID)
    }

    return false
}
```

---

## Implementation Roadmap

### Phase 1: Core Foundation (Week 1-2)

- [ ] Create migrations for core tables (novels, volumes, chapters)
- [ ] Add enum types
- [ ] Create audit fields and version triggers
- [ ] Implement domain models
- [ ] Basic repositories

### Phase 2: Ownership System (Week 3)

- [ ] Ownership transfers table
- [ ] Transfer workflow service
- [ ] Approval system

### Phase 3: Teams & Collaboration (Week 4)

- [ ] Translation teams
- [ ] Team assignments
- [ ] Exclusive rights & reports

### Phase 4: Synopsis Translations (Week 5)

- [ ] Synopsis translations table
- [ ] Contribution workflow
- [ ] Review system

### Phase 5: Chapter Translations (Week 6-7)

- [ ] Chapter translations table
- [ ] Contribution workflow
- [ ] History tracking

### Phase 6: Supporting Features (Week 8)

- [ ] Genres system
- [ ] Authors, artists, translators
- [ ] Language tracking
- [ ] Statistics updates

### Phase 7: Audit System (Week 9)

- [ ] Audit service implementation
- [ ] History tables
- [ ] Queue workers
- [ ] Monitoring

### Phase 8: Testing & Optimization (Week 10)

- [ ] Unit tests
- [ ] Integration tests
- [ ] Performance testing
- [ ] Index optimization

---

## Summary of Changes from v3.0

### ✅ Added

1. Audit fields to all core tables (created_by, updated_by, version, deleted_by)
2. Application-layer audit strategy document
3. Quality metrics to contributions (quality_score, reviewer_rating)
4. Status field to novel_team_assignments (active/inactive/blocked)
5. Request context to history tables (ip_address, user_agent, request_id)
6. Structured permissions to team_members (boolean flags)
7. More check constraints for data validation
8. Performance considerations section
9. Detailed workflows
10. Implementation roadmap

### 🔧 Fixed

1. Made original_language NOT NULL
2. Added proper cascade behaviors
3. Fixed missing indexes for common queries
4. Added comments to polymorphic fields
5. Removed full content indexing for chapters (performance)

### 📝 Clarified

1. Audit logging strategy (application layer, not triggers)
2. Polymorphic reference validation approach
3. History table purposes and scope
4. Permission model for teams

---

**Status:** ✅ Design Complete & Reviewed
**Next Step:** Begin Phase 1 Implementation
**Document Version:** 3.1 (Final)
