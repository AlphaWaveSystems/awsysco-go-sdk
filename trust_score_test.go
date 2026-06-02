package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestTrustScoreScan(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-trust-score",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	ref := link.ShortCode
	if ref == "" {
		ref = link.ID
	}
	defer func() { _ = client.Links.Delete(ctx, ref) }()

	result, err := client.TrustScore.Scan(ctx, ref)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) {
			t.Skipf("TrustScore.Scan not available on this environment: %v", err)
		}
		t.Fatalf("TrustScore.Scan failed: %v", err)
	}
	t.Logf("trust score result: short=%s status=%v", result.Short, result.Status)
}
