# CLI output consistency (TDD)

## Overview
- Eliminate three silent/output-shape failures observed in 1551 Claude Code transcripts:
  - Plain-text errors break `linear-cli ... --json | jq` pipelines (no structured envelope).
  - `comment list --json` returns `null` for empty collections while `attachment list --json` returns `[]` — inconsistent.
  - Empty-result commands in table mode print zero bytes with exit 0 (`project issues "AI QC Bot"`), so callers cannot distinguish "no rows" from "command did nothing".
- Three centralized fixes:
  - `cmd/linear-cli/main.go` — JSON-shaped error envelope on stderr when `--json` is set.
  - `internal/api/pagination.go` — initialize `all` as zero-length slice instead of `nil`.
  - `internal/output/formatter.go` — `JSONFormatter` substitutes nil slice with empty slice via reflection; `TableFormatter` prints `(no results)` line when empty.

## Context (from discovery)
- Entry point: `cmd/linear-cli/main.go` (19 LoC, currently `fmt.Fprintln(os.Stderr, err)`)
- Root command: `internal/cmd/root.go` (declares `--json` PersistentFlag, `SilenceErrors: true`, `SilenceUsage: true`)
- Pagination: `internal/api/pagination.go` (line 28: `var all []T` — root cause of `null` JSON)
- Formatter: `internal/output/formatter.go` (`JSONFormatter` lines 27-35; `TableFormatter.Format` returns nil at line 47-48 when empty)
- Tests: `internal/output/formatter_test.go`, `internal/api/pagination_test.go`, `internal/cmd/root_test.go`
- Affected commands: every list command (issue list, comment list, attachment list, project list, view list, etc.) — fix is centralized so per-command changes are zero

## Development Approach
- **Testing approach**: TDD (tests first, then implementation)
- Each task: write failing tests first, then implement code to make them pass
- **CRITICAL: every task MUST include new/updated tests** for code changes
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run `make build` after each task (includes linter)

## Testing Strategy
- **Unit tests**: table-driven, in-memory buffers (`bytes.Buffer`) for stdout/stderr capture
- No e2e tests in this project
- Audit existing list-command tests for expectations on empty-output (table mode) — update those that asserted zero bytes
- Test JSON-mode error path with cobra cmd that returns synthetic error
- No regression on non-empty output paths (header layout, JSON shape unchanged)

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with `+` prefix
- Document issues/blockers with `!` prefix
- Update plan if implementation deviates from original scope

## What Goes Where
- **Implementation Steps** (`[ ]` checkboxes): code, tests, docs in this repo
- **Post-Completion** (no checkboxes): release/cask/manual verification

## Implementation Steps

### Task 1: Fix `PaginateAll` to never return nil slice

**Tests first:**
- [x] in `internal/api/pagination_test.go` update `TestPaginateAll_EmptyResult` to assert `nodes != nil && len(nodes) == 0`, and additionally assert `json.Marshal(nodes)` produces `[]` not `null`
- [x] add `TestPaginateAll_EmptyJSONShape` if cleaner as a separate test

**Implementation:**
- [x] in `internal/api/pagination.go` change `var all []T` (line 28) to `all := make([]T, 0)`
- [x] verify other tests in the same file still pass (single-page, multi-page, error paths unaffected)
- [x] run `go test ./internal/api/...` - must pass before Task 2

### Task 2: `JSONFormatter` defends against nil slices via reflection

**Tests first:**
- [x] in `internal/output/formatter_test.go` add `TestJSONFormatter_NilSlice`: pass `var nilSlice []int = nil`, assert formatted output is `[]\n`
- [x] add `TestJSONFormatter_TypedNilSlice`: pass `var nilSlice []someStruct = nil`, assert formatted output is `[]\n`
- [x] add `TestJSONFormatter_NonSlice`: pass a struct value, assert it is marshaled normally (no behavior change)
- [x] add `TestJSONFormatter_EmptySlice`: pass `[]int{}`, assert `[]\n` (regression check)

**Implementation:**
- [x] in `internal/output/formatter.go` `JSONFormatter.Format`:
  - at top: `v := reflect.ValueOf(data); if v.Kind() == reflect.Slice && v.IsNil() { data = reflect.MakeSlice(v.Type(), 0, 0).Interface() }`
- [x] add `reflect` import if not present
- [x] run `go test ./internal/output/...` - must pass before Task 3

### Task 3: `TableFormatter` prints `(no results)` for empty input

**Tests first:**
- [x] in `internal/output/formatter_test.go` add `TestTableFormatter_EmptySlice`: pass `[]someRow{}`, assert output is `(no results)\n`
- [x] add `TestTableFormatter_NilSlice`: pass nil slice, assert `(no results)\n`
- [x] verify existing `TestTableFormatter_*` tests for non-empty input still pass without modification

**Implementation:**
- [x] in `internal/output/formatter.go` `TableFormatter.Format`:
  - replace `if v.Len() == 0 { return nil }` with `if v.Len() == 0 { _, err := fmt.Fprintln(w, "(no results)"); return err }`
- [x] run `go test ./internal/output/...` - must pass before Task 4

### Task 4: Audit and fix command tests broken by Task 3

**Tests first:**
- [x] grep test files for assertions on empty output: `grep -rn 'assert.*Equal.*""' internal/cmd/*_test.go` and `grep -rn 'len(out)' internal/cmd/*_test.go`
- [x] identify any test that expected zero-byte output for empty list result; update it to expect `(no results)\n`

**Implementation:**
- [x] update each affected test (likely in `comment_test.go`, `issue_test.go`, `attachment_test.go`, `project_test.go`, `view_test.go`)
- [x] do NOT change command behavior — only test expectations
- [x] run `go test ./...` - all tests must pass before Task 5

### Task 5: JSON error envelope in `main.go`

**Tests first:**
- [x] add `cmd/linear-cli/main_test.go` (or extend existing if any), test the printing helper directly:
  - extract `formatExecError(err error, jsonMode bool, w io.Writer)` from main into a testable function
  - `TestFormatExecError_JSON`: jsonMode=true, err with message containing quotes/newlines, asserts output is valid JSON parseable as `{"error": "..."}` and equals expected escaped form, ends with `\n`
  - `TestFormatExecError_Plain`: jsonMode=false, asserts output is `<msg>\n` (current behavior)
  - `TestFormatExecError_NilError`: nil error, asserts no output (defensive)

**Implementation:**
- [x] refactor `cmd/linear-cli/main.go`:
  ```go
  func main() {
      root := cmd.NewRootCommand(version)
      err := root.Execute()
      if err == nil {
          return
      }
      jsonMode, _ := root.PersistentFlags().GetBool("json")
      formatExecError(err, jsonMode, os.Stderr)
      os.Exit(1)
  }

  func formatExecError(err error, jsonMode bool, w io.Writer) {
      if err == nil {
          return
      }
      if jsonMode {
          b, _ := json.Marshal(map[string]string{"error": err.Error()})
          fmt.Fprintln(w, string(b))
          return
      }
      fmt.Fprintln(w, err)
  }
  ```
- [x] add `encoding/json` and `io` imports
- [x] run `go test ./...` - must pass before Task 6

### Task 6: Update README

- [x] add a short subsection documenting JSON error envelope shape (`{"error": "<message>"}` written to stderr when `--json` is set, exit code 1)
- [x] note that empty list results print `(no results)` in table mode and `[]` in JSON mode
- [x] no test changes for this task

### Task 7: Verify acceptance criteria
- [x] `make build` passes (linter + go build)
- [x] `go test -race ./...` passes
- [x] manual smoke test: `./linear-cli issue list --team NONEXISTENT --json 2>&1 1>/dev/null | jq -e .error` returns non-zero only if error not parseable (skipped - requires live API auth)
- [x] no regression on non-empty list outputs in any list command test
- [x] coverage report still above project baseline (skipped - no baseline recorded; full suite green with -race)

## Technical Details

### Error envelope shape
```json
{"error": "team \"NONEXISTENT\" not found"}
```
- Single string field, lowercase key
- Written to stderr with trailing newline
- Exit code remains 1
- No nested fields, no error code, no stack — keep minimal (YAGNI)
- Plain-text fallback unchanged when `--json` is not set

### Nil-slice handling
- `PaginateAll` is the root cause for list queries — fix at source.
- `JSONFormatter` reflection check is defense-in-depth for direct query handlers (singletons returning nil arrays from GraphQL).
- Both fixes safe: they only affect the empty-collection case; non-empty paths untouched.

### Empty table output
- `TableFormatter` prints literal `(no results)` plus newline to its writer (stdout via `cmd.OutOrStdout()`).
- JSON mode never reaches `TableFormatter` — empty becomes `[]\n` from JSONFormatter.

### Testability of `main.go`
- `main()` becomes thin wrapper.
- Testable helper `formatExecError(err, jsonMode, w)` gets full coverage including escaping edge cases.

## Post-Completion
*Items requiring manual intervention or external systems - no checkboxes, informational only*

**Manual verification** (out of plan):
- Live API smoke: `./linear-cli auth status --json` (success path), `./linear-cli issue list --team ZZZ --json 2>&1` (error path), `./linear-cli comment list ENG-NEW --json` (empty path).
- Pipe through `jq`: `./linear-cli ... --json 2>&1 | jq .` should always parse.

**External system updates** (out of plan):
- Cut new release tag.
- Update Homebrew cask formula `linear-cli` SHA256 and version.
- Note in changelog: error output shape changed under `--json` (potential script breakage if anyone consumed plain-text errors).
