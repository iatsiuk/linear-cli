package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func searchResponse(issues []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"searchIssues": map[string]any{
				"nodes": issues,
			},
		},
	}
}

func TestSearchCommand_TableOutput(t *testing.T) {
	issues := []map[string]any{
		makeIssue("ENG-1", "Fix the login bug", "In Progress", "Urgent", "Alice"),
		makeIssue("ENG-2", "Add search feature", "Backlog", "Normal", ""),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		searchResponse(issues),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "login"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["term"] != "login" {
		t.Errorf("term = %v, want login", (*bodies)[0]["term"])
	}

	result := out.String()
	if !strings.Contains(result, "ID") {
		t.Errorf("output should contain ID column header, got:\n%s", result)
	}
	if !strings.Contains(result, "TITLE") {
		t.Errorf("output should contain TITLE column header, got:\n%s", result)
	}
	if !strings.Contains(result, "STATUS") {
		t.Errorf("output should contain STATUS column header, got:\n%s", result)
	}
	if !strings.Contains(result, "TEAM") {
		t.Errorf("output should contain TEAM column header, got:\n%s", result)
	}
	if !strings.Contains(result, "ENG-1") {
		t.Errorf("output should contain ENG-1, got:\n%s", result)
	}
	if !strings.Contains(result, "Fix the login bug") {
		t.Errorf("output should contain issue title, got:\n%s", result)
	}
}

func TestSearchCommand_JSONOutput(t *testing.T) {
	issues := []map[string]any{
		makeIssue("ENG-5", "JSON search result", "Done", "High", ""),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		searchResponse(issues),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "search", "JSON"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 issue in JSON, got %d", len(decoded))
	}
	if decoded[0]["identifier"] != "ENG-5" {
		t.Errorf("identifier = %v, want ENG-5", decoded[0]["identifier"])
	}
}

func TestSearchCommand_EmptyResults(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		searchResponse([]map[string]any{}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "noresults"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error on empty results: %v", err)
	}
}

func TestSearchCommand_WithTeamFlag(t *testing.T) {
	issues := []map[string]any{
		makeIssue("ENG-10", "Team filtered result", "Todo", "Low", ""),
	}

	// team resolve response + search response
	server, bodies := newQueuedServer(t, []map[string]any{
		teamResolveResponse("team-uuid-123"),
		searchResponse(issues),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "filtered", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 2 {
		t.Fatalf("expected 2 requests (team resolve + search), got %d", len(*bodies))
	}
	// second request is the search, should include teamId and preserve term
	searchVars := (*bodies)[1]
	if searchVars["teamId"] != "team-uuid-123" {
		t.Errorf("teamId = %v, want team-uuid-123", searchVars["teamId"])
	}
	if searchVars["term"] != "filtered" {
		t.Errorf("term = %v, want filtered", searchVars["term"])
	}
}

func TestSearchCommand_LimitFlag(t *testing.T) {
	server, bodies := newQueuedServer(t, []map[string]any{
		searchResponse([]map[string]any{}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "test", "--limit", "10"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if (*bodies)[0]["first"] != float64(10) {
		t.Errorf("first = %v, want 10", (*bodies)[0]["first"])
	}
}

func TestSearchCommand_RequiresQuery(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no query provided")
	}
}

func TestSearchCommand_InvalidLimit(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	for _, limit := range []string{"0", "-1"} {
		t.Run("limit="+limit, func(t *testing.T) {
			var out bytes.Buffer
			root := cmd.NewRootCommand("test")
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs([]string{"search", "test", "--limit", limit})

			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error for --limit %s", limit)
			}
		})
	}
}
