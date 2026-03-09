package model

import (
	"encoding/json"
	"testing"
)

func TestDocumentDeserialization(t *testing.T) {
	t.Parallel()

	content := "# Hello\nSome content."
	raw := `{
		"id": "doc-1",
		"title": "My Document",
		"content": "# Hello\nSome content.",
		"slugId": "my-document",
		"url": "https://linear.app/team/doc/my-document-doc-1",
		"createdAt": "2026-01-01T10:00:00.000Z",
		"updatedAt": "2026-01-02T12:00:00.000Z",
		"creator": {
			"id": "u1",
			"displayName": "Alice",
			"email": "alice@example.com"
		},
		"project": {
			"id": "p1",
			"name": "Alpha Project",
			"description": "",
			"color": "#blue",
			"progress": 0.5,
			"url": "https://linear.app/team/project/alpha-p1",
			"createdAt": "2026-01-01T00:00:00.000Z",
			"updatedAt": "2026-01-01T00:00:00.000Z",
			"status": {"id": "ps1", "name": "In Progress", "type": "inProgress"},
			"teams": {"nodes": []}
		}
	}`

	var doc Document
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if doc.ID != "doc-1" {
		t.Errorf("ID: got %q, want %q", doc.ID, "doc-1")
	}
	if doc.Title != "My Document" {
		t.Errorf("Title: got %q", doc.Title)
	}
	if doc.Content == nil || *doc.Content != content {
		t.Errorf("Content: got %v", doc.Content)
	}
	if doc.SlugID != "my-document" {
		t.Errorf("SlugID: got %q", doc.SlugID)
	}
	if doc.URL != "https://linear.app/team/doc/my-document-doc-1" {
		t.Errorf("URL: got %q", doc.URL)
	}
	if doc.CreatedAt != "2026-01-01T10:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", doc.CreatedAt)
	}
	if doc.UpdatedAt != "2026-01-02T12:00:00.000Z" {
		t.Errorf("UpdatedAt: got %q", doc.UpdatedAt)
	}
	if doc.Creator == nil {
		t.Fatal("Creator should not be nil")
	}
	if doc.Creator.ID != "u1" {
		t.Errorf("Creator.ID: got %q", doc.Creator.ID)
	}
	if doc.Project == nil {
		t.Fatal("Project should not be nil")
	}
	if doc.Project.Name != "Alpha Project" {
		t.Errorf("Project.Name: got %q", doc.Project.Name)
	}
}

func TestDocumentNullableFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "doc-2",
		"title": "Minimal Doc",
		"slugId": "minimal-doc",
		"url": "https://linear.app/team/doc/minimal-doc-2",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z"
	}`

	var doc Document
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if doc.Content != nil {
		t.Errorf("Content: expected nil, got %v", doc.Content)
	}
	if doc.Creator != nil {
		t.Errorf("Creator: expected nil, got %v", doc.Creator)
	}
	if doc.Project != nil {
		t.Errorf("Project: expected nil, got %v", doc.Project)
	}
	if doc.ArchivedAt != nil {
		t.Errorf("ArchivedAt: expected nil, got %v", doc.ArchivedAt)
	}
	if doc.HiddenAt != nil {
		t.Errorf("HiddenAt: expected nil, got %v", doc.HiddenAt)
	}
	if doc.Trashed != nil {
		t.Errorf("Trashed: expected nil, got %v", doc.Trashed)
	}
}

func TestDocumentTrashed(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "doc-3",
		"title": "Deleted Doc",
		"slugId": "deleted-doc",
		"url": "https://linear.app/team/doc/deleted-doc-3",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-05T00:00:00.000Z",
		"archivedAt": "2026-01-05T00:00:00.000Z",
		"trashed": true
	}`

	var doc Document
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if doc.Trashed == nil || !*doc.Trashed {
		t.Error("Trashed: expected true")
	}
	if doc.ArchivedAt == nil {
		t.Error("ArchivedAt: expected non-nil")
	}
}
