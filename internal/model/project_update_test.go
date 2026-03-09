package model

import (
	"encoding/json"
	"testing"
)

func TestProjectUpdateDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "pu-1",
		"body": "All systems go",
		"health": "onTrack",
		"user": {
			"id": "user-1",
			"displayName": "Alice",
			"email": "alice@example.com"
		},
		"project": {
			"id": "proj-1",
			"name": "My Project",
			"description": "",
			"color": "#FF0000",
			"progress": 0.5,
			"url": "https://linear.app/project/proj-1",
			"createdAt": "2026-01-01T00:00:00Z",
			"updatedAt": "2026-01-02T00:00:00Z",
			"status": {"id": "s1", "name": "started", "type": "started"},
			"teams": {"nodes": []}
		},
		"createdAt": "2026-03-01T00:00:00Z",
		"updatedAt": "2026-03-02T00:00:00Z"
	}`

	var pu ProjectUpdate
	if err := json.Unmarshal([]byte(raw), &pu); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if pu.ID != "pu-1" {
		t.Errorf("ID: got %q, want %q", pu.ID, "pu-1")
	}
	if pu.Body != "All systems go" {
		t.Errorf("Body: got %q, want %q", pu.Body, "All systems go")
	}
	if pu.Health != "onTrack" {
		t.Errorf("Health: got %q, want %q", pu.Health, "onTrack")
	}
	if pu.User.DisplayName != "Alice" {
		t.Errorf("User.DisplayName: got %q, want %q", pu.User.DisplayName, "Alice")
	}
	if pu.Project.ID != "proj-1" {
		t.Errorf("Project.ID: got %q, want %q", pu.Project.ID, "proj-1")
	}
	if pu.CreatedAt != "2026-03-01T00:00:00Z" {
		t.Errorf("CreatedAt: got %q", pu.CreatedAt)
	}
}

func TestProjectUpdateConnectionDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"nodes": [
			{
				"id": "pu-1",
				"body": "Progress update",
				"health": "atRisk",
				"user": {"id": "u1", "displayName": "Bob", "email": "bob@example.com"},
				"project": {
					"id": "p1", "name": "P1", "description": "", "color": "#000",
					"progress": 0.3, "url": "https://linear.app/p1",
					"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
					"status": {"id": "s1", "name": "started", "type": "started"},
					"teams": {"nodes": []}
				},
				"createdAt": "2026-03-01T00:00:00Z",
				"updatedAt": "2026-03-01T00:00:00Z"
			},
			{
				"id": "pu-2",
				"body": "Behind schedule",
				"health": "offTrack",
				"user": {"id": "u2", "displayName": "Carol", "email": "carol@example.com"},
				"project": {
					"id": "p1", "name": "P1", "description": "", "color": "#000",
					"progress": 0.1, "url": "https://linear.app/p1",
					"createdAt": "2026-01-01T00:00:00Z", "updatedAt": "2026-01-01T00:00:00Z",
					"status": {"id": "s1", "name": "started", "type": "started"},
					"teams": {"nodes": []}
				},
				"createdAt": "2026-02-01T00:00:00Z",
				"updatedAt": "2026-02-01T00:00:00Z"
			}
		]
	}`

	var conn ProjectUpdateConnection
	if err := json.Unmarshal([]byte(raw), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(conn.Nodes) != 2 {
		t.Fatalf("Nodes: got %d, want 2", len(conn.Nodes))
	}
	if conn.Nodes[0].Health != "atRisk" {
		t.Errorf("Nodes[0].Health: got %q, want atRisk", conn.Nodes[0].Health)
	}
	if conn.Nodes[1].Health != "offTrack" {
		t.Errorf("Nodes[1].Health: got %q, want offTrack", conn.Nodes[1].Health)
	}
}
