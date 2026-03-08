package model

// ProjectStatus represents a Linear project status.
type ProjectStatus struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// TeamConnection wraps a list of teams (relay connection).
type TeamConnection struct {
	Nodes []Team `json:"nodes"`
}

// Project represents a Linear project.
type Project struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Color       string         `json:"color"`
	Icon        *string        `json:"icon,omitempty"`
	Health      *string        `json:"health,omitempty"`
	Status      ProjectStatus  `json:"status"`
	Progress    float64        `json:"progress"`
	StartDate   *string        `json:"startDate,omitempty"`
	TargetDate  *string        `json:"targetDate,omitempty"`
	Creator     *User          `json:"creator,omitempty"`
	Teams       TeamConnection `json:"teams"`
	URL         string         `json:"url"`
	CreatedAt   string         `json:"createdAt"`
	UpdatedAt   string         `json:"updatedAt"`
}

// ProjectConnection wraps a list of projects (relay connection).
type ProjectConnection struct {
	Nodes []Project `json:"nodes"`
}
