package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func makeState(id, name, stateType, color string, position float64, teamKey string) map[string]any {
	s := map[string]any{
		"id":        id,
		"name":      name,
		"type":      stateType,
		"color":     color,
		"position":  position,
		"createdAt": "2026-01-01T00:00:00Z",
	}
	if teamKey != "" {
		s["team"] = map[string]any{
			"id":   "team-" + teamKey,
			"name": teamKey + " Team",
			"key":  teamKey,
		}
	}
	return s
}

func stateListResponse(states []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"workflowStates": map[string]any{
				"nodes":    states,
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			},
		},
	}
}

// TestStateListCommand_TableOutput verifies table output for state list.
func TestStateListCommand_TableOutput(t *testing.T) {
	states := []map[string]any{
		makeState("s1", "Backlog", "backlog", "#808080", 0, "ENG"),
		makeState("s2", "In Progress", "started", "#0000FF", 1, "ENG"),
		makeState("s3", "Done", "completed", "#00FF00", 2, "ENG"),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, stateListResponse(states))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"state", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Backlog") {
		t.Errorf("output should contain Backlog state, got:\n%s", result)
	}
	if !strings.Contains(result, "In Progress") {
		t.Errorf("output should contain In Progress state, got:\n%s", result)
	}
	if !strings.Contains(result, "Done") {
		t.Errorf("output should contain Done state, got:\n%s", result)
	}
}

// TestStateListCommand_GroupedByType verifies that output groups states by type.
func TestStateListCommand_GroupedByType(t *testing.T) {
	states := []map[string]any{
		makeState("s1", "In Progress", "started", "#0000FF", 0, "ENG"),
		makeState("s2", "Todo", "unstarted", "#808080", 0, "ENG"),
		makeState("s3", "Backlog", "backlog", "#AAAAAA", 0, "ENG"),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, stateListResponse(states))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"state", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	// verify group headers appear
	if !strings.Contains(result, "Backlog") {
		t.Errorf("output should contain Backlog group header, got:\n%s", result)
	}
	if !strings.Contains(result, "Unstarted") {
		t.Errorf("output should contain Unstarted group header, got:\n%s", result)
	}
	if !strings.Contains(result, "Started") {
		t.Errorf("output should contain Started group header, got:\n%s", result)
	}
	// backlog should appear before started in the output
	backlogIdx := strings.Index(result, "Backlog")
	startedIdx := strings.Index(result, "In Progress")
	if backlogIdx > startedIdx {
		t.Errorf("Backlog group should appear before Started group, got:\n%s", result)
	}
}

// TestStateListCommand_TeamFilter verifies that --team sets the filter variable.
func TestStateListCommand_TeamFilter(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, stateListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"state", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("filter not set, got: %v", gotVars["filter"])
	}
	team, ok := filter["team"].(map[string]any)
	if !ok {
		t.Fatalf("filter.team not set, got: %v", filter["team"])
	}
	key, ok := team["key"].(map[string]any)
	if !ok {
		t.Fatalf("filter.team.key not set, got: %v", team["key"])
	}
	if key["eq"] != "ENG" {
		t.Errorf("filter.team.key.eq = %v, want ENG", key["eq"])
	}
}

// TestStateListCommand_RequiresTeam verifies that --team flag is required.
func TestStateListCommand_RequiresTeam(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"state", "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --team is missing")
	}
}

// TestStateListCommand_JSONOutput verifies JSON output for state list.
func TestStateListCommand_JSONOutput(t *testing.T) {
	states := []map[string]any{
		makeState("s1", "Backlog", "backlog", "#808080", 0, "ENG"),
		makeState("s2", "In Progress", "started", "#0000FF", 1, "ENG"),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, stateListResponse(states))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "state", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 2 {
		t.Fatalf("expected 2 states, got %d", len(decoded))
	}
}
