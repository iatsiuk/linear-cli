package query

// projectFields is the common field selection for Project.
const projectFields = `
	id
	name
	description
	color
	icon
	health
	status { id name type }
	progress
	startDate
	targetDate
	creator { id displayName email }
	teams { nodes { id name key } }
	url
	createdAt
	updatedAt
`

// ProjectListQuery fetches projects with optional pagination and filter.
const ProjectListQuery = `
query ProjectList($first: Int, $after: String, $filter: ProjectFilter, $includeArchived: Boolean, $orderBy: PaginationOrderBy) {
	projects(first: $first, after: $after, filter: $filter, includeArchived: $includeArchived, orderBy: $orderBy) {
		nodes {` + projectFields + `}
		pageInfo { hasNextPage endCursor }
	}
}
`

// ProjectGetQuery fetches a single project by ID.
const ProjectGetQuery = `
query ProjectGet($id: String!) {
	project(id: $id) {` + projectFields + `}
}
`

// ProjectCreateMutation creates a new project.
const ProjectCreateMutation = `
mutation ProjectCreate($input: ProjectCreateInput!) {
	projectCreate(input: $input) {
		success
		project {` + projectFields + `}
	}
}
`

// ProjectUpdateMutation updates an existing project.
const ProjectUpdateMutation = `
mutation ProjectUpdate($id: String!, $input: ProjectUpdateInput!) {
	projectUpdate(id: $id, input: $input) {
		success
		project {` + projectFields + `}
	}
}
`

// ProjectDeleteMutation deletes (trashes) a project.
const ProjectDeleteMutation = `
mutation ProjectDelete($id: String!) {
	projectDelete(id: $id) {
		success
	}
}
`
