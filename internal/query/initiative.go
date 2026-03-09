package query

const initiativeFields = `
	id
	name
	description
	status
`

// InitiativeListQuery fetches initiatives in the organization.
const InitiativeListQuery = `
query InitiativeList($first: Int) {
	initiatives(first: $first) {
		nodes {` + initiativeFields + `}
	}
}
`

// InitiativeShowQuery fetches a single initiative by ID.
const InitiativeShowQuery = `
query InitiativeShow($id: String!) {
	initiative(id: $id) {` + initiativeFields + `}
}
`

// InitiativeCreateMutation creates a new initiative.
const InitiativeCreateMutation = `
mutation InitiativeCreate($input: InitiativeCreateInput!) {
	initiativeCreate(input: $input) {
		success
		initiative {` + initiativeFields + `}
	}
}
`

// InitiativeUpdateMutation updates an existing initiative.
const InitiativeUpdateMutation = `
mutation InitiativeUpdate($id: String!, $input: InitiativeUpdateInput!) {
	initiativeUpdate(id: $id, input: $input) {
		success
		initiative {` + initiativeFields + `}
	}
}
`

// InitiativeDeleteMutation deletes an initiative.
const InitiativeDeleteMutation = `
mutation InitiativeDelete($id: String!) {
	initiativeDelete(id: $id) {
		success
	}
}
`
