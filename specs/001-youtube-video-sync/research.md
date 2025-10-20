# Research: YouTube Video Sync

**Feature**: 001-youtube-video-sync
**Date**: 2025-10-20
**Purpose**: Resolve technology choices and design patterns for implementing the YouTube sync CLI tool in Go

## Research Questions

From Technical Context, these areas need clarification:
1. Which YouTube API client library for Go?
2. Which PostgreSQL driver and ORM (if any)?
3. Which CLI framework for Go?
4. How to handle YouTube API rate limiting and quotas?
5. Scheduling mechanism for automatic syncs
6. Structured logging library

## 1. YouTube API Client Library

### Decision: `google.golang.org/api/youtube/v3`

### Rationale:
- Official Google API client library for Go
- Actively maintained by Google
- Comprehensive coverage of YouTube Data API v3
- Built-in support for API key and OAuth2 authentication
- Type-safe Go interfaces for all API operations
- Handles pagination automatically with iterators
- Integrated with `golang.org/x/oauth2` for credential management

### Alternatives Considered:
- **Custom HTTP client**: Rejected because it requires manually implementing YouTube API spec, error handling, and pagination. The official library is well-tested and handles edge cases.
- **`github.com/kkdai/youtube`**: Rejected because it's focused on video downloading, not metadata fetching via official API. Would violate YouTube TOS.

### Usage Pattern:
```go
import "google.golang.org/api/youtube/v3"

service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
video, err := service.Videos.List([]string{"snippet", "statistics"}).Id(videoID).Do()
```

## 2. PostgreSQL Driver and Data Access

### Decision: `github.com/lib/pq` (driver) + `database/sql` (standard library)

### Rationale:
- `lib/pq` is the most mature and widely-used pure Go PostgreSQL driver
- Using standard library `database/sql` avoids ORM complexity for simple CRUD operations
- Direct SQL provides better control over upsert logic (INSERT ... ON CONFLICT)
- No ORM layer means clearer mapping to existing database schema
- Easier to write contract tests against actual schema
- Better performance for bulk operations (batch inserts)

### Alternatives Considered:
- **GORM**: Popular ORM with migrations and model generation. Rejected because:
  - We cannot modify the existing schema
  - ORM abstractions add complexity for simple sync operations
  - Harder to ensure we don't accidentally touch presenters/organizations tables
- **sqlx**: Provides nice extensions (StructScan, Named queries). Rejected because:
  - Adds dependency for marginal benefit over database/sql
  - Our queries are simple enough for standard library
- **sqlc**: Generates type-safe Go code from SQL. Considered but rejected for v1:
  - Requires learning new tooling
  - Our schema already exists (no need for code generation)
  - Can revisit in future if query complexity increases

### Usage Pattern:
```go
import (
    "database/sql"
    _ "github.com/lib/pq"
)

db, err := sql.Open("postgres", connString)

// Upsert video (idempotent)
_, err = db.Exec(`
    INSERT INTO episodes (video_id, title, description, published_at, view_count, active)
    VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (video_id) DO UPDATE SET
        title = EXCLUDED.title,
        description = EXCLUDED.description,
        view_count = EXCLUDED.view_count,
        updated_at = NOW()
`, videoID, title, desc, publishedAt, viewCount, true)
```

## 3. CLI Framework

### Decision: `github.com/spf13/cobra`

### Rationale:
- Industry standard for Go CLI applications (used by kubectl, Hugo, GitHub CLI)
- Excellent support for subcommands (`youtube-sync video`, `youtube-sync channel`)
- Automatic help generation and flag parsing
- POSIX-compliant flag handling (supports both `--flag` and `-f`)
- Built-in completion generation (bash, zsh, fish, powershell)
- Integrates well with `github.com/spf13/viper` for configuration management
- Active community and extensive documentation

### Alternatives Considered:
- **`flag` (standard library)**: Rejected because:
  - No built-in subcommand support
  - Basic help generation
  - Would require significant manual work for complex CLI
- **`github.com/urfave/cli`**: Good alternative but:
  - Less intuitive API than Cobra
  - Smaller ecosystem and community
  - Cobra is more widely adopted in enterprise Go projects

### Usage Pattern:
```go
import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
    Use:   "youtube-sync",
    Short: "Sync YouTube videos to PostgreSQL database",
}

var videoCmd = &cobra.Command{
    Use:   "video [url or id]",
    Short: "Sync a single YouTube video",
    Args:  cobra.ExactArgs(1),
    RunE:  func(cmd *cobra.Command, args []string) error {
        // sync logic
    },
}

rootCmd.AddCommand(videoCmd)
```

## 4. Rate Limiting and Retry Logic

### Decision: `github.com/cenkalti/backoff/v4` + custom rate limiter

### Rationale:
- **Backoff library**: Provides battle-tested exponential backoff with jitter for retry logic
- **Custom rate limiter**: Use `golang.org/x/time/rate` (standard extended library) to respect YouTube quota
- YouTube API quota: 10,000 units/day default
  - Videos.list costs 1 unit
  - Channels.list costs 1 unit
  - PlaylistItems.list costs 1 unit
  - Need to track quota usage and throttle requests

### Strategy:
1. **Rate Limiter**: Limit to ~100 requests/minute (well below quota limits, leaves buffer)
2. **Exponential Backoff**: Retry on 429 (rate limit), 500 (server error), network errors
3. **Quota Tracking**: Log quota usage per sync operation for monitoring

### Implementation:
```go
import (
    "github.com/cenkalti/backoff/v4"
    "golang.org/x/time/rate"
)

// Rate limiter: 100 requests per minute
limiter := rate.NewLimiter(rate.Every(600*time.Millisecond), 1)

// Exponential backoff with retries
operation := func() error {
    limiter.Wait(ctx) // throttle request
    return makeYouTubeAPICall()
}

err := backoff.Retry(operation, backoff.NewExponentialBackOff())
```

### Alternatives Considered:
- **`github.com/go-resty/resty`**: HTTP client with built-in retry. Rejected because YouTube client already handles HTTP, and backoff is more flexible.
- **Manual implementation**: Rejected because backoff library is well-tested and handles edge cases (jitter, max retry time).

## 5. Scheduling Mechanism

### Decision: `github.com/robfig/cron/v3` (for User Story 3 - P3 priority)

### Rationale:
- Production-ready cron library with cron expression support
- Thread-safe for concurrent jobs
- Supports second-level granularity if needed
- Can run as daemon or standalone command
- Well-documented and widely used

### Implementation Approach:
```go
import "github.com/robfig/cron/v3"

c := cron.New()

// Run daily at 2 AM
c.AddFunc("0 2 * * *", func() {
    // Run sync for all tracked channels/playlists
    performScheduledSync()
})

c.Start()
```

### Alternatives Considered:
- **External cron (system crontab)**: Simpler but:
  - Requires system configuration
  - Less portable across platforms
  - Harder to test
  - No programmatic control
- **`github.com/go-co-op/gocron`**: Similar functionality, but robfig/cron is more established

### Note: This is P3 priority - defer implementation until P1 and P2 are complete

## 6. Structured Logging

### Decision: `github.com/sirupsen/logrus` or `go.uber.org/zap`

### Rationale:
Both are excellent structured logging libraries with JSON output support.

**Recommendation: `go.uber.org/zap`** because:
- Highest performance (zero-allocation logging)
- Structured logging with strong typing
- JSON output for machine parsing
- Supports log levels (DEBUG, INFO, WARN, ERROR)
- Can add correlation IDs for request tracing
- Actively maintained by Uber

### Usage Pattern:
```go
import "go.uber.org/zap"

logger, _ := zap.NewProduction() // JSON output
defer logger.Sync()

logger.Info("syncing video",
    zap.String("video_id", videoID),
    zap.String("correlation_id", correlationID),
    zap.Int("view_count", viewCount),
)
```

### Alternatives Considered:
- **`log` (standard library)**: Too basic, no structured logging
- **`logrus`**: Slower than zap, but easier API. Good alternative if zap proves too complex.

## 7. Configuration Management

### Decision: `github.com/spf13/viper`

### Rationale:
- Integrates seamlessly with Cobra CLI framework
- Supports environment variables, config files (.env, YAML, JSON), and CLI flags
- Automatic precedence: flags > env vars > config file > defaults
- Popular in Go ecosystem

### Configuration Schema:
```yaml
# config.yaml or environment variables
database:
  host: localhost
  port: 5432
  user: postgres
  password: ${POSTGRES_PASSWORD}  # from env var
  database: esg_tube

youtube:
  api_key: ${YOUTUBE_API_KEY}  # from env var

logging:
  level: info
  format: json
```

## 8. Testing Strategy

### Tools:
- **Unit Tests**: Standard library `testing` package + `github.com/stretchr/testify` for assertions
- **Mocking**: `github.com/stretchr/testify/mock` for mocking YouTube API and database
- **Integration Tests**: Use `dockertest` to spin up PostgreSQL container for real DB tests
- **Table-Driven Tests**: Go idiom for testing multiple scenarios

### Coverage Tool:
- Use `go test -cover` and `go tool cover` for coverage reports
- Target: >80% coverage for all business logic

## Summary of Dependencies

### Core Dependencies:
```go
require (
    google.golang.org/api v0.150.0           // YouTube API client
    github.com/lib/pq v1.10.9                 // PostgreSQL driver
    github.com/spf13/cobra v1.8.0             // CLI framework
    github.com/spf13/viper v1.18.0            // Configuration
    github.com/cenkalti/backoff/v4 v4.2.1     // Retry logic
    golang.org/x/time v0.5.0                  // Rate limiting
    go.uber.org/zap v1.26.0                   // Structured logging
    github.com/robfig/cron/v3 v3.0.1          // Scheduling (P3 priority)
)
```

### Testing Dependencies:
```go
require (
    github.com/stretchr/testify v1.8.4        // Test assertions and mocking
    github.com/ory/dockertest/v3 v3.10.0      // Docker-based integration tests
)
```

## Next Steps

With these technology decisions made, proceed to:
1. **Phase 1**: Generate `data-model.md` with entity-to-database mappings
2. **Phase 1**: Create CLI command contracts in `contracts/cli-commands.md`
3. **Phase 1**: Generate `quickstart.md` with manual testing scenarios
