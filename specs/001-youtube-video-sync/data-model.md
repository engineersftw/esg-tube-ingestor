# Data Model: YouTube Video Sync

**Feature**: 001-youtube-video-sync
**Date**: 2025-10-20
**Purpose**: Define entity mappings between YouTube API responses and existing PostgreSQL schema

## Overview

This document maps YouTube Data API entities to the existing database schema. **Critical**: The sync tool will ONLY write to `episodes`, `playlists`, and `playlist_items` tables. The `presenters`, `organizations`, and their junction tables are manually maintained.

## Entity Mappings

### 1. YouTube Video → Episode (public.episodes)

**Source**: YouTube Data API `videos.list` endpoint with `snippet` and `statistics` parts

**Mapping**:

| YouTube API Field | Database Column | Type | Transformation | Notes |
|-------------------|-----------------|------|----------------|-------|
| `id` | `video_id` | VARCHAR | Direct | YouTube video ID (e.g., "dQw4w9WgXcQ") |
| `snippet.title` | `title` | VARCHAR | Direct | Video title |
| `snippet.description` | `description` | TEXT | Direct | Full video description |
| `snippet.publishedAt` | `published_at` | TIMESTAMP | Parse RFC3339 → timestamp | Upload date/time |
| `snippet.thumbnails.default.url` | `image1` | VARCHAR | Direct | Default thumbnail (120x90) |
| `snippet.thumbnails.medium.url` | `image2` | VARCHAR | Direct | Medium thumbnail (320x180) |
| `snippet.thumbnails.high.url` | `image3` | VARCHAR | Direct | High res thumbnail (480x360) |
| `statistics.viewCount` | `view_count` | INTEGER | Parse string → int | Current view count |
| N/A | `video_site` | INTEGER | Constant: 1 | 1 = YouTube (hardcoded) |
| N/A | `active` | BOOLEAN | Default: true | Set to false if video deleted/private |
| N/A | `sort_order` | INTEGER | NULL or user-specified | For manual ordering |
| N/A | `created_at` | TIMESTAMP | NOW() | Record creation time |
| N/A | `updated_at` | TIMESTAMP | NOW() | Record update time |

**Validation Rules**:
- `video_id` must be unique (primary key)
- `video_id` must match YouTube ID format (11 characters, alphanumeric + underscore/hyphen)
- `title` must not be empty
- `view_count` must be >= 0
- `published_at` must be valid timestamp

**Upsert Strategy** (FR-007):
```sql
INSERT INTO episodes (
    video_id, title, description, published_at,
    image1, image2, image3, view_count,
    video_site, active, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 1, true, NOW(), NOW())
ON CONFLICT (video_id) DO UPDATE SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    published_at = EXCLUDED.published_at,
    image1 = EXCLUDED.image1,
    image2 = EXCLUDED.image2,
    image3 = EXCLUDED.image3,
    view_count = EXCLUDED.view_count,
    active = EXCLUDED.active,
    updated_at = NOW();
```

**Not Mapped from YouTube**:
- `snippet.channelTitle` (channel name): Logged but NOT stored in database
- `snippet.channelId` (channel ID): Logged but NOT stored in database
- `statistics.likeCount`: Available in API but not in current schema (ignore for now)
- `statistics.commentCount`: Available in API but not in current schema (ignore for now)

### 2. YouTube Playlist → Playlist (public.playlists)

**Source**: YouTube Data API `playlists.list` endpoint with `snippet` and `contentDetails` parts

**Mapping**:

| YouTube API Field | Database Column | Type | Transformation | Notes |
|-------------------|-----------------|------|----------------|-------|
| `id` | `playlist_id` | VARCHAR | Direct | YouTube playlist ID |
| `snippet.title` | `name` | VARCHAR | Direct | Playlist name |
| `snippet.description` | `description` | TEXT | Direct | Playlist description |
| `snippet.publishedAt` | `publish_date` | DATE | Parse RFC3339 → date | Playlist creation date |
| `snippet.thumbnails.high.url` | `image` | VARCHAR | Direct | Playlist thumbnail |
| N/A | `website` | VARCHAR | NULL | Not available from YouTube API |
| N/A | `hashtag` | VARCHAR | NULL | Not available from YouTube API |
| N/A | `playlist_category_id` | INTEGER | NULL | Manual categorization required |
| `snippet.title` | `slug` | VARCHAR | Generate slug | Slugify title for URL-friendly format |
| N/A | `active` | BOOLEAN | Default: true | Can be marked inactive manually |
| N/A | `created_at` | TIMESTAMP | NOW() | Record creation time |
| N/A | `updated_at` | TIMESTAMP | NOW() | Record update time |

**Slug Generation**:
- Convert title to lowercase
- Replace spaces and special chars with hyphens
- Remove duplicate hyphens
- Trim leading/trailing hyphens
- Example: "My Awesome Playlist!" → "my-awesome-playlist"

**Upsert Strategy**:
```sql
INSERT INTO playlists (
    playlist_id, name, description, publish_date,
    image, slug, active, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
ON CONFLICT (playlist_id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    publish_date = EXCLUDED.publish_date,
    image = EXCLUDED.image,
    slug = EXCLUDED.slug,
    updated_at = NOW();
```

### 3. YouTube Playlist Items → PlaylistItem (public.playlist_items)

**Source**: YouTube Data API `playlistItems.list` endpoint

**Mapping**:

| YouTube API Field | Database Column | Type | Transformation | Notes |
|-------------------|-----------------|------|----------------|-------|
| N/A (from playlist) | `playlist_id` | INTEGER | Lookup from playlists table | Foreign key to playlists.id |
| `snippet.resourceId.videoId` | `episode_id` | INTEGER | Lookup from episodes table | Foreign key to episodes.id |
| `snippet.position` | `sort_order` | INTEGER | Direct | Order of video in playlist (0-indexed) |
| N/A | `created_at` | TIMESTAMP | NOW() | Record creation time |
| N/A | `updated_at` | TIMESTAMP | NOW() | Record update time |

**Lookup Strategy**:
1. First, ensure playlist exists in `playlists` table (create/update if needed)
2. For each playlist item:
   - Fetch video metadata if not already in `episodes` table
   - Lookup `playlists.id` using `playlist_id` (YouTube ID)
   - Lookup `episodes.id` using `video_id` (YouTube ID)
   - Insert/update playlist_items with internal IDs

**Upsert Strategy**:
```sql
INSERT INTO playlist_items (playlist_id, episode_id, sort_order, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
ON CONFLICT (playlist_id, episode_id) DO UPDATE SET
    sort_order = EXCLUDED.sort_order,
    updated_at = NOW();
```

**Constraint**: `(playlist_id, episode_id)` should be unique to avoid duplicate entries

### 4. Sync Job Tracking (Optional - Future Enhancement)

**Note**: Not in current database schema. For v1, use structured logging to track sync operations.

**Logged Information** (JSON format):
- Job ID (UUID)
- Job Type (video/channel/playlist)
- Target (YouTube ID or URL)
- Start Time
- End Time
- Status (success/partial/failed)
- Items Processed
- Success Count
- Failure Count
- Error Messages

**Example Log Entry**:
```json
{
  "job_id": "a1b2c3d4-e5f6-4789-a0b1-c2d3e4f5g6h7",
  "job_type": "channel_sync",
  "target": "UCXuqSBlHAE6Xw-yeJA0Tunw",
  "start_time": "2025-10-20T14:30:00Z",
  "end_time": "2025-10-20T14:35:23Z",
  "status": "success",
  "items_processed": 150,
  "success_count": 148,
  "failure_count": 2,
  "errors": [
    {"video_id": "abc123", "error": "Video unavailable (deleted)"},
    {"video_id": "def456", "error": "Video unavailable (private)"}
  ]
}
```

## Tables NOT Modified by Sync Tool

### Presenters (public.presenters)
**Manually Maintained** - sync tool does NOT create or modify records

Fields: `id`, `name`, `biography`, `twitter`, `email`, `website`, `avatar_url`, `byline`, `active`

**Rationale**: Presenters represent specific individuals (speakers, hosts) who are manually curated. YouTube channels do not directly map to presenters (one channel may have multiple presenters, or one presenter may appear across multiple channels).

### Video Presenters (public.video_presenters)
**Manually Maintained** - junction table linking episodes to presenters

Fields: `id`, `episode_id` (FK), `presenter_id` (FK)

**Rationale**: Administrators manually associate videos with presenters after syncing. The sync tool cannot automatically determine which presenter(s) appear in a video.

### Organizations (public.organizations)
**Manually Maintained** - sync tool does NOT create or modify records

Fields: `id`, `title`, `description`, `website`, `twitter`, `contact_person`, `image`, `slug`, `active`

**Rationale**: Organizations are manually managed entities representing companies, institutions, or groups. YouTube channels may or may not represent organizations.

### Video Organizations (public.video_organizations)
**Manually Maintained** - junction table linking episodes to organizations

Fields: `id`, `episode_id` (FK), `organization_id` (FK)

**Rationale**: Administrators manually tag videos with relevant organizations after syncing.

## Data Integrity Constraints

### Foreign Key Constraints (Existing Schema)

1. **playlist_items.playlist_id** → playlists.id (CASCADE on delete)
2. **playlist_items.episode_id** → episodes.id (CASCADE on delete)
3. **video_presenters.episode_id** → episodes.id (CASCADE on delete)
4. **video_presenters.presenter_id** → presenters.id (CASCADE on delete)
5. **video_organizations.episode_id** → episodes.id (CASCADE on delete)
6. **video_organizations.organization_id** → organizations.id (CASCADE on delete)

### Sync Tool Responsibilities

**Must Ensure**:
- Valid foreign key references when creating playlist_items
- No orphaned records (always create parent before child)
- Proper transaction handling for atomicity

**Must NOT**:
- Modify presenters table
- Modify organizations table
- Modify video_presenters table
- Modify video_organizations table
- Modify any auth.* or storage.* tables

## State Transitions

### Episode Active Status

```
[New Video] → active = true
[Video Sync Success] → active = true (remains active)
[Video Deleted/Private] → active = false (marked unavailable)
[Manual Reactivation] → active = true (admin can toggle)
```

**Handling Unavailable Videos** (FR-013):
- On sync, if YouTube API returns 404 or "video unavailable", set `active = false`
- Do NOT delete the record (preserve historical data)
- Log the unavailability for admin review
- Continue processing other videos (don't fail entire sync)

### Playlist Active Status

```
[New Playlist] → active = true
[Playlist Sync Success] → active = true (remains active)
[Playlist Deleted] → active = false (manual marking by admin)
```

## Validation Rules Summary

### Episode Validation
- [ ] `video_id` matches YouTube ID format (11 chars, alphanumeric + `-_`)
- [ ] `title` is not empty and <= 255 chars
- [ ] `view_count` >= 0
- [ ] `published_at` is valid timestamp
- [ ] `video_site` = 1 (YouTube constant)
- [ ] Thumbnail URLs are valid HTTP(S) URLs or NULL

### Playlist Validation
- [ ] `playlist_id` matches YouTube playlist ID format (starts with "PL" or "UU")
- [ ] `name` is not empty and <= 255 chars
- [ ] `slug` is URL-safe (lowercase, hyphens, no special chars)
- [ ] `publish_date` is valid date

### Playlist Item Validation
- [ ] `playlist_id` exists in playlists table
- [ ] `episode_id` exists in episodes table
- [ ] `sort_order` >= 0
- [ ] No duplicate (playlist_id, episode_id) pairs

## Error Handling Strategy

### YouTube API Errors

| Error Code | Handling | Impact |
|------------|----------|--------|
| 404 (Not Found) | Mark episode as `active = false` | Continue sync |
| 403 (Forbidden/Private) | Mark episode as `active = false` | Continue sync |
| 429 (Rate Limit) | Exponential backoff, retry | Pause sync, resume |
| 400 (Bad Request) | Log error, skip item | Continue sync |
| 500 (Server Error) | Retry with backoff (max 3 attempts) | Continue or fail |

### Database Errors

| Error Type | Handling | Impact |
|------------|----------|--------|
| Connection failure | Retry with backoff (max 5 attempts) | Fail sync |
| Constraint violation | Log error, skip item | Continue sync |
| Transaction deadlock | Retry transaction (max 3 attempts) | Fail sync |
| Foreign key violation | Log error, skip item (data inconsistency) | Continue sync |

## Performance Considerations

### Batch Operations
- Use database transactions for bulk inserts
- Batch playlist_items inserts (e.g., 100 items per transaction)
- Commit transactions regularly to avoid long-running locks

### Indexing (Existing Schema)
- Ensure `episodes.video_id` has unique index
- Ensure `playlists.playlist_id` has unique index
- Ensure `playlist_items (playlist_id, episode_id)` has unique constraint

### Caching Strategy (Future Enhancement)
- Cache YouTube API responses for duplicate requests within same sync
- Cache database ID lookups for videos/playlists during bulk operations

## Next Steps

With data model defined, proceed to:
1. Create CLI command contracts (`contracts/cli-commands.md`)
2. Generate quickstart manual testing guide (`quickstart.md`)
3. Update agent context with technology stack
