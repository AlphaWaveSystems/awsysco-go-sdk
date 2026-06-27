package awsysco

import "context"

// UsageResource provides access to the account usage-stats API.
type UsageResource struct {
	client *Client
}

// Get returns usage statistics and tier limits for the authenticated account.
func (r *UsageResource) Get(ctx context.Context) (*UsageStats, error) {
	var stats UsageStats
	if err := r.client.doRequest(ctx, "GET", "/api/user/stats", nil, &stats); err != nil {
		return nil, err
	}
	return &stats, nil
}
