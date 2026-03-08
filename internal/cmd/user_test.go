package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func makeUser(id, name, email string, active, admin, guest bool) map[string]any {
	return map[string]any{
		"id":          id,
		"email":       email,
		"displayName": name,
		"avatarUrl":   nil,
		"active":      active,
		"admin":       admin,
		"guest":       guest,
		"isMe":        false,
		"createdAt":   "2026-01-01T00:00:00Z",
		"updatedAt":   "2026-01-02T00:00:00Z",
	}
}

func userListResponse(users []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"users": map[string]any{
				"nodes":    users,
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			},
		},
	}
}

func userGetResponse(user map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"user": user,
		},
	}
}

func TestUserListCommand_TableOutput(t *testing.T) {
	users := []map[string]any{
		makeUser("u1", "Alice Smith", "alice@example.com", true, false, false),
		makeUser("u2", "Bob Admin", "bob@example.com", true, true, false),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, userListResponse(users))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"user", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Alice Smith") {
		t.Errorf("output should contain Alice Smith, got:\n%s", result)
	}
	if !strings.Contains(result, "alice@example.com") {
		t.Errorf("output should contain alice@example.com, got:\n%s", result)
	}
	if !strings.Contains(result, "Bob Admin") {
		t.Errorf("output should contain Bob Admin, got:\n%s", result)
	}
}

func TestUserListCommand_TableHeaders(t *testing.T) {
	users := []map[string]any{
		makeUser("u1", "Alice Smith", "alice@example.com", true, false, false),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, userListResponse(users))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"user", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, col := range []string{"NAME", "EMAIL", "ROLE", "ACTIVE"} {
		if !strings.Contains(result, col) {
			t.Errorf("output should contain %s column header, got:\n%s", col, result)
		}
	}
}

func TestUserListCommand_RoleDisplay(t *testing.T) {
	users := []map[string]any{
		makeUser("u1", "Alice Member", "alice@example.com", true, false, false),
		makeUser("u2", "Bob Admin", "bob@example.com", true, true, false),
		makeUser("u3", "Carol Guest", "carol@example.com", true, false, true),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, userListResponse(users))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"user", "list", "--include-guests"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Member") {
		t.Errorf("output should contain Member role, got:\n%s", result)
	}
	if !strings.Contains(result, "Admin") {
		t.Errorf("output should contain Admin role, got:\n%s", result)
	}
	if !strings.Contains(result, "Guest") {
		t.Errorf("output should contain Guest role, got:\n%s", result)
	}
}

func TestUserListCommand_GuestFilteredByDefault(t *testing.T) {
	users := []map[string]any{
		makeUser("u1", "Alice Member", "alice@example.com", true, false, false),
		makeUser("u2", "Bob Guest", "bob@example.com", true, false, true),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, userListResponse(users))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"user", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Alice Member") {
		t.Errorf("output should contain Alice Member, got:\n%s", result)
	}
	if strings.Contains(result, "Bob Guest") {
		t.Errorf("output should not contain guest user Bob Guest, got:\n%s", result)
	}
}

func TestUserListCommand_IncludeDisabled(t *testing.T) {
	var capturedVars map[string]any

	users := []map[string]any{
		makeUser("u1", "Alice Smith", "alice@example.com", true, false, false),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		capturedVars = body.Variables
		writeJSONResponse(w, userListResponse(users))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"user", "list", "--include-disabled"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedVars["includeDisabled"] != true {
		t.Errorf("expected includeDisabled=true in request vars, got: %v", capturedVars)
	}
}

func TestUserListCommand_JSONOutput(t *testing.T) {
	users := []map[string]any{
		makeUser("u1", "Alice Smith", "alice@example.com", true, false, false),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, userListResponse(users))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "user", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 user, got %d", len(decoded))
	}
	if decoded[0]["displayName"] != "Alice Smith" {
		t.Errorf("expected displayName Alice Smith, got %v", decoded[0]["displayName"])
	}
	if decoded[0]["email"] != "alice@example.com" {
		t.Errorf("expected email alice@example.com, got %v", decoded[0]["email"])
	}
}

func TestUserShowCommand_TableOutput(t *testing.T) {
	user := makeUser("u1", "Alice Smith", "alice@example.com", true, false, false)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, userGetResponse(user))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"user", "show", "u1"})

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
	if !strings.Contains(result, "Active:") || !strings.Contains(result, "yes") {
		t.Errorf("output should contain Active: yes, got:\n%s", result)
	}
}

func TestUserShowCommand_NotFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{"data": map[string]any{"user": nil}})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"user", "show", "missing-id"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing user, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected not found error, got: %v", err)
	}
}

func TestUserShowCommand_JSONOutput(t *testing.T) {
	user := makeUser("u1", "Alice Smith", "alice@example.com", true, true, false)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, userGetResponse(user))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "user", "show", "u1"})

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
	if decoded["admin"] != true {
		t.Errorf("expected admin=true, got %v", decoded["admin"])
	}
}

func TestUserShowCommand_RequiresArg(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"user", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no arg provided")
	}
}
