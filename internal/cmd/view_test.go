package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/iatsiuk/linear-cli/internal/cmd"
	"github.com/iatsiuk/linear-cli/internal/model"
)

func makeCustomView(id, name, modelName string, shared bool, description *string) map[string]any {
	m := map[string]any{
		"id":        id,
		"name":      name,
		"modelName": modelName,
		"shared":    shared,
	}
	if description != nil {
		m["description"] = *description
	}
	return m
}

func customViewListResponse(views []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"customViews": map[string]any{
				"nodes": views,
			},
		},
	}
}

func customViewShowResponse(view map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"customView": view,
		},
	}
}

func TestViewListCommand_Basic(t *testing.T) {
	views := []map[string]any{
		makeCustomView("cv-1", "My Issues", "Issue", false, nil),
		makeCustomView("cv-2", "Team Projects", "Project", true, strPtr("Shared with team")),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		customViewListResponse(views),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "My Issues") {
		t.Errorf("output should contain view name, got: %s", result)
	}
	if !strings.Contains(result, "Team Projects") {
		t.Errorf("output should contain second view name, got: %s", result)
	}
	if !strings.Contains(result, "Issue") {
		t.Errorf("output should contain type, got: %s", result)
	}
}

func TestViewListCommand_Empty(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		customViewListResponse([]map[string]any{}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := out.String(); got != "(no results)\n" {
		t.Errorf("expected %q, got %q", "(no results)\n", got)
	}
}

func TestViewListCommand_JSONOutput(t *testing.T) {
	views := []map[string]any{
		makeCustomView("cv-1", "My Issues", "Issue", false, nil),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		customViewListResponse(views),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "view", "list"})

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
	if decoded[0]["name"] != "My Issues" {
		t.Errorf("expected name 'My Issues', got %v", decoded[0]["name"])
	}
}

func TestViewShowCommand_Basic(t *testing.T) {
	const viewID = "11111111-2222-3333-4444-555555555555"
	desc := "Shows all open issues assigned to me"
	view := makeCustomView(viewID, "My Open Issues", "Issue", false, &desc)

	server, bodies := newQueuedServer(t, []map[string]any{
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "show", viewID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != viewID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], viewID)
	}

	result := out.String()
	if !strings.Contains(result, "My Open Issues") {
		t.Errorf("output should contain name, got: %s", result)
	}
	if !strings.Contains(result, "Issue") {
		t.Errorf("output should contain type, got: %s", result)
	}
	if !strings.Contains(result, desc) {
		t.Errorf("output should contain description, got: %s", result)
	}
}

func TestViewShowCommand_Shared(t *testing.T) {
	const viewID = "22222222-3333-4444-5555-666666666666"
	view := makeCustomView(viewID, "Team Board", "Issue", true, nil)

	server, _ := newQueuedServer(t, []map[string]any{
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "show", viewID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "yes") {
		t.Errorf("output should show shared=yes, got: %s", result)
	}
}

func TestViewShowCommand_JSONOutput(t *testing.T) {
	const viewID = "33333333-4444-5555-6666-777777777777"
	view := makeCustomView(viewID, "Backlog", "Issue", false, nil)

	server, _ := newQueuedServer(t, []map[string]any{
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "view", "show", viewID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Backlog" {
		t.Errorf("expected name 'Backlog', got %v", decoded["name"])
	}
}

func TestViewShowCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when view id is missing")
	}
}

func viewIssuesResponse(issues []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"customView": map[string]any{
				"issues": map[string]any{
					"nodes":    issues,
					"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
				},
			},
		},
	}
}

func TestViewIssuesCommand_TableOutput(t *testing.T) {
	const viewID = "44444444-5555-6666-7777-888888888888"
	issues := []map[string]any{
		makeIssue("ENG-10", "View issue one", "In Progress", "Medium", "Alice"),
		makeIssue("ENG-11", "View issue two", "Backlog", "Low", ""),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		viewIssuesResponse(issues),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "issues", viewID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != viewID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], viewID)
	}

	result := out.String()
	if !strings.Contains(result, "ENG-10") {
		t.Errorf("output should contain ENG-10, got: %s", result)
	}
	if !strings.Contains(result, "ENG-11") {
		t.Errorf("output should contain ENG-11, got: %s", result)
	}
	if !strings.Contains(result, "View issue one") {
		t.Errorf("output should contain issue title, got: %s", result)
	}
}

func TestViewIssuesCommand_JSONOutput(t *testing.T) {
	const viewID = "55555555-6666-7777-8888-999999999999"
	issues := []map[string]any{
		makeIssue("ENG-12", "JSON view issue", "Todo", "High", "Bob"),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		viewIssuesResponse(issues),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "view", "issues", viewID})

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
	if decoded[0]["identifier"] != "ENG-12" {
		t.Errorf("identifier = %v, want ENG-12", decoded[0]["identifier"])
	}
}

func TestViewIssuesCommand_WithLimit(t *testing.T) {
	const viewID = "66666666-7777-8888-9999-aaaaaaaaaaaa"
	server, bodies := newQueuedServer(t, []map[string]any{
		viewIssuesResponse([]map[string]any{}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "issues", viewID, "--limit", "5"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	first, ok := (*bodies)[0]["first"]
	if !ok {
		t.Fatalf("expected 'first' variable in request body")
	}
	// JSON numbers are float64 when decoded into any
	if first.(float64) != 5 {
		t.Errorf("first = %v, want 5", first)
	}
}

func TestViewIssuesCommand_MissingArg(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "issues"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when view id is missing")
	}
}

func TestViewShowCommand_HelpMentionsAcceptedForms(t *testing.T) {
	root := cmd.NewRootCommand("test")
	var showCmd *cobra.Command
	for _, sub := range root.Commands() {
		if sub.Use == "view" {
			for _, s := range sub.Commands() {
				if strings.HasPrefix(s.Use, "show") {
					showCmd = s
					break
				}
			}
			break
		}
	}
	if showCmd == nil {
		t.Fatal("view show command not found")
	}
	short := strings.ToLower(showCmd.Short)
	for _, want := range []string{"name", "uuid", "slug"} {
		if !strings.Contains(short, want) {
			t.Errorf("Short = %q, want to mention %q", showCmd.Short, want)
		}
	}
}

func customViewResolveResponse(viewID string) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"customViews": map[string]any{
				"nodes": []map[string]any{{"id": viewID}},
			},
		},
	}
}

func TestViewShow_ByName(t *testing.T) {
	const viewID = "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	desc := "Issues without estimates"
	view := makeCustomView(viewID, "Without Estimates", "Issue", false, &desc)

	server, bodies := newQueuedServer(t, []map[string]any{
		customViewResolveResponse(viewID),
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "show", "Without Estimates"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 2 {
		t.Fatalf("expected 2 requests (resolve + show), got %d", len(*bodies))
	}
	if (*bodies)[0]["name"] != "Without Estimates" {
		t.Errorf("resolve name = %v, want 'Without Estimates'", (*bodies)[0]["name"])
	}
	if (*bodies)[1]["id"] != viewID {
		t.Errorf("show id = %v, want %s", (*bodies)[1]["id"], viewID)
	}

	if !strings.Contains(out.String(), "Without Estimates") {
		t.Errorf("output should contain view name, got: %s", out.String())
	}
}

func TestViewShow_ByUUID(t *testing.T) {
	const viewID = "bbbbbbbb-cccc-dddd-eeee-ffffffffffff"
	view := makeCustomView(viewID, "My View", "Issue", false, nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "show", viewID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 1 {
		t.Fatalf("expected 1 request (no resolve call), got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != viewID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], viewID)
	}
}

func TestViewIssues_ByName(t *testing.T) {
	const viewID = "cccccccc-dddd-eeee-ffff-000000000000"
	issues := []map[string]any{
		makeIssue("ENG-100", "An issue", "Todo", "Medium", ""),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		customViewResolveResponse(viewID),
		viewIssuesResponse(issues),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "issues", "Without Estimates"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 2 {
		t.Fatalf("expected 2 requests (resolve + issues), got %d", len(*bodies))
	}
	if (*bodies)[0]["name"] != "Without Estimates" {
		t.Errorf("resolve name = %v, want 'Without Estimates'", (*bodies)[0]["name"])
	}
	if (*bodies)[1]["id"] != viewID {
		t.Errorf("issues id = %v, want %s", (*bodies)[1]["id"], viewID)
	}
}

func TestViewIssues_NotFound(t *testing.T) {
	// name lookup returns empty, slug fallback passes "Nonexistent" to API which returns null
	server, _ := newQueuedServer(t, []map[string]any{
		customViewListResponse([]map[string]any{}),
		{"data": map[string]any{"customView": nil}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "issues", "Nonexistent"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestViewShow_BySlug(t *testing.T) {
	const slug = "my-team-bugs"
	const viewID = "dddddddd-eeee-ffff-0000-111111111111"
	view := makeCustomView(viewID, "My Team Bugs", "Issue", false, nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		customViewListResponse([]map[string]any{}),
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "show", slug})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 2 {
		t.Fatalf("expected 2 requests (resolve + show), got %d", len(*bodies))
	}
	if (*bodies)[0]["name"] != slug {
		t.Errorf("resolve name = %v, want %q", (*bodies)[0]["name"], slug)
	}
	if (*bodies)[1]["id"] != slug {
		t.Errorf("show id = %v, want %q (slug passthrough)", (*bodies)[1]["id"], slug)
	}
}

func TestViewIssues_BySlug(t *testing.T) {
	const slug = "without-estimates"
	issues := []map[string]any{
		makeIssue("ENG-200", "A slug issue", "Todo", "Medium", ""),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		customViewListResponse([]map[string]any{}),
		viewIssuesResponse(issues),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "issues", slug})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 2 {
		t.Fatalf("expected 2 requests (resolve + issues), got %d", len(*bodies))
	}
	if (*bodies)[0]["name"] != slug {
		t.Errorf("resolve name = %v, want %q", (*bodies)[0]["name"], slug)
	}
	if (*bodies)[1]["id"] != slug {
		t.Errorf("issues id = %v, want %q (slug passthrough)", (*bodies)[1]["id"], slug)
	}
}

func TestCustomViewDeserialization(t *testing.T) {
	data := `{"id":"cv-1","name":"Test View","modelName":"Issue","shared":true,"description":"A test view"}`

	var v model.CustomView
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if v.ID != "cv-1" {
		t.Errorf("ID = %v, want cv-1", v.ID)
	}
	if v.Name != "Test View" {
		t.Errorf("Name = %v, want 'Test View'", v.Name)
	}
	if v.ModelName != "Issue" {
		t.Errorf("ModelName = %v, want 'Issue'", v.ModelName)
	}
	if !v.Shared {
		t.Errorf("Shared = %v, want true", v.Shared)
	}
	if v.Description == nil || *v.Description != "A test view" {
		t.Errorf("Description = %v, want 'A test view'", v.Description)
	}
}
