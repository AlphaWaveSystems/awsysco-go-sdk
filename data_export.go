package awsysco

import (
	"context"
)

// DataExportResource provides access to the CSV data export API.
type DataExportResource struct {
	client *Client
}

// ExportLinks returns a CSV string containing all links for the authenticated user.
func (r *DataExportResource) ExportLinks(ctx context.Context) (string, error) {
	return r.client.doText(ctx, "GET", pathExportLinks, nil)
}

// ExportLinkStats returns a CSV string containing click stats for the given short path.
func (r *DataExportResource) ExportLinkStats(ctx context.Context, shortPath string) (string, error) {
	return r.client.doText(ctx, "GET", pathExportStats(shortPath), nil)
}
