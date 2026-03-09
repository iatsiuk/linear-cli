package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func issueGetResponse(issue map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issue": issue,
		},
	}
}

func makeDetailedIssue() map[string]any {
	desc := "This is a detailed description of the issue."
	estimate := 3.0
	dueDate := "2026-04-01"
	return map[string]any{
		"id":            "id-ENG-42",
		"identifier":    "ENG-42",
		"title":         "Implement feature X",
		"description":   desc,
		"priority":      2.0,
		"priorityLabel": "Medium",
		"estimate":      estimate,
		"dueDate":       dueDate,
		"url":           "https://linear.app/issue/ENG-42",
		"createdAt":     "2026-01-10T00:00:00Z",
		"updatedAt":     "2026-02-15T00:00:00Z",
		"state": map[string]any{
			"id":    "state-2",
			"name":  "In Progress",
			"color": "#FF0000",
			"type":  "started",
		},
		"assignee": map[string]any{
			"id":          "user-1",
			"displayName": "Alice",
			"email":       "alice@example.com",
		},
		"team": map[string]any{
			"id":   "team-1",
			"name": "Engineering",
		},
		"labels": map[string]any{
			"nodes": []any{
				map[string]any{"id": "label-1", "name": "bug", "color": "#FF0000"},
			},
		},
	}
}

func TestIssueShowCommand_TableOutput(t *testing.T) {
	issue := makeDetailedIssue()
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueGetResponse(issue))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "show", "ENG-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()

	checks := []struct {
		label string
		want  string
	}{
		{"identifier", "ENG-42"},
		{"title", "Implement feature X"},
		{"status", "In Progress"},
		{"priority", "Medium"},
		{"assignee", "Alice"},
		{"team", "Engineering"},
		{"url", "https://linear.app/issue/ENG-42"},
		{"due date", "2026-04-01"},
		{"description", "This is a detailed description"},
		{"label", "bug"},
	}
	for _, c := range checks {
		if !strings.Contains(result, c.want) {
			t.Errorf("output should contain %s (%q), got:\n%s", c.label, c.want, result)
		}
	}
}

func TestIssueShowCommand_JSONOutput(t *testing.T) {
	issue := makeDetailedIssue()
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueGetResponse(issue))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "issue", "show", "ENG-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["identifier"] != "ENG-42" {
		t.Errorf("expected identifier ENG-42, got %v", decoded["identifier"])
	}
	if decoded["title"] != "Implement feature X" {
		t.Errorf("expected title 'Implement feature X', got %v", decoded["title"])
	}
}

func TestIssueShowCommand_PassesIdentifierToAPI(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, issueGetResponse(makeDetailedIssue()))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "show", "ENG-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["id"] != "ENG-42" {
		t.Errorf("expected id variable = ENG-42, got %v", gotVars["id"])
	}
}

func TestIssueShowCommand_MissingIdentifier(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueGetResponse(makeDetailedIssue()))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when identifier is missing")
	}
	if !strings.Contains(err.Error(), "identifier") {
		t.Errorf("error should mention identifier, got: %v", err)
	}
}

func TestIssueShowCommand_NotFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueGetResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "show", "NONE-999"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when issue is not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

func TestIssueShowCommand_NullableFields(t *testing.T) {
	// issue with no assignee, no description, no due date, no estimate
	issue := map[string]any{
		"id":            "id-ENG-1",
		"identifier":    "ENG-1",
		"title":         "Simple issue",
		"priority":      0.0,
		"priorityLabel": "No priority",
		"url":           "https://linear.app/issue/ENG-1",
		"createdAt":     "2026-01-01T00:00:00Z",
		"updatedAt":     "2026-01-01T00:00:00Z",
		"state": map[string]any{
			"id":    "state-1",
			"name":  "Todo",
			"color": "#000000",
			"type":  "unstarted",
		},
		"team": map[string]any{
			"id":   "team-1",
			"name": "Engineering",
		},
		"labels": map[string]any{"nodes": []any{}},
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueGetResponse(issue))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "show", "ENG-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "ENG-1") {
		t.Errorf("output should contain identifier ENG-1, got:\n%s", result)
	}
	if !strings.Contains(result, "Simple issue") {
		t.Errorf("output should contain title, got:\n%s", result)
	}
}
