# Novel System - Database Design Document

**Version:** 3.0
**Date:** 2025-11-18
**Status:** Design Phase - Pre-Implementation

---

## 📋 Table of Contents

1. [Overview](#overview)
2. [Core Requirements](#core-requirements)
3. [Database Architecture](#database-architecture)
4. [Detailed Schema Design](#detailed-schema-design)
5. [Ownership System](#ownership-system)
6. [Translation System](#translation-system)
7. [Teams & Collaboration](#teams--collaboration)
8. [Workflows](#workflows)
9. [Permissions & Security](#permissions--security)
10. [Implementation Plan](#implementation-plan)

---

## Overview

### System Purpose

Hệ thống quản lý novel với khả năng:
- Đăng novel bởi cá nhân (user) hoặc nhóm (tenant)
- Chuyển ownership giữa user và tenant
- Dịch synopsis và chapter content riêng biệt
- Hỗ trợ nhiều team dịch cùng lúc
- Community contributions với review workflow
- Report system cho exclusive translation rights

### Key Features

✅ **Dual Ownership**: User hoặc Tenant làm chủ novel
✅ **Ownership Transfer**: Chuyển quyền sở hữu 2 chiều với approval
✅ **Separate Translations**: Synopsis và Chapter translations độc lập
✅ **Multi-Team Support**: Nhiều team dịch cùng ngôn ngữ
✅ **Exclusive Rights**: Report system cho độc quyền dịch
✅ **Version Control**: Translation history tracking
✅ **Community Driven**: Contributions với credit system

---

## Core Requirements

### 1. Ownership Model

**Requirement:**
- Novel có thể thuộc về `user` (cá nhân) hoặc `tenant` (nhóm/tổ chức)
- Owner có thể chuyển novel cho nhau
- Transfer cần approval từ admin tenant (nếu chuyển cho tenant)

**Use Cases:**
```
UC1: User đăng novel cá nhân
UC2: Tenant (team dịch) đăng novel
UC3: User chuyển novel cho tenant của mình
UC4: Tenant chuyển novel cho user (member)
UC5: Tenant admin approve/reject transfer request
```

### 2. Translation Model

**Requirement:**
- **Synopsis translation**: Riêng biệt, ít thay đổi
- **Chapter translation**: Riêng biệt, thường xuyên update
- Một ngôn ngữ có thể có nhiều contributors
- Contributors chọn ngôn ngữ khi đóng góp

**Rationale:**
```
Synopsis:
- Dịch 1 lần, hiếm update
- Khối lượng nhỏ
- Review đơn giản
- Một user có thể chỉ dịch synopsis không dịch chapters

Chapter:
- Dịch liên tục (mỗi chapter mới)
- Khối lượng lớn
- Review phức tạp hơn
- Requires ongoing commitment
```

### 3. Team Collaboration

**Requirement:**
- 1 novel có thể có nhiều team dịch cùng 1 ngôn ngữ
- Tác giả có thể muốn exclusive rights (báo cáo team khác)
- Team có member với roles khác nhau

**Workflow:**
```
1. Team A register để dịch novel X sang tiếng Việt
2. Team B cũng muốn dịch novel X sang tiếng Việt
3. Cả 2 team có thể dịch song song
4. Nếu tác giả muốn exclusive cho Team A:
   → Tác giả report Team B
   → Admin review
   → Có thể block Team B
```

---

## Database Architecture

### Schema Organization

```
catalog (Novel Content Schema)
├── Core Entities
│   ├── novels (top level)
│   ├── volumes (middle level)
│   └── chapters (bottom level)
│
├── Ownership & Access
│   ├── novel_ownership (polymorphic: user/tenant)
│   ├── ownership_transfers (transfer workflow)
│   └── exclusive_translation_reports (report system)
│
├── Classification
│   ├── genres (hierarchical)
│   └── novel_genres (many-to-many)
│
├── Contributors
│   ├── authors
│   ├── artists
│   ├── translators
│   ├── novel_authors (junction)
│   ├── novel_artists (junction)
│   └── novel_translators (junction)
│
├── Translation Teams
│   ├── translation_teams (linked to tenants)
│   ├── team_members (team composition)
│   └── novel_team_assignments (team → novel → language)
│
├── Synopsis Translations (SEPARATE)
│   ├── novel_synopsis_translations (official)
│   └── synopsis_translation_contributions (community)
│
├── Chapter Translations (SEPARATE)
│   ├── chapter_translations (official)
│   ├── translation_contributions (community)
│   └── translation_history (version control)
│
└── Language Tracking
    └── novel_languages (available languages per novel)
```

---

## Detailed Schema Design

### 1. Core Tables

#### `catalog.novels`

```sql
CREATE TABLE catalog.novels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Basic Info
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL UNIQUE,

    -- Ownership (Polymorphic)
    owner_type VARCHAR(20) NOT NULL, -- 'user' or 'tenant'
    owner_id UUID NOT NULL, -- User ID or Tenant ID
    created_by UUID NOT NULL REFERENCES identify.users(id), -- Actual creator

    -- Synopsis in ORIGINAL language
    synopsis JSONB,

    -- Images
    cover_image_url VARCHAR(1000),
    thumbnail_url VARCHAR(1000),

    -- Original Info
    original_language VARCHAR(10) NOT NULL, -- ISO 639-1 (vi, en, zh, ja, ko)
    original_title VARCHAR(500),

    -- Status
    status catalog.novel_status NOT NULL DEFAULT 'draft',

    -- Statistics (auto-updated by triggers)
    total_volumes INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,
    total_words BIGINT NOT NULL DEFAULT 0,
    view_count BIGINT NOT NULL DEFAULT 0,
    favorite_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER NOT NULL DEFAULT 0,

    -- Additional metadata
    metadata JSONB DEFAULT '{}',

    -- Dates
    first_published_at TIMESTAMP WITH TIME ZONE,
    last_chapter_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    CONSTRAINT novels_rating_check CHECK (rating_average >= 0 AND rating_average <= 5),
    CONSTRAINT novels_owner_type_check CHECK (owner_type IN ('user', 'tenant'))
);

-- Indexes
CREATE INDEX idx_novels_owner ON catalog.novels(owner_type, owner_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_created_by ON catalog.novels(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_status ON catalog.novels(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_original_language ON catalog.novels(original_language) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_slug ON catalog.novels(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_novels_synopsis ON catalog.novels USING GIN(synopsis);
CREATE INDEX idx_novels_metadata ON catalog.novels USING GIN(metadata);
```

**Note:** `owner_id` có thể reference đến `identify.users.id` HOẶC `identify.tenants.id` tùy theo `owner_type`

#### `catalog.volumes`

```sql
CREATE TABLE catalog.volumes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    volume_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,
    description TEXT,

    cover_image_url VARCHAR(1000),

    -- Statistics (auto-updated)
    chapter_count INTEGER NOT NULL DEFAULT 0,
    word_count BIGINT NOT NULL DEFAULT 0,

    -- Ordering
    display_order INTEGER NOT NULL,
    is_published BOOLEAN NOT NULL DEFAULT FALSE,

    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(novel_id, volume_number),
    UNIQUE(novel_id, slug)
);

CREATE INDEX idx_volumes_novel_id ON catalog.volumes(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_volumes_display_order ON catalog.volumes(novel_id, display_order);
```

#### `catalog.chapters`

```sql
CREATE TABLE catalog.chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    volume_id UUID REFERENCES catalog.volumes(id) ON DELETE SET NULL,

    chapter_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,

    -- Content in ORIGINAL language (JSONB)
    content JSONB NOT NULL,

    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,

    -- Access control
    is_free BOOLEAN NOT NULL DEFAULT TRUE,
    price DECIMAL(10,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'VND',

    status catalog.chapter_status NOT NULL DEFAULT 'draft',

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,

    display_order INTEGER NOT NULL,

    -- Author notes (JSONB)
    author_notes JSONB,

    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(novel_id, chapter_number),
    UNIQUE(novel_id, slug)
);

CREATE INDEX idx_chapters_novel_id ON catalog.chapters(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_volume_id ON catalog.chapters(volume_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_status ON catalog.chapters(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_chapters_content ON catalog.chapters USING GIN(content);
```

---

### 2. Ownership System

#### `catalog.ownership_transfers`

```sql
CREATE TABLE catalog.ownership_transfers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    -- From owner
    from_owner_type VARCHAR(20) NOT NULL,
    from_owner_id UUID NOT NULL,

    -- To owner
    to_owner_type VARCHAR(20) NOT NULL,
    to_owner_id UUID NOT NULL,

    -- Transfer metadata
    initiated_by UUID NOT NULL REFERENCES identify.users(id),
    approved_by UUID REFERENCES identify.users(id),
    rejected_by UUID REFERENCES identify.users(id),

    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, cancelled

    reason TEXT,
    notes TEXT,
    admin_notes TEXT, -- Notes từ admin khi approve/reject

    -- Dates
    transferred_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT transfers_status_check CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    CONSTRAINT transfers_owner_type_check CHECK (
        from_owner_type IN ('user', 'tenant') AND to_owner_type IN ('user', 'tenant')
    )
);

CREATE INDEX idx_ownership_transfers_novel_id ON catalog.ownership_transfers(novel_id);
CREATE INDEX idx_ownership_transfers_status ON catalog.ownership_transfers(status);
CREATE INDEX idx_ownership_transfers_to_owner ON catalog.ownership_transfers(to_owner_type, to_owner_id) WHERE status = 'pending';
```

**Workflow:**
```
1. User/Tenant initiate transfer → Create record với status='pending'
2. If to_owner_type='tenant' → Tenant admin phải approve
3. If to_owner_type='user' → User phải accept
4. Admin approve → Update status='approved', transferred_at=NOW()
5. Update novels.owner_type và owner_id
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

    -- Primary language team focuses on
    primary_language VARCHAR(10) NOT NULL,

    -- Supported languages (JSONB array)
    supported_languages JSONB NOT NULL DEFAULT '[]', -- ["vi", "en", "zh"]

    -- Contact info
    website VARCHAR(500),
    discord_url VARCHAR(500),
    facebook_url VARCHAR(500),
    email VARCHAR(200),

    -- Statistics (auto-updated)
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

    role VARCHAR(50) NOT NULL DEFAULT 'translator', -- leader, translator, editor, proofreader, qc

    -- Permissions (JSONB)
    permissions JSONB DEFAULT '{}', -- {"can_approve": true, "can_publish": true}

    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,

    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX idx_team_members_user_id ON catalog.team_members(user_id);
CREATE INDEX idx_team_members_role ON catalog.team_members(team_id, role);
```

#### `catalog.novel_team_assignments`

```sql
CREATE TABLE catalog.novel_team_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES catalog.translation_teams(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    -- Assignment type
    assignment_type VARCHAR(20) NOT NULL DEFAULT 'collaborative', -- exclusive, collaborative

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    -- If exclusive, can block other teams
    is_exclusive BOOLEAN NOT NULL DEFAULT FALSE,
    exclusive_approved_by UUID REFERENCES identify.users(id), -- Novel owner approved exclusive

    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(novel_id, team_id, language)
);

CREATE INDEX idx_novel_team_assignments_novel_id ON catalog.novel_team_assignments(novel_id);
CREATE INDEX idx_novel_team_assignments_team_id ON catalog.novel_team_assignments(team_id);
CREATE INDEX idx_novel_team_assignments_language ON catalog.novel_team_assignments(language);
CREATE INDEX idx_novel_team_assignments_exclusive ON catalog.novel_team_assignments(novel_id, language) WHERE is_exclusive = TRUE;
```

---

### 4. Exclusive Rights & Reports

#### `catalog.exclusive_translation_reports`

```sql
CREATE TABLE catalog.exclusive_translation_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    -- Reporter (usually novel owner)
    reported_by UUID NOT NULL REFERENCES identify.users(id),

    -- Team being reported
    reported_team_id UUID NOT NULL REFERENCES catalog.translation_teams(id) ON DELETE CASCADE,

    -- Request details
    reason TEXT NOT NULL,
    evidence TEXT, -- Links, screenshots, etc.

    -- Review
    reviewed_by UUID REFERENCES identify.users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, approved, rejected

    -- Action taken
    action_taken VARCHAR(50), -- team_blocked, warning_issued, dismissed

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT reports_status_check CHECK (status IN ('pending', 'approved', 'rejected'))
);

CREATE INDEX idx_exclusive_reports_novel_id ON catalog.exclusive_translation_reports(novel_id);
CREATE INDEX idx_exclusive_reports_team_id ON catalog.exclusive_translation_reports(reported_team_id);
CREATE INDEX idx_exclusive_reports_status ON catalog.exclusive_translation_reports(status);
```

**Workflow:**
```
1. Novel owner discover unauthorized team translating
2. Owner submit report với evidence
3. Admin review
4. If approved:
   - Mark team assignment as blocked
   - Delete unauthorized translations (optional)
   - Issue warning to team
5. Team có thể appeal
```

---

### 5. Synopsis Translations (SEPARATE)

#### `catalog.novel_synopsis_translations`

```sql
CREATE TABLE catalog.novel_synopsis_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    -- Synopsis content (JSONB)
    synopsis JSONB NOT NULL,

    -- Translator info
    translator_id UUID REFERENCES catalog.translators(id) ON DELETE SET NULL,
    translation_team_id UUID REFERENCES catalog.translation_teams(id) ON DELETE SET NULL,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Quality
    word_count INTEGER NOT NULL DEFAULT 0,

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER NOT NULL DEFAULT 0,

    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(novel_id, language)
);

CREATE INDEX idx_synopsis_translations_novel_id ON catalog.novel_synopsis_translations(novel_id);
CREATE INDEX idx_synopsis_translations_language ON catalog.novel_synopsis_translations(language);
CREATE INDEX idx_synopsis_translations_translator_id ON catalog.novel_synopsis_translations(translator_id);
CREATE INDEX idx_synopsis_translations_team_id ON catalog.novel_synopsis_translations(translation_team_id);
CREATE INDEX idx_synopsis_translations_synopsis ON catalog.novel_synopsis_translations USING GIN(synopsis);
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

    -- Review workflow
    status catalog.translation_status NOT NULL DEFAULT 'pending_review',

    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,

    -- Link to official if approved
    official_translation_id UUID REFERENCES catalog.novel_synopsis_translations(id) ON DELETE SET NULL,

    -- Credits & community feedback
    credit_points INTEGER NOT NULL DEFAULT 0,
    is_credited BOOLEAN NOT NULL DEFAULT FALSE,

    upvote_count INTEGER NOT NULL DEFAULT 0,
    downvote_count INTEGER NOT NULL DEFAULT 0,

    word_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_synopsis_contributions_novel_id ON catalog.synopsis_translation_contributions(novel_id);
CREATE INDEX idx_synopsis_contributions_contributor_id ON catalog.synopsis_translation_contributions(contributor_id);
CREATE INDEX idx_synopsis_contributions_status ON catalog.synopsis_translation_contributions(status);
CREATE INDEX idx_synopsis_contributions_language ON catalog.synopsis_translation_contributions(language);
```

---

### 6. Chapter Translations (SEPARATE)

#### `catalog.chapter_translations`

```sql
CREATE TABLE catalog.chapter_translations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    -- Content (JSONB)
    title VARCHAR(500) NOT NULL,
    content JSONB NOT NULL,
    translator_notes JSONB,

    -- Translator info
    translator_id UUID REFERENCES catalog.translators(id) ON DELETE SET NULL,
    translation_team_id UUID REFERENCES catalog.translation_teams(id) ON DELETE SET NULL,

    -- Version tracking
    version INTEGER NOT NULL DEFAULT 1,

    -- Status
    status catalog.translation_status NOT NULL DEFAULT 'draft',

    -- Quality metrics
    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,

    -- Statistics
    view_count BIGINT NOT NULL DEFAULT 0,
    like_count INTEGER NOT NULL DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER NOT NULL DEFAULT 0,

    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(chapter_id, language)
);

CREATE INDEX idx_chapter_translations_chapter_id ON catalog.chapter_translations(chapter_id);
CREATE INDEX idx_chapter_translations_language ON catalog.chapter_translations(language);
CREATE INDEX idx_chapter_translations_translator_id ON catalog.chapter_translations(translator_id);
CREATE INDEX idx_chapter_translations_team_id ON catalog.chapter_translations(translation_team_id);
CREATE INDEX idx_chapter_translations_status ON catalog.chapter_translations(status);
CREATE INDEX idx_chapter_translations_content ON catalog.chapter_translations USING GIN(content);
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

    -- Review workflow
    status catalog.translation_status NOT NULL DEFAULT 'pending_review',

    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,

    -- Link to official if approved
    official_translation_id UUID REFERENCES catalog.chapter_translations(id) ON DELETE SET NULL,

    -- Credits & community feedback
    credit_points INTEGER NOT NULL DEFAULT 0,
    is_credited BOOLEAN NOT NULL DEFAULT FALSE,

    upvote_count INTEGER NOT NULL DEFAULT 0,
    downvote_count INTEGER NOT NULL DEFAULT 0,

    word_count INTEGER NOT NULL DEFAULT 0,
    character_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_translation_contributions_chapter_id ON catalog.translation_contributions(chapter_id);
CREATE INDEX idx_translation_contributions_contributor_id ON catalog.translation_contributions(contributor_id);
CREATE INDEX idx_translation_contributions_status ON catalog.translation_contributions(status);
CREATE INDEX idx_translation_contributions_language ON catalog.translation_contributions(language);
```

#### `catalog.translation_history`

```sql
CREATE TABLE catalog.translation_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    translation_id UUID NOT NULL REFERENCES catalog.chapter_translations(id) ON DELETE CASCADE,

    version INTEGER NOT NULL,

    -- Snapshot
    title VARCHAR(500) NOT NULL,
    content JSONB NOT NULL,

    changed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    change_description TEXT,

    word_count INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_translation_history_translation_id ON catalog.translation_history(translation_id);
CREATE INDEX idx_translation_history_version ON catalog.translation_history(translation_id, version DESC);
```

---

### 7. Language Tracking

#### `catalog.novel_languages`

```sql
CREATE TABLE catalog.novel_languages (
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    language VARCHAR(10) NOT NULL,

    is_original BOOLEAN NOT NULL DEFAULT FALSE,

    -- Translation progress
    synopsis_translated BOOLEAN NOT NULL DEFAULT FALSE,
    chapters_translated INTEGER NOT NULL DEFAULT 0,
    total_chapters INTEGER NOT NULL DEFAULT 0,

    -- Primary team/translator for this language
    primary_team_id UUID REFERENCES catalog.translation_teams(id) ON DELETE SET NULL,
    primary_translator_id UUID REFERENCES catalog.translators(id) ON DELETE SET NULL,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (novel_id, language)
);

CREATE INDEX idx_novel_languages_novel_id ON catalog.novel_languages(novel_id);
CREATE INDEX idx_novel_languages_language ON catalog.novel_languages(language);
CREATE INDEX idx_novel_languages_team_id ON catalog.novel_languages(primary_team_id);
```

---

### 8. Other Tables (Keep from previous design)

#### Genres, Authors, Artists, Translators

```sql
-- catalog.genres (hierarchical)
-- catalog.novel_genres (junction)
-- catalog.authors
-- catalog.artists
-- catalog.translators
-- catalog.novel_authors (junction)
-- catalog.novel_artists (junction)
-- catalog.novel_translators (junction)
```

*(Schema giữ nguyên như design trước)*

---

## Workflows

### 1. Novel Creation Workflow

```
User creates novel:
1. Choose owner_type (user/tenant)
2. Enter novel info
3. Select original_language
4. Enter synopsis in original language
5. System auto-creates record in novel_languages with is_original=true
6. Novel status = 'draft'

Tenant creates novel:
1. Same as user but owner_type='tenant', owner_id=tenant.id
2. created_by = user who initiated
```

### 2. Ownership Transfer Workflow

```
Transfer User → Tenant:
1. User initiate transfer to tenant X
2. Create ownership_transfer (status=pending)
3. Tenant admin of X receives notification
4. Admin approve/reject
5. If approved:
   - Update novels.owner_type = 'tenant'
   - Update novels.owner_id = tenant_id
   - Update transfer.status = 'approved'
   - Update transfer.transferred_at = NOW()

Transfer Tenant → User:
1. Tenant admin initiate transfer to user Y
2. Create ownership_transfer (status=pending)
3. User Y receives notification
4. User Y accept/reject
5. If accepted:
   - Update novels.owner_type = 'user'
   - Update novels.owner_id = user_id
   - Update transfer status
```

### 3. Synopsis Translation Workflow

```
Official Synopsis Translation:
1. Translator/Team select novel
2. Select target language
3. Translate synopsis
4. Create novel_synopsis_translations (status=draft)
5. When ready, publish (status=published)
6. Auto-update novel_languages.synopsis_translated = true

Community Synopsis Contribution:
1. User select novel to contribute
2. Select target language
3. Translate synopsis
4. Submit → Create synopsis_translation_contributions (status=pending_review)
5. Moderator review
6. If approved:
   - Update status=approved
   - Credit points awarded
   - Can merge to official translation
```

### 4. Chapter Translation Workflow

```
Official Chapter Translation:
1. Team/Translator assign to novel + language
2. For each new chapter:
   - Translate content
   - Create chapter_translations (status=draft)
   - When ready, publish (status=published)
3. Auto-update novel_languages.chapters_translated++

Community Chapter Contribution:
1. User select chapter
2. Select target language
3. Choose contribution_type (new_translation/improvement/correction)
4. Submit → Create translation_contributions (status=pending_review)
5. Review workflow
6. If approved → Merge to official translation
```

### 5. Team Assignment Workflow

```
Team wants to translate novel:
1. Team leader navigate to novel
2. Click "Request Translation Assignment"
3. Select language
4. Submit → Create novel_team_assignments (is_active=true)
5. Team can start translating

Multiple teams same language:
1. Team A already translating novel X to Vietnamese
2. Team B also wants to translate to Vietnamese
3. Team B can also create assignment (collaborative by default)
4. Both teams can publish translations

Exclusive Rights Request:
1. Novel owner wants only Team A to translate
2. Owner report Team B via exclusive_translation_reports
3. Admin review
4. If approved:
   - Team A: is_exclusive=true
   - Team B: is_active=false (blocked)
5. Team B notified and translations may be removed
```

---

## Permissions & Security

### Ownership Permissions

```
Novel Owner (User/Tenant):
- Full CRUD on novel
- Manage volumes & chapters
- Approve/reject ownership transfers
- Report unauthorized translations
- Grant exclusive translation rights
- Manage official translations

Tenant Admin:
- All tenant-owned novel permissions
- Approve incoming transfer requests
- Initiate outgoing transfers
- Manage team assignments
```

### Team Permissions

```
Team Leader:
- Assign members to novel projects
- Approve member translations
- Publish official translations
- Request exclusive rights

Team Translator:
- Create draft translations
- Submit for review
- Edit own translations

Team Editor:
- Review translations
- Approve for publishing
- Edit any team translation

Community Contributor:
- Submit contributions
- Edit own contributions (before review)
- Vote on other contributions
```

---

## Implementation Plan

### Phase 1: Core Structure (Priority 1)

**Migration 001:**
```
- novels (with ownership)
- volumes
- chapters
- Enum types
- Basic triggers
```

### Phase 2: Ownership System (Priority 1)

**Migration 002:**
```
- ownership_transfers
- Update novels with owner fields
- Transfer workflow triggers
```

### Phase 3: Teams & Collaboration (Priority 2)

**Migration 003:**
```
- translation_teams
- team_members
- novel_team_assignments
- exclusive_translation_reports
```

### Phase 4: Synopsis Translations (Priority 2)

**Migration 004:**
```
- novel_synopsis_translations
- synopsis_translation_contributions
```

### Phase 5: Chapter Translations (Priority 3)

**Migration 005:**
```
- chapter_translations
- translation_contributions
- translation_history
- Version control triggers
```

### Phase 6: Supporting Tables (Priority 3)

**Migration 006:**
```
- genres
- novel_genres
- authors
- artists
- translators
- Junction tables
- novel_languages
```

---

## Next Steps

1. **Review Design**: Kiểm tra lại design với stakeholders
2. **Approve Design**: Confirm tất cả requirements đã cover
3. **Create Migrations**: Implement từng phase
4. **Create Domain Models**: Go domain models cho từng entity
5. **Create Repositories**: Data access layer
6. **Create Services**: Business logic layer
7. **Create API**: HTTP handlers và routes
8. **Testing**: Unit tests và integration tests

---

**Status:** ✅ Design Complete - Ready for Review
**Next:** Awaiting approval to proceed with implementation

