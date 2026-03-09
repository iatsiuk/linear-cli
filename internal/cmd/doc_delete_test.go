package cmd_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func docDeleteResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"documentDelete": map[string]any{
				"success": success,
			},
		},
	}
}

func docUnarchiveResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"documentUnarchive": map[string]any{
				"success": success,
			},
		},
	}
}

func TestDocDeleteCommand_Basic(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000010"

	server, bodies := newQueuedServer(t, []map[string]any{
		docDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "delete", docID, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != docID {
		t.Errorf("mutation id = %v, want %q", (*bodies)[0]["id"], docID)
	}

	result := out.String()
	if !strings.Contains(result, "trash") {
		t.Errorf("output should mention trash, got: %s", result)
	}
}

func TestDocDeleteCommand_RestoreFlag(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000011"

	server, bodies := newQueuedServer(t, []map[string]any{
		docUnarchiveResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "delete", docID, "--restore", "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != docID {
		t.Errorf("mutation id = %v, want %q", (*bodies)[0]["id"], docID)
	}

	result := out.String()
	if !strings.Contains(result, "restored") {
		t.Errorf("output should mention restored, got: %s", result)
	}
}

func TestDocDeleteCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "delete"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when ID is missing")
	}
}

func TestDocDeleteCommand_SuccessFalse(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000012"

	server, _ := newQueuedServer(t, []map[string]any{
		docDeleteResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "delete", docID, "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestDocDeleteCommand_RestoreSuccessFalse(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000013"

	server, _ := newQueuedServer(t, []map[string]any{
		docUnarchiveResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"doc", "delete", docID, "--restore", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when restore success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestDocDeleteCommand_ConfirmationAbort(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000014"

	// no mutation expected — only the GET (which is none, since doc delete goes direct)
	server, bodies := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"doc", "delete", docID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user declines confirmation")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("error should mention aborted, got: %v", err)
	}
	if len(*bodies) != 0 {
		t.Errorf("no mutation should have been called, got %d requests", len(*bodies))
	}
}

func TestDocDeleteCommand_ConfirmationAccept(t *testing.T) {
	const docID = "00000000-0000-0000-0000-000000000015"

	server, _ := newQueuedServer(t, []map[string]any{
		docDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("y\n"))
	root.SetArgs([]string{"doc", "delete", docID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "trash") {
		t.Errorf("output should mention trash, got: %s", result)
	}
}
