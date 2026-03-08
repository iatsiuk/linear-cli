package query

// labelFields is the common field selection for IssueLabel.
const labelFields = `
	id
	name
	color
	description
	isGroup
	createdAt
	team { id name key }
	parent { id name color }
`

// LabelListQuery fetches issue labels with optional filter.
const LabelListQuery = `
query LabelList($first: Int, $after: String, $filter: IssueLabelFilter) {
	issueLabels(first: $first, after: $after, filter: $filter) {
		nodes {` + labelFields + `}
		pageInfo { hasNextPage endCursor }
	}
}
`

// LabelCreateMutation creates a new issue label.
const LabelCreateMutation = `
mutation LabelCreate($input: IssueLabelCreateInput!) {
	issueLabelCreate(input: $input) {
		success
		issueLabel {` + labelFields + `}
	}
}
`

// LabelUpdateMutation updates an existing issue label.
const LabelUpdateMutation = `
mutation LabelUpdate($id: String!, $input: IssueLabelUpdateInput!) {
	issueLabelUpdate(id: $id, input: $input) {
		success
		issueLabel {` + labelFields + `}
	}
}
`
