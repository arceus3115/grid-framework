# grid-framework

A lightweight Go client for the GRID GraphQL APIs (**Central Data**, **Series State**, and **Stats Feed**) with no third-party dependencies.

## Prerequisites

- **Go 1.25+** installed (`go version`)
- A **GRID API key** exported in your environment:

```bash
printf 'GRID API key: '
IFS= read -r -s GRID_API_KEY
printf '\n'
export GRID_API_KEY
```

## Install

Run this in your Go project's module directory. For a new project, run `go mod init` with your project's module path first.

```bash
go get github.com/arceus3115/grid-framework@v0.1.0
```

Import the package in your Go code. Its package name is `grid`.

```go
import "github.com/arceus3115/grid-framework"
```

For local development before release, create a Go workspace from your consumer project's module directory. Adjust the path to your clone of `grid-framework`.

```bash
go work init . ../grid-framework
```

If you already use a workspace, add the checkout with `go work use ../grid-framework`. The import path is the same in both cases.

## Quick Start

Save the following as `main.go` in your Go project and run it with `go run .`:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/arceus3115/grid-framework"
)

func main() {
	ctx := context.Background()

	apiKey := os.Getenv("GRID_API_KEY")
	if apiKey == "" {
		log.Fatal("GRID_API_KEY is not set")
	}

	// Defaults to Riot Open Access endpoints (api-op.grid.gg) with a 30s timeout.
	client := grid.NewClient(apiKey)

	// 1. Central Data: list titles
	titles, err := client.CentralData.Titles(ctx, grid.TitlesParams{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Titles:", len(titles))

	// 2. Central Data: fetch recent League of Legends series
	seriesPage, err := client.CentralData.AllSeries(ctx, grid.AllSeriesParams{
		Page:           grid.PageParams{First: grid.Int(10)},
		OrderBy:        grid.SeriesOrderByStartTimeScheduled,
		OrderDirection: grid.OrderDirectionDesc,
		Filter: &grid.SeriesFilter{
			TitleIDs: &grid.IDFilter{In: []string{grid.TitleIDLeagueOfLegends}},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range seriesPage.Nodes() {
		fmt.Println("Series:", s.ID, s.StartTimeScheduled)
	}

	// 3. Series State: inspect a series by ID
	state, err := client.SeriesState.Get(ctx, "2589176")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Series finished:", state.Finished)
}
```

For a browser frontend, call these methods from a Go backend and have the browser call your backend. Read `GRID_API_KEY` on the server.

## Client Configuration

`grid.NewClient(apiKey, opts...)` accepts functional options:

```go
client := grid.NewClient(
	os.Getenv("GRID_API_KEY"),
	grid.WithCommercialEndpoints(),       // Switch from api-op.grid.gg to api.grid.gg
	grid.WithTimeout(15*time.Second),     // Override default 30s HTTP timeout
	grid.WithHeader("User-Agent", "app"), // Add custom headers to every request
)
```

With a commercial GRID key that has Stats Feed access, use `grid.WithCommercialEndpoints()` and call methods such as `client.Stats.TeamStatistics`.

Available options:

- `WithCommercialEndpoints()` — use `https://api.grid.gg/...` endpoints instead of Riot Open Access (`https://api-op.grid.gg/...`).
- `WithEndpoints(Endpoints{...})` — override all three API URLs at once.
- `WithCentralDataEndpoint(url)` / `WithSeriesStateEndpoint(url)` / `WithStatsEndpoint(url)` — override a single API URL.
- `WithHTTPClient(hc)` — supply a custom `*http.Client`.
- `WithTimeout(d)` — set the HTTP client timeout.
- `WithHeader(key, value)` — attach a custom header to every request.

## API Overview

| Service | Methods |
| --- | --- |
| `client.CentralData` | `Titles`, `Title`, `AllSeries`, `Series`, `SeriesFormats`, `Teams`, `Team`, `Players`, `Player`, `PlayerRoles`, `PlayerRole`, `Tournaments`, `Tournament`, `Organizations`, `Organization`, `DataProviders`, `ContentCatalogEntities`, `ContentCatalogEntity`, `ContentCatalogVersions`, `ContentCatalogVersion`, `SeriesIDByExternalID`, `TeamIDByExternalID`, `PlayerIDByExternalID`, `TournamentIDByExternalID` |
| `client.SeriesState` | `Get` |
| `client.Stats` | `TeamStatistics`, `TeamGameStatistics`, `PlayerStatistics`, `SeriesStatistics`, `GameStatistics` |

For custom raw queries not covered by a helper, call `client.DoRaw(ctx, api, operationName, query, variables)` or `grid.Execute(ctx, client, api, spec, form)`.

## Pagination

Paginated Central Data queries accept `grid.PageParams` (use `First`/`After` for forward pagination or `Last`/`Before` for backward pagination; max page size is `grid.MaxPageSize` = 50) and return `grid.Connection[T]`.

Use `grid.CollectAll` to automatically follow `PageInfo.EndCursor` across multiple pages:

```go
allTeams, err := grid.CollectAll(ctx, 5, func(ctx context.Context, after *string) (grid.Connection[grid.Team], error) {
	return client.CentralData.Teams(ctx, grid.TeamsParams{
		Page: grid.PageParams{
			First: grid.Int(50),
			After: after,
		},
	})
})
```

Pass `maxPages <= 0` to fetch until `HasNextPage` is `false`.

## Modular Return Forms (`Form[T]`)

Every query has a `*WithForm` generic counterpart that accepts any `Form[T]`, decoupling *which GraphQL fields are requested* (`Selection`) from *how the JSON response is decoded into Go*.

### 1. Custom DTO + Composed Selection

```go
type MatchupCard struct {
	ID         string `json:"id"`
	StartTime  string `json:"startTimeScheduled"`
	Tournament struct {
		Name string `json:"name"`
	} `json:"tournament"`
}

matchupForm := grid.NewForm[MatchupCard](
	grid.Scalars("id", "startTimeScheduled").With(
		grid.Field("tournament", grid.Scalar("name")),
	),
)

page, err := grid.AllSeriesWithForm(ctx, client, grid.AllSeriesParams{
	Page: grid.PageParams{First: grid.Int(5)},
}, matchupForm)
```

### 2. Transforming a Form (`Transform`)

```go
teamNameForm := grid.Transform(grid.DefaultTeamForm(), func(t grid.Team) (string, error) {
	return t.Name, nil
})
name, err := grid.TeamWithForm(ctx, client, "47351", teamNameForm)
```

### 3. Inline Fragments & Dynamic Maps / Raw JSON

```go
lolStatsForm := grid.MapForm(
	grid.TeamGameStatisticsSelectionBasic.With(
		grid.InlineFragment("TeamGameStatisticsLol",
			grid.Field("baronPowerPlay", grid.Scalar("avg")),
		),
	),
)
rawMap, err := grid.TeamGameStatisticsWithForm(ctx, client, grid.TeamGameStatisticsParams{
	TeamID: "47351",
}, lolStatsForm)
```

Use `grid.RawForm(selection)` when you want `json.RawMessage` directly, or `grid.NewCustomForm(selection, decodeFn)` for custom unmarshaling logic.

## Error Handling

Use the built-in helpers or `errors.As` to inspect GraphQL (`*grid.APIError`) and HTTP (`*grid.HTTPError`) errors:

```go
series, err := client.CentralData.Series(ctx, "2589176")
if err != nil {
	switch {
	case grid.IsUnauthenticated(err):
		log.Fatal("invalid or missing GRID_API_KEY")
	case grid.IsRateLimited(err):
		log.Println("rate limited by GRID; retry with backoff")
	case grid.IsNotFound(err):
		log.Println("series not found")
	default:
		log.Fatal(err)
	}
}
```

## Running Tests

Run the framework's offline unit and `httptest` integration tests from the repository root:

```bash
env -u GRID_API_KEY go test -v -race ./...
go vet ./...
```

To also run the opt-in live integration tests against the GRID API, export `GRID_API_KEY` as shown in [Prerequisites](#prerequisites):

```bash
go test -v -race -count=1 ./...
```

## Release

Run the offline tests and commit the release changes. From the repository root, tag that commit and push the tag to GitHub.

```bash
git tag -a v0.1.0 -m "Release grid-framework v0.1.0"
git push origin v0.1.0
```
