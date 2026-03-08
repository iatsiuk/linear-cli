package model

import (
	"encoding/json"
	"testing"
)

func TestTeamDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "t1",
		"name": "Engineering",
		"displayName": "Eng Team",
		"description": "Core engineering team",
		"icon": "lightning",
		"color": "#4ea7fc",
		"key": "ENG",
		"cyclesEnabled": true,
		"issueEstimationType": "fibonacci",
		"createdAt": "2025-01-01T00:00:00.000Z",
		"updatedAt": "2025-06-01T00:00:00.000Z"
	}`

	var team Team
	if err := json.Unmarshal([]byte(raw), &team); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if team.ID != "t1" {
		t.Errorf("ID: got %q, want %q", team.ID, "t1")
	}
	if team.Name != "Engineering" {
		t.Errorf("Name: got %q", team.Name)
	}
	if team.DisplayName != "Eng Team" {
		t.Errorf("DisplayName: got %q", team.DisplayName)
	}
	if team.Description == nil || *team.Description != "Core engineering team" {
		t.Errorf("Description: unexpected value")
	}
	if team.Icon == nil || *team.Icon != "lightning" {
		t.Errorf("Icon: unexpected value")
	}
	if team.Color == nil || *team.Color != "#4ea7fc" {
		t.Errorf("Color: unexpected value")
	}
	if team.Key != "ENG" {
		t.Errorf("Key: got %q", team.Key)
	}
	if !team.CyclesEnabled {
		t.Errorf("CyclesEnabled: got false, want true")
	}
	if team.IssueEstimationType != "fibonacci" {
		t.Errorf("IssueEstimationType: got %q", team.IssueEstimationType)
	}
	if team.CreatedAt != "2025-01-01T00:00:00.000Z" {
		t.Errorf("CreatedAt: got %q", team.CreatedAt)
	}
	if team.UpdatedAt != "2025-06-01T00:00:00.000Z" {
		t.Errorf("UpdatedAt: got %q", team.UpdatedAt)
	}
}

func TestTeamNullableFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "t2",
		"name": "Platform",
		"displayName": "Platform",
		"key": "PLT",
		"cyclesEnabled": false,
		"issueEstimationType": "notUsed",
		"createdAt": "2025-01-01T00:00:00.000Z",
		"updatedAt": "2025-01-01T00:00:00.000Z"
	}`

	var team Team
	if err := json.Unmarshal([]byte(raw), &team); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if team.Description != nil {
		t.Errorf("Description should be nil, got %v", team.Description)
	}
	if team.Icon != nil {
		t.Errorf("Icon should be nil, got %v", team.Icon)
	}
	if team.Color != nil {
		t.Errorf("Color should be nil, got %v", team.Color)
	}
	if team.CyclesEnabled {
		t.Errorf("CyclesEnabled: got true, want false")
	}
}
