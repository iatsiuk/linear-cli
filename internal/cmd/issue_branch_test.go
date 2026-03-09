package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func branchSearchResponse(issue map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueVcsBranchSearch": issue,
		},
	}
}

func TestIssueBranchCommand_TableOutput(t *testing.T) {
	issue := makeDetailedIssue()
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, branchSearchResponse(issue))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "branch", "feature/eng-42-implement-x"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	checks := []string{"ENG-42", "Implement feature X", "In Progress"}
	for _, want := range checks {
		if !strings.Contains(result, want) {
			t.Errorf("output should contain %q, got:\n%s", want, result)
		}
	}
}

func TestIssueBranchCommand_SendsBranchName(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, branchSearchResponse(makeDetailedIssue()))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "branch", "feature/eng-42-fix"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["branchName"] != "feature/eng-42-fix" {
		t.Errorf("expected branchName = 'feature/eng-42-fix', got %v", gotVars["branchName"])
	}
}

func TestIssueBranchCommand_JSONOutput(t *testing.T) {
	issue := makeDetailedIssue()
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, branchSearchResponse(issue))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "issue", "branch", "feature/eng-42"})

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
}

func TestIssueBranchCommand_AutoDetect(t *testing.T) {
	var gotBranch string
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if v, ok := body.Variables["branchName"].(string); ok {
			gotBranch = v
		}
		writeJSONResponse(w, branchSearchResponse(makeDetailedIssue()))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "branch"})

	if err := root.Execute(); err != nil {
		t.Skipf("git not available or not in repo: %v", err)
	}

	if gotBranch == "" || gotBranch == "HEAD" {
		t.Errorf("expected non-empty branch name from auto-detect, got %q", gotBranch)
	}
}

func TestIssueBranchCommand_NotFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, branchSearchResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "branch", "unknown-branch"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when issue is not found")
	}
	if !strings.Contains(err.Error(), "no issue found") {
		t.Errorf("error should mention 'no issue found', got: %v", err)
	}
}
