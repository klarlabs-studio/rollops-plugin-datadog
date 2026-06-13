// Package datadog is a Rollops metric-provider plugin backed by Datadog's
// metrics query API. It resolves a Datadog metrics query to a single scalar —
// the most recent point of the returned series — so rollout analysis can gate a
// canary on Datadog metrics, the way it does on Prometheus.
package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Provider talks to Datadog's metrics query API. BaseURL, APIKey, and AppKey
// come from the plugin's environment (see Config). Window is how far back the
// query reads (default 5m); the latest point is returned.
type Provider struct {
	BaseURL string // e.g. https://api.datadoghq.com
	APIKey  string // DD-API-KEY
	AppKey  string // DD-APPLICATION-KEY
	Window  time.Duration
	HTTP    *http.Client
	now     func() time.Time
}

func (p Provider) client() *http.Client {
	if p.HTTP != nil {
		return p.HTTP
	}
	return http.DefaultClient
}

func (p Provider) clock() time.Time {
	if p.now != nil {
		return p.now()
	}
	return time.Now()
}

func (p Provider) window() time.Duration {
	if p.Window > 0 {
		return p.Window
	}
	return 5 * time.Minute
}

type queryResponse struct {
	Series []struct {
		Pointlist [][]float64 `json:"pointlist"`
	} `json:"series"`
	Error string `json:"error"`
}

// Query runs a Datadog metrics query over the lookback window and returns the
// most recent non-null point of the first series.
func (p Provider) Query(ctx context.Context, query string) (float64, error) {
	if p.APIKey == "" || p.AppKey == "" {
		return 0, fmt.Errorf("datadog: DD_API_KEY and DD_APP_KEY are required")
	}
	to := p.clock().Unix()
	from := p.clock().Add(-p.window()).Unix()
	u := fmt.Sprintf("%s/api/v1/query?from=%d&to=%d&query=%s", p.BaseURL, from, to, url.QueryEscape(query))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("DD-API-KEY", p.APIKey)
	req.Header.Set("DD-APPLICATION-KEY", p.AppKey)
	resp, err := p.client().Do(req)
	if err != nil {
		return 0, fmt.Errorf("datadog: query: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return 0, fmt.Errorf("datadog: status %d", resp.StatusCode)
	}
	var qr queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return 0, fmt.Errorf("datadog: decode: %w", err)
	}
	if qr.Error != "" {
		return 0, fmt.Errorf("datadog: %s", qr.Error)
	}
	if len(qr.Series) == 0 || len(qr.Series[0].Pointlist) == 0 {
		return 0, fmt.Errorf("datadog: query %q returned no data", query)
	}
	// pointlist is [[timestamp, value], …] ordered oldest→newest; take the last
	// point with a value present.
	points := qr.Series[0].Pointlist
	for i := len(points) - 1; i >= 0; i-- {
		if len(points[i]) == 2 {
			return points[i][1], nil
		}
	}
	return 0, fmt.Errorf("datadog: query %q returned no usable point", query)
}
