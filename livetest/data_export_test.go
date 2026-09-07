package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestDataExportExportLinks(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	csv, err := client.DataExport.ExportLinks(ctx)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("DataExport.ExportLinks not available on this environment: %v", err)
		}
		t.Fatalf("DataExport.ExportLinks failed: %v", err)
	}
	if csv == "" {
		t.Fatal("expected non-empty CSV response")
	}
	t.Logf("exported links CSV: %d bytes", len(csv))
}

func TestDataExportExportLinkStats(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
		URL: "https://example.com/go-sdk-test-export-stats",
	})
	skipSetupError(t, "Links.Create (setup)", err)
	ref := link.ShortCode
	if ref == "" {
		ref = link.ID
	}
	defer func() { _ = client.Links.Delete(ctx, ref) }()

	csv, err := client.DataExport.ExportLinkStats(ctx, ref)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) {
			t.Skipf("DataExport.ExportLinkStats not available on this environment: %v", err)
		}
		t.Fatalf("DataExport.ExportLinkStats failed: %v", err)
	}
	t.Logf("exported link stats CSV: %d bytes", len(csv))
}
