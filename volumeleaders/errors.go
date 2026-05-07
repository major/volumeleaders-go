package volumeleaders

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const errorBodyPreviewLimit = 4096

// ErrCrossOriginRedirect reports that a redirect attempted to leave the
// original request origin.
var ErrCrossOriginRedirect = errors.New("volumeleaders cross-origin redirect blocked")

// ErrBrowserCookiesUnavailable reports that browser cookie stores did not
// contain the cookies required for authenticated VolumeLeaders requests.
var ErrBrowserCookiesUnavailable = errors.New("volumeleaders browser cookies unavailable")

// ErrSessionExpired reports that the supplied browser cookies no longer reach
// authenticated VolumeLeaders pages.
var ErrSessionExpired = errors.New("volumeleaders session expired")

// ErrXSRFTokenNotFound reports that the authenticated page did not contain the
// ASP.NET request verification token expected by XHR endpoints.
var ErrXSRFTokenNotFound = errors.New("volumeleaders XSRF token not found")

// ErrInvalidSession reports that a Session is missing authentication material
// required by browser-backed VolumeLeaders endpoints.
var ErrInvalidSession = errors.New("volumeleaders invalid session")

// ErrInvalidQuery reports that a typed request contains unsupported filter or
// pagination values.
var ErrInvalidQuery = errors.New("volumeleaders invalid query")

// ErrUnexpectedContent reports that VolumeLeaders returned content that could
// not be decoded as the expected response model.
var ErrUnexpectedContent = errors.New("volumeleaders unexpected response content")

type sessionExpiredError struct {
	redirectPath string
}

func (e sessionExpiredError) Error() string {
	return "volumeleaders session expired: redirected to " + e.redirectPath
}

func (e sessionExpiredError) Unwrap() error {
	return ErrSessionExpired
}

// SessionValidationError reports which Session fields are missing.
type SessionValidationError struct {
	Missing []string
}

// Error returns a human-readable session validation failure.
func (e *SessionValidationError) Error() string {
	return "volumeleaders invalid session: missing " + strings.Join(e.Missing, ", ")
}

// Unwrap returns ErrInvalidSession for [errors.Is] matching.
func (e *SessionValidationError) Unwrap() error {
	return ErrInvalidSession
}

// StatusError represents a non-2xx HTTP response from VolumeLeaders.
// Body contains a bounded preview for callers that need diagnostics, but Error
// intentionally omits it to avoid leaking authenticated response content into
// logs.
type StatusError struct {
	StatusCode  int
	Method      string
	Path        string
	ContentType string
	Header      http.Header
	RetryAfter  string
	Body        string
}

// Error returns a human-readable HTTP status failure.
func (e *StatusError) Error() string {
	if e.Method != "" || e.Path != "" {
		return fmt.Sprintf("volumeleaders API error %s %s %d", e.Method, e.Path, e.StatusCode)
	}
	return fmt.Sprintf("volumeleaders API error %d", e.StatusCode)
}

// StatusCode returns the HTTP status code from err when it wraps a StatusError.
func StatusCode(err error) (int, bool) {
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr == nil {
		return 0, false
	}
	return statusErr.StatusCode, true
}

// IsStatusCode reports whether err wraps a StatusError with statusCode.
func IsStatusCode(err error, statusCode int) bool {
	actual, ok := StatusCode(err)
	return ok && actual == statusCode
}

// IsRateLimit reports whether err wraps an HTTP 429 StatusError.
func IsRateLimit(err error) bool {
	return IsStatusCode(err, http.StatusTooManyRequests)
}

// RetryAfterHeader returns the raw Retry-After value from err when it wraps a
// StatusError with retry metadata.
func RetryAfterHeader(err error) (string, bool) {
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr == nil || statusErr.RetryAfter == "" {
		return "", false
	}
	return statusErr.RetryAfter, true
}

// RetryAfter returns parsed Retry-After metadata from err when VolumeLeaders
// returns either delta-seconds or an HTTP-date value.
func RetryAfter(err error) (time.Duration, bool) {
	header, ok := RetryAfterHeader(err)
	if !ok {
		return 0, false
	}
	duration, parseErr := time.ParseDuration(header + "s")
	if parseErr == nil && duration >= 0 {
		return duration, true
	}
	retryAt, parseErr := http.ParseTime(header)
	if parseErr != nil {
		return 0, false
	}
	duration = time.Until(retryAt)
	if duration < 0 {
		return 0, true
	}
	return duration, true
}

// BodyLimitError reports that a response body exceeded the configured limit.
type BodyLimitError struct {
	Limit int64
}

// Error returns a human-readable body-size failure.
func (e *BodyLimitError) Error() string {
	return fmt.Sprintf("volumeleaders response body exceeded %d byte limit", e.Limit)
}

// IsBodyLimit reports whether err wraps a BodyLimitError.
func IsBodyLimit(err error) bool {
	var limitErr *BodyLimitError
	return errors.As(err, &limitErr)
}

// IsInvalidQuery reports whether err indicates invalid typed request values.
func IsInvalidQuery(err error) bool {
	return errors.Is(err, ErrInvalidQuery)
}

// ContentError reports a response that could not be decoded as the expected
// typed content.
type ContentError struct {
	Method      string
	Path        string
	ContentType string
	Body        string
	Err         error
}

// Error returns a human-readable content decoding failure.
func (e *ContentError) Error() string {
	if e.Method != "" || e.Path != "" {
		return fmt.Sprintf("volumeleaders unexpected response content for %s %s: %v", e.Method, e.Path, e.Err)
	}
	return fmt.Sprintf("volumeleaders unexpected response content: %v", e.Err)
}

// Unwrap returns both the classification sentinel and underlying decode error.
func (e *ContentError) Unwrap() []error {
	if e.Err == nil {
		return []error{ErrUnexpectedContent}
	}
	return []error{ErrUnexpectedContent, e.Err}
}

// IsUnexpectedContent reports whether err indicates a response content mismatch
// or schema decoding failure.
func IsUnexpectedContent(err error) bool {
	return errors.Is(err, ErrUnexpectedContent)
}

// IsSessionExpired reports whether err indicates an expired browser session.
func IsSessionExpired(err error) bool {
	return errors.Is(err, ErrSessionExpired)
}

// IsAuthError reports whether err indicates missing, invalid, or expired
// VolumeLeaders authentication material.
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrInvalidSession) ||
		errors.Is(err, ErrBrowserCookiesUnavailable) ||
		errors.Is(err, ErrSessionExpired) ||
		errors.Is(err, ErrXSRFTokenNotFound) {
		return true
	}
	return IsStatusCode(err, http.StatusUnauthorized) || IsStatusCode(err, http.StatusForbidden)
}
