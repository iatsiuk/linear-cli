package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

const labelUUID = "00000000-0000-0000-0000-000000000099"

func makeLabel(id, name, color string, isGroup bool, teamKey string) map[string]any {
	l := map[string]any{
		"id":        id,
		"name":      name,
		"color":     color,
		"isGroup":   isGroup,
		"createdAt": "2026-01-01T00:00:00Z",
	}
	if teamKey != "" {
		l["team"] = map[string]any{
			"id":   "team-" + teamKey,
			"name": teamKey + " Team",
			"key":  teamKey,
		}
	}
	return l
}

func labelListResponse(labels []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueLabels": map[string]any{
				"nodes":    labels,
				"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
			},
		},
	}
}

func labelCreateResponse(label map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueLabelCreate": map[string]any{
				"success":    true,
				"issueLabel": label,
			},
		},
	}
}

func labelUpdateResponse(label map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issueLabelUpdate": map[string]any{
				"success":    true,
				"issueLabel": label,
			},
		},
	}
}

// TestLabelListCommand_TableOutput verifies table output for label list.
func TestLabelListCommand_TableOutput(t *testing.T) {
	labels := []map[string]any{
		makeLabel("l1", "Bug", "#FF0000", false, "ENG"),
		makeLabel("l2", "Feature", "#00FF00", false, ""),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, labelListResponse(labels))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, col := range []string{"NAME", "COLOR", "DESCRIPTION", "TEAM", "GROUP"} {
		if !strings.Contains(result, col) {
			t.Errorf("output should contain %s column header, got:\n%s", col, result)
		}
	}
	if !strings.Contains(result, "Bug") {
		t.Errorf("output should contain Bug label, got:\n%s", result)
	}
	if !strings.Contains(result, "Feature") {
		t.Errorf("output should contain Feature label, got:\n%s", result)
	}
	if !strings.Contains(result, "ENG") {
		t.Errorf("output should contain team key ENG, got:\n%s", result)
	}
}

// TestLabelListCommand_TeamFilter verifies that --team sets the filter variable.
func TestLabelListCommand_TeamFilter(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, labelListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "list", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := gotVars["filter"].(map[string]any)
	if !ok {
		t.Fatalf("filter not set, got: %v", gotVars["filter"])
	}
	team, ok := filter["team"].(map[string]any)
	if !ok {
		t.Fatalf("filter.team not set, got: %v", filter["team"])
	}
	key, ok := team["key"].(map[string]any)
	if !ok {
		t.Fatalf("filter.team.key not set, got: %v", team["key"])
	}
	if key["eq"] != "ENG" {
		t.Errorf("filter.team.key.eq = %v, want ENG", key["eq"])
	}
}

// TestLabelListCommand_JSONOutput verifies JSON output for label list.
func TestLabelListCommand_JSONOutput(t *testing.T) {
	labels := []map[string]any{
		makeLabel("l1", "Bug", "#FF0000", false, "ENG"),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, labelListResponse(labels))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "label", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 label, got %d", len(decoded))
	}
	if decoded[0]["name"] != "Bug" {
		t.Errorf("expected name Bug, got %v", decoded[0]["name"])
	}
}

// TestLabelListCommand_NoFilter verifies that no filter variable is set without --team.
func TestLabelListCommand_NoFilter(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, labelListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := gotVars["filter"]; ok {
		t.Errorf("filter should not be set without --team, got: %v", gotVars["filter"])
	}
}

// TestLabelListCommand_IncludeArchived verifies that --include-archived sets the variable.
func TestLabelListCommand_IncludeArchived(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, labelListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "list", "--include-archived"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotVars["includeArchived"] != true {
		t.Errorf("includeArchived should be true, got: %v", gotVars["includeArchived"])
	}
}

// TestLabelCreateCommand_Basic verifies that create sends required fields.
func TestLabelCreateCommand_Basic(t *testing.T) {
	l := makeLabel("l-new", "Urgent", "#FF6600", false, "")

	server, bodies := newQueuedServer(t, []map[string]any{
		labelCreateResponse(l),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "create", "--name", "Urgent", "--color", "#FF6600"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Urgent") {
		t.Errorf("output should contain label name, got: %s", result)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	input, ok := (*bodies)[0]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[0])
	}
	if input["name"] != "Urgent" {
		t.Errorf("name = %v, want Urgent", input["name"])
	}
	if input["color"] != "#FF6600" {
		t.Errorf("color = %v, want #FF6600", input["color"])
	}
}

// TestLabelCreateCommand_MissingName verifies error when --name is missing.
func TestLabelCreateCommand_MissingName(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "create", "--color", "#FF0000"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --name is missing")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention name, got: %v", err)
	}
}

// TestLabelCreateCommand_MissingColor verifies error when --color is missing.
func TestLabelCreateCommand_MissingColor(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "create", "--name", "Test"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --color is missing")
	}
	if !strings.Contains(err.Error(), "color") {
		t.Errorf("error should mention color, got: %v", err)
	}
}

// TestLabelCreateCommand_WithTeam verifies that --team resolves to teamId in input.
func TestLabelCreateCommand_WithTeam(t *testing.T) {
	const teamID = "team-uuid-1234-5678-90ab-cdef01234567"
	l := makeLabel("l-new", "Bug", "#FF0000", false, "ENG")

	server, bodies := newQueuedServer(t, []map[string]any{
		teamResolveResponse(teamID),
		labelCreateResponse(l),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "create", "--name", "Bug", "--color", "#FF0000", "--team", "ENG"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	input, ok := (*bodies)[1]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[1])
	}
	if input["teamId"] != teamID {
		t.Errorf("teamId = %v, want %s", input["teamId"], teamID)
	}
}

// TestLabelCreateCommand_WithDescription verifies optional description field.
func TestLabelCreateCommand_WithDescription(t *testing.T) {
	l := makeLabel("l-new", "Bug", "#FF0000", false, "")

	server, bodies := newQueuedServer(t, []map[string]any{
		labelCreateResponse(l),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "create", "--name", "Bug", "--color", "#FF0000", "--description", "A bug label"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input, ok := (*bodies)[0]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[0])
	}
	if input["description"] != "A bug label" {
		t.Errorf("description = %v, want A bug label", input["description"])
	}
}

// TestLabelCreateCommand_JSONOutput verifies JSON output for label create.
func TestLabelCreateCommand_JSONOutput(t *testing.T) {
	l := makeLabel("l-new", "Bug", "#FF0000", false, "")

	server, _ := newQueuedServer(t, []map[string]any{
		labelCreateResponse(l),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "label", "create", "--name", "Bug", "--color", "#FF0000"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Bug" {
		t.Errorf("expected name Bug, got %v", decoded["name"])
	}
}

// TestLabelCreateCommand_SuccessFalse verifies error when mutation returns success=false.
func TestLabelCreateCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"issueLabelCreate": map[string]any{
					"success":    false,
					"issueLabel": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "create", "--name", "Bug", "--color", "#FF0000"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

// TestLabelUpdateCommand_PartialUpdate verifies that only changed flags are sent.
func TestLabelUpdateCommand_PartialUpdate(t *testing.T) {
	l := makeLabel(labelUUID, "Bug Updated", "#FF0000", false, "")

	server, bodies := newQueuedServer(t, []map[string]any{
		labelResolveResponse(labelUUID),
		labelUpdateResponse(l),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "update", "Bug", "--name", "Bug Updated"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests, got %d", len(*bodies))
	}
	input, ok := (*bodies)[1]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[1])
	}
	if input["name"] != "Bug Updated" {
		t.Errorf("name = %v, want Bug Updated", input["name"])
	}
	if _, present := input["color"]; present {
		t.Errorf("color should not be in input when not provided")
	}
}

// TestLabelUpdateCommand_MultipleFlags verifies that multiple flags are sent.
func TestLabelUpdateCommand_MultipleFlags(t *testing.T) {
	l := makeLabel(labelUUID, "Bug", "#0000FF", false, "")

	server, bodies := newQueuedServer(t, []map[string]any{
		labelResolveResponse(labelUUID),
		labelUpdateResponse(l),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "update", "Bug", "--color", "#0000FF", "--description", "A blue bug"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input, ok := (*bodies)[1]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[1])
	}
	if input["color"] != "#0000FF" {
		t.Errorf("color = %v, want #0000FF", input["color"])
	}
	if input["description"] != "A blue bug" {
		t.Errorf("description = %v, want A blue bug", input["description"])
	}
	if _, present := input["name"]; present {
		t.Errorf("name should not be in input when not provided")
	}
}

// TestLabelUpdateCommand_NoFlags verifies error when no flags are provided.
func TestLabelUpdateCommand_NoFlags(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "update", "Bug"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no flags provided")
	}
	if !strings.Contains(err.Error(), "no fields to update") {
		t.Errorf("error should mention no fields to update, got: %v", err)
	}
}

// TestLabelUpdateCommand_MissingID verifies error when no argument is provided.
func TestLabelUpdateCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "update"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when id is missing")
	}
}

// TestLabelUpdateCommand_JSONOutput verifies JSON output for label update.
func TestLabelUpdateCommand_JSONOutput(t *testing.T) {
	l := makeLabel(labelUUID, "Bug", "#FF0000", false, "")

	server, _ := newQueuedServer(t, []map[string]any{
		labelResolveResponse(labelUUID),
		labelUpdateResponse(l),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "label", "update", "Bug", "--name", "Bug"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Bug" {
		t.Errorf("expected name Bug, got %v", decoded["name"])
	}
}

// TestLabelUpdateCommand_SuccessFalse verifies error when mutation returns success=false.
func TestLabelUpdateCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		labelResolveResponse(labelUUID),
		{
			"data": map[string]any{
				"issueLabelUpdate": map[string]any{
					"success":    false,
					"issueLabel": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"label", "update", "Bug", "--name", "Fail"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}
