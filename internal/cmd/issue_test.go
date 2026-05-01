package cmd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
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
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, fmt.Sprintf("encode: %v", err), http.StatusInternalServerError)
	}
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

func TestIssueListCommand_AssigneeFilter(t *testing.T) {

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
	root.SetArgs([]string{"issue", "list", "--assignee", "Alice"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	assignee, ok := filter["assignee"].(map[string]any)
	if !ok {
		t.Fatalf("filter.assignee not set, got: %v", filter["assignee"])
	}
	displayName, ok := assignee["displayName"].(map[string]any)
	if !ok {
		t.Fatalf("filter.assignee.displayName not set, got: %v", assignee["displayName"])
	}
	if displayName["eq"] != "Alice" {
		t.Errorf("filter.assignee.displayName.eq = %v, want Alice", displayName["eq"])
	}
}

func TestIssueListCommand_CreatedAfterFilter(t *testing.T) {

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
	root.SetArgs([]string{"issue", "list", "--created-after", "7d"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	createdAt, ok := filter["createdAt"].(map[string]any)
	if !ok {
		t.Fatalf("filter.createdAt not set, got: %v", filter)
	}
	if createdAt["gt"] != "-P7D" {
		t.Errorf("filter.createdAt.gt = %v, want -P7D", createdAt["gt"])
	}
}

func TestIssueListCommand_PriorityGteFilter(t *testing.T) {

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
	root.SetArgs([]string{"issue", "list", "--priority-gte", "2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	priority, ok := filter["priority"].(map[string]any)
	if !ok {
		t.Fatalf("filter.priority not set, got: %v", filter)
	}
	if priority["gte"] != float64(2) {
		t.Errorf("filter.priority.gte = %v, want 2", priority["gte"])
	}
}

func TestIssueListCommand_NoAssigneeFilter(t *testing.T) {

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
	root.SetArgs([]string{"issue", "list", "--no-assignee"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	assignee, ok := filter["assignee"].(map[string]any)
	if !ok {
		t.Fatalf("filter.assignee not set, got: %v", filter)
	}
	if assignee["null"] != true {
		t.Errorf("filter.assignee.null = %v, want true", assignee["null"])
	}
}

func TestIssueListCommand_MyFilter(t *testing.T) {

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
	root.SetArgs([]string{"issue", "list", "--my"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	assignee, ok := filter["assignee"].(map[string]any)
	if !ok {
		t.Fatalf("filter.assignee not set, got: %v", filter)
	}
	isMe, ok := assignee["isMe"].(map[string]any)
	if !ok {
		t.Fatalf("filter.assignee.isMe not set, got: %v", assignee)
	}
	if isMe["eq"] != true {
		t.Errorf("filter.assignee.isMe.eq = %v, want true", isMe["eq"])
	}
}

func TestIssueListCommand_CombinedTeamAndAdvancedFilters(t *testing.T) {

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
	root.SetArgs([]string{"issue", "list", "--team", "ENG", "--created-after", "7d", "--priority-gte", "2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	// team from base flags
	if _, ok := filter["team"]; !ok {
		t.Error("filter.team not set")
	}
	// createdAt from advanced filter
	if _, ok := filter["createdAt"]; !ok {
		t.Error("filter.createdAt not set")
	}
	// priority from advanced filter
	if _, ok := filter["priority"]; !ok {
		t.Error("filter.priority not set")
	}
}

func TestIssueListCommand_OrderByFlag(t *testing.T) {

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
	root.SetArgs([]string{"issue", "list", "--order-by", "createdAt"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["orderBy"] != "createdAt" {
		t.Errorf("variables.orderBy = %v, want createdAt", gotVars["orderBy"])
	}
}

func TestIssueListCommand_InvalidLimit(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	for _, limit := range []string{"0", "-1"} {
		t.Run("limit="+limit, func(t *testing.T) {
			var out bytes.Buffer
			root := cmd.NewRootCommand("test")
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs([]string{"issue", "list", "--limit", limit})

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error for --limit %s", limit)
			}
		})
	}
}

func TestIssueListCommand_OrWithBaseAndAdvancedFilters(t *testing.T) {
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
	root.SetArgs([]string{"issue", "list", "--team", "ENG", "--no-assignee", "--or"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	orList, ok := filter["or"].([]any)
	if !ok {
		t.Fatalf("filter.or not found or wrong type: %v", filter)
	}
	if len(orList) != 2 {
		t.Errorf("filter.or len = %d, want 2", len(orList))
	}
	// verify top level has no team or assignee keys (all moved into or)
	if _, ok := filter["team"]; ok {
		t.Error("filter.team should not exist at top level when --or is set")
	}
}

func TestIssueListCommand_OrSkipsMutualExclusion(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--assignee", "alice", "--no-assignee", "--or"})

	if err := root.Execute(); err != nil {
		t.Fatalf("--assignee + --no-assignee + --or should not return error, got: %v", err)
	}
}

func TestIssueListCommand_LabelFilter(t *testing.T) {
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
	root.SetArgs([]string{"issue", "list", "--label", "bug"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	labels, ok := filter["labels"].(map[string]any)
	if !ok {
		t.Fatalf("filter.labels not set, got: %v", filter["labels"])
	}
	some, ok := labels["some"].(map[string]any)
	if !ok {
		t.Fatalf("filter.labels.some not set, got: %v", labels["some"])
	}
	name, ok := some["name"].(map[string]any)
	if !ok {
		t.Fatalf("filter.labels.some.name not set, got: %v", some["name"])
	}
	if name["eq"] != "bug" {
		t.Errorf("filter.labels.some.name.eq = %v, want bug", name["eq"])
	}
}

func TestIssueListCommand_ProjectFilter(t *testing.T) {
	const projectUUID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	var reqCount int
	var gotVars map[string]any

	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		reqCount++

		if strings.Contains(body.Query, "ResolveProject") {
			writeJSONResponse(w, map[string]any{
				"data": map[string]any{
					"projects": map[string]any{
						"nodes": []any{map[string]any{"id": projectUUID}},
					},
				},
			})
			return
		}

		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--project", "My Project"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reqCount != 2 {
		t.Errorf("expected 2 requests (resolve + list), got %d", reqCount)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	project, ok := filter["project"].(map[string]any)
	if !ok {
		t.Fatalf("filter.project not set, got: %v", filter["project"])
	}
	id, ok := project["id"].(map[string]any)
	if !ok {
		t.Fatalf("filter.project.id not set, got: %v", project["id"])
	}
	if id["eq"] != projectUUID {
		t.Errorf("filter.project.id.eq = %v, want %s", id["eq"], projectUUID)
	}
}

func TestIssueListCommand_ProjectFilter_MutualExclusive(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--project", "My Project", "--no-project"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --project and --no-project used together")
	}
	if !strings.Contains(err.Error(), "--project") || !strings.Contains(err.Error(), "--no-project") {
		t.Errorf("error should mention --project and --no-project, got: %v", err)
	}
}

func TestIssueList_ParentFilter_Identifier(t *testing.T) {
	const parentUUID = "11111111-2222-3333-4444-555555555555"
	var reqCount int
	var gotVars map[string]any

	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		reqCount++

		if strings.Contains(body.Query, "ResolveIssue") {
			if body.Variables["id"] != "ENG-727" {
				t.Errorf("ResolveIssue variable id = %v, want ENG-727", body.Variables["id"])
			}
			writeJSONResponse(w, map[string]any{
				"data": map[string]any{
					"issue": map[string]any{"id": parentUUID},
				},
			})
			return
		}

		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--parent", "ENG-727"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reqCount != 2 {
		t.Errorf("expected 2 requests (resolve + list), got %d", reqCount)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	parent, ok := filter["parent"].(map[string]any)
	if !ok {
		t.Fatalf("filter.parent not set, got: %v", filter["parent"])
	}
	id, ok := parent["id"].(map[string]any)
	if !ok {
		t.Fatalf("filter.parent.id not set, got: %v", parent["id"])
	}
	if id["eq"] != parentUUID {
		t.Errorf("filter.parent.id.eq = %v, want %s", id["eq"], parentUUID)
	}
}

func TestIssueList_ParentFilter_UUID(t *testing.T) {
	const parentUUID = "11111111-2222-3333-4444-555555555555"
	var reqCount int
	var gotVars map[string]any

	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		reqCount++

		if strings.Contains(body.Query, "ResolveIssue") {
			t.Errorf("unexpected ResolveIssue call for UUID input")
			writeJSONResponse(w, map[string]any{"data": map[string]any{"issue": nil}})
			return
		}

		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--parent", parentUUID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reqCount != 1 {
		t.Errorf("expected 1 request (list only), got %d", reqCount)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	parent, ok := filter["parent"].(map[string]any)
	if !ok {
		t.Fatalf("filter.parent not set, got: %v", filter["parent"])
	}
	id, ok := parent["id"].(map[string]any)
	if !ok {
		t.Fatalf("filter.parent.id not set, got: %v", parent["id"])
	}
	if id["eq"] != parentUUID {
		t.Errorf("filter.parent.id.eq = %v, want %s", id["eq"], parentUUID)
	}
}

func TestIssueList_ParentFilter_NotFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if strings.Contains(body.Query, "ResolveIssue") {
			writeJSONResponse(w, map[string]any{
				"data": map[string]any{"issue": nil},
			})
			return
		}
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--parent", "ENG-999"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when parent issue not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestIssueList_ParentFilter_CombinedWithOther(t *testing.T) {
	const parentUUID = "11111111-2222-3333-4444-555555555555"
	var gotVars map[string]any

	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)

		if strings.Contains(body.Query, "ResolveIssue") {
			writeJSONResponse(w, map[string]any{
				"data": map[string]any{
					"issue": map[string]any{"id": parentUUID},
				},
			})
			return
		}

		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "list", "--parent", "ENG-727", "--state", "In Progress"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	parent, ok := filter["parent"].(map[string]any)
	if !ok {
		t.Fatalf("filter.parent not set, got: %v", filter["parent"])
	}
	pid, ok := parent["id"].(map[string]any)
	if !ok {
		t.Fatalf("filter.parent.id not set, got: %v", parent["id"])
	}
	if pid["eq"] != parentUUID {
		t.Errorf("filter.parent.id.eq = %v, want %s", pid["eq"], parentUUID)
	}
	state, ok := filter["state"].(map[string]any)
	if !ok {
		t.Fatalf("filter.state not set, got: %v", filter["state"])
	}
	name, ok := state["name"].(map[string]any)
	if !ok {
		t.Fatalf("filter.state.name not set, got: %v", state["name"])
	}
	if name["eq"] != "In Progress" {
		t.Errorf("filter.state.name.eq = %v, want In Progress", name["eq"])
	}
}

func TestIssueListCommand_LabelFilter_JSON(t *testing.T) {
	var gotVars map[string]any
	issues := []map[string]any{
		makeIssue("ENG-1", "Bug fix", "Done", "High", ""),
	}
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, issueListResponse(issues))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "issue", "list", "--label", "bug"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// verify filter was set
	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set")
	}
	if _, ok := filter["labels"]; !ok {
		t.Error("filter.labels not set")
	}

	// verify JSON output
	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 issue in JSON output, got %d", len(decoded))
	}
}
