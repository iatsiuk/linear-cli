package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
	"linear-cli/internal/model"
)

func makeTemplate(id, name, tmplType string, description *string) map[string]any {
	m := map[string]any{
		"id":   id,
		"name": name,
		"type": tmplType,
	}
	if description != nil {
		m["description"] = *description
	}
	return m
}

func templateListResponse(templates []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"templates": templates,
		},
	}
}

func templateShowResponse(template map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"template": template,
		},
	}
}

func TestTemplateListCommand_Basic(t *testing.T) {
	templates := []map[string]any{
		makeTemplate("t-1", "Bug Report", "issue", nil),
		makeTemplate("t-2", "Feature Request", "issue", strPtr("A feature request template")),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		templateListResponse(templates),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"template", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Bug Report") {
		t.Errorf("output should contain template name, got: %s", result)
	}
	if !strings.Contains(result, "Feature Request") {
		t.Errorf("output should contain second template name, got: %s", result)
	}
	if !strings.Contains(result, "issue") {
		t.Errorf("output should contain template type, got: %s", result)
	}
}

func TestTemplateListCommand_Empty(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		templateListResponse([]map[string]any{}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"template", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTemplateListCommand_JSONOutput(t *testing.T) {
	templates := []map[string]any{
		makeTemplate("t-1", "Bug Report", "issue", nil),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		templateListResponse(templates),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "template", "list"})

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
	if decoded[0]["name"] != "Bug Report" {
		t.Errorf("expected name 'Bug Report', got %v", decoded[0]["name"])
	}
}

func TestTemplateShowCommand_Basic(t *testing.T) {
	desc := "Template for bug reports"
	tmpl := makeTemplate("t-1", "Bug Report", "issue", &desc)
	tmpl["templateData"] = map[string]any{"title": "Bug: ", "description": "Steps to reproduce:"}

	server, bodies := newQueuedServer(t, []map[string]any{
		templateShowResponse(tmpl),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"template", "show", "t-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != "t-1" {
		t.Errorf("id = %v, want t-1", (*bodies)[0]["id"])
	}

	result := out.String()
	if !strings.Contains(result, "Bug Report") {
		t.Errorf("output should contain name, got: %s", result)
	}
	if !strings.Contains(result, "issue") {
		t.Errorf("output should contain type, got: %s", result)
	}
	if !strings.Contains(result, desc) {
		t.Errorf("output should contain description, got: %s", result)
	}
}

func TestTemplateShowCommand_JSONOutput(t *testing.T) {
	tmpl := makeTemplate("t-2", "Feature Request", "issue", nil)

	server, _ := newQueuedServer(t, []map[string]any{
		templateShowResponse(tmpl),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "template", "show", "t-2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Feature Request" {
		t.Errorf("expected name 'Feature Request', got %v", decoded["name"])
	}
	if decoded["type"] != "issue" {
		t.Errorf("expected type 'issue', got %v", decoded["type"])
	}
}

func TestTemplateShowCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"template", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when template id is missing")
	}
}

func TestTemplateModelDeserialization(t *testing.T) {
	t.Parallel()
	input := `{
		"id": "tmpl-1",
		"name": "Bug Fix",
		"type": "issue",
		"description": "For fixing bugs",
		"templateData": {"key": "value"}
	}`

	var tmpl model.Template
	if err := json.Unmarshal([]byte(input), &tmpl); err != nil {
		t.Fatalf("deserialization failed: %v", err)
	}
	if tmpl.ID != "tmpl-1" {
		t.Errorf("ID = %s, want tmpl-1", tmpl.ID)
	}
	if tmpl.Name != "Bug Fix" {
		t.Errorf("Name = %s, want Bug Fix", tmpl.Name)
	}
	if tmpl.Type != "issue" {
		t.Errorf("Type = %s, want issue", tmpl.Type)
	}
	if tmpl.Description == nil || *tmpl.Description != "For fixing bugs" {
		t.Errorf("Description = %v, want 'For fixing bugs'", tmpl.Description)
	}
	if len(tmpl.TemplateData) == 0 {
		t.Error("TemplateData should not be empty")
	}
}
