package model

// IssueLabelConnection wraps a list of issue labels (relay connection).
type IssueLabelConnection struct {
	Nodes []IssueLabel `json:"nodes"`
}

// IssueRef is a lightweight reference to a parent issue.
type IssueRef struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
}

// ProjectRef is a lightweight reference to a project.
type ProjectRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// CycleRef is a lightweight reference to a cycle.
type CycleRef struct {
	ID     string  `json:"id"`
	Name   *string `json:"name,omitempty"`
	Number float64 `json:"number"`
}

// Issue represents a Linear issue.
type Issue struct {
	ID                  string               `json:"id"`
	Identifier          string               `json:"identifier"`
	Number              float64              `json:"number"`
	Title               string               `json:"title"`
	Description         *string              `json:"description,omitempty"`
	BranchName          string               `json:"branchName"`
	Priority            float64              `json:"priority"`
	PriorityLabel       string               `json:"priorityLabel"`
	Estimate            *float64             `json:"estimate,omitempty"`
	DueDate             *string              `json:"dueDate,omitempty"`
	URL                 string               `json:"url"`
	Trashed             *bool                `json:"trashed,omitempty"`
	CustomerTicketCount int                  `json:"customerTicketCount"`
	CreatedAt           string               `json:"createdAt"`
	UpdatedAt           string               `json:"updatedAt"`
	ArchivedAt          *string              `json:"archivedAt,omitempty"`
	AutoArchivedAt      *string              `json:"autoArchivedAt,omitempty"`
	AutoClosedAt        *string              `json:"autoClosedAt,omitempty"`
	CanceledAt          *string              `json:"canceledAt,omitempty"`
	CompletedAt         *string              `json:"completedAt,omitempty"`
	StartedAt           *string              `json:"startedAt,omitempty"`
	StartedTriageAt     *string              `json:"startedTriageAt,omitempty"`
	TriagedAt           *string              `json:"triagedAt,omitempty"`
	SnoozedUntilAt      *string              `json:"snoozedUntilAt,omitempty"`
	AddedToCycleAt      *string              `json:"addedToCycleAt,omitempty"`
	AddedToProjectAt    *string              `json:"addedToProjectAt,omitempty"`
	AddedToTeamAt       *string              `json:"addedToTeamAt,omitempty"`
	SlaBreachesAt       *string              `json:"slaBreachesAt,omitempty"`
	SlaHighRiskAt       *string              `json:"slaHighRiskAt,omitempty"`
	SlaMediumRiskAt     *string              `json:"slaMediumRiskAt,omitempty"`
	SlaStartedAt        *string              `json:"slaStartedAt,omitempty"`
	SlaType             *string              `json:"slaType,omitempty"`
	State               WorkflowState        `json:"state"`
	Assignee            *User                `json:"assignee,omitempty"`
	Creator             *User                `json:"creator,omitempty"`
	Team                Team                 `json:"team"`
	Labels              IssueLabelConnection `json:"labels"`
	Parent              *IssueRef            `json:"parent,omitempty"`
	Project             *ProjectRef          `json:"project,omitempty"`
	Cycle               *CycleRef            `json:"cycle,omitempty"`
}
