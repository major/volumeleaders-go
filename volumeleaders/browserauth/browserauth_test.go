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
		{Cookie: http.Cookie{Name: volumeleaders.RequestVerificationCookieName, Value: "verification"}},
	}

	got := volumeLeadersCookies(cookies)
	if len(got) != 3 {
		t.Fatalf("volumeLeadersCookies(%d cookies) returned %d cookies, want 3", len(cookies), len(got))
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
			t.Errorf("volumeLeadersCookies() cookie %q = %q, want %q", cookie.Name, cookie.Value, want[cookie.Name])
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
		{Cookie: http.Cookie{Name: "extra"}, Browser: testBrowserInfo{browser: "chromium", profile: "default"}},
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
					{Cookie: http.Cookie{Name: volumeleaders.SessionCookieName, Value: "session-id"}},
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
			browserCookie("firefox", "default", volumeleaders.RequestVerificationCookieName, "firefox-verification"),
			browserCookie("chrome", "default", volumeleaders.SessionCookieName, "chrome-session"),
			browserCookie("chrome", "default", volumeleaders.FormsAuthCookieName, "chrome-auth"),
			browserCookie("chrome", "default", volumeleaders.RequestVerificationCookieName, "chrome-verification"),
		}, nil
	})

	session, err := FindSession(context.Background(), WithBrowser("firefox"))
	require.NoError(t, err, "FindSession(WithBrowser(%q)) returned unexpected error", "firefox")

	assertSessionCookieValues(t, session, map[string]string{
		volumeleaders.SessionCookieName:             "firefox-session",
		volumeleaders.FormsAuthCookieName:           "firefox-auth",
		volumeleaders.RequestVerificationCookieName: "firefox-verification",
	})
	assert.NotContains(t, sessionCookieValueList(session), "chrome-session", "FindSession(WithBrowser(%q)) must not use other browser cookies", "firefox")
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
			stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
				return kooky.Cookies{
					browserCookie("firefox", "default", volumeleaders.SessionCookieName, "default-session"),
					browserCookie("firefox", "default", volumeleaders.FormsAuthCookieName, "default-auth"),
					browserCookie("firefox", "default", volumeleaders.RequestVerificationCookieName, "default-verification"),
					browserCookie("firefox", "work", volumeleaders.SessionCookieName, "work-session"),
					browserCookie("firefox", "work", volumeleaders.FormsAuthCookieName, "work-auth"),
					browserCookie("firefox", "work", volumeleaders.RequestVerificationCookieName, "work-verification"),
				}, nil
			})

			session, err := FindSession(context.Background(), WithProfile(tt.profile))
			require.NoError(t, err, "FindSession(WithProfile(%q)) returned unexpected error", tt.profile)
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
			browserCookie("firefox", "default", volumeleaders.RequestVerificationCookieName, "firefox-verification"),
		}, nil
	})

	session, err := FindSession(context.Background(), WithBrowser("chrome"))
	require.Error(t, err, "FindSession(WithBrowser(%q)) must fail when selected browser lacks required cookies", "chrome")
	assert.Empty(t, session.Cookies, "FindSession(WithBrowser(%q)) must not fall back to unselected browser cookies", "chrome")
}

func TestFindSessionWithProfileNoFallback(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie("firefox", "default", "unrelated", "default-cookie"),
			browserCookie("firefox", "work", volumeleaders.SessionCookieName, "work-session"),
			browserCookie("firefox", "work", volumeleaders.FormsAuthCookieName, "work-auth"),
			browserCookie("firefox", "work", volumeleaders.RequestVerificationCookieName, "work-verification"),
		}, nil
	})

	session, err := FindSession(context.Background(), WithProfile("default"))
	require.Error(t, err, "FindSession(WithProfile(%q)) must fail when selected profile lacks required cookies", "default")
	assert.Empty(t, session.Cookies, "FindSession(WithProfile(%q)) must not fall back to unselected profile cookies", "default")
}

func TestNewBuildsClientFromBrowserCookiesAndFetchedXSRFToken(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return requiredKookyCookies(), nil
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ExecutiveSummary" {
			t.Errorf("New() XSRF request path = %q, want /ExecutiveSummary", r.URL.Path)
		}
		if got := r.Header.Get("Cookie"); !strings.Contains(got, volumeleaders.SessionCookieName+"=session-id") {
			t.Errorf("New() XSRF request Cookie = %q, want session cookie", got)
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><input name="__RequestVerificationToken" value="xsrf-token"></body></html>`))
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

func stubReadBrowserCookies(t *testing.T, stub func(context.Context, ...kooky.Filter) (kooky.Cookies, error)) {
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
		{Cookie: http.Cookie{Name: volumeleaders.RequestVerificationCookieName, Value: "verification"}},
	}
}

func browserCookie(browser, profile, name, value string) *kooky.Cookie {
	return &kooky.Cookie{
		Cookie:  http.Cookie{Name: name, Value: value},
		Browser: testBrowserInfo{browser: browser, profile: profile},
	}
}

func assertSessionCookieValues(t *testing.T, session volumeleaders.Session, want map[string]string) {
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
