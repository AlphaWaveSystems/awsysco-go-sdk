<!-- HARNESS:START
     version=0.32.0
     schema=1
     updated=2026-07-18T02:25:54Z
     DO NOT EDIT — regenerate with: harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk
-->

# Architecture — awsysco-go-sdk

> Auto-generated from constitution scan on 2026-07-18T02:25:54Z.
> Reflects the state of the repo at install time — update manually as the project evolves,
> or re-run `harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk` to refresh from the latest scan.

---

## Project identity

| Field | Value |
|---|---|
| Name | awsysco-go-sdk |
| Path | `/Users/patrickbertsch/dev/awsysco-go-sdk` |
| Repository | https://github.com/AlphaWaveSystems/awsysco-go-sdk.git |
| Stack | go |
| Language(s) | Go |
| Runtime | go1.26.4 |
| Package manager | go mod |
| Zeus owner | `hephaestus` |

---

## Project overview


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


---

## Stack overview

Go module. Binary built with `go build`, tested with `go test ./...`.

### Key entry points


### Build and test commands

| Action | Command |
|---|---|
| Install deps | `go mod download` |
| Build | `go build ./...` |
| Test | `go test ./...` |
| Lint | `(not detected — configure manually)` |
| Dev server | `go run ./cmd/...` |
| Deploy (staging) | `bash scripts/deploy-staging.sh` |
| Deploy (production) | `bash scripts/deploy-server.sh` |



---

## Directory structure

```
├── AGENTS.md
├── CLAUDE.md
├── README.md
├── SECURITY.md
├── affiliate.go
├── affiliate_test.go
├── agentlink.go
├── agentlink_test.go
├── analytics.go
├── analytics_test.go
├── awsysco/
├── awsysco.go
├── awsysco_test.go
├── bulk.go
├── bulk_test.go
├── custom_domains.go
├── custom_domains_test.go
├── data_export.go
├── data_export_test.go
├── errors.go
├── examples/
├── folders.go
├── folders_test.go
├── go.mod
├── go.sum
├── harness/
├── http.go
├── links.go
├── links_test.go
├── me.go
├── me_test.go
├── namespace.go
├── namespace_test.go
├── qr.go
├── qr_test.go
├── reports/
├── saved_views.go
├── saved_views_test.go
├── tags.go
├── tags_test.go
├── trust_score.go
├── trust_score_test.go
├── types.go
├── utm_templates.go
├── utm_templates_test.go
├── webhooks.go
├── webhooks_test.go
```

---

## Dependencies

**Runtime dependencies (1):**

- `github.com/joho/godotenv` v1.5.1



**Dev dependencies:**


---

## Environment variables

Variables the project reads at runtime. Do not commit values — use the harness vault.

| Variable | Required | Purpose |
|---|---|---|

| `AWSYS_API_KEY` | yes | (see .env.example) |

| `AWSYS_BASE_URL` | yes | (see .env.example) |



---

## External services



*(none detected)*


---

## Constitution context

Rules extracted from `CLAUDE.md` at install time:

<!-- HARNESS:START
     version=0.31.0
     schema=1
     agent=awsysco-go-sdk
     updated=2026-07-04T02:31:42Z
     DO NOT EDIT THIS BLOCK — regenerate with: harness-ctl update /Users/patrickbertsch/dev/awsysco-go-sdk
-->

# Harness — Active Constraints

**This file is the entry point for every task in this project — always start here.**

**Agent:** `awsysco-go-sdk` · trust: `worker` · model: `coder`
**Budget:** 40 steps · 80000 tokens · $3.00 per session
**Privacy:** local_preferred — local models preferred; cloud only on low confidence
**Memory namespace:** `awsysco-go-sdk-worker`


## Must escalate (blocks until human approves)

*(truncated — see CLAUDE.md for full rules)*

*(Full rules in `CLAUDE.md` — this is a harness-generated summary only)*



---

## Notes from previous version

---

<!-- Add architecture decisions, diagrams, and notes below.
     The harness block above is managed automatically — everything below is yours. -->



<!-- HARNESS:END -->

---

<!-- Add architecture decisions, diagrams, and notes below.
     The harness block above is managed automatically — everything below is yours. -->
