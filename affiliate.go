package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// AffiliateResource provides access to the affiliate programs API.
type AffiliateResource struct {
	client *Client
}

// AffiliateProgram represents an affiliate program.
type AffiliateProgram struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	CommissionType string  `json:"commissionType"`
	CpcRate        float64 `json:"cpcRate,omitempty"`
	CpaRate        float64 `json:"cpaRate,omitempty"`
	CookieDays     int     `json:"cookieDays,omitempty"`
	Status         string  `json:"status,omitempty"`
}

// CreateAffiliateProgramInput is the input for creating an affiliate program.
type CreateAffiliateProgramInput struct {
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	CommissionType string  `json:"commissionType"`
	CpcRate        float64 `json:"cpcRate,omitempty"`
	CpaRate        float64 `json:"cpaRate,omitempty"`
	CookieDays     int     `json:"cookieDays,omitempty"`
}

// JoinProgramInput is the input for joining an affiliate program.
type JoinProgramInput struct {
	PartnerCode string `json:"partnerCode,omitempty"`
}

// CreateProgram creates a new affiliate program.
func (r *AffiliateResource) CreateProgram(ctx context.Context, input CreateAffiliateProgramInput) (*AffiliateProgram, error) {
	var program AffiliateProgram
	if err := r.client.doRequest(ctx, "POST", "/api/affiliate/programs", input, &program); err != nil {
		return nil, err
	}
	return &program, nil
}

// ListPrograms returns all affiliate programs owned by the authenticated user.
func (r *AffiliateResource) ListPrograms(ctx context.Context) ([]AffiliateProgram, error) {
	var resp struct {
		Programs []AffiliateProgram `json:"programs"`
	}
	if err := r.client.doRequest(ctx, "GET", "/api/affiliate/programs", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Programs, nil
}

// GetProgram retrieves a single affiliate program by ID.
func (r *AffiliateResource) GetProgram(ctx context.Context, programID string) (*AffiliateProgram, error) {
	var program AffiliateProgram
	path := fmt.Sprintf("/api/affiliate/programs/%s", url.PathEscape(programID))
	if err := r.client.doRequest(ctx, "GET", path, nil, &program); err != nil {
		return nil, err
	}
	return &program, nil
}

// UpdateProgram updates an existing affiliate program.
func (r *AffiliateResource) UpdateProgram(ctx context.Context, programID string, input CreateAffiliateProgramInput) (*AffiliateProgram, error) {
	var program AffiliateProgram
	path := fmt.Sprintf("/api/affiliate/programs/%s", url.PathEscape(programID))
	if err := r.client.doRequest(ctx, "PATCH", path, input, &program); err != nil {
		return nil, err
	}
	return &program, nil
}

// GetProgramStats returns analytics for the given affiliate program.
// period examples: "7d", "30d".
func (r *AffiliateResource) GetProgramStats(ctx context.Context, programID, period string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/affiliate/programs/%s/stats?period=%s",
		url.PathEscape(programID),
		url.QueryEscape(period),
	)
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListPartners returns the partners enrolled in the given affiliate program.
func (r *AffiliateResource) ListPartners(ctx context.Context, programID string) ([]map[string]interface{}, error) {
	var resp struct {
		Partners []map[string]interface{} `json:"partners"`
	}
	path := fmt.Sprintf("/api/affiliate/programs/%s/partners", url.PathEscape(programID))
	if err := r.client.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Partners, nil
}

// UpdatePartnerStatus approves, rejects, or suspends a partner.
// status values: "approved", "rejected", "suspended".
func (r *AffiliateResource) UpdatePartnerStatus(ctx context.Context, programID, partnerID, status string) (map[string]interface{}, error) {
	body := map[string]string{"status": status}
	var result map[string]interface{}
	path := fmt.Sprintf("/api/affiliate/programs/%s/partners/%s",
		url.PathEscape(programID),
		url.PathEscape(partnerID),
	)
	if err := r.client.doRequest(ctx, "PATCH", path, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Discover returns publicly discoverable affiliate programs.
// limit controls the maximum number returned (0 uses the API default of 20).
func (r *AffiliateResource) Discover(ctx context.Context, limit int) ([]AffiliateProgram, error) {
	path := "/api/affiliate/discover"
	if limit > 0 {
		path += fmt.Sprintf("?limit=%d", limit)
	}
	var resp struct {
		Programs []AffiliateProgram `json:"programs"`
	}
	if err := r.client.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Programs, nil
}

// Join joins a discovered affiliate program.
func (r *AffiliateResource) Join(ctx context.Context, programID string, input JoinProgramInput) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/affiliate/join/%s", url.PathEscape(programID))
	if err := r.client.doRequest(ctx, "POST", path, input, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListPartnerships returns the affiliate programs the authenticated user has joined as a partner.
func (r *AffiliateResource) ListPartnerships(ctx context.Context) ([]map[string]interface{}, error) {
	var resp struct {
		Partnerships []map[string]interface{} `json:"partnerships"`
	}
	if err := r.client.doRequest(ctx, "GET", "/api/affiliate/partnerships", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Partnerships, nil
}

// GetPartnershipStats returns analytics for the given partnership.
func (r *AffiliateResource) GetPartnershipStats(ctx context.Context, partnershipID, period string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/affiliate/partnerships/%s/stats?period=%s",
		url.PathEscape(partnershipID),
		url.QueryEscape(period),
	)
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LeaveProgram cancels the authenticated user's partnership with the given program.
func (r *AffiliateResource) LeaveProgram(ctx context.Context, partnershipID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/affiliate/partnerships/%s", url.PathEscape(partnershipID))
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetLimits returns the affiliate feature limits for the authenticated user's tier.
func (r *AffiliateResource) GetLimits(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.client.doRequest(ctx, "GET", "/api/affiliate/limits", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
