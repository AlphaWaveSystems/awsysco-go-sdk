package awsysco_test

import (
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
