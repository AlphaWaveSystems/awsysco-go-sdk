package awsysco_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestUsageGet(t *testing.T) {
	const body = `{
		"totalLinks": 42,
		"totalClicks": 1337,
		"linksCreatedThisMonth": 5,
		"qrCodesThisMonth": 3,
		"folderCount": 2,
		"apiCallsThisMonth": 88,
		"trackedClicksThisMonth": 900,
		"tier": "pro",
		"limits": {
			"linksPerMonth": "unlimited",
			"monthlyLinks": "unlimited",
			"dailyLinks": 100,
			"monthlyTrackedClicks": "unlimited",
			"apiCallsPerMonth": 1000,
			"qrCodes": 50,
			"folders": "unlimited",
			"customSlugs": 25
		},
		"hasApiKey": true,
		"apiKeyCreatedAt": "2026-06-01T00:00:00Z",
		"userPrefix": "px",
		"isPremium": true,
		"overage": {
			"active": true,
			"startedAt": "2026-06-20T00:00:00Z",
			"expiresAt": "2026-06-27T00:00:00Z",
			"hoursUntilDrop": 12.5,
			"clicksThisCycle": 2500,
			"spendingLimitCents": 5000,
			"estimatedChargeCents": 45
		}
	}`

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	stats, err := client.Usage.Get(context.Background())
	if err != nil {
		t.Fatalf("Usage.Get failed: %v", err)
	}

	if gotPath != "/api/user/stats" {
		t.Fatalf("expected path /api/user/stats, got %s", gotPath)
	}

	if stats.TotalLinks != 42 {
		t.Errorf("TotalLinks = %d, want 42", stats.TotalLinks)
	}
	if stats.TotalClicks != 1337 {
		t.Errorf("TotalClicks = %d, want 1337", stats.TotalClicks)
	}
	if stats.Tier != "pro" {
		t.Errorf("Tier = %q, want pro", stats.Tier)
	}
	if !stats.HasAPIKey {
		t.Error("HasAPIKey = false, want true")
	}
	if stats.APIKeyCreatedAt == nil || *stats.APIKeyCreatedAt != "2026-06-01T00:00:00Z" {
		t.Errorf("APIKeyCreatedAt = %v, want 2026-06-01T00:00:00Z", stats.APIKeyCreatedAt)
	}
	if stats.UserPrefix == nil || *stats.UserPrefix != "px" {
		t.Errorf("UserPrefix = %v, want px", stats.UserPrefix)
	}
	if !stats.IsPremium {
		t.Error("IsPremium = false, want true")
	}

	// Flexible int-or-unlimited fields.
	if !stats.Limits.LinksPerMonth.Unlimited {
		t.Error("Limits.LinksPerMonth should be unlimited")
	}
	if !stats.Limits.MonthlyTrackedClicks.Unlimited {
		t.Error("Limits.MonthlyTrackedClicks should be unlimited")
	}
	if !stats.Limits.Folders.Unlimited {
		t.Error("Limits.Folders should be unlimited")
	}
	if stats.Limits.DailyLinks.Unlimited || stats.Limits.DailyLinks.Value != 100 {
		t.Errorf("Limits.DailyLinks = %+v, want {Value:100 Unlimited:false}", stats.Limits.DailyLinks)
	}
	if stats.Limits.QRCodes.Unlimited || stats.Limits.QRCodes.Value != 50 {
		t.Errorf("Limits.QRCodes = %+v, want {Value:50 Unlimited:false}", stats.Limits.QRCodes)
	}

	// Plain int limits.
	if stats.Limits.APICallsPerMonth != 1000 {
		t.Errorf("Limits.APICallsPerMonth = %d, want 1000", stats.Limits.APICallsPerMonth)
	}
	if stats.Limits.CustomSlugs != 25 {
		t.Errorf("Limits.CustomSlugs = %d, want 25", stats.Limits.CustomSlugs)
	}

	// Overage.
	if !stats.Overage.Active {
		t.Error("Overage.Active = false, want true")
	}
	if stats.Overage.HoursUntilDrop == nil || *stats.Overage.HoursUntilDrop != 12.5 {
		t.Errorf("Overage.HoursUntilDrop = %v, want 12.5", stats.Overage.HoursUntilDrop)
	}
	if stats.Overage.ClicksThisCycle != 2500 {
		t.Errorf("Overage.ClicksThisCycle = %d, want 2500", stats.Overage.ClicksThisCycle)
	}
	if stats.Overage.EstimatedChargeCents != 45 {
		t.Errorf("Overage.EstimatedChargeCents = %d, want 45", stats.Overage.EstimatedChargeCents)
	}
}
