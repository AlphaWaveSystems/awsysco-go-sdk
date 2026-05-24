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
}

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

// IsValidationError returns true if err is a 400 Validation error.
func IsValidationError(err error) bool {
	var e *AwsysError
	if errors.As(err, &e) {
		return e.Status == 400
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
