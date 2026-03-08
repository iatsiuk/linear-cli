package api

import (
	"context"
	"errors"
	"testing"
)

func TestPaginateAll_SinglePage(t *testing.T) {
	t.Parallel()
	calls := 0
	fetch := func(_ context.Context, after *string, first int) (Connection[string], error) {
		calls++
		if after != nil {
			t.Errorf("unexpected cursor on single-page fetch")
		}
		if first != 50 {
			t.Errorf("want first=50, got %d", first)
		}
		return Connection[string]{
			Nodes:    []string{"a", "b", "c"},
			PageInfo: PageInfo{HasNextPage: false},
		}, nil
	}

	nodes, err := PaginateAll(context.Background(), fetch, 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("want 1 call, got %d", calls)
	}
	if len(nodes) != 3 {
		t.Errorf("want 3 nodes, got %d", len(nodes))
	}
	for i, want := range []string{"a", "b", "c"} {
		if nodes[i] != want {
			t.Errorf("nodes[%d]: want %q, got %q", i, want, nodes[i])
		}
	}
}

func TestPaginateAll_MultiPage(t *testing.T) {
	t.Parallel()
	cursor1 := "cursor-after-page1"
	responses := []Connection[string]{
		{
			Nodes:    []string{"a", "b"},
			PageInfo: PageInfo{HasNextPage: true, EndCursor: &cursor1},
		},
		{
			Nodes:    []string{"c"},
			PageInfo: PageInfo{HasNextPage: false},
		},
	}
	page := 0
	fetch := func(_ context.Context, after *string, first int) (Connection[string], error) {
		if page == 1 && (after == nil || *after != cursor1) {
			t.Errorf("want cursor %q on page 2, got %v", cursor1, after)
		}
		resp := responses[page]
		page++
		return resp, nil
	}

	nodes, err := PaginateAll(context.Background(), fetch, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page != 2 {
		t.Errorf("want 2 fetches, got %d", page)
	}
	want := []string{"a", "b", "c"}
	if len(nodes) != len(want) {
		t.Fatalf("want %d nodes, got %d", len(want), len(nodes))
	}
	for i := range want {
		if nodes[i] != want[i] {
			t.Errorf("nodes[%d]: want %q, got %q", i, want[i], nodes[i])
		}
	}
}

func TestPaginateAll_StopsAtLastPage(t *testing.T) {
	t.Parallel()
	// three pages, stops after HasNextPage=false
	cursors := []*string{strPtr("c1"), strPtr("c2"), nil}
	hasNext := []bool{true, true, false}
	page := 0
	fetch := func(_ context.Context, _ *string, _ int) (Connection[int], error) {
		resp := Connection[int]{
			Nodes:    []int{page},
			PageInfo: PageInfo{HasNextPage: hasNext[page], EndCursor: cursors[page]},
		}
		page++
		return resp, nil
	}

	nodes, err := PaginateAll(context.Background(), fetch, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page != 3 {
		t.Errorf("want 3 fetches, got %d", page)
	}
	if len(nodes) != 3 {
		t.Errorf("want 3 nodes, got %d", len(nodes))
	}
}

func TestPaginateAll_EmptyResult(t *testing.T) {
	t.Parallel()
	fetch := func(_ context.Context, _ *string, _ int) (Connection[string], error) {
		return Connection[string]{
			Nodes:    nil,
			PageInfo: PageInfo{HasNextPage: false},
		}, nil
	}

	nodes, err := PaginateAll(context.Background(), fetch, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("want 0 nodes, got %d", len(nodes))
	}
}

func TestPaginateAll_FetchError(t *testing.T) {
	t.Parallel()
	fetchErr := errors.New("network error")
	fetch := func(_ context.Context, _ *string, _ int) (Connection[string], error) {
		return Connection[string]{}, fetchErr
	}

	_, err := PaginateAll(context.Background(), fetch, 10)
	if !errors.Is(err, fetchErr) {
		t.Errorf("want fetchErr, got %v", err)
	}
}

func TestPaginateAll_HasNextPageWithNilCursor(t *testing.T) {
	t.Parallel()
	fetch := func(_ context.Context, _ *string, _ int) (Connection[string], error) {
		return Connection[string]{
			Nodes:    []string{"a"},
			PageInfo: PageInfo{HasNextPage: true, EndCursor: nil},
		}, nil
	}

	_, err := PaginateAll(context.Background(), fetch, 10)
	if err == nil {
		t.Fatal("expected error when hasNextPage=true but endCursor=nil")
	}
}

func TestPaginateAll_ContextCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	fetch := func(ctx context.Context, _ *string, _ int) (Connection[string], error) {
		called = true
		return Connection[string]{}, ctx.Err()
	}

	_, err := PaginateAll(ctx, fetch, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !called {
		t.Error("fetch was not called")
	}
}

func strPtr(s string) *string { return &s }
