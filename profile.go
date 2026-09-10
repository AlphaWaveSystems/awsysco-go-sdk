package awsysco

import "context"

// ProfileResource provides access to the account profile API.
type ProfileResource struct {
	client *Client
}

// preferredLanguages is the set of locale codes the platform accepts for
// ProfileUpdateInput.PreferredLanguage.
var preferredLanguages = map[string]bool{
	"en":    true,
	"es":    true,
	"fr-CA": true,
	"de":    true,
}

// Get returns the authenticated user's profile via GET /api/user/profile.
func (r *ProfileResource) Get(ctx context.Context) (*Profile, error) {
	var profile Profile
	if err := r.client.doRequest(ctx, "GET", pathUserProfile, nil, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// Update updates the authenticated user's profile via PATCH
// /api/user/profile. Only non-nil fields on input are sent. If
// input.PreferredLanguage is set, it is validated client-side against the
// platform's accepted locale codes ("en", "es", "fr-CA", "de") before any
// network call is made; an unsupported value returns a 400 *AwsysError
// immediately (use IsValidationError to detect it).
func (r *ProfileResource) Update(ctx context.Context, input ProfileUpdateInput) error {
	if input.PreferredLanguage != nil && !preferredLanguages[*input.PreferredLanguage] {
		return &AwsysError{
			Message: "preferredLanguage must be one of: en, es, fr-CA, de",
			Status:  400,
		}
	}
	return r.client.doRequest(ctx, "PATCH", pathUserProfile, input, nil)
}
