# Issue Relations & Notifications

## Overview
- Implement issue relations management and notification commands
- Commands: `linear issue relation list/create/delete`, `linear issue branch`, `linear notification list/read/archive`
- Depends on Foundation and Issues plans

## Context
- IssueRelation: id, type, issue, relatedIssue
- IssueRelationType enum: `blocks`, `duplicate`, `related`, `similar`
- IssueRelation mutations: `issueRelationCreate`, `issueRelationUpdate`, `issueRelationDelete`
- `issueRelationDelete` returns `DeletePayload` (entityId, success, no entity object)
- `issueVcsBranchSearch(branchName)` - lookup issue by git branch name
- Notification: id, type, readAt, archivedAt, createdAt
- Notification mutations: `notificationMarkReadAll`, `notificationArchiveAll`, `notificationUpdate`, etc.
- **CRITICAL: verify all GraphQL types against `docs/schema.graphql`**

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

### Task 1: Issue Relation model and queries
- [ ] write tests for IssueRelation struct deserialization
- [ ] create `internal/model/relation.go`:
  - IssueRelation struct: ID, Type, Issue(Issue), RelatedIssue(Issue), CreatedAt
- [ ] create `internal/query/relation.go`:
  - RelationListQuery (via issue.relations + issue.inverseRelations)
  - RelationCreateMutation, RelationUpdateMutation, RelationDeleteMutation
- [ ] write tests for query construction
- [ ] run `make test` - must pass before next task

### Task 2: `linear issue relation` commands
- [ ] write tests for relation list: fetches relations for issue, shows type + direction
- [ ] write tests for relation create: required fields, correct mutation
- [ ] write tests for relation delete: sends delete, returns DeletePayload (entityId only)
- [ ] create `internal/cmd/issue_relation.go`:
  - `issue relation list <identifier>`: shows relations (blocks, blocked by, duplicates, related)
  - `issue relation create <identifier>`: flags `--related <identifier>`, `--type` (blocks|duplicate|related|similar)
  - `issue relation delete <relation-id>`: `--yes` to skip confirmation
  - table columns: Type | Direction | Related Issue | Title
- [ ] run `make test` - must pass before next task

### Task 3: `linear issue branch` command
- [ ] write tests for branch lookup: sends issueVcsBranchSearch, returns issue
- [ ] create `internal/cmd/issue_branch.go`:
  - `issue branch <branch-name>`: looks up issue by git branch name
  - `issue branch` (no args): uses current git branch from `git rev-parse --abbrev-ref HEAD`
  - displays issue details (same as `issue show`)
  - supports `--json`
- [ ] write tests for auto-detect current git branch
- [ ] write tests for no issue found (error handling)
- [ ] run `make test` - must pass before next task

### Task 4: Notification model and queries
- [ ] write tests for Notification struct deserialization
- [ ] create `internal/model/notification.go`:
  - Notification struct: ID, Type, ReadAt, ArchivedAt, CreatedAt
- [ ] create `internal/query/notification.go`:
  - NotificationListQuery, NotificationUpdateMutation
  - NotificationMarkReadAllMutation, NotificationArchiveAllMutation
- [ ] write tests for query construction
- [ ] run `make test` - must pass before next task

### Task 5: `linear notification` commands
- [ ] write tests for notification list: fetches notifications, table output
- [ ] write tests for notification read/archive: sends correct mutations
- [ ] create `internal/cmd/notification.go`:
  - `notification list`: flags `--unread`, `--limit`, `--json`
  - `notification read <id>`: marks single notification as read
  - `notification read --all`: marks all as read (`notificationMarkReadAll`)
  - `notification archive <id>`: archives single notification
  - `notification archive --all`: archives all (`notificationArchiveAll`)
  - table columns: Type | Created | Read
- [ ] run `make test` - must pass before next task

### Task 6: Verify acceptance criteria
- [ ] verify `linear issue relation list ENG-123` shows issue relations
- [ ] verify `linear issue relation create ENG-123 --related ENG-456 --type blocks` works
- [ ] verify `linear issue branch` auto-detects current git branch
- [ ] verify `linear issue branch feature/eng-123-fix` finds issue
- [ ] verify `linear notification list --unread` shows unread notifications
- [ ] verify `linear notification read --all` marks all as read
- [ ] run `make test` - full suite must pass
- [ ] run `make build` - lint + build must pass

## Technical Details

### IssueRelationType values
- `blocks` - this issue blocks the related issue
- `duplicate` - this issue is a duplicate of the related issue
- `related` - issues are related
- `similar` - issues are similar

### Relation direction
`issue.relations` = outgoing (this issue -> related)
`issue.inverseRelations` = incoming (related -> this issue)
Display both for complete picture.

### DeletePayload pattern
```graphql
mutation { issueRelationDelete(id: "...") { success entityId } }
```
Returns `entityId: String!` (no entity object).

### Task 7: [Final] Update documentation
- [ ] update README.md with relation, branch, and notification commands usage
- [ ] document all command flags and options

## Post-Completion
- Manual testing with real issue relations
- Test branch lookup with various branch naming conventions
