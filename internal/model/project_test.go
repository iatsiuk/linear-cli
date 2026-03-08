package model

import (
	"encoding/json"
	"testing"
)

func TestProjectDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "proj-1",
		"name": "New Website",
		"description": "Redesign the website",
		"color": "#ff5733",
		"icon": "rocket",
		"health": "onTrack",
		"status": {"id": "ps-1", "name": "In Progress", "type": "started"},
		"progress": 0.45,
		"startDate": "2026-01-01",
		"targetDate": "2026-06-01",
		"creator": {"id": "u1", "displayName": "Alice", "email": "alice@example.com", "active": true, "admin": false, "guest": false, "isMe": false, "createdAt": "2025-01-01T00:00:00.000Z", "updatedAt": "2025-01-01T00:00:00.000Z"},
		"teams": {"nodes": [{"id": "t1", "name": "Engineering", "displayName": "Engineering", "key": "ENG", "cyclesEnabled": true, "issueEstimationType": "points", "createdAt": "2025-01-01T00:00:00.000Z", "updatedAt": "2025-01-01T00:00:00.000Z"}]},
		"url": "https://linear.app/project/new-website",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-02-01T00:00:00.000Z"
	}`

	var project Project
	if err := json.Unmarshal([]byte(raw), &project); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if project.ID != "proj-1" {
		t.Errorf("ID: got %q, want %q", project.ID, "proj-1")
	}
	if project.Name != "New Website" {
		t.Errorf("Name: got %q", project.Name)
	}
	if project.Description != "Redesign the website" {
		t.Errorf("Description: got %q", project.Description)
	}
	if project.Color != "#ff5733" {
		t.Errorf("Color: got %q", project.Color)
	}
	if project.Icon == nil || *project.Icon != "rocket" {
		t.Errorf("Icon: unexpected value")
	}
	if project.Health == nil || *project.Health != "onTrack" {
		t.Errorf("Health: unexpected value")
	}
	if project.Status.ID != "ps-1" {
		t.Errorf("Status.ID: got %q", project.Status.ID)
	}
	if project.Status.Name != "In Progress" {
		t.Errorf("Status.Name: got %q", project.Status.Name)
	}
	if project.Status.Type != "started" {
		t.Errorf("Status.Type: got %q", project.Status.Type)
	}
	if project.Progress != 0.45 {
		t.Errorf("Progress: got %v, want 0.45", project.Progress)
	}
	if project.StartDate == nil || *project.StartDate != "2026-01-01" {
		t.Errorf("StartDate: unexpected value")
	}
	if project.TargetDate == nil || *project.TargetDate != "2026-06-01" {
		t.Errorf("TargetDate: unexpected value")
	}
	if project.Creator == nil || project.Creator.DisplayName != "Alice" {
		t.Errorf("Creator: unexpected value")
	}
	if len(project.Teams.Nodes) != 1 || project.Teams.Nodes[0].Key != "ENG" {
		t.Errorf("Teams: unexpected value")
	}
	if project.URL != "https://linear.app/project/new-website" {
		t.Errorf("URL: got %q", project.URL)
	}
}

func TestProjectNullableFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "proj-2",
		"name": "Minimal Project",
		"description": "",
		"color": "#000",
		"status": {"id": "ps-2", "name": "Backlog", "type": "backlog"},
		"progress": 0,
		"teams": {"nodes": []},
		"url": "https://linear.app/project/minimal",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z"
	}`

	var project Project
	if err := json.Unmarshal([]byte(raw), &project); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if project.Icon != nil {
		t.Errorf("Icon should be nil")
	}
	if project.Health != nil {
		t.Errorf("Health should be nil")
	}
	if project.StartDate != nil {
		t.Errorf("StartDate should be nil")
	}
	if project.TargetDate != nil {
		t.Errorf("TargetDate should be nil")
	}
	if project.Creator != nil {
		t.Errorf("Creator should be nil")
	}
	if len(project.Teams.Nodes) != 0 {
		t.Errorf("Teams should be empty")
	}
}

func TestProjectConnectionDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"nodes": [
			{
				"id": "p1",
				"name": "Alpha",
				"description": "",
				"color": "#f00",
				"status": {"id": "ps-1", "name": "Started", "type": "started"},
				"progress": 0.1,
				"teams": {"nodes": []},
				"url": "https://linear.app/p1",
				"createdAt": "2026-01-01T00:00:00.000Z",
				"updatedAt": "2026-01-01T00:00:00.000Z"
			},
			{
				"id": "p2",
				"name": "Beta",
				"description": "",
				"color": "#0f0",
				"status": {"id": "ps-2", "name": "Planned", "type": "planned"},
				"progress": 0,
				"teams": {"nodes": []},
				"url": "https://linear.app/p2",
				"createdAt": "2026-01-01T00:00:00.000Z",
				"updatedAt": "2026-01-01T00:00:00.000Z"
			}
		]
	}`

	var conn ProjectConnection
	if err := json.Unmarshal([]byte(raw), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(conn.Nodes) != 2 {
		t.Fatalf("Nodes: got %d, want 2", len(conn.Nodes))
	}
	if conn.Nodes[0].Name != "Alpha" {
		t.Errorf("Nodes[0].Name: got %q", conn.Nodes[0].Name)
	}
	if conn.Nodes[1].Status.Type != "planned" {
		t.Errorf("Nodes[1].Status.Type: got %q", conn.Nodes[1].Status.Type)
	}
}
