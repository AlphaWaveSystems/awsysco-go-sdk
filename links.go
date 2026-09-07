package awsysco

import (
	"context"
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
	if err := r.client.doRequest(ctx, "POST", pathLinks, input, &link); err != nil {
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

	path := pathLinks
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
	if err := r.client.doRequest(ctx, "GET", pathLink(id), nil, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// Update updates a link's attributes.
func (r *LinksResource) Update(ctx context.Context, id string, input UpdateLinkInput) (*Link, error) {
	var link Link
	if err := r.client.doRequest(ctx, "PATCH", pathLink(id), input, &link); err != nil {
		return nil, err
	}
	return &link, nil
}

// Delete deletes a link by its ID.
func (r *LinksResource) Delete(ctx context.Context, id string) error {
	return r.client.doRequest(ctx, "DELETE", pathLink(id), nil, nil)
}

// LinkIterator iterates over all links matching a query, transparently
// paging through the platform's offset-based pagination. Obtain one via
// LinksResource.Iter.
type LinkIterator struct {
	client *Client
	ctx    context.Context
	input  ListLinksInput

	buf  []Link
	idx  int
	cur  *Link
	err  error
	done bool
}

// Iter returns an iterator over links matching opts. opts.Limit is clamped
// to the platform maximum of 100 (and defaults to 100 if unset or invalid).
//
//	it := client.Links.Iter(ctx, awsysco.ListLinksInput{})
//	for it.Next() {
//	    link := it.Link()
//	    _ = link
//	}
//	if err := it.Err(); err != nil {
//	    // handle err
//	}
func (r *LinksResource) Iter(ctx context.Context, opts ListLinksInput) *LinkIterator {
	if opts.Limit <= 0 || opts.Limit > 100 {
		opts.Limit = 100
	}
	return &LinkIterator{client: r.client, ctx: ctx, input: opts}
}

// Next advances the iterator and reports whether a link is available via
// Link. It returns false when iteration is exhausted or an error occurred
// (check Err to distinguish the two).
func (it *LinkIterator) Next() bool {
	if it.idx < len(it.buf) {
		it.cur = &it.buf[it.idx]
		it.idx++
		return true
	}
	if it.done {
		return false
	}

	resp, err := it.client.Links.List(it.ctx, it.input)
	if err != nil {
		it.err = err
		it.done = true
		return false
	}

	it.buf = resp.Links
	it.idx = 0
	it.input.Offset += len(resp.Links)

	if !resp.HasMore || len(resp.Links) < it.input.Limit || len(resp.Links) == 0 {
		it.done = true
	}

	if len(it.buf) == 0 {
		return false
	}
	it.cur = &it.buf[0]
	it.idx = 1
	return true
}

// Link returns the link at the iterator's current position. It is only
// valid after a call to Next that returned true.
func (it *LinkIterator) Link() *Link { return it.cur }

// Err returns the first error encountered during iteration, if any.
// Reaching the end of results is not an error.
func (it *LinkIterator) Err() error { return it.err }
