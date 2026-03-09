package model

// Comment represents a Linear comment on an issue.
type Comment struct {
	ID        string   `json:"id"`
	Body      string   `json:"body"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
	EditedAt  *string  `json:"editedAt,omitempty"`
	URL       string   `json:"url"`
	User      *User    `json:"user,omitempty"`
	Issue     *Issue   `json:"issue,omitempty"`
	Parent    *Comment `json:"parent,omitempty"`
}
