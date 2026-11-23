// Package grid queries the GRID Central Data, Series State, and Stats Feed
// GraphQL APIs using typed Go methods.
//
// Set GRID_API_KEY and use [NewClientFromEnv], or pass a key to [NewClient].
// Both constructors default to Riot Open Access. For a commercial key, pass
// [WithCommercialEndpoints]. Stats Feed requires access on the account.
//
// Start with [CentralDataService.ListTitles] or [CentralDataService.RecentSeries].
// Pass a series ID from Central Data to [SeriesStateService.Get] for its state.
// The paginated List methods handle cursors and fetch further pages as needed.
//
// For custom GraphQL selections, use a [Form] with an endpoint's WithForm
// function, or use [Client.DoRaw] to send a GraphQL document directly.
package grid
