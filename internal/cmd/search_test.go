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

func projectSearchResponse(projects []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"searchProjects": map[string]any{
				"nodes": projects,
			},
		},
	}
}

func documentSearchResponse(docs []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"searchDocuments": map[string]any{
				"nodes": docs,
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

func TestSearchCommand_TypeProject_TableOutput(t *testing.T) {
	projects := []map[string]any{
		makeProject("proj-1", "API Redesign", "started", "onTrack", 0.5, "2026-06-01"),
		makeProject("proj-2", "Mobile App", "planned", "", 0, ""),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		projectSearchResponse(projects),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "API", "--type", "project"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["term"] != "API" {
		t.Errorf("term = %v, want API", (*bodies)[0]["term"])
	}

	result := out.String()
	if !strings.Contains(result, "API Redesign") {
		t.Errorf("output should contain project name, got: %s", result)
	}
	if !strings.Contains(result, "Mobile App") {
		t.Errorf("output should contain second project name, got: %s", result)
	}
}

func TestSearchCommand_TypeProject_JSONOutput(t *testing.T) {
	projects := []map[string]any{
		makeProject("proj-1", "API Redesign", "started", "onTrack", 0.5, "2026-06-01"),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		projectSearchResponse(projects),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "search", "API", "--type", "project"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 project in JSON, got %d", len(decoded))
	}
	if decoded[0]["name"] != "API Redesign" {
		t.Errorf("name = %v, want API Redesign", decoded[0]["name"])
	}
}

func TestSearchCommand_TypeProject_Empty(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		projectSearchResponse([]map[string]any{}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "noresults", "--type", "project"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error on empty results: %v", err)
	}
}

func TestSearchCommand_TypeDocument_TableOutput(t *testing.T) {
	docs := []map[string]any{
		makeDoc("doc-1", "Architecture Overview", "", "Backend API", "Alice"),
		makeDoc("doc-2", "Onboarding Guide", "", "", ""),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		documentSearchResponse(docs),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "arch", "--type", "document"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["term"] != "arch" {
		t.Errorf("term = %v, want arch", (*bodies)[0]["term"])
	}

	result := out.String()
	if !strings.Contains(result, "Architecture Overview") {
		t.Errorf("output should contain doc title, got: %s", result)
	}
	if !strings.Contains(result, "Onboarding Guide") {
		t.Errorf("output should contain second doc title, got: %s", result)
	}
	if !strings.Contains(result, "Backend API") {
		t.Errorf("output should contain project name, got: %s", result)
	}
}

func TestSearchCommand_TypeDocument_JSONOutput(t *testing.T) {
	docs := []map[string]any{
		makeDoc("doc-1", "Architecture Overview", "content here", "Backend API", "Alice"),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		documentSearchResponse(docs),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "search", "arch", "--type", "document"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 document in JSON, got %d", len(decoded))
	}
	if decoded[0]["title"] != "Architecture Overview" {
		t.Errorf("title = %v, want 'Architecture Overview'", decoded[0]["title"])
	}
}

func TestSearchCommand_TypeDocument_Empty(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		documentSearchResponse([]map[string]any{}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "noresults", "--type", "document"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error on empty results: %v", err)
	}
}

func TestSearchCommand_TypeIssueExplicit(t *testing.T) {
	issues := []map[string]any{
		makeIssue("ENG-1", "Fix login bug", "In Progress", "Urgent", "Alice"),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		searchResponse(issues),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "login", "--type", "issue"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if (*bodies)[0]["term"] != "login" {
		t.Errorf("term = %v, want login", (*bodies)[0]["term"])
	}

	result := out.String()
	if !strings.Contains(result, "ENG-1") {
		t.Errorf("output should contain ENG-1, got: %s", result)
	}
}

func TestSearchCommand_InvalidType(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"search", "query", "--type", "invalid"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for invalid --type")
	}
	if !strings.Contains(err.Error(), "--type must be one of") {
		t.Errorf("error should mention valid types, got: %v", err)
	}
}
