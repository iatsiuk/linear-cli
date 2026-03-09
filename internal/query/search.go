package query

// ProjectSearchQuery performs full-text search across projects.
const ProjectSearchQuery = `
query SearchProjects($term: String!, $first: Int) {
	searchProjects(term: $term, first: $first) {
		nodes {` + projectFields + `}
	}
}
`

// DocumentSearchQuery performs full-text search across documents.
const DocumentSearchQuery = `
query SearchDocuments($term: String!, $first: Int) {
	searchDocuments(term: $term, first: $first) {
		nodes {` + documentListFields + `}
	}
}
`
