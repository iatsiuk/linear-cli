package query

// attachmentFields is the common field selection for Attachment.
const attachmentFields = `
	id
	title
	subtitle
	url
	createdAt
	updatedAt
	creator { id displayName email }
`

// AttachmentListQuery fetches attachments for a given issue.
const AttachmentListQuery = `
query AttachmentList($issueId: String!) {
	issue(id: $issueId) {
		attachments(first: 50) {
			nodes {` + attachmentFields + `}
		}
	}
}
`

// AttachmentCreateMutation creates a new attachment (idempotent: same url+issueId = update).
const AttachmentCreateMutation = `
mutation AttachmentCreate($input: AttachmentCreateInput!) {
	attachmentCreate(input: $input) {
		success
		attachment {` + attachmentFields + `
			issue { id identifier title }
		}
	}
}
`

// AttachmentDeleteMutation deletes an attachment by ID.
const AttachmentDeleteMutation = `
mutation AttachmentDelete($id: String!) {
	attachmentDelete(id: $id) {
		success
	}
}
`
