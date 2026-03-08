package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestClient creates a test HTTP server and a Client pointed at it.
// The client's sleep is replaced with a no-op to avoid delays in tests.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := NewClient("lin_api_testkey", WithEndpoint(server.URL))
	c.sleep = func(context.Context, time.Duration) error { return nil }
	return server, c
}

// writeJSON encodes resp as JSON to w with Content-Type header.
func writeJSON(w http.ResponseWriter, resp any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func TestClientDo_Success(t *testing.T) {
	t.Parallel()
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"data": map[string]any{"viewer": map[string]any{"id": "user1"}},
		})
	})

	var result struct {
		Viewer struct {
			ID string `json:"id"`
		} `json:"viewer"`
	}
	if err := c.Do(context.Background(), "query { viewer { id } }", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Viewer.ID != "user1" {
		t.Errorf("got ID %q, want %q", result.Viewer.ID, "user1")
	}
}

func TestClientDo_AuthHeader(t *testing.T) {
	t.Parallel()
	var gotAuth string
	_, c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		writeJSON(w, map[string]any{"data": map[string]any{}})
	})

	_ = c.Do(context.Background(), "query { viewer { id } }", nil, nil)
	if gotAuth != "Bearer lin_api_testkey" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer lin_api_testkey")
	}
}

func TestClientDo_GraphQLErrors_Single(t *testing.T) {
	t.Parallel()
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"errors": []map[string]any{
				{"message": "entity not found", "extensions": map[string]any{"code": CodeEntityNotFound}},
			},
		})
	})

	err := c.Do(context.Background(), "query { issue(id: \"x\") { id } }", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var gqlErrs GraphQLErrors
	if !errors.As(err, &gqlErrs) {
		t.Fatalf("expected GraphQLErrors, got %T: %v", err, err)
	}
	if gqlErrs[0].Code() != CodeEntityNotFound {
		t.Errorf("code = %q, want %q", gqlErrs[0].Code(), CodeEntityNotFound)
	}
}

func TestClientDo_GraphQLErrors_Multiple(t *testing.T) {
	t.Parallel()
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"errors": []map[string]any{
				{"message": "error one"},
				{"message": "error two"},
			},
		})
	})

	err := c.Do(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "error one; error two" {
		t.Errorf("Error() = %q, want %q", err.Error(), "error one; error two")
	}
}

func TestClientDo_PartialDataWithErrors(t *testing.T) {
	t.Parallel()
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"data":   map[string]any{"partial": "data"},
			"errors": []map[string]any{{"message": "partial error"}},
		})
	})

	var result map[string]any
	err := c.Do(context.Background(), "query {}", nil, &result)
	if err == nil {
		t.Fatal("expected error for partial response with errors")
	}
	if result != nil {
		t.Errorf("result should remain nil when errors are returned, got: %v", result)
	}
}

func TestClientDo_Unauthorized(t *testing.T) {
	t.Parallel()
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	err := c.Do(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestClientDo_ServerError_Retries(t *testing.T) {
	t.Parallel()
	attempts := 0
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	})

	err := c.Do(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error after all retries")
	}
	// should have retried (maxRetries=3 means 4 total attempts)
	if attempts != 4 {
		t.Errorf("expected 4 attempts (1 + 3 retries), got %d", attempts)
	}
}

func TestClientDo_NetworkError(t *testing.T) {
	t.Parallel()
	// nothing listening on port 1
	c := NewClient("key", WithEndpoint("http://127.0.0.1:1"))
	c.sleep = func(context.Context, time.Duration) error { return nil }

	err := c.Do(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}

func TestClientDo_CustomEndpoint(t *testing.T) {
	t.Parallel()
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		writeJSON(w, map[string]any{"data": map[string]any{}})
	}))
	defer server.Close()

	c := NewClient("key", WithEndpoint(server.URL+"/custom/graphql"))
	c.sleep = func(context.Context, time.Duration) error { return nil }
	_ = c.Do(context.Background(), "query {}", nil, nil)

	if gotPath != "/custom/graphql" {
		t.Errorf("request path = %q, want %q", gotPath, "/custom/graphql")
	}
}

func TestClientDo_MutationPayload(t *testing.T) {
	t.Parallel()
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"issueCreate": map[string]any{
					"success": true,
					"issue":   map[string]any{"id": "issue1"},
				},
			},
		})
	})

	var result struct {
		IssueCreate struct {
			Success bool `json:"success"`
			Issue   struct {
				ID string `json:"id"`
			} `json:"issue"`
		} `json:"issueCreate"`
	}
	if err := c.Do(context.Background(), "mutation {}", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IssueCreate.Success {
		t.Error("expected success=true")
	}
	if result.IssueCreate.Issue.ID != "issue1" {
		t.Errorf("issue ID = %q, want %q", result.IssueCreate.Issue.ID, "issue1")
	}
}

func TestClientDo_Variables(t *testing.T) {
	t.Parallel()
	var gotBody map[string]any
	_, c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		writeJSON(w, map[string]any{"data": map[string]any{}})
	})

	vars := map[string]any{"id": "test-id", "title": "Test"}
	_ = c.Do(context.Background(), "query($id: ID!) { issue(id: $id) { id } }", vars, nil)

	gotVars, ok := gotBody["variables"].(map[string]any)
	if !ok {
		t.Fatal("variables not present in request body")
	}
	if gotVars["id"] != "test-id" {
		t.Errorf("variables.id = %v, want %q", gotVars["id"], "test-id")
	}
}
