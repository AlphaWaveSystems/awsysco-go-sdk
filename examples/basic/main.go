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

	// Get current user info
	me, err := client.Me.Get(ctx)
	if err != nil {
		log.Fatalf("failed to get user info: %v", err)
	}
	fmt.Printf("Logged in as: %s (tier: %s)\n", me.Email, me.SubscriptionTier)

	// Create a shortened link
	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/very/long/url/that/needs/shortening",
	})
	if err != nil {
		log.Fatalf("failed to create link: %v", err)
	}
	fmt.Printf("Created link: %s -> %s\n", link.ShortURL, link.Long)

	// List recent links
	list, err := client.Links.List(ctx, awsysco.ListLinksInput{Limit: 5})
	if err != nil {
		log.Fatalf("failed to list links: %v", err)
	}
	fmt.Printf("\nRecent links (%d total):\n", list.Total)
	for _, l := range list.Links {
		fmt.Printf("  %s -> %s (%d clicks)\n", l.ShortURL, l.Long, l.Clicks)
	}

	// Get analytics for the created link
	stats, err := client.Analytics.GetStats(ctx, link.ID)
	if err != nil {
		log.Fatalf("failed to get stats: %v", err)
	}
	fmt.Printf("\nStats for %s: %d total clicks\n", link.ShortCode, stats.TotalClicks)

	// Generate a QR code URL
	qrURL := client.QR.GetURL(link.ShortCode, awsysco.WithSize(400))
	fmt.Printf("QR Code URL: %s\n", qrURL)

	// Create a folder and assign the link to it
	folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
		Name:  "Example Folder",
		Color: "#10B981",
	})
	if err != nil {
		log.Fatalf("failed to create folder: %v", err)
	}
	fmt.Printf("\nCreated folder: %s (id: %s)\n", folder.Name, folder.ID)

	if err := client.Folders.AssignLink(ctx, link.ID, folder.ID); err != nil {
		log.Fatalf("failed to assign link to folder: %v", err)
	}
	fmt.Printf("Assigned link %s to folder %s\n", link.ShortCode, folder.Name)

	// Bulk create links
	bulk, err := client.Bulk.Create(ctx, awsysco.BulkCreateInput{
		URLs: []awsysco.BulkLinkInput{
			{URL: "https://example.com/page1"},
			{URL: "https://example.com/page2"},
			{URL: "https://example.com/page3"},
		},
	})
	if err != nil {
		log.Fatalf("failed to bulk create: %v", err)
	}
	fmt.Printf("\nBulk created %d links (%d failed)\n", bulk.Created, bulk.Failed)

	// Clean up the example link and folder
	_ = client.Folders.RemoveLink(ctx, link.ID)
	_ = client.Links.Delete(ctx, link.ID)
	_ = client.Folders.Delete(ctx, folder.ID)
	fmt.Println("\nCleanup complete.")
}
