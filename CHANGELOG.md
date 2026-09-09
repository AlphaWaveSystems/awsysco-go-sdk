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
  auto-filing a GitHub issue on drift), plus `.golangci.yml`. CI now also
  compiles/vets the `livetest/` module (no secrets required — build/vet
  only, not `go test`) so it can't silently bit-rot, and pins `govulncheck`
  to `v1.1.4` instead of `@latest`. `release.yml` now hard-fails before
  building/testing if the pushed tag doesn't exactly match
  `v<sdkVersion>` (read from `http.go`, not hardcoded).
- Vendored contract fixture (`testdata/sdk-contract.json`) bumped to
  `v1.0.9` (79 capabilities, 27 errors, 15 behaviors — up from v1.0.4's 21
  errors/8 behaviors), adding coverage for 6 new error scenarios
  (`err_422_validation`, `err_429_resets_at_only`,
  `err_429_retry_after_oversized`, `err_503_retry_after_oversized`,
  `err_2xx_malformed_json`, `err_user_cancel`) and 7 new cross-cutting
  behaviors (`user_agent`, `unknown_fields_preserved`, `timestamp_variants`,
  `timestamp_never_raises`, `iterator_links_limit_zero`, `redaction_str`,
  `config_warnings`, `links_list_has_more_from_pagination`,
  `body_read_within_timeout`), all covered by new/extended tests
  (`contract_errors_test.go`, `behaviors_test.go`, `retry_test.go`).
- `AwsysError.RetryAfter` — the parsed `Retry-After` response header,
  populated for any status (not only 429) that sends one.
- `IsConfigurationError`, `IsNetworkError`, `IsTimeoutError`, `IsSDKError`
  predicates, alongside the existing `Is*` family.
- `NewClient`/`WithBaseURL` now log one `awsysco: warning: ...` line (via
  the standard `log` package — no new dependency) when the API key doesn't
  start with `awsys_`, or the base URL doesn't use `https://`. Neither
  condition blocks configuration; both are informational.
- `CreateWebhookInput`/`UpdateWebhookInput`/`ImportStartOptions` now
  implement `String()`/`GoString()`, redacting `Secret`/`AccessToken` (as
  `[REDACTED]`) so they never leak through `%v`/`%+v`/`%s`/`%#v` formatting
  or accidental `Println`/logging of an input value.

### Fixed

- `AffiliateProgram.CookieDays` — corrected the wire field name from
  `cookieDays` to the platform's actual `cookieDurationDays` (the old tag
  decoded silently to `0` on every real response). Also gained `MerchantID`,
  `MaxPartners`, `PartnerCount`, `IsPublic`, `CreatedAt`, and `UpdatedAt`
  (the latter two normalizing the platform's Firestore
  `{_seconds,_nanoseconds}` shape to an ISO-8601 string, same pattern as
  `SavedView`). The same type is shared between the owner's full view and
  `Discover`'s public-summary subset; fields Discover omits (status,
  merchantId, etc.) simply decode as zero values.
- `TrustScoreResult.Score`/`Status` — corrected the wire field names from
  `score`/`status` to the platform's actual `trustScore`/`trustStatus`
  (`GET /api/link-scan/*`); the old tags decoded silently to `nil` on every
  real response. Also gained `Source` and `CreatedAt`, which the platform
  actually sends.
- `NamespaceInfo` — gained `CanClaimCustomDomain`, `CanClaimSubdomain`, and
  `NamespaceData`, which `GET /api/user/namespace` actually returns and the
  struct was previously missing entirely.
- `AggregateAnalytics` — gained `BotClicksExcluded`, present on the real
  `GET /api/v1/links/*/stats/aggregate` response but previously missing.
- `Link` — gained `GeoRestriction`, `OgMeta`, `IsCustom`, `IsDisabled`, and
  `DisabledReason`, all real fields the platform can return that the struct
  didn't expose.
- `Analytics.GetRecentClicks` — corrected the request path from the
  nonexistent `/api/user/recent-clicks` to the platform's actual
  `GET /api/user/clicks/recent`, and updated the response envelope to match
  (`{clicks, count}`). Also gained an optional `since ...time.Time`
  parameter (variadic, at most the first value used) to add `since`
  filtering support without breaking the existing two-argument call sites —
  a backward-compatible signature widening under ADR-014, not a breaking
  change.
- `Folders.Update` — corrected to use the unversioned
  `PATCH /api/folders/{id}` route; the `/api/v1/folders/{id}` PATCH route
  404s live on the platform (List/Create/Delete remain on `/api/v1/folders`).
- `Webhooks.List` / `Create` / `Delete` / `Test` — corrected from the
  unversioned `/api/webhooks` paths to the platform's actual
  `/api/v1/webhooks` paths. `Update` deliberately remains on the unversioned
  `/api/webhooks/{id}` twin, matching the live platform.
- `UtmTemplate` / `CreateUtmTemplateInput` field names — **re-corrected**
  back to the platform's actual wire names `source`/`medium`/`campaign`/
  `term`/`content` (not `utmSource`/`utmMedium`/`utmCampaign`/`utmTerm`/
  `utmContent` as an earlier pass in this same unreleased version assumed).
  Verified live against `POST /api/user/utm-templates`
  (`functions/app/routes/user.js:355`).
- `UtmTemplates.List` now reads the real `GET /api/user/utm-templates`
  route (`{"templates": [...]}`), added by platform PR #833 (ADR-021,
  superseding ADR-020/ADR-003). An earlier pass in this same unreleased
  version briefly had `List` always return an empty slice with a warning
  log because no working list route existed yet on the platform; that
  route now exists and `List` returns real data.
- `UsageLimits.CustomSlugs` — corrected from `int` to `bool`. It's a tier
  feature flag on the platform (`functions/app/config/tierLimits.js`,
  `routes/user.js:475`, `routes/apiV1.js:91`), not a count; decoding a real
  `usage` response with the previous `int` field would have failed outright.
- `Link.UnmarshalJSON` — fixed a shadowed `clicks` decode field that
  silently discarded the real click count on every `Link` decode, leaving
  `Link.Clicks` always `0` regardless of the API response. Caught by the new
  `unknown_fields_preserved` behavior test.
- `firestoreTimestamp` — now also recognizes the bare `{seconds,
  nanoseconds}` shape (previously only the underscore-prefixed
  `{_seconds,_nanoseconds}`), and no longer conflates a genuinely-present
  `seconds: 0` (a legitimate epoch timestamp) with an absent field; a
  garbage value on one sub-field (e.g. `nanoseconds: "q"`) no longer blocks
  parsing a valid `seconds` value alongside it. Still never errors on an
  unrecognized shape, per ADR-017.
- Body-read failures (`io.ReadAll(resp.Body)`, as opposed to `httpClient.Do`
  itself) are now classified through the same transport-error path as
  connection-level failures, so a server that sends headers/flushes and
  then stalls the body correctly surfaces as `*TimeoutError`, not a bare
  wrapped `encoding/json`-adjacent error the caller can't type-switch on.
- A caller-cancelled context (`context.Canceled`) no longer risks being
  reported as `*TimeoutError` — `wrapTransportError` now checks
  `context.Canceled` before `context.DeadlineExceeded`/`net.Error.Timeout()`
  and passes it through as a plain `*NetworkError` instead, so
  `errors.Is(err, context.Canceled)` holds and `errors.As(err, &timeoutErr)`
  correctly returns `false`.
- A 2xx response with a body that isn't valid JSON (e.g. an HTML
  interstitial) now surfaces as a typed `*AwsysError` (`Status` set to the
  actual 2xx status code) instead of a raw, unhandled `encoding/json`
  unmarshal error.
- `IsValidationError` now also returns `true` for HTTP 422 (in addition to
  400) — the platform uses 422 for some validation failures
  (`VALIDATION_FAILED`) and 400 for others; both are the same conceptual
  class.
- 429 responses with a `resetsAt` body field but no recognized quota `code`
  are now correctly treated as quota-class and never retried (previously
  only a recognized `code` triggered this).
- A server-supplied `Retry-After` above the 30s backoff cap (429, or a
  retryable 5xx that now also carries `AwsysError.RetryAfter`) now fails
  the request immediately instead of sleeping for the full duration.
- `Webhook.Enabled` is now `*bool` (was `bool`), so a response that omits
  the field decodes as `nil` ("unknown") instead of `false`
  ("disabled") — previously an active webhook with an omitted `enabled`
  field would have misreported as disabled.
- `LinksResource.List` now clamps `Limit` to the platform maximum of 100
  when it exceeds that, matching `Iter`'s existing clamping (a `Limit` of
  `0` or negative still omits the query parameter entirely, unchanged).
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
