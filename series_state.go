package grid

import (
	"context"
	"fmt"
)

// SeriesStateService provides convenience methods for the GRID Series State GraphQL API.
// For custom/modular return forms, use SeriesStateWithForm.
type SeriesStateService struct {
	client *Client
}

// Series State Domain Models (matching the Series State GraphQL API).

type SeriesState struct {
	ID       string            `json:"id,omitempty"`
	Started  bool              `json:"started"`
	Finished bool              `json:"finished"`
	Valid    bool              `json:"valid,omitempty"`
	Format   string            `json:"format,omitempty"`
	Teams    []SeriesStateTeam `json:"teams,omitempty"`
	Games    []SeriesStateGame `json:"games,omitempty"`
}

type SeriesStateTeam struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Score  int    `json:"score,omitempty"`
	Won    bool   `json:"won"`
	Kills  int    `json:"kills,omitempty"`
	Deaths int    `json:"deaths,omitempty"`
}

type SeriesStateGame struct {
	ID             string                `json:"id,omitempty"`
	SequenceNumber int                   `json:"sequenceNumber,omitempty"`
	Started        bool                  `json:"started,omitempty"`
	Finished       bool                  `json:"finished,omitempty"`
	Teams          []SeriesStateGameTeam `json:"teams,omitempty"`
}

type SeriesStateGameTeam struct {
	ID      string                  `json:"id"`
	Name    string                  `json:"name,omitempty"`
	Side    string                  `json:"side,omitempty"`
	Won     bool                    `json:"won,omitempty"`
	Score   int                     `json:"score,omitempty"`
	Kills   int                     `json:"kills,omitempty"`
	Deaths  int                     `json:"deaths,omitempty"`
	Players []SeriesStateGamePlayer `json:"players,omitempty"`
}

type SeriesStateGamePlayer struct {
	ID               string                  `json:"id"`
	Name             string                  `json:"name"`
	Kills            int                     `json:"kills"`
	Deaths           int                     `json:"deaths"`
	KillAssistsGiven int                     `json:"killAssistsGiven,omitempty"`
	Character        *SeriesStateNamedEntity `json:"character,omitempty"`
}

type SeriesStateNamedEntity struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
}

// Composable Selection presets for Series State.

var (
	SeriesStateSelectionBasic = Select(
		Scalar("started"),
		Scalar("finished"),
		Field("teams", Scalar("id"), Scalar("name"), Scalar("won")),
	)

	SeriesStateSelectionWithGames = SeriesStateSelectionBasic.With(
		Field("games",
			Scalar("id"),
			Scalar("sequenceNumber"),
			Scalar("started"),
			Scalar("finished"),
			Field("teams",
				Scalar("id"),
				Scalar("name"),
				Scalar("side"),
				Scalar("won"),
				Field("players",
					Scalar("id"),
					Scalar("name"),
					Scalar("kills"),
					Scalar("deaths"),
				),
			),
		),
	)
)

// DefaultSeriesStateForm returns the default typed Form for seriesState queries.
func DefaultSeriesStateForm() Form[SeriesState] {
	return NewForm[SeriesState](SeriesStateSelectionWithGames)
}

// Get retrieves the current or final state of a series by ID using the default return form.
func (s *SeriesStateService) Get(ctx context.Context, seriesID string) (SeriesState, error) {
	return SeriesStateWithForm(ctx, s.client, seriesID, DefaultSeriesStateForm())
}

// SeriesStateWithForm retrieves a series state using any modular Form[T].
func SeriesStateWithForm[T any](ctx context.Context, client *Client, seriesID string, form Form[T]) (T, error) {
	var zero T
	if seriesID == "" {
		return zero, fmt.Errorf("grid series-state: series id is required")
	}
	return Execute(ctx, client, APISeriesState, QuerySpec{
		OperationName: "SeriesState",
		FieldName:     "seriesState",
		Args:          []Arg{{Name: "id", GraphQLType: "ID!", Value: seriesID}},
	}, form)
}
