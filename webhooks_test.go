package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestWebhooksListEventTypes(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	result, err := client.Webhooks.ListEventTypes(ctx)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Webhooks.ListEventTypes not available on this environment: %v", err)
		}
		t.Fatalf("Webhooks.ListEventTypes failed: %v", err)
	}
	t.Logf("event types: %v", result)
}

func TestWebhooksCreateUpdateDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	webhook, err := client.Webhooks.Create(ctx, awsysco.CreateWebhookInput{
		URL:    "https://example.com/go-sdk-test-webhook",
		Events: []string{"link.created"},
		Name:   "go-sdk-test",
	})
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Webhooks.Create not available on this environment: %v", err)
		}
		t.Fatalf("Webhooks.Create failed: %v", err)
	}
	t.Logf("created webhook: id=%s url=%s", webhook.ID, webhook.URL)

	defer func() {
		_, _ = client.Webhooks.Delete(ctx, webhook.ID)
	}()

	if webhook.ID == "" {
		t.Fatal("expected non-empty webhook ID")
	}

	enabled := false
	updated, err := client.Webhooks.Update(ctx, webhook.ID, awsysco.UpdateWebhookInput{
		Enabled: &enabled,
	})
	if err != nil {
		t.Fatalf("Webhooks.Update failed: %v", err)
	}
	t.Logf("updated webhook: id=%s enabled=%v", updated.ID, updated.Enabled)

	_, err = client.Webhooks.Delete(ctx, webhook.ID)
	if err != nil {
		t.Fatalf("Webhooks.Delete failed: %v", err)
	}
	t.Logf("deleted webhook %s", webhook.ID)
}

func TestWebhooksList(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	result, err := client.Webhooks.List(ctx)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Webhooks.List not available on this environment: %v", err)
		}
		t.Fatalf("Webhooks.List failed: %v", err)
	}
	t.Logf("webhooks list: %v", result)
}
