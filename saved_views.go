package awsysco

import (
	"context"
	"encoding/json"
	"time"
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

// UnmarshalJSON tolerates both an ISO-8601 string and a Firestore
// {_seconds, _nanoseconds} object for CreatedAt/UpdatedAt, normalizing either
// shape to an RFC3339 string so the exported field type stays *string.
func (v *SavedView) UnmarshalJSON(b []byte) error {
	type SavedViewAlias SavedView
	aux := &struct {
		CreatedAt *firestoreTimestamp `json:"createdAt"`
		UpdatedAt *firestoreTimestamp `json:"updatedAt"`
		*SavedViewAlias
	}{
		SavedViewAlias: (*SavedViewAlias)(v),
	}
	if err := json.Unmarshal(b, aux); err != nil {
		return err
	}
	if aux.CreatedAt != nil && !aux.CreatedAt.IsZero() {
		s := aux.CreatedAt.Time.Format(time.RFC3339)
		v.CreatedAt = &s
	}
	if aux.UpdatedAt != nil && !aux.UpdatedAt.IsZero() {
		s := aux.UpdatedAt.Time.Format(time.RFC3339)
		v.UpdatedAt = &s
	}
	return nil
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
	if err := r.client.doRequest(ctx, "GET", pathViews, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Views, nil
}

// Create saves a new dashboard view.
func (r *SavedViewsResource) Create(ctx context.Context, input CreateSavedViewInput) (*SavedView, error) {
	var view SavedView
	if err := r.client.doRequest(ctx, "POST", pathViews, input, &view); err != nil {
		return nil, err
	}
	return &view, nil
}

// Update modifies an existing saved view.
func (r *SavedViewsResource) Update(ctx context.Context, viewID string, input UpdateSavedViewInput) (*SavedView, error) {
	var view SavedView
	path := pathView(viewID)
	if err := r.client.doRequest(ctx, "PATCH", path, input, &view); err != nil {
		return nil, err
	}
	return &view, nil
}

// Delete removes a saved view by ID.
func (r *SavedViewsResource) Delete(ctx context.Context, viewID string) error {
	path := pathView(viewID)
	return r.client.doRequest(ctx, "DELETE", path, nil, nil)
}
