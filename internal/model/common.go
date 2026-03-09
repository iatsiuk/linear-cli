package model

// WorkflowState represents a Linear workflow state.
type WorkflowState struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Color       string   `json:"color"`
	Type        string   `json:"type"`
	Description *string  `json:"description,omitempty"`
	Position    *float64 `json:"position,omitempty"`
	Team        *Team    `json:"team,omitempty"`
	CreatedAt   string   `json:"createdAt,omitempty"`
}

// User represents a Linear user.
type User struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	Active      bool    `json:"active"`
	Admin       bool    `json:"admin"`
	Guest       bool    `json:"guest"`
	IsMe        bool    `json:"isMe"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// Team represents a Linear team.
type Team struct {
	ID                  string  `json:"id"`
	Name                string  `json:"name"`
	DisplayName         string  `json:"displayName"`
	Description         *string `json:"description,omitempty"`
	Icon                *string `json:"icon,omitempty"`
	Color               *string `json:"color,omitempty"`
	Key                 string  `json:"key"`
	CyclesEnabled       bool    `json:"cyclesEnabled"`
	IssueEstimationType string  `json:"issueEstimationType"`
	CreatedAt           string  `json:"createdAt"`
	UpdatedAt           string  `json:"updatedAt"`
}

// IssueLabel represents a Linear issue label.
type IssueLabel struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Color       string      `json:"color"`
	Description *string     `json:"description,omitempty"`
	IsGroup     bool        `json:"isGroup"`
	Team        *Team       `json:"team,omitempty"`
	Parent      *IssueLabel `json:"parent,omitempty"`
	CreatedAt   string      `json:"createdAt,omitempty"`
}
