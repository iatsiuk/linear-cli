package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func projectListResponse(projects []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projects": map[string]any{
				"nodes":    projects,
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			},
		},
	}
}

func makeProject(id, name, statusType, health string, progress float64, targetDate string) map[string]any {
	p := map[string]any{
		"id":          id,
		"name":        name,
		"description": "test project",
		"color":       "#FF0000",
		"icon":        nil,
		"progress":    progress,
		"url":         "https://linear.app/project/" + id,
		"createdAt":   "2026-01-01T00:00:00Z",
		"updatedAt":   "2026-01-02T00:00:00Z",
		"status": map[string]any{
			"id":   "status-1",
			"name": statusType,
			"type": statusType,
		},
		"teams": map[string]any{
			"nodes": []any{
				map[string]any{"id": "team-1", "name": "Engineering", "key": "ENG"},
			},
		},
		"creator": map[string]any{
			"id":          "user-1",
			"displayName": "Alice",
			"email":       "alice@example.com",
		},
	}
	if health != "" {
		p["health"] = health
	}
	if targetDate != "" {
		p["targetDate"] = targetDate
	}
	return p
}

func projectGetResponse(project map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"project": project,
		},
	}
}

func TestProjectListCommand_TableOutput(t *testing.T) {
	projects := []map[string]any{
		makeProject("proj-1", "Auth Redesign", "started", "onTrack", 0.45, "2026-06-01"),
		makeProject("proj-2", "Mobile App", "planned", "", 0.0, ""),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, projectListResponse(projects))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "NAME") {
		t.Errorf("output should contain NAME header, got:\n%s", result)
	}
	if !strings.Contains(result, "STATUS") {
		t.Errorf("output should contain STATUS header, got:\n%s", result)
	}
	if !strings.Contains(result, "HEALTH") {
		t.Errorf("output should contain HEALTH header, got:\n%s", result)
	}
	if !strings.Contains(result, "PROGRESS") {
		t.Errorf("output should contain PROGRESS header, got:\n%s", result)
	}
	if !strings.Contains(result, "Auth Redesign") {
		t.Errorf("output should contain project name, got:\n%s", result)
	}
	if !strings.Contains(result, "started") {
		t.Errorf("output should contain status, got:\n%s", result)
	}
	if !strings.Contains(result, "onTrack") {
		t.Errorf("output should contain health, got:\n%s", result)
	}
	if !strings.Contains(result, "2026-06-01") {
		t.Errorf("output should contain target date, got:\n%s", result)
	}
}

func TestProjectListCommand_JSONOutput(t *testing.T) {
	projects := []map[string]any{
		makeProject("proj-1", "Auth Redesign", "started", "onTrack", 0.5, "2026-06-01"),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, projectListResponse(projects))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "project", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 project, got %d", len(decoded))
	}
	if decoded[0]["name"] != "Auth Redesign" {
		t.Errorf("expected name Auth Redesign, got %v", decoded[0]["name"])
	}
}

func TestProjectListCommand_TeamFilter(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, projectListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	teams, ok := filter["accessibleTeams"].(map[string]any)
	if !ok {
		t.Fatalf("filter.accessibleTeams not set, got: %v", filter["accessibleTeams"])
	}
	some, ok := teams["some"].(map[string]any)
	if !ok {
		t.Fatalf("filter.accessibleTeams.some not set, got: %v", teams["some"])
	}
	key, ok := some["key"].(map[string]any)
	if !ok {
		t.Fatalf("filter.accessibleTeams.some.key not set, got: %v", some["key"])
	}
	if key["eq"] != "ENG" {
		t.Errorf("team key eq = %v, want ENG", key["eq"])
	}
}

func TestProjectListCommand_StatusFilter(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, projectListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "list", "--status", "started"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	status, ok := filter["status"].(map[string]any)
	if !ok {
		t.Fatalf("filter.status not set, got: %v", filter["status"])
	}
	typ, ok := status["type"].(map[string]any)
	if !ok {
		t.Fatalf("filter.status.type not set, got: %v", status["type"])
	}
	if typ["eq"] != "started" {
		t.Errorf("filter.status.type.eq = %v, want started", typ["eq"])
	}
}

func TestProjectListCommand_HealthFilter(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, projectListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "list", "--health", "atRisk"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	health, ok := filter["health"].(map[string]any)
	if !ok {
		t.Fatalf("filter.health not set, got: %v", filter["health"])
	}
	if health["eq"] != "atRisk" {
		t.Errorf("filter.health.eq = %v, want atRisk", health["eq"])
	}
}

func TestProjectListCommand_LimitFlag(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, projectListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "list", "--limit", "5"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first, ok := gotVars["first"].(float64)
	if !ok {
		t.Fatalf("variables.first not set, got: %v (%T)", gotVars["first"], gotVars["first"])
	}
	if int(first) != 5 {
		t.Errorf("variables.first = %v, want 5", first)
	}
}

func TestProjectListCommand_IncludeArchived(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, projectListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "list", "--include-archived"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["includeArchived"] != true {
		t.Errorf("variables.includeArchived = %v, want true", gotVars["includeArchived"])
	}
}

func TestProjectListCommand_OrderByFlag(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, projectListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "list", "--order-by", "createdAt"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["orderBy"] != "createdAt" {
		t.Errorf("variables.orderBy = %v, want createdAt", gotVars["orderBy"])
	}
}

func TestProjectShowCommand_TableOutput(t *testing.T) {
	project := makeProject("proj-1", "Auth Redesign", "started", "onTrack", 0.65, "2026-12-31")

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, projectGetResponse(project))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "show", "proj-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Auth Redesign") {
		t.Errorf("output should contain project name, got:\n%s", result)
	}
	if !strings.Contains(result, "started") {
		t.Errorf("output should contain status, got:\n%s", result)
	}
	if !strings.Contains(result, "onTrack") {
		t.Errorf("output should contain health, got:\n%s", result)
	}
	if !strings.Contains(result, "2026-12-31") {
		t.Errorf("output should contain target date, got:\n%s", result)
	}
	if !strings.Contains(result, "Engineering") {
		t.Errorf("output should contain team name, got:\n%s", result)
	}
}

func TestProjectShowCommand_JSONOutput(t *testing.T) {
	project := makeProject("proj-1", "Auth Redesign", "started", "onTrack", 0.65, "2026-12-31")

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, projectGetResponse(project))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "project", "show", "proj-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Auth Redesign" {
		t.Errorf("expected name Auth Redesign, got %v", decoded["name"])
	}
}

func TestProjectShowCommand_NotFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{"data": map[string]any{"project": nil}})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "show", "nonexistent"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for not found project")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestProjectShowCommand_MissingID(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when ID is missing")
	}
}
