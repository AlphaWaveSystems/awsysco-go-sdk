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
	"net/http"
	"time"
)

const defaultBaseURL = "https://awsys.co"

// clientConfig holds the internal configuration for the client.
type clientConfig struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
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

	cfg *clientConfig
}

// Option is a functional option for configuring the client.
type Option func(*clientConfig)

// WithBaseURL overrides the default API base URL.
func WithBaseURL(u string) Option {
	return func(c *clientConfig) {
		c.baseURL = u
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *clientConfig) {
		c.httpClient = hc
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = d
	}
}

// NewClient creates a new AWSYS.CO API client.
func NewClient(apiKey string, opts ...Option) *Client {
	cfg := &clientConfig{
		apiKey:  apiKey,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(cfg)
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

	return c
}
