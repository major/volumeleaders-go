package volumeleaders

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCodeHelpers(t *testing.T) {
	err := fmt.Errorf(
		"wrap: %w",
		&StatusError{StatusCode: http.StatusTooManyRequests, Body: "rate limited", RetryAfter: "30"},
	)

	statusCode, ok := StatusCode(err)
	retryAfter, retryOK := RetryAfter(err)
	retryAfterHeader, retryHeaderOK := RetryAfterHeader(err)

	assert.True(t, ok, "StatusCode(wrapped StatusError) ok")
	assert.Equal(t, http.StatusTooManyRequests, statusCode, "StatusCode(wrapped StatusError)")
	assert.True(t, IsStatusCode(err, http.StatusTooManyRequests), "IsStatusCode(wrapped StatusError, 429)")
	assert.False(t, IsStatusCode(err, http.StatusUnauthorized), "IsStatusCode(wrapped StatusError, 401)")
	assert.True(t, IsRateLimit(err), "IsRateLimit(wrapped StatusError)")
	assert.True(t, retryOK, "RetryAfter(wrapped StatusError) ok")
	assert.Equal(t, 30*time.Second, retryAfter, "RetryAfter(wrapped StatusError)")
	assert.True(t, retryHeaderOK, "RetryAfterHeader(wrapped StatusError) ok")
	assert.Equal(t, "30", retryAfterHeader, "RetryAfterHeader(wrapped StatusError)")
}

func TestStatusCodeHelpersReportMissingStatusError(t *testing.T) {
	statusCode, ok := StatusCode(errors.New("plain error"))

	assert.False(t, ok, "StatusCode(plain error) ok")
	assert.Equal(t, 0, statusCode, "StatusCode(plain error)")
	assert.False(t, IsStatusCode(errors.New("plain error"), http.StatusUnauthorized), "IsStatusCode(plain error, 401)")
}

func TestErrorStrings(t *testing.T) {
	assert.Equal(
		t,
		"volumeleaders API error 401",
		(&StatusError{StatusCode: http.StatusUnauthorized, Body: "not authenticated"}).Error(),
		"StatusError.Error()",
	)
	assert.Equal(
		t,
		"volumeleaders API error POST /Trades/GetTrades 401",
		(&StatusError{StatusCode: http.StatusUnauthorized, Method: http.MethodPost, Path: TradesGetTradesPath, Body: "not authenticated"}).Error(),
		"StatusError.Error() with metadata",
	)
	assert.Equal(
		t,
		"volumeleaders response body exceeded 8 byte limit",
		(&BodyLimitError{Limit: 8}).Error(),
		"BodyLimitError.Error()",
	)
	assert.Equal(
		t,
		"volumeleaders unexpected response content for POST /Trades/GetTrades: boom",
		(&ContentError{Method: http.MethodPost, Path: TradesGetTradesPath, Err: errors.New("boom")}).Error(),
		"ContentError.Error()",
	)
	assert.Equal(
		t,
		"volumeleaders invalid session: missing ASP.NET_SessionId, .ASPXAUTH",
		(&SessionValidationError{Missing: []string{sessionCookieName, formsAuthCookieName}}).Error(),
		"SessionValidationError.Error()",
	)
}

func TestRetryAfterHTTPDatesAndInvalidHeaders(t *testing.T) {
	future := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)

	tests := []struct {
		name         string
		err          error
		wantOK       bool
		wantZeroTime bool
	}{
		{
			name:   "future HTTP date",
			err:    &StatusError{RetryAfter: future.Format(http.TimeFormat)},
			wantOK: true,
		},
		{
			name:         "past HTTP date",
			err:          &StatusError{RetryAfter: past.Format(http.TimeFormat)},
			wantOK:       true,
			wantZeroTime: true,
		},
		{name: "invalid header", err: &StatusError{RetryAfter: "later"}},
		{name: "missing header", err: &StatusError{}},
		{name: "plain error", err: errors.New("plain error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, ok := RetryAfter(tt.err)

			assert.Equal(t, tt.wantOK, ok, "RetryAfter(%v) ok", tt.err)
			if tt.wantZeroTime {
				assert.Equal(t, time.Duration(0), duration, "RetryAfter(%v) duration", tt.err)
			}
			if tt.wantOK && !tt.wantZeroTime {
				assert.Positive(t, duration, "RetryAfter(%v) duration", tt.err)
			}
		})
	}
}

func TestBodyLimitAndSessionExpiredHelpers(t *testing.T) {
	bodyErr := fmt.Errorf("wrap: %w", &BodyLimitError{Limit: 8})
	sessionErr := fmt.Errorf("wrap: %w", sessionExpiredError{redirectPath: "/login"})

	assert.True(t, IsBodyLimit(bodyErr), "IsBodyLimit(wrapped BodyLimitError)")
	assert.False(t, IsBodyLimit(errors.New("plain error")), "IsBodyLimit(plain error)")
	assert.True(t, IsSessionExpired(sessionErr), "IsSessionExpired(wrapped sessionExpiredError)")
	assert.False(t, IsSessionExpired(errors.New("plain error")), "IsSessionExpired(plain error)")
}

func TestUnexpectedContentHelpers(t *testing.T) {
	err := fmt.Errorf("wrap: %w", &ContentError{Err: errors.New("bad json")})

	assert.True(t, IsUnexpectedContent(err), "IsUnexpectedContent(wrapped ContentError)")
	require.ErrorIs(t, err, ErrUnexpectedContent, "errors.Is(ContentError, ErrUnexpectedContent)")
	assert.False(t, IsUnexpectedContent(errors.New("plain error")), "IsUnexpectedContent(plain error)")
}

func TestContentErrorUnwrapWithoutCause(t *testing.T) {
	err := &ContentError{}

	require.ErrorIs(t, err, ErrUnexpectedContent, "errors.Is(ContentError without cause, ErrUnexpectedContent)")
	assert.NotErrorIs(t, err, ErrInvalidQuery, "errors.Is(ContentError without cause, ErrInvalidQuery)")
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "invalid session", err: fmt.Errorf("wrap: %w", ErrInvalidSession), want: true},
		{name: "browser cookies unavailable", err: fmt.Errorf("wrap: %w", ErrBrowserCookiesUnavailable), want: true},
		{name: "session expired", err: fmt.Errorf("wrap: %w", ErrSessionExpired), want: true},
		{name: "XSRF token not found", err: fmt.Errorf("wrap: %w", ErrXSRFTokenNotFound), want: true},
		{name: "unauthorized status", err: &StatusError{StatusCode: http.StatusUnauthorized}, want: true},
		{name: "forbidden status", err: &StatusError{StatusCode: http.StatusForbidden}, want: true},
		{name: "rate limited status", err: &StatusError{StatusCode: http.StatusTooManyRequests}, want: false},
		{name: "plain error", err: errors.New("plain error"), want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsAuthError(tt.err), "IsAuthError(%v)", tt.err)
		})
	}
}
