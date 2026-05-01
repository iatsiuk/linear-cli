# CLI input flexibility (TDD)

## Overview
- Eliminate three repeated input mistakes from LLM/script callers, found in an audit of 1551 Claude Code transcripts:
  - `comment create`/`comment update` lack `--body-file`; LLMs assume gh-cli pattern.
  - `issue create`/`issue update` lack `--description-file`; same root cause.
  - `issue list` lacks `--parent <ID-or-key>`; LLMs expect it because all sibling fields exist as flags.
  - `view show`/`view issues` accept only UUID/slug; LLMs expect name resolution like `project`/`label`/`team`/`user`/`state`.
- Two new resolver functions, two file-reading helpers, one new filter flag, two call-site updates in `view`.
- All changes follow existing patterns (resolver short-circuit on UUID, filter map construction, table-driven tests with httptest mocks).

## Context (from discovery)
- Commands: `internal/cmd/comment.go`, `internal/cmd/issue.go`, `internal/cmd/issue_create.go`, `internal/cmd/issue_update.go`, `internal/cmd/view.go`
- Resolvers: `internal/api/resolver.go` (existing patterns for team/label/user/state/project, all use `looksLikeUUID` short-circuit)
- Filters: `internal/filter/builder.go` (advanced filters); `runIssueList` in `internal/cmd/issue.go` for command-local filters
- Queries: `internal/query/issue.go`, `internal/query/comment.go`, `internal/query/custom_view.go`
- Tests: `internal/cmd/comment_test.go`, `internal/cmd/issue_test.go`, `internal/cmd/issue_create_test.go`, `internal/cmd/issue_update_test.go`, `internal/cmd/view_test.go`, plus resolver tests in `internal/api/`
- Test helpers: `setupIssueTest`, `newIssueTestServer`, `makeIssue` (used in prior plans), httptest mock pattern that handles multiple sequential GraphQL calls
- Schema: `IssueIDComparator.eq: ID` accepts UUID only — must resolve `ENG-727` to UUID via top-level `issue(id: String!)` query

## Development Approach
- **Testing approach**: TDD (tests first, then implementation)
- Each task: write failing tests first, then implement code to make them pass
- **CRITICAL: every task MUST include new/updated tests** for code changes
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run `make build` after each task (includes linter)

## Testing Strategy
- **Unit tests**: table-driven, httptest mock servers, capture GraphQL operation name and variables
- No e2e tests in this project
- Test both table and JSON output modes where relevant
- Test mutual exclusivity constraints (`--body` vs `--body-file`)
- Stdin tests use `cmd.SetIn(strings.NewReader(...))`
- File-input tests use `t.TempDir()` + `os.WriteFile`

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with `+` prefix
- Document issues/blockers with `!` prefix
- Update plan if implementation deviates from original scope

## What Goes Where
- **Implementation Steps** (`[ ]` checkboxes): code, tests, docs in this repo
- **Post-Completion** (no checkboxes): release/cask/manual smoke tests against live Linear API

## Implementation Steps

### Task 1: Add `ResolveIssueID` resolver

**Tests first:**
- [x] in `internal/api/resolver_test.go` (create if missing) write `TestResolveIssueID_UUID`: input is UUID, asserts no HTTP call made, returns input unchanged
- [x] write `TestResolveIssueID_Identifier`: input is `ENG-727`, mock server returns `{"data":{"issue":{"id":"<uuid>"}}}`, assert returned id matches and the GraphQL variable `id` equals `"ENG-727"`
- [x] write `TestResolveIssueID_NotFound`: mock returns `{"data":{"issue":null}}` or graphql error, assert error contains `issue "ENG-999" not found`

**Implementation:**
- [x] add `ResolveIssueID(ctx context.Context, c *Client, idOrKey string) (string, error)` to `internal/api/resolver.go`:
  - short-circuit on `looksLikeUUID(idOrKey)`
  - GraphQL: `query ResolveIssue($id: String!) { issue(id: $id) { id } }`
  - error format: `fmt.Errorf("resolve issue %q: %w", idOrKey, err)` for transport errors, `fmt.Errorf("issue %q not found", idOrKey)` for empty result
- [x] run tests - must pass before Task 2

### Task 2: Add `ResolveCustomViewID` resolver

**Tests first:**
- [x] in `internal/api/resolver_test.go` write `TestResolveCustomViewID_UUID`: input UUID, no HTTP call
- [x] write `TestResolveCustomViewID_Name`: input `"Without Estimates"`, mock returns `{"data":{"customViews":{"nodes":[{"id":"<uuid>"}]}}}`, assert variable `name` equals `"Without Estimates"` and returned id matches
- [x] write `TestResolveCustomViewID_NotFound`: mock returns empty `nodes`, assert error `custom view "X" not found`

**Implementation:**
- [x] add `ResolveCustomViewID(ctx context.Context, c *Client, nameOrID string) (string, error)` to `internal/api/resolver.go`:
  - short-circuit on `looksLikeUUID(nameOrID)`
  - GraphQL: `query ResolveCustomView($name: String!) { customViews(filter: { name: { eq: $name } }, first: 1) { nodes { id } } }`
  - first match wins (matches `ResolveProjectID`)
- [x] run tests - must pass before Task 3

### Task 3: Add `--body-file` to `comment create` with `readBody` helper

**Tests first:**
- [x] in `internal/cmd/comment_test.go` write `TestCommentCreate_BodyFile`: write content to `t.TempDir()` file, run `comment create ENG-1 --body-file <path>`, assert mutation input.body equals file content
- [x] write `TestCommentCreate_BodyFileStdin`: pass `--body-file -` and `cmd.SetIn(strings.NewReader("from stdin"))`, assert mutation input.body == `"from stdin"`
- [x] write `TestCommentCreate_BodyAndBodyFileMutuallyExclusive`: pass both, assert error from cobra mentions `mutually exclusive`
- [x] write `TestCommentCreate_NoBodyOrBodyFile`: pass neither, assert error from cobra mentions one of the flags is required
- [x] write `TestCommentCreate_BodyFileMissing`: pass non-existent path, assert error contains `read body file` and the path

**Implementation:**
- [x] add helper `readBody(cmd *cobra.Command) (string, error)` in `internal/cmd/comment.go`:
  - if `cmd.Flags().Changed("body-file")`:
    - read flag value `path`
    - if `path == "-"`: `io.ReadAll(cmd.InOrStdin())`
    - else: `os.ReadFile(path)`
    - return `string(data)` (no trim — preserve trailing newline as user wrote it)
  - else: return `cmd.Flags().GetString("body")`
- [x] in `newCommentCreateCommand`: register `f.String("body-file", "", "read comment body from file ('-' for stdin)")`; replace `MarkFlagRequired("body")` with `cmd.MarkFlagsMutuallyExclusive("body", "body-file")` and `cmd.MarkFlagsOneRequired("body", "body-file")`
- [x] in `runCommentCreate`: replace `body, _ := f.GetString("body")` with `body, err := readBody(cmd); if err != nil { return fmt.Errorf("read body file: %w", err) }`
- [x] run tests - must pass before Task 4

### Task 4: Add `--body-file` to `comment update`

**Tests first:**
- [x] in `internal/cmd/comment_test.go` write `TestCommentUpdate_BodyFile`: similar to create test, asserts input.body equals file content
- [x] write `TestCommentUpdate_BodyFileStdin`: stdin path
- [x] write `TestCommentUpdate_BodyAndBodyFileMutuallyExclusive`: both flags -> cobra error
- [x] write `TestCommentUpdate_NoBodyOrBodyFile`: neither flag -> cobra error (one required, since update needs new content)

**Implementation:**
- [x] in `newCommentUpdateCmd`: register `--body-file`; replace `MarkFlagRequired("body")` with mutual-exclusive + one-required pair
- [x] in `runCommentUpdate`: use `readBody(cmd)` helper
- [x] run tests - must pass before Task 5

### Task 5: Add `--description-file` to `issue create`

**Tests first:**
- [x] in `internal/cmd/issue_create_test.go` write `TestIssueCreate_DescriptionFile`: file path -> mutation input.description equals file content
- [x] write `TestIssueCreate_DescriptionFileStdin`: `--description-file -` reads stdin
- [x] write `TestIssueCreate_DescriptionAndDescriptionFileMutuallyExclusive`: both -> cobra error
- [x] write `TestIssueCreate_DescriptionFileMissing`: non-existent path -> clear error

**Implementation:**
- [x] add helper `readDescription(cmd *cobra.Command) (string, ok bool, err error)` to `internal/cmd/issue.go` (shared by create and update):
  - returns `(value, true, nil)` if either `--description` or `--description-file` was changed
  - returns `("", false, nil)` if neither was changed
  - file-reading semantics identical to `readBody`
- [x] in `internal/cmd/issue_create.go`:
  - register `f.String("description-file", "", "read issue description from file ('-' for stdin)")`
  - add `cmd.MarkFlagsMutuallyExclusive("description", "description-file")`
  - in `runIssueCreate`: call `desc, hasDesc, err := readDescription(cmd)`; if err return wrapped; if `hasDesc` set `input["description"] = desc`
- [x] run tests - must pass before Task 6

### Task 6: Add `--description-file` to `issue update`

**Tests first:**
- [x] in `internal/cmd/issue_update_test.go` write `TestIssueUpdate_DescriptionFile`: file -> mutation input.description equals content
- [x] write `TestIssueUpdate_DescriptionFileStdin`: stdin path
- [x] write `TestIssueUpdate_DescriptionAndDescriptionFileMutuallyExclusive`
- [x] write `TestIssueUpdate_NoDescriptionFlags`: neither flag -> input map does NOT contain `description` key (preserve existing partial-update semantics)

**Implementation:**
- [x] in `internal/cmd/issue_update.go`:
  - register `--description-file`
  - `cmd.MarkFlagsMutuallyExclusive("description", "description-file")`
  - replace `if f.Changed("description")` block with `if desc, ok, err := readDescription(cmd); err != nil { return fmt.Errorf("read description file: %w", err) } else if ok { input["description"] = desc }`
- [x] run tests - must pass before Task 7

### Task 7: Add `--parent` filter to `issue list`

**Tests first:**
- [x] in `internal/cmd/issue_test.go` write `TestIssueList_ParentFilter_Identifier`: pass `--parent ENG-727`, mock server first responds to ResolveIssue query (returns UUID), then to IssueListQuery; assert IssueListQuery variables include `filter.parent.id.eq` equal to UUID
- [x] write `TestIssueList_ParentFilter_UUID`: pass `--parent <uuid>`, assert no resolve call, filter contains UUID directly
- [x] write `TestIssueList_ParentFilter_NotFound`: ResolveIssue returns null/error -> command returns error containing `not found`
- [x] write `TestIssueList_ParentFilter_CombinedWithOther`: pass `--parent ENG-727 --state "In Progress"`, assert filter contains both `parent.id.eq` and `state.name.eq`

**Implementation:**
- [x] in `internal/cmd/issue.go` `newIssueListCommand`: register `f.String("parent", "", "filter by parent issue (UUID or identifier like ENG-727)")`
- [x] in `runIssueList`: after building base `issueFilter`, if `parent` flag was set:
  - call `parentID, err := api.ResolveIssueID(ctx, client, parentRaw)`; on error return wrapped
  - merge into existing filter: `issueFilter["parent"] = map[string]any{"id": map[string]any{"eq": parentID}}`
  - merge happens before advanced-filter merge so AND/OR logic stays intact
- [x] run tests - must pass before Task 8

### Task 8: Wire `ResolveCustomViewID` into `view show` and `view issues`

**Tests first:**
- [x] in `internal/cmd/view_test.go` write `TestViewShow_ByName`: pass `view show "Without Estimates"`, mock server first handles ResolveCustomView (returns UUID), then CustomViewShowQuery with that UUID; assert succeeds
- [x] write `TestViewShow_ByUUID`: pass UUID directly, assert no resolve call
- [x] write `TestViewIssues_ByName`: similar for `view issues`
- [x] write `TestViewIssues_NotFound`: resolve returns no nodes -> command error contains `not found`

**Implementation:**
- [x] in `internal/cmd/view.go` `runViewShow`: replace direct `args[0]` use with `id, err := api.ResolveCustomViewID(ctx, client, args[0])`
- [x] in `runViewIssues`: same — resolve first, then pass UUID to `ViewIssuesQuery`
- [x] update help text on `view show` and `view issues` to mention "name" in addition to UUID/slug
- [x] run tests - must pass before Task 9

### Task 9: Update README

- [x] document `--body-file` flag in `comment create` and `comment update` sections (with `-` for stdin example)
- [x] document `--description-file` flag in `issue create` and `issue update` sections
- [x] document `--parent` flag in `issue list` filter list
- [x] document name acceptance for `view show` and `view issues`
- [x] no test changes for this task

### Task 10: Verify acceptance criteria
- [x] `make build` passes (linter + go build)
- [x] `go test -race ./...` passes
- [x] `./linear-cli comment create --help` shows `--body-file string`
- [x] `./linear-cli comment update --help` shows `--body-file string`
- [x] `./linear-cli issue create --help` shows `--description-file string`
- [x] `./linear-cli issue update --help` shows `--description-file string`
- [x] `./linear-cli issue list --help` shows `--parent string`
- [x] all new tests included in coverage report
- [x] no regression in existing `comment_test.go`, `issue_test.go`, `view_test.go`

## Technical Details

### `ResolveIssueID` query
```graphql
query ResolveIssue($id: String!) {
  issue(id: $id) { id }
}
```
Linear API accepts both UUID and identifier (e.g. `ENG-727`) in the `id` argument of the top-level `issue` field; we leverage this rather than parsing identifier locally.

### `ResolveCustomViewID` query
```graphql
query ResolveCustomView($name: String!) {
  customViews(filter: { name: { eq: $name } }, first: 1) {
    nodes { id }
  }
}
```

### `--parent` filter shape
After resolution to UUID, merge into IssueFilter:
```json
{ "parent": { "id": { "eq": "<uuid>" } } }
```

### `readBody` / `readDescription` helpers
- Path `"-"` -> `io.ReadAll(cmd.InOrStdin())`
- Other path -> `os.ReadFile(path)`
- Trailing newlines preserved as-is (LLMs and editors typically add one; consistent with how `cat file | gh ...` would behave)

### Cobra flag combinations
- `comment create`: `MarkFlagsMutuallyExclusive("body","body-file")` + `MarkFlagsOneRequired("body","body-file")`
- `comment update`: same pair (update needs new content)
- `issue create`: mutual exclusivity only (description was always optional)
- `issue update`: mutual exclusivity only (preserves partial-update semantics — neither flag means "leave description unchanged")

## Post-Completion
*Items requiring manual intervention or external systems - no checkboxes, informational only*

**Manual verification** (out of plan):
- Smoke test against live Linear API: `linear-cli comment create ENG-X --body-file /tmp/test.md`, `linear-cli issue list --parent ENG-727`, `linear-cli view issues "Without Estimates"`.

**External system updates** (out of plan):
- Cut new release tag.
- Update Homebrew cask formula `linear-cli` SHA256 and version.
- Announce changes in relevant channels.
