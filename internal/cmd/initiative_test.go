package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func makeInitiative(id, name, status string, description *string) map[string]any {
	m := map[string]any{
		"id":     id,
		"name":   name,
		"status": status,
	}
	if description != nil {
		m["description"] = *description
	}
	return m
}

func initiativeListResponse(initiatives []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"initiatives": map[string]any{
				"nodes": initiatives,
			},
		},
	}
}

func initiativeShowResponse(initiative map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"initiative": initiative,
		},
	}
}

func initiativeCreateResponse(initiative map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"initiativeCreate": map[string]any{
				"success":    true,
				"initiative": initiative,
			},
		},
	}
}

func initiativeUpdateResponse(initiative map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"initiativeUpdate": map[string]any{
				"success":    true,
				"initiative": initiative,
			},
		},
	}
}

func initiativeDeleteResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"initiativeDelete": map[string]any{
				"success": success,
			},
		},
	}
}

func TestInitiativeListCommand_Basic(t *testing.T) {
	initiatives := []map[string]any{
		makeInitiative("ini-1", "Q1 Goals", "Active", strPtr("First quarter")),
		makeInitiative("ini-2", "Platform Upgrade", "Planned", nil),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		initiativeListResponse(initiatives),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "Q1 Goals") {
		t.Errorf("output should contain initiative name, got: %s", result)
	}
	if !strings.Contains(result, "Platform Upgrade") {
		t.Errorf("output should contain second initiative name, got: %s", result)
	}
	if !strings.Contains(result, "Active") {
		t.Errorf("output should contain status, got: %s", result)
	}
}

func TestInitiativeListCommand_Empty(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		initiativeListResponse([]map[string]any{}),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInitiativeListCommand_JSONOutput(t *testing.T) {
	initiatives := []map[string]any{
		makeInitiative("ini-1", "Q1 Goals", "Active", nil),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		initiativeListResponse(initiatives),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "initiative", "list"})

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
	if decoded[0]["name"] != "Q1 Goals" {
		t.Errorf("expected name 'Q1 Goals', got %v", decoded[0]["name"])
	}
}

func TestInitiativeShowCommand_Basic(t *testing.T) {
	desc := "First quarter goals"
	ini := makeInitiative("ini-1", "Q1 Goals", "Active", &desc)

	server, bodies := newQueuedServer(t, []map[string]any{
		initiativeShowResponse(ini),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "show", "ini-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != "ini-1" {
		t.Errorf("id = %v, want ini-1", (*bodies)[0]["id"])
	}

	result := out.String()
	if !strings.Contains(result, "Q1 Goals") {
		t.Errorf("output should contain name, got: %s", result)
	}
	if !strings.Contains(result, "Active") {
		t.Errorf("output should contain status, got: %s", result)
	}
	if !strings.Contains(result, desc) {
		t.Errorf("output should contain description, got: %s", result)
	}
}

func TestInitiativeShowCommand_JSONOutput(t *testing.T) {
	ini := makeInitiative("ini-2", "Platform Upgrade", "Planned", nil)

	server, _ := newQueuedServer(t, []map[string]any{
		initiativeShowResponse(ini),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "initiative", "show", "ini-2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["name"] != "Platform Upgrade" {
		t.Errorf("expected name 'Platform Upgrade', got %v", decoded["name"])
	}
}

func TestInitiativeShowCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "show"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when initiative id is missing")
	}
}

func TestInitiativeCreateCommand_Basic(t *testing.T) {
	ini := makeInitiative("ini-new", "New Initiative", "Planned", nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		initiativeCreateResponse(ini),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "create", "--name", "New Initiative"})

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
	if input["name"] != "New Initiative" {
		t.Errorf("name = %v, want 'New Initiative'", input["name"])
	}
	if _, present := input["description"]; present {
		t.Errorf("description should not be in input when not provided")
	}
}

func TestInitiativeCreateCommand_WithDescription(t *testing.T) {
	ini := makeInitiative("ini-2", "Upgrade Plan", "Planned", strPtr("Some description"))

	server, bodies := newQueuedServer(t, []map[string]any{
		initiativeCreateResponse(ini),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "create", "--name", "Upgrade Plan", "--description", "Some description"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := (*bodies)[0]["input"].(map[string]any)
	if input["description"] != "Some description" {
		t.Errorf("description = %v, want 'Some description'", input["description"])
	}
}

func TestInitiativeCreateCommand_MissingName(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "create"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when name is missing")
	}
}

func TestInitiativeCreateCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"initiativeCreate": map[string]any{
					"success":    false,
					"initiative": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "create", "--name", "test"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestInitiativeUpdateCommand_Basic(t *testing.T) {
	const iniID = "ini-update-1"
	ini := makeInitiative(iniID, "Updated Name", "Active", nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		initiativeUpdateResponse(ini),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "update", iniID, "--name", "Updated Name"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != iniID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], iniID)
	}
	input := (*bodies)[0]["input"].(map[string]any)
	if input["name"] != "Updated Name" {
		t.Errorf("name = %v, want 'Updated Name'", input["name"])
	}
}

func TestInitiativeUpdateCommand_NoFlags(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "update", "ini-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no flags are provided")
	}
	if !strings.Contains(err.Error(), "no fields to update") {
		t.Errorf("error should mention no fields to update, got: %v", err)
	}
}

func TestInitiativeUpdateCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "update"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when initiative id is missing")
	}
}

func TestInitiativeDeleteCommand_Basic(t *testing.T) {
	const iniID = "ini-delete-1"

	server, bodies := newQueuedServer(t, []map[string]any{
		initiativeDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "delete", iniID, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != iniID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], iniID)
	}
	if !strings.Contains(out.String(), "deleted") {
		t.Errorf("output should contain 'deleted', got: %s", out.String())
	}
}

func TestInitiativeDeleteCommand_Aborted(t *testing.T) {
	server, bodies := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"initiative", "delete", "ini-1"})

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

func TestInitiativeDeleteCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "delete"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when initiative id is missing")
	}
}

func TestInitiativeDeleteCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		initiativeDeleteResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"initiative", "delete", "ini-1", "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}
