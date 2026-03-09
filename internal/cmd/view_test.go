package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

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
	desc := "Shows all open issues assigned to me"
	view := makeCustomView("cv-1", "My Open Issues", "Issue", false, &desc)

	server, bodies := newQueuedServer(t, []map[string]any{
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "show", "cv-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != "cv-1" {
		t.Errorf("id = %v, want cv-1", (*bodies)[0]["id"])
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
	view := makeCustomView("cv-2", "Team Board", "Issue", true, nil)

	server, _ := newQueuedServer(t, []map[string]any{
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"view", "show", "cv-2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "yes") {
		t.Errorf("output should show shared=yes, got: %s", result)
	}
}

func TestViewShowCommand_JSONOutput(t *testing.T) {
	view := makeCustomView("cv-3", "Backlog", "Issue", false, nil)

	server, _ := newQueuedServer(t, []map[string]any{
		customViewShowResponse(view),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "view", "show", "cv-3"})

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
