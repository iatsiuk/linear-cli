# Organization, Templates, Initiatives & Extras

## Overview
- Implement organization info, project updates/milestones, team membership, templates, initiatives, custom views, extended search
- Commands: `linear org`, `linear project update/milestone`, `linear team member`, `linear template`, `linear initiative`, `linear view`, extended `linear search`
- Depends on Foundation, Issues, Projects plans

## Context
- Organization: id, name, urlKey, logoUrl
- ProjectUpdate (status check-ins): id, body, health, user, project, createdAt. NOT the `projectUpdate` mutation.
- ProjectMilestone: id, name, description, targetDate, sortOrder
- TeamMembership: user + team association
- Template: id, name, type, templateData
- Initiative: replacement for deprecated Roadmaps
- Custom Views: saved filter/sort configurations
- `searchProjects(term)`, `searchDocuments(term)` - additional search types
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

### Task 1: `linear org` command
- [x] write tests for org command: calls organization query, displays info
- [x] create `internal/cmd/org.go`:
  - `org`: displays organization name, urlKey, logoUrl
  - supports `--json`
- [x] run `make test` - must pass before next task

### Task 2: Project Updates (status check-ins)
- [ ] write tests for ProjectUpdate struct deserialization
- [ ] write tests for query/mutation construction
- [ ] create `internal/model/project_update.go`:
  - ProjectUpdate struct: ID, Body, Health, User, Project, CreatedAt
- [ ] create `internal/query/project_update.go`:
  - ProjectUpdateListQuery, ProjectUpdateCreateMutation, ProjectUpdateArchiveMutation
- [ ] create `internal/cmd/project_update_cmd.go`:
  - `project update list <project-id>`: shows status check-ins
  - `project update create <project-id>`: flags `--body` (required), `--health` (onTrack|atRisk|offTrack)
  - `project update archive <id>`: archives check-in (replaces deprecated `projectUpdateDelete`)
  - table columns: Health | Author | Date | Body (truncated)
- [ ] run `make test` - must pass before next task

### Task 3: Project Milestones
- [ ] write tests for ProjectMilestone struct deserialization
- [ ] write tests for query/mutation construction
- [ ] create `internal/model/milestone.go`:
  - ProjectMilestone struct: ID, Name, Description, TargetDate, SortOrder
- [ ] create `internal/query/milestone.go`:
  - MilestoneListQuery, MilestoneCreateMutation, MilestoneUpdateMutation, MilestoneDeleteMutation
- [ ] create `internal/cmd/milestone.go`:
  - `project milestone list <project-id>`: shows milestones
  - `project milestone create <project-id>`: flags `--name` (required), `--description`, `--target-date`
  - `project milestone update <id>`: flags `--name`, `--description`, `--target-date`
  - `project milestone delete <id>`: `--yes` to skip confirmation
  - table columns: Name | Target Date | Description
- [ ] run `make test` - must pass before next task

### Task 4: Team Membership
- [ ] write tests for TeamMembership struct deserialization
- [ ] write tests for query/mutation construction
- [ ] create `internal/cmd/team_member.go`:
  - `team member list <team-key>`: shows team members
  - `team member add <team-key> <user>`: creates team membership
  - `team member remove <team-key> <user>`: deletes team membership
  - table columns: Name | Email | Role
- [ ] run `make test` - must pass before next task

### Task 5: Templates
- [ ] write tests for Template struct deserialization
- [ ] write tests for query construction
- [ ] create `internal/cmd/template.go`:
  - `template list`: shows all templates (root `templates` query returns plain list, not connection)
  - `template show <id>`: shows template details with templateData
  - supports `--json`
  - table columns: Name | Type
- [ ] run `make test` - must pass before next task

### Task 6: Initiatives
- [ ] write tests for Initiative struct deserialization
- [ ] write tests for query/mutation construction
- [ ] create `internal/cmd/initiative.go`:
  - `initiative list`: flags `--limit`, `--json`
  - `initiative show <id>`: full initiative details
  - `initiative create`: flags `--name` (required), `--description`
  - `initiative update <id>`: flags `--name`, `--description`
  - `initiative delete <id>`: `--yes` to skip confirmation
  - table columns: Name | Status | Description
- [ ] run `make test` - must pass before next task

### Task 7: Extended search (projects, documents)
- [ ] write tests for searchProjects query
- [ ] write tests for searchDocuments query
- [ ] extend `internal/cmd/search.go`:
  - `search <query>`: default searches issues (existing)
  - `search <query> --type project`: searches projects via `searchProjects(term)`
  - `search <query> --type document`: searches documents via `searchDocuments(term)`
  - `search <query> --type issue`: explicit issue search (default)
  - `IssueSearchResult`, `ProjectSearchResult`, `DocumentSearchResult` implement respective entity fields
- [ ] run `make test` - must pass before next task

### Task 8: Custom Views
- [ ] write tests for CustomView struct deserialization
- [ ] create `internal/cmd/view.go`:
  - `view list`: shows saved custom views
  - `view show <id>`: shows view details (filters, sorting)
  - supports `--json`
  - table columns: Name | Type | Shared
- [ ] run `make test` - must pass before next task

### Task 9: Verify acceptance criteria
- [ ] verify `linear org` shows organization info
- [ ] verify `linear project update list/create` works for status check-ins
- [ ] verify `linear project milestone list/create` works
- [ ] verify `linear team member list/add/remove` works
- [ ] verify `linear template list` shows templates
- [ ] verify `linear initiative list/create` works
- [ ] verify `linear search "query" --type project` searches projects
- [ ] verify `linear view list` shows custom views
- [ ] run `make test` - full suite must pass
- [ ] run `make build` - lint + build must pass

### Task 10: [Final] Update documentation
- [ ] update README.md with all new commands
- [ ] document all available commands and flags

## Technical Details

### ProjectUpdate vs projectUpdate mutation
`ProjectUpdate` is a status check-in entity (query: `projectUpdates`).
`projectUpdate` is the mutation to update a Project's fields.
These are completely different operations - naming is confusing.

### Templates query
Root `templates` returns `[Template!]!` (plain list, NOT a connection).
No pagination needed unlike most other queries.

### Initiative replaces Roadmap
Roadmaps are deprecated. Initiatives are the replacement.
Mutations: `initiativeCreate`, `initiativeUpdate`, `initiativeArchive`, `initiativeDelete`.

## Post-Completion
- Manual testing with real Linear workspace
- Verify all new commands in help output
