package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func makeTeamMembership(id, userID, displayName, email string, owner bool) map[string]any {
	return map[string]any{
		"id":        id,
		"owner":     owner,
		"sortOrder": 1.0,
		"user": map[string]any{
			"id":          userID,
			"displayName": displayName,
			"email":       email,
		},
	}
}

func teamMemberListResponse(memberships []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"team": map[string]any{
				"memberships": map[string]any{
					"nodes": memberships,
				},
			},
		},
	}
}

func teamMemberAddResponse(membership map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"teamMembershipCreate": map[string]any{
				"success":        true,
				"teamMembership": membership,
			},
		},
	}
}

func teamMemberRemoveResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"teamMembershipDelete": map[string]any{
				"success": success,
			},
		},
	}
}

const testTeamID = "00000000-0000-0000-0000-000000000010"
const testUserID = "00000000-0000-0000-0000-000000000020"

func TestTeamMemberListCommand_Basic(t *testing.T) {
	memberships := []map[string]any{
		makeTeamMembership("tm-1", "u-1", "Alice", "alice@example.com", true),
		makeTeamMembership("tm-2", "u-2", "Bob", "bob@example.com", false),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		teamMemberListResponse(memberships),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "list", testTeamID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["teamId"] != testTeamID {
		t.Errorf("teamId = %v, want %s", (*bodies)[0]["teamId"], testTeamID)
	}

	result := out.String()
	if !strings.Contains(result, "Alice") {
		t.Errorf("output should contain Alice, got: %s", result)
	}
	if !strings.Contains(result, "bob@example.com") {
		t.Errorf("output should contain bob@example.com, got: %s", result)
	}
	if !strings.Contains(result, "Owner") {
		t.Errorf("output should contain Owner role, got: %s", result)
	}
	if !strings.Contains(result, "Member") {
		t.Errorf("output should contain Member role, got: %s", result)
	}
}

func TestTeamMemberListCommand_JSONOutput(t *testing.T) {
	memberships := []map[string]any{
		makeTeamMembership("tm-1", "u-1", "Alice", "alice@example.com", true),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		teamMemberListResponse(memberships),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "team", "member", "list", testTeamID})

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
}

func TestTeamMemberListCommand_MissingArg(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when team key is missing")
	}
}

func TestTeamMemberAddCommand_Basic(t *testing.T) {
	// team UUID and user UUID - no resolver calls needed
	membership := makeTeamMembership("tm-new", testUserID, "Carol", "carol@example.com", false)

	server, bodies := newQueuedServer(t, []map[string]any{
		teamMemberAddResponse(membership),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "add", testTeamID, testUserID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	input, ok := (*bodies)[0]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[0])
	}
	if input["teamId"] != testTeamID {
		t.Errorf("teamId = %v, want %s", input["teamId"], testTeamID)
	}
	if input["userId"] != testUserID {
		t.Errorf("userId = %v, want %s", input["userId"], testUserID)
	}

	if !strings.Contains(out.String(), "Carol") {
		t.Errorf("output should contain user name, got: %s", out.String())
	}
}

func TestTeamMemberAddCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"teamMembershipCreate": map[string]any{
					"success":        false,
					"teamMembership": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "add", testTeamID, testUserID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestTeamMemberAddCommand_MissingArgs(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "add", testTeamID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user is missing")
	}
}

func TestTeamMemberRemoveCommand_Basic(t *testing.T) {
	const membershipID = "tm-remove-1"
	memberships := []map[string]any{
		makeTeamMembership(membershipID, testUserID, "Dave", "dave@example.com", false),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		teamMemberListResponse(memberships),
		teamMemberRemoveResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "remove", testTeamID, testUserID, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	// first request: list memberships
	if (*bodies)[0]["teamId"] != testTeamID {
		t.Errorf("list request teamId = %v, want %s", (*bodies)[0]["teamId"], testTeamID)
	}
	// second request: delete membership
	if (*bodies)[1]["id"] != membershipID {
		t.Errorf("delete request id = %v, want %s", (*bodies)[1]["id"], membershipID)
	}

	if !strings.Contains(out.String(), "Removed") {
		t.Errorf("output should contain 'Removed', got: %s", out.String())
	}
}

func TestTeamMemberRemoveCommand_UserNotMember(t *testing.T) {
	// user is not in the team
	memberships := []map[string]any{
		makeTeamMembership("tm-1", "u-other", "Other", "other@example.com", false),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		teamMemberListResponse(memberships),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "remove", testTeamID, testUserID, "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user is not a member")
	}
	if !strings.Contains(err.Error(), "not a member") {
		t.Errorf("error should mention 'not a member', got: %v", err)
	}
}

func TestTeamMemberRemoveCommand_Aborted(t *testing.T) {
	server, bodies := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"team", "member", "remove", testTeamID, testUserID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user declines confirmation")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("error should mention aborted, got: %v", err)
	}
	if len(*bodies) != 0 {
		t.Errorf("expected 0 requests, got %d", len(*bodies))
	}
}

func TestTeamMemberRemoveCommand_SuccessFalse(t *testing.T) {
	memberships := []map[string]any{
		makeTeamMembership("tm-1", testUserID, "Eve", "eve@example.com", false),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		teamMemberListResponse(memberships),
		teamMemberRemoveResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "remove", testTeamID, testUserID, "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestTeamMemberRemoveCommand_MissingArgs(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"team", "member", "remove", testTeamID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user is missing")
	}
}
