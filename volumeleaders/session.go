package volumeleaders

import (
	"net/http"
	"slices"
)

// CookieDomain is the VolumeLeaders browser cookie domain used by optional auth
// integrations.
const CookieDomain = volumeLeadersCookieDomain

// SessionCookieName is the ASP.NET session cookie name used by VolumeLeaders.
// Most callers should use NewSession or SessionFromCookies instead of this
// constant.
const SessionCookieName = sessionCookieName

// FormsAuthCookieName is the ASP.NET forms authentication cookie name used by
// VolumeLeaders. Most callers should use NewSession or SessionFromCookies
// instead of this constant.
const FormsAuthCookieName = formsAuthCookieName

// RequestVerificationCookieName is the optional request verification cookie name
// set by VolumeLeaders browser sessions.
const RequestVerificationCookieName = requestVerificationCookie

// Session carries browser authentication material supplied by the caller.
type Session struct {
	Cookies   []*http.Cookie
	XSRFToken string
	Headers   http.Header
}

// NewSession builds a Session from the three values explicit-session callers
// usually store in secret management.
func NewSession(sessionID, authCookie, xsrfToken string) Session {
	return SessionFromCookies([]*http.Cookie{
		// #nosec G124 - callers supply browser cookies, and the client only sends them to the configured origin.
		{Name: sessionCookieName, Value: sessionID},
		// #nosec G124 - callers supply browser cookies, and the client only sends them to the configured origin.
		{Name: formsAuthCookieName, Value: authCookie},
	}, xsrfToken, nil)
}

// SessionFromCookies builds a Session from caller-supplied cookies, XSRF token,
// and additional headers.
func SessionFromCookies(cookies []*http.Cookie, xsrfToken string, headers http.Header) Session {
	return Session{
		Cookies:   cloneCookies(cookies),
		XSRFToken: xsrfToken,
		Headers:   cloneHeader(headers),
	}
}

// MissingSessionFields returns the browser authentication fields that are empty
// in session.
func MissingSessionFields(session Session) []string {
	missing := missingCookieNames(session.Cookies)
	if session.XSRFToken == "" {
		missing = append(missing, "XSRFToken")
	}
	return missing
}

// MissingCookieFields returns required authentication cookie names that are not
// present in cookies.
func MissingCookieFields(cookies []*http.Cookie) []string {
	return missingCookieNames(cookies)
}

// ValidateSession reports whether session contains the cookies and XSRF token
// normally required by authenticated VolumeLeaders browser endpoints.
func ValidateSession(session Session) error {
	missing := MissingSessionFields(session)
	if len(missing) == 0 {
		return nil
	}
	return &SessionValidationError{Missing: missing}
}

func (s Session) clone() Session {
	return Session{
		Cookies:   cloneCookies(s.Cookies),
		XSRFToken: s.XSRFToken,
		Headers:   cloneHeader(s.Headers),
	}
}

func cloneCookies(cookies []*http.Cookie) []*http.Cookie {
	if len(cookies) == 0 {
		return nil
	}
	cloned := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		copyCookie := *cookie
		cloned = append(cloned, &copyCookie)
	}
	return cloned
}

func missingCookieNames(cookies []*http.Cookie) []string {
	names := requiredCookieNames()
	missing := make([]string, 0, len(names))
	for _, name := range names {
		if !slices.ContainsFunc(cookies, func(cookie *http.Cookie) bool {
			return cookie != nil && cookie.Name == name && cookie.Value != ""
		}) {
			missing = append(missing, name)
		}
	}
	return missing
}

func cloneHeader(headers http.Header) http.Header {
	cloned := http.Header{}
	copyHeaderInto(cloned, headers)
	return cloned
}

func copyHeaderInto(dst, src http.Header) {
	for name, values := range src {
		if name == "" {
			continue
		}
		dst.Del(name)
		for _, value := range values {
			dst.Add(name, value)
		}
	}
}
