package query

// relationIssueFields is a minimal field selection for Issue inside a relation.
const relationIssueFields = `
	id
	identifier
	title
	priority
	priorityLabel
	url
	createdAt
	updatedAt
	state { id name color type }
	team { id name key }
	labels { nodes { id name color } }
`

// relationFields is the common field selection for IssueRelation.
const relationFields = `
	id
	type
	createdAt
	updatedAt
	issue {` + relationIssueFields + `}
	relatedIssue {` + relationIssueFields + `}
`

// RelationListQuery fetches both outgoing and incoming relations for an issue.
const RelationListQuery = `
query RelationList($issueId: String!) {
	issue(id: $issueId) {
		relations(first: 50) {
			nodes {` + relationFields + `}
		}
		inverseRelations(first: 50) {
			nodes {` + relationFields + `}
		}
	}
}
`

// RelationCreateMutation creates a new issue relation.
const RelationCreateMutation = `
mutation RelationCreate($input: IssueRelationCreateInput!) {
	issueRelationCreate(input: $input) {
		success
		issueRelation {` + relationFields + `}
	}
}
`

// RelationUpdateMutation updates an existing issue relation.
const RelationUpdateMutation = `
mutation RelationUpdate($id: String!, $input: IssueRelationUpdateInput!) {
	issueRelationUpdate(id: $id, input: $input) {
		success
		issueRelation {` + relationFields + `}
	}
}
`

// RelationDeleteMutation deletes an issue relation by ID.
const RelationDeleteMutation = `
mutation RelationDelete($id: String!) {
	issueRelationDelete(id: $id) {
		success
		entityId
	}
}
`
