package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// NamespaceResource provides access to the branded namespace API.
type NamespaceResource struct {
	client *Client
}

// NamespaceInfo describes the authenticated user's namespace status.
type NamespaceInfo struct {
	HasAccess       bool    `json:"hasAccess"`
	Namespace       *string `json:"namespace"`
	Tier            string  `json:"tier"`
	UpgradeRequired bool    `json:"upgradeRequired"`
}

// NamespaceCheckResult is the result of checking namespace availability.
type NamespaceCheckResult struct {
	Namespace  string  `json:"namespace"`
	Available  bool    `json:"available"`
	Reason     *string `json:"reason"`
	PreviewURL *string `json:"previewUrl"`
}

// Get returns the authenticated user's current namespace info.
func (r *NamespaceResource) Get(ctx context.Context) (*NamespaceInfo, error) {
	var info NamespaceInfo
	if err := r.client.doRequest(ctx, "GET", "/api/user/namespace", nil, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Check tests whether the given namespace is available to claim.
func (r *NamespaceResource) Check(ctx context.Context, namespace string) (*NamespaceCheckResult, error) {
	var result NamespaceCheckResult
	path := fmt.Sprintf("/api/namespace/check/%s", url.PathEscape(namespace))
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Claim claims the given namespace for the authenticated user.
func (r *NamespaceResource) Claim(ctx context.Context, namespace string) (*NamespaceInfo, error) {
	body := map[string]string{"namespace": namespace}
	var info NamespaceInfo
	if err := r.client.doRequest(ctx, "POST", "/api/user/namespace", body, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// Release releases the authenticated user's current namespace.
func (r *NamespaceResource) Release(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.client.doRequest(ctx, "DELETE", "/api/user/namespace", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}
