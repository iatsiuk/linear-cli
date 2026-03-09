package model

// ProjectUpdate represents a status check-in for a project.
type ProjectUpdate struct {
	ID        string  `json:"id"`
	Body      string  `json:"body"`
	Health    string  `json:"health"`
	User      User    `json:"user"`
	Project   Project `json:"project"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// ProjectUpdateConnection wraps a list of project updates (relay connection).
type ProjectUpdateConnection struct {
	Nodes []ProjectUpdate `json:"nodes"`
}
