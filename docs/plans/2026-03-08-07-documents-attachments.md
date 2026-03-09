# Documents & Attachments

## Overview
- Implement CRUD for documents and attachments, including file upload
- Commands: `linear doc list/show/create/update/delete`, `linear attachment list/create/delete`
- Depends on Foundation plan

## Context
- Document key fields: id, title, content (markdown), project, creator, createdAt, updatedAt
- DocumentCreateInput: `title: String!` (required), `content`, `projectId` optional
- DocumentUpdateInput: `title`, `content`, `hiddenAt`, `trashed`, `issueId`, `teamId`
- `documentDelete` = moves to trash (30-day grace period), `documentUnarchive` = restore
- Attachment: id, title, url, issueId. `attachmentCreate` is idempotent (same url+issueId = update)
- File Upload is two-step: (1) `fileUpload` mutation -> get `uploadUrl` + `assetUrl`, (2) PUT file to `uploadUrl`
- Integration-specific attachment mutations exist (Slack, GitHub, Jira, etc.) but are out of scope for CLI
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

### Task 1: Document model and queries
- [x] write tests for Document struct deserialization (all key fields, nullable fields)
- [x] create `internal/model/document.go`:
  - Document struct: ID, Title, Content(*string), Creator(*User), Project(*Project), CreatedAt, UpdatedAt, ArchivedAt
- [x] create `internal/query/document.go`:
  - DocumentListQuery, DocumentGetQuery, DocumentCreateMutation, DocumentUpdateMutation, DocumentDeleteMutation
- [x] write tests for query construction
- [x] run `make test` - must pass before next task

### Task 2: `linear doc list/show` commands
- [ ] write tests for doc list: fetches documents, table output, filters
- [ ] write tests for doc show: fetches by ID, displays full content
- [ ] create `internal/cmd/doc.go`:
  - `doc list`: flags `--project`, `--limit`, `--json`, `--include-archived`
  - `doc show <id>`: full document details with content
  - table columns: Title | Project | Creator | Updated
- [ ] run `make test` - must pass before next task

### Task 3: `linear doc create/update/delete` commands
- [ ] write tests for doc create: required title, sends mutation
- [ ] write tests for doc update: partial update, correct mutation
- [ ] write tests for doc delete: trash mutation (30-day grace), confirmation
- [ ] create `internal/cmd/doc_create.go`:
  - flags: `--title` (required), `--content`, `--project`
  - `--content-file` flag: read content from file
- [ ] create `internal/cmd/doc_update.go`:
  - flags: `--title`, `--content`, `--content-file`
  - accepts document ID or URL identifier as argument
- [ ] create `internal/cmd/doc_delete.go`:
  - `documentDelete` (trash, 30-day grace)
  - `--restore` flag: `documentUnarchive` to restore from trash
  - `--yes` to skip confirmation
- [ ] run `make test` - must pass before next task

### Task 4: Attachment model and queries
- [ ] write tests for Attachment struct deserialization
- [ ] create `internal/model/attachment.go`:
  - Attachment struct: ID, Title, URL, Issue, Creator, CreatedAt, UpdatedAt
- [ ] create `internal/query/attachment.go`:
  - AttachmentListQuery, AttachmentGetQuery, AttachmentCreateMutation, AttachmentDeleteMutation
- [ ] write tests for query construction
- [ ] run `make test` - must pass before next task

### Task 5: `linear attachment list/create/delete` commands
- [ ] write tests for attachment list: fetches for issue, table output
- [ ] write tests for attachment create: sends mutation with url + issueId
- [ ] write tests for attachment delete: sends delete mutation
- [ ] create `internal/cmd/attachment.go`:
  - `attachment list <issue-identifier>`: flags `--json`
  - `attachment create <issue-identifier>`: flags `--url` (required), `--title`
  - `attachment delete <id>`: `--yes` to skip confirmation
  - table columns: Title | URL | Created
- [ ] run `make test` - must pass before next task

### Task 6: File upload support
- [ ] write tests for two-step upload: fileUpload mutation, PUT to uploadUrl
- [ ] create `internal/api/upload.go`:
  - `Upload(ctx, client, filePath) (assetURL string, error)`:
    1. call `fileUpload` mutation with contentType, filename, size
    2. PUT file to returned `uploadUrl` with provided headers
    3. return `assetUrl` for use in descriptions/attachments
- [ ] add `--file` flag to `attachment create`: uploads file, then creates attachment with assetUrl
- [ ] write tests for upload errors (file not found, upload failure)
- [ ] run `make test` - must pass before next task

### Task 7: Verify acceptance criteria
- [ ] verify `linear doc list` shows documents
- [ ] verify `linear doc create --title "Test" --content "..."` creates document
- [ ] verify `linear doc delete` trashes document, `--restore` restores
- [ ] verify `linear attachment create ENG-123 --url "..."` creates attachment
- [ ] verify `linear attachment create ENG-123 --file ./screenshot.png` uploads and attaches
- [ ] verify idempotent attachment creation (same url+issue = update)
- [ ] run `make test` - full suite must pass
- [ ] run `make build` - lint + build must pass

## Technical Details

### File upload flow
```
1. fileUpload(contentType, filename, size) -> { uploadUrl, assetUrl, headers }
2. PUT file to uploadUrl with headers
3. Use assetUrl in descriptions/attachments
```

### Table columns
**doc list**: Title | Project | Creator | Updated
**attachment list**: Title | URL | Created

### Task 8: [Final] Update documentation
- [ ] update README.md with document and attachment commands usage
- [ ] document file upload workflow

## Post-Completion
- Manual testing with file uploads (images, PDFs)
- Verify idempotent attachment behavior
