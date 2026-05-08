package browserauth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/browserutils/kooky"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/major/volumeleaders-go/volumeleaders"
)

type testBrowserInfo struct {
	browser string
	profile string
}

func (b testBrowserInfo) Browser() string {
	return b.browser
}

func (b testBrowserInfo) Profile() string {
	return b.profile
}

func (b testBrowserInfo) IsDefaultProfile() bool {
	return b.profile == "default"
}

func (b testBrowserInfo) FilePath() string {
	return "/tmp/" + b.browser + "/" + b.profile
}

func TestVolumeLeadersCookiesSelectsRequiredCookieValues(t *testing.T) {
	cookies := kooky.Cookies{
		nil,
		{Cookie: http.Cookie{Name: "unrelated", Value: "ignore"}},
		{Cookie: http.Cookie{Name: volumeleaders.SessionCookieName, Value: "session-id"}},
		{Cookie: http.Cookie{Name: volumeleaders.FormsAuthCookieName, Value: "forms-auth"}},
		{
			Cookie: http.Cookie{
				Name:  volumeleaders.RequestVerificationCookieName,
				Value: "verification",
			},
		},
	}

	got := volumeLeadersCookies(cookies)
	if len(got) != 3 {
		t.Fatalf(
			"volumeLeadersCookies(%d cookies) returned %d cookies, want 3",
			len(cookies),
			len(got),
		)
	}

	want := map[string]string{
		volumeleaders.SessionCookieName:             "session-id",
		volumeleaders.FormsAuthCookieName:           "forms-auth",
		volumeleaders.RequestVerificationCookieName: "verification",
	}
	for _, cookie := range got {
		if cookie == nil {
			t.Fatal("volumeLeadersCookies() returned nil cookie")
		}
		if cookie.Value != want[cookie.Name] {
			t.Errorf(
				"volumeLeadersCookies() cookie %q = %q, want %q",
				cookie.Name,
				cookie.Value,
				want[cookie.Name],
			)
		}
	}
}

func TestBrowserStoreCountCountsDistinctBrowserProfiles(t *testing.T) {
	cookies := kooky.Cookies{
		nil,
		{
			Cookie:  http.Cookie{Name: volumeleaders.SessionCookieName},
			Browser: testBrowserInfo{browser: "firefox", profile: "default"},
		},
		{
			Cookie:  http.Cookie{Name: volumeleaders.FormsAuthCookieName},
			Browser: testBrowserInfo{browser: "firefox", profile: "default"},
		},
		{
			Cookie:  http.Cookie{Name: volumeleaders.RequestVerificationCookieName},
			Browser: testBrowserInfo{browser: "firefox", profile: "work"},
		},
		{
			Cookie:  http.Cookie{Name: "extra"},
			Browser: testBrowserInfo{browser: "chromium", profile: "default"},
		},
	}

	if got, want := browserStoreCount(cookies), 4; got != want {
		t.Errorf("browserStoreCount(%d cookies) = %d, want %d", len(cookies), got, want)
	}
}

func TestExtractCookiesReturnsValidBrowserCookies(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return requiredKookyCookies(), nil
	})

	got, err := ExtractCookies(context.Background())
	if err != nil {
		t.Fatalf("ExtractCookies() error = %v, want nil", err)
	}
	if missing := volumeleaders.MissingCookieFields(got); len(missing) != 0 {
		t.Fatalf("ExtractCookies() missing fields = %v, want none", missing)
	}
}

func TestExtractCookiesReportsBrowserDiagnostics(t *testing.T) {
	readCalls := 0
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		readCalls++
		if readCalls == 1 {
			return kooky.Cookies{
					{
						Cookie: http.Cookie{
							Name:  volumeleaders.SessionCookieName,
							Value: "session-id",
						},
					},
				}, errors.New(
					"valid store failed",
				)
		}
		return kooky.Cookies{
			{
				Cookie:  http.Cookie{Name: volumeleaders.SessionCookieName, Value: "session-id"},
				Browser: testBrowserInfo{browser: "firefox", profile: "default"},
			},
			{
				Cookie:  http.Cookie{Name: volumeleaders.FormsAuthCookieName, Value: "forms-auth"},
				Browser: testBrowserInfo{browser: "chromium", profile: "work"},
			},
		}, errors.New("all stores failed")
	})

	_, err := ExtractCookies(context.Background())
	if !errors.Is(err, volumeleaders.ErrBrowserCookiesUnavailable) {
		t.Fatalf("ExtractCookies() error = %v, want ErrBrowserCookiesUnavailable", err)
	}
	for _, want := range []string{"missing .ASPXAUTH", "valid cookies found: 1", "browser stores with matching cookies: 2"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("ExtractCookies() error = %q, want substring %q", err.Error(), want)
		}
	}
}

func TestFindSessionWithBrowserSelectsOnlyMatchingBrowser(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie("firefox", "default", volumeleaders.SessionCookieName, "firefox-session"),
			browserCookie("firefox", "default", volumeleaders.FormsAuthCookieName, "firefox-auth"),
			browserCookie(
				"firefox",
				"default",
				volumeleaders.RequestVerificationCookieName,
				"firefox-verification",
			),
			browserCookie("chrome", "default", volumeleaders.SessionCookieName, "chrome-session"),
			browserCookie("chrome", "default", volumeleaders.FormsAuthCookieName, "chrome-auth"),
			browserCookie(
				"chrome",
				"default",
				volumeleaders.RequestVerificationCookieName,
				"chrome-verification",
			),
		}, nil
	})

	session, err := FindSession(context.Background(), WithBrowser("firefox"), WithoutValidation())
	require.NoError(t, err, "FindSession(WithBrowser(%q)) returned unexpected error", "firefox")

	assertSessionCookieValues(t, session, map[string]string{
		volumeleaders.SessionCookieName:             "firefox-session",
		volumeleaders.FormsAuthCookieName:           "firefox-auth",
		volumeleaders.RequestVerificationCookieName: "firefox-verification",
	})
	assert.NotContains(t, sessionCookieValueList(session), "chrome-session",
		"FindSession(WithBrowser(%q)) must not use other browser cookies", "firefox")
}

func TestFindSessionWithProfileSelectsOnlyMatchingProfile(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		wantCookies map[string]string
	}{
		{
			name:    "default profile",
			profile: "default",
			wantCookies: map[string]string{
				volumeleaders.SessionCookieName:             "default-session",
				volumeleaders.FormsAuthCookieName:           "default-auth",
				volumeleaders.RequestVerificationCookieName: "default-verification",
			},
		},
		{
			name:    "work profile",
			profile: "work",
			wantCookies: map[string]string{
				volumeleaders.SessionCookieName:             "work-session",
				volumeleaders.FormsAuthCookieName:           "work-auth",
				volumeleaders.RequestVerificationCookieName: "work-verification",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stubReadBrowserCookies(
				t,
				func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
					return kooky.Cookies{
						browserCookie(
							"firefox",
							"default",
							volumeleaders.SessionCookieName,
							"default-session",
						),
						browserCookie(
							"firefox",
							"default",
							volumeleaders.FormsAuthCookieName,
							"default-auth",
						),
						browserCookie(
							"firefox",
							"default",
							volumeleaders.RequestVerificationCookieName,
							"default-verification",
						),
						browserCookie(
							"firefox",
							"work",
							volumeleaders.SessionCookieName,
							"work-session",
						),
						browserCookie(
							"firefox",
							"work",
							volumeleaders.FormsAuthCookieName,
							"work-auth",
						),
						browserCookie(
							"firefox",
							"work",
							volumeleaders.RequestVerificationCookieName,
							"work-verification",
						),
					}, nil
				},
			)

			session, err := FindSession(
				context.Background(),
				WithProfile(tt.profile),
				WithoutValidation(),
			)
			require.NoError(
				t,
				err,
				"FindSession(WithProfile(%q)) returned unexpected error",
				tt.profile,
			)
			assertSessionCookieValues(t, session, tt.wantCookies)
		})
	}
}

func TestFindSessionWithBrowserNoFallback(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie("chrome", "default", "unrelated", "chrome-cookie"),
			browserCookie("firefox", "default", volumeleaders.SessionCookieName, "firefox-session"),
			browserCookie("firefox", "default", volumeleaders.FormsAuthCookieName, "firefox-auth"),
			browserCookie(
				"firefox",
				"default",
				volumeleaders.RequestVerificationCookieName,
				"firefox-verification",
			),
		}, nil
	})

	session, err := FindSession(context.Background(), WithBrowser("chrome"), WithoutValidation())
	require.Error(
		t,
		err,
		"FindSession(WithBrowser(%q)) must fail when selected browser lacks required cookies",
		"chrome",
	)
	assert.Empty(t, session.Cookies,
		"FindSession(WithBrowser(%q)) must not fall back to unselected browser cookies", "chrome")
}

func TestFindSessionWithProfileNoFallback(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie("firefox", "default", "unrelated", "default-cookie"),
			browserCookie("firefox", "work", volumeleaders.SessionCookieName, "work-session"),
			browserCookie("firefox", "work", volumeleaders.FormsAuthCookieName, "work-auth"),
			browserCookie(
				"firefox",
				"work",
				volumeleaders.RequestVerificationCookieName,
				"work-verification",
			),
		}, nil
	})

	session, err := FindSession(context.Background(), WithProfile("default"), WithoutValidation())
	require.Error(
		t,
		err,
		"FindSession(WithProfile(%q)) must fail when selected profile lacks required cookies",
		"default",
	)
	assert.Empty(t, session.Cookies,
		"FindSession(WithProfile(%q)) must not fall back to unselected profile cookies", "default")
}

func TestFindSessionParityDiscoversValidatedSession(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return requiredKookyCookies(), nil
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method, "FindSession() validation method mismatch")
		assert.Equal(
			t,
			"/ExecutiveSummary",
			r.URL.Path,
			"FindSession() validation request path mismatch",
		)
		assert.Contains(
			t,
			r.Header.Get("Cookie"),
			volumeleaders.SessionCookieName+"=session-id",
			"FindSession() validation request must include session cookie",
		)
		assert.Contains(
			t,
			r.Header.Get("Cookie"),
			volumeleaders.FormsAuthCookieName+"=forms-auth",
			"FindSession() validation request must include forms auth cookie",
		)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(
				`<html><body><input name="__RequestVerificationToken" value="test-xsrf-token"></body></html>`,
			),
		)
	}))
	t.Cleanup(server.Close)

	session, err := FindSession(context.Background(), WithClientOptions(
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	))
	require.NoError(t, err, "FindSession() returned unexpected error")
	assert.Equal(
		t,
		"test-xsrf-token",
		session.XSRFToken,
		"FindSession() must return fetched XSRF token",
	)
	assertSessionCookieValues(t, session, map[string]string{
		volumeleaders.SessionCookieName:             "session-id",
		volumeleaders.FormsAuthCookieName:           "forms-auth",
		volumeleaders.RequestVerificationCookieName: "verification",
	})
}

func TestFindSessionToleratesPartialBrowserReadErrors(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie("firefox", "default", volumeleaders.SessionCookieName, "session-id"),
			browserCookie("firefox", "default", volumeleaders.FormsAuthCookieName, "forms-auth"),
			browserCookie(
				"firefox",
				"default",
				volumeleaders.RequestVerificationCookieName,
				"verification",
			),
		}, errors.New("chrome cookie store read failed")
	})

	session, err := FindSession(context.Background(), WithoutValidation())
	require.NoError(
		t,
		err,
		"FindSession() must tolerate browser read errors when required cookies are found",
	)
	assertSessionCookieValues(t, session, map[string]string{
		volumeleaders.SessionCookieName:             "session-id",
		volumeleaders.FormsAuthCookieName:           "forms-auth",
		volumeleaders.RequestVerificationCookieName: "verification",
	})
}

func TestFindSessionWithUnsupportedBrowserReturnsError(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie("chrome", "default", volumeleaders.SessionCookieName, "chrome-session"),
			browserCookie("chrome", "default", volumeleaders.FormsAuthCookieName, "chrome-auth"),
			browserCookie(
				"chrome",
				"default",
				volumeleaders.RequestVerificationCookieName,
				"chrome-verification",
			),
		}, nil
	})

	session, err := FindSession(context.Background(), WithBrowser("firefox"), WithoutValidation())
	require.Error(
		t,
		err,
		"FindSession(WithBrowser(%q)) must fail when selected browser has no cookies",
		"firefox",
	)
	assert.Empty(
		t,
		session.Cookies,
		"FindSession(WithBrowser(%q)) must not return cookies from another browser",
		"firefox",
	)
	require.ErrorIs(t, err, ErrBrowserUnavailable,
		"FindSession(WithBrowser(%q)) error must identify unavailable browser", "firefox")
	require.ErrorIs(t, err, ErrRequiredCookieMissing,
		"FindSession(WithBrowser(%q)) error must identify missing required cookies", "firefox")
}

func TestFindSessionMissingRequiredCookiesReturnsError(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie(
				"firefox",
				"default",
				volumeleaders.SessionCookieName,
				"secret-session-id",
			),
			browserCookie(
				"firefox",
				"default",
				volumeleaders.RequestVerificationCookieName,
				"secret-xsrf-token",
			),
		}, nil
	})

	session, err := FindSession(context.Background(), WithoutValidation())
	require.Error(
		t,
		err,
		"FindSession() must fail when %q is missing",
		volumeleaders.FormsAuthCookieName,
	)
	assert.Empty(
		t,
		session.Cookies,
		"FindSession() must not return a partial session when required cookies are missing",
	)
	require.ErrorIs(
		t,
		err,
		ErrRequiredCookieMissing,
		"FindSession() error must identify missing required cookies",
	)
	assertNoSecretLeakage(t, err, fixtureSecrets())
}

func TestFindSessionMissingXSRFTokenReturnsError(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return secretKookyCookies(), nil
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/ExecutiveSummary", r.URL.Path, "FindSession() XSRF request path mismatch")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(`<html><body>credential-bearing-body without token input</body></html>`),
		)
	}))
	t.Cleanup(server.Close)

	session, err := FindSession(context.Background(), WithClientOptions(
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	))
	require.Error(t, err, "FindSession() must fail when validation page lacks XSRF token")
	assert.Empty(
		t,
		session.Cookies,
		"FindSession() must not return a session when XSRF validation fails",
	)
	require.ErrorIs(t, err, volumeleaders.ErrXSRFTokenNotFound,
		"FindSession() error must preserve missing XSRF token cause")
	require.ErrorIs(t, err, ErrRequestVerificationTokenMissing,
		"FindSession() error must identify missing request verification token")
	require.ErrorIs(
		t,
		err,
		ErrValidationFailed,
		"FindSession() error must identify validation failure",
	)
	assertNoSecretLeakage(t, err, fixtureSecrets())
}

func TestFindSessionExpiredSessionReturnsError(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return secretKookyCookies(), nil
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ExecutiveSummary":
			http.Redirect(w, r, "/Login", http.StatusFound)
		case "/Login":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(
				[]byte(`<html><body class="login">credential-bearing-body</body></html>`),
			)
		default:
			assert.Failf(
				t,
				"FindSession() validation request path mismatch",
				"path = %q",
				r.URL.Path,
			)
		}
	}))
	t.Cleanup(server.Close)

	session, err := FindSession(context.Background(), WithClientOptions(
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	))
	require.Error(t, err, "FindSession() must fail when validation redirects to login")
	assert.Empty(t, session.Cookies, "FindSession() must not return an expired session")
	require.ErrorIs(t, err, volumeleaders.ErrSessionExpired,
		"FindSession() error must identify expired sessions")
	require.ErrorIs(
		t,
		err,
		ErrValidationFailed,
		"FindSession() error must identify validation failure",
	)
	assertNoSecretLeakage(t, err, fixtureSecrets())
}

func TestFindSessionValidatesDiscoveredSession(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return requiredKookyCookies(), nil
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(
			t,
			"/ExecutiveSummary",
			r.URL.Path,
			"FindSession() validation request path mismatch",
		)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(
				`<html><body><input name="__RequestVerificationToken" value="test-xsrf-token"></body></html>`,
			),
		)
	}))
	t.Cleanup(server.Close)

	session, err := FindSession(context.Background(), WithClientOptions(
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	))
	require.NoError(t, err, "FindSession() returned unexpected error")
	assert.Equal(
		t,
		"test-xsrf-token",
		session.XSRFToken,
		"FindSession() must return XSRF token from validation",
	)
	assert.NotEmpty(t, session.Cookies, "FindSession() must return cookies")
}

func TestFindSessionRespectsContextCancellation(t *testing.T) {
	tests := []struct {
		name    string
		cookies kooky.Cookies
	}{
		{
			name: "no cookies",
		},
		{
			name: "partial cookies",
			cookies: kooky.Cookies{
				{Cookie: http.Cookie{Name: volumeleaders.SessionCookieName, Value: "session-id"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			stubReadBrowserCookies(
				t,
				func(ctx context.Context, _ ...kooky.Filter) (kooky.Cookies, error) {
					return tt.cookies, ctx.Err()
				},
			)

			_, err := FindSession(ctx)
			require.Error(t, err, "FindSession() with canceled context must return error")
			assert.ErrorIs(
				t,
				err,
				context.Canceled,
				"FindSession() with canceled context must return context.Canceled",
			)
		})
	}
}

func TestNewBuildsClientFromBrowserCookiesAndFetchedXSRFToken(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return requiredKookyCookies(), nil
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ExecutiveSummary" {
			t.Errorf("New() XSRF request path = %q, want /ExecutiveSummary", r.URL.Path)
		}
		if got := r.Header.Get("Cookie"); !strings.Contains(
			got,
			volumeleaders.SessionCookieName+"=session-id",
		) {
			t.Errorf("New() XSRF request Cookie = %q, want session cookie", got)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(
				`<html><body><input name="__RequestVerificationToken" value="xsrf-token"></body></html>`,
			),
		)
	}))
	t.Cleanup(server.Close)

	client, err := New(
		context.Background(),
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if client == nil {
		t.Fatal("New() client = nil, want client")
	}
}

func stubReadBrowserCookies(
	t *testing.T,
	stub func(context.Context, ...kooky.Filter) (kooky.Cookies, error),
) {
	t.Helper()
	original := readBrowserCookies
	readBrowserCookies = stub
	t.Cleanup(func() {
		readBrowserCookies = original
	})
}

func requiredKookyCookies() kooky.Cookies {
	return kooky.Cookies{
		{Cookie: http.Cookie{Name: volumeleaders.SessionCookieName, Value: "session-id"}},
		{Cookie: http.Cookie{Name: volumeleaders.FormsAuthCookieName, Value: "forms-auth"}},
		{
			Cookie: http.Cookie{
				Name:  volumeleaders.RequestVerificationCookieName,
				Value: "verification",
			},
		},
	}
}

func browserCookie(browser, profile, name, value string) *kooky.Cookie {
	return &kooky.Cookie{
		Cookie:  http.Cookie{Name: name, Value: value},
		Browser: testBrowserInfo{browser: browser, profile: profile},
	}
}

func assertSessionCookieValues(
	t *testing.T,
	session volumeleaders.Session,
	want map[string]string,
) {
	t.Helper()

	got := sessionCookieValues(session)
	assert.Equal(t, want, got, "FindSession() selected cookies mismatch")
}

func sessionCookieValues(session volumeleaders.Session) map[string]string {
	values := make(map[string]string, len(session.Cookies))
	for _, cookie := range session.Cookies {
		if cookie == nil {
			continue
		}
		values[cookie.Name] = cookie.Value
	}
	return values
}

func sessionCookieValueList(session volumeleaders.Session) []string {
	values := make([]string, 0, len(session.Cookies))
	for _, cookie := range session.Cookies {
		if cookie == nil {
			continue
		}
		values = append(values, cookie.Value)
	}
	return values
}

// fixtureSecrets returns well-known credential values injected by redaction
// tests. If any appear in an error message, the error path leaks sensitive
// material.
func fixtureSecrets() []string {
	return []string{
		"secret-session-id",
		"secret-forms-auth",
		"secret-xsrf-token",
		"/tmp/secret-profile/Default",
		"credential-bearing-body",
	}
}

// assertNoSecretLeakage checks that err.Error() does not contain any of the
// provided fixture secrets.
func assertNoSecretLeakage(t *testing.T, err error, secrets []string) {
	t.Helper()
	require.Error(t, err)
	errMsg := err.Error()
	for _, secret := range secrets {
		assert.NotContains(t, errMsg, secret,
			"error message must not leak sensitive value")
	}
}

// secretKookyCookies returns kooky cookies carrying fixture secret values and
// browser metadata that must never appear in error messages.
// testBrowserInfo{browser: "secret-profile", profile: "Default"} produces
// FilePath() = "/tmp/secret-profile/Default", matching a fixture secret.
func secretKookyCookies() kooky.Cookies {
	return kooky.Cookies{
		{
			Cookie:  http.Cookie{Name: volumeleaders.SessionCookieName, Value: "secret-session-id"},
			Browser: testBrowserInfo{browser: "secret-profile", profile: "Default"},
		},
		{
			Cookie: http.Cookie{
				Name:  volumeleaders.FormsAuthCookieName,
				Value: "secret-forms-auth",
			},
			Browser: testBrowserInfo{browser: "secret-profile", profile: "Default"},
		},
		{
			Cookie: http.Cookie{
				Name:  volumeleaders.RequestVerificationCookieName,
				Value: "secret-xsrf-token",
			},
			Browser: testBrowserInfo{browser: "secret-profile", profile: "Default"},
		},
	}
}

// --- Secret redaction tests ---
//
// Each test injects fixture secrets into an error path and asserts that
// err.Error() does not expose them.

func TestFindSessionRedactsBrowserReadErrorSecrets(t *testing.T) {
	// kooky returns an error whose message embeds a browser profile path.
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return nil, errors.New(
			"sqlite: open /tmp/secret-profile/Default/Cookies: permission denied",
		)
	})

	_, err := FindSession(context.Background())
	assertNoSecretLeakage(t, err, fixtureSecrets())
	assert.ErrorIs(t, err, volumeleaders.ErrBrowserCookiesUnavailable,
		"browser read error must remain matchable")
}

func TestFindSessionRedactsMissingCookieSecrets(t *testing.T) {
	// Only session cookie present with a secret value; .ASPXAUTH missing.
	// The secret cookie value must not appear in the error.
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			{
				Cookie: http.Cookie{
					Name:  volumeleaders.SessionCookieName,
					Value: "secret-session-id",
				},
				Browser: testBrowserInfo{browser: "secret-profile", profile: "Default"},
			},
		}, nil
	})

	_, err := FindSession(context.Background())
	assertNoSecretLeakage(t, err, fixtureSecrets())
	assert.ErrorIs(t, err, volumeleaders.ErrBrowserCookiesUnavailable,
		"missing cookie error must remain matchable")
}

func TestFindSessionRedactsValidationErrorSecrets(t *testing.T) {
	// Browser filter selects a browser that has only a partial set of
	// VolumeLeaders cookies. The cookie values present must not leak.
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie(
				"chrome",
				"Default",
				volumeleaders.SessionCookieName,
				"secret-session-id",
			),
			// chrome is missing .ASPXAUTH; firefox has it but shouldn't be selected.
			browserCookie(
				"firefox",
				"default",
				volumeleaders.FormsAuthCookieName,
				"secret-forms-auth",
			),
		}, nil
	})

	_, err := FindSession(context.Background(), WithBrowser("chrome"))
	assertNoSecretLeakage(t, err, fixtureSecrets())
	assert.ErrorIs(t, err, volumeleaders.ErrBrowserCookiesUnavailable,
		"validation error must remain matchable")
}

func TestFindSessionRedactsSessionExpiredSecrets(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return secretKookyCookies(), nil
	})

	// Server returns a page that looks like a login page, triggering session
	// expired detection. The response body contains fixture secrets.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(
			`<html><body class="login">` +
				`<form action="/Login/Login"><input type="password" name="Password"></form>` +
				`credential-bearing-body</body></html>`,
		))
	}))
	t.Cleanup(server.Close)

	_, err := FindSession(context.Background(), WithClientOptions(
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	))
	assertNoSecretLeakage(t, err, fixtureSecrets())
	assert.ErrorIs(t, err, volumeleaders.ErrSessionExpired,
		"session expired error must remain matchable")
}

func TestFindSessionRedactsMissingXSRFTokenSecrets(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return secretKookyCookies(), nil
	})

	// Server returns valid HTML without a __RequestVerificationToken input.
	// The body embeds fixture secrets that must not leak.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(`<html><body>secret-xsrf-token credential-bearing-body</body></html>`),
		)
	}))
	t.Cleanup(server.Close)

	_, err := FindSession(context.Background(), WithClientOptions(
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	))
	assertNoSecretLeakage(t, err, fixtureSecrets())
	assert.ErrorIs(t, err, volumeleaders.ErrXSRFTokenNotFound,
		"missing XSRF token error must remain matchable")
}

func TestFindSessionRedactsStatusErrorSecrets(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return secretKookyCookies(), nil
	})

	// Server returns 500 with a body that contains fixture secrets.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`credential-bearing-body`))
	}))
	t.Cleanup(server.Close)

	_, err := FindSession(context.Background(), WithClientOptions(
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	))
	assertNoSecretLeakage(t, err, fixtureSecrets())

	var statusErr *volumeleaders.StatusError
	assert.ErrorAs(t, err, &statusErr,
		"status error must remain matchable with errors.As")
}
