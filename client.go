package grid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
	return endpointsAt("https://api-op.grid.gg")
}

// CommercialEndpoints returns the standard commercial GRID endpoints (api.grid.gg).
func CommercialEndpoints() Endpoints {
	return endpointsAt("https://api.grid.gg")
}

func endpointsAt(baseURL string) Endpoints {
	return Endpoints{
		CentralData: baseURL + "/central-data/graphql",
		SeriesState: baseURL + "/live-data-feed/series-state/graphql",
		StatsFeed:   baseURL + "/statistics-feed/graphql",
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

// Client is the main entry point for querying GRID GraphQL APIs. Construct it
// with [NewClient] or [NewClientFromEnv] and reuse it across requests. A client
// can make concurrent requests as long as its service fields stay unchanged.
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

// WithHTTPClient copies the settings from a custom HTTP client for outbound
// requests. Its transport and cookie jar are shared with the GRID client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		if httpClient != nil {
			clientCopy := *httpClient
			c.httpClient = &clientCopy
		}
	}
}

// WithTimeout sets the timeout on the underlying HTTP client. Apply it after
// [WithHTTPClient] to override the supplied client's timeout.
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
		c.headers[http.CanonicalHeaderKey(key)] = value
	}
}

// NewClient creates a new GRID API client authenticated with apiKey.
// It defaults to Riot Open Access endpoints (api-op.grid.gg) and a 30-second
// HTTP timeout. Options are applied in the order supplied. Redirects to a
// different origin are returned as HTTP errors to protect the API key.
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:    apiKey,
		endpoints: OpenAccessEndpoints(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	checkRedirect := c.httpClient.CheckRedirect
	c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		origin := via[0].URL
		if !strings.EqualFold(req.URL.Scheme, origin.Scheme) || !strings.EqualFold(req.URL.Host, origin.Host) {
			return http.ErrUseLastResponse
		}
		if checkRedirect != nil {
			return checkRedirect(req, via)
		}
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
	c.CentralData = &CentralDataService{client: c}
	c.SeriesState = &SeriesStateService{client: c}
	c.Stats = &StatsService{client: c}
	return c
}

// NewClientFromEnv reads the API key from GRID_API_KEY and creates a client
// with the same defaults and options as [NewClient]. It returns
// [ErrMissingAPIKey] if the value is unset or blank. It makes no API requests.
func NewClientFromEnv(opts ...ClientOption) (*Client, error) {
	apiKey := strings.TrimSpace(os.Getenv("GRID_API_KEY"))
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	return NewClient(apiKey, opts...), nil
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
