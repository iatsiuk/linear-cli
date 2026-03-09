package model

// Document represents a Linear document.
type Document struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Content    *string  `json:"content,omitempty"`
	SlugID     string   `json:"slugId"`
	URL        string   `json:"url"`
	Creator    *User    `json:"creator,omitempty"`
	Project    *Project `json:"project,omitempty"`
	ArchivedAt *string  `json:"archivedAt,omitempty"`
	HiddenAt   *string  `json:"hiddenAt,omitempty"`
	Trashed    *bool    `json:"trashed,omitempty"`
	CreatedAt  string   `json:"createdAt"`
	UpdatedAt  string   `json:"updatedAt"`
}

// DocumentConnection wraps a list of documents (relay connection).
type DocumentConnection struct {
	Nodes []Document `json:"nodes"`
}
