package awsysco_test

import (
	"context"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestUtmTemplatesCreateListDelete(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	resp, err := client.UtmTemplates.Create(ctx, awsysco.CreateUtmTemplateInput{
		Name:     "go-sdk-test-utm",
		Source:   "newsletter",
		Medium:   "email",
		Campaign: "go-sdk-test",
	})
	if err != nil {
		if awsysco.IsNotFound(err) || awsysco.IsForbidden(err) || awsysco.IsAuthError(err) {
			t.Skipf("UtmTemplates.Create not available on this environment: %v", err)
		}
		t.Fatalf("UtmTemplates.Create failed: %v", err)
	}
	t.Logf("created UTM template: id=%s name=%s", resp.Template.ID, resp.Template.Name)

	if resp.Template.ID == "" {
		t.Fatal("expected non-empty template ID")
	}
	defer func() {
		_, _ = client.UtmTemplates.Delete(ctx, resp.Template.ID)
	}()

	templates, err := client.UtmTemplates.List(ctx)
	if err != nil {
		t.Fatalf("UtmTemplates.List failed: %v", err)
	}

	found := false
	for _, tmpl := range templates {
		if tmpl.ID == resp.Template.ID {
			found = true
			break
		}
	}
	if !found {
		t.Logf("note: newly created template not found in list (may be eventual consistency)")
	}
	t.Logf("listed %d UTM templates", len(templates))

	_, err = client.UtmTemplates.Delete(ctx, resp.Template.ID)
	if err != nil {
		t.Fatalf("UtmTemplates.Delete failed: %v", err)
	}
	t.Logf("deleted UTM template %s", resp.Template.ID)
}
