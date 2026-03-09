package model

// ProjectMilestone represents a milestone within a project.
type ProjectMilestone struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	TargetDate  *string `json:"targetDate,omitempty"`
	SortOrder   float64 `json:"sortOrder"`
	Status      string  `json:"status"`
}

// ProjectMilestoneConnection wraps a list of project milestones (relay connection).
type ProjectMilestoneConnection struct {
	Nodes []ProjectMilestone `json:"nodes"`
}
