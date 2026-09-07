package awsysco

import (
	"context"
	"encoding/json"
)

// UtmTemplatesResource provides access to the UTM template API.
type UtmTemplatesResource struct {
	client *Client
}

// UtmTemplate represents a saved UTM parameter template.
type UtmTemplate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Source   string `json:"utmSource"`
	Medium   string `json:"utmMedium"`
	Campaign string `json:"utmCampaign"`
	Term     string `json:"utmTerm,omitempty"`
	Content  string `json:"utmContent,omitempty"`
}

// CreateUtmTemplateInput is the input for creating a UTM template.
type CreateUtmTemplateInput struct {
	Name     string `json:"name"`
	Source   string `json:"utmSource"`
	Medium   string `json:"utmMedium"`
	Campaign string `json:"utmCampaign"`
	Term     string `json:"utmTerm,omitempty"`
	Content  string `json:"utmContent,omitempty"`
}

// CreateUtmTemplateResponse is the response from creating a UTM template.
type CreateUtmTemplateResponse struct {
	Success  bool        `json:"success"`
	Template UtmTemplate `json:"template"`
}

// List reads templates from the utmTemplates field of GET /api/v1/me — no
// dedicated list endpoint exists on the platform (see ADR-003). If the field
// is missing, null, or an unexpected shape, List returns an empty slice
// rather than an error; genuine transport/HTTP errors are still propagated.
func (r *UtmTemplatesResource) List(ctx context.Context) ([]UtmTemplate, error) {
	var resp struct {
		UtmTemplates json.RawMessage `json:"utmTemplates"`
	}
	if err := r.client.doRequest(ctx, "GET", pathMe, nil, &resp); err != nil {
		return nil, err
	}
	var templates []UtmTemplate
	if len(resp.UtmTemplates) > 0 {
		_ = json.Unmarshal(resp.UtmTemplates, &templates)
	}
	if templates == nil {
		templates = []UtmTemplate{}
	}
	return templates, nil
}

// Create saves a new UTM template.
func (r *UtmTemplatesResource) Create(ctx context.Context, input CreateUtmTemplateInput) (*CreateUtmTemplateResponse, error) {
	var resp CreateUtmTemplateResponse
	if err := r.client.doRequest(ctx, "POST", pathUserUtmTemplates, input, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Delete removes a UTM template by ID.
func (r *UtmTemplatesResource) Delete(ctx context.Context, id string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := pathUtmTemplate(id)
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
