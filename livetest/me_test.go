package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestMeGet(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	me, err := client.Me.Get(ctx)
	if err != nil {
		if awsysco.IsRateLimitError(err) {
			t.Skipf("Me.Get rate-limited (hourly limit): %v", err)
		}
		t.Fatalf("Me.Get failed: %v", err)
	}
	if me.Email == "" {
		t.Fatal("expected Email to be non-empty")
	}
	if me.SubscriptionTier == "" {
		t.Fatal("expected SubscriptionTier to be non-empty")
	}
	t.Logf("me: email=%s tier=%s premium=%v", me.Email, me.SubscriptionTier, me.IsPremium)
}
