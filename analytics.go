package awsysco

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// AnalyticsResource provides access to the analytics API.
type AnalyticsResource struct {
	client *Client
}

// GetStats returns click statistics for the given link short path.
// The period parameter filters results (e.g. "7d", "30d", "all").
// Pass an empty string to use the API default.
func (r *AnalyticsResource) GetStats(ctx context.Context, shortPath string, period string) (*LinkStats, error) {
	path := fmt.Sprintf("/api/v1/links/%s/stats", url.PathEscape(shortPath))
	if period != "" {
		path += "?period=" + url.QueryEscape(period)
	}
	var stats LinkStats
	if err := r.client.doRequest(ctx, "GET", path, nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetRecentClicks returns recent click events across all links for the authenticated user.
// limit controls the maximum number of events returned (0 uses the API default).
func (r *AnalyticsResource) GetRecentClicks(ctx context.Context, limit int) ([]ClickEvent, error) {
	path := "/api/user/recent-clicks"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	var resp struct {
		Clicks []ClickEvent `json:"clicks"`
	}
	if err := r.client.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Clicks, nil
}
