# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [1.2.0] - 2026-09-06

Contract-parity release: brings the SDK's request paths, request/response
field names, and error/retry behavior in line with the actual
`awsys-shortener` platform contract, adds an offline contract-fixture test
suite to keep it that way, and rounds out resource coverage.

### Added

- `ProfileResource` (`client.Profile`) — `Get`/`Update` against
  `GET`/`PATCH /api/user/profile`, with client-side validation of
  `PreferredLanguage` against the platform's accepted locale codes
  (`en`, `es`, `fr-CA`, `de`) before any network call.
- `ImportsResource.GetRedirectMapCSV` / `GetRedirectMapJSON` — retrieve a
  completed import job's redirect map via
  `GET /api/v1/imports/{jobID}/redirect-map.{csv,json}`.
- `ImportsResource.WaitForCompletion` — polls `GetStatus` until an import job
  reaches a terminal state (`completed`, `partial`, `failed`, `cancelled`),
  the context is cancelled, or a timeout elapses (configurable via
  `WaitOptions`, default 2s poll / 120s timeout).
- `AnalyticsResource.GetAggregateStats` — pre-aggregated analytics via
  `GET /api/v1/links/{shortPath}/stats/aggregate`, including paid-tier
  breakdowns (`DeviceBreakdown`, `UTMBreakdown`, `HourBreakdown`,
  `ReferrerBreakdown`, `BrowserBreakdown`, `OSBreakdown`, `SourceBreakdown`)
  and the free-tier `UpgradeForMore` gating signal.
- `LinksResource.Iter` / `LinkIterator` — transparent pagination helper over
  `Links.List`'s offset-based paging, with limit clamping to the platform
  maximum of 100 and defensive stop conditions (short page, missing
  `hasMore`, empty page).
- New error types: `ConfigurationError` (bad/missing API key or base URL,
  surfaced before any network call), `NetworkError` (transport-level
  failures), and `TimeoutError` (a `NetworkError` specialization for
  request/context deadline exceeded, unwrapping to `context.DeadlineExceeded`).
- `IsServerError(err)` predicate for HTTP 5xx errors.
- `RateLimitError.ResetsAt` — parsed from the platform's `resetsAt` field on
  quota-class 429s, when present and RFC3339-formatted.
- A full retry engine: full-jitter exponential backoff
  (`min(1s*2^attempt, 30s)`, then random in `[0, that]`), `Retry-After`
  support (integer seconds or HTTP-date), method-aware retry eligibility
  (idempotent methods retry on 5xx/network errors, quota-class 429s never
  retry, timeouts never retry), and `WithMaxRetries(n)` to configure it
  (default 3, `0` disables retries). Backoff waits honor context
  cancellation.
- `NewClient` / `WithBaseURL` environment-variable fallback:
  `AWSYS_API_KEY` and `AWSYS_BASE_URL` are read when the corresponding
  argument/option is omitted, and base URLs (explicit or from the
  environment) are validated to require an `http://`/`https://` scheme.
- `Client.String()` / `GoString()` and `clientConfig.String()` / `GoString()`
  — redact the API key (masked to a `awsys_...<last4>` form) in every
  `fmt` verb (`%v`, `%+v`, `%#v`) and explicit `.String()` calls, so the key
  can never leak through logging or debug printing of the client.
- Offline contract-fixture test suite: `contract_fixture_test.go`,
  `contract_errors_test.go`, `retry_test.go`, `redaction_test.go`,
  `iterator_test.go`, `internal_test.go`, backed by
  `testdata/sdk-contract.json`. Runs with `go test -race ./...` in under 2
  seconds, no network access or API key required.
- `livetest/` — the live integration test suite was split out into its own
  Go module (own `go.mod`, `replace`d onto the root module) so the root
  module has zero third-party dependencies and `go test ./...` at the repo
  root never requires network access or a real API key.
- CI/release tooling: `.github/workflows/ci.yml` (build/vet/lint/test/
  govulncheck across Go 1.22/1.23/1.24), `.github/workflows/release.yml`
  (tag-triggered GitHub Release with notes extracted from this file), and
  `.github/workflows/contract-drift.yml` (weekly + dispatch-triggered check
  of the vendored contract fixture against the live upstream contract,
  auto-filing a GitHub issue on drift), plus `.golangci.yml`.

### Fixed

- `Analytics.GetRecentClicks` — corrected the request path from the
  nonexistent `/api/user/recent-clicks` to the platform's actual
  `GET /api/user/clicks/recent`, and updated the response envelope to match
  (`{clicks, count}`).
- `Folders.Update` — corrected to use the unversioned
  `PATCH /api/folders/{id}` route; the `/api/v1/folders/{id}` PATCH route
  404s live on the platform (List/Create/Delete remain on `/api/v1/folders`).
- `Webhooks.List` / `Create` / `Delete` / `Test` — corrected from the
  unversioned `/api/webhooks` paths to the platform's actual
  `/api/v1/webhooks` paths. `Update` deliberately remains on the unversioned
  `/api/webhooks/{id}` twin, matching the live platform.
- `UtmTemplate` / `CreateUtmTemplateInput` field names — corrected from
  `source`/`medium`/`campaign`/`term`/`content` to the platform's actual
  wire names `utmSource`/`utmMedium`/`utmCampaign`/`utmTerm`/`utmContent`.
- `Tags.Add` — the platform's endpoint takes a batch (`{"tags": [...]}`),
  not one tag per call; `Add` is now variadic (`Add(ctx, shortPath, tags...)`)
  so existing single-tag call sites keep compiling while also supporting
  multi-tag batches.
- `RateLimitError` now implements `Unwrap() error` (returning the embedded
  `AwsysError`), so `errors.As(err, &awsysErr)` works uniformly whether `err`
  is a `*RateLimitError` or a plain `*AwsysError`.
- `sdkVersion` (and the `User-Agent` header it feeds) bumped from the
  placeholder `0.1.0` to track the actual release version.

### Changed

- `UpdateDomainInput` gained a `DefaultRedirect` field (previously
  unrepresentable — the platform accepts it but the SDK had no field for it).
- `CreateAffiliateProgramInput` gained a `CommissionRate` field, and
  `CommissionType` is now optional (`omitempty`) rather than required.
- All internal endpoint paths were centralized into `paths.go` as
  unexported helper functions/constants, replacing ad hoc `fmt.Sprintf` path
  construction scattered across resource files. This is an internal
  refactor with no change to the public API.
- `WithTimeout` now composes safely with `WithHTTPClient` regardless of
  option order — the timeout is applied once, after all options run, to
  whatever `*http.Client` is in effect at that point.
- `go.mod` bumped to `go 1.22`; the root module now has **zero** third-party
  dependencies (previously depended on `github.com/joho/godotenv`, now moved
  to `livetest/go.mod` alongside the live-integration tests that use it).

### Deprecated

- `CustomDomainsResource.Activate` — the platform route requires Firebase
  session authentication (`requireAuthStrict`) and cannot be called with an
  API key (see ADR-006). It now returns a `403 *AwsysError` immediately,
  without making any network call, and will be removed in a future major
  version. Use the AWSYS.CO dashboard to activate domains instead.

## [1.1.0]

- Added `.token-publish` to `.gitignore` to prevent accidental publish-token
  exposure.
- Added `ExpireFallbackURL` to the link create/update input structs and the
  `Link` response struct.

## [1.0.0]

- Full SDK parity pass: 10 new resources added (analytics, folders, bulk, QR,
  tags, trust score, custom domains, agentlink, affiliate, and more), enhanced
  existing resources, example program, and initial test coverage.
- Added a CODEOWNERS file requiring review on all PRs.

## [0.2.0]

- Initial public Go SDK implementation: `Client`, `LinksResource`
  (create/list/get/update/delete), basic error handling, and the first
  functional options (`WithBaseURL`, `WithHTTPClient`, `WithTimeout`).
