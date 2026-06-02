package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestTagsAddRemove(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-tags",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	ref := link.ShortCode
	if ref == "" {
		ref = link.ID
	}
	defer func() { _ = client.Links.Delete(ctx, ref) }()

	resp, err := client.Tags.Add(ctx, ref, "go-sdk-tag")
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) {
			t.Skipf("Tags.Add not available on this environment: %v", err)
		}
		t.Fatalf("Tags.Add failed: %v", err)
	}
	t.Logf("tags after add: %v", resp.Tags)

	resp2, err := client.Tags.Remove(ctx, ref, "go-sdk-tag")
	if err != nil {
		t.Fatalf("Tags.Remove failed: %v", err)
	}
	t.Logf("tags after remove: %v", resp2.Tags)
}
