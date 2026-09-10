package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestAnalyticsGetStats(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Create a link to get stats for.
	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-stats",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	defer func() {
		if link.ShortCode != "" {
			_ = client.Links.Delete(ctx, link.ShortCode)
		}
	}()

	// Use shortCode (or ID as fallback) for stats lookup.
	id := link.ShortCode
	if id == "" {
		id = link.ID
	}

	stats, err := client.Analytics.GetStats(ctx, id, "7d")
	if err != nil {
		if awsysco.IsNotFound(err) {
			t.Skipf("Analytics.GetStats not available for this link on this environment: %v", err)
		}
		t.Fatalf("Analytics.GetStats failed: %v", err)
	}
	// TotalClicks exists on the struct (may be 0 for a fresh link, that's fine).
	t.Logf("stats for %s: totalClicks=%d", id, stats.TotalClicks)
}

func TestAnalyticsGetRecentClicks(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	clicks, err := client.Analytics.GetRecentClicks(ctx, 5)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsAuthError(err) || awsysco.IsForbidden(err) {
			t.Skipf("Analytics.GetRecentClicks not available on this environment: %v", err)
		}
		t.Fatalf("Analytics.GetRecentClicks failed: %v", err)
	}
	t.Logf("recent clicks: %d events", len(clicks))
}
