package grid

import (
	"encoding/json"
	"net/http"
	"reflect"
	"slices"
	"testing"
	"time"
)

func TestRecentSeries(t *testing.T) {
	var captured struct {
		OperationName string
		Variables     struct {
			First          int
			OrderBy        SeriesOrderBy
			OrderDirection OrderDirection
			Filter         SeriesFilter
		}
	}
	calls := 0
	client := NewClient("key", WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls++
		if err := json.NewDecoder(req.Body).Decode(&captured); err != nil {
			t.Fatal(err)
		}
		return stubResponse(http.StatusOK, `{"data":{"allSeries":{"edges":[{"node":{"id":"series-1"}}]}}}`), nil
	})}))
	for _, invalid := range []struct {
		titleID string
		limit   int
	}{{"", 1}, {TitleIDLeagueOfLegends, 0}, {TitleIDLeagueOfLegends, MaxPageSize + 1}} {
		if _, err := client.CentralData.RecentSeries(t.Context(), invalid.titleID, invalid.limit); err == nil {
			t.Fatalf("expected validation error for %+v", invalid)
		}
	}
	if calls != 0 {
		t.Fatalf("invalid arguments made %d requests", calls)
	}

	start := time.Now()
	series, err := client.CentralData.RecentSeries(t.Context(), TitleIDLeagueOfLegends, 3)
	if err != nil || len(series) != 1 || series[0].ID != "series-1" || calls != 1 {
		t.Fatalf("RecentSeries = %#v, %v (%d requests)", series, err, calls)
	}
	args := captured.Variables
	if captured.OperationName != "AllSeries" || args.First != 3 || args.OrderBy != SeriesOrderByStartTimeScheduled || args.OrderDirection != OrderDirectionDesc {
		t.Fatalf("unexpected series request: %+v", captured)
	}
	if args.Filter.TitleIDs == nil || !slices.Equal(args.Filter.TitleIDs.In, []string{TitleIDLeagueOfLegends}) {
		t.Fatalf("unexpected title filter: %+v", args.Filter.TitleIDs)
	}
	if args.Filter.StartTimeScheduled == nil || args.Filter.StartTimeScheduled.LTE == nil {
		t.Fatal("missing scheduled start time bound")
	}
	bound, err := time.Parse(time.RFC3339Nano, *args.Filter.StartTimeScheduled.LTE)
	if err != nil || bound.Before(start) || bound.After(time.Now()) {
		t.Fatalf("scheduled start time bound = %s, %v", bound, err)
	}
}

func TestListTeamsFetchesPagesOnDemand(t *testing.T) {
	var requests []graphQLRequest
	client := NewClient("key", WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		var query graphQLRequest
		if err := json.NewDecoder(req.Body).Decode(&query); err != nil {
			t.Fatal(err)
		}
		requests = append(requests, query)
		if query.Variables["after"] == "cursor-1" {
			return stubResponse(http.StatusOK, `{"data":{"teams":{"edges":[{"node":{"id":"2"}}]}}}`), nil
		}
		return stubResponse(http.StatusOK, `{"data":{"teams":{"pageInfo":{"hasNextPage":true,"endCursor":"cursor-1"},"edges":[{"node":{"id":"1"}}]}}}`), nil
	})}))
	teams := client.CentralData.ListTeams(t.Context(), TeamsParams{Filter: &TeamFilter{TitleID: String(TitleIDValorant)}})
	if len(requests) != 0 {
		t.Fatal("iterator fetched before iteration")
	}
	for team, err := range teams {
		if err != nil || team.ID != "1" {
			t.Fatalf("first result = %#v, %v", team, err)
		}
		break
	}
	if len(requests) != 1 {
		t.Fatalf("early stop made %d requests, want 1", len(requests))
	}

	var ids []string
	for team, err := range teams {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, team.ID)
	}
	if !slices.Equal(ids, []string{"1", "2"}) || len(requests) != 3 {
		t.Fatalf("replayed list returned ids %v and made %d total requests", ids, len(requests))
	}
	for i, request := range requests {
		want := map[string]any{
			"first":  float64(MaxPageSize),
			"filter": map[string]any{"titleId": TitleIDValorant},
		}
		if i == 2 {
			want["after"] = "cursor-1"
		}
		if request.OperationName != "Teams" || !reflect.DeepEqual(request.Variables, want) {
			t.Fatalf("request %d = %+v, want Teams with variables %v", i, request, want)
		}
	}
}
