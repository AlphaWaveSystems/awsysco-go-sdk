# awsysco-go-sdk

[![Go Version](https://img.shields.io/badge/go-1.22%2B-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Official Go SDK for the [AWSYS.CO](https://awsys.co) URL Shortener API.

## Installation

```bash
go get github.com/AlphaWaveSystems/awsysco-go-sdk
```

The root module has **zero third-party dependencies** and requires Go 1.22+.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

func main() {
    client := awsysco.NewClient("awsys_your_api_key_here")

    link, err := client.Links.Create(context.Background(), awsysco.CreateLinkInput{
        URL: "https://example.com/very/long/url",
    })
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Short URL:", link.ShortURL)
}
```

## Configuration

```go
client := awsysco.NewClient("awsys_your_key",
    awsysco.WithBaseURL("https://staging.awsys.co"), // override base URL
    awsysco.WithTimeout(15 * time.Second),           // custom HTTP client timeout
    awsysco.WithHTTPClient(&http.Client{}),          // bring your own http.Client
    awsysco.WithMaxRetries(5),                       // retry attempts beyond the initial try (default 3)
)
```

### Environment variables

`NewClient` falls back to environment variables when the corresponding
argument/option is omitted:

| Variable          | Used when                                          |
| ------------------ | --------------------------------------------------- |
| `AWSYS_API_KEY`    | `apiKey` passed to `NewClient` is `""`               |
| `AWSYS_BASE_URL`   | `WithBaseURL` was not used (or left the base URL unset) |

```go
// Reads AWSYS_API_KEY and AWSYS_BASE_URL from the environment.
client := awsysco.NewClient("")
```

### Base URL validation

`WithBaseURL` (and the `AWSYS_BASE_URL` fallback) require a URL with an
`http://` or `https://` scheme. `NewClient` **never panics** on bad
configuration — an empty API key with no `AWSYS_API_KEY` fallback, or an
invalid/non-http(s) base URL, records a `*awsysco.ConfigurationError` on the
client. Every method call returns that error immediately, before any network
request is made:

```go
client := awsysco.NewClient("", awsysco.WithBaseURL("ftp://example.com"))

_, err := client.Links.Get(ctx, "abc123")
var cfgErr *awsysco.ConfigurationError
if errors.As(err, &cfgErr) {
    fmt.Println("misconfigured client:", cfgErr)
}
```

### Functional options

| Option                       | Effect                                                                 |
| ----------------------------- | ----------------------------------------------------------------------- |
| `WithBaseURL(u string)`       | Overrides the default base URL (`https://awsys.co`). Validated eagerly. |
| `WithHTTPClient(hc *http.Client)` | Supplies a custom `*http.Client` (proxies, custom transports, etc.). |
| `WithTimeout(d time.Duration)`    | Sets the HTTP client timeout. Composes safely with `WithHTTPClient` regardless of option order. |
| `WithMaxRetries(n int)`       | Maximum retry attempts beyond the first try. `0` disables retries. Negative values clamp to `0`. Default `3`. |

## API Reference

Every resource method takes a `context.Context` as its first argument and
returns `(<result>, error)` unless noted otherwise.

### Links

```go
// Create a link
link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
    URL:        "https://example.com",
    CustomSlug: "my-slug",          // optional
    MaxClicks:  &maxClicks,         // optional *int
    ExpiresAt:  &expiresAt,         // optional *time.Time
    Tags:       []string{"demo"},   // optional
    FolderID:   "folder_id",        // optional
})

// List links (single page)
resp, err := client.Links.List(ctx, awsysco.ListLinksInput{
    Limit:  20,
    Offset: 0,
})
// resp.Links []Link, resp.Total int, resp.HasMore bool

// Get a link by ID
link, err := client.Links.Get(ctx, "link_id")

// Update a link
maxClicks := 500
link, err := client.Links.Update(ctx, "link_id", awsysco.UpdateLinkInput{
    MaxClicks: &maxClicks,
})

// Delete a link
err := client.Links.Delete(ctx, "link_id")
```

#### Pagination with `Links.Iter`

`Links.Iter` wraps the offset-based `List` pagination in a `database/sql`-`Rows`-style
iterator, so you don't have to track offsets by hand:

```go
it := client.Links.Iter(ctx, awsysco.ListLinksInput{Limit: 50})
for it.Next() {
    link := it.Link()
    fmt.Println(link.ShortURL)
}
if err := it.Err(); err != nil {
    log.Fatal(err) // any transport/HTTP error surfaced mid-pagination
}
```

Notes on `LinkIterator`:

- `opts.Limit` is clamped to the platform maximum of 100 (and defaults to 100
  if unset or invalid).
- Iteration stops when the server reports no more pages, when a page comes
  back shorter than the requested limit (even if `hasMore` says otherwise),
  or when a page is empty.
- `Next()` returns `false` both at the end of iteration and on error — always
  check `Err()` after the loop to tell them apart.
- Errors mid-pagination don't discard already-buffered results: items already
  fetched in the current page are still yielded by `Next()` before it starts
  returning `false`.

### Analytics

```go
stats, err := client.Analytics.GetStats(ctx, "link_id", "7d")
// stats.ShortCode string
// stats.TotalClicks int
// stats.Clicks []ClickEvent — per-click breakdown (country, device, browser, OS, referrer)
```

#### Aggregate stats

`GetAggregateStats` returns pre-aggregated analytics for a link via
`GET /api/v1/links/{shortPath}/stats/aggregate`. The `Period` option is
`"7d"`, `"30d"`, or `"90d"`. Paid-tier breakdowns (`DeviceBreakdown`,
`UTMBreakdown`, `HourBreakdown`, `ReferrerBreakdown`, `BrowserBreakdown`,
`OSBreakdown`, `SourceBreakdown`) are pointers/maps that are **nil on the
free tier**, where the response instead populates `UpgradeForMore`.

```go
agg, err := client.Analytics.GetAggregateStats(ctx, "abc123", &awsysco.AggregateOptions{
    Period: "30d",
})
// agg.TotalClicks, agg.UniqueVisitors int
// agg.ClicksByDay []DayClicks, agg.CountryBreakdown map[string]int
// agg.Tier string, agg.TierLimit int

if agg.UpgradeForMore != nil {
    // free tier — paid breakdowns gated:
    fmt.Println(agg.UpgradeForMore.Message, agg.UpgradeForMore.Available)
} else {
    // paid tier — richer breakdowns available:
    fmt.Println("mobile clicks:", agg.DeviceBreakdown.Mobile)
    fmt.Println("utm sources:", agg.UTMBreakdown.Sources)
}
```

#### Recent clicks

`GetRecentClicks` returns recent click events across **all** of the
authenticated user's links via `GET /api/user/clicks/recent` (fixed in
v1.2.0 — it previously pointed at a nonexistent path). `since` is variadic
purely to keep the signature backward compatible; pass at most one value:

```go
clicks, err := client.Analytics.GetRecentClicks(ctx, 20)
// or, filtered to a time window:
clicks, err := client.Analytics.GetRecentClicks(ctx, 20, time.Now().Add(-1*time.Hour))
```

This endpoint is gated behind the "Live Globe" feature flag on some
accounts; when disabled it returns a 403 with `Code` `"FEATURE_DISABLED"` —
detect it with `awsysco.IsForbidden(err)` and inspect `AwsysError.Code`.

### QR Codes

QR code URL construction is a pure function — no HTTP request is made.

```go
// Default options (300px, black on white)
url := client.QR.GetURL("abc123")
// https://awsys.co/api/qr/abc123?bgColor=FFFFFF&color=000000&size=300

// Custom options
url := client.QR.GetURL("abc123",
    awsysco.WithSize(512),
    awsysco.WithColor("1D4ED8"),
    awsysco.WithBGColor("F0F9FF"),
)
```

Persisted QR settings (distinct from the URL-construction helper above) are
managed via `GetSettings`/`UpdateSettings`:

```go
settings, err := client.QR.GetSettings(ctx, "abc123")
settings.LogoURL = "https://example.com/logo.png"
updated, err := client.QR.UpdateSettings(ctx, "abc123", *settings)
```

### Folders

```go
// Create a folder
folder, err := client.Folders.Create(ctx, awsysco.CreateFolderInput{
    Name:  "My Folder",
    Color: "#3B82F6",
})

// List folders
resp, err := client.Folders.List(ctx)
// resp.Folders []Folder, resp.Limit int, resp.Used int

// Assign a link to a folder (linkID may be the Firestore doc ID or shortCode)
err := client.Folders.AssignLink(ctx, "link_id", "folder_id")

// Remove a link from its folder
err := client.Folders.RemoveLink(ctx, "link_id")

// Update a folder's name/color
folder, err := client.Folders.Update(ctx, "folder_id", awsysco.UpdateFolderInput{
    Name: "Renamed Folder",
})

// Delete a folder
err := client.Folders.Delete(ctx, "folder_id")
```

> **Platform routing quirk:** `Folders.List`/`Create`/`Delete` are mounted
> under `/api/v1/folders`, but `Folders.Update` only exists at the
> unversioned `/api/folders/{id}` twin — the `/api/v1` PATCH route 404s on
> the live platform. The SDK routes each method to the correct path
> internally; you don't need to think about this, but it explains why
> `Update`'s behavior can look inconsistent with the others if you're
> watching raw HTTP traffic.

### Tags

```go
// Add one or more tags in a single call (the platform's endpoint is
// batch-oriented: {"tags": [...]})
resp, err := client.Tags.Add(ctx, "link_id", "featured", "campaign-q3")
// resp.Tags []string — the link's full tag set after the add

// Remove a single tag
resp, err := client.Tags.Remove(ctx, "link_id", "featured")
```

### Custom Domains

```go
domains, err := client.CustomDomains.List(ctx)
result, err := client.CustomDomains.Add(ctx, "links.example.com")
status, err := client.CustomDomains.Verify(ctx, "links.example.com")
isDefault := true
updated, err := client.CustomDomains.Update(ctx, "links.example.com", awsysco.UpdateDomainInput{
    IsDefault:       &isDefault,
    DefaultRedirect: "https://example.com/404",
})
_, err = client.CustomDomains.Remove(ctx, "links.example.com")
avail, err := client.CustomDomains.Check(ctx, "links.example.com")
```

#### Deprecation: `CustomDomains.Activate`

```go
// Deprecated: Firebase-only, cannot be called with an API key (see ADR-006).
_, err := client.CustomDomains.Activate(ctx, "links.example.com")
```

`Activate` requires Firebase session authentication (`requireAuthStrict`) on
the platform side and can never succeed with an API-key-authenticated SDK
client. As of v1.2.0 it returns a `403 *awsysco.AwsysError` **immediately,
without making any network call** — it is kept only for source
compatibility and will be removed in a future major version. Activate
domains through the AWSYS.CO dashboard instead.

### Webhooks

```go
eventTypes, err := client.Webhooks.ListEventTypes(ctx)
webhooks, err := client.Webhooks.List(ctx)

webhook, err := client.Webhooks.Create(ctx, awsysco.CreateWebhookInput{
    URL:    "https://example.com/webhook",
    Events: []string{"link.created", "link.click"},
    Name:   "my-integration",
})

enabled := false
updated, err := client.Webhooks.Update(ctx, webhook.ID, awsysco.UpdateWebhookInput{
    Enabled: &enabled,
})

_, err = client.Webhooks.Test(ctx, webhook.ID, "link.created")
_, err = client.Webhooks.Delete(ctx, webhook.ID)
```

> **Platform routing quirk:** `List`/`Create`/`Delete`/`Test` are mounted
> under `/api/v1/webhooks`, but `Update` only exists at the unversioned
> `/api/webhooks/{id}` twin. As with folders, the SDK handles the routing
> internally.

### Saved Views

```go
views, err := client.SavedViews.List(ctx)

view, err := client.SavedViews.Create(ctx, awsysco.CreateSavedViewInput{
    Name:    "Active Q3 campaign links",
    Filters: awsysco.SavedViewFilters{Tag: "q3-campaign", Status: "active"},
})

updated, err := client.SavedViews.Update(ctx, view.ID, awsysco.UpdateSavedViewInput{
    Name: "Q3 campaign links (renamed)",
})

err = client.SavedViews.Delete(ctx, view.ID)
```

### UTM Templates

```go
resp, err := client.UtmTemplates.Create(ctx, awsysco.CreateUtmTemplateInput{
    Name:     "email-newsletter",
    Source:   "newsletter",
    Medium:   "email",
    Campaign: "q3-launch",
})
// resp.Template.ID, resp.Template.Name, ...

templates, err := client.UtmTemplates.List(ctx)

_, err = client.UtmTemplates.Delete(ctx, resp.Template.ID)
```

> **List is currently non-functional on the platform (ADR-020, retracting
> ADR-003):** verified live 2026-09-08 — `GET /api/v1/me` does not actually
> return a `utmTemplates` field, and there is no working
> `GET /api/user/utm-templates` route either (tracked upstream as platform
> issue #831). `UtmTemplates.List` always returns an **empty slice** until
> that ships — never silently treat this as an authoritative "you have no
> templates" answer, and note it now logs one `awsysco: warning: ...` line
> per call to make the limitation visible. Genuine transport/HTTP failures
> from the underlying `/api/v1/me` call still propagate normally.
>
> The wire field names are `source`/`medium`/`campaign`/`term`/`content`
> (not `utmSource`/`utmMedium`/`utmCampaign`/...) — verified live against
> `POST /api/user/utm-templates`.

### Data Export

```go
csv, err := client.DataExport.ExportLinks(ctx)          // all links, as CSV text
csv, err := client.DataExport.ExportLinkStats(ctx, "abc123") // click stats for one link, as CSV text
```

These return plain CSV strings (not JSON) — the SDK exposes them via a
non-JSON request path (`doText`) for exactly this reason.

### Namespace

```go
info, err := client.Namespace.Get(ctx)                 // current namespace status
check, err := client.Namespace.Check(ctx, "acme")      // availability check
info, err = client.Namespace.Claim(ctx, "acme")        // claim it
_, err = client.Namespace.Release(ctx)                 // release the current namespace
```

### Affiliate

```go
program, err := client.Affiliate.CreateProgram(ctx, awsysco.CreateAffiliateProgramInput{
    Name:           "Referral Program",
    CommissionType: "cpc",
    CommissionRate: 0.10,
    CpcRate:        0.05,
})

programs, err := client.Affiliate.ListPrograms(ctx)
program, err = client.Affiliate.GetProgram(ctx, program.ID)
program, err = client.Affiliate.UpdateProgram(ctx, program.ID, awsysco.CreateAffiliateProgramInput{
    Name: "Referral Program (updated)",
})
stats, err := client.Affiliate.GetProgramStats(ctx, program.ID, "30d")
partners, err := client.Affiliate.ListPartners(ctx, program.ID)
_, err = client.Affiliate.UpdatePartnerStatus(ctx, program.ID, "partner_id", "approved")

// Discoverable programs & partnerships (the "join as a partner" side)
discovered, err := client.Affiliate.Discover(ctx, 20)
_, err = client.Affiliate.Join(ctx, discovered[0].ID, awsysco.JoinProgramInput{})
partnerships, err := client.Affiliate.ListPartnerships(ctx)
stats, err = client.Affiliate.GetPartnershipStats(ctx, "partnership_id", "30d")
_, err = client.Affiliate.LeaveProgram(ctx, "partnership_id")

limits, err := client.Affiliate.GetLimits(ctx)
```

### Agentlink

```go
_, err := client.Agentlink.Subscribe(ctx, "person@example.com") // public, no auth required
stats, err := client.Agentlink.GetLinkStats(ctx, "abc123", 30)   // 30-day lookback
account, err := client.Agentlink.GetAccountStats(ctx, 7)         // 7-day lookback
```

### Bulk

```go
resp, err := client.Bulk.Create(ctx, awsysco.BulkCreateInput{
    URLs: []awsysco.BulkLinkInput{
        {URL: "https://example.com/page1"},
        {URL: "https://example.com/page2"},
        {URL: "https://example.com/page3", CustomSlug: "page3"},
    },
})
// resp.Created int
// resp.Failed int
// resp.Results []BulkLinkResult
```

### Me (Current User)

```go
me, err := client.Me.Get(ctx)
// me.Email string
// me.SubscriptionTier string
// me.IsPremium bool
// me.Features map[string]interface{}
// me.Limits map[string]interface{}
```

### Trust Score

```go
result, err := client.TrustScore.Scan(ctx, "abc123")
// result.Score *float64, result.Status *string, result.Threats []string
```

### Usage (Account Stats & Limits)

```go
stats, err := client.Usage.Get(ctx)
// stats.TotalLinks, stats.TotalClicks int
// stats.LinksCreatedThisMonth, stats.QRCodesThisMonth int
// stats.FolderCount, stats.APICallsThisMonth, stats.TrackedClicksThisMonth int
// stats.Tier string
// stats.HasAPIKey bool, stats.APIKeyCreatedAt *string
// stats.UserPrefix *string, stats.IsPremium bool
// stats.Overage — metered-overage state (active, clicks, estimated charge, ...)

// Tier limits. Fields that can be "unlimited" use IntOrUnlimited:
if stats.Limits.MonthlyLinks.Unlimited {
    fmt.Println("monthly links: unlimited")
} else {
    fmt.Println("monthly links:", stats.Limits.MonthlyLinks.Value)
}
// Plain int limits: stats.Limits.APICallsPerMonth
// stats.Limits.CustomSlugs bool — a tier feature flag, not a count
```

### Profile

```go
profile, err := client.Profile.Get(ctx)
// profile.Email, profile.SubscriptionTier, profile.PreferredLanguage string
// profile.Trial *TrialInfo (nil when the account has no active trial)
// profile.UtmTemplates []UtmTemplate

name := "Jane Doe"
lang := "es"
err = client.Profile.Update(ctx, awsysco.ProfileUpdateInput{
    DisplayName:       &name,
    PreferredLanguage: &lang, // validated client-side against en/es/fr-CA/de
})
```

`Profile.Update` only sends fields you set (all fields are pointers). If
`PreferredLanguage` is set to anything other than `"en"`, `"es"`, `"fr-CA"`,
or `"de"`, the SDK returns a `400 *awsysco.AwsysError` **before making any
network call** — detect it with `awsysco.IsValidationError(err)`.

### Web2App (Attribution Sessions)

Consume a deferred-deep-link attribution token. Sessions are **single-use** (a
successful call deletes the token server-side) and expire **24 hours** after
creation. A consumed or expired token returns 404.

```go
session, err := client.Web2App.ConsumeSession(ctx, token)
if err != nil {
    if awsysco.IsNotFound(err) {
        // token already consumed or expired
    }
    log.Fatal(err)
}
// session.LinkID string
// session.UTMParams map[string]string
// session.RoutingRule map[string]interface{} (may be nil)
// session.Country *string, session.ClickedAt *string
```

### Imports (Provider Migration)

Import links from another provider (e.g. Bitly, Rebrandly). `Start` accepts an
access token for the source provider; `ScanOnly` performs a dry run without
writing links. Job `Status` progresses through `pending` → `running` →
a terminal state (`completed`, `partial`, `failed`, or `cancelled`).

```go
// Kick off an import
job, err := client.Imports.Start(ctx, awsysco.ImportStartOptions{
    Provider:        "bitly",
    AccessToken:     "bitly_access_token",
    TargetNamespace: "promo", // optional
    ScanOnly:        false,   // optional dry-run
})

// Poll a single status
job, err = client.Imports.GetStatus(ctx, job.ID)
// job.Status string, job.Counts (Fetched/Transformed/Written/Errored int)
// job.Errors []string

// List recent jobs
jobs, err := client.Imports.List(ctx, &awsysco.ImportListOptions{Limit: 25})

// Cancel a running import
job, err = client.Imports.Cancel(ctx, job.ID)

// Block until the job reaches a terminal state (defaults: poll 2s, timeout 120s)
final, err := client.Imports.WaitForCompletion(ctx, job.ID, &awsysco.WaitOptions{
    PollInterval: 5 * time.Second,
    Timeout:      10 * time.Minute,
})

// Once a job is terminal, pull the redirect map it produced.
csv, err := client.Imports.GetRedirectMapCSV(ctx, job.ID)
raw, err := client.Imports.GetRedirectMapJSON(ctx, job.ID) // []byte — json.Unmarshal it yourself
```

## Pagination & Iteration

Two resources page over the platform's offset-based `List` endpoints:

- **Links**: use `client.Links.Iter(ctx, opts)` — see [Links](#links) above
  for the full contract (limit clamping, stop conditions, error handling).
- **Imports**: `client.Imports.List` returns a single page (no iterator is
  provided; the platform's import job list is expected to stay small).

For everything else (`Folders.List`, `SavedViews.List`, `UtmTemplates.List`,
`Webhooks.List`, etc.) the platform returns the full collection in one
response — no pagination parameters exist for those endpoints.

## Error Handling

All errors returned by the SDK are one of:

- `*awsysco.AwsysError` — a mapped HTTP error response (4xx/5xx).
- `*awsysco.RateLimitError` — embeds `AwsysError`; returned for HTTP 429.
- `*awsysco.ConfigurationError` — client misconfiguration (bad/missing API
  key, invalid base URL), returned before any network call.
- `*awsysco.NetworkError` — a transport-level failure (connection refused,
  DNS failure, etc.).
- `*awsysco.TimeoutError` — embeds `NetworkError`; a request or context
  deadline was exceeded. `errors.Is(err, context.DeadlineExceeded)` works
  through its `Unwrap` chain.

```go
link, err := client.Links.Get(ctx, "nonexistent_id")
if err != nil {
    switch {
    case awsysco.IsNotFound(err):
        fmt.Println("link not found")
    case awsysco.IsAuthError(err):
        fmt.Println("invalid or expired API key")
    case awsysco.IsForbidden(err):
        fmt.Println("insufficient permissions")
    case awsysco.IsValidationError(err):
        fmt.Println("invalid input:", err)
    case awsysco.IsConflict(err):
        fmt.Println("resource conflict (e.g. slug already taken)")
    case awsysco.IsRateLimitError(err):
        fmt.Println("rate limited")
    case awsysco.IsServerError(err):
        fmt.Println("platform 5xx error")
    default:
        fmt.Println("unexpected error:", err)
    }
}
```

### `Is*` predicates

| Function                          | True when                                    |
| ----------------------------------- | ----------------------------------------------- |
| `IsNotFound(err)`                  | HTTP 404                                        |
| `IsAuthError(err)`                 | HTTP 401                                        |
| `IsForbidden(err)`                 | HTTP 403                                        |
| `IsValidationError(err)`           | HTTP 400                                        |
| `IsConflict(err)`                  | HTTP 409                                        |
| `IsRateLimitError(err)`            | HTTP 429 (`*RateLimitError` or any `*AwsysError` with `Status == 429`) |
| `IsServerError(err)`               | HTTP 5xx                                        |

All predicates use `errors.As` internally, so they work correctly whether
`err` is a bare `*AwsysError`, a `*RateLimitError` (embeds `AwsysError`), or
those types wrapped by `fmt.Errorf("...: %w", err)`.

### Error type inspection

```go
var awsysErr *awsysco.AwsysError
if errors.As(err, &awsysErr) {
    fmt.Println("HTTP status:", awsysErr.Status)
    fmt.Println("API error code:", awsysErr.Code)
    fmt.Println("Raw response:", string(awsysErr.Raw))
}
```

```go
var rlErr *awsysco.RateLimitError
if errors.As(err, &rlErr) {
    fmt.Println("retry after:", rlErr.RetryAfter)
    if rlErr.ResetsAt != nil {
        fmt.Println("quota resets at:", rlErr.ResetsAt)
    }
}
```

### Tolerant error-body parsing

The platform has returned error bodies in more than one shape over its
lifetime. `parseErrorResponse` tolerates all of them without ever panicking
or surfacing a JSON-decode error to the caller:

- `{"error": true, "code": "...", "message": "..."}`
- `{"error": "<string message>", "code": "..."}` — the string *is* the message
- `{"error": true, "code": "..."}` — no message; the SDK synthesizes one from the code
- `{"success": false, "message": "...", "code": "..."}`
- A non-JSON body (e.g. an Express default HTML 404 page) or an empty body —
  the SDK falls back to the HTTP status text for `Message`, while `Raw`
  always retains the original response bytes for debugging.

## Retry & Rate Limiting

The SDK automatically retries certain failures with full-jitter exponential
backoff: `min(1s * 2^attempt, 30s)`, then a random duration in `[0, that]`.
`WithMaxRetries(n)` controls how many retries are attempted beyond the
initial try (default `3`; `0` disables retries).

Retry behavior depends on the failure and the HTTP method:

| Failure                                   | Retried?                                             |
| -------------------------------------------- | ------------------------------------------------------- |
| HTTP 429, quota-class code (`HOURLY_LIMIT_EXCEEDED`, `DAILY_LIMIT_EXCEEDED`, `MONTHLY_LIMIT_EXCEEDED`) | **Never** — the window can't reset any sooner, so the SDK fails fast. |
| HTTP 429, any other code                  | Yes, on every method. Honors the `Retry-After` header (integer seconds or an HTTP-date) when present; falls back to jittered backoff otherwise. |
| HTTP 502 / 503 / 504                      | Yes, but **only** for idempotent methods (`GET`, `PUT`, `DELETE`) — never for `POST`/`PATCH`. |
| Network error (connection refused/reset, DNS failure) | Yes, but only for idempotent methods — same rule as 5xx. |
| Timeout (`*TimeoutError`, i.e. context/request deadline exceeded) | Never — a timeout is assumed intentional/unrecoverable within this call. |
| Any other 4xx                             | Never.                                                 |

Context cancellation is honored during a backoff wait — cancelling `ctx` cuts
a pending sleep short instead of running out the full schedule.

```go
var rlErr *awsysco.RateLimitError
if errors.As(err, &rlErr) {
    fmt.Println("rate limited, retry after:", rlErr.RetryAfter)
}

if awsysco.IsRateLimitError(err) {
    // handle rate limit after retries exhausted
}

client := awsysco.NewClient(apiKey, awsysco.WithMaxRetries(0)) // disable retries entirely
```

## Deprecations

| Symbol                          | Status                                                                 |
| ---------------------------------- | ------------------------------------------------------------------------- |
| `CustomDomainsResource.Activate`  | Deprecated in v1.2.0. Firebase-only on the platform side; cannot succeed with API-key auth. Returns a `403 *AwsysError` immediately, no network call. Will be removed in a future major version. Use the AWSYS.CO dashboard to activate domains. |
| `QROptions`                       | Deprecated. Use the `QROption` functional options (`WithSize`, `WithColor`, `WithBGColor`) with `QRResource.GetURL` instead. |

## Supported Go Versions

CI tests against Go **1.22, 1.23, and 1.24** (see `.github/workflows/ci.yml`).
`go.mod` declares `go 1.22` as the minimum.

## Testing

This repository ships two independent test suites:

### Offline unit / contract suite (root module)

Runs entirely against `httptest` servers and in-process fixtures — **no
network access and no API key required**:

```bash
go build ./...
go test -race ./...
```

This includes:

- Standard unit tests per resource (e.g. `analytics_test.go`, `usage_test.go`, `web2app_test.go`).
- `contract_fixture_test.go` / `contract_errors_test.go` — replay fixtures from
  `testdata/sdk-contract.json` against the SDK's request/response handling,
  so drift from the platform's actual contract shows up as a local test
  failure (see also `.github/workflows/contract-drift.yml`, which fetches the
  live upstream contract weekly and on a platform-side dispatch).
- `retry_test.go` — backoff/jitter bounds, `Retry-After` handling, quota-class
  fast-fail, method-based retry eligibility, `WithMaxRetries`, context
  cancellation during backoff.
- `redaction_test.go` — proves the API key never leaks through `String()`,
  `GoString()`, or any `fmt` verb, and that configuration errors fail fast
  (< 100ms, no network call).
- `iterator_test.go` — `Links.Iter` pagination edge cases (empty pages, short
  pages, missing `hasMore`, limit clamping, mid-pagination errors).

### Live integration suite (`livetest/`, separate module)

A sibling Go module (its own `go.mod`, `replace`d onto this repo) that
exercises the SDK against a **real** AWSYS.CO environment. Requires a valid
`AWSYS_API_KEY`:

```bash
cd livetest
export AWSYS_API_KEY=awsys_your_test_key   # or put it in livetest/.env.test (gitignored)
go test ./... -v -timeout 60s
```

`livetest` tests skip themselves cleanly (via `t.Skip`) when `AWSYS_API_KEY`
is not set, so `go test ./...` at the repo root never depends on it — CI's
nightly live-integration job also skips gracefully when the secret isn't
configured (e.g. in forks).

## Development Setup

```bash
git clone https://github.com/AlphaWaveSystems/awsysco-go-sdk.git
cd awsysco-go-sdk

go build ./...
go vet ./...
gofmt -l .            # should print nothing
go test -race ./...

# Optional: golangci-lint (see .golangci.yml), matches CI
golangci-lint run

# Live integration suite (see Testing above)
cd livetest && go test ./... -v -timeout 60s
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Make your changes with tests
4. Run `go build ./...`, `go vet ./...`, `gofmt -l .`, and `go test -race ./...`
   (and `cd livetest && go test ./...` if you touched behavior the live suite
   covers)
5. Open a pull request — CI runs the full matrix (Go 1.22/1.23/1.24, lint,
   vet, build, test, `govulncheck`) on every push and PR

## License

MIT License — see [LICENSE](LICENSE) for details.

AWSYS.CO is a product of Alpha Wave Systems.
