// Package awsysco provides a Go client for the AWSYS.CO URL Shortener API.
//
// Usage:
//
//	client := awsysco.NewClient("awsys_your_api_key")
//
//	link, err := client.Links.Create(ctx, awsysco.CreateLinkInput{
//	    URL: "https://example.com",
//	})
package awsysco

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultBaseURL = "https://awsys.co"

// clientConfig holds the internal configuration for the client.
type clientConfig struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	clock      retryClock
	maxRetries int
	configErr  *ConfigurationError
	timeoutSet bool
	timeoutVal time.Duration
}

// Client is the AWSYS.CO API client.
type Client struct {
	Links         *LinksResource
	Analytics     *AnalyticsResource
	QR            *QRResource
	Folders       *FoldersResource
	Bulk          *BulkResource
	Me            *MeResource
	Tags          *TagsResource
	TrustScore    *TrustScoreResource
	DataExport    *DataExportResource
	Namespace     *NamespaceResource
	UtmTemplates  *UtmTemplatesResource
	Webhooks      *WebhooksResource
	SavedViews    *SavedViewsResource
	CustomDomains *CustomDomainsResource
	Agentlink     *AgentlinkResource
	Affiliate     *AffiliateResource
	Usage         *UsageResource
	Web2App       *Web2AppResource
	Imports       *ImportsResource
	Profile       *ProfileResource

	cfg *clientConfig
}

// Option is a functional option for configuring the client.
type Option func(*clientConfig)

// WithBaseURL overrides the default API base URL. u must include an
// http:// or https:// scheme; an invalid or unsupported value sets a
// ConfigurationError (surfaced on the first API call) rather than panicking.
func WithBaseURL(u string) Option {
	return func(c *clientConfig) {
		if c.configErr != nil {
			return
		}
		parsed, err := validateBaseURL(u)
		if err != nil {
			c.configErr = &ConfigurationError{Message: err.Error(), Err: err}
			return
		}
		warnIfInsecureBaseURL(parsed)
		c.baseURL = parsed
	}
}

// warnIfInsecureBaseURL logs a warning-level line (via the standard log
// package — no logging dependency) if u doesn't use https. It never fails
// or blocks configuration; a non-https base URL is allowed (e.g. for local
// testing against an httptest server) but deserves a visible nudge since API
// traffic, including the API key, would otherwise travel unencrypted.
func warnIfInsecureBaseURL(u string) {
	if !strings.HasPrefix(u, "https://") {
		log.Printf("awsysco: warning: base URL %q does not use https — API traffic, including your API key, will not be encrypted in transit", u)
	}
}

// warnIfKeyMissingPrefix logs a warning-level line if apiKey is non-empty
// but doesn't look like a real AWSYS.CO API key (which always starts with
// "awsys_"). It never fails or blocks configuration.
func warnIfKeyMissingPrefix(apiKey string) {
	if apiKey != "" && !strings.HasPrefix(apiKey, "awsys_") {
		log.Printf("awsysco: warning: API key does not start with %q — this does not look like a valid AWSYS.CO API key", "awsys_")
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = hc
	}
}

// WithTimeout sets the HTTP client timeout. Safe to combine with
// WithHTTPClient regardless of option order — the timeout is applied once,
// after all options have run, to whatever http.Client is in effect.
func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) {
		c.timeoutSet = true
		c.timeoutVal = d
	}
}

// WithMaxRetries sets the maximum number of retry attempts (in addition to
// the initial attempt). 0 disables retries entirely. Negative values are
// clamped to 0. Default is 3.
func WithMaxRetries(n int) Option {
	return func(c *clientConfig) {
		if n < 0 {
			n = 0
		}
		c.maxRetries = n
	}
}

// withClock is unexported — used by the SDK's own tests to inject a fake
// retryClock and avoid real sleeps. Not part of the public API.
func withClock(clock retryClock) Option {
	return func(c *clientConfig) {
		c.clock = clock
	}
}

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

type invalidBaseURLError struct{ u string }

func (e *invalidBaseURLError) Error() string {
	return fmt.Sprintf("invalid base URL %q: must use http:// or https://", e.u)
}

// NewClient creates a new AWSYS.CO API client. It always returns a non-nil
// Client, even when configuration is invalid: an empty apiKey with no
// AWSYS_API_KEY fallback, or an invalid base URL (explicit or via
// AWSYS_BASE_URL), records a ConfigurationError that every API call returns
// immediately, before any network call is made.
func NewClient(apiKey string, opts ...Option) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("AWSYS_API_KEY")
	}
	warnIfKeyMissingPrefix(apiKey)

	cfg := &clientConfig{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		clock:      realClock{},
		maxRetries: 3,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.baseURL == "" {
		if envURL := os.Getenv("AWSYS_BASE_URL"); envURL != "" {
			if parsed, err := validateBaseURL(envURL); err == nil {
				warnIfInsecureBaseURL(parsed)
				cfg.baseURL = parsed
			} else if cfg.configErr == nil {
				cfg.configErr = &ConfigurationError{Message: err.Error(), Err: err}
			}
		}
	}
	if cfg.baseURL == "" {
		cfg.baseURL = defaultBaseURL
	}

	if cfg.timeoutSet {
		if cfg.httpClient == nil {
			cfg.httpClient = &http.Client{}
		}
		cfg.httpClient.Timeout = cfg.timeoutVal
	}

	if apiKey == "" && cfg.configErr == nil {
		cfg.configErr = &ConfigurationError{
			Message: "API key is required: pass one to NewClient or set AWSYS_API_KEY",
		}
	}

	c := &Client{cfg: cfg}
	c.Links = &LinksResource{client: c}
	c.Analytics = &AnalyticsResource{client: c}
	c.QR = &QRResource{client: c}
	c.Folders = &FoldersResource{client: c}
	c.Bulk = &BulkResource{client: c}
	c.Me = &MeResource{client: c}
	c.Tags = &TagsResource{client: c}
	c.TrustScore = &TrustScoreResource{client: c}
	c.DataExport = &DataExportResource{client: c}
	c.Namespace = &NamespaceResource{client: c}
	c.UtmTemplates = &UtmTemplatesResource{client: c}
	c.Webhooks = &WebhooksResource{client: c}
	c.SavedViews = &SavedViewsResource{client: c}
	c.CustomDomains = &CustomDomainsResource{client: c}
	c.Agentlink = &AgentlinkResource{client: c}
	c.Affiliate = &AffiliateResource{client: c}
	c.Usage = &UsageResource{client: c}
	c.Web2App = &Web2AppResource{client: c}
	c.Imports = &ImportsResource{client: c}
	c.Profile = &ProfileResource{client: c}

	return c
}

// maskKey redacts an API key for safe display, keeping only its last 4
// characters (e.g. "awsys_...ab12").
func maskKey(key string) string {
	const prefix = "awsys_"
	trimmed := key
	if len(trimmed) >= len(prefix) && trimmed[:len(prefix)] == prefix {
		trimmed = trimmed[len(prefix):]
	}
	if len(trimmed) <= 4 {
		return prefix + "****"
	}
	return prefix + "..." + trimmed[len(trimmed)-4:]
}

// redactSecret masks an arbitrary secret value (webhook signing secret,
// provider OAuth access token, etc.) for safe display in logging/debug
// output. Unlike maskKey it never reveals any part of the original value —
// these secrets don't have a conventional prefix worth preserving, and
// callers shouldn't be able to reconstruct or narrow down the value from a
// partial leak.
func redactSecret(s string) string {
	if s == "" {
		return `""`
	}
	return "[REDACTED]"
}

// String implements fmt.Stringer, redacting the API key.
func (c *Client) String() string {
	if c == nil || c.cfg == nil {
		return "awsysco.Client{}"
	}
	return "awsysco.Client{baseURL: " + c.cfg.baseURL + ", apiKey: " + maskKey(c.cfg.apiKey) + "}"
}

// GoString implements fmt.GoStringer, redacting the API key.
func (c *Client) GoString() string { return c.String() }

// String implements fmt.Stringer, redacting the API key.
func (c *clientConfig) String() string {
	if c == nil {
		return "awsysco.clientConfig{}"
	}
	return "awsysco.clientConfig{baseURL: " + c.baseURL + ", apiKey: " + maskKey(c.apiKey) + "}"
}

// GoString implements fmt.GoStringer, redacting the API key.
func (c *clientConfig) GoString() string { return c.String() }
