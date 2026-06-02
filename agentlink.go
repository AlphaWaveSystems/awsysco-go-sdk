package awsysco

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// AgentlinkResource provides access to the AgentLink API.
type AgentlinkResource struct {
	client *Client
}

// Subscribe subscribes an email address to AgentLink updates.
// This endpoint is public and does not require authentication.
func (r *AgentlinkResource) Subscribe(ctx context.Context, email string) (map[string]interface{}, error) {
	body := map[string]string{"email": email}
	var result map[string]interface{}
	if err := r.client.doRequest(ctx, "POST", "/api/agentlink/subscribe", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetLinkStats returns AgentLink analytics for a specific short path.
// periodDays specifies the look-back window in days (e.g. 7, 30).
func (r *AgentlinkResource) GetLinkStats(ctx context.Context, shortPath string, periodDays int) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/agentlink/links/%s/stats?period=%s",
		url.PathEscape(shortPath),
		strconv.Itoa(periodDays),
	)
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetAccountStats returns aggregate AgentLink analytics for the authenticated account.
// periodDays specifies the look-back window in days (e.g. 7, 30).
func (r *AgentlinkResource) GetAccountStats(ctx context.Context, periodDays int) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/agentlink/account/stats?period=%s", strconv.Itoa(periodDays))
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
