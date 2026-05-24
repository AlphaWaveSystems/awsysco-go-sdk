package awsysco

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	sdkVersion = "0.1.0"
	userAgent  = "awsysco-go-sdk/" + sdkVersion
)

// doRequest performs an HTTP request, handles auth, encodes/decodes JSON,
// maps error responses, and auto-retries on 429.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	const maxAttempts = 3
	backoff := time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := c.doRequestOnce(ctx, method, path, body, result)
		if err == nil {
			return nil
		}

		// Check if rate limited
		var rlErr *RateLimitError
		if isRateLimit(err, &rlErr) {
			if attempt == maxAttempts-1 {
				return err
			}
			wait := backoff
			if rlErr != nil && rlErr.RetryAfter > 0 {
				wait = rlErr.RetryAfter
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			backoff *= 2
			continue
		}

		return err
	}

	return fmt.Errorf("awsysco: max retry attempts exceeded")
}

func isRateLimit(err error, out **RateLimitError) bool {
	if err == nil {
		return false
	}
	if rl, ok := err.(*RateLimitError); ok {
		if out != nil {
			*out = rl
		}
		return true
	}
	if ae, ok := err.(*AwsysError); ok {
		return ae.Status == 429
	}
	return false
}

func (c *Client) doRequestOnce(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("awsysco: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.cfg.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("awsysco: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.apiKey)
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("awsysco: http request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("awsysco: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp.StatusCode, raw, resp.Header)
	}

	if result != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, result); err != nil {
			return fmt.Errorf("awsysco: decode response: %w", err)
		}
	}

	return nil
}

func parseErrorResponse(status int, raw []byte, headers http.Header) error {
	var apiErr struct {
		Error string `json:"error"`
		Code  string `json:"code"`
	}
	_ = json.Unmarshal(raw, &apiErr)

	msg := apiErr.Error
	if msg == "" {
		msg = http.StatusText(status)
	}

	base := AwsysError{
		Message: msg,
		Code:    apiErr.Code,
		Status:  status,
		Raw:     raw,
	}

	if status == 429 {
		var retryAfter time.Duration
		if ra := headers.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil {
				retryAfter = time.Duration(secs) * time.Second
			}
		}
		return &RateLimitError{
			AwsysError: base,
			RetryAfter: retryAfter,
		}
	}

	return &base
}
