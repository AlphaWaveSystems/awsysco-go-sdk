package awsysco

import "context"

// MeResource provides access to the current user info API.
type MeResource struct {
	client *Client
}

// Get returns information about the currently authenticated user.
func (r *MeResource) Get(ctx context.Context) (*MeResponse, error) {
	var me MeResponse
	if err := r.client.doRequest(ctx, "GET", "/api/v1/me", nil, &me); err != nil {
		return nil, err
	}
	return &me, nil
}
