package api

import (
	"context"
	"fmt"
)

// PageInfo holds cursor-based pagination info from a GraphQL connection.
type PageInfo struct {
	HasNextPage bool    `json:"hasNextPage"`
	EndCursor   *string `json:"endCursor"`
}

// Connection is a generic GraphQL connection with nodes and pagination info.
type Connection[T any] struct {
	Nodes    []T      `json:"nodes"`
	PageInfo PageInfo `json:"pageInfo"`
}

// FetchFunc fetches a single page of results given an optional after cursor and page size.
type FetchFunc[T any] func(ctx context.Context, after *string, first int) (Connection[T], error)

// PaginateAll fetches all pages using fetch, collecting all nodes.
// Each page requests pageSize items. Stops when PageInfo.HasNextPage is false.
func PaginateAll[T any](ctx context.Context, fetch FetchFunc[T], pageSize int) ([]T, error) {
	var (
		all   []T
		after *string
	)
	for {
		page, err := fetch(ctx, after, pageSize)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Nodes...)
		if !page.PageInfo.HasNextPage {
			break
		}
		if page.PageInfo.EndCursor == nil {
			return nil, fmt.Errorf("pagination: hasNextPage is true but endCursor is nil")
		}
		after = page.PageInfo.EndCursor
	}
	return all, nil
}
