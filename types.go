package awsysco

import (
	"encoding/json"
	"time"
)

// Link represents a shortened URL.
// The API may return shortCode as either "shortCode" or "short" depending on
// the endpoint; both are handled transparently.
type Link struct {
	ID                string     `json:"id"`
	ShortURL          string     `json:"shortUrl"`
	ShortCode         string     `json:"shortCode"`
	Long              string     `json:"long"`
	Clicks            int        `json:"clicks"`
	Created           time.Time  `json:"created"`
	ExpiresAt         *time.Time `json:"expiresAt"`
	MaxClicks         *int       `json:"maxClicks"`
	PasswordProtected bool       `json:"passwordProtected"`
	Namespace         string     `json:"namespace"`
	FullPath          string     `json:"fullPath"`
}

// UnmarshalJSON handles the API inconsistency where "shortCode" may be
// returned as "short" on some endpoints.
func (l *Link) UnmarshalJSON(b []byte) error {
	// Use an alias to avoid infinite recursion.
	type LinkAlias Link
	aux := &struct {
		Short    string      `json:"short"`
		Clicks   interface{} `json:"clicks"`
		Created  interface{} `json:"created"`
		ExpiresAt interface{} `json:"expiresAt"`
		*LinkAlias
	}{
		LinkAlias: (*LinkAlias)(l),
	}

	if err := json.Unmarshal(b, aux); err != nil {
		return err
	}

	// Prefer shortCode; fall back to short.
	if l.ShortCode == "" && aux.Short != "" {
		l.ShortCode = aux.Short
	}

	// Handle created as either time.Time or string.
	if aux.Created != nil {
		switch v := aux.Created.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				l.Created = t
			}
		}
	}

	// Handle expiresAt as either time.Time or string.
	if aux.ExpiresAt != nil {
		switch v := aux.ExpiresAt.(type) {
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				l.ExpiresAt = &t
			}
		}
	}

	return nil
}

// listLinksRaw is used for JSON decoding before normalization.
type listLinksRaw struct {
	Links      []Link `json:"links"`
	Total      int    `json:"total"`
	HasMore    bool   `json:"hasMore"`
	Pagination *struct {
		HasMore bool `json:"hasMore"`
	} `json:"pagination"`
}

// ListLinksResponse is the response from listing links.
type ListLinksResponse struct {
	Links   []Link `json:"links"`
	Total   int    `json:"total"`
	HasMore bool   `json:"hasMore"`
}

// UnmarshalJSON handles both API response shapes for the list endpoint.
// The production API nests hasMore under "pagination"; staging puts "total" at top level.
func (r *ListLinksResponse) UnmarshalJSON(b []byte) error {
	var raw listLinksRaw
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	r.Links = raw.Links
	r.Total = raw.Total
	r.HasMore = raw.HasMore
	if raw.Pagination != nil {
		r.HasMore = raw.Pagination.HasMore
	}
	return nil
}

// CreateLinkInput is the input for creating a link.
type CreateLinkInput struct {
	URL        string     `json:"url"`
	CustomSlug string     `json:"customSlug,omitempty"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	MaxClicks  *int       `json:"maxClicks,omitempty"`
}

// UpdateLinkInput is the input for updating a link.
type UpdateLinkInput struct {
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	MaxClicks *int       `json:"maxClicks,omitempty"`
}

// ListLinksInput is the input for listing links.
type ListLinksInput struct {
	Limit  int
	Offset int
}

// LinkStats holds analytics for a link.
type LinkStats struct {
	ShortCode   string       `json:"shortCode"`
	TotalClicks int          `json:"totalClicks"`
	Clicks      []ClickEvent `json:"clicks"`
}

// ClickEvent represents a single click on a link.
type ClickEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Country   string    `json:"country"`
	Device    string    `json:"device"`
	Browser   string    `json:"browser"`
	OS        string    `json:"os"`
	Referrer  string    `json:"referrer"`
}

// firestoreTimestamp handles both ISO string and Firestore {_seconds, _nanoseconds} formats.
type firestoreTimestamp struct {
	time.Time
}

func (f *firestoreTimestamp) UnmarshalJSON(b []byte) error {
	// Try plain string first.
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		t, err := time.Parse(time.RFC3339, s)
		if err == nil {
			f.Time = t
			return nil
		}
		// Not a recognized date string — leave zero.
		return nil
	}
	// Try Firestore object {_seconds: N, _nanoseconds: N}.
	var obj struct {
		Seconds     int64 `json:"_seconds"`
		Nanoseconds int64 `json:"_nanoseconds"`
	}
	if err := json.Unmarshal(b, &obj); err == nil && obj.Seconds != 0 {
		f.Time = time.Unix(obj.Seconds, obj.Nanoseconds).UTC()
		return nil
	}
	return nil
}

// Folder represents a link folder.
type Folder struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	LinkCount int       `json:"linkCount"`
	CreatedAt time.Time `json:"createdAt"`
	Created   time.Time `json:"created"`
}

// UnmarshalJSON handles both ISO string and Firestore timestamp formats for Created/CreatedAt.
func (f *Folder) UnmarshalJSON(b []byte) error {
	type FolderAlias Folder
	aux := &struct {
		Created   firestoreTimestamp `json:"created"`
		CreatedAt firestoreTimestamp `json:"createdAt"`
		*FolderAlias
	}{
		FolderAlias: (*FolderAlias)(f),
	}
	if err := json.Unmarshal(b, aux); err != nil {
		return err
	}
	if !aux.Created.IsZero() {
		f.Created = aux.Created.Time
		if f.CreatedAt.IsZero() {
			f.CreatedAt = aux.Created.Time
		}
	}
	if !aux.CreatedAt.IsZero() {
		f.CreatedAt = aux.CreatedAt.Time
	}
	return nil
}

// ListFoldersResponse is the response from listing folders.
type ListFoldersResponse struct {
	Folders []Folder `json:"folders"`
	Limit   int      `json:"limit"`
	Used    int      `json:"used"`
}

// CreateFolderInput is the input for creating a folder.
type CreateFolderInput struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

// BulkCreateInput is the input for bulk link creation.
type BulkCreateInput struct {
	URLs []BulkLinkInput `json:"urls"`
}

// BulkLinkInput is a single link entry in a bulk create request.
type BulkLinkInput struct {
	URL        string     `json:"url"`
	CustomSlug string     `json:"customSlug,omitempty"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	MaxClicks  *int       `json:"maxClicks,omitempty"`
}

// BulkCreateResponse is the response from a bulk create operation.
type BulkCreateResponse struct {
	Created int              `json:"created"`
	Failed  int              `json:"failed"`
	Results []BulkLinkResult `json:"results"`
}

// BulkLinkResult is the result of a single link in a bulk create.
type BulkLinkResult struct {
	Success  bool   `json:"success"`
	ShortURL string `json:"shortUrl"`
	Long     string `json:"long"`
	Error    string `json:"error"`
}

// QROptions configures QR code generation.
// Deprecated: use QROption functional options with QRResource.GetURL instead.
type QROptions struct {
	Size    int
	Color   string
	BGColor string
}

// MeResponse is the response from the /api/v1/me endpoint.
type MeResponse struct {
	UID              string                 `json:"uid"`
	Email            string                 `json:"email"`
	SubscriptionTier string                 `json:"subscriptionTier"`
	UserPrefix       string                 `json:"userPrefix"`
	IsPremium        bool                   `json:"isPremium"`
	Features         map[string]interface{} `json:"features"`
	Limits           map[string]interface{} `json:"limits"`
}
