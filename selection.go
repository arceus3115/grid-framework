package grid

import "strings"

// SelectionNode represents a single item inside a GraphQL selection set:
// a scalar field, an object field with sub-selections, an inline fragment, or raw GraphQL text.
type SelectionNode struct {
	Name           string
	Alias          string
	Args           string
	InlineFragment bool
	Raw            string
	Sub            []SelectionNode
}

// Selection is an immutable, composable GraphQL selection set.
type Selection struct {
	nodes []SelectionNode
}

// Select creates a new Selection from the provided nodes.
func Select(nodes ...SelectionNode) Selection {
	cloned := make([]SelectionNode, len(nodes))
	copy(cloned, nodes)
	return Selection{nodes: cloned}
}

// Scalars creates a Selection containing only scalar field names.
func Scalars(names ...string) Selection {
	nodes := make([]SelectionNode, 0, len(names))
	for _, n := range names {
		if trimmed := strings.TrimSpace(n); trimmed != "" {
			nodes = append(nodes, SelectionNode{Name: trimmed})
		}
	}
	return Selection{nodes: nodes}
}

// Scalar creates a single scalar SelectionNode.
func Scalar(name string) SelectionNode {
	return SelectionNode{Name: strings.TrimSpace(name)}
}

// Field creates a SelectionNode for a field with optional nested sub-selections.
func Field(name string, sub ...SelectionNode) SelectionNode {
	cloned := make([]SelectionNode, len(sub))
	copy(cloned, sub)
	return SelectionNode{
		Name: strings.TrimSpace(name),
		Sub:  cloned,
	}
}

// FieldSelect creates a SelectionNode whose nested selection comes from an existing Selection.
func FieldSelect(name string, sub Selection) SelectionNode {
	return Field(name, sub.nodes...)
}

// AliasedField creates a SelectionNode with a GraphQL alias (e.g. `alias: field { ... }`).
func AliasedField(alias, name string, sub ...SelectionNode) SelectionNode {
	node := Field(name, sub...)
	node.Alias = strings.TrimSpace(alias)
	return node
}

// InlineFragment creates a GraphQL inline fragment (`... on TypeName { ... }`).
// This is useful for GRID interfaces such as ContentCatalogEntity or title-specific Stats Feed types.
func InlineFragment(typeName string, sub ...SelectionNode) SelectionNode {
	cloned := make([]SelectionNode, len(sub))
	copy(cloned, sub)
	return SelectionNode{
		Name:           strings.TrimSpace(typeName),
		InlineFragment: true,
		Sub:            cloned,
	}
}

// RawSelection creates a Selection from a raw GraphQL selection body (without outer braces).
func RawSelection(gql string) Selection {
	trimmed := strings.TrimSpace(gql)
	if trimmed == "" {
		return Selection{}
	}
	return Selection{nodes: []SelectionNode{{Raw: trimmed}}}
}

// IsEmpty reports whether the selection contains no nodes.
func (s Selection) IsEmpty() bool {
	return len(s.nodes) == 0
}

// Nodes returns a defensive copy of the selection's top-level nodes.
func (s Selection) Nodes() []SelectionNode {
	out := make([]SelectionNode, len(s.nodes))
	copy(out, s.nodes)
	return out
}

// With returns a new Selection with additional nodes appended and merged without mutating s.
func (s Selection) With(extra ...SelectionNode) Selection {
	merged := mergeNodes(s.nodes, extra)
	return Selection{nodes: merged}
}

// Merge combines multiple Selections into a single deduplicated Selection.
func Merge(selections ...Selection) Selection {
	var merged []SelectionNode
	for _, sel := range selections {
		merged = mergeNodes(merged, sel.nodes)
	}
	return Selection{nodes: merged}
}

// ConnectionSelection wraps a node Selection in a standard GRID Relay connection selection set.
func ConnectionSelection(nodeSelection Selection) Selection {
	return Select(
		Scalar("totalCount"),
		Field("pageInfo",
			Scalar("hasPreviousPage"),
			Scalar("hasNextPage"),
			Scalar("startCursor"),
			Scalar("endCursor"),
		),
		Field("edges",
			Scalar("cursor"),
			FieldSelect("node", nodeSelection),
		),
	)
}

// Render formats the Selection as a GraphQL selection block (excluding outer braces).
func (s Selection) Render() string {
	return renderNodes(s.nodes, 0)
}

func mergeNodes(base, extra []SelectionNode) []SelectionNode {
	out := make([]SelectionNode, 0, len(base)+len(extra))
	index := make(map[string]int, len(base)+len(extra))

	add := func(n SelectionNode) {
		if n.Raw != "" {
			key := "raw:" + n.Raw
			if _, exists := index[key]; !exists {
				index[key] = len(out)
				out = append(out, n)
			}
			return
		}
		if n.Name == "" {
			return
		}
		key := nodeMergeKey(n)
		if idx, exists := index[key]; exists {
			if len(n.Sub) > 0 {
				out[idx].Sub = mergeNodes(out[idx].Sub, n.Sub)
			}
			return
		}
		index[key] = len(out)
		cloned := n
		if len(n.Sub) > 0 {
			cloned.Sub = mergeNodes(nil, n.Sub)
		}
		out = append(out, cloned)
	}

	for _, n := range base {
		add(n)
	}
	for _, n := range extra {
		add(n)
	}
	return out
}

func nodeMergeKey(n SelectionNode) string {
	if n.InlineFragment {
		return "frag:" + n.Name
	}
	return "field:" + n.Alias + ":" + n.Name + ":" + n.Args
}

func renderNodes(nodes []SelectionNode, depth int) string {
	if len(nodes) == 0 {
		return ""
	}
	indent := strings.Repeat("  ", depth)
	lines := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.Raw != "" {
			for _, line := range strings.Split(n.Raw, "\n") {
				lines = append(lines, indent+strings.TrimSpace(line))
			}
			continue
		}
		if n.Name == "" {
			continue
		}
		var header strings.Builder
		header.WriteString(indent)
		if n.InlineFragment {
			header.WriteString("... on ")
			header.WriteString(n.Name)
		} else {
			if n.Alias != "" {
				header.WriteString(n.Alias)
				header.WriteString(": ")
			}
			header.WriteString(n.Name)
			if n.Args != "" {
				header.WriteString("(")
				header.WriteString(n.Args)
				header.WriteString(")")
			}
		}
		if len(n.Sub) > 0 {
			header.WriteString(" {\n")
			header.WriteString(renderNodes(n.Sub, depth+1))
			header.WriteString("\n")
			header.WriteString(indent)
			header.WriteString("}")
		}
		lines = append(lines, header.String())
	}
	return strings.Join(lines, "\n")
}
