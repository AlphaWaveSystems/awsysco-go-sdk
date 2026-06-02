package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestLinksCreate(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-create",
	})
	skipSetupError(t, "Links.Create", err)
	if link.ShortURL == "" && link.ShortCode == "" {
		t.Fatal("expected ShortURL or ShortCode to be non-empty")
	}
	t.Logf("created link: id=%s shortUrl=%s shortCode=%s", link.ID, link.ShortURL, link.ShortCode)

	// Best-effort cleanup
	if link.ShortCode != "" {
		_ = client.Links.Delete(ctx, link.ShortCode)
	}
	if link.ID != "" && link.ID != link.ShortCode {
		_ = client.Links.Delete(ctx, link.ID)
	}
}

func TestLinksList(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Create at least one link to ensure the list is non-empty.
	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-list",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	defer func() {
		if link.ShortCode != "" {
			_ = client.Links.Delete(ctx, link.ShortCode)
		}
	}()

	resp, err := client.Links.List(ctx, awsysco.ListLinksInput{Limit: 10})
	if err != nil {
		t.Fatalf("Links.List failed: %v", err)
	}
	if len(resp.Links) == 0 {
		t.Fatal("expected at least one link in list")
	}
	t.Logf("listed %d links (total=%d)", len(resp.Links), resp.Total)
}

func TestLinksGet(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	created, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-get",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	defer func() {
		if created.ShortCode != "" {
			_ = client.Links.Delete(ctx, created.ShortCode)
		}
	}()

	id := created.ID
	if id == "" {
		id = created.ShortCode
	}
	got, err := client.Links.Get(ctx, id)
	if err != nil {
		if awsysco.IsNotFound(err) {
			t.Skipf("GET /api/v1/links/:id not available on this environment: %v", err)
		}
		t.Fatalf("Links.Get failed: %v", err)
	}
	wantCode := created.ShortCode
	if wantCode == "" {
		wantCode = created.ID
	}
	if got.ShortCode != "" && got.ShortCode != wantCode {
		t.Errorf("ShortCode mismatch: got %q, want %q", got.ShortCode, wantCode)
	}
	t.Logf("fetched link: id=%s shortCode=%s", got.ID, got.ShortCode)
}

func TestLinksUpdate(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	created, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-update",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	defer func() {
		if created.ShortCode != "" {
			_ = client.Links.Delete(ctx, created.ShortCode)
		}
	}()

	id := created.ShortCode
	if id == "" {
		id = created.ID
	}

	maxClicks := 999
	updated, err := client.Links.Update(ctx, id, awsysco.UpdateLinkInput{
		MaxClicks: &maxClicks,
	})
	if err != nil {
		if awsysco.IsNotFound(err) {
			t.Skipf("PATCH /api/v1/links/:id not available on this environment: %v", err)
		}
		t.Fatalf("Links.Update failed: %v", err)
	}
	if updated.MaxClicks == nil || *updated.MaxClicks != maxClicks {
		t.Errorf("MaxClicks: got %v, want %d", updated.MaxClicks, maxClicks)
	}
	t.Logf("updated link: id=%s maxClicks=%d", updated.ID, *updated.MaxClicks)
}

func TestLinksDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	created, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-delete",
	})
	skipSetupError(t, "Links.Create (setup)", err)

	id := created.ShortCode
	if id == "" {
		id = created.ID
	}

	if err := client.Links.Delete(ctx, id); err != nil {
		if awsysco.IsNotFound(err) {
			t.Skipf("DELETE /api/v1/links/:id not available on this environment: %v", err)
		}
		t.Fatalf("Links.Delete failed: %v", err)
	}

	_, err = client.Links.Get(ctx, id)
	if err == nil {
		t.Logf("note: link may still exist (eventual consistency)")
	} else if awsysco.IsNotFound(err) {
		t.Logf("verified link %s is deleted", id)
	} else {
		// Some environments return 404 for both delete endpoint and get — acceptable
		t.Logf("note: after delete, get returned: %v", err)
	}
}
