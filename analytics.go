package awsysco

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// AnalyticsResource provides access to the analytics API.
type AnalyticsResource struct {
	client *Client
}

// GetStats returns click statistics for the given link short path.
// The period parameter filters results (e.g. "7d", "30d", "all").
// Pass an empty string to use the API default.
func (r *AnalyticsResource) GetStats(ctx context.Context, shortPath string, period string) (*LinkStats, error) {
	path := pathLinkStats(shortPath)
	if period != "" {
		path += "?period=" + url.QueryEscape(period)
	}
	var stats LinkStats
	if err := r.client.doRequest(ctx, "GET", path, nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetAggregateStats returns the aggregated analytics breakdown for a link via
// GET /api/v1/links/{shortPath}/stats/aggregate. opts.Period filters the window
// ("7d", "30d", "90d"); pass nil or an empty period to use the API default.
//
// Free-tier responses populate UpgradeForMore and leave the paid-tier
// breakdowns (DeviceBreakdown, UTMBreakdown, etc.) nil.
func (r *AnalyticsResource) GetAggregateStats(ctx context.Context, shortPath string, opts *AggregateOptions) (*AggregateAnalytics, error) {
	path := pathLinkAggregateStats(shortPath)
	if opts != nil && opts.Period != "" {
		path += "?period=" + url.QueryEscape(opts.Period)
	}
	var stats AggregateAnalytics
	if err := r.client.doRequest(ctx, "GET", path, nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}

// GetRecentClicks returns recent click events across all links for the
// authenticated user via GET /api/user/clicks/recent. limit controls the
// maximum number of events returned (0 uses the API default, currently capped
// at 50 by the platform). An optional since restricts results to click
// events at or after that time — pass no value, or a zero time.Time, to omit
// it. The total number of events returned is available via len() on the
// result.
//
// since is variadic (at most the first value is used) solely to keep this
// method's existing signature backward compatible; only ever pass zero or one
// value.
//
// This endpoint is gated behind the "Live Globe" feature flag on some
// accounts; when disabled it returns a 403 with Code "FEATURE_DISABLED" —
// detect it with IsForbidden(err) and inspect the AwsysError's Code field.
func (r *AnalyticsResource) GetRecentClicks(ctx context.Context, limit int, since ...time.Time) ([]ClickEvent, error) {
	q := url.Values{}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	if len(since) > 0 && !since[0].IsZero() {
		q.Set("since", since[0].UTC().Format(time.RFC3339))
	}
	path := pathRecentClicks
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var resp struct {
		Clicks []ClickEvent `json:"clicks"`
		Count  int          `json:"count"`
	}
	if err := r.client.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Clicks, nil
}
