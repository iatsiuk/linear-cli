# Comment & Attachment CRUD Completion

## Overview
- Implement missing `comment update`, `comment delete` CLI commands
- Implement `attachment show` (view metadata) and `attachment download` (save file to disk) commands
- Investigate and fix `attachment create --file` upload failure (HTTP 400)
- All mutations exist in GraphQL API (`docs/schema.graphql`) but lack CLI wrappers

## Context (from discovery)

**Schema reference:** `docs/schema.graphql` (authoritative source for all types and mutations)

**Files involved:**
- `internal/cmd/comment.go` -- comment command tree (create, list)
- `internal/cmd/comment_test.go` -- 13 existing tests
- `internal/cmd/attachment.go` -- attachment command tree (create, list, delete)
- `internal/cmd/attachment_test.go` -- 13 existing tests
- `internal/query/comment.go` -- GraphQL queries (CommentListQuery, CommentCreateMutation)
- `internal/query/comment_test.go` -- 2 query structure tests
- `internal/query/attachment.go` -- GraphQL queries (List, Create, Delete)
- `internal/query/attachment_test.go` -- 3 query structure tests
- `internal/model/comment.go` -- Comment struct
- `internal/model/attachment.go` -- Attachment struct
- `internal/api/upload.go` -- file upload flow (fileUpload mutation + PUT)

**Patterns to follow:**
- `internal/cmd/issue_update.go` -- update pattern: fetch by ID -> build `map[string]any` input -> call mutation -> verify success -> output
- `internal/cmd/issue_delete.go` -- delete pattern: fetch by ID -> confirm prompt (`--yes` to skip) -> call mutation -> verify success -> print message
- `internal/cmd/doc.go` -- doc show as reference for single-entity display

**GraphQL mutations (from `docs/schema.graphql`):**
- `commentUpdate(id: String!, input: CommentUpdateInput!)` -> `CommentPayload!` (line ~17919)
- `commentDelete(id: String!)` -> `DeletePayload!` (line ~17891)
- `fileUpload(contentType: String!, filename: String!, size: Int!)` -> `UploadPayload!` (line ~18313)

**GraphQL queries for attachment:**
- `attachment(id: String!)` -> `Attachment!` (schema line ~16567) -- fetch single attachment by ID
- `Attachment.url` field (line ~1368) -- HTTPS URL to file on Linear CDN, can be downloaded with HTTP GET

**CommentUpdateInput (from schema, line ~2755):**
- `body: String` -- the only user-facing field; rest are internal

## Development Approach
- **Testing approach**: TDD -- write tests first, then implement code to make them pass
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run `make build` after each change (includes linter)

## Testing Strategy
- **Unit tests**: table-driven tests using `newQueuedServer()` mock pattern
- **Test files**: `*_test.go` in same package (`internal/cmd`, `internal/query`)
- Test both success and error paths
- Test JSON output mode
- Test flag validation
- Capture and verify mutation variables

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix

## Implementation Steps

### Task 1: Add CommentUpdateMutation and CommentDeleteMutation queries

Write query strings and tests first, then verify they compile.

- [x] write tests in `internal/query/comment_test.go`:
  - `TestCommentUpdateMutation` -- verify operation name, `$id` and `$input` variables, `commentUpdate` block, comment fields returned
  - `TestCommentDeleteMutation` -- verify operation name, `$id` variable, `commentDelete` block, `success` field
- [x] add `CommentUpdateMutation` to `internal/query/comment.go`:
  ```
  mutation CommentUpdate($id: String!, $input: CommentUpdateInput!) {
    commentUpdate(id: $id, input: $input) {
      comment { ...commentFields }
    }
  }
  ```
- [x] add `CommentDeleteMutation` to `internal/query/comment.go`:
  ```
  mutation CommentDelete($id: String!) {
    commentDelete(id: $id) { success }
  }
  ```
- [x] run `make build` and tests -- must pass before next task

### Task 2: Implement `comment update` command

TDD: write cmd tests first, then implement the command.

- [x] write tests in `internal/cmd/comment_test.go`:
  - `TestCommentUpdate_Success` -- mock server returns updated comment; verify output shows updated body
  - `TestCommentUpdate_JSON` -- verify `--json` outputs valid JSON with updated comment
  - `TestCommentUpdate_MissingBody` -- verify error when `--body` not provided
  - `TestCommentUpdate_NotFound` -- verify error message when API returns error
  - `TestCommentUpdate_VerifyInput` -- capture request body, verify `id` and `input.body` sent correctly
- [x] implement `newCommentUpdateCmd()` in `internal/cmd/comment.go`:
  - accepts comment UUID as positional arg
  - `--body` flag (required) -- new comment text
  - calls `CommentUpdateMutation` with `map[string]any{"body": body}`
  - output: `Comment <id> updated.`; with `--json` outputs full comment JSON
- [x] register in comment command tree: `cmd.AddCommand(newCommentUpdateCmd())`
- [x] run `make build` and tests -- must pass before next task

### Task 3: Implement `comment delete` command

TDD: write cmd tests first, then implement the command.

- [x] write tests in `internal/cmd/comment_test.go`:
  - `TestCommentDelete_Success` -- mock server returns success=true; verify confirmation message
  - `TestCommentDelete_WithYesFlag` -- verify `--yes` skips confirmation
  - `TestCommentDelete_Abort` -- verify user declining confirmation aborts
  - `TestCommentDelete_NotFound` -- verify error on API error
  - `TestCommentDelete_MutationFails` -- verify error when success=false
- [x] implement `newCommentDeleteCmd()` in `internal/cmd/comment.go`:
  - accepts comment UUID as positional arg
  - `--yes` flag to skip confirmation
  - confirmation prompt: `Delete comment <id>? [y/N]`
  - calls `CommentDeleteMutation`
  - output: `Comment <id> deleted.`
- [x] register in comment command tree: `cmd.AddCommand(newCommentDeleteCmd())`
- [x] run `make build` and tests -- must pass before next task

### Task 4: Add AttachmentShowQuery

Query to fetch a single attachment by ID (needed for `show` and `download`).

- [x] write test in `internal/query/attachment_test.go`:
  - `TestAttachmentShowQuery` -- verify operation name, `$id` variable, `attachment` block, fields include `url`, `title`, `creator`
- [x] add `AttachmentShowQuery` to `internal/query/attachment.go`:
  ```
  query AttachmentShow($id: String!) {
    attachment(id: $id) { ...attachmentFields }
  }
  ```
- [x] run `make build` and tests -- must pass before next task

### Task 5: Implement `attachment show` command

Display attachment metadata (title, URL, creator, dates).

- [ ] write tests in `internal/cmd/attachment_test.go`:
  - `TestAttachmentShow_Success` -- mock server returns attachment; verify table output (Title, URL, Creator, Created, Updated)
  - `TestAttachmentShow_JSON` -- verify `--json` outputs valid JSON with all fields
  - `TestAttachmentShow_NotFound` -- verify error message when attachment not found
- [ ] implement `newAttachmentShowCommand()` in `internal/cmd/attachment.go`:
  - accepts attachment UUID as positional arg
  - display format (like `doc show` or `issue show`):
    ```
    Title:     Screenshot link
    URL:       https://cdn.linear.app/...
    Creator:   aleksei.i
    Created:   2026-03-09T12:28:16.572Z
    Updated:   2026-03-09T12:28:16.572Z
    ```
  - with `--json` outputs full attachment JSON
- [ ] register: `cmd.AddCommand(newAttachmentShowCommand())`
- [ ] run `make build` and tests -- must pass before next task

### Task 6: Implement `attachment download` command

Download attachment file to disk. Fetches attachment metadata via GraphQL, then HTTP GET on the `url` field.

- [ ] write tests in `internal/cmd/attachment_test.go`:
  - `TestAttachmentDownload_Success` -- mock GraphQL + mock HTTP file server; verify file saved to disk with correct content
  - `TestAttachmentDownload_CustomOutput` -- verify `--output` flag saves to specified path
  - `TestAttachmentDownload_StdoutDash` -- verify `--output -` writes to stdout
  - `TestAttachmentDownload_NotFound` -- verify error when attachment not found
  - `TestAttachmentDownload_HTTPError` -- verify error when file URL returns non-2xx
  - `TestAttachmentDownload_FilenameFromURL` -- verify default filename derived from URL path
- [ ] implement `newAttachmentDownloadCommand()` in `internal/cmd/attachment.go`:
  - accepts attachment UUID as positional arg
  - `--output` / `-o` flag -- destination path (default: filename from URL, saved to current dir)
  - `--output -` -- write to stdout (for piping)
  - flow:
    1. fetch attachment via `AttachmentShowQuery` to get `url`
    2. HTTP GET on the `url`
    3. save response body to file (or stdout)
    4. print: `Downloaded: <filename> (<size>)` (unless stdout mode)
  - use `client.HTTPClient` or plain `http.Client` for the download (no auth needed, Linear CDN URLs are signed)
- [ ] register: `cmd.AddCommand(newAttachmentDownloadCommand())`
- [ ] run `make build` and tests -- must pass before next task

### Task 7: Fix attachment file upload (HTTP 400)

**Root cause found:** Google Cloud Storage signed URL includes `content-type` in
`X-Goog-SignedHeaders`, but `upload.go` does not set `Content-Type` header on the
PUT request. Linear API returns `Content-Disposition` and `x-goog-content-length-range`
in the `headers` array, but NOT `Content-Type` -- the client must set it explicitly
using the same value passed to the `fileUpload` mutation.

**Error from GCS:** `MalformedSecurityHeader: Header was included in signedheaders, but not in the request. ParameterName: content-type`

- [ ] write test in `internal/api/upload_test.go`: verify PUT request includes `Content-Type` header matching the contentType passed to `fileUpload` mutation
- [ ] fix `internal/api/upload.go`: add `req.Header.Set("Content-Type", contentType)` after applying API headers (after line 94); this requires passing `contentType` into the PUT section or restructuring slightly
- [ ] run `make build` and tests -- must pass before next task

### Task 8: Verify acceptance criteria

- [ ] verify `comment update <id> --body "new text"` works (unit tests pass)
- [ ] verify `comment delete <id> --yes` works (unit tests pass)
- [ ] verify `attachment show <id>` works (unit tests pass)
- [ ] verify `attachment download <id>` works (unit tests pass)
- [ ] verify `attachment download <id> -o -` writes to stdout (unit tests pass)
- [ ] verify existing comment/attachment tests still pass (no regressions)
- [ ] verify file upload fix (if root cause identified)
- [ ] run full test suite: `go test -race ./...`
- [ ] run linter: `make build`

## Technical Details

**Comment update flow:**
1. Parse comment UUID from args
2. Validate `--body` flag is set
3. Build input: `map[string]any{"body": bodyValue}`
4. Call `CommentUpdateMutation` with `{id, input}`
5. Check response for comment payload
6. Output confirmation or JSON

**Comment delete flow:**
1. Parse comment UUID from args
2. Show confirmation (unless `--yes`)
3. Call `CommentDeleteMutation` with `{id}`
4. Check `success` field in `DeletePayload`
5. Print confirmation message

**Attachment show flow:**
1. Parse attachment UUID from args
2. Call `AttachmentShowQuery` with `{id}`
3. Display key-value pairs (Title, URL, Creator, Created, Updated) or JSON

**Attachment download flow:**
1. Parse attachment UUID from args
2. Call `AttachmentShowQuery` with `{id}` to get `url`
3. Derive filename: `--output` flag, or last path segment of URL, or `attachment-<id>`
4. HTTP GET on the `url` (no auth -- Linear CDN URLs are pre-signed)
5. Stream response body to file (or stdout if `--output -`)
6. Print `Downloaded: <filename> (<size>)` on success

**Upload fix (Task 7):**
- `fileUpload` mutation returns `uploadUrl`, `assetUrl`, `headers` array
- GCS signed URL requires `Content-Type` in PUT request (listed in `X-Goog-SignedHeaders`)
- Linear API does NOT include `Content-Type` in the returned `headers` array
- Fix: explicitly set `Content-Type` on PUT request using the same value passed to `fileUpload`

## Post-Completion

**E2E verification:**
- Re-run E2E step 2 tests from `e2e-step-2.md` with new commands
- Test `comment update` and `comment delete` against real Linear API
- Test `attachment show` and `attachment download` against real Linear API
- Re-test `attachment create --file` after upload fix
