package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// UtmTemplatesResource provides access to the UTM template API.
type UtmTemplatesResource struct {
	client *Client
}

// UtmTemplate represents a saved UTM parameter template.
type UtmTemplate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	Medium   string `json:"medium"`
	Campaign string `json:"campaign"`
	Term     string `json:"term,omitempty"`
	Content  string `json:"content,omitempty"`
}

// CreateUtmTemplateInput is the input for creating a UTM template.
type CreateUtmTemplateInput struct {
	Name     string `json:"name"`
	Source   string `json:"source"`
	Medium   string `json:"medium"`
	Campaign string `json:"campaign"`
	Term     string `json:"term,omitempty"`
	Content  string `json:"content,omitempty"`
}

// CreateUtmTemplateResponse is the response from creating a UTM template.
type CreateUtmTemplateResponse struct {
	Success  bool        `json:"success"`
	Template UtmTemplate `json:"template"`
}

// List returns all UTM templates for the authenticated user.
func (r *UtmTemplatesResource) List(ctx context.Context) ([]UtmTemplate, error) {
	var resp struct {
		UtmTemplates []UtmTemplate `json:"utmTemplates"`
	}
	if err := r.client.doRequest(ctx, "GET", "/api/v1/me", nil, &resp); err != nil {
		return nil, err
	}
	return resp.UtmTemplates, nil
}

// Create saves a new UTM template.
func (r *UtmTemplatesResource) Create(ctx context.Context, input CreateUtmTemplateInput) (*CreateUtmTemplateResponse, error) {
	var resp CreateUtmTemplateResponse
	if err := r.client.doRequest(ctx, "POST", "/api/user/utm-templates", input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete removes a UTM template by ID.
func (r *UtmTemplatesResource) Delete(ctx context.Context, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/user/utm-templates/%s", url.PathEscape(id))
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
