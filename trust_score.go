package awsysco

import (
	"context"
)

// TrustScoreResource provides access to the link trust-score scanning API.
type TrustScoreResource struct {
	client *Client
}

// TrustScoreResult holds the safety scan result for a link.
//
// Score/Status use the real wire field names trustScore/trustStatus (a
// prior version of this struct used score/status, which silently decoded
// to nil against the actual platform response — see GET /api/link-scan/*,
// links.js:550+). Long is kept for backward compatibility (ADR-014) but is
// not present on the real response and will always decode empty; Source
// and CreatedAt reflect fields the real response actually sends.
type TrustScoreResult struct {
	Short     string   `json:"short"`
	Long      string   `json:"long"`
	Score     *float64 `json:"trustScore"`
	Status    *string  `json:"trustStatus"`
	Threats   []string `json:"threats"`
	ScannedAt *string  `json:"scannedAt"`
	Source    string   `json:"source,omitempty"`
	CreatedAt *int64   `json:"createdAt,omitempty"`
}

// Scan retrieves the trust-score scan result for the given short path.
func (r *TrustScoreResource) Scan(ctx context.Context, shortPath string) (*TrustScoreResult, error) {
	var result TrustScoreResult
	path := pathLinkScan(shortPath)
	if err := r.client.doRequest(ctx, "GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
