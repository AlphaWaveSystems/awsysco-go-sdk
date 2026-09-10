package awsysco_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func linkJSON(code string) string {
	return `{"shortCode":"` + code + `"}`
}

func TestIteratorEmptyFirstPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"links":[],"hasMore":false}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	it := client.Links.Iter(context.Background(), awsysco.ListLinksInput{})

	if it.Next() {
		t.Fatal("Next() = true on an empty first page, want false")
	}
	if it.Err() != nil {
		t.Errorf("Err() = %v, want nil", it.Err())
	}
}

func TestIteratorTwoPages(t *testing.T) {
	var reqs int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs++
		w.Header().Set("Content-Type", "application/json")
		if reqs == 1 {
			_, _ = w.Write([]byte(`{"links":[` + linkJSON("a") + `,` + linkJSON("b") + `],"hasMore":true}`))
			return
		}
		_, _ = w.Write([]byte(`{"links":[` + linkJSON("c") + `],"hasMore":false}`))
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
	if len(got) != 3 || got[0] != "a" || got[1] != "b" || got[2] != "c" {
		t.Fatalf("got = %v, want [a b c]", got)
	}
	if reqs != 2 {
		t.Fatalf("requests = %d, want 2", reqs)
	}
}

func TestIteratorShortPageStopsDespiteHasMoreTrue(t *testing.T) {
	var reqs int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs++
		w.Header().Set("Content-Type", "application/json")
		// Contradictory: fewer than `limit` items but hasMore:true.
		_, _ = w.Write([]byte(`{"links":[` + linkJSON("a") + `],"hasMore":true}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	it := client.Links.Iter(context.Background(), awsysco.ListLinksInput{Limit: 5})

	var got []string
	for it.Next() {
		got = append(got, it.Link().ShortCode)
	}
	if it.Err() != nil {
		t.Fatalf("Err() = %v, want nil", it.Err())
	}
	if len(got) != 1 {
		t.Fatalf("got = %v, want exactly 1 item (short-page guard should stop iteration)", got)
	}
	if reqs != 1 {
		t.Fatalf("requests = %d, want 1 (must not loop forever)", reqs)
	}
}

func TestIteratorMissingHasMoreStops(t *testing.T) {
	var reqs int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs++
		w.Header().Set("Content-Type", "application/json")
		// hasMore omitted entirely, and the page is empty.
		_, _ = w.Write([]byte(`{"links":[]}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	it := client.Links.Iter(context.Background(), awsysco.ListLinksInput{})

	if it.Next() {
		t.Fatal("Next() = true, want false")
	}
	if it.Err() != nil {
		t.Errorf("Err() = %v, want nil", it.Err())
	}
	if reqs != 1 {
		t.Fatalf("requests = %d, want 1 (must not loop forever)", reqs)
	}
}

func TestIteratorClampsLimitTo100(t *testing.T) {
	var gotLimit string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLimit = r.URL.Query().Get("limit")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"links":[],"hasMore":false}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	it := client.Links.Iter(context.Background(), awsysco.ListLinksInput{Limit: 500})
	it.Next()

	if gotLimit != "100" {
		t.Errorf("limit query param = %q, want 100 (clamped to platform max)", gotLimit)
	}
}

func TestIteratorErrorMidPagination(t *testing.T) {
	var reqs int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqs++
		if reqs == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"links":[` + linkJSON("a") + `,` + linkJSON("b") + `],"hasMore":true}`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":true,"code":"INTERNAL","message":"boom"}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))
	it := client.Links.Iter(context.Background(), awsysco.ListLinksInput{Limit: 2})

	var got []string
	for it.Next() {
		got = append(got, it.Link().ShortCode)
	}
	if len(got) != 2 {
		t.Fatalf("got = %v, want the 2 items from the first page before the error", got)
	}
	if it.Err() == nil {
		t.Fatal("Err() = nil, want the underlying 500 error")
	}
	if !awsysco.IsServerError(it.Err()) {
		t.Errorf("Err() = %v, want a 5xx server error", it.Err())
	}

	// A further Next() call must not panic or loop.
	if it.Next() {
		t.Error("Next() = true after a terminal error, want false")
	}
}
