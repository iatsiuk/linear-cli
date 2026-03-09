package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"linear-cli/internal/cmd"
)

func makeNotification(id, notifType string, readAt *string) map[string]any {
	n := map[string]any{
		"id":         id,
		"type":       notifType,
		"createdAt":  "2026-01-01T00:00:00Z",
		"updatedAt":  "2026-01-01T00:00:00Z",
		"archivedAt": nil,
		"title":      "Test notification",
		"subtitle":   "",
		"url":        "https://linear.app/notify/" + id,
	}
	if readAt != nil {
		n["readAt"] = *readAt
	} else {
		n["readAt"] = nil
	}
	return n
}

func notificationListResponse(notifications []map[string]any) map[string]any {
	if notifications == nil {
		notifications = []map[string]any{}
	}
	return map[string]any{
		"data": map[string]any{
			"notifications": map[string]any{
				"nodes": notifications,
			},
		},
	}
}

func notificationUpdateResponse(notification map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"notificationUpdate": map[string]any{
				"success":      true,
				"notification": notification,
			},
		},
	}
}

func notificationArchiveResponse() map[string]any {
	return map[string]any{
		"data": map[string]any{
			"notificationArchive": map[string]any{
				"success": true,
			},
		},
	}
}

func notificationMarkReadAllResponse() map[string]any {
	return map[string]any{
		"data": map[string]any{
			"notificationMarkReadAll": map[string]any{
				"success": true,
			},
		},
	}
}

func notificationArchiveAllResponse() map[string]any {
	return map[string]any{
		"data": map[string]any{
			"notificationArchiveAll": map[string]any{
				"success": true,
			},
		},
	}
}

// TestNotificationListCommand_TableOutput verifies table columns and data.
func TestNotificationListCommand_TableOutput(t *testing.T) {
	readAt := "2026-01-02T00:00:00Z"
	notifications := []map[string]any{
		makeNotification("notif-1", "issueAssignedToYou", nil),
		makeNotification("notif-2", "issueMention", &readAt),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, notificationListResponse(notifications))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, col := range []string{"ID", "TYPE", "CREATED", "READ"} {
		if !strings.Contains(result, col) {
			t.Errorf("output should contain %s column header, got:\n%s", col, result)
		}
	}
	if !strings.Contains(result, "notif-1") {
		t.Errorf("output should contain notif-1, got:\n%s", result)
	}
	if !strings.Contains(result, "issueAssignedToYou") {
		t.Errorf("output should contain notification type, got:\n%s", result)
	}
}

// TestNotificationListCommand_UnreadFilter verifies --unread filters out read notifications.
func TestNotificationListCommand_UnreadFilter(t *testing.T) {
	readAt := "2026-01-02T00:00:00Z"
	notifications := []map[string]any{
		makeNotification("notif-unread", "issueAssignedToYou", nil),
		makeNotification("notif-read", "issueMention", &readAt),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, notificationListResponse(notifications))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "list", "--unread"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "notif-unread") {
		t.Errorf("output should contain unread notification, got:\n%s", result)
	}
	if strings.Contains(result, "notif-read") {
		t.Errorf("output should not contain read notification, got:\n%s", result)
	}
}

// TestNotificationListCommand_JSONOutput verifies JSON output for notification list.
func TestNotificationListCommand_JSONOutput(t *testing.T) {
	notifications := []map[string]any{
		makeNotification("notif-json-1", "issueAssignedToYou", nil),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, notificationListResponse(notifications))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "notification", "list"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(decoded))
	}
}

// TestNotificationListCommand_LimitSentInVars verifies --limit flag is sent as 'first' variable.
func TestNotificationListCommand_LimitSentInVars(t *testing.T) {
	var gotVars map[string]any
	server := newIssueTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotVars = body.Variables
		writeJSONResponse(w, notificationListResponse(nil))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "list", "--limit", "10"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// JSON numbers decode as float64
	if gotVars["first"] != float64(10) {
		t.Errorf("first = %v, want 10", gotVars["first"])
	}
}

// TestNotificationReadCommand_Single verifies marking a single notification as read.
func TestNotificationReadCommand_Single(t *testing.T) {
	notif := makeNotification("notif-read-1", "issueAssignedToYou", nil)

	server, bodies := newQueuedServer(t, []map[string]any{
		notificationUpdateResponse(notif),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "read", "notif-read-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "marked as read") {
		t.Errorf("output should mention marked as read, got: %s", result)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != "notif-read-1" {
		t.Errorf("id = %v, want notif-read-1", (*bodies)[0]["id"])
	}
	input, ok := (*bodies)[0]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[0])
	}
	if _, hasReadAt := input["readAt"]; !hasReadAt {
		t.Errorf("input should contain readAt, got: %v", input)
	}
}

// TestNotificationReadCommand_All verifies --all flag sends markReadAll mutation.
func TestNotificationReadCommand_All(t *testing.T) {
	server, bodies := newQueuedServer(t, []map[string]any{
		notificationMarkReadAllResponse(),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "read", "--all"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "All notifications marked as read") {
		t.Errorf("output should confirm all read, got: %s", result)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if _, hasReadAt := (*bodies)[0]["readAt"]; !hasReadAt {
		t.Errorf("request should contain readAt, got: %v", (*bodies)[0])
	}
}

// TestNotificationReadCommand_MissingID verifies error when no id and no --all.
func TestNotificationReadCommand_MissingID(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "read"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no id provided")
	}
	if !strings.Contains(err.Error(), "required") && !strings.Contains(err.Error(), "id") && !strings.Contains(err.Error(), "--all") {
		t.Errorf("error should mention id or --all, got: %v", err)
	}
}

// TestNotificationArchiveCommand_Single verifies archiving a single notification.
func TestNotificationArchiveCommand_Single(t *testing.T) {
	server, bodies := newQueuedServer(t, []map[string]any{
		notificationArchiveResponse(),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "archive", "notif-arch-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "archived") {
		t.Errorf("output should mention archived, got: %s", result)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != "notif-arch-1" {
		t.Errorf("id = %v, want notif-arch-1", (*bodies)[0]["id"])
	}
}

// TestNotificationArchiveCommand_All verifies --all flag sends archiveAll mutation.
func TestNotificationArchiveCommand_All(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		notificationArchiveAllResponse(),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "archive", "--all"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "All notifications archived") {
		t.Errorf("output should confirm all archived, got: %s", result)
	}
}

// TestNotificationArchiveCommand_MissingID verifies error when no id and no --all.
func TestNotificationArchiveCommand_MissingID(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"notification", "archive"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no id provided")
	}
}
