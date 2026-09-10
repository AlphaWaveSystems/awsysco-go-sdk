package awsysco_test

import (
	"context"
	"os"
	"testing"

	"github.com/joho/godotenv"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func newTestClient(t *testing.T) *awsysco.Client {
	t.Helper()
	_ = godotenv.Load(".env.test")
	key := os.Getenv("AWSYS_API_KEY")
	if key == "" {
		t.Skip("AWSYS_API_KEY not set")
	}
	baseURL := os.Getenv("AWSYS_BASE_URL")
	if baseURL == "" {
		baseURL = "https://staging.awsys.co"
	}
	client := awsysco.NewClient(key, awsysco.WithBaseURL(baseURL))

	// Probe reachability. Skip if:
	//   - staging gate requires a __session cookie (STAGING_AUTH_REQUIRED / 401)
	//   - account is not verified (EMAIL_NOT_VERIFIED / 403)
	//   - any other auth/forbidden error
	//   - rate limit exceeded (run again later)
	// This makes `go test ./...` a no-op when credentials are not fully configured.
	_, err := client.Me.Get(context.Background())
	if err != nil {
		if awsysco.IsAuthError(err) || awsysco.IsForbidden(err) || awsysco.IsRateLimitError(err) {
			t.Skipf("skipping: server unreachable, auth blocked, or rate-limited: %v", err)
		}
	}

	return client
}

// skipSetupError calls t.Skip if err represents a blocking environment condition
// (staging gate, unverified email, tier gate, rate limit) so that setup failures
// don't cause false FAIL results in CI without proper credentials.
func skipSetupError(t *testing.T, op string, err error) {
	t.Helper()
	if err == nil {
		return
	}
	if awsysco.IsAuthError(err) || awsysco.IsForbidden(err) || awsysco.IsRateLimitError(err) {
		t.Skipf("skipping: %s returned environment-specific error: %v", op, err)
	}
	t.Fatalf("%s failed: %v", op, err)
}
