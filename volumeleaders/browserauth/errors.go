package browserauth

import "errors"

// ErrBrowserUnavailable reports that the requested browser has no VolumeLeaders cookies.
var ErrBrowserUnavailable = errors.New("volumeleaders browser unavailable")

// ErrProfileUnavailable reports that the requested profile has no VolumeLeaders cookies.
var ErrProfileUnavailable = errors.New("volumeleaders profile unavailable")

// ErrRequiredCookieMissing reports that required VolumeLeaders cookies were not found.
var ErrRequiredCookieMissing = errors.New("volumeleaders required cookie missing")

// ErrRequestVerificationTokenMissing reports that the request verification token was not found.
var ErrRequestVerificationTokenMissing = errors.New("volumeleaders request verification token missing")

// ErrValidationFailed reports that session validation against VolumeLeaders failed.
var ErrValidationFailed = errors.New("volumeleaders session validation failed")

// ValidationError wraps a validation request failure with the underlying cause.
type ValidationError struct {
	Err error
}

// Error returns a human-readable validation failure.
func (e *ValidationError) Error() string {
	return "volumeleaders session validation failed: " + e.Err.Error()
}

// Unwrap returns validation failure causes for [errors.Is] matching.
func (e *ValidationError) Unwrap() []error {
	return []error{ErrValidationFailed, e.Err}
}
