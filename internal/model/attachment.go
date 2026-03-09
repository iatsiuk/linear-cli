package model

// Attachment represents a Linear issue attachment.
type Attachment struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Subtitle  *string `json:"subtitle,omitempty"`
	URL       string  `json:"url"`
	Creator   *User   `json:"creator,omitempty"`
	Issue     *Issue  `json:"issue,omitempty"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
}

// AttachmentConnection wraps a list of attachments (relay connection).
type AttachmentConnection struct {
	Nodes []Attachment `json:"nodes"`
}
