package query

// commentFields is the common field selection for Comment.
const commentFields = `
	id
	body
	createdAt
	updatedAt
	editedAt
	url
	user { id displayName email }
	parent { id body createdAt updatedAt url }
`

// CommentListQuery fetches comments for an issue.
const CommentListQuery = `
query CommentList($issueId: String!, $first: Int, $after: String) {
	issue(id: $issueId) {
		comments(first: $first, after: $after) {
			nodes {` + commentFields + `}
			pageInfo { hasNextPage endCursor }
		}
	}
}
`

// CommentCreateMutation creates a new comment.
const CommentCreateMutation = `
mutation CommentCreate($input: CommentCreateInput!) {
	commentCreate(input: $input) {
		success
		comment {` + commentFields + `}
	}
}
`
