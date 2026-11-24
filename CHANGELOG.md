# Changelog

## v0.2.0 (unreleased)

- Added `NewClientFromEnv` for `GRID_API_KEY` and documented Riot Open Access, commercial, and custom endpoint configuration.
- Added `ListTitles`, `ListPlayerRoles`, and `RecentSeries` to Central Data, plus paginated `List` iterators that fetch pages on demand and support forward and backward traversal.
- Pagination detects missing and repeated cursors, and the client returns cross-origin redirects as HTTP errors to protect credentials.
- Composed GraphQL selections now own their nested fields so callers can customize a selection without mutating shared presets.
- Added a title-agnostic quick start, credential-free CI, and an opt-in live test suite for Open Access and commercial accounts. To run the live suite, pass `-tags integration` alongside `GRID_API_KEY`.

## v0.1.0

- Initial standalone Go module with typed Central Data, Series State, and Stats Feed clients, configurable GraphQL selections, and raw query support.
