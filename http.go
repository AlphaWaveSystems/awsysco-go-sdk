package awsysco

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	sdkVersion = "1.2.0"
	userAgent  = "awsysco-go-sdk/" + sdkVersion
)

// retryClock abstracts the backoff wait so tests can avoid real sleeps.
type retryClock interface {
	// Sleep blocks until d elapses or ctx is done, whichever comes first.
	// It returns ctx.Err() if ctx is done before d elapses.
	Sleep(ctx context.Context, d time.Duration) error
}

type realClock struct{}

func (realClock) Sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// quotaCodes are 429 error codes that represent an exhausted quota window —
// retrying cannot help until the window resets, so these fail immediately.
var quotaCodes = map[string]bool{
	"HOURLY_LIMIT_EXCEEDED":  true,
	"DAILY_LIMIT_EXCEEDED":   true,
	"MONTHLY_LIMIT_EXCEEDED": true,
}

func isIdempotentMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPut, http.MethodDelete:
		return true
	default:
		return false
	}
}

// maxRetryAfter caps how long the SDK will honor a server-supplied
// Retry-After value; anything larger means "don't bother retrying at all"
// rather than sleeping for the full duration.
const maxRetryAfter = 30 * time.Second

// retryDecision reports whether err is retryable for the given method and,
// if so, how long to wait before the next attempt.
func retryDecision(err error, method string, attempt int) (wait time.Duration, retryable bool) {
	var rl *RateLimitError
	if errors.As(err, &rl) {
		// Quota-class 429s are never retried: either a recognized quota code,
		// or (even with no code at all) a resetsAt in the body — both mean
		// retrying before the window resets cannot help.
		if quotaCodes[rl.Code] || rl.ResetsAt != nil {
			return 0, false
		}
		if rl.RetryAfter > 0 {
			if rl.RetryAfter > maxRetryAfter {
				// Server asked us to wait longer than we're willing to sleep
				// for — fail fast instead of blocking the caller.
				return 0, false
			}
			return rl.RetryAfter, true
		}
		return backoff(attempt), true
	}

	var timeoutErr *TimeoutError
	if errors.As(err, &timeoutErr) {
		return 0, false
	}

	var netErr *NetworkError
	if errors.As(err, &netErr) {
		if isIdempotentMethod(method) {
			return backoff(attempt), true
		}
		return 0, false
	}

	var ae *AwsysError
	if errors.As(err, &ae) {
		if (ae.Status == 502 || ae.Status == 503 || ae.Status == 504) && isIdempotentMethod(method) {
			if ae.RetryAfter > 0 {
				if ae.RetryAfter > maxRetryAfter {
					return 0, false
				}
				return ae.RetryAfter, true
			}
			return backoff(attempt), true
		}
		return 0, false
	}

	return 0, false
}

// backoff returns a full-jitter exponential backoff duration for the given
// (0-indexed) attempt: min(1s*2^attempt, 30s), then random(0, that).
func backoff(attempt int) time.Duration {
	base := time.Second * time.Duration(1<<uint(attempt))
	const maxBackoff = 30 * time.Second
	if base > maxBackoff {
		base = maxBackoff
	}
	// Full jitter (not security-sensitive — just needs to spread out retries).
	return time.Duration(rand.Int63n(int64(base) + 1))
}

// executeWithRetry runs attempt, retrying per retryDecision up to
// c.cfg.maxRetries additional times, honoring ctx cancellation during any
// backoff wait.
func (c *Client) executeWithRetry(ctx context.Context, method string, attempt func(context.Context) error) error {
	var lastErr error
	for i := 0; ; i++ {
		lastErr = attempt(ctx)
		if lastErr == nil {
			return nil
		}
		if i >= c.cfg.maxRetries {
			return lastErr
		}
		wait, retryable := retryDecision(lastErr, method, i)
		if !retryable {
			return lastErr
		}
		if err := c.cfg.clock.Sleep(ctx, wait); err != nil {
			return err
		}
	}
}

// buildURL joins the client's base URL with path, preserving any query
// string already present in path.
func (c *Client) buildURL(path string) (string, error) {
	base, err := url.Parse(c.cfg.baseURL)
	if err != nil {
		return "", fmt.Errorf("awsysco: invalid base URL: %w", err)
	}
	p := path
	rawQuery := ""
	if idx := indexByte(p, '?'); idx >= 0 {
		rawQuery = p[idx+1:]
		p = p[:idx]
	}
	joined := base.JoinPath(p)
	joined.RawQuery = rawQuery
	return joined.String(), nil
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// wrapTransportError converts a transport-level error from httpClient.Do
// into a NetworkError or TimeoutError.
//
// A caller-cancelled context (context.Canceled) is deliberately never
// promoted to *TimeoutError, even though it also aborts the in-flight
// request — a TimeoutError implies the server was too slow, but here the
// caller gave up on purpose. It surfaces as a plain *NetworkError instead,
// so errors.Is(err, context.Canceled) still works through NetworkError's
// Unwrap, and errors.As(err, &timeoutErr) correctly returns false.
func wrapTransportError(method, url string, err error) error {
	base := NetworkError{Op: method, URL: url, Err: err}
	if errors.Is(err, context.Canceled) {
		// context.Canceled is a stable stdlib sentinel across Go versions,
		// so err already unwraps to it reliably — no need to re-wrap.
		return &base
	}
	var netErr net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()) {
		// Whether the underlying error already unwraps to
		// context.DeadlineExceeded depends on the Go toolchain version and
		// exactly which stage of the request timed out (net.Error.Timeout()
		// can be satisfied by errors that never mention
		// context.DeadlineExceeded at all, e.g. some net.OpError/poll
		// deadline errors). Guarantee the contract's
		// errors.Is(err, context.DeadlineExceeded) requirement explicitly
		// rather than relying on that stdlib behavior.
		base.Err = fmt.Errorf("%w: %w", context.DeadlineExceeded, err)
		return &TimeoutError{NetworkError: base}
	}
	return &base
}

// doRequest performs an HTTP request, handles auth, encodes/decodes JSON,
// maps error responses, and auto-retries per the SDK's retry policy.
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	if c.cfg.configErr != nil {
		return c.cfg.configErr
	}
	return c.executeWithRetry(ctx, method, func(ctx context.Context) error {
		return c.doRequestOnce(ctx, method, path, body, result)
	})
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

	reqURL, err := c.buildURL(path)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
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
		return wrapTransportError(method, reqURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		// A Client.Timeout deadline (or a caller-cancelled context) can fire
		// mid-body-read just as easily as during Do() — e.g. a server that
		// sends headers, flushes, then stalls the body. Route this through
		// the same transport-error classifier as Do()'s error so it still
		// becomes *TimeoutError/*NetworkError as appropriate, not a bare
		// wrapped error the caller can't type-switch on.
		return wrapTransportError(method, reqURL, err)
	}

	if resp.StatusCode >= 400 {
		return parseErrorResponse(resp.StatusCode, raw, resp.Header)
	}

	if result != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, result); err != nil {
			// A 2xx response with a body that isn't valid JSON (e.g. an HTML
			// interstitial from a proxy) must never surface as a raw
			// encoding/json error — wrap it into a well-typed SDK error so
			// callers can handle it uniformly via errors.As/IsSDKError.
			return &AwsysError{
				Message: "received a 2xx response with an unparseable body",
				Status:  resp.StatusCode,
				Raw:     raw,
			}
		}
	}

	return nil
}

// doText performs a request and returns the response body as a plain string.
// It is used for endpoints that return non-JSON bodies (e.g. CSV exports).
func (c *Client) doText(ctx context.Context, method, path string, body interface{}) (string, error) {
	if c.cfg.configErr != nil {
		return "", c.cfg.configErr
	}
	var result string
	err := c.executeWithRetry(ctx, method, func(ctx context.Context) error {
		r, err := c.doTextOnce(ctx, method, path, body)
		if err != nil {
			return err
		}
		result = r
		return nil
	})
	return result, err
}

func (c *Client) doTextOnce(ctx context.Context, method, path string, body interface{}) (string, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return "", fmt.Errorf("awsysco: marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	reqURL, err := c.buildURL(path)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return "", fmt.Errorf("awsysco: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.cfg.apiKey)
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.cfg.httpClient.Do(req)
	if err != nil {
		return "", wrapTransportError(method, reqURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		// See the matching comment in doRequestOnce: a body-read failure can
		// be a timeout/cancellation just as easily as Do() itself.
		return "", wrapTransportError(method, reqURL, err)
	}

	if resp.StatusCode >= 400 {
		return "", parseErrorResponse(resp.StatusCode, raw, resp.Header)
	}

	return string(raw), nil
}

// errorBody covers all four platform error-body shapes:
//
//	{error:true,  code, message}
//	{error:"<string>", code}        -- the string IS the message
//	{error:true,  code}              -- no message, synthesize from code
//	{success:false, message, code}
type errorBody struct {
	Error    json.RawMessage `json:"error"`
	Message  string          `json:"message"`
	Code     string          `json:"code"`
	Success  *bool           `json:"success"`
	ResetsAt string          `json:"resetsAt"`
}

// parseErrorResponse builds an error from a non-2xx HTTP response, tolerating
// all documented body shapes plus non-JSON/empty bodies (e.g. an Express HTML
// 404 page). It never panics or returns a decode error — worst case it falls
// back to the HTTP status text. Raw always retains the original response
// bytes for debugging; the Authorization request header is never captured
// here (only response bytes ever land in Raw).
func parseErrorResponse(status int, raw []byte, headers http.Header) error {
	var body errorBody
	msg := ""
	code := ""

	if len(raw) > 0 && json.Unmarshal(raw, &body) == nil {
		code = body.Code
		if len(body.Error) > 0 {
			var errStr string
			if json.Unmarshal(body.Error, &errStr) == nil {
				// Shape: {error:"<string>", code} — the string IS the message.
				msg = errStr
			} else {
				// Shape: {error:true, code, message?}
				msg = body.Message
			}
		} else if body.Success != nil && !*body.Success {
			// Shape: {success:false, message, code}
			msg = body.Message
		} else {
			msg = body.Message
		}
	}

	if msg == "" {
		if code != "" {
			msg = fmt.Sprintf("request failed with code %s", code)
		} else {
			msg = http.StatusText(status)
		}
	}

	retryAfter := parseRetryAfterHeader(headers)

	base := AwsysError{
		Message:    msg,
		Code:       code,
		Status:     status,
		Raw:        raw,
		RetryAfter: retryAfter,
	}

	if status == 429 {
		var resetsAt *time.Time
		if body.ResetsAt != "" {
			if t, err := time.Parse(time.RFC3339, body.ResetsAt); err == nil {
				resetsAt = &t
			}
		}
		return &RateLimitError{
			AwsysError: base,
			RetryAfter: retryAfter,
			ResetsAt:   resetsAt,
		}
	}

	return &base
}

// parseRetryAfterHeader parses a Retry-After response header (either an
// integer number of seconds or an HTTP-date), returning 0 if absent,
// malformed, or already in the past. Applies to any status code — the
// platform can send Retry-After on a retryable 5xx as well as on 429.
func parseRetryAfterHeader(headers http.Header) time.Duration {
	ra := headers.Get("Retry-After")
	if ra == "" {
		return 0
	}
	if secs, err := strconv.Atoi(ra); err == nil {
		return time.Duration(secs) * time.Second
	}
	if t, err := http.ParseTime(ra); err == nil {
		if d := time.Until(t); d > 0 {
			return d
		}
	}
	return 0
}
