package grid

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

// integrationTitleID selects a title from past series visible to the test account.
func integrationTitleID(ctx context.Context, client *Client, configuredID string) (string, error) {
	if id := strings.TrimSpace(configuredID); id != "" {
		return id, nil
	}
	page, err := AllSeriesWithForm(ctx, client, AllSeriesParams{
		Page:           PageParams{First: Int(5)},
		OrderBy:        SeriesOrderByStartTimeScheduled,
		OrderDirection: OrderDirectionDesc,
		Filter: &SeriesFilter{
			StartTimeScheduled: &DateTimeFilter{LTE: String(time.Now().UTC().Format(time.RFC3339Nano))},
		},
	}, NewForm[Series](Select(Field("title", Scalar("id")))))
	if err != nil {
		return "", err
	}
	for _, edge := range page.Edges {
		if title := edge.Node.Title; title != nil && title.ID != "" {
			return title.ID, nil
		}
	}
	return "", errors.New("GRID returned no recent series with a title (set GRID_TEST_TITLE_ID to select a title explicitly)")
}

func TestIntegrationTitleSelection(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{
			name: "title from visible series",
			body: `{"data":{"allSeries":{"edges":[{"node":{"title":null}},{"node":{"title":{"id":"6"}}}]}}}`,
			want: "6",
		},
		{
			name: "empty series results",
			body: `{"data":{"allSeries":{"edges":[]}}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient("test-key", WithHTTPClient(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				var request graphQLRequest
				if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
					t.Fatal(err)
				}
				if request.OperationName != "AllSeries" {
					t.Fatalf("queried %q, want AllSeries", request.OperationName)
				}
				filter, _ := request.Variables["filter"].(map[string]any)
				if len(filter) != 1 || filter["startTimeScheduled"] == nil {
					t.Fatalf("discovery filter = %v, want only a scheduled start-time limit", filter)
				}
				return stubResponse(http.StatusOK, tc.body), nil
			})}))
			id, err := integrationTitleID(t.Context(), client, "")
			if id != tc.want || (err != nil) != (tc.want == "") {
				t.Fatalf("integrationTitleID() = %q, %v, want ID %q", id, err, tc.want)
			}
		})
	}

	t.Run("explicit override", func(t *testing.T) {
		id, err := integrationTitleID(t.Context(), nil, " 3 ")
		if err != nil || id != "3" {
			t.Fatalf("integrationTitleID() = %q, %v, want ID 3", id, err)
		}
	})
}
