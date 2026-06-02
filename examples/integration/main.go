// Command integration is a manual integration test script for the awsysco Go SDK.
//
// Run with a real API key against staging or production:
//
//	AWSYS_API_KEY=awsys_xxx go run examples/integration/main.go
//	AWSYS_API_KEY=awsys_xxx AWSYS_BASE_URL=https://staging.awsys.co go run examples/integration/main.go
//
// Each test prints PASS or FAIL. All operations are best-effort; failures in
// optional features (namespace, affiliates, etc.) do not stop the run.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

var (
	passed int
	failed int
)

func check(name string, err error) bool {
	if err != nil {
		fmt.Printf("  FAIL  %s: %v\n", name, err)
		failed++
		return false
	}
	fmt.Printf("  PASS  %s\n", name)
	passed++
	return true
}

func skip(name, reason string) {
	fmt.Printf("  SKIP  %s: %s\n", name, reason)
}

func main() {
	apiKey := os.Getenv("AWSYS_API_KEY")
	if apiKey == "" {
		log.Fatal("AWSYS_API_KEY environment variable is required")
	}
	baseURL := os.Getenv("AWSYS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://awsys.co"
	}

	client := awsysco.NewClient(apiKey,
		awsysco.WithBaseURL(baseURL),
		awsysco.WithTimeout(20*time.Second),
	)
	ctx := context.Background()

	fmt.Printf("Integration test against: %s\n\n", baseURL)

	// ── Me ────────────────────────────────────────────────────────────────────
	fmt.Println("=== Me ===")
	me, err := client.Me.Get(ctx)
	if check("Me.Get", err) {
		fmt.Printf("        user=%s tier=%s\n", me.Email, me.SubscriptionTier)
	}

	// ── Links ─────────────────────────────────────────────────────────────────
	fmt.Println("\n=== Links ===")
	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL:  "https://example.com/integration-test-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		Tags: []string{"integration-test"},
	})
	if !check("Links.Create", err) {
		log.Fatal("cannot continue without a link")
	}

	linkRef := link.ShortCode
	if linkRef == "" {
		linkRef = link.ID
	}
	fmt.Printf("        created short=%s\n", linkRef)

	_, err = client.Links.List(ctx, awsysco.ListLinksInput{Limit: 5})
	check("Links.List", err)

	_, err = client.Links.Get(ctx, linkRef)
	check("Links.Get", err)

	maxClicks := 500
	_, err = client.Links.Update(ctx, linkRef, awsysco.UpdateLinkInput{MaxClicks: &maxClicks})
	check("Links.Update", err)

	// ── Analytics ─────────────────────────────────────────────────────────────
	fmt.Println("\n=== Analytics ===")
	_, err = client.Analytics.GetStats(ctx, linkRef, "7d")
	check("Analytics.GetStats", err)

	_, err = client.Analytics.GetRecentClicks(ctx, 5)
	check("Analytics.GetRecentClicks", err)

	// ── Tags ──────────────────────────────────────────────────────────────────
	fmt.Println("\n=== Tags ===")
	tagResp, err := client.Tags.Add(ctx, linkRef, "integration")
	if check("Tags.Add", err) {
		_, err = client.Tags.Remove(ctx, linkRef, "integration")
		check("Tags.Remove", err)
		_ = tagResp
	}

	// ── QR ────────────────────────────────────────────────────────────────────
	fmt.Println("\n=== QR ===")
	qrURL := client.QR.GetURL(linkRef, awsysco.WithSize(256))
	if strings.Contains(qrURL, linkRef) {
		check("QR.GetURL", nil)
	} else {
		check("QR.GetURL", fmt.Errorf("URL missing short code"))
	}

	_, err = client.QR.GetSettings(ctx, linkRef)
	if err != nil && awsysco.IsNotFound(err) {
		skip("QR.GetSettings", "endpoint not available for this link")
	} else {
		check("QR.GetSettings", err)
	}

	// ── Folders ───────────────────────────────────────────────────────────────
	fmt.Println("\n=== Folders ===")
	folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
		Name:  "integration-test-folder",
		Color: "#6366F1",
	})
	if check("Folders.Create", err) {
		check("Folders.AssignLink", client.Folders.AssignLink(ctx, linkRef, folder.ID))
		check("Folders.RemoveLink", client.Folders.RemoveLink(ctx, linkRef))
		_, err = client.Folders.Update(ctx, folder.ID, awsysco.UpdateFolderInput{Name: "integration-test-folder-renamed"})
		check("Folders.Update", err)
		check("Folders.Delete", client.Folders.Delete(ctx, folder.ID))
	}

	_, err = client.Folders.List(ctx)
	check("Folders.List", err)

	// ── Bulk ──────────────────────────────────────────────────────────────────
	fmt.Println("\n=== Bulk ===")
	bulkResp, err := client.Bulk.Create(ctx, awsysco.BulkCreateInput{
		URLs: []awsysco.BulkLinkInput{
			{URL: "https://example.com/bulk-1"},
			{URL: "https://example.com/bulk-2"},
		},
	})
	if check("Bulk.Create", err) {
		fmt.Printf("        created=%d failed=%d\n", bulkResp.Created, bulkResp.Failed)
	}

	// ── Webhooks ──────────────────────────────────────────────────────────────
	fmt.Println("\n=== Webhooks ===")
	_, err = client.Webhooks.ListEventTypes(ctx)
	check("Webhooks.ListEventTypes", err)

	webhook, err := client.Webhooks.Create(ctx, awsysco.CreateWebhookInput{
		URL:    "https://example.com/wh-integration",
		Events: []string{"link.created"},
		Name:   "integration-test",
	})
	if check("Webhooks.Create", err) {
		enabled := false
		_, err = client.Webhooks.Update(ctx, webhook.ID, awsysco.UpdateWebhookInput{Enabled: &enabled})
		check("Webhooks.Update", err)

		_, err = client.Webhooks.Test(ctx, webhook.ID, "link.created")
		if err != nil {
			skip("Webhooks.Test", fmt.Sprintf("non-fatal: %v", err))
		} else {
			check("Webhooks.Test", nil)
		}

		_, err = client.Webhooks.Delete(ctx, webhook.ID)
		check("Webhooks.Delete", err)
	}

	// ── Namespace ─────────────────────────────────────────────────────────────
	fmt.Println("\n=== Namespace ===")
	nsInfo, err := client.Namespace.Get(ctx)
	if check("Namespace.Get", err) {
		fmt.Printf("        hasAccess=%v tier=%s\n", nsInfo.HasAccess, nsInfo.Tier)
	}

	_, err = client.Namespace.Check(ctx, "go-sdk-integration")
	if err != nil && (awsysco.IsNotFound(err) || awsysco.IsForbidden(err)) {
		skip("Namespace.Check", fmt.Sprintf("tier gate: %v", err))
	} else {
		check("Namespace.Check", err)
	}

	// ── UTM Templates ──────────────────────────────────────────────────────────
	fmt.Println("\n=== UTM Templates ===")
	utmResp, err := client.UtmTemplates.Create(ctx, awsysco.CreateUtmTemplateInput{
		Name:     "integration-test-utm",
		Source:   "newsletter",
		Medium:   "email",
		Campaign: "go-sdk",
	})
	if check("UtmTemplates.Create", err) {
		_, err = client.UtmTemplates.List(ctx)
		check("UtmTemplates.List", err)

		_, err = client.UtmTemplates.Delete(ctx, utmResp.Template.ID)
		check("UtmTemplates.Delete", err)
	}

	// ── Saved Views ───────────────────────────────────────────────────────────
	fmt.Println("\n=== Saved Views ===")
	viewResp, err := client.SavedViews.Create(ctx, awsysco.CreateSavedViewInput{
		Name:    "integration-test-view",
		Filters: awsysco.SavedViewFilters{Tag: "integration-test"},
	})
	if check("SavedViews.Create", err) {
		_, err = client.SavedViews.List(ctx)
		check("SavedViews.List", err)

		_, err = client.SavedViews.Update(ctx, viewResp.ID, awsysco.UpdateSavedViewInput{Name: "integration-test-view-v2"})
		check("SavedViews.Update", err)

		check("SavedViews.Delete", client.SavedViews.Delete(ctx, viewResp.ID))
	}

	// ── Custom Domains ────────────────────────────────────────────────────────
	fmt.Println("\n=== Custom Domains ===")
	_, err = client.CustomDomains.List(ctx)
	check("CustomDomains.List", err)

	_, err = client.CustomDomains.Check(ctx, "example.com")
	if err != nil && (awsysco.IsNotFound(err) || awsysco.IsForbidden(err)) {
		skip("CustomDomains.Check", fmt.Sprintf("tier gate: %v", err))
	} else {
		check("CustomDomains.Check", err)
	}

	// ── Trust Score ───────────────────────────────────────────────────────────
	fmt.Println("\n=== Trust Score ===")
	_, err = client.TrustScore.Scan(ctx, linkRef)
	if err != nil && (awsysco.IsNotFound(err) || awsysco.IsForbidden(err)) {
		skip("TrustScore.Scan", fmt.Sprintf("not available: %v", err))
	} else {
		check("TrustScore.Scan", err)
	}

	// ── Data Export ───────────────────────────────────────────────────────────
	fmt.Println("\n=== Data Export ===")
	csv, err := client.DataExport.ExportLinks(ctx)
	if check("DataExport.ExportLinks", err) {
		fmt.Printf("        csv bytes=%d\n", len(csv))
	}

	// ── Affiliate ─────────────────────────────────────────────────────────────
	fmt.Println("\n=== Affiliate ===")
	_, err = client.Affiliate.GetLimits(ctx)
	check("Affiliate.GetLimits", err)

	_, err = client.Affiliate.Discover(ctx, 5)
	if err != nil && (awsysco.IsNotFound(err) || awsysco.IsForbidden(err)) {
		skip("Affiliate.Discover", fmt.Sprintf("tier gate: %v", err))
	} else {
		check("Affiliate.Discover", err)
	}

	program, err := client.Affiliate.CreateProgram(ctx, awsysco.CreateAffiliateProgramInput{
		Name:           "integration-test-program",
		CommissionType: "cpc",
		CpcRate:        0.05,
	})
	if err != nil && (awsysco.IsForbidden(err) || awsysco.IsValidationError(err)) {
		skip("Affiliate.CreateProgram", fmt.Sprintf("tier/validation gate: %v", err))
	} else if check("Affiliate.CreateProgram", err) {
		_, err = client.Affiliate.GetProgram(ctx, program.ID)
		check("Affiliate.GetProgram", err)

		_, err = client.Affiliate.ListPartners(ctx, program.ID)
		check("Affiliate.ListPartners", err)
	}

	// ── Agentlink ─────────────────────────────────────────────────────────────
	fmt.Println("\n=== Agentlink ===")
	_, err = client.Agentlink.GetAccountStats(ctx, 7)
	if err != nil && (awsysco.IsNotFound(err) || awsysco.IsForbidden(err)) {
		skip("Agentlink.GetAccountStats", fmt.Sprintf("not available: %v", err))
	} else {
		check("Agentlink.GetAccountStats", err)
	}

	// ── Cleanup ───────────────────────────────────────────────────────────────
	fmt.Println("\n=== Cleanup ===")
	check("Links.Delete", client.Links.Delete(ctx, linkRef))

	// ── Summary ───────────────────────────────────────────────────────────────
	fmt.Printf("\n%s\n", strings.Repeat("─", 40))
	fmt.Printf("Results: %d passed, %d failed\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}
