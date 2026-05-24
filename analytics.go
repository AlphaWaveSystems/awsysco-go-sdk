package awsysco

import (
	"context"
	"fmt"
)

// AnalyticsResource provides access to the analytics API.
type AnalyticsResource struct {
	client *Client
}

// GetStats returns click statistics for the given link ID.
func (r *AnalyticsResource) GetStats(ctx context.Context, id string) (*LinkStats, error) {
	var stats LinkStats
	if err := r.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/links/%s/stats", id), nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}
