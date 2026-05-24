package awsysco

import (
	"fmt"
	"net/url"
)

// QRResource provides QR code URL construction.
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
