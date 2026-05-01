package cmd_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func issueUpdateResponse(issue map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueUpdate": map[string]any{
				"success": true,
				"issue":   issue,
			},
		},
	}
}

func labelResolveResponse(labelID string) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueLabels": map[string]any{
				"nodes": []map[string]any{{"id": labelID}},
			},
		},
	}
}

func stateResolveResponse(stateID string) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"workflowStates": map[string]any{
				"nodes": []map[string]any{{"id": stateID}},
			},
		},
	}
}

func TestIssueUpdateCommand_Basic(t *testing.T) {
	issue := makeIssue("ENG-10", "Updated title", "In Progress", "High", "")
	issue["id"] = "issue-uuid-1234"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueUpdateResponse(issue),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update", "ENG-10", "--title", "Updated title"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "ENG-10") {
		t.Errorf("output should contain ENG-10, got: %s", result)
	}

	// verify mutation vars: id should be issue UUID, input should have title
	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[1]
	if mutationVars["id"] != "issue-uuid-1234" {
		t.Errorf("mutation id = %v, want issue-uuid-1234", mutationVars["id"])
	}
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set in mutation vars: %v", mutationVars)
	}
	if input["title"] != "Updated title" {
		t.Errorf("input.title = %v, want Updated title", input["title"])
	}
}

func TestIssueUpdateCommand_JSONOutput(t *testing.T) {
	issue := makeIssue("ENG-20", "JSON update", "Done", "Low", "")
	issue["id"] = "issue-uuid-2345"

	server, _ := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueUpdateResponse(issue),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "issue", "update", "ENG-20", "--title", "JSON update"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["identifier"] != "ENG-20" {
		t.Errorf("expected identifier ENG-20, got %v", decoded["identifier"])
	}
}

func TestIssueUpdateCommand_PartialUpdate(t *testing.T) {
	issue := makeIssue("ENG-30", "Original", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-3456"
	updated := makeIssue("ENG-30", "Updated", "Todo", "High", "")

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueUpdateResponse(updated),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update", "ENG-30", "--title", "Updated"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mutationVars := (*bodies)[1]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set in mutation vars")
	}
	// only title should be present, no description/assigneeId/stateId/etc.
	for _, key := range []string{"description", "assigneeId", "stateId", "priority", "labelIds", "addedLabelIds", "removedLabelIds", "dueDate", "estimate", "cycleId", "projectId", "parentId"} {
		if _, present := input[key]; present {
			t.Errorf("optional field %q should not be in input when not provided", key)
		}
	}
	if input["title"] != "Updated" {
		t.Errorf("input.title = %v, want Updated", input["title"])
	}
}

func TestIssueUpdateCommand_StateResolution(t *testing.T) {
	issue := makeIssue("ENG-40", "Some issue", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-4567"
	const stateID = "state-uuid-done"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		stateResolveResponse(stateID),
		issueUpdateResponse(makeIssue("ENG-40", "Some issue", "Done", "No priority", "")),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update", "ENG-40", "--state", "Done"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mutationVars := (*bodies)[2]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set")
	}
	if input["stateId"] != stateID {
		t.Errorf("input.stateId = %v, want %s", input["stateId"], stateID)
	}
}

func TestIssueUpdateCommand_AddLabel(t *testing.T) {
	issue := makeIssue("ENG-50", "Label test", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-5678"
	const labelID = "label-uuid-bug"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		labelResolveResponse(labelID),
		issueUpdateResponse(issue),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update", "ENG-50", "--add-label", "Bug"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mutationVars := (*bodies)[2]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set")
	}
	addedIDs, ok := input["addedLabelIds"].([]any)
	if !ok {
		t.Fatalf("input.addedLabelIds not set or wrong type: %v", input["addedLabelIds"])
	}
	if len(addedIDs) != 1 || addedIDs[0] != labelID {
		t.Errorf("addedLabelIds = %v, want [%s]", addedIDs, labelID)
	}
	if _, present := input["removedLabelIds"]; present {
		t.Error("removedLabelIds should not be set when only --add-label used")
	}
}

func TestIssueUpdateCommand_RemoveLabel(t *testing.T) {
	issue := makeIssue("ENG-60", "Remove label", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-6789"
	const labelID = "label-uuid-feat"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		labelResolveResponse(labelID),
		issueUpdateResponse(issue),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update", "ENG-60", "--remove-label", "Feature"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mutationVars := (*bodies)[2]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set")
	}
	removedIDs, ok := input["removedLabelIds"].([]any)
	if !ok {
		t.Fatalf("input.removedLabelIds not set or wrong type: %v", input["removedLabelIds"])
	}
	if len(removedIDs) != 1 || removedIDs[0] != labelID {
		t.Errorf("removedLabelIds = %v, want [%s]", removedIDs, labelID)
	}
	if _, present := input["addedLabelIds"]; present {
		t.Error("addedLabelIds should not be set when only --remove-label used")
	}
}

func TestIssueUpdateCommand_MissingIdentifier(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when identifier is missing")
	}
}

func TestIssueUpdateCommand_PayloadSuccessFalse(t *testing.T) {
	issue := makeIssue("ENG-70", "Fail", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-7890"

	server, _ := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		{
			"data": map[string]any{
				"issueUpdate": map[string]any{
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
	root.SetArgs([]string{"issue", "update", "ENG-70", "--title", "Fail"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

// TestIssueUpdate_DescriptionFile verifies that --description-file reads description from a file.
func TestIssueUpdate_DescriptionFile(t *testing.T) {
	const fileContent = "Updated description from file\n\n## Section\n- bullet\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "desc.md")
	if err := os.WriteFile(path, []byte(fileContent), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	issue := makeIssue("ENG-80", "Title", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-8888"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueUpdateResponse(issue),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update", "ENG-80", "--description-file", path})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[1]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", mutationVars)
	}
	if input["description"] != fileContent {
		t.Errorf("description = %q, want %q", input["description"], fileContent)
	}
}

// TestIssueUpdate_DescriptionFileStdin verifies that --description-file - reads from stdin.
func TestIssueUpdate_DescriptionFileStdin(t *testing.T) {
	const stdinContent = "description from stdin"
	issue := makeIssue("ENG-81", "Stdin", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-8181"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueUpdateResponse(issue),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdinContent))
	root.SetArgs([]string{"issue", "update", "ENG-81", "--description-file", "-"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[1]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", mutationVars)
	}
	if input["description"] != stdinContent {
		t.Errorf("description = %q, want %q", input["description"], stdinContent)
	}
}

// TestIssueUpdate_DescriptionAndDescriptionFileMutuallyExclusive verifies that
// --description and --description-file cannot be combined.
func TestIssueUpdate_DescriptionAndDescriptionFileMutuallyExclusive(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{
		"issue", "update", "ENG-82",
		"--description", "inline",
		"--description-file", "/tmp/x",
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both --description and --description-file are set")
	}
	msg := err.Error()
	if !strings.Contains(msg, "description") || !strings.Contains(msg, "description-file") {
		t.Errorf("error should mention both flag names, got: %v", err)
	}
	if !strings.Contains(msg, "none of the others can be") {
		t.Errorf("error should signal mutual exclusion, got: %v", err)
	}
}

// TestIssueUpdate_NoDescriptionFlags verifies that omitting both --description and
// --description-file leaves the description field absent from the mutation input,
// preserving partial-update semantics.
func TestIssueUpdate_NoDescriptionFlags(t *testing.T) {
	issue := makeIssue("ENG-83", "Original", "Todo", "No priority", "")
	issue["id"] = "issue-uuid-8383"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
		issueUpdateResponse(issue),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update", "ENG-83", "--title", "New title"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	mutationVars := (*bodies)[1]
	input, ok := mutationVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", mutationVars)
	}
	if _, present := input["description"]; present {
		t.Errorf("description should not be in input when neither flag is set, got: %v", input["description"])
	}
}

func TestIssueUpdate_DescriptionFileMissing(t *testing.T) {
	issue := makeIssue("ENG-10", "Existing title", "In Progress", "High", "")
	issue["id"] = "issue-uuid-1234"

	server, _ := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue),
	})
	setupIssueTest(t, server)

	missingPath := filepath.Join(t.TempDir(), "does-not-exist.md")

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{
		"issue", "update", "ENG-10",
		"--description-file", missingPath,
	})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when description file is missing")
	}
	if !strings.Contains(err.Error(), "read description file") {
		t.Errorf("error should mention 'read description file', got: %v", err)
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Errorf("error should contain path %q, got: %v", missingPath, err)
	}
}

func TestIssueUpdateCommand_IssueNotFound(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{"data": map[string]any{"issue": nil}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "update", "ENG-99", "--title", "New title"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when issue not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}
