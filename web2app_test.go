package awsysco_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func TestWeb2AppConsumeSession(t *testing.T) {
	const token = "0123456789abcdef0123456789abcdef"
	const body = `{
		"success": true,
		"linkId": "abc123",
		"utmParams": {
			"utm_source": "newsletter",
			"utm_medium": "email"
		},
		"routingRule": {
			"country": "US",
			"redirectUrl": "https://example.com/us"
		},
		"country": "US",
		"clickedAt": "2026-06-27T10:00:00Z"
	}`

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	session, err := client.Web2App.ConsumeSession(context.Background(), token)
	if err != nil {
		t.Fatalf("Web2App.ConsumeSession failed: %v", err)
	}

	if gotPath != "/api/v1/web2app/"+token {
		t.Fatalf("expected path /api/v1/web2app/%s, got %s", token, gotPath)
	}

	if !session.Success {
		t.Error("Success = false, want true")
	}
	if session.LinkID != "abc123" {
		t.Errorf("LinkID = %q, want abc123", session.LinkID)
	}
	if session.UTMParams["utm_source"] != "newsletter" {
		t.Errorf("UTMParams[utm_source] = %q, want newsletter", session.UTMParams["utm_source"])
	}
	if session.UTMParams["utm_medium"] != "email" {
		t.Errorf("UTMParams[utm_medium] = %q, want email", session.UTMParams["utm_medium"])
	}
	if session.RoutingRule["country"] != "US" {
		t.Errorf("RoutingRule[country] = %v, want US", session.RoutingRule["country"])
	}
	if session.Country == nil || *session.Country != "US" {
		t.Errorf("Country = %v, want US", session.Country)
	}
	if session.ClickedAt == nil || *session.ClickedAt != "2026-06-27T10:00:00Z" {
		t.Errorf("ClickedAt = %v, want 2026-06-27T10:00:00Z", session.ClickedAt)
	}
}

func TestWeb2AppConsumeSessionNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":true,"code":"NOT_FOUND","message":"Token not found or already consumed"}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	_, err := client.Web2App.ConsumeSession(context.Background(), "0123456789abcdef0123456789abcdef")
	if err == nil {
		t.Fatal("expected error for consumed/expired token, got nil")
	}
	if !awsysco.IsNotFound(err) {
		t.Errorf("expected IsNotFound to be true, got %v", err)
	}
}
