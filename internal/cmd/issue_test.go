package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

// issueListResponse builds a minimal GraphQL response for issue list queries.
func issueListResponse(issues []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issues": map[string]any{
				"nodes":    issues,
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			},
		},
	}
}

func makeIssue(identifier, title, stateName, priorityLabel string, assigneeName string) map[string]any {
	issue := map[string]any{
		"id":            "id-" + identifier,
		"identifier":    identifier,
		"title":         title,
		"priority":      1.0,
		"priorityLabel": priorityLabel,
		"url":           "https://linear.app/issue/" + identifier,
		"createdAt":     "2026-01-01T00:00:00Z",
		"updatedAt":     "2026-01-02T00:00:00Z",
		"state": map[string]any{
			"id":    "state-1",
			"name":  stateName,
			"color": "#000000",
			"type":  "started",
		},
		"team": map[string]any{
			"id":   "team-1",
			"name": "Engineering",
			"key":  "ENG",
		},
		"labels": map[string]any{"nodes": []any{}},
	}
	if assigneeName != "" {
		issue["assignee"] = map[string]any{
			"id":          "user-1",
			"displayName": assigneeName,
			"email":       "user@example.com",
		}
	}
	return issue
}

func newIssueTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func setupIssueTest(t *testing.T, server *httptest.Server) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)
	t.Setenv("LINEAR_API_ENDPOINT", server.URL)
	err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("api_key: lin_api_testkey\n"), 0o600)
	if err != nil {
		t.Fatalf("setup config: %v", err)
	}
}

func writeJSONResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func TestIssueListCommand_TableOutput(t *testing.T) {

	issues := []map[string]any{
		makeIssue("ENG-1", "Fix the login bug", "In Progress", "Urgent", "Alice"),
		makeIssue("ENG-2", "Add dark mode", "Backlog", "Medium", ""),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(issues))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()

	// check header row
	if !strings.Contains(result, "ID") {
		t.Errorf("output should contain ID column header, got:\n%s", result)
	}
	if !strings.Contains(result, "TITLE") {
		t.Errorf("output should contain TITLE column header, got:\n%s", result)
	}
	if !strings.Contains(result, "STATUS") {
		t.Errorf("output should contain STATUS column header, got:\n%s", result)
	}
	if !strings.Contains(result, "PRIORITY") {
		t.Errorf("output should contain PRIORITY column header, got:\n%s", result)
	}
	if !strings.Contains(result, "ASSIGNEE") {
		t.Errorf("output should contain ASSIGNEE column header, got:\n%s", result)
	}

	// check data rows
	if !strings.Contains(result, "ENG-1") {
		t.Errorf("output should contain ENG-1, got:\n%s", result)
	}
	if !strings.Contains(result, "Fix the login bug") {
		t.Errorf("output should contain issue title, got:\n%s", result)
	}
	if !strings.Contains(result, "In Progress") {
		t.Errorf("output should contain state name, got:\n%s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("output should contain assignee name, got:\n%s", result)
	}
	if !strings.Contains(result, "ENG-2") {
		t.Errorf("output should contain ENG-2, got:\n%s", result)
	}
}

func TestIssueListCommand_JSONOutput(t *testing.T) {

	issues := []map[string]any{
		makeIssue("ENG-1", "Fix bug", "Done", "No priority", ""),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(issues))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "issue", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 issue in JSON output, got %d", len(decoded))
	}
	if decoded[0]["identifier"] != "ENG-1" {
		t.Errorf("expected identifier ENG-1, got %v", decoded[0]["identifier"])
	}
}

func TestIssueListCommand_LimitFlag(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--limit", "10"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first, ok := gotVars["first"].(float64)
	if !ok {
		t.Fatalf("variables.first not set or wrong type, got: %v (%T)", gotVars["first"], gotVars["first"])
	}
	if int(first) != 10 {
		t.Errorf("variables.first = %v, want 10", first)
	}
}

func TestIssueListCommand_TeamFilter(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--team", "ENG"})

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

func TestIssueListCommand_StateFilter(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--state", "In Progress"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	state, ok := filter["state"].(map[string]any)
	if !ok {
		t.Fatalf("filter.state not set")
	}
	name, ok := state["name"].(map[string]any)
	if !ok {
		t.Fatalf("filter.state.name not set")
	}
	if name["eq"] != "In Progress" {
		t.Errorf("filter.state.name.eq = %v, want In Progress", name["eq"])
	}
}

func TestIssueListCommand_PriorityFilter(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--priority", "1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	prio, ok := filter["priority"].(map[string]any)
	if !ok {
		t.Fatalf("filter.priority not set")
	}
	if prio["eq"] != float64(1) {
		t.Errorf("filter.priority.eq = %v, want 1.0", prio["eq"])
	}
}

func TestIssueListCommand_IncludeArchived(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--include-archived"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["includeArchived"] != true {
		t.Errorf("variables.includeArchived = %v, want true", gotVars["includeArchived"])
	}
}

func TestIssueListCommand_TitleTruncation(t *testing.T) {

	longTitle := strings.Repeat("a", 50)
	issues := []map[string]any{
		makeIssue("ENG-1", longTitle, "Todo", "Low", ""),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(issues))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if strings.Contains(result, longTitle) {
		t.Errorf("long title should be truncated, but full title appears in output")
	}
	if !strings.Contains(result, "...") {
		t.Errorf("truncated title should end with ..., got:\n%s", result)
	}
}

func TestIssueListCommand_EmptyResult(t *testing.T) {

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// empty table produces no output (TableFormatter returns nil for empty slice)
	if out.Len() != 0 {
		t.Errorf("expected empty output for no issues, got: %q", out.String())
	}
}

func TestIssueListCommand_NoAPIKey(t *testing.T) {

	dir := t.TempDir()
	t.Setenv("LINEAR_CONFIG_DIR", dir)
	t.Setenv("LINEAR_API_ENDPOINT", "http://127.0.0.1:1")
	t.Setenv("LINEAR_API_KEY", "") // ensure env key is cleared

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no API key configured")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("error should mention authentication, got: %v", err)
	}
}

func TestIssueListCommand_DefaultLimit(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first, ok := gotVars["first"].(float64)
	if !ok {
		t.Fatalf("variables.first not set, got: %v", gotVars["first"])
	}
	if int(first) != 50 {
		t.Errorf("default limit = %v, want 50", first)
	}
}
