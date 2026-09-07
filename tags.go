package awsysco

import (
	"context"
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

// Add adds one or more tags to the given link. The platform's endpoint takes
// a batch ({"tags": [...]}) rather than one tag per call; Add is variadic so
// existing single-tag call sites (Add(ctx, shortPath, "a")) keep compiling.
func (r *TagsResource) Add(ctx context.Context, shortPath string, tags ...string) (*TagsResponse, error) {
	body := map[string][]string{"tags": tags}
	var resp TagsResponse
	path := pathTags(shortPath)
	if err := r.client.doRequest(ctx, "POST", path, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Remove removes a tag from the given link.
func (r *TagsResource) Remove(ctx context.Context, shortPath, tag string) (*TagsResponse, error) {
	var resp TagsResponse
	path := pathTag(shortPath, tag)
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
