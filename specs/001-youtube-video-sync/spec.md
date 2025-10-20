# Feature Specification: YouTube Video Sync

**Feature Branch**: `001-youtube-video-sync`
**Created**: 2025-10-20
**Status**: Draft
**Input**: User description: "Build a CLI tool that will sync YouTube videos to a Database"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manual Single Video Sync (Priority: P1)

An administrator wants to sync a specific YouTube video's metadata to the database to begin tracking it. They run a CLI command with a video URL or ID, and the system fetches and stores the video information.

**Why this priority**: This is the minimal viable product - the core capability that all other features depend on. It proves the fundamental sync mechanism works and delivers immediate value for basic use cases.

**Independent Test**: Can be fully tested by providing a valid YouTube video URL/ID via CLI and verifying the video metadata appears in the database. Delivers standalone value for manual video tracking.

**Acceptance Scenarios**:

1. **Given** a valid YouTube video URL, **When** the user runs the sync command with that URL, **Then** the video metadata (title, description, duration, upload date, channel info) is stored in the database
2. **Given** a valid YouTube video ID, **When** the user runs the sync command with that ID, **Then** the video metadata is fetched and stored
3. **Given** an invalid YouTube URL, **When** the user runs the sync command, **Then** a clear error message is displayed and no database entry is created
4. **Given** a video that already exists in the database, **When** the user runs the sync command for that video, **Then** the existing record is updated with current metadata

---

### User Story 2 - Bulk Video Sync from Channel (Priority: P2)

An administrator wants to sync all videos from a specific YouTube channel to build a comprehensive video database. They provide a channel URL or ID, and the system fetches metadata for all public videos from that channel. The channel information itself is stored as metadata with each video but does NOT create or update presenter records.

**Why this priority**: Significantly reduces manual effort for bulk operations and is a common use case for content aggregation. Builds on P1's single-video capability.

**Independent Test**: Can be tested by providing a YouTube channel identifier and verifying all channel videos appear in the database. Delivers value for content managers tracking specific channels.

**Acceptance Scenarios**:

1. **Given** a valid YouTube channel URL, **When** the user runs the bulk sync command, **Then** all public videos from that channel are fetched and stored in the database
2. **Given** a channel with 100+ videos, **When** the bulk sync runs, **Then** the system handles pagination correctly and stores all videos
3. **Given** a channel sync in progress, **When** the user requests status, **Then** progress information is displayed (e.g., "Syncing 45/150 videos")
4. **Given** a previously synced channel, **When** the bulk sync runs again, **Then** only new videos are added and existing videos are updated
5. **Given** videos are synced from a channel, **When** reviewing the database, **Then** the channel name and ID are stored with each video but NO presenter records are created or modified

---

### User Story 3 - Scheduled Automatic Sync (Priority: P3)

An administrator wants videos to stay up-to-date automatically without manual intervention. They configure a sync schedule (e.g., daily at 2 AM), and the system automatically refreshes tracked videos and channels.

**Why this priority**: Provides automation for ongoing maintenance but is not essential for initial value delivery. Users can manually re-run syncs until this is implemented.

**Independent Test**: Can be tested by configuring a schedule, waiting for the scheduled time, and verifying the database was updated automatically. Delivers value for "set and forget" operations.

**Acceptance Scenarios**:

1. **Given** a configured sync schedule, **When** the scheduled time arrives, **Then** all tracked videos and channels are automatically synced
2. **Given** a scheduled sync completes, **When** checking the sync logs, **Then** the sync results (success/failure counts, timestamp) are recorded
3. **Given** a scheduled sync fails, **When** reviewing the error logs, **Then** the failure details (timestamp, error message, affected items) are recorded in log files

---

### User Story 4 - Playlist Sync (Priority: P3)

An administrator wants to sync all videos from a YouTube playlist to track curated collections. They provide a playlist URL or ID, and the system fetches metadata for all videos in that playlist.

**Why this priority**: Useful for tracking themed collections but less common than channel-based syncing. Can be deferred if resources are limited.

**Independent Test**: Can be tested by providing a playlist identifier and verifying all playlist videos appear in the database. Delivers value for curated content tracking.

**Acceptance Scenarios**:

1. **Given** a valid YouTube playlist URL, **When** the user runs the playlist sync command, **Then** all videos in the playlist are fetched and stored
2. **Given** a playlist containing videos from multiple channels, **When** the sync runs, **Then** videos are correctly associated with their respective channels
3. **Given** a playlist with duplicate videos, **When** the sync runs, **Then** each unique video is stored only once in the database

---

### Edge Cases

- What happens when YouTube API rate limits are exceeded during a bulk sync?
- How does the system handle private or deleted videos that were previously synced?
- What happens when a video's metadata changes significantly (e.g., title/description edit)?
- How does the system handle videos that are geo-restricted or age-restricted?
- What happens when the database connection fails mid-sync?
- How does the system handle malformed or unusual YouTube URLs (e.g., shortened URLs, embed URLs)?
- What happens when a channel is deleted or made private after being synced?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept YouTube video URLs (full and shortened formats) and video IDs as input
- **FR-002**: System MUST accept YouTube channel URLs (various formats) and channel IDs for bulk operations
- **FR-003**: System MUST accept YouTube playlist URLs and playlist IDs for collection syncing
- **FR-004**: System MUST fetch video metadata including: title, description, duration, upload date, thumbnail URLs, view count, like count (note: channel name and ID are fetched but stored separately for reference only)
- **FR-005**: System MUST fetch channel metadata for logging and reference purposes, but NOT persist it to the presenters table (presenters are manually maintained)
- **FR-006**: System MUST persist video and channel metadata to a PostgreSQL database
- **FR-007**: System MUST update existing records when syncing previously tracked videos (upsert operation)
- **FR-008**: System MUST handle YouTube API rate limits gracefully with appropriate retry logic and backoff
- **FR-009**: System MUST provide clear error messages for invalid inputs (malformed URLs, non-existent videos)
- **FR-010**: System MUST log all sync operations with timestamps and results (success/failure counts)
- **FR-011**: System MUST support authentication with YouTube Data API using API keys for accessing public content only
- **FR-012**: Users MUST be able to query sync status for in-progress bulk operations
- **FR-013**: System MUST handle videos that are deleted or made private (mark as unavailable rather than fail entire sync)
- **FR-014**: System MUST respect YouTube's Terms of Service and API usage quotas
- **FR-015**: CLI MUST follow POSIX conventions (exit codes, standard output/error streams)
- **FR-016**: System MUST use the existing PostgreSQL database schema and map YouTube entities to existing tables (episodes, playlists, playlist_items, presenters, organizations)
- **FR-017**: System MUST map YouTube videos to the `episodes` table with video_id storing the YouTube video ID
- **FR-018**: System MUST NOT automatically create or modify records in the `presenters` table - presenters are manually maintained by administrators
- **FR-019**: System MUST map YouTube playlists to the `playlists` table with playlist_id storing the YouTube playlist ID
- **FR-020**: System MAY log channel information (name, ID) for reference but MUST NOT persist it to database tables

### Key Entities

**Note**: The system uses an existing PostgreSQL schema. The CLI tool will map YouTube data to these existing tables:

- **Episode** (existing `public.episodes` table): Represents a YouTube video with fields: video_id (YouTube ID), title, description, published_at, image1/2/3 (thumbnails), view_count, video_site (platform identifier), active (availability status), sort_order. This is the primary table that will be populated by the sync tool.
- **Playlist** (existing `public.playlists` table): Represents a YouTube playlist with fields: playlist_id (YouTube ID), name, description, publish_date, image, website, hashtag, playlist_category_id, slug, active. This will be populated when syncing playlists.
- **PlaylistItem** (existing `public.playlist_items` table): Junction table linking playlists to episodes with fields: playlist_id (FK to playlists), episode_id (FK to episodes), sort_order. This maintains the order of videos within playlists.
- **Presenter** (existing `public.presenters` table): Represents content creators/speakers. **IMPORTANT**: This table is manually maintained by administrators and will NOT be automatically populated or modified by the sync tool. YouTube channel information will NOT be stored here.
- **VideoPresenter** (existing `public.video_presenters` table): Junction table linking episodes to presenters (many-to-many relationship). Administrators manually create these associations after syncing videos.
- **Organization** (existing `public.organizations` table): Represents organizations associated with content. This table is manually maintained by administrators.
- **VideoOrganization** (existing `public.video_organizations` table): Junction table linking episodes to organizations. Administrators manually create these associations.

## Assumptions

- **Existing Database Schema**: The system will integrate with an existing PostgreSQL database schema (see backup at `docs/db_cluster-06-06-2025@08-19-26.backup`). No new tables will be created; YouTube data will be mapped to existing tables
- **Database Technology**: PostgreSQL was selected for its robust support of structured data, strong query capabilities, and industry-standard reliability for production deployments
- **Schema Mapping**: YouTube videos map to `episodes` table, YouTube playlists map to `playlists` table. YouTube channel information is logged but NOT persisted to any table.
- **Manual Content Curation**: The `presenters`, `organizations`, `video_presenters`, and `video_organizations` tables are manually maintained by administrators. The sync tool does NOT create or modify these records.
- **Authentication Scope**: API key authentication is sufficient as the system only needs to access publicly available YouTube content
- **Error Notification**: File-based logging is adequate for error tracking; administrators will monitor log files for scheduled sync failures rather than receiving active notifications

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can sync a single video in under 5 seconds for videos with standard metadata
- **SC-002**: System successfully syncs 95% of valid YouTube videos without errors
- **SC-003**: System handles bulk syncs of 1000+ videos from a channel without manual intervention
- **SC-004**: Video metadata in database reflects current YouTube state within configured sync interval
- **SC-005**: System processes bulk channel syncs at a rate of at least 100 videos per minute (accounting for API rate limits)
- **SC-006**: 100% of sync operations are logged with sufficient detail for troubleshooting
- **SC-007**: Users can successfully configure and run the tool within 10 minutes of installation (including API key setup)
