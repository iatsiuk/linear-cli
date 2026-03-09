package query

// documentFields is the common field selection for Document (includes content).
const documentFields = `
	id
	title
	content
	slugId
	url
	createdAt
	updatedAt
	archivedAt
	hiddenAt
	trashed
	creator { id displayName email }
	project { id name }
`

// documentListFields is the field selection for listing documents (excludes content for efficiency).
const documentListFields = `
	id
	title
	slugId
	url
	createdAt
	updatedAt
	archivedAt
	hiddenAt
	trashed
	creator { id displayName email }
	project { id name }
`

// DocumentListQuery fetches documents, optionally filtered by project.
const DocumentListQuery = `
query DocumentList($first: Int, $after: String, $filter: DocumentFilter, $includeArchived: Boolean) {
	documents(first: $first, after: $after, filter: $filter, includeArchived: $includeArchived) {
		nodes {` + documentListFields + `}
		pageInfo { hasNextPage endCursor }
	}
}
`

// DocumentGetQuery fetches a single document by ID.
const DocumentGetQuery = `
query DocumentGet($id: String!) {
	document(id: $id) {` + documentFields + `}
}
`

// DocumentCreateMutation creates a new document.
const DocumentCreateMutation = `
mutation DocumentCreate($input: DocumentCreateInput!) {
	documentCreate(input: $input) {
		success
		document {` + documentFields + `}
	}
}
`

// DocumentUpdateMutation updates an existing document.
const DocumentUpdateMutation = `
mutation DocumentUpdate($id: String!, $input: DocumentUpdateInput!) {
	documentUpdate(id: $id, input: $input) {
		success
		document {` + documentFields + `}
	}
}
`

// DocumentDeleteMutation moves a document to trash (30-day grace period).
const DocumentDeleteMutation = `
mutation DocumentDelete($id: String!) {
	documentDelete(id: $id) {
		success
	}
}
`

// DocumentUnarchiveMutation restores a document from trash.
const DocumentUnarchiveMutation = `
mutation DocumentUnarchive($id: String!) {
	documentUnarchive(id: $id) {
		success
	}
}
`
