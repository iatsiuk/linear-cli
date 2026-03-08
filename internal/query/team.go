package query

const teamFields = `
	id
	name
	displayName
	description
	icon
	color
	key
	cyclesEnabled
	issueEstimationType
	createdAt
	updatedAt
`

// TeamListQuery fetches all teams with pagination.
const TeamListQuery = `
query TeamList($first: Int, $after: String) {
	teams(first: $first, after: $after) {
		nodes {` + teamFields + `}
		pageInfo { hasNextPage endCursor }
	}
}
`

// TeamGetQuery fetches a single team by ID.
const TeamGetQuery = `
query TeamGet($id: String!) {
	team(id: $id) {` + teamFields + `}
}
`
