package awsysco

import (
	"context"
	"fmt"
	"net/url"
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
// body is sent as snake_case ({provider, access_token, target_namespace?,
// scan_only?}).
func (r *ImportsResource) Start(ctx context.Context, opts ImportStartOptions) (*ImportJob, error) {
	var job ImportJob
	if err := r.client.doRequest(ctx, "POST", "/api/v1/imports", opts, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// GetStatus retrieves the current state of an import job via
// GET /api/v1/imports/{jobID}.
func (r *ImportsResource) GetStatus(ctx context.Context, jobID string) (*ImportJob, error) {
	var job ImportJob
	path := fmt.Sprintf("/api/v1/imports/%s", url.PathEscape(jobID))
	if err := r.client.doRequest(ctx, "GET", path, nil, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// Cancel cancels a running import job via DELETE /api/v1/imports/{jobID} and
// returns the updated job.
func (r *ImportsResource) Cancel(ctx context.Context, jobID string) (*ImportJob, error) {
	var job ImportJob
	path := fmt.Sprintf("/api/v1/imports/%s", url.PathEscape(jobID))
	if err := r.client.doRequest(ctx, "DELETE", path, nil, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

// List returns recent import jobs via GET /api/v1/imports. The response is a
// {"jobs": [...]} wrapper; the slice is returned directly.
func (r *ImportsResource) List(ctx context.Context, opts *ImportListOptions) ([]ImportJob, error) {
	path := "/api/v1/imports"
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
