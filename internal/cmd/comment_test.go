package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func makeComment(id, body, authorName string, parent map[string]any) map[string]any {
	c := map[string]any{
		"id":        id,
		"body":      body,
		"createdAt": "2026-01-01T10:00:00Z",
		"updatedAt": "2026-01-01T10:00:00Z",
		"url":       "https://linear.app/comment/" + id,
	}
	if authorName != "" {
		c["user"] = map[string]any{
			"id":          "user-" + id,
			"displayName": authorName,
			"email":       authorName + "@example.com",
		}
	}
	if parent != nil {
		c["parent"] = parent
	}
	return c
}

func commentListResponse(identifier string, comments []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issue": map[string]any{
				"id":         "issue-uuid-" + identifier,
				"identifier": identifier,
				"comments": map[string]any{
					"nodes":    comments,
					"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
				},
			},
		},
	}
}

func commentCreateResponse(comment map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"commentCreate": map[string]any{
				"success": true,
				"comment": comment,
			},
		},
	}
}

// TestCommentListCommand_TableOutput verifies table output for comment list.
func TestCommentListCommand_TableOutput(t *testing.T) {
	comments := []map[string]any{
		makeComment("c1", "First comment", "Alice", nil),
		makeComment("c2", "Second comment", "Bob", nil),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, commentListResponse("ENG-1", comments))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "list", "ENG-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, col := range []string{"AUTHOR", "DATE", "BODY"} {
		if !strings.Contains(result, col) {
			t.Errorf("output should contain %s column header, got:\n%s", col, result)
		}
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("output should contain author Alice, got:\n%s", result)
	}
	if !strings.Contains(result, "First comment") {
		t.Errorf("output should contain comment body, got:\n%s", result)
	}
	if !strings.Contains(result, "Bob") {
		t.Errorf("output should contain author Bob, got:\n%s", result)
	}
}

// TestCommentListCommand_ThreadedOutput verifies that replies are prefixed.
func TestCommentListCommand_ThreadedOutput(t *testing.T) {
	parent := makeComment("c1", "Parent comment", "Alice", nil)
	reply := makeComment("c2", "Reply comment", "Bob", map[string]any{"id": "c1", "body": "Parent comment"})
	comments := []map[string]any{parent, reply}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, commentListResponse("ENG-2", comments))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "list", "ENG-2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "> Reply comment") {
		t.Errorf("reply should be prefixed with '> ', got:\n%s", result)
	}
	if !strings.Contains(result, "Parent comment") {
		t.Errorf("output should contain parent comment body, got:\n%s", result)
	}
	if strings.Contains(result, "> Parent comment") {
		t.Errorf("parent comment should not be prefixed with '> ', got:\n%s", result)
	}
}

// TestCommentListCommand_JSONOutput verifies JSON output for comment list.
func TestCommentListCommand_JSONOutput(t *testing.T) {
	comments := []map[string]any{
		makeComment("c1", "Hello world", "Alice", nil),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, commentListResponse("ENG-1", comments))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "comment", "list", "ENG-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 comment, got %d", len(decoded))
	}
	if decoded[0]["body"] != "Hello world" {
		t.Errorf("expected body 'Hello world', got %v", decoded[0]["body"])
	}
}

// TestCommentListCommand_IssueNotFound verifies error when issue is not found.
func TestCommentListCommand_IssueNotFound(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, map[string]any{
			"data": map[string]any{"issue": nil},
		})
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "list", "ENG-999"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when issue not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

// TestCommentListCommand_MissingIdentifier verifies error when identifier is missing.
func TestCommentListCommand_MissingIdentifier(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when identifier is missing")
	}
}

// TestCommentCreateCommand_Basic verifies that create sends issue identifier and body.
func TestCommentCreateCommand_Basic(t *testing.T) {
	comment := makeComment("new-comment-id", "This is a comment", "Alice", nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		commentCreateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "create", "ENG-1", "--body", "This is a comment"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "new-comment-id") {
		t.Errorf("output should contain comment ID, got: %s", result)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	input, ok := (*bodies)[0]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[0])
	}
	if input["issueId"] != "ENG-1" {
		t.Errorf("issueId = %v, want ENG-1", input["issueId"])
	}
	if input["body"] != "This is a comment" {
		t.Errorf("body = %v, want 'This is a comment'", input["body"])
	}
}

// TestCommentCreateCommand_WithParent verifies threading via --parent flag.
func TestCommentCreateCommand_WithParent(t *testing.T) {
	const parentID = "parent-comment-uuid-001"
	comment := makeComment("reply-id", "Reply text", "Bob", map[string]any{"id": parentID})

	server, bodies := newQueuedServer(t, []map[string]any{
		commentCreateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "create", "ENG-2", "--body", "Reply text", "--parent", parentID})

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
	if input["parentId"] != parentID {
		t.Errorf("parentId = %v, want %s", input["parentId"], parentID)
	}
}

// TestCommentCreateCommand_MissingBody verifies error when --body is not provided.
func TestCommentCreateCommand_MissingBody(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "create", "ENG-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --body is missing")
	}
	if !strings.Contains(err.Error(), "body") {
		t.Errorf("error should mention body, got: %v", err)
	}
}

// TestCommentCreateCommand_SuccessFalse verifies error when mutation returns success=false.
func TestCommentCreateCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"commentCreate": map[string]any{
					"success": false,
					"comment": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "create", "ENG-3", "--body", "Test"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func commentUpdateResponse(comment map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"commentUpdate": map[string]any{
				"success": true,
				"comment": comment,
			},
		},
	}
}

// TestCommentUpdate_Success verifies that update outputs confirmation.
func TestCommentUpdate_Success(t *testing.T) {
	comment := makeComment("comment-uuid-1", "Updated body", "Alice", nil)

	server, _ := newQueuedServer(t, []map[string]any{
		commentUpdateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "update", "comment-uuid-1", "--body", "Updated body"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "comment-uuid-1") {
		t.Errorf("output should contain comment ID, got: %s", result)
	}
	if !strings.Contains(result, "updated") {
		t.Errorf("output should contain 'updated', got: %s", result)
	}
}

// TestCommentUpdate_JSON verifies JSON output for comment update.
func TestCommentUpdate_JSON(t *testing.T) {
	comment := makeComment("comment-uuid-2", "New body", "Bob", nil)

	server, _ := newQueuedServer(t, []map[string]any{
		commentUpdateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "comment", "update", "comment-uuid-2", "--body", "New body"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["body"] != "New body" {
		t.Errorf("expected body 'New body', got %v", decoded["body"])
	}
}

// TestCommentUpdate_MissingBody verifies error when --body not provided.
func TestCommentUpdate_MissingBody(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "update", "comment-uuid-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --body is missing")
	}
	if !strings.Contains(err.Error(), "body") {
		t.Errorf("error should mention body, got: %v", err)
	}
}

// TestCommentUpdate_NotFound verifies error when API returns error.
func TestCommentUpdate_NotFound(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{"errors": []map[string]any{{"message": "Comment not found"}}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "update", "nonexistent-id", "--body", "text"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when comment not found")
	}
}

// TestCommentUpdate_BodyFile verifies that --body-file reads body from a file.
func TestCommentUpdate_BodyFile(t *testing.T) {
	const fileContent = "Updated body from file\nwith multiple lines\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "body.md")
	if err := os.WriteFile(path, []byte(fileContent), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	comment := makeComment("update-file-id", fileContent, "Alice", nil)
	server, bodies := newQueuedServer(t, []map[string]any{
		commentUpdateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "update", "update-file-id", "--body-file", path})

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
	if input["body"] != fileContent {
		t.Errorf("input.body = %q, want %q", input["body"], fileContent)
	}
}

// TestCommentUpdate_BodyFileStdin verifies that --body-file - reads body from stdin.
func TestCommentUpdate_BodyFileStdin(t *testing.T) {
	const stdinContent = "from stdin update"
	comment := makeComment("update-stdin-id", stdinContent, "Alice", nil)
	server, bodies := newQueuedServer(t, []map[string]any{
		commentUpdateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdinContent))
	root.SetArgs([]string{"comment", "update", "update-stdin-id", "--body-file", "-"})

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
	if input["body"] != stdinContent {
		t.Errorf("input.body = %q, want %q", input["body"], stdinContent)
	}
}

// TestCommentUpdate_BodyAndBodyFileMutuallyExclusive verifies that --body and
// --body-file cannot be combined.
func TestCommentUpdate_BodyAndBodyFileMutuallyExclusive(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "update", "comment-uuid-1", "--body", "x", "--body-file", "/tmp/x"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both --body and --body-file are set")
	}
	msg := err.Error()
	if !strings.Contains(msg, "body") || !strings.Contains(msg, "body-file") {
		t.Errorf("error should mention both flag names, got: %v", err)
	}
	if !strings.Contains(msg, "none of the others can be") {
		t.Errorf("error should signal mutual exclusion, got: %v", err)
	}
}

// TestCommentUpdate_NoBodyOrBodyFile verifies that one of --body / --body-file is required.
func TestCommentUpdate_NoBodyOrBodyFile(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "update", "comment-uuid-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when neither --body nor --body-file is set")
	}
	msg := err.Error()
	if !strings.Contains(msg, "body") || !strings.Contains(msg, "required") {
		t.Errorf("error should mention body and required, got: %v", err)
	}
}

// TestCommentUpdate_VerifyInput verifies that id and input.body are sent correctly.
func TestCommentUpdate_VerifyInput(t *testing.T) {
	comment := makeComment("target-id", "New text", "Carol", nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		commentUpdateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "update", "target-id", "--body", "New text"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	req := (*bodies)[0]
	if req["id"] != "target-id" {
		t.Errorf("id = %v, want target-id", req["id"])
	}
	input, ok := req["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", req)
	}
	if input["body"] != "New text" {
		t.Errorf("input.body = %v, want 'New text'", input["body"])
	}
}

func commentDeleteResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"commentDelete": map[string]any{
				"success": success,
			},
		},
	}
}

// TestCommentDelete_Success verifies deletion with --yes flag outputs confirmation.
func TestCommentDelete_Success(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		commentDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "delete", "comment-uuid-1", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "comment-uuid-1") {
		t.Errorf("output should contain comment ID, got: %s", result)
	}
	if !strings.Contains(result, "deleted") {
		t.Errorf("output should contain 'deleted', got: %s", result)
	}
}

// TestCommentDelete_WithYesFlag verifies that --yes skips confirmation prompt.
func TestCommentDelete_WithYesFlag(t *testing.T) {
	server, bodies := newQueuedServer(t, []map[string]any{
		commentDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("")) // empty stdin - should not be read
	root.SetArgs([]string{"comment", "delete", "skip-confirm-id", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != "skip-confirm-id" {
		t.Errorf("id = %v, want skip-confirm-id", (*bodies)[0]["id"])
	}
}

// TestCommentDelete_Abort verifies that declining confirmation aborts deletion.
func TestCommentDelete_Abort(t *testing.T) {
	server, bodies := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"comment", "delete", "comment-abort-id"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user declines confirmation")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("error should mention aborted, got: %v", err)
	}
	if len(*bodies) != 0 {
		t.Errorf("expected 0 requests (no mutation), got %d", len(*bodies))
	}
}

// TestCommentDelete_NotFound verifies error when API returns error.
func TestCommentDelete_NotFound(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{"errors": []map[string]any{{"message": "Comment not found"}}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "delete", "nonexistent-id", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when comment not found")
	}
}

// TestCommentDelete_MutationFails verifies error when mutation returns success=false.
func TestCommentDelete_MutationFails(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		commentDeleteResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "delete", "fail-id", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

// TestCommentCreate_BodyFile verifies that --body-file reads body from a file.
func TestCommentCreate_BodyFile(t *testing.T) {
	const fileContent = "This is body from file\nwith multiple lines\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "body.md")
	if err := os.WriteFile(path, []byte(fileContent), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	comment := makeComment("file-comment-id", fileContent, "Alice", nil)
	server, bodies := newQueuedServer(t, []map[string]any{
		commentCreateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "create", "ENG-1", "--body-file", path})

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
	if input["body"] != fileContent {
		t.Errorf("body = %q, want %q", input["body"], fileContent)
	}
}

// TestCommentCreate_BodyFileStdin verifies that --body-file - reads body from stdin.
func TestCommentCreate_BodyFileStdin(t *testing.T) {
	const stdinContent = "from stdin"
	comment := makeComment("stdin-comment-id", stdinContent, "Alice", nil)
	server, bodies := newQueuedServer(t, []map[string]any{
		commentCreateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdinContent))
	root.SetArgs([]string{"comment", "create", "ENG-1", "--body-file", "-"})

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
	if input["body"] != stdinContent {
		t.Errorf("body = %q, want %q", input["body"], stdinContent)
	}
}

// TestCommentCreate_BodyAndBodyFileMutuallyExclusive verifies that --body and
// --body-file cannot be combined.
func TestCommentCreate_BodyAndBodyFileMutuallyExclusive(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "create", "ENG-1", "--body", "x", "--body-file", "/tmp/x"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both --body and --body-file are set")
	}
	msg := err.Error()
	if !strings.Contains(msg, "body") || !strings.Contains(msg, "body-file") {
		t.Errorf("error should mention both flag names, got: %v", err)
	}
	if !strings.Contains(msg, "none of the others can be") {
		t.Errorf("error should signal mutual exclusion, got: %v", err)
	}
}

// TestCommentCreate_NoBodyOrBodyFile verifies that one of --body / --body-file is required.
func TestCommentCreate_NoBodyOrBodyFile(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "create", "ENG-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when neither --body nor --body-file is set")
	}
	msg := err.Error()
	if !strings.Contains(msg, "body") || !strings.Contains(msg, "required") {
		t.Errorf("error should mention body and required, got: %v", err)
	}
}

// TestCommentCreate_BodyFileMissing verifies error includes path when file is missing.
func TestCommentCreate_BodyFileMissing(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	missingPath := filepath.Join(t.TempDir(), "does-not-exist.md")

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"comment", "create", "ENG-1", "--body-file", missingPath})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when body file is missing")
	}
	if !strings.Contains(err.Error(), "read body file") {
		t.Errorf("error should mention 'read body file', got: %v", err)
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Errorf("error should contain path %q, got: %v", missingPath, err)
	}
}

// TestCommentCreateCommand_JSONOutput verifies JSON output for comment create.
func TestCommentCreateCommand_JSONOutput(t *testing.T) {
	comment := makeComment("json-comment-id", "A comment", "Carol", nil)

	server, _ := newQueuedServer(t, []map[string]any{
		commentCreateResponse(comment),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "comment", "create", "ENG-4", "--body", "A comment"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["id"] != "json-comment-id" {
		t.Errorf("expected id 'json-comment-id', got %v", decoded["id"])
	}
	if decoded["body"] != "A comment" {
		t.Errorf("expected body 'A comment', got %v", decoded["body"])
	}
}
