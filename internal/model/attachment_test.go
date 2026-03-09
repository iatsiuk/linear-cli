package model

import (
	"encoding/json"
	"testing"
)

func TestAttachmentDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "att-1",
		"title": "Screenshot",
		"url": "https://uploads.linear.app/screenshot.png",
		"createdAt": "2026-01-01T10:00:00.000Z",
		"updatedAt": "2026-01-02T12:00:00.000Z",
		"creator": {
			"id": "u1",
			"displayName": "Alice",
			"email": "alice@example.com"
		},
		"issue": {
			"id": "iss-1",
			"identifier": "ENG-42",
			"title": "Bug report"
		}
	}`

	var att Attachment
	if err := json.Unmarshal([]byte(raw), &att); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if att.ID != "att-1" {
		t.Errorf("ID: got %q, want %q", att.ID, "att-1")
	}
	if att.Title != "Screenshot" {
		t.Errorf("Title: got %q", att.Title)
	}
	if att.URL != "https://uploads.linear.app/screenshot.png" {
		t.Errorf("URL: got %q", att.URL)
	}
	if att.CreatedAt != "2026-01-01T10:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", att.CreatedAt)
	}
	if att.UpdatedAt != "2026-01-02T12:00:00.000Z" {
		t.Errorf("UpdatedAt: got %q", att.UpdatedAt)
	}
	if att.Creator == nil {
		t.Fatal("Creator should not be nil")
	}
	if att.Creator.ID != "u1" {
		t.Errorf("Creator.ID: got %q", att.Creator.ID)
	}
	if att.Issue == nil {
		t.Fatal("Issue should not be nil")
	}
	if att.Issue.Identifier != "ENG-42" {
		t.Errorf("Issue.Identifier: got %q", att.Issue.Identifier)
	}
}

func TestAttachmentNullableFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "att-2",
		"title": "Link",
		"url": "https://example.com",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"issue": {
			"id": "iss-2",
			"identifier": "ENG-1",
			"title": "Issue"
		}
	}`

	var att Attachment
	if err := json.Unmarshal([]byte(raw), &att); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if att.Creator != nil {
		t.Errorf("Creator: expected nil, got %v", att.Creator)
	}
	if att.Subtitle != nil {
		t.Errorf("Subtitle: expected nil, got %v", att.Subtitle)
	}
}

func TestAttachmentWithSubtitle(t *testing.T) {
	t.Parallel()

	sub := "branch: main"
	raw := `{
		"id": "att-3",
		"title": "PR #42",
		"url": "https://github.com/org/repo/pull/42",
		"subtitle": "branch: main",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"issue": {
			"id": "iss-3",
			"identifier": "ENG-10",
			"title": "Feature"
		}
	}`

	var att Attachment
	if err := json.Unmarshal([]byte(raw), &att); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if att.Subtitle == nil || *att.Subtitle != sub {
		t.Errorf("Subtitle: got %v, want %q", att.Subtitle, sub)
	}
}
