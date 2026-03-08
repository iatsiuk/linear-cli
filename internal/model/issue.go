package model

// IssueLabelConnection wraps a list of issue labels (relay connection).
type IssueLabelConnection struct {
	Nodes []IssueLabel `json:"nodes"`
}

// Issue represents a Linear issue.
type Issue struct {
	ID            string               `json:"id"`
	Identifier    string               `json:"identifier"`
	Title         string               `json:"title"`
	Description   *string              `json:"description,omitempty"`
	Priority      float64              `json:"priority"`
	PriorityLabel string               `json:"priorityLabel"`
	Estimate      *float64             `json:"estimate,omitempty"`
	DueDate       *string              `json:"dueDate,omitempty"`
	URL           string               `json:"url"`
	CreatedAt     string               `json:"createdAt"`
	UpdatedAt     string               `json:"updatedAt"`
	State         WorkflowState        `json:"state"`
	Assignee      *User                `json:"assignee,omitempty"`
	Team          Team                 `json:"team"`
	Labels        IssueLabelConnection `json:"labels"`
}
