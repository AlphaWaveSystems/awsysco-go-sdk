package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// SavedViewsResource provides access to the saved views API.
type SavedViewsResource struct {
	client *Client
}

// SavedViewFilters defines the filter criteria stored in a saved view.
type SavedViewFilters struct {
	FolderID  string `json:"folderId,omitempty"`
	Tag       string `json:"tag,omitempty"`
	Status    string `json:"status,omitempty"`
	Search    string `json:"search,omitempty"`
	DateRange string `json:"dateRange,omitempty"`
}

// SavedView represents a persisted dashboard filter preset.
type SavedView struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Filters   SavedViewFilters `json:"filters"`
	CreatedAt *string          `json:"createdAt"`
	UpdatedAt *string          `json:"updatedAt"`
}

// CreateSavedViewInput is the input for creating a saved view.
type CreateSavedViewInput struct {
	Name    string           `json:"name"`
	Filters SavedViewFilters `json:"filters"`
}

// UpdateSavedViewInput is the input for updating a saved view.
type UpdateSavedViewInput struct {
	Name    string            `json:"name,omitempty"`
	Filters *SavedViewFilters `json:"filters,omitempty"`
}

// List returns all saved views for the authenticated user.
func (r *SavedViewsResource) List(ctx context.Context) ([]SavedView, error) {
	var resp struct {
		Views []SavedView `json:"views"`
	}
	if err := r.client.doRequest(ctx, "GET", "/api/views", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Views, nil
}

// Create saves a new dashboard view.
func (r *SavedViewsResource) Create(ctx context.Context, input CreateSavedViewInput) (*SavedView, error) {
	var view SavedView
	if err := r.client.doRequest(ctx, "POST", "/api/views", input, &view); err != nil {
		return nil, err
	}
	return &view, nil
}

// Update modifies an existing saved view.
func (r *SavedViewsResource) Update(ctx context.Context, viewID string, input UpdateSavedViewInput) (*SavedView, error) {
	var view SavedView
	path := fmt.Sprintf("/api/views/%s", url.PathEscape(viewID))
	if err := r.client.doRequest(ctx, "PATCH", path, input, &view); err != nil {
		return nil, err
	}
	return &view, nil
}

// Delete removes a saved view by ID.
func (r *SavedViewsResource) Delete(ctx context.Context, viewID string) error {
	path := fmt.Sprintf("/api/views/%s", url.PathEscape(viewID))
	return r.client.doRequest(ctx, "DELETE", path, nil, nil)
}
