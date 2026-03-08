package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestClientDo_RateLimit_Retries(t *testing.T) {
	t.Parallel()
	attempts := 0
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts < 3 {
			writeJSON(w, map[string]any{
				"errors": []map[string]any{
					{"message": "rate limited", "extensions": map[string]any{"code": CodeRateLimited}},
				},
			})
			return
		}
		writeJSON(w, map[string]any{"data": map[string]any{"ok": true}})
	})

	var result map[string]any
	if err := c.Do(context.Background(), "query {}", nil, &result); err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientDo_RateLimit_ResetHeader(t *testing.T) {
	t.Parallel()
	var sleptFor time.Duration
	attempts := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			reset := strconv.FormatInt(time.Now().Add(2*time.Second).UnixMilli(), 10)
			w.Header().Set("X-RateLimit-Requests-Reset", reset)
			writeJSON(w, map[string]any{
				"errors": []map[string]any{
					{"message": "rate limited", "extensions": map[string]any{"code": CodeRateLimited}},
				},
			})
			return
		}
		writeJSON(w, map[string]any{"data": map[string]any{"ok": true}})
	}))
	defer server.Close()

	c := NewClient("key", WithEndpoint(server.URL))
	c.sleep = func(d time.Duration) { sleptFor = d }

	var result map[string]any
	if err := c.Do(context.Background(), "query {}", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// should have slept for ~2s from the reset header
	if sleptFor < time.Second {
		t.Errorf("expected sleep >= 1s from reset header, got %v", sleptFor)
	}
}

func TestClientDo_RateLimit_ExhaustedRetries(t *testing.T) {
	t.Parallel()
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"errors": []map[string]any{
				{"message": "rate limited", "extensions": map[string]any{"code": CodeRateLimited}},
			},
		})
	})

	err := c.Do(context.Background(), "query {}", nil, nil)
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	var gqlErrs GraphQLErrors
	if !errors.As(err, &gqlErrs) {
		t.Fatalf("expected GraphQLErrors, got %T: %v", err, err)
	}
	if len(gqlErrs) == 0 || gqlErrs[0].Code() != CodeRateLimited {
		t.Errorf("expected RATELIMITED error, got %v", gqlErrs)
	}
}

func TestRateLimitDelay_FallbackBackoff(t *testing.T) {
	t.Parallel()
	// no reset header -> exponential backoff
	resp := &http.Response{Header: make(http.Header)}

	d0 := rateLimitDelay(resp, 0) // base 1s
	d1 := rateLimitDelay(resp, 1) // base 2s

	if d0 < time.Second {
		t.Errorf("attempt 0 delay = %v, want >= 1s", d0)
	}
	if d1 < 2*time.Second {
		t.Errorf("attempt 1 delay = %v, want >= 2s", d1)
	}
}

func TestRateLimitDelay_ResetHeaderParsing(t *testing.T) {
	t.Parallel()
	future := time.Now().Add(3 * time.Second)
	resp := &http.Response{Header: make(http.Header)}
	resp.Header.Set("X-RateLimit-Requests-Reset", strconv.FormatInt(future.UnixMilli(), 10))

	d := rateLimitDelay(resp, 0)
	if d < 2*time.Second {
		t.Errorf("delay = %v, want >= 2s (reset is 3s away)", d)
	}
}

func TestServerErrorDelay_Backoff(t *testing.T) {
	t.Parallel()
	d0 := serverErrorDelay(0)
	d1 := serverErrorDelay(1)
	d4 := serverErrorDelay(4) // capped at 5s

	if d0 < time.Second {
		t.Errorf("attempt 0 delay = %v, want >= 1s", d0)
	}
	if d1 < 2*time.Second {
		t.Errorf("attempt 1 delay = %v, want >= 2s", d1)
	}
	// cap is 5s base + up to 1s jitter = up to 6s
	if d4 > 7*time.Second {
		t.Errorf("attempt 4 delay = %v, want <= 7s (cap + jitter)", d4)
	}
}

func TestClientDo_ContentTypeHeader(t *testing.T) {
	t.Parallel()
	var gotCT string
	_, c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotCT = r.Header.Get("Content-Type")
		writeJSON(w, map[string]any{"data": map[string]any{}})
	})

	_ = c.Do(context.Background(), "query {}", nil, nil)
	if gotCT != "application/json" {
		t.Errorf("Content-Type = %q, want %q", gotCT, "application/json")
	}
}

func TestClientDo_RateLimit_StatusCode400(t *testing.T) {
	t.Parallel()
	// HTTP 400 with RATELIMITED should trigger retry, not immediate failure
	attempts := 0
	_, c := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"errors": []map[string]any{
					{"message": "rate limited", "extensions": map[string]any{"code": CodeRateLimited}},
				},
			})
			return
		}
		writeJSON(w, map[string]any{"data": map[string]any{"ok": true}})
	})

	var result map[string]any
	if err := c.Do(context.Background(), "query {}", nil, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
