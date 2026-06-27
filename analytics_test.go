package awsysco_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func TestAnalyticsGetAggregateStatsFreeTier(t *testing.T) {
	const body = `{
		"shortCode": "abc123",
		"fullPath": "promo/abc123",
		"period": "30d",
		"totalClicks": 120,
		"uniqueVisitors": 95,
		"clicksByDay": [
			{"date": "2026-06-01", "clicks": 10},
			{"date": "2026-06-02", "clicks": 14}
		],
		"countryBreakdown": {"US": 80, "CA": 40},
		"tierLimit": 30,
		"tier": "free",
		"upgradeForMore": {
			"available": ["deviceBreakdown", "utmBreakdown", "hourBreakdown"],
			"message": "Upgrade to Pro for device, UTM and hourly breakdowns."
		}
	}`

	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	stats, err := client.Analytics.GetAggregateStats(context.Background(), "abc123", &awsysco.AggregateOptions{Period: "30d"})
	if err != nil {
		t.Fatalf("GetAggregateStats failed: %v", err)
	}

	if gotPath != "/api/v1/links/abc123/stats/aggregate" {
		t.Errorf("path = %s, want /api/v1/links/abc123/stats/aggregate", gotPath)
	}
	if gotQuery != "period=30d" {
		t.Errorf("query = %q, want period=30d", gotQuery)
	}
	if stats.TotalClicks != 120 {
		t.Errorf("TotalClicks = %d, want 120", stats.TotalClicks)
	}
	if stats.UniqueVisitors != 95 {
		t.Errorf("UniqueVisitors = %d, want 95", stats.UniqueVisitors)
	}
	if stats.Tier != "free" {
		t.Errorf("Tier = %q, want free", stats.Tier)
	}
	if len(stats.ClicksByDay) != 2 || stats.ClicksByDay[0].Clicks != 10 {
		t.Errorf("ClicksByDay = %+v, unexpected", stats.ClicksByDay)
	}
	if stats.CountryBreakdown["US"] != 80 {
		t.Errorf("CountryBreakdown[US] = %d, want 80", stats.CountryBreakdown["US"])
	}
	// Paid-tier fields must be nil on free tier.
	if stats.DeviceBreakdown != nil {
		t.Error("DeviceBreakdown should be nil on free tier")
	}
	if stats.UTMBreakdown != nil {
		t.Error("UTMBreakdown should be nil on free tier")
	}
	// UpgradeForMore must be present.
	if stats.UpgradeForMore == nil {
		t.Fatal("UpgradeForMore should be present on free tier")
	}
	if len(stats.UpgradeForMore.Available) != 3 {
		t.Errorf("UpgradeForMore.Available = %v, want 3 entries", stats.UpgradeForMore.Available)
	}
}

func TestAnalyticsGetAggregateStatsProTier(t *testing.T) {
	const body = `{
		"shortCode": "abc123",
		"period": "7d",
		"totalClicks": 500,
		"uniqueVisitors": 410,
		"clicksByDay": [{"date": "2026-06-20", "clicks": 70}],
		"countryBreakdown": {"US": 300},
		"tierLimit": 90,
		"tier": "pro",
		"deviceBreakdown": {"mobile": 250, "desktop": 200, "tablet": 50},
		"referrerBreakdown": {"twitter.com": 120, "direct": 380},
		"browserBreakdown": {"Chrome": 400, "Safari": 100},
		"osBreakdown": {"iOS": 220, "Windows": 180},
		"sourceBreakdown": {"newsletter": 90},
		"hourBreakdown": [{"hour": 9, "clicks": 40}, {"hour": 10, "clicks": 55}],
		"utmBreakdown": {
			"sources": {"newsletter": 90},
			"mediums": {"email": 90},
			"campaigns": {"launch": 90}
		}
	}`

	var gotPath, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	stats, err := client.Analytics.GetAggregateStats(context.Background(), "abc123", &awsysco.AggregateOptions{Period: "7d"})
	if err != nil {
		t.Fatalf("GetAggregateStats failed: %v", err)
	}

	if gotPath != "/api/v1/links/abc123/stats/aggregate" {
		t.Errorf("path = %s, unexpected", gotPath)
	}
	if gotQuery != "period=7d" {
		t.Errorf("query = %q, want period=7d", gotQuery)
	}
	if stats.Tier != "pro" {
		t.Errorf("Tier = %q, want pro", stats.Tier)
	}
	if stats.DeviceBreakdown == nil {
		t.Fatal("DeviceBreakdown should be present on pro tier")
	}
	if stats.DeviceBreakdown.Mobile != 250 || stats.DeviceBreakdown.Desktop != 200 || stats.DeviceBreakdown.Tablet != 50 {
		t.Errorf("DeviceBreakdown = %+v, unexpected", stats.DeviceBreakdown)
	}
	if stats.ReferrerBreakdown["direct"] != 380 {
		t.Errorf("ReferrerBreakdown[direct] = %d, want 380", stats.ReferrerBreakdown["direct"])
	}
	if stats.BrowserBreakdown["Chrome"] != 400 {
		t.Errorf("BrowserBreakdown[Chrome] = %d, want 400", stats.BrowserBreakdown["Chrome"])
	}
	if stats.OSBreakdown["iOS"] != 220 {
		t.Errorf("OSBreakdown[iOS] = %d, want 220", stats.OSBreakdown["iOS"])
	}
	if len(stats.HourBreakdown) != 2 || stats.HourBreakdown[1].Hour != 10 {
		t.Errorf("HourBreakdown = %+v, unexpected", stats.HourBreakdown)
	}
	if stats.UTMBreakdown == nil {
		t.Fatal("UTMBreakdown should be present on pro tier")
	}
	if stats.UTMBreakdown.Sources["newsletter"] != 90 {
		t.Errorf("UTMBreakdown.Sources[newsletter] = %d, want 90", stats.UTMBreakdown.Sources["newsletter"])
	}
	if stats.UpgradeForMore != nil {
		t.Error("UpgradeForMore should be nil on pro tier")
	}
	if stats.FullPath != nil {
		t.Error("FullPath should be nil when absent in response")
	}
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
