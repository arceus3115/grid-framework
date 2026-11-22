package grid

import (
	"context"
	"fmt"
)

// CentralDataService provides convenience methods using default return forms
// for the GRID Central Data GraphQL API.
// For custom/modular return forms, use the matching package-level *WithForm functions.
type CentralDataService struct {
	client *Client
}

// Parameter structs for Central Data queries.

type TitlesParams struct {
	Filter *TitleFilter
}

type AllSeriesParams struct {
	Page           PageParams
	Filter         *SeriesFilter
	OrderBy        SeriesOrderBy
	OrderDirection OrderDirection
}

type TeamsParams struct {
	Page   PageParams
	Filter *TeamFilter
}

type PlayersParams struct {
	Page   PageParams
	Filter *PlayerFilter
}

type PlayerRolesParams struct {
	Filter *PlayerRoleFilter
}

type TournamentsParams struct {
	Page           PageParams
	Filter         *TournamentFilter
	OrderDirection OrderDirection
}

type OrganizationsParams struct {
	Page   PageParams
	Filter *OrganizationFilter
}

type ContentCatalogEntitiesParams struct {
	Page                    PageParams
	ContentCatalogVersionID *string
	Filter                  *ContentCatalogEntityFilter
}

type ContentCatalogEntityParams struct {
	ID                      string
	ContentCatalogVersionID *string
	AcceptLanguage          *string
}

type ContentCatalogVersionsParams struct {
	Page    PageParams
	Filter  *ContentCatalogVersionFilter
	OrderBy []ContentCatalogVersionOrder
}

// --- Titles ---

func (s *CentralDataService) Titles(ctx context.Context, params TitlesParams) ([]Title, error) {
	return TitlesWithForm(ctx, s.client, params, DefaultTitlesForm())
}

func TitlesWithForm[T any](ctx context.Context, client *Client, params TitlesParams, form Form[T]) (T, error) {
	var args []Arg
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "TitleFilter", Value: params.Filter})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Titles",
		FieldName:     "titles",
		Args:          args,
	}, form)
}

func (s *CentralDataService) Title(ctx context.Context, id string) (Title, error) {
	return TitleWithForm(ctx, s.client, id, DefaultTitleForm())
}

func TitleWithForm[T any](ctx context.Context, client *Client, id string, form Form[T]) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("grid central-data: title id is required")
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Title",
		FieldName:     "title",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: id}},
	}, form)
}

// --- Series ---

func (s *CentralDataService) AllSeries(ctx context.Context, params AllSeriesParams) (Connection[Series], error) {
	return AllSeriesWithForm(ctx, s.client, params, DefaultSeriesForm())
}

func AllSeriesWithForm[T any](ctx context.Context, client *Client, params AllSeriesParams, nodeForm Form[T]) (Connection[T], error) {
	if err := params.Page.Validate(); err != nil {
		return Connection[T]{}, err
	}
	args := appendPageArgs(nil, params.Page)
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "SeriesFilter", Value: params.Filter})
	}
	if params.OrderBy != "" {
		args = append(args, Arg{Name: "orderBy", GraphQLType: "SeriesOrderBy!", Value: params.OrderBy})
	}
	if params.OrderDirection != "" {
		args = append(args, Arg{Name: "orderDirection", GraphQLType: "OrderDirection!", Value: params.OrderDirection})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "AllSeries",
		FieldName:     "allSeries",
		Args:          args,
	}, ConnectionForm(nodeForm))
}

func (s *CentralDataService) Series(ctx context.Context, id string) (Series, error) {
	return SeriesWithForm(ctx, s.client, id, DefaultSeriesForm())
}

func SeriesWithForm[T any](ctx context.Context, client *Client, id string, form Form[T]) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("grid central-data: series id is required")
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Series",
		FieldName:     "series",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: id}},
	}, form)
}

func (s *CentralDataService) SeriesFormats(ctx context.Context) ([]SeriesFormat, error) {
	return SeriesFormatsWithForm(ctx, s.client, DefaultSeriesFormatListForm())
}

func SeriesFormatsWithForm[T any](ctx context.Context, client *Client, form Form[T]) (T, error) {
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "SeriesFormats",
		FieldName:     "seriesFormats",
	}, form)
}

// --- Teams ---

func (s *CentralDataService) Teams(ctx context.Context, params TeamsParams) (Connection[Team], error) {
	return TeamsWithForm(ctx, s.client, params, DefaultTeamForm())
}

func TeamsWithForm[T any](ctx context.Context, client *Client, params TeamsParams, nodeForm Form[T]) (Connection[T], error) {
	if err := params.Page.Validate(); err != nil {
		return Connection[T]{}, err
	}
	args := appendPageArgs(nil, params.Page)
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "TeamFilter", Value: params.Filter})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Teams",
		FieldName:     "teams",
		Args:          args,
	}, ConnectionForm(nodeForm))
}

func (s *CentralDataService) Team(ctx context.Context, id string) (Team, error) {
	return TeamWithForm(ctx, s.client, id, DefaultTeamForm())
}

func TeamWithForm[T any](ctx context.Context, client *Client, id string, form Form[T]) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("grid central-data: team id is required")
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Team",
		FieldName:     "team",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: id}},
	}, form)
}

// --- Players & Roles ---

func (s *CentralDataService) Players(ctx context.Context, params PlayersParams) (Connection[Player], error) {
	return PlayersWithForm(ctx, s.client, params, DefaultPlayerForm())
}

func PlayersWithForm[T any](ctx context.Context, client *Client, params PlayersParams, nodeForm Form[T]) (Connection[T], error) {
	if err := params.Page.Validate(); err != nil {
		return Connection[T]{}, err
	}
	args := appendPageArgs(nil, params.Page)
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "PlayerFilter", Value: params.Filter})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Players",
		FieldName:     "players",
		Args:          args,
	}, ConnectionForm(nodeForm))
}

func (s *CentralDataService) Player(ctx context.Context, id string) (Player, error) {
	return PlayerWithForm(ctx, s.client, id, DefaultPlayerForm())
}

func PlayerWithForm[T any](ctx context.Context, client *Client, id string, form Form[T]) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("grid central-data: player id is required")
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Player",
		FieldName:     "player",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: id}},
	}, form)
}

func (s *CentralDataService) PlayerRoles(ctx context.Context, params PlayerRolesParams) ([]PlayerRole, error) {
	return PlayerRolesWithForm(ctx, s.client, params, DefaultPlayerRolesForm())
}

func PlayerRolesWithForm[T any](ctx context.Context, client *Client, params PlayerRolesParams, form Form[T]) (T, error) {
	var args []Arg
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "PlayerRoleFilter", Value: params.Filter})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "PlayerRoles",
		FieldName:     "playerRoles",
		Args:          args,
	}, form)
}

func (s *CentralDataService) PlayerRole(ctx context.Context, id string) (PlayerRole, error) {
	return PlayerRoleWithForm(ctx, s.client, id, DefaultPlayerRoleForm())
}

func PlayerRoleWithForm[T any](ctx context.Context, client *Client, id string, form Form[T]) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("grid central-data: player role id is required")
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "PlayerRole",
		FieldName:     "playerRole",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: id}},
	}, form)
}

// --- Tournaments ---

func (s *CentralDataService) Tournaments(ctx context.Context, params TournamentsParams) (Connection[Tournament], error) {
	return TournamentsWithForm(ctx, s.client, params, DefaultTournamentForm())
}

func TournamentsWithForm[T any](ctx context.Context, client *Client, params TournamentsParams, nodeForm Form[T]) (Connection[T], error) {
	if err := params.Page.Validate(); err != nil {
		return Connection[T]{}, err
	}
	args := appendPageArgs(nil, params.Page)
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "TournamentFilter", Value: params.Filter})
	}
	if params.OrderDirection != "" {
		args = append(args, Arg{Name: "orderDirection", GraphQLType: "OrderDirection!", Value: params.OrderDirection})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Tournaments",
		FieldName:     "tournaments",
		Args:          args,
	}, ConnectionForm(nodeForm))
}

func (s *CentralDataService) Tournament(ctx context.Context, id string) (Tournament, error) {
	return TournamentWithForm(ctx, s.client, id, DefaultTournamentForm())
}

func TournamentWithForm[T any](ctx context.Context, client *Client, id string, form Form[T]) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("grid central-data: tournament id is required")
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Tournament",
		FieldName:     "tournament",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: id}},
	}, form)
}

// --- Organizations & Data Providers ---

func (s *CentralDataService) Organizations(ctx context.Context, params OrganizationsParams) (Connection[Organization], error) {
	return OrganizationsWithForm(ctx, s.client, params, DefaultOrganizationForm())
}

func OrganizationsWithForm[T any](ctx context.Context, client *Client, params OrganizationsParams, nodeForm Form[T]) (Connection[T], error) {
	if err := params.Page.Validate(); err != nil {
		return Connection[T]{}, err
	}
	args := appendPageArgs(nil, params.Page)
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "OrganizationFilter", Value: params.Filter})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Organizations",
		FieldName:     "organizations",
		Args:          args,
	}, ConnectionForm(nodeForm))
}

func (s *CentralDataService) Organization(ctx context.Context, id string) (Organization, error) {
	return OrganizationWithForm(ctx, s.client, id, DefaultOrganizationForm())
}

func OrganizationWithForm[T any](ctx context.Context, client *Client, id string, form Form[T]) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("grid central-data: organization id is required")
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "Organization",
		FieldName:     "organization",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: id}},
	}, form)
}

func (s *CentralDataService) DataProviders(ctx context.Context) ([]DataProvider, error) {
	return Execute(ctx, s.client, APICentralData, QuerySpec{
		OperationName: "DataProviders",
		FieldName:     "dataProviders",
	}, DefaultDataProvidersForm())
}

// --- Content Catalog ---

func (s *CentralDataService) ContentCatalogEntities(ctx context.Context, params ContentCatalogEntitiesParams) (Connection[ContentCatalogEntity], error) {
	return ContentCatalogEntitiesWithForm(ctx, s.client, params, DefaultContentCatalogEntityForm())
}

func ContentCatalogEntitiesWithForm[T any](ctx context.Context, client *Client, params ContentCatalogEntitiesParams, nodeForm Form[T]) (Connection[T], error) {
	if err := params.Page.Validate(); err != nil {
		return Connection[T]{}, err
	}
	args := appendPageArgs(nil, params.Page)
	if params.ContentCatalogVersionID != nil && *params.ContentCatalogVersionID != "" {
		args = append(args, Arg{Name: "contentCatalogVersionId", GraphQLType: "ID", Value: *params.ContentCatalogVersionID})
	}
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "ContentCatalogEntityFilter", Value: params.Filter})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "ContentCatalogEntities",
		FieldName:     "contentCatalogEntities",
		Args:          args,
	}, ConnectionForm(nodeForm))
}

func (s *CentralDataService) ContentCatalogEntity(ctx context.Context, params ContentCatalogEntityParams) (ContentCatalogEntity, error) {
	return ContentCatalogEntityWithForm(ctx, s.client, params, DefaultContentCatalogEntityForm())
}

func ContentCatalogEntityWithForm[T any](ctx context.Context, client *Client, params ContentCatalogEntityParams, form Form[T]) (T, error) {
	var zero T
	if params.ID == "" {
		return zero, fmt.Errorf("grid central-data: content catalog entity id is required")
	}
	args := []Arg{{Name: "id", GraphQLType: "ID!", Value: params.ID}}
	if params.ContentCatalogVersionID != nil && *params.ContentCatalogVersionID != "" {
		args = append(args, Arg{Name: "contentCatalogVersionId", GraphQLType: "ID", Value: *params.ContentCatalogVersionID})
	}
	if params.AcceptLanguage != nil && *params.AcceptLanguage != "" {
		args = append(args, Arg{Name: "acceptLanguage", GraphQLType: "Locale", Value: *params.AcceptLanguage})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "ContentCatalogEntity",
		FieldName:     "contentCatalogEntity",
		Args:          args,
	}, form)
}

func (s *CentralDataService) ContentCatalogVersions(ctx context.Context, params ContentCatalogVersionsParams) (Connection[ContentCatalogVersion], error) {
	return ContentCatalogVersionsWithForm(ctx, s.client, params, DefaultContentCatalogVersionForm())
}

func ContentCatalogVersionsWithForm[T any](ctx context.Context, client *Client, params ContentCatalogVersionsParams, nodeForm Form[T]) (Connection[T], error) {
	if err := params.Page.Validate(); err != nil {
		return Connection[T]{}, err
	}
	args := appendPageArgs(nil, params.Page)
	if params.Filter != nil {
		args = append(args, Arg{Name: "filter", GraphQLType: "ContentCatalogVersionFilter", Value: params.Filter})
	}
	if len(params.OrderBy) > 0 {
		args = append(args, Arg{Name: "orderBy", GraphQLType: "[ContentCatalogVersionOrder!]", Value: params.OrderBy})
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "ContentCatalogVersions",
		FieldName:     "contentCatalogVersions",
		Args:          args,
	}, ConnectionForm(nodeForm))
}

func (s *CentralDataService) ContentCatalogVersion(ctx context.Context, id string) (ContentCatalogVersion, error) {
	return ContentCatalogVersionWithForm(ctx, s.client, id, DefaultContentCatalogVersionForm())
}

func ContentCatalogVersionWithForm[T any](ctx context.Context, client *Client, id string, form Form[T]) (T, error) {
	var zero T
	if id == "" {
		return zero, fmt.Errorf("grid central-data: content catalog version id is required")
	}
	return Execute(ctx, client, APICentralData, QuerySpec{
		OperationName: "ContentCatalogVersion",
		FieldName:     "contentCatalogVersion",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: id}},
	}, form)
}

// --- External ID Lookups ---

func (s *CentralDataService) SeriesIDByExternalID(ctx context.Context, dataProviderName, externalSeriesID string) (*string, error) {
	return Execute(ctx, s.client, APICentralData, QuerySpec{
		OperationName: "SeriesIdByExternalId",
		FieldName:     "seriesIdByExternalId",
		Args: []Arg{
			{Name: "dataProviderName", GraphQLType: "String!", Value: dataProviderName},
			{Name: "externalSeriesId", GraphQLType: "ID!", Value: externalSeriesID},
		},
	}, ScalarForm[*string]())
}

func (s *CentralDataService) TeamIDByExternalID(ctx context.Context, dataProviderName, externalTeamID, titleID string) (*string, error) {
	return Execute(ctx, s.client, APICentralData, QuerySpec{
		OperationName: "TeamIdByExternalId",
		FieldName:     "teamIdByExternalId",
		Args: []Arg{
			{Name: "dataProviderName", GraphQLType: "String!", Value: dataProviderName},
			{Name: "externalTeamId", GraphQLType: "ID!", Value: externalTeamID},
			{Name: "titleId", GraphQLType: "ID!", Value: titleID},
		},
	}, ScalarForm[*string]())
}

func (s *CentralDataService) PlayerIDByExternalID(ctx context.Context, dataProviderName, externalPlayerID string, titleID *string) (*string, error) {
	args := []Arg{
		{Name: "dataProviderName", GraphQLType: "String!", Value: dataProviderName},
		{Name: "externalPlayerId", GraphQLType: "ID!", Value: externalPlayerID},
	}
	if titleID != nil && *titleID != "" {
		args = append(args, Arg{Name: "titleId", GraphQLType: "ID", Value: *titleID})
	}
	return Execute(ctx, s.client, APICentralData, QuerySpec{
		OperationName: "PlayerIdByExternalId",
		FieldName:     "playerIdByExternalId",
		Args:          args,
	}, ScalarForm[*string]())
}

func (s *CentralDataService) TournamentIDByExternalID(ctx context.Context, dataProviderName, externalTournamentID string) (*string, error) {
	return Execute(ctx, s.client, APICentralData, QuerySpec{
		OperationName: "TournamentIdByExternalId",
		FieldName:     "tournamentIdByExternalId",
		Args: []Arg{
			{Name: "dataProviderName", GraphQLType: "String!", Value: dataProviderName},
			{Name: "externalTournamentId", GraphQLType: "ID!", Value: externalTournamentID},
		},
	}, ScalarForm[*string]())
}

func (s *CentralDataService) GameIDByExternalID(ctx context.Context, dataProviderName, externalGameID string) (*string, error) {
	return Execute(ctx, s.client, APICentralData, QuerySpec{
		OperationName: "GameIdByExternalId",
		FieldName:     "gameIdByExternalId",
		Args: []Arg{
			{Name: "dataProviderName", GraphQLType: "String!", Value: dataProviderName},
			{Name: "externalGameId", GraphQLType: "ID!", Value: externalGameID},
		},
	}, ScalarForm[*string]())
}
