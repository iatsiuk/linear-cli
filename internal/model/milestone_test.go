package model

import (
	"encoding/json"
	"testing"
)

func TestProjectMilestoneDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "ms-1",
		"name": "v1.0 Release",
		"description": "First stable release",
		"targetDate": "2026-06-01",
		"sortOrder": 1.0,
		"status": "unstarted"
	}`

	var ms ProjectMilestone
	if err := json.Unmarshal([]byte(raw), &ms); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ms.ID != "ms-1" {
		t.Errorf("ID: got %q, want %q", ms.ID, "ms-1")
	}
	if ms.Name != "v1.0 Release" {
		t.Errorf("Name: got %q, want %q", ms.Name, "v1.0 Release")
	}
	if ms.Description == nil || *ms.Description != "First stable release" {
		t.Errorf("Description: got %v, want %q", ms.Description, "First stable release")
	}
	if ms.TargetDate == nil || *ms.TargetDate != "2026-06-01" {
		t.Errorf("TargetDate: got %v, want %q", ms.TargetDate, "2026-06-01")
	}
	if ms.SortOrder != 1.0 {
		t.Errorf("SortOrder: got %v, want 1.0", ms.SortOrder)
	}
	if ms.Status != "unstarted" {
		t.Errorf("Status: got %q, want %q", ms.Status, "unstarted")
	}
}

func TestProjectMilestoneDeserialization_NullOptionals(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "ms-2",
		"name": "Beta",
		"sortOrder": 0.5,
		"status": "next"
	}`

	var ms ProjectMilestone
	if err := json.Unmarshal([]byte(raw), &ms); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ms.Description != nil {
		t.Errorf("Description: expected nil, got %v", ms.Description)
	}
	if ms.TargetDate != nil {
		t.Errorf("TargetDate: expected nil, got %v", ms.TargetDate)
	}
}

func TestProjectMilestoneConnectionDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"nodes": [
			{"id": "ms-1", "name": "Alpha", "sortOrder": 1.0, "status": "done"},
			{"id": "ms-2", "name": "Beta", "sortOrder": 2.0, "status": "unstarted"}
		]
	}`

	var conn ProjectMilestoneConnection
	if err := json.Unmarshal([]byte(raw), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(conn.Nodes) != 2 {
		t.Fatalf("Nodes: got %d, want 2", len(conn.Nodes))
	}
	if conn.Nodes[0].Name != "Alpha" {
		t.Errorf("Nodes[0].Name: got %q, want Alpha", conn.Nodes[0].Name)
	}
	if conn.Nodes[1].Status != "unstarted" {
		t.Errorf("Nodes[1].Status: got %q, want unstarted", conn.Nodes[1].Status)
	}
}
