package grid

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	defer req.Body.Close()
	return f(req)
}

func stubResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func TestNewClientFromEnv(t *testing.T) {
	for _, value := range []string{"", " \t\n"} {
		t.Setenv("GRID_API_KEY", value)
		client, err := NewClientFromEnv()
		if client != nil || !errors.Is(err, ErrMissingAPIKey) {
			t.Fatalf("NewClientFromEnv with blank key returned %v, %v", client, err)
		}
	}

	t.Setenv("GRID_API_KEY", " test-key\n")
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.URL.String(), CommercialEndpoints().CentralData; got != want {
			t.Fatalf("endpoint = %q, want %q", got, want)
		}
		if got := req.Header.Get("X-Api-Key"); got != "test-key" {
			t.Error("missing or incorrect API key header")
		}
		return stubResponse(http.StatusOK, `{"data":{"titles":[{"id":"3"}]}}`), nil
	})}
	client, err := NewClientFromEnv(WithCommercialEndpoints(), WithHTTPClient(httpClient))
	if err != nil {
		t.Fatal(err)
	}
	titles, err := client.CentralData.ListTitles(t.Context())
	if err != nil || len(titles) != 1 || titles[0].ID != "3" {
		t.Fatalf("ListTitles = %#v, %v", titles, err)
	}
}

func TestCustomHTTPClientSettings(t *testing.T) {
	httpClient := &http.Client{Timeout: time.Second}
	first := NewClient("key", WithHTTPClient(httpClient), WithTimeout(2*time.Second))
	second := NewClient("key", WithHTTPClient(httpClient))
	if first.httpClient.Timeout != 2*time.Second || second.httpClient.Timeout != time.Second {
		t.Fatalf("timeouts = %s and %s", first.httpClient.Timeout, second.httpClient.Timeout)
	}
	if httpClient.Timeout != time.Second || httpClient.CheckRedirect != nil {
		t.Fatal("construction changed the caller's HTTP client")
	}
}

func TestClientRedirects(t *testing.T) {
	for _, tc := range []struct {
		name     string
		location string
		blocked  bool
	}{
		{name: "same origin", location: "/canonical"},
		{name: "different host", location: "https://other.example/canonical", blocked: true},
		{name: "different port", location: "https://grid.example:8443/canonical", blocked: true},
		{name: "different scheme", location: "http://grid.example/canonical", blocked: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests, checks := 0, 0
			httpClient := &http.Client{
				CheckRedirect: func(*http.Request, []*http.Request) error {
					checks++
					return nil
				},
				Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					requests++
					if req.URL.Path == "/graphql" {
						resp := stubResponse(http.StatusTemporaryRedirect, "")
						resp.Header.Set("Location", tc.location)
						return resp, nil
					}
					if req.Header.Get("X-Api-Key") != "test-key" {
						t.Error("same-origin redirect lost the API key header")
					}
					return stubResponse(http.StatusOK, `{"data":{"titles":[]}}`), nil
				}),
			}
			client := NewClient("test-key", WithHTTPClient(httpClient), WithCentralDataEndpoint("https://grid.example/graphql"))
			_, err := client.CentralData.ListTitles(t.Context())
			if tc.blocked {
				var httpErr *HTTPError
				if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusTemporaryRedirect || requests != 1 || checks != 0 {
					t.Fatalf("blocked redirect returned err=%v, requests=%d, checks=%d", err, requests, checks)
				}
			} else if err != nil || requests != 2 || checks != 1 {
				t.Fatalf("same-origin redirect returned err=%v, requests=%d, checks=%d", err, requests, checks)
			}
		})
	}
}

func TestClientEndpointsAndHeaders(t *testing.T) {
	var gotAPIKey string
	var gotCustomHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("x-api-key")
		gotCustomHeader = r.Header.Get("X-Custom")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"titles":[{"id":"3","name":"League of Legends"}]}}`))
	}))
	defer srv.Close()

	client := NewClient("secret-key",
		WithCentralDataEndpoint(srv.URL),
		WithHeader("X-Custom", "example-value"),
	)

	titles, err := client.CentralData.Titles(context.Background(), TitlesParams{})
	if err != nil {
		t.Fatalf("Titles failed: %v", err)
	}
	if gotAPIKey != "secret-key" {
		t.Fatalf("expected x-api-key=secret-key, got %q", gotAPIKey)
	}
	if gotCustomHeader != "example-value" {
		t.Fatalf("expected X-Custom=example-value, got %q", gotCustomHeader)
	}
	if len(titles) != 1 || titles[0].ID != "3" {
		t.Fatalf("unexpected titles: %#v", titles)
	}
}

func TestClientGraphQLErrorHelpers(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		body      string
		checkFunc func(error) bool
	}{
		{
			name:   "unauthenticated",
			status: http.StatusOK,
			body: `{
				"errors": [{
					"message": "Requester unauthorized to make query",
					"extensions": {"errorType": "UNAUTHENTICATED"}
				}]
			}`,
			checkFunc: IsUnauthenticated,
		},
		{
			name:   "rate limited",
			status: http.StatusTooManyRequests,
			body: `{
				"errors": [{
					"message": "Too many requests",
					"extensions": {"errorType": "UNAVAILABLE", "errorDetail": "TOO_MANY_REQUESTS"}
				}]
			}`,
			checkFunc: IsRateLimited,
		},
		{
			name:   "not found",
			status: http.StatusOK,
			body: `{
				"errors": [{
					"message": "Entity not found",
					"extensions": {"errorType": "NOT_FOUND"}
				}]
			}`,
			checkFunc: IsNotFound,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := NewClient("key", WithCentralDataEndpoint(srv.URL))
			_, err := client.CentralData.Titles(context.Background(), TitlesParams{})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.checkFunc(err) {
				t.Fatalf("expected error helper to match %v", err)
			}
		})
	}
}

func TestPageParamsValidation(t *testing.T) {
	if err := (PageParams{First: Int(50)}).Validate(); err != nil {
		t.Fatalf("expected valid page params, got %v", err)
	}
	if err := (PageParams{First: Int(51)}).Validate(); err == nil {
		t.Fatal("expected error for first > 50")
	}
	if err := (PageParams{First: Int(10), Last: Int(10)}).Validate(); err == nil {
		t.Fatal("expected error when combining first and last")
	}
	if err := (PageParams{After: String("c1"), Before: String("c2")}).Validate(); err == nil {
		t.Fatal("expected error when combining after and before")
	}
}

func TestCollectAllPagination(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		calls++
		w.Header().Set("Content-Type", "application/json")
		if req.Variables["after"] == nil {
			_, _ = w.Write([]byte(`{
				"data": {
					"teams": {
						"totalCount": 3,
						"pageInfo": {"hasPreviousPage": false, "hasNextPage": true, "startCursor": "c1", "endCursor": "c2"},
						"edges": [
							{"cursor": "c1", "node": {"id": "1", "name": "T1"}},
							{"cursor": "c2", "node": {"id": "2", "name": "Gen.G"}}
						]
					}
				}
			}`))
			return
		}
		_, _ = w.Write([]byte(`{
			"data": {
				"teams": {
					"totalCount": 3,
					"pageInfo": {"hasPreviousPage": true, "hasNextPage": false, "startCursor": "c3", "endCursor": "c3"},
					"edges": [
						{"cursor": "c3", "node": {"id": "3", "name": "G2"}}
					]
				}
			}
		}`))
	}))
	defer srv.Close()

	client := NewClient("key", WithCentralDataEndpoint(srv.URL))
	all, err := CollectAll(context.Background(), 0, func(ctx context.Context, after *string) (Connection[Team], error) {
		return client.CentralData.Teams(ctx, TeamsParams{
			Page: PageParams{First: Int(2), After: after},
		})
	})
	if err != nil {
		t.Fatalf("CollectAll failed: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 page fetches, got %d", calls)
	}
	if len(all) != 3 || all[0].Name != "T1" || all[2].Name != "G2" {
		t.Fatalf("unexpected collected teams: %#v", all)
	}
}
