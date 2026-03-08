package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func projectUpdateResponse(project map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projectUpdate": map[string]any{
				"success": true,
				"project": project,
			},
		},
	}
}

func TestProjectUpdateCommand_Basic(t *testing.T) {
	p := makeProject("proj-1", "Updated Name", "started", "onTrack", 0.5, "")

	server, bodies := newQueuedServer(t, []map[string]any{
		projectUpdateResponse(p),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "proj-1", "--name", "Updated Name"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	vars := (*bodies)[0]
	if vars["id"] != "proj-1" {
		t.Errorf("id = %v, want proj-1", vars["id"])
	}
	input, ok := vars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", vars)
	}
	if input["name"] != "Updated Name" {
		t.Errorf("name = %v, want Updated Name", input["name"])
	}
}

func TestProjectUpdateCommand_JSONOutput(t *testing.T) {
	p := makeProject("proj-1", "Updated Name", "started", "onTrack", 0.5, "")

	server, _ := newQueuedServer(t, []map[string]any{
		projectUpdateResponse(p),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "project", "update", "proj-1", "--name", "Updated Name"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Updated Name" {
		t.Errorf("expected name Updated Name, got %v", decoded["name"])
	}
}

func TestProjectUpdateCommand_PartialUpdate(t *testing.T) {
	p := makeProject("proj-2", "Same Name", "paused", "atRisk", 0.3, "2026-09-01")

	server, bodies := newQueuedServer(t, []map[string]any{
		projectUpdateResponse(p),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "proj-2", "--health", "atRisk", "--target-date", "2026-09-01"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := (*bodies)[0]["input"].(map[string]any)
	// only health and targetDate should be present
	if _, present := input["name"]; present {
		t.Errorf("name should not be in input when not provided")
	}
	if input["health"] != "atRisk" {
		t.Errorf("health = %v, want atRisk", input["health"])
	}
	if input["targetDate"] != "2026-09-01" {
		t.Errorf("targetDate = %v, want 2026-09-01", input["targetDate"])
	}
}

func TestProjectUpdateCommand_NoFlags(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "proj-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no flags provided")
	}
	if !strings.Contains(err.Error(), "no fields to update") {
		t.Errorf("error should mention no fields to update, got: %v", err)
	}
}

func TestProjectUpdateCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when id is missing")
	}
}

func TestProjectUpdateCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"projectUpdate": map[string]any{
					"success": false,
					"project": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "proj-1", "--name", "Fail"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestProjectUpdateCommand_AllFlags(t *testing.T) {
	p := makeProject("proj-3", "Full Update", "completed", "onTrack", 1.0, "2026-12-01")

	server, bodies := newQueuedServer(t, []map[string]any{
		projectUpdateResponse(p),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{
		"project", "update", "proj-3",
		"--name", "Full Update",
		"--description", "new desc",
		"--state", "completed",
		"--target-date", "2026-12-01",
		"--start-date", "2026-06-01",
		"--health", "onTrack",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := (*bodies)[0]["input"].(map[string]any)
	if input["name"] != "Full Update" {
		t.Errorf("name = %v, want Full Update", input["name"])
	}
	if input["description"] != "new desc" {
		t.Errorf("description = %v, want new desc", input["description"])
	}
	if input["statusType"] != "completed" {
		t.Errorf("statusType = %v, want completed", input["statusType"])
	}
	if input["targetDate"] != "2026-12-01" {
		t.Errorf("targetDate = %v, want 2026-12-01", input["targetDate"])
	}
	if input["startDate"] != "2026-06-01" {
		t.Errorf("startDate = %v, want 2026-06-01", input["startDate"])
	}
	if input["health"] != "onTrack" {
		t.Errorf("health = %v, want onTrack", input["health"])
	}
}
