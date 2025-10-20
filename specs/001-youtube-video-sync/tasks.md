# Tasks: YouTube Video Sync

**Input**: Design documents from `/specs/001-youtube-video-sync/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cli-commands.md

**Tests**: REQUIRED per Constitution Principle II (Test-First Development)

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic Go project structure

- [X] T001 Initialize Go module with `go mod init github.com/engineersftw/youtube-sync`
- [X] T002 Create directory structure: cmd/youtube-sync/, internal/, pkg/config/, tests/
- [X] T003 [P] Install core dependencies: cobra, viper, zap, lib/pq, google.golang.org/api/youtube/v3
- [X] T004 [P] Install development dependencies: testify, dockertest for tests
- [X] T005 [P] Create .env.example with YOUTUBE_API_KEY and database config examples
- [X] T006 [P] Create README.md with setup instructions and quickstart guide
- [X] T007 [P] Create .gitignore for Go projects (bin/, .env, go.sum variations)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T008 Write test for configuration loading in tests/unit/config/config_test.go (MUST FAIL)
- [X] T009 Implement configuration management in pkg/config/config.go (viper-based, env vars + file)
- [X] T010 Write test for database connection in tests/integration/database_test.go (MUST FAIL)
- [X] T011 Implement database client with connection pooling in internal/database/client.go
- [X] T012 [P] Write contract test for episodes table schema in tests/contract/schema_test.go (MUST FAIL)
- [X] T013 [P] Implement episodes table operations (upsert) in internal/database/episodes.go
- [X] T014 Write test for structured logger in tests/unit/logger/logger_test.go (MUST FAIL)
- [X] T015 Implement structured logging (zap) with correlation IDs in internal/logger/logger.go
- [X] T016 Write test for YouTube API client initialization in tests/unit/youtube/client_test.go (MUST FAIL)
- [X] T017 Implement YouTube API client wrapper with rate limiting in internal/youtube/client.go
- [X] T018 Write test for exponential backoff retry logic in tests/unit/youtube/client_test.go (MUST FAIL)
- [X] T019 Implement retry logic with backoff in internal/youtube/client.go
- [X] T020 Create main CLI entry point skeleton in cmd/youtube-sync/main.go (cobra root command)

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Manual Single Video Sync (Priority: P1) 🎯 MVP

**Goal**: Sync a single YouTube video to the database by providing a URL or video ID

**Independent Test**: Run `youtube-sync video dQw4w9WgXcQ` and verify video appears in episodes table

### Tests for User Story 1 (Test-First Development)

**NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T021 [P] [US1] Write unit test for YouTube video ID extraction from URLs in tests/unit/parser/youtube_url_test.go (MUST FAIL)
- [ ] T022 [P] [US1] Write unit test for YouTube URL validation in tests/unit/parser/youtube_url_test.go (MUST FAIL)
- [ ] T023 [P] [US1] Write unit test for video metadata fetching in tests/unit/youtube/video_test.go (MUST FAIL with mocked API)
- [ ] T024 [P] [US1] Write unit test for video metadata mapping to episodes table in tests/unit/sync/video_sync_test.go (MUST FAIL)
- [ ] T025 [US1] Write integration test for end-to-end video sync in tests/integration/video_sync_test.go (MUST FAIL)

### Implementation for User Story 1

- [ ] T026 [P] [US1] Implement YouTube URL parser (extract video ID from various URL formats) in internal/parser/youtube_url.go
- [ ] T027 [P] [US1] Implement URL validator (validate format, handle malformed URLs) in internal/parser/validator.go
- [ ] T028 [US1] Implement video metadata fetcher (YouTube API videos.list) in internal/youtube/video.go
- [ ] T029 [US1] Implement video sync logic (fetch + map + upsert to episodes) in internal/sync/video_sync.go
- [ ] T030 [US1] Implement "video" CLI command with cobra in cmd/youtube-sync/main.go
- [ ] T031 [US1] Add --force, --dry-run, --json, --verbose flags to video command
- [ ] T032 [US1] Add error handling for invalid video IDs (FR-009) in internal/sync/video_sync.go
- [ ] T033 [US1] Add error handling for unavailable videos (404/403) mark as active=false (FR-013)
- [ ] T034 [US1] Add structured logging for video sync operations with correlation IDs
- [ ] T035 [US1] Add success/error output formatting (human-readable + JSON modes)

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently
**Manual Test**: Follow quickstart.md Test Case 1.1-1.6 to validate single video sync

---

## Phase 4: User Story 2 - Bulk Video Sync from Channel (Priority: P2)

**Goal**: Sync all videos from a YouTube channel with pagination support and progress tracking

**Independent Test**: Run `youtube-sync channel UCuAXFkgsw1L7xaCfnd5JJOw --limit 10` and verify all channel videos appear in episodes table

### Tests for User Story 2 (Test-First Development)

- [ ] T036 [P] [US2] Write unit test for channel ID extraction from URLs in tests/unit/parser/youtube_url_test.go (MUST FAIL)
- [ ] T037 [P] [US2] Write unit test for channel metadata fetching in tests/unit/youtube/channel_test.go (MUST FAIL with mocked API)
- [ ] T038 [P] [US2] Write unit test for pagination handling in tests/unit/youtube/channel_test.go (MUST FAIL)
- [ ] T039 [US2] Write unit test for channel sync logic in tests/unit/sync/channel_sync_test.go (MUST FAIL)
- [ ] T040 [US2] Write integration test for bulk channel sync in tests/integration/channel_sync_test.go (MUST FAIL)

### Implementation for User Story 2

- [ ] T041 [P] [US2] Extend URL parser to handle channel URLs (various formats) in internal/parser/youtube_url.go
- [ ] T042 [US2] Implement channel metadata fetcher (YouTube API channels.list) in internal/youtube/channel.go
- [ ] T043 [US2] Implement channel video listing with pagination in internal/youtube/channel.go
- [ ] T044 [US2] Implement channel sync logic (iterate videos, call video sync for each) in internal/sync/channel_sync.go
- [ ] T045 [US2] Add progress tracking for bulk operations (current/total count) in internal/sync/channel_sync.go
- [ ] T046 [US2] Implement "channel" CLI command with cobra in cmd/youtube-sync/main.go
- [ ] T047 [US2] Add --limit, --since, --force flags to channel command
- [ ] T048 [US2] Add progress bar/indicator for channel sync operations
- [ ] T049 [US2] Add handling for unavailable videos (continue sync, don't fail) (FR-013)
- [ ] T050 [US2] Add structured logging with summary (success/failed counts) for channel syncs
- [ ] T051 [US2] Ensure channel info is logged but NOT stored in presenters table (FR-018, FR-020)

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently
**Manual Test**: Follow quickstart.md Test Case 2.1-2.5 to validate bulk channel sync

---

## Phase 5: User Story 4 - Playlist Sync (Priority: P3)

**Goal**: Sync YouTube playlists with all videos and maintain playlist order

**Independent Test**: Run `youtube-sync playlist PLrAXtmErZgOeiKm4sgNOknGvNjby9efdf` and verify playlist + all videos + playlist_items appear in database

**Note**: Implementing US4 before US3 because US3 (scheduling) depends on US1, US2, and US4 sync capabilities

### Tests for User Story 4 (Test-First Development)

- [ ] T052 [P] [US4] Write unit test for playlist ID extraction from URLs in tests/unit/parser/youtube_url_test.go (MUST FAIL)
- [ ] T053 [P] [US4] Write unit test for playlist metadata fetching in tests/unit/youtube/playlist_test.go (MUST FAIL with mocked API)
- [ ] T054 [P] [US4] Write contract test for playlists table operations in tests/contract/schema_test.go (MUST FAIL)
- [ ] T055 [P] [US4] Write contract test for playlist_items table operations in tests/contract/schema_test.go (MUST FAIL)
- [ ] T056 [US4] Write unit test for playlist sync logic in tests/unit/sync/playlist_sync_test.go (MUST FAIL)
- [ ] T057 [US4] Write integration test for end-to-end playlist sync in tests/integration/playlist_sync_test.go (MUST FAIL)

### Implementation for User Story 4

- [ ] T058 [P] [US4] Extend URL parser to handle playlist URLs in internal/parser/youtube_url.go
- [ ] T059 [P] [US4] Implement slug generation from playlist title in internal/parser/youtube_url.go
- [ ] T060 [P] [US4] Implement playlists table operations (upsert) in internal/database/playlists.go
- [ ] T061 [P] [US4] Implement playlist_items table operations (upsert with sort_order) in internal/database/playlist_items.go
- [ ] T062 [US4] Implement playlist metadata fetcher in internal/youtube/playlist.go
- [ ] T063 [US4] Implement playlist items fetcher with pagination in internal/youtube/playlist.go
- [ ] T064 [US4] Implement playlist sync logic (playlist + videos + playlist_items) in internal/sync/playlist_sync.go
- [ ] T065 [US4] Add database transaction handling for atomicity (playlist + items)
- [ ] T066 [US4] Add ID lookup logic (playlist_id and episode_id FKs) in internal/sync/playlist_sync.go
- [ ] T067 [US4] Implement "playlist" CLI command with cobra in cmd/youtube-sync/main.go
- [ ] T068 [US4] Add --force flag to playlist command
- [ ] T069 [US4] Add handling for duplicate videos in playlists (unique constraint)
- [ ] T070 [US4] Add structured logging for playlist sync operations

**Checkpoint**: Playlist sync should work independently
**Manual Test**: Follow quickstart.md Test Case 3.1-3.3 to validate playlist sync

---

## Phase 6: User Story 3 - Scheduled Automatic Sync (Priority: P3)

**Goal**: Configure scheduled syncs that run automatically at specified times

**Independent Test**: Configure a 1-minute schedule, wait, and verify database was updated automatically

### Tests for User Story 3 (Test-First Development)

- [ ] T071 [P] [US3] Write unit test for cron schedule parsing in tests/unit/sync/scheduler_test.go (MUST FAIL)
- [ ] T072 [P] [US3] Write unit test for scheduled job tracking in tests/unit/sync/scheduler_test.go (MUST FAIL)
- [ ] T073 [US3] Write integration test for scheduled sync execution in tests/integration/scheduler_test.go (MUST FAIL)

### Implementation for User Story 3

- [ ] T074 [US3] Implement scheduler with cron library in internal/sync/scheduler.go
- [ ] T075 [US3] Add configuration for scheduled jobs (cron expressions) in pkg/config/config.go
- [ ] T076 [US3] Implement scheduled job registry (track which channels/playlists to sync)
- [ ] T077 [US3] Implement "schedule" CLI subcommands (add, list, remove, start, stop) in cmd/youtube-sync/main.go
- [ ] T078 [US3] Add daemon mode for running scheduler continuously
- [ ] T079 [US3] Add structured logging for scheduled sync results (success/failure counts, timestamps)
- [ ] T080 [US3] Add file-based error logging for failed scheduled syncs (FR-010, US3 Scenario 3)
- [ ] T081 [US3] Add graceful shutdown handling for scheduler daemon

**Checkpoint**: All user stories should now be independently functional
**Manual Test**: Follow quickstart.md Scenario 3 to validate scheduled syncs

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories and production readiness

- [ ] T082 [P] Add version command showing build info in cmd/youtube-sync/main.go
- [ ] T083 [P] Add global flags (--config, --verbose, --json, --quiet, --dry-run) to root command
- [ ] T084 [P] Implement JSON output mode for all commands (parseable for scripts)
- [ ] T085 [P] Add POSIX-compliant exit codes (0=success, 1=error, 2=usage, 3=config, 4=API, 5=DB)
- [ ] T086 [P] Add actionable error messages with suggestions for common failures
- [ ] T087 [P] Add comprehensive help text for all commands with examples
- [ ] T088 [P] Add configuration file support (~/.youtube-sync/config.yaml)
- [ ] T089 [P] Add environment variable support for all config options
- [ ] T090 [P] Create comprehensive unit tests for untested edge cases in tests/unit/
- [ ] T091 [P] Add performance benchmarks for sync operations (target: <5s per video)
- [ ] T092 [P] Add code comments and godoc for all public APIs
- [ ] T093 [P] Run gofmt, golint, go vet and fix all issues
- [ ] T094 [P] Update README.md with complete usage examples and troubleshooting
- [ ] T095 Run full quickstart.md validation suite (all test scenarios)
- [ ] T096 Verify test coverage meets >80% target with `go test -cover`
- [ ] T097 [P] Create Makefile with build, test, lint, install targets
- [ ] T098 [P] Add build scripts for cross-platform binaries (Linux, macOS, Windows)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion
- **User Story 2 (Phase 4)**: Depends on Foundational phase completion + can reuse video sync from US1
- **User Story 4 (Phase 5)**: Depends on Foundational phase completion + can reuse video sync from US1
- **User Story 3 (Phase 6)**: Depends on US1, US2, US4 being complete (schedules need sync capabilities)
- **Polish (Phase 7)**: Depends on all desired user stories being complete

### User Story Dependencies

```
Foundational (Phase 2)
    ├─→ US1: Manual Video Sync (P1) ✓ Independent
    ├─→ US2: Bulk Channel Sync (P2) [uses video sync from US1]
    └─→ US4: Playlist Sync (P3) [uses video sync from US1]
         └─→ US3: Scheduled Sync (P3) [schedules US1, US2, US4]
```

- **User Story 1 (P1)**: Fully independent - can start immediately after Foundational
- **User Story 2 (P2)**: Reuses video sync from US1 but independently testable
- **User Story 4 (P3)**: Reuses video sync from US1 but independently testable
- **User Story 3 (P3)**: Schedules the operations from US1, US2, US4

### Within Each User Story

1. **Tests FIRST** (RED phase):
   - Write all tests for the story
   - Run tests - they MUST FAIL
   - Get approval on test scenarios

2. **Implementation** (GREEN phase):
   - Implement parsers/models (can be parallel if marked [P])
   - Implement services/sync logic
   - Implement CLI commands
   - Run tests - they should now PASS

3. **Refactor** (REFACTOR phase):
   - Improve code while keeping tests green
   - Add logging, error handling
   - Add CLI flags and output formatting

### Parallel Opportunities

**Phase 1 - Setup**: All tasks marked [P] (T003, T004, T005, T006, T007)

**Phase 2 - Foundational**: Tests can be written in parallel, then implementations:
- T008, T010, T012, T014, T016, T018 (all tests, parallel)
- T009, T011, T013, T015, T017, T019 (implementations after tests)

**Phase 3 - User Story 1**: Tests in parallel (T021-T024), then implementation

**Phase 4 - User Story 2**: Tests in parallel (T036-T039), then implementation

**Phase 5 - User Story 4**: Tests in parallel (T052-T055), then implementation

**Phase 7 - Polish**: Most tasks marked [P] can run in parallel

**Cross-Story Parallelism**:
- With multiple developers, US1, US2, US4 can be developed in parallel after Foundational is complete
- US3 must wait for US1, US2, US4 to provide sync capabilities to schedule

---

## Parallel Example: User Story 1 Tests

```bash
# Launch all tests for User Story 1 together (they MUST fail):
Task T021: "Write unit test for YouTube video ID extraction"
Task T022: "Write unit test for YouTube URL validation"
Task T023: "Write unit test for video metadata fetching"
Task T024: "Write unit test for video metadata mapping"
Task T025: "Write integration test for end-to-end video sync"

# After tests fail, launch parallel implementation:
Task T026: "Implement YouTube URL parser"
Task T027: "Implement URL validator"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T007)
2. Complete Phase 2: Foundational (T008-T020) **CRITICAL - blocks all stories**
3. Complete Phase 3: User Story 1 (T021-T035)
4. **STOP and VALIDATE**: Run quickstart.md Test Cases 1.1-1.6
5. Deploy/demo if ready (can sync single videos)

**MVP Deliverable**: A working CLI that can sync individual YouTube videos to the database

### Incremental Delivery

1. **Foundation** (Phase 1 + 2): Project setup + core infrastructure → Can start building features
2. **MVP Release** (+ Phase 3): Add User Story 1 → Test independently → **v0.1.0 Release**
   - Value: Manual video sync capability
3. **Enhanced Release** (+ Phase 4): Add User Story 2 → Test independently → **v0.2.0 Release**
   - Value: Bulk channel sync capability
4. **Full Release** (+ Phase 5): Add User Story 4 → Test independently → **v0.3.0 Release**
   - Value: Playlist sync capability
5. **Automated Release** (+ Phase 6): Add User Story 3 → Test independently → **v1.0.0 Release**
   - Value: Scheduled automatic syncs
6. **Polished Release** (+ Phase 7): Polish & production hardening → **v1.1.0 Release**

Each release adds value without breaking previous functionality.

### Parallel Team Strategy

With multiple developers (after Foundational Phase 2 completes):

```
Timeline:
Week 1-2: ALL developers → Phase 1 (Setup) + Phase 2 (Foundational)

Then in parallel:
Week 3:   Developer A → US1 (P1 - Manual Video Sync)
          Developer B → US2 (P2 - Channel Sync)
          Developer C → US4 (P3 - Playlist Sync)

Week 4:   ALL developers → US3 (P3 - Scheduling) after US1/US2/US4 complete

Week 5:   ALL developers → Phase 7 (Polish)
```

Stories complete and integrate independently.

---

## Test-First Development (TDD) Enforcement

**Constitution Principle II Requires**: Test-First Development (Red-Green-Refactor)

### RED Phase (Write Failing Tests)
- Every implementation task (T026+) has corresponding test tasks (T021-T025)
- Tests MUST be written first
- Tests MUST fail initially
- Get approval on test scenarios before proceeding

### GREEN Phase (Make Tests Pass)
- Implement minimum code to make tests pass
- No gold-plating or premature optimization
- Focus on meeting test requirements

### REFACTOR Phase (Improve While Green)
- Refactor code for clarity and performance
- Tests must stay green throughout
- Add logging, error handling, documentation

### Validation Checkpoints

**After Foundational (Phase 2)**:
```bash
go test ./internal/database/... -v
go test ./internal/logger/... -v
go test ./internal/youtube/... -v
# All tests should PASS
```

**After User Story 1 (Phase 3)**:
```bash
go test ./tests/unit/parser/... -v
go test ./tests/unit/sync/... -v
go test ./tests/integration/video_sync_test.go -v
# All tests should PASS
# Manual validation: quickstart.md Test Cases 1.1-1.6
```

**After User Story 2 (Phase 4)**:
```bash
go test ./tests/unit/youtube/channel_test.go -v
go test ./tests/integration/channel_sync_test.go -v
# All tests should PASS
# Manual validation: quickstart.md Test Cases 2.1-2.5
```

**Final Validation (Phase 7)**:
```bash
go test ./... -cover
# Target: >80% coverage
# Run full quickstart.md validation suite
```

---

## Notes

- **[P] tasks** = different files, no dependencies, can run in parallel
- **[Story] label** maps task to specific user story for traceability
- **Test-First** is NON-NEGOTIABLE per Constitution Principle II
- Each user story should be independently completable and testable
- Verify tests FAIL before implementing (RED phase)
- Commit after each logical group of tasks
- Stop at any checkpoint to validate story independently
- **Database constraint**: Do NOT modify presenters, organizations, or their junction tables (FR-018)
- **Channel info**: Log channel metadata but do NOT store in database (FR-020)

## Total Task Count

- **Phase 1 (Setup)**: 7 tasks
- **Phase 2 (Foundational)**: 13 tasks
- **Phase 3 (User Story 1)**: 15 tasks (5 tests + 10 implementation)
- **Phase 4 (User Story 2)**: 16 tasks (5 tests + 11 implementation)
- **Phase 5 (User Story 4)**: 19 tasks (6 tests + 13 implementation)
- **Phase 6 (User Story 3)**: 11 tasks (3 tests + 8 implementation)
- **Phase 7 (Polish)**: 17 tasks

**Total: 98 tasks**

**Test tasks**: 32 (33% of total - enforces TDD)
**Implementation tasks**: 66

**Parallel opportunities**: 40+ tasks marked [P] across all phases

**MVP scope**: Phases 1-3 (35 tasks) delivers working single video sync
