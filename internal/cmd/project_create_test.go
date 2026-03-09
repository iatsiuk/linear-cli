package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func projectCreateResponse(project map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projectCreate": map[string]any{
				"success": true,
				"project": project,
			},
		},
	}
}

func TestProjectCreateCommand_Basic(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"
	p := makeProject("proj-new", "My Project", "planned", "", 0.0, "")

	server, bodies := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		projectCreateResponse(p),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "create", "--name", "My Project", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "My Project") {
		t.Errorf("output should contain project name, got: %s", result)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[1]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set in mutation vars: %v", mutationVars)
	}
	if input["name"] != "My Project" {
		t.Errorf("name = %v, want My Project", input["name"])
	}
	teamIDs, ok := input["teamIds"].([]any)
	if !ok {
		t.Fatalf("teamIds not set in input: %v", input)
	}
	if len(teamIDs) != 1 || teamIDs[0] != teamID {
		t.Errorf("teamIds = %v, want [%s]", teamIDs, teamID)
	}
}

func TestProjectCreateCommand_JSONOutput(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"
	p := makeProject("proj-new", "My Project", "planned", "", 0.0, "")

	server, _ := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		projectCreateResponse(p),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "project", "create", "--name", "My Project", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "My Project" {
		t.Errorf("expected name My Project, got %v", decoded["name"])
	}
}

func TestProjectCreateCommand_MissingName(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "create", "--team", "ENG"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention name, got: %v", err)
	}
}

func TestProjectCreateCommand_MissingTeam(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "create", "--name", "Test"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --team is missing")
	}
	if !strings.Contains(err.Error(), "team") {
		t.Errorf("error should mention team, got: %v", err)
	}
}

func TestProjectCreateCommand_OptionalFields(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"
	p := makeProject("proj-opt", "Opt Project", "planned", "", 0.0, "2026-12-31")

	server, bodies := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		projectCreateResponse(p),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{
		"project", "create",
		"--name", "Opt Project",
		"--team", "ENG",
		"--description", "desc",
		"--color", "#FF0000",
		"--target-date", "2026-12-31",
		"--start-date", "2026-01-01",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	input := (*bodies)[1]["input"].(map[string]any)
	if input["description"] != "desc" {
		t.Errorf("description = %v, want desc", input["description"])
	}
	if input["color"] != "#FF0000" {
		t.Errorf("color = %v, want #FF0000", input["color"])
	}
	if input["targetDate"] != "2026-12-31" {
		t.Errorf("targetDate = %v, want 2026-12-31", input["targetDate"])
	}
	if input["startDate"] != "2026-01-01" {
		t.Errorf("startDate = %v, want 2026-01-01", input["startDate"])
	}
}

func TestProjectCreateCommand_SuccessFalse(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"

	server, _ := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		{
			"data": map[string]any{
				"projectCreate": map[string]any{
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
	root.SetArgs([]string{"project", "create", "--name", "Test", "--team", "ENG"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}
