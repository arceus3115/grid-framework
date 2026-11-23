package grid

import (
	"context"
	"errors"
	"fmt"
	"iter"
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
	forward := p.First != nil || hasCursor(p.After)
	if forward && p.backward() {
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

func (p PageParams) backward() bool {
	return p.Last != nil || hasCursor(p.Before)
}

func hasCursor(cursor *string) bool {
	return cursor != nil && *cursor != ""
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
// or maxPages is reached (maxPages <= 0 fetches all pages). It starts with a
// nil cursor and returns an error if GRID indicates a next page without
// providing a new cursor.
func CollectAll[T any](
	ctx context.Context,
	maxPages int,
	fetchPage func(ctx context.Context, after *string) (Connection[T], error),
) ([]T, error) {
	var all []T
	pages := paginate(ctx, PageParams{}, maxPages, func(page PageParams) (Connection[T], error) {
		return fetchPage(ctx, page.After)
	})
	for conn, err := range pages {
		if err != nil {
			return nil, err
		}
		for _, edge := range conn.Edges {
			all = append(all, edge.Node)
		}
	}
	return all, nil
}

func iterateNodes[T any](ctx context.Context, page PageParams, fetchPage func(PageParams) (Connection[T], error)) iter.Seq2[T, error] {
	backward := page.backward()
	if page.First == nil && page.Last == nil {
		if backward {
			page.Last = Int(MaxPageSize)
		} else {
			page.First = Int(MaxPageSize)
		}
	}
	pages := paginate(ctx, page, 0, fetchPage)
	return func(yield func(T, error) bool) {
		for conn, err := range pages {
			if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			for i := range conn.Edges {
				if backward {
					i = len(conn.Edges) - 1 - i
				}
				if !yield(conn.Edges[i].Node, nil) {
					return
				}
			}
		}
	}
}

func paginate[T any](ctx context.Context, initial PageParams, maxPages int, fetchPage func(PageParams) (Connection[T], error)) iter.Seq2[Connection[T], error] {
	return func(yield func(Connection[T], error) bool) {
		page := initial
		if err := page.Validate(); err != nil {
			yield(Connection[T]{}, err)
			return
		}

		backward := page.backward()
		cursor := page.After
		if backward {
			cursor = page.Before
		}
		seen := make(map[string]struct{})
		if hasCursor(cursor) {
			seen[*cursor] = struct{}{}
		}

		for pageNumber := 1; ; pageNumber++ {
			if err := ctx.Err(); err != nil {
				yield(Connection[T]{}, err)
				return
			}
			conn, err := fetchPage(page)
			if err != nil {
				yield(Connection[T]{}, err)
				return
			}
			if !yield(conn, nil) || pageNumber == maxPages {
				return
			}

			more, next := conn.PageInfo.HasNextPage, conn.PageInfo.EndCursor
			if backward {
				more, next = conn.PageInfo.HasPreviousPage, conn.PageInfo.StartCursor
			}
			if !more {
				return
			}
			if !hasCursor(next) {
				yield(Connection[T]{}, errors.New("grid pagination: GRID reported another page without a cursor"))
				return
			}
			if _, repeated := seen[*next]; repeated {
				yield(Connection[T]{}, errors.New("grid pagination: GRID returned a repeated cursor"))
				return
			}
			seen[*next] = struct{}{}
			if backward {
				page.Before = next
			} else {
				page.After = next
			}
		}
	}
}
