package awsysco_test

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	awsysco "github.com/AlphaWaveSystems/awsysco-go-sdk"
)

// contractErrorFixture mirrors one entry of the "errors" array in
// testdata/sdk-contract.json.
type contractErrorFixture struct {
	ID          string            `json:"id"`
	Status      *int              `json:"status"`
	Body        json.RawMessage   `json:"body"`
	Headers     map[string]string `json:"headers"`
	ExpectError string            `json:"expect_error"`
	Retry       *bool             `json:"retry"`
}

type contractErrorsFile struct {
	Errors []contractErrorFixture `json:"errors"`
}

func loadContractErrors(t *testing.T) []contractErrorFixture {
	t.Helper()
	raw, err := os.ReadFile("testdata/sdk-contract.json")
	if err != nil {
		t.Fatalf("reading testdata/sdk-contract.json: %v", err)
	}
	var f contractErrorsFile
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("parsing testdata/sdk-contract.json: %v", err)
	}
	return f.Errors
}

// predicateFor maps the contract's conceptual error classes to this SDK's
// Is* predicates.
var predicateFor = map[string]func(error) bool{
	"AuthenticationError": awsysco.IsAuthError,
	"AuthorizationError":  awsysco.IsForbidden,
	"ValidationError":     awsysco.IsValidationError,
	"NotFoundError":       awsysco.IsNotFound,
	"ConflictError":       awsysco.IsConflict,
	"RateLimitError":      awsysco.IsRateLimitError,
	"ServerError":         awsysco.IsServerError,
	// SDKError is the contract's umbrella class for "some well-typed SDK
	// error, not a raw unwrapped decode/parse exception" — used by
	// err_2xx_malformed_json, where a 2xx status rules out every
	// status-specific predicate above.
	"SDKError": awsysco.IsSDKError,
}

// probe exercises the shared error-parsing path in http.go — the exact
// resource/method doesn't matter, only that it's a single round trip through
// doRequest. WithMaxRetries(0) keeps every scenario a single request: this
// test validates error-body parsing and Is*/Status/Code/ResetsAt mapping,
// not retry behavior (see the dedicated retry tests for that).
func probe(ctx context.Context, client *awsysco.Client) error {
	_, err := client.Links.Get(ctx, "x")
	return err
}

func TestContractErrors(t *testing.T) {
	fixtures := loadContractErrors(t)

	for _, f := range fixtures {
		f := f
		t.Run(f.ID, func(t *testing.T) {
			switch f.ID {
			case "err_timeout":
				testContractTimeout(t)
				return
			case "err_network":
				testContractNetworkError(t)
				return
			case "err_user_cancel":
				testContractUserCancel(t)
				return
			}

			if f.Status == nil {
				t.Fatalf("fixture %s has no status and isn't one of the special-cased transport scenarios", f.ID)
			}

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for k, v := range f.Headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(*f.Status)
				if len(f.Body) > 0 {
					var raw string
					// A handful of fixtures encode a non-JSON (e.g. HTML)
					// body as a plain JSON string; write it as raw bytes so
					// the SDK really sees a non-JSON response.
					if json.Unmarshal(f.Body, &raw) == nil {
						_, _ = w.Write([]byte(raw))
					} else {
						_, _ = w.Write(f.Body)
					}
				}
			}))
			defer srv.Close()

			client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL), awsysco.WithMaxRetries(0))
			err := probe(context.Background(), client)

			if err == nil {
				t.Fatal("expected an error, got nil")
			}

			predicate, ok := predicateFor[f.ExpectError]
			if !ok {
				t.Fatalf("no predicate mapping registered for expect_error %q", f.ExpectError)
			}
			if !predicate(err) {
				t.Errorf("predicate for %s returned false for error: %v", f.ExpectError, err)
			}

			// RateLimitError embeds AwsysError by value with no Unwrap of its
			// own, so errors.As(&ae) does not reach through it — extract the
			// embedded AwsysError directly for that case instead.
			var ae *awsysco.AwsysError
			var re *awsysco.RateLimitError
			switch {
			case errors.As(err, &re):
				ae = &re.AwsysError
			case errors.As(err, &ae):
				// already set
			default:
				t.Fatalf("error does not unwrap to *awsysco.AwsysError or *awsysco.RateLimitError: %v (%T)", err, err)
			}
			if ae.Status != *f.Status {
				t.Errorf("Status = %d, want %d", ae.Status, *f.Status)
			}
			assertErrorBodyFields(t, f.Body, ae)

			if f.ExpectError == "RateLimitError" {
				if re == nil {
					t.Errorf("expected *awsysco.RateLimitError in the chain, got %T", err)
				} else if resetsAt := extractResetsAt(f.Body); resetsAt != "" {
					// err_429_hourly's resetsAt ("2026-09-06-04:00 UTC") is
					// deliberately not RFC3339 — the parser must leave
					// ResetsAt nil rather than erroring; a well-formed
					// resetsAt should parse successfully.
					if _, perr := time.Parse(time.RFC3339, resetsAt); perr == nil && re.ResetsAt == nil {
						t.Errorf("ResetsAt should be populated from a valid RFC3339 resetsAt field, got nil")
					}
				}
			}

			// Fixture-specific extras for the new 1.0.9 error ids: confirm
			// the header/body-driven fields that drive Part B's retry-cap
			// and quota-class fixes are actually populated as expected. The
			// generic loop above always uses WithMaxRetries(0), so it can't
			// observe retry *behavior* (see the dedicated retry_test.go
			// cases for that) — this only checks the parsed fields.
			switch f.ID {
			case "err_429_resets_at_only":
				if re == nil || re.ResetsAt == nil {
					t.Error("expected ResetsAt to be populated from the resetsAt body field, got nil")
				}
				if re != nil && re.Code != "" {
					t.Errorf("Code = %q, want empty (this fixture has no quota code — resetsAt alone must drive quota-class detection)", re.Code)
				}
			case "err_429_retry_after_oversized", "err_503_retry_after_oversized":
				if ae.RetryAfter <= 30*time.Second {
					t.Errorf("RetryAfter = %v, want > 30s (parsed from the oversized Retry-After header)", ae.RetryAfter)
				}
			}
		})
	}
}

// testContractUserCancel covers err_user_cancel: a request aborted by the
// CALLER cancelling its context (as opposed to a deadline elapsing) must
// surface as something satisfying errors.Is(err, context.Canceled), and must
// never be reported as *awsysco.TimeoutError — a TimeoutError implies the
// server was too slow, not that the caller gave up on purpose.
func testContractUserCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until the client gives up (or the test times out) —
		// simulates a slow server the caller deliberately cancels out on.
		<-r.Context().Done()
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL), awsysco.WithMaxRetries(0))

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	err := probe(ctx, client)

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("errors.Is(err, context.Canceled) = false, want true: %v", err)
	}
	var te *awsysco.TimeoutError
	if errors.As(err, &te) {
		t.Error("a caller-cancelled context must not surface as *awsysco.TimeoutError")
	}
}

func assertErrorBodyFields(t *testing.T, body json.RawMessage, ae *awsysco.AwsysError) {
	t.Helper()
	if len(body) == 0 {
		return
	}
	var raw string
	if json.Unmarshal(body, &raw) == nil {
		// Non-JSON body case (err_non_json): Message should fall back to
		// the HTTP status text, and Raw must still retain the original
		// bytes for debugging — never crash, never leak the raw HTML as if
		// it were a structured message field.
		if ae.Message == "" {
			t.Error("Message should fall back to a non-empty status-text default for a non-JSON body")
		}
		if string(ae.Raw) != raw {
			t.Errorf("Raw = %q, want original non-JSON body %q", ae.Raw, raw)
		}
		return
	}

	var shape struct {
		Error   json.RawMessage `json:"error"`
		Message string          `json:"message"`
		Code    string          `json:"code"`
		Success *bool           `json:"success"`
	}
	if err := json.Unmarshal(body, &shape); err != nil {
		t.Fatalf("fixture body is not valid JSON: %v", err)
	}

	wantMessage := shape.Message
	if wantMessage == "" && len(shape.Error) > 0 {
		var errStr string
		if json.Unmarshal(shape.Error, &errStr) == nil {
			wantMessage = errStr // {"error": "<string>", code} shape
		}
	}
	if wantMessage != "" && ae.Message != wantMessage {
		t.Errorf("Message = %q, want %q", ae.Message, wantMessage)
	}
	if shape.Code != "" && ae.Code != shape.Code {
		t.Errorf("Code = %q, want %q", ae.Code, shape.Code)
	}
}

func extractResetsAt(body json.RawMessage) string {
	var shape struct {
		ResetsAt string `json:"resetsAt"`
	}
	_ = json.Unmarshal(body, &shape)
	return shape.ResetsAt
}

// testContractTimeout covers err_timeout: a request that exceeds the
// client's timeout must surface as *awsysco.TimeoutError, wrapping
// context.DeadlineExceeded.
func testContractTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL(srv.URL), awsysco.WithTimeout(10*time.Millisecond), awsysco.WithMaxRetries(0))
	err := probe(context.Background(), client)

	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	var te *awsysco.TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("expected *awsysco.TimeoutError, got %T: %v", err, err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Error("errors.Is(err, context.DeadlineExceeded) should be true through TimeoutError's Unwrap chain")
	}
}

// testContractNetworkError covers err_network: a connection that can never
// succeed (nothing listening) must surface as *awsysco.NetworkError, not a
// TimeoutError and not a generic error.
func testContractNetworkError(t *testing.T) {
	// Reserve a port, then close the listener immediately so nothing is
	// listening on it — guarantees a connection-refused error without
	// depending on any particular unused port being free system-wide.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to reserve a local port: %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	client := awsysco.NewClient("awsys_test", awsysco.WithBaseURL("http://"+addr), awsysco.WithMaxRetries(0))
	callErr := probe(context.Background(), client)

	if callErr == nil {
		t.Fatal("expected a network error, got nil")
	}
	var ne *awsysco.NetworkError
	if !errors.As(callErr, &ne) {
		t.Fatalf("expected *awsysco.NetworkError, got %T: %v", callErr, callErr)
	}
	var te *awsysco.TimeoutError
	if errors.As(callErr, &te) {
		t.Error("a connection-refused failure should be NetworkError, not TimeoutError")
	}
}
