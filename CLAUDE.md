# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**ESG Tube Ingestor** is a video ingestion system built using the SpecKit framework - a specification-driven development methodology that emphasizes test-first development, feature isolation, and independent user story delivery.

This is an early-stage project. The codebase currently contains only the SpecKit framework scaffolding with no application code yet implemented.

## Architecture: SpecKit Framework

This project uses **SpecKit**, a structured specification-driven development workflow. All feature development follows a strict lifecycle:

```
/speckit.specify → /speckit.plan → /speckit.tasks → /speckit.implement
```

### Feature Development Lifecycle

1. **Specification Phase** (`/speckit.specify <description>`)
   - Creates a feature specification in `specs/[###-feature-name]/spec.md`
   - Generates user stories with priorities (P1, P2, P3) and acceptance criteria
   - Each user story MUST be independently testable

2. **Planning Phase** (`/speckit.plan`)
   - Generates implementation plan in `specs/[###-feature-name]/plan.md`
   - **Runs Constitution Check** - validates against project principles (see below)
   - Phase 0: Creates `research.md` (resolves unknowns, evaluates libraries)
   - Phase 1: Creates `data-model.md`, `contracts/`, `quickstart.md`
   - Updates agent context in `.specify/templates/agent-file-template.md`

3. **Task Generation** (`/speckit.tasks`)
   - Generates dependency-ordered tasks in `specs/[###-feature-name]/tasks.md`
   - Tasks are organized by user story for parallel/incremental delivery
   - Includes explicit parallel execution markers `[P]` for concurrent work

4. **Implementation** (`/speckit.implement`)
   - Executes tasks from `tasks.md` in dependency order
   - Enforces test-first development (write failing tests before implementation)
   - Checks checklists (if present) before proceeding
   - Marks tasks complete in tasks.md as work progresses

### Constitution-Driven Development

**CRITICAL**: All development is governed by `.specify/memory/constitution.md` (v1.0.0). The constitution defines four NON-NEGOTIABLE principles:

1. **Code Quality & Maintainability**
   - Type hints/annotations required
   - Single Responsibility Principle
   - Mandatory code reviews and linting

2. **Testing Discipline (NON-NEGOTIABLE)**
   - Test-first development (Red-Green-Refactor)
   - 80%+ unit test coverage
   - Integration tests for all external dependencies
   - Contract tests for all public interfaces
   - All tests MUST pass in CI before merge

3. **User Experience Consistency**
   - POSIX CLI conventions, RESTful API design
   - Actionable error messages with context
   - Structured JSON logging with correlation IDs

4. **Performance & Reliability**
   - API response time <200ms (p95)
   - Idempotent operations, circuit breakers
   - Prometheus metrics, distributed tracing
   - Health checks at `/health` and `/ready`

**When implementing features**: The `/speckit.plan` command automatically validates your design against these principles. Any violations MUST be justified in the "Complexity Tracking" section of `plan.md`.

## Directory Structure

```
.specify/
├── memory/
│   └── constitution.md          # Project governance and principles (v1.0.0)
├── templates/
│   ├── spec-template.md         # Feature specification template
│   ├── plan-template.md         # Implementation plan template
│   ├── tasks-template.md        # Task list template
│   ├── checklist-template.md    # Quality checklist template
│   └── agent-file-template.md   # AI agent context (auto-generated from plans)
└── scripts/bash/
    ├── setup-plan.sh            # Initialize planning phase
    ├── check-prerequisites.sh   # Validate design artifacts exist
    └── common.sh                # Shared utility functions

.claude/commands/                # Custom slash commands
├── speckit.specify.md           # Create feature spec
├── speckit.plan.md              # Generate implementation plan
├── speckit.tasks.md             # Generate task list
├── speckit.implement.md         # Execute implementation
├── speckit.analyze.md           # Cross-artifact consistency check
├── speckit.clarify.md           # Identify underspecified areas
├── speckit.checklist.md         # Generate custom checklists
└── speckit.constitution.md      # Update project constitution

specs/                           # Feature specifications (created per feature)
├── [###-feature-name]/
│   ├── spec.md                  # User stories and requirements
│   ├── plan.md                  # Implementation plan with constitution check
│   ├── tasks.md                 # Dependency-ordered task list
│   ├── research.md              # Technology/library research
│   ├── data-model.md            # Entity definitions
│   ├── quickstart.md            # Manual test scenarios
│   ├── contracts/               # API/interface contracts
│   └── checklists/              # Optional quality checklists

src/                             # Application code (not yet created)
tests/                           # Test suites (not yet created)
├── unit/
├── integration/
└── contract/
```

## Key Workflow Concepts

### Independent User Stories

Every user story in `spec.md` MUST be:
- Assigned a priority (P1 = MVP, P2, P3, etc.)
- Independently implementable (can work on its own)
- Independently testable (has clear test criteria)
- Delivers standalone value when completed

This enables:
- Parallel development by multiple developers
- Incremental delivery (deploy P1, then add P2, etc.)
- MVP validation (implement only P1 first)

### Parallel Task Execution

Tasks marked with `[P]` in `tasks.md` can run in parallel because they:
- Modify different files
- Have no dependencies on each other
- Are within the same phase

Example:
```markdown
- [ ] T012 [P] [US1] Create Entity1 model in src/models/entity1.py
- [ ] T013 [P] [US1] Create Entity2 model in src/models/entity2.py
```

### Test-First Development (Red-Green-Refactor)

**Enforced by Constitution Principle II**:

1. **RED**: Write test(s) that define expected behavior → Tests FAIL
2. Get approval on test scenarios
3. **GREEN**: Write minimum code to make tests pass
4. **REFACTOR**: Improve code while keeping tests green

The `/speckit.implement` command enforces this by requiring tests to be written before implementation tasks.

### Constitution Check Gate

During `/speckit.plan`, the system validates your design against constitution principles:

- **Blocking errors**: Violations of NON-NEGOTIABLE principles (e.g., no tests planned)
- **Warnings**: Complexity that needs justification (e.g., adding 4th project when 3 is the standard)

If violations exist, you MUST document in plan.md:
```markdown
| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| No integration tests | ... | ... |
```

## Slash Commands Reference

| Command | Purpose | Output |
|---------|---------|--------|
| `/speckit.specify` | Create feature spec from description | `specs/###-name/spec.md` |
| `/speckit.plan` | Generate implementation plan | `specs/###-name/plan.md` + artifacts |
| `/speckit.tasks` | Generate task list from design | `specs/###-name/tasks.md` |
| `/speckit.implement` | Execute tasks (test-first) | Code + tests |
| `/speckit.analyze` | Validate cross-artifact consistency | Analysis report |
| `/speckit.clarify` | Find underspecified areas | Updated spec.md |
| `/speckit.checklist` | Generate custom quality checklist | `specs/###-name/checklists/*.md` |
| `/speckit.constitution` | Update project principles | Updated constitution.md |

## Important Notes

### When Working on Features

1. **Always start with `/speckit.specify`** - Don't write code before specification
2. **Follow the lifecycle** - Don't skip steps (specify → plan → tasks → implement)
3. **Respect Constitution Check results** - Address violations or justify them
4. **Write tests first** - The framework enforces TDD via task ordering
5. **Check task dependencies** - Follow the execution order in tasks.md
6. **Mark tasks complete** - Update tasks.md checkboxes as you progress

### When Reading the Codebase

- Check `specs/` first to understand what features exist
- Read `constitution.md` to understand project principles
- Feature documentation lives in `specs/[feature]/` not in code comments
- Look for `[P]` markers in tasks.md to identify parallelizable work

### Branch Naming Convention

Feature branches follow the pattern: `[###-feature-name]` where `###` is a sequential number generated by `/speckit.specify`.

Example: `001-video-metadata-extraction`

### Commit Message Format

Follow Conventional Commits:
- `feat: add video thumbnail extraction`
- `fix: handle null metadata gracefully`
- `docs: update API documentation for /ingest endpoint`
- `test: add integration tests for S3 upload`
- `refactor: simplify metadata parser logic`

See constitution.md § Development Workflow → Commit Standards for full requirements.

## When There's No Application Code Yet

This project is in its initial state with only SpecKit scaffolding. To begin development:

1. Define your first feature: `/speckit.specify <feature description>`
2. Follow the workflow through to implementation
3. The first feature will establish the project structure (src/, tests/, etc.)

The templates will guide you through:
- Choosing the tech stack (language, frameworks)
- Defining project structure (single/web/mobile)
- Setting up testing infrastructure
- Implementing the MVP (P1 user stories)
