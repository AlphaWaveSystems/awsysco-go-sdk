package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// FoldersResource provides access to the folders API.
type FoldersResource struct {
	client *Client
}

// Create creates a new folder.
func (r *FoldersResource) Create(ctx context.Context, input CreateFolderInput) (*Folder, error) {
	var folder Folder
	if err := r.client.doRequest(ctx, "POST", "/api/v1/folders", input, &folder); err != nil {
		return nil, err
	}
	return &folder, nil
}

// List returns all folders for the authenticated user.
func (r *FoldersResource) List(ctx context.Context) (*ListFoldersResponse, error) {
	var resp ListFoldersResponse
	if err := r.client.doRequest(ctx, "GET", "/api/v1/folders", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AssignLink assigns a link to a folder.
// linkID may be the link's Firestore document ID or its shortCode.
func (r *FoldersResource) AssignLink(ctx context.Context, linkID, folderID string) error {
	body := map[string]string{"folderId": folderID}
	return r.client.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/links/%s/folder", linkID), body, nil)
}

// RemoveLink removes a link from its folder by passing a null folderId.
// linkID may be the link's Firestore document ID or its shortCode.
func (r *FoldersResource) RemoveLink(ctx context.Context, linkID string) error {
	body := map[string]interface{}{"folderId": nil}
	return r.client.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/links/%s/folder", linkID), body, nil)
}

// Update updates a folder's name or color.
func (r *FoldersResource) Update(ctx context.Context, folderID string, input UpdateFolderInput) (*Folder, error) {
	var folder Folder
	if err := r.client.doRequest(ctx, "PATCH", fmt.Sprintf("/api/v1/folders/%s", url.PathEscape(folderID)), input, &folder); err != nil {
		return nil, err
	}
	return &folder, nil
}

// Delete deletes a folder by its ID.
func (r *FoldersResource) Delete(ctx context.Context, id string) error {
	return r.client.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/folders/%s", id), nil, nil)
}
