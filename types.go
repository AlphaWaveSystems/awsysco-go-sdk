package awsysco

import (
	"encoding/json"
	"fmt"
	"time"
)

// Link represents a shortened URL.
// The API may return shortCode as either "shortCode" or "short" depending on
// the endpoint; both are handled transparently.
type Link struct {
	ID                string          `json:"id"`
	ShortURL          string          `json:"shortUrl"`
	ShortCode         string          `json:"shortCode"`
	Long              string          `json:"long"`
	Clicks            int             `json:"clicks"`
	Created           time.Time       `json:"created"`
	ExpiresAt         *time.Time      `json:"expiresAt"`
	MaxClicks         *int            `json:"maxClicks"`
	ExpireFallbackURL string          `json:"expireFallbackUrl,omitempty"`
	PasswordProtected bool            `json:"passwordProtected"`
	Namespace         string          `json:"namespace"`
	FullPath          string          `json:"fullPath"`
	GeoRestriction    *GeoRestriction `json:"geoRestriction,omitempty"`
	OgMeta            *OgMeta         `json:"ogMeta,omitempty"`
	IsCustom          bool            `json:"isCustom,omitempty"`
	IsDisabled        bool            `json:"isDisabled,omitempty"`
	DisabledReason    string          `json:"disabledReason,omitempty"`
}

// UnmarshalJSON handles the API inconsistency where "shortCode" may be
// returned as "short" on some endpoints.
func (l *Link) UnmarshalJSON(b []byte) error {
	// Use an alias to avoid infinite recursion. Deliberately does NOT
	// shadow Clicks here (unlike Created/ExpiresAt below) — a same-tagged
	// field declared directly on this anonymous struct takes priority over
	// the same tag promoted from the embedded *LinkAlias, which would
	// silently swallow the real "clicks" value into an unused field and
	// leave l.Clicks always 0. Let it decode straight through the
	// embedded alias instead.
	type LinkAlias Link
	aux := &struct {
		Short     string      `json:"short"`
		Created   interface{} `json:"created"`
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

// RoutingRule defines a geo-based routing rule for a link.
type RoutingRule struct {
	Country     string `json:"country"`
	RedirectURL string `json:"redirectUrl"`
}

// OgMeta defines OpenGraph metadata overrides for a link.
type OgMeta struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       string `json:"image,omitempty"`
}

// GeoRestriction defines geographic allow/block rules for a link.
type GeoRestriction struct {
	AllowedCountries []string `json:"allowedCountries,omitempty"`
	BlockedCountries []string `json:"blockedCountries,omitempty"`
}

// CreateLinkInput is the input for creating a link.
type CreateLinkInput struct {
	URL               string          `json:"url"`
	CustomSlug        string          `json:"customSlug,omitempty"`
	ExpiresAt         *time.Time      `json:"expiresAt,omitempty"`
	MaxClicks         *int            `json:"maxClicks,omitempty"`
	ExpireFallbackURL string          `json:"expireFallbackUrl,omitempty"`
	RoutingRules      []RoutingRule   `json:"routingRules,omitempty"`
	OgMeta            *OgMeta         `json:"ogMeta,omitempty"`
	GeoRestriction    *GeoRestriction `json:"geoRestriction,omitempty"`
	Password          string          `json:"password,omitempty"`
	PassAdClickIds    bool            `json:"passAdClickIds,omitempty"`
	FolderID          string          `json:"folderId,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
}

// UpdateLinkInput is the input for updating a link.
type UpdateLinkInput struct {
	URL               string          `json:"url,omitempty"`
	ExpiresAt         *time.Time      `json:"expiresAt,omitempty"`
	MaxClicks         *int            `json:"maxClicks,omitempty"`
	ExpireFallbackURL string          `json:"expireFallbackUrl,omitempty"`
	RoutingRules      []RoutingRule   `json:"routingRules,omitempty"`
	OgMeta            *OgMeta         `json:"ogMeta,omitempty"`
	GeoRestriction    *GeoRestriction `json:"geoRestriction,omitempty"`
	Password          string          `json:"password,omitempty"`
	PassAdClickIds    bool            `json:"passAdClickIds,omitempty"`
	FolderID          string          `json:"folderId,omitempty"`
	Tags              []string        `json:"tags,omitempty"`
}

// ListLinksInput is the input for listing links.
type ListLinksInput struct {
	Limit  int
	Offset int
}

// AggregateStats holds aggregated click breakdown data attached to LinkStats.
type AggregateStats struct {
	Countries map[string]int `json:"countries"`
	Devices   map[string]int `json:"devices"`
	Browsers  map[string]int `json:"browsers"`
	Referrers map[string]int `json:"referrers"`
}

// LinkStats holds analytics for a link.
type LinkStats struct {
	ShortCode      string          `json:"shortCode"`
	TotalClicks    int             `json:"totalClicks"`
	Clicks         []ClickEvent    `json:"clicks"`
	AggregateStats *AggregateStats `json:"aggregateStats,omitempty"`
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

// firestoreTimestamp handles both ISO string and Firestore
// {_seconds,_nanoseconds} / {seconds,nanoseconds} formats. Per ADR-017 it
// must never raise on an unknown/garbage shape — worst case the field ends
// up zero and decoding of the containing struct still succeeds.
type firestoreTimestamp struct {
	time.Time
}

func (f *firestoreTimestamp) UnmarshalJSON(b []byte) error {
	// Try plain string first.
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			f.Time = t
		}
		// Not a recognized date string — leave zero, never error.
		return nil
	}

	// Try a Firestore-shaped object: {_seconds,_nanoseconds} (underscore) or
	// {seconds,nanoseconds} (bare) — the platform has emitted both forms.
	// Decode field-by-field via a raw map instead of a fixed struct so a
	// garbage value on one key (e.g. nanoseconds:"q", seconds:[1]) doesn't
	// blow up decoding of the whole object — it's simply treated as absent.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		// Not an object either (or some other shape entirely) — never error.
		return nil
	}

	getInt := func(keys ...string) (int64, bool) {
		for _, k := range keys {
			v, ok := raw[k]
			if !ok {
				continue
			}
			var n int64
			if json.Unmarshal(v, &n) == nil {
				return n, true
			}
			// Present but not a valid int64 (wrong type, or a number too
			// large/small to fit — e.g. 1e300) — treat as absent, not fatal.
		}
		return 0, false
	}

	secs, ok := getInt("_seconds", "seconds")
	if !ok {
		// No usable seconds field at all — leave zero, never error.
		return nil
	}
	nanos, _ := getInt("_nanoseconds", "nanoseconds")
	f.Time = time.Unix(secs, nanos).UTC()
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

// UpdateFolderInput is the input for updating a folder.
type UpdateFolderInput struct {
	Name  string `json:"name,omitempty"`
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
	Total   int              `json:"total"`
	Results []BulkLinkResult `json:"results"`
}

// UnmarshalJSON reads Created/Failed/Total from the platform's real response
// envelope ({success, summary:{total,created,failed}, results}), falling back
// to top-level created/failed/total fields if summary is absent.
func (b *BulkCreateResponse) UnmarshalJSON(data []byte) error {
	type BulkCreateResponseAlias BulkCreateResponse
	aux := &struct {
		Summary *struct {
			Total   int `json:"total"`
			Created int `json:"created"`
			Failed  int `json:"failed"`
		} `json:"summary"`
		*BulkCreateResponseAlias
	}{
		BulkCreateResponseAlias: (*BulkCreateResponseAlias)(b),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if aux.Summary != nil {
		b.Created = aux.Summary.Created
		b.Failed = aux.Summary.Failed
		b.Total = aux.Summary.Total
	}
	return nil
}

// BulkLinkResult is the result of a single link in a bulk create.
type BulkLinkResult struct {
	Success  bool   `json:"success"`
	ShortURL string `json:"shortUrl"`
	Long     string `json:"long"`
	Error    string `json:"error"`
}

// QROptions configures QR code generation.
//
// Deprecated: use QROption functional options with QRResource.GetURL instead.
type QROptions struct {
	Size    int
	Color   string
	BGColor string
}

// IntOrUnlimited represents a usage limit value that the API returns either as
// a numeric quantity or as the string "unlimited". When the limit is unlimited,
// Unlimited is true and Value is 0.
type IntOrUnlimited struct {
	Value     int
	Unlimited bool
}

// UnmarshalJSON decodes a value that may be a JSON number or the string
// "unlimited".
func (u *IntOrUnlimited) UnmarshalJSON(b []byte) error {
	// Try string form first (e.g. "unlimited").
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		if s == "unlimited" {
			u.Unlimited = true
			u.Value = 0
			return nil
		}
		// Some other string — treat as unbounded/unknown rather than failing.
		u.Unlimited = true
		u.Value = 0
		return nil
	}
	// Otherwise decode as a plain integer.
	var n int
	if err := json.Unmarshal(b, &n); err != nil {
		return err
	}
	u.Value = n
	u.Unlimited = false
	return nil
}

// UsageLimits holds the per-tier usage limits for the current account.
// Fields that can be "unlimited" use IntOrUnlimited; the remainder are plain
// ints. CustomSlugs is a plain bool, not a count — the platform models it as
// a tier feature flag (functions/app/config/tierLimits.js), not a quota; see
// functions/app/routes/user.js and apiV1.js for the same field on /api/user/*
// and /api/v1/* responses.
type UsageLimits struct {
	LinksPerMonth        IntOrUnlimited `json:"linksPerMonth"`
	MonthlyLinks         IntOrUnlimited `json:"monthlyLinks"`
	DailyLinks           IntOrUnlimited `json:"dailyLinks"`
	MonthlyTrackedClicks IntOrUnlimited `json:"monthlyTrackedClicks"`
	QRCodes              IntOrUnlimited `json:"qrCodes"`
	Folders              IntOrUnlimited `json:"folders"`
	APICallsPerMonth     int            `json:"apiCallsPerMonth"`
	CustomSlugs          bool           `json:"customSlugs"`
}

// UsageOverage describes the account's metered-overage state.
type UsageOverage struct {
	Active               bool     `json:"active"`
	StartedAt            *string  `json:"startedAt"`
	ExpiresAt            *string  `json:"expiresAt"`
	HoursUntilDrop       *float64 `json:"hoursUntilDrop"`
	ClicksThisCycle      int      `json:"clicksThisCycle"`
	SpendingLimitCents   int      `json:"spendingLimitCents"`
	EstimatedChargeCents int      `json:"estimatedChargeCents"`
}

// UsageStats is the response from the /api/user/stats endpoint.
type UsageStats struct {
	TotalLinks             int          `json:"totalLinks"`
	TotalClicks            int          `json:"totalClicks"`
	LinksCreatedThisMonth  int          `json:"linksCreatedThisMonth"`
	QRCodesThisMonth       int          `json:"qrCodesThisMonth"`
	FolderCount            int          `json:"folderCount"`
	APICallsThisMonth      int          `json:"apiCallsThisMonth"`
	TrackedClicksThisMonth int          `json:"trackedClicksThisMonth"`
	Tier                   string       `json:"tier"`
	Limits                 UsageLimits  `json:"limits"`
	HasAPIKey              bool         `json:"hasApiKey"`
	APIKeyCreatedAt        *string      `json:"apiKeyCreatedAt"`
	UserPrefix             *string      `json:"userPrefix"`
	IsPremium              bool         `json:"isPremium"`
	Overage                UsageOverage `json:"overage"`
}

// Web2AppSession is the response from consuming a Web2App attribution token via
// GET /api/v1/web2app/{token}.
//
// Web2App sessions are single-use: a successful consume deletes the token
// server-side, so a second call with the same token returns 404. Tokens also
// expire 24 hours after creation, after which they are deleted and return 404.
type Web2AppSession struct {
	Success     bool                   `json:"success"`
	LinkID      string                 `json:"linkId"`
	UTMParams   map[string]string      `json:"utmParams"`
	RoutingRule map[string]interface{} `json:"routingRule"`
	Country     *string                `json:"country"`
	ClickedAt   *string                `json:"clickedAt"`
}

// ImportCounts holds the per-stage record counts for an import job.
type ImportCounts struct {
	Fetched     int `json:"fetched"`
	Transformed int `json:"transformed"`
	Written     int `json:"written"`
	Errored     int `json:"errored"`
}

// ImportJob represents a provider import job created via the imports API.
//
// Status is one of: pending, running, completed, partial, failed, cancelled.
type ImportJob struct {
	ID              string       `json:"id"`
	UserID          string       `json:"userId"`
	Provider        string       `json:"provider"`
	Status          string       `json:"status"`
	ScanOnly        bool         `json:"scanOnly"`
	TargetNamespace *string      `json:"targetNamespace"`
	ScopeFilter     *string      `json:"scopeFilter"`
	Counts          ImportCounts `json:"counts"`
	Errors          []string     `json:"errors"`
	CreatedAt       *string      `json:"createdAt"`
	UpdatedAt       *string      `json:"updatedAt"`
}

// ImportStartOptions is the input for starting a provider import.
type ImportStartOptions struct {
	Provider        string `json:"provider"`
	AccessToken     string `json:"accessToken"`
	TargetNamespace string `json:"targetNamespace,omitempty"`
	ScanOnly        bool   `json:"scanOnly,omitempty"`
}

// String implements fmt.Stringer, redacting AccessToken so a provider OAuth
// token never leaks through logging/debug output of an ImportStartOptions
// value (or pointer — the method set of *ImportStartOptions includes this
// value-receiver method too).
func (o ImportStartOptions) String() string {
	return fmt.Sprintf(
		"awsysco.ImportStartOptions{Provider: %q, AccessToken: %s, TargetNamespace: %q, ScanOnly: %v}",
		o.Provider, redactSecret(o.AccessToken), o.TargetNamespace, o.ScanOnly,
	)
}

// GoString implements fmt.GoStringer, redacting AccessToken.
func (o ImportStartOptions) GoString() string { return o.String() }

// ImportListOptions filters the imports List request.
type ImportListOptions struct {
	Limit int
}

// DayClicks holds the click count for a single calendar day.
type DayClicks struct {
	Date   string `json:"date"`
	Clicks int    `json:"clicks"`
}

// HourClicks holds the click count for a single hour bucket.
type HourClicks struct {
	Hour   int `json:"hour"`
	Clicks int `json:"clicks"`
}

// DeviceBreakdown holds device-category click counts (paid-tier field).
type DeviceBreakdown struct {
	Mobile  int `json:"mobile"`
	Desktop int `json:"desktop"`
	Tablet  int `json:"tablet"`
}

// UTMBreakdown holds UTM-parameter click breakdowns (paid-tier field).
type UTMBreakdown struct {
	Sources   map[string]int `json:"sources"`
	Mediums   map[string]int `json:"mediums"`
	Campaigns map[string]int `json:"campaigns"`
}

// UpgradeForMore describes paid-tier analytics fields gated behind an upgrade.
type UpgradeForMore struct {
	Available []string `json:"available"`
	Message   string   `json:"message"`
}

// AggregateAnalytics is the response from the aggregate stats endpoint
// (GET /api/v1/links/{shortPath}/stats/aggregate). Paid-tier breakdowns are
// pointers/omitempty and are nil for free-tier responses, which instead
// populate UpgradeForMore.
type AggregateAnalytics struct {
	ShortCode         string           `json:"shortCode"`
	FullPath          *string          `json:"fullPath,omitempty"`
	Period            string           `json:"period"`
	TotalClicks       int              `json:"totalClicks"`
	BotClicksExcluded int              `json:"botClicksExcluded"`
	UniqueVisitors    int              `json:"uniqueVisitors"`
	ClicksByDay       []DayClicks      `json:"clicksByDay"`
	CountryBreakdown  map[string]int   `json:"countryBreakdown"`
	TierLimit         int              `json:"tierLimit"`
	Tier              string           `json:"tier"`
	DeviceBreakdown   *DeviceBreakdown `json:"deviceBreakdown,omitempty"`
	ReferrerBreakdown map[string]int   `json:"referrerBreakdown,omitempty"`
	BrowserBreakdown  map[string]int   `json:"browserBreakdown,omitempty"`
	OSBreakdown       map[string]int   `json:"osBreakdown,omitempty"`
	SourceBreakdown   map[string]int   `json:"sourceBreakdown,omitempty"`
	HourBreakdown     []HourClicks     `json:"hourBreakdown,omitempty"`
	UTMBreakdown      *UTMBreakdown    `json:"utmBreakdown,omitempty"`
	UpgradeForMore    *UpgradeForMore  `json:"upgradeForMore,omitempty"`
}

// AggregateOptions filters the aggregate stats request. Period is one of
// "7d", "30d", or "90d".
type AggregateOptions struct {
	Period string
}

// WaitOptions configures ImportsResource.WaitForCompletion polling behaviour.
type WaitOptions struct {
	// PollInterval is the delay between status checks (default 2s).
	PollInterval time.Duration
	// Timeout is the maximum total wait before returning an error (default 120s).
	Timeout time.Duration
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

// TrialInfo describes an account's active trial, if any. Profile.Trial is nil
// when the account has no trial.
type TrialInfo struct {
	Active   bool    `json:"active"`
	Locked   bool    `json:"locked"`
	DaysLeft int     `json:"daysLeft"`
	Tier     string  `json:"tier"`
	EndsAt   *string `json:"endsAt"`
}

// Profile is the response from GET /api/user/profile.
type Profile struct {
	UID                   string                 `json:"uid"`
	Email                 string                 `json:"email"`
	EmailVerified         bool                   `json:"emailVerified"`
	IsPremium             bool                   `json:"isPremium"`
	UserPrefix            string                 `json:"userPrefix"`
	SubscriptionTier      string                 `json:"subscriptionTier"`
	SubscriptionSource    string                 `json:"subscriptionSource"`
	LinksCreatedThisMonth int                    `json:"linksCreatedThisMonth"`
	APICallsThisMonth     int                    `json:"apiCallsThisMonth"`
	HasAPIKey             bool                   `json:"hasApiKey"`
	Created               string                 `json:"created"`
	PreferredLanguage     string                 `json:"preferredLanguage"`
	Trial                 *TrialInfo             `json:"trial"`
	OfferDeclinedAt       *string                `json:"offerDeclinedAt"`
	Limits                map[string]interface{} `json:"limits"`
	Features              map[string]interface{} `json:"features"`
	FeatureFlags          map[string]interface{} `json:"featureFlags"`
	UtmTemplates          []UtmTemplate          `json:"utmTemplates"`
	FiscalData            map[string]interface{} `json:"fiscalData"`
}

// ProfileUpdateInput is the input for ProfileResource.Update. Only non-nil
// fields are sent in the request body.
type ProfileUpdateInput struct {
	DisplayName       *string `json:"displayName,omitempty"`
	PhotoURL          *string `json:"photoURL,omitempty"`
	PreferredLanguage *string `json:"preferredLanguage,omitempty"`
}
