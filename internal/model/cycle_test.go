package model

import (
	"encoding/json"
	"testing"
)

func TestCycleDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "cycle-1",
		"name": "Sprint 42",
		"number": 42,
		"description": "Q1 sprint",
		"startsAt": "2026-01-06T00:00:00.000Z",
		"endsAt": "2026-01-20T00:00:00.000Z",
		"isActive": true,
		"isFuture": false,
		"isPast": false,
		"progress": 0.6,
		"team": {"id": "t1", "name": "Engineering", "displayName": "Engineering", "key": "ENG", "cyclesEnabled": true, "issueEstimationType": "points", "createdAt": "2025-01-01T00:00:00.000Z", "updatedAt": "2025-01-01T00:00:00.000Z"},
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-10T00:00:00.000Z"
	}`

	var cycle Cycle
	if err := json.Unmarshal([]byte(raw), &cycle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cycle.ID != "cycle-1" {
		t.Errorf("ID: got %q, want %q", cycle.ID, "cycle-1")
	}
	if cycle.Name == nil || *cycle.Name != "Sprint 42" {
		t.Errorf("Name: unexpected value")
	}
	if cycle.Number != 42 {
		t.Errorf("Number: got %v, want 42", cycle.Number)
	}
	if cycle.Description == nil || *cycle.Description != "Q1 sprint" {
		t.Errorf("Description: unexpected value")
	}
	if cycle.StartsAt != "2026-01-06T00:00:00.000Z" {
		t.Errorf("StartsAt: got %q", cycle.StartsAt)
	}
	if cycle.EndsAt != "2026-01-20T00:00:00.000Z" {
		t.Errorf("EndsAt: got %q", cycle.EndsAt)
	}
	if !cycle.IsActive {
		t.Errorf("IsActive: expected true")
	}
	if cycle.IsFuture {
		t.Errorf("IsFuture: expected false")
	}
	if cycle.IsPast {
		t.Errorf("IsPast: expected false")
	}
	if cycle.Progress != 0.6 {
		t.Errorf("Progress: got %v, want 0.6", cycle.Progress)
	}
	if cycle.Team.Key != "ENG" {
		t.Errorf("Team.Key: got %q", cycle.Team.Key)
	}
}

func TestCycleNullableFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "cycle-2",
		"number": 1,
		"startsAt": "2026-02-01T00:00:00.000Z",
		"endsAt": "2026-02-15T00:00:00.000Z",
		"isActive": false,
		"isFuture": true,
		"isPast": false,
		"progress": 0,
		"team": {"id": "t2", "name": "Platform", "displayName": "Platform", "key": "PLT", "cyclesEnabled": true, "issueEstimationType": "points", "createdAt": "2025-01-01T00:00:00.000Z", "updatedAt": "2025-01-01T00:00:00.000Z"},
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z"
	}`

	var cycle Cycle
	if err := json.Unmarshal([]byte(raw), &cycle); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if cycle.Name != nil {
		t.Errorf("Name should be nil")
	}
	if cycle.Description != nil {
		t.Errorf("Description should be nil")
	}
	if cycle.IsActive {
		t.Errorf("IsActive: expected false")
	}
	if !cycle.IsFuture {
		t.Errorf("IsFuture: expected true")
	}
}

func TestCycleConnectionDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"nodes": [
			{
				"id": "c1",
				"number": 1,
				"startsAt": "2026-01-01T00:00:00.000Z",
				"endsAt": "2026-01-15T00:00:00.000Z",
				"isActive": false,
				"isFuture": false,
				"isPast": true,
				"progress": 1,
				"team": {"id": "t1", "name": "Engineering", "displayName": "Engineering", "key": "ENG", "cyclesEnabled": true, "issueEstimationType": "points", "createdAt": "2025-01-01T00:00:00.000Z", "updatedAt": "2025-01-01T00:00:00.000Z"},
				"createdAt": "2026-01-01T00:00:00.000Z",
				"updatedAt": "2026-01-15T00:00:00.000Z"
			},
			{
				"id": "c2",
				"number": 2,
				"startsAt": "2026-01-15T00:00:00.000Z",
				"endsAt": "2026-01-29T00:00:00.000Z",
				"isActive": true,
				"isFuture": false,
				"isPast": false,
				"progress": 0.3,
				"team": {"id": "t1", "name": "Engineering", "displayName": "Engineering", "key": "ENG", "cyclesEnabled": true, "issueEstimationType": "points", "createdAt": "2025-01-01T00:00:00.000Z", "updatedAt": "2025-01-01T00:00:00.000Z"},
				"createdAt": "2026-01-01T00:00:00.000Z",
				"updatedAt": "2026-01-20T00:00:00.000Z"
			}
		]
	}`

	var conn CycleConnection
	if err := json.Unmarshal([]byte(raw), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(conn.Nodes) != 2 {
		t.Fatalf("Nodes: got %d, want 2", len(conn.Nodes))
	}
	if conn.Nodes[0].Number != 1 {
		t.Errorf("Nodes[0].Number: got %v", conn.Nodes[0].Number)
	}
	if !conn.Nodes[0].IsPast {
		t.Errorf("Nodes[0].IsPast: expected true")
	}
	if !conn.Nodes[1].IsActive {
		t.Errorf("Nodes[1].IsActive: expected true")
	}
}
