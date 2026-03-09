package model

import (
	"encoding/json"
	"testing"
)

func TestCommentDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "cmt-1",
		"body": "This is a comment.",
		"createdAt": "2026-01-01T10:00:00.000Z",
		"updatedAt": "2026-01-01T11:00:00.000Z",
		"editedAt": "2026-01-01T11:00:00.000Z",
		"url": "https://linear.app/issue/ENG-42#comment-cmt-1",
		"user": {
			"id": "u1",
			"displayName": "Alice",
			"email": "alice@example.com"
		}
	}`

	var comment Comment
	if err := json.Unmarshal([]byte(raw), &comment); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if comment.ID != "cmt-1" {
		t.Errorf("ID: got %q, want %q", comment.ID, "cmt-1")
	}
	if comment.Body != "This is a comment." {
		t.Errorf("Body: got %q", comment.Body)
	}
	if comment.CreatedAt != "2026-01-01T10:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", comment.CreatedAt)
	}
	if comment.UpdatedAt != "2026-01-01T11:00:00.000Z" {
		t.Errorf("UpdatedAt: got %q", comment.UpdatedAt)
	}
	if comment.EditedAt == nil || *comment.EditedAt != "2026-01-01T11:00:00.000Z" {
		t.Errorf("EditedAt: unexpected value %v", comment.EditedAt)
	}
	if comment.URL != "https://linear.app/issue/ENG-42#comment-cmt-1" {
		t.Errorf("URL: got %q", comment.URL)
	}
	if comment.User == nil {
		t.Fatal("User should not be nil")
	}
	if comment.User.ID != "u1" {
		t.Errorf("User.ID: got %q", comment.User.ID)
	}
	if comment.User.DisplayName != "Alice" {
		t.Errorf("User.DisplayName: got %q", comment.User.DisplayName)
	}
}

func TestCommentWithParent(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "cmt-2",
		"body": "A reply.",
		"createdAt": "2026-01-02T00:00:00.000Z",
		"updatedAt": "2026-01-02T00:00:00.000Z",
		"url": "https://linear.app/issue/ENG-42#comment-cmt-2",
		"user": {"id": "u2", "displayName": "Bob", "email": "bob@example.com"},
		"parent": {
			"id": "cmt-1",
			"body": "Original comment.",
			"createdAt": "2026-01-01T00:00:00.000Z",
			"updatedAt": "2026-01-01T00:00:00.000Z",
			"url": "https://linear.app/issue/ENG-42#comment-cmt-1"
		}
	}`

	var comment Comment
	if err := json.Unmarshal([]byte(raw), &comment); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if comment.Parent == nil {
		t.Fatal("Parent should not be nil")
	}
	if comment.Parent.ID != "cmt-1" {
		t.Errorf("Parent.ID: got %q, want %q", comment.Parent.ID, "cmt-1")
	}
	if comment.Parent.Body != "Original comment." {
		t.Errorf("Parent.Body: got %q", comment.Parent.Body)
	}
}

func TestCommentNullableFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "cmt-3",
		"body": "Simple comment.",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"url": "https://linear.app/issue/ENG-1#comment-cmt-3"
	}`

	var comment Comment
	if err := json.Unmarshal([]byte(raw), &comment); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if comment.EditedAt != nil {
		t.Errorf("EditedAt: expected nil, got %v", comment.EditedAt)
	}
	if comment.User != nil {
		t.Errorf("User: expected nil, got %v", comment.User)
	}
	if comment.Parent != nil {
		t.Errorf("Parent: expected nil, got %v", comment.Parent)
	}
	if comment.Issue != nil {
		t.Errorf("Issue: expected nil, got %v", comment.Issue)
	}
}

func TestCommentWithIssue(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "cmt-4",
		"body": "Comment with issue ref.",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"url": "https://linear.app/issue/ENG-5#comment-cmt-4",
		"issue": {
			"id": "iss-5",
			"identifier": "ENG-5",
			"title": "Some issue",
			"priority": 0,
			"priorityLabel": "No priority",
			"url": "https://linear.app/issue/ENG-5",
			"createdAt": "2026-01-01T00:00:00.000Z",
			"updatedAt": "2026-01-01T00:00:00.000Z",
			"state": {"id": "s1", "name": "Backlog", "color": "#ccc", "type": "backlog"},
			"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
			"labels": {"nodes": []}
		}
	}`

	var comment Comment
	if err := json.Unmarshal([]byte(raw), &comment); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if comment.Issue == nil {
		t.Fatal("Issue should not be nil")
	}
	if comment.Issue.Identifier != "ENG-5" {
		t.Errorf("Issue.Identifier: got %q, want %q", comment.Issue.Identifier, "ENG-5")
	}
}
