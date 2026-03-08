# Projects & Cycles: Project and Cycle Management

## Overview
- Implement CRUD for projects and read/manage commands for cycles
- Commands: `linear project list/show/create/update/delete`, `linear cycle list/show/active`
- Depends on Foundation plan

## Context
- Project key fields: id, name, description, color, icon, health, state, progress, startDate, targetDate, creator, teams, url
- Cycle key fields: id, name, number, description, startsAt, endsAt, isActive, isFuture, isPast, progress, team
- ProjectCreateInput: name(required), teamIds(required), description, color, icon, startDate, targetDate, statusId, leadId, memberIds, priority, labelIds
- CycleCreateInput: teamId(required), startsAt(required, DateTime), endsAt(required, DateTime), name, description
- **Project status**: `status` is a `ProjectStatus` object with `type: ProjectStatusType!`. Values: `backlog`, `planned`, `started`, `paused`, `completed`, `canceled`
- **Project health**: `ProjectUpdateHealthType` values: `onTrack`, `atRisk`, `offTrack`
- **projectArchive is deprecated** - use `projectDelete` instead

## Development Approach
- **Testing approach**: TDD (tests first, then implementation)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
- **CRITICAL: all tests must pass before starting next task** - no exceptions
- **CRITICAL: update this plan file when scope changes during implementation**
- **CRITICAL: verify all GraphQL types, field names, nullability, and input types against `docs/schema.graphql`**

## Testing Strategy
- **Unit tests**: required for every task (see Development Approach above)
- **E2E tests**: not applicable for this plan (CLI commands)

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix
- Update plan if implementation deviates from original scope
- Keep plan in sync with actual work done

## What Goes Where
- **Implementation Steps** (`[ ]` checkboxes): tasks achievable within this codebase - code changes, tests, documentation updates
- **Post-Completion** (no checkboxes): items requiring external action - manual testing, changes in consuming projects, deployment configs, third-party verifications

## Implementation Steps

### Task 1: Project and Cycle models
- [x] write tests for Project struct deserialization
- [x] write tests for Cycle struct deserialization
- [x] create `internal/model/project.go`:
  - Project struct with key fields + JSON tags
- [x] create `internal/model/cycle.go`:
  - Cycle struct with key fields + JSON tags
- [x] run `make test` - must pass before next task

### Task 2: Project GraphQL queries and mutations
- [x] write tests for project query/mutation string construction
- [x] create `internal/query/project.go`:
  - ProjectListQuery, ProjectGetQuery, ProjectCreateMutation, ProjectUpdateMutation, ProjectDeleteMutation
- [x] run `make test` - must pass before next task

### Task 3: `linear project list/show` commands
- [x] write tests for project list: fetches projects, table output, filters
- [x] write tests for project show: fetches by ID, detailed output
- [x] create `internal/cmd/project.go`:
  - `project list`: flags `--team`, `--status`, `--health`, `--limit`, `--json`, `--include-archived`, `--order-by`
  - `project show <id>`: full project details
  - table columns: Name | Status | Health | Progress | Target Date
- [x] run `make test` - must pass before next task

### Task 4: `linear project create/update/delete` commands
- [ ] write tests for project create: required flags, mutation variables
- [ ] write tests for project update: partial update, correct mutation
- [ ] write tests for project delete: archive mutation, confirmation
- [ ] create `internal/cmd/project_create.go`:
  - flags: `--name` (required), `--team` (required, repeatable), `--description`, `--color`, `--target-date`, `--start-date`
- [ ] create `internal/cmd/project_update.go`:
  - flags: `--name`, `--description`, `--state`, `--target-date`, `--start-date`, `--health`
- [ ] create `internal/cmd/project_delete.go`:
  - uses `projectDelete` mutation (not deprecated `projectArchive`)
  - `--yes` to skip confirmation
- [ ] run `make test` - must pass before next task

### Task 5: Cycle GraphQL queries
- [ ] write tests for cycle query string construction
- [ ] create `internal/query/cycle.go`:
  - CycleListQuery, CycleGetQuery
- [ ] run `make test` - must pass before next task

### Task 6: `linear cycle list/show/active` commands
- [ ] write tests for cycle list: fetches cycles for team, table output
- [ ] write tests for cycle show: fetches by ID, detailed output
- [ ] write tests for cycle active: shows active cycle for team
- [ ] create `internal/cmd/cycle.go`:
  - `cycle list`: flags `--team` (required), `--limit`, `--json`, `--include-archived`, `--order-by`
  - `cycle show <id>`: full cycle details
  - `cycle active --team <key>`: shows currently active cycle
  - table columns: Number | Name | Start | End | Progress | Status
- [ ] run `make test` - must pass before next task

### Task 7: Verify acceptance criteria
- [ ] verify all project commands work end-to-end
- [ ] verify all cycle commands work end-to-end
- [ ] verify `--json` output for all commands
- [ ] run `make test` - full suite must pass
- [ ] run `make build` - lint + build must pass

## Technical Details

### Table columns
**project list**: Name | Status | Health | Progress% | Target Date
**cycle list**: # | Name | Start | End | Progress% | Status (Active/Past/Future)

### Task 8: [Final] Update documentation
- [ ] update README.md with project and cycle commands usage
- [ ] document all command flags and options

## Post-Completion
- Manual testing with real Linear workspace
