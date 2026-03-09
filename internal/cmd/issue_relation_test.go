package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func makeRelation(id, relType string, issue, relatedIssue map[string]any) map[string]any {
	return map[string]any{
		"id":           id,
		"type":         relType,
		"createdAt":    "2026-01-01T00:00:00Z",
		"updatedAt":    "2026-01-01T00:00:00Z",
		"issue":        issue,
		"relatedIssue": relatedIssue,
	}
}

func relationListResponse(relations, inverseRelations []map[string]any) map[string]any {
	if relations == nil {
		relations = []map[string]any{}
	}
	if inverseRelations == nil {
		inverseRelations = []map[string]any{}
	}
	return map[string]any{
		"data": map[string]any{
			"issue": map[string]any{
				"relations":        map[string]any{"nodes": relations},
				"inverseRelations": map[string]any{"nodes": inverseRelations},
			},
		},
	}
}

func relationCreateResponse(relation map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueRelationCreate": map[string]any{
				"success":       true,
				"issueRelation": relation,
			},
		},
	}
}

func relationDeleteResponse(entityID string) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueRelationDelete": map[string]any{
				"success":  true,
				"entityId": entityID,
			},
		},
	}
}

// TestRelationListCommand_TableOutput verifies table output showing type and direction.
func TestRelationListCommand_TableOutput(t *testing.T) {
	issue1 := makeIssue("ENG-1", "Base issue", "In Progress", "High", "")
	issue2 := makeIssue("ENG-2", "Blocked issue", "Todo", "Medium", "")
	issue3 := makeIssue("ENG-3", "Incoming issue", "Done", "Low", "")

	outgoing := makeRelation("rel-1", "blocks", issue1, issue2)
	incoming := makeRelation("rel-2", "related", issue3, issue1)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, relationListResponse(
			[]map[string]any{outgoing},
			[]map[string]any{incoming},
		))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "list", "ENG-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, col := range []string{"TYPE", "DIRECTION", "RELATED_ISSUE", "TITLE"} {
		if !strings.Contains(result, col) {
			t.Errorf("output should contain %s column header, got:\n%s", col, result)
		}
	}
	if !strings.Contains(result, "blocks") {
		t.Errorf("output should contain relation type 'blocks', got:\n%s", result)
	}
	if !strings.Contains(result, "outgoing") {
		t.Errorf("output should contain direction 'outgoing', got:\n%s", result)
	}
	if !strings.Contains(result, "ENG-2") {
		t.Errorf("output should contain related issue identifier ENG-2, got:\n%s", result)
	}
	if !strings.Contains(result, "incoming") {
		t.Errorf("output should contain direction 'incoming', got:\n%s", result)
	}
	// incoming: the "issue" field is the other issue (ENG-3)
	if !strings.Contains(result, "ENG-3") {
		t.Errorf("output should contain incoming issue identifier ENG-3, got:\n%s", result)
	}
}

// TestRelationListCommand_IssueNotFound verifies error when issue is not found.
func TestRelationListCommand_IssueNotFound(t *testing.T) {
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
	root.SetArgs([]string{"issue", "relation", "list", "ENG-999"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when issue not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

// TestRelationListCommand_MissingIdentifier verifies error when identifier is missing.
func TestRelationListCommand_MissingIdentifier(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when identifier is missing")
	}
	if !strings.Contains(err.Error(), "issue identifier is required") {
		t.Errorf("error should mention issue identifier is required, got: %v", err)
	}
}

// TestRelationListCommand_IssueIdSentInVars verifies correct variable is sent.
func TestRelationListCommand_IssueIdSentInVars(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, relationListResponse(nil, nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "list", "ENG-42"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["issueId"] != "ENG-42" {
		t.Errorf("issueId = %v, want ENG-42", gotVars["issueId"])
	}
}

// TestRelationCreateCommand_Basic verifies correct mutation input is sent.
func TestRelationCreateCommand_Basic(t *testing.T) {
	issue1 := makeIssue("ENG-1", "Issue 1", "Todo", "Medium", "")
	issue2 := makeIssue("ENG-2", "Issue 2", "Todo", "Medium", "")
	rel := makeRelation("new-rel-id", "blocks", issue1, issue2)

	server, bodies := newQueuedServer(t, []map[string]any{
		relationCreateResponse(rel),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "create", "ENG-1", "--related", "ENG-2", "--type", "blocks"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "new-rel-id") {
		t.Errorf("output should contain relation ID, got: %s", result)
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
	if input["relatedIssueId"] != "ENG-2" {
		t.Errorf("relatedIssueId = %v, want ENG-2", input["relatedIssueId"])
	}
	if input["type"] != "blocks" {
		t.Errorf("type = %v, want blocks", input["type"])
	}
}

// TestRelationCreateCommand_DefaultTypeRelated verifies default type is "related".
func TestRelationCreateCommand_DefaultTypeRelated(t *testing.T) {
	issue1 := makeIssue("ENG-1", "Issue 1", "Todo", "Medium", "")
	issue2 := makeIssue("ENG-2", "Issue 2", "Todo", "Medium", "")
	rel := makeRelation("rel-default", "related", issue1, issue2)

	server, bodies := newQueuedServer(t, []map[string]any{
		relationCreateResponse(rel),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "create", "ENG-1", "--related", "ENG-2"})

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
	if input["type"] != "related" {
		t.Errorf("default type = %v, want related", input["type"])
	}
}

// TestRelationCreateCommand_MissingRelated verifies error when --related is missing.
func TestRelationCreateCommand_MissingRelated(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "create", "ENG-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --related is missing")
	}
	if !strings.Contains(err.Error(), "related") {
		t.Errorf("error should mention related, got: %v", err)
	}
}

// TestRelationCreateCommand_SuccessFalse verifies error when mutation returns success=false.
func TestRelationCreateCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"issueRelationCreate": map[string]any{
					"success":       false,
					"issueRelation": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "create", "ENG-1", "--related", "ENG-2"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

// TestRelationDeleteCommand_WithYes verifies delete sends correct mutation with --yes flag.
func TestRelationDeleteCommand_WithYes(t *testing.T) {
	const relID = "rel-uuid-1234-5678-90ab-cdef01234567"

	server, bodies := newQueuedServer(t, []map[string]any{
		relationDeleteResponse(relID),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "delete", relID, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "deleted") {
		t.Errorf("output should mention deleted, got: %s", result)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != relID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], relID)
	}
}

// TestRelationDeleteCommand_ConfirmAbort verifies abort when user declines confirmation.
func TestRelationDeleteCommand_ConfirmAbort(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"issue", "relation", "delete", "rel-uuid-1234-5678-90ab-cdef01234567"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when user aborts")
	}
	if !strings.Contains(err.Error(), "aborted") {
		t.Errorf("error should mention aborted, got: %v", err)
	}
}

// TestRelationDeleteCommand_SuccessFalse verifies error when mutation returns success=false.
func TestRelationDeleteCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"issueRelationDelete": map[string]any{
					"success":  false,
					"entityId": "",
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "delete", "rel-id", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

// TestRelationListCommand_ExtraArgs verifies error when too many arguments are provided.
func TestRelationListCommand_ExtraArgs(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "list", "ENG-1", "ENG-2"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments provided")
	}
	if !strings.Contains(err.Error(), "accepts 1 argument") {
		t.Errorf("error should mention accepts 1 argument, got: %v", err)
	}
}

// TestRelationDeleteCommand_ExtraArgs verifies error when too many arguments are provided.
func TestRelationDeleteCommand_ExtraArgs(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"issue", "relation", "delete", "rel-1", "rel-2", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when too many arguments provided")
	}
	if !strings.Contains(err.Error(), "accepts 1 argument") {
		t.Errorf("error should mention accepts 1 argument, got: %v", err)
	}
}

// TestRelationListCommand_JSONOutput verifies JSON output for relation list.
func TestRelationListCommand_JSONOutput(t *testing.T) {
	issue1 := makeIssue("ENG-1", "Issue 1", "Todo", "Medium", "")
	issue2 := makeIssue("ENG-2", "Issue 2", "Done", "Low", "")
	outgoing := makeRelation("rel-json-1", "duplicate", issue1, issue2)

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, relationListResponse(
			[]map[string]any{outgoing},
			nil,
		))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "issue", "relation", "list", "ENG-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(decoded))
	}
}
