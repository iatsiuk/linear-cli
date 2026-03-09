package model

import (
	"encoding/json"
	"testing"
)

func TestIssueLabelDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "lbl-1",
		"name": "bug",
		"color": "#ff0000",
		"description": "Something is broken",
		"isGroup": false,
		"createdAt": "2026-01-01T00:00:00.000Z",
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"}
	}`

	var label IssueLabel
	if err := json.Unmarshal([]byte(raw), &label); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if label.ID != "lbl-1" {
		t.Errorf("ID: got %q, want %q", label.ID, "lbl-1")
	}
	if label.Name != "bug" {
		t.Errorf("Name: got %q, want %q", label.Name, "bug")
	}
	if label.Color != "#ff0000" {
		t.Errorf("Color: got %q, want %q", label.Color, "#ff0000")
	}
	if label.Description == nil || *label.Description != "Something is broken" {
		t.Errorf("Description: unexpected value %v", label.Description)
	}
	if label.IsGroup {
		t.Errorf("IsGroup: got true, want false")
	}
	if label.CreatedAt != "2026-01-01T00:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", label.CreatedAt)
	}
	if label.Team == nil {
		t.Fatal("Team should not be nil")
	}
	if label.Team.ID != "t1" {
		t.Errorf("Team.ID: got %q, want %q", label.Team.ID, "t1")
	}
	if label.Team.Key != "ENG" {
		t.Errorf("Team.Key: got %q, want %q", label.Team.Key, "ENG")
	}
	if label.Parent != nil {
		t.Errorf("Parent: expected nil, got %v", label.Parent)
	}
}

func TestIssueLabelWithParent(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "lbl-2",
		"name": "backend/auth",
		"color": "#0000ff",
		"isGroup": false,
		"createdAt": "2026-01-02T00:00:00.000Z",
		"parent": {
			"id": "lbl-0",
			"name": "backend",
			"color": "#000088",
			"isGroup": true,
			"createdAt": "2026-01-01T00:00:00.000Z"
		}
	}`

	var label IssueLabel
	if err := json.Unmarshal([]byte(raw), &label); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if label.Parent == nil {
		t.Fatal("Parent should not be nil")
	}
	if label.Parent.ID != "lbl-0" {
		t.Errorf("Parent.ID: got %q, want %q", label.Parent.ID, "lbl-0")
	}
	if label.Parent.Name != "backend" {
		t.Errorf("Parent.Name: got %q, want %q", label.Parent.Name, "backend")
	}
	if !label.Parent.IsGroup {
		t.Errorf("Parent.IsGroup: got false, want true")
	}
}

func TestIssueLabelWorkspaceLabelNilTeam(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "lbl-ws",
		"name": "workspace-label",
		"color": "#aabbcc",
		"isGroup": false,
		"createdAt": "2026-01-01T00:00:00.000Z"
	}`

	var label IssueLabel
	if err := json.Unmarshal([]byte(raw), &label); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if label.Team != nil {
		t.Errorf("Team: expected nil for workspace label, got %v", label.Team)
	}
	if label.Description != nil {
		t.Errorf("Description: expected nil, got %v", label.Description)
	}
}
