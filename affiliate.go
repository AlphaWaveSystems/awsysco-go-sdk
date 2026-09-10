package awsysco

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"
)

// AffiliateResource provides access to the affiliate programs API.
type AffiliateResource struct {
	client *Client
}

// AffiliateProgram represents an affiliate program. The same type is used
// for both the owner's full view (CreateProgram/ListPrograms/GetProgram/
// UpdateProgram) and the public discover-subset view (Discover) — Discover
// responses omit MerchantID/MaxPartners/IsPublic/Status/CreatedAt/UpdatedAt,
// which simply decode as zero values; see ADR-024's "discover = public
// summary subset" note.
//
// CookieDays' json tag is cookieDurationDays, not cookieDays — the Go field
// name is kept as-is under ADR-014 (minor releases never break exported
// field names) even though it no longer matches the wire name.
type AffiliateProgram struct {
	ID             string  `json:"id"`
	MerchantID     string  `json:"merchantId,omitempty"`
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	CommissionType string  `json:"commissionType"`
	CpcRate        float64 `json:"cpcRate,omitempty"`
	CpaRate        float64 `json:"cpaRate,omitempty"`
	CookieDays     int     `json:"cookieDurationDays,omitempty"`
	MaxPartners    int     `json:"maxPartners,omitempty"`
	PartnerCount   int     `json:"partnerCount,omitempty"`
	Status         string  `json:"status,omitempty"`
	IsPublic       bool    `json:"isPublic,omitempty"`
	CreatedAt      *string `json:"createdAt,omitempty"`
	UpdatedAt      *string `json:"updatedAt,omitempty"`
}

// UnmarshalJSON normalizes CreatedAt/UpdatedAt, which the platform sends as
// either an ISO-8601 string or a Firestore {_seconds,_nanoseconds} object
// (ADR-017: timestamps stay ISO-8601 strings in 1.x, never raise on an
// unknown shape).
func (p *AffiliateProgram) UnmarshalJSON(b []byte) error {
	type AffiliateProgramAlias AffiliateProgram
	aux := &struct {
		CreatedAt *firestoreTimestamp `json:"createdAt"`
		UpdatedAt *firestoreTimestamp `json:"updatedAt"`
		*AffiliateProgramAlias
	}{
		AffiliateProgramAlias: (*AffiliateProgramAlias)(p),
	}
	if err := json.Unmarshal(b, aux); err != nil {
		return err
	}
	if aux.CreatedAt != nil && !aux.CreatedAt.IsZero() {
		s := aux.CreatedAt.Time.Format(time.RFC3339)
		p.CreatedAt = &s
	}
	if aux.UpdatedAt != nil && !aux.UpdatedAt.IsZero() {
		s := aux.UpdatedAt.Time.Format(time.RFC3339)
		p.UpdatedAt = &s
	}
	return nil
}

// CreateAffiliateProgramInput is the input for creating an affiliate program.
type CreateAffiliateProgramInput struct {
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	CommissionType string  `json:"commissionType,omitempty"`
	CommissionRate float64 `json:"commissionRate,omitempty"`
	CpcRate        float64 `json:"cpcRate,omitempty"`
	CpaRate        float64 `json:"cpaRate,omitempty"`
	CookieDays     int     `json:"cookieDurationDays,omitempty"`
}

// JoinProgramInput is the input for joining an affiliate program.
type JoinProgramInput struct {
	PartnerCode string `json:"partnerCode,omitempty"`
}

// CreateProgram creates a new affiliate program.
func (r *AffiliateResource) CreateProgram(ctx context.Context, input CreateAffiliateProgramInput) (*AffiliateProgram, error) {
	var program AffiliateProgram
	if err := r.client.doRequest(ctx, "POST", pathAffiliatePrograms, input, &program); err != nil {
		return nil, err
	}
	return &program, nil
}

// ListPrograms returns all affiliate programs owned by the authenticated user.
func (r *AffiliateResource) ListPrograms(ctx context.Context) ([]AffiliateProgram, error) {
	var resp struct {
		Programs []AffiliateProgram `json:"programs"`
	}
	if err := r.client.doRequest(ctx, "GET", pathAffiliatePrograms, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Programs, nil
}

// GetProgram retrieves a single affiliate program by ID.
func (r *AffiliateResource) GetProgram(ctx context.Context, programID string) (*AffiliateProgram, error) {
	var program AffiliateProgram
	path := pathAffiliateProgram(programID)
	if err := r.client.doRequest(ctx, "GET", path, nil, &program); err != nil {
		return nil, err
	}
	return &program, nil
}

// UpdateProgram updates an existing affiliate program.
func (r *AffiliateResource) UpdateProgram(ctx context.Context, programID string, input CreateAffiliateProgramInput) (*AffiliateProgram, error) {
	var program AffiliateProgram
	path := pathAffiliateProgram(programID)
	if err := r.client.doRequest(ctx, "PATCH", path, input, &program); err != nil {
		return nil, err
	}
	return &program, nil
}

// GetProgramStats returns analytics for the given affiliate program.
// period examples: "7d", "30d".
func (r *AffiliateResource) GetProgramStats(ctx context.Context, programID, period string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := pathAffiliateProgramStats(programID) + "?period=" + url.QueryEscape(period)
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
	path := pathAffiliateProgramPartners(programID)
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
	path := pathAffiliateProgramPartner(programID, partnerID)
	if err := r.client.doRequest(ctx, "PATCH", path, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Discover returns publicly discoverable affiliate programs.
// limit controls the maximum number returned (0 uses the API default of 20).
func (r *AffiliateResource) Discover(ctx context.Context, limit int) ([]AffiliateProgram, error) {
	path := pathAffiliateDiscover
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
	path := pathAffiliateJoin(programID)
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
	if err := r.client.doRequest(ctx, "GET", pathAffiliatePartnerships, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Partnerships, nil
}

// GetPartnershipStats returns analytics for the given partnership.
func (r *AffiliateResource) GetPartnershipStats(ctx context.Context, partnershipID, period string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := pathAffiliatePartnershipStats(partnershipID) + "?period=" + url.QueryEscape(period)
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// LeaveProgram cancels the authenticated user's partnership with the given program.
func (r *AffiliateResource) LeaveProgram(ctx context.Context, partnershipID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := pathAffiliatePartnership(partnershipID)
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetLimits returns the affiliate feature limits for the authenticated user's tier.
func (r *AffiliateResource) GetLimits(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.client.doRequest(ctx, "GET", pathAffiliateLimits, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
