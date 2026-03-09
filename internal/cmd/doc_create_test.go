package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func docCreateResponse(doc map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"documentCreate": map[string]any{
				"success":  true,
				"document": doc,
			},
		},
	}
}

func TestDocCreateCommand_Basic(t *testing.T) {
	doc := makeDoc("doc-new", "My New Doc", "", "", "Alice")

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, docCreateResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "create", "--title", "My New Doc"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "My New Doc") {
		t.Errorf("output should contain document title, got: %s", result)
	}
}

func TestDocCreateCommand_MissingTitle(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "create"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --title is missing")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention title, got: %v", err)
	}
}

func TestDocCreateCommand_SendsMutationInput(t *testing.T) {
	doc := makeDoc("doc-inp", "Test Input", "some content", "", "")

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, docCreateResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "create", "--title", "Test Input", "--content", "some content"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input, ok := gotVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("variables.input not set, got: %v", gotVars)
	}
	if input["title"] != "Test Input" {
		t.Errorf("input.title = %v, want Test Input", input["title"])
	}
	if input["content"] != "some content" {
		t.Errorf("input.content = %v, want 'some content'", input["content"])
	}
}

func TestDocCreateCommand_ContentFile(t *testing.T) {
	// write a temp file
	dir := t.TempDir()
	fpath := filepath.Join(dir, "content.md")
	if err := os.WriteFile(fpath, []byte("# From file\n"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	doc := makeDoc("doc-cf", "From File Doc", "# From file\n", "", "")

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, docCreateResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "create", "--title", "From File Doc", "--content-file", fpath})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input, ok := gotVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("variables.input not set, got: %v", gotVars)
	}
	if input["content"] != "# From file\n" {
		t.Errorf("input.content = %v, want '# From file\\n'", input["content"])
	}
}

func TestDocCreateCommand_ContentAndFileMutuallyExclusive(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "create", "--title", "T", "--content", "inline", "--content-file", "/tmp/x"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both --content and --content-file are provided")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention mutually exclusive, got: %v", err)
	}
}

func TestDocCreateCommand_JSONOutput(t *testing.T) {
	doc := makeDoc("doc-json", "JSON Doc", "", "", "")

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, docCreateResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "doc", "create", "--title", "JSON Doc"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["title"] != "JSON Doc" {
		t.Errorf("expected title JSON Doc, got %v", decoded["title"])
	}
}

func TestDocCreateCommand_SuccessFalse(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{
			"data": map[string]any{
				"documentCreate": map[string]any{
					"success":  false,
					"document": nil,
				},
			},
		})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "create", "--title", "Fail"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}
