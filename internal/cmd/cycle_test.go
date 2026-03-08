package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func cycleListResponse(cycles []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"cycles": map[string]any{
				"nodes":    cycles,
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			},
		},
	}
}

func makeCycle(id string, number float64, name string, isActive, isFuture, isPast bool, progress float64) map[string]any {
	return map[string]any{
		"id":          id,
		"name":        name,
		"number":      number,
		"description": nil,
		"startsAt":    "2026-01-01T00:00:00Z",
		"endsAt":      "2026-01-14T00:00:00Z",
		"isActive":    isActive,
		"isFuture":    isFuture,
		"isPast":      isPast,
		"progress":    progress,
		"createdAt":   "2026-01-01T00:00:00Z",
		"updatedAt":   "2026-01-01T00:00:00Z",
		"team": map[string]any{
			"id":   "team-1",
			"name": "Engineering",
			"key":  "ENG",
		},
	}
}

func cycleGetResponse(cycle map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"cycle": cycle,
		},
	}
}

func TestCycleListCommand_TableOutput(t *testing.T) {
	cycles := []map[string]any{
		makeCycle("cycle-1", 1, "Sprint 1", false, false, true, 1.0),
		makeCycle("cycle-2", 2, "Sprint 2", true, false, false, 0.5),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, cycleListResponse(cycles))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cycle", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, want := range []string{"NUMBER", "NAME", "START", "END", "PROGRESS", "STATUS"} {
		if !strings.Contains(result, want) {
			t.Errorf("output should contain %s header, got:\n%s", want, result)
		}
	}
	if !strings.Contains(result, "Sprint 1") {
		t.Errorf("output should contain cycle name, got:\n%s", result)
	}
	if !strings.Contains(result, "Past") {
		t.Errorf("output should contain Past status, got:\n%s", result)
	}
	if !strings.Contains(result, "Active") {
		t.Errorf("output should contain Active status, got:\n%s", result)
	}
}

func TestCycleListCommand_JSONOutput(t *testing.T) {
	cycles := []map[string]any{
		makeCycle("cycle-1", 1, "Sprint 1", true, false, false, 0.3),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, cycleListResponse(cycles))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "cycle", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 cycle, got %d", len(decoded))
	}
	if decoded[0]["name"] != "Sprint 1" {
		t.Errorf("expected name Sprint 1, got %v", decoded[0]["name"])
	}
}

func TestCycleListCommand_TeamFilter(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, cycleListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cycle", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
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

func TestCycleShowCommand_DetailOutput(t *testing.T) {
	cycle := makeCycle("cycle-1", 3, "Sprint 3", true, false, false, 0.65)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, cycleGetResponse(cycle))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cycle", "show", "cycle-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, want := range []string{"Sprint 3", "Active", "65%", "Engineering"} {
		if !strings.Contains(result, want) {
			t.Errorf("output should contain %q, got:\n%s", want, result)
		}
	}
}

func TestCycleShowCommand_JSONOutput(t *testing.T) {
	cycle := makeCycle("cycle-1", 1, "Sprint 1", false, true, false, 0.0)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, cycleGetResponse(cycle))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "cycle", "show", "cycle-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Sprint 1" {
		t.Errorf("expected name Sprint 1, got %v", decoded["name"])
	}
}

func TestCycleShowCommand_NotFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{"data": map[string]any{"cycle": nil}})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cycle", "show", "nonexistent"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for not found cycle")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestCycleShowCommand_MissingID(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cycle", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when ID is missing")
	}
}

func TestCycleActiveCommand_ShowsActiveCycle(t *testing.T) {
	var gotVars map[string]any
	cycles := []map[string]any{
		makeCycle("cycle-2", 2, "Sprint 2", true, false, false, 0.4),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, cycleListResponse(cycles))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cycle", "active", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// verify filter sent correctly
	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	isActive, ok := filter["isActive"].(map[string]any)
	if !ok {
		t.Fatalf("filter.isActive not set, got: %v", filter["isActive"])
	}
	if isActive["eq"] != true {
		t.Errorf("filter.isActive.eq = %v, want true", isActive["eq"])
	}

	result := out.String()
	if !strings.Contains(result, "Sprint 2") {
		t.Errorf("output should contain cycle name, got:\n%s", result)
	}
	if !strings.Contains(result, "Active") {
		t.Errorf("output should contain Active status, got:\n%s", result)
	}
}

func TestCycleActiveCommand_NoCycleFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, cycleListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"cycle", "active", "--team", "ENG"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no active cycle found")
	}
	if !strings.Contains(err.Error(), "no active cycle") {
		t.Errorf("error should mention no active cycle, got: %v", err)
	}
}
