# Implementation Plan: YouTube Video Sync

**Branch**: `001-youtube-video-sync` | **Date**: 2025-10-20 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-youtube-video-sync/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build a CLI tool in Go to sync YouTube video metadata to an existing PostgreSQL database. The tool will fetch video, channel, and playlist metadata from YouTube Data API and map it to the existing database schema (episodes, playlists, playlist_items tables). Supports single video sync, bulk channel sync, playlist sync, and scheduled automatic updates. Presenters and organizations tables are manually maintained and will NOT be auto-populated.

## Technical Context

**Language/Version**: Go 1.21+
**Primary Dependencies**:
- YouTube API: `google.golang.org/api/youtube/v3` (official Google client)
- Database: `github.com/lib/pq` (PostgreSQL driver) + `database/sql` (stdlib)
- CLI Framework: `github.com/spf13/cobra`
- Configuration: `github.com/spf13/viper`
- Retry Logic: `github.com/cenkalti/backoff/v4`
- Rate Limiting: `golang.org/x/time/rate`
- Logging: `go.uber.org/zap`
- Scheduling (P3): `github.com/robfig/cron/v3`
**Storage**: PostgreSQL (existing database schema - see `docs/db_cluster-06-06-2025@08-19-26.backup`)
**Testing**: Go Test (standard library testing package)
**Target Platform**: Linux/macOS/Windows (cross-platform CLI)
**Project Type**: single (CLI application)
**Performance Goals**:
- Single video sync: <5 seconds
- Bulk sync throughput: 100 videos/minute (respecting YouTube API rate limits)
- Handle 1000+ video channel syncs without manual intervention

**Constraints**:
- YouTube API quota limits (10,000 units/day default)
- Must respect existing database schema (no schema changes)
- Must NOT modify presenters, organizations, or their junction tables
- CLI must follow POSIX conventions

**Scale/Scope**:
- Support channels with 1000+ videos
- Handle multiple concurrent sync operations
- Production-ready error handling and retry logic

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Code Quality & Maintainability ✅

- **Type Safety**: Go provides strong static typing with type inference
- **Single Responsibility**: CLI tool has one clear purpose - sync YouTube data to database
- **Documentation**: Will document all public APIs, complex logic, and database mappings
- **Code Review**: Required before merge per constitution
- **Linting**: Go has `golint`, `go vet`, and `gofmt` for automated checks

**Status**: ✅ PASS - Go's tooling and type system align with quality requirements

### Principle II: Testing Discipline (NON-NEGOTIABLE) ✅

- **Test-First Development**: Will use TDD with Go's testing package
- **Unit Tests**: Target >80% coverage for business logic (sync operations, URL parsing, data mapping)
- **Integration Tests**: Required for YouTube API calls and PostgreSQL operations
- **Contract Tests**: Required for database schema interactions (episodes, playlists, playlist_items)
- **CI Requirements**: All tests must pass before merge

**Status**: ✅ PASS - Test coverage plan meets constitution requirements

### Principle III: User Experience Consistency ✅

- **CLI Standards**:
  - POSIX conventions (specified in FR-015)
  - `--help` for all commands
  - stdout for data, stderr for errors
  - Exit codes: 0 for success, non-zero for failures
  - `--json` flag for machine-readable output
- **Error Messages**: Actionable with context (e.g., "Failed to sync video abc123: API quota exceeded. Wait 24 hours or request quota increase")
- **Logging**: Structured logging with correlation IDs for tracing sync operations

**Status**: ✅ PASS - CLI design follows POSIX and constitution UX standards

### Principle IV: Performance & Reliability ✅

- **Performance**:
  - <5s for single video sync (meets SC-001)
  - 100 videos/min throughput (meets SC-005)
- **Error Handling**:
  - Retry logic with exponential backoff for YouTube API (FR-008)
  - Handle rate limits gracefully (FR-008)
  - Mark unavailable videos without failing entire sync (FR-013)
- **Idempotency**: Upsert operations ensure safe retries (FR-007)
- **Observability**: Comprehensive logging for all sync operations (FR-010, SC-006)

**Status**: ✅ PASS - Performance targets and reliability patterns defined

### Additional Standards Check ✅

- **Security**:
  - API keys via environment variables (no secrets in code)
  - Input validation for YouTube URLs (FR-009)
- **Data Integrity**:
  - Use PostgreSQL transactions for atomic operations
  - Validate existing schema constraints (no modifications to presenters/organizations)
- **Configuration**:
  - Database connection via environment variables
  - API key via environment variable

**Status**: ✅ PASS - Security and data integrity requirements met

### Overall Gate Status: ✅ PASS

No constitution violations identified. All principles can be satisfied with the proposed Go implementation.

## Project Structure

### Documentation (this feature)

```
specs/001-youtube-video-sync/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
│   └── cli-commands.md  # CLI command specifications
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```
cmd/
└── youtube-sync/        # Main CLI entry point
    └── main.go

internal/
├── youtube/            # YouTube API client wrapper
│   ├── client.go
│   ├── video.go
│   ├── channel.go
│   └── playlist.go
├── database/           # PostgreSQL operations
│   ├── client.go
│   ├── episodes.go
│   ├── playlists.go
│   └── playlist_items.go
├── sync/               # Core sync logic
│   ├── video_sync.go
│   ├── channel_sync.go
│   ├── playlist_sync.go
│   └── scheduler.go
├── parser/             # URL parsing and validation
│   ├── youtube_url.go
│   └── validator.go
└── logger/             # Structured logging
    └── logger.go

pkg/                    # Public packages (if needed for extensibility)
└── config/
    └── config.go       # Configuration management

tests/
├── unit/              # Unit tests (mirrors internal/ structure)
│   ├── youtube/
│   ├── database/
│   ├── sync/
│   └── parser/
├── integration/       # Integration tests
│   ├── youtube_api_test.go
│   └── database_test.go
└── contract/          # Database schema contract tests
    └── schema_test.go

go.mod                 # Go module definition
go.sum                 # Go dependency checksums
.env.example           # Example environment configuration
README.md              # Setup and usage instructions
```

**Structure Decision**: Using Go's standard project layout with `cmd/` for entry points, `internal/` for private application code, and `pkg/` for any potentially reusable components. This structure supports:
- Clear separation of concerns (YouTube API, database, sync logic)
- Private internal packages (cannot be imported by external projects)
- Standard Go testing structure
- Idiomatic Go project organization

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

No constitution violations identified - this section is empty.

## Post-Design Constitution Re-Check

After completing Phase 0 (Research) and Phase 1 (Design), re-evaluating constitution compliance:

### Phase 0 - Research Outcomes
- ✅ All technology choices documented in [research.md](research.md)
- ✅ Decisions justified with rationale and alternatives considered
- ✅ Selected dependencies are production-ready and widely adopted

### Phase 1 - Design Artifacts
- ✅ [data-model.md](data-model.md) defines clear entity mappings
- ✅ [contracts/cli-commands.md](contracts/cli-commands.md) specifies CLI interface
- ✅ [quickstart.md](quickstart.md) provides manual testing scenarios
- ✅ Agent context updated with technology stack

### Constitution Compliance Review

**Principle I: Code Quality & Maintainability** ✅
- Clear project structure defined (`cmd/`, `internal/`, `tests/`)
- Single Responsibility maintained (separate packages for youtube, database, sync, parser)
- Go's type system provides strong type safety
- Standard Go tooling (gofmt, golint, go vet) for linting

**Principle II: Testing Discipline** ✅
- Test-first development planned in tasks
- Unit tests for all packages (youtube, database, sync, parser)
- Integration tests for API and database operations
- Contract tests for schema compliance
- Target coverage >80% documented

**Principle III: User Experience Consistency** ✅
- CLI follows POSIX conventions (FR-015)
- Cobra provides excellent help and flag handling
- JSON output mode for scripting (`--json`)
- Actionable error messages with suggestions
- Structured logging with zap (JSON format)

**Principle IV: Performance & Reliability** ✅
- Performance targets defined and testable
- Retry logic with exponential backoff (cenkalti/backoff)
- Rate limiting to respect YouTube quotas (golang.org/x/time/rate)
- Idempotent upsert operations for safe retries
- Comprehensive logging for observability

**Final Status**: ✅ ALL PRINCIPLES SATISFIED

No new violations introduced during design phase. Implementation can proceed to Phase 2 (Task Generation).

## Next Steps

Phase 0 and Phase 1 are complete. To proceed:

1. **Review the design artifacts**:
   - [research.md](research.md) - Technology decisions
   - [data-model.md](data-model.md) - Entity mappings
   - [contracts/cli-commands.md](contracts/cli-commands.md) - CLI interface
   - [quickstart.md](quickstart.md) - Manual testing guide

2. **Generate implementation tasks**:
   ```bash
   /speckit.tasks
   ```
   This will create `tasks.md` with dependency-ordered tasks for TDD implementation.

3. **Execute implementation**:
   ```bash
   /speckit.implement
   ```
   This will execute tasks in order, enforcing test-first development.

