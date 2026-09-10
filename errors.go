package awsysco

import (
	"errors"
	"fmt"
	"time"
)

// AwsysError is the error type returned by all SDK operations.
type AwsysError struct {
	Message string
	Code    string
	Status  int
	Raw     []byte
	// RetryAfter is the parsed value of a Retry-After response header, if the
	// server sent one — populated for any status, not only 429. Zero when
	// absent or unparseable.
	RetryAfter time.Duration
}

// Error implements the error interface, formatting the message, API error
// code (if any), and HTTP status.
func (e *AwsysError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("awsysco: %s (code=%s, status=%d)", e.Message, e.Code, e.Status)
	}
	return fmt.Sprintf("awsysco: %s (status=%d)", e.Message, e.Status)
}

// RateLimitError is returned when the API rate limit is exceeded (HTTP 429).
type RateLimitError struct {
	AwsysError
	RetryAfter time.Duration
	// ResetsAt is when the current quota window resets, if the platform
	// supplied one (quota-class 429s only). Nil when not provided.
	ResetsAt *time.Time
}

// Unwrap exposes the embedded AwsysError so errors.As(err, &awsysErr) works
// uniformly whether err is a *RateLimitError or a plain *AwsysError.
func (e *RateLimitError) Unwrap() error { return &e.AwsysError }

// ConfigurationError is returned when the client is misconfigured — e.g. no
// API key and no AWSYS_API_KEY fallback, or an invalid base URL — before any
// network call is made.
type ConfigurationError struct {
	Message string
	Err     error
}

// Error implements the error interface.
func (e *ConfigurationError) Error() string { return "awsysco: " + e.Message }

// Unwrap returns the underlying error that caused the configuration error,
// if any, so errors.Is/errors.As can see through it.
func (e *ConfigurationError) Unwrap() error { return e.Err }

// NetworkError wraps a transport-level failure (connection refused/reset,
// DNS failure, etc.) that occurred while attempting a request.
type NetworkError struct {
	Op  string
	URL string
	Err error
}

// Error implements the error interface.
func (e *NetworkError) Error() string {
	return fmt.Sprintf("awsysco: network error (%s): %v", e.Op, e.Err)
}

// Unwrap returns the underlying transport error, so errors.Is (e.g. against
// context.DeadlineExceeded) and errors.As can see through it.
func (e *NetworkError) Unwrap() error { return e.Err }

// TimeoutError is a NetworkError caused by a request or context deadline
// being exceeded. errors.Is(err, context.DeadlineExceeded) works through the
// embedded NetworkError's Unwrap.
type TimeoutError struct {
	NetworkError
}

// IsNotFound returns true if err is a 404 Not Found error.
func IsNotFound(err error) bool {
	var e *AwsysError
	if errors.As(err, &e) {
		return e.Status == 404
	}
	return false
}

// IsAuthError returns true if err is a 401 Unauthorized error.
func IsAuthError(err error) bool {
	var e *AwsysError
	if errors.As(err, &e) {
		return e.Status == 401
	}
	return false
}

// IsRateLimitError returns true if err is a 429 rate limit error.
func IsRateLimitError(err error) bool {
	var e *RateLimitError
	if errors.As(err, &e) {
		return true
	}
	var ae *AwsysError
	if errors.As(err, &ae) {
		return ae.Status == 429
	}
	return false
}

// IsForbidden returns true if err is a 403 Forbidden error.
func IsForbidden(err error) bool {
	var e *AwsysError
	if errors.As(err, &e) {
		return e.Status == 403
	}
	return false
}

// IsValidationError returns true if err is a 400 or 422 Validation error.
// The platform uses 422 for some validation failures (e.g. VALIDATION_FAILED)
// and 400 for others; both map to the same conceptual ValidationError class.
func IsValidationError(err error) bool {
	var e *AwsysError
	if errors.As(err, &e) {
		return e.Status == 400 || e.Status == 422
	}
	return false
}

// IsConflict returns true if err is a 409 Conflict error.
func IsConflict(err error) bool {
	var e *AwsysError
	if errors.As(err, &e) {
		return e.Status == 409
	}
	return false
}

// IsServerError returns true if err is a 5xx server error.
func IsServerError(err error) bool {
	var e *AwsysError
	if errors.As(err, &e) {
		return e.Status >= 500
	}
	return false
}

// IsConfigurationError returns true if err is a *ConfigurationError (client
// misconfiguration detected before any network call was made).
func IsConfigurationError(err error) bool {
	var e *ConfigurationError
	return errors.As(err, &e)
}

// IsNetworkError returns true if err is a *NetworkError OR a *TimeoutError (a
// transport-level failure such as connection refused/reset, DNS failure, or
// a request/context deadline being exceeded). *TimeoutError embeds
// NetworkError by value, not pointer, so errors.As against *NetworkError
// alone would NOT match it — both are checked explicitly here so
// IsNetworkError means "any transport-level problem"; use IsTimeoutError to
// narrow to the timeout-specific case.
func IsNetworkError(err error) bool {
	var e *NetworkError
	if errors.As(err, &e) {
		return true
	}
	var te *TimeoutError
	return errors.As(err, &te)
}

// IsTimeoutError returns true if err is a *TimeoutError (a request or
// context deadline was exceeded). It returns false for a caller-cancelled
// context (context.Canceled), which surfaces as a plain *NetworkError.
func IsTimeoutError(err error) bool {
	var e *TimeoutError
	return errors.As(err, &e)
}

// IsSDKError returns true if err is any of this SDK's well-typed error kinds
// (AwsysError, RateLimitError, ConfigurationError, NetworkError, or
// TimeoutError) as opposed to an unexpected/unwrapped error such as a raw
// encoding/json failure. Every error this SDK returns from an API call
// satisfies IsSDKError.
//
// TimeoutError is checked explicitly, not only via NetworkError: it embeds
// NetworkError by value, so errors.As against *NetworkError alone does not
// match a *TimeoutError (Unwrap resolves to the *underlying* transport
// error, not the embedded NetworkError struct itself).
func IsSDKError(err error) bool {
	if err == nil {
		return false
	}
	var ae *AwsysError
	if errors.As(err, &ae) {
		return true
	}
	var ce *ConfigurationError
	if errors.As(err, &ce) {
		return true
	}
	var ne *NetworkError
	if errors.As(err, &ne) {
		return true
	}
	var te *TimeoutError
	return errors.As(err, &te)
}
