package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func issueDeleteResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueDelete": map[string]any{
				"success": success,
			},
		},
	}
}

func issueArchiveResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueArchive": map[string]any{
				"success": success,
			},
		},
	}
}

func TestIssueDeleteCommand_Basic(t *testing.T) {
	issue := makeIssue("ENG-10", "Delete me", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-del1"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "delete", "ENG-10", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// first request: GET (resolve identifier)
	// second request: DELETE mutation
	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[1]
	if mutationVars["id"] != "issue-uuid-del1" {
		t.Errorf("mutation id = %v, want issue-uuid-del1", mutationVars["id"])
	}

	result := out.String()
	if !strings.Contains(result, "deleted") || !strings.Contains(result, "ENG-10") {
		t.Errorf("output should mention deleted and ENG-10, got: %s", result)
	}
}

func TestIssueDeleteCommand_ArchiveFlag(t *testing.T) {
	issue := makeIssue("ENG-20", "Archive me", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-arc1"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueArchiveResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "delete", "ENG-20", "--archive", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[1]
	if mutationVars["id"] != "issue-uuid-arc1" {
		t.Errorf("mutation id = %v, want issue-uuid-arc1", mutationVars["id"])
	}

	result := out.String()
	if !strings.Contains(result, "archived") || !strings.Contains(result, "ENG-20") {
		t.Errorf("output should mention archived and ENG-20, got: %s", result)
	}
}

func TestIssueDeleteCommand_MissingIdentifier(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "delete"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when identifier is missing")
	}
}

func TestIssueDeleteCommand_IssueNotFound(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{"data": map[string]any{"issue": nil}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "delete", "ENG-99", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when issue not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestIssueDeleteCommand_PayloadSuccessFalse(t *testing.T) {
	issue := makeIssue("ENG-30", "Fail delete", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-fail1"

	server, _ := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueDeleteResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "delete", "ENG-30", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestIssueArchivePayloadSuccessFalse(t *testing.T) {
	issue := makeIssue("ENG-31", "Fail archive", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-fail2"

	server, _ := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueArchiveResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "delete", "ENG-31", "--archive", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when archive success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestIssueDeleteCommand_ConfirmationPrompt(t *testing.T) {
	issue := makeIssue("ENG-40", "Confirm me", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-conf1"

	// only one response queued (GET): the delete should not proceed without confirmation
	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	// feed "n" to deny confirmation
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"issue", "delete", "ENG-40"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user declines confirmation")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("error should mention aborted, got: %v", err)
	}
	// only GET was called, delete mutation was not
	if len(*bodies) != 1 {
		t.Errorf("expected 1 request (GET only), got %d", len(*bodies))
	}
}
