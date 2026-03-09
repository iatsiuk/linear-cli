package model

// CustomView represents a Linear custom view (saved filter/sort configuration).
type CustomView struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Shared      bool    `json:"shared"`
	ModelName   string  `json:"modelName"`
}

// CustomViewConnection wraps a list of custom views (relay connection).
type CustomViewConnection struct {
	Nodes []CustomView `json:"nodes"`
}
