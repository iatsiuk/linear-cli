package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func docUpdateResponse(doc map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"documentUpdate": map[string]any{
				"success":  true,
				"document": doc,
			},
		},
	}
}

func TestDocUpdateCommand_Title(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000042"
	doc := makeDoc(docID, "Updated Title", "", "", "")

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, docUpdateResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "update", docID, "--title", "Updated Title"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["id"] != docID {
		t.Errorf("id = %v, want %q", gotVars["id"], docID)
	}
	input, ok := gotVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("variables.input not set, got: %v", gotVars)
	}
	if input["title"] != "Updated Title" {
		t.Errorf("input.title = %v, want Updated Title", input["title"])
	}
	// content should not be present when not specified
	if _, present := input["content"]; present {
		t.Error("input.content should not be set when --content not provided")
	}
}

func TestDocUpdateCommand_Content(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000043"
	doc := makeDoc(docID, "Existing Title", "new content", "", "")

	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, docUpdateResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "update", docID, "--content", "new content"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input, ok := gotVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("variables.input not set, got: %v", gotVars)
	}
	if input["content"] != "new content" {
		t.Errorf("input.content = %v, want 'new content'", input["content"])
	}
	// title should not be set when not provided
	if _, present := input["title"]; present {
		t.Error("input.title should not be set when --title not provided")
	}
}

func TestDocUpdateCommand_MissingID(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "update"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when ID is missing")
	}
}

func TestDocUpdateCommand_NoFlags(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000044"

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "update", docID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no flags specified")
	}
	if !strings.Contains(err.Error(), "no fields to update") {
		t.Errorf("error should mention no fields, got: %v", err)
	}
}

func TestDocUpdateCommand_JSONOutput(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000045"
	doc := makeDoc(docID, "JSON Updated", "", "", "")

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, docUpdateResponse(doc))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "doc", "update", docID, "--title", "JSON Updated"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["title"] != "JSON Updated" {
		t.Errorf("expected title JSON Updated, got %v", decoded["title"])
	}
}

func TestDocUpdateCommand_SuccessFalse(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000046"

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{
			"data": map[string]any{
				"documentUpdate": map[string]any{
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
	root.SetArgs([]string{"doc", "update", docID, "--title", "Fail"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestDocUpdateCommand_ContentAndFileMutuallyExclusive(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000047"

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "update", docID, "--content", "inline", "--content-file", "/tmp/x"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both --content and --content-file are provided")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention mutually exclusive, got: %v", err)
	}
}

func TestDocUpdateCommand_HTTPError(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000048"

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "update", docID, "--title", "fail"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error on HTTP failure")
	}
}
