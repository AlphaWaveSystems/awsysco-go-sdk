package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestAffiliateGetLimits(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	result, err := client.Affiliate.GetLimits(ctx)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Affiliate.GetLimits not available on this environment: %v", err)
		}
		t.Fatalf("Affiliate.GetLimits failed: %v", err)
	}
	t.Logf("affiliate limits: %v", result)
}

func TestAffiliateDiscover(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	programs, err := client.Affiliate.Discover(ctx, 5)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Affiliate.Discover not available on this environment: %v", err)
		}
		t.Fatalf("Affiliate.Discover failed: %v", err)
	}
	t.Logf("discovered %d affiliate programs", len(programs))
}

func TestAffiliateProgramCreateGetDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	program, err := client.Affiliate.CreateProgram(ctx, awsysco.CreateAffiliateProgramInput{
		Name:           "go-sdk-test-program",
		CommissionType: "cpc",
		CpcRate:        0.05,
	})
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsValidationError(err) || awsysco.IsAuthError(err) {
			t.Skipf("Affiliate.CreateProgram not available on this environment: %v", err)
		}
		t.Fatalf("Affiliate.CreateProgram failed: %v", err)
	}
	t.Logf("created affiliate program: id=%s name=%s", program.ID, program.Name)

	if program.ID == "" {
		t.Fatal("expected non-empty program ID")
	}

	got, err := client.Affiliate.GetProgram(ctx, program.ID)
	if err != nil {
		t.Fatalf("Affiliate.GetProgram failed: %v", err)
	}
	if got.ID != program.ID {
		t.Errorf("program ID mismatch: got %q want %q", got.ID, program.ID)
	}
	t.Logf("fetched program: id=%s", got.ID)

	_, err = client.Affiliate.ListPartners(ctx, program.ID)
	if err != nil {
		t.Fatalf("Affiliate.ListPartners failed: %v", err)
	}
	t.Logf("listed partners for program %s", program.ID)

	programs, err := client.Affiliate.ListPrograms(ctx)
	if err != nil {
		t.Fatalf("Affiliate.ListPrograms failed: %v", err)
	}
	t.Logf("listed %d programs", len(programs))
}

func TestAffiliateListPartnerships(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	partnerships, err := client.Affiliate.ListPartnerships(ctx)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Affiliate.ListPartnerships not available on this environment: %v", err)
		}
		t.Fatalf("Affiliate.ListPartnerships failed: %v", err)
	}
	t.Logf("listed %d partnerships", len(partnerships))
}
