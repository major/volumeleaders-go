package browserauth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/browserutils/kooky"

	"github.com/major/volumeleaders-go/volumeleaders"
)

// ExampleFindSession shows the default discovery path for desktop automation.
// FindSession reads cookies from all supported local browser stores, validates
// the session against VolumeLeaders, and returns a Session ready for NewClient.
func ExampleFindSession() {
	restore := stubExampleBrowserCookies()
	defer restore()
	server := newExampleValidationServer()
	defer server.Close()

	session, err := FindSession(context.Background(), WithClientOptions(
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	))
	if err != nil {
		fmt.Println(err)
		return
	}

	client, err := volumeleaders.NewClient(session)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(client != nil, session.XSRFToken)
	// Output:
	// true example-xsrf-token
}

// ExampleFindSession_withBrowser shows explicit browser selection when multiple
// supported browsers are installed. Only cookies from the named browser are
// considered; FindSession returns ErrBrowserUnavailable if that browser has no
// matching VolumeLeaders cookies.
func ExampleFindSession_withBrowser() {
	restore := stubExampleBrowserCookies()
	defer restore()

	session, err := FindSession(
		context.Background(),
		WithBrowser("firefox"),
		WithoutValidation(),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(session.Cookies[0].Value)
	// Output:
	// firefox-session-id
}

// ExampleFindSession_withoutValidation shows skipping the XSRF token fetch.
// The returned Session contains the discovered browser cookies but has an empty
// XSRFToken field. Use this when the caller will supply the token separately or
// when the validation request itself is not needed.
func ExampleFindSession_withoutValidation() {
	restore := stubExampleBrowserCookies()
	defer restore()

	session, err := FindSession(context.Background(), WithoutValidation())
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(session.XSRFToken == "")
	// Output:
	// true
}

func stubExampleBrowserCookies() func() {
	original := readBrowserCookies
	readBrowserCookies = func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return kooky.Cookies{
			browserCookie(
				"firefox",
				"default",
				volumeleaders.SessionCookieName,
				"firefox-session-id",
			),
			browserCookie("firefox", "default", volumeleaders.FormsAuthCookieName, "firefox-auth"),
			browserCookie(
				"firefox",
				"default",
				volumeleaders.RequestVerificationCookieName,
				"firefox-verification",
			),
		}, nil
	}
	return func() {
		readBrowserCookies = original
	}
}

func newExampleValidationServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(
			[]byte(`<input name="__RequestVerificationToken" value="example-xsrf-token">`),
		)
	}))
}
