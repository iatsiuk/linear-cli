# Linear-CLI Project Instructions

## Project Files

- `docs/schema.graphql` - Linear GraphQL API schema (authoritative source)
- `docs/linear-api.md` - Linear API reference with examples, patterns, and gotchas

### API Reference Priority

`docs/schema.graphql` is the primary source of truth for implementation.
Always verify GraphQL types, field names, nullability, and input types against
the schema when writing code and tests. Use `docs/linear-api.md` as a
supplementary guide for usage patterns, but resolve any discrepancies in favor
of the schema.

## Code Style

### Imports

Group imports in order, separated by blank lines:

1. Standard library
2. External packages
3. Local packages (`linear-cli/...`)

```go
import (
    "context"
    "fmt"

    "github.com/spf13/cobra"

    "linear-cli/internal/config"
)
```

### Naming

- Package names: short, lowercase, no underscores (`db`, `config`, `query`)
- Exported types: PascalCase (`Config`, `Session`, `Query`)
- Unexported: camelCase (`validateQuery`, `buildRequest`)
- Acronyms: consistent case (`URL`, `HTTP`, `API` or `url`, `http`, `api`)
- Receivers: short, 1-2 letters (`s` for `*Session`, `q` for `*Query`)
- Errors: `Err` prefix for sentinel errors (`ErrConnectionFailed`)

### Functions

- Early returns for error handling
- Group related functions together

### Error Handling

- Wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Check all errors (enforced by `errcheck` linter)
- Use `errors.Is`/`errors.As` for error comparison
- Sentinel errors as package-level variables

### Comments

- Only for non-obvious logic
- English, lowercase, brief
- No comments for self-explanatory code

### Structs

- JSON tags on all exported fields: `json:"field_name"`
- Use `omitempty` for optional fields
- Pointer types for optional values (`*float64`, `*int`)
- Group related fields together

### Variables

- Package-level constants in `const` block
- Related constants grouped together
- Unexported package variables with `var`

### Control Flow

- Use `range` with index for modifying slices
- Prefer `for i := range n` over `for i := 0; i < n; i++` (Go 1.22+)
- Use `switch` over long `if-else` chains

### Concurrency

- Use `context.Context` as first parameter
- Use `sync.Mutex` for simple locking
- Use `errgroup` for parallel operations

## Testing

- Use table-driven tests for multiple scenarios
- Use stdlib `testing` package only (no testify)
- Test error paths: timeouts, context cancellation
- Run with race detector: `go test -race ./...`
- Use `t.Parallel()` for independent tests
- Test files: `*_test.go` in same package

## Language

All documentation, comments, and text must be in English.

## Building

- Always build with `make build` (runs linter automatically)
- Direct `go build` skips linting - avoid it

## Linting and Formatting

- Run `golangci-lint run` before committing (executed automatically via `make build`)
- Fix formatting issues with `goimports -w <file>` or `gofmt -w <file>`
- Config: `.golangci.yml` defines enabled linters
- No trailing whitespace, proper import grouping (stdlib, external, local)

## API Patterns

### ID Resolution

Use functions in `internal/api/resolver.go` to convert human-readable names to UUIDs:
- `ResolveTeamID` - team key (e.g. "ENG") or UUID
- `ResolveLabelID` - label name or UUID; accepts optional teamID to restrict search
- `ResolveUserID` - display name or email or UUID; tries name first, then email
- `ResolveStateID` - workflow state name or UUID; accepts optional teamID
- `ResolveProjectID` - project name or UUID
- `ResolveViewerID` - returns authenticated user ID (used for "me" assignee)

Pattern: if input already matches UUID regex, return immediately without API call.

### Filter Builder

Use `internal/filter` to add advanced issue filter flags and build `IssueFilter` variables:
- `filter.AddFlags(cmd)` - registers all filter flags on a command
- `filter.BuildFromFlags(f)` - constructs an `IssueFilter` map from parsed flags; returns nil if no flags set
- `filter.ParseDate(s)` - converts CLI date aliases (7d, 2w, today, yesterday) to Linear API format

Flags added: `--created-after`, `--created-before`, `--updated-after`, `--updated-before`,
`--due-after`, `--due-before`, `--completed-after`, `--completed-before`, `--no-assignee`,
`--no-project`, `--no-cycle`, `--priority-gte`, `--priority-lte`, `--my`, `--or`.

### Partial Update Mutations

Use `map[string]any` for mutation input variables when partial updates are needed.
Omitting a key leaves the field unchanged. Do not use typed structs for update
inputs - `omitempty` cannot distinguish "not provided" from "zero value".
