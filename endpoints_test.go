package grid

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCentralDataAllSeriesWithCustomForm(t *testing.T) {
	var received graphQLRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"allSeries": {
					"totalCount": 1,
					"pageInfo": {"hasPreviousPage": false, "hasNextPage": false, "startCursor": "cur1", "endCursor": "cur1"},
					"edges": [{
						"cursor": "cur1",
						"node": {
							"id": "2589176",
							"startTimeScheduled": "2026-01-15T18:00:00Z",
							"tournament": {"name": "LCK Spring"}
						}
					}]
				}
			}
		}`))
	}))
	defer srv.Close()

	type SeriesSummary struct {
		ID                 string `json:"id"`
		StartTimeScheduled string `json:"startTimeScheduled"`
		Tournament         struct {
			Name string `json:"name"`
		} `json:"tournament"`
	}

	summaryForm := NewForm[SeriesSummary](
		Scalars("id", "startTimeScheduled").With(
			Field("tournament", Scalar("name")),
		),
	)

	client := NewClient("key", WithCentralDataEndpoint(srv.URL))
	conn, err := AllSeriesWithForm(context.Background(), client, AllSeriesParams{
		Page:           PageParams{First: Int(5)},
		OrderBy:        SeriesOrderByStartTimeScheduled,
		OrderDirection: OrderDirectionDesc,
		Filter: &SeriesFilter{
			TitleIDs: &IDFilter{In: []string{TitleIDLeagueOfLegends}},
		},
	}, summaryForm)
	if err != nil {
		t.Fatalf("AllSeriesWithForm failed: %v", err)
	}

	if received.OperationName != "AllSeries" {
		t.Fatalf("expected operation AllSeries, got %q", received.OperationName)
	}
	if !strings.Contains(received.Query, "tournament") {
		t.Fatalf("expected query to include custom tournament selection, got:\n%s", received.Query)
	}
	if len(conn.Edges) != 1 || conn.Edges[0].Node.ID != "2589176" || conn.Edges[0].Node.Tournament.Name != "LCK Spring" {
		t.Fatalf("unexpected decoded connection: %#v", conn)
	}
}

func TestCentralDataExternalIDLookup(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"seriesIdByExternalId":"991122"}}`))
	}))
	defer srv.Close()

	client := NewClient("key", WithCentralDataEndpoint(srv.URL))
	id, err := client.CentralData.SeriesIDByExternalID(context.Background(), "RIOT", "ESPORTSTMNT01:123")
	if err != nil {
		t.Fatalf("SeriesIDByExternalID failed: %v", err)
	}
	if id == nil || *id != "991122" {
		t.Fatalf("expected id 991122, got %v", id)
	}
}

func TestSeriesStateGetAndTransformedForm(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"seriesState": {
					"started": true,
					"finished": true,
					"teams": [
						{"id": "1", "name": "T1", "won": true},
						{"id": "2", "name": "Gen.G", "won": false}
					],
					"games": [{
						"id": "g1",
						"sequenceNumber": 1,
						"started": true,
						"finished": true,
						"teams": [{
							"id": "1",
							"name": "T1",
							"side": "blue",
							"won": true,
							"players": [{"id": "p1", "name": "Faker", "kills": 7, "deaths": 1}]
						}]
					}]
				}
			}
		}`))
	}))
	defer srv.Close()

	client := NewClient("key", WithSeriesStateEndpoint(srv.URL))
	state, err := client.SeriesState.Get(context.Background(), "2589176")
	if err != nil {
		t.Fatalf("SeriesState.Get failed: %v", err)
	}
	if !state.Finished || len(state.Teams) != 2 || state.Games[0].Teams[0].Players[0].Name != "Faker" {
		t.Fatalf("unexpected series state: %#v", state)
	}

	winnerForm := Transform(NewForm[SeriesState](SeriesStateSelectionBasic), func(s SeriesState) (string, error) {
		for _, team := range s.Teams {
			if team.Won {
				return team.Name, nil
			}
		}
		return "", nil
	})
	winner, err := SeriesStateWithForm(context.Background(), client, "2589176", winnerForm)
	if err != nil {
		t.Fatalf("SeriesStateWithForm failed: %v", err)
	}
	if winner != "T1" {
		t.Fatalf("expected winner T1, got %q", winner)
	}
}

func TestStatsFeedTeamAndCustomInlineFragmentForm(t *testing.T) {
	var received graphQLRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.Header().Set("Content-Type", "application/json")
		switch received.OperationName {
		case "TeamStatistics":
			_, _ = w.Write([]byte(`{
				"data": {
					"teamStatistics": {
						"id": "47351",
						"aggregationSeriesIds": ["s1", "s2"],
						"series": {"count": 2, "kills": {"sum": 40, "min": 18, "max": 22, "avg": 20}},
						"game": {"count": 5, "kills": {"sum": 75, "min": 10, "max": 20, "avg": 15}}
					}
				}
			}`))
		case "TeamGameStatistics":
			_, _ = w.Write([]byte(`{
				"data": {
					"teamGameStatistics": {
						"count": 5,
						"kills": {"sum": 75, "min": 10, "max": 20, "avg": 15},
						"baronPowerPlay": {"avg": 1250.5}
					}
				}
			}`))
		}
	}))
	defer srv.Close()

	client := NewClient("key", WithStatsEndpoint(srv.URL))
	stats, err := client.Stats.TeamStatistics(context.Background(), TeamStatisticsParams{
		TeamID: "47351",
		Filter: TeamStatisticsFilter{TimeWindow: TimeRangeLast3Months},
	})
	if err != nil {
		t.Fatalf("TeamStatistics failed: %v", err)
	}
	if stats.ID != "47351" || stats.Series == nil || stats.Series.Count != 2 || stats.Game.Kills.Avg != 15 {
		t.Fatalf("unexpected team statistics: %#v", stats)
	}

	lolForm := MapForm(TeamGameStatisticsSelectionBasic.With(
		InlineFragment("TeamGameStatisticsLol",
			Field("baronPowerPlay", Scalar("avg")),
		),
	))
	rawMap, err := TeamGameStatisticsWithForm(context.Background(), client, TeamGameStatisticsParams{
		TeamID: "47351",
		Selection: &GameSelection{
			Filter: &GameStateFilter{
				StartedAt: &StatsDateTimeFilter{Period: TimeRangeLastMonth},
			},
			First: Int(5),
		},
	}, lolForm)
	if err != nil {
		t.Fatalf("TeamGameStatisticsWithForm failed: %v", err)
	}
	if !strings.Contains(received.Query, "... on TeamGameStatisticsLol") {
		t.Fatalf("expected inline fragment in query, got:\n%s", received.Query)
	}
	selection, err := json.Marshal(received.Variables["selection"])
	if err != nil {
		t.Fatalf("marshal selection variable: %v", err)
	}
	if want := `{"filter":{"startedAt":{"period":"LAST_MONTH"}},"first":5}`; string(selection) != want {
		t.Fatalf("selection variable = %s, want %s", selection, want)
	}
	if rawMap["count"].(float64) != 5 {
		t.Fatalf("unexpected map result: %#v", rawMap)
	}
}

func TestRequiredIDValidation(t *testing.T) {
	client := NewClient("key")
	ctx := context.Background()

	if _, err := client.CentralData.Series(ctx, ""); err == nil {
		t.Fatal("expected error for empty series id")
	}
	if _, err := client.SeriesState.Get(ctx, ""); err == nil {
		t.Fatal("expected error for empty series state id")
	}
	if _, err := client.Stats.TeamStatistics(ctx, TeamStatisticsParams{}); err == nil {
		t.Fatal("expected error for empty team id")
	}
	if _, err := client.Stats.PlayerStatistics(ctx, PlayerStatisticsParams{}); err == nil {
		t.Fatal("expected error for empty player id")
	}
	if _, err := client.Stats.SeriesStatistics(ctx, SeriesStatisticsParams{}); err == nil {
		t.Fatal("expected error for empty title id")
	}
	if _, err := client.Stats.GameStatistics(ctx, GameStatisticsParams{}); err == nil {
		t.Fatal("expected error for empty title id")
	}
}
