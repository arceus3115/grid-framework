# grid-framework

A Go client for GRID Central Data, Series State, and Stats Feed. Typed methods handle authentication, GraphQL queries, and decoding with no third-party dependencies.

## Prerequisites

- **Go 1.25+** installed (`go version`)
- A **GRID API key** in the `GRID_API_KEY` environment variable. Get a key from [Riot Games Open Access powered by GRID](https://grid.gg/riot-open-access/) or use your commercial GRID key.

```bash
printf 'GRID API key: '
IFS= read -r -s GRID_API_KEY
printf '\n'
export GRID_API_KEY
```

## Install

If you are starting a new Go project, run `go mod init example.com/myapp` in its directory. Substitute your own module path.

The examples below use APIs added after `v0.1.0`. To use them before the next release, run this command in your Go project's module directory. Adjust the path to your checkout of `grid-framework`.

```bash
go work init . ../grid-framework
```

If you already use a workspace, add the checkout with `go work use ../grid-framework`. To install the latest release, run `go get github.com/arceus3115/grid-framework@latest`. The import path is the same in both cases, and the package name is `grid`.

## Quick Start

Save this as `main.go` and run `go run .`. The program finds a recent League of Legends series and fetches its state using the returned ID.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/arceus3115/grid-framework"
)

func main() {
	ctx := context.Background()

	client, err := grid.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	series, err := client.CentralData.RecentSeries(ctx, grid.TitleIDLeagueOfLegends, 1)
	if err != nil {
		log.Fatal(err)
	}
	if len(series) == 0 {
		fmt.Println("No recent series available")
		return
	}

	state, err := client.SeriesState.Get(ctx, series[0].ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Series %s finished: %t\n", series[0].ID, state.Finished)
}
```

To discover title IDs, call `client.CentralData.ListTitles(ctx)`. Run `go doc .` from the SDK checkout for the package overview.

For a browser frontend, call these methods from a Go backend and have the browser call your backend. Read `GRID_API_KEY` on the server.

## Client Configuration

`grid.NewClientFromEnv()` reads `GRID_API_KEY` and reports `grid.ErrMissingAPIKey` if it is unset or blank. Use `grid.NewClient(apiKey, opts...)` when your application already manages the key. Both default to Riot Open Access and a 30-second HTTP timeout. Reuse the client across requests and use contexts for request or workflow deadlines.

For a commercial key, use `grid.NewClientFromEnv(grid.WithCommercialEndpoints())`. Stats Feed methods such as `client.Stats.TeamStatistics` require Stats Feed access on the account. [GRID access plans](https://grid.gg/get-access/)

Both constructors apply options in the order supplied.

- `WithCommercialEndpoints()` — use `https://api.grid.gg/...` endpoints instead of Riot Open Access (`https://api-op.grid.gg/...`).
- `WithEndpoints(Endpoints{...})` — override all three API URLs at once.
- `WithCentralDataEndpoint(url)` / `WithSeriesStateEndpoint(url)` / `WithStatsEndpoint(url)` — override a single API URL.
- `WithHTTPClient(hc)` — copy a custom HTTP client's settings and share its transport and cookie jar.
- `WithTimeout(d)` — set the HTTP client timeout. Use it after `WithHTTPClient` to override that client's timeout.
- `WithHeader(key, value)` — attach a custom header to every request.

Use overrides for trusted GraphQL URLs. The client rejects cross-origin redirects to protect the API key. Configure a redirected URL directly with the appropriate endpoint option.

## API Overview

| Service | Methods |
| --- | --- |
| `client.CentralData` | `ListTitles`, `ListPlayerRoles`, `RecentSeries`, and the paginated `List` methods below. Use `Titles` or `PlayerRoles` for filtered results, `AllSeries` for one page of series, and singular methods such as `Series`, `Team`, or `Tournament` for lookups by ID. External ID helpers include `SeriesIDByExternalID` and `GameIDByExternalID`. |
| `client.SeriesState` | `Get` |
| `client.Stats` | `TeamStatistics`, `TeamGameStatistics`, `PlayerStatistics`, `SeriesStatistics`, `GameStatistics` |

For a custom GraphQL document, use `client.DoRaw(ctx, api, operationName, query, variables)`. To build and decode a custom root query, use `grid.Execute(ctx, client, api, spec, form)`.

## Pagination

Use `ListSeries`, `ListTeams`, `ListPlayers`, `ListTournaments`, `ListOrganizations`, `ListContentCatalogEntities`, or `ListContentCatalogVersions` to walk matching records. Each fetches a page when needed. Breaking the loop stops further requests.

```go
for team, err := range client.CentralData.ListTeams(ctx, grid.TeamsParams{}) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(team.ID, team.Name)
}
```

Zero-valued `PageParams` start at the first page with up to `grid.MaxPageSize` (50) records per request. Set `First` or `Last` to change the page size, or `After` or `Before` to resume from a cursor. `Last` or `Before` traverses backward and yields nodes in reverse order. An iterator yields a terminal error if a fetch fails or a cursor cannot advance.

Use methods such as `Teams` or `AllSeries` for a single `grid.Connection[T]` with page metadata and cursors. `grid.CollectAll` can collect pages from a custom forward fetch function with a page limit. A `maxPages` value of zero or less fetches all pages.

## Modular Return Forms (`Form[T]`)

Most typed queries have a `WithForm` counterpart that accepts a `Form[T]`. Use it to choose GraphQL fields with `Selection` and decode the result into your own Go type.

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

Run the offline unit and `httptest` integration tests from the repository root.

```bash
env -u GRID_API_KEY go test -v -race ./...
go vet ./...
```

To also run the live integration tests against the GRID API, export `GRID_API_KEY` as shown in [Prerequisites](#prerequisites).

```bash
go test -v -race -count=1 ./...
```

## Release

Run the offline tests, commit the release changes, and publish a new version tag for that commit.
