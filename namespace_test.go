package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestNamespaceGet(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	info, err := client.Namespace.Get(ctx)
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Namespace.Get not available on this environment: %v", err)
		}
		t.Fatalf("Namespace.Get failed: %v", err)
	}
	t.Logf("namespace info: hasAccess=%v tier=%s", info.HasAccess, info.Tier)
}

func TestNamespaceCheck(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	result, err := client.Namespace.Check(ctx, "go-sdk-test-ns")
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("Namespace.Check not available on this environment: %v", err)
		}
		t.Fatalf("Namespace.Check failed: %v", err)
	}
	t.Logf("namespace check: namespace=%s available=%v", result.Namespace, result.Available)
}
