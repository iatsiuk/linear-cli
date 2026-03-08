package model

import (
	"encoding/json"
	"testing"
)

func TestIssueDeserialization(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "abc-1",
		"identifier": "ENG-42",
		"title": "Fix login bug",
		"description": "Details here",
		"priority": 2,
		"priorityLabel": "Medium",
		"estimate": 3,
		"dueDate": "2026-04-01",
		"url": "https://linear.app/issue/ENG-42",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-02T00:00:00.000Z",
		"state": {"id": "s1", "name": "In Progress", "color": "#ff0", "type": "started"},
		"assignee": {"id": "u1", "displayName": "Alice", "email": "alice@example.com"},
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
		"labels": {"nodes": [{"id": "l1", "name": "bug", "color": "#f00"}]}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if issue.ID != "abc-1" {
		t.Errorf("ID: got %q, want %q", issue.ID, "abc-1")
	}
	if issue.Identifier != "ENG-42" {
		t.Errorf("Identifier: got %q, want %q", issue.Identifier, "ENG-42")
	}
	if issue.Title != "Fix login bug" {
		t.Errorf("Title: got %q", issue.Title)
	}
	if issue.Description == nil || *issue.Description != "Details here" {
		t.Errorf("Description: unexpected value")
	}
	if issue.Priority != 2 {
		t.Errorf("Priority: got %v, want 2", issue.Priority)
	}
	if issue.PriorityLabel != "Medium" {
		t.Errorf("PriorityLabel: got %q", issue.PriorityLabel)
	}
	if issue.Estimate == nil || *issue.Estimate != 3 {
		t.Errorf("Estimate: unexpected value")
	}
	if issue.DueDate == nil || *issue.DueDate != "2026-04-01" {
		t.Errorf("DueDate: unexpected value")
	}
	if issue.URL != "https://linear.app/issue/ENG-42" {
		t.Errorf("URL: got %q", issue.URL)
	}
}

func TestIssueNullableFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "abc-2",
		"identifier": "ENG-1",
		"title": "No description",
		"priority": 0,
		"priorityLabel": "No priority",
		"url": "https://linear.app/issue/ENG-1",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"state": {"id": "s2", "name": "Backlog", "color": "#ccc", "type": "backlog"},
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
		"labels": {"nodes": []}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if issue.Description != nil {
		t.Errorf("Description should be nil, got %v", issue.Description)
	}
	if issue.Estimate != nil {
		t.Errorf("Estimate should be nil, got %v", issue.Estimate)
	}
	if issue.DueDate != nil {
		t.Errorf("DueDate should be nil, got %v", issue.DueDate)
	}
	if issue.Assignee != nil {
		t.Errorf("Assignee should be nil, got %v", issue.Assignee)
	}
	if len(issue.Labels.Nodes) != 0 {
		t.Errorf("Labels should be empty")
	}
}

func TestIssueNestedStructs(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "abc-3",
		"identifier": "ENG-99",
		"title": "Test nested",
		"priority": 1,
		"priorityLabel": "Urgent",
		"url": "https://linear.app/issue/ENG-99",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"state": {"id": "s3", "name": "Done", "color": "#0f0", "type": "completed"},
		"assignee": {"id": "u2", "displayName": "Bob", "email": "bob@example.com"},
		"team": {"id": "t2", "name": "Platform", "key": "PLT"},
		"labels": {"nodes": [
			{"id": "l1", "name": "bug", "color": "#f00"},
			{"id": "l2", "name": "urgent", "color": "#f80"}
		]}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// state
	if issue.State.ID != "s3" {
		t.Errorf("State.ID: got %q", issue.State.ID)
	}
	if issue.State.Name != "Done" {
		t.Errorf("State.Name: got %q", issue.State.Name)
	}
	if issue.State.Color != "#0f0" {
		t.Errorf("State.Color: got %q", issue.State.Color)
	}
	if issue.State.Type != "completed" {
		t.Errorf("State.Type: got %q", issue.State.Type)
	}

	// assignee
	if issue.Assignee == nil {
		t.Fatal("Assignee should not be nil")
	}
	if issue.Assignee.ID != "u2" {
		t.Errorf("Assignee.ID: got %q", issue.Assignee.ID)
	}
	if issue.Assignee.DisplayName != "Bob" {
		t.Errorf("Assignee.DisplayName: got %q", issue.Assignee.DisplayName)
	}
	if issue.Assignee.Email != "bob@example.com" {
		t.Errorf("Assignee.Email: got %q", issue.Assignee.Email)
	}

	// team
	if issue.Team.ID != "t2" {
		t.Errorf("Team.ID: got %q", issue.Team.ID)
	}
	if issue.Team.Name != "Platform" {
		t.Errorf("Team.Name: got %q", issue.Team.Name)
	}
	if issue.Team.Key != "PLT" {
		t.Errorf("Team.Key: got %q", issue.Team.Key)
	}

	// labels
	if len(issue.Labels.Nodes) != 2 {
		t.Fatalf("Labels: got %d, want 2", len(issue.Labels.Nodes))
	}
	if issue.Labels.Nodes[0].Name != "bug" {
		t.Errorf("Labels[0].Name: got %q", issue.Labels.Nodes[0].Name)
	}
	if issue.Labels.Nodes[1].Name != "urgent" {
		t.Errorf("Labels[1].Name: got %q", issue.Labels.Nodes[1].Name)
	}
}
