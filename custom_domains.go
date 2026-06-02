package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// CustomDomainsResource provides access to the custom domains API.
type CustomDomainsResource struct {
	client *Client
}

// CustomDomain represents a custom domain registered by the user.
type CustomDomain struct {
	Domain            string  `json:"domain"`
	Status            string  `json:"status"`
	VerificationToken string  `json:"verificationToken,omitempty"`
	IsDefault         bool    `json:"isDefault,omitempty"`
	LinkCount         int     `json:"linkCount,omitempty"`
	CreatedAt         *string `json:"createdAt"`
}

// AddDomainInput is the input for adding a custom domain.
type AddDomainInput struct {
	Domain string `json:"domain"`
}

// UpdateDomainInput is the input for updating a custom domain's settings.
type UpdateDomainInput struct {
	IsDefault    *bool  `json:"isDefault,omitempty"`
	NotFoundHTML string `json:"notFoundHtml,omitempty"`
}

// List returns all custom domains registered by the authenticated user.
func (r *CustomDomainsResource) List(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.client.doRequest(ctx, "GET", "/api/user/domains", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Add registers a new custom domain.
func (r *CustomDomainsResource) Add(ctx context.Context, domain string) (map[string]interface{}, error) {
	body := AddDomainInput{Domain: domain}
	var result map[string]interface{}
	if err := r.client.doRequest(ctx, "POST", "/api/user/domains", body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Verify checks the DNS verification status of a custom domain.
func (r *CustomDomainsResource) Verify(ctx context.Context, domain string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/user/domains/%s/verify", url.PathEscape(domain))
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Activate activates a verified custom domain.
func (r *CustomDomainsResource) Activate(ctx context.Context, domain string) (*CustomDomain, error) {
	var result CustomDomain
	path := fmt.Sprintf("/api/user/domains/%s/activate", url.PathEscape(domain))
	if err := r.client.doRequest(ctx, "POST", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates settings for a custom domain.
func (r *CustomDomainsResource) Update(ctx context.Context, domain string, input UpdateDomainInput) (*CustomDomain, error) {
	var result CustomDomain
	path := fmt.Sprintf("/api/user/domains/%s", url.PathEscape(domain))
	if err := r.client.doRequest(ctx, "PATCH", path, input, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Remove deletes a custom domain registration.
func (r *CustomDomainsResource) Remove(ctx context.Context, domain string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/user/domains/%s", url.PathEscape(domain))
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Check tests whether a hostname is available for use as a custom domain.
func (r *CustomDomainsResource) Check(ctx context.Context, hostname string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/domains/check/%s", url.PathEscape(hostname))
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
