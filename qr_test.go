package awsysco_test

import (
	"context"
	"strings"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestQRGetURL(t *testing.T) {
	client := newTestClient(t)

	shortCode := "abc123"
	u := client.QR.GetURL(shortCode)

	if !strings.Contains(u, "/api/qr/"+shortCode) {
		t.Errorf("expected URL to contain /api/qr/%s, got: %s", shortCode, u)
	}
	t.Logf("QR URL: %s", u)
}

func TestQRGetURLOptions(t *testing.T) {
	client := newTestClient(t)

	shortCode := "xyz789"
	u := client.QR.GetURL(shortCode, awsysco.WithSize(512))

	if !strings.Contains(u, "size=512") {
		t.Errorf("expected URL to contain size=512, got: %s", u)
	}
	if !strings.Contains(u, "/api/qr/"+shortCode) {
		t.Errorf("expected URL to contain /api/qr/%s, got: %s", shortCode, u)
	}
	t.Logf("QR URL with options: %s", u)
}

func TestQRGetSettings(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-qr-settings",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	ref := link.ShortCode
	if ref == "" {
		ref = link.ID
	}
	defer func() { _ = client.Links.Delete(ctx, ref) }()

	settings, err := client.QR.GetSettings(ctx, ref)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) {
			t.Skipf("QR.GetSettings not available on this environment: %v", err)
		}
		t.Fatalf("QR.GetSettings failed: %v", err)
	}
	t.Logf("QR settings: size=%d color=%s", settings.Size, settings.Color)
}
