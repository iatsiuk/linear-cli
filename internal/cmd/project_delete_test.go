package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func projectDeleteResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projectDelete": map[string]any{
				"success": success,
			},
		},
	}
}

func TestProjectDeleteCommand_Basic(t *testing.T) {
	const projDelID = "00000000-0000-0000-0000-000000000011"
	p := makeProject(projDelID, "Delete Me", "planned", "", 0.0, "")

	server, bodies := newQueuedServer(t, []map[string]any{
		projectGetResponse(p),
		projectDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "delete", projDelID, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	// second request is the delete mutation
	mutationVars := (*bodies)[1]
	if mutationVars["id"] != projDelID {
		t.Errorf("mutation id = %v, want %s", mutationVars["id"], projDelID)
	}

	result := out.String()
	if !strings.Contains(result, "deleted") {
		t.Errorf("output should mention deleted, got: %s", result)
	}
	if !strings.Contains(result, "Delete Me") {
		t.Errorf("output should contain project name, got: %s", result)
	}
}

func TestProjectDeleteCommand_NotFound(t *testing.T) {
	const nonexistentID = "00000000-0000-0000-0000-000000000099"
	server, _ := newQueuedServer(t, []map[string]any{
		{"data": map[string]any{"project": nil}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "delete", nonexistentID, "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when project not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestProjectDeleteCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "delete"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when id is missing")
	}
}

func TestProjectDeleteCommand_SuccessFalse(t *testing.T) {
	const projFailID = "00000000-0000-0000-0000-000000000012"
	p := makeProject(projFailID, "Fail Delete", "planned", "", 0.0, "")

	server, _ := newQueuedServer(t, []map[string]any{
		projectGetResponse(p),
		projectDeleteResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "delete", projFailID, "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestProjectDeleteCommand_ConfirmationPrompt(t *testing.T) {
	const projConfirmID = "00000000-0000-0000-0000-000000000013"
	p := makeProject(projConfirmID, "Confirm Me", "planned", "", 0.0, "")

	// only GET queued; delete should not proceed
	server, bodies := newQueuedServer(t, []map[string]any{
		projectGetResponse(p),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"project", "delete", projConfirmID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user declines confirmation")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("error should mention aborted, got: %v", err)
	}
	// only GET was called
	if len(*bodies) != 1 {
		t.Errorf("expected 1 request (GET only), got %d", len(*bodies))
	}
}

func TestProjectDeleteCommand_ConfirmationYes(t *testing.T) {
	const projYID = "00000000-0000-0000-0000-000000000014"
	p := makeProject(projYID, "Yes Delete", "planned", "", 0.0, "")

	server, _ := newQueuedServer(t, []map[string]any{
		projectGetResponse(p),
		projectDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("y\n"))
	root.SetArgs([]string{"project", "delete", projYID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "deleted") {
		t.Errorf("output should mention deleted, got: %s", result)
	}
}
