# Linear GraphQL API Reference

## Endpoint

```
POST https://api.linear.app/graphql
Content-Type: application/json
```

The API supports introspection. The full schema is available locally at
[`schema.graphql`](schema.graphql) (~40k lines, ~160 query fields, ~380 mutations).

No API versioning. Deprecations are communicated via `@deprecated` directives in
the schema and developer notifications.

## Authentication

### Personal API Key

For scripts and personal tooling. Keys can be scoped to specific permissions
(Read, Write, Admin, Create issues, Create comments) and restricted to specific teams.

```
Authorization: <API_KEY>
```

Verify your key works with `{ viewer { id name email } }` -- this is the
recommended auth smoke test.

### OAuth2

OAuth2 also supported for multi-user apps (see https://linear.app/developers).
OAuth2 tokens require the `Bearer` prefix:

```
Authorization: Bearer <TOKEN>
```

## Viewer Context

Omitting `assigneeId` in `issueCreate` does NOT assign to the current user --
the issue goes to unassigned. To implement `--assign me` in CLI, first fetch
the current user's ID:

```graphql
query { viewer { id } }
```

Then pass it as `assigneeId` in the mutation input.

To resolve the current user's teams (pick first or let user choose):

```graphql
query { viewer { teams { nodes { id key name } } } }
```

### Viewer Shortcuts

The `User` type (returned by `viewer`) has convenience fields that are more
efficient than filter-based queries for common CLI operations:

```graphql
query {
  viewer {
    assignedIssues { nodes { id identifier title } }
    createdIssues { nodes { id identifier title } }
    delegatedIssues { nodes { id identifier title } }
  }
}
```

## Making Requests

```sh
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: <API_KEY>" \
  --data '{ "query": "{ viewer { id name email } }" }' \
  https://api.linear.app/graphql
```

Request body is JSON with `query` (string) and optional `variables` (object):

```json
{
  "query": "query($id: String!) { issue(id: $id) { title } }",
  "variables": { "id": "ENG-123" }
}
```

## Data Model

### Issue Priority

Priority values (see Type Gotchas for type mismatch: `Float!` on query, `Int` on input):

| Value | Meaning |
|-------|---------|
| 0 | No priority |
| 1 | Urgent |
| 2 | High |
| 3 | Normal |
| 4 | Low |

### Workflow State Types

The `WorkflowState.type` field is a plain `String!` (not a GraphQL enum) with these values:

| Type | Description |
|------|-------------|
| `triage` | needs triage |
| `backlog` | in backlog |
| `unstarted` | not yet started |
| `started` | in progress |
| `completed` | done |
| `canceled` | canceled |

### Issue Identifier

Issues have a human-readable `identifier` field in `TEAM-NUMBER` format
(e.g., `ENG-123`). Most issue-related fields accept both UUID and identifier.
Root mutation `id` params (`issueUpdate(id)`, `issueArchive(id)`,
`issueDelete(id)`) likely support identifiers too, but the schema doesn't
explicitly state it -- the `id: String!` parameter is typed as plain string.

### DateTime Format

ISO 8601 format. `DateTime` inputs accept UTC (`Z` suffix) and timezone offsets
(`+02:00`). Go formatting: `time.RFC3339`.

`TimelessDate` is date-only: `YYYY-MM-DD` (no timezone component).
Go formatting: `time.DateOnly` / `"2006-01-02"`.

Date comparator filters also accept:
- shortcuts like `"2021"` (midnight Jan 1, 2021)
- ISO 8601 durations relative to now: `"P2W"` (2 weeks from now),
  `"-P2W1D"` (2 weeks and 1 day ago)

### Markdown in Content

Issue descriptions and comments support markdown. Special features:

- Mentions via plain URLs: `https://linear.app/workspace/issue/ABC-123/title`
- Collapsible sections: `+++ Title\nContent\n+++`
- Rich content: use `descriptionData` (input, JSON) / `descriptionState` (output, String)
  for issues, `content` (input/output, String) for documents.
  These are `[Internal]` -- use plain `description` / `content` markdown fields for CLI

### Trashed vs Archived

Entities can be in two "soft-deleted" states:

- **Archived**: hidden from default queries. Use `includeArchived: true` to include.
  Restore with `*Unarchive` mutations.
- **Trashed**: moved to trash (30-day grace period before permanent deletion).
  Use `issueArchive(id, trash: true)` to trash. The `trashed` field exists on
  Issue, Project, Document, Initiative (not all entities).

See the Delete/Archive/Trash matrix in the Mutation Examples section for full details.

Field size limits are not in the schema; handle validation errors at the API boundary.

### Go Scalar Mapping

| GraphQL | Go | Notes |
|---------|-----|-------|
| `ID` / `String` | `string` | |
| `Float` (nullable) | `*float64` | priority/estimate are ints |
| `Float!` (non-null) | `float64` | |
| `Int` (nullable) | `*int` | |
| `Int!` (non-null) | `int` | |
| `Boolean` (nullable) | `*bool` | |
| `Boolean!` (non-null) | `bool` | |
| `DateTime` | `string` or `time.Time` | ISO 8601 |
| `TimelessDate` | `string` | `YYYY-MM-DD` |
| `JSON` / `JSONObject` | `json.RawMessage` | |
| `UUID` | `string` | |

### Handling Nullability and Updates in Go

To unset a nullable field (unassign, remove due date, etc.), explicitly pass
`null` in mutation variables:

```json
{ "variables": { "id": "ENG-123", "input": { "assigneeId": null } } }
```

Omitting a field from the input means "don't change it". Passing `null` means
"clear this value".

Go note: `omitempty` is a compile-time tag, so a single struct cannot dynamically
switch between omit/null/set per field. Practical patterns for tri-state handling:

1. **`map[string]any` for variables** -- build the input map dynamically, only
   including keys that need to be sent. Keys present with `nil` value send `null`.
2. **Separate structs** -- use distinct create/update input structs with different
   `omitempty` tags for each operation.
3. **Custom `MarshalJSON`** -- implement `json.Marshaler` on input structs to
   control which fields are serialized per-call.
4. **Wrapper type** -- use `Optional[T]` with explicit `Set(v)`/`Null()`/`Omit()`
   states (e.g., `github.com/guregu/null` or a custom generic type).

### Type Gotchas

| Field | Query Type | Input Type | Notes |
|-------|-----------|------------|-------|
| `Issue.priority` | `Float!` | `Int` | Always integer values 0-4 |
| `Issue.estimate` | `Float` | `Int` | |
| `Issue.dueDate` | `TimelessDate` | `TimelessDate` | Date-only scalar, not `DateTime`. Format: `YYYY-MM-DD` (Go: `time.DateOnly` / `"2006-01-02"`) |
| `IssueRelation.type` | -- | Create: `IssueRelationType!`, Update: `String` | Type enum on create, plain string on update |
| `Comment.bodyData` | `String!` | `JSON` | Serialized string on output, JSON object on input |
| `Issue.number` | `Float!` | -- | Always integer value |
| `Cycle.number` | `Float!` | -- | Always integer value |
| `Project.updateRemindersHour` | `Float` | `Int` | |
| `Initiative.updateRemindersHour` | `Float` | `Int` | |
| `Team.cycleCooldownTime` | `Float!` | `Int` | |
| `Team.cycleDuration` | `Float!` | `Int` | |

Rich content field naming is inconsistent across types (output vs input):
- `Issue` output: `description` (String), `descriptionState` (String, `[Internal]`)
- `Issue` input: `descriptionData` (JSON, `[Internal]`)
- `Comment` output: `body` (String!), `bodyData` (String!)
- `Comment` input: `bodyData` (JSON)
- `Document` output: `content` (String), `contentState` (String, `[Internal]`)
- `Document` input: `content` (String)

`descriptionState`, `contentState` are `[Internal]` output fields.
`descriptionData` is an `[Internal]` input field (Issue only).
Prefer `description`/`content`/`body` markdown fields for CLI use.

Nullable comparator types in filters: `NullableNumberComparator`,
`NullableTimelessDateComparator`, `NullableDateComparator`, `NullableUserFilter`,
`NullableIssueFilter` -- used for optional fields that support `null: true/false`.

### Mutation Payloads

Most create/update mutations return a payload type with:
- `success: Boolean!` -- whether the operation succeeded
- `lastSyncId: Float!` -- monotonically increasing sync ID. Track the highest
  value seen; use it to detect whether your local state is up-to-date with the
  server. Higher values indicate more recent changes
- entity field -- the created/updated object. Create/update payloads use
  type-specific field names (`issue`, `project`, `comment`), while archive
  payloads use a generic `entity` field name (e.g., `IssueArchivePayload.entity`,
  `ProjectArchivePayload.entity`). Entity field nullability varies: some are
  nullable (`IssuePayload.issue`, `ProjectPayload.project`), others are non-null
  (`CommentPayload.comment!`, `DocumentPayload.document!`,
  `IssueLabelPayload.issueLabel!`). Always check `success` before accessing
  the entity field

**Delete payloads** (`DeletePayload` pattern): return `entityId: String!`,
`lastSyncId: Float!`, `success: Boolean!` -- no entity object, only the ID of
the deleted entity. Examples of mutations returning `DeletePayload`: `commentDelete`,
`attachmentDelete`, `issueLabelDelete`, `webhookDelete`, `customViewDelete`,
`templateDelete`, `teamMembershipDelete`, `projectRelationDelete`,
`issueRelationDelete`, `reactionDelete` (see schema.graphql for full list).

**Archive payloads** (e.g., `IssueArchivePayload`): return the same fields as
create/update payloads, but `entity` is null if the entity was permanently
deleted. See the Delete/Archive/Trash matrix below for entity-specific behavior.

**Exceptions**: some mutations have non-standard payloads (e.g.,
`CreateCsvExportReportPayload`, `ContactPayload` return only `success`).
Check schema for exact payload types.

## Pagination

Standard Relay cursor-based pagination (`first`/`after`, `last`/`before`,
`pageInfo`). Default page size: 50. Recommended: `first: 50` or `first: 100`.
Prefer `nodes` over `edges` unless you need per-item cursors.

Linear-specific connection arguments:
- `includeArchived: Boolean` -- include archived resources (default: `false`)
- `orderBy: PaginationOrderBy` -- `createdAt` (default) or `updatedAt`

### Iterating All Pages

```graphql
query($cursor: String) {
  issues(first: 50, after: $cursor) {
    nodes { id title }
    pageInfo { hasNextPage endCursor }
  }
}
```

Repeat with `variables: { "cursor": "<endCursor>" }` while `hasNextPage` is `true`.

## Filtering

Most paginated collections accept a `filter` argument. Filters use typed comparator
inputs defined in the schema (e.g., `IssueFilter`, `ProjectFilter`).

### Comparators

**Universal (string, numeric, date):**

| Comparator | Description |
|------------|-------------|
| `eq` | equals |
| `neq` | not equals |
| `in` | value in collection |
| `nin` | value not in collection |

**Numeric and date only:**

| Comparator | Description |
|------------|-------------|
| `lt` | less than |
| `lte` | less than or equal |
| `gt` | greater than |
| `gte` | greater than or equal |

**String only:**

| Comparator | Description |
|------------|-------------|
| `eqIgnoreCase` / `neqIgnoreCase` | case-insensitive equality |
| `startsWith` / `notStartsWith` | prefix match |
| `endsWith` / `notEndsWith` | suffix match |
| `contains` / `notContains` | substring match |
| `containsIgnoreCase` / `notContainsIgnoreCase` | case-insensitive substring |
| `startsWithIgnoreCase` | case-insensitive prefix match |
| `containsIgnoreCaseAndAccent` | case and accent insensitive substring |

**Optional fields:**

| Comparator | Description |
|------------|-------------|
| `null: true` | field is null |
| `null: false` | field is not null |

### Logical Operators

All filter fields are combined with AND by default. Use `or` for OR logic:

```graphql
query {
  issues(filter: {
    or: [
      { assignee: { email: { eq: "alice@co.com" } } },
      { assignee: { email: { eq: "bob@co.com" } } }
    ]
  }) {
    nodes { id title }
  }
}
```

### Relationship Filters

One-to-one -- filter by nested fields:

```graphql
filter: { assignee: { email: { eq: "alice@co.com" } } }
```

Many-to-many -- collection filters (like `IssueLabelCollectionFilter`) support
direct field access (implicit `some`) and explicit operators:

```graphql
# implicit match (any label named "Bug")
filter: { labels: { name: { eq: "Bug" } } }

# explicit some -- same as above
filter: { labels: { some: { name: { eq: "Bug" } } } }

# every -- all labels must match
filter: { labels: { every: { name: { eq: "Bug" } } } }

# length -- check collection size
filter: { labels: { length: { eq: 0 } } }
```

### Relative Date Filters

Date fields accept ISO 8601 durations relative to now:

```graphql
# issues due within 2 weeks
filter: { dueDate: { lt: "P2W" } }

# issues completed in past 2 weeks
filter: { completedAt: { gt: "-P2W" } }
```

### IssueFilter Fields

The most commonly used filter fields beyond `assignee` and `labels`:

| Field | Type | Example |
|-------|------|---------|
| `state` | WorkflowStateFilter | `state: { type: { eq: "started" } }` |
| `team` | TeamFilter | `team: { key: { eq: "ENG" } }` |
| `project` | NullableProjectFilter | `project: { name: { eq: "MVP" } }` |
| `cycle` | NullableCycleFilter | `cycle: { name: { eq: "Sprint 1" } }` |
| `priority` | NullableNumberComparator | `priority: { lte: 2 }` |
| `estimate` | EstimateComparator | `estimate: { gte: 3 }` |
| `dueDate` | NullableTimelessDateComparator | `dueDate: { lt: "P2W" }` |
| `createdAt` | DateComparator | `createdAt: { gt: "-P30D" }` |
| `updatedAt` | DateComparator | `updatedAt: { gt: "-P7D" }` |
| `completedAt` | NullableDateComparator | `completedAt: { gt: "-P2W" }` |
| `creator` | NullableUserFilter | `creator: { email: { eq: "..." } }` |
| `assignee` (isMe) | NullableUserFilter | `assignee: { isMe: { eq: true } }` -- filter by current user without resolving viewer ID |
| `parent` | NullableIssueFilter | `parent: { null: true }` (top-level only) |

Also available: `canceledAt`, `startedAt`, `snoozedUntilAt`, `attachments`,
`hasBlockedByRelations`, `hasBlockingRelations`, `hasDuplicateRelations`,
`hasRelatedRelations`.

### ProjectFilter Fields

Key fields: `health`, `status`, `lead`, `accessibleTeams`, `members`, `startDate`, `targetDate`,
`completedAt`, `canceledAt`, `priority`.

### CycleFilter Fields

Key fields: `isActive`, `isFuture`, `isNext`, `isPast`, `startsAt`, `endsAt`,
`team`, `name`.

### CommentFilter Fields

Key fields: `body`, `issue`, `user`, `parent`, `projectUpdate`, `documentContent`.

## Ordering and Sorting

### `orderBy` -- pagination ordering

The `orderBy` connection argument controls cursor/pagination ordering:
`createdAt` (default) or `updatedAt`.

```graphql
query {
  issues(orderBy: updatedAt) {
    nodes { id title updatedAt }
  }
}
```

### `sort` -- result sorting

The `sort` argument provides richer sorting by multiple fields. Available on
`issues`, `projects`, `users`, and other collections.

```graphql
query {
  issues(sort: [{ createdAt: { order: Descending } }]) {
    nodes { id title createdAt }
  }
}
```

```graphql
query {
  projects(sort: [{ name: { order: Ascending } }]) {
    nodes { id name }
  }
}
```

Sort input types and common fields:
- `IssueSortInput`: `createdAt`, `updatedAt`, `priority`, `dueDate`, `title`, `estimate`, `completedAt`
- `ProjectSortInput`: `name`, `createdAt`, `updatedAt`, `startDate`, `targetDate`
- `UserSortInput`, etc. -- check `schema.graphql` for full list.

Note: the `customers` query uses `sorts` (plural) instead of `sort`. Argument
naming is not 100% consistent across all collections.

## Resolving CLI Inputs to IDs

Common patterns for resolving human-readable inputs to Linear UUIDs:

**Team key -> team ID:**
```graphql
teams(filter: { key: { eq: "ENG" } }) { nodes { id } }
```

**User by email:**
```graphql
users(filter: { email: { eq: "alice@company.com" } }) { nodes { id } }
```

**User by name:**
```graphql
users(filter: { name: { containsIgnoreCase: "alice" } }) { nodes { id } }
```

**Workflow state by team + type/name:**
```graphql
workflowStates(filter: {
  team: { key: { eq: "ENG" } }
  name: { eqIgnoreCase: "In Progress" }
}) { nodes { id name type } }
```

**Label by name:**
```graphql
issueLabels(filter: { name: { eqIgnoreCase: "bug" } }) { nodes { id name } }
```

**Project by name:**
```graphql
projects(filter: { name: { containsIgnoreCase: "onboarding" } }) { nodes { id name } }
```

**Project statuses:**
```graphql
projectStatuses { nodes { id name type } }
```

**Milestones by project + name:**
```graphql
projectMilestones(filter: {
  project: { name: { eq: "MVP" } }
  name: { containsIgnoreCase: "launch" }
}) { nodes { id name targetDate } }
```

**Cycles by team (active/next):**
```graphql
query {
  team(id: "<TEAM_ID>") {
    activeCycle { id name startsAt endsAt }
    cycles(filter: { isNext: { eq: true } }, first: 1) {
      nodes { id name startsAt endsAt }
    }
  }
}
```

**Identifier format support matrix:**

| Operation | UUID | Human-readable (`ENG-123`) | URL identifier | Special |
|-----------|------|---------------------------|----------------|---------|
| Issue mutations (single) | yes | yes | -- | -- |
| `issueBatchUpdate` | yes | **no** | -- | -- |
| `userUpdate` | yes | -- | -- | `"me"` = current user |
| `documentUpdate` | yes | -- | yes | -- |
| `projectUpdate`, `projectArchive`, `projectUnarchive` | yes | -- | yes | -- |
| `projectMilestoneUpdate` | yes | -- | yes | -- |

## Query Examples

### Common Queries

| Query | Pattern |
|-------|---------|
| Current user | `{ viewer { id name email } }` |
| Teams | `{ teams { nodes { id name key } } }` |
| Users | `{ users { nodes { id name email active admin } } }` |
| Team members | `{ team(id: "...") { members { nodes { id name email } } } }` |
| Issues by team | `{ team(id: "...") { issues { nodes { id identifier title state { name } } } } }` |

Use `users(includeDisabled: true)` to include suspended/disabled users (default: false).

### Single Issue (by identifier)

```graphql
query GetIssue($id: String!) {
  issue(id: $id) {
    id title description
    state { name }
    assignee { name }
    labels { nodes { name } }
    project { name }
  }
}
```

```json
{ "variables": { "id": "ENG-123" } }
```

Note: `issue(id)` returns `Issue!` (non-null). When the entity is not found,
`data` itself becomes `null` with an error in the `errors` array -- see
Error Handling section for details.

Additional useful Issue fields (not shown above):
`url`, `branchName`, `number` (Float!), `priorityLabel` ("Urgent"/"High"/...),
`creator { name }`, `estimate` (Float), `dueDate`, `sortOrder`,
`completedAt`, `canceledAt`, `startedAt`, `snoozedUntilAt`,
`trashed` (Boolean, nullable), `previousIdentifiers` (old identifiers after team move),
`customerTicketCount`, `inverseRelations` (issues blocked by this one).

Note: `estimate` and `priority` have type mismatches between query and input -- see Type Gotchas.

### Issue with Relationships

```graphql
query GetIssueRelationships($id: String!) {
  issue(id: $id) {
    id title
    parent { id identifier title }
    children { nodes { id identifier title state { name } } }
    relations { nodes { id type relatedIssue { identifier title } } }
    attachments { nodes { id title url } }
    subscribers { nodes { id name } }
    history(first: 10) {
      nodes {
        id createdAt
        fromStateId toStateId
        fromAssigneeId toAssigneeId
        addedLabelIds removedLabelIds
        actor { name }
      }
    }
  }
}
```

```json
{ "variables": { "id": "ENG-123" } }
```

### Issues Assigned to User

```graphql
query {
  issues(filter: {
    assignee: { email: { eq: "alice@co.com" } }
  }) {
    nodes { id identifier title state { name } }
  }
}
```

### Search Issues

Returns `IssueSearchPayload` with `nodes: [IssueSearchResult!]!` (not `Issue`).
`IssueSearchResult` implements all `Issue` fields directly alongside search-specific
fields (implements `Node`).

```graphql
query {
  searchIssues(term: "login bug", first: 20) {
    nodes { id identifier title state { name } }
    pageInfo { hasNextPage endCursor }
  }
}
```

Additional parameters: `filter: IssueFilter` (filter results),
`includeComments: Boolean` (search in comments, default: false),
`teamId: String` (boost results for a specific team).

Also available: `searchProjects(term)` (returns `ProjectSearchResult`),
`searchDocuments(term)` (returns `DocumentSearchResult`),
`semanticSearch(query)` (natural language / AI search).

### Workflow States

```graphql
query {
  workflowStates {
    nodes { id name type team { name } }
  }
}
```

### Projects

```graphql
query {
  projects {
    nodes { id name status { name } startDate targetDate }
  }
}
```

Additional useful Project fields: `url`, `slugId`, `description`, `progress` (Float 0-1),
`health` (ProjectUpdateHealthType), `lead { name }`, `members { nodes { name } }`,
`teams { nodes { name } }`, `projectMilestones { nodes { id name } }`,
`completedAt`, `canceledAt`.

**Project status**: `status` is a `ProjectStatus` object (not a plain string)
with `type: ProjectStatusType!`. `ProjectStatusType` values: `backlog`,
`planned`, `started`, `paused`, `completed`, `canceled`.

**Project health**: `ProjectUpdateHealthType` values: `onTrack`, `atRisk`,
`offTrack`.

### Cycles

```graphql
query {
  cycles {
    nodes { id name startsAt endsAt }  # name is nullable (String)
  }
}
```

### Documents

```graphql
query {
  documents {
    nodes { id title content project { name } }
  }
}
```

### Labels

```graphql
query {
  issueLabels {
    nodes { id name color parent { name } }
  }
}
```

### Single-Item Queries

All major entities support lookup by ID: `project(id)`, `cycle(id)`,
`document(id)`, `user(id)`, `issueLabel(id)`, `workflowState(id)`,
`notification(id)`, `comment(id)`, `attachment(id)`. Same pattern as `issue(id)`.

### Lookup: Issue by Git Branch

```graphql
query {
  issueVcsBranchSearch(branchName: "feature/eng-123-login-fix") {
    id identifier title state { name }
  }
}
```

### Issue Priority Values

```graphql
query {
  issuePriorityValues { priority label }
}
```

### Template Queries

```graphql
query { templates { id name type } }
query { template(id: "<TEMPLATE_ID>") { id name type templateData } }
```

Note: the root `templates` query returns `[Template!]!` (plain list, not a
connection -- no pagination). In contrast, `Organization.templates` returns
`TemplateConnection!` (paginated connection with `first`/`after` etc.).

### Document Content History

```graphql
query {
  documentContentHistory(id: "<DOCUMENT_ID>") {
    success
    history { id actorIds createdAt contentData contentDataSnapshotAt }
  }
}
```

## Mutation Examples

### Required Fields Cheat Sheet

| Input Type | Required Fields |
|------------|----------------|
| `IssueCreateInput` | `teamId: String!` (all others optional, including `title`) |
| `IssueUpdateInput` | all fields optional |
| `ProjectCreateInput` | `name: String!`, `teamIds: [String!]!` |
| `CycleCreateInput` | `teamId: String!`, `startsAt: DateTime!`, `endsAt: DateTime!` |
| `DocumentCreateInput` | `title: String!` (projectId, content optional) |
| `IssueLabelCreateInput` | `name: String!`, `color: String` (teamId optional -- omit for workspace label) |
| `CommentCreateInput` | all fields optional (but `body` or `bodyData` should be provided in practice) |

### Create Issue

```graphql
mutation CreateIssue($input: IssueCreateInput!) {
  issueCreate(input: $input) {
    success
    issue { id identifier title url }
  }
}
```

```json
{
  "variables": {
    "input": {
      "title": "Bug report",
      "description": "Detailed description in markdown",
      "teamId": "<TEAM_ID>",
      "assigneeId": "<USER_ID>",
      "labelIds": ["<LABEL_ID>"],
      "priority": 2
    }
  }
}
```

Without `stateId`, the issue is assigned to the team's first Backlog state (or
Triage if enabled).

Note: `title` is technically optional (`String`, not `String!`) in `IssueCreateInput`.

Additional `IssueCreateInput` fields: `projectId`, `cycleId`, `parentId` (sub-issue),
`projectMilestoneId`, `dueDate`, `estimate`, `stateId`, `subscriberIds`,
`templateId`, `useDefaultTemplate`, `delegateId`, `createdAt`, `completedAt`,
`sourceCommentId`, `referenceCommentId`, `slaType`.

### Update Issue

```graphql
mutation UpdateIssue($id: String!, $input: IssueUpdateInput!) {
  issueUpdate(id: $id, input: $input) {
    success
    issue { id title state { name } }
  }
}
```

```json
{
  "variables": {
    "id": "ENG-123",
    "input": {
      "title": "Updated title",
      "stateId": "<STATE_ID>",
      "assigneeId": "<USER_ID>"
    }
  }
}
```

`IssueUpdateInput` also supports `addedLabelIds` and `removedLabelIds` for
incremental label changes (instead of replacing all `labelIds`).

Additional `IssueUpdateInput` fields: `teamId` (move issue to another team),
`trashed`, `snoozedById`, `subscriberIds`.

### Delete / Archive Issue

```graphql
mutation {
  issueDelete(id: "ENG-123") { success }
}
# issueDelete = soft-delete (trash) by default
# optional: permanentlyDelete: Boolean -- permanent delete, skip 30-day grace period (admin only)
# returns IssueArchivePayload -- entity is null after delete

mutation {
  issueArchive(id: "ENG-123") { success }
}
# optional: trash: Boolean -- move to trash instead of archiving
# issueArchive(trash: true) is equivalent to issueDelete (both move to trash)

mutation {
  issueUnarchive(id: "ENG-123") { success }
}
```

**Delete/Archive/Trash matrix for core entities:**

| Entity | Archive | Trash | Delete | Restore | `trashed` field |
|--------|---------|-------|--------|---------|----------------|
| Issue | `issueArchive` | `issueArchive(trash: true)` | `issueDelete` | `issueUnarchive` | yes |
| Project | deprecated | deprecated `projectArchive(trash: true)` | `projectDelete` | `projectUnarchive` | yes |
| Document | -- | `documentDelete` (30-day grace) | -- | `documentUnarchive` | yes |

`documentDelete` moves to trash (30-day grace period), not permanent deletion.
Use `documentUnarchive` to restore a trashed document.

`issueDelete` returns `IssueArchivePayload!` (not a generic delete payload).
The `entity` field is null after deletion.

### Batch Operations

```graphql
mutation {
  issueBatchCreate(input: {
    issues: [
      { title: "Issue 1", teamId: "<TEAM_ID>" }
      { title: "Issue 2", teamId: "<TEAM_ID>" }
    ]
  }) {
    success
    issues { id identifier }
  }
}
```

Also available: `issueBatchUpdate` -- note: accepts `ids: [UUID!]!` (only UUIDs,
not human-readable identifiers like `ENG-123`). Maximum 50 items per batch.

### Issue Labels

```graphql
mutation {
  issueAddLabel(id: "ENG-123", labelId: "<LABEL_ID>") { success }
}

mutation {
  issueRemoveLabel(id: "ENG-123", labelId: "<LABEL_ID>") { success }
}
```

### Issue Relations

`IssueRelationType` enum values: `blocks`, `duplicate`, `related`, `similar`.

```graphql
mutation {
  issueRelationCreate(input: {
    issueId: "<ISSUE_ID>"
    relatedIssueId: "<RELATED_ISSUE_ID>"
    type: blocks
  }) {
    success
    issueRelation { id type }
  }
}
```

Also available: `issueRelationUpdate`, `issueRelationDelete`.

### Comments

```graphql
query {
  comment(id: "<COMMENT_ID>") {
    id body user { name } createdAt
    parent { id }
    children { nodes { id body user { name } } }
    resolvedAt
  }
}
```

```graphql
mutation CreateComment($input: CommentCreateInput!) {
  commentCreate(input: $input) {
    success
    comment { id body }
  }
}
```

```json
{ "variables": { "input": { "issueId": "<ISSUE_ID>", "body": "Comment in markdown" } } }
```

Note: `issueId` is optional (`String`, not `String!`) -- comments can also
target `documentContentId`, `projectUpdateId`, `initiativeUpdateId`, or
`postId` instead.

Additional `CommentCreateInput` fields: `parentId` (threaded reply),
`projectUpdateId` (comment on project update), `documentContentId` (comment on
document), `postId` (comment on post), `createAsUser` (create as another user,
OAuth apps), `displayIconUrl` (external user avatar), `initiativeUpdateId`,
`doNotSubscribeToIssue`, `quotedText`, `createdAt`.

Also available: `commentUpdate`, `commentDelete`, `commentResolve`,
`commentUnresolve`.

### Create Project

```graphql
mutation {
  projectCreate(input: {
    name: "Project name"
    teamIds: ["<TEAM_ID>"]
  }) {
    success
    project { id name url }
  }
}
```

Additional `ProjectCreateInput` fields: `leadId`, `memberIds`, `statusId`,
`priority`, `labelIds`, `startDate`, `targetDate`, `description`, `templateId`.

Additional `ProjectUpdateInput` fields: `statusId`, `leadId`, `memberIds`,
`teamIds`, `completedAt`, `canceledAt`, `trashed`.

Also available: `projectUpdate`, `projectDelete`, `projectUnarchive`,
`projectAddLabel(id, labelId)`, `projectRemoveLabel(id, labelId)`.
`projectArchive` is **deprecated** -- use `projectDelete` instead.

### Label Management

Single label: `issueLabel(id)`.

```graphql
mutation {
  issueLabelCreate(input: {
    name: "critical"
    color: "#ff0000"
    teamId: "<TEAM_ID>"
  }) {
    success
    issueLabel { id name }
  }
}
```

Also available: `issueLabelUpdate`, `issueLabelDelete`, `issueLabelRestore`,
`issueLabelRetire`.

Labels can be workspace-level (no `teamId`) or team-level (with `teamId`).
When resolving label names, filter by team to avoid ambiguity:
`issueLabels(filter: { team: { key: { eq: "ENG" } }, name: { eqIgnoreCase: "bug" } })`.

### Cycle Management

```graphql
mutation {
  cycleCreate(input: {
    teamId: "<TEAM_ID>"
    name: "Sprint 1"
    startsAt: "2026-03-09T00:00:00Z"
    endsAt: "2026-03-23T00:00:00Z"
  }) {
    success
    cycle { id name }
  }
}
```

Also available: `cycleUpdate`, `cycleArchive`, `cycleShiftAll`,
`cycleStartUpcomingCycleToday`.

### Document Management

```graphql
mutation {
  documentCreate(input: {
    title: "Design doc"
    content: "Markdown content"
    projectId: "<PROJECT_ID>"
  }) {
    success
    document { id title }
  }
}
```

Additional `DocumentUpdateInput` fields: `hiddenAt`, `trashed`, `issueId`, `teamId`.

Also available: `documentUpdate`, `documentDelete`, `documentUnarchive`.

### Workflow State Management

```graphql
mutation {
  workflowStateCreate(input: {
    name: "In Review"
    type: "started"
    teamId: "<TEAM_ID>"
    color: "#f0ad4e"
  }) {
    success
    workflowState { id name type }
  }
}
```

Also available: `workflowStateUpdate`, `workflowStateArchive`.

### File Upload

Two-step process:

1. Request upload URL:

```graphql
mutation {
  fileUpload(
    contentType: "image/png"
    filename: "screenshot.png"
    size: 102400
  ) {
    success
    uploadFile {
      uploadUrl
      assetUrl
      headers {
        key
        value
      }
    }
  }
}
```

2. PUT the file to the returned `uploadUrl` with the provided headers:

```sh
curl -X PUT \
  -H "Content-Type: image/png" \
  -H "<key>: <value>" \
  --data-binary @screenshot.png \
  "<uploadUrl>"
```

Then use `assetUrl` in issue descriptions or `attachmentCreate`.

### Attachment

Single: `attachment(id)`. List: `attachments`.

```graphql
mutation {
  attachmentCreate(input: {
    issueId: "<ISSUE_ID>"
    url: "https://example.com/file.pdf"
    title: "Design spec"
  }) {
    success
    attachment { id title url }
  }
}
```

Also available: `attachmentUpdate`, `attachmentDelete`.

`attachmentCreate` is idempotent: if an attachment with the same `url` + `issueId`
already exists, it updates the existing attachment instead of creating a duplicate.

Integration-specific: `attachmentLinkSlack`, `attachmentLinkDiscord`,
`attachmentLinkGitHubIssue`, `attachmentLinkGitHubPR`, `attachmentLinkURL`,
`attachmentLinkGitLabMR`, `attachmentLinkZendesk`, `attachmentLinkIntercom`,
`attachmentLinkJiraIssue`, `attachmentLinkFront`, `attachmentLinkSalesforce`.

### Team Management

```graphql
mutation {
  teamCreate(input: {
    name: "Backend"
    key: "BE"  # optional -- auto-generated from name if omitted
  }) {
    success
    team { id name key }
  }
}
```

Also available: `teamUpdate`, `teamDelete`, `teamUnarchive`.

### Notification Management

Available mutations: `notificationMarkReadAll`, `notificationMarkUnreadAll`,
`notificationArchiveAll`, `notificationSnoozeAll`, `notificationUnsnoozeAll`,
`notificationUpdate`, `notificationArchive`, `notificationUnarchive`,
`notificationCategoryChannelSubscriptionUpdate`.

### User Management

Available mutations: `userUpdate`, `userSuspend`, `userUnsuspend`,
`userChangeRole(id: String!, role: UserRoleType!): UserAdminPayload!`.

### Organization Invites

Mutations: `organizationInviteCreate(input: OrganizationInviteCreateInput!)`,
`organizationInviteUpdate(id: String!, input: OrganizationInviteUpdateInput!)`,
`organizationInviteDelete(id: String!)`, `resendOrganizationInvite(id: String!)`.

### Other Available Mutations

- Templates: `templateCreate`, `templateUpdate`, `templateDelete`
- Custom views: `customViewCreate`, `customViewUpdate`, `customViewDelete`
- Initiatives: `initiativeCreate`, `initiativeUpdate`, `initiativeArchive`,
  `initiativeDelete`, `initiativeUnarchive`.
  Queries: `initiative(id)`, `initiatives`.
- Roadmaps: **deprecated** -- use initiatives instead
- Releases: `releaseCreate`, `releaseUpdate`, `releaseComplete`
- Issue subscriptions: `issueSubscribe`, `issueUnsubscribe`
- Reminders: `issueReminder`

See [`schema.graphql`](schema.graphql) for full input types and parameters.

### Organization

```graphql
query {
  organization { id name urlKey logoUrl }
}
```

### Team Membership

```graphql
mutation {
  teamMembershipCreate(input: {
    teamId: "<TEAM_ID>"
    userId: "<USER_ID>"
  }) {
    success
    teamMembership { id }
  }
}
```

```graphql
query {
  teamMemberships {
    nodes { id user { name email } team { name key } }
  }
}
```

Also available: `teamMembershipUpdate`, `teamMembershipDelete`.
`TeamMembershipUpdateInput` also accepts `owner: Boolean` (`[Internal]`).

### Project Updates (Status Check-ins)

Not to be confused with the `projectUpdate` mutation. `ProjectUpdate` is a
separate entity representing project status check-ins.

```graphql
query {
  projectUpdates {
    nodes { id body health user { name } project { name } createdAt }
  }
}
```

```graphql
mutation {
  projectUpdateCreate(input: {
    projectId: "<PROJECT_ID>"
    body: "Status update in markdown"
    health: onTrack
  }) {
    success
    projectUpdate { id body health }
  }
}
```

Also available: `projectUpdateUpdate`, `projectUpdateDelete` (**deprecated** --
use `projectUpdateArchive` instead), `projectUpdateArchive`,
`projectUpdateUnarchive`.

### Project Milestones

```graphql
query {
  projectMilestones {
    nodes { id name description targetDate sortOrder }
  }
}
```

```graphql
mutation {
  projectMilestoneCreate(input: {
    projectId: "<PROJECT_ID>"
    name: "MVP"
    description: "Minimum viable product"
    targetDate: "2026-06-01"
  }) {
    success
    projectMilestone { id name }
  }
}
```

Also available: `projectMilestoneUpdate`, `projectMilestoneDelete`,
`projectMilestoneMove`.

### Project Labels, Relations, and Statuses

**Project Labels** -- separate from issue labels:

Queries: `projectLabel(id)`, `projectLabels`.
Mutations: `projectLabelCreate`, `projectLabelUpdate`, `projectLabelDelete`,
`projectLabelRestore`, `projectLabelRetire`.

**Project Relations** -- links between projects:

Queries: `projectRelation(id)`, `projectRelations`.
Mutations: `projectRelationCreate`, `projectRelationUpdate`, `projectRelationDelete`.

**Project Statuses** -- custom project status types:

Queries: `projectStatus(id)`, `projectStatuses`.
Mutations: `projectStatusCreate`, `projectStatusUpdate`, `projectStatusArchive`,
`projectStatusUnarchive`.

### Initiative Updates

Not to be confused with the `initiativeUpdate` mutation. `InitiativeUpdate` is a
separate entity for initiative status check-ins (similar to `ProjectUpdate`).

Queries: `initiativeUpdate(id)`, `initiativeUpdates`.
Mutations: `initiativeUpdateCreate`, `initiativeUpdateUpdate`,
`initiativeUpdateArchive`, `initiativeUpdateUnarchive`.

### Notification Queries

```graphql
query {
  notifications {
    nodes { id type readAt archivedAt createdAt }
  }
}
```

Single: `notification(id)`.

### Notification Subscriptions

Manage subscriptions to notifications for teams, projects, cycles, labels,
custom views, customers, initiatives, and users.

Queries: `notificationSubscription(id)`, `notificationSubscriptions`.
Mutations: `notificationSubscriptionCreate`, `notificationSubscriptionUpdate`,
`notificationSubscriptionDelete` (**deprecated**).

### Favorites

Queries: `favorite(id)`, `favorites`.
Mutations: `favoriteCreate`, `favoriteUpdate`, `favoriteDelete`.

### Reactions

Mutations: `reactionCreate`, `reactionDelete`.

### Other Queries and Mutations

Time schedules, triage responsibilities, audit log, entity external links,
git automation states, CSV export, and releases are available. See
`schema.graphql` for full details.

## Error Handling

### HTTP Status Codes

| Code | Meaning |
|------|---------|
| 200 | success, or partial success (check `errors` array) |
| 400 | malformed query, rate limiting (`"code": "RATELIMITED"`) |
| 401 | invalid or expired token |
| 5xx | server error |

GraphQL queries can partially succeed -- returning HTTP 200 with both `data` and
`errors`:

```json
{
  "data": null,
  "errors": [
    {
      "message": "Entity not found",
      "path": ["issue"],
      "extensions": { "code": "ENTITY_NOT_FOUND" }
    }
  ]
}
```

Note: `issue(id)` returns `Issue!` (non-null). When the entity is not found,
`data` itself becomes `null` (not `data.issue`). This is standard GraphQL
behavior for non-null root fields -- if a non-null field resolves to null,
the error propagates up to the nearest nullable parent (`data`).

Common error codes in `extensions.code`: `ENTITY_NOT_FOUND`, `RATELIMITED`,
`FORBIDDEN` (permission denied), `VALIDATION_ERROR` (input validation),
`GRAPHQL_VALIDATION_FAILED` (malformed query).

## Rate Limiting

### Request Limits

| Type | Limit |
|------|-------|
| API key / OAuth | 5,000 requests/hour |
| Unauthenticated | 60 requests/hour |

Key header: `X-RateLimit-Requests-Reset` (UTC epoch ms) -- use for backoff.
Also available: `X-RateLimit-Requests-Limit`, `X-RateLimit-Requests-Remaining`,
and per-endpoint variants (`X-RateLimit-Endpoint-*`).

### Complexity Limits

| Type | Limit |
|------|-------|
| API key | 250,000 points/hour |
| OAuth app | 2,000,000 points/hour |
| Unauthenticated | 10,000 points/hour |
| Max single query | 10,000 points |

Headers: `X-Complexity`, `X-RateLimit-Complexity-Limit`,
`X-RateLimit-Complexity-Remaining`, `X-RateLimit-Complexity-Reset`.

Keep nesting to 3-4 levels to stay within the 10,000 point per-query limit.

Rate limit errors return HTTP 400 with `"code": "RATELIMITED"` in extensions.

### Rate Limit Status Query

```graphql
query {
  rateLimitStatus {
    identifier
    kind
    limits {
      type
      allowedAmount
      remainingAmount
      requestedAmount
      period
      reset
    }
  }
}
```

### Backoff Strategy

When rate limited, read `X-RateLimit-Requests-Reset` (UTC epoch ms) to know
when the bucket refills. Use exponential backoff starting at 1 second. For
server errors (5xx), retry with jitter after 1-5 seconds.

## Real-Time Updates

The Linear GraphQL API has no `Subscription` type -- real-time updates are not
available via GraphQL subscriptions. Use webhooks for push-based notifications
or poll with `updatedAt` filters for pull-based updates.

## Schema Reference

For full type definitions, input types, and filter types, see
[`schema.graphql`](schema.graphql). Use it as the authoritative source for
exact field names, types, and enum values.

## Resources

- Developer docs: https://linear.app/developers
- GraphQL explorer: https://studio.apollographql.com/public/Linear-API/variant/current/home
