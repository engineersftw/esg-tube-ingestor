# YouTube Video Sync

A command-line tool to sync YouTube video metadata to PostgreSQL database. Supports single video sync, bulk channel sync, playlist sync, and scheduled automatic updates.

## Features

- **Single Video Sync**: Sync individual YouTube videos by URL or ID
- **Bulk Channel Sync**: Sync all videos from a YouTube channel with pagination
- **Playlist Sync**: Sync YouTube playlists with all videos and maintain order
- **Scheduled Syncs**: Configure automatic syncs to run on a schedule (P3)
- **Idempotent Operations**: Safe to run multiple times without duplicates
- **Robust Error Handling**: Retry logic, rate limiting, and graceful failure handling
- **Structured Logging**: JSON or text format with correlation IDs

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database with existing schema
  - **Easy option**: Use Docker Compose (see below)
  - **Manual option**: Install PostgreSQL locally
- YouTube Data API key ([Get one here](https://console.developers.google.com/))

### Quick Start with Docker (Recommended)

The easiest way to get started is using Docker Compose for the database:

```bash
# Start PostgreSQL database
docker-compose up -d

# Verify database is running
docker-compose ps

# The database is now available at localhost:5432
# Default credentials: postgres/postgres
# Database name: esg_tube
```

See [docs/local-development.md](docs/local-development.md) for detailed Docker setup, schema initialization, and troubleshooting.

### Installation

```bash
# Clone the repository
git clone https://github.com/engineersftw/youtube-sync.git
cd youtube-sync

# Install dependencies
go mod download

# Build the binary
go build -o youtube-sync cmd/youtube-sync/main.go

# Or install to $GOPATH/bin
go install cmd/youtube-sync/main.go
```

### Configuration

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and set your credentials:
   ```bash
   YOUTUBE_API_KEY=your-api-key-here
   POSTGRES_PASSWORD=your-db-password
   ```

3. Or create a config file at `~/.youtube-sync/config.yaml`:
   ```yaml
   youtube:
     api_key: your-api-key-here

   database:
     host: localhost
     port: 5432
     user: postgres
     password: your-password
     database: esg_tube
     sslmode: disable

   logging:
     level: info
     format: text
   ```

### Usage

```bash
# Sync a single video
youtube-sync video "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

# Sync a video by ID only
youtube-sync video dQw4w9WgXcQ

# Sync all videos from a channel (limited to 50)
youtube-sync channel "https://www.youtube.com/@TEDEd" --limit 50

# Sync only recent videos from a channel
youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --since 2024-01-01

# Sync a playlist
youtube-sync playlist "https://www.youtube.com/playlist?list=PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf"

# Dry run (simulate without database writes)
youtube-sync video dQw4w9WgXcQ --dry-run

# Verbose output
youtube-sync video dQw4w9WgXcQ --verbose

# JSON output for scripting
youtube-sync video dQw4w9WgXcQ --json
```

## Commands

### `video` - Sync Single Video

```bash
youtube-sync video <url-or-id> [flags]

Flags:
  --force         Force update even if video already exists
  --dry-run       Simulate sync without writing to database
  --json          Output in JSON format
  --verbose       Enable verbose logging
```

### `channel` - Sync Channel Videos

```bash
youtube-sync channel <url-or-id> [flags]

Flags:
  --limit int     Max number of videos to sync (0 = unlimited)
  --since string  Only sync videos published after this date (YYYY-MM-DD)
  --force         Force update all videos
```

### `playlist` - Sync Playlist

```bash
youtube-sync playlist <url-or-id> [flags]

Flags:
  --force    Force update all videos and playlist metadata
```

### `version` - Show Version

```bash
youtube-sync version [--short]
```

## Exit Codes

- `0`: Success
- `1`: General error
- `2`: Invalid arguments/usage
- `3`: Configuration error (missing API key, bad DB connection)
- `4`: API error (YouTube API failure)
- `5`: Database error

## Database Schema

This tool syncs to an existing PostgreSQL schema with the following tables:

- **episodes**: Stores YouTube video metadata
- **playlists**: Stores YouTube playlist metadata
- **playlist_items**: Links videos to playlists with sort order

**Note**: The `presenters` and `organizations` tables are manually maintained and will NOT be auto-populated by this tool.

## Performance

- Single video sync: < 5 seconds
- Bulk sync throughput: ~100 videos/minute
- Handles channels with 1000+ videos

## Development

### Using the Makefile

A Makefile is provided for common development tasks:

```bash
# See all available commands
make help

# Start database and build in one command
make dev

# Build the binary
make build

# Run all tests
make test

# Run unit tests only
make test-unit

# Run with coverage report
make test-coverage

# Start database
make db-up

# View database logs
make db-logs

# Connect to database shell
make db-shell

# Stop database
make db-down

# Format code
make fmt

# Run linters
make lint
```

### Running Tests

```bash
# Run all tests
make test
# or: go test ./...

# Run tests with coverage
make test-coverage
# or: go test ./... -cover

# Run unit tests only
make test-unit

# Run integration tests (requires PostgreSQL)
make test-integration
# or: go test ./tests/integration/... -v

# Run contract tests
go test ./tests/contract/... -v
```

### Project Structure

```
cmd/youtube-sync/       # Main CLI entry point
internal/
  ├── youtube/          # YouTube API client wrapper
  ├── database/         # PostgreSQL operations
  ├── sync/             # Core sync logic
  ├── parser/           # URL parsing and validation
  └── logger/           # Structured logging
pkg/config/             # Configuration management
tests/                  # Test suites
  ├── unit/             # Unit tests
  ├── integration/      # Integration tests
  └── contract/         # Database schema contract tests
```

## Troubleshooting

### "Failed to connect to database"

- Check PostgreSQL is running: `pg_isready -h localhost -p 5432`
- Verify credentials: `psql -U postgres -d esg_tube -c "SELECT 1;"`

### "YouTube API quota exceeded"

- Wait until quota resets (midnight Pacific Time)
- Request quota increase in Google Cloud Console
- Reduce sync frequency

### "Video not found" for valid video

- Check video is public (not private/unlisted)
- Verify API key has correct permissions

## License

[Add your license here]

## Contributing

[Add contributing guidelines here]

## Support

For issues and questions, please use the GitHub issue tracker.
