package grid

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBuildQueryWithVariablesAndSelection(t *testing.T) {
	spec := QuerySpec{
		OperationName: "GetSeries",
		FieldName:     "series",
		Args: []Arg{
			{Name: "id", GraphQLType: "ID!", Value: "2589176"},
		},
	}

	query, vars, err := BuildQuery(spec, Scalars("id", "startTimeScheduled"))
	if err != nil {
		t.Fatalf("BuildQuery returned error: %v", err)
	}
	if vars["id"] != "2589176" {
		t.Fatalf("expected variable id=2589176, got %#v", vars["id"])
	}
	want := strings.Join([]string{
		"query GetSeries($id: ID!) {",
		"  series(id: $id) {",
		"    id",
		"    startTimeScheduled",
		"  }",
		"}",
	}, "\n")
	if query != want {
		t.Fatalf("unexpected query:\ngot:\n%s\nwant:\n%s", query, want)
	}
}

func TestBuildQueryScalarWithoutSelection(t *testing.T) {
	spec := QuerySpec{
		OperationName: "SeriesIdByExternalId",
		FieldName:     "seriesIdByExternalId",
		Args: []Arg{
			{Name: "dataProviderName", GraphQLType: "String!", Value: "RIOT"},
			{Name: "externalSeriesId", GraphQLType: "ID!", Value: "ext-1"},
		},
	}

	query, vars, err := BuildQuery(spec, Selection{})
	if err != nil {
		t.Fatalf("BuildQuery returned error: %v", err)
	}
	if vars["dataProviderName"] != "RIOT" || vars["externalSeriesId"] != "ext-1" {
		t.Fatalf("unexpected variables: %#v", vars)
	}
	if strings.Contains(query, "{ \n") || strings.Count(query, "{") != 1 {
		t.Fatalf("expected single outer brace for scalar query, got:\n%s", query)
	}
}

func TestBuildQueryRequiresFieldName(t *testing.T) {
	if _, _, err := BuildQuery(QuerySpec{}, Selection{}); err == nil {
		t.Fatal("expected error when FieldName is empty")
	}
}

func TestModularFormsDecodeAndTransform(t *testing.T) {
	raw := json.RawMessage(`{"id":"3","name":"League of Legends"}`)

	baseForm := NewForm[Title](TitleSelectionBasic)
	transformed := Transform(baseForm, func(title Title) (string, error) {
		return title.ID + ":" + title.Name, nil
	})
	label, err := transformed.decodeRaw(raw)
	if err != nil {
		t.Fatalf("Transform decode failed: %v", err)
	}
	if label != "3:League of Legends" {
		t.Fatalf("unexpected transformed value: %q", label)
	}

	rawForm := RawForm(TitleSelectionBasic)
	rawOut, err := rawForm.decodeRaw(raw)
	if err != nil {
		t.Fatalf("RawForm decode failed: %v", err)
	}
	if string(rawOut) != string(raw) {
		t.Fatalf("expected raw json %s, got %s", raw, rawOut)
	}

	mapForm := MapForm(TitleSelectionBasic)
	mapOut, err := mapForm.decodeRaw(raw)
	if err != nil {
		t.Fatalf("MapForm decode failed: %v", err)
	}
	if mapOut["name"] != "League of Legends" {
		t.Fatalf("unexpected map value: %#v", mapOut)
	}
}
