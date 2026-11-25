# grid-framework

An independent Go SDK for GRID Central Data, Series State, and Stats Feed. Typed methods handle authentication, GraphQL queries, and decoding with no third-party dependencies.

[Quick start](#quick-start) · [API overview](#api-overview) · [Go package reference](https://pkg.go.dev/github.com/arceus3115/grid-framework) · [Quick reference for AI agents](#quick-reference-for-ai-agents)

## Install

Use Go 1.25 or later. If your project needs a `go.mod`, run `go mod init example.com/myapp` with your own module path first.

Run this from your Go module.

```bash
go get github.com/arceus3115/grid-framework@latest
```

The examples require `v0.2.0` or later. Until that release is published, use `go get github.com/arceus3115/grid-framework@main`.

Import `github.com/arceus3115/grid-framework` and refer to the package as `grid`.

## Quick start

Set `GRID_API_KEY` in the environment where your Go program runs. The example uses Riot Open Access endpoints. For a commercial key, pass `grid.WithCommercialEndpoints()` to `grid.NewClientFromEnv`. [Riot Games Open Access](https://grid.gg/riot-open-access/) · [GRID access plans](https://grid.gg/plans/)

In Bash or Zsh, you can enter the key without echoing it.

```bash
printf 'GRID API key: '
IFS= read -r -s GRID_API_KEY
printf '\n'
export GRID_API_KEY
```

Save this as `main.go`. Run `go run .` to list title IDs from GRID's catalog, then run `go run . TITLE_ID` to look up a recent series and its state. Series coverage depends on your key, and a catalog title can return an empty series list.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/arceus3115/grid-framework"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := grid.NewClientFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) == 1 {
		titles, err := client.CentralData.ListTitles(ctx)
		if err != nil {
			log.Fatal(err)
		}
		for _, title := range titles {
			fmt.Printf("%s\t%s\n", title.ID, title.Name)
		}
		return
	}

	series, err := client.CentralData.RecentSeries(ctx, os.Args[1], 5)
	if err != nil {
		log.Fatal(err)
	}
	for _, match := range series {
		state, err := client.SeriesState.Get(ctx, match.ID)
		if grid.IsNotFound(err) {
			continue
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Series %s finished: %t\n", match.ID, state.Finished)
		return
	}
	fmt.Println("No state available for recent series")
}
```

`RecentSeries` filters by scheduled start time, so a returned series can still be waiting for state data.

Keep the API key in your backend when building a browser frontend. [GRID authentication docs](https://docs.grid.gg/public/documentation/api-reference/getting-started/authentication)

## Client configuration

`grid.NewClientFromEnv()` reads `GRID_API_KEY` and reports `grid.ErrMissingAPIKey` if it is unset or blank. Use `grid.NewClient(apiKey, opts...)` when your application already manages the key. Both default to Riot Open Access and a 30-second HTTP timeout. Reuse the client across requests and use contexts for request or workflow deadlines.

Riot Open Access includes Central Data and Series State. Stats Feed requires separate access on your GRID account. [Riot Open Access coverage](https://grid.gg/riot-open-access/) · [GRID access plans](https://grid.gg/plans/)

Both constructors apply options in the order supplied.

- `WithCommercialEndpoints()` — use the commercial GraphQL endpoints at `https://api.grid.gg`.
- `WithEndpoints(Endpoints{...})` — override all three API URLs at once.
- `WithCentralDataEndpoint(url)` / `WithSeriesStateEndpoint(url)` / `WithStatsEndpoint(url)` — override a single API URL.
- `WithHTTPClient(hc)` — copy a custom HTTP client's settings and share its transport and cookie jar.
- `WithTimeout(d)` — set the HTTP client timeout. Use it after `WithHTTPClient` to override that client's timeout.
- `WithHeader(key, value)` — attach a custom header to every request.

Each configured endpoint receives the API key. Use trusted HTTPS URLs for overrides. The client rejects cross-origin redirects to protect the key. Configure a redirected URL directly with the appropriate endpoint option.

## API overview

| Service | Use it for |
| --- | --- |
| `client.CentralData` | Discover titles with `ListTitles`, find scheduled series with `RecentSeries` or `AllSeries`, and look up metadata with methods such as `Series`, `Team`, `Player`, and `Tournament`. Use `Titles` or `PlayerRoles` for filtered results. |
| `client.SeriesState` | Read the current or final state of one series with `Get(ctx, seriesID)`, including whether it started or finished, its games, and which teams won. |
| `client.Stats` | Query aggregates with `TeamStatistics` or `TeamGameStatistics` by team ID, `PlayerStatistics` by player ID, and `SeriesStatistics` or `GameStatistics` by title ID. |

Use Central Data IDs across services. If you have a provider's external ID, use an appropriate `*IDByExternalID` helper, such as `SeriesIDByExternalID` or `GameIDByExternalID`. These helpers return `*string`, so check for `nil` before dereferencing the result.

## Pagination

Use methods such as `Teams` or `AllSeries` for one page. These return a `grid.Connection[T]` with `Nodes()` for the records and `PageInfo` for pagination. This example fetches up to five teams for a `titleID` from `ListTitles`.

```go
params := grid.TeamsParams{
	Page:   grid.PageParams{First: grid.Int(5)},
	Filter: &grid.TeamFilter{TitleID: grid.String(titleID)},
}
page, err := client.CentralData.Teams(ctx, params)
if err != nil {
	log.Fatal(err)
}
for _, team := range page.Nodes() {
	fmt.Println(team.ID, team.Name)
}
```

To fetch the next page yourself, check `page.PageInfo.HasNextPage` and pass its `EndCursor` as `params.Page.After` on the next call. Use `First`/`After` for forward pagination or `Last`/`Before` for backward pagination. `First` and `Last` accept values from 1 through `grid.MaxPageSize` (50).

Use `ListSeries`, `ListTeams`, `ListPlayers`, `ListTournaments`, `ListOrganizations`, `ListContentCatalogEntities`, or `ListContentCatalogVersions` to walk matching records across pages. The following loop reuses the filter and page size from the `params` above.

```go
for team, err := range client.CentralData.ListTeams(ctx, params) {
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(team.ID, team.Name)
}
```

These iterators fetch pages as needed and stop making requests when you break the loop. `PageParams{}` starts at the first page and requests up to 50 records per page. `First` and `Last` set the page size. Backward iteration yields records in reverse order. A failed fetch or a cursor that cannot advance yields a terminal error.

`grid.CollectAll` collects records from a custom forward page fetcher. A `maxPages` value of zero or less fetches all pages.

## Stats Feed example

With a Stats Feed-enabled key, pass a team ID from Central Data as `teamID` to fetch its aggregate statistics.

```go
stats, err := client.Stats.TeamStatistics(ctx, grid.TeamStatisticsParams{
	TeamID: teamID,
	Filter: grid.TeamStatisticsFilter{TimeWindow: grid.TimeRangeLastYear},
})
if err != nil {
	log.Fatal(err)
}
if stats.Series != nil {
	fmt.Printf("Series in the past year: %d\n", stats.Series.Count)
}
```

## Custom GraphQL fields

Most typed queries have a package-level `WithForm` function that takes `ctx`, `client`, the endpoint's arguments, and a `grid.Form[T]`. Use the form's `Selection` to choose GraphQL fields and Go JSON tags to map them to your type.

### Select fields into a custom type

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
if err != nil {
	log.Fatal(err)
}
for _, card := range page.Nodes() {
	fmt.Println(card.ID, card.Tournament.Name)
}
```

### Transform a typed result

Use `Transform` to map a decoded result to another Go type. Here `teamID` is a GRID team ID from Central Data.

```go
teamNameForm := grid.Transform(grid.DefaultTeamForm(), func(t grid.Team) (string, error) {
	return t.Name, nil
})
name, err := grid.TeamWithForm(ctx, client, teamID, teamNameForm)
if err != nil {
	log.Fatal(err)
}
fmt.Println(name)
```

### Decode title-specific fields into a map

This selection uses a League of Legends fragment. Supply a League of Legends `teamID` and a client with Stats Feed access.

```go
lolStatsForm := grid.MapForm(
	grid.TeamGameStatisticsSelectionBasic.With(
		grid.InlineFragment("TeamGameStatisticsLol",
			grid.Field("baronPowerPlay", grid.Scalar("avg")),
		),
	),
)
rawMap, err := grid.TeamGameStatisticsWithForm(ctx, client, grid.TeamGameStatisticsParams{
	TeamID: teamID,
}, lolStatsForm)
if err != nil {
	log.Fatal(err)
}
fmt.Println(rawMap["baronPowerPlay"])
```

Use `grid.RawForm(selection)` when you want `json.RawMessage` directly, or `grid.NewCustomForm(selection, decodeFn)` for custom unmarshaling logic.

Use `grid.Execute(ctx, client, api, spec, form)` to build and decode a custom root query. To send a complete GraphQL document, call `client.DoRaw(ctx, api, operationName, query, variables)`. Both accept `grid.APICentralData`, `grid.APISeriesState`, or `grid.APIStatsFeed` as `api`.

## Error handling

Use the built-in helpers for common failures and `errors.As` to inspect `*grid.APIError` or `*grid.HTTPError`. `DoRaw` and the typed methods return an error if GRID sends GraphQL errors alongside data. The client returns rate-limit errors to the caller, which can decide when to retry.

```go
series, err := client.CentralData.Series(ctx, seriesID)
switch {
case err == nil:
	fmt.Println(series.ID)
case grid.IsUnauthenticated(err):
	log.Fatal("invalid or missing GRID_API_KEY")
case grid.IsRateLimited(err):
	log.Println("rate limited by GRID")
case grid.IsNotFound(err):
	log.Println("series not found")
default:
	log.Fatal(err)
}
```

## Quick reference for AI agents

- Import `github.com/arceus3115/grid-framework` as `grid`. Use `grid.NewClientFromEnv()` for Open Access or pass `grid.WithCommercialEndpoints()` for a commercial key. The constructor reads `GRID_API_KEY`. Use one client across calls and pass a context to each request.
- Start with the typed service methods. Use `CentralData.ListTitles` to discover title IDs, `CentralData.RecentSeries` to get series IDs, and `SeriesState.Get` to fetch state for one series. For aggregate stats, inspect the relevant `Stats` parameter type to see which ID it takes. Stats Feed access depends on the key.
- Pick the result shape before writing the loop. `AllSeries` and `Teams` return `Connection[T]` for a single page, with `Nodes()` and `PageInfo`. `ListSeries` and `ListTeams` return `iter.Seq2[T, error]` for traversal. Handle the error inside the loop and break once you have enough records.
- Use a package-level `WithForm` function for custom fields. For example, `grid.SeriesStateWithForm(ctx, client, seriesID, form)` decodes the fields selected by `form`. `grid.AllSeriesWithForm` takes a node form and returns `Connection[T]`. Use `grid.String`, `grid.Int`, and `grid.Bool` for optional pointer-valued inputs.
- Check signatures in the installed version before generating code. For example, run the following commands from the consuming Go module or the SDK checkout.

```bash
go doc github.com/arceus3115/grid-framework CentralDataService
go doc github.com/arceus3115/grid-framework SeriesStateWithForm
go doc github.com/arceus3115/grid-framework TeamStatisticsParams
```

In the SDK checkout, use [`example_test.go`](example_test.go) for consumer examples and [`endpoints_test.go`](endpoints_test.go) for request and response fixtures. Run the offline tests below to validate SDK changes without an API key.

## Development

Run the unit and local HTTP tests from the repository root. These are the checks used in CI and require no GRID credentials.

```bash
go test -race ./...
go vet ./...
```

To use a local SDK checkout from another module, run `go work init . ../grid-framework` from that module. If you already have a workspace, add the checkout with `go work use ../grid-framework`.

### Live integration tests

Export `GRID_API_KEY` as in [Quick start](#quick-start), then opt in with the `integration` build tag.

```bash
go test -tags integration -run '^TestLive' -v -count=1 ./...
```

The suite defaults to Riot Open Access endpoints and selects a title from recent series visible to your key. Set `GRID_TEST_ACCESS=commercial` for commercial endpoints or `GRID_TEST_TITLE_ID` to select a title explicitly. Set `GRID_TEST_STATS=1` when your key has Stats Feed access.

### Releases

See the [changelog](CHANGELOG.md) for release notes. From the repository root, run the development checks and compile the live suite with `go test -tags integration -run '^$' ./...`. Once the release commit passes CI, publish a new semantic version tag on that commit. Verify that Go can resolve the exact tag from a separate module. [Go module publishing guide](https://go.dev/doc/modules/publishing)

## License

The SDK is released under the [MIT License](LICENSE). Access to GRID data is governed by your GRID agreement. [GRID access plans](https://grid.gg/plans/)
