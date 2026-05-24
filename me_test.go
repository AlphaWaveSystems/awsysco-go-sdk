package awsysco_test

import (
	"context"
	"testing"
)

func TestMeGet(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	me, err := client.Me.Get(ctx)
	if err != nil {
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
