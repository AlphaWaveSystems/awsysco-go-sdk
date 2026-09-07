package awsysco_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

const secretKey = "awsys_supersecretkeyvalue1234"

func TestRedactionNeverLeaksKey(t *testing.T) {
	client := awsysco.NewClient(secretKey, awsysco.WithBaseURL("https://example.com"))

	forms := map[string]string{
		"%v":       fmt.Sprintf("%v", client),
		"%+v":      fmt.Sprintf("%+v", client),
		"%#v":      fmt.Sprintf("%#v", client),
		"String()": client.String(),
	}

	for label, out := range forms {
		if strings.Contains(out, "supersecretkeyvalue1234") {
			t.Errorf("%s output leaked the raw key: %q", label, out)
		}
		if !strings.Contains(out, "1234") && !strings.Contains(out, "awsys_") {
			t.Errorf("%s output = %q, want a masked form showing awsys_ prefix or last 4 chars", label, out)
		}
	}
}

func TestConfigurationErrorNoKeyFast(t *testing.T) {
	orig, hadOrig := os.LookupEnv("AWSYS_API_KEY")
	os.Unsetenv("AWSYS_API_KEY")
	defer func() {
		if hadOrig {
			os.Setenv("AWSYS_API_KEY", orig)
		}
	}()

	client := awsysco.NewClient("")

	start := time.Now()
	_, err := client.Links.Get(context.Background(), "x")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var cfgErr *awsysco.ConfigurationError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error = %T, want *ConfigurationError", err)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("elapsed = %v, want well under 100ms (should fail before any network call)", elapsed)
	}
}

func TestConfigurationErrorInvalidBaseURLFast(t *testing.T) {
	client := awsysco.NewClient("valid_key", awsysco.WithBaseURL("ftp://example.com"))

	start := time.Now()
	_, err := client.Links.Get(context.Background(), "x")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	var cfgErr *awsysco.ConfigurationError
	if !errors.As(err, &cfgErr) {
		t.Fatalf("error = %T, want *ConfigurationError", err)
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("elapsed = %v, want well under 100ms (should fail before any network call)", elapsed)
	}
}

func TestEnvFallbackAPIKeyAndBaseURL(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	origKey, hadKey := os.LookupEnv("AWSYS_API_KEY")
	origURL, hadURL := os.LookupEnv("AWSYS_BASE_URL")
	os.Setenv("AWSYS_API_KEY", "awsys_env_fallback_key")
	os.Setenv("AWSYS_BASE_URL", srv.URL)
	defer func() {
		if hadKey {
			os.Setenv("AWSYS_API_KEY", origKey)
		} else {
			os.Unsetenv("AWSYS_API_KEY")
		}
		if hadURL {
			os.Setenv("AWSYS_BASE_URL", origURL)
		} else {
			os.Unsetenv("AWSYS_BASE_URL")
		}
	}()

	client := awsysco.NewClient("")

	_, err := client.Links.Get(context.Background(), "x")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if gotAuth != "Bearer awsys_env_fallback_key" {
		t.Errorf("Authorization = %q, want Bearer awsys_env_fallback_key (AWSYS_API_KEY fallback)", gotAuth)
	}
	if gotPath != "/api/v1/links/x" {
		t.Errorf("path = %q — request did not land on the AWSYS_BASE_URL-configured server", gotPath)
	}
}
