# CLI Command Contracts: YouTube Video Sync

**Feature**: 001-youtube-video-sync
**Date**: 2025-10-20
**Purpose**: Define the command-line interface contract for the YouTube sync tool

## Binary Name

`youtube-sync`

## Global Flags

These flags are available for all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--config` | `-c` | string | `""` | Path to config file (optional) |
| `--verbose` | `-v` | bool | false | Enable verbose logging (DEBUG level) |
| `--json` | `-j` | bool | false | Output results in JSON format |
| `--quiet` | `-q` | bool | false | Suppress all output except errors |
| `--dry-run` | | bool | false | Simulate sync without writing to database |

**Exit Codes**:
- `0`: Success
- `1`: General error
- `2`: Invalid arguments/usage
- `3`: Configuration error (missing API key, bad DB connection)
- `4`: API error (YouTube API failure)
- `5`: Database error

## Commands

### 1. Root Command

```bash
youtube-sync [--help]
```

**Description**: Display help and list available commands

**Output**:
```
YouTube Video Sync Tool - Sync YouTube video metadata to PostgreSQL

Usage:
  youtube-sync [command]

Available Commands:
  video       Sync a single YouTube video
  channel     Sync all videos from a YouTube channel
  playlist    Sync all videos from a YouTube playlist
  schedule    Manage scheduled sync jobs (P3 - not in MVP)
  version     Display version information

Flags:
  -c, --config string   Path to config file
  -v, --verbose         Enable verbose logging
  -j, --json            Output in JSON format
  -q, --quiet           Suppress all output except errors
      --dry-run         Simulate sync without writing to database
  -h, --help            Help for youtube-sync

Use "youtube-sync [command] --help" for more information about a command.
```

---

### 2. Video Command (User Story 1 - P1)

**Syntax**:
```bash
youtube-sync video <url-or-id> [flags]
```

**Arguments**:
- `url-or-id` (required): YouTube video URL or video ID

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Force update even if video already exists |

**Supported URL Formats**:
- Full URL: `https://www.youtube.com/watch?v=dQw4w9WgXcQ`
- Short URL: `https://youtu.be/dQw4w9WgXcQ`
- Video ID only: `dQw4w9WgXcQ`
- Embed URL: `https://www.youtube.com/embed/dQw4w9WgXcQ`

**Examples**:
```bash
# Sync by full URL
youtube-sync video "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Sync by video ID
youtube-sync video dQw4w9WgXcQ

# Sync with verbose output
youtube-sync video dQw4w9WgXcQ --verbose

# Dry run (simulate only)
youtube-sync video dQw4w9WgXcQ --dry-run

# Force update existing video
youtube-sync video dQw4w9WgXcQ --force

# JSON output for scripting
youtube-sync video dQw4w9WgXcQ --json
```

**Success Output (human-readable)**:
```
✓ Successfully synced video: "Never Gonna Give You Up"
  Video ID:     dQw4w9WgXcQ
  Channel:      Rick Astley
  Published:    2009-10-25
  Views:        1,234,567,890
  Duration:     3:33
  Database ID:  42
```

**Success Output (JSON format)**:
```json
{
  "status": "success",
  "video": {
    "id": "dQw4w9WgXcQ",
    "title": "Never Gonna Give You Up",
    "channel_name": "Rick Astley",
    "channel_id": "UCuAXFkgsw1L7xaCfnd5JJOw",
    "published_at": "2009-10-25T06:57:33Z",
    "view_count": 1234567890,
    "duration": "PT3M33S",
    "database_id": 42
  },
  "operation": "created",
  "timestamp": "2025-10-20T14:30:00Z"
}
```

**Error Output**:
```
✗ Error: Failed to sync video
  Reason: Video not found (404)
  Video ID: invalid123

Suggestion: Check that the video ID or URL is correct and the video is publicly accessible.

Exit code: 4
```

**Error Output (JSON)**:
```json
{
  "status": "error",
  "error": {
    "code": "VIDEO_NOT_FOUND",
    "message": "Video not found (404)",
    "video_id": "invalid123"
  },
  "timestamp": "2025-10-20T14:30:00Z"
}
```

---

### 3. Channel Command (User Story 2 - P2)

**Syntax**:
```bash
youtube-sync channel <url-or-id> [flags]
```

**Arguments**:
- `url-or-id` (required): YouTube channel URL or channel ID

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | 0 | Max number of videos to sync (0 = unlimited) |
| `--force` | bool | false | Force update all videos even if they exist |
| `--since` | string | "" | Only sync videos published after this date (YYYY-MM-DD) |

**Supported URL Formats**:
- Channel URL: `https://www.youtube.com/channel/UCuAXFkgsw1L7xaCfnd5JJOw`
- Custom URL: `https://www.youtube.com/@RickAstleyYT`
- Channel ID only: `UCuAXFkgsw1L7xaCfnd5JJOw`

**Examples**:
```bash
# Sync all videos from channel
youtube-sync channel "https://www.youtube.com/@RickAstleyYT"

# Sync only recent videos (last 50)
youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --limit 50

# Sync only videos published after specific date
youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --since 2024-01-01

# Verbose progress tracking
youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --verbose

# JSON output for automation
youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --json
```

**Success Output (human-readable)**:
```
Syncing channel: Rick Astley
Channel ID: UCuAXFkgsw1L7xaCfnd5JJOw
Total videos: 150

Progress: [████████████████████] 100% (150/150)

✓ Sync complete
  Success: 148 videos
  Skipped: 2 videos (unavailable)
  Duration: 5m 23s

Failed videos:
  - abc123: Video deleted
  - def456: Video private
```

**Success Output (JSON format)**:
```json
{
  "status": "success",
  "channel": {
    "id": "UCuAXFkgsw1L7xaCfnd5JJOw",
    "name": "Rick Astley",
    "video_count": 150
  },
  "results": {
    "total": 150,
    "success": 148,
    "failed": 2,
    "skipped": 0,
    "duration_seconds": 323
  },
  "failed_videos": [
    {"id": "abc123", "error": "Video deleted"},
    {"id": "def456", "error": "Video private"}
  ],
  "timestamp": "2025-10-20T14:35:23Z"
}
```

**Progress Output (verbose mode)**:
```
[2025-10-20 14:30:00] INFO  Starting channel sync: UCuAXFkgsw1L7xaCfnd5JJOw
[2025-10-20 14:30:01] INFO  Fetching channel metadata...
[2025-10-20 14:30:02] INFO  Channel: Rick Astley (150 videos)
[2025-10-20 14:30:03] INFO  Syncing video 1/150: dQw4w9WgXcQ
[2025-10-20 14:30:04] INFO  ✓ Synced: Never Gonna Give You Up
[2025-10-20 14:30:05] INFO  Syncing video 2/150: xyz789abc
...
[2025-10-20 14:35:23] INFO  Sync complete: 148 success, 2 failed
```

---

### 4. Playlist Command (User Story 4 - P3)

**Syntax**:
```bash
youtube-sync playlist <url-or-id> [flags]
```

**Arguments**:
- `url-or-id` (required): YouTube playlist URL or playlist ID

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--force` | bool | false | Force update all videos and playlist metadata |

**Supported URL Formats**:
- Playlist URL: `https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf`
- Playlist ID only: `PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf`

**Examples**:
```bash
# Sync playlist with all videos
youtube-sync playlist "https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"

# Sync by playlist ID
youtube-sync playlist PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf

# Verbose output
youtube-sync playlist PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf --verbose

# JSON output
youtube-sync playlist PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf --json
```

**Success Output (human-readable)**:
```
Syncing playlist: "Best of Rick Astley"
Playlist ID: PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf
Total videos: 25

Progress: [████████████████████] 100% (25/25)

✓ Sync complete
  Playlist saved (ID: 5)
  Videos synced: 25
  Playlist items created: 25
  Duration: 1m 12s
```

**Success Output (JSON format)**:
```json
{
  "status": "success",
  "playlist": {
    "id": "PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf",
    "name": "Best of Rick Astley",
    "video_count": 25,
    "database_id": 5
  },
  "results": {
    "total_videos": 25,
    "videos_synced": 25,
    "playlist_items_created": 25,
    "duration_seconds": 72
  },
  "timestamp": "2025-10-20T14:45:00Z"
}
```

---

### 5. Schedule Command (User Story 3 - P3)

**Note**: This command is P3 priority and will be implemented after P1 and P2 are complete.

**Syntax**:
```bash
youtube-sync schedule <subcommand> [args] [flags]
```

**Subcommands**:
- `add`: Add a new scheduled sync job
- `list`: List all scheduled jobs
- `remove`: Remove a scheduled job
- `start`: Start the schedule daemon
- `stop`: Stop the schedule daemon

**Examples**:
```bash
# Add daily channel sync at 2 AM
youtube-sync schedule add --type channel --target UCuAXFkgsw1L7xaCfnd5JJOw --cron "0 2 * * *"

# List all scheduled jobs
youtube-sync schedule list

# Remove a scheduled job
youtube-sync schedule remove --id abc123

# Start scheduler daemon
youtube-sync schedule start

# Stop scheduler daemon
youtube-sync schedule stop
```

*Detailed specification deferred until P3 implementation.*

---

### 6. Version Command

**Syntax**:
```bash
youtube-sync version [--short]
```

**Flags**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--short` | bool | false | Show only version number |

**Examples**:
```bash
# Full version info
youtube-sync version

# Short version
youtube-sync version --short
```

**Output (full)**:
```
YouTube Video Sync Tool
Version:    1.0.0
Commit:     a1b2c3d
Build Date: 2025-10-20T10:00:00Z
Go Version: go1.21.5
Platform:   linux/amd64
```

**Output (short)**:
```
1.0.0
```

---

## Configuration File

**Location**: `~/.youtube-sync/config.yaml` or specified via `--config` flag

**Format**: YAML

**Example**:
```yaml
# YouTube API configuration
youtube:
  api_key: ${YOUTUBE_API_KEY}  # Prefer environment variable

# Database configuration
database:
  host: localhost
  port: 5432
  user: postgres
  password: ${POSTGRES_PASSWORD}  # Prefer environment variable
  database: esg_tube
  sslmode: disable
  max_connections: 10

# Logging configuration
logging:
  level: info           # debug, info, warn, error
  format: text          # text or json
  file: ""              # Path to log file (empty = stdout)

# Rate limiting
rate_limit:
  requests_per_minute: 100

# Retry configuration
retry:
  max_attempts: 3
  initial_backoff: 1s
  max_backoff: 60s
```

## Environment Variables

These override config file settings:

| Variable | Description | Example |
|----------|-------------|---------|
| `YOUTUBE_API_KEY` | YouTube Data API key (required) | `AIzaSyABC123...` |
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Database user | `postgres` |
| `POSTGRES_PASSWORD` | Database password | `secret` |
| `POSTGRES_DATABASE` | Database name | `esg_tube` |
| `POSTGRES_SSLMODE` | SSL mode | `disable` |
| `LOG_LEVEL` | Logging level | `info` |
| `LOG_FORMAT` | Log format | `json` |

**Precedence**: CLI flags > Environment variables > Config file > Defaults

## Error Handling

### Common Error Scenarios

#### 1. Missing API Key
```
✗ Error: YouTube API key not configured

Configuration error: YOUTUBE_API_KEY environment variable is not set.

To fix:
  1. Export the environment variable: export YOUTUBE_API_KEY="your-key-here"
  2. Or add to config file: youtube.api_key: "your-key-here"
  3. Get an API key from: https://console.developers.google.com/

Exit code: 3
```

#### 2. Database Connection Failure
```
✗ Error: Failed to connect to database

Database error: connection refused (host: localhost:5432)

To fix:
  1. Ensure PostgreSQL is running
  2. Check database connection settings in config or environment variables
  3. Test connection: psql -h localhost -U postgres -d esg_tube

Exit code: 5
```

#### 3. API Quota Exceeded
```
✗ Error: YouTube API quota exceeded

API error: Daily quota limit reached (10,000 units)

To fix:
  1. Wait until quota resets (typically at midnight Pacific Time)
  2. Request quota increase: https://console.developers.google.com/
  3. Reduce sync frequency to stay within quota

Exit code: 4
```

#### 4. Invalid URL Format
```
✗ Error: Invalid YouTube URL

Parse error: Could not extract video ID from "https://example.com/video"

Supported formats:
  - https://www.youtube.com/watch?v=VIDEO_ID
  - https://youtu.be/VIDEO_ID
  - VIDEO_ID (11 characters)
  - https://www.youtube.com/embed/VIDEO_ID

Exit code: 2
```

## Testing Contract

Each command must have:
1. **Unit tests** for argument parsing and validation
2. **Integration tests** with mocked YouTube API
3. **Contract tests** for database operations
4. **End-to-end tests** with test fixtures

See `tests/` directory for test implementation.

## Next Steps

With CLI contracts defined, proceed to:
1. Generate `quickstart.md` with manual testing scenarios
2. Update agent context file with complete technology stack
