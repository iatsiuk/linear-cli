# E2E Bugfixes

## Overview
- Fix 7 bugs discovered during E2E testing (documented in `e2e-errors.md`)
- Bugs span config path, issue model, resolver logic, attachment download, and doc create
- All fixes must be validated against `docs/schema.graphql` (authoritative source for GraphQL types)

## Context (from discovery)

**Schema reference:** `docs/schema.graphql` (authoritative source for all types and mutations)

**Files involved:**
- `internal/config/config.go` -- config path resolution (`configPath()`, line 27)
- `internal/model/issue.go` -- Issue struct (line 8), missing `parent` and `project` fields
- `internal/query/issue.go` -- `issueFields` constant (line 4), missing parent/project in query
- `internal/api/resolver.go` -- `ResolveUserID` (line 89), `ResolveStateID` (line 129)
- `internal/cmd/issue_batch.go` -- batch update passes empty teamID to `ResolveStateID` (line 107)
- `internal/cmd/attachment.go` -- `runAttachmentDownload` (line 147), bare http.Client without auth
- `internal/api/client.go` -- `Client.apiKey` field (line 18), auth header format (line 99)
- `internal/cmd/doc_create.go` -- `runDocCreate` (line 38), no handling for workspace-required project

**Schema types to verify:**
- `Issue.parent` -- `Issue` type (nullable), schema line ~12450
- `Issue.project` -- `Project` type (nullable), schema line ~12470
- `User.name` vs `User.displayName` -- schema line ~37673
- `DocumentCreateInput` -- `projectId: String` (nullable), schema line ~6652
- `WorkflowState` -- team association, used by `ResolveStateID`

**Patterns to follow:**
- `internal/api/resolver.go` -- existing resolver pattern for name-to-UUID conversion
- `internal/api/client.go:99` -- `Authorization: <apiKey>` header format (no Bearer prefix)
- `internal/config/config.go` -- config path resolution

## Development Approach
- **Testing approach**: TDD -- write tests first, then implement code to make them pass
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- **CRITICAL: verify all GraphQL types and field names against `docs/schema.graphql`**
- Run `make build` after each change (includes linter)

## Testing Strategy
- **Unit tests**: table-driven tests using existing mock patterns (`newQueuedServer()`)
- **Test files**: `*_test.go` in same package
- Test both success and error paths
- Capture and verify request variables where applicable

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix

## Implementation Steps

### Task 1: Fix config path for Linux/macOS

Use `os.UserHomeDir()` + `.config/linear-cli` on Linux and macOS (GOOS=linux, GOOS=darwin).
Fall back to `os.UserConfigDir()` + `linear-cli` on other OSes (Windows, etc.).

- [x] write test in `internal/config/config_test.go`:
  - `TestConfigPath_DefaultDir` -- verify that on linux/darwin the path ends with `.config/linear-cli/config.yaml`
- [x] update `configPath()` in `internal/config/config.go`:
  - check `runtime.GOOS` for "linux" or "darwin"
  - if match: use `os.UserHomeDir()` + `.config/linear-cli`
  - otherwise: use `os.UserConfigDir()` + `linear-cli`
- [x] run `make build` and tests -- must pass before next task

### Task 2: Add parent and project fields to Issue model and query

Verify against schema: `Issue.parent` (nullable `Issue`), `Issue.project` (nullable `Project`).
Check `Project` type in schema for available fields.

- [x] write tests in `internal/query/issue_test.go`:
  - `TestIssueFieldsContainsParent` -- verify `issueFields` contains `parent { id identifier title }`
  - `TestIssueFieldsContainsProject` -- verify `issueFields` contains `project { id name }`
- [x] write tests in `internal/model/issue_test.go`:
  - `TestIssueUnmarshal_WithParent` -- verify Issue JSON with parent unmarshals correctly
  - `TestIssueUnmarshal_WithProject` -- verify Issue JSON with project unmarshals correctly
  - `TestIssueUnmarshal_NilParentProject` -- verify omitted parent/project unmarshal as nil
- [x] add `Parent` field to `Issue` struct in `internal/model/issue.go`:
  - type: pointer to a lightweight struct (id, identifier, title) -- avoid recursive `*Issue`
  - JSON tag: `json:"parent,omitempty"`
- [x] add `Project` field to `Issue` struct in `internal/model/issue.go`:
  - type: pointer to a lightweight struct (id, name)
  - JSON tag: `json:"project,omitempty"`
- [x] update `issueFields` constant in `internal/query/issue.go`:
  - add `parent { id identifier title }`
  - add `project { id name }`
- [x] run `make build` and tests -- must pass before next task

### Task 3: Fix ResolveUserID to try displayName

Schema: `User.name: String!` (full name), `User.displayName: String!` (display name).
Current code only tries `name` then `email`. Users may pass `displayName` value.

- [x] write tests in `internal/api/resolver_test.go`:
  - `TestResolveUserID_ByDisplayName` -- name query returns empty, displayName query returns match
  - `TestResolveUserID_ByName` -- name query returns match (existing behavior)
  - `TestResolveUserID_ByEmail` -- name and displayName return empty, email returns match
  - `TestResolveUserID_NotFound` -- all three queries return empty, error returned
- [x] update `ResolveUserID` in `internal/api/resolver.go`:
  - add `qDisplayName` query: `users(filter: { displayName: { eq: $displayName } }, first: 1)`
  - resolution order: name -> displayName -> email
  - verify `displayName` filter field exists in schema's `UserFilter` type
- [x] run `make build` and tests -- must pass before next task

### Task 4: Fix batch update state resolution with team context

Current code passes empty teamID to `ResolveStateID` (line 107 in `issue_batch.go`).
This resolves the first matching state workspace-wide, which may belong to a different team.

- [x] write tests in `internal/cmd/issue_batch_test.go`:
  - `TestBatchUpdate_StateResolvesWithTeam` -- verify state resolution query includes team filter
  - `TestBatchUpdate_StateAcrossTeams` -- verify error when issues span multiple teams with --state
- [x] fix `runIssueBatchUpdate` in `internal/cmd/issue_batch.go`:
  - when `--state` is used: fetch team from the first issue in the batch
  - pass teamID to `api.ResolveStateID(ctx, client, stateName, teamID)`
  - if issues span multiple teams, resolve state per team or return clear error
- [x] run `make build` and tests -- must pass before next task

### Task 5: Fix attachment download authentication

`uploads.linear.app` requires `Authorization: <API_KEY>` header (no Bearer prefix).
Reference: https://developers.linear.app/docs/oauth/file-storage-authentication

- [ ] write tests in `internal/cmd/attachment_test.go`:
  - `TestAttachmentDownload_AuthHeader` -- verify download request includes Authorization header
  - `TestAttachmentDownload_AuthHeaderNoBearerPrefix` -- verify header value has no Bearer prefix
- [ ] update `runAttachmentDownload` in `internal/cmd/attachment.go`:
  - load API key from config (already available via `newClientFromConfig()`)
  - add `req.Header.Set("Authorization", apiKey)` to download request
  - refactor to expose API key from client or pass it through
- [ ] run `make build` and tests -- must pass before next task

### Task 6: Improve doc create error handling

Schema: `DocumentCreateInput.projectId` is nullable, but some workspaces require it.
Also: `--json` is already implemented but `e2e-errors.md` reports it doesn't work --
investigate and fix if needed.

- [ ] write tests in `internal/cmd/doc_create_test.go`:
  - `TestDocCreate_ArgumentValidationError` -- verify clear error message when API returns "Argument Validation Error"
  - `TestDocCreate_HintProjectFlag` -- verify error suggests `--project` flag
- [ ] update `runDocCreate` in `internal/cmd/doc_create.go`:
  - catch "Argument Validation Error" and append hint: `(try adding --project flag)`
  - verify `--json` output works in tests (may already work -- confirm and add test if missing)
- [ ] run `make build` and tests -- must pass before next task

### Task 7: Verify acceptance criteria

- [ ] verify config path resolves to `~/.config/linear-cli/config.yaml` on macOS (unit tests)
- [ ] verify `issue show --json` includes parent and project fields (unit tests)
- [ ] verify `ResolveUserID` finds users by displayName (unit tests)
- [ ] verify batch update resolves state within correct team (unit tests)
- [ ] verify attachment download passes auth header (unit tests)
- [ ] verify doc create shows helpful error without --project (unit tests)
- [ ] run full test suite: `go test -race ./...`
- [ ] run linter: `make build`

## Technical Details

**Config path fix:**
- `runtime.GOOS == "linux" || runtime.GOOS == "darwin"` -> `os.UserHomeDir()` + `.config/linear-cli`
- other OS -> `os.UserConfigDir()` + `linear-cli`
- `LINEAR_CONFIG_DIR` env override unchanged (highest priority)

**Issue model additions:**
- `Parent *IssueRef` where `IssueRef` has `ID`, `Identifier`, `Title` -- avoids recursive Issue
- `Project *ProjectRef` where `ProjectRef` has `ID`, `Name`
- Or reuse existing model types if suitable

**User resolution order:**
1. UUID check (existing)
2. `name` exact match (existing)
3. `displayName` exact match (new)
4. `email` exact match (existing)

**Batch state resolution:**
- Fetch team from first issue: `issue(id: $id) { team { id } }`
- Pass teamID to `ResolveStateID`
- If batch spans multiple teams with --state, return error explaining limitation

**Attachment download auth:**
- Header format: `Authorization: <apiKey>` (no Bearer prefix, same as GraphQL API)
- API key source: same config/env as GraphQL client

## Post-Completion

**E2E verification:**
- Re-run E2E tests from `e2e-tests.md` and `e2e-step-2.md` for affected commands
- Test `issue show BACK-14 --json` -- verify parent/project in output
- Test `issue batch update` with --state across team issues
- Test `attachment download` for uploaded files
- Test `doc create` without --project -- verify improved error
- Test user resolution with display names
- Verify config path on macOS after fix
