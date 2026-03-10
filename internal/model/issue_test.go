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

func TestIssueUnmarshal_WithParent(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "abc-5",
		"identifier": "ENG-5",
		"title": "Child issue",
		"priority": 0,
		"priorityLabel": "No priority",
		"url": "https://linear.app/issue/ENG-5",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"state": {"id": "s1", "name": "Backlog", "color": "#ccc", "type": "backlog"},
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
		"labels": {"nodes": []},
		"parent": {"id": "p1", "identifier": "ENG-1", "title": "Parent issue"}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if issue.Parent == nil {
		t.Fatal("Parent should not be nil")
	}
	if issue.Parent.ID != "p1" {
		t.Errorf("Parent.ID: got %q, want %q", issue.Parent.ID, "p1")
	}
	if issue.Parent.Identifier != "ENG-1" {
		t.Errorf("Parent.Identifier: got %q, want %q", issue.Parent.Identifier, "ENG-1")
	}
	if issue.Parent.Title != "Parent issue" {
		t.Errorf("Parent.Title: got %q, want %q", issue.Parent.Title, "Parent issue")
	}
}

func TestIssueUnmarshal_WithProject(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "abc-6",
		"identifier": "ENG-6",
		"title": "Issue with project",
		"priority": 0,
		"priorityLabel": "No priority",
		"url": "https://linear.app/issue/ENG-6",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"state": {"id": "s1", "name": "Backlog", "color": "#ccc", "type": "backlog"},
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
		"labels": {"nodes": []},
		"project": {"id": "proj1", "name": "My Project"}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if issue.Project == nil {
		t.Fatal("Project should not be nil")
	}
	if issue.Project.ID != "proj1" {
		t.Errorf("Project.ID: got %q, want %q", issue.Project.ID, "proj1")
	}
	if issue.Project.Name != "My Project" {
		t.Errorf("Project.Name: got %q, want %q", issue.Project.Name, "My Project")
	}
}

func TestIssueUnmarshal_NilParentProject(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "abc-7",
		"identifier": "ENG-7",
		"title": "No parent no project",
		"priority": 0,
		"priorityLabel": "No priority",
		"url": "https://linear.app/issue/ENG-7",
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"state": {"id": "s1", "name": "Backlog", "color": "#ccc", "type": "backlog"},
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
		"labels": {"nodes": []}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if issue.Parent != nil {
		t.Errorf("Parent should be nil, got %+v", issue.Parent)
	}
	if issue.Project != nil {
		t.Errorf("Project should be nil, got %+v", issue.Project)
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

func TestIssueDeserialization_NewFields(t *testing.T) {
	t.Parallel()

	trashed := true
	raw := `{
		"id": "new-1",
		"identifier": "ENG-100",
		"number": 100,
		"title": "New fields issue",
		"priority": 1,
		"priorityLabel": "Urgent",
		"branchName": "eng-100-new-fields-issue",
		"url": "https://linear.app/issue/ENG-100",
		"trashed": true,
		"customerTicketCount": 5,
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-02T00:00:00.000Z",
		"archivedAt": "2026-02-01T00:00:00.000Z",
		"autoArchivedAt": "2026-02-02T00:00:00.000Z",
		"autoClosedAt": "2026-02-03T00:00:00.000Z",
		"canceledAt": "2026-02-04T00:00:00.000Z",
		"completedAt": "2026-02-05T00:00:00.000Z",
		"startedAt": "2026-02-06T00:00:00.000Z",
		"startedTriageAt": "2026-02-07T00:00:00.000Z",
		"triagedAt": "2026-02-08T00:00:00.000Z",
		"snoozedUntilAt": "2026-02-09T00:00:00.000Z",
		"addedToCycleAt": "2026-02-10T00:00:00.000Z",
		"addedToProjectAt": "2026-02-11T00:00:00.000Z",
		"addedToTeamAt": "2026-02-12T00:00:00.000Z",
		"slaBreachesAt": "2026-03-01T00:00:00.000Z",
		"slaHighRiskAt": "2026-03-02T00:00:00.000Z",
		"slaMediumRiskAt": "2026-03-03T00:00:00.000Z",
		"slaStartedAt": "2026-03-04T00:00:00.000Z",
		"slaType": "standard",
		"state": {"id": "s1", "name": "In Progress", "color": "#ff0", "type": "started"},
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
		"labels": {"nodes": []},
		"creator": {"id": "u2", "displayName": "Bob", "email": "bob@example.com"},
		"cycle": {"id": "c1", "name": "Sprint 1", "number": 1}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if issue.Number != 100 {
		t.Errorf("Number: got %v, want 100", issue.Number)
	}
	if issue.BranchName != "eng-100-new-fields-issue" {
		t.Errorf("BranchName: got %q", issue.BranchName)
	}
	if issue.Trashed == nil || *issue.Trashed != trashed {
		t.Errorf("Trashed: got %v, want true", issue.Trashed)
	}
	if issue.CustomerTicketCount != 5 {
		t.Errorf("CustomerTicketCount: got %d, want 5", issue.CustomerTicketCount)
	}

	// timestamps
	checkStr := func(field string, got *string, want string) {
		t.Helper()
		if got == nil {
			t.Errorf("%s: got nil, want %q", field, want)
		} else if *got != want {
			t.Errorf("%s: got %q, want %q", field, *got, want)
		}
	}
	checkStr("ArchivedAt", issue.ArchivedAt, "2026-02-01T00:00:00.000Z")
	checkStr("AutoArchivedAt", issue.AutoArchivedAt, "2026-02-02T00:00:00.000Z")
	checkStr("AutoClosedAt", issue.AutoClosedAt, "2026-02-03T00:00:00.000Z")
	checkStr("CanceledAt", issue.CanceledAt, "2026-02-04T00:00:00.000Z")
	checkStr("CompletedAt", issue.CompletedAt, "2026-02-05T00:00:00.000Z")
	checkStr("StartedAt", issue.StartedAt, "2026-02-06T00:00:00.000Z")
	checkStr("StartedTriageAt", issue.StartedTriageAt, "2026-02-07T00:00:00.000Z")
	checkStr("TriagedAt", issue.TriagedAt, "2026-02-08T00:00:00.000Z")
	checkStr("SnoozedUntilAt", issue.SnoozedUntilAt, "2026-02-09T00:00:00.000Z")
	checkStr("AddedToCycleAt", issue.AddedToCycleAt, "2026-02-10T00:00:00.000Z")
	checkStr("AddedToProjectAt", issue.AddedToProjectAt, "2026-02-11T00:00:00.000Z")
	checkStr("AddedToTeamAt", issue.AddedToTeamAt, "2026-02-12T00:00:00.000Z")
	checkStr("SlaBreachesAt", issue.SlaBreachesAt, "2026-03-01T00:00:00.000Z")
	checkStr("SlaHighRiskAt", issue.SlaHighRiskAt, "2026-03-02T00:00:00.000Z")
	checkStr("SlaMediumRiskAt", issue.SlaMediumRiskAt, "2026-03-03T00:00:00.000Z")
	checkStr("SlaStartedAt", issue.SlaStartedAt, "2026-03-04T00:00:00.000Z")
	checkStr("SlaType", issue.SlaType, "standard")

	// creator
	if issue.Creator == nil {
		t.Fatal("Creator should not be nil")
	}
	if issue.Creator.ID != "u2" {
		t.Errorf("Creator.ID: got %q", issue.Creator.ID)
	}
	if issue.Creator.DisplayName != "Bob" {
		t.Errorf("Creator.DisplayName: got %q", issue.Creator.DisplayName)
	}
	if issue.Creator.Email != "bob@example.com" {
		t.Errorf("Creator.Email: got %q", issue.Creator.Email)
	}

	// cycle
	if issue.Cycle == nil {
		t.Fatal("Cycle should not be nil")
	}
	if issue.Cycle.ID != "c1" {
		t.Errorf("Cycle.ID: got %q", issue.Cycle.ID)
	}
	if issue.Cycle.Name == nil || *issue.Cycle.Name != "Sprint 1" {
		t.Errorf("Cycle.Name: unexpected value")
	}
	if issue.Cycle.Number != 1 {
		t.Errorf("Cycle.Number: got %v, want 1", issue.Cycle.Number)
	}
}

func TestIssueNullableFields_NewFields(t *testing.T) {
	t.Parallel()

	raw := `{
		"id": "null-1",
		"identifier": "ENG-200",
		"number": 0,
		"title": "Minimal issue",
		"priority": 0,
		"priorityLabel": "No priority",
		"branchName": "",
		"url": "https://linear.app/issue/ENG-200",
		"customerTicketCount": 0,
		"createdAt": "2026-01-01T00:00:00.000Z",
		"updatedAt": "2026-01-01T00:00:00.000Z",
		"state": {"id": "s1", "name": "Backlog", "color": "#ccc", "type": "backlog"},
		"team": {"id": "t1", "name": "Engineering", "key": "ENG"},
		"labels": {"nodes": []}
	}`

	var issue Issue
	if err := json.Unmarshal([]byte(raw), &issue); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if issue.Trashed != nil {
		t.Errorf("Trashed should be nil, got %v", issue.Trashed)
	}
	if issue.Creator != nil {
		t.Errorf("Creator should be nil, got %+v", issue.Creator)
	}
	if issue.Cycle != nil {
		t.Errorf("Cycle should be nil, got %+v", issue.Cycle)
	}
	if issue.ArchivedAt != nil {
		t.Errorf("ArchivedAt should be nil")
	}
	if issue.AutoArchivedAt != nil {
		t.Errorf("AutoArchivedAt should be nil")
	}
	if issue.AutoClosedAt != nil {
		t.Errorf("AutoClosedAt should be nil")
	}
	if issue.CanceledAt != nil {
		t.Errorf("CanceledAt should be nil")
	}
	if issue.CompletedAt != nil {
		t.Errorf("CompletedAt should be nil")
	}
	if issue.StartedAt != nil {
		t.Errorf("StartedAt should be nil")
	}
	if issue.StartedTriageAt != nil {
		t.Errorf("StartedTriageAt should be nil")
	}
	if issue.TriagedAt != nil {
		t.Errorf("TriagedAt should be nil")
	}
	if issue.SnoozedUntilAt != nil {
		t.Errorf("SnoozedUntilAt should be nil")
	}
	if issue.AddedToCycleAt != nil {
		t.Errorf("AddedToCycleAt should be nil")
	}
	if issue.AddedToProjectAt != nil {
		t.Errorf("AddedToProjectAt should be nil")
	}
	if issue.AddedToTeamAt != nil {
		t.Errorf("AddedToTeamAt should be nil")
	}
	if issue.SlaBreachesAt != nil {
		t.Errorf("SlaBreachesAt should be nil")
	}
	if issue.SlaHighRiskAt != nil {
		t.Errorf("SlaHighRiskAt should be nil")
	}
	if issue.SlaMediumRiskAt != nil {
		t.Errorf("SlaMediumRiskAt should be nil")
	}
	if issue.SlaStartedAt != nil {
		t.Errorf("SlaStartedAt should be nil")
	}
	if issue.SlaType != nil {
		t.Errorf("SlaType should be nil")
	}
}

func TestCycleRefDeserialization(t *testing.T) {
	t.Parallel()

	t.Run("with name", func(t *testing.T) {
		t.Parallel()
		raw := `{"id": "c1", "name": "Sprint 42", "number": 42}`
		var c CycleRef
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if c.ID != "c1" {
			t.Errorf("ID: got %q", c.ID)
		}
		if c.Name == nil || *c.Name != "Sprint 42" {
			t.Errorf("Name: unexpected value %v", c.Name)
		}
		if c.Number != 42 {
			t.Errorf("Number: got %v, want 42", c.Number)
		}
	})

	t.Run("without name", func(t *testing.T) {
		t.Parallel()
		raw := `{"id": "c2", "number": 7}`
		var c CycleRef
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if c.Name != nil {
			t.Errorf("Name should be nil, got %v", c.Name)
		}
		if c.Number != 7 {
			t.Errorf("Number: got %v, want 7", c.Number)
		}
	})

	t.Run("number as float64", func(t *testing.T) {
		t.Parallel()
		raw := `{"id": "c3", "number": 3.0}`
		var c CycleRef
		if err := json.Unmarshal([]byte(raw), &c); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if c.Number != 3.0 {
			t.Errorf("Number: got %v, want 3.0", c.Number)
		}
	})
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
