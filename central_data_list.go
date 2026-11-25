package grid

import (
	"context"
	"errors"
	"iter"
	"strings"
	"time"
)

// ListTitles returns the titles in GRID's Central Data catalog. Use
// [CentralDataService.RecentSeries] to find past series visible to your account for a title.
func (s *CentralDataService) ListTitles(ctx context.Context) ([]Title, error) {
	return s.Titles(ctx, TitlesParams{})
}

// ListPlayerRoles returns player roles available to the configured GRID account.
func (s *CentralDataService) ListPlayerRoles(ctx context.Context) ([]PlayerRole, error) {
	return s.PlayerRoles(ctx, PlayerRolesParams{})
}

// RecentSeries returns up to limit series for a title in descending scheduled
// start time, through the current time. Limit must be between 1 and [MaxPageSize].
func (s *CentralDataService) RecentSeries(ctx context.Context, titleID string, limit int) ([]Series, error) {
	if strings.TrimSpace(titleID) == "" {
		return nil, errors.New("grid central-data: title id is required")
	}
	page, err := s.AllSeries(ctx, AllSeriesParams{
		Page:           PageParams{First: Int(limit)},
		OrderBy:        SeriesOrderByStartTimeScheduled,
		OrderDirection: OrderDirectionDesc,
		Filter: &SeriesFilter{
			TitleIDs:           &IDFilter{In: []string{titleID}},
			StartTimeScheduled: &DateTimeFilter{LTE: String(time.Now().UTC().Format(time.RFC3339Nano))},
		},
	})
	if err != nil {
		return nil, err
	}
	return page.Nodes(), nil
}

// ListSeries yields matching series across pages. See [CentralDataService]
// for iterator behavior.
func (s *CentralDataService) ListSeries(ctx context.Context, params AllSeriesParams) iter.Seq2[Series, error] {
	return iterateNodes(ctx, params.Page, func(page PageParams) (Connection[Series], error) {
		request := params
		request.Page = page
		return s.AllSeries(ctx, request)
	})
}

// ListTeams yields matching teams across pages. See [CentralDataService]
// for iterator behavior.
func (s *CentralDataService) ListTeams(ctx context.Context, params TeamsParams) iter.Seq2[Team, error] {
	return iterateNodes(ctx, params.Page, func(page PageParams) (Connection[Team], error) {
		request := params
		request.Page = page
		return s.Teams(ctx, request)
	})
}

// ListPlayers yields matching players across pages. See [CentralDataService]
// for iterator behavior.
func (s *CentralDataService) ListPlayers(ctx context.Context, params PlayersParams) iter.Seq2[Player, error] {
	return iterateNodes(ctx, params.Page, func(page PageParams) (Connection[Player], error) {
		request := params
		request.Page = page
		return s.Players(ctx, request)
	})
}

// ListTournaments yields matching tournaments across pages. See [CentralDataService]
// for iterator behavior.
func (s *CentralDataService) ListTournaments(ctx context.Context, params TournamentsParams) iter.Seq2[Tournament, error] {
	return iterateNodes(ctx, params.Page, func(page PageParams) (Connection[Tournament], error) {
		request := params
		request.Page = page
		return s.Tournaments(ctx, request)
	})
}

// ListOrganizations yields matching organizations across pages. See [CentralDataService]
// for iterator behavior.
func (s *CentralDataService) ListOrganizations(ctx context.Context, params OrganizationsParams) iter.Seq2[Organization, error] {
	return iterateNodes(ctx, params.Page, func(page PageParams) (Connection[Organization], error) {
		request := params
		request.Page = page
		return s.Organizations(ctx, request)
	})
}

// ListContentCatalogEntities yields matching catalog entities across pages.
// See [CentralDataService] for iterator behavior.
func (s *CentralDataService) ListContentCatalogEntities(ctx context.Context, params ContentCatalogEntitiesParams) iter.Seq2[ContentCatalogEntity, error] {
	return iterateNodes(ctx, params.Page, func(page PageParams) (Connection[ContentCatalogEntity], error) {
		request := params
		request.Page = page
		return s.ContentCatalogEntities(ctx, request)
	})
}

// ListContentCatalogVersions yields matching catalog versions across pages.
// See [CentralDataService] for iterator behavior.
func (s *CentralDataService) ListContentCatalogVersions(ctx context.Context, params ContentCatalogVersionsParams) iter.Seq2[ContentCatalogVersion, error] {
	return iterateNodes(ctx, params.Page, func(page PageParams) (Connection[ContentCatalogVersion], error) {
		request := params
		request.Page = page
		return s.ContentCatalogVersions(ctx, request)
	})
}
