package query

const projectUpdateFields = `
	id
	body
	health
	user { id displayName email }
	project { id name }
	createdAt
	updatedAt
`

// ProjectUpdateListQuery fetches status check-ins for a specific project.
const ProjectUpdateListQuery = `
query ProjectUpdateList($projectId: String!, $first: Int) {
	project(id: $projectId) {
		projectUpdates(first: $first) {
			nodes {` + projectUpdateFields + `}
		}
	}
}
`

// ProjectUpdateCreateMutation creates a new status check-in for a project.
const ProjectUpdateCreateMutation = `
mutation ProjectUpdateCreate($input: ProjectUpdateCreateInput!) {
	projectUpdateCreate(input: $input) {
		success
		projectUpdate {` + projectUpdateFields + `}
	}
}
`

// ProjectUpdateArchiveMutation archives a status check-in.
const ProjectUpdateArchiveMutation = `
mutation ProjectUpdateArchive($id: String!) {
	projectUpdateArchive(id: $id) {
		success
	}
}
`
