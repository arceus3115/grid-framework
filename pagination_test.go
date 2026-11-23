package grid

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
)

func TestIterateNodesBackward(t *testing.T) {
	nodes := iterateNodes(t.Context(), PageParams{Before: String("start")}, func(page PageParams) (Connection[int], error) {
		if page.Last == nil || *page.Last != MaxPageSize || page.First != nil || page.After != nil || page.Before == nil {
			t.Fatalf("unexpected backward request: %+v", page)
		}
		switch *page.Before {
		case "start":
			return Connection[int]{
				PageInfo: PageInfo{HasPreviousPage: true, StartCursor: String("cursor-3")},
				Edges:    []Edge[int]{{Node: 3}, {Node: 4}},
			}, nil
		case "cursor-3":
			return Connection[int]{Edges: []Edge[int]{{Node: 1}, {Node: 2}}}, nil
		default:
			t.Fatalf("unexpected before cursor %q", *page.Before)
			return Connection[int]{}, nil
		}
	})
	var got []int
	for node, err := range nodes {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, node)
	}
	if !slices.Equal(got, []int{4, 3, 2, 1}) {
		t.Fatalf("backward traversal = %v", got)
	}
}

func TestCollectAllRejectsBrokenCursor(t *testing.T) {
	for _, tc := range []struct {
		name     string
		cursor   *string
		message  string
		maxCalls int
	}{
		{name: "missing", message: "without a cursor", maxCalls: 1},
		{name: "repeated", cursor: String("cursor-1"), message: "repeated cursor", maxCalls: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			fetch := func(context.Context, *string) (Connection[int], error) {
				calls++
				if calls > tc.maxCalls {
					t.Fatal("fetched another page with a broken cursor")
				}
				return Connection[int]{
					PageInfo: PageInfo{HasNextPage: true, EndCursor: tc.cursor},
					Edges:    []Edge[int]{{Node: 1}},
				}, nil
			}
			if got, err := CollectAll(t.Context(), 0, fetch); got != nil || err == nil || !strings.Contains(err.Error(), tc.message) || calls != tc.maxCalls {
				t.Fatalf("CollectAll = %v, %v (%d requests)", got, err, calls)
			}
			calls = 0
			if got, err := CollectAll(t.Context(), 1, fetch); !slices.Equal(got, []int{1}) || err != nil || calls != 1 {
				t.Fatalf("limited CollectAll = %v, %v (%d requests)", got, err, calls)
			}
		})
	}
}

func TestIterateNodesCancellationBetweenPages(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	calls := 0
	nodes := iterateNodes(ctx, PageParams{}, func(PageParams) (Connection[int], error) {
		calls++
		return Connection[int]{
			PageInfo: PageInfo{HasNextPage: true, EndCursor: String("cursor-1")},
			Edges:    []Edge[int]{{Node: 1}},
		}, nil
	})
	var failures int
	for node, err := range nodes {
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("iteration error = %v, want context.Canceled", err)
			}
			failures++
			continue
		}
		if node != 1 {
			t.Fatalf("unexpected node %d", node)
		}
		cancel()
	}
	if calls != 1 || failures != 1 {
		t.Fatalf("got %d requests and %d terminal errors", calls, failures)
	}
}
