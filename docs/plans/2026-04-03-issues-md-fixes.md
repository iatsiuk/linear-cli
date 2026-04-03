# Fix all issues from issues.md (TDD)

## Overview
- Implement 5 feature gaps described in `issues.md`
- #1: `--label` filter for `issue list`
- #2: `--project` filter for `issue list`
- #3: `view issues` subcommand (list issues belonging to a custom view)
- #4: `view show` slug support (help text + docs)
- #5: `project issues` subcommand (list issues belonging to a project)
- All changes follow existing patterns: Cobra commands, filter maps, resolver functions, httptest mocks

## Context (from discovery)
- Commands: `internal/cmd/issue.go`, `internal/cmd/view.go`, `internal/cmd/project.go`
- Filters: `internal/filter/builder.go` (AddFlags/BuildFromFlags pattern)
- Resolvers: `internal/api/resolver.go` (ResolveLabelID, ResolveProjectID exist)
- Queries: `internal/query/issue.go` (IssueListQuery), `internal/query/custom_view.go`, `internal/query/project.go`
- Models: `internal/model/issue.go`, `internal/model/custom_view.go`, `internal/model/project.go`
- Tests: `internal/cmd/issue_test.go`, `internal/cmd/view_test.go`, `internal/cmd/project_test.go`
- Output: `internal/output/formatter.go` (TableFormatter/JSONFormatter)
- Test helpers: `setupIssueTest`, `newIssueTestServer`, `makeIssue`, `issueListResponse`

## Development Approach
- **Testing approach**: TDD (tests first, then implementation)
- Each task: write failing tests first, then implement code to make them pass
- **CRITICAL: every task MUST include new/updated tests** for code changes
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run `make build` after each task (includes linter)

## Testing Strategy
- **Unit tests**: table-driven, httptest mock servers, capture GraphQL variables to verify filter construction
- No e2e tests in this project
- Test both table and JSON output modes where relevant
- Test mutual exclusivity constraints (e.g. --project vs --no-project)

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Add `--label` flag to `issue list`

**Tests first:**
- [x] write test `TestIssueListCommand_LabelFilter`: mock server captures GraphQL variables, verify `filter.labels.some.name.eq` is set correctly when `--label "bug"` is passed
- [x] write test `TestIssueListCommand_LabelFilter_JSON`: verify JSON output works with label filter
- [x] write test for multiple labels if StringSlice is used (or skip if single label for now)

**Implementation:**
- [x] add `--label` string flag in `newIssueListCommand()` (`internal/cmd/issue.go`)
- [x] in `runIssueList()`: read flag, build filter condition `"labels": {"some": {"name": {"eq": value}}}`
- [x] run tests - must pass
- [x] run `make build` - must pass

### Task 2: Add `--project` flag to `issue list`

**Tests first:**
- [x] write test `TestIssueListCommand_ProjectFilter`: mock server handles ResolveProjectID query + issue list query, verify `filter.project.id.eq` is set with resolved UUID
- [x] write test `TestIssueListCommand_ProjectFilter_MutualExclusive`: verify error when both `--project` and `--no-project` are used (without `--or`)

**Implementation:**
- [x] add `--project` string flag in `newIssueListCommand()` (`internal/cmd/issue.go`)
- [x] in `runIssueList()`: read flag, resolve via `api.ResolveProjectID()`, build filter `"project": {"id": {"eq": uuid}}`
- [x] add mutual exclusivity check with `--no-project` (when not in `--or` mode)
- [x] run tests - must pass
- [x] run `make build` - must pass

### Task 3: Add `view issues` tests
- [x] write test `TestViewIssuesCommand_TableOutput`: mock server returns customView with issues connection, verify table output contains issue identifiers
- [x] write test `TestViewIssuesCommand_JSONOutput`: verify JSON output with `--json`
- [x] write test `TestViewIssuesCommand_WithLimit`: verify `first` variable is passed correctly
- [x] write test `TestViewIssuesCommand_MissingArg`: verify error when no ID provided

### Task 4: Add `view issues` implementation
- [x] add `ViewIssuesQuery` in `internal/query/custom_view.go`: `customView(id) { issues(first, orderBy, includeArchived) { nodes { ...issueListFields } pageInfo } }`
- [x] add response model if needed (or reuse existing issue connection types)
- [x] add `newViewIssuesCommand()` in `internal/cmd/view.go` with flags: `--limit`, `--order-by`, `--include-archived`
- [x] implement `runViewIssues()`: query API, format output with `IssueRow` / `printIssueRow`
- [x] register subcommand in `newViewCommand()`
- [x] run tests - must pass
- [x] run `make build` - must pass

### Task 5: Update `view show` for slug support
- [x] write test verifying command `Use` field contains `<id-or-slug>` instead of `<id>`
- [x] update `Use` from `"show <id>"` to `"show <id-or-slug>"` in `newViewShowCommand()`
- [x] update `Short`/`Long` description to mention slug support: "Accepts UUID or URL slug"
- [x] run tests - must pass
- [x] run `make build` - must pass

### Task 6: Add `project issues` tests
- [ ] write test `TestProjectIssuesCommand_TableOutput`: mock server handles ResolveProjectID + project issues query, verify table output
- [ ] write test `TestProjectIssuesCommand_JSONOutput`: verify JSON output
- [ ] write test `TestProjectIssuesCommand_WithLimit`: verify `first` variable
- [ ] write test `TestProjectIssuesCommand_ByName`: verify project name is resolved to UUID via ResolveProjectID

### Task 7: Add `project issues` implementation
- [ ] add `ProjectIssuesQuery` in `internal/query/project.go`: `project(id) { issues(first, orderBy, includeArchived) { nodes { ...issueListFields } pageInfo } }`
- [ ] add `newProjectIssuesCommand()` in `internal/cmd/project.go` with flags: `--limit`, `--order-by`, `--include-archived`
- [ ] implement `runProjectIssues()`: resolve project ID, query API, format with `IssueRow` / `printIssueRow`
- [ ] register subcommand in `newProjectCommand()`
- [ ] run tests - must pass
- [ ] run `make build` - must pass

### Task 8: Verify acceptance criteria
- [ ] verify all 5 issues from issues.md are addressed
- [ ] verify edge cases: mutual exclusivity, missing args, empty results
- [ ] run full test suite: `go test -race ./...`
- [ ] run linter: `golangci-lint run`
- [ ] verify test coverage meets project standard
- [ ] run `make build` - final verification

### Task 9: [Final] Update documentation
- [ ] update README.md if new flags/commands need to be documented
- [ ] update issues.md to mark resolved items (or remove file)

## Technical Details

### Filter conditions format
- Label: `"labels": {"some": {"name": {"eq": "bug"}}}`
- Project: `"project": {"id": {"eq": "uuid-here"}}`

### GraphQL queries to add

**ViewIssuesQuery:**
```graphql
query ViewIssues($id: String!, $first: Int, $orderBy: PaginationOrderBy, $includeArchived: Boolean) {
  customView(id: $id) {
    issues(first: $first, orderBy: $orderBy, includeArchived: $includeArchived) {
      nodes { ...issueListFields }
      pageInfo { hasNextPage endCursor }
    }
  }
}
```

**ProjectIssuesQuery:**
```graphql
query ProjectIssues($id: String!, $first: Int, $orderBy: PaginationOrderBy, $includeArchived: Boolean) {
  project(id: $id) {
    issues(first: $first, orderBy: $orderBy, includeArchived: $includeArchived) {
      nodes { ...issueListFields }
      pageInfo { hasNextPage endCursor }
    }
  }
}
```

### Resolver usage
- `--label`: filter by name directly (no resolver needed, filter uses `name.eq`)
- `--project`: resolve via `api.ResolveProjectID(ctx, client, value)` then filter by `id.eq`
- `view issues`: pass ID/slug directly to query (API resolves)
- `project issues`: resolve via `api.ResolveProjectID(ctx, client, value)`

## Post-Completion

**Manual verification:**
- Test with real Linear API: `linear issue list --label "bug"`
- Test with real Linear API: `linear issue list --project "My Project"`
- Test with real Linear API: `linear view issues <slug>`
- Test with real Linear API: `linear project issues "My Project"`
- Verify `--help` output for new flags/commands
