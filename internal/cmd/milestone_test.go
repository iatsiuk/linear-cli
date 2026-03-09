package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func makeMilestone(id, name, status string, targetDate *string, description *string) map[string]any {
	m := map[string]any{
		"id":        id,
		"name":      name,
		"status":    status,
		"sortOrder": 1.0,
	}
	if targetDate != nil {
		m["targetDate"] = *targetDate
	}
	if description != nil {
		m["description"] = *description
	}
	return m
}

func milestoneListResponse(milestones []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"project": map[string]any{
				"projectMilestones": map[string]any{
					"nodes": milestones,
				},
			},
		},
	}
}

func milestoneCreateResponse(milestone map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projectMilestoneCreate": map[string]any{
				"success":          true,
				"projectMilestone": milestone,
			},
		},
	}
}

func milestoneUpdateResponse(milestone map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projectMilestoneUpdate": map[string]any{
				"success":          true,
				"projectMilestone": milestone,
			},
		},
	}
}

func milestoneDeleteResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projectMilestoneDelete": map[string]any{
				"success": success,
			},
		},
	}
}

func strPtr(s string) *string { return &s }

func TestMilestoneListCommand_Basic(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	milestones := []map[string]any{
		makeMilestone("ms-1", "v1.0 Release", "unstarted", strPtr("2026-06-01"), nil),
		makeMilestone("ms-2", "Beta", "done", nil, strPtr("Beta description")),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		milestoneListResponse(milestones),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "list", projID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["projectId"] != projID {
		t.Errorf("projectId = %v, want %s", (*bodies)[0]["projectId"], projID)
	}

	result := out.String()
	if !strings.Contains(result, "v1.0 Release") {
		t.Errorf("output should contain milestone name, got: %s", result)
	}
	if !strings.Contains(result, "2026-06-01") {
		t.Errorf("output should contain target date, got: %s", result)
	}
}

func TestMilestoneListCommand_JSONOutput(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	milestones := []map[string]any{
		makeMilestone("ms-1", "v1.0", "unstarted", strPtr("2026-06-01"), nil),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		milestoneListResponse(milestones),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "project", "milestone", "list", projID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 item, got %d", len(decoded))
	}
	if decoded[0]["name"] != "v1.0" {
		t.Errorf("expected name v1.0, got %v", decoded[0]["name"])
	}
}

func TestMilestoneListCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when project id is missing")
	}
}

func TestMilestoneCreateCommand_Basic(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	ms := makeMilestone("ms-new", "v2.0", "unstarted", strPtr("2026-12-01"), nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		milestoneCreateResponse(ms),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "create", projID, "--name", "v2.0"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	input, ok := (*bodies)[0]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[0])
	}
	if input["projectId"] != projID {
		t.Errorf("projectId = %v, want %s", input["projectId"], projID)
	}
	if input["name"] != "v2.0" {
		t.Errorf("name = %v, want v2.0", input["name"])
	}
	if _, present := input["description"]; present {
		t.Errorf("description should not be in input when not provided")
	}
	if _, present := input["targetDate"]; present {
		t.Errorf("targetDate should not be in input when not provided")
	}
}

func TestMilestoneCreateCommand_WithOptionals(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000002"
	ms := makeMilestone("ms-3", "v3.0", "unstarted", strPtr("2027-01-01"), strPtr("Third version"))

	server, bodies := newQueuedServer(t, []map[string]any{
		milestoneCreateResponse(ms),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{
		"project", "milestone", "create", projID,
		"--name", "v3.0",
		"--description", "Third version",
		"--target-date", "2027-01-01",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := (*bodies)[0]["input"].(map[string]any)
	if input["description"] != "Third version" {
		t.Errorf("description = %v, want 'Third version'", input["description"])
	}
	if input["targetDate"] != "2027-01-01" {
		t.Errorf("targetDate = %v, want 2027-01-01", input["targetDate"])
	}
}

func TestMilestoneCreateCommand_MissingName(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "create", projID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when name is missing")
	}
}

func TestMilestoneCreateCommand_SuccessFalse(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"projectMilestoneCreate": map[string]any{
					"success":          false,
					"projectMilestone": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "create", projID, "--name", "test"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestMilestoneUpdateCommand_Basic(t *testing.T) {
	const msID = "ms-update-1"
	ms := makeMilestone(msID, "Updated Name", "next", nil, nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		milestoneUpdateResponse(ms),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "update", msID, "--name", "Updated Name"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != msID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], msID)
	}
	input := (*bodies)[0]["input"].(map[string]any)
	if input["name"] != "Updated Name" {
		t.Errorf("name = %v, want 'Updated Name'", input["name"])
	}
}

func TestMilestoneUpdateCommand_NoFlags(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "update", "ms-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no flags are provided")
	}
	if !strings.Contains(err.Error(), "no fields to update") {
		t.Errorf("error should mention no fields to update, got: %v", err)
	}
}

func TestMilestoneUpdateCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "update"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when milestone id is missing")
	}
}

func TestMilestoneDeleteCommand_Basic(t *testing.T) {
	const msID = "ms-delete-1"

	server, bodies := newQueuedServer(t, []map[string]any{
		milestoneDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "delete", msID, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != msID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], msID)
	}

	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("output should contain 'deleted', got: %s", out.String())
	}
}

func TestMilestoneDeleteCommand_Aborted(t *testing.T) {
	server, bodies := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"project", "milestone", "delete", "ms-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user declines confirmation")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("error should mention aborted, got: %v", err)
	}
	if len(*bodies) != 0 {
		t.Errorf("expected 0 requests, got %d", len(*bodies))
	}
}

func TestMilestoneDeleteCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "delete"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when milestone id is missing")
	}
}

func TestMilestoneDeleteCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		milestoneDeleteResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "milestone", "delete", "ms-1", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}
