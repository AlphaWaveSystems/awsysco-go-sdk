package awsysco

import (
	"context"
	"encoding/json"
	"log"
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

// List reads templates from the utmTemplates field of GET /api/v1/me.
//
// Deprecated/known-broken: as of ADR-020 (retracting ADR-003), the platform's
// GET /api/v1/me response does not actually carry a utmTemplates field, and
// there is no dedicated GET /api/user/utm-templates route either — tracked
// upstream as platform issue #831. Until that ships, this always returns an
// empty slice (never silently claims "no templates exist" without warning):
// it logs a one-line warning on every call so the limitation is visible
// rather than silently swallowed. Genuine transport/HTTP errors from the
// underlying /api/v1/me call are still returned as errors.
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
		log.Printf("awsysco: warning: UtmTemplates.List has no working platform endpoint yet (see ADR-020, platform issue #831) — always returning an empty slice, not an authoritative \"no templates\" result")
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
