package query

// issueFields is the common field selection for Issue.
const issueFields = `
	id
	identifier
	title
	description
	priority
	priorityLabel
	estimate
	dueDate
	url
	createdAt
	updatedAt
	state { id name color type }
	assignee { id displayName email }
	team { id name key }
	labels { nodes { id name color } }
`

// IssueListQuery fetches issues with optional pagination and filter.
const IssueListQuery = `
query IssueList($first: Int, $after: String, $filter: IssueFilter, $includeArchived: Boolean, $orderBy: PaginationOrderBy) {
	issues(first: $first, after: $after, filter: $filter, includeArchived: $includeArchived, orderBy: $orderBy) {
		nodes {` + issueFields + `}
		pageInfo { hasNextPage endCursor }
	}
}
`

// IssueGetQuery fetches a single issue by ID.
const IssueGetQuery = `
query IssueGet($id: String!) {
	issue(id: $id) {` + issueFields + `}
}
`

// IssueCreateMutation creates a new issue.
const IssueCreateMutation = `
mutation IssueCreate($input: IssueCreateInput!) {
	issueCreate(input: $input) {
		success
		issue {` + issueFields + `}
	}
}
`

// IssueUpdateMutation updates an existing issue.
const IssueUpdateMutation = `
mutation IssueUpdate($id: String!, $input: IssueUpdateInput!) {
	issueUpdate(id: $id, input: $input) {
		success
		issue {` + issueFields + `}
	}
}
`

// IssueDeleteMutation soft-deletes (trashes) an issue.
const IssueDeleteMutation = `
mutation IssueDelete($id: String!) {
	issueDelete(id: $id) {
		success
	}
}
`

// IssueArchiveMutation archives an issue (hidden from default queries).
const IssueArchiveMutation = `
mutation IssueArchive($id: String!) {
	issueArchive(id: $id) {
		success
	}
}
`

// IssueBatchUpdateMutation updates multiple issues at once (max 50).
const IssueBatchUpdateMutation = `
mutation IssueBatchUpdate($ids: [UUID!]!, $input: IssueUpdateInput!) {
	issueBatchUpdate(ids: $ids, input: $input) {
		issues {` + issueFields + `}
	}
}
`

// IssueSearchQuery performs full-text search across issues.
const IssueSearchQuery = `
query SearchIssues($term: String!, $first: Int, $teamId: String) {
	searchIssues(term: $term, first: $first, teamId: $teamId) {
		nodes {` + issueFields + `}
	}
}
`

// IssueBranchQuery looks up an issue by VCS branch name.
const IssueBranchQuery = `
query IssueBranch($branchName: String!) {
	issueVcsBranchSearch(branchName: $branchName) {` + issueFields + `}
}
`
