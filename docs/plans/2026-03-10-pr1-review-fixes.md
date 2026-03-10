# PR #1 Review Fixes: Issue Full Fields

## Overview
Fix all confirmed review findings for PR #1 "Feature/issue full fields".
The PR adds ~30 new fields to the Issue model, GraphQL query, and text output.
Confirmed issues: over-fetching via shared `issueFields`, 9 undisplayed fields,
zero tests for new code, stale README, empty PR description.

## Context
- PR branch: `feature/issue-full-fields`
- Schema source of truth: `docs/schema.graphql`
- Files involved:
  - `internal/query/issue.go` -- GraphQL query constants
  - `internal/model/issue.go` -- Issue struct and CycleRef
  - `internal/cmd/issue_show.go` -- detail display logic
  - `internal/query/issue_test.go` -- query field presence tests
  - `internal/model/issue_test.go` -- deserialization tests
  - `internal/cmd/issue_show_test.go` -- display output tests
  - `README.md` -- user-facing docs

## Development Approach
- **Testing approach**: TDD -- write/update tests FIRST, then fix code to pass them
- **Schema verification**: every field type in model/query MUST be verified against
  `docs/schema.graphql` before writing code. Check nullability (`!` suffix),
  scalar types (`String`, `Float`, `Int`, `Boolean`, `DateTime`), and nested types.
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** written BEFORE code changes
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run `make build` after each change (includes linter)

## Schema Verification Rules
Before writing any code that references a GraphQL field:
1. Open `docs/schema.graphql` and find the `type Issue implements Node` block (~line 11963)
2. Confirm field name exists exactly as written
3. Check nullability: `Type!` = non-null (use Go value type), `Type` = nullable (use Go pointer)
4. Map types: `String`/`DateTime` -> `string`/`*string`, `Float` -> `float64`/`*float64`,
   `Int` -> `int`/`*int`, `Boolean` -> `bool`/`*bool`
5. For nested types (e.g. `cycle: Cycle`), check the referenced type's fields too

## Testing Strategy
- **Unit tests**: required for every task, written BEFORE implementation
- Test files: `*_test.go` in same package
- Use table-driven tests, stdlib `testing` only
- Run with `go test -race ./...`

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix

## Implementation Steps

### Task 1: Split `issueFields` into list and detail variants (tests first)

TDD: write tests that assert field separation, then split the constant.

- [x] verify against `docs/schema.graphql` which fields exist on `type Issue`
- [x] in `internal/query/issue_test.go`: add `TestIssueListFieldsCompact` -- assert `issueListFields` contains only: id, identifier, title, description, priority, priorityLabel, estimate, dueDate, url, createdAt, updatedAt, state, assignee, team, labels, parent, project
- [x] in `internal/query/issue_test.go`: add `TestIssueDetailFieldsContainsAll` -- assert `issueDetailFields` contains all fields from list PLUS: number, branchName, trashed, customerTicketCount, all timestamps, all SLA fields, creator, cycle
- [x] in `internal/query/issue_test.go`: update existing `TestIssueFieldsPresence` to reference new constant names
- [x] run tests -- expect failures (constants don't exist yet)
- [x] in `internal/query/issue.go`: split `issueFields` into `issueListFields` (compact) and `issueDetailFields` (full set with all PR fields)
- [x] update query constants: `IssueListQuery`, `IssueSearchQuery`, `IssueBatchUpdateMutation` use `issueListFields`; `IssueGetQuery`, `IssueCreateMutation`, `IssueUpdateMutation`, `IssueBranchQuery` use `issueDetailFields`
- [x] run tests -- must pass
- [x] run `make build` -- must pass

### Task 2: Display all 9 undisplayed fields in `issue show` (tests first)

TDD: write display tests for new fields, then add display code.

Fields to add to display: `slaHighRiskAt`, `slaMediumRiskAt`, `startedTriageAt`,
`snoozedUntilAt`, `addedToCycleAt`, `addedToProjectAt`, `addedToTeamAt`,
`number`, `customerTicketCount`.

- [x] verify field types against `docs/schema.graphql`:
  - `number: Float!` -> display as integer
  - `customerTicketCount: Int!` -> display as integer
  - all 7 timestamps: `DateTime` (nullable) -> display if non-nil
- [x] in `internal/cmd/issue_show_test.go`: extend `makeDetailedIssue()` to include all 9 fields in mock response
- [x] in `internal/cmd/issue_show_test.go`: add test assertions for new fields in output (e.g. "SLA High Risk", "Triage Started", "Number", "Tickets")
- [x] run tests -- expect failures (display code missing)
- [x] in `internal/cmd/issue_show.go`: add `number` and `customerTicketCount` display (non-zero check)
- [x] in `internal/cmd/issue_show.go`: add 7 missing timestamps to the display loop
- [x] run tests -- must pass
- [x] run `make build` -- must pass

### Task 3: Add deserialization tests for all new Issue fields

TDD: write tests to verify JSON -> Go struct mapping for all new fields.

- [x] verify ALL new field types against `docs/schema.graphql` before writing tests
- [x] in `internal/model/issue_test.go`: add `TestIssueDeserialization_NewFields` -- JSON with all new fields (CycleRef with/without name, Creator, BranchName, Trashed true, Number, CustomerTicketCount, all timestamps, all SLA fields)
- [x] in `internal/model/issue_test.go`: add `TestIssueNullableFields_NewFields` -- JSON without optional fields, verify nil pointers for: Creator, Cycle, Trashed, all timestamp pointers, SLA pointers
- [x] in `internal/model/issue_test.go`: add `TestCycleRefDeserialization` -- test CycleRef with name, without name (nil), verify Number as float64
- [x] run tests -- must pass (these test existing PR code)
- [x] run `make build` -- must pass

### Task 4: Add query field presence tests for new fields

- [x] in `internal/query/issue_test.go`: add `TestIssueDetailFieldsContainsCycle` -- assert `issueDetailFields` contains `cycle { id name number }`
- [x] in `internal/query/issue_test.go`: add `TestIssueDetailFieldsContainsCreator` -- assert `issueDetailFields` contains `creator { id displayName email }`
- [x] in `internal/query/issue_test.go`: extend `TestIssueFieldsPresence` to include new fields: number, branchName, trashed, customerTicketCount, archivedAt, canceledAt, completedAt, startedAt, slaType, slaBreachesAt, slaHighRiskAt, slaMediumRiskAt, slaStartedAt, startedTriageAt, snoozedUntilAt, addedToCycleAt, addedToProjectAt, addedToTeamAt
- [x] run tests -- must pass
- [x] run `make build` -- must pass

### Task 5: Add display tests for existing new fields (cycle, creator, branch, trashed)

TDD: verify the existing PR display code is tested.

- [ ] in `internal/cmd/issue_show_test.go`: extend `makeDetailedIssue()` with cycle (id, name, number), creator (id, displayName, email), branchName, trashed=true
- [ ] add assertions for: cycle format "#N Name", creator name, branch name, "Trashed: yes"
- [ ] add test case for issue WITHOUT cycle/creator/branch/trashed -- verify they don't appear in output
- [ ] run tests -- must pass
- [ ] run `make build` -- must pass

### Task 6: Verify acceptance criteria
- [ ] verify all 9 previously undisplayed fields now appear in `issue show`
- [ ] verify `issueListFields` is compact (no detail-only fields)
- [ ] verify `issueDetailFields` contains all fields
- [ ] verify all field types match `docs/schema.graphql` (final check)
- [ ] run full test suite: `go test -race ./...`
- [ ] run `make build` (includes linter)
- [ ] verify test coverage for changed files

### Task 7: [Final] Update documentation
- [ ] update `README.md:156` -- add all new displayed fields to the `issue show` description
- [ ] run `make build` -- must pass

## Technical Details

### Field split strategy
- `issueListFields`: id, identifier, title, description, priority, priorityLabel, estimate, dueDate, url, createdAt, updatedAt, state { id name color type }, assignee { id displayName email }, team { id name key }, labels { nodes { id name color } }, parent { id identifier title }, project { id name }
- `issueDetailFields`: all of the above PLUS: number, branchName, trashed, customerTicketCount, archivedAt, autoArchivedAt, autoClosedAt, canceledAt, completedAt, startedAt, startedTriageAt, triagedAt, snoozedUntilAt, addedToCycleAt, addedToProjectAt, addedToTeamAt, slaBreachesAt, slaHighRiskAt, slaMediumRiskAt, slaStartedAt, slaType, creator { id displayName email }, cycle { id name number }

### Schema type mapping (verified against docs/schema.graphql)
| GraphQL field | Schema type | Go type | Nullable |
|---|---|---|---|
| number | Float! | float64 | no |
| branchName | String! | string | no |
| customerTicketCount | Int! | int | no |
| trashed | Boolean | *bool | yes |
| creator | User | *User | yes |
| cycle | Cycle | *CycleRef | yes |
| slaType | String | *string | yes |
| all timestamps | DateTime | *string | yes |

### Query usage after split
| Query | Fields constant |
|---|---|
| IssueListQuery | issueListFields |
| IssueSearchQuery | issueListFields |
| IssueBatchUpdateMutation | issueListFields |
| IssueGetQuery | issueDetailFields |
| IssueCreateMutation | issueDetailFields |
| IssueUpdateMutation | issueDetailFields |
| IssueBranchQuery | issueDetailFields |

## Post-Completion

**PR updates:**
- Fill in PR #1 description with summary of changes
- Request re-review after fixes are pushed
