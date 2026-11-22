package grid

import (
	"encoding/json"
	"fmt"
)

// Form defines a modular return shape for any GRID GraphQL query.
// It pairs a GraphQL Selection (what fields to request from the API) with
// a typed decoder (how to materialize the JSON field into Go type T).
type Form[T any] struct {
	Selection Selection
	Decode    func(raw json.RawMessage) (T, error)
}

// NewForm creates a Form[T] that requests the given Selection and decodes
// the response field into T via standard encoding/json.
func NewForm[T any](selection Selection) Form[T] {
	return Form[T]{
		Selection: selection,
	}
}

// NewCustomForm creates a Form[T] with a custom decoder/transformer function.
// Use this to project, flatten, or enrich GRID responses into application-specific types.
func NewCustomForm[T any](selection Selection, decode func(raw json.RawMessage) (T, error)) Form[T] {
	return Form[T]{
		Selection: selection,
		Decode:    decode,
	}
}

// RawForm creates a Form that returns the unparsed JSON payload for the given Selection.
func RawForm(selection Selection) Form[json.RawMessage] {
	return Form[json.RawMessage]{
		Selection: selection,
		Decode: func(raw json.RawMessage) (json.RawMessage, error) {
			out := make(json.RawMessage, len(raw))
			copy(out, raw)
			return out, nil
		},
	}
}

// MapForm creates a Form that decodes the response into a generic map[string]any.
func MapForm(selection Selection) Form[map[string]any] {
	return NewForm[map[string]any](selection)
}

// ScalarForm creates a Form for queries that return a GraphQL scalar (such as ID or String)
// and therefore require no selection set.
func ScalarForm[T any]() Form[T] {
	return Form[T]{}
}

// WithSelection returns a new Form[T] with additional GraphQL selection nodes merged in.
func (f Form[T]) WithSelection(extra ...SelectionNode) Form[T] {
	return Form[T]{
		Selection: f.Selection.With(extra...),
		Decode:    f.Decode,
	}
}

// Transform wraps an existing Form[In] with a pure mapping function to produce a Form[Out].
func Transform[In, Out any](base Form[In], mapFn func(In) (Out, error)) Form[Out] {
	return Form[Out]{
		Selection: base.Selection,
		Decode: func(raw json.RawMessage) (Out, error) {
			var zero Out
			in, err := base.decodeRaw(raw)
			if err != nil {
				return zero, err
			}
			return mapFn(in)
		},
	}
}

// ConnectionForm wraps a node Form[T] into a paginated Relay Connection[T] Form.
func ConnectionForm[T any](nodeForm Form[T]) Form[Connection[T]] {
	return Form[Connection[T]]{
		Selection: ConnectionSelection(nodeForm.Selection),
		Decode: func(raw json.RawMessage) (Connection[T], error) {
			var wire struct {
				TotalCount int      `json:"totalCount"`
				PageInfo   PageInfo `json:"pageInfo"`
				Edges      []struct {
					Cursor string          `json:"cursor"`
					Node   json.RawMessage `json:"node"`
				} `json:"edges"`
			}
			if err := json.Unmarshal(raw, &wire); err != nil {
				return Connection[T]{}, fmt.Errorf("decode connection: %w", err)
			}
			edges := make([]Edge[T], len(wire.Edges))
			for i, e := range wire.Edges {
				node, err := nodeForm.decodeRaw(e.Node)
				if err != nil {
					return Connection[T]{}, fmt.Errorf("decode connection edge %d: %w", i, err)
				}
				edges[i] = Edge[T]{
					Cursor: e.Cursor,
					Node:   node,
				}
			}
			return Connection[T]{
				TotalCount: wire.TotalCount,
				PageInfo:   wire.PageInfo,
				Edges:      edges,
			}, nil
		},
	}
}

func (f Form[T]) decodeRaw(raw json.RawMessage) (T, error) {
	if f.Decode != nil {
		return f.Decode(raw)
	}
	var out T
	if len(raw) == 0 || string(raw) == "null" {
		return out, nil
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return out, fmt.Errorf("decode response: %w", err)
	}
	return out, nil
}
