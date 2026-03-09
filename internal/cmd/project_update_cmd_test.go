package cmd_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func makeProjectUpdateCheckin(id, health, authorName, createdAt, body string) map[string]any {
	return map[string]any{
		"id":     id,
		"body":   body,
		"health": health,
		"user": map[string]any{
			"id":          "user-1",
			"displayName": authorName,
			"email":       authorName + "@example.com",
		},
		"project": map[string]any{
			"id":   "proj-1",
			"name": "My Project",
		},
		"createdAt": createdAt,
		"updatedAt": createdAt,
	}
}

func projectUpdateCheckinListResponse(updates []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"project": map[string]any{
				"projectUpdates": map[string]any{
					"nodes": updates,
				},
			},
		},
	}
}

func projectUpdateCheckinCreateResponse(update map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projectUpdateCreate": map[string]any{
				"success":       true,
				"projectUpdate": update,
			},
		},
	}
}

func projectUpdateCheckinArchiveResponse() map[string]any {
	return map[string]any{
		"data": map[string]any{
			"projectUpdateArchive": map[string]any{
				"success": true,
			},
		},
	}
}

func TestProjectUpdateCheckinListCommand_Basic(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	updates := []map[string]any{
		makeProjectUpdateCheckin("pu-1", "onTrack", "Alice", "2026-03-01T00:00:00Z", "Going well"),
		makeProjectUpdateCheckin("pu-2", "atRisk", "Bob", "2026-02-01T00:00:00Z", "Slight delay"),
	}

	server, bodies := newQueuedServer(t, []map[string]any{
		projectUpdateCheckinListResponse(updates),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "list", projID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	vars := (*bodies)[0]
	if vars["projectId"] != projID {
		t.Errorf("projectId = %v, want %s", vars["projectId"], projID)
	}

	result := out.String()
	if !strings.Contains(result, "onTrack") {
		t.Errorf("output should contain onTrack, got: %s", result)
	}
	if !strings.Contains(result, "Alice") {
		t.Errorf("output should contain Alice, got: %s", result)
	}
}

func TestProjectUpdateCheckinListCommand_JSONOutput(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	updates := []map[string]any{
		makeProjectUpdateCheckin("pu-1", "onTrack", "Alice", "2026-03-01T00:00:00Z", "Going well"),
	}

	server, _ := newQueuedServer(t, []map[string]any{
		projectUpdateCheckinListResponse(updates),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "project", "update", "list", projID})

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
	if decoded[0]["health"] != "onTrack" {
		t.Errorf("expected health onTrack, got %v", decoded[0]["health"])
	}
}

func TestProjectUpdateCheckinListCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when project id is missing")
	}
}

func TestProjectUpdateCheckinCreateCommand_Basic(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	update := makeProjectUpdateCheckin("pu-new", "onTrack", "Alice", "2026-03-09T00:00:00Z", "Everything is great")

	server, bodies := newQueuedServer(t, []map[string]any{
		projectUpdateCheckinCreateResponse(update),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "create", projID, "--body", "Everything is great"})

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
	if input["projectId"] != projID {
		t.Errorf("projectId = %v, want %s", input["projectId"], projID)
	}
	if input["body"] != "Everything is great" {
		t.Errorf("body = %v, want 'Everything is great'", input["body"])
	}
	if _, present := input["health"]; present {
		t.Errorf("health should not be in input when not provided")
	}
}

func TestProjectUpdateCheckinCreateCommand_WithHealth(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000002"
	update := makeProjectUpdateCheckin("pu-2", "atRisk", "Bob", "2026-03-09T00:00:00Z", "Some delays")

	server, bodies := newQueuedServer(t, []map[string]any{
		projectUpdateCheckinCreateResponse(update),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "create", projID, "--body", "Some delays", "--health", "atRisk"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	input := (*bodies)[0]["input"].(map[string]any)
	if input["health"] != "atRisk" {
		t.Errorf("health = %v, want atRisk", input["health"])
	}
}

func TestProjectUpdateCheckinCreateCommand_MissingBody(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "create", projID})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when body is missing")
	}
}

func TestProjectUpdateCheckinCreateCommand_SuccessFalse(t *testing.T) {
	const projID = "00000000-0000-0000-0000-000000000001"
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"projectUpdateCreate": map[string]any{
					"success":       false,
					"projectUpdate": nil,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "create", projID, "--body", "test"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

func TestProjectUpdateCheckinArchiveCommand_Basic(t *testing.T) {
	const checkinID = "pu-archive-1"

	server, bodies := newQueuedServer(t, []map[string]any{
		projectUpdateCheckinArchiveResponse(),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "archive", checkinID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != checkinID {
		t.Errorf("id = %v, want %s", (*bodies)[0]["id"], checkinID)
	}

	if !strings.Contains(out.String(), "archived") {
		t.Errorf("output should contain 'archived', got: %s", out.String())
	}
}

func TestProjectUpdateCheckinArchiveCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "archive"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when check-in id is missing")
	}
}

func TestProjectUpdateCheckinArchiveCommand_SuccessFalse(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{
			"data": map[string]any{
				"projectUpdateArchive": map[string]any{
					"success": false,
				},
			},
		},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"project", "update", "archive", "pu-1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}
