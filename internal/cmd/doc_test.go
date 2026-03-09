package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func makeDoc(id, title, content, projectName, creatorName string) map[string]any {
	d := map[string]any{
		"id":        id,
		"title":     title,
		"slugId":    "slug-" + id,
		"url":       "https://linear.app/doc/" + id,
		"createdAt": "2026-01-01T00:00:00Z",
		"updatedAt": "2026-02-01T00:00:00Z",
	}
	if content != "" {
		d["content"] = content
	}
	if projectName != "" {
		d["project"] = map[string]any{"id": "proj-1", "name": projectName}
	}
	if creatorName != "" {
		d["creator"] = map[string]any{"id": "user-1", "displayName": creatorName, "email": "user@example.com"}
	}
	return d
}

func docListResponse(docs []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"documents": map[string]any{
				"nodes":    docs,
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			},
		},
	}
}

func docGetResponse(doc map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"document": doc,
		},
	}
}

func TestDocListCommand_TableOutput(t *testing.T) {
	docs := []map[string]any{
		makeDoc("doc-1", "Design Doc", "", "Auth Project", "Alice"),
		makeDoc("doc-2", "API Spec", "", "", ""),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, docListResponse(docs))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "TITLE") {
		t.Errorf("output should contain TITLE header, got:\n%s", result)
	}
	if !strings.Contains(result, "PROJECT") {
		t.Errorf("output should contain PROJECT header, got:\n%s", result)
	}
	if !strings.Contains(result, "CREATOR") {
		t.Errorf("output should contain CREATOR header, got:\n%s", result)
	}
	if !strings.Contains(result, "UPDATED") {
		t.Errorf("output should contain UPDATED header, got:\n%s", result)
	}
	if !strings.Contains(result, "Design Doc") {
		t.Errorf("output should contain document title, got:\n%s", result)
	}
	if !strings.Contains(result, "Auth Project") {
		t.Errorf("output should contain project name, got:\n%s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("output should contain creator name, got:\n%s", result)
	}
}

func TestDocListCommand_JSONOutput(t *testing.T) {

	docs := []map[string]any{
		makeDoc("doc-1", "Design Doc", "some content", "Auth Project", "Alice"),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, docListResponse(docs))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "doc", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 document, got %d", len(decoded))
	}
	if decoded[0]["title"] != "Design Doc" {
		t.Errorf("expected title Design Doc, got %v", decoded[0]["title"])
	}
}

func TestDocListCommand_ProjectFilter(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, docListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "list", "--project", "proj-abc"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("variables.filter not set, got: %v", gotVars["filter"])
	}
	proj, ok := filter["project"].(map[string]any)
	if !ok {
		t.Fatalf("filter.project not set, got: %v", filter["project"])
	}
	idFilter, ok := proj["id"].(map[string]any)
	if !ok {
		t.Fatalf("filter.project.id not set, got: %v", proj["id"])
	}
	if idFilter["eq"] != "proj-abc" {
		t.Errorf("filter.project.id.eq = %v, want proj-abc", idFilter["eq"])
	}
}

func TestDocListCommand_LimitFlag(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, docListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "list", "--limit", "10"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	first, ok := gotVars["first"].(float64)
	if !ok {
		t.Fatalf("variables.first not set, got: %v (%T)", gotVars["first"], gotVars["first"])
	}
	if int(first) != 10 {
		t.Errorf("variables.first = %v, want 10", first)
	}
}

func TestDocListCommand_IncludeArchived(t *testing.T) {

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, docListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "list", "--include-archived"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["includeArchived"] != true {
		t.Errorf("variables.includeArchived = %v, want true", gotVars["includeArchived"])
	}
}

func TestDocShowCommand_TableOutput(t *testing.T) {

	const docID = "00000000-0000-0000-0000-000000000001"
	content := "# Hello\n\nThis is the content."
	doc := makeDoc(docID, "Design Doc", content, "Auth Project", "Alice")

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, docGetResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "show", docID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Design Doc") {
		t.Errorf("output should contain title, got:\n%s", result)
	}
	if !strings.Contains(result, "Auth Project") {
		t.Errorf("output should contain project, got:\n%s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("output should contain creator, got:\n%s", result)
	}
	if !strings.Contains(result, content) {
		t.Errorf("output should contain content, got:\n%s", result)
	}
}

func TestDocShowCommand_JSONOutput(t *testing.T) {

	const docID = "00000000-0000-0000-0000-000000000001"
	doc := makeDoc(docID, "Design Doc", "content here", "Auth Project", "Alice")

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, docGetResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "doc", "show", docID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["title"] != "Design Doc" {
		t.Errorf("expected title Design Doc, got %v", decoded["title"])
	}
}

func TestDocShowCommand_NotFound(t *testing.T) {

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{"data": map[string]any{"document": nil}})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "show", "nonexistent"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for not found document")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestDocShowCommand_MissingID(t *testing.T) {

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when ID is missing")
	}
}
