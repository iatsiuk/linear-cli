package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/iatsiuk/linear-cli/internal/cmd"
)

func makeAttachment(id, title, url string) map[string]any {
	return map[string]any{
		"id":        id,
		"title":     title,
		"url":       url,
		"createdAt": "2026-01-01T10:00:00Z",
		"updatedAt": "2026-01-02T12:00:00Z",
	}
}

func attachmentListResponse(identifier string, attachments []map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"issue": map[string]any{
				"id":         "issue-uuid-" + identifier,
				"identifier": identifier,
				"attachments": map[string]any{
					"nodes":    attachments,
					"pageInfo": map[string]any{"hasNextPage": false, "endCursor": nil},
				},
			},
		},
	}
}

func attachmentCreateResponse(attachment map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"attachmentCreate": map[string]any{
				"success":    true,
				"attachment": attachment,
			},
		},
	}
}

func attachmentDeleteResponse(success bool) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"attachmentDelete": map[string]any{
				"success": success,
			},
		},
	}
}

// TestAttachmentListCommand_TableOutput verifies table output for attachment list.
func TestAttachmentListCommand_TableOutput(t *testing.T) {
	attachments := []map[string]any{
		makeAttachment("att-1", "Screenshot", "https://uploads.linear.app/screenshot.png"),
		makeAttachment("att-2", "PR Link", "https://github.com/org/repo/pull/42"),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, attachmentListResponse("ENG-1", attachments))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "list", "ENG-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, col := range []string{"TITLE", "URL", "CREATED"} {
		if !strings.Contains(result, col) {
			t.Errorf("output should contain %s column header, got:\n%s", col, result)
		}
	}
	if !strings.Contains(result, "Screenshot") {
		t.Errorf("output should contain attachment title, got:\n%s", result)
	}
	if !strings.Contains(result, "PR Link") {
		t.Errorf("output should contain second title, got:\n%s", result)
	}
}

// TestAttachmentListCommand_JSONOutput verifies JSON output for attachment list.
func TestAttachmentListCommand_JSONOutput(t *testing.T) {
	attachments := []map[string]any{
		makeAttachment("att-1", "Screenshot", "https://example.com/shot.png"),
	}

	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSONResponse(w, attachmentListResponse("ENG-2", attachments))
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "attachment", "list", "ENG-2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded []map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 attachment, got %d", len(decoded))
	}
	if decoded[0]["title"] != "Screenshot" {
		t.Errorf("expected title Screenshot, got %v", decoded[0]["title"])
	}
}

// TestAttachmentListCommand_IssueNotFound verifies error when issue is not found.
func TestAttachmentListCommand_IssueNotFound(t *testing.T) {
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
	root.SetArgs([]string{"attachment", "list", "ENG-999"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when issue not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

// TestAttachmentListCommand_MissingIdentifier verifies error when identifier is missing.
func TestAttachmentListCommand_MissingIdentifier(t *testing.T) {
	server := newIssueTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "list"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when identifier is missing")
	}
}

// TestAttachmentCreateCommand_Basic verifies that create sends url, title, and issueId.
func TestAttachmentCreateCommand_Basic(t *testing.T) {
	att := makeAttachment("new-att-id", "My Link", "https://example.com/link")

	server, bodies := newQueuedServer(t, []map[string]any{
		attachmentCreateResponse(att),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "create", "ENG-10", "--url", "https://example.com/link", "--title", "My Link"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "new-att-id") {
		t.Errorf("output should contain attachment ID, got: %s", result)
	}

	if len(*bodies) < 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	input, ok := (*bodies)[0]["input"].(map[string]any)
	if !ok {
		t.Fatalf("input not set: %v", (*bodies)[0])
	}
	if input["issueId"] != "ENG-10" {
		t.Errorf("issueId = %v, want ENG-10", input["issueId"])
	}
	if input["url"] != "https://example.com/link" {
		t.Errorf("url = %v, want https://example.com/link", input["url"])
	}
	if input["title"] != "My Link" {
		t.Errorf("title = %v, want My Link", input["title"])
	}
}

// TestAttachmentCreateCommand_MissingURL verifies error when --url is not provided.
func TestAttachmentCreateCommand_MissingURL(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "create", "ENG-1", "--title", "My Link"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --url is missing")
	}
	if !strings.Contains(err.Error(), "url") {
		t.Errorf("error should mention url, got: %v", err)
	}
}

// TestAttachmentCreateCommand_MissingTitle verifies error when --title is not provided.
func TestAttachmentCreateCommand_MissingTitle(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "create", "ENG-1", "--url", "https://example.com"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --title is missing")
	}
	if !strings.Contains(err.Error(), "title") {
		t.Errorf("error should mention title, got: %v", err)
	}
}

// TestAttachmentCreateCommand_JSONOutput verifies JSON output for attachment create.
func TestAttachmentCreateCommand_JSONOutput(t *testing.T) {
	att := makeAttachment("json-att-id", "Link", "https://example.com")

	server, _ := newQueuedServer(t, []map[string]any{
		attachmentCreateResponse(att),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "attachment", "create", "ENG-5", "--url", "https://example.com", "--title", "Link"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["id"] != "json-att-id" {
		t.Errorf("expected id 'json-att-id', got %v", decoded["id"])
	}
}

// TestAttachmentDeleteCommand_Basic verifies that delete sends the attachment ID.
func TestAttachmentDeleteCommand_Basic(t *testing.T) {
	const attID = "00000000-0000-0000-0000-000000000020"

	server, bodies := newQueuedServer(t, []map[string]any{
		attachmentDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "delete", attID, "--yes"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(*bodies) != 1 {
		t.Fatalf("expected 1 request, got %d", len(*bodies))
	}
	if (*bodies)[0]["id"] != attID {
		t.Errorf("mutation id = %v, want %q", (*bodies)[0]["id"], attID)
	}

	result := out.String()
	if !strings.Contains(result, "deleted") {
		t.Errorf("output should mention deleted, got: %s", result)
	}
}

// TestAttachmentDeleteCommand_SuccessFalse verifies error when mutation returns success=false.
func TestAttachmentDeleteCommand_SuccessFalse(t *testing.T) {
	const attID = "00000000-0000-0000-0000-000000000021"

	server, _ := newQueuedServer(t, []map[string]any{
		attachmentDeleteResponse(false),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "delete", attID, "--yes"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when success=false")
	}
	if !strings.Contains(err.Error(), "success=false") {
		t.Errorf("error should mention success=false, got: %v", err)
	}
}

// TestAttachmentDeleteCommand_ConfirmationAbort verifies abort when user declines.
func TestAttachmentDeleteCommand_ConfirmationAbort(t *testing.T) {
	const attID = "00000000-0000-0000-0000-000000000022"

	server, bodies := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("n\n"))
	root.SetArgs([]string{"attachment", "delete", attID})

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

// TestAttachmentDeleteCommand_ConfirmationAccept verifies delete proceeds when user confirms.
func TestAttachmentDeleteCommand_ConfirmationAccept(t *testing.T) {
	const attID = "00000000-0000-0000-0000-000000000023"

	server, _ := newQueuedServer(t, []map[string]any{
		attachmentDeleteResponse(true),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader("y\n"))
	root.SetArgs([]string{"attachment", "delete", attID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "deleted") {
		t.Errorf("output should mention deleted, got: %s", result)
	}
}

// TestAttachmentDeleteCommand_MissingID verifies error when ID is missing.
func TestAttachmentDeleteCommand_MissingID(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "delete"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when ID is missing")
	}
}

func attachmentShowResponse(attachment map[string]any) map[string]any {
	return map[string]any{
		"data": map[string]any{
			"attachment": attachment,
		},
	}
}

// TestAttachmentShow_Success verifies table output for attachment show.
func TestAttachmentShow_Success(t *testing.T) {
	att := makeAttachment("att-show-1", "Screenshot link", "https://cdn.linear.app/screenshot.png")
	att["creator"] = map[string]any{"id": "u1", "displayName": "aleksei.i", "email": "aleksei@example.com"}

	server, _ := newQueuedServer(t, []map[string]any{
		attachmentShowResponse(att),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "show", "att-show-1"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	for _, want := range []string{"Screenshot link", "https://cdn.linear.app/screenshot.png", "aleksei.i", "2026-01-01", "2026-01-02"} {
		if !strings.Contains(result, want) {
			t.Errorf("output should contain %q, got:\n%s", want, result)
		}
	}
}

// TestAttachmentShow_JSON verifies JSON output for attachment show.
func TestAttachmentShow_JSON(t *testing.T) {
	att := makeAttachment("att-show-2", "PR Link", "https://github.com/org/repo/pull/1")

	server, _ := newQueuedServer(t, []map[string]any{
		attachmentShowResponse(att),
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--json", "attachment", "show", "att-show-2"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, out.String())
	}
	if decoded["id"] != "att-show-2" {
		t.Errorf("expected id att-show-2, got %v", decoded["id"])
	}
	if decoded["title"] != "PR Link" {
		t.Errorf("expected title PR Link, got %v", decoded["title"])
	}
}

// TestAttachmentShow_NotFound verifies error when attachment is not found.
func TestAttachmentShow_NotFound(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{"data": map[string]any{"attachment": nil}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "show", "nonexistent-id"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when attachment not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

// newUploadServer creates a PUT server and a queued GraphQL server whose
// fileUpload response points to the PUT server. Returns both servers and
// a flag indicating whether a PUT was received.
func newUploadServer(t *testing.T, assetURL string, putStatus int, extraResponses []map[string]any) (*httptest.Server, *httptest.Server, *atomic.Bool, *[]map[string]any) {
	t.Helper()

	var putReceived atomic.Bool
	putServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			putReceived.Store(true)
		}
		w.WriteHeader(putStatus)
	}))
	t.Cleanup(putServer.Close)

	uploadResp := map[string]any{
		"data": map[string]any{
			"fileUpload": map[string]any{
				"success": true,
				"uploadFile": map[string]any{
					"assetUrl":  assetURL,
					"uploadUrl": putServer.URL + "/upload",
					"headers":   []map[string]any{},
				},
			},
		},
	}

	allResponses := append([]map[string]any{uploadResp}, extraResponses...)
	bodies := &[]map[string]any{}
	var mu sync.Mutex
	idx := 0
	gqlServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		defer mu.Unlock()
		*bodies = append(*bodies, body.Variables)
		if idx >= len(allResponses) {
			t.Errorf("unexpected request %d", idx+1)
			http.Error(w, "too many requests", 500)
			return
		}
		resp := allResponses[idx]
		idx++
		writeJSONResponse(w, resp)
	}))
	t.Cleanup(gqlServer.Close)

	return gqlServer, putServer, &putReceived, bodies
}

// TestAttachmentCreateCommand_WithFile verifies two-step upload: fileUpload mutation + PUT + attachmentCreate.
func TestAttachmentCreateCommand_WithFile(t *testing.T) {
	const assetURL = "https://cdn.linear.app/org/screenshot.png"
	att := makeAttachment("upload-att-id", "screenshot.png", assetURL)

	gqlServer, _, putReceived, bodies := newUploadServer(t, assetURL, http.StatusOK, []map[string]any{
		attachmentCreateResponse(att),
	})
	setupIssueTest(t, gqlServer)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "screenshot.png")
	if err := os.WriteFile(filePath, []byte("fake png"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "create", "ENG-5", "--file", filePath, "--title", "screenshot.png"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !putReceived.Load() {
		t.Error("expected PUT to upload server, but none received")
	}

	result := out.String()
	if !strings.Contains(result, "upload-att-id") {
		t.Errorf("output should contain attachment ID, got: %s", result)
	}

	// verify attachmentCreate received the assetURL
	if len(*bodies) < 2 {
		t.Fatalf("expected 2 requests (fileUpload + attachmentCreate), got %d", len(*bodies))
	}
	input, ok := (*bodies)[1]["input"].(map[string]any)
	if !ok {
		t.Fatalf("attachmentCreate input not set: %v", (*bodies)[1])
	}
	if input["url"] != assetURL {
		t.Errorf("attachmentCreate url = %v, want %q", input["url"], assetURL)
	}
}

// TestAttachmentCreateCommand_FileNotFound verifies error when --file points to missing file.
func TestAttachmentCreateCommand_FileNotFound(t *testing.T) {
	gqlServer, _, _, _ := newUploadServer(t, "", http.StatusOK, nil)
	setupIssueTest(t, gqlServer)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "create", "ENG-1", "--file", "/nonexistent/file.png", "--title", "Nope"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// TestAttachmentCreateCommand_UploadFailure verifies error when PUT returns non-2xx.
func TestAttachmentCreateCommand_UploadFailure(t *testing.T) {
	gqlServer, _, _, _ := newUploadServer(t, "https://cdn.linear.app/x.png", http.StatusForbidden, nil)
	setupIssueTest(t, gqlServer)

	dir := t.TempDir()
	filePath := filepath.Join(dir, "image.png")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "create", "ENG-1", "--file", filePath, "--title", "Image"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when upload returns 403, got nil")
	}
}

// TestAttachmentCreateCommand_MutuallyExclusive verifies error when both --url and --file are provided.
func TestAttachmentCreateCommand_MutuallyExclusive(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "create", "ENG-1", "--url", "https://example.com", "--file", "/tmp/f.png", "--title", "T"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when both --url and --file are provided")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention mutually exclusive, got: %v", err)
	}
}

// TestAttachmentCreateCommand_NeitherURLNorFile verifies error when neither --url nor --file is provided.
func TestAttachmentCreateCommand_NeitherURLNorFile(t *testing.T) {
	server, _ := newQueuedServer(t, nil)
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "create", "ENG-1", "--title", "T"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when neither --url nor --file is provided")
	}
}

// newDownloadTestServersCapture creates a file server that captures request headers.
// Returns GraphQL server, file server, and a pointer to the captured Authorization header value.
func newDownloadTestServersCapture(t *testing.T, attID, filename string, fileContent []byte) (*httptest.Server, *httptest.Server, *string) {
	t.Helper()

	var capturedAuth string
	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(fileContent)
	}))
	t.Cleanup(fileServer.Close)

	fileURL := fileServer.URL + "/files/" + filename
	att := makeAttachment(attID, "Test Attachment", fileURL)
	gqlServer, _ := newQueuedServer(t, []map[string]any{
		attachmentShowResponse(att),
	})

	return gqlServer, fileServer, &capturedAuth
}

// TestAttachmentDownload_AuthHeaderNoBearerPrefix verifies Authorization header is set to the API key with no Bearer prefix
// when the attachment URL is from a trusted linear.app host.
func TestAttachmentDownload_AuthHeaderNoBearerPrefix(t *testing.T) {
	const attID = "dl-auth-2"
	content := []byte("secure content")

	// override trusted host suffix so the local test server is treated as trusted
	orig := cmd.TrustedDownloadHostSuffix
	cmd.TrustedDownloadHostSuffix = "127.0.0.1"
	t.Cleanup(func() { cmd.TrustedDownloadHostSuffix = orig })

	gqlServer, _, capturedAuth := newDownloadTestServersCapture(t, attID, "secret2.pdf", content)
	setupIssueTest(t, gqlServer)

	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "download", attID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	auth := *capturedAuth
	if auth == "" {
		t.Fatal("download request should include Authorization header, got empty")
	}
	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		t.Errorf("Authorization header should not have Bearer prefix, got: %s", auth)
	}
	// verify it's the API key value directly
	if auth != "lin_api_testkey" {
		t.Errorf("Authorization header should be API key, got: %s", auth)
	}
}

// TestAttachmentDownload_NoAuthForThirdPartyURL verifies that the Authorization header is NOT sent
// to non-linear.app hosts (third-party attachment URLs) to prevent API key leakage.
func TestAttachmentDownload_NoAuthForThirdPartyURL(t *testing.T) {
	const attID = "dl-auth-3"
	content := []byte("public content")

	// TrustedDownloadHostSuffix is "linear.app" by default; test server is at 127.0.0.1 which won't match
	gqlServer, _, capturedAuth := newDownloadTestServersCapture(t, attID, "public.pdf", content)
	setupIssueTest(t, gqlServer)

	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "download", attID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *capturedAuth != "" {
		t.Errorf("Authorization header should NOT be sent to non-linear.app hosts, got: %s", *capturedAuth)
	}
}

// newDownloadTestServers creates a GraphQL server and a file server for download tests.
// The GraphQL server returns an attachment whose URL points to the file server.
func newDownloadTestServers(t *testing.T, attID, filename string, fileContent []byte, fileStatus int) (*httptest.Server, *httptest.Server) {
	t.Helper()

	fileServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(fileStatus)
		if fileStatus >= 200 && fileStatus < 300 {
			_, _ = w.Write(fileContent)
		}
	}))
	t.Cleanup(fileServer.Close)

	fileURL := fileServer.URL + "/files/" + filename

	att := makeAttachment(attID, "Test Attachment", fileURL)
	gqlServer, _ := newQueuedServer(t, []map[string]any{
		attachmentShowResponse(att),
	})

	return gqlServer, fileServer
}

// TestAttachmentDownload_Success verifies file is saved to current directory with filename from URL.
func TestAttachmentDownload_Success(t *testing.T) {
	const attID = "dl-att-1"
	content := []byte("hello file content")

	gqlServer, _ := newDownloadTestServers(t, attID, "report.pdf", content, http.StatusOK)
	setupIssueTest(t, gqlServer)

	dir := t.TempDir()
	// run from the temp dir so the default filename lands there
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "download", attID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result := out.String()
	if !strings.Contains(result, "report.pdf") {
		t.Errorf("output should mention filename, got: %s", result)
	}

	saved, err := os.ReadFile(filepath.Join(dir, "report.pdf"))
	if err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
	if string(saved) != string(content) {
		t.Errorf("file content mismatch: got %q, want %q", saved, content)
	}
}

// TestAttachmentDownload_CustomOutput verifies --output flag saves to specified path.
func TestAttachmentDownload_CustomOutput(t *testing.T) {
	const attID = "dl-att-2"
	content := []byte("custom output content")

	gqlServer, _ := newDownloadTestServers(t, attID, "original.png", content, http.StatusOK)
	setupIssueTest(t, gqlServer)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "custom-name.png")

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "download", attID, "--output", outPath})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	saved, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("saved file not found: %v", err)
	}
	if string(saved) != string(content) {
		t.Errorf("file content mismatch: got %q, want %q", saved, content)
	}
}

// TestAttachmentDownload_StdoutDash verifies --output - writes to stdout.
func TestAttachmentDownload_StdoutDash(t *testing.T) {
	const attID = "dl-att-3"
	content := []byte("stdout content here")

	gqlServer, _ := newDownloadTestServers(t, attID, "file.txt", content, http.StatusOK)
	setupIssueTest(t, gqlServer)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "download", attID, "--output", "-"})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.String() != string(content) {
		t.Errorf("stdout content mismatch: got %q, want %q", out.String(), content)
	}
}

// TestAttachmentDownload_NotFound verifies error when attachment not found.
func TestAttachmentDownload_NotFound(t *testing.T) {
	server, _ := newQueuedServer(t, []map[string]any{
		{"data": map[string]any{"attachment": nil}},
	})
	setupIssueTest(t, server)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "download", "nonexistent-id"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when attachment not found")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found, got: %v", err)
	}
}

// TestAttachmentDownload_HTTPError verifies error when file URL returns non-2xx.
func TestAttachmentDownload_HTTPError(t *testing.T) {
	const attID = "dl-att-4"

	gqlServer, _ := newDownloadTestServers(t, attID, "file.bin", nil, http.StatusForbidden)
	setupIssueTest(t, gqlServer)

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "download", attID, "--output", "-"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when file URL returns 403")
	}
}

// TestAttachmentDownload_FilenameFromURL verifies default filename is derived from URL path.
func TestAttachmentDownload_FilenameFromURL(t *testing.T) {
	const attID = "dl-att-5"
	content := []byte("data")

	gqlServer, _ := newDownloadTestServers(t, attID, "myimage.jpg", content, http.StatusOK)
	setupIssueTest(t, gqlServer)

	dir := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	var out bytes.Buffer
	root := cmd.NewRootCommand("test")
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"attachment", "download", attID})

	if err := root.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, statErr := os.Stat(filepath.Join(dir, "myimage.jpg")); statErr != nil {
		t.Errorf("expected file myimage.jpg to be saved: %v", statErr)
	}
}
