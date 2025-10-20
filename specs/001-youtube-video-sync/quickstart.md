# Quickstart Guide: YouTube Video Sync

**Feature**: 001-youtube-video-sync
**Date**: 2025-10-20
**Purpose**: Manual testing scenarios to validate the implementation before automated tests

## Prerequisites

Before testing, ensure you have:
- [ ] PostgreSQL database running with existing schema (from `docs/db_cluster-06-06-2025@08-19-26.backup`)
- [ ] YouTube Data API key (get from [Google Cloud Console](https://console.developers.google.com/))
- [ ] Go 1.21+ installed
- [ ] `youtube-sync` binary built and in PATH

## Setup

### 1. Database Setup

```bash
# Restore existing database schema (if needed)
psql -U postgres -d esg_tube < docs/db_cluster-06-06-2025@08-19-26.backup

# Verify tables exist
psql -U postgres -d esg_tube -c "\dt"
# Should show: episodes, playlists, playlist_items, presenters, organizations, etc.
```

### 2. Configuration

Create `~/.youtube-sync/config.yaml`:

```yaml
youtube:
  api_key: ${YOUTUBE_API_KEY}

database:
  host: localhost
  port: 5432
  user: postgres
  password: ${POSTGRES_PASSWORD}
  database: esg_tube
  sslmode: disable

logging:
  level: info
  format: text
```

Set environment variables:

```bash
export YOUTUBE_API_KEY="your-api-key-here"
export POSTGRES_PASSWORD="your-db-password"
```

### 3. Verify Installation

```bash
youtube-sync version
# Should show version info

youtube-sync --help
# Should show command list
```

## Test Scenarios

### Scenario 1: Sync Single Video (P1 - MVP)

**User Story**: As an administrator, I want to sync a single YouTube video to verify basic functionality.

**Test Case 1.1: Sync Public Video**

```bash
# Use a well-known public video
youtube-sync video "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
```

**Expected Result**:
- ✓ Command succeeds with exit code 0
- ✓ Console shows success message with video details
- ✓ Database query confirms record exists:
  ```sql
  SELECT video_id, title, view_count, active
  FROM episodes
  WHERE video_id = 'dQw4w9WgXcQ';
  ```
- ✓ `active` field is `true`
- ✓ `video_site` field is `1` (YouTube)
- ✓ Thumbnails (image1, image2, image3) are populated with URLs

**Acceptance Criteria**: Meets User Story 1, Scenario 1

---

**Test Case 1.2: Sync by Video ID Only**

```bash
youtube-sync video dQw4w9WgXcQ
```

**Expected Result**:
- ✓ Command accepts bare video ID (no URL required)
- ✓ Same video synced successfully

**Acceptance Criteria**: Meets User Story 1, Scenario 2

---

**Test Case 1.3: Sync Invalid Video**

```bash
youtube-sync video "https://www.youtube.com/watch?v=INVALIDID123"
```

**Expected Result**:
- ✗ Command fails with exit code 4 (API error)
- ✗ Error message explains video not found
- ✗ No database record created:
  ```sql
  SELECT * FROM episodes WHERE video_id = 'INVALIDID123';
  -- Should return 0 rows
  ```

**Acceptance Criteria**: Meets User Story 1, Scenario 3

---

**Test Case 1.4: Update Existing Video**

```bash
# Sync same video again
youtube-sync video dQw4w9WgXcQ
```

**Expected Result**:
- ✓ Command succeeds
- ✓ Database record updated (check `updated_at` timestamp):
  ```sql
  SELECT video_id, created_at, updated_at
  FROM episodes
  WHERE video_id = 'dQw4w9WgXcQ';
  -- updated_at should be more recent than created_at
  ```
- ✓ `view_count` reflects current YouTube value

**Acceptance Criteria**: Meets User Story 1, Scenario 4

---

**Test Case 1.5: Dry Run Mode**

```bash
youtube-sync video "https://www.youtube.com/watch?v=9bZkp7q19f0" --dry-run
```

**Expected Result**:
- ✓ Command succeeds
- ✓ Console shows what would be synced
- ✗ No database record created:
  ```sql
  SELECT * FROM episodes WHERE video_id = '9bZkp7q19f0';
  -- Should return 0 rows
  ```

---

**Test Case 1.6: JSON Output Mode**

```bash
youtube-sync video dQw4w9WgXcQ --json
```

**Expected Result**:
- ✓ Output is valid JSON
- ✓ JSON includes: `status`, `video` object, `operation`, `timestamp`
- ✓ Can be parsed by `jq`:
  ```bash
  youtube-sync video dQw4w9WgXcQ --json | jq '.video.title'
  ```

---

### Scenario 2: Sync Channel Videos (P2)

**User Story**: As an administrator, I want to sync all videos from a YouTube channel.

**Test Case 2.1: Sync Small Channel (< 50 videos)**

```bash
# Use a channel with known small video count
youtube-sync channel "https://www.youtube.com/@TEDEd" --limit 10
```

**Expected Result**:
- ✓ Command succeeds
- ✓ Progress indicator shows "Syncing X/10"
- ✓ Database contains videos from channel:
  ```sql
  SELECT COUNT(*) FROM episodes WHERE created_at > NOW() - INTERVAL '5 minutes';
  -- Should show ~10 new records
  ```
- ✓ All videos have `active = true`

**Acceptance Criteria**: Meets User Story 2, Scenarios 1 and 3

---

**Test Case 2.2: Sync with Pagination**

```bash
# Sync channel with 100+ videos
youtube-sync channel "UCuAXFkgsw1L7xaCfnd5JJOw" --limit 100 --verbose
```

**Expected Result**:
- ✓ Verbose logs show pagination (fetching page 1, page 2, etc.)
- ✓ All 100 videos synced
- ✓ Database count matches:
  ```sql
  SELECT COUNT(*) FROM episodes WHERE created_at > NOW() - INTERVAL '10 minutes';
  ```

**Acceptance Criteria**: Meets User Story 2, Scenario 2

---

**Test Case 2.3: Re-sync Channel (Update Existing)**

```bash
# Run same channel sync again
youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --limit 10
```

**Expected Result**:
- ✓ Command succeeds quickly (videos already exist)
- ✓ Existing videos updated (check `updated_at`)
- ✓ No duplicate records:
  ```sql
  SELECT video_id, COUNT(*) as count
  FROM episodes
  GROUP BY video_id
  HAVING COUNT(*) > 1;
  -- Should return 0 rows
  ```

**Acceptance Criteria**: Meets User Story 2, Scenario 4

---

**Test Case 2.4: Channel Info NOT Stored**

```bash
# Sync channel and verify channel data is NOT in presenters table
youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --limit 5

# Check presenters table count before and after
psql -U postgres -d esg_tube -c "SELECT COUNT(*) FROM presenters;"
```

**Expected Result**:
- ✓ Presenter count unchanged (no auto-created presenters)
- ✓ Logs may mention channel name, but NOT stored in database
- ✓ `video_presenters` table unchanged

**Acceptance Criteria**: Meets User Story 2, Scenario 5 + FR-018

---

**Test Case 2.5: Handle Unavailable Videos**

```bash
# Sync channel that contains deleted/private videos
youtube-sync channel <channel-with-unavailable-videos> --verbose
```

**Expected Result**:
- ✓ Command succeeds (doesn't fail entire sync)
- ✓ Unavailable videos marked as `active = false`:
  ```sql
  SELECT video_id, title, active
  FROM episodes
  WHERE active = false;
  ```
- ✓ Logs show "Video unavailable" for failed videos
- ✓ Success count reflects only available videos

**Acceptance Criteria**: Meets FR-013 and edge case handling

---

### Scenario 3: Sync Playlist (P3)

**User Story**: As an administrator, I want to sync a YouTube playlist with all its videos.

**Test Case 3.1: Sync Public Playlist**

```bash
youtube-sync playlist "https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"
```

**Expected Result**:
- ✓ Command succeeds
- ✓ Playlist metadata stored:
  ```sql
  SELECT playlist_id, name, description
  FROM playlists
  WHERE playlist_id = 'PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf';
  ```
- ✓ All playlist videos synced to `episodes` table
- ✓ Playlist items created:
  ```sql
  SELECT COUNT(*) FROM playlist_items
  WHERE playlist_id = (SELECT id FROM playlists WHERE playlist_id = 'PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf');
  ```
- ✓ `sort_order` matches YouTube playlist order (0, 1, 2, ...)

**Acceptance Criteria**: Meets User Story 4, Scenario 1

---

**Test Case 3.2: Playlist with Multi-Channel Videos**

```bash
# Sync playlist containing videos from different channels
youtube-sync playlist <playlist-id>
```

**Expected Result**:
- ✓ All videos synced regardless of channel
- ✓ Videos correctly associated with their channels in logs
- ✓ No presenter records auto-created

**Acceptance Criteria**: Meets User Story 4, Scenario 2

---

**Test Case 3.3: Playlist with Duplicate Videos**

```bash
# Sync playlist where same video appears multiple times
youtube-sync playlist <playlist-with-duplicates>
```

**Expected Result**:
- ✓ Each unique video stored only once in `episodes` table:
  ```sql
  SELECT video_id, COUNT(*) FROM episodes GROUP BY video_id HAVING COUNT(*) > 1;
  -- Should return 0 rows
  ```
- ✓ Multiple `playlist_items` entries may exist if video appears at different positions

**Acceptance Criteria**: Meets User Story 4, Scenario 3

---

### Scenario 4: Error Handling

**Test Case 4.1: Missing API Key**

```bash
# Unset API key
unset YOUTUBE_API_KEY

youtube-sync video dQw4w9WgXcQ
```

**Expected Result**:
- ✗ Exit code 3 (configuration error)
- ✗ Clear error message about missing API key
- ✗ Suggestion to set environment variable or config file

---

**Test Case 4.2: Database Connection Failure**

```bash
# Stop PostgreSQL or use invalid connection
export POSTGRES_HOST="invalid-host"

youtube-sync video dQw4w9WgXcQ
```

**Expected Result**:
- ✗ Exit code 5 (database error)
- ✗ Error message about connection failure
- ✗ Retry attempts visible in verbose mode

---

**Test Case 4.3: API Rate Limit**

```bash
# Trigger rate limit by syncing many videos quickly
for i in {1..200}; do
  youtube-sync video <different-video-ids>
done
```

**Expected Result**:
- ✓ Commands pause and retry when hitting rate limit
- ✓ Exponential backoff visible in logs
- ✓ Eventually succeeds after backoff

---

### Scenario 5: Performance Validation

**Test Case 5.1: Single Video Sync Speed**

```bash
time youtube-sync video dQw4w9WgXcQ
```

**Expected Result**:
- ✓ Completes in < 5 seconds (SC-001)

---

**Test Case 5.2: Bulk Sync Throughput**

```bash
# Sync 100 videos from channel
time youtube-sync channel <channel-id> --limit 100
```

**Expected Result**:
- ✓ Processes at ~100 videos/minute (SC-005)
- ✓ Duration ≤ 60 seconds for 100 videos

---

### Scenario 6: Logging Validation

**Test Case 6.1: Structured Logging**

```bash
youtube-sync video dQw4w9WgXcQ --verbose 2>&1 | grep -E "(INFO|ERROR)"
```

**Expected Result**:
- ✓ Logs include timestamps
- ✓ Logs include log levels (INFO, ERROR, WARN)
- ✓ Logs include video IDs and relevant context

---

**Test Case 6.2: JSON Logging**

```bash
# Enable JSON logging in config
LOG_FORMAT=json youtube-sync video dQw4w9WgXcQ --verbose
```

**Expected Result**:
- ✓ All log lines are valid JSON
- ✓ JSON includes: timestamp, level, message, context fields

---

## Post-Testing Validation

After completing test scenarios, verify:

### Database Integrity

```sql
-- No duplicate video_ids
SELECT video_id, COUNT(*)
FROM episodes
GROUP BY video_id
HAVING COUNT(*) > 1;
-- Should return 0 rows

-- All videos have required fields
SELECT COUNT(*)
FROM episodes
WHERE video_id IS NULL OR title IS NULL OR title = '';
-- Should return 0

-- All videos have video_site = 1 (YouTube)
SELECT COUNT(*)
FROM episodes
WHERE video_site != 1;
-- Should return 0

-- Presenters table unchanged (manual only)
SELECT COUNT(*) FROM presenters;
-- Compare to count before testing - should be identical

-- Organizations table unchanged (manual only)
SELECT COUNT(*) FROM organizations;
-- Compare to count before testing - should be identical

-- Valid foreign keys in playlist_items
SELECT COUNT(*)
FROM playlist_items pi
LEFT JOIN playlists p ON pi.playlist_id = p.id
LEFT JOIN episodes e ON pi.episode_id = e.id
WHERE p.id IS NULL OR e.id IS NULL;
-- Should return 0
```

### Success Criteria Validation

- [ ] **SC-001**: Single video sync completes in < 5 seconds
- [ ] **SC-002**: 95%+ success rate for valid videos
- [ ] **SC-003**: Successfully synced 1000+ video channel without errors
- [ ] **SC-006**: All sync operations logged with sufficient detail
- [ ] **SC-007**: Tool setup completed in < 10 minutes

## Troubleshooting

### Common Issues

**Issue**: "Failed to connect to database"
- **Solution**: Check PostgreSQL is running: `pg_isready -h localhost -p 5432`
- **Solution**: Verify credentials: `psql -U postgres -d esg_tube -c "SELECT 1;"`

**Issue**: "YouTube API quota exceeded"
- **Solution**: Wait until quota resets (midnight Pacific Time)
- **Solution**: Request quota increase in Google Cloud Console

**Issue**: "Video not found" for valid video
- **Solution**: Check video is public (not private/unlisted)
- **Solution**: Verify API key has correct permissions

**Issue**: Slow sync performance
- **Solution**: Check network latency to YouTube API
- **Solution**: Verify database connection pooling is configured
- **Solution**: Monitor API rate limits in logs

## Next Steps

After manual testing:
1. Implement automated tests based on these scenarios
2. Set up CI pipeline to run tests on every commit
3. Deploy to staging environment for integration testing
4. Document any issues found and create bug reports
