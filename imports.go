package awsysco

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

// ImportsResource provides access to the provider-import API
// (POST/GET/DELETE /api/v1/imports).
type ImportsResource struct {
	client *Client
}

// terminal import statuses — a job in any of these states will not change again.
func isTerminalImportStatus(status string) bool {
	switch status {
	case "completed", "partial", "failed", "cancelled":
		return true
	default:
		return false
	}
}

// Start kicks off a new provider import via POST /api/v1/imports. The request
// body is sent as {provider, accessToken, targetNamespace?, scanOnly?}.
func (r *ImportsResource) Start(ctx context.Context, opts ImportStartOptions) (*ImportJob, error) {
	var job ImportJob
	if err := r.client.doRequest(ctx, "POST", pathImports, opts, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// GetStatus retrieves the current state of an import job via
// GET /api/v1/imports/{jobID}.
func (r *ImportsResource) GetStatus(ctx context.Context, jobID string) (*ImportJob, error) {
	var job ImportJob
	path := pathImport(jobID)
	if err := r.client.doRequest(ctx, "GET", path, nil, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// Cancel cancels a running import job via DELETE /api/v1/imports/{jobID} and
// returns the updated job.
func (r *ImportsResource) Cancel(ctx context.Context, jobID string) (*ImportJob, error) {
	var job ImportJob
	path := pathImport(jobID)
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// List returns recent import jobs via GET /api/v1/imports. The response is a
// {"jobs": [...]} wrapper; the slice is returned directly.
func (r *ImportsResource) List(ctx context.Context, opts *ImportListOptions) ([]ImportJob, error) {
	path := pathImports
	if opts != nil && opts.Limit > 0 {
		path += "?limit=" + strconv.Itoa(opts.Limit)
	}
	var resp struct {
		Jobs []ImportJob `json:"jobs"`
	}
	if err := r.client.doRequest(ctx, "GET", path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Jobs, nil
}

// GetRedirectMapCSV returns the completed import job's redirect map as a CSV
// string via GET /api/v1/imports/{jobID}/redirect-map.csv.
func (r *ImportsResource) GetRedirectMapCSV(ctx context.Context, jobID string) (string, error) {
	return r.client.doText(ctx, "GET", pathImportRedirectMapCSV(jobID), nil)
}

// GetRedirectMapJSON returns the completed import job's redirect map as raw
// JSON bytes via GET /api/v1/imports/{jobID}/redirect-map.json. Callers that
// want a typed result can json.Unmarshal the returned bytes themselves.
func (r *ImportsResource) GetRedirectMapJSON(ctx context.Context, jobID string) ([]byte, error) {
	body, err := r.client.doText(ctx, "GET", pathImportRedirectMapJSON(jobID), nil)
	if err != nil {
		return nil, err
	}
	return []byte(body), nil
}

// WaitForCompletion polls GetStatus until the job reaches a terminal status
// (completed, partial, failed, or cancelled), the context is cancelled, or the
// timeout elapses. The poll interval defaults to 2s and the timeout to 120s.
func (r *ImportsResource) WaitForCompletion(ctx context.Context, jobID string, opts *WaitOptions) (*ImportJob, error) {
	poll := 2 * time.Second
	timeout := 120 * time.Second
	if opts != nil {
		if opts.PollInterval > 0 {
			poll = opts.PollInterval
		}
		if opts.Timeout > 0 {
			timeout = opts.Timeout
		}
	}

	deadline := time.Now().Add(timeout)

	// Check once immediately so an already-terminal job returns without delay.
	job, err := r.GetStatus(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if isTerminalImportStatus(job.Status) {
		return job, nil
	}

	ticker := time.NewTicker(poll)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return nil, fmt.Errorf("awsysco: timed out waiting for import %s to complete after %s", jobID, timeout)
			}
			job, err := r.GetStatus(ctx, jobID)
			if err != nil {
				return nil, err
			}
			if isTerminalImportStatus(job.Status) {
				return job, nil
			}
		}
	}
}
