package awsysco_test

import (
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
	return awsysco.NewClient(key, awsysco.WithBaseURL(baseURL))
}
