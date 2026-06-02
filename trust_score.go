package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// TrustScoreResource provides access to the link trust-score scanning API.
type TrustScoreResource struct {
	client *Client
}

// TrustScoreResult holds the safety scan result for a link.
type TrustScoreResult struct {
	Short     string   `json:"short"`
	Long      string   `json:"long"`
	Score     *float64 `json:"score"`
	Status    *string  `json:"status"`
	Threats   []string `json:"threats"`
	ScannedAt *string  `json:"scannedAt"`
}

// Scan retrieves the trust-score scan result for the given short path.
func (r *TrustScoreResource) Scan(ctx context.Context, shortPath string) (*TrustScoreResult, error) {
	var result TrustScoreResult
	path := fmt.Sprintf("/api/link-scan/%s", url.PathEscape(shortPath))
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
