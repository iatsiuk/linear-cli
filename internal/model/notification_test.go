package model

import (
	"encoding/json"
	"testing"
)

func TestNotificationDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "notif-1",
		"type": "issueAssignedToYou",
		"readAt": "2026-01-02T10:00:00.000Z",
		"archivedAt": null,
		"createdAt": "2026-01-01T08:00:00.000Z",
		"updatedAt": "2026-01-02T10:00:00.000Z",
		"title": "ENG-123 assigned to you",
		"subtitle": "Fix the bug",
		"url": "https://linear.app/inbox/notif-1"
	}`

	var n Notification
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if n.ID != "notif-1" {
		t.Errorf("ID: got %q, want %q", n.ID, "notif-1")
	}
	if n.Type != "issueAssignedToYou" {
		t.Errorf("Type: got %q, want %q", n.Type, "issueAssignedToYou")
	}
	if n.ReadAt == nil || *n.ReadAt != "2026-01-02T10:00:00.000Z" {
		t.Errorf("ReadAt: got %v, want %q", n.ReadAt, "2026-01-02T10:00:00.000Z")
	}
	if n.ArchivedAt != nil {
		t.Errorf("ArchivedAt: got %v, want nil", n.ArchivedAt)
	}
	if n.CreatedAt != "2026-01-01T08:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", n.CreatedAt)
	}
	if n.Title != "ENG-123 assigned to you" {
		t.Errorf("Title: got %q", n.Title)
	}
	if n.URL != "https://linear.app/inbox/notif-1" {
		t.Errorf("URL: got %q", n.URL)
	}
}

func TestNotificationUnread(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "notif-2",
		"type": "issueMention",
		"readAt": null,
		"archivedAt": null,
		"createdAt": "2026-01-03T09:00:00.000Z",
		"updatedAt": "2026-01-03T09:00:00.000Z",
		"title": "You were mentioned",
		"subtitle": "In a comment",
		"url": "https://linear.app/inbox/notif-2"
	}`

	var n Notification
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if n.ReadAt != nil {
		t.Errorf("ReadAt: got %v, want nil (unread)", n.ReadAt)
	}
}

func TestNotificationConnectionDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"nodes": [
			{
				"id": "notif-1",
				"type": "issueAssignedToYou",
				"readAt": null,
				"archivedAt": null,
				"createdAt": "2026-01-01T08:00:00.000Z",
				"updatedAt": "2026-01-01T08:00:00.000Z",
				"title": "ENG-1 assigned to you",
				"subtitle": "Do the thing",
				"url": "https://linear.app/inbox/notif-1"
			},
			{
				"id": "notif-2",
				"type": "issueMention",
				"readAt": "2026-01-02T10:00:00.000Z",
				"archivedAt": null,
				"createdAt": "2026-01-01T07:00:00.000Z",
				"updatedAt": "2026-01-02T10:00:00.000Z",
				"title": "You were mentioned",
				"subtitle": "In a comment",
				"url": "https://linear.app/inbox/notif-2"
			}
		]
	}`

	var conn NotificationConnection
	if err := json.Unmarshal([]byte(raw), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(conn.Nodes) != 2 {
		t.Fatalf("Nodes: got %d, want 2", len(conn.Nodes))
	}
	if conn.Nodes[0].ID != "notif-1" {
		t.Errorf("Nodes[0].ID: got %q, want %q", conn.Nodes[0].ID, "notif-1")
	}
	if conn.Nodes[1].ReadAt == nil {
		t.Errorf("Nodes[1].ReadAt: got nil, want non-nil")
	}
}
