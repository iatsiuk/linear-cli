package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func makeViewerResponse(name, email string, admin, guest, active bool, teams []map[string]any) map[string]any {
	viewer := map[string]any{
		"id":          "user-me",
		"displayName": name,
		"email":       email,
		"active":      active,
		"admin":       admin,
		"guest":       guest,
		"isMe":        true,
		"createdAt":   "2026-01-01T00:00:00Z",
		"updatedAt":   "2026-01-02T00:00:00Z",
		"teams":       map[string]any{"nodes": teams},
	}
	return map[string]any{"data": map[string]any{"viewer": viewer}}
}

func makeViewerIssuesResponse(field string, issues []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"viewer": map[string]any{
				field: map[string]any{"nodes": issues},
			},
		},
	}
}

func makeShortIssue(identifier, title, stateName string) map[string]any {
	return map[string]any{
		"id":         "id-" + identifier,
		"identifier": identifier,
		"title":      title,
		"state": map[string]any{
			"id":    "state-1",
			"name":  stateName,
			"color": "#000000",
			"type":  "started",
		},
		"team": map[string]any{
			"id":   "team-1",
			"name": "Engineering",
			"key":  "ENG",
		},
	}
}

func TestMeCommand_TableOutput(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		teams := []map[string]any{
			{"id": "t1", "name": "Engineering", "key": "ENG"},
		}
		writeJSONResponse(w, makeViewerResponse("Alice Smith", "alice@example.com", false, false, true, teams))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"me"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Alice Smith") {
		t.Errorf("output should contain name, got:\n%s", result)
	}
	if !strings.Contains(result, "alice@example.com") {
		t.Errorf("output should contain email, got:\n%s", result)
	}
	if !strings.Contains(result, "Member") {
		t.Errorf("output should contain role=Member, got:\n%s", result)
	}
	if !strings.Contains(result, "Active:") || !strings.Contains(result, "yes") {
		t.Errorf("output should contain Active: yes, got:\n%s", result)
	}
	if !strings.Contains(result, "Engineering") {
		t.Errorf("output should contain team name, got:\n%s", result)
	}
	if !strings.Contains(result, "ENG") {
		t.Errorf("output should contain team key, got:\n%s", result)
	}
}

func TestMeCommand_AdminRole(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, makeViewerResponse("Bob Admin", "bob@example.com", true, false, true, nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"me"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Admin") {
		t.Errorf("output should contain role=Admin, got:\n%s", out.String())
	}
}

func TestMeCommand_GuestRole(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, makeViewerResponse("Carol Guest", "carol@example.com", false, true, true, nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"me"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(out.String(), "Guest") {
		t.Errorf("output should contain role=Guest, got:\n%s", out.String())
	}
}

func TestMeCommand_InactiveUser(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, makeViewerResponse("Dan Inactive", "dan@example.com", false, false, false, nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"me"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Active:") || !strings.Contains(result, "no") {
		t.Errorf("output should contain Active: no, got:\n%s", result)
	}
}

func TestMeCommand_JSONOutput(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, makeViewerResponse("Alice Smith", "alice@example.com", false, false, true, nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "me"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["displayName"] != "Alice Smith" {
		t.Errorf("expected displayName Alice Smith, got %v", decoded["displayName"])
	}
	if decoded["email"] != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %v", decoded["email"])
	}
}

func TestMeCommand_AssignedFlag(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)

		issues := []map[string]any{
			makeShortIssue("ENG-5", "Assigned issue", "In Progress"),
		}
		writeJSONResponse(w, makeViewerIssuesResponse("assignedIssues", issues))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"me", "--assigned"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "ENG-5") {
		t.Errorf("output should contain ENG-5, got:\n%s", result)
	}
	if !strings.Contains(result, "Assigned issue") {
		t.Errorf("output should contain issue title, got:\n%s", result)
	}
	if !strings.Contains(result, "In Progress") {
		t.Errorf("output should contain state name, got:\n%s", result)
	}
}

func TestMeCommand_CreatedFlag(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		issues := []map[string]any{
			makeShortIssue("ENG-7", "Created issue", "Backlog"),
		}
		writeJSONResponse(w, makeViewerIssuesResponse("createdIssues", issues))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"me", "--created"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "ENG-7") {
		t.Errorf("output should contain ENG-7, got:\n%s", result)
	}
	if !strings.Contains(result, "Created issue") {
		t.Errorf("output should contain issue title, got:\n%s", result)
	}
}

func TestMeCommand_BothFlagsError(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"me", "--assigned", "--created"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both --assigned and --created are provided")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected mutually exclusive error, got: %v", err)
	}
}

func TestMeCommand_AssignedFlag_JSONOutput(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		issues := []map[string]any{
			makeShortIssue("ENG-5", "Assigned issue", "In Progress"),
		}
		writeJSONResponse(w, makeViewerIssuesResponse("assignedIssues", issues))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "me", "--assigned"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(decoded))
	}
	if decoded[0]["identifier"] != "ENG-5" {
		t.Errorf("expected identifier ENG-5, got %v", decoded[0]["identifier"])
	}
}
