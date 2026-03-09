package query

// stateFields is the common field selection for WorkflowState.
const stateFields = `
	id
	name
	color
	type
	description
	position
	createdAt
	team { id name key }
`

// StateListQuery fetches workflow states with optional filter.
const StateListQuery = `
query StateList($first: Int, $after: String, $filter: WorkflowStateFilter) {
	workflowStates(first: $first, after: $after, filter: $filter) {
		nodes {` + stateFields + `}
		pageInfo { hasNextPage endCursor }
	}
}
`
