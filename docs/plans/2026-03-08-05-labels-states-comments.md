# Labels, States & Comments

## Overview
- Implement management commands for issue labels, workflow states, and comments
- Commands: `linear label list/create/update`, `linear state list`, `linear comment list/create`
- Depends on Foundation plan and Issue models

## Context
- IssueLabel: id, name, color, description, isGroup, team(optional - workspace vs team labels), parent
- WorkflowState: id, name, color, description, position, type (backlog/unstarted/started/completed/canceled), team
- Comment: id, body, createdAt, updatedAt, user, issue, parent (threading)
- Labels and states are essential for issue filtering and updates
- Comments are threaded (parent/children)

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

### Task 1: Label, State, Comment models (extend existing)
- [ ] write tests for IssueLabel full struct deserialization
- [ ] write tests for WorkflowState full struct deserialization
- [ ] write tests for Comment struct deserialization (including nested user, parent)
- [ ] extend models in `internal/model/`:
  - IssueLabel full: ID, Name, Color, Description, IsGroup, Team, Parent, CreatedAt
  - WorkflowState full: ID, Name, Color, Description, Position, Type, Team, CreatedAt
  - Comment: ID, Body, CreatedAt, UpdatedAt, EditedAt, URL, User, Issue(identifier)
- [ ] run `make test` - must pass before next task

### Task 2: GraphQL queries for labels, states, comments
- [ ] write tests for query construction
- [ ] create `internal/query/label.go`: LabelListQuery, LabelCreateMutation, LabelUpdateMutation
- [ ] create `internal/query/state.go`: StateListQuery
- [ ] create `internal/query/comment.go`: CommentListQuery, CommentCreateMutation
- [ ] run `make test` - must pass before next task

### Task 3: `linear label list/create/update` commands
- [ ] write tests for label list: fetches labels, filters by team, table output, `--include-archived`
- [ ] write tests for label create: required flags, mutation
- [ ] write tests for label update: partial update
- [ ] create `internal/cmd/label.go`:
  - `label list`: flags `--team`, `--json`. Columns: Name | Color | Team | Group
  - `label create`: flags `--name` (required), `--color` (required), `--team`, `--description`
  - `label update <id>`: flags `--name`, `--color`, `--description`
- [ ] run `make test` - must pass before next task

### Task 4: `linear state list` command
- [ ] write tests for state list: fetches states for team, groups by type, table output
- [ ] create `internal/cmd/state.go`:
  - `state list`: flags `--team` (required), `--json`
  - table columns: Name | Type | Color | Position
  - group output by type (Triage, Backlog, Unstarted, Started, Completed, Canceled)
- [ ] run `make test` - must pass before next task

### Task 5: `linear comment list/create` commands
- [ ] write tests for comment list: fetches comments for issue, displays threaded
- [ ] write tests for comment create: sends mutation with correct input
- [ ] create `internal/cmd/comment.go`:
  - `comment list <issue-identifier>`: shows comments for an issue, flag `--json`
  - `comment create <issue-identifier>`: flags `--body` (required), `--parent` (for threading)
  - table columns: Author | Date | Body (truncated)
- [ ] run `make test` - must pass before next task

### Task 6: Verify acceptance criteria
- [ ] verify `linear label list` shows workspace and team labels
- [ ] verify `linear label create` creates label
- [ ] verify `linear state list --team X` shows grouped workflow states
- [ ] verify `linear comment list ABC-123` shows issue comments
- [ ] verify `linear comment create ABC-123 --body "text"` creates comment
- [ ] run `make test` - full suite must pass
- [ ] run `make build` - lint + build must pass

## Technical Details

### WorkflowState types (from linear-api.md - `type` is plain `String!`, not a GraphQL enum)
- `triage` - needs triage
- `backlog` - in backlog
- `unstarted` - not yet started
- `started` - in progress
- `completed` - done
- `canceled` - canceled

### Table columns
**label list**: Name | Color | Description | Team | Group
**state list**: Name | Type | Color
**comment list**: Author | Date | Body

### Task 7: [Final] Update documentation
- [ ] update README.md with label, state, and comment commands usage
- [ ] document all command flags and options

## Post-Completion
- Manual testing with real Linear workspace
