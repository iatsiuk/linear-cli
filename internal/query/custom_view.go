package query

const customViewFields = `
	id
	name
	description
	shared
	modelName
`

// CustomViewListQuery fetches custom views for the user.
const CustomViewListQuery = `
query CustomViewList($first: Int) {
	customViews(first: $first) {
		nodes {` + customViewFields + `}
	}
}
`

// CustomViewShowQuery fetches a single custom view by ID including filter data.
const CustomViewShowQuery = `
query CustomViewShow($id: String!) {
	customView(id: $id) {` + customViewFields + `
		filterData
	}
}
`

// ViewIssuesQuery fetches issues belonging to a custom view.
const ViewIssuesQuery = `
query ViewIssues($id: String!, $first: Int, $orderBy: PaginationOrderBy, $includeArchived: Boolean) {
	customView(id: $id) {
		issues(first: $first, orderBy: $orderBy, includeArchived: $includeArchived) {
			nodes {` + issueListFields + `}
			pageInfo { hasNextPage endCursor }
		}
	}
}
`
