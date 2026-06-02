package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// TagsResource provides access to the link tags API.
type TagsResource struct {
	client *Client
}

// TagsResponse is the response from tag operations.
type TagsResponse struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
}

// Add adds a tag to the given link.
func (r *TagsResource) Add(ctx context.Context, shortPath, tag string) (*TagsResponse, error) {
	body := map[string]string{"tag": tag}
	var resp TagsResponse
	path := fmt.Sprintf("/api/link/%s/tags", url.PathEscape(shortPath))
	if err := r.client.doRequest(ctx, "POST", path, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Remove removes a tag from the given link.
func (r *TagsResource) Remove(ctx context.Context, shortPath, tag string) (*TagsResponse, error) {
	var resp TagsResponse
	path := fmt.Sprintf("/api/link/%s/tags/%s", url.PathEscape(shortPath), url.PathEscape(tag))
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
