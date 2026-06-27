package awsysco_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestImportsStart(t *testing.T) {
	const body = `{
		"id": "imp_123",
		"userId": "user_42",
		"provider": "bitly",
		"status": "pending",
		"scanOnly": false,
		"targetNamespace": "promo",
		"scopeFilter": null,
		"counts": {"fetched": 0, "transformed": 0, "written": 0, "errored": 0},
		"errors": [],
		"createdAt": "2026-06-27T10:00:00Z",
		"updatedAt": "2026-06-27T10:00:00Z"
	}`

	var (
		gotPath   string
		gotMethod string
		gotBody   map[string]interface{}
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	job, err := client.Imports.Start(context.Background(), awsysco.ImportStartOptions{
		Provider:        "bitly",
		AccessToken:     "secret_token",
		TargetNamespace: "promo",
		ScanOnly:        false,
	})
	if err != nil {
		t.Fatalf("Imports.Start failed: %v", err)
	}

	if gotMethod != "POST" {
		t.Errorf("method = %s, want POST", gotMethod)
	}
	if gotPath != "/api/v1/imports" {
		t.Errorf("path = %s, want /api/v1/imports", gotPath)
	}

	// Body must be snake_case.
	if gotBody["provider"] != "bitly" {
		t.Errorf("body provider = %v, want bitly", gotBody["provider"])
	}
	if gotBody["access_token"] != "secret_token" {
		t.Errorf("body access_token = %v, want secret_token", gotBody["access_token"])
	}
	if gotBody["target_namespace"] != "promo" {
		t.Errorf("body target_namespace = %v, want promo", gotBody["target_namespace"])
	}
	if _, hasCamel := gotBody["accessToken"]; hasCamel {
		t.Error("body should not contain camelCase accessToken")
	}

	if job.ID != "imp_123" {
		t.Errorf("ID = %q, want imp_123", job.ID)
	}
	if job.Status != "pending" {
		t.Errorf("Status = %q, want pending", job.Status)
	}
	if job.TargetNamespace == nil || *job.TargetNamespace != "promo" {
		t.Errorf("TargetNamespace = %v, want promo", job.TargetNamespace)
	}
}

func TestImportsStartOmitsOptionalFields(t *testing.T) {
	var gotBody map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"imp_1","provider":"bitly","status":"pending"}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	_, err := client.Imports.Start(context.Background(), awsysco.ImportStartOptions{
		Provider:    "bitly",
		AccessToken: "tok",
	})
	if err != nil {
		t.Fatalf("Imports.Start failed: %v", err)
	}

	if _, ok := gotBody["target_namespace"]; ok {
		t.Error("target_namespace should be omitted when empty")
	}
	if _, ok := gotBody["scan_only"]; ok {
		t.Error("scan_only should be omitted when false")
	}
}

func TestImportsGetStatus(t *testing.T) {
	const body = `{"id":"imp_123","provider":"bitly","status":"running","counts":{"fetched":10,"transformed":8,"written":5,"errored":1}}`

	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	job, err := client.Imports.GetStatus(context.Background(), "imp_123")
	if err != nil {
		t.Fatalf("Imports.GetStatus failed: %v", err)
	}
	if gotMethod != "GET" {
		t.Errorf("method = %s, want GET", gotMethod)
	}
	if gotPath != "/api/v1/imports/imp_123" {
		t.Errorf("path = %s, want /api/v1/imports/imp_123", gotPath)
	}
	if job.Status != "running" {
		t.Errorf("Status = %q, want running", job.Status)
	}
	if job.Counts.Fetched != 10 || job.Counts.Errored != 1 {
		t.Errorf("Counts = %+v, want fetched=10 errored=1", job.Counts)
	}
}

func TestImportsCancel(t *testing.T) {
	const body = `{"id":"imp_123","provider":"bitly","status":"cancelled"}`

	var gotPath, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	job, err := client.Imports.Cancel(context.Background(), "imp_123")
	if err != nil {
		t.Fatalf("Imports.Cancel failed: %v", err)
	}
	if gotMethod != "DELETE" {
		t.Errorf("method = %s, want DELETE", gotMethod)
	}
	if gotPath != "/api/v1/imports/imp_123" {
		t.Errorf("path = %s, want /api/v1/imports/imp_123", gotPath)
	}
	if job.Status != "cancelled" {
		t.Errorf("Status = %q, want cancelled", job.Status)
	}
}

func TestImportsList(t *testing.T) {
	const body = `{"jobs":[
		{"id":"imp_1","provider":"bitly","status":"completed"},
		{"id":"imp_2","provider":"rebrandly","status":"failed"}
	]}`

	var gotPath, gotMethod, gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	jobs, err := client.Imports.List(context.Background(), &awsysco.ImportListOptions{Limit: 25})
	if err != nil {
		t.Fatalf("Imports.List failed: %v", err)
	}
	if gotMethod != "GET" {
		t.Errorf("method = %s, want GET", gotMethod)
	}
	if gotPath != "/api/v1/imports" {
		t.Errorf("path = %s, want /api/v1/imports", gotPath)
	}
	if gotQuery != "limit=25" {
		t.Errorf("query = %q, want limit=25", gotQuery)
	}
	if len(jobs) != 2 {
		t.Fatalf("len(jobs) = %d, want 2", len(jobs))
	}
	if jobs[0].ID != "imp_1" || jobs[1].Status != "failed" {
		t.Errorf("unexpected jobs: %+v", jobs)
	}
}

func TestImportsListNoOptions(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jobs":[]}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	jobs, err := client.Imports.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("Imports.List failed: %v", err)
	}
	if gotQuery != "" {
		t.Errorf("query = %q, want empty", gotQuery)
	}
	if len(jobs) != 0 {
		t.Errorf("len(jobs) = %d, want 0", len(jobs))
	}
}

func TestImportsWaitForCompletionResolves(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls++
		n := calls
		mu.Unlock()
		status := "pending"
		if n >= 2 {
			status = "completed"
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"imp_123","provider":"bitly","status":"` + status + `"}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	job, err := client.Imports.WaitForCompletion(context.Background(), "imp_123", &awsysco.WaitOptions{
		PollInterval: 10 * time.Millisecond,
		Timeout:      2 * time.Second,
	})
	if err != nil {
		t.Fatalf("WaitForCompletion failed: %v", err)
	}
	if job.Status != "completed" {
		t.Errorf("Status = %q, want completed", job.Status)
	}
}

func TestImportsWaitForCompletionTimesOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"imp_123","provider":"bitly","status":"running"}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	_, err := client.Imports.WaitForCompletion(context.Background(), "imp_123", &awsysco.WaitOptions{
		PollInterval: 10 * time.Millisecond,
		Timeout:      40 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestImportsWaitForCompletionRespectsContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"imp_123","provider":"bitly","status":"running"}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	_, err := client.Imports.WaitForCompletion(ctx, "imp_123", &awsysco.WaitOptions{
		PollInterval: 10 * time.Millisecond,
		Timeout:      10 * time.Second,
	})
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}
