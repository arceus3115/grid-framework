package grid

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func newLiveClient(t *testing.T) *Client {
	t.Helper()
	apiKey := os.Getenv("GRID_API_KEY")
	if apiKey == "" {
		t.Skip("GRID_API_KEY not set; skipping live GRID API integration test")
	}
	return NewClient(apiKey, WithTimeout(30*time.Second))
}

func liveCall[T any](t *testing.T, fn func() (T, error)) (T, error) {
	t.Helper()
	var lastErr error
	var zero T
	for attempt := 0; attempt < 4; attempt++ {
		res, err := fn()
		if err == nil || !IsRateLimited(err) {
			return res, err
		}
		lastErr = err
		wait := time.Duration(attempt+1) * 5 * time.Second
		t.Logf("hit GRID rate limit (%v); retrying in %s...", err, wait)
		time.Sleep(wait)
	}
	return zero, lastErr
}

func TestLiveCentralData(t *testing.T) {
	client := newLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	t.Run("Titles", func(t *testing.T) {
		titles, err := client.CentralData.Titles(ctx, TitlesParams{})
		if err != nil {
			t.Fatalf("Titles() error: %v", err)
		}
		if len(titles) == 0 {
			t.Fatal("Titles() returned 0 titles")
		}
		t.Logf("Titles returned: %d (first: id=%s name=%q)", len(titles), titles[0].ID, titles[0].Name)
	})

	t.Run("TitleByID", func(t *testing.T) {
		title, err := client.CentralData.Title(ctx, "3")
		if err != nil {
			t.Fatalf("Title(3) error: %v", err)
		}
		if title.ID != "3" {
			t.Fatalf("Title(3) got id=%q, want 3", title.ID)
		}
		t.Logf("Title(3): name=%q nameShortened=%q", title.Name, title.NameShortened)
	})

	t.Run("Tournaments", func(t *testing.T) {
		page, err := client.CentralData.Tournaments(ctx, TournamentsParams{
			Page: PageParams{First: Int(5)},
			Filter: &TournamentFilter{
				Title: &TournamentTitleFilter{ID: &IDFilter{In: []string{"3"}}},
			},
		})
		if err != nil {
			t.Fatalf("Tournaments() error: %v", err)
		}
		nodes := page.Nodes()
		t.Logf("Tournaments: totalCount=%d returned=%d", page.TotalCount, len(nodes))
		if len(nodes) > 0 {
			t.Logf("First tournament: id=%s name=%q", nodes[0].ID, nodes[0].Name)
			single, err := client.CentralData.Tournament(ctx, nodes[0].ID)
			if err != nil {
				t.Fatalf("Tournament(%s) error: %v", nodes[0].ID, err)
			}
			if single.ID != nodes[0].ID {
				t.Fatalf("Tournament(%s) got id=%q", nodes[0].ID, single.ID)
			}
		}
	})

	t.Run("AllSeriesAndSeriesByID", func(t *testing.T) {
		page, err := client.CentralData.AllSeries(ctx, AllSeriesParams{
			Page:           PageParams{First: Int(5)},
			OrderBy:        SeriesOrderByStartTimeScheduled,
			OrderDirection: OrderDirectionDesc,
			Filter:         &SeriesFilter{TitleID: String("3")},
		})
		if err != nil {
			t.Fatalf("AllSeries() error: %v", err)
		}
		nodes := page.Nodes()
		t.Logf("AllSeries: totalCount=%d returned=%d", page.TotalCount, len(nodes))
		if len(nodes) == 0 {
			t.Fatal("AllSeries() returned 0 series for LoL")
		}
		t.Logf("First series: id=%s start=%s teams=%d", nodes[0].ID, nodes[0].StartTimeScheduled, len(nodes[0].Teams))

		single, err := client.CentralData.Series(ctx, nodes[0].ID)
		if err != nil {
			t.Fatalf("Series(%s) error: %v", nodes[0].ID, err)
		}
		if single.ID != nodes[0].ID {
			t.Fatalf("Series(%s) got id=%q", nodes[0].ID, single.ID)
		}
	})

	t.Run("SeriesFormats", func(t *testing.T) {
		formats, err := client.CentralData.SeriesFormats(ctx)
		if err != nil {
			t.Fatalf("SeriesFormats() error: %v", err)
		}
		t.Logf("SeriesFormats: count=%d", len(formats))
	})

	t.Run("TeamsAndTeamByID", func(t *testing.T) {
		page, err := client.CentralData.Teams(ctx, TeamsParams{
			Page:   PageParams{First: Int(5)},
			Filter: &TeamFilter{TitleID: String("3")},
		})
		if err != nil {
			t.Fatalf("Teams() error: %v", err)
		}
		nodes := page.Nodes()
		t.Logf("Teams: totalCount=%d returned=%d", page.TotalCount, len(nodes))
		if len(nodes) > 0 {
			single, err := client.CentralData.Team(ctx, nodes[0].ID)
			if err != nil {
				t.Fatalf("Team(%s) error: %v", nodes[0].ID, err)
			}
			if single.ID != nodes[0].ID {
				t.Fatalf("Team(%s) got id=%q", nodes[0].ID, single.ID)
			}
		}
	})

	t.Run("PlayersAndPlayerByID", func(t *testing.T) {
		page, err := client.CentralData.Players(ctx, PlayersParams{
			Page:   PageParams{First: Int(5)},
			Filter: &PlayerFilter{TitleID: String("3")},
		})
		if err != nil {
			t.Fatalf("Players() error: %v", err)
		}
		nodes := page.Nodes()
		t.Logf("Players: totalCount=%d returned=%d", page.TotalCount, len(nodes))
		if len(nodes) > 0 {
			single, err := client.CentralData.Player(ctx, nodes[0].ID)
			if err != nil {
				t.Fatalf("Player(%s) error: %v", nodes[0].ID, err)
			}
			if single.ID != nodes[0].ID {
				t.Fatalf("Player(%s) got id=%q", nodes[0].ID, single.ID)
			}
		}
	})

	t.Run("PlayerRoles", func(t *testing.T) {
		roles, err := client.CentralData.PlayerRoles(ctx, PlayerRolesParams{})
		if err != nil {
			t.Fatalf("PlayerRoles() error: %v", err)
		}
		t.Logf("PlayerRoles: count=%d", len(roles))
	})

	t.Run("Organizations", func(t *testing.T) {
		page, err := client.CentralData.Organizations(ctx, OrganizationsParams{
			Page: PageParams{First: Int(5)},
		})
		if err != nil {
			t.Fatalf("Organizations() error: %v", err)
		}
		t.Logf("Organizations: totalCount=%d returned=%d", page.TotalCount, len(page.Nodes()))
	})

	t.Run("DataProviders", func(t *testing.T) {
		providers, err := client.CentralData.DataProviders(ctx)
		if err != nil {
			t.Fatalf("DataProviders() error: %v", err)
		}
		t.Logf("DataProviders: count=%d", len(providers))
	})

	t.Run("CollectAllPagination", func(t *testing.T) {
		items, err := liveCall(t, func() ([]Tournament, error) {
			return CollectAll(ctx, 2, func(ctx context.Context, after *string) (Connection[Tournament], error) {
				return client.CentralData.Tournaments(ctx, TournamentsParams{
					Page: PageParams{First: Int(2), After: after},
					Filter: &TournamentFilter{
						Title: &TournamentTitleFilter{ID: &IDFilter{In: []string{"3"}}},
					},
				})
			})
		})
		if err != nil {
			t.Fatalf("CollectAll(Tournaments) error: %v", err)
		}
		if len(items) < 2 {
			t.Fatalf("CollectAll(Tournaments) got %d items, expected at least 2", len(items))
		}
		t.Logf("CollectAll(2 pages x 2): collected %d tournaments", len(items))
	})
}

func TestLiveModularForms(t *testing.T) {
	client := newLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	t.Run("CustomDTOWithAllSeriesWithForm", func(t *testing.T) {
		type SeriesCard struct {
			ID         string `json:"id"`
			Tournament struct {
				Name string `json:"name"`
			} `json:"tournament"`
		}
		cardForm := NewForm[SeriesCard](Select(
			Scalar("id"),
			Field("tournament", Scalar("name")),
		))
		page, err := liveCall(t, func() (Connection[SeriesCard], error) {
			return AllSeriesWithForm(ctx, client, AllSeriesParams{
				Page:   PageParams{First: Int(3)},
				Filter: &SeriesFilter{TitleID: String("3")},
			}, cardForm)
		})
		if err != nil {
			t.Fatalf("AllSeriesWithForm() error: %v", err)
		}
		nodes := page.Nodes()
		if len(nodes) == 0 {
			t.Fatal("AllSeriesWithForm() returned 0 nodes")
		}
		t.Logf("SeriesCard[0]: id=%s tournament=%q", nodes[0].ID, nodes[0].Tournament.Name)
	})

	t.Run("TransformForm", func(t *testing.T) {
		titleMapForm := Transform(DefaultTitlesForm(), func(titles []Title) (map[string]string, error) {
			out := make(map[string]string, len(titles))
			for _, item := range titles {
				out[item.ID] = item.Name
			}
			return out, nil
		})
		m, err := liveCall(t, func() (map[string]string, error) {
			return TitlesWithForm(ctx, client, TitlesParams{}, titleMapForm)
		})
		if err != nil {
			t.Fatalf("TitlesWithForm(Transform) error: %v", err)
		}
		if m["3"] == "" {
			t.Fatalf("expected title ID 3 in transformed map, got %v", m)
		}
		t.Logf("Transformed titles map: %d entries (id 3=%q)", len(m), m["3"])
	})

	t.Run("RawAndMapForm", func(t *testing.T) {
		raw, err := liveCall(t, func() (json.RawMessage, error) {
			return TitleWithForm(ctx, client, "3", RawForm(TitleSelectionBasic))
		})
		if err != nil {
			t.Fatalf("TitleWithForm(RawForm) error: %v", err)
		}
		if !json.Valid(raw) {
			t.Fatalf("RawForm returned invalid JSON: %s", string(raw))
		}
		m, err := liveCall(t, func() (map[string]any, error) {
			return TitleWithForm(ctx, client, "3", MapForm(TitleSelectionBasic))
		})
		if err != nil {
			t.Fatalf("TitleWithForm(MapForm) error: %v", err)
		}
		if m["id"] != "3" {
			t.Fatalf("MapForm got id=%v, want 3", m["id"])
		}
	})
}

func TestLiveSeriesState(t *testing.T) {
	client := newLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Find a recent LoL series that already started in the past.
	now := time.Now().UTC().Format(time.RFC3339)
	page, err := liveCall(t, func() (Connection[Series], error) {
		return client.CentralData.AllSeries(ctx, AllSeriesParams{
			Page:           PageParams{First: Int(10)},
			OrderBy:        SeriesOrderByStartTimeScheduled,
			OrderDirection: OrderDirectionDesc,
			Filter: &SeriesFilter{
				TitleID: String("3"),
				StartTimeScheduled: &DateTimeFilter{
					LTE: String(now),
				},
			},
		})
	})
	if err != nil {
		t.Fatalf("AllSeries() error: %v", err)
	}
	nodes := page.Nodes()
	if len(nodes) == 0 {
		t.Fatal("no past LoL series found")
	}

	var tested bool
	for _, s := range nodes {
		state, err := client.SeriesState.Get(ctx, s.ID)
		if err != nil {
			t.Logf("SeriesState.Get(%s) error (trying next): %v", s.ID, err)
			continue
		}
		t.Logf("SeriesState(%s): started=%v finished=%v teams=%d games=%d",
			s.ID, state.Started, state.Finished, len(state.Teams), len(state.Games))
		tested = true

		basic, err := SeriesStateWithForm(ctx, client, s.ID, NewForm[SeriesState](SeriesStateSelectionBasic))
		if err != nil {
			t.Fatalf("SeriesStateWithForm(basic, %s) error: %v", s.ID, err)
		}
		t.Logf("SeriesState basic(%s): started=%v finished=%v teams=%d", s.ID, basic.Started, basic.Finished, len(basic.Teams))
		break
	}
	if !tested {
		t.Fatalf("none of the %d recent LoL series returned a SeriesState", len(nodes))
	}
}

func TestLiveStats(t *testing.T) {
	client := newLiveClient(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Grab a real LoL team ID and player ID from Central Data.
	teamsPage, err := liveCall(t, func() (Connection[Team], error) {
		return client.CentralData.Teams(ctx, TeamsParams{
			Page:   PageParams{First: Int(3)},
			Filter: &TeamFilter{TitleID: String("3")},
		})
	})
	if err != nil {
		t.Fatalf("Teams() error: %v", err)
	}
	teams := teamsPage.Nodes()
	if len(teams) == 0 {
		t.Fatal("no LoL teams returned")
	}

	t.Run("TeamStatistics", func(t *testing.T) {
		stats, err := client.Stats.TeamStatistics(ctx, TeamStatisticsParams{
			TeamID: teams[0].ID,
			Filter: TeamStatisticsFilter{TimeWindow: TimeRangeLastYear},
		})
		if err != nil {
			t.Fatalf("TeamStatistics(%s) error: %v", teams[0].ID, err)
		}
		t.Logf("TeamStatistics(%s): id=%s seriesCount=%d gameCount=%d",
			teams[0].ID, stats.ID, stats.Series.Count, stats.Game.Count)
	})

	t.Run("TeamGameStatistics", func(t *testing.T) {
		stats, err := client.Stats.TeamGameStatistics(ctx, TeamGameStatisticsParams{
			TeamID: teams[0].ID,
		})
		if err != nil {
			t.Fatalf("TeamGameStatistics(%s) error: %v", teams[0].ID, err)
		}
		t.Logf("TeamGameStatistics(%s): count=%d", teams[0].ID, stats.Count)
	})

	t.Run("SeriesStatistics", func(t *testing.T) {
		stats, err := client.Stats.SeriesStatistics(ctx, SeriesStatisticsParams{
			TitleID: "3",
			Filter: SeriesStatisticsFilter{
				StartedAt: &StatsDateTimeFilter{Period: TimeRangeLastMonth},
			},
		})
		if err != nil {
			t.Fatalf("SeriesStatistics(3) error: %v", err)
		}
		t.Logf("SeriesStatistics(3, LAST_MONTH): count=%d", stats.Count)
	})

	t.Run("GameStatistics", func(t *testing.T) {
		stats, err := client.Stats.GameStatistics(ctx, GameStatisticsParams{
			TitleID: "3",
			Filter:  GameStatisticsFilter{},
		})
		if err != nil {
			t.Fatalf("GameStatistics(3) error: %v", err)
		}
		t.Logf("GameStatistics(3): count=%d", stats.Count)
	})

	playersPage, err := liveCall(t, func() (Connection[Player], error) {
		return client.CentralData.Players(ctx, PlayersParams{
			Page:   PageParams{First: Int(3)},
			Filter: &PlayerFilter{TitleID: String("3")},
		})
	})
	if err != nil {
		t.Fatalf("Players() error: %v", err)
	}
	players := playersPage.Nodes()
	if len(players) > 0 {
		t.Run("PlayerStatistics", func(t *testing.T) {
			stats, err := client.Stats.PlayerStatistics(ctx, PlayerStatisticsParams{
				PlayerID: players[0].ID,
				Filter:   PlayerStatisticsFilter{TimeWindow: TimeRangeLastYear},
			})
			if err != nil {
				t.Fatalf("PlayerStatistics(%s) error: %v", players[0].ID, err)
			}
			t.Logf("PlayerStatistics(%s): id=%s seriesCount=%d gameCount=%d",
				players[0].ID, stats.ID, stats.Series.Count, stats.Game.Count)
		})
	}
}

func TestLiveErrorHandling(t *testing.T) {
	if os.Getenv("GRID_API_KEY") == "" {
		t.Skip("GRID_API_KEY not set; skipping live error handling test")
	}
	badClient := NewClient("invalid-api-key-for-testing", WithTimeout(15*time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := badClient.CentralData.Titles(ctx, TitlesParams{})
	if err == nil {
		t.Fatal("expected error with invalid API key, got nil")
	}
	if !IsUnauthenticated(err) {
		t.Fatalf("expected IsUnauthenticated(err)=true, got false; err=%v", err)
	}
	t.Logf("Invalid key properly detected as unauthenticated: %v", err)
}
