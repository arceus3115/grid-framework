package grid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// API identifies one of the GRID GraphQL APIs.
type API string

const (
	APICentralData API = "central-data"
	APISeriesState API = "series-state"
	APIStatsFeed   API = "stats-feed"
)

// Known GRID Title IDs documented for Riot Open Access.
const (
	TitleIDLeagueOfLegends = "3"
	TitleIDValorant        = "6"
)

// Endpoints holds the GraphQL URLs for each GRID API.
type Endpoints struct {
	CentralData string
	SeriesState string
	StatsFeed   string
}

// OpenAccessEndpoints returns the default Riot Open Access GRID endpoints (api-op.grid.gg).
func OpenAccessEndpoints() Endpoints {
	return Endpoints{
		CentralData: "https://api-op.grid.gg/central-data/graphql",
		SeriesState: "https://api-op.grid.gg/live-data-feed/series-state/graphql",
		StatsFeed:   "https://api-op.grid.gg/statistics-feed/graphql",
	}
}

// CommercialEndpoints returns the standard commercial GRID endpoints (api.grid.gg).
func CommercialEndpoints() Endpoints {
	return Endpoints{
		CentralData: "https://api.grid.gg/central-data/graphql",
		SeriesState: "https://api.grid.gg/live-data-feed/series-state/graphql",
		StatsFeed:   "https://api.grid.gg/statistics-feed/graphql",
	}
}

// URLFor returns the configured endpoint URL for the given API.
func (e Endpoints) URLFor(api API) (string, error) {
	switch api {
	case APICentralData:
		if e.CentralData == "" {
			return "", fmt.Errorf("grid client: central-data endpoint is not configured")
		}
		return e.CentralData, nil
	case APISeriesState:
		if e.SeriesState == "" {
			return "", fmt.Errorf("grid client: series-state endpoint is not configured")
		}
		return e.SeriesState, nil
	case APIStatsFeed:
		if e.StatsFeed == "" {
			return "", fmt.Errorf("grid client: stats-feed endpoint is not configured")
		}
		return e.StatsFeed, nil
	default:
		return "", fmt.Errorf("grid client: unknown api %q", api)
	}
}

// Client is the main entry point for querying GRID GraphQL APIs.
type Client struct {
	apiKey     string
	endpoints  Endpoints
	httpClient *http.Client
	headers    map[string]string

	// CentralData provides typed helper methods for the Central Data GraphQL API.
	CentralData *CentralDataService
	// SeriesState provides typed helper methods for the Series State GraphQL API.
	SeriesState *SeriesStateService
	// Stats provides typed helper methods for the Stats Feed GraphQL API.
	Stats *StatsService
}

// ClientOption configures a Client instance.
type ClientOption func(*Client)

// WithEndpoints overrides all API endpoints.
func WithEndpoints(endpoints Endpoints) ClientOption {
	return func(c *Client) {
		c.endpoints = endpoints
	}
}

// WithCommercialEndpoints configures the client to use commercial api.grid.gg endpoints.
func WithCommercialEndpoints() ClientOption {
	return WithEndpoints(CommercialEndpoints())
}

// WithCentralDataEndpoint overrides only the Central Data GraphQL endpoint URL.
func WithCentralDataEndpoint(url string) ClientOption {
	return func(c *Client) {
		c.endpoints.CentralData = url
	}
}

// WithSeriesStateEndpoint overrides only the Series State GraphQL endpoint URL.
func WithSeriesStateEndpoint(url string) ClientOption {
	return func(c *Client) {
		c.endpoints.SeriesState = url
	}
}

// WithStatsEndpoint overrides only the Stats Feed GraphQL endpoint URL.
func WithStatsEndpoint(url string) ClientOption {
	return func(c *Client) {
		c.endpoints.StatsFeed = url
	}
}

// WithHTTPClient sets a custom *http.Client for outbound requests.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			c.httpClient = httpClient
		}
	}
}

// WithTimeout sets the timeout on the underlying HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithHeader adds a custom HTTP header sent on every GRID request.
func WithHeader(key, value string) ClientOption {
	return func(c *Client) {
		if c.headers == nil {
			c.headers = make(map[string]string)
		}
		c.headers[key] = value
	}
}

// NewClient creates a new GRID API client authenticated with apiKey.
// By default it targets the Riot Open Access endpoints (api-op.grid.gg).
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:    apiKey,
		endpoints: OpenAccessEndpoints(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	c.CentralData = &CentralDataService{client: c}
	c.SeriesState = &SeriesStateService{client: c}
	c.Stats = &StatsService{client: c}
	return c
}

type graphQLRequest struct {
	OperationName string         `json:"operationName,omitempty"`
	Query         string         `json:"query"`
	Variables     map[string]any `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []GraphQLError             `json:"errors"`
}

// DoRaw executes a raw GraphQL query against the selected GRID API and returns the top-level data map.
func (c *Client) DoRaw(
	ctx context.Context,
	api API,
	operationName string,
	query string,
	variables map[string]any,
) (map[string]json.RawMessage, error) {
	endpoint, err := c.endpoints.URLFor(api)
	if err != nil {
		return nil, err
	}

	payload := graphQLRequest{
		OperationName: operationName,
		Query:         query,
		Variables:     variables,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("grid client: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("grid client: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("grid client: send %s request: %w", api, err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("grid client: read %s response: %w", api, err)
	}

	var envelope graphQLResponse
	if unmarshalErr := json.Unmarshal(respBytes, &envelope); unmarshalErr != nil {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, &HTTPError{
				API:        api,
				StatusCode: resp.StatusCode,
				Body:       string(respBytes),
			}
		}
		return nil, fmt.Errorf("grid client: decode %s response: %w", api, unmarshalErr)
	}

	if len(envelope.Errors) > 0 {
		return nil, &APIError{
			API:        api,
			StatusCode: resp.StatusCode,
			Errors:     envelope.Errors,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{
			API:        api,
			StatusCode: resp.StatusCode,
			Body:       string(respBytes),
		}
	}

	return envelope.Data, nil
}
