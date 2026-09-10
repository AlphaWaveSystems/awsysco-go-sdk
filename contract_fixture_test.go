package awsysco_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// contractFixture mirrors the shape of testdata/sdk-contract.json.
type contractFixture struct {
	Version      string               `json:"version"`
	Capabilities []contractCapability `json:"capabilities"`
}

type contractCapability struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
	Request    struct {
		Method string          `json:"method"`
		Path   string          `json:"path"`
		Query  map[string]any  `json:"query"`
		Body   json.RawMessage `json:"body"`
	} `json:"request"`
	Response struct {
		Status int             `json:"status"`
		Body   json.RawMessage `json:"body"`
	} `json:"response"`
}

func loadContractFixture(t *testing.T) contractFixture {
	t.Helper()
	raw, err := os.ReadFile("testdata/sdk-contract.json")
	if err != nil {
		t.Fatalf("reading testdata/sdk-contract.json: %v", err)
	}
	var f contractFixture
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("parsing testdata/sdk-contract.json: %v", err)
	}
	return f
}

// capabilityInvoker calls the one SDK method that should hit a capability's
// fixture endpoint. It returns an error only for genuine SDK-side failures
// (the httptest server always serves the fixture's canned response, so a
// non-2xx fixture status is expected to surface as a typed SDK error, not a
// test failure — the invoker itself doesn't need to branch on status).
type capabilityInvoker func(ctx context.Context, c *awsysco.Client) error

// invokers maps each fixture id to the SDK call that should produce the
// matching request. Deliberately keyed by id (not the shared "capability"
// number) since sibling scenarios for the same capability (e.g. the four
// list_links variants) exercise the same method with different arguments.
var contractInvokers = map[string]capabilityInvoker{
	"create_link": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Links.Create(ctx, awsysco.CreateLinkInput{URL: "https://example.com/"})
		return err
	},
	"create_link_custom_slug": func(ctx context.Context, c *awsysco.Client) error {
		maxClicks := 10
		expires := mustParseTime("2027-01-01T00:00:00.000Z")
		_, err := c.Links.Create(ctx, awsysco.CreateLinkInput{
			URL:        "https://example.com/",
			CustomSlug: "my-slug",
			MaxClicks:  &maxClicks,
			ExpiresAt:  &expires,
		})
		return err
	},
	"list_links": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Links.List(ctx, awsysco.ListLinksInput{Limit: 2, Offset: 0})
		return err
	},
	"list_links_last_page": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Links.List(ctx, awsysco.ListLinksInput{Limit: 2, Offset: 2})
		return err
	},
	"list_links_missing_hasmore": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Links.List(ctx, awsysco.ListLinksInput{Limit: 2, Offset: 0})
		return err
	},
	"list_links_limit_clamped": func(ctx context.Context, c *awsysco.Client) error {
		// Fixture note: "caller passed 500; SDK clamps to 100" — Links.List
		// now clamps Limit above the platform max the same way Iter does,
		// so passing 500 here must still put "limit=100" on the wire.
		_, err := c.Links.List(ctx, awsysco.ListLinksInput{Limit: 500, Offset: 0})
		return err
	},
	"get_link": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Links.Get(ctx, "abc123")
		return err
	},
	"get_link_namespaced": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Links.Get(ctx, "ns/slug")
		return err
	},
	"update_link": func(ctx context.Context, c *awsysco.Client) error {
		maxClicks := 5
		expires := mustParseTime("2027-01-01T00:00:00.000Z")
		_, err := c.Links.Update(ctx, "abc123", awsysco.UpdateLinkInput{MaxClicks: &maxClicks, ExpiresAt: &expires})
		return err
	},
	"delete_link": func(ctx context.Context, c *awsysco.Client) error {
		return c.Links.Delete(ctx, "abc123")
	},
	"link_stats": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Analytics.GetStats(ctx, "abc123", "7d")
		return err
	},
	"aggregate_stats": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Analytics.GetAggregateStats(ctx, "abc123", &awsysco.AggregateOptions{Period: "7d"})
		return err
	},
	"bulk_create": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Bulk.Create(ctx, awsysco.BulkCreateInput{URLs: []awsysco.BulkLinkInput{
			{URL: "https://a.example/"},
			{URL: "https://b.example/", CustomSlug: "b"},
		}})
		return err
	},
	"me": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Me.Get(ctx)
		return err
	},
	"usage": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Usage.Get(ctx)
		return err
	},
	"recent_clicks": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Analytics.GetRecentClicks(ctx, 10)
		return err
	},
	"profile_get": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Profile.Get(ctx)
		return err
	},
	"profile_update": func(ctx context.Context, c *awsysco.Client) error {
		name := "New"
		return c.Profile.Update(ctx, awsysco.ProfileUpdateInput{DisplayName: &name})
	},
	// qr_url has no corresponding invoker: QR.GetURL is a pure local URL
	// builder and never makes an HTTP request — see TestContractFixtures'
	// explicit skip list.
	"qr_settings_get": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.QR.GetSettings(ctx, "abc123")
		return err
	},
	"qr_settings_update": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.QR.UpdateSettings(ctx, "abc123", awsysco.QRSettings{Color: "#ff0000"})
		return err
	},
	"folders_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Folders.List(ctx)
		return err
	},
	"folder_create": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Folders.Create(ctx, awsysco.CreateFolderInput{Name: "Work"})
		return err
	},
	"folder_update": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Folders.Update(ctx, "f1", awsysco.UpdateFolderInput{Name: "Work2"})
		return err
	},
	"folder_delete": func(ctx context.Context, c *awsysco.Client) error {
		return c.Folders.Delete(ctx, "f1")
	},
	"folder_assign": func(ctx context.Context, c *awsysco.Client) error {
		return c.Folders.AssignLink(ctx, "abc123", "f1")
	},
	"folder_remove": func(ctx context.Context, c *awsysco.Client) error {
		return c.Folders.RemoveLink(ctx, "abc123")
	},
	"tags_add": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Tags.Add(ctx, "abc123", "a", "b")
		return err
	},
	"tag_remove": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Tags.Remove(ctx, "abc123", "a")
		return err
	},
	"views_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.SavedViews.List(ctx)
		return err
	},
	"view_create": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.SavedViews.Create(ctx, awsysco.CreateSavedViewInput{Name: "Mine", Filters: awsysco.SavedViewFilters{Tag: "a"}})
		return err
	},
	"view_update": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.SavedViews.Update(ctx, "v1", awsysco.UpdateSavedViewInput{Name: "Yours"})
		return err
	},
	"view_delete": func(ctx context.Context, c *awsysco.Client) error {
		return c.SavedViews.Delete(ctx, "v1")
	},
	"utm_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.UtmTemplates.List(ctx)
		return err
	},
	"utm_create": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.UtmTemplates.Create(ctx, awsysco.CreateUtmTemplateInput{Name: "Launch", Source: "newsletter", Medium: "email", Campaign: "sept"})
		return err
	},
	"utm_delete": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.UtmTemplates.Delete(ctx, "t1")
		return err
	},
	"webhook_event_types": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Webhooks.ListEventTypes(ctx)
		return err
	},
	"webhooks_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Webhooks.List(ctx)
		return err
	},
	"webhook_create": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Webhooks.Create(ctx, awsysco.CreateWebhookInput{URL: "https://h.example/", Events: []string{"link.created"}})
		return err
	},
	"webhook_update": func(ctx context.Context, c *awsysco.Client) error {
		enabled := false
		_, err := c.Webhooks.Update(ctx, "w1", awsysco.UpdateWebhookInput{Enabled: &enabled})
		return err
	},
	"webhook_delete": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Webhooks.Delete(ctx, "w1")
		return err
	},
	"webhook_test": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Webhooks.Test(ctx, "w1", "link.created")
		return err
	},
	"domains_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.CustomDomains.List(ctx)
		return err
	},
	"domain_add": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.CustomDomains.Add(ctx, "go.example.com")
		return err
	},
	"domain_verify": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.CustomDomains.Verify(ctx, "go.example.com")
		return err
	},
	// domain_activate_deprecated is intentionally NOT mapped here: since
	// Batch 3, CustomDomains.Activate returns a 403 error locally without
	// making any network call (ADR-006) — see the dedicated deprecation
	// check in TestCustomDomainsActivateDeprecated instead.
	"domain_update": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.CustomDomains.Update(ctx, "go.example.com", awsysco.UpdateDomainInput{DefaultRedirect: "https://example.com/"})
		return err
	},
	"domain_remove": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.CustomDomains.Remove(ctx, "go.example.com")
		return err
	},
	"domain_check": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.CustomDomains.Check(ctx, "go.example.com")
		return err
	},
	"namespace_get": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Namespace.Get(ctx)
		return err
	},
	"namespace_check": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Namespace.Check(ctx, "acme")
		return err
	},
	"namespace_claim": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Namespace.Claim(ctx, "acme")
		return err
	},
	"namespace_release": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Namespace.Release(ctx)
		return err
	},
	"affiliate_program_create": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.CreateProgram(ctx, awsysco.CreateAffiliateProgramInput{Name: "P", CommissionRate: 10})
		return err
	},
	"affiliate_programs_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.ListPrograms(ctx)
		return err
	},
	"affiliate_program_get": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.GetProgram(ctx, "p1")
		return err
	},
	"affiliate_program_update": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.UpdateProgram(ctx, "p1", awsysco.CreateAffiliateProgramInput{Name: "P2"})
		return err
	},
	"affiliate_program_stats": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.GetProgramStats(ctx, "p1", "30d")
		return err
	},
	"affiliate_partners_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.ListPartners(ctx, "p1")
		return err
	},
	"affiliate_partner_status": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.UpdatePartnerStatus(ctx, "p1", "pt1", "approved")
		return err
	},
	"affiliate_discover": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.Discover(ctx, 20)
		return err
	},
	"affiliate_join": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.Join(ctx, "p9", awsysco.JoinProgramInput{PartnerCode: "CODE"})
		return err
	},
	"affiliate_partnerships_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.ListPartnerships(ctx)
		return err
	},
	"affiliate_partnership_stats": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.GetPartnershipStats(ctx, "ps1", "30d")
		return err
	},
	"affiliate_leave": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.LeaveProgram(ctx, "ps1")
		return err
	},
	"affiliate_limits": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Affiliate.GetLimits(ctx)
		return err
	},
	"agentlink_link_stats": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Agentlink.GetLinkStats(ctx, "abc123", 30)
		return err
	},
	"agentlink_account_stats": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Agentlink.GetAccountStats(ctx, 30)
		return err
	},
	"agentlink_subscribe": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Agentlink.Subscribe(ctx, "t@example.com")
		return err
	},
	"web2app_consume": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Web2App.ConsumeSession(ctx, "tok123")
		return err
	},
	"import_start": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Imports.Start(ctx, awsysco.ImportStartOptions{Provider: "bitly", AccessToken: "bitly_token", ScanOnly: true})
		return err
	},
	"imports_list": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Imports.List(ctx, &awsysco.ImportListOptions{Limit: 20})
		return err
	},
	"import_get": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Imports.GetStatus(ctx, "j1")
		return err
	},
	"import_cancel": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Imports.Cancel(ctx, "j1")
		return err
	},
	"import_redirect_map_csv": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Imports.GetRedirectMapCSV(ctx, "j1")
		return err
	},
	"import_redirect_map_json": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.Imports.GetRedirectMapJSON(ctx, "j1")
		return err
	},
	"export_links_csv": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.DataExport.ExportLinks(ctx)
		return err
	},
	"export_link_stats_csv": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.DataExport.ExportLinkStats(ctx, "abc123")
		return err
	},
	"trust_scan": func(ctx context.Context, c *awsysco.Client) error {
		_, err := c.TrustScore.Scan(ctx, "abc123")
		return err
	},
}

// fixtures with no corresponding SDK-method invoker, and why.
var contractSkips = map[string]string{
	"qr_url":                     "QR.GetURL is a pure local URL builder — it never makes an HTTP request, there is nothing to contract-test against an httptest server",
	"domain_activate_deprecated": "CustomDomains.Activate is deprecated (ADR-006) and returns an error locally without any network call — see TestCustomDomainsActivateDeprecated",
}

// TestCustomDomainsActivateDeprecated covers the domain_activate_deprecated
// fixture's intent directly: Activate must fail fast with no network call.
func TestCustomDomainsActivateDeprecated(t *testing.T) {
	var hit bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	_, err := client.CustomDomains.Activate(context.Background(), "go.example.com")

	if err == nil {
		t.Fatal("Activate should return an error, got nil")
	}
	if !awsysco.IsForbidden(err) {
		t.Errorf("Activate error should be a 403 (IsForbidden), got: %v", err)
	}
	if hit {
		t.Error("Activate should not make any network call — the deprecated route requires Firebase auth an API key can never satisfy")
	}
}

func TestContractFixtures(t *testing.T) {
	fixture := loadContractFixture(t)

	for _, cap := range fixture.Capabilities {
		cap := cap
		t.Run(cap.ID, func(t *testing.T) {
			if reason, skip := contractSkips[cap.ID]; skip {
				t.Skip(reason)
			}
			invoke, ok := contractInvokers[cap.ID]
			if !ok {
				t.Fatalf("no contract invoker registered for fixture id %q (capability %s, %s %s) — add one to contractInvokers or contractSkips", cap.ID, cap.Capability, cap.Request.Method, cap.Request.Path)
			}

			var gotMethod, gotPath, gotAuth string
			var gotQuery url.Values
			var gotBody []byte

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path
				gotQuery = r.URL.Query()
				gotAuth = r.Header.Get("Authorization")
				gotBody, _ = io.ReadAll(r.Body)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(cap.Response.Status)
				if len(cap.Response.Body) > 0 {
					_, _ = w.Write(cap.Response.Body)
				}
			}))
			defer srv.Close()

			client := awsysco.NewClient("awsys_contracttest", awsysco.WithBaseURL(srv.URL), awsysco.WithMaxRetries(0))

			err := invoke(context.Background(), client)

			if cap.Response.Status >= 200 && cap.Response.Status < 300 {
				if err != nil {
					t.Fatalf("SDK call failed for a fixture with a %d response: %v", cap.Response.Status, err)
				}
			}

			if gotMethod != cap.Request.Method {
				t.Errorf("method = %s, want %s", gotMethod, cap.Request.Method)
			}
			wantPath := cap.Request.Path
			if gotPath != wantPath {
				t.Errorf("path = %s, want %s", gotPath, wantPath)
			}
			if gotAuth != "Bearer awsys_contracttest" {
				t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer awsys_contracttest")
			}
			assertContractQuery(t, cap.Request.Query, gotQuery)
			assertContractBody(t, cap.Request.Body, gotBody)
		})
	}

	// Every fixture id must be either invoked or explicitly skipped —
	// catches new capabilities added to the contract with no coverage here.
	covered := 0
	for _, cap := range fixture.Capabilities {
		if _, ok := contractInvokers[cap.ID]; ok {
			covered++
		} else if _, ok := contractSkips[cap.ID]; ok {
			covered++
		}
	}
	if covered != len(fixture.Capabilities) {
		t.Errorf("%d/%d fixture capabilities have no invoker or skip entry", len(fixture.Capabilities)-covered, len(fixture.Capabilities))
	}
}

// findFixture returns the single capability with the given id, failing the
// test immediately if it's missing (a fixture rename/removal should be
// caught here, not silently skipped).
func findFixture(t *testing.T, fixture contractFixture, id string) contractCapability {
	t.Helper()
	for _, cap := range fixture.Capabilities {
		if cap.ID == id {
			return cap
		}
	}
	t.Fatalf("fixture id %q not found in testdata/sdk-contract.json", id)
	return contractCapability{}
}

// fixtureServer spins up an httptest server that always returns the given
// capability's canned response, and returns a client pointed at it.
func fixtureServer(t *testing.T, cap contractCapability) *awsysco.Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(cap.Response.Status)
		if len(cap.Response.Body) > 0 {
			_, _ = w.Write(cap.Response.Body)
		}
	}))
	t.Cleanup(srv.Close)
	return awsysco.NewClient("awsys_fieldtest", awsysco.WithBaseURL(srv.URL), awsysco.WithMaxRetries(0))
}

// TestContractFixtureFieldsDecodeThroughTypedAccessors asserts that decoding
// a REAL fixture response body through the SDK's typed struct actually
// populates the fields the fixture body carries — not just that the call
// returns no error (ADR-019: a passing test must prove the SDK's typed
// model exposes the documented fields, not merely that it echoes the
// server). A wrong or missing json tag decodes silently to a zero value in
// Go, so each assertion below is chosen specifically to have caught a real
// bug found in this SDK during contract-parity work (TrustScoreResult's
// score/status tags, NamespaceInfo's missing canClaim*/namespaceData
// fields, AggregateAnalytics' missing botClicksExcluded).
func TestContractFixtureFieldsDecodeThroughTypedAccessors(t *testing.T) {
	fixture := loadContractFixture(t)
	ctx := context.Background()

	t.Run("trust_scan", func(t *testing.T) {
		cap := findFixture(t, fixture, "trust_scan")
		client := fixtureServer(t, cap)
		result, err := client.TrustScore.Scan(ctx, "abc123")
		if err != nil {
			t.Fatalf("TrustScore.Scan: %v", err)
		}
		if result.Short == "" {
			t.Error("Short is empty — expected the fixture's \"short\" value")
		}
		if result.Status == nil || *result.Status == "" {
			t.Error("Status is nil/empty — trustStatus tag regressed to score/status?")
		}
		if result.Score == nil {
			t.Error("Score is nil — trustScore tag regressed to score/status?")
		}
		if result.Source == "" {
			t.Error("Source is empty — expected the fixture's \"source\" value")
		}
		if result.CreatedAt == nil {
			t.Error("CreatedAt is nil — expected the fixture's \"createdAt\" value")
		}
	})

	t.Run("namespace_get", func(t *testing.T) {
		cap := findFixture(t, fixture, "namespace_get")
		client := fixtureServer(t, cap)
		info, err := client.Namespace.Get(ctx)
		if err != nil {
			t.Fatalf("Namespace.Get: %v", err)
		}
		if !info.HasAccess {
			t.Error("HasAccess is false — expected true from the fixture")
		}
		if info.Namespace == nil || *info.Namespace == "" {
			t.Error("Namespace is nil/empty")
		}
		if !info.CanClaimCustomDomain {
			t.Error("CanClaimCustomDomain is false — expected true from the fixture")
		}
		if len(info.NamespaceData) == 0 {
			t.Error("NamespaceData is empty — expected the fixture's nested object")
		}
	})

	t.Run("aggregate_stats", func(t *testing.T) {
		cap := findFixture(t, fixture, "aggregate_stats")
		client := fixtureServer(t, cap)
		agg, err := client.Analytics.GetAggregateStats(ctx, "abc123", &awsysco.AggregateOptions{Period: "7d"})
		if err != nil {
			t.Fatalf("Analytics.GetAggregateStats: %v", err)
		}
		if len(agg.ClicksByDay) == 0 {
			t.Error("ClicksByDay is empty — expected one entry from the fixture")
		}
		if len(agg.CountryBreakdown) == 0 {
			t.Error("CountryBreakdown is empty — expected the fixture's {MX:1}")
		}
		if agg.DeviceBreakdown == nil {
			t.Error("DeviceBreakdown is nil")
		}
		if len(agg.HourBreakdown) == 0 {
			t.Error("HourBreakdown is empty — expected one entry from the fixture")
		}
		if agg.Tier == "" {
			t.Error("Tier is empty — expected \"builder\" from the fixture")
		}
	})

	t.Run("affiliate_program_create", func(t *testing.T) {
		cap := findFixture(t, fixture, "affiliate_program_create")
		client := fixtureServer(t, cap)
		program, err := client.Affiliate.CreateProgram(ctx, awsysco.CreateAffiliateProgramInput{Name: "P", CommissionRate: 10})
		if err != nil {
			t.Fatalf("Affiliate.CreateProgram: %v", err)
		}
		if program.MerchantID == "" {
			t.Error("MerchantID is empty")
		}
		if program.CommissionType == "" {
			t.Error("CommissionType is empty")
		}
		if program.CookieDays == 0 {
			t.Error("CookieDays is 0 — cookieDurationDays tag regressed to cookieDays?")
		}
		if program.MaxPartners == 0 {
			t.Error("MaxPartners is 0")
		}
		if program.Status == "" {
			t.Error("Status is empty")
		}
		if !program.IsPublic {
			t.Error("IsPublic is false — expected true from the fixture")
		}
		if program.CreatedAt == nil || *program.CreatedAt == "" {
			t.Error("CreatedAt is nil/empty — Firestore {_seconds,_nanoseconds} normalization broken?")
		}
	})

	t.Run("affiliate_discover", func(t *testing.T) {
		cap := findFixture(t, fixture, "affiliate_discover")
		client := fixtureServer(t, cap)
		programs, err := client.Affiliate.Discover(ctx, 20)
		if err != nil {
			t.Fatalf("Affiliate.Discover: %v", err)
		}
		if len(programs) != 1 {
			t.Fatalf("expected 1 program, got %d", len(programs))
		}
		p := programs[0]
		if p.CommissionType == "" {
			t.Error("CommissionType is empty")
		}
		if p.CookieDays == 0 {
			t.Error("CookieDays is 0 — cookieDurationDays tag regressed to cookieDays?")
		}
		if p.PartnerCount == 0 {
			t.Error("PartnerCount is 0")
		}
		// Discover is a public summary subset — these fields are genuinely
		// absent from the response and must decode as zero, not error.
		if p.Status != "" {
			t.Error("Status should be empty on a discover-subset response")
		}
		if p.MerchantID != "" {
			t.Error("MerchantID should be empty on a discover-subset response")
		}
	})
}

func assertContractQuery(t *testing.T, want map[string]any, got url.Values) {
	t.Helper()
	for k, v := range want {
		wantStr, ok := v.(string)
		if !ok {
			continue // fixture query values are always strings in this contract
		}
		if !got.Has(k) && wantStr == "0" {
			// The SDK omits zero-valued limit/offset from the query string
			// entirely rather than sending an explicit "0" — the platform
			// treats a missing param the same as 0, so this is an accepted,
			// intentional wire-format difference, not a bug.
			continue
		}
		if got.Get(k) != wantStr {
			t.Errorf("query[%s] = %q, want %q (full query: %v)", k, got.Get(k), wantStr, got)
		}
	}
}

func assertContractBody(t *testing.T, want json.RawMessage, gotRaw []byte) {
	t.Helper()
	if len(want) == 0 || string(want) == "null" {
		if len(bytes.TrimSpace(gotRaw)) > 0 && string(bytes.TrimSpace(gotRaw)) != "null" {
			t.Errorf("request body = %s, want empty/no body", gotRaw)
		}
		return
	}
	var wantVal, gotVal any
	if err := json.Unmarshal(want, &wantVal); err != nil {
		t.Fatalf("fixture body is not valid JSON: %v", err)
	}
	if err := json.Unmarshal(gotRaw, &gotVal); err != nil {
		t.Errorf("request body is not valid JSON: %v (body: %s)", err, gotRaw)
		return
	}
	wantMap, wantIsMap := wantVal.(map[string]any)
	gotMap, gotIsMap := gotVal.(map[string]any)
	if !wantIsMap || !gotIsMap {
		return // non-object bodies aren't used by this contract; skip deep compare
	}
	for k, wv := range wantMap {
		gv, present := gotMap[k]
		if !present {
			t.Errorf("request body missing field %q (want %v)", k, wv)
			continue
		}
		if !jsonEqual(wv, gv) {
			t.Errorf("request body field %q = %v, want %v", k, gv, wv)
		}
	}
}

func jsonEqual(a, b any) bool {
	if as, aok := a.(string); aok {
		if bs, bok := b.(string); bok {
			at, aerr := time.Parse(time.RFC3339, as)
			bt, berr := time.Parse(time.RFC3339, bs)
			if aerr == nil && berr == nil {
				// Both sides are timestamps: compare the instant, not the
				// exact formatting (e.g. a trailing ".000" is not
				// significant — time.Time doesn't preserve it on marshal).
				return at.Equal(bt)
			}
		}
	}
	ab, _ := json.Marshal(a)
	bb, _ := json.Marshal(b)
	return string(ab) == string(bb)
}
