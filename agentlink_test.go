package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestAgentlinkGetAccountStats(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	result, err := client.Agentlink.GetAccountStats(ctx, 7)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Agentlink.GetAccountStats not available on this environment: %v", err)
		}
		t.Fatalf("Agentlink.GetAccountStats failed: %v", err)
	}
	t.Logf("agentlink account stats (7d): %v", result)
}

func TestAgentlinkGetLinkStats(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-agentlink-stats",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	ref := link.ShortCode
	if ref == "" {
		ref = link.ID
	}
	defer func() { _ = client.Links.Delete(ctx, ref) }()

	result, err := client.Agentlink.GetLinkStats(ctx, ref, 7)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) {
			t.Skipf("Agentlink.GetLinkStats not available on this environment: %v", err)
		}
		t.Fatalf("Agentlink.GetLinkStats failed: %v", err)
	}
	t.Logf("agentlink link stats (7d): %v", result)
}
