# Security Review — awsysco-go-sdk (v1.2.0 contract-parity effort)

This is a point-in-time engineering security review of the Go SDK client
code at the root of this repository (not `livetest/`, not `examples/`),
performed as part of the v1.2.0 contract-parity effort. It is distinct from
`SECURITY.md`, which is the project's ongoing vulnerability-disclosure and
secret-hygiene policy — that document is unchanged by this review.

Findings are organized by severity. It is expected and fine for most
severity buckets to be empty in a small, well-scoped HTTP client SDK.

## Critical

None found.

## High

None found.

## Medium

None found.

## Low

### L-1: Standard-library CVEs flagged by `govulncheck` (toolchain-version dependent, not an SDK code issue)

`govulncheck ./...` was installed and run for real in this environment
(`go install golang.org/x/vuln/cmd/govulncheck@latest`, then
`govulncheck ./...` from the repo root). Verbatim output:

```
=== Symbol Results ===

Vulnerability #1: GO-2026-6218
    Avoid quadratic complexity in resolvePath in net/url
  More info: https://pkg.go.dev/vuln/GO-2026-6218
  Standard library
    Found in: net/url@go1.26.4
    Fixed in: net/url@go1.26.6
    Example traces found:
      #1: http.go:211:34: awsysco.Client.doRequestOnce calls http.Client.Do, which eventually calls url.URL.Parse

Vulnerability #2: GO-2026-6090
    Limit handshake messages we are willing to accept post-handshake in
    crypto/tls
  More info: https://pkg.go.dev/vuln/GO-2026-6090
  Standard library
    Found in: crypto/tls@go1.26.4
    Fixed in: crypto/tls@go1.26.6
    Example traces found:
      #1: http.go:211:34: awsysco.Client.doRequestOnce calls http.Client.Do, which eventually calls tls.Conn.HandshakeContext
      #2: http.go:217:24: awsysco.Client.doRequestOnce calls io.ReadAll, which eventually calls tls.Conn.Read
      #3: examples/integration/main.go:59:12: integration.main calls fmt.Printf, which eventually calls tls.Conn.Write
      #4: http.go:211:34: awsysco.Client.doRequestOnce calls http.Client.Do, which eventually calls tls.Dialer.DialContext

Vulnerability #3: GO-2026-5972
    Enforce maximum recursion depth in encoding/asn1
  More info: https://pkg.go.dev/vuln/GO-2026-5972
  Standard library
    Found in: encoding/asn1@go1.26.4
    Fixed in: encoding/asn1@go1.26.6
    Example traces found:
      #1: examples/integration/main.go:59:12: integration.main calls fmt.Printf, which eventually calls asn1.Unmarshal

Vulnerability #4: GO-2026-5856
    Invoking Encrypted Client Hello privacy leak in crypto/tls
  More info: https://pkg.go.dev/vuln/GO-2026-5856
  Standard library
    Found in: crypto/tls@go1.26.4
    Fixed in: crypto/tls@go1.26.5
    Example traces found:
      #1: http.go:211:34: awsysco.Client.doRequestOnce calls http.Client.Do, which eventually calls tls.Conn.HandshakeContext
      #2: http.go:217:24: awsysco.Client.doRequestOnce calls io.ReadAll, which eventually calls tls.Conn.Read
      #3: examples/integration/main.go:59:12: integration.main calls fmt.Printf, which eventually calls tls.Conn.Write
      #4: http.go:211:34: awsysco.Client.doRequestOnce calls http.Client.Do, which eventually calls tls.Dialer.DialContext

Vulnerability #5: GO-2026-5026
    Invoking failure to reject ASCII-only Punycode-encoded labels in
    golang.org/x/net/idna
  More info: https://pkg.go.dev/vuln/GO-2026-5026
  Standard library
    Found in: net/http@go1.26.4
    Fixed in: net/http@go1.26.6
    Example traces found:
      #1: http.go:211:34: awsysco.Client.doRequestOnce calls http.Client.Do

Your code is affected by 5 vulnerabilities from the Go standard library.
This scan also found 3 vulnerabilities in packages you import and 2
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
```

**Assessment:** all five reachable findings are in the **Go standard
library** as shipped by the locally installed toolchain (`go1.26.4`), not in
this module's own code or its (zero) third-party dependencies. Every one is
already fixed in a later `go1.26.x` patch release (`go1.26.5` or
`go1.26.6`). This is a toolchain-currency issue, not an SDK design or coding
flaw — resolved by building/releasing with a current Go patch release.

**Recommendation:** CI's `ci.yml` matrix pins Go **minor** versions
(`1.22`, `1.23`, `1.24`) via `actions/setup-go`, which resolves to the
latest patch of each on every run — so CI naturally tracks patched
toolchains without any change needed. `release.yml` similarly requests
`1.24` (unpinned patch). No action is required in the SDK itself; this is
recorded here as a snapshot of the local dev environment's toolchain at
review time, not a defect to fix in this repository.

## Informational

### I-1: API key redaction verified

`redaction_test.go` (`TestRedactionNeverLeaksKey`) constructs a client with
a known secret key and asserts that none of `fmt.Sprintf("%v", client)`,
`"%+v"`, `"%#v"`, or `client.String()` contain the raw key — only a masked
`awsys_...<last4>` form (via `maskKey` in `awsysco.go`). `Client.GoString()`
and `clientConfig.String()`/`GoString()` all delegate to the same masking
logic, so no `fmt` verb or explicit stringification can leak the raw key.
The same test also confirms `ConfigurationError` (missing key / invalid
base URL) fails in well under 100ms — i.e., before any network call, so a
misconfigured client can't accidentally leak a partial key over the wire
either. No gaps found.

### I-2: No TLS certificate verification bypass

`grep -rn "InsecureSkipVerify\|tls.Config" *.go` (root module, excluding
`_test.go`) returns no matches — the SDK never constructs a custom
`tls.Config` and never sets `InsecureSkipVerify`. `NewClient`'s default
`http.Client` (`awsysco.go`) and any client supplied via `WithHTTPClient`
rely entirely on Go's standard-library TLS defaults (system root CA pool,
certificate verification always on). A caller who supplies their own
`*http.Client` via `WithHTTPClient` could disable verification themselves,
but that is an explicit opt-in outside the SDK's control, not something the
SDK does or enables by default.

### I-3: Base URL scheme validation confirmed

`validateBaseURL` (`awsysco.go`) parses the URL with `net/url` and rejects
anything whose `Scheme` is not exactly `"http"` or `"https"`:

```go
func validateBaseURL(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", &invalidBaseURLError{u: u}
	}
	return u, nil
}
```

This same function gates **both** paths that can set the base URL:
`WithBaseURL(u)` calls it directly and stores a `*ConfigurationError` on
rejection instead of applying the URL; `NewClient`'s `AWSYS_BASE_URL`
environment-variable fallback also routes through it before assigning
`cfg.baseURL`. Either rejection path fails the client (via
`ConfigurationError`, surfaced on the first API call) rather than silently
falling through to an insecure or unintended scheme (e.g. `ftp://`,
`javascript:`, or a bare host with no scheme). Confirmed by
`TestConfigurationErrorInvalidBaseURLFast` in `redaction_test.go` and the
contract-error tests.

### I-4: No logging in the library itself

`grep -rn "log\.\|fmt.Print" *.go` (excluding `_test.go`) returns no
matches in the root package — the SDK itself never writes to stdout/stderr
or a logger. There is therefore no code path in the library where the API
key, `Authorization` header, or full request/response bodies could be
inadvertently written to application logs by the SDK. (Consumers who choose
to log an `*AwsysError`'s `.Raw` field, or a request built with a custom
`http.Client`'s own logging `Transport`, are outside the SDK's control —
see I-5.)

### I-5: Error `Raw` field carries full response bodies, not request headers

`parseErrorResponse` (`http.go`) attaches the complete raw HTTP response
body to `AwsysError.Raw` for debugging. This is response data only — the
function never sees or captures the request's `Authorization` header (the
comment directly above `parseErrorResponse` states this explicitly). A
platform response body is not expected to echo back the caller's API key,
but consumers who print `AwsysError.Raw` (e.g. in verbose logs) should be
aware it contains the platform's unredacted error payload verbatim, which
could include other account-identifying information the platform chose to
return (e.g. email addresses in some error bodies). This is a property of
the API contract, not an SDK defect, but is worth noting for consumers
building their own logging around SDK errors.

### I-6: Default transport is HTTPS; no plaintext HTTP by default

`defaultBaseURL` (`awsysco.go`) is `"https://awsys.co"`. The only way a
client ever talks plaintext HTTP is if a caller explicitly passes an
`http://` URL to `WithBaseURL` or `AWSYS_BASE_URL` — both scenarios require
deliberate action by the integrator (e.g. pointing at a local `httptest`
server in tests, as the SDK's own test suite does). No code path degrades a
`https://` configuration to `http://` automatically.

### I-7: Documentation reference to a non-existent `.env.example`

Unrelated to application security, but noted from the read-through: both
the pre-existing `SECURITY.md` and the pre-v1.2.0 `README.md` instructed
contributors to `cp .env.example .env.test` / `livetest/.env.test`, but no
`.env.example` file exists anywhere in the repository. The v1.2.0
`README.md` rewrite (this effort) removes the dangling reference from the
`README.md` Testing section and instead documents setting `AWSYS_API_KEY`
directly or via `livetest/.env.test`. `SECURITY.md` itself was left
untouched, per instructions, and still references the missing file —
flagged here so it can be corrected or the example file added in a
follow-up.

## Methodology

- **Redaction:** read `redaction_test.go` in full; traced `maskKey`,
  `Client.String/GoString`, `clientConfig.String/GoString` in `awsysco.go`.
- **TLS:** `grep -rn "InsecureSkipVerify\|tls.Config" *.go` across the root
  module (zero matches); read `NewClient`'s default `http.Client`
  construction.
- **Base URL validation:** read `validateBaseURL`, `WithBaseURL`, and the
  `AWSYS_BASE_URL` fallback branch in `NewClient` (`awsysco.go`); cross-
  referenced `TestConfigurationErrorInvalidBaseURLFast` and the contract
  network-error test (`contract_errors_test.go`).
- **Dependency/CVE scan:** installed `govulncheck` fresh
  (`go install golang.org/x/vuln/cmd/govulncheck@latest`) and ran
  `govulncheck ./...` from the repo root; output pasted verbatim above,
  not paraphrased or fabricated.
- **General read-through:** performed alongside the README rewrite and
  GoDoc pass, which required reading every root-package `.go` file in full;
  logging, HTTP-scheme defaults, and error-body handling were reviewed as
  part of that pass.
