package query

const teamMembershipFields = `
	id
	owner
	sortOrder
	user {
		id
		displayName
		email
	}
`

// TeamMemberListQuery fetches memberships for a specific team.
const TeamMemberListQuery = `
query TeamMemberList($teamId: String!, $first: Int, $after: String) {
	team(id: $teamId) {
		memberships(first: $first, after: $after) {
			nodes {` + teamMembershipFields + `}
			pageInfo {
				hasNextPage
				endCursor
			}
		}
	}
}
`

// TeamMemberAddMutation creates a new team membership.
const TeamMemberAddMutation = `
mutation TeamMemberAdd($input: TeamMembershipCreateInput!) {
	teamMembershipCreate(input: $input) {
		success
		teamMembership {` + teamMembershipFields + `}
	}
}
`

// TeamMemberRemoveMutation deletes a team membership by its ID.
const TeamMemberRemoveMutation = `
mutation TeamMemberRemove($id: String!) {
	teamMembershipDelete(id: $id) {
		success
	}
}
`
