package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func issueBatchUpdateResponse(issues []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueBatchUpdate": map[string]any{
				"issues": issues,
			},
		},
	}
}

// validUUID is a properly formatted UUID for use in tests where we want
// to skip resolution API calls (resolver returns it as-is).
const validUUID1 = "aaaaaaaa-1111-2222-3333-444444444444"

func TestIssueBatchUpdate_ResolvesIdentifiersToUUIDs(t *testing.T) {
	issue1 := makeIssue("ENG-1", "Issue one", "Done", "High", "")
	issue1["id"] = "uuid-eng-1"
	issue2 := makeIssue("ENG-2", "Issue two", "Done", "Low", "")
	issue2["id"] = "uuid-eng-2"

	// state is a valid UUID so no resolution API call needed
	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue1),
		issueGetResponse(issue2),
		issueBatchUpdateResponse([]map[string]any{issue1, issue2}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-1", "ENG-2", "--state", validUUID1})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 3 {
		t.Fatalf("expected 3 requests (2 gets + 1 batch), got %d", len(*bodies))
	}

	// first two requests should be identifier lookups
	if (*bodies)[0]["id"] != "ENG-1" {
		t.Errorf("first get request id = %v, want ENG-1", (*bodies)[0]["id"])
	}
	if (*bodies)[1]["id"] != "ENG-2" {
		t.Errorf("second get request id = %v, want ENG-2", (*bodies)[1]["id"])
	}

	// batch mutation should use resolved UUIDs
	batchVars := (*bodies)[2]
	ids, ok := batchVars["ids"].([]any)
	if !ok {
		t.Fatalf("ids not set or wrong type: %v", batchVars["ids"])
	}
	if len(ids) != 2 {
		t.Errorf("expected 2 ids, got %d", len(ids))
	}
	if ids[0] != "uuid-eng-1" {
		t.Errorf("ids[0] = %v, want uuid-eng-1", ids[0])
	}
	if ids[1] != "uuid-eng-2" {
		t.Errorf("ids[1] = %v, want uuid-eng-2", ids[1])
	}

	input, ok := batchVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set in batch vars: %v", batchVars)
	}
	if input["stateId"] != validUUID1 {
		t.Errorf("input.stateId = %v, want %s", input["stateId"], validUUID1)
	}
}

func TestIssueBatchUpdate_TableOutput(t *testing.T) {
	issue1 := makeIssue("ENG-10", "First issue", "Done", "High", "Alice")
	issue1["id"] = "uuid-10"

	server, _ := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue1),
		issueBatchUpdateResponse([]map[string]any{issue1}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-10", "--state", validUUID1})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "ENG-10") {
		t.Errorf("output should contain ENG-10, got: %s", result)
	}
	if !strings.Contains(result, "First issue") {
		t.Errorf("output should contain issue title, got: %s", result)
	}
}

func TestIssueBatchUpdate_JSONOutput(t *testing.T) {
	issue1 := makeIssue("ENG-20", "JSON issue", "Done", "High", "")
	issue1["id"] = "uuid-20"

	server, _ := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue1),
		issueBatchUpdateResponse([]map[string]any{issue1}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "issue", "batch", "update", "ENG-20", "--state", validUUID1})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Errorf("expected 1 issue in JSON, got %d", len(decoded))
	}
	if decoded[0]["identifier"] != "ENG-20" {
		t.Errorf("identifier = %v, want ENG-20", decoded[0]["identifier"])
	}
}

func TestIssueBatchUpdate_StdinInput(t *testing.T) {
	issue1 := makeIssue("ENG-30", "Stdin issue one", "Done", "High", "")
	issue1["id"] = "uuid-30"
	issue2 := makeIssue("ENG-31", "Stdin issue two", "Done", "Low", "")
	issue2["id"] = "uuid-31"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue1),
		issueGetResponse(issue2),
		issueBatchUpdateResponse([]map[string]any{issue1, issue2}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	// provide identifiers via stdin, no args
	root.SetIn(strings.NewReader("ENG-30\nENG-31\n"))
	root.SetArgs([]string{"issue", "batch", "update", "--state", validUUID1})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != "ENG-30" {
		t.Errorf("first get id = %v, want ENG-30", (*bodies)[0]["id"])
	}
	if (*bodies)[1]["id"] != "ENG-31" {
		t.Errorf("second get id = %v, want ENG-31", (*bodies)[1]["id"])
	}
}

func TestIssueBatchUpdate_StdinSkipsEmptyLines(t *testing.T) {
	issue1 := makeIssue("ENG-40", "Line issue", "Done", "High", "")
	issue1["id"] = "uuid-40"

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue1),
		issueBatchUpdateResponse([]map[string]any{issue1}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	// stdin has empty lines and whitespace
	root.SetIn(strings.NewReader("\n  ENG-40  \n\n"))
	root.SetArgs([]string{"issue", "batch", "update", "--state", validUUID1})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	batchVars := (*bodies)[1]
	ids, ok := batchVars["ids"].([]any)
	if !ok {
		t.Fatalf("ids not set: %v", batchVars)
	}
	if len(ids) != 1 {
		t.Errorf("expected 1 id, got %d", len(ids))
	}
}

func TestIssueBatchUpdate_ValidationNoIdentifiers(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("")) // empty stdin
	root.SetArgs([]string{"issue", "batch", "update", "--state", validUUID1})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no identifiers given")
	}
	if !strings.Contains(err.Error(), "at least one identifier") {
		t.Errorf("error should mention 'at least one identifier', got: %v", err)
	}
}

func TestIssueBatchUpdate_ValidationNoChangeFlag(t *testing.T) {
	// the "no fields to update" check happens before identifier resolution, so no API calls are made
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-50"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no change flag given")
	}
	if !strings.Contains(err.Error(), "no fields to update") {
		t.Errorf("error should mention 'no fields to update', got: %v", err)
	}
}

func TestIssueBatchUpdate_InvalidPriority(t *testing.T) {
	// priority validation happens before identifier resolution, so no API calls are made
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-1", "--priority", "5"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for priority out of range")
	}
	if !strings.Contains(err.Error(), "priority must be 0-4") {
		t.Errorf("error should mention priority range, got: %v", err)
	}
}

func TestIssueBatchUpdate_ValidationTooManyItems(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	// build 51 identifiers
	ids := make([]string, 51)
	for i := range ids {
		ids[i] = "ENG-" + strings.Repeat("1", i+1)
	}

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	args := append([]string{"issue", "batch", "update"}, ids...)
	args = append(args, "--state", validUUID1)
	root.SetArgs(args)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when more than 50 items")
	}
	if !strings.Contains(err.Error(), "too many identifiers") {
		t.Errorf("error should mention 'too many identifiers', got: %v", err)
	}
}

func TestIssueBatchUpdate_ValidationLabelConflict(t *testing.T) {
	// label conflict check happens before identifier resolution, so no API calls are made
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-60", "--label", "Bug", "--add-label", "Feature"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --label combined with --add-label")
	}
	if !strings.Contains(err.Error(), "--label cannot be combined") {
		t.Errorf("error should mention --label conflict, got: %v", err)
	}
}

func TestIssueBatchUpdate_IssueNotFound(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{"data": map[string]any{"issue": nil}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-99", "--state", validUUID1})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when issue not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

func TestIssueBatchUpdate_PriorityFlag(t *testing.T) {
	issue1 := makeIssue("ENG-70", "Priority issue", "Todo", "Low", "")
	issue1["id"] = "uuid-70"
	updated := makeIssue("ENG-70", "Priority issue", "Todo", "Urgent", "")

	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue1),
		issueBatchUpdateResponse([]map[string]any{updated}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-70", "--priority", "1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	batchVars := (*bodies)[1]
	input, ok := batchVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", batchVars)
	}
	if input["priority"] != float64(1) {
		t.Errorf("input.priority = %v, want 1", input["priority"])
	}
}

func TestIssueBatchUpdate_AddLabelFlag(t *testing.T) {
	issue1 := makeIssue("ENG-80", "Add label issue", "Todo", "Low", "")
	issue1["id"] = "uuid-80"

	// validUUID1 is already a UUID, so ResolveLabelID returns it without an API call
	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue1),
		issueBatchUpdateResponse([]map[string]any{issue1}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-80", "--add-label", validUUID1})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	batchVars := (*bodies)[1]
	input, ok := batchVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", batchVars)
	}
	addedIDs, ok := input["addedLabelIds"].([]any)
	if !ok {
		t.Fatalf("addedLabelIds not set or wrong type: %v", input)
	}
	if len(addedIDs) != 1 || addedIDs[0] != validUUID1 {
		t.Errorf("addedLabelIds = %v, want [%s]", addedIDs, validUUID1)
	}
}

func TestIssueBatchUpdate_RemoveLabelFlag(t *testing.T) {
	issue1 := makeIssue("ENG-81", "Remove label issue", "Todo", "Low", "")
	issue1["id"] = "uuid-81"

	// validUUID1 is already a UUID, so ResolveLabelID returns it without an API call
	server, bodies := newQueuedServer(t, []map[string]any{
		issueGetResponse(issue1),
		issueBatchUpdateResponse([]map[string]any{issue1}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "batch", "update", "ENG-81", "--remove-label", validUUID1})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	batchVars := (*bodies)[1]
	input, ok := batchVars["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", batchVars)
	}
	removedIDs, ok := input["removedLabelIds"].([]any)
	if !ok {
		t.Fatalf("removedLabelIds not set or wrong type: %v", input)
	}
	if len(removedIDs) != 1 || removedIDs[0] != validUUID1 {
		t.Errorf("removedLabelIds = %v, want [%s]", removedIDs, validUUID1)
	}
}
