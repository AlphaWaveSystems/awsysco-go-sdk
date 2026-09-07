package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestCustomDomainsList(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	result, err := client.CustomDomains.List(ctx)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("CustomDomains.List not available on this environment: %v", err)
		}
		t.Fatalf("CustomDomains.List failed: %v", err)
	}
	t.Logf("custom domains list: %v", result)
}

func TestCustomDomainsCheck(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	result, err := client.CustomDomains.Check(ctx, "example.com")
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("CustomDomains.Check not available on this environment: %v", err)
		}
		t.Fatalf("CustomDomains.Check failed: %v", err)
	}
	t.Logf("domain check result: %v", result)
}
