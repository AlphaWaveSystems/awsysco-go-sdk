package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestFoldersCreate(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
		Name:  "go-sdk-test-folder",
		Color: "#3B82F6",
	})
	if err != nil {
		t.Fatalf("Folders.Create failed: %v", err)
	}
	if folder.ID == "" {
		t.Fatal("expected folder ID to be non-empty")
	}
	t.Logf("created folder: id=%s name=%s", folder.ID, folder.Name)

	// Clean up
	_ = client.Folders.Delete(ctx, folder.ID)
}

func TestFoldersList(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	// Create a folder first to ensure the list is non-empty.
	folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
		Name: "go-sdk-test-list-folder",
	})
	if err != nil {
		t.Fatalf("Folders.Create (setup) failed: %v", err)
	}
	defer func() { _ = client.Folders.Delete(ctx, folder.ID) }()

	resp, err := client.Folders.List(ctx)
	if err != nil {
		t.Fatalf("Folders.List failed: %v", err)
	}
	found := false
	for _, f := range resp.Folders {
		if f.ID == folder.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created folder %s not found in list", folder.ID)
	}
	t.Logf("listed %d folders", len(resp.Folders))
}

func TestFoldersAssignLink(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-assign",
	})
	if err != nil {
		t.Fatalf("Links.Create (setup) failed: %v", err)
	}
	// Use shortCode (or ID) as the link identifier for folder assignment.
	linkRef := link.ShortCode
	if linkRef == "" {
		linkRef = link.ID
	}
	defer func() { _ = client.Links.Delete(ctx, linkRef) }()

	folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
		Name: "go-sdk-test-assign-folder",
	})
	if err != nil {
		t.Fatalf("Folders.Create (setup) failed: %v", err)
	}
	defer func() { _ = client.Folders.Delete(ctx, folder.ID) }()

	if err := client.Folders.AssignLink(ctx, linkRef, folder.ID); err != nil {
		t.Fatalf("Folders.AssignLink failed: %v", err)
	}
	t.Logf("assigned link %s to folder %s", linkRef, folder.ID)
}

func TestFoldersRemoveLink(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-remove-folder",
	})
	if err != nil {
		t.Fatalf("Links.Create (setup) failed: %v", err)
	}
	linkRef := link.ShortCode
	if linkRef == "" {
		linkRef = link.ID
	}
	defer func() { _ = client.Links.Delete(ctx, linkRef) }()

	folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
		Name: "go-sdk-test-remove-folder",
	})
	if err != nil {
		t.Fatalf("Folders.Create (setup) failed: %v", err)
	}
	defer func() { _ = client.Folders.Delete(ctx, folder.ID) }()

	// Assign first, then remove.
	if err := client.Folders.AssignLink(ctx, linkRef, folder.ID); err != nil {
		t.Fatalf("Folders.AssignLink (setup) failed: %v", err)
	}

	if err := client.Folders.RemoveLink(ctx, linkRef); err != nil {
		t.Fatalf("Folders.RemoveLink failed: %v", err)
	}
	t.Logf("removed link %s from folder", linkRef)
}

func TestFoldersDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
		Name: "go-sdk-test-delete-folder",
	})
	if err != nil {
		t.Fatalf("Folders.Create (setup) failed: %v", err)
	}

	if err := client.Folders.Delete(ctx, folder.ID); err != nil {
		t.Fatalf("Folders.Delete failed: %v", err)
	}
	t.Logf("deleted folder %s", folder.ID)
}
