package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

// newUploadTestServers creates a GraphQL test server (for the fileUpload mutation)
// and a separate PUT server (for the actual file upload).
// The GraphQL server returns the PUT server URL as uploadUrl.
func newUploadTestServers(t *testing.T, putStatus int, assetUrl string) (gqlServer *httptest.Server, putServer *httptest.Server, putReceived *atomic.Bool) {
	t.Helper()

	var received atomic.Bool
	putServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "expected PUT", http.StatusMethodNotAllowed)
			return
		}
		received.Store(true)
		w.WriteHeader(putStatus)
	}))
	t.Cleanup(putServer.Close)

	uploadUrl := putServer.URL + "/upload"
	gqlServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		resp := map[string]any{
			"data": map[string]any{
				"fileUpload": map[string]any{
					"success": true,
					"uploadFile": map[string]any{
						"assetUrl":  assetUrl,
						"uploadUrl": uploadUrl,
						"headers": []map[string]any{
							{"key": "Content-Type", "value": "image/png"},
						},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	t.Cleanup(gqlServer.Close)

	return gqlServer, putServer, &received
}

func TestUpload_HappyPath(t *testing.T) {
	t.Parallel()

	const assetUrl = "https://cdn.linear.app/test-asset.png"
	gqlServer, _, putReceived := newUploadTestServers(t, http.StatusOK, assetUrl)

	c := NewClient("lin_api_testkey", WithEndpoint(gqlServer.URL))

	// write a temp file
	dir := t.TempDir()
	filePath := filepath.Join(dir, "screenshot.png")
	if err := os.WriteFile(filePath, []byte("fake png data"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	got, err := c.Upload(context.Background(), filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != assetUrl {
		t.Errorf("assetUrl = %q, want %q", got, assetUrl)
	}
	if !putReceived.Load() {
		t.Error("expected PUT request to upload server, but none received")
	}
}

func TestUpload_FileNotFound(t *testing.T) {
	t.Parallel()

	gqlServer, _, _ := newUploadTestServers(t, http.StatusOK, "")

	c := NewClient("lin_api_testkey", WithEndpoint(gqlServer.URL))

	_, err := c.Upload(context.Background(), "/nonexistent/path/file.png")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestUpload_PutFailure(t *testing.T) {
	t.Parallel()

	const assetUrl = "https://cdn.linear.app/test-asset.png"
	gqlServer, _, _ := newUploadTestServers(t, http.StatusForbidden, assetUrl)

	c := NewClient("lin_api_testkey", WithEndpoint(gqlServer.URL))

	dir := t.TempDir()
	filePath := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(filePath, []byte("fake pdf"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := c.Upload(context.Background(), filePath)
	if err == nil {
		t.Fatal("expected error for PUT failure, got nil")
	}
}

func TestUpload_MutationFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"errors": []map[string]any{
				{"message": "forbidden"},
			},
		})
	}))
	t.Cleanup(server.Close)

	c := NewClient("lin_api_testkey", WithEndpoint(server.URL))

	dir := t.TempDir()
	filePath := filepath.Join(dir, "image.png")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	_, err := c.Upload(context.Background(), filePath)
	if err == nil {
		t.Fatal("expected error when mutation fails, got nil")
	}
}

func TestContentTypeFromName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		filename string
		want     string
	}{
		{"image.png", "image/png"},
		{"doc.pdf", "application/pdf"},
		{"noext", "application/octet-stream"},
		{"archive.tar.gz", "application/gzip"},
		{"readme.txt", "text/plain"},
	}
	for _, tc := range cases {
		got := contentTypeFromName(tc.filename)
		if got != tc.want {
			t.Errorf("contentTypeFromName(%q) = %q, want %q", tc.filename, got, tc.want)
		}
	}
}
