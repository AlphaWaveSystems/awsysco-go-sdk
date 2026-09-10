package awsysco_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

// This file covers the testdata/sdk-contract.json "behaviors" array —
// cross-cutting SDK behavior that isn't tied to a single capability or error
// fixture. See each test's doc comment for the specific behavior id it
// covers.

// TestBehaviorUserAgent covers the "user_agent" behavior: the User-Agent
// header must match ^awsysco-go-sdk/\d+\.\d+\.\d+ and its version component
// must equal the package version.
func TestBehaviorUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	if _, err := client.Links.Get(context.Background(), "x"); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	re := regexp.MustCompile(`^awsysco-go-sdk/\d+\.\d+\.\d+$`)
	if !re.MatchString(gotUA) {
		t.Errorf("User-Agent = %q, want match of %s", gotUA, re.String())
	}
	want := "awsysco-go-sdk/" + awsysco.TestingSDKVersion
	if gotUA != want {
		t.Errorf("User-Agent = %q, want %q (package version)", gotUA, want)
	}
}

// TestBehaviorUnknownFieldsPreserved covers "unknown_fields_preserved":
// extra/unexpected JSON fields in a response must not raise, and known
// fields must still populate correctly.
func TestBehaviorUnknownFieldsPreserved(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"id": "abc123",
			"shortUrl": "https://awsys.co/abc123",
			"shortCode": "abc123",
			"long": "https://example.com/",
			"clicks": 7,
			"totallyNewField": {"nested": true, "n": 42},
			"anotherNewField": [1, 2, 3],
			"yetAnotherNewField": null,
			"aFutureBooleanFlag": true
		}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	link, err := client.Links.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed on a response with unexpected extra fields: %v", err)
	}
	if link.ID != "abc123" {
		t.Errorf("ID = %q, want abc123", link.ID)
	}
	if link.Clicks != 7 {
		t.Errorf("Clicks = %d, want 7", link.Clicks)
	}
	if link.Long != "https://example.com/" {
		t.Errorf("Long = %q, want https://example.com/", link.Long)
	}
}

// folderWithTimestamp is a minimal JSON document exercising firestoreTimestamp
// via Folder.Created, without depending on any other Folder field.
func folderJSON(createdRaw string) string {
	return `{"id":"f1","name":"n","created":` + createdRaw + `}`
}

// TestBehaviorTimestampVariants covers "timestamp_variants": a plain ISO
// string and a valid Firestore {_seconds,_nanoseconds} object must both
// parse to the expected instant. It also exercises the bare (no underscore)
// {seconds,nanoseconds} form the platform has also been observed to emit.
func TestBehaviorTimestampVariants(t *testing.T) {
	want := time.Date(2026, 1, 15, 12, 30, 0, 0, time.UTC)

	cases := []struct {
		name string
		raw  string
	}{
		{"iso_string", `"2026-01-15T12:30:00Z"`},
		{"firestore_underscore", `{"_seconds":1768480200,"_nanoseconds":0}`},
		{"firestore_bare", `{"seconds":1768480200,"nanoseconds":0}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var f awsysco.Folder
			if err := json.Unmarshal([]byte(folderJSON(c.raw)), &f); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if !f.Created.Equal(want) {
				t.Errorf("Created = %v, want %v", f.Created, want)
			}
		})
	}
}

// TestBehaviorTimestampNeverRaises covers "timestamp_never_raises": the
// timestamp normalizer must never error for any of the four garbage shapes
// from the contract's assert text, even though the resulting field ends up
// zero/unparsed — decoding the containing struct must still succeed.
func TestBehaviorTimestampNeverRaises(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"string_nanoseconds", `{"seconds":1,"nanoseconds":"q"}`},
		{"seconds_overflow", `{"_seconds":1e300}`},
		{"seconds_large_negative", `{"seconds":-1e14}`},
		{"seconds_as_array", `{"seconds":[1]}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var f awsysco.Folder
			err := json.Unmarshal([]byte(folderJSON(c.raw)), &f)
			if err != nil {
				t.Fatalf("Unmarshal returned an error for garbage shape %s, want nil: %v", c.raw, err)
			}
		})
	}

	// Also exercise it as the response body of a real API call, not just a
	// direct json.Unmarshal — the normalizer must not raise there either.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"folders":[` + folderJSON(`{"seconds":[1]}`) + `],"limit":10,"used":1}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	resp, err := client.Folders.List(context.Background())
	if err != nil {
		t.Fatalf("Folders.List failed on a garbage timestamp shape: %v", err)
	}
	if len(resp.Folders) != 1 {
		t.Fatalf("got %d folders, want 1", len(resp.Folders))
	}
}

// TestBehaviorIteratorLinksLimitZero covers "iterator_links_limit_zero":
// Iter with a zero or negative Limit must clamp to a sane positive value
// (rather than sending limit=0 or a negative limit) and terminate normally.
func TestBehaviorIteratorLinksLimitZero(t *testing.T) {
	for _, tc := range []struct {
		name  string
		limit int
	}{
		{"zero", 0},
		{"negative", -5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotLimit string
			var reqs int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				reqs++
				gotLimit = r.URL.Query().Get("limit")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"links":[],"hasMore":false}`))
			}))
			defer srv.Close()

			client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
			it := client.Links.Iter(context.Background(), awsysco.ListLinksInput{Limit: tc.limit})

			if it.Next() {
				t.Fatal("Next() = true on an empty page, want false")
			}
			if it.Err() != nil {
				t.Errorf("Err() = %v, want nil", it.Err())
			}
			if gotLimit == "" || gotLimit == "0" {
				t.Errorf("limit query param = %q, want a clamped positive value", gotLimit)
			}
			if reqs != 1 {
				t.Fatalf("requests = %d, want 1 (must terminate, not loop forever)", reqs)
			}
		})
	}
}

// TestBehaviorLinksListHasMoreFromPagination covers
// "links_list_has_more_from_pagination": HasMore must be read from the
// nested pagination.hasMore field, not a top-level hasMore key — and when
// both are present and disagree, pagination wins.
func TestBehaviorLinksListHasMoreFromPagination(t *testing.T) {
	cases := []struct {
		name string
		body string
		want bool
	}{
		{
			name: "pagination_only_true",
			body: `{"links":[],"pagination":{"hasMore":true}}`,
			want: true,
		},
		{
			name: "pagination_only_false",
			body: `{"links":[],"pagination":{"hasMore":false}}`,
			want: false,
		},
		{
			name: "pagination_overrides_conflicting_top_level",
			body: `{"links":[],"hasMore":false,"pagination":{"hasMore":true}}`,
			want: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var resp awsysco.ListLinksResponse
			if err := json.Unmarshal([]byte(c.body), &resp); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if resp.HasMore != c.want {
				t.Errorf("HasMore = %v, want %v", resp.HasMore, c.want)
			}
		})
	}
}

// TestBehaviorIteratorLinksFixtureScenario covers the "iterator_links"
// behavior precisely: draining Links.Iter over the real list_links →
// list_links_last_page fixture scenarios must yield exactly 3 links across
// exactly 2 requests.
func TestBehaviorIteratorLinksFixtureScenario(t *testing.T) {
	fixture := loadContractFixture(t)
	bodies := map[string]json.RawMessage{}
	for _, cap := range fixture.Capabilities {
		if cap.ID == "list_links" || cap.ID == "list_links_last_page" {
			bodies[cap.ID] = cap.Response.Body
		}
	}
	if len(bodies) != 2 {
		t.Fatalf("expected list_links and list_links_last_page fixtures in testdata/sdk-contract.json, got %d", len(bodies))
	}

	var reqs int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("offset") == "2" {
			_, _ = w.Write(bodies["list_links_last_page"])
			return
		}
		_, _ = w.Write(bodies["list_links"])
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	it := client.Links.Iter(context.Background(), awsysco.ListLinksInput{Limit: 2})

	var got []string
	for it.Next() {
		got = append(got, it.Link().ShortCode)
	}
	if it.Err() != nil {
		t.Fatalf("Err() = %v, want nil", it.Err())
	}
	if reqs != 2 {
		t.Errorf("requests = %d, want 2", reqs)
	}
	want := []string{"abc123", "def456", "ghi789"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("link[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestBehaviorRedactionStrFormat covers "redaction_str": %s formatting (not
// only %v/%+v/%#v) of the client must redact the API key, and any model
// carrying a secret (CreateWebhookInput/UpdateWebhookInput.Secret,
// ImportStartOptions.AccessToken) must redact it across %v/%+v/%s/%#v, both
// by value and by pointer.
func TestBehaviorRedactionStrFormat(t *testing.T) {
	client := awsysco.NewClient("awsys_supersecretclientkey1234", awsysco.WithBaseURL("https://example.com"))
	checkNoLeak(t, "Client", client, "supersecretclientkey1234")

	wh := awsysco.CreateWebhookInput{
		URL:    "https://hooks.example/",
		Events: []string{"link.created"},
		Secret: "whsec_averysecretsigningvalue",
	}
	checkNoLeak(t, "CreateWebhookInput (value)", wh, "averysecretsigningvalue")
	checkNoLeak(t, "CreateWebhookInput (pointer)", &wh, "averysecretsigningvalue")

	upd := awsysco.UpdateWebhookInput{Secret: "whsec_anotherverysecretvalue"}
	checkNoLeak(t, "UpdateWebhookInput (value)", upd, "anotherverysecretvalue")
	checkNoLeak(t, "UpdateWebhookInput (pointer)", &upd, "anotherverysecretvalue")

	imp := awsysco.ImportStartOptions{Provider: "bitly", AccessToken: "bitly_supersecretaccesstoken"}
	checkNoLeak(t, "ImportStartOptions (value)", imp, "supersecretaccesstoken")
	checkNoLeak(t, "ImportStartOptions (pointer)", &imp, "supersecretaccesstoken")
}

func checkNoLeak(t *testing.T, label string, v interface{}, secretFragment string) {
	t.Helper()
	forms := map[string]string{
		"%v":  fmt.Sprintf("%v", v),
		"%+v": fmt.Sprintf("%+v", v),
		"%s":  fmt.Sprintf("%s", v),
		"%#v": fmt.Sprintf("%#v", v),
	}
	for verb, out := range forms {
		if strings.Contains(out, secretFragment) {
			t.Errorf("%s %s output leaked the secret: %q", label, verb, out)
		}
	}
}

// TestBehaviorConfigWarnings covers "config_warnings": constructing a client
// with an API key that doesn't start with "awsys_", or with a non-https
// base URL, must each emit one warning-level log line via the standard log
// package (no new logging dependency).
func TestBehaviorConfigWarnings(t *testing.T) {
	orig := log.Writer()
	defer log.SetOutput(orig)

	var buf bytes.Buffer

	log.SetOutput(&buf)
	buf.Reset()
	_ = awsysco.NewClient("not_a_real_awsys_key")
	if !strings.Contains(buf.String(), "does not start with") {
		t.Errorf("expected a warning for a key not starting with awsys_, got: %q", buf.String())
	}

	log.SetOutput(&buf)
	buf.Reset()
	_ = awsysco.NewClient("awsys_validlookingkey", awsysco.WithBaseURL("http://insecure.example.com"))
	if !strings.Contains(buf.String(), "does not use https") {
		t.Errorf("expected a warning for a non-https base URL, got: %q", buf.String())
	}

	log.SetOutput(&buf)
	buf.Reset()
	_ = awsysco.NewClient("awsys_validlookingkey", awsysco.WithBaseURL("https://example.com"))
	if buf.Len() != 0 {
		t.Errorf("expected no warning for a valid awsys_ key + https URL, got: %q", buf.String())
	}
}

// TestIsNetworkErrorAndIsSDKErrorMatchTimeoutError is a regression test for
// the new IsNetworkError/IsSDKError predicates: *TimeoutError embeds
// NetworkError BY VALUE with no explicit Unwrap() override of its own, so
// errors.As(err, &networkErrPtr) does NOT match a *TimeoutError purely via
// promotion (Unwrap resolves to the underlying transport error, not the
// embedded NetworkError struct) — both predicates must check *TimeoutError
// explicitly rather than assuming the embedding makes it "fall out for
// free".
func TestIsNetworkErrorAndIsSDKErrorMatchTimeoutError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL), awsysco.WithTimeout(10*time.Millisecond), awsysco.WithMaxRetries(0))
	_, err := client.Links.Get(context.Background(), "x")
	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if !awsysco.IsTimeoutError(err) {
		t.Error("IsTimeoutError = false, want true")
	}
	if !awsysco.IsNetworkError(err) {
		t.Error("IsNetworkError = false, want true (a timeout is a transport-level problem)")
	}
	if !awsysco.IsSDKError(err) {
		t.Error("IsSDKError = false, want true")
	}
}

// TestBehaviorBodyReadWithinTimeout covers "body_read_within_timeout": the
// client timeout must cover the response body read, not just the headers —
// a server that sends headers, flushes, then stalls the body must still
// raise *TimeoutError.
func TestBehaviorBodyReadWithinTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// Stall well past the client's timeout before writing the body.
		time.Sleep(150 * time.Millisecond)
		_, _ = w.Write([]byte(`{"shortCode":"abc123"}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient(
		"awsys_test",
		awsysco.WithBaseURL(srv.URL),
		awsysco.WithTimeout(20*time.Millisecond),
		awsysco.WithMaxRetries(0),
	)

	start := time.Now()
	_, err := client.Links.Get(context.Background(), "abc123")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	var te *awsysco.TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("expected *awsysco.TimeoutError, got %T: %v", err, err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("elapsed = %v, want well under the 150ms body stall", elapsed)
	}
}
