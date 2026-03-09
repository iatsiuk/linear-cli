package model

import (
	"encoding/json"
	"testing"
)

func TestWorkflowStateDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "st-1",
		"name": "In Progress",
		"color": "#ff9900",
		"type": "started",
		"description": "Work has begun",
		"position": 2.5,
		"createdAt": "2026-01-01T00:00:00.000Z",
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"}
	}`

	var state WorkflowState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if state.ID != "st-1" {
		t.Errorf("ID: got %q, want %q", state.ID, "st-1")
	}
	if state.Name != "In Progress" {
		t.Errorf("Name: got %q, want %q", state.Name, "In Progress")
	}
	if state.Color != "#ff9900" {
		t.Errorf("Color: got %q, want %q", state.Color, "#ff9900")
	}
	if state.Type != "started" {
		t.Errorf("Type: got %q, want %q", state.Type, "started")
	}
	if state.Description == nil || *state.Description != "Work has begun" {
		t.Errorf("Description: unexpected value %v", state.Description)
	}
	if state.Position == nil || *state.Position != 2.5 {
		t.Errorf("Position: got %v, want 2.5", state.Position)
	}
	if state.CreatedAt != "2026-01-01T00:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", state.CreatedAt)
	}
	if state.Team == nil {
		t.Fatal("Team should not be nil")
	}
	if state.Team.ID != "t1" {
		t.Errorf("Team.ID: got %q, want %q", state.Team.ID, "t1")
	}
}

func TestWorkflowStateMinimalFields(t *testing.T) {
	t.Parallel()

	// minimal state as embedded in issue queries (only id, name, color, type)
	raw := `{"id": "st-2", "name": "Done", "color": "#00cc00", "type": "completed"}`

	var state WorkflowState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if state.ID != "st-2" {
		t.Errorf("ID: got %q", state.ID)
	}
	if state.Type != "completed" {
		t.Errorf("Type: got %q", state.Type)
	}
	if state.Description != nil {
		t.Errorf("Description: expected nil, got %v", state.Description)
	}
	if state.Team != nil {
		t.Errorf("Team: expected nil, got %v", state.Team)
	}
}
