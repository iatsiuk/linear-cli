package model

import "encoding/json"

// Template represents a Linear template.
type Template struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Type         string          `json:"type"`
	Description  *string         `json:"description,omitempty"`
	TemplateData json.RawMessage `json:"templateData,omitempty"`
}
