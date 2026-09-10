package awsysco_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

// fakeClock records requested sleep durations and returns immediately,
// so retry tests never actually wait.
type fakeClock struct {
	mu     sync.Mutex
	sleeps []time.Duration
}

func (c *fakeClock) Sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	c.mu.Lock()
	c.sleeps = append(c.sleeps, d)
	c.mu.Unlock()
	return nil
}

func (c *fakeClock) durations() []time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]time.Duration, len(c.sleeps))
	copy(out, c.sleeps)
	return out
}

func newTestClientWithClock(baseURL string, clock *fakeClock, opts ...awsysco.Option) *awsysco.Client {
	all := append([]awsysco.Option{
		awsysco.WithBaseURL(baseURL),
		awsysco.TestingOptionWithClock(clock),
	}, opts...)
	return awsysco.NewClient("awsys_test", all...)
}

func TestRetry429ThenSuccess(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&count, 1)
		if n <= 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":true,"code":"RATE_LIMITED","message":"slow down"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"shortCode":"abc123"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	link, err := client.Links.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if link.ShortCode != "abc123" {
		t.Errorf("ShortCode = %q, want abc123", link.ShortCode)
	}
	if got := atomic.LoadInt32(&count); got != 4 {
		t.Fatalf("request count = %d, want 4 (3 failures + 1 success)", got)
	}

	sleeps := clock.durations()
	if len(sleeps) != 3 {
		t.Fatalf("recorded sleeps = %d, want 3", len(sleeps))
	}
	for i, d := range sleeps {
		max := time.Duration(1<<uint(i)) * time.Second
		if max > 30*time.Second {
			max = 30 * time.Second
		}
		if d < 0 || d > max {
			t.Errorf("sleep[%d] = %v, want in [0, %v]", i, d, max)
		}
	}
}

func TestRetryQuota429NotRetried(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":true,"code":"MONTHLY_LIMIT_EXCEEDED","message":"quota exhausted","resetsAt":"2026-10-01T00:00:00Z"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	_, err := client.Links.Get(context.Background(), "x")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("request count = %d, want 1 (no retry on quota-class 429)", got)
	}
	var rl *awsysco.RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("error = %T, want *RateLimitError", err)
	}
	if rl.Code != "MONTHLY_LIMIT_EXCEEDED" {
		t.Errorf("Code = %q, want MONTHLY_LIMIT_EXCEEDED", rl.Code)
	}
	if len(clock.durations()) != 0 {
		t.Errorf("expected no sleeps, got %v", clock.durations())
	}
}

func TestRetry503RetriedOnGET(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&count, 1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"shortCode":"abc123"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	_, err := client.Links.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got := atomic.LoadInt32(&count); got != 3 {
		t.Fatalf("request count = %d, want 3", got)
	}
}

func TestRetry503NotRetriedOnPOST(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	_, err := client.Links.Create(context.Background(), awsysco.CreateLinkInput{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("request count = %d, want 1 (POST must not retry on 503)", got)
	}
	if !awsysco.IsServerError(err) {
		t.Errorf("expected IsServerError, got %v", err)
	}
}

func TestRetryWithMaxRetriesZero(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":true,"code":"RATE_LIMITED","message":"slow down"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock, awsysco.WithMaxRetries(0))

	_, err := client.Links.Get(context.Background(), "x")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("request count = %d, want 1 (WithMaxRetries(0) disables retries)", got)
	}
}

func TestRetryWithMaxRetriesOne(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":true,"code":"RATE_LIMITED","message":"slow down"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock, awsysco.WithMaxRetries(1))

	_, err := client.Links.Get(context.Background(), "x")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := atomic.LoadInt32(&count); got != 2 {
		t.Fatalf("request count = %d, want 2 (1 initial + 1 retry)", got)
	}
}

func TestRetryContextCancelledDuringWait(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":true,"code":"RATE_LIMITED","message":"slow down"}`))
	}))
	defer srv.Close()

	// Use the real clock here (not the fake) so the retry loop actually
	// waits — then prove ctx cancellation cuts that wait short instead of
	// running the full ~1s/2s/4s backoff schedule.
	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := client.Links.Get(ctx, "x")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("elapsed = %v, want well under the 1s+ backoff schedule (ctx cancellation should cut it short)", elapsed)
	}
}

func TestRetryAfterIntegerSeconds(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&count, 1) == 1 {
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":true,"code":"RATE_LIMITED","message":"slow down"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"shortCode":"abc123"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	_, err := client.Links.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	sleeps := clock.durations()
	if len(sleeps) != 1 {
		t.Fatalf("recorded sleeps = %d, want 1", len(sleeps))
	}
	if sleeps[0] != 5*time.Second {
		t.Errorf("sleep = %v, want exactly 5s (explicit Retry-After, no jitter)", sleeps[0])
	}
}

// TestRetryQuotaResetsAtOnlyNotRetried covers err_429_resets_at_only: a 429
// with a resetsAt field but no recognized quota code must still be treated
// as quota-class and never retried.
func TestRetryQuotaResetsAtOnlyNotRetried(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":true,"message":"Limit exceeded","resetsAt":"2026-10-01T00:00:00Z"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	_, err := client.Links.Get(context.Background(), "x")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("request count = %d, want 1 (no retry — resetsAt alone marks this quota-class)", got)
	}
	var rl *awsysco.RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("error = %T, want *RateLimitError", err)
	}
	if rl.Code != "" {
		t.Errorf("Code = %q, want empty (no quota code in this fixture)", rl.Code)
	}
	if rl.ResetsAt == nil {
		t.Error("ResetsAt should be populated")
	}
	if len(clock.durations()) != 0 {
		t.Errorf("expected no sleeps, got %v", clock.durations())
	}
}

// TestRetry429RetryAfterOversizedNotRetried covers err_429_retry_after_oversized:
// a Retry-After value above the 30s cap must fail immediately, never sleep
// for the full duration.
func TestRetry429RetryAfterOversizedNotRetried(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.Header().Set("Retry-After", "86400")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":true,"message":"Too many requests"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	start := time.Now()
	_, err := client.Links.Get(context.Background(), "x")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("request count = %d, want 1 (Retry-After above the 30s cap must not retry)", got)
	}
	if len(clock.durations()) != 0 {
		t.Errorf("expected no sleeps (fake clock never invoked), got %v", clock.durations())
	}
	if elapsed > 500*time.Millisecond {
		t.Errorf("elapsed = %v, want fast — must never actually sleep 86400s", elapsed)
	}
	var rl *awsysco.RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("error = %T, want *RateLimitError", err)
	}
	if rl.RetryAfter != 86400*time.Second {
		t.Errorf("RetryAfter = %v, want 24h (parsed from the header even though it's not honored)", rl.RetryAfter)
	}
}

// TestRetry503RetryAfterOversizedNotRetried covers err_503_retry_after_oversized:
// a retryable 5xx with a Retry-After above the 30s cap must also fail
// immediately rather than sleeping, even on an idempotent GET.
func TestRetry503RetryAfterOversizedNotRetried(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.Header().Set("Retry-After", "3600")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"error":true,"message":"unavailable"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	_, err := client.Links.Get(context.Background(), "x")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if got := atomic.LoadInt32(&count); got != 1 {
		t.Fatalf("request count = %d, want 1 (Retry-After above the 30s cap must not retry, even on GET)", got)
	}
	if !awsysco.IsServerError(err) {
		t.Errorf("expected IsServerError, got %v", err)
	}
	if len(clock.durations()) != 0 {
		t.Errorf("expected no sleeps, got %v", clock.durations())
	}
}

// TestRetry503RetryAfterUnderCapHonored is the control case: a 503's
// Retry-After under the 30s cap IS honored as the wait, same as 429.
func TestRetry503RetryAfterUnderCapHonored(t *testing.T) {
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&count, 1) == 1 {
			w.Header().Set("Retry-After", "7")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"shortCode":"abc123"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	_, err := client.Links.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	sleeps := clock.durations()
	if len(sleeps) != 1 {
		t.Fatalf("recorded sleeps = %d, want 1", len(sleeps))
	}
	if sleeps[0] != 7*time.Second {
		t.Errorf("sleep = %v, want exactly 7s (explicit Retry-After under the cap, no jitter)", sleeps[0])
	}
}

func TestRetryAfterHTTPDate(t *testing.T) {
	var count int32
	target := time.Now().Add(2 * time.Second).UTC()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&count, 1) == 1 {
			w.Header().Set("Retry-After", target.Format(http.TimeFormat))
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":true,"code":"RATE_LIMITED","message":"slow down"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"shortCode":"abc123"}`))
	}))
	defer srv.Close()

	clock := &fakeClock{}
	client := newTestClientWithClock(srv.URL, clock)

	_, err := client.Links.Get(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	sleeps := clock.durations()
	if len(sleeps) != 1 {
		t.Fatalf("recorded sleeps = %d, want 1", len(sleeps))
	}
	if sleeps[0] <= 0 || sleeps[0] > 3*time.Second {
		t.Errorf("sleep = %v, want roughly ~2s (HTTP-date Retry-After ~2s in the future)", sleeps[0])
	}
}
