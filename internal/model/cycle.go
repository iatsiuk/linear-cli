package model

// Cycle represents a Linear cycle (sprint).
type Cycle struct {
	ID          string  `json:"id"`
	Name        *string `json:"name,omitempty"`
	Number      float64 `json:"number"`
	Description *string `json:"description,omitempty"`
	StartsAt    string  `json:"startsAt"`
	EndsAt      string  `json:"endsAt"`
	IsActive    bool    `json:"isActive"`
	IsFuture    bool    `json:"isFuture"`
	IsPast      bool    `json:"isPast"`
	Progress    float64 `json:"progress"`
	Team        Team    `json:"team"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

// CycleConnection wraps a list of cycles (relay connection).
type CycleConnection struct {
	Nodes []Cycle `json:"nodes"`
}
