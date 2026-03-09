package query

// ProjectSearchQuery performs full-text search across projects.
const ProjectSearchQuery = `
query SearchProjects($term: String!, $first: Int, $teamId: String) {
	searchProjects(term: $term, first: $first, teamId: $teamId) {
		nodes {` + projectFields + `}
	}
}
`

// DocumentSearchQuery performs full-text search across documents.
const DocumentSearchQuery = `
query SearchDocuments($term: String!, $first: Int, $teamId: String) {
	searchDocuments(term: $term, first: $first, teamId: $teamId) {
		nodes {` + documentListFields + `}
	}
}
`
