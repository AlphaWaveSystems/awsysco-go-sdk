package awsysco

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// LinksResource provides access to the links API.
type LinksResource struct {
	client *Client
}

// Create creates a new shortened link.
func (r *LinksResource) Create(ctx context.Context, input CreateLinkInput) (*Link, error) {
	var link Link
	if err := r.client.doRequest(ctx, "POST", "/api/v1/links", input, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// List returns a paginated list of links.
func (r *LinksResource) List(ctx context.Context, input ListLinksInput) (*ListLinksResponse, error) {
	q := url.Values{}
	if input.Limit > 0 {
		q.Set("limit", strconv.Itoa(input.Limit))
	}
	if input.Offset > 0 {
		q.Set("offset", strconv.Itoa(input.Offset))
	}

	path := "/api/v1/links"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	var resp ListLinksResponse
	if err := r.client.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Get retrieves a link by its ID.
func (r *LinksResource) Get(ctx context.Context, id string) (*Link, error) {
	var link Link
	if err := r.client.doRequest(ctx, "GET", fmt.Sprintf("/api/v1/links/%s", id), nil, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// Update updates a link's attributes.
func (r *LinksResource) Update(ctx context.Context, id string, input UpdateLinkInput) (*Link, error) {
	var link Link
	if err := r.client.doRequest(ctx, "PATCH", fmt.Sprintf("/api/v1/links/%s", id), input, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// Delete deletes a link by its ID.
func (r *LinksResource) Delete(ctx context.Context, id string) error {
	return r.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/links/%s", id), nil, nil)
}
