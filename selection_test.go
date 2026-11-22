package grid

import (
	"strings"
	"testing"
)

func TestSelectionRenderAndImmutability(t *testing.T) {
	base := Select(
		Scalar("id"),
		Field("teams", Scalar("id"), Scalar("name")),
	)
	extended := base.With(
		Scalar("name"),
		Field("teams", Scalar("name"), Scalar("won")),
	)

	baseRendered := base.Render()
	if strings.Contains(baseRendered, "won") {
		t.Fatalf("expected base selection to remain immutable, got:\n%s", baseRendered)
	}

	got := extended.Render()
	want := strings.Join([]string{
		"id",
		"teams {",
		"  id",
		"  name",
		"  won",
		"}",
		"name",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected rendered selection:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestSelectionInlineFragmentAndAlias(t *testing.T) {
	sel := Select(
		AliasedField("seriesId", "id"),
		InlineFragment("TeamGameStatisticsLol",
			Field("kills", Scalar("avg")),
		),
	)

	got := sel.Render()
	want := strings.Join([]string{
		"seriesId: id",
		"... on TeamGameStatisticsLol {",
		"  kills {",
		"    avg",
		"  }",
		"}",
	}, "\n")
	if got != want {
		t.Fatalf("unexpected rendered selection:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestSelectionMergeAndConnectionSelection(t *testing.T) {
	merged := Merge(
		Scalars("id", "name"),
		Scalars("name", "logoUrl"),
	)
	conn := ConnectionSelection(merged)
	rendered := conn.Render()

	for _, needle := range []string{"totalCount", "pageInfo", "hasNextPage", "endCursor", "edges", "cursor", "node", "logoUrl"} {
		if !strings.Contains(rendered, needle) {
			t.Fatalf("expected connection selection to contain %q, got:\n%s", needle, rendered)
		}
	}
}
