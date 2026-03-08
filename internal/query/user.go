package query

const userFields = `
	id
	email
	displayName
	avatarUrl
	active
	admin
	guest
	isMe
	createdAt
	updatedAt
`

const issueShortFields = `
	id
	identifier
	title
	state { id name color type }
	team { id name key }
`

// UserListQuery fetches organization members with optional includeDisabled flag.
const UserListQuery = `
query UserList($first: Int, $after: String, $includeDisabled: Boolean) {
	users(first: $first, after: $after, includeDisabled: $includeDisabled) {
		nodes {` + userFields + `}
		pageInfo { hasNextPage endCursor }
	}
}
`

// UserGetQuery fetches a single user by ID.
const UserGetQuery = `
query UserGet($id: String!) {
	user(id: $id) {` + userFields + `}
}
`

// ViewerQuery fetches the authenticated user with their teams.
const ViewerQuery = `
query Viewer {
	viewer {` + userFields + `
		teams { nodes { id name key } }
	}
}
`

// ViewerAssignedIssuesQuery fetches issues assigned to the current user.
const ViewerAssignedIssuesQuery = `
query ViewerAssignedIssues {
	viewer {
		assignedIssues {
			nodes {` + issueShortFields + `}
		}
	}
}
`

// ViewerCreatedIssuesQuery fetches issues created by the current user.
const ViewerCreatedIssuesQuery = `
query ViewerCreatedIssues {
	viewer {
		createdIssues {
			nodes {` + issueShortFields + `}
		}
	}
}
`
