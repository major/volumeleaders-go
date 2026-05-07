package volumeleaders

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissingSessionFields(t *testing.T) {
	session := Session{
		Cookies: []*http.Cookie{{Name: sessionCookieName, Value: "session-123"}},
	}

	missing := MissingSessionFields(session)

	assert.Equal(t, []string{formsAuthCookieName, "XSRFToken"}, missing, "MissingSessionFields(partial session)")
}

func TestValidateSession(t *testing.T) {
	valid := NewSession("session-123", "auth-123", "xsrf-123")

	require.NoError(t, ValidateSession(valid), "ValidateSession(valid session)")

	err := ValidateSession(Session{})
	require.Error(t, err, "ValidateSession(empty session)")
	require.ErrorIs(t, err, ErrInvalidSession, "ValidateSession(empty session) error")
	assert.True(t, IsAuthError(err), "IsAuthError(ValidateSession(empty session))")

	var validationErr *SessionValidationError
	require.ErrorAs(t, err, &validationErr, "errors.As(ValidateSession error, *SessionValidationError)")
	assert.Equal(
		t,
		[]string{sessionCookieName, formsAuthCookieName, "XSRFToken"},
		validationErr.Missing,
		"SessionValidationError.Missing",
	)
}

func TestNewSessionBuildsRequiredCookies(t *testing.T) {
	session := NewSession("session-123", "auth-123", "xsrf-123")

	assert.Equal(t, "session-123", cookieValueFromSession(t, session, SessionCookieName), "NewSession() session cookie")
	assert.Equal(t, "auth-123", cookieValueFromSession(t, session, FormsAuthCookieName), "NewSession() auth cookie")
	assert.Equal(t, "xsrf-123", session.XSRFToken, "NewSession() XSRFToken")
	assert.NoError(t, ValidateSession(session), "ValidateSession(NewSession())")
}

func TestSessionFromCookiesDefensivelyCopiesInputs(t *testing.T) {
	cookie := &http.Cookie{Name: SessionCookieName, Value: "before"}
	headers := http.Header{"X-Test": {"before"}}

	session := SessionFromCookies([]*http.Cookie{cookie}, "xsrf-123", headers)
	cookie.Value = "after"
	headers.Set("X-Test", "after")

	assert.Equal(
		t,
		"before",
		cookieValueFromSession(t, session, SessionCookieName),
		"SessionFromCookies() copied cookie",
	)
	assert.Equal(t, "before", session.Headers.Get("X-Test"), "SessionFromCookies() copied header")
}

func TestSessionFromCookiesSkipsNilCookiesAndBlankHeaderNames(t *testing.T) {
	session := SessionFromCookies(
		[]*http.Cookie{nil, {Name: SessionCookieName, Value: "session-123"}},
		"xsrf-123",
		http.Header{"": {"ignored"}, "X-Test": {"kept"}},
	)

	assert.Len(t, session.Cookies, 1, "SessionFromCookies() copied cookies")
	assert.Equal(t, "session-123", cookieValueFromSession(t, session, SessionCookieName), "SessionFromCookies() copied cookie")
	assert.Empty(t, session.Headers.Values(""), "SessionFromCookies() skipped blank header name")
	assert.Equal(t, []string{"kept"}, session.Headers.Values("X-Test"), "SessionFromCookies() copied header")
}

func TestMissingCookieFields(t *testing.T) {
	missing := MissingCookieFields([]*http.Cookie{{Name: SessionCookieName, Value: "session-123"}})

	assert.Equal(t, []string{FormsAuthCookieName}, missing, "MissingCookieFields(partial cookies)")
}

func cookieValueFromSession(t *testing.T, session Session, name string) string {
	t.Helper()
	for _, cookie := range session.Cookies {
		if cookie != nil && cookie.Name == name {
			return cookie.Value
		}
	}
	t.Fatalf("Session cookie %q missing", name)
	return ""
}
