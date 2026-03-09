package model

// Initiative represents a Linear initiative (replacement for deprecated Roadmaps).
type Initiative struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Status      string  `json:"status"`
}

// InitiativeConnection wraps a list of initiatives (relay connection).
type InitiativeConnection struct {
	Nodes []Initiative `json:"nodes"`
}
