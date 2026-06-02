package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// WebhooksResource provides access to the webhooks API.
type WebhooksResource struct {
	client *Client
}

// Webhook represents a registered webhook endpoint.
type Webhook struct {
	ID            string   `json:"id"`
	URL           string   `json:"url"`
	Events        []string `json:"events"`
	Name          string   `json:"name,omitempty"`
	Enabled       bool     `json:"enabled"`
	CreatedAt     *string  `json:"createdAt"`
	UpdatedAt     *string  `json:"updatedAt"`
	LastTriggered *string  `json:"lastTriggered"`
	FailureCount  int      `json:"failureCount,omitempty"`
}

// CreateWebhookInput is the input for registering a new webhook.
type CreateWebhookInput struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Name   string   `json:"name,omitempty"`
	Secret string   `json:"secret,omitempty"`
}

// UpdateWebhookInput is the input for updating an existing webhook.
type UpdateWebhookInput struct {
	URL     string   `json:"url,omitempty"`
	Events  []string `json:"events,omitempty"`
	Name    string   `json:"name,omitempty"`
	Secret  string   `json:"secret,omitempty"`
	Enabled *bool    `json:"enabled,omitempty"`
}

// ListEventTypes returns the available webhook event types.
func (r *WebhooksResource) ListEventTypes(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.client.doRequest(ctx, "GET", "/api/webhooks/event-types", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// List returns all webhooks registered for the authenticated user.
func (r *WebhooksResource) List(ctx context.Context) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := r.client.doRequest(ctx, "GET", "/api/webhooks", nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Create registers a new webhook.
func (r *WebhooksResource) Create(ctx context.Context, input CreateWebhookInput) (*Webhook, error) {
	var webhook Webhook
	if err := r.client.doRequest(ctx, "POST", "/api/webhooks", input, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Update modifies an existing webhook.
func (r *WebhooksResource) Update(ctx context.Context, webhookID string, input UpdateWebhookInput) (*Webhook, error) {
	var webhook Webhook
	path := fmt.Sprintf("/api/webhooks/%s", url.PathEscape(webhookID))
	if err := r.client.doRequest(ctx, "PATCH", path, input, &webhook); err != nil {
		return nil, err
	}
	return &webhook, nil
}

// Delete removes a webhook by ID.
func (r *WebhooksResource) Delete(ctx context.Context, webhookID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	path := fmt.Sprintf("/api/webhooks/%s", url.PathEscape(webhookID))
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Test sends a test event to the given webhook.
func (r *WebhooksResource) Test(ctx context.Context, webhookID, eventType string) (map[string]interface{}, error) {
	body := map[string]string{"eventType": eventType}
	var result map[string]interface{}
	path := fmt.Sprintf("/api/webhooks/%s/test", url.PathEscape(webhookID))
	if err := r.client.doRequest(ctx, "POST", path, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}
