package awsysco

import (
	"context"
)

// UtmTemplatesResource provides access to the UTM template API.
type UtmTemplatesResource struct {
	client *Client
}

// UtmTemplate represents a saved UTM parameter template. Wire field names
// are source/medium/campaign/term/content (not utmSource/utmMedium/...) —
// verified live against POST /api/user/utm-templates (functions/app/routes/
// user.js:355); see ADR-020, which retracts the earlier utmSource/utmMedium/
// utmCampaign assumption from ADR-003.
type UtmTemplate struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	Medium   string `json:"medium"`
	Campaign string `json:"campaign"`
	Term     string `json:"term,omitempty"`
	Content  string `json:"content,omitempty"`
}

// CreateUtmTemplateInput is the input for creating a UTM template. See
// UtmTemplate's doc comment for the source/medium/campaign wire naming.
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

// List returns the authenticated user's saved UTM templates via
// GET /api/user/utm-templates.
//
// This route was added by platform PR #833 (ADR-021, which supersedes
// ADR-020/ADR-003's prior "no working list endpoint" guidance): the response
// is {"templates": [...]}. Previously List read a nonexistent utmTemplates
// field off GET /api/v1/me and always returned an empty slice with a warning
// log — that behavior is gone now that the real endpoint exists.
func (r *UtmTemplatesResource) List(ctx context.Context) ([]UtmTemplate, error) {
	var resp struct {
		Templates []UtmTemplate `json:"templates"`
	}
	if err := r.client.doRequest(ctx, "GET", pathUserUtmTemplates, nil, &resp); err != nil {
		return nil, err
	}
	if resp.Templates == nil {
		resp.Templates = []UtmTemplate{}
	}
	return resp.Templates, nil
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
