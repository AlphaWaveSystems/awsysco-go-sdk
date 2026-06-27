# awsysco-go-sdk

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Official Go SDK for the [AWSYS.CO](https://awsys.co) URL Shortener API.

## Installation

```bash
go get github.com/AlphaWaveSystems/awsysco-go-sdk
```

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
    awsysco.WithTimeout(15 * time.Second),           // custom timeout
    awsysco.WithHTTPClient(&http.Client{}),          // bring your own http.Client
)
```

## API Reference

### Links

```go
// Create a link
link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
    URL:        "https://example.com",
    CustomSlug: "my-slug",          // optional
    MaxClicks:  &maxClicks,         // optional *int
    ExpiresAt:  &expiresAt,         // optional *time.Time
})

// List links
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

### Analytics

```go
stats, err := client.Analytics.GetStats(ctx, "link_id")
// stats.ShortCode string
// stats.TotalClicks int
// stats.Clicks []ClickEvent — per-click breakdown (country, device, browser, OS, referrer)
```

#### Aggregate stats

`GetAggregateStats` returns pre-aggregated analytics for a link. The `period`
option is `"7d"`, `"30d"`, or `"90d"`. Paid-tier breakdowns
(`DeviceBreakdown`, `UTMBreakdown`, `HourBreakdown`, `ReferrerBreakdown`,
`BrowserBreakdown`, `OSBreakdown`, `SourceBreakdown`) are pointers/maps that are
**nil on the free tier**, where the response instead populates `UpgradeForMore`.

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

// Assign a link to a folder
err := client.Folders.AssignLink(ctx, "link_id", "folder_id")

// Remove a link from its folder
err := client.Folders.RemoveLink(ctx, "link_id")

// Delete a folder
err := client.Folders.Delete(ctx, "folder_id")
```

### Bulk Create

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

### Me (Current User)

```go
me, err := client.Me.Get(ctx)
// me.Email string
// me.SubscriptionTier string
// me.IsPremium bool
// me.Features map[string]interface{}
// me.Limits map[string]interface{}
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
// Plain int limits: stats.Limits.APICallsPerMonth, stats.Limits.CustomSlugs
```

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
```

## Error Handling

All errors returned by the SDK are either `*awsysco.AwsysError` or `*awsysco.RateLimitError` (which embeds `AwsysError`).

```go
link, err := client.Links.Get(ctx, "nonexistent_id")
if err != nil {
    if awsysco.IsNotFound(err) {
        fmt.Println("link not found")
    } else if awsysco.IsAuthError(err) {
        fmt.Println("invalid or expired API key")
    } else if awsysco.IsForbidden(err) {
        fmt.Println("insufficient permissions")
    } else if awsysco.IsValidationError(err) {
        fmt.Println("invalid input:", err)
    } else if awsysco.IsConflict(err) {
        fmt.Println("resource conflict (e.g. slug already taken)")
    } else {
        fmt.Println("unexpected error:", err)
    }
}
```

### Error type inspection

```go
var awsysErr *awsysco.AwsysError
if errors.As(err, &awsysErr) {
    fmt.Println("HTTP status:", awsysErr.Status)
    fmt.Println("API error code:", awsysErr.Code)
    fmt.Println("Raw response:", string(awsysErr.Raw))
}
```

## Rate Limiting

The SDK automatically retries on HTTP 429 with exponential backoff (1s, 2s, 4s, max 3 attempts). The `Retry-After` header is respected when present.

```go
var rlErr *awsysco.RateLimitError
if errors.As(err, &rlErr) {
    fmt.Println("rate limited, retry after:", rlErr.RetryAfter)
}
```

You can also check with the convenience function:

```go
if awsysco.IsRateLimitError(err) {
    // handle rate limit after retries exhausted
}
```

## Development Setup

```bash
git clone https://github.com/AlphaWaveSystems/awsysco-go-sdk.git
cd awsysco-go-sdk

# Copy env template and fill in your staging API key
cp .env.example .env.test
# Edit .env.test with your AWSYS_API_KEY

# Install dependencies
go mod tidy

# Build
go build ./...

# Run tests (requires .env.test with a valid API key)
go test ./... -v -timeout 60s
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/your-feature`)
3. Make your changes with tests
4. Run `go build ./...` and `go test ./...`
5. Open a pull request

## License

MIT License — see [LICENSE](LICENSE) for details.

AWSYS.CO is a product of Alpha Wave Systems.
