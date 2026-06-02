package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestBulkCreate(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	resp, err := client.Bulk.Create(ctx, awsysco.BulkCreateInput{
		URLs: []awsysco.BulkLinkInput{
			{URL: "https://example.com/go-sdk-bulk-1"},
			{URL: "https://example.com/go-sdk-bulk-2"},
			{URL: "https://example.com/go-sdk-bulk-3"},
		},
	})
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("POST /api/v1/bulk not available on this environment: %v", err)
		}
		t.Fatalf("Bulk.Create failed: %v", err)
	}
	if resp.Created != 3 {
		t.Errorf("expected Created=3, got %d (failed=%d)", resp.Created, resp.Failed)
	}
	for i, result := range resp.Results {
		if !result.Success {
			t.Errorf("result[%d] failed: %s", i, result.Error)
			continue
		}
		if result.ShortURL == "" {
			t.Errorf("result[%d]: expected ShortURL to be non-empty", i)
		}
		t.Logf("bulk result[%d]: shortUrl=%s", i, result.ShortURL)
	}

	// Best-effort cleanup via list
	listResp, _ := client.Links.List(ctx, awsysco.ListLinksInput{Limit: 50})
	if listResp != nil {
		targets := map[string]bool{
			"https://example.com/go-sdk-bulk-1": true,
			"https://example.com/go-sdk-bulk-2": true,
			"https://example.com/go-sdk-bulk-3": true,
		}
		for _, link := range listResp.Links {
			if targets[link.Long] {
				id := link.ShortCode
				if id == "" {
					id = link.ID
				}
				_ = client.Links.Delete(ctx, id)
			}
		}
	}
}
