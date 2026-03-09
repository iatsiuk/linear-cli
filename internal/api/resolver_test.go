package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeServer(t *testing.T, resp map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
}

// -- ResolveTeamID --

func TestResolveTeamID_UUIDPassthrough(t *testing.T) {
	t.Parallel()
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	got, err := ResolveTeamID(context.Background(), nil, uuid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid {
		t.Errorf("want %q, got %q", uuid, got)
	}
}

func TestResolveTeamID_ByKey(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"teams": map[string]any{
				"nodes": []map[string]any{
					{"id": "team-uuid-1234-5678-90ab-cdef01234567", "key": "ENG"},
				},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveTeamID(context.Background(), c, "ENG")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "team-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
}

func TestResolveTeamID_NotFound(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"teams": map[string]any{
				"nodes": []map[string]any{},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	_, err := ResolveTeamID(context.Background(), c, "NOPE")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// -- ResolveLabelID --

func TestResolveLabelID_UUIDPassthrough(t *testing.T) {
	t.Parallel()
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	got, err := ResolveLabelID(context.Background(), nil, uuid, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid {
		t.Errorf("want %q, got %q", uuid, got)
	}
}

func TestResolveLabelID_ByName(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"issueLabels": map[string]any{
				"nodes": []map[string]any{
					{"id": "label-uuid-1234-5678-90ab-cdef01234567", "name": "Bug"},
				},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveLabelID(context.Background(), c, "Bug", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "label-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
}

func TestResolveLabelID_NotFound(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"issueLabels": map[string]any{
				"nodes": []map[string]any{},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	_, err := ResolveLabelID(context.Background(), c, "NonExistent", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveLabelID_ByNameWithTeam(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"issueLabels": map[string]any{
				"nodes": []map[string]any{
					{"id": "label-uuid-1234-5678-90ab-cdef01234567", "name": "Bug"},
				},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveLabelID(context.Background(), c, "Bug", "team-uuid-1234-5678-90ab-cdef01234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "label-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
}

// -- ResolveUserID --

func TestResolveUserID_UUIDPassthrough(t *testing.T) {
	t.Parallel()
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	got, err := ResolveUserID(context.Background(), nil, uuid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid {
		t.Errorf("want %q, got %q", uuid, got)
	}
}

func TestResolveUserID_ByName(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"users": map[string]any{
				"nodes": []map[string]any{
					{"id": "user-uuid-1234-5678-90ab-cdef01234567", "name": "Alice"},
				},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveUserID(context.Background(), c, "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "user-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
}

func TestResolveUserID_ByDisplayName(t *testing.T) {
	t.Parallel()
	// name lookup returns empty; displayName lookup returns match
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		var resp map[string]any
		if calls == 1 {
			// name lookup - no match
			resp = map[string]any{"data": map[string]any{"users": map[string]any{"nodes": []any{}}}}
		} else {
			// displayName lookup - match
			resp = map[string]any{"data": map[string]any{"users": map[string]any{
				"nodes": []map[string]any{{"id": "user-uuid-1234-5678-90ab-cdef01234567"}},
			}}}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveUserID(context.Background(), c, "alice_display")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "user-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
	if calls != 2 {
		t.Errorf("expected 2 API calls (name then displayName), got %d", calls)
	}
}

func TestResolveUserID_ByEmail(t *testing.T) {
	t.Parallel()
	// name and displayName lookups return empty; email lookup returns match
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		calls++
		var resp map[string]any
		if calls < 3 {
			// name and displayName lookups - no match
			resp = map[string]any{"data": map[string]any{"users": map[string]any{"nodes": []any{}}}}
		} else {
			// email lookup - match
			resp = map[string]any{"data": map[string]any{"users": map[string]any{
				"nodes": []map[string]any{{"id": "user-uuid-1234-5678-90ab-cdef01234567"}},
			}}}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveUserID(context.Background(), c, "alice@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "user-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
	if calls != 3 {
		t.Errorf("expected 3 API calls (name, displayName, email), got %d", calls)
	}
}

func TestResolveUserID_NotFound(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"users": map[string]any{
				"nodes": []map[string]any{},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	_, err := ResolveUserID(context.Background(), c, "nobody@example.com")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// -- ResolveStateID --

func TestResolveStateID_UUIDPassthrough(t *testing.T) {
	t.Parallel()
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	got, err := ResolveStateID(context.Background(), nil, uuid, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid {
		t.Errorf("want %q, got %q", uuid, got)
	}
}

func TestResolveStateID_ByName(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"workflowStates": map[string]any{
				"nodes": []map[string]any{
					{"id": "state-uuid-1234-5678-90ab-cdef01234567", "name": "In Progress"},
				},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveStateID(context.Background(), c, "In Progress", "team-uuid-1234-5678-90ab-cdef01234567")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "state-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
}

func TestResolveStateID_NotFound(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"workflowStates": map[string]any{
				"nodes": []map[string]any{},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	_, err := ResolveStateID(context.Background(), c, "Nonexistent", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestResolveStateID_ByNameNoTeam(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"workflowStates": map[string]any{
				"nodes": []map[string]any{
					{"id": "state-uuid-1234-5678-90ab-cdef01234567", "name": "Done"},
				},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveStateID(context.Background(), c, "Done", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "state-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
}

// -- ResolveViewerID --

func TestResolveViewerID_ReturnsID(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"viewer": map[string]any{"id": "viewer-uuid-1234-5678-90ab-cdef01234567"},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveViewerID(context.Background(), c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "viewer-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
}

func TestResolveViewerID_NotFound(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"viewer": map[string]any{"id": ""},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	_, err := ResolveViewerID(context.Background(), c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// -- ResolveProjectID --

func TestResolveProjectID_UUIDPassthrough(t *testing.T) {
	t.Parallel()
	uuid := "550e8400-e29b-41d4-a716-446655440000"
	got, err := ResolveProjectID(context.Background(), nil, uuid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != uuid {
		t.Errorf("want %q, got %q", uuid, got)
	}
}

func TestResolveProjectID_ByName(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"projects": map[string]any{
				"nodes": []map[string]any{
					{"id": "proj-uuid-1234-5678-90ab-cdef01234567", "name": "My Project"},
				},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	got, err := ResolveProjectID(context.Background(), c, "My Project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "proj-uuid-1234-5678-90ab-cdef01234567" {
		t.Errorf("unexpected id: %q", got)
	}
}

func TestResolveProjectID_NotFound(t *testing.T) {
	t.Parallel()
	srv := makeServer(t, map[string]any{
		"data": map[string]any{
			"projects": map[string]any{
				"nodes": []map[string]any{},
			},
		},
	})
	defer srv.Close()

	c := NewClient("key", WithEndpoint(srv.URL))
	_, err := ResolveProjectID(context.Background(), c, "Nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
