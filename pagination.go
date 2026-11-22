package grid

import (
	"context"
	"fmt"
)

// MaxPageSize is the maximum number of items GRID allows in a single paginated request.
const MaxPageSize = 50

// PageParams represents Relay-style cursor pagination parameters supported by GRID Central Data.
// Specify either First/After (forward pagination) or Last/Before (backward pagination), never both.
type PageParams struct {
	First  *int
	After  *string
	Last   *int
	Before *string
}

// Validate checks pagination invariants documented by GRID so invalid requests fail fast.
func (p PageParams) Validate() error {
	forward := p.First != nil || (p.After != nil && *p.After != "")
	backward := p.Last != nil || (p.Before != nil && *p.Before != "")
	if forward && backward {
		return fmt.Errorf("grid pagination: cannot combine first/after with last/before")
	}
	if p.First != nil && (*p.First < 1 || *p.First > MaxPageSize) {
		return fmt.Errorf("grid pagination: first must be between 1 and %d, got %d", MaxPageSize, *p.First)
	}
	if p.Last != nil && (*p.Last < 1 || *p.Last > MaxPageSize) {
		return fmt.Errorf("grid pagination: last must be between 1 and %d, got %d", MaxPageSize, *p.Last)
	}
	return nil
}

// PageInfo contains cursor boundaries returned by Relay-style GRID connections.
type PageInfo struct {
	HasPreviousPage bool    `json:"hasPreviousPage"`
	HasNextPage     bool    `json:"hasNextPage"`
	StartCursor     *string `json:"startCursor"`
	EndCursor       *string `json:"endCursor"`
}

// Edge pairs a cursor with a decoded node of type T.
type Edge[T any] struct {
	Cursor string `json:"cursor"`
	Node   T      `json:"node"`
}

// Connection is a generic Relay-style paginated result.
type Connection[T any] struct {
	TotalCount int       `json:"totalCount"`
	PageInfo   PageInfo  `json:"pageInfo"`
	Edges      []Edge[T] `json:"edges"`
}

// Nodes extracts a flat slice of all node values from the connection's edges.
func (c Connection[T]) Nodes() []T {
	nodes := make([]T, len(c.Edges))
	for i, e := range c.Edges {
		nodes[i] = e.Node
	}
	return nodes
}

// CollectAll fetches pages sequentially using cursor pagination until no more pages exist
// or maxPages is reached (maxPages <= 0 fetches all pages).
func CollectAll[T any](
	ctx context.Context,
	maxPages int,
	fetchPage func(ctx context.Context, after *string) (Connection[T], error),
) ([]T, error) {
	var all []T
	var cursor *string
	for page := 0; maxPages <= 0 || page < maxPages; page++ {
		conn, err := fetchPage(ctx, cursor)
		if err != nil {
			return nil, err
		}
		all = append(all, conn.Nodes()...)
		if !conn.PageInfo.HasNextPage || conn.PageInfo.EndCursor == nil || *conn.PageInfo.EndCursor == "" {
			break
		}
		cursor = conn.PageInfo.EndCursor
	}
	return all, nil
}
