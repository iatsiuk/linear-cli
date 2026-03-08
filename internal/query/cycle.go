package query

// cycleFields is the common field selection for Cycle.
const cycleFields = `
	id
	name
	number
	description
	startsAt
	endsAt
	isActive
	isFuture
	isPast
	progress
	team { id name key }
	createdAt
	updatedAt
`

// CycleListQuery fetches cycles with optional pagination and filter.
const CycleListQuery = `
query CycleList($first: Int, $after: String, $filter: CycleFilter, $includeArchived: Boolean, $orderBy: PaginationOrderBy) {
	cycles(first: $first, after: $after, filter: $filter, includeArchived: $includeArchived, orderBy: $orderBy) {
		nodes {` + cycleFields + `}
		pageInfo { hasNextPage endCursor }
	}
}
`

// CycleGetQuery fetches a single cycle by ID.
const CycleGetQuery = `
query CycleGet($id: String!) {
	cycle(id: $id) {` + cycleFields + `}
}
`
