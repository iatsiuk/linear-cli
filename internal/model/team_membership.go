package model

// TeamMembership represents a user's membership in a Linear team.
type TeamMembership struct {
	ID        string  `json:"id"`
	Owner     bool    `json:"owner"`
	SortOrder float64 `json:"sortOrder"`
	User      User    `json:"user"`
	Team      Team    `json:"team"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// TeamMembershipConnection wraps a paginated list of TeamMembership nodes.
type TeamMembershipConnection struct {
	Nodes []TeamMembership `json:"nodes"`
}
