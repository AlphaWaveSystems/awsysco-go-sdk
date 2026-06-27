package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// Web2AppResource provides access to the Web2App attribution API.
type Web2AppResource struct {
	client *Client
}

// ConsumeSession retrieves and consumes a Web2App attribution session for the
// given token via GET /api/v1/web2app/{token}.
//
// Sessions are single-use: a successful call deletes the token server-side, so a
// subsequent call with the same token returns a 404 (use IsNotFound to detect).
// Tokens also expire 24 hours after creation, after which they return 404.
// An invalid token format returns a 400 (use IsValidationError to detect).
func (r *Web2AppResource) ConsumeSession(ctx context.Context, token string) (*Web2AppSession, error) {
	var session Web2AppSession
	path := fmt.Sprintf("/api/v1/web2app/%s", url.PathEscape(token))
	if err := r.client.doRequest(ctx, "GET", path, nil, &session); err != nil {
		return nil, err
	}
	return &session, nil
}
