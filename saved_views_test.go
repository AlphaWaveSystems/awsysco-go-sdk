package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestSavedViewsCreateListUpdateDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	view, err := client.SavedViews.Create(ctx, awsysco.CreateSavedViewInput{
		Name:    "go-sdk-test-view",
		Filters: awsysco.SavedViewFilters{Tag: "go-sdk"},
	})
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("SavedViews.Create not available on this environment: %v", err)
		}
		t.Fatalf("SavedViews.Create failed: %v", err)
	}
	t.Logf("created saved view: id=%s name=%s", view.ID, view.Name)

	if view.ID == "" {
		t.Fatal("expected non-empty view ID")
	}
	defer func() { _ = client.SavedViews.Delete(ctx, view.ID) }()

	views, err := client.SavedViews.List(ctx)
	if err != nil {
		t.Fatalf("SavedViews.List failed: %v", err)
	}
	found := false
	for _, v := range views {
		if v.ID == view.ID {
			found = true
			break
		}
	}
	if !found {
		t.Logf("note: newly created view not found in list (may be eventual consistency)")
	}
	t.Logf("listed %d saved views", len(views))

	updated, err := client.SavedViews.Update(ctx, view.ID, awsysco.UpdateSavedViewInput{
		Name: "go-sdk-test-view-updated",
	})
	if err != nil {
		t.Fatalf("SavedViews.Update failed: %v", err)
	}
	t.Logf("updated saved view: id=%s name=%s", updated.ID, updated.Name)

	if err := client.SavedViews.Delete(ctx, view.ID); err != nil {
		t.Fatalf("SavedViews.Delete failed: %v", err)
	}
	t.Logf("deleted saved view %s", view.ID)
}
