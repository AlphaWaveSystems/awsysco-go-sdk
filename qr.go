package awsysco

import (
	"context"
	"fmt"
	"net/url"
)

// QRResource provides QR code URL construction and settings management.
type QRResource struct {
	client *Client
}

// QROption is a functional option for QR code URL parameters.
type QROption func(*qrConfig)

type qrConfig struct {
	size    int
	color   string
	bgColor string
}

// WithSize sets the QR code size in pixels.
func WithSize(size int) QROption {
	return func(c *qrConfig) {
		c.size = size
	}
}

// WithColor sets the QR code foreground color (hex, without #).
func WithColor(color string) QROption {
	return func(c *qrConfig) {
		c.color = color
	}
}

// WithBGColor sets the QR code background color (hex, without #).
func WithBGColor(bgColor string) QROption {
	return func(c *qrConfig) {
		c.bgColor = bgColor
	}
}

// GetURL constructs the QR code URL for the given short code.
// The returned URL points to the API endpoint that generates the QR image.
func (r *QRResource) GetURL(shortCode string, opts ...QROption) string {
	cfg := &qrConfig{
		size:    300,
		color:   "000000",
		bgColor: "FFFFFF",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	q := url.Values{}
	q.Set("size", fmt.Sprintf("%d", cfg.size))
	q.Set("color", cfg.color)
	q.Set("bgColor", cfg.bgColor)

	return fmt.Sprintf("%s/api/qr/%s?%s", r.client.cfg.baseURL, shortCode, q.Encode())
}

// QRSettings holds the persisted QR code settings for a link.
type QRSettings struct {
	Size            int    `json:"size,omitempty"`
	Color           string `json:"color,omitempty"`
	BgColor         string `json:"bgColor,omitempty"`
	ErrorCorrection string `json:"errorCorrection,omitempty"`
	Margin          int    `json:"margin,omitempty"`
	LogoURL         string `json:"logoUrl,omitempty"`
}

// GetSettings retrieves the saved QR code settings for the given short path.
func (r *QRResource) GetSettings(ctx context.Context, shortPath string) (*QRSettings, error) {
	var settings QRSettings
	path := fmt.Sprintf("/api/link/%s/qr-settings", url.PathEscape(shortPath))
	if err := r.client.doRequest(ctx, "GET", path, nil, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

// UpdateSettings saves QR code settings for the given short path.
func (r *QRResource) UpdateSettings(ctx context.Context, shortPath string, settings QRSettings) (*QRSettings, error) {
	var result QRSettings
	path := fmt.Sprintf("/api/link/%s/qr-settings", url.PathEscape(shortPath))
	if err := r.client.doRequest(ctx, "PUT", path, settings, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
