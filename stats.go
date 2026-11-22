package grid

import (
	"context"
	"fmt"
)

// StatsService provides convenience methods for the GRID Stats Feed GraphQL API.
// For custom/modular return forms, use the matching package-level *WithForm functions.
type StatsService struct {
	client *Client
}

// Stats Feed Enums.

type TimeRangeFilter string

const (
	TimeRangeLastWeek    TimeRangeFilter = "LAST_WEEK"
	TimeRangeLastMonth   TimeRangeFilter = "LAST_MONTH"
	TimeRangeLast3Months TimeRangeFilter = "LAST_3_MONTHS"
	TimeRangeLast6Months TimeRangeFilter = "LAST_6_MONTHS"
	TimeRangeLastYear    TimeRangeFilter = "LAST_YEAR"
)

type GameOrderField string

const (
	GameOrderFieldStartedAt GameOrderField = "STARTED_AT"
)

// Stats Feed Filter Inputs.

type StatsTournamentFilter struct {
	ID              *IDFilter `json:"id,omitempty"`
	IncludeChildren bool      `json:"includeChildren"`
}

type StatsDateTimeFilter struct {
	Period TimeRangeFilter `json:"period,omitempty"`
}

type TeamStatisticsFilter struct {
	TournamentIDs *IDFilter              `json:"tournamentIds,omitempty"`
	Tournament    *StatsTournamentFilter `json:"tournament,omitempty"`
	TimeWindow    TimeRangeFilter        `json:"timeWindow,omitempty"`
	StartedAt     *StatsDateTimeFilter   `json:"startedAt,omitempty"`
}

type PlayerStatisticsFilter struct {
	TournamentIDs *IDFilter              `json:"tournamentIds,omitempty"`
	Tournament    *StatsTournamentFilter `json:"tournament,omitempty"`
	TimeWindow    TimeRangeFilter        `json:"timeWindow,omitempty"`
	StartedAt     *StatsDateTimeFilter   `json:"startedAt,omitempty"`
}

type SeriesStatisticsFilter struct {
	Tournament *StatsTournamentFilter `json:"tournament,omitempty"`
	StartedAt  *StatsDateTimeFilter   `json:"startedAt,omitempty"`
}

type GameStatisticsVersionFilter struct {
	ID *IDFilter `json:"id,omitempty"`
}

type GameStatisticsFilter struct {
	Tournament *StatsTournamentFilter       `json:"tournament,omitempty"`
	StartedAt  *StatsDateTimeFilter         `json:"startedAt,omitempty"`
	Version    *GameStatisticsVersionFilter `json:"version,omitempty"`
}

type GameOrder struct {
	Field     GameOrderField `json:"field"`
	Direction OrderDirection `json:"direction"`
}

type CharacterFilter struct {
	ID   *IDFilter     `json:"id,omitempty"`
	Name *StringFilter `json:"name,omitempty"`
}

type MapFilter struct {
	Name *StringFilter `json:"name,omitempty"`
}

type TitleVersionFilter struct {
	Name *StringFilter `json:"name,omitempty"`
}

type ObjectiveFilter struct {
	Type           *StringFilter  `json:"type,omitempty"`
	CompletedFirst *BooleanFilter `json:"completedFirst,omitempty"`
}

type GamePlayerStateFilter struct {
	Character *CharacterFilter `json:"character,omitempty"`
}

type GameTeamStateFilter struct {
	ID         *IDFilter              `json:"id,omitempty"`
	Side       *StringFilter          `json:"side,omitempty"`
	Objectives *ObjectiveFilter       `json:"objectives,omitempty"`
	Players    *GamePlayerStateFilter `json:"players,omitempty"`
	FirstKill  *BooleanFilter         `json:"firstKill,omitempty"`
}

type SegmentStateFilter struct {
	Type     *StringFilter  `json:"type,omitempty"`
	WonFirst *BooleanFilter `json:"wonFirst,omitempty"`
}

type GameStateFilter struct {
	Map           *MapFilter             `json:"map,omitempty"`
	Teams         *GameTeamStateFilter   `json:"teams,omitempty"`
	TitleVersion  *TitleVersionFilter    `json:"titleVersion,omitempty"`
	Segments      *SegmentStateFilter    `json:"segments,omitempty"`
	TournamentIDs *IDFilter              `json:"tournamentIds,omitempty"`
	Tournament    *StatsTournamentFilter `json:"tournament,omitempty"`
	TimeWindow    TimeRangeFilter        `json:"timeWindow,omitempty"`
	StartedAt     *DateTimeFilter        `json:"startedAt,omitempty"`
	And           []GameStateFilter      `json:"and,omitempty"`
	Or            []GameStateFilter      `json:"or,omitempty"`
}

type GameSelection struct {
	Filter  *GameStateFilter `json:"filter,omitempty"`
	First   *int             `json:"first,omitempty"`
	OrderBy []GameOrder      `json:"orderBy,omitempty"`
}

// Stats Feed Parameter Structs.

type TeamStatisticsParams struct {
	TeamID string
	Filter TeamStatisticsFilter
}

type TeamGameStatisticsParams struct {
	TeamID    string
	Selection *GameSelection
}

type PlayerStatisticsParams struct {
	PlayerID string
	Filter   PlayerStatisticsFilter
}

type SeriesStatisticsParams struct {
	TitleID string
	Filter  SeriesStatisticsFilter
}

type GameStatisticsParams struct {
	TitleID string
	Filter  GameStatisticsFilter
}

// Stats Feed Domain Models.

type RateStatistic struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
	Avg float64 `json:"avg"`
}

type AggregateIntStatistic struct {
	Sum           int            `json:"sum"`
	Min           int            `json:"min"`
	Max           int            `json:"max"`
	Avg           float64        `json:"avg"`
	RatePerMinute *RateStatistic `json:"ratePerMinute,omitempty"`
}

type AggregateFloatStatistic struct {
	Sum           float64        `json:"sum"`
	Min           float64        `json:"min"`
	Max           float64        `json:"max"`
	Avg           float64        `json:"avg"`
	RatePerMinute *RateStatistic `json:"ratePerMinute,omitempty"`
}

type BooleanOccurrenceStatistic struct {
	Value      bool    `json:"value"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type DurationStatistic struct {
	Sum string `json:"sum"`
	Min string `json:"min"`
	Max string `json:"max"`
	Avg string `json:"avg"`
}

type TeamSeriesStats struct {
	Count    int                          `json:"count"`
	Kills    *AggregateIntStatistic       `json:"kills,omitempty"`
	Deaths   *AggregateIntStatistic       `json:"deaths,omitempty"`
	Won      []BooleanOccurrenceStatistic `json:"won,omitempty"`
	Duration *DurationStatistic           `json:"duration,omitempty"`
}

type TeamGameStats struct {
	Count               int                          `json:"count"`
	Kills               *AggregateIntStatistic       `json:"kills,omitempty"`
	Deaths              *AggregateIntStatistic       `json:"deaths,omitempty"`
	NetWorth            *AggregateIntStatistic       `json:"netWorth,omitempty"`
	StructuresDestroyed *AggregateIntStatistic       `json:"structuresDestroyed,omitempty"`
	Won                 []BooleanOccurrenceStatistic `json:"won,omitempty"`
}

type TeamStatistics struct {
	ID                   string           `json:"id"`
	AggregationSeriesIDs []string         `json:"aggregationSeriesIds,omitempty"`
	Series               *TeamSeriesStats `json:"series,omitempty"`
	Game                 *TeamGameStats   `json:"game,omitempty"`
}

type PlayerSeriesStats struct {
	Count            int                          `json:"count"`
	Kills            *AggregateIntStatistic       `json:"kills,omitempty"`
	Deaths           *AggregateIntStatistic       `json:"deaths,omitempty"`
	KillAssistsGiven *AggregateIntStatistic       `json:"killAssistsGiven,omitempty"`
	Won              []BooleanOccurrenceStatistic `json:"won,omitempty"`
}

type PlayerGameStats struct {
	Count            int                          `json:"count"`
	Kills            *AggregateIntStatistic       `json:"kills,omitempty"`
	Deaths           *AggregateIntStatistic       `json:"deaths,omitempty"`
	KillAssistsGiven *AggregateIntStatistic       `json:"killAssistsGiven,omitempty"`
	NetWorth         *AggregateIntStatistic       `json:"netWorth,omitempty"`
	Won              []BooleanOccurrenceStatistic `json:"won,omitempty"`
}

type PlayerStatistics struct {
	ID                   string             `json:"id"`
	AggregationSeriesIDs []string           `json:"aggregationSeriesIds,omitempty"`
	Series               *PlayerSeriesStats `json:"series,omitempty"`
	Game                 *PlayerGameStats   `json:"game,omitempty"`
}

type SeriesStatistics struct {
	AggregationSeriesIDs []string `json:"aggregationSeriesIds,omitempty"`
	Count                int      `json:"count"`
}

type GameStatistics struct {
	AggregationGameIDs []string `json:"aggregationGameIds,omitempty"`
	Count              int      `json:"count"`
}

// Composable Selection presets for Stats Feed.

var (
	AggregateIntSelection   = Scalars("sum", "min", "max", "avg")
	BoolOccurrenceSelection = Scalars("value", "count", "percentage")
	DurationSelection       = Scalars("sum", "min", "max", "avg")

	TeamStatisticsSelectionBasic = Select(
		Scalar("id"),
		Scalar("aggregationSeriesIds"),
		Field("series",
			Scalar("count"),
			FieldSelect("kills", AggregateIntSelection),
			FieldSelect("deaths", AggregateIntSelection),
			FieldSelect("won", BoolOccurrenceSelection),
		),
		Field("game",
			Scalar("count"),
			FieldSelect("kills", AggregateIntSelection),
			FieldSelect("deaths", AggregateIntSelection),
			FieldSelect("won", BoolOccurrenceSelection),
		),
	)

	TeamGameStatisticsSelectionBasic = Select(
		Scalar("count"),
		FieldSelect("kills", AggregateIntSelection),
		FieldSelect("deaths", AggregateIntSelection),
		FieldSelect("netWorth", AggregateIntSelection),
		FieldSelect("structuresDestroyed", AggregateIntSelection),
		FieldSelect("won", BoolOccurrenceSelection),
	)

	PlayerStatisticsSelectionBasic = Select(
		Scalar("id"),
		Scalar("aggregationSeriesIds"),
		Field("series",
			Scalar("count"),
			FieldSelect("kills", AggregateIntSelection),
			FieldSelect("deaths", AggregateIntSelection),
			FieldSelect("killAssistsGiven", AggregateIntSelection),
			FieldSelect("won", BoolOccurrenceSelection),
		),
		Field("game",
			Scalar("count"),
			FieldSelect("kills", AggregateIntSelection),
			FieldSelect("deaths", AggregateIntSelection),
			FieldSelect("killAssistsGiven", AggregateIntSelection),
			FieldSelect("won", BoolOccurrenceSelection),
		),
	)

	SeriesStatisticsSelectionBasic = Scalars("aggregationSeriesIds", "count")
	GameStatisticsSelectionBasic   = Scalars("aggregationGameIds", "count")
)

// Default typed return Forms for Stats Feed queries.

func DefaultTeamStatisticsForm() Form[TeamStatistics] {
	return NewForm[TeamStatistics](TeamStatisticsSelectionBasic)
}

func DefaultTeamGameStatisticsForm() Form[TeamGameStats] {
	return NewForm[TeamGameStats](TeamGameStatisticsSelectionBasic)
}

func DefaultPlayerStatisticsForm() Form[PlayerStatistics] {
	return NewForm[PlayerStatistics](PlayerStatisticsSelectionBasic)
}

func DefaultSeriesStatisticsForm() Form[SeriesStatistics] {
	return NewForm[SeriesStatistics](SeriesStatisticsSelectionBasic)
}

func DefaultGameStatisticsForm() Form[GameStatistics] {
	return NewForm[GameStatistics](GameStatisticsSelectionBasic)
}

// --- Stats Feed Query Functions ---

func (s *StatsService) TeamStatistics(ctx context.Context, params TeamStatisticsParams) (TeamStatistics, error) {
	return TeamStatisticsWithForm(ctx, s.client, params, DefaultTeamStatisticsForm())
}

func TeamStatisticsWithForm[T any](ctx context.Context, client *Client, params TeamStatisticsParams, form Form[T]) (T, error) {
	var zero T
	if params.TeamID == "" {
		return zero, fmt.Errorf("grid stats-feed: team id is required")
	}
	return Execute(ctx, client, APIStatsFeed, QuerySpec{
		OperationName: "TeamStatistics",
		FieldName:     "teamStatistics",
		Args: []Arg{
			{Name: "teamId", GraphQLType: "ID!", Value: params.TeamID},
			{Name: "filter", GraphQLType: "TeamStatisticsFilter!", Value: params.Filter},
		},
	}, form)
}

func (s *StatsService) TeamGameStatistics(ctx context.Context, params TeamGameStatisticsParams) (TeamGameStats, error) {
	return TeamGameStatisticsWithForm(ctx, s.client, params, DefaultTeamGameStatisticsForm())
}

func TeamGameStatisticsWithForm[T any](ctx context.Context, client *Client, params TeamGameStatisticsParams, form Form[T]) (T, error) {
	var zero T
	if params.TeamID == "" {
		return zero, fmt.Errorf("grid stats-feed: team id is required")
	}
	args := []Arg{{Name: "teamId", GraphQLType: "ID!", Value: params.TeamID}}
	if params.Selection != nil {
		args = append(args, Arg{Name: "selection", GraphQLType: "GameSelection", Value: params.Selection})
	}
	return Execute(ctx, client, APIStatsFeed, QuerySpec{
		OperationName: "TeamGameStatistics",
		FieldName:     "teamGameStatistics",
		Args:          args,
	}, form)
}

func (s *StatsService) PlayerStatistics(ctx context.Context, params PlayerStatisticsParams) (PlayerStatistics, error) {
	return PlayerStatisticsWithForm(ctx, s.client, params, DefaultPlayerStatisticsForm())
}

func PlayerStatisticsWithForm[T any](ctx context.Context, client *Client, params PlayerStatisticsParams, form Form[T]) (T, error) {
	var zero T
	if params.PlayerID == "" {
		return zero, fmt.Errorf("grid stats-feed: player id is required")
	}
	return Execute(ctx, client, APIStatsFeed, QuerySpec{
		OperationName: "PlayerStatistics",
		FieldName:     "playerStatistics",
		Args: []Arg{
			{Name: "playerId", GraphQLType: "ID!", Value: params.PlayerID},
			{Name: "filter", GraphQLType: "PlayerStatisticsFilter!", Value: params.Filter},
		},
	}, form)
}

func (s *StatsService) SeriesStatistics(ctx context.Context, params SeriesStatisticsParams) (SeriesStatistics, error) {
	return SeriesStatisticsWithForm(ctx, s.client, params, DefaultSeriesStatisticsForm())
}

func SeriesStatisticsWithForm[T any](ctx context.Context, client *Client, params SeriesStatisticsParams, form Form[T]) (T, error) {
	var zero T
	if params.TitleID == "" {
		return zero, fmt.Errorf("grid stats-feed: title id is required")
	}
	return Execute(ctx, client, APIStatsFeed, QuerySpec{
		OperationName: "SeriesStatistics",
		FieldName:     "seriesStatistics",
		Args: []Arg{
			{Name: "titleId", GraphQLType: "ID!", Value: params.TitleID},
			{Name: "filter", GraphQLType: "SeriesStatisticsFilter!", Value: params.Filter},
		},
	}, form)
}

func (s *StatsService) GameStatistics(ctx context.Context, params GameStatisticsParams) (GameStatistics, error) {
	return GameStatisticsWithForm(ctx, s.client, params, DefaultGameStatisticsForm())
}

func GameStatisticsWithForm[T any](ctx context.Context, client *Client, params GameStatisticsParams, form Form[T]) (T, error) {
	var zero T
	if params.TitleID == "" {
		return zero, fmt.Errorf("grid stats-feed: title id is required")
	}
	return Execute(ctx, client, APIStatsFeed, QuerySpec{
		OperationName: "GameStatistics",
		FieldName:     "gameStatistics",
		Args: []Arg{
			{Name: "titleId", GraphQLType: "ID!", Value: params.TitleID},
			{Name: "filter", GraphQLType: "GameStatisticsFilter!", Value: params.Filter},
		},
	}, form)
}
