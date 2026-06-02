// Command basic demonstrates basic usage of the awsysco Go SDK.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func main() {
	apiKey := os.Getenv("AWSYS_API_KEY")
	if apiKey == "" {
		log.Fatal("AWSYS_API_KEY environment variable is required")
	}

	// Create a client (defaults to https://awsys.co)
	client := awsysco.NewClient(apiKey,
		awsysco.WithTimeout(15*time.Second),
	)

	ctx := context.Background()

	// ── Me ──────────────────────────────────────────────────────────────────
	me, err := client.Me.Get(ctx)
	if err != nil {
		log.Fatalf("failed to get user info: %v", err)
	}
	fmt.Printf("Logged in as: %s (tier: %s)\n", me.Email, me.SubscriptionTier)

	// ── Links ────────────────────────────────────────────────────────────────
	maxClicks := 1000
	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL:        "https://example.com/very/long/url/that/needs/shortening",
		MaxClicks:  &maxClicks,
		Tags:       []string{"demo", "go-sdk"},
		OgMeta:     &awsysco.OgMeta{Title: "Example", Description: "SDK demo link"},
	})
	if err != nil {
		log.Fatalf("failed to create link: %v", err)
	}
	fmt.Printf("Created link: %s -> %s\n", link.ShortURL, link.Long)

	list, err := client.Links.List(ctx, awsysco.ListLinksInput{Limit: 5})
	if err != nil {
		log.Fatalf("failed to list links: %v", err)
	}
	fmt.Printf("\nRecent links (%d total):\n", list.Total)
	for _, l := range list.Links {
		fmt.Printf("  %s -> %s (%d clicks)\n", l.ShortURL, l.Long, l.Clicks)
	}

	// ── Analytics ────────────────────────────────────────────────────────────
	stats, err := client.Analytics.GetStats(ctx, link.ShortCode, "7d")
	if err != nil {
		log.Printf("warning: failed to get stats: %v", err)
	} else {
		fmt.Printf("\nStats for %s (7d): %d total clicks\n", link.ShortCode, stats.TotalClicks)
	}

	recentClicks, err := client.Analytics.GetRecentClicks(ctx, 10)
	if err != nil {
		log.Printf("warning: failed to get recent clicks: %v", err)
	} else {
		fmt.Printf("Recent click events: %d\n", len(recentClicks))
	}

	// ── QR ───────────────────────────────────────────────────────────────────
	qrURL := client.QR.GetURL(link.ShortCode, awsysco.WithSize(400))
	fmt.Printf("QR Code URL: %s\n", qrURL)

	// ── Tags ─────────────────────────────────────────────────────────────────
	tagResp, err := client.Tags.Add(ctx, link.ShortCode, "featured")
	if err != nil {
		log.Printf("warning: Tags.Add: %v", err)
	} else {
		fmt.Printf("Tags after add: %v\n", tagResp.Tags)
	}

	// ── Folders ───────────────────────────────────────────────────────────────
	folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
		Name:  "Example Folder",
		Color: "#10B981",
	})
	if err != nil {
		log.Fatalf("failed to create folder: %v", err)
	}
	fmt.Printf("\nCreated folder: %s (id: %s)\n", folder.Name, folder.ID)

	if err := client.Folders.AssignLink(ctx, link.ShortCode, folder.ID); err != nil {
		log.Printf("warning: AssignLink: %v", err)
	}

	_, err = client.Folders.Update(ctx, folder.ID, awsysco.UpdateFolderInput{Name: "Example Folder (renamed)"})
	if err != nil {
		log.Printf("warning: Folders.Update: %v", err)
	}

	// ── Bulk ─────────────────────────────────────────────────────────────────
	bulk, err := client.Bulk.Create(ctx, awsysco.BulkCreateInput{
		URLs: []awsysco.BulkLinkInput{
			{URL: "https://example.com/page1"},
			{URL: "https://example.com/page2"},
			{URL: "https://example.com/page3"},
		},
	})
	if err != nil {
		log.Printf("warning: Bulk.Create: %v", err)
	} else {
		fmt.Printf("\nBulk created %d links (%d failed)\n", bulk.Created, bulk.Failed)
	}

	// ── Webhooks ─────────────────────────────────────────────────────────────
	webhook, err := client.Webhooks.Create(ctx, awsysco.CreateWebhookInput{
		URL:    "https://example.com/webhook",
		Events: []string{"link.created", "link.click"},
		Name:   "go-sdk-demo",
	})
	if err != nil {
		log.Printf("warning: Webhooks.Create: %v", err)
	} else {
		fmt.Printf("\nCreated webhook: id=%s url=%s\n", webhook.ID, webhook.URL)
		_, _ = client.Webhooks.Delete(ctx, webhook.ID)
	}

	// ── Namespace ─────────────────────────────────────────────────────────────
	nsInfo, err := client.Namespace.Get(ctx)
	if err != nil {
		log.Printf("warning: Namespace.Get: %v", err)
	} else {
		fmt.Printf("\nNamespace access: hasAccess=%v tier=%s\n", nsInfo.HasAccess, nsInfo.Tier)
	}

	// ── UTM Templates ──────────────────────────────────────────────────────────
	utmResp, err := client.UtmTemplates.Create(ctx, awsysco.CreateUtmTemplateInput{
		Name:     "go-sdk-email",
		Source:   "newsletter",
		Medium:   "email",
		Campaign: "go-sdk-demo",
	})
	if err != nil {
		log.Printf("warning: UtmTemplates.Create: %v", err)
	} else {
		fmt.Printf("\nCreated UTM template: id=%s name=%s\n", utmResp.Template.ID, utmResp.Template.Name)
		_, _ = client.UtmTemplates.Delete(ctx, utmResp.Template.ID)
	}

	// ── Saved Views ───────────────────────────────────────────────────────────
	views, err := client.SavedViews.List(ctx)
	if err != nil {
		log.Printf("warning: SavedViews.List: %v", err)
	} else {
		fmt.Printf("\nSaved views: %d\n", len(views))
	}

	// ── Affiliate ─────────────────────────────────────────────────────────────
	program, err := client.Affiliate.CreateProgram(ctx, awsysco.CreateAffiliateProgramInput{
		Name:           "Go SDK Demo Program",
		CommissionType: "cpc",
		CpcRate:        0.05,
	})
	if err != nil {
		log.Printf("warning: Affiliate.CreateProgram: %v", err)
	} else {
		fmt.Printf("\nCreated affiliate program: id=%s name=%s\n", program.ID, program.Name)
	}

	// ── Data Export ───────────────────────────────────────────────────────────
	csv, err := client.DataExport.ExportLinks(ctx)
	if err != nil {
		log.Printf("warning: DataExport.ExportLinks: %v", err)
	} else {
		lines := 0
		for _, c := range csv {
			if c == '\n' {
				lines++
			}
		}
		fmt.Printf("\nExported links CSV: ~%d lines\n", lines)
	}

	// ── Cleanup ───────────────────────────────────────────────────────────────
	_ = client.Folders.RemoveLink(ctx, link.ShortCode)
	_ = client.Links.Delete(ctx, link.ShortCode)
	_ = client.Folders.Delete(ctx, folder.ID)
	fmt.Println("\nCleanup complete.")
}
