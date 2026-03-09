package model

// Notification represents a Linear user notification.
type Notification struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"`
	ReadAt     *string `json:"readAt"`
	ArchivedAt *string `json:"archivedAt"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
	Title      string  `json:"title"`
	Subtitle   string  `json:"subtitle"`
	URL        string  `json:"url"`
}

// NotificationConnection wraps a list of notifications (relay connection).
type NotificationConnection struct {
	Nodes []Notification `json:"nodes"`
}
