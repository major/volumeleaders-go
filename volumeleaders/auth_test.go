package volumeleaders

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchXSRFTokenReadsAuthenticatedExecutiveSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, executiveSummaryPath, r.URL.Path, "FetchXSRFToken() path")
		assert.Equal(t, "session-123", cookieValue(t, r, sessionCookieName), "FetchXSRFToken() session cookie")
		assert.Equal(t, "auth-123", cookieValue(t, r, formsAuthCookieName), "FetchXSRFToken() auth cookie")
		assert.Equal(t, defaultUserAgent, r.Header.Get("User-Agent"), "FetchXSRFToken() User-Agent")
		assert.Equal(
			t,
			"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
			r.Header.Get("Accept"),
			"FetchXSRFToken() Accept",
		)
		_, _ = w.Write([]byte(`<html><input name="__RequestVerificationToken" type="hidden" value="xsrf-123"></html>`))
	}))
	t.Cleanup(server.Close)

	token, err := FetchXSRFToken(context.Background(), Session{
		Cookies: []*http.Cookie{
			{Name: sessionCookieName, Value: "session-123"},
			{Name: formsAuthCookieName, Value: "auth-123"},
		},
	}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))

	require.NoError(t, err, "FetchXSRFToken()")
	assert.Equal(t, "xsrf-123", token, "FetchXSRFToken() token")
}

func TestFetchXSRFTokenReportsMissingToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html></html>`))
	}))
	t.Cleanup(server.Close)

	_, err := FetchXSRFToken(context.Background(), Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))

	assert.ErrorIs(t, err, ErrXSRFTokenNotFound, "FetchXSRFToken() missing token")
}

func TestFetchXSRFTokenReportsSessionExpiredRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == executiveSummaryPath {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		_, _ = w.Write([]byte(`<html>login</html>`))
	}))
	t.Cleanup(server.Close)

	_, err := FetchXSRFToken(context.Background(), Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))

	require.Error(t, err, "FetchXSRFToken() expired session")
	assert.True(t, IsSessionExpired(err), "IsSessionExpired(FetchXSRFToken error)")
}

func TestFetchXSRFTokenReportsSessionExpiredLoginPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(
			`<html><title>Login</title><body>` +
				`<form action="/Login/Login"><input type="password" name="Password"></form>` +
				`</body></html>`,
		))
	}))
	t.Cleanup(server.Close)

	_, err := FetchXSRFToken(context.Background(), Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))

	require.Error(t, err, "FetchXSRFToken() login page")
	assert.True(t, IsSessionExpired(err), "IsSessionExpired(FetchXSRFToken login page error)")
}

func TestFetchXSRFTokenIgnoresIncidentalLoginString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// Authenticated page containing "Login" in a JS conditional (not a login form).
		_, _ = w.Write([]byte(
			`<html><body>` +
				`<input name="__RequestVerificationToken" type="hidden" value="xsrf-ok">` +
				`<script>if ("ExecutiveSummary" != "Login") { init(); }</script>` +
				`</body></html>`,
		))
	}))
	t.Cleanup(server.Close)

	token, err := FetchXSRFToken(
		context.Background(), Session{},
		WithBaseURL(server.URL), WithHTTPClient(server.Client()),
	)

	require.NoError(t, err, "FetchXSRFToken() with incidental login string")
	assert.Equal(t, "xsrf-ok", token, "FetchXSRFToken() token")
}

func TestParseXSRFTokenHandlesAttributeOrder(t *testing.T) {
	token, err := parseXSRFToken([]byte(`<input type="hidden" value="xsrf-456" name="__RequestVerificationToken">`))

	require.NoError(t, err, "parseXSRFToken()")
	assert.Equal(t, "xsrf-456", token, "parseXSRFToken() token")
}

func TestParseXSRFTokenRejectsEmptyValue(t *testing.T) {
	_, err := parseXSRFToken([]byte(`<input name="__RequestVerificationToken" value="">`))

	assert.ErrorIs(t, err, ErrXSRFTokenNotFound, "parseXSRFToken() empty value")
}

func TestSafeResponsePath(t *testing.T) {
	tests := []struct {
		name string
		resp *http.Response
		want string
	}{
		{name: "nil response", want: "unknown redirect target"},
		{name: "nil request", resp: &http.Response{}, want: "unknown redirect target"},
		{name: "nil URL", resp: &http.Response{Request: &http.Request{}}, want: "unknown redirect target"},
		{
			name: "empty path",
			resp: &http.Response{Request: &http.Request{URL: &url.URL{Scheme: "https", Host: "volumeleaders.com"}}},
			want: "/",
		},
		{
			name: "escaped path",
			resp: &http.Response{Request: &http.Request{URL: &url.URL{Path: "/Login Path"}}},
			want: "/Login%20Path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, safeResponsePath(tt.resp), "safeResponsePath(%s)", tt.name)
		})
	}
}

func TestNormalizeResponsePath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "", want: "/"},
		{path: "login", want: "/login"},
		{path: "/LOGIN", want: "/login"},
		{path: "/trades/../Login", want: "/login"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeResponsePath(tt.path), "normalizeResponsePath(%q)", tt.path)
		})
	}
}

func TestMissingCookieNames(t *testing.T) {
	missing := missingCookieNames([]*http.Cookie{{Name: sessionCookieName, Value: "session-123"}})

	assert.Equal(t, []string{formsAuthCookieName}, missing, "missingCookieNames()")
}

func cookieValue(t *testing.T, r *http.Request, name string) string {
	t.Helper()
	cookie, err := r.Cookie(name)
	require.NoError(t, err, "Request.Cookie(%q)", name)
	return cookie.Value
}
