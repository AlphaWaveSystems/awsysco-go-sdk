package awsysco

// TestingOptionWithClock exposes the unexported withClock option to the
// SDK's own black-box test files (package awsysco_test), so retry tests can
// inject a fake retryClock and avoid real sleeps. This file is a _test.go
// file and is never compiled into consumers' binaries — it is not part of
// the public API.
func TestingOptionWithClock(clock retryClock) Option {
	return withClock(clock)
}
