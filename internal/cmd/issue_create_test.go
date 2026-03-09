package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

// newQueuedServer creates a test server that serves queued responses in order and
// records all request variable bodies.
func newQueuedServer(t *testing.T, responses []map[string]any) (*httptest.Server, *[]map[string]any) {
	t.Helper()
	bodies := &[]map[string]any{}
	var mu sync.Mutex
	idx := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		*bodies = append(*bodies, body.Variables)
		if idx >= len(responses) {
			t.Errorf("unexpected request %d (max %d)", idx+1, len(responses))
			mu.Unlock()
			http.Error(w, "too many requests", 500)
			return
		}
		resp := responses[idx]
		idx++
		mu.Unlock()
		writeJSONResponse(w, resp)
	}))
	t.Cleanup(server.Close)
	return server, bodies
}

func teamResolveResponse(teamID string) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"teams": map[string]any{
				"nodes": []map[string]any{{"id": teamID}},
			},
		},
	}
}

func issueCreateResponse(issue map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueCreate": map[string]any{
				"success": true,
				"issue":   issue,
			},
		},
	}
}

func TestIssueCreateCommand_Basic(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"
	created := makeIssue("ENG-10", "New feature", "Todo", "No priority", "")

	server, _ := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		issueCreateResponse(created),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "create", "--title", "New feature", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "ENG-10") {
		t.Errorf("output should contain ENG-10, got: %s", result)
	}
	if !strings.Contains(result, "New feature") {
		t.Errorf("output should contain title, got: %s", result)
	}
}

func TestIssueCreateCommand_JSONOutput(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"
	created := makeIssue("ENG-11", "JSON test", "Todo", "High", "Alice")

	server, _ := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		issueCreateResponse(created),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "issue", "create", "--title", "JSON test", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["identifier"] != "ENG-11" {
		t.Errorf("expected identifier ENG-11, got %v", decoded["identifier"])
	}
}

func TestIssueCreateCommand_PayloadSuccessFalse(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"

	server, _ := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		{
			"data": map[string]any{
				"issueCreate": map[string]any{
					"success": false,
					"issue":   nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "create", "--title", "Test", "--team", "ENG"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestIssueCreateCommand_MissingTitle(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "create", "--team", "ENG"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --title is missing")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention title, got: %v", err)
	}
}

func TestIssueCreateCommand_MissingTeam(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, issueListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "create", "--title", "Test"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --team is missing")
	}
	if !strings.Contains(err.Error(), "team") {
		t.Errorf("error should mention team, got: %v", err)
	}
}

func TestIssueCreateCommand_InvalidPriority(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"

	server, _ := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "create", "--title", "Test", "--team", "ENG", "--priority", "5"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for priority 5")
	}
	if !strings.Contains(err.Error(), "priority") {
		t.Errorf("error should mention priority, got: %v", err)
	}
}

func TestIssueCreateCommand_OptionalFieldsOmitted(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"
	created := makeIssue("ENG-12", "Minimal", "Todo", "No priority", "")

	server, bodies := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		issueCreateResponse(created),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "create", "--title", "Minimal", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// second request is the mutation; check its input
	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[1]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set in mutation vars: %v", mutationVars)
	}

	// only teamId and title should be present
	for _, key := range []string{"description", "assigneeId", "stateId", "priority", "labelIds", "dueDate", "estimate", "cycleId", "projectId", "parentId"} {
		if _, present := input[key]; present {
			t.Errorf("optional field %q should not be in input when not provided", key)
		}
	}
	if input["teamId"] != teamID {
		t.Errorf("teamId = %v, want %q", input["teamId"], teamID)
	}
	if input["title"] != "Minimal" {
		t.Errorf("title = %v, want Minimal", input["title"])
	}
}

func TestIssueCreateCommand_AssigneeMe(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"
	const viewerID = "viewer-uuid-1234-5678-90ab-cdef01234567"
	created := makeIssue("ENG-13", "My issue", "Todo", "No priority", "Me")

	server, bodies := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		// viewer query
		{"data": map[string]any{"viewer": map[string]any{"id": viewerID}}},
		issueCreateResponse(created),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "create", "--title", "My issue", "--team", "ENG", "--assignee", "me"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// third request is the mutation
	if len(*bodies) < 3 {
		t.Fatalf("expected 3 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[2]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set in mutation vars: %v", mutationVars)
	}
	if input["assigneeId"] != viewerID {
		t.Errorf("assigneeId = %v, want %q", input["assigneeId"], viewerID)
	}
}
