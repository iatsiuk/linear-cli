package query

const milestoneFields = `
	id
	name
	description
	targetDate
	sortOrder
	status
`

// MilestoneListQuery fetches milestones for a specific project.
const MilestoneListQuery = `
query MilestoneList($projectId: String!, $first: Int) {
	project(id: $projectId) {
		projectMilestones(first: $first) {
			nodes {` + milestoneFields + `}
		}
	}
}
`

// MilestoneCreateMutation creates a new project milestone.
const MilestoneCreateMutation = `
mutation MilestoneCreate($input: ProjectMilestoneCreateInput!) {
	projectMilestoneCreate(input: $input) {
		success
		projectMilestone {` + milestoneFields + `}
	}
}
`

// MilestoneUpdateMutation updates an existing project milestone.
const MilestoneUpdateMutation = `
mutation MilestoneUpdate($id: String!, $input: ProjectMilestoneUpdateInput!) {
	projectMilestoneUpdate(id: $id, input: $input) {
		success
		projectMilestone {` + milestoneFields + `}
	}
}
`

// MilestoneDeleteMutation deletes a project milestone.
const MilestoneDeleteMutation = `
mutation MilestoneDelete($id: String!) {
	projectMilestoneDelete(id: $id) {
		success
	}
}
`
