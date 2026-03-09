package model

// IssueRelation represents a relation between two Linear issues.
type IssueRelation struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Issue        Issue  `json:"issue"`
	RelatedIssue Issue  `json:"relatedIssue"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// IssueRelationConnection wraps a list of issue relations (relay connection).
type IssueRelationConnection struct {
	Nodes []IssueRelation `json:"nodes"`
}
