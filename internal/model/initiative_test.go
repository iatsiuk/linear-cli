package model

import (
	"encoding/json"
	"testing"
)

func TestInitiativeDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "ini-1",
		"name": "Q1 Goals",
		"description": "First quarter goals",
		"status": "Active"
	}`

	var ini Initiative
	if err := json.Unmarshal([]byte(raw), &ini); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ini.ID != "ini-1" {
		t.Errorf("ID: got %q, want ini-1", ini.ID)
	}
	if ini.Name != "Q1 Goals" {
		t.Errorf("Name: got %q, want Q1 Goals", ini.Name)
	}
	if ini.Description == nil || *ini.Description != "First quarter goals" {
		t.Errorf("Description: got %v, want 'First quarter goals'", ini.Description)
	}
	if ini.Status != "Active" {
		t.Errorf("Status: got %q, want Active", ini.Status)
	}
}

func TestInitiativeDeserialization_NoDescription(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "ini-2",
		"name": "Platform Upgrade",
		"status": "Planned"
	}`

	var ini Initiative
	if err := json.Unmarshal([]byte(raw), &ini); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if ini.Description != nil {
		t.Errorf("Description: got %v, want nil", ini.Description)
	}
	if ini.Status != "Planned" {
		t.Errorf("Status: got %q, want Planned", ini.Status)
	}
}

func TestInitiativeConnectionDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"nodes": [
			{"id": "ini-1", "name": "Q1 Goals", "status": "Active"},
			{"id": "ini-2", "name": "Platform Upgrade", "status": "Planned", "description": "Upgrade platform"}
		]
	}`

	var conn InitiativeConnection
	if err := json.Unmarshal([]byte(raw), &conn); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(conn.Nodes) != 2 {
		t.Fatalf("Nodes: got %d, want 2", len(conn.Nodes))
	}
	if conn.Nodes[0].Name != "Q1 Goals" {
		t.Errorf("Nodes[0].Name: got %q, want Q1 Goals", conn.Nodes[0].Name)
	}
	if conn.Nodes[1].Description == nil || *conn.Nodes[1].Description != "Upgrade platform" {
		t.Errorf("Nodes[1].Description: got %v, want 'Upgrade platform'", conn.Nodes[1].Description)
	}
}
