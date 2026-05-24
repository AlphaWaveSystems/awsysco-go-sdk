package awsysco

import "context"

// BulkResource provides access to the bulk link creation API.
type BulkResource struct {
	client *Client
}

// Create creates multiple links in a single request.
func (r *BulkResource) Create(ctx context.Context, input BulkCreateInput) (*BulkCreateResponse, error) {
	var resp BulkCreateResponse
	if err := r.client.doRequest(ctx, "POST", "/api/v1/bulk", input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
